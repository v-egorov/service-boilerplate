# Authentication & Audit Logging Troubleshooting Guide

This guide helps diagnose and resolve common issues with JWT authentication, user context population, and audit logging in the service-boilerplate project.

## Overview

The authentication and audit logging system consists of:
- JWT middleware for token validation and user context
- Three-tier logging architecture (application, standard, audit)
- Distributed tracing integration
- Grafana/Loki integration for log analysis

## Quick Diagnosis

### Check Service Health
```bash
# Check all services are running
make health

# Check specific service logs
docker logs service-boilerplate-auth-service
docker logs service-boilerplate-user-service
```

### Verify JWT Configuration
```bash
# Check JWT public key is set
docker exec service-boilerplate-auth-service env | grep JWT

# Test authentication endpoint
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}'
```

## Common Issues

### 1. JWT Token Validation Fails

**Symptoms:**
- 401 Unauthorized responses
- User context not populated in handlers
- Audit logs show empty `user_id`

**Possible Causes:**
- JWT public key not configured
- Token expired or malformed
- Wrong signing algorithm
- Clock skew between services

**Solutions:**

1. **Verify JWT Public Key:**
   ```bash
   # Check environment variable
   docker exec service-boilerplate-user-service env | grep JWT_PUBLIC_KEY

   # Ensure key is base64 encoded
   echo $JWT_PUBLIC_KEY | base64 -d
   ```

2. **Check Token Format:**
   ```bash
   # Decode JWT token (replace TOKEN with actual token)
   echo "TOKEN" | jq -R 'split(".") | .[0],.[1] | @base64d | fromjson'
   ```

3. **Verify Service Configuration:**
   ```go
   // In cmd/main.go - ensure JWT middleware is configured
   jwtPublicKey := os.Getenv("JWT_PUBLIC_KEY")
   if jwtPublicKey != "" {
       decodedKey, err := base64.StdEncoding.DecodeString(jwtPublicKey)
       if err != nil {
           log.Fatal("Invalid JWT public key")
       }
       router.Use(middleware.JWTMiddleware(decodedKey, logger.Logger))
   }
   ```

### 2. User Context Not Populated

**Symptoms:**
- `middleware.GetAuthenticatedUserID(c)` returns empty string
- Audit logs missing actor identification
- User-specific operations fail

**Possible Causes:**
- JWT middleware not applied to route
- Middleware order incorrect
- Token missing from request

**Solutions:**

1. **Check Middleware Order:**
   ```go
   // Correct order in cmd/main.go
   router.Use(gin.Recovery())
   router.Use(corsMiddleware())
   router.Use(requestIDMiddleware())
   if cfg.Tracing.Enabled {
       router.Use(tracing.HTTPMiddleware(cfg.Tracing.ServiceName))
   }
   router.Use(middleware.JWTMiddleware(jwtPublicKey, logger.Logger))  // Before logging middleware
   router.Use(serviceLogger.RequestResponseLogger())
   ```

2. **Verify Route Protection:**
   ```go
   // Ensure routes require authentication
   protected := v1.Group("/protected")
   protected.Use(middleware.RequireAuth())
   ```

3. **Check Request Headers:**
   ```bash
   # Include Authorization header
   curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
        http://localhost:8080/api/v1/users
   ```

### 3. Audit Logs Not Appearing

**Symptoms:**
- No audit events in logs
- Grafana dashboards missing audit data
- Security events not logged

**Possible Causes:**
- Audit logger not initialized
- Audit methods not called
- Log level filtering audit logs
- Loki ingestion issues

**Solutions:**

1. **Verify Audit Logger Initialization:**
   ```go
   // In handlers - ensure audit logger is available
   type Handler struct {
       logger       *logrus.Logger
       standardLogger *logging.StandardLogger
       auditLogger   *logging.AuditLogger  // Must be initialized
   }
   ```

