package apprise

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
		token    string
		urls     []string
		wantPath string
	}{
		{ //nolint:gosec // The Apprise config key is a test fixture, not a credential.
			name:     "saved configuration",
			token:    "ftpgrab-key",
			wantPath: "/api/notify/ftpgrab-key",
		},
		{
			name: "direct URLs",
			urls: []string{
				"mailto://ops@example.com",
				"ntfy://ntfy.sh/reports",
			},
			wantPath: "/api/notify",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var method, target string
			var headers http.Header
			var payload struct {
				Body  string   `json:"body"`
				Title string   `json:"title"`
				Tags  []string `json:"tags"`
				URLs  []string `json:"urls"`
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
			client := newTestClient(srv.URL + "/api/?source=ftpgrab")
			client.cfg.Token = tt.token
			client.cfg.URLs = tt.urls
			client.cfg.Tags = []string{"reports", "ops"}
			jnl := journal.Journal{ServerHost: "ftp.example.com", Duration: time.Minute}
			jnl.Count.Success = 2
			jnl.Count.Skip = 1
			jnl.Count.Error = 3
			require.NoError(t, client.Send(jnl))
			assert.Equal(t, http.MethodPost, method)
			assert.Equal(t, tt.wantPath+"?source=ftpgrab", target)
			assert.Equal(t, "application/json", headers.Get("Content-Type"))
			assert.Equal(t, "ftpgrab-test", headers.Get("User-Agent"))
			assert.Equal(t, "FTPGrab report for ftp.example.com on destination", payload.Title)
			assert.Equal(t, "FTPGrab has successfully downloaded 2 files in 1 minute.\n1 have been skipped and 3 errors occurred.", payload.Body)
			assert.Equal(t, []string{"reports", "ops"}, payload.Tags)
			assert.Equal(t, tt.urls, payload.URLs)
		})
	}
}

func TestTokenFile(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(filename, []byte("file-key"), 0o600))
	tests := []struct {
		name      string
		token     string
		tokenFile string
		wantPath  string
		wantError bool
	}{
		{
			name:      "file key",
			tokenFile: filename,
			wantPath:  "/notify/file-key",
		},
		{
			name:      "inline takes precedence",
			token:     "inline-key",
			tokenFile: filename + ".missing",
			wantPath:  "/notify/inline-key",
		},
		{
			name:      "unreadable file",
			tokenFile: filename + ".missing",
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
			client := newTestClient(srv.URL)
			client.cfg.Token = tt.token
			client.cfg.TokenFile = tt.tokenFile
			err := client.Send(journal.Journal{})
			if tt.wantError {
				require.ErrorContains(t, err, "cannot retrieve token secret for Apprise notifier")
				assert.EqualValues(t, 0, requests.Load())
				return
			}
			require.NoError(t, err)
			assert.EqualValues(t, 1, requests.Load())
			assert.Equal(t, tt.wantPath, target)
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
			name:   "invalid configuration",
			status: http.StatusBadRequest,
			body:   `{"error":"invalid notification URL"}`,
		},
		{
			name:   "destination failure",
			status: http.StatusFailedDependency,
			body:   `{"error":"notification failed"}`,
		},
		{
			name:   "non-JSON proxy error",
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
	client := newTestClient("http://example.invalid/api")
	client.cfg.Token = "ftpgrab-key"
	client.cfg.Proxy = proxy.URL
	require.NoError(t, client.Send(journal.Journal{}))
	assert.Equal(t, "http://example.invalid/api/notify/ftpgrab-key", target)
}

func newTestClient(endpoint string) *Client {
	cfg := (&config.NotifApprise{}).GetDefaults()
	cfg.Endpoint = endpoint
	return &Client{
		cfg: cfg,
		meta: config.Meta{
			Name:      "FTPGrab",
			Hostname:  "destination",
			UserAgent: "ftpgrab-test",
		},
	}
}
