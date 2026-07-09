# Go Coding Conventions

Project-wide rules for writing idiomatic, bug-free Go code in the service boilerplate.

---

## Rule 1: Initialize Collection Slices as Empty, Never Nil

**Scope:** All repository methods and any function returning `[]*T` over JSON APIs.

```go
// ❌ WRONG — nil zero-value serializes to JSON null
var items []*models.Object
for rows.Next() { ... }
return items  // → "data": null when empty

// ✅ CORRECT — non-nil empty slice serializes to JSON []
items := make([]*models.Object, 0)
for rows.Next() { ... }
return items  // → "data": [] when empty
```

**Why:** Go's nil zero-value for slices is `nil`, which gin serializes as `null`. Consumers (API clients, MCP tools, frontend apps) expect arrays from list endpoints. `make([]*T, 0)` guarantees consistent JSON serialization regardless of query results.

**Applies to:** Every repository method returning a collection (`[]*Model`). Includes: List, Search, FindBy*, GetChildren/Descendants/Ancestors/Path, tree traversal methods.

**Known exception (deferred):** `user-service/internal/repository/user_repository.go:133` — same pattern, will be fixed in a follow-up delta.

---

## Rule 2: Pagination Parameters — limit/offset Canonical, Default=50

**Scope:** All list endpoints returning paginated collections.

Always use **`limit` and `offset`** query parameters. Never use `page/page_size`. Default to `limit=50`, `offset=0` when client omits pagination params.

```go
// Handler extracts params with defaults
filter.Limit = 50
if l := c.Query("limit"); l != "" {
    fmt.Sscanf(l, "%d", &filter.Limit)
}
filter.Offset = 0
if o := c.Query("offset"); o != "" {
    fmt.Sscanf(o, "%d", &filter.Offset)
}

---

## Rule 3: Handler Error Mapping Must Dispatch on Error Types

**Scope:** All `handleServiceError()` / error-response functions in handlers.

```go
// ❌ WRONG — everything becomes 500
func handleServiceError(c *gin.Context, err error) {
    c.JSON(http.StatusInternalServerError, gin.H{
        "error": "Internal server error",
        "type":  "internal_error",
    })
}

// ✅ CORRECT — use shared HandleError dispatcher (internal/handlers/error.go)
func handleServiceError(c *gin.Context, err error) {
    handlers.HandleError(c, err, requestID)  // centralized mapping table
}
```
**Note:** The shared dispatcher lives at `internal/handlers/error.go`. Every handler calls it — never write per-handler error dispatch logic.

**Why:** The API response standards require HTTP status codes to match error types. A 500 with `"type":"internal_error"` for a missing resource (404) or invalid input (422) misleads clients and breaks automated error handling.

---

## Rule 4: Use Schema-Qualified Table Names in Queries

**Scope:** All SQL queries across all services.

```sql
-- ✅ CORRECT
SELECT * FROM objects_service.objects WHERE id = $1;
SELECT * FROM auth_service.permissions WHERE name = 'read';

-- ❌ WRONG — ambiguous in multi-schema DB
SELECT * FROM permissions WHERE name = 'read';
```

The development database uses PostgreSQL schemas (`auth_service`, `objects_service`, `user_service`). Unqualified names may work if the schema is in `search_path` today but will break with schema-aware clients or future migrations.

---

## Rule 5: Error Wrapping — One Level Per Call Stack Frame

**Scope:** All error handling across services.

```go
// ✅ CORRECT — one level of wrapping per layer
func (r *repo) List(ctx context.Context, filter *Filter) ([]*T, error) {
    rows, err := r.db.Query(...)
    if err != nil {
        return nil, fmt.Errorf("failed to query objects: %w", err)  // repo layer
    }
}

func (s *service) List(ctx context.Context, filter *Filter) ([]*T, int64, error) {
    items, total, err := s.repo.List(ctx, filter)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to list objects: %w", err)  // service layer (one more level)
    }
}
```

**Never:** `fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", err))` — double-wrapping in one call chain. Each layer wraps once; the outermost handler unwraps to match error types.

