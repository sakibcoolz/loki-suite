package service_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/sakibcoolz/loki-suite/internal/models"
	"github.com/sakibcoolz/loki-suite/internal/service"
	"github.com/sakibcoolz/loki-suite/mocks"
	"github.com/sakibcoolz/zcornor/pkg/config"
	"github.com/sakibcoolz/zcornor/pkg/security"
)

// WebhookServiceTestSuite is the test suite for webhook service
type WebhookServiceTestSuite struct {
	suite.Suite
	service      service.WebhookService
	mockRepo     *mocks.MockWebhookRepository
	mockChainSvc *mocks.MockExecutionChainService
	securitySvc  *security.SecurityService
	config       *config.Config
	testServer   *httptest.Server
}

// SetupTest initializes test dependencies before each test
func (suite *WebhookServiceTestSuite) SetupTest() {
	// Create mocks
	suite.mockRepo = mocks.NewMockWebhookRepository(suite.T())
	suite.mockChainSvc = mocks.NewMockExecutionChainService(suite.T())

	// Create test configuration
	suite.config = &config.Config{}

	// Create security service
	suite.securitySvc = security.NewSecurityService("test-jwt-secret", 3600, 300)

	// Create webhook service
	webhookService := service.NewWebhookService(
		suite.mockRepo,
		suite.securitySvc,
		suite.config,
	)

	// Set chain service
	webhookService.SetChainService(suite.mockChainSvc)

	// Create test HTTP server for webhook delivery testing
	suite.testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "success"}`))
		case "/failure":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal server error"}`))
		case "/client-error":
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "bad request"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	suite.service = webhookService
}

// TearDownTest cleans up after each test
func (suite *WebhookServiceTestSuite) TearDownTest() {
	suite.testServer.Close()
}

// TestGenerateWebhook_Success tests successful webhook generation
func (suite *WebhookServiceTestSuite) TestGenerateWebhook_Success() {
	// Arrange
	req := &models.GenerateWebhookRequest{
		TenantID:        "tenant-123",
		AppName:         "test-app",
		SubscribedEvent: "user.created",
		Type:            models.WebhookTypePublic,
		QueryParams:     map[string]string{"source": "test"},
		Payload:         map[string]interface{}{"extra": "data"},
		RetryPolicy: &models.RetryPolicy{
			MaxRetries:        3,
			RetryDelaySeconds: 5,
		},
	}

	// Mock repository call
	suite.mockRepo.EXPECT().
		CreateSubscription(mock.MatchedBy(func(sub *models.WebhookSubscription) bool {
			return sub.TenantID == req.TenantID &&
				sub.AppName == req.AppName &&
				sub.SubscribedEvent == req.SubscribedEvent &&
				sub.Type == req.Type &&
				sub.MaxRetries == 3 &&
				sub.RetryDelaySeconds == 5 &&
				len(sub.QueryParams) == 1 &&
				sub.Payload != ""
		})).
		Return(nil).
		Once()

	// Act
	result, err := suite.service.GenerateWebhook(req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.NotEmpty(suite.T(), result.WebhookURL)
	assert.NotEmpty(suite.T(), result.SecretToken)
	assert.Equal(suite.T(), models.WebhookTypePublic, result.Type)
	assert.NotEqual(suite.T(), uuid.Nil, result.WebhookID)
	assert.Equal(suite.T(), req.QueryParams, result.QueryParams)
	assert.Equal(suite.T(), req.Payload, result.Payload)
	assert.Equal(suite.T(), req.RetryPolicy, result.RetryPolicy)
	assert.Nil(suite.T(), result.JWTToken) // Public webhook shouldn't have JWT token
}

// TestGenerateWebhook_PrivateWebhook tests private webhook generation
func (suite *WebhookServiceTestSuite) TestGenerateWebhook_PrivateWebhook() {
	// Arrange
	req := &models.GenerateWebhookRequest{
		TenantID:        "tenant-123",
		AppName:         "test-app",
		SubscribedEvent: "user.created",
		Type:            models.WebhookTypePrivate,
	}

	// Mock repository call
	suite.mockRepo.EXPECT().
		CreateSubscription(mock.MatchedBy(func(sub *models.WebhookSubscription) bool {
			return sub.Type == models.WebhookTypePrivate && sub.JWTToken != nil
		})).
		Return(nil).
		Once()

	// Act
	result, err := suite.service.GenerateWebhook(req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), models.WebhookTypePrivate, result.Type)
	assert.NotNil(suite.T(), result.JWTToken) // Private webhook should have JWT token
}

