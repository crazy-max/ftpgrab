package grabber

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/db"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/crazy-max/ftpgrab/v7/internal/server"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGrabForce(t *testing.T) {
	for _, tempFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("tempFirst=%t", tempFirst), func(t *testing.T) {
			cfg := tempFirstDownloadConfig(t.TempDir(), tempFirst)
			cfg.Force = new(true)
			dbCli, err := db.New(&config.Db{Path: filepath.Join(t.TempDir(), "history.db")})
			require.NoError(t, err)
			var calls int
			client := &Client{
				config: cfg,
				db:     dbCli,
				server: &server.Client{Handler: &stubServerHandler{
					retrieve: func(_ string, dest io.Writer) error {
						calls++
						_, err := fmt.Fprintf(dest, "data%d", calls)
						return err
					},
				}},
			}
			openTestRoot(t, client)
			if tempFirst {
				require.NoError(t, client.initTempDir())
			}
			t.Cleanup(client.Close)
			file := File{
				Base:    "/backups",
				SrcDir:  "/backups",
				DestDir: ".",
				Info:    stubFileInfo{name: "nightly.tar", size: 5, modTime: time.Now()},
			}
			dest := filepath.Join(cfg.Output, file.DestDir, file.Info.Name())
			for run := 1; run <= 3; run++ {
				if run == 3 {
					// History alone must not prevent another download either.
					require.NoError(t, os.Remove(dest))
				}
				got := client.Grab([]File{file})
				require.Equal(t, 1, got.Count.Success)
				require.Len(t, got.Entries, 1)
				assert.Equal(t, journal.EntryStatusForcedDl, got.Entries[0].Status)
				data, err := os.ReadFile(dest)
				require.NoError(t, err)
				assert.Equal(t, fmt.Sprintf("data%d", run), string(data))
				assert.True(t, dbCli.HasDigest(file.Base, file.SrcDir, file.Info))
			}
			assert.Equal(t, 3, calls)
			cfg.Force = new(false)
			assert.Equal(t, 1, client.Grab([]File{file}).Count.Skip)
			assert.Equal(t, 3, calls)
		})
	}
}

