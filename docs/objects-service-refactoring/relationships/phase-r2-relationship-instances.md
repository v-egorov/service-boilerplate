# Phase R2: Relationship Instance System

## Overview

This phase implements the relationship instance management system. Relationship instances are actual links between objects, created based on relationship type rules.

## Estimated Time

12-15 hours

## Prerequisites

- Phase R1: Relationship Type System completed
- Relationship type records exist in the database

## Deliverables

1. Database migrations for relationship instances CTI table
2. Go models for relationships
3. Repository layer for CRUD operations and queries
4. Service layer with validation logic
5. HTTP handlers for API endpoints
6. Route registration
7. Unit tests
8. Dev migration for seed data
9. RBAC permissions for relationships
10. End-to-end test script

---

## Task Breakdown

### R2.1 Create Relationship Marker in object_types

**Objective:** Add a special object_type record to mark relationship instances.

**Migration File:** `services/objects-service/migrations/development/000007_add_relationship_marker.up.sql`

```sql
-- Create Relationship marker in object_types (if not exists)
INSERT INTO objects_service.object_types (name, description, created_at, updated_at)
VALUES ('Relationship', 'Marker type for relationship instances', NOW(), NOW())
ON CONFLICT (name) DO NOTHING
RETURNING id;
```

**Down Migration:** `000007_add_relationship_marker.down.sql`

```sql
DELETE FROM objects_service.object_types WHERE name = 'Relationship';
```

### R2.2 Create objects_relationships CTI Table

**Objective:** Create the concrete CTI table for relationship instances.

**Migration File:** `services/objects-service/migrations/development/000008_create_objects_relationships.up.sql`

```sql
CREATE TABLE objects_service.objects_relationships (
    object_id BIGINT PRIMARY KEY REFERENCES objects_service.objects(id) ON DELETE CASCADE,
    source_object_id BIGINT NOT NULL REFERENCES objects_service.objects(id) ON DELETE CASCADE,
    target_object_id BIGINT NOT NULL REFERENCES objects_service.objects(id) ON DELETE CASCADE,
    relationship_type_id BIGINT NOT NULL REFERENCES objects_service.objects(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'active',
    relationship_metadata JSONB DEFAULT '{}',
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_relationship UNIQUE (source_object_id, target_object_id, relationship_type_id)
);

-- Indexes for common queries
CREATE INDEX idx_relationships_source ON objects_service.objects_relationships(source_object_id);
CREATE INDEX idx_relationships_target ON objects_service.objects_relationships(target_object_id);
CREATE INDEX idx_relationships_type_status ON objects_service.objects_relationships(relationship_type_id, status);
CREATE INDEX idx_relationships_type ON objects_service.objects_relationships(relationship_type_id);

-- Self-referencing indexes for circular detection
CREATE INDEX idx_relationships_source_target ON objects_service.objects_relationships(source_object_id, target_object_id);

COMMENT ON TABLE objects_service.objects_relationships IS 
    'CTI concrete table for relationship instances';
```

**Down Migration:** `000008_create_objects_relationships.down.sql`

```sql
DROP TABLE IF EXISTS objects_service.objects_relationships;
```

### R2.3 Add Go Models

**Objective:** Create Go structs for relationship instance data.

**File:** `services/objects-service/internal/models/relationship.go`

```go
package models

import (
	"time"
)

type Relationship struct {
	ObjectID               int64           `json:"object_id" db:"object_id"`
	PublicID              uuid.UUID       `json:"public_id" db:"public_id"`
	SourceObjectID        int64           `json:"source_object_id" db:"source_object_id"`
	SourceObjectPublicID  uuid.UUID       `json:"source_object_public_id" db:"source_object_public_id"`
	TargetObjectID        int64           `json:"target_object_id" db:"target_object_id"`
	TargetObjectPublicID  uuid.UUID       `json:"target_object_public_id" db:"target_object_public_id"`
	RelationshipTypeID    int64           `json:"relationship_type_id" db:"relationship_type_id"`
	RelationshipTypeKey   string          `json:"relationship_type_key" db:"relationship_type_key"`
	Status                string          `json:"status" db:"status"`
	RelationshipMetadata  json.RawMessage `json:"relationship_metadata" db:"relationship_metadata"`
	CreatedBy             *string         `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy             *string         `json:"updated_by,omitempty" db:"updated_by"`
	CreatedAt             time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at" db:"updated_at"`
}

