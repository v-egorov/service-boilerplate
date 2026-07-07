# Known Issues & Deferred Items

Items tracked for future investigation or deferred to a later delta. Each entry has discovery date, affected change (if applicable), and the current state of understanding.

---

## Handler Error Mapping — All Service Errors Return 500

**Discovered:** 2026-07-07 (during `fix-objects-service-null-data` container verification)  
**Affected service:** objects-service (all handlers: object, object_type, relationship)

### Problem
The handler's error mapping function (`handleServiceError`) only special-cases one sentinel error (`repository.ErrOptimisticLock` → 409 Conflict). **All other errors map to HTTP 500 Internal Server Error**, including:
- `sql.ErrNoRows` (resource not found) — should be 404 or validation_error
- Invalid input parameters — should be 422 validation_error
- Business logic violations — should be specific error types, not internal_error

### Example
```
GET /api/v1/objects?object_type_id=999&limit=5   (ID doesn't exist)
→ Returns: {"error":"Internal server error","type":"internal_error"}  (HTTP 500)
→ Should be: {"error":"Object type not found","type":"not_found"}  (HTTP 404) or validation_error
```

### Root Cause
`services/objects-service/internal/handlers/object_handler.go:75-98` — `handleServiceError()` has no error-type dispatch. It checks for one sentinel, then falls through to a generic 500:
```go
if errors.Is(err, repository.ErrOptimisticLock) {
    c.JSON(http.StatusConflict, ...)
    return
}
c.JSON(http.StatusInternalServerError, gin.H{
    "error": "Internal server error",
    "type":  "internal_error",
})
```

### Impact
Clients cannot distinguish between unexpected server failures and known validation/lookup errors. This violates the API response standards which require HTTP status codes to match error types.

### Fix Strategy (future delta)
Add error-type dispatch in `handleServiceError`:
- Check for sentinel errors via `errors.Is()` (ErrNotFound, ErrInvalidInput, etc.)
- Map each to appropriate HTTP status + error type per API response standards
- Fall back to 500 only for truly unexpected errors

---

## E2E Test Script — Unverified Response Parsing

**Discovered:** 2026-07-04 (same commit as script creation)  
**Status:** May be fixed by `fix-mcp-e2e-parsing` delta

### Problem
`scripts/test-mcp-e2e.sh` was written during debugging but never run end-to-end with a clean pass. The SSE response parsing logic may still have issues — earlier attempts failed with jq errors on malformed input, and the script's own output showed empty results for steps 4–6 despite raw data being present in the SSE stream.

### Current State
- Script exists and is executable
- Initial SSE connection (Step 1–2) likely works
- Subsequent tool call responses (Steps 3–6): **unconfirmed** — may fail silently due to JSON parsing errors, missing SSE response lines, or timing issues with the `sleep` delays between POST requests and SSE stream arrivals

### Action Needed
Run `bash scripts/test-mcp-e2e.sh` on a fresh environment (or at least after restarting all containers) to verify:
1. All 7 steps pass cleanly with colored output
2. Tool call results are correctly extracted from SSE stream and parsed by jq
3. Auth-chain check in Step 7 reflects actual DB state

---

## Object Types — Hardcoded Route Configuration Gap

**Discovered:** 2026-07-04  
**Related delta:** `enforce-type-key-mandatory` (non-goal)

### Problem
Adding a new object type that fits the base schema (`objects_service.objects` table) requires manual code changes in `main.go`. Every route group is hardcoded with literal `TypeKey` values:

```go
objectTypesRead.Use(perm(permiddleware.RouteConfig{TypeKey: "object-types", HTTPMethod: "GET"}))
// ...
objectsCreate.Use(perm(permiddleware.RouteConfig{TypeKey: "objects", HTTPMethod: "POST"}))
```

This means every new type requires:
1. Running a migration to insert the `object_type` row (with its `type_key`)
2. Manually editing `main.go` route configuration — but there's no per-type routing needed since all types share the same CRUD endpoints (`/api/v1/object-types/*`, `/api/v1/objects/*`)

The real friction is not type-specific routes — it's that **adding a new object type should require zero code changes** beyond the migration. The current system works for existing types but requires manual maintenance of `main.go` when adding new ones.

### Current State
- Relationships already use CTI (separate table) and work correctly ✅
- All other object types share one flat `objects_service.objects` table with JSONB metadata
- Route groups are hardcoded in `cmd/main.go` — but they're shared across all types, so adding a new type doesn't actually require route changes (the existing `/api/v1/objects/*` endpoints already serve every type)
- The actual pain point is less about code changes and more about the **missing declarative link** between a new `object_type` row and permission entries in `auth_service.permissions`

