package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/ftpgrab/ftpgrab/internal/app"
	"github.com/ftpgrab/ftpgrab/internal/config"
	"github.com/ftpgrab/ftpgrab/internal/logging"
	"github.com/ftpgrab/ftpgrab/internal/model"
	"github.com/rs/zerolog/log"
)

var (
	ftpgrab *app.FtpGrab
	cli     model.Cli
	version = "dev"
)

func main() {
	// Parse command line
	_ = kong.Parse(&cli,
		kong.Name("ftpgrab"),
		kong.Description(`Grab your files periodically from a remote FTP or SFTP server easily. More info: https://ftpgrab.github.io`),
		kong.UsageOnError(),
		kong.Vars{
			"version": fmt.Sprintf("%s", version),
		},
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))

	// Load timezone location
	location, err := time.LoadLocation(cli.Timezone)
	if err != nil {
		log.Panic().Err(err).Msgf("Cannot load timezone %s", cli.Timezone)
	}

	// Init
	logging.Configure(&cli, location)
	log.Info().Msgf("Starting FTPGrab %s", version)

	// Handle os signals
	channel := make(chan os.Signal)
	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-channel
		ftpgrab.Close()
		log.Warn().Msgf("Caught signal %v", sig)
		os.Exit(0)
	}()

	// Load and check configuration
	cfg, err := config.Load(cli, version)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot load configuration")
	}
	if err := cfg.Check(); err != nil {
		cfg.Display()
		log.Fatal().Err(err).Msg("Improper configuration")
	}
	cfg.Display()

	// Init
	if ftpgrab, err = app.New(cfg, location); err != nil {
		log.Fatal().Err(err).Msg("Cannot initialize FTPGrab")
	}

	// Start
	if err = ftpgrab.Start(); err != nil {
		log.Fatal().Err(err).Msg("Cannot start FTPGrab")
	}
}
