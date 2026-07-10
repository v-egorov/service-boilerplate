## MODIFIED Requirements

### Requirement: Common error package provides shared sentinel variables and typed error structs for cross-service error handling

The `common/errors` package MUST provide both sentinel error variables AND typed error structs that all services import. Sentinels serve as the base of error chains via `%w` wrapping; typed structs carry structured data (field names, resource identifiers) that handlers render into JSON response fields.

**Sentinel errors** (existing):

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

#### Scenario: Common errors package provides typed struct definitions

- **WHEN** any service imports `"github.com/v-egorov/service-boilerplate/common/errors"`
- **THEN** it can use `common/errors.NewNotFoundError("user", "email", "test@example.com")` to create a NotFoundError with structured fields

#### Scenario: Typed errors carry field-level metadata for JSON responses

- **WHEN** user-service returns `errors.NewValidationError("password", "must be at least 6 characters")`
- **THEN** the handler renders it as `{ "error": "...", "type": "validation_error", "field": "password" }` — the Field struct member populates the JSON response

#### Scenario: Typed errors coexist with sentinel errors in common package

- **WHEN** a service needs both infra-level signaling (ErrNotFound) and domain-level metadata (NotFoundError with field info)
- **THEN** it can use sentinels for error chain base types and typed structs where structured response data is required

#### Scenario: InternalError supports error unwrapping

- **WHEN** `errors.As(err, &cause)` is called on an InternalError
- **THEN** the underlying wrapped error is returned via `Unwrap()`, enabling inspection of the root cause
