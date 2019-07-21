package ftp

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"time"

	"github.com/ftpgrab/ftpgrab/internal/logging"
	"github.com/ftpgrab/ftpgrab/internal/model"
	"github.com/ftpgrab/ftpgrab/internal/server"
	"github.com/jlaffaye/ftp"
)

// Client represents an active ftp object
type Client struct {
	*server.Client
	ftp *ftp.ServerConn
	cfg *model.FTP
}

// New creates new ftp instance
func New(config *model.FTP) (*server.Client, error) {
	var err error
	var tlsConfig *tls.Config
	var client = &Client{cfg: config}

	ftpConfig := []ftp.DialOption{
		ftp.DialWithTimeout(time.Duration(config.Timeout) * time.Second),
		ftp.DialWithDisabledEPSV(config.DisableEPSV),
		ftp.DialWithDebugOutput(&logging.FtpWriter{
			Enabled: config.LogTrace,
		}),
	}

	if config.TLS {
		tlsConfig = &tls.Config{
			ServerName:         config.Host,
			InsecureSkipVerify: config.InsecureSkipVerify,
		}
		ftpConfig = append(ftpConfig, ftp.DialWithTLS(tlsConfig))
	}

	if client.ftp, err = ftp.Dial(fmt.Sprintf("%s:%d", config.Host, config.Port), ftpConfig...); err != nil {
		return nil, err
	}

	if len(config.Username) > 0 {
		if err = client.ftp.Login(config.Username, config.Password); err != nil {
			return nil, err
		}
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
	var files []*ftp.Entry
	files, err := c.ftp.List(regexp.QuoteMeta(path))
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

	buf, err := ioutil.ReadAll(resp)
	if err != nil {
		return err
	}

	_, err = dest.Write(buf)
	return err
}

// Close closes ftp connection
func (c *Client) Close() error {
	return c.ftp.Quit()
}
