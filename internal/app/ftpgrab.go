package app

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/docker/go-units"
	"github.com/ftpgrab/ftpgrab/internal/config"
	"github.com/ftpgrab/ftpgrab/internal/db"
	"github.com/ftpgrab/ftpgrab/internal/journal"
	"github.com/ftpgrab/ftpgrab/internal/model"
	"github.com/ftpgrab/ftpgrab/internal/notif"
	"github.com/ftpgrab/ftpgrab/internal/server"
	"github.com/ftpgrab/ftpgrab/internal/server/ftp"
	"github.com/ftpgrab/ftpgrab/internal/server/sftp"
	"github.com/ftpgrab/ftpgrab/internal/utl"
	"github.com/hako/durafmt"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

// FtpGrab represents an active ftpgrab object
type FtpGrab struct {
	cfg    *config.Configuration
	cron   *cron.Cron
	srv    *server.Client
	db     *db.Client
	notif  *notif.Client
	jnl    *journal.Client
	jobID  cron.EntryID
	locker uint32
}

const (
	outdated    = model.EntryStatus("Outdated file")
	notIncluded = model.EntryStatus("Not included")
	excluded    = model.EntryStatus("Excluded")
	neverDl     = model.EntryStatus("Never downloaded")
	alreadyDl   = model.EntryStatus("Already downloaded")
	sizeDiff    = model.EntryStatus("Exists but size is different")
	hashExists  = model.EntryStatus("Hash sum exists")
)

// New creates new ftpgrab instance
func New(cfg *config.Configuration, location *time.Location) (*FtpGrab, error) {
	if err := os.MkdirAll(cfg.Download.Output, os.ModePerm); err != nil {
		return nil, fmt.Errorf("cannot create output download folder %s, %v", cfg.Download.Output, err)
	}

	if err := os.MkdirAll(path.Dir(cfg.Db.Path), os.ModePerm); err != nil {
		return nil, fmt.Errorf("cannot create database destination folder %s, %v", path.Dir(cfg.Db.Path), err)
	}

	return &FtpGrab{
		cfg:  cfg,
		cron: cron.New(cron.WithLocation(location), cron.WithSeconds()),
	}, nil
}

// Start starts ftpgrab
func (fg *FtpGrab) Start() error {
	var err error

	// Run on startup
	fg.Run()

	// Init scheduler if defined
	if fg.cfg.Flags.Schedule == "" {
		return nil
	}
	fg.jobID, err = fg.cron.AddJob(fg.cfg.Flags.Schedule, fg)
	if err != nil {
		return err
	}
	log.Info().Msgf("Cron initialized with schedule %s", fg.cfg.Flags.Schedule)

	// Start scheduler
	fg.cron.Start()
	log.Info().Msgf("Next run in %s (%s)",
		durafmt.ParseShort(fg.cron.Entry(fg.jobID).Next.Sub(time.Now())).String(),
		fg.cron.Entry(fg.jobID).Next)

	select {}
}

