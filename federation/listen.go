package federation

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/secrets"
)

func ListenAndServe(addr, version string) error {
	s := New(version)

	// Check if mTLS is required
	requireMTLS := strings.ToLower(strings.TrimSpace(os.Getenv("AGENTOS_FED_REQUIRE_MTLS"))) == "true"

	if !requireMTLS {
		// Standard HTTP server (dev mode)
		return http.ListenAndServe(addr, s.Handler())
	}

	// Configure mTLS server
	tlsConfig, err := configureMTLSServer()
	if err != nil {
		return fmt.Errorf("failed to configure mTLS: %w", err)
	}

	server := &http.Server{
		Addr:      addr,
		Handler:   s.Handler(),
		TLSConfig: tlsConfig,
	}

	// ListenAndServeTLS requires cert/key files or we use the TLSConfig
	// Since we load from env/secrets, we need to use a custom listener
	return server.ListenAndServeTLS("", "")
}

// configureMTLSServer configures TLS for the federation server.
// Requires client certificates if AGENTOS_FED_REQUIRE_MTLS=true.
func configureMTLSServer() (*tls.Config, error) {
	loader := secrets.NewLoader()

	// Load server cert and key
	serverCert, err := loader.Require("AGENTOS_FED_SERVER_CERT")
	if err != nil {
		return nil, fmt.Errorf("AGENTOS_FED_SERVER_CERT required for mTLS: %w", err)
	}

	serverKey, err := loader.Require("AGENTOS_FED_SERVER_KEY")
	if err != nil {
		return nil, fmt.Errorf("AGENTOS_FED_SERVER_KEY required for mTLS: %w", err)
	}

	// Load CA cert for client verification
	caCert, err := loader.Require("AGENTOS_FED_CA_CERT")
	if err != nil {
		return nil, fmt.Errorf("AGENTOS_FED_CA_CERT required for mTLS: %w", err)
	}

	// Parse server certificate
	cert, err := tls.X509KeyPair([]byte(serverCert), []byte(serverKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse server cert/key: %w", err)
	}

	// Parse CA certificate
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM([]byte(caCert)) {
		return nil, fmt.Errorf("failed to parse CA cert")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12,
	}

	return tlsConfig, nil
}
