package model

// SFTP holds data necessary for SFTP configuration
type SFTP struct {
	Host               string   `yaml:"host,omitempty"`
	Port               int      `yaml:"port,omitempty"`
	Username           string   `yaml:"username,omitempty"`
	Password           string   `yaml:"password,omitempty"`
	Sources            []string `yaml:"sources,omitempty"`
	MaxPacketSize      int      `yaml:"max_packet_size,omitempty"`
	InsecureSkipVerify bool     `yaml:"insecure_skip_verify,omitempty"`
}
