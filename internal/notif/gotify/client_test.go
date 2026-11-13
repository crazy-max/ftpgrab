package gotify

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSend(t *testing.T) {
	var method, target string
	var headers http.Header
	var contentLength int64
	var payload struct {
		Message  string                       `json:"message"`
		Title    string                       `json:"title"`
		Priority int                          `json:"priority"`
		Extras   map[string]map[string]string `json:"extras"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		target = r.URL.RequestURI()
		headers = r.Header.Clone()
		contentLength = r.ContentLength
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	client := newTestClient(srv.URL + "/base/?source=ftpgrab")
	client.cfg.Priority = 5
	jnl := journal.Journal{ServerHost: "ftp.example.com", Duration: time.Minute}
	jnl.Count.Success = 2
	jnl.Count.Skip = 1
	jnl.Count.Error = 3
	require.NoError(t, client.Send(jnl))
	assert.Equal(t, http.MethodPost, method)
	assert.Equal(t, "/base/message?source=ftpgrab", target)
	assert.Equal(t, "application/json", headers.Get("Content-Type"))
	assert.Equal(t, "ftpgrab-test", headers.Get("User-Agent"))
	assert.Equal(t, "gotify-token", headers.Get("X-Gotify-Key"))
	assert.Positive(t, contentLength)
	assert.Equal(t, "FTPGrab report for ftp.example.com on destination", payload.Title)
	assert.Equal(t, "FTPGrab has successfully downloaded **2** files in **1 minute**.\n**1** have been skipped and **3** errors occurred.", payload.Message)
	assert.Equal(t, 5, payload.Priority)
	assert.Equal(t, map[string]map[string]string{
		"client::display": {
			"contentType": "text/markdown",
		},
	}, payload.Extras)
}

func TestTokenFile(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(filename, []byte("file-token"), 0o600))
	tests := []struct {
		name      string
		token     string
		tokenFile string
		wantToken string
		wantError bool
	}{
		{
			name:      "file token",
			tokenFile: filename,
			wantToken: "file-token",
		},
		{
			name:      "inline takes precedence",
			token:     "inline-token",
			tokenFile: filename + ".missing",
			wantToken: "inline-token",
		},
		{
			name:      "unreadable file",
			tokenFile: filename + ".missing",
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var token string
			var requests atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				token = r.Header.Get("X-Gotify-Key")
				requests.Add(1)
				w.WriteHeader(http.StatusOK)
			}))
			t.Cleanup(srv.Close)
			client := newTestClient(srv.URL)
			client.cfg.Token = tt.token
			client.cfg.TokenFile = tt.tokenFile
			err := client.Send(journal.Journal{})
			if tt.wantError {
				require.ErrorContains(t, err, "cannot retrieve token secret for Gotify notifier")
				assert.EqualValues(t, 0, requests.Load())
				return
			}
			require.NoError(t, err)
			assert.EqualValues(t, 1, requests.Load())
			assert.Equal(t, tt.wantToken, token)
		})
	}
}

func TestErrorResponses(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		body      string
		wantError string
	}{
		{
			name:      "Gotify error",
			status:    http.StatusUnauthorized,
			body:      `{"error":"Unauthorized","errorCode":401,"errorDescription":"invalid application token"}`,
			wantError: "401 Unauthorized: invalid application token",
		},
		{
			name:      "non-JSON proxy error",
			status:    http.StatusBadGateway,
			body:      "<html>bad gateway</html>",
			wantError: "cannot decode JSON error response for HTTP 502 Bad Gateway status",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			t.Cleanup(srv.Close)
			require.ErrorContains(t, newTestClient(srv.URL).Send(journal.Journal{}), tt.wantError)
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
	client := newTestClient("http://example.invalid/base")
	client.cfg.Proxy = proxy.URL
	require.NoError(t, client.Send(journal.Journal{}))
	assert.Equal(t, "http://example.invalid/base/message", target)
}

func newTestClient(endpoint string) *Client {
	cfg := (&config.NotifGotify{}).GetDefaults()
	cfg.Endpoint = endpoint
	cfg.Token = "gotify-token"
	return &Client{
		cfg: cfg,
		meta: config.Meta{
			Name:      "FTPGrab",
			Hostname:  "destination",
			UserAgent: "ftpgrab-test",
		},
	}
}
