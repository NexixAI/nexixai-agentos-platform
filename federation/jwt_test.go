package federation

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestJWTVerifyValidToken(t *testing.T) {
	// Generate test RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	verifier := &JWTVerifier{publicKey: &privateKey.PublicKey}

	// Create a valid token
	header := map[string]string{"alg": "RS256", "typ": "JWT"}
	claims := map[string]any{
		"tenant_id":    "tenant_123",
		"principal_id": "principal_456",
		"sub":          "user@example.com",
		"iss":          "agentos",
		"exp":          time.Now().Add(time.Hour).Unix(),
		"iat":          time.Now().Unix(),
	}

	token := createTestToken(t, header, claims, privateKey)

	result, err := verifier.Verify(token)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if result.TenantID != "tenant_123" {
		t.Errorf("expected tenant_id tenant_123, got %s", result.TenantID)
	}
	if result.PrincipalID != "principal_456" {
		t.Errorf("expected principal_id principal_456, got %s", result.PrincipalID)
	}
	if result.Subject != "user@example.com" {
		t.Errorf("expected subject user@example.com, got %s", result.Subject)
	}
}

func TestJWTVerifyExpiredToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	verifier := &JWTVerifier{publicKey: &privateKey.PublicKey}

	header := map[string]string{"alg": "RS256", "typ": "JWT"}
	claims := map[string]any{
		"tenant_id": "tenant_123",
		"exp":       time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		"iat":       time.Now().Add(-2 * time.Hour).Unix(),
	}

	token := createTestToken(t, header, claims, privateKey)

	_, err = verifier.Verify(token)
	if err != ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestJWTVerifyInvalidSignature(t *testing.T) {
	// Generate two different key pairs
	privateKey1, _ := rsa.GenerateKey(rand.Reader, 2048)
	privateKey2, _ := rsa.GenerateKey(rand.Reader, 2048)

	// Verifier uses key 2's public key
	verifier := &JWTVerifier{publicKey: &privateKey2.PublicKey}

	// Token signed with key 1
	header := map[string]string{"alg": "RS256", "typ": "JWT"}
	claims := map[string]any{
		"tenant_id": "tenant_123",
		"exp":       time.Now().Add(time.Hour).Unix(),
	}

	token := createTestToken(t, header, claims, privateKey1)

	_, err := verifier.Verify(token)
	if err == nil {
		t.Error("expected signature verification to fail")
	}
}

func TestJWTVerifyMalformedToken(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	verifier := &JWTVerifier{publicKey: &privateKey.PublicKey}

	testCases := []struct {
		name  string
		token string
	}{
		{"empty", ""},
		{"no dots", "invalidtoken"},
		{"one dot", "part1.part2"},
		{"too many parts", "a.b.c.d"},
		{"invalid base64 header", "!!!.abc.def"},
		{"invalid base64 payload", "eyJhbGciOiJSUzI1NiJ9.!!!.def"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := verifier.Verify(tc.token)
			if err != ErrInvalidToken {
				t.Errorf("expected ErrInvalidToken for %s, got %v", tc.name, err)
			}
		})
	}
}

func TestJWTVerifyECDSA(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ECDSA key: %v", err)
	}

	verifier := &JWTVerifier{publicKey: &privateKey.PublicKey}

	header := map[string]string{"alg": "ES256", "typ": "JWT"}
	claims := map[string]any{
		"tenant_id": "tenant_ecdsa",
		"exp":       time.Now().Add(time.Hour).Unix(),
	}

	token := createTestTokenECDSA(t, header, claims, privateKey)

	result, err := verifier.Verify(token)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if result.TenantID != "tenant_ecdsa" {
		t.Errorf("expected tenant_id tenant_ecdsa, got %s", result.TenantID)
	}
}

func TestJWTMiddlewarePassthrough(t *testing.T) {
	// Test that middleware passes through when no verifier configured (dev mode)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := JWTMiddleware(nil, handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK with nil verifier, got %d", rec.Code)
	}
}

func TestJWTMiddlewareMissingToken(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	verifier := &JWTVerifier{publicKey: &privateKey.PublicKey}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := JWTMiddleware(verifier, handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized without token, got %d", rec.Code)
	}
}

func TestJWTMiddlewareSetsHeaders(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	verifier := &JWTVerifier{publicKey: &privateKey.PublicKey}

	var capturedTenantID, capturedPrincipalID, capturedSubject string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTenantID = r.Header.Get("X-JWT-Tenant-Id")
		capturedPrincipalID = r.Header.Get("X-JWT-Principal-Id")
		capturedSubject = r.Header.Get("X-JWT-Subject")
		w.WriteHeader(http.StatusOK)
	})

	wrapped := JWTMiddleware(verifier, handler)

	header := map[string]string{"alg": "RS256", "typ": "JWT"}
	claims := map[string]any{
		"tenant_id":    "t_test",
		"principal_id": "p_test",
		"sub":          "sub_test",
		"exp":          time.Now().Add(time.Hour).Unix(),
	}
	token := createTestToken(t, header, claims, privateKey)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
	if capturedTenantID != "t_test" {
		t.Errorf("expected X-JWT-Tenant-Id t_test, got %s", capturedTenantID)
	}
	if capturedPrincipalID != "p_test" {
		t.Errorf("expected X-JWT-Principal-Id p_test, got %s", capturedPrincipalID)
	}
	if capturedSubject != "sub_test" {
		t.Errorf("expected X-JWT-Subject sub_test, got %s", capturedSubject)
	}
}

// Helper function to create a test RSA-signed JWT token
func createTestToken(t *testing.T, header map[string]string, claims map[string]any, privateKey *rsa.PrivateKey) string {
	t.Helper()

	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signingInput := headerB64 + "." + claimsB64

	h := sha256.New()
	h.Write([]byte(signingInput))
	hashed := h.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + signatureB64
}

// Helper function to create a test ECDSA-signed JWT token
func createTestTokenECDSA(t *testing.T, header map[string]string, claims map[string]any, privateKey *ecdsa.PrivateKey) string {
	t.Helper()

	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signingInput := headerB64 + "." + claimsB64

	h := sha256.New()
	h.Write([]byte(signingInput))
	hashed := h.Sum(nil)

	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, hashed)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + signatureB64
}
