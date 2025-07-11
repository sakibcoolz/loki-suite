# API Reference Guide

This document provides a complete reference for all Loki Suite Webhook Service APIs.

## Base URL

```
http://localhost:8080/api
```

## Authentication

All API requests require proper authentication. The service supports:
- **Bearer Token Authentication** for API access
- **HMAC Signature Verification** for webhook security
- **JWT Token Authentication** for private webhooks

---

## Webhook Management APIs

### Generate Webhook Subscription

Creates a new webhook subscription with auto-generated credentials.

```http
POST /api/webhooks/generate
```

**Request Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "tenant_id": "string",        // Required: Tenant identifier
  "app_name": "string",         // Required: Application name
  "subscribed_event": "string", // Required: Event name to subscribe to
  "type": "public|private"      // Required: Webhook type
}
```

**Response (201):**
```json
{
  "webhook_url": "string",      // Generated webhook URL
  "secret_token": "string",     // HMAC secret token
  "jwt_token": "string",        // JWT token (private webhooks only)
  "type": "public|private",     // Webhook type
  "webhook_id": "uuid"          // Unique webhook identifier
}
```

**Response (400):**
```json
{
  "error": "validation_error",
  "message": "tenant_id is required",
  "code": 400
}
```

---

### Manual Webhook Subscription

Creates a webhook subscription for external endpoints.

```http
POST /api/webhooks/subscribe
```

**Request Body:**
```json
{
  "tenant_id": "string",        // Required: Tenant identifier
  "app_name": "string",         // Required: Application name
  "target_url": "string",       // Required: External webhook URL
  "subscribed_event": "string", // Required: Event name to subscribe to
  "type": "public|private"      // Required: Webhook type
}
```

**Response (201):**
```json
{
  "webhook_url": "string",      // Target webhook URL
  "secret_token": "string",     // HMAC secret token
  "jwt_token": "string",        // JWT token (private webhooks only)
  "type": "public|private",     // Webhook type
  "webhook_id": "uuid"          // Unique webhook identifier
}
```

---

### Send Event

Triggers webhook deliveries to all subscribed endpoints.

```http
POST /api/webhooks/event
```

**Request Body:**
```json
{
  "tenant_id": "string",        // Required: Tenant identifier
  "event": "string",            // Required: Event name
  "source": "string",           // Required: Event source identifier
  "payload": {}                 // Required: Event payload (any JSON object)
}
```

**Response (200):**
```json
{
  "event_id": "uuid",           // Unique event identifier
  "total_sent": 3,              // Number of successful deliveries
  "total_failed": 0,            // Number of failed deliveries
  "webhooks": [                 // Delivery results per webhook
    {
      "webhook_id": "uuid",
      "target_url": "string",
      "success": true,
      "response_code": 200,
      "attempt_count": 1,
      "error": null
    }
  ]
}
```

---

### List Webhooks

Retrieves all webhook subscriptions for a tenant.

```http
GET /api/webhooks
```

**Query Parameters:**
- `tenant_id` (required): Tenant identifier
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 10, max: 100)
- `event` (optional): Filter by subscribed event
- `type` (optional): Filter by webhook type (public/private)
- `active` (optional): Filter by active status (true/false)

**Response (200):**
```json
{
  "webhooks": [
    {
      "id": "uuid",
      "tenant_id": "string",
      "app_name": "string",
      "target_url": "string",
      "subscribed_event": "string",
      "type": "public|private",
      "retry_count": 0,
      "is_active": true,
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10
}
```

---

### Receive Webhook

Internal endpoint for services using generated webhook URLs.

```http
POST /api/webhooks/receive/:id
```

**Path Parameters:**
- `id` (required): Webhook ID

**Request Headers:**
```
Content-Type: application/json
Authorization: Bearer <jwt_token>          // Required for private webhooks
X-Loki-Signature: sha256=<hmac_signature>  // Required for all webhooks
X-Loki-Timestamp: <unix_timestamp>         // Required for all webhooks
X-Loki-Event: <event_name>                 // Optional event identifier
```

**Request Body:**
```json
{
  // Any JSON payload
}
```

**Response (200):**
```json
{
  "message": "Webhook received successfully",
  "webhook_id": "uuid",
  "received_at": "2024-01-15T10:30:00Z"
}
```

**Response (401):**
```json
{
  "error": "unauthorized",
  "message": "Invalid JWT token or HMAC signature",
  "code": 401
}
```

---

## Execution Chain APIs

### Create Execution Chain

Creates a sequence of webhooks to be executed in order.

```http
POST /api/execution-chains
```

**Request Body:**
```json
{
  "tenant_id": "string",        // Required: Tenant identifier
  "name": "string",             // Required: Chain name
  "description": "string",      // Optional: Chain description
  "trigger_event": "string",    // Required: Event that triggers this chain
  "steps": [                    // Required: Array of chain steps (min: 1)
    {
      "webhook_id": "uuid",     // Required: Webhook to call
      "name": "string",         // Required: Step name
      "description": "string",  // Optional: Step description
      "request_params": {},     // Optional: Additional request parameters
      "on_success_action": "continue|stop|pause", // Optional: Action on success (default: continue)
      "on_failure_action": "continue|stop|retry", // Optional: Action on failure (default: stop)
      "max_retries": 3,         // Optional: Maximum retry attempts (default: 3)
      "delay_seconds": 0        // Optional: Delay before execution (default: 0)
    }
  ]
}
```

**Response (201):**
```json
{
  "chain_id": "uuid",
  "name": "string",
  "trigger_event": "string",
  "steps_count": 4,
  "status": "pending",
  "created_at": "2024-01-15T10:00:00Z"
}
```

---

### Get Execution Chain

Retrieves a specific execution chain.

```http
GET /api/execution-chains/:id
```

**Path Parameters:**
- `id` (required): Chain ID

**Response (200):**
```json
{
  "id": "uuid",
  "tenant_id": "string",
  "name": "string",
  "description": "string",
  "status": "pending|running|completed|failed|paused",
  "trigger_event": "string",
  "is_active": true,
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z",
  "steps": [
    {
      "id": "uuid",
      "chain_id": "uuid",
      "step_order": 1,
      "webhook_id": "uuid",
      "name": "string",
      "description": "string",
      "request_params": {},
      "on_success_action": "continue",
      "on_failure_action": "stop",
      "retry_count": 0,
      "max_retries": 3,
      "delay_seconds": 0,
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z",
      "webhook": {
        "id": "uuid",
        "app_name": "string",
        "target_url": "string",
        "type": "private"
      }
    }
  ]
}
```

---

### List Execution Chains

Retrieves all execution chains for a tenant.

```http
GET /api/execution-chains
```

**Query Parameters:**
- `tenant_id` (required): Tenant identifier
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 10, max: 100)
- `status` (optional): Filter by status
- `trigger_event` (optional): Filter by trigger event
- `active` (optional): Filter by active status

**Response (200):**
```json
{
  "chains": [
    {
      "id": "uuid",
      "tenant_id": "string",
      "name": "string",
      "description": "string",
      "status": "pending",
      "trigger_event": "string",
      "is_active": true,
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z",
      "steps": []
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10
}
```

---

### Update Execution Chain

Updates chain properties.

```http
PUT /api/execution-chains/:id
```

**Path Parameters:**
- `id` (required): Chain ID

**Request Body:**
```json
{
  "name": "string",             // Optional: New chain name
  "description": "string",      // Optional: New description
  "is_active": true             // Optional: Active status
}
```

**Response (200):**
```json
{
  "message": "Execution chain updated successfully"
}
```

---

### Delete Execution Chain

Permanently deletes an execution chain and all its runs.

```http
DELETE /api/execution-chains/:id
```

**Path Parameters:**
- `id` (required): Chain ID

**Response (200):**
```json
{
  "message": "Execution chain deleted successfully"
}
```

---

### Execute Chain

Manually executes an execution chain with trigger data.

```http
POST /api/execution-chains/:id/execute
```

**Path Parameters:**
- `id` (required): Chain ID

**Request Body:**
```json
{
  "trigger_data": {}            // Optional: Data to pass to chain execution
}
```

**Response (202):**
```json
{
  "run_id": "uuid",
  "chain_id": "uuid",
  "status": "running",
  "total_steps": 4,
  "started_at": "2024-01-15T10:30:00Z"
}
```

---

### Get Chain Run Status

Retrieves detailed status of a chain execution.

```http
GET /api/execution-chains/runs/:runId
```

**Path Parameters:**
- `runId` (required): Run ID

**Response (200):**
```json
{
  "id": "uuid",
  "chain_id": "uuid",
  "tenant_id": "string",
  "status": "pending|running|completed|failed|paused",
  "trigger_event": "string",
  "trigger_data": {},
  "current_step": 2,
  "total_steps": 4,
  "started_at": "2024-01-15T10:30:00Z",
  "completed_at": "2024-01-15T10:32:15Z",
  "last_error": null,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:32:15Z",
  "chain": {
    "id": "uuid",
    "name": "string",
    "description": "string"
  },
  "step_runs": [
    {
      "id": "uuid",
      "run_id": "uuid",
      "step_id": "uuid",
      "step_order": 1,
      "status": "pending|sent|failed",
      "request_payload": {},
      "response_code": 200,
      "response_body": {},
      "attempt_count": 1,
      "last_error": null,
      "started_at": "2024-01-15T10:30:00Z",
      "completed_at": "2024-01-15T10:30:15Z",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:15Z",
      "step": {
        "id": "uuid",
        "name": "string",
        "description": "string"
      }
    }
  ]
}
```

---

### List Chain Runs

Retrieves execution history for a specific chain.

```http
GET /api/execution-chains/:id/runs
```

**Path Parameters:**
- `id` (required): Chain ID

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 10, max: 100)
- `status` (optional): Filter by status
- `start_date` (optional): Filter runs after date (ISO 8601)
- `end_date` (optional): Filter runs before date (ISO 8601)

**Response (200):**
```json
{
  "runs": [
    {
      "id": "uuid",
      "chain_id": "uuid",
      "status": "completed",
      "current_step": 4,
      "total_steps": 4,
      "started_at": "2024-01-15T10:30:00Z",
      "completed_at": "2024-01-15T10:32:15Z",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:32:15Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10
}
```

---

## System APIs

### Health Check

Checks the health status of the service.

```http
GET /health
```

**Response (200):**
```json
{
  "status": "healthy",
  "service": "loki-suite-webhook-service",
  "version": "2.0.0",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Response (503):**
```json
{
  "status": "unhealthy",
  "service": "loki-suite-webhook-service",
  "version": "2.0.0",
  "timestamp": "2024-01-15T10:30:00Z",
  "errors": [
    "Database connection failed"
  ]
}
```

---

## Data Models

### WebhookSubscription

```json
{
  "id": "uuid",                 // Unique identifier
  "tenant_id": "string",        // Tenant identifier
  "app_name": "string",         // Application name
  "target_url": "string",       // Webhook URL
  "subscribed_event": "string", // Event name
  "type": "public|private",     // Webhook type
  "retry_count": 0,             // Current retry count
  "is_active": true,            // Active status
  "created_at": "timestamp",    // Creation timestamp
  "updated_at": "timestamp"     // Last update timestamp
}
```

### WebhookEvent

```json
{
  "id": "uuid",                 // Unique identifier
  "tenant_id": "string",        // Tenant identifier
  "event_name": "string",       // Event name
  "source": "string",           // Event source
  "payload": {},                // Event payload
  "status": "pending|sent|failed", // Delivery status
  "response_code": 200,         // HTTP response code
  "attempts": 1,                // Delivery attempts
  "last_error": null,           // Last error message
  "sent_at": "timestamp",       // Delivery timestamp
  "created_at": "timestamp",    // Creation timestamp
  "updated_at": "timestamp"     // Last update timestamp
}
```

### ExecutionChain

```json
{
  "id": "uuid",                 // Unique identifier
  "tenant_id": "string",        // Tenant identifier
  "name": "string",             // Chain name
  "description": "string",      // Chain description
  "status": "pending|running|completed|failed|paused", // Chain status
  "trigger_event": "string",    // Triggering event
  "is_active": true,            // Active status
  "created_at": "timestamp",    // Creation timestamp
  "updated_at": "timestamp",    // Last update timestamp
  "steps": []                   // Array of ExecutionChainStep
}
```

### ExecutionChainStep

```json
{
  "id": "uuid",                 // Unique identifier
  "chain_id": "uuid",           // Parent chain ID
  "step_order": 1,              // Execution order
  "webhook_id": "uuid",         // Target webhook ID
  "name": "string",             // Step name
  "description": "string",      // Step description
  "request_params": {},         // Request parameters
  "on_success_action": "continue|stop|pause", // Success action
  "on_failure_action": "continue|stop|retry", // Failure action
  "retry_count": 0,             // Current retry count
  "max_retries": 3,             // Maximum retries
  "delay_seconds": 0,           // Execution delay
  "created_at": "timestamp",    // Creation timestamp
  "updated_at": "timestamp"     // Last update timestamp
}
```

### ExecutionChainRun

```json
{
  "id": "uuid",                 // Unique identifier
  "chain_id": "uuid",           // Parent chain ID
  "tenant_id": "string",        // Tenant identifier
  "status": "pending|running|completed|failed|paused", // Run status
  "trigger_event": "string",    // Triggering event
  "trigger_data": {},           // Original trigger data
  "current_step": 2,            // Current step number
  "total_steps": 4,             // Total steps in chain
  "started_at": "timestamp",    // Start timestamp
  "completed_at": "timestamp",  // Completion timestamp
  "last_error": null,           // Last error message
  "created_at": "timestamp",    // Creation timestamp
  "updated_at": "timestamp"     // Last update timestamp
}
```

### ExecutionChainStepRun

```json
{
  "id": "uuid",                 // Unique identifier
  "run_id": "uuid",             // Parent run ID
  "step_id": "uuid",            // Step definition ID
  "step_order": 1,              // Step order in chain
  "status": "pending|sent|failed", // Step status
  "request_payload": {},        // Actual request sent
  "response_code": 200,         // HTTP response code
  "response_body": {},          // Response body
  "attempt_count": 1,           // Number of attempts
  "last_error": null,           // Last error message
  "started_at": "timestamp",    // Start timestamp
  "completed_at": "timestamp",  // Completion timestamp
  "created_at": "timestamp",    // Creation timestamp
  "updated_at": "timestamp"     // Last update timestamp
}
```

---

## Error Codes

### HTTP Status Codes

| Status | Code | Description |
|--------|------|-------------|
| OK | 200 | Request successful |
| Created | 201 | Resource created |
| Accepted | 202 | Async operation started |
| Bad Request | 400 | Invalid request |
| Unauthorized | 401 | Authentication failed |
| Forbidden | 403 | Access denied |
| Not Found | 404 | Resource not found |
| Conflict | 409 | Resource conflict |
| Unprocessable Entity | 422 | Business logic error |
| Internal Server Error | 500 | Server error |

### Error Response Format

```json
{
  "error": "error_code",        // Machine-readable error code
  "message": "string",          // Human-readable error message
  "code": 400                   // HTTP status code
}
```

### Common Error Codes

| Error Code | Description |
|------------|-------------|
| `validation_error` | Request validation failed |
| `unauthorized` | Authentication failed |
| `forbidden` | Access denied |
| `not_found` | Resource not found |
| `conflict` | Resource already exists |
| `invalid_signature` | HMAC signature verification failed |
| `invalid_token` | JWT token invalid or expired |
| `chain_not_found` | Execution chain not found |
| `webhook_not_found` | Webhook subscription not found |
| `execution_failed` | Chain execution failed |
| `database_error` | Database operation failed |
| `network_error` | Network communication failed |

---

## Rate Limiting

The API implements rate limiting to ensure fair usage:

- **Global Rate Limit**: 1000 requests per minute per IP
- **Per-Tenant Limit**: 500 requests per minute per tenant
- **Webhook Delivery**: 100 concurrent deliveries per tenant

Rate limit headers are included in responses:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1642262400
```

---

## Webhooks Payload Format

### Outgoing Webhook Payload

When the service delivers webhooks to external endpoints:

```json
{
  "event": "string",            // Event name
  "source": "string",           // Event source
  "timestamp": "string",        // ISO 8601 timestamp
  "payload": {},                // Event payload
  "event_id": "uuid"            // Unique event identifier
}
```

### Required Headers

```
Content-Type: application/json
X-Loki-Signature: sha256=<hmac_signature>
X-Loki-Timestamp: <unix_timestamp>
X-Loki-Event: <event_name>
Authorization: Bearer <jwt_token>  // For private webhooks only
```

---

## Template Variables

Execution chains support template variables for dynamic request generation:

### Available Variables

- `{{.trigger_data.field_name}}` - Access trigger data fields
- `{{.step_N.response.field_name}}` - Access response from step N
- `{{.step_N.request.field_name}}` - Access request from step N

### Example Template

```json
{
  "order_id": "{{.trigger_data.order_id}}",
  "customer_id": "{{.trigger_data.customer_id}}",
  "payment_id": "{{.step_1.response.payment_id}}",
  "shipping_cost": "{{.step_2.response.cost}}"
}
```

### Variable Resolution

Variables are resolved at execution time with proper type preservation (strings, numbers, booleans, objects).

---

This completes the comprehensive API reference guide for the Loki Suite Webhook Service.
