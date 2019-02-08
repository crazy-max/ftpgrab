package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/crazy-max/cron"
	"github.com/ftpgrab/ftpgrab/internal/app"
	"github.com/ftpgrab/ftpgrab/internal/config"
	"github.com/ftpgrab/ftpgrab/internal/logging"
	"github.com/rs/zerolog/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	fg         *app.FtpGrab
	c          *cron.Cron
	cfgfile    string
	schedule   string
	timezone   string
	logLevel   string
	logNocolor bool
	logFile    bool
	logFtp     bool
	version    = "dev"
)

func main() {
	// Parse command line
	kingpin.Flag("config", "Yaml configuration file.").Envar("CONFIG").Required().StringVar(&cfgfile)
	kingpin.Flag("schedule", "CRON expression format.").Envar("SCHEDULE").StringVar(&schedule)
	kingpin.Flag("timezone", "Timezone.").Envar("TZ").Default("UTC").StringVar(&timezone)
	kingpin.Flag("log-level", "Set log level.").Envar("LOG_LEVEL").Default("info").StringVar(&logLevel)
	kingpin.Flag("log-file", "Enable logging to file.").Envar("LOG_FILE").Default("false").BoolVar(&logFile)
	kingpin.Flag("log-nocolor", "Disable the colorized output.").Envar("LOG_NOCOLOR").Default("false").BoolVar(&logNocolor)
	kingpin.Flag("log-ftp", "Enable FTP log.").Envar("LOG_FTP").Default("false").BoolVar(&logFtp)
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version(version).Author("CrazyMax")
	kingpin.CommandLine.Name = "ftpgrab"
	kingpin.CommandLine.Help = `Grab your files from a remote FTP server easily. More info : https://ftpgrab.github.io`
	kingpin.Parse()

	// Init
	logging.Configure(logNocolor, logLevel, logFile)
	log.Info().Msgf("Starting FTPGrab %s", version)

	// Handle os signals
	channel := make(chan os.Signal)
	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-channel
		if c != nil {
			c.Stop()
		}
		fg.Close()
		log.Warn().Msgf("Caught signal %v", sig)
		os.Exit(0)
	}()

	// Load and check configuration
	cfg, err := config.Load(cfgfile, logFtp, version)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot load configuration")
	}
	if err := cfg.Check(); err != nil {
		log.Fatal().Err(err).Msg("Improper configuration")
	}

	// Init
	if fg, err = app.New(cfg); err != nil {
		log.Fatal().Err(err).Msg("Cannot initialize FTPGrab")
	}

	// Run immediately if schedule is not defined
	if schedule == "" {
		fg.Run()
		return
	}

	// Start cronjob
	c = cron.NewWithLocation(cfg.Location)
	log.Info().Msgf("Add cronjob with schedule %s", schedule)
	if err := c.AddJob(schedule, fg); err != nil {
		log.Fatal().Err(err).Msg("Cannot create cron task")
	}
	c.Start()

	for {
		runtime.Gosched()
	}
}
