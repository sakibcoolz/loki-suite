package models

import (
	"time"

	"github.com/google/uuid"
)

// WebhookType defines the type of webhook subscription
// Determines the security model and authentication requirements for webhook delivery
type WebhookType string

const (
	// WebhookTypePublic represents webhooks that use only HMAC signature verification
	// Suitable for public endpoints that don't require JWT authentication
	WebhookTypePublic WebhookType = "public"

	// WebhookTypePrivate represents webhooks that require both HMAC and JWT authentication
	// Provides enhanced security for sensitive endpoints requiring authorization
	WebhookTypePrivate WebhookType = "private"
)

// WebhookStatus defines the delivery status of webhook events
// Tracks the lifecycle of webhook event processing and delivery attempts
type WebhookStatus string

const (
	// WebhookStatusPending indicates the webhook event is queued for delivery
	// Initial state when event is created but not yet processed
	WebhookStatusPending WebhookStatus = "pending"

	// WebhookStatusSent indicates successful delivery to at least one subscriber
	// Event was successfully processed and delivered to target endpoints
	WebhookStatusSent WebhookStatus = "sent"

	// WebhookStatusFailed indicates delivery failed to all subscribers
	// All delivery attempts failed due to network, authentication, or target errors
	WebhookStatusFailed WebhookStatus = "failed"
)

// ExecutionChainStatus defines the execution state of workflow chains
// Tracks the progress and outcome of multi-step webhook execution workflows
type ExecutionChainStatus string

const (
	// ExecutionChainStatusPending indicates the chain is created but not yet started
	// Waiting for trigger event or manual execution command
	ExecutionChainStatusPending ExecutionChainStatus = "pending"

	// ExecutionChainStatusRunning indicates the chain is currently executing
	// One or more steps are in progress or awaiting execution
	ExecutionChainStatusRunning ExecutionChainStatus = "running"

	// ExecutionChainStatusCompleted indicates successful completion of all steps
	// All steps executed successfully according to their success criteria
	ExecutionChainStatusCompleted ExecutionChainStatus = "completed"

	// ExecutionChainStatusFailed indicates the chain execution failed
	// A critical step failed and chain was terminated according to failure policy
	ExecutionChainStatusFailed ExecutionChainStatus = "failed"

	// ExecutionChainStatusPaused indicates execution was paused by success/failure action
	// Chain can be resumed manually or by external trigger
	ExecutionChainStatusPaused ExecutionChainStatus = "paused"
)

