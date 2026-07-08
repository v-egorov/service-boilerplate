## MODIFIED Requirements

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

**Removed:** The dispatcher no longer maps raw `sql.ErrNoRows` via a fallback case. Repository methods return `repository.ErrNotFound` sentinel directly when rows are not found, eliminating the need for SQL-level error handling in the handler layer.

#### Scenario: Missing object returns 404 not_found

- **WHEN** a client requests `/api/v1/objects/:id` for an ID that does not exist in the database
- **THEN** the repository returns `repository.ErrNotFound`, the service layer wraps it, and the handler dispatches via `HandleError()` which matches `repository.ErrNotFound` and returns HTTP 404 with `"type": "not_found"`

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