func TestGrabConcurrency(t *testing.T) {
	for _, tc := range []struct {
		name        string
		concurrency int
		wantWorkers int
	}{
		{
			name:        "sequential default",
			concurrency: 1,
			wantWorkers: 1,
		},
		{
			name:        "parallel",
			concurrency: 3,
			wantWorkers: 3,
		},
		{
			name:        "bounded by files",
			concurrency: 20,
			wantWorkers: 6,
		},
	} {
		for _, tempFirst := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/tempFirst=%t", tc.name, tempFirst), func(t *testing.T) {
				cfg := tempFirstDownloadConfig(t.TempDir(), tempFirst)
				cfg.Concurrency = tc.concurrency
				dbCli, err := db.New(&config.Db{Path: filepath.Join(t.TempDir(), "history.db")})
				require.NoError(t, err)
				gate := make(chan struct{})
				var release sync.Once
				started := make(chan struct{}, 32)
				var active, closed atomic.Int32
				newSession := func() *server.Client {
					var attempts sync.Map
					return &server.Client{Handler: &stubServerHandler{
						common: config.ServerCommon{Host: "ftp.example.com"},
						retrieve: func(src string, dest io.Writer) error {
							assert.LessOrEqual(t, active.Add(1), int32(tc.wantWorkers))
							defer active.Add(-1)
							started <- struct{}{}
							<-gate
							_, retried := attempts.LoadOrStore(src, true)
							if path.Base(src) == "file-4" || (path.Base(src) == "file-1" && !retried) {
								return errors.New("transfer failed")
							}
							_, err := io.WriteString(dest, "data")
							return err
						},
						close: func() error {
							assert.Zero(t, active.Load(), "client closed during transfer")
							closed.Add(1)
							return nil
						},
					}}
				}
				client := &Client{
					config: cfg,
					db:     dbCli,
					server: newSession(),
				}
				openTestRoot(t, client)
				if tempFirst {
					require.NoError(t, client.initTempDir())
				}
				t.Cleanup(client.Close)
				t.Cleanup(func() { release.Do(func() { close(gate) }) })
				var files []File
				for i := range 6 {
					files = append(files, File{
						Base:    "/source",
						SrcDir:  "/source",
						DestDir: ".",
						Info:    stubFileInfo{name: fmt.Sprintf("file-%d", i), size: 4, modTime: time.Now().Add(-time.Hour)},
					})
				}
				result := make(chan journal.Journal, 1)
				go func() { result <- client.Grab(files) }()
				for range tc.wantWorkers {
					select {
					case <-started:
					case <-time.After(5 * time.Second):
						t.Fatal("workers did not start overlapping transfers")
					}
				}
				release.Do(func() { close(gate) })
				var got journal.Journal
				select {
				case got = <-result:
				case <-time.After(5 * time.Second):
					t.Fatal("downloads did not finish")
				}
				assert.Zero(t, closed.Load(), "Grab must leave the server client open")
				assert.Equal(t, "ftp.example.com", got.ServerHost)
				assert.Equal(t, 5, got.Count.Success)
				assert.Equal(t, 1, got.Count.Error)
				require.Len(t, got.Entries, len(files))
				for i, file := range files {
					assert.Equal(t, path.Join(file.SrcDir, file.Info.Name()), got.Entries[i].File)
					if i == 4 {
						assert.False(t, dbCli.HasDigest(file.Base, file.SrcDir, file.Info))
						continue
					}
					data, err := os.ReadFile(filepath.Join(cfg.Output, file.DestDir, file.Info.Name()))
					require.NoError(t, err)
					assert.Equal(t, "data", string(data))
					assert.True(t, dbCli.HasDigest(file.Base, file.SrcDir, file.Info))
				}
				if tempFirst {
					var remaining []string
					require.NoError(t, filepath.WalkDir(filepath.Join(cfg.Output, client.tempdirRun), func(name string, entry os.DirEntry, err error) error {
						if err == nil && !entry.IsDir() {
							remaining = append(remaining, name)
						}
						return err
					}))
					assert.Empty(t, remaining)
				}
			})
		}
	}
}

func TestGrabHistoryConflicts(t *testing.T) {
	for _, transitive := range []bool{false, true} {
		t.Run(fmt.Sprintf("transitive=%t", transitive), func(t *testing.T) {
			cfg := tempFirstDownloadConfig(t.TempDir(), true)
			cfg.Concurrency = 3
			dbCli, err := db.New(&config.Db{Path: filepath.Join(t.TempDir(), "history.db")})
			require.NoError(t, err)
			var sources []string
			client := &Client{
				config: cfg,
				db:     dbCli,
				server: &server.Client{Handler: &stubServerHandler{
					retrieve: func(src string, dest io.Writer) error {
						sources = append(sources, src)
						_, err := io.WriteString(dest, "payload")
						return err
					},
				}},
			}
			openTestRoot(t, client)
			require.NoError(t, client.initTempDir())
			t.Cleanup(client.Close)
			var files []File
			for i, src := range []string{"/a", "/b", "/c"} {
				file := File{
					Base:    src,
					SrcDir:  src,
					DestDir: strings.TrimPrefix(src, "/"),
					Info:    stubFileInfo{name: "file", size: 7, modTime: time.Now()},
				}
				if transitive {
					file.Base = "/"
					switch i {
					case 1:
						file.DestDir = files[0].DestDir
					case 2:
						file.SrcDir = files[1].SrcDir
					}
				}
				files = append(files, file)
			}
			got := client.Grab(files)
			assert.Equal(t, []string{"/a/file"}, sources)
			assert.Equal(t, 1, got.Count.Success)
			assert.Equal(t, 2, got.Count.Skip)
			require.Len(t, got.Entries, 3)
			if transitive {
				assert.Equal(t, journal.EntryStatusAlreadyDl, got.Entries[1].Status)
			} else {
				assert.Equal(t, journal.EntryStatusDigestExists, got.Entries[1].Status)
			}
			assert.Equal(t, journal.EntryStatusDigestExists, got.Entries[2].Status)
			require.NoFileExists(t, filepath.Join(cfg.Output, files[2].DestDir, "file"))
		})
	}
}

