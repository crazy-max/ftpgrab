package logging

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFtpWriterDisabled(t *testing.T) {
	var buf bytes.Buffer
	prevLogger := log.Logger
	prevLevel := zerolog.GlobalLevel()
	t.Cleanup(func() {
		log.Logger = prevLogger
		zerolog.SetGlobalLevel(prevLevel)
	})

	log.Logger = zerolog.New(&buf)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	writer := &FtpWriter{}
	n, err := writer.Write([]byte(" hello \n"))
	require.NoError(t, err)
	assert.Equal(t, len(" hello \n"), n)
	assert.Empty(t, buf.String())
}

func TestFtpWriterEnabled(t *testing.T) {
	var buf bytes.Buffer
	prevLogger := log.Logger
	prevLevel := zerolog.GlobalLevel()
	t.Cleanup(func() {
		log.Logger = prevLogger
		zerolog.SetGlobalLevel(prevLevel)
	})

	log.Logger = zerolog.New(&buf)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	writer := &FtpWriter{Enabled: true}
	n, err := writer.Write([]byte(" hello \n"))
	require.NoError(t, err)
	assert.Equal(t, len(" hello \n"), n)
	assert.Contains(t, buf.String(), `"message":"hello"`)
}
