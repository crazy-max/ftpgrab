package httputil

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/pkg/errors"
)

// LoadTLSConfig creates a TLS configuration with optional custom root CAs.
func LoadTLSConfig(insecureSkipVerify bool, caCertFiles []string) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
	}
	if len(caCertFiles) > 0 {
		certPool := x509.NewCertPool()
		for _, caCertFile := range caCertFiles {
			caCert, err := os.ReadFile(caCertFile)
			if err != nil {
				return nil, err
			}
			if !certPool.AppendCertsFromPEM(caCert) {
				return nil, errors.Errorf("no certificates found in CA file %q", caCertFile)
			}
		}
		tlsConfig.RootCAs = certPool
	}
	return tlsConfig, nil
}
