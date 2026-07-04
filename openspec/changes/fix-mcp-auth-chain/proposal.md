## Why

Every MCP tool/resource/prompt call currently returns 401 from objects-service. The gateway correctly injects identity headers (`X-User-ID`, `X-User-Email`, `X-User-Roles`) for `/mcp/*` requests and mcp-go surfaces them onto each handler's `request.Header`, but `ObjectsClient.httpGet()` builds brand-new outbound HTTP requests with zero headers — identity dies at hop 2 (mcp-server → objects-service). The existing spec's "preserving headers and body" scenario only covers the gateway → mcp-server hop and never required the onward hop to objects-service, so this gap is also a spec gap.

## What Changes

- Add `WithIdentity(ctx, hdr)` / `IdentityFromContext(ctx)` helpers and a single allow-listed header set (`X-User-ID`, `X-User-Email`, `X-User-Roles`) to the mcp-server `client` package
- Change the 7 data-fetching methods on `ObjectsClient` to take a leading `ctx context.Context` and forward only allow-listed identity headers in `httpGet`
- Inject identity at each tool/resource/prompt handler entry point: `ctx = client.WithIdentity(ctx, request.Header)`
- Add a regression test asserting the three identity headers arrive at a stub objects-service (and nothing else is copied wholesale) — fails today, passes after the fix
- Document the semantic: when no identity is present in context (health checks, background callers), calls proceed with no identity headers rather than being rejected

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `mcp-server`: add a requirement that the mcp-server MUST forward the gateway-injected identity headers from each handler's inbound request onto all outbound objects-service calls, scoped to an allow-list, with explicit no-identity-proceeds behavior. This closes the spec gap where the existing "preserving headers" scenario only covered gateway → mcp-server.

## Impact

- **Code:** `services/mcp-server/internal/client/` (new `identity.go`, modified `objects_client.go`), `services/mcp-server/internal/tools/{type_tools,object_tools}.go`, `services/mcp-server/internal/resources/hierarchy.go`, `services/mcp-server/internal/prompts/{browse_schema,get_object_info}.go`
- **Untouched:** `internal/handlers/health_handler.go` (builds its own request to `/health`, intentionally unauthenticated), `HealthCheckURL()` (returns a string, no HTTP), api-gateway, common middleware, and all three backend services — they receive headers exactly as before
- **Tests:** existing health-handler tests stay green; new unit test in `client/` proves the chain
- **Dependencies:** none added — `context` and `net/http` only
- **Scope:** mcp-server only; no breaking API changes
