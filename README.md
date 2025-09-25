# Golang Service Boilerplate

A comprehensive boilerplate for building scalable Golang-based REST API services with microservice architecture, API gateway, and PostgreSQL support.

## Features

- **Microservice Architecture**: API Gateway with service discovery
- **PostgreSQL Integration**: Connection pooling and migrations
- **Advanced Logging System**: Structured JSON logging with file rotation and Docker integration
- **Configuration Management**: Environment-based config with Viper
- **Docker Support**: Containerized deployment with docker-compose
- **REST API Framework**: Gin-based HTTP server with middleware
- **Service Instantiation**: Automated script to create new services
- **Makefile Workflow**: Complete build, test, and deployment automation

## 📚 Documentation

- **[Logging System](docs/logging-system.md)**: Comprehensive guide to logging configuration, options, and troubleshooting
- **[Distributed Tracing Implementation Plan](docs/distributed-tracing-implementation-plan.md)**: Detailed roadmap for implementing OpenTelemetry tracing across microservices
- **[Service Creation Guide](docs/service-creation-guide.md)**: How to create new services using the boilerplate
- **[Air Hot Reload](docs/air-hot-reload/)**: Development setup with live reloading
- **[Migrations](docs/migrations/)**: Database migration management and best practices
- **[Future Development Plan](docs/future_development_plan.md)**: Roadmap of planned features and enhancements

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

## Quick Start

### Prerequisites

- Docker & Docker Compose (recommended)
- Go 1.23+ (for local development)
- PostgreSQL 15+ (for local development)

### 🚀 Docker Development

1. **Quick start:**

   ```bash
   # create volumes dirs and start dev environment with hot reload
   # logs will be shown in current terminal window
   make dev
   
   # then swich to other terminal window - since `make dev` 
   # will continue to stream logs 
   # and run database migrations:
   make db-migrate
   
   # create test users and check auth flow
   ./scripts/test-auth-flow.sh

   # watch logs and then look 
   # at the database service_db.auth_service schema:
   # tables user_sessions and auth_tokens,
   # in service_db.user_service schema: users table
   ```

2. **Development with hot reload (starts all services):**

   ```bash
   make dev
   ```

3. **View service logs:**

   ```bash
   make logs
   ```

4. **Stop services:**

   ```bash
   make down
   ```

5. **Interactive development menu:**

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
    make start

    # Or build and start separately:
    make build-prod  # Build production images
    make up          # Start production containers
    ```

2. **Switch from development to production:**

    ```bash
    # Stop development environment
    make down

    # Build and start production
    make start
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

| Mode | Build Target | Start Target | Hot Reload | Image Size |
|------|-------------|--------------|------------|------------|
| **Development** | `make build-dev` | `make dev` | ✅ Air | ~1.2GB |
| **Production** | `make build-prod` | `make up` | ❌ None | ~15MB |

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
make up                # Start all services with Docker
make down              # Stop all services
make logs              # View service logs
make dev               # Start development environment with hot reload
make dev-build         # Build development images with Air

# Testing & Maintenance
make test              # Run all tests
make clean             # Clean build artifacts
make docker-build      # Build Docker images
make migrate-up        # Run database migrations
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
make docker-build
make docker-run
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
