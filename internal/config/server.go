package config

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
