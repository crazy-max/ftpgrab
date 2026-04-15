package notif

import (
	"errors"
	"testing"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/crazy-max/ftpgrab/v7/internal/notif/notifier"
	"github.com/crazy-max/ftpgrab/v7/pkg/utl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	client, err := New(nil, config.Meta{})
	require.NoError(t, err)
	assert.Empty(t, client.notifiers)

	client, err = New(&config.Notif{
		Mail:    &config.NotifMail{},
		Script:  &config.NotifScript{},
		Slack:   &config.NotifSlack{},
		Webhook: &config.NotifWebhook{Timeout: utl.NewDuration(0)},
	}, config.Meta{})
	require.NoError(t, err)
	require.Len(t, client.notifiers, 4)
	assert.Equal(t, "mail", client.notifiers[0].Name())
	assert.Equal(t, "script", client.notifiers[1].Name())
	assert.Equal(t, "slack", client.notifiers[2].Name())
	assert.Equal(t, "webhook", client.notifiers[3].Name())
}

func TestSend(t *testing.T) {
	first := &stubNotifier{name: "first", err: errors.New("boom")}
	second := &stubNotifier{name: "second"}

	client := &Client{
		notifiers: []notifier.Notifier{
			{Handler: first},
			{Handler: second},
		},
	}

	client.Send(journal.Journal{})

	assert.Equal(t, 1, first.calls)
	assert.Equal(t, 1, second.calls)
}

type stubNotifier struct {
	name  string
	err   error
	calls int
}

func (n *stubNotifier) Name() string {
	return n.name
}

func (n *stubNotifier) Send(journal.Journal) error {
	n.calls++
	return n.err
}
