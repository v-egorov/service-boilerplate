# Implementation Details

This document provides code-level details of the distributed tracing implementation.

## 📁 Code Structure

### Common Tracing Package

```
common/tracing/
├── tracer.go      # Global tracer provider setup
└── middleware.go  # HTTP request tracing middleware
```

### Service Integration Points

```
services/{service-name}/
├── cmd/main.go           # Tracing initialization
├── config.yaml           # Tracing configuration
└── internal/
    └── handlers/         # Business logic (auto-instrumented)
```

## 🔧 Core Implementation

### Tracer Setup (`common/tracing/tracer.go`)

#### `InitTracer` Function

```go
func InitTracer(cfg config.TracingConfig) (*trace.TracerProvider, error) {
    // 1. Early return if tracing disabled
    if !cfg.Enabled {
        return nil, nil
    }

    // 2. Create OTLP HTTP exporter
    exp, err := otlptracehttp.New(context.Background(),
        otlptracehttp.WithEndpointURL(cfg.CollectorURL),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
    }

    // 3. Create resource with service metadata
    res, err := resource.New(context.Background(),
        resource.WithAttributes(
            semconv.ServiceNameKey.String(cfg.ServiceName),
            semconv.ServiceVersionKey.String("1.0.0"),
        ),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create resource: %w", err)
    }

    // 4. Configure batch span processor
    bsp := trace.NewBatchSpanProcessor(exp)

    // 5. Create tracer provider with sampling
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exp),  // Batch processor
        trace.WithResource(res), // Service metadata
        trace.WithSampler(trace.ParentBased(
            trace.TraceIDRatioBased(cfg.SamplingRate),
        )),
    )

    // 6. Set global tracer provider
    otel.SetTracerProvider(tp)

    // 7. Configure global text map propagator
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{}, // W3C traceparent
        propagation.Baggage{},      // W3C tracestate
    ))

    return tp, nil
}
```

#### `ShutdownTracer` Function

```go
func ShutdownTracer(tp *trace.TracerProvider) error {
    if tp == nil {
        return nil
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := tp.Shutdown(ctx); err != nil {
        return fmt.Errorf("failed to shutdown tracer provider: %w", err)
    }

    return nil
}
```

### HTTP Middleware (`common/tracing/middleware.go`)

#### `HTTPMiddleware` Function

```go
func HTTPMiddleware(serviceName string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extract trace context from request headers
        ctx := otel.GetTextMapPropagator().Extract(
            c.Request.Context(),
            propagation.HeaderCarrier(c.Request.Header),
        )

        // 2. Get tracer instance
        tracer := otel.Tracer(serviceName)

        // 3. Create span name from HTTP method and path
        spanName := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)

        // 4. Start span with extracted context
        ctx, span := tracer.Start(ctx, spanName)
        defer span.End()

        // 5. Add HTTP attributes
        span.SetAttributes(
            attribute.String("http.method", c.Request.Method),
            attribute.String("http.url", c.Request.URL.String()),
            attribute.String("http.user_agent", c.Request.UserAgent()),
            attribute.String("http.scheme", c.Request.URL.Scheme),
            attribute.String("http.host", c.Request.Host),
        )

        // 6. Add service-specific attributes
        span.SetAttributes(
            attribute.String("service.name", serviceName),
            attribute.String("http.route", c.Request.URL.Path),
        )

        // 7. Store trace context in request
        c.Request = c.Request.WithContext(ctx)

        // 8. Process request
        c.Next()

        // 9. Add response attributes
        span.SetAttributes(
            attribute.Int("http.status_code", c.Writer.Status()),
        )

        // 10. Record errors if any
        if len(c.Errors) > 0 {
            span.SetStatus(codes.Error, c.Errors.String())
            for _, err := range c.Errors {
                span.RecordError(err)
            }
        } else {
            span.SetStatus(codes.Ok, "")
        }
    }
}
```

## 🔄 Service Integration

### Main Application Setup

#### Tracing Initialization in `main.go`

```go
func main() {
    // Load configuration
    cfg, err := config.Load(".")
    if err != nil {
        fmt.Printf("Failed to load config: %v\n", err)
        os.Exit(1)
    }

    // Initialize logger
    logger := logging.NewLogger(logging.Config{...})

    // Initialize tracing
    tracerProvider, err := tracing.InitTracer(cfg.Tracing)
    if err != nil {
        logger.Warn("Failed to initialize tracing", err)
    } else if tracerProvider != nil {
        defer func() {
            if err := tracing.ShutdownTracer(tracerProvider); err != nil {
                logger.Error("Failed to shutdown tracer", err)
            }
        }()
    }

    // ... database, services setup ...

    // Setup Gin router
    router := gin.New()

    // Middleware (order matters)
    router.Use(gin.Recovery())
    router.Use(corsMiddleware())
    router.Use(serviceLogger.RequestResponseLogger())

    // Add tracing middleware (after logger, before routes)
    if cfg.Tracing.Enabled {
        router.Use(tracing.HTTPMiddleware(cfg.Tracing.ServiceName))
    }

    // ... route definitions ...

    // Start server
    srv := &http.Server{...}
    go func() {
        logger.Info(fmt.Sprintf("Starting %s service on %s", cfg.App.Name, srv.Addr))
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatal("Failed to start service", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logger.Info("Shutting down service...")

    // Shutdown server
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        logger.Error("Service forced to shutdown", err)
    }

    logger.Info("Service exited")
}
```