// WebhookSubscription represents a webhook subscription in the database
// Stores configuration and security credentials for webhook endpoints that receive event notifications
type WebhookSubscription struct {
	// ID is the unique identifier for this webhook subscription
	// Generated automatically using PostgreSQL's gen_random_uuid() function
	ID uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`

	// TenantID identifies the tenant/organization that owns this webhook
	// Used for multi-tenancy isolation and access control
	TenantID string `json:"tenant_id" gorm:"index;not null"`

	// AppName identifies the application or service that created this webhook
	// Used for organizing and filtering webhooks by source application
	AppName string `json:"app_name" gorm:"not null"`

	// Description provides additional context about this webhook's purpose
	// Optional field for documenting the webhook's business logic and usage
	Description *string `json:"description"`

	// TargetURL is the HTTP endpoint where webhook payloads will be delivered
	// Must be a valid HTTP/HTTPS URL accessible by the webhook service
	TargetURL string `json:"target_url" gorm:"not null"`

	// SubscribedEvent specifies which event type this webhook should receive
	// Acts as a filter to determine which events trigger this webhook
	SubscribedEvent string `json:"subscribed_event" gorm:"not null"`

	// Type determines the security model (public with HMAC or private with HMAC+JWT)
	// Affects authentication requirements and security headers sent with webhooks
	Type WebhookType `json:"type" gorm:"not null"`

	// SecretToken is the HMAC secret used for signature verification
	// Hidden from JSON responses for security, used to generate X-Shavix-Signature header
	SecretToken string `json:"-" gorm:"not null"`

	// JWTToken contains the JWT for private webhook authentication
	// Only populated for private webhooks, sent in Authorization header
	JWTToken *string `json:"-" gorm:"type:text"`

	// RetryCount tracks the number of failed delivery attempts
	// Used for implementing retry policies and delivery statistics
	RetryCount int `json:"retry_count" gorm:"default:0"`

	// MaxRetries sets the maximum number of retry attempts for failed deliveries
	// Prevents infinite retry loops and allows control over retry behavior
	MaxRetries int `json:"max_retries" gorm:"default:3"`

	// RetryDelaySeconds is the delay in seconds before retrying a failed delivery
	// Allows subscribers to control the backoff strategy for retries
	RetryDelaySeconds int `json:"retry_delay_seconds" gorm:"default:5"`

	// QueryParams is an optional map of query parameters included in webhook requests
	// Allows subscribers to specify additional parameters for the webhook URL
	QueryParams map[string]string `json:"query_params,omitempty" gorm:"type:jsonb"`

	// Headers is an optional map of custom headers to include in webhook requests
	// Allows subscribers to specify additional metadata or authentication headers
	Headers map[string]string `json:"headers,omitempty" gorm:"type:jsonb"`

	// Payload contains additional static data to be included with each webhook delivery
	// Stored as JSONB and merged with event payload when sending webhooks
	Payload string `json:"payload,omitempty" gorm:"type:jsonb"`

	// IsActive controls whether this webhook should receive events
	// Allows temporary disabling without deleting the subscription
	IsActive bool `json:"is_active" gorm:"default:true"`

	// CreatedAt timestamp when the subscription was first created
	// Automatically managed by GORM for audit trails
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt timestamp when the subscription was last modified
	// Automatically updated by GORM on any field changes
	UpdatedAt time.Time `json:"updated_at"`
}

// WebhookEvent represents a webhook event in the database
// Stores event data and delivery tracking information for webhook notifications
type WebhookEvent struct {
	// ID is the unique identifier for this webhook event
	// Generated automatically for tracking individual event deliveries
	ID uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`

	// TenantID identifies the tenant that generated this event
	// Used for isolation and filtering events by tenant
	TenantID string `json:"tenant_id" gorm:"index;not null"`

	// EventName specifies the type of event that occurred
	// Used to match against webhook subscriptions and trigger appropriate handlers
	EventName string `json:"event_name" gorm:"not null"`

	// Source identifies the system or service that generated this event
	// Provides context about the event origin for debugging and routing
	Source string `json:"source" gorm:"not null"`

	// Payload contains the JSON-encoded event data
	// Stored as JSONB in PostgreSQL for efficient querying and indexing
	Payload string `json:"payload" gorm:"type:jsonb"`

	// Status tracks the delivery status of this event
	// Indicates whether the event was successfully delivered to subscribers
	Status WebhookStatus `json:"status" gorm:"default:'pending'"`

	// ResponseCode stores the HTTP response code from the last delivery attempt
	// Used for debugging delivery failures and monitoring webhook health
	ResponseCode *int `json:"response_code"`

	// Attempts counts the number of delivery attempts made for this event
	// Used for retry logic and delivery statistics
	Attempts int `json:"attempts" gorm:"default:0"`

	// LastError contains the error message from the most recent failed delivery
	// Provides diagnostic information for troubleshooting webhook issues
	LastError *string `json:"last_error"`

	// SentAt timestamp when the event was successfully delivered
	// Only set when at least one webhook delivery succeeds
	SentAt *time.Time `json:"sent_at"`

	// CreatedAt timestamp when the event was first created
	// Automatically managed by GORM for audit trails
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt timestamp when the event was last modified
	// Updated when delivery status or attempts change
	UpdatedAt time.Time `json:"updated_at"`
}

