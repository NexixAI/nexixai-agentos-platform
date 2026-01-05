package federation

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/secrets"
)

type Forwarder struct {
	Client      *http.Client
	MaxAttempts int
	BaseBackoff time.Duration
}

func NewForwarder() *Forwarder {
	maxAttempts := 3
	if v := strings.TrimSpace(os.Getenv("AGENTOS_FED_FORWARD_MAX_ATTEMPTS")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			maxAttempts = parsed
		}
	}
	baseBackoff := 250 * time.Millisecond
	if v := strings.TrimSpace(os.Getenv("AGENTOS_FED_FORWARD_BASE_BACKOFF_MS")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			baseBackoff = time.Duration(parsed) * time.Millisecond
		}
	}

	// Configure mTLS if certs are provided
	transport := &http.Transport{}
	tlsConfig := configureMTLSClient()
	if tlsConfig != nil {
		transport.TLSClientConfig = tlsConfig
	}

	return &Forwarder{
		Client:      &http.Client{Timeout: 10 * time.Second, Transport: transport},
		MaxAttempts: maxAttempts,
		BaseBackoff: baseBackoff,
	}
}

// configureMTLSClient loads client cert/key and CA cert for mTLS.
// Returns nil if certs not configured (falls back to standard HTTP/HTTPS).
func configureMTLSClient() *tls.Config {
	loader := secrets.NewLoader()

	// Load client cert and key
	clientCert, _ := loader.Load("AGENTOS_FED_CLIENT_CERT")
	clientKey, _ := loader.Load("AGENTOS_FED_CLIENT_KEY")
	caCert, _ := loader.Load("AGENTOS_FED_CA_CERT")

	// If no certs configured, return nil (dev mode)
	if clientCert == "" && clientKey == "" && caCert == "" {
		return nil
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Load client certificate if provided
	if clientCert != "" && clientKey != "" {
		cert, err := tls.X509KeyPair([]byte(clientCert), []byte(clientKey))
		if err != nil {
			// Log error but don't fail - fallback to no client cert
			fmt.Fprintf(os.Stderr, "WARN: failed to load federation client cert: %v\n", err)
		} else {
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}

	// Load CA cert if provided
	if caCert != "" {
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM([]byte(caCert)) {
			fmt.Fprintf(os.Stderr, "WARN: failed to parse federation CA cert\n")
		} else {
			tlsConfig.RootCAs = caCertPool
		}
	}

	return tlsConfig
}

// ForwardRun calls the remote Agent Orchestrator Run Create endpoint and returns (remote_run_id, remote_events_url, status).
func (f *Forwarder) ForwardRun(remoteAgentOrchestratorBaseURL string, agentID string, tenantID string, principalID string, bearerToken string, runCreateReq map[string]any) (string, string, string, error) {
	maxAttempts := f.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	backoff := f.BaseBackoff
	if backoff <= 0 {
		backoff = 250 * time.Millisecond
	}

	url := strings.TrimRight(remoteAgentOrchestratorBaseURL, "/") + "/v1/agents/" + agentID + "/runs"

	body, _ := json.Marshal(runCreateReq)

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-Id", tenantID)
		if principalID != "" {
			req.Header.Set("X-Principal-Id", principalID)
		}
		if bearerToken != "" {
			req.Header.Set("Authorization", "Bearer "+bearerToken)
		}

		resp, err := f.Client.Do(req)
		if err != nil {
			lastErr = err
		} else {
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				var decoded map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
					_ = resp.Body.Close()
					return "", "", "", err
				}
				_ = resp.Body.Close()

				runObj, _ := decoded["run"].(map[string]any)
				runID, _ := runObj["run_id"].(string)
				eventsURL, _ := runObj["events_url"].(string)
				status, _ := runObj["status"].(string)

				if runID == "" || eventsURL == "" {
					return "", "", "", fmt.Errorf("remote response missing run_id/events_url")
				}

				// events_url is expected to be relative; normalize to absolute.
				absEvents := eventsURL
				if strings.HasPrefix(eventsURL, "/") {
					absEvents = strings.TrimRight(remoteAgentOrchestratorBaseURL, "/") + eventsURL
				} else if strings.HasPrefix(eventsURL, "http://") || strings.HasPrefix(eventsURL, "https://") {
					absEvents = eventsURL
				} else {
					absEvents = strings.TrimRight(remoteAgentOrchestratorBaseURL, "/") + "/" + eventsURL
				}

				if status == "" {
					status = "queued"
				}

				return runID, absEvents, status, nil
			}

			lastErr = fmt.Errorf("remote agent-orchestrator returned %s", resp.Status)
			_ = resp.Body.Close()

			// Do not retry non-retryable client errors.
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				return "", "", "", lastErr
			}
		}

		// Retry transient failures with simple linear backoff.
		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * backoff)
		}
	}

	return "", "", "", lastErr
}
