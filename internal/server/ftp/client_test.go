package ftp

import (
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
)

type stubFTPConn struct {
	loginErr error
	quitErr  error
	quitCall int
}

func (c *stubFTPConn) Login(_, _ string) error {
	return c.loginErr
}

func (c *stubFTPConn) List(string) ([]*ftplib.Entry, error) {
	return nil, nil
}

func (c *stubFTPConn) Retr(string) (*ftplib.Response, error) {
	return nil, nil
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
	ln, err := net.Listen("tcp", "127.0.0.1:0")
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
