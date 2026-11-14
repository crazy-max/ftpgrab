//go:build integration

package integration

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"net/textproto"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/grabber"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/crazy-max/ftpgrab/v7/internal/server"
	ftpserver "github.com/crazy-max/ftpgrab/v7/internal/server/ftp"
	sftpserver "github.com/crazy-max/ftpgrab/v7/internal/server/sftp"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var serverNames = []string{
	"vsftpd",
	"pureftpd",
	"crazymax-pureftpd",
	"proftpd",
	"sftpgo",
	"sftp",
	"sftpgo-sftp",
}

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	if name := os.Getenv("TEST_SERVER"); name != "" {
		if !slices.Contains(serverNames, name) {
			fmt.Fprintf(os.Stderr, "unknown integration server %q\n", name)
			os.Exit(1)
		}
		serverNames = []string{name}
		if name == "sftpgo" {
			serverNames = append(serverNames, "sftpgo-sftp")
		}
	}
	os.Exit(m.Run())
}

func TestSFTPHostKeyRejection(t *testing.T) {
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	key, err := ssh.NewPublicKey(publicKey)
	require.NoError(t, err)
	for _, name := range serverNames {
		if name != "sftp" && name != "sftpgo-sftp" {
			continue
		}
		for _, tc := range []struct {
			name    string
			entry   string
			wantErr string
		}{
			{
				name:    "unknown host",
				wantErr: "knownhosts: key is unknown",
			},
			{
				name:    "wrong key",
				entry:   knownhosts.Line([]string{"sftp", "[sftpgo]:2022"}, key),
				wantErr: "knownhosts: key mismatch",
			},
		} {
			t.Run(name+"/"+tc.name, func(t *testing.T) {
				cfg := serverConfig(name).SFTP
				cfg.KnownHostsFile = filepath.Join(t.TempDir(), "known_hosts")
				require.NoError(t, os.WriteFile(cfg.KnownHostsFile, []byte(tc.entry+"\n"), 0o600))
				client, err := sftpserver.New(cfg)
				require.ErrorContains(t, err, tc.wantErr)
				assert.Nil(t, client)
			})
		}
	}
}

func TestConcurrentRetrievals(t *testing.T) {
	files := []struct {
		path    string
		content string
	}{
		{
			path:    "basic/bar.txt",
			content: "bar\n",
		},
		{
			path:    "basic/baz/baz.txt",
			content: "baz\n",
		},
	}
	for _, name := range serverNames {
		t.Run(name, func(t *testing.T) {
			cfg := serverConfig(name)
			var client *server.Client
			var source string
			var err error
			if cfg.FTP != nil {
				client, err = ftpserver.New(cfg.FTP)
				source = cfg.FTP.Sources[0]
			} else {
				client, err = sftpserver.New(cfg.SFTP)
				source = cfg.SFTP.Sources[0]
			}
			require.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, client.Close()) })

			started := make(chan struct{}, len(files))
			proceed := make(chan struct{})
			var release sync.Once
			writers := make([]transferWriter, len(files))
			errs := make([]error, len(files))
			var transfers sync.WaitGroup
			for i, file := range files {
				writers[i].started = started
				writers[i].proceed = proceed
				transfers.Go(func() {
					errs[i] = client.Retrieve(path.Join(source, file.path), &writers[i])
				})
			}
			done := make(chan struct{})
			go func() {
				transfers.Wait()
				close(done)
			}()
			t.Cleanup(func() {
				release.Do(func() { close(proceed) })
				<-done
			})

			// Both transfers must receive data before either is allowed to finish.
			// Holding the first destination also keeps its FTP session checked out.
			timer := time.NewTimer(30 * time.Second)
			defer timer.Stop()
			for range files {
				select {
				case <-started:
				case <-done:
					t.Fatalf("transfers finished before overlapping: %v", errs)
				case <-timer.C:
					t.Fatal("transfers did not receive data concurrently")
				}
			}
			release.Do(func() { close(proceed) })
			<-done
			for i, file := range files {
				require.NoError(t, errs[i], file.path)
				assert.Equal(t, file.content, writers[i].buffer.String(), file.path)
			}
		})
	}
}

type transferWriter struct {
	buffer  bytes.Buffer
	started chan<- struct{}
	proceed <-chan struct{}
	once    sync.Once
}

func (w *transferWriter) Write(p []byte) (int, error) {
	w.once.Do(func() {
		w.started <- struct{}{}
		<-w.proceed
	})
	return w.buffer.Write(p)
}

