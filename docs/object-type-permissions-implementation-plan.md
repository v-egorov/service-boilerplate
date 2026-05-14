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

---

## Permission Resolution Model

Two rules operate at different levels:

### Rule 1 — Within-role: Most Specific Wins

When multiple permissions for the same action exist in a single role, the most granular available variant is used:
```
Role grants: read:own + read:all → read:all takes effect (broader scope covers more)
Role grants: create:own           → only that variant applies (no :all in union)
```

### Rule 2 — Across-roles: Union then Broadest Wins

When multiple roles grant permissions for the same action, all are collected into a union set. If ANY role provides the broader `:all` variant, ownership checks are bypassed entirely:

| Available in Union | Resolution | Ownership Check? |
|--------------------|------------|------------------|
| `*:all` only       | Allow unconditionally | No |
| `*:own` only       | Conditional allow     | Yes — on endpoints |
| Both              | Same as `*:all`        | No |
| Neither           | Deny                  | N/A |

**Key insight:** Different actions are independent. Having `create:own` + `read:all` means create requires ownership verification but read does not. Actions never interfere with each other.

---

## Final Roles & Permissions Matrix

### Object Types Permissions

| Role | create | update | delete |
|------|--------|--------|--------|
| admin | ✓ | ✓ | ✓ |
| object-type-admin | ✓ | ✓ | ✓ |
| user, relationship-admin, relationship-viewer | — | — | — |

### Objects Permissions

| Role | create | read | update | delete | Scope constraint |
|------|--------|------|--------|--------|------------------|
| admin | all 3 variants | both variants | both variants | both variants | None |
| object-type-admin | — | read:all | all 3 variants | both variants | — |
| user | create (default) | own, all | own, all | own, all | :own checks ownership |

### Relationships Permissions (NEW scoped model)

| Role | create | update | delete | read | Notes |
|------|--------|--------|--------|------|-------|
| **admin** | `:own`, `:all` | `:own`, `:all` | `:own`, `:all` | `:own`, `:all` | Full CRUD, no restrictions |
| **relationship-admin** (NEW) | same as admin | same as admin | same as admin | same as admin | Dedicated for relationship lifecycle mgmt |
| **relationship-viewer** (NEW) | — | — | — | `:own`, `:all` | Read-only, audit/discovery |
| **user** | `:own` only | `:own` only | `:own` only | `:own` only | Own objects endpoint constraint |
| **object-type-admin** | — | — | — | — | Not applicable to relationships |

### Scoped Variants Definition

| Permission | Enforcement (for "own" variants) |
|------------|----------------------------------|
| `create:own` | User owns BOTH source AND target objects |
| `read:own`   | User owns AT LEAST ONE endpoint (source OR target) |
| `update:own` / `delete:own` | User created the relationship (`created_by = user_id`) |

### Ownership Check Logic

```
For :own permissions, middleware verifies ownership BEFORE allowing action.
For :all permissions (or if any role grants :all), no ownership check is performed.
Ownership verification checks:
  - create:own → owner_id of source_object AND target_object match user ID
  - read:own   → owner_id of source_object OR target_object matches user ID  
  - update/delete:own → created_by field on relationship object matches user ID
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

#### Step 0.3 — Update `docs/rbac-objects-service.md` [ ]
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
- Add `relationship-admin` → ALL scoped variants (same as admin)
- Update `user` → only scoped variants (`:own` for all CRUD, `:own+all` for read)
- Add `relationship-viewer` → ONLY `read:own` + `read:all`

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
**After Phase 1 fix:** User gets `read:own` + `read:all` → should return 200

#### Step 3.2 — Add scoped permission enforcement tests [ ]
- [ ] Test that `create:own` blocks creation when user doesn't own both endpoints
- [ ] Test that `read:own` filters results to relationships where user owns at least one endpoint  
- [ ] Test that `update/delete:own` only allows modification of self-created relationships

#### Step 3.3 — Add multi-role union tests [ ]
- [ ] User with BOTH `user` role + `relationship-viewer` → should get union (create:own from user, read:all from viewer)
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