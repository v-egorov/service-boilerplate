# Phase R3: Advanced Features (Future Work)

## Overview

This phase covers advanced features for the relationship system. These are optional enhancements that can be implemented after the core functionality is complete.

## Estimated Time

8-10 hours (optional)

---

## Features

### R3.1 Bulk Relationship Operations

**Description:** Support creating or deleting multiple relationships in a single request.

**API Endpoints:**

```
POST   /api/v1/relationships/bulk-create
DELETE /api/v1/relationships/bulk-delete
```

**Request/Response:**

```json
// POST /api/v1/relationships/bulk-create
Request:
{
  "relationships": [
    {"source_object_id": "...", "target_object_id": "...", "type_key": "contains"},
    {"source_object_id": "...", "target_object_id": "...", "type_key": "contains"}
  ]
}

Response:
{
  "created": 2,
  "failed": 0,
  "errors": []
}
```

**Implementation Notes:**
- Process all operations in a transaction
- Collect errors but continue processing
- Return summary of successes/failures

---

### R3.2 Pagination for Relationship Queries

**Description:** Add proper pagination to all relationship list endpoints.

**Implementation:**
- Add `total` count to all list responses
- Add `has_next`/`has_prev` boolean fields
- Support cursor-based pagination for large datasets

**Response Format:**

```json
{
  "relationships": [...],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 150,
    "total_pages": 8,
    "has_next": true,
    "has_prev": false
  }
}
```

---

### R3.3 Relationship Path Queries

**Description:** Query paths between related objects.

**API Endpoints:**

```
GET /api/v1/objects/:id/paths/:target_id
```

**Query Parameters:**
- `type_key`: Filter by relationship type
- `max_depth`: Maximum path length (default 5)

**Response:**

```json
{
  "source_object_id": "...",
  "target_object_id": "...",
  "paths": [
    {
      "distance": 2,
      "relationships": [
        {"type_key": "contains", "from": "A", "to": "B"},
        {"type_key": "contains", "from": "B", "to": "C"}
      ]
    }
  ]
}
```

**SQL Implementation:**

```sql
WITH RECURSIVE relationship_path AS (
    -- Start from source
    SELECT 
        source_object_id,
        target_object_id,
        relationship_type_id,
        ARRAY[source_object_id] as path,
        1 as depth
    FROM objects_relationships
    WHERE source_object_id = $sourceID
    
    UNION ALL
    
    -- Continue traversing
    SELECT 
        r.source_object_id,
        r.target_object_id,
        r.relationship_type_id,
        rp.path || r.source_object_id,
        rp.depth + 1
    FROM objects_relationships r
    INNER JOIN relationship_path rp ON r.source_object_id = rp.target_object_id
    WHERE rp.depth < $maxDepth
)
SELECT * FROM relationship_path WHERE target_object_id = $targetID;
```

---

### R3.4 Performance Tuning

**Description:** Optimize query performance for large datasets.

**Potential Additions:**

1. **Additional Indexes:**
   - Composite indexes for common query patterns
   - Partial indexes for filtered queries (e.g., only active relationships)

2. **Query Optimization:**
   - Review and optimize slow queries
   - Add query hints if needed

3. **Caching:**
   - Cache relationship type lookups
   - Cache object relationship counts

---

### R3.5 API Enhancements

**Description:** Additional API features for better developer experience.

**Potential Additions:**

1. **Filtering:**
   ```bash
   GET /api/v1/relationships?source_object_id=...&status=active
   ```

2. **Sorting:**
   ```bash
   GET /api/v1/relationships?sort=created_at&order=desc
   ```

3. **Field Selection:**
   ```bash
   GET /api/v1/relationships?fields=public_id,source_object_id,target_object_id
   ```

4. **Expansion:**
   ```bash
   GET /api/v1/relationships?expand=source_object,target_object
   ```

---

### R3.6 Bidirectional Relationship Enhancements

**Description:** More sophisticated handling of bidirectional relationships.

**Potential Features:**

1. **Auto-sync:**
   - Automatically create reverse relationship when forward is created
   - Sync status changes between directions

2. **Shared vs Independent Metadata:**
   - Option to share metadata between directions
   - Option to have separate metadata per direction

3. **Consistency Validation:**
   - Ensure both directions exist
   - Detect and fix inconsistencies

---

### R3.7 Relationship Metadata Schema Validation

**Description:** Validate relationship metadata against a schema.

**Implementation:**

1. **Add schema field to relationship type:**
   ```sql
   ALTER TABLE objects_relationship_types 
   ADD COLUMN metadata_schema JSONB;
   ```

2. **Validate on relationship create/update:**
   ```go
   func ValidateMetadata(schema, metadata json.RawMessage) error {
       // Use JSON Schema validation
   }
   ```

---

### R3.8 Graph Traversal Operations

**Description:** Advanced graph operations on the relationship graph.

**Potential Operations:**

1. **Get all connected objects (any depth):**
   ```bash
   GET /api/v1/objects/:id/connected
   ```

2. **Get objects at specific depth:**
   ```bash
   GET /api/v1/objects/:id/connected?depth=2
   ```

3. **Get common neighbors:**
   ```bash
   GET /api/v1/objects/:id/common-neighbors?other_id=:other_id
   ```

4. **Graph statistics:**
   ```bash
   GET /api/v1/objects/:id/graph-stats
   # Returns: degree, connected_components, etc.
   ```

---

## Priority Order

Recommended implementation order:

1. **R3.2** - Pagination (high value, moderate effort)
2. **R3.1** - Bulk operations (high value, moderate effort)
3. **R3.4** - Performance tuning (as needed)
4. **R3.5** - API enhancements (moderate value)
5. **R3.3** - Path queries (moderate value)
6. **R3.6-R3.8** - Advanced features (lower priority)

---

## Out of Scope

The following are explicitly out of scope for this project:

- Graph databases or specialized graph query engines
- Real-time relationship subscriptions
- Relationship versioning (full audit trail)
- Cross-service relationships (relationships between objects in different services)

---

## References

- [Phase R1: Relationship Types](phase-r1-relationship-types.md)
- [Phase R2: Relationship Instances](phase-r2-relationship-instances.md)
- [Objects Service Refactoring README](../README.md)
