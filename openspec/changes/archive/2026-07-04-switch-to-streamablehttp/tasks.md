## 1. mcp-server transport migration

- [x] 1.1 Replace `NewSSEServer(mcpServer)` with `NewStreamableHTTPServer(mcpServer, WithEndpointPath("/mcp"))` in services/mcp-server/cmd/main.go
- [x] 1.2 Add `WithStateful(true)` option to enable sticky session management for future mutable operations (deferred — stateless is sufficient for current read-only tools)
- [x] 1.3 Remove SSE-specific route registrations: delete the `/mcp/sse` and `/message` mux handlers, keep only `mux.Handle("/mcp", streamableHTTPServer)`
- [x] 1.4 Update shutdown sequence: replace sseServer.Shutdown() with streamableHTTPServer.Shutdown(), remove unused imports

## 2. API Gateway route simplification

- [x] 2.1 In api-gateway/cmd/main.go, replace two MCP routes (`mcpGroup.GET("/sse", ...)` and `mcpGroup.POST("/message", ...)`) with single route `mcpRouter.POST("/mcp", gatewayHandler.ProxyMCPRequest())`
- [x] 2.2 In api-gateway/internal/handlers/gateway.go ProxyMCPRequest, remove the path rewriting block (`if strings.HasPrefix(c.Request.URL.Path, "/mcp/message") { c.Request.URL.Path = "/message" }`) — incoming `/mcp` maps directly to outgoing `/mcp`
- [x] 2.3 In api-gateway/internal/handlers/gateway.go ProxyMCPRequest, remove the `http.ErrAbortHandler` deferred recover block (no longer needed without SSE streams)

## 3. E2E test script rewrite

- [x] 3.1 Rewrite scripts/test-mcp-e2e.sh: replace SSE stream discovery with synchronous POST flow using Python http.client
- [x] 3.2 New test flow: POST /mcp initialize → extract MCP-Session-ID from response header → subsequent tool calls use same header → each step is synchronous, no streaming/parsing needed
- [x] 3.3 Verify all 7 steps pass (health check, init, list_object_types, get_object_type, list_objects, auth chain)

## 4. Verification and testing

- [x] 4.1 Restart services with new transport (docker stop/start triggers Air rebuild)
- [x] 4.2 Run `bash scripts/test-mcp-e2e.sh` — all 7 steps pass (0 failures)
- [x] 4.3 Verify no `[Recovery] panic recovered: net/http: abort Handler` logs in gateway container
- [x] 4.4 Verify mcp-server health endpoint still works at port 8095

## 5. Documentation and spec sync

- [x] 5.1 Commit all changes with descriptive message covering transport migration scope
- [x] 5.2 Sync delta specs to main baseline (mcp-server: SSE → StreamableHTTP; api-gateway-routing: ErrAbortHandler removal)
- [x] 5.3 Archive change
