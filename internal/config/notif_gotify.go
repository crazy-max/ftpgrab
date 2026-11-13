package config

import "time"

// NotifGotify holds Gotify notification configuration details.
type NotifGotify struct {
	Endpoint  string         `yaml:"endpoint,omitempty" json:"endpoint,omitempty" validate:"required"`
	Token     string         `yaml:"token,omitempty" json:"token,omitempty" validate:"required_without=TokenFile"`
	TokenFile string         `yaml:"tokenFile,omitempty" json:"tokenFile,omitempty" validate:"omitempty,file"`
	Priority  int            `yaml:"priority,omitempty" json:"priority,omitempty" validate:"min=0"`
	Timeout   *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required,gt=0"`
	Proxy     string         `yaml:"proxy,omitempty" json:"proxy,omitempty" validate:"omitempty,url"`
}

// GetDefaults gets the default values.
func (s *NotifGotify) GetDefaults() *NotifGotify {
	n := &NotifGotify{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values.
func (s *NotifGotify) SetDefaults() {
	s.Priority = 1
	s.Timeout = new(10 * time.Second)
}
