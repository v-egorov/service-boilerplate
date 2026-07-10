# Specification: common-errors

## Purpose

Define the shared error package (`common/errors`) that provides both sentinel error variables AND typed error structs for cross-service error handling. Sentinels serve as the base of error chains via `%w` wrapping; typed structs carry structured data (field names, resource identifiers) that handlers render into JSON response fields.
## Requirements
### Requirement: Common error package provides shared sentinel variables and typed error structs for cross-service error handling

The `common/errors` package MUST provide both sentinel error variables AND typed error structs that all services import. Sentinels serve as the base of error chains via `%w` wrapping; typed structs carry structured data (field names, resource identifiers) that handlers render into JSON response fields.

| Sentinel | HTTP Status | Error Type | Description |
|----------|-------------|------------|-------------|
| `common/errors.ErrNotFound` | 404 | `not_found` | Resource not found in the database |
| `common/errors.ErrAlreadyExists` | 409 | `conflict` | Resource with same identifier already exists |
| `common/errors.ErrInvalidInput` | 400 | `validation_error` | Request contains invalid or missing required fields |
| `common/errors.ErrUnauthorized` | 401 | `unauthorized` | Authentication failed — invalid or expired credentials |
| `common/errors.ErrForbidden` | 403 | `permission_denied` | User lacks permission to perform the requested action |

**Typed error structs** (ADDED):

| Struct | HTTP Status | Fields | Description |
|--------|-------------|--------|-------------|
| `ValidationError` | 400 | `Field string`, `Message string` | Validation failure on a specific field |
| `ConflictError` | 409 | `Resource string`, `Field string`, `Value string` | Resource conflict (duplicate) |
| `NotFoundError` | 404 | `Resource string`, `Field string`, `Value string` | Resource not found with identifiers |
| `InternalError` | 500 | `Operation string`, `Err error` + `Unwrap()` | Internal server error wrapping a cause |

Each typed struct MUST have an `Error() string` method and constructor helper functions:
- `NewValidationError(field, message string) ValidationError`
- `NewConflictError(resource, field, value string) ConflictError`
- `NewNotFoundError(resource, field, value string) NotFoundError`
- `NewInternalError(operation string, err error) InternalError`

Typed structs are used by service layers that need to carry structured metadata (field names, resource types) into handler responses. Handlers match these via Go's type-switch (`switch e := err.(type)`), not `errors.Is()`.

#### Scenario: ErrNotFound is importable by all services

- **WHEN** any service imports `"github.com/v-egorov/service-boilerplate/common/errors"`
- **THEN** it can reference `errors.ErrNotFound`, `errors.ErrAlreadyExists`, and `errors.ErrInvalidInput` as sentinel variables usable with `errors.Is()`

#### Scenario: ErrUnauthorized and ErrForbidden are importable by auth-relevant services

- **WHEN** any service imports `"github.com/v-egorov/service-boilerplate/common/errors"`
- **THEN** it can reference `errors.ErrUnauthorized` and `errors.ErrForbidden` as sentinel variables usable with `errors.Is()`

#### Scenario: Services wrap domain errors with common sentinels via %w

- **WHEN** an objects-service repository method encounters a missing row (`sql.ErrNoRows`)
- **THEN** it wraps the result: `fmt.Errorf("failed to get object: %w", common/errors.ErrNotFound)` so that handler dispatchers can match via `errors.Is()`

#### Scenario: User-service typed structs coexist with common sentinels

- **WHEN** a user-service service layer returns `models.NotFoundError{Resource: "user", Field: "email"}`
- **THEN** the type-switch in the handler still matches on the struct type for rich metadata, while `errors.Is()` can match any wrapped `common/errors.ErrNotFound` at deeper layers

#### Scenario: Typed errors coexist with sentinel errors in common package

- **WHEN** a service needs both infra-level signaling (ErrNotFound) and domain-level metadata (NotFoundError with field info)
- **THEN** it can use sentinels for error chain base types and typed structs where structured response data is required