---

## Rule 6: Context Propagation for Identity and Tracing

**Scope:** All inter-service communication and internal request handling.

- **Identity forwarding:** Extract user identity from inbound headers → store in context via helper (`WithIdentity(ctx, hdr)`) → read from context when making outbound calls (`IdentityFromContext(ctx)`).
- **Request ID:** Always propagate `X-Request-ID` from the original request through all layers (handler → service → repository → HTTP client).
- **Tracing:** Use OpenTelemetry context propagation for distributed tracing across services.

---

## Rule 7: Repository Get*() Methods Return ErrNotFound for Missing Rows

**Scope:** All repository `Get*()` methods that retrieve a single row.

```go
func (r *objectRepository) GetByID(ctx context.Context, id int64) (*models.Object, error) {
    err := r.db.QueryRow(ctx, query, id).Scan(&obj.ID, ...)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {  // ← CHECK FIRST
            return nil, ErrNotFound           // ← RETURN UNWRAPPED SENTINEL
        }
        return nil, fmt.Errorf("failed to get object: %w", err)
    }
}
```

**Why:** This is the single most common source of error handling bugs in this codebase. When a repository method passes raw `sql.ErrNoRows` through `%w` wrapping instead of converting it to the `repository.ErrNotFound` sentinel, two cascading problems occur:

1. **Service layer direct comparison fails:** Service methods check `err == ErrNotFound` (pointer equality). A wrapped pgx error never matches, so the service layer falls through to its own generic wrap — hiding the "not found" intent.
2. **Handler dispatcher workaround required:** The handler must include a `sql.ErrNoRows` fallback case as a bandage to catch these unwrapped errors. This masks the real issue at the repository boundary.

**Contract between layers:**
- Repository returns `repository.ErrNotFound` directly (unwrapped) for missing rows → service layer's direct comparison works ✅
- Service wraps with its own sentinel via `fmt.Errorf("...: %w", err)` when needed → handler uses `errors.Is()` to match ✅

**Anti-patterns:**
```go
// ❌ WRONG — raw ErrNoRows wrapped through, service layer direct comparison fails
if err != nil {
    return nil, fmt.Errorf("failed to get object: %w", err)
}

// ❌ WRONG — wrapping ErrNotFound twice creates double-wrapped error
return nil, fmt.Errorf("service failed: %w", fmt.Errorf("repo failed: %w", repository.ErrNotFound))

// ✅ CORRECT — check first, return sentinel unwrapped
if errors.Is(err, sql.ErrNoRows) {
    return nil, ErrNotFound
}
```

**Applies to:** All single-row `Get*()` methods across all repositories (object_repository, object_type_repository, relationship_repository, relationship_type_repository, etc.). Collection-returning methods (List, Search, tree traversal) do NOT apply — they return empty slices instead.

---

## Rule 8: Test Coverage Expectations

**Scope:** All Go packages in every service. See [`docs/testify-overview.md`](testify-overview.md) for full framework guidance.

### Framework
All tests use **[testify](https://github.com/stretchr/testify)**:
```go
import "github.com/stretchr/testify/assert"
// or assert + require via:
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```
- `assert` — non-fatal checks, continues test on failure
- `require` — fatal checks, stops test immediately (use for critical setup)

### By Layer

| Layer | Pattern | Mocking Approach |
|-------|---------|------------------|
| Repository | `MockDBPool` implementing `DBInterface` | Manual mocks — full control over pgx.Rows iteration |
| Service | Interface-based mock of repository layer | Test both happy path and error wrapping (`fmt.Errorf("...: %w", err)`) |
| Handler | Smoke-test through real middleware chain (no writer mocks) | gin.Context + httptest/recorder — `gin.ResponseWriter` interface is not mockable |

### Empty Result Guarantee
Every collection-returning method must have at least one test asserting non-nil empty results:
```go
assert.NotNil(t, result)
assert.Len(t, result, 0)
```
This verifies Rule 1 (slice initialization) is working correctly in the handler response path.