2. **Check Audit Method Calls:**
   ```go
   // Ensure audit methods are called in handlers
   func (h *UserHandler) CreateUser(c *gin.Context) {
       actorUserID := middleware.GetAuthenticatedUserID(c)
       // ... business logic ...

       h.auditLogger.LogUserCreation(
           actorUserID, c.GetHeader("X-Request-ID"), user.ID.String(),
           c.ClientIP(), c.GetHeader("User-Agent"),
           traceID, spanID, true, "")
   }
   ```

3. **Verify Log Levels:**
   ```bash
   # Audit logs use warn/error levels - ensure not filtered
   docker exec service-boilerplate-user-service env | grep LOGGING_LEVEL
   ```

4. **Check Loki Ingestion:**
   ```bash
   # Verify Promtail is running
   docker ps | grep promtail

   # Check Promtail logs
   docker logs service-boilerplate-promtail
   ```

### 4. Trace Correlation Issues

**Symptoms:**
- Trace ID/Span ID missing from audit logs
- Jaeger traces incomplete
- Request correlation broken

**Possible Causes:**
- Tracing middleware not applied
- Context not propagated
- Span extraction failing

**Solutions:**

1. **Verify Tracing Middleware:**
   ```go
   // Ensure tracing middleware is applied before JWT middleware
   if cfg.Tracing.Enabled {
       router.Use(tracing.HTTPMiddleware(cfg.Tracing.ServiceName))
   }
   ```

2. **Check Trace Extraction:**
   ```go
   // In handlers - extract trace information
   span := trace.SpanFromContext(c.Request.Context())
   traceID := span.SpanContext().TraceID().String()
   spanID := span.SpanContext().SpanID().String()
   ```

3. **Verify Jaeger Configuration:**
   ```bash
   # Check Jaeger endpoint
   docker exec service-boilerplate-jaeger curl http://localhost:16686/api/services
   ```

### 5. Grafana Queries Not Working

**Symptoms:**
- Dashboards show no data
- Log queries return empty results
- User ID/Entity ID fields missing

**Possible Causes:**
- Loki data source misconfigured
- Query syntax incorrect
- Log format issues

**Solutions:**

1. **Verify Loki Data Source:**
   ```bash
   # Check Grafana data sources
   curl http://localhost:3000/api/datasources \
        -H "Authorization: Bearer YOUR_GRAFANA_TOKEN"
   ```

2. **Test Loki Queries:**
   ```bash
   # Query audit logs
   curl "http://localhost:3100/loki/api/v1/query?query={service=\"user-service\"}" | jq
   ```

3. **Check Log Format:**
   ```json
   // Ensure audit logs have required fields
   {
     "time": "2025-09-27T10:30:00Z",
     "level": "warn",
     "event_type": "user_creation",
     "user_id": "user-123",
     "entity_id": "user-456",
     "service": "user-service",
     "trace_id": "abc123",
     "span_id": "def456"
   }
   ```

### 6. Authentication Middleware Not Working

**Symptoms:**
- Routes not protected
- Public endpoints accessible without auth
- Role-based access failing

**Possible Causes:**
- Middleware not applied
- Role configuration incorrect
- Auth bypass for health endpoints

**Solutions:**

1. **Check Route Groups:**
   ```go
   // Public routes (no auth required)
   router.GET("/health", healthHandler.LivenessHandler)
   router.POST("/api/v1/auth/login", authHandler.Login)

   // Protected routes
   protected := router.Group("/api/v1")
   protected.Use(middleware.RequireAuth())
   ```

2. **Verify Role Requirements:**
   ```go
   // Role-based access
   admin := protected.Group("/admin")
   admin.Use(middleware.RequireRole("admin"))
   ```

3. **Test Authentication Flow:**
   ```bash
   # 1. Login to get token
   TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"admin@example.com","password":"password"}' | jq -r '.token')

   # 2. Use token for protected endpoint
   curl -H "Authorization: Bearer $TOKEN" \
        http://localhost:8080/api/v1/users
   ```

## Debugging Tools

