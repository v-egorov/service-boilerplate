## Context

The objects-service repository layer has an inconsistency in how it handles "not found" conditions:

- `object_repository.GetByPublicID()` — checks `errors.Is(err, sql.ErrNoRows)` → returns `repository.ErrNotFound` ✅
- `relationship_repository.GetByObjectID()`, `GetByPublicID()` — same pattern ✅
- `relationship_type_repository.GetByID()`, `GetByTypeKey()`, `GetByReverseTypeKey()` — same pattern ✅
- **3 methods do NOT check:** `object_repository.GetByID()`, `object_type_repository.GetByID()`, `object_type_repository.GetByName()` ❌

When these 3 methods encounter a missing row, pgx returns `sql.ErrNoRows`. The method wraps it with `%w` and passes it up. The service layer checks `err == repository.ErrNotFound` (direct pointer comparison), which **fails** on the wrapped error. The handler dispatcher then catches it via its `sql.ErrNoRows` fallback case — working, but as a workaround.

```
Current broken flow:
  Repository.GetByID → fmt.Errorf("failed to get object: %w", sql.ErrNoRows)
    ↓ (service layer checks err == ErrNotFound → FAILS)
  Service.GetByID → fmt.Errorf("failed to get object: %w", wrapped_err)
    ↓ (handler dispatcher catches via sql.ErrNoRows fallback case)
  Handler.HandleError → HTTP 404 "not_found" ✅ works but masks the repo gap
```

## Goals / Non-Goals

**Goals:**
- Fix all 3 broken repository methods to return `repository.ErrNotFound` sentinel consistently
- Remove the `sql.ErrNoRows` fallback case from handler dispatcher (dead code cleanup)
- Verify no behavior change — HTTP 404 responses already work correctly

**Non-Goals:**
- Service layer refactoring (direct comparisons vs `errors.Is()` — currently mixed but both work after this fix since sentinels are returned unwrapped)
- Cross-service fixes (user-service, auth-service have their own patterns)
- Adding new error types or sentinel definitions

## Decisions

### Decision 1: Fix at repository layer only
**Choice:** Add `errors.Is(err, sql.ErrNoRows)` check in the 3 broken repo methods. Return `repository.ErrNotFound` directly (unwrapped).

**Rationale:** The service layer's direct comparison (`err == ErrNotFound`) works when the sentinel is returned unwrapped. Wrapping it again would require changing all service-layer comparisons to use `errors.Is()`, which is scope creep. Keeping sentinels unwrapped at repo→service boundary matches what relationship_repository already does correctly.

### Decision 2: Remove sql.ErrNoRows fallback from dispatcher
**Choice:** Delete the `case errors.Is(err, sql.ErrNoRows)` branch from `error.go`.

**Rationale:** After fixing the 3 repo methods, this case is unreachable dead code. Removing it simplifies the dispatcher and makes the error handling more explicit — every mapped error now has a named sentinel source.

### Decision 3: No test changes for repository layer
**Choice:** Don't add new repository tests. The existing `object_repository_test.go` doesn't have integration-level "GetByID returns ErrNotFound" tests, but adding them would expand scope beyond the narrow fix.

**Rationale:** The handler dispatcher test (`error_test.go`) already verifies the 404 behavior end-to-end via service layer mocks. Once we verify builds pass and existing tests compile, that's sufficient for this delta.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Service layer direct comparisons might miss wrapped errors from other callers | Only applies to unwrapped sentinels returned directly from repo — which is the intended contract. If service layers wrap with `%w` before returning to handlers, `errors.Is()` is used there. |
| Removing sql.ErrNoRows fallback could unmask a bug if another caller passes raw pgx error upstream | After this fix, all single-row Get methods in objects-service repo return ErrNotFound for missing rows. The only risk would be from non-objects-service code paths — unlikely given the narrow scope. |

## Migration Plan

1. Apply changes (3 repo files + 2 handler files)
2. Run `make build-objects-service` and `make test-objects-service` to verify no regressions
3. No database migration needed, no runtime restart required beyond container rebuild
4. Rollback: revert the 5 file changes — no data impact