func TestGrabDestinationConflicts(t *testing.T) {
	for _, tempFirst := range []bool{false, true} {
		for _, hideSkipped := range []bool{false, true} {
			t.Run(fmt.Sprintf("tempFirst=%t/hideSkipped=%t", tempFirst, hideSkipped), func(t *testing.T) {
				cfg := tempFirstDownloadConfig(t.TempDir(), tempFirst)
				cfg.Concurrency = 4
				cfg.HideSkipped = new(hideSkipped)
				dbCli, err := db.New(nil)
				require.NoError(t, err)
				var sources []string
				handler := &stubServerHandler{
					retrieve: func(src string, dest io.Writer) error {
						sources = append(sources, src)
						_, err := io.WriteString(dest, src)
						return err
					},
				}
				client := &Client{
					config: cfg,
					db:     dbCli,
					server: &server.Client{Handler: handler},
				}
				openTestRoot(t, client)
				if tempFirst {
					require.NoError(t, client.initTempDir())
				}
				t.Cleanup(client.Close)
				var files []File
				for _, src := range []string{"/a", "/b", "/c"} {
					files = append(files, File{
						Base:    src,
						SrcDir:  src,
						DestDir: ".",
						Info:    stubFileInfo{name: "file", size: int64(len(src + "/file")), modTime: time.Now()},
					})
				}
				// Later sources have the same size and must see the first download.
				got := client.Grab(files)
				assert.Equal(t, []string{"/a/file"}, sources)
				assert.Equal(t, 1, got.Count.Success)
				if hideSkipped {
					assert.Len(t, got.Entries, 1)
					assert.Zero(t, got.Count.Skip)
				} else {
					require.Len(t, got.Entries, 3)
					assert.Equal(t, journal.EntryStatusAlreadyDl, got.Entries[2].Status)
					assert.Equal(t, 2, got.Count.Skip)
				}
				data, err := os.ReadFile(filepath.Join(cfg.Output, "file"))
				require.NoError(t, err)
				assert.Equal(t, "/a/file", string(data))
			})
		}
	}
}

func TestCreateDownloadFileUsesSessionTempDirectory(t *testing.T) {
	destdir := t.TempDir()
	client := tempFirstClient(t, destdir)

	file, temppath, err := client.createDownloadFile(filepath.Join("shows", "episode.mkv"))
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	})

	assert.Equal(t, filepath.Join(destdir, client.tempdirRun), filepath.Dir(filepath.Dir(file.Name())))
	assert.Equal(t, filepath.Join(destdir, temppath), file.Name())
	assert.Equal(t, "episode.mkv", filepath.Base(file.Name()))
}

func TestCloseAndRemoveTempFileRemovesTempFile(t *testing.T) {
	destdir := t.TempDir()
	client := tempFirstClient(t, destdir)

	file, temppath, err := client.createDownloadFile(filepath.Join("shows", "episode.mkv"))
	require.NoError(t, err)
	require.NoError(t, client.closeAndRemoveTempFile(file, temppath))

	_, err = os.Stat(file.Name())
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestCreateDownloadFileUsesFreshFiles(t *testing.T) {
	client := tempFirstClient(t, t.TempDir())
	first, firstPath, err := client.createDownloadFile("file")
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.closeAndRemoveTempFile(first, firstPath) })
	second, secondPath, err := client.createDownloadFile("file")
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.closeAndRemoveTempFile(second, secondPath) })
	assert.NotEqual(t, firstPath, secondPath)
	firstInfo, err := first.Stat()
	require.NoError(t, err)
	secondInfo, err := second.Stat()
	require.NoError(t, err)
	assert.False(t, os.SameFile(firstInfo, secondInfo))
}

