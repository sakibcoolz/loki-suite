# Practical Implementation Examples

This document provides real-world implementation examples for the Loki Suite Webhook Service.

## Example 1: Complete E-commerce Order Processing

### Step 1: Setup Webhook Subscriptions

First, create webhook subscriptions for each service in your e-commerce platform:

```bash
# Payment Service Webhook
curl -X POST http://localhost:8080/api/webhooks/generate \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "ecommerce_platform",
    "app_name": "payment_service",
    "subscribed_event": "payment.process",
    "type": "private"
  }'

# Response: Save the webhook_id for payment service
# {
#   "webhook_id": "payment-webhook-uuid",
#   "webhook_url": "http://localhost:8080/api/webhooks/receive/payment-webhook-uuid",
#   "secret_token": "payment_secret_token",
#   "jwt_token": "payment_jwt_token"
# }

# Inventory Service Webhook  
curl -X POST http://localhost:8080/api/webhooks/generate \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "ecommerce_platform",
    "app_name": "inventory_service", 
    "subscribed_event": "inventory.update",
    "type": "private"
  }'

# Shipping Service Webhook
curl -X POST http://localhost:8080/api/webhooks/generate \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "ecommerce_platform",
    "app_name": "shipping_service",
    "subscribed_event": "shipping.create",
    "type": "private"
  }'

# Email Service Webhook
curl -X POST http://localhost:8080/api/webhooks/generate \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "ecommerce_platform", 
    "app_name": "email_service",
    "subscribed_event": "email.send",
    "type": "private"
  }'
```

### Step 2: Create Execution Chain

```bash
curl -X POST http://localhost:8080/api/execution-chains \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "ecommerce_platform",
    "name": "Complete Order Processing",
    "description": "End-to-end order processing from payment to delivery notification",
    "trigger_event": "order.placed",
    "steps": [
      {
        "webhook_id": "payment-webhook-uuid",
        "name": "Process Payment",
        "description": "Charge customer payment method and verify transaction",
        "request_params": {
          "order_id": "{{.trigger_data.order_id}}",
          "customer_id": "{{.trigger_data.customer_id}}",
          "amount": "{{.trigger_data.total_amount}}",
          "currency": "{{.trigger_data.currency}}",
          "payment_method": "{{.trigger_data.payment_method}}",
          "billing_address": "{{.trigger_data.billing_address}}"
        },
        "on_success_action": "continue",
        "on_failure_action": "stop",
        "max_retries": 3,
        "delay_seconds": 0
      },
      {
        "webhook_id": "inventory-webhook-uuid",
        "name": "Reserve Inventory",
        "description": "Reserve products and update inventory levels",
        "request_params": {
          "order_id": "{{.trigger_data.order_id}}",
          "items": "{{.trigger_data.items}}",
          "payment_id": "{{.step_1.response.payment_id}}",
          "payment_status": "{{.step_1.response.status}}"
        },
        "on_success_action": "continue", 
        "on_failure_action": "stop",
        "max_retries": 2,
        "delay_seconds": 5
      },
      {
        "webhook_id": "shipping-webhook-uuid",
        "name": "Create Shipping Label",
        "description": "Generate shipping label and calculate delivery estimates",
        "request_params": {
          "order_id": "{{.trigger_data.order_id}}",
          "shipping_address": "{{.trigger_data.shipping_address}}",
          "items": "{{.trigger_data.items}}",
          "shipping_method": "{{.trigger_data.shipping_method}}",
          "payment_confirmed": true,
          "inventory_reserved": "{{.step_2.response.reserved}}"
        },
        "on_success_action": "continue",
        "on_failure_action": "continue",
        "max_retries": 2,
        "delay_seconds": 10
      },
      {
        "webhook_id": "email-webhook-uuid",
        "name": "Send Order Confirmation",
        "description": "Email order confirmation with tracking information",
        "request_params": {
          "template": "order_confirmation",
          "to": "{{.trigger_data.customer_email}}",
          "customer_name": "{{.trigger_data.customer_name}}",
          "order_id": "{{.trigger_data.order_id}}",
          "items": "{{.trigger_data.items}}",
          "total_amount": "{{.trigger_data.total_amount}}",
          "tracking_number": "{{.step_3.response.tracking_number}}",
          "estimated_delivery": "{{.step_3.response.estimated_delivery_date}}",
          "payment_id": "{{.step_1.response.payment_id}}"
        },
        "on_success_action": "continue",
        "on_failure_action": "continue", 
        "max_retries": 3,
        "delay_seconds": 0
      }
    ]
  }'
```

