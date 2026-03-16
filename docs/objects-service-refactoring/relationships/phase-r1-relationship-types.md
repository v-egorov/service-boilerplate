# Phase R1: Relationship Type System

## Overview

This phase implements the relationship type management system. Relationship types define the rules for relationships between objects (e.g., "contains", "belongs_to", "references").

## Estimated Time

8-10 hours

## Prerequisites

- Objects-service database schema from previous phases
- Migration orchestrator configured for objects-service

## Deliverables

1. Database migrations for relationship types CTI table
2. Go models for relationship types
3. Repository layer for CRUD operations
4. Service layer with validation logic
5. HTTP handlers for API endpoints
6. Route registration
7. Unit tests
8. Dev migration for seed data

---

## Task Breakdown

### R1.1 Create Relationship Type Marker in object_types

**Objective:** Add a special object_type record to mark relationship types.

**Migration File:** `services/objects-service/migrations/development/000004_dev_add_relationship_type_marker.up.sql`

```sql
-- Create RelationshipType marker in object_types (if not exists)
INSERT INTO objects_service.object_types (name, description, created_at, updated_at)
VALUES ('RelationshipType', 'Marker type for relationship type instances', NOW(), NOW())
ON CONFLICT (name) DO NOTHING
RETURNING id;
```

**Down Migration:** `000004_dev_add_relationship_type_marker.down.sql`

```sql
DELETE FROM objects_service.object_types WHERE name = 'RelationshipType';
```

### R1.2 Create objects_relationship_types CTI Table

**Objective:** Create the concrete CTI table for relationship types.

**Migration File:** `services/objects-service/migrations/development/000005_dev_create_objects_relationship_types.up.sql`

```sql
CREATE TABLE objects_service.objects_relationship_types (
    object_id BIGINT PRIMARY KEY REFERENCES objects_service.objects(id) ON DELETE CASCADE,
    type_key VARCHAR(100) NOT NULL UNIQUE,
    relationship_name VARCHAR(255),
    reverse_type_key VARCHAR(100) NULL,
    cardinality VARCHAR(20) NOT NULL DEFAULT 'many_to_many',
    required BOOLEAN DEFAULT FALSE,
    min_count INTEGER DEFAULT 0,
    max_count INTEGER DEFAULT -1,
    validation_rules JSONB DEFAULT '{}',
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for reverse lookup
CREATE INDEX idx_relationship_types_reverse_key 
    ON objects_service.objects_relationship_types(reverse_type_key);

-- Index for cardinality filter
CREATE INDEX idx_relationship_types_cardinality 
    ON objects_service.objects_relationship_types(cardinality);

COMMENT ON TABLE objects_service.objects_relationship_types IS 
    'CTI concrete table for relationship type instances';
```

**Down Migration:** `000005_dev_create_objects_relationship_types.down.sql`

```sql
DROP TABLE IF EXISTS objects_service.objects_relationship_types;
```

### R1.3 Add Go Models

**Objective:** Create Go structs for relationship type data.

**File:** `services/objects-service/internal/models/relationship_type.go`

```go
package models

import (
	"time"
)

type RelationshipType struct {
	ObjectID          int64           `json:"object_id" db:"object_id"`
	TypeKey           string          `json:"type_key" db:"type_key"`
	RelationshipName  string          `json:"relationship_name" db:"relationship_name"`
	ReverseTypeKey    *string         `json:"reverse_type_key,omitempty" db:"reverse_type_key"`
	Cardinality       string          `json:"cardinality" db:"cardinality"`
	Required          bool            `json:"required" db:"required"`
	MinCount          int             `json:"min_count" db:"min_count"`
	MaxCount          int             `json:"max_count" db:"max_count"`
	ValidationRules   json.RawMessage `json:"validation_rules" db:"validation_rules"`
	CreatedBy         *string         `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy         *string         `json:"updated_by,omitempty" db:"updated_by"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
}

// Valid cardinality values
const (
	CardinalityOneToOne   = "one_to_one"
	CardinalityOneToMany  = "one_to_many"
	CardinalityManyToOne  = "many_to_one"
	CardinalityManyToMany = "many_to_many"
)

// Valid statuses
const (
	StatusActive     = "active"
	StatusInactive   = "inactive"
	StatusDeprecated = "deprecated"
)

// CreateRelationshipTypeRequest represents creation input
type CreateRelationshipTypeRequest struct {
	TypeKey          string          `json:"type_key" db:"type_key"`
	RelationshipName string          `json:"relationship_name" db:"relationship_name"`
	ReverseTypeKey  *string         `json:"reverse_type_key,omitempty" db:"reverse_type_key"`
	Cardinality     string          `json:"cardinality" db:"cardinality"`
	Required        bool            `json:"required" db:"required"`
	MinCount        int             `json:"min_count" db:"min_count"`
	MaxCount        int             `json:"max_count" db:"max_count"`
	ValidationRules json.RawMessage `json:"validation_rules" db:"validation_rules"`
}

// UpdateRelationshipTypeRequest represents update input
type UpdateRelationshipTypeRequest struct {
	RelationshipName *string         `json:"relationship_name,omitempty"`
	ReverseTypeKey   *string         `json:"reverse_type_key,omitempty"`
	Cardinality      *string         `json:"cardinality,omitempty"`
	Required        *bool            `json:"required,omitempty"`
	MinCount        *int             `json:"min_count,omitempty"`
	MaxCount        *int             `json:"max_count,omitempty"`
	ValidationRules json.RawMessage `json:"validation_rules,omitempty"`
}

// RelationshipTypeFilter for listing
type RelationshipTypeFilter struct {
	Cardinality string
	Required    *bool
	Page        int
	PageSize    int
}
```

