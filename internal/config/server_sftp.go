package config

import (
	"time"

	"github.com/crazy-max/ftpgrab/v7/pkg/utl"
)

// ServerSFTP holds sftp server configuration
type ServerSFTP struct {
	Host              string         `yaml:"host,omitempty" json:"host,omitempty" validate:"required"`
	Port              int            `yaml:"port,omitempty" json:"port,omitempty" validate:"required,min=1"`
	Username          string         `yaml:"username,omitempty" json:"username,omitempty"`
	UsernameFile      string         `yaml:"usernameFile,omitempty" json:"usernameFile,omitempty" validate:"omitempty,file"`
	Password          string         `yaml:"password,omitempty" json:"password,omitempty"`
	PasswordFile      string         `yaml:"passwordFile,omitempty" json:"passwordFile,omitempty" validate:"omitempty,file"`
	KeyFile           string         `yaml:"keyFile,omitempty" json:"keyFile,omitempty" validate:"omitempty,file"`
	KeyPassphrase     string         `yaml:"keyPassphrase,omitempty" json:"keyPassphrase,omitempty"`
	KeyPassphraseFile string         `yaml:"keyPassphraseFile,omitempty" json:"keyPassphraseFile,omitempty" validate:"omitempty,file"`
	Sources           []string       `yaml:"sources,omitempty" json:"sources,omitempty"`
	Timeout           *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	MaxPacketSize     int            `yaml:"maxPacketSize,omitempty" json:"maxPacketSize,omitempty"`
}

// GetDefaults gets the default values
func (s *ServerSFTP) GetDefaults() *ServerSFTP {
	n := &ServerSFTP{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *ServerSFTP) SetDefaults() {
	s.Port = 22
	s.Sources = []string{}
	s.Timeout = utl.NewDuration(30 * time.Second)
	s.MaxPacketSize = 32768
}