func TestConcurrentDownloads(t *testing.T) {
	expected := make(map[string]fixture)
	for name, data := range map[string]string{
		"bar.txt":             "bar\n",
		"baz/baz.txt":         "baz\n",
		"file name.txt":       "spaces\n",
		"file [1](a)+.^$.txt": "regexp\n",
		"café-世界.txt":         "unicode\n",
	} {
		expected[name] = fixture{
			size:   int64(len(data)),
			digest: fmt.Sprintf("%x", sha256.Sum256([]byte(data))),
		}
	}
	for _, name := range serverNames {
		for _, tempFirst := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/tempFirst=%t", name, tempFirst), func(t *testing.T) {
				dl := (&config.Download{}).GetDefaults()
				dl.Output = t.TempDir()
				dl.TempFirst = new(tempFirst)
				dl.Concurrency = 4
				srv := serverConfig(name)
				var sources []string
				root := "/fixtures"
				if name == "crazymax-pureftpd" || name == "sftpgo" || name == "sftpgo-sftp" {
					root = "/"
				}
				for _, dir := range []string{"basic", "path with spaces", "regexp [dir](v1)+.^$", "日本語"} {
					sources = append(sources, path.Join(root, dir))
				}
				if srv.FTP != nil {
					srv.FTP.Sources = sources
				} else {
					srv.SFTP.Sources = sources
				}
				client := openGrabber(t, dl, &config.Db{Path: filepath.Join(t.TempDir(), "history.db")}, srv)
				files := client.ListFiles()
				require.Len(t, files, len(expected))
				got := client.Grab(files)
				assert.Equal(t, len(expected), got.Count.Success)
				assert.Zero(t, got.Count.Error)
				require.Len(t, got.Entries, len(files))
				for i, file := range files {
					assert.Equal(t, path.Join(file.SrcDir, file.Info.Name()), got.Entries[i].File)
				}
				// Temporary session directories contain no files after completion.
				checkFiles(t, dl.Output, expected)
				again := client.Grab(files)
				assert.Equal(t, len(expected), again.Count.Skip)
				assert.Zero(t, again.Count.Error)
				assert.Zero(t, again.Count.Success)
			})
		}
	}
}

func TestDownloads(t *testing.T) {
	expected := fixtureFiles()
	for _, name := range serverNames {
		for _, tempFirst := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/tempFirst=%t", name, tempFirst), func(t *testing.T) {
				dl := (&config.Download{}).GetDefaults()
				dl.Output = t.TempDir()
				dl.TempFirst = new(tempFirst)
				db := &config.Db{Path: filepath.Join(t.TempDir(), "ftpgrab.db")}
				srv := serverConfig(name)
				// Closing between runs verifies history survives new connections and a reopened database.
				require.True(t, t.Run("download", func(t *testing.T) {
					client := openGrabber(t, dl, db, srv)
					files := client.ListFiles()
					require.Len(t, files, len(expected))
					result := client.Grab(files)
					assert.Equal(t, len(expected), result.Count.Success)
					assert.Zero(t, result.Count.Error)
					checkFiles(t, dl.Output, expected)
				}))
				assertNoTemporaryDirectories(t, dl.Output)
				t.Run("existing files", func(t *testing.T) {
					client := openGrabber(t, dl, db, srv)
					result := client.Grab(client.ListFiles())
					require.Len(t, result.Entries, len(expected))
					for _, entry := range result.Entries {
						assert.Equal(t, journal.EntryStatusAlreadyDl, entry.Status)
						assert.Equal(t, journal.EntryLevelSkip, entry.Level)
					}
				})
				require.NoError(t, os.Remove(filepath.Join(dl.Output, "basic", "bar.txt")))
				t.Run("persisted history", func(t *testing.T) {
					client := openGrabber(t, dl, db, srv)
					result := client.Grab(client.ListFiles())
					require.Len(t, result.Entries, len(expected))
					for _, entry := range result.Entries {
						want := journal.EntryStatusAlreadyDl
						if path.Base(entry.File) == "bar.txt" {
							want = journal.EntryStatusDigestExists
						}
						assert.Equal(t, want, entry.Status)
						assert.Equal(t, journal.EntryLevelSkip, entry.Level)
					}
					require.NoFileExists(t, filepath.Join(dl.Output, "basic", "bar.txt"))
				})
			})
		}
	}
}

