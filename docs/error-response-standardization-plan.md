# Error Response Standardization - Implementation Plan

## Executive Summary

This document outlines the comprehensive plan to standardize error response formats across all services in the boilerplate. The goal is to ensure consistent, secure, and client-friendly error handling.

## Current State Analysis

### Audit Summary

| Service | Files | Error Responses | Issues |
|---------|-------|-----------------|--------|
| Auth-Service | 3 handlers | ~50+ | ❌ Missing `type` field |
| Objects-Service | 4 handlers | ~20+ | ⚠️ Inconsistent `type` usage, ❌ `details` field |
| User-Service | 2 handlers | ~10+ | ✅ Complete (reference implementation) |

### Problem Areas

#### 1. Auth-Service (High Priority)
- **File:** `auth-service/internal/handlers/auth_handler.go`
- **Issue:** All error responses use single-field format: `{"error": "message"}`
- **Impact:** Cannot distinguish error types programmatically
- **Fix:** Add `type` field to all error responses

#### 2. Objects-Service - Relationship Handler (High Priority)
- **File:** `objects-service/internal/handlers/relationship_handler.go`
- **Issue:** `details` field exposes internal error chains
- **Security Risk:** Stack traces and internal paths exposed
- **Fix:** Remove `details` field

#### 3. Objects-Service - Relationship Type Handler (Medium Priority)
- **File:** `objects-service/internal/handlers/relationship_type_handler.go`
- **Issue:** `details` field present
- **Fix:** Remove `details` field

#### 4. Objects-Service - Object Handler (Medium Priority)
- **File:** `objects-service/internal/handlers/object_handler.go`
- **Issue:** Only handles `repository.ErrOptimisticLock` explicitly
- **Fix:** Add handling for other common service errors

## Target State

### Standard Error Format

```json
{
  "error": "Human-readable error message",
  "type": "<error_type>"
}
```

### Error Type Mapping

| HTTP Status | Type | Example |
|-------------|------|---------|
| 400 | `validation_error` | "Invalid request body" |
| 401 | `unauthorized` | "Invalid credentials" |
| 403 | `permission_denied` | "Insufficient permissions" |
| 404 | `not_found` | "User not found" |
| 409 | `conflict` | "User already exists" |
| 422 | `validation_error` | "Cardinality constraint violated" |
| 500 | `internal_error` | "Internal server error" |

## Implementation Tasks

### Phase 1: Auth-Service Standardization

#### 1.1 Update auth_handler.go
**File:** `services/auth-service/internal/handlers/auth_handler.go`

**Changes Required:**

