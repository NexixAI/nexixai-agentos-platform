\
package stacka

import (
	"bufio"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/audit"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/auth"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/httpx"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/id"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/middleware"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/quota"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/types"
)

type Server struct {
	version string

	mu   sync.Mutex
	runs map[string]map[string]types.Run // tenant -> run_id -> run

	limiter *quota.Limiter
	audit   audit.Logger
}

func New(version string) *Server {
	return &Server{
		version: version,
		runs:    make(map[string]map[string]types.Run),
		limiter: quota.NewFromEnv("AGENTOS_QUOTA_RUN_CREATE_QPS", "AGENTOS_QUOTA_CONCURRENT_RUNS", 10, 25),
		audit:   audit.NewFromEnv(),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/health", s.handleHealth)
	mux.HandleFunc("/v1/agents/", s.handleAgents) // /v1/agents/{agent_id}/runs
	mux.HandleFunc("/v1/runs/", s.handleRuns)     // /v1/runs/{run_id} and /v1/runs/{run_id}/events

	h := middleware.WithAuth(mux)
	h = middleware.EnsureRequestID(h)
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
	tenantID, ok := auth.RequireTenant(ac)
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized", "tenant_id required", httpx.CorrelationID(r), false)
		return
	}

	if !s.limiter.AllowQPS(tenantID) {
		httpx.Error(w, http.StatusTooManyRequests, "quota_exceeded", "run create QPS exceeded", httpx.CorrelationID(r), true)
		s.audit.Log(audit.Entry{
			TenantID: tenantID, PrincipalID: ac.PrincipalID, Action: "runs.create", Resource: "stack-a", Outcome: "denied",
			CorrelationID: httpx.CorrelationID(r), RequestID: r.Header.Get("X-Request-Id"),
			Meta: map[string]any{"reason": "qps_exceeded"},
		})
		return
	}
	if !s.limiter.TryIncConcurrent(tenantID) {
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

	s.mu.Lock()
	if s.runs[tenantID] == nil {
		s.runs[tenantID] = make(map[string]types.Run)
	}
	s.runs[tenantID][runID] = run
	s.mu.Unlock()

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
	tenantID, ok := auth.RequireTenant(ac)
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized", "tenant_id required", httpx.CorrelationID(r), false)
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

	s.mu.Lock()
	run, ok := s.runs[tenantID][runID]
	s.mu.Unlock()
	if !ok {
		httpx.Error(w, http.StatusNotFound, "not_found", "run not found", httpx.CorrelationID(r), false)
		return
	}

	// For scaffold: auto-progress and auto-complete after ~5s
	created, _ := time.Parse(time.RFC3339, run.CreatedAt)
	age := time.Since(created)
	if run.Status == "queued" && age > 1*time.Second {
		run.Status = "running"
	}
	if (run.Status == "running" || run.Status == "queued") && age > 5*time.Second {
		run.Status = "completed"
		run.CompletedAt = time.Now().UTC().Format(time.RFC3339)
		s.limiter.DecConcurrent(tenantID)
		run.Output = &types.RunOutput{Type: "text", Text: "stub completed output"}
	}

	s.mu.Lock()
	s.runs[tenantID][runID] = run
	s.mu.Unlock()

	resp := types.RunGetResponse{Run: run, CorrelationID: httpx.CorrelationID(r)}
	httpx.JSON(w, http.StatusOK, resp)
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request, tenantID, runID string) {
	if r.Method != http.MethodGet {
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		return
	}

	s.mu.Lock()
	run, ok := s.runs[tenantID][runID]
	s.mu.Unlock()
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
			TenantID:  run.TenantID,
			AgentID:   run.AgentID,
			RunID:     run.RunID,
			StepID:    "step_1",
			Trace:     types.TraceContext{Traceparent: "00-00000000000000000000000000000000-0000000000000000-01"},
			Payload:   map[string]any{"status": "ok"},
		},
	}

	b, _ := json.Marshal(env)
	bw := bufio.NewWriter(w)
	_, _ = bw.WriteString("event: agentos.event\n")
	_, _ = bw.WriteString("data: " + string(b) + "\n\n")
	_ = bw.Flush()
	flusher.Flush()
}
