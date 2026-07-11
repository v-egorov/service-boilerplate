## ADDED Requirements

### Requirement: MCP server initializes OpenTelemetry tracer provider on startup

The mcp-server MUST initialize an OpenTelemetry TracerProvider using the shared `common/tracing.InitTracer` function with configuration from `cfg.Tracing`. The tracer provider MUST be gracefully shut down via `common/tracing.ShutdownTracer` during server shutdown. When tracing is disabled (`tracing.enabled: false`), initialization returns nil without error and no spans are emitted.

#### Scenario: TracerProvider initialized successfully
- **WHEN** the mcp-server starts with `tracing.enabled: true` in config.yaml
- **THEN** `common/tracing.InitTracer(cfg.Tracing)` creates an OTLP HTTP tracer provider that exports spans to `cfg.Tracing.CollectorURL` and is registered as the global TracerProvider

#### Scenario: Shutdown is called during graceful server shutdown
- **WHEN** the mcp-server receives SIGINT or SIGTERM
- **THEN** `common/tracing.ShutdownTracer(tracerProvider)` is called before the process exits, ensuring all in-flight spans are flushed

#### Scenario: Tracing disabled does not cause errors
- **WHEN** the mcp-server starts with `tracing.enabled: false` in config.yaml
- **THEN** InitTracer returns nil for the provider and no error; ShutdownTracer handles nil gracefully

### Requirement: MCP server creates HTTP-level spans for all requests to /mcp

The `/mcp` endpoint handler MUST be wrapped so that every incoming HTTP request creates a span via OpenTelemetry's `otelhttp` middleware. The span MUST extract trace context from incoming W3C TraceContext headers (`traceparent`, `baggage`) and create a child span of the upstream service (api-gateway) when context is present. The span name SHOULD be `"mcp-server"` with HTTP method, route, and status code as attributes.

#### Scenario: Incoming request creates an HTTP span
- **WHEN** an MCP client sends POST /mcp to the mcp-server
- **THEN** a span named `"mcp-server"` is created with `http.method`, `http.route`, `http.status_code` attributes, and spans are recorded for both success and error responses

#### Scenario: Trace context is extracted from gateway headers
- **WHEN** the api-gateway proxies an MCP request with a `traceparent` header
- **THEN** the HTTP span in mcp-server becomes a child of the gateway's proxy span, continuing the distributed trace chain

### Requirement: MCP server propagates trace context on outbound calls to objects-service

When the mcp-server makes HTTP requests to objects-service (from tool handlers, resource handlers, and prompt handlers), it MUST inject the active W3C TraceContext into the outbound request headers using `otel.GetTextMapPropagator().Inject()`. This ensures that objects-service receives the correct trace parent and can create child spans.

#### Scenario: Tool call propagates trace context to objects-service
- **WHEN** a tool handler (e.g., list_objects) calls ObjectsClient.httpGet() with an active trace context
- **THEN** the outbound HTTP request to objects-service includes `traceparent` header in W3C TraceContext format

#### Scenario: Resource read propagates trace context to objects-service
- **WHEN** a resource handler (e.g., hierarchy) calls ObjectsClient methods with an active trace context
- **THEN** the outbound HTTP request carries the trace parent from the inbound HTTP span

#### Scenario: Prompt invocation propagates trace context to objects-service
- **WHEN** a prompt handler (e.g., browse_schema) calls ObjectsClient methods with an active trace context
- **THEN** the outbound HTTP request carries the trace parent from the inbound HTTP span

### Requirement: MCP server creates business-level spans for tool execution

Each MCP tool handler MUST create a manual span via `otel.Tracer("mcp-server").Start()` that wraps the entire tool execution — argument parsing, ObjectsClient call, and response rendering. The span name SHOULD follow the pattern `"tool.<tool_name>"` (e.g., `"tool.list_objects"`, `"tool.get_object"`). Input parameters MUST be recorded as span attributes for filtering and debugging.

#### Scenario: list_objects creates a tool span with input attributes
- **WHEN** the list_objects tool handler is invoked
- **THEN** a span named `"tool.list_objects"` is created with `tool.name` = "list_objects", `object_type_id`, `limit`, `offset`, and `type_key_prefix` (when provided) as string attributes

#### Scenario: get_object creates a tool span with input attributes
- **WHEN** the get_object tool handler is invoked
- **THEN** a span named `"tool.get_object"` is created with `tool.name` = "get_object" and `public_id` as string attributes

#### Scenario: Tool error sets span status to Error
- **WHEN** a tool handler encounters an error (e.g., objects-service returns 500)
- **THEN** the tool span calls `span.RecordError(err)` and `span.SetStatus(codes.Error, ...)` before returning the MCP error result

#### Scenario: Tool success sets span status to Ok
- **WHEN** a tool handler completes successfully
- **THEN** the tool span calls `span.SetStatus(codes.Ok, "")` after rendering the response

### Requirement: All MCP handler types create explicit business-level spans

Every code path in the mcp-server that handles an inbound request — tools, resources, and prompts — MUST create an explicit manual span via `otel.Tracer("mcp-server").Start()`. This ensures complete visibility into every user-facing operation and prevents trace gaps when errors occur. Span naming conventions:
- **Tools:** `"tool.<name>"` (e.g., `"tool.list_objects"`)
- **Resources:** `"resource.<uri_basename>"` (e.g., `"resource.hierarchy"` for the `objects-types://hierarchy` URI)
- **Prompts:** `"prompt.<name>"` (e.g., `"prompt.browse_schema"`)

Each span MUST record input parameters as attributes and set status Ok/Error appropriately. All spans propagate trace context to outbound ObjectsClient calls via the propagated context in their `ctx`.

#### Scenario: Resource read creates a manual span with visibility
- **WHEN** an agent reads the `objects-types://hierarchy` resource URI
- **THEN** a span named `"resource.hierarchy"` is created, traceparent is injected onto the outbound request to objects-service, and any error during fetch is recorded on the span with `span.RecordError(err)` and `span.SetStatus(codes.Error, ...)`

#### Scenario: Prompt invocation creates a manual span with visibility
- **WHEN** an agent invokes the `browse_schema` prompt template
- **THEN** a span named `"prompt.browse_schema"` is created, traceparent is injected onto outbound ObjectsClient calls, and errors during type fetching are recorded on the span

#### Scenario: Error in resource handler does not break the trace chain
- **WHEN** objects-service returns an error (500) while serving a resource request
- **THEN** the `"resource.hierarchy"` span captures the error with `span.RecordError(err)` and status `codes.Error`, allowing Grafana/Jaeger to display the failure rather than leaving a gap in the trace
