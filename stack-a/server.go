package stacka

import (
	"bufio"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/audit"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/auth"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/httpx"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/id"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/metrics"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/middleware"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/quota"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/storage"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/tenants"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/types"
)

type Server struct {
	version string

	runs    storage.RunStore
	tenants *tenants.Store
	limiter *quota.Limiter
	audit   audit.Logger
}

func New(version string) (*Server, error) {
	runStore, err := storage.NewRunStoreFromEnv()
	if err != nil {
		return nil, err
	}
	tenantStore := tenants.NewStore()
	if def := auth.DefaultTenant(); def != "" {
		tenantStore.EnsureDefault(def)
	}
	return &Server{
		version: version,
		runs:    runStore,
		tenants: tenantStore,
		limiter: quota.NewFromEnv("AGENTOS_QUOTA_RUN_CREATE_QPS", "AGENTOS_QUOTA_CONCURRENT_RUNS", 10, 25),
		audit:   audit.NewFromEnv(),
	}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/health", s.handleHealth)
	mux.HandleFunc("/v1/agents/", s.handleAgents) // /v1/agents/{agent_id}/runs
	mux.HandleFunc("/v1/runs/", s.handleRuns)     // /v1/runs/{run_id} and /v1/runs/{run_id}/events
	mux.HandleFunc("/v1/admin/tenants", s.handleTenants)
	mux.HandleFunc("/v1/admin/tenants/", s.handleTenants)
	mux.Handle("/metrics", middleware.ProtectMetrics(metrics.Handler()))

	h := middleware.WithAuth(mux)
	h = middleware.EnsureRequestID(h)
	h = metrics.Instrument("stack-a", h)
	return h
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		return
	}
	httpx.JSON(w, http.StatusOK, types.HealthResponse{Status: "ok", Service: "stack-a", Version: s.version})
}

func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		return
	}

	ac, _ := auth.Get(r.Context())
	tenantID, ok := resolveTenant(w, r, ac)
	if !ok {
		return
	}
	if !s.tenantsExists(tenantID) {
		httpx.Error(w, http.StatusForbidden, "tenant_unknown", "tenant not found", httpx.CorrelationID(r), false)
		return
	}

	if !s.limiter.AllowQPS(tenantID) {
		metrics.IncQuotaDenied("stack-a", "runs_create_qps")
		httpx.Error(w, http.StatusTooManyRequests, "quota_exceeded", "run create QPS exceeded", httpx.CorrelationID(r), true)
		s.audit.Log(audit.Entry{
			TenantID: tenantID, PrincipalID: ac.PrincipalID, Action: "runs.create", Resource: "stack-a", Outcome: "denied",
			CorrelationID: httpx.CorrelationID(r), RequestID: r.Header.Get("X-Request-Id"),
			Meta: map[string]any{"reason": "qps_exceeded"},
		})
		return
	}
	if !s.limiter.TryIncConcurrent(tenantID) {
		metrics.IncQuotaDenied("stack-a", "runs_concurrency")
		httpx.Error(w, http.StatusTooManyRequests, "quota_exceeded", "concurrent runs exceeded", httpx.CorrelationID(r), true)
		s.audit.Log(audit.Entry{
			TenantID: tenantID, PrincipalID: ac.PrincipalID, Action: "runs.create", Resource: "stack-a", Outcome: "denied",
			CorrelationID: httpx.CorrelationID(r), RequestID: r.Header.Get("X-Request-Id"),
			Meta: map[string]any{"reason": "concurrent_exceeded"},
		})
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/v1/agents/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[1] != "runs" {
		s.limiter.DecConcurrent(tenantID)
		httpx.Error(w, http.StatusNotFound, "not_found", "not found", httpx.CorrelationID(r), false)
		return
	}
	agentID := parts[0]

	var req types.RunCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.limiter.DecConcurrent(tenantID)
		httpx.Error(w, http.StatusBadRequest, "invalid_json", "invalid json body", httpx.CorrelationID(r), false)
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	runID := id.New("run")

	run := types.Run{
		TenantID:   tenantID,
		AgentID:    agentID,
		RunID:      runID,
		Status:     "queued",
		CreatedAt:  now,
		EventsURL:  "/v1/runs/" + runID + "/events",
		RunOptions: req.RunOptions,
	}

	if err := s.runs.Create(r.Context(), run); err != nil {
		s.limiter.DecConcurrent(tenantID)
		code := http.StatusInternalServerError
		errCode := "run_persist_failed"
		retryable := true
		if errors.Is(err, storage.ErrRunExists) {
			code = http.StatusConflict
			errCode = "conflict"
			retryable = false
		} else if errors.Is(err, storage.ErrInvalidRun) {
			code = http.StatusBadRequest
			errCode = "invalid_request"
			retryable = false
		}
		httpx.Error(w, code, errCode, "failed to persist run", httpx.CorrelationID(r), retryable)
		return
	}

	s.audit.Log(audit.Entry{
		TenantID: tenantID, PrincipalID: ac.PrincipalID, Action: "runs.create", Resource: "run/" + runID, Outcome: "allowed",
		CorrelationID: httpx.CorrelationID(r), RequestID: r.Header.Get("X-Request-Id"),
		Meta: map[string]any{"agent_id": agentID},
	})

	resp := types.RunCreateResponse{Run: run, CorrelationID: httpx.CorrelationID(r)}
	httpx.JSON(w, http.StatusCreated, resp)
}

