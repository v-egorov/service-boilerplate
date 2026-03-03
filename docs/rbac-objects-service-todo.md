# RBAC Implementation - TODO List

**Project**: Role-Based Access Control for Objects-Service
**Status**: Phase 4 Complete
**Started**: 2026-02-12
**Completed**: 2026-02-28
**Total Estimated Time**: 17-23 hours

---

## Progress Summary

| Phase | Status | Tasks | Progress |
|-------|--------|-------|----------|
| Phase 1: Auth-Service Foundation | ✅ | 5 | 5/5 |
| Phase 2: Auth-Service Migration | ✅ | 4 | 4/4 |
| Phase 3: Auth-Client Wrapper | ✅ | 4 | 4/4 |
| Phase 4: Objects-Service Integration | ✅ | 4 | 4/4 |
| Phase 5: Ownership Validation | ⬜ | 4 | 0/4 |
| Phase 6: Testing & Documentation | ⬜ | 4 | 0/4 |
| **Total** | ✅ | **25** | **17/25** |

---

## Phase 1: Auth-Service Foundation

**Goal**: Add permission caching and permission check API to auth-service
**Estimated Time**: 4-5 hours
**Status**: ✅ COMPLETED

### Tasks

- [x] **1.1** Extend config with cache settings (0.5h)
  - File: `services/auth-service/internal/config/config.go`
  - Add `PermissionCacheConfig` struct

- [x] **1.2** Implement permission cache with interface (1.5h)
  - File: `services/auth-service/internal/cache/permission_cache.go`
  - Interface-based design (Redis-ready)
  - TTL: 60s configurable
  - MaxEntries: 10000
  - Thread-safe with RWMutex
  - LRU eviction

- [x] **1.3** Add permission check endpoint (1h)
  - File: `services/auth-service/internal/handlers/permission_handler.go`
  - POST `/api/v1/auth/permissions/check`
  - Request: `{user_id, permission}`
  - Response: `{allowed, user_id, permission}`

- [x] **1.4** Add user permissions endpoint (1h)
  - File: `services/auth-service/internal/handlers/permission_handler.go`
  - GET `/api/v1/auth/users/{user_id}/permissions`
  - Returns user's permissions list

- [x] **1.5** Write unit tests (1h)
  - File: `services/auth-service/internal/cache/permission_cache_test.go`
  - Cache hit/miss tests
  - TTL expiration tests
  - Concurrent access tests

### Deliverables
- [x] Auth-service cache package
- [x] New API endpoints registered in `auth_handler.go`
- [x] Unit tests passing

---

## Phase 2: Auth-Service Migration

**Goal**: Populate objects-service permissions and create dedicated role
**Estimated Time**: 2-3 hours
**Status**: ✅ COMPLETED

### Tasks

- [x] **2.1** Create permission migration (0.5h)
  - File: `services/auth-service/migrations/000005_dev_object_permissions.up.sql`
  - 11 permissions (4 object-types + 7 objects)
  - Includes `objects:read:all` and `objects:read:own`

- [x] **2.2** Create object-type-admin role (0.5h)
  - File: `services/auth-service/migrations/development/000006_dev_object_permissions_seed.up.sql`
  - Add role to `auth_service.roles`

- [x] **2.3** Assign permissions to roles (0.5h)
  - Admin: All 11 permissions
  - object-type-admin: All object-types:* + objects:read:* + objects:update:all + objects:delete:all
  - User: object-types:read, objects:create, objects:read:own, objects:update:own, objects:delete:own

- [x] **2.4** Create rollback migration (0.5h)
  - File: `services/auth-service/migrations/000005_dev_object_permissions.down.sql`
  - Removes all permissions
  - Removes object-type-admin role

### Deliverables
- [x] Migration files created
- [x] Permissions and roles seeded
- [x] Migration tested

---

## Phase 3: Auth-Client Wrapper

