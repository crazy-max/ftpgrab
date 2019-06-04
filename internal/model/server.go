package model

const (
	ServerTypeFTP  = ServerType("ftp")
	ServerTypeSFTP = ServerType("sftp")
)

// Server holds data necessary for server configuration
type Server struct {
	Type     ServerType `yaml:"type,omitempty"`
	FTP      FTP        `yaml:"ftp,omitempty"`
	SFTP     SFTP       `yaml:"sftp,omitempty"`
	Encoding string     `yaml:"encoding,omitempty"`
}

// ServerType is the server type, can be ftp or sftp
type ServerType string

// Common holds common data server configuration
type Common struct {
	Host     string   `yaml:"host,omitempty"`
	Port     int      `yaml:"port,omitempty"`
	Username string   `yaml:"username,omitempty"`
	Password string   `yaml:"password,omitempty"`
	Sources  []string `yaml:"sources,omitempty"`
}
