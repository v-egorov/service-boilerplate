# Distributed Tracing Overview & Architecture

## 🎯 What is Distributed Tracing?

Distributed tracing is a method of tracking requests as they flow through distributed systems - in our case, across multiple microservices. It provides visibility into the performance and behavior of requests as they traverse service boundaries.

## 🏗️ Architecture Overview

### Core Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │───▶│  Auth Service  │───▶│  User Service   │
│   (Entry Point) │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 ▼
                    ┌─────────────────────┐
                    │   Jaeger Collector  │
                    │   (OTLP HTTP: 4318) │
                    └─────────────────────┘
                                 │
                                 ▼
                    ┌─────────────────────┐
                    │     Jaeger UI       │
                    │   (Web UI: 16686)   │
                    └─────────────────────┘
```

### Data Flow

1. **Request Entry**: API Gateway receives HTTP request
2. **Span Creation**: Gateway middleware creates initial span
3. **Header Injection**: Trace context injected into proxied request
4. **Service Processing**: Each service creates child spans
5. **Client Propagation**: Service clients inject trace headers
6. **Data Export**: All spans sent to Jaeger via OTLP HTTP
7. **Visualization**: Traces viewable in Jaeger UI

## 🔄 Trace Propagation

### W3C TraceContext Standard

We use the industry-standard W3C TraceContext specification for trace propagation:

- **traceparent**: Contains trace ID, span ID, and flags
  - Format: `00-{trace-id}-{span-id}-{flags}`
  - Example: `00-76429b34e7b89ca17f590cf6ae8d3649-c54dcdb426a9a4fb-01`

- **tracestate**: Vendor-specific trace information (optional)

### Propagation Points

1. **API Gateway → Services**: HTTP headers injected in reverse proxy
2. **Service → Service**: HTTP client injects headers in outgoing requests
3. **Async Operations**: Context passed through goroutines and channels

## 📊 Key Concepts

### Spans
- **Definition**: A unit of work in a trace (e.g., HTTP request, database query)
- **Hierarchy**: Parent-child relationships show request flow
- **Attributes**: Key-value metadata (HTTP method, status codes, custom data)
- **Events**: Timestamped logs within spans

### Traces
- **Definition**: Complete journey of a request through the system
- **Trace ID**: Unique identifier linking all spans in a request
- **Root Span**: Initial span (usually at API Gateway)
- **Child Spans**: Subsequent operations in the request flow

### Sampling
- **Purpose**: Control performance impact by not tracing all requests
- **Types**: Parent-based, trace ID ratio-based
- **Configuration**: 100% in development, 1-10% in production

## 🛠️ Technology Stack

### OpenTelemetry
- **Specification**: Vendor-neutral observability framework
- **Language**: Go SDK for instrumentation
- **Protocol**: OTLP (OpenTelemetry Protocol) for data export

### Jaeger
- **Role**: Trace collector, storage, and visualization
- **Backend**: In-memory for development, configurable for production
- **UI**: Web interface for trace exploration and analysis

### Integration Points

| Component | Responsibility |
|-----------|----------------|
| **API Gateway** | Request entry, initial span creation, header injection |
| **Services** | HTTP middleware, business logic spans, client propagation |
| **Common Package** | Tracer setup, middleware, propagation utilities |
| **Configuration** | Per-service tracing settings, sampling rates |
| **Docker** | Jaeger service orchestration, network setup |

## 🎯 Benefits

### Observability
- **Request Tracking**: Follow requests across service boundaries
- **Performance Monitoring**: Identify bottlenecks and slow operations
- **Error Correlation**: Link errors across distributed components

### Debugging
- **Root Cause Analysis**: Trace issues to specific services/operations
- **Dependency Mapping**: Understand service communication patterns
- **Historical Analysis**: Review past request patterns

### Operations
- **SLA Monitoring**: Track response times and success rates
- **Capacity Planning**: Understand system load patterns
- **Incident Response**: Quick identification of affected components

## 🔧 Development Workflow

### For New Services
1. Use `create-service.sh` - tracing automatically included
2. Configure service-specific settings in `config.yaml`
3. HTTP endpoints automatically instrumented
4. Add custom spans for complex business logic

### For Existing Code
1. Import tracing package
2. Add middleware to router
3. Configure tracer provider
4. Instrument custom operations as needed

### For Debugging
1. Enable 100% sampling in development
2. Access Jaeger UI at `http://localhost:16686`
3. Search traces by service, operation, or tags
4. Analyze span durations and error patterns

## 📈 Performance Considerations

### Overhead
- **Minimal Impact**: <5% performance overhead with proper sampling
- **Memory**: Small per-request memory allocation for spans
- **Network**: OTLP batches reduce network traffic

### Optimization
- **Sampling**: Reduce in production (1-10% typical)
- **Batching**: OTLP exporter batches spans automatically
- **Resource Limits**: Configure appropriate span/trace limits

## 🚀 Getting Started

1. **Read [Tools & Libraries](tools.md)** for technical details
2. **Check [Configuration](configuration.md)** for your environment
3. **Follow [Developer Guide](developer-guide.md)** for instrumentation
4. **Use [Monitoring](monitoring.md)** for troubleshooting

---

*Next: [Tools & Libraries](tools.md) | [Implementation Details](implementation.md)*</content>
</xai:function_call">docs/tracing/overview.md