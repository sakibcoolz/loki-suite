package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"loki-suite/internal/config"
	"loki-suite/internal/models"
	"loki-suite/internal/repository"
	"loki-suite/pkg/logger"
	"loki-suite/pkg/security"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ExecutionChainService handles execution chain business logic
type ExecutionChainService interface {
	// Chain management
	CreateChain(ctx context.Context, req *models.CreateExecutionChainRequest) (*models.CreateExecutionChainResponse, error)
	GetChain(ctx context.Context, chainID uuid.UUID) (*models.ExecutionChain, error)
	ListChains(ctx context.Context, tenantID string, page, limit int) (*models.ExecutionChainListResponse, error)
	UpdateChain(ctx context.Context, chainID uuid.UUID, req *models.UpdateExecutionChainRequest) error
	DeleteChain(ctx context.Context, chainID uuid.UUID) error

	// Chain execution
	ExecuteChain(ctx context.Context, req *models.ExecuteChainRequest) (*models.ExecuteChainResponse, error)
	ExecuteChainByEvent(ctx context.Context, tenantID, event string, eventData map[string]interface{}) error
	GetChainRun(ctx context.Context, runID uuid.UUID) (*models.ExecutionChainRun, error)
	ListChainRuns(ctx context.Context, chainID uuid.UUID, page, limit int) (*models.ExecutionChainRunsResponse, error)
}

// executionChainService implements ExecutionChainService
type executionChainService struct {
	chainRepo   repository.ExecutionChainRepository
	webhookRepo repository.WebhookRepository
	security    *security.SecurityService
	config      *config.Config
	httpClient  *http.Client
}

// NewExecutionChainService creates a new execution chain service
func NewExecutionChainService(
	chainRepo repository.ExecutionChainRepository,
	webhookRepo repository.WebhookRepository,
	security *security.SecurityService,
	config *config.Config,
) ExecutionChainService {
	return &executionChainService{
		chainRepo:   chainRepo,
		webhookRepo: webhookRepo,
		security:    security,
		config:      config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Webhook.TimeoutSeconds) * time.Second,
		},
	}
}

