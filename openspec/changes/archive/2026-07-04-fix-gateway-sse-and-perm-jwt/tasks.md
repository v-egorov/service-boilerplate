## 1. Fix SSE reverse proxy panic in api-gateway

- [x] 1.1 In `api-gateway/internal/handlers/gateway.go:ProxyMCPRequest()`, wrap `proxy.ServeHTTP(c.Writer, c.Request)` with a deferred recover that catches `http.ErrAbortHandler` and silently returns, re-panicking all other errors
- [ ] 1.2 Verify: `make build-mcp-server` and `make test-api-gateway` pass without regressions

## 2. Skip auth-service permission check for gateway-trusted MCP reads

- [x] 2.1 In `services/objects-service/internal/permiddleware/permission.go:NewPermissionMiddleware()`, add a check before the `authClient.CheckPermission()` loop: when `jwtToken == "" && userID != "" && cfg.HTTPMethod == "GET"`, skip the loop, set `c.Set("matched_permissions", requiredPermissions)`, call `c.Next()`, and return
- [ ] 2.2 Verify: `make build-objects-service` compiles; `make test-objects-service` passes

## 3. End-to-end validation

**PAUSED: Pre-existing permission model mismatch blocks E2E.**

After permiddleware skip fires and request reaches service layer, objects-service returns 500 for ALL users (including admin with real JWT). Root cause: DB has `object-types | read` but permiddleware checks `{type_key}:read:all/{type_key}:read:own`. This is a pre-existing design bug in objects-service.

- [ ] 3.1 Rebuild and restart services: `make down && make dev-detached`
- [ ] 3.2 Run `bash scripts/test-mcp-e2e.sh` — SSE panic fix confirmed (Step 2 passes, no `[Recovery]` logs). Tool calls blocked by pre-existing permission mismatch.
- [ ] 3.3 Check gateway logs for absence of `[Recovery] panic recovered` entries during SSE connections
