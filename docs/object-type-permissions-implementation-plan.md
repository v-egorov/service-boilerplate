# Object-Type-Scoped Permissions: Implementation Plan

**Status**: Draft — pending review  
**Date**: 2026-05-14  
**Related**: `object-type-permissions-architecture.md`, `rbac-objects-service.md`

---

## Background & Motivation

The current permission system uses flat resource names (`objects:read:own`, `relationships:create`) with no type-level granularity. As new object types are introduced (portfolio, asset, etc.), this approach won't scale because:

1. All typed objects share the same `objects:*` permissions — no per-type control
2. Relationships are CTE-derived objects but have separate permission management that can contradict object-level rules
3. No way to grant granular access like "read portfolios but not assets"

This plan addresses these concerns while maintaining compatibility with existing code.

### Cache Status: Disabled (Deferred for Cross-Container Support)

The `PermissionCache` type exists in the auth-service but is **not wired at startup** — permission checks always hit the database directly. This was a deliberate choice because:

1. The current cache uses an in-memory LRU map with TTL expiration
2. In production, multiple auth-service replicas will be running simultaneously
3. `Invalidate()` / `InvalidateAll()` only clear one container's memory — there is no cross-container invalidation mechanism (e.g., Redis PubSub)

Without distributed caching infrastructure, an in-memory cache causes subtle bugs: after mutating permissions via admin API endpoints on replica A, requests routed to replicas B/C serve stale cached results until TTL expires. This makes end-to-end testing impossible without 60+ second delays between mutation calls and verification checks.

**Current behavior:** All permission lookups go straight through `auth_repository.GetUserPermissions()` — correct but with no caching benefit. When a distributed cache (Redis, etc.) is added later, the fix will be straightforward: swap in an external store, implement cross-container invalidation via pub/sub or event-driven updates, then re-enable the existing `PermissionCache` wrapper around it.

---

## Permission Resolution Model

Two rules operate at different levels:

### Rule A — Within-role: One Scoped Level Per Role/Action

Each role gets exactly one scoped variant per type+action pair. This is enforced at the permission assignment API level (POST/PATCH/PUT). No ambiguity exists in normal operation:

```
Role grants: read:own  → only `read:own` applies
Role grants: read:all  → only `read:all` applies
```

> **Note:** The `create` action uses flat permission names on objects/types (`object-types:create`, `portfolio:create`, etc.) — you always own what you create. Scoped variants for `create` only apply to relationships where endpoint ownership must be verified (`relationships:create:own`, `relationships:create:all`).

### Rule A Edge Case: Duplicate Scoped Variants in Same Role

The API enforces one scoped level per role+action pair, but if a race condition or data corruption causes both `type:action:own` and `type:action:all` to exist for the same role, resolution follows **Rule B semantics** — broadest scope wins (`:all` overrides `:own`). This is logged at warning level with role_id, permission names involved, and affected user count. The system treats this identically to cross-role union because the practical effect is the same: a broader scope exists and takes precedence.

### Rule B — Across-roles: Union then Broadest Wins

When multiple roles grant permissions for the same action, all are collected into a union set. If ANY role provides the broader `:all` variant, ownership checks are bypassed entirely:

| Available in Union | Resolution | Ownership Check? |
|--------------------|------------|------------------|
| `*:all` only       | Allow unconditionally | No |
| `*:own` only       | Conditional allow     | Yes — on endpoints |
| Both              | Same as `*:all`        | No |
| Neither           | Deny                  | N/A |

**Key insight:** Different actions are independent. Having `update:own` + `read:all` means update requires ownership verification but read does not. Actions never interfere with each other.

---

## Final Roles & Permissions Matrix

### Object Types Permissions (registry)

Permissions on the `object_types` table itself. The `create` action uses flat permission names — you always own what you create, so no scoped variants apply.

| Role | create | read | update | delete |
|------|--------|------|--------|--------|
| admin | `create` | `all` | `all` | `all` |
| object-type-admin | `create` | `all` | `all` | `all` |
| user, relationship-admin, relationship-viewer | — | `all` | — | — |

### Objects Permissions (data instances)

Permissions on actual object records. The `create` action uses flat permission names since you always own what you create.

