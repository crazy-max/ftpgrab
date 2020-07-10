package model

import (
	"github.com/pkg/errors"
)

// Server represents a server configuration
type Server struct {
	FTP  *ServerFTP  `yaml:"ftp,omitempty" json:"ftp,omitempty"`
	SFTP *ServerSFTP `yaml:"sftp,omitempty" json:"sftp,omitempty"`
}

// ServerCommon holds common data server configuration
type ServerCommon struct {
	Host    string
	Port    int
	Sources []string
}

// GetDefaults gets the default values
func (s *Server) GetDefaults() *Server {
	return nil
}

// SetDefaults sets the default values
func (s *Server) SetDefaults() {
	// noop
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (s *Server) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Server
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}

	if s.FTP == nil && s.SFTP == nil {
		return errors.New("one server (ftp or sftp) must be defined")
	} else if s.FTP != nil && s.SFTP != nil {
		return errors.New("only one server (ftp or sftp) is allowed")
	}

	return nil
}
