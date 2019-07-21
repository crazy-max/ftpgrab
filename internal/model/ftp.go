package model

// FTP holds data necessary for FTP configuration
type FTP struct {
	Host               string   `yaml:"host,omitempty"`
	Port               int      `yaml:"port,omitempty"`
	Username           string   `yaml:"username,omitempty"`
	Password           string   `yaml:"password,omitempty"`
	Sources            []string `yaml:"sources,omitempty"`
	Timeout            int      `yaml:"timeout,omitempty"`
	DisableEPSV        bool     `yaml:"disable_epsv,omitempty"`
	TLS                bool     `yaml:"tls,omitempty"`
	InsecureSkipVerify bool     `yaml:"insecure_skip_verify,omitempty"`
	LogTrace           bool     `yaml:"log_trace,omitempty"`
}
