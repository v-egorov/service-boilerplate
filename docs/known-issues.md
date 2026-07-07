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

## Pagination Inconsistency Across Objects-Service Endpoints

**Discovered:** 2026-07-07 (during `fix-objects-service-null-data` container verification)  
**Affected service:** objects-service

### Problem
Different endpoints use different pagination parameter names and response field conventions:

| Endpoint | Param Format | Response Field |
|----------|-------------|----------------|
| `/api/v1/objects` (List, Search) | `?limit=N&offset=M` | `"count"` + `"total"` |
| `/api/v1/object-types` (List) | `?limit=N&offset=M` | `"count"` + `"total"` |
| `/api/v1/relationships` (List) | `?page=P&page_size=S` | `"limit"` + `"total"` |

### Example
```bash
# Objects-service uses limit/offset
curl 'http://localhost:8085/api/v1/objects?limit=3&offset=0'   # ✅ works

# Relationships uses page/page_size — limit/offset ignored!
curl 'http://localhost:8085/api/v1/relationships?limit=3&offset=9999'  # ❌ returns all (defaults to page=1, page_size=20)

curl 'http://localhost:8085/api/v1/relationships?page=1&page_size=3'   # ✅ works
```

### Root Cause
- `ObjectFilter` struct uses `Limit int` / `Offset int` with form tags matching query params
- `RelationshipFilter` struct uses `Page int` / `PageSize int` — completely different field names
- No shared pagination interface or helper to normalize across endpoints

### Impact
Clients must remember which endpoint uses which param format. Breaking changes when calling different list endpoints from the same tool.

### Fix Strategy (future delta)
Standardize on one convention across all endpoints:
1. Choose `limit`/`offset` OR `page`/`page_size` as canonical
2. Update filter structs to use consistent field names
3. Add a shared pagination response struct with consistent fields (`count`, `total`, `has_more`)

**Discovered:** 2026-07-04 (during `fix-gateway-sse-and-perm-jwt` delta investigation)  
**Resolved by:** revert commit `5aab940`

### Problem (resolved)
During debugging of a persistent **400 Bad Request** bug on StreamableHTTP POST requests, there was a suspicion that Air's volume mount + inotify watching caused Docker session state loss between rebuilds. This led to considering switching all `.air.toml` files from `poll = false` to `poll = true`.

### Root Cause (debunked)
The 400 bug was **NOT** an Air/watcher issue. It was a legitimate request-handling bug in the mcp-go HTTP handler chain — specifically how `handlePost` reads the body under certain client conditions. The issue persisted even when running the binary directly (no Air involved), proving it was not related to rebuild timing or session state.

### Current State
- **All 5 services** use `poll = false` (inotify mode) — confirmed working correctly
- File changes are detected and hot-reloaded without issues across all services
- Volume mounts work as expected: host file edits → Air detects → container rebuilds → new binary runs
- **Triggering a rebuild**: Use `touch -m <file.go>` (not plain `touch`). The `-m` flag performs an actual metadata write syscall that Docker volume mounts translate into inotify IN_MODIFY events. Plain `touch` only updates stat info and does NOT trigger Air.

### Lesson for Future Debugging
When encountering bugs in containers with Air + volume mounts:
1. **First** try running the compiled binary directly (`./tmp/service-name`) to rule out Air/watcher issues
2. If `curl`/raw TCP still reproduces the bug → it's a code-level issue, NOT an Air issue
3. **Never** switch `.air.toml` files to `poll = true` without first verifying the binary directly
4. Inotify mode works fine — this is the correct configuration for all services

### Affected Files (all unchanged)
- `services/mcp-server/.air.toml`
- `services/objects-service/.air.toml`
- `services/auth-service/.air.toml`
- `services/user-service/.air.toml`
- `api-gateway/.air.toml`

---

## MCP Server Identity Forwarding

**Discovered:** 2026-07-04  
**Resolved by:** `fix-mcp-auth-chain` delta

### Problem (resolved)
Gateway injects identity headers (`X-User-ID`, `X-User-Email`, `X-User-Roles`) for all `/mcp/*` requests. mcp-server receives them on the POST /message request, but when tool handlers call objects-service via `ObjectsClient.httpGet()`, they created a **brand new HTTP request with zero headers**. Identity was lost at hop 2 — every object-tool call returns 401 from objects-service's permiddleware.

### Current State (post-fix)
- Gateway → mcp-server: ✅ headers present and correct (verified in logs)
- mcp-server → objects-service: ✅ identity headers forwarded via context threading
- Objects-service permiddleware: sees valid user context → authorizes read operations

