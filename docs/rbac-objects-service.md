# RBAC for Objects-Service

**Service**: Objects-Service with Centralized Authorization
**Auth Service**: auth-service (permission management)
**Status**: Implemented

---

## Overview

Objects-service implements Role-Based Access Control (RBAC) where permission checks are centralized in auth-service. This document describes the system architecture, permission model, and operational guidance.

### How It Works

1. Client sends request with JWT token (validated by API Gateway)
2. Objects-service extracts user identity from JWT context
3. Permission middleware calls auth-service to verify permissions
4. Auth-service checks its cache (60s TTL) or queries database
5. If authorized, request proceeds; otherwise returns 403 Forbidden

### Security Model

- **Fail-closed**: If auth-service is unavailable, all requests are denied
- **Ownership-based**: Users can only modify their own objects unless they have admin/object-type-admin role
- **JWT handled by gateway**: Objects-service trusts JWT validated by API Gateway

---

## Permission Model

### Available Permissions

#### Object Types (4 permissions)

| Permission | Description | Who Can Use |
|------------|-------------|-------------|
| `object-types:create` | Create new object types | admin, object-type-admin |
| `object-types:read` | View object types | admin, object-type-admin, user |
| `object-types:update` | Modify object types | admin, object-type-admin |
| `object-types:delete` | Delete object types | admin, object-type-admin |

#### Objects (7 permissions)

| Permission | Description | Who Can Use |
|------------|-------------|-------------|
| `objects:create` | Create new objects | admin, user |
| `objects:read:all` | Read all objects | admin, object-type-admin |
| `objects:read:own` | Read only own objects | admin, user |
| `objects:update:all` | Update any object | admin, object-type-admin |
| `objects:update:own` | Update own objects only | admin, user |
| `objects:delete:all` | Delete any object | admin, object-type-admin |
| `objects:delete:own` | Delete own objects only | admin, user |

### Role Definitions

| Role | Permissions | Use Case |
|------|-------------|----------|
| `admin` | All 11 permissions | Full system access |
| `object-type-admin` | All object-types:* + objects:read:* + objects:update:all + objects:delete:all | Object and object-type management |
| `user` | object-types:read, objects:create, objects:read:own, objects:update:own, objects:delete:own | Standard user operations |

### Ownership Rules

- **Regular users** (role `user`): Can only read, create, update, delete objects they created
- **Object-type-admin**: Can read all objects, update/delete any object
- **admin**: Full access to all objects regardless of ownership

---

## Configuration

### Objects-Service Config

File: `services/objects-service/config.yaml`

```yaml
auth_service:
  url: "http://auth-service:8083"
  timeout_seconds: 10

jwt:
  enabled: false  # JWT validation handled by API Gateway
```

### Auth-Service Config

File: `services/auth-service/config.yaml`

```yaml
permission_cache:
  ttl: 60s        # Cache duration for permissions
  max_entries: 10000
```

### Environment Variables

Objects-service requires these environment variables (in Docker Compose):

```yaml
services:
  objects-service:
    environment:
      - AUTH_SERVICE_URL=http://auth-service:8083
      - AUTH_SERVICE_TIMEOUT=10s
```

---

## API Endpoints

### Auth-Service Permission APIs

These endpoints are called internally by objects-service:

#### Check Permission

```
POST /api/v1/auth/permissions/check
Content-Type: application/json

Request:
{
  "user_id": "user-123",
  "permission": "objects:create"
}

Response (allowed):
{
  "allowed": true,
  "user_id": "user-123",
  "permission": "objects:create"
}

Response (denied):
{
  "allowed": false,
  "user_id": "user-123",
  "permission": "objects:create"
}
```

#### Get User Permissions

```
GET /api/v1/auth/users/{user_id}/permissions

Response:
{
  "permissions": [
    "object-types:read",
    "objects:create",
    "objects:read:own",
    "objects:update:own",
    "objects:delete:own"
  ]
}
```

#### Get User Roles

```
GET /api/v1/auth/users/{user_id}/roles

Response:
{
  "roles": ["user"]
}
```

### Objects-Service Routes

All routes require authentication (JWT). Permission requirements:

| Method | Endpoint | Required Permission |
|--------|----------|---------------------|
| POST | /api/v1/object-types | object-types:create |
| GET | /api/v1/object-types | object-types:read |
| PUT | /api/v1/object-types/:id | object-types:update |
| DELETE | /api/v1/object-types/:id | object-types:delete |
| POST | /api/v1/objects | objects:create |
| GET | /api/v1/objects | objects:read:all OR objects:read:own |
| PUT | /api/v1/objects/:id | objects:update:all OR objects:update:own |
| DELETE | /api/v1/objects/:id | objects:delete:all OR objects:delete:own |

---

## Test Users

For development/testing:

