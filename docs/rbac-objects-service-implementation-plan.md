# RBAC Implementation Plan - Multi-Phase

**Project**: Role-Based Access Control for Objects-Service
**Architecture**: Centralized Authorization with Auth-Service
**Status**: Phase 2 Complete
**Total Estimated Time**: 20-24 hours

## Table of Contents

- [Executive Summary](#executive-summary)
- [Architecture Overview](#architecture-overview)
- [Phase Breakdown](#phase-breakdown)
- [Detailed Phase Tasks](#detailed-phase-tasks)
- [File Changes Summary](#file-changes-summary)
- [Risk Assessment](#risk-assessment)
- [Rollback Plan](#rollback-plan)

---

## Executive Summary

Implement comprehensive RBAC for objects-service with centralized permission management in auth-service:

| Component | Description |
|-----------|-------------|
| **Auth-Service** | Permission checks, caching, API endpoints |
| **Objects-Service** | JWT context extraction, permission calls, ownership validation |
| **API Gateway** | JWT signature validation (existing) |

### Key Features

- Fine-grained permissions (11 permissions)
- Dedicated `object-type-admin` role
- Ownership-based access (users manage own objects)
- Centralized caching in auth-service (TTL configurable)
- Fail-closed security model

### Test Users

| Email | Roles | Purpose |
|-------|-------|---------|
| `dev.admin@example.com` | admin, object-type-admin, user | Full admin access |
| `object.admin@example.com` | object-type-admin | Objects administration |
| `test.user@example.com` | user | Standard user access |
| `qa.tester@example.com` | user | QA testing |

---

## Architecture Overview

### Authentication Flow

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│   Client    │────▶│  API Gateway │────▶│ Objects-Service │
└─────────────┘     │  (validates  │     │                 │
                    │   JWT RS256) │     │ Extract from    │
                    └──────────────┘     │ JWT claims:     │
                                         │ - user_id       │
                                         │ - email         │
                                         │ - roles         │
                                         └────────┬────────┘
```

### Authorization Flow

```
┌─────────────────┐                              ┌─────────────────┐
│ Objects-Service │                              │   Auth-Service  │
│                 │                              │                 │
│ Extract user_id │──── POST /api/v1/auth/ ─────▶│ Check Cache     │
│ from JWT        │     permissions/check        │                 │
│                 │◀──── {"allowed": true} ──────│ Cache Miss →    │
│                 │                              │ Query DB        │
│                 │                              │ Update Cache    │
└─────────────────┘                              │ (TTL 60s)       │
                                                 └────────┬────────┘
```

### Permission Model

#### Object Types Permissions

| Permission | Description | Roles |
|------------|-------------|-------|
| `object-types:create` | Create new object types | `admin`, `object-type-admin` |
| `object-types:read` | View object types | `admin`, `object-type-admin`, `user` |
| `object-types:update` | Modify object types | `admin`, `object-type-admin` |
| `object-types:delete` | Delete object types | `admin`, `object-type-admin` |

#### Objects Permissions

| Permission | Description | Roles |
|------------|-------------|-------|
| `objects:create` | Create new objects | `admin`, `user` |
| `objects:read:all` | Read all objects | `admin`, `object-type-admin` |
| `objects:read:own` | Read only own objects | `admin`, `user` |
| `objects:update:all` | Update any object | `admin`, `object-type-admin` |
| `objects:update:own` | Update own objects only | `admin`, `user` |
| `objects:delete:all` | Delete any object | `admin`, `object-type-admin` |
| `objects:delete:own` | Delete own objects only | `admin`, `user` |

#### Role Definitions

| Role | Permissions | Purpose |
|------|-------------|---------|
| `admin` | All permissions | Full system access |
| `object-type-admin` | All `object-types:*` + `objects:read:*` + `objects:update:all` + `objects:delete:all` | Dedicated object-type management |
| `user` | `object-types:read`, `objects:create`, `objects:read:own`, `objects:update:own`, `objects:delete:own` | Standard user access |

---

## Phase Breakdown

### Phase 1: Auth-Service Foundation (4-5 hours) ✅ COMPLETED

**Goal**: Add permission caching and permission check API to auth-service

| Task | Description | Effort | Status |
|------|-------------|--------|--------|
| 1.1 | Extend config with cache settings | 0.5h | ✅ Done |
| 1.2 | Implement permission cache with interface | 1.5h | ✅ Done |
| 1.3 | Add permission check endpoint | 1h | ✅ Done |
| 1.4 | Add user permissions endpoint | 1h | ✅ Done |
| 1.5 | Write unit tests | 1h | ✅ Done |

**Deliverables**:
- ✅ Auth-service cache package (`internal/cache/permission_cache.go`)
- ✅ New API endpoints (`internal/handlers/permission_handler.go`)
- ✅ Unit tests (11 tests passing)

### Phase 2: Auth-Service Migration (2-3 hours) ✅ COMPLETED

**Goal**: Populate objects-service permissions and create dedicated role

| Task | Description | Effort | Status |
|------|-------------|--------|--------|
| 2.1 | Create permission migration | 0.5h | ✅ Done |
| 2.2 | Create object-type-admin role | 0.5h | ✅ Done |
| 2.3 | Assign permissions to roles | 0.5h | ✅ Done |
| 2.4 | Run migrations | 0.5h | ✅ Done |

**Deliverables**:
- ✅ Migration files created (`000005`, `000006`)
- ✅ Seeded permissions and roles
- ✅ `object.admin@example.com` user created

### Phase 3: Auth-Client Wrapper (2-3 hours)

**Goal**: Create client wrapper in objects-service for auth-service API calls

| Task | Description | Effort |
|------|-------------|--------|
| 3.1 | Create auth-client package | 1h |
| 3.2 | Implement permission check call | 0.5h |
| 3.3 | Implement user permissions call | 0.5h |
| 3.4 | Write unit tests | 0.5h |

**Deliverables**:
- `services/objects-service/internal/client/auth_client.go`
- Unit tests

### Phase 4: Objects-Service Integration (4-5 hours)

**Goal**: Integrate auth-client, add permission middleware, protect routes

| Task | Description | Effort |
|------|-------------|--------|
| 4.1 | Initialize auth-client in main.go | 0.5h |
| 4.2 | Create permission middleware | 1.5h |
| 4.3 | Protect routes with middleware | 1.5h |
| 4.4 | Add JWT config to config.yaml | 0.5h |

**Deliverables**:
- Auth-client initialized
- Permission middleware
- Protected routes

### Phase 5: Ownership Validation (2-3 hours)

**Goal**: Add ownership checks to update/delete operations

| Task | Description | Effort |
|------|-------------|--------|
| 5.1 | Add CreatedBy to Create request | 0.5h |
| 5.2 | Update Object model (if needed) | 0.5h |
| 5.3 | Add ownership validation in handlers | 1h |
| 5.4 | Write integration tests | 0.5h |

**Deliverables**:
- Ownership validation
- Integration tests

### Phase 6: Testing & Documentation (3-4 hours)

**Goal**: Comprehensive testing and documentation

| Task | Description | Effort |
|------|-------------|--------|
| 6.1 | Integration tests | 1.5h |
| 6.2 | Update documentation | 1h |
| 6.3 | Update swagger.yaml | 0.5h |
| 6.4 | Test script | 0.5h |

**Deliverables**:
- All tests passing
- Documentation complete

---

## Detailed Phase Tasks

### Phase 1: Auth-Service Foundation

#### Task 1.1: Extend Config with Cache Settings

**File**: `services/auth-service/internal/config/config.go`

```go
type PermissionCacheConfig struct {
    TTL        time.Duration `mapstructure:"ttl" json:"ttl"`
    MaxEntries int           `mapstructure:"max_entries" json:"max_entries"`
}

type Config struct {
    // ... existing fields ...
    PermissionCache PermissionCacheConfig `mapstructure:"permission_cache" json:"permission_cache"`
}
```

**File**: `services/auth-service/config.yaml`

```yaml
permission_cache:
  ttl: 60s
  max_entries: 10000
```

#### Task 1.2: Implement Permission Cache

**File**: `services/auth-service/internal/cache/permission_cache.go`

```go
package cache

import (
    "context"
    "sync"
    "time"

    "github.com/sirupsen/logrus"
)

type PermissionCache interface {
    GetPermissions(userID string) ([]string, bool)
    SetPermissions(userID string, permissions []string)
    GetRoles(userID string) ([]string, bool)
    SetRoles(userID string, roles []string)
    Invalidate(userID string)
    InvalidateAll()
}

type PermissionCacheConfig struct {
    TTL        time.Duration
    MaxEntries int
}

type permissionCache struct {
    mu           sync.RWMutex
    permissions  map[string]cacheEntry
    roles        map[string]cacheEntry
    ttl          time.Duration
    maxEntries   int
    logger       *logrus.Logger
}

type cacheEntry struct {
    data      interface{}
    timestamp time.Time
}

func NewPermissionCache(cfg PermissionCacheConfig, logger *logrus.Logger) PermissionCache {
    if cfg.TTL == 0 {
        cfg.TTL = 60 * time.Second
    }
    if cfg.MaxEntries == 0 {
        cfg.MaxEntries = 10000
    }

    return &permissionCache{
        permissions: make(map[string]cacheEntry),
        roles:       make(map[string]cacheEntry),
        ttl:         cfg.TTL,
        maxEntries:  cfg.MaxEntries,
        logger:      logger,
    }
}

func (c *permissionCache) GetPermissions(userID string) ([]string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    entry, exists := c.permissions[userID]
    if !exists {
        return nil, false
    }

    if time.Since(entry.timestamp) > c.ttl {
        return nil, false
    }

    if perms, ok := entry.data.([]string); ok {
        return perms, true
    }

    return nil, false
}

func (c *permissionCache) SetPermissions(userID string, permissions []string) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if len(c.permissions) >= c.maxEntries {
        c.evictOldestPermissions()
    }

    c.permissions[userID] = cacheEntry{
        data:      permissions,
        timestamp: time.Now(),
    }
}

func (c *permissionCache) GetRoles(userID string) ([]string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    entry, exists := c.roles[userID]
    if !exists {
        return nil, false
    }

    if time.Since(entry.timestamp) > c.ttl {
        return nil, false
    }

    if roles, ok := entry.data.([]string); ok {
        return roles, true
    }

    return nil, false
}

func (c *permissionCache) SetRoles(userID string, roles []string) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if len(c.roles) >= c.maxEntries {
        c.evictOldestRoles()
    }

    c.roles[userID] = cacheEntry{
        data:      roles,
        timestamp: time.Now(),
    }
}

func (c *permissionCache) Invalidate(userID string) {
    c.mu.Lock()
    defer c.mu.Unlock()

    delete(c.permissions, userID)
    delete(c.roles, userID)
}

func (c *permissionCache) InvalidateAll() {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.permissions = make(map[string]cacheEntry)
    c.roles = make(map[string]cacheEntry)
}

func (c *permissionCache) evictOldestPermissions() {
    var oldestID string
    var oldestTime time.Time

    for id, entry := range c.permissions {
        if oldestTime.IsZero() || entry.timestamp.Before(oldestTime) {
            oldestTime = entry.timestamp
            oldestID = id
        }
    }

    if oldestID != "" {
        delete(c.permissions, oldestID)
    }
}

func (c *permissionCache) evictOldestRoles() {
    var oldestID string
    var oldestTime time.Time

    for id, entry := range c.roles {
        if oldestTime.IsZero() || entry.timestamp.Before(oldestTime) {
            oldestTime = entry.timestamp
            oldestID = id
        }
    }

    if oldestID != "" {
        delete(c.roles, oldestID)
    }
}
```

#### Task 1.3: Add Permission Check Endpoint

**File**: `services/auth-service/internal/handlers/permission_handler.go`

```go
package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    "github.com/v-egorov/service-boilerplate/services/auth-service/internal/cache"
    "github.com/v-egorov/service-boilerplate/services/auth-service/internal/services"
)

type PermissionHandler struct {
    authService *services.AuthService
    cache       cache.PermissionCache
    logger      *logrus.Logger
}

type CheckPermissionRequest struct {
    UserID     string `json:"user_id" binding:"required"`
    Permission string `json:"permission" binding:"required"`
}

type CheckPermissionResponse struct {
    Allowed    bool   `json:"allowed"`
    UserID    string `json:"user_id"`
    Permission string `json:"permission"`
}

func NewPermissionHandler(authService *services.AuthService, cache cache.PermissionCache, logger *logrus.Logger) *PermissionHandler {
    return &PermissionHandler{
        authService: authService,
        cache:       cache,
        logger:      logger,
    }
}

func (h *PermissionHandler) CheckPermission(c *gin.Context) {
    var req CheckPermissionRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
        return
    }

    allowed, err := h.authService.CheckPermission(c.Request.Context(), req.UserID, req.Permission)
    if err != nil {
        h.logger.WithError(err).Error("Failed to check permission")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "permission check failed"})
        return
    }

    c.JSON(http.StatusOK, CheckPermissionResponse{
        Allowed:    allowed,
        UserID:     req.UserID,
        Permission: req.Permission,
    })
}

func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
    userID := c.Param("user_id")
    if userID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
        return
    }

    permissions, err := h.authService.GetUserPermissions(c.Request.Context(), userID)
    if err != nil {
        h.logger.WithError(err).Error("Failed to get user permissions")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get permissions"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

func (h *PermissionHandler) GetUserRoles(c *gin.Context) {
    userID := c.Param("user_id")
    if userID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
        return
    }

    roles, err := h.authService.GetUserRoles(c.Request.Context(), userID)
    if err != nil {
        h.logger.WithError(err).Error("Failed to get user roles")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get roles"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"roles": roles})
}
```

#### Task 1.4: Add Service Methods

**File**: `services/auth-service/internal/services/auth_service.go`

Add methods:

```go
func (s *AuthService) CheckPermission(ctx context.Context, userID, permission string) (bool, error) {
    // Check cache first
    if cached, found := s.cache.GetPermissions(userID); found {
        return hasPermission(cached, permission), nil
    }

    // Query database
    permissions, err := s.repo.GetUserPermissions(ctx, userID)
    if err != nil {
        return false, err
    }

    // Update cache
    s.cache.SetPermissions(userID, permissions)

    return hasPermission(permissions, permission), nil
}

func (s *AuthService) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
    // Check cache
    if cached, found := s.cache.GetPermissions(userID); found {
        return cached, nil
    }

    // Query database
    permissions, err := s.repo.GetUserPermissions(ctx, userID)
    if err != nil {
        return nil, err
    }

    // Update cache
    s.cache.SetPermissions(userID, permissions)

    return permissions, nil
}