### Deferred Action
Not part of this delta. Future work could explore:
- Auto-registering permissions for new object types based on their `type_key`
- A plugin or hook system that triggers permission/role assignments when an `object_type` is created
- Dynamic route generation from the database (overkill for current scope)

---

## Nil Slice Initialization — user-service Deferred to Follow-up Delta

**Discovered:** 2026-07-04 (during `fix-objects-service-null-data` exploration)  
**Delta:** `fix-objects-service-null-data` (objects-service only, scoped)

### Problem
Go's zero-value behavior: `var items []*Model` produces a nil slice, which gin serializes as JSON `null`. Empty result sets should return `[]` (empty array), not `null`. This is idiomatic Go best practice for collection-returning functions — non-nil slices serialize consistently in JSON, behave predictably with equality checks and third-party libraries.

### Current State
- **objects-service**: 25 occurrences across 3 repository files (`object_repository.go`, `object_type_repository.go`, `relationship_repository.go`) — all fixed by this delta ✅
- **user-service**: 1 occurrence in `services/user-service/internal/repository/user_repository.go:133`
  - `var users []*models.User` → serializes as `"data": null` when no users match a query
  - Same root cause, same fix pattern (`make([]*User, 0)`)
- **auth-service**: ✅ Already clean — uses maps and individual objects (no collection nil-slice issue)

### Delta Scope Decision
This delta fixes all 25 occurrences in objects-service only. The user-service instance is deferred to a follow-up change for these reasons:
1. Smaller, lower-risk diff — 25 changes vs 26 across two services
2. Independent deploy risk — don't want objects-service test failures blocking user-service deployment
3. Pattern reuse — once this delta ships cleanly, the same one-line fix applies to user-service with zero design decisions needed

### Action Needed (future)
One line change in `user_service/internal/repository/user_repository.go`:
```go
// Before:
var users []*models.User
// After:
users := make([]*User, 0)
```
No handler changes required — gin serialization fix is automatic.

---

## MCP Resources and Prompts — Output Schema Declarations Out of Scope

**Discovered:** 2026-07-04 (during `fix-mcp-compliance` exploration)  
**Delta:** `fix-mcp-compliance` (deferred assessment)

### Problem
The current fix adds output schemas only to the 4 MCP tools. Resources (`ReadResourceResult`) and prompts (`GetPromptResult`) use different result types that mcp-go also supports schema declarations on, but they were not included in scope.

**Resources:** The hierarchy resource returns `[]mcp.ResourceContents` (currently `TextResourceContents`). Could declare an output schema for the resource's data structure.

**Prompts:** Prompt handlers return `*GetPromptResult` containing `[]PromptMessage`. Each prompt message has structured content that could be validated against a schema.

### Assessment Needed
Before adding schemas to resources/prompts, verify:
1. Do Python SDK clients actually validate resource/prompt results? (Likely only tool call results are validated)
2. Is there a practical benefit beyond consistency with tools?
3. Would it add value for agent debugging/tracing?

**Decision:** Defer until a concrete client or use case demonstrates the need. Track here as a best-practice consideration.

---

## Resolved Items (Archived Deltas)

The following items were resolved by archived OpenSpec deltas and are retained here for reference only:

| Issue | Delta | Archive Location |
|-------|-------|-----------------|
| **Pagination Inconsistency** — endpoints used mixed `page/page_size` vs `limit/offset` | `fix-pagination-inconsistency` | `2026-07-07-fix-pagination-inconsistency` |
| **Air/Polling Mode Suspicion** — suspected Air inotify issues on Docker volumes | N/A (debunked — code-level bug) | N/A |
| **MCP Server Identity Forwarding** — identity headers lost at mcp-server → objects-service hop | `fix-mcp-auth-chain` | `2026-07-04-fix-mcp-auth-chain` |
| **Migration 00012** — user-role linkage not applied | `mcp-server-initial` (archived) | Fixed in DB |
| **Gateway SSE Panic** — `http.ErrAbortHandler` panic from reverse proxy + gin Recovery conflict | `fix-gateway-sse-and-perm-jwt` | `2026-07-04-fix-gateway-sse-and-perm-jwt` |
| **MCP Auth Chain Permission Gap** — auth-service round-trip failed for gateway-trusted requests | `fix-gateway-sse-and-perm-jwt` | `2026-07-04-fix-gateway-sse-and-perm-jwt` |
