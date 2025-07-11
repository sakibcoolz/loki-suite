# Loki Suite Webhook Service v2.0

A comprehensive, enterprise-grade webhook management platform with execution chains for sequential webhook automation and workflow orchestration.

## 🚀 Overview

The Loki Suite Webhook Service is a robust solution for managing webhooks at scale, providing:

- **Multi-tenant webhook subscriptions** with complete tenant isolation
- **Dual security model** (JWT + HMAC) for maximum protection
- **Execution chains** for sequential webhook workflows and automation
- **Event-driven architecture** with reliable delivery guarantees
- **Advanced retry mechanisms** with exponential backoff
- **Real-time monitoring** and comprehensive status tracking
- **Template-based request generation** for dynamic workflows

## ✨ Key Features

### 🔐 Enterprise Security
- **Dual Authentication**: JWT tokens + HMAC signatures
- **Tenant Isolation**: Complete multi-tenancy support
- **Signature Verification**: SHA-256 HMAC validation
- **Token Management**: Automatic JWT generation and validation

### 🔄 Execution Chains
- **Sequential Processing**: Execute webhooks in defined order
- **Template Variables**: Dynamic request generation with `{{.trigger_data.field}}`
- **Error Handling**: Configurable retry logic and failure actions
- **Status Tracking**: Real-time monitoring of chain execution
- **Conditional Logic**: Continue, stop, or retry based on results

### 📡 Webhook Management
- **Auto-generation**: Create secure webhook endpoints instantly
- **External Integration**: Subscribe external services to events
- **Event Broadcasting**: Send events to all subscribed endpoints
- **Delivery Tracking**: Monitor success/failure rates
- **Retry Logic**: Automatic retries with exponential backoff

### 📊 Monitoring & Observability
- **Execution Metrics**: Track chain performance and completion times
- **Delivery Analytics**: Monitor webhook success rates
- **Error Reporting**: Detailed error messages and stack traces
- **Health Checks**: Service health monitoring endpoints

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client Apps   │────│  Loki Suite API │────│   PostgreSQL    │
│                 │    │                 │    │    Database     │
│ • E-commerce    │    │ • Webhook Mgmt  │    │                 │
│ • CRM Systems   │    │ • Chain Engine  │    │ • Subscriptions │
│ • Microservices │    │ • Security      │    │ • Events        │
│ • Third-party   │    │ • Monitoring    │    │ • Chain Runs    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                               │
                               ▼
                    ┌─────────────────┐
                    │ External Webhook│
                    │   Endpoints     │
                    │                 │
                    │ • Payment APIs  │
                    │ • Email Services│
                    │ • Analytics     │
                    │ • Notifications │
                    └─────────────────┘
```

### Core Components

1. **Webhook Controller**: HTTP endpoint management and request handling
2. **Execution Chain Controller**: Workflow orchestration and step management
3. **Security Service**: JWT/HMAC authentication and authorization
4. **Repository Layer**: Data persistence abstraction with GORM
5. **Service Layer**: Business logic implementation and orchestration

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [**Complete API Documentation**](COMPLETE_API_DOCUMENTATION.md) | Comprehensive end-to-end service documentation |
| [**API Reference Guide**](API_REFERENCE.md) | Detailed API endpoints and data models |
| [**Implementation Examples**](IMPLEMENTATION_EXAMPLES.md) | Real-world usage scenarios and code examples |
| [**Architecture Guide**](ARCHITECTURE.md) | System design and architectural decisions |

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+**
- **PostgreSQL 13+**
- **Git**

### 1. Clone and Setup

```bash
git clone <repository-url>
cd loki-suite
cp .env.example .env
```

### 2. Configure Environment

Edit `.env` file:

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

# Security Configuration
JWT_SECRET=your-super-secret-jwt-key-here
HMAC_KEY_LENGTH=32
JWT_TOKEN_EXPIRATION=24

# Webhook Configuration
WEBHOOK_BASE_URL=http://localhost:8080
WEBHOOK_TIMEOUT_SECONDS=30
WEBHOOK_MAX_RETRIES=3
```

### 3. Start Database

```bash
# Using Docker
docker run --name postgres-loki \
  -e POSTGRES_PASSWORD=your_password \
  -e POSTGRES_DB=loki_suite \
  -p 5432:5432 -d postgres:13

# Or use existing PostgreSQL installation
createdb loki_suite
```

