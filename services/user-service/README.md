# User Service

A microservice for managing user data in the service-boilerplate microservices architecture. Provides user CRUD operations and is integrated with auth-service for authentication.

## Overview

The user-service handles:
- User profile management (CRUD operations)
- User lookup by ID and email
- Integration with auth-service for authentication
- Comprehensive audit logging

## Features

### User Management
- Create, read, update, delete user profiles
- Lookup by UUID or email address
- Partial updates (PATCH) and full replacement (PUT)
- Soft delete support

### Integration
- Consumed by auth-service for user authentication
- Provides user data for permission checks
- Health checks with database connectivity

### Monitoring
- Health check endpoints (liveness, readiness)
- Structured logging with trace correlation
- Metrics collection
- Alerting support

## API Endpoints

### Base URL

```
Development (via API Gateway): http://localhost:8080
Direct (internal): http://localhost:8081
```

### Authentication

All requests should go through the API Gateway (port 8080) which handles JWT validation. The gateway forwards user identity via headers:
- `X-User-ID`: User's unique identifier
- `X-User-Email`: User's email address
- `X-User-Roles`: Comma-separated list of user roles

In development mode, internal services trust these headers from the gateway.

### User Management

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/users` | Create new user |
| GET | `/api/v1/users` | List users |
| GET | `/api/v1/users/:id` | Get user by UUID |
| GET | `/api/v1/users/by-email/:email` | Get user by email |
| PUT | `/api/v1/users/:id` | Replace user (full update) |
| PATCH | `/api/v1/users/:id` | Update user (partial update) |
| DELETE | `/api/v1/users/:id` | Delete user |

### Health & Status

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Basic health check |
| GET | `/ready` | Readiness check (includes DB) |
| GET | `/live` | Liveness check |
| GET | `/status` | Comprehensive status |
| GET | `/ping` | Simple ping response |

### Metrics & Alerts

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/metrics` | Service metrics |
| GET | `/api/v1/alerts` | Active alerts |

## Configuration

### Configuration File

The service uses `config.yaml`:

```yaml
app:
  name: "user-service"
  version: "1.0.0"
  environment: "development"

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  database: "service_db"
  ssl_mode: "disable"
  schema: "user_service"
  max_conns: 10
  min_conns: 2

logging:
  level: "debug"
  format: "json"
  output: "stdout"

server:
  host: "0.0.0.0"
  port: 8081

tracing:
  enabled: true
  service_name: "user-service"
  collector_url: "http://jaeger:4318/v1/traces"
  sampling_rate: 1.0

monitoring:
  health_check_timeout: 5
  status_cache_duration: 30
  enable_detailed_metrics: true

alerting:
  enabled: false
  error_rate_threshold: 0.1
  response_time_threshold_ms: 5000
```

## Database Schema

### Users Table

```sql
CREATE TABLE user_service.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(50) DEFAULT 'user',
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
```

## Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL 15+
- Docker (optional)

### Installation

```bash
# Clone repository
git clone <repository-url>
cd service-boilerplate

# Run migrations
make db-migrate SERVICE_NAME=user-service

# Start services via API Gateway
make dev-detached
```

The service is accessed via API Gateway at http://localhost:8080

## Development

### Running Tests

```bash
# Run all tests
make test-user-service

# Or directly
cd services/user-service && go test ./...
```

### Database Migrations

Migrations are managed via the migration orchestrator:

```bash
# Run migrations
make db-migrate SERVICE_NAME=user-service

# Rollback one migration
make db-migrate-down SERVICE_NAME=user-service
```

## Usage Examples

### Create User

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### Get User by Email

```bash
curl http://localhost:8080/api/v1/users/by-email/user@example.com \
  -H "Authorization: Bearer <token>"
```

### Update User

```bash
curl -X PATCH http://localhost:8080/api/v1/users/{user-id} \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "first_name": "Jane"
  }'
```

### List Users

```bash
curl "http://localhost:8080/api/v1/users?limit=20&offset=0" \
  -H "Authorization: Bearer <token>"
```

## Integration with Auth-Service

The auth-service communicates with user-service for:

- **User Registration**: Creates user profiles during registration
- **Login**: Validates user credentials
- **User Lookup**: Retrieves user by email for authentication
- **User Updates**: Synchronizes profile changes

Internal communication happens via HTTP through the API Gateway.

## Security

### JWT Authentication

- Requests should go through API Gateway (port 8080)
- Gateway validates JWT tokens and forwards user context via headers
- In development mode, internal services trust gateway headers

### Access Control

- Role-based access: `admin`, `user`
- Admin role required for user management operations
- Users can only modify their own profile (except admins)

## Monitoring

### Health Checks

```bash
# Basic health
curl http://localhost:8081/health

# Readiness (includes DB)
curl http://localhost:8081/ready

# Status
curl http://localhost:8081/status
```

### Metrics

```bash
curl http://localhost:8081/api/v1/metrics
```

### Logging

- Structured JSON logging
- Request correlation with trace_id
- Audit logging for user modifications

## Troubleshooting

### Common Issues

**Service won't start**
- Check port 8081 is not in use
- Verify PostgreSQL is running
- Check database credentials in config.yaml

**Database connection errors**
- Ensure PostgreSQL is accessible
- Verify database name exists
- Check user permissions

**Authentication errors**
- Ensure requests go through API Gateway
- Verify JWT token is valid

## Related Documentation

- [Auth Service](../services/auth-service/README.md) - Authentication service
- [API Gateway](../api-gateway/README.md) - Request routing
- [Security Architecture](../docs/security-architecture.md) - Security model
