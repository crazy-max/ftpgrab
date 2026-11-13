package ntfy

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
	tests := []struct {
		name     string
		icon     string
		wantIcon string
	}{
		{
			name:     "configured icon",
			icon:     "https://example.com/custom.png",
			wantIcon: "https://example.com/custom.png",
		},
		{
			name:     "logo fallback",
			wantIcon: "https://example.com/logo.png",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var method, target string
			var headers http.Header
			var payload struct {
				Topic    string   `json:"topic"`
				Message  string   `json:"message"`
				Title    string   `json:"title"`
				Priority int      `json:"priority"`
				Tags     []string `json:"tags"`
				Icon     string   `json:"icon"`
				Click    string   `json:"click"`
				Markdown bool     `json:"markdown"`
			}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				method = r.Method
				target = r.URL.RequestURI()
				headers = r.Header.Clone()
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Errorf("decode payload: %v", err)
				}
				w.WriteHeader(http.StatusOK)
			}))
			t.Cleanup(srv.Close)
			client := newTestClient(srv.URL + "/base?source=ftpgrab")
			client.cfg.Icon = tt.icon
			client.cfg.Click = "https://example.com/reports"
			client.cfg.Priority = 5
			client.cfg.Tags = []string{"warning", "package"}
			jnl := journal.Journal{ServerHost: "ftp.example.com", Duration: time.Minute}
			jnl.Count.Success = 2
			jnl.Count.Skip = 1
			jnl.Count.Error = 3
			require.NoError(t, client.Send(jnl))
			assert.Equal(t, http.MethodPost, method)
			assert.Equal(t, "/base?source=ftpgrab", target)
			assert.Equal(t, "application/json", headers.Get("Content-Type"))
			assert.Equal(t, "ftpgrab-test", headers.Get("User-Agent"))
			assert.Equal(t, "reports", payload.Topic)
			assert.Equal(t, "FTPGrab report for ftp.example.com on destination", payload.Title)
			assert.Equal(t, "FTPGrab has successfully downloaded **2** files in **1 minute**.\n**1** have been skipped and **3** errors occurred.", payload.Message)
			assert.Equal(t, 5, payload.Priority)
			assert.Equal(t, []string{"warning", "package"}, payload.Tags)
			assert.Equal(t, tt.wantIcon, payload.Icon)
			assert.Equal(t, client.cfg.Click, payload.Click)
			assert.True(t, payload.Markdown)
		})
	}
}

func TestToken(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(filename, []byte("file-token"), 0o600))
	tests := []struct {
		name      string
		token     string
		tokenFile string
		wantAuth  string
		wantError bool
	}{
		{
			name: "anonymous",
		},
		{
			name:     "inline token",
			token:    "inline-token",
			wantAuth: "Bearer inline-token",
		},
		{
			name:      "file token",
			tokenFile: filename,
			wantAuth:  "Bearer file-token",
		},
		{
			name:      "inline takes precedence",
			token:     "inline-token",
			tokenFile: filename + ".missing",
			wantAuth:  "Bearer inline-token",
		},
		{
			name:      "unreadable file",
			tokenFile: filename + ".missing",
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var auth string
			var requests atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				auth = r.Header.Get("Authorization")
				requests.Add(1)
				w.WriteHeader(http.StatusOK)
			}))
			t.Cleanup(srv.Close)
			client := newTestClient(srv.URL)
			client.cfg.Token = tt.token
			client.cfg.TokenFile = tt.tokenFile
			err := client.Send(journal.Journal{})
			if tt.wantError {
				require.ErrorContains(t, err, "cannot retrieve token secret for ntfy notifier")
				assert.EqualValues(t, 0, requests.Load())
				return
			}
			require.NoError(t, err)
			assert.EqualValues(t, 1, requests.Load())
			assert.Equal(t, tt.wantAuth, auth)
		})
	}
}

func TestErrorResponses(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{
			name:   "ntfy JSON error",
			status: http.StatusBadRequest,
			body:   `{"code":40001,"http":400,"error":"invalid topic"}`,
		},
		{
			name:   "authentication error",
			status: http.StatusUnauthorized,
			body:   "unauthorized",
		},
		{
			name:   "rate limit",
			status: http.StatusTooManyRequests,
			body:   "rate limited",
		},
		{
			name:   "proxy error",
			status: http.StatusBadGateway,
			body:   "<html>bad gateway</html>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			t.Cleanup(srv.Close)
			err := newTestClient(srv.URL).Send(journal.Journal{})
			require.EqualError(t, err, fmt.Sprintf("unexpected HTTP status %d: %s", tt.status, tt.body))
		})
	}
}

func TestTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
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
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(proxy.Close)
	client := newTestClient("http://example.invalid/ntfy")
	client.cfg.Proxy = proxy.URL
	require.NoError(t, client.Send(journal.Journal{}))
	assert.Equal(t, client.cfg.Endpoint, target)
}

func newTestClient(endpoint string) *Client {
	cfg := (&config.NotifNtfy{}).GetDefaults()
	cfg.Endpoint = endpoint
	cfg.Topic = "reports"
	return &Client{
		cfg: cfg,
		meta: config.Meta{
			Name:      "FTPGrab",
			Hostname:  "destination",
			Logo:      "https://example.com/logo.png",
			UserAgent: "ftpgrab-test",
		},
	}
}
