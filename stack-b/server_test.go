package stackb

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/types"
)

func TestInvokePolicyDeniedWhenOptionDeny(t *testing.T) {
	s := New("test")
	rec := httptest.NewRecorder()
	reqBody := types.ModelInvokeRequest{
		Operation: "chat",
		ModelID:   "local-stub-llm",
		Input:     map[string]any{"text": "hello"},
		Options:   map[string]any{"deny": true},
	}
	payload, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/models:invoke", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Id", "tnt_test")
	req.Header.Set("X-Principal-Id", "usr_test")

	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestInvokeReturnsUsage(t *testing.T) {
	s := New("test")
	rec := httptest.NewRecorder()
	reqBody := types.ModelInvokeRequest{
		Operation: "chat",
		ModelID:   "local-stub-llm",
		Input:     map[string]any{"text": "hello world"},
		Options:   map[string]any{},
	}
	payload, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/models:invoke", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Id", "tnt_test")
	req.Header.Set("X-Principal-Id", "usr_test")

	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp types.ModelInvokeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Usage == nil || resp.Usage["total_tokens"] == nil {
		t.Fatalf("expected usage in response, got %+v", resp)
	}
}
