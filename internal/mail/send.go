package mail

import (
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"github.com/ftpgrab/ftpgrab/internal/config"
	"github.com/ftpgrab/ftpgrab/internal/journal"
	"github.com/go-gomail/gomail"
	"github.com/hako/durafmt"
	"github.com/matcornic/hermes/v2"
)

// Send creates and sends an email with journal entries
func Send(jnl *journal.Client, cfg *config.Configuration) error {
	h := hermes.Hermes{
		Theme: new(Theme),
		Product: hermes.Product{
			Name: cfg.App.Name,
			Link: "https://ftpgrab.github.io",
			Logo: "https://ftpgrab.github.io/img/logo.png",
			Copyright: fmt.Sprintf("%s © 2014 - %d %s %s",
				cfg.App.Author,
				time.Now().Year(),
				cfg.App.Name,
				cfg.App.Version),
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
			Title: fmt.Sprintf("%s report", cfg.App.Name),
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
			Signature: "Thanks for your support,",
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
	msg.SetHeader("From", fmt.Sprintf("%s <%s>", cfg.App.Name, cfg.Mail.From))
	msg.SetHeader("To", cfg.Mail.To)
	msg.SetHeader("Subject", fmt.Sprintf("%s report for %s on %s",
		cfg.App.Name,
		cfg.Ftp.Host,
		hostname,
	))
	msg.SetBody("text/plain", textpart)
	msg.AddAlternative("text/html", htmlpart)

	dialer := &gomail.Dialer{
		Host:     cfg.Mail.Host,
		Port:     cfg.Mail.Port,
		Username: cfg.Mail.Username,
		Password: cfg.Mail.Password,
		SSL:      cfg.Mail.Port == 465,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return dialer.DialAndSend(msg)
}
