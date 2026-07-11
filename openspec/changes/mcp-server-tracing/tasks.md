## 1. Tracer initialization and shutdown in main.go

- [x] 1.1 Add imports for `common/tracing`, `"go.opentelemetry.io/otel"`, and `"context"` to mcp-server/cmd/main.go (check if already present)
- [x] 1.2 After logger initialization, call `tracerProvider, err := tracing.InitTracer(cfg.Tracing)` with same error-handling pattern as other services (warn on failure, nil when disabled)
- [x] 1.3 Add deferred shutdown: `defer func() { if tracerProvider != nil { tracing.ShutdownTracer(tracerProvider) } }()`

## 2. Wrap /mcp endpoint with otelhttp middleware

- [x] 2.1 Import `"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"` in main.go
- [x] 2.2 Replace direct `mux.Handle("/mcp", streamableHTTPServer)` with `mux.Handle("/mcp", otelhttp.NewHandler(streamableHTTPServer, "mcp-server"))` so all /mcp requests create HTTP spans
- [x] 2.3 Verify health endpoints (/health, /live, /ready, /ping, /status) remain unwrapped (no otelhttp on those handlers)

## 3. Add trace context propagation to ObjectsClient

- [x] 3.1 Import `"go.opentelemetry.io/otel"` and `"go.opentelemetry.io/otel/propagation"` in objects_client.go
- [x] 3.2 In `httpGet()`, after creating the request, add: `propagator := otel.GetTextMapPropagator(); propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))` to inject traceparent on outbound calls

## 4. Add business-level spans to tool handlers

- [x] 4.1 In object_tools.go: import `"go.opentelemetry.io/otel"`, `"go.opentelemetry.io/otel/attribute"`, `"go.opentelemetry.io/otel/codes"`
- [x] 4.2 Wrap list_objects handler body with `tracer := otel.Tracer("mcp-server"); ctx, span := tracer.Start(ctx, "tool.list_objects")` and add input params (object_type_id, limit, offset, type_key_prefix) as attributes; set status Ok/Error appropriately
- [x] 4.3 Wrap get_object handler body with `tracer.Start(ctx, "tool.get_object")` and add public_id attribute; set status Ok/Error appropriately

## 5. Add business-level spans to type tools

- [x] 5.1 In type_tools.go: import `"go.opentelemetry.io/otel"`, `"go.opentelemetry.io/otel/attribute"`, `"go.opentelemetry.io/otel/codes"`
- [x] 5.2 Wrap list_object_types handler body with `tracer.Start(ctx, "tool.list_object_types")` and add type_key_prefix, parent_type_id attributes; set status appropriately
- [x] 5.3 Wrap get_object_type handler body with `tracer.Start(ctx, "tool.get_object_type")` and add id/type_key attribute; set status appropriately

## 6. Add business-level spans to resource and prompt handlers

- [x] 6.1 In hierarchy.go: import `"go.opentelemetry.io/otel"`, `"go.opentelemetry.io/otel/attribute"`, `"go.opentelemetry.io/otel/codes"`; wrap handler body with `tracer.Start(ctx, "resource.hierarchy")` and set status appropriately
- [x] 6.2 In browse_schema.go: import otel packages; wrap handler body with `tracer.Start(ctx, "prompt.browse_schema")` and set status appropriately
- [x] 6.3 In get_object_info.go: import otel packages; wrap handler body with `tracer.Start(ctx, "prompt.get_object_info")` and set status appropriately

## 7. Build verification

- [x] 7.1 Run `go build ./services/mcp-server/...` — must compile without errors
- [x] 7.2 Verify no unused imports remain (check if otelhttp, propagation are actually used)
- [x] 7.3 Verify config.yaml tracing block is already present and correctly configured

## 8. Test verification

- [x] 8.1 Run `go test ./services/mcp-server/...` — all existing tests must pass
- [x] 8.2 Verify handler tests that mock ObjectsClient still work (context propagation doesn't break mocked calls)
