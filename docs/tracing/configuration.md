# Configuration & Environment Setup

This document covers how to configure distributed tracing for different environments and deployment scenarios.

## 📋 Configuration Structure

### Tracing Configuration Schema

```yaml
tracing:
  enabled: true                    # Enable/disable tracing
  service_name: "my-service"       # Unique service identifier
  collector_url: "http://jaeger:4318/v1/traces"  # Jaeger endpoint
  sampling_rate: 1.0               # Sampling ratio (0.0-1.0)
```

### Configuration Location

#### Service-Specific Configuration
```
services/{service-name}/config.yaml
```

#### Template Configuration
```
templates/service-template/config.yaml  # Used for new services
```

## 🌍 Environment-Specific Configurations

### Development Environment

**File:** `services/{service-name}/config.yaml`

```yaml
tracing:
  enabled: true
  service_name: "user-service"
  collector_url: "http://localhost:4318/v1/traces"  # Local Jaeger
  sampling_rate: 1.0                               # 100% sampling
```

**Docker Override:** `docker/docker-compose.override.yml`

```yaml
user-service:
  environment:
    - TRACING_ENABLED=true
    - TRACING_SERVICE_NAME=user-service
    - TRACING_COLLECTOR_URL=http://jaeger:4318/v1/traces
    - TRACING_SAMPLING_RATE=1.0
```

### Staging Environment

```yaml
tracing:
  enabled: true
  service_name: "user-service"
  collector_url: "http://jaeger-staging.company.com:4318/v1/traces"
  sampling_rate: 0.1  # 10% sampling
```

### Production Environment

```yaml
tracing:
  enabled: true
  service_name: "user-service"
  collector_url: "http://jaeger-collector.company.com:4318/v1/traces"
  sampling_rate: 0.01  # 1% sampling
```

## 🐳 Docker Configuration

### Docker Compose Services

#### Jaeger Service Configuration

```yaml
jaeger:
  image: jaegertracing/all-in-one:latest
  ports:
    - "16686:16686"    # Jaeger UI
    - "4318:4318"      # OTLP HTTP collector
    - "14268:14268"    # Jaeger Thrift (HTTP)
    - "14250:14250"    # Jaeger Thrift (gRPC)
  environment:
    - COLLECTOR_OTLP_ENABLED=true
    - COLLECTOR_ZIPKIN_HOST_PORT=:9411
    - SPAN_STORAGE_TYPE=memory  # For development
  volumes:
    - jaeger_data:/tmp
  networks:
    - service-network
  restart: unless-stopped
```

#### Service Tracing Environment Variables

```yaml
user-service:
  environment:
    - TRACING_ENABLED=${TRACING_ENABLED:-true}
    - TRACING_SERVICE_NAME=${USER_SERVICE_NAME}
    - TRACING_COLLECTOR_URL=${TRACING_COLLECTOR_URL:-http://jaeger:4318/v1/traces}
    - TRACING_SAMPLING_RATE=${TRACING_SAMPLING_RATE:-1.0}
  depends_on:
    jaeger:
      condition: service_started  # Jaeger doesn't have health checks
```

### Environment Variables Mapping

| Environment Variable | Config Field | Default |
|---------------------|--------------|---------|
| `TRACING_ENABLED` | `tracing.enabled` | `true` |
| `TRACING_SERVICE_NAME` | `tracing.service_name` | Service name |
| `TRACING_COLLECTOR_URL` | `tracing.collector_url` | `http://jaeger:4318/v1/traces` |
| `TRACING_SAMPLING_RATE` | `tracing.sampling_rate` | `1.0` |

## 📊 Sampling Strategies

### Development
```yaml
sampling_rate: 1.0  # 100% - capture all traces
```
- **Purpose:** Full visibility for debugging
- **Impact:** Higher resource usage, detailed traces
- **Use Case:** Local development, feature testing

### Staging
```yaml
sampling_rate: 0.1  # 10% - representative sample
```
- **Purpose:** Performance testing with realistic load
- **Impact:** Balanced resource usage
- **Use Case:** Load testing, integration testing

### Production
```yaml
sampling_rate: 0.01  # 1% - minimal overhead
```
- **Purpose:** Observability with minimal performance impact
- **Impact:** Low resource overhead
- **Use Case:** Live production systems

### Dynamic Sampling

For advanced sampling based on service or operation:

```yaml
# High sampling for critical services
tracing:
  sampling_rate: 0.1  # 10% for auth-service

# Lower sampling for high-traffic services
tracing:
  sampling_rate: 0.01  # 1% for user-service
```

## 🔧 Advanced Configuration

### Custom Resource Attributes

```go
// In tracer initialization (advanced customization)
res, err := resource.New(context.Background(),
    resource.WithAttributes(
        semconv.ServiceNameKey.String(cfg.ServiceName),
        semconv.ServiceVersionKey.String("1.0.0"),
        semconv.ServiceInstanceIDKey.String(hostname),
        semconv.DeploymentEnvironmentKey.String("production"),
        semconv.TelemetrySDKLanguageGo,
    ),
)
```

### Batch Span Processor Configuration

```go
// Custom batch processor (if needed)
bsp := trace.NewBatchSpanProcessor(exp,
    trace.WithBatchTimeout(5*time.Second),      // Send batch every 5s
    trace.WithMaxExportBatchSize(512),          // Max spans per batch
    trace.WithMaxQueueSize(2048),               // Queue size
    trace.WithExportTimeout(10*time.Second),    // Export timeout
)
```

### Custom Sampling Rules

