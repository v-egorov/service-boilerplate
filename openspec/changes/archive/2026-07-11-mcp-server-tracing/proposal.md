## Why

The MCP server is a complete blind spot in the project's distributed tracing. While all other services (auth-service, user-service, objects-service, api-gateway) emit spans to Jaeger via OpenTelemetry, the mcp-server creates zero spans and does not propagate trace context on its outbound HTTP calls to objects-service. This means traces that start at the gateway and flow through to backend services are broken at the MCP hop — observability tools show disconnected, orphaned traces with no visibility into which tool was called, what parameters were used, or how long business logic took versus network I/O.

## What Changes

- Add OpenTelemetry tracer initialization and graceful shutdown to `mcp-server/cmd/main.go`
- Wrap the `/mcp` endpoint handler with `otelhttp.NewHandler` for automatic HTTP-level span creation and W3C trace context extraction (replacing the need for a Gin-specific middleware since mcp-server uses raw `http.ServeMux`)
- Add trace context propagation (`propagator.Inject`) to `ObjectsClient.httpGet()` so outbound calls to objects-service carry the active trace parent header
- Add manual business-level spans around each MCP tool handler (list_object_types, get_object_type, list_objects, get_object) with operation name and input parameters as span attributes
- Add manual spans around resource and prompt handlers for completeness

## Capabilities

### New Capabilities

- `mcp-server-tracing`: OpenTelemetry tracing for the mcp-server — HTTP transport layer (otelhttp middleware), business-level tool spans, and trace context propagation to objects-service

### Modified Capabilities

None. This change does not modify requirements in existing specs; it adds new observability capabilities.

## Impact

- **Affected code**: `mcp-server/cmd/main.go` (init/shutdown + handler wrapping), `services/mcp-server/internal/client/objects_client.go` (context propagation), `services/mcp-server/internal/tools/*.go` (tool-level spans), `services/mcp-server/internal/resources/hierarchy.go` (resource span), `services/mcp-server/internal/prompts/*.go` (prompt spans)
- **Dependencies**: Adds import of `common/tracing`, `go.opentelemetry.io/otel`, `go.opentelemetry.io/otel/propagation`, `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` — all already present in go.mod as indirect dependencies (will become direct)
- **API behavior change**: None. Tracing is transparent to clients; no API contract changes
- **Config**: The existing `config.yaml` tracing block is already wired (`enabled: true, collector_url: http://jaeger:4318/v1/traces`) — no config changes needed
