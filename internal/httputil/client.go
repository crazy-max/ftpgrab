package httputil

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// NewClient creates an HTTP client with optional proxy and TLS settings.
// An empty proxy preserves the default transport's environment proxy settings.
func NewClient(proxy string, insecureSkipVerify bool, caCertFiles []string) (http.Client, error) {
	tlsConfig, err := LoadTLSConfig(insecureSkipVerify, caCertFiles)
	if err != nil {
		return http.Client{}, err
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig
	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return http.Client{}, err
		}
		if proxyURL.Scheme == "" || proxyURL.Host == "" {
			return http.Client{}, errors.New("proxy URL must include scheme and host")
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	return http.Client{Transport: transport}, nil
}
