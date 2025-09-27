# Tools & Libraries

This document details the technical stack and dependencies used for distributed tracing implementation.

## 📚 OpenTelemetry Go Libraries

### Core Dependencies

```go
// Primary OpenTelemetry packages
go.opentelemetry.io/otel v1.38.0                    // Core API and types
go.opentelemetry.io/otel/sdk v1.38.0                // SDK implementation
go.opentelemetry.io/otel/trace v1.38.0              // Tracing API
go.opentelemetry.io/otel/metric v1.38.0             // Metrics API (future use)
go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.38.0           // OTLP exporter base
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.38.0  // HTTP transport
```

### Semantic Conventions

```go
go.opentelemetry.io/otel/semconv/v1.21.0            // Standard attribute names
```

### Additional Dependencies

```go
// Auto-instrumentation (included automatically)
go.opentelemetry.io/auto/sdk v1.1.0

// Protocol buffers and gRPC (for OTLP)
go.opentelemetry.io/proto/otlp v1.7.1
google.golang.org/grpc v1.75.0
google.golang.org/protobuf v1.36.8
```

## 🏗️ Infrastructure Components

### Jaeger (Trace Collector & UI)

**Docker Image:**
```yaml
jaegertracing/all-in-one:latest
```

**Ports:**
- `16686` - Web UI
- `14268` - Jaeger Thrift (HTTP)
- `14250` - Jaeger Thrift (gRPC)
- `4318` - OTLP HTTP (custom addition)

**Environment Variables:**
```yaml
COLLECTOR_OTLP_ENABLED=true          # Enable OTLP ingestion
COLLECTOR_ZIPKIN_HOST_PORT=:9411     # Zipkin compatibility
```

**Storage:** In-memory (development), configurable for production

### Docker Compose Configuration

```yaml
jaeger:
  image: jaegertracing/all-in-one:latest
  ports:
    - "16686:16686"    # Jaeger UI
    - "4318:4318"      # OTLP HTTP collector
  environment:
    - COLLECTOR_OTLP_ENABLED=true
  volumes:
    - jaeger_data:/tmp
  networks:
    - service-network
```

## 🔧 Custom Implementation

### Common Tracing Package

**Location:** `common/tracing/`

**Files:**
- `tracer.go` - Global tracer provider setup
- `middleware.go` - Gin HTTP middleware

**Key Functions:**

#### `InitTracer(cfg TracingConfig) (*trace.TracerProvider, error)`
- Creates OTLP HTTP exporter
- Sets up resource with service information
- Configures batch span processor
- Initializes global text map propagator
- Returns tracer provider for shutdown

#### `ShutdownTracer(tp *trace.TracerProvider) error`
- Gracefully shuts down tracer provider
- Flushes remaining spans
- Cleans up resources

#### `HTTPMiddleware(serviceName string) gin.HandlerFunc`
- Gin middleware for automatic HTTP request tracing
- Extracts trace context from request headers
- Creates spans with HTTP attributes
- Records errors and response status
- Injects trace context into request context

### Configuration Structure

```go
type TracingConfig struct {
    Enabled      bool    `mapstructure:"enabled"`
    ServiceName  string  `mapstructure:"service_name"`
    CollectorURL string  `mapstructure:"collector_url"`
    SamplingRate float64 `mapstructure:"sampling_rate"`
}
```

## 📊 Sampling Strategies

### Parent-Based Sampling
- **Default Strategy**: `trace.ParentBased(trace.TraceIDRatioBased(cfg.SamplingRate))`
- **Behavior**: If parent span is sampled, child spans are sampled
- **Fallback**: Uses trace ID ratio for root spans

### Sampling Rates
- **Development**: `1.0` (100% sampling)
- **Staging**: `0.1` (10% sampling)
- **Production**: `0.01` - `0.1` (1-10% sampling)

## 🌐 Propagation Mechanisms

### W3C TraceContext Propagator

**Global Setup:**
```go
otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
    propagation.TraceContext{},  // W3C traceparent header
    propagation.Baggage{},       // W3C tracestate header
))
```

**Header Format:**
```
traceparent: 00-{trace-id}-{span-id}-{flags}
tracestate: vendor-specific-data
```

### Injection Points

#### API Gateway Proxy
```go
// In reverse proxy director
ctx := c.Request.Context()
otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
```

#### Service Clients
```go
// In HTTP client requests
otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(httpReq.Header))
```

## 📈 Span Attributes

### Automatic HTTP Attributes
```go
attribute.String("http.method", c.Request.Method)
attribute.String("http.url", c.Request.URL.String())
attribute.String("http.user_agent", c.Request.UserAgent())
attribute.String("http.scheme", c.Request.URL.Scheme)
attribute.String("http.host", c.Request.Host)
attribute.Int("http.status_code", c.Writer.Status())
```

### Custom Attributes
```go
attribute.String("service.name", serviceName)
attribute.String("http.route", c.Request.URL.Path)
attribute.String("db.operation", "SELECT")
attribute.String("external.service", "payment-gateway")
```

## 🔄 Data Flow

### Span Processing Pipeline

1. **Span Creation**: Application code creates spans
2. **Attribute Addition**: Automatic and custom attributes added
3. **Event Recording**: Errors and custom events logged
4. **Status Setting**: Success/error status recorded
5. **Span Ending**: `defer span.End()` called
6. **Batch Processing**: SDK batches spans
7. **Export**: OTLP HTTP exporter sends to Jaeger
8. **Storage**: Jaeger stores in configured backend
9. **Query**: Jaeger UI queries for visualization

### Batch Configuration

**Default Settings:**
- Batch size: Configurable (default: efficient for HTTP)
- Timeout: 30 seconds
- Retry: Exponential backoff
- Queue size: Memory-bounded

## 🐳 Docker Integration

### Network Configuration
```yaml
networks:
  service-network:
    driver: bridge
    labels:
      - "com.service-boilerplate.network=backend"
```

### Volume Management
```yaml
volumes:
  jaeger_data:
    driver: local
```

### Health Checks
```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:4318/v1/traces"]
  interval: 30s
  timeout: 10s
  retries: 3
```

## 📋 Version Compatibility

### Go Version Requirements
- **Minimum**: Go 1.19+
- **Recommended**: Go 1.21+
- **Current**: Go 1.23.0

### OpenTelemetry Compatibility
- **API**: v1.38.0
- **SDK**: v1.38.0
- **Protocol**: OTLP v1.7.1

### Jaeger Compatibility
- **Version**: All-in-one latest
- **Protocol**: OTLP HTTP
- **Storage**: In-memory (dev), Elasticsearch/Cassandra (prod)

## 🚀 Future Extensions

### Potential Additions
- **Metrics**: `go.opentelemetry.io/otel/metric`
- **Logs**: Structured logging correlation
- **gRPC**: Protocol-specific propagation
- **Database**: Auto-instrumentation for PostgreSQL
- **External Services**: HTTP client auto-instrumentation

### Production Considerations
- **Persistent Storage**: Jaeger with Elasticsearch
- **Load Balancing**: Multiple Jaeger collectors
- **Security**: Authentication for OTLP endpoints
- **Monitoring**: Jaeger metrics and health checks

---

*Next: [Implementation Details](implementation.md) | [Configuration](configuration.md)*</content>
</xai:function_call">docs/tracing/tools.md