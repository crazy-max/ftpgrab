package config

import (
	"encoding/json"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/crazy-max/gonfig"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Config holds configuration details
type Config struct {
	Cli      Cli       `yaml:"-" json:"-" label:"-" file:"-"`
	Meta     Meta      `yaml:"-" json:"-" label:"-" file:"-"`
	Db       *Db       `yaml:"db,omitempty" json:"db,omitempty" validate:"omitempty"`
	Server   *Server   `yaml:"server,omitempty" json:"server,omitempty" validate:"required"`
	Download *Download `yaml:"download,omitempty" json:"download,omitempty" validate:"required"`
	Notif    *Notif    `yaml:"notif,omitempty" json:"notif,omitempty"`
}

// Load returns Config struct
func Load(cli Cli, meta Meta) (*Config, error) {
	cfg := Config{
		Cli:  cli,
		Meta: meta,
		Db:   (&Db{}).GetDefaults(),
	}

	fileLoader := gonfig.NewFileLoader(gonfig.FileLoaderConfig{
		Filename: cli.Cfgfile,
		Finder: gonfig.Finder{
			BasePaths:  []string{"/etc/ftpgrab/ftpgrab", "$XDG_CONFIG_HOME/ftpgrab", "$HOME/.config/ftpgrab", "./ftpgrab"},
			Extensions: []string{"yaml", "yml"},
		},
	})
	if found, err := fileLoader.Load(&cfg); err != nil {
		return nil, errors.Wrap(err, "Failed to decode configuration from file")
	} else if !found {
		log.Debug().Msg("No configuration file found")
	} else {
		log.Info().Msgf("Configuration loaded from file: %s", fileLoader.GetFilename())
	}

	envLoader := gonfig.NewEnvLoader(gonfig.EnvLoaderConfig{
		Prefix: "FTPGRAB_",
	})
	if found, err := envLoader.Load(&cfg); err != nil {
		return nil, errors.Wrap(err, "Failed to decode configuration from environment variables")
	} else if !found {
		log.Debug().Msg("No FTPGRAB_* environment variables defined")
	} else {
		log.Info().Msgf("Configuration loaded from %d environment variables", len(envLoader.GetVars()))
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *Config) validate() error {
	var err error

	if cfg.Db != nil {
		if len(cfg.Db.Path) > 0 {
			if err := os.MkdirAll(path.Dir(cfg.Db.Path), os.ModePerm); err != nil {
				return errors.Wrap(err, "Cannot create database destination folder")
			}
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
		if len(cfg.Download.Since) > 0 {
			cfg.Download.SinceTime, err = time.Parse("2006-01-02T15:04:05Z", cfg.Download.Since)
			if err != nil {
				return err
			}
		}
	}

	return validator.New().Struct(cfg)
}

// String returns the string representation of configuration
func (cfg *Config) String() string {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return string(b)
}
