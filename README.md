# Golang Service Boilerplate

A comprehensive boilerplate for building scalable Golang-based REST API services with microservice architecture, API gateway, and PostgreSQL support.

## Features

- **Microservice Architecture**: API Gateway with service discovery
- **PostgreSQL Integration**: Connection pooling and migrations
- **Structured Logging**: JSON logging with logrus
- **Configuration Management**: Environment-based config with Viper
- **Docker Support**: Containerized deployment with docker-compose
- **REST API Framework**: Gin-based HTTP server with middleware
- **Service Instantiation**: Automated script to create new services
- **Makefile Workflow**: Complete build, test, and deployment automation

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

### 🚀 Recommended: Docker Development (Primary)

1. **Clone and setup:**
   ```bash
   git clone <repository-url>
   cd service-boilerplate
   ```

2. **Start all services:**
   ```bash
   make up
   ```

3. **View service logs:**
   ```bash
   make logs
   ```

4. **Stop services:**
   ```bash
   make down
   ```

**That's it!** All services (PostgreSQL, API Gateway, User Service) will start automatically with proper networking and health checks.

### 💻 Alternative: Local Development (Secondary)

1. **Setup local environment:**
   ```bash
   make setup
   ```

2. **Start PostgreSQL locally:**
   ```bash
   docker run -d --name postgres \
     -e POSTGRES_DB=service_db \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=postgres \
     -p 5432:5432 postgres:15-alpine
   ```

3. **Run database migrations:**
   ```bash
   make migrate-up
   ```

4. **Build and run services locally:**
   ```bash
   make build
   make run-local     # Runs all services locally
   # OR run individually:
   make run-gateway-local
   make run-user-service-local
   ```

5. **Stop local services:**
   ```bash
   make stop-local
   ```

### 🔧 Development Workflows

1. **Start all services:**
   ```bash
   make docker-run
   ```

2. **View logs:**
   ```bash
   make docker-logs
   ```

3. **Stop services:**
   ```bash
   make docker-stop
   ```

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
make up                # Start all services with Docker (RECOMMENDED)
make down              # Stop all services
make logs              # View service logs
make dev               # Start development environment

# Local Development (Secondary)
make setup             # Initialize project
make build             # Build all services
make run-local         # Run all services locally
make stop-local        # Stop local services
make run-gateway-local # Run API Gateway locally
make run-user-service-local # Run User Service locally

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