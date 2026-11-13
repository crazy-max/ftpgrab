//go:build integration

package integration

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/textproto"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/grabber"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				require.NoDirExists(t, filepath.Join(dl.Output, ".ftpgrab-tmp"))
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
	assert.NoDirExists(t, filepath.Join(output, ".ftpgrab-tmp"))
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
