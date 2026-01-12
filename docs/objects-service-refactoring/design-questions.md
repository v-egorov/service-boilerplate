# Design Questions

This document contains questions that need to be answered to refine the implementation plan for the objects-service refactoring.

---

## Question 1: Hierarchy Depth Limits

**Question**: Should there be any limits on the depth of hierarchies for object_types and objects?

**Context**:
- Both object_types and objects support parent-child relationships
- Deep hierarchies can impact query performance
- No limit could allow infinite nesting

**Options**:
- A: No limit - allow arbitrary depth
- B: Fixed maximum depth (e.g., 10 levels)
- C: Configurable maximum depth per type

**Recommendation**: Option A - No limit, but add database-level cycle detection triggers to prevent infinite loops.

**Decision**: [ANSWER NEEDED]

---

## Question 2: Type Compatibility Rules for Parent-Child Objects

**Question**: Should there be any constraints on object types when creating parent-child object relationships?

**Context**:
- Objects have parent_object_id linking to other objects
- Each object has an object_type_id
- Parent and child objects may have different types

**Options**:
- A: No restrictions - any object can be parent of any other object
- B: Child must be same type or subtype of parent
- C: Child must be same type as parent
- D: Configurable compatibility rules per type

**Recommendation**: Option A - No restrictions initially for maximum flexibility. Add metadata-based constraints later if needed.

**Decision**: [ANSWER NEEDED]

---

## Question 3: Sealed Type Behavior

**Question**: What should happen when attempting to create an object_type that extends a sealed type?

**Context**:
- `is_sealed` flag on object_types indicates whether type can be extended
- When creating a new object_type with parent_type_id pointing to a sealed type

**Options**:
- A: Allow creation but emit warning
- B: Reject creation with validation error (403/400)
- C: Require special permission/role to extend sealed types
- D: Allow but mark child type as "special" and restrict some operations

**Recommendation**: Option B - Reject creation with clear validation error explaining sealed type cannot be extended.

**Decision**: [ANSWER NEEDED]

---

## Question 4: Public ID Generation Strategy

**Question**: How should public_id (UUID) be generated?

**Context**:
- Internal id is BIGINT IDENTITY (auto-increment)
- Public_id is UUID for external API
- Different UUID versions have different properties

**Options**:
- A: UUID v4 (random) - simple, no guarantees on ordering
- B: UUID v7 (time-ordered) - timestamp-based, sortable
- C: UUID v5 (namespace-based) - deterministic based on object data
- D: ULID - sortable, URL-safe alternative

**Recommendation**: Option B - UUID v7 (time-ordered) because:
- Sortable by creation time
- Better for indexing and range queries
- Monotonically increasing (helps with database performance)
- Still UUID standard format

**Decision**: [ANSWER NEEDED]

---

## Question 5: Metadata Schema Validation

**Question**: How should metadata JSONB field be validated?

**Context**:
- metadata is JSONB for flexible attributes
- Need to balance flexibility with data integrity
- Validation can happen at application or database level

