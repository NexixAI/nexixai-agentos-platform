package federation

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/audit"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/auth"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/httpx"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/metrics"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/middleware"
)

type Server struct {
	version string

	registry    *Registry
	forward     *Forwarder
	proxy       *SSEProxy
	index       *forwardIndex
	events      *eventStore
	jwtVerifier *auth.JWTVerifier

	audit audit.Logger
}

func New(version string) *Server {
	reg, _ := LoadRegistryFromEnv()
	idxPath := os.Getenv("AGENTOS_FED_FORWARD_INDEX_FILE")
	if idxPath == "" {
		idxPath = "data/federation/forward-index.json"
	}

	// Initialize JWT verifier (may be disabled in dev mode)
	jwtVerifier, err := auth.NewJWTVerifierFromEnv()
	if err != nil {
		// Log error but continue - JWT verification will be disabled
		audit.NewFromEnv().Log(audit.Entry{
			Action:  "federation.jwt_init_failed",
			Outcome: "error",
			Meta:    map[string]any{"error": err.Error()},
		})
	}

	return &Server{
		version:     version,
		registry:    reg,
		forward:     NewForwarder(),
		proxy:       NewSSEProxy(),
		index:       newForwardIndexPersistent(idxPath),
		events:      newEventStore(),
		jwtVerifier: jwtVerifier,
		audit:       audit.NewFromEnv(),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Required by Federation OpenAPI
	mux.HandleFunc("/v1/federation/health", s.handleHealth)
	mux.HandleFunc("/v1/federation/peer", s.handlePeerInfo)
	mux.HandleFunc("/v1/federation/peer/capabilities", s.handlePeerCapabilities)
	mux.HandleFunc("/v1/federation/runs:forward", s.handleForwardRun)
	mux.HandleFunc("/v1/federation/runs/", s.handleRunEvents) // /v1/federation/runs/{run_id}/events
	mux.HandleFunc("/v1/federation/events:ingest", s.handleEventsIngest)
	mux.Handle("/metrics", middleware.ProtectMetrics(metrics.Handler()))

	h := s.withJWTVerification(mux)
	h = middleware.WithAuth(h)
	h = middleware.EnsureRequestID(h)
	h = metrics.Instrument("federation", h)
	return h
}

// withJWTVerification adds JWT verification middleware.
// Verifies bearer token if JWT verification is enabled.
// Returns 401 Unauthorized if verification fails.
func (s *Server) withJWTVerification(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip verification if JWT verifier not configured or not enabled
		if s.jwtVerifier == nil || !s.jwtVerifier.Enabled() {
			next.ServeHTTP(w, r)
			return
		}

		// Extract bearer token
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			// No bearer token - allow request (may be authenticated by other means)
			next.ServeHTTP(w, r)
			return
		}

		token := strings.TrimSpace(authHeader[len("bearer "):])
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Verify JWT and extract claims
		tenantID, principalID, err := s.jwtVerifier.VerifyAndExtract(token)
		if err != nil && !errors.Is(err, auth.ErrJWTVerifyDisabled) {
			// Verification failed
			httpx.Error(w, http.StatusUnauthorized, "jwt_verification_failed", "JWT verification failed", httpx.CorrelationID(r), false)
			return
		}

		// Add extracted claims to headers for downstream auth middleware
		if tenantID != "" {
			r.Header.Set("X-Tenant-Id", tenantID)
		}
		if principalID != "" {
			r.Header.Set("X-Principal-Id", principalID)
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "federation", "version": s.version})
}

func (s *Server) handlePeerInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		return
	}
	if s.registry == nil {
		httpx.Error(w, http.StatusServiceUnavailable, "unavailable", "peer registry not configured (AGENTOS_PEERS_FILE)", httpx.CorrelationID(r), true)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"peer": s.registry.Local})
}

func (s *Server) handlePeerCapabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		return
	}
	resp := map[string]any{
		"peer_id":  strings.TrimSpace(os.Getenv("AGENTOS_STACK_ID")),
		"protocol": "1.0",
		"capabilities": []string{
			"runs.forward",
			"events.ingest",
			"events.sse_proxy",
		},
		"event_backhaul": map[string]any{"mode": "sse_proxy"},
	}
	if resp["peer_id"] == "" {
		resp["peer_id"] = "stk_local"
	}
	httpx.JSON(w, http.StatusOK, resp)
}

