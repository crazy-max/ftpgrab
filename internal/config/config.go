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

	"github.com/ftpgrab/ftpgrab/internal/model"
	"gopkg.in/yaml.v2"
)

// Configuration holds configuration details
type Configuration struct {
	Flags    *model.Flags
	App      *model.App      `yaml:"app,omitempty"`
	Server   *model.Server   `yaml:"server,omitempty"`
	Download *model.Download `yaml:"download,omitempty"`
	Mail     *model.Mail     `yaml:"mail,omitempty"`
	File     os.FileInfo
	Location *time.Location
}

// Load returns Configuration struct
func Load(fl *model.Flags, version string) (*Configuration, error) {
	var err error
	var cfg = &Configuration{Flags: fl}

	cfg.App = &model.App{
		ID:      "ftpgrab",
		Name:    "FTPGrab",
		Desc:    "Grab your files from a remote FTP server easily",
		URL:     "https://ftpgrab.github.io",
		Author:  "CrazyMax",
		Version: version,
	}

	if cfg.File, err = os.Lstat(fl.Cfgfile); err != nil {
		return nil, fmt.Errorf("unable to open config file, %s", err)
	}

	bytes, err := ioutil.ReadFile(fl.Cfgfile)
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

	if cfg.Flags.Output == "" {
		return errors.New("output destination folder is required")
	}
	cfg.Flags.Output = path.Clean(cfg.Flags.Output)

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
