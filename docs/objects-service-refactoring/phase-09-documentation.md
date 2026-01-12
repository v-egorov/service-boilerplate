# Phase 9: Documentation

**Estimated Time**: 2 hours
**Status**: ⬜ Not Started
**Dependencies**: Phase 8 (Tests)

## Overview

Update and create comprehensive documentation for the refactored objects-service, including README, API documentation, and usage examples.

## Tasks

### 9.1 Update Service README

**File**: `README.md`

**Steps**:
1. Update service overview
2. Document new API endpoints
3. Add setup instructions
4. Add usage examples
5. Document configuration

```markdown
# Objects Service

A microservice for managing generic object taxonomies with hierarchical relationships, flexible attributes, and comprehensive audit capabilities.

## Features

- **Hierarchical Object Types**: Create nested taxonomic categories
- **Flexible Objects**: Store any type of object with custom metadata
- **Dual ID System**: Internal BIGINT for performance, public UUID for API
- **Version Control**: Optimistic locking with version field
- **Soft Delete**: Preserve deleted objects with deleted_at tracking
- **Comprehensive Audit**: Track created_by, updated_by with timestamps
- **Advanced Search**: Filter by type, status, tags, and metadata
- **Batch Operations**: Create and update multiple objects in a single request
- **Type Sealing**: Prevent inheritance from sealed types
- **Tag System**: Categorize objects with array-based tags

## Architecture

### Core Concepts

**Object Types**
- Define taxonomic categories and hierarchies
- Support parent-child relationships
- Optional concrete table mapping
- Sealed types prevent inheritance
- Flexible metadata for type-specific attributes

**Objects**
- Instantiations of object types
- Internal BIGINT ID for performance
- Public UUID for external API
- Parent-child relationships
- Soft delete support
- Version control
- Status management
- Flexible metadata and tags

## Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL 15+
- Make or Docker for local development

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

### Database Connection

```go
// config/config.go
type Config struct {
    DatabaseURL   string `env:"DATABASE_URL,required"`
    Port         string `env:"PORT" envDefault:"8085"`
    Environment   string `env:"ENVIRONMENT" envDefault:"development"`
    LogLevel     string `env:"LOG_LEVEL" envDefault:"info"`
    JWTSecret    string `env:"JWT_SECRET,required"`
    EnableCORS   bool   `env:"ENABLE_CORS" envDefault:"true"`
}
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

### Object Types API

#### List Object Types

```http
GET /api/v1/object-types
```

Query Parameters:
- `parent_type_id` (int): Filter by parent type
- `is_sealed` (bool): Filter by sealed status
- `search` (string): Search in name and description

Response:
```json
{
  "object_types": [
    {
      "id": 1,
      "name": "Category",
      "parent_type_id": null,
      "concrete_table_name": null,
      "description": "Root category",
      "is_sealed": false,
      "metadata": {},
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20
}
```

#### Get Object Type Tree

```http
GET /api/v1/object-types/tree
```

Returns full hierarchical tree of object types.

#### Create Object Type

```http
POST /api/v1/object-types
Content-Type: application/json

{
  "name": "Product",
  "parent_type_id": 1,
  "description": "Product type",
  "is_sealed": false,
  "metadata": {
    "icon": "product",
    "color": "#FF5722"
  }
}
```

#### Get Object Type by ID

```http
GET /api/v1/object-types/:id
```

#### Update Object Type

```http
PUT /api/v1/object-types/:id
Content-Type: application/json

{
  "description": "Updated description",
  "metadata": {
    "icon": "product-new"
  }
}
```

#### Delete Object Type

```http
DELETE /api/v1/object-types/:id
```

### Objects API

#### List Objects

```http
GET /api/v1/objects
```

Query Parameters:
- `object_type_id` (int): Filter by object type
- `parent_object_id` (int): Filter by parent object
- `status` (string): Filter by status (active/inactive/archived/deleted/pending)
- `include_deleted` (bool): Include soft-deleted objects
- `search` (string): Search in name and description
- `tags` (string): Comma-separated tags
- `tags_mode` (string): Tags match mode (any/all)
- `page` (int): Page number (default: 1)
- `page_size` (int): Page size (default: 20, max: 100)
- `sort_by` (string): Sort field (name/created_at/updated_at)
- `sort_order` (string): Sort order (asc/desc)

