package modelpolicy

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/audit"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/auth"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/httpx"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/metrics"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/middleware"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/quota"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/types"
)

type Server struct {
	version   string
	limiter   *quota.Limiter
	audit     audit.Logger
	providers *registry
	policy    *policyEngine
	usage     *usageMeter
}

func New(version string) *Server {
	return &Server{
		version:   version,
		limiter:   quota.NewFromEnv("AGENTOS_QUOTA_INVOKE_QPS", "AGENTOS_QUOTA_UNUSED_CONCURRENT", 20, 999999),
		audit:     audit.NewFromEnv(),
		providers: newRegistry(),
		policy:    newPolicyEngine(),
		usage:     newUsageMeter(),
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
	h = metrics.Instrument("model-policy", h)
	return h
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		return
	}
	httpx.JSON(w, http.StatusOK, types.HealthResponse{Status: "ok", Service: "model-policy", Version: s.version})
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		return
	}
	resp := types.ModelsListResponse{Models: s.providers.Models()}
	httpx.JSON(w, http.StatusOK, resp)
}

func (s *Server) handleInvoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		return
	}

	ac, _ := auth.Get(r.Context())
	tenantID, ok := resolveTenant(w, r, ac)
	if !ok {
		return
	}
	if !s.limiter.AllowQPS(tenantID) {
		metrics.IncQuotaDenied("model-policy", "models_invoke_qps")
		httpx.Error(w, http.StatusTooManyRequests, "quota_exceeded", "invoke QPS exceeded", httpx.CorrelationID(r), true)
		s.audit.Log(audit.Entry{
			TenantID: tenantID, PrincipalID: ac.PrincipalID, Action: "models.invoke", Resource: "model-policy", Outcome: "denied",
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

	decision, reasons := s.policy.Evaluate(tenantID, ac, req)
	if decision != "allow" {
		httpx.Error(w, http.StatusForbidden, "policy_blocked", strings.Join(reasons, "; "), httpx.CorrelationID(r), false)
		s.audit.Log(audit.Entry{
			TenantID: tenantID, PrincipalID: ac.PrincipalID, Action: "models.invoke", Resource: "model/" + req.ModelID, Outcome: "denied",
			CorrelationID: httpx.CorrelationID(r), RequestID: r.Header.Get("X-Request-Id"),
			Meta: map[string]any{"policy_reasons": reasons},
		})
		return
	}

	prov, model, ok := s.providers.Resolve(req.ModelID)
	if !ok {
		httpx.Error(w, http.StatusNotFound, "model_not_found", "model not found", httpx.CorrelationID(r), false)
		return
	}

	output, usage, err := prov.Invoke(req)
	if err != nil {
		httpx.Error(w, http.StatusBadGateway, "provider_error", err.Error(), httpx.CorrelationID(r), true)
		return
	}
	s.usage.Record(tenantID, usage)

	resp := types.ModelInvokeResponse{
		Output:        output,
		Usage:         usage,
		CorrelationID: httpx.CorrelationID(r),
	}

	s.audit.Log(audit.Entry{
		TenantID: tenantID, PrincipalID: ac.PrincipalID, Action: "models.invoke", Resource: "model/" + req.ModelID, Outcome: "allowed",
		CorrelationID: httpx.CorrelationID(r), RequestID: r.Header.Get("X-Request-Id"),
		Meta: map[string]any{"operation": req.Operation, "model_id": model.ModelID, "provider": model.Provider},
	})

	httpx.JSON(w, http.StatusOK, resp)
}

func (s *Server) handlePolicyCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
		return
	}

	ac, _ := auth.Get(r.Context())
	tenantID, ok := resolveTenant(w, r, ac)
	if !ok {
		return
	}

	var req types.PolicyCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_json", "invalid json body", httpx.CorrelationID(r), false)
		return
	}

	decision, reasons := s.policy.EvaluatePolicyCheck(tenantID, ac, req)
	resp := types.PolicyCheckResponse{
		Decision:      decision,
		Reasons:       reasons,
		CorrelationID: httpx.CorrelationID(r),
	}

	s.audit.Log(audit.Entry{
		TenantID: tenantID, PrincipalID: ac.PrincipalID, Action: "policy.check", Resource: "policy", Outcome: decision,
		CorrelationID: httpx.CorrelationID(r), RequestID: r.Header.Get("X-Request-Id"),
		Meta: map[string]any{"action": req.Action, "resource": req.Resource},
	})

	httpx.JSON(w, http.StatusOK, resp)
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
