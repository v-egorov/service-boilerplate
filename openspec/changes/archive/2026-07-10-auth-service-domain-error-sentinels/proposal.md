## Why

Auth-service's error handling is "mixed" — it uses sentinels for auth failures (ErrUnauthorized) and repo-level not-found errors, but falls back to raw strings and inline handler detection for business rule violations. This creates two problems:

1. **HTTP 500 bug**: Two service-layer returns ("cannot delete role", "cannot delete permission") are plain strings that fall through HandleAuthError's default case → clients get HTTP 500 instead of the correct status code (should be 422 or similar).

2. **Handler knows about PostgreSQL internals**: Two handler methods (CreateRole, CreatePermission) use `strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "23505")` to detect DB constraint violations inline — bypassing HandleAuthError entirely and returning hardcoded JSON responses. This is fragile (breaks if PG error format changes) and violates the principle that handlers should only dispatch, not inspect error contents.

Objects-service solved this by defining domain-specific sentinels in its service package (`ErrRelationshipTypeInUse`, `ErrCircularRelationship`, etc.), each with a specific HTTP status and message matched by HandleError(). Auth-service should follow the same pattern.

## What Changes

- Define two domain-specific sentinel errors in auth-service's services package:
  - `ErrRoleInUse` — role has assigned users (DeleteRole business rule)
  - `ErrPermissionInUse` — permission has roles assigned (DeletePermission business rule)
- Wrap the two raw-string returns with these sentinels via `%w`
- Add matching cases to HandleAuthError dispatcher → HTTP 422 (unprocessable entity, business rule violation)
- Move DB duplicate detection from handlers to service layer:
  - CreateRole: detect `strings.Contains(err.Error(), "duplicate key")` after repo.CreateRole call → wrap with a new sentinel or use common/errors.ErrAlreadyExists
  - CreatePermission: same pattern for repo.CreatePermission
- Add corresponding HandleAuthError cases for the duplicate-key sentinels

## Capabilities

### Modified Capabilities

- **common-errors**: ADD auth-service domain-specific error sentinel documentation. The sentinel package already provides ErrUnauthorized, ErrNotFound, and ErrAlreadyExists — this change documents how auth-service extends with its own domain sentinels in the service layer (following objects-service pattern).

## Impact

- **Affected code**: `auth-service/internal/services/auth_service.go` (define 2 new sentinels + wrap existing returns), `auth-service/internal/handlers/auth_handler.go` (HandleAuthError dispatcher expansion)
- **API behavior change**: Two endpoints that currently return HTTP 500 will now return HTTP 422 with a structured error type. This is correct — these are business rule violations, not server errors.
- **Breaking change**: No — the new status codes match what the API should have returned all along
