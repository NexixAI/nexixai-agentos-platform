package agentorchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/types"
)

func TestCancelQueuedRunSucceeds(t *testing.T) {
	srv, err := New("test")
	if err != nil {
		t.Fatalf("New server error: %v", err)
	}

	// Create a tenant first (requires tenants:admin scope)
	createTenantReq := httptest.NewRequest(http.MethodPost, "/v1/admin/tenants", bytes.NewBufferString(`{"tenant_id":"tnt_test"}`))
	createTenantReq.Header.Set("Authorization", "Bearer test-token")
	createTenantReq.Header.Set("X-Tenant-Id", "tnt_test")
	createTenantReq.Header.Set("X-Scopes", "tenants:admin")
	createTenantRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(createTenantRec, createTenantReq)

	// Create a run
	createRunReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agt_test/runs", bytes.NewBufferString(`{"input":{"type":"text","text":"hello"}}`))
	createRunReq.Header.Set("Authorization", "Bearer test-token")
	createRunReq.Header.Set("X-Tenant-Id", "tnt_test")
	createRunRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(createRunRec, createRunReq)

	if createRunRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for run creation, got %d: %s", createRunRec.Code, createRunRec.Body.String())
	}

	var createResp types.RunCreateResponse
	if err := json.Unmarshal(createRunRec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}
	runID := createResp.Run.RunID

	if createResp.Run.Status != "queued" {
		t.Fatalf("expected status queued, got %s", createResp.Run.Status)
	}

	// Cancel the run
	cancelReq := httptest.NewRequest(http.MethodPost, "/v1/runs/"+runID+":cancel", nil)
	cancelReq.Header.Set("Authorization", "Bearer test-token")
	cancelReq.Header.Set("X-Tenant-Id", "tnt_test")
	cancelRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(cancelRec, cancelReq)

	if cancelRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for cancel, got %d: %s", cancelRec.Code, cancelRec.Body.String())
	}

	var cancelResp types.RunCancelResponse
	if err := json.Unmarshal(cancelRec.Body.Bytes(), &cancelResp); err != nil {
		t.Fatalf("unmarshal cancel response: %v", err)
	}

	if cancelResp.Run.Status != "canceled" {
		t.Fatalf("expected status canceled, got %s", cancelResp.Run.Status)
	}

	if cancelResp.Run.CompletedAt == "" {
		t.Fatalf("expected completed_at to be set")
	}
}

func TestCancelCompletedRunReturns409(t *testing.T) {
	srv, err := New("test")
	if err != nil {
		t.Fatalf("New server error: %v", err)
	}

	// Create tenant (requires tenants:admin scope)
	createTenantReq := httptest.NewRequest(http.MethodPost, "/v1/admin/tenants", bytes.NewBufferString(`{"tenant_id":"tnt_test2"}`))
	createTenantReq.Header.Set("Authorization", "Bearer test-token")
	createTenantReq.Header.Set("X-Tenant-Id", "tnt_test2")
	createTenantReq.Header.Set("X-Scopes", "tenants:admin")
	createTenantRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(createTenantRec, createTenantReq)

	// Create a run
	createRunReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agt_test/runs", bytes.NewBufferString(`{"input":{"type":"text","text":"hello"}}`))
	createRunReq.Header.Set("Authorization", "Bearer test-token")
	createRunReq.Header.Set("X-Tenant-Id", "tnt_test2")
	createRunRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(createRunRec, createRunReq)

	var createResp types.RunCreateResponse
	if err := json.Unmarshal(createRunRec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}
	runID := createResp.Run.RunID

	// Manually complete the run by getting it (triggers auto-complete after 5s)
	// Instead, we'll directly update it via storage
	run := createResp.Run
	run.Status = "completed"
	run.CompletedAt = time.Now().UTC().Format(time.RFC3339)
	if err := srv.runs.Save(context.Background(), run); err != nil {
		t.Fatalf("save completed run: %v", err)
	}

	// Try to cancel the completed run
	cancelReq := httptest.NewRequest(http.MethodPost, "/v1/runs/"+runID+":cancel", nil)
	cancelReq.Header.Set("Authorization", "Bearer test-token")
	cancelReq.Header.Set("X-Tenant-Id", "tnt_test2")
	cancelRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(cancelRec, cancelReq)

	if cancelRec.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict for cancel of completed run, got %d: %s", cancelRec.Code, cancelRec.Body.String())
	}
}

