package ftp

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/pkg/utl"
	ftplib "github.com/jlaffaye/ftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type stubFTPConn struct {
	listEntries []*ftplib.Entry
	listErr     error
	listPath    string
	loginErr    error
	quitErr     error
	quitCall    int
	retrErr     error
	retrPath    string
}

func (c *stubFTPConn) Login(_, _ string) error {
	return c.loginErr
}

func (c *stubFTPConn) List(path string) ([]*ftplib.Entry, error) {
	c.listPath = path
	return c.listEntries, c.listErr
}

func (c *stubFTPConn) Retr(path string) (*ftplib.Response, error) {
	c.retrPath = path
	return nil, c.retrErr
}

func (c *stubFTPConn) Quit() error {
	c.quitCall++
	return c.quitErr
}

func TestGetTLSMode(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		cfg := (&config.ServerFTP{}).GetDefaults()
		assert.Equal(t, tlsModeDisabled, getTLSMode(cfg))
	})

	t.Run("implicit", func(t *testing.T) {
		cfg := (&config.ServerFTP{}).GetDefaults()
		cfg.TLS = utl.NewTrue()
		assert.Equal(t, tlsModeImplicit, getTLSMode(cfg))
	})

	t.Run("explicit", func(t *testing.T) {
		cfg := (&config.ServerFTP{}).GetDefaults()
		cfg.ExplicitTLS = utl.NewTrue()
		assert.Equal(t, tlsModeExplicit, getTLSMode(cfg))
	})
}

func TestNewTimeoutDialFunc(t *testing.T) {
	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = ln.Close()
	})

	accepted := make(chan net.Conn, 4)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				close(accepted)
				return
			}
			accepted <- conn
		}
	}()

	tlsConfig := &tls.Config{InsecureSkipVerify: true}

	t.Run("disabled", func(t *testing.T) {
		dial := newTimeoutDialFunc(50*time.Millisecond, tlsModeDisabled, tlsConfig)
		conn, err := dial("tcp", ln.Addr().String())
		require.NoError(t, err)
		t.Cleanup(func() {
			_ = conn.Close()
		})

		_, isTLS := conn.(*tls.Conn)
		assert.False(t, isTLS)

		serverConn := <-accepted
		t.Cleanup(func() {
			_ = serverConn.Close()
		})
	})

	t.Run("implicit", func(t *testing.T) {
		dial := newTimeoutDialFunc(50*time.Millisecond, tlsModeImplicit, tlsConfig)
		conn, err := dial("tcp", ln.Addr().String())
		require.NoError(t, err)
		t.Cleanup(func() {
			_ = conn.Close()
		})

		_, isTLS := conn.(*tls.Conn)
		assert.True(t, isTLS)

		serverConn := <-accepted
		t.Cleanup(func() {
			_ = serverConn.Close()
		})
	})

	t.Run("explicit", func(t *testing.T) {
		dial := newTimeoutDialFunc(50*time.Millisecond, tlsModeExplicit, tlsConfig)

		controlConn, err := dial("tcp", ln.Addr().String())
		require.NoError(t, err)
		t.Cleanup(func() {
			_ = controlConn.Close()
		})
		_, isTLS := controlConn.(*tls.Conn)
		assert.False(t, isTLS)

		serverControlConn := <-accepted
		t.Cleanup(func() {
			_ = serverControlConn.Close()
		})

		dataConn, err := dial("tcp", ln.Addr().String())
		require.NoError(t, err)
		t.Cleanup(func() {
			_ = dataConn.Close()
		})
		_, isTLS = dataConn.(*tls.Conn)
		assert.True(t, isTLS)

		serverDataConn := <-accepted
		t.Cleanup(func() {
			_ = serverDataConn.Close()
		})
	})
}

func TestClientLogin(t *testing.T) {
	t.Run("skips login without username", func(t *testing.T) {
		conn := &stubFTPConn{}
		client := &Client{ftp: conn}
		require.NoError(t, client.login("", "secret"))
		assert.Equal(t, 0, conn.quitCall)
	})

	t.Run("closes connection when login fails", func(t *testing.T) {
		conn := &stubFTPConn{loginErr: errors.New("boom")}
		client := &Client{ftp: conn}
		err := client.login("user", "secret")
		require.EqualError(t, err, "boom")
		assert.Equal(t, 1, conn.quitCall)
	})

	t.Run("does not close connection on successful login", func(t *testing.T) {
		conn := &stubFTPConn{}
		client := &Client{ftp: conn}
		require.NoError(t, client.login("user", "secret"))
		assert.Equal(t, 0, conn.quitCall)
	})
}

func TestNewPathEnc(t *testing.T) {
	t.Run("normalizes utf8", func(t *testing.T) {
		pathenc, err := newPathEnc("utf-8")
		require.NoError(t, err)

		value, err := pathenc.Encode("/Проекты")
		require.NoError(t, err)
		assert.Equal(t, "/Проекты", value)
	})

	t.Run("supports windows1251", func(t *testing.T) {
		pathenc, err := newPathEnc("windows-1251")
		require.NoError(t, err)

		encoded, err := pathenc.Encode("/Проекты")
		require.NoError(t, err)
		assert.NotEqual(t, "/Проекты", encoded)

		decoded, err := pathenc.Decode(encoded)
		require.NoError(t, err)
		assert.Equal(t, "/Проекты", decoded)
	})
}

func TestReadDirDecodesEntryNames(t *testing.T) {
	pathenc, err := newPathEnc("windows-1251")
	require.NoError(t, err)

	conn := &stubFTPConn{
		listEntries: []*ftplib.Entry{
			{
				Name: mustEncodeWindows1251(t, "Инструкция_по_заливке+договор"),
				Type: ftplib.EntryTypeFolder,
			},
		},
	}
	client := &Client{
		cfg:     &config.ServerFTP{EscapeRegexpMeta: utl.NewFalse()},
		ftp:     conn,
		pathenc: pathenc,
	}

	entries, err := client.ReadDir("/Проекты")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "/\xcf\xf0\xee\xe5\xea\xf2\xfb", conn.listPath)
	assert.Equal(t, "Инструкция_по_заливке+договор", entries[0].Name())
	assert.True(t, entries[0].IsDir())
}

func TestRetrieveEncodesPath(t *testing.T) {
	pathenc, err := newPathEnc("windows-1251")
	require.NoError(t, err)

	conn := &stubFTPConn{retrErr: errors.New("stop")}
	client := &Client{
		ftp:     conn,
		pathenc: pathenc,
	}

	err = client.Retrieve("/Проекты/Инструкция.txt", nil)
	require.EqualError(t, err, "stop")
	assert.Equal(t, mustEncodeWindows1251(t, "/Проекты/Инструкция.txt"), conn.retrPath)
}

func mustEncodeWindows1251(t *testing.T, value string) string {
	t.Helper()

	encoded, _, err := transform.String(charmap.Windows1251.NewEncoder(), value)
	require.NoError(t, err)
	return encoded
}
