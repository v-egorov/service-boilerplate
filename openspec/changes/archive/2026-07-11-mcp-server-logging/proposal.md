## Why

The MCP server currently produces zero operational logs during normal request handling — only startup messages appear. This makes debugging, monitoring, and troubleshooting impossible. Unlike other services (auth-service, objects-service) which log every request with method, path, status, trace ID, and duration, the mcp-server is effectively silent after launch.

## What Changes

- Add a `slog.Handler` bridge that routes mcp-go's internal transport logging to logrus (the project's unified logging system), capturing session lifecycle, SSE events, heartbeats, and connection errors
- Register MCP hooks (`OnBeforeAny`, `OnSuccess`, `OnError`) on the MCPServer to emit structured operation-level logs for every tool call, resource read, prompt invocation, ping, and initialization
- Extract and log key fields from each operation: method name (e.g., `tools/call`), tool/resource/prompt identifier, input parameters, response status, error details, and request duration

## Capabilities

### New Capabilities
- `mcp-server-logging`: Structured, operation-level logging for all MCP server requests — tools, resources, prompts, pings, and session lifecycle events. Bridges mcp-go's internal slog transport logging to logrus and registers hooks on the MCPServer for per-operation visibility.

### Modified Capabilities
(none)

## Impact

- **Affected code**: `services/mcp-server/cmd/main.go` — add slog bridge struct + registration, register MCP hooks on MCPServer
- **New dependency**: none (slog is Go stdlib since 1.21; mcp-go v0.55.1 already includes hook infrastructure)
- **No behavioral changes**: logging is additive only; request handling and response format unchanged
- **No API contract changes**: MCP protocol, tool/resource/prompt signatures unmodified
