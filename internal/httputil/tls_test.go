package httputil

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadTLSConfig(t *testing.T) {
	tlsConfig, err := LoadTLSConfig(false, nil)
	require.NoError(t, err)
	assert.False(t, tlsConfig.InsecureSkipVerify)
	assert.Nil(t, tlsConfig.RootCAs)
	tlsConfig, err = LoadTLSConfig(true, nil)
	require.NoError(t, err)
	assert.True(t, tlsConfig.InsecureSkipVerify)

	first, firstPEM := newTestCACert(t, 1)
	second, secondPEM := newTestCACert(t, 2)
	dir := t.TempDir()
	firstFile := filepath.Join(dir, "first.pem")
	secondFile := filepath.Join(dir, "second.pem")
	bundleFile := filepath.Join(dir, "bundle.pem")
	require.NoError(t, os.WriteFile(firstFile, firstPEM, 0o600))
	require.NoError(t, os.WriteFile(secondFile, secondPEM, 0o600))
	require.NoError(t, os.WriteFile(bundleFile, append(firstPEM, secondPEM...), 0o600))
	for _, files := range [][]string{{firstFile, secondFile}, {bundleFile}} {
		tlsConfig, err := LoadTLSConfig(false, files)
		require.NoError(t, err)
		assert.False(t, tlsConfig.InsecureSkipVerify)
		require.NotNil(t, tlsConfig.RootCAs)
		for _, cert := range []*x509.Certificate{first, second} {
			_, err := cert.Verify(x509.VerifyOptions{Roots: tlsConfig.RootCAs})
			require.NoError(t, err)
		}
	}

	t.Run("missing file", func(t *testing.T) {
		_, err := LoadTLSConfig(false, []string{filepath.Join(dir, "missing.pem")})
		require.Error(t, err)
	})
	for _, content := range []string{"", "not a PEM certificate"} {
		t.Run("invalid PEM/"+content, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "invalid.pem")
			require.NoError(t, os.WriteFile(filename, []byte(content), 0o600))
			_, err := LoadTLSConfig(false, []string{filename})
			require.ErrorContains(t, err, "no certificates found in CA file")
		})
	}
}

func TestClientTLS(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
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
		serverName string
		wantError  bool
	}{
		{
			name:      "reject untrusted server by default",
			wantError: true,
		},
		{
			name:    "trust custom CA",
			caFiles: []string{certFile},
		},
		{
			name:       "custom CA still verifies hostname",
			caFiles:    []string{certFile},
			serverName: "wrong.invalid",
			wantError:  true,
		},
		{
			name:       "skip verification explicitly",
			skipVerify: true,
			serverName: "wrong.invalid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient("", tt.skipVerify, tt.caFiles)
			require.NoError(t, err)
			t.Cleanup(client.CloseIdleConnections)
			transport := client.Transport.(*http.Transport)
			transport.TLSClientConfig.ServerName = tt.serverName
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
			require.NoError(t, err)
			resp, err := client.Do(req)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			resp.Body.Close()
			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		})
	}
}

func newTestCACert(t *testing.T, serial int64) (*x509.Certificate, []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(serial),
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	return cert, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}
