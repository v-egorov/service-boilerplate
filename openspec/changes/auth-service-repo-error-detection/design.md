## Context

auth-service's repository (`services/auth-service/internal/repository/auth_repository.go`) has 30+ methods but NO `interfaces.go` file and NO sql.ErrNoRows detection at the repository boundary. This was established during Delta 1 (`auth-service-error-consolidation`) when we wrapped ~50 service-layer errors with sentinels, but the repo layer itself remained untouched.

Currently, only 2 service-layer methods manually check `sql.ErrNoRows` (GetPermission and GetRole in `auth_service.go:764-778`). All other single-row GET methods propagate raw pgx errors through `%w` wrapping directly to callers. The service layer then either:
1. Wraps with ErrUnauthorized/ErrNotFound (the 2 manual checks) → handler dispatcher matches ✅
2. Propagates raw err → defaults to HTTP 500 ❌

This is inconsistent with objects-service, which does sql.ErrNoRows detection in every Get* method at the repository layer (Rule 7). It's also a violation of Rule 7 itself — this delta brings auth-service into compliance.

## Goals / Non-Goals

**Goals:**
- Create `auth-service/internal/repository/interfaces.go` with sentinel aliases and RepositoryInterface definition
- Add sql.ErrNoRows → ErrNotFound detection to all single-row GET methods in auth_repository.go (~5 methods)
- Remove manual sql.ErrNoRows checks from service layer (they're now handled at repo level)
- Ensure compilation, tests pass, zero API behavior change

**Non-Goals:**
- Adding new sentinels beyond ErrNotFound — Delta 1 already established ErrUnauthorized for auth failures
- Wrapping infrastructure errors with sentinels — those correctly default to 500
- Refactoring repository patterns (no interface extraction for other methods)
- Modifying any methods that return slices or perform non-entity queries

## Decisions

### Decision 1: Create interfaces.go as a new file (not merge into auth_repository.go)

**Choice:** New `repository/interfaces.go` file containing sentinel aliases and the RepositoryInterface.

**Rationale:** Objects-service uses this pattern (`interfaces.go` + concrete implementation in separate file). It keeps the repo package organized, allows easy reference for Rule 7 compliance documentation, and follows the established boilerplate convention.

### Decision 2: ErrNotFound as repository alias to common/errors.ErrNotFound

**Choice:** `ErrNotFound = errors.ErrNotFound` (type alias to the shared sentinel).

**Rationale:** This is identical to objects-service's approach (`repository/interfaces.go:133`). It ensures `errors.Is(err, repository.ErrNotFound)` works correctly in both service layer and handler dispatcher. The alias preserves pointer identity so existing direct comparisons (`err == ErrNotFound`) still work if any exist (though Delta 1 removed those).

### Decision 3: Service layer removes manual sql.ErrNoRows checks

**Choice:** Delete the ~2 manual `if errors.Is(err, sql.ErrNoRows)` blocks from auth_service.go now that repo methods return ErrNotFound directly.

**Rationale:** The service layer was doing this check because the repo didn't handle it. With repo-level detection, these checks are redundant and create double-wrapping risk (anti-pattern per Rule 5). Service methods should simply propagate whatever the repo returns.

### Decision 4: Only modify single-row GET methods — not collection queries or mutations

**Choice:** Target exactly 5 methods: `GetAuthTokenByHash`, `GetUserSession`, `GetRoleByName`, `GetRole`, `GetPermission`. Skip `ListRoles`, `ListPermissions`, `GetUserRoles`, `GetUserPermissions` (return slices, empty = no error). Skip all mutation methods (Create*, Update*, Delete*).

**Rationale:** Rule 7 explicitly scopes to single-row Get*() methods. Collection queries return empty non-nil slices per Rule 1 — no ErrNoRows possible. Mutation errors are infrastructure failures that correctly default to 500.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Service layer code changes might break tests | Delta 1 already updated handler tests to use wrapped sentinels; service-layer tests mock repo calls and will need minor updates for new ErrNotFound return paths |
| Manual sql.ErrNoRows checks in auth_service.go become dead code after this change | Explicitly remove them as part of the delta — no stale code left behind |
| Handler dispatcher (HandleAuthError) already catches ErrNotFound → 404, so behavior unchanged | Verify via existing handler tests that GetRole/GetPermission not-found paths still return 401 or whatever the service layer wraps them with |

## Migration Plan

1. Create `repository/interfaces.go` with sentinel alias and interface
2. Modify auth_repository.go: add sql.ErrNoRows → ErrNotFound to 5 GET methods
3. Remove manual sql.ErrNoRows checks from auth_service.go (service layer)
4. Run `go build ./services/auth-service/...` — verify compilation
5. Run `go test ./services/auth-service/...` — fix any test failures
6. Verify no behavioral regression: compiled binary returns same HTTP status codes

## Open Questions

None — all decisions resolved by existing patterns in objects-service and Rule 7 of go-coding-conventions.md.
