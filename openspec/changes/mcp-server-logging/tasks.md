## 1. Slog → logrus bridge implementation

- [ ] 1.1 In `services/mcp-server/cmd/main.go`, create a `logrusSlogHandler` struct that implements `slog.Handler` — fields: `logger *logrus.Entry`, `group string`, `attrs []slog.Attr`
- [ ] 1.2 Implement `Enabled(ctx, level)` → return true for all levels (we filter by logrus config)
- [ ] 1.3 Implement `Handle(ctx, record)` — map slog level to logrus level (debug→Debug, info→Info, warn→Warn, error→Error), convert `record.Attributes()` into a `logrus.Fields` map keyed by attribute name/value, call appropriate logrus method with `record.Message`
- [ ] 1.4 Implement `WithAttrs(attrs)` — return new handler with merged attrs slice (append to existing)
- [ ] 1.5 Implement `WithGroup(name)` — append group name to group field for namespacing

## 2. Register slog bridge on StreamableHTTPServer

- [ ] 2.1 Create a `slog.New` instance using the logrusSlogHandler: `slog.New(&logrusSlogHandler{logger: logger.Logger.WithField("component", "mcp-transport")})`
- [ ] 2.2 Pass it to StreamableHTTPServer constructor via `server.WithStreamableHTTPLogger(...)`: wrap existing call with the option added

## 3. Operation-level logging hooks

- [ ] 3.1 Define a context key type for storing start time: `type ctxKey struct{}` — use unexported type to avoid collisions
- [ ] 3.2 Implement `onBeforeAny` hook function: capture `time.Now()` on context, extract method name and request ID into logrus fields, perform type assertion on message to extract tool/resource/prompt identifier and params (handle nil/uncastable gracefully)
- [ ] 3.3 Implement `onSuccess` hook function: retrieve start time from context via `ctx.Value(ctxKey{})`, compute duration_ms, emit info-level log with op name, request ID, status="success", duration_ms, and any result summary fields
- [ ] 3.4 Implement `onError` hook function: retrieve start time from context, compute duration_ms, emit error-level log with op name, request ID, error message, and duration_ms (handle nil ctx value for dispatcher-rejected requests)

## 4. Register hooks on MCPServer

- [ ] 4.1 In `initMCPServer()`, after creating the MCPServer, call `mcpServer.GetHooks().AddOnBeforeAny(onBeforeAny)`
- [ ] 4.2 Call `mcpServer.GetHooks().AddOnSuccess(onSuccess)`
- [ ] 4.3 Call `mcpServer.GetHooks().AddOnError(onError)`

## 5. Build verification

- [ ] 5.1 Run `go build ./services/mcp-server/...` — must compile without errors
- [ ] 5.2 Verify no unused imports (slog, time, context)
- [ ] 5.3 Verify slog is already in go.mod as stdlib (Go 1.21+)

## 6. Runtime verification

- [ ] 6.1 Trigger `make dev-detached` to rebuild and restart mcp-server container
- [ ] 6.2 Send a test request: `curl -s http://localhost:8095/mcp -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"initialize",...}'`
- [ ] 6.3 Verify logs appear in `./docker/volumes/mcp-server/logs/mcp-server.log` with fields: timestamp, level (info), service ("mcp-server"), op name, request ID
