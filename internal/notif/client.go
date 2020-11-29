package notif

import (
	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/crazy-max/ftpgrab/v7/internal/notif/mail"
	"github.com/crazy-max/ftpgrab/v7/internal/notif/notifier"
	"github.com/crazy-max/ftpgrab/v7/internal/notif/script"
	"github.com/crazy-max/ftpgrab/v7/internal/notif/slack"
	"github.com/crazy-max/ftpgrab/v7/internal/notif/webhook"
	"github.com/rs/zerolog/log"
)

// Client represents an active notification object
type Client struct {
	cfg       *config.Notif
	meta      config.Meta
	notifiers []notifier.Notifier
}

// New creates a new notification instance
func New(cfg *config.Notif, meta config.Meta) (*Client, error) {
	var c = &Client{
		cfg:       cfg,
		meta:      meta,
		notifiers: []notifier.Notifier{},
	}

	if cfg == nil {
		log.Warn().Msg("No notifier available")
		return c, nil
	}

	// Add notifiers
	if cfg.Mail != nil {
		c.notifiers = append(c.notifiers, mail.New(cfg.Mail, meta))
	}
	if cfg.Script != nil {
		c.notifiers = append(c.notifiers, script.New(cfg.Script, meta))
	}
	if cfg.Slack != nil {
		c.notifiers = append(c.notifiers, slack.New(cfg.Slack, meta))
	}
	if cfg.Webhook != nil {
		c.notifiers = append(c.notifiers, webhook.New(cfg.Webhook, meta))
	}

	log.Debug().Msgf("%d notifier(s) created", len(c.notifiers))
	return c, nil
}

// Send creates and sends notifications to notifiers
func (c *Client) Send(jnl journal.Journal) {
	for _, n := range c.notifiers {
		log.Debug().Msgf("Sending %s notification...", n.Name())
		if err := n.Send(jnl); err != nil {
			log.Error().Err(err).Msgf("%s notification failed", n.Name())
		}
	}
}
