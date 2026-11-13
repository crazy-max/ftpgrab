package notif

import (
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/crazy-max/ftpgrab/v7/internal/notif/notifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	client, err := New(nil, config.Meta{})
	require.NoError(t, err)
	assert.Empty(t, client.notifiers)

	client, err = New(&config.Notif{
		Apprise: &config.NotifApprise{},
		Discord: &config.NotifDiscord{},
		Gotify:  &config.NotifGotify{},
		Mail:    &config.NotifMail{},
		Ntfy:    &config.NotifNtfy{},
		Script:  &config.NotifScript{},
		Slack:   &config.NotifSlack{},
		Webhook: &config.NotifWebhook{Timeout: new(time.Duration)},
	}, config.Meta{})
	require.NoError(t, err)
	require.Len(t, client.notifiers, 8)
	assert.Equal(t, "apprise", client.notifiers[0].Name())
	assert.Equal(t, "discord", client.notifiers[1].Name())
	assert.Equal(t, "gotify", client.notifiers[2].Name())
	assert.Equal(t, "mail", client.notifiers[3].Name())
	assert.Equal(t, "ntfy", client.notifiers[4].Name())
	assert.Equal(t, "script", client.notifiers[5].Name())
	assert.Equal(t, "slack", client.notifiers[6].Name())
	assert.Equal(t, "webhook", client.notifiers[7].Name())
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

func TestHTTPNotifierTLS(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/discord" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	certFile := filepath.Join(t.TempDir(), "ca.pem")
	require.NoError(t, os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: srv.Certificate().Raw,
	}), 0o600))
	tests := []struct {
		name       string
		skipVerify bool
		caFiles    []string
		wantError  bool
	}{
		{
			name:      "untrusted server",
			wantError: true,
		},
		{
			name:    "custom CA",
			caFiles: []string{certFile},
		},
		{
			name:       "skip verification",
			skipVerify: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(&config.Notif{
				Apprise: &config.NotifApprise{
					Endpoint:       srv.URL + "/apprise",
					Timeout:        new(time.Second),
					TLSSkipVerify:  tt.skipVerify,
					TLSCACertFiles: tt.caFiles,
				},
				Discord: &config.NotifDiscord{
					WebhookURL:     srv.URL + "/discord",
					RenderEmbeds:   new(false),
					Timeout:        new(time.Second),
					TLSSkipVerify:  tt.skipVerify,
					TLSCACertFiles: tt.caFiles,
				},
				Gotify: &config.NotifGotify{
					Endpoint:       srv.URL + "/gotify",
					Timeout:        new(time.Second),
					TLSSkipVerify:  tt.skipVerify,
					TLSCACertFiles: tt.caFiles,
				},
				Ntfy: &config.NotifNtfy{
					Endpoint:       srv.URL + "/ntfy",
					Timeout:        new(time.Second),
					TLSSkipVerify:  tt.skipVerify,
					TLSCACertFiles: tt.caFiles,
				},
				Slack: &config.NotifSlack{
					WebhookURL:     srv.URL + "/slack",
					Timeout:        new(time.Second),
					TLSSkipVerify:  tt.skipVerify,
					TLSCACertFiles: tt.caFiles,
				},
				Webhook: &config.NotifWebhook{
					Endpoint:       srv.URL + "/webhook",
					Method:         http.MethodPost,
					Timeout:        new(time.Second),
					TLSSkipVerify:  tt.skipVerify,
					TLSCACertFiles: tt.caFiles,
				},
			}, config.Meta{})
			require.NoError(t, err)
			require.Len(t, client.notifiers, 6)
			for _, n := range client.notifiers {
				t.Run(n.Name(), func(t *testing.T) {
					err := n.Send(journal.Journal{})
					if tt.wantError {
						require.Error(t, err)
					} else {
						require.NoError(t, err)
					}
				})
			}
		})
	}
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
