# Service Creation Guide

This guide explains how to create new microservices in the service-boilerplate project using the automated service creation mechanism.

## Overview

The service creation mechanism provides a streamlined way to add new microservices to your project. It automatically:

- Creates service structure from template
- Updates Docker Compose configuration
- Registers service with API gateway
- Creates database migrations (optional)
- Updates Makefile with service-specific targets
- Updates environment configuration
- Updates project documentation

## Prerequisites

Before creating a new service, ensure you have:

1. The service template exists at `templates/service-template/`
2. Available port number (1024-65535) not used by other services
3. PostgreSQL database running (for database-enabled services)

## Quick Start

### Basic Service Creation

```bash
# Create a new service
./scripts/create-service.sh product-service 8082

# Build and run the service
make build-product-service
make run-product-service
```

### Advanced Service Creation

```bash
# Create service without database schema
./scripts/create-service.sh notification-service 8083 --no-db-schema

# Create service without documentation updates
./scripts/create-service.sh analytics-service 8084 --no-docs-update

# Create service with all options
./scripts/create-service.sh inventory-service 8085
```

## Service Structure

When you create a new service, the following structure is generated:

```
services/your-service-name/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── handlers/
│   │   └── service_handler.go  # HTTP request handlers
│   ├── models/
│   │   └── service.go          # Data models and DTOs
│   ├── repository/
│   │   └── service_repository.go # Database operations
│   └── services/
│       └── service_service.go  # Business logic layer
├── migrations/                 # Database migrations
│   ├── 000001_initial.up.sql
│   └── 000001_initial.down.sql
├── .air.toml                   # Hot reload configuration
├── config.yaml                 # Service configuration
├── Dockerfile                  # Production container
├── Dockerfile.dev              # Development container
└── README.md                   # Service documentation
```

## Configuration Files

### Environment Variables (.env)

The script automatically adds service-specific environment variables:

```bash
# Example for product-service
PRODUCT_SERVICE_NAME=product-service
PRODUCT_SERVICE_PORT=8082
PRODUCT_SERVICE_IMAGE=docker-product-service
PRODUCT_SERVICE_CONTAINER=service-boilerplate-product-service
PRODUCT_SERVICE_TMP_VOLUME=service-boilerplate-product-service-tmp
```

### Docker Compose

**docker-compose.yml** (Production):
```yaml
product-service:
  build:
    context: ..
    dockerfile: services/product-service/Dockerfile
  image: ${PRODUCT_SERVICE_IMAGE}
  container_name: ${PRODUCT_SERVICE_CONTAINER}
  ports:
    - "${PRODUCT_SERVICE_PORT:-8082}:${PRODUCT_SERVICE_PORT:-8082}"
  environment:
    - APP_ENV=${APP_ENV:-production}
    - SERVER_PORT=${PRODUCT_SERVICE_PORT:-8082}
    - DATABASE_HOST=${POSTGRES_NAME}
    # ... more environment variables
  depends_on:
    postgres:
      condition: service_healthy
  networks:
    service-network:
      aliases:
        - ${PRODUCT_SERVICE_NAME}
        - product-service
        - product-service-svc
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:${PRODUCT_SERVICE_PORT:-8082}/health"]
    interval: 30s
    timeout: 10s
    retries: 3
    start_period: 10s
  restart: unless-stopped
```

**docker-compose.override.yml** (Development):
```yaml
product-service:
  build:
    context: ..
    dockerfile: services/product-service/Dockerfile.dev
  environment:
    - APP_ENV=development
    - LOGGING_LEVEL=debug
  volumes:
    - ../services/product-service:/app/services/product-service
    - product_service_tmp:/app/services/product-service/tmp
  ports:
    - "${PRODUCT_SERVICE_PORT}:${PRODUCT_SERVICE_PORT}"
  working_dir: /app/services/product-service
  networks:
    service-network:
      aliases:
        - ${PRODUCT_SERVICE_NAME}
        - product-service
        - product-service-svc
```

## API Gateway Integration

The script automatically registers your service with the API gateway:

```go
// In api-gateway/cmd/main.go
serviceRegistry.RegisterService("product-service", "http://product-service:8082")
```

### Adding Routes

After creating the service, you need to manually add routes in the API gateway:

```go
// Add to api-gateway/cmd/main.go
api := router.Group("/api")
api.Use(middleware.AuthMiddleware())
{
    // Product service routes
    products := api.Group("/v1/products")
    {
        products.POST("", gatewayHandler.ProxyRequest("product-service"))
        products.GET("/:id", gatewayHandler.ProxyRequest("product-service"))
        products.PUT("/:id", gatewayHandler.ProxyRequest("product-service"))
        products.DELETE("/:id", gatewayHandler.ProxyRequest("product-service"))
        products.GET("", gatewayHandler.ProxyRequest("product-service"))
    }
}
```

## Database Integration

### Automatic Schema Creation

By default, the script creates database migration files. To apply them:

```bash
# Run migrations for your service
make db-migrate-up SERVICE_NAME=product-service

# Check migration status
make db-migrate-status SERVICE_NAME=product-service

# Rollback if needed
make db-migrate-down SERVICE_NAME=product-service
```

### Migration Files

The template includes basic migration files:

- `migrations/000001_initial.up.sql` - Creates the services table
- `migrations/000001_initial.down.sql` - Drops the services table

Customize these files according to your service's data model.

## Makefile Integration

