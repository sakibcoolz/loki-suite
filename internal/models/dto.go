package models

import (
	"time"

	"github.com/google/uuid"
)

// Request DTOs - Data Transfer Objects for incoming API requests

// RetryPolicy defines the retry behavior for webhook deliveries
type RetryPolicy struct {
	MaxRetries        int `json:"max_retries,omitempty"`
	RetryDelaySeconds int `json:"retry_delay_seconds,omitempty"`
}

// GenerateWebhookRequest represents the request to generate a new webhook subscription
// Used when creating a new webhook endpoint that will receive event notifications
type GenerateWebhookRequest struct {
	// TenantID identifies the tenant/organization requesting the webhook
	// Required for multi-tenancy isolation and access control
	TenantID string `json:"tenant_id" binding:"required"`

	// AppName identifies the application or service creating the webhook
	// Used for organizing and filtering webhooks by source application
	AppName string `json:"app_name" binding:"required"`

	// SubscribedEvent specifies which event type this webhook should receive
	// Acts as a filter to determine which events will trigger this webhook
	SubscribedEvent string `json:"subscribed_event" binding:"required"`

	// Type determines the security model (public with HMAC or private with HMAC+JWT)
	// Affects authentication requirements and security credentials generated
	Type WebhookType `json:"type" binding:"required"`

	// retry policy defines how failed deliveries should be retried
	// Allows subscribers to specify retry behavior for failed webhook deliveries
	RetryPolicy *RetryPolicy `json:"retry_policy,omitempty"`

	// QueryParams is an optional map of query parameters to include in webhook requests
	// Allows subscribers to specify additional parameters for the webhook URL
	QueryParams map[string]string `json:"query_params,omitempty"`

	// Payload is an optional field for additional data to be sent with the webhook
	Payload interface{} `json:"payload,omitempty"`
}

// SubscribeWebhookRequest represents the manual subscription request
// Used when an external service wants to register their own endpoint for webhooks
type SubscribeWebhookRequest struct {
	// TenantID identifies the tenant/organization creating the subscription
	// Required for multi-tenancy isolation and event filtering
	TenantID string `json:"tenant_id" binding:"required"`

	// AppName identifies the application or service creating the subscription
	// Used for organizing and managing webhook subscriptions
	AppName string `json:"app_name" binding:"required"`

	// TargetURL is the HTTP endpoint where webhook payloads will be delivered
	// Must be a valid, accessible HTTP/HTTPS URL controlled by the subscriber
	TargetURL string `json:"target_url" binding:"required"`

	// SubscribedEvent specifies which event type should trigger webhook delivery
	// Filters events to only those matching this subscription
	SubscribedEvent string `json:"subscribed_event" binding:"required"`

	// Type determines the security model and authentication requirements
	// Affects what security credentials are generated and required for verification
	Type WebhookType `json:"type" binding:"required"`

	// SecretToken is the HMAC secret used for verifying webhook authenticity
	// Required for public webhooks to generate X-Shavix-Signature headers
	SecretToken *string `json:"secret_token,omitempty"`

	// JWTToken is the JWT used for private webhook authentication
	// Only required for private webhooks, used in Authorization headers
	JWTToken *string `json:"jwt_token,omitempty"`

	// Description is an optional human-readable description of this subscription
	// Provides additional context about the purpose of this webhook
	Description *string `json:"description,omitempty"`

	// IsActive indicates whether this subscription is currently active
	// If false, the webhook will not receive events until reactivated
	IsActive *bool `json:"is_active,omitempty"`

	// Headers is an optional map of custom headers to include in webhook requests
	// Allows subscribers to specify additional metadata or authentication headers
	Headers map[string]string `json:"headers,omitempty"`

	// RetryPolicy defines how failed deliveries should be retried
	// Allows subscribers to specify retry behavior for failed webhook deliveries
	RetryPolicy *RetryPolicy `json:"retry_policy,omitempty"`

	// QueryParams is an optional map of query parameters to include in webhook requests
	// Allows subscribers to specify additional parameters for the webhook URL
	QueryParams map[string]string `json:"query_params,omitempty"`

	// IsPublic indicates whether this subscription is public or private
	// Public subscriptions use HMAC for verification, private use HMAC+JWT
	IsPublic bool `json:"is_public" binding:"required"`
}

