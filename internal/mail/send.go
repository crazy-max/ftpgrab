package mail

import (
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"github.com/ftpgrab/ftpgrab/internal/journal"
	"github.com/ftpgrab/ftpgrab/internal/model"
	"github.com/go-gomail/gomail"
	"github.com/hako/durafmt"
	"github.com/matcornic/hermes/v2"
)

// Send creates and sends an email with journal entries
func Send(jnl *journal.Client, app model.App, cmn model.Common, mail model.Mail) error {
	h := hermes.Hermes{
		Theme: new(Theme),
		Product: hermes.Product{
			Name: app.Name,
			Link: "https://ftpgrab.github.io",
			Logo: "https://ftpgrab.github.io/img/logo.png",
			Copyright: fmt.Sprintf("%s © 2014 - %d %s %s",
				app.Author,
				time.Now().Year(),
				app.Name,
				app.Version),
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
			Title: fmt.Sprintf("%s report", app.Name),
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
	msg.SetHeader("From", fmt.Sprintf("%s <%s>", app.Name, mail.From))
	msg.SetHeader("To", mail.To)
	msg.SetHeader("Subject", fmt.Sprintf("%s report for %s on %s",
		app.Name,
		cmn.Host,
		hostname,
	))
	msg.SetBody("text/plain", textpart)
	msg.AddAlternative("text/html", htmlpart)

	var tlsConfig *tls.Config
	if mail.InsecureSkipVerify {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: mail.InsecureSkipVerify,
		}
	}

	dialer := &gomail.Dialer{
		Host:      mail.Host,
		Port:      mail.Port,
		Username:  mail.Username,
		Password:  mail.Password,
		SSL:       mail.SSL,
		TLSConfig: tlsConfig,
	}

	return dialer.DialAndSend(msg)
}
