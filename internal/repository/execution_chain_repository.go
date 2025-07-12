package repository

import (
	"context"
	"encoding/json"

	"github.com/sakibcoolz/loki-suite/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExecutionChainRepository defines the interface for execution chain data access
// This interface provides comprehensive data access methods for managing execution chains,
// their runs, and step executions with proper transaction support and relationship loading
type ExecutionChainRepository interface {
	// Chain management methods for CRUD operations on execution chains

	// CreateChain creates a new execution chain with its associated steps in a transaction
	// Validates JSON format of request parameters and sets proper step ordering
	CreateChain(ctx context.Context, chain *models.ExecutionChain) error

	// GetChainByID retrieves a single execution chain by its unique identifier
	// Preloads associated steps and webhook relationships for complete chain data
	GetChainByID(ctx context.Context, id uuid.UUID) (*models.ExecutionChain, error)

	// GetChainsByTenant retrieves all execution chains for a specific tenant with pagination
	// Returns chains with preloaded steps and total count for pagination metadata
	GetChainsByTenant(ctx context.Context, tenantID string, offset, limit int) ([]*models.ExecutionChain, int64, error)

	// GetChainsByTriggerEvent finds active chains that respond to a specific event type
	// Used by the webhook system to determine which chains to execute for incoming events
	GetChainsByTriggerEvent(ctx context.Context, tenantID, event string) ([]*models.ExecutionChain, error)

	// UpdateChain modifies specific fields of an execution chain using a map of updates
	// Allows partial updates without affecting unchanged fields
	UpdateChain(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error

	// DeleteChain soft deletes an execution chain and all its associated steps
	// Uses transaction to ensure data consistency during cascading deletion
	DeleteChain(ctx context.Context, id uuid.UUID) error

	// Chain execution methods for managing runtime execution instances

	// CreateChainRun initiates a new execution instance of a chain
	// Records the start of chain execution with initial status and metadata
	CreateChainRun(ctx context.Context, run *models.ExecutionChainRun) error

	// GetChainRunByID retrieves a specific chain execution run with full relationship data
	// Includes chain definition, step runs, and associated webhook information
	GetChainRunByID(ctx context.Context, runID uuid.UUID) (*models.ExecutionChainRun, error)

	// GetChainRunsByChain retrieves all execution runs for a specific chain with pagination
	// Provides execution history and audit trail for chain performance analysis
	GetChainRunsByChain(ctx context.Context, chainID uuid.UUID, offset, limit int) ([]*models.ExecutionChainRun, int64, error)

	// UpdateChainRunStatus updates the execution status of a chain run
	// Automatically sets completion timestamp for terminal statuses
	UpdateChainRunStatus(ctx context.Context, runID uuid.UUID, status models.ExecutionChainStatus) error

	// UpdateChainRunStep advances the current step pointer during chain execution
	// Tracks progress through the execution sequence
	UpdateChainRunStep(ctx context.Context, runID uuid.UUID, currentStep int) error

	// Step execution methods for managing individual step executions within a chain run

	// CreateStepRun records the execution of a single step within a chain run
	// Captures step-specific execution data, status, and results
	CreateStepRun(ctx context.Context, stepRun *models.ExecutionChainStepRun) error

	// GetStepRunsByRun retrieves all step executions for a specific chain run
	// Returns steps in execution order with preloaded step and webhook data
	GetStepRunsByRun(ctx context.Context, runID uuid.UUID) ([]*models.ExecutionChainStepRun, error)

	// UpdateStepRun modifies specific fields of a step execution run
	// Used to update status, results, or error information during step execution
	UpdateStepRun(ctx context.Context, stepRunID uuid.UUID, updates map[string]interface{}) error
}

// executionChainRepository implements ExecutionChainRepository interface
// Provides concrete implementation of execution chain data access using GORM ORM
// Handles database transactions, relationship loading, and error handling
type executionChainRepository struct {
	// db is the GORM database instance for executing queries
	// Provides transaction support and relationship management
	db *gorm.DB
}

// NewExecutionChainRepository creates a new execution chain repository instance
// Factory function that initializes the repository with a database connection
// Returns: ExecutionChainRepository interface implementation
func NewExecutionChainRepository(db *gorm.DB) ExecutionChainRepository {
	return &executionChainRepository{db: db}
}

// CreateChain creates a new execution chain with its associated steps
// Performs atomic creation of chain and steps within a database transaction
// Validates JSON format of request parameters and establishes proper step ordering
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - chain: ExecutionChain model with populated steps slice
//
// Returns: error if creation fails, nil on success
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

// GetChainByID retrieves a single execution chain by its unique identifier
// Loads complete chain data including steps and associated webhook relationships
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - id: UUID of the execution chain to retrieve
//
// Returns: ExecutionChain pointer with preloaded relationships, error if not found
func (r *executionChainRepository) GetChainByID(ctx context.Context, id uuid.UUID) (*models.ExecutionChain, error) {
	var chain models.ExecutionChain
	err := r.db.WithContext(ctx).Preload("Steps.Webhook").Where("id = ?", id).First(&chain).Error
	if err != nil {
		return nil, err
	}
	return &chain, nil
}

// GetChainsByTenant retrieves all execution chains for a specific tenant with pagination
// Provides complete chain data with steps and webhook relationships for tenant dashboard
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - tenantID: Tenant identifier to filter chains
//   - offset: Number of records to skip for pagination
//   - limit: Maximum number of records to return
//
// Returns: Slice of ExecutionChain pointers, total count, error if query fails
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

// GetChainsByTriggerEvent finds active execution chains that respond to a specific event
// Critical method for webhook event processing to determine which chains to execute
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - tenantID: Tenant identifier to scope the search
//   - event: Event type that triggers chain execution
//
// Returns: Slice of ExecutionChain with steps ordered by execution sequence, error if query fails
func (r *executionChainRepository) GetChainsByTriggerEvent(ctx context.Context, tenantID, event string) ([]*models.ExecutionChain, error) {
	var chains []*models.ExecutionChain
	err := r.db.WithContext(ctx).
		Preload("Steps", "steps.step_order ASC").
		Preload("Steps.Webhook").
		Where("tenant_id = ? AND trigger_event = ? AND is_active = ?", tenantID, event, true).
		Find(&chains).Error
	return chains, err
}

// UpdateChain modifies specific fields of an execution chain
// Allows partial updates using a map of field names to new values
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - id: UUID of the execution chain to update
//   - updates: Map of field names to new values for selective updating
//
// Returns: error if update fails, nil on success
func (r *executionChainRepository) UpdateChain(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.ExecutionChain{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteChain performs soft deletion of an execution chain and its associated steps
// Uses database transaction to ensure data consistency during cascading deletion
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - id: UUID of the execution chain to delete
//
// Returns: error if deletion fails, nil on success
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

// CreateChainRun initiates a new execution instance of a chain
// Records the start of chain execution with initial status and trigger context
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - run: ExecutionChainRun model with execution metadata and initial status
//
// Returns: error if creation fails, nil on success
func (r *executionChainRepository) CreateChainRun(ctx context.Context, run *models.ExecutionChainRun) error {
	return r.db.WithContext(ctx).Create(run).Error
}

// GetChainRunByID retrieves a specific chain execution run with complete relationship data
// Loads chain definition, step runs, and associated webhook information for execution tracking
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - runID: UUID of the chain execution run to retrieve
//
// Returns: ExecutionChainRun pointer with preloaded relationships, error if not found
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

// GetChainRunsByChain retrieves all execution runs for a specific chain with pagination
// Provides execution history and audit trail for chain performance analysis and debugging
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - chainID: UUID of the execution chain to get runs for
//   - offset: Number of records to skip for pagination
//   - limit: Maximum number of records to return
//
// Returns: Slice of ExecutionChainRun pointers, total count, error if query fails
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

// UpdateChainRunStatus updates the execution status of a chain run
// Automatically sets completion timestamp for terminal statuses (completed/failed)
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - runID: UUID of the chain execution run to update
//   - status: New execution status to set
//
// Returns: error if update fails, nil on success
func (r *executionChainRepository) UpdateChainRunStatus(ctx context.Context, runID uuid.UUID, status models.ExecutionChainStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == models.ExecutionChainStatusCompleted || status == models.ExecutionChainStatusFailed {
		updates["completed_at"] = gorm.Expr("NOW()")
	}

	return r.db.WithContext(ctx).Model(&models.ExecutionChainRun{}).Where("id = ?", runID).Updates(updates).Error
}

// UpdateChainRunStep advances the current step pointer during chain execution
// Tracks progress through the execution sequence for monitoring and resume capability
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - runID: UUID of the chain execution run to update
//   - currentStep: Step number (1-based) currently being executed
//
// Returns: error if update fails, nil on success
func (r *executionChainRepository) UpdateChainRunStep(ctx context.Context, runID uuid.UUID, currentStep int) error {
	return r.db.WithContext(ctx).Model(&models.ExecutionChainRun{}).
		Where("id = ?", runID).
		Update("current_step", currentStep).Error
}

// CreateStepRun records the execution of a single step within a chain run
// Captures step-specific execution data, status, results, and error information
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - stepRun: ExecutionChainStepRun model with step execution details
//
// Returns: error if creation fails, nil on success
func (r *executionChainRepository) CreateStepRun(ctx context.Context, stepRun *models.ExecutionChainStepRun) error {
	return r.db.WithContext(ctx).Create(stepRun).Error
}

// GetStepRunsByRun retrieves all step executions for a specific chain run
// Returns steps in execution order with preloaded step and webhook data for analysis
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - runID: UUID of the chain execution run to get step runs for
//
// Returns: Slice of ExecutionChainStepRun pointers ordered by step execution sequence, error if query fails
func (r *executionChainRepository) GetStepRunsByRun(ctx context.Context, runID uuid.UUID) ([]*models.ExecutionChainStepRun, error) {
	var stepRuns []*models.ExecutionChainStepRun
	err := r.db.WithContext(ctx).
		Preload("Step.Webhook").
		Where("run_id = ?", runID).
		Order("step_order ASC").
		Find(&stepRuns).Error
	return stepRuns, err
}

// UpdateStepRun modifies specific fields of a step execution run
// Used to update status, results, error information, or timing data during step execution
// Parameters:
//   - ctx: Context for request cancellation and timeout control
//   - stepRunID: UUID of the step execution run to update
//   - updates: Map of field names to new values for selective updating
//
// Returns: error if update fails, nil on success
func (r *executionChainRepository) UpdateStepRun(ctx context.Context, stepRunID uuid.UUID, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.ExecutionChainStepRun{}).Where("id = ?", stepRunID).Updates(updates).Error
}
