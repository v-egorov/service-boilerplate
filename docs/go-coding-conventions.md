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

## Rule 2: Standardize Pagination Parameters Across Endpoints

**Scope:** All list endpoints returning paginated collections.

Choose **one convention** and use it everywhere:
- `?limit=N&offset=M` (used by objects-service List/Search)
- OR `?page=P&page_size=S` (currently used only by relationship handler)

Do NOT mix conventions within the same service. Response field names must also be consistent (`count`, `total`, `has_more`).

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

// ✅ CORRECT — dispatch via errors.Is()
func handleServiceError(c *gin.Context, err error) {
    if errors.Is(err, repository.ErrNotFound) {
        c.JSON(http.StatusNotFound, ...)
        return
    }
    if errors.Is(err, repository.ErrInvalidInput) {
        c.JSON(http.StatusBadRequest, ...)
        return
    }
    // fallback for truly unexpected errors only
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "type": "internal_error"})
}
```

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

## Rule 7: Test Coverage Expectations

**Scope:** All Go packages in every service.

| Layer | Minimum Coverage | Pattern |
|-------|-----------------|---------|
| Repository | Mock DB interface, test empty/non-empty paths | `MockDBPool` implementing `DBInterface` |
| Service | Interface-based mock of repository layer | Test both happy path and error wrapping |
| Handler | Integration through full middleware chain (no writer mocks) | Smoke-test via real gin engine |

**Empty result guarantee:** Every collection-returning method must have at least one test asserting non-nil empty results (`assert.NotNil(t, result)` + `assert.Len(t, result, 0)`).