### R1.4 Add Repository Layer

**Objective:** Implement database operations for relationship types.

**File:** `services/objects-service/internal/repository/relationship_type_repository.go`

**Methods to implement:**

```go
type RelationshipTypeRepository interface {
	Create(ctx context.Context, input *models.CreateRelationshipTypeRequest) (*models.RelationshipType, error)
	GetByID(ctx context.Context, id int64) (*models.RelationshipType, error)
	GetByTypeKey(ctx context.Context, typeKey string) (*models.RelationshipType, error)
	Update(ctx context.Context, id int64, input *models.UpdateRelationshipTypeRequest) (*models.RelationshipType, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter *models.RelationshipTypeFilter) ([]*models.RelationshipType, error)
	Exists(ctx context.Context, typeKey string) (bool, error)
	GetByReverseTypeKey(ctx context.Context, reverseKey string) (*models.RelationshipType, error)
}
```

**Implementation notes:**
- Use schema-qualified table names: `objects_service.objects_relationship_types`
- Use database tracing wrappers (see existing repository patterns)
- Follow existing repository patterns for query building

### R1.5 Add Service Layer

**Objective:** Business logic for relationship type management.

**File:** `services/objects-service/internal/services/relationship_type_service.go`

**Methods to implement:**

```go
type RelationshipTypeService interface {
	Create(ctx context.Context, input *models.CreateRelationshipTypeRequest) (*models.RelationshipType, error)
	GetByTypeKey(ctx context.Context, typeKey string) (*models.RelationshipType, error)
	Update(ctx context.Context, typeKey string, input *models.UpdateRelationshipTypeRequest) (*models.RelationshipType, error)
	Delete(ctx context.Context, typeKey string) error
	List(ctx context.Context, filter *models.RelationshipTypeFilter) ([]*models.RelationshipType, error)
}
```

**Validation logic:**
1. **type_key uniqueness**: Check no existing type with same type_key
2. **reverse_type_key exists**: If provided, must reference existing type_key
3. **cardinality valid**: Must be one of valid values
4. **min_count <= max_count**: If both set, min must not exceed max

### R1.6 Add HTTP Handlers

**Objective:** HTTP endpoints for relationship type management.

**File:** `services/objects-service/internal/handlers/relationship_type_handler.go`

**Endpoints:**

```go
type RelationshipTypeHandler interface {
	Create(c *gin.Context)      // POST /api/v1/relationship-types
	List(c *gin.Context)        // GET /api/v1/relationship-types
	GetByTypeKey(c *gin.Context) // GET /api/v1/relationship-types/:type_key
	Update(c *gin.Context)     // PUT /api/v1/relationship-types/:type_key
	Delete(c *gin.Context)     // DELETE /api/v1/relationship-types/:type_key
}
```

**Request/Response formats:**