// SendEventRequest represents the request to send a webhook event
// Used to broadcast events to all matching webhook subscriptions
type SendEventRequest struct {
	// TenantID identifies the tenant/organization generating the event
	// Used for isolating events and finding matching subscriptions
	TenantID string `json:"tenant_id" binding:"required"`

	// Event specifies the type of event being sent
	// Must match the SubscribedEvent in webhook subscriptions to trigger delivery
	Event string `json:"event" binding:"required"`

	// Source identifies the system or service that generated this event
	// Provides context about the event origin for subscribers
	Source string `json:"source" binding:"required"`

	// Payload contains the event data to be delivered to webhook endpoints
	// Can be any JSON-serializable data structure
	Payload interface{} `json:"payload" binding:"required"`
}

// Response DTOs - Data Transfer Objects for API responses

// GenerateWebhookResponse represents the response after generating a webhook
// Contains all information needed for the client to use the new webhook
type GenerateWebhookResponse struct {
	// WebhookURL is the endpoint where the webhook will receive events
	// This URL should be registered with external services for event delivery
	WebhookURL string `json:"webhook_url"`

	// SecretToken is the HMAC secret for verifying webhook authenticity
	// Used to generate and verify X-Shavix-Signature headers
	SecretToken string `json:"secret_token"`

	// JWTToken contains the JWT for private webhook authentication
	// Only present for private webhooks, used in Authorization headers
	JWTToken *string `json:"jwt_token,omitempty"`

	// Type indicates the security model of this webhook
	// Determines which authentication methods are required
	Type WebhookType `json:"type"`

	// WebhookID is the unique identifier for this webhook subscription
	// Used for management operations and webhook identification
	WebhookID uuid.UUID `json:"webhook_id"`

	// Payload is an optional field for additional data to be sent with the webhook
	Payload interface{} `json:"payload,omitempty"`

	// QueryParams is an optional map of query parameters included in webhook requests
	QueryParams map[string]string `json:"query_params,omitempty"`

	// RetryPolicy defines how failed deliveries should be retried
	// Allows subscribers to specify retry behavior for failed webhook deliveries
	RetryPolicy *RetryPolicy `json:"retry_policy,omitempty"`
}

// WebhookListResponse represents the response for listing webhooks
// Provides paginated results with metadata for webhook management interfaces
type WebhookListResponse struct {
	// Webhooks contains the array of webhook subscriptions for the current page
	// Each entry includes configuration and status information
	Webhooks []WebhookSubscription `json:"webhooks"`

	// Total is the total number of webhooks available (across all pages)
	// Used for pagination controls and result counting
	Total int64 `json:"total"`

	// Page is the current page number (1-based)
	// Indicates which page of results is being returned
	Page int `json:"page"`

	// Limit is the maximum number of results per page
	// Indicates the page size used for this request
	Limit int `json:"limit"`
}

// Webhook payload sent to external endpoints

// WebhookPayload represents the payload sent to webhook endpoints
// This is the standardized format delivered to all webhook subscribers
type WebhookPayload struct {
	// Event specifies the type of event that occurred
	// Matches the event name used in webhook subscriptions
	Event string `json:"event"`

	// Source identifies the system or service that generated this event
	// Provides context about the event origin for processing
	Source string `json:"source"`

	// Timestamp is the RFC3339-formatted time when the event occurred
	// Used for ordering events and detecting replay attacks
	Timestamp string `json:"timestamp"`

	// Payload contains the actual event data
	// Structure varies based on the event type and source system
	Payload interface{} `json:"payload"`

	// EventID is the unique identifier for this specific event
	// Used for deduplication and event tracking across systems
	EventID uuid.UUID `json:"event_id"`
}

// Common response types

// ErrorResponse represents an error response
// Standardized error format for consistent API error handling
type ErrorResponse struct {
	// Error is a machine-readable error code or identifier
	// Used for programmatic error handling and categorization
	Error string `json:"error"`

	// Message is a human-readable description of the error
	// Provides additional context and details for debugging
	Message string `json:"message,omitempty"`

	// Code is the HTTP status code associated with this error
	// Redundant with HTTP response code but useful for client-side handling
	Code int `json:"code,omitempty"`
}

