package federation

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/httpx"
)

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token expired")
	ErrInvalidSignature  = errors.New("invalid signature")
	ErrUnsupportedAlg    = errors.New("unsupported algorithm")
	ErrPublicKeyNotFound = errors.New("public key not found")
)

// JWTClaims represents the claims extracted from a JWT token.
type JWTClaims struct {
	TenantID    string `json:"tenant_id"`
	PrincipalID string `json:"principal_id"`
	Subject     string `json:"sub"`
	Issuer      string `json:"iss"`
	Audience    string `json:"aud"`
	ExpiresAt   int64  `json:"exp"`
	IssuedAt    int64  `json:"iat"`
}

// JWTVerifier verifies JWT tokens using a public key.
type JWTVerifier struct {
	publicKey crypto.PublicKey
}

// NewJWTVerifier creates a JWT verifier from environment configuration.
// Returns nil if public key not configured (dev mode - skip verification).
func NewJWTVerifier() *JWTVerifier {
	publicKeyPath := strings.TrimSpace(os.Getenv("AGENTOS_FED_JWT_PUBLIC_KEY"))
	if publicKeyPath == "" {
		return nil
	}

	keyData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil
	}

	publicKey, err := parsePublicKey(keyData)
	if err != nil {
		return nil
	}

	return &JWTVerifier{publicKey: publicKey}
}

// parsePublicKey parses a PEM-encoded public key.
func parsePublicKey(data []byte) (crypto.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	switch block.Type {
	case "PUBLIC KEY":
		return x509.ParsePKIXPublicKey(block.Bytes)
	case "RSA PUBLIC KEY":
		return x509.ParsePKCS1PublicKey(block.Bytes)
	default:
		return nil, errors.New("unsupported key type: " + block.Type)
	}
}

// Verify verifies a JWT token and extracts claims.
func (v *JWTVerifier) Verify(tokenStr string) (*JWTClaims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	// Decode header
	headerData, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var header struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
	if err := json.Unmarshal(headerData, &header); err != nil {
		return nil, ErrInvalidToken
	}

	// Decode payload
	payloadData, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims JWTClaims
	if err := json.Unmarshal(payloadData, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	// Decode signature
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Verify signature
	signingInput := parts[0] + "." + parts[1]
	if err := v.verifySignature(header.Alg, signingInput, signature); err != nil {
		return nil, err
	}

	// Check expiration
	if claims.ExpiresAt > 0 && time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrTokenExpired
	}

	return &claims, nil
}

// verifySignature verifies the JWT signature.
func (v *JWTVerifier) verifySignature(alg string, signingInput string, signature []byte) error {
	data := []byte(signingInput)

	switch alg {
	case "RS256", "RS384", "RS512":
		rsaKey, ok := v.publicKey.(*rsa.PublicKey)
		if !ok {
			return ErrInvalidSignature
		}
		return verifyRSA(alg, rsaKey, data, signature)

	case "ES256", "ES384", "ES512":
		ecKey, ok := v.publicKey.(*ecdsa.PublicKey)
		if !ok {
			return ErrInvalidSignature
		}
		return verifyECDSA(alg, ecKey, data, signature)

	case "EdDSA":
		edKey, ok := v.publicKey.(ed25519.PublicKey)
		if !ok {
			return ErrInvalidSignature
		}
		if !ed25519.Verify(edKey, data, signature) {
			return ErrInvalidSignature
		}
		return nil

	default:
		return ErrUnsupportedAlg
	}
}

// verifyRSA verifies RSA signatures.
func verifyRSA(alg string, key *rsa.PublicKey, data, signature []byte) error {
	var hash crypto.Hash
	switch alg {
	case "RS256":
		hash = crypto.SHA256
	case "RS384":
		hash = crypto.SHA384
	case "RS512":
		hash = crypto.SHA512
	default:
		return ErrUnsupportedAlg
	}

	h := hash.New()
	h.Write(data)
	hashed := h.Sum(nil)

	return rsa.VerifyPKCS1v15(key, hash, hashed, signature)
}

// verifyECDSA verifies ECDSA signatures.
func verifyECDSA(alg string, key *ecdsa.PublicKey, data, signature []byte) error {
	var hash crypto.Hash
	switch alg {
	case "ES256":
		hash = crypto.SHA256
	case "ES384":
		hash = crypto.SHA384
	case "ES512":
		hash = crypto.SHA512
	default:
		return ErrUnsupportedAlg
	}

	h := hash.New()
	h.Write(data)
	hashed := h.Sum(nil)

	if !ecdsa.VerifyASN1(key, hashed, signature) {
		return ErrInvalidSignature
	}
	return nil
}

// JWTMiddleware returns an HTTP middleware that verifies JWT tokens.
// If no verifier is configured, it passes through (dev mode).
func JWTMiddleware(verifier *JWTVerifier, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip verification if no verifier configured (dev mode)
		if verifier == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Extract bearer token
		authHeader := r.Header.Get("Authorization")
		token := ""
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token = strings.TrimSpace(authHeader[7:])
		}

		if token == "" {
			httpx.Error(w, http.StatusUnauthorized, "unauthorized", "missing bearer token", httpx.CorrelationID(r), false)
			return
		}

		// Verify token
		claims, err := verifier.Verify(token)
		if err != nil {
			errMsg := "invalid token"
			if errors.Is(err, ErrTokenExpired) {
				errMsg = "token expired"
			}
			httpx.Error(w, http.StatusUnauthorized, "unauthorized", errMsg, httpx.CorrelationID(r), false)
			return
		}

		// Add claims to headers for downstream processing
		if claims.TenantID != "" {
			r.Header.Set("X-JWT-Tenant-Id", claims.TenantID)
		}
		if claims.PrincipalID != "" {
			r.Header.Set("X-JWT-Principal-Id", claims.PrincipalID)
		}
		if claims.Subject != "" {
			r.Header.Set("X-JWT-Subject", claims.Subject)
		}

		next.ServeHTTP(w, r)
	})
}