**Goal**: Create client wrapper in objects-service for auth-service API calls
**Estimated Time**: 2-3 hours
**Status**: ✅ COMPLETED

### Tasks

- [x] **3.1** Create auth-client package (1h)
  - File: `services/objects-service/internal/client/auth_client.go`
  - Interface-based design
  - HTTP client with configurable timeout

- [x] **3.2** Implement permission check call (0.5h)
  - Method: `CheckPermission(ctx, userID, permission) (bool, error)`
  - Calls `POST /api/v1/auth/permissions/check`

- [x] **3.3** Implement user permissions call (0.5h)
  - Method: `GetUserPermissions(ctx, userID) ([]string, error)`
  - Calls `GET /api/v1/auth/users/{user_id}/permissions`

- [x] **3.4** Write unit tests (0.5h)
  - File: `services/objects-service/internal/client/auth_client_test.go`
  - Mock HTTP responses
  - Error handling tests (6 tests passing)

### Deliverables
- [x] `services/objects-service/internal/client/auth_client.go`
- [x] Unit tests passing

---

## Phase 4: Objects-Service Integration

**Goal**: Integrate auth-client, add permission middleware, protect routes
**Estimated Time**: 4-5 hours
**Status**: ✅ COMPLETED

### Tasks

- [x] **4.1** Add config for auth-service (0.5h)
  - File: `services/objects-service/config.yaml`
  - Add `auth_service.url` and `auth_service.timeout_seconds`
  - ✅ DONE in Phase 3

- [x] **4.2** Initialize auth-client in main.go (0.5h)
  - File: `services/objects-service/cmd/main.go`
  - Create `client.NewAuthClient()`
  - Pass to middleware

- [x] **4.3** Create permission middleware (1.5h)
  - File: `services/objects-service/internal/permiddleware/permission.go`
  - Function: `RequirePermission(permissions ...string) gin.HandlerFunc`
  - Calls auth-client for permission checks
  - Fail-closed behavior

- [x] **4.4** Protect routes with middleware (1.5h)
  - File: `services/objects-service/cmd/main.go`
  - Object Types: POST/PUT/DELETE → object-types:create/update/delete
  - Object Types: GET → object-types:read
  - Objects: POST → objects:create
  - Objects: GET → objects:read:all OR objects:read:own
  - Objects: PUT → objects:update:all OR objects:update:own
  - Objects: DELETE → objects:delete:all OR objects:delete:own
  - JWT validation: handled by API Gateway (config pre-existing)

### Deliverables
- [x] Auth-client initialized
- [x] Permission middleware created
- [x] All routes protected (fail-closed)
- [x] Unit tests passing (5 tests)
- [x] JWT config (pre-existing, handled by gateway)

- [ ] **4.2** Initialize auth-client in main.go (0.5h)
  - File: `services/objects-service/cmd/main.go`
  - Create `client.NewAuthClient()`
  - Pass to middleware

- [ ] **4.3** Create permission middleware (1.5h)
  - File: `services/objects-service/internal/middleware/permission.go`
  - Function: `RequirePermission(permissions ...string) gin.HandlerFunc`
  - Calls auth-client for permission checks

- [ ] **4.4** Protect routes with middleware (1.5h)
  - File: `services/objects-service/cmd/main.go`
  - Object Types: POST/PUT/DELETE → object-types:create/update/delete
  - Object Types: GET → object-types:read
  - Objects: POST → objects:create
  - Objects: GET → objects:read:all OR objects:read:own
  - Objects: PUT → objects:update:all OR objects:update:own
  - Objects: DELETE → objects:delete:all OR objects:delete:own

### Deliverables
- [ ] Auth-client initialized
- [ ] Permission middleware created
- [ ] All routes protected

---

## Phase 5: Ownership Validation

**Goal**: Add ownership checks to update/delete operations
**Estimated Time**: 2-3 hours
**Status**: ⬜ Not Started

### Tasks

