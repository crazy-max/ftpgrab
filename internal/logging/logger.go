package logging

import (
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ftpgrab/ftpgrab/internal/model"
	"github.com/ilya1st/rotatewriter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Configure configures logger
func Configure(fl *model.Flags) {
	var err error
	var w io.Writer

	w = zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC1123,
		NoColor:    fl.LogNocolor,
	}

	if fl.LogFile {
		if err := os.MkdirAll("./logs", os.ModePerm); err != nil {
			log.Fatal().Err(err).Msgf("Cannot create log folder")
		}
		rwriter, err := rotatewriter.NewRotateWriter("./logs/ftpgrab.log", 5)
		if err != nil {
			log.Fatal().Err(err).Msgf("Cannot create log file writer")
		}
		sighupChan := make(chan os.Signal, 1)
		signal.Notify(sighupChan, syscall.SIGHUP)
		go func() {
			for {
				_, ok := <-sighupChan
				if !ok {
					return
				}
				rwriter.Rotate(nil)
			}
		}()
		w = zerolog.MultiLevelWriter(w, rwriter)
	}

	log.Logger = zerolog.New(w).With().Timestamp().Logger()

	logLevel, err := zerolog.ParseLevel(fl.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msgf("Unknown log level")
	} else {
		zerolog.SetGlobalLevel(logLevel)
	}
}
