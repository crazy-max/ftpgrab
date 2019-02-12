package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/crazy-max/cron"
	"github.com/ftpgrab/ftpgrab/internal/app"
	"github.com/ftpgrab/ftpgrab/internal/config"
	"github.com/ftpgrab/ftpgrab/internal/logging"
	"github.com/ftpgrab/ftpgrab/internal/model"
	"github.com/rs/zerolog/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	ftpgrab *app.FtpGrab
	flags   model.Flags
	c       *cron.Cron
	version = "dev"
)

func main() {
	// Parse command line
	kingpin.Flag("config", "FTPGrab configuration file.").Envar("CONFIG").Required().StringVar(&flags.Cfgfile)
	kingpin.Flag("schedule", "CRON expression format.").Envar("SCHEDULE").StringVar(&flags.Schedule)
	kingpin.Flag("timezone", "Timezone assigned to FTPGrab.").Envar("TZ").Default("UTC").StringVar(&flags.Timezone)
	kingpin.Flag("log-level", "Set log level.").Envar("LOG_LEVEL").Default("info").StringVar(&flags.LogLevel)
	kingpin.Flag("log-json", "Enable JSON logging output.").Envar("LOG_JSON").Default("false").BoolVar(&flags.LogJson)
	kingpin.Flag("log-file", "Add logging to a specific file.").Envar("LOG_FILE").StringVar(&flags.LogFile)
	kingpin.Flag("log-ftp", "Enable low-level FTP log.").Envar("LOG_FTP").Default("false").BoolVar(&flags.LogFtp)
	kingpin.Flag("docker", "Enable Docker mode.").Envar("DOCKER").Default("false").BoolVar(&flags.Docker)
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version(version).Author("CrazyMax")
	kingpin.CommandLine.Name = "ftpgrab"
	kingpin.CommandLine.Help = `Grab your periodically files from a remote FTP server easily. More info on https://ftpgrab.github.io`
	kingpin.Parse()

	// Load timezone location
	location, err := time.LoadLocation(flags.Timezone)
	if err != nil {
		log.Panic().Err(err).Msgf("Cannot load timezone %s", flags.Timezone)
	}

	// Init
	logging.Configure(&flags, location)
	log.Info().Msgf("Starting FTPGrab %s", version)

	// Handle os signals
	channel := make(chan os.Signal)
	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-channel
		if c != nil {
			c.Stop()
		}
		ftpgrab.Close()
		log.Warn().Msgf("Caught signal %v", sig)
		os.Exit(0)
	}()

	// Load and check configuration
	cfg, err := config.Load(flags, version)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot load configuration")
	}
	if err := cfg.Check(); err != nil {
		log.Fatal().Err(err).Msg("Improper configuration")
	}
	cfg.Display()

	// Init
	if ftpgrab, err = app.New(cfg); err != nil {
		log.Fatal().Err(err).Msg("Cannot initialize FTPGrab")
	}

	// Run immediately if schedule is not defined
	if flags.Schedule == "" {
		ftpgrab.Run()
		return
	}

	// Start cronjob
	c = cron.NewWithLocation(location)
	log.Info().Msgf("Add cronjob with schedule %s", flags.Schedule)
	if err := c.AddJob(flags.Schedule, ftpgrab); err != nil {
		log.Fatal().Err(err).Msg("Cannot create cron task")
	}
	c.Start()

	for {
		runtime.Gosched()
	}
}
