# 🔗 Service Integration Patterns

This document outlines best practices and patterns for integrating services with the authentication system, including how to consume auth endpoints, handle authentication failures, and implement secure service-to-service communication.

## 📋 Table of Contents

- [Overview](#overview)
- [Authentication Integration](#authentication-integration)
- [Service-to-Service Communication](#service-to-service-communication)
- [Error Handling Patterns](#error-handling-patterns)
- [Token Management](#token-management)
- [Testing Integration](#testing-integration)
- [Security Best Practices](#security-best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

The service-boilerplate architecture uses a centralized authentication service with the following integration patterns:

- **API Gateway**: Single entry point with JWT validation
- **Internal Services**: Trust gateway validation (no duplicate auth)
- **Service-to-Service**: Secure internal communication patterns
- **External Clients**: Direct authentication with auth service

### Integration Layers

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   External      │────│   API Gateway    │────│   Auth Service  │
│   Clients       │    │   (JWT Validation│    │   (Token Mgmt)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
       │                        │                        │
       └────────────────────────┴────────────────────────┴─────────┐
                                                                 │
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Service A      │────│  Service B       │────│  Service C      │
│  (Internal)     │    │  (Internal)      │    │  (Internal)     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Authentication Integration

### Consuming Auth Endpoints

#### User Registration Integration

```go
// In user service - integrate with auth service registration
func (s *UserService) RegisterUser(ctx context.Context, req *RegisterRequest) (*User, error) {
    // First, call auth service to handle authentication
    authReq := &auth.RegisterRequest{
        Email:     req.Email,
        Password:  req.Password,
        FirstName: req.FirstName,
        LastName:  req.LastName,
    }

    authResp, err := s.authClient.Register(ctx, authReq)
    if err != nil {
        return nil, fmt.Errorf("auth service registration failed: %w", err)
    }

    // Auth service creates user in user service internally
    // Just return the created user info
    return &User{
        ID:        authResp.User.ID,
        Email:     authResp.User.Email,
        FirstName: authResp.User.FirstName,
        LastName:  authResp.User.LastName,
        Roles:     authResp.User.Roles,
        CreatedAt: authResp.User.CreatedAt,
    }, nil
}
```

#### Login Integration

```go
// In API gateway or auth-consuming service
func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    // Call auth service
    authResp, err := h.authClient.Login(c.Request.Context(), &auth.LoginRequest{
        Email:    req.Email,
        Password: req.Password,
    })
    if err != nil {
        h.logger.WithError(err).Warn("Login failed")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    // Return tokens to client
    c.JSON(http.StatusOK, gin.H{
        "access_token":  authResp.AccessToken,
        "refresh_token": authResp.RefreshToken,
        "token_type":    "Bearer",
        "expires_in":    3600,
    })
}
```

#### Token Validation Integration

```go
// In services that need to validate tokens directly
func (h *Handler) ValidateToken(c *gin.Context) {
    authHeader := c.GetHeader("Authorization")
    if authHeader == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
        return
    }

    token := strings.TrimPrefix(authHeader, "Bearer ")

    // Call auth service for validation
    validationReq := &auth.ValidateTokenRequest{Token: token}
    resp, err := h.authClient.ValidateToken(c.Request.Context(), validationReq)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
        return
    }

    // Token is valid, extract user info
    c.Set("user_id", resp.User.ID)
    c.Set("user_email", resp.User.Email)
    c.Set("user_roles", resp.User.Roles)
    c.Next()
}
```

### Auth Service Client

```go
// Auth service client interface
type AuthServiceClient interface {
    Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
    Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
    RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error)
    ValidateToken(ctx context.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error)
    Logout(ctx context.Context, req *LogoutRequest) (*LogoutResponse, error)
    GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error)
}

// HTTP client implementation
type httpAuthClient struct {
    baseURL    string
    httpClient *http.Client
    logger     *logrus.Logger
}

func (c *httpAuthClient) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
    return c.doRequest(ctx, "POST", "/api/v1/auth/register", req)
}

func (c *httpAuthClient) doRequest(ctx context.Context, method, path string, body interface{}) (*RegisterResponse, error) {
    jsonData, err := json.Marshal(body)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Service-Caller", "user-service") // Identify calling service

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("auth service returned %d: %s", resp.StatusCode, string(body))
    }

    var response RegisterResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, err
    }

    return &response, nil
}
```

## Service-to-Service Communication

### Internal Service Authentication

When services communicate internally (bypassing the API gateway), use these patterns:

#### Service Tokens

```go
// Generate service-to-service tokens
func (s *AuthService) GenerateServiceToken(serviceName string) (string, error) {
    claims := &middleware.JWTClaims{
        UserID:    uuid.Nil, // No specific user
        Email:     serviceName + "@internal",
        Roles:     []string{"service", serviceName},
        TokenType: "service",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Subject:   serviceName,
            Issuer:    "auth-service",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    return token.SignedString(s.jwtUtils.GetPrivateKey())
}
```

#### Service Authentication Headers

```go
// Service-to-service call with authentication
func callUserService(ctx context.Context, userID string, serviceToken string) (*User, error) {
    req, err := http.NewRequestWithContext(ctx, "GET",
        "http://user-service:8081/api/v1/internal/users/"+userID, nil)
    if err != nil {
        return nil, err
    }

    // Service authentication headers
    req.Header.Set("Authorization", "Bearer "+serviceToken)
    req.Header.Set("X-Service-Caller", "auth-service")
    req.Header.Set("X-Internal-Call", "true")

    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("user service returned %d: %s", resp.StatusCode, string(body))
    }

    var user User
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, err
    }

    return &user, nil
}
```

#### Service Authentication Middleware

```go
// Middleware for internal service calls
func ServiceAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Check if this is an internal call
        if c.GetHeader("X-Internal-Call") != "true" {
            c.Next()
            return
        }

        // Validate service caller
        serviceCaller := c.GetHeader("X-Service-Caller")
        if serviceCaller == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Service caller not identified"})
            c.Abort()
            return
        }

        // Validate service token
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Service token required"})
            c.Abort()
            return
        }

        token := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := validateServiceToken(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid service token"})
            c.Abort()
            return
        }

        // Verify token is for calling service
        if claims.Subject != serviceCaller {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not authorized for this service"})
            c.Abort()
            return
        }

        c.Set("service_caller", serviceCaller)
        c.Set("internal_call", true)
        c.Next()
    }
}
```

### Circuit Breaker Pattern

```go
// Circuit breaker for auth service calls
type AuthCircuitBreaker struct {
    failureCount int
    lastFailure  time.Time
    state        string // "closed", "open", "half-open"
    mutex        sync.Mutex
}

func (cb *AuthCircuitBreaker) Call(fn func() error) error {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()

    if cb.state == "open" {
        if time.Since(cb.lastFailure) > 60*time.Second {
            cb.state = "half-open"
        } else {
            return errors.New("circuit breaker is open")
        }
    }

    err := fn()
    if err != nil {
        cb.failureCount++
        cb.lastFailure = time.Now()
        if cb.failureCount >= 5 {
            cb.state = "open"
        }
        return err
    }

    // Success - reset circuit breaker
    cb.failureCount = 0
    cb.state = "closed"
    return nil
}
```

## Error Handling Patterns

### Authentication Errors

```go
// Comprehensive error handling for auth operations
func (h *AuthHandler) handleAuthError(c *gin.Context, err error, operation string) {
    var statusCode int
    var errorMsg string

    switch {
    case errors.Is(err, auth.ErrInvalidCredentials):
        statusCode = http.StatusUnauthorized
        errorMsg = "Invalid email or password"
    case errors.Is(err, auth.ErrUserNotFound):
        statusCode = http.StatusNotFound
        errorMsg = "User not found"
    case errors.Is(err, auth.ErrTokenExpired):
        statusCode = http.StatusUnauthorized
        errorMsg = "Token has expired"
    case errors.Is(err, auth.ErrTokenRevoked):
        statusCode = http.StatusUnauthorized
        errorMsg = "Token has been revoked"
    case errors.Is(err, auth.ErrInsufficientPermissions):
        statusCode = http.StatusForbidden
        errorMsg = "Insufficient permissions"
    default:
        h.logger.WithError(err).Error("Unexpected auth error")
        statusCode = http.StatusInternalServerError
        errorMsg = "Authentication service unavailable"
    }

    // Audit log the error
    h.auditLogger.LogAuthFailure(
        middleware.GetAuthenticatedUserID(c),
        c.GetHeader("X-Request-ID"),
        operation,
        c.ClientIP(),
        c.GetHeader("User-Agent"),
        c.GetString("trace_id"),
        c.GetString("span_id"),
        errorMsg,
    )

    c.JSON(statusCode, gin.H{
        "error": errorMsg,
        "type":  "authentication_error",
        "operation": operation,
    })
}
```

### Retry Logic

```go
// Retry logic for auth service calls
func (c *httpAuthClient) callWithRetry(ctx context.Context, method, path string, body interface{}, maxRetries int) (*http.Response, error) {
    var lastErr error

    for attempt := 0; attempt < maxRetries; attempt++ {
        resp, err := c.doRequest(ctx, method, path, body)
        if err == nil {
            return resp, nil
        }

        lastErr = err

        // Check if error is retryable
        if isRetryableError(err) {
            backoff := time.Duration(attempt+1) * time.Second
            c.logger.WithFields(logrus.Fields{
                "attempt": attempt + 1,
                "backoff": backoff,
                "error":   err.Error(),
            }).Warn("Auth service call failed, retrying")

            select {
            case <-time.After(backoff):
                continue
            case <-ctx.Done():
                return nil, ctx.Err()
            }
        }

        // Non-retryable error
        break
    }

    return nil, lastErr
}

func isRetryableError(err error) bool {
    if urlErr, ok := err.(*url.Error); ok {
        return urlErr.Timeout() || urlErr.Temporary()
    }

    if httpErr, ok := err.(*http.Response); ok {
        return httpErr.StatusCode >= 500
    }

    return false
}
```

### Fallback Strategies

```go
// Fallback for auth service unavailability
func (h *AuthHandler) LoginWithFallback(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    // Try primary auth service
    resp, err := h.authClient.Login(c.Request.Context(), &auth.LoginRequest{
        Email:    req.Email,
        Password: req.Password,
    })

    if err != nil {
        h.logger.WithError(err).Warn("Primary auth service failed, trying fallback")

        // Fallback to cached/local validation
        if h.fallbackAuth != nil {
            resp, err = h.fallbackAuth.Login(c.Request.Context(), &auth.LoginRequest{
                Email:    req.Email,
                Password: req.Password,
            })
            if err != nil {
                h.handleAuthError(c, err, "login_fallback")
                return
            }
        } else {
            h.handleAuthError(c, err, "login")
            return
        }
    }

    c.JSON(http.StatusOK, gin.H{
        "access_token":  resp.AccessToken,
        "refresh_token": resp.RefreshToken,
        "token_type":    "Bearer",
        "expires_in":    3600,
    })
}
```

## Token Management

### Token Refresh Integration

```go
// Automatic token refresh in API clients
type AuthenticatedClient struct {
    baseURL       string
    accessToken   string
    refreshToken  string
    tokenMutex    sync.RWMutex
    authClient    AuthServiceClient
    logger        *logrus.Logger
}

func (c *AuthenticatedClient) doAuthenticatedRequest(method, path string, body io.Reader) (*http.Response, error) {
    c.tokenMutex.RLock()
    token := c.accessToken
    c.tokenMutex.RUnlock()

    req, err := http.NewRequest(method, c.baseURL+path, body)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", "Bearer "+token)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }

    // Check if token needs refresh
    if resp.StatusCode == http.StatusUnauthorized {
        resp.Body.Close()

        // Try to refresh token
        if err := c.refreshAccessToken(); err != nil {
            return nil, fmt.Errorf("token refresh failed: %w", err)
        }

        // Retry with new token
        c.tokenMutex.RLock()
        token = c.accessToken
        c.tokenMutex.RUnlock()

        req.Header.Set("Authorization", "Bearer "+token)
        return http.DefaultClient.Do(req)
    }

    return resp, nil
}

func (c *AuthenticatedClient) refreshAccessToken() error {
    c.tokenMutex.Lock()
    defer c.tokenMutex.Unlock()

    refreshResp, err := c.authClient.RefreshToken(context.Background(), &auth.RefreshTokenRequest{
        RefreshToken: c.refreshToken,
    })
    if err != nil {
        return err
    }

    c.accessToken = refreshResp.AccessToken
    c.refreshToken = refreshResp.RefreshToken
    return nil
}
```

### Token Caching

```go
// Token validation caching
type TokenCache struct {
    cache  *bigcache.BigCache
    logger *logrus.Logger
}

func (tc *TokenCache) IsValid(token string) (bool, error) {
    key := "token:" + token

    // Check cache first
    if data, err := tc.cache.Get(key); err == nil {
        return string(data) == "valid", nil
    }

    // Validate with auth service
    valid, err := tc.validateWithAuthService(token)
    if err != nil {
        return false, err
    }

    // Cache result
    status := "invalid"
    if valid {
        status = "valid"
    }

    if err := tc.cache.Set(key, []byte(status)); err != nil {
        tc.logger.WithError(err).Warn("Failed to cache token validation")
    }

    return valid, nil
}
```

## Testing Integration

### Unit Testing Auth Integration

```go
func TestAuthServiceIntegration(t *testing.T) {
    // Mock auth service
    mockAuth := &mockAuthClient{}
    service := NewUserService(mockAuth)

    // Test successful registration
    mockAuth.On("Register", mock.Anything, mock.AnythingOfType("*auth.RegisterRequest")).
        Return(&auth.RegisterResponse{
            User: &auth.User{
                ID:    uuid.New(),
                Email: "test@example.com",
            },
        }, nil)

    user, err := service.RegisterUser(context.Background(), &RegisterRequest{
        Email:    "test@example.com",
        Password: "password123",
    })

    assert.NoError(t, err)
    assert.Equal(t, "test@example.com", user.Email)
    mockAuth.AssertExpectations(t)
}
```

### Integration Testing

```go
func TestAuthServiceEndToEnd(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Start test auth service
    authService := startTestAuthService()
    defer authService.Stop()

    client := NewAuthServiceClient("http://localhost:8083")

    // Test registration
    resp, err := client.Register(context.Background(), &auth.RegisterRequest{
        Email:    "integration@example.com",
        Password: "password123",
    })

    assert.NoError(t, err)
    assert.NotNil(t, resp.User)
    assert.Equal(t, "integration@example.com", resp.User.Email)

    // Test login
    loginResp, err := client.Login(context.Background(), &auth.LoginRequest{
        Email:    "integration@example.com",
        Password: "password123",
    })

    assert.NoError(t, err)
    assert.NotEmpty(t, loginResp.AccessToken)
    assert.NotEmpty(t, loginResp.RefreshToken)
}
```

### Load Testing

```go
func TestAuthServiceLoad(t *testing.T) {
    client := NewAuthServiceClient("http://localhost:8083")

    // Simulate concurrent users
    var wg sync.WaitGroup
    errors := make(chan error, 100)

    for i := 0; i < 50; i++ {
        wg.Add(1)
        go func(userID int) {
            defer wg.Done()

            // Register user
            _, err := client.Register(context.Background(), &auth.RegisterRequest{
                Email:    fmt.Sprintf("loadtest%d@example.com", userID),
                Password: "password123",
            })
            if err != nil {
                errors <- err
                return
            }

            // Login user
            _, err = client.Login(context.Background(), &auth.LoginRequest{
                Email:    fmt.Sprintf("loadtest%d@example.com", userID),
                Password: "password123",
            })
            if err != nil {
                errors <- err
            }
        }(i)
    }

    wg.Wait()
    close(errors)

    // Check for errors
    errorCount := 0
    for err := range errors {
        t.Logf("Load test error: %v", err)
        errorCount++
    }

    if errorCount > 0 {
        t.Errorf("Load test had %d errors", errorCount)
    }
}
```

## Security Best Practices

### Service Authentication

1. **Use Service Tokens**: Never use user tokens for service-to-service calls
2. **Validate Callers**: Always verify the calling service identity
3. **Short-Lived Tokens**: Use short expiration times for service tokens
4. **Secure Storage**: Never log or store service tokens in plain text

### Error Handling

1. **Don't Leak Information**: Avoid exposing internal error details
2. **Log Security Events**: Audit all authentication failures
3. **Rate Limiting**: Implement rate limiting on auth endpoints
4. **Fail Secure**: Default to denying access when in doubt

### Network Security

1. **Internal Networks**: Use internal Docker networks for service communication
2. **TLS**: Use HTTPS for all service-to-service communication
3. **Firewall Rules**: Restrict service access to authorized callers
4. **Monitoring**: Monitor all service-to-service traffic

### Token Security

1. **Regular Rotation**: Implement automatic token rotation
2. **Secure Transmission**: Always use HTTPS for token transmission
3. **Token Validation**: Validate tokens on every request
4. **Revocation**: Implement immediate token revocation capabilities

## Troubleshooting

### Common Integration Issues

#### Auth Service Unavailable

**Symptoms:**
- Registration/login requests fail
- Services can't authenticate users
- 503 Service Unavailable errors

**Solutions:**
- Check auth service health: `curl http://localhost:8083/health`
- Verify network connectivity between services
- Check auth service logs for errors
- Implement circuit breaker pattern

#### Token Validation Failures

**Symptoms:**
- 401 Unauthorized for valid requests
- Intermittent authentication failures
- Services rejecting valid tokens

**Solutions:**
- Verify JWT public key configuration
- Check token expiration times
- Validate token format and claims
- Check for clock skew between services

#### Service-to-Service Auth Issues

**Symptoms:**
- Internal API calls failing with 401
- Services can't communicate internally
- Missing service authentication headers

**Solutions:**
- Verify service tokens are properly generated
- Check X-Service-Caller headers
- Validate service authentication middleware
- Check service token expiration

#### High Latency

**Symptoms:**
- Slow authentication responses
- Service timeouts
- Degraded user experience

**Solutions:**
- Implement token caching
- Use connection pooling
- Add request timeouts
- Monitor auth service performance

### Debug Commands

```bash
# Check service connectivity
curl -v http://localhost:8083/health
curl -v http://localhost:8081/health

# Test authentication flow
curl -X POST http://localhost:8083/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Check service logs
docker logs auth-service
docker logs user-service

# Verify network connectivity
docker exec auth-service ping user-service
docker exec user-service ping auth-service

# Check token validation
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8083/api/v1/auth/validate-token
```

### Monitoring Integration

```go
// Add auth integration metrics
func (c *httpAuthClient) observeRequest(method, path string, start time.Time, err error) {
    duration := time.Since(start).Seconds()

    status := "success"
    if err != nil {
        status = "error"
    }

    // Prometheus metrics
    authRequestsTotal.WithLabelValues(method, path, status).Inc()
    authRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// Prometheus metrics
var (
    authRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "auth_client_requests_total",
            Help: "Total number of auth service requests",
        },
        []string{"method", "endpoint", "status"},
    )

    authRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "auth_client_request_duration_seconds",
            Help: "Duration of auth service requests",
        },
        []string{"method", "endpoint"},
    )
)
```

### Logging Best Practices

```go
// Structured logging for auth integration
func (c *httpAuthClient) logRequest(ctx context.Context, method, path string, err error) {
    fields := logrus.Fields{
        "method":     method,
        "path":       path,
        "service":    "auth-client",
        "request_id": middleware.GetRequestID(ctx),
        "trace_id":   middleware.GetTraceID(ctx),
    }

    if err != nil {
        c.logger.WithFields(fields).WithError(err).Warn("Auth service request failed")
    } else {
        c.logger.WithFields(fields).Debug("Auth service request completed")
    }
}
```

This comprehensive guide provides patterns and best practices for integrating services with the authentication system, ensuring secure, reliable, and maintainable service-to-service communication.</content>
</xai:function_call ><xai:function_call name="todowrite">
<parameter name="todos">[{"content":"Update docs/auth-api-examples.md with RBAC examples and service-to-service auth patterns","status":"completed","priority":"high","id":"update_auth_api_examples_rbac"},{"content":"Update docs/service-creation-guide.md with comprehensive RBAC examples","status":"completed","priority":"high","id":"update_service_creation_rbac"},{"content":"Create docs/service-integration-patterns.md with auth endpoint consumption best practices","status":"completed","priority":"medium","id":"create_service_integration_patterns"}]