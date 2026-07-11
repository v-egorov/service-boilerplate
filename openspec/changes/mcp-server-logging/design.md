## Context

The MCP server uses `github.com/mark3labs/mcp-go` v0.55.1 with a raw `http.ServeMux` (not Gin). It has no logging middleware on the HTTP layer and no hooks configured on the MCPServer. As a result, after startup messages, the server produces zero log output regardless of how many requests it handles — tool calls, resource reads, prompt invocations, errors, everything is invisible in logs.

Other services (auth-service, objects-service) use `common/logging.ServiceRequestLogger` with Gin middleware to emit structured per-request logs including method, path, status, duration, trace ID, and user identity. The mcp-server has no equivalent observability layer.

## Goals / Non-Goals

**Goals:**
- Emit operation-level structured logs for every MCP request (tools, resources, prompts) with method name, identifier, parameters, status, errors, and duration
- Route mcp-go's internal transport logging (session lifecycle, SSE events, heartbeats) to logrus via a slog bridge so transport issues are visible in logs
- Maintain the project's unified logging conventions: JSON format, service name field, file output via lumberjack rotation

**Non-Goals:**
- Adding HTTP middleware for path-level request logging (the mcp-server has only one business endpoint `/mcp` — operation-level hooks provide better granularity)
- Modifying health endpoint logs (`/health`, `/live`, `/ready`) — these are infrastructure probes, not MCP operations
- Changing the JSON-RPC response format or API contract

## Decisions

### Decision 1: Bridge slog to logrus (not replace with pure logrus hooks)

**Choice:** Implement a `slog.Handler` that delegates to `*logrus.Entry`, register it via `WithStreamableHTTPLogger()`. Also register MCP operation hooks on the MCPServer for business-level visibility.

```
┌───────────────────────────────────────────────┐
│  mcp-go internal (slog)                       │
│       ↓                                       │
│  slog.Handler bridge                          │  ← routes transport logs to logrus
│       ↓                                       │
│  logrus.Info/Warn/Error with JSON formatter   │
│                                               │
│  ┌───────────────────────────────────────┐    │
│  │  MCPServer hooks (OnBeforeAny, ...)   │    │  ← business-level ops logging
│  │       ↓                               │    │     directly to logrus
│  │  logger.WithFields(...).Info(...)     │    │
│  └───────────────────────────────────────┘    │
└───────────────────────────────────────────────┘
```

**Rationale:** mcp-go uses slog internally for transport events (session registration, SSE writes, heartbeat sends, panic recovery in notification forwarders). These are invisible to MCP hooks and would be lost without the bridge. The bridge is ~20 lines — a one-time plumbing job, not ongoing maintenance.

**Alternatives considered:**
- *Hooks-only approach:* Misses all transport-level events (session lifecycle, SSE errors, heartbeats) that operators need for debugging connection issues.
- *Pure slog approach:* Would diverge from the project's unified logrus logging convention, making it harder to correlate mcp-server logs with other services in Grafana/ Loki.

### Decision 2: Hook registration via MCPServer.GetHooks().Add*()

**Choice:** After creating `MCPServer`, call `GetHooks()` and register `OnBeforeAny`, `OnSuccess`, and `OnError` hooks using the fluent `Add*` methods (not `WithHooks()` option).

```go
mcpServer := server.NewMCPServer(name, version)

// Register operation-level hooks
hooks := mcpServer.GetHooks()
hooks.AddOnBeforeAny(onBeforeAny)
hooks.AddOnSuccess(onSuccess)
hooks.AddOnError(onError)
```

**Rationale:** `WithHooks()` replaces the entire Hooks instance (risking loss of library-internal hooks). `GetHooks().Add*()` appends to the existing slice, preserving any defaults. This is the safer incremental approach.

### Decision 3: Field extraction from generic message payload

**Choice:** Use type assertions on the `message interface{}` parameter in hooks to extract operation-specific fields. Map known method names to expected types:

