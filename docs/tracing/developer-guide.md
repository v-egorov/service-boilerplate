# Developer Guide: Instrumenting New Endpoints

This guide shows developers how to add distributed tracing to new endpoints, business logic, and external integrations.

## 🎯 Automatic Instrumentation

### HTTP Endpoints (Automatic)

Most HTTP endpoints are automatically instrumented via middleware. No additional code needed:

```go
// These endpoints are automatically traced:
router.GET("/api/v1/users", handler.ListUsers)           // ✅ Auto-traced
router.POST("/api/v1/users", handler.CreateUser)         // ✅ Auto-traced
router.GET("/api/v1/users/:id", handler.GetUser)         // ✅ Auto-traced
router.PUT("/api/v1/users/:id", handler.UpdateUser)      // ✅ Auto-traced
router.DELETE("/api/v1/users/:id", handler.DeleteUser)   // ✅ Auto-traced
```

**What gets captured automatically:**
- HTTP method, URL, status code
- Request duration
- User agent, headers
- Errors and exceptions

### Auth Service Example

The auth service demonstrates comprehensive tracing for security-critical operations:

#### Repository Layer Database Tracing

```go
// Automatic database operation tracing
func (r *AuthRepository) CreateAuthToken(ctx context.Context, token *models.AuthToken) error {
    query := `INSERT INTO auth_service.auth_tokens...`
    return database.TraceDBInsert(ctx, "auth_tokens", query, func(ctx context.Context) error {
        _, err := r.db.Exec(ctx, query, token.ID, token.UserID, token.TokenHash, token.TokenType, token.ExpiresAt)
        return err
    })
}
```

#### Service Layer Business Logic Tracing

```go
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest, ipAddress, userAgent string) (*models.TokenResponse, error) {
    tracer := otel.Tracer("auth-service")

    ctx, span := tracer.Start(ctx, "auth.login",
        trace.WithAttributes(
            attribute.String("user.email", req.Email),
            attribute.String("client.ip", ipAddress),
            attribute.String("auth.operation", "login"),
        ))
    defer span.End()

    // Business logic with error recording
    userLogin, err := s.userClient.GetUserWithPasswordByEmail(ctx, req.Email)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "Failed to get user from user service")
        return nil, fmt.Errorf("invalid credentials")
    }

    // Success attributes
    span.SetAttributes(
        attribute.String("user.id", userID.String()),
        attribute.Int("auth.tokens_created", 2),
        attribute.Bool("auth.session_created", true),
    )
    span.SetStatus(codes.Ok, "Login successful")

    return tokenResponse, nil
}
```

#### Handler Layer Context Propagation

```go
func (h *AuthHandler) Login(c *gin.Context) {
    // Extract trace information from request context
    span := trace.SpanFromContext(c.Request.Context())
    traceID := span.SpanContext().TraceID().String()
    spanID := span.SpanContext().SpanID().String()

    // Use trace context in audit logging
    h.auditLogger.LogAuthAttempt("", c.GetHeader("X-Request-ID"), c.ClientIP(), c.GetHeader("User-Agent"),
        req.Email, traceID, spanID, false, "Invalid request format")
}
```

## 🔧 Custom Span Creation

### Basic Custom Spans

For complex business logic that spans multiple operations:

```go
func (h *UserHandler) ComplexBusinessLogic(c *gin.Context) {
    // Get tracer from global provider
    tracer := otel.Tracer("user-service")

    // Create custom span
    ctx, span := tracer.Start(c.Request.Context(), "user.registration.process")
    defer span.End()

    // Add business-specific attributes
    span.SetAttributes(
        attribute.String("user.email", c.PostForm("email")),
        attribute.String("registration.type", "email"),
    )

    // Your business logic
    user, err := h.createUser(ctx, userData)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "User creation failed")
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    // Add success attributes
    span.SetAttributes(
        attribute.String("user.id", user.ID.String()),
        attribute.Bool("registration.success", true),
    )

    span.SetStatus(codes.Ok, "User registered successfully")
    c.JSON(201, gin.H{"user": user})
}
```

