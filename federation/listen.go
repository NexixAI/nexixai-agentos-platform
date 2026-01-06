package federation

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"strings"
)

func ListenAndServe(addr, version string) error {
	s := New(version)
	handler := s.Handler()

	// Check if mTLS is required
	tlsConfig := loadMTLSServerConfig()
	if tlsConfig != nil {
		server := &http.Server{
			Addr:      addr,
			Handler:   handler,
			TLSConfig: tlsConfig,
		}

		serverCertPath := strings.TrimSpace(os.Getenv("AGENTOS_FED_SERVER_CERT"))
		serverKeyPath := strings.TrimSpace(os.Getenv("AGENTOS_FED_SERVER_KEY"))
		return server.ListenAndServeTLS(serverCertPath, serverKeyPath)
	}

	return http.ListenAndServe(addr, handler)
}

// loadMTLSServerConfig loads mTLS server configuration from environment.
// Returns nil if mTLS not required (dev mode).
func loadMTLSServerConfig() *tls.Config {
	requireMTLS := strings.TrimSpace(os.Getenv("AGENTOS_FED_REQUIRE_MTLS"))
	if !strings.EqualFold(requireMTLS, "true") {
		return nil
	}

	serverCertPath := strings.TrimSpace(os.Getenv("AGENTOS_FED_SERVER_CERT"))
	serverKeyPath := strings.TrimSpace(os.Getenv("AGENTOS_FED_SERVER_KEY"))
	caCertPath := strings.TrimSpace(os.Getenv("AGENTOS_FED_CA_CERT"))

	if serverCertPath == "" || serverKeyPath == "" {
		return nil
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Require and verify client certs if CA is provided
	if caCertPath != "" {
		caCert, err := os.ReadFile(caCertPath)
		if err == nil {
			caCertPool := x509.NewCertPool()
			if caCertPool.AppendCertsFromPEM(caCert) {
				tlsConfig.ClientCAs = caCertPool
				tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
			}
		}
	}

	return tlsConfig
}

// GetPeerIdentity extracts peer identity from a verified client certificate.
// Returns the Common Name (CN) or first DNS SAN from the client cert.
func GetPeerIdentity(r *http.Request) string {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return ""
	}

	cert := r.TLS.PeerCertificates[0]

	// Prefer DNS SAN if available
	if len(cert.DNSNames) > 0 {
		return cert.DNSNames[0]
	}

	// Fall back to Common Name
	return cert.Subject.CommonName
}
