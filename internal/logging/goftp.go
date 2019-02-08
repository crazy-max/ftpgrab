package logging

import (
	"strings"

	"github.com/rs/zerolog/log"
)

type GoftpWriter struct {
	Enabled bool
}

func (w *GoftpWriter) Write(p []byte) (n int, err error) {
	if w.Enabled {
		log.Debug().Msgf("%s", strings.TrimSpace(string(p)))
	}
	return len(p), nil
}
