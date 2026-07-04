## MODIFIED Requirements

### Requirement: Reverse Proxy Execution (MCP path)
For MCP requests routed to `/mcp`, the gateway SHALL execute a reverse proxy to the mcp-server service URL without modifying the request path. The incoming path (`/mcp`) matches the outgoing path exactly — no prefix stripping or rewriting is performed. Request body is read and rewound for tracing purposes (same as existing behavior), and `X-User-*` identity headers are injected via context before forwarding.

#### Scenario: Successful MCP proxy with no path rewrite
- **WHEN** a request hits `/mcp` POST through api-gateway
- **THEN** the gateway forwards it to mcp-server at the resolved URL with the original path preserved as `/mcp`, and returns the backend's response to the client

### Requirement: MCP reverse proxy does not require ErrAbortHandler recovery
The api-gateway's MCP reverse proxy handler does NOT need special panic recovery for `http.ErrAbortHandler` because StreamableHTTP uses discrete synchronous POST requests (no long-lived SSE streams). Client disconnects during a POST are handled normally by the HTTP stack without raising recoverable panics.

#### Scenario: No ErrAbortHandler recovery needed
- **WHEN** an MCP client disconnects during a POST request to `/mcp` through the api-gateway
- **THEN** the reverse proxy handles the disconnect via standard HTTP error paths, no panic is raised, and no special recovery code is required

## REMOVED Requirements

### Requirement: MCP SSE reverse proxy recovers from ErrAbortHandler panics
**Reason**: Transport changed from SSE to StreamableHTTP. StreamableHTTP has no long-lived streams — each request is a discrete POST with synchronous response. No `http.ErrAbortHandler` panic can occur from client disconnect during streaming because there's no streaming connection.

### Requirement: Route-to-Service Mapping (MCP routes)
**Reason**: MCP routes simplified from two endpoints (`GET /sse`, `POST /message`) to single endpoint (`POST /mcp`). The spec requirement is now covered by the "Reverse Proxy Execution (MCP path)" modified requirement above.
