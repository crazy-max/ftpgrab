package model

import "time"

// App holds application configuration details
type App struct {
	ID       string
	Name     string
	Desc     string
	URL      string
	Author   string
	Version  string
	Timezone string `yaml:"timezone,omitempty"`
}

// Server holds data necessary for server configuration
type Server struct {
	Host               string `yaml:"host,omitempty"`
	Port               int    `yaml:"port,omitempty"`
	Username           string `yaml:"username,omitempty"`
	Password           string `yaml:"password,omitempty"`
	ConnectionsPerHost int    `yaml:"connections_per_host,omitempty"`
	Timeout            int    `yaml:"timeout,omitempty"`
	DisableEPSV        bool   `yaml:"disable_epsv,omitempty"`
	TLS                struct {
		Enable             bool `yaml:"enable,omitempty"`
		Implicit           bool `yaml:"implicit,omitempty"`
		InsecureSkipVerify bool `yaml:"insecure_skip_verify,omitempty"`
	} `yaml:"tls,omitempty"`
	Sources []string `yaml:"sources,omitempty"`
}

// Download holds download configuration details
type Download struct {
	UID           int       `yaml:"uid,omitempty"`
	GID           int       `yaml:"gid,omitempty"`
	ChmodFile     int       `yaml:"chmod_file,omitempty"`
	ChmodDir      int       `yaml:"chmod_dir,omitempty"`
	Include       []string  `yaml:"include,omitempty"`
	Exclude       []string  `yaml:"exclude,omitempty"`
	Since         time.Time `yaml:"since,omitempty"`
	Retry         int       `yaml:"retry,omitempty"`
	HashEnabled   bool      `yaml:"hash_enabled,omitempty"`
	HideSkipped   bool      `yaml:"hide_skipped,omitempty"`
	CreateBasedir bool      `yaml:"create_basedir,omitempty"`
}

// Mail holds mail notification configuration details
type Mail struct {
	Enabled  bool   `yaml:"enabled,omitempty"`
	Host     string `yaml:"host,omitempty"`
	Port     int    `yaml:"port,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	From     string `yaml:"from,omitempty"`
	To       string `yaml:"to,omitempty"`
}
