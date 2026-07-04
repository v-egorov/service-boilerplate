## 1. Fix SSE reverse proxy panic in api-gateway

- [x] 1.1 In `api-gateway/internal/handlers/gateway.go:ProxyMCPRequest()`, wrap `proxy.ServeHTTP(c.Writer, c.Request)` with a deferred recover that catches `http.ErrAbortHandler` and silently returns, re-panicking all other errors
- [ ] 1.2 Verify: `make build-mcp-server` and `make test-api-gateway` pass without regressions

## 2. Skip auth-service permission check for gateway-trusted MCP reads

- [x] 2.1 In `services/objects-service/internal/permiddleware/permission.go:NewPermissionMiddleware()`, add a check before the `authClient.CheckPermission()` loop: when `jwtToken == "" && userID != "" && cfg.HTTPMethod == "GET"`, skip the loop, set `c.Set("matched_permissions", requiredPermissions)`, call `c.Next()`, and return
- [ ] 2.2 Verify: `make build-objects-service` compiles; `make test-objects-service` passes

## 3. End-to-end validation

- [ ] 3.1 Rebuild and restart services: `make down && make dev-detached`
- [ ] 3.2 Run `bash scripts/test-mcp-e2e.sh` — confirm all steps pass (no panic logs, no 401 errors)
- [ ] 3.3 Check gateway logs for absence of `[Recovery] panic recovered` entries during SSE connections
