# Distributed Tracing Implementation Plan

## Overview

This document outlines the comprehensive implementation plan for adding distributed tracing capabilities to the microservices boilerplate. The implementation follows OpenTelemetry standards with Jaeger as the tracing backend, enabling complete observability of request flows across API Gateway → User Service → Auth Service.

## Current Status

- **Phase 3.1 (Enhanced Logging)**: ✅ COMPLETED
- **Phase 3.2 (Advanced Monitoring)**: 🔄 PENDING
- **Phase 3.3 (Distributed Tracing)**: 🎯 READY FOR IMPLEMENTATION

## Architecture Overview

### Tech Stack
- **OpenTelemetry Go SDK**: Industry-standard tracing library
- **Jaeger**: Tracing backend with visualization UI
- **W3C Trace Context**: Standard header propagation
- **HTTP Instrumentation**: Automatic span creation for HTTP requests

### Request Flow Tracing
```
Client Request → API Gateway → User Service → Auth Service
     ↓             ↓            ↓            ↓
  Root Span    Gateway Span  Service Spans  Auth Spans
     ↓             ↓            ↓            ↓
 Jaeger UI: Complete request timeline with spans
```

## Implementation Phases

### Phase 3.3.1: Core Tracing Infrastructure

#### 1. Add Tracing Dependencies
**Location**: `go.mod`
**Dependencies**:
```go
go.opentelemetry.io/otel v1.24.0
go.opentelemetry.io/otel/exporters/jaeger v1.17.0
go.opentelemetry.io/otel/sdk v1.24.0
go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.49.0
```

**Commands**:
```bash
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/exporters/jaeger
go get go.opentelemetry.io/otel/sdk
go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
```

#### 2. Create Tracing Configuration
**Location**: `common/config/config.go`
**Changes**:
```go
type TracingConfig struct {
    Enabled         bool    `mapstructure:"enabled"`
    ServiceName     string  `mapstructure:"service_name"`
    CollectorURL    string  `mapstructure:"collector_url"`
    SamplingRate    float64 `mapstructure:"sampling_rate"`
}

type Config struct {
    // ... existing fields
    Tracing TracingConfig `mapstructure:"tracing"`
}
```

**Environment Variables**:
```bash
TRACING_ENABLED=true
TRACING_SERVICE_NAME=api-gateway
TRACING_COLLECTOR_URL=http://jaeger:14268/api/traces
TRACING_SAMPLING_RATE=1.0
```

#### 3. Create Common Tracing Package
**Location**: `common/tracing/tracer.go`
**Files to Create**:
- `common/tracing/tracer.go` - Core tracer initialization
- `common/tracing/middleware.go` - HTTP middleware for span creation

**Core Implementation**:
```go
package tracing

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func InitTracer(config TracingConfig) (*trace.TracerProvider, error) {
    exp, err := jaeger.New(jaeger.WithCollectorEndpoint(
        jaeger.WithEndpoint(config.CollectorURL),
    ))
    if err != nil {
        return nil, err
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exp),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String(config.ServiceName),
        )),
        trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(config.SamplingRate))),
    )

    otel.SetTracerProvider(tp)
    return tp, nil
}
```

### Phase 3.3.2: Gateway Tracing Implementation

#### 4. Add Gateway Tracing Middleware
**Location**: `api-gateway/internal/middleware/tracing.go`
**Implementation**:
```go
func TracingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tracer := otel.Tracer("api-gateway")

        ctx, span := tracer.Start(c.Request.Context(),
            fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path))
        defer span.End()

        // Add span attributes
        span.SetAttributes(
            semconv.HTTPMethodKey.String(c.Request.Method),
            semconv.HTTPURLKey.String(c.Request.URL.String()),
            semconv.HTTPUserAgentKey.String(c.Request.UserAgent()),
        )

        c.Request = c.Request.WithContext(ctx)
        c.Next()

        // Set response attributes
        span.SetAttributes(
            semconv.HTTPStatusCodeKey.Int(c.Writer.Status()),
        )

        if len(c.Errors) > 0 {
            span.RecordError(c.Errors.Last())
        }
    }
}
```

