# Monitoring & Troubleshooting

This document covers how to monitor distributed tracing in production and troubleshoot common issues.

## 📊 Jaeger UI Usage

### Accessing Jaeger UI

**Development:**
- URL: `http://localhost:16686`
- No authentication required
- All traces visible by default

**Production:**
- URL: `https://jaeger.company.com` (example)
- May require authentication
- Sampling may limit visible traces

### UI Navigation

#### Search Interface
- **Service**: Select service to view traces for
- **Operation**: Filter by specific operations
- **Tags**: Filter by span attributes
- **Duration**: Filter by trace duration
- **Time Range**: Lookback period for traces

#### Trace View
- **Timeline**: Visual representation of span durations
- **Span Details**: Click spans to see attributes and events
- **Service Dependencies**: Shows service communication patterns

### Common Search Patterns

```bash
# Find all traces for a service
Service: user-service

# Find slow requests (>1 second)
Service: api-gateway, Duration: >1000ms

# Find error traces
Service: auth-service, Tags: error=true

# Find traces with specific operation
Service: user-service, Operation: db.get_user_by_email
```

## 🔍 Debugging with Traces

### Identifying Performance Issues

#### Slow Request Analysis
1. **Find Slow Traces**: Search for traces exceeding SLA
2. **Identify Bottlenecks**: Look for spans with longest duration
3. **Check Dependencies**: See if external services are slow
4. **Database Queries**: Check for N+1 queries or missing indexes

#### Example: Slow API Response
```
Trace: GET /api/v1/users
├── API Gateway (50ms)
├── Auth Service (200ms) ← Slow authentication
└── User Service (800ms) ← Very slow database query
```

### Error Correlation

#### Finding Error Patterns
1. **Search Error Traces**: Filter by error status
2. **Check Error Messages**: Look at span events and attributes
3. **Identify Root Cause**: Follow error propagation across services
4. **Check Dependencies**: See if errors cascade from one service to another

#### Example: Database Connection Error
```
Trace: POST /api/v1/users
├── API Gateway (20ms)
├── User Service (5000ms)
│   ├── db.connect (4990ms) ← Connection timeout
│   └── span.record_error: "connection refused"
```

### Service Dependency Analysis

#### Understanding Service Communication
1. **View Dependencies**: Use Jaeger dependency graph
2. **Identify Coupling**: See which services communicate frequently
3. **Check Propagation**: Ensure trace context flows correctly
4. **Monitor Latency**: Track inter-service call durations

## 🚨 Troubleshooting Common Issues

### Traces Not Appearing

#### Check Configuration
```bash
# Verify tracing is enabled
grep "tracing.enabled" services/*/config.yaml

# Check service names are unique
grep "service_name" services/*/config.yaml | sort | uniq -c

# Verify collector URL
grep "collector_url" services/*/config.yaml
```

#### Check Connectivity
```bash
# Test Jaeger OTLP endpoint
curl -X POST http://localhost:4318/v1/traces \
  -H "Content-Type: application/x-protobuf"

# Check service logs for export errors
docker logs user-service | grep -i trace

# Verify network connectivity
docker exec user-service curl -f http://jaeger:4318/v1/traces
```

#### Check Sampling
```bash
# Temporarily increase sampling for debugging
echo "tracing:
  sampling_rate: 1.0" > services/debug-service/config.yaml

# Restart service
make restart-debug-service
```

### Incorrect Trace Context

#### Missing Trace Headers
```bash
# Check if API Gateway injects headers
curl -v http://localhost:8080/api/v1/users \
  -H "Accept: application/json"

# Look for traceparent header in response
# Should see: traceparent: 00-...
```

#### Service Client Issues
```go
// Check client code injects headers
req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
```

### Performance Issues

#### High Memory Usage
```bash
# Check span processor queue
# Default queue size is 2048 spans
# Increase if needed in advanced config

# Monitor span creation rate
docker stats user-service
```

#### High CPU Usage
```bash
# Check sampling rate
grep "sampling_rate" services/*/config.yaml

# Reduce sampling in production
# 0.01 (1%) is typical for high-traffic services
```

### Jaeger Issues

#### Jaeger Not Receiving Traces
```bash
# Check Jaeger logs
docker logs jaeger

# Verify OTLP endpoint is enabled
docker exec jaeger env | grep OTLP

# Test OTLP ingestion
curl http://localhost:4318/v1/traces
```

#### Jaeger UI Not Loading
```bash
# Check Jaeger container status
docker ps | grep jaeger

# Verify port mapping
docker port jaeger

# Check Jaeger health
curl http://localhost:16686/api/services
```

## 📈 Performance Monitoring