func TestFilters(t *testing.T) {
	cases := []struct {
		name    string
		include []string
		exclude []string
		status  journal.EntryStatus
	}{
		{
			name:    "include",
			include: []string{`^bar\.txt$`},
			status:  journal.EntryStatusNotIncluded,
		},
		{
			name:    "exclude",
			exclude: []string{`^baz\.txt$`},
			status:  journal.EntryStatusExcluded,
		},
	}
	for _, name := range serverNames {
		for _, tc := range cases {
			t.Run(name+"/"+tc.name, func(t *testing.T) {
				dl := (&config.Download{}).GetDefaults()
				dl.Output, dl.Include, dl.Exclude = t.TempDir(), tc.include, tc.exclude
				srv := serverConfig(name)
				if srv.FTP != nil {
					srv.FTP.Sources = []string{path.Join(srv.FTP.Sources[0], "basic")}
				} else {
					srv.SFTP.Sources = []string{path.Join(srv.SFTP.Sources[0], "basic")}
				}
				client := openGrabber(t, dl, nil, srv)
				result := client.Grab(client.ListFiles())
				require.Len(t, result.Entries, 2)
				assert.Equal(t, 1, result.Count.Success)
				assert.Equal(t, 1, result.Count.Skip)
				for _, entry := range result.Entries {
					if path.Base(entry.File) == "baz.txt" {
						assert.Equal(t, tc.status, entry.Status)
					}
				}
				assert.FileExists(t, filepath.Join(dl.Output, "bar.txt"))
				assert.NoFileExists(t, filepath.Join(dl.Output, "baz", "baz.txt"))
			})
		}
	}
}

func TestInvalidCredentials(t *testing.T) {
	for _, name := range serverNames {
		t.Run(name, func(t *testing.T) {
			srv := serverConfig(name)
			if srv.FTP != nil {
				srv.FTP.Password = "invalid"
			} else {
				srv.SFTP.Password = "invalid"
			}
			dl := (&config.Download{}).GetDefaults()
			dl.Output = t.TempDir()
			client, err := grabber.New(dl, nil, srv)
			if client != nil {
				t.Cleanup(client.Close)
			}
			require.ErrorContains(t, err, "Cannot connect to server")
			if srv.FTP != nil {
				var response *textproto.Error
				require.ErrorAs(t, err, &response)
				assert.Equal(t, 530, response.Code)
			} else {
				assert.ErrorContains(t, err, "unable to authenticate")
			}
		})
	}
}

func TestCLI(t *testing.T) {
	if !slices.Contains(serverNames, "vsftpd") {
		t.Skip("CLI smoke test runs with vsftpd")
	}
	output := t.TempDir()
	cfg := fmt.Sprintf(`db:
  path: %q
server:
  ftp:
    host: vsftpd
    username: foo
    password: bar
    sources: ["/fixtures"]
download:
  output: %q
  tempFirst: true
`, filepath.Join(t.TempDir(), "ftpgrab.db"), output)
	filename := filepath.Join(t.TempDir(), "ftpgrab.yml")
	require.NoError(t, os.WriteFile(filename, []byte(cfg), 0o600))
	cmd := exec.CommandContext(t.Context(), "ftpgrab", "--config", filename)
	logs, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s", logs)
	checkFiles(t, output, fixtureFiles())
	assertNoTemporaryDirectories(t, output)
}

func serverConfig(name string) *config.Server {
	if name != "sftp" && name != "sftpgo-sftp" {
		cfg := (&config.ServerFTP{}).GetDefaults()
		cfg.Host, cfg.Username, cfg.Password = name, "foo", "bar"
		cfg.Sources = []string{"/fixtures"}
		cfg.Timeout = new(30 * time.Second)
		switch name {
		case "sftpgo":
			cfg.Port = 2121
			cfg.Sources = []string{"/"}
		case "crazymax-pureftpd":
			cfg.Port = 2100
			cfg.Sources = []string{"/"}
		}
		return &config.Server{FTP: cfg}
	}
	cfg := (&config.ServerSFTP{}).GetDefaults()
	cfg.Host, cfg.Username, cfg.Password = "sftp", "foo", "bar"
	cfg.KnownHostsFile = "/etc/ftpgrab/known_hosts"
	cfg.Sources = []string{"/fixtures"}
	if name == "sftpgo-sftp" {
		cfg.Host, cfg.Port = "sftpgo", 2022
		cfg.Sources = []string{"/"}
	}
	return &config.Server{SFTP: cfg}
}

func openGrabber(t *testing.T, dl *config.Download, db *config.Db, srv *config.Server) *grabber.Client {
	t.Helper()
	client, err := grabber.New(dl, db, srv)
	require.NoError(t, err)
	t.Cleanup(client.Close)
	return client
}