**POST /api/v1/relationship-types**
```json
Request:
{
  "type_key": "contains",
  "relationship_name": "contains",
  "reverse_type_key": "contained_by",
  "cardinality": "one_to_many",
  "required": false,
  "min_count": 0,
  "max_count": -1,
  "validation_rules": {}
}

Response (201):
{
  "object_id": 100,
  "type_key": "contains",
  "relationship_name": "contains",
  "reverse_type_key": "contained_by",
  "cardinality": "one_to_many",
  "required": false,
  "min_count": 0,
  "max_count": -1,
  "validation_rules": {},
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

### R1.7 Register Routes

**Objective:** Add relationship type routes to the router.

**File:** `services/objects-service/internal/router.go` (or similar)

```go
relationshipTypes := r.Group("/api/v1/relationship-types")
{
    relationshipTypes.POST("", handler.CreateRelationshipType)
    relationshipTypes.GET("", handler.ListRelationshipTypes)
    relationshipTypes.GET("/:type_key", handler.GetRelationshipTypeByKey)
    relationshipTypes.PUT("/:type_key", handler.UpdateRelationshipType)
    relationshipTypes.DELETE("/:type_key", handler.DeleteRelationshipType)
}
```

### R1.8 Add Unit Tests

**Objective:** Test relationship type functionality.

**Files:**
- `services/objects-service/internal/services/relationship_type_service_test.go`
- `services/objects-service/internal/repository/relationship_type_repository_test.go` (optional, if testing repo directly)

**Testing patterns to follow:**
- Manual mock structs (see `object_type_service_test.go`)
- Testify assertions
- Test both success and error cases

**Test cases:**
- Create relationship type - success
- Create duplicate type_key - error
- Create with invalid reverse_type_key - error
- Create with invalid cardinality - error
- Update relationship type - success
- Delete relationship type - success
- List relationship types - success
- List with filters - success

### R1.9 Dev Migration: Seed Relationship Types

**Objective:** Add development test data for relationship types.

**Migration File:** `services/objects-service/migrations/development/000006_dev_seed_relationship_types.up.sql`

```sql
-- Seed standard relationship types
-- Note: First create the objects entries, then the relationship type entries

-- Step 1: Create base objects for each relationship type
INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'RelationshipType'),
    'Relationship: contains',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects o
    JOIN objects_service.object_types ot ON o.object_type_id = ot.id
    WHERE ot.name = 'RelationshipType'
    AND o.name = 'Relationship: contains'
);

-- Repeat for other types: belongs_to, references, parent_of, depends_on

-- Step 2: Create relationship type entries
INSERT INTO objects_service.objects_relationship_types (
    object_id, type_key, relationship_name, reverse_type_key, 
    cardinality, required, min_count, max_count, validation_rules,
    created_at, updated_at
)
SELECT 
    o.id,
    'contains',
    'contains',
    'contained_by',
    'one_to_many',
    false,
    0,
    -1,
    '{}',
    NOW(),
    NOW()
FROM objects_service.objects o
JOIN objects_service.object_types ot ON o.object_type_id = ot.id
WHERE ot.name = 'RelationshipType' AND o.name = 'Relationship: contains'
ON CONFLICT (type_key) DO NOTHING;

