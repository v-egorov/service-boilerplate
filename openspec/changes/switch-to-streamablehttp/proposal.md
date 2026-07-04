## Why

The SSE transport adds unnecessary complexity to the MCP pipeline: two endpoints (`/sse` + `/message`) require path rewriting, async stream waiting in tests, and `http.ErrAbortHandler` panic recovery. StreamableHTTP consolidates everything into a single `/mcp` endpoint using HTTP headers for session management, making it trivially proxyable through the API Gateway and eliminating all SSE-specific code paths. This change also establishes stateful sessions from day one to support future mutable MCP operations (create/update/delete objects) without requiring auth rework later.

## What Changes

- Replace mcp-server's SSE transport (`NewSSEServer`) with StreamableHTTP transport (`NewStreamableHTTPServer`)
- Use `WithStateful(true)` for sticky session management — first POST establishes identity, subsequent calls reuse the session via header echo
- Remove all SSE-specific routes: `/mcp/sse` and `/message` endpoints on both mcp-server and API Gateway
- Add single StreamableHTTP route: `POST /mcp` on the gateway (no path rewriting needed)
- Delete `http.ErrAbortHandler` panic recovery in gateway (only relevant for SSE streams that no longer exist)
- Rewrite e2e test script to use synchronous POST requests instead of SSE stream parsing
- Update MCP server spec requirement from "SSE transport" to "StreamableHTTP transport with stateful sessions"

## Capabilities

### Modified Capabilities

- `mcp-server`: Transport changes from SSE to StreamableHTTP; session management becomes stateful; gateway-trust skip remains for GET-only reads (unchanged semantics)
- `api-gateway-routing`: MCP routing simplifies from two endpoints (`/sse` + `/message`) to one endpoint (`POST /mcp`); removes ErrAbortHandler recovery requirement

### New Capabilities

<!-- None — this is a transport migration, no new capabilities -->

## Impact

**Affected code:**
- `services/mcp-server/cmd/main.go` — replace SSE server with StreamableHTTP server (~20 line replacement)
- `api-gateway/internal/handlers/gateway.go` — remove path rewriting for `/mcp/message`, delete ErrAbortHandler recovery
- `api-gateway/cmd/main.go` — replace two MCP routes with single `POST /mcp` route
- `scripts/test-mcp-e2e.sh` — complete rewrite from SSE stream parsing to synchronous POST requests

**API changes:**
- **BREAKING**: Clients connecting via SSE (`GET /mcp/sse`) will no longer work. All clients must use StreamableHTTP (POST to `/mcp`).
- Session ID moves from URL query param (`?sessionId=abc`) to HTTP header (`MCP-Session-ID: abc`)

**Dependencies:** No new dependencies — both transports are built into mcp-go v0.55.1
