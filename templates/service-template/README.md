# SERVICE_NAME Service

This is the SERVICE_NAME service for the service-boilerplate project.

## Overview

The SERVICE_NAME service provides REST API endpoints for managing entities.

## Features

- RESTful API for entity management
- PostgreSQL database integration
- **Comprehensive Observability**:
  - Structured request/response logging
  - Performance metrics collection
  - Audit logging for security events
  - Configurable alerting system
- Health check endpoints
- Docker support
- Hot reload for development

## API Endpoints

### Health Check
- `GET /health` - Service health check
- `GET /ready` - Service readiness check
- `GET /live` - Service liveness check
- `GET /status` - Comprehensive service status
- `GET /ping` - Simple ping response

### Observability
- `GET /api/v1/metrics` - Real-time performance metrics
- `GET /api/v1/alerts` - Active alerts and notifications

### Entities
- `POST /api/v1/entities` - Create a new entity
- `GET /api/v1/entities/:id` - Get entity by ID
- `PUT /api/v1/entities/:id` - Update entity
- `DELETE /api/v1/entities/:id` - Delete entity
- `GET /api/v1/entities` - List entities

## Configuration

The service uses a YAML configuration file (`config.yaml`) with the following structure:

```yaml
app:
  name: "SERVICE_NAME"
  version: "1.0.0"
  environment: "development"

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  database: "service_db"
  ssl_mode: "disable"

logging:
  level: "debug"
  format: "json"
  output: "stdout"

server:
  host: "0.0.0.0"
  port: PORT

monitoring:
  health_check_timeout: 5
  status_cache_duration: 30
  enable_detailed_metrics: true

alerting:
  enabled: false
  error_rate_threshold: 0.1  # 10% error rate
  response_time_threshold_ms: 5000  # 5 seconds
  alert_interval_minutes: 5  # Alert every 5 minutes max
```

## Development

### Prerequisites

- Go 1.23+
- PostgreSQL
- Docker (optional)

### Running Locally

1. Start PostgreSQL database
2. Run migrations: `make db-migrate-up SERVICE_NAME=SERVICE_NAME`
3. Start the service: `make run-SERVICE_NAME`

### Running with Docker

1. Build and start: `make up`
2. View logs: `make logs`

### Development with Hot Reload

1. Start development environment: `make dev`
2. The service will automatically reload on code changes

## Observability

The SERVICE_NAME service includes comprehensive observability features for monitoring, debugging, and alerting.

### Logging

- **Structured Logging**: All requests and responses are logged with consistent JSON format
- **Request Correlation**: X-Request-ID header propagation for tracing requests across services
- **Audit Logging**: Security events (entity creation, modifications, deletions) are logged with user context
- **Performance Logging**: Slow requests (>5 seconds) are automatically flagged

### Metrics

The `/api/v1/metrics` endpoint provides comprehensive performance metrics including per-endpoint breakdown:

```json
{
  "service_name": "SERVICE_NAME",
  "uptime": "1h30m45s",
  "request_count": 1250,
  "error_count": 12,
  "error_rate": 0.0096,
  "avg_response_time": "245ms",
  "p95_response_time": "890ms",
  "p99_response_time": "2.1s",
  "endpoint_metrics": {
    "GET /api/v1/entities": {
      "path": "/api/v1/entities",
      "method": "GET",
      "requests": 450,
      "error_rate": 0.004,
      "avg_response_time": "28ms",
      "p95_response_time": "85ms",
      "p99_response_time": "120ms"
    },
    "POST /api/v1/entities": {
      "path": "/api/v1/entities",
      "method": "POST",
      "requests": 120,
      "error_rate": 0.025,
      "avg_response_time": "95ms",
      "p95_response_time": "180ms",
      "p99_response_time": "250ms"
    },
    "GET /api/v1/entities/{id}": {
      "path": "/api/v1/entities/{id}",
      "method": "GET",
      "requests": 380,
      "error_rate": 0.008,
      "avg_response_time": "35ms",
      "p95_response_time": "92ms",
      "p99_response_time": "150ms"
    }
  }
}
```

**Path Normalization**: Parameterized routes are automatically normalized (e.g., `/entities/123` becomes `/entities/{id}`) for consistent grouping and analysis.

### Alerting

The service includes configurable alerting for critical events:

- **High Error Rate**: Alerts when error rate exceeds threshold (default: 10%)
- **Slow Response Times**: Alerts when average response time exceeds threshold (default: 5 seconds)
- **Service Unavailability**: Alerts when no requests processed in 5+ minutes

Active alerts can be viewed at `/api/v1/alerts`:

```json
{
  "alerts": [
    {
      "id": "service_high_error_rate_1640995200",
      "service_name": "SERVICE_NAME",
      "type": "error_rate",
      "severity": "warning",
      "message": "High error rate: 15.2% (threshold: 10.0%)",
      "timestamp": "2023-12-31T23:59:59Z",
      "acked": false
    }
  ]
}
```

### Configuration

Alerting can be configured in `config.yaml`:

```yaml
alerting:
  enabled: true  # Set to true to enable alerting
  error_rate_threshold: 0.1  # 10% error rate threshold
  response_time_threshold_ms: 5000  # 5 second response time threshold
  alert_interval_minutes: 5  # Minimum interval between similar alerts
```

## Database Schema

The service uses the following database schema:

```sql
CREATE TABLE entities (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## Project Structure

```
services/SERVICE_NAME/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── handlers/
│   │   └── entity_handler.go   # HTTP handlers
│   ├── models/
│   │   └── entity.go           # Data models
│   ├── repository/
│   │   └── entity_repository.go # Database operations
│   └── services/
│       └── entity_service.go   # Business logic
├── migrations/                 # Database migrations
│   ├── 000001_initial.up.sql
│   └── 000001_initial.down.sql
├── .air.toml                   # Air configuration for hot reload
├── config.yaml                 # Service configuration
├── Dockerfile                  # Production Docker image
├── Dockerfile.dev              # Development Docker image
└── README.md                   # This file
```

## Contributing

1. Follow the existing code style and patterns
2. Add tests for new functionality
3. Update documentation as needed
4. Ensure all tests pass before submitting PR

## License

This project is part of the service-boilerplate and follows the same license terms.