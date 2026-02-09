# Objects Service

A microservice for managing generic object taxonomies with hierarchical relationships, flexible attributes, and comprehensive audit capabilities.

## Features

- **Hierarchical Object Types**: Create nested taxonomic categories with parent-child relationships
- **Flexible Objects**: Store any type of object with custom JSONB metadata
- **Dual ID System**: Internal BIGINT ID for database performance, public UUID for API exposure
- **Version Control**: Optimistic locking with version field to prevent concurrent update conflicts
- **Soft Delete**: Preserve deleted objects with deleted_at tracking
- **Comprehensive Audit**: Track created_by, updated_by with timestamps
- **Advanced Search**: Filter by type, status, tags, and metadata
- **Batch Operations**: Create, update, and delete multiple objects in a single request
- **Type Sealing**: Prevent inheritance from sealed types
- **Tag System**: Categorize objects with array-based tags
- **Hierarchical Queries**: Get tree, children, descendants, ancestors, and path

## Architecture

### Core Concepts

**Object Types**
- Define taxonomic categories and hierarchies
- Support parent-child relationships (unlimited depth)
- Optional concrete table mapping for specialized storage
- Sealed types prevent further inheritance
- Flexible metadata for type-specific attributes

**Objects**
- Instantiations of object types
- Internal BIGINT ID for performance
- Public UUID for external API exposure
- Parent-child relationships for hierarchical organization
- Soft delete support with deleted_at tracking
- Version control for optimistic locking
- Status management (active, inactive, archived, deleted, pending)
- Flexible JSONB metadata and string array tags

## Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL 15+
- Docker (optional)

### Installation

```bash
# Clone repository
git clone <repository-url>
cd service-boilerplate/services/objects-service

# Install dependencies
go mod download

# Run migrations
go run cmd/migrate/main.go up

# Start service
go run cmd/main.go
```

### Docker

```bash
# Build image
docker build -t objects-service .

# Run container
docker run -p 8085:8085 \
  -e DATABASE_URL="postgresql://postgres:password@postgres:5432/objects_service" \
  objects-service
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | - |
| `PORT` | Service port | 8085 |
| `ENVIRONMENT` | Environment (development/staging/production) | development |
| `LOG_LEVEL` | Log level (debug/info/warn/error) | info |
| `JWT_SECRET` | JWT secret for authentication | - |
| `ENABLE_CORS` | Enable CORS | true |

### Configuration File

The service uses `config.yaml`:

```yaml
app:
  name: "objects-service"
  version: "1.0.0"
  environment: "development"

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  database: "objects_service"
  ssl_mode: "disable"
  max_conns: 25
  min_conns: 5

logging:
  level: "info"
  format: "json"
  output: "stdout"

server:
  host: "0.0.0.0"
  port: 8085
```

## API Documentation

### Base URL

```
Development: http://localhost:8085
```

### Authentication

Include JWT token in Authorization header:

```
Authorization: Bearer <token>
```

### Health Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Service health check |
| GET | `/ready` | Readiness check (includes DB) |
| GET | `/live` | Liveness check |
| GET | `/status` | Comprehensive status |
| GET | `/ping` | Simple ping response |

### Object Types API

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/object-types` | Create object type |
| GET | `/api/v1/object-types` | List object types |
| GET | `/api/v1/object-types/:id` | Get object type by ID |
| GET | `/api/v1/object-types/name/:name` | Get object type by name |
| PUT | `/api/v1/object-types/:id` | Update object type |
| DELETE | `/api/v1/object-types/:id` | Delete object type |
| GET | `/api/v1/object-types/:id/tree` | Get tree from type |
| GET | `/api/v1/object-types/:id/children` | Get direct children |
| GET | `/api/v1/object-types/:id/descendants` | Get all descendants |
| GET | `/api/v1/object-types/:id/ancestors` | Get all ancestors |
| GET | `/api/v1/object-types/:id/path` | Get path to root |
| GET | `/api/v1/object-types/search` | Search object types |
| POST | `/api/v1/object-types/:id/validate-move` | Validate move operation |
| GET | `/api/v1/object-types/:id/subtree-count` | Count objects in subtree |

#### Create Object Type

```http
POST /api/v1/object-types
Content-Type: application/json

{
  "name": "Product Category",
  "parent_type_id": null,
  "description": "Root category for products",
  "is_sealed": false,
  "metadata": {
    "icon": "product",
    "color": "#FF5722"
  }
}
```

Response:
```json
{
  "data": {
    "id": 1,
    "name": "Product Category",
    "parent_type_id": null,
    "description": "Root category for products",
    "is_sealed": false,
    "metadata": {"icon": "product", "color": "#FF5722"},
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "message": "Object type created successfully"
}
```

#### List Object Types

```http
GET /api/v1/object-types?limit=20&offset=0
```

Response:
```json
{
  "data": [...],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 10,
    "count": 10
  }
}
```

### Objects API

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/objects` | Create object |
| GET | `/api/v1/objects` | List objects |
| GET | `/api/v1/objects/:id` | Get object by internal ID |
| GET | `/api/v1/objects/public-id/:public_id` | Get object by public UUID |
| GET | `/api/v1/objects/name/:name` | Get object by name |
| PUT | `/api/v1/objects/:id` | Update object |
| DELETE | `/api/v1/objects/:id` | Delete object (soft) |
| GET | `/api/v1/objects/:id/children` | Get direct children |
| GET | `/api/v1/objects/:id/descendants` | Get all descendants |
| GET | `/api/v1/objects/:id/ancestors` | Get all ancestors |
| GET | `/api/v1/objects/:id/path` | Get path to root |
| PUT | `/api/v1/objects/:id/metadata` | Update metadata |
| POST | `/api/v1/objects/:id/tags` | Add tags |
| DELETE | `/api/v1/objects/:id/tags` | Remove tags |
| GET | `/api/v1/objects/search` | Search objects |
| POST | `/api/v1/objects/bulk` | Bulk create |
| PUT | `/api/v1/objects/bulk` | Bulk update |
| DELETE | `/api/v1/objects/bulk` | Bulk delete |
| GET | `/api/v1/objects/stats` | Get statistics |

#### Create Object

```http
POST /api/v1/objects
Content-Type: application/json

