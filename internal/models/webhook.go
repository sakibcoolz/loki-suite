package models

import (
	"time"

	"github.com/google/uuid"
)

// WebhookType defines the type of webhook
type WebhookType string

const (
	WebhookTypePublic  WebhookType = "public"
	WebhookTypePrivate WebhookType = "private"
)

// WebhookStatus defines the status of webhook events
type WebhookStatus string

const (
	WebhookStatusPending WebhookStatus = "pending"
	WebhookStatusSent    WebhookStatus = "sent"
	WebhookStatusFailed  WebhookStatus = "failed"
)

// ExecutionChainStatus defines the status of execution chains
type ExecutionChainStatus string

const (
	ExecutionChainStatusPending   ExecutionChainStatus = "pending"
	ExecutionChainStatusRunning   ExecutionChainStatus = "running"
	ExecutionChainStatusCompleted ExecutionChainStatus = "completed"
	ExecutionChainStatusFailed    ExecutionChainStatus = "failed"
	ExecutionChainStatusPaused    ExecutionChainStatus = "paused"
)

// WebhookSubscription represents a webhook subscription in the database
type WebhookSubscription struct {
	ID              uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID        string      `json:"tenant_id" gorm:"index;not null"`
	AppName         string      `json:"app_name" gorm:"not null"`
	TargetURL       string      `json:"target_url" gorm:"not null"`
	SubscribedEvent string      `json:"subscribed_event" gorm:"not null"`
	Type            WebhookType `json:"type" gorm:"not null"`
	SecretToken     string      `json:"-" gorm:"not null"`  // HMAC secret, hidden from JSON
	JWTToken        *string     `json:"-" gorm:"type:text"` // JWT token for private webhooks, hidden from JSON
	RetryCount      int         `json:"retry_count" gorm:"default:0"`
	IsActive        bool        `json:"is_active" gorm:"default:true"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// WebhookEvent represents a webhook event in the database
type WebhookEvent struct {
	ID           uuid.UUID     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID     string        `json:"tenant_id" gorm:"index;not null"`
	EventName    string        `json:"event_name" gorm:"not null"`
	Source       string        `json:"source" gorm:"not null"`
	Payload      string        `json:"payload" gorm:"type:jsonb"` // JSONB for PostgreSQL
	Status       WebhookStatus `json:"status" gorm:"default:'pending'"`
	ResponseCode *int          `json:"response_code"`
	Attempts     int           `json:"attempts" gorm:"default:0"`
	LastError    *string       `json:"last_error"`
	SentAt       *time.Time    `json:"sent_at"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// ExecutionChain represents a sequence of webhooks to be executed
type ExecutionChain struct {
	ID           uuid.UUID            `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID     string               `json:"tenant_id" gorm:"index;not null"`
	Name         string               `json:"name" gorm:"not null"`
	Description  string               `json:"description"`
	Status       ExecutionChainStatus `json:"status" gorm:"default:'pending'"`
	TriggerEvent string               `json:"trigger_event" gorm:"not null"` // Event that triggers this chain
	IsActive     bool                 `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`

	// Relationships
	Steps []ExecutionChainStep `json:"steps" gorm:"foreignKey:ChainID;constraint:OnDelete:CASCADE"`
}

// ExecutionChainStep represents a single step in an execution chain
type ExecutionChainStep struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ChainID         uuid.UUID `json:"chain_id" gorm:"type:uuid;not null"`
	StepOrder       int       `json:"step_order" gorm:"not null"` // Order of execution
	WebhookID       uuid.UUID `json:"webhook_id" gorm:"type:uuid;not null"`
	Name            string    `json:"name" gorm:"not null"`
	Description     string    `json:"description"`
	RequestParams   string    `json:"request_params" gorm:"type:jsonb"`            // Additional parameters for this step
	OnSuccessAction string    `json:"on_success_action" gorm:"default:'continue'"` // continue, stop, pause
	OnFailureAction string    `json:"on_failure_action" gorm:"default:'stop'"`     // continue, stop, retry
	RetryCount      int       `json:"retry_count" gorm:"default:0"`
	MaxRetries      int       `json:"max_retries" gorm:"default:3"`
	DelaySeconds    int       `json:"delay_seconds" gorm:"default:0"` // Delay before executing this step
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Relationships
	Chain   ExecutionChain      `json:"-" gorm:"foreignKey:ChainID"`
	Webhook WebhookSubscription `json:"webhook" gorm:"foreignKey:WebhookID"`
}

// ExecutionChainRun represents a single execution of a chain
type ExecutionChainRun struct {
	ID           uuid.UUID            `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ChainID      uuid.UUID            `json:"chain_id" gorm:"type:uuid;not null"`
	TenantID     string               `json:"tenant_id" gorm:"index;not null"`
	Status       ExecutionChainStatus `json:"status" gorm:"default:'pending'"`
	TriggerEvent string               `json:"trigger_event"`
	TriggerData  string               `json:"trigger_data" gorm:"type:jsonb"` // Original event data that triggered this run
	CurrentStep  int                  `json:"current_step" gorm:"default:0"`
	TotalSteps   int                  `json:"total_steps"`
	StartedAt    *time.Time           `json:"started_at"`
	CompletedAt  *time.Time           `json:"completed_at"`
	LastError    *string              `json:"last_error"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`

	// Relationships
	Chain    ExecutionChain          `json:"chain" gorm:"foreignKey:ChainID"`
	StepRuns []ExecutionChainStepRun `json:"step_runs" gorm:"foreignKey:RunID;constraint:OnDelete:CASCADE"`
}

// ExecutionChainStepRun represents the execution of a single step
type ExecutionChainStepRun struct {
	ID             uuid.UUID     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	RunID          uuid.UUID     `json:"run_id" gorm:"type:uuid;not null"`
	StepID         uuid.UUID     `json:"step_id" gorm:"type:uuid;not null"`
	StepOrder      int           `json:"step_order"`
	Status         WebhookStatus `json:"status" gorm:"default:'pending'"`
	RequestPayload string        `json:"request_payload" gorm:"type:jsonb"`
	ResponseCode   *int          `json:"response_code"`
	ResponseBody   *string       `json:"response_body" gorm:"type:text"`
	AttemptCount   int           `json:"attempt_count" gorm:"default:0"`
	LastError      *string       `json:"last_error"`
	StartedAt      *time.Time    `json:"started_at"`
	CompletedAt    *time.Time    `json:"completed_at"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`

	// Relationships
	Run  ExecutionChainRun  `json:"-" gorm:"foreignKey:RunID"`
	Step ExecutionChainStep `json:"step" gorm:"foreignKey:StepID"`
}

// TableName sets the table name for WebhookSubscription
func (WebhookSubscription) TableName() string {
	return "webhook_subscriptions"
}

// TableName sets the table name for WebhookEvent
func (WebhookEvent) TableName() string {
	return "webhook_events"
}

// TableName sets the table name for ExecutionChain
func (ExecutionChain) TableName() string {
	return "execution_chains"
}

// TableName sets the table name for ExecutionChainStep
func (ExecutionChainStep) TableName() string {
	return "execution_chain_steps"
}

// TableName sets the table name for ExecutionChainRun
func (ExecutionChainRun) TableName() string {
	return "execution_chain_runs"
}

// TableName sets the table name for ExecutionChainStepRun
func (ExecutionChainStepRun) TableName() string {
	return "execution_chain_step_runs"
}
