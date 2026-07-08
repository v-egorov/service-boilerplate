# Specification: objects-service

## Purpose

Define the behavior of the objects-service QueryBuilder SQL query builder (specifically WHERE clause placeholder handling) and the List endpoint filtering rules, ensuring correct multi-parameter queries and proper permission-based access control without implicit self-filtering.
## Requirements
### Requirement: QueryBuilder Where() uses distinct placeholder indices per call
The QueryBuilder.Where() method MUST generate distinct `$N` placeholders for each invocation, starting from the current argIndex value and incrementing sequentially. Each argument appended to args[] must correspond to exactly one unique placeholder in the generated SQL string. This ensures that when multiple WHERE clauses are chained (e.g., object_type_id filter + user ID filter), PostgreSQL receives correctly typed values for each parameter without positional collision.

#### Scenario: Two Where() calls produce $1 and $2
- **WHEN** a QueryBuilder has two consecutive .Where("column = $1", value) calls
- **THEN** the generated SQL contains "WHERE col1 = $1 AND col2 = $2" with args[0]=value1, args[1]=value2

#### Scenario: Three Where() calls produce distinct indices across mixed methods
- **WHEN** a QueryBuilder chains Where(), then WhereJsonContains(), then WhereDateRange()
- **THEN** all placeholders in the final SQL are unique ($1 through $N with no gaps or duplicates) and args[] length matches placeholder count

#### Scenario: Mixed filter query does not cause type mismatch errors
- **WHEN** a List request includes both object_type_id (bigint) and user ID (varchar UUID) filters
- **THEN** the generated SQL has "WHERE object_type_id = $1 AND created_by = $2" with args[0]=3, args[1]="uuid-string", and PostgreSQL executes without SQLSTATE 42883 errors

### Requirement: QueryBuilder WhereTagsContain() uses OR logic per tag
The QueryBuilder.WhereTagsContain() method MUST use AND for the first tag and OR for each subsequent tag, each consuming exactly one placeholder index. The argIndex must increment by 1 per tag (not by len(tags)).

#### Scenario: Single tag produces AND clause
- **WHEN** WhereTagsContain(["tag1"]) is called on a query that already has WHERE
- **THEN** the SQL contains "AND ($1 = ANY(tags))" with args[0]="tag1"

#### Scenario: Multiple tags produce AND ... OR chain
- **WHEN** WhereTagsContain(["tag1", "tag2"]) is called
- **THEN** the SQL contains "AND ($1 = ANY(tags)) OR ($2 = ANY(tags))" with correct arg indices and args array

### Requirement: List endpoint does not apply implicit ownership filtering
The List endpoints (object_handler.go List() and relationship_handler.go List()) MUST NOT unconditionally add `filter.UserID` based on the authenticated user's identity. Permission enforcement is handled at the gateway level via permiddleware. If clients need ownership-based filtering, they SHALL pass an explicit query parameter (future enhancement).

#### Scenario: Unauthenticated request with no filters returns all accessible objects
- **WHEN** a request arrives without X-User-ID header and no filter parameters
- **THEN** the List endpoint queries all non-deleted objects without any created_by filter

#### Scenario: Authenticated request respects only explicit query parameters
- **WHEN** a request carries X-User-ID but no object_type_id or name filter
- **THEN** the generated SQL does NOT include "WHERE created_by = ..." — permission checks are enforced by permiddleware, not by self-filtering in the handler

#### Scenario: Explicit ownership filter can be added via query parameter (future)
- **WHEN** a client passes ?owner_id=<uuid> query parameter to List
- **THEN** the generated SQL includes "WHERE created_by = $N" for that specific owner value

### Requirement: List endpoints return JSON array for empty results
Every objects-service endpoint that returns a paginated collection (`data` field in the response) MUST serialize an empty result set as `[]` (JSON array), never `null` (JSON null). This applies to all list, search, tree traversal, and hierarchical methods across object types, relationships, and relationship types.

#### Scenario: List objects returns empty array when no match
- **WHEN** a request queries objects with an offset beyond the total count (e.g., `?object_type_id=3&limit=5&offset=100`)
- **THEN** the response includes `"data": []` in the JSON body, not `"data": null`

#### Scenario: List object types returns empty array when no match
- **WHEN** a request queries object types with a prefix that matches nothing (e.g., `?type_key_prefix=nonexistent_xyz`)
- **THEN** the response includes `"data": []` in the JSON body, not `"data": null`

#### Scenario: List relationships returns empty array when no match
- **WHEN** a request queries relationships with filters that match nothing (e.g., `?from_type_key=nonexistent&to_type_key=also_nonexistent`)
- **THEN** the response includes `"data": []` in the JSON body, not `"data": null`

#### Scenario: Tree traversal methods return empty array for missing nodes
- **WHEN** a request queries descendants/ancestors/path for an object type that does not exist or has no children
- **THEN** each method returns `"data": []` in the JSON body, not `"data": null`

### Requirement: Repository layer initializes slices as empty, never nil
All repository methods that return collection types (`[]*Model`) MUST initialize their result variable using `make([]*Model, 0)` instead of `var items []*Model`. This ensures the Go zero value is an empty slice (which serializes to JSON `[]`), not a nil slice (which serializes to JSON `null`).

