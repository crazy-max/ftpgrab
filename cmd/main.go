package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/crazy-max/ftpgrab/v7/internal/app"
	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/logging"
	"github.com/crazy-max/ftpgrab/v7/internal/model"
	"github.com/rs/zerolog/log"
)

var (
	ftpgrab *app.FtpGrab
	cli     model.Cli
	version = "dev"
	meta    = model.Meta{
		ID:     "ftpgrab",
		Name:   "FTPGrab",
		Desc:   "Grab your files periodically from a remote FTP or SFTP server easily",
		URL:    "https://github.com/crazy-max/ftpgrab",
		Logo:   "https://raw.githubusercontent.com/crazy-max/ftpgrab/master/.res/ftpgrab.png",
		Author: "CrazyMax",
	}
)

func main() {
	var err error
	runtime.GOMAXPROCS(runtime.NumCPU())

	meta.Version = version
	meta.UserAgent = fmt.Sprintf("%s/%s go/%s %s", meta.ID, meta.Version, runtime.Version()[2:], strings.Title(runtime.GOOS))
	if meta.Hostname, err = os.Hostname(); err != nil {
		log.Fatal().Err(err).Msg("Cannot resolve hostname")
	}

	// Parse command line
	_ = kong.Parse(&cli,
		kong.Name(meta.ID),
		kong.Description(fmt.Sprintf("%s. More info: %s", meta.Desc, meta.URL)),
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
	log.Info().Str("version", version).Msgf("Starting %s", meta.Name)

	// Handle os signals
	channel := make(chan os.Signal)
	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-channel
		ftpgrab.Close()
		log.Warn().Msgf("Caught signal %v", sig)
		os.Exit(0)
	}()

	// Load configuration
	cfg, err := config.Load(cli.Cfgfile, cli.Schedule)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot load configuration")
	}
	log.Debug().Msg(cfg.String())

	// Init
	if ftpgrab, err = app.New(meta, cfg, location); err != nil {
		log.Fatal().Err(err).Msgf("Cannot initialize %s", meta.Name)
	}

	// Start
	if err = ftpgrab.Start(); err != nil {
		log.Fatal().Err(err).Msgf("Cannot start %s", meta.Name)
	}
}