### Step 3: Trigger Order Processing

```bash
curl -X POST http://localhost:8080/api/webhooks/event \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "ecommerce_platform",
    "event": "order.placed",
    "source": "checkout_service",
    "payload": {
      "order_id": "ORDER-2024-001234",
      "customer_id": "CUST-789456",
      "customer_name": "Sarah Johnson",
      "customer_email": "sarah.johnson@email.com",
      "total_amount": 249.97,
      "currency": "USD",
      "payment_method": {
        "type": "credit_card",
        "last_four": "4242",
        "brand": "visa"
      },
      "items": [
        {
          "product_id": "PROD-LAPTOP-001", 
          "name": "Gaming Laptop",
          "quantity": 1,
          "price": 199.99,
          "sku": "LP-GAM-001"
        },
        {
          "product_id": "PROD-MOUSE-001",
          "name": "Wireless Mouse", 
          "quantity": 2,
          "price": 24.99,
          "sku": "MS-WIR-001"
        }
      ],
      "shipping_address": {
        "name": "Sarah Johnson",
        "street_1": "123 Technology Lane",
        "street_2": "Apt 4B",
        "city": "San Francisco",
        "state": "CA",
        "zip": "94105",
        "country": "US",
        "phone": "+1-555-0123"
      },
      "billing_address": {
        "name": "Sarah Johnson", 
        "street_1": "123 Technology Lane",
        "city": "San Francisco",
        "state": "CA",
        "zip": "94105",
        "country": "US"
      },
      "shipping_method": "standard",
      "order_date": "2024-01-15T14:30:00Z"
    }
  }'
```

### Expected Execution Flow

1. **Payment Step**: Process $249.97 charge
   - Request to payment service with order details
   - Response: `{"payment_id": "PAY-XYZ789", "status": "completed", "transaction_id": "TXN-ABC123"}`

2. **Inventory Step**: Reserve products
   - Request includes payment confirmation
   - Response: `{"reserved": true, "reservation_id": "RES-456", "items_reserved": [...]}`

3. **Shipping Step**: Create shipping label
   - Request includes confirmed payment and reservation
   - Response: `{"tracking_number": "1Z999AA1234567890", "estimated_delivery_date": "2024-01-18", "shipping_cost": 9.99}`

4. **Email Step**: Send confirmation
   - Request includes all previous step results
   - Response: `{"email_sent": true, "message_id": "MSG-EMAIL-001"}`

## Example 2: User Registration & Onboarding

### Setup Webhooks

```bash
# Email Service
curl -X POST http://localhost:8080/api/webhooks/subscribe \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "saas_platform",
    "app_name": "email_service",
    "target_url": "https://api.mailservice.com/webhooks/loki",
    "subscribed_event": "email.send",
    "type": "public"
  }'

# User Profile Service  
curl -X POST http://localhost:8080/api/webhooks/generate \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "saas_platform",
    "app_name": "profile_service",
    "subscribed_event": "profile.create", 
    "type": "private"
  }'

# Analytics Service
curl -X POST http://localhost:8080/api/webhooks/generate \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "saas_platform",
    "app_name": "analytics_service",
    "subscribed_event": "analytics.track",
    "type": "private"
  }'
```

### Create Onboarding Chain

```bash
curl -X POST http://localhost:8080/api/execution-chains \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "saas_platform",
    "name": "User Onboarding Workflow",
    "description": "Complete new user setup and welcome process",
    "trigger_event": "user.registered",
    "steps": [
      {
        "webhook_id": "email-service-webhook-uuid",
        "name": "Send Welcome Email",
        "description": "Send personalized welcome email to new user",
        "request_params": {
          "template": "welcome_new_user",
          "to": "{{.trigger_data.email}}",
          "variables": {
            "first_name": "{{.trigger_data.first_name}}",
            "company_name": "{{.trigger_data.company_name}}",
            "verification_link": "https://app.example.com/verify?token={{.trigger_data.verification_token}}"
          }
        },
        "on_success_action": "continue",
        "on_failure_action": "continue",
        "max_retries": 3,
        "delay_seconds": 0
      },
      {
        "webhook_id": "profile-service-webhook-uuid",
        "name": "Initialize User Profile",
        "description": "Create user profile with default settings",
        "request_params": {
          "user_id": "{{.trigger_data.user_id}}",
          "email": "{{.trigger_data.email}}",
          "first_name": "{{.trigger_data.first_name}}",
          "last_name": "{{.trigger_data.last_name}}",
          "company_name": "{{.trigger_data.company_name}}",
          "role": "{{.trigger_data.role}}",
          "subscription_tier": "{{.trigger_data.subscription_tier}}",
          "preferences": {
            "email_notifications": true,
            "marketing_emails": "{{.trigger_data.marketing_consent}}",
            "theme": "light",
            "timezone": "{{.trigger_data.timezone}}"
          }
        },
        "on_success_action": "continue",
        "on_failure_action": "stop",
        "max_retries": 2,
        "delay_seconds": 5
      },
      {
        "webhook_id": "analytics-service-webhook-uuid", 
        "name": "Track User Signup",
        "description": "Record user registration in analytics system",
        "request_params": {
          "event": "user_registered",
          "user_id": "{{.trigger_data.user_id}}",
          "properties": {
            "email": "{{.trigger_data.email}}",
            "company_name": "{{.trigger_data.company_name}}",
            "subscription_tier": "{{.trigger_data.subscription_tier}}",
            "signup_source": "{{.trigger_data.signup_source}}",
            "referrer": "{{.trigger_data.referrer}}",
            "profile_created": "{{.step_2.response.profile_id}}",
            "welcome_email_sent": "{{.step_1.response.message_id}}"
          },
          "timestamp": "{{.trigger_data.registration_timestamp}}"
        },
        "on_success_action": "continue",
        "on_failure_action": "continue",
        "max_retries": 1,
        "delay_seconds": 0
      }
    ]
  }'
```

