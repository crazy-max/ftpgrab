package apprise

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/httputil"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/crazy-max/ftpgrab/v7/internal/notif/notifier"
	"github.com/crazy-max/ftpgrab/v7/internal/secret"
	"github.com/hako/durafmt"
	"github.com/pkg/errors"
)

// Client represents an active Apprise notification object.
type Client struct {
	*notifier.Notifier
	cfg  *config.NotifApprise
	meta config.Meta
}

// New creates a new Apprise notification instance.
func New(cfg *config.NotifApprise, meta config.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  cfg,
			meta: meta,
		},
	}
}

// Name returns the notifier's name.
func (c *Client) Name() string {
	return "apprise"
}

// Send creates and sends an Apprise notification with journal entries.
func (c *Client) Send(jnl journal.Journal) error {
	token, err := secret.GetSecret(c.cfg.Token, c.cfg.TokenFile)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve token secret for Apprise notifier")
	}
	body, err := json.Marshal(struct {
		Message string   `json:"body"`
		Title   string   `json:"title"`
		Tags    []string `json:"tags"`
		URLs    []string `json:"urls"`
	}{
		Message: fmt.Sprintf("FTPGrab has successfully downloaded %d files in %s.\n%d have been skipped and %d errors occurred.",
			jnl.Count.Success, durafmt.ParseShort(jnl.Duration).String(), jnl.Count.Skip, jnl.Count.Error),
		Title: fmt.Sprintf("%s report for %s on %s", c.meta.Name, jnl.ServerHost, c.meta.Hostname),
		Tags:  c.cfg.Tags,
		URLs:  c.cfg.URLs,
	})
	if err != nil {
		return err
	}
	u, err := url.Parse(c.cfg.Endpoint)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "notify")
	if token != "" {
		u.Path = path.Join(u.Path, token)
	}

	ctx, cancel := context.WithTimeoutCause(context.Background(), *c.cfg.Timeout, errors.WithStack(context.DeadlineExceeded))
	defer cancel()
	hc, err := httputil.NewClient(c.cfg.Proxy)
	if err != nil {
		return errors.Wrap(err, "cannot create HTTP client for Apprise notifier")
	}
	defer hc.CloseIdleConnections()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return err
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
			return errors.Wrap(err, "cannot read Apprise error response")
		}
		return errors.Errorf("unexpected HTTP status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