```go
func onBeforeAny(ctx context.Context, id any, method mcp.MCPMethod, message any) {
    fields := logrus.Fields{"op": string(method), "request_id": fmt.Sprintf("%v", id)}

    switch method {
    case mcp.MethodCallTool:
        if p, ok := message.(*mcp.CallToolParams); ok {
            fields["tool"] = p.Name
            fields["params"] = p.Params // map[string]interface{}
        }
    case mcp.MethodReadResource:
        if p, ok := message.(*mcp.ReadResourceParams); ok {
            fields["resource_uri"] = p.ResourceURI
            fields["arguments"] = p.Arguments
        }
    case mcp.MethodGetPrompt:
        if p, ok := message.(*mcp.GetPromptParams); ok {
            fields["prompt_name"] = p.Name
            fields["arguments"] = p.Arguments
        }
    }

    logger.WithFields(fields).Info("mcp_operation_start")
}
```

**Rationale:** mcp-go v0.55.1 defines typed params structs for each method (`CallToolParams`, `ReadResourceParams`, `GetPromptParams`). Type assertions are safe — if the cast fails, we still log the operation name and request ID (graceful degradation).

### Decision 4: Duration measurement via time.Since in hooks

**Choice:** Capture start time in `OnBeforeAny` using `time.Now()`, store it on a simple context value or closure variable, then report duration in both `OnSuccess` and `OnError`.

Since mcp-go hooks don't share state between Before/After callbacks, we use a lightweight per-request context key:

```go
type ctxKey struct{}

func onBeforeAny(ctx context.Context, id any, method mcp.MCPMethod, message any) {
    ctx = context.WithValue(ctx, ctxKey{}, time.Now())
}

func onSuccess(ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
    start, _ := ctx.Value(ctxKey{}).(time.Time)
    duration := time.Since(start).Milliseconds()
    // ... log with duration_ms field
}
```

**Rationale:** Context is already passed through all hooks and carries trace context. Using it for timing avoids additional global state. The key collision risk is nil (unexported struct type as key).

## Risks / Trade-offs

| Risk | Impact | Mitigation |
|------|--------|------------|
| Slog bridge adds overhead to every transport log | Low — mcp-go transport logs are infrequent (session events, not per-request) | Bridge only creates a logrus entry; no JSON marshaling beyond what logrus already does |
| Hook registration order affects behavior | Medium — OnError might fire without OnBeforeAny for dispatcher-level rejections | Check: documentation says "OnError may fire without prior OnBeforeAny when the dispatcher rejects invalid requests" — handle nil ctx value gracefully |
| Type assertion failures on message payload | Low — if cast fails, still log op name + request ID (graceful degradation) | Add fallback logging for uncastable messages; monitor first deployment logs for missing fields |
| Two logging paths produce duplicate entries | Medium — a successful operation produces both an OnBeforeAny + OnSuccess pair AND potentially a slog transport entry | These are different log levels and scopes: hooks = business-level (info), bridge = transport-level (debug/info). No duplication of the same event. |

## Migration Plan

1. Add `slogHandler` struct and its three methods (`Enabled`, `Handle`, `WithAttrs`, `WithGroup`) to `cmd/main.go`
2. Register the handler via `server.WithStreamableHTTPLogger(slog.New(logrusSlogHandler{...}))` on the StreamableHTTPServer constructor
3. Add hook functions (`onBeforeAny`, `onSuccess`, `onError`) and register them on MCPServer.GetHooks()
4. Build, deploy, verify logs appear in `./docker/volumes/mcp-server/logs/mcp-server.log` with expected fields

**Rollback:** Remove the three code additions (bridge struct, handler registration, hook registration). No config changes needed — if the binary doesn't have logging code, it simply produces no extra output (same as current behavior).

## Open Questions

1. **Log level for successful operations:** Should `OnSuccess` emit at info or debug level? Info is more visible but increases log volume for high-frequency tool calls. Recommendation: info-level with short messages (tool name + duration), full details only on error.
2. **Parameter redaction:** Should we redact any fields from input parameters that might contain sensitive data (e.g., user IDs, emails)? Current analysis suggests no — the mcp-server is an internal service and params are tool/resource identifiers (type IDs, public UUIDs, URI strings), not PII.