### Trigger User Registration

```bash
curl -X POST http://localhost:8080/api/webhooks/event \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "saas_platform", 
    "event": "user.registered",
    "source": "registration_form",
    "payload": {
      "user_id": "USER-2024-556677",
      "email": "mike.developer@techcorp.com",
      "first_name": "Mike",
      "last_name": "Developer", 
      "company_name": "TechCorp Inc",
      "role": "developer",
      "subscription_tier": "professional",
      "verification_token": "verify_abc123xyz789",
      "marketing_consent": true,
      "timezone": "America/New_York",
      "signup_source": "website",
      "referrer": "google_ads",
      "registration_timestamp": "2024-01-15T15:45:00Z"
    }
  }'
```

## Example 3: Content Moderation Pipeline

### Setup Content Pipeline

```bash
# AI Moderation Service
curl -X POST http://localhost:8080/api/webhooks/subscribe \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "content_platform",
    "app_name": "ai_moderation",
    "target_url": "https://ai-service.example.com/moderate",
    "subscribed_event": "content.moderate",
    "type": "public"
  }'

# Human Review Queue
curl -X POST http://localhost:8080/api/webhooks/generate \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "content_platform",
    "app_name": "review_queue", 
    "subscribed_event": "review.queue",
    "type": "private"
  }'

# Publishing Service
curl -X POST http://localhost:8080/api/webhooks/generate \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "content_platform",
    "app_name": "publishing_service",
    "subscribed_event": "content.publish",
    "type": "private"
  }'
```

### Create Moderation Chain

```bash
curl -X POST http://localhost:8080/api/execution-chains \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "content_platform",
    "name": "Content Moderation Pipeline",
    "description": "Multi-stage content review and publishing process",
    "trigger_event": "content.submitted",
    "steps": [
      {
        "webhook_id": "ai-moderation-webhook-uuid",
        "name": "AI Content Analysis",
        "description": "Automated content analysis for policy violations",
        "request_params": {
          "content_id": "{{.trigger_data.content_id}}",
          "content_type": "{{.trigger_data.content_type}}",
          "content_text": "{{.trigger_data.content_text}}",
          "content_images": "{{.trigger_data.images}}",
          "author_id": "{{.trigger_data.author_id}}",
          "moderation_level": "strict"
        },
        "on_success_action": "continue",
        "on_failure_action": "continue",
        "max_retries": 2,
        "delay_seconds": 0
      },
      {
        "webhook_id": "review-queue-webhook-uuid",
        "name": "Human Review Queue",
        "description": "Queue content for human review if needed",
        "request_params": {
          "content_id": "{{.trigger_data.content_id}}",
          "ai_confidence_score": "{{.step_1.response.confidence_score}}",
          "ai_flags": "{{.step_1.response.policy_violations}}",
          "ai_recommendation": "{{.step_1.response.recommendation}}",
          "priority": "{{.step_1.response.priority}}",
          "content_preview": {
            "title": "{{.trigger_data.title}}",
            "author": "{{.trigger_data.author_name}}",
            "type": "{{.trigger_data.content_type}}"
          }
        },
        "on_success_action": "continue",
        "on_failure_action": "stop",
        "max_retries": 1,
        "delay_seconds": 10
      },
      {
        "webhook_id": "publishing-webhook-uuid",
        "name": "Publish Content",
        "description": "Publish approved content to platform",
        "request_params": {
          "content_id": "{{.trigger_data.content_id}}",
          "review_status": "{{.step_2.response.status}}",
          "reviewer_id": "{{.step_2.response.reviewer_id}}",
          "ai_score": "{{.step_1.response.confidence_score}}",
          "publish_immediately": "{{.step_2.response.auto_publish}}",
          "content_data": {
            "title": "{{.trigger_data.title}}",
            "body": "{{.trigger_data.content_text}}",
            "author_id": "{{.trigger_data.author_id}}",
            "category": "{{.trigger_data.category}}"
          }
        },
        "on_success_action": "continue",
        "on_failure_action": "stop",
        "max_retries": 2,
        "delay_seconds": 5
      }
    ]
  }'
```

