# Best Practices for Distributed Tracing

This document outlines guidelines and recommendations for implementing and maintaining distributed tracing in production systems.

## 🎯 General Principles

### Start with Business Value
- **Focus on User Impact**: Trace requests that affect user experience
- **Measure Business Metrics**: Track conversion funnels, not just technical metrics
- **Prioritize Critical Paths**: Instrument payment flows, user registration, search operations

### Keep It Simple
- **Avoid Over-Instrumentation**: Don't trace every function call
- **Use Appropriate Sampling**: Balance observability with performance
- **Standardize Naming**: Consistent span and attribute names across services

## 📏 Span Management

### Span Naming Conventions

#### HTTP Operations
```go
// ✅ Good: Descriptive and consistent
spanName := "http.get_user_profile"

// ❌ Bad: Too generic
spanName := "handler"

// ❌ Bad: Too specific
spanName := "GET /api/v1/users/123/profile?include=preferences"
```

#### Database Operations
```go
// ✅ Good: Operation and table
ctx, span := tracer.Start(ctx, "db.select_users")

// ✅ Good: Include query type
ctx, span := tracer.Start(ctx, "db.get_user_by_email")

// ❌ Bad: Too vague
ctx, span := tracer.Start(ctx, "database_call")
```

#### Business Operations
```go
// ✅ Good: Business context
ctx, span := tracer.Start(ctx, "user.registration.process")

// ✅ Good: Action and entity
ctx, span := tracer.Start(ctx, "order.payment.charge")

// ❌ Bad: Technical implementation
ctx, span := tracer.Start(ctx, "gin_handler")
```

### Span Lifecycle

#### Always End Spans
```go
// ✅ Good: Proper cleanup
func operation(ctx context.Context) error {
    ctx, span := tracer.Start(ctx, "operation")
    defer span.End() // Always called

    // ... work ...
    return nil
}

// ❌ Bad: Span may leak
func operation(ctx context.Context) error {
    ctx, span := tracer.Start(ctx, "operation")
    // Missing defer span.End()

    // ... work ...
    return nil
}
```

#### Handle Errors Properly
```go
// ✅ Good: Record errors and set status
func operation(ctx context.Context) error {
    ctx, span := tracer.Start(ctx, "operation")
    defer span.End()

    err := doWork()
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    span.SetStatus(codes.Ok, "")
    return nil
}
```

### Span Attributes

#### Use Semantic Conventions
```go
// ✅ Good: Standard attributes
span.SetAttributes(
    semconv.HTTPMethodKey.String("GET"),
    semconv.HTTPStatusCodeKey.Int(200),
    semconv.DBSystemKey.String("postgresql"),
    semconv.DBNameKey.String("user_db"),
)

// ❌ Bad: Custom attributes for standard things
span.SetAttributes(
    attribute.String("http_method", "GET"), // Use semconv.HTTPMethodKey
    attribute.String("db_type", "postgres"), // Use semconv.DBSystemKey
)
```

#### Business Attributes
```go
// ✅ Good: Business-relevant attributes
span.SetAttributes(
    attribute.String("user.tier", "premium"),
    attribute.Float64("order.amount", 99.99),
    attribute.String("payment.method", "credit_card"),
)

// ❌ Bad: Sensitive data
span.SetAttributes(
    attribute.String("user.password", "secret123"), // Never log sensitive data
    attribute.String("payment.card_number", "4111111111111111"), // Never log PII
)
```

## 🔄 Context Propagation

### Always Pass Context
```go
// ✅ Good: Context flows through all layers
func (h *Handler) HandleRequest(c *gin.Context) {
    ctx := c.Request.Context()
    user, err := h.userService.GetUser(ctx, userID)
    // Context contains trace information
}

func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    return s.userRepo.GetUser(ctx, id) // Pass context to repository
}

// ❌ Bad: Context not passed
func (s *UserService) GetUser(id string) (*User, error) {
    return s.userRepo.GetUser(id) // Missing context
}
```

### HTTP Client Propagation
```go
// ✅ Good: Inject trace headers
func (c *HTTPClient) Get(ctx context.Context, url string) error {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return err
    }

    // Critical: Inject trace context
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

    return c.client.Do(req)
}

// ❌ Bad: Missing header injection
func (c *HTTPClient) Get(ctx context.Context, url string) error {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    // Missing: otel.GetTextMapPropagator().Inject(...)
    return c.client.Do(req)
}
```

## ⚡ Performance Considerations

### Sampling Strategy

#### Development Environment
```yaml
# 100% sampling for full visibility
tracing:
  sampling_rate: 1.0
```

