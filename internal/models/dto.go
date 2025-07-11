package models

import (
	"time"

	"github.com/google/uuid"
)

// Request DTOs

// GenerateWebhookRequest represents the request to generate a new webhook
type GenerateWebhookRequest struct {
	TenantID        string      `json:"tenant_id" binding:"required"`
	AppName         string      `json:"app_name" binding:"required"`
	SubscribedEvent string      `json:"subscribed_event" binding:"required"`
	Type            WebhookType `json:"type" binding:"required"`
}

// SubscribeWebhookRequest represents the manual subscription request
type SubscribeWebhookRequest struct {
	TenantID        string      `json:"tenant_id" binding:"required"`
	AppName         string      `json:"app_name" binding:"required"`
	TargetURL       string      `json:"target_url" binding:"required"`
	SubscribedEvent string      `json:"subscribed_event" binding:"required"`
	Type            WebhookType `json:"type" binding:"required"`
}

// SendEventRequest represents the request to send a webhook event
type SendEventRequest struct {
	TenantID string      `json:"tenant_id" binding:"required"`
	Event    string      `json:"event" binding:"required"`
	Source   string      `json:"source" binding:"required"`
	Payload  interface{} `json:"payload" binding:"required"`
}

// Response DTOs

// GenerateWebhookResponse represents the response after generating a webhook
type GenerateWebhookResponse struct {
	WebhookURL  string      `json:"webhook_url"`
	SecretToken string      `json:"secret_token"`
	JWTToken    *string     `json:"jwt_token,omitempty"` // For private webhooks
	Type        WebhookType `json:"type"`
	WebhookID   uuid.UUID   `json:"webhook_id"`
}

// WebhookListResponse represents the response for listing webhooks
type WebhookListResponse struct {
	Webhooks []WebhookSubscription `json:"webhooks"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	Limit    int                   `json:"limit"`
}

// Webhook payload sent to external endpoints

// WebhookPayload represents the payload sent to webhook endpoints
type WebhookPayload struct {
	Event     string      `json:"event"`
	Source    string      `json:"source"`
	Timestamp string      `json:"timestamp"`
	Payload   interface{} `json:"payload"`
	EventID   uuid.UUID   `json:"event_id"`
}

// Common response types

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Version   string `json:"version"`
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
