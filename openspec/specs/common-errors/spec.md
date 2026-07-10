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
