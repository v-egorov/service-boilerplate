# Tracing Implementation Guide

This guide covers how to implement tracing in services using OpenTelemetry.

---

## Overview

The boilerplate provides three levels of tracing:

1. **HTTP Request Tracing** - Automatic via middleware
2. **Database Operation Tracing** - Via wrapper functions
3. **Business Operation Tracing** - Manual span creation

---

## 1. HTTP Request Tracing

### Setup

```go
// cmd/main.go
import (
    "github.com/v-egorov/service-boilerplate/common/tracing"
)

func main() {
    // Initialize tracer
    if cfg.Tracing.Enabled {
        tp, err := tracing.InitTracer(cfg.Tracing)
        if err != nil {
            logger.Error("Failed to initialize tracer", err)
        }
        defer tp.Shutdown()
    }

    // Add middleware to router
    router.Use(tracing.HTTPMiddleware("service-name"))
}
```

### Configuration

```yaml
# config.yaml
tracing:
  enabled: true
  service_name: "objects-service"
  endpoint: "http://localhost:4318/v1/traces"
  sample_rate: 0.1
```

### Automatic Attributes

The HTTP middleware adds these attributes to each span:
- `http.method`
- `http.url`
- `http.route`
- `http.status_code`
- `http.request.content_length`
- `http.response.content_length`

---

## 2. Database Operation Tracing

### Using Wrapper Functions

```go
// internal/repository/user_repository.go
import (
    "github.com/v-egorov/service-boilerplate/common/database"
)

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
    query := `SELECT id, email FROM app.users WHERE id = $1`

    var user models.User
    err := database.TraceDBQuery(ctx, "users", query, func(ctx context.Context) error {
        return r.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email)
    })
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    return &user, nil
}
```

### Available Wrappers

| Function | Use For |
|----------|---------|
| `TraceDBQuery` | SELECT operations |
| `TraceDBInsert` | INSERT operations |
| `TraceDBUpdate` | UPDATE operations |
| `TraceDBDelete` | DELETE operations |
| `TraceDBOperation` | Generic operations |

### Wrapper Signature

```go
func TraceDBQuery(ctx context.Context, table, query string, fn func(ctx context.Context) error) error {
    span, ctx := otel.Tracer("database").Start(ctx, "db.query."+table)
    defer span.End()

    span.SetAttributes(
        attribute.String("db.system", "postgresql"),
        attribute.String("db.operation", "SELECT"),
        attribute.String("db.name", table),
    )

    err := fn(ctx)
    if err != nil {
        span.RecordError(err)
    }

    return err
}
```

---

## 3. Business Operation Tracing

### Manual Span Creation

```go
// internal/services/user_service.go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
)

func (s *userService) Create(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
    // Start a new span
    span, ctx := otel.Tracer("user-service").Start(ctx, "service.user.Create")
    defer span.End()

    // Add attributes
    span.SetAttributes(
        attribute.String("user.email", req.Email),
    )

    // Validation
    if req.Email == "" {
        span.SetStatus(codes.Error, "validation failed")
        span.RecordError(fmt.Errorf("email is required"))
        return nil, ErrInvalidInput
    }

    // Business logic
    user, err := s.repo.Create(ctx, user)
    if err != nil {
        span.SetStatus(codes.Error, "create failed")
        span.RecordError(err)
        return nil, err
    }

    // Add result attributes
    span.SetAttributes(
        attribute.String("user.id", user.ID.String()),
    )

    return user, nil
}
```

### Repository-Level Tracing

```go
// internal/repository/user_repository.go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
)

func (r *userRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
    span, ctx := otel.Tracer("user-repository").Start(ctx, "repository.user.Create")
    defer span.End()

    span.SetAttributes(
        attribute.String("user.id", user.ID.String()),
        attribute.String("user.email", user.Email),
    )

    // ... database operation
    // Database wrapper will create child span automatically
}
```

---

## 4. Trace Context Propagation

### Outgoing HTTP Calls

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/propagation"
)

func (c *AuthClient) CallExternalAPI(ctx context.Context) error {
    // Get propagator
    propagator := otel.GetTextMapPropagator()

    // Create HTTP request with trace context
    req, _ := http.NewRequestWithContext(ctx, "GET", "http://external-api/users", nil)

    // Inject trace context into headers
    propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

    // Make request
    resp, err := c.client.Do(req)
    // ...
}
```

### Receiving Trace Context

The HTTP middleware automatically extracts trace context from incoming requests using W3C Trace Context format.

---

## 5. Best Practices

### Do's

✅ Always end spans with `defer span.End()`  
✅ Record errors with `span.RecordError(err)`  
✅ Set informative span names: `service.method.operation`  
✅ Add relevant attributes for filtering  
✅ Use child spans for nested operations  

### Don'ts

❌ Don't create spans without ending them  
❌ Don't set error status then return nil  
❌ Don't add PII/sensitive data to spans  
❌ Don't create too many spans (performance)  

---

## 6. Viewing Traces

### Jaeger

The boilerplate outputs to Jaeger-compatible OTLP endpoint:

```bash
# Access Jaeger UI
open http://localhost:16686

# Search for spans by:
# - Service name: "objects-service"
# - Operation name: "service.user.Create"
# - Attribute: user.email = "test@example.com"
```

### Common Issues

| Issue | Solution |
|-------|----------|
| Spans not appearing | Check OTLP endpoint is running |
| Missing attributes | Ensure span is ended before accessing |
| Performance impact | Reduce sample rate in config |
| Trace breaks across services | Ensure context propagation (Inject/Extract) |

---

## 7. Example: Full Service Call Trace

```
HTTP GET /api/v1/users/123
├── HTTPmiddleware (http.server)
│   └── db.users.Query (database)
│   └── service.user.GetByID (service)
│       └── repository.user.GetByID (repository)
│           └── db.users.Query (database)
└── Response
```

Each level creates a child span, allowing you to trace the full request path through all layers.