// CreateChain creates a new execution chain
func (s *executionChainService) CreateChain(ctx context.Context, req *models.CreateExecutionChainRequest) (*models.CreateExecutionChainResponse, error) {
	logger.Info("Creating execution chain",
		zap.String("tenant_id", req.TenantID),
		zap.String("name", req.Name),
		zap.String("trigger_event", req.TriggerEvent),
		zap.Int("steps_count", len(req.Steps)))

	// Validate that all webhook IDs exist and belong to the tenant
	for i, step := range req.Steps {
		webhook, err := s.webhookRepo.GetSubscriptionByID(step.WebhookID)
		if err != nil {
			return nil, fmt.Errorf("step %d: webhook not found: %w", i+1, err)
		}
		if webhook.TenantID != req.TenantID {
			return nil, fmt.Errorf("step %d: webhook belongs to different tenant", i+1)
		}
	}

	// Create execution chain
	chain := &models.ExecutionChain{
		ID:           uuid.New(),
		TenantID:     req.TenantID,
		Name:         req.Name,
		Description:  req.Description,
		TriggerEvent: req.TriggerEvent,
		Status:       models.ExecutionChainStatusPending,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Create steps
	for i, stepReq := range req.Steps {
		// Convert request params to JSON
		var requestParamsJSON string
		if stepReq.RequestParams != nil {
			paramsBytes, err := json.Marshal(stepReq.RequestParams)
			if err != nil {
				return nil, fmt.Errorf("step %d: invalid request params: %w", i+1, err)
			}
			requestParamsJSON = string(paramsBytes)
		}

		// Set default actions
		onSuccessAction := stepReq.OnSuccessAction
		if onSuccessAction == "" {
			onSuccessAction = "continue"
		}

		onFailureAction := stepReq.OnFailureAction
		if onFailureAction == "" {
			onFailureAction = "stop"
		}

		maxRetries := stepReq.MaxRetries
		if maxRetries == 0 {
			maxRetries = 3
		}

		step := models.ExecutionChainStep{
			ID:              uuid.New(),
			WebhookID:       stepReq.WebhookID,
			Name:            stepReq.Name,
			Description:     stepReq.Description,
			RequestParams:   requestParamsJSON,
			OnSuccessAction: onSuccessAction,
			OnFailureAction: onFailureAction,
			MaxRetries:      maxRetries,
			DelaySeconds:    stepReq.DelaySeconds,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		chain.Steps = append(chain.Steps, step)
	}

	// Save to database
	if err := s.chainRepo.CreateChain(ctx, chain); err != nil {
		logger.Error("Failed to create execution chain", zap.Error(err))
		return nil, fmt.Errorf("failed to create execution chain: %w", err)
	}

	logger.Info("Execution chain created successfully",
		zap.String("chain_id", chain.ID.String()),
		zap.String("tenant_id", req.TenantID))

	return &models.CreateExecutionChainResponse{
		ChainID:      chain.ID,
		Name:         chain.Name,
		TriggerEvent: chain.TriggerEvent,
		StepsCount:   len(chain.Steps),
		Status:       string(chain.Status),
		CreatedAt:    chain.CreatedAt,
	}, nil
}

// GetChain retrieves a chain by ID
func (s *executionChainService) GetChain(ctx context.Context, chainID uuid.UUID) (*models.ExecutionChain, error) {
	return s.chainRepo.GetChainByID(ctx, chainID)
}

// ListChains lists chains for a tenant with pagination
func (s *executionChainService) ListChains(ctx context.Context, tenantID string, page, limit int) (*models.ExecutionChainListResponse, error) {
	offset := (page - 1) * limit
	chains, total, err := s.chainRepo.GetChainsByTenant(ctx, tenantID, offset, limit)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responseChains := make([]models.ExecutionChain, len(chains))
	for i, chain := range chains {
		responseChains[i] = *chain
	}

	return &models.ExecutionChainListResponse{
		Chains: responseChains,
		Total:  total,
		Page:   page,
		Limit:  limit,
	}, nil
}

// UpdateChain updates a chain's properties
func (s *executionChainService) UpdateChain(ctx context.Context, chainID uuid.UUID, req *models.UpdateExecutionChainRequest) error {
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		return s.chainRepo.UpdateChain(ctx, chainID, updates)
	}

	return nil
}

// DeleteChain deletes a chain
func (s *executionChainService) DeleteChain(ctx context.Context, chainID uuid.UUID) error {
	return s.chainRepo.DeleteChain(ctx, chainID)
}

// ExecuteChain manually executes a chain
func (s *executionChainService) ExecuteChain(ctx context.Context, req *models.ExecuteChainRequest) (*models.ExecuteChainResponse, error) {
	logger.Info("Executing chain manually",
		zap.String("chain_id", req.ChainID.String()))

	// Get the chain
	chain, err := s.chainRepo.GetChainByID(ctx, req.ChainID)
	if err != nil {
		return nil, fmt.Errorf("chain not found: %w", err)
	}

	if !chain.IsActive {
		return nil, fmt.Errorf("chain is not active")
	}

	// Create trigger data JSON
	var triggerDataJSON string
	if req.TriggerData != nil {
		triggerBytes, err := json.Marshal(req.TriggerData)
		if err != nil {
			return nil, fmt.Errorf("invalid trigger data: %w", err)
		}
		triggerDataJSON = string(triggerBytes)
	}

	// Create chain run
	now := time.Now()
	run := &models.ExecutionChainRun{
		ID:           uuid.New(),
		ChainID:      req.ChainID,
		TenantID:     chain.TenantID,
		Status:       models.ExecutionChainStatusRunning,
		TriggerEvent: chain.TriggerEvent,
		TriggerData:  triggerDataJSON,
		CurrentStep:  0,
		TotalSteps:   len(chain.Steps),
		StartedAt:    &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.chainRepo.CreateChainRun(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to create chain run: %w", err)
	}

	// Start executing the chain asynchronously
	go s.executeChainSteps(context.Background(), run.ID, chain, req.TriggerData)

	return &models.ExecuteChainResponse{
		RunID:      run.ID,
		ChainID:    req.ChainID,
		Status:     string(run.Status),
		TotalSteps: run.TotalSteps,
		StartedAt:  *run.StartedAt,
	}, nil
}

// ExecuteChainByEvent executes chains triggered by an event
func (s *executionChainService) ExecuteChainByEvent(ctx context.Context, tenantID, event string, eventData map[string]interface{}) error {
	logger.Info("Executing chains by event",
		zap.String("tenant_id", tenantID),
		zap.String("event", event))

	// Find chains that listen to this event
	chains, err := s.chainRepo.GetChainsByTriggerEvent(ctx, tenantID, event)
	if err != nil {
		return fmt.Errorf("failed to find chains for event: %w", err)
	}

	logger.Info("Found chains for event",
		zap.String("event", event),
		zap.Int("chains_count", len(chains)))

	// Execute each chain
	for _, chain := range chains {
		req := &models.ExecuteChainRequest{
			ChainID:     chain.ID,
			TriggerData: eventData,
		}

		if _, err := s.ExecuteChain(ctx, req); err != nil {
			logger.Error("Failed to execute chain",
				zap.String("chain_id", chain.ID.String()),
				zap.Error(err))
			// Continue with other chains even if one fails
		}
	}

	return nil
}

// GetChainRun retrieves a chain run by ID
func (s *executionChainService) GetChainRun(ctx context.Context, runID uuid.UUID) (*models.ExecutionChainRun, error) {
	return s.chainRepo.GetChainRunByID(ctx, runID)
}

// ListChainRuns lists runs for a chain with pagination
func (s *executionChainService) ListChainRuns(ctx context.Context, chainID uuid.UUID, page, limit int) (*models.ExecutionChainRunsResponse, error) {
	offset := (page - 1) * limit
	runs, total, err := s.chainRepo.GetChainRunsByChain(ctx, chainID, offset, limit)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responseRuns := make([]models.ExecutionChainRun, len(runs))
	for i, run := range runs {
		responseRuns[i] = *run
	}

	return &models.ExecutionChainRunsResponse{
		Runs:  responseRuns,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

// executeChainSteps executes the steps of a chain sequentially
func (s *executionChainService) executeChainSteps(ctx context.Context, runID uuid.UUID, chain *models.ExecutionChain, triggerData map[string]interface{}) {
	logger.Info("Starting chain execution",
		zap.String("run_id", runID.String()),
		zap.String("chain_id", chain.ID.String()),
		zap.Int("total_steps", len(chain.Steps)))

	for _, step := range chain.Steps {
		logger.Info("Executing step",
			zap.String("run_id", runID.String()),
			zap.Int("step_order", step.StepOrder),
			zap.String("step_name", step.Name))

		// Update current step
		if err := s.chainRepo.UpdateChainRunStep(ctx, runID, step.StepOrder); err != nil {
			logger.Error("Failed to update current step", zap.Error(err))
		}

		// Apply delay if specified
		if step.DelaySeconds > 0 {
			logger.Info("Applying step delay",
				zap.Int("delay_seconds", step.DelaySeconds))
			time.Sleep(time.Duration(step.DelaySeconds) * time.Second)
		}

		// Execute the step
		success := s.executeStep(ctx, runID, &step, triggerData)

		// Handle step result
		if success {
			logger.Info("Step executed successfully",
				zap.String("step_name", step.Name))

			if step.OnSuccessAction == "stop" {
				logger.Info("Stopping chain execution due to success action")
				break
			} else if step.OnSuccessAction == "pause" {
				logger.Info("Pausing chain execution")
				s.chainRepo.UpdateChainRunStatus(ctx, runID, models.ExecutionChainStatusPaused)
				return
			}
		} else {
			logger.Error("Step execution failed",
				zap.String("step_name", step.Name))

			if step.OnFailureAction == "stop" {
				logger.Info("Stopping chain execution due to failure")
				s.chainRepo.UpdateChainRunStatus(ctx, runID, models.ExecutionChainStatusFailed)
				return
			} else if step.OnFailureAction == "continue" {
				logger.Info("Continuing chain execution despite failure")
				continue
			}
		}
	}

	// Mark chain as completed
	logger.Info("Chain execution completed", zap.String("run_id", runID.String()))
	s.chainRepo.UpdateChainRunStatus(ctx, runID, models.ExecutionChainStatusCompleted)
}

// executeStep executes a single step with retry logic
func (s *executionChainService) executeStep(ctx context.Context, runID uuid.UUID, step *models.ExecutionChainStep, triggerData map[string]interface{}) bool {
	// Create step run
	now := time.Now()
	stepRun := &models.ExecutionChainStepRun{
		ID:        uuid.New(),
		RunID:     runID,
		StepID:    step.ID,
		StepOrder: step.StepOrder,
		Status:    models.WebhookStatusPending,
		StartedAt: &now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.chainRepo.CreateStepRun(ctx, stepRun); err != nil {
		logger.Error("Failed to create step run", zap.Error(err))
		return false
	}

	// Retry logic
	for attempt := 0; attempt <= step.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := time.Duration(attempt*attempt) * time.Second
			logger.Info("Retrying step execution",
				zap.Int("attempt", attempt),
				zap.Duration("delay", delay))
			time.Sleep(delay)
		}

		success, responseCode, responseBody, err := s.sendStepWebhook(ctx, step, triggerData)

		// Update step run
		updates := map[string]interface{}{
			"attempt_count": attempt + 1,
			"updated_at":    time.Now(),
		}

		if responseCode != nil {
			updates["response_code"] = *responseCode
		}

		if responseBody != nil {
			updates["response_body"] = *responseBody
		}

		if success {
			updates["status"] = models.WebhookStatusSent
			updates["completed_at"] = time.Now()
		} else {
			if err != nil {
				errMsg := err.Error()
				updates["last_error"] = errMsg
			}
			if attempt == step.MaxRetries {
				updates["status"] = models.WebhookStatusFailed
				updates["completed_at"] = time.Now()
			}
		}

		if err := s.chainRepo.UpdateStepRun(ctx, stepRun.ID, updates); err != nil {
			logger.Error("Failed to update step run", zap.Error(err))
		}

		if success {
			return true
		}

		logger.Error("Step execution attempt failed",
			zap.Int("attempt", attempt+1),
			zap.Error(err))
	}

	return false
}

// sendStepWebhook sends the webhook for a step
func (s *executionChainService) sendStepWebhook(ctx context.Context, step *models.ExecutionChainStep, triggerData map[string]interface{}) (bool, *int, *string, error) {
	// Prepare payload
	payload := map[string]interface{}{
		"step_name":    step.Name,
		"step_order":   step.StepOrder,
		"trigger_data": triggerData,
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	// Merge step-specific request params
	if step.RequestParams != "" {
		var stepParams map[string]interface{}
		if err := json.Unmarshal([]byte(step.RequestParams), &stepParams); err == nil {
			payload["request_params"] = stepParams
		}
	}

	// Convert to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return false, nil, nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", step.Webhook.TargetURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return false, nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "loki-suite-execution-chain/2.0")

	// Generate HMAC signature
	signature := s.security.GenerateHMACSignature(payloadBytes, step.Webhook.SecretToken)
	req.Header.Set("X-Shavix-Signature", fmt.Sprintf("sha256=%s", signature))
	req.Header.Set("X-Shavix-Timestamp", time.Now().Format(time.RFC3339))

	// Add JWT token for private webhooks
	if step.Webhook.Type == models.WebhookTypePrivate && step.Webhook.JWTToken != nil {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *step.Webhook.JWTToken))
	}

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, nil, nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, _ := io.ReadAll(resp.Body)
	responseBody := string(bodyBytes)

	// Check response status
	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	return success, &resp.StatusCode, &responseBody, nil
}
