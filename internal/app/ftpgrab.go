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
	"github.com/rs/zerolog/log"
)

// FtpGrab represents an active ftpgrab object
type FtpGrab struct {
	cfg    *config.Configuration
	srv    *server.Client
	db     *db.Client
	notif  *notif.Client
	jnl    *journal.Client
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
func New(cfg *config.Configuration) (*FtpGrab, error) {
	if err := os.MkdirAll(cfg.Download.Output, os.ModePerm); err != nil {
		return nil, fmt.Errorf("cannot create output download folder %s, %v", cfg.Download.Output, err)
	}

	if err := os.MkdirAll(path.Dir(cfg.Db.Path), os.ModePerm); err != nil {
		return nil, fmt.Errorf("cannot create database destination folder %s, %v", path.Dir(cfg.Db.Path), err)
	}

	return &FtpGrab{
		cfg: cfg,
	}, nil
}

// Run starts ftpgrab process
func (fg *FtpGrab) Run() {
	if !atomic.CompareAndSwapUint32(&fg.locker, 0, 1) {
		log.Warn().Msg("Already running")
		return
	}
	defer atomic.StoreUint32(&fg.locker, 0)

	start := time.Now()
	var err error

	defer fg.trackTime(start, "Finished, total time spent: ")
	log.Info().Msg("########")

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
		log.Info().Msg("########")
		log.Info().Msgf("Grabbing from %s", src)

		// Check basedir
		dest := fg.cfg.Download.Output
		if src != "/" && fg.cfg.Download.CreateBasedir {
			dest = path.Join(dest, src)
		}

		// Retrieve recursively
		fg.retrieveRecursive(src, src, dest)
	}

	fg.Close()
	fg.jnl.Duration = time.Since(start)
	log.Info().Msg("########")

	// Check journal before sending report
	if fg.jnl.IsEmpty() {
		log.Warn().Msg("Journal empty, skip sending report")
		return
	}

	// Send notifications
	fg.notif.Send(*fg.jnl)
}

// Close closes ftpgrab (ftp and db connection)
func (fg *FtpGrab) Close() {
	if err := fg.srv.Close(); err != nil {
		log.Warn().Err(err).Msg("Cannot close server connection")
	}
	if err := fg.db.Close(); err != nil {
		log.Warn().Err(err).Msg("Cannot close database")
	}
}

func (fg *FtpGrab) retrieveRecursive(base string, source string, dest string) {
	// Check source dir exists
	files, err := fg.srv.ReadDir(source)
	if err != nil {
		log.Error().Err(err).Msgf("Cannot read directory %s", source)
		return
	}

	for _, file := range files {
		fg.retrieve(base, source, dest, file, 0)
	}
}

func (fg *FtpGrab) retrieve(base string, src string, dest string, file os.FileInfo, retry int) {
	srcpath := path.Join(src, file.Name())
	destpath := path.Join(dest, file.Name())

	if file.IsDir() {
		fg.retrieveRecursive(base, srcpath, destpath)
		return
	}

	status := fg.fileStatus(base, src, dest, file)
	jnlEntry := model.Entry{
		File:       srcpath,
		StatusText: string(status),
	}

	if !fg.isSkipped(status) || (fg.isSkipped(status) && !fg.cfg.Download.HideSkipped) {
		log.Info().Msg("--------")
		log.Info().Msgf("Checking %s", srcpath)
		log.Info().Msg(string(status))
	}

	if status == alreadyDl && !fg.db.HasHash(base, src, file) {
		if err := fg.db.PutHash(base, src, file); err != nil {
			log.Error().Err(err).Msgf("Cannot add hash into db for %s", srcpath)
		}
	}
	if fg.isSkipped(status) {
		if !fg.cfg.Download.HideSkipped {
			log.Warn().Msgf("Skipped: %s", jnlEntry.StatusText)
			jnlEntry.StatusType = "skip"
			fg.jnl.AddEntry(jnlEntry)
		}
		return
	}

	if retry == 0 {
		defer fg.trackTime(time.Now(), "Time spent: ")
	}
	retrieveStart := time.Now()
	log.Info().Msgf("Downloading file (%s) to %s...", units.HumanSize(float64(file.Size())), destpath)

	destfolder := path.Dir(destpath)
	if err := os.MkdirAll(destfolder, os.ModePerm); err != nil {
		log.Error().Err(err).Msg("Cannot create destination dir")
		jnlEntry.StatusType = "error"
		jnlEntry.StatusText = fmt.Sprintf("Cannot create destination dir: %v", err)
		fg.jnl.AddEntry(jnlEntry)
		return
	}
	fg.fixPerms(destfolder)

	destfile, err := os.Create(destpath)
	if err != nil {
		log.Error().Err(err).Msg("Cannot create destination file")
		jnlEntry.StatusType = "error"
		jnlEntry.StatusText = fmt.Sprintf("Cannot create destination file: %v", err)
		fg.jnl.AddEntry(jnlEntry)
		return
	}

	err = fg.srv.Retrieve(srcpath, destfile)
	if err != nil {
		retry++
		log.Error().Err(err).Msgf("Error downloading, retry %d/%d", retry, fg.cfg.Download.Retry)
		if retry == fg.cfg.Download.Retry {
			log.Error().Err(err).Msg("Cannot download file")
			jnlEntry.StatusType = "error"
			jnlEntry.StatusText = fmt.Sprintf("Cannot download file: %v", err)
		} else {
			fg.retrieve(base, src, dest, file, retry)
			return
		}
	} else {
		log.Info().Msg("File successfully downloaded!")
		jnlEntry.StatusType = "success"
		jnlEntry.StatusText = fmt.Sprintf("%s successfully downloaded in %s",
			units.HumanSize(float64(file.Size())),
			durafmt.ParseShort(time.Since(retrieveStart)).String(),
		)
		fg.fixPerms(destpath)
		if err := fg.db.PutHash(base, src, file); err != nil {
			log.Error().Err(err).Msg("Cannot add hash into db")
			jnlEntry.StatusType = "warning"
			jnlEntry.StatusText = fmt.Sprintf("Successfully downloaded but cannot add hash into db: %v", err)
		}
		if err = os.Chtimes(destpath, file.ModTime(), file.ModTime()); err != nil {
			log.Warn().Err(err).Msg("Cannot change modtime of destination file")
		}
	}

	fg.jnl.AddEntry(jnlEntry)
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

func (fg *FtpGrab) fixPerms(filepath string) {
	if runtime.GOOS == "windows" {
		return
	}

	fileinfo, err := os.Stat(filepath)
	if err != nil {
		log.Warn().Err(err).Msgf("Cannot stat %s", filepath)
		return
	}

	chmod := os.FileMode(fg.cfg.Download.ChmodFile)
	if fileinfo.IsDir() {
		chmod = os.FileMode(fg.cfg.Download.ChmodDir)
	}

	if err := os.Chmod(filepath, chmod); err != nil {
		log.Warn().Err(err).Msgf("Cannot chmod %s", filepath)
	}

	if err := os.Chown(filepath, fg.cfg.Download.UID, fg.cfg.Download.GID); err != nil {
		log.Warn().Err(err).Msgf("Cannot chown %s", filepath)
	}
}

func (fg *FtpGrab) trackTime(start time.Time, prefix string) {
	log.Info().Msgf("%s%s", prefix, durafmt.ParseShort(time.Since(start)).String())
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
