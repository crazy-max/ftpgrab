package config

import (
	"time"

	"github.com/crazy-max/ftpgrab/v7/pkg/utl"
)

// ServerFTP holds ftp server configuration
type ServerFTP struct {
	Host               string         `yaml:"host,omitempty" json:"host,omitempty" validate:"required"`
	Port               int            `yaml:"port,omitempty" json:"port,omitempty" validate:"required,min=1"`
	Username           string         `yaml:"username,omitempty" json:"username,omitempty"`
	UsernameFile       string         `yaml:"usernameFile,omitempty" json:"usernameFile,omitempty" validate:"omitempty,file"`
	Password           string         `yaml:"password,omitempty" json:"password,omitempty"`
	PasswordFile       string         `yaml:"passwordFile,omitempty" json:"passwordFile,omitempty" validate:"omitempty,file"`
	Sources            []string       `yaml:"sources,omitempty" json:"sources,omitempty"`
	Timeout            *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	DisableUTF8        *bool          `yaml:"disableUTF8,omitempty" json:"disableUTF8,omitempty"`
	DisableEPSV        *bool          `yaml:"disableEPSV,omitempty" json:"disableEPSV,omitempty"`
	DisableMLSD        *bool          `yaml:"disableMLSD,omitempty" json:"disableMLSD,omitempty"`
	TLS                *bool          `yaml:"tls,omitempty" json:"tls,omitempty"`
	InsecureSkipVerify *bool          `yaml:"insecureSkipVerify,omitempty" json:"insecureSkipVerify,omitempty"`
	LogTrace           *bool          `yaml:"logTrace,omitempty" json:"logTrace,omitempty"`
}

// GetDefaults gets the default values
func (s *ServerFTP) GetDefaults() *ServerFTP {
	n := &ServerFTP{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *ServerFTP) SetDefaults() {
	s.Port = 21
	s.Sources = []string{}
	s.Timeout = utl.NewDuration(5 * time.Second)
	s.DisableUTF8 = utl.NewFalse()
	s.DisableEPSV = utl.NewFalse()
	s.DisableMLSD = utl.NewFalse()
	s.TLS = utl.NewFalse()
	s.InsecureSkipVerify = utl.NewFalse()
	s.LogTrace = utl.NewFalse()
}
