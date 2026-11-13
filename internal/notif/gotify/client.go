package gotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

// Client represents an active Gotify notification object.
type Client struct {
	*notifier.Notifier
	cfg  *config.NotifGotify
	meta config.Meta
}

// New creates a new Gotify notification instance.
func New(cfg *config.NotifGotify, meta config.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  cfg,
			meta: meta,
		},
	}
}

// Name returns the notifier's name.
func (c *Client) Name() string {
	return "gotify"
}

// Send creates and sends a Gotify notification with journal entries.
func (c *Client) Send(jnl journal.Journal) error {
	token, err := secret.GetSecret(c.cfg.Token, c.cfg.TokenFile)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve token secret for Gotify notifier")
	}
	body, err := json.Marshal(struct {
		Message  string                       `json:"message"`
		Title    string                       `json:"title"`
		Priority int                          `json:"priority"`
		Extras   map[string]map[string]string `json:"extras"`
	}{
		Message: fmt.Sprintf("FTPGrab has successfully downloaded **%d** files in **%s**.\n**%d** have been skipped and **%d** errors occurred.",
			jnl.Count.Success, durafmt.ParseShort(jnl.Duration).String(), jnl.Count.Skip, jnl.Count.Error),
		Title:    fmt.Sprintf("%s report for %s on %s", c.meta.Name, jnl.ServerHost, c.meta.Hostname),
		Priority: c.cfg.Priority,
		Extras: map[string]map[string]string{
			"client::display": {
				"contentType": "text/markdown",
			},
		},
	})
	if err != nil {
		return err
	}
	u, err := url.Parse(c.cfg.Endpoint)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "message")

	ctx, cancel := context.WithTimeoutCause(context.Background(), *c.cfg.Timeout, errors.WithStack(context.DeadlineExceeded))
	defer cancel()
	hc, err := httputil.NewClient(c.cfg.Proxy, c.cfg.TLSSkipVerify, c.cfg.TLSCACertFiles)
	if err != nil {
		return errors.Wrap(err, "cannot create HTTP client for Gotify notifier")
	}
	defer hc.CloseIdleConnections()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.meta.UserAgent)
	req.Header.Set("X-Gotify-Key", token)
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var errBody struct {
			Error            string `json:"error"`
			ErrorCode        int    `json:"errorCode"`
			ErrorDescription string `json:"errorDescription"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			return errors.Wrapf(err, "cannot decode JSON error response for HTTP %d %s status", resp.StatusCode, http.StatusText(resp.StatusCode))
		}
		return errors.Errorf("%d %s: %s", errBody.ErrorCode, errBody.Error, errBody.ErrorDescription)
	}
	return nil
}
