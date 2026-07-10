## Context

Auth-service has three error handling patterns coexisting in one handler file:
1. HandleAuthError() — matches 2 sentinel errors (ErrUnauthorized, ErrNotFound) via errors.Is()
2. validationError()/errorResponse() — inline JSON responses for request format and middleware checks
3. Inline strings.Contains() detection — two places where handlers inspect raw error messages to detect DB constraint violations

The "cannot delete role/permission" business rules currently return raw strings that fall through HandleAuthError's default case → HTTP 500. This is incorrect: these are business rule violations, not server failures. The fix requires domain-specific sentinels (like objects-service uses) so the dispatcher can classify them correctly.

## Goals / Non-Goals

**Goals:**
- Define two domain-specific sentinel errors in auth-service service layer (ErrRoleInUse, ErrPermissionInUse)
- Wrap existing raw-string returns with these sentinels via %w
- Add matching cases to HandleAuthError dispatcher → HTTP 422
- Remove inline strings.Contains("duplicate key") detection from CreateRole/CreatePermission handlers
- Move DB duplicate detection to service layer (like user-service does it, but wrap with common.ErrAlreadyExists)

**Non-Goals:**
- Fixing the DB string-matching pattern itself — that's a known issue deferred to a future delta
- Adding more domain sentinels beyond these two business rules
- Changing error response format or JSON structure
- Modifying any other service (user-service, objects-service)

## Decisions

### Decision 1: Domain-specific sentinels in services package (not repository)

**Choice:** Define ErrRoleInUse and ErrPermissionInUse as `var` in `auth_service.go` alongside existing sentinel definitions.

**Rationale:** Objects-service defines its service-layer sentinels (`ErrRelationshipTypeInUse`, etc.) in the same file where they're used. Auth-service follows this pattern. These are business rule violations known only at the service layer (they require checking counts before deletion), not repository conditions like sql.ErrNoRows.

### Decision 2: HTTP 422 for "resource in use" violations

**Choice:** Return HTTP 422 Unprocessable Entity with error type "validation_error" for both domain sentinels.

**Rationale:** These are business rule violations — the request is well-formed but cannot be processed due to data constraints. Objects-service uses 422 for similar conditions (ErrCircularRelationship, ErrCardinalityViolation). HTTP 409 Conflict could also work but 422 better expresses "your input is valid but violates a business constraint."

### Decision 3: Common.ErrAlreadyExists for DB duplicate key detection

**Choice:** Move the two `strings.Contains("duplicate key")` blocks from handlers to service layer CreateRole/CreatePermission methods, wrapping with common/errors.ErrAlreadyExists.

**Rationale:** The string-matching itself is fragile (known issue), but moving it from handler to service layer eliminates the most problematic pattern (handler knowing about PostgreSQL internals). This change defers fixing the string-matching approach entirely — just relocations the detection point so errors flow through HandleAuthError consistently. A future delta can replace `strings.Contains()` with proper pgx error code checking or constraint-specific repository methods.

### Decision 4: Keep existing handler inline patterns for auth middleware checks

**Choice:** Do NOT refactor validationError()/errorResponse() calls used for request format validation and auth middleware checks (Authorization header, token validation).

**Rationale:** These are correct uses of inline responses — they handle pre-service-layer conditions (malformed JSON, missing headers) that don't go through the service layer. Refactoring them would add complexity without benefit. Only refactor error paths that involve actual service layer calls.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| DB duplicate key string-matching still fragile after this delta | Documented in known-issues.md as deferred — future delta can improve detection mechanism; current change only relocates to service layer |
| Handler tests may need updates for new error paths | Existing handler tests mock repo errors and check string matching — they'll need to be updated to expect wrapped sentinels instead of raw strings, or test the dispatcher behavior directly |
| HTTP status code change (500 → 422) could break clients expecting 500 | These endpoints never returned correct status codes before; 422 is the correct status for business rule violations. Any client that was handling "Internal server error" will now get a meaningful response |

## Migration Plan

1. Define ErrRoleInUse and ErrPermissionInUse sentinels in auth_service.go
2. Wrap DeleteRole/DeletePermission raw-string returns with these sentinels
3. Move DB duplicate detection from CreateRole/CreatePermission handlers to service layer (service layer wraps with common.ErrAlreadyExists)
4. Add 3 new cases to HandleAuthError dispatcher (ErrRoleInUse, ErrPermissionInUse → 422; ErrAlreadyExists → 409)
5. Remove inline strings.Contains("duplicate key") blocks from CreateRole/CreatePermission handlers
6. Run `go build` and fix any compilation errors
7. Update handler tests to match new error paths
8. Verify all auth-service tests pass

## Open Questions

None — all decisions resolved by following objects-service pattern for domain sentinels and user-service pattern for relocating DB detection from handler to service layer.
