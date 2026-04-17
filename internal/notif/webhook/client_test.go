package webhook

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSend(t *testing.T) {
	var gotMethod string
	var gotUserAgent string
	var gotHeaders http.Header
	var gotBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		gotMethod = r.Method
		gotUserAgent = r.UserAgent()
		gotHeaders = r.Header.Clone()
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	client := &Client{
		cfg: &config.NotifWebhook{
			Endpoint: srv.URL,
			Method:   http.MethodPost,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"X-Test":       "true",
				"User-Agent":   "wrong",
			},
			Timeout: new(200 * time.Millisecond),
		},
		meta: config.Meta{
			Version:   "v1.2.3",
			Hostname:  "dest-host",
			UserAgent: "ftpgrab/test",
		},
	}

	err := client.Send(journal.Journal{
		ServerHost: "10.0.0.1",
		Entries: []journal.Entry{
			{File: "/shows/episode.mkv", Level: journal.EntryLevelSuccess, Text: "downloaded"},
		},
		Duration: 90 * time.Second,
	})
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "ftpgrab/test", gotUserAgent)
	assert.Equal(t, "true", gotHeaders.Get("X-Test"))
	assert.Equal(t, "v1.2.3", gotBody["ftpgrab_version"])
	assert.Equal(t, "10.0.0.1", gotBody["server_ip"])
	assert.Equal(t, "dest-host", gotBody["dest_hostname"])

	journalBody := gotBody["journal"].(map[string]any)
	assert.Equal(t, "1 minute", journalBody["duration"])
	entries := journalBody["entries"].([]any)
	require.Len(t, entries, 1)
	assert.Equal(t, "/shows/episode.mkv", entries[0].(map[string]any)["file"])
}

func TestTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	client := &Client{
		cfg: &config.NotifWebhook{
			Endpoint: srv.URL,
			Method:   http.MethodPost,
			Timeout:  new(10 * time.Millisecond),
		},
	}

	err := client.Send(journal.Journal{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}