// TestGenerateWebhook_RepositoryError tests repository failure
func (suite *WebhookServiceTestSuite) TestGenerateWebhook_RepositoryError() {
	// Arrange
	req := &models.GenerateWebhookRequest{
		TenantID:        "tenant-123",
		AppName:         "test-app",
		SubscribedEvent: "user.created",
		Type:            models.WebhookTypePublic,
	}

	// Mock repository error
	suite.mockRepo.EXPECT().
		CreateSubscription(mock.AnythingOfType("*models.WebhookSubscription")).
		Return(fmt.Errorf("database error")).
		Once()

	// Act
	result, err := suite.service.GenerateWebhook(req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to create webhook subscription")
}

// TestSubscribeWebhook_Success tests successful webhook subscription
func (suite *WebhookServiceTestSuite) TestSubscribeWebhook_Success() {
	// Arrange
	req := &models.SubscribeWebhookRequest{
		TenantID:        "tenant-123",
		AppName:         "external-app",
		TargetURL:       "https://example.com/webhook",
		SubscribedEvent: "order.completed",
		Type:            models.WebhookTypePublic,
		Headers:         map[string]string{"Authorization": "Bearer token"},
		QueryParams:     map[string]string{"version": "v1"},
		Description:     stringPtr("Test webhook subscription"),
		IsActive:        boolPtr(true),
		IsPublic:        true,
		RetryPolicy: &models.RetryPolicy{
			MaxRetries:        5,
			RetryDelaySeconds: 10,
		},
	}

	// Mock repository call
	suite.mockRepo.EXPECT().
		CreateSubscription(mock.MatchedBy(func(sub *models.WebhookSubscription) bool {
			return sub.TenantID == req.TenantID &&
				sub.AppName == req.AppName &&
				sub.TargetURL == req.TargetURL &&
				sub.SubscribedEvent == req.SubscribedEvent &&
				sub.Type == req.Type &&
				len(sub.Headers) == 1 &&
				len(sub.QueryParams) == 1 &&
				sub.Description != nil &&
				*sub.Description == *req.Description &&
				sub.IsActive == *req.IsActive &&
				sub.MaxRetries == 5 &&
				sub.RetryDelaySeconds == 10
		})).
		Return(nil).
		Once()

	// Act
	result, err := suite.service.SubscribeWebhook(req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), req.TargetURL, result.WebhookURL)
	assert.NotEmpty(suite.T(), result.SecretToken)
	assert.Equal(suite.T(), req.Type, result.Type)
	assert.Equal(suite.T(), req.QueryParams, result.QueryParams)
	assert.Equal(suite.T(), req.RetryPolicy, result.RetryPolicy)
}

// TestSendEvent_Success tests successful event sending
func (suite *WebhookServiceTestSuite) TestSendEvent_Success() {
	// Arrange
	req := &models.SendEventRequest{
		TenantID: "tenant-123",
		Event:    "user.created",
		Source:   "user-service",
		Payload:  map[string]interface{}{"user_id": "123", "email": "test@example.com"},
	}

	subscriptions := []models.WebhookSubscription{
		{
			ID:                uuid.New(),
			TenantID:          req.TenantID,
			TargetURL:         suite.testServer.URL + "/success",
			SubscribedEvent:   req.Event,
			Type:              models.WebhookTypePublic,
			SecretToken:       "test-secret",
			MaxRetries:        3,
			RetryDelaySeconds: 1,
			IsActive:          true,
		},
	}

	// Mock repository calls
	suite.mockRepo.EXPECT().
		GetActiveSubscriptionsByTenantAndEvent(req.TenantID, req.Event).
		Return(subscriptions, nil).
		Once()

	suite.mockRepo.EXPECT().
		CreateEvent(mock.MatchedBy(func(event *models.WebhookEvent) bool {
			return event.TenantID == req.TenantID &&
				event.EventName == req.Event &&
				event.Source == req.Source
		})).
		Return(nil).
		Once()

	suite.mockRepo.EXPECT().
		UpdateEvent(mock.MatchedBy(func(event *models.WebhookEvent) bool {
			return event.Status == models.WebhookStatusSent &&
				event.SentAt != nil
		})).
		Return(nil).
		Once()

	// Mock chain service call
	suite.mockChainSvc.EXPECT().
		ExecuteChainByEvent(mock.Anything, req.TenantID, req.Event, mock.Anything).
		Return(nil).
		Once()

	// Act
	result, err := suite.service.SendEvent(req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), 1, result.TotalSent)
	assert.Equal(suite.T(), 0, result.TotalFailed)
	assert.Len(suite.T(), result.Webhooks, 1)
	assert.True(suite.T(), result.Webhooks[0].Success)
	assert.Equal(suite.T(), http.StatusOK, *result.Webhooks[0].ResponseCode)
}

