package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
)

type ContextKey string

const AuthContextKey ContextKey = "agentos_auth_context"

type AuthContext struct {
	TenantID         string   `json:"tenant_id"`
	PrincipalID      string   `json:"principal_id,omitempty"`
	Scopes           []string `json:"scopes,omitempty"`
	SubjectType      string   `json:"subject_type,omitempty"`
	APIKeyID         string   `json:"api_key_id,omitempty"`
	BearerToken      string   `json:"-"`
	TokenTenantID    string   `json:"-"`
	TokenPrincipalID string   `json:"-"`
}

// FromRequest derives auth context from headers and (optionally) a bearer token payload.
// v1.02 policy: tenant_id is primarily derived from auth context; an X-Tenant-Id header may be accepted in dev/demo.
func FromRequest(r *http.Request) AuthContext {
	headerTenant := strings.TrimSpace(r.Header.Get("X-Tenant-Id"))
	headerPrincipal := strings.TrimSpace(r.Header.Get("X-Principal-Id"))
	scopes := parseScopes(r.Header.Get("X-Scopes"))
	subjectType := strings.TrimSpace(r.Header.Get("X-Subject-Type"))
	apiKey := strings.TrimSpace(r.Header.Get("X-Api-Key-Id"))

	ac := AuthContext{
		TenantID:    headerTenant,
		PrincipalID: headerPrincipal,
		Scopes:      scopes,
		SubjectType: subjectType,
		APIKeyID:    apiKey,
	}

	authz := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
		token := strings.TrimSpace(authz[len("bearer "):])
		if token != "" {
			ac.BearerToken = token
			claims := parseJWTClaims(token)
			if v, ok := claims["tenant_id"].(string); ok {
				ac.TokenTenantID = strings.TrimSpace(v)
			} else if v, ok := claims["tid"].(string); ok {
				ac.TokenTenantID = strings.TrimSpace(v)
			}
			if v, ok := claims["principal_id"].(string); ok {
				ac.TokenPrincipalID = strings.TrimSpace(v)
			} else if v, ok := claims["sub"].(string); ok {
				ac.TokenPrincipalID = strings.TrimSpace(v)
			}
			if ac.PrincipalID == "" && ac.TokenPrincipalID != "" {
				ac.PrincipalID = ac.TokenPrincipalID
			}
			if ac.TenantID == "" && ac.TokenTenantID != "" {
				ac.TenantID = ac.TokenTenantID
			}
			if len(ac.Scopes) == 0 {
				if scp := parseScopesClaim(claims); len(scp) > 0 {
					ac.Scopes = scp
				}
			}
			if ac.SubjectType == "" {
				if v, ok := claims["subject_type"].(string); ok {
					ac.SubjectType = strings.TrimSpace(v)
				} else if v, ok := claims["principal_type"].(string); ok {
					ac.SubjectType = strings.TrimSpace(v)
				}
			}
			if ac.APIKeyID == "" {
				if v, ok := claims["api_key_id"].(string); ok {
					ac.APIKeyID = strings.TrimSpace(v)
				}
			}
		}
	}

	return ac
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

// DefaultTenant enables "works out of the box" local/dev deployments without a full identity system yet.
func DefaultTenant() string {
	return strings.TrimSpace(os.Getenv("AGENTOS_DEFAULT_TENANT"))
}

var (
	ErrTenantRequired = errors.New("tenant_id required")
	ErrTenantMismatch = errors.New("tenant_id mismatch between token and header")
)

// RequireTenant returns the resolved tenant_id or an error when missing or mismatched.
func RequireTenant(ac AuthContext) (string, error) {
	tokenTenant := strings.TrimSpace(ac.TokenTenantID)
	headerTenant := strings.TrimSpace(ac.TenantID)

	if tokenTenant != "" && headerTenant != "" && tokenTenant != headerTenant {
		return "", ErrTenantMismatch
	}

	if tokenTenant != "" {
		return tokenTenant, nil
	}
	if headerTenant != "" {
		return headerTenant, nil
	}
	if dt := DefaultTenant(); dt != "" {
		return dt, nil
	}
	return "", ErrTenantRequired
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

func parseScopesClaim(claims map[string]any) []string {
	if claims == nil {
		return nil
	}
	if v, ok := claims["scopes"].([]any); ok {
		return parseAnyScope(v)
	}
	if v, ok := claims["scope"].(string); ok {
		return parseScopes(v)
	}
	if v, ok := claims["scp"].([]any); ok {
		return parseAnyScope(v)
	}
	if v, ok := claims["scp"].(string); ok {
		return parseScopes(v)
	}
	return nil
}

func parseAnyScope(raw []any) []string {
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok {
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func parseJWTClaims(token string) map[string]any {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil
	}
	return claims
}
