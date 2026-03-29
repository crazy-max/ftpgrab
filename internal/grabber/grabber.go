package grabber

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/db"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/crazy-max/ftpgrab/v7/internal/server"
	"github.com/crazy-max/ftpgrab/v7/internal/server/ftp"
	"github.com/crazy-max/ftpgrab/v7/internal/server/sftp"
	"github.com/crazy-max/ftpgrab/v7/pkg/utl"
	"github.com/docker/go-units"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Client represents an active grabber object
type Client struct {
	config     *config.Download
	db         *db.Client
	server     *server.Client
	tempdirRun string
}

// New creates new grabber instance
func New(dlConfig *config.Download, dbConfig *config.Db, serverConfig *config.Server) (client *Client, err error) {
	var dbCli *db.Client
	var serverctl *server.Client

	if dbCli, err = db.New(dbConfig); err != nil {
		return nil, errors.Wrap(err, "Cannot open database")
	}
	defer func() {
		if err == nil {
			return
		}
		if client != nil {
			client.Close()
			return
		}
		if serverctl != nil {
			_ = serverctl.Close()
		}
		if dbCli != nil {
			_ = dbCli.Close()
		}
	}()

	if serverConfig.FTP != nil {
		serverctl, err = ftp.New(serverConfig.FTP)
	} else if serverConfig.SFTP != nil {
		serverctl, err = sftp.New(serverConfig.SFTP)
	} else {
		return nil, errors.New("No server defined")
	}
	if err != nil {
		return nil, errors.Wrap(err, "Cannot connect to server")
	}

	client = &Client{
		config: dlConfig,
		db:     dbCli,
		server: serverctl,
	}

	if *dlConfig.TempFirst {
		if err := client.initTempDir(); err != nil {
			return nil, errors.Wrap(err, "Cannot create temporary destination dir")
		}
	}

	return client, nil
}

func (c *Client) Grab(files []File) journal.Journal {
	jnl := journal.New()
	jnl.ServerHost = c.server.Common().Host

	for _, file := range files {
		if entry := c.download(file, 0); entry != nil {
			jnl.Add(*entry)
		}
	}

	return jnl.Journal
}

func (c *Client) download(file File, retry int) *journal.Entry {
	srcpath := path.Join(file.SrcDir, file.Info.Name())
	destpath := filepath.Join(file.DestDir, file.Info.Name())

	entry := &journal.Entry{
		File:   srcpath,
		Status: c.getStatus(file),
	}

	sublogger := log.With().
		Str("src", entry.File).
		Str("dest", file.DestDir).
		Str("size", units.HumanSize(float64(file.Info.Size()))).
		Logger()

	if entry.Status == journal.EntryStatusAlreadyDl && !c.db.HasHash(file.Base, file.SrcDir, file.Info) {
		if err := c.db.PutHash(file.Base, file.SrcDir, file.Info); err != nil {
			sublogger.Error().Err(err).Msg("Cannot add hash into db")
			entry.Level = journal.EntryLevelWarning
			entry.Text = fmt.Sprintf("Already downloaded but cannot add hash into db: %v", err)
			return entry
		}
	}

	if entry.Status.IsSkipped() {
		if !*c.config.HideSkipped {
			sublogger.Warn().Msgf("Skipped (%s)", entry.Status)
			entry.Level = journal.EntryLevelSkip
			return entry
		}
		return nil
	}

	retrieveStart := time.Now()

	destfolder := filepath.Dir(destpath)
	if err := os.MkdirAll(destfolder, os.ModePerm); err != nil {
		sublogger.Error().Err(err).Msg("Cannot create destination dir")
		entry.Level = journal.EntryLevelError
		entry.Text = fmt.Sprintf("Cannot create destination dir: %v", err)
		return entry
	}
	if err := c.fixPerms(destfolder); err != nil {
		sublogger.Warn().Err(err).Msg("Cannot fix parent folder permissions")
	}

	destfile, err := c.createDownloadFile(destpath)
	if err != nil {
		sublogger.Error().Err(err).Msg("Cannot create destination file")
		entry.Level = journal.EntryLevelError
		entry.Text = fmt.Sprintf("Cannot create destination file: %v", err)
		return entry
	}

	err = c.server.Retrieve(srcpath, destfile)
	if err != nil {
		if cleanupErr := c.closeAndRemoveTempFile(destfile); cleanupErr != nil {
			sublogger.Warn().Err(cleanupErr).Str("path", destfile.Name()).Msg("Cannot clean destination file after download failure")
		}
		retry++
		sublogger.Error().Err(err).Msgf("Error downloading, retry %d/%d", retry, c.config.Retry)
		if retry == c.config.Retry {
			sublogger.Error().Err(err).Msg("Cannot download file")
			entry.Level = journal.EntryLevelError
			entry.Text = fmt.Sprintf("Cannot download file: %v", err)
		} else {
			return c.download(file, retry)
		}
	} else {
		if err := destfile.Close(); err != nil {
			if cleanupErr := c.removeTempPath(destfile.Name()); cleanupErr != nil {
				sublogger.Warn().Err(cleanupErr).Str("path", destfile.Name()).Msg("Cannot clean temporary destination file")
			}
			sublogger.Error().Err(err).Msg("Cannot close destination file")
			entry.Level = journal.EntryLevelError
			entry.Text = fmt.Sprintf("Cannot close destination file: %v", err)
			return entry
		}

		if *c.config.TempFirst {
			log.Debug().
				Str("tempfile", destfile.Name()).
				Str("destfile", destpath).
				Msgf("Move temp file")
			if err := moveFile(destfile.Name(), destpath); err != nil {
				if cleanupErr := c.removeTempPath(destfile.Name()); cleanupErr != nil {
					sublogger.Warn().Err(cleanupErr).Str("path", destfile.Name()).Msg("Cannot clean temporary destination file")
				}
				sublogger.Error().Err(err).Msg("Cannot move file")
				entry.Level = journal.EntryLevelError
				entry.Text = fmt.Sprintf("Cannot move file: %v", err)
				return entry
			}
		}

		sublogger.Info().
			Str("duration", time.Since(retrieveStart).Round(time.Millisecond).String()).
			Msg("File successfully downloaded")

		entry.Level = journal.EntryLevelSuccess
		entry.Text = fmt.Sprintf("%s successfully downloaded in %s",
			units.HumanSize(float64(file.Info.Size())),
			time.Since(retrieveStart).Round(time.Millisecond).String(),
		)
		if err := c.fixPerms(destpath); err != nil {
			sublogger.Warn().Err(err).Msg("Cannot fix file permissions")
		}
		if err := c.db.PutHash(file.Base, file.SrcDir, file.Info); err != nil {
			sublogger.Error().Err(err).Msg("Cannot add hash into db")
			entry.Level = journal.EntryLevelWarning
			entry.Text = fmt.Sprintf("Successfully downloaded but cannot add hash into db: %v", err)
		}
		if err = os.Chtimes(destpath, file.Info.ModTime(), file.Info.ModTime()); err != nil {
			sublogger.Warn().Err(err).Msg("Cannot change modtime of destination file")
		}
	}

	return entry
}