func (s *Server) handleForwardRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		return
	}
	if s.registry == nil {
		httpx.Error(w, http.StatusServiceUnavailable, "unavailable", "peer registry not configured (AGENTOS_PEERS_FILE)", httpx.CorrelationID(r), true)
		return
	}

	ac, _ := auth.Get(r.Context())
	tenantHdr, err := auth.RequireTenant(ac)
	if err != nil {
		if errors.Is(err, auth.ErrTenantMismatch) {
			httpx.Error(w, http.StatusBadRequest, "tenant_mismatch", err.Error(), httpx.CorrelationID(r), false)
			return
		}
		tenantHdr = "" // allow payload to supply tenant when header/default missing
	}

	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_json", "invalid json body", httpx.CorrelationID(r), false)
		return
	}

	forwardObj, _ := req["forward"].(map[string]any)
	if forwardObj == nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", "missing forward object", httpx.CorrelationID(r), false)
		return
	}

	selector, _ := forwardObj["target_selector"].(map[string]any)
	authObj, _ := forwardObj["auth"].(map[string]any)
	runReq, _ := forwardObj["run_request"].(map[string]any)
	if selector == nil || authObj == nil || runReq == nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", "missing required forward fields", httpx.CorrelationID(r), false)
		return
	}

	targetStackID, _ := selector["stack_id"].(string)
	if targetStackID == "" {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", "target_selector.stack_id required", httpx.CorrelationID(r), false)
		return
	}

	tenantPayload, _ := authObj["tenant_id"].(string)
	principalPayload, _ := authObj["principal_id"].(string)

	tenantID := tenantPayload
	if tenantID == "" {
		tenantID = tenantHdr
	}
	if tenantID == "" {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized", "tenant_id required", httpx.CorrelationID(r), false)
		return
	}
	if tenantHdr != "" && tenantPayload != "" && tenantHdr != tenantPayload {
		httpx.Error(w, http.StatusBadRequest, "tenant_mismatch", "tenant_id mismatch between header and payload auth", httpx.CorrelationID(r), false)
		return
	}

	peer, ok := s.registry.Get(targetStackID)
	if !ok {
		httpx.Error(w, http.StatusNotFound, "peer_not_found", "target peer not found", httpx.CorrelationID(r), false)
		return
	}

	agentID, _ := runReq["agent_id"].(string)
	if agentID == "" {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", "run_request.agent_id required", httpx.CorrelationID(r), false)
		return
	}

	runCreate := map[string]any{
		"input":           runReq["input"],
		"context":         runReq["context"],
		"tooling":         runReq["tooling"],
		"run_options":     runReq["run_options"],
		"idempotency_key": runReq["idempotency_key"],
	}

	bearer := bearerToken(r.Header.Get("Authorization"))

	remoteRunID, remoteEventsURL, status, err := s.forward.ForwardRun(peer.Endpoints.AgentOrchestratorBaseURL, agentID, tenantID, principalPayload, bearer, runCreate)
	if err != nil {
		metrics.IncFederationForwardFailure("federation", "forward_run_failed")
		httpx.Error(w, http.StatusBadGateway, "forward_failed", err.Error(), httpx.CorrelationID(r), true)
		return
	}

	s.index.Set(tenantID, remoteRunID, peer.StackID, remoteEventsURL)

	resp := map[string]any{
		"forwarded": map[string]any{
			"tenant_id":         tenantID,
			"remote_stack_id":   peer.StackID,
			"remote_run_id":     remoteRunID,
			"remote_events_url": remoteEventsURL,
			"status":            status,
		},
		"correlation_id": httpx.CorrelationID(r),
	}

	s.audit.Log(audit.Entry{
		TenantID: tenantID, PrincipalID: principalPayload, Action: "federation.runs.forward", Resource: "run/" + remoteRunID, Outcome: "allowed",
		CorrelationID: httpx.CorrelationID(r), RequestID: r.Header.Get("X-Request-Id"),
		Meta: map[string]any{"target_stack_id": peer.StackID},
	})

	httpx.JSON(w, http.StatusOK, resp)
}