// SuccessResponse represents a success response
// Generic success format for operations that don't return specific data
type SuccessResponse struct {
	// Message is a human-readable success confirmation
	// Provides feedback about the completed operation
	Message string `json:"message"`

	// Data contains optional additional information about the success
	// Structure varies based on the operation performed
	Data interface{} `json:"data,omitempty"`
}

// HealthResponse represents health check response
// Provides system status information for monitoring and load balancing
type HealthResponse struct {
	// Status indicates the overall health of the service
	// Common values: "healthy", "degraded", "unhealthy"
	Status string `json:"status"`

	// Service is the name identifier of this service
	// Used for identifying the service in multi-service environments
	Service string `json:"service"`

	// Version is the current version of the service
	// Useful for deployment tracking and compatibility checks
	Version string `json:"version"`

	// Timestamp is the RFC3339-formatted time when the health check was performed
	// Indicates the freshness of the health status
	Timestamp string `json:"timestamp"`
}

// EventProcessingResult represents the result of event processing
type EventProcessingResult struct {
	EventID     uuid.UUID               `json:"event_id"`
	TotalSent   int                     `json:"total_sent"`
	TotalFailed int                     `json:"total_failed"`
	Webhooks    []WebhookDeliveryResult `json:"webhooks"`
}

// WebhookDeliveryResult represents the result of a single webhook delivery
type WebhookDeliveryResult struct {
	WebhookID    uuid.UUID `json:"webhook_id"`
	TargetURL    string    `json:"target_url"`
	Success      bool      `json:"success"`
	ResponseCode *int      `json:"response_code,omitempty"`
	Error        *string   `json:"error,omitempty"`
	AttemptCount int       `json:"attempt_count"`
}

// ===== Execution Chain DTOs =====

// CreateExecutionChainRequest represents the request to create an execution chain
type CreateExecutionChainRequest struct {
	TenantID     string                     `json:"tenant_id" binding:"required"`
	Name         string                     `json:"name" binding:"required"`
	Description  string                     `json:"description"`
	TriggerEvent string                     `json:"trigger_event" binding:"required"`
	Steps        []CreateExecutionChainStep `json:"steps" binding:"required,min=1"`
}

// CreateExecutionChainStep represents a step in the chain creation request
type CreateExecutionChainStep struct {
	WebhookID       uuid.UUID              `json:"webhook_id" binding:"required"`
	Name            string                 `json:"name" binding:"required"`
	Description     string                 `json:"description"`
	RequestParams   map[string]interface{} `json:"request_params"`
	OnSuccessAction string                 `json:"on_success_action,omitempty"` // continue, stop, pause
	OnFailureAction string                 `json:"on_failure_action,omitempty"` // continue, stop, retry
	MaxRetries      int                    `json:"max_retries,omitempty"`
	DelaySeconds    int                    `json:"delay_seconds,omitempty"`
}

// CreateExecutionChainResponse represents the response for chain creation
type CreateExecutionChainResponse struct {
	ChainID      uuid.UUID `json:"chain_id"`
	Name         string    `json:"name"`
	TriggerEvent string    `json:"trigger_event"`
	StepsCount   int       `json:"steps_count"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// ExecuteChainRequest represents the request to manually execute a chain
type ExecuteChainRequest struct {
	ChainID     uuid.UUID              `json:"chain_id" binding:"required"`
	TriggerData map[string]interface{} `json:"trigger_data,omitempty"`
}

// ExecuteChainResponse represents the response for chain execution
type ExecuteChainResponse struct {
	RunID      uuid.UUID `json:"run_id"`
	ChainID    uuid.UUID `json:"chain_id"`
	Status     string    `json:"status"`
	TotalSteps int       `json:"total_steps"`
	StartedAt  time.Time `json:"started_at"`
}

// ExecutionChainListResponse represents the response for listing chains
type ExecutionChainListResponse struct {
	Chains []ExecutionChain `json:"chains"`
	Total  int64            `json:"total"`
	Page   int              `json:"page"`
	Limit  int              `json:"limit"`
}

// ExecutionChainRunsResponse represents the response for listing chain runs
type ExecutionChainRunsResponse struct {
	Runs  []ExecutionChainRun `json:"runs"`
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
	Limit int                 `json:"limit"`
}

// UpdateExecutionChainRequest represents the request to update a chain
type UpdateExecutionChainRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}