func (c *Client) createDownloadFile(filename string) (*os.File, error) {
	if !*c.config.TempFirst {
		return os.Create(filename)
	}
	tempfilepath := c.tempFilePath(filename)
	tempdir := filepath.Dir(tempfilepath)
	if err := os.MkdirAll(tempdir, os.ModePerm); err != nil {
		return nil, err
	}
	if err := c.fixPerms(tempdir); err != nil {
		log.Warn().Err(err).Str("path", tempdir).Msg("Cannot fix temporary destination folder permissions")
	}
	return os.Create(tempfilepath)
}

func (c *Client) getStatus(file File) journal.EntryStatus {
	if !c.isIncluded(file) {
		return journal.EntryStatusNotIncluded
	} else if c.isExcluded(file) {
		return journal.EntryStatusExcluded
	} else if file.Info.ModTime().Before(c.config.SinceTime) {
		return journal.EntryStatusOutdated
	} else if destfile, err := os.Stat(path.Join(file.DestDir, file.Info.Name())); err == nil {
		if destfile.Size() == file.Info.Size() {
			return journal.EntryStatusAlreadyDl
		}
		return journal.EntryStatusSizeDiff
	} else if c.db.HasHash(file.Base, file.SrcDir, file.Info) {
		return journal.EntryStatusHashExists
	}
	return journal.EntryStatusNeverDl
}

func (c *Client) isIncluded(file File) bool {
	if len(c.config.Include) == 0 {
		return true
	}
	for _, include := range c.config.Include {
		if utl.MatchString(include, file.Info.Name()) {
			return true
		}
	}
	return false
}

func (c *Client) isExcluded(file File) bool {
	if len(c.config.Exclude) == 0 {
		return false
	}
	for _, exclude := range c.config.Exclude {
		if utl.MatchString(exclude, file.Info.Name()) {
			return true
		}
	}
	return false
}

// Close closes grabber
func (c *Client) Close() {
	if err := c.db.Close(); err != nil {
		log.Warn().Err(err).Msg("Cannot close database")
	}
	if err := c.server.Close(); err != nil {
		log.Warn().Err(err).Msg("Cannot close server connection")
	}
	if *c.config.TempFirst {
		if c.tempdirRun != "" {
			if err := os.RemoveAll(c.tempdirRun); err != nil && !os.IsNotExist(err) {
				log.Warn().Err(err).Str("path", c.tempdirRun).Msg("Cannot remove temporary session folder")
			}
		}
		tempRootDir := c.tempRootDir()
		if err := os.Remove(tempRootDir); err != nil && !os.IsNotExist(err) {
			log.Debug().Err(err).Str("path", tempRootDir).Msg("Temporary root folder not removed")
		}
	}
}

func (c *Client) initTempDir() error {
	tempRootDir := c.tempRootDir()
	if err := os.MkdirAll(tempRootDir, os.ModePerm); err != nil {
		return err
	}
	if err := c.fixPerms(tempRootDir); err != nil {
		log.Warn().Err(err).Str("path", tempRootDir).Msg("Cannot fix temporary root folder permissions")
	}

	tempdirRun, err := os.MkdirTemp(tempRootDir, "")
	if err != nil {
		return err
	}
	c.tempdirRun = tempdirRun
	if err := c.fixPerms(c.tempdirRun); err != nil {
		log.Warn().Err(err).Str("path", c.tempdirRun).Msg("Cannot fix temporary session folder permissions")
	}
	return nil
}

func (c *Client) tempRootDir() string {
	return filepath.Join(c.config.Output, ".ftpgrab-tmp")
}

func (c *Client) tempFilePath(filename string) string {
	relpath, err := filepath.Rel(c.config.Output, filename)
	if err != nil || relpath == "." || relpath == "" {
		return filepath.Join(c.tempdirRun, filepath.Base(filename))
	}
	return filepath.Join(c.tempdirRun, relpath)
}

func (c *Client) closeAndRemoveTempFile(file *os.File) error {
	closeErr := file.Close()
	removeErr := c.removeTempPath(file.Name())
	if closeErr != nil {
		return closeErr
	}
	return removeErr
}

func (c *Client) removeTempPath(filename string) error {
	if !*c.config.TempFirst {
		return nil
	}
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
