# Design: Fix Pagination Inconsistency in objects-service

## Decision 1: Canonical format is limit/offset

**Choice:** All endpoints use `?limit=N&offset=M` query params. Remove all `Page`/`PageSize` fields from filter structs.

**Rationale:**
- Already used by the most endpoints (objects, object-types) — less migration surface
- Maps directly to SQL `LIMIT/OFFSET` — no `(page-1)*size` arithmetic needed in handlers/repos
- More explicit: offset=0 is clearer than page=1 for "first page"
- Cursor/keyset pagination can build on offset naturally later

### Before/After Field Mapping

```go
// BEFORE (relationship.go)
type RelationshipFilter struct {
    Page     int `form:"page"`       // → remove
    PageSize int `form:"page_size"`  // → remove
}

// AFTER
type RelationshipFilter struct {
    Limit  int `form:"limit"`   // new
    Offset int `form:"offset"`  // new
}
```

### Default Values Standardized to limit=50, offset=0

All handlers set:
```go
filter := &models.SomeFilter{
    Limit:  50,
    Offset: 0,
}
// Then override from query params if provided
if limitStr := c.Query("limit"); limitStr != "" { ... }
if offsetStr := c.Query("offset"); offsetStr != "" { ... }
```

## Decision 2: Response pagination fields are consistent everywhere

**Choice:** All list responses include `pagination` object with exactly these fields:
- `count` — number of items in this response page (len(data))
- `limit` — requested limit value
- `offset` — requested offset value
- `total` — total matching rows across all pages

```json
{
  "data": [...],
  "pagination": {
    "count": 3,
    "limit": 50,
    "offset": 100,
    "total": 247
  },
  "meta": {"request_id": "..."}
}
```

**Before this fix:** relationship-types returned raw `gin.H{"page": ..., "page_size": ...}` — no count, no total. This was a bug regardless of param format.

## Decision 3: Remove Page/PageSize arithmetic from all layers

The `(filter.Page - 1) * filter.PageSize` calculation currently appears in:
- **Handlers:** relationship_handler.go (2 places), relationship_type_handler.go (2 places)
- **Repository:** relationship_repository.go (3 methods with offset calc), relation_type_repository.go (1 method)

**After:** Repos receive `filter.Offset` directly. No arithmetic needed anywhere.

```go
// BEFORE — handler converts page to offset, passes to repo
offset := (filter.Page - 1) * filter.PageSize
args = append(args, limit, offset)  // manual math in multiple places

// AFTER — handler parses offset from query, passes directly
if offsetStr := c.Query("offset"); offsetStr != "" { ... }
args = append(args, filter.Limit, filter.Offset)  // direct values
```

## Decision 4: DefaultPageSize/MaxPageSize config stays but is renamed

The repository options struct has `DefaultPageSize` and `MaxPageSize` fields used only for the relationship endpoints. These become `DefaultLimit` and `MaxLimit` — same semantics, clearer naming.

## Implementation Map

### Files Changed (8 total)

| File | Changes |
|------|---------|
| `models/relationship.go` | RelationshipFilter: Page→Limit, PageSize→Offset (2 structs) |
| `models/relationship_type.go` | RelationshipTypeFilter: same field rename |
| `handlers/relationship_handler.go` | Switch ShouldBindQuery to manual limit/offset parsing; fix default values (3 methods) |
| `handlers/relationship_type_handler.go` | Same + replace hand-rolled gin.H pagination with shared PaginationResponse struct |
| `services/relationship_type_service.go` | Remove Page/PageSize defaults in List() |
| `repository/relationship_repository.go` | Replace Page→Offset arithmetic in 3 methods (List, GetForObject, GetForObjectByType) |
| `repository/relation_type_repository.go` | Replace Page→Offset arithmetic in 1 method |
| `repository/interfaces.go` | DefaultPageSize → DefaultLimit, MaxPageSize → MaxLimit |

### Tests Changed (~5 locations)

| File | Changes |
|------|---------|
| `services/relationship_service_test.go` | ~4 test cases: Page→Limit, PageSize→Offset in filter construction and assertions |
| `repository/*_test.go` | Any offset arithmetic in SQL query expectations (verify none exist — tests use MockDB) |

### No Changes Needed

- `/api/v1/objects` handler — already uses limit/offset ✅
- `/api/v1/object-types` handler — already uses limit/offset ✅
- Handler test files — no pagination assertions found
- API Gateway, mcp-server, user-service, auth-service — no affected code