| Line | Current | Target |
|------|---------|--------|
| 44 | `{"error": "Invalid request format"}` | `{"error": "Invalid request format", "type": "validation_error"}` |
| 56 | `{"error": "Invalid credentials"}` | `{"error": "Invalid credentials", "type": "unauthorized"}` |
| 75 | `{"error": "Invalid request format"}` | `{"error": "Invalid request format", "type": "validation_error"}` |
| 87 | `{"error": "Registration failed"}` | `{"error": "Registration failed", "type": "internal_error"}` |
| 111 | `{"error": "Authorization header required"}` | `{"error": "Authorization header required", "type": "unauthorized"}` |
| 118 | `{"error": "Invalid authorization header format"}` | `{"error": "Invalid authorization header format", "type": "unauthorized"}` |
| 151 | `{"error": "Invalid request format"}` | `{"error": "Invalid request format", "type": "validation_error"}` |
| 163 | `{"error": "Invalid refresh token"}` | `{"error": "Invalid refresh token", "type": "unauthorized"}` |
| 175 | `{"error": "User not authenticated"}` | `{"error": "User not authenticated", "type": "unauthorized"}` |
| 182 | `{"error": "Internal server error"}` | `{"error": "Internal server error", "type": "internal_error"}` |
| 189 | `{"error": "Internal server error"}` | `{"error": "Internal server error", "type": "internal_error"}` |
| 196 | `{"error": "User email not found"}` | `{"error": "User email not found", "type": "not_found"}` |
| 203 | `{"error": "Internal server error"}` | `{"error": "Internal server error", "type": "internal_error"}` |
| 210 | `{"error": "Failed to get user information"}` | `{"error": "Failed to get user information", "type": "internal_error"}` |
| 220 | `{"error": "Authorization header required"}` | `{"error": "Authorization header required", "type": "unauthorized"}` |
| 226 | `{"error": "Invalid authorization header format"}` | `{"error": "Invalid authorization header format", "type": "unauthorized"}` |
| 234 | `{"error": "Invalid or revoked token"}` | `{"error": "Invalid or revoked token", "type": "unauthorized"}` |
| 245 | `{"error": "Failed to get public key"}` | `{"error": "Failed to get public key", "type": "internal_error"}` |
| 278 | `{"error": "Admin privileges required"}` | `{"error": "Admin privileges required", "type": "permission_denied"}` |
| 285 | `{"error": "Failed to rotate keys"}` | `{"error": "Failed to rotate keys", "type": "internal_error"}` |
| 298 | `{"error": "Authorization header required"}` | `{"error": "Authorization header required", "type": "unauthorized"}` |
| 305 | `{"error": "Invalid authorization header format"}` | `{"error": "Invalid authorization header format", "type": "unauthorized"}` |
| 314 | `{"error": "Invalid or expired token"}` | `{"error": "Invalid or expired token", "type": "unauthorized"}` |
| 347 | `{"error": "Invalid request data"}` | `{"error": "Invalid request data", "type": "validation_error"}` |
| 358 | `{"error": "Role with this name already exists"}` | `{"error": "Role with this name already exists", "type": "conflict"}` |
| 363 | `{"error": "Failed to create role"}` | `{"error": "Failed to create role", "type": "internal_error"}` |
| 376 | `{"error": "Failed to list roles"}` | `{"error": "Failed to list roles", "type": "internal_error"}` |
| 387 | `{"error": "Invalid role ID"}` | `{"error": "Invalid role ID", "type": "validation_error"}` |
| 394 | `{"error": "Role not found"}` | `{"error": "Role not found", "type": "not_found"}` |
| 414 | `{"error": "Invalid role ID"}` | `{"error": "Invalid role ID", "type": "validation_error"}` |
| 425 | `{"error": "Invalid request data"}` | `{"error": "Invalid request data", "type": "validation_error"}` |
| 433 | `{"error": "Failed to update role"}` | `{"error": "Failed to update role", "type": "internal_error"}` |
| 454 | `{"error": "Invalid role ID"}` | `{"error": "Invalid role ID", "type": "validation_error"}` |
| 462 | `{"error": err.Error()}` | `{"error": err.Error(), "type": "internal_error"}` |
| 490 | `{"error": "Invalid request data"}` | `{"error": "Invalid request data", "type": "validation_error"}` |
| 501 | `{"error": "Permission with this name already exists"}` | `{"error": "Permission with this name already exists", "type": "conflict"}` |
| 506 | `{"error": "Failed to create permission"}` | `{"error": "Failed to create permission", "type": "internal_error"}` |
| 519 | `{"error": "Failed to list permissions"}` | `{"error": "Failed to list permissions", "type": "internal_error"}` |
| 530 | `{"error": "Invalid permission ID"}` | `{"error": "Invalid permission ID", "type": "validation_error"}` |
| 537 | `{"error": "Permission not found"}` | `{"error": "Permission not found", "type": "not_found"}` |
| 557 | `{"error": "Invalid permission ID"}` | `{"error": "Invalid permission ID", "type": "validation_error"}` |
| 569 | `{"error": "Invalid request data"}` | `{"error": "Invalid request data", "type": "validation_error"}` |
| 577 | `{"error": "Failed to update permission"}` | `{"error": "Failed to update permission", "type": "internal_error"}` |
| 598 | `{"error": "Invalid permission ID"}` | `{"error": "Invalid permission ID", "type": "validation_error"}` |
| 606 | `{"error": err.Error()}` | `{"error": err.Error(), "type": "internal_error"}` |
| 629 | `{"error": "Invalid role ID"}` | `{"error": "Invalid role ID", "type": "validation_error"}` |
| 639 | `{"error": "Invalid request data"}` | `{"error": "Invalid request data", "type": "validation_error"}` |
| 646 | `{"error": "Invalid permission ID"}` | `{"error": "Invalid permission ID", "type": "validation_error"}` |
| 654 | `{"error": "Failed to assign permission to role"}` | `{"error": "Failed to assign permission to role", "type": "internal_error"}` |
| 675 | `{"error": "Invalid role ID"}` | `{"error": "Invalid role ID", "type": "validation_error"}` |

