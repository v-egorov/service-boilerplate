# 🔐 Auth Service

The auth-service provides comprehensive authentication and authorization for the service-boilerplate microservices architecture, including JWT token management, user authentication, and role-based access control.

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [API Endpoints](#api-endpoints)
- [JWT Key Management](#jwt-key-management)
- [Configuration](#configuration)
- [Database Schema](#database-schema)
- [Development](#development)
- [Security](#security)
- [Monitoring](#monitoring)

## Overview

The auth-service is the central authentication authority for the microservices ecosystem, handling:

- User registration and login
- JWT token generation and validation
- Token refresh and revocation
- Role-based access control (RBAC)
- Automatic JWT key rotation
- Comprehensive audit logging

## Features

### 🔑 Authentication

- Secure user registration with password hashing (bcrypt)
- JWT-based authentication with access/refresh tokens
- Token refresh capabilities
- Secure logout with token revocation

### 🛡️ Authorization

- Role-based access control (RBAC)
- Admin-protected endpoints for sensitive operations
- JWT middleware integration across services

### 🔄 Key Management

- Automatic JWT key rotation (30-day intervals)
- Manual key rotation for security incidents
- Key overlap periods for zero-downtime rotation
- Cryptographically secure key generation (RSA-2048)

### 📊 Monitoring & Audit

- Comprehensive health checks including rotation status
- Detailed audit logging for all security events
- Structured logging with trace correlation
- Performance metrics and monitoring

### 🏗️ Architecture

- PostgreSQL database for persistence
- RESTful API design
- Docker containerization
- Hot reload for development

## API Endpoints

### Public Endpoints

#### Authentication

- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - User logout
- `GET /api/v1/auth/me` - Get current user info
- `POST /api/v1/auth/validate-token` - Validate JWT token

#### Health & Status

- `GET /health` - Basic health check
- `GET /ping` - Simple ping response
- `GET /status` - Comprehensive service status

### Admin Endpoints (Require `admin` Role)

#### Key Management

- `POST /api/v1/admin/rotate-keys` - Manual JWT key rotation

### User Service Integration

The auth-service communicates with the user-service for user data management:

- User creation during registration
- User lookup for authentication
- User updates and profile management

## JWT Key Management

### Automatic Rotation

- **Interval**: 30 days (configurable)
- **Trigger**: Time-based background monitoring
- **Overlap**: 60 minutes to prevent service disruption
- **Audit**: All rotations logged with actor identification

### Manual Rotation

```bash
# Admin-only endpoint
curl -X POST http://localhost:8083/api/v1/admin/rotate-keys \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

### Key Storage

- RSA-2048 key pairs stored in PostgreSQL
- Private keys encrypted at rest
- Public keys distributed to services via API
- Key metadata tracked (creation, rotation, expiration)

## Configuration

### Environment Variables

```bash
# Service Configuration
APP_NAME=auth-service
APP_VERSION=1.0.0
APP_ENV=development

# Database
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
DATABASE_NAME=service_db
DATABASE_SSL_MODE=disable

# JWT Configuration
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=24h
JWT_PUBLIC_KEY_PATH=/path/to/public/key

# Key Rotation
JWT_ROTATION_ENABLED=true
JWT_ROTATION_TYPE=time
JWT_ROTATION_INTERVAL_DAYS=30
JWT_ROTATION_OVERLAP_MINUTES=60

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8083

# Logging
LOGGING_LEVEL=info
LOGGING_FORMAT=json
LOGGING_OUTPUT=stdout

# Tracing
TRACING_ENABLED=true
TRACING_SERVICE_NAME=auth-service
TRACING_COLLECTOR_URL=http://jaeger:4318/v1/traces
```

### Service Dependencies

```yaml
# docker-compose.yml
depends_on:
  postgres:
    condition: service_healthy
  user-service:
    condition: service_started
```

## Database Schema

### Core Tables

#### Users (via user-service integration)

- User profiles managed by user-service
- Auth-service handles authentication logic

#### JWT Keys

```sql
CREATE TABLE auth_service.jwt_keys (
    id BIGSERIAL PRIMARY KEY,
    key_id VARCHAR(255) UNIQUE NOT NULL,
    private_key_pem TEXT NOT NULL,
    public_key_pem TEXT NOT NULL,
    algorithm VARCHAR(50) DEFAULT 'RS256',
    is_active BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    rotation_reason TEXT,
    rotated_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB
);
```

#### Key Rotation Config

```sql
CREATE TABLE auth_service.key_rotation_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rotation_type VARCHAR(50) NOT NULL DEFAULT 'time',
    interval_days INTEGER DEFAULT 30,
    max_tokens INTEGER DEFAULT 100000,
    overlap_minutes INTEGER DEFAULT 60,
    enabled BOOLEAN DEFAULT true,
    last_rotation_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

#### Refresh Tokens

```sql
CREATE TABLE auth_service.refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    user_id UUID NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP WITH TIME ZONE
);
```

## Development

### Prerequisites

- Go 1.23+
- PostgreSQL 13+
- Docker & Docker Compose
- Make

### Quick Start

1. **Clone and setup:**

   ```bash
   git clone <repository>
   cd service-boilerplate
   cp .env.example .env
   ```

2. **Start dependencies:**

   ```bash
   make up-postgres
   make db-migrate-up SERVICE_NAME=auth-service
   ```

3. **Run the service:**

   ```bash
   # With hot reload
   make air-auth-service

   # Or regular run
   make run-auth-service
   ```

4. **Test authentication:**

   ```bash
   # Register a user
   curl -X POST http://localhost:8083/api/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","password":"password123","first_name":"Test","last_name":"User"}'

   # Login
   curl -X POST http://localhost:8083/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","password":"password123"}'
   ```

### Development Workflow

1. **Code Changes**: Modify handlers, services, or utilities
2. **Testing**: Run unit tests with `make test-auth-service`
3. **Integration**: Test with full stack using `make dev`
4. **Migrations**: Add database changes in `migrations/` directory

### Project Structure

```
services/auth-service/
├── cmd/
│   └── main.go                    # Application entry point
├── internal/
│   ├── handlers/
│   │   ├── auth_handler.go        # Authentication endpoints
│   │   └── health_handler.go      # Health check endpoints
│   ├── models/
│   │   └── auth.go                # Data models and DTOs
│   ├── repository/
│   │   └── auth_repository.go     # Database operations
│   ├── services/
│   │   ├── auth_service.go        # Authentication business logic
│   │   └── key_rotation_manager.go # Key rotation logic
│   └── utils/
│       └── jwt.go                 # JWT utilities
├── migrations/                    # Database migrations
│   ├── 000001_initial.up.sql
│   ├── 000001_initial.down.sql
│   ├── 000002_refresh_tokens.up.sql
│   └── 000003_key_rotation.up.sql
├── config.yaml                    # Service configuration
├── .air.toml                      # Hot reload configuration
├── Dockerfile                     # Production container
├── Dockerfile.dev                 # Development container
└── README.md                      # This documentation
```

## Security

### Password Security

- bcrypt hashing with cost factor 12
- Secure password validation (length, complexity)
- No plaintext password storage

### Token Security

- RSA-2048 for JWT signing
- Short-lived access tokens (15 minutes)
- Secure refresh token storage with hashing
- Automatic token revocation on logout

### Key Rotation Security

- Regular key rotation prevents long-term compromise
- Overlap periods ensure service continuity
- Admin-only manual rotation controls
- Comprehensive audit logging

### Access Control

- Role-based permissions (admin, user)
- JWT middleware validation
- Request context user identification
- Audit logging for security events

## Monitoring

### Health Checks

The service provides multiple health check endpoints:

```bash
# Basic health
curl http://localhost:8083/health

# Detailed status
curl http://localhost:8083/status
```

**Status Response:**

```json
{
  "status": "healthy",
  "service": {
    "name": "auth-service",
    "version": "1.0.0",
    "uptime": "2h30m15s"
  },
  "database": {
    "status": "healthy",
    "response_time": "5ms"
  },
  "jwt_keys": {
    "status": "healthy",
    "key_id": "key_20250101_abc123"
  },
  "rotation": {
    "status": "healthy",
    "type": "time",
    "enabled": true,
    "days_since_last": 15.5,
    "next_rotation": "2025-02-14T10:30:00Z"
  }
}
```

### Metrics

- Authentication success/failure rates
- Token issuance and validation metrics
- Key rotation events
- Database connection pool stats
- Request latency and throughput

### Logging

Structured JSON logging with:

- Request correlation (trace_id, span_id)
- User context (user_id, roles)
- Security events (login, logout, rotation)
- Error tracking with stack traces
- Performance metrics

### Alerting

Configure alerts for:

- Authentication failures spikes
- Key rotation failures
- Database connectivity issues
- High latency responses
- Security events (admin actions)

## Troubleshooting

### Common Issues

1. **401 Unauthorized**: Check JWT token validity and expiration
2. **Database Connection**: Verify PostgreSQL is running and accessible
3. **Key Rotation Failed**: Check database permissions and key generation
4. **Health Check Failed**: Review service logs and dependencies

### Debug Commands

```bash
# Check service logs
docker logs auth-service

# Verify database connectivity
make db-connect

# Test authentication flow
curl -v -X POST http://localhost:8083/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Check key rotation status
curl http://localhost:8083/status | jq '.rotation'
```

## API Examples

See [Authentication API Examples](../docs/auth-api-examples.md) for comprehensive usage examples including:

- Complete authentication flows
- Token refresh patterns
- Error handling
- Multi-service integration

## Related Documentation

- [JWT Key Rotation](../docs/jwt-key-rotation.md) - Key rotation operations
- [Security Architecture](../docs/security-architecture.md) - Overall security model
- [Service Integration Patterns](../docs/service-integration-patterns.md) - Integration guidelines
- [RBAC Implementation](../docs/rbac-implementation.md) - Role-based access control
- [Troubleshooting Auth & Logging](../docs/troubleshooting-auth-logging.md) - Issue resolution

