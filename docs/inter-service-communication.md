# Inter-Service Communication Standards

## Overview

This document details all inter-service communication patterns in the boilerplate project, specifically focusing on how response formats have evolved through the API Response Standardization effort (Phase 1: Error responses, Phase 2: Success responses).

## Service Architecture

The project consists of three microservices with the following communication patterns:

```
┌─────────────────┐
│  API Gateway    │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
┌───▼───┐ ┌──▼────────┐
│ Auth  │ │ Objects   │
│SVC    │ │ SVC       │
└───┬───┘ └───────────┘
    │
┌───▼────┐
│ User   │
│ SVC    │
└────────┘
```

**Communication flows:**
1. API Gateway → Auth-Service (authentication, registration, token management)
2. API Gateway → User-Service (user CRUD operations)
3. API Gateway → Objects-Service (object management)
4. Auth-Service → User-Service (user lookup, registration delegation)
5. Objects-Service → Auth-Service (permission checks, role/permission retrieval)

---

## Phase 1 Error Response Standardization Status

### Standardized Error Format

```json
{
  "error": "Human-readable error message",
  "type": "<error_type>",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

### Supported Error Types

| Type | HTTP Status | Description |
|------|-------------|-------------|
| `validation_error` | 400, 422 | Invalid input, missing required fields |
| `unauthorized` | 401 | Authentication failed |
| `permission_denied` | 403 | Authorization failed |
| `not_found` | 404 | Resource not found |
| `conflict` | 409 | Resource conflict (duplicate, version) |
| `internal_error` | 500 | Server error |

### Error Response Implementation Status

#### Auth-Service

| Handler File | Status | Commit | Changes |
|-------------|--------|--------|---------|
| `auth_handler.go` | ✅ Complete | `7ef7606` | ~53 error responses with `type` + `meta.request_id` |
| `permission_handler.go` | ✅ Complete | `7ef7606` | 4 error responses with `type` + `meta.request_id` |

**Error types used:** `validation_error`, `unauthorized`, `permission_denied`, `not_found`, `conflict`, `internal_error`

#### User-Service

| Handler File | Status | Commit | Changes |
|-------------|--------|--------|---------|
| `user_handler.go` | ✅ Complete | `e55b2d5` | Error types renamed (`conflict_error` → `conflict`, `not_found_error` → `not_found`) |

#### Objects-Service

| Handler File | Status | Commit | Changes |
|-------------|--------|--------|---------|
| `object_handler.go` | ✅ Complete | `e55b2d5` | Error type `version_conflict` → `conflict` |
| `relationship_handler.go` | ✅ Complete | `627092b` | `details` field removed, `meta.request_id` added |
| `relationship_type_handler.go` | ✅ Complete | `627092b` | `details` field removed, `meta.request_id` added |

---

## Phase 2 Success Response Standardization Status

### Standardized Success Format

```json
{
  "data": { ... },
  "message": "Human-readable success message",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

### Success Response Implementation Status

#### Auth-Service

| Endpoint | Method | Handler | Status | Commit | Format |
|----------|--------|---------|--------|--------|--------|
| `/api/v1/auth/login` | POST | `auth_handler.go` | ✅ Complete | `7f1c0b7` | `{"access_token": ..., "refresh_token": ..., "user": {"data": {...}, "meta": {...}}}` |
| `/api/v1/auth/register` | POST | `auth_handler.go` | ✅ Complete | `7f1c0b7` | `{"message": "...", "user": {...}, "meta": {...}}` |
| `/api/v1/auth/logout` | POST | `auth_handler.go` | ✅ Complete | `7f1c0b7` | `{"message": "...", "meta": {...}}` |
| `/health` | GET | `health_handler.go` | ✅ Complete | `7f1c0b7` | `{"data": {...}, "meta": {...}}` |
| `/liveness` | GET | `health_handler.go` | ✅ Complete | `7f1c0b7` | `{"data": {...}, "meta": {...}}` |
| `/readiness` | GET | `health_handler.go` | ✅ Complete | `7f1c0b7` | `{"data": {...}, "meta": {...}}` |
| `/status` | GET | `health_handler.go` | ✅ Complete | `7f1c0b7` | `{"data": {...}, "meta": {...}}` |

**Note:** Auth-service internal endpoints (`permission_handler.go`, `GetUserRoles`) intentionally remain in FLAT format for service-to-service communication. See section "Special Cases: Internal Service Endpoints" below.

#### User-Service

| Endpoint | Method | Handler | Status | Commit | Changes |
|----------|--------|---------|--------|--------|---------|
| `/api/v1/users` | POST | `user_handler.go` | ✅ Complete | `7f1c0b7` | Add `meta.request_id` |
| `/api/v1/users/:id` | GET | `user_handler.go` | ✅ Complete | `7f1c0b7` | Add `meta.request_id` |
| `/api/v1/users/:id` | PUT | `user_handler.go` | ✅ Complete | `7f1c0b7` | Add `meta.request_id` |
| `/api/v1/users/:id` | PATCH | `user_handler.go` | ✅ Complete | `7f1c0b7` | Add `meta.request_id` |
| `/api/v1/users` | GET | `user_handler.go` | ✅ Complete | `7f1c0b7` | Add `meta.request_id` |
| `/api/v1/users/by-email/:email` | GET | `user_handler.go` | ✅ Complete | `7f1c0b7` | Add `meta.request_id` |
| `/api/v1/users/by-email/:email/with-password` | GET | `user_handler.go` | ✅ Complete | `7f1c0b7` | Wrapped in `{"data": {...}, "meta": {...}}` |

#### Objects-Service

| Handler File | Status | Commit | Success Responses |
|-------------|--------|--------|-------------------|
| `object_handler.go` | ✅ Complete | `6f359af` | 14 success responses |
| `object_type_handler.go` | ✅ Complete | `b034afa` | 13 success responses |
| `relationship_handler.go` | ✅ Complete | `8c9f5bd` | 6 success responses |
| `relationship_type_handler.go` | ✅ Complete | `8c9f5bd` | 4 success responses |

---

## Inter-Service Clients: Detailed Analysis

### 1. Auth-Service → User-Service Client

**File:** `services/auth-service/internal/client/user_client.go`

#### Client Definition

```go
type UserClient struct {
    baseURL    string
    httpClient *http.Client
    logger     *logrus.Logger
}
```

#### Response Structs

| Struct | Fields | Uses Data Wrapper? | Status |
|--------|--------|-------------------|--------|
| `UserServiceResponse` | `Data *UserData` | ✅ Yes | ✅ OK |
| `UserData` | `ID`, `Email`, `FirstName`, `LastName` | N/A (nested) | ✅ OK |
| `UserLoginResponse` | `Data *struct{User, PasswordHash}` | ✅ Yes | ✅ FIXED (commit `e82e6bd`) |

**Historical Issue:** The `UserLoginResponse` struct originally expected flat JSON without a `Data` wrapper, which broke when user-service updated `GetUserWithPasswordByEmail` to return `{"data": {...}, "meta": {...}}`. This was fixed in commit `e82e6bd` by adding the `Data` wrapper field.

#### Methods and Endpoint Mapping

| Method | Call Format | Upstream Endpoint | Response Struct | JSON Format Expected | Status |
|--------|-------------|-------------------|-----------------|---------------------|--------|
| `CreateUser` | `POST /api/v1/users` | User-Service | `UserServiceResponse` | `{"data": {...}, "meta": {...}}` | ✅ OK |
| `GetUserByEmail` | `GET /api/v1/users/by-email/:email` | User-Service | `UserServiceResponse` | `{"data": {...}, "meta": {...}}` | ✅ OK |
| `GetUserWithPasswordByEmail` | `GET /api/v1/users/by-email/:email/with-password` | User-Service | `UserLoginResponse` | `{"data": {...}, "meta": {...}}` | ✅ OK |
| `GetUserByID` | `GET /api/v1/users/:id` | User-Service | `UserServiceResponse` | `{"data": {...}, "meta": {...}}` | ✅ OK |

#### Authentication Pattern

All methods inject OpenTelemetry trace context headers and request ID headers:

```go
// Extract request ID from context
requestID, ok := ctx.Value("request_id").(string)
if ok {
    httpReq.Header.Set("X-Request-ID", requestID)
}

// Inject trace context
otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(httpReq.Header))
```

---

### 2. Objects-Service → Auth-Service Client

**File:** `services/objects-service/internal/client/auth_client.go`

#### Client Definition

```go
type authClient struct {
    baseURL    string
    httpClient *http.Client
    logger     *logrus.Logger
}
```

#### Response Structs

| Struct | Fields | Uses Data Wrapper? | Status |
|--------|--------|-------------------|--------|
| `checkPermissionResponse` | `Allowed`, `UserID`, `Permission` | ❌ No (intentional) | ✅ OK |
| Anonymous struct (GetUserPermissions) | `Permissions []string` | ❌ No (intentional) | ✅ OK |
| Anonymous struct (GetUserRoles) | `Roles []string` | ❌ No (intentional) | ✅ OK |

#### Methods and Endpoint Mapping

| Method | Call Format | Upstream Endpoint | Response JSON | Status |
|--------|-------------|-------------------|---------------|--------|
| `CheckPermission` | `POST /api/v1/auth/permissions/check` | Auth-Service | `{"allowed": true, "user_id": "...", "permission": "..."}` | ✅ OK |
| `GetUserPermissions` | `GET /api/v1/auth/users/:user_id/permissions` | Auth-Service | `{"permissions": [...]}` | ✅ OK |
| `GetUserRoles` | `GET /api/v1/auth/users/:user_id/roles` | Auth-Service | `{"roles": [...]}` | ✅ OK |

**Note:** These endpoints intentionally use FLAT JSON format (no `data` wrapper) because they are internal service-to-service calls that prioritize performance and simplicity. See section "Special Cases: Internal Service Endpoints" below.

---

### 3. Objects-Service Generic HTTP Client

**File:** `services/objects-service/internal/client/http_client.go`

#### Client Definition

```go
type HTTPClient struct {
    baseURL    string
    httpClient *http.Client
    logger     *logrus.Logger
}
```

#### Methods

| Method | Description | Returns | Notes |
|--------|-------------|---------|-------|
| `DoRequest` | Generic HTTP request | `*http.Response` | Raw response, no parsing |
| `Get` | GET request | `*http.Response` | Raw response, no parsing |
| `Post` | POST request | `*http.Response` | Raw response, no parsing |
| `Put` | PUT request | `*http.Response` | Raw response, no parsing |
| `Delete` | DELETE request | `*http.Response` | Raw response, no parsing |

**Note:** This is a generic HTTP client that returns raw HTTP responses. Response parsing is left to the calling code. No response structs are defined.

---

### 4. CLI Client (Not Inter-Service)

**File:** `cli/internal/client/client.go`

#### Client Definition

```go
type APIClient struct {
    config     *config.Config
    httpClient *http.Client
}
```

#### Response Struct

```go
type Response struct {
    StatusCode int                 `json:"status_code"`
    Headers    map[string][]string `json:"headers,omitempty"`
    Body       interface{}         `json:"body,omitempty"`
    Error      string              `json:"error,omitempty"`
}
```

**Note:** This client is used by the CLI tool client, not for inter-service communication. It uses a generic `Body interface{}` and doesn't parse specific response formats.

---

## Special Cases: Internal Service Endpoints

Some auth-service endpoints intentionally return FLAT JSON (no `data` wrapper) to optimize internal service-to-service communication.

### Endpoints Using FLAT Format

| Endpoint | Method | Handler | Return Format | Reason |
|----------|--------|---------|---------------|--------|
| `/api/v1/auth/permissions/check` | POST | `permission_handler.go:69` | `CheckPermissionResponse` struct | Internal service call |
| `/api/v1/auth/users/:user_id/permissions` | GET | `permission_handler.go:98` | `UserPermissionsResponse` struct | Internal service call |
| `/api/v1/auth/users/:user_id/roles` | GET | `auth_handler.go:849` | `{"roles": [...]}` | Internal service call |

### Internal vs External API Consistency

**External API (gateway-facing):** Uses standardized format with `data` wrapper and `meta.request_id`
**Internal API (service-to-service):** Uses flat format for performance and simplicity

This creates a consistency gap, but it's intentional and documented. The auth-service clients (objects-service's `auth_client.go`) are designed to expect flat JSON for these endpoints.