func (s *AuthService) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
    // Check cache
    if cached, found := s.cache.GetRoles(userID); found {
        return cached, nil
    }

    // Query database
    roles, err := s.repo.GetUserRoles(ctx, userID)
    if err != nil {
        return nil, err
    }

    // Update cache
    s.cache.SetRoles(userID, roles)

    return roles, nil
}

func hasPermission(permissions []string, required string) bool {
    for _, p := range permissions {
        if p == required {
            return true
        }
    }
    return false
}
```

---

### Phase 2: Auth-Service Migration

#### Task 2.1: Create Permission Migration

**File**: `services/auth-service/migrations/development/000005_dev_object_permissions.up.sql`

```sql
-- Insert fine-grained permissions for objects-service
INSERT INTO auth_service.permissions (name, resource, action) VALUES
    -- Object Types permissions
    ('object-types:create', 'object-types', 'create'),
    ('object-types:read', 'object-types', 'read'),
    ('object-types:update', 'object-types', 'update'),
    ('object-types:delete', 'object-types', 'delete'),
    -- Objects permissions
    ('objects:create', 'objects', 'create'),
    ('objects:read:all', 'objects', 'read:all'),
    ('objects:read:own', 'objects', 'read:own'),
    ('objects:update:all', 'objects', 'update:all'),
    ('objects:update:own', 'objects', 'update:own'),
    ('objects:delete:all', 'objects', 'delete:all'),
    ('objects:delete:own', 'objects', 'delete:own')
