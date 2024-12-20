package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	_ "time/tzdata"

	"github.com/alecthomas/kong"
	"github.com/crazy-max/ftpgrab/v7/internal/app"
	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/logging"
	"github.com/crazy-max/ftpgrab/v7/pkg/utl"
	"github.com/rs/zerolog/log"
)

var (
	ftpgrab *app.FtpGrab
	cli     config.Cli
	version = "dev"
	meta    = config.Meta{
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
	meta.UserAgent = fmt.Sprintf("%s/%s go/%s %s", meta.ID, meta.Version, runtime.Version()[2:], strings.Title(runtime.GOOS)) //nolint:staticcheck // ignoring "SA1019: strings.Title is deprecated", as for our use we don't need full unicode support
	if meta.Hostname, err = os.Hostname(); err != nil {
		log.Fatal().Err(err).Msg("Cannot resolve hostname")
	}

	// Parse command line
	_ = kong.Parse(&cli,
		kong.Name(meta.ID),
		kong.Description(fmt.Sprintf("%s. More info: %s", meta.Desc, meta.URL)),
		kong.UsageOnError(),
		kong.Vars{
			"version": version,
		},
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))

	// Init
	logging.Configure(cli)
	log.Info().Str("version", version).Msgf("Starting %s", meta.Name)

	// Handle os signals
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, utl.SIGTERM)
	go func() {
		sig := <-channel
		ftpgrab.Close()
		log.Warn().Msgf("Caught signal %v", sig)
		os.Exit(0)
	}()

	// Load configuration
	cfg, err := config.Load(cli, meta)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot load configuration")
	}
	log.Debug().Msg(cfg.String())

	// Init
	if ftpgrab, err = app.New(cfg); err != nil {
		log.Fatal().Err(err).Msgf("Cannot initialize %s", meta.Name)
	}

	// Start
	if err = ftpgrab.Start(); err != nil {
		log.Fatal().Err(err).Msgf("Cannot start %s", meta.Name)
	}
}
