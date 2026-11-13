package slack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSend(t *testing.T) {
	var payload slack.WebhookMessage
	var headers http.Header
	var method string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		headers = r.Header.Clone()
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	cfg := (&config.NotifSlack{}).GetDefaults()
	cfg.WebhookURL = srv.URL
	client := &Client{
		cfg: cfg,
		meta: config.Meta{
			Name:      "FTPGrab",
			Hostname:  "destination",
			UserAgent: "ftpgrab/test",
		},
	}
	jnl := journal.Journal{
		ServerHost: "ftp.example.com",
		Duration:   time.Minute,
	}
	jnl.Count.Success = 2
	jnl.Count.Skip = 1
	jnl.Count.Error = 3
	require.NoError(t, client.Send(jnl))
	assert.Equal(t, http.MethodPost, method)
	assert.Equal(t, "application/json", headers.Get("Content-Type"))
	require.Len(t, payload.Attachments, 1)
	attachment := payload.Attachments[0]
	assert.Equal(t, "#b60205", attachment.Color)
	assert.Equal(t, "FTPGrab", attachment.AuthorName)
	assert.Contains(t, attachment.Text, "<!channel>")
	assert.Contains(t, attachment.Text, "*2* files in *1 minute*")
	assert.Contains(t, attachment.Text, "*1* have been skipped and *3* errors")
	require.Len(t, attachment.Fields, 2)
	assert.Equal(t, "ftp.example.com", attachment.Fields[0].Value)
	assert.Equal(t, "destination", attachment.Fields[1].Value)
}

func TestWebhookURLFile(t *testing.T) {
	tests := []struct {
		name      string
		inline    bool
		missing   bool
		wantPath  string
		wantError bool
	}{
		{
			name:     "file URL",
			wantPath: "/file",
		},
		{
			name:     "inline takes precedence",
			inline:   true,
			wantPath: "/inline",
		},
		{
			name:      "unreadable file",
			missing:   true,
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target string
			var requests atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				target = r.URL.Path
				requests.Add(1)
				w.WriteHeader(http.StatusOK)
			}))
			t.Cleanup(srv.Close)
			filename := filepath.Join(t.TempDir(), "webhook_url")
			if !tt.missing {
				require.NoError(t, os.WriteFile(filename, []byte(srv.URL+"/file"), 0o600))
			}
			cfg := (&config.NotifSlack{}).GetDefaults()
			cfg.WebhookURLFile = filename
			if tt.inline {
				cfg.WebhookURL = srv.URL + "/inline"
			}
			err := (&Client{cfg: cfg}).Send(journal.Journal{})
			if tt.wantError {
				require.ErrorContains(t, err, "cannot retrieve webhook URL for Slack notifier")
				assert.EqualValues(t, 0, requests.Load())
				return
			}
			require.NoError(t, err)
			assert.EqualValues(t, 1, requests.Load())
			assert.Equal(t, tt.wantPath, target)
		})
	}
}

func TestDeliveryResponses(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		retryAfter string
		succeed    bool
		proxy      bool
		attempts   int32
		wantError  string
	}{
		{
			name:     "success",
			status:   http.StatusOK,
			attempts: 1,
		},
		{
			name:     "explicit proxy",
			status:   http.StatusOK,
			proxy:    true,
			attempts: 1,
		},
		{
			name:       "retry rate limit",
			status:     http.StatusTooManyRequests,
			retryAfter: "0",
			succeed:    true,
			attempts:   2,
		},
		{
			name:       "stop after three attempts",
			status:     http.StatusTooManyRequests,
			retryAfter: "0",
			attempts:   3,
			wantError:  "slack rate limit exceeded",
		},
		{
			name:      "do not retry bad request",
			status:    http.StatusBadRequest,
			attempts:  1,
			wantError: "400 Bad Request",
		},
		{
			name:      "do not retry server error",
			status:    http.StatusInternalServerError,
			attempts:  1,
			wantError: "500 Internal Server Error",
		},
		{
			name:      "missing retry header",
			status:    http.StatusTooManyRequests,
			attempts:  1,
			wantError: "invalid syntax",
		},
		{
			name:       "invalid retry header",
			status:     http.StatusTooManyRequests,
			retryAfter: "invalid",
			attempts:   1,
			wantError:  "invalid syntax",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attempts atomic.Int32
			var payloads [][]byte
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var payload json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Errorf("decode payload: %v", err)
				}
				payloads = append(payloads, payload)
				if attempts.Add(1) > 1 && tt.succeed {
					w.WriteHeader(http.StatusOK)
					return
				}
				w.Header().Set("Retry-After", tt.retryAfter)
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte("response details"))
			}))
			t.Cleanup(srv.Close)
			cfg := (&config.NotifSlack{}).GetDefaults()
			cfg.WebhookURL = srv.URL
			if tt.proxy {
				cfg.Proxy = srv.URL
				cfg.WebhookURL = "http://example.invalid/slack"
			}
			err := (&Client{cfg: cfg}).Send(journal.Journal{})
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
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

func TestRetryAfter(t *testing.T) {
	var firstAttempt time.Time
	var delay time.Duration
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			firstAttempt = time.Now()
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		delay = time.Since(firstAttempt)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	cfg := (&config.NotifSlack{}).GetDefaults()
	cfg.WebhookURL = srv.URL
	require.NoError(t, (&Client{cfg: cfg}).Send(journal.Journal{}))
	assert.EqualValues(t, 2, attempts.Load())
	assert.GreaterOrEqual(t, delay, time.Second)
}

func TestTimeout(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	cfg := (&config.NotifSlack{}).GetDefaults()
	cfg.WebhookURL = srv.URL
	cfg.Timeout = new(100 * time.Millisecond)
	start := time.Now()
	err := (&Client{cfg: cfg}).Send(journal.Journal{})
	require.ErrorContains(t, err, "context deadline exceeded")
	assert.Less(t, time.Since(start), time.Second)
	assert.EqualValues(t, 1, attempts.Load())
}