ON CONFLICT (name) DO NOTHING;

-- Create object-type-admin role
INSERT INTO auth_service.roles (name, description) VALUES
    ('object-type-admin', 'Dedicated role for managing object types and objects')
ON CONFLICT (name) DO NOTHING;

-- Assign object-types permissions to object-type-admin role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
JOIN auth_service.permissions p ON p.name LIKE 'object-types:%'
WHERE r.name = 'object-type-admin'
ON CONFLICT DO NOTHING;

-- Assign read permissions to object-type-admin role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
JOIN auth_service.permissions p ON p.name IN ('objects:read:all', 'objects:read:own')
WHERE r.name = 'object-type-admin'
ON CONFLICT DO NOTHING;

-- Assign update:all and delete:all to object-type-admin role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
JOIN auth_service.permissions p ON p.name IN ('objects:update:all', 'objects:delete:all')
WHERE r.name = 'object-type-admin'
ON CONFLICT DO NOTHING;

-- Assign all permissions to admin role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

-- Assign basic permissions to user role
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
JOIN auth_service.permissions p ON p.name IN (
    'object-types:read',
    'objects:create',
    'objects:read:own',
    'objects:update:own',
    'objects:delete:own'
)
WHERE r.name = 'user'
ON CONFLICT DO NOTHING;
```

**File**: `services/auth-service/migrations/development/000005_dev_object_permissions.down.sql`

```sql
-- Remove permissions from user role
DELETE FROM auth_service.role_permissions
WHERE role_id IN (SELECT id FROM auth_service.roles WHERE name = 'user')
  AND permission_id IN (SELECT id FROM auth_service.permissions WHERE name LIKE 'object-types:%' OR name LIKE 'objects:%');

