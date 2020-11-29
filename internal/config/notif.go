package config

// Notif holds data necessary for notification configuration
type Notif struct {
	Mail    *NotifMail    `yaml:"mail,omitempty" json:"mail,omitempty"`
	Slack   *NotifSlack   `yaml:"slack,omitempty" json:"slack,omitempty"`
	Script  *NotifScript  `yaml:"script,omitempty" json:"script,omitempty"`
	Webhook *NotifWebhook `yaml:"webhook,omitempty" json:"webhook,omitempty"`
}

// GetDefaults gets the default values
func (s *Notif) GetDefaults() *Notif {
	return nil
}

// SetDefaults sets the default values
func (s *Notif) SetDefaults() {
	// noop
}
