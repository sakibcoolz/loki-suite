package repository

import (
	"context"
	"encoding/json"

	"github.com/sakibcoolz/loki-suite/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExecutionChainRepository defines the interface for execution chain data access
type ExecutionChainRepository interface {
	// Chain management
	CreateChain(ctx context.Context, chain *models.ExecutionChain) error
	GetChainByID(ctx context.Context, id uuid.UUID) (*models.ExecutionChain, error)
	GetChainsByTenant(ctx context.Context, tenantID string, offset, limit int) ([]*models.ExecutionChain, int64, error)
	GetChainsByTriggerEvent(ctx context.Context, tenantID, event string) ([]*models.ExecutionChain, error)
	UpdateChain(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	DeleteChain(ctx context.Context, id uuid.UUID) error

	// Chain execution
	CreateChainRun(ctx context.Context, run *models.ExecutionChainRun) error
	GetChainRunByID(ctx context.Context, runID uuid.UUID) (*models.ExecutionChainRun, error)
	GetChainRunsByChain(ctx context.Context, chainID uuid.UUID, offset, limit int) ([]*models.ExecutionChainRun, int64, error)
	UpdateChainRunStatus(ctx context.Context, runID uuid.UUID, status models.ExecutionChainStatus) error
	UpdateChainRunStep(ctx context.Context, runID uuid.UUID, currentStep int) error

	// Step execution
	CreateStepRun(ctx context.Context, stepRun *models.ExecutionChainStepRun) error
	GetStepRunsByRun(ctx context.Context, runID uuid.UUID) ([]*models.ExecutionChainStepRun, error)
	UpdateStepRun(ctx context.Context, stepRunID uuid.UUID, updates map[string]interface{}) error
}

// executionChainRepository implements ExecutionChainRepository
type executionChainRepository struct {
	db *gorm.DB
}

// NewExecutionChainRepository creates a new execution chain repository
func NewExecutionChainRepository(db *gorm.DB) ExecutionChainRepository {
	return &executionChainRepository{db: db}
}

// CreateChain creates a new execution chain
func (r *executionChainRepository) CreateChain(ctx context.Context, chain *models.ExecutionChain) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create the chain
		if err := tx.Create(chain).Error; err != nil {
			return err
		}

		// Create steps with proper ordering
		for i, step := range chain.Steps {
			step.ChainID = chain.ID
			step.StepOrder = i + 1

			// Convert request params to JSON string
			if step.RequestParams != "" {
				// Validate JSON format
				var params map[string]interface{}
				if err := json.Unmarshal([]byte(step.RequestParams), &params); err != nil {
					return err
				}
			}

			if err := tx.Create(&step).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetChainByID retrieves a chain by its ID with steps
func (r *executionChainRepository) GetChainByID(ctx context.Context, id uuid.UUID) (*models.ExecutionChain, error) {
	var chain models.ExecutionChain
	err := r.db.WithContext(ctx).Preload("Steps.Webhook").Where("id = ?", id).First(&chain).Error
	if err != nil {
		return nil, err
	}
	return &chain, nil
}

// GetChainsByTenant retrieves chains for a specific tenant with pagination
func (r *executionChainRepository) GetChainsByTenant(ctx context.Context, tenantID string, offset, limit int) ([]*models.ExecutionChain, int64, error) {
	var chains []*models.ExecutionChain
	var total int64

	// Count total
	if err := r.db.WithContext(ctx).Model(&models.ExecutionChain{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get chains with steps
	err := r.db.WithContext(ctx).
		Preload("Steps.Webhook").
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&chains).Error

	return chains, total, err
}

// GetChainsByTriggerEvent retrieves active chains that listen to a specific event
func (r *executionChainRepository) GetChainsByTriggerEvent(ctx context.Context, tenantID, event string) ([]*models.ExecutionChain, error) {
	var chains []*models.ExecutionChain
	err := r.db.WithContext(ctx).
		Preload("Steps", "steps.step_order ASC").
		Preload("Steps.Webhook").
		Where("tenant_id = ? AND trigger_event = ? AND is_active = ?", tenantID, event, true).
		Find(&chains).Error
	return chains, err
}

// UpdateChain updates a chain's fields
func (r *executionChainRepository) UpdateChain(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.ExecutionChain{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteChain soft deletes a chain and its steps
func (r *executionChainRepository) DeleteChain(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete steps first
		if err := tx.Where("chain_id = ?", id).Delete(&models.ExecutionChainStep{}).Error; err != nil {
			return err
		}
		// Delete chain
		return tx.Where("id = ?", id).Delete(&models.ExecutionChain{}).Error
	})
}

// CreateChainRun creates a new chain execution run
func (r *executionChainRepository) CreateChainRun(ctx context.Context, run *models.ExecutionChainRun) error {
	return r.db.WithContext(ctx).Create(run).Error
}

// GetChainRunByID retrieves a chain run by ID with steps
func (r *executionChainRepository) GetChainRunByID(ctx context.Context, runID uuid.UUID) (*models.ExecutionChainRun, error) {
	var run models.ExecutionChainRun
	err := r.db.WithContext(ctx).
		Preload("Chain").
		Preload("StepRuns.Step.Webhook").
		Where("id = ?", runID).
		First(&run).Error
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// GetChainRunsByChain retrieves runs for a specific chain with pagination
func (r *executionChainRepository) GetChainRunsByChain(ctx context.Context, chainID uuid.UUID, offset, limit int) ([]*models.ExecutionChainRun, int64, error) {
	var runs []*models.ExecutionChainRun
	var total int64

	// Count total
	if err := r.db.WithContext(ctx).Model(&models.ExecutionChainRun{}).Where("chain_id = ?", chainID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get runs
	err := r.db.WithContext(ctx).
		Preload("StepRuns.Step").
		Where("chain_id = ?", chainID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&runs).Error

	return runs, total, err
}

// UpdateChainRunStatus updates the status of a chain run
func (r *executionChainRepository) UpdateChainRunStatus(ctx context.Context, runID uuid.UUID, status models.ExecutionChainStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == models.ExecutionChainStatusCompleted || status == models.ExecutionChainStatusFailed {
		updates["completed_at"] = gorm.Expr("NOW()")
	}

	return r.db.WithContext(ctx).Model(&models.ExecutionChainRun{}).Where("id = ?", runID).Updates(updates).Error
}

// UpdateChainRunStep updates the current step of a chain run
func (r *executionChainRepository) UpdateChainRunStep(ctx context.Context, runID uuid.UUID, currentStep int) error {
	return r.db.WithContext(ctx).Model(&models.ExecutionChainRun{}).
		Where("id = ?", runID).
		Update("current_step", currentStep).Error
}

// CreateStepRun creates a new step execution run
func (r *executionChainRepository) CreateStepRun(ctx context.Context, stepRun *models.ExecutionChainStepRun) error {
	return r.db.WithContext(ctx).Create(stepRun).Error
}

// GetStepRunsByRun retrieves all step runs for a chain run
func (r *executionChainRepository) GetStepRunsByRun(ctx context.Context, runID uuid.UUID) ([]*models.ExecutionChainStepRun, error) {
	var stepRuns []*models.ExecutionChainStepRun
	err := r.db.WithContext(ctx).
		Preload("Step.Webhook").
		Where("run_id = ?", runID).
		Order("step_order ASC").
		Find(&stepRuns).Error
	return stepRuns, err
}

// UpdateStepRun updates a step run's fields
func (r *executionChainRepository) UpdateStepRun(ctx context.Context, stepRunID uuid.UUID, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.ExecutionChainStepRun{}).Where("id = ?", stepRunID).Updates(updates).Error
}