-- Remove all permissions from admin role
DELETE FROM auth_service.role_permissions
WHERE role_id IN (SELECT id FROM auth_service.roles WHERE name = 'admin')
  AND permission_id IN (SELECT id FROM auth_service.permissions WHERE name LIKE 'object-types:%' OR name LIKE 'objects:%');

-- Remove all permissions from object-type-admin role
DELETE FROM auth_service.role_permissions
WHERE role_id IN (SELECT id FROM auth_service.roles WHERE name = 'object-type-admin')
  AND permission_id IN (SELECT id FROM auth_service.permissions WHERE name LIKE 'object-types:%' OR name LIKE 'objects:%');

-- Delete object-type-admin role
DELETE FROM auth_service.roles WHERE name = 'object-type-admin';

-- Delete permissions
DELETE FROM auth_service.permissions WHERE name LIKE 'object-types:%' OR name LIKE 'objects:%';
```

---

### Phase 3: Auth-Client Wrapper

#### Task 3.1: Create Auth-Client Package

**File**: `services/objects-service/internal/client/auth_client.go`

```go
package client

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/sirupsen/logrus"
)

type AuthClient interface {
    CheckPermission(ctx context.Context, userID, permission string) (bool, error)
    GetUserPermissions(ctx context.Context, userID string) ([]string, error)
    GetUserRoles(ctx context.Context, userID string) ([]string, error)
}

