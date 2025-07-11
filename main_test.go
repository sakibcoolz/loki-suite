package main

import (
	"testing"
	"time"

	"loki-suite/internal/config"
	"loki-suite/pkg/security"

	"github.com/stretchr/testify/assert"
)

func TestSecurityFunctions(t *testing.T) {
	securitySvc := security.NewSecurityService("test-jwt-secret", 32, 24)

	t.Run("Generate HMAC Secret Key", func(t *testing.T) {
		key1 := securitySvc.GenerateHMACSecretKey()
		key2 := securitySvc.GenerateHMACSecretKey()

		assert.Len(t, key1, 64) // 32 bytes = 64 hex chars
		assert.Len(t, key2, 64)
		assert.NotEqual(t, key1, key2) // Should be different
	})

	t.Run("JWT Token Generation and Verification", func(t *testing.T) {
		tenantID := "test-tenant"
		webhookID := "123e4567-e89b-12d3-a456-426614174000"
		appName := "test-app"

		// Generate JWT token
		token, err := securitySvc.GenerateJWTToken(tenantID, webhookID, appName)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify JWT token
		claims, err := securitySvc.VerifyJWTToken(token)
		assert.NoError(t, err)
		assert.Equal(t, tenantID, claims.TenantID)
		assert.Equal(t, webhookID, claims.WebhookID)
		assert.Equal(t, appName, claims.AppName)
		assert.Equal(t, "loki-suite", claims.Issuer)
	})

	t.Run("HMAC Signature Generation and Verification", func(t *testing.T) {
		payload := []byte(`{"test": "data"}`)
		secret := "test-secret"

		signature := securitySvc.GenerateHMACSignature(payload, secret)
		assert.NotEmpty(t, signature)

		// Verify correct signature
		assert.True(t, securitySvc.VerifyHMACSignature(payload, signature, secret))

		// Verify incorrect signature
		assert.False(t, securitySvc.VerifyHMACSignature(payload, "wrong-signature", secret))

		// Verify with wrong secret
		assert.False(t, securitySvc.VerifyHMACSignature(payload, signature, "wrong-secret"))
	})

	t.Run("Timestamp Validation", func(t *testing.T) {
		// Valid timestamp (current time)
		now := time.Now().Format(time.RFC3339)
		assert.True(t, securitySvc.ValidateTimestamp(now, 5))

		// Valid timestamp (4 minutes ago)
		past := time.Now().Add(-4 * time.Minute).Format(time.RFC3339)
		assert.True(t, securitySvc.ValidateTimestamp(past, 5))

		// Invalid timestamp (10 minutes ago)
		oldPast := time.Now().Add(-10 * time.Minute).Format(time.RFC3339)
		assert.False(t, securitySvc.ValidateTimestamp(oldPast, 5))

		// Invalid timestamp format
		assert.False(t, securitySvc.ValidateTimestamp("invalid-timestamp", 5))
	})

	t.Run("Extract Signature From Header", func(t *testing.T) {
		// Valid signature header
		signature, err := securitySvc.ExtractSignatureFromHeader("sha256=abcdef123456")
		assert.NoError(t, err)
		assert.Equal(t, "abcdef123456", signature)

		// Invalid format
		_, err = securitySvc.ExtractSignatureFromHeader("invalid-format")
		assert.Error(t, err)

		// Empty header
		_, err = securitySvc.ExtractSignatureFromHeader("")
		assert.Error(t, err)
	})

	t.Run("Extract Bearer Token", func(t *testing.T) {
		// Valid bearer token
		token, err := securitySvc.ExtractBearerToken("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")
		assert.NoError(t, err)
		assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", token)

		// Invalid format
		_, err = securitySvc.ExtractBearerToken("invalid-format")
		assert.Error(t, err)

		// Empty header
		_, err = securitySvc.ExtractBearerToken("")
		assert.Error(t, err)
	})

	t.Run("Generate Webhook Security", func(t *testing.T) {
		// Public webhook
		publicSecurity, err := securitySvc.GenerateWebhookSecurity(false, "tenant-1", "webhook-1", "app-1")
		assert.NoError(t, err)
		assert.NotEmpty(t, publicSecurity.SecretToken)
		assert.Nil(t, publicSecurity.JWTToken)

		// Private webhook
		privateSecurity, err := securitySvc.GenerateWebhookSecurity(true, "tenant-1", "webhook-1", "app-1")
		assert.NoError(t, err)
		assert.NotEmpty(t, privateSecurity.SecretToken)
		assert.NotNil(t, privateSecurity.JWTToken)
		assert.NotEmpty(t, *privateSecurity.JWTToken)
	})
}

func TestConfigLoading(t *testing.T) {
	t.Run("Load Default Config", func(t *testing.T) {
		cfg := config.LoadConfig()

		// Should have default values
		assert.Equal(t, "8080", cfg.Server.Port)
		assert.Equal(t, "localhost", cfg.Database.Host)
		assert.Equal(t, 30, cfg.Webhook.TimeoutSeconds)
		assert.Equal(t, 3, cfg.Webhook.MaxRetries)
		assert.Equal(t, 5, cfg.Security.TimestampToleranceMinutes)
		assert.Equal(t, 32, cfg.Security.HMACKeyLength)
		assert.Equal(t, 24, cfg.Security.JWTTokenExpiration)
	})

	t.Run("Config Validation", func(t *testing.T) {
		cfg := config.LoadConfig()
		err := cfg.Validate()
		assert.NoError(t, err)
	})
}
