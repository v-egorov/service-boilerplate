# API Error Response Standards

## Overview

This document describes the standardized error response format used across all services in the boilerplate.

## Current State

### Existing Patterns

| Service | Pattern | Consistency |
|---------|---------|-------------|
| Auth-Service | `{"error": "message"}` | ⚠️ Missing `type` field |
| Objects-Service (Object) | `{"error": "message"}` or `{"error": "message", "type": "..."}` | ⚠️ Inconsistent |
| Objects-Service (Relationship) | `{"error": "message", "type": "...", "details": "chain"}` | ⚠️ Has `details` field |
| User-Service | `{"error": "message", "type": "...", "field": "...", "resource": "..."}` | ✅ Complete |

### Issues Identified

1. **Missing `type` field** - Auth-service returns only error messages without type identifiers
2. **Inconsistent structure** - `type` field exists in some handlers, missing in others
3. **`details` field exposure** - Relationship handlers expose internal error chains (security risk)
4. **No unified error types** - Different services use different error type naming conventions

## Standardization Plan

### Target Format

All error responses should follow this structure:

```json
{
  "error": "Human-readable error message",
  "type": "<error_type>"
}
```

#### HTTP Status Codes

| Status | Type | Description |
|--------|------|-------------|
| 400 | `validation_error` | Invalid input, missing required fields |
| 401 | `unauthorized` | Authentication failed |
| 403 | `permission_denied` | Authorization failed |
| 404 | `not_found` | Resource not found |
| 409 | `conflict` | Resource conflict (duplicate, version conflict) |
| 422 | `validation_error` | Business logic validation failed |
| 500 | `internal_error` | Server error |

### Error Type Values

```
- validation_error
- unauthorized
- permission_denied
- not_found
- conflict
- internal_error
- request_error (for malformed requests)
```

### Special Cases

#### Field-Level Validation

```json
{
  "error": "email is required",
  "type": "validation_error",
  "field": "email"
}
```

#### Resource Conflicts

```json
{
  "error": "User with this email already exists",
  "type": "conflict",
  "resource": "user"
}
```

**Note:** Only add `field` or `resource` when it provides meaningful client information.

## Implementation Tasks

### Phase 1: Documentation ✅
- [x] Document existing error patterns
- [x] Create AGENTS.md reference

### Phase 2: Auth-Service Standardization
- [ ] Update `auth_handler.go` - Add `type` field to all error responses
- [ ] Update `permission_handler.go` - Add `type` field to error responses

### Phase 3: Objects-Service Cleanup
- [ ] Remove `details` field from `relationship_handler.go` handleError
- [ ] Remove `details` field from `relationship_type_handler.go`
- [ ] Update `object_handler.go` - Handle more service error types with appropriate `type`

### Phase 4: User-Service Alignment
- [ ] Verify `user_handler.go` error types match standard
- [ ] Remove overly verbose fields from error responses if needed

### Phase 5: Testing & Verification
- [ ] Run unit tests for all services
- [ ] Update integration tests if needed
- [ ] Verify error handling in test scripts

## Migration Notes

### Breaking Changes

**None.** The changes are additive (adding `type` field) or removals of internal implementation details (`details` field). Clients that:

- Rely on HTTP status codes: ✅ No impact
- Parse error body for messages: ✅ No impact
- Use `type` field: ✅ Will gain consistency

### Implementation Priority

**High:**
1. Auth-Service - Add `type` field (affects most endpoints)
2. Remove `details` field (security improvement)

**Medium:**
3. Objects-Service - Standardize `type` field usage
4. Update tests to verify new error structure

**Low:**
5. Documentation updates
6. Client migration guides (if needed)

## References

- [HTTP Status Codes RFC 7231](https://datatracker.ietf.org/doc/html/rfc7231)
- [JSON:API Error Format](https://jsonapi.org/format/#error-objects)
- [AGENTS.md - Error Handling](../AGENTS.md#error-handling-standards)