# RBAC (Role-Based Access Control) API Guide

## Overview

This guide provides comprehensive documentation for the Role-Based Access Control (RBAC) system implemented in the service boilerplate. It covers both API usage and implementation details.

## Table of Contents

- [Quick Start](#quick-start)
- [API Endpoints](#api-endpoints)
- [Authentication & Authorization](#authentication--authorization)
- [RBAC Architecture](#rbac-architecture)
- [Database Schema](#database-schema)
- [Implementation Details](#implementation-details)
- [Testing](#testing)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Prerequisites

1. **Start the development environment:**

   ```bash
   make dev
   ```

2. **Run database migrations** (creates dev admin account):

   ```bash
   make db-migrate
   ```

3. **Dev Admin Account** (automatically created):
   - **Email**: `dev.admin@example.com`
   - **Password**: `devadmin123`
   - **Roles**: `admin`, `user`

### Basic RBAC Operations

```bash
# 1. Login as admin
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "dev.admin@example.com", "password": "devadmin123"}'

# 2. Extract token from response and use for RBAC operations
TOKEN="your_jwt_token_here"

# 3. List all roles
curl -X GET http://localhost:8080/api/v1/auth/roles \
  -H "Authorization: Bearer $TOKEN"

# 4. Run comprehensive RBAC tests
./scripts/test-rbac-endpoints.sh
```

## API Endpoints

All RBAC endpoints require admin authentication and are protected by role-based middleware.

### Role Management

#### Create Role

```http
POST /api/v1/auth/roles
Authorization: Bearer <admin_jwt_token>
Content-Type: application/json

{
  "name": "moderator",
  "description": "Content moderator with limited admin access"
}
```

**Response:**

```json
{
  "id": "uuid",
  "name": "moderator",
  "description": "Content moderator with limited admin access",
  "created_at": "2025-10-08T12:00:00Z"
}
```

#### List Roles

```http
GET /api/v1/auth/roles
Authorization: Bearer <admin_jwt_token>
```

**Response:**

```json
{
  "roles": [
    {
      "id": "uuid",
      "name": "admin",
      "description": "Administrator with full access",
      "created_at": "2025-10-08T12:00:00Z"
    }
  ]
}
```

#### Get Specific Role

```http
GET /api/v1/auth/roles/{role_id}
Authorization: Bearer <admin_jwt_token>
```

#### Update Role

```http
PUT /api/v1/auth/roles/{role_id}
Authorization: Bearer <admin_jwt_token>
Content-Type: application/json

{
  "name": "moderator",
  "description": "Updated description"
}
```

#### Delete Role

```http
DELETE /api/v1/auth/roles/{role_id}
Authorization: Bearer <admin_jwt_token>
```

**Response:**

```json
{
  "message": "Role deleted successfully"
}
```

### Permission Management

#### Create Permission

```http
POST /api/v1/auth/permissions
Authorization: Bearer <admin_jwt_token>
Content-Type: application/json

{
  "name": "read:reports",
  "resource": "reports",
  "action": "read"
}
```

#### List Permissions

```http
GET /api/v1/auth/permissions
Authorization: Bearer <admin_jwt_token>
```

#### Get Specific Permission

```http
GET /api/v1/auth/permissions/{permission_id}
Authorization: Bearer <admin_jwt_token>
```

#### Update Permission

```http
PUT /api/v1/auth/permissions/{permission_id}
Authorization: Bearer <admin_jwt_token>
Content-Type: application/json

{
  "name": "read:reports",
  "resource": "reports",
  "action": "read"
}
```

#### Delete Permission

```http
DELETE /api/v1/auth/permissions/{permission_id}
Authorization: Bearer <admin_jwt_token>
```

### Role-Permission Relationships

#### Assign Permission to Role

```http
POST /api/v1/auth/roles/{role_id}/permissions
Authorization: Bearer <admin_jwt_token>
Content-Type: application/json

{
  "permission_id": "uuid"
}
```

**Response:**

```json
{
  "message": "Permission assigned to role successfully"
}
```

#### Get Role Permissions

```http
GET /api/v1/auth/roles/{role_id}/permissions
Authorization: Bearer <admin_jwt_token>
```

**Response:**

```json
{
  "permissions": [
    {
      "id": "uuid",
      "name": "read:reports",
      "resource": "reports",
      "action": "read",
      "created_at": "2025-10-08T12:00:00Z"
    }
  ]
}
```

#### Remove Permission from Role

```http
DELETE /api/v1/auth/roles/{role_id}/permissions/{permission_id}
Authorization: Bearer <admin_jwt_token>
```

### User-Role Management

#### Assign Role to User

```http
POST /api/v1/auth/users/{user_id}/roles
Authorization: Bearer <admin_jwt_token>
Content-Type: application/json

{
  "role_id": "uuid"
}
```

#### Get User Roles

```http
GET /api/v1/auth/users/{user_id}/roles
Authorization: Bearer <admin_jwt_token>
```

**Response:**

```json
{
  "roles": [
    {
      "id": "uuid",
      "name": "admin",
      "description": "Administrator with full access",
      "created_at": "2025-10-08T12:00:00Z"
    }
  ]
}
```

#### Remove Role from User

```http
DELETE /api/v1/auth/users/{user_id}/roles/{role_id}
Authorization: Bearer <admin_jwt_token>
```

#### Bulk Update User Roles

```http
PUT /api/v1/auth/users/{user_id}/roles
Authorization: Bearer <admin_jwt_token>
Content-Type: application/json

{
  "role_ids": ["uuid1", "uuid2"]
}
```

## Authentication & Authorization

### JWT Token Requirements

All RBAC endpoints require:

1. **Valid JWT token** in Authorization header
2. **Admin role** in token claims
3. **Non-expired token**

### Middleware Chain

```
Request → JWT Validation → Role Check (admin) → RBAC Handler
```

### Error Responses

#### Unauthorized (401)

```json
{
  "error": "Authorization header required"
}
```

#### Forbidden (403)

```json
{
  "error": "Admin privileges required"
}
```

#### Bad Request (400)

```json
{
  "error": "Invalid role ID"
}
```

## RBAC Architecture

### Core Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │    │  Auth Service   │    │  User Service   │
│                 │    │                 │    │                 │
│ - Route Proxy   │    │ - RBAC Logic    │    │ - User Data     │
│ - Auth Check    │    │ - JWT Tokens    │    │ - Passwords     │
│ - Admin Role    │    │ - Audit Logs    │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                    │                       │
         └────────────────────┼───────────────────────┘
                              │
                    ┌───────────────────┐
                    │   PostgreSQL      │
                    │                   │
                    │ - roles           │
                    │ - permissions     │
                    │ - user_roles      │
                    │ - role_permissions│
                    └───────────────────┘
```

### Service Responsibilities

#### API Gateway

- **Routes RBAC requests** to auth-service
- **Validates JWT tokens** and admin role
- **Proxies requests** with user context

#### Auth Service

- **Implements RBAC business logic**
- **Manages roles and permissions**
- **Handles user-role assignments**
- **Provides audit logging**

#### User Service

- **Stores user accounts and passwords**
- **Provides user data** for RBAC operations

## Database Schema

### Tables Overview

```sql
-- Core RBAC tables in auth_service schema
auth_service.roles
auth_service.permissions
auth_service.role_permissions
auth_service.user_roles
```

### Roles Table

```sql
CREATE TABLE auth_service.roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Permissions Table

```sql
CREATE TABLE auth_service.permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Role-Permissions Junction

```sql
CREATE TABLE auth_service.role_permissions (
    role_id UUID REFERENCES auth_service.roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES auth_service.permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);
```

### User-Roles Junction

```sql
CREATE TABLE auth_service.user_roles (
    user_id UUID NOT NULL, -- References user_service.users(id)
    role_id UUID REFERENCES auth_service.roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id)
);
```

### Indexes

```sql
CREATE INDEX idx_roles_name ON auth_service.roles(name);
CREATE INDEX idx_permissions_name ON auth_service.permissions(name);
CREATE INDEX idx_permissions_resource_action ON auth_service.permissions(resource, action);
```

## Implementation Details

### Code Structure

```
services/auth-service/
├── internal/
│   ├── handlers/
│   │   └── auth_handler.go      # RBAC HTTP handlers
│   ├── services/
│   │   └── auth_service.go      # RBAC business logic
│   ├── repository/
│   │   └── auth_repository.go   # Database operations
│   └── models/
│       └── rbac.go              # RBAC data models
├── migrations/
│   └── development/
│       └── 0004_dev_admin_setup.up.sql
└── README.md
```

### Key Implementation Files

#### Handler Layer (`auth_handler.go`)

**Role Handlers:**

```go
func (h *AuthHandler) CreateRole(c *gin.Context) {
    // Extract trace info
    span := trace.SpanFromContext(c.Request.Context())
    traceID := span.SpanContext().TraceID().String()

    // Get authenticated admin user
    actorUserID := middleware.GetAuthenticatedUserID(c)

    // Parse request
    var req struct {
        Name        string `json:"name" binding:"required"`
        Description string `json:"description"`
    }

    // Validate and create role
    role, err := h.authService.CreateRole(c.Request.Context(), req.Name, req.Description)

    // Audit log and respond
    h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"),
        role.ID.String(), c.ClientIP(), c.GetHeader("User-Agent"),
        "create_role", traceID, spanID, true, "")
    c.JSON(http.StatusCreated, role)
}
```

#### Service Layer (`auth_service.go`)

**Business Logic:**

```go
func (s *AuthService) CreateRole(ctx context.Context, name, description string) (*models.Role, error) {
    // Validate input
    if name == "" {
        return nil, errors.New("role name is required")
    }

    // Check for duplicates
    existing, err := s.repo.GetRoleByName(ctx, name)
    if err == nil && existing != nil {
        return nil, errors.New("role already exists")
    }

    // Create role
    role := &models.Role{
        Name:        name,
        Description: description,
    }

    return s.repo.CreateRole(ctx, role)
}
```

#### Repository Layer (`auth_repository.go`)

**Database Operations:**

```go
func (r *AuthRepository) CreateRole(ctx context.Context, role *models.Role) (*models.Role, error) {
    query := `
        INSERT INTO auth_service.roles (name, description)
        VALUES ($1, $2)
        RETURNING id, name, description, created_at`

    err := r.pool.QueryRow(ctx, query, role.Name, role.Description).Scan(
        &role.ID, &role.Name, &role.Description, &role.CreatedAt)

    return role, err
}
```

### Middleware Implementation

#### Admin Role Check

```go
// middleware.RequireRole("admin")
func RequireRole(requiredRole string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get user roles from JWT claims
        userRoles := middleware.GetAuthenticatedUserRoles(c)

        // Check if user has required role
        hasRole := false
        for _, role := range userRoles {
            if role == requiredRole {
                hasRole = true
                break
            }
        }

        if !hasRole {
            c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

### Audit Logging

**Admin Actions:**

```go
func (al *AuditLogger) LogAdminAction(
    actorUserID, requestID, entityID, ipAddress, userAgent, action, traceID, spanID string,
    success bool, errorMsg string) {

    event := AuditEvent{
        EventType: "admin_action",
        UserID:    actorUserID,
        EntityID:  entityID,
        Action:    action,
        Resource:  "admin",
        // ... other fields
    }

    al.logEvent(event)
}
```

## Testing

### Automated Testing Script

Run comprehensive RBAC tests:

```bash
./scripts/test-rbac-endpoints.sh
```

### Manual Testing Examples

#### Create and Assign Role

```bash
# 1. Login as admin
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "dev.admin@example.com", "password": "devadmin123"}' | jq -r '.access_token')

# 2. Create role
curl -X POST http://localhost:8080/api/v1/auth/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "editor", "description": "Content editor"}'

# 3. Create permission
curl -X POST http://localhost:8080/api/v1/auth/permissions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "edit:articles", "resource": "articles", "action": "edit"}'

# 4. Assign permission to role
ROLE_ID="role-uuid-here"
PERM_ID="permission-uuid-here"
curl -X POST http://localhost:8080/api/v1/auth/roles/$ROLE_ID/permissions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"permission_id\": \"$PERM_ID\"}"
```

### Test Coverage

The testing script covers:

- ✅ Role CRUD operations
- ✅ Permission CRUD operations
- ✅ Role-permission assignments
- ✅ User-role assignments
- ✅ Error handling
- ✅ Data cleanup

## Best Practices

### Role Naming Conventions

```go
// Use resource:action pattern
"read:users"      // Read access to users
"create:articles" // Create articles
"update:profile"  // Update own profile
"delete:posts"    // Delete posts
"manage:roles"    // Administrative role management
```

### Permission Granularity

- **Coarse-grained**: `manage:users` (all user operations)
- **Fine-grained**: `read:users`, `create:users`, `update:users`, `delete:users`

### Role Hierarchy

```go
// Define clear role hierarchy
admin (full access)
├── manager (department management)
│   ├── supervisor (team supervision)
│   └── editor (content editing)
└── user (basic access)
```

### Security Considerations

1. **Principle of Least Privilege**: Grant minimum required permissions
2. **Regular Audits**: Review role assignments periodically
3. **Audit Logging**: All admin actions are logged
4. **Token Expiration**: JWT tokens expire and require refresh

### Performance Optimization

```sql
-- Use appropriate indexes
CREATE INDEX idx_user_roles_user_id ON auth_service.user_roles(user_id);
CREATE INDEX idx_role_permissions_role_id ON auth_service.role_permissions(role_id);

-- Cache frequently accessed permissions
// Implement Redis caching for user permissions
```

## Troubleshooting

### Common Issues

#### "Admin privileges required" (403)

**Cause**: User doesn't have admin role
**Solution**: Check JWT token contains "admin" in roles array

#### "Invalid role ID" (400)

**Cause**: UUID format incorrect or role doesn't exist
**Solution**: Verify UUID format and role exists in database

#### Migration Errors

**Cause**: Database schema conflicts
**Solution**:

```bash
# Reset database
make db-recreate
make db-migrate
```

#### Token Expired (401)

**Cause**: JWT token expired
**Solution**: Login again to get new token

### Debug Commands

```bash
# Check user roles
docker exec service-boilerplate-postgres psql -U postgres -d service_db -c "
SELECT u.email, array_agg(r.name) as roles
FROM user_service.users u
LEFT JOIN auth_service.user_roles ur ON u.id = ur.user_id
LEFT JOIN auth_service.roles r ON ur.role_id = r.id
GROUP BY u.id, u.email;"

# Check role permissions
docker exec service-boilerplate-postgres psql -U postgres -d service_db -c "
SELECT r.name as role, array_agg(p.name) as permissions
FROM auth_service.roles r
LEFT JOIN auth_service.role_permissions rp ON r.id = rp.role_id
LEFT JOIN auth_service.permissions p ON rp.permission_id = p.id
GROUP BY r.id, r.name;"

# View audit logs
docker logs service-boilerplate-auth-service 2>&1 | grep audit_event
```

### Health Checks

```bash
# Check RBAC service health
curl http://localhost:8083/health

# Test RBAC endpoints
./scripts/test-rbac-endpoints.sh
```

## Migration Guide

### From No RBAC to RBAC

1. **Run migrations** to create RBAC tables
2. **Create initial roles** (admin, user, etc.)
3. **Create permissions** for your resources
4. **Assign permissions to roles**
5. **Assign roles to existing users**
6. **Update application code** to check permissions
7. **Test thoroughly** before production deployment

### Database Migration Commands

```bash
# Run all migrations
make db-migrate

# Check migration status
make db-migrate-status

# Rollback if needed
make db-rollback
```

## API Reference

### Response Codes

- **200 OK**: Success
- **201 Created**: Resource created
- **400 Bad Request**: Invalid input
- **401 Unauthorized**: Missing/invalid token
- **403 Forbidden**: Insufficient permissions
- **404 Not Found**: Resource not found
- **409 Conflict**: Resource already exists
- **500 Internal Server Error**: Server error

### Rate Limiting

RBAC endpoints are subject to the same rate limiting as other admin endpoints.

### Pagination

List endpoints support pagination:

```http
GET /api/v1/auth/roles?page=1&limit=20
```

---

## Summary

This RBAC system provides:

- **Complete CRUD operations** for roles and permissions
- **Flexible relationships** between users, roles, and permissions
- **Comprehensive audit logging** for security compliance
- **RESTful API design** with proper HTTP status codes
- **Automated testing** and documentation
- **Production-ready architecture** with proper error handling

For questions or issues, refer to the troubleshooting section or check the service logs.