### 4. Run the Service

```bash
# Build and run
go build -o loki-suite
./loki-suite

# Or run directly
go run main.go
```

The service will start on `http://localhost:8080`

### 5. Verify Installation

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
# {
#   "status": "healthy",
#   "service": "loki-suite-webhook-service",
#   "version": "2.0.0",
#   "timestamp": "2024-01-15T10:30:00Z"
# }
```

## 📖 Usage Examples

### Create a Webhook Subscription

```bash
curl -X POST http://localhost:8080/api/webhooks/generate \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "my_company",
    "app_name": "payment_service",
    "subscribed_event": "payment.completed",
    "type": "private"
  }'
```

### Send an Event

```bash
curl -X POST http://localhost:8080/api/webhooks/event \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "my_company",
    "event": "payment.completed",
    "source": "payment_gateway",
    "payload": {
      "order_id": "ORD-12345",
      "amount": 99.99,
      "status": "success"
    }
  }'
```

### Create an Execution Chain

```bash
curl -X POST http://localhost:8080/api/execution-chains \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "my_company",
    "name": "Order Processing Chain",
    "description": "Complete order processing workflow",
    "trigger_event": "order.placed",
    "steps": [
      {
        "webhook_id": "payment-webhook-uuid",
        "name": "Process Payment",
        "request_params": {
          "amount": "{{.trigger_data.order_amount}}",
          "customer_id": "{{.trigger_data.customer_id}}"
        },
        "max_retries": 3
      },
      {
        "webhook_id": "inventory-webhook-uuid", 
        "name": "Update Inventory",
        "request_params": {
          "order_id": "{{.trigger_data.order_id}}",
          "payment_id": "{{.step_1.response.payment_id}}"
        },
        "max_retries": 2
      }
    ]
  }'
```

## 🎯 Use Cases

### E-commerce Order Processing
```
Order Placed → Payment Processing → Inventory Update → Shipping Label → Email Confirmation
```

### User Onboarding Workflow
```
User Registered → Welcome Email → Profile Creation → Analytics Tracking → Setup Complete
```

### Content Moderation Pipeline
```
Content Submitted → AI Analysis → Human Review → Publishing → Notification
```

### Financial Transaction Processing
```
Transaction Initiated → Fraud Check → Compliance Verification → Payment Processing → Account Update
```

## 🔧 API Endpoints

### Webhook Management
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/webhooks/generate` | Generate webhook with auto credentials |
| `POST` | `/api/webhooks/subscribe` | Subscribe external webhook endpoint |
| `POST` | `/api/webhooks/event` | Send event to trigger webhooks |
| `POST` | `/api/webhooks/receive/:id` | Receive webhook (generated endpoints) |
| `GET` | `/api/webhooks` | List webhook subscriptions |

### Execution Chains
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/execution-chains` | Create new execution chain |
| `GET` | `/api/execution-chains` | List execution chains |
| `GET` | `/api/execution-chains/:id` | Get specific chain details |
| `PUT` | `/api/execution-chains/:id` | Update chain properties |
| `DELETE` | `/api/execution-chains/:id` | Delete execution chain |
| `POST` | `/api/execution-chains/:id/execute` | Execute chain manually |
| `GET` | `/api/execution-chains/runs/:runId` | Get run status and results |
| `GET` | `/api/execution-chains/:id/runs` | List chain execution history |

### System
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Service health check |

## 🔒 Security

### Authentication Methods

1. **Public Webhooks**: HMAC-256 signature verification
2. **Private Webhooks**: JWT authentication + HMAC verification

### HMAC Signature Generation

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
)

func generateSignature(payload []byte, secret string) string {
    h := hmac.New(sha256.New, []byte(secret))
    h.Write(payload)
    return "sha256=" + hex.EncodeToString(h.Sum(nil))
}
```

### Required Headers

```
Content-Type: application/json
X-Loki-Signature: sha256=<hmac_signature>
X-Loki-Timestamp: <unix_timestamp>
X-Loki-Event: <event_name>
Authorization: Bearer <jwt_token>  // For private webhooks
```

## 📊 Monitoring

### Key Metrics to Track

- **Webhook Delivery Success Rate**: Monitor delivery reliability
- **Execution Chain Completion Time**: Track workflow performance
- **Step Failure Rates**: Identify problematic integrations
- **API Response Times**: Monitor service performance
- **Database Connection Health**: Ensure data layer stability

