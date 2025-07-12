package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sakibcoolz/loki-suite/internal/models"
	"github.com/sakibcoolz/loki-suite/internal/service"
	"go.uber.org/zap"
)

// WebhookController handles webhook HTTP requests
type WebhookController struct {
	webhookSvc service.WebhookService
}

// NewWebhookController creates a new webhook controller
func NewWebhookController(webhookSvc service.WebhookService) *WebhookController {
	return &WebhookController{
		webhookSvc: webhookSvc,
	}
}

// GenerateWebhook handles POST /api/webhooks/generate
func (wc *WebhookController) GenerateWebhook(c *gin.Context) {
	var req models.GenerateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid generate webhook request",
			zap.Error(err),
			zap.String("remote_addr", c.ClientIP()))

		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	response, err := wc.webhookSvc.GenerateWebhook(&req)
	if err != nil {
		logger.Error("Failed to generate webhook",
			zap.Error(err),
			zap.String("tenant_id", req.TenantID),
			zap.String("app_name", req.AppName))

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "webhook_generation_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	logger.Info("Webhook generated successfully",
		zap.String("webhook_id", response.WebhookID.String()),
		zap.String("tenant_id", req.TenantID),
		zap.String("type", string(req.Type)))

	c.JSON(http.StatusCreated, response)
}

// SubscribeWebhook handles POST /api/webhooks/subscribe
func (wc *WebhookController) SubscribeWebhook(c *gin.Context) {
	var req models.SubscribeWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid subscribe webhook request",
			zap.Error(err),
			zap.String("remote_addr", c.ClientIP()))

		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	response, err := wc.webhookSvc.SubscribeWebhook(&req)
	if err != nil {
		logger.Error("Failed to subscribe webhook",
			zap.Error(err),
			zap.String("tenant_id", req.TenantID),
			zap.String("target_url", req.TargetURL))

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "webhook_subscription_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	logger.Info("Webhook subscription created successfully",
		zap.String("webhook_id", response.WebhookID.String()),
		zap.String("tenant_id", req.TenantID),
		zap.String("target_url", req.TargetURL))

	c.JSON(http.StatusCreated, response)
}

// SendEvent handles POST /api/webhooks/event
func (wc *WebhookController) SendEvent(c *gin.Context) {
	var req models.SendEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid send event request",
			zap.Error(err),
			zap.String("remote_addr", c.ClientIP()))

		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := wc.webhookSvc.SendEvent(&req)
	if err != nil {
		logger.Error("Failed to send webhook event",
			zap.Error(err),
			zap.String("tenant_id", req.TenantID),
			zap.String("event", req.Event))

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "event_processing_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	logger.Info("Webhook event processed successfully",
		zap.String("event_id", result.EventID.String()),
		zap.String("tenant_id", req.TenantID),
		zap.String("event", req.Event),
		zap.Int("total_sent", result.TotalSent),
		zap.Int("total_failed", result.TotalFailed))

	message := "Webhook event processed successfully"
	if result.TotalSent == 0 {
		message = "No active webhook subscriptions found for this event"
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: message,
		Data:    result,
	})
}

// ReceiveWebhook handles POST /api/webhooks/receive/:id
func (wc *WebhookController) ReceiveWebhook(c *gin.Context) {
	webhookIDStr := c.Param("id")

	// Parse webhook ID
	webhookID, err := uuid.Parse(webhookIDStr)
	if err != nil {
		logger.Warn("Invalid webhook ID in receive request",
			zap.String("webhook_id", webhookIDStr),
			zap.String("remote_addr", c.ClientIP()))

		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_webhook_id",
			Message: "Invalid webhook ID format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Read request body
	payload, err := c.GetRawData()
	if err != nil {
		logger.Warn("Failed to read request body",
			zap.String("webhook_id", webhookIDStr),
			zap.Error(err))

		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_payload",
			Message: "Failed to read request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Get headers
	signature := c.GetHeader("X-Shavix-Signature")
	timestamp := c.GetHeader("X-Shavix-Timestamp")
	authHeader := c.GetHeader("Authorization")

	// Verify webhook
	err = wc.webhookSvc.VerifyWebhook(webhookID, payload, signature, timestamp, authHeader)
	if err != nil {
		logger.Warn("Webhook verification failed",
			zap.String("webhook_id", webhookIDStr),
			zap.Error(err),
			zap.String("remote_addr", c.ClientIP()))

		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "webhook_verification_failed",
			Message: err.Error(),
			Code:    http.StatusUnauthorized,
		})
		return
	}

	logger.Info("Webhook received and verified successfully",
		zap.String("webhook_id", webhookIDStr),
		zap.String("remote_addr", c.ClientIP()))

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Webhook received and verified successfully",
		Data: gin.H{
			"webhook_id": webhookID,
			"timestamp":  time.Now().Format(time.RFC3339),
		},
	})
}

// ListWebhooks handles GET /api/webhooks
func (wc *WebhookController) ListWebhooks(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		logger.Warn("Missing tenant_id in list webhooks request",
			zap.String("remote_addr", c.ClientIP()))

		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "missing_tenant_id",
			Message: "tenant_id query parameter is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	response, err := wc.webhookSvc.ListWebhooks(tenantID, page, limit)
	if err != nil {
		logger.Error("Failed to list webhooks",
			zap.Error(err),
			zap.String("tenant_id", tenantID))

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "list_webhooks_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	logger.Debug("Webhooks listed successfully",
		zap.String("tenant_id", tenantID),
		zap.Int64("total", response.Total),
		zap.Int("page", page),
		zap.Int("limit", limit))

	c.JSON(http.StatusOK, response)
}

// HealthCheck handles GET /health
func (wc *WebhookController) HealthCheck(c *gin.Context) {
	response := models.HealthResponse{
		Status:    "healthy",
		Service:   "github.com/sakibcoolz/loki-suite",
		Version:   "2.0.0",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}
