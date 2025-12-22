package federation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestForwardRunRetriesAndSucceeds(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"run": map[string]any{
				"run_id":     "run123",
				"events_url": "/v1/runs/run123/events",
				"status":     "running",
			},
		})
	}))
	defer srv.Close()

	f := NewForwarder()
	f.BaseBackoff = 1 * time.Millisecond
	f.MaxAttempts = 3

	runID, eventsURL, status, err := f.ForwardRun(srv.URL, "agt_demo", "t1", "p1", "", map[string]any{"input": map[string]any{"text": "hi"}})
	if err != nil {
		t.Fatalf("forward run failed: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
	if runID != "run123" {
		t.Fatalf("unexpected run_id %s", runID)
	}
	if status != "running" {
		t.Fatalf("unexpected status %s", status)
	}
	expectedEventsURL := srv.URL + "/v1/runs/run123/events"
	if eventsURL != expectedEventsURL {
		t.Fatalf("expected events_url %s, got %s", expectedEventsURL, eventsURL)
	}
}

func TestForwardRunStopsOnClientError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	f := NewForwarder()
	f.BaseBackoff = 1 * time.Millisecond
	f.MaxAttempts = 3

	_, _, _, err := f.ForwardRun(srv.URL, "agt_demo", "t1", "p1", "", map[string]any{"input": map[string]any{"text": "hi"}})
	if err == nil {
		t.Fatalf("expected error for 400 response")
	}
}
