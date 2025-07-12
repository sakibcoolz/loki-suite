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
		// Webhook routes - Handle webhook subscription and event management
		webhooks := api.Group("/webhooks")
		{
			// POST /api/webhooks/generate - Generates a new webhook subscription URL
			// Purpose: Creates a new webhook endpoint that can receive HTTP callbacks
			// Used when you need to create a webhook URL for external services to call
			webhooks.POST("/generate", r.webhookController.GenerateWebhook)

			// POST /api/webhooks/subscribe - Subscribes to webhook events with custom configuration
			// Purpose: Creates a webhook subscription with specific event filters and target URLs
			// Used when you want to register to receive specific webhook events
			webhooks.POST("/subscribe", r.webhookController.SubscribeWebhook)

			// POST /api/webhooks/event - Sends a webhook event to all matching subscribers
			// Purpose: Triggers webhook notifications to all registered subscribers for a specific event
			// Used when you want to broadcast an event to all listening webhooks
			webhooks.POST("/event", r.webhookController.SendEvent)

			// POST /api/webhooks/receive/:id - Receives incoming webhook payloads
			// Purpose: Endpoint that external services call to deliver webhook payloads
			// Used as the callback URL that gets registered with external webhook providers
			webhooks.POST("/receive/:id", r.webhookController.ReceiveWebhook)

			// GET /api/webhooks - Lists all webhook subscriptions for a tenant
			// Purpose: Retrieves all webhook subscriptions, with optional filtering by tenant
			// Used for management dashboards and webhook subscription overview
			webhooks.GET("", r.webhookController.ListWebhooks)
		}

		// Execution chain routes - Manage sequential webhook execution workflows
		chains := api.Group("/execution-chains")
		{
			// POST /api/execution-chains - Creates a new execution chain
			// Purpose: Defines a sequence of webhooks to be executed in order when triggered
			// Used to set up complex workflows that involve multiple webhook calls
			chains.POST("", r.executionChainController.CreateChain)

			// GET /api/execution-chains - Lists all execution chains for a tenant
			// Purpose: Retrieves all configured execution chains with optional filtering
			// Used for workflow management and chain overview dashboards
			chains.GET("", r.executionChainController.ListChains)

			// GET /api/execution-chains/:id - Gets details of a specific execution chain
			// Purpose: Retrieves complete configuration and steps of a single execution chain
			// Used to view chain details and for editing workflows
			chains.GET("/:id", r.executionChainController.GetChain)

			// PUT /api/execution-chains/:id - Updates an existing execution chain
			// Purpose: Modifies the configuration, steps, or settings of an execution chain
			// Used to edit workflows, add/remove steps, or change chain behavior
			chains.PUT("/:id", r.executionChainController.UpdateChain)

			// DELETE /api/execution-chains/:id - Deletes an execution chain
			// Purpose: Removes an execution chain and all its associated steps
			// Used to clean up unused workflows or decommission old automation
			chains.DELETE("/:id", r.executionChainController.DeleteChain)

			// POST /api/execution-chains/:id/execute - Manually triggers an execution chain
			// Purpose: Starts immediate execution of a chain with optional custom parameters
			// Used for testing workflows or manually triggering automation outside normal events
			chains.POST("/:id/execute", r.executionChainController.ExecuteChain)

			// GET /api/execution-chains/:id/runs - Lists execution history for a specific chain
			// Purpose: Retrieves all past executions of a chain with their status and results
			// Used for monitoring chain performance and debugging failed executions
			chains.GET("/:id/runs", r.executionChainController.ListChainRuns)

			// GET /api/execution-chains/runs/:runId - Gets details of a specific chain execution
			// Purpose: Retrieves detailed information about a single chain run including step results
			// Used for detailed debugging and execution analysis
			chains.GET("/runs/:runId", r.executionChainController.GetChainRun)
		}
	}

	// Health check endpoint
	// GET /health - Application health and readiness check
	// Purpose: Provides system health status for load balancers and monitoring tools
	// Returns 200 OK when the application is ready to serve requests
	r.engine.GET("/health", r.webhookController.HealthCheck)

	// Metrics endpoint (if needed)
	// r.engine.GET("/metrics", r.webhookController.Metrics)
}

// GetEngine returns the Gin engine
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
