package secret

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSecret(t *testing.T) {
	secretFile := filepath.Join(t.TempDir(), "secret.txt")
	require.NoError(t, os.WriteFile(secretFile, []byte("from-file"), 0o644))

	value, err := GetSecret("from-plain", secretFile)
	require.NoError(t, err)
	assert.Equal(t, "from-plain", value)

	value, err = GetSecret("", secretFile)
	require.NoError(t, err)
	assert.Equal(t, "from-file", value)

	value, err = GetSecret("", "")
	require.NoError(t, err)
	assert.Empty(t, value)

	_, err = GetSecret("", filepath.Join(t.TempDir(), "missing.txt"))
	require.Error(t, err)
}
