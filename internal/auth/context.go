package auth

import (
	"context"
	"net/http"
	"os"
	"strings"
)

type ContextKey string

const AuthContextKey ContextKey = "agentos_auth_context"

type AuthContext struct {
	TenantID    string   `json:"tenant_id"`
	PrincipalID string   `json:"principal_id,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
	SubjectType string   `json:"subject_type,omitempty"`
	APIKeyID    string   `json:"api_key_id,omitempty"`
}

// FromRequest derives auth context from headers.
// v1.02 policy: tenant_id is primarily derived from auth context; an X-Tenant-Id header may be accepted in dev/demo.
func FromRequest(r *http.Request) AuthContext {
	t := strings.TrimSpace(r.Header.Get("X-Tenant-Id"))
	p := strings.TrimSpace(r.Header.Get("X-Principal-Id"))
	scopes := parseScopes(r.Header.Get("X-Scopes"))
	st := strings.TrimSpace(r.Header.Get("X-Subject-Type"))
	return AuthContext{
		TenantID:    t,
		PrincipalID: p,
		Scopes:      scopes,
		SubjectType: st,
	}
}

func WithContext(ctx context.Context, ac AuthContext) context.Context {
	return context.WithValue(ctx, AuthContextKey, ac)
}

func Get(ctx context.Context) (AuthContext, bool) {
	v := ctx.Value(AuthContextKey)
	if v == nil {
		return AuthContext{}, false
	}
	ac, ok := v.(AuthContext)
	return ac, ok
}

// DefaultTenant enables “works out of the box” local/dev deployments without a full identity system yet.
func DefaultTenant() string {
	return strings.TrimSpace(os.Getenv("AGENTOS_DEFAULT_TENANT"))
}

// RequireTenant returns (tenant_id, true) if either an explicit tenant is provided, or a default is configured.
func RequireTenant(ac AuthContext) (string, bool) {
	if ac.TenantID != "" {
		return ac.TenantID, true
	}
	if dt := DefaultTenant(); dt != "" {
		return dt, true
	}
	return "", false
}

func parseScopes(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	raw = strings.ReplaceAll(raw, ",", " ")
	parts := strings.Fields(raw)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