| Role | create | read | update | delete |
|------|--------|------|--------|--------|
| admin | `create` | `all` | `all` | `all` |
| object-type-admin | — | `all` | — | — |
| user | `create` | `own` | `own` | `own` |

### Relationships Permissions (NEW scoped model)

The `create` action on relationships uses scoped variants because creating a relationship requires ownership of pre-existing endpoint objects (source AND target). This differs from plain object creation, where you always own what you create.

| Role | create | read | update | delete | Notes |
|------|--------|------|--------|--------|-------|
| **admin** | `all` | `all` | `all` | `all` | Full CRUD, no restrictions |
| **relationship-admin** (NEW) | `all` | `all` | `all` | `all` | Dedicated relationship lifecycle management |
| **relationship-viewer** (NEW) | — | `all` | — | — | Read-only, audit/discovery |
| **user** | `own` | `own` | `own` | `own` | Can act on relationships involving owned endpoints; if also assigned `relationship-viewer`, union grants `read:all` |
| **object-type-admin** | — | — | — | — | Not applicable to relationships |

### Three Distinct Ownership Models for Relationships

Unlike plain objects where "owned = created by this user" applies uniformly, relationship permissions use **three different ownership models** depending on the action type:

| Action | Ownership Model | Logic | What Is Checked |
|--------|----------------|-------|-----------------|
| `create:own` | Endpoint ownership | AND — must own BOTH endpoints | `owner_id` on source_object AND target_object match user ID |
| `read:own`   | Partial endpoint ownership | OR — own at least ONE endpoint | `owner_id` on source_object **OR** target_object matches user ID |
| `update:own` / `delete:own` | Creator ownership | N/A — checks the record itself | `created_by` field on the relationship object matches user ID |

**Rationale:** Create requires owning both endpoints because you can only link objects under your control. Read uses OR logic to enable discovery — if I own either side of a relationship, I should be able to see it for context. Update/delete checks `created_by` (the relationship record owner) rather than endpoint ownership, since modifying or removing a relationship is an action on the link itself, not on its endpoints.

### Ownership Check Logic

```
For :own permissions, middleware verifies ownership BEFORE allowing action.
For :all permissions (or if any role grants :all), no ownership check is performed.
Ownership verification checks use three distinct models:

  - create:own → Endpoint ownership model (AND)
    owner_id of source_object AND target_object match user ID (relationships only)

  - read:own   → Partial endpoint ownership model (OR)
    owner_id of source_object OR target_object matches user ID

  - update/delete:own → Creator ownership model (N/A — record-level check)
    created_by field on relationship object matches user ID
```

---

## Implementation Plan (Todo-Style)

### Phase 0: Documentation & Audit

> **Goal:** Update docs first, then audit middleware to understand current enforcement. No code changes yet.

#### Step 0.1 — Update `docs/object-type-permissions-architecture.md` [ ]
- [x] Rewrite Section 3 (precedence) with two-rule model clarification
- [x] Replace unscoped permissions table in Section 4 with scoped-only variants
- [x] Add new roles section: relationship-admin, relationship-viewer + assignment matrix
- [x] Fix transition section — remove contradictory `objects:read:own` statement (line 201)
- [ ] Update role list and permission assignments throughout document

#### Step 0.2 — Audit & enforce permission assignment constraints [ ]
**Files to check:**
- `services/auth-service/internal/handlers/` — look for handlers that manage role_permissions (POST/PATCH/PUT)
- Any admin API endpoints for role management, permission grants

**What to verify & implement:**
1. Identify which API endpoints assign permissions to roles
2. Add validation: when granting a scoped variant to a role, check if an overlapping scoped variant already exists on the same type+action → warn or error
   - Example: if role "X" has `portfolio:read:own` and assignment tries to add `portfolio:read:all` → block with clear message about redundancy
3. Implement enforcement constraint: PREVENT assignments where both scoped variants for the same type+action exist on the same role
   - Rule: a single role should never have both `type:action:own` AND `type:action:all` simultaneously
   - This is enforced at the permission assignment API level (POST/PATCH/PUT to auth_service.role_permissions)

**Output:** Validation logic + updated permission assignment endpoints with conflict detection and blocking.

