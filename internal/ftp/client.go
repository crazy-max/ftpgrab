package ftp

import (
	"crypto/tls"
	"fmt"
	"io"
	"time"

	"github.com/ftpgrab/ftpgrab/internal/config"
	"github.com/secsy/goftp"
)

// Client represents an active ftp object
type Client struct {
	*goftp.Client
}

// New creates new ftp instance
func New(config *config.Server, logger io.Writer) (*Client, error) {
	var err error
	var client *goftp.Client
	var tlsConfig *tls.Config

	if config.TLS.Enable {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: config.TLS.InsecureSkipVerify,
		}
	}
	tlsMode := goftp.TLSExplicit
	if config.TLS.Implicit {
		tlsMode = goftp.TLSImplicit
	}

	if client, err = goftp.DialConfig(goftp.Config{
		User:               config.Username,
		Password:           config.Password,
		ConnectionsPerHost: config.ConnectionsPerHost,
		Timeout:            time.Duration(config.Timeout) * time.Second,
		Logger:             logger,
		DisableEPSV:        config.DisableEPSV,
		TLSConfig:          tlsConfig,
		TLSMode:            tlsMode,
	}, fmt.Sprintf("%s:%d", config.Host, config.Port)); err != nil {
		return nil, err
	}

	rawConn, err := client.OpenRawConn()
	if err != nil {
		return nil, err
	}

	if code, msg, err := rawConn.SendCommand("FEAT"); err != nil {
		return nil, err
	} else if code/100 != 2 {
		return nil, fmt.Errorf("server doesn't support FEAT: %d-%s", code, msg)
	}

	return &Client{client}, err
}

// Close closes ftp connection
func (c *Client) Close() error {
	return c.Client.Close()
}
