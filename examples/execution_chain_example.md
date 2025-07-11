# Execution Chain API Examples

This document provides examples of how to use the execution chain feature to create sequences of webhooks that will be called in order with request parameters from the database.

## Overview

The execution chain feature allows you to:
- Create sequences of webhook calls that execute in order
- Pass data between webhook steps
- Monitor execution status and results
- Handle failures with retry logic

## API Endpoints

### 1. Create Execution Chain
```http
POST /api/execution-chains
Content-Type: application/json
Authorization: Bearer <your-jwt-token>

{
  "tenant_id": "your-tenant-id",
  "name": "Order Processing Chain",
  "description": "Process order through payment, inventory, and shipping",
  "is_active": true,
  "steps": [
    {
      "webhook_id": "webhook-payment-uuid",
      "name": "Process Payment",
      "order": 1,
      "retry_count": 3,
      "timeout_seconds": 30,
      "continue_on_failure": false,
      "request_template": {
        "amount": "{{.trigger_data.order_amount}}",
        "currency": "USD",
        "order_id": "{{.trigger_data.order_id}}"
      }
    },
    {
      "webhook_id": "webhook-inventory-uuid", 
      "name": "Update Inventory",
      "order": 2,
      "retry_count": 2,
      "timeout_seconds": 15,
      "continue_on_failure": false,
      "request_template": {
        "product_id": "{{.trigger_data.product_id}}",
        "quantity": "{{.trigger_data.quantity}}",
        "order_id": "{{.trigger_data.order_id}}"
      }
    },
    {
      "webhook_id": "webhook-shipping-uuid",
      "name": "Create Shipping Label", 
      "order": 3,
      "retry_count": 1,
      "timeout_seconds": 20,
      "continue_on_failure": true,
      "request_template": {
        "order_id": "{{.trigger_data.order_id}}",
        "address": "{{.trigger_data.shipping_address}}",
        "payment_confirmed": "{{.step_1.response.payment_id}}"
      }
    }
  ]
}
```

Response:
```json
{
  "chain_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Execution chain created successfully"
}
```

### 2. Execute Chain
```http
POST /api/execution-chains/550e8400-e29b-41d4-a716-446655440000/execute
Content-Type: application/json
Authorization: Bearer <your-jwt-token>

{
  "trigger_data": {
    "order_id": "ORD-12345",
    "order_amount": 99.99,
    "product_id": "PROD-ABC123",
    "quantity": 2,
    "shipping_address": {
      "street": "123 Main St",
      "city": "Anytown",
      "zip": "12345"
    }
  }
}
```

Response:
```json
{
  "run_id": "660e8400-e29b-41d4-a716-446655440001",
  "chain_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "running",
  "message": "Execution chain started successfully"
}
```

### 3. Get Chain Run Status
```http
GET /api/execution-chains/runs/660e8400-e29b-41d4-a716-446655440001
Authorization: Bearer <your-jwt-token>
```

Response:
```json
{
  "run_id": "660e8400-e29b-41d4-a716-446655440001",
  "chain_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "started_at": "2024-01-15T10:30:00Z",
  "completed_at": "2024-01-15T10:30:45Z",
  "trigger_data": {
    "order_id": "ORD-12345",
    "order_amount": 99.99
  },
  "step_runs": [
    {
      "step_id": "step-1-uuid",
      "step_name": "Process Payment",
      "status": "completed",
      "started_at": "2024-01-15T10:30:00Z",
      "completed_at": "2024-01-15T10:30:15Z",
      "request_data": {
        "amount": 99.99,
        "currency": "USD",
        "order_id": "ORD-12345"
      },
      "response_data": {
        "payment_id": "PAY-XYZ789",
        "status": "success"
      },
      "http_status": 200,
      "attempt_count": 1
    },
    {
      "step_id": "step-2-uuid", 
      "step_name": "Update Inventory",
      "status": "completed",
      "started_at": "2024-01-15T10:30:16Z",
      "completed_at": "2024-01-15T10:30:30Z",
      "request_data": {
        "product_id": "PROD-ABC123",
        "quantity": 2,
        "order_id": "ORD-12345"
      },
      "response_data": {
        "inventory_updated": true,
        "remaining_stock": 48
      },
      "http_status": 200,
      "attempt_count": 1
    }
  ]
}
```

### 4. List Execution Chains
```http
GET /api/execution-chains?tenant_id=your-tenant-id&page=1&limit=10
Authorization: Bearer <your-jwt-token>
```

### 5. List Chain Runs
```http
GET /api/execution-chains/550e8400-e29b-41d4-a716-446655440000/runs?page=1&limit=10
Authorization: Bearer <your-jwt-token>
```

## Template Variables

The execution chain supports template variables in request data:

- `{{.trigger_data.field_name}}` - Access trigger data fields
- `{{.step_N.response.field_name}}` - Access response from previous step N
- `{{.step_N.request.field_name}}` - Access request from previous step N

## Error Handling

- If `continue_on_failure` is false (default), the chain stops on step failure
- If `continue_on_failure` is true, the chain continues to next step
- Each step can have retry logic with configurable retry count
- Exponential backoff is used between retries

## Use Cases

1. **E-commerce Order Processing**: Payment → Inventory → Shipping → Notification
2. **User Registration**: Email Verification → Profile Creation → Welcome Email
3. **Data Pipeline**: Extract → Transform → Load → Notification
4. **Approval Workflow**: Request → Manager Approval → HR Approval → Final Processing

## Security

- All execution chain operations require JWT authentication
- Webhook URLs and secrets are stored securely in the database
- Request templates are sanitized to prevent injection attacks
- Tenant isolation ensures data privacy
