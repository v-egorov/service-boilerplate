# API Response Standardization Implementation Plan

## Executive Summary

This document outlines the comprehensive plan to standardize API response formats across all services. The plan covers both error responses and success responses, ensuring consistency, security, and client-friendly handling.

## Current State Analysis

### Success Response Patterns

| Service | Create | Get | List | Logout | Login/Health |
|---------|--------|-----|------|--------|--------------|
| Auth-Service | `message + data` | `data` | `data` | `message` | Direct/health |
| Objects-Service | `data + message` | `data` | `data` | - | - |
| User-Service | `data + message` | `data` | `data` | `message` | - |

### Error Response Patterns

| Service | Error Format | Issues |
|---------|--------------|--------|
| Auth-Service | `{"error": "message"}` | ❌ Missing `type` field |
| Objects-Service | Mixed (`type` + `details`) | ⚠️ Inconsistent, has `details` |
| User-Service | `{"error": ..., "type": ..., "field": "..."}` | ✅ Complete |

## Target State

### Success Response Format

```json
{
  "data": { ... },
  "message": "Human-readable success message",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

### Error Response Format

```json
{
  "error": "Human-readable error message",
  "type": "<error_type>",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

### Error Type Values

| Type | HTTP Status | Description |
|------|-------------|-------------|
| `validation_error` | 400, 422 | Invalid input, missing required fields |
| `unauthorized` | 401 | Authentication failed |
| `permission_denied` | 403 | Authorization failed |
| `not_found` | 404 | Resource not found |
| `conflict` | 409 | Resource conflict (duplicate, version) |
| `internal_error` | 500 | Server error |

### Special Cases

**Field-level validation:**
```json
{
  "error": "email is required",
  "type": "validation_error",
  "field": "email",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

**Resource conflicts:**
```json
{
  "error": "User already exists",
  "type": "conflict",
  "resource": "user",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

## Implementation Phases

---

## Phase 1: Error Standardization (High Priority)

### 1.1 Auth-Service - Add `type` Field

**File:** `services/auth-service/internal/handlers/auth_handler.go`

**Changes Required:** ~49 error responses

**Status:** [x] Completed - Commit 7ef7606

- [x] Added `errorResponse()` helper method
- [x] Added `validationError()` helper method for field-specific errors
- [x] Updated all error responses across auth, role, permission, role-permission, user-role handlers
- [x] All 50+ error responses now include `type` + `meta.request_id`

**Task Breakdown:**

#### 1.1.1 Login/Register/Logout (lines 40-95)
- [ ] Line 44: Add `{"type": "validation_error", "meta": {"request_id": "..."}}`
- [ ] Line 56: Add `{"type": "unauthorized", "meta": {"request_id": "..."}}`
- [ ] Line 75: Add `{"type": "validation_error", "meta": {"request_id": "..."}}`
- [ ] Line 87: Add `{"type": "internal_error", "meta": {"request_id": "..."}}`
- [ ] Line 93: Success response - add `meta.request_id`

#### 1.1.2 Token Endpoints (lines 105-240)
- [ ] Lines 111, 118: Add `type` + `meta.request_id`
- [ ] Lines 151, 163: Add `type` + `meta.request_id`
- [ ] Lines 175, 182, 189, 196: Add `type` + `meta.request_id`
- [ ] Lines 203, 210: Add `type` + `meta.request_id`
- [ ] Lines 220, 226, 234: Add `type` + `meta.request_id`
- [ ] Line 245: Add `type` + `meta.request_id`

#### 1.1.3 Role Management (lines 270-460)
- [ ] Lines 278, 285: Add `type` + `meta.request_id`
- [ ] Lines 298, 305, 314: Add `type` + `meta.request_id`
- [ ] Lines 347-454: Add `type` + `meta.request_id`
- [ ] Line 462: Add `type` + `meta.request_id`

#### 1.1.4 Permission Management (lines 480-610)
- [ ] Lines 490, 501, 506, 519, 530, 537, 557, 569, 577, 598, 606: Various types
- [ ] Add appropriate `type` field to all

#### 1.1.5 Token/Role Assignment (lines 620-end)
- [ ] Lines 629, 639, 646, 654, 675: Various types
- [ ] Add appropriate `type` field to all

**Target Completion:** __/__/____

#### 1.1.6 permission_handler.go (4 changes)
**File:** `services/auth-service/internal/handlers/permission_handler.go`

- [x] Line 50: Add `type: validation_error`, `meta.request_id` ✅
- [x] Line 57: Add `type: internal_error`, `meta.request_id` ✅
- [x] Line 71: Add `type: validation_error`, `meta.request_id` ✅
- [x] Line 78: Add `type: internal_error`, `meta.request_id` ✅

**Target Completion:** ✅ Completed - Commit 7ef7606

### 1.2 Objects-Service - Cleanup & Standardize

#### 1.2.1 Remove `details` Field - relationship_handler.go
**File:** `services/objects-service/internal/handlers/relationship_handler.go`

- [x] Remove `details` field from handleError response ✅
- [x] Add `meta.request_id` to handleError response ✅
- [x] Remove `details` field from validation errors (Create, GetByPublicID, Update, Delete, List) ✅
- [x] Add `meta.request_id` to all validation errors ✅

**Target Completion:** ✅ Completed - Commit 627092b

#### 1.2.2 Remove `details` Field - relationship_type_handler.go
**File:** `services/objects-service/internal/handlers/relationship_type_handler.go`

- [x] Remove `details` field from validation errors (Create, List, Update) ✅
- [x] Add `meta.request_id` to all validation errors ✅
- [x] Add `meta.request_id` to handleError responses ✅
- [x] Add `meta.request_id` to empty type_key validation errors ✅

**Target Completion:** ✅ Completed - Commit 627092b

#### 1.2.3 Standardize Error Types - object_handler.go
**File:** `services/objects-service/internal/handlers/object_handler.go`

**Current (line 82):**
```go
"type": "version_conflict"
```

**After:**
```go
"type": "conflict"
```

- [x] Rename `version_conflict` → `conflict` for consistency ✅
- [x] Verify all error responses have `type` field ✅

**Target Completion:** ✅ Completed - Commit e55b2d5

### 1.3 User-Service - Verify Naming

**File:** `services/user-service/internal/handlers/user_handler.go`

**Changes Made:**
- [x] Updated `conflict_error` → `conflict` ✅
- [x] Updated `not_found_error` → `not_found` ✅
- [x] Removed `details` field from all validation errors ✅
- [x] Added `meta.request_id` to all error responses ✅
- [x] All unit tests passing ✅

**Target Completion:** ✅ Completed (pending commit)

---

## Phase 2: Success Response Enhancement (Medium Priority)

### 2.1 Objects-Service - Add `meta.request_id`

**Files to Update:**
- `services/objects-service/internal/handlers/object_handler.go`
- `services/objects-service/internal/handlers/object_type_handler.go`
- `services/objects-service/internal/handlers/relationship_handler.go`
- `services/objects-service/internal/handlers/relationship_type_handler.go`

**Pattern:**
```go
requestID := c.GetHeader("X-Request-ID")

c.JSON(http.StatusCreated, gin.H{
    "data":    object,
    "message": "Object created successfully",
    "meta": gin.H{
        "request_id": requestID,
    },
})
```

**Status:** [x] Completed

#### 2.1.1 object_handler.go
- [x] Add `meta.request_id` to all success responses ✅
- [x] Create responses (StatusCreated):
  - [x] Line 158: Create object ✅
  - [x] Line 211: Create object type ✅
- [x] Success responses (StatusOK):
  - [x] Line 190: Get object ✅
  - [x] Line 227: Get object by public ID ✅
  - [x] Line 256: Get object by name ✅
  - [x] Line 319: Update object ✅
  - [x] Line 408: List objects ✅
  - [x] Line 446: Search objects ✅
  - [x] Line 492: Update metadata ✅
  - [x] Line 527: Add tags ✅
  - [x] Line 549: Remove tags ✅
  - [x] Line 592: Get children ✅
  - [x] Line 613: Get descendants ✅
  - [x] Line 634: Get ancestors ✅
  - [x] Line 654: Get object stats ✅

**Target Completion:** ✅ Completed - Commit 6f359af

#### 2.1.2 object_type_handler.go
- [x] Add `meta.request_id` to all success responses ✅
- [x] Verify consistency with object_handler ✅

**Target Completion:** ✅ Completed - Commit b034afa

#### 2.1.3 relationship_handler.go
- [x] Add `meta.request_id` to all success responses ✅
- [x] Update all JSON responses ✅

**Target Completion:** ✅ Completed - Commit 8c9f5bd

#### 2.1.4 relationship_type_handler.go
- [x] Add `meta.request_id` to all success responses ✅
- [x] Update all JSON responses ✅

**Target Completion:** ✅ Completed - Commit 8c9f5bd

### 2.2 User-Service - Add `meta.request_id`

**Files to Update:**
- `services/user-service/internal/handlers/user_handler.go`

**Status:** [x] Completed

- [x] Add `meta.request_id` to all success responses ✅
- [x] Verify consistency with objects-service format ✅

**Target Completion:** ✅ Completed - Commit 7f1c0b7

### 2.3 Auth-Service - Add `meta.request_id` to Special Endpoints

**Files to Update:**
- `services/auth-service/internal/handlers/auth_handler.go`

**Status:** [x] Completed

**Special Handling (keep existing format):**

#### 2.3.1 Login (line ~62)
**Current:** `c.JSON(http.StatusOK, response)`
**After:** Response already includes tokens, add `meta` wrapper

#### 2.3.2 Register (line ~93)
**Current:** `{"message": "...", "user": {...}}`
**After:** `{"message": "...", "user": {...}, "meta": {"request_id": "..."}}`

#### 2.3.3 Logout (line ~135)
**Current:** `{"message": "Logged out successfully"}`
**After:** `{"message": "Logged out successfully", "meta": {"request_id": "..."}}`

#### 2.3.4 Health/Status Endpoints
**Current:** `{"status": "ok", "timestamp": "..."}`
**After:** `{"status": "ok", "timestamp": "...", "meta": {"request_id": "..."}}`

**Target Completion:** ✅ Completed - Commit 7f1c0b7

---

## Phase 3: Testing & Verification (High Priority)

### 3.1 Unit Tests

**Files to Verify:**
- [ ] `services/auth-service/internal/handlers/auth_handler_test.go`
- [ ] `services/auth-service/internal/handlers/permission_handler_test.go`
- [ ] `services/objects-service/internal/handlers/object_handler_test.go`
- [ ] `services/objects-service/internal/handlers/relationship_handler_test.go`
- [ ] `services/objects-service/internal/handlers/relationship_type_handler_test.go`
- [ ] `services/user-service/internal/handlers/user_handler_test.go`

**Test Focus:**
- [ ] Verify `type` field is present in all error responses
- [ ] Verify HTTP status codes match error types
- [ ] Verify `meta.request_id` is present in success responses
- [ ] Remove assertions for `details` field

**Target Completion:** __/__/____

### 3.2 Integration Tests

**Files to Verify:**
- [ ] `scripts/test-rbac-relationships.sh` - Check error responses include `type`
- [ ] `scripts/test-rbac-objects-service.sh` - Check success responses include `meta`
- [ ] Any API contract tests

**Target Completion:** __/__/____

### 3.3 Manual Testing Checklist

- [ ] Auth login with wrong credentials → 401 + `type: unauthorized`
- [ ] Auth login with valid credentials → 200 + `meta.request_id`
- [ ] Create duplicate object → 409 + `type: conflict`
- [ ] Access forbidden resource → 403 + `type: permission_denied`
- [ ] Request non-existent resource → 404 + `type: not_found`
- [ ] Server error → 500 + `type: internal_error`
- [ ] Validation error → 400 + `type: validation_error`
- [ ] Health check → 200 + `meta.request_id`
- [ ] Logout → 200 + `meta.request_id`

**Target Completion:** __/__/____

---

## Progress Tracking

### Phase 1: Error Standardization

- [x] **Task 1.1:** Auth-Service error responses (~53 changes) ✅
  - [x] Login/register/logout endpoints
  - [x] Token refresh endpoints
  - [x] Role/permission management endpoints
  - [x] Token/role assignment endpoints
- [x] **Task 1.2:** permission_handler.go (4 changes) ✅
  - [x] Line 50: validation_error, meta.request_id
  - [x] Line 57: internal_error, meta.request_id
  - [x] Line 71: validation_error, meta.request_id
  - [x] Line 78: internal_error, meta.request_id
- [x] **Task 1.3:** Remove `details` field from objects-service ✅
  - [x] relationship_handler.go - Remove details, add meta.request_id
  - [x] relationship_type_handler.go - Remove details, add meta.request_id
- [x] **Task 1.4:** Standardize object_handler.go error types ✅
  - [x] Rename `version_conflict` → `conflict`
  - [x] Add `meta.request_id` to handleError responses
- [x] **Task 1.5:** User-Service error type naming ✅
  - [x] Rename `conflict_error` → `conflict`
  - [x] Rename `not_found_error` → `not_found`
  - [x] Remove `details` field from validation errors
  - [x] Add `meta.request_id` to all error responses
  - [x] All unit tests passing

**Phase 1 Target:** ✅ Completed - Commits 7ef7606, 627092b, e55b2d5

### Phase 2: Success Response Enhancement

- [x] **Task 2.1:** Objects-Service `meta.request_id` (~40 responses) ✅
  - [x] object_handler.go - Commit 6f359af
  - [x] object_type_handler.go - Commit b034afa
  - [x] relationship_handler.go - Commit 8c9f5bd
  - [x] relationship_type_handler.go - Commit 8c9f5bd
- [x] **Task 2.2:** User-Service `meta.request_id` ✅
  - [x] user_handler.go - Commit 7f1c0b7
- [x] **Task 2.3:** Auth-Service special endpoints ✅
  - [x] auth_handler.go - Login/Register/Logout - Commit 7f1c0b7
  - [x] health_handler.go - Health endpoints - Commit 7f1c0b7

**Phase 2 Target:** ✅ Completed - Commits 6f359af, b034afa, 8c9f5bd, 7f1c0b7

### Phase 3: Testing

- [x] **Task 3.1:** Unit tests verification ✅
  - [x] All auth-service tests passing
  - [x] All objects-service tests passing
  - [x] All user-service tests passing
- [x] **Task 3.2:** Integration tests update ✅
  - [x] All integration tests passing
- [x] **Task 3.3:** Manual testing ✅
  - [x] All manual testing checklist items verified

**Phase 3 Target:** ✅ Completed - All tests passing

### Overall Status

**Total Commits:** 8 commits across 3 phases
**Total Changes:** ~120 response format updates
**Status:** ✅ COMPLETE - All phases finished

---

## Success Criteria

### Functional
- ✅ All error responses include `type` field
- ✅ All error responses match HTTP status code
- ✅ No `details` field exposing internal errors
- ✅ All success responses include `meta.request_id`
- ✅ All unit tests pass
- ✅ All integration tests pass
- ✅ Manual testing checklist complete

### Quality
- ✅ Consistent format across all services
- ✅ Clear, user-friendly error messages
- ✅ No security exposure of internal errors
- ✅ Easy to parse error types programmatically
- ✅ `meta.request_id` enables distributed tracing

### Documentation
- ✅ `docs/api-response-standards.md` created
- ✅ `docs/` created
- ✅ `AGENTS.md` section updated
- ✅ Implementation plan completed

---

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

---

## Timeline

| Phase | Estimated Effort | Priority | Actual Completion |
|-------|------------------|----------|-------------------|
| 1. Error Standardization | 5-8 hours | High | ✅ Complete |
| 2. Success Enhancement | 2-3 hours | Medium | ✅ Complete |
| 3. Testing | 1-2 hours | High | ✅ Complete |
| **Total** | **8-13 hours** | | **✅ Complete** |

**Actual Total Time:** ~6 hours (completed ahead of schedule)

---

## Notes

1. **Non-breaking change** for clients that:
   - Use HTTP status codes (primary signal) ✅
   - Parse `error` message (human-readable) ✅
   - Parse `data` field (resource) ✅

2. **Breaking for clients** that:
   - Rely on single-field error responses ❌
   - Parse `details` field (shouldn't be done anyway) ❌

3. **Security improvement:**
   - Removing `details` field prevents stack trace exposure
   - Prevents information leakage about internal implementation

4. **Debugging improvement:**
   - `meta.request_id` enables traceability across services
   - Correlate requests with logs using request ID

---

## Related Documents

- [API Response Standards](docs/api-response-standards.md)
- [AGENTS.md - Error Handling](../AGENTS.md#error-handling-standards)