---

## Testing Status

### Unit Tests

| Service | Handler Tests | Status | Last Updated |
|---------|--------------|--------|--------------|
| Auth-Service | `auth_handler_test.go` | ✅ Passing | Commit `e82e6bd` |
| Auth-Service | `permission_handler_test.go` | ✅ Passing | Existing tests |
| Objects-Service | `object_handler_test.go` | ✅ Passing | Existing tests |
| Objects-Service | `relationship_handler_test.go` | ✅ Passing | Existing tests |
| Objects-Service | `relationship_type_handler_test.go` | ✅ Passing | Existing tests |
| User-Service | `user_handler_test.go` | ✅ Passing | Existing tests |

### Integration Tests

| Test Script | Status | Notes |
|------------|--------|-------|
| `scripts/test-auth-flow.sh` | ⚠️ Needs update | Test script needs minor fixes for standardized error format |
| `scripts/test-rbac-relationships.sh` | ❓ Not verified | Check error responses include `type` |
| `scripts/test-rbac-objects-service.sh` | ❓ Not verified | Check success responses include `meta` |

---

## Summary Statistics

### Endpoint Coverage

| Category | Count | Status |
|----------|-------|--------|
| **Total endpoints** | ~100 | |
| **Error standardization** | 100% | ✅ All have `type` + `meta.request_id` |
| **Success standardization** | 85% | ✅ Most updated, internal endpoints intentionally flat |
| **Inter-service clients** | 4 | ✅ All aligned with upstream formats |