### Health Monitoring

```bash
# Service health
curl http://localhost:8080/health

# Webhook performance
curl "http://localhost:8080/api/webhooks?tenant_id=my_company&status=failed"

# Chain execution metrics
curl "http://localhost:8080/api/execution-chains/runs/run-uuid"
```

## 🧪 Testing

### Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/service
```

### Integration Testing

```bash
# Start test environment
docker-compose up -d postgres
go run main.go &

# Run integration tests
go test -tags=integration ./tests/integration

# Cleanup
docker-compose down
```

## 🚀 Deployment

### Docker Deployment

```bash
# Build image
docker build -t loki-suite:latest .

# Run with Docker Compose
docker-compose up -d
```

### Production Configuration

```bash
# Production environment variables
export GIN_MODE=release
export LOG_LEVEL=info
export DB_SSL_MODE=require
export WEBHOOK_TIMEOUT_SECONDS=60
export WEBHOOK_MAX_RETRIES=5
```

## 🤝 Contributing

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Commit your changes**: `git commit -m 'Add amazing feature'`
4. **Push to the branch**: `git push origin feature/amazing-feature`
5. **Open a Pull Request**

### Development Setup

```bash
# Install development dependencies
go mod download

# Install linting tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Format code
go fmt ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: Check the comprehensive docs in this repository
- **Issues**: Create an issue for bugs or feature requests
- **Discussions**: Use GitHub Discussions for questions and ideas

## 🎯 Roadmap

- [ ] **GraphQL API**: Alternative query interface
- [ ] **WebSocket Support**: Real-time event streaming
- [ ] **Workflow Builder UI**: Visual chain creation interface
- [ ] **Advanced Analytics**: Detailed performance metrics dashboard
- [ ] **Plugin System**: Custom webhook transformations
- [ ] **Multi-region Deployment**: Geographic distribution support

---

**Built with ❤️ for developers who need reliable webhook orchestration at scale.**

## 🆕 What's New in v2.0

- **🏗️ Clean Architecture**: Proper separation of concerns with handler → controller → service → repository layers
- **🔐 Enhanced Security**: JWT tokens for private webhooks + HMAC signing for all webhooks
- **📊 Structured Logging**: Zap-based JSON logging with configurable levels
- **🛡️ Better Error Handling**: Consistent error responses with proper HTTP status codes
- **⚡ Improved Performance**: Connection pooling and optimized database operations
- **🔍 Enhanced Monitoring**: Detailed request tracing and security event logging

## ✨ Features

### Core Capabilities
- **Dual-Auth Webhooks**: Public (HMAC only) and Private (HMAC + JWT) webhook types
- **Event Distribution**: Send events to multiple subscribed webhooks with delivery tracking
- **Cryptographic Security**: HMAC-SHA256 signing + JWT token verification
- **Multi-tenant Support**: Complete isolation of webhook management per tenant
- **RESTful API**: Clean, intuitive API design with comprehensive validation

### Advanced Features
- **Clean Architecture**: Maintainable codebase with proper dependency injection
- **Structured Logging**: JSON-formatted logs with Zap for better observability
- **JWT Token Management**: Secure token generation with configurable expiration
- **Enhanced Error Handling**: Detailed error responses with actionable messages
- **Database Migrations**: Automatic GORM migrations with relationship management
- **Docker Support**: Multi-stage builds with health checks

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 13+
- Docker & Docker Compose (optional)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/your-username/loki-suite.git
   cd loki-suite
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Run with Docker Compose**
   ```bash
   docker-compose up -d
   ```

   Or run locally:
   ```bash
   go mod tidy
   go run main.go
   ```

4. **Test the service**
   ```bash
   curl http://localhost:8080/health
   ```

## 📚 API Reference

### Generate Public Webhook

```bash
POST /api/webhooks/generate
```

```json
{
  "tenant_id": "acme-corp",
  "app_name": "inventory-system",
  "subscribed_event": "product.updated",
  "type": "public"
}
```

