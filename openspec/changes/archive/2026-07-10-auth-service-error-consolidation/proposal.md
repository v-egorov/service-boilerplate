## Why

Auth-service's service layer returns all errors as raw strings (`fmt.Errorf("invalid credentials")`, `fmt.Errorf("token not found: %w", err)`) with no semantic error type. The handler catches every service error and hardcodes a single HTTP status code per method — regardless of what the actual error is. This means "failed to generate token" (server error) and "invalid credentials" (client error) both get the same status code, determined by which method called rather than what went wrong. Consolidating auth errors into shared sentinels enables proper error classification.

## What Changes

- Wrap service-layer infrastructure errors with `common/errors.ErrNotFound` or `common/errors.ErrInvalidInput` via `%w`
- Wrap authentication failures (invalid credentials, revoked tokens) with `common/errors.ErrUnauthorized`
- Extract shared `HandleAuthError(err, requestID)` dispatcher on AuthHandler that maps sentinels to HTTP status codes
- Replace ~76 inline `errorResponse()`/`validationError()` calls in auth_handler.go with the dispatcher
- Re-export local typed error types from `models/errors.go` as aliases to common/errors (same pattern as user-service)

## Capabilities

### Modified Capabilities

- `common-errors`: ADD `ErrUnauthorized` and `ErrForbidden` sentinel usage documentation for auth-domain errors. These were already defined in the package but never adopted by any service layer.
- `auth-service-permission-resolution`: No requirement changes — error handling is orthogonal to permission semantics. (Leave empty)

## Impact

- **Affected code**: `common/errors/auth.go` (imports updated), `services/auth-service/internal/services/auth_service.go` (~50 error returns wrapped with sentinels), `services/auth-service/internal/handlers/auth_handler.go` (extracted dispatcher, replaced ~76 inline calls), `services/auth-service/internal/models/errors.go` (re-export aliases)
- **API behavior change**: Some errors may now return different HTTP status codes (e.g., service-layer "failed to get user roles" was always 500; if it wrapped ErrNotFound via sql.ErrNoRows detection, it would correctly return 404). Most existing error paths remain the same status code.
- **Breaking change**: No — response format unchanged, only classification accuracy improves