#### Scenario: Empty query path produces non-nil slice
- **WHEN** a repository method executes a query that returns zero rows
- **THEN** the returned slice is non-nil and has length 0 (`len(result) == 0 && result != nil`)

### Requirement: All list endpoints use limit/offset query parameters consistently
Every objects-service endpoint that returns a paginated collection MUST accept `limit` and `offset` as query parameters. No endpoint shall use `page` or `page_size` parameters. This applies to all list, search, and hierarchical methods across object types, relationships, and relationship types.

#### Scenario: Objects list uses limit/offset
- **WHEN** a client requests `/api/v1/objects?limit=5&offset=10`
- **THEN** the endpoint returns 5 objects starting from offset 10

#### Scenario: Relationships list uses limit/offset (was page/page_size)
- **WHEN** a client requests `/api/v1/relationships?limit=5&offset=10`
- **THEN** the endpoint returns 5 relationships starting from offset 10, not via `page=` or `page_size=` parameters

#### Scenario: Relationship types list uses limit/offset (was page/page_size)
- **WHEN** a client requests `/api/v1/relationship-types?limit=5&offset=10`
- **THEN** the endpoint returns 5 relationship types starting from offset 10, not via `page=` or `page_size=` parameters

#### Scenario: GetForObject uses limit/offset (was page/page_size)
- **WHEN** a client requests `/api/v1/objects/public-id/:id/relationships?limit=3&offset=0`
- **THEN** the endpoint returns 3 relationships starting from offset 0

### Requirement: All list endpoints return consistent pagination metadata
Every objects-service endpoint that returns a paginated collection MUST include a `pagination` object in the response with exactly these fields: `count`, `limit`, `offset`, and `total`. The `count` field is the number of items in this page, `limit` and `offset` are the requested values, and `total` is the total matching rows across all pages.

#### Scenario: All endpoints include pagination metadata
- **WHEN** a client lists objects, object types, relationships, or relationship types
- **THEN** the response includes `"pagination": {"count": N, "limit": L, "offset": O, "total": T}`

#### Scenario: Relationship-types returns pagination (was missing)
- **WHEN** a client requests `/api/v1/relationship-types?limit=5&offset=0`
- **THEN** the response includes `"pagination": {"count": 2, "limit": 5, "offset": 0, "total": 7}`

### Requirement: Default limit is 50 for all list endpoints when not specified
Every objects-service list endpoint MUST default to `limit=50` and `offset=0` when the client does not provide pagination parameters. No endpoint shall use a different default (e.g., page_size=20).

#### Scenario: Missing pagination params defaults to limit 50, offset 0
- **WHEN** a client requests `/api/v1/relationships` without pagination parameters
- **THEN** the endpoint returns up to 50 relationships starting from offset 0 (not 20)

### Requirement: Handler error dispatcher maps sentinel errors to correct HTTP status codes

The objects-service handler layer MUST include a shared `HandleError` function that dispatches service-layer and repository-layer sentinel errors to appropriate HTTP status codes. Every request path that calls into the service layer MUST use this dispatcher instead of defaulting to HTTP 500 for all errors.

The dispatcher SHALL map the following error types:

| Sentinel Error | HTTP Status | Error Type |
|---------------|-------------|------------|
| `repository.ErrNotFound` | 404 | `not_found` |
| `services.*ErrNotFound` (all service variants) | 404 | `not_found` |
| `repository.ErrAlreadyExists` / `services.*ErrDuplicate*` | 409 | `conflict` |
| `repository.ErrOptimisticLock` / `repository.ErrVersionConflict` | 409 | `conflict` |
| `services.ErrCircularRelationship` / `services.ErrCardinalityViolation` | 422 | `validation_error` |
| `repository.ErrInvalidInput` / `services.*ErrTypeKeyRequired` / `services.*ErrCardinalityRequired` | 400 | `validation_error` |
| All other errors (unknown) | 500 | `internal_error` |

#### Scenario: Missing object returns 404 not_found

- **WHEN** a client requests `/api/v1/objects/:id` for an ID that does not exist in the database
- **THEN** the handler dispatches via `HandleError()` which matches `repository.ErrNotFound` (wrapped by service layer) and returns HTTP 404 with `"type": "not_found"`

#### Scenario: Circular relationship returns 422 validation_error

- **WHEN** a client attempts to create a relationship that would form a cycle
- **THEN** the handler dispatches via `HandleError()` which matches `services.ErrCircularRelationship` and returns HTTP 422 with `"type": "validation_error"`

#### Scenario: Duplicate resource returns 409 conflict

- **WHEN** a client attempts to create an object type with a `type_key` that already exists
- **THEN** the handler dispatches via `HandleError()` which matches `services.ErrTypeKeyExists` and returns HTTP 409 with `"type": "conflict"`

#### Scenario: Unknown error returns 500 internal_error

- **WHEN** a database connection failure or unexpected panic occurs during request processing
- **THEN** the handler dispatches via `HandleError()` which falls through to default case and returns HTTP 500 with `"type": "internal_error"`

#### Scenario: All four handlers use the same dispatcher function

- **WHEN** any of the four objects-service handlers (`ObjectHandler`, `ObjectTypeHandler`, `RelationshipHandler`, `RelationshipTypeHandler`) encounters a service-layer error
- **THEN** each handler calls the shared `HandleError()` from `internal/handlers/error.go` (not a per-handler method) and receives consistent status code mapping

