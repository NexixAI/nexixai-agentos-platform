    package federation

    import (
        "bufio"
        "encoding/json"
        "net/http"
        "strings"
        "time"

        "github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/httpx"
        "github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/id"
        "github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/types"
    )

    type Server struct {
        version string
    }

    func New(version string) *Server {
        return &Server{version: version}
    }

    func (s *Server) Handler() http.Handler {
        mux := http.NewServeMux()
        mux.HandleFunc("/v1/federation/health", s.handleHealth)
        mux.HandleFunc("/v1/federation/peers", s.handlePeers)
        mux.HandleFunc("/v1/federation/peers/", s.handlePeerSubroutes)
        mux.HandleFunc("/v1/federation/runs:forward", s.handleForwardRun)
        mux.HandleFunc("/v1/federation/events:ingest", s.handleEventsIngest)
        mux.HandleFunc("/v1/federation/runs/", s.handleRunEvents) // /v1/federation/runs/{run_id}/events
        return mux
    }

    func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
            return
        }
        httpx.JSON(w, http.StatusOK, types.HealthResponse{Status: "ok", Service: "federation", Version: s.version})
    }

    func (s *Server) handlePeers(w http.ResponseWriter, r *http.Request) {
        // Minimal: respond with a single local peer.
        if r.Method != http.MethodGet {
            httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
            return
        }
        resp := map[string]any{
            "peers": []map[string]any{
                {"peer_id": "peer_local", "name": "Local Node", "base_url": "http://localhost:8083"},
            },
        }
        w.Header().Set("X-Request-Id", id.New("req"))
        httpx.JSON(w, http.StatusOK, resp)
    }

    func (s *Server) handlePeerSubroutes(w http.ResponseWriter, r *http.Request) {
        // /v1/federation/peers/{peer_id}/capabilities
        path := strings.TrimPrefix(r.URL.Path, "/v1/federation/peers/")
        parts := strings.Split(strings.Trim(path, "/"), "/")
        if len(parts) == 2 && parts[1] == "capabilities" && r.Method == http.MethodGet {
            peerID := parts[0]
            resp := types.PeerCapabilitiesResponse{
                PeerID:       peerID,
                Protocol:     "1.0",
                Capabilities: []string{"runs.forward", "events.ingest", "events.sse_proxy"},
                EventBackhaul: map[string]any{"mode": "sse_proxy"},
            }
            w.Header().Set("X-Request-Id", id.New("req"))
            httpx.JSON(w, http.StatusOK, resp)
            return
        }
        if len(parts) == 1 && r.Method == http.MethodGet {
            peerID := parts[0]
            resp := types.PeerInfoResponse{PeerID: peerID, Name: "Stub Peer", BaseURL: "http://example.invalid"}
            w.Header().Set("X-Request-Id", id.New("req"))
            httpx.JSON(w, http.StatusOK, resp)
            return
        }
        httpx.Error(w, http.StatusNotFound, "not_found", "not found", httpx.CorrelationID(r), false)
    }

    func (s *Server) handleForwardRun(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
            return
        }
        var req types.FederationForwardRunRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            httpx.Error(w, http.StatusBadRequest, "invalid_json", "invalid json body", httpx.CorrelationID(r), false)
            return
        }

        // Stub: "accept" and return a synthesized run
        runID := id.New("run")
        now := time.Now().UTC().Format(time.RFC3339)
        run := types.Run{
            TenantID:  req.Auth.TenantID,
            AgentID:   "agt_forwarded",
            RunID:     runID,
            Status:    "queued",
            CreatedAt: now,
            EventsURL: "/v1/federation/runs/" + runID + "/events",
        }

        resp := types.FederationForwardRunResponse{
            ForwardedTo:  map[string]any{"peer_id": "peer_stub", "remote_run_id": runID},
            Run:          run,
            CorrelationID: httpx.CorrelationID(r),
        }
        w.Header().Set("X-Request-Id", id.New("req"))
        httpx.JSON(w, http.StatusOK, resp)
    }

    func (s *Server) handleEventsIngest(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
            return
        }
        var req types.FederationEventIngestRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            httpx.Error(w, http.StatusBadRequest, "invalid_json", "invalid json body", httpx.CorrelationID(r), false)
            return
        }

        resp := types.FederationEventIngestResponse{
            Accepted:     len(req.Events),
            Rejected:     0,
            CorrelationID: httpx.CorrelationID(r),
        }
        w.Header().Set("X-Request-Id", id.New("req"))
        httpx.JSON(w, http.StatusOK, resp)
    }

    func (s *Server) handleRunEvents(w http.ResponseWriter, r *http.Request) {
        // /v1/federation/runs/{run_id}/events
        if r.Method != http.MethodGet {
            httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
            return
        }
        path := strings.TrimPrefix(r.URL.Path, "/v1/federation/runs/")
        parts := strings.Split(strings.Trim(path, "/"), "/")
        if len(parts) != 2 || parts[1] != "events" {
            httpx.Error(w, http.StatusNotFound, "not_found", "not found", httpx.CorrelationID(r), false)
            return
        }
        runID := parts[0]

        w.Header().Set("Content-Type", "text/event-stream")
        w.Header().Set("Cache-Control", "no-cache")
        w.Header().Set("Connection", "keep-alive")
        w.Header().Set("X-Request-Id", id.New("req"))

        flusher, ok := w.(http.Flusher)
        if !ok {
            http.Error(w, "streaming unsupported", http.StatusInternalServerError)
            return
        }

        env := types.EventEnvelope{
            Event: types.Event{
                EventID:  id.New("evt"),
                Sequence: 1,
                Time:     time.Now().UTC().Format(time.RFC3339),
                Type:     "agentos.federation.event",
                TenantID: "tnt_demo",
                AgentID:  "agt_remote",
                RunID:    runID,
                Trace:    types.TraceContext{Traceparent: "00-00000000000000000000000000000000-0000000000000000-01"},
                Payload:  map[string]any{"note": "stub federated event"},
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
