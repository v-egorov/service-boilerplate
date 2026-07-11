# Known Issues & Deferred Items

Items tracked for future investigation or deferred to a later delta. Each entry has discovery date, affected change (if applicable), and the current state of understanding.

---

## Repository `GetByID()` Wraps `sql.ErrNoRows` Instead of Returning `ErrNotFound`

**Discovered:** 2026-07-08 (during `fix-handler-error-dispatch` implementation)  
**Priority:** Medium — causes inconsistent error wrapping that the handler dispatcher must work around

### Problem
The repository layer's `GetByID()` method does **not** check for `sql.ErrNoRows`. When a query returns zero rows, pgx propagates the raw `sql.ErrNoRows` directly wrapped with `%w`: `fmt.Errorf("failed to get object: %w", err)`. The service layer then checks `err == repository.ErrNotFound` (direct pointer comparison), which **fails** because direct equality doesn't unwrap.

```go
// Repository (object_repository.go:208-210)
if err != nil {
    return nil, fmt.Errorf("failed to get object: %w", err)  // ErrNoRows wrapped directly!
}

// Service (object_service.go:93)
if err == repository.ErrNotFound {  // Direct comparison — FAILS on wrapped error
    return nil, fmt.Errorf("object not found: %w", err)
}
```

### Impact
- The handler dispatcher must include a `sql.ErrNoRows` fallback case to catch these unwrapped errors (workaround added in `error.go`)
- This pattern is **consistent across all repository methods** that use direct comparisons instead of `errors.Is()`
- Service layers that DO use `errors.Is(err, sql.ErrNoRows)` then wrap with their own sentinel → dispatcher catches those correctly
- But the gap means some errors reach handlers as double-wrapped `sql.ErrNoRows` without a service-layer sentinel

### Root Cause
Inconsistent error conversion patterns across repository methods:
- `GetByPublicID()` checks `errors.Is(err, sql.ErrNoRows)` → returns `repository.ErrNotFound` ✅
- `GetByID()` does NOT check — passes raw `sql.ErrNoRows` through ❌

### Fix (future)
Each repository `Get*()` method should convert `sql.ErrNoRows` to `repository.ErrNotFound` consistently. Currently only some do. This would eliminate the need for the `sql.ErrNoRows` fallback in the handler dispatcher.

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

## JSON Validation Errors Lack Field-Level Detail

**Discovered:** 2026-07-08 (during `fix-handler-error-dispatch` exploration)  
**Priority:** Low

### Problem
When `ShouldBindJSON()` fails, gin returns a generic error without field-level path information. The handler catches this and returns a 400 with no indication of which field or why parsing failed.

```go
// Current — loses field context
var req models.CreateObjectRequest
if err := c.ShouldBindJSON(&req); err != nil {
    // Returns: {"error": "invalid request body", "type": "validation_error"}
    // No idea if it's missing 'name', invalid object_type_id, etc.
}
```

### Current State
- gin's `ShouldBindJSON` uses Go's `encoding/json` which doesn't track field paths on parse errors
- Validation library (`go-playground/validator`) only validates struct tags — it doesn't help with JSON parsing errors (malformed input, unknown fields)

### Possible Improvements (future)
1. Use `json.Decoder` with custom `UnmarshalJSON` methods per request type to provide field-level error messages
2. Use a validation framework like `go-playground/validator/v10` + manual parsing for richer errors
3. Accept the current behavior — most clients send well-formed JSON, parse errors indicate client bugs

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

## DB Constraint Violations Detected via Fragile String Matching

**Discovered:** 2026-07-10 (during auth-service error handling exploration)  
**Affected services:** auth-service (handlers), user-service (service layer)

### Problem
Both auth-service and user-service detect PostgreSQL constraint violations by matching raw error message strings:

