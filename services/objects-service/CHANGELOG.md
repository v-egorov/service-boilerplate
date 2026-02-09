# Changelog

All notable changes to the objects-service will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - Phase 9 Complete

### Added

- **OpenAPI 3.0 Specification** (`api/swagger.yaml`): Complete API documentation with 29 endpoints including:
  - Object Types: CRUD, hierarchy operations (Tree, Children, Descendants, Ancestors, Path), Search, ValidateMove
  - Objects: CRUD, hierarchy operations, metadata/tags management, bulk operations, statistics
  - Health and Metrics endpoints

- **README Documentation**: Comprehensive documentation including:
  - Feature overview and architectural patterns
  - Configuration and setup instructions
  - Complete API endpoint reference with examples
  - Observability features (health checks, metrics, alerts)

### Changed

- **Complete Refactoring from Entities to Taxonomy**: Replaced entity-based architecture with taxonomy-focused design:
  - Renamed `Entity` to `Object` and `EntityType` to `ObjectType`
  - Updated all internal packages: `models/`, `repository/`, `services/`, `handlers/`
  - Removed 8 entity-related files across all layers
  - Added 16 new taxonomy-focused files

- **Database Schema Updates**:
  - `object_types` table: Added `validation_schema` JSONB column for metadata validation
  - `objects` table: Added `metadata` JSONB and `tags` TEXT[] columns
  - Added hierarchical indexes for efficient tree operations
  - Added triggers for cycle detection in parent-child relationships

- **Repository Layer Enhancements**:
  - Added hierarchical query methods using recursive CTEs
  - Added bulk operations (BulkCreate, BulkUpdate, BulkDelete)
  - Added tag and metadata management methods
  - Added search and filtering capabilities

- **Service Layer Enhancements**:
  - Implemented transaction wrapper pattern with `Transaction` interface
  - Added validation for sealed types and parent-child type compatibility
  - Added circular dependency detection for hierarchy operations
  - Implemented business logic for all CRUD and hierarchical operations

- **Handler Layer Enhancements**:
  - Added `GetByName` method to both ObjectType and Object handlers
  - Fixed `ObjectStats` type definition in response schemas
  - Implemented RESTful routing with proper HTTP methods and status codes

### Fixed

- **YAML Syntax Error**: Corrected nested mapping indentation in `api/swagger.yaml` (line 680)
- **JWT Infrastructure**: Fixed race conditions and implemented automatic key refresh
- **Thread-Safe Operations**: Added mutex protection for concurrent key access

### Removed

- Entity-related files (pre Phase 2):
  - `internal/models/entity.go` and `entity_test.go`
  - `internal/repository/entity_repository.go` and `entity_repository_test.go`
  - `internal/services/entity_service.go` and `entity_service_test.go`
  - `internal/handlers/entity_handler.go` and `entity_handler_test.go`

### Deprecated

- Direct database access to `entities` table (use `objects` table instead)
- Old migration files (`000003_dev_test_data.*.sql`) - replaced with taxonomy test data

## [0.1.0] - Phase 8 Complete

### Added

- **Unit Tests** (132 tests total):
  - `internal/models/object_type_test.go`: 16 tests
  - `internal/models/object_test.go`: 20 tests
  - `internal/repository/repository_test.go`: 29 tests
  - `internal/services/object_type_service_test.go`: 15 tests
  - `internal/services/object_service_test.go`: 26 tests
  - `internal/handlers/object_type_handler_test.go`: 15 tests
  - `internal/handlers/object_handler_test.go`: 20 tests

- **Integration Tests** (`tests/integration/api_integration_test.go`):
  - Health check endpoint test
  - Object types CRUD integration test
  - Objects CRUD integration test

### Changed

- **Test Coverage**: Improved from ~40% to ~75% across all packages

## [0.1.0-rc1] - Phase 5 Complete

### Added

- **API Endpoints** (29 total):
  - `POST /api/v1/object-types` - Create object type
  - `GET /api/v1/object-types/{id}` - Get object type by ID
  - `GET /api/v1/object-types/by-name/{name}` - Get object type by name
  - `PUT /api/v1/object-types/{id}` - Update object type
  - `DELETE /api/v1/object-types/{id}` - Delete object type
  - `GET /api/v1/object-types/{id}/tree` - Get object type tree
  - `GET /api/v1/object-types/{id}/children` - Get direct children
  - `GET /api/v1/object-types/{id}/descendants` - Get all descendants
  - `GET /api/v1/object-types/{id}/ancestors` - Get all ancestors
  - `GET /api/v1/object-types/{id}/path` - Get path from root
  - `GET /api/v1/object-types` - List object types
  - `GET /api/v1/object-types/search` - Search object types
  - `POST /api/v1/object-types/{id}/validate-move` - Validate move operation
  - `GET /api/v1/object-types/{id}/subtree-object-count` - Count objects in subtree
  - `POST /api/v1/objects` - Create object
  - `GET /api/v1/objects/{id}` - Get object by ID
  - `GET /api/v1/objects/by-public-id/{public_id}` - Get by public ID
  - `GET /api/v1/objects/by-name/{name}` - Get by name
  - `PUT /api/v1/objects/{id}` - Update object
  - `DELETE /api/v1/objects/{id}` - Delete object
  - `GET /api/v1/objects` - List objects
  - `GET /api/v1/objects/search` - Search objects
  - `PATCH /api/v1/objects/{id}/metadata` - Update metadata
  - `POST /api/v1/objects/{id}/tags` - Add tags
  - `DELETE /api/v1/objects/{id}/tags/{tag}` - Remove tag
  - `GET /api/v1/objects/{id}/children` - Get direct children
  - `GET /api/v1/objects/{id}/descendants` - Get all descendants
  - `GET /api/v1/objects/{id}/ancestors` - Get all ancestors
  - `GET /api/v1/objects/{id}/path` - Get path from root
  - `POST /api/v1/objects/bulk` - Bulk create objects
  - `PUT /api/v1/objects/bulk` - Bulk update objects
  - `DELETE /api/v1/objects/bulk` - Bulk delete objects
  - `GET /api/v1/objects/stats` - Get object statistics
  - `GET /health` - Health check
  - `GET /metrics` - Prometheus metrics
  - `GET /alerts` - Active alerts

- **Transaction Support**: `Transaction` interface with `Commit()` and `Rollback()` methods

### Changed

- **Routes Configuration**: All routes wired in `cmd/main.go`
- **Response Formats**: Consistent JSON response structure across all endpoints

## Design Decisions

| Question | Decision | Notes |
|----------|----------|-------|
| Q1: Hierarchy depth limits | No limit | Database-level cycle detection triggers |
| Q2: Type compatibility rules | No restrictions | Maximum flexibility initially |
| Q3: Sealed type behavior | Reject creation | Clear validation error |
| Q4: Public ID generation | UUID v7 | Time-ordered UUID |
| Q5: Metadata schema validation | JSON Schema | Per object_type stored in validation_schema |
| Q6: API Gateway routing | Clear separation | Separate endpoints |
| Q7: Authentication | Permission-based RBAC | JWT with claims |
| Q8: Batch operations | Full batch support | Transaction safety |
| Q9: Search sophistication | Indexed search | Full-text search capability |
| Q10: Audit field storage | Store user_id | JWT sub claim |

---

**Last Updated**: 2026-02-09
**Version**: 0.1.0 (Unreleased)