### Key Metrics to Monitor

#### Tracing Metrics
- **Span Creation Rate**: Spans/second per service
- **Export Success Rate**: Percentage of spans successfully exported
- **Export Latency**: Time to send spans to Jaeger
- **Queue Size**: Number of spans waiting to be exported

#### Application Metrics
- **Request Latency**: End-to-end request duration
- **Error Rate**: Percentage of requests with errors
- **Throughput**: Requests/second per service

### Setting Up Alerts

#### Jaeger Health Alerts
```yaml
# Prometheus alert rules
groups:
  - name: jaeger
    rules:
      - alert: JaegerDown
        expr: up{job="jaeger"} == 0
        for: 5m
        labels:
          severity: critical

      - alert: JaegerHighQueue
        expr: jaeger_collector_queue_size > 1000
        for: 5m
        labels:
          severity: warning
```

#### Tracing Performance Alerts
```yaml
- alert: HighSpanDropRate
  expr: rate(jaeger_collector_spans_dropped_total[5m]) > 0.1
  for: 5m
  labels:
    severity: warning

- alert: SlowTraceExport
  expr: histogram_quantile(0.95, rate(jaeger_collector_batch_send_duration_seconds_bucket[5m])) > 10
  for: 5m
  labels:
    severity: warning
```

## 🔧 Advanced Troubleshooting

### Trace Sampling Issues

#### Inconsistent Sampling
```go
// Check sampling decision in code
sampler := trace.ParentBased(
    trace.TraceIDRatioBased(0.1),
)

// For debugging, force sampling
sampler := trace.AlwaysSample()
```

#### Missing Root Spans
- **Cause**: Sampling decision at API Gateway
- **Solution**: Ensure gateway uses appropriate sampling
- **Debug**: Check gateway configuration

### Context Propagation Issues

#### Async Operations
```go
// Ensure context is passed to goroutines
go func(ctx context.Context) {
    // Create child span
    tracer := otel.Tracer("service")
    _, span := tracer.Start(ctx, "async.operation")
    defer span.End()

    // Work...
}(ctx)
```

#### Message Queues
```go
// Inject trace context into messages
carrier := propagation.MapCarrier{}
otel.GetTextMapPropagator().Inject(ctx, carrier)

// Add to message headers
headers := make(map[string]string)
carrier.ForeachKey(func(k, v string) bool {
    headers[k] = v
    return true
})
```

### Database Tracing

#### Missing Database Spans
```go
// Ensure context is passed to database calls
func (r *Repository) GetUser(ctx context.Context, id string) (*User, error) {
    return r.db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = ?", id).Scan(...)
}
```

#### Slow Database Queries
- **Check Indexes**: Look for missing database indexes
- **Query Optimization**: Review query execution plans
- **Connection Pooling**: Ensure proper connection pool configuration

## 📊 Analyzing Trace Data

### Performance Analysis

#### Latency Breakdown
1. **Network Latency**: Time spent in HTTP transport
2. **Serialization**: JSON marshaling/unmarshaling time
3. **Database Time**: Query execution time
4. **Business Logic**: Application processing time

#### Throughput Analysis
- **Concurrent Requests**: How many requests processed simultaneously
- **Queue Time**: Time spent waiting in queues
- **Resource Utilization**: CPU, memory, network usage

### Error Analysis

#### Error Patterns
- **Frequent Errors**: Common error types and locations
- **Error Propagation**: How errors spread across services
- **Recovery Time**: Time to recover from errors

#### Root Cause Analysis
1. **Identify Symptoms**: What the user experienced
2. **Find Error Source**: Which service/component failed
3. **Check Dependencies**: What caused the failure
4. **Implement Fix**: Address root cause and add monitoring

## 🚀 Production Best Practices

### Monitoring Setup
- **Centralized Logging**: Correlate traces with logs
- **Metrics Collection**: Track key performance indicators
- **Alert Configuration**: Set up appropriate alerting thresholds
- **Dashboard Creation**: Build operational dashboards

### Capacity Planning
- **Traffic Patterns**: Understand peak and off-peak usage
- **Resource Scaling**: Plan for traffic increases
- **Storage Requirements**: Estimate trace storage needs
- **Network Capacity**: Ensure sufficient bandwidth for trace export

### Incident Response
- **Runbook Creation**: Document troubleshooting procedures
- **Escalation Paths**: Define when to escalate issues
- **Post-Mortem Process**: Learn from incidents
- **Continuous Improvement**: Update monitoring based on findings

---

*Next: [Best Practices](best-practices.md) | [Configuration](configuration.md)*</content>
</xai:function_call">write">
<parameter name="filePath">docs/tracing/best-practices.md