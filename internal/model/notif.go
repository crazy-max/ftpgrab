package model

// Notif holds data necessary for notification configuration
type Notif struct {
	Mail    Mail    `yaml:"mail,omitempty"`
	Webhook Webhook `yaml:"webhook,omitempty"`
}
