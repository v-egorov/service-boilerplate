## Why

Two infrastructure bugs block the MCP server from functioning end-to-end. First, the api-gateway's SSE reverse proxy crashes every time a client disconnects because `httputil.ReverseProxy` uses `panic(http.ErrAbortHandler)` as control flow, and gin's recovery middleware treats it as a real panic, returning HTTP 500 and breaking all SSE connections. Second, after `fix-mcp-auth-chain` restores identity forwarding from mcp-server to objects-service, the authorization hop from objects-service to auth-service fails — `permiddleware` calls `authClient.CheckPermission()` with an empty JWT token, and auth-service's `RequireAuth()` rejects because no user identity was set in its gin context. Combined, these make the MCP server unusable: SSE connections crash, and any surviving tool calls get 401 from auth-service.

## What Changes

- **Gateway SSE reverse proxy**: wrap `proxy.ServeHTTP` in `ProxyMCPRequest()` with a deferred recover that catches `http.ErrAbortHandler` and silently returns, re-panicking other errors. This allows SSE streams to terminate cleanly without triggering gin's recovery middleware.
- **Permiddleware permission check skip**: in objects-service's `permiddleware.go`, when a request carries valid identity headers (`X-User-ID`) but no `Authorization` header (no JWT), skip the `auth-service.CheckPermission()` call for GET requests. The gateway-vouched identity is trusted directly in dev mode. This path only activates when `jwtToken` is empty — in production (or any real user request), the gateway always injects a JWT, so the normal RBAC path remains unchanged.

## Capabilities

### Modified Capabilities
- `mcp-server`: add a requirement that objects-service MUST skip the auth-service permission check for requests with gateway-injected identity headers but no JWT token (dev-mode MCP reads only). This closes the gap where the existing spec's "Auth-service permission check succeeds for MCP agent" scenario assumed a round-trip that doesn't work without a JWT.
- `api-gateway-routing`: add a requirement that the MCP SSE reverse proxy MUST recover from `http.ErrAbortHandler` panics without triggering gin recovery, allowing SSE connections to terminate cleanly.

### New Capabilities
(none)

## Impact

- **Code:** `api-gateway/internal/handlers/gateway.go` (~5 lines), `services/objects-service/internal/permiddleware/permission.go` (~3 lines)
- **Untouched:** mcp-server, user-service, auth-service, docker config, migrations
- **Tests:** existing gateway_mcp_test.go and permiddleware tests stay green
- **Dev mode only:** both changes only affect paths without actual JWT tokens — unreachable in production where the gateway always validates JWTs
- **Dependencies:** none added
- **Known compromise documented in** `docs/known-issues.md`: MCP reads bypass RBAC in dev mode; proper fix requires API-key-based client identity and real JWT injection (separate scope)
