# 🔧 Middleware Architecture

This document outlines the middleware architecture of the service-boilerplate project, including authentication, logging, tracing, and request processing patterns used across all services.

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture Layers](#architecture-layers)
- [API Gateway Middleware](#api-gateway-middleware)
- [Common Shared Middleware](#common-shared-middleware)
- [Logging Infrastructure](#logging-infrastructure)
- [Tracing Integration](#tracing-integration)
- [Service Implementation Patterns](#service-implementation-patterns)
- [Future Service Guidelines](#future-service-guidelines)

## Overview

The middleware architecture follows a **layered approach** with different middleware implementations serving specific purposes in the microservice ecosystem:

- **API Gateway Layer**: Lightweight routing and basic request processing
- **Service Layer**: Business logic authentication and detailed logging
- **Shared Components**: Common utilities used across all services

## Architecture Layers

### Layered Middleware Stack

```
┌─────────────────┐
│   API Gateway   │  ← Lightweight routing, basic auth
├─────────────────┤
│  Service Layer  │  ← JWT validation, business logic
├─────────────────┤
│ Shared Commons  │  ← Tracing, logging, audit
└─────────────────┘
```

### Design Principles

1. **Separation of Concerns**: Each layer has distinct responsibilities
2. **Progressive Enhancement**: Basic → Advanced authentication
3. **Shared Components**: Reduce duplication across services
4. **Observability**: Tracing and logging at all layers

## API Gateway Middleware

**Location:** `api-gateway/internal/middleware/`

**Purpose:** Handle request routing, basic authentication, and gateway-specific concerns.

### Components

#### 1. Authentication Middleware (`auth.go`)

**Purpose:** Simple authentication for development/demo purposes.

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        // Demo implementation - accepts any Bearer token
        if !strings.HasPrefix(authHeader, "Bearer ") {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

**Usage:** Development environments, not production.

#### 2. Request Processing Middleware (`auth.go`)

**CORS Middleware:**
```go
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }

        c.Next()
    }
}
```

**Request ID Middleware:**
```go
func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)

        c.Next()
    }
}
```

#### 3. Logging Middleware (`logging.go`)

**Purpose:** Gateway-level request/response logging with metrics collection.

**Key Features:**
- Request/response logging with structured fields
- Metrics collection for monitoring
- Response size and status tracking
- User context extraction

### Gateway Middleware Stack

```go
// API Gateway middleware order
router.Use(gin.Recovery())
router.Use(corsMiddleware())                    // CORS headers
router.Use(middleware.RequestIDMiddleware())    // Request ID generation
router.Use(commonMiddleware.JWTMiddleware())    // Optional auth
router.Use(gatewayLogger.RequestResponseLogger()) // Gateway logging
```

## Common Shared Middleware

**Location:** `common/middleware/`

**Purpose:** Production-grade middleware shared across all services.

### JWT Authentication Middleware (`auth.go`)

**Purpose:** Full JWT token validation with revocation checking.

#### Key Features

**Token Validation:**
- RSA and HMAC JWT support
- Token expiration checking
- Signature verification
- Claims extraction

**Revocation Checking:**
```go
type TokenRevocationChecker interface {
    IsTokenRevoked(tokenString string) bool
}
```

**User Context:**
- Extracts user ID, email, roles from JWT claims
- Sets context variables for handlers
- Integrates with tracing

#### Implementation

```go
func JWTMiddleware(jwtSecret interface{}, logger *logrus.Logger, revocationChecker TokenRevocationChecker) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract and validate JWT token
        // Check revocation if checker provided
        // Set user context variables
        // Continue or abort with error
    }
}
```

#### Usage Patterns

**API Gateway (with revocation):**
```go
revocationChecker := &httpTokenRevocationChecker{
    authServiceURL: "http://auth-service:8083",
    logger:         logger.Logger,
}
router.Use(middleware.JWTMiddleware(jwtPublicKey, logger.Logger, revocationChecker))
```

**Internal Services (trust gateway):**
```go
// No revocation checking - trusts gateway validation
router.Use(middleware.JWTMiddleware(nil, logger.Logger, nil))
```

**Direct Exposed Services:**
```go
// Full validation for services that may be directly accessed
router.Use(middleware.JWTMiddleware(jwtPublicKey, logger.Logger, revocationChecker))
```

## Logging Infrastructure

**Location:** `common/logging/`

**Purpose:** Comprehensive logging system with tracing integration and audit capabilities.

### Components

#### 1. Core Logger (`logger.go`)

**Features:**
- Structured JSON logging
- File rotation with lumberjack
- ANSI escape code stripping
- Multiple output targets

#### 2. Service Request Logger (`middleware.go`)

**Purpose:** HTTP request/response logging for all services.

**Key Features:**
- Trace context extraction (trace_id, span_id)
- User context logging (user_id)
- Request/response metrics
- Error tracking

**Standardized Fields:**
```go
logEntry := logger.WithFields(logrus.Fields{
    "timestamp":     time.Now().Format(time.RFC3339),
    "service":       serviceName,
    "request_id":    requestID,
    "trace_id":      traceID,
    "span_id":       spanID,
    "user_id":       userID,
    "method":        c.Request.Method,
    "path":          c.Request.URL.Path,
    "status":        statusCode,
    "duration_ms":   duration.Milliseconds(),
    "user_agent":    userAgent,
    "ip":            clientIP,
    "request_size":  requestSize,
    "response_size": responseSize,
})
```

#### 3. Audit Logger (`audit.go`)

**Purpose:** Security event logging for compliance and monitoring.

**Features:**
- Actor-target separation
- Security event categorization
- Trace correlation
- Structured audit trails

### Logging Levels

- **INFO**: Successful operations
- **WARN**: Client errors (4xx)
- **ERROR**: Server errors (5xx)
- **DEBUG**: Development/troubleshooting

## Tracing Integration

**Location:** `common/tracing/`

**Purpose:** Distributed tracing across all middleware layers.

### HTTP Middleware (`middleware.go`)

**Purpose:** Creates spans for incoming HTTP requests.

```go
func HTTPMiddleware(serviceName string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract trace context from headers
        ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

        // Create span for HTTP request
        tracer := otel.Tracer(serviceName)
        ctx, span := tracer.Start(ctx, fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path))

        // Set span attributes
        span.SetAttributes(
            attribute.String("http.method", c.Request.Method),
            attribute.String("http.url", c.Request.URL.String()),
            attribute.String("service.name", serviceName),
        )

        // Store context for child spans
        c.Request = c.Request.WithContext(ctx)

        c.Next()

        // Record response attributes
        span.SetAttributes(attribute.Int("http.status_code", c.Writer.Status()))
        span.SetStatus(codes.Ok, "")
    }
}
```

### Database Tracing (`database/tracing.go`)

**Purpose:** Instrument database operations with spans.

```go
func TraceDBInsert(ctx context.Context, table string, query string, fn func(ctx context.Context) error) error {
    return TraceDBOperation(ctx, DBOpInsert, table, query, fn)
}
```

## Service Implementation Patterns

### Standard Service Middleware Stack

```go
func main() {
    // Basic middleware
    router.Use(gin.Recovery())
    router.Use(corsMiddleware())

    // Tracing (creates HTTP spans)
    if cfg.Tracing.Enabled {
        router.Use(tracing.HTTPMiddleware(cfg.Tracing.ServiceName))
    }

    // Authentication (JWT validation)
    router.Use(middleware.JWTMiddleware(jwtSecret, logger.Logger, revocationChecker))

    // Logging (request/response with tracing context)
    router.Use(serviceLogger.RequestResponseLogger())

    // Routes...
}
```

### Service-Specific Variations

#### API Gateway Pattern
```go
// Additional gateway middleware
router.Use(middleware.RequestIDMiddleware())    // Generate request IDs
router.Use(commonMiddleware.RequireAuth())      // Route-level auth
```

#### Auth Service Pattern
```go
// Full JWT validation with revocation checking
revocationChecker := &dbTokenRevocationChecker{db: db}
router.Use(middleware.JWTMiddleware(jwtPublicKey, logger.Logger, revocationChecker))
```

#### Internal Service Pattern
```go
// Trust gateway validation
router.Use(middleware.JWTMiddleware(nil, logger.Logger, nil))
```

## Future Service Guidelines

### Adding New Middleware

#### 1. Assess Requirements
- **Gateway Level**: Request routing, basic validation
- **Service Level**: Business logic validation, detailed logging
- **Shared Level**: Common utilities, tracing integration

#### 2. Follow Patterns
- **Context Propagation**: Pass context through middleware chain
- **Error Handling**: Use appropriate HTTP status codes
- **Logging**: Include trace context in all log entries
- **Metrics**: Record relevant metrics for monitoring

#### 3. Integration Points
- **Tracing**: Create spans for significant operations
- **Logging**: Use standardized field names
- **Metrics**: Record performance and error metrics
- **Audit**: Log security-relevant events

### Middleware Ordering Best Practices

```go
// Recommended order
router.Use(gin.Recovery())                    // 1. Error recovery
router.Use(corsMiddleware())                  // 2. CORS headers
router.Use(requestIDMiddleware())             // 3. Request ID
router.Use(tracing.HTTPMiddleware())          // 4. Tracing spans
router.Use(middleware.JWTMiddleware())        // 5. Authentication
router.Use(serviceLogger.RequestResponseLogger()) // 6. Logging
```

### Testing Middleware

#### Unit Testing
```go
func TestJWTMiddleware(t *testing.T) {
    // Test valid tokens
    // Test invalid tokens
    // Test missing tokens
    // Test revoked tokens
}
```

#### Integration Testing
```go
func TestMiddlewareStack(t *testing.T) {
    // Test complete middleware chain
    // Verify context propagation
    // Check logging output
    // Validate tracing spans
}
```

### Performance Considerations

- **Minimal Overhead**: Design middleware to add minimal latency
- **Conditional Execution**: Skip expensive operations when not needed
- **Resource Pooling**: Reuse connections and resources
- **Async Processing**: Offload non-critical operations

## Related Documentation

- [Security Architecture](security-architecture.md) - Authentication and authorization patterns
- [Distributed Tracing](tracing/) - Complete tracing implementation guide
- [Logging System](logging-system.md) - Logging configuration and usage
- [Service Creation Guide](service-creation-guide.md) - How to implement middleware in new services</content>
</xai:function_call<parameter name="filePath">docs/middleware-architecture.md