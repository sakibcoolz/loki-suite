# Loki Suite v2.0 - Complete API Examples

This document provides comprehensive examples for testing the **restructured** Loki Suite webhook service with **JWT token support** for private webhooks.

## 🏗️ New Architecture Features

- **Clean Architecture**: Handler → Controller → Service → Repository layers
- **JWT Tokens**: Private webhooks now use JWT tokens instead of simple auth tokens
- **Zap Logging**: Structured JSON logging with Zap
- **Enhanced Security**: Both HMAC signing AND JWT authentication for private webhooks
- **Better Error Handling**: Consistent error responses with proper HTTP status codes

## Prerequisites

- Loki Suite v2.0 service running on `http://localhost:8080`
- PostgreSQL database configured and running
- `curl` and `jq` installed for testing

## 🚀 Complete Workflow Examples

### 1. Health Check

```bash
curl -X GET http://localhost:8080/health | jq '.'
```

**Expected Response:**
```json
{
  "status": "healthy",
  "service": "loki-suite",
  "version": "2.0.0",
  "timestamp": "2025-07-11T10:30:00Z"
}
```

### 2. Generate Public Webhook

```bash
curl -X POST http://localhost:8080/api/webhooks/generate \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "acme-corp",
    "app_name": "inventory-system",
    "subscribed_event": "product.updated",
    "type": "public"
  }' | jq '.'
```

**Expected Response:**
```json
{
  "webhook_url": "https://loki-suite.shavix.com/api/webhooks/receive/550e8400-e29b-41d4-a716-446655440000",
  "secret_token": "a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890",
  "type": "public",
  "webhook_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 3. Generate Private Webhook (NEW: With JWT Token)

```bash
curl -X POST http://localhost:8080/api/webhooks/generate \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "acme-corp",
    "app_name": "billing-system",
    "subscribed_event": "payment.processed",
    "type": "private"
  }' | jq '.'
