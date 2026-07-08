## Context

objects-service has 4 handler files with inconsistent error handling:

| Handler | Method Name | Coverage | Pattern |
|---------|-------------|----------|---------|
| `object_handler.go` | `handleServiceError()` | 1 case (ErrOptimisticLock → 409) | `errors.Is()` if-then, then default 500 |
| `object_type_handler.go` | `handleServiceError()` | 0 cases (switch nil only) | Falls through to 500 for everything |
| `relationship_handler.go` | `handleError()` | ~9 cases | Full switch with errors.Is() — **good pattern** |
| `relationship_type_handler.go` | `handleError()` | ~6 cases + default 500 | Switch, but missing some sentinel checks |

The service and repository layers already define proper sentinel errors:

```go
// Repository (interfaces.go)
ErrNotFound, ErrAlreadyExists, ErrInvalidInput, ErrOptimisticLock, ...

// Services (object_service.go, relationship_type_service.go, etc.)
ErrObjectNotFound, ErrTypeKeyExists, ErrCircularRelationship, ... (~23 total)
```

These errors are wrapped by services with `fmt.Errorf("ctx: %w", err)` and the handlers **can** detect them via `errors.Is()`, but most handlers simply don't check. The relationship handler is the exception — it has a well-structured switch that covers all service-layer sentinels.

## Goals / Non-Goals

**Goals:**
- Every objects-service error returns the correct HTTP status code (404, 409, 422, 400) matching its semantic meaning
- All 4 handler files use a single shared dispatcher — no more duplicate logic or missing cases
- The dispatch covers all ~35 sentinel errors across repository and service layers

**Non-Goals:**
- No changes to error creation in services/repositories (they already wrap correctly with `%w`)
- No cross-service unification (auth-service, user-service remain untouched)
- No structural error types — continue using sentinel `var ErrX = fmt.Errorf(...)` pattern
- No changes to error messages or response format (already compliant per API Response Standards)

## Decisions

### Decision 1: Shared dispatcher in handlers package vs common/handlers/

**Choice:** Create a shared function inside the objects-service handlers package (`internal/handlers/error.go`).

**Rationale:**
- Cross-service unification is out of scope (user-service uses typed structs, auth-service uses inline calls — different dispatch mechanisms)
- A `common/handlers/` package would require standardizing error types across services first — unnecessary complexity for a single-service fix
- The relationship handler already has a good internal pattern; factoring it into the same package avoids import cycles and keeps scope tight

### Decision 2: Function signature — single dispatcher vs per-handler helpers

**Choice:** Single `HandleError(c *gin.Context, err error, operation string) int` function that returns both status code and writes JSON response. The caller logs and assigns the return value for logging.

```go
func HandleError(c *gin.Context, err error, requestID string) {
    // switch on errors.Is() cases → c.JSON(statusCode, gin.H{...})
}
```

**Rationale:**
- Eliminates per-handler method duplication (each handler currently has its own `handleServiceError` or `handleError`)
- Consistent logging format across all handlers via shared implementation
- Return int allows callers to log the resolved status code separately if needed

### Decision 3: Error case ordering — service-layer first, then repository-layer

**Choice:** Check service-layer sentinels first (they're more specific), fall back to repository-layer sentinels for generic cases.

```go
switch {
case errors.Is(err, services.ErrObjectNotFound):          // 404
case errors.Is(err, services.ErrCircularRelationship):     // 422
case errors.Is(err, repository.ErrOptimisticLock):         // 409
case errors.Is(err, repository.ErrNotFound):               // 404 (catch-all)
default:                                                   // 500
}
```

**Rationale:** `errors.Is()` unwraps through all layers automatically. Service-layer wrappers like `"object not found: %w"` wrap a repo sentinel, so checking the service sentinel first is more precise and catches domain-specific errors before generic ones.

### Decision 4: Response format — match existing per-handler patterns

**Choice:** Each case writes its own `c.JSON()` with `{error, type, meta}` following API Response Standards. No shared response struct needed since all handlers already use the same gin.H shape.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Missing a sentinel error in the dispatcher (same problem as now) | Exhaustive audit of `var Err` across services/rep during implementation; add regression test for each case |
| Changing 500→404 could break clients that relied on the (wrong) behavior | Dev mode only — no production users yet. Safe to fix immediately |
| Adding ~60 lines of switch cases adds cognitive load to a new file | The relationship_handler.go already has this pattern documented; it's proven, not theoretical |

## Migration Plan

1. Create `internal/handlers/error.go` with the shared dispatcher (~60 lines)
2. Update each handler's error sites: replace method call with shared function call
3. Run tests — all existing tests that check for 500 status codes will need updates to reflect correct codes (404, 422, etc.)
4. Deploy and verify MCP tool calls return proper error responses

No database migrations or config changes required. Zero-downtime deploy since this only affects response code mapping.

## Open Questions

1. Should the dispatcher also log structured fields like `resource_type` based on the operation name? (Low priority — can add later)
2. What about errors from validation middleware (ShouldBindJSON failures)? Those are already handled inline with 400 responses before reaching service layer — no change needed.