// CreateRelationshipRequest represents creation input
type CreateRelationshipRequest struct {
	SourceObjectPublicID  string          `json:"source_object_id" db:"source_object_id"` // or UUID
	TargetObjectPublicID string          `json:"target_object_id" db:"target_object_id"` // or UUID
	RelationshipTypeKey  string          `json:"type_key" db:"type_key"` // or ID
	Status                string          `json:"status" db:"status"`
	RelationshipMetadata  json.RawMessage `json:"metadata" db:"metadata"`
}

// UpdateRelationshipRequest represents update input
type UpdateRelationshipRequest struct {
	Status               *string         `json:"status,omitempty"`
	RelationshipMetadata json.RawMessage `json:"metadata,omitempty"`
}

// RelationshipFilter for querying
type RelationshipFilter struct {
	SourceObjectID       *int64
	TargetObjectID       *int64
	RelationshipTypeKey *string
	Status              *string
	Page                int
	PageSize            int
}
```

### R2.4 Add Repository Layer

**Objective:** Implement database operations for relationships.

**File:** `services/objects-service/internal/repository/relationship_repository.go`

**Methods to implement:**

```go
type RelationshipRepository interface {
	Create(ctx context.Context, input *models.CreateRelationshipRequest) (*models.Relationship, error)
	GetByID(ctx context.Context, id int64) (*models.Relationship, error)
	GetByPublicID(ctx context.Context, publicID uuid.UUID) (*models.Relationship, error)
	Update(ctx context.Context, id int64, input *models.UpdateRelationshipRequest) (*models.Relationship, error)
	Delete(ctx context.Context, id int64) error
	DeleteByPublicID(ctx context.Context, publicID uuid.UUID) error
	
	// Query methods
	List(ctx context.Context, filter *models.RelationshipFilter) ([]*models.Relationship, error)
	GetForObject(ctx context.Context, objectID int64) ([]*models.Relationship, error)
	GetForObjectByType(ctx context.Context, objectID int64, typeKey string) ([]*models.Relationship, error)
	GetRelatedObjects(ctx context.Context, objectID int64, typeKey *string) ([]*models.Object, error)
	
	// Validation helpers
	Exists(ctx context.Context, sourceID, targetID int64, typeID int64) (bool, error)
	CountForObject(ctx context.Context, objectID int64, typeKey *string) (int, error)
	GetByType(ctx context.Context, typeKey string) ([]*models.Relationship, error)
	
	// Circular detection
	CheckCircular(ctx context.Context, sourceID, targetID, typeID int64) (bool, error)
}
```

**Implementation notes:**
- Use schema-qualified table names
- Use CTE for joins with objects table
- Implement circular detection using recursive CTE

### R2.5 Add Service Layer

**Objective:** Business logic for relationship management with validation.

**File:** `services/objects-service/internal/services/relationship_service.go`

**Methods to implement:**

```go
type RelationshipService interface {
	Create(ctx context.Context, input *models.CreateRelationshipRequest) (*models.Relationship, error)
	GetByPublicID(ctx context.Context, publicID uuid.UUID) (*models.Relationship, error)
	Update(ctx context.Context, publicID uuid.UUID, input *models.UpdateRelationshipRequest) (*models.Relationship, error)
	Delete(ctx context.Context, publicID uuid.UUID) error
	
	// Query methods
	List(ctx context.Context, filter *models.RelationshipFilter) ([]*models.Relationship, error)
	GetForObject(ctx context.Context, objectPublicID uuid.UUID) ([]*models.Relationship, error)
	GetForObjectByType(ctx context.Context, objectPublicID uuid.UUID, typeKey string) ([]*models.Relationship, error)
	GetRelatedObjects(ctx context.Context, objectPublicID uuid.UUID, typeKey *string) ([]*models.Object, error)
}
```

**Validation logic:**

1. **Object existence**: Both source and target objects must exist
2. **Relationship type existence**: Type must exist and be valid
3. **Circular detection**: Prevent circular relationships based on cardinality rules
4. **Cardinality constraints**: Enforce min_count/max_count
5. **Type compatibility**: Optional - can validate if source/target object types are allowed
6. **Duplicate check**: Same source, target, type combination must be unique
7. **Status**: Default to 'active', validate if provided

**Cardinality validation rules:**

| Cardinality | Source→Target | Target→Source |
|-------------|---------------|---------------|
| one_to_one | Must have 0 or 1 | Must have 0 or 1 |
| one_to_many | Unlimited | Must have 0 or 1 |
| many_to_one | Must have 0 or 1 | Unlimited |
| many_to_many | Unlimited | Unlimited |

### R2.6 Add HTTP Handlers

**Objective:** HTTP endpoints for relationship management.

**File:** `services/objects-service/internal/handlers/relationship_handler.go`

**Endpoints:**

```go
type RelationshipHandler interface {
	Create(c *gin.Context)           // POST /api/v1/relationships
	List(c *gin.Context)             // GET /api/v1/relationships
	GetByPublicID(c *gin.Context)   // GET /api/v1/relationships/:public_id
	Update(c *gin.Context)          // PUT /api/v1/relationships/:public_id
	Delete(c *gin.Context)          // DELETE /api/v1/relationships/:public_id
	
	// Object relationships
	GetForObject(c *gin.Context)    // GET /api/v1/objects/:id/relationships
	GetForObjectByType(c *gin.Context) // GET /api/v1/objects/:id/relationships/:type_key
}
```

**Request/Response formats:**

**POST /api/v1/relationships**
```json
Request:
{
  "source_object_id": "550e8400-e29b-41d4-a716-446655440000",
  "target_object_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "type_key": "contains",
  "status": "active",
  "metadata": {}
}

