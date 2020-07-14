package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

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

func (cfg *Config) loadFile(cfgfile string, out interface{}) error {
	if len(cfgfile) == 0 {
		log.Debug().Msg("No configuration file defined")
		return nil
	}

	if _, err := os.Lstat(cfgfile); os.IsNotExist(err) {
		return fmt.Errorf("config file %s not found", cfgfile)
	}

	if err := file.Decode(cfgfile, out); err != nil {
		return errors.Wrap(err, "failed to decode configuration from file")
	}

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

	if cfg.Server == nil || (cfg.Server.FTP == nil && cfg.Server.SFTP == nil) {
		return errors.New("a server must be defined")
	} else if cfg.Server != nil && cfg.Server.FTP != nil && cfg.Server.SFTP != nil {
		return errors.New("only one server is allowed")
	}

	if cfg.Download != nil && cfg.Download.Output != "" {
		if err := os.MkdirAll(cfg.Download.Output, os.ModePerm); err != nil {
			return errors.Wrap(err, "Cannot create download output folder")
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
