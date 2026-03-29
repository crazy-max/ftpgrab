package server

import (
	"io"
	"net"
	"os"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
)

// Handler is a server interface
type Handler interface {
	Common() config.ServerCommon
	ReadDir(source string) ([]os.FileInfo, error)
	Retrieve(path string, dest io.Writer) error
	Close() error
}

// Client represents an active server object
type Client struct {
	Handler
}

type timeoutConn struct {
	net.Conn
	timeout time.Duration
}

// NewTimeoutConn refreshes read and write deadlines before each network
// operation so long transfers keep making progress while stalled ones time out.
func NewTimeoutConn(conn net.Conn, timeout time.Duration) net.Conn {
	if timeout <= 0 {
		return conn
	}
	return &timeoutConn{
		Conn:    conn,
		timeout: timeout,
	}
}

func (c *timeoutConn) Read(p []byte) (int, error) {
	if err := c.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
		return 0, err
	}
	return c.Conn.Read(p)
}

func (c *timeoutConn) Write(p []byte) (int, error) {
	if err := c.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
		return 0, err
	}
	return c.Conn.Write(p)
}