-- Add more types:
-- - belongs_to / owns / many_to_one
-- - references / NULL / many_to_many
-- - parent_of / child_of / one_to_many
-- - depends_on / NULL / many_to_many
```

**Down Migration:** `000006_dev_seed_relationship_types.down.sql`

```sql
DELETE FROM objects_service.objects_relationship_types;
DELETE FROM objects_service.objects 
WHERE object_type_id = (SELECT id FROM objects_service.object_types WHERE name = 'RelationshipType');
```

---

## Implementation Order

1. **Database migrations first** (R1.1, R1.2)
2. **Models** (R1.3)
3. **Repository** (R1.4)
4. **Service** (R1.5)
5. **Handlers** (R1.6)
6. **Routes** (R1.7)
7. **Tests** (R1.8)
8. **Dev seed data** (R1.9)

---

## API Reference

### Create Relationship Type

**Endpoint:** `POST /api/v1/relationship-types`

**Request Body:**
```json
{
  "type_key": "string (required, unique)",
  "relationship_name": "string (optional, defaults to type_key)",
  "reverse_type_key": "string (optional, must reference existing type_key)",
  "cardinality": "string (required): one_to_one|one_to_many|many_to_one|many_to_many",
  "required": "boolean (optional, default false)",
  "min_count": "integer (optional, default 0)",
  "max_count": "integer (optional, default -1 for unlimited)",
  "validation_rules": "object (optional, default {})"
}
```

**Response (201):** Returns created relationship type object

**Errors:**
- 400: Invalid request body
- 409: Duplicate type_key
- 422: Invalid reverse_type_key or cardinality

### List Relationship Types

**Endpoint:** `GET /api/v1/relationship-types`

**Query Parameters:**
- `cardinality`: Filter by cardinality
- `required`: Filter by required flag
- `page`: Page number (default 1)
- `page_size`: Items per page (default 20)

**Response (200):** Array of relationship type objects

### Get Relationship Type

**Endpoint:** `GET /api/v1/relationship-types/:type_key`

**Response (200):** Relationship type object

**Errors:**
- 404: type_key not found

### Update Relationship Type

**Endpoint:** `PUT /api/v1/relationship-types/:type_key`

**Request Body:** Same as create, all fields optional

**Response (200):** Updated relationship type object

**Errors:**
- 404: type_key not found
- 422: Invalid reverse_type_key

### Delete Relationship Type

**Endpoint:** `DELETE /api/v1/relationship-types/:type_key`

**Response (204):** No content

**Errors:**
- 404: type_key not found
- 409: Relationship type in use

---

## Dependencies and Configuration

### Migration Dependencies

Update `services/objects-service/migrations/development/dependencies.json`:

```json
{
  "000004": {
    "description": "Add RelationshipType marker to object_types",
    "depends_on": ["000003"],
    "affects_tables": ["objects_service.object_types"],
    "estimated_duration": "10s",
    "risk_level": "low",
    "rollback_safe": true,
    "environment": "development"
  },
  "000005": {
    "description": "Create objects_relationship_types CTI table",
    "depends_on": ["000004"],
    "affects_tables": ["objects_service.objects_relationship_types"],
    "estimated_duration": "30s",
    "risk_level": "medium",
    "rollback_safe": true,
    "environment": "development"
  },
  "000006": {
    "description": "Seed development relationship types",
    "depends_on": ["000005"],
    "affects_tables": ["objects_service.objects", "objects_service.objects_relationship_types"],
    "estimated_duration": "20s",
    "risk_level": "low",
    "rollback_safe": true,
    "environment": "development"
  }
}
```

Update `services/objects-service/migrations/development/environments.json`:

```json
{
  "development": {
    "migrations": [
      "development/000004_dev_add_relationship_type_marker.up.sql",
      "development/000005_dev_create_objects_relationship_types.up.sql",
      "development/000006_dev_seed_relationship_types.up.sql"
    ]
  }
}
```

---

## Validation Rules Detail

### Cardinality

| Value | Description | Example |
|-------|-------------|---------|
| one_to_one | Each A has one B | Person ↔ Passport |
| one_to_many | Each A has many B | Department → Employees |
| many_to_one | Many A map to one B | Employee → Department |
| many_to_many | Many A relate to many B | Student ↔ Course |

### Reverse Type Key

When setting reverse_type_key:
- Must reference an existing type_key
- The referenced type should have this type as its reverse
- Optional - if NULL, relationship is unidirectional

### Count Constraints

- `min_count`: Minimum relationships of this type an object must have
- `max_count`: Maximum relationships (-1 = unlimited)
- If max_count = 0, no relationships of this type allowed
- Validation runs when creating/deleting relationships

---

## Error Handling

### Service Layer Errors

```go
var (
	ErrRelationshipTypeNotFound    = errors.New("relationship type not found")
	ErrDuplicateRelationshipType   = errors.New("relationship type already exists")
	ErrInvalidCardinality          = errors.New("invalid cardinality value")
	ErrInvalidReverseType          = errors.New("reverse type key does not exist")
	ErrRelationshipTypeInUse      = errors.New("relationship type is in use and cannot be deleted")
	ErrInvalidCountConstraint      = errors.New("min_count cannot exceed max_count")
)
```

### HTTP Status Codes

- 200: Success
- 201: Created
- 204: No Content (deleted)
- 400: Bad Request (invalid input)
- 404: Not Found
- 409: Conflict (duplicate/in use)
- 422: Unprocessable Entity (validation error)
- 500: Internal Server Error

---

## Testing Checklist

- [ ] Create relationship type with valid data
- [ ] Create duplicate type_key returns 409
- [ ] Create with invalid reverse_type_key returns 422
- [ ] Create with invalid cardinality returns 422
- [ ] Create with min_count > max_count returns 422
- [ ] Update relationship type successfully
- [ ] Update to duplicate type_key returns 409
- [ ] Delete relationship type successfully
- [ ] Delete type in use returns 409
- [ ] List relationship types with filters
- [ ] Get by type_key returns correct object

---

## Next Steps

After completing Phase R1, proceed to Phase R2: Relationship Instance System.

## References

- [Phase 10: Class Table Inheritance](../phase-10-class-table-inheritance.md)
- [Objects Service Refactoring README](../README.md)
