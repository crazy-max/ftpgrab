package config

import "time"

// NotifNtfy holds ntfy notification configuration details.
type NotifNtfy struct {
	Endpoint  string         `yaml:"endpoint,omitempty" json:"endpoint,omitempty" validate:"required"`
	Token     string         `yaml:"token,omitempty" json:"token,omitempty"`
	TokenFile string         `yaml:"tokenFile,omitempty" json:"tokenFile,omitempty" validate:"omitempty,file"`
	Topic     string         `yaml:"topic,omitempty" json:"topic,omitempty" validate:"required"`
	Priority  int            `yaml:"priority,omitempty" json:"priority,omitempty" validate:"min=0,max=5"`
	Tags      []string       `yaml:"tags,omitempty" json:"tags,omitempty"`
	Icon      string         `yaml:"icon,omitempty" json:"icon,omitempty"`
	Click     string         `yaml:"click,omitempty" json:"click,omitempty"`
	Timeout   *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required,gt=0"`
	Proxy     string         `yaml:"proxy,omitempty" json:"proxy,omitempty" validate:"omitempty,url"`
}

// GetDefaults gets the default values.
func (s *NotifNtfy) GetDefaults() *NotifNtfy {
	n := &NotifNtfy{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values.
func (s *NotifNtfy) SetDefaults() {
	s.Endpoint = "https://ntfy.sh"
	s.Priority = 3
	s.Tags = []string{"package"}
	s.Timeout = new(10 * time.Second)
}
