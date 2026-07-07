## 1. Update filter model structs (models layer)

- [ ] 1.1 `services/objects-service/internal/models/relationship.go`: Change `RelationshipFilter` — remove `Page int form:"page"` and `PageSize int form:"page_size"`, add `Limit int form:"limit"` and `Offset int form:"offset"`
- [ ] 1.2 `services/objects-service/internal/models/relationship.go`: Same change for `RelationshipFilterForType` struct (same fields, same location)
- [ ] 1.3 `services/objects-service/internal/models/relationship_type.go`: Change `RelationshipTypeFilter` — remove Page/PageSize, add Limit/Offset with form tags

## 2. Update handler layer — switch from page/page_size to limit/offset parsing

### 2.1 Relationship Handler (3 methods)
- [ ] 2.1.1 `handlers/relationship_handler.go:List()` — replace `c.ShouldBindQuery(&filter)` with manual query param parsing for limit and offset, set defaults limit=50/offset=0, remove Page/PageSize default assignments
- [ ] 2.1.2 `handlers/relationship_handler.go:GetForObject()` — same pattern: parse limit/offset from query params instead of relying on ShouldBindQuery with old struct fields
- [ ] 2.1.3 `handlers/relationship_handler.go:GetForObjectByType()` — same pattern as GetForObject

### 2.2 Relationship Type Handler (1 method)
- [ ] 2.2.1 `handlers/relationship_type_handler.go:List()` — replace ShouldBindQuery with manual limit/offset parsing, set defaults limit=50/offset=0, remove Page/PageSize default assignments
- [ ] 2.2.2 `handlers/relationship_type_handler.go:List()` — replace hand-rolled `gin.H{"page": filter.Page, "page_size": filter.PageSize}` with `PaginationResponse{Limit: filter.Limit, Offset: filter.Offset, Total: int64(len(rts)), Count: len(responses)}`

## 3. Update service layer defaults

- [ ] 3.1 `services/relationship_type_service.go` — in List() method, remove Page < 1 / PageSize default assignments (lines ~188-192), replace with Limit=50 / Offset=0 defaults if not set by handler

## 4. Update repository layer — replace offset arithmetic

### 4.1 Relationship Repository
- [ ] 4.1.1 `repository/relationship_repository.go:List()` — remove Page < 1 / PageSize default checks, remove `(filter.Page - 1) * filter.PageSize` calculation (line ~350), use `filter.Offset` directly in SQL LIMIT/OFFSET (line ~404), update args append to use Limit and Offset (line ~422)
- [ ] 4.1.2 `repository/relationship_repository.go:GetForObject()` — same pattern: remove Page defaults, replace offset calc with direct filter.Offset usage
- [ ] 4.1.3 `repository/relationship_repository.go:GetForObjectByType()` — same pattern as GetForObject

### 4.2 Relationship Type Repository
- [ ] 4.2.1 `repository/relation_type_repository.go:List()` — remove Page defaults, replace offset calc with direct filter.Offset usage in SQL LIMIT/OFFSET and args append

## 5. Update repository options struct (interfaces)

- [ ] 5.1 `repository/interfaces.go` — rename `DefaultPageSize int` to `DefaultLimit int`, rename `MaxPageSize int` to `MaxLimit int`, update comment if present, update DefaultRepositoryOptions() initializer values

## 6. Update tests

### 6.1 Service layer tests
- [ ] 6.1.1 `services/relationship_service_test.go:278` — change filter assertion from `filter.Page == 2 && filter.PageSize == 10` to `filter.Limit == 10 && filter.Offset == ...` (update based on actual offset calculation)
- [ ] 6.1.2 `services/relationship_service_test.go:288-289` — change test setup from `Page: 2, PageSize: 10` to `Limit: 10, Offset: ...`
- [ ] 6.1.3 `services/relationship_service_test.go:1187` — same pattern as 6.1.1
- [ ] 6.1.4 `services/relationship_service_test.go:1197-1198` — same pattern as 6.1.2

## 7. Build and test verification

- [ ] 7.1 Run `go build ./services/objects-service/...` — expect zero errors
- [ ] 7.2 Run `go test ./services/objects-service/internal/services/... -v` — verify relationship service tests pass with new filter struct fields
- [ ] 7.3 Run `go test ./services/objects-service/internal/repository/... -v` — all existing repo tests still pass (no pagination-specific assertions in mocks)
- [ ] 7.4 Verify: `curl 'http://localhost:8085/api/v1/relationships?limit=3&offset=0' | jq '.pagination'` → returns `{count, limit, offset, total}` with correct values
- [ ] 7.5 Verify: `curl 'http://localhost:8085/api/v1/relationship-types?limit=2&offset=4' | jq '.pagination'` → returns proper pagination object (not raw gin.H)
