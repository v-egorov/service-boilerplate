# Objects Service Refactoring Plan

## Overview

This refactoring transforms the objects-service from a simple generic "entity" CRUD service into a sophisticated "generic objects taxonomy" service. The new implementation supports hierarchical object types and objects, flexible attributes, versioning, and comprehensive audit capabilities.

**Status**: Service is NOT in production or development - no migration risks

## Current State

The objects-service was created using `create-service.sh` script and implements a basic CRUD system:

- Single `entities` table with fields: id, name, description, created_at, updated_at
- Simple Entity model with basic CRUD operations
- RESTful API endpoints: `/api/v1/entities`
- Service runs on port 8085
- Basic repository pattern implementation
- No hierarchical relationships
- No flexible attributes
- No versioning or soft delete

## Target State

The refactored objects-service will be a full-featured taxonomy management system:

### Core Concepts

**Object Types**
- Define taxonomic categories and hierarchies
- Support parent-child relationships (self-referencing)
- Optional concrete table mapping for type-specific storage
- Sealed types to prevent inheritance
- Flexible metadata for type-specific attributes

**Objects**
- Instantiations of object types
- Internal BIGINT ID for performance
- Public UUID for external API usage
- Parent-child object relationships
- Comprehensive audit fields (created_by, updated_by, version)
- Soft delete support (deleted_at)
- Flexible metadata and tags
- Status management (active, inactive, archived, deleted, pending)

### New Schema

```sql
object_types:
  - id (BIGINT PRIMARY KEY)
  - name (VARCHAR UNIQUE)
  - parent_type_id (FK to object_types, nullable)
  - concrete_table_name (VARCHAR, nullable)
  - description (TEXT)
  - is_sealed (BOOLEAN)
  - metadata (JSONB)
  - created_at, updated_at (TIMESTAMP)

objects:
  - id (BIGINT IDENTITY - internal PK)
  - public_id (UUID - external API ID)
  - object_type_id (FK to object_types)
  - parent_object_id (FK to objects, nullable)
  - name (VARCHAR)
  - description (TEXT)
  - created_at, updated_at, deleted_at (TIMESTAMP)
  - version (BIGINT)
  - created_by (VARCHAR)
  - updated_by (VARCHAR)
  - metadata (JSONB)
  - status (VARCHAR)
  - tags (TEXT[])
```

## Key Features

1. **Hierarchical Organization**: Both object types and objects support tree structures
2. **Flexible Attributes**: JSONB metadata and TEXT[] tags for extensible data
3. **Dual ID System**: Internal BIGINT for performance, external UUID for API
4. **Version Control**: Optimistic locking with version field
5. **Soft Delete**: Delete tracking without permanent removal
6. **Comprehensive Audit**: Created/updated by tracking with timestamps
7. **Status Management**: Enum-based status with defined transitions
8. **Type Sealing**: Prevent inheritance from sealed types
9. **Advanced Search**: GIN indexes on JSONB and arrays for fast queries
10. **Concrete Tables**: Optional type-specific table mapping

## Architecture Changes

### Files to Delete
- `internal/models/entity.go`
- `internal/models/entity_test.go`
- `internal/repository/entity_repository.go`
- `internal/repository/entity_repository_test.go`
- `internal/services/entity_service.go`
- `internal/services/entity_service_test.go`
- `internal/handlers/entity_handler.go`
- `internal/handlers/entity_handler_test.go`

### Files to Rewrite
- `migrations/000001_initial.up.sql` - Replace with new schema
- `migrations/000001_initial.down.sql` - Replace with new rollback
- `migrations/dependencies.json` - Update for new schema
- `migrations/environments.json` - Update for development data
- `migrations/development/000003_dev_test_data.up.sql` - Replace with new test data

### Files to Create
- `internal/models/object_type.go` - Object type model
- `internal/models/object.go` - Object model
- `internal/models/object_type_request.go` - DTOs for object types
- `internal/models/object_request.go` - DTOs for objects
- `internal/repository/object_type_repository.go` - Object type repository
- `internal/repository/object_repository.go` - Object repository
- `internal/services/object_type_service.go` - Object type service
- `internal/services/object_service.go` - Object service
- `internal/handlers/object_type_handler.go` - Object type handler
- `internal/handlers/object_handler.go` - Object handler
- `migrations/development/000002_dev_tax_test_data.up.sql` - Development taxonomy data

### Files to Update
- `cmd/main.go` - Wire new handlers

## Documentation Structure

This folder contains the following documents:

1. **README.md** (this file) - Overview and summary
2. **design-questions.md** - Questions that need answers to refine implementation
3. **phases.md** - Detailed 9-phase implementation plan with steps
4. **progress.md** - Progress tracking checklist

## Risk Assessment

**RISK LEVEL: NONE**

The objects-service has not been deployed to production or development environments. There are no existing users, no data migration requirements, and no backward compatibility concerns. All changes can be made freely.

## Estimated Effort

- Phase 1 (Migrations): 2 hours
- Phase 2 (Models): 2.5 hours
- Phase 3 (Repositories): 4 hours
- Phase 4 (Services): 4 hours
- Phase 5 (Handlers): 4 hours
- Phase 6 (Main): 1 hour
- Phase 7 (Test Data): 1 hour
- Phase 8 (Tests): 4 hours
- Phase 9 (Documentation): 2 hours

**Total: ~24.5 hours**

## Getting Started

1. Review **design-questions.md** and answer the questions
2. Review **phases.md** to understand implementation steps
3. Begin with Phase 1 once design questions are resolved
4. Update **progress.md** as you complete each phase