func TestCancelRunFromDifferentTenantReturns404(t *testing.T) {
	srv, err := New("test")
	if err != nil {
		t.Fatalf("New server error: %v", err)
	}

	// Create tenant A (requires tenants:admin scope)
	createTenantAReq := httptest.NewRequest(http.MethodPost, "/v1/admin/tenants", bytes.NewBufferString(`{"tenant_id":"tnt_alpha"}`))
	createTenantAReq.Header.Set("Authorization", "Bearer test-token")
	createTenantAReq.Header.Set("X-Tenant-Id", "tnt_alpha")
	createTenantAReq.Header.Set("X-Scopes", "tenants:admin")
	createTenantARec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(createTenantARec, createTenantAReq)

	// Create tenant B (requires tenants:admin scope)
	createTenantBReq := httptest.NewRequest(http.MethodPost, "/v1/admin/tenants", bytes.NewBufferString(`{"tenant_id":"tnt_beta"}`))
	createTenantBReq.Header.Set("Authorization", "Bearer test-token")
	createTenantBReq.Header.Set("X-Tenant-Id", "tnt_beta")
	createTenantBReq.Header.Set("X-Scopes", "tenants:admin")
	createTenantBRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(createTenantBRec, createTenantBReq)

	// Create a run in tenant A
	createRunReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agt_test/runs", bytes.NewBufferString(`{"input":{"type":"text","text":"hello"}}`))
	createRunReq.Header.Set("Authorization", "Bearer test-token")
	createRunReq.Header.Set("X-Tenant-Id", "tnt_alpha")
	createRunRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(createRunRec, createRunReq)

	var createResp types.RunCreateResponse
	if err := json.Unmarshal(createRunRec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}
	runID := createResp.Run.RunID

	// Try to cancel from tenant B - should get 404
	cancelReq := httptest.NewRequest(http.MethodPost, "/v1/runs/"+runID+":cancel", nil)
	cancelReq.Header.Set("Authorization", "Bearer test-token")
	cancelReq.Header.Set("X-Tenant-Id", "tnt_beta")
	cancelRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(cancelRec, cancelReq)

	if cancelRec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for cancel from different tenant, got %d: %s", cancelRec.Code, cancelRec.Body.String())
	}
}

func TestCancelAlreadyCanceledRunReturns409(t *testing.T) {
	srv, err := New("test")
	if err != nil {
		t.Fatalf("New server error: %v", err)
	}

	// Create tenant (requires tenants:admin scope)
	createTenantReq := httptest.NewRequest(http.MethodPost, "/v1/admin/tenants", bytes.NewBufferString(`{"tenant_id":"tnt_test3"}`))
	createTenantReq.Header.Set("Authorization", "Bearer test-token")
	createTenantReq.Header.Set("X-Tenant-Id", "tnt_test3")
	createTenantReq.Header.Set("X-Scopes", "tenants:admin")
	createTenantRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(createTenantRec, createTenantReq)

	// Create a run
	createRunReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agt_test/runs", bytes.NewBufferString(`{"input":{"type":"text","text":"hello"}}`))
	createRunReq.Header.Set("Authorization", "Bearer test-token")
	createRunReq.Header.Set("X-Tenant-Id", "tnt_test3")
	createRunRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(createRunRec, createRunReq)

	var createResp types.RunCreateResponse
	if err := json.Unmarshal(createRunRec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}
	runID := createResp.Run.RunID

	// Cancel the run first time
	cancelReq1 := httptest.NewRequest(http.MethodPost, "/v1/runs/"+runID+":cancel", nil)
	cancelReq1.Header.Set("Authorization", "Bearer test-token")
	cancelReq1.Header.Set("X-Tenant-Id", "tnt_test3")
	cancelRec1 := httptest.NewRecorder()
	srv.Handler().ServeHTTP(cancelRec1, cancelReq1)

	if cancelRec1.Code != http.StatusOK {
		t.Fatalf("expected 200 for first cancel, got %d", cancelRec1.Code)
	}

	// Try to cancel again - should get 409
	cancelReq2 := httptest.NewRequest(http.MethodPost, "/v1/runs/"+runID+":cancel", nil)
	cancelReq2.Header.Set("Authorization", "Bearer test-token")
	cancelReq2.Header.Set("X-Tenant-Id", "tnt_test3")
	cancelRec2 := httptest.NewRecorder()
	srv.Handler().ServeHTTP(cancelRec2, cancelReq2)

	if cancelRec2.Code != http.StatusConflict {
		t.Fatalf("expected 409 for second cancel, got %d: %s", cancelRec2.Code, cancelRec2.Body.String())
	}
}
