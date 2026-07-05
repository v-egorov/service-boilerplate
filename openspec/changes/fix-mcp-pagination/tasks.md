## 1. Update MCP tool input parameters

- [x] 1.1 Rename `ListObjectsParams.page` → `Limit` with json tag `"limit,omitempty"`
- [x] 1.2 Rename `ListObjectsParams.pageSize` → `Offset` with json tag `"offset,omitempty"`
- [x] 1.3 Update object_tools.go handler to use `args.Limit` and `args.Offset` (rename local vars from `page`/`pageSize`)

## 2. Fix HTTP client URL construction

- [x] 2.1 In objects_client.go `ListObjects()`, replace `fmt.Sprintf("?object_type_id=%d&page=%d&page_size=%d", ...)` with `&limit=` and `&offset=` parameters
- [x] 2.2 Ensure limit/offset values are properly used (default to 50/0 when nil, matching objects-service defaults)

## 3. Add pagination metadata to output schema

- [x] 3.1 Define `PaginationMeta` struct with `Total int64`, `Limit int`, `Offset int` fields
- [x] 3.2 Extend `ListObjectsResult` to include `Pagination *PaginationMeta` with json tag `"pagination,omitempty"`
- [x] 3.3 Update object_tools.go handler to pass pagination data from objects-service response into structuredContent

## 4. Add unit tests for mcp-server changes

- [x] 4.1 Create/extend `object_tools_test.go`: test that `limit=3, offset=0` produces correct URL query string via client
- [x] 4.2 Test edge case: zero limit — should be rejected or defaulted to objects-service default (50)
- [x] 4.3 Test edge case: negative limit/offset — should be rejected or ignored
- [x] 4.4 Test edge case: nil PaginationMeta — handler returns result without pagination key when metadata unavailable
- [x] 4.5 Verify `ListObjectsResult.Pagination` is present and correct in structuredContent via mock tool call
- [x] 4.6 Update/extend `objects_client_test.go`: verify ListObjects URL contains `limit=`/`offset=` (not `page=`/`page_size=`)
- [x] 4.7 Test pagination metadata passthrough: mock response with `pagination:{total,limit,offset}` is correctly decoded

## 5. Update e2e test script

- [x] 5.1 Replace all occurrences of `page:`/`page_size:` with `limit:`/`offset:` in tool call arguments
- [x] 5.2 Add assertions for pagination metadata in structuredContent (check for `"pagination"` key with `total`, `limit`, `offset`)

## 6. Build and verify fix

- [x] 6.1 Rebuild mcp-server binary: `cd services/mcp-server && go build -o ../../build/mcp-server ./cmd`
- [x] 6.2 Run all unit tests: `go test ./services/mcp-server/...` — all pass
- [x] 6.3 Deploy to container and test: `curl ... list_objects({object_type_id:3, limit:3})` returns exactly 3 objects
- [x] 6.4 Verify pagination metadata present in structuredContent response
