# SERVICE_NAME Service

This is the SERVICE_NAME service for the service-boilerplate project.

## Overview

The SERVICE_NAME service provides REST API endpoints for managing entities.

## Features

- RESTful API for entity management
- PostgreSQL database integration
- Structured logging
- Health check endpoint
- Docker support
- Hot reload for development

## API Endpoints

### Health Check
- `GET /health` - Service health check

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