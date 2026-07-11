## Context

The objects-service uses a custom `QueryBuilder` in `services/objects-service/internal/repository/database.go` to build SQL queries with proper placeholder indexing. The `List()` method in the object repository performs two steps: (1) execute a COUNT query via `BuildCount()`, and (2) fetch actual rows via `Build()`. Currently, `BuildCount()` reconstructs the query by extracting only the table name from the SELECT statement and building a bare `SELECT COUNT(*) FROM <table>`, discarding all WHERE clauses. This means pagination metadata always reports the global object count instead of the filtered count.

The QueryBuilder has ~15 methods (Select, From, Where, WhereIn, WhereTagsContain, WhereJsonContains, WhereDateRange, OrderBy, Limit, Offset, BuildCount) but only `Build()` and `BuildCount()` are public consumers. All WHERE methods append to the same `qb.query` string.

## Goals / Non-Goals

**Goals:**
- Fix `BuildCount()` so it produces a COUNT query that includes all WHERE filters from the builder chain
- Preserve existing placeholder indexing behavior (distinct `$N` per Where call)
- Add tests covering filtered COUNT queries with multiple filter combinations

**Non-Goals:**
- Refactor QueryBuilder API or add new methods (except BuildCount fix)
- Change pagination metadata response shape
- Fix related issues in other services (mcp-server, auth-service, user-service)
- Address performance of filtered COUNT on large tables (N+1 is a separate concern)

## Decisions

### Decision 1: Rewrite `BuildCount()` to transform the existing query string instead of reconstructing from scratch

**Approach:** Parse `qb.query` to find `FROM <table>` and everything after it, then replace the SELECT clause with `SELECT COUNT(*)` while preserving FROM, WHERE, ORDER BY (drop for COUNT), LIMIT, OFFSET.

```
Before: "SELECT id, name FROM objects_service.objects WHERE object_type_id = $1 AND deleted_at IS NULL"
After:  "SELECT COUNT(*) FROM objects_service.objects WHERE object_type_id = $1 AND deleted_at IS NULL"
```

**Rationale:** This approach is simpler than rebuilding the query from filter state (which would require storing each filter separately). The existing `qb.query` string already contains all conditions with correct placeholder indices — we just need to swap the SELECT clause.

**Alternatives considered:**
1. Store filters in a separate struct and rebuild both queries independently → More code, more complexity, risk of divergence between Build() and BuildCount() queries
2. Add a `filters []filterClause` field to QueryBuilder → Breaks current API, adds memory overhead for every builder instance

### Decision 2: Drop ORDER BY and LIMIT/OFFSET from COUNT query

COUNT queries don't need sorting or pagination clauses — they only need SELECT columns + FROM + WHERE. Dropping these reduces unnecessary work on the database side.

**Implementation:** After extracting the table name and WHERE clause, stop parsing at ORDER BY / LIMIT / OFFSET keywords.

### Decision 3: Keep `extractTable()` helper but use it for both Build() and BuildCount()

The existing `extractTable()` function already correctly extracts the table name from a SELECT query by finding "FROM " and reading until the next SQL keyword. No change needed there — just reuse it.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| String parsing edge cases (e.g., table names containing keywords) | The existing `extractTable()` already handles this for Build(). Same logic applies to BuildCount() since the FROM clause position is identical |
| Performance regression on large tables with complex WHERE | COUNT queries are inherently O(N) — no index can shortcut a full scan. This was always true; we're just making it correct, not faster |
| Breaking existing tests that expect global count | Tests should be updated to expect filtered counts. No external API contract changes |

## Migration Plan

1. Modify `BuildCount()` in `database.go`
2. Add unit tests for filtered COUNT queries (single filter, multi-filter, tags, JSON, date range)
3. Verify existing QueryBuilder placeholder-indexing tests still pass
4. Build and run objects-service tests — no migration needed (no schema/API changes)

## Open Questions

- None identified. The fix is localized to one method in one file with clear test coverage targets.
