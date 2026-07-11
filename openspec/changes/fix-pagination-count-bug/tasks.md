## 1. Understand existing tests and baseline

- [x] 1.1 Read `repository_test.go` to understand mock patterns (MockDBPool, MockRow, MockRows)
- [x] 1.2 Run existing objects-service repository tests: `cd services/objects-service && go test ./internal/repository/... -v` — confirm all pass

## 2. Rewrite BuildCount() in database.go

- [x] 2.1 Replace the current `BuildCount()` body with a query-transforming approach: replace SELECT columns with COUNT(*) while preserving FROM + WHERE
- [x] 2.2 Ensure ORDER BY, LIMIT, OFFSET are dropped from the COUNT query (they're unnecessary for counting)
- [x] 2.3 Keep `extractTable()` helper unchanged — reuse it to find the table name

## 3. Add unit tests for BuildCount() with filters

- [x] 3.1 Test: single WHERE filter (`object_type_id = $1`) → COUNT query includes WHERE clause with correct placeholder
- [x] 3.2 Test: multiple chained WHERE clauses (Where + WhereTagsContain) → all conditions preserved in COUNT
- [x] 3.3 Test: WhereJsonContains filter → `metadata::jsonb @> $N::jsonb` included in COUNT
- [x] 3.4 Test: WhereDateRange filter → both start/end conditions included in COUNT
- [x] 3.5 Test: no filters applied → COUNT query is bare SELECT with just FROM (no WHERE)
- [x] 3.6 Test: args array matches placeholder count in the generated COUNT SQL

## 4. Integration verification

- [x] 4.1 Run full objects-service test suite: `go test ./...` — confirm all pass
- [x] 4.2 Verify via curl that filtered List endpoints now return correct total values (e.g., type_id=1 → total=20, not 56)
- [x] 4.3 Build and run objects-service in Docker: `make build-objects-service && make dev-detached` — verify services come up clean

## 5. Commit

- [ ] 5.1 Review changes with `git diff`
- [ ] 5.2 Stage and commit all changed files