#### 5. Modify Proxy Handler for Trace Injection
**Location**: `api-gateway/internal/handlers/gateway.go`
**Changes**:
- Extract trace context from current span
- Inject `traceparent` and `tracestate` headers into downstream requests
- Ensure trace context propagation

**Implementation**:
```go
func (h *GatewayHandler) ProxyRequest(serviceName string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // ... existing code ...

        // Inject trace context headers
        span := trace.SpanFromContext(c.Request.Context())
        if span.SpanContext().IsValid() {
            traceparent := fmt.Sprintf("00-%s-%s-%s",
                span.SpanContext().TraceID().String(),
                span.SpanContext().SpanID().String(),
                span.SpanContext().TraceFlags().String(),
            )
            req.Header.Set("traceparent", traceparent)

            if tracestate := span.SpanContext().TraceState().String(); tracestate != "" {
                req.Header.Set("tracestate", tracestate)
            }
        }

        // ... rest of proxy logic ...
    }
}
```

### Phase 3.3.3: Service-Level Tracing

#### 6. Add User Service Spans
**Location**: `services/user-service/internal/handlers/user_handler.go`
**Operations to Instrument**:
- `CreateUser` - User creation span
- `GetUser` - User retrieval span
- `UpdateUser` - User update span
- `DeleteUser` - User deletion span
- `ListUsers` - User listing span

**Example Implementation**:
```go
func (h *UserHandler) CreateUser(c *gin.Context) {
    tracer := otel.Tracer("user-service")

    ctx, span := tracer.Start(c.Request.Context(), "CreateUser")
    defer span.End()

    span.SetAttributes(
        attribute.String("operation", "create_user"),
        attribute.String("service", "user-service"),
    )

    // ... existing logic with ctx ...
}
```

#### 7. Add Auth Service Spans
**Location**: `services/auth-service/internal/handlers/auth_handler.go`
**Operations to Instrument**:
- `Login` - Authentication span
- `Register` - User registration span
- `RefreshToken` - Token refresh span
- `Logout` - Logout span
- `GetCurrentUser` - User info retrieval span

**Example Implementation**:
```go
func (h *AuthHandler) Login(c *gin.Context) {
    tracer := otel.Tracer("auth-service")

    ctx, span := tracer.Start(c.Request.Context(), "Login")
    defer span.End()

    span.SetAttributes(
        attribute.String("operation", "user_login"),
        attribute.String("auth.method", "password"),
        attribute.String("service", "auth-service"),
    )

    // ... existing logic with ctx ...
}
```

### Phase 3.3.4: Infrastructure Setup

#### 8. Add Jaeger to Docker Compose
**Location**: `docker/docker-compose.yml`
**Service Configuration**:
```yaml
services:
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"  # Jaeger UI
      - "14268:14268"  # Accept jaeger.thrift over HTTP
      - "14250:14250"  # Accept jaeger.thrift over gRPC
    environment:
      - COLLECTOR_OTLP_ENABLED=true
      - COLLECTOR_ZIPKIN_HOST_PORT=:9411
    volumes:
      - jaeger_data:/tmp
    networks:
      - service-network
    restart: unless-stopped

volumes:
  jaeger_data:
```

#### 9. Update Service Configurations
**Location**: All `config.yaml` files
**Configuration Template**:
```yaml
tracing:
  enabled: true
  service_name: "api-gateway"  # Change per service
  collector_url: "http://jaeger:14268/api/traces"
  sampling_rate: 1.0  # 1.0 = 100% sampling in dev
```