type fixture struct {
	size   int64
	digest string
}

func fixtureFiles() map[string]fixture {
	files := map[string][]byte{
		"basic/bar.txt":                            []byte("bar\n"),
		"basic/baz/baz.txt":                        []byte("baz\n"),
		"empty.bin":                                {},
		"path with spaces/file name.txt":           []byte("spaces\n"),
		"regexp [dir](v1)+.^$/file [1](a)+.^$.txt": []byte("regexp\n"),
		"quotes 'single'/file 'quoted'.txt":        []byte("single quotes\n"),
		"quotes \"double\"/file \"quoted\".txt":    []byte("double quotes\n"),
		"日本語/café-世界.txt":                          []byte("unicode\n"),
		"nested/one/two/three/data.bin":            {0, 1, 2, 255},
	}
	for i := range 5000 {
		files[fmt.Sprintf("many/file-%04d.txt", i)] = fmt.Appendf(nil, "fixture %d\n", i)
	}
	expected := make(map[string]fixture, len(files)+6)
	for name, data := range files {
		expected[name] = fixture{
			size:   int64(len(data)),
			digest: fmt.Sprintf("%x", sha256.Sum256(data)),
		}
	}
	zeroes := make([]byte, 1024*1024)
	trailer := []byte("\x01\x02\xfftail\n")
	for _, size := range []int64{16, 32, 64, 128, 256, 512} {
		bytes := size * 1024 * 1024
		hash := sha256.New()
		for remaining := bytes - int64(len(trailer)); remaining > 0; {
			n := min(remaining, int64(len(zeroes)))
			hash.Write(zeroes[:n])
			remaining -= n
		}
		hash.Write(trailer)
		expected[fmt.Sprintf("large-%dMiB.bin", size)] = fixture{
			size:   bytes,
			digest: fmt.Sprintf("%x", hash.Sum(nil)),
		}
	}
	return expected
}

func checkFiles(t *testing.T, output string, expected map[string]fixture) {
	t.Helper()
	var wantNames, gotNames []string
	for name, want := range expected {
		wantNames = append(wantNames, name)
		file, err := os.Open(filepath.Join(output, name))
		require.NoError(t, err)
		hash := sha256.New()
		size, readErr := io.Copy(hash, file)
		info, statErr := file.Stat()
		closeErr := file.Close()
		require.NoError(t, readErr, name)
		require.NoError(t, statErr, name)
		require.NoError(t, closeErr, name)
		assert.Equal(t, want.size, size, name)
		assert.Equal(t, want.digest, fmt.Sprintf("%x", hash.Sum(nil)), name)
		assert.Equal(t, time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC), info.ModTime().UTC(), name)
	}
	require.NoError(t, filepath.WalkDir(output, func(filename string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			name, err := filepath.Rel(output, filename)
			if err != nil {
				return err
			}
			gotNames = append(gotNames, filepath.ToSlash(name))
		}
		return nil
	}))
	slices.Sort(wantNames)
	slices.Sort(gotNames)
	assert.Equal(t, wantNames, gotNames)
}

func assertNoTemporaryDirectories(t *testing.T, output string) {
	t.Helper()
	require.NoError(t, filepath.WalkDir(output, func(_ string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		assert.False(t, strings.HasPrefix(entry.Name(), ".ftpgrab-tmp-"), "staging entry remains: %s", entry.Name())
		return nil
	}))
}

func TestRemoteSymlinkDownload(t *testing.T) {
	for _, name := range serverNames {
		if name != "vsftpd" && name != "sftp" {
			continue
		}
		for _, tempFirst := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/tempFirst=%t", name, tempFirst), func(t *testing.T) {
				dl := (&config.Download{}).GetDefaults()
				dl.Output = t.TempDir()
				dl.TempFirst = new(tempFirst)
				srv := serverConfig(name)
				if srv.FTP != nil {
					srv.FTP.Sources[0] = "/symlink-fixtures"
				} else {
					srv.SFTP.Sources[0] = "/symlink-fixtures"
				}
				client := openGrabber(t, dl, nil, srv)
				files := client.ListFiles()
				require.Len(t, files, 2)
				if name == "sftp" {
					// OpenSSH exposes the fixture's lstat mode. FTP servers may
					// instead report the target's type in directory listings.
					var linked bool
					for _, file := range files {
						if file.Info.Name() == "alias.txt" {
							linked = file.Info.Mode()&os.ModeSymlink != 0
						}
					}
					require.True(t, linked, "the remote fixture must actually be a symlink")
				}
				got := client.Grab(files)
				assert.Equal(t, 2, got.Count.Success)
				assert.Zero(t, got.Count.Error)
				for _, filename := range []string{"alias.txt", "original.txt"} {
					local := filepath.Join(dl.Output, filename)
					info, err := os.Lstat(local)
					require.NoError(t, err)
					assert.True(t, info.Mode().IsRegular(), "download must create a regular file")
					data, err := os.ReadFile(local)
					require.NoError(t, err)
					assert.Equal(t, "linked file\n", string(data))
					assert.Equal(t, time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC), info.ModTime().UTC())
				}
				client.Close()
				assertNoTemporaryDirectories(t, dl.Output)
			})
		}
	}
}