### Database Operations

```go
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
    tracer := otel.Tracer("user-service")

    ctx, span := tracer.Start(ctx, "db.get_user_by_email")
    defer span.End()

    // Add database-specific attributes
    span.SetAttributes(
        attribute.String("db.operation", "SELECT"),
        attribute.String("db.table", "users"),
        attribute.String("db.query_type", "by_email"),
        attribute.String("user.email", email),
    )

    // Database query
    user, err := r.db.QueryRowContext(ctx,
        "SELECT id, email, first_name, last_name FROM users WHERE email = $1",
        email,
    ).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName)

    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "Database query failed")
        return nil, err
    }

    span.SetAttributes(attribute.Bool("user.found", true))
    span.SetStatus(codes.Ok, "")
    return &user, nil
}
```

### External API Calls

```go
func (c *PaymentClient) ChargeCard(ctx context.Context, amount float64, cardToken string) error {
    tracer := otel.Tracer("payment-service")

    ctx, span := tracer.Start(ctx, "payment.charge_card")
    defer span.End()

    span.SetAttributes(
        attribute.String("payment.provider", "stripe"),
        attribute.Float64("payment.amount", amount),
        attribute.String("payment.currency", "USD"),
    )

    // Create HTTP request with trace context
    req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/charge", body)
    if err != nil {
        span.RecordError(err)
        return err
    }

    // Inject trace headers
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

    // Make the call
    resp, err := c.httpClient.Do(req)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "HTTP request failed")
        return err
    }
    defer resp.Body.Close()

    span.SetAttributes(
        attribute.Int("http.status_code", resp.StatusCode),
        attribute.Bool("payment.success", resp.StatusCode == 200),
    )

    if resp.StatusCode != 200 {
        span.SetStatus(codes.Error, "Payment failed")
        return fmt.Errorf("payment failed with status %d", resp.StatusCode)
    }

    span.SetStatus(codes.Ok, "Payment processed successfully")
    return nil
}
```

## 🔄 Async Operations

### Goroutines with Trace Context

```go
func (h *BatchHandler) ProcessBatch(c *gin.Context) {
    tracer := otel.Tracer("batch-service")

    // Parent span for batch operation
    ctx, parentSpan := tracer.Start(c.Request.Context(), "batch.process")
    defer parentSpan.End()

    parentSpan.SetAttributes(
        attribute.Int("batch.size", len(batch.Items)),
        attribute.String("batch.type", "user_import"),
    )

    // Process items concurrently
    var wg sync.WaitGroup
    for _, item := range batch.Items {
        wg.Add(1)
        go func(item BatchItem) {
            defer wg.Done()

            // Child span for each item
            _, childSpan := tracer.Start(ctx, "batch.process_item")
            defer childSpan.End()

            childSpan.SetAttributes(
                attribute.String("item.id", item.ID),
                attribute.String("item.type", item.Type),
            )

            err := h.processItem(ctx, item)
            if err != nil {
                childSpan.RecordError(err)
                childSpan.SetStatus(codes.Error, err.Error())
            } else {
                childSpan.SetStatus(codes.Ok, "")
            }
        }(item)
    }

    wg.Wait()
    parentSpan.SetStatus(codes.Ok, "Batch processing completed")
}
```

### Worker Pools

