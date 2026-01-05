package agentorchestrator

import (
	"bufio"
	"context"
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
	agents  storage.AgentStore
	tenants *tenants.Store
	limiter *quota.Limiter
	audit   audit.Logger
}

func New(version string) (*Server, error) {
	runStore, err := storage.NewRunStoreFromEnv()
	if err != nil {
		return nil, err
	}
	agentStore, err := storage.NewAgentStoreFromEnv()
	if err != nil {
		return nil, err
	}
	tenantStore := tenants.NewStore()
	defaultTenant := auth.DefaultTenant()
	if defaultTenant != "" {
		tenantStore.EnsureDefault(defaultTenant)
	}
	srv := &Server{
		version: version,
		runs:    runStore,
		agents:  agentStore,
		tenants: tenantStore,
		limiter: quota.NewFromEnv("AGENTOS_QUOTA_RUN_CREATE_QPS", "AGENTOS_QUOTA_CONCURRENT_RUNS", 10, 25),
		audit:   audit.NewFromEnv(),
	}
	// Seed demo agent for default tenant
	if defaultTenant != "" {
		srv.seedDemoAgent(defaultTenant)
	}
	return srv, nil
}

func (s *Server) seedDemoAgent(tenantID string) {
	now := time.Now().UTC().Format(time.RFC3339)
	demoAgent := types.Agent{
		AgentID:     "agt_demo",
		TenantID:    tenantID,
		Name:        "Demo Agent",
		Description: "Sample agent for validation",
		Version:     "1.0",
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	// Ignore error if agent already exists
	_ = s.agents.Create(context.Background(), demoAgent)
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
	h = metrics.Instrument("agent-orchestrator", h)
	return h
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		return
	}
	httpx.JSON(w, http.StatusOK, types.HealthResponse{Status: "ok", Service: "agent-orchestrator", Version: s.version})
}

func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	ac, _ := auth.Get(r.Context())
	tenantID, ok := resolveTenant(w, r, ac)
	if !ok {
		return
	}
	if !s.tenantsExists(tenantID) {
		httpx.Error(w, http.StatusForbidden, "tenant_unknown", "tenant not found", httpx.CorrelationID(r), false)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/v1/agents/")
	path = strings.Trim(path, "/")

	// GET /v1/agents - List agents
	if path == "" {
		if r.Method != http.MethodGet {
			httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
			return
		}
		s.handleAgentList(w, r, tenantID)
		return
	}

	parts := strings.Split(path, "/")

	// GET /v1/agents/{agent_id} - Get agent metadata
	if len(parts) == 1 && r.Method == http.MethodGet {
		s.handleAgentGet(w, r, tenantID, parts[0])
		return
	}

	// POST /v1/agents/{agent_id}/runs - Create run
	if len(parts) == 2 && parts[1] == "runs" && r.Method == http.MethodPost {
		s.handleRunCreate(w, r, tenantID, parts[0], ac)
		return
	}

	httpx.Error(w, http.StatusNotFound, "not_found", "not found", httpx.CorrelationID(r), false)
}

func (s *Server) handleAgentList(w http.ResponseWriter, r *http.Request, tenantID string) {
	agents, err := s.agents.List(r.Context(), tenantID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "agent_list_failed", "failed to list agents", httpx.CorrelationID(r), true)
		return
	}
	if agents == nil {
		agents = []types.Agent{}
	}
	resp := types.AgentListResponse{Agents: agents, CorrelationID: httpx.CorrelationID(r)}
	httpx.JSON(w, http.StatusOK, resp)
}

func (s *Server) handleAgentGet(w http.ResponseWriter, r *http.Request, tenantID, agentID string) {
	agent, ok, err := s.agents.Get(r.Context(), tenantID, agentID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "agent_lookup_failed", "failed to load agent", httpx.CorrelationID(r), true)
		return
	}
	if !ok {
		// Return 404 for non-existent or different tenant agent
		httpx.Error(w, http.StatusNotFound, "not_found", "agent not found", httpx.CorrelationID(r), false)
		return
	}
	resp := types.AgentGetResponse{Agent: agent, CorrelationID: httpx.CorrelationID(r)}
	httpx.JSON(w, http.StatusOK, resp)
}