{
  "object_type_id": 1,
  "parent_object_id": null,
  "name": "Smartphone",
  "description": "Mobile phone device",
  "metadata": {
    "brand": "Example",
    "price": 999.99,
    "screen_size": 6.5
  },
  "status": "active",
  "tags": ["mobile", "electronics", "phone"]
}
```

Response:
```json
{
  "data": {
    "id": 1,
    "public_id": "550e8400-e29b-41d4-a716-446655440000",
    "object_type_id": 1,
    "parent_object_id": null,
    "name": "Smartphone",
    "description": "Mobile phone device",
    "metadata": {"brand": "Example", "price": 999.99},
    "status": "active",
    "tags": ["mobile", "electronics"],
    "version": 0,
    "created_by": "user123",
    "updated_by": "user123",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "deleted_at": null
  },
  "message": "Object created successfully"
}
```

#### List Objects

```http
GET /api/v1/objects?object_type_id=1&status=active&limit=20
```

#### Bulk Create

```http
POST /api/v1/objects/bulk
Content-Type: application/json

[
  {"object_type_id": 1, "name": "Product 1"},
  {"object_type_id": 1, "name": "Product 2"},
  {"object_type_id": 1, "name": "Product 3"}
]
```

## Usage Examples

### Creating a Product Category Hierarchy

```bash
# Create root category
curl -X POST http://localhost:8085/api/v1/object-types \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Product Category",
    "description": "Root category for products"
  }'

# Create sub-category
curl -X POST http://localhost:8085/api/v1/object-types \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Electronics",
    "parent_type_id": 1,
    "description": "Electronic products"
  }'

# Create product object
curl -X POST http://localhost:8085/api/v1/objects \
  -H "Content-Type: application/json" \
  -d '{
    "object_type_id": 2,
    "name": "Smartphone",
    "metadata": {
      "brand": "Example",
      "price": 999.99
    },
    "tags": ["mobile", "electronics"]
  }'
```

### Working with Hierarchies

```bash
# Get all children of a category
curl http://localhost:8085/api/v1/object-types/1/children

# Get full tree from root
curl http://localhost:8085/api/v1/object-types/1/tree

# Get ancestors of a type
curl http://localhost:8085/api/v1/object-types/2/ancestors

# Get descendants with max depth of 2
curl "http://localhost:8085/api/v1/object-types/1/descendants?max_depth=2"
```

### Searching Objects

```bash
# Search by name
curl "http://localhost:8085/api/v1/objects/search?q=phone"

# Filter by type and status
curl "http://localhost:8085/api/v1/objects?object_type_id=2&status=active"

# Filter by tags (any match)
curl "http://localhost:8085/api/v1/objects?tags=mobile,electronics"

# Filter by tags (all match)
curl "http://localhost:8085/api/v1/objects?tags=mobile,electronics&tags_mode=all"

# Get statistics
curl http://localhost:8085/api/v1/objects/stats
```

### Version Control (Optimistic Locking)

```bash
# Get object with version
curl http://localhost:8085/api/v1/objects/1

# Update with version (must match current version)
curl -X PUT http://localhost:8085/api/v1/objects/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Name",
    "version": 0
  }'

# If version mismatch, returns 409 Conflict
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific package tests
go test ./internal/models/...
go test ./internal/repository/...
go test ./internal/services/...
go test ./internal/handlers/...
go test ./tests/...
```

### Database Migrations

```bash
# Run migrations
go run cmd/migrate/main.go up

# Rollback migrations
go run cmd/migrate/main.go down

# Run migrations for specific environment
go run cmd/migrate/main.go up --env=development
```

### Code Quality

```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run

# Run vet
go vet ./...
```

## Observability

### Logging

- **Structured Logging**: All requests and responses are logged with consistent JSON format
- **Request Correlation**: X-Request-ID header propagation for tracing requests across services
- **Audit Logging**: Security events are logged with user context
- **Performance Logging**: Slow requests are automatically flagged

### Metrics

The `/api/v1/metrics` endpoint provides comprehensive performance metrics:

```json
{
  "service_name": "objects-service",
  "uptime": "1h30m45s",
  "request_count": 1250,
  "error_count": 12,
  "error_rate": 0.0096,
  "avg_response_time": "245ms"
}
```

### Alerting

The service includes configurable alerting for critical events:

- **High Error Rate**: Alerts when error rate exceeds threshold
- **Slow Response Times**: Alerts when average response time exceeds threshold
- **Service Unavailability**: Alerts when no requests processed

Active alerts can be viewed at `/api/v1/alerts`.

## Troubleshooting

### Common Issues

**Migration fails**
- Verify DATABASE_URL is set correctly
- Ensure PostgreSQL is running
- Check database user has necessary permissions

**Service won't start**
- Check port 8085 is not in use
- Verify environment variables are set
- Check logs for specific error messages

**Authentication errors**
- Ensure JWT_SECRET is set
- Verify token format is correct
- Check token expiration

## License

This project is part of the service-boilerplate and follows the same license terms.
