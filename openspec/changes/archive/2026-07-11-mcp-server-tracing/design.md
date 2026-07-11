## Context

The mcp-server is a Go service that exposes MCP (Model Context Protocol) tools, resources, and prompts via StreamableHTTP transport. It runs on `http.ServeMux` (not Gin) and proxies all data requests to objects-service. Currently it has zero tracing — no tracer initialization, no HTTP middleware, no context propagation, and no business-level spans. All other services in the project already use OpenTelemetry with Jaeger via OTLP HTTP exporter at port 4318.

The common infrastructure (`common/tracing/`) provides:
- `InitTracer(cfg)` — creates TracerProvider with OTLP HTTP exporter to a configurable collector URL
- `ShutdownTracer(tp)` — graceful shutdown with span flushing
- `HTTPMiddleware(serviceName string)` — Gin middleware for automatic HTTP span creation (not usable here since mcp-server uses raw ServeMux)

The common database layer (`common/database/tracing.go`) provides DB operation tracing wrappers used by auth-service and user-service repositories. The mcp-server has no direct DB access, so these are not needed.

## Goals / Non-Goals

**Goals:**
- Initialize OpenTelemetry tracer in mcp-server with the same OTLP HTTP exporter pattern as other services
- Create HTTP-level spans for all `/mcp` requests using `otelhttp.NewHandler` (since Gin middleware is unavailable)
- Propagate trace context on outbound HTTP calls to objects-service via W3C TraceContext headers
- Create manual business-level spans around MCP tool handlers with input parameters as attributes
- Ensure zero behavioral change — tracing is transparent to clients and tools

**Non-Goals:**
- Adding DB operation tracing (mcp-server has no direct database access)
- Modifying the common/tracing package or its API surface
- Adding metrics, logging enrichment, or other observability signals
- Changing the mcp-go library version or transport mechanism
- Tracing health check endpoints (/health, /ping, etc.) — they are trivial and low-value

## Decisions

### Decision 1: Use `otelhttp.NewHandler` for HTTP-level tracing (not common/tracing.HTTPMiddleware)

**Choice:** Wrap only the `/mcp` endpoint with `otelhttp.NewHandler(streamableHTTPServer, "mcp-server")`. Health endpoints remain unwrapped.

**Rationale:** The existing `common/tracing.HTTPMiddleware` returns a `gin.HandlerFunc`, which is incompatible with mcp-server's raw `http.ServeMux` architecture. The `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` package already exists in go.mod (as indirect dependency) and provides the standard-library equivalent: automatic span creation, W3C context extraction, HTTP attribute recording. Wrapping only `/mcp` avoids noisy spans for trivial health endpoints while capturing all meaningful request traffic.

**Alternatives considered:**
- Write a custom `http.Handler` middleware — unnecessary duplication of otelhttp's work
- Wrap the entire ServeMux — would create low-value spans for /health, /ping, etc.

### Decision 2: Inject trace context in ObjectsClient.httpGet() (not per-call site)

**Choice:** Add propagator injection inside `ObjectsClient.httpGet()` so all callers automatically propagate context without modification to each call site.

**Rationale:** All outbound calls from mcp-server go through a single method (`httpGet`). Adding injection there means:
- One change location, no risk of missing a call site
- The `ctx` parameter already flows from tool handlers (which have the HTTP span context) → ObjectsClient → httpGet
- Health checks that don't carry trace context will simply inject an empty/no-op propagator — harmless

**Alternatives considered:**
- Inject at each tool handler before calling ObjectsClient — more code changes, higher risk of inconsistency
- Create a dedicated tracing client wrapper — overkill for two callers (tools + resources/prompts)

### Decision 3: Manual spans on ALL handler types (tools, resources, prompts)

**Choice:** Every MCP handler — tools, resources, and prompts — creates an explicit manual span via `otel.Tracer("mcp-server").Start()`. Resource spans named `"resource.<uri>"` (e.g., `"resource.hierarchy"`). Prompt spans named `"prompt.<name>"` (e.g., `"prompt.browse_schema"`).

**Rationale:** While resource and prompt handlers are thin wrappers around single ObjectsClient calls, they still represent distinct user-facing operations that deserve visibility in Grafana/Jaeger. A trace broken at the MCP hop means any error from a resource or prompt handler is invisible — no span name, no error recorded, no correlation with upstream gateway spans. Adding manual spans gives complete coverage: every inbound request path produces a visible span tree with meaningful names.

**Cost:** Minimal — 3 extra `tracer.Start()`/`span.End()` pairs across 2 files (hierarchy.go + prompt files). Resource/prompt calls are low-frequency compared to tool calls, so noise is negligible. The visibility gain (error reporting, complete trace chains) far outweighs the cost.

**Alternatives considered:**
- Manual spans only on tools — risks invisible errors from resource/prompts that break the trace chain; Grafana UI shows gaps in the MCP server's behavior

### Decision 4: Span naming convention follows existing patterns

**Choice:** Tool spans use `"tool.<name>"` pattern (e.g., `tool.list_objects`). HTTP span uses the service name prefix from otelhttp.

**Rationale:** Consistent with auth-service's pattern (`auth.login`, `auth.logout`) and objects-service's approach. The `tool.` prefix makes it easy to filter spans in Jaeger by operation type vs infrastructure (HTTP, DB).

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| otelhttp adds overhead to every request | Overhead is minimal (<1ms per span); sampling rate already configured at 1.0 for dev |
| mcp-go's context lifecycle may not carry trace info | Verified: mcp-go passes the incoming request context through to tool handlers — we confirmed this in code inspection where `ctx context.Context` is received from mcp-go and flows to ObjectsClient |
| Tracing adds dependencies (OTLP exporter, otelhttp) | Dependencies already present in go.mod as indirect; will become direct but no new major versions |
| Business spans with input params could leak PII | Tool parameters are object IDs, type keys, pagination values — no user emails, tokens, or sensitive data. Safe to include as attributes. |

## Migration Plan

No migration needed. The change is additive:
1. Deploy the updated mcp-server alongside existing services
2. Verify spans appear in Jaeger UI at `http://localhost:16686` under service name "mcp-server"
3. Confirm trace chains from gateway → mcp-server → objects-service are continuous

## Open Questions

None — all decisions resolved by following the established patterns from auth-service and api-gateway.