```go
func (w *WorkerPool) StartWorkers(ctx context.Context) {
    tracer := otel.Tracer("worker-service")

    for i := 0; i < w.numWorkers; i++ {
        go func(workerID int) {
            // Worker span
            workerCtx, workerSpan := tracer.Start(ctx, "worker.lifecycle")
            defer workerSpan.End()

            workerSpan.SetAttributes(
                attribute.Int("worker.id", workerID),
                attribute.String("worker.type", "batch_processor"),
            )

            for {
                select {
                case job := <-w.jobQueue:
                    // Job processing span
                    _, jobSpan := tracer.Start(workerCtx, "worker.process_job")
                    jobSpan.SetAttributes(
                        attribute.String("job.id", job.ID),
                        attribute.String("job.type", job.Type),
                    )

                    err := w.processJob(job)
                    if err != nil {
                        jobSpan.RecordError(err)
                        jobSpan.SetStatus(codes.Error, err.Error())
                    } else {
                        jobSpan.SetStatus(codes.Ok, "")
                    }
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

## 📊 Custom Attributes and Events

### Business Metrics

```go
func (h *AnalyticsHandler) TrackUserAction(c *gin.Context) {
    tracer := otel.Tracer("analytics-service")

    ctx, span := tracer.Start(c.Request.Context(), "analytics.track_action")
    defer span.End()

    action := c.PostForm("action")
    userID := c.GetString("user_id")

    // Add business attributes
    span.SetAttributes(
        attribute.String("analytics.action", action),
        attribute.String("user.id", userID),
        attribute.String("analytics.category", "engagement"),
        attribute.String("analytics.source", "web"),
    )

    // Record custom event
    span.AddEvent("action_tracked", trace.WithAttributes(
        attribute.String("event.type", "user_interaction"),
        attribute.String("event.timestamp", time.Now().Format(time.RFC3339)),
    ))

    // Process analytics
    err := h.recordAnalytics(ctx, userID, action)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "Analytics recording failed")
        c.JSON(500, gin.H{"error": "Failed to track action"})
        return
    }

    span.SetStatus(codes.Ok, "Action tracked successfully")
    c.JSON(200, gin.H{"status": "tracked"})
}
```

### Performance Metrics

```go
func (h *SearchHandler) SearchProducts(c *gin.Context) {
    tracer := otel.Tracer("search-service")

    startTime := time.Now()
    ctx, span := tracer.Start(c.Request.Context(), "search.products")
    defer span.End()

    query := c.Query("q")
    limit := parseInt(c.DefaultQuery("limit", "20"))

    span.SetAttributes(
        attribute.String("search.query", query),
        attribute.Int("search.limit", limit),
        attribute.String("search.index", "products"),
    )

    // Perform search
    results, total, err := h.searchEngine.Search(ctx, query, limit)
    searchDuration := time.Since(startTime)

    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "Search failed")
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    // Add performance metrics
    span.SetAttributes(
        attribute.Int("search.results_count", len(results)),
        attribute.Int("search.total_matches", total),
        attribute.Float64("search.duration_ms", float64(searchDuration.Nanoseconds())/1e6),
        attribute.Bool("search.cache_hit", false),
    )

    span.SetStatus(codes.Ok, "Search completed")
    c.JSON(200, gin.H{
        "results": results,
        "total": total,
        "query": query,
    })
}
```

## 🔗 Context Propagation

### Service-to-Service Calls

```go
func (c *OrderClient) CreateOrder(ctx context.Context, order *OrderRequest) (*Order, error) {
    tracer := otel.Tracer("order-service")

    ctx, span := tracer.Start(ctx, "order.create")
    defer span.End()

    span.SetAttributes(
        attribute.String("order.customer_id", order.CustomerID),
        attribute.Float64("order.total", order.Total),
        attribute.Int("order.items_count", len(order.Items)),
    )

    // Prepare request
    jsonData, _ := json.Marshal(order)
    req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/orders", bytes.NewBuffer(jsonData))
    if err != nil {
        span.RecordError(err)
        return nil, err
    }

    // Critical: Inject trace context
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

    // Make call
    resp, err := c.httpClient.Do(req)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "HTTP request failed")
        return nil, err
    }
    defer resp.Body.Close()

    span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

    if resp.StatusCode != 201 {
        span.SetStatus(codes.Error, "Order creation failed")
        return nil, fmt.Errorf("order creation failed: %d", resp.StatusCode)
    }

    var createdOrder Order
    json.NewDecoder(resp.Body).Decode(&createdOrder)

    span.SetAttributes(attribute.String("order.id", createdOrder.ID))
    span.SetStatus(codes.Ok, "Order created successfully")

    return &createdOrder, nil
}
```

### Message Queues

```go
func (p *MessageProducer) PublishMessage(ctx context.Context, topic string, message interface{}) error {
    tracer := otel.Tracer("message-service")

    ctx, span := tracer.Start(ctx, "message.publish")
    defer span.End()

    span.SetAttributes(
        attribute.String("message.topic", topic),
        attribute.String("message.type", fmt.Sprintf("%T", message)),
    )

    // Inject trace context into message headers/metadata
    carrier := propagation.MapCarrier{}
    otel.GetTextMapPropagator().Inject(ctx, carrier)

    // Convert carrier to message headers
    headers := make(map[string]string)
    carrier.ForeachKey(func(key, value string) bool {
        headers[key] = value
        return true
    })

    // Publish message with trace headers
    err := p.publisher.Publish(topic, message, headers)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "Message publish failed")
        return err
    }

    span.SetStatus(codes.Ok, "Message published successfully")
    return nil
}
```

## 🛠️ Utility Functions

### Span Helper Functions

```go
// Common span creation helper
func createSpan(ctx context.Context, tracer trace.Tracer, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
    ctx, span := tracer.Start(ctx, name)
    span.SetAttributes(attrs...)
    return ctx, span
}

