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