#### Step 0.3 — Audit current permission middleware for enforcement [ ]
**Files to check:**
- `services/objects-service/internal/middleware/*`
- `common/middleware/permission*.go` or shared auth middleware
- Router setup in `cmd/main.go` (how permissions are wired to endpoints)

**What to verify:**
1. Does middleware resolve scoped variants (`:own`, `:all`) or just check flat permission names?
2. If scoped — is ownership actually enforced for `:own` variants? (Test with current test data)
3. How does multi-role lookup work currently? (Does it union across roles?)
4. Does middleware look up `object_type_id` during checks, or rely on endpoint path alone?

**Output:** A summary of what the middleware DOES vs. what it SHOULD do per this plan.

#### Step 0.4 — Update `docs/rbac-objects-service.md` [ ]
- [ ] Add relationship roles (relationship-admin, relationship-viewer) to role definitions table
- [ ] Note that object-type-admin has NO relationship permissions
- [ ] Clarify multi-role union rule in security model section

---

### Phase 1: New Roles & Scoped Permissions (migrations only)

> **Goal:** Extend permission definitions and add new roles. No code changes — just SQL migrations.

#### Step 1.1 — Expand migration `000008_add_relationships_permissions.up.sql` [ ]
**Current state:** Only defines flat permissions (`create`, `read`, `update`, `delete`)  
**Target state:** Define scoped variants for all CRUD actions

```sql
-- Add relationships scoped permissions (R2.14 — extended)
INSERT INTO auth_service.permissions (name, resource, action) VALUES
    ('relationships:create:own',     'relationships', 'create'),
    ('relationships:create:all',     'relationships', 'create'),
    -- read variants already exist as flat names; replace with scoped:
    ('relationships:read:own',       'relationships', 'read'),
    ('relationships:read:all',       'relationships', 'read'),
    -- update/delete scoped variants
    ('relationships:update:own',     'relationships', 'update'),
    ('relationships:update:all',     'relationships', 'update'),
    ('relationships:delete:own',     'relationships', 'delete'),
    ('relationships:delete:all',     'relationships', 'delete')
ON CONFLICT (name) DO NOTHING;

-- Keep flat names for backward compatibility during transition period (deprecated)
```

#### Step 1.2 — Add new roles via migration `000010_add_relationship_roles.up.sql` [ ]
**New file needed:**
```sql
INSERT INTO auth_service.roles (name, description) VALUES
    ('relationship-admin', 'Dedicated role for full relationship instance management'),
    ('relationship-viewer', 'Read-only access to relationship instances for audit and discovery')
ON CONFLICT (name) DO NOTHING;
```

#### Step 1.3 — Update migration `000009_assign_relationships_permissions.up.sql` [ ]
**Current state:** Assigns ALL permissions to admin + object-type-admin  
**Target state:** Per-role assignment matrix per the table above

Changes needed:
- Remove `object-type-admin` from relationships assignments entirely
- Add `relationship-admin` → all scoped variants (`all` for read/update/delete; `all` for create since scope on create applies only to endpoint ownership)
- Update `user` → flat `create` (objects/types), `own` for all CRUD on relationships, `own` for read/update/delete on objects
- Add `relationship-viewer` → ONLY `read:all`

#### Step 1.4 — Rollback and re-apply [ ]
```bash
make db-migrate-down SERVICE_NAME=auth-service   # rolls back 000009, then 000008
# (or manually run .down.sql for each)
# Verify clean state
Makefile: make db-migrate-up SERVICE_NAME=auth-service   # re-apply all in sequence
```

---

### Phase 2: Middleware Enhancement (code changes)

> **Goal:** Update permission middleware to enforce scoped variants and implement multi-role union logic.

#### Step 2.1 — Scoped variant enforcement [ ]
**Location:** Permission checking function (to be identified during audit in Phase 0.2)  
**Changes needed:**

