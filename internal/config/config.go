package config

import (
	"encoding/json"
	"os"
	"path"
	"regexp"

	"github.com/containous/traefik/v2/pkg/config/env"
	"github.com/containous/traefik/v2/pkg/config/file"
	"github.com/ftpgrab/ftpgrab/v7/internal/model"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Config holds configuration details
type Config struct {
	Schedule string          `yaml:"schedule,omitempty" json:"schedule,omitempty"`
	Db       *model.Db       `yaml:"db,omitempty" json:"db,omitempty" validate:"omitempty"`
	Server   *model.Server   `yaml:"server,omitempty" json:"server,omitempty" validate:"required"`
	Download *model.Download `yaml:"download,omitempty" json:"download,omitempty" validate:"required"`
	Notif    *model.Notif    `yaml:"notif,omitempty" json:"notif,omitempty"`
}

// Load returns Config struct
func Load(cfgfile string, schedule string) (*Config, error) {
	cfg := Config{
		Schedule: schedule,
		Db:       (&model.Db{}).GetDefaults(),
	}

	if err := cfg.loadFile(cfgfile, &cfg); err != nil {
		return nil, err
	}

	if err := cfg.loadEnv(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *Config) loadFile(configFile string, out interface{}) error {
	finder := Finder{
		BasePaths:  []string{"/etc/ftpgrab/ftpgrab", "$XDG_CONFIG_HOME/ftpgrab", "$HOME/.config/ftpgrab", "./ftpgrab"},
		Extensions: []string{"yaml", "yml"},
	}

	filePath, err := finder.Find(configFile)
	if err != nil {
		return err
	}

	if len(filePath) == 0 {
		log.Debug().Msg("No configuration file defined")
		return nil
	}

	if err := file.Decode(filePath, out); err != nil {
		return errors.Wrap(err, "Failed to decode configuration from file")
	}

	log.Info().Msgf("Configuration loaded from file: %s", filePath)
	return nil
}

func (cfg *Config) loadEnv(out interface{}) error {
	var envvars []string
	for _, envvar := range env.FindPrefixedEnvVars(os.Environ(), "FTPGRAB_", out) {
		envvars = append(envvars, envvar)
	}
	if len(envvars) == 0 {
		log.Debug().Msg("No FTPGRAB_* environment variables defined")
		return nil
	}

	if err := env.Decode(envvars, "FTPGRAB_", out); err != nil {
		return errors.Wrap(err, "failed to decode configuration from environment variables")
	}

	return nil
}

func (cfg *Config) validate() error {
	if cfg.Db != nil && cfg.Db.Path != "" {
		if err := os.MkdirAll(path.Dir(cfg.Db.Path), os.ModePerm); err != nil {
			return errors.Wrap(err, "Cannot create database destination folder")
		}
	}

	if cfg.Server != nil {
		if cfg.Server.FTP == nil && cfg.Server.SFTP == nil {
			return errors.New("A server must be defined")
		} else if cfg.Server.FTP != nil && cfg.Server.SFTP != nil {
			return errors.New("Only one server is allowed")
		}
		if cfg.Server.FTP != nil {
			if len(cfg.Server.FTP.Sources) == 0 {
				return errors.New("At least one FTP source is required")
			}
		}
		if cfg.Server.SFTP != nil {
			if len(cfg.Server.SFTP.Sources) == 0 {
				return errors.New("At least one SFTP source is required")
			}
		}
	}

	if cfg.Download != nil {
		if err := os.MkdirAll(cfg.Download.Output, os.ModePerm); err != nil {
			return errors.Wrap(err, "Cannot create download output folder")
		}
		for _, include := range cfg.Download.Include {
			if _, err := regexp.Compile(include); err != nil {
				return errors.Wrapf(err, "Include regex '%s' cannot compile", include)
			}
		}
		for _, exclude := range cfg.Download.Exclude {
			if _, err := regexp.Compile(exclude); err != nil {
				return errors.Wrapf(err, "Exclude regex '%s' cannot compile", exclude)
			}
		}
	}

	validate := validator.New()
	return validate.Struct(cfg)
}

// String returns the string representation of configuration
func (cfg *Config) String() string {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return string(b)
}