### Log Analysis Commands

```bash
# View real-time logs with filtering
docker logs -f service-boilerplate-user-service | grep audit

# Search for specific user operations
docker logs service-boilerplate-user-service | jq 'select(.event_type == "user_creation")'

# Check for authentication failures
docker logs service-boilerplate-auth-service | grep "authentication_failed"
```

### Database Queries

```bash
# Connect to database
make db-connect

# Check user sessions
SELECT * FROM auth_service.user_sessions WHERE expires_at > NOW();

# Verify user data
SELECT id, email, created_at FROM user_service.users LIMIT 5;
```

### Network Debugging

```bash
# Test service connectivity
curl http://localhost:8080/health
curl http://localhost:8081/health  # user-service
curl http://localhost:8082/health  # auth-service

# Check service discovery
docker exec service-boilerplate-api-gateway nslookup user-service
```

## Advanced Troubleshooting

### JWT Token Inspection

```bash
# Decode JWT header
echo "TOKEN" | cut -d'.' -f1 | base64 -d | jq

# Decode JWT payload
echo "TOKEN" | cut -d'.' -f2 | base64 -d | jq

# Verify signature (requires jwt-cli or similar)
jwt decode TOKEN --key PUBLIC_KEY
```

### Audit Log Analysis

```bash
# Count audit events by type
docker logs service-boilerplate-user-service 2>&1 | \
  jq -r 'select(.event_type) | .event_type' | sort | uniq -c

# Find failed operations
docker logs service-boilerplate-user-service 2>&1 | \
  jq 'select(.result == "failure")'
```

### Performance Issues

**Symptoms:**
- Slow authentication
- Audit logging impacting performance
- High memory usage

**Solutions:**

1. **JWT Validation Optimization:**
   - Cache public keys
   - Use faster crypto libraries
   - Implement token blacklisting

2. **Audit Logging Performance:**
   - Use async logging
   - Batch audit writes
   - Implement log buffering

3. **Memory Usage:**
   - Monitor goroutine leaks
   - Check for context leaks
   - Profile memory usage

## Configuration Validation

### Environment Variables Checklist

```bash
# Required for authentication
JWT_PUBLIC_KEY=base64_encoded_key
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=24h

# Required for logging
LOGGING_LEVEL=info
LOGGING_FORMAT=json
LOGGING_OUTPUT=file
LOGGING_DUAL_OUTPUT=true

# Required for tracing
TRACING_ENABLED=true
TRACING_SERVICE_NAME=user-service
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
```

### Service Dependencies

```yaml
# docker-compose.yml dependencies
depends_on:
  postgres:
    condition: service_healthy
  jaeger:
    condition: service_started
  loki:
    condition: service_started
```

## Best Practices

### Development
- Use debug logging level for troubleshooting
- Enable dual output for console + file logs
- Test authentication flows with curl/Postman

### Production
- Use structured JSON logging
- Implement log retention policies
- Monitor authentication failure rates
- Set up alerting for security events

### Security
- Rotate JWT keys regularly
- Implement token blacklisting
- Log all authentication attempts
- Monitor for suspicious patterns

## Support

If issues persist:

1. **Check Existing Issues:** Search project repository for similar problems
2. **Gather Diagnostics:**
   ```bash
   # Collect system information
   make health > health.log
   docker-compose logs > all_logs.log
   ```
3. **Create Issue:** Include logs, configuration, and reproduction steps
4. **Community Support:** Check documentation and examples

## Quick Reference

### Key Commands
```bash
# Health checks
make health
curl http://localhost:8080/health

# Log inspection
docker logs service-boilerplate-user-service
make logs

# Authentication testing
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}'

# Database access
make db-connect
```

### Common Error Codes
- `401`: Authentication required/token invalid
- `403`: Insufficient permissions/role missing
- `500`: Internal server error (check logs)

This guide covers the most common authentication and audit logging issues. For service-specific problems, refer to individual service documentation and logs.