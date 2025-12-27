package modelpolicy

import (
	"strings"

	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/auth"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/types"
)

type policyEngine struct{}

func newPolicyEngine() *policyEngine {
	return &policyEngine{}
}

func (p *policyEngine) Evaluate(tenantID string, ac auth.AuthContext, req types.ModelInvokeRequest) (string, []string) {
	if denies := denyReasons(req); len(denies) > 0 {
		return "deny", denies
	}
	if len(ac.Scopes) > 0 && !hasScope(ac.Scopes, "models:invoke") {
		return "deny", []string{"scope_missing:models:invoke"}
	}
	if tenantID == "" {
		return "deny", []string{"tenant_missing"}
	}
	return "allow", []string{"policy_allow"}
}

func (p *policyEngine) EvaluatePolicyCheck(tenantID string, ac auth.AuthContext, req types.PolicyCheckRequest) (string, []string) {
	if tenantID == "" {
		return "deny", []string{"tenant_missing"}
	}
	if strings.EqualFold(req.Action, "deny") {
		return "deny", []string{"explicit_deny_action"}
	}
	if len(ac.Scopes) > 0 && !hasScope(ac.Scopes, "policy:check") {
		return "deny", []string{"scope_missing:policy:check"}
	}
	return "allow", []string{"policy_allow"}
}

func denyReasons(req types.ModelInvokeRequest) []string {
	reasons := []string{}
	if val, ok := req.Options["deny"].(bool); ok && val {
		reasons = append(reasons, "option_deny=true")
	}
	if strings.EqualFold(req.Operation, "deny") || strings.EqualFold(req.Operation, "block") {
		reasons = append(reasons, "operation_blocked")
	}
	return reasons
}

func hasScope(scopes []string, target string) bool {
	for _, s := range scopes {
		if strings.EqualFold(strings.TrimSpace(s), target) {
			return true
		}
	}
	return false
}
