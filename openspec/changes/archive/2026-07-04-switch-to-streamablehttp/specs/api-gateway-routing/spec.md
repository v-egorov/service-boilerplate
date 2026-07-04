## MODIFIED Requirements

### Requirement: Reverse Proxy Execution
For MCP requests routed to `/mcp`, the gateway SHALL execute a reverse proxy to the mcp-server service URL without modifying the request path. The incoming path (`/mcp`) matches the outgoing path exactly — no prefix stripping or rewriting is performed. Request body is read and rewound for tracing purposes (same as existing behavior), and `X-User-*` identity headers are injected via context before forwarding.

#### Scenario: Successful MCP proxy with no path rewrite
- **WHEN** a request hits `/mcp` POST through api-gateway
- **THEN** the gateway forwards it to mcp-server at the resolved URL with the original path preserved as `/mcp`, and returns the backend's response to the client

## REMOVED Requirements

### Requirement: MCP SSE reverse proxy recovers from ErrAbortHandler panics
**Reason**: Transport changed from SSE to StreamableHTTP. StreamableHTTP has no long-lived streams — each request is a discrete POST with synchronous response. No `http.ErrAbortHandler` panic can occur from client disconnect during streaming because there's no streaming connection.