type authClient struct {
    baseURL    string
    httpClient *http.Client
    logger     *logrus.Logger
}

type AuthClientConfig struct {
    BaseURL string
    Timeout time.Duration
}

func NewAuthClient(cfg AuthClientConfig, logger *logrus.Logger) AuthClient {
    if cfg.Timeout == 0 {
        cfg.Timeout = 10 * time.Second
    }

    return &authClient{
        baseURL: cfg.BaseURL,
        httpClient: &http.Client{
            Timeout: cfg.Timeout,
        },
        logger: logger,
    }
}

type checkPermissionRequest struct {
    UserID     string `json:"user_id"`
    Permission string `json:"permission"`
}

type checkPermissionResponse struct {
    Allowed    bool   `json:"allowed"`
    UserID     string `json:"user_id"`
    Permission string `json:"permission"`
}

func (c *authClient) CheckPermission(ctx context.Context, userID, permission string) (bool, error) {
    url := fmt.Sprintf("%s/api/v1/auth/permissions/check", c.baseURL)

    reqBody := checkPermissionRequest{
        UserID:     userID,
        Permission: permission,
    }

    body, err := json.Marshal(reqBody)
    if err != nil {
        return false, fmt.Errorf("failed to marshal request: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
    if err != nil {
        return false, fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return false, fmt.Errorf("failed to call auth-service: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return false, fmt.Errorf("auth-service returned status %d", resp.StatusCode)
    }

    var result checkPermissionResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return false, fmt.Errorf("failed to decode response: %w", err)
    }

    return result.Allowed, nil
}

func (c *authClient) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
    url := fmt.Sprintf("%s/api/v1/auth/users/%s/permissions", c.baseURL, userID)

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to call auth-service: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("auth-service returned status %d", resp.StatusCode)
    }

    var result struct {
        Permissions []string `json:"permissions"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    return result.Permissions, nil
}

func (c *authClient) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
    url := fmt.Sprintf("%s/api/v1/auth/users/%s/roles", c.baseURL, userID)

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to call auth-service: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("auth-service returned status %d", resp.StatusCode)
    }

    var result struct {
        Roles []string `json:"roles"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    return result.Roles, nil
}
```

---

### Phase 4: Objects-Service Integration

#### Task 4.1: Initialize Auth-Client

**File**: `services/objects-service/cmd/main.go`

```go
import (
    // ... existing imports
    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/client"
)

func main() {
    // ... existing setup ...

    // Initialize auth-client for permission checks
    authClient := client.NewAuthClient(client.AuthClientConfig{
        BaseURL: cfg.AuthService.URL,
        Timeout: 10 * time.Second,
    }, logger.Logger)

    // ... rest of setup ...
}
```

#### Task 4.2: Create Permission Middleware

**File**: `services/objects-service/internal/middleware/permission.go`

```go
package middleware

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/client"
)

