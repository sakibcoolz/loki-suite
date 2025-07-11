package main

import (
	"testing"
	"time"

	"loki-suite/internal/config"
	"loki-suite/pkg/security"
)

func TestServiceIntegration(t *testing.T) {
	t.Run("Config Service Integration", func(t *testing.T) {
		cfg := config.LoadConfig()
		if err := cfg.Validate(); err != nil {
			t.Fatalf("Config validation failed: %v", err)
		}

		if cfg.Server.Port == "" {
			t.Error("Server port should not be empty")
		}

		if cfg.Database.Host == "" {
			t.Error("Database host should not be empty")
		}
	})

	t.Run("Security Service Integration", func(t *testing.T) {
		cfg := config.LoadConfig()

		// Initialize security service
		securitySvc := security.NewSecurityService(
			cfg.Security.JWTSecret,
			cfg.Security.HMACKeyLength,
			cfg.Security.JWTTokenExpiration,
		)

		// Test HMAC key generation
		hmacKey := securitySvc.GenerateHMACSecretKey()
		if len(hmacKey) == 0 {
			t.Error("HMAC key should not be empty")
		}

		// Test JWT token generation - correct parameter order: tenantID, webhookID, appName
		token, err := securitySvc.GenerateJWTToken("test-tenant", "test-webhook-id", "test-app")
		if err != nil {
			t.Fatalf("JWT token generation failed: %v", err)
		}

		if token == "" {
			t.Error("JWT token should not be empty")
		}

		// Test JWT token verification
		claims, err := securitySvc.VerifyJWTToken(token)
		if err != nil {
			t.Fatalf("JWT token verification failed: %v", err)
		}

		if claims.WebhookID != "test-webhook-id" {
			t.Errorf("Expected webhook_id 'test-webhook-id', got '%s'", claims.WebhookID)
		}

		if claims.TenantID != "test-tenant" {
			t.Errorf("Expected tenant_id 'test-tenant', got '%s'", claims.TenantID)
		}
	})

	t.Run("Service Layer Dependencies", func(t *testing.T) {
		cfg := config.LoadConfig()

		// Verify all required dependencies are available
		securitySvc := security.NewSecurityService(
			cfg.Security.JWTSecret,
			cfg.Security.HMACKeyLength,
			cfg.Security.JWTTokenExpiration,
		)

		if securitySvc == nil {
			t.Error("Security service should not be nil")
		}

		// Test HMAC operations
		payload := []byte("test payload")
		secretKey := "test-secret-key"

		signature := securitySvc.GenerateHMACSignature(payload, secretKey)
		if signature == "" {
			t.Error("HMAC signature should not be empty")
		}

		if !securitySvc.VerifyHMACSignature(payload, signature, secretKey) {
			t.Error("HMAC signature verification should succeed")
		}

		// Test timestamp validation
		currentTime := time.Now().Format(time.RFC3339)
		if !securitySvc.ValidateTimestamp(currentTime, 5) {
			t.Error("Current timestamp should be valid")
		}
	})
}
