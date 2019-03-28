package mail

import (
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"github.com/ftpgrab/ftpgrab/internal/journal"
	"github.com/ftpgrab/ftpgrab/internal/model"
	"github.com/ftpgrab/ftpgrab/internal/notif/notifier"
	"github.com/go-gomail/gomail"
	"github.com/hako/durafmt"
	"github.com/matcornic/hermes/v2"
)

// Client represents an active mail notification object
type Client struct {
	*notifier.Notifier
	cfg model.Mail
	app model.App
	cmn model.Common
}

// New creates a new mail notification instance
func New(config model.Mail, app model.App, cmn model.Common) notifier.Notifier {
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
	return "mail"
}

// Send creates and sends an email notification with journal entries
func (c *Client) Send(jnl journal.Client) error {
	h := hermes.Hermes{
		Theme: new(Theme),
		Product: hermes.Product{
			Name: c.app.Name,
			Link: "https://ftpgrab.github.io",
			Logo: "https://ftpgrab.github.io/img/logo.png",
			Copyright: fmt.Sprintf("%s © 2014 - %d %s %s",
				c.app.Author,
				time.Now().Year(),
				c.app.Name,
				c.app.Version),
		},
	}

	var entriesData [][]hermes.Entry
	for _, entry := range jnl.Entries {
		entriesData = append(entriesData, []hermes.Entry{
			{Key: "Status", Value: entry.StatusType},
			{Key: "Info", Value: string(entry.StatusText)},
			{Key: "File", Value: entry.File},
		})
	}

	email := hermes.Email{
		Body: hermes.Body{
			Title: fmt.Sprintf("%s report", c.app.Name),
			FreeMarkdown: hermes.Markdown(fmt.Sprintf(
				`**%d** files have been download successfully, **%d** have been skipped and **%d** errors occurred in %s.`,
				jnl.Count.Success,
				jnl.Count.Skip,
				jnl.Count.Error,
				durafmt.ParseShort(time.Duration(jnl.Duration)).String())),
			Table: hermes.Table{
				Data: entriesData,
				Columns: hermes.Columns{
					CustomWidth: map[string]string{
						"Status": "5%",
						"Info":   "20%",
					},
					CustomAlignment: map[string]string{
						"Status": "center",
					},
				},
			},
			Signature: "Thanks for your support",
		},
	}

	// Generate an HTML email with the provided contents (for modern clients)
	htmlpart, err := h.GenerateHTML(email)
	if err != nil {
		return fmt.Errorf("hermes: %v", err)
	}

	// Generate the plaintext version of the e-mail (for clients that do not support xHTML)
	textpart, err := h.GeneratePlainText(email)
	if err != nil {
		return fmt.Errorf("hermes: %v", err)
	}

	hostname, _ := os.Hostname()
	msg := gomail.NewMessage()
	msg.SetHeader("From", fmt.Sprintf("%s <%s>", c.app.Name, c.cfg.From))
	msg.SetHeader("To", c.cfg.To)
	msg.SetHeader("Subject", fmt.Sprintf("%s report for %s on %s",
		c.app.Name,
		c.cmn.Host,
		hostname,
	))
	msg.SetBody("text/plain", textpart)
	msg.AddAlternative("text/html", htmlpart)

	var tlsConfig *tls.Config
	if c.cfg.InsecureSkipVerify {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: c.cfg.InsecureSkipVerify,
		}
	}

	dialer := &gomail.Dialer{
		Host:      c.cfg.Host,
		Port:      c.cfg.Port,
		Username:  c.cfg.Username,
		Password:  c.cfg.Password,
		SSL:       c.cfg.SSL,
		TLSConfig: tlsConfig,
	}

	return dialer.DialAndSend(msg)
}
