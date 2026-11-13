package mail

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/crazy-max/ftpgrab/v7/internal/notif/notifier"
	"github.com/crazy-max/ftpgrab/v7/internal/secret"
	"github.com/hako/durafmt"
	"github.com/matcornic/hermes/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	email "github.com/wneessen/go-mail"
)

// Client represents an active mail notification object
type Client struct {
	*notifier.Notifier
	cfg  *config.NotifMail
	meta config.Meta
}

// New creates a new mail notification instance
func New(cfg *config.NotifMail, meta config.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  cfg,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "mail"
}

// Send creates and sends an email notification with journal entries
func (c *Client) Send(jnl journal.Journal) error {
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
			{Key: "Status", Value: string(entry.Level)},
			{Key: "Info", Value: entry.Text},
			{Key: "File", Value: entry.File},
		})
	}

	hermesEmail := hermes.Email{
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
	htmlpart, err := h.GenerateHTML(hermesEmail)
	if err != nil {
		return errors.Wrap(err, "Cannot generate HTML content for email notification")
	}

	// Generate the plaintext version of the e-mail (for clients that do not support xHTML)
	textpart, err := h.GeneratePlainText(hermesEmail)
	if err != nil {
		return errors.Wrap(err, "Cannot generate plaintext content for email notification")
	}

	msg := email.NewMsg()
	if err = msg.FromFormat(c.meta.Name, c.cfg.From); err != nil {
		return errors.Wrap(err, "cannot set mail FROM address")
	}
	if err = msg.To(c.cfg.To); err != nil {
		return errors.Wrap(err, "cannot set mail TO address")
	}
	msg.Subject(fmt.Sprintf("%s report for %s on %s",
		c.meta.Name,
		jnl.ServerHost,
		c.meta.Hostname,
	))
	msg.SetBodyString(email.TypeTextPlain, textpart)
	msg.AddAlternativeString(email.TypeTextHTML, htmlpart)

	username, err := secret.GetSecret(c.cfg.Username, c.cfg.UsernameFile)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot retrieve username secret for mail notifier")
	}
	password, err := secret.GetSecret(c.cfg.Password, c.cfg.PasswordFile)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot retrieve password secret for mail notifier")
	}

	client, err := c.mailClient(username, password)
	if err != nil {
		return errors.Wrap(err, "cannot create mail client")
	}

	if err = client.DialAndSend(msg); err != nil {
		if msg.IsDelivered() {
			return nil
		}
		return errors.Wrap(err, "cannot send mail notification")
	}
	return nil
}

func (c *Client) mailClient(username, password string) (*email.Client, error) {
	opts := []email.Option{
		email.WithPort(c.cfg.Port),
		email.WithTLSPolicy(email.TLSOpportunistic),
		email.WithoutNoop(),
	}
	if *c.cfg.SSL {
		opts = append(opts, email.WithSSL())
	}
	if *c.cfg.InsecureSkipVerify {
		opts = append(opts, email.WithTLSConfig(&tls.Config{
			ServerName:         c.cfg.Host,
			MinVersion:         email.DefaultTLSMinVersion,
			InsecureSkipVerify: true,
		}))
	}
	if username != "" {
		opts = append(opts,
			email.WithUsername(username),
			email.WithPassword(password),
			email.WithSMTPAuth(email.SMTPAuthAutoDiscover),
		)
	}
	return email.NewClient(c.cfg.Host, opts...)
}