func (s *Server) handleRunCreate(w http.ResponseWriter, r *http.Request, tenantID, agentID string, ac auth.AuthContext) {
	if !s.limiter.AllowQPS(tenantID) {
		metrics.IncQuotaDenied("agent-orchestrator", "runs_create_qps")
		httpx.Error(w, http.StatusTooManyRequests, "quota_exceeded", "run create QPS exceeded", httpx.CorrelationID(r), true)
		s.audit.Log(audit.Entry{
			TenantID: tenantID, PrincipalID: ac.PrincipalID, Action: "runs.create", Resource: "agent-orchestrator", Outcome: "denied",
			CorrelationID: httpx.CorrelationID(r), RequestID: r.Header.Get("X-Request-Id"),
			Meta: map[string]any{"reason": "qps_exceeded"},
		})
		return
	}
	if !s.limiter.TryIncConcurrent(tenantID) {
		metrics.IncQuotaDenied("agent-orchestrator", "runs_concurrency")
		httpx.Error(w, http.StatusTooManyRequests, "quota_exceeded", "concurrent runs exceeded", httpx.CorrelationID(r), true)
		s.audit.Log(audit.Entry{
			TenantID: tenantID, PrincipalID: ac.PrincipalID, Action: "runs.create", Resource: "agent-orchestrator", Outcome: "denied",
			CorrelationID: httpx.CorrelationID(r), RequestID: r.Header.Get("X-Request-Id"),
			Meta: map[string]any{"reason": "concurrent_exceeded"},
		})
		return
	}

	var req types.RunCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.limiter.DecConcurrent(tenantID)
		httpx.Error(w, http.StatusBadRequest, "invalid_json", "invalid json body", httpx.CorrelationID(r), false)
		return
	}

	// Check idempotency key if provided
	if req.IdempotencyKey != "" {
		existingRun, found, err := s.runs.GetByIdempotencyKey(r.Context(), tenantID, req.IdempotencyKey)
		if err != nil {
			// Fail-open for availability: log error and proceed with creation
			// In production, this should use structured logging
		} else if found {
			// Return existing run with 200 OK (not 201 Created)
			s.limiter.DecConcurrent(tenantID)
			resp := types.RunCreateResponse{Run: existingRun, CorrelationID: httpx.CorrelationID(r)}
			httpx.JSON(w, http.StatusOK, resp)
			return
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	runID := id.New("run")

	run := types.Run{
		TenantID:       tenantID,
		AgentID:        agentID,
		RunID:          runID,
		Status:         "queued",
		CreatedAt:      now,
		EventsURL:      "/v1/runs/" + runID + "/events",
		RunOptions:     req.RunOptions,
		IdempotencyKey: req.IdempotencyKey,
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

	// Check for :cancel action (e.g., /v1/runs/run_123:cancel)
	if strings.HasSuffix(runID, ":cancel") {
		runID = strings.TrimSuffix(runID, ":cancel")
		s.handleCancel(w, r, tenantID, runID, ac)
		return
	}

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

func (s *Server) handleCancel(w http.ResponseWriter, r *http.Request, tenantID, runID string, ac auth.AuthContext) {
	if r.Method != http.MethodPost {
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		return
	}

	run, ok, err := s.runs.Get(r.Context(), tenantID, runID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "run_lookup_failed", "failed to load run", httpx.CorrelationID(r), true)
		return
	}
	if !ok {
		// Tenant isolation: return 404 (not 403) for runs not in this tenant
		httpx.Error(w, http.StatusNotFound, "not_found", "run not found", httpx.CorrelationID(r), false)
		return
	}

	// State transition validation
	switch run.Status {
	case "completed", "failed", "canceled":
		// Cannot cancel from terminal states
		httpx.Error(w, http.StatusConflict, "invalid_state_transition", "cannot cancel run in "+run.Status+" state", httpx.CorrelationID(r), false)
		return
	case "queued", "running":
		// Can cancel from these states
	default:
		// Unknown state, allow cancellation
	}

	// Update run to canceled state
	run.Status = "canceled"
	run.CompletedAt = time.Now().UTC().Format(time.RFC3339)

	if err := s.runs.Save(r.Context(), run); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "run_persist_failed", "failed to persist run cancellation", httpx.CorrelationID(r), true)
		return
	}

	// Decrement concurrent run quota
	s.limiter.DecConcurrent(tenantID)

	// Audit log
	s.audit.Log(audit.Entry{
		TenantID: tenantID, PrincipalID: ac.PrincipalID, Action: "runs.cancel", Resource: "run/" + runID, Outcome: "allowed",
		CorrelationID: httpx.CorrelationID(r), RequestID: r.Header.Get("X-Request-Id"),
	})

	resp := types.RunCancelResponse{Run: run, CorrelationID: httpx.CorrelationID(r)}
	httpx.JSON(w, http.StatusOK, resp)
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
