#!/bin/bash

# Loki Suite Test Script
# This script demonstrates the complete webhook workflow

BASE_URL="http://localhost:8080/api/webhooks"

echo "🚀 Testing Loki Suite Webhook Service"
echo "======================================"

# Check if service is running
echo "1. Health Check..."
curl -s $BASE_URL/../health | jq '.' || echo "❌ Service not running"
echo ""

# Test 1: Generate a public webhook
echo "2. Generating Public Webhook..."
PUBLIC_RESPONSE=$(curl -s -X POST $BASE_URL/generate \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "test-tenant",
    "app_name": "test-app",
    "subscribed_event": "user.created",
    "type": "public"
  }')

echo "Public Webhook Response:"
echo $PUBLIC_RESPONSE | jq '.'
echo ""

# Test 2: Generate a private webhook  
echo "3. Generating Private Webhook..."
PRIVATE_RESPONSE=$(curl -s -X POST $BASE_URL/generate \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "test-tenant",
    "app_name": "secure-app",
    "subscribed_event": "order.completed", 
    "type": "private"
  }')

echo "Private Webhook Response:"
echo $PRIVATE_RESPONSE | jq '.'
echo ""

# Test 3: Manual subscription
echo "4. Manual Subscription..."
MANUAL_RESPONSE=$(curl -s -X POST $BASE_URL/subscribe \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "test-tenant",
    "app_name": "external-system",
    "target_url": "https://httpbin.org/post",
    "subscribed_event": "payment.processed",
    "type": "public"
  }')

echo "Manual Subscription Response:"
echo $MANUAL_RESPONSE | jq '.'
echo ""

# Test 4: List webhooks
echo "5. Listing Webhooks..."
LIST_RESPONSE=$(curl -s "$BASE_URL?tenant_id=test-tenant")
echo "Webhook List:"
echo $LIST_RESPONSE | jq '.'
echo ""

# Test 5: Send webhook event
echo "6. Sending Webhook Event..."
EVENT_RESPONSE=$(curl -s -X POST $BASE_URL/event \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "test-tenant",
    "event": "user.created",
    "source": "registration-service",
    "payload": {
      "user_id": "user-123",
      "email": "test@example.com",
      "name": "Test User"
    }
  }')

echo "Event Response:"
echo $EVENT_RESPONSE | jq '.'
echo ""

# Test 6: Test webhook verification endpoint
echo "7. Testing Webhook Verification..."
if [ ! -z "$PUBLIC_RESPONSE" ]; then
  WEBHOOK_ID=$(echo $PUBLIC_RESPONSE | jq -r '.webhook_id')
  SECRET_TOKEN=$(echo $PUBLIC_RESPONSE | jq -r '.secret_token')
  
  if [ "$WEBHOOK_ID" != "null" ] && [ "$SECRET_TOKEN" != "null" ]; then
    # Create test payload
    PAYLOAD='{"event":"test.event","source":"test","timestamp":"'$(date -Iseconds)'","payload":{"test":"data"}}'
    
    # Generate HMAC signature (requires openssl)
    SIGNATURE=$(echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "$SECRET_TOKEN" -binary | xxd -p -c 256)
    
    VERIFY_RESPONSE=$(curl -s -X POST "$BASE_URL/receive/$WEBHOOK_ID" \
      -H "Content-Type: application/json" \
      -H "X-Shavix-Signature: sha256=$SIGNATURE" \
      -H "X-Shavix-Timestamp: $(date -Iseconds)" \
      -d "$PAYLOAD")
    
    echo "Verification Response:"
    echo $VERIFY_RESPONSE | jq '.'
  else
    echo "❌ Could not extract webhook details for verification test"
  fi
else
  echo "❌ No webhook generated for verification test"
fi

echo ""
echo "✅ Test completed!"
echo "Check the responses above for any errors."