// TestSendEvent_WithRetries tests event sending with retry logic
func (suite *WebhookServiceTestSuite) TestSendEvent_WithRetries() {
	// Arrange
	req := &models.SendEventRequest{
		TenantID: "tenant-123",
		Event:    "user.created",
		Source:   "user-service",
		Payload:  map[string]interface{}{"user_id": "123"},
	}

	subscriptions := []models.WebhookSubscription{
		{
			ID:                uuid.New(),
			TenantID:          req.TenantID,
			TargetURL:         suite.testServer.URL + "/failure",
			SubscribedEvent:   req.Event,
			Type:              models.WebhookTypePublic,
			SecretToken:       "test-secret",
			MaxRetries:        3,
			RetryDelaySeconds: 1,
			IsActive:          true,
		},
	}

	// Mock repository calls
	suite.mockRepo.EXPECT().
		GetActiveSubscriptionsByTenantAndEvent(req.TenantID, req.Event).
		Return(subscriptions, nil).
		Once()

	suite.mockRepo.EXPECT().
		CreateEvent(mock.AnythingOfType("*models.WebhookEvent")).
		Return(nil).
		Once()

	suite.mockRepo.EXPECT().
		UpdateEvent(mock.MatchedBy(func(event *models.WebhookEvent) bool {
			return event.Status == models.WebhookStatusFailed
		})).
		Return(nil).
		Once()

	// Mock chain service call
	suite.mockChainSvc.EXPECT().
		ExecuteChainByEvent(mock.Anything, req.TenantID, req.Event, mock.Anything).
		Return(nil).
		Once()

	// Act
	result, err := suite.service.SendEvent(req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), 0, result.TotalSent)
	assert.Equal(suite.T(), 1, result.TotalFailed)
	assert.Len(suite.T(), result.Webhooks, 1)
	assert.False(suite.T(), result.Webhooks[0].Success)
	assert.Equal(suite.T(), 3, result.Webhooks[0].AttemptCount) // Should retry 3 times
	assert.Equal(suite.T(), http.StatusInternalServerError, *result.Webhooks[0].ResponseCode)
}

// TestSendEvent_ClientErrorNoRetry tests that client errors don't trigger retries
func (suite *WebhookServiceTestSuite) TestSendEvent_ClientErrorNoRetry() {
	// Arrange
	req := &models.SendEventRequest{
		TenantID: "tenant-123",
		Event:    "user.created",
		Source:   "user-service",
		Payload:  map[string]interface{}{"user_id": "123"},
	}

	subscriptions := []models.WebhookSubscription{
		{
			ID:                uuid.New(),
			TenantID:          req.TenantID,
			TargetURL:         suite.testServer.URL + "/client-error",
			SubscribedEvent:   req.Event,
			Type:              models.WebhookTypePublic,
			SecretToken:       "test-secret",
			MaxRetries:        3,
			RetryDelaySeconds: 1,
			IsActive:          true,
		},
	}

	// Mock repository calls
	suite.mockRepo.EXPECT().
		GetActiveSubscriptionsByTenantAndEvent(req.TenantID, req.Event).
		Return(subscriptions, nil).
		Once()

	suite.mockRepo.EXPECT().
		CreateEvent(mock.AnythingOfType("*models.WebhookEvent")).
		Return(nil).
		Once()

	suite.mockRepo.EXPECT().
		UpdateEvent(mock.MatchedBy(func(event *models.WebhookEvent) bool {
			return event.Status == models.WebhookStatusFailed
		})).
		Return(nil).
		Once()

	// Mock chain service call
	suite.mockChainSvc.EXPECT().
		ExecuteChainByEvent(mock.Anything, req.TenantID, req.Event, mock.Anything).
		Return(nil).
		Once()

	// Act
	result, err := suite.service.SendEvent(req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), 0, result.TotalSent)
	assert.Equal(suite.T(), 1, result.TotalFailed)
	assert.Len(suite.T(), result.Webhooks, 1)
	assert.False(suite.T(), result.Webhooks[0].Success)
	assert.Equal(suite.T(), 1, result.Webhooks[0].AttemptCount) // Should NOT retry on client error
	assert.Equal(suite.T(), http.StatusBadRequest, *result.Webhooks[0].ResponseCode)
}

