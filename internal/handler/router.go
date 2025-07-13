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
		// Webhooks provide real-time event notifications and enable seamless integration between services
		webhooks := api.Group("/webhooks")
		{
			// POST /api/webhooks/generate - Generates a new webhook subscription URL
			// Purpose: Creates a secure webhook endpoint with auto-generated credentials for receiving HTTP callbacks
			// Workflow: Request validation → Generate UUID → Create security tokens → Store subscription → Return webhook URL
			//
			// Example 1 - Payment Provider Integration:
			//   POST /api/webhooks/generate
			//   {
			//     "tenant_id": "ecommerce-store",
			//     "app_name": "stripe-payments",
			//     "subscribed_event": "payment.completed",
			//     "type": "private"
			//   }
			//   Response: {
			//     "webhook_url": "https://api.loki-suite.com/api/webhooks/receive/uuid-12345",
			//     "secret_token": "sk_generated_secret_token",
			//     "jwt_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			//     "webhook_id": "uuid-12345"
			//   }
			//
			// Example 2 - Shipping Service Callback:
			//   POST /api/webhooks/generate
			//   {
			//     "tenant_id": "logistics-company",
			//     "app_name": "fedex-tracking",
			//     "subscribed_event": "shipment.status_updated",
			//     "type": "public"
			//   }
			//   Response: {
			//     "webhook_url": "https://api.loki-suite.com/api/webhooks/receive/uuid-67890",
			//     "secret_token": "fedex_webhook_secret",
			//     "type": "public"
			//   }
			//
			// Example 3 - Internal Microservice Communication:
			//   POST /api/webhooks/generate
			//   {
			//     "tenant_id": "saas-platform",
			//     "app_name": "notification-service",
			//     "subscribed_event": "user.action_required",
			//     "type": "private"
			//   }
			//   Response: {
			//     "webhook_url": "https://api.loki-suite.com/api/webhooks/receive/uuid-internal-001",
			//     "secret_token": "internal_service_secret",
			//     "jwt_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			//     "webhook_id": "uuid-internal-001"
			//   }
			webhooks.POST("/generate", r.webhookController.GenerateWebhook)

			// POST /api/webhooks/subscribe - Subscribes to webhook events with custom configuration
			// Purpose: Registers a webhook subscription with custom target URLs and event filters
			// Workflow: Validate target URL → Check event permissions → Generate security tokens → Store subscription → Verify endpoint
			//
			// Example 1 - CRM Integration for User Events:
			//   POST /api/webhooks/subscribe
			//   {
			//     "tenant_id": "tech-startup",
			//     "app_name": "salesforce-crm",
			//     "target_url": "https://hooks.salesforce.com/services/webhook/loki-integration",
			//     "subscribed_event": "user.created",
			//     "type": "public",
			//     "retry_policy": {"max_retries": 3, "backoff": "exponential"}
			//   }
			//   Response: {
			//     "webhook_id": "crm-integration-uuid",
			//     "secret_token": "salesforce_webhook_secret",
			//     "subscription_status": "active",
			//     "verification_required": true
			//   }
			//
			// Example 2 - Slack Notifications for Order Events:
			//   POST /api/webhooks/subscribe
			//   {
			//     "tenant_id": "restaurant-chain",
			//     "app_name": "slack-notifications",
			//     "target_url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
			//     "subscribed_event": "order.completed",
			//     "type": "public",
			//     "message_format": "slack",
			//     "filter_conditions": {"order_value": {"gt": 100}}
			//   }
			//   Response: {
			//     "webhook_id": "slack-orders-uuid",
			//     "secret_token": "slack_verification_token",
			//     "message": "Slack webhook configured for high-value orders"
			//   }
			//
			// Example 3 - Internal Billing System Integration:
			//   POST /api/webhooks/subscribe
			//   {
			//     "tenant_id": "finance-dept",
			//     "app_name": "billing-system",
			//     "target_url": "https://internal.billing.company.com/webhooks/invoicing",
			//     "subscribed_event": "invoice.paid",
			//     "type": "private",
			//     "headers": {"Authorization": "Bearer internal_api_token"},
			//     "timeout_seconds": 30
			//   }
			//   Response: {
			//     "webhook_id": "billing-invoice-uuid",
			//     "secret_token": "billing_webhook_secret",
			//     "jwt_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			//     "security_level": "enterprise"
			//   }
			webhooks.POST("/subscribe", r.webhookController.SubscribeWebhook)

			// POST /api/webhooks/event - Sends a webhook event to all matching subscribers
			// Purpose: Broadcasts events to all registered webhook subscribers with reliable delivery guarantees
			// Workflow: Event validation → Find subscribers → Parallel delivery → Retry failed attempts → Return delivery summary
			//
			// Example 1 - E-commerce Order Completion Event:
			//   POST /api/webhooks/event
			//   {
			//     "tenant_id": "ecommerce-store",
			//     "event": "order.completed",
			//     "source": "checkout-service",
			//     "payload": {
			//       "order_id": "ORD-2024-001234",
			//       "customer_id": "CUST-567890",
			//       "items": [
			//         {"product_id": "PROD-001", "quantity": 2, "price": 29.99},
			//         {"product_id": "PROD-002", "quantity": 1, "price": 49.99}
			//       ],
			//       "total_amount": 109.97,
			//       "payment_method": "credit_card",
			//       "shipping_address": {"street": "123 Main St", "city": "Anytown", "zip": "12345"}
			//     }
			//   }
			//   Response: {
			//     "event_id": "evt-001234",
			//     "total_subscribers": 5,
			//     "successful_deliveries": 4,
			//     "failed_deliveries": 1,
			//     "delivery_results": [
			//       {"webhook_id": "payment-service", "status": "delivered", "response_time": "120ms"},
			//       {"webhook_id": "inventory-service", "status": "delivered", "response_time": "85ms"},
			//       {"webhook_id": "email-service", "status": "failed", "error": "timeout", "retry_scheduled": true}
			//     ]
			//   }
			//
			// Example 2 - User Registration Event:
			//   POST /api/webhooks/event
			//   {
			//     "tenant_id": "saas-platform",
			//     "event": "user.registered",
			//     "source": "user-management-service",
			//     "payload": {
			//       "user_id": "usr_new_123456",
			//       "email": "john.doe@company.com",
			//       "registration_timestamp": "2024-01-15T10:30:00Z",
			//       "signup_source": "website",
			//       "user_preferences": {"newsletter": true, "notifications": false},
			//       "account_tier": "premium"
			//     }
			//   }
			//   Response: {
			//     "event_id": "evt-user-reg-001",
			//     "triggered_execution_chains": [
			//       {"chain_id": "user-onboarding-chain", "run_id": "run-456789", "status": "started"}
			//     ],
			//     "webhook_deliveries": 3,
			//     "estimated_completion": "2024-01-15T10:32:00Z"
			//   }
			//
			// Example 3 - System Alert Event:
			//   POST /api/webhooks/event
			//   {
			//     "tenant_id": "infrastructure-team",
			//     "event": "system.alert.critical",
			//     "source": "monitoring-service",
			//     "payload": {
			//       "alert_id": "ALERT-CPU-001",
			//       "severity": "critical",
			//       "service": "payment-processor",
			//       "metric": "cpu_usage",
			//       "current_value": 95.2,
			//       "threshold": 80.0,
			//       "timestamp": "2024-01-15T15:45:00Z",
			//       "affected_instances": ["prod-payment-01", "prod-payment-02"],
			//       "runbook_url": "https://wiki.company.com/runbooks/high-cpu"
			//     }
			//   }
			//   Response: {
			//     "event_id": "evt-alert-critical-001",
			//     "priority_delivery": true,
			//     "notification_channels": ["slack", "pagerduty", "email"],
			//     "escalation_triggered": true,
			//     "incident_id": "INC-2024-001"
			//   }
			webhooks.POST("/event", r.webhookController.SendEvent)

			// POST /api/webhooks/receive/:id - Receives incoming webhook payloads
			// Purpose: Secure endpoint for external services to deliver webhook payloads with authentication and validation
			// Workflow: ID validation → Security verification → Payload processing → Event triggering → Response
			//
			// Example 1 - Payment Provider Callback:
			//   POST /api/webhooks/receive/payment-webhook-uuid
			//   Headers: {
			//     "Content-Type": "application/json",
			//     "X-Loki-Signature": "sha256=abc123...",
			//     "X-Loki-Timestamp": "2024-01-15T10:30:00Z",
			//     "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
			//   }
			//   Body: {
			//     "event_type": "payment.succeeded",
			//     "payment_id": "pay_1234567890",
			//     "amount": 2999,
			//     "currency": "usd",
			//     "customer": "cus_customer123",
			//     "metadata": {"order_id": "ORD-001", "internal_ref": "payment-ref-456"}
			//   }
			//   Response: {
			//     "message": "Webhook received and processed successfully",
			//     "webhook_id": "payment-webhook-uuid",
			//     "event_triggered": "payment.completed",
			//     "processing_time": "45ms"
			//   }
			//
			// Example 2 - Shipping Status Update:
			//   POST /api/webhooks/receive/shipping-tracking-uuid
			//   Headers: {
			//     "Content-Type": "application/json",
			//     "X-Loki-Signature": "sha256=def456...",
			//     "X-Loki-Timestamp": "2024-01-15T14:20:00Z"
			//   }
			//   Body: {
			//     "tracking_number": "1Z999AA1234567890",
			//     "status": "delivered",
			//     "location": "123 Customer St, Anytown, ST 12345",
			//     "timestamp": "2024-01-15T14:15:00Z",
			//     "signature": "J. Customer",
			//     "delivery_photo_url": "https://tracking.fedex.com/photo/abc123"
			//   }
			//   Response: {
			//     "message": "Shipping status updated successfully",
			//     "customer_notified": true,
			//     "internal_systems_updated": ["order-management", "customer-portal"],
			//     "next_actions": ["trigger-review-request", "update-analytics"]
			//   }
			//
			// Example 3 - External System Integration:
			//   POST /api/webhooks/receive/external-system-uuid
			//   Headers: {
			//     "Content-Type": "application/json",
			//     "X-Loki-Signature": "sha256=ghi789...",
			//     "X-Loki-Timestamp": "2024-01-15T09:00:00Z",
			//     "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
			//   }
			//   Body: {
			//     "integration_event": "user_profile_updated",
			//     "external_user_id": "ext_user_789",
			//     "changes": {
			//       "email": "updated.email@company.com",
			//       "department": "engineering",
			//       "access_level": "senior"
			//     },
			//     "sync_timestamp": "2024-01-15T08:58:30Z"
			//   }
			//   Response: {
			//     "message": "External integration processed",
			//     "internal_user_updated": true,
			//     "permissions_synced": true,
			//     "audit_log_created": "audit-ext-sync-001"
			//   }
			webhooks.POST("/receive/:id", r.webhookController.ReceiveWebhook)

			// GET /api/webhooks - Lists all webhook subscriptions for a tenant
			// Purpose: Retrieves webhook subscriptions with filtering, pagination, and health status information
			// Workflow: Permission validation → Apply filters → Database query → Health checks → Format response
			//
			// Example 1 - Management Dashboard Query:
			//   GET /api/webhooks?tenant_id=ecommerce-store&page=1&limit=10&status=active
			//   Response: {
			//     "webhooks": [
			//       {
			//         "id": "payment-webhook-uuid",
			//         "app_name": "stripe-payments",
			//         "target_url": "https://api.loki-suite.com/api/webhooks/receive/payment-webhook-uuid",
			//         "subscribed_event": "payment.completed",
			//         "type": "private",
			//         "status": "active",
			//         "health": {"last_success": "2024-01-15T10:30:00Z", "success_rate": "99.2%", "avg_response_time": "120ms"},
			//         "created_at": "2024-01-01T00:00:00Z"
			//       },
			//       {
			//         "id": "inventory-webhook-uuid",
			//         "app_name": "inventory-service",
			//         "target_url": "https://api.loki-suite.com/api/webhooks/receive/inventory-webhook-uuid",
			//         "subscribed_event": "inventory.low_stock",
			//         "type": "private",
			//         "status": "active",
			//         "health": {"last_success": "2024-01-15T09:45:00Z", "success_rate": "98.7%", "avg_response_time": "85ms"}
			//       }
			//     ],
			//     "pagination": {"total": 25, "page": 1, "limit": 10, "total_pages": 3}
			//   }
			//
			// Example 2 - Health Monitoring Query:
			//   GET /api/webhooks?tenant_id=saas-platform&health_status=degraded&include_metrics=true
			//   Response: {
			//     "webhooks": [
			//       {
			//         "id": "email-service-uuid",
			//         "app_name": "email-notifications",
			//         "health": {
			//           "status": "degraded",
			//           "success_rate": "87.3%",
			//           "recent_failures": 8,
			//           "last_error": "timeout after 30s",
			//           "error_trend": "increasing"
			//         },
			//         "performance_metrics": {
			//           "avg_response_time": "2.8s",
			//           "p95_response_time": "8.5s",
			//           "total_requests_24h": 1250,
			//           "failed_requests_24h": 158
			//         },
			//         "recommendations": ["Increase timeout threshold", "Check email service capacity"]
			//       }
			//     ]
			//   }
			//
			// Example 3 - Audit and Compliance View:
			//   GET /api/webhooks?tenant_id=finance-dept&include_security=true&audit_mode=true
			//   Response: {
			//     "webhooks": [
			//       {
			//         "id": "billing-webhook-uuid",
			//         "app_name": "billing-system",
			//         "security_info": {
			//           "type": "private",
			//           "encryption": "TLS 1.3",
			//           "authentication": "JWT + HMAC",
			//           "last_rotation": "2024-01-01T00:00:00Z",
			//           "compliance": ["PCI-DSS", "SOX"]
			//         },
			//         "audit_trail": {
			//           "created_by": "admin@company.com",
			//           "last_modified": "2024-01-10T15:30:00Z",
			//           "modification_count": 3,
			//           "access_log_entries": 1247
			//         },
			//         "data_classification": "confidential",
			//         "retention_policy": "7_years"
			//       }
			//     ]
			//   }
			webhooks.GET("", r.webhookController.ListWebhooks)
		}

		// Execution chain routes - Manage sequential webhook execution workflows
		// Execution chains enable complex business process automation by orchestrating multiple webhook calls
		// in a specific sequence with data passing between steps and configurable error handling.
		chains := api.Group("/execution-chains")
		{
			// POST /api/execution-chains - Creates a new execution chain
			// Purpose: Defines a sequence of webhooks to be executed in order when triggered by events
			// Workflow: Event trigger → Step 1 → Step 2 → ... → Step N (with data passing between steps)
			//
			// Example 1 - E-commerce Order Processing Chain:
			//   POST /api/execution-chains
			//   {
			//     "tenant_id": "ecommerce-store",
			//     "name": "Complete Order Processing",
			//     "trigger_event": "order.placed",
			//     "steps": [
			//       {"webhook_id": "payment-service", "name": "Process Payment", "request_params": {"amount": "{{.trigger_data.total}}"}},
			//       {"webhook_id": "inventory-service", "name": "Update Inventory", "request_params": {"payment_id": "{{.step_1.response.payment_id}}"}},
			//       {"webhook_id": "shipping-service", "name": "Create Label", "request_params": {"order_id": "{{.trigger_data.order_id}}"}},
			//       {"webhook_id": "email-service", "name": "Send Confirmation", "request_params": {"tracking": "{{.step_3.response.tracking_number}}"}}
			//     ]
			//   }
			//
			// Example 2 - User Onboarding Workflow:
			//   POST /api/execution-chains
			//   {
			//     "tenant_id": "saas-platform",
			//     "name": "New User Onboarding",
			//     "trigger_event": "user.registered",
			//     "steps": [
			//       {"webhook_id": "email-service", "name": "Welcome Email", "request_params": {"template": "welcome", "user_email": "{{.trigger_data.email}}"}},
			//       {"webhook_id": "profile-service", "name": "Create Profile", "request_params": {"user_data": "{{.trigger_data}}"}},
			//       {"webhook_id": "team-service", "name": "Assign Team", "request_params": {"profile_id": "{{.step_2.response.profile_id}}"}},
			//       {"webhook_id": "analytics-service", "name": "Track Signup", "request_params": {"user_id": "{{.trigger_data.user_id}}", "source": "{{.trigger_data.signup_source}}"}}
			//     ]
			//   }
			//
			// Example 3 - Content Publishing Pipeline:
			//   POST /api/execution-chains
			//   {
			//     "tenant_id": "media-company",
			//     "name": "Article Publishing Workflow",
			//     "trigger_event": "content.submitted",
			//     "steps": [
			//       {"webhook_id": "seo-service", "name": "SEO Optimization", "request_params": {"content": "{{.trigger_data.article_body}}"}},
			//       {"webhook_id": "image-service", "name": "Process Images", "request_params": {"images": "{{.trigger_data.images}}"}},
			//       {"webhook_id": "cdn-service", "name": "Upload to CDN", "request_params": {"optimized_content": "{{.step_1.response.optimized_content}}"}},
			//       {"webhook_id": "social-service", "name": "Post to Social", "request_params": {"article_url": "{{.step_3.response.cdn_url}}", "seo_tags": "{{.step_1.response.tags}}"}}
			//     ]
			//   }
			chains.POST("", r.executionChainController.CreateChain)

			// GET /api/execution-chains - Lists all execution chains for a tenant
			// Purpose: Retrieves all configured execution chains with pagination and filtering
			// Workflow: Client request → Tenant validation → Database query → Formatted response
			//
			// Example 1 - DevOps Dashboard Query:
			//   GET /api/execution-chains?tenant_id=ecommerce-store&page=1&limit=10&status=active
			//   Response: {
			//     "chains": [
			//       {"id": "chain-1", "name": "Order Processing", "trigger_event": "order.placed", "status": "active", "total_runs": 1250},
			//       {"id": "chain-2", "name": "Return Processing", "trigger_event": "return.initiated", "status": "active", "total_runs": 45},
			//       {"id": "chain-3", "name": "Inventory Sync", "trigger_event": "inventory.low", "status": "paused", "total_runs": 892}
			//     ],
			//     "total": 3, "page": 1, "limit": 10
			//   }
			//
			// Example 2 - Performance Analysis Query:
			//   GET /api/execution-chains?tenant_id=saas-platform&filter=high_volume&sort=execution_count
			//   Response: {
			//     "chains": [
			//       {"id": "user-onboard", "name": "User Onboarding", "avg_duration": "2.3s", "success_rate": "99.2%", "total_runs": 5420},
			//       {"id": "password-reset", "name": "Password Reset Flow", "avg_duration": "1.1s", "success_rate": "98.7%", "total_runs": 2100}
			//     ]
			//   }
			//
			// Example 3 - Audit and Compliance Report:
			//   GET /api/execution-chains?tenant_id=finance-dept&include_inactive=true&audit_mode=true
			//   Response: {
			//     "chains": [
			//       {"id": "payment-chain", "name": "Payment Processing", "compliance_level": "PCI-DSS", "last_audit": "2024-12-01", "status": "active"},
			//       {"id": "invoice-chain", "name": "Invoice Generation", "compliance_level": "SOX", "last_audit": "2024-11-15", "status": "active"},
			//       {"id": "old-billing", "name": "Legacy Billing", "compliance_level": "deprecated", "deactivated_date": "2024-10-01", "status": "inactive"}
			//     ]
			//   }
			chains.GET("", r.executionChainController.ListChains)

			// GET /api/execution-chains/:id - Gets details of a specific execution chain
			// Purpose: Retrieves complete configuration, steps, and metadata of a single execution chain
			// Workflow: Chain ID validation → Permission check → Database lookup → Step details → Response formatting
			//
			// Example 1 - Order Processing Chain Details:
			//   GET /api/execution-chains/order-processing-chain-uuid
			//   Response: {
			//     "id": "order-processing-chain-uuid",
			//     "name": "Complete Order Processing",
			//     "description": "Handles end-to-end order processing from payment to delivery",
			//     "trigger_event": "order.placed",
			//     "status": "active",
			//     "steps": [
			//       {"id": "step-1", "name": "Process Payment", "webhook_id": "payment-service", "step_order": 1, "max_retries": 3, "timeout": "30s"},
			//       {"id": "step-2", "name": "Update Inventory", "webhook_id": "inventory-service", "step_order": 2, "max_retries": 2, "timeout": "15s"},
			//       {"id": "step-3", "name": "Create Shipping Label", "webhook_id": "shipping-service", "step_order": 3, "max_retries": 2, "timeout": "20s"},
			//       {"id": "step-4", "name": "Send Confirmation", "webhook_id": "email-service", "step_order": 4, "max_retries": 1, "timeout": "10s"}
			//     ],
			//     "created_at": "2024-01-15T10:00:00Z", "updated_at": "2024-01-20T14:30:00Z"
			//   }
			//
			// Example 2 - User Onboarding Chain Configuration:
			//   GET /api/execution-chains/user-onboarding-uuid
			//   Response: {
			//     "id": "user-onboarding-uuid",
			//     "name": "New User Onboarding",
			//     "trigger_event": "user.registered",
			//     "steps": [
			//       {"name": "Welcome Email", "request_params": {"template": "welcome", "personalization": "{{.trigger_data.user_preferences}}"}},
			//       {"name": "Create Profile", "request_params": {"user_data": "{{.trigger_data}}", "default_settings": true}},
			//       {"name": "Analytics Tracking", "request_params": {"event": "user_onboarded", "properties": "{{.step_2.response.profile_data}}"}}
			//     ],
			//     "error_handling": {"on_failure": "continue", "notification_webhook": "admin-alerts-service"}
			//   }
			//
			// Example 3 - Content Publishing Pipeline Details:
			//   GET /api/execution-chains/content-pipeline-uuid
			//   Response: {
			//     "id": "content-pipeline-uuid",
			//     "name": "Article Publishing Workflow",
			//     "trigger_event": "content.submitted",
			//     "steps": [
			//       {"name": "SEO Optimization", "conditional_logic": {"execute_if": "{{.trigger_data.content_type}} == 'article'"}},
			//       {"name": "Image Processing", "parallel_execution": true, "request_params": {"batch_size": 10}},
			//       {"name": "CDN Upload", "depends_on": ["step-1", "step-2"], "cache_policy": "aggressive"},
			//       {"name": "Social Media Posting", "schedule_delay": "5m", "platforms": ["twitter", "linkedin", "facebook"]}
			//     ],
			//     "performance_metrics": {"avg_duration": "45s", "success_rate": "97.8%", "last_24h_runs": 156}
			//   }
			chains.GET("/:id", r.executionChainController.GetChain)

			// PUT /api/execution-chains/:id - Updates an existing execution chain
			// Purpose: Modifies chain configuration, steps, or settings with validation and versioning
			// Workflow: Request validation → Permission check → Configuration diff → Database update → Cache invalidation
			//
			// Example 1 - Add New Step to Order Processing:
			//   PUT /api/execution-chains/order-processing-chain-uuid
			//   {
			//     "name": "Enhanced Order Processing with Fraud Check",
			//     "steps": [
			//       {"webhook_id": "payment-service", "name": "Process Payment", "step_order": 1},
			//       {"webhook_id": "fraud-detection", "name": "Fraud Check", "step_order": 2, "request_params": {"transaction_id": "{{.step_1.response.transaction_id}}"}},
			//       {"webhook_id": "inventory-service", "name": "Update Inventory", "step_order": 3, "conditional": "{{.step_2.response.fraud_score < 0.3}}"},
			//       {"webhook_id": "shipping-service", "name": "Create Label", "step_order": 4},
			//       {"webhook_id": "email-service", "name": "Send Confirmation", "step_order": 5}
			//     ]
			//   }
			//
			// Example 2 - Update Error Handling Policy:
			//   PUT /api/execution-chains/user-onboarding-uuid
			//   {
			//     "steps": [
			//       {"name": "Welcome Email", "on_failure_action": "continue", "max_retries": 3},
			//       {"name": "Create Profile", "on_failure_action": "stop", "max_retries": 2, "failure_webhook": "admin-alert-service"},
			//       {"name": "Analytics Tracking", "on_failure_action": "continue", "max_retries": 1}
			//     ],
			//     "global_error_handling": {"notify_on_failure": true, "rollback_on_critical_failure": true}
			//   }
			//
			// Example 3 - Performance Optimization Update:
			//   PUT /api/execution-chains/content-pipeline-uuid
			//   {
			//     "name": "Optimized Content Pipeline",
			//     "steps": [
			//       {"name": "SEO Optimization", "timeout_seconds": 20, "priority": "high"},
			//       {"name": "Parallel Image Processing", "parallel_execution": true, "max_concurrent": 5, "timeout_seconds": 60},
			//       {"name": "CDN Upload", "batch_processing": true, "batch_size": 50},
			//       {"name": "Scheduled Social Posting", "execution_delay": "300s", "retry_policy": "exponential_backoff"}
			//     ],
			//     "performance_settings": {"max_execution_time": "10m", "resource_limits": {"cpu": "500m", "memory": "1Gi"}}
			//   }
			chains.PUT("/:id", r.executionChainController.UpdateChain)

			// DELETE /api/execution-chains/:id - Deletes an execution chain
			// Purpose: Safely removes an execution chain and all associated data
			// Workflow: Permission check → Active run validation → Cascade deletion → Cleanup → Audit logging
			//
			// Example 1 - Decommission Legacy Order Process:
			//   DELETE /api/execution-chains/legacy-order-process-uuid
			//   Pre-checks: Verify no active runs, archive 6 months of execution history
			//   Response: {
			//     "message": "Execution chain 'Legacy Order Process' successfully deleted",
			//     "archived_runs": 2847,
			//     "cleanup_summary": {
			//       "chain_definition_removed": true,
			//       "steps_removed": 5,
			//       "historical_data_archived": true,
			//       "dependent_configurations_updated": 2
			//     },
			//     "audit_trail_id": "audit-del-12345"
			//   }
			//
			// Example 2 - Remove Test Workflow:
			//   DELETE /api/execution-chains/test-user-flow-uuid?force=true
			//   Response: {
			//     "message": "Test execution chain 'User Flow Test' forcefully deleted",
			//     "warning": "Force deletion bypassed active run checks",
			//     "active_runs_terminated": 3,
			//     "cleanup_summary": {
			//       "immediate_cleanup": true,
			//       "cascade_deletion": true,
			//       "notification_sent": true
			//     }
			//   }
			//
			// Example 3 - Compliance-Driven Deletion:
			//   DELETE /api/execution-chains/gdpr-user-data-process-uuid
			//   Headers: {"Compliance-Reason": "GDPR Article 17 - Right to Erasure"}
			//   Response: {
			//     "message": "Execution chain deleted for compliance reasons",
			//     "compliance_verification": {
			//       "regulation": "GDPR Article 17",
			//       "data_retention_policy": "immediate_deletion",
			//       "verification_hash": "sha256:abc123...",
			//       "compliance_officer_notified": true
			//     },
			//     "deletion_certificate": "cert-gdpr-del-67890"
			//   }
			chains.DELETE("/:id", r.executionChainController.DeleteChain)

			// POST /api/execution-chains/:id/execute - Manually triggers an execution chain
			// Purpose: Starts immediate execution of a chain with custom trigger data
			// Workflow: Chain validation → Parameter injection → Async execution → Run tracking → Response
			//
			// Example 1 - Emergency Order Processing:
			//   POST /api/execution-chains/order-processing-uuid/execute
			//   {
			//     "trigger_data": {
			//       "order_id": "URGENT-ORD-2024-001",
			//       "customer_id": "VIP-CUSTOMER-789",
			//       "priority": "high",
			//       "total_amount": 2999.99,
			//       "payment_method": "corporate_account",
			//       "special_instructions": "VIP customer - expedite processing"
			//     },
			//     "execution_options": {"priority": "high", "bypass_rate_limits": true}
			//   }
			//   Response: {
			//     "run_id": "run-emergency-12345",
			//     "status": "running",
			//     "estimated_duration": "45s",
			//     "tracking_url": "/api/execution-chains/runs/run-emergency-12345"
			//   }
			//
			// Example 2 - Batch User Migration:
			//   POST /api/execution-chains/user-migration-uuid/execute
			//   {
			//     "trigger_data": {
			//       "batch_id": "migration-batch-2024-Q1",
			//       "users": [
			//         {"user_id": "user123", "email": "john@company.com", "department": "engineering"},
			//         {"user_id": "user456", "email": "jane@company.com", "department": "marketing"},
			//         {"user_id": "user789", "email": "bob@company.com", "department": "sales"}
			//       ],
			//       "migration_settings": {"preserve_permissions": true, "notify_users": false}
			//     },
			//     "execution_options": {"parallel_processing": true, "max_concurrent": 5}
			//   }
			//   Response: {
			//     "run_id": "run-migration-67890",
			//     "batch_info": {"total_users": 3, "estimated_duration": "2m 30s"},
			//     "progress_webhook": "https://api.company.com/migration-progress"
			//   }
			//
			// Example 3 - Testing and Validation:
			//   POST /api/execution-chains/content-pipeline-uuid/execute
			//   {
			//     "trigger_data": {
			//       "test_mode": true,
			//       "content_id": "test-article-001",
			//       "content_type": "blog_post",
			//       "author": "test-author",
			//       "content_body": "This is a test article for validation...",
			//       "images": ["test-image-1.jpg", "test-image-2.png"],
			//       "tags": ["test", "validation", "pipeline"]
			//     },
			//     "execution_options": {
			//       "dry_run": true,
			//       "validation_only": true,
			//       "skip_external_apis": true,
			//       "detailed_logging": true
			//     }
			//   }
			//   Response: {
			//     "run_id": "run-test-validation-999",
			//     "test_mode": true,
			//     "validation_results": {"config_valid": true, "all_webhooks_reachable": true},
			//     "dry_run_summary": "All steps would execute successfully"
			//   }
			chains.POST("/:id/execute", r.executionChainController.ExecuteChain)

			// GET /api/execution-chains/:id/runs - Lists execution history for a specific chain
			// Purpose: Retrieves paginated execution history with status and performance metrics
			// Workflow: Chain validation → Permission check → Database query → Metrics calculation → Response formatting
			//
			// Example 1 - Performance Monitoring Dashboard:
			//   GET /api/execution-chains/order-processing-uuid/runs?page=1&limit=20&status=completed&date_range=last_7_days
			//   Response: {
			//     "runs": [
			//       {
			//         "run_id": "run-001", "status": "completed", "duration": "42s", "started_at": "2024-01-15T10:30:00Z",
			//         "trigger_event": "order.placed", "steps_completed": 4, "success_rate": "100%",
			//         "performance_metrics": {"avg_step_duration": "10.5s", "total_api_calls": 8, "data_processed": "2.3KB"}
			//       },
			//       {
			//         "run_id": "run-002", "status": "completed", "duration": "38s", "started_at": "2024-01-15T11:45:00Z",
			//         "trigger_event": "order.placed", "steps_completed": 4, "success_rate": "100%",
			//         "performance_metrics": {"avg_step_duration": "9.5s", "total_api_calls": 8, "data_processed": "1.8KB"}
			//       }
			//     ],
			//     "summary": {"total_runs": 156, "success_rate": "98.7%", "avg_duration": "41.2s"},
			//     "pagination": {"page": 1, "limit": 20, "total_pages": 8}
			//   }
			//
			// Example 2 - Failure Analysis Report:
			//   GET /api/execution-chains/user-onboarding-uuid/runs?status=failed&include_errors=true&sort=latest_first
			//   Response: {
			//     "runs": [
			//       {
			//         "run_id": "run-failed-001", "status": "failed", "failed_at": "2024-01-14T16:20:00Z",
			//         "failed_step": "step-2-profile-creation", "error_type": "timeout", "retry_attempts": 3,
			//         "error_details": "Profile service timeout after 30s - database connection pool exhausted",
			//         "impact_analysis": {"users_affected": 1, "business_impact": "medium"}
			//       },
			//       {
			//         "run_id": "run-failed-002", "status": "failed", "failed_at": "2024-01-14T14:15:00Z",
			//         "failed_step": "step-1-welcome-email", "error_type": "service_unavailable", "retry_attempts": 2,
			//         "error_details": "Email service returned 503 - temporary maintenance mode",
			//         "recovery_action": "auto_retried_after_service_recovery"
			//       }
			//     ],
			//     "error_summary": {"total_failures": 12, "most_common_error": "timeout", "affected_steps": ["profile-creation", "email-service"]}
			//   }
			//
			// Example 3 - Business Intelligence and Trends:
			//   GET /api/execution-chains/content-pipeline-uuid/runs?analytics=true&group_by=date&date_range=last_30_days
			//   Response: {
			//     "runs": [
			//       {"date": "2024-01-15", "total_runs": 45, "success_rate": "97.8%", "avg_duration": "52s", "content_processed": "245 articles"},
			//       {"date": "2024-01-14", "total_runs": 52, "success_rate": "96.2%", "avg_duration": "48s", "content_processed": "267 articles"},
			//       {"date": "2024-01-13", "total_runs": 38, "success_rate": "98.7%", "avg_duration": "51s", "content_processed": "198 articles"}
			//     ],
			//     "trends": {
			//       "success_rate_trend": "+1.2% vs last month",
			//       "performance_trend": "-3.2s avg duration improvement",
			//       "volume_trend": "+15% content processing increase"
			//     },
			//     "recommendations": ["Consider scaling image processing step", "Optimize SEO service response times"]
			//   }
			chains.GET("/:id/runs", r.executionChainController.ListChainRuns)

			// GET /api/execution-chains/runs/:runId - Gets details of a specific chain execution
			// Purpose: Retrieves comprehensive execution details including step-by-step results
			// Workflow: Run ID validation → Permission check → Deep data fetch → Step analysis → Detailed response
			//
			// Example 1 - Successful Order Processing Investigation:
			//   GET /api/execution-chains/runs/run-order-success-12345
			//   Response: {
			//     "run_id": "run-order-success-12345",
			//     "chain_name": "Complete Order Processing",
			//     "status": "completed", "duration": "42.5s",
			//     "started_at": "2024-01-15T10:30:00Z", "completed_at": "2024-01-15T10:30:42Z",
			//     "trigger_data": {"order_id": "ORD-2024-001", "customer_id": "CUST-789", "total": 199.99},
			//     "step_executions": [
			//       {
			//         "step_name": "Process Payment", "status": "completed", "duration": "12.3s",
			//         "request_payload": {"amount": 199.99, "customer_id": "CUST-789", "payment_method": "card_ending_4242"},
			//         "response_data": {"payment_id": "PAY-ABC123", "status": "approved", "transaction_id": "TXN-XYZ789"},
			//         "http_status": 200, "attempt_count": 1
			//       },
			//       {
			//         "step_name": "Update Inventory", "status": "completed", "duration": "8.7s",
			//         "request_payload": {"payment_id": "PAY-ABC123", "items": [{"sku": "WIDGET-001", "quantity": 2}]},
			//         "response_data": {"inventory_updated": true, "new_stock_levels": {"WIDGET-001": 48}},
			//         "http_status": 200, "attempt_count": 1
			//       }
			//     ]
			//   }
			//
			// Example 2 - Failed User Onboarding Debug:
			//   GET /api/execution-chains/runs/run-onboard-failed-67890
			//   Response: {
			//     "run_id": "run-onboard-failed-67890",
			//     "chain_name": "New User Onboarding",
			//     "status": "failed", "duration": "125.2s",
			//     "started_at": "2024-01-14T16:15:00Z", "failed_at": "2024-01-14T16:17:05Z",
			//     "trigger_data": {"user_id": "USER-001", "email": "new_user@company.com", "signup_source": "website"},
			//     "step_executions": [
			//       {
			//         "step_name": "Welcome Email", "status": "completed", "duration": "3.2s",
			//         "request_payload": {"template": "welcome", "user_email": "new_user@company.com"},
			//         "response_data": {"email_sent": true, "message_id": "MSG-001"},
			//         "http_status": 200, "attempt_count": 1
			//       },
			//       {
			//         "step_name": "Create Profile", "status": "failed", "duration": "122s",
			//         "request_payload": {"user_data": {"user_id": "USER-001", "email": "new_user@company.com"}},
			//         "error_details": "Database connection timeout after 30s, retried 3 times",
			//         "http_status": 504, "attempt_count": 4,
			//         "troubleshooting": {"possible_causes": ["Database overload", "Network connectivity"], "suggested_actions": ["Check DB metrics", "Retry with backoff"]}
			//       }
			//     ],
			//     "failure_analysis": {"root_cause": "Database performance issue", "business_impact": "User registration incomplete"}
			//   }
			//
			// Example 3 - Content Pipeline Performance Analysis:
			//   GET /api/execution-chains/runs/run-content-perf-99999?include_metrics=true
			//   Response: {
			//     "run_id": "run-content-perf-99999",
			//     "chain_name": "Article Publishing Workflow",
			//     "status": "completed", "duration": "156.8s",
			//     "performance_breakdown": {
			//       "seo_optimization": {"duration": "15.2s", "cpu_usage": "45%", "memory_peak": "128MB"},
			//       "image_processing": {"duration": "89.5s", "cpu_usage": "78%", "memory_peak": "512MB", "images_processed": 12},
			//       "cdn_upload": {"duration": "32.1s", "bandwidth_used": "45MB", "upload_speed": "1.4MB/s"},
			//       "social_posting": {"duration": "20.0s", "api_calls": 6, "platforms": ["twitter", "linkedin", "facebook"]}
			//     },
			//     "resource_utilization": {"total_cpu_time": "4.2s", "peak_memory": "512MB", "network_io": "67MB"},
			//     "optimization_suggestions": [
			//       "Image processing is the bottleneck - consider parallel processing",
			//       "CDN upload speed can be improved with compression",
			//       "Social posting can be made asynchronous"
			//     ],
			//     "data_flow": {
			//       "input_size": "2.3MB", "intermediate_transformations": 4,
			//       "output_artifacts": {"optimized_article": "1.8MB", "processed_images": "12.4MB", "social_posts": "0.5KB"}
			//     }
			//   }
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