Response:
```json
{
  "objects": [
    {
      "id": 1,
      "public_id": "550e8400-e29b-41d4-a716-446655440000",
      "object_type_id": 1,
      "parent_object_id": null,
      "name": "Test Object",
      "description": "Test description",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "deleted_at": null,
      "version": 0,
      "created_by": "user123",
      "updated_by": "user123",
      "metadata": {},
      "status": "active",
      "tags": ["tag1", "tag2"]
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20
}
```

#### Create Object

```http
POST /api/v1/objects
Content-Type: application/json

{
  "object_type_id": 1,
  "parent_object_id": null,
  "name": "New Object",
  "description": "Object description",
  "metadata": {
    "custom_field": "value"
  },
  "status": "active",
  "tags": ["tag1", "tag2"]
}
```

#### Get Object by Public ID

```http
GET /api/v1/objects/:public_id
```

#### Update Object

```http
PUT /api/v1/objects/:public_id
Content-Type: application/json

{
  "name": "Updated Name",
  "description": "Updated description",
  "version": 0
}
```

#### Soft Delete Object

```http
DELETE /api/v1/objects/:public_id
```

#### Hard Delete Object

```http
DELETE /api/v1/objects/:public_id/hard
```

#### Batch Create Objects

```http
POST /api/v1/objects/batch
Content-Type: application/json

{
  "objects": [
    {
      "object_type_id": 1,
      "name": "Object 1",
      "tags": ["tag1"]
    },
    {
      "object_type_id": 1,
      "name": "Object 2",
      "tags": ["tag2"]
    }
  ]
}
```

#### Batch Update Objects

```http
PATCH /api/v1/objects/batch
Content-Type: application/json

{
  "updates": [
    {
      "id": 1,
      "version": 0,
      "changes": {
        "name": "Updated Name 1"
      }
    },
    {
      "id": 2,
      "version": 0,
      "changes": {
        "name": "Updated Name 2"
      }
    }
  ]
}
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

### Searching Objects

```bash
# Search by name
curl "http://localhost:8085/api/v1/objects?search=phone"

# Filter by type and status
curl "http://localhost:8085/api/v1/objects?object_type_id=2&status=active"

# Filter by tags
curl "http://localhost:8085/api/v1/objects?tags=mobile,electronics&tags_mode=all"

# Filter by metadata
curl "http://localhost:8085/api/v1/objects?metadata.brand=Example"
```

### Version Control (Optimistic Locking)

```bash
# Get object with version
curl http://localhost:8085/api/v1/objects/{public_id}

# Update with version (must match current version)
curl -X PUT http://localhost:8085/api/v1/objects/{public_id} \
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

[License Information]

## Support

For issues and questions:
- Create an issue in the repository
- Contact: [support email]
```

---

### 9.2 Create API OpenAPI/Swagger Documentation

**File**: `api/swagger.yaml`

**Steps**:
1. Create OpenAPI 3.0 specification
2. Document all endpoints
3. Add request/response schemas
4. Add authentication requirements

```yaml
openapi: 3.0.0
info:
  title: Objects Service API
  description: API for managing generic object taxonomies
  version: 1.0.0
servers:
  - url: http://localhost:8085
    description: Development server

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

  schemas:
    ObjectType:
      type: object
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
        parent_type_id:
          type: integer
          format: int64
          nullable: true
        description:
          type: string
          nullable: true
        is_sealed:
          type: boolean
        metadata:
          type: object
          additionalProperties: true
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time

    Object:
      type: object
      properties:
        id:
          type: integer
          format: int64
        public_id:
          type: string
          format: uuid
        object_type_id:
          type: integer
          format: int64
        parent_object_id:
          type: integer
          format: int64
          nullable: true
        name:
          type: string
        description:
          type: string
          nullable: true
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time
        deleted_at:
          type: string
          format: date-time
          nullable: true
        version:
          type: integer
          format: int64
        created_by:
          type: string
        updated_by:
          type: string
        metadata:
          type: object
          additionalProperties: true
        status:
          type: string
          enum: [active, inactive, archived, deleted, pending]
        tags:
          type: array
          items:
            type: string

