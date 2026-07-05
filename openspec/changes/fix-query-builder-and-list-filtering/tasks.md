## Tasks

- [ ] **Task 1: Fix QueryBuilder.Where() placeholder injection**
  - Replace hardcoded `$1` in Where() with dynamic indices starting from argIndex
  - Scan condition string for `$N` patterns and replace sequentially
  - Update argIndex to reflect actual consumed placeholders
  - Run `go build ./services/objects-service/cmd` to verify compilation

- [ ] **Task 2: Fix QueryBuilder.WhereTagsContain() OR logic and argIndex**
  - Replace identical branches with first-tag AND / subsequent-tags OR pattern
  - Change `argIndex += len(tags)` to `argIndex++` (increment per tag)
  - Verify tags filter still works correctly

- [ ] **Task 3: Remove self-filtering from object_handler.go List()**
  - Delete the unconditional `filter.UserID = &userID` block
  - Keep the rest of handler logic intact (query params, pagination)
  - Run existing tests to verify no regressions

- [ ] **Task 4: Remove self-filtering from relationship_handler.go List()**
  - Delete the unconditional `filter.UserID = &userID` block  
  - Run existing tests to verify no regressions

- [ ] **Task 5: Write QueryBuilder multi-filter unit test**
  - Create test that chains 3+ Where() calls with different types
  - Verify generated SQL has distinct `$1`, `$2`, `$3` placeholders
  - Verify args array matches placeholder count and order

- [ ] **Task 6: Build, deploy, and verify fix**
  - `go build ./services/objects-service/cmd && docker cp binary to container`
  - Test: `curl -H "X-User-ID: ..." "http://localhost:8085/api/v1/objects?object_type_id=3&limit=2"` → should return 200 OK with data
  - Test: same request without object_type_id → should still work (regression check)
  - Verify MCP server tool call `list_objects` returns results