// TestSendEvent_WithPayloadMerging tests payload merging functionality
func (suite *WebhookServiceTestSuite) TestSendEvent_WithPayloadMerging() {
	// Arrange
	req := &models.SendEventRequest{
		TenantID: "tenant-123",
		Event:    "user.created",
		Source:   "user-service",
		Payload:  map[string]interface{}{"user_id": "123", "name": "John"},
	}

	// Subscription with additional payload
	subscriptionPayload := map[string]interface{}{"source": "subscription", "priority": "high"}
	payloadJSON, _ := json.Marshal(subscriptionPayload)

	subscriptions := []models.WebhookSubscription{
		{
			ID:                uuid.New(),
			TenantID:          req.TenantID,
			TargetURL:         suite.testServer.URL + "/success",
			SubscribedEvent:   req.Event,
			Type:              models.WebhookTypePublic,
			SecretToken:       "test-secret",
			MaxRetries:        1,
			RetryDelaySeconds: 1,
			IsActive:          true,
			Payload:           string(payloadJSON),
		},
	}

	// Mock repository calls
	suite.mockRepo.EXPECT().
		GetActiveSubscriptionsByTenantAndEvent(req.TenantID, req.Event).
		Return(subscriptions, nil).
		Once()

	suite.mockRepo.EXPECT().
		CreateEvent(mock.AnythingOfType("*models.WebhookEvent")).
		Return(nil).
		Once()

	suite.mockRepo.EXPECT().
		UpdateEvent(mock.MatchedBy(func(event *models.WebhookEvent) bool {
			return event.Status == models.WebhookStatusSent
		})).
		Return(nil).
		Once()

	// Mock chain service call
	suite.mockChainSvc.EXPECT().
		ExecuteChainByEvent(mock.Anything, req.TenantID, req.Event, mock.Anything).
		Return(nil).
		Once()

	// Act
	result, err := suite.service.SendEvent(req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), 1, result.TotalSent)
	assert.Equal(suite.T(), 0, result.TotalFailed)
	assert.True(suite.T(), result.Webhooks[0].Success)
}

// TestSendEvent_NoSubscriptions tests sending event with no matching subscriptions
func (suite *WebhookServiceTestSuite) TestSendEvent_NoSubscriptions() {
	// Arrange
	req := &models.SendEventRequest{
		TenantID: "tenant-123",
		Event:    "user.created",
		Source:   "user-service",
		Payload:  map[string]interface{}{"user_id": "123"},
	}

	// Mock repository calls - return empty subscriptions
	suite.mockRepo.EXPECT().
		GetActiveSubscriptionsByTenantAndEvent(req.TenantID, req.Event).
		Return([]models.WebhookSubscription{}, nil).
		Once()

	// Mock CreateEvent call since the service always creates an event record
	suite.mockRepo.EXPECT().
		CreateEvent(mock.AnythingOfType("*models.WebhookEvent")).
		Return(nil).
		Once()

	// Mock UpdateEvent call since the service always updates the event status
	suite.mockRepo.EXPECT().
		UpdateEvent(mock.AnythingOfType("*models.WebhookEvent")).
		Return(nil).
		Once()

	// Mock chain service call
	suite.mockChainSvc.EXPECT().
		ExecuteChainByEvent(mock.Anything, req.TenantID, req.Event, mock.Anything).
		Return(nil).
		Once()

	// Act
	result, err := suite.service.SendEvent(req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), 0, result.TotalSent)
	assert.Equal(suite.T(), 0, result.TotalFailed)
	assert.Len(suite.T(), result.Webhooks, 0)
}