```

**Expected Response:**
```json
{
  "webhook_url": "https://loki-suite.shavix.com/api/webhooks/receive/660e8400-e29b-41d4-a716-446655440001",
  "secret_token": "b2c3d4e5f6789012345678901234567890123456789012345678901234567890a1",
  "jwt_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0ZW5hbnRfaWQiOiJhY21lLWNvcnAiLCJ3ZWJob29rX2lkIjoiNjYwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDAxIiwiYXBwX25hbWUiOiJiaWxsaW5nLXN5c3RlbSIsImlzcyI6Imxva2ktc3VpdGUiLCJzdWIiOiI2NjBlODQwMC1lMjliLTQxZDQtYTcxNi00NDY2NTU0NDAwMDEiLCJpYXQiOjE2ODkwNzIwMDAsImV4cCI6MTY4OTE1ODQwMCwibmJmIjoxNjg5MDcyMDAwfQ.signature",
  "type": "private",
  "webhook_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

### 4. Manual Webhook Subscription

```bash
curl -X POST http://localhost:8080/api/webhooks/subscribe \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "acme-corp",
    "app_name": "external-crm",
    "target_url": "https://crm.acme-corp.com/webhooks/shavix",
    "subscribed_event": "customer.created",
    "type": "public"
  }' | jq '.'
```

### 5. Send Webhook Event

```bash
curl -X POST http://localhost:8080/api/webhooks/event \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "acme-corp",
    "event": "product.updated",
    "source": "warehouse-management",
    "payload": {
      "product_id": "PROD-12345",
      "sku": "ABC-123",
      "name": "Premium Widget",
      "quantity": 250,
      "location": "Warehouse-A",
      "last_updated": "2025-07-11T10:30:00Z",
      "updated_by": "system"
    }
  }' | jq '.'
```

**Expected Response (Enhanced):**
```json
{
  "message": "Webhook event processed successfully",
  "data": {
    "event_id": "770e8400-e29b-41d4-a716-446655440002",
    "total_sent": 1,
    "total_failed": 0,
    "webhooks": [
      {
        "webhook_id": "550e8400-e29b-41d4-a716-446655440000",
        "target_url": "https://loki-suite.shavix.com/api/webhooks/receive/550e8400-e29b-41d4-a716-446655440000",
        "success": true,
        "response_code": 200,
        "attempt_count": 1
      }
    ]
  }
}
```

### 6. List All Webhooks for Tenant (Enhanced)

```bash
curl -X GET "http://localhost:8080/api/webhooks?tenant_id=acme-corp&page=1&limit=10" | jq '.'
```

**Expected Response:**
```json
{
  "webhooks": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "tenant_id": "acme-corp",
      "app_name": "inventory-system",
      "target_url": "https://loki-suite.shavix.com/api/webhooks/receive/550e8400-e29b-41d4-a716-446655440000",
      "subscribed_event": "product.updated",
      "type": "public",
      "retry_count": 0,
      "is_active": true,
      "created_at": "2025-07-11T10:25:00Z",
      "updated_at": "2025-07-11T10:25:00Z"
    }
  ],
  "total": 3,
  "page": 1,
  "limit": 10
}
```

### 7. Test Webhook Verification - Public Webhook

```bash
# First, get the webhook details from step 2
WEBHOOK_ID="550e8400-e29b-41d4-a716-446655440000"
SECRET_TOKEN="a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890"

# Create the payload
PAYLOAD='{
  "event": "product.updated",
  "source": "inventory-system",
  "timestamp": "'$(date -Iseconds)'",
  "payload": {
    "product_id": "PROD-12345",
    "changes": ["quantity", "location"]
  },
  "event_id": "550e8400-e29b-41d4-a716-446655440000"
}'

# Generate HMAC signature (requires openssl)
SIGNATURE=$(echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "$SECRET_TOKEN" -binary | xxd -p -c 256)

# Send verification request (PUBLIC webhook - only HMAC signature required)
curl -X POST "http://localhost:8080/api/webhooks/receive/$WEBHOOK_ID" \
  -H "Content-Type: application/json" \
  -H "X-Shavix-Signature: sha256=$SIGNATURE" \
  -H "X-Shavix-Timestamp: $(date -Iseconds)" \
  -d "$PAYLOAD" | jq '.'
```

### 8. Test Webhook Verification - Private Webhook (NEW)

```bash
# Use the private webhook details from step 3
PRIVATE_WEBHOOK_ID="660e8400-e29b-41d4-a716-446655440001"
PRIVATE_SECRET_TOKEN="b2c3d4e5f6789012345678901234567890123456789012345678901234567890a1"
JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0ZW5hbnRfaWQiOiJhY21lLWNvcnAiLCJ3ZWJob29rX2lkIjoiNjYwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDAxIiwiYXBwX25hbWUiOiJiaWxsaW5nLXN5c3RlbSIsImlzcyI6Imxva2ktc3VpdGUiLCJzdWIiOiI2NjBlODQwMC1lMjliLTQxZDQtYTcxNi00NDY2NTU0NDAwMDEiLCJpYXQiOjE2ODkwNzIwMDAsImV4cCI6MTY4OTE1ODQwMCwibmJmIjoxNjg5MDcyMDAwfQ.signature"

# Create the payload
PRIVATE_PAYLOAD='{
  "event": "payment.processed",
  "source": "billing-system",
  "timestamp": "'$(date -Iseconds)'",
  "payload": {
    "payment_id": "PAY-67890",
    "amount": 99.99,
    "currency": "USD"
  },
  "event_id": "660e8400-e29b-41d4-a716-446655440001"
}'

# Generate HMAC signature
PRIVATE_SIGNATURE=$(echo -n "$PRIVATE_PAYLOAD" | openssl dgst -sha256 -hmac "$PRIVATE_SECRET_TOKEN" -binary | xxd -p -c 256)

# Send verification request (PRIVATE webhook - requires both HMAC signature AND JWT token)
curl -X POST "http://localhost:8080/api/webhooks/receive/$PRIVATE_WEBHOOK_ID" \
  -H "Content-Type: application/json" \
  -H "X-Shavix-Signature: sha256=$PRIVATE_SIGNATURE" \
  -H "X-Shavix-Timestamp: $(date -Iseconds)" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d "$PRIVATE_PAYLOAD" | jq '.'
```

**Expected Response:**
```json
{
  "message": "Webhook received and verified successfully",
  "data": {
    "webhook_id": "660e8400-e29b-41d4-a716-446655440001",
    "timestamp": "2025-07-11T10:30:00Z"
  }
}
```

## 🔐 Enhanced Security Testing Examples

### Test Missing JWT Token for Private Webhook

```bash
curl -X POST "http://localhost:8080/api/webhooks/receive/$PRIVATE_WEBHOOK_ID" \
  -H "Content-Type: application/json" \
  -H "X-Shavix-Signature: sha256=$PRIVATE_SIGNATURE" \
  -H "X-Shavix-Timestamp: $(date -Iseconds)" \
  -d "$PRIVATE_PAYLOAD" | jq '.'
```

**Expected Response (401):**
```json
{
  "error": "webhook_verification_failed",
  "message": "authorization header is required for private webhooks",
  "code": 401
}
```

### Test Invalid JWT Token

```bash
curl -X POST "http://localhost:8080/api/webhooks/receive/$PRIVATE_WEBHOOK_ID" \
  -H "Content-Type: application/json" \
  -H "X-Shavix-Signature: sha256=$PRIVATE_SIGNATURE" \
  -H "X-Shavix-Timestamp: $(date -Iseconds)" \
  -H "Authorization: Bearer invalid-jwt-token" \
  -d "$PRIVATE_PAYLOAD" | jq '.'
```

**Expected Response (401):**
```json
{
  "error": "webhook_verification_failed",
  "message": "JWT token verification failed: token is malformed: token contains an invalid number of segments",
  "code": 401
}
```

### Test Invalid HMAC Signature

```bash
curl -X POST "http://localhost:8080/api/webhooks/receive/$WEBHOOK_ID" \
  -H "Content-Type: application/json" \
  -H "X-Shavix-Signature: sha256=invalid-signature" \
  -H "X-Shavix-Timestamp: $(date -Iseconds)" \
  -d "$PAYLOAD" | jq '.'
```

**Expected Response (401):**
```json
{
  "error": "webhook_verification_failed",
  "message": "HMAC signature verification failed",
  "code": 401
}
```

## 📊 Enhanced Client Examples

### Node.js Client with JWT Support

```javascript
const crypto = require('crypto');
const axios = require('axios');
const jwt = require('jsonwebtoken');

class LokiSuiteClientV2 {
  constructor(baseUrl) {
    this.baseUrl = baseUrl;
  }

  async generateWebhook(tenantId, appName, event, type = 'public') {
    const response = await axios.post(`${this.baseUrl}/api/webhooks/generate`, {
      tenant_id: tenantId,
      app_name: appName,
      subscribed_event: event,
      type: type
    });
    return response.data;
  }

  async sendEvent(tenantId, event, source, payload) {
    const response = await axios.post(`${this.baseUrl}/api/webhooks/event`, {
      tenant_id: tenantId,
      event: event,
      source: source,
      payload: payload
    });
    return response.data;
  }

  verifyWebhookHMAC(payload, signature, secret) {
    const expectedSignature = crypto
      .createHmac('sha256', secret)
      .update(payload)
      .digest('hex');
    
    return crypto.timingSafeEqual(
      Buffer.from(signature),
      Buffer.from(expectedSignature)
    );
  }

  verifyJWTToken(token, secret) {
    try {
      const decoded = jwt.verify(token, secret);
      return { valid: true, claims: decoded };
    } catch (error) {
      return { valid: false, error: error.message };
    }
  }

  // Full webhook verification for private webhooks
  verifyPrivateWebhook(payload, signature, jwtToken, hmacSecret, jwtSecret) {
    // Verify HMAC signature
    if (!this.verifyWebhookHMAC(payload, signature, hmacSecret)) {
      return { valid: false, error: 'HMAC verification failed' };
    }

    // Verify JWT token
    const jwtResult = this.verifyJWTToken(jwtToken, jwtSecret);
    if (!jwtResult.valid) {
      return { valid: false, error: `JWT verification failed: ${jwtResult.error}` };
    }

    return { valid: true, claims: jwtResult.claims };
  }
}

// Usage Example
const client = new LokiSuiteClientV2('http://localhost:8080');

async function example() {
  // Generate private webhook
  const webhook = await client.generateWebhook(
    'my-tenant',
    'my-app',
    'user.created',
    'private'
  );
  
  console.log('Generated private webhook:', {
    id: webhook.webhook_id,
    url: webhook.webhook_url,
    hasJWT: !!webhook.jwt_token
  });
  
  // Send event
  const result = await client.sendEvent(
    'my-tenant',
    'user.created',
    'auth-service',
    { user_id: '12345', email: 'user@example.com' }
  );
  
  console.log('Event processing result:', result);
}
```

### Python Client with JWT Support

```python
import hashlib
import hmac
import json
import jwt
import requests
from datetime import datetime

class LokiSuiteClientV2:
    def __init__(self, base_url):
        self.base_url = base_url
    
    def generate_webhook(self, tenant_id, app_name, event, webhook_type='public'):
        response = requests.post(f'{self.base_url}/api/webhooks/generate', json={
            'tenant_id': tenant_id,
            'app_name': app_name,
            'subscribed_event': event,
            'type': webhook_type
        })
        return response.json()
    
    def send_event(self, tenant_id, event, source, payload):
        response = requests.post(f'{self.base_url}/api/webhooks/event', json={
            'tenant_id': tenant_id,
            'event': event,
            'source': source,
            'payload': payload
        })
        return response.json()
    
    def verify_webhook_hmac(self, payload, signature, secret):
        expected_signature = hmac.new(
            secret.encode(),
            payload.encode(),
            hashlib.sha256
        ).hexdigest()
        return hmac.compare_digest(signature, expected_signature)
    
    def verify_jwt_token(self, token, secret):
        try:
            decoded = jwt.decode(token, secret, algorithms=['HS256'])
            return {'valid': True, 'claims': decoded}
        except jwt.InvalidTokenError as e:
            return {'valid': False, 'error': str(e)}
    
    def verify_private_webhook(self, payload, signature, jwt_token, hmac_secret, jwt_secret):
        # Verify HMAC signature
        if not self.verify_webhook_hmac(payload, signature, hmac_secret):
            return {'valid': False, 'error': 'HMAC verification failed'}
        
        # Verify JWT token
        jwt_result = self.verify_jwt_token(jwt_token, jwt_secret)
        if not jwt_result['valid']:
            return {'valid': False, 'error': f"JWT verification failed: {jwt_result['error']}"}
        
        return {'valid': True, 'claims': jwt_result['claims']}

# Usage
client = LokiSuiteClientV2('http://localhost:8080')

# Generate private webhook
webhook = client.generate_webhook(
    'my-tenant',
    'my-app', 
    'order.completed',
    'private'
)
print(f"Generated private webhook: {webhook['webhook_id']}")
print(f"Has JWT token: {'jwt_token' in webhook}")

# Send event
result = client.send_event(
    'my-tenant',
    'order.completed',
    'ecommerce-platform',
    {'order_id': '67890', 'total': 99.99}
)
print(f"Event processing result: {result}")
```

## 🆕 New Features in v2.0

### 1. **Enhanced Security for Private Webhooks**
- **Dual Authentication**: Both HMAC signature AND JWT token required
- **JWT Claims Validation**: Webhook ID and tenant ID must match
- **Token Expiration**: Configurable JWT token expiration (default 24 hours)

### 2. **Improved Error Handling**
- **Consistent Error Format**: All errors return consistent JSON structure
- **HTTP Status Codes**: Proper HTTP status codes for different error types
- **Detailed Error Messages**: Clear, actionable error messages

### 3. **Enhanced Logging**
- **Structured Logging**: JSON-formatted logs with Zap
- **Request Tracing**: All HTTP requests logged with relevant metadata
- **Security Events**: Security-related events (failed verifications) logged

### 4. **Better API Responses**
- **Detailed Event Results**: Full webhook delivery results in event responses
- **Enhanced Pagination**: Page and limit information in list responses
- **Timestamp Information**: RFC3339 timestamps throughout

---

**💡 Pro Tips for v2.0:**

1. **Always store JWT tokens securely** - they contain sensitive webhook information
2. **Implement JWT token refresh** - tokens expire after 24 hours by default
3. **Use both HMAC and JWT verification** for private webhooks
4. **Monitor logs** - structured logging provides better debugging
5. **Validate JWT claims** - ensure webhook ID and tenant ID match expectations
6. **Handle token expiration gracefully** - implement token refresh workflows
7. **Use proper HTTP status codes** - leverage enhanced error responses for better error handling

**🔧 Migration from v1.0:**
- Private webhooks now use JWT tokens instead of simple auth tokens
- Response format enhanced with additional metadata
- Error responses now include HTTP status codes
- All timestamps are now in RFC3339 format
