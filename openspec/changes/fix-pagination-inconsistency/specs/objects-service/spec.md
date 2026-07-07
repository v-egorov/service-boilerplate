# Delta: objects-service pagination standardization

## ADDED Requirements

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
