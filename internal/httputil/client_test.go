package httputil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient("", false, nil)
	require.NoError(t, err)
	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	require.NotNil(t, transport.Proxy)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)
	gotProxy, err := transport.Proxy(req)
	require.NoError(t, err)
	wantProxy, err := http.ProxyFromEnvironment(req)
	require.NoError(t, err)
	assert.Equal(t, wantProxy, gotProxy)

	var target string
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target = r.URL.String()
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(proxy.Close)
	client, err = NewClient(proxy.URL, false, nil)
	require.NoError(t, err)
	t.Cleanup(client.CloseIdleConnections)
	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.invalid/report", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, "http://example.invalid/report", target)
	assert.NotSame(t, http.DefaultTransport, client.Transport)
}

func TestInvalidProxy(t *testing.T) {
	for _, proxy := range []string{":", "proxy.example.com:3128", "http:///", "http://%"} {
		t.Run(proxy, func(t *testing.T) {
			_, err := NewClient(proxy, false, nil)
			require.Error(t, err)
		})
	}
}
