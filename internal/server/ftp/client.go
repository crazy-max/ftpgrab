package ftp

import (
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/ftpgrab/ftpgrab/v7/internal/logging"
	"github.com/ftpgrab/ftpgrab/v7/internal/model"
	"github.com/ftpgrab/ftpgrab/v7/internal/server"
	"github.com/ftpgrab/ftpgrab/v7/pkg/utl"
	"github.com/jlaffaye/ftp"
	"github.com/rs/zerolog/log"
)

// Client represents an active ftp object
type Client struct {
	*server.Client
	config *model.ServerFTP
	ftp    *ftp.ServerConn
}

// New creates new ftp instance
func New(config *model.ServerFTP) (*server.Client, error) {
	var err error
	var client = &Client{config: config}

	ftpConfig := []ftp.DialOption{
		ftp.DialWithTimeout(*config.Timeout),
		ftp.DialWithDisabledEPSV(*config.DisableEPSV),
		ftp.DialWithDebugOutput(&logging.FtpWriter{
			Enabled: *config.LogTrace,
		}),
	}

	if *config.TLS {
		ftpConfig = append(ftpConfig, ftp.DialWithTLS(&tls.Config{
			ServerName:         config.Host,
			InsecureSkipVerify: *config.InsecureSkipVerify,
		}))
	}

	if client.ftp, err = ftp.Dial(fmt.Sprintf("%s:%d", config.Host, config.Port), ftpConfig...); err != nil {
		return nil, err
	}

	username, err := utl.GetSecret(config.Username, config.UsernameFile)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot retrieve username secret for ftp server")
	}
	password, err := utl.GetSecret(config.Password, config.PasswordFile)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot retrieve password secret for ftp server")
	}

	if len(username) > 0 {
		if err = client.ftp.Login(username, password); err != nil {
			return nil, err
		}
	}

	return &server.Client{Handler: client}, err
}

// Common return common configuration
func (c *Client) Common() model.ServerCommon {
	return model.ServerCommon{
		Host:    c.config.Host,
		Port:    c.config.Port,
		Sources: c.config.Sources,
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

	_, err = io.Copy(dest, resp)
	return err
}

// Close closes ftp connection
func (c *Client) Close() error {
	return c.ftp.Quit()
}
