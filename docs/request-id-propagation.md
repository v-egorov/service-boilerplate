# Request ID Propagation Guide

This document explains how request ID propagation works in the microservice architecture and provides implementation guidelines for new services and endpoints.

## 🎯 Overview

Request ID propagation ensures that every request can be traced across all services in the architecture. This enables:

- **Correlation**: Link related log entries across services
- **Debugging**: Trace request flow through the entire system
- **Monitoring**: Track request lifecycle and performance
- **Support**: Provide unique identifiers for user issues

## 🔄 Request Flow Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Client    │───▶│ API Gateway │───▶│ Auth Service│───▶│ User Service│
│             │    │             │    │             │    │             │
│             │    │ 1. Generate │    │ 3. Extract  │    │ 5. Extract  │
│             │    │    Request  │    │    & Prop. │    │    & Log    │
│             │    │       ID    │    │             │    │             │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                        │                       │                       │
                        │ 2. Set X-Request-ID   │ 4. Set X-Request-ID   │
                        │    Header             │    Header             │
                        ▼                       ▼                       ▼
```

## 📋 Implementation Scenarios

### Scenario 1: Proxied Requests (Through Auth Service)

For endpoints that require authentication and are proxied through the auth service:

#### 1. API Gateway Configuration
```go
// In api-gateway/cmd/main.go
// RequestIDMiddleware automatically generates and sets X-Request-ID
router.Use(middleware.RequestIDMiddleware())

// Auth-protected routes
api := router.Group("/api")
api.Use(middleware.AuthMiddleware())
{
    users := api.Group("/v1/users")
    {
        users.POST("", gatewayHandler.ProxyRequest("user-service"))
        users.GET("/:id", gatewayHandler.ProxyRequest("user-service"))
    }
}
```

#### 2. Auth Service Setup
```go
// In auth-service/cmd/main.go
// Extract request ID from incoming headers and store in context
requestIDMiddleware := func() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.GetHeader("X-Request-ID")
        if requestID != "" {
            ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
            c.Request = c.Request.WithContext(ctx)
        }
        c.Next()
    }
}

router.Use(requestIDMiddleware())
```

#### 3. Auth Service Client Calls
```go
// In auth-service/internal/client/user_client.go
func (c *UserClient) GetUserByEmail(ctx context.Context, email string) (*UserData, error) {
    // Extract request ID from context and set in HTTP headers
    if requestID, ok := ctx.Value("request_id").(string); ok {
        httpReq.Header.Set("X-Request-ID", requestID)
    }

    // Inject trace context (OpenTelemetry)
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(httpReq.Header))

    // Make HTTP call...
}
```

#### 4. User Service Handlers
```go
// In user-service/internal/handlers/user_handler.go
func (h *UserHandler) GetUser(c *gin.Context) {
    requestID := c.GetHeader("X-Request-ID")

    // Include request_id in all log statements
    h.logger.WithFields(logrus.Fields{
        "request_id": requestID,
        "user_id":    user.ID,
    }).Info("User retrieved successfully")

    // Use standard logger for structured operation logging
    h.standardLogger.UserOperation(requestID, user.ID.String(), "get", true, nil)
}
```

### Scenario 2: Direct Requests (No Authentication Required)

For public endpoints that don't require authentication:

#### 1. API Gateway Configuration
```go
// In api-gateway/cmd/main.go
// RequestIDMiddleware automatically generates and sets X-Request-ID
router.Use(middleware.RequestIDMiddleware())

// Public routes (no auth required)
router.GET("/health", gatewayHandler.ProxyRequest("user-service"))
router.GET("/api/v1/public/data", gatewayHandler.ProxyRequest("data-service"))
```

#### 2. Service Setup (No Auth Service Involvement)
```go
// In user-service/cmd/main.go
// ServiceRequestLogger automatically extracts request_id from headers
router.Use(serviceLogger.RequestResponseLogger())

// Handlers automatically get request_id in logs
func (h *HealthHandler) LivenessHandler(c *gin.Context) {
    requestID := c.GetHeader("X-Request-ID")

    h.logger.WithFields(logrus.Fields{
        "request_id": requestID,
        "service":    "user-service",
    }).Debug("Health check performed")
}
```

## 🛠️ Implementation Checklist

### For New Services (Proxied Through Auth)

#### 1. Service Middleware Setup
```go
// In your-service/cmd/main.go
requestIDMiddleware := func() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.GetHeader("X-Request-ID")
        if requestID != "" {
            ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
            c.Request = c.Request.WithContext(ctx)
        }
        c.Next()
    }
}

