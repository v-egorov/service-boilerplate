## Context

Objects-service was the first to implement a shared handler dispatcher (`HandleError()` in `internal/handlers/error.go`), but it still uses repository-local sentinel definitions for `ErrNotFound` and `ErrInvalidInput`:

```go
// current (repository/interfaces.go)
var (
    ErrNotFound        = fmt.Errorf("resource not found")
    ErrAlreadyExists   = fmt.Errorf("resource already exists")  // never returned!
    ErrInvalidInput    = fmt.Errorf("invalid input")
    ErrOptimisticLock  = fmt.Errorf("optimistic lock failed")   # objects-service specific
)
```

The `common-errors-infrastructure` delta created `common/errors/` with matching sentinels (`ErrNotFound`, `ErrAlreadyExists`, `ErrInvalidInput`), but objects-service never adopted them. This means:
- The handler dispatcher references `repository.ErrNotFound` instead of the shared type
- Future services importing only `common/errors` won't match these errors via `errors.Is()`
- ErrAlreadyExists is defined and referenced in tests/dispatcher but **never actually returned** by any repository method

## Goals / Non-Goals

**Goals:**
- Replace `ErrNotFound` references (8 repo returns + ~60 service refs) with `common/errors.ErrNotFound`
- Replace `ErrInvalidInput` references (~25 service refs) with `common/errors.ErrInvalidInput`
- Remove `ErrAlreadyExists` entirely — definition, handler case, test data
- Update handler dispatcher to reference `common/errors.*` for migrated sentinels
- Build passes and all tests pass (zero behavior change)

**Non-Goals:**
- Migrate objects-service-specific sentinels (`ErrOptimisticLock`, `ErrVersionConflict`, `ErrSoftDeleted`, etc.) — these stay in repository package as they're not shared infra types
- Update service-layer direct comparisons (`err == ErrNotFound`) to use `errors.Is()` — this is a future delta
- Touch user-service or auth-service (those are separate deltas)

## Decisions

### Decision 1: Remove ErrAlreadyExists entirely instead of migrating it
**Choice:** Delete the definition from repository/interfaces.go, remove the dispatcher case, and update test references.

**Rationale:** Nothing in objects-service ever returns `ErrAlreadyExists`. The only references are:
- Definition in interfaces.go (dead code)
- One handler dispatcher case (dead code path — no error can reach it)
- Two test entries in error_test.go

Moving dead code is pointless. If a future delta needs this sentinel, it can be re-added to common/errors or objects-service's repository layer.

### Decision 2: Keep ErrOptimisticLock/ErrVersionConflict local
**Choice:** These two remain defined in `repository/interfaces.go` and referenced by the handler dispatcher as-is.

**Rationale:** They're objects-service-specific concepts (optimistic locking for concurrent updates). No other service has an equivalent error type, so there's no benefit to moving them to common/errors. The handler dispatcher already matches them correctly via `errors.Is(err, repository.ErrOptimisticLock)`.

### Decision 3: Update import paths incrementally
**Choice:** Each file adds `"github.com/v-egorov/service-boilerplate/common/errors"` and replaces references in-place. No bulk refactor.

**Rationale:** Keeps each commit small and reviewable. The Go compiler will catch any missed references during `go build`.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| A missed reference causes compile error (common/errors not imported) | `go build ./...` catches all missing imports; test run verifies behavior |
| ErrAlreadyExists removal breaks something we don't know about | It's dead code — nothing returns it. Tests that reference it will fail to compile until updated, which is the correct signal |

## Migration Plan

1. Update repository/interfaces.go: remove ErrAlreadyExists var, update ErrNotFound/ErrInvalidInput to use common/errors types
2. Update handler/error.go: change sentinel references in dispatcher cases
3. Update handler/error_test.go: update test data, remove ErrAlreadyExists tests
4. Update service files (object_service.go, object_type_service.go, relationship_service.go, relationship_type_service.go): replace all `repository.ErrNotFound` and `repository.ErrInvalidInput` refs with `errors.Err...`
5. Build + test to verify no regressions

## Open Questions

None — scope is well-defined by the existing codebase patterns.
