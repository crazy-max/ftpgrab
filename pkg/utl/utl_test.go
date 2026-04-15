package utl

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEnv(t *testing.T) {
	t.Setenv("FTPGRAB_TEST_ENV", "value")
	assert.Equal(t, "value", GetEnv("FTPGRAB_TEST_ENV", "fallback"))
	assert.Equal(t, "fallback", GetEnv("FTPGRAB_MISSING_ENV", "fallback"))
}

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

func TestExists(t *testing.T) {
	dir := t.TempDir()
	assert.True(t, Exists(dir))
	assert.False(t, Exists(filepath.Join(dir, "missing")))
}

func TestHash(t *testing.T) {
	assert.Equal(t, "8ed3f6ad685b959ead7022518e1af76cd816f8e8ec7ccdda1ed4018e8f2223f8", Hash("alpha"))
}

func TestBasename(t *testing.T) {
	assert.Equal(t, "episode", Basename("/shows/episode.mkv"))
	assert.Equal(t, "README", Basename("/docs/README"))
}

func TestMatchString(t *testing.T) {
	assert.True(t, MatchString(`\.mkv$`, "episode.mkv"))
	assert.False(t, MatchString(`\.mkv$`, "episode.txt"))
	assert.False(t, MatchString(`(`, "episode.mkv"))
}

func TestNewValues(t *testing.T) {
	assert.False(t, *NewFalse())
	assert.True(t, *NewTrue())
	assert.Equal(t, 5*time.Second, *NewDuration(5 * time.Second))
}
