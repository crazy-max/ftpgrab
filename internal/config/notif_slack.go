package config

import "time"

// NotifSlack holds slack notification configuration details
type NotifSlack struct {
	Proxy      string         `yaml:"proxy,omitempty" json:"proxy,omitempty" validate:"omitempty,url"`
	WebhookURL string         `yaml:"webhookURL,omitempty" json:"webhookURL,omitempty" validate:"required"`
	Timeout    *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"required,gt=0"`
}

// GetDefaults gets the default values
func (s *NotifSlack) GetDefaults() *NotifSlack {
	n := &NotifSlack{}
	n.SetDefaults()
	return n
}

// SetDefaults sets the default values
func (s *NotifSlack) SetDefaults() {
	s.Timeout = new(10 * time.Second)
}