### API Gateway Proxy Implementation

#### Trace Header Injection (`api-gateway/internal/handlers/gateway.go`)

```go
func (h *GatewayHandler) ProxyRequest(serviceName string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get target service URL
        targetURL, exists := h.registry.GetServiceURL(serviceName)
        if !exists {
            c.JSON(503, gin.H{"error": "Service unavailable"})
            return
        }

        // Create reverse proxy
        proxy := &httputil.ReverseProxy{
            Director: func(req *http.Request) {
                // Standard proxy director setup
                req.URL.Scheme = "http"
                req.URL.Host = targetURL.Host
                req.URL.Path = c.Request.URL.Path
                req.URL.RawQuery = c.Request.URL.RawQuery

                // Copy headers
                for key, values := range c.Request.Header {
                    for _, value := range values {
                        req.Header.Add(key, value)
                    }
                }

                // CRITICAL: Inject trace context headers
                ctx := c.Request.Context()
                otel.GetTextMapPropagator().Inject(
                    ctx,
                    propagation.HeaderCarrier(req.Header),
                )

                // Handle request body
                if c.Request.Body != nil {
                    bodyBytes, err := io.ReadAll(c.Request.Body)
                    if err != nil {
                        h.logger.WithError(err).Error("Failed to read request body")
                        return
                    }
                    req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
                    c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
                }
            },
            ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
                h.logger.WithError(err).Error("Proxy error")
                c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
            },
        }

        // Serve the proxied request
        proxy.ServeHTTP(c.Writer, c.Request)
    }
}
```

### Service Client Propagation

#### HTTP Client with Trace Headers (`services/auth-service/internal/client/user_client.go`)

```go
func (c *UserClient) CreateUser(ctx context.Context, req *CreateUserRequest) (*UserData, error) {
    url := fmt.Sprintf("%s/users", c.baseURL)

    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }

    httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    httpReq.Header.Set("Content-Type", "application/json")

    // CRITICAL: Inject trace context into outgoing request
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(httpReq.Header))

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        c.logger.WithError(err).Error("Failed to call user service")
        return nil, fmt.Errorf("failed to call user service: %w", err)
    }
    defer resp.Body.Close()

    // ... response handling ...
}
```

## ⚙️ Configuration Structure

### Tracing Configuration (`common/config/config.go`)

```go
type TracingConfig struct {
    Enabled      bool    `mapstructure:"enabled"`
    ServiceName  string  `mapstructure:"service_name"`
    CollectorURL string  `mapstructure:"collector_url"`
    SamplingRate float64 `mapstructure:"sampling_rate"`
}
```

### Service Configuration Examples

#### API Gateway (`api-gateway/config.yaml`)
```yaml
tracing:
  enabled: true
  service_name: "api-gateway"
  collector_url: "http://jaeger:4318/v1/traces"
  sampling_rate: 1.0
```

#### Auth Service (`services/auth-service/config.yaml`)
```yaml
tracing:
  enabled: true
  service_name: "auth-service"
  collector_url: "http://jaeger:4318/v1/traces"
  sampling_rate: 1.0
```

#### User Service (`services/user-service/config.yaml`)
```yaml
tracing:
  enabled: true
  service_name: "user-service"
  collector_url: "http://jaeger:4318/v1/traces"
  sampling_rate: 1.0
```

## 🔄 Context Propagation Patterns

### Synchronous Operations
```go
func (h *Handler) ProcessRequest(c *gin.Context) {
    // Context already contains trace info from middleware
    ctx := c.Request.Context()

    // Pass context to service layer
    result, err := h.service.ProcessBusinessLogic(ctx, data)

    // Context flows through all layers automatically
}
```

### Asynchronous Operations
```go
func (h *Handler) ProcessAsync(c *gin.Context) {
    ctx := c.Request.Context()

    go func() {
        // Create child span for async operation
        tracer := otel.Tracer("service-name")
        asyncCtx, span := tracer.Start(ctx, "async.operation")
        defer span.End()

        // Process in background with trace context
        h.processAsync(asyncCtx, data)
    }()
}
```

### Database Operations
```go
func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
    // Context passed from handler contains trace info
    // Database operations are automatically traced if using instrumented driver
    return r.db.GetUser(ctx, id)
}
```

## 🐛 Error Handling

### Span Error Recording
```go
func operation(ctx context.Context) error {
    tracer := otel.Tracer("service-name")
    ctx, span := tracer.Start(ctx, "operation.name")
    defer span.End()

    err := doSomething()
    if err != nil {
        // Record error on span
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    span.SetStatus(codes.Ok, "")
    return nil
}
```

### Middleware Error Handling
```go
// In middleware - errors are automatically recorded
if len(c.Errors) > 0 {
    span.SetStatus(codes.Error, c.Errors.String())
    for _, err := range c.Errors {
        span.RecordError(err)
    }
}
```

## 📊 Performance Optimizations

### Batch Processing
- OTLP exporter automatically batches spans
- Reduces network overhead
- Configurable batch size and timeout

### Sampling
- Parent-based sampling respects parent decisions
- Trace ID ratio sampling for root spans
- Configurable rates per environment

### Resource Management
- Graceful shutdown ensures span delivery
- Context timeouts prevent hanging
- Memory-bounded queues

---

*Next: [Service Template Integration](template.md) | [Developer Guide](developer-guide.md)*</content>
</xai:function_call">docs/tracing/template.md