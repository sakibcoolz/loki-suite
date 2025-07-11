# Loki Suite Webhook Service v2.0 - Complete API Documentation

## Table of Contents
1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Getting Started](#getting-started)
4. [Authentication & Security](#authentication--security)
5. [Webhook Management APIs](#webhook-management-apis)
6. [Execution Chain APIs](#execution-chain-apis)
7. [Event Processing](#event-processing)
8. [Error Handling](#error-handling)
9. [Examples & Use Cases](#examples--use-cases)
10. [Troubleshooting](#troubleshooting)

## Overview

The Loki Suite Webhook Service is a comprehensive webhook management platform that provides:

- **Multi-tenant webhook subscriptions** with JWT and HMAC security
- **Execution chains** for sequential webhook automation
- **Event-driven architecture** with reliable delivery
- **Retry mechanisms** with exponential backoff
- **Real-time monitoring** and status tracking

### Key Features

- ✅ **Dual Security**: JWT tokens + HMAC signatures
- ✅ **Multi-tenancy**: Complete tenant isolation
- ✅ **Execution Chains**: Sequential webhook workflows
- ✅ **Reliable Delivery**: Retry logic with status tracking
- ✅ **Event Processing**: Async event handling
- ✅ **RESTful APIs**: Complete CRUD operations
- ✅ **PostgreSQL Backend**: ACID compliance with JSONB support

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client Apps   │────│  Loki Suite API │────│   PostgreSQL    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                               │
                               ▼
                    ┌─────────────────┐
                    │ External Webhook│
                    │   Endpoints     │
                    └─────────────────┘
```

### Core Components

1. **Webhook Controller**: HTTP endpoint management
2. **Execution Chain Controller**: Workflow orchestration  
3. **Security Service**: JWT/HMAC authentication
4. **Repository Layer**: Data persistence abstraction
5. **Service Layer**: Business logic implementation

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 13+
- Environment variables configured

### Environment Setup

Create a `.env` file:

```bash
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
GIN_MODE=release

# Database Configuration
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=loki_suite
DB_PORT=5432
DB_SSL_MODE=disable

# Webhook Configuration
WEBHOOK_BASE_URL=http://localhost:8080
WEBHOOK_TIMEOUT_SECONDS=30
WEBHOOK_MAX_RETRIES=3

# Security Configuration
JWT_SECRET=your-super-secret-jwt-key-here
TIMESTAMP_TOLERANCE_MINUTES=5
HMAC_KEY_LENGTH=32
JWT_TOKEN_EXPIRATION=24

# Logger Configuration
LOG_LEVEL=info
LOG_ENCODING=json
LOG_OUTPUT_PATH=stdout
```

### Database Setup

```sql
-- Create database
CREATE DATABASE loki_suite;

-- The application will auto-migrate tables on startup
-- Tables created: webhook_subscriptions, webhook_events, 
-- execution_chains, execution_chain_steps, execution_chain_runs, 
-- execution_chain_step_runs
```

### Starting the Service

```bash
# Build and run
go build -o loki-suite
./loki-suite

# Or run directly
go run main.go
```

The service will start on `http://localhost:8080`

## Authentication & Security

### Security Models

1. **Public Webhooks**: HMAC-256 signature verification only
2. **Private Webhooks**: JWT authentication + HMAC verification

### HMAC Signature Generation

```go
// Example HMAC signature generation (Go)
func generateHMACSignature(payload []byte, secret string) string {
    h := hmac.New(sha256.New, []byte(secret))
    h.Write(payload)
    return "sha256=" + hex.EncodeToString(h.Sum(nil))
}
```

### JWT Token Format

```json
{
  "sub": "webhook_id",
  "tenant_id": "tenant_123",
  "webhook_id": "webhook_uuid",
  "iat": 1641024000,
  "exp": 1641110400
}
```

### Request Headers

All webhook deliveries include:

```
Content-Type: application/json
X-Loki-Signature: sha256=<hmac_signature>
X-Loki-Timestamp: <unix_timestamp>
X-Loki-Event: <event_name>
Authorization: Bearer <jwt_token>  // For private webhooks only
```

## Webhook Management APIs

### 1. Generate Webhook Subscription

Creates a new webhook subscription with auto-generated credentials.

**Endpoint:** `POST /api/webhooks/generate`

**Request:**
```json
{
  "tenant_id": "tenant_123",
  "app_name": "payment_service",
  "subscribed_event": "payment.completed",
  "type": "private"
}
```

**Response:**
```json
{
  "webhook_url": "http://localhost:8080/api/webhooks/receive/550e8400-e29b-41d4-a716-446655440000",
  "secret_token": "generated_32_char_secret",
  "jwt_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "type": "private",
  "webhook_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Use Case:** Auto-generate secure webhook endpoints for internal services

### 2. Manual Webhook Subscription

Creates a webhook subscription for external endpoints.

**Endpoint:** `POST /api/webhooks/subscribe`

**Request:**
```json
{
  "tenant_id": "tenant_123",
  "app_name": "external_crm",
  "target_url": "https://api.example.com/webhooks/loki",
  "subscribed_event": "user.created",
  "type": "public"
}
```

**Response:**
```json
{
  "webhook_url": "https://api.example.com/webhooks/loki",
  "secret_token": "generated_32_char_secret",
  "type": "public",
  "webhook_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

**Use Case:** Subscribe external services to receive webhook events

### 3. Send Event

Triggers webhook deliveries to all subscribed endpoints.

**Endpoint:** `POST /api/webhooks/event`

**Request:**
```json
{
  "tenant_id": "tenant_123",
  "event": "order.completed",
  "source": "order_service",
  "payload": {
    "order_id": "ORD-12345",
    "amount": 99.99,
    "customer_id": "CUST-789",
    "items": [
      {
        "product_id": "PROD-001",
        "quantity": 2,
        "price": 49.99
      }
    ]
  }
}
```

**Response:**
```json
{
  "event_id": "770e8400-e29b-41d4-a716-446655440002",
  "total_sent": 3,
  "total_failed": 0,
  "webhooks": [
    {
      "webhook_id": "550e8400-e29b-41d4-a716-446655440000",
      "target_url": "http://localhost:8080/api/webhooks/receive/550e8400-e29b-41d4-a716-446655440000",
      "success": true,
      "response_code": 200,
      "attempt_count": 1
    },
    {
      "webhook_id": "660e8400-e29b-41d4-a716-446655440001",
      "target_url": "https://api.example.com/webhooks/loki",
      "success": true,
      "response_code": 200,
      "attempt_count": 1
    }
  ]
}
```

**Use Case:** Broadcast events to all subscribed services

### 4. List Webhooks

Retrieves all webhook subscriptions for a tenant.

**Endpoint:** `GET /api/webhooks?tenant_id=tenant_123&page=1&limit=10`

**Response:**
```json
{
  "webhooks": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "tenant_id": "tenant_123",
      "app_name": "payment_service",
      "target_url": "http://localhost:8080/api/webhooks/receive/550e8400-e29b-41d4-a716-446655440000",
      "subscribed_event": "payment.completed",
      "type": "private",
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

### 5. Receive Webhook (For Generated Webhooks)

Internal endpoint for services using generated webhook URLs.

**Endpoint:** `POST /api/webhooks/receive/:id`

**Headers:**
```
Content-Type: application/json
Authorization: Bearer <jwt_token>
X-Loki-Signature: sha256=<hmac_signature>
X-Loki-Timestamp: <unix_timestamp>
```

**Request:**
```json
{
  "custom_data": "value",
  "processed_at": "2024-01-15T10:30:00Z"
}
```

**Response:**
```json
{
  "message": "Webhook received successfully",
  "webhook_id": "550e8400-e29b-41d4-a716-446655440000",
  "received_at": "2024-01-15T10:30:00Z"
}
```

## Execution Chain APIs

### 1. Create Execution Chain

Creates a sequence of webhooks to be executed in order.

**Endpoint:** `POST /api/execution-chains`

**Request:**
```json
{
  "tenant_id": "tenant_123",
  "name": "E-commerce Order Processing",
  "description": "Complete order processing workflow",
  "trigger_event": "order.placed",
  "steps": [
    {
      "webhook_id": "payment-webhook-uuid",
      "name": "Process Payment",
      "description": "Charge customer payment method",
      "request_params": {
        "amount": "{{.trigger_data.order_amount}}",
        "currency": "USD",
        "customer_id": "{{.trigger_data.customer_id}}",
        "payment_method": "{{.trigger_data.payment_method}}"
      },
      "on_success_action": "continue",
      "on_failure_action": "stop",
      "max_retries": 3,
      "delay_seconds": 0
    },
    {
      "webhook_id": "inventory-webhook-uuid",
      "name": "Update Inventory",
      "description": "Reduce product inventory",
      "request_params": {
        "order_id": "{{.trigger_data.order_id}}",
        "items": "{{.trigger_data.items}}",
        "payment_id": "{{.step_1.response.payment_id}}"
      },
      "on_success_action": "continue",
      "on_failure_action": "stop",
      "max_retries": 2,
      "delay_seconds": 5
    },
    {
      "webhook_id": "shipping-webhook-uuid",
      "name": "Create Shipping Label",
      "description": "Generate shipping label and tracking",
      "request_params": {
        "order_id": "{{.trigger_data.order_id}}",
        "shipping_address": "{{.trigger_data.shipping_address}}",
        "items": "{{.trigger_data.items}}",
        "payment_confirmed": true
      },
      "on_success_action": "continue",
      "on_failure_action": "continue",
      "max_retries": 1,
      "delay_seconds": 10
    },
    {
      "webhook_id": "notification-webhook-uuid",
      "name": "Send Confirmation Email",
      "description": "Email order confirmation to customer",
      "request_params": {
        "customer_email": "{{.trigger_data.customer_email}}",
        "order_id": "{{.trigger_data.order_id}}",
        "tracking_number": "{{.step_3.response.tracking_number}}",
        "estimated_delivery": "{{.step_3.response.estimated_delivery}}"
      },
      "on_success_action": "continue",
      "on_failure_action": "continue",
      "max_retries": 2,
      "delay_seconds": 0
    }
  ]
}
```

**Response:**
```json
{
  "chain_id": "880e8400-e29b-41d4-a716-446655440003",
  "name": "E-commerce Order Processing",
  "trigger_event": "order.placed",
  "steps_count": 4,
  "status": "pending",
  "created_at": "2024-01-15T10:00:00Z"
}
```

### 2. Execute Chain

Manually executes an execution chain with trigger data.

**Endpoint:** `POST /api/execution-chains/:id/execute`

**Request:**
```json
{
  "trigger_data": {
    "order_id": "ORD-12345",
    "customer_id": "CUST-789",
    "customer_email": "customer@example.com",
    "order_amount": 149.99,
    "payment_method": "card_1234",
    "items": [
      {
        "product_id": "PROD-001",
        "quantity": 2,
        "price": 49.99
      },
      {
        "product_id": "PROD-002", 
        "quantity": 1,
        "price": 49.99
      }
    ],
    "shipping_address": {
      "name": "John Doe",
      "street": "123 Main St",
      "city": "Anytown",
      "state": "CA",
      "zip": "12345",
      "country": "US"
    }
  }
}
```

**Response:**
```json
{
  "run_id": "990e8400-e29b-41d4-a716-446655440004",
  "chain_id": "880e8400-e29b-41d4-a716-446655440003",
  "status": "running",
  "total_steps": 4,
  "started_at": "2024-01-15T10:30:00Z"
}
```

### 3. Get Chain Run Status

Retrieves detailed status of a chain execution.

**Endpoint:** `GET /api/execution-chains/runs/:runId`

**Response:**
```json
{
  "id": "990e8400-e29b-41d4-a716-446655440004",
  "chain_id": "880e8400-e29b-41d4-a716-446655440003",
  "tenant_id": "tenant_123",
  "status": "completed",
  "trigger_event": "order.placed",
  "trigger_data": {
    "order_id": "ORD-12345",
    "customer_id": "CUST-789",
    "order_amount": 149.99
  },
  "current_step": 4,
  "total_steps": 4,
  "started_at": "2024-01-15T10:30:00Z",
  "completed_at": "2024-01-15T10:32:15Z",
  "last_error": null,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:32:15Z",
  "chain": {
    "id": "880e8400-e29b-41d4-a716-446655440003",
    "name": "E-commerce Order Processing",
    "description": "Complete order processing workflow"
  },
  "step_runs": [
    {
      "id": "step-run-1-uuid",
      "run_id": "990e8400-e29b-41d4-a716-446655440004",
      "step_id": "step-1-uuid",
      "step_order": 1,
      "status": "sent",
      "request_payload": {
        "amount": 149.99,
        "currency": "USD",
        "customer_id": "CUST-789",
        "payment_method": "card_1234"
      },
      "response_code": 200,
      "response_body": {
        "payment_id": "PAY-XYZ789",
        "status": "completed",
        "transaction_id": "TXN-ABC123"
      },
      "attempt_count": 1,
      "last_error": null,
      "started_at": "2024-01-15T10:30:00Z",
      "completed_at": "2024-01-15T10:30:15Z",
      "step": {
        "id": "step-1-uuid",
        "name": "Process Payment",
        "description": "Charge customer payment method"
      }
    },
    {
      "id": "step-run-2-uuid",
      "run_id": "990e8400-e29b-41d4-a716-446655440004",
      "step_id": "step-2-uuid",
      "step_order": 2,
      "status": "sent",
      "request_payload": {
        "order_id": "ORD-12345",
        "items": [...],
        "payment_id": "PAY-XYZ789"
      },
      "response_code": 200,
      "response_body": {
        "inventory_updated": true,
        "reserved_items": [...]
      },
      "attempt_count": 1,
      "started_at": "2024-01-15T10:30:20Z",
      "completed_at": "2024-01-15T10:30:35Z",
      "step": {
        "id": "step-2-uuid",
        "name": "Update Inventory"
      }
    }
  ]
}
```

### 4. List Execution Chains

Retrieves all execution chains for a tenant.

**Endpoint:** `GET /api/execution-chains?tenant_id=tenant_123&page=1&limit=10`

**Response:**
```json
{
  "chains": [
    {
      "id": "880e8400-e29b-41d4-a716-446655440003",
      "tenant_id": "tenant_123",
      "name": "E-commerce Order Processing",
      "description": "Complete order processing workflow",
      "status": "pending",
      "trigger_event": "order.placed",
      "is_active": true,
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z",
      "steps": [
        {
          "id": "step-1-uuid",
          "chain_id": "880e8400-e29b-41d4-a716-446655440003",
          "step_order": 1,
          "webhook_id": "payment-webhook-uuid",
          "name": "Process Payment",
          "description": "Charge customer payment method"
        }
      ]
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10
}
```

### 5. List Chain Runs

Retrieves execution history for a specific chain.

**Endpoint:** `GET /api/execution-chains/:id/runs?page=1&limit=10`

**Response:**
```json
{
  "runs": [
    {
      "id": "990e8400-e29b-41d4-a716-446655440004",
      "chain_id": "880e8400-e29b-41d4-a716-446655440003",
      "status": "completed",
      "current_step": 4,
      "total_steps": 4,
      "started_at": "2024-01-15T10:30:00Z",
      "completed_at": "2024-01-15T10:32:15Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10
}
```

### 6. Update Execution Chain

Updates chain properties (name, description, active status).

**Endpoint:** `PUT /api/execution-chains/:id`

**Request:**
```json
{
  "name": "Updated Order Processing Chain",
  "description": "Enhanced order processing with new features",
  "is_active": true
}
```

**Response:**
```json
{
  "message": "Execution chain updated successfully"
}
```

### 7. Delete Execution Chain

Permanently deletes an execution chain and all its runs.

**Endpoint:** `DELETE /api/execution-chains/:id`

**Response:**
```json
{
  "message": "Execution chain deleted successfully"
}
```

## Event Processing

### Automatic Chain Triggering

When events are sent via `/api/webhooks/event`, the system automatically:

1. **Finds matching chains** based on `trigger_event`
2. **Validates tenant access** for security
3. **Starts chain execution** with event payload as trigger data
4. **Processes steps sequentially** according to chain configuration

### Template Variable Resolution

The system supports dynamic request generation using template variables:

- `{{.trigger_data.field_name}}` - Access original event data
- `{{.step_N.response.field_name}}` - Access response from step N  
- `{{.step_N.request.field_name}}` - Access request from step N

**Example Template:**
```json
{
  "order_id": "{{.trigger_data.order_id}}",
  "payment_status": "{{.step_1.response.status}}",
  "total_amount": "{{.trigger_data.amount}}"
}
```

**Resolved Result:**
```json
{
  "order_id": "ORD-12345",
  "payment_status": "completed", 
  "total_amount": 149.99
}
```

## Error Handling

### HTTP Status Codes

| Code | Description | When It Occurs |
|------|-------------|----------------|
| 200 | Success | Request completed successfully |
| 201 | Created | Resource created successfully |
| 202 | Accepted | Async operation started |
| 400 | Bad Request | Invalid request data |
| 401 | Unauthorized | Missing/invalid authentication |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Resource already exists |
| 422 | Unprocessable Entity | Valid JSON but business logic error |
| 500 | Internal Server Error | Server-side error |

### Error Response Format

```json
{
  "error": "validation_error",
  "message": "tenant_id is required",
  "code": 400
}
```

### Retry Logic

**Webhook Delivery Retries:**
- Initial attempt: Immediate
- Retry 1: 2 seconds delay
- Retry 2: 4 seconds delay  
- Retry 3: 8 seconds delay
- Maximum: 3 retries (configurable)

**Execution Chain Step Retries:**
- Configurable per step via `max_retries`
- Exponential backoff: 2^attempt seconds
- Continue on failure: Configurable via `on_failure_action`

## Examples & Use Cases

### Use Case 1: E-commerce Order Processing

**Scenario:** Process customer orders through payment, inventory, shipping, and notification.

**Setup:**
1. Create webhook subscriptions for each service
2. Create execution chain with 4 steps
3. Configure event trigger on `order.placed`

**Flow:**
```
Order Placed → Payment → Inventory → Shipping → Email Notification
```

### Use Case 2: User Onboarding Workflow

**Scenario:** Automated user registration process.

**Setup:**
```json
{
  "name": "User Onboarding",
  "trigger_event": "user.registered",
  "steps": [
    {
      "name": "Send Welcome Email",
      "webhook_id": "email-service-webhook",
      "request_params": {
        "template": "welcome",
        "to": "{{.trigger_data.email}}",
        "name": "{{.trigger_data.name}}"
      }
    },
    {
      "name": "Create User Profile",
      "webhook_id": "profile-service-webhook", 
      "request_params": {
        "user_id": "{{.trigger_data.user_id}}",
        "initial_preferences": "{{.trigger_data.preferences}}"
      }
    },
    {
      "name": "Setup Analytics",
      "webhook_id": "analytics-webhook",
      "request_params": {
        "user_id": "{{.trigger_data.user_id}}",
        "signup_source": "{{.trigger_data.source}}"
      }
    }
  ]
}
```

### Use Case 3: Content Moderation Pipeline

**Scenario:** Multi-stage content review process.

**Setup:**
```json
{
  "name": "Content Moderation",
  "trigger_event": "content.submitted",
  "steps": [
    {
      "name": "AI Content Scan", 
      "webhook_id": "ai-moderation-webhook",
      "on_failure_action": "continue"
    },
    {
      "name": "Human Review Queue",
      "webhook_id": "review-queue-webhook",
      "request_params": {
        "content_id": "{{.trigger_data.content_id}}",
        "ai_score": "{{.step_1.response.confidence_score}}",
        "flagged_issues": "{{.step_1.response.issues}}"
      }
    },
    {
      "name": "Publish Content",
      "webhook_id": "publishing-webhook",
      "request_params": {
        "content_id": "{{.trigger_data.content_id}}",
        "review_status": "{{.step_2.response.status}}"
      }
    }
  ]
}
```

### Use Case 4: Financial Transaction Processing

**Scenario:** Multi-step financial transaction with compliance checks.

**Setup:**
```json
{
  "name": "Financial Transaction",
  "trigger_event": "transaction.initiated",
  "steps": [
    {
      "name": "Fraud Detection",
      "webhook_id": "fraud-detection-webhook",
      "on_failure_action": "stop",
      "max_retries": 1
    },
    {
      "name": "Compliance Check",
      "webhook_id": "compliance-webhook",
      "on_failure_action": "stop"
    },
    {
      "name": "Process Payment",
      "webhook_id": "payment-processor-webhook"
    },
    {
      "name": "Update Account Balance",
      "webhook_id": "account-service-webhook",
      "request_params": {
        "account_id": "{{.trigger_data.account_id}}",
        "amount": "{{.trigger_data.amount}}",
        "transaction_id": "{{.step_3.response.transaction_id}}"
      }
    },
    {
      "name": "Send Receipt",
      "webhook_id": "notification-webhook",
      "on_failure_action": "continue"
    }
  ]
}
```

## Health Check & Monitoring

### Health Check Endpoint

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy",
  "service": "loki-suite-webhook-service",
  "version": "2.0.0",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Monitoring Best Practices

1. **Monitor webhook delivery success rates**
2. **Track execution chain completion times**
3. **Alert on failed chain executions**
4. **Monitor database connection health**
5. **Track API response times**

## Troubleshooting

### Common Issues

**1. Webhook Delivery Failures**
```
Problem: HTTP 401 Unauthorized
Solution: Check HMAC signature generation and JWT token validity
```

**2. Chain Execution Stuck**
```
Problem: Chain status remains "running"
Solution: Check individual step statuses and logs for errors
```

**3. Template Variable Resolution**
```
Problem: Variables not resolved (showing {{.trigger_data.field}})
Solution: Ensure field exists in trigger data and proper JSON structure
```

**4. Database Connection Issues**
```
Problem: "failed to connect to database"
Solution: Verify PostgreSQL is running and connection parameters
```

### Debug Mode

Enable debug logging by setting:
```bash
LOG_LEVEL=debug
GIN_MODE=debug
```

### API Testing with cURL

**Create Webhook:**
```bash
curl -X POST http://localhost:8080/api/webhooks/generate \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "test_tenant",
    "app_name": "test_app", 
    "subscribed_event": "test.event",
    "type": "public"
  }'
```

**Send Event:**
```bash
curl -X POST http://localhost:8080/api/webhooks/event \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "test_tenant",
    "event": "test.event",
    "source": "test_service",
    "payload": {"message": "Hello World"}
  }'
```

**Execute Chain:**
```bash
curl -X POST http://localhost:8080/api/execution-chains/{chain-id}/execute \
  -H "Content-Type: application/json" \
  -d '{
    "trigger_data": {
      "order_id": "test_order",
      "amount": 100.00
    }
  }'
```

---

## Conclusion

The Loki Suite Webhook Service provides a comprehensive solution for webhook management and workflow automation. With its robust security model, flexible execution chains, and reliable delivery mechanisms, it's designed to handle enterprise-scale webhook orchestration requirements.

For additional support or feature requests, please refer to the project repository or contact the development team.
