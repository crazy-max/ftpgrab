package ftp

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/logging"
	"github.com/crazy-max/ftpgrab/v7/internal/server"
	"github.com/crazy-max/ftpgrab/v7/pkg/utl"
	"github.com/jlaffaye/ftp"
	"github.com/rs/zerolog/log"
)

type ftpConn interface {
	Login(username, password string) error
	List(path string) ([]*ftp.Entry, error)
	Retr(path string) (*ftp.Response, error)
	Quit() error
}

// Client represents an active ftp object
type Client struct {
	*server.Client
	cfg *config.ServerFTP
	ftp ftpConn
}

type tlsMode uint8

const (
	tlsModeDisabled tlsMode = iota
	tlsModeImplicit
	tlsModeExplicit
)

func getTLSMode(cfg *config.ServerFTP) tlsMode {
	switch {
	case *cfg.ExplicitTLS:
		return tlsModeExplicit
	case *cfg.TLS:
		return tlsModeImplicit
	default:
		return tlsModeDisabled
	}
}

// New creates new ftp instance
func New(cfg *config.ServerFTP) (*server.Client, error) {
	var err error
	var client = &Client{cfg: cfg}
	mode := getTLSMode(cfg)

	tlsConfig := &tls.Config{
		ServerName:         cfg.Host,
		InsecureSkipVerify: *cfg.InsecureSkipVerify,
	}
	ftpConfig := []ftp.DialOption{
		ftp.DialWithDialFunc(newTimeoutDialFunc(*cfg.Timeout, mode, tlsConfig)),
		ftp.DialWithDisabledEPSV(*cfg.DisableEPSV),
		ftp.DialWithDisabledUTF8(*cfg.DisableUTF8),
		ftp.DialWithDisabledMLSD(*cfg.DisableMLSD),
		ftp.DialWithDebugOutput(&logging.FtpWriter{
			Enabled: *cfg.LogTrace,
		}),
	}

	switch mode {
	case tlsModeImplicit:
		ftpConfig = append(ftpConfig, ftp.DialWithTLS(tlsConfig))
	case tlsModeExplicit:
		ftpConfig = append(ftpConfig, ftp.DialWithExplicitTLS(tlsConfig))
	}

	if client.ftp, err = ftp.Dial(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), ftpConfig...); err != nil {
		return nil, err
	}

	username, err := utl.GetSecret(cfg.Username, cfg.UsernameFile)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot retrieve username secret for ftp server")
	}
	password, err := utl.GetSecret(cfg.Password, cfg.PasswordFile)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot retrieve password secret for ftp server")
	}

	if err = client.login(username, password); err != nil {
		return nil, err
	}

	return &server.Client{Handler: client}, err
}

func (c *Client) login(username, password string) error {
	if len(username) == 0 {
		return nil
	}
	if err := c.ftp.Login(username, password); err != nil {
		if closeErr := c.ftp.Quit(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("Cannot close ftp connection after login failure")
		}
		return err
	}
	return nil
}

// Common return common configuration
func (c *Client) Common() config.ServerCommon {
	return config.ServerCommon{
		Host:    c.cfg.Host,
		Port:    c.cfg.Port,
		Sources: c.cfg.Sources,
	}
}

// ReadDir fetches the contents of a directory, returning a list of os.FileInfo's
func (c *Client) ReadDir(dir string) ([]os.FileInfo, error) {
	var files []*ftp.Entry

	if *c.cfg.EscapeRegexpMeta {
		dir = regexp.QuoteMeta(dir)
	}
	files, err := c.ftp.List(dir)
	if err != nil {
		return nil, err
	}

	var entries []os.FileInfo
	for _, file := range files {
		if file.Name == "." || file.Name == ".." {
			continue
		}
		var mode os.FileMode
		switch file.Type {
		case ftp.EntryTypeFolder:
			mode |= os.ModeDir
		case ftp.EntryTypeLink:
			mode |= os.ModeSymlink
		}
		fileInfo := &fileInfo{
			name:  file.Name,
			mode:  mode,
			mtime: file.Time,
			size:  int64(file.Size),
		}
		entries = append(entries, fileInfo)
	}

	return entries, nil
}

// Retrieve file "path" from server and write bytes to "dest".
func (c *Client) Retrieve(path string, dest io.Writer) error {
	resp, err := c.ftp.Retr(path)
	if err != nil {
		return err
	}
	defer resp.Close()

	_, err = io.Copy(dest, resp)
	return err
}

// Close closes ftp connection
func (c *Client) Close() error {
	done := make(chan error, 1)
	go func() {
		done <- c.ftp.Quit()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(*c.cfg.Timeout):
		log.Warn().Dur("timeout", *c.cfg.Timeout).Msg("Cannot close ftp connection, quit timed out")
		return fmt.Errorf("quit timed out after %s", *c.cfg.Timeout)
	}
}

func newTimeoutDialFunc(timeout time.Duration, mode tlsMode, tlsConfig *tls.Config) func(network, address string) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout}
	useTLS := mode == tlsModeImplicit
	return func(network, address string) (net.Conn, error) {
		conn, err := dialer.Dial(network, address)
		if err != nil {
			return nil, err
		}
		conn = server.NewTimeoutConn(conn, timeout)
		if useTLS {
			return tls.Client(conn, tlsConfig), nil
		}
		useTLS = mode == tlsModeExplicit
		return conn, nil
	}
}