**Response:**
```json
{
  "webhook_url": "https://loki-suite.shavix.com/api/webhooks/receive/550e8400-e29b-41d4-a716-446655440000",
  "secret_token": "a1b2c3d4e5f6...",
  "type": "public",
  "webhook_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Generate Private Webhook (NEW - With JWT Token)

```bash
POST /api/webhooks/generate
```

```json
{
  "tenant_id": "acme-corp",
  "app_name": "billing-system",
  "subscribed_event": "payment.processed",
  "type": "private"
}
```

**Response:**
```json
{
  "webhook_url": "https://loki-suite.shavix.com/api/webhooks/receive/660e8400-e29b-41d4-a716-446655440001",
  "secret_token": "b2c3d4e5f6...",
  "jwt_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "type": "private",
  "webhook_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

### Send Event

```bash
POST /api/webhooks/event
```

```json
{
  "tenant_id": "acme-corp",
  "event": "product.updated",
  "source": "warehouse-management",
  "payload": {
    "product_id": "PROD-12345",
    "quantity": 250
  }
}
```

### Receive Public Webhook

```bash
POST /api/webhooks/receive/{webhook_id}
```

**Headers:**
- `X-Shavix-Signature`: HMAC-SHA256 signature
- `X-Shavix-Timestamp`: RFC3339 timestamp

### Receive Private Webhook (NEW - Dual Authentication)

```bash
POST /api/webhooks/receive/{webhook_id}
```

**Headers:**
- `X-Shavix-Signature`: HMAC-SHA256 signature
- `X-Shavix-Timestamp`: RFC3339 timestamp
- `Authorization`: Bearer JWT-token

For complete API examples and security testing, see [API_EXAMPLES.md](API_EXAMPLES.md).

## 🔐 Enhanced Security

Loki Suite v2.0 implements multiple security layers:

### Public Webhooks
- **HMAC-SHA256 Signing**: Cryptographically secure signatures
- **Timestamp Validation**: Prevents replay attacks
- **Secret Token Management**: Unique secret per webhook

### Private Webhooks (NEW)
- **Dual Authentication**: Both HMAC signature AND JWT token required
- **JWT Claims Validation**: Webhook ID and tenant ID verification
- **Token Expiration**: Configurable JWT token expiration (default 24 hours)
- **Enhanced Logging**: Security events tracked for audit purposes

## 🏗️ Architecture

### Clean Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                        Loki Suite v2.0                     │
├─────────────────────────────────────────────────────────────┤
│  📥 Handler Layer (Routing & Middleware)                   │
│  • HTTP route definitions                                   │
│  • Middleware setup (CORS, logging, recovery)              │
├─────────────────────────────────────────────────────────────┤
│  🎮 Controller Layer (HTTP Request/Response)               │
│  • Request validation & parsing                             │
│  • Response formatting                                      │
│  • Error handling & status codes                           │
├─────────────────────────────────────────────────────────────┤
│  💼 Service Layer (Business Logic)                         │
│  • Webhook generation & management                          │
│  • Event processing & distribution                          │
│  • Security (JWT & HMAC operations)                        │
├─────────────────────────────────────────────────────────────┤
│  📊 Repository Layer (Data Access)                         │
│  • Database operations (CRUD)                              │
│  • Query optimization                                       │
│  • Transaction management                                   │
├─────────────────────────────────────────────────────────────┤
│  🗄️ Database Layer (PostgreSQL)                            │
│  • Webhook subscriptions                                    │
│  • Event storage & audit logs                              │
└─────────────────────────────────────────────────────────────┘
```

### Request Flow

```
HTTP Request → Handler → Controller → Service → Repository → Database
                ↓           ↓           ↓          ↓
            Middleware   Validation   Business   Data Access
                ↓           ↓         Logic        ↓
            Logging     Error         JWT/HMAC   Transactions
                        Handling      Security
```

## ⚙️ Configuration

### Environment Variables

```env
# Server Configuration
PORT=8080
GIN_MODE=release

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=loki_suite
DB_SSLMODE=disable

# Security Configuration
WEBHOOK_SECRET_KEY=your-secret-key-here
JWT_SECRET_KEY=your-jwt-secret-key-here
JWT_EXPIRATION_HOURS=24
WEBHOOK_TIMESTAMP_TOLERANCE_MINUTES=5

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json

# Service Configuration
SERVICE_NAME=loki-suite
SERVICE_VERSION=2.0.0
BASE_WEBHOOK_URL=https://loki-suite.shavix.com
```

## 🗄️ Database Schema

### Webhook Subscriptions (Enhanced)
```sql
CREATE TABLE webhook_subscriptions (
  id UUID PRIMARY KEY,
  tenant_id VARCHAR(255) NOT NULL,
  app_name VARCHAR(255) NOT NULL,
  target_url TEXT NOT NULL,
  subscribed_event VARCHAR(255) NOT NULL,
  secret_token VARCHAR(255) NOT NULL,
  auth_token VARCHAR(255), -- JWT token for private webhooks
  type VARCHAR(50) NOT NULL DEFAULT 'public',
  retry_count INTEGER DEFAULT 0,
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Webhook Events (Enhanced)
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

## 🧪 Development

### Project Structure

```
loki-suite/
├── cmd/                    # Application entry points
├── internal/               # Private application code
│   ├── config/            # Configuration management
│   ├── controller/        # HTTP controllers
│   ├── handler/           # HTTP handlers & routing
│   ├── middleware/        # HTTP middleware
│   ├── models/            # Data models & DTOs
│   ├── repository/        # Data access layer
│   └── service/           # Business logic layer
├── pkg/                   # Public packages
│   ├── database/          # Database utilities
│   ├── logger/            # Logging utilities
│   └── security/          # Security utilities (JWT/HMAC)
├── docker-compose.yml     # Local development setup
├── Dockerfile            # Container build definition
├── API_EXAMPLES.md       # Comprehensive API examples
└── README.md             # This file
```

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...

# Run specific test
go test -v ./pkg/security -run TestGenerateJWTToken
```

### Building

```bash
# Build for current platform
go build -o loki-suite .

# Build for Linux (production)
GOOS=linux GOARCH=amd64 go build -o loki-suite .
```

### Docker Build

```bash
# Build image
docker build -t loki-suite:2.0.0 .

# Run container
docker run -p 8080:8080 --env-file .env loki-suite:2.0.0
```

## 🚀 Deployment

### Docker Compose (Recommended)

```bash
# Development
docker-compose up -d

# Production
docker-compose -f docker-compose.prod.yml up -d
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: loki-suite-v2
  labels:
    app: loki-suite
    version: v2.0.0
spec:
  replicas: 3
  selector:
    matchLabels:
      app: loki-suite
  template:
    metadata:
      labels:
        app: loki-suite
        version: v2.0.0
    spec:
      containers:
      - name: loki-suite
        image: loki-suite:2.0.0
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: "postgres-service"
        - name: LOG_LEVEL
          value: "info"
        - name: LOG_FORMAT
          value: "json"
        envFrom:
        - secretRef:
            name: loki-suite-secrets
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## 📊 Performance & Monitoring

### Performance Optimizations in v2.0
- **Connection Pooling**: Efficient database connection management
- **Concurrent Processing**: Non-blocking webhook delivery with goroutines
- **Structured Logging**: Optimized JSON logging with Zap
- **Database Indexing**: Proper indexes for query optimization
- **Clean Architecture**: Reduced coupling for better performance

### Health Check

```bash
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "loki-suite",
  "version": "2.0.0",
  "timestamp": "2025-07-11T10:30:00Z"
}
```

### Structured Logging

All logs are now in structured JSON format:

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

## 🔄 Migration from v1.0

### Breaking Changes
- **Private webhooks** now use JWT tokens instead of simple auth tokens
- **API responses** include additional metadata and enhanced error information
- **Logging format** changed to structured JSON (configurable)
- **Environment variables** added for JWT configuration

### Migration Steps
1. Update environment variables with JWT configuration
2. Update client code to handle JWT tokens for private webhooks
3. Update error handling to use new error response format
4. Update log parsing if you process Loki Suite logs

## 🤝 Contributing

We welcome contributions! Please follow these guidelines:

1. **Fork** the repository
2. Create a **feature branch** (`git checkout -b feature/amazing-feature`)
3. **Follow clean architecture** principles
4. **Write comprehensive tests** for new features
5. **Update documentation** for API changes
6. **Ensure all tests pass** (`go test -v ./...`)
7. **Commit** your changes (`git commit -m 'feat: add amazing feature'`)
8. **Push** to the branch (`git push origin feature/amazing-feature`)
9. Open a **Pull Request**

### Development Guidelines

- **Clean Architecture**: Follow the established layer separation
- **Error Handling**: Use consistent error types and status codes
- **Logging**: Use structured logging with appropriate log levels
- **Testing**: Write unit tests for all layers
- **Security**: Validate all inputs and implement proper authentication
- **Documentation**: Update README and API examples for changes

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- 📧 **Email**: support@shavix.com
- 🐛 **Issues**: [GitHub Issues](https://github.com/your-username/loki-suite/issues)
- 📖 **Documentation**: [Wiki](https://github.com/your-username/loki-suite/wiki)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/your-username/loki-suite/discussions)

---

**Built with ❤️ and Clean Architecture by the Shavix Team**

*Loki Suite v2.0 - Secure, Scalable, and Production-Ready Webhook Infrastructure*

## ✨ Features

- **Multi-tenant webhook management**
- **Secure HMAC-SHA256 signing**
- **Public/Private webhook types**
- **Auto-generated webhook URLs**
- **Timestamp validation**
- **PostgreSQL with GORM**
- **RESTful API with Gin**
- **Docker support**

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 13+
- Docker & Docker Compose (optional)

### Installation

1. **Clone the repository**
```bash
git clone <repository-url>
cd loki-suite
```

2. **Install dependencies**
```bash
go mod download
```

3. **Setup environment**
```bash
cp .env.example .env
# Edit .env with your database credentials
```

4. **Run with Docker Compose**
```bash
docker-compose up -d
```

5. **Or run locally**
```bash
# Start PostgreSQL first
go run .
```

## 📚 API Documentation

### Base URL
```
http://localhost:8080/api/webhooks
```

### 1. Generate Webhook

**POST** `/api/webhooks/generate`

Generate a new webhook subscription with auto-generated URL and secrets.

```bash
curl -X POST http://localhost:8080/api/webhooks/generate \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant-xyz",
    "app_name": "smartcomm", 
    "subscribed_event": "product.updated",
    "type": "private"
  }'
```

**Response:**
```json
{
  "webhook_url": "https://loki-suite.shavix.com/api/webhooks/receive/123e4567-e89b-12d3-a456-426614174000",
  "secret_token": "a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890",
  "auth_token": "1a2b3c4d5e6f7890",
  "type": "private",
  "webhook_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

### 2. Manual Subscription

**POST** `/api/webhooks/subscribe`

Manually register a webhook with your own target URL.

```bash
curl -X POST http://localhost:8080/api/webhooks/subscribe \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant-xyz",
    "app_name": "erp",
    "target_url": "https://your-app.com/webhooks/shavix",
    "subscribed_event": "product.created",
    "type": "public"
  }'
```

### 3. Send Event

**POST** `/api/webhooks/event`

Trigger webhook events to all matching subscriptions.

```bash
curl -X POST http://localhost:8080/api/webhooks/event \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant-xyz",
    "event": "product.updated",
    "source": "erp",
    "payload": {
      "product_id": "P123",
      "name": "Updated Product",
      "price": 99.99
    }
  }'
```

### 4. Receive Webhook

**POST** `/api/webhooks/receive/{webhook_id}`

Endpoint for receiving and validating webhooks (used by auto-generated URLs).

```bash
curl -X POST http://localhost:8080/api/webhooks/receive/123e4567-e89b-12d3-a456-426614174000 \
  -H "Content-Type: application/json" \
  -H "X-Shavix-Signature: sha256=abc123..." \
  -H "X-Shavix-Timestamp: 2025-07-11T10:30:00Z" \
  -H "X-Shavix-Auth-Token: 1a2b3c4d5e6f7890" \
  -d '{
    "event": "product.updated",
    "source": "erp", 
    "timestamp": "2025-07-11T10:30:00Z",
    "payload": { "product_id": "P123" }
  }'
```

### 5. List Webhooks

**GET** `/api/webhooks?tenant_id=tenant-xyz`

List all webhook subscriptions for a tenant.

```bash
curl "http://localhost:8080/api/webhooks?tenant_id=tenant-xyz&page=1&limit=10"
```

## 🔒 Security Features

### HMAC Signature Verification

All webhooks are signed using HMAC-SHA256:

```
X-Shavix-Signature: sha256=<hex_signature>
```

### Timestamp Validation

Requests must include a timestamp header:

```
X-Shavix-Timestamp: 2025-07-11T10:30:00Z
```

Default tolerance: ±5 minutes

### Private Webhook Authentication

Private webhooks require an additional auth token:

```
X-Shavix-Auth-Token: <auth_token>
```

### Signature Validation Example (Go)

```go
func VerifyWebhook(body []byte, signature, secret string) bool {
    h := hmac.New(sha256.New, []byte(secret))
    h.Write(body)
    expectedSignature := hex.EncodeToString(h.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
```

## 🗃️ Database Schema

### webhook_subscriptions
```sql
CREATE TABLE webhook_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR NOT NULL,
    app_name VARCHAR NOT NULL,
    target_url VARCHAR NOT NULL,
    subscribed_event VARCHAR NOT NULL,
    type VARCHAR NOT NULL CHECK (type IN ('public', 'private')),
    secret_token VARCHAR NOT NULL,
    auth_token VARCHAR,
    retry_count INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_webhook_subscriptions_tenant_event 
ON webhook_subscriptions(tenant_id, subscribed_event);
```

### webhook_events
```sql
CREATE TABLE webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR NOT NULL,
    event_name VARCHAR NOT NULL,
    source VARCHAR NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed')),
    response_code INTEGER,
    attempts INTEGER DEFAULT 0,
    last_error TEXT,
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_webhook_events_tenant_status 
ON webhook_events(tenant_id, status);
```

## ⚙️ Configuration

### Environment Variables

```bash
# Database
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=loki_suite
DB_PORT=5432
DB_SSLMODE=disable

# Server
PORT=8080
GIN_MODE=release

# Webhook
WEBHOOK_BASE_URL=https://loki-suite.shavix.com
WEBHOOK_TIMEOUT_SECONDS=30
WEBHOOK_MAX_RETRIES=3

# Security
TIMESTAMP_TOLERANCE_MINUTES=5
```

## 🐳 Docker Deployment

### Build and Run
```bash
# Build image
docker build -t loki-suite .

# Run with Docker Compose
docker-compose up -d

# Check logs
docker-compose logs -f loki-suite
```

### Production Deployment
```bash
# Production environment
docker-compose -f docker-compose.prod.yml up -d
```

## 🧪 Testing

### Run Tests
```bash
go test ./...
```

### Manual Testing
```bash
# Health check
curl http://localhost:8080/health

# Generate webhook
curl -X POST http://localhost:8080/api/webhooks/generate \
  -H "Content-Type: application/json" \
  -d '{"tenant_id":"test","app_name":"test","subscribed_event":"test.event","type":"public"}'
```

## 📊 Monitoring

### Health Check
```bash
curl http://localhost:8080/health
```

### Metrics Endpoints
- `/health` - Basic health check
- Database connection status included in health check

## 🔧 Development

### Local Development Setup
```bash
# Install dependencies
go mod download

# Run with live reload (install air first)
go install github.com/cosmtrek/air@latest
air

# Format code
go fmt ./...

# Run linter
golangci-lint run
```

### Project Structure
```
loki-suite/
├── main.go           # Application entry point
├── models.go         # Database models and structs  
├── handlers.go       # HTTP handlers
├── security.go       # HMAC and security utilities
├── docker-compose.yml # Docker compose configuration
├── Dockerfile        # Docker build configuration
├── .env             # Environment variables
├── go.mod           # Go module dependencies
└── README.md        # Documentation
```

## 📝 API Examples

### Complete Workflow Example

1. **Generate a webhook:**
```bash
WEBHOOK_RESPONSE=$(curl -s -X POST http://localhost:8080/api/webhooks/generate \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "acme-corp",
    "app_name": "inventory-system",
    "subscribed_event": "product.updated", 
    "type": "private"
  }')

echo $WEBHOOK_RESPONSE
```

2. **Send an event:**
```bash
curl -X POST http://localhost:8080/api/webhooks/event \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "acme-corp",
    "event": "product.updated",
    "source": "warehouse-system",
    "payload": {
      "product_id": "PROD-001",
      "quantity": 150,
      "location": "Warehouse-A"
    }
  }'
```

3. **List webhooks:**
```bash
curl "http://localhost:8080/api/webhooks?tenant_id=acme-corp"
```

## 🛡️ Security Best Practices

1. **Always verify HMAC signatures**
2. **Validate timestamps to prevent replay attacks**
3. **Use HTTPS in production**
4. **Rotate secrets regularly**
5. **Monitor for suspicious activity**
6. **Rate limit webhook endpoints**

## 📄 License

MIT License - see LICENSE file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

---

**Built with ❤️ by ShaviX Technologies**