// Run runs ftpgrab process
func (fg *FtpGrab) Run() {
	if !atomic.CompareAndSwapUint32(&fg.locker, 0, 1) {
		log.Warn().Msg("Already running")
		return
	}
	defer atomic.StoreUint32(&fg.locker, 0)
	if fg.jobID > 0 {
		defer log.Info().Msgf("Next run in %s (%s)",
			durafmt.ParseShort(fg.cron.Entry(fg.jobID).Next.Sub(time.Now())).String(),
			fg.cron.Entry(fg.jobID).Next)
	}

	start := time.Now()
	var err error

	// Journal client
	fg.jnl = journal.New()

	// Server client
	switch fg.cfg.Server.Type {
	case model.ServerTypeFTP:
		fg.srv, err = ftp.New(&fg.cfg.Server.FTP)
	case model.ServerTypeSFTP:
		fg.srv, err = sftp.New(&fg.cfg.Server.SFTP)
	default:
		log.Fatal().Err(err).Msgf("Unknown server type %s", fg.cfg.Server.Type)
	}
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot connect to server")
	}

	// DB client
	if fg.db, err = db.New(&fg.cfg.Db); err != nil {
		log.Fatal().Err(err).Msg("Cannot open database")
	}

	// Notification client
	if fg.notif, err = notif.New(fg.cfg.Notif, fg.cfg.App, fg.srv.Common()); err != nil {
		log.Fatal().Err(err).Msg("Cannot create notifiers")
	}

	// Iterate sources
	for _, src := range fg.srv.Common().Sources {
		log.Info().Str("source", src).Msg("Grabbing")

		// Check basedir
		dest := fg.cfg.Download.Output
		if src != "/" && fg.cfg.Download.CreateBasedir {
			dest = path.Join(dest, src)
		}

		// Retrieve recursively
		fg.retrieveRecursive(src, src, dest)
	}

	if err := fg.srv.Close(); err != nil {
		log.Warn().Err(err).Msg("Cannot close server connection")
	}
	if err := fg.db.Close(); err != nil {
		log.Warn().Err(err).Msg("Cannot close database")
	}
	fg.jnl.Duration = time.Since(start)

	log.Info().
		Str("duration", time.Since(start).Round(time.Millisecond).String()).
		Msg("Finished")

	// Check journal before sending report
	if fg.jnl.IsEmpty() {
		log.Warn().Msg("Journal empty, skip sending report")
		return
	}

	// Send notifications
	fg.notif.Send(*fg.jnl)
}

// Close closes ftpgrab
func (fg *FtpGrab) Close() {
	if err := fg.srv.Close(); err != nil {
		log.Warn().Err(err).Msg("Cannot close server connection")
	}
	if err := fg.db.Close(); err != nil {
		log.Warn().Err(err).Msg("Cannot close database")
	}
	if fg.cron != nil {
		fg.cron.Stop()
	}
}

func (fg *FtpGrab) retrieveRecursive(base string, source string, dest string) {
	// Check source dir exists
	files, err := fg.srv.ReadDir(source)
	if err != nil {
		log.Error().Err(err).Str("source", base).
			Msgf("Cannot read directory %s", source)
		return
	}

	for _, file := range files {
		if jnlEntry := fg.retrieve(base, source, dest, file, 0); jnlEntry != nil {
			fg.jnl.AddEntry(*jnlEntry)
		}
	}
}

