## 1. Fix object_repository.go (10 occurrences)

- [ ] 1.1 `List()` — change `var objects []*models.Object` → `objects := make([]*models.Object, 0)`
- [ ] 1.2 `Search()` — change `var objects []*models.Object` → `objects := make([]*models.Object, 0)`
- [ ] 1.3 `FindByTags()` — change `var objects []*models.Object` → `objects := make([]*models.Object, 0)`
- [ ] 1.4 `GetChildren()` — change `var objects []*models.Object` → `objects := make([]*models.Object, 0)`
- [ ] 1.5 `GetDescendants()` — change `var objects []*models.Object` → `objects := make([]*models.Object, 0)`
- [ ] 1.6 `GetAncestors()` — change `var objects []*models.Object` → `objects := make([]*models.Object, 0)`
- [ ] 1.7 `GetPath()` — change `var objects []*models.Object` → `objects := make([]*models.Object, 0)`
- [ ] 1.8 `BulkCreate()` response collection variable
- [ ] 1.9 `GetObjectStats()` / stats-related slice initialization
- [ ] 1.10 Any remaining `var .* \[\]\*models.Object` in object_repository.go

## 2. Fix object_type_repository.go (8 occurrences)

- [ ] 2.1 `ListObjectTypesByParentType()` — change `var objectTypes []*models.ObjectType` → `objectTypes := make([]*models.ObjectType, 0)`
- [ ] 2.2 `GetRootTree()` / roots collection variable
- [ ] 2.3 `GetChildrenTypes()` / children collection variable
- [ ] 2.4 `GetAncestorsTree()` — change `var objectTypes []*models.ObjectType` → use `make([]*models.ObjectType, 0)`
- [ ] 2.5 `GetPathToRoot()` — change `var path []*models.ObjectType` → `path := make([]*models.ObjectType, 0)`
- [ ] 2.6 `SearchObjectTypes()` or similar search method slice variable
- [ ] 2.7 Any remaining tree traversal collection variables in object_type_repository.go
- [ ] 2.8 Verify no remaining `var .* \[\]\*models.ObjectType` declarations

## 3. Fix relationship_repository.go (5 occurrences)

- [ ] 3.1 `ListRelationships()` — change `var rels []*models.Relationship` → `rels := make([]*models.Relationship, 0)`
- [ ] 3.2 `SearchRelationships()` — change `var rels []*models.Relationship` → `rels := make([]*models.Relationship, 0)`
- [ ] 3.3 Any remaining relationship collection slice variables in relationship_repository.go
- [ ] 3.4 Verify no remaining `var .* \[\]\*models.Relationship` declarations

## 4. Update existing repository unit tests (assertions will break after fix)

- [ ] 4.1 Fix `TestObjectTypeRepository_List` — change `assert.Nil(t, result)` to `assert.NotNil(t, result) && assert.Len(t, result, 0)`
- [ ] 4.2 Fix `TestObjectTypeRepository_GetTree` — same pattern: nil → non-nil empty
- [ ] 4.3 Fix test at line ~439 (GetDescendants) — change nil assertion to non-nil + len check
- [ ] 4.4 Fix test at line ~458 (GetAncestors) — change nil assertion to non-nil + len check
- [ ] 4.5 Fix test at line ~477 (GetPath) — change nil assertion to non-nil + len check
- [ ] 4.6 Fix test at line ~502 (BulkUpdate result) — update nil → non-nil empty
- [ ] 4.7 Verify `assert.Len(t, result, 0)` at line ~515 is already correct (doesn't assert nil)

## 5. Add new unit tests for non-nil empty slice guarantee

- [ ] 5.1 Add test: `TestObjectRepository_List_EmptyResultIsNonNil` — verify List() returns non-nil slice with no rows
- [ ] 5.2 Add test: `TestObjectRepository_Search_EmptyResultIsNonNil` — same for Search()
- [ ] 5.3 Add test: `TestObjectTypeRepository_GetChildrenTypes_EmptyResultIsNonNil`
- [ ] 5.4 Add test: `TestRelationshipRepository_List_EmptyResultIsNonNil`

## 6. Build and deploy verification

- [ ] 6.1 Run `go build ./services/objects-service/...` — expect zero errors
- [ ] 6.2 Run `go test ./services/objects-service/internal/repository/... -v` — all tests pass (including updated assertions)
- [ ] 6.3 Deploy to container and verify: empty queries return `"data": []` not `"data": null`
- [ ] 6.4 Verify: `curl 'http://localhost:8085/api/v1/objects?object_type_id=999&limit=5' -H 'X-User-ID: ...' | jq '.data'` → prints `[]`
