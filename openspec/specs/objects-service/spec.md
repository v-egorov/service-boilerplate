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

