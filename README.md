# Golang Service Boilerplate

A comprehensive boilerplate for building scalable Golang-based REST API services with microservice architecture, API gateway, and PostgreSQL support.

## Features

- **Microservice Architecture**: API Gateway with service discovery
- **Distributed Tracing**: OpenTelemetry with Jaeger for observability across services
- **PostgreSQL Integration**: Connection pooling and migrations
- **Advanced Logging System**: Structured JSON logging with file rotation and Docker integration
- **Configuration Management**: Environment-based config with Viper
- **Docker Support**: Containerized deployment with docker-compose
- **REST API Framework**: Gin-based HTTP server with middleware
- **Service Instantiation**: Automated script to create new services
- **Makefile Workflow**: Complete build, test, and deployment automation

## 📚 Documentation

### Core Features

- **[Middleware Architecture](docs/middleware-architecture.md)**: Authentication, logging, tracing, and request processing patterns
- **[Security Architecture](docs/security-architecture.md)**: Authentication, authorization, and service exposure guidelines
- **[Logging System](docs/logging-system.md)**: Comprehensive guide to logging configuration, options, and troubleshooting
- **[Distributed Tracing](docs/tracing/)**: Complete OpenTelemetry implementation with Jaeger
  - [Overview & Architecture](docs/tracing/overview.md)
  - [Developer Guide](docs/tracing/developer-guide.md)
  - [Configuration](docs/tracing/configuration.md)
  - [Database Tracing](docs/tracing/database-tracing.md): Instrumenting database operations for performance monitoring
  - [Monitoring & Troubleshooting](docs/tracing/monitoring.md)
  - [Best Practices](docs/tracing/best-practices.md)

### Development & Deployment

- **[Service Creation Guide](docs/service-creation-guide.md)**: How to create new services using the boilerplate
- **[Air Hot Reload](docs/air-hot-reload/)**: Development setup with live reloading
- **[Migrations](docs/migrations/)**: Database migration management and best practices
- **[CLI Utilities](docs/cli-utility-comprehensive.md)**: Command-line tools for development and operations

### API & Examples

- **[Authentication API Examples](docs/auth-api-examples.md)**: Complete API usage examples with authentication
- **[Distributed Tracing Implementation Plan](docs/distributed-tracing-implementation-plan.md)**: Detailed roadmap for implementing OpenTelemetry tracing across microservices

### Planning & Future

- **[Future Development Plan](docs/future_development_plan.md)**: Roadmap of planned features and enhancements
- **[CLI Utility Plan](docs/cli-utility-plan.md)**: Planned CLI enhancements and features

## Project Structure

```
service-boilerplate/
├── api-gateway/           # Central API gateway service
│   ├── cmd/
│   ├── internal/
│   │   ├── config/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   └── services/
│   └── config.yaml
├── services/              # Individual microservices
│   └── user-service/      # Example user management service
│       ├── cmd/
│       ├── internal/
│       │   ├── config/
│       │   ├── database/
│       │   ├── handlers/
│       │   ├── models/
│       │   ├── repository/
│       │   └── services/
│       ├── migrations/
│       └── config.yaml
├── common/                # Shared libraries
│   ├── logging/
│   ├── database/
│   └── config/
├── docker/                # Docker configurations
│   └── docker-compose.yml
├── scripts/               # Utility scripts
│   └── create-service.sh
├── templates/             # Service templates
└── Makefile               # Build automation
```

## 🔐 Security Architecture

The service-boilerplate implements a **secure-by-design microservice architecture** with centralized authentication and authorization.

### Key Security Features

- **API Gateway Security Model**: Single entry point for all external requests with JWT token validation and revocation checking
- **Token Management**: JWT access tokens with refresh token rotation and immediate revocation on logout
- **Service Isolation**: Internal services are not directly exposed in production, accessed only through the secure API Gateway
- **Audit Logging**: Comprehensive security event logging with distributed tracing correlation
- **Role-Based Access Control**: JWT claims include user roles for fine-grained authorization

### Security Flow

```
External Client → API Gateway (Port 8080) → Auth Service (Port 8083)
                                      ↓
                               Internal Services (Ports 8081+)
```

**Production Security:**

- ✅ Only API Gateway exposed externally
- ✅ All requests validated for authentication and token revocation
- ✅ Internal services trust gateway validation
- ✅ Comprehensive audit trails for security events

**Development Security:**

- ⚠️ Direct service access allowed for testing/debugging
- ⚠️ Must not be used for production workflows
- ✅ Same authentication and logging as production

### Security Documentation

- **[Security Architecture Guide](docs/security-architecture.md)**: Complete security model, TokenRevocationChecker patterns, and service exposure guidelines
- **[Authentication Examples](docs/auth-api-examples.md)**: API usage with JWT tokens, token refresh, and error handling
- **[Troubleshooting Auth](docs/troubleshooting-auth-logging.md)**: Debug authentication issues and token problems

## Quick Start

### Prerequisites

- Docker & Docker Compose (recommended)
- Go 1.23+ (for local development)
- PostgreSQL 15+ (for local development)

### 🚀 Docker Development

1. **Quick start (Development with hot reload):**

   ```bash
   # 🛠️  Start DEVELOPMENT environment with hot reload and debug logging
   make dev

   # In another terminal, run database migrations to set up the database:
   make db-migrate

   # This automatically creates:
   # - Database tables and schemas for all services
   # - Dev admin account: dev.admin@example.com / devadmin123 (full admin access)
   # - Test users for development and testing

   # Test basic authentication flow
   ./scripts/test-auth-flow.sh

   # Test RBAC (Role-Based Access Control) endpoints with dev admin account
   ./scripts/test-rbac-endpoints.sh

   # View distributed traces in Jaeger UI:
   # http://localhost:16686
   ```