#### Scenario: InternalError supports error unwrapping

- **WHEN** `errors.As(err, &cause)` is called on an InternalError
- **THEN** the underlying wrapped error is returned via `Unwrap()`, enabling inspection of the root cause

#### Scenario: Sentinel variables are simple fmt.Errorf values

- **WHEN** the common/errors package is compiled
- **THEN** each sentinel is defined as a package-level variable using `fmt.Errorf("...")`, making them compatible with Go's error wrapping and `errors.Is()` semantics

### Requirement: Auth-domain services use ErrUnauthorized for authentication failures

Auth-domain service layers MUST return `common/errors.ErrUnauthorized` (wrapped via `%w`) for all authentication-related errors, including but not limited to: invalid credentials, expired tokens, revoked tokens, unauthenticated requests. This enables handler dispatchers to match auth failures uniformly and return HTTP 401 regardless of the specific failure reason.

The following service-layer error messages SHALL use `ErrUnauthorized`:
- "invalid credentials" → `fmt.Errorf("...: %w", errors.ErrUnauthorized)`
- "token not found" / "refresh token not found" → `fmt.Errorf("...: %w", errors.ErrUnauthorized)`
- "token has been revoked" → `fmt.Errorf("...: %w", errors.ErrUnauthorized)`
- "invalid token" / "invalid refresh token" → `fmt.Errorf("...: %w", errors.ErrUnauthorized)`

#### Scenario: Invalid credentials returns ErrUnauthorized sentinel

- **WHEN** a service method encounters invalid username/password during login
- **THEN** it returns `fmt.Errorf("login failed: %w", common/errors.ErrUnauthorized)` so that handler dispatchers can match via `errors.Is(err, errors.ErrUnauthorized)` and return HTTP 401

#### Scenario: Revoked token returns ErrUnauthorized sentinel

- **WHEN** a service method detects that a refresh token has been revoked
- **THEN** it returns `fmt.Errorf("token expired: %w", common/errors.ErrUnauthorized)` so that handler dispatchers match the same sentinel as invalid credentials (both are auth failures)

### Requirement: Auth-domain services use ErrNotFound for lookup failures

Auth-domain service layers MUST return `common/errors.ErrNotFound` (wrapped via `%w`) when a database query fails to find a requested resource. This distinguishes "resource not found" from other infrastructure errors and enables handlers to return HTTP 404 instead of defaulting to 500.

#### Scenario: Missing user returns ErrNotFound sentinel

- **WHEN** a service method queries for a user that does not exist in the database
- **THEN** it detects `sql.ErrNoRows` and returns `fmt.Errorf("...: %w", common/errors.ErrNotFound)` so that handler dispatchers return HTTP 404

### Requirement: Auth-domain services use ErrInvalidInput for malformed requests

Auth-domain service layers MUST return `common/errors.ErrInvalidInput` (wrapped via `%w`) when request data fails validation at the service layer (not at the request parsing stage, which returns a different error type). This covers cases where business rules reject input that passed basic JSON parsing.

#### Scenario: Invalid role name returns ErrInvalidInput sentinel

- **WHEN** a service method validates and rejects a role name that violates naming constraints
- **THEN** it returns `fmt.Errorf("...: %w", common/errors.ErrInvalidInput)` so that handler dispatchers return HTTP 400

### Requirement: Auth-service repository methods detect sql.ErrNoRows at the boundary

The auth-service repository layer MUST detect `sql.ErrNoRows` in all single-row GET methods and return `common/errors.ErrNotFound` directly (unwrapped). This follows Rule 7 of `docs/go-coding-conventions.md` and ensures consistent error flow across services.

Every `Get*()` method that retrieves a single row SHALL check for missing rows before returning:

```go
func (r *AuthRepository) GetRole(ctx context.Context, roleID uuid.UUID) (*models.Role, error) {
    err := r.db.QueryRow(ctx, query, id).Scan(...)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, repository.ErrNotFound  // ← unwrapped sentinel
        }
        return nil, fmt.Errorf("failed to get role: %w", err)
    }
}
```

