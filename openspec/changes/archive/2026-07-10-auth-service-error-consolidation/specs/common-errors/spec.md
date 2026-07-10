## ADDED Requirements

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