**Environment Variables** (`.env`):
```bash
# API Gateway Tracing
TRACING_ENABLED=true
TRACING_SERVICE_NAME=api-gateway
TRACING_COLLECTOR_URL=http://jaeger:14268/api/traces
TRACING_SAMPLING_RATE=1.0

# User Service Tracing
USER_SERVICE_TRACING_ENABLED=true
USER_SERVICE_TRACING_SERVICE_NAME=user-service
# ... etc for other services
```

### Phase 3.3.5: Service Initialization Updates

#### 10. Update Service Main Functions
**Location**: All `cmd/main.go` files
**Pattern**:
```go
func main() {
    // Load config
    cfg, err := config.Load(".")
    // ... error handling ...

    // Initialize tracer if enabled
    if cfg.Tracing.Enabled {
        tracerProvider, err := tracing.InitTracer(cfg.Tracing)
        if err != nil {
            logger.Warn("Failed to initialize tracing", err)
        } else {
            defer func() {
                if err := tracerProvider.Shutdown(context.Background()); err != nil {
                    logger.Error("Failed to shutdown tracer", err)
                }
            }()
        }
    }

    // ... rest of initialization ...
}
```

### Phase 3.3.6: Testing & Validation

#### 11. End-to-End Testing
**Test Scenarios**:
1. **Simple Request**: Gateway → User Service (GET user)
2. **Complex Flow**: Gateway → User Service → Auth Service (user creation with auth)
3. **Error Scenarios**: Failed requests, timeouts, service unavailability
4. **Header Propagation**: Verify trace context headers are passed correctly

**Test Commands**:
```bash
# Start services with tracing
make dev

# Access Jaeger UI
open http://localhost:16686

# Make test requests
curl http://localhost:8080/api/v1/users
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Check trace propagation
curl http://localhost:8080/api/v1/users/1
```

#### 12. Performance Validation
**Metrics to Monitor**:
- Request latency increase (<5% overhead)
- Memory usage of tracing components
- CPU overhead of span creation
- Network traffic to Jaeger collector

**Performance Tests**:
```bash
# Load testing with tracing enabled/disabled
ab -n 1000 -c 10 http://localhost:8080/health
```

## Implementation Timeline

### Week 1: Core Infrastructure (Days 1-2)
- [ ] Add OpenTelemetry dependencies
- [ ] Create tracing configuration structure
- [ ] Implement common tracing package
- [ ] Add Jaeger to docker-compose.yml

### Week 2: Gateway Implementation (Days 3-4)
- [ ] Create gateway tracing middleware
- [ ] Modify proxy handler for trace injection
- [ ] Update gateway configuration
- [ ] Test gateway-level tracing

### Week 3: Service Integration (Days 5-6)
- [ ] Add spans to user-service handlers
- [ ] Add spans to auth-service handlers
- [ ] Update all service configurations
- [ ] Test service-level tracing

### Week 4: Testing & Documentation (Days 7-8)
- [ ] End-to-end testing across all services
- [ ] Performance validation and optimization
- [ ] Update documentation
- [ ] Create troubleshooting guide

## Success Criteria

### Functional Requirements
- [ ] **Complete Trace Visibility**: All request flows visible in Jaeger UI
- [ ] **Span Coverage**: Spans for all major operations (CRUD, auth, proxy)
- [ ] **Header Propagation**: W3C trace context passed between services
- [ ] **Error Recording**: Exceptions and errors recorded in spans
- [ ] **Service Identification**: Clear service attribution in traces

### Performance Requirements
- [ ] **Latency Overhead**: <5% increase in request response times
- [ ] **Memory Usage**: Minimal additional memory consumption
- [ ] **CPU Overhead**: Negligible CPU impact from tracing
- [ ] **Network Traffic**: Efficient batch export to Jaeger

### Operational Requirements
- [ ] **Configuration**: Enable/disable tracing without code changes
- [ ] **Sampling**: Configurable sampling rates for production
- [ ] **Graceful Degradation**: Continue operation if tracing fails
- [ ] **Resource Cleanup**: Proper shutdown and resource management

## Configuration Examples

