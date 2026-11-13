package sftp

import (
	"context"
	stderrors "errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/secret"
	"github.com/crazy-max/ftpgrab/v7/internal/server"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"github.com/rs/zerolog/log"
	"github.com/skeema/knownhosts"
	"golang.org/x/crypto/ssh"
)

var newSFTPClient = sftp.NewClient

// Client represents an active sftp object
type Client struct {
	*server.Client
	config *config.ServerSFTP
	sftp   *sftp.Client
	ssh    *ssh.Client
	closer io.Closer
}

// New creates new ftp instance
func New(config *config.ServerSFTP) (*server.Client, error) {
	var err error
	client := &Client{config: config}
	var sshConf *ssh.ClientConfig
	var sshAuth []ssh.AuthMethod

	// SSH Auth
	if len(config.KeyFile) > 0 {
		keyPassphrase, err := secret.GetSecret(config.KeyPassphrase, config.KeyPassphraseFile)
		if err != nil {
			log.Warn().Err(err).Msg("Cannot retrieve key passphrase secret for sftp server")
		}
		if sshAuth, err = client.readPublicKey(config.KeyFile, keyPassphrase); err != nil {
			return nil, errors.Wrap(err, "Unable to read SFTP public key")
		}
	} else if len(config.Password) > 0 || len(config.PasswordFile) > 0 {
		password, err := secret.GetSecret(config.Password, config.PasswordFile)
		if err != nil {
			log.Warn().Err(err).Msg("Cannot retrieve password secret for sftp server")
		}
		sshAuth = []ssh.AuthMethod{
			ssh.Password(password),
		}
	}

	username, err := secret.GetSecret(config.Username, config.UsernameFile)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot retrieve username secret for sftp server")
	}

	knownHosts, err := loadKnownHosts(config.KnownHostsFile)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot configure SFTP host key verification")
	}

	addr := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
	sshConf = &ssh.ClientConfig{
		User:              username,
		Auth:              sshAuth,
		HostKeyCallback:   knownHosts.HostKeyCallback(),
		HostKeyAlgorithms: knownHosts.HostKeyAlgorithms(addr),
		Timeout:           *config.Timeout,
	}
	// Prefer trusted key types while allowing a CA to sign other host key types.
	for _, algo := range ssh.SupportedAlgorithms().HostKeys {
		if !slices.Contains(sshConf.HostKeyAlgorithms, algo) {
			sshConf.HostKeyAlgorithms = append(sshConf.HostKeyAlgorithms, algo)
		}
	}

	sshConf.SetDefaults()
	client.ssh, err = dialSSH("tcp", addr, sshConf)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot open ssh connection")
	}
	client.closer = client.ssh

	if err = client.openSFTP(); err != nil {
		return nil, err
	}

	return &server.Client{Handler: client}, err
}

func loadKnownHosts(filename string) (*knownhosts.HostKeyDB, error) {
	if filename == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, errors.Wrap(err, "resolve user home directory for known_hosts; configure knownHostsFile explicitly")
		}
		filename = filepath.Join(home, ".ssh", "known_hosts")
	}
	db, err := knownhosts.NewDB(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "load known hosts file %q; configure knownHostsFile with trusted server keys", filename)
	}
	return db, nil
}

func (c *Client) openSFTP() error {
	sftpClient, err := newSFTPClient(c.ssh, sftp.MaxPacket(c.config.MaxPacketSize))
	if err != nil {
		if c.closer != nil {
			if closeErr := c.closer.Close(); closeErr != nil {
				log.Warn().Err(closeErr).Msg("Cannot close ssh connection after sftp client initialization failure")
			}
		}
		return err
	}
	c.sftp = sftpClient
	return nil
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
	err := c.sftp.Close()
	if err == io.EOF {
		err = nil
	}
	return stderrors.Join(err, c.ssh.Close())
}

func dialSSH(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	dialer := &net.Dialer{Timeout: config.Timeout}
	conn, err := dialer.DialContext(context.Background(), network, addr)
	if err != nil {
		return nil, err
	}
	conn = server.NewTimeoutConn(conn, config.Timeout)
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return ssh.NewClient(c, chans, reqs), nil
}
