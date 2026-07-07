# Delta: objects-service empty list response contract

## ADDED Requirements

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