Response (201):
{
  "object_id": 200,
  "public_id": "550e8400-e29b-41d4-a716-446655440000",
  "source_object_id": 100,
  "source_object_public_id": "...",
  "target_object_id": 101,
  "target_object_public_id": "...",
  "relationship_type_id": 50,
  "relationship_type_key": "contains",
  "status": "active",
  "relationship_metadata": {},
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

**GET /api/v1/objects/:id/relationships**
```json
Response (200):
{
  "relationships": [...],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 100
  }
}
```

### R2.7 Register Routes

**Objective:** Add relationship routes to the router.

**File:** `services/objects-service/internal/router.go`

```go
relationships := r.Group("/api/v1/relationships")
{
    relationships.POST("", handler.CreateRelationship)
    relationships.GET("", handler.ListRelationships)
    relationships.GET("/:public_id", handler.GetRelationshipByPublicID)
    relationships.PUT("/:public_id", handler.UpdateRelationship)
    relationships.DELETE("/:public_id", handler.DeleteRelationship)
}

// Object relationships
objects := r.Group("/api/v1/objects")
{
    objects.GET("/:public_id/relationships", handler.GetRelationshipsForObject)
    objects.GET("/:public_id/relationships/:type_key", handler.GetRelationshipsForObjectByType)
}
```

### R2.8 Implement Validation Logic

**Objective:** Service-level validation for relationships.

**Circular Detection:**

```go
// CheckCircular checks if creating this relationship would create a cycle
// For hierarchical relationships (parent_of), we must prevent cycles
func (s *RelationshipService) CheckCircular(ctx context.Context, sourceID, targetID int64, typeID int64) (bool, error) {
    // Get relationship type to check if it's hierarchical
    relType, err := s.relationshipTypeRepo.GetByID(ctx, typeID)
    if err != nil {
        return false, err
    }
    
    // Only check for hierarchical types (parent_of, contains, etc.)
    // Many-to-many can have cycles without issue
    if relType.Cardinality != models.CardinalityOneToMany && 
       relType.Cardinality != models.CardinalityManyToOne {
        return false, nil
    }
    
    // Check if target is already an ancestor of source
    return s.repo.CheckCircular(ctx, sourceID, targetID, typeID)
}
```

**SQL for circular detection:**
```sql
WITH RECURSIVE relationship_path AS (
    -- Start from target
    SELECT target_object_id, 1 as depth
    FROM objects_relationships
    WHERE source_object_id = $sourceID
      AND relationship_type_id = $typeID
    
    UNION
    
    -- Traverse further
    SELECT r.target_object_id, rp.depth + 1
    FROM objects_relationships r
    INNER JOIN relationship_path rp ON r.source_object_id = rp.target_object_id
    WHERE r.relationship_type_id = $typeID
      AND rp.depth < 100  -- Prevent infinite recursion
)
SELECT EXISTS(SELECT 1 FROM relationship_path WHERE target_object_id = $targetID);
```

**Cardinality Validation:**

```go
// ValidateCardinality checks if adding this relationship would violate cardinality constraints
func (s *RelationshipService) ValidateCardinality(ctx context.Context, sourceID, targetID int64, typeKey string) error {
    relType, err := s.relationshipTypeRepo.GetByTypeKey(ctx, typeKey)
    if err != nil {
        return err
    }
    
    // Check source cardinality
    sourceCount, err := s.repo.CountForObject(ctx, sourceID, &typeKey)
    if err != nil {
        return err
    }
    
    if relType.MaxCount >= 0 && sourceCount >= relType.MaxCount {
        return ErrCardinalityViolation // source has max relationships
    }
    
    // For many_to_one, also check target cardinality
    if relType.Cardinality == models.CardinalityManyToOne {
        targetCount, err := s.repo.CountForObject(ctx, targetID, nil)
        if err != nil {
            return err
        }
        if targetCount >= 1 {
            return ErrCardinalityViolation // target already has a source
        }
    }
    
    return nil
}
```

### R2.9 Add Query Methods

**Objective:** Additional repository methods for querying relationships.

**Get Related Objects:**
```go
// GetRelatedObjects returns objects related to the given object
// Optionally filtered by relationship type
func (r *RelationshipRepository) GetRelatedObjects(ctx context.Context, objectID int64, typeKey *string) ([]*models.Object, error) {
    // Query that joins relationships to objects
    // Return objects that are targets (or sources) of relationships
}
```

**Get for Object:**
```go
// GetForObject returns all relationships for an object (as source or target)
func (r *RelationshipRepository) GetForObject(ctx context.Context, objectID int64) ([]*models.Relationship, error) {
    query := `
        SELECT r.*, o.public_id as object_public_id
        FROM objects_relationships r
        JOIN objects o ON r.object_id = o.id
        WHERE r.source_object_id = $1 OR r.target_object_id = $1
        ORDER BY r.created_at DESC
    `
}
```

### R2.10 Add Unit Tests ✅ COMPLETED

**Objective:** Test relationship functionality.

**Files:**
- `services/objects-service/internal/services/relationship_service_test.go`
- `services/objects-service/internal/services/relationship_type_service_test.go` (mocks)

**Test Cases (51 total):**
- Create: 21 tests (success, validation, circular detection, cardinality, errors)
- GetByPublicID: 2 tests (success, not found)
- Update: 3 tests (success, not found, with metadata)
- Delete: 3 tests (success, not found, database error)
- List: 5 tests (success, filters, pagination, nil filter, empty)
- GetForObject: 6 tests (success, source, target, filters, pagination)
- GetForObjectByType: 3 tests (success, not found, no matches)
- GetRelatedObjects: 6 tests (source, target, type, bidirectional, error)

### R2.11 Dev Migration: Seed Test Relationships

**Objective:** Add development test data for relationships.

**Migration File:** `services/objects-service/migrations/development/000009_seed_relationships.up.sql`

```sql
-- Create sample objects for relationship testing
-- This assumes there are existing objects to create relationships between

-- Example: Create relationships between test objects
-- 1. First, ensure we have test objects (or use existing ones)
-- 2. Create relationship instances

-- Create base objects for relationships
INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    (SELECT id FROM objects_service.object_types WHERE name = 'Relationship'),
    'Relationship: Portfolio A contains Asset X',
    'active',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM objects_service.objects o
    JOIN objects_service.object_types ot ON o.object_type_id = ot.id
    WHERE ot.name = 'Relationship'
    AND o.name = 'Relationship: Portfolio A contains Asset X'
);

-- Get the relationship type ID for "contains"
-- Get object IDs for source and target
-- Create the relationship

INSERT INTO objects_service.objects_relationships (
    object_id, source_object_id, target_object_id, relationship_type_id,
    status, relationship_metadata, created_at, updated_at
)
SELECT 
    o.id,
    (SELECT id FROM objects_service.objects WHERE name = 'Test Portfolio A' LIMIT 1),
    (SELECT id FROM objects_service.objects WHERE name = 'Test Asset X' LIMIT 1),
    (SELECT ort.object_id FROM objects_service.objects_relationship_types ort WHERE ort.type_key = 'contains'),
    'active',
    '{}',
    NOW(),
    NOW()
FROM objects_service.objects o
JOIN objects_service.object_types ot ON o.object_type_id = ot.id
WHERE ot.name = 'Relationship'
AND o.name = 'Relationship: Portfolio A contains Asset X'
ON CONFLICT DO NOTHING;
```

**Down Migration:** `000009_seed_relationships.down.sql`

```sql
DELETE FROM objects_service.objects_relationships;
DELETE FROM objects_service.objects 
WHERE object_type_id = (SELECT id FROM objects_service.object_types WHERE name = 'Relationship');
```

### R2.12 End-to-End Test Script

**Objective:** Create shell script to test the full relationship system.

**File:** `scripts/test-relationships-e2e.sh`

```bash
#!/bin/bash
set -e

BASE_URL="${BASE_URL:-http://localhost:8085}"

echo "=== Relationship System E2E Tests ==="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

pass() { echo -e "${GREEN}✓ PASS${NC}: $1"; }
fail() { echo -e "${RED}✗ FAIL${NC}: $1"; }

# Test 1: Create relationship type
echo "Test 1: Create relationship type"
RESP=$(curl -s -X POST "$BASE_URL/api/v1/relationship-types" \
  -H "Content-Type: application/json" \
  -d '{"type_key": "test_contains", "relationship_name": "test contains", "reverse_type_key": "test_contained_by", "cardinality": "one_to_many"}')
  
if echo "$RESP" | grep -q "type_key"; then
    pass "Created relationship type"
else
    fail "Failed to create relationship type: $RESP"
    exit 1
fi

# Test 2: List relationship types
echo "Test 2: List relationship types"
RESP=$(curl -s "$BASE_URL/api/v1/relationship-types")
if echo "$RESP" | grep -q "test_contains"; then
    pass "Listed relationship types"
else
    fail "Failed to list: $RESP"
fi

# Test 3: Create relationship
echo "Test 3: Create relationship"
# (Requires existing objects - would need to set up test data first)

echo ""
echo "=== Tests Complete ==="
```

---



### R2.13 RBAC Permissions ✅ COMPLETED

**Objective:** Add permissions for relationship instance management and assign to roles.

**Migration Files:**
- `services/auth-service/migrations/{development,staging,production}/000008_add_relationships_permissions.up.sql`
- `services/auth-service/migrations/{development,staging,production}/000008_add_relationships_permissions.down.sql`

**Permissions Added:**
- `relationships:create` - Create new relationship instances
- `relationships:read` - Read relationship instances  
- `relationships:update` - Update relationship instances
- `relationships:delete` - Delete relationship instances

```sql
-- Add relationships permissions
INSERT INTO auth_service.permissions (name, resource, action) VALUES
    ('relationships:create', 'relationships', 'create'),
    ('relationships:read', 'relationships', 'read'),
    ('relationships:update', 'relationships', 'update'),
    ('relationships:delete', 'relationships', 'delete')
ON CONFLICT (name) DO NOTHING;
```

**Down Migration:** `000008_add_relationships_permissions.down.sql`

```sql
DELETE FROM auth_service.permissions WHERE name LIKE 'relationships:%';
```

### R2.14 Assign Permissions to Roles ✅ COMPLETED

**Objective:** Assign relationship permissions to admin and object-type-admin roles.

**Migration Files:**
- `services/auth-service/migrations/{development,staging,production}/000009_assign_relationships_permissions.up.sql`
- `services/auth-service/migrations/{development,staging,production}/000009_assign_relationships_permissions.down.sql`

**Roles Assigned:**
- `admin` role - All 4 relationship permissions (create, read, update, delete)
- `object-type-admin` role - All 4 relationship permissions (create, read, update, delete)

**Note:** Routes already have permission middleware configured in `cmd/main.go` (lines 305-334).

```sql
-- Assign relationships permissions to admin role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'admin'
  AND p.name LIKE 'relationships:%'
ON CONFLICT DO NOTHING;

-- Assign relationships permissions to object-type-admin role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'object-type-admin'
  AND p.name LIKE 'relationships:%'
ON CONFLICT DO NOTHING;
```

**Down Migration:** `000009_assign_relationships_permissions.down.sql`

```sql
-- Remove relationships permissions from roles
DELETE FROM auth_service.role_permissions
WHERE permission_id IN (
    SELECT id FROM auth_service.permissions WHERE name LIKE 'relationships:%'
);
```

### R2.15 End-to-End RBAC Test Script

**Objective:** Create shell script to test relationship instance RBAC.

**File:** `scripts/test-rbac-relationships.sh`

```bash
#!/bin/bash

# RBAC Test Script for Relationships
# Tests permission-based access control for relationships endpoints
# Usage: ./scripts/test-rbac-relationships.sh [--keep-data]

BASE_URL="${BASE_URL:-http://localhost:8080}"

# Test users
ADMIN_EMAIL="dev.admin@example.com"
ADMIN_PASSWORD="devadmin123"
OBJECT_ADMIN_EMAIL="object.admin@example.com"
OBJECT_ADMIN_PASSWORD="devadmin123"
TEST_USER_EMAIL="test.user@example.com"
TEST_USER_PASSWORD="devadmin123"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Login and get tokens
login() {
    local email=$1
    local password=$2
    curl -s -X POST "$BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\": \"$email\", \"password\": \"$password\"}" | \
        jq -r '.access_token'
}

# Test relationship permissions
ADMIN_TOKEN=$(login "$ADMIN_EMAIL" "$ADMIN_PASSWORD")
OBJECT_ADMIN_TOKEN=$(login "$OBJECT_ADMIN_EMAIL" "$OBJECT_ADMIN_PASSWORD")
TEST_USER_TOKEN=$(login "$TEST_USER_EMAIL" "$TEST_USER_PASSWORD")

# RL-1: object.admin CREATE relationship → 201
curl -s -X POST "$BASE_URL/api/v1/relationships" \
    -H "Authorization: Bearer $OBJECT_ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"source_object_id": "...", "target_object_id": "...", "type_key": "contains"}'

# RL-2: test.user CREATE relationship → 403
# RL-3: test.user READ relationships → 200
# ... (similar to relationship-types test pattern)
```

---

## RBAC Permissions Reference

### Required Permissions

| Permission | Description |
|------------|-------------|
| relationships:create | Create new relationship instances |
| relationships:read | List/get relationship instances |
| relationships:update | Update relationship instances |
| relationships:delete | Delete relationship instances |

### Role Assignments

| Role | relationships:create | relationships:read | relationships:update | relationships:delete |
|------|----------------------|-------------------|---------------------|---------------------|
| admin | ✅ | ✅ | ✅ | ✅ |
| object-type-admin | ✅ | ✅ | ✅ | ✅ |
| user | ❌ | ✅ | ❌ | ❌ |

---

## Implementation Order

1. **Database migrations first** (R2.1, R2.2)
2. **Models** (R2.3)
3. **Repository** (R2.4)
4. **Service** (R2.5)
5. **Handlers** (R2.6)
6. **Routes** (R2.7)
7. **Validation logic** (R2.8)
8. **Query methods** (R2.9)
9. **Tests** (R2.10)
10. **Dev seed data** (R2.11)
11. **RBAC permissions** (R2.13)
12. **E2E script** (R2.12)

---

## API Reference

### Create Relationship

**Endpoint:** `POST /api/v1/relationships`

**Request Body:**
```json
{
  "source_object_id": "uuid (required)",
  "target_object_id": "uuid (required)",
  "type_key": "string (required)",
  "status": "string (optional, default active)",
  "metadata": object (optional)
}
```

**Response (201):** Returns created relationship object

**Errors:**
- 400: Invalid request body
- 404: Source/target object not found
- 404: Relationship type not found
- 409: Duplicate relationship
- 422: Validation error (circular, cardinality)

### List Relationships

**Endpoint:** `GET /api/v1/relationships`

**Query Parameters:**
- `source_object_id`: Filter by source
- `target_object_id`: Filter by target
- `type_key`: Filter by type
- `status`: Filter by status
- `page`: Page number
- `page_size`: Items per page

**Response (200):** Array of relationship objects

### Get Relationship

**Endpoint:** `GET /api/v1/relationships/:public_id`

**Response (200):** Relationship object

**Errors:**
- 404: Not found

### Update Relationship

**Endpoint:** `PUT /api/v1/relationships/:public_id`

**Request Body:**
```json
{
  "status": "string (optional)",
  "metadata": object (optional)"
}
```

**Response (200):** Updated relationship object

### Delete Relationship

**Endpoint:** `DELETE /api/v1/relationships/:public_id`

**Response (204):** No content

### Get Relationships for Object

**Endpoint:** `GET /api/v1/objects/:public_id/relationships`

**Response (200):**
```json
{
  "relationships": [...],
  "pagination": {...}
}
```

### Get Relationships for Object by Type

**Endpoint:** `GET /api/v1/objects/:public_id/relationships/:type_key`

**Response (200):** Array of relationships of that type

---

## Validation Rules Detail

### Circular Detection

Circular detection is performed for hierarchical relationships:
- `one_to_many`: Check if target is ancestor of source
- `many_to_one`: Check if target is descendant of source
- `many_to_many`: No circular check needed (allowed)

### Cardinality Enforcement

When creating a relationship:

1. **one_to_one**: Both source and target must have 0 or 1 relationship of this type
2. **one_to_many**: Source can have unlimited, target must have 0 or 1
3. **many_to_one**: Source must have 0 or 1, target can have unlimited
4. **many_to_many**: No restrictions

### Status Values

| Status | Description |
|--------|-------------|
| active | Relationship is currently valid |
| inactive | Relationship is temporarily disabled |
| deprecated | Relationship type is deprecated |

---

## Error Handling

### Service Layer Errors

```go
var (
	ErrRelationshipNotFound        = errors.New("relationship not found")
	ErrDuplicateRelationship       = errors.New("relationship already exists")
	ErrSourceObjectNotFound        = errors.New("source object not found")
	ErrTargetObjectNotFound        = errors.New("target object not found")
	ErrRelationshipTypeNotFound    = errors.New("relationship type not found")
	ErrCircularRelationship       = errors.New("creating this relationship would create a cycle")
	ErrCardinalityViolation       = errors.New("relationship cardinality constraint violated")
	ErrRelationshipTypeIncompatible = errors.New("source/target objects incompatible with relationship type")
)
```

---

## Testing Checklist

- [ ] Create relationship with valid data
- [ ] Create duplicate relationship returns 409
- [ ] Create with non-existent source object returns 404
- [ ] Create with non-existent target object returns 404
- [ ] Create with non-existent type returns 404
- [ ] Create circular relationship returns 422
- [ ] Create violating cardinality returns 422
- [ ] Update relationship status successfully
- [ ] Delete relationship successfully
- [ ] List relationships with filters
- [ ] Get relationships for object returns correct data
- [ ] Get relationships by type returns correct data
- [ ] Bidirectional relationship lookup works

---

## Performance Considerations

### Indexes

Indexes created:
- `idx_relationships_source`: For queries by source
- `idx_relationships_target`: For queries by target
- `idx_relationships_type_status`: For filtered queries
- `idx_relationships_type`: For type lookups
- `idx_relationships_source_target`: For circular detection

### Query Optimization

- Use CTE for joins with objects table
- Paginate large result sets
- Consider caching frequently accessed relationships

---

## Definition of Done

**All items must be verified before phase is marked complete.**

### Required Verification

- [x] Migrations applied successfully to database
- [x] Database schema verified: `objects_relationships` table exists with correct columns
- [x] Seed data verified: relationships created between test objects
- [ ] API endpoint: POST creates new relationship
- [ ] API endpoint: GET lists relationships with filters
- [ ] API endpoint: GET by public_id returns relationship
- [ ] API endpoint: DELETE removes relationship
- [ ] API endpoint: GET `/objects/{id}/relationships` returns relationships for object
- [ ] API endpoint: GET `/objects/{id}/relationships/{type_key}` returns by type
- [ ] Validation: circular detection prevents cycles
- [ ] Validation: cardinality constraints enforced
- [ ] Unit tests pass
- [ ] Code compiles without errors

### Implementation Checklist

- [x] R2.1 Create Relationship marker migration
- [x] R2.2 Create objects_relationships CTI table
- [x] R2.3 Add Go models
- [x] R2.4 Add repository layer
- [x] R2.5 Add service layer
- [x] R2.6 Add HTTP handlers
- [x] R2.7 Register routes
- [x] R2.8 Implement validation logic
- [x] R2.9 Add query methods
- [x] R2.10 Add unit tests
- [x] R2.11 Dev migration: seed relationships
- [x] R2.13 RBAC permissions migration
- [x] R2.14 Assign permissions to roles
- [ ] R2.15 End-to-end RBAC test script

---

## Next Steps

After completing Phase R2, consider Phase R3: Advanced Features.

## References

- [Phase R1: Relationship Types](phase-r1-relationship-types.md)
- [Objects Service Refactoring README](../README.md)
