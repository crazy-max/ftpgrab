package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/crazy-max/ftpgrab/v7/internal/notif/notifier"
)

// Client represents an active webhook notification object
type Client struct {
	*notifier.Notifier
	cfg  *config.NotifWebhook
	meta config.Meta
}

// New creates a new webhook notification instance
func New(config *config.NotifWebhook, meta config.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "webhook"
}

// Send creates and sends a webhook notification with journal entries
func (c *Client) Send(jnl journal.Journal) error {
	hc := http.Client{
		Timeout: *c.cfg.Timeout,
	}

	body, err := json.Marshal(struct {
		Version  string          `json:"ftpgrab_version,omitempty"`
		ServerIP string          `json:"server_ip,omitempty"`
		Dest     string          `json:"dest_hostname,omitempty"`
		Journal  journal.Journal `json:"journal,omitempty"`
	}{
		Version:  c.meta.Version,
		ServerIP: jnl.ServerHost,
		Dest:     c.meta.Hostname,
		Journal:  jnl,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(c.cfg.Method, c.cfg.Endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	if len(c.cfg.Headers) > 0 {
		for key, value := range c.cfg.Headers {
			req.Header.Add(key, value)
		}
	}

	req.Header.Set("User-Agent", c.meta.UserAgent)

	_, err = hc.Do(req)
	return err
}
