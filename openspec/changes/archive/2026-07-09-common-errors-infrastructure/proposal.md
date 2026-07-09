## Why

Each service implements error handling independently — objects-service uses shared sentinels with a dispatcher, user-service uses typed struct errors with per-handler type switches, and auth-service has hardcoded `errorResponse()` calls scattered across handlers. This fragmentation makes it impossible to have consistent error semantics across services and blocks any future cross-service error aggregation or monitoring.

## What Changes

- Create `common/errors/` package with 5 sentinel variables: `ErrNotFound`, `ErrAlreadyExists`, `ErrInvalidInput`, `ErrUnauthorized`, `ErrForbidden`
- Provide these as base errors that services import and use when wrapping domain-specific errors via `%w`
- No breaking changes — existing typed structs, per-handler dispatchers, and hardcoded error responses remain functional during migration

## Capabilities

### New Capabilities
- `common-errors`: Shared sentinel error types for all services to use as the base of their error chains

### Modified Capabilities
(none in this delta — service-level migrations are separate deltas)

## Impact

**New files:**
- `common/errors/infra.go` — ErrNotFound, ErrAlreadyExists, ErrInvalidInput
- `common/errors/auth.go` — ErrUnauthorized, ErrForbidden

**No changes to existing services yet.** Each subsequent migration delta (objects-service, user-service, auth-service) will adopt the package independently. The gateway remains a verbatim HTTP proxy — no translation layer needed.