func TestDownloadLocalLinks(t *testing.T) {
	for _, name := range serverNames {
		for _, tempFirst := range []bool{false, true} {
			for _, link := range []string{"file symlink", "directory symlink", "relative directory symlink", "hard link"} {
				t.Run(fmt.Sprintf("%s/tempFirst=%t/%s", name, tempFirst, link), func(t *testing.T) {
					dl := (&config.Download{}).GetDefaults()
					dl.Output = t.TempDir()
					dl.TempFirst = new(tempFirst)
					dl.Force = new(true)
					outside := t.TempDir()
					victim := filepath.Join(outside, "baz.txt")
					require.NoError(t, os.WriteFile(victim, []byte("outside sentinel\n"), 0o600))
					require.NoError(t, os.Chtimes(victim, time.Unix(1600000000, 0), time.Unix(1600000000, 0)))
					before, err := os.Stat(victim)
					require.NoError(t, err)
					dirBefore, err := os.Stat(outside)
					require.NoError(t, err)
					local := filepath.Join(dl.Output, "bar.txt")
					relative, err := filepath.Rel(dl.Output, outside)
					require.NoError(t, err)
					switch link {
					case "file symlink":
						require.NoError(t, os.Symlink(filepath.Join(relative, "baz.txt"), local))
					case "directory symlink":
						require.NoError(t, os.Symlink(relative, filepath.Join(dl.Output, "baz")))
					case "relative directory symlink":
						require.NoError(t, os.Mkdir(filepath.Join(dl.Output, "inside"), 0o755))
						require.NoError(t, os.Symlink("inside", filepath.Join(dl.Output, "baz")))
					case "hard link":
						require.NoError(t, os.Link(victim, local))
					}
					srv := serverConfig(name)
					if srv.FTP != nil {
						srv.FTP.Sources[0] = path.Join(srv.FTP.Sources[0], "basic")
					} else {
						srv.SFTP.Sources[0] = path.Join(srv.SFTP.Sources[0], "basic")
					}
					client := openGrabber(t, dl, nil, srv)
					files := client.ListFiles()
					require.Len(t, files, 2)
					got := client.Grab(files)
					if link == "directory symlink" {
						assert.Equal(t, 1, got.Count.Success)
						require.Equal(t, 1, got.Count.Error)
						for _, entry := range got.Entries {
							if path.Base(entry.File) == "baz.txt" {
								assert.Equal(t, journal.EntryLevelError, entry.Level)
							}
						}
					} else {
						assert.Equal(t, 2, got.Count.Success)
						assert.Zero(t, got.Count.Error)
					}
					if link == "relative directory symlink" {
						data, err := os.ReadFile(filepath.Join(dl.Output, "inside", "baz.txt"))
						require.NoError(t, err)
						assert.Equal(t, "baz\n", string(data))
					}
					data, err := os.ReadFile(victim)
					require.NoError(t, err)
					assert.Equal(t, "outside sentinel\n", string(data))
					after, err := os.Stat(victim)
					require.NoError(t, err)
					assert.True(t, os.SameFile(before, after))
					assert.Equal(t, before.Mode(), after.Mode())
					assert.Equal(t, before.ModTime(), after.ModTime())
					dirAfter, err := os.Stat(outside)
					require.NoError(t, err)
					assert.Equal(t, dirBefore.Mode(), dirAfter.Mode())
					assert.Equal(t, dirBefore.ModTime(), dirAfter.ModTime())
					info, err := os.Lstat(local)
					require.NoError(t, err)
					assert.True(t, info.Mode().IsRegular())
					assert.False(t, os.SameFile(info, after))
					data, err = os.ReadFile(local)
					require.NoError(t, err)
					assert.Equal(t, "bar\n", string(data))
					client.Close()
					assertNoTemporaryDirectories(t, dl.Output)
				})
			}
		}
	}
}
