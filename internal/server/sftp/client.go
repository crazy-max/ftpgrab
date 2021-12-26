package sftp

import (
	"fmt"
	"io"
	"os"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/server"
	"github.com/crazy-max/ftpgrab/v7/pkg/utl"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

// Client represents an active sftp object
type Client struct {
	*server.Client
	config *config.ServerSFTP
	sftp   *sftp.Client
	ssh    *ssh.Client
}

// New creates new ftp instance
func New(config *config.ServerSFTP) (*server.Client, error) {
	var err error
	var client = &Client{config: config}
	var sshConf *ssh.ClientConfig
	var sshAuth []ssh.AuthMethod

	// SSH Auth
	if len(config.KeyFile) > 0 {
		keyPassphrase, err := utl.GetSecret(config.KeyPassphrase, config.KeyPassphraseFile)
		if err != nil {
			log.Warn().Err(err).Msg("Cannot retrieve key passphrase secret for sftp server")
		}
		if sshAuth, err = client.readPublicKey(config.KeyFile, keyPassphrase); err != nil {
			return nil, errors.Wrap(err, "Unable to read SFTP public key")
		}
	} else if len(config.Password) > 0 || len(config.PasswordFile) > 0 {
		password, err := utl.GetSecret(config.Password, config.PasswordFile)
		if err != nil {
			log.Warn().Err(err).Msg("Cannot retrieve password secret for sftp server")
		}
		sshAuth = []ssh.AuthMethod{
			ssh.Password(password),
		}
	}

	username, err := utl.GetSecret(config.Username, config.UsernameFile)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot retrieve username secret for sftp server")
	}

	sshConf = &ssh.ClientConfig{
		User:            username,
		Auth:            sshAuth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         *config.Timeout,
	}

	sshConf.SetDefaults()
	client.ssh, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), sshConf)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot open ssh connection")
	}

	if client.sftp, err = sftp.NewClient(client.ssh, sftp.MaxPacket(config.MaxPacketSize)); err != nil {
		return nil, err
	}

	return &server.Client{Handler: client}, err
}

// Common return common configuration
func (c *Client) Common() config.ServerCommon {
	return config.ServerCommon{
		Host:    c.config.Host,
		Port:    c.config.Port,
		Sources: c.config.Sources,
	}
}

func (c *Client) readPublicKey(key string, password string) ([]ssh.AuthMethod, error) {
	var signer ssh.Signer
	var err error

	buffer, err := os.ReadFile(key)
	if err != nil {
		return nil, err
	}

	if len(password) > 0 {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(buffer, []byte(password))
	} else {
		signer, err = ssh.ParsePrivateKey(buffer)
	}
	if err != nil {
		return nil, err
	}

	return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil
}

// ReadDir fetches the contents of a directory, returning a list of os.FileInfo's
func (c *Client) ReadDir(path string) ([]os.FileInfo, error) {
	return c.sftp.ReadDir(path)
}

// Retrieve file "path" from server and write bytes to "dest".
func (c *Client) Retrieve(path string, dest io.Writer) error {
	reader, err := c.sftp.Open(path)
	if err != nil {
		return err
	}
	defer reader.Close()

	if _, err := io.Copy(dest, reader); err != nil {
		return err
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
