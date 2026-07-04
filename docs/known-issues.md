# Known Issues & Deferred Items

Items tracked for future investigation or deferred to a later delta. Each entry has discovery date, affected change (if applicable), and the current state of understanding.

---

## MCP Server Identity Forwarding

**Discovered:** 2026-07-04  
**Delta:** `mcp-server-initial` (archived)

### Problem
Gateway injects identity headers (`X-User-ID`, `X-User-Email`, `X-User-Roles`) for all `/mcp/*` requests. mcp-server receives them on the POST /message request, but when tool handlers call objects-service via `ObjectsClient.httpGet()`, they create a **brand new HTTP request with zero headers**. Identity is lost at hop 2 — every object-tool call returns 401 from objects-service's permiddleware.

### Current State
- Gateway → mcp-server: ✅ headers present and correct (verified in logs)
- mcp-server → objects-service: ❌ no identity headers forwarded
- Objects-service permiddleware: sees empty user context → rejects with 401

### Root Cause Location
`services/mcp-server/internal/client/objects_client.go` — `httpGet()` creates a new request without copying any headers from the original request. mcp-go's session context does not carry HTTP headers.

### Possible Approaches (for next delta)
1. **Context threading**: Extract identity from gateway request, store in mcp-go session/context during message handling, retrieve inside tool handlers before calling objects-service
2. **Explicit header forwarding**: Pass X-User-* as MCP tool call arguments and have the client layer attach them
3. **Service account config**: Have mcp-server read its own config for identity (simpler, but loses per-client context)

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

**Discovered:** 2026-07-04  
**Delta:** `mcp-server-initial` (archived)

### Problem
The delta spec defines requirements that aren't fully implemented:

1. **`list_object_types` filter parameters**: Spec mentions optional `type_key_prefix` and `parent_type_id` filters, but current implementation doesn't pass these through to objects-service (or accepts them as arguments at all).

2. **Full hierarchy resource**: Delta 1 registers a root-level types resource only. Full recursive tree traversal is deferred.

### Current State
- `list_object_types()` calls work but return everything — no filtering support
- Resource `objects-types://hierarchy` returns root types with children (one level) only

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

*Last updated: 2026-07-04*
