package config

import "time"

// NotifApprise holds Apprise notification configuration details.
type NotifApprise struct {
	Endpoint  string         `yaml:"endpoint,omitempty" json:"endpoint,omitempty" validate:"required"`
	Token     string         `yaml:"token,omitempty" json:"token,omitempty"`
	TokenFile string         `yaml:"tokenFile,omitempty" json:"tokenFile,omitempty" validate:"omitempty,file"`
	Tags      []string       `yaml:"tags,omitempty" json:"tags,omitempty"`
	URLs      []string       `yaml:"urls,omitempty" json:"urls,omitempty"`
	Timeout   *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required,gt=0"`
	Proxy     string         `yaml:"proxy,omitempty" json:"proxy,omitempty" validate:"omitempty,url"`
}

// GetDefaults gets the default values.
func (s *NotifApprise) GetDefaults() *NotifApprise {
	n := &NotifApprise{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values.
func (s *NotifApprise) SetDefaults() {
	s.Timeout = new(10 * time.Second)
}
