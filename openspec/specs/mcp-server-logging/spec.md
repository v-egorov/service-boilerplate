# Specification: mcp-server-logging

## Purpose
Define structured logging requirements for the MCP server, including transport-level slog→logrus bridge and operation-level hooks that provide full visibility into every inbound request (tools, resources, prompts).

## Requirements
### Requirement: MCP server bridges mcp-go's internal slog transport logging to logrus

The mcp-server MUST implement a custom `slog.Handler` that translates all slog records emitted by mcp-go into structured logrus entries. The bridge MUST be registered via `WithStreamableHTTPLogger()` on the StreamableHTTPServer so that mcp-go's internal events — session lifecycle, SSE connection management, heartbeats, transport errors — are emitted through the project's unified logrus logging system with JSON format and file output.

#### Scenario: Session registration is logged
- **WHEN** a client sends POST /mcp establishing a new session
- **THEN** mcp-go's internal slog emits an info-level message that the bridge routes to logrus as `{"level":"info","service":"mcp-server","msg":"session registered",...}`

#### Scenario: Transport errors are logged at error level
- **WHEN** an SSE write fails or a connection is unexpectedly closed
- **THEN** mcp-go's internal slog emits an error-level message that the bridge routes to logrus as `{"level":"error","service":"mcp-server","msg":"Failed to write...","err":...}`

#### Scenario: Slog output matches logrus JSON format conventions
- **WHEN** a log entry is emitted via the bridge
- **THEN** it includes standard fields: timestamp, level, service name ("mcp-server"), message, and any additional key=value attributes from the slog record

### Requirement: MCP server emits operation-level logs for all requests via hooks

The MCPServer MUST register hooks (`OnBeforeAny`, `OnSuccess`, `OnError`) that emit structured logrus entries for every inbound MCP operation. Each operation log entry MUST include: method name (e.g., `tools/call`, `resources/read`, `prompts/get`), request ID, and — when available — the specific tool/resource/prompt identifier and input parameters.

#### Scenario: Tool call logs method, name, and parameters
- **WHEN** a client calls `tools/call` with the `list_objects` tool and parameters `{"object_type_id": 5}`
- **THEN** an info-level log entry is emitted containing `"op":"tools/call"`, `"tool":"list_objects"`, `"params":{"object_type_id":5}`, and `"status":"success"` (on success) or `"error":...` (on failure)

#### Scenario: Resource read logs method, URI, and parameters
- **WHEN** a client calls `resources/read` for the `objects-types://hierarchy` resource with arguments `{"type_key_prefix":"auth"}`
- **THEN** an info-level log entry is emitted containing `"op":"resources/read"`, `"resource_uri":"objects-types://hierarchy"`, and input parameters

#### Scenario: Prompt invocation logs method, name, and arguments
- **WHEN** a client calls `prompts/get` for the `browse_schema` prompt with argument `{"type_key_prefix":"obj"}`
- **THEN** an info-level log entry is emitted containing `"op":"prompts/get"`, `"prompt_name":"browse_schema"`, and input arguments

#### Scenario: Error operations log error details
- **WHEN** any MCP operation fails (tool handler panics, objects-service returns 500, invalid JSON-RPC)
- **THEN** the `OnError` hook emits an error-level entry with `"op":<method>`, `"error":"..."`, and request ID for traceability

#### Scenario: Ping operations are logged at debug level
- **WHEN** a client sends a ping (`ping` method) to check connectivity
- **THEN** a debug-level log entry is emitted containing `"op":"ping"` (low-frequency health signal, not noisy)

### Requirement: MCP server logs request duration for all operations

The hooks MUST measure and report the elapsed time from operation start to completion. Duration must be recorded as `duration_ms` in milliseconds on every info-level operation log entry.

#### Scenario: Successful tool call reports duration
- **WHEN** a tool handler completes successfully
- **THEN** the log entry includes `"duration_ms":<number>` reflecting total wall-clock time from OnBeforeAny to OnSuccess

#### Scenario: Failed tool call reports duration
- **WHEN** a tool handler fails with an error
- **THEN** the error-level log entry includes `"duration_ms":<number>` and `"error":"..."`
