package sftp

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	pkgsftp "github.com/pkg/sftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
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

func TestNewHostKeyNegotiation(t *testing.T) {
	_, edKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	edSigner, err := ssh.NewSignerFromKey(edKey)
	require.NoError(t, err)
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	rsaSigner, err := ssh.NewSignerFromKey(rsaKey)
	require.NoError(t, err)
	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	ecdsaSigner, err := ssh.NewSignerFromKey(ecdsaKey)
	require.NoError(t, err)
	cert := &ssh.Certificate{
		Key:             edSigner.PublicKey(),
		CertType:        ssh.HostCert,
		ValidPrincipals: []string{"127.0.0.1"},
		ValidBefore:     ssh.CertTimeInfinity,
	}
	require.NoError(t, cert.SignCert(rand.Reader, rsaSigner))
	certSigner, err := ssh.NewCertSigner(cert, edSigner)
	require.NoError(t, err)

	origNewSFTPClient := newSFTPClient
	t.Cleanup(func() { newSFTPClient = origNewSFTPClient })
	// Exercise New's real SSH handshake without needing an SFTP subsystem.
	newSFTPClient = func(_ *ssh.Client, _ ...pkgsftp.ClientOption) (*pkgsftp.Client, error) {
		return nil, nil
	}
	for _, tc := range []struct {
		name         string
		keys         []ssh.Signer
		trusted      ssh.PublicKey
		marker       string
		extraTrusted ssh.PublicKey
	}{
		{
			name:    "trusted Ed25519 on a multi-key server",
			keys:    []ssh.Signer{rsaSigner, edSigner},
			trusted: edSigner.PublicKey(),
		},
		{
			name:    "RSA authority signing an Ed25519 host certificate",
			keys:    []ssh.Signer{certSigner},
			trusted: rsaSigner.PublicKey(),
			marker:  "@cert-authority ",
		},
		{
			name:         "raw key and authority entries on a multi-key server",
			keys:         []ssh.Signer{ecdsaSigner, edSigner},
			trusted:      rsaSigner.PublicKey(),
			marker:       "@cert-authority ",
			extraTrusted: edSigner.PublicKey(),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
			require.NoError(t, err)
			defer listener.Close()
			serverResult := make(chan error, 1)
			go func() {
				conn, err := listener.Accept()
				if err != nil {
					serverResult <- err
					return
				}
				defer conn.Close()
				if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
					serverResult <- err
					return
				}
				serverConfig := &ssh.ServerConfig{NoClientAuth: true}
				for _, key := range tc.keys {
					serverConfig.AddHostKey(key)
				}
				_, _, _, err = ssh.NewServerConn(conn, serverConfig)
				serverResult <- err
			}()
			cfg := (&config.ServerSFTP{}).GetDefaults()
			cfg.Host = "127.0.0.1"
			cfg.Port = listener.Addr().(*net.TCPAddr).Port
			cfg.Timeout = new(2 * time.Second)
			cfg.KnownHostsFile = filepath.Join(t.TempDir(), "known_hosts")
			entry := tc.marker + knownhosts.Line([]string{knownhosts.Normalize(listener.Addr().String())}, tc.trusted)
			if tc.extraTrusted != nil {
				entry += "\n" + knownhosts.Line([]string{knownhosts.Normalize(listener.Addr().String())}, tc.extraTrusted)
			}
			require.NoError(t, os.WriteFile(cfg.KnownHostsFile, []byte(entry+"\n"), 0o600))
			client, err := New(cfg)
			if client != nil {
				_ = client.Handler.(*Client).ssh.Close()
			}
			require.NoError(t, err)
			require.NoError(t, <-serverResult)
		})
	}
}

