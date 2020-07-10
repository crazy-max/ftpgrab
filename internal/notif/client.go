package notif

import (
	"github.com/ftpgrab/ftpgrab/v7/internal/journal"
	"github.com/ftpgrab/ftpgrab/v7/internal/model"
	"github.com/ftpgrab/ftpgrab/v7/internal/notif/mail"
	"github.com/ftpgrab/ftpgrab/v7/internal/notif/notifier"
	"github.com/ftpgrab/ftpgrab/v7/internal/notif/slack"
	"github.com/ftpgrab/ftpgrab/v7/internal/notif/webhook"
	"github.com/rs/zerolog/log"
)

// Client represents an active webhook notification object
type Client struct {
	cfg       *model.Notif
	meta      model.Meta
	notifiers []notifier.Notifier
}

// New creates a new notification instance
func New(config *model.Notif, meta model.Meta) (*Client, error) {
	var c = &Client{
		cfg:       config,
		meta:      meta,
		notifiers: []notifier.Notifier{},
	}

	if config == nil {
		log.Warn().Msg("No notifier available")
		return c, nil
	}

	// Add notifiers
	if config.Mail != nil {
		c.notifiers = append(c.notifiers, mail.New(config.Mail, meta))
	}
	if config.Slack != nil {
		c.notifiers = append(c.notifiers, slack.New(config.Slack, meta))
	}
	if config.Webhook != nil {
		c.notifiers = append(c.notifiers, webhook.New(config.Webhook, meta))
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
