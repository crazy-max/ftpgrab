package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/mail"
	"os"
	"path"
	"regexp"
	"time"

	"gopkg.in/yaml.v2"
)

// Configuration holds configuration details
type Configuration struct {
	App      *App      `yaml:"app,omitempty"`
	Server   *Server   `yaml:"server,omitempty"`
	Download *Download `yaml:"download,omitempty"`
	Mail     *Mail     `yaml:"mail,omitempty"`
	File     os.FileInfo
	Location *time.Location
}

// App holds application configuration details
type App struct {
	ID       string
	Name     string
	Desc     string
	URL      string
	Author   string
	Version  string
	Timezone string `yaml:"timezone,omitempty"`
	LogFtp   bool
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
	Dest          string    `yaml:"dest,omitempty"`
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

// Load returns Configuration struct
func Load(file string, logFtp bool, version string) (*Configuration, error) {
	var err error
	var cfg = new(Configuration)

	cfg.App = new(App)
	cfg.App.ID = "ftpgrab"
	cfg.App.Name = "FTPGrab"
	cfg.App.Desc = "Grab your files from a remote FTP server easily"
	cfg.App.URL = "https://ftpgrab.github.io"
	cfg.App.Author = "CrazyMax"
	cfg.App.Version = version
	cfg.App.LogFtp = logFtp

	if cfg.File, err = os.Lstat(file); err != nil {
		return nil, fmt.Errorf("unable to open config file, %s", err)
	}

	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file, %s", err)
	}

	if err := yaml.Unmarshal(bytes, cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %v", err)
	}

	return cfg, nil
}

// Check verifies Configuration values
func (cfg *Configuration) Check() error {
	var err error

	if cfg.Location, err = time.LoadLocation(cfg.App.Timezone); err != nil {
		return fmt.Errorf("cannot load timezone, %v", err)
	}

	if cfg.Server.Host == "" {
		return errors.New("host is required")
	}

	if len(cfg.Server.Sources) == 0 {
		return errors.New("at least one source is required")
	}

	if cfg.Download.Dest == "" {
		return errors.New("download destination path is required")
	}
	cfg.Download.Dest = path.Clean(cfg.Download.Dest)

	for _, include := range cfg.Download.Include {
		if _, err := regexp.Compile(include); err != nil {
			return fmt.Errorf("include regex '%s' cannot compile, %v", include, err)
		}
	}

	for _, exclude := range cfg.Download.Exclude {
		if _, err := regexp.Compile(exclude); err != nil {
			return fmt.Errorf("exclude regex '%s' cannot compile, %v", exclude, err)
		}
	}

	if cfg.Mail.Enabled {
		if _, err := mail.ParseAddress(cfg.Mail.From); err != nil {
			return fmt.Errorf("cannot load timezone, %v", err)
		}

		if _, err := mail.ParseAddress(cfg.Mail.To); err != nil {
			return errors.New("invalid recipient mail address")
		}
	}

	return nil
}
