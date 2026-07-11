## Why

The objects-service `List` endpoint returns an incorrect `total` value in its pagination metadata. The `BuildCount()` method discards all WHERE filters when constructing the COUNT query, so `total` always equals the count of ALL non-deleted objects (56) regardless of the applied filter (e.g., `object_type_id=1`). This breaks client-side pagination: consumers see "total=56" but only receive 20 items for a filtered query, making it impossible to calculate correct page counts or detect whether more pages exist.

## What Changes

- **Fix `QueryBuilder.BuildCount()`** to produce a COUNT query that respects all WHERE filters applied via the builder chain (Where, WhereIn, WhereTagsContain, WhereJsonContains, WhereDateRange)
- The fix requires extracting the full filtered SQL from the QueryBuilder state rather than reconstructing a bare `SELECT COUNT(*) FROM <table>`
- No API contract changes — the response shape remains identical; only the `total` value becomes correct

## Capabilities

### Modified Capabilities

- **objects-service**: "All list endpoints return consistent pagination metadata" requirement — the `total` field must reflect filtered count, not global object count. Currently violated when any filter is applied.

## Impact

- **Affected code**: `services/objects-service/internal/repository/database.go` — `BuildCount()` method and `extractTable()` helper
- **Test impact**: Existing tests for QueryBuilder (placeholder indexing) remain valid; new tests needed for filtered COUNT queries
- **No breaking changes** to API contracts, handlers, or service layer
