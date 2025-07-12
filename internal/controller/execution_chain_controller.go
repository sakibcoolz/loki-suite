package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sakibcoolz/loki-suite/internal/models"
	"github.com/sakibcoolz/loki-suite/internal/service"
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
}

// ExecutionChainController handles HTTP requests for execution chains
type ExecutionChainController struct {
	service service.ExecutionChainService
}

// NewExecutionChainController creates a new execution chain controller
func NewExecutionChainController(service service.ExecutionChainService) *ExecutionChainController {
	return &ExecutionChainController{
		service: service,
	}
}

// CreateChain handles POST /api/execution-chains
func (c *ExecutionChainController) CreateChain(ctx *gin.Context) {
	var req models.CreateExecutionChainRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request for chain creation", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Validate steps
	if len(req.Steps) == 0 {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "at least one step is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	response, err := c.service.CreateChain(ctx.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to create execution chain", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "chain_creation_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	logger.Info("Execution chain created successfully",
		zap.String("chain_id", response.ChainID.String()),
		zap.String("tenant_id", req.TenantID))

	ctx.JSON(http.StatusCreated, response)
}

// GetChain handles GET /api/execution-chains/:id
func (c *ExecutionChainController) GetChain(ctx *gin.Context) {
	chainIDStr := ctx.Param("id")
	chainID, err := uuid.Parse(chainIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_chain_id",
			Message: "Invalid chain ID format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	chain, err := c.service.GetChain(ctx.Request.Context(), chainID)
	if err != nil {
		logger.Error("Failed to get execution chain", zap.Error(err))
		ctx.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "chain_not_found",
			Message: "Execution chain not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	ctx.JSON(http.StatusOK, chain)
}

// ListChains handles GET /api/execution-chains
func (c *ExecutionChainController) ListChains(ctx *gin.Context) {
	tenantID := ctx.Query("tenant_id")
	if tenantID == "" {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "missing_tenant_id",
			Message: "tenant_id query parameter is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	response, err := c.service.ListChains(ctx.Request.Context(), tenantID, page, limit)
	if err != nil {
		logger.Error("Failed to list execution chains", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "chains_listing_failed",
			Message: "Failed to retrieve execution chains",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateChain handles PUT /api/execution-chains/:id
func (c *ExecutionChainController) UpdateChain(ctx *gin.Context) {
	chainIDStr := ctx.Param("id")
	chainID, err := uuid.Parse(chainIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_chain_id",
			Message: "Invalid chain ID format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req models.UpdateExecutionChainRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request for chain update", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := c.service.UpdateChain(ctx.Request.Context(), chainID, &req); err != nil {
		logger.Error("Failed to update execution chain", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "chain_update_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	logger.Info("Execution chain updated successfully",
		zap.String("chain_id", chainID.String()))

	ctx.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Execution chain updated successfully",
	})
}

// DeleteChain handles DELETE /api/execution-chains/:id
func (c *ExecutionChainController) DeleteChain(ctx *gin.Context) {
	chainIDStr := ctx.Param("id")
	chainID, err := uuid.Parse(chainIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_chain_id",
			Message: "Invalid chain ID format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := c.service.DeleteChain(ctx.Request.Context(), chainID); err != nil {
		logger.Error("Failed to delete execution chain", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "chain_deletion_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	logger.Info("Execution chain deleted successfully",
		zap.String("chain_id", chainID.String()))

	ctx.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Execution chain deleted successfully",
	})
}

// ExecuteChain handles POST /api/execution-chains/:id/execute
func (c *ExecutionChainController) ExecuteChain(ctx *gin.Context) {
	chainIDStr := ctx.Param("id")
	chainID, err := uuid.Parse(chainIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_chain_id",
			Message: "Invalid chain ID format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var requestBody struct {
		TriggerData map[string]interface{} `json:"trigger_data,omitempty"`
	}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		logger.Error("Invalid request for chain execution", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	req := &models.ExecuteChainRequest{
		ChainID:     chainID,
		TriggerData: requestBody.TriggerData,
	}

	response, err := c.service.ExecuteChain(ctx.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute chain", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "chain_execution_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	logger.Info("Execution chain started successfully",
		zap.String("chain_id", chainID.String()),
		zap.String("run_id", response.RunID.String()))

	ctx.JSON(http.StatusAccepted, response)
}

// GetChainRun handles GET /api/execution-chains/runs/:runId
func (c *ExecutionChainController) GetChainRun(ctx *gin.Context) {
	runIDStr := ctx.Param("runId")
	runID, err := uuid.Parse(runIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_run_id",
			Message: "Invalid run ID format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	run, err := c.service.GetChainRun(ctx.Request.Context(), runID)
	if err != nil {
		logger.Error("Failed to get chain run", zap.Error(err))
		ctx.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "run_not_found",
			Message: "Chain run not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	ctx.JSON(http.StatusOK, run)
}

// ListChainRuns handles GET /api/execution-chains/:id/runs
func (c *ExecutionChainController) ListChainRuns(ctx *gin.Context) {
	chainIDStr := ctx.Param("id")
	chainID, err := uuid.Parse(chainIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_chain_id",
			Message: "Invalid chain ID format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	response, err := c.service.ListChainRuns(ctx.Request.Context(), chainID, page, limit)
	if err != nil {
		logger.Error("Failed to list chain runs", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "runs_listing_failed",
			Message: "Failed to retrieve chain runs",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}