### Root Cause (resolved)
`services/mcp-server/internal/client/objects_client.go` — `httpGet()` created a new request without copying any headers. **mcp-go DOES carry HTTP headers** on each handler's `request.Header` (via SSE transport in v0.55.1). The gap was that:
1. Tool/resource/prompt handlers never extracted identity from `request.Header`
2. `ObjectsClient.httpGet()` had no mechanism to receive or forward identity
3. No context-based threading existed between handler entry point and outbound HTTP call

### Fix Applied
- Added `identity.go` with `WithIdentity(ctx, hdr)` / `IdentityFromContext(ctx)` helpers
- Modified 7 data methods on `ObjectsClient` to accept `context.Context`
- `httpGet()` reads identity from context and forwards only the allow-listed headers (`X-User-ID`, `X-User-Email`, `X-User-Roles`)
- Each handler injects identity at entry: `ctx = client.WithIdentity(ctx, request.Header)`
- Regression test in `objects_client_identity_test.go` covers all scenarios

See [`design.md`](../openspec/changes/fix-mcp-auth-chain/design.md) for full architectural rationale.

---

## Migration 00012 — User-Role Linkage Not Applied

**Discovered:** 2026-07-04  
**Delta:** `mcp-server-initial` (archived)

### Problem
Migration file `services/auth-service/migrations/development/000012_mcp_agent_permissions.up.sql` exists and version 12 is recorded in `auth_service.schema_migrations`. However, the linking query produced zero rows — no entry in `auth_service.user_roles` connecting the mcp-agent user to the role.

### Current State
- ✅ User exists: `user_service.users` has `000...1 | mcp-agent@system.internal` (version 7 applied)
- ✅ Role exists: `auth_service.roles` has `mcp-agent-read-only` (created by migration 000011)
- ❌ No row in `auth_service.user_roles` linking them
- ⚠️ Manually linked outside of migrations for now

### Suspected Cause
The migration uses `ON CONFLICT DO NOTHING` with a JOIN:
```sql
INSERT INTO auth_service.user_roles (user_id, role_id)
SELECT u.id, r.id FROM user_service.users u
CROSS JOIN auth_service.roles r
WHERE u.email = 'mcp-agent@system.internal' AND r.name = 'mcp-agent-read-only'
ON CONFLICT DO NOTHING;
```

Likely failure modes:
1. Role was created by an earlier migration (000001–000010) with a **different UUID** than what the JOIN resolves to, causing the link query to match the wrong row or fail silently due to ON CONFLICT
2. Migration 000011 (`ON CONFLICT DO NOTHING`) created the role, but its UUID doesn't match any previous reference — the join should still work unless there's a timing/ordering issue with how golang-migrate applies migrations vs. how rows are referenced

### Investigation Needed
- Compare actual role UUID in DB against migration 000011 content
- Check if the JOIN query returns results when run manually against current DB state
- Verify migration execution order and whether any earlier migration touched `roles` table

---

## MCP Spec Gaps — Deferred to Future Delta

**Discovered:** 2026-07-04 (initially), **re-audited:** 2026-07-04
**Delta:** `mcp-server-initial` (archived)

### Problem
During a full spec-vs-reality audit, two requirements from the mcp-server baseline spec are defined but not implemented. The MCP server has 11 registered capabilities in total — 9 work correctly, 2 have no implementation.

#### Gap 1: `list_object_types` with `parent_type_id` filter (spec line ~20)
**Spec says:**
> The tool SHALL accept optional filter parameters (`type_key_prefix`, **`parent_type_id`**) to narrow results.
> Scenario: listing by parent_type_id — "querying Product's children returns Electronics, Clothing, Books"

**Reality:** `ListObjectTypesParams` struct in `type_tools.go` has only `TypeKeyPrefix`. No `ParentTypeID` field. Objects-service repository supports filtering by `ObjectTypeID`, but there's no MCP tool param or client method to request "children of type X".

**Impact:** Agents cannot ask "what are the children of Product?" through the MCP protocol. They must already know child IDs/names and call `get_object_type` individually.

#### Gap 2: `list_objects` with `type_key_prefix` cross-type filtering (spec line ~195)
**Spec says:**
> The mcp-server MUST support filtering objects by **`type_key_prefix`** in addition to `object_type_id`. When a user wants all objects across a type namespace, the tool SHALL resolve child type_keys matching the prefix and query them.
> Scenario: "list_objects with type_key_prefix=product" → resolves product, product-electronics, product-clothing

