package sftp

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

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
	var sshConf *ssh.ClientConfig
	var sshAuth []ssh.AuthMethod

	// SSH Auth
	if config.Key != "" {
		if sshAuth, err = client.readPublicKey(config.Key, config.Password); err != nil {
			return nil, fmt.Errorf("unable to read SFTP public key, %v", err)
		}
	} else {
		sshAuth = []ssh.AuthMethod{
			ssh.Password(config.Password),
		}
	}
	sshConf = &ssh.ClientConfig{
		User:            config.Username,
		Auth:            sshAuth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         config.Timeout * time.Second,
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

func (c *Client) readPublicKey(key string, password string) ([]ssh.AuthMethod, error) {
	var signer ssh.Signer
	var err error

	buffer, err := ioutil.ReadFile(key)
	if err != nil {
		return nil, err
	}

	if password != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(buffer, []byte(password))
	} else {
		signer, err = ssh.ParsePrivateKey(buffer)
	}
	if err != nil {
		return nil, err
	}

	return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil
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
