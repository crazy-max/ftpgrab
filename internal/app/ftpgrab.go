package app

import (
	"sync/atomic"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/grabber"
	"github.com/crazy-max/ftpgrab/v7/internal/notif"
	"github.com/hako/durafmt"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

// FtpGrab represents an active ftpgrab object
type FtpGrab struct {
	cfg     *config.Config
	cron    *cron.Cron
	notif   *notif.Client
	grabber *grabber.Client
	jobID   cron.EntryID
	locker  uint32
}

// New creates new ftpgrab instance
func New(cfg *config.Config) (*FtpGrab, error) {
	return &FtpGrab{
		cfg: cfg,
		cron: cron.New(cron.WithParser(cron.NewParser(
			cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor),
		)),
	}, nil
}

// Start starts ftpgrab
func (fg *FtpGrab) Start() error {
	var err error

	// Run on startup
	fg.Run()

	// Init scheduler if defined
	if len(fg.cfg.Cli.Schedule) == 0 {
		return nil
	}
	if fg.jobID, err = fg.cron.AddJob(fg.cfg.Cli.Schedule, fg); err != nil {
		return err
	}
	log.Info().Msgf("Cron initialized with schedule %s", fg.cfg.Cli.Schedule)

	// Start scheduler
	fg.cron.Start()
	log.Info().Msgf("Next run in %s (%s)",
		durafmt.Parse(time.Until(fg.cron.Entry(fg.jobID).Next)).LimitFirstN(2).String(),
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
			durafmt.Parse(time.Until(fg.cron.Entry(fg.jobID).Next)).LimitFirstN(2).String(),
			fg.cron.Entry(fg.jobID).Next)
	}

	start := time.Now()
	var err error

	// Notification client
	if fg.notif, err = notif.New(fg.cfg.Notif, fg.cfg.Meta); err != nil {
		log.Error().Err(err).Msg("Cannot create notifiers")
		return
	}

	// Grabber client
	if fg.grabber, err = grabber.New(fg.cfg.Download, fg.cfg.Db, fg.cfg.Server); err != nil {
		log.Error().Err(err).Msg("Cannot create grabber")
		return
	}
	defer fg.grabber.Close()

	// List files
	files := fg.grabber.ListFiles()
	if len(files) == 0 {
		log.Warn().Msg("No file found from the provided sources")
		return
	}
	log.Info().Msgf("%d file(s) found", len(files))

	// Grab
	jnl := fg.grabber.Grab(files)
	jnl.Duration = time.Since(start)
	log.Info().
		Str("duration", time.Since(start).Round(time.Millisecond).String()).
		Msg("Finished")

	// Check journal before sending report
	if jnl.IsEmpty() {
		log.Warn().Msg("Journal empty, skip sending report")
		return
	}

	// Send notifications
	fg.notif.Send(jnl)
}

// Close closes ftpgrab
func (fg *FtpGrab) Close() {
	fg.grabber.Close()
	if fg.cron != nil {
		fg.cron.Stop()
	}
}
