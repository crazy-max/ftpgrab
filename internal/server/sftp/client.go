package sftp

import (
	"fmt"
	"io"
	"os"

	"github.com/ftpgrab/ftpgrab/internal/model"
	"github.com/ftpgrab/ftpgrab/internal/server"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Client represents an active sftp object
type Client struct {
	*server.Client
	sftp *sftp.Client
	ssh  *ssh.Client
	cfg  *model.SFTP
}

// New creates new ftp instance
func New(config *model.SFTP) (*server.Client, error) {
	var err error
	var client = &Client{cfg: config}

	var hostKeyCallback ssh.HostKeyCallback
	if config.InsecureSkipVerify {
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	sshConf := &ssh.ClientConfig{
		User:            config.Username,
		HostKeyCallback: hostKeyCallback,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
		},
	}

	sshConf.SetDefaults()
	client.ssh, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), sshConf)
	if err != nil {
		return nil, fmt.Errorf("cannot open ssh connection, %v", err)
	}

	if client.sftp, err = sftp.NewClient(client.ssh); err != nil {
		return nil, err
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
	return c.sftp.ReadDir(path)
}

// Retrieve file "path" from server and write bytes to "dest".
func (c *Client) Retrieve(path string, dest io.Writer) error {
	remoteFile, err := c.sftp.Open(path)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	b := make([]byte, 1024)
	for {
		n, err := remoteFile.Read(b)
		dest.Write(b[:n])
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
	}

	return nil
}

// Close closes sftp connection
func (c *Client) Close() error {
	if err := c.ssh.Close(); err != nil {
		return err
	}
	if err := c.sftp.Close(); err != nil && err != io.EOF {
		return err
	}
	return nil
}