```go
// AUTH-SERVICE — in HANDLER (most fragile ❌)
if strings.Contains(err.Error(), "duplicate key value") || 
   strings.Contains(err.Error(), "23505") {
    h.errorResponse(c, http.StatusConflict, ...)
}

// USER-SERVICE — in SERVICE LAYER (still fragile ⚠️)
if strings.Contains(err.Error(), "duplicate key") ||
   strings.Contains(err.Error(), "unique constraint") ||
   strings.Contains(err.Error(), "already exists") {
    return nil, models.NewConflictError(...)
}
```

This is fragile because:
1. **Hardcodes PostgreSQL error messages/codes** — breaks if DB driver or error format changes
2. **Auth-service handles it in the handler layer** — not just service layer, which means the handler knows about PostgreSQL internals (error code `23505`)
3. **Inconsistent placement** — auth-service detects in handlers via inline `strings.Contains()`, user-service detects in service layer but still uses string matching
4. **No structured error type** — auth-service's handler returns a hardcoded status message instead of wrapping with an error sentinel that the dispatcher can match

### Current State by Service
| Service | Detection Location | Pattern | Status |
|---------|-------------------|---------|--------|
| objects-service | N/A (uses domain sentinels) | `ErrRelationshipTypeInUse` etc. | ✅ No string matching |
| user-service | Service layer (3 places) | `strings.Contains("duplicate key")` → wraps with typed `ConflictError` | ⚠️ Fragile but consistent error flow |
| auth-service | Handler layer (2 places) | `strings.Contains(err.Error(), "23505")` → inline JSON response | ❌ Most fragile — handler knows DB internals |

### Root Cause
No standard sentinel type exists for "resource in use" / constraint violation conditions. Each service either:
- Defines domain-specific sentinels (objects-service) ✅
- Wraps with generic typed structs (user-service) ⚠️  
- Falls back to inline handler detection (auth-service) ❌

### Impact on Auth-Service Error Unification
This is the primary blocker for making auth-service error handling more unified. The handler currently has two places where it:
1. Calls `h.authService.CreateRole()` or `CreatePermission()`
2. Checks if err contains "duplicate key" or "23505"
3. Returns HTTP 409 directly (bypassing HandleAuthError dispatcher)

To unify, these should move to the service layer and return a wrapped sentinel that HandleAuthError can match.

### Fix Strategy (future delta)
1. Define domain-specific sentinels in auth-service's services package:
   - `ErrRoleNameConflict` — role with this name already exists (unique constraint on roles.name)
   - `ErrPermissionNameConflict` — permission with this name/resource/action already exists
2. Move duplicate detection from handlers to service layer CreateRole/CreatePermission methods
3. Wrap detected violations with `%w ErrAlreadyExists` or domain-specific sentinels
4. Add corresponding cases to HandleAuthError dispatcher
5. Remove all `strings.Contains("duplicate key")` and `strings.Contains("23505")` from handlers

---

## Test Coverage Gaps — Systemic Blind Spots

**Discovered:** 2026-07-11 (during `fix-pagination-count-bug` investigation)  
**Status:** Open — will split into multiple future deltas

### Problem
The existing test suite has **zero coverage for SQL string correctness and pagination metadata**. A bug where `QueryBuilder.BuildCount()` discarded all WHERE filters went undetected because:

1. No unit test ever asserted what SQL `BuildCount()` produces (only verified placeholder distinctness)
2. Repository tests mocked DBPool entirely — no SQL strings captured or validated
3. Handler tests only checked HTTP status codes, never parsed response body structure
4. Service tests stubbed repository methods with hardcoded return values
5. The `total` field from `List()` was consistently discarded via `_`

Every layer mocks the one below it. No test exercises a real SQL query against a real database. This creates a **false sense of coverage** — 241+ unit tests pass, but structural bugs in SQL generation and response shape are invisible.

### The Test Chain — Where Bugs Fall Through