**Total changes:** ~45 error responses to update

#### 1.2 Update permission_handler.go
**File:** `services/auth-service/internal/handlers/permission_handler.go`

**Changes Required:**

| Line | Current | Target |
|------|---------|--------|
| 50 | `{"error": "invalid request"}` | `{"error": "invalid request", "type": "validation_error"}` |
| 57 | `{"error": "permission check failed"}` | `{"error": "permission check failed", "type": "internal_error"}` |
| 71 | `{"error": "user_id required"}` | `{"error": "user_id required", "type": "validation_error"}` |
| 78 | `{"error": "failed to get permissions"}` | `{"error": "failed to get permissions", "type": "internal_error"}` |

**Total changes:** 4 error responses to update

### Phase 2: Objects-Service Cleanup

#### 2.1 Remove details field - relationship_handler.go
**File:** `services/objects-service/internal/handlers/relationship_handler.go`

**Lines to update:** 344-348

**Before:**
```go
c.JSON(statusCode, gin.H{
    "error":   errorMessage,
    "details": err.Error(),
    "type":    errorType,
})
```

**After:**
```go
c.JSON(statusCode, gin.H{
    "error": errorMessage,
    "type":  errorType,
})
```

#### 2.2 Remove details field - relationship_type_handler.go
**File:** `services/objects-service/internal/handlers/relationship_type_handler.go`

**Lines to update:** Multiple (bad request handlers)

**Pattern:**
```go
// Before
c.JSON(http.StatusBadRequest, gin.H{
    "error":   "Invalid request format",
    "details": err.Error(),
    "type":    "validation_error",
})

// After
c.JSON(http.StatusBadRequest, gin.H{
    "error": "Invalid request format",
    "type":  "validation_error",
})
```

#### 2.3 Expand error handling - object_handler.go
**File:** `services/objects-service/internal/handlers/object_handler.go`

**Current handleServiceError (lines 74-93):**
```go
func (h *ObjectHandler) handleServiceError(c *gin.Context, err error, operation string, requestID string) {
    // ...
    switch err {
    case nil:
        return
    case repository.ErrOptimisticLock:
        c.JSON(http.StatusConflict, gin.H{
            "error": "Version conflict - the object has been modified by another request",
            "type":  "version_conflict",  // Note: non-standard type name
        })
        return
    default:
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Internal server error",
            "type":  "internal_error",
        })
    }
}
```

**Proposed changes:**
1. Rename `version_conflict` → `conflict` for consistency
2. Add handling for `repository.ErrNotFound` (if needed)
3. Keep generic fallback for other errors

### Phase 3: User-Service Verification

#### 3.1 Verify error type naming
**File:** `services/user-service/internal/handlers/user_handler.go`

**Check:**
- `validation_error` ✅ (standard)
- `conflict_error` ⚠️ → Should be `conflict`
- `not_found_error` ⚠️ → Should be `not_found`
- `internal_error` ✅ (standard)
- `unknown_error` ✅ (acceptable for fallback)

**If changes needed:** Update error type strings to match standards.

## Testing Plan

### Unit Tests

**Files to verify:**
- `services/auth-service/internal/handlers/auth_handler_test.go`
- `services/auth-service/internal/handlers/permission_handler_test.go`
- `services/objects-service/internal/handlers/object_handler_test.go`
- `services/objects-service/internal/handlers/relationship_handler_test.go`
- `services/objects-service/internal/handlers/relationship_type_handler_test.go`
- `services/user-service/internal/handlers/user_handler_test.go`

