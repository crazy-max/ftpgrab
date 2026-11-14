package grabber

import (
	"crypto/rand"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/db"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/crazy-max/ftpgrab/v7/internal/server"
	"github.com/crazy-max/ftpgrab/v7/internal/server/ftp"
	"github.com/crazy-max/ftpgrab/v7/internal/server/sftp"
	"github.com/docker/go-units"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Client represents an active grabber object
type Client struct {
	config     *config.Download
	db         *db.Client
	server     *server.Client
	root       *os.Root
	tempdirRun string
	closeOnce  sync.Once
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

	if client.root, err = os.OpenRoot(dlConfig.Output); err != nil {
		return nil, errors.Wrap(err, "Cannot open download output folder")
	}

	if *dlConfig.TempFirst {
		if client.tempdirRun, err = client.makeTempDir("."); err != nil {
			return nil, errors.Wrap(err, "Cannot create temporary destination dir")
		}
	}

	return client, nil
}

func (c *Client) Grab(files []File) journal.Journal {
	jnl := journal.New()
	jnl.ServerHost = c.server.Common().Host

	if c.config.Concurrency <= 1 || len(files) < 2 {
		for _, file := range files {
			if entry := c.download(file, 0); entry != nil {
				jnl.Add(*entry)
			}
		}
		return jnl.Journal
	}

	groups := c.downloadGroups(files)

	entries := make([]*journal.Entry, len(files))
	jobs := make(chan []int)
	var workers sync.WaitGroup
	for range min(c.config.Concurrency, len(groups)) {
		workers.Go(func() {
			for group := range jobs {
				for _, i := range group {
					entries[i] = c.download(files[i], 0)
				}
			}
		})
	}
	for _, group := range groups {
		jobs <- group
	}
	close(jobs)
	workers.Wait()

	// Assemble the journal after all workers finish to preserve input order.
	for _, entry := range entries {
		if entry != nil {
			jnl.Add(*entry)
		}
	}

	return jnl.Journal
}

// downloadGroups keeps files sharing a destination or history key in input order.
// A file can connect two groups, so merge their roots before assembling jobs.
func (c *Client) downloadGroups(files []File) [][]int {
	parents := make([]int, len(files))
	for i := range parents {
		parents[i] = i
	}
	root := func(i int) int {
		for parents[i] != i {
			parents[i] = parents[parents[i]]
			i = parents[i]
		}
		return i
	}

	previous := make(map[string]int)
	for i, file := range files {
		// Conservatively serialize case variants, including on case-insensitive
		// filesystems mounted on Unix. Temporary paths follow destination paths.
		keys := []string{"dest:" + strings.ToLower(filepath.Join(file.DestDir, file.Info.Name()))}
		if c.db.Enabled() {
			keys = append(keys, "history:"+strings.TrimPrefix(path.Join(file.SrcDir, file.Info.Name()), file.Base))
		}
		for _, key := range keys {
			if prev, ok := previous[key]; ok {
				parents[root(prev)] = i
			}
			previous[key] = i
		}
	}

	var groups [][]int
	indices := make(map[int]int)
	for i := range files {
		r := root(i)
		group, ok := indices[r]
		if !ok {
			group = len(groups)
			indices[r] = group
			groups = append(groups, nil)
		}
		groups[group] = append(groups[group], i)
	}

	return groups
}

func (c *Client) download(file File, retry int) *journal.Entry {
	srcpath := path.Join(file.SrcDir, file.Info.Name())
	destpath := filepath.Join(file.DestDir, file.Info.Name())

	entry := &journal.Entry{
		File: srcpath,
	}
	if !validRemoteName(file.Info.Name()) || !filepath.IsLocal(file.DestDir) {
		entry.Level = journal.EntryLevelError
		entry.Text = "Unsafe download destination"
		log.Error().Str("src", srcpath).Str("dest", destpath).Msg(entry.Text)
		return entry
	}
	entry.Status = c.getStatus(file)

	sublogger := log.With().
		Str("src", entry.File).
		Str("dest", file.DestDir).
		Str("size", units.HumanSize(float64(file.Info.Size()))).
		Logger()

	if entry.Status == journal.EntryStatusAlreadyDl && !c.db.HasDigest(file.Base, file.SrcDir, file.Info) {
		if err := c.db.PutDigest(file.Base, file.SrcDir, file.Info); err != nil {
			sublogger.Error().Err(err).Msg("Cannot add digest into db")
			entry.Level = journal.EntryLevelWarning
			entry.Text = fmt.Sprintf("Already downloaded but cannot add digest into db: %v", err)
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
	if err := c.root.MkdirAll(destfolder, os.ModePerm); err != nil {
		sublogger.Error().Err(err).Msg("Cannot create destination dir")
		entry.Level = journal.EntryLevelError
		entry.Text = fmt.Sprintf("Cannot create destination dir: %v", err)
		return entry
	}
	if err := c.fixDirPerms(destfolder); err != nil {
		sublogger.Warn().Err(err).Msg("Cannot fix parent folder permissions")
	}

	destfile, temppath, err := c.createDownloadFile(destpath)
	if err != nil {
		sublogger.Error().Err(err).Msg("Cannot create destination file")
		entry.Level = journal.EntryLevelError
		entry.Text = fmt.Sprintf("Cannot create destination file: %v", err)
		return entry
	}

	if err := c.server.Retrieve(srcpath, destfile); err != nil {
		if cleanupErr := c.closeAndRemoveTempFile(destfile, temppath); cleanupErr != nil {
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
		if err := c.fixFilePerms(destfile); err != nil {
			sublogger.Warn().Err(err).Msg("Cannot fix file permissions")
		}
		if err := destfile.Close(); err != nil {
			if cleanupErr := c.removeTempPath(temppath); cleanupErr != nil {
				sublogger.Warn().Err(cleanupErr).Str("path", destfile.Name()).Msg("Cannot clean temporary destination file")
			}
			sublogger.Error().Err(err).Msg("Cannot close destination file")
			entry.Level = journal.EntryLevelError
			entry.Text = fmt.Sprintf("Cannot close destination file: %v", err)
			return entry
		}

		if err := c.root.Chtimes(temppath, file.Info.ModTime(), file.Info.ModTime()); err != nil {
			sublogger.Warn().Err(err).Msg("Cannot change modtime of destination file")
		}

		if err := c.root.Rename(temppath, destpath); err != nil {
			if cleanupErr := c.removeTempPath(temppath); cleanupErr != nil {
				sublogger.Warn().Err(cleanupErr).Str("path", temppath).Msg("Cannot clean temporary destination file")
			}
			sublogger.Error().Err(err).Msg("Cannot move file")
			entry.Level = journal.EntryLevelError
			entry.Text = fmt.Sprintf("Cannot move file: %v", err)
			return entry
		}
		if err := c.removeTempPath(temppath); err != nil {
			sublogger.Warn().Err(err).Str("path", temppath).Msg("Cannot clean temporary destination folder")
		}

		sublogger.Info().
			Str("duration", time.Since(retrieveStart).Round(time.Millisecond).String()).
			Msg("File successfully downloaded")

		entry.Level = journal.EntryLevelSuccess
		entry.Text = fmt.Sprintf("%s successfully downloaded in %s",
			units.HumanSize(float64(file.Info.Size())),
			time.Since(retrieveStart).Round(time.Millisecond).String(),
		)
		if err := c.db.PutDigest(file.Base, file.SrcDir, file.Info); err != nil {
			sublogger.Error().Err(err).Msg("Cannot add digest into db")
			entry.Level = journal.EntryLevelWarning
			entry.Text = fmt.Sprintf("Successfully downloaded but cannot add digest into db: %v", err)
		}
	}

	return entry
}

func (c *Client) createDownloadFile(filename string) (*os.File, string, error) {
	parent := filepath.Dir(filename)
	if *c.config.TempFirst {
		parent = c.tempdirRun
	}
	tempdir, err := c.makeTempDir(parent)
	if err != nil {
		return nil, "", err
	}
	temppath := filepath.Join(tempdir, filepath.Base(filename))
	// Never open an existing inode, including links, devices, and FIFOs.
	file, err := c.root.OpenFile(temppath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if cleanupErr := c.root.Remove(tempdir); cleanupErr != nil {
			log.Warn().Err(cleanupErr).Str("path", tempdir).Msg("Cannot clean temporary destination folder")
		}
		return nil, "", err
	}
	return file, temppath, nil
}

func (c *Client) getStatus(file File) journal.EntryStatus {
	if !c.isIncluded(file) {
		return journal.EntryStatusNotIncluded
	} else if c.isExcluded(file) {
		return journal.EntryStatusExcluded
	} else if file.Info.ModTime().Before(c.config.SinceTime) {
		return journal.EntryStatusOutdated
	} else if *c.config.Force {
		return journal.EntryStatusForcedDl
	} else if destfile, err := c.root.Stat(filepath.Join(file.DestDir, file.Info.Name())); err == nil {
		if destfile.Size() == file.Info.Size() {
			return journal.EntryStatusAlreadyDl
		}
		return journal.EntryStatusSizeDiff
	} else if c.db.HasDigest(file.Base, file.SrcDir, file.Info) {
		return journal.EntryStatusDigestExists
	}
	return journal.EntryStatusNeverDl
}

func (c *Client) isIncluded(file File) bool {
	if len(c.config.Include) == 0 {
		return true
	}
	for _, include := range c.config.Include {
		if matchString(include, file.Info.Name()) {
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
		if matchString(exclude, file.Info.Name()) {
			return true
		}
	}
	return false
}

// Close closes grabber
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		if err := c.db.Close(); err != nil {
			log.Warn().Err(err).Msg("Cannot close database")
		}
		if err := c.server.Close(); err != nil {
			log.Warn().Err(err).Msg("Cannot close server connection")
		}
		if c.root != nil {
			if c.tempdirRun != "" {
				if err := c.root.RemoveAll(c.tempdirRun); err != nil {
					log.Warn().Err(err).Str("path", c.tempdirRun).Msg("Cannot remove temporary session folder")
				}
			}
			if err := c.root.Close(); err != nil {
				log.Warn().Err(err).Msg("Cannot close download output folder")
			}
		}
	})
}

func (c *Client) makeTempDir(parent string) (string, error) {
	// Root has no MkdirTemp. Mkdir exclusively claims a random private directory.
	for {
		name := filepath.Join(parent, ".ftpgrab-tmp-"+rand.Text())
		if err := c.root.Mkdir(name, 0o700); os.IsExist(err) {
			continue
		} else if err != nil {
			return "", err
		}
		return name, nil
	}
}

func (c *Client) closeAndRemoveTempFile(file *os.File, filename string) error {
	closeErr := file.Close()
	removeErr := c.removeTempPath(filename)
	if closeErr != nil {
		return closeErr
	}
	return removeErr
}

func (c *Client) removeTempPath(filename string) error {
	if err := c.root.Remove(filename); err != nil && !os.IsNotExist(err) {
		return err
	}
	return c.root.Remove(filepath.Dir(filename))
}

func matchString(exp string, s string) bool {
	re, err := regexp.Compile(exp)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}
