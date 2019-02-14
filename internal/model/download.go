package model

import "time"

// Download holds download configuration details
type Download struct {
	Output        string    `yaml:"output,omitempty"`
	UID           int       `yaml:"uid,omitempty"`
	GID           int       `yaml:"gid,omitempty"`
	ChmodFile     int       `yaml:"chmod_file,omitempty"`
	ChmodDir      int       `yaml:"chmod_dir,omitempty"`
	Include       []string  `yaml:"include,omitempty"`
	Exclude       []string  `yaml:"exclude,omitempty"`
	Since         time.Time `yaml:"since,omitempty"`
	Retry         int       `yaml:"retry,omitempty"`
	HideSkipped   bool      `yaml:"hide_skipped,omitempty"`
	CreateBasedir bool      `yaml:"create_basedir,omitempty"`
}