| Email | Roles | Capabilities |
|-------|-------|--------------|
| dev.admin@example.com | admin, object-type-admin, user | Full access to everything |
| object.admin@example.com | object-type-admin | Manage all objects, limited object-type admin |
| test.user@example.com | user | Can only manage own objects |
| qa.tester@example.com | user | Standard user for QA testing |

---

## Caching

Auth-service caches user permissions and roles to reduce database load.

### Cache Behavior

- **TTL**: 60 seconds (configurable)
- **Max Entries**: 10,000 users
- **Cache Invalidation**: Automatic after TTL expires
- **Storage**: In-memory (interface designed for Redis migration)

### Cache Performance

| Scenario | Latency |
|----------|---------|
| Cache hit | < 1ms |
| Cache miss (DB query) | ~10-50ms |
| Auth-service unavailable | Returns error (fail-closed) |

---

## Troubleshooting

### Permission Denied (403)

**Symptoms**: User receives 403 Forbidden even though they should have access.

**Possible Causes**:
1. User doesn't have the required role/permission in auth-service
2. Cache returning stale data (wait 60s for TTL expiry)
3. User ID mismatch between JWT and auth-service records

**Resolution**:
1. Verify user's roles in auth-service database:
   ```sql
   SELECT u.email, array_agg(r.name) as roles
   FROM auth_service.users u
   JOIN auth_service.user_roles ur ON u.id = ur.user_id
   JOIN auth_service.roles r ON ur.role_id = r.id
   WHERE u.email = 'user@example.com'
   GROUP BY u.email;
   ```
2. Verify user's permissions:
   ```sql
   SELECT u.email, array_agg(p.name) as permissions
   FROM auth_service.users u
   JOIN auth_service.user_roles ur ON u.id = ur.user_id
   JOIN auth_service.role_permissions rp ON ur.role_id = rp.role_id
   JOIN auth_service.permissions p ON rp.permission_id = p.id
   WHERE u.email = 'user@example.com'
   GROUP BY u.email;
   ```

### Auth-Service Unavailable (500)

**Symptoms**: Objects-service returns 500 Internal Server Error.

**Possible Causes**:
1. Auth-service is down
2. Network connectivity issue between objects-service and auth-service
3. Timeout waiting for auth-service response

**Resolution**:
1. Check auth-service status:
   ```bash
   docker ps | grep auth-service
   curl http://localhost:8083/health
   ```
2. Check objects-service logs:
   ```bash
   make logs SERVICE_NAME=objects-service
   ```
3. Verify network connectivity:
   ```bash
   docker exec objects-service curl http://auth-service:8083/health
   ```

### Ownership Validation Error

**Symptoms**: User cannot update/delete their own object.

**Possible Causes**:
1. `created_by` field not properly set when object was created
2. User ID in JWT doesn't match `created_by` in database

**Resolution**:
1. Check object's created_by:
   ```sql
   SELECT id, name, created_by FROM objects.objects WHERE id = <object_id>;
   ```
2. Verify user ID from JWT matches created_by

### JWT Authentication Issues

**Symptoms**: 401 Unauthorized

**Possible Causes**:
1. Missing or invalid JWT token
2. Token expired
3. API Gateway not validating JWT correctly

**Resolution**:
1. Ensure request includes valid JWT in Authorization header
2. Check API Gateway configuration
3. Verify JWT payload contains required claims (user_id, email, roles)

---

## Monitoring

### Key Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| permission_check_latency | Time for permission check | > 100ms |
| permission_check_errors | Failed permission checks | > 0 |
| auth_service_health | Auth-service availability | not healthy |

### Log Fields

Permission-related log fields:
- `user_id`: User making the request
- `required_permissions`: Permissions being checked
- `matched_permissions`: Permissions user actually has
- `ownership_valid`: Whether ownership check passed

---

## Migration Reference

### Apply RBAC Migrations

```bash
# Run up migration
make db-migrate SERVICE_NAME=auth-service

# Or via migration orchestrator
docker exec migration-orchestrator migrate -service auth-service
```

### Rollback RBAC Migrations

```bash
# Run down migration
make db-migrate-down SERVICE_NAME=auth-service
```

---

## Security Considerations

### Fail-Closed Behavior

If auth-service becomes unavailable:
- All permission checks fail
- Objects-service denies all requests
- Service remains operational but read-only

### Permission Escalation Prevention

- Ownership is checked after permission check
- Users cannot escalate privileges by modifying JWT
- Database enforces role-permission relationships

### Audit Trail

All permission checks are logged with:
- User ID
- Requested permission
- Result (allowed/denied)
- Timestamp

---

## Future Enhancements

Potential improvements (not implemented):
- Redis-backed permission cache for distributed deployments
- Permission cache warming on startup
- Granular object-level permissions (beyond ownership)
- Audit logs for permission decisions
- Rate limiting on permission check endpoints