**Reality:** `ListObjectsParams` struct has `object_type_id`, `page`, `page_size`. No `type_key_prefix` field. The objects-service `List()` repository method does NOT support filtering by `type_key` at all (it only filters by `ObjectTypeID`). No client method exists to resolve a prefix into multiple type IDs.

**Impact:** Agents cannot query "show me all products and their variants" in one call. They must know each variant's object_type_id separately and issue multiple list_objects calls, then merge results client-side.

#### What IS working (verified against spec)
| Spec Requirement | Status |
|-----------------|--------|
| `list_object_types` with `type_key_prefix` filter | ✅ Implemented (`TypeKeyPrefix` param → objects-service) |
| `get_object_type` by ID or name/type_key | ✅ Both paths: `GetObjectTypeByID`, `GetObjectTypeByName` |
| Resource `objects-types://hierarchy` | ✅ Registered, returns full type tree with children |
| Prompt `browse_schema` (with optional prefix filter) | ✅ Registered, works via SSE |
| Prompt `get_object_info` (requires object_type_id) | ✅ Registered, works via SSE |
| SSE transport via Gateway reverse proxy | ✅ Verified in e2e test |
| Identity forwarding (3 headers only) | ✅ Via `identity.go` context threading |
| Gateway-trust skip for MCP reads | ✅ permiddleware skip path active |
| NOT NULL type_key guarantee on responses | ✅ DB constraint enforced |

#### Deferred Actions
- **Gap 1:** Add `ParentTypeID *int64` to `ListObjectTypesParams`, implement `ObjectTypeRepository.GetChildren(ctx, parentID)` method (or use existing `GetDescendants` with depth=1), wire through MCP tool.
- **Gap 2:** Add `TypeKeyPrefix string` to `ListObjectsParams`, create client method that calls objects-service `/api/v1/object-types?type_key_prefix=X` → resolves matching type_keys → batches `list_objects` calls per ID → merges results. Or better: add a new endpoint on objects-service for prefix-based multi-type queries.

---

## gin.ResponseWriter Interface Not Mockable

**Discovered:** During api-gateway unit test development  
**Delta:** `unit-tests-api-gateway-*` changes (archived)

### Problem
The `gin.ResponseWriter` interface requires 9 methods to implement. Creating a mock for it is impractical — any unimplemented method causes runtime panics in gin's internal code paths.

### Current State
- Established pattern: smoke testing through actual middleware execution rather than unit mocks
- Tests verify behavior by running real middleware chains and checking response status/body, not by mocking the writer
- This is a known limitation of the test strategy, not a bug — it affects how much coverage we can achieve for handlers like `ProxyRequest` (which uses `httputil.ReverseProxy`)

---

## E2E Test Script — Unverified Response Parsing

**Discovered:** 2026-07-04 (same commit as script creation)

### Problem
`scripts/test-mcp-e2e.sh` was written during debugging but never run end-to-end with a clean pass. The SSE response parsing logic may still have issues — earlier attempts failed with jq errors on malformed input, and the script's own output showed empty results for steps 4–6 despite raw data being present in the SSE stream.

### Current State
- Script exists and is executable
- Initial SSE connection (Step 1–2) likely works
- Subsequent tool call responses (Steps 3–6): **unconfirmed** — may fail silently due to JSON parsing errors, missing SSE response lines, or timing issues with the `sleep` delays between POST requests and SSE stream arrivals
- No clean validation run has been performed

### Action Needed
Run `bash scripts/test-mcp-e2e.sh` on a fresh environment (or at least after restarting all containers) to verify:
1. All 7 steps pass cleanly with colored output
2. Tool call results are correctly extracted from SSE stream and parsed by jq
3. Auth-chain check in Step 7 reflects actual DB state

---

## Gateway SSE Panic — httputil.ReverseProxy + gin Recovery Conflict

**Discovered:** 2026-07-04
**Planned for:** `fix-gateway-sse-and-perm-jwt` delta

### Problem
`ProxyMCPRequest()` uses `httputil.ReverseProxy` to stream SSE to clients. When the client disconnects or the SSE stream ends, Go's reverse proxy calls `panic(http.ErrAbortHandler)` as **control flow** (not a real error). Gin's `gin.Recovery()` middleware catches this panic and treats it as a server error, returning HTTP 500 and logging:

```
[Recovery] 2026/07/04 - 10:18:17 panic recovered:
net/http: abort Handler
```

This breaks all SSE connections — the gateway can't maintain persistent event streams.