**Test focus:**
1. Verify `type` field is present in error responses
2. Verify HTTP status codes match error types
3. Remove assertions for `details` field (if present)

### Integration Tests

**Files to verify:**
- `scripts/test-rbac-relationships.sh` - Check error responses
- Any API contract tests

### Manual Testing

**Checklist:**
- [ ] Auth login with wrong credentials → 401 + type field
- [ ] Auth login with valid credentials → 200
- [ ] Create duplicate object → 409 + type field
- [ ] Access forbidden resource → 403 + type field
- [ ] Request non-existent resource → 404 + type field
- [ ] Server error scenario → 500 + type field

## Progress Tracking

### Phase 1: Auth-Service Standardization

- [ ] **Task 1.1:** Update auth_handler.go error responses (~45 changes)
  - [ ] Login/register/logout endpoints (lines 40-240)
  - [ ] Token refresh endpoints (lines 290-320)
  - [ ] Role management endpoints (lines 340-460)
  - [ ] Permission management endpoints (lines 480-610)
  - [ ] Token management endpoints (lines 620-end)
- [ ] **Task 1.2:** Update permission_handler.go (4 changes)

**Phase 1 Completion:** __/__/____

### Phase 2: Objects-Service Cleanup

- [ ] **Task 2.1:** Remove `details` from relationship_handler.go
- [ ] **Task 2.2:** Remove `details` from relationship_type_handler.go
- [ ] **Task 2.3:** Update object_handler.go error types

**Phase 2 Completion:** __/__/____

### Phase 3: User-Service Verification

- [ ] **Task 3.1:** Verify and align error type naming

**Phase 3 Completion:** __/__/____

### Phase 4: Testing & Documentation

- [ ] **Task 4.1:** Run all unit tests
- [ ] **Task 4.2:** Update integration tests if needed
- [ ] **Task 4.3:** Verify test scripts work with new format
- [ ] **Task 4.4:** Update documentation

**Phase 4 Completion:** __/__/____

## Rollback Plan

If issues arise during implementation:

1. **Identify problematic changes:**
   - Check which error responses are causing issues
   - Verify HTTP status codes are still correct

2. **Partial rollback:**
   - Revert changes to specific handlers
   - Keep changes to other handlers if working

3. **Communication:**
   - Document any breaking changes
   - Update client migration guides if needed

## Success Criteria

### Functional
- ✅ All error responses include `type` field
- ✅ HTTP status codes match error types
- ✅ No `details` field exposing internal errors
- ✅ All unit tests pass
- ✅ All integration tests pass

### Quality
- ✅ Consistent error format across all services
- ✅ Clear, user-friendly error messages
- ✅ No security exposure of internal errors
- ✅ Easy to parse error types programmatically

### Documentation
- ✅ AGENTS.md section added
- ✅ Full standards document in docs/
- ✅ Migration guide available if needed

## Timeline

| Phase | Estimated Effort | Priority |
|-------|------------------|----------|
| 1. Auth-Service | 2-3 hours | High |
| 2. Objects-Service | 1-2 hours | High |
| 3. User-Service | 30 minutes | Medium |
| 4. Testing | 1-2 hours | High |
| **Total** | **5-8 hours** | |

## Notes

1. **This is a non-breaking change** for clients that:
   - Use HTTP status codes (primary signal)
   - Parse `error` message (human-readable)

2. **Breaking for clients** that:
   - Rely on single-field error responses
   - Parse `details` field (shouldn't be done anyway)

3. **Security improvement:**
   - Removing `details` field prevents stack trace exposure
   - Prevents information leakage about internal implementation

## Related Documents

- [API Error Response Standards](api-error-response-standards.md)
- [AGENTS.md - Error Handling](../AGENTS.md#error-handling-standards)
- [Service Patterns Reference](service-patterns-reference.md)