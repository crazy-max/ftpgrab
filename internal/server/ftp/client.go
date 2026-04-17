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
	"github.com/crazy-max/ftpgrab/v7/internal/secret"
	"github.com/crazy-max/ftpgrab/v7/internal/server"
	"github.com/hashicorp/go-multierror"
	"github.com/jlaffaye/ftp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type ftpConn interface {
	Login(username, password string) error
	List(path string) ([]*ftp.Entry, error)
	Retr(path string) (io.ReadCloser, error)
	Quit() error
}

// Client represents an active ftp object
type Client struct {
	*server.Client
	cfg         *config.ServerFTP
	addr        string
	username    string
	password    string
	connectErr  error
	ftp         ftpConn
	dial        func(addr string, options ...ftp.DialOption) (ftpConn, error)
	dialOptions []ftp.DialOption
	pathenc     pathEncoder
}

type serverConn struct {
	*ftp.ServerConn
}

func (c serverConn) Retr(path string) (io.ReadCloser, error) {
	return c.ServerConn.Retr(path)
}

func defaultDialFTP(addr string, options ...ftp.DialOption) (ftpConn, error) {
	conn, err := ftp.Dial(addr, options...)
	if err != nil {
		return nil, err
	}
	return serverConn{ServerConn: conn}, nil
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
	client := &Client{
		cfg:  cfg,
		addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		dial: defaultDialFTP,
	}
	mode := getTLSMode(cfg)
	if client.pathenc, err = newPathEnc(cfg.PathEncoding); err != nil {
		return nil, err
	}

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
	client.dialOptions = ftpConfig

	username, err := secret.GetSecret(cfg.Username, cfg.UsernameFile)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot retrieve username secret for ftp server")
	}
	password, err := secret.GetSecret(cfg.Password, cfg.PasswordFile)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot retrieve password secret for ftp server")
	}
	client.username = username
	client.password = password

	if err = client.connect(); err != nil {
		return nil, err
	}

	return &server.Client{Handler: client}, err
}

func (c *Client) connect() error {
	dial := c.dial
	if dial == nil {
		dial = defaultDialFTP
	}
	conn, err := dial(c.addr, c.dialOptions...)
	if err != nil {
		return err
	}
	if c.username != "" {
		if err := conn.Login(c.username, c.password); err != nil {
			if closeErr := conn.Quit(); closeErr != nil {
				log.Warn().Err(closeErr).Msg("Cannot close ftp connection after login failure")
			}
			return err
		}
	}
	c.connectErr = nil
	c.ftp = conn
	return nil
}

func (c *Client) ensureConnected() error {
	if c.ftp != nil {
		return nil
	}
	if c.connectErr != nil {
		return c.connectErr
	}
	if err := c.connect(); err != nil {
		c.connectErr = err
		return err
	}
	return nil
}

func (c *Client) handleTransferError(err error) error {
	if !isTimeoutError(err) {
		return err
	}
	if closeErr := c.Close(); closeErr != nil {
		log.Warn().Err(closeErr).Msg("Cannot close ftp connection after transfer timeout")
	}
	return err
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
	if *c.cfg.EscapeRegexpMeta {
		dir = regexp.QuoteMeta(dir)
	}
	dir, err := c.pathenc.Encode(dir)
	if err != nil {
		return nil, err
	}
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}
	files, err := c.ftp.List(dir)
	if err != nil {
		return nil, c.handleTransferError(err)
	}

	var entries []os.FileInfo
	for _, file := range files {
		if file.Name == "." || file.Name == ".." {
			continue
		}
		name, err := c.pathenc.Decode(file.Name)
		if err != nil {
			return nil, err
		}
		var mode os.FileMode
		switch file.Type {
		case ftp.EntryTypeFolder:
			mode |= os.ModeDir
		case ftp.EntryTypeLink:
			mode |= os.ModeSymlink
		}
		entries = append(entries, &fileInfo{
			name:  name,
			mode:  mode,
			mtime: file.Time,
			size:  int64(file.Size),
		})
	}

	return entries, nil
}

// Retrieve file from server and write bytes to "dest".
func (c *Client) Retrieve(filename string, dest io.Writer) error {
	filename, err := c.pathenc.Encode(filename)
	if err != nil {
		return err
	}
	if err := c.ensureConnected(); err != nil {
		return err
	}
	resp, err := c.ftp.Retr(filename)
	if err != nil {
		return c.handleTransferError(err)
	}
	_, copyErr := io.Copy(dest, resp)
	closeErr := resp.Close()
	if err := multierror.Append(copyErr, closeErr).ErrorOrNil(); err != nil {
		return c.handleTransferError(err)
	}
	return nil
}

// Close closes ftp connection
func (c *Client) Close() error {
	if c.ftp == nil {
		return nil
	}
	conn := c.ftp
	c.ftp = nil
	return conn.Quit()
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

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	var timeoutErr interface {
		Timeout() bool
	}
	return errors.As(err, &timeoutErr) && timeoutErr.Timeout()
}

type pathEncoder struct {
	name     string
	encoding encoding.Encoding
}

func newPathEnc(pathEncoding string) (pathEncoder, error) {
	switch pathEncoding {
	case "", "utf-8":
		return pathEncoder{}, nil
	case "windows-1251":
		return pathEncoder{
			name:     pathEncoding,
			encoding: charmap.Windows1251,
		}, nil
	default:
		return pathEncoder{}, errors.Errorf("unsupported FTP path encoding %q", pathEncoding)
	}
}

func (c pathEncoder) Encode(value string) (string, error) {
	if c.encoding == nil {
		return value, nil
	}
	encoded, _, err := transform.String(c.encoding.NewEncoder(), value)
	if err != nil {
		return "", errors.Wrapf(err, "encode FTP path using %s", c.name)
	}
	return encoded, nil
}

func (c pathEncoder) Decode(value string) (string, error) {
	if c.encoding == nil {
		return value, nil
	}
	decoded, _, err := transform.String(c.encoding.NewDecoder(), value)
	if err != nil {
		return "", errors.Wrapf(err, "decode FTP path using %s", c.name)
	}
	return decoded, nil
}

type fileInfo struct {
	name  string
	size  int64
	mode  os.FileMode
	mtime time.Time
}

func (f *fileInfo) Name() string {
	return f.name
}

func (f *fileInfo) Size() int64 {
	return f.size
}

func (f *fileInfo) Mode() os.FileMode {
	return f.mode
}

func (f *fileInfo) ModTime() time.Time {
	return f.mtime
}

func (f *fileInfo) IsDir() bool {
	return f.mode.IsDir()
}

func (f *fileInfo) Sys() any {
	return nil
}