// TestVerifyWebhook_Success tests successful webhook verification
func (suite *WebhookServiceTestSuite) TestVerifyWebhook_Success() {
	// Arrange
	webhookID := uuid.New()
	payload := []byte(`{"test": "data"}`)
	secretToken := "test-secret"
	timestamp := time.Now().Format(time.RFC3339)

	subscription := &models.WebhookSubscription{
		ID:          webhookID,
		Type:        models.WebhookTypePublic,
		SecretToken: secretToken,
		IsActive:    true,
	}

	signature := suite.securitySvc.GenerateHMACSignature(payload, secretToken)

	// Mock repository call
	suite.mockRepo.EXPECT().
		GetSubscriptionByID(webhookID).
		Return(subscription, nil).
		Once()

	// Act
	err := suite.service.VerifyWebhook(webhookID, payload, fmt.Sprintf("sha256=%s", signature), timestamp, "")

	// Assert
	assert.NoError(suite.T(), err)
}

// TestVerifyWebhook_InvalidSignature tests verification with invalid signature
func (suite *WebhookServiceTestSuite) TestVerifyWebhook_InvalidSignature() {
	// Arrange
	webhookID := uuid.New()
	payload := []byte(`{"test": "data"}`)
	secretToken := "test-secret"
	timestamp := time.Now().Format(time.RFC3339)
	invalidSignature := "invalid-signature"

	subscription := &models.WebhookSubscription{
		ID:          webhookID,
		Type:        models.WebhookTypePublic,
		SecretToken: secretToken,
		IsActive:    true,
	}

	// Mock repository call
	suite.mockRepo.EXPECT().
		GetSubscriptionByID(webhookID).
		Return(subscription, nil).
		Once()

	// Act
	err := suite.service.VerifyWebhook(webhookID, payload, invalidSignature, timestamp, "")

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid signature format")
}

// TestVerifyWebhook_SubscriptionNotFound tests verification when subscription is not found
func (suite *WebhookServiceTestSuite) TestVerifyWebhook_SubscriptionNotFound() {
	// Arrange
	webhookID := uuid.New()
	payload := []byte(`{"test": "data"}`)
	timestamp := time.Now().Format(time.RFC3339)
	signature := "sha256=test"

	// Mock repository call - return error
	suite.mockRepo.EXPECT().
		GetSubscriptionByID(webhookID).
		Return(nil, fmt.Errorf("subscription not found")).
		Once()

	// Act
	err := suite.service.VerifyWebhook(webhookID, payload, signature, timestamp, "")

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "subscription not found")
}

// TestListWebhooks_Success tests successful webhook listing
func (suite *WebhookServiceTestSuite) TestListWebhooks_Success() {
	// Arrange
	tenantID := "tenant-123"
	page := 1
	limit := 10

	subscriptions := []models.WebhookSubscription{
		{
			ID:              uuid.New(),
			TenantID:        tenantID,
			AppName:         "app1",
			SubscribedEvent: "user.created",
			Type:            models.WebhookTypePublic,
			IsActive:        true,
		},
		{
			ID:              uuid.New(),
			TenantID:        tenantID,
			AppName:         "app2",
			SubscribedEvent: "order.completed",
			Type:            models.WebhookTypePrivate,
			IsActive:        true,
		},
	}

	// Mock repository call
	suite.mockRepo.EXPECT().
		GetSubscriptionsByTenant(tenantID, 0, limit).
		Return(subscriptions, int64(2), nil).
		Once()

	// Act
	result, err := suite.service.ListWebhooks(tenantID, page, limit)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Webhooks, 2)
	assert.Equal(suite.T(), int64(2), result.Total)
	assert.Equal(suite.T(), page, result.Page)
	assert.Equal(suite.T(), limit, result.Limit)
}

// TestListWebhooks_RepositoryError tests listing when repository returns error
func (suite *WebhookServiceTestSuite) TestListWebhooks_RepositoryError() {
	// Arrange
	tenantID := "tenant-123"
	page := 1
	limit := 10

	// Mock repository call - return error
	suite.mockRepo.EXPECT().
		GetSubscriptionsByTenant(tenantID, 0, limit).
		Return(nil, int64(0), fmt.Errorf("database error")).
		Once()

	// Act
	result, err := suite.service.ListWebhooks(tenantID, page, limit)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to fetch webhooks")
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

// TestWebhookServiceTestSuite runs the test suite
func TestWebhookServiceTestSuite(t *testing.T) {
	suite.Run(t, new(WebhookServiceTestSuite))
}
