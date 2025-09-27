# Service Template Integration

This document explains how distributed tracing is automatically integrated into new services created with the service template.

## 📋 Template Structure

### Files Modified in Template

```
templates/service-template/
├── cmd/main.go           # Tracing initialization added
├── config.yaml           # Tracing configuration added
└── .air.toml            # Common directory watching (already present)
```

### Automatic Integration Points

1. **Configuration**: Tracing section in `config.yaml`
2. **Initialization**: Tracer setup in `main.go`
3. **Middleware**: HTTP tracing middleware in router
4. **Hot Reload**: Common package watching in `.air.toml`

## 🔧 Template Changes

### Configuration Addition (`config.yaml`)

```yaml
# Added to templates/service-template/config.yaml
tracing:
  enabled: true
  service_name: "SERVICE_NAME"  # Replaced during service creation
  collector_url: "http://jaeger:4318/v1/traces"
  sampling_rate: 1.0
```

### Main Application Changes (`cmd/main.go`)

#### Import Addition
```go
import (
    // ... existing imports ...
    "github.com/v-egorov/service-boilerplate/common/tracing"
    // ENTITY_IMPORT_HANDLERS
    // ENTITY_IMPORT_REPOSITORY
    // ENTITY_IMPORT_SERVICES
)
```

#### Tracing Initialization
```go
func main() {
    // ... existing setup ...

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

    // ... database setup ...

    // Setup Gin router
    router := gin.New()

    // Middleware
    router.Use(gin.Recovery())
    router.Use(corsMiddleware())
    router.Use(serviceLogger.RequestResponseLogger())

    // Add tracing middleware
    if cfg.Tracing.Enabled {
        router.Use(tracing.HTTPMiddleware(cfg.Tracing.ServiceName))
    }

    // ... routes and server setup ...
}
```

### Hot Reload Configuration (`.air.toml`)

```toml
# Already present in templates/service-template/.air.toml
[build]
include_dir = ["common", "services/SERVICE_NAME"]
```

## 🚀 Service Creation Process

### Using `create-service.sh`

```bash
./scripts/create-service.sh my-new-service 8085
```

### What Happens Automatically

1. **Template Copy**: `templates/service-template/` copied to `services/my-new-service/`
2. **Placeholder Replacement**:
   - `SERVICE_NAME` → `my-new-service`
   - `PORT` → `8085`
3. **Configuration Update**: Tracing service name updated to `my-new-service`
4. **Docker Integration**: Service added to `docker-compose.yml` and `docker-compose.override.yml`
5. **Makefile Targets**: Build, run, test targets added
6. **API Gateway Registration**: Service registered in API Gateway service registry

### Generated Configuration

#### `services/my-new-service/config.yaml`
```yaml
app:
  name: "my-new-service"
  version: "1.0.0"
  environment: "development"

# ... other config ...

tracing:
  enabled: true
  service_name: "my-new-service"        # ✅ Auto-generated
  collector_url: "http://jaeger:4318/v1/traces"
  sampling_rate: 1.0
```

#### `services/my-new-service/cmd/main.go`
```go
// Tracing import added
import "github.com/v-egorov/service-boilerplate/common/tracing"

// Tracing initialization added
tracerProvider, err := tracing.InitTracer(cfg.Tracing)
// ... error handling ...

// Tracing middleware added
if cfg.Tracing.Enabled {
    router.Use(tracing.HTTPMiddleware(cfg.Tracing.ServiceName))
}
```

## 🔄 Integration Verification

### Build Verification

```bash
# Build the new service
make build-my-new-service

# Expected output: successful compilation
```

### Runtime Verification

```bash
# Start the service
make run-my-new-service

# Check logs for tracing initialization
# Expected: No tracing-related errors
```

### Jaeger Verification

```bash
# Make a request to the service
curl http://localhost:8085/health

# Check Jaeger UI at http://localhost:16686
# Expected: Traces from my-new-service appear
```

## 📊 What Gets Traced Automatically

### HTTP Endpoints
- All REST endpoints automatically instrumented
- Health checks, metrics, and business endpoints
- Request/response tracing with HTTP attributes

### Generated Routes (Example)
```go
// These routes are automatically traced:
GET    /health
GET    /ready
GET    /status
GET    /ping
GET    /api/v1/status
GET    /api/v1/ping
GET    /api/v1/metrics
GET    /api/v1/alerts
POST   /api/v1/entities      # Business logic
GET    /api/v1/entities/:id
PUT    /api/v1/entities/:id
PATCH  /api/v1/entities/:id
DELETE /api/v1/entities/:id
GET    /api/v1/entities
```

## 🔧 Manual Customization

### After Service Creation