func TestListFiles(t *testing.T) {
	cases := []struct {
		source string
		dest   string
	}{
		{
			source: "/shows",
			dest:   filepath.Join("shows", "season1"),
		},
		{
			source: "/shows/.",
			dest:   filepath.Join("shows", "season1"),
		},
		{
			source: "/",
			dest:   "season1",
		},
		{
			source: "/.",
			dest:   "season1",
		},
		{
			source: "//",
			dest:   "season1",
		},
		{
			source: "/./",
			dest:   "season1",
		},
	}
	for _, tc := range cases {
		t.Run(tc.source, func(t *testing.T) {
			srcdir := path.Join(tc.source, "season1")
			client := &Client{
				config: &config.Download{
					Output:        "downloads",
					CreateBaseDir: new(true),
				},
				server: &server.Client{
					Handler: &stubServerHandler{
						common: config.ServerCommon{
							Sources: []string{tc.source},
						},
						entries: map[string][]os.FileInfo{
							tc.source: {
								stubFileInfo{name: "season1", dir: true},
							},
							srcdir: {
								stubFileInfo{name: "episode.mkv", size: 42, modTime: time.Unix(1700000000, 0)},
							},
						},
					},
				},
			}

			files := client.ListFiles()
			require.Len(t, files, 1)
			assert.Equal(t, tc.source, files[0].Base)
			assert.Equal(t, srcdir, files[0].SrcDir)
			assert.Equal(t, tc.dest, files[0].DestDir)
			assert.Equal(t, "episode.mkv", files[0].Info.Name())
			assert.Equal(t, int64(42), files[0].Info.Size())
		})
	}
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

	openTestRoot(t, client)
	cases := []struct {
		name  string
		build func(t *testing.T) File
		want  journal.EntryStatus
	}{
		{
			name: "not included",
			build: func(t *testing.T) File {
				return File{
					DestDir: filepath.FromSlash(t.Name()),
					Info:    stubFileInfo{name: "readme.txt", size: 5, modTime: now},
				}
			},
			want: journal.EntryStatusNotIncluded,
		},
		{
			name: "excluded",
			build: func(t *testing.T) File {
				return File{
					DestDir: filepath.FromSlash(t.Name()),
					Info:    stubFileInfo{name: "sample.mkv", size: 5, modTime: now},
				}
			},
			want: journal.EntryStatusExcluded,
		},
		{
			name: "outdated",
			build: func(t *testing.T) File {
				return File{
					DestDir: filepath.FromSlash(t.Name()),
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
					DestDir: filepath.FromSlash(t.Name()),
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
					DestDir: filepath.FromSlash(t.Name()),
					Info:    stubFileInfo{name: "episode.mkv", size: 4, modTime: now},
				}
			},
			want: journal.EntryStatusSizeDiff,
		},
		{
			name: "never downloaded",
			build: func(t *testing.T) File {
				return File{
					DestDir: filepath.FromSlash(t.Name()),
					Info:    stubFileInfo{name: "new.mkv", size: 4, modTime: now},
				}
			},
			want: journal.EntryStatusNeverDl,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, client.getStatus(tt.build(t)))
			cfg.Force = new(true)
			t.Cleanup(func() { cfg.Force = new(false) })
			want := tt.want
			if !want.IsSkipped() || want == journal.EntryStatusAlreadyDl {
				want = journal.EntryStatusForcedDl
			}
			assert.Equal(t, want, client.getStatus(tt.build(t)))
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
		DestDir: "downloads",
		Info:    stubFileInfo{name: "episode.mkv", size: 4, modTime: time.Now()},
	}

	require.NoError(t, dbCli.PutDigest(file.Base, file.SrcDir, file.Info))

	client := &Client{
		config: cfg,
		db:     dbCli,
	}

	openTestRoot(t, client)
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

func openTestRoot(t *testing.T, client *Client) {
	t.Helper()
	root, err := os.OpenRoot(client.config.Output)
	require.NoError(t, err)
	client.root = root
	t.Cleanup(func() { _ = root.Close() })
}

func tempFirstClient(t *testing.T, output string) *Client {
	t.Helper()
	client := &Client{config: tempFirstDownloadConfig(output, true)}
	openTestRoot(t, client)
	require.NoError(t, client.initTempDir())
	t.Cleanup(func() { _ = client.root.RemoveAll(client.tempdirRun) })
	return client
}

type stubServerHandler struct {
	common   config.ServerCommon
	entries  map[string][]os.FileInfo
	retrieve func(string, io.Writer) error
	close    func() error
}

func (h *stubServerHandler) Common() config.ServerCommon {
	return h.common
}

func (h *stubServerHandler) ReadDir(source string) ([]os.FileInfo, error) {
	return h.entries[source], nil
}

func (h *stubServerHandler) Retrieve(src string, dest io.Writer) error {
	if h.retrieve != nil {
		return h.retrieve(src, dest)
	}
	return nil
}

func (h *stubServerHandler) Close() error {
	if h.close != nil {
		return h.close()
	}
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

func TestReadFileRejectsUnsafeNames(t *testing.T) {
	for _, name := range []string{"", ".", "..", "../outside", "dir/file", "/absolute", "./file", "nested/../file", "/", `dir\file`, `..\outside`, `C:\outside`, `\\server\share`, "nul\x00byte"} {
		for _, dir := range []bool{false, true} {
			t.Run(fmt.Sprintf("%q/dir=%t", name, dir), func(t *testing.T) {
				client := &Client{}
				// An invalid directory must be rejected before contacting the server.
				assert.Empty(t, client.readFile("/source", "/source", t.TempDir(), stubFileInfo{
					name: name,
					dir:  dir,
				}))
			})
		}
	}
}

func TestValidRemoteName(t *testing.T) {
	for _, name := range []string{"episode.mkv", "a..b", ".hidden", "spaces and 日本語", ".ftpgrab-tmp-backup", "file [1](a)+.^$.txt", "café-世界.txt"} {
		t.Run(name, func(t *testing.T) {
			assert.True(t, validRemoteName(name))
		})
	}
}

func FuzzValidRemoteName(f *testing.F) {
	for _, name := range []string{"", ".", "..", "../escape", `..\escape`, "nul\x00byte", ".. ", "café-日本語.txt", ".ftpgrab-tmp-backup"} {
		f.Add(name)
	}
	f.Fuzz(func(t *testing.T, name string) {
		if !validRemoteName(name) {
			return
		}
		base := filepath.Join("downloads", "nested")
		joined := filepath.Join(base, name)
		if filepath.Dir(joined) != base || filepath.Base(joined) != name {
			t.Fatalf("accepted name %q changes the destination path to %q", name, joined)
		}
	})
}

func TestGrabConfinesDestination(t *testing.T) {
	for _, tempFirst := range []bool{false, true} {
		for _, escape := range []string{"parent", "symlink directory", "symlink file"} {
			t.Run(fmt.Sprintf("%s/tempFirst=%t", escape, tempFirst), func(t *testing.T) {
				output := t.TempDir()
				outside := t.TempDir()
				victim := filepath.Join(outside, "file")
				require.NoError(t, os.WriteFile(victim, []byte("untouched"), 0o600))
				before, err := os.Stat(victim)
				require.NoError(t, err)
				cfg := tempFirstDownloadConfig(output, tempFirst)
				cfg.Force = new(true)
				dbCli, err := db.New(nil)
				require.NoError(t, err)
				client := &Client{
					config: cfg,
					db:     dbCli,
					server: &server.Client{Handler: &stubServerHandler{
						retrieve: func(_ string, dest io.Writer) error {
							_, err := io.WriteString(dest, "overwrite")
							return err
						},
					}},
				}
				openTestRoot(t, client)
				if tempFirst {
					require.NoError(t, client.initTempDir())
				}
				t.Cleanup(client.Close)
				dir := "."
				switch escape {
				case "parent":
					dir = filepath.Join("..", filepath.Base(outside))
				case "symlink directory":
					if err := os.Symlink(outside, filepath.Join(output, "link")); err != nil {
						t.Skipf("symlinks unavailable: %v", err)
					}
					dir = "link"
				case "symlink file":
					if err := os.Symlink(victim, filepath.Join(output, "file")); err != nil {
						t.Skipf("symlinks unavailable: %v", err)
					}
				}
				got := client.Grab([]File{{
					Base:    "/source",
					SrcDir:  "/source",
					DestDir: dir,
					Info:    stubFileInfo{name: "file", size: 9, modTime: time.Now()},
				}})
				if escape == "symlink file" {
					// Rename replaces the link itself without following its target.
					assert.Equal(t, 1, got.Count.Success)
				} else {
					assert.Equal(t, 1, got.Count.Error)
				}
				data, err := os.ReadFile(victim)
				require.NoError(t, err)
				assert.Equal(t, "untouched", string(data))
				after, err := os.Stat(victim)
				require.NoError(t, err)
				assert.Equal(t, before.Mode(), after.Mode())
				assert.Equal(t, before.ModTime(), after.ModTime())
			})
		}
	}
}

func TestGrabReplacesFilesAfterSuccess(t *testing.T) {
	cases := []struct {
		name     string
		existing bool
		hardlink bool
		fail     bool
	}{
		{
			name:     "replace regular file",
			existing: true,
		},
		{
			name:     "replace hard link",
			existing: true,
			hardlink: true,
		},
		{
			name: "create new file",
		},
		{
			name:     "failed replacement",
			existing: true,
			fail:     true,
		},
		{
			name: "failed new file",
			fail: true,
		},
	}
	for _, tc := range cases {
		for _, tempFirst := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/tempFirst=%t", tc.name, tempFirst), func(t *testing.T) {
				output := t.TempDir()
				require.NoError(t, os.Mkdir(filepath.Join(output, "shows"), 0o755))
				dest := filepath.Join(output, "shows", "file")
				original := filepath.Join(t.TempDir(), "original")
				require.NoError(t, os.WriteFile(original, []byte("original"), 0o600))
				before, err := os.Stat(original)
				require.NoError(t, err)
				if tc.hardlink {
					require.NoError(t, os.Link(original, dest))
				} else if tc.existing {
					require.NoError(t, os.WriteFile(dest, []byte("original"), 0o600))
				}
				cfg := tempFirstDownloadConfig(output, tempFirst)
				cfg.Force = new(true)
				dbCli, err := db.New(&config.Db{Path: filepath.Join(t.TempDir(), "history.db")})
				require.NoError(t, err)
				client := &Client{
					config: cfg,
					db:     dbCli,
				}
				openTestRoot(t, client)
				if tempFirst {
					require.NoError(t, client.initTempDir())
				}
				assertUnpublished := func() {
					t.Helper()
					if tc.existing {
						data, err := os.ReadFile(dest)
						require.NoError(t, err)
						assert.Equal(t, "original", string(data))
					} else {
						assert.NoFileExists(t, dest)
					}
				}
				calls := 0
				client.server = &server.Client{Handler: &stubServerHandler{
					retrieve: func(_ string, writer io.Writer) error {
						calls++
						assertUnpublished()
						stage := writer.(*os.File).Name()
						parent := filepath.Dir(filepath.Dir(stage))
						if tempFirst {
							assert.Equal(t, filepath.Join(output, client.tempdirRun), parent)
						} else {
							assert.Equal(t, filepath.Join(output, "shows"), parent)
						}
						_, err := io.WriteString(writer, "download")
						require.NoError(t, err)
						assertUnpublished()
						if tc.fail {
							return io.ErrUnexpectedEOF
						}
						return nil
					},
				}}
				t.Cleanup(client.Close)
				mtime := time.Unix(1700000000, 0)
				file := File{
					Base:    "/source",
					SrcDir:  "/source",
					DestDir: "shows",
					Info:    stubFileInfo{name: "file", size: 8, modTime: mtime},
				}
				got := client.Grab([]File{file})
				if tc.fail {
					assert.Equal(t, 1, got.Count.Error)
					assert.Equal(t, cfg.Retry, calls)
					assertUnpublished()
					assert.False(t, dbCli.HasDigest(file.Base, file.SrcDir, file.Info))
				} else {
					assert.Equal(t, 1, got.Count.Success)
					assert.Equal(t, 1, calls)
					data, err := os.ReadFile(dest)
					require.NoError(t, err)
					assert.Equal(t, "download", string(data))
					info, err := os.Stat(dest)
					require.NoError(t, err)
					assert.True(t, info.Mode().IsRegular())
					assert.Equal(t, mtime, info.ModTime())
					assert.True(t, dbCli.HasDigest(file.Base, file.SrcDir, file.Info))
				}
				data, err := os.ReadFile(original)
				require.NoError(t, err)
				assert.Equal(t, "original", string(data))
				after, err := os.Stat(original)
				require.NoError(t, err)
				assert.Equal(t, before.Mode(), after.Mode())
				assert.Equal(t, before.ModTime(), after.ModTime())
				var stages []string
				require.NoError(t, filepath.WalkDir(output, func(name string, entry os.DirEntry, err error) error {
					if err == nil && strings.HasPrefix(entry.Name(), ".ftpgrab-tmp-") && name != filepath.Join(output, client.tempdirRun) {
						stages = append(stages, name)
					}
					return err
				}))
				assert.Empty(t, stages)
			})
		}
	}
}