// ExecutionChain represents a sequence of webhooks to be executed in order
// Defines a workflow that automatically executes multiple webhook calls when triggered by events
type ExecutionChain struct {
	// ID is the unique identifier for this execution chain
	// Generated automatically for referencing and managing chains
	ID uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`

	// TenantID identifies the tenant that owns this execution chain
	// Used for multi-tenancy isolation and access control
	TenantID string `json:"tenant_id" gorm:"index;not null"`

	// Name is a human-readable identifier for this chain
	// Used in management interfaces and logging for easy identification
	Name string `json:"name" gorm:"not null"`

	// Description provides additional context about the chain's purpose
	// Optional field for documenting the workflow's business logic
	Description string `json:"description"`

	// Status tracks the current execution state of the chain
	// Indicates whether the chain is ready, running, completed, or failed
	Status ExecutionChainStatus `json:"status" gorm:"default:'pending'"`

	// TriggerEvent specifies which event type will start this chain execution
	// When this event is received, the chain will begin executing its steps
	TriggerEvent string `json:"trigger_event" gorm:"not null"`

	// IsActive controls whether this chain should respond to trigger events
	// Allows temporary disabling of workflows without deletion
	IsActive bool `json:"is_active" gorm:"default:true"`

	// CreatedAt timestamp when the chain was first created
	// Automatically managed by GORM for audit trails
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt timestamp when the chain configuration was last modified
	// Updated when chain properties or steps are changed
	UpdatedAt time.Time `json:"updated_at"`

	// Steps contains the ordered list of webhook calls in this chain
	// Foreign key relationship with cascading delete for data consistency
	Steps []ExecutionChainStep `json:"steps" gorm:"foreignKey:ChainID;constraint:OnDelete:CASCADE"`
}

// ExecutionChainStep represents a single step in an execution chain workflow
// Defines the configuration and behavior for one webhook call within a multi-step process
type ExecutionChainStep struct {
	// ID is the unique identifier for this step
	// Generated automatically for referencing individual steps
	ID uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`

	// ChainID links this step to its parent execution chain
	// Foreign key relationship for maintaining workflow structure
	ChainID uuid.UUID `json:"chain_id" gorm:"type:uuid;not null"`

	// StepOrder defines the sequence position of this step in the chain
	// Determines the execution order, starting from 1
	StepOrder int `json:"step_order" gorm:"not null"`

	// WebhookID references the webhook subscription to call in this step
	// Links to the actual webhook endpoint configuration and security settings
	WebhookID uuid.UUID `json:"webhook_id" gorm:"type:uuid;not null"`

	// Name is a human-readable identifier for this step
	// Used for logging, debugging, and management interfaces
	Name string `json:"name" gorm:"not null"`

	// Description provides additional context about this step's purpose
	// Optional field for documenting the step's role in the workflow
	Description string `json:"description"`

	// RequestParams contains additional parameters to include in the webhook call
	// Stored as JSONB for flexible parameter passing and merging with event data
	RequestParams string `json:"request_params" gorm:"type:jsonb"`

	// OnSuccessAction defines what to do when this step succeeds
	// Options: "continue" (next step), "stop" (end chain), "pause" (wait for manual resume)
	OnSuccessAction string `json:"on_success_action" gorm:"default:'continue'"`

	// OnFailureAction defines what to do when this step fails
	// Options: "continue" (ignore failure), "stop" (end chain), "retry" (attempt again)
	OnFailureAction string `json:"on_failure_action" gorm:"default:'stop'"`

	// RetryCount tracks the number of retry attempts made for this step
	// Incremented on each retry until MaxRetries is reached
	RetryCount int `json:"retry_count" gorm:"default:0"`

	// MaxRetries sets the maximum number of retry attempts for failed executions
	// Prevents infinite retry loops and provides failure handling
	MaxRetries int `json:"max_retries" gorm:"default:3"`

	// DelaySeconds specifies the wait time before executing this step
	// Used for implementing delays, rate limiting, or sequencing requirements
	DelaySeconds int `json:"delay_seconds" gorm:"default:0"`

	// CreatedAt timestamp when the step was first created
	// Automatically managed by GORM for audit trails
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt timestamp when the step configuration was last modified
	// Updated when step parameters or behavior settings change
	UpdatedAt time.Time `json:"updated_at"`

	// Chain provides access to the parent execution chain
	// Relationship for accessing chain-level configuration and metadata
	Chain ExecutionChain `json:"-" gorm:"foreignKey:ChainID"`

	// Webhook provides access to the webhook subscription details
	// Includes target URL, security credentials, and delivery configuration
	Webhook WebhookSubscription `json:"webhook" gorm:"foreignKey:WebhookID"`
}

