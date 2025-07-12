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
// This interface provides methods for managing webhook subscriptions, sending events,
// and handling webhook security verification
type WebhookService interface {
	// GenerateWebhook creates a new webhook subscription and returns the webhook URL
	// Parameters:
	//   - req: Contains webhook configuration including tenant ID, app name, event type, and webhook type
	// Returns:
	//   - GenerateWebhookResponse: Contains the generated webhook URL, security tokens, and webhook ID
	//   - error: If webhook creation fails due to validation or database errors
	GenerateWebhook(req *models.GenerateWebhookRequest) (*models.GenerateWebhookResponse, error)

	// SubscribeWebhook creates a webhook subscription with a custom target URL
	// Parameters:
	//   - req: Contains subscription details including target URL, tenant ID, and event filters
	// Returns:
	//   - GenerateWebhookResponse: Contains security credentials and webhook configuration
	//   - error: If subscription creation fails
	SubscribeWebhook(req *models.SubscribeWebhookRequest) (*models.GenerateWebhookResponse, error)

	// SendEvent broadcasts an event to all matching webhook subscriptions
	// Parameters:
	//   - req: Contains event data, tenant ID, event name, source, and payload
	// Returns:
	//   - EventProcessingResult: Summary of delivery results including success/failure counts
	//   - error: If event processing fails
	SendEvent(req *models.SendEventRequest) (*models.EventProcessingResult, error)

	// VerifyWebhook validates the authenticity and authorization of incoming webhook requests
	// Parameters:
	//   - webhookID: UUID of the webhook subscription
	//   - payload: Raw request body bytes for signature verification
	//   - signature: HMAC signature from request headers
	//   - timestamp: Request timestamp for replay attack prevention
	//   - authHeader: Authorization header containing JWT token (for private webhooks)
	// Returns:
	//   - error: If verification fails due to invalid signature, expired timestamp, or unauthorized access
	VerifyWebhook(webhookID uuid.UUID, payload []byte, signature, timestamp, authHeader string) error

	// ListWebhooks retrieves paginated webhook subscriptions for a tenant
	// Parameters:
	//   - tenantID: Filter webhooks by tenant identifier
	//   - page: Page number for pagination (1-based)
	//   - limit: Maximum number of results per page (1-100, default 10)
	// Returns:
	//   - WebhookListResponse: Contains webhooks array, total count, and pagination info
	//   - error: If database query fails
	ListWebhooks(tenantID string, page, limit int) (*models.WebhookListResponse, error)

	// SetChainService injects the execution chain service dependency
	// This is used to avoid circular dependencies between webhook and chain services
	// Parameters:
	//   - chainService: The execution chain service instance for triggering workflows
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

// NewWebhookService creates a new webhook service instance with required dependencies
// This constructor initializes the service with repository, security service, and configuration
// Parameters:
//   - repo: WebhookRepository for database operations (subscriptions, events)
//   - securitySvc: SecurityService for generating tokens, signatures, and verification
//   - cfg: Application configuration containing webhook and security settings
//
// Returns:
//   - WebhookService: Configured service instance ready for use
//
// Note: The execution chain service is set separately via SetChainService to avoid circular dependencies
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
// This method is called after service initialization to inject the chain service dependency
// Parameters:
//   - chainService: ExecutionChainService instance for triggering workflow executions
//
// Purpose: Enables webhook events to automatically trigger execution chains when configured
func (s *webhookService) SetChainService(chainService ExecutionChainService) {
	s.chainService = chainService
}

// GenerateWebhook creates a new webhook subscription and generates a unique webhook URL
// This method handles the complete webhook creation flow including security credential generation
// Parameters:
//   - req: GenerateWebhookRequest containing tenant ID, app name, subscribed event, and webhook type
//
// Returns:
//   - GenerateWebhookResponse: Contains webhook URL, security tokens, and webhook ID
//   - error: If validation fails, security generation fails, or database operation fails
//
// Process:
//  1. Validates webhook type (public/private)
//  2. Generates unique webhook ID and security credentials
//  3. Creates subscription record in database
//  4. Returns webhook URL and security information
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

// SubscribeWebhook creates a webhook subscription with a custom target URL
// This method allows external services to register their own endpoints for webhook delivery
// Parameters:
//   - req: SubscribeWebhookRequest containing target URL, tenant ID, app name, event filter, and type
//
// Returns:
//   - GenerateWebhookResponse: Contains security credentials and webhook configuration
//   - error: If validation fails, security generation fails, or database operation fails
//
// Process:
//  1. Validates webhook type and target URL
//  2. Generates webhook ID and security credentials
//  3. Creates subscription with provided target URL
//  4. Returns security information for the subscriber
//
// Use case: When external services want to receive webhooks at their own endpoints
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

// SendEvent broadcasts an event to all matching webhook subscriptions and triggers execution chains
// This is the core event delivery method that handles the complete webhook notification flow
// Parameters:
//   - req: SendEventRequest containing tenant ID, event name, source, and payload data
//
// Returns:
//   - EventProcessingResult: Summary containing event ID, delivery results, and success/failure counts
//   - error: If event creation fails or critical processing errors occur
//
// Process:
//  1. Finds all active subscriptions matching tenant and event
//  2. Creates event record in database for tracking
//  3. Delivers webhook to each subscription with proper security headers
//  4. Updates event status based on delivery results
//  5. Triggers any execution chains configured for this event
//
// Note: Chain execution failures don't fail the entire operation
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

// sendWebhookToSubscription delivers a webhook payload to a single subscription endpoint
// This is an internal helper method that handles the HTTP delivery and security headers
// Parameters:
//   - subscription: WebhookSubscription containing target URL and security credentials
//   - payload: JSON-encoded webhook payload to be delivered
//
// Returns:
//   - WebhookDeliveryResult: Contains delivery status, response code, error details, and attempt count
//
// Process:
//  1. Creates HTTP POST request to target URL
//  2. Adds security headers (Content-Type, User-Agent, HMAC signature, timestamp)
//  3. Adds JWT authorization for private webhooks
//  4. Sends request and evaluates response status
//  5. Logs delivery success/failure with details
//
// Security: Includes HMAC signature verification and JWT tokens for private webhooks
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

// VerifyWebhook validates the authenticity and authorization of incoming webhook requests
// This method provides comprehensive security validation for webhook endpoints
// Parameters:
//   - webhookID: UUID identifying the webhook subscription
//   - payload: Raw request body bytes used for HMAC signature verification
//   - signature: HMAC signature from X-Shavix-Signature header
//   - timestamp: Request timestamp from X-Shavix-Timestamp header
//   - authHeader: Authorization header containing JWT token (required for private webhooks)
//
// Returns:
//   - error: nil if verification succeeds, descriptive error if validation fails
//
// Security Checks:
//  1. Webhook subscription exists and is active
//  2. HMAC signature matches payload and secret token
//  3. Timestamp is within acceptable tolerance (prevents replay attacks)
//  4. JWT token is valid and claims match webhook (for private webhooks only)
//
// Use case: Called by webhook receive endpoints to ensure request authenticity
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

// ListWebhooks retrieves paginated webhook subscriptions for a specific tenant
// This method provides filtered and paginated access to webhook subscriptions
// Parameters:
//   - tenantID: Tenant identifier to filter subscriptions (required)
//   - page: Page number for pagination, 1-based (minimum 1, defaults to 1)
//   - limit: Maximum results per page (range 1-100, defaults to 10)
//
// Returns:
//   - WebhookListResponse: Contains webhooks array, total count, current page, and limit
//   - error: If database query fails or parameters are invalid
//
// Process:
//  1. Validates and normalizes pagination parameters
//  2. Calculates offset for database query
//  3. Retrieves subscriptions with total count
//  4. Returns structured response with pagination metadata
//
// Use case: Management dashboards, webhook administration, and subscription overview
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
