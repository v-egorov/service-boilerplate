## Why

Three repository `Get*()` methods (`object_repository.GetByID`, `object_type_repository.GetByID`, `object_type_repository.GetByName`) pass raw `sql.ErrNoRows` through `%w` wrapping instead of converting it to the `repository.ErrNotFound` sentinel. This forces the handler dispatcher (added in `fix-handler-error-dispatch`) to include a `sql.ErrNoRows` fallback case — a workaround bandage that masks an inconsistency at the repository layer.

## What Changes

- Add `errors.Is(err, sql.ErrNoRows)` check + return `repository.ErrNotFound` sentinel in three repository methods
- Remove the `sql.ErrNoRows` fallback case from the handler dispatcher (`error.go`) since it becomes dead code
- Update handler tests to verify no workaround is needed
- No service layer changes required — existing direct comparisons (`err == repository.ErrNotFound`) work correctly once sentinels are returned unwrapped

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `objects-service`: Handler dispatcher requirement updated — removes the `sql.ErrNoRows` fallback mapping, simplifying the dispatcher to rely solely on repository/service sentinel errors

## Impact

**Files changed:** ~4 files across objects-service:
- `services/objects-service/internal/repository/object_repository.go` (2 methods)
- `services/objects-service/internal/repository/object_type_repository.go` (1 method)
- `services/objects-service/internal/handlers/error.go` (remove 1 case)
- `services/objects-service/internal/handlers/error_test.go` (update test coverage)

**No API behavior change:** HTTP 404 `not_found` responses already work correctly via the workaround fallback. This delta removes the workaround at its source.
