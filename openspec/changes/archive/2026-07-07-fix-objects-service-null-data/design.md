## Context

Every list endpoint in objects-service returns `"data": null` instead of `"data": []` when no rows match the query. The root cause is Go's zero-value behavior: `var items []*Model` produces a nil slice, and gin.H serializes nil slices as JSON `null`.

Current code pattern across 3 repository files (24 occurrences):
```go
func (r *repo) List(...) ([]*models.Object, int64, error) {
    var objects []*models.Object   // ← nil by default
    for rows.Next() { ... }        // ← no rows → loop never runs
    return objects, total, nil     // ← nil slice → JSON null
}
```

Handler layer passes this through unchanged:
```go
c.JSON(http.StatusOK, gin.H{
    "data":   objects,           // ← nil → {"data": null}
    "meta":   gin.H{"request_id": reqID},
})
```

## Goals / Non-Goals

**Goals:**
- Ensure all collection endpoints return `"data": []` (empty JSON array) instead of `"data": null` for zero-result queries
- Fix the root cause in repository layer by initializing result slices as empty (`make([]*T, 0)`) rather than nil (`var items []*T`)

**Non-Goals:**
- Changing handler response format or adding new fields
- Modifying QueryBuilder, database schema, auth-service, gateway, or mcp-server
- Adding pagination metadata for empty results (already handled by existing code)

## Decisions

### Decision 1: Fix at repository layer, not handler layer
**Choice**: Change `var items []*T` → `items := make([]*T, 0)` in every repository method that returns a collection.  
**Rationale**: This is idiomatic Go best practice for collection-returning functions — non-nil slices serialize consistently in JSON (`[]` vs `null`), behave predictably with equality checks and third-party libraries, and eliminate the need for callers to handle both nil and empty-array cases. It's also the single point of failure: all handlers simply pass through whatever the repository returns.

### Decision 2: One-line change per method
**Choice**: Direct variable initialization replacement — no wrapper functions, no helper methods.  
**Rationale**: The fix is mechanical and identical in every location. Adding a helper would be unnecessary indirection for 24 one-line changes. Each change is self-evident and trivially reviewable.

### Decision 3: No test framework changes
**Choice**: Rely on existing integration tests plus manual verification via curl.  
**Rationale**: The fix is purely about Go slice initialization — it doesn't change logic, query behavior, or error handling. Existing tests already pass when results are non-empty; the null-vs-array distinction is a serialization concern that's trivially verified with `curl ... | jq '.data'`.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Some existing clients may depend on `null` being truthy (falsy in JS) | Negligible — `null` and `[]` are both falsy in JavaScript, so no behavioral change for consumers. The fix only affects explicit null-checks like `if data === null`. |
| Existing repository unit tests assert `assert.Nil(t, result)` on empty collections (6 locations) | Update all assertions to `assert.NotNil(t, result) && assert.Len(t, result, 0)`. Add new dedicated tests verifying non-nil guarantee for each collection-returning method. |
| 24 locations to change across 3 files increases review surface | Each change is a one-line mechanical replacement with identical semantics. Use `sed` for consistency and review via `git diff`. No logic changes, no new code paths. |

## Migration Plan

1. Apply all 24 variable initialization changes in the 3 repository files
2. Run `go build ./services/objects-service/...` — expect zero errors (same types, just different initialization)
3. Run `go test ./services/objects-service/...` — verify existing tests still pass
4. Deploy to container and manually verify: `curl ... | jq '.data'` returns `[]` for empty queries
5. No rollback needed — the change is purely a serialization fix with no behavioral impact
