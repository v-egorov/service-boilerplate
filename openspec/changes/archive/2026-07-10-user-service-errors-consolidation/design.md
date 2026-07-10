## Context

User-service (`services/user-service/internal/models/errors.go`) and auth-service (`services/auth-service/internal/models/errors.go`) both define identical typed error structs: `ValidationError`, `ConflictError`, `NotFoundError`, `InternalError`. Each has the same struct definitions, `Error()` methods, and constructor helper functions — duplicated across two services. The objects-service uses a different pattern (sentinel aliases via common/errors), so this delta focuses on user-service only.

### Current State

```
common/errors/
  infra.go: ErrNotFound, ErrAlreadyExists, ErrInvalidInput (sentinels)
  auth.go:   ErrUnauthorized, ErrForbidden (sentinels)

user-service/internal/models/errors.go:
  ValidationError {Field, Message}         ← DUPLICATED from auth-service
  ConflictError     {Resource, Field, Value}
  NotFoundError   {Resource, Field, Value}
  InternalError   {Operation, Err} + Unwrap()
  Constructor functions for each

auth-service/internal/models/errors.go:
  ValidationError {Field, Message}         ← DUPLICATED from user-service (same code)
  ConflictError     {Resource, Field, Value}
  NotFoundError   {Resource, Field, Value}
  InternalError   {Operation, Err} + Unwrap()
  Constructor functions for each (EXTRA: PermissionParseError, ScopedVariantConflictError — auth-specific)

auth-service/internal/services/*.go:
  ~2 uses of models.NewNotFoundError (shared infra type)
```

## Goals / Non-Goals

**Goals:**
- Create `common/errors/typed.go` with the four shared typed error structs + constructors
- Remove duplicated types from `user-service/internal/models/errors.go`; keep re-export aliases for backward compat
- Update all ~68 service-layer references in user-service to import from `common/errors` instead of `models`
- Update handler dispatcher imports

**Non-Goals:**
- **Auth-service changes**: Auth-service has the same duplication but also defines auth-specific types (`ScopedVariantConflictError`, `PermissionParseError`). That's Delta 2.
- **Shared dispatcher extraction**: User-service keeps its per-handler `handleServiceError()` method — no shared dispatcher yet.
- **Objects-service changes**: Objects-service uses sentinel aliases, not typed structs. No changes needed.

## Decisions

### Decision 1: Keep user-service re-export aliases for backward compat
**Choice:** After removing types from `user-service/internal/models/errors.go`, add `type ValidationError = errors.ValidationError` etc. as aliases.

**Rationale:** The handler file (`user_handler.go`) already imports `"github.com/v-egorov/service-boilerplate/services/user-service/internal/models"` and uses `models.ValidationError`. Re-export aliases mean the type-switch in `handleServiceError()` works without changes — same pointer, same type identity. If we removed them entirely, every reference in user-handler would need updating too.

### Decision 2: Type structs, not sentinels, for user-service domain errors
**Choice:** User-service continues using typed structs (`ValidationError`, etc.) rather than switching to sentinel-wrapped approach like objects-service.

**Rationale:** User-service's typed errors carry structured data (field names, resource identifiers) that handlers render into JSON response fields. A plain `common/errors.ErrNotFound` doesn't tell the handler which field or resource was not found — the typed struct does. This is a domain concern where rich metadata matters. Objects-service uses sentinels because its error dispatching happens at a different granularity (handler matches wrapped sentinels via errors.Is()). User-service wants per-field detail in responses.

### Decision 3: One file, one package for common/errors
**Choice:** New typed structs go into `common/errors/typed.go`, alongside existing `infra.go` and `auth.go`. All three files live in the same `common/errors` package.

**Rationale:** Keeping everything in one package means services import a single path (`"github.com/v-egorov/service-boilerplate/common/errors"`). Splitting into sub-packages would add complexity for minimal benefit — these types are always used together conceptually (infra sentinels + domain typed errors).

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Re-export aliases create confusion about where types "live" | Clear comments in models/errors.go pointing to common/errors; document in AGENTS.md |
| Auth-service still has duplicate definitions (same types) | Delta 2 will address this after user-service is stable |
| Handler type-switch depends on exact struct match, not errors.Is() | This is by design — typed structs carry metadata that errors.Is() can't express. The handler needs the actual struct value to extract field/resource/value for JSON rendering. |

## Migration Plan

1. Create `common/errors/typed.go` with all four typed error structs and constructors (copied verbatim from user-service definitions)
2. Update `user-service/internal/models/errors.go`: remove type definitions, replace with re-export aliases + clear comment pointing to common/errors
3. Bulk-replace service-layer imports: `models.NewValidationError` → `errors.NewValidationError`, etc. (~68 references across 4 service files)
4. Update handler dispatcher imports in user_handler.go
5. Build + test to verify no regressions

## Open Questions

None — all four types are well-defined and identical between user-service and auth-service (minus auth-specific extras). No behavioral change expected.