paths:
  /api/v1/object-types:
    get:
      summary: List object types
      tags:
        - Object Types
      parameters:
        - name: parent_type_id
          in: query
          schema:
            type: integer
        - name: is_sealed
          in: query
          schema:
            type: boolean
        - name: search
          in: query
          schema:
            type: string
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: object
                properties:
                  object_types:
                    type: array
                    items:
                      $ref: '#/components/schemas/ObjectType'
                  total:
                    type: integer

    post:
      summary: Create object type
      tags:
        - Object Types
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
                parent_type_id:
                  type: integer
                description:
                  type: string
                is_sealed:
                  type: boolean
                metadata:
                  type: object
      responses:
        '201':
          description: Object type created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ObjectType'

  /api/v1/object-types/{id}:
    get:
      summary: Get object type by ID
      tags:
        - Object Types
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ObjectType'

  /api/v1/objects:
    get:
      summary: List objects
      tags:
        - Objects
      parameters:
        - name: object_type_id
          in: query
          schema:
            type: integer
        - name: status
          in: query
          schema:
            type: string
        - name: page
          in: query
          schema:
            type: integer
            default: 1
        - name: page_size
          in: query
          schema:
            type: integer
            default: 20
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: object
                properties:
                  objects:
                    type: array
                    items:
                      $ref: '#/components/schemas/Object'
                  total:
                    type: integer
                  page:
                    type: integer
                  page_size:
                    type: integer

    post:
      summary: Create object
      tags:
        - Objects
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                object_type_id:
                  type: integer
                parent_object_id:
                  type: integer
                name:
                  type: string
                description:
                  type: string
                metadata:
                  type: object
                status:
                  type: string
                tags:
                  type: array
                  items:
                    type: string
      responses:
        '201':
          description: Object created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Object'
```

---

### 9.3 Create Changelog

**File**: `CHANGELOG.md`

```markdown
# Changelog

All notable changes to the Objects Service will be documented in this file.

## [Unreleased]

### Added
- Hierarchical object types with parent-child relationships
- Flexible objects with JSONB metadata
- Dual ID system (internal BIGINT + public UUID)
- Version control with optimistic locking
- Soft delete support
- Comprehensive audit fields (created_by, updated_by)
- Advanced search and filtering
- Batch operations (create, update)
- Tag system for categorization
- Type sealing to prevent inheritance
- Status management (active, inactive, archived, deleted, pending)

### Changed
- Complete refactoring from generic entity system to taxonomy service
- Replaced `/api/v1/entities` endpoints with `/api/v1/object-types` and `/api/v1/objects`
- New database schema with object_types and objects tables

### Removed
- Old entity CRUD endpoints
- Generic entity table
- Entity-related models, repositories, services, handlers

### Fixed
- None

### Security
- Added JWT authentication support
- Added RBAC permissions (object_types:read/write/delete, objects:read/write/delete/admin)

## [0.1.0] - Initial Release

### Added
- Initial service scaffold
- Basic entity CRUD operations
- Database migrations
- Health check endpoint
```

---

## Checklist

- [ ] Update `README.md` with new API documentation
- [ ] Create `api/swagger.yaml` OpenAPI specification
- [ ] Create `CHANGELOG.md`
- [ ] Update `docs/objects-service-refactoring/progress.md` with completion status
- [ ] Verify all documentation is accurate
- [ ] Test API examples from README
- [ ] Verify Swagger file is valid
- [ ] Update AGENTS.md if needed

## Testing

```bash
# Test README examples
curl http://localhost:8085/health
curl http://localhost:8085/api/v1/object-types
curl http://localhost:8085/api/v1/objects

# Validate Swagger file
swagger validate api/swagger.yaml

# Generate HTML documentation from Swagger
swagger generate spec -o api/swagger.yaml
```

## Common Issues

**Issue**: API examples don't work
**Solution**: Verify service is running and port is correct

**Issue**: Swagger file validation fails
**Solution**: Check for syntax errors in YAML and ensure all references are valid

**Issue**: Documentation outdated
**Solution**: Keep documentation in sync with code changes

## Completion

All phases are now complete! The objects-service has been successfully refactored from a generic entity system to a full-featured taxonomy management service.

## Post-Implementation Tasks

1. Deploy to development environment
2. Run integration tests
3. Set up monitoring and logging
4. Configure API Gateway routing
5. Update user documentation
6. Train team on new API
7. Plan production rollout
