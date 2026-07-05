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

- [x] **Task 6: Build, deploy, and verify fix**
  - ✅ Built from source inside container via `go build`
  - ✅ Test 1: `object_type_id=3 + X-User-ID` → 200 OK with 2 results (was 500 crash)
  - ✅ Test 2: no filters → 200 OK with data (self-filtering removed)
  - ✅ Test 5: `object_type_id=3 + status=active` → 200 OK (cross-type filter combo works)
  - ⚠️ MCP E2E via gateway needs gateway restart — objects-service fix is verified directly
