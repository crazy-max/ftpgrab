package model

import "time"

// SFTP holds data necessary for SFTP configuration
type SFTP struct {
	Host          string        `yaml:"host,omitempty"`
	Port          int           `yaml:"port,omitempty"`
	Username      string        `yaml:"username,omitempty"`
	Password      string        `yaml:"password,omitempty"`
	Key           string        `yaml:"key,omitempty"`
	Sources       []string      `yaml:"sources,omitempty"`
	Timeout       time.Duration `yaml:"timeout,omitempty"`
	MaxPacketSize int           `yaml:"max_packet_size,omitempty"`
}