### Fix (planned)
Wrap `proxy.ServeHTTP(c.Writer, c.Request)` with a deferred recover that catches `http.ErrAbortHandler` specifically and silently returns, re-panicking everything else. ~5 lines in `gateway.go`.

---

## MCP Auth Chain — Permission Check Gap (Objects → Auth Hop)

**Discovered:** 2026-07-04
**Planned for:** `fix-gateway-sse-and-perm-jwt` delta

### Problem
With the identity forwarding fix (`fix-mcp-auth-chain`), gateway-injected `X-User-*` headers now flow correctly:

```
Gateway → mcp-server → objects-service
   ✅         ✅            ✅ identity headers arrive
```

Authentication works — objects-service's JWTMiddleware reads `X-User-ID` from headers and sets `user_id` in gin context (dev mode: `jwtSecret == nil`).

But authorization breaks at the next hop:

```
objects-service ──CheckPermission──────────▶ auth-service
  (no JWT token)           RequireAuth() → user_id empty → 401
```

`permiddleware.NewPermissionMiddleware()` calls `authClient.CheckPermission(userID, permission, jwtToken)` where `jwtToken` is empty (no Authorization header on the original MCP request). The auth client doesn't forward `X-User-*` headers. Auth-service's `RequireAuth()` middleware rejects because `GetAuthenticatedUserID()` returns empty.

In production this path is unreachable — all services are isolated behind the gateway, everything has a JWT. But in dev mode with MCP (no JWT validation by design), the round-trip to auth-service is the gap.

### Chosen Approach: Option B — Skip auth-service for gateway-trusted GET requests

Architectural options evaluated:

| Option | Approach | Verdict |
|--------|----------|---------|
| A | Forward `X-User-ID` to auth-service so its JWT middleware reads it | Papering over the gap, implicit trust |
| **B** | **Skip `CheckPermission()` when `jwtToken == "" && userID != ""` for GET requests** | **Chosen — clean, dev-mode only** |
| C | Inject a shared internal token header across all services | Overengineered for dev mode |

**Rationale for Option B:**
- In production, all requests have a JWT (gateway validates), so the `Authorization`-absent branch never triggers
- In dev mode, the gateway IS the authenticator — skipping the auth-service round-trip is consistent with the existing gateway-trust model
- Change is scoped to one file (`permiddleware.go`), ~3 lines

**Known compromise:** This skips RBAC for MCP read operations in dev mode — the `mcp-agent` system user effectively has unrestricted read access. This is intentional: per-user MCP identity would require API keys (not login/password), which is a separate architectural scope. The current goal is "make the damn thing work" for a single developer interacting with their own system.

### Future Consideration
When API-key-based MCP client identity is implemented, the proper fix is to give mcp-server a valid JWT (gateway logs in as the MCP client, injects the token). Then the entire RBAC chain — `user_roles → role_permissions → permissions` — works natively without any shortcuts.

### Delta Implementation Status: Paused

**Date:** 2026-07-04  
**Delta:** `fix-gateway-sse-and-perm-jwt`

#### What worked (confirmed)
1. **SSE panic recovery**: Gateway no longer crashes on SSE disconnect — E2E Step 2 passes, no `[Recovery]` panic logs
2. **Permiddleware skip fires correctly**: When `jwtToken == "" && userID != "" && GET`, the auth-service round-trip is skipped and permiddleware sets `matched_permissions` → request reaches service layer

#### What blocked E2E (pre-existing bug)
After permiddleware passes through, objects-service returns **500 Internal Server Error** with an empty error (`{}`). This happens for ALL users — not just MCP:

| User | Auth method | Result |
|------|------------|--------|
| MCP agent (no JWT) | Skip path → service layer | 500 `{}` |
| Dev admin (real JWT) | Normal RBAC → auth-service returns false | 403 "Insufficient permissions" |

**Root cause: permission model mismatch between permiddleware and DB.**

- Permiddleware builds permission strings for GET as `{type_key}:read:all` and `{type_key}:read:own`
- Auth-service migrations store permissions with `action = "read"` (not `":read:all"` or `":read:own"`)
- Result: `CheckPermission()` returns false for both → empty matchedPermissions → 500 from handler

**This is a pre-existing design bug in objects-service — not introduced by this delta.** It affects the entire service, not just MCP. Requires a separate fix.

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
- **objects-service**: 25 occurrences across 3 repository files (`object_repository.go`, `object_type_repository.go`, `relationship_repository.go`) — all fixed by this delta
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
users := make([]*models.User, 0)
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
