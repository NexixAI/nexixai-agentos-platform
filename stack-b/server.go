package stackb

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/audit"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/auth"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/httpx"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/metrics"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/middleware"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/quota"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/types"
)

type Server struct {
	version string
	limiter *quota.Limiter
	audit   audit.Logger
}

func New(version string) *Server {
	return &Server{
		version: version,
		limiter: quota.NewFromEnv("AGENTOS_QUOTA_INVOKE_QPS", "AGENTOS_QUOTA_UNUSED_CONCURRENT", 20, 999999),
		audit:   audit.NewFromEnv(),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/health", s.handleHealth)
	mux.HandleFunc("/v1/models", s.handleModels)
	mux.HandleFunc("/v1/models:invoke", s.handleInvoke)
	mux.HandleFunc("/v1/policy:check", s.handlePolicyCheck)
	mux.Handle("/metrics", middleware.ProtectMetrics(metrics.Handler()))

	h := middleware.WithAuth(mux)
	h = middleware.EnsureRequestID(h)
	h = metrics.Instrument("stack-b", h)
	return h
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		return
	}
	httpx.JSON(w, http.StatusOK, types.HealthResponse{Status: "ok", Service: "stack-b", Version: s.version})
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		return
	}
	resp := types.ModelsListResponse{
		Models: []types.Model{
			{
				ModelID:      "local-stub-llm",
				Provider:     "stub",
				DisplayName:  "Local Stub LLM",
				Capabilities: map[string]any{"chat": true, "embeddings": true},
			},
		},
	}
	httpx.JSON(w, http.StatusOK, resp)
}

func (s *Server) handleInvoke(w http.ResponseWriter, r *http.Request) {
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
		metrics.IncQuotaDenied("stack-b", "models_invoke_qps")
		httpx.Error(w, http.StatusTooManyRequests, "quota_exceeded", "invoke QPS exceeded", httpx.CorrelationID(r), true)
		s.audit.Log(audit.Entry{
			TenantID: tenantID, PrincipalID: ac.PrincipalID, Action: "models.invoke", Resource: "stack-b", Outcome: "denied",
			CorrelationID: httpx.CorrelationID(r), RequestID: r.Header.Get("X-Request-Id"),
			Meta: map[string]any{"reason": "qps_exceeded"},
		})
		return
	}

	var req types.ModelInvokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_json", "invalid json body", httpx.CorrelationID(r), false)
		return
	}

	out := map[string]any{
		"type": "text",
		"text": "stub response",
		"echo": req.Input,
		"ts":   time.Now().UTC().Format(time.RFC3339),
	}

	resp := types.ModelInvokeResponse{
		Output:        out,
		CorrelationID: httpx.CorrelationID(r),
	}

	s.audit.Log(audit.Entry{
		TenantID: tenantID, PrincipalID: ac.PrincipalID, Action: "models.invoke", Resource: "model/" + req.ModelID, Outcome: "allowed",
		CorrelationID: httpx.CorrelationID(r), RequestID: r.Header.Get("X-Request-Id"),
		Meta: map[string]any{"operation": req.Operation},
	})

	httpx.JSON(w, http.StatusOK, resp)
}

func (s *Server) handlePolicyCheck(w http.ResponseWriter, r *http.Request) {
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

	var req types.PolicyCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_json", "invalid json body", httpx.CorrelationID(r), false)
		return
	}

	resp := types.PolicyCheckResponse{
		Decision:      "allow",
		Reasons:       []string{"stub_policy_allows"},
		CorrelationID: httpx.CorrelationID(r),
	}

	s.audit.Log(audit.Entry{
		TenantID: tenantID, PrincipalID: ac.PrincipalID, Action: "policy.check", Resource: "policy", Outcome: "allowed",
		CorrelationID: httpx.CorrelationID(r), RequestID: r.Header.Get("X-Request-Id"),
		Meta: map[string]any{"action": req.Action},
	})

	httpx.JSON(w, http.StatusOK, resp)
}