// ExecutionChainRun represents a single execution instance of an execution chain
// Tracks the progress and results of a workflow execution from start to completion
type ExecutionChainRun struct {
	// ID is the unique identifier for this execution run
	// Generated automatically for tracking individual workflow executions
	ID uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`

	// ChainID links this run to the execution chain being executed
	// References the workflow definition and configuration
	ChainID uuid.UUID `json:"chain_id" gorm:"type:uuid;not null"`

	// TenantID identifies the tenant that owns this execution
	// Used for isolation and access control of execution results
	TenantID string `json:"tenant_id" gorm:"index;not null"`

	// Status tracks the current state of this execution run
	// Indicates whether the workflow is pending, running, completed, failed, or paused
	Status ExecutionChainStatus `json:"status" gorm:"default:'pending'"`

	// TriggerEvent stores the event name that initiated this execution
	// Provides context about what caused the workflow to start
	TriggerEvent string `json:"trigger_event"`

	// TriggerData contains the original event payload that started this execution
	// Stored as JSONB for passing context data through the workflow steps
	TriggerData string `json:"trigger_data" gorm:"type:jsonb"`

	// CurrentStep tracks which step is currently being executed or was last attempted
	// Zero-based index into the chain's steps array
	CurrentStep int `json:"current_step" gorm:"default:0"`

	// TotalSteps contains the total number of steps in this chain execution
	// Used for progress calculation and completion tracking
	TotalSteps int `json:"total_steps"`

	// StartedAt timestamp when the execution run began
	// Set when the first step starts executing
	StartedAt *time.Time `json:"started_at"`

	// CompletedAt timestamp when the execution finished (success or failure)
	// Set when the workflow reaches a terminal state
	CompletedAt *time.Time `json:"completed_at"`

	// LastError contains the error message from the most recent step failure
	// Provides diagnostic information for troubleshooting workflow issues
	LastError *string `json:"last_error"`

	// CreatedAt timestamp when the run was first created
	// Automatically managed by GORM for audit trails
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt timestamp when the run status was last modified
	// Updated as steps execute and the workflow progresses
	UpdatedAt time.Time `json:"updated_at"`

	// Chain provides access to the execution chain definition
	// Includes workflow configuration, steps, and metadata
	Chain ExecutionChain `json:"chain" gorm:"foreignKey:ChainID"`

	// StepRuns contains the execution results for each step in this run
	// Foreign key relationship with cascading delete for data consistency
	StepRuns []ExecutionChainStepRun `json:"step_runs" gorm:"foreignKey:RunID;constraint:OnDelete:CASCADE"`
}

// ExecutionChainStepRun represents the execution of a single step within a workflow run
// Captures the detailed results and timing information for each webhook call in a chain
type ExecutionChainStepRun struct {
	// ID is the unique identifier for this step execution
	// Generated automatically for tracking individual step results
	ID uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`

	// RunID links this step execution to its parent workflow run
	// Foreign key relationship for maintaining execution hierarchy
	RunID uuid.UUID `json:"run_id" gorm:"type:uuid;not null"`

	// StepID references the step definition that was executed
	// Links to the configuration and parameters for this step
	StepID uuid.UUID `json:"step_id" gorm:"type:uuid;not null"`

	// StepOrder indicates the position of this step in the execution sequence
	// Copied from the step definition for easier querying and sorting
	StepOrder int `json:"step_order"`

	// Status tracks the delivery status of this step's webhook call
	// Indicates whether the step is pending, sent successfully, or failed
	Status WebhookStatus `json:"status" gorm:"default:'pending'"`

	// RequestPayload contains the actual payload sent to the webhook endpoint
	// Includes merged trigger data and step-specific parameters as JSONB
	RequestPayload string `json:"request_payload" gorm:"type:jsonb"`

	// ResponseCode stores the HTTP status code returned by the webhook endpoint
	// Used for determining success/failure and debugging delivery issues
	ResponseCode *int `json:"response_code"`

	// ResponseBody contains the response body returned by the webhook endpoint
	// Stored as text for debugging and potential response processing
	ResponseBody *string `json:"response_body" gorm:"type:text"`

	// AttemptCount tracks the number of delivery attempts made for this step
	// Incremented on each retry until successful or max retries reached
	AttemptCount int `json:"attempt_count" gorm:"default:0"`

	// LastError contains the error message from the most recent failed attempt
	// Provides diagnostic information for troubleshooting step failures
	LastError *string `json:"last_error"`

	// StartedAt timestamp when this step began executing
	// Set when the webhook call is initiated
	StartedAt *time.Time `json:"started_at"`

	// CompletedAt timestamp when this step finished (success or final failure)
	// Set when the step reaches a terminal state
	CompletedAt *time.Time `json:"completed_at"`

	// CreatedAt timestamp when the step run was first created
	// Automatically managed by GORM for audit trails
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt timestamp when the step execution was last modified
	// Updated when status, attempts, or results change
	UpdatedAt time.Time `json:"updated_at"`

	// Run provides access to the parent workflow execution
	// Includes overall execution context and status
	Run ExecutionChainRun `json:"-" gorm:"foreignKey:RunID"`

	// Step provides access to the step definition and configuration
	// Includes webhook details, retry settings, and action policies
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