#### 1. Adjust Sampling Rate
```yaml
# services/my-new-service/config.yaml
tracing:
  sampling_rate: 0.5  # 50% sampling instead of 100%
```

#### 2. Change Collector URL
```yaml
# For production deployment
tracing:
  collector_url: "http://jaeger-collector.company.com:4318/v1/traces"
```

#### 3. Disable Tracing
```yaml
# If tracing not needed
tracing:
  enabled: false
```

### Adding Custom Tracing

#### Business Logic Instrumentation
```go
// In handlers or services
func (h *MyHandler) CustomOperation(c *gin.Context) {
    tracer := otel.Tracer("my-new-service")
    ctx, span := tracer.Start(c.Request.Context(), "custom.operation")
    defer span.End()

    // Your business logic
    result := h.doSomething(ctx)

    c.JSON(200, result)
}
```

#### Database Operation Tracing
```go
func (r *MyRepository) GetEntity(ctx context.Context, id string) (*Entity, error) {
    tracer := otel.Tracer("my-new-service")
    ctx, span := tracer.Start(ctx, "db.get_entity")
    defer span.End()

    span.SetAttributes(
        attribute.String("db.operation", "SELECT"),
        attribute.String("entity.id", id),
    )

    // Database query
    return r.db.GetEntity(ctx, id)
}
```

## 🐳 Docker Integration

### Automatic Docker Configuration

#### `docker-compose.yml` Addition
```yaml
my-new-service:
  build:
    context: ..
    dockerfile: services/my-new-service/Dockerfile
  image: ${MY_NEW_SERVICE_SERVICE_IMAGE}
  container_name: ${MY_NEW_SERVICE_SERVICE_CONTAINER}
  ports:
    - "${MY_NEW_SERVICE_SERVICE_PORT:-8085}:${MY_NEW_SERVICE_SERVICE_PORT:-8085}"
  environment:
    - APP_ENV=${APP_ENV:-production}
    - SERVER_PORT=${MY_NEW_SERVICE_SERVICE_PORT:-8085}
    # ... database and tracing environment variables
  depends_on:
    postgres:
      condition: service_healthy
  networks:
    service-network:
      aliases:
        - my-new-service
```

#### `docker-compose.override.yml` Addition
```yaml
my-new-service:
  build:
    context: ..
    dockerfile: services/my-new-service/Dockerfile.dev
  environment:
    - APP_ENV=development
    - LOGGING_LEVEL=debug
  volumes:
    - ../services/my-new-service:/app/services/my-new-service
    - ../common:/app/common  # Enables hot reload for tracing changes
```

### Environment Variables

#### `.env` Additions
```bash
# Service configuration
MY_NEW_SERVICE_SERVICE_NAME=my-new-service
MY_NEW_SERVICE_SERVICE_PORT=8085
MY_NEW_SERVICE_SERVICE_IMAGE=docker-my-new-service
MY_NEW_SERVICE_SERVICE_CONTAINER=service-boilerplate-my-new-service
MY_NEW_SERVICE_SERVICE_TMP_VOLUME=service-boilerplate-my-new-service-tmp
MY_NEW_SERVICE_LOGS_VOLUME=service-boilerplate-my-new-service-logs
MY_NEW_SERVICE_SCHEMA=my_new_service
```

## 📈 Scaling Considerations

### Multiple Instances
- Each instance gets unique service name
- Traces distinguish between instances
- Load balancing preserves trace context

### Service Dependencies
- API Gateway automatically routes with trace headers
- Service-to-service calls need manual client instrumentation
- Database operations may need custom tracing

## 🚨 Troubleshooting Template Issues

### Service Won't Start
```bash
# Check tracing configuration
cat services/my-new-service/config.yaml

# Verify tracing import
grep -n "tracing" services/my-new-service/cmd/main.go
```

### Traces Not Appearing
```bash
# Check Jaeger connectivity
curl http://localhost:4318/v1/traces

# Verify service name in config
grep "service_name" services/my-new-service/config.yaml

# Check service logs for tracing errors
make logs-my-new-service
```

### Hot Reload Not Working
```bash
# Verify .air.toml configuration
cat services/my-new-service/.air.toml

# Check that common directory is included
grep "common" services/my-new-service/.air.toml
```

## 📚 Related Documentation

- **[Developer Guide](developer-guide.md)**: How to add custom tracing
- **[Configuration](configuration.md)**: Environment-specific setup
- **[Monitoring](monitoring.md)**: Jaeger UI usage and debugging
- **[Best Practices](best-practices.md)**: Guidelines and recommendations

---

*Next: [Developer Guide](developer-guide.md) | [Configuration](configuration.md)*</content>
</xai:function_call">docs/tracing/developer-guide.md