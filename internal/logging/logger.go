package logging

import (
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ilya1st/rotatewriter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Configure(nocolor bool, level string, logFile bool) {
	var err error
	var w io.Writer

	w = zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC1123,
		NoColor:    nocolor,
	}

	if logFile {
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

	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		log.Fatal().Err(err).Msgf("Unknown log level")
	} else {
		zerolog.SetGlobalLevel(logLevel)
	}
}
