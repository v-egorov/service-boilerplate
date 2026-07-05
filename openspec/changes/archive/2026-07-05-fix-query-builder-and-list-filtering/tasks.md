## Tasks

- [x] **Task 1: Fix QueryBuilder.Where() placeholder injection**
  - ✅ Replaced hardcoded `$1` with dynamic indices from argIndex, scanning for $N patterns and replacing sequentially
  - Run `go build ./services/objects-service/cmd` verified — builds clean

- [x] **Task 2: Fix QueryBuilder.WhereTagsContain() OR logic and argIndex**
  - ✅ Replaced identical branches with first-tag AND / subsequent-tags OR pattern
  - ✅ Changed `argIndex += len(tags)` to `argIndex++`

- [x] **Task 3: Remove self-filtering from object_handler.go List()**
  - ✅ Deleted unconditional `filter.UserID = &userID` block
  - ✅ Existing tests pass — no regressions

- [x] **Task 4: Remove self-filtering from relationship_handler.go List()**
  - ✅ Deleted unconditional `filter.UserID = &userID` block
  - ✅ Existing tests pass — no regressions

- [x] **Task 5: Write QueryBuilder multi-filter unit test**
  - ✅ Created 6 tests: MultipleWhereDistinctPlaceholders, ThreeWhereDistinctPlaceholders, WhereWithMixedMethods, WhereTagsContain_ORLogic, WhereTagsContain_SingleTag, WhereTagsContain_TwoTagsDistinctIndices
  - ✅ All 34 repository tests pass

- [x] **Task 6: Build, deploy, and verify fix**
  - ✅ Built from source inside container via `go build`
  - ✅ Test 1: `object_type_id=3 + X-User-ID` → 200 OK with 2 results (was 500 crash)
  - ✅ Test 2: no filters → 200 OK with data (self-filtering removed)
  - ✅ Test 5: `object_type_id=3 + status=active` → 200 OK (cross-type filter combo works)
  - ⚠️ MCP E2E via gateway needs gateway restart — objects-service fix is verified directly
