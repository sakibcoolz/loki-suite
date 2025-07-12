package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sakibcoolz/zcornor/pkg/config"
	"github.com/sakibcoolz/zcornor/pkg/security"

	"github.com/sakibcoolz/loki-suite/internal/models"
	"github.com/sakibcoolz/loki-suite/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// WebhookService defines the interface for webhook business logic
type WebhookService interface {
	GenerateWebhook(req *models.GenerateWebhookRequest) (*models.GenerateWebhookResponse, error)
	SubscribeWebhook(req *models.SubscribeWebhookRequest) (*models.GenerateWebhookResponse, error)
	SendEvent(req *models.SendEventRequest) (*models.EventProcessingResult, error)
	VerifyWebhook(webhookID uuid.UUID, payload []byte, signature, timestamp, authHeader string) error
	ListWebhooks(tenantID string, page, limit int) (*models.WebhookListResponse, error)
	SetChainService(chainService ExecutionChainService)
}

// webhookService implements WebhookService
type webhookService struct {
	repo         repository.WebhookRepository
	securitySvc  *security.SecurityService
	config       *config.Config
	httpClient   *http.Client
	chainService ExecutionChainService
}

// NewWebhookService creates a new webhook service
func NewWebhookService(
	repo repository.WebhookRepository,
	securitySvc *security.SecurityService,
	cfg *config.Config,
) WebhookService {
	return &webhookService{
		repo:        repo,
		securitySvc: securitySvc,
		config:      cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second, // Default timeout
		},
		chainService: nil, // Will be set via SetChainService
	}
}

// SetChainService sets the execution chain service (for avoiding circular dependencies)
func (s *webhookService) SetChainService(chainService ExecutionChainService) {
	s.chainService = chainService
}

func (s *webhookService) GenerateWebhook(req *models.GenerateWebhookRequest) (*models.GenerateWebhookResponse, error) {
	// Validate webhook type
	if req.Type != models.WebhookTypePublic && req.Type != models.WebhookTypePrivate {
		return nil, fmt.Errorf("invalid webhook type: %s", req.Type)
	}

	// Generate webhook ID
	webhookID := uuid.New()

	// Generate security credentials
	isPrivate := req.Type == models.WebhookTypePrivate
	securityData, err := s.securitySvc.GenerateWebhookSecurity(isPrivate, req.TenantID, webhookID.String(), req.AppName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate security credentials: %w", err)
	}

	// Create webhook URL
	webhookURL := fmt.Sprintf("http://localhost:8080/api/webhooks/receive/%s", webhookID.String())

	// Create subscription
	subscription := &models.WebhookSubscription{
		ID:              webhookID,
		TenantID:        req.TenantID,
		AppName:         req.AppName,
		TargetURL:       webhookURL,
		SubscribedEvent: req.SubscribedEvent,
		Type:            req.Type,
		SecretToken:     securityData.SecretToken,
		JWTToken:        securityData.JWTToken,
		IsActive:        true,
	}

	// Save to database
	if err := s.repo.CreateSubscription(subscription); err != nil {
		logger.Error("Failed to create webhook subscription",
			zap.Error(err),
			zap.String("tenant_id", req.TenantID),
			zap.String("app_name", req.AppName))
		return nil, fmt.Errorf("failed to create webhook subscription: %w", err)
	}

	logger.Info("Webhook subscription created",
		zap.String("webhook_id", webhookID.String()),
		zap.String("tenant_id", req.TenantID),
		zap.String("app_name", req.AppName),
		zap.String("type", string(req.Type)))

	// Prepare response
	response := &models.GenerateWebhookResponse{
		WebhookURL:  webhookURL,
		SecretToken: securityData.SecretToken,
		Type:        req.Type,
		WebhookID:   webhookID,
	}

	if securityData.JWTToken != nil {
		response.JWTToken = securityData.JWTToken
	}

	return response, nil
}

