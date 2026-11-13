package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/httputil"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/crazy-max/ftpgrab/v7/internal/notif/notifier"
	"github.com/crazy-max/ftpgrab/v7/internal/secret"
	"github.com/hako/durafmt"
	"github.com/pkg/errors"
)

const discordMaxRateLimitAttempts = 3

// Client represents an active Discord notification object.
type Client struct {
	*notifier.Notifier
	cfg  *config.NotifDiscord
	meta config.Meta
}

// New creates a new Discord notification instance.
func New(cfg *config.NotifDiscord, meta config.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  cfg,
			meta: meta,
		},
	}
}

// Name returns the notifier's name.
func (c *Client) Name() string {
	return "discord"
}

// Send creates and sends a Discord notification with journal entries.
func (c *Client) Send(jnl journal.Journal) error {
	webhookURL, err := secret.GetSecret(c.cfg.WebhookURL, c.cfg.WebhookURLFile)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve webhook URL for Discord notifier")
	}

	content := fmt.Sprintf("FTPGrab has successfully downloaded **%d** files in **%s**.\n**%d** have been skipped and **%d** errors occurred.",
		jnl.Count.Success, durafmt.ParseShort(jnl.Duration).String(), jnl.Count.Skip, jnl.Count.Error)
	if len(c.cfg.Mentions) > 0 {
		content = strings.Join(c.cfg.Mentions, " ") + " " + content
	}

	var embeds []Embed
	if *c.cfg.RenderEmbeds {
		color := 0x4caf50
		if jnl.Count.Error > 0 {
			color = 0xb60205
		} else if jnl.Count.Success == 0 {
			color = 0xfbca04
		}
		var fields []EmbedField
		if *c.cfg.RenderFields {
			fields = []EmbedField{
				{
					Name:  "Server",
					Value: jnl.ServerHost,
				},
				{
					Name:  "Destination hostname",
					Value: c.meta.Hostname,
				},
			}
		}
		embeds = []Embed{
			{
				Color: color,
				Author: EmbedAuthor{
					Name:    c.meta.Name,
					URL:     c.meta.URL,
					IconURL: c.meta.Logo,
				},
				Fields: fields,
				Footer: EmbedFooter{
					Text: fmt.Sprintf("%s © %d %s %s", c.meta.Author, time.Now().Year(), c.meta.Name, c.meta.Version),
				},
			},
		}
	}
	payload, err := json.Marshal(Message{
		Content:   content,
		Username:  c.meta.Name,
		AvatarURL: c.meta.Logo,
		Embeds:    embeds,
	})
	if err != nil {
		return err
	}

	hc, err := httputil.NewClient(c.cfg.Proxy, c.cfg.TLSSkipVerify, c.cfg.TLSCACertFiles)
	if err != nil {
		return errors.Wrap(err, "cannot create HTTP client for Discord notifier")
	}
	defer hc.CloseIdleConnections()
	for attempt := 1; attempt <= discordMaxRateLimitAttempts; attempt++ {
		ctx, cancel := context.WithTimeoutCause(context.Background(), *c.cfg.Timeout, errors.WithStack(context.DeadlineExceeded))
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(payload))
		if err != nil {
			cancel()
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", c.meta.UserAgent)
		resp, err := hc.Do(req)
		if err != nil {
			cancel()
			return err
		}
		body, err := io.ReadAll(resp.Body)
		if closeErr := resp.Body.Close(); err == nil {
			err = closeErr
		}
		cancel()
		if err != nil {
			return errors.Wrap(err, "cannot read Discord response")
		}
		if resp.StatusCode == http.StatusNoContent {
			return nil
		}
		err = errors.Errorf("unexpected HTTP status %d: %s", resp.StatusCode, string(body))
		if resp.StatusCode != http.StatusTooManyRequests || attempt == discordMaxRateLimitAttempts {
			return err
		}
		time.Sleep(discordRetryAfter(resp, body))
	}
	return nil
}

// https://discord.com/developers/docs/topics/rate-limits
func discordRetryAfter(resp *http.Response, body []byte) time.Duration {
	if value := resp.Header.Get("Retry-After"); value != "" {
		if seconds, err := strconv.ParseFloat(value, 64); err == nil && seconds > 0 {
			return time.Duration(seconds * float64(time.Second))
		}
		if retryAt, err := http.ParseTime(value); err == nil {
			if delay := time.Until(retryAt); delay > 0 {
				return delay
			}
		}
	}
	var errBody struct {
		RetryAfter float64 `json:"retry_after"`
	}
	if err := json.Unmarshal(body, &errBody); err == nil && errBody.RetryAfter > 0 {
		return time.Duration(errBody.RetryAfter * float64(time.Second))
	}
	if value := resp.Header.Get("X-RateLimit-Reset-After"); value != "" {
		if seconds, err := strconv.ParseFloat(value, 64); err == nil && seconds > 0 {
			return time.Duration(seconds * float64(time.Second))
		}
	}
	return 5 * time.Second
}