func TestGrabKeepsDestinationDirectory(t *testing.T) {
	for _, tempFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("tempFirst=%t", tempFirst), func(t *testing.T) {
			cfg := tempFirstDownloadConfig(t.TempDir(), tempFirst)
			cfg.Force = new(true)
			dest := filepath.Join(cfg.Output, "file")
			require.NoError(t, os.Mkdir(dest, 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(dest, "keep"), []byte("original"), 0o600))
			dbCli, err := db.New(nil)
			require.NoError(t, err)
			client := &Client{
				config: cfg,
				db:     dbCli,
				server: &server.Client{Handler: &stubServerHandler{
					retrieve: func(_ string, writer io.Writer) error {
						_, err := io.WriteString(writer, "download")
						return err
					},
				}},
			}
			openTestRoot(t, client)
			if tempFirst {
				require.NoError(t, client.initTempDir())
			}
			t.Cleanup(client.Close)
			got := client.Grab([]File{
				{
					Base:    "/source",
					SrcDir:  "/source",
					DestDir: ".",
					Info:    stubFileInfo{name: "file", size: 8, modTime: time.Now()},
				},
			})
			assert.Equal(t, 1, got.Count.Error)
			data, err := os.ReadFile(filepath.Join(dest, "keep"))
			require.NoError(t, err)
			assert.Equal(t, "original", string(data))
			if tempFirst {
				entries, err := os.ReadDir(filepath.Join(cfg.Output, client.tempdirRun))
				require.NoError(t, err)
				assert.Empty(t, entries)
			} else {
				entries, err := os.ReadDir(cfg.Output)
				require.NoError(t, err)
				assert.Len(t, entries, 1)
			}
		})
	}
}

func TestCloseRemovesSessionOnce(t *testing.T) {
	cfg := tempFirstDownloadConfig(t.TempDir(), true)
	dbCli, err := db.New(nil)
	require.NoError(t, err)
	closed := 0
	client := &Client{
		config: cfg,
		db:     dbCli,
		server: &server.Client{Handler: &stubServerHandler{
			close: func() error {
				closed++
				return nil
			},
		}},
	}
	openTestRoot(t, client)
	require.NoError(t, client.initTempDir())
	stage := filepath.Join(cfg.Output, client.tempdirRun)
	require.DirExists(t, stage)
	client.Close()
	assert.NoDirExists(t, stage)
	client.Close()
	assert.Equal(t, 1, closed)
	_, err = client.root.Stat(".")
	assert.ErrorIs(t, err, os.ErrClosed)
}