The script adds service-specific targets to the Makefile:

```makefile
.PHONY: build-product-service
build-product-service: ## Build product-service
	@echo "Building product-service..."
	@mkdir -p $(BUILD_DIR)
	@cd services/product-service && $(GOBUILD) -o ../$(BUILD_DIR)/product-service ./cmd

.PHONY: run-product-service
run-product-service: ## Run product-service
	@echo "Running product-service..."
	@cd services/product-service && $(GO) run ./cmd

.PHONY: test-product-service
test-product-service: ## Run product-service tests
	@echo "Running product-service tests..."
	@cd services/product-service && $(GOTEST) ./...

.PHONY: air-product-service
air-product-service: ## Run product-service with Air locally
	@echo "Starting product-service with Air..."
	@cd services/product-service && air
```

## Development Workflow

### Local Development

1. **Create the service:**
   ```bash
   ./scripts/create-service.sh product-service 8082
   ```

2. **Customize the code:**
   - Update models in `internal/models/`
   - Modify handlers in `internal/handlers/`
   - Adjust business logic in `internal/services/`
   - Update database operations in `internal/repository/`

3. **Update configuration:**
   - Modify `config.yaml` for service-specific settings
   - Update API gateway routes

4. **Run database migrations:**
   ```bash
   make db-migrate-up SERVICE_NAME=product-service
   ```

5. **Start the service:**
   ```bash
   # With hot reload
   make air-product-service

   # Or regular run
   make run-product-service
   ```

### Docker Development

1. **Start all services:**
   ```bash
   make dev
   ```

2. **Check service health:**
   ```bash
   make health
   ```

3. **View logs:**
   ```bash
   make logs
   ```

## Testing

### Unit Tests

```bash
# Run tests for your service
make test-product-service

# Run all tests
make test
```

### Integration Tests

Add integration tests that verify:

- Service startup
- Database connectivity
- API endpoints
- API gateway routing

## Customization

### Modifying the Template

To customize the service template:

1. Edit files in `templates/service-template/`
2. Use placeholders like `SERVICE_NAME` and `PORT`
3. The creation script will replace these with actual values

### Service-Specific Configuration

Each service can have its own configuration in `config.yaml`:

```yaml
app:
  name: "product-service"
  version: "1.0.0"
  environment: "development"

database:
  # Database settings
  host: "localhost"
  port: 5432
  # ... more settings

server:
  host: "0.0.0.0"
  port: 8082
```

## Troubleshooting

### Common Issues

1. **Port already in use:**
   - Choose a different port number
   - Check `netstat -tlnp | grep :PORT`

2. **Service not accessible:**
   - Verify API gateway routes are added
   - Check service logs: `make logs`
   - Test service directly: `curl http://localhost:PORT/health`

3. **Database connection issues:**
   - Ensure PostgreSQL is running
   - Check database credentials in `.env`
   - Verify migrations are applied

4. **Template not found:**
   - Ensure `templates/service-template/` exists
   - Check file permissions

### Debugging

```bash
# Check service status
make health

# View service logs
docker-compose logs product-service

# Connect to database
make db-connect

# Check migration status
make db-migrate-status SERVICE_NAME=product-service
```

## Best Practices

### Service Design

1. **Single Responsibility:** Each service should have one clear purpose
2. **API Consistency:** Follow RESTful conventions
3. **Error Handling:** Implement proper error responses
4. **Logging:** Use structured logging throughout
5. **Configuration:** Externalize configuration settings

### Database Design

1. **Migrations:** Always use migrations for schema changes
2. **Indexes:** Add appropriate indexes for performance
3. **Constraints:** Use foreign keys and constraints
4. **Naming:** Follow consistent naming conventions

### Docker Best Practices

1. **Multi-stage Builds:** Use multi-stage Dockerfiles
2. **Health Checks:** Implement proper health checks
3. **Resource Limits:** Set appropriate resource limits
4. **Security:** Run as non-root user when possible

## Advanced Topics

### Service Discovery

The API gateway uses a simple service registry. For production:

- Consider using Consul, etcd, or Kubernetes service discovery
- Implement health checks and circuit breakers
- Add service mesh (Istio, Linkerd)

### Monitoring and Observability

Add to each service:

- Metrics collection (Prometheus)
- Distributed tracing (Jaeger)
- Centralized logging (ELK stack)
- Health check endpoints

### Security

Implement:

- Authentication and authorization
- Input validation and sanitization
- Rate limiting
- CORS configuration
- HTTPS/TLS encryption

## Contributing

When adding new services:

1. Follow the established patterns
2. Add comprehensive tests
3. Update documentation
4. Ensure CI/CD pipeline compatibility
5. Add monitoring and logging

## Support

For issues with service creation:

1. Check this documentation
2. Review the troubleshooting section
3. Check existing services for examples
4. Create an issue in the project repository

## Examples

### Complete Service Creation Workflow

```bash
# 1. Create the service
./scripts/create-service.sh product-service 8082

# 2. Customize the models
# Edit services/product-service/internal/models/service.go

# 3. Update handlers
# Edit services/product-service/internal/handlers/service_handler.go

# 4. Add API gateway routes
# Edit api-gateway/cmd/main.go

# 5. Run migrations
make db-migrate-up SERVICE_NAME=product-service

# 6. Test the service
make test-product-service

# 7. Start development environment
make dev
```

This workflow creates a fully functional microservice integrated with your existing infrastructure.