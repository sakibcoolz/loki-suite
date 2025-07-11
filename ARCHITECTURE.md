# Loki Suite v2.0 - Architecture Documentation

## 🏗️ Overview

Loki Suite v2.0 has been completely restructured to follow **Clean Architecture** principles, providing better separation of concerns, improved testability, and enhanced maintainability. This document outlines the architectural decisions and structure of the new codebase.

## 📐 Clean Architecture Layers

### 1. Handler Layer (`internal/handler/`)
**Purpose**: HTTP routing and middleware configuration
- Route definitions and registration
- Middleware setup (CORS, logging, recovery)
- HTTP server configuration
- Entry point for HTTP requests

**Files:**
- `webhook_handler.go` - Route definitions and middleware setup

### 2. Controller Layer (`internal/controller/`)
**Purpose**: HTTP request/response handling and validation
- Request parsing and validation
- Response formatting and serialization
- HTTP status code management
- Error response formatting

**Files:**
- `webhook_controller.go` - HTTP controllers for webhook endpoints

### 3. Service Layer (`internal/service/`)
**Purpose**: Business logic and orchestration
- Core business rules implementation
- Security operations (JWT, HMAC)
- Event processing and distribution
- Cross-cutting concerns coordination

**Files:**
- `webhook_service.go` - Business logic for webhook operations

### 4. Repository Layer (`internal/repository/`)
**Purpose**: Data access and persistence
- Database operations (CRUD)
- Query optimization
- Transaction management
- Data mapping

**Files:**
- `webhook_repository.go` - Data access interface and implementation

### 5. Model Layer (`internal/models/`)
**Purpose**: Data structures and DTOs
- Domain models
- Data Transfer Objects (DTOs)
- Request/Response structures
- Database entity definitions

**Files:**
- `webhook.go` - Core webhook data models
- `dto.go` - Data Transfer Objects for API

## 🎯 Dependency Flow

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Handler   │───▶│ Controller  │───▶│   Service   │───▶│ Repository  │
│             │    │             │    │             │    │             │
│ • Routes    │    │ • Validation│    │ • Business  │    │ • Database  │
│ • Middleware│    │ • Responses │    │ • Security  │    │ • CRUD      │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

**Key Principles:**
- **Dependency Inversion**: Higher layers depend on interfaces, not implementations
- **Single Responsibility**: Each layer has a specific purpose
- **Interface Segregation**: Small, focused interfaces
- **Open/Closed**: Open for extension, closed for modification

## 🛠️ Supporting Packages (`pkg/`)

### `pkg/logger/`
**Purpose**: Structured logging with Zap
- JSON-formatted logging
- Configurable log levels
- Request tracing and correlation
- Performance logging

### `pkg/security/`
**Purpose**: Security utilities
- JWT token generation and verification
- HMAC signature operations
- Cryptographic functions
- Security validation helpers

### `pkg/database/`
**Purpose**: Database connection and management
- Connection pooling
- Migration management
- Health checks
- Configuration management

## 🔧 Configuration (`internal/config/`)

Centralized configuration management with:
- Environment variable loading
- Configuration validation
- Default value management
- Type-safe configuration structures

## 🔀 Request Flow

### 1. Webhook Generation Flow
```
HTTP POST /api/webhooks/generate
    ↓
Handler (routing)
    ↓
Controller (validation, parsing)
    ↓
Service (business logic, security)
    ↓
Repository (database save)
    ↓
Response (formatted JSON)
```

### 2. Event Processing Flow
```
HTTP POST /api/webhooks/event
    ↓
Handler (routing)
    ↓
Controller (validation, parsing)
    ↓
Service (find subscribers, send webhooks)
    ↓
Repository (query subscriptions, log events)
    ↓
Response (delivery status)
```

### 3. Webhook Verification Flow
```
HTTP POST /api/webhooks/receive/{id}
    ↓
Handler (routing)
    ↓
Controller (header extraction)
    ↓
Service (HMAC + JWT verification)
    ↓
Repository (lookup webhook details)
    ↓
Response (verification result)
```

## 🔐 Security Architecture

### Two-Tier Security Model

#### Public Webhooks
1. **HMAC-SHA256 Signature** verification
2. **Timestamp validation** (replay protection)

#### Private Webhooks (NEW in v2.0)
1. **HMAC-SHA256 Signature** verification
2. **JWT Token** authentication
3. **Claims validation** (webhook_id, tenant_id)
4. **Token expiration** checking

### Security Flow
```
Incoming Request
    ↓
Extract Headers (Signature, Timestamp, Authorization)
    ↓
Webhook Type Check (Public/Private)
    ↓
HMAC Verification (All webhooks)
    ↓
JWT Verification (Private webhooks only)
    ↓
Claims Validation (Private webhooks only)
    ↓
Success/Failure Response
```

## 📊 Data Architecture

### Database Schema Evolution

