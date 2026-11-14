package grabber

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoteNamesOnHost(t *testing.T) {
	for _, name := range []string{".. ", ". ", "...", "file.", "file ", "NUL", "CON", "COM1", "LPT1", "CONIN$", "CONOUT$", "C:file", "file:stream"} {
		t.Run(name, func(t *testing.T) {
			assert.False(t, validRemoteName(name))
		})
	}
}
