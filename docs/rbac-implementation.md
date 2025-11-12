# 🛡️ Role-Based Access Control (RBAC) Implementation

This document provides comprehensive guidance on implementing and using Role-Based Access Control (RBAC) in the service-boilerplate project, including role definitions, assignment patterns, middleware usage, and best practices.

## 📋 Table of Contents

- [Overview](#overview)
- [RBAC Architecture](#rbac-architecture)
- [Role Definitions](#role-definitions)
- [Role Assignment](#role-assignment)
- [Middleware Implementation](#middleware-implementation)
- [Code Examples](#code-examples)
- [API Protection Patterns](#api-protection-patterns)
- [Testing RBAC](#testing-rbac)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

Role-Based Access Control (RBAC) provides fine-grained authorization by assigning permissions to roles and roles to users. The service-boilerplate implements RBAC through:

- **JWT Token Claims**: Roles embedded in JWT tokens
- **Middleware Enforcement**: Automatic role checking on protected endpoints
- **Flexible Permissions**: Support for multiple roles per user
- **Audit Logging**: All authorization decisions logged

### Key Benefits

- **Security**: Granular access control prevents unauthorized operations
- **Maintainability**: Centralized permission management
- **Auditability**: Complete authorization event logging
- **Flexibility**: Easy role assignment and permission changes

## RBAC Architecture

### Components

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   User          │────│   JWT Token      │────│   Middleware    │
│   (Roles)       │    │   (Claims)       │    │   (Enforcement) │
└─────────────────┘    └──────────────────┘    └─────────────────┘
       │                        │                        │
       └────────────────────────┴────────────────────────┴─────────┐
                                                                   │
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Auth Service  │────│   Role Assignment│────│   Audit Logs    │
│   (Management)  │    │   (Business      │    │   (Compliance)  │
│                 │    │    Logic)        │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Data Flow

1. **User Authentication**: User logs in, roles retrieved from user profile
2. **Token Generation**: Roles included in JWT claims during token creation
3. **Request Processing**: JWT middleware extracts roles from token
4. **Authorization Check**: RequireRole middleware validates user permissions
5. **Audit Logging**: Authorization decisions logged with user context

## Role Definitions

### Current Role System

The system currently supports the following roles:

#### `admin`

**Description**: Superuser role with full system access
**Permissions**:

- Manual JWT key rotation
- Access to all admin endpoints
- System configuration management
- User management across all services

**Use Cases**:

- Security administrators
- System operators
- Emergency key rotation

#### `user` (Default)

**Description**: Standard user role for normal application access
**Permissions**:

- Access to user-specific endpoints
- Profile management
- Standard application features

**Use Cases**:

- Regular application users
- API consumers
- Standard service interactions

### Role Hierarchy

```
admin (highest)
  │
  └─ user (default)
```

**Note**: Current implementation uses flat role structure. Hierarchical roles can be added by extending the middleware logic.

### Future Role Extensions

The system is designed to support additional roles:

```go
// Potential future roles
const (
    RoleAdmin     = "admin"
    RoleUser      = "user"
    RoleModerator = "moderator"
    RoleAuditor   = "auditor"
    RoleService   = "service"  // For service-to-service auth
)
```

## Role Assignment

### During User Registration

Roles are assigned during user creation in the auth service:

```go
// In auth service - user registration
user := &models.User{
    Email:     registrationRequest.Email,
    FirstName: registrationRequest.FirstName,
    LastName:  registrationRequest.LastName,
    Roles:     []string{"user"}, // Default role assignment
}

// For admin user creation (manual process)
adminUser := &models.User{
    Email:    "admin@example.com",
    Roles:    []string{"admin", "user"}, // Multiple roles supported
}
```

### Dynamic Role Assignment

Roles can be updated through user management endpoints:

```go
// Update user roles (admin only)
updatedUser := &models.User{
    Roles: []string{"admin", "user"}, // Add admin role
}
```

### Role Validation

The system validates role assignments:

- Roles must be from predefined list
- At least one role required (defaults to "user")
- Role changes audited
- Invalid roles rejected

## Middleware Implementation

### JWT Middleware Role Extraction

Roles are extracted from JWT tokens and stored in Gin context:

```go
// In common/middleware/auth.go
claims, ok := token.Claims.(*JWTClaims)
if !ok {
    // Handle invalid claims
}

// Set user roles in context
c.Set("user_roles", claims.Roles)
```

**JWTClaims Structure:**

```go
type JWTClaims struct {
    UserID    uuid.UUID `json:"user_id"`
    Email     string    `json:"email"`
    Roles     []string  `json:"roles"`      // Role array
    TokenType string    `json:"token_type"`
    jwt.RegisteredClaims
}
```

### RequireRole Middleware

The `RequireRole` middleware enforces role-based access:

```go
// RequireRole middleware implementation
func RequireRole(requiredRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRoles := GetAuthenticatedUserRoles(c)
        if len(userRoles) == 0 {
            c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }

        // Check if user has any of the required roles
        hasRequiredRole := false
        for _, requiredRole := range requiredRoles {
            for _, userRole := range userRoles {
                if userRole == requiredRole {
                    hasRequiredRole = true
                    break
                }
            }
            if hasRequiredRole {
                break
            }
        }

        if !hasRequiredRole {
            c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

### Helper Functions

```go
// Get authenticated user roles from context
func GetAuthenticatedUserRoles(c *gin.Context) []string {
    if roles, exists := c.Get("user_roles"); exists {
        if r, ok := roles.([]string); ok {
            return r
        }
    }
    return []string{}
}

// Check if user has specific role
func HasRole(c *gin.Context, role string) bool {
    userRoles := GetAuthenticatedUserRoles(c)
    for _, r := range userRoles {
        if r == role {
            return true
        }
    }
    return false
}
```

## Code Examples

### Basic Route Protection

```go
// In service main.go
router := gin.New()

// Public routes (no auth required)
public := router.Group("/api/v1")
{
    public.POST("/auth/login", authHandler.Login)
    public.POST("/auth/register", authHandler.Register)
}

// Protected routes (authentication required)
protected := router.Group("/api/v1")
protected.Use(middleware.JWTMiddleware(jwtSecret, logger, nil))
protected.Use(middleware.RequireAuth())
{
    protected.GET("/users/profile", userHandler.GetProfile)
    protected.PUT("/users/profile", userHandler.UpdateProfile)
}

// Admin-only routes
admin := router.Group("/api/v1/admin")
admin.Use(middleware.JWTMiddleware(jwtSecret, logger, nil))
admin.Use(middleware.RequireAuth())
admin.Use(middleware.RequireRole("admin"))
{
    admin.POST("/rotate-keys", adminHandler.RotateKeys)
    admin.GET("/system/status", adminHandler.SystemStatus)
}
```

### Handler Implementation

```go
// In handlers
func (h *UserHandler) UpdateProfile(c *gin.Context) {
    userID := middleware.GetAuthenticatedUserID(c)
    userRoles := middleware.GetAuthenticatedUserRoles(c)

    // Business logic with role context
    canUpdate := h.canUserUpdateProfile(userID, userRoles)

    if !canUpdate {
        c.JSON(http.StatusForbidden, gin.H{"error": "Cannot update this profile"})
        return
    }

    // Proceed with update...
}

func (h *UserHandler) canUserUpdateProfile(userID string, roles []string) bool {
    // Admin can update any profile
    for _, role := range roles {
        if role == "admin" {
            return true
        }
    }

    // Users can only update their own profile
    requestUserID := middleware.GetAuthenticatedUserID(c)
    return userID == requestUserID
}
```

### Role-Based Business Logic

```go
// Service layer with role-based decisions
func (s *UserService) DeleteUser(ctx context.Context, userID string, actorRoles []string) error {
    // Check if actor has permission to delete
    hasPermission := false
    for _, role := range actorRoles {
        if role == "admin" {
            hasPermission = true
            break
        }
    }

    if !hasPermission {
        return errors.New("insufficient permissions to delete user")
    }

    // Admin audit logging
    s.auditLogger.LogUserDeletion(
        middleware.GetAuthenticatedUserID(c),
        c.GetHeader("X-Request-ID"),
        userID,
        // ... other audit fields
    )

    return s.userRepo.DeleteUser(ctx, userID)
}
```

## API Protection Patterns

### Public API Pattern

```go
// No authentication required
router.GET("/health", healthHandler.Check)
router.POST("/auth/login", authHandler.Login)
router.POST("/auth/register", authHandler.Register)
```

### User API Pattern

```go
// Authentication required, any authenticated user
userRoutes := router.Group("/api/v1/user")
userRoutes.Use(middleware.JWTMiddleware(jwtSecret, logger, nil))
userRoutes.Use(middleware.RequireAuth())
{
    userRoutes.GET("/profile", userHandler.GetProfile)
    userRoutes.PUT("/profile", userHandler.UpdateProfile)
}
```

### Admin API Pattern

```go
// Authentication + admin role required
adminRoutes := router.Group("/api/v1/admin")
adminRoutes.Use(middleware.JWTMiddleware(jwtSecret, logger, nil))
adminRoutes.Use(middleware.RequireAuth())
adminRoutes.Use(middleware.RequireRole("admin"))
{
    adminRoutes.POST("/rotate-keys", adminHandler.RotateKeys)
    adminRoutes.GET("/system/config", adminHandler.GetConfig)
    adminRoutes.POST("/users", adminHandler.CreateUser) // Admin user creation
}
```

### Mixed Permission API Pattern

```go
// Different endpoints with different role requirements
api := router.Group("/api/v1")
api.Use(middleware.JWTMiddleware(jwtSecret, logger, nil))
api.Use(middleware.RequireAuth())
{
    // Any authenticated user
    api.GET("/public-data", dataHandler.GetPublicData)

    // Admin or specific role
    adminData := api.Group("/admin-data")
    adminData.Use(middleware.RequireRole("admin", "auditor"))
    {
        adminData.GET("", dataHandler.GetAdminData)
    }

    // User-specific data (ownership check in handler)
    api.GET("/my-data/:id", dataHandler.GetMyData)
}
```

## Testing RBAC

### Unit Testing Middleware

```go
func TestRequireRole(t *testing.T) {
    // Test cases
    tests := []struct {
        name         string
        userRoles    []string
        requiredRole string
        expectAllow  bool
    }{
        {"admin access", []string{"admin"}, "admin", true},
        {"user denied admin", []string{"user"}, "admin", false},
        {"multi-role access", []string{"user", "admin"}, "admin", true},
        {"no roles", []string{}, "admin", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Create test context with roles
            c, _ := gin.CreateTestContext(httptest.NewRecorder())
            c.Set("user_roles", tt.userRoles)

            // Test middleware
            middleware := RequireRole(tt.requiredRole)
            middleware(c)

            if tt.expectAllow {
                assert.False(t, c.IsAborted())
            } else {
                assert.True(t, c.IsAborted())
                assert.Equal(t, http.StatusForbidden, c.Writer.Status())
            }
        })
    }
}
```

### Integration Testing

```go
func TestAdminEndpoint(t *testing.T) {
    // Setup test server with RBAC
    router := setupTestRouter()

    // Test with admin token
    adminToken := createTestToken("admin-user", []string{"admin"})
    req := httptest.NewRequest("POST", "/api/v1/admin/rotate-keys", nil)
    req.Header.Set("Authorization", "Bearer "+adminToken)

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    // Test with user token (should fail)
    userToken := createTestToken("regular-user", []string{"user"})
    req = httptest.NewRequest("POST", "/api/v1/admin/rotate-keys", nil)
    req.Header.Set("Authorization", "Bearer "+userToken)

    w = httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusForbidden, w.Code)
}
```

### Token Creation Helper

```go
func createTestToken(userID string, roles []string) string {
    claims := &middleware.JWTClaims{
        UserID:    uuid.MustParse(userID),
        Email:     "test@example.com",
        Roles:     roles,
        TokenType: "access",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString([]byte("test-secret"))
    return tokenString
}
```

## Best Practices

### Role Design

1. **Principle of Least Privilege**: Assign minimum required roles
2. **Role Naming**: Use clear, descriptive role names
3. **Role Groups**: Consider role inheritance for complex hierarchies
4. **Regular Review**: Audit role assignments periodically

### Implementation

1. **Consistent Middleware Order**:

   ```go
   // Recommended order
   router.Use(corsMiddleware())
   router.Use(requestIDMiddleware())
   router.Use(tracing.HTTPMiddleware())
   router.Use(middleware.JWTMiddleware(jwtSecret, logger, revocationChecker))
   router.Use(middleware.RequireAuth())        // If needed
   router.Use(middleware.RequireRole("admin")) // If needed
   router.Use(serviceLogger.RequestResponseLogger())
   ```

2. **Error Handling**: Provide clear error messages for authorization failures

3. **Audit Logging**: Log all authorization decisions

4. **Fail-Safe Defaults**: Default to denying access when in doubt

### Security

1. **Token Validation**: Always validate JWT tokens before role checking
2. **Role Validation**: Validate roles exist and are assigned correctly
3. **Audit Everything**: Log role changes and authorization decisions
4. **Regular Rotation**: Rotate JWT keys regularly to limit role compromise impact

### Performance

1. **Cache Role Checks**: Avoid repeated role lookups for same request
2. **Efficient Storage**: Use appropriate indexes on role-related database fields
3. **Minimal Middleware**: Only apply role checks where necessary

## Troubleshooting

### Common Issues

#### 403 Forbidden Errors

**Symptoms:**

- User gets 403 despite having correct role
- Intermittent authorization failures

**Causes & Solutions:**

- **Role not in JWT**: Token may be outdated, refresh token
- **Middleware order**: Ensure JWT middleware runs before RequireRole
- **Role case sensitivity**: Check role name casing
- **Context corruption**: Verify Gin context not modified

**Debug:**

```bash
# Check token contents
echo "TOKEN" | jwt decode --secret test-secret

# Verify middleware order in logs
docker logs service-name | grep "JWT middleware\|RequireRole"
```

#### Missing User Context

**Symptoms:**

- `GetAuthenticatedUserRoles()` returns empty array
- Authorization bypassed unexpectedly

**Causes & Solutions:**

- **JWT middleware not applied**: Check route configuration
- **Invalid token**: Token parsing failed
- **Context key mismatch**: Verify context key names

**Debug:**

```go
// Add debug logging
userRoles := middleware.GetAuthenticatedUserRoles(c)
logger.WithField("user_roles", userRoles).Debug("User roles in context")
```

#### Role Assignment Issues

**Symptoms:**

- Users don't have expected roles
- Role changes not reflected in tokens

**Causes & Solutions:**

- **Database not updated**: Check user role storage
- **Token not refreshed**: Old tokens contain old roles
- **Service restart**: Auth service may need restart for config changes

**Debug:**

```sql
-- Check user roles in database
SELECT id, email, roles FROM user_service.users WHERE email = 'user@example.com';
```

### Audit Log Analysis

```bash
# Find authorization failures
docker logs auth-service | jq 'select(.level == "warn" and .error == "Insufficient permissions")'

# Check role assignment changes
docker logs auth-service | jq 'select(.event_type == "user_role_changed")'

# Monitor admin actions
docker logs auth-service | jq 'select(.user_roles[] == "admin")'
```

### Performance Issues

**Symptoms:**

- Slow authorization checks
- High CPU usage on role validation

**Solutions:**

- **Cache role lookups**: Implement role caching
- **Optimize middleware**: Reduce middleware layers
- **Database indexes**: Add indexes on role columns
- **Profile code**: Use Go profiling tools

### Testing Issues

**Symptoms:**

- Tests pass locally but fail in CI
- Role checks work in development but not production

**Solutions:**

- **Environment differences**: Check JWT secrets match
- **Token expiration**: Use valid, non-expired test tokens
- **Middleware setup**: Ensure test router configured correctly
- **Database state**: Verify test database has correct user roles

## Migration Guide

### Adding New Roles

1. **Define role constants**:

   ```go
   const (
       RoleAdmin     = "admin"
       RoleUser      = "user"
       RoleModerator = "moderator"  // New role
   )
   ```

2. **Update validation logic**:

   ```go
   func isValidRole(role string) bool {
       validRoles := []string{RoleAdmin, RoleUser, RoleModerator}
       // Check if role is in valid list
   }
   ```

3. **Update middleware** (if needed for hierarchical roles)

4. **Migrate existing users**:

   ```sql
   -- Assign new roles to existing users
   UPDATE user_service.users
   SET roles = array_append(roles, 'moderator')
   WHERE /* condition */;
   ```

5. **Update documentation**

### Changing Role Permissions

1. **Identify affected endpoints**
2. **Update RequireRole calls**:

   ```go
   // Before
   admin.Use(middleware.RequireRole("admin"))

   // After
   admin.Use(middleware.RequireRole("admin", "super_admin"))
   ```

3. **Test thoroughly**
4. **Update user assignments**
5. **Communicate changes**

## Related Documentation

- [Security Architecture](security-architecture.md) - Overall security model
- [JWT Key Rotation](jwt-key-rotation.md) - Key management security
- [Authentication API Examples](auth-api-examples.md) - Auth flow examples
- [Troubleshooting Auth & Logging](troubleshooting-auth-logging.md) - Issue resolution
- [Middleware Architecture](middleware-architecture.md) - Middleware patterns

