## Context

Three services use three different error handling patterns:

```
objects-service  → shared sentinels + HandleError() dispatcher ✅
user-service     → typed struct errors (ValidationError, NotFoundError...) + per-handler type-switch ⚠️
auth-service     → hardcoded errorResponse()/validationError() scattered ❌
```

The gateway is a verbatim reverse proxy — it forwards HTTP status codes and JSON bodies without parsing or translation. The shared error package serves internal consistency: each service imports the same sentinels so `errors.Is()` works uniformly across layers within that service.

## Goals / Non-Goals

**Goals:**
- Create `common/errors/` with 5 sentinel variables that services import as base errors
- Provide non-breaking foundation for subsequent service migration deltas
- Support both bare-sentinel pattern (objects-service style) and wrapped-typed-struct pattern (user-service style) via `errors.Is()`

**Non-Goals:**
- Migrating any existing service in this delta — each gets its own separate delta
- Creating a gateway error dispatcher — the gateway forwards HTTP status codes verbatim
- Replacing user-service's typed struct errors — they stay local and can wrap common sentinels via `%w`
- Adding new error types beyond the 5 infra-level sentinels

## Decisions

### Decision 1: Sentinel vars (not structs) in common/errors
**Choice:** `var ErrNotFound = fmt.Errorf("resource not found")` — simple, untyped sentinel variables.

**Rationale:** Objects-service already uses this pattern and it works well for infra-level errors that don't carry structured metadata. User-service can keep its rich typed structs locally and wrap the common sentinels if needed: `fmt.Errorf("%s: %w", err, common.ErrNotFound)`. The handler dispatcher matches via `errors.Is()` which traverses `%w` chains regardless of intermediate types.

**Alternatives considered:**
- Struct-based errors in common/errors — too rigid for auth-service domain (what's the "resource" for a JWT key rotation?). Forces awkward struct fields on every service.
- String constants — no type safety, can't use `errors.Is()`.

### Decision 2: Five infra-level sentinels only
**Choice:** ErrNotFound, ErrAlreadyExists, ErrInvalidInput (infra) + ErrUnauthorized, ErrForbidden (auth).

**Rationale:** These five cover every cross-boundary error the gateway needs to recognize. Objects-service uses all three infra types. Auth-service adds the two auth-specific ones. No other services need shared errors at this granularity — business-domain errors stay local.

### Decision 3: Non-breaking by design
**Choice:** Delta 0 creates only new code (common/errors/). Zero changes to existing services, zero spec modifications.

**Rationale:** Keeps risk minimal. If the package has issues, nothing downstream breaks. Service migration deltas can adopt it incrementally without a monolithic "all-or-nothing" change.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Services ignore the shared package and create their own sentinels | Document in go-coding-conventions Rule 7; add lint rule or code review gate later |
| Auth-service ErrUnauthorized/ErrForbidden conflict with HTTP-level auth middleware | These are service-layer errors (returned from handlers), not middleware — different concern. The gateway passes HTTP status codes verbatim so no conflict exists. |

## Migration Plan

1. Create `common/errors/` package (this delta)
2. Validate: `make build-all && make test-all` pass with zero changes to existing services
3. Subsequent deltas migrate each service independently: objects-service → user-service → auth-service
4. Each migration delta updates go-coding-conventions.md and adds its own spec coverage

## Open Questions

None — scope is intentionally minimal (5 vars, new package only).
