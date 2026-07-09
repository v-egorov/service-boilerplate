# Specification: common-errors

## Purpose

Define the shared sentinel error package (`common/errors`) that provides five error types for cross-service error handling. All services import this package and use these sentinels as the base of their error chains via `%w` wrapping, enabling uniform matching with `errors.Is()` across all service layers.

## Requirements

### Requirement: Common error package provides shared sentinel variables for cross-service error handling

The `common/errors` package MUST provide five sentinel error variables that all services import and use as the base of their error chains. Services SHALL wrap domain-specific errors with these sentinels via `%w` so that downstream handlers can match them using `errors.Is()`.

| Sentinel | HTTP Status | Error Type | Description |
|----------|-------------|------------|-------------|
| `common/errors.ErrNotFound` | 404 | `not_found` | Resource not found in the database |
| `common/errors.ErrAlreadyExists` | 409 | `conflict` | Resource with same identifier already exists |
| `common/errors.ErrInvalidInput` | 400 | `validation_error` | Request contains invalid or missing required fields |
| `common/errors.ErrUnauthorized` | 401 | `unauthorized` | Authentication failed — invalid or expired credentials |
| `common/errors.ErrForbidden` | 403 | `permission_denied` | User lacks permission to perform the requested action |

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

#### Scenario: Sentinel variables are simple fmt.Errorf values

- **WHEN** the common/errors package is compiled
- **THEN** each sentinel is defined as a package-level variable using `fmt.Errorf("...")`, making them compatible with Go's error wrapping and `errors.Is()` semantics
