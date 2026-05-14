# API Response Standards

## Overview

This document describes the standardized API response format used across all services in the boilerplate.

## Success Response Format

### Standard Format

All successful API responses follow this structure:

```json
{
  "data": { ... },
  "message": "Human-readable success message",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

**Fields:**
- `data` - The actual response payload (resource or collection)
- `message` - Human-readable success message (for logging/debugging)
- `meta` - Machine-readable metadata
  - `request_id` - Unique request identifier for distributed tracing

### Examples

**Create Object:**
```json
{
  "data": {
    "id": 1,
    "public_id": "fa5c27bc-7369-41d2-95b1-7561f072c863",
    "name": "Test Object",
    "status": "active"
  },
  "message": "Object created successfully",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

**Get Object:**
```json
{
  "data": {
    "id": 1,
    "public_id": "fa5c27bc-7369-41d2-95b1-7561f072c863",
    "name": "Test Object",
    "status": "active"
  },
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

**List Objects:**
```json
{
  "data": [
    { "id": 1, "name": "Object 1" },
    { "id": 2, "name": "Object 2" }
  ],
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

### Special Endpoints

**Authentication Endpoints:**

**POST /login** (returns tokens directly)
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

**POST /register** (returns message + user)
```json
{
  "message": "User registered successfully",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "created_at": "2026-04-30T16:32:16Z"
  },
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

**POST /logout** (returns success message)
```json
{
  "message": "Logged out successfully",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

**Health Check Endpoints:**

**GET /health**
```json
{
  "status": "ok",
  "timestamp": "2026-04-30T16:32:16Z",
  "service": "objects-service",
  "version": "1.0.0",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

**GET /ready**
```json
{
  "status": "ok",
  "timestamp": "2026-04-30T16:32:16Z",
  "service": "objects-service",
  "version": "1.0.0",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

**GET /ping**
```json
{
  "status": "pong",
  "timestamp": "2026-04-30T16:32:16Z",
  "service": "objects-service",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

### Format Rationale

**Why `{"data": ..., "message": ..., "meta": {...}}`?**

1. **Consistency**: Same structure across all successful responses
2. **Human-readable**: `message` field is helpful for debugging and logging
3. **Machine-readable**: `meta.request_id` enables distributed tracing
4. **Separation of concerns**:
   - `data` = business payload
   - `message` = human feedback
   - `meta` = system metadata

**Why keep current format instead of pure REST?**

- Acceptable for internal microservices
- Less breaking changes for existing clients
- Simpler client parsing (no need to adapt to strict REST)
- Human-readable messages are valuable in development
- Can migrate to JSON:API or pure REST when public API is needed

**Error Response Standards** - See below

## Error Response Format

All error responses follow this structure:

```json
{
  "error": "Human-readable error message",
  "type": "<error_type>",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

**Fields:**
- `error` - User-facing error message
- `type` - Machine-readable error type for programmatic handling
- `meta` - Machine-readable metadata
  - `request_id` - Unique request ID for distributed tracing (same as in success responses)

### HTTP Status Codes

| Status | Type | Description |
|--------|------|-------------|
| 400 | `validation_error` | Invalid input, missing required fields |
| 401 | `unauthorized` | Authentication failed |
| 403 | `permission_denied` | Authorization failed |
| 404 | `not_found` | Resource not found |
| 409 | `conflict` | Resource conflict (duplicate, version) |
| 422 | `validation_error` | Business logic validation failed |
| 500 | `internal_error` | Server error |

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

## Implementation Notes

### Adding `meta.request_id`

**How to add:**

1. Extract from request header:
```go
requestID := c.GetHeader("X-Request-ID")
```

2. Include in response:
```go
c.JSON(http.StatusCreated, gin.H{
    "data":    object,
    "message": "Object created successfully",
    "meta": gin.H{
        "request_id": requestID,
    },
})
```

**Best practices:**
- Always include `request_id` even if header is empty
- Use for troubleshooting and log correlation
- Don't expose internal system info in `meta`

### Response Patterns by Service

| Service | Create | Get | List | Delete |
|---------|--------|-----|------|--------|
| Auth-Service | `message + data` | `data` | `data` | `message` |
| Objects-Service | `data + message` | `data` | `data` | - |
| User-Service | `data + message` | `data` | `data` | `message` |

## Client Examples

### JavaScript/TypeScript

```typescript
async function createObject(data: CreateObjectRequest): Promise<Object> {
  const response = await fetch('/api/v1/objects', {
    method: 'POST',
    body: JSON.stringify(data)
  });

  if (response.status >= 400) {
    const error = await response.json();
    // Check error type for programmatic handling
    switch (error.type) {
      case 'validation_error':
        throw new ValidationError(error.error);
      case 'conflict':
        throw new ConflictError(error.error);
      default:
        throw new ServerError(error.error);
    }
  }

  const result = await response.json();
  // Access resource from 'data' field
  return result.data;
}
```

### Go

```go
func CreateObject(ctx context.Context, client *http.Client, data CreateObjectRequest) (*Object, error) {
    resp, err := client.Post("http://api/v1/objects", "application/json", body)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        var apiError struct {
            Error string `json:"error"`
            Type  string `json:"type"`
        }
        json.NewDecoder(resp.Body).Decode(&apiError)
        return nil, fmt.Errorf("%s: %s", apiError.Type, apiError.Error)
    }

    var result struct {
        Data   *Object `json:"data"`
        Meta   struct{ RequestID string } `json:"meta"`
    }
    json.NewDecoder(resp.Body).Decode(&result)

    // Use result.Meta.RequestID for logging/tracing
    return result.Data, nil
}
```

## Related Documents

- [API Response Standardization Plan](api-response-standardization-plan.md)
- [AGENTS.md - API Response Standards](../AGENTS.md#api-response-standards)