#### Scenario: GetRole returns ErrNotFound for missing role

- **WHEN** `GetRole` queries for a role that does not exist in the database
- **THEN** it detects `sql.ErrNoRows` and returns `repository.ErrNotFound` (unwrapped sentinel) so that service/handler layers can match via `errors.Is()` or direct comparison

#### Scenario: GetPermission returns ErrNotFound for missing permission

- **WHEN** `GetPermission` queries for a permission ID that does not exist
- **THEN** it detects `sql.ErrNoRows` and returns `repository.ErrNotFound` (unwrapped sentinel)

#### Scenario: GetAuthTokenByHash returns ErrNotFound for missing token

- **WHEN** `GetAuthTokenByHash` queries for an auth token by hash that does not exist
- **THEN** it detects `sql.ErrNoRows` and returns `repository.ErrNotFound` (unwrapped sentinel)

#### Scenario: GetRoleByName returns ErrNotFound for unknown role name

- **WHEN** `GetRoleByName` queries for a role with a name that does not exist
- **THEN** it detects `sql.ErrNoRows` and returns `repository.ErrNotFound` (unwrapped sentinel)

#### Scenario: Collection-returning methods do NOT apply this rule

- **WHEN** a repository method returns a slice (`[]*models.Role`, etc.)
- **THEN** it does NOT check for sql.ErrNoRows — empty results are represented as an empty non-nil slice, not an error (Rule 1 convention)

### Requirement: Auth-service defines domain-specific sentinel errors for business rule violations

Auth-service service layer MUST define domain-specific sentinel error variables for business rule conditions that cannot be expressed with generic sentinels (ErrUnauthorized, ErrNotFound, etc.). These follow the same pattern as objects-service's service-layer sentinels (`ErrRelationshipTypeInUse`, `ErrCircularRelationship`, etc.) and are matched by HandleAuthError via `errors.Is()`.

The following domain-specific sentinels SHALL be defined in `auth-service/internal/services/auth_service.go`:

| Sentinel | HTTP Status | Description |
|----------|-------------|-------------|
| `ErrRoleInUse` | 422 (unprocessable) | Attempt to delete a role that has assigned users |
| `ErrPermissionInUse` | 422 (unprocessable) | Attempt to delete a permission that has roles assigned |

Each sentinel is defined as:
```go
var ErrRoleInUse = errors.New("role has assigned users and cannot be deleted")
var ErrPermissionInUse = errors.New("permission has assigned roles and cannot be deleted")
```

Service layer methods wrap these with `%w` when returning:
```go
return fmt.Errorf("cannot delete role: %d users are assigned to this role", userCount)
// BECOMES:
return fmt.Errorf("role deletion blocked: %w", ErrRoleInUse)
```

#### Scenario: DeleteRole returns ErrRoleInUse via errors.Is() matching

- **WHEN** `DeleteRole` is called for a role that has 3 assigned users
- **THEN** it detects the constraint and returns: `fmt.Errorf("role deletion blocked: %w", ErrRoleInUse)` so that HandleAuthError dispatcher can match via `errors.Is()` and return HTTP 422

#### Scenario: DeletePermission returns ErrPermissionInUse via errors.Is() matching

- **WHEN** `DeletePermission` is called for a permission that has 5 roles assigned
- **THEN** it detects the constraint and returns: `fmt.Errorf("permission deletion blocked: %w", ErrPermissionInUse)` so that HandleAuthError dispatcher can match via `errors.Is()` and return HTTP 422

#### Scenario: HandleAuthError dispatches domain sentinels to correct HTTP status codes

- **WHEN** a service-layer error wrapped with `ErrRoleInUse` reaches the handler
- **THEN** HandleAuthError matches it via `errors.Is(err, services.ErrRoleInUse)` and returns HTTP 422 Unprocessable Entity with error type "validation_error" — NOT HTTP 500