func (fg *FtpGrab) retrieve(base string, src string, dest string, file os.FileInfo, retry int) *model.Entry {
	srcpath := path.Join(src, file.Name())
	destpath := path.Join(dest, file.Name())

	if file.IsDir() {
		fg.retrieveRecursive(base, srcpath, destpath)
		return nil
	}

	status := fg.fileStatus(base, src, dest, file)
	jnlEntry := &model.Entry{
		File:       srcpath,
		StatusText: string(status),
	}

	sublogger := log.With().
		Str("file", jnlEntry.File).
		Str("size", units.HumanSize(float64(file.Size()))).
		Logger()

	if status == alreadyDl && !fg.db.HasHash(base, src, file) {
		if err := fg.db.PutHash(base, src, file); err != nil {
			sublogger.Error().Err(err).Msg("Cannot add hash into db")
		}
	}
	if fg.isSkipped(status) {
		if !fg.cfg.Download.HideSkipped {
			sublogger.Warn().Str(".status", jnlEntry.StatusText).Msg("Skipped")
			jnlEntry.StatusType = "skip"
			return jnlEntry
		}
		return nil
	}

	retrieveStart := time.Now()
	sublogger.Info().Str("dest", destpath).Msg("Downloading...")

	destfolder := path.Dir(destpath)
	if err := os.MkdirAll(destfolder, os.ModePerm); err != nil {
		sublogger.Error().Err(err).Msg("Cannot create destination dir")
		jnlEntry.StatusType = "error"
		jnlEntry.StatusText = fmt.Sprintf("Cannot create destination dir: %v", err)
		return jnlEntry
	}
	if err := fg.fixPerms(destfolder); err != nil {
		sublogger.Warn().Err(err).Msg("Cannot fix parent folder permissions")
	}

	destfile, err := os.Create(destpath)
	if err != nil {
		sublogger.Error().Err(err).Msg("Cannot create destination file")
		jnlEntry.StatusType = "error"
		jnlEntry.StatusText = fmt.Sprintf("Cannot create destination file: %v", err)
		return jnlEntry
	}

	err = fg.srv.Retrieve(srcpath, destfile)
	if err != nil {
		retry++
		sublogger.Error().Err(err).Msgf("Error downloading, retry %d/%d", retry, fg.cfg.Download.Retry)
		if retry == fg.cfg.Download.Retry {
			sublogger.Error().Err(err).Msg("Cannot download file")
			jnlEntry.StatusType = "error"
			jnlEntry.StatusText = fmt.Sprintf("Cannot download file: %v", err)
		} else {
			fg.retrieve(base, src, dest, file, retry)
			return nil
		}
	} else {
		sublogger.Info().
			Str("duration", time.Since(retrieveStart).Round(time.Millisecond).String()).
			Msg("File successfully downloaded")
		jnlEntry.StatusType = "success"
		jnlEntry.StatusText = fmt.Sprintf("%s successfully downloaded in %s",
			units.HumanSize(float64(file.Size())),
			time.Since(retrieveStart).Round(time.Millisecond).String(),
		)
		if err := fg.fixPerms(destpath); err != nil {
			sublogger.Warn().Err(err).Msg("Cannot fix file permissions")
		}
		if err := fg.db.PutHash(base, src, file); err != nil {
			sublogger.Error().Err(err).Msg("Cannot add hash into db")
			jnlEntry.StatusType = "warning"
			jnlEntry.StatusText = fmt.Sprintf("Successfully downloaded but cannot add hash into db: %v", err)
		}
		if err = os.Chtimes(destpath, file.ModTime(), file.ModTime()); err != nil {
			sublogger.Warn().Err(err).Msg("Cannot change modtime of destination file")
		}
	}

	return jnlEntry
}

func (fg *FtpGrab) fileStatus(base string, src string, dest string, file os.FileInfo) model.EntryStatus {
	if !fg.isIncluded(file.Name()) {
		return notIncluded
	} else if fg.isExcluded(file.Name()) {
		return excluded
	} else if file.ModTime().Before(fg.cfg.Download.Since) {
		return outdated
	} else if destfile, err := os.Stat(path.Join(dest, file.Name())); err == nil {
		if destfile.Size() == file.Size() {
			return alreadyDl
		}
		return sizeDiff
	} else if fg.db.HasHash(base, src, file) {
		return hashExists
	}

	return neverDl
}

func (fg *FtpGrab) fixPerms(filepath string) error {
	if runtime.GOOS == "windows" {
		return nil
	}

	fileinfo, err := os.Stat(filepath)
	if err != nil {
		return err
	}

	chmod := os.FileMode(fg.cfg.Download.ChmodFile)
	if fileinfo.IsDir() {
		chmod = os.FileMode(fg.cfg.Download.ChmodDir)
	}

	if err := os.Chmod(filepath, chmod); err != nil {
		return err
	}

	if err := os.Chown(filepath, fg.cfg.Download.UID, fg.cfg.Download.GID); err != nil {
		return err
	}

	return nil
}

func (fg *FtpGrab) isIncluded(filename string) bool {
	if len(fg.cfg.Download.Include) == 0 {
		return true
	}
	for _, include := range fg.cfg.Download.Include {
		if utl.MatchString(include, filename) {
			return true
		}
	}
	return false
}

func (fg *FtpGrab) isExcluded(filename string) bool {
	if len(fg.cfg.Download.Exclude) == 0 {
		return false
	}
	for _, exclude := range fg.cfg.Download.Exclude {
		if utl.MatchString(exclude, filename) {
			return true
		}
	}
	return false
}

func (fg *FtpGrab) isSkipped(status model.EntryStatus) bool {
	return status == alreadyDl ||
		status == hashExists ||
		status == outdated ||
		status == notIncluded ||
		status == excluded
}
