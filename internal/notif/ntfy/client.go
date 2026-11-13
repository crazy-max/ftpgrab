package ntfy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/httputil"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/crazy-max/ftpgrab/v7/internal/notif/notifier"
	"github.com/crazy-max/ftpgrab/v7/internal/secret"
	"github.com/hako/durafmt"
	"github.com/pkg/errors"
)

// Client represents an active ntfy notification object.
type Client struct {
	*notifier.Notifier
	cfg  *config.NotifNtfy
	meta config.Meta
}

// New creates a new ntfy notification instance.
func New(cfg *config.NotifNtfy, meta config.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  cfg,
			meta: meta,
		},
	}
}

// Name returns the notifier's name.
func (c *Client) Name() string {
	return "ntfy"
}

// Send creates and sends an ntfy notification with journal entries.
func (c *Client) Send(jnl journal.Journal) error {
	token, err := secret.GetSecret(c.cfg.Token, c.cfg.TokenFile)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve token secret for ntfy notifier")
	}
	icon := c.cfg.Icon
	if icon == "" {
		icon = c.meta.Logo
	}
	body, err := json.Marshal(struct {
		Topic    string   `json:"topic"`
		Message  string   `json:"message"`
		Title    string   `json:"title"`
		Priority int      `json:"priority"`
		Tags     []string `json:"tags"`
		Icon     string   `json:"icon,omitempty"`
		Click    string   `json:"click,omitempty"`
		Markdown bool     `json:"markdown"`
	}{
		Topic: c.cfg.Topic,
		Message: fmt.Sprintf("FTPGrab has successfully downloaded **%d** files in **%s**.\n**%d** have been skipped and **%d** errors occurred.",
			jnl.Count.Success, durafmt.ParseShort(jnl.Duration).String(), jnl.Count.Skip, jnl.Count.Error),
		Title:    fmt.Sprintf("%s report for %s on %s", c.meta.Name, jnl.ServerHost, c.meta.Hostname),
		Priority: c.cfg.Priority,
		Tags:     c.cfg.Tags,
		Icon:     icon,
		Click:    c.cfg.Click,
		Markdown: true,
	})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeoutCause(context.Background(), *c.cfg.Timeout, errors.WithStack(context.DeadlineExceeded))
	defer cancel()
	hc, err := httputil.NewClient(c.cfg.Proxy, c.cfg.TLSSkipVerify, c.cfg.TLSCACertFiles)
	if err != nil {
		return errors.Wrap(err, "cannot create HTTP client for ntfy notifier")
	}
	defer hc.CloseIdleConnections()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.Endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.meta.UserAgent)
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "cannot read ntfy error response")
		}
		return errors.Errorf("unexpected HTTP status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