```go
// Advanced sampling based on operation
sampler := trace.ParentBased(
    trace.NewSampler(func(params trace.SamplingParameters) trace.SamplingResult {
        // Sample all errors
        if params.Kind == trace.SpanKindServer && containsError(params.Tags) {
            return trace.SamplingResult{Decision: trace.RecordAndSample}
        }
        // Sample based on URL patterns
        if strings.Contains(params.Name, "/api/v1/auth/") {
            return trace.SamplingResult{Decision: trace.RecordAndSample}
        }
        // Default ratio sampling
        return trace.TraceIDRatioBased(0.1).ShouldSample(params)
    }),
)
```

## 🌐 Multi-Environment Setup

### Local Development

**`.env` file:**
```bash
# Tracing Configuration
TRACING_ENABLED=true
TRACING_COLLECTOR_URL=http://localhost:4318/v1/traces
TRACING_SAMPLING_RATE=1.0

# Service-specific overrides
USER_SERVICE_TRACING_SAMPLING_RATE=1.0
AUTH_SERVICE_TRACING_SAMPLING_RATE=1.0
```

**Docker Compose Override:**
```yaml
jaeger:
  ports:
    - "16686:16686"
    - "4318:4318"
```

### Kubernetes Deployment

**ConfigMap:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tracing-config
data:
  TRACING_ENABLED: "true"
  TRACING_COLLECTOR_URL: "http://jaeger-collector:4318/v1/traces"
  TRACING_SAMPLING_RATE: "0.01"
```

**Deployment Environment Variables:**
```yaml
envFrom:
  - configMapRef:
      name: tracing-config
env:
  - name: TRACING_SERVICE_NAME
    valueFrom:
      fieldRef:
        fieldPath: metadata.labels['app.kubernetes.io/name']
```

### AWS ECS/Fargate

**Task Definition Environment:**
```json
{
  "environment": [
    {"name": "TRACING_ENABLED", "value": "true"},
    {"name": "TRACING_COLLECTOR_URL", "value": "http://jaeger-collector.local:4318/v1/traces"},
    {"name": "TRACING_SAMPLING_RATE", "value": "0.01"},
    {"name": "TRACING_SERVICE_NAME", "value": "user-service"}
  ]
}
```

## 🔍 Configuration Validation

### Health Check Integration

```go
// Add tracing health check
func (h *HealthHandler) TracingHealth() gin.H {
    // Check if tracer is initialized
    if tracer := otel.Tracer("health-check"); tracer != nil {
        ctx, span := tracer.Start(context.Background(), "health.tracing")
        span.End()
        return gin.H{"status": "ok", "tracing": "enabled"}
    }
    return gin.H{"status": "degraded", "tracing": "disabled"}
}
```

### Configuration Validation

```go
func validateTracingConfig(cfg config.TracingConfig) error {
    if cfg.Enabled {
        if cfg.ServiceName == "" {
            return fmt.Errorf("service_name is required when tracing is enabled")
        }
        if cfg.CollectorURL == "" {
            return fmt.Errorf("collector_url is required when tracing is enabled")
        }
        if cfg.SamplingRate < 0 || cfg.SamplingRate > 1 {
            return fmt.Errorf("sampling_rate must be between 0.0 and 1.0")
        }
    }
    return nil
}
```

## 📈 Monitoring Configuration

### Jaeger UI Access

**Development:**
- URL: `http://localhost:16686`
- No authentication required

**Production:**
- URL: `https://jaeger.company.com`
- Authentication may be required
- HTTPS with proper certificates

### Metrics Integration

```yaml
# Prometheus metrics for tracing
tracing:
  enabled: true
  metrics:
    enabled: true
    endpoint: "/metrics"
```

## 🚨 Troubleshooting Configuration

### Common Issues

#### Traces Not Appearing
```bash
# Check Jaeger connectivity
curl http://localhost:4318/v1/traces

# Verify configuration
cat services/my-service/config.yaml

# Check service logs
docker logs my-service
```

#### Incorrect Service Names
```bash
# Check service name in traces
grep "service_name" services/*/config.yaml

# Verify uniqueness
grep "service_name" services/*/config.yaml | sort | uniq -c
```

#### Sampling Too Low
```bash
# Temporarily increase sampling for debugging
echo "tracing:\n  sampling_rate: 1.0" > services/my-service/config.debug.yaml
# Restart service
```

### Configuration Testing

#### Unit Tests
```go
func TestTracingConfig(t *testing.T) {
    cfg := config.TracingConfig{
        Enabled:      true,
        ServiceName:  "test-service",
        CollectorURL: "http://jaeger:4318/v1/traces",
        SamplingRate: 0.1,
    }

    err := validateTracingConfig(cfg)
    assert.NoError(t, err)

    // Test tracer initialization
    tracer, err := tracing.InitTracer(cfg)
    assert.NoError(t, err)
    assert.NotNil(t, tracer)
}
```

#### Integration Tests
```go
func TestTracingEndToEnd(t *testing.T) {
    // Start Jaeger
    // Deploy service with tracing
    // Make requests
    // Verify traces in Jaeger API
}
```

## 📚 Related Documentation

- **[Tools & Libraries](tools.md)**: Technical dependencies and versions
- **[Monitoring](monitoring.md)**: Jaeger UI usage and debugging
- **[Best Practices](best-practices.md)**: Configuration recommendations

---

*Next: [Monitoring & Troubleshooting](monitoring.md) | [Best Practices](best-practices.md)*</content>
</xai:function_call">docs/tracing/monitoring.md