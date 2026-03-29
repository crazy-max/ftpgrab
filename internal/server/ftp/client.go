package ftp

import (
	"crypto/tls"
	"errors"
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

const (
	// checkInterval is how often to check download speed during transfers.
	checkInterval = 30 * time.Second
	// minSlowChecks is the number of consecutive slow checks before aborting.
	minSlowChecks = 2
)

var (
	// ErrSlowTransfer is returned when sustained throughput drops below the minimum speed.
	ErrSlowTransfer = errors.New("transfer speed too slow, reconnecting")
	// ErrListTimeout is returned when a List operation exceeds the configured timeout.
	ErrListTimeout = errors.New("list operation timed out")
)

// speedMonitor wraps an io.Reader to abort transfers whose sustained
// throughput drops below minSpeed. Unlike TimeoutConn (which catches
// complete stalls at the TCP level), this detects slow trickles that
// would otherwise keep the connection alive indefinitely.
type speedMonitor struct {
	reader          io.Reader
	lastCheck       time.Time
	lastBytes       int64
	totalBytes      int64
	minSpeed        int64
	interval        time.Duration
	consecutiveSlow int
}

func (s *speedMonitor) Read(p []byte) (n int, err error) {
	n, err = s.reader.Read(p)
	s.totalBytes += int64(n)

	now := time.Now()
	if now.Sub(s.lastCheck) >= s.interval {
		elapsed := now.Sub(s.lastCheck).Seconds()
		bytesSinceCheck := s.totalBytes - s.lastBytes

		if elapsed > 0 {
			speed := float64(bytesSinceCheck) / elapsed
			if speed < float64(s.minSpeed) && s.lastBytes > 0 {
				s.consecutiveSlow++
				log.Warn().
					Float64("speed_bps", speed).
					Int64("min_speed_bps", s.minSpeed).
					Int("consecutive_slow", s.consecutiveSlow).
					Msgf("Transfer speed below threshold (%d/%d)", s.consecutiveSlow, minSlowChecks)

				if s.consecutiveSlow >= minSlowChecks {
					return n, ErrSlowTransfer
				}
			} else {
				s.consecutiveSlow = 0
			}
		}

		s.lastCheck = now
		s.lastBytes = s.totalBytes
	}

	return n, err
}

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

	// Add timeout to List operation to prevent hanging
	type listResult struct {
		files []*ftp.Entry
		err   error
	}
	resultCh := make(chan listResult, 1)

	go func() {
		f, e := c.ftp.List(dir)
		resultCh <- listResult{files: f, err: e}
	}()

	select {
	case result := <-resultCh:
		files = result.files
		if err := result.err; err != nil {
			return nil, err
		}
	case <-time.After(*c.cfg.Timeout):
		log.Warn().
			Dur("timeout", *c.cfg.Timeout).
			Msg("Cannot list directory, timeout reached")
		return nil, ErrListTimeout
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

	var src io.Reader = resp
	if *c.cfg.MinSpeed > 0 {
		src = &speedMonitor{
			reader:    resp,
			lastCheck: time.Now(),
			minSpeed:  *c.cfg.MinSpeed * 1024, // convert KB/s to bytes/s
			interval:  checkInterval,
		}
	}

	_, err = io.Copy(dest, src)
	return err
}

// Close closes ftp connection
func (c *Client) Close() error {
	return c.ftp.Quit()
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
