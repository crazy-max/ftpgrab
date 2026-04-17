package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisabled(t *testing.T) {
	client, err := New(nil)
	require.NoError(t, err)

	file := stubFileInfo{name: "episode.mkv", size: 42, modTime: time.Now()}

	assert.False(t, client.Enabled())
	assert.False(t, client.HasDigest("/shows", "/shows/season1", file))
	assert.NoError(t, client.PutDigest("/shows", "/shows/season1", file))
	assert.NoError(t, client.Close())
}

func TestDigest(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ftpgrab.db")
	client, err := New(&config.Db{Path: dbPath})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = client.Close()
	})

	file := stubFileInfo{name: "episode.mkv", size: 42, modTime: time.Now()}

	assert.False(t, client.HasDigest("/shows", "/shows/season1", file))
	require.NoError(t, client.PutDigest("/shows", "/shows/season1", file))
	assert.True(t, client.HasDigest("/shows", "/shows/season1", file))
	require.NoError(t, client.Close())

	reopened, err := New(&config.Db{Path: dbPath})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = reopened.Close()
	})

	assert.True(t, reopened.HasDigest("/shows", "/shows/season1", file))
}

func TestSHA256Hex(t *testing.T) {
	assert.Equal(t, "8ed3f6ad685b959ead7022518e1af76cd816f8e8ec7ccdda1ed4018e8f2223f8", sha256Hex("alpha"))
}

type stubFileInfo struct {
	name    string
	size    int64
	modTime time.Time
}

func (f stubFileInfo) Name() string {
	return f.name
}

func (f stubFileInfo) Size() int64 {
	return f.size
}

func (f stubFileInfo) Mode() os.FileMode {
	return 0o644
}

func (f stubFileInfo) ModTime() time.Time {
	return f.modTime
}

func (f stubFileInfo) IsDir() bool {
	return false
}

func (f stubFileInfo) Sys() any {
	return nil
}
