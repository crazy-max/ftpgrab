package grabber

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/db"
	"github.com/crazy-max/ftpgrab/v7/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoteNamesOnHost(t *testing.T) {
	for _, name := range []string{".. ", ". ", "...", "file.", "file ", "NUL", "CON", "COM1", "LPT1", "CONIN$", "CONOUT$", "C:file", "file:stream"} {
		t.Run(name, func(t *testing.T) {
			assert.False(t, validRemoteName(name))
		})
	}
}

func TestGrabJunctions(t *testing.T) {
	cases := []struct {
		name     string
		asOutput bool
		inside   bool
	}{
		{
			name:     "output junction",
			asOutput: true,
		},
		{
			name:   "nested junction to inside",
			inside: true,
		},
		{
			name: "nested junction to outside",
		},
	}
	for _, tc := range cases {
		for _, tempFirst := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/tempFirst=%t", tc.name, tempFirst), func(t *testing.T) {
				workspace := t.TempDir()
				output := filepath.Join(workspace, "output")
				target := filepath.Join(workspace, "target")
				link, destdir := output, "."
				if !tc.asOutput {
					require.NoError(t, os.Mkdir(output, 0o755))
					link, destdir = filepath.Join(output, "junction"), "junction"
					if tc.inside {
						target = filepath.Join(output, "target")
					}
				}
				require.NoError(t, os.Mkdir(target, 0o755))
				victim := filepath.Join(target, "file")
				require.NoError(t, os.WriteFile(victim, []byte("original"), 0o600))
				cmd := exec.CommandContext(t.Context(), "powershell.exe", "-NoProfile", "-NonInteractive", "-Command",
					`New-Item -ItemType Junction -Path $env:FTPGRAB_TEST_JUNCTION -Target $env:FTPGRAB_TEST_TARGET -ErrorAction Stop | Out-Null`)
				cmd.Env = append(os.Environ(), "FTPGRAB_TEST_JUNCTION="+link, "FTPGRAB_TEST_TARGET="+target)
				logs, err := cmd.CombinedOutput()
				require.NoError(t, err, "%s", logs)
				t.Cleanup(func() { require.NoError(t, os.Remove(link)) })
				before, err := os.Stat(victim)
				require.NoError(t, err)
				dirBefore, err := os.Stat(target)
				require.NoError(t, err)

				cfg := tempFirstDownloadConfig(output, tempFirst)
				cfg.Force = new(true)
				dbCli, err := db.New(nil)
				require.NoError(t, err)
				calls := 0
				client := &Client{
					config: cfg,
					db:     dbCli,
					server: &server.Client{Handler: &stubServerHandler{
						retrieve: func(_ string, dest io.Writer) error {
							calls++
							_, err := io.WriteString(dest, "download")
							return err
						},
					}},
				}
				initTestOutput(t, client)
				t.Cleanup(client.Close)
				got := client.Grab([]File{
					{
						Base:    "/source",
						SrcDir:  "/source",
						DestDir: destdir,
						Info:    stubFileInfo{name: "file", size: 8, modTime: time.Unix(1700000000, 0)},
					},
				})
				client.Close()
				data, err := os.ReadFile(victim)
				require.NoError(t, err)
				if tc.asOutput {
					assert.Equal(t, 1, got.Count.Success)
					assert.Zero(t, got.Count.Error)
					assert.Equal(t, 1, calls)
					assert.Equal(t, "download", string(data))
				} else {
					assert.Equal(t, 1, got.Count.Error)
					assert.Zero(t, got.Count.Success)
					assert.Zero(t, calls)
					assert.Equal(t, "original", string(data))
					after, err := os.Stat(victim)
					require.NoError(t, err)
					assert.Equal(t, before.Mode(), after.Mode())
					assert.Equal(t, before.ModTime(), after.ModTime())
					dirAfter, err := os.Stat(target)
					require.NoError(t, err)
					assert.Equal(t, dirBefore.Mode(), dirAfter.Mode())
					assert.Equal(t, dirBefore.ModTime(), dirAfter.ModTime())
				}
				entries, err := os.ReadDir(target)
				require.NoError(t, err)
				require.Len(t, entries, 1, "no staging directories should remain in the target")
				assert.Equal(t, "file", entries[0].Name())
			})
		}
	}
}
