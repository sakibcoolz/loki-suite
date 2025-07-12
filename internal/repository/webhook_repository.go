package repository

import (
	"github.com/sakibcoolz/loki-suite/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WebhookRepository defines the interface for webhook data operations
// This interface provides comprehensive data access methods for managing webhook subscriptions
// and webhook events with proper relationship handling and status-based querying
type WebhookRepository interface {
	// Subscription management methods for webhook endpoint registrations

	// CreateSubscription registers a new webhook subscription for event notifications
	// Establishes a webhook endpoint to receive specific event types
	CreateSubscription(subscription *models.WebhookSubscription) error

	// GetSubscriptionByID retrieves a specific webhook subscription by its unique identifier
	// Used for subscription verification and configuration retrieval
	GetSubscriptionByID(id uuid.UUID) (*models.WebhookSubscription, error)

	// GetActiveSubscriptionsByTenantAndEvent finds active subscriptions for event delivery
	// Critical method for determining which endpoints to notify when events occur
	GetActiveSubscriptionsByTenantAndEvent(tenantID, event string) ([]models.WebhookSubscription, error)

	// GetSubscriptionsByTenant retrieves all webhook subscriptions for a tenant with pagination
	// Provides subscription management dashboard data with pagination support
	GetSubscriptionsByTenant(tenantID string, offset, limit int) ([]models.WebhookSubscription, int64, error)

	// UpdateSubscription modifies an existing webhook subscription
	// Allows changes to endpoint URL, event types, security settings, and active status
	UpdateSubscription(subscription *models.WebhookSubscription) error

	// DeleteSubscription removes a webhook subscription from the system
	// Permanently deletes the subscription and stops future event deliveries
	DeleteSubscription(id uuid.UUID) error

	// Event management methods for webhook delivery tracking and retry logic

	// CreateEvent records a new webhook event for delivery processing
	// Initiates the webhook delivery pipeline with event metadata and payload
	CreateEvent(event *models.WebhookEvent) error

	// GetEventByID retrieves a specific webhook event by its unique identifier
	// Used for event status checking and delivery result analysis
	GetEventByID(id uuid.UUID) (*models.WebhookEvent, error)

	// UpdateEvent modifies webhook event fields during delivery processing
	// Updates delivery status, retry count, error information, and completion timestamps
	UpdateEvent(event *models.WebhookEvent) error

	// GetEventsByStatus retrieves webhook events filtered by delivery status
	// Essential for retry processing and delivery queue management
	GetEventsByStatus(status models.WebhookStatus, limit int) ([]models.WebhookEvent, error)
}

// webhookRepository implements WebhookRepository interface
// Provides concrete implementation of webhook data access using GORM ORM
// Handles database operations, relationship management, and query optimization
type webhookRepository struct {
	// db is the GORM database instance for executing queries
	// Provides transaction support and advanced querying capabilities
	db *gorm.DB
}

// NewWebhookRepository creates a new webhook repository instance
// Factory function that initializes the repository with a database connection
// Returns: WebhookRepository interface implementation for dependency injection
func NewWebhookRepository(db *gorm.DB) WebhookRepository {
	return &webhookRepository{db: db}
}

// Subscription operations - Methods for managing webhook endpoint registrations

// CreateSubscription registers a new webhook subscription in the database
// Establishes a webhook endpoint to receive notifications for specific event types
// Parameters:
//   - subscription: WebhookSubscription model with endpoint URL, events, and security config
//
// Returns: error if creation fails, nil on success
func (r *webhookRepository) CreateSubscription(subscription *models.WebhookSubscription) error {
	return r.db.Create(subscription).Error
}

// GetSubscriptionByID retrieves a specific webhook subscription by unique identifier
// Used for subscription verification, configuration retrieval, and access control
// Parameters:
//   - id: UUID of the webhook subscription to retrieve
//
// Returns: WebhookSubscription pointer if found, error if not found or query fails
func (r *webhookRepository) GetSubscriptionByID(id uuid.UUID) (*models.WebhookSubscription, error) {
	var subscription models.WebhookSubscription
	err := r.db.Where("id = ?", id).First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

// GetActiveSubscriptionsByTenantAndEvent finds active subscriptions for event delivery
// Critical method for webhook delivery pipeline to determine notification targets
// Parameters:
//   - tenantID: Tenant identifier to scope subscription search
//   - event: Event type that needs to be delivered to subscribed endpoints
//
// Returns: Slice of active WebhookSubscriptions, error if query fails
func (r *webhookRepository) GetActiveSubscriptionsByTenantAndEvent(tenantID, event string) ([]models.WebhookSubscription, error) {
	var subscriptions []models.WebhookSubscription
	err := r.db.Where("tenant_id = ? AND subscribed_event = ? AND is_active = ?",
		tenantID, event, true).Find(&subscriptions).Error
	return subscriptions, err
}

// GetSubscriptionsByTenant retrieves all webhook subscriptions for a tenant with pagination
// Provides subscription management dashboard data with total count for pagination
// Parameters:
//   - tenantID: Tenant identifier to filter subscriptions
//   - offset: Number of records to skip for pagination
//   - limit: Maximum number of records to return
//
// Returns: Slice of WebhookSubscriptions, total count, error if query fails
func (r *webhookRepository) GetSubscriptionsByTenant(tenantID string, offset, limit int) ([]models.WebhookSubscription, int64, error) {
	var subscriptions []models.WebhookSubscription
	var total int64

	// Get total count
	if err := r.db.Model(&models.WebhookSubscription{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&subscriptions).Error

	return subscriptions, total, err
}

// UpdateSubscription modifies an existing webhook subscription
// Allows changes to endpoint configuration, event types, and security settings
// Parameters:
//   - subscription: WebhookSubscription model with updated fields
//
// Returns: error if update fails, nil on success
func (r *webhookRepository) UpdateSubscription(subscription *models.WebhookSubscription) error {
	return r.db.Save(subscription).Error
}

// DeleteSubscription permanently removes a webhook subscription
// Stops future event deliveries to the associated endpoint
// Parameters:
//   - id: UUID of the webhook subscription to delete
//
// Returns: error if deletion fails, nil on success
func (r *webhookRepository) DeleteSubscription(id uuid.UUID) error {
	return r.db.Delete(&models.WebhookSubscription{}, id).Error
}

// Event operations - Methods for managing webhook delivery tracking and processing

// CreateEvent records a new webhook event for delivery processing
// Initiates the webhook delivery pipeline with event metadata and payload data
// Parameters:
//   - event: WebhookEvent model with payload, subscription info, and initial status
//
// Returns: error if creation fails, nil on success
func (r *webhookRepository) CreateEvent(event *models.WebhookEvent) error {
	return r.db.Create(event).Error
}

// GetEventByID retrieves a specific webhook event by unique identifier
// Used for event status checking, delivery result analysis, and debugging
// Parameters:
//   - id: UUID of the webhook event to retrieve
//
// Returns: WebhookEvent pointer if found, error if not found or query fails
func (r *webhookRepository) GetEventByID(id uuid.UUID) (*models.WebhookEvent, error) {
	var event models.WebhookEvent
	err := r.db.Where("id = ?", id).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// UpdateEvent modifies webhook event fields during delivery processing
// Updates delivery status, retry count, error information, and completion timestamps
// Parameters:
//   - event: WebhookEvent model with updated delivery status and metadata
//
// Returns: error if update fails, nil on success
func (r *webhookRepository) UpdateEvent(event *models.WebhookEvent) error {
	return r.db.Save(event).Error
}

// GetEventsByStatus retrieves webhook events filtered by delivery status
// Essential for retry processing, delivery queue management, and failure analysis
// Parameters:
//   - status: WebhookStatus to filter events (pending, delivered, failed, etc.)
//   - limit: Maximum number of events to return for batch processing
//
// Returns: Slice of WebhookEvents matching the status, error if query fails
func (r *webhookRepository) GetEventsByStatus(status models.WebhookStatus, limit int) ([]models.WebhookEvent, error) {
	var events []models.WebhookEvent
	err := r.db.Where("status = ?", status).
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}