func (s *Server) handleRuns(w http.ResponseWriter, r *http.Request) {
	ac, _ := auth.Get(r.Context())
	tenantID, ok := resolveTenant(w, r, ac)
	if !ok {
		return
	}
	if !s.tenantsExists(tenantID) {
		httpx.Error(w, http.StatusForbidden, "tenant_unknown", "tenant not found", httpx.CorrelationID(r), false)
		return
	}
	if !s.tenantsExists(tenantID) {
		httpx.Error(w, http.StatusForbidden, "tenant_unknown", "tenant not found", httpx.CorrelationID(r), false)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/v1/runs/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		httpx.Error(w, http.StatusNotFound, "not_found", "not found", httpx.CorrelationID(r), false)
		return
	}
	runID := parts[0]

	if len(parts) == 2 && parts[1] == "events" {
		s.handleEvents(w, r, tenantID, runID)
		return
	}

	if r.Method != http.MethodGet || len(parts) != 1 {
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		return
	}

	run, ok, err := s.runs.Get(r.Context(), tenantID, runID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "run_lookup_failed", "failed to load run", httpx.CorrelationID(r), true)
		return
	}
	if !ok {
		httpx.Error(w, http.StatusNotFound, "not_found", "run not found", httpx.CorrelationID(r), false)
		return
	}

	// For scaffold: auto-progress and auto-complete after ~5s
	updated := false
	if created, err := time.Parse(time.RFC3339, run.CreatedAt); err == nil {
		age := time.Since(created)
		if run.Status == "queued" && age > 1*time.Second {
			run.Status = "running"
			if run.StartedAt == "" {
				run.StartedAt = time.Now().UTC().Format(time.RFC3339)
			}
			updated = true
		}
		if (run.Status == "running" || run.Status == "queued") && age > 5*time.Second {
			run.Status = "completed"
			run.CompletedAt = time.Now().UTC().Format(time.RFC3339)
			s.limiter.DecConcurrent(tenantID)
			run.Output = &types.RunOutput{Type: "text", Text: "stub completed output"}
			updated = true
		}
	}

	if updated {
		if err := s.runs.Save(r.Context(), run); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "run_persist_failed", "failed to persist run update", httpx.CorrelationID(r), true)
			return
		}
	}

	resp := types.RunGetResponse{Run: run, CorrelationID: httpx.CorrelationID(r)}
	httpx.JSON(w, http.StatusOK, resp)
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request, tenantID, runID string) {
	if r.Method != http.MethodGet {
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		return
	}

	run, ok, err := s.runs.Get(r.Context(), tenantID, runID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "run_lookup_failed", "failed to load run", httpx.CorrelationID(r), true)
		return
	}
	if !ok {
		httpx.Error(w, http.StatusNotFound, "not_found", "run not found", httpx.CorrelationID(r), false)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, okf := w.(http.Flusher)
	if !okf {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	env := types.EventEnvelope{
		Event: types.Event{
			EventID:  id.New("evt"),
			Sequence: 1,
			Time:     time.Now().UTC().Format(time.RFC3339),
			Type:     "agentos.run.step.completed",
			TenantID: run.TenantID,
			AgentID:  run.AgentID,
			RunID:    run.RunID,
			StepID:   "step_1",
			Trace:    types.TraceContext{Traceparent: "00-00000000000000000000000000000000-0000000000000000-01"},
			Payload:  map[string]any{"status": "ok"},
		},
	}

	b, _ := json.Marshal(env)
	bw := bufio.NewWriter(w)
	_, _ = bw.WriteString("event: agentos.event\n")
	_, _ = bw.WriteString("data: " + string(b) + "\n\n")
	_ = bw.Flush()
	flusher.Flush()
}

func (s *Server) handleTenants(w http.ResponseWriter, r *http.Request) {
	ac, _ := auth.Get(r.Context())
	if !hasScope(ac.Scopes, "tenants:admin") {
		httpx.Error(w, http.StatusForbidden, "forbidden", "admin scope required", httpx.CorrelationID(r), false)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/v1/admin/tenants")
	path = strings.Trim(path, "/")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			list := s.tenants.List()
			httpx.JSON(w, http.StatusOK, map[string]any{"tenants": list, "correlation_id": httpx.CorrelationID(r)})
		case http.MethodPost:
			var t types.Tenant
			if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
				httpx.Error(w, http.StatusBadRequest, "invalid_json", "invalid json body", httpx.CorrelationID(r), false)
				return
			}
			if err := s.tenants.Create(t); err != nil {
				code := http.StatusBadRequest
				errCode := "invalid_request"
				if errors.Is(err, tenants.ErrTenantExists) {
					code = http.StatusConflict
					errCode = "conflict"
				}
				httpx.Error(w, code, errCode, err.Error(), httpx.CorrelationID(r), false)
				return
			}
			s.audit.Log(audit.Entry{
				TenantID: t.TenantID, PrincipalID: ac.PrincipalID, Action: "tenants.create", Resource: "tenant/" + t.TenantID, Outcome: "allowed",
				CorrelationID: httpx.CorrelationID(r), RequestID: r.Header.Get("X-Request-Id"),
			})
			httpx.JSON(w, http.StatusCreated, map[string]any{"tenant": t, "correlation_id": httpx.CorrelationID(r)})
		default:
			httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		}
		return
	}

	tenantID := path
	switch r.Method {
	case http.MethodGet:
		if t, ok := s.tenants.Get(tenantID); ok {
			httpx.JSON(w, http.StatusOK, map[string]any{"tenant": t, "correlation_id": httpx.CorrelationID(r)})
			return
		}
		httpx.Error(w, http.StatusNotFound, "not_found", "tenant not found", httpx.CorrelationID(r), false)
	case http.MethodPut, http.MethodPatch:
		var t types.Tenant
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			httpx.Error(w, http.StatusBadRequest, "invalid_json", "invalid json body", httpx.CorrelationID(r), false)
			return
		}
		updated, err := s.tenants.Update(tenantID, t)
		if err != nil {
			code := http.StatusBadRequest
			errCode := "invalid_request"
			if errors.Is(err, tenants.ErrNotFound) {
				code = http.StatusNotFound
				errCode = "not_found"
			}
			httpx.Error(w, code, errCode, err.Error(), httpx.CorrelationID(r), false)
			return
		}
		s.audit.Log(audit.Entry{
			TenantID: tenantID, PrincipalID: ac.PrincipalID, Action: "tenants.update", Resource: "tenant/" + tenantID, Outcome: "allowed",
			CorrelationID: httpx.CorrelationID(r), RequestID: r.Header.Get("X-Request-Id"),
		})
		httpx.JSON(w, http.StatusOK, map[string]any{"tenant": updated, "correlation_id": httpx.CorrelationID(r)})
	case http.MethodDelete:
		if _, err := s.tenants.Delete(tenantID); err != nil {
			code := http.StatusBadRequest
			errCode := "invalid_request"
			if errors.Is(err, tenants.ErrNotFound) {
				code = http.StatusNotFound
				errCode = "not_found"
			}
			httpx.Error(w, code, errCode, err.Error(), httpx.CorrelationID(r), false)
			return
		}
		s.audit.Log(audit.Entry{
			TenantID: tenantID, PrincipalID: ac.PrincipalID, Action: "tenants.delete", Resource: "tenant/" + tenantID, Outcome: "allowed",
			CorrelationID: httpx.CorrelationID(r), RequestID: r.Header.Get("X-Request-Id"),
		})
		httpx.JSON(w, http.StatusOK, map[string]any{"deleted": tenantID, "correlation_id": httpx.CorrelationID(r)})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
	}
}

func resolveTenant(w http.ResponseWriter, r *http.Request, ac auth.AuthContext) (string, bool) {
	tenantID, err := auth.RequireTenant(ac)
	if err != nil {
		code := http.StatusUnauthorized
		errCode := "unauthorized"
		msg := err.Error()
		retryable := false
		if errors.Is(err, auth.ErrTenantMismatch) {
			code = http.StatusBadRequest
			errCode = "tenant_mismatch"
		}
		httpx.Error(w, code, errCode, msg, httpx.CorrelationID(r), retryable)
		return "", false
	}
	return tenantID, true
}

func (s *Server) tenantsExists(tenantID string) bool {
	if tenantID == "" {
		return false
	}
	_, ok := s.tenants.Get(tenantID)
	return ok
}

func hasScope(scopes []string, target string) bool {
	target = strings.TrimSpace(target)
	for _, s := range scopes {
		if strings.EqualFold(strings.TrimSpace(s), target) {
			return true
		}
	}
	return false
}
