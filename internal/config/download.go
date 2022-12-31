package config

import (
	"os"
	"time"

	"github.com/crazy-max/ftpgrab/v7/pkg/utl"
)

// Download holds download configuration details
type Download struct {
	Output        string      `yaml:"output,omitempty" json:"output,omitempty" validate:"required,dir"`
	UID           int         `yaml:"uid,omitempty" json:"uid,omitempty"`
	GID           int         `yaml:"gid,omitempty" json:"gid,omitempty"`
	ChmodFile     os.FileMode `yaml:"chmodFile,omitempty" json:"chmodFile,omitempty"`
	ChmodDir      os.FileMode `yaml:"chmodDir,omitempty" json:"chmodDir,omitempty"`
	Include       []string    `yaml:"include,omitempty" json:"include,omitempty"`
	Exclude       []string    `yaml:"exclude,omitempty" json:"exclude,omitempty"`
	Since         string      `yaml:"since,omitempty" json:"since,omitempty"`
	SinceTime     time.Time   `yaml:"-" json:"-" label:"-" file:"-"`
	Retry         int         `yaml:"retry,omitempty" json:"retry,omitempty"`
	HideSkipped   *bool       `yaml:"hideSkipped,omitempty" json:"hideSkipped,omitempty"`
	TempFirst     *bool       `yaml:"tempFirst,omitempty" json:"tempFirst,omitempty"`
	CreateBaseDir *bool       `yaml:"createBaseDir,omitempty" json:"createBaseDir,omitempty"`
}

// GetDefaults gets the default values
func (s *Download) GetDefaults() *Download {
	n := &Download{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *Download) SetDefaults() {
	s.UID = os.Getuid()
	s.GID = os.Getgid()
	s.ChmodFile = 0o644
	s.ChmodDir = 0o755
	s.Retry = 3
	s.HideSkipped = utl.NewFalse()
	s.TempFirst = utl.NewFalse()
	s.CreateBaseDir = utl.NewFalse()
}
