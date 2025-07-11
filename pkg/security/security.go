package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the JWT claims for private webhooks
type JWTClaims struct {
	TenantID  string `json:"tenant_id"`
	WebhookID string `json:"webhook_id"`
	AppName   string `json:"app_name"`
	jwt.RegisteredClaims
}

// SecurityService handles all security-related operations
type SecurityService struct {
	jwtSecret               string
	hmacKeyLength           int
	jwtTokenExpirationHours int
}

// NewSecurityService creates a new security service
func NewSecurityService(jwtSecret string, hmacKeyLength, jwtTokenExpirationHours int) *SecurityService {
	return &SecurityService{
		jwtSecret:               jwtSecret,
		hmacKeyLength:           hmacKeyLength,
		jwtTokenExpirationHours: jwtTokenExpirationHours,
	}
}

// GenerateHMACSecretKey generates a secure hex-encoded secret key
func (s *SecurityService) GenerateHMACSecretKey() string {
	b := make([]byte, s.hmacKeyLength)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateJWTToken generates a JWT token for private webhooks
func (s *SecurityService) GenerateJWTToken(tenantID, webhookID, appName string) (string, error) {
	claims := JWTClaims{
		TenantID:  tenantID,
		WebhookID: webhookID,
		AppName:   appName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtTokenExpirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "loki-suite",
			Subject:   webhookID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// VerifyJWTToken verifies and parses a JWT token
func (s *SecurityService) VerifyJWTToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GenerateHMACSignature generates HMAC-SHA256 signature for the given payload and secret
func (s *SecurityService) GenerateHMACSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyHMACSignature verifies the HMAC signature
func (s *SecurityService) VerifyHMACSignature(payload []byte, signature, secret string) bool {
	expectedSignature := s.GenerateHMACSignature(payload, secret)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// ValidateTimestamp checks if the timestamp is within the allowed tolerance
func (s *SecurityService) ValidateTimestamp(timestampStr string, toleranceMinutes int) bool {
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return false
	}

	now := time.Now()
	tolerance := time.Duration(toleranceMinutes) * time.Minute

	return timestamp.After(now.Add(-tolerance)) && timestamp.Before(now.Add(tolerance))
}

// ExtractSignatureFromHeader extracts the signature from the X-Shavix-Signature header
// Expected format: "sha256=<hex_signature>"
func (s *SecurityService) ExtractSignatureFromHeader(signatureHeader string) (string, error) {
	if signatureHeader == "" {
		return "", fmt.Errorf("signature header is empty")
	}

	const prefix = "sha256="
	if !strings.HasPrefix(signatureHeader, prefix) {
		return "", fmt.Errorf("invalid signature format")
	}

	return signatureHeader[len(prefix):], nil
}

// ExtractBearerToken extracts JWT token from Authorization header
// Expected format: "Bearer <jwt_token>"
func (s *SecurityService) ExtractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is empty")
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", fmt.Errorf("invalid authorization format")
	}

	return authHeader[len(prefix):], nil
}

// WebhookSecurityData holds all security information for a webhook
type WebhookSecurityData struct {
	SecretToken string  `json:"secret_token"`
	JWTToken    *string `json:"jwt_token,omitempty"`
}

// GenerateWebhookSecurity generates both HMAC secret and JWT token if needed
func (s *SecurityService) GenerateWebhookSecurity(isPrivate bool, tenantID, webhookID, appName string) (*WebhookSecurityData, error) {
	data := &WebhookSecurityData{
		SecretToken: s.GenerateHMACSecretKey(),
	}

	if isPrivate {
		jwtToken, err := s.GenerateJWTToken(tenantID, webhookID, appName)
		if err != nil {
			return nil, fmt.Errorf("failed to generate JWT token: %w", err)
		}
		data.JWTToken = &jwtToken
	}

	return data, nil
}
