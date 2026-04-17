package logging

import (
	"io"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/ilya1st/rotatewriter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func newLogWriter(out io.Writer, logJSON bool, logTimestamp bool, noColor bool) io.Writer {
	if logJSON {
		return out
	}

	var excludeParts []string
	if !logTimestamp {
		excludeParts = []string{
			zerolog.TimestampFieldName,
		}
	}

	return zerolog.ConsoleWriter{
		Out:          out,
		NoColor:      noColor,
		TimeFormat:   time.RFC1123,
		PartsExclude: excludeParts,
	}
}

// Configure configures logger
func Configure(cli config.Cli) {
	var err error
	w := newLogWriter(os.Stdout, cli.LogJSON, cli.LogTimestamp, false)

	if len(cli.LogFile) > 0 {
		logFile := path.Clean(cli.LogFile)
		if err := os.MkdirAll(path.Dir(logFile), os.ModePerm); err != nil {
			log.Fatal().Err(err).Msgf("Cannot create log folder")
		}
		rwriter, err := rotatewriter.NewRotateWriter(logFile, 5)
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
				if err := rwriter.Rotate(nil); err != nil {
					log.Error().Err(err).Msgf("Cannot rotate log")
				}
			}
		}()
		w = zerolog.MultiLevelWriter(w, newLogWriter(rwriter, cli.LogJSON, cli.LogTimestamp, true))
	}

	log.Logger = zerolog.New(w)
	if cli.LogCaller {
		log.Logger = log.Logger.With().Caller().Logger()
	}
	if cli.LogTimestamp {
		log.Logger = log.Logger.With().Timestamp().Logger()
	}

	logLevel, err := zerolog.ParseLevel(cli.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msgf("Unknown log level")
	} else {
		zerolog.SetGlobalLevel(logLevel)
	}
}
