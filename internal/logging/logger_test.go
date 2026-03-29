package logging

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewLogWriterPlainText(t *testing.T) {
	var buf bytes.Buffer

	logger := zerolog.New(newLogWriter(&buf, false, false, true))
	logger.Info().Str("foo", "bar").Msg("hello")

	got := buf.String()
	assert.Contains(t, got, "INF")
	assert.Contains(t, got, "hello")
	assert.Contains(t, got, "foo=bar")
	assert.NotContains(t, got, "{\"level\":")
	assert.NotContains(t, got, "\x1b[")
}

func TestNewLogWriterJSON(t *testing.T) {
	var buf bytes.Buffer

	logger := zerolog.New(newLogWriter(&buf, true, true, true)).With().Timestamp().Logger()
	logger.Info().Str("foo", "bar").Msg("hello")

	got := buf.String()
	assert.Contains(t, got, "{\"level\":\"info\"")
	assert.Contains(t, got, "\"foo\":\"bar\"")
	assert.Contains(t, got, "\"message\":\"hello\"")
}
