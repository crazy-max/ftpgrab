package model

// FTP holds data necessary for FTP configuration
type FTP struct {
	Host        string   `yaml:"host,omitempty"`
	Port        int      `yaml:"port,omitempty"`
	Username    string   `yaml:"username,omitempty"`
	Password    string   `yaml:"password,omitempty"`
	Sources     []string `yaml:"sources,omitempty"`
	Timeout     int      `yaml:"timeout,omitempty"`
	DisableEPSV bool     `yaml:"disable_epsv,omitempty"`
	TLS         TLS      `yaml:"tls,omitempty"`
	LogTrace    bool     `yaml:"log_trace,omitempty"`
}

// TLS holds data necessary for TLS FTP configuration
type TLS struct {
	Enable             bool `yaml:"enable,omitempty"`
	Implicit           bool `yaml:"implicit,omitempty"`
	InsecureSkipVerify bool `yaml:"insecure_skip_verify,omitempty"`
}