```
┌──────────┬─────────────────────┬─────────────────────┬──────────────────────┐
│  Layer   │ What IS tested      │ What is NOT tested  │ Why bugs escape      │
├──────────┼─────────────────────┼─────────────────────┼──────────────────────┤
│ Models   │ Struct fields, tag  │ N/A                 │ Irrelevant           │
│          │ logic               │                     │                      │
├──────────┼─────────────────────┼─────────────────────┼──────────────────────┤
│ Query    │ Placeholder         │ SQL string content; │ Tests verify $1,$2   │
│ Builder  │ distinctness        │ BuildCount() output │ are ≠, not what the  │
│          │                     │                     │ final SQL looks like │
├──────────┼─────────────────────┼─────────────────────┼──────────────────────┤
│ Re       │ Mock rows return    │ Actual SQL executed;│ DBPool is mocked; no │
│ pository │ non-nil slices      │ total value from    │ query ever runs, so  │
│          │                     │ List() return value │ output shape never   │
│          │                     │ (discarded via _)   │ validated            │
├──────────┼─────────────────────┼─────────────────────┼──────────────────────┤
│ Service  │ Validation logic,   │ End-to-end data     │ Repo method stubbed; │
│          │ business rules      │ flow                │ pagination total     │
│          │                     │                     │ never flows through  │
├──────────┼─────────────────────┼─────────────────────┼──────────────────────┤
│ Handler  │ HTTP status codes   │ Response body JSON; │ Service is mocked;   │
│          │                     │ pagination structure│ handler returns      │
│          │                     │ invariants          │ whatever mock said   │
└──────────┴─────────────────────┴─────────────────────┴──────────────────────┘
```

### Specific Gaps (objects-service)

| Gap | Location | Current Behavior | What Should Be Tested |
|-----|----------|-----------------|----------------------|
| **G1: BuildCount SQL content** | `repository_test.go` | Placeholder distinctness checked only | `BuildCount()` SQL contains WHERE clauses for every filter type (Where, WhereTagsContain, WhereJsonContains, WhereDateRange) |
| **G2: List() total value captured** | `repository_test.go` | `result, _, err := repo.List(...)` — total discarded | Assert `total >= int64(len(result))`, test with known filter counts |
| **G3: Handler response body structure** | `object_handler_test.go` | Only `assert.Equal(t, 200, w.Code)` | Parse JSON, verify pagination fields exist and are consistent (`count == len(data)`, `total >= count`) |
| **G4: Cross-layer filter flow** | (anywhere) | Each layer mocked in isolation | When a filter is applied at handler → service → repo → SQL contains WHERE → total reflects filtered set |
| **G5: RBAC does not affect List() pagination total** | `object_handler.go:389` + `permiddleware/permission.go` | Permission middleware gates ACCESS (403 vs allow) but does NOT filter DATA. Handler comment: "No self-filtering by user ID for List()." A user with only `objects:read:own` sees ALL objects in List(), not just their own. Total = 56 regardless of ownership. | Pagination total should reflect RBAC scope: if user has only `:read:own`, total = count of objects where `created_by = current_user_id`. Currently the permission system is a gatekeeper, not a data filter.

### Concrete Fix Patterns Needed

**Pattern A — Capture SQL strings in mocks:**
```go
capturedSQL := ""
mockDB.QueryFunc = func(ctx, sql string, args...) { capturedSQL = sql; return ... }
repo.List(...)
assert.Contains(t, capturedSQL, "WHERE object_type_id")  // catches filter loss
```

**Pattern B — Assert response body structure:**
```go
var resp map[string]interface{}
json.Unmarshal(w.Body.Bytes(), &resp)
pagination := resp["pagination"].(map[string]interface{})
assert.Equal(t, float64(len(data)), pagination["count"])
```

**Pattern C — SQL content assertions on QueryBuilder:**
```go
sql, _ := qb.BuildCount()
assert.Contains(t, sql, "WHERE object_type_id = $1")
assert.NotContains(t, sql, "ORDER BY")  // if stripped
```

### Scope & Splitting Plan (future deltas)

This is too large for a single delta. Expected split:

