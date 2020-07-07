package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"time"

	"github.com/ftpgrab/ftpgrab/internal/journal"
	"github.com/ftpgrab/ftpgrab/internal/model"
	"github.com/ftpgrab/ftpgrab/internal/notif/notifier"
	"github.com/hako/durafmt"
	"github.com/nlopes/slack"
)

// Client represents an active slack notification object
type Client struct {
	*notifier.Notifier
	cfg model.NotifSlack
	app model.App
	cmn model.Common
}

// New creates a new slack notification instance
func New(config model.NotifSlack, app model.App, cmn model.Common) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg: config,
			app: app,
			cmn: cmn,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "slack"
}

// Send creates and sends a slack notification with journal entries
func (c *Client) Send(jnl journal.Client) error {
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

	hostname, _ := os.Hostname()

	return slack.PostWebhook(c.cfg.WebhookURL, &slack.WebhookMessage{
		Attachments: []slack.Attachment{slack.Attachment{
			Color:         color,
			AuthorName:    "FTPGrab",
			AuthorSubname: "github.com/ftpgrab/ftpgrab",
			AuthorLink:    "https://github.com/ftpgrab/ftpgrab",
			AuthorIcon:    "https://raw.githubusercontent.com/ftpgrab/ftpgrab/master/.res/ftpgrab.png",
			Text:          fmt.Sprintf("%s %s", "<!channel>", textBuf.String()),
			Footer:        fmt.Sprintf("%s © %d %s %s", c.app.Author, time.Now().Year(), c.app.Name, c.app.Version),
			Fields: []slack.AttachmentField{
				{
					Title: "Server",
					Value: c.cmn.Host,
					Short: false,
				},
				{
					Title: "Destination hostname",
					Value: hostname,
					Short: false,
				},
			},
			Ts: json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
		}},
	})
}
