package notif

import (
	"github.com/ftpgrab/ftpgrab/internal/journal"
	"github.com/ftpgrab/ftpgrab/internal/model"
	"github.com/ftpgrab/ftpgrab/internal/notif/mail"
	"github.com/ftpgrab/ftpgrab/internal/notif/notifier"
	"github.com/ftpgrab/ftpgrab/internal/notif/webhook"
	"github.com/rs/zerolog/log"
)

// Client represents an active webhook notification object
type Client struct {
	cfg       model.Notif
	app       model.App
	cmn       model.Common
	notifiers []notifier.Notifier
}

// New creates a new notification instance
func New(config model.Notif, app model.App, cmn model.Common) (*Client, error) {
	var c = &Client{
		cfg:       config,
		app:       app,
		cmn:       cmn,
		notifiers: []notifier.Notifier{},
	}

	// Add notifiers
	if config.Mail.Enable {
		c.notifiers = append(c.notifiers, mail.New(config.Mail, app, cmn))
	}
	if config.Webhook.Enable {
		c.notifiers = append(c.notifiers, webhook.New(config.Webhook, app, cmn))
	}

	log.Debug().Msgf("%d notifier(s) created", len(c.notifiers))
	return c, nil
}

// Send creates and sends notifications to notifiers
func (c *Client) Send(jnl journal.Client) {
	for _, n := range c.notifiers {
		log.Debug().Msgf("Sending %s notification...", n.Name())
		if err := n.Send(jnl); err != nil {
			log.Error().Err(err).Msgf("%s notification failed", n.Name())
		}
	}
}