func (s *webhookService) SubscribeWebhook(req *models.SubscribeWebhookRequest) (*models.GenerateWebhookResponse, error) {
	// Validate webhook type
	if req.Type != models.WebhookTypePublic && req.Type != models.WebhookTypePrivate {
		return nil, fmt.Errorf("invalid webhook type: %s", req.Type)
	}

	// Generate webhook ID
	webhookID := uuid.New()

	// Generate security credentials
	isPrivate := req.Type == models.WebhookTypePrivate
	securityData, err := s.securitySvc.GenerateWebhookSecurity(isPrivate, req.TenantID, webhookID.String(), req.AppName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate security credentials: %w", err)
	}

	// Create subscription
	subscription := &models.WebhookSubscription{
		ID:              webhookID,
		TenantID:        req.TenantID,
		AppName:         req.AppName,
		TargetURL:       req.TargetURL,
		SubscribedEvent: req.SubscribedEvent,
		Type:            req.Type,
		SecretToken:     securityData.SecretToken,
		JWTToken:        securityData.JWTToken,
		IsActive:        true,
	}

	// Save to database
	if err := s.repo.CreateSubscription(subscription); err != nil {
		logger.Error("Failed to create webhook subscription",
			zap.Error(err),
			zap.String("tenant_id", req.TenantID),
			zap.String("app_name", req.AppName))
		return nil, fmt.Errorf("failed to create webhook subscription: %w", err)
	}

	logger.Info("Manual webhook subscription created",
		zap.String("webhook_id", webhookID.String()),
		zap.String("tenant_id", req.TenantID),
		zap.String("target_url", req.TargetURL))

	// Prepare response
	response := &models.GenerateWebhookResponse{
		WebhookURL:  req.TargetURL,
		SecretToken: securityData.SecretToken,
		Type:        req.Type,
		WebhookID:   webhookID,
	}

	if securityData.JWTToken != nil {
		response.JWTToken = securityData.JWTToken
	}

	return response, nil
}

func (s *webhookService) SendEvent(req *models.SendEventRequest) (*models.EventProcessingResult, error) {
	// Find matching subscriptions
	subscriptions, err := s.repo.GetActiveSubscriptionsByTenantAndEvent(req.TenantID, req.Event)
	if err != nil {
		logger.Error("Failed to find webhook subscriptions",
			zap.Error(err),
			zap.String("tenant_id", req.TenantID),
			zap.String("event", req.Event))
		return nil, fmt.Errorf("failed to find webhook subscriptions: %w", err)
	}

	// Create event record
	eventID := uuid.New()
	webhookPayload := &models.WebhookPayload{
		Event:     req.Event,
		Source:    req.Source,
		Timestamp: time.Now().Format(time.RFC3339),
		Payload:   req.Payload,
		EventID:   eventID,
	}

	payloadBytes, err := json.Marshal(webhookPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize webhook payload: %w", err)
	}

	event := &models.WebhookEvent{
		ID:        eventID,
		TenantID:  req.TenantID,
		EventName: req.Event,
		Source:    req.Source,
		Payload:   string(payloadBytes),
		Status:    models.WebhookStatusPending,
	}

	if err := s.repo.CreateEvent(event); err != nil {
		logger.Error("Failed to create webhook event",
			zap.Error(err),
			zap.String("event_id", eventID.String()))
	}

	// Send webhooks
	result := &models.EventProcessingResult{
		EventID:  eventID,
		Webhooks: make([]models.WebhookDeliveryResult, len(subscriptions)),
	}

	for i, subscription := range subscriptions {
		deliveryResult := s.sendWebhookToSubscription(subscription, payloadBytes)
		result.Webhooks[i] = deliveryResult

		if deliveryResult.Success {
			result.TotalSent++
		} else {
			result.TotalFailed++
		}
	}

	// Update event status
	if result.TotalSent > 0 && result.TotalFailed == 0 {
		event.Status = models.WebhookStatusSent
		now := time.Now()
		event.SentAt = &now
	} else if result.TotalFailed > 0 {
		event.Status = models.WebhookStatusFailed
		if result.TotalFailed == len(subscriptions) {
			errMsg := fmt.Sprintf("all %d webhook deliveries failed", result.TotalFailed)
			event.LastError = &errMsg
		}
	}

	event.Attempts = 1
	s.repo.UpdateEvent(event)

	logger.Info("Webhook event processed",
		zap.String("event_id", eventID.String()),
		zap.String("tenant_id", req.TenantID),
		zap.String("event", req.Event),
		zap.Int("total_sent", result.TotalSent),
		zap.Int("total_failed", result.TotalFailed))

	// Execute chains triggered by this event
	if s.chainService != nil {
		ctx := context.Background()

		// Convert payload to map[string]interface{}
		var eventData map[string]interface{}
		if req.Payload != nil {
			if payloadMap, ok := req.Payload.(map[string]interface{}); ok {
				eventData = payloadMap
			} else {
				// Try to convert via JSON marshal/unmarshal
				if payloadBytes, err := json.Marshal(req.Payload); err == nil {
					json.Unmarshal(payloadBytes, &eventData)
				}
			}
		}

		if err := s.chainService.ExecuteChainByEvent(ctx, req.TenantID, req.Event, eventData); err != nil {
			logger.Error("Failed to execute chains for event",
				zap.String("event", req.Event),
				zap.String("tenant_id", req.TenantID),
				zap.Error(err))
			// Don't fail the entire operation if chain execution fails
		}
	}

	return result, nil
}

