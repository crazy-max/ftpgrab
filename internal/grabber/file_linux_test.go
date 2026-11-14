package grabber

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/db"
	"github.com/crazy-max/ftpgrab/v7/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func TestGrabReplacesFIFO(t *testing.T) {
	for _, tempFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("tempFirst=%t", tempFirst), func(t *testing.T) {
			cfg := tempFirstDownloadConfig(t.TempDir(), tempFirst)
			cfg.Force = new(true)
			dest := filepath.Join(cfg.Output, "file")
			require.NoError(t, unix.Mkfifo(dest, 0o600))
			// Keep a nonblocking reader so a regression writes into the FIFO
			// instead of hanging the test while opening or writing it.
			fd, err := unix.Open(dest, unix.O_RDWR|unix.O_NONBLOCK, 0)
			require.NoError(t, err)
			t.Cleanup(func() { _ = unix.Close(fd) })
			dbCli, err := db.New(nil)
			require.NoError(t, err)
			client := &Client{
				config: cfg,
				db:     dbCli,
				server: &server.Client{Handler: &stubServerHandler{
					retrieve: func(_ string, writer io.Writer) error {
						_, err := io.WriteString(writer, "data")
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
					Info:    stubFileInfo{name: "file", size: 4, modTime: time.Now()},
				},
			})
			assert.Equal(t, 1, got.Count.Success)
			info, err := os.Stat(dest)
			require.NoError(t, err)
			assert.True(t, info.Mode().IsRegular())
			assert.Equal(t, cfg.ChmodFile, info.Mode().Perm())
			var buffer [4]byte
			_, err = unix.Read(fd, buffer[:])
			assert.ErrorIs(t, err, unix.EAGAIN, "the original FIFO must receive no bytes")
		})
	}
}
