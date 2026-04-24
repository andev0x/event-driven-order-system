# Event-driven Order System

A production-grade event-driven order management system built with Go, demonstrating scalable microservices architecture, asynchronous event processing, and distributed system design patterns.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Services](#services)
- [Getting Started](#getting-started)
- [API Reference](#api-reference)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## Overview

### Features

| Feature | Description |
|---------|-------------|
| Microservices Architecture | Independent services with well-defined responsibilities |
| Event-driven Communication | Asynchronous messaging via RabbitMQ for loose coupling |
| Database-per-Service | Isolated data stores ensuring service autonomy |
| Cache-Aside Pattern | Redis-based caching for optimized read performance |
| Clean Architecture | Layered design separating handlers, services, and repositories |
| Production Patterns | Structured logging, metrics, and comprehensive error handling |

### Technology Stack

| Component | Technology | Version |
|-----------|------------|---------|
| Language | Go | 1.21+ |
| Message Broker | RabbitMQ | 3.x |
| Cache | Redis | 7.x |
| Database | MySQL | 8.x |
| Containerization | Docker | Latest |
| Orchestration | Docker Compose | v2.10+ |

## Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                           External Client                           │
└───────────────────────────────┬─────────────────────────────────────┘
                                │ REST API
                                ▼
                     ┌──────────────────────┐
                     │    Order Service     │───── MySQL (order_db)
                     │       :8080          │───── Redis (cache)
                     └──────────┬───────────┘
                                │ Publish Events
                                ▼
                     ┌──────────────────────┐
                     │      RabbitMQ        │
                     │       :5672          │
                     └──────────┬───────────┘
                                │
              ┌─────────────────┴─────────────────┐
              │                                   │
              ▼                                   ▼
   ┌────────────────────┐            ┌────────────────────┐
   │  Analytics Service │            │ Notification Worker│
   │       :8081        │            │    (background)    │
   └────────────────────┘            └────────────────────┘
              │                               │
        MySQL + Redis                  Event Consumer
```

### Design Principles

1. **Service Decoupling** - Services communicate exclusively through events
2. **Database Isolation** - Each service owns and manages its data independently
3. **Asynchronous-First** - Event-driven design enables scalability and resilience
4. **Failure Isolation** - Service failures are contained and do not cascade
5. **Contract-driven** - Well-defined event schemas for inter-service communication

## Services

### Order Service

The primary entry point for order management operations.

**Responsibilities:**
- Order creation and validation
- Data persistence and caching
- Domain event publication

**Stack:** REST API, MySQL, Redis, RabbitMQ Producer

**Port:** `8080`

### Analytics Service

Processes order events and provides business intelligence.

**Responsibilities:**
- Event consumption and processing
- Metrics aggregation and storage
- Analytics query endpoints

**Stack:** RabbitMQ Consumer, MySQL, Redis, REST API

**Port:** `8081`

### Notification Worker

Background service for notification delivery.

**Responsibilities:**
- Event consumption
- Notification dispatch (email, SMS)
- Fan-out pattern demonstration

**Stack:** RabbitMQ Consumer

**Port:** Background process (no HTTP endpoint)

## Getting Started

### Prerequisites

- Docker & Docker Compose v2.10+
- Go 1.21+ (for local development)
- Make (optional)

### Quick Start

```bash
# Clone the repository
git clone https://github.com/andev0x/event-driven-order-system.git
cd event-driven-order-system

# Start all services
docker compose up --build

# Verify services are healthy (allow ~30 seconds for initialization)
docker compose ps
```

### Service Endpoints

| Service | URL | Credentials |
|---------|-----|-------------|
| Order Service | http://localhost:8080 | - |
| Analytics Service | http://localhost:8081 | - |
| Order Service Swagger UI | http://localhost:8080/swagger/index.html | - |
| Analytics Service Swagger UI | http://localhost:8081/swagger/index.html | - |
| Internal Auth Token (Order) | http://localhost:8080/internal/auth/token | `X-Internal-Auth-Key` header |
| Internal Auth Token (Analytics) | http://localhost:8081/internal/auth/token | `X-Internal-Auth-Key` header |
| RabbitMQ Management | http://localhost:15672 | guest / guest |
| Jaeger UI | http://localhost:16686 | - |
| Grafana | http://localhost:3000 | admin / admin |
| Prometheus | http://localhost:9090 | - |
| MySQL (Order DB) | localhost:3306 | orderuser / orderpass |
| MySQL (Analytics DB) | localhost:3307 | analyticsuser / analyticspass |
| Redis | localhost:6379 | - |

### Verify Installation

```bash
# Create a test order
make order

# Print raw JWT tokens (for scripting/manual requests)
make token-order
make token-analytics

# Retrieve analytics summary
make analytics

# View notification worker logs
make notification
```

### Shutdown

```bash
# Stop all services
docker compose down

# Stop and remove volumes (clean state)
docker compose down -v
```

## API Reference

### Order Service

#### Create Order

```http
POST /orders
Content-Type: application/json
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "customer_id": "customer-123",
  "product_id": "product-456",
  "quantity": 2,
  "total_amount": 99.99
}
```

**Response (201 Created):**
```json
{
  "id": "order-uuid-xxxx",
  "customer_id": "customer-123",
  "product_id": "product-456",
  "quantity": 2,
  "total_amount": 99.99,
  "status": "created",
  "created_at": "2026-01-09T12:34:56Z",
  "updated_at": "2026-01-09T12:34:56Z"
}
```

**cURL Example:**
```bash
TOKEN=$(make -s token-order)

curl -X POST http://localhost:8080/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"customer-123","product_id":"product-456","quantity":2,"total_amount":99.99}'
```

#### Get Order

```http
GET /orders/{order_id}
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "id": "order-uuid-xxxx",
  "customer_id": "customer-123",
  "product_id": "product-456",
  "quantity": 2,
  "total_amount": 99.99,
  "status": "created",
  "created_at": "2026-01-09T12:34:56Z",
  "updated_at": "2026-01-09T12:34:56Z"
}
```

### Analytics Service

#### Get Summary

```http
GET /analytics/summary
Authorization: Bearer <token>
```

#### Issue Internal Token

```http
POST /internal/auth/token
X-Internal-Auth-Key: <internal-auth-key>
Content-Type: application/json
```

**Request Body (optional):**
```json
{
  "subject": "internal-cli",
  "ttl_seconds": 3600
}
```

**Response (200 OK):**
```json
{
  "access_token": "jwt-token",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/internal/auth/token \
  -H "X-Internal-Auth-Key: ${INTERNAL_AUTH_KEY}" \
  -H "Content-Type: application/json" \
  -d '{"subject":"internal-cli","ttl_seconds":3600}'
```

**Response (200 OK):**
```json
{
  "total_orders": 42,
  "total_revenue": 12450.50,
  "average_order_value": 296.44,
  "last_updated": "2026-01-09T12:45:30Z"
}
```

### Error Responses

All endpoints return errors in a consistent format:

```json
{
  "error": "error_code",
  "message": "Human-readable description",
  "details": {}
}
```

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | `invalid_request` | Malformed request or validation failure |
| 404 | `not_found` | Requested resource does not exist |
| 500 | `internal_error` | Server-side error |

## Development

### Project Structure

```
event-driven-order-system/
├── services/
│   ├── order-service/
│   │   ├── cmd/order-api/          # Application entry point
│   │   ├── internal/
│   │   │   ├── api/                # HTTP handlers
│   │   │   ├── order/              # Domain logic
│   │   │   └── infrastructure/     # External dependencies
│   │   ├── migrations/             # Database schemas
│   │   └── Dockerfile
│   ├── analytics-service/          # Similar structure
│   └── notification-worker/        # Similar structure
├── pkg/                            # Shared packages
│   ├── config/                     # Configuration utilities
│   ├── database/                   # Database connections
│   ├── events/                     # Event definitions
│   ├── httputil/                   # HTTP utilities
│   ├── rabbitmq/                   # Message queue client
│   └── redis/                      # Cache client
├── docs/                           # Documentation
├── scripts/                        # Utility scripts
├── docker-compose.yml
├── Makefile
└── README.md
```

### Local Development

```bash
# Start infrastructure only
docker compose up -d order-db analytics-db redis rabbitmq

# Run services locally (in separate terminals)
cd services/order-service && go run cmd/order-api/main.go
cd services/analytics-service && go run cmd/analytics-api/main.go
cd services/notification-worker && go run cmd/notification-worker/main.go
```

### Make Commands

```bash
make help          # Display available commands
make tidy          # Tidy Go modules
make test          # Run all tests
make build         # Build Docker images
make build-go      # Build Go binaries
make swagger       # Generate Swagger docs
make up            # Start all services
make down          # Stop all services
make logs          # Stream service logs
make restart       # Restart all services
make clean         # Remove containers and volumes
make token-order   # Print Order Service JWT token
make token-analytics # Print Analytics Service JWT token
```

Swagger docs are generated into:

- `services/order-service/internal/api/docs`
- `services/analytics-service/internal/api/docs`

Regenerate docs after changing API annotations:

```bash
make swagger
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | localhost | Database host |
| `DB_PORT` | 3306 | Database port |
| `DB_USER` | - | Database username |
| `DB_PASSWORD` | - | Database password |
| `DB_NAME` | - | Database name |
| `REDIS_HOST` | localhost | Redis host |
| `REDIS_PORT` | 6379 | Redis port |
| `RABBITMQ_URL` | - | RabbitMQ connection URL |
| `JWT_SECRET` | - | Shared HMAC secret for API token validation |
| `INTERNAL_AUTH_KEY` | - | Key required by `/internal/auth/token` endpoint |
| `INTERNAL_AUTH_ISSUER` | order-service / analytics-service | Issuer claim for generated tokens |
| `INTERNAL_AUTH_TOKEN_TTL` | 1h | Default token lifetime |
| `SERVICE_PORT` | - | HTTP server port |

### Code Standards

This project adheres to Go best practices:

- **Architecture:** Domain-driven design with clean architecture layers
- **Error Handling:** Explicit error propagation without panics in production code
- **Naming:** Clear, descriptive identifiers following Go conventions
- **Documentation:** Package-level comments and exported function documentation
- **Testing:** Unit tests with mocked dependencies for business logic

## Testing

### Run Tests

```bash
# Run all tests
make test

# Run tests for a specific service
cd services/order-service && go test ./...

# Run with coverage report
go test -cover ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestOrderServiceCreate ./...
```

### Testing Strategy

| Type | Scope | Location |
|------|-------|----------|
| Unit | Business logic, service layer | `*_test.go` files |
| Integration | Database, cache, message queue | `tests/` directories |

## Deployment

### Docker Compose

```bash
docker compose up --build
```

### Manual Docker Build

```bash
# Build images
docker build -t order-service:latest services/order-service/
docker build -t analytics-service:latest services/analytics-service/
docker build -t notification-worker:latest services/notification-worker/

# Run container
docker run -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  -e RABBITMQ_URL=amqp://guest:guest@host.docker.internal:5672/ \
  order-service:latest
```

## Monitoring

### Structured Logging

All services emit structured JSON logs:

```json
{
  "level": "info",
  "service": "order-service",
  "event": "order_created",
  "order_id": "123",
  "timestamp": "2026-01-09T12:34:56Z"
}
```

### Metrics

Prometheus-format metrics are available at `/metrics`:

| Metric | Type | Description |
|--------|------|-------------|
| `orders_created_total` | Counter | Total orders created |
| `orders_created_duration_seconds` | Histogram | Order creation latency |
| `http_request_duration_seconds` | Histogram | HTTP request latency |
| `http_requests_total` | Counter | Total HTTP requests |

### Distributed Tracing

OpenTelemetry tracing is enabled in `order-service` and `analytics-service`:

- HTTP requests create/continue traces via Gorilla Mux OTel middleware.
- `order-service` injects trace context into RabbitMQ message headers (`traceparent` and baggage).
- `analytics-service` extracts headers and continues the trace when consuming events.
- Jaeger collects traces via OTLP gRPC (`jaeger:4317`) and exposes UI at `http://localhost:16686`.

### Grafana Dashboards

Grafana is pre-provisioned with Prometheus datasource and sample infrastructure dashboards:

- `MySQL Overview`
- `Redis Overview`
- `RabbitMQ Overview`

After `docker compose up --build`, open `http://localhost:3000` and navigate to Dashboards.
Run `make order` repeatedly and you should see queue depth, DB activity, and Redis command rate move.

### RabbitMQ Management

Monitor queues and message throughput at http://localhost:15672 (guest/guest).

### Health Monitoring

```bash
docker compose logs order-service
docker compose logs analytics-service
docker compose logs notification-worker
```

## Troubleshooting

### Service Connection Errors

```bash
# Verify all containers are healthy
docker compose ps

# Check service logs
docker compose logs [service-name]

# Test database connectivity
docker compose exec order-db mysql -u orderuser -p -e "SELECT 1;"

# Full restart
docker compose down -v && docker compose up --build
```

### Messages Not Processing

```bash
# Check RabbitMQ queue status
curl -u guest:guest http://localhost:15672/api/queues

# Inspect DLQ message counts
curl -u guest:guest http://localhost:15672/api/queues | jq '.[] | select(.name | endswith(".dlq")) | {name, messages}'

# Verify consumer is running
docker compose logs analytics-service | grep "consumer"
```

### Redis Cache Issues

```bash
# Connect to Redis CLI
docker exec -it redis redis-cli

# List all keys
KEYS *

# Check TTL
TTL [key-name]

# Flush cache
FLUSHALL
```

## Event Schema

### OrderCreated

```json
{
  "event_type": "OrderCreated",
  "event_id": "uuid",
  "timestamp": "2026-01-09T12:34:56Z",
  "data": {
    "order_id": "uuid",
    "customer_id": "string",
    "product_id": "string",
    "quantity": 2,
    "total_amount": 99.99
  }
}
```


### OrderProcessed

```json
{
  "event_type": "OrderProcessed",
  "event_id": "uuid",
  "timestamp": "2026-01-09T12:35:00Z",
  "data": {
    "order_id": "uuid",
    "processed_by": "analytics-service",
    "metrics_updated": true
  }
}
```

## Roadmap

- [ ] gRPC for inter-service communication
- [x] Dead-letter queue (DLQ) for failed messages
- [x] OpenTelemetry distributed tracing
- [ ] Circuit breaker pattern
- [ ] Rate limiting
- [ ] JWT authentication and RBAC
- [ ] API versioning
- [ ] Kubernetes deployment manifests
- [ ] CQRS and Saga patterns

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Commit changes (`git commit -m 'Add your feature'`)
4. Push to branch (`git push origin feature/your-feature`)
5. Open a Pull Request

Please ensure all tests pass and follow the existing code style.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

---

**Author:** [andev0x](https://github.com/andev0x)

**Go Version:** 1.21+ | **Status:** Active Development | **License:** MIT
