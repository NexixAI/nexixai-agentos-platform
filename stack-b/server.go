package stackb

import (
    "encoding/json"
    "net/http"
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
    mux.HandleFunc("/v1/health", s.handleHealth)
    mux.HandleFunc("/v1/models", s.handleModels)
    mux.HandleFunc("/v1/models:invoke", s.handleInvoke)
    mux.HandleFunc("/v1/policy:check", s.handlePolicyCheck)
    return mux
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
                ModelID:     "local-stub-llm",
                Provider:    "stub",
                DisplayName: "Local Stub LLM",
                Capabilities: map[string]any{"chat": true, "embeddings": true},
            },
        },
    }
    w.Header().Set("X-Request-Id", id.New("req"))
    httpx.JSON(w, http.StatusOK, resp)
}

func (s *Server) handleInvoke(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
        return
    }
    var req types.ModelInvokeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        httpx.Error(w, http.StatusBadRequest, "invalid_json", "invalid json body", httpx.CorrelationID(r), false)
        return
    }

    // Stub output: echo input with a minimal "text" result.
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
    w.Header().Set("X-Request-Id", id.New("req"))
    httpx.JSON(w, http.StatusOK, resp)
}

func (s *Server) handlePolicyCheck(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        httpx.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.CorrelationID(r), false)
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
    w.Header().Set("X-Request-Id", id.New("req"))
    httpx.JSON(w, http.StatusOK, resp)
}