### Development Configuration
```yaml
tracing:
  enabled: true
  service_name: "api-gateway"
  collector_url: "http://jaeger:14268/api/traces"
  sampling_rate: 1.0  # Sample all requests
```

### Production Configuration
```yaml
tracing:
  enabled: true
  service_name: "api-gateway"
  collector_url: "http://jaeger-collector.company.com:14268/api/traces"
  sampling_rate: 0.1  # Sample 10% of requests
```

## Troubleshooting Guide

### Common Issues

#### Traces Not Appearing in Jaeger
**Symptoms**: No traces visible in Jaeger UI
**Solutions**:
1. Check Jaeger service is running: `docker ps | grep jaeger`
2. Verify collector URL configuration
3. Check service logs for tracing initialization errors
4. Ensure sampling rate > 0

#### Missing Spans in Trace
**Symptoms**: Incomplete traces, missing service spans
**Solutions**:
1. Verify tracing is enabled in service configuration
2. Check for span creation in handler code
3. Ensure trace context propagation between services
4. Validate header injection in proxy requests

#### High Performance Overhead
**Symptoms**: Increased latency, high resource usage
**Solutions**:
1. Reduce sampling rate in production
2. Disable tracing for health check endpoints
3. Optimize span attribute collection
4. Use batch export settings

#### Header Propagation Issues
**Symptoms**: Broken trace chains between services
**Solutions**:
1. Verify `traceparent` header injection in gateway
2. Check W3C trace context format compliance
3. Ensure downstream services extract trace context
4. Validate span context extraction in service handlers

## Dependencies & Prerequisites

### Required Skills
- Go programming language
- OpenTelemetry concepts
- Distributed systems understanding
- Docker containerization
- HTTP middleware patterns

### Required Tools
- Go 1.23+
- Docker & Docker Compose
- Jaeger (tracing backend)
- Web browser (for Jaeger UI)

### Testing Requirements
- Multi-service test environment
- Load testing tools (Apache Bench, hey, etc.)
- Network traffic monitoring
- Performance profiling tools

## Future Enhancements

### Phase 3.4: Advanced Tracing Features
- [ ] Custom span attributes and tags
- [ ] Business metric correlation
- [ ] Trace-based alerting
- [ ] Integration with metrics systems
- [ ] Custom sampling strategies

### Phase 3.5: Observability Dashboard
- [ ] Jaeger UI integration
- [ ] Custom dashboards with Grafana
- [ ] Alerting based on trace data
- [ ] Performance trend analysis

## References

- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/go/)
- [Jaeger Tracing](https://www.jaegertracing.io/)
- [W3C Trace Context](https://www.w3.org/TR/trace-context/)
- [OpenTelemetry Semantic Conventions](https://opentelemetry.io/docs/semantic_conventions/)

---

## Implementation Checklist

### Pre-Implementation
- [ ] Review OpenTelemetry best practices
- [ ] Set up Jaeger development environment
- [ ] Create tracing configuration templates
- [ ] Plan span naming conventions

### Core Implementation
- [ ] Add OpenTelemetry dependencies
- [ ] Create tracing configuration structures
- [ ] Implement common tracing package
- [ ] Add Jaeger infrastructure

### Gateway Implementation
- [ ] Create tracing middleware
- [ ] Modify proxy handler for trace injection
- [ ] Update gateway configuration

### Service Implementation
- [ ] Add spans to user-service operations
- [ ] Add spans to auth-service operations
- [ ] Update service initialization code

### Testing & Validation
- [ ] End-to-end tracing tests
- [ ] Performance impact assessment
- [ ] Error scenario testing
- [ ] Documentation updates

### Deployment & Monitoring
- [ ] Production configuration setup
- [ ] Monitoring and alerting setup
- [ ] Performance baseline establishment
- [ ] Operational runbook creation

This comprehensive plan provides a complete roadmap for implementing distributed tracing across the microservices architecture, ensuring full observability and debugging capabilities for production systems.</content>
</xai:function_call