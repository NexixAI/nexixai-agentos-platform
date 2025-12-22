package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequireTenantPrefersTokenAndRejectsMismatch(t *testing.T) {
	ac := AuthContext{TenantID: "tnt_header", TokenTenantID: "tnt_token"}
	if _, err := RequireTenant(ac); err == nil || !errors.Is(err, ErrTenantMismatch) {
		t.Fatalf("expected ErrTenantMismatch, got %v", err)
	}
}

func TestRequireTenantFromBearerToken(t *testing.T) {
	ac := AuthContext{TokenTenantID: "tnt_token", TokenPrincipalID: "usr_token"}
	tenant, err := RequireTenant(ac)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tenant != "tnt_token" {
		t.Fatalf("unexpected tenant: %s", tenant)
	}
}

func TestFromRequestParsesBearerClaims(t *testing.T) {
	token := makeJWT(map[string]any{
		"tenant_id":      "tnt_jwt",
		"principal_id":   "usr_jwt",
		"subject_type":   "user",
		"api_key_id":     "key123",
		"scope":          "runs:create runs:read",
		"principal_type": "user",
	})

	req := httptest.NewRequest("GET", "/v1/health", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	ac := FromRequest(req)

	if ac.TokenTenantID != "tnt_jwt" || ac.TenantID != "tnt_jwt" {
		t.Fatalf("expected tenant from jwt, got token=%s header=%s", ac.TokenTenantID, ac.TenantID)
	}
	if ac.PrincipalID != "usr_jwt" || ac.TokenPrincipalID != "usr_jwt" {
		t.Fatalf("expected principal from jwt, got %s/%s", ac.PrincipalID, ac.TokenPrincipalID)
	}
	if ac.SubjectType != "user" {
		t.Fatalf("expected subject_type user, got %s", ac.SubjectType)
	}
	if ac.APIKeyID != "key123" {
		t.Fatalf("expected api_key_id key123, got %s", ac.APIKeyID)
	}
	if len(ac.Scopes) != 2 || ac.Scopes[0] != "runs:create" {
		t.Fatalf("unexpected scopes: %+v", ac.Scopes)
	}
}

func makeJWT(claims map[string]any) string {
	b, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(b)
	return strings.Join([]string{"hdr", payload, "sig"}, ".")
}