func (s *Server) handleEventsIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		return
	}

	ac, _ := auth.Get(r.Context())
	tenantHdr, err := auth.RequireTenant(ac)
	if err != nil {
		if errors.Is(err, auth.ErrTenantMismatch) {
			httpx.Error(w, http.StatusBadRequest, "tenant_mismatch", err.Error(), httpx.CorrelationID(r), false)
			return
		}
		tenantHdr = ""
	}

	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_json", "invalid json body", httpx.CorrelationID(r), false)
		return
	}

	peerID, _ := req["peer_id"].(string)
	authObj, _ := req["auth"].(map[string]any)
	eventsArr, _ := req["events"].([]any)
	if authObj == nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", "missing auth", httpx.CorrelationID(r), false)
		return
	}

	tenantPayload, _ := authObj["tenant_id"].(string)
	principalPayload, _ := authObj["principal_id"].(string)

	tenantID := tenantPayload
	if tenantID == "" {
		tenantID = tenantHdr
	}
	if tenantID == "" {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized", "tenant_id required", httpx.CorrelationID(r), false)
		return
	}
	if tenantHdr != "" && tenantPayload != "" && tenantHdr != tenantPayload {
		httpx.Error(w, http.StatusBadRequest, "tenant_mismatch", "tenant_id mismatch between header and payload auth", httpx.CorrelationID(r), false)
		return
	}

	acceptedTotal := 0
	rejectedTotal := 0

	for _, ev := range eventsArr {
		env, ok := ev.(map[string]any)
		if !ok {
			rejectedTotal++
			continue
		}
		eventObj, _ := env["event"].(map[string]any)
		runID, _ := eventObj["run_id"].(string)
		if runID == "" {
			rejectedTotal++
			continue
		}
		a, rj := s.events.Ingest(tenantID, runID, []map[string]any{env})
		acceptedTotal += a
		rejectedTotal += rj
	}

	resp := map[string]any{
		"accepted":       acceptedTotal,
		"rejected":       rejectedTotal,
		"correlation_id": httpx.CorrelationID(r),
	}

	s.audit.Log(audit.Entry{
		TenantID: tenantID, PrincipalID: principalPayload, Action: "federation.events.ingest", Resource: "peer/" + peerID, Outcome: "allowed",
		CorrelationID: httpx.CorrelationID(r), RequestID: r.Header.Get("X-Request-Id"),
		Meta: map[string]any{"accepted": acceptedTotal, "rejected": rejectedTotal},
	})

	httpx.JSON(w, http.StatusOK, resp)
}

func (s *Server) handleRunEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		return
	}

	ac, _ := auth.Get(r.Context())
	tenantID, ok := resolveTenant(w, r, ac)
	if !ok {
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/v1/federation/runs/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[1] != "events" {
		httpx.Error(w, http.StatusNotFound, "not_found", "not found", httpx.CorrelationID(r), false)
		return
	}
	runID := parts[0]

	fromSeq := 0
	if v := r.URL.Query().Get("from_sequence"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			fromSeq = parsed
		}
	}

	// Prefer SSE proxy if forwarded.
	if tgt, ok := s.index.Get(tenantID, runID); ok && tgt.RemoteEventsURL != "" {
		bearer := bearerToken(r.Header.Get("Authorization"))
		if err := s.proxy.Proxy(w, tgt.RemoteEventsURL, tenantID, ac.PrincipalID, bearer, fromSeq); err != nil {
			metrics.IncFederationForwardFailure("federation", "events_proxy_failed")
			httpx.Error(w, http.StatusBadGateway, "events_proxy_failed", err.Error(), httpx.CorrelationID(r), true)
		}
		return
	}

	// Otherwise, stream ingested events (push mode).
	if envs, ok := s.events.ListFromSequence(tenantID, runID, fromSeq); ok {
		_ = StreamStoredEvents(w, envs)
		return
	}

	httpx.Error(w, http.StatusNotFound, "not_found", "run not found", httpx.CorrelationID(r), false)
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

func bearerToken(header string) string {
	h := strings.TrimSpace(header)
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[len("bearer "):])
	}
	return ""
}
