package federation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Forwarder struct {
	Client      *http.Client
	MaxAttempts int
	BaseBackoff time.Duration
}

func NewForwarder() *Forwarder {
	return &Forwarder{
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

// ForwardRun calls the remote Stack A Run Create endpoint and returns (remote_run_id, remote_events_url, status).
func (f *Forwarder) ForwardRun(remoteStackABaseURL string, agentID string, tenantID string, principalID string, runCreateReq map[string]any) (string, string, string, error) {
	url := strings.TrimRight(remoteStackABaseURL, "/") + "/v1/agents/" + agentID + "/runs"

	body, _ := json.Marshal(runCreateReq)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Id", tenantID)
	if principalID != "" {
		req.Header.Set("X-Principal-Id", principalID)
	}

	resp, err := f.Client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", "", fmt.Errorf("remote stack-a returned %s", resp.Status)
	}

	var decoded map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", "", "", err
	}

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
		absEvents = strings.TrimRight(remoteStackABaseURL, "/") + eventsURL
	} else if strings.HasPrefix(eventsURL, "http://") || strings.HasPrefix(eventsURL, "https://") {
		absEvents = eventsURL
	} else {
		absEvents = strings.TrimRight(remoteStackABaseURL, "/") + "/" + eventsURL
	}

	if status == "" {
		status = "queued"
	}

	return runID, absEvents, status, nil
}