#### Production Environment
```yaml
# Varies by service traffic
tracing:
  sampling_rate: 0.01  # 1% for high-traffic services
  sampling_rate: 0.1   # 10% for medium-traffic services
```

#### Dynamic Sampling
```go
// Sample based on operation importance
sampler := trace.ParentBased(
    trace.NewSampler(func(params trace.SamplingParameters) trace.SamplingResult {
        // Always sample errors
        if containsError(params.Tags) {
            return trace.SamplingResult{Decision: trace.RecordAndSample}
        }

        // Sample important operations
        if isCriticalOperation(params.Name) {
            return trace.SamplingResult{Decision: trace.RecordAndSample}
        }

        // Default ratio sampling
        return trace.TraceIDRatioBased(0.01).ShouldSample(params)
    }),
)
```

### Resource Management

#### Batch Configuration
```go
// Optimize for your workload
bsp := trace.NewBatchSpanProcessor(exp,
    trace.WithBatchTimeout(5*time.Second),      // Balance latency vs throughput
    trace.WithMaxExportBatchSize(512),          // Batch size
    trace.WithMaxQueueSize(2048),               // Memory usage
    trace.WithExportTimeout(10*time.Second),    // Fail fast
)
```

#### Memory Bounds
- **Queue Size**: Limit spans in memory (default: 2048)
- **Batch Size**: Balance network efficiency vs latency
- **Timeout**: Prevent hanging on export failures

## 🏗️ Architecture Patterns

### Service Boundaries

#### Clear Service Ownership
```go
// ✅ Good: Each service owns its operations
// API Gateway: Request routing and header injection
// Auth Service: Authentication and authorization
// User Service: User CRUD operations
// Payment Service: Payment processing

// ❌ Bad: Cross-service operations
// Don't have User Service call Payment Service directly
// Use events or API Gateway orchestration instead
```

#### Consistent Error Handling
```go
// ✅ Good: Consistent error propagation
func (h *Handler) ProcessRequest(c *gin.Context) {
    ctx, span := tracer.Start(c.Request.Context(), "process_request")
    defer span.End()

    err := h.businessLogic(ctx)
    if err != nil {
        // Record error on span
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())

        // Return appropriate HTTP status
        c.JSON(getHTTPStatus(err), gin.H{"error": err.Error()})
        return
    }

    span.SetStatus(codes.Ok, "")
    c.JSON(200, result)
}
```

### Async Operations

#### Goroutines with Context
```go
// ✅ Good: Pass context to goroutines
func (h *Handler) ProcessAsync(c *gin.Context) {
    ctx := c.Request.Context()

    go func(operationCtx context.Context) {
        // Create child span for async work
        asyncCtx, span := tracer.Start(operationCtx, "async_processing")
        defer span.End()

        // Process with trace context
        h.processAsync(asyncCtx, data)
    }(ctx)
}

// ❌ Bad: Context not propagated
go func() {
    // Missing context, no tracing
    h.processAsync(data)
}()
```

#### Worker Pools
```go
func (w *WorkerPool) StartWorkers(ctx context.Context) {
    for i := 0; i < w.numWorkers; i++ {
        go func(workerID int) {
            // Worker lifecycle span
            workerCtx, workerSpan := tracer.Start(ctx, "worker_lifecycle")
            defer workerSpan.End()

            workerSpan.SetAttributes(attribute.Int("worker.id", workerID))

            for {
                select {
                case job := <-w.jobQueue:
                    // Job processing span
                    jobCtx, jobSpan := tracer.Start(workerCtx, "process_job")
                    jobSpan.SetAttributes(attribute.String("job.id", job.ID))
                    // ... process job ...
                    jobSpan.End()

                case <-ctx.Done():
                    workerSpan.SetStatus(codes.Ok, "Worker shutdown")
                    return
                }
            }
        }(i)
    }
}
```

## 📊 Monitoring and Alerting

### Key Metrics to Track

#### Tracing Health
- **Export Success Rate**: >99% of spans should be exported
- **Export Latency**: <1 second average
- **Queue Utilization**: <80% of max queue size
- **Span Drop Rate**: 0 drops in normal operation

#### Application Performance
- **P95 Latency**: Track by operation and service
- **Error Rate**: <1% for most operations
- **Throughput**: Monitor for performance regressions

### Alert Configuration

#### Critical Alerts
```yaml
# Jaeger down
- alert: JaegerCollectorDown
  expr: up{job="jaeger-collector"} == 0
  for: 5m
  severity: critical

# High span drop rate
- alert: HighSpanDropRate
  expr: rate(jaeger_collector_spans_dropped_total[5m]) > 0.05
  for: 5m
  severity: warning
```