#### Webhook Subscriptions Table
```sql
CREATE TABLE webhook_subscriptions (
  id UUID PRIMARY KEY,
  tenant_id VARCHAR(255) NOT NULL,
  app_name VARCHAR(255) NOT NULL,
  target_url TEXT NOT NULL,
  subscribed_event VARCHAR(255) NOT NULL,
  secret_token VARCHAR(255) NOT NULL,
  auth_token VARCHAR(255),  -- NEW: JWT token for private webhooks
  type VARCHAR(50) NOT NULL DEFAULT 'public',
  retry_count INTEGER DEFAULT 0,
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### Webhook Events Table
```sql
CREATE TABLE webhook_events (
  id UUID PRIMARY KEY,
  tenant_id VARCHAR(255) NOT NULL,
  event VARCHAR(255) NOT NULL,
  source VARCHAR(255) NOT NULL,
  payload JSONB,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Repository Pattern Implementation

```go
type WebhookRepository interface {
    CreateSubscription(ctx context.Context, subscription *models.WebhookSubscription) error
    GetSubscriptionByID(ctx context.Context, id string) (*models.WebhookSubscription, error)
    GetSubscriptionsByTenantAndEvent(ctx context.Context, tenantID, event string) ([]*models.WebhookSubscription, error)
    CreateEvent(ctx context.Context, event *models.WebhookEvent) error
    // ... more methods
}
```

## 🧪 Testing Strategy

### Test Structure
```
pkg/
├── security/
│   └── security_test.go    # Security function unit tests
└── ...

internal/
├── config/
│   └── config_test.go      # Configuration loading tests
└── ...
```

### Test Coverage Areas
1. **Security Functions**: JWT, HMAC, validation
2. **Configuration Loading**: Environment variable handling
3. **Business Logic**: Service layer testing
4. **Data Access**: Repository layer testing
5. **HTTP Endpoints**: Controller integration tests

## 🚀 Performance Optimizations

### Database Optimizations
- **Connection Pooling**: Configured in `pkg/database/`
- **Query Optimization**: Efficient queries in repository layer
- **Indexing**: Proper database indexes for lookup operations

### Concurrency
- **Goroutines**: Non-blocking webhook delivery
- **Context Propagation**: Request cancellation and timeouts
- **Resource Management**: Proper cleanup and resource limits

### Logging Performance
- **Structured Logging**: Efficient JSON serialization with Zap
- **Log Levels**: Configurable verbosity
- **Async Logging**: Non-blocking log operations

## 📈 Monitoring and Observability

### Structured Logging
```json
{
  "level": "info",
  "ts": "2025-07-11T10:30:00.000Z",
  "caller": "controller/webhook_controller.go:45",
  "msg": "webhook generated successfully",
  "tenant_id": "acme-corp",
  "webhook_id": "550e8400-e29b-41d4-a716-446655440000",
  "webhook_type": "private",
  "request_id": "req-123456"
}
```

### Key Metrics to Monitor
- Request latency per endpoint
- Database query performance
- Webhook delivery success rates
- JWT token validation performance
- Error rates by type and endpoint

## 🔄 Migration Guide

### From v1.0 to v2.0

#### Code Changes
1. **Import Paths**: Update to new package structure
2. **Authentication**: Handle JWT tokens for private webhooks
3. **Error Handling**: Use new error response format
4. **Logging**: Adapt to structured JSON logs

#### Configuration Changes
```env
# NEW in v2.0
JWT_SECRET_KEY=your-jwt-secret-key-here
JWT_EXPIRATION_HOURS=24
LOG_LEVEL=info
LOG_FORMAT=json
SERVICE_VERSION=2.0.0
```

#### API Changes
- Private webhooks now return `jwt_token` in generation response
- Error responses include detailed error codes and messages
- Enhanced response metadata for better debugging

## 🔮 Future Enhancements

### Potential Improvements
1. **Metrics Endpoint**: Prometheus-compatible metrics
2. **Rate Limiting**: Request rate limiting per tenant
3. **Webhook Retry Logic**: Exponential backoff for failed deliveries
4. **Event Sourcing**: Complete audit trail of webhook events
5. **GraphQL API**: Alternative query interface
6. **Message Queues**: Async processing with Redis/RabbitMQ

### Scalability Considerations
1. **Horizontal Scaling**: Stateless design for easy scaling
2. **Database Sharding**: Tenant-based data partitioning
3. **Caching Layer**: Redis for frequently accessed data
4. **Load Balancing**: Multi-instance deployment support

---

## 📝 Architecture Decision Records (ADRs)

### ADR-001: Clean Architecture Adoption
**Decision**: Implement Clean Architecture with clear layer separation
**Rationale**: Improve maintainability, testability, and code organization
**Consequences**: Better separation of concerns, easier testing, more complex initial setup

### ADR-002: JWT Token Authentication for Private Webhooks
**Decision**: Add JWT tokens alongside HMAC for private webhooks
**Rationale**: Enhanced security, better claims validation, standard authentication
**Consequences**: Dual authentication complexity, improved security posture

### ADR-003: Zap Structured Logging
**Decision**: Replace standard logging with Zap structured logging
**Rationale**: Better observability, JSON format, performance benefits
**Consequences**: Enhanced monitoring capabilities, learning curve for developers

### ADR-004: Repository Pattern Implementation
**Decision**: Implement repository pattern for data access
**Rationale**: Database abstraction, easier testing, cleaner service layer
**Consequences**: Additional abstraction layer, improved testability

---

*This architecture supports the core requirements of security, scalability, and maintainability while providing a solid foundation for future enhancements.*