## Example 4: Financial Transaction Processing

### Setup Financial Services

```bash
# Fraud Detection
curl -X POST http://localhost:8080/api/webhooks/subscribe \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "fintech_platform",
    "app_name": "fraud_detection",
    "target_url": "https://fraud-api.fintech.com/analyze",
    "subscribed_event": "transaction.analyze",
    "type": "public"
  }'

# Compliance Service
curl -X POST http://localhost:8080/api/webhooks/subscribe \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "fintech_platform", 
    "app_name": "compliance_service",
    "target_url": "https://compliance-api.fintech.com/check",
    "subscribed_event": "transaction.compliance",
    "type": "public"
  }'

# Payment Processor
curl -X POST http://localhost:8080/api/webhooks/generate \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "fintech_platform",
    "app_name": "payment_processor",
    "subscribed_event": "payment.process",
    "type": "private"
  }'
```

### Monitor Chain Execution

```bash
# Get real-time status of a running chain
curl -X GET "http://localhost:8080/api/execution-chains/runs/{run-id}" \
  -H "Content-Type: application/json"

# List all recent chain runs
curl -X GET "http://localhost:8080/api/execution-chains/{chain-id}/runs?page=1&limit=20" \
  -H "Content-Type: application/json"

# Get chain performance metrics
curl -X GET "http://localhost:8080/api/execution-chains?tenant_id=ecommerce_platform" \
  -H "Content-Type: application/json"
```

## Testing Webhook Endpoints

### Test Generated Webhook Endpoint

```bash
# Simulate external service calling your generated webhook
curl -X POST "http://localhost:8080/api/webhooks/receive/{webhook-id}" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {jwt_token}" \
  -H "X-Loki-Signature: sha256={hmac_signature}" \
  -H "X-Loki-Timestamp: $(date +%s)" \
  -H "X-Loki-Event: test.callback" \
  -d '{
    "callback_data": {
      "status": "completed",
      "result_id": "RESULT-12345",
      "processed_at": "2024-01-15T16:00:00Z"
    }
  }'
```

### Generate HMAC Signature (Python Example)

```python
import hmac
import hashlib
import json
import time

def generate_hmac_signature(payload, secret):
    """Generate HMAC signature for webhook verification"""
    payload_bytes = json.dumps(payload, separators=(',', ':')).encode('utf-8')
    signature = hmac.new(
        secret.encode('utf-8'),
        payload_bytes,
        hashlib.sha256
    ).hexdigest()
    return f"sha256={signature}"

# Example usage
payload = {
    "callback_data": {
        "status": "completed",
        "result_id": "RESULT-12345"
    }
}
secret = "your_webhook_secret_token"
signature = generate_hmac_signature(payload, secret)
print(f"X-Loki-Signature: {signature}")
```

## Performance Monitoring

### Key Metrics to Track

1. **Webhook Delivery Success Rate**
   ```bash
   # Monitor successful vs failed deliveries
   curl -X GET "http://localhost:8080/api/webhooks?tenant_id=your_tenant&status=failed"
   ```

2. **Execution Chain Completion Times**
   ```bash
   # Track chain execution performance
   curl -X GET "http://localhost:8080/api/execution-chains/{chain-id}/runs?limit=100"
   ```

3. **Step Failure Rates**
   ```bash
   # Analyze step-level failures
   curl -X GET "http://localhost:8080/api/execution-chains/runs/{run-id}"
   ```

### Setting Up Alerts

Monitor these conditions:
- Chain execution time > 5 minutes
- Step failure rate > 10%
- Webhook delivery failure rate > 5%
- Database connection issues
- High API response times

This completes the practical implementation examples showing real-world usage of the Loki Suite Webhook Service across different industries and use cases.