| Delta | Scope | Effort |
|-------|-------|--------|
| **delta A** — QueryBuilder SQL assertions | Add content-assertion tests to existing `repository_test.go` BuildCount cases; add SQL capture pattern to a subset of repository tests | ~30 min |
| **delta B** — Handler response body checks | Add JSON parsing + pagination invariant assertions to handler List tests (object, object_type, relationship) | ~1 hour |
| **delta C** — Repository total value validation | Stop discarding `total` from List() return; add invariant tests (`total >= count`, `count <= limit`) | ~30 min |
| **delta D** — Cross-layer filter flow test | Contract-style test: handler filter → service call → repo SQL contains WHERE → correct total. Could use a thin real-DB or sophisticated mock capture layer | ~2 hours |

### RBAC Permissions Gate Access But Never Scope Data

**Discovered:** 2026-07-11 (during pagination bug investigation)  
**Status:** Architectural gap — RBAC is a gatekeeper, not a data filter

The `fix-pagination-count-bug` delta corrected BuildCount() to preserve WHERE filters. But there's a deeper structural issue: **RBAC permissions are never translated into data-scoping filters on List queries.** The permission system gates which endpoints you can call — once inside, all matching rows return regardless of the user's RBAC scope.

```
┌─────────────────────────────────────────────────────┐
│  Current RBAC → Pagination Flow                     │
├─────────────────────────────────────────────────────┤
│                                                     │
│  permiddleware (gateway):                           │
│    checks: objects:read:all OR objects:read:own     │
│    result: ALLOW or 403                             │
│         ↓                                             │
│      handler.List()                                  │
│         ↓  ← NO RBAC-SCOPED FILTERING               │
│         ↓  Comment (object_handler.go:389):          │
│         ↓  "No self-filtering by user ID for List" │
│         ↓                                             │
│      service.List(filter)                            │
│         ↓                                             │
│      repo.List(sql → SELECT * FROM objects...)       │
│         ↓                                             │
│      total = ALL rows (56), regardless of scope    │
│                                                     │
└─────────────────────────────────────────────────────┘
```

**This is not just about `:own`.** RBAC scopes data access in multiple ways that never reach the query layer:

| Scope Type | Example | How it gates | How it should filter |
|------------|---------|-------------|---------------------|
| **Type-level** | User has no permission for `object-types` type | 403 at middleware, never reaches handler | N/A — already blocked |
| **Method-level** | User can `GET /objects` but not `POST /objects` | Middleware allows GET only | N/A — method routing handles this |
| **Ownership (`:own`)** | User has `objects:read:own` only | Allows request, no ownership filter applied | `WHERE created_by = user_id`, total scoped to user's objects |
| **Cross-type scope** | User can read `objects` but not `relationships` | Middleware blocks relationship routes | N/A — already blocked at route level |

The gap is specifically in **ownership-scoped access (`:own`)** where the request IS allowed through but data returns unfiltered. This affects pagination metadata:
- A user with only `objects:read:own` sees all 56 objects in List()
- Their actual scoped total should be N (their own objects)
- The client receives `total=56, count=50` which is misleading

**Current state by operation:**
| Operation | RBAC enforcement | Data filtered? |
|-----------|-----------------|----------------|
| GetByID/GetByPublicID | ✓ Ownership check via `checkOwnership()` (per-object) | N/A — single object return |
| Update/Delete | ✓ Ownership check via `checkOwnership()` (per-operation) | N/A — operates on one object |
| **List** | ✗ Only gates endpoint access | **NO — returns all matching rows regardless of scope** |

This is likely intentional design (the comment notes ownership filtering can be opt-in via explicit query params), but it creates a **misleading pagination total** for scoped users. The fix-pagination-count-bug delta solved the WHERE-preservation bug, but RBAC-scoped totals remain unaddressed.

**Design decision needed (future delta):**
- Option A: Add implicit scope filter to List() based on `matched_permissions` — if user has only `:own`, add `WHERE created_by = user_id`; total reflects scoped count
- Option B: Document that pagination total represents global query match and ownership filtering must be explicit via query params (`?created_by=me`)
- Option C: Return dual metadata — scoped_total (what user can see) + global_total (admin view)

---

### Principle to Adopt Going Forward

> **Never assert just status codes.** For any endpoint returning structured data, always parse and validate response body structure. Status code is necessary but not sufficient.

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