- [ ] **5.1** Add CreatedBy to Create request (0.5h)
  - File: `services/objects-service/internal/models/object_request.go`
  - Field: `CreatedBy string` (json:"-", db:"created_by")
  - Set from JWT context in handler

- [ ] **5.2** Update Object model (0.5h)
  - Verify `created_by` column exists in schema
  - May need migration update

- [ ] **5.3** Add ownership validation in handlers (1h)
  - File: `services/objects-service/internal/handlers/object_handler.go`
  - Update: Check `object.CreatedBy == userID` or admin role
  - Delete: Check `object.CreatedBy == userID` or admin role

- [ ] **5.4** Write integration tests (0.5h)
  - Test owner can update/delete own object
  - Test user cannot update/delete other's object
  - Test admin can update/delete any object

### Deliverables
- [ ] CreatedBy field populated on object creation
- [ ] Ownership validation working
- [ ] Integration tests passing

---

## Phase 6: Testing & Documentation

**Goal**: Comprehensive testing and documentation
**Estimated Time**: 3-4 hours
**Status**: ⬜ Not Started

### Tasks

- [ ] **6.1** Integration tests (1.5h)
  - File: `tests/integration/rbac_integration_test.go`
  - Test permission check flow end-to-end
  - Test cache behavior
  - Test fail-closed on auth-service unavailability

- [ ] **6.2** Update documentation (1h)
  - File: `docs/rbac-objects-service.md`
  - Architecture diagram
  - Permission model
  - Usage examples

- [ ] **6.3** Update swagger.yaml (0.5h)
  - File: `services/objects-service/api/swagger.yaml`
  - Document permission requirements

- [ ] **6.4** Create test script (0.5h)
  - File: `scripts/test-rbac-objects-service.sh`
  - End-to-end RBAC test

### Deliverables
- [ ] All tests passing
- [ ] Documentation complete
- [ ] Swagger updated

---

## File Changes Summary

### New Files Created

| File | Phase | Status |
|------|-------|--------|
| `services/auth-service/internal/cache/permission_cache.go` | 1 | ✅ |
| `services/auth-service/internal/cache/permission_cache_test.go` | 1 | ✅ |
| `services/auth-service/internal/handlers/permission_handler.go` | 1 | ✅ |
| `services/auth-service/migrations/000005_dev_object_permissions.up.sql` | 2 | ✅ |
| `services/auth-service/migrations/000005_dev_object_permissions.down.sql` | 2 | ✅ |
| `services/auth-service/migrations/development/000006_dev_object_permissions_seed.up.sql` | 2 | ✅ |
| `services/auth-service/migrations/development/000006_dev_object_permissions_seed.down.sql` | 2 | ✅ |
| `services/user-service/migrations/development/000005_dev_add_object_admin.up.sql` | 2 | ✅ |
| `services/user-service/migrations/development/000005_dev_add_object_admin.down.sql` | 2 | ✅ |
| `services/objects-service/internal/client/auth_client.go` | 3 | ✅ |
| `services/objects-service/internal/client/auth_client_test.go` | 3 | ✅ |
| `services/objects-service/internal/permiddleware/permission.go` | 4 | ✅ |
| `services/objects-service/internal/permiddleware/permission_test.go` | 4 | ✅ |
| `tests/integration/rbac_integration_test.go` | 6 | ⬜ |
| `docs/rbac-objects-service.md` | 6 | ⬜ |
| `scripts/test-rbac-objects-service.sh` | 6 | ⬜ |

### Files Modified