type PermissionMiddlewareConfig struct {
    AuthClient client.AuthClient
    Logger     *logrus.Logger
}

type RequirePermissionFunc func(requiredPermissions ...string) gin.HandlerFunc

func NewPermissionMiddleware(cfg PermissionMiddlewareConfig) RequirePermissionFunc {
    return func(requiredPermissions ...string) gin.HandlerFunc {
        return func(c *gin.Context) {
            userID := GetAuthenticatedUserID(c)
            if userID == "" {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
                c.Abort()
                return
            }

            // Check each required permission
            for _, permission := range requiredPermissions {
                allowed, err := cfg.AuthClient.CheckPermission(c.Request.Context(), userID, permission)
                if err != nil {
                    cfg.Logger.WithError(err).Error("Permission check failed")
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Permission check failed"})
                    c.Abort()
                    return
                }

                if allowed {
                    // User has this permission, continue
                    c.Next()
                    return
                }
            }

            // User doesn't have any of the required permissions
            cfg.Logger.WithFields(logrus.Fields{
                "user_id": userID,
                "required": requiredPermissions,
            }).Warn("Permission denied")

            c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
            c.Abort()
        }
    }
}
```

#### Task 4.3: Protect Routes

**File**: `services/objects-service/cmd/main.go`

```go
// Initialize permission middleware
permMiddleware := middleware.NewPermissionMiddleware(middleware.PermissionMiddlewareConfig{
    AuthClient: authClient,
    Logger:     logger.Logger,
})

// Protected routes
v1 := router.Group("/api/v1")
{
    // Object Types - Admin/Object-Type-Admin only
    objectTypesAdmin := v1.Group("/object-types")
    objectTypesAdmin.Use(middleware.RequireAuth())
    objectTypesAdmin.Use(permMiddleware("object-types:create", "object-types:update", "object-types:delete"))
    {
        objectTypesAdmin.POST("", objectTypeHandler.Create)
        objectTypesAdmin.PUT("/:id", objectTypeHandler.Update)
        objectTypesAdmin.DELETE("/:id", objectTypeHandler.Delete)
    }

    // Object Types - Read for authenticated users
    objectTypesRead := v1.Group("/object-types")
    objectTypesRead.Use(middleware.RequireAuth())
    objectTypesRead.Use(permMiddleware("object-types:read"))
    {
        objectTypesRead.GET("/:id", objectTypeHandler.GetByID)
        objectTypesRead.GET("/name/:name", objectTypeHandler.GetByName)
        objectTypesRead.GET("", objectTypeHandler.List)
        objectTypesRead.GET("/search", objectTypeHandler.Search)
        objectTypesRead.GET("/:id/tree", objectTypeHandler.GetTree)
        objectTypesRead.GET("/:id/children", objectTypeHandler.GetChildren)
        objectTypesRead.GET("/:id/descendants", objectTypeHandler.GetDescendants)
        objectTypesRead.GET("/:id/ancestors", objectTypeHandler.GetAncestors)
        objectTypesRead.GET("/:id/path", objectTypeHandler.GetPath)
        objectTypesRead.GET("/:id/subtree-count", objectTypeHandler.GetSubtreeObjectCount)
    }

    // Objects - Create
    objectsCreate := v1.Group("/objects")
    objectsCreate.Use(middleware.RequireAuth())
    objectsCreate.Use(permMiddleware("objects:create"))
    {
        objectsCreate.POST("", objectHandler.Create)
    }

    // Objects - Read
    objectsRead := v1.Group("/objects")
    objectsRead.Use(middleware.RequireAuth())
    objectsRead.Use(permMiddleware("objects:read:all", "objects:read:own"))
    {
        objectsRead.GET("/:id", objectHandler.GetByID)
        objectsRead.GET("/public-id/:public_id", objectHandler.GetByPublicID)
        objectsRead.GET("/name/:name", objectHandler.GetByName)
        objectsRead.GET("", objectHandler.List)
        objectsRead.GET("/search", objectHandler.Search)
        objectsRead.GET("/:id/children", objectHandler.GetChildren)
        objectsRead.GET("/:id/descendants", objectHandler.GetDescendants)
        objectsRead.GET("/:id/ancestors", objectHandler.GetAncestors)
        objectsRead.GET("/:id/path", objectHandler.GetPath)
        objectsRead.GET("/stats", objectHandler.GetStats)
    }

    // Objects - Update
    objectsUpdate := v1.Group("/objects")
    objectsUpdate.Use(middleware.RequireAuth())
    objectsUpdate.Use(permMiddleware("objects:update:all", "objects:update:own"))
    {
        objectsUpdate.PUT("/:id", objectHandler.Update)
        objectsUpdate.PUT("/:id/metadata", objectHandler.UpdateMetadata)
        objectsUpdate.POST("/:id/tags", objectHandler.AddTags)
        objectsUpdate.DELETE("/:id/tags", objectHandler.RemoveTags)
    }

    // Objects - Delete
    objectsDelete := v1.Group("/objects")
    objectsDelete.Use(middleware.RequireAuth())
    objectsDelete.Use(permMiddleware("objects:delete:all", "objects:delete:own"))
    {
        objectsDelete.DELETE("/:id", objectHandler.Delete)
    }

    // Bulk operations
    objectsBulk := v1.Group("/objects")
    objectsBulk.Use(middleware.RequireAuth())
    objectsBulk.Use(permMiddleware("objects:create", "objects:update:all", "objects:delete:all"))
    {
        objectsBulk.POST("/bulk", objectHandler.BulkCreate)
        objectsBulk.PUT("/bulk", objectHandler.BulkUpdate)
        objectsBulk.DELETE("/bulk", objectHandler.BulkDelete)
    }
}
```

#### Task 4.4: Add Config

**File**: `services/objects-service/config.yaml`

```yaml
auth_service:
  url: "http://auth-service:8083"
  timeout_seconds: 10

