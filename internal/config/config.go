package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/mail"
	"os"
	"path"
	"regexp"

	"github.com/ftpgrab/ftpgrab/internal/model"
	"github.com/imdario/mergo"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// Configuration holds configuration details
type Configuration struct {
	Flags    model.Flags
	App      model.App
	Server   model.Server   `yaml:"server,omitempty"`
	Db       model.Db       `yaml:"db,omitempty"`
	Download model.Download `yaml:"download,omitempty"`
	Notif    model.Notif    `yaml:"notif,omitempty"`
	File     os.FileInfo
}

// Load returns Configuration struct
func Load(fl model.Flags, version string) (*Configuration, error) {
	var err error
	var cfg = Configuration{
		Flags: fl,
		App: model.App{
			ID:      "ftpgrab",
			Name:    "FTPGrab",
			Desc:    "Grab your files periodically from a remote FTP or SFTP server easily",
			URL:     "https://ftpgrab.github.io",
			Author:  "CrazyMax",
			Version: version,
		},
		Server: model.Server{
			Type: model.ServerTypeFTP,
			FTP: model.FTP{
				Port:               21,
				Sources:            []string{},
				ConnectionsPerHost: 5,
				Timeout:            5,
				DisableEPSV:        false,
				TLS: model.TLS{
					Enable:             false,
					Implicit:           true,
					InsecureSkipVerify: false,
				},
				LogTrace: false,
			},
			SFTP: model.SFTP{
				Port:          22,
				Sources:       []string{},
				Timeout:       30,
				MaxPacketSize: 32768,
			},
		},
		Db: model.Db{
			Enable: true,
			Path:   "ftpgrab.db",
		},
		Download: model.Download{
			UID:           os.Getuid(),
			GID:           os.Getgid(),
			ChmodFile:     0644,
			ChmodDir:      0755,
			Retry:         3,
			HideSkipped:   false,
			CreateBasedir: false,
		},
		Notif: model.Notif{
			Mail: model.Mail{
				Enable:             false,
				Host:               "localhost",
				Port:               25,
				SSL:                false,
				InsecureSkipVerify: false,
			},
			Webhook: model.Webhook{
				Enable:  false,
				Method:  "GET",
				Timeout: 10,
			},
		},
	}

	if cfg.File, err = os.Lstat(fl.Cfgfile); err != nil {
		return nil, fmt.Errorf("unable to open config file, %s", err)
	}

	bytes, err := ioutil.ReadFile(fl.Cfgfile)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file, %s", err)
	}

	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %v", err)
	}

	return &cfg, nil
}

// Check verifies Configuration values
func (cfg *Configuration) Check() error {
	if err := checkServer(&cfg.Server); err != nil {
		return err
	}

	if cfg.Flags.Docker {
		cfg.Db.Path = "/db/ftpgrab.db"
		cfg.Download.Output = "/download"
	}

	if cfg.Db.Enable && cfg.Db.Path == "" {
		return errors.New("path to database path is required if enabled")
	}
	cfg.Db.Path = path.Clean(cfg.Db.Path)

	if cfg.Download.Output == "" {
		return errors.New("output download folder is required")
	}
	cfg.Download.Output = path.Clean(cfg.Download.Output)

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

	if cfg.Notif.Mail.Enable {
		if _, err := mail.ParseAddress(cfg.Notif.Mail.From); err != nil {
			return fmt.Errorf("cannot sender mail address, %v", err)
		}
		if _, err := mail.ParseAddress(cfg.Notif.Mail.To); err != nil {
			return fmt.Errorf("cannot recipient mail address, %v", err)
		}
	}

	return nil
}

func checkServer(cfg *model.Server) error {
	switch cfg.Type {
	case model.ServerTypeFTP:
		return checkServerFtp(cfg.FTP)
	case model.ServerTypeSFTP:
		return checkServerSftp(cfg.SFTP)
	default:
		return fmt.Errorf("unknown server type %s", cfg.Type)
	}
}

func checkServerFtp(cfg model.FTP) error {
	if cfg.Host == "" {
		return errors.New("FTP host is required")
	}

	if len(cfg.Sources) == 0 {
		return errors.New("at least one FTP source is required")
	}

	return nil
}

func checkServerSftp(cfg model.SFTP) error {
	if cfg.Host == "" {
		return errors.New("SFTP host is required")
	}

	if len(cfg.Sources) == 0 {
		return errors.New("at least one SFTP source is required")
	}

	return nil
}

// Display logs configuration in a pretty JSON format
func (cfg *Configuration) Display() {
	var out = Configuration{
		Server: model.Server{
			FTP: model.FTP{
				Username: "********",
				Password: "********",
			},
			SFTP: model.SFTP{
				Username: "********",
				Password: "********",
				Key:      "********",
			},
		},
		Notif: model.Notif{
			Mail: model.Mail{
				Username: "********",
				Password: "********",
			},
		},
	}
	if err := mergo.Merge(&out, cfg); err != nil {
		log.Error().Err(err).Msg("Cannot merge config")
		return
	}
	b, _ := json.MarshalIndent(out, "", "  ")
	log.Debug().Msg(string(b))
}
