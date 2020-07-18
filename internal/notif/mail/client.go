package mail

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/crazy-max/ftpgrab/v7/internal/model"
	"github.com/crazy-max/ftpgrab/v7/internal/notif/notifier"
	"github.com/crazy-max/ftpgrab/v7/pkg/utl"
	"github.com/go-gomail/gomail"
	"github.com/hako/durafmt"
	"github.com/matcornic/hermes/v2"
	"github.com/rs/zerolog/log"
)

// Client represents an active mail notification object
type Client struct {
	*notifier.Notifier
	cfg  *model.NotifMail
	meta model.Meta
}

// New creates a new mail notification instance
func New(config *model.NotifMail, meta model.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
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
			Name: c.meta.Name,
			Link: c.meta.URL,
			Logo: c.meta.Logo,
			Copyright: fmt.Sprintf("%s © %d %s %s",
				c.meta.Author,
				time.Now().Year(),
				c.meta.Name,
				c.meta.Version),
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
			Title: fmt.Sprintf("%s ⚡️ report", c.meta.Name),
			FreeMarkdown: hermes.Markdown(fmt.Sprintf(
				`**%d** files have been download successfully, **%d** have been skipped and **%d** errors occurred in %s.`,
				jnl.Count.Success,
				jnl.Count.Skip,
				jnl.Count.Error,
				durafmt.ParseShort(jnl.Duration).String())),
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

	msg := gomail.NewMessage()
	msg.SetHeader("From", fmt.Sprintf("%s <%s>", c.meta.Name, c.cfg.From))
	msg.SetHeader("To", c.cfg.To)
	msg.SetHeader("Subject", fmt.Sprintf("%s report for %s on %s",
		c.meta.Name,
		jnl.ServerHost,
		c.meta.Hostname,
	))
	msg.SetBody("text/plain", textpart)
	msg.AddAlternative("text/html", htmlpart)

	var tlsConfig *tls.Config
	if *c.cfg.InsecureSkipVerify {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: *c.cfg.InsecureSkipVerify,
		}
	}

	username, err := utl.GetSecret(c.cfg.Username, c.cfg.UsernameFile)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot retrieve username secret for mail notifier")
	}
	password, err := utl.GetSecret(c.cfg.Password, c.cfg.PasswordFile)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot retrieve password secret for mail notifier")
	}

	dialer := &gomail.Dialer{
		Host:      c.cfg.Host,
		Port:      c.cfg.Port,
		Username:  username,
		Password:  password,
		SSL:       *c.cfg.SSL,
		TLSConfig: tlsConfig,
	}

	return dialer.DialAndSend(msg)
}
