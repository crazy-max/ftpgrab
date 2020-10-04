package grabber

import (
	"fmt"
	"os"
	"path"
	"runtime"
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
	config *config.Download
	db     *db.Client
	server *server.Client
}

// New creates new grabber instance
func New(dlConfig *config.Download, dbConfig *config.Db, serverConfig *config.Server) (*Client, error) {
	var dbCli *db.Client
	var serverCli *server.Client
	var err error

	// DB client
	if dbCli, err = db.New(dbConfig); err != nil {
		return nil, errors.Wrap(err, "Cannot open database")
	}

	// Server client
	if serverConfig.FTP != nil {
		serverCli, err = ftp.New(serverConfig.FTP)
	} else if serverConfig.SFTP != nil {
		serverCli, err = sftp.New(serverConfig.SFTP)
	} else {
		return nil, errors.New("No server defined")
	}
	if err != nil {
		return nil, errors.Wrap(err, "Cannot connect to server")
	}

	return &Client{
		config: dlConfig,
		db:     dbCli,
		server: serverCli,
	}, nil
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
	destpath := path.Join(file.DestDir, file.Info.Name())

	entry := &journal.Entry{
		File:   srcpath,
		Status: c.getStatus(file),
	}

	sublogger := log.With().
		Str("file", entry.File).
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
	sublogger.Info().Str("dest", destpath).Msg("Downloading...")

	destfolder := path.Dir(destpath)
	if err := os.MkdirAll(destfolder, os.ModePerm); err != nil {
		sublogger.Error().Err(err).Msg("Cannot create destination dir")
		entry.Level = journal.EntryLevelError
		entry.Text = fmt.Sprintf("Cannot create destination dir: %v", err)
		return entry
	}
	if err := c.fixPerms(destfolder); err != nil {
		sublogger.Warn().Err(err).Msg("Cannot fix parent folder permissions")
	}

	destfile, err := os.Create(destpath)
	if err != nil {
		sublogger.Error().Err(err).Msg("Cannot create destination file")
		entry.Level = journal.EntryLevelError
		entry.Text = fmt.Sprintf("Cannot create destination file: %v", err)
		return entry
	}

	err = c.server.Retrieve(srcpath, destfile)
	if err != nil {
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

func (c *Client) fixPerms(filepath string) error {
	if runtime.GOOS == "windows" {
		return nil
	}

	fileinfo, err := os.Stat(filepath)
	if err != nil {
		return err
	}

	chmod := os.FileMode(c.config.ChmodFile)
	if fileinfo.IsDir() {
		chmod = os.FileMode(c.config.ChmodDir)
	}

	if err := os.Chmod(filepath, chmod); err != nil {
		return err
	}

	if err := os.Chown(filepath, c.config.UID, c.config.GID); err != nil {
		return err
	}

	return nil
}

// Close closes grabber
func (c *Client) Close() {
	if err := c.db.Close(); err != nil {
		log.Warn().Err(err).Msg("Cannot close database")
	}
	if err := c.server.Close(); err != nil {
		log.Warn().Err(err).Msg("Cannot close server connection")
	}
}
