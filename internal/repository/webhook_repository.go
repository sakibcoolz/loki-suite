package repository

import (
	"github.com/sakibcoolz/loki-suite/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WebhookRepository defines the interface for webhook data operations
type WebhookRepository interface {
	CreateSubscription(subscription *models.WebhookSubscription) error
	GetSubscriptionByID(id uuid.UUID) (*models.WebhookSubscription, error)
	GetActiveSubscriptionsByTenantAndEvent(tenantID, event string) ([]models.WebhookSubscription, error)
	GetSubscriptionsByTenant(tenantID string, offset, limit int) ([]models.WebhookSubscription, int64, error)
	UpdateSubscription(subscription *models.WebhookSubscription) error
	DeleteSubscription(id uuid.UUID) error

	CreateEvent(event *models.WebhookEvent) error
	GetEventByID(id uuid.UUID) (*models.WebhookEvent, error)
	UpdateEvent(event *models.WebhookEvent) error
	GetEventsByStatus(status models.WebhookStatus, limit int) ([]models.WebhookEvent, error)
}

// webhookRepository implements WebhookRepository
type webhookRepository struct {
	db *gorm.DB
}

// NewWebhookRepository creates a new webhook repository
func NewWebhookRepository(db *gorm.DB) WebhookRepository {
	return &webhookRepository{db: db}
}

// Subscription operations

func (r *webhookRepository) CreateSubscription(subscription *models.WebhookSubscription) error {
	return r.db.Create(subscription).Error
}

func (r *webhookRepository) GetSubscriptionByID(id uuid.UUID) (*models.WebhookSubscription, error) {
	var subscription models.WebhookSubscription
	err := r.db.Where("id = ?", id).First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *webhookRepository) GetActiveSubscriptionsByTenantAndEvent(tenantID, event string) ([]models.WebhookSubscription, error) {
	var subscriptions []models.WebhookSubscription
	err := r.db.Where("tenant_id = ? AND subscribed_event = ? AND is_active = ?",
		tenantID, event, true).Find(&subscriptions).Error
	return subscriptions, err
}

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

func (r *webhookRepository) UpdateSubscription(subscription *models.WebhookSubscription) error {
	return r.db.Save(subscription).Error
}

func (r *webhookRepository) DeleteSubscription(id uuid.UUID) error {
	return r.db.Delete(&models.WebhookSubscription{}, id).Error
}

// Event operations

func (r *webhookRepository) CreateEvent(event *models.WebhookEvent) error {
	return r.db.Create(event).Error
}

func (r *webhookRepository) GetEventByID(id uuid.UUID) (*models.WebhookEvent, error) {
	var event models.WebhookEvent
	err := r.db.Where("id = ?", id).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *webhookRepository) UpdateEvent(event *models.WebhookEvent) error {
	return r.db.Save(event).Error
}

func (r *webhookRepository) GetEventsByStatus(status models.WebhookStatus, limit int) ([]models.WebhookEvent, error) {
	var events []models.WebhookEvent
	err := r.db.Where("status = ?", status).
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}
