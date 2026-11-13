package slack

import (
	"bytes"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"html/template"
	"strconv"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/httputil"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/crazy-max/ftpgrab/v7/internal/notif/notifier"
	"github.com/crazy-max/ftpgrab/v7/internal/secret"
	"github.com/hako/durafmt"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
)

const slackMaxRateLimitAttempts = 3

// Client represents an active slack notification object
type Client struct {
	*notifier.Notifier
	cfg  *config.NotifSlack
	meta config.Meta
}

// New creates a new slack notification instance
func New(cfg *config.NotifSlack, meta config.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  cfg,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "slack"
}

// Send creates and sends a slack notification with journal entries
func (c *Client) Send(jnl journal.Journal) error {
	webhookURL, err := secret.GetSecret(c.cfg.WebhookURL, c.cfg.WebhookURLFile)
	if err != nil {
		return errors.Wrap(err, "cannot retrieve webhook URL for Slack notifier")
	}
	var textBuf bytes.Buffer
	textTpl := template.Must(template.New("text").Parse("FTPGrab has successfully downloaded *{{ .Success }}* files in *{{ .Duration }}*.\n*{{ .Skip }}* have been skipped and *{{ .Error }}* errors occurred."))
	if err := textTpl.Execute(&textBuf, struct {
		Success  int
		Skip     int
		Error    int
		Duration string
	}{
		jnl.Count.Success,
		jnl.Count.Skip,
		jnl.Count.Error,
		durafmt.ParseShort(jnl.Duration).String(),
	}); err != nil {
		return err
	}

	color := "#4caf50"
	if jnl.Count.Error > 0 {
		color = "#b60205"
	} else if jnl.Count.Success == 0 {
		color = "#fbca04"
	}

	payload := &slack.WebhookMessage{
		Attachments: []slack.Attachment{{
			Color:         color,
			AuthorName:    c.meta.Name,
			AuthorSubname: "github.com/crazy-max/ftpgrab",
			AuthorLink:    c.meta.URL,
			AuthorIcon:    c.meta.Logo,
			Text:          fmt.Sprintf("%s %s", "<!channel>", textBuf.String()),
			Footer:        fmt.Sprintf("%s © %d %s %s", c.meta.Author, time.Now().Year(), c.meta.Name, c.meta.Version),
			Fields: []slack.AttachmentField{
				{
					Title: "Server",
					Value: jnl.ServerHost,
					Short: false,
				},
				{
					Title: "Destination hostname",
					Value: c.meta.Hostname,
					Short: false,
				},
			},
			Ts: json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
		}},
	}

	hc, err := httputil.NewClient(c.cfg.Proxy, c.cfg.TLSSkipVerify, c.cfg.TLSCACertFiles)
	if err != nil {
		return errors.Wrap(err, "cannot create HTTP client for Slack notifier")
	}
	defer hc.CloseIdleConnections()
	hc.Timeout = *c.cfg.Timeout
	for attempt := 1; attempt <= slackMaxRateLimitAttempts; attempt++ {
		err = slack.PostWebhookCustomHTTP(webhookURL, &hc, payload)
		if err == nil {
			return nil
		}

		rateLimitErr, ok := stderrors.AsType[*slack.RateLimitedError](err)
		if !ok || attempt == slackMaxRateLimitAttempts {
			return err
		}

		time.Sleep(rateLimitErr.RetryAfter)
	}
	return nil
}
