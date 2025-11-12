# 🔐 Security Architecture

This document outlines the security architecture of the service-boilerplate project, including authentication, authorization, token management, and service exposure guidelines.

## 📋 Table of Contents

- [Overview](#overview)
- [API Gateway Security Model](#api-gateway-security-model)
- [JWT Token Management](#jwt-token-management)
- [TokenRevocationChecker Pattern](#tokenrevocationchecker-pattern)
- [Service Exposure Guidelines](#service-exposure-guidelines)
- [Development vs Production](#development-vs-production)
- [Implementation Examples](#implementation-examples)
- [Security Best Practices](#security-best-practices)

## Overview

The service-boilerplate implements a microservice architecture with centralized authentication and authorization through an API Gateway. Security is enforced at the gateway level, with internal services operating under the assumption that requests have already been validated.

### Key Security Components

- **API Gateway**: Single entry point for all external requests, enforces authentication and token revocation
- **Auth Service**: Manages user authentication, JWT token generation, and token validation
- **JWT Middleware**: Validates JWT tokens and extracts user context
- **TokenRevocationChecker**: Interface for checking if tokens have been revoked

## API Gateway Security Model

### Single Entry Point Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   External      │────│   API Gateway   │────│   Auth Service  │
│   Clients       │    │   (Port 8080)   │    │   (Port 8083)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ├─────────────────┐
                              │   User Service  │
                              │   (Port 8081)   │
                              └─────────────────┘
```

**Security Flow:**

1. All external requests go through API Gateway (port 8080)
2. API Gateway validates JWT tokens and checks revocation status
3. Validated requests are proxied to internal services
4. Internal services trust the gateway's validation

### Why This Architecture?

- **Centralized Security**: All authentication logic in one place
- **Reduced Complexity**: Internal services don't need complex auth logic
- **Performance**: Token validation happens once at the gateway
- **Security**: Revocation checking ensures compromised tokens are immediately invalidated

## JWT Token Management

### Token Lifecycle

1. **Registration/Login**: User credentials → Auth Service → JWT tokens issued
2. **Request Flow**: Client → API Gateway → Token validation → Service proxy
3. **Revocation**: Logout or security event → Token marked invalid → Future requests rejected
4. **Refresh**: Expired access tokens → Refresh token → New access token issued

### Token Types

- **Access Token**: Short-lived (1 hour), used for API access
- **Refresh Token**: Long-lived (7 days), used to get new access tokens
- **JWT Claims**: User ID, email, roles, token type, expiration

### Token Storage

- **Access Tokens**: Stateless, validated using public key cryptography
- **Refresh Tokens**: Stored in database with revocation tracking
- **Revoked Tokens**: Marked in database, checked on each request

### JWT Key Rotation

The system implements automatic JWT key rotation for enhanced security:

- **Automatic Rotation**: Keys rotated every 30 days by default
- **Manual Rotation**: Admin-initiated rotation via API endpoint
- **Key Overlap**: Old keys remain valid during transition period (60 minutes)
- **Audit Logging**: All rotation operations are logged with actor identification
- **Health Monitoring**: Rotation status included in service health checks

**Security Benefits:**

- Limits impact of key compromise to 30-day windows
- Ensures cryptographic key freshness
- Provides emergency rotation capabilities
- Maintains audit trail for compliance

See [JWT Key Rotation](jwt-key-rotation.md) for complete operational details.

## TokenRevocationChecker Pattern

### Interface Definition

```go
type TokenRevocationChecker interface {
    IsTokenRevoked(tokenString string) bool
}
```

### When to Implement

**✅ REQUIRED for services that are:**

- Directly exposed to external traffic in production
- API Gateway (enforces revocation for all requests)

**❌ NOT needed for services that are:**

- Internal microservices accessed only through API Gateway
- Development-only direct access services

### Implementation Examples

#### HTTP-based Revocation Checker (API Gateway)

```go
type httpTokenRevocationChecker struct {
    authServiceURL string
    logger         *logrus.Logger
}

func (c *httpTokenRevocationChecker) IsTokenRevoked(tokenString string) bool {
    req, err := http.NewRequest("POST", c.authServiceURL+"/api/v1/auth/validate-token", nil)
    if err != nil {
        return true // Consider revoked if can't check
    }

    req.Header.Set("Authorization", "Bearer "+tokenString)

    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return true // Consider revoked if service unavailable
    }
    defer resp.Body.Close()

    return resp.StatusCode != http.StatusOK
}
```

#### Database-based Revocation Checker (Auth Service)

```go
type dbTokenRevocationChecker struct {
    db     *sql.DB
    logger *logrus.Logger
}

func (c *dbTokenRevocationChecker) IsTokenRevoked(tokenString string) bool {
    // Parse JWT to get JTI claim
    // Check database for revocation status
    // Return true if revoked
}
```

## Service Exposure Guidelines

### Production Deployment

**🚫 NEVER expose internal services directly in production:**

- User Service (port 8081)
- Auth Service (port 8083)
- Any future microservices

**✅ ONLY expose:**

- API Gateway (port 8080)

### Development Environment

**⚠️ Development-only direct access allowed:**

- For testing and debugging individual services
- Must not be relied upon for production workflows
- Should be documented as development-only features

### Service Creation Guidelines

When creating new services:

1. **Internal Services**: Do NOT implement TokenRevocationChecker
2. **External Services**: Implement TokenRevocationChecker if directly exposed
3. **API Gateway**: Always implement TokenRevocationChecker
4. **Documentation**: Clearly document service exposure requirements

## Development vs Production

### Development Environment

```yaml
# docker-compose.yml (development)
services:
  user-service:
    ports:
      - "8081:8081" # Direct access for development
  api-gateway:
    ports:
      - "8080:8080" # Main entry point
```

**Development Features:**

- Direct service access for testing
- Detailed logging and debugging
- Hot reload capabilities
- Relaxed security for development workflow

### Production Environment

```yaml
# Production deployment
# Only API Gateway exposed externally
api-gateway:
  ports:
    - "8080:8080"

# Internal services not exposed
user-service:
  # No external ports
auth-service:
  # No external ports
```

**Production Requirements:**

- Only API Gateway accessible from external networks
- Internal services communicate via Docker networks
- All requests must go through API Gateway
- Token revocation enforced at gateway level

## Implementation Examples

### API Gateway Configuration

```go
// api-gateway/cmd/main.go
func main() {
    // JWT middleware with revocation checking
    revocationChecker := &httpTokenRevocationChecker{
        authServiceURL: "http://auth-service:8083",
        logger:         logger.Logger,
    }

    router.Use(commonMiddleware.JWTMiddleware(
        jwtPublicKey,
        logger.Logger,
        revocationChecker, // Enforces revocation
    ))
}
```

### Internal Service Configuration

```go
// services/user-service/cmd/main.go
func main() {
    // JWT middleware without revocation checking
    // (trusts API Gateway validation)
    router.Use(commonMiddleware.JWTMiddleware(
        nil,              // No JWT secret (optional auth)
        logger.Logger,
        nil,              // No revocation checker (internal service)
    ))
}
```

### Service Template for New Services

```go
// For internal services (recommended)
router.Use(commonMiddleware.JWTMiddleware(
    nil, logger, nil, // Optional auth, no revocation checking
))

// For external services (if required)
revocationChecker := &httpTokenRevocationChecker{
    authServiceURL: "http://auth-service:8083",
    logger: logger,
}
router.Use(commonMiddleware.JWTMiddleware(
    jwtPublicKey, logger, revocationChecker, // Full auth + revocation
))
```

## Security Best Practices

### Authentication

- Always use HTTPS in production
- Implement proper password policies
- Use secure token storage on clients
- Implement token refresh logic
- Log authentication failures

### Authorization

- Implement role-based access control (RBAC)
- Use middleware for route protection
- Validate user permissions on sensitive operations
- Implement audit logging for security events

### Token Management

- Use short-lived access tokens (1 hour)
- Implement secure refresh token rotation
- Immediately revoke tokens on logout/security events
- Monitor for token abuse patterns

### Service Security

- Never expose internal services directly in production
- Use network segmentation (Docker networks)
- Implement proper CORS policies
- Validate all input data
- Use parameterized queries to prevent SQL injection

### Monitoring & Alerting

- Monitor authentication failure rates
- Alert on suspicious token usage patterns
- Log all security events with full context
- Implement distributed tracing for security investigations

## Troubleshooting

### Common Security Issues

1. **401 Unauthorized after logout**: Token not properly revoked

   - Check: Token appears in revoked tokens table
   - Solution: Verify logout endpoint called correctly

2. **Internal services accepting invalid tokens**: Missing gateway routing

   - Check: Request going directly to service instead of through gateway
   - Solution: Ensure production deployment doesn't expose internal ports

3. **Token validation failures**: JWT secret mismatch
   - Check: Public key configuration in API Gateway
   - Solution: Verify auth-service public key endpoint accessible

### Debug Commands

```bash
# Check token revocation status
curl -H "Authorization: Bearer <token>" \
  http://localhost:8083/api/v1/auth/validate-token

# Check service exposure (should fail in production)
curl http://localhost:8081/health

# Verify gateway routing
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/v1/users
```

## Related Documentation

- [Authentication API Examples](auth-api-examples.md)
- [Troubleshooting Auth & Logging](troubleshooting-auth-logging.md)
- [Service Creation Guide](service-creation-guide.md)
- [Logging System](logging-system.md)
