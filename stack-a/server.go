    package stacka

    import (
        "bufio"
        "encoding/json"
        "net/http"
        "strings"
        "sync"
        "time"

        "github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/httpx"
        "github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/id"
        "github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/types"
    )

    type Server struct {
        version string
        mu      sync.Mutex
        runs    map[string]types.Run
    }

    func New(version string) *Server {
        return &Server{
            version: version,
            runs:    make(map[string]types.Run),
        }
    }

    func (s *Server) Handler() http.Handler {
        mux := http.NewServeMux()
        mux.HandleFunc("/v1/health", s.handleHealth)
        mux.HandleFunc("/v1/agents/", s.handleAgents) // expects /v1/agents/{agent_id}/runs
        mux.HandleFunc("/v1/runs/", s.handleRuns)     // expects /v1/runs/{run_id} and /v1/runs/{run_id}/events
        return mux
    }

    func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
            return
        }
        httpx.JSON(w, http.StatusOK, types.HealthResponse{Status: "ok", Service: "stack-a", Version: s.version})
    }

    func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
        // /v1/agents/{agent_id}/runs
        if r.Method != http.MethodPost {
            httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
            return
        }
        path := strings.TrimPrefix(r.URL.Path, "/v1/agents/")
        parts := strings.Split(strings.Trim(path, "/"), "/")
        if len(parts) != 2 || parts[1] != "runs" {
            httpx.Error(w, http.StatusNotFound, "not_found", "not found", httpx.CorrelationID(r), false)
            return
        }
        agentID := parts[0]

        var req types.RunCreateRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            httpx.Error(w, http.StatusBadRequest, "invalid_json", "invalid json body", httpx.CorrelationID(r), false)
            return
        }

        now := time.Now().UTC().Format(time.RFC3339)
        runID := id.New("run")
        tenantID := r.Header.Get("X-Tenant-Id")
        if tenantID == "" {
            tenantID = "tnt_demo"
        }

        run := types.Run{
            TenantID:  tenantID,
            AgentID:   agentID,
            RunID:     runID,
            Status:    "queued",
            CreatedAt: now,
            EventsURL: "/v1/runs/" + runID + "/events",
            RunOptions: req.RunOptions,
        }

        s.mu.Lock()
        s.runs[runID] = run
        s.mu.Unlock()

        resp := types.RunCreateResponse{
            Run:           run,
            CorrelationID: httpx.CorrelationID(r),
        }
        w.Header().Set("X-Request-Id", id.New("req"))
        httpx.JSON(w, http.StatusCreated, resp)
    }

    func (s *Server) handleRuns(w http.ResponseWriter, r *http.Request) {
        // /v1/runs/{run_id} OR /v1/runs/{run_id}/events
        path := strings.TrimPrefix(r.URL.Path, "/v1/runs/")
        parts := strings.Split(strings.Trim(path, "/"), "/")
        if len(parts) == 0 || parts[0] == "" {
            httpx.Error(w, http.StatusNotFound, "not_found", "not found", httpx.CorrelationID(r), false)
            return
        }
        runID := parts[0]

        if len(parts) == 2 && parts[1] == "events" {
            s.handleEvents(w, r, runID)
            return
        }

        if r.Method != http.MethodGet || len(parts) != 1 {
            httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
            return
        }

        s.mu.Lock()
        run, ok := s.runs[runID]
        s.mu.Unlock()
        if !ok {
            httpx.Error(w, http.StatusNotFound, "not_found", "run not found", httpx.CorrelationID(r), false)
            return
        }

        // For demo: mark as running quickly
        if run.Status == "queued" {
            run.Status = "running"
            s.mu.Lock()
            s.runs[runID] = run
            s.mu.Unlock()
        }

        resp := types.RunGetResponse{Run: run, CorrelationID: httpx.CorrelationID(r)}
        w.Header().Set("X-Request-Id", id.New("req"))
        httpx.JSON(w, http.StatusOK, resp)
    }

    func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request, runID string) {
        if r.Method != http.MethodGet {
            httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
            return
        }
        s.mu.Lock()
        run, ok := s.runs[runID]
        s.mu.Unlock()
        if !ok {
            httpx.Error(w, http.StatusNotFound, "not_found", "run not found", httpx.CorrelationID(r), false)
            return
        }

        w.Header().Set("Content-Type", "text/event-stream")
        w.Header().Set("Cache-Control", "no-cache")
        w.Header().Set("Connection", "keep-alive")
        w.Header().Set("X-Request-Id", id.New("req"))

        flusher, okf := w.(http.Flusher)
        if !okf {
            http.Error(w, "streaming unsupported", http.StatusInternalServerError)
            return
        }

        // Write one event then close (stub). Clients can reconnect as needed.
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
        _, _ = bw.WriteString("event: agentos.event
")
        _, _ = bw.WriteString("data: " + string(b) + "

")
        _ = bw.Flush()
        flusher.Flush()
    }
