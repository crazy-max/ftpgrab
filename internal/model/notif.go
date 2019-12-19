package model

// Notif holds data necessary for notification configuration
type Notif struct {
	Mail    NotifMail    `yaml:"mail,omitempty"`
	Slack   NotifSlack   `yaml:"slack,omitempty"`
	Webhook NotifWebhook `yaml:"webhook,omitempty"`
}

// Mail holds mail notification configuration details
type NotifMail struct {
	Enable             bool   `yaml:"enable,omitempty"`
	Host               string `yaml:"host,omitempty"`
	Port               int    `yaml:"port,omitempty"`
	SSL                bool   `yaml:"ssl,omitempty"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify,omitempty"`
	Username           string `yaml:"username,omitempty"`
	Password           string `yaml:"password,omitempty"`
	From               string `yaml:"from,omitempty"`
	To                 string `yaml:"to,omitempty"`
}

// NotifSlack holds slack notification configuration details
type NotifSlack struct {
	Enable     bool   `yaml:"enable,omitempty"`
	WebhookURL string `yaml:"webhook_url,omitempty"`
}

// NotifWebhook holds webhook notification configuration details
type NotifWebhook struct {
	Enable   bool              `yaml:"enable,omitempty"`
	Endpoint string            `yaml:"endpoint,omitempty"`
	Method   string            `yaml:"method,omitempty"`
	Headers  map[string]string `yaml:"headers,omitempty"`
	Timeout  int               `yaml:"timeout,omitempty"`
}