router.Use(requestIDMiddleware())
```

#### 2. Client HTTP Calls
```go
// In your-service/internal/client/other_client.go
func (c *OtherClient) MakeCall(ctx context.Context, data interface{}) error {
    httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/endpoint", body)

    // Extract and set request ID
    if requestID, ok := ctx.Value("request_id").(string); ok {
        httpReq.Header.Set("X-Request-ID", requestID)
    }

    // Inject OpenTelemetry trace context
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(httpReq.Header))

    return c.httpClient.Do(httpReq)
}
```

#### 3. Handler Implementation
```go
// In your-service/internal/handlers/your_handler.go
func (h *YourHandler) YourMethod(c *gin.Context) {
    requestID := c.GetHeader("X-Request-ID")

    // All log statements must include request_id
    h.logger.WithFields(logrus.Fields{
        "request_id": requestID,
        "operation":  "your_operation",
    }).Info("Operation completed")

    // Use standard logger for operations
    h.standardLogger.YourOperation(requestID, entityID, "create", true, nil)
}
```

### For New Services (Direct/Public Access)

#### 1. Service Middleware Setup
```go
// In your-service/cmd/main.go
// ServiceRequestLogger automatically handles request_id extraction and logging
serviceLogger := logging.NewServiceRequestLogger(logger.Logger, "your-service")
router.Use(serviceLogger.RequestResponseLogger())
```

#### 2. Handler Implementation
```go
// In your-service/internal/handlers/your_handler.go
func (h *YourHandler) PublicEndpoint(c *gin.Context) {
    requestID := c.GetHeader("X-Request-ID")

    // Include request_id in all custom logs
    h.logger.WithFields(logrus.Fields{
        "request_id": requestID,
        "endpoint":   "public",
    }).Info("Public endpoint accessed")
}
```

## 🔍 Verification Steps

### 1. Check API Gateway Logs
```bash
docker-compose logs api-gateway | grep "Proxying request"
# Should show: request_id=uuid-here service=user-service
```

### 2. Check Service Logs
```bash
docker-compose logs user-service | grep "request_id="
# Should show consistent request_id across all log entries for the same request
```

### 3. Test End-to-End Flow
```bash
# Make a request and check logs across all services
curl -X GET "http://localhost:8080/api/v1/users" -H "Authorization: Bearer <token>"
```

### 4. Verify Log Correlation
- Same `request_id` should appear in API Gateway, Auth Service, and User Service logs
- Request flow should be traceable from client to final service

## 🚨 Common Issues & Solutions

### Issue: Missing Request ID in Service Logs
**Symptom**: Service logs show `request_id=` (empty)
**Cause**: Service not extracting request ID from headers
**Solution**: Ensure handlers extract `c.GetHeader("X-Request-ID")`

### Issue: Inconsistent Request IDs
**Symptom**: Different request IDs in different services
**Cause**: Service generating new request ID instead of using forwarded one
**Solution**: Always use `c.GetHeader("X-Request-ID")`, don't generate new ones

### Issue: Service-to-Service Calls Lose Request ID
**Symptom**: Client service has request ID, but called service doesn't
**Cause**: HTTP client not setting X-Request-ID header
**Solution**: Extract from context and set header: `httpReq.Header.Set("X-Request-ID", requestID)`

### Issue: Context Not Propagated
**Symptom**: `ctx.Value("request_id")` returns nil
**Cause**: Request ID middleware not storing in context
**Solution**: Ensure middleware does: `ctx := context.WithValue(c.Request.Context(), "request_id", requestID)`

## 📋 Best Practices

### 1. Always Extract Request ID
```go
func (h *Handler) AnyMethod(c *gin.Context) {
    requestID := c.GetHeader("X-Request-ID") // Always do this first
    // ... rest of handler
}
```

### 2. Include Request ID in All Logs
```go
h.logger.WithFields(logrus.Fields{
    "request_id": requestID,  // Always include
    "user_id":    userID,     // Business context
    "operation":  "create",   // Operation details
}).Info("User created")
```

### 3. Propagate in HTTP Clients
```go
// Always set both request ID and trace context
if requestID, ok := ctx.Value("request_id").(string); ok {
    httpReq.Header.Set("X-Request-ID", requestID)
}
otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(httpReq.Header))
```

### 4. Use Standard Logger for Operations
```go
// Structured operation logging
h.standardLogger.UserOperation(requestID, entityID, "create", success, err)
```

### 5. Handle Missing Request IDs Gracefully
```go
requestID := c.GetHeader("X-Request-ID")
if requestID == "" {
    requestID = "unknown" // Fallback for debugging
}
```

## 🔗 Related Documentation

- [Distributed Tracing Overview](../tracing/overview.md)
- [Service Creation Guide](../service-creation-guide.md)
- [Logging System Documentation](../logging-system.md)

## 📞 Support

When implementing request ID propagation:

1. **Test thoroughly**: Verify request IDs appear in all service logs
2. **Check middleware order**: Request ID middleware must run before handlers
3. **Monitor logs**: Use consistent request IDs for debugging
4. **Update documentation**: Add new endpoints to this guide

---

*Request ID propagation ensures observability across the entire microservice architecture.*</content>
</xai:function_call">Write file to docs/request-id-propagation.md