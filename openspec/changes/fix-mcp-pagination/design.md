## Context

The `list_objects` MCP tool receives pagination parameters (`page`, `page_size`) from clients but objects-service only recognizes `limit` and `offset`. This mismatch causes the HTTP client in mcp-server to send wrong query parameter names, so objects-service falls back to defaults (limit=50, offset=0) and returns all matching results.

Current request flow:
```
MCP Client → list_objects({object_type_id:1, page:1, page_size:3})
    ↓
mcp-server tool handler extracts page=1, pageSize=3 ✓
    ↓
HTTP client sends: GET /api/v1/objects?object_type_id=1&page=1&page_size=3  ← WRONG PARAMS
    ↓
objects-service List() reads c.Query("limit") → "" (ignored)
                       c.Query("offset") → "" (ignored)
                       filter.Limit = 50, filter.Offset = 0  ← DEFAULTS
    ↓
Returns up to 50 results instead of 3 ✗
```

## Goals / Non-Goals

**Goals:**
- Align MCP tool input parameters with objects-service API contract (`limit`/`offset`)
- Add pagination metadata to `list_objects` structuredContent output
- Zero translation layer — direct parameter pass-through from MCP → HTTP → objects-service

**Non-Goals:**
- Cursor-based pagination (future enhancement)
- Pagination for other tools (get_object, list_object_types have no pagination need)
- Changing objects-service handler code

## Decisions

### Decision 1: Rename input parameters in ListObjectsParams struct
**Choice**: `page` → `limit`, `pageSize` → `offset`  
**Rationale**: Objects-service is the API truth. Its handler reads `c.Query("limit")` and `c.Query("offset")`. Any translation layer adds indirection and maintenance cost. MCP protocol imposes no pagination semantics — tool input schemas are arbitrary JSON Schema defined by the developer.

```go
// Before:
type ListObjectsParams struct {
    ObjectTypeID  int64   `json:"object_type_id"`
    Page          *int    `json:"page,omitempty"`
    PageSize      *int    `json:"page_size,omitempty"`
}

// After:
type ListObjectsParams struct {
    ObjectTypeID  int64   `json:"object_type_id"`
    Limit         *int    `json:"limit,omitempty"`       // max results (default 50)
    Offset        *int    `json:"offset,omitempty"`      // skip N results (default 0)
}
```

### Decision 2: Update HTTP URL construction in ObjectsClient.ListObjects()
**Choice**: Replace `fmt.Sprintf("?object_type_id=%d&page=%d&page_size=%d", ...)` with `&limit=`/`&offset=`  
**Rationale**: Direct parameter substitution, no mapping logic needed.

```go
// Before:
params := fmt.Sprintf("?object_type_id=%d&page=%d&page_size=%d", objectTypeID, page, pageSize)

// After:
params := fmt.Sprintf("?object_type_id=%d&limit=%d&offset=%d", objectTypeID, limitVal, offsetVal)
```

### Decision 3: Extend ListObjectsResult with pagination metadata
**Choice**: Add `Pagination *PaginationMeta` field to output struct  
**Rationale**: MCP spec allows arbitrary JSON in `structuredContent`. objects-service already returns `{data:[], pagination:{total,limit,offset,count}}`. We should surface this metadata so agents know total count and can navigate pages.

```go
type PaginationMeta struct {
    Total  int64 `json:"total"`
    Limit  int   `json:"limit"`
    Offset int   `json:"offset"`
}

type ListObjectsResult struct {
    Items      []map[string]interface{} `json:"objects"`
    Pagination *PaginationMeta           `json:"pagination,omitempty"`
}
```

The tool handler passes pagination data from objects-service response:
```go
// After extracting objects from response, read pagination metadata
result := ObjectListResponse{...}  // decoded from HTTP response
structuredContent.ListObjectsResult{
    Items:      result.Data,
    Pagination: &PaginationMeta{Total: total, Limit: filter.Limit, Offset: filter.Offset},
}
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Breaking change for MCP clients using `page`/`page_size` | These clients were already broken (pagination ignored). Validation error is correct behavior. Update e2e test and any external integrations. |
| Pagination metadata not available if objects-service response format changes | Parse pagination from the existing `pagination` field in objects-service JSON response (already decoded in ObjectListResponse or similar) |
| Nil pointer dereference when PaginationMeta is nil | Use `omitempty` on Pagination field; only set when we have valid data |

## Unit Tests

Add unit tests in `services/mcp-server/internal/tools/object_tools_test.go`:
- **Test pagination params encoding**: verify `limit=3, offset=0` produces correct URL query string
- **Edge case: zero limit** — should be rejected or defaulted to 50 (objects-service behavior)
- **Edge case: negative limit/offset** — should be rejected or ignored
- **Edge case: nil pagination** — handler returns `PaginationMeta` only when objects-service response includes it
- **Output schema validation**: verify `ListObjectsResult.Pagination` is present and correct in structuredContent

Add unit tests in `services/mcp-server/internal/client/objects_client_test.go`:
- **Test ListObjects URL construction**: verify `limit=`/`offset=` params (not `page=`/`page_size=`)
- **Test pagination metadata passthrough**: mock response with `pagination:{total,limit,offset}` is correctly decoded

## Migration Plan

1. Update MCP tool input schema (`ListObjectsParams`) — clients must use new param names
2. Update HTTP client URL construction — direct parameter replacement
3. Add pagination metadata to output struct and handler logic
4. Update e2e test script for new parameter names and pagination assertions
5. Deploy and verify with `list_objects({object_type_id:1, limit:3})` returns exactly 3 objects

No database migrations, no config changes, no service restart coordination needed — single mcp-server redeploy suffices.
