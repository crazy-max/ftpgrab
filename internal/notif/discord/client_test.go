package discord

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSend(t *testing.T) {
	var payload Message
	var method string
	var headers http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		headers = r.Header.Clone()
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	client := newTestClient(srv.URL)
	client.cfg.Mentions = []string{"<@123>", "<@&456>"}
	jnl := journal.Journal{ServerHost: "ftp.example.com", Duration: time.Minute}
	jnl.Count.Success = 2
	jnl.Count.Skip = 1
	jnl.Count.Error = 3
	require.NoError(t, client.Send(jnl))
	assert.Equal(t, http.MethodPost, method)
	assert.Equal(t, "application/json", headers.Get("Content-Type"))
	assert.Equal(t, "ftpgrab-test", headers.Get("User-Agent"))
	assert.Equal(t, "<@123> <@&456> FTPGrab has successfully downloaded **2** files in **1 minute**.\n**1** have been skipped and **3** errors occurred.", payload.Content)
	assert.Equal(t, "FTPGrab", payload.Username)
	assert.Equal(t, client.meta.Logo, payload.AvatarURL)
	require.Len(t, payload.Embeds, 1)
	embed := payload.Embeds[0]
	assert.Equal(t, 0xb60205, embed.Color)
	assert.Equal(t, EmbedAuthor{
		Name:    client.meta.Name,
		URL:     client.meta.URL,
		IconURL: client.meta.Logo,
	}, embed.Author)
	assert.Equal(t, fmt.Sprintf("CrazyMax © %d FTPGrab 7.0.0", time.Now().Year()), embed.Footer.Text)
	assert.Equal(t, []EmbedField{
		{
			Name:  "Server",
			Value: "ftp.example.com",
		},
		{
			Name:  "Destination hostname",
			Value: "destination",
		},
	}, embed.Fields)
}

func TestRenderOptions(t *testing.T) {
	tests := []struct {
		name         string
		renderEmbeds bool
		renderFields bool
		success      int
		wantColor    int
	}{
		{
			name:         "successful downloads",
			renderEmbeds: true,
			renderFields: true,
			success:      1,
			wantColor:    0x4caf50,
		},
		{
			name:         "no downloads without fields",
			renderEmbeds: true,
			wantColor:    0xfbca04,
		},
		{
			name:         "embeds disabled",
			renderFields: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var payload Message
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Errorf("decode payload: %v", err)
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			t.Cleanup(srv.Close)
			client := newTestClient(srv.URL)
			client.cfg.RenderEmbeds = new(tt.renderEmbeds)
			client.cfg.RenderFields = new(tt.renderFields)
			jnl := journal.Journal{}
			jnl.Count.Success = tt.success
			require.NoError(t, client.Send(jnl))
			assert.NotEmpty(t, payload.Content)
			if !tt.renderEmbeds {
				assert.Empty(t, payload.Embeds)
				return
			}
			require.Len(t, payload.Embeds, 1)
			assert.Equal(t, tt.wantColor, payload.Embeds[0].Color)
			if tt.renderFields {
				assert.Len(t, payload.Embeds[0].Fields, 2)
			} else {
				assert.Empty(t, payload.Embeds[0].Fields)
			}
		})
	}
}

func TestDeliveryResponses(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		succeed   bool
		attempts  int32
		wantError string
	}{
		{
			name:     "retry rate limit",
			status:   http.StatusTooManyRequests,
			succeed:  true,
			attempts: 2,
		},
		{
			name:      "stop after three attempts",
			status:    http.StatusTooManyRequests,
			attempts:  3,
			wantError: "unexpected HTTP status 429",
		},
		{
			name:      "do not retry bad request",
			status:    http.StatusBadRequest,
			attempts:  1,
			wantError: "unexpected HTTP status 400",
		},
		{
			name:      "do not retry server error",
			status:    http.StatusInternalServerError,
			attempts:  1,
			wantError: "unexpected HTTP status 500",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attempts atomic.Int32
			var payloads []Message
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var payload Message
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Errorf("decode payload: %v", err)
				}
				payloads = append(payloads, payload)
				if attempts.Add(1) > 1 && tt.succeed {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(`{"message":"response details","retry_after":0.001}`))
			}))
			t.Cleanup(srv.Close)
			err := newTestClient(srv.URL).Send(journal.Journal{})
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				assert.Contains(t, err.Error(), "response details")
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.attempts, attempts.Load())
			for _, payload := range payloads[1:] {
				assert.Equal(t, payloads[0], payload)
			}
		})
	}
}

func TestWebhookURLFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	filename := filepath.Join(t.TempDir(), "webhook_url")
	require.NoError(t, os.WriteFile(filename, []byte(srv.URL), 0o600))
	client := newTestClient("")
	client.cfg.WebhookURLFile = filename
	require.NoError(t, client.Send(journal.Journal{}))
	client.cfg.WebhookURL = srv.URL
	client.cfg.WebhookURLFile = filename + ".missing"
	require.NoError(t, client.Send(journal.Journal{}))
	client.cfg.WebhookURL = ""
	require.ErrorContains(t, client.Send(journal.Journal{}), "cannot retrieve webhook URL for Discord notifier")
}

func TestTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	client := newTestClient(srv.URL)
	client.cfg.Timeout = new(10 * time.Millisecond)
	require.ErrorContains(t, client.Send(journal.Journal{}), "context deadline exceeded")
}

func TestProxy(t *testing.T) {
	var target string
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target = r.URL.String()
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(proxy.Close)
	client := newTestClient("http://example.invalid/discord")
	client.cfg.Proxy = proxy.URL
	require.NoError(t, client.Send(journal.Journal{}))
	assert.Equal(t, client.cfg.WebhookURL, target)
}

func TestDiscordRetryAfter(t *testing.T) {
	tests := []struct {
		name   string
		header http.Header
		body   string
		want   time.Duration
	}{
		{
			name:   "header takes precedence",
			header: http.Header{"Retry-After": {"0.25"}},
			body:   `{"retry_after":0.5}`,
			want:   250 * time.Millisecond,
		},
		{
			name:   "body fallback",
			header: http.Header{"Retry-After": {"invalid"}},
			body:   `{"retry_after":0.5}`,
			want:   500 * time.Millisecond,
		},
		{
			name:   "reset header fallback",
			header: http.Header{"X-Ratelimit-Reset-After": {"0.75"}},
			want:   750 * time.Millisecond,
		},
		{
			name: "default delay",
			body: "invalid JSON",
			want: 5 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, discordRetryAfter(&http.Response{Header: tt.header}, []byte(tt.body)))
		})
	}
	t.Run("HTTP date", func(t *testing.T) {
		retryAt := time.Now().Add(time.Minute).UTC().Truncate(time.Second)
		resp := &http.Response{Header: http.Header{"Retry-After": {retryAt.Format(http.TimeFormat)}}}
		delay := discordRetryAfter(resp, nil)
		assert.InDelta(t, time.Until(retryAt).Seconds(), delay.Seconds(), 1)
	})
}

func newTestClient(webhookURL string) *Client {
	cfg := (&config.NotifDiscord{}).GetDefaults()
	cfg.WebhookURL = webhookURL
	return &Client{
		cfg: cfg,
		meta: config.Meta{
			Name:      "FTPGrab",
			URL:       "https://example.com/ftpgrab",
			Logo:      "https://example.com/logo.png",
			Author:    "CrazyMax",
			Version:   "7.0.0",
			Hostname:  "destination",
			UserAgent: "ftpgrab-test",
		},
	}
}
