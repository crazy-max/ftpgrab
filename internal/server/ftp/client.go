package ftp

import (
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ftpgrab/ftpgrab/internal/logging"
	"github.com/ftpgrab/ftpgrab/internal/model"
	"github.com/ftpgrab/ftpgrab/internal/server"
	"github.com/secsy/goftp"
)

// Client represents an active ftp object
type Client struct {
	*server.Client
	ftp *goftp.Client
	cfg *model.FTP
}

// New creates new ftp instance
func New(config *model.FTP) (*server.Client, error) {
	var err error
	var tlsConfig *tls.Config
	var client = &Client{cfg: config}

	if config.TLS.Enable {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: config.TLS.InsecureSkipVerify,
		}
	}
	tlsMode := goftp.TLSExplicit
	if config.TLS.Implicit {
		tlsMode = goftp.TLSImplicit
	}

	if client.ftp, err = goftp.DialConfig(goftp.Config{
		User:     config.Username,
		Password: config.Password,
		Timeout:  time.Duration(config.Timeout) * time.Second,
		Logger: &logging.GoftpWriter{
			Enabled: config.LogTrace,
		},
		DisableEPSV: config.DisableEPSV,
		TLSConfig:   tlsConfig,
		TLSMode:     tlsMode,
	}, fmt.Sprintf("%s:%d", config.Host, config.Port)); err != nil {
		return nil, err
	}

	rawConn, err := client.ftp.OpenRawConn()
	if err != nil {
		return nil, err
	}

	if code, msg, err := rawConn.SendCommand("FEAT"); err != nil {
		return nil, err
	} else if code/100 != 2 {
		return nil, fmt.Errorf("server doesn't support FEAT: %d-%s", code, msg)
	}

	return &server.Client{Handler: client}, err
}

// Common return common configuration
func (c *Client) Common() model.Common {
	return model.Common{
		Host:     c.cfg.Host,
		Port:     c.cfg.Port,
		Username: c.cfg.Username,
		Password: c.cfg.Password,
		Sources:  c.cfg.Sources,
	}
}

// ReadDir fetches the contents of a directory, returning a list of os.FileInfo's
func (c *Client) ReadDir(path string) ([]os.FileInfo, error) {
	return c.ftp.ReadDir(path)
}

// Retrieve file "path" from server and write bytes to "dest".
func (c *Client) Retrieve(path string, dest io.Writer) error {
	return c.ftp.Retrieve(path, dest)
}

// Close closes ftp connection
func (c *Client) Close() error {
	return c.ftp.Close()
}
