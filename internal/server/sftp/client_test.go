package sftp

import (
	"errors"
	"io"
	"testing"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	pkgsftp "github.com/pkg/sftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

type stubCloser struct {
	closeCall int
	err       error
}

func (c *stubCloser) Close() error {
	c.closeCall++
	return c.err
}

func TestClientOpenSFTP(t *testing.T) {
	origNewSFTPClient := newSFTPClient
	t.Cleanup(func() {
		newSFTPClient = origNewSFTPClient
	})

	t.Run("closes ssh connection when sftp init fails", func(t *testing.T) {
		closer := &stubCloser{}
		client := &Client{
			config: (&config.ServerSFTP{}).GetDefaults(),
			ssh:    &ssh.Client{},
			closer: closer,
		}
		newSFTPClient = func(_ *ssh.Client, _ ...pkgsftp.ClientOption) (*pkgsftp.Client, error) {
			return nil, errors.New("boom")
		}

		err := client.openSFTP()
		require.EqualError(t, err, "boom")
		assert.Equal(t, 1, closer.closeCall)
	})

	t.Run("keeps ssh connection open when sftp init succeeds", func(t *testing.T) {
		closer := &stubCloser{}
		client := &Client{
			config: (&config.ServerSFTP{}).GetDefaults(),
			ssh:    &ssh.Client{},
			closer: closer,
		}
		newSFTPClient = func(_ *ssh.Client, _ ...pkgsftp.ClientOption) (*pkgsftp.Client, error) {
			return nil, nil
		}

		require.NoError(t, client.openSFTP())
		assert.Equal(t, 0, closer.closeCall)
		assert.Nil(t, client.sftp)
	})
}

var _ io.Closer = (*stubCloser)(nil)