2. **Production deployment:**

   ```bash
   # 🚀 Start PRODUCTION environment with pre-built optimized images
   make prod

   # Run database migrations
   make db-migrate
   ```

3. **Environment commands:**

   ```bash
   # Check current environment status
   make status

   # View service logs
   make logs

   # Stop all services
   make down
   ```

   **⚠️ Important:** Use `make dev` for development/debugging and `make prod` for production.

4. **Interactive development menu:**

   ```bash
   ./scripts/dev.sh
   ```

### 🔧 Development Workflows

#### **Hot Reload Development**

The project includes **Air** for hot reloading during development:

- **Docker Hot Reload**: `make dev` - Services automatically restart on file changes

#### **Development Tools**

- **Air**: Live reloading for Go applications
- **Development Script**: `./scripts/dev.sh` - Interactive development menu
- **Advanced Logging**: See [Logging System Documentation](docs/logging-system.md)
- **Health Checks**: Automatic service health monitoring
- **Environment Configuration**: Flexible config management

#### **Development Admin Account**

For development and testing purposes, a pre-configured admin account is **automatically created** when you run `make db-migrate`:

- **Email**: `dev.admin@example.com`
- **Password**: `devadmin123`
- **Roles**: `admin`, `user`

This account has full administrative privileges and can be used to test RBAC functionality, manage roles/permissions, and perform administrative operations.

**Example usage:**

```bash
# Login as dev admin
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "dev.admin@example.com", "password": "devadmin123"}'

# Use JWT token for admin operations
./scripts/test-rbac-endpoints.sh
```

1. **Start all services with Air hot-reload:**

   ```bash
   make dev
   ```

2. **View logs for running services:**

   ```bash
   make logs
   ```

3. **Stop services:**

   ```bash
   make down
   ```

### 🏭 Production Deployment

For production deployment with optimized Docker images:

1. **Build and start production services:**

   ```bash
   # Build optimized production images and start services
   make build-prod  # Build production images
   make prod        # Start production containers
   ```

2. **Switch from development to production:**

   ```bash
   # Stop development environment
   make down

   # Build and start production
   make build-prod
   make prod
   ```

3. **Switch from production to development:**

   ```bash
   # Stop production environment
   make down

   # Build development images and start
   make build-dev
   make dev
   ```

### 🔄 Environment Modes

| Mode            | Build Target      | Start Target | Hot Reload | Image Size |
| --------------- | ----------------- | ------------ | ---------- | ---------- |
| **Development** | `make build-dev`  | `make dev`   | ✅ Air     | ~1.2GB     |
| **Production**  | `make build-prod` | `make prod`  | ❌ None    | ~15MB      |

**Note:** Always run `make down` before switching between development and production modes to avoid image conflicts.

## API Usage

### API Gateway

The API Gateway runs on `http://localhost:8080` and proxies requests to individual services.

**Health Check:**

```bash
curl http://localhost:8080/health
```

### User Service

**Create User:**

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

**Get User:**

```bash
curl http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer your-token"
```

**List Users:**

```bash
curl http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer your-token"
```

## Creating New Services

Use the automated script to create new services:

```bash
make create-service SERVICE_NAME=product-service PORT=8082
```

This will:

- Create the service directory structure
- Copy boilerplate code with proper naming
- Update docker-compose.yml
- Update Makefile targets
- Register with API gateway

## Available Make Targets

```bash
# Primary Docker Commands
make dev               # 🛠️  Start DEVELOPMENT environment (hot reload, debug logs)
make prod              # 🚀 Start PRODUCTION environment (pre-built images)
make smart-start       # 🧠 Smart start - automatically detects environment
make down              # Stop all services
make logs              # View service logs
make status            # Show current environment status

# Build Commands
make build-prod        # Build production Docker images
make build-dev         # Build development images with Air
make build             # Build all services

# Database Commands
make db-migrate        # Run database migrations
make db-setup          # Complete database setup
make db-health         # Check database connectivity

# Testing & Maintenance
make test              # Run all tests
make clean             # Clean build artifacts
make health            # Comprehensive health check
make create-service    # Create new service
```

## Configuration

Each service has its own `config.yaml` file. Environment variables can override config values:

```bash
# Override database settings
export APP_DATABASE_HOST=prod-db.example.com
export APP_DATABASE_PASSWORD=secure-password

# Override server settings
export APP_SERVER_PORT=8080
```

## Development Guidelines

### Code Structure

- **cmd/**: Application entry points
- **internal/**: Private application code
  - **handlers/**: HTTP request handlers
  - **services/**: Business logic layer
  - **repository/**: Data access layer
  - **models/**: Data structures
- **pkg/**: Public libraries
- **migrations/**: Database schema changes

### Logging

Structured JSON logging is enabled by default:

```go
logger.WithFields(logrus.Fields{
    "user_id": 123,
    "action": "login",
}).Info("User logged in")
```

### Database

- Uses pgx driver with connection pooling
- Migrations managed with golang-migrate
- Repository pattern for data access

### Testing

```bash
make test                    # Run all tests
make test-user-service       # Run specific service tests
```

## Deployment

### Docker Production

```bash
make build-prod    # Build optimized production images
make prod          # Start production environment
```

### Environment Variables

Set these for production:

```bash
APP_ENV=production
DATABASE_HOST=your-db-host
DATABASE_PASSWORD=your-secure-password
LOGGING_LEVEL=info
```

## Contributing

1. Follow the established code structure
2. Add tests for new features
3. Update documentation
4. Use `make check` before committing

## License

This project is licensed under the MIT License.
