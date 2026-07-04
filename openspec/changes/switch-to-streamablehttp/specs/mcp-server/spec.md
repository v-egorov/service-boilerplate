## MODIFIED Requirements

### Requirement: MCP server uses StreamableHTTP transport via api-gateway
The mcp-server MUST expose a StreamableHTTP endpoint at `/mcp` that MCP clients connect to. The api-gateway MUST route incoming POST requests from the path prefix `/mcp/*` to the mcp-server service using standard HTTP reverse proxying without path rewriting or body manipulation. Sessions are stateful — the first POST establishes session context, and subsequent requests echo back `MCP-Session-ID` in the header for session reuse.

#### Scenario: Client connects via StreamableHTTP through gateway
- **WHEN** an MCP client sends POST /mcp with Content-Type: application/json through api-gateway
- **THEN** api-gateway forwards the request to mcp-server without path rewriting, and mcp-server responds with HTTP 200 and a JSON-RPC result containing the server-assigned session ID in the `MCP-Session-ID` response header

#### Scenario: Subsequent tool calls reuse the same session
- **WHEN** an MCP client sends POST /mcp with header `MCP-Session-ID: <session-id>` (echoed from initialize)
- **THEN** mcp-server recognizes the session, reuses existing context, and returns the tool result in a single synchronous response

#### Scenario: Tool call reaches objects-service via gateway-mcp pipeline
- **WHEN** a client POSTs a tool_call request through api-gateway to `/mcp` with `MCP-Session-ID` header
- **THEN** the request is forwarded to mcp-server which executes the tool and returns the result synchronously, with each hop (gateway → mcp-server) preserving headers and body