**Options**:
- A: No schema validation - completely flexible
- B: JSON Schema validation per object_type (stored in type's metadata)
- C: Database-level CHECK constraints with JSONB operators
- D: Optional validation mode - can be enabled/disabled per type

**Recommendation**: Option B - JSON Schema validation per object_type stored in the object_type's metadata field. This allows:
- Type-specific schemas defined in object_type.metadata.validation_schema
- Flexible but controlled attributes
- Schema evolution support
- Can be optional (null schema = no validation)

**Implementation**:
```go
type ObjectType struct {
    Metadata map[string]interface{}
    // Metadata can contain:
    // {
    //   "validation_schema": { ... JSON Schema ... },
    //   "allowed_attributes": ["attr1", "attr2"],
    //   "required_attributes": ["attr1"]
    // }
}
```

**Decision**: [ANSWER NEEDED]

---

## Question 6: API Gateway Routing

**Question**: How should API Gateway route requests to objects-service?

**Context**:
- Existing microservices: api-gateway, user-service, auth-service, objects-service
- Current objects-service runs on port 8085 with `/api/v1/entities`
- Need to define new routing paths

**Options**:
- A: `/api/v1/object-types/*` and `/api/v1/objects/*`
- B: `/api/v1/objects/types/*` and `/api/v1/objects/items/*`
- C: `/api/v1/taxonomy/types/*` and `/api/v1/taxonomy/objects/*`
- D: Single prefix `/api/v1/objects/*` with nested routes

**Recommendation**: Option A - Clear separation:
- `/api/v1/object-types/*` - Object type CRUD operations
- `/api/v1/objects/*` - Object CRUD operations

Example endpoints:
- `GET /api/v1/object-types` - List all object types
- `GET /api/v1/object-types/{id}` - Get specific object type
- `POST /api/v1/object-types` - Create object type
- `GET /api/v1/objects` - List objects (with filters)
- `GET /api/v1/objects/{public_id}` - Get object by public_id
- `POST /api/v1/objects` - Create object
- `GET /api/v1/object-types/{id}/objects` - List objects of specific type

**Decision**: [ANSWER NEEDED]

---

## Question 7: Authentication Requirements per Endpoint

**Question**: What authentication/authorization should be required for each endpoint?

**Context**:
- Existing services use JWT tokens
- Role-based access control (RBAC) with roles and permissions
- Different operations may require different permission levels

**Options**:
- A: All endpoints require authenticated user
- B: GET endpoints public, POST/PUT/DELETE require auth
- C: Permission-based - specific permissions per operation type
- D: Configurable - can be set per object_type

**Recommendation**: Option C - Permission-based RBAC:

Required permissions:
- `object_types:read` - View object types
- `object_types:write` - Create/update object types
- `object_types:delete` - Delete object types
- `objects:read` - View objects
- `objects:write` - Create/update objects
- `objects:delete` - Delete objects (soft/hard)
- `objects:admin` - Admin operations (sealed types, type modifications)

Special cases:
- Soft delete: `objects:write`
- Hard delete: `objects:admin`
- Sealed type operations: `object_types:admin`

**Decision**: [ANSWER NEEDED]

---

## Question 8: Batch Operations

**Question**: Should the API support batch operations (bulk create, update, delete)?

**Context**:
- Bulk operations can improve efficiency for many objects
- Increases API complexity and error handling requirements
- Transaction safety considerations

**Options**:
- A: No batch operations - only single item operations
- B: Basic batch operations for create only
- C: Full batch support (create, update, delete) with transaction safety
- D: Configurable - batch support as optional feature

**Recommendation**: Option C - Full batch support with transaction safety:

Endpoints:
- `POST /api/v1/objects/batch` - Bulk create (array of objects)
- `PATCH /api/v1/objects/batch` - Bulk update (array of {id, changes})
- `DELETE /api/v1/objects/batch` - Bulk delete (array of IDs)

Features:
- Single transaction for all operations
- Partial success support (return errors per item)
- Configurable batch size limits
- Async option for very large batches

**Decision**: [ANSWER NEEDED]

---

## Question 9: Search Sophistication Level

**Question**: How sophisticated should object search and filtering capabilities be?

**Context**:
- metadata is JSONB with flexible attributes
- tags is TEXT[] array
- Need efficient search capabilities

**Options**:
- A: Basic filtering (by object_type_id, status, name, tags)
- B: Full JSONB metadata query support (any JSON path)
- C: Indexed search on common metadata fields
- D: Integration with search service (Elasticsearch/Opensearch)

**Recommendation**: Option C - Indexed search with query builder:

Query capabilities:
- Basic filters: object_type_id, status, created_at range, updated_at range
- Tags: exact match, contains any, contains all
- Name/description: full-text search with trigrams
- Metadata: indexed fields (configured in object_type.metadata.indexed_fields)
- Hierarchical: include descendants, level-based filtering

API examples:
```
GET /api/v1/objects?object_type_id=1&status=active
GET /api/v1/objects?tags=red,blue&tags_mode=any
GET /api/v1/objects?name_like=widget
GET /api/v1/objects?metadata.color=red
GET /api/v1/objects?object_type_id=1&include_descendants=true
```

Database indexes:
- GIN index on metadata (jsonb_path_ops)
- GIN index on tags
- BTREE index on (object_type_id, status)
- GIN index on name, description (gin_trgm_ops)

**Decision**: [ANSWER NEEDED]

---

## Question 10: Audit Field Storage Format

**Question**: How should created_by and updated_by fields be stored?

**Context**:
- Audit fields track who made changes
- Need to consider JWT user info, system operations, deleted users

**Options**:
- A: Store user_id (from JWT sub claim)
- B: Store email (from JWT email claim)
- C: Store username (from JWT preferred_username claim)
- D: Store all three in JSON object
- E: Store user reference to user-service

**Recommendation**: Option A - Store user_id (JWT sub claim):

Field format:
- `created_by VARCHAR(255)` - user_id or "system" for automated operations
- `updated_by VARCHAR(255)` - user_id or "system" for automated operations

Special values:
- "system" - For automated/migration operations
- "anonymous" - For operations without authentication (if allowed)

Rationale:
- Simple, stable identifier
- Can lookup user details from user-service if needed
- Survives email/username changes
- Matches existing JWT pattern

**Implementation**:
```go
// Extract from JWT context
userID := c.GetString("user_id")
if userID == "" {
    userID = "system"
}
```

**Decision**: [ANSWER NEEDED]

---

## Summary

All questions need answers before starting implementation. Once answered, update this document with the decisions, then proceed to **phases.md** for the detailed implementation plan.

**Progress**: 0/10 questions answered