#### Performance Alerts
```yaml
# Slow service operations
- alert: SlowServiceOperation
  expr: histogram_quantile(0.95, rate(http_request_duration_seconds{operation=~".*"}[5m])) > 5
  for: 5m
  severity: warning

# High error rate
- alert: HighErrorRate
  expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.05
  for: 5m
  severity: warning
```

## 🔧 Operational Practices

### Configuration Management

#### Environment-Specific Configs
```yaml
# development.yaml
tracing:
  enabled: true
  sampling_rate: 1.0
  collector_url: "http://localhost:4318/v1/traces"

# staging.yaml
tracing:
  enabled: true
  sampling_rate: 0.1
  collector_url: "http://jaeger-staging.company.com:4318/v1/traces"

# production.yaml
tracing:
  enabled: true
  sampling_rate: 0.01
  collector_url: "http://jaeger-prod.company.com:4318/v1/traces"
```

#### Feature Flags
```go
// Allow runtime configuration changes
if featureFlags.TracingEnabled() {
    tracerProvider, err := tracing.InitTracer(cfg.Tracing)
    // ...
}
```

### Deployment Considerations

#### Rolling Deployments
- **Canary Releases**: Monitor tracing metrics during rollout
- **Rollback Plan**: Ability to disable tracing if issues arise
- **Gradual Rollout**: Start with low sampling rate

#### Service Mesh Integration
- **Istio/Service Mesh**: Consider using service mesh for automatic tracing
- **Sidecar Proxies**: Offload tracing to sidecar containers
- **Control Plane**: Centralized tracing configuration

### Security Considerations

#### Sensitive Data
```go
// ✅ Good: Sanitize before adding attributes
span.SetAttributes(
    attribute.String("user.id", userID), // OK
    // attribute.String("user.email", email), // Don't log PII
    // attribute.String("payment.card", cardNumber), // Never log
)

// ❌ Bad: Log sensitive information
span.SetAttributes(
    attribute.String("password", "secret123"),
    attribute.String("api_key", "sk-1234567890"),
)
```

#### Access Control
- **Jaeger UI**: Restrict access in production
- **Trace Data**: Consider data retention policies
- **Audit Logging**: Log access to trace data

## 🚀 Advanced Patterns

### Custom Samplers

#### Error-Always Sampling
```go
type ErrorAlwaysSampler struct {
    fallback trace.Sampler
}

func (s *ErrorAlwaysSampler) ShouldSample(params trace.SamplingParameters) trace.SamplingResult {
    // Always sample errors
    if params.Kind == trace.SpanKindServer {
        for _, attr := range params.Attributes {
            if attr.Key == "error" && attr.Value.AsBool() {
                return trace.SamplingResult{Decision: trace.RecordAndSample}
            }
        }
    }

    return s.fallback.ShouldSample(params)
}
```

#### Adaptive Sampling
```go
// Sample more during incidents
func (s *AdaptiveSampler) ShouldSample(params trace.SamplingParameters) trace.SamplingResult {
    // Increase sampling during high error rates
    errorRate := s.getCurrentErrorRate()
    if errorRate > 0.05 { // 5% error rate
        return trace.SamplingResult{Decision: trace.RecordAndSample}
    }

    return trace.TraceIDRatioBased(0.01).ShouldSample(params)
}
```

### Custom Propagators

#### Baggage Propagation
```go
// Propagate custom metadata
baggageCtx := baggage.ContextWithValues(ctx,
    "user.tier", "premium",
    "request.priority", "high",
)

otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
    propagation.TraceContext{},
    propagation.Baggage{}, // Propagate custom baggage
))
```

### Integration Testing

#### Trace Validation
```go
func TestTracingEndToEnd(t *testing.T) {
    // Make request
    resp, err := client.Get("/api/v1/users")
    require.NoError(t, err)

    // Wait for spans to be exported
    time.Sleep(100 * time.Millisecond)

    // Query Jaeger API for traces
    traces := queryTraces(t, "user-service", "http.get_users")

    // Validate trace structure
    require.Len(t, traces, 1)
    trace := traces[0]

    // Check spans exist
    assertSpanExists(t, trace, "http.get_users")
    assertSpanExists(t, trace, "db.select_users")

    // Validate attributes
    span := findSpan(trace, "db.select_users")
    assert.Equal(t, "SELECT", span.Attributes["db.operation"])
}
```

---

*Next: [Tools & Libraries](tools.md) | [Monitoring & Troubleshooting](monitoring.md)*</content>
</xai:function_call">write">
<parameter name="filePath">docs/tracing/README.md