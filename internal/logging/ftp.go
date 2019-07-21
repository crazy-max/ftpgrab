package logging

import (
	"strings"

	"github.com/rs/zerolog/log"
)

// FtpWriter is a ftp logger
type FtpWriter struct {
	Enabled bool
}

func (w *FtpWriter) Write(p []byte) (n int, err error) {
	if w.Enabled {
		log.Debug().Msgf("%s", strings.TrimSpace(string(p)))
	}
	return len(p), nil
}