| File | Phase | Changes | Status |
|------|-------|---------|--------|
| `services/auth-service/internal/config/config.go` | 1 | Add `PermissionCacheConfig` | ✅ |
| `services/auth-service/internal/services/auth_service.go` | 1 | Add `CheckPermission`, `GetUserPermissions`, `GetUserRoles` methods | ✅ |
| `services/auth-service/cmd/main.go` | 1 | Initialize cache, register permission routes | ✅ |
| `services/auth-service/config.yaml` | 1 | Add `permission_cache` settings | ✅ |
| `services/auth-service/internal/handlers/auth_handler.go` | 1 | Register new permission endpoints | ✅ |
| `services/auth-service/internal/handlers/auth_handler_test.go` | 1 | Update MockAuthService | ✅ |
| `services/auth-service/internal/services/auth_service_test.go` | 1 | Update MockAuthRepository | ✅ |
| `services/auth-service/migrations/dependencies.json` | 2 | Add 000005, 000006 entries | ✅ |
| `services/auth-service/migrations/environments.json` | 2 | Add migrations to dev/staging | ✅ |
| `services/user-service/migrations/dependencies.json` | 2 | Add 000005 entry | ✅ |
| `services/user-service/migrations/environments.json` | 2 | Add migration to dev/staging | ✅ |
| `common/config/config.go` | 3 | Add `AuthServiceConfig` struct | ✅ |
| `services/objects-service/config.yaml` | 3 | Add `auth_service` settings | ✅ |
| `services/objects-service/cmd/main.go` | 4 | Initialize auth-client, add route protection | ✅ |
| `services/objects-service/internal/handlers/object_handler.go` | 5 | Add ownership validation | ⬜ |
| `services/objects-service/internal/models/object_request.go` | 5 | Add `CreatedBy` field | ⬜ |
| `services/objects-service/api/swagger.yaml` | 6 | Update with permission requirements | ⬜ |

---

## Commands Reference

### Run Tests
```bash
# Auth-service tests
cd services/auth-service && go test ./...

# Objects-service tests
cd services/objects-service && go test ./...

# Integration tests
go test ./tests/integration/...
```

### Run Migration
```bash
# Run up migration
psql -U postgres -d service_db -f services/auth-service/migrations/development/000005_dev_object_permissions.up.sql

# Run down migration
psql -U postgres -d service_db -f services/auth-service/migrations/development/000005_dev_object_permissions.down.sql
```

### Run Test Script
```bash
./scripts/test-rbac-objects-service.sh
```

---

## Permission Model Reference

### Object Types Permissions

| Permission | Description | Roles |
|------------|-------------|-------|
| `object-types:create` | Create new object types | admin, object-type-admin |
| `object-types:read` | View object types | admin, object-type-admin, user |
| `object-types:update` | Modify object types | admin, object-type-admin |
| `object-types:delete` | Delete object types | admin, object-type-admin |

### Objects Permissions

| Permission | Description | Roles |
|------------|-------------|-------|
| `objects:create` | Create new objects | admin, user |
| `objects:read:all` | Read all objects | admin, object-type-admin |
| `objects:read:own` | Read only own objects | admin, user |
| `objects:update:all` | Update any object | admin, object-type-admin |
| `objects:update:own` | Update own objects only | admin, user |
| `objects:delete:all` | Delete any object | admin, object-type-admin |
| `objects:delete:own` | Delete own objects only | admin, user |

### Role Definitions

| Role | Permissions |
|------|-------------|
| `admin` | All permissions |
| `object-type-admin` | object-types:*, objects:read:*, objects:update:all, objects:delete:all |
| `user` | object-types:read, objects:create, objects:read:own, objects:update:own, objects:delete:own |

---

## Notes

- **Cache TTL**: 60 seconds (configurable)
- **Cache Strategy**: Interface-based, Redis-ready
- **Security Model**: Fail-closed (deny all on auth-service unavailability)
- **Ownership**: Users can only modify own objects (unless admin/object-type-admin)

---

## Changelog

| Date | Phase | Change | Author |
|------|-------|--------|--------|
| 2026-02-12 | - | Created TODO list | - |
| 2026-02-28 | 1,2 | Implemented Phase 1 (permission cache) and Phase 2 (migrations) | - |
| 2026-03-03 | 3 | Implemented Phase 3 (auth-client wrapper) | - |
| 2026-03-03 | 4 | Implemented Phase 4 (objects-service integration, route protection) | - |
