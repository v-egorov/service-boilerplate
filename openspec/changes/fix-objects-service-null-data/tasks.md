## 1. Fix object_repository.go (10 occurrences)

- [x] 1.1 `List()` — change `var objects []*models.Object` → `objects := make([]*models.Object, 0)`
- [x] 1.2 `Search()` — change `var objects []*models.Object` → `objects := make([]*models.Object, 0)`
- [x] 1.3 `FindByMetadata()` + `FindByTags()` — both fixed
- [x] 1.4 `GetChildren()` — change `var objects []*models.Object` → `objects := make([]*models.Object, 0)`
- [x] 1.5 `GetDescendants()` — change `var objects []*models.Object` → `objects := make([]*models.Object, 0)`
- [x] 1.6 `GetAncestors()` — change `var objects []*models.Object` → `objects := make([]*models.Object, 0)`
- [x] 1.7 `GetPath()` — change `var objects []*models.Object` → `objects := make([]*models.Object, 0)`
- [x] 1.8 `BulkCreate()` response collection variable
- [x] 1.9 `GetObjectStats()` — no slice vars (uses maps initialized with `make()`) ✅
- [x] 1.10 Any remaining `var .* \[\]\*models.Object` in object_repository.go — verified zero remaining

## 2. Fix object_type_repository.go (8 occurrences)

- [x] 2.1 `GetTree()` — change `var objectTypes []*models.ObjectType` → `objectTypes := make([]*models.ObjectType, 0)`
- [x] 2.2 `GetRootTree()` / roots collection variable
- [x] 2.3 `GetChildrenTypes()` / children collection variable
- [x] 2.4 `GetDescendants()` — change `var objectTypes []*models.ObjectType` → use `make([]*models.ObjectType, 0)`
- [x] 2.5 `GetPathToRoot()` — change `var path []*models.ObjectType` → `path := make([]*models.ObjectType, 0)`
- [x] 2.6 `SearchObjectTypes()` or similar search method slice variable
- [x] 2.7 Any remaining tree traversal collection variables in object_type_repository.go
- [x] 2.8 Verify no remaining `var .* \[\]\*models.ObjectType` declarations — verified zero remaining

## 3. Fix relationship_repository.go (5 occurrences)

- [x] 3.1 `ListRelationships()` — change `var rels []*models.Relationship` → `rels := make([]*models.Relationship, 0)`
- [x] 3.2 `GetForObject()` + `GetForObjectByType()` — both fixed (task merged)
- [x] 3.3 Any remaining relationship collection slice variables in relationship_repository.go
- [x] 3.4 Verify no remaining `var .* \[\]\*models.Relationship` declarations — verified zero remaining
- [x] 3.5 Bonus: Fixed `GetRelatedObjects()` (returns objects, not relationships) + `relationship_type_repository.go:446`

## 4. Update existing repository unit tests (assertions will break after fix)

- [x] 4.1 Fix `TestObjectTypeRepository_List` — change `assert.Nil(t, result)` to `assert.NotNil(t, result) && assert.Len(t, result, 0)`
- [x] 4.2 Fix `TestObjectTypeRepository_GetTree` — same pattern: nil → non-nil empty
- [x] 4.3 Fix test at line ~439 (GetDescendants) — change nil assertion to non-nil + len check
- [x] 4.4 Fix test at line ~458 (GetAncestors) — change nil assertion to non-nil + len check
- [x] 4.5 Fix test at line ~477 (GetPath) — change nil assertion to non-nil + len check
- [x] 4.6 Fix test at line ~502 (BulkUpdate result) — update nil → non-nil empty
- [x] 4.7 Verify `assert.Len(t, result, 0)` at line ~515 is already correct (doesn't assert nil)
- All 6 nil assertions updated in one sed pass; all tests now pass

## 5. Add new unit tests for non-nil empty slice guarantee

- [x] 5.1 Add test: `TestObjectRepository_List_EmptyResultIsNonNil`
- [x] 5.2 Add test: `TestObjectRepository_Search_EmptyResultIsNonNil`
- [x] 5.3 Add test: `TestObjectTypeRepository_GetChildrenTypes_EmptyResultIsNonNil`
- [x] 5.4 Add test: `TestRelationshipRepository_List_EmptyResultIsNonNil`

## 6. Build and deploy verification

- [x] 6.1 Run `go build ./services/objects-service/...` — ✅ zero errors
- [x] 6.2 Run `go test ./services/objects-service/internal/repository/... -v` — ✅ all 38 tests pass (26 existing updated + 4 new dedicated)
- [x] 6.3 Deploy to container and verify: empty queries return `"data": []` not `"data": null`
  - List() offset past end → `[ ]` ✅
  - GetAncestors() no ancestors → `[ ]` ✅
  - GetChildren() ID=3, no children → `[ ]` ✅
  - GetRelatedObjects() valid UUID, no rels → `[ ]` ✅
- [x] 6.4 Verify: all empty list endpoints return `"data": []` — verified via curl against running container