jwt:
  enabled: false  # JWT validation handled by gateway
```

---

### Phase 5: Ownership Validation

#### Task 5.1: Add CreatedBy to Create Request

**File**: `services/objects-service/internal/models/object_request.go`

```go
type CreateObjectRequest struct {
    ObjectTypeID   int64     `json:"object_type_id" binding:"required"`
    Name           string    `json:"name" binding:"required,min=1,max=255"`
    Description    *string   `json:"description,omitempty"`
    ParentObjectID *int64    `json:"parent_object_id,omitempty"`
    Metadata       Metadata  `json:"metadata,omitempty"`
    Tags           []string  `json:"tags,omitempty"`

    // CreatedBy is set from JWT context, not from request body
    CreatedBy string `json:"-" db:"created_by"`
}
```

#### Task 5.3: Add Ownership Validation

**File**: `services/objects-service/internal/handlers/object_handler.go`

```go
import "slices"

func (h *ObjectHandler) Update(c *gin.Context) {
    objectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid object ID"})
        return
    }

    userID := middleware.GetAuthenticatedUserID(c)
    if userID == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
        return
    }

    object, err := h.service.GetByID(c.Request.Context(), objectID)
    if err != nil {
        handleObjectError(c, err)
        return
    }

    // Ownership check
    userRoles := middleware.GetAuthenticatedUserRoles(c)
    isAdmin := slices.Contains(userRoles, "admin")
    isObjectTypeAdmin := slices.Contains(userRoles, "object-type-admin")

    canUpdateAll := isAdmin || isObjectTypeAdmin
    ownsObject := object.CreatedBy == userID

    if !canUpdateAll && !ownsObject {
        c.JSON(http.StatusForbidden, gin.H{"error": "you can only update your own objects"})
        return
    }

    // ... rest of update logic
}

