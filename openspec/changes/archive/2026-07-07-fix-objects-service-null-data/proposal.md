## Why

Every list endpoint in objects-service returns `null` instead of an empty array (`[]`) when no results match the query. This is caused by Go's zero-value behavior: `var items []*Model` produces a nil slice, and gin's JSON serializer encodes nil slices as JSON `null`. Clients must defensively handle both `data: null` and `data: []` as valid "empty" responses — fragile and inconsistent with the documented response format.

## What Changes

- **Fix all repository list methods** to initialize result slices as empty (`make([]*T, 0)`) instead of nil (`var items []*T`). This is idiomatic Go best practice for collection-returning functions: non-nil slices serialize consistently in JSON (`[]` vs `null`), behave predictably with equality checks and third-party libraries, and eliminate the need for callers to handle both nil and empty-array cases.
- **Update affected endpoints**: object_handler.go (List, Search, FindByTags, GetChildren/Descendants/Ancestors/Path), object_type_handler.go (tree/list operations), relationship_handler.go (list/search) — all return `"data": []` instead of `"data": null` when no rows match

## Capabilities

### Modified Capabilities
- `objects-service`: change API response contract for all list endpoints — empty results SHALL produce `"data": []` (JSON array), never `"data": null` (JSON null)

## Impact

- **Code:** 24 variable declarations across 3 repository files (`object_repository.go`, `object_type_repository.go`, `relationship_repository.go`)
- **Untouched:** objects-service handler layer, query builder, database schema, auth-service, user-service, gateway, mcp-server
- **Tests:** Existing integration tests may need adjustment if they assert `data` is truthy rather than checking for empty array explicitly