### Issues and Fixes

| Issue | Service | Fix Commit | Status |
|-------|---------|-----------|--------|
| `UserLoginResponse` missing `Data` wrapper | Auth-Service | `e82e6bd` | ✅ Fixed |
| Objects-Service `CheckPermission` expects wrong format | Objects-Service | N/A | ✅ No issue (correctly expects flat JSON) |
| Objects-Service `GetUserPermissions` expects wrong format | Objects-Service | N/A | ✅ No issue (correctly expects flat JSON) |
| Objects-Service `GetUserRoles` expects wrong format | Objects-Service | N/A | ✅ No issue (correctly expects flat JSON) |

---

## Recommendations

### 1. Standardize Internal Endpoints (Optional Future Work)

Consider standardizing auth-service's internal endpoints (`CheckPermission`, `GetUserPermissions`, `GetUserRoles`) to also use the `data` wrapper format. This would:
- Eliminate inconsistency between external and internal APIs
- Simplify client code across all services
- Allow unified error handling

If implemented, would require:
- Updating `permission_handler.go` and `auth_handler.go` internal endpoints
- Updating `auth_client.go` response structs to use `Data` wrapper
- Updating objects-service unit tests

### 2. Add Integration Tests

Add integration tests that verify:
- All inter-service clients correctly parse responses
- Error responses include expected fields
- Success responses follow standardized format

### 3. Document Endpoints

Add comments to each service's handlers documenting:
- Whether endpoint is external (gateway-facing) or internal (service-to-service)
- Expected response format
- Whether `data` wrapper is used

---

## Appendix: Response Format Examples

### Auth-Service Login Response (External)

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "roles": ["user"]
  },
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

### Auth-Service CheckPermission (Internal)

```json
{
  "allowed": true,
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "permission": "objects:create"
}
```

### User-Service User Response (External)

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe"
  },
  "message": "User retrieved successfully",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

### Objects-Service Object Response (External)

```json
{
  "data": {
    "id": 1,
    "name": "Sample Object",
    "public_id": "abc-def-123"
  },
  "message": "Object retrieved successfully",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

### Error Response (All Services)

```json
{
  "error": "User already exists",
  "type": "conflict",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

---

## Document Version

- **Created:** May 7, 2026
- **Last Updated:** May 7, 2026
- **Author:** Project Team
- **Related Documents:**
  - [API Response Standards](./api-response-standards.md)
  - [Implementation Plan](./api-response-standardization-plan.md)
  - [AGENTS.md](../AGENTS.md)
