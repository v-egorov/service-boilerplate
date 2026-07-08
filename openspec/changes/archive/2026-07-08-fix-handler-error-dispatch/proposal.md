## Why

Every service error in objects-service maps to HTTP 500 Internal Server Error regardless of the actual error type, violating API response standards that require HTTP status codes to match error types. A missing object returns 500 instead of 404, a validation failure returns 500 instead of 422/400. This makes it impossible for clients (including the MCP server) to distinguish between expected errors and unexpected failures.

## What Changes

- Extract a shared `handleError` dispatcher in objects-service handlers that maps sentinel errors to correct HTTP status codes
- Replace the broken dispatch logic across all 4 handler files (`object_handler.go`, `object_type_handler.go`, `relationship_handler.go`, `relationship_type_handler.go`) with calls to the shared dispatcher
- Add error-type coverage for all service-layer and repository-layer sentinels (~23 errors across 5 methods)

## Capabilities

### Modified Capabilities
- **objects-service**: Error handling — handlers now dispatch sentinel errors to correct HTTP status codes (404, 409, 422, 400) instead of defaulting to 500 for all errors

## Impact

**Affected code:**
- `services/objects-service/internal/handlers/object_handler.go` (~25 error sites)
- `services/objects-service/internal/handlers/object_type_handler.go` (~15 error sites)
- `services/objects-service/internal/handlers/relationship_handler.go` (~8 error sites)
- `services/objects-service/internal/handlers/relationship_type_handler.go` (~6 error sites)

**No spec changes for external API consumers** — the behavior change (404 instead of 500 on missing resources) is a correctness fix, not a new feature. Existing callers that check status codes will get more accurate responses.
