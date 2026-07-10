## Why

User-service and auth-service both define identical typed error structs (`ValidationError`, `ConflictError`, `NotFoundError`, `InternalError`) in their local `models/errors.go` files. This duplication means the same four types exist three times across the codebase (user-service, auth-service, plus a near-duplicate pattern in objects-service's repository). Moving these shared infra errors to `common/errors/typed.go` eliminates duplication and establishes a single source of truth for cross-service error handling.

## What Changes

- Create `common/errors/typed.go` with four typed error structs (`ValidationError`, `ConflictError`, `NotFoundError`, `InternalError`) matching user-service's existing definitions
- Remove the duplicated types from `user-service/internal/models/errors.go`; keep re-export aliases for backward compat during transition
- Update all ~68 service-layer references in user-service to import from `common/errors` instead of `models`
- Update handler dispatcher imports to reference common/errors types

## Capabilities

### Modified Capabilities

- `common-errors`: ADD typed error structs (`ValidationError`, `ConflictError`, `NotFoundError`, `InternalError`) with constructor functions. These complement the existing sentinel errors (`ErrNotFound`, `ErrAlreadyExists`, etc.) by carrying structured data (field names, resource IDs) that handlers render into JSON response fields.

## Impact

- **Affected code**: `user-service/internal/models/errors.go` (remove types), `common/errors/typed.go` (new file), `user-service/internal/services/*.go` (~68 references updated), `user-service/internal/handlers/user_handler.go` (import update)
- **API behavior change**: None — same HTTP status codes, same JSON response shapes
- **Breaking change**: No. Re-export aliases in user-service/models/errors.go maintain backward compatibility for any external consumers
