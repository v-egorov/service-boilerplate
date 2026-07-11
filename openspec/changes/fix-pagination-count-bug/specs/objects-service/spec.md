## MODIFIED Requirements

### Requirement: All list endpoints return consistent pagination metadata

Every objects-service endpoint that returns a paginated collection MUST include a `pagination` object in the response with exactly these fields: `count`, `limit`, `offset`, and `total`. The `count` field is the number of items in this page, `limit` and `offset` are the requested values, and `total` is the total matching rows across all pages.

**CRITICAL:** The `total` value MUST reflect ALL WHERE filters applied to the query (object_type_id, name, status, tags, created_by, parent_object_id, metadata key-value pairs, date ranges). It must NOT be a global count of all non-deleted objects in the table when any filter is active.

#### Scenario: All endpoints include pagination metadata
- **WHEN** a client lists objects, object types, relationships, or relationship types
- **THEN** the response includes `"pagination": {"count": N, "limit": L, "offset": O, "total": T}`

#### Scenario: Relationship-types returns pagination (was missing)
- **WHEN** a client requests `/api/v1/relationship-types?limit=5&offset=0`
- **THEN** the response includes `"pagination": {"count": 2, "limit": 5, "offset": 0, "total": 7}`

#### Scenario: Filtered count matches actual matching rows
- **WHEN** a client requests `/api/v1/objects?object_type_id=1&limit=50&offset=0` and there are exactly 20 objects of type_id=1
- **THEN** the response includes `"pagination": {"count": 20, "limit": 50, "offset": 0, "total": 20}` — NOT a global count

#### Scenario: Multi-filter pagination is accurate
- **WHEN** a client requests `/api/v1/objects?object_type_id=1&status=active` and there are exactly 15 matching objects
- **THEN** the response includes `"pagination": {"count": 15, "limit": 50, "offset": 0, "total": 15}`

## ADDED Requirements

### Requirement: QueryBuilder.BuildCount() preserves WHERE filters

The `QueryBuilder.BuildCount()` method MUST produce a COUNT query that applies all WHERE clauses added via the builder chain (`Where`, `WhereIn`, `WhereTagsContain`, `WhereJsonContains`, `WhereDateRange`). The implementation SHALL replace the SELECT columns with `COUNT(*)` while preserving the FROM table and ALL AND/OR conditions. It must NOT reconstruct a bare `SELECT COUNT(*) FROM <table>` without filters.

#### Scenario: BuildCount with single WHERE filter
- **WHEN** a QueryBuilder has one `.Where("object_type_id = $1", 3)` call
- **THEN** `BuildCount()` returns `"SELECT COUNT(*) FROM objects_service.objects WHERE object_type_id = $1"` with the same args

#### Scenario: BuildCount with multiple chained WHERE clauses
- **WHEN** a QueryBuilder chains `.Where("object_type_id = $1", 3)` then `.Where("deleted_at IS NULL")`
- **THEN** `BuildCount()` returns `"SELECT COUNT(*) FROM objects_service.objects WHERE object_type_id = $1 AND deleted_at IS NULL"` with args[0]=3

#### Scenario: BuildCount with WhereTagsContain
- **WHEN** a QueryBuilder calls `.WhereTagsContain(["tag1", "tag2"])` after an existing WHERE clause
- **THEN** `BuildCount()` includes the tag condition as `"AND ($N = ANY(tags)) OR ($M = ANY(tags))"` in the COUNT query

#### Scenario: BuildCount with WhereJsonContains
- **WHEN** a QueryBuilder calls `.WhereJsonContains("metadata", map[string]interface{}{"key": "value"})`
- **THEN** `BuildCount()` includes `"AND metadata::jsonb @> $N::jsonb"` in the COUNT query

#### Scenario: BuildCount with WhereDateRange
- **WHEN** a QueryBuilder calls `.WhereDateRange("created_at", startTime, endTime)`
- **THEN** `BuildCount()` includes both `"AND created_at >= $N AND created_at <= $M"` conditions in the COUNT query
