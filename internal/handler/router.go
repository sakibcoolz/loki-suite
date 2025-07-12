package handler

import (
	"github.com/sakibcoolz/loki-suite/internal/controller"
	"github.com/sakibcoolz/loki-suite/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Router holds the HTTP router and controllers
type Router struct {
	engine                   *gin.Engine
	webhookController        *controller.WebhookController
	executionChainController *controller.ExecutionChainController
}

// NewRouter creates a new HTTP router
func NewRouter(
	webhookController *controller.WebhookController,
	executionChainController *controller.ExecutionChainController,
) *Router {
	return &Router{
		engine:                   gin.New(),
		webhookController:        webhookController,
		executionChainController: executionChainController,
	}
}

// Setup configures all routes and middleware
func (r *Router) Setup() {
	// Add middleware
	r.engine.Use(gin.Recovery())
	r.engine.Use(middleware.RequestLogger())
	r.engine.Use(middleware.CORS())

	// API routes
	api := r.engine.Group("/api")
	{
		// Webhook routes
		webhooks := api.Group("/webhooks")
		{
			webhooks.POST("/generate", r.webhookController.GenerateWebhook)
			webhooks.POST("/subscribe", r.webhookController.SubscribeWebhook)
			webhooks.POST("/event", r.webhookController.SendEvent)
			webhooks.POST("/receive/:id", r.webhookController.ReceiveWebhook)
			webhooks.GET("", r.webhookController.ListWebhooks)
		}

		// Execution chain routes
		chains := api.Group("/execution-chains")
		{
			chains.POST("", r.executionChainController.CreateChain)
			chains.GET("", r.executionChainController.ListChains)
			chains.GET("/:id", r.executionChainController.GetChain)
			chains.PUT("/:id", r.executionChainController.UpdateChain)
			chains.DELETE("/:id", r.executionChainController.DeleteChain)
			chains.POST("/:id/execute", r.executionChainController.ExecuteChain)
			chains.GET("/:id/runs", r.executionChainController.ListChainRuns)
			chains.GET("/runs/:runId", r.executionChainController.GetChainRun)
		}
	}

	// Health check endpoint
	r.engine.GET("/health", r.webhookController.HealthCheck)

	// Metrics endpoint (if needed)
	// r.engine.GET("/metrics", r.webhookController.Metrics)
}

// GetEngine returns the Gin engine
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
