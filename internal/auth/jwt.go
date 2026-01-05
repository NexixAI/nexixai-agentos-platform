package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrJWTInvalid        = errors.New("jwt: invalid token")
	ErrJWTExpired        = errors.New("jwt: token expired")
	ErrJWTVerifyDisabled = errors.New("jwt: verification disabled (dev mode)")
)

// JWTVerifier verifies JWT signatures using RS256 (RSA with SHA-256).
type JWTVerifier struct {
	publicKey *rsa.PublicKey
	enabled   bool
}

// NewJWTVerifierFromEnv creates a JWT verifier from environment configuration.
// If AGENTOS_FED_JWT_PUBLIC_KEY or AGENTOS_FED_JWT_PUBLIC_KEY_FILE is not set,
// returns a disabled verifier (dev mode).
func NewJWTVerifierFromEnv() (*JWTVerifier, error) {
	// Load public key from env or file
	pubKeyPEM := strings.TrimSpace(os.Getenv("AGENTOS_FED_JWT_PUBLIC_KEY"))
	if pubKeyPEM == "" {
		if path := strings.TrimSpace(os.Getenv("AGENTOS_FED_JWT_PUBLIC_KEY_FILE")); path != "" {
			b, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read JWT public key file: %w", err)
			}
			pubKeyPEM = string(b)
		}
	}

	if pubKeyPEM == "" {
		// Dev mode: JWT verification disabled
		return &JWTVerifier{enabled: false}, nil
	}

	// Parse PEM block
	block, _ := pem.Decode([]byte(pubKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block for JWT public key")
	}

	// Parse public key
	var pubKey *rsa.PublicKey
	parsedKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT public key: %w", err)
	}

	var ok bool
	pubKey, ok = parsedKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("JWT public key is not RSA")
	}

	return &JWTVerifier{
		publicKey: pubKey,
		enabled:   true,
	}, nil
}

// VerifyAndExtract verifies the JWT signature and extracts claims.
// Returns tenant_id, principal_id, and any error.
// If verification is disabled (dev mode), returns ErrJWTVerifyDisabled and still extracts claims.
func (v *JWTVerifier) VerifyAndExtract(tokenString string) (tenantID, principalID string, err error) {
	if !v.enabled {
		// Dev mode: verification disabled, return empty claims
		return "", "", ErrJWTVerifyDisabled
	}

	// Parse and verify token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.publicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", "", ErrJWTExpired
		}
		return "", "", fmt.Errorf("%w: %v", ErrJWTInvalid, err)
	}

	if !token.Valid {
		return "", "", ErrJWTInvalid
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", ErrJWTInvalid
	}

	tenant, principal := extractTenantAndPrincipal(map[string]any(claims))
	return tenant, principal, nil
}

func extractTenantAndPrincipal(claims map[string]any) (string, string) {
	var tenantID, principalID string

	// Extract tenant_id
	if v, ok := claims["tenant_id"].(string); ok {
		tenantID = strings.TrimSpace(v)
	} else if v, ok := claims["tid"].(string); ok {
		tenantID = strings.TrimSpace(v)
	}

	// Extract principal_id
	if v, ok := claims["principal_id"].(string); ok {
		principalID = strings.TrimSpace(v)
	} else if v, ok := claims["sub"].(string); ok {
		principalID = strings.TrimSpace(v)
	}

	return tenantID, principalID
}

// Enabled returns true if JWT verification is enabled.
func (v *JWTVerifier) Enabled() bool {
	return v != nil && v.enabled
}
