## ADDED Requirements

### Requirement: MCP SSE reverse proxy recovers from ErrAbortHandler panics
The api-gateway's MCP SSE reverse proxy handler SHALL recover from `http.ErrAbortHandler` panics raised by `httputil.ReverseProxy` without triggering gin's recovery middleware. When a client disconnects or the SSE stream ends naturally, the reverse proxy SHALL terminate cleanly without logging a panic or returning an HTTP error response.

#### Scenario: SSE stream disconnects without gateway panic
- **WHEN** an MCP client disconnects from the SSE stream (`GET /mcp/sse`) through the api-gateway
- **THEN** the reverse proxy terminates normally, no panic is logged, and no HTTP 500 response is written

#### Scenario: Non-ErrAbortHandler panics still trigger gin recovery
- **WHEN** the SSE reverse proxy encounters a real panic (not `http.ErrAbortHandler`)
- **THEN** the panic propagates to gin's recovery middleware and is handled as a server error

#### Scenario: SSE streaming continues for active connections
- **WHEN** an MCP client maintains an active SSE connection through the gateway and sends tool call messages via POST `/mcp/message`
- **THEN** the SSE stream remains open, responses flow back through the reverse proxy, and the connection persists until either the client disconnects or the session expires
