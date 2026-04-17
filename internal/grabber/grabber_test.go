package grabber

import (
	"bytes"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/db"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/crazy-max/ftpgrab/v7/internal/server"
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

func TestListFiles(t *testing.T) {
	client := &Client{
		config: &config.Download{
			Output:        "downloads",
			CreateBaseDir: new(true),
		},
		server: &server.Client{
			Handler: &stubServerHandler{
				common: config.ServerCommon{
					Sources: []string{"/shows"},
				},
				entries: map[string][]os.FileInfo{
					"/shows": {
						stubFileInfo{name: "season1", dir: true},
					},
					"/shows/season1": {
						stubFileInfo{name: "episode.mkv", size: 42, modTime: time.Unix(1700000000, 0)},
					},
				},
			},
		},
	}

	files := client.ListFiles()
	require.Len(t, files, 1)
	assert.Equal(t, "/shows", files[0].Base)
	assert.Equal(t, "/shows/season1", files[0].SrcDir)
	assert.Equal(t, path.Join("downloads", "/shows", "season1"), files[0].DestDir)
	assert.Equal(t, "episode.mkv", files[0].Info.Name())
	assert.Equal(t, int64(42), files[0].Info.Size())
}

func TestGetStatus(t *testing.T) {
	output := t.TempDir()
	now := time.Now()
	dbCli, err := db.New(nil)
	require.NoError(t, err)

	cfg := (&config.Download{}).GetDefaults()
	cfg.Output = output
	cfg.Include = []string{`\.mkv$`}
	cfg.Exclude = []string{`sample`}
	cfg.SinceTime = now.Add(-time.Hour)

	client := &Client{
		config: cfg,
		db:     dbCli,
	}

	cases := []struct {
		name  string
		build func(t *testing.T) File
		want  journal.EntryStatus
	}{
		{
			name: "not included",
			build: func(t *testing.T) File {
				return File{
					DestDir: filepath.Join(output, t.Name()),
					Info:    stubFileInfo{name: "readme.txt", size: 5, modTime: now},
				}
			},
			want: journal.EntryStatusNotIncluded,
		},
		{
			name: "excluded",
			build: func(t *testing.T) File {
				return File{
					DestDir: filepath.Join(output, t.Name()),
					Info:    stubFileInfo{name: "sample.mkv", size: 5, modTime: now},
				}
			},
			want: journal.EntryStatusExcluded,
		},
		{
			name: "outdated",
			build: func(t *testing.T) File {
				return File{
					DestDir: filepath.Join(output, t.Name()),
					Info:    stubFileInfo{name: "old.mkv", size: 5, modTime: now.Add(-2 * time.Hour)},
				}
			},
			want: journal.EntryStatusOutdated,
		},
		{
			name: "already downloaded",
			build: func(t *testing.T) File {
				destdir := filepath.Join(output, t.Name())
				require.NoError(t, os.MkdirAll(destdir, os.ModePerm))
				require.NoError(t, os.WriteFile(filepath.Join(destdir, "episode.mkv"), bytes.Repeat([]byte("a"), 4), 0o644))
				return File{
					DestDir: destdir,
					Info:    stubFileInfo{name: "episode.mkv", size: 4, modTime: now},
				}
			},
			want: journal.EntryStatusAlreadyDl,
		},
		{
			name: "size different",
			build: func(t *testing.T) File {
				destdir := filepath.Join(output, t.Name())
				require.NoError(t, os.MkdirAll(destdir, os.ModePerm))
				require.NoError(t, os.WriteFile(filepath.Join(destdir, "episode.mkv"), bytes.Repeat([]byte("a"), 3), 0o644))
				return File{
					DestDir: destdir,
					Info:    stubFileInfo{name: "episode.mkv", size: 4, modTime: now},
				}
			},
			want: journal.EntryStatusSizeDiff,
		},
		{
			name: "never downloaded",
			build: func(t *testing.T) File {
				return File{
					DestDir: filepath.Join(output, t.Name()),
					Info:    stubFileInfo{name: "new.mkv", size: 4, modTime: now},
				}
			},
			want: journal.EntryStatusNeverDl,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, client.getStatus(tt.build(t)))
		})
	}
}

func TestGetStatusDigestExists(t *testing.T) {
	dbCli, err := db.New(&config.Db{
		Path: filepath.Join(t.TempDir(), "ftpgrab.db"),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = dbCli.Close()
	})

	cfg := (&config.Download{}).GetDefaults()
	cfg.Output = t.TempDir()

	file := File{
		Base:    "/shows",
		SrcDir:  "/shows/season1",
		DestDir: filepath.Join(cfg.Output, "downloads"),
		Info:    stubFileInfo{name: "episode.mkv", size: 4, modTime: time.Now()},
	}

	require.NoError(t, dbCli.PutDigest(file.Base, file.SrcDir, file.Info))

	client := &Client{
		config: cfg,
		db:     dbCli,
	}

	assert.Equal(t, journal.EntryStatusDigestExists, client.getStatus(file))
}

func TestMatchString(t *testing.T) {
	assert.True(t, matchString(`\.mkv$`, "episode.mkv"))
	assert.False(t, matchString(`\.mkv$`, "episode.txt"))
	assert.False(t, matchString(`(`, "episode.mkv"))
}

func tempFirstDownloadConfig(output string, tempFirst bool) *config.Download {
	cfg := (&config.Download{}).GetDefaults()
	cfg.Output = output
	cfg.TempFirst = new(tempFirst)
	return cfg
}

func tempFirstClient(output string) *Client {
	return &Client{
		config:     tempFirstDownloadConfig(output, true),
		tempdirRun: filepath.Join(output, ".ftpgrab-tmp", "session"),
	}
}

type stubServerHandler struct {
	common  config.ServerCommon
	entries map[string][]os.FileInfo
}

func (h *stubServerHandler) Common() config.ServerCommon {
	return h.common
}

func (h *stubServerHandler) ReadDir(source string) ([]os.FileInfo, error) {
	return h.entries[source], nil
}

func (*stubServerHandler) Retrieve(string, io.Writer) error {
	return nil
}

func (*stubServerHandler) Close() error {
	return nil
}

type stubFileInfo struct {
	name    string
	size    int64
	modTime time.Time
	dir     bool
}

func (f stubFileInfo) Name() string {
	return f.name
}

func (f stubFileInfo) Size() int64 {
	return f.size
}

func (f stubFileInfo) Mode() os.FileMode {
	if f.dir {
		return os.ModeDir | 0o755
	}
	return 0o644
}

func (f stubFileInfo) ModTime() time.Time {
	return f.modTime
}

func (f stubFileInfo) IsDir() bool {
	return f.dir
}

func (f stubFileInfo) Sys() any {
	return nil
}