// Error handling helper
func handleSpanError(span trace.Span, err error, message string) {
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, message)
    } else {
        span.SetStatus(codes.Ok, "")
    }
}

// Database operation helper
func traceDBOperation(ctx context.Context, operation, table string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
    tracer := otel.Tracer("db-service")
    ctx, span := tracer.Start(ctx, fmt.Sprintf("db.%s", operation))

    baseAttrs := []attribute.KeyValue{
        attribute.String("db.operation", operation),
        attribute.String("db.table", table),
    }

    span.SetAttributes(append(baseAttrs, attrs...)...)
    return ctx, span
}
```

### Usage Examples

```go
func (r *UserRepository) UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) error {
    ctx, span := traceDBOperation(ctx, "update", "users",
        attribute.String("user.id", userID),
        attribute.Int("update.fields_count", len(updates)),
    )
    defer span.End()

    err := r.db.UpdateUser(ctx, userID, updates)
    handleSpanError(span, err, "User update failed")

    return err
}

func (h *UserHandler) RegisterUser(c *gin.Context) {
    ctx, span := createSpan(c.Request.Context(), otel.Tracer("user-service"), "user.register",
        attribute.String("registration.method", "email"),
    )
    defer span.End()

    // Registration logic
    user, err := h.registerUser(ctx, registrationData)
    if err != nil {
        handleSpanError(span, err, "User registration failed")
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    span.SetAttributes(attribute.String("user.id", user.ID))
    span.SetStatus(codes.Ok, "User registered successfully")
    c.JSON(201, gin.H{"user": user})
}
```

## 📋 Best Practices Summary

### Span Naming
- Use lowercase with dots: `user.registration.process`
- Include operation type: `db.select`, `http.post`, `cache.get`
- Be descriptive but concise

### Attributes
- Use semantic conventions when possible
- Add business-relevant attributes
- Include IDs, counts, and status information
- Avoid sensitive data in attributes

### Error Handling
- Always record errors with `span.RecordError(err)`
- Set appropriate status codes
- Include descriptive error messages

### Context Propagation
- Always pass context through function calls
- Inject trace headers in HTTP clients
- Use context in database operations

### Performance
- Create spans for meaningful operations only
- Use appropriate sampling rates
- Avoid excessive attributes

---

*Next: [Configuration](configuration.md) | [Best Practices](best-practices.md)*</content>
</xai:function_call">docs/tracing/configuration.md