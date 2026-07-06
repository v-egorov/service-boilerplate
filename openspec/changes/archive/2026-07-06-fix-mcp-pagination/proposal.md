## Why

The `list_objects` MCP tool returns all matching objects instead of respecting pagination parameters. The root cause is a parameter name mismatch: the MCP client sends `page` and `page_size`, but objects-service only reads `limit` and `offset`. This means agents requesting "first 3 items" get up to 50 results, breaking cursor-based navigation and wasting tokens.

## What Changes

- **Rename tool input parameters**: Change `ListObjectsParams.page` → `limit` and `page_size` → `offset` to match objects-service API contract
- **Update HTTP client URL construction**: Replace `page=`/`page_size=` query params with `limit=`/`offset=` in `objects_client.go:ListObjects()`
- **Add pagination metadata to output schema**: Extend `ListObjectsResult` struct with `Pagination *PaginationMeta` field so MCP responses include `{pagination: {total, limit, offset}}` alongside the objects array
- **Update e2e test**: Fix any assertions that reference old parameter names or missing pagination data

## Capabilities

### Modified Capabilities
- `mcp-server`: change `list_objects` tool input parameters from `page`/`page_size` to `limit`/`offset`; add pagination metadata to output schema via `ListObjectsResult.Pagination` field

## Impact

- **Code:** `services/mcp-server/internal/tools/object_tools.go` (input param names), `services/mcp-server/internal/client/objects_client.go` (URL construction + response parsing), `services/mcp-server/internal/tools/object_tools.go` (output struct fields)
- **Untouched:** objects-service handler code, gateway routing, auth-service, docker config
- **Tests:** `scripts/test-mcp-e2e.sh` needs update for new parameter names and pagination metadata in structuredContent
- **Breaking change**: MCP clients using old `page`/`page_size` parameters will get validation errors. This is intentional — aligns with the actual API contract.
