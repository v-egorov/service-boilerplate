## Context

Auth-service is the last service without structured error handling. Its service layer returns ~50 errors as raw strings with no semantic type, and its handler hardcodes HTTP status codes per method rather than reading error semantics from the error chain itself. This means:

1. "invalid credentials" (401) and "failed to generate token" (also 401 in Login, but would be 500 in Register) get identical handling despite different root causes
2. Infrastructure failures ("failed to get user roles") always return 500 even when the underlying cause is a missing resource (should be 404)
3. Handler code has ~76 inline `errorResponse()`/`validationError()` calls scattered across methods, each duplicating the same status-code-lookup pattern

The service layer uses `fmt.Errorf("...: %w", err)` for wrapping but never wraps with a sentinel — just string context around another wrapped error. There's no `errors.Is(err, common/errors.ErrNotFound)` anywhere in auth-service because no sentinel was ever introduced at any layer.

### Current Error Flow (broken)

```
Service Layer                    Handler Layer
─────────────                    ───────────────
Login() → err ← "invalid credentials"  errorResponse(c, 401, ...)  ✓ OK by coincidence
Register() → err ← "failed to..."      errorResponse(c, 500, ...)  ✗ Wrong — should distinguish infra vs auth failures
CreateRole() → err ← "duplicate key"   errorResponse(c, 500, ...)  ✗ Same method, different cause, same status
```

## Goals / Non-Goals

**Goals:**
- Wrap ~50 service-layer errors with `common/errors.*` sentinels (ErrUnauthorized for auth failures, ErrNotFound for lookups, ErrInvalidInput for validation)
- Extract shared `HandleAuthError(err, requestID)` dispatcher method on AuthHandler that reads sentinels and maps to HTTP status codes
- Replace inline `errorResponse()`/`validationError()` calls in handler with the dispatcher
- Re-export local typed error types as aliases (same pattern as user-service)

**Non-Goals:**
- **Auth-specific typed errors**: PermissionParseError, ScopedVariantConflictError remain in auth-service/models/errors.go — they're domain concepts handled via existing `errors.As()` type-switches
- **Permission handler cleanup**: permission_handler.go has its own patterns (some already use errors.As) — scope limited to auth_handler.go
- **Auth middleware error handling**: JWT validation errors are hardcoded 401 in the middleware itself — acceptable for now since they're always auth failures
- **Health handler**: Already minimal, no changes needed

## Decisions

### Decision 1: ErrUnauthorized for ALL authentication failures (not separate sentinels)
**Choice:** Map invalid credentials, expired tokens, revoked tokens, and "invalid token" all to `ErrUnauthorized`.

**Rationale:** These are fundamentally the same class of error — the client failed to authenticate or maintain a valid session. There's no business value in distinguishing them at the HTTP level (all → 401). Using separate sentinels would add complexity without benefit and create more code paths to test.

### Decision 2: ErrNotFound detection via sql.ErrNoRows in service layer
**Choice:** Where service methods query databases and detect `sql.ErrNoRows`, wrap with `ErrNotFound`. For string-wrapped errors that already contain "not found" text (like "token not found"), also use `ErrUnauthorized` since the semantic meaning is "auth failure" rather than "resource missing."

**Rationale:** This follows objects-service's pattern. The key insight: sql.ErrNoRows → ErrNotFound is a repository-level detection, but in auth-service the service layer does direct queries (no separate repo for most operations). So we detect it at the service level where the query happens.

### Decision 3: Dispatcher as method on handler struct (not standalone)
**Choice:** `func (h *AuthHandler) HandleAuthError(c *gin.Context, err error)` — a method that writes JSON response and returns.

**Rationale:** This matches user-service's pattern (`handleServiceError` is a method). It keeps the dispatcher close to the handler receiver, allows access to logger for consistent logging, and avoids changing every call site from `h.errorResponse()` to standalone function calls (only need to change the body of each call, not the syntax).

### Decision 4: Keep errorResponse()/validationError() as thin wrappers
**Choice:** After extracting HandleAuthError, remove errorResponse() entirely but keep validationError() for request-parsing errors that don't come from service layer.

**Rationale:** Request parsing failures (invalid JSON, missing fields) are client input errors — they're not domain/service errors and should always be 400 with "validation_error" type. These bypass the service layer entirely so HandleAuthError would never see them. Keeping validationError() avoids over-engineering.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Changing error wrapping changes log messages (different string format) | Logs use err.Error() which preserves the full wrapped chain — logs remain informative |
| Some hardcoded 500s now return different status codes if wrapped with ErrNotFound | This is a fix, not a regression. Verify via tests that existing behaviors are preserved for non-sentinel errors |
| Handler method signatures unchanged but internal logic changes significantly | No API contract change — response format identical, only classification accuracy improves |

## Migration Plan

1. **Phase 1**: Wrap service-layer errors with sentinels (auth_service.go + key_rotation_manager.go)
2. **Phase 2**: Extract HandleAuthError dispatcher on AuthHandler struct
3. **Phase 3**: Replace ~76 inline errorResponse()/validationError() calls in auth_handler.go
4. **Phase 4**: Re-export typed errors from models/errors.go (same alias pattern as user-service)
5. **Verify**: Build + tests pass, no behavior regression

## Open Questions

None — all decisions resolved by design choices above. The only remaining detail is which specific error strings map to which sentinel, determined during implementation by reading each service-layer return site.
