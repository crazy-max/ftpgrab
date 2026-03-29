package grabber

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDownloadFileUsesSessionTempDirectory(t *testing.T) {
	destdir := t.TempDir()
	client := tempFirstClient(destdir)

	file, err := client.createDownloadFile(filepath.Join(destdir, "shows", "episode.mkv"))
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	})

	assert.Equal(t, filepath.Join(destdir, ".ftpgrab-tmp", "session", "shows"), filepath.Dir(file.Name()))
	assert.Equal(t, filepath.Join(destdir, ".ftpgrab-tmp", "session", "shows", "episode.mkv"), file.Name())
	assert.Equal(t, "episode.mkv", filepath.Base(file.Name()))
}

func TestCloseAndRemoveTempFileRemovesTempFile(t *testing.T) {
	destdir := t.TempDir()
	client := tempFirstClient(destdir)

	file, err := client.createDownloadFile(filepath.Join(destdir, "shows", "episode.mkv"))
	require.NoError(t, err)
	require.NoError(t, client.closeAndRemoveTempFile(file))

	_, err = os.Stat(file.Name())
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestTempFilePathUsesRunDirectoryAndPreservesRelativePath(t *testing.T) {
	destdir := t.TempDir()
	client := tempFirstClient(destdir)

	assert.Equal(
		t,
		filepath.Join(destdir, ".ftpgrab-tmp", "session", "shows", "episode.mkv"),
		client.tempFilePath(filepath.Join(destdir, "shows", "episode.mkv")),
	)
}

func tempFirstDownloadConfig(output string, tempFirst bool) *config.Download {
	cfg := (&config.Download{}).GetDefaults()
	cfg.Output = output
	cfg.TempFirst = &tempFirst
	return cfg
}

func tempFirstClient(output string) *Client {
	return &Client{
		config:     tempFirstDownloadConfig(output, true),
		tempdirRun: filepath.Join(output, ".ftpgrab-tmp", "session"),
	}
}
