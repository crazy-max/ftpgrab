package ftp

import (
	"testing"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/pkg/utl"
	"github.com/stretchr/testify/assert"
)

func TestGetTLSMode(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		cfg := (&config.ServerFTP{}).GetDefaults()
		assert.Equal(t, tlsModeDisabled, getTLSMode(cfg))
	})

	t.Run("implicit", func(t *testing.T) {
		cfg := (&config.ServerFTP{}).GetDefaults()
		cfg.TLS = utl.NewTrue()
		assert.Equal(t, tlsModeImplicit, getTLSMode(cfg))
	})

	t.Run("explicit", func(t *testing.T) {
		cfg := (&config.ServerFTP{}).GetDefaults()
		cfg.ExplicitTLS = utl.NewTrue()
		assert.Equal(t, tlsModeExplicit, getTLSMode(cfg))
	})
}