func (h *ObjectHandler) Delete(c *gin.Context) {
    objectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid object ID"})
        return
    }

    userID := middleware.GetAuthenticatedUserID(c)
    if userID == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
        return
    }

    object, err := h.service.GetByID(c.Request.Context(), objectID)
    if err != nil {
        handleObjectError(c, err)
        return
    }

    // Ownership check
    userRoles := middleware.GetAuthenticatedUserRoles(c)
    isAdmin := slices.Contains(userRoles, "admin")
    isObjectTypeAdmin := slices.Contains(userRoles, "object-type-admin")

    canDeleteAll := isAdmin || isObjectTypeAdmin
    ownsObject := object.CreatedBy == userID

    if !canDeleteAll && !ownsObject {
        c.JSON(http.StatusForbidden, gin.H{"error": "you can only delete your own objects"})
        return
    }

    // ... rest of delete logic
}
```

---

## File Changes Summary

### New Files

| File | Phase | Description |
|------|-------|-------------|
| `services/auth-service/internal/cache/permission_cache.go` | 1 | Cache with interface |
| `services/auth-service/internal/handlers/permission_handler.go` | 1 | Permission check API |
| `services/auth-service/migrations/development/000005_dev_object_permissions.up.sql` | 2 | Permission migration |
| `services/auth-service/migrations/development/000005_dev_object_permissions.down.sql` | 2 | Rollback migration |
| `services/objects-service/internal/client/auth_client.go` | 3 | Auth-service client |
| `services/objects-service/internal/middleware/permission.go` | 4 | Permission middleware |

### Modified Files

| File | Phase | Changes |
|------|-------|---------|
| `services/auth-service/internal/config/config.go` | 1 | Add cache config |
| `services/auth-service/internal/services/auth_service.go` | 1 | Add permission methods |
| `services/auth-service/cmd/main.go` | 1 | Initialize cache |
| `services/auth-service/config.yaml` | 1 | Add cache settings |
| `services/auth-service/internal/handlers/auth_handler.go` | 1 | Register new routes |
| `services/objects-service/cmd/main.go` | 4 | Initialize auth-client, protect routes |
| `services/objects-service/config.yaml` | 4 | Add auth-service config |
| `services/objects-service/internal/handlers/object_handler.go` | 5 | Ownership validation |
| `services/objects-service/internal/models/object_request.go` | 5 | Add CreatedBy field |

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Auth-service unavailable | High | Low | Fail-closed (deny all) |
| Permission cache stale | Medium | Low | TTL 60s, short-lived inconsistencies |
| Network latency to auth-service | Medium | Medium | Cache reduces calls |
| Migration conflicts | Low | Low | Use ON CONFLICT in SQL |
| Circular dependency | Medium | Low | Objects-service starts without auth-service |

---

## Rollback Plan

### Phase Rollback (Migration)

```bash
# Run down migration
psql -U postgres -d service_db -f services/auth-service/migrations/development/000005_dev_object_permissions.down.sql
```

### Code Rollback

```bash
# Revert to previous commit
git checkout HEAD~1 -- services/objects-service/cmd/main.go
git checkout HEAD~1 -- services/auth-service/internal/services/auth_service.go
```

### Emergency Procedure

If auth-service becomes unavailable:
1. Objects-service returns 500 for permission checks (fail-closed)
2. Existing cached permissions remain valid until TTL expires
3. To bypass: temporarily comment out permission middleware

---

## Estimated Timeline

| Phase | Effort | Cumulative |
|-------|--------|------------|
| Phase 1: Auth-Service Foundation | 4-5h | 4-5h |
| Phase 2: Auth-Service Migration | 2-3h | 6-8h |
| Phase 3: Auth-Client Wrapper | 2-3h | 8-11h |
| Phase 4: Objects-Service Integration | 4-5h | 12-16h |
| Phase 5: Ownership Validation | 2-3h | 14-19h |
| Phase 6: Testing & Documentation | 3-4h | 17-23h |
| **Total** | **17-23h** | |

---

## Success Criteria

### Functional

- [x] All permission endpoints return correct allow/deny (Phase 1)
- [ ] Objects-service routes protected correctly (Phase 4)
- [ ] Ownership validation works for update/delete (Phase 5)
- [x] Cache expires after TTL (Phase 1)
- [ ] Fail-closed behavior on auth-service unavailability (Phase 4)

### Non-Functional

- [x] All tests passing (Phase 1)
- [ ] < 100ms overhead for permission checks (cached) (Phase 4)
- [ ] Documentation complete (Phase 6)
- [x] Migration reversible (Phase 2)

---

## Questions

1. **Auth-service URL**: Should objects-service discover auth-service via environment variable or config?
   - Current plan: Config (`auth_service.url`)

2. **Timeout**: 10 seconds for auth-client HTTP calls - acceptable?

3. **Cache warming**: Should objects-service pre-fetch permissions on startup?
   - Not needed (lazy loading on first request)
