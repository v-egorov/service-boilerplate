## Why

Objects-service still defines `ErrNotFound` and `ErrInvalidInput` locally in the repository package instead of using the shared `common/errors` sentinels created by the `common-errors-infrastructure` delta. This means objects-service doesn't benefit from the cross-service error consistency that common/errors provides, and any future handler dispatcher that imports only `common/errors` won't match these errors.

## What Changes

- **Remove** `repository.ErrAlreadyExists` — defined but never returned by any method; remove definition + handler case + test references
- **Migrate** `ErrNotFound` (8 repository returns, ~60 service references) → `common/errors.ErrNotFound`
- **Migrate** `ErrInvalidInput` (~25 service references) → `common/errors.ErrInvalidInput`
- Handler dispatcher updates to use `common/errors.*` for migrated sentinels
- Service layer test data updated to reference new types

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `objects-service`: Handler error dispatcher requirement — sentinel mappings now reference `common/errors.ErrNotFound` and `common/errors.ErrInvalidInput` instead of repository-local variants; removes `ErrAlreadyExists` from the mapping table

## Impact

**Files changed:** ~12 files across objects-service:
- `internal/repository/interfaces.go` — remove ErrAlreadyExists, update ErrNotFound/ErrInvalidInput definitions
- `internal/handlers/error.go` — update sentinel references in dispatcher cases
- `internal/handlers/error_test.go` — update test data to use common/errors types, remove ErrAlreadyExists tests
- `internal/services/object_service.go` — ~25 refs updated
- `internal/services/object_type_service.go` — ~10 refs updated  
- `internal/services/relationship_service.go` — ~8 refs updated
- `internal/services/relationship_type_service.go` — ~6 refs updated
- Service test files — update references

**No API behavior change:** HTTP status codes and error types remain identical. Only the sentinel source changes from `repository.*` to `common/errors.*`.