```go
// Pseudocode for the updated check:
func CheckPermission(userID, objectType, action, scope string) bool {
    // 1. Collect ALL permissions from ALL roles assigned to this user → union set U
    perms := collectAllPermissions(userID)
    
    // 2. If any permission matches :all variant → allow unconditionally
    if hasPermission(perms, objectType, action, "all") {
        return true  // broad access wins
    
    // 3. If only :own variants exist in union → enforce ownership check
    if hasOwnVariant(perms, objectType, action) {
        return verifyOwnership(userID, objectType, action)
    
    // 4. No matching permission at all → deny
    return false
}

func verifyOwnership(userID string, objectType string, action string) bool {
    switch {
    case action == "create":
        // User must own BOTH source AND target objects
        return userOwnsObject(userID, sourceObjID) && userOwnsObject(userID, targetObjID)
    
    case action == "read" || action == "list":
        // User owns at least ONE endpoint (source OR target)  
        return userOwnsObject(userID, sourceObjID) || userOwnsObject(userID, targetObjID)
    
    case action == "update" || action == "delete":
        // User created the relationship
        return relationshipCreatedByUser(userID, relationshipID)
    
    default:
        return false
    }
}
```

#### Step 2.2 — Multi-role union collection [ ]
**Current behavior:** Need to check if middleware already unions across roles  
**If not:** Add logic to iterate all user-role assignments and collect permissions from each role

#### Step 2.3 — Object type resolution (optional for v1) [ ]
Currently, the permission middleware likely checks permissions based on endpoint path (`/api/v1/relationships`) rather than the object's actual `object_type_id`. For v1:
- Relationships continue to use their own routing and `relationships:*` permissions
- When new typed objects (portfolio, asset, etc.) are added in future phases, type resolution via `object_type_id` lookup will be needed

---

### Phase 3: Test Updates & Verification

> **Goal:** Ensure all tests pass with the new permission model.

#### Step 3.1 — Fix test-rbac-relationships.sh RL-3 [ ]
**Current issue:** GET `/api/v1/relationships` returns 403 for `user` role  
**After Phase 1 fix:** User gets `relationships:read:own` (no :all assigned to user role) → returns 200 with filtered list containing only relationships where user owns at least one endpoint

#### Step 3.2 — Add scoped permission enforcement tests [ ]
- [ ] Test that `relationships:create:own` blocks creation when user doesn't own both endpoints (source AND target)
- [ ] Test that `read:own` filters results to relationships where user owns at least one endpoint  
- [ ] Test that `update/delete:own` only allows modification of self-created relationships

#### Step 3.3 — Add multi-role union tests [ ]
- [ ] User with BOTH `user` role + `relationship-viewer` → should get union (relationships:create:own from user, read:all from viewer)
- [ ] User with ONLY `user` role → no access to relationships they don't own

---

### Phase 4: Documentation & Migration Guide

> **Goal:** Finalize documentation and prepare deployment guidance.

#### Step 4.1 — Finalize architecture document [ ]
- [ ] Add middleware pseudocode from Phase 2 as implementation reference
- [ ] Document ownership verification logic per permission type
- [ ] Add migration steps for existing deployments (how to go from flat → scoped)

#### Step 4.2 — Migration guide for existing deployments [ ]
Document the safe upgrade path:
1. Deploy new permissions (Phase 1) — backward compatible, no breaking changes
2. Update middleware (Phase 2) — users with old role assignments still work via union semantics
3. Reassign roles as needed (`object-type-admin` → remove from relationships; add `relationship-admin/viewer`)

---

## Summary Checklist

| Phase | Key Deliverable | Status |
|-------|----------------|--------|
| 0.1 | Updated architecture doc (two-rule model, scoped-only perms) | Done |
| 0.2 | Permission assignment validation + constraint enforcement | Not started |
| 0.3 | Middleware audit report | Not started |
| 0.4 | Updated RBAC docs with roles | Not started |
| 1.1 | Scoped permission migration (000008) | Not started |
| 1.2 | New role migrations (relationship-admin/viewer) | Not started |
| 1.3 | Updated assignments (000009) + rollback/reapply | Not started |
| 2.1 | Scoped variant enforcement in middleware | Not started |
| 2.2 | Multi-role union collection logic | Not started |
| 2.3 | Object type resolution (future-proofing) | Deferred to Phase N+1 |
| 3.1 | Fix RL-3 test failure | Not started |
| 3.2 | Scoped enforcement tests | Not started |
| 3.3 | Multi-role union tests | Not started |
| 4.1 | Final architecture doc with pseudocode | Not started |
| 4.2 | Deployment migration guide | Not started |
