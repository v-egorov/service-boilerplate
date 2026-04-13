# Golang Service Boilerplate

Microservice architecture with API Gateway, PostgreSQL, and distributed tracing.

## Features

- Microservice Architecture with API Gateway
- PostgreSQL with golang-migrate migrations
- OpenTelemetry/Jaeger distributed tracing
- Gin REST framework with middleware
- Docker & Air hot-reload development

## Quick Start

```bash
# Bootstrap - starts services + sets up database + creates test users
make dev-bootstrap

# Test authentication
./scripts/test-auth-flow.sh

# View traces in Jaeger
# http://localhost:16686
```

## Project Structure

```
service-boilerplate/
├── api-gateway/           # Central API gateway (port 8080)
├── services/              # Microservices
│   ├── auth-service/      # Authentication
│   ├── user-service/      # User management
│   └── objects-service/   # Objects API
├── common/                # Shared libraries
├── docker/                # Docker compose
├── scripts/               # Utility scripts
└── Makefile              # Build automation
```

## Development

```bash
make dev                  # Start with hot reload
make down                 # Stop all services
make logs                 # View logs
make status              # Check status

# Run tests
make test
make test-user-service    # Specific service
```

### Database Migrations

Environment-specific migrations: `development/`, `staging/`, `production/`

```bash
# Initialize (run once per service)
make db-migrate-init SERVICE_NAME=auth-service

# Apply migrations
make db-migrate-up SERVICE_NAME=auth-service

# Rollback
make db-migrate-down SERVICE_NAME=auth-service

# Or run all at once
make db-migrate
```

## API Usage

API Gateway: `http://localhost:8080`

```bash
# Health check
curl http://localhost:8080/health

# Login (dev admin: dev.admin@example.com / devadmin123)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"dev.admin@example.com","password":"devadmin123"}'

# Create user (with JWT token)
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"email":"user@example.com","first_name":"John","last_name":"Doe"}'
```

## Configuration

Environment variables override `config.yaml`:

```bash
export APP_DATABASE_HOST=prod-db.example.com
export APP_DATABASE_PASSWORD=secure-password
export APP_SERVER_PORT=8080
```

## Creating Services

```bash
make create-service SERVICE_NAME=product-service PORT=8082
```

## Available Commands

```bash
# Primary
make dev          # Start dev environment (hot reload)
make prod         # Start production (pre-built images)
make down         # Stop all services

# Build
make build-dev    # Build dev images
make build-prod   # Build production images

# Database
make db-migrate              # Migrate all services
make db-migrate-init SERVICE_NAME=<name>
make db-migrate-up SERVICE_NAME=<name>

# Other
make test        # Run tests
make health      # Health check
```

## Documentation

- [Service Creation Guide](docs/service-creation-guide.md)
- [Migrations](docs/migrations/)
- [Security](docs/security-architecture.md)
- [Logging](docs/logging-system.md)
- [Distributed Tracing](docs/tracing/)
- [RBAC](docs/rbac-api-guide.md)

## Contributing

1. Follow code structure (cmd/, internal/, migrations/)
2. Add tests for new features
3. Update documentation
4. Use `make check` before committing

## License

MIT