func (s *webhookService) sendWebhookToSubscription(subscription models.WebhookSubscription, payload []byte) models.WebhookDeliveryResult {
	result := models.WebhookDeliveryResult{
		WebhookID: subscription.ID,
		TargetURL: subscription.TargetURL,
		Success:   false,
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", subscription.TargetURL, bytes.NewBuffer(payload))
	if err != nil {
		errMsg := fmt.Sprintf("failed to create request: %v", err)
		result.Error = &errMsg
		return result
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "github.com/sakibcoolz/loki-suite/2.0")

	// Generate HMAC signature
	signature := s.securitySvc.GenerateHMACSignature(payload, subscription.SecretToken)
	req.Header.Set("X-Shavix-Signature", fmt.Sprintf("sha256=%s", signature))
	req.Header.Set("X-Shavix-Timestamp", time.Now().Format(time.RFC3339))

	// Add JWT token for private webhooks
	if subscription.Type == models.WebhookTypePrivate && subscription.JWTToken != nil {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *subscription.JWTToken))
	}

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		errMsg := fmt.Sprintf("failed to send request: %v", err)
		result.Error = &errMsg
		result.AttemptCount = 1
		return result
	}
	defer resp.Body.Close()

	result.ResponseCode = &resp.StatusCode
	result.AttemptCount = 1

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
		logger.Debug("Webhook delivered successfully",
			zap.String("webhook_id", subscription.ID.String()),
			zap.String("target_url", subscription.TargetURL),
			zap.Int("status_code", resp.StatusCode))
	} else {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("webhook returned status %d: %s", resp.StatusCode, string(bodyBytes))
		result.Error = &errMsg
		logger.Warn("Webhook delivery failed",
			zap.String("webhook_id", subscription.ID.String()),
			zap.String("target_url", subscription.TargetURL),
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(bodyBytes)))
	}

	return result
}

func (s *webhookService) VerifyWebhook(webhookID uuid.UUID, payload []byte, signature, timestamp, authHeader string) error {
	// Find webhook subscription
	subscription, err := s.repo.GetSubscriptionByID(webhookID)
	if err != nil {
		return fmt.Errorf("webhook subscription not found: %w", err)
	}

	if !subscription.IsActive {
		return fmt.Errorf("webhook subscription is inactive")
	}

	// Extract and verify HMAC signature
	sig, err := s.securitySvc.ExtractSignatureFromHeader(signature)
	if err != nil {
		return fmt.Errorf("invalid signature header: %w", err)
	}

	if !s.securitySvc.VerifyHMACSignature(payload, sig, subscription.SecretToken) {
		return fmt.Errorf("HMAC signature verification failed")
	}

	// Verify timestamp
	if !s.securitySvc.ValidateTimestamp(timestamp, 5) { // Default 5 minute tolerance
		return fmt.Errorf("timestamp is outside allowed tolerance")
	}

	// Verify JWT token for private webhooks
	if subscription.Type == models.WebhookTypePrivate {
		if authHeader == "" {
			return fmt.Errorf("authorization header is required for private webhooks")
		}

		token, err := s.securitySvc.ExtractBearerToken(authHeader)
		if err != nil {
			return fmt.Errorf("invalid authorization header: %w", err)
		}

		claims, err := s.securitySvc.VerifyJWTToken(token)
		if err != nil {
			return fmt.Errorf("JWT token verification failed: %w", err)
		}

		// Verify claims match the webhook
		if claims.WebhookID != webhookID.String() || claims.TenantID != subscription.TenantID {
			return fmt.Errorf("JWT token claims do not match webhook")
		}
	}

	logger.Debug("Webhook verification successful",
		zap.String("webhook_id", webhookID.String()),
		zap.String("tenant_id", subscription.TenantID),
		zap.String("type", string(subscription.Type)))

	return nil
}

func (s *webhookService) ListWebhooks(tenantID string, page, limit int) (*models.WebhookListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	webhooks, total, err := s.repo.GetSubscriptionsByTenant(tenantID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch webhooks: %w", err)
	}

	return &models.WebhookListResponse{
		Webhooks: webhooks,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}, nil
}
