package ftp

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"os"
	"strings"
	"syscall"
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

type netDialFunc func(network, address string) (net.Conn, error)

// Client represents an active ftp object
type Client struct {
	*server.Client
	cfg         *config.ServerFTP
	addr        string
	username    string
	password    string
	connectErr  error
	ftp         ftpConn
	dial        func(addr string, netDial netDialFunc, options ...ftp.DialOption) (ftpConn, error)
	newDialFunc func() netDialFunc
	dialOptions []ftp.DialOption
	pathenc     pathEncoder
}

type serverConn struct {
	*ftp.ServerConn
}

var errRestoreDir = errors.New("cannot restore FTP working directory")

func (c serverConn) List(dir string) ([]*ftp.Entry, error) {
	if dir == "" {
		return c.ServerConn.List("")
	}
	cwd, err := c.CurrentDir()
	if err != nil {
		return nil, err
	}
	// CWD takes a literal pathname. LIST may interpret it as a glob, while
	// escaping glob characters changes the pathname passed to MLSD.
	if err := c.ChangeDir(dir); err != nil {
		return nil, err
	}
	entries, listErr := c.ServerConn.List("")
	// CurrentDir leaves RFC 959 doubled quotes in the PWD response escaped.
	if err := c.ChangeDir(strings.ReplaceAll(cwd, `""`, `"`)); err != nil {
		restoreErr := errors.Wrap(errRestoreDir, err.Error())
		if listErr != nil {
			return nil, multierror.Append(listErr, restoreErr)
		}
		return nil, restoreErr
	}
	return entries, listErr
}

func (c serverConn) Retr(path string) (io.ReadCloser, error) {
	return c.ServerConn.Retr(path)
}

func defaultDialFTP(addr string, netDial netDialFunc, options ...ftp.DialOption) (ftpConn, error) {
	if netDial != nil {
		options = append([]ftp.DialOption{ftp.DialWithDialFunc(netDial)}, options...)
	}
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
	client.newDialFunc = func() netDialFunc {
		return newTimeoutDialFunc(*cfg.Timeout, mode, tlsConfig)
	}
	ftpConfig := []ftp.DialOption{
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
	var netDial netDialFunc
	if c.newDialFunc != nil {
		netDial = c.newDialFunc()
	}
	conn, err := dial(c.addr, netDial, c.dialOptions...)
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
	if !isConnectionError(err) && !errors.Is(err, errRestoreDir) {
		return err
	}
	if closeErr := c.Close(); closeErr != nil {
		log.Warn().Err(closeErr).Msg("Cannot close broken ftp connection")
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
	dir, err := c.pathenc.Encode(dir)
	if err != nil {
		return nil, err
	}
	files, err := c.list(dir)
	if err != nil && c.ftp == nil && c.connectErr == nil {
		// The session was dropped as broken, which says nothing about this
		// directory. Retry it once on a fresh connection rather than skip it
		// for the rest of the run.
		log.Warn().Err(err).Str("dir", dir).Msg("Listing failed on a broken connection, retrying on a new one")
		files, err = c.list(dir)
	}
	if err != nil {
		return nil, err
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

func (c *Client) list(dir string) ([]*ftp.Entry, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}
	files, err := c.ftp.List(dir)
	if err != nil {
		return nil, c.handleTransferError(err)
	}
	return files, nil
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

// newTimeoutDialFunc must be recreated per session: explicit TLS starts with a
// plaintext control connection, then uses TLS for data connections.
func newTimeoutDialFunc(timeout time.Duration, mode tlsMode, tlsConfig *tls.Config) netDialFunc {
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

// isConnectionError reports whether err means the FTP session itself is no
// longer usable, as opposed to a failure scoped to one path. Reusing a broken
// session fails every later command of the run, often with a misleading error
// read from a stale reply.
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	if isTimeoutError(err) {
		return true
	}
	var protoErr *textproto.Error
	if errors.As(err, &protoErr) {
		switch protoErr.Code {
		case ftp.StatusNotAvailable, ftp.StatusCanNotOpenDataConnection, ftp.StatusTransfertAborted:
			return true
		}
		// vsftpd answers "500 OOPS: ..." once its session process has failed,
		// and "500 PASV: <errno>" / "500 EPSV: <errno>" when it can no longer
		// set up a data channel. Neither recovers on the same session.
		for _, prefix := range []string{"OOPS:", "PASV:", "EPSV:"} {
			if strings.HasPrefix(protoErr.Msg, prefix) {
				return true
			}
		}
		return false
	}
	var opErr *net.OpError
	var recordErr tls.RecordHeaderError
	return errors.As(err, &opErr) ||
		errors.As(err, &recordErr) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, net.ErrClosed) ||
		errors.Is(err, syscall.EPIPE) ||
		errors.Is(err, syscall.ECONNRESET)
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