func TestLoadKnownHosts(t *testing.T) {
	key := newTestHostKey(t)
	otherKey := newTestHostKey(t)
	for _, tc := range []struct {
		name    string
		entry   string
		host    string
		key     ssh.PublicKey
		wantErr string
	}{
		{
			name:  "trusted host",
			entry: knownhosts.Line([]string{"example.com"}, key),
			host:  "example.com:22",
			key:   key,
		},
		{
			name:  "nonstandard port",
			entry: knownhosts.Line([]string{"[example.com]:2222"}, key),
			host:  "example.com:2222",
			key:   key,
		},
		{
			name:  "hashed hostname",
			entry: knownhosts.Line([]string{knownhosts.HashHostname("example.com")}, key),
			host:  "example.com:22",
			key:   key,
		},
		{
			name:    "unknown host",
			entry:   knownhosts.Line([]string{"other.example.com"}, key),
			host:    "example.com:22",
			key:     key,
			wantErr: "knownhosts: key is unknown",
		},
		{
			name:    "wrong key",
			entry:   knownhosts.Line([]string{"example.com"}, key),
			host:    "example.com:22",
			key:     otherKey,
			wantErr: "knownhosts: key mismatch",
		},
		{
			name:    "wrong port",
			entry:   knownhosts.Line([]string{"[example.com]:2222"}, key),
			host:    "example.com:22",
			key:     key,
			wantErr: "knownhosts: key is unknown",
		},
		{
			name:    "empty file",
			host:    "example.com:22",
			key:     key,
			wantErr: "knownhosts: key is unknown",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "known_hosts")
			require.NoError(t, os.WriteFile(filename, []byte(tc.entry+"\n"), 0o600))
			db, err := loadKnownHosts(filename)
			require.NoError(t, err)
			err = db.HostKeyCallback()(tc.host, &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}, tc.key)
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, []string{ssh.KeyAlgoED25519}, db.HostKeyAlgorithms(tc.host))
		})
	}

	t.Run("certificate authority", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), "known_hosts")
		entry := "@cert-authority " + knownhosts.Line([]string{"example.com"}, key)
		require.NoError(t, os.WriteFile(filename, []byte(entry+"\n"), 0o600))
		db, err := loadKnownHosts(filename)
		require.NoError(t, err)
		assert.Equal(t, []string{ssh.CertAlgoED25519v01}, db.HostKeyAlgorithms("example.com:22"))
	})

	t.Run("default file", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", home)
		require.NoError(t, os.Mkdir(filepath.Join(home, ".ssh"), 0o700))
		filename := filepath.Join(home, ".ssh", "known_hosts")
		entry := knownhosts.Line([]string{"example.com"}, key)
		require.NoError(t, os.WriteFile(filename, []byte(entry+"\n"), 0o600))
		db, err := loadKnownHosts("")
		require.NoError(t, err)
		require.NoError(t, db.HostKeyCallback()("example.com:22", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}, key))
	})

	t.Run("missing default file is not created", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", home)
		db, err := loadKnownHosts("")
		require.ErrorContains(t, err, "load known hosts file")
		assert.Nil(t, db)
		assert.NoDirExists(t, filepath.Join(home, ".ssh"))
	})

	t.Run("no home directory", func(t *testing.T) {
		t.Setenv("HOME", "")
		t.Setenv("USERPROFILE", "")
		db, err := loadKnownHosts("")
		require.ErrorContains(t, err, "resolve user home directory")
		assert.Nil(t, db)
	})

	t.Run("missing configured file", func(t *testing.T) {
		cfg := (&config.ServerSFTP{}).GetDefaults()
		cfg.KnownHostsFile = filepath.Join(t.TempDir(), "missing")
		client, err := New(cfg)
		require.ErrorContains(t, err, "Cannot configure SFTP host key verification")
		assert.Nil(t, client)
	})

	t.Run("malformed file", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), "known_hosts")
		require.NoError(t, os.WriteFile(filename, []byte("example.com ssh-ed25519 invalid\n"), 0o600))
		db, err := loadKnownHosts(filename)
		require.ErrorContains(t, err, "load known hosts file")
		assert.Nil(t, db)
	})
}

func newTestHostKey(t *testing.T) ssh.PublicKey {
	t.Helper()
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	key, err := ssh.NewPublicKey(publicKey)
	require.NoError(t, err)
	return key
}

var _ io.Closer = (*stubCloser)(nil)
