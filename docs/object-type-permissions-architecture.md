# Object-Type-Scoped Permissions Architecture

**Status**: Proposed  
**Date**: 2026-05-14  
**Related**: `rbac-objects-service.md`, `object-type-permissions-implementation-plan.md`

---

## Overview

This document describes the long-term approach for permission management in the objects-service, designed to support the evolution of new object types derived from the abstract CTE base object.

The approach addresses:
1. How permissions are scoped to specific object types
2. How permission checks determine which permission applies to a given request
3. How new object types integrate with the permission system (locked down by default)
4. How relationships (as object types) fit into this model

---

## Core Principles

### 1. Object-Type-Scoped Permissions

Each concrete object type has its own namespace of permissions:

```
<object_type>:<action>[:<scope>]
```

Examples:
```
portfolio:read:own          → read only owned Portfolios  
portfolio:read:all          → read any Portfolio (broad access, no ownership check)
object-types:create         → create new object type definitions (no scope — you always own what you create)
relationships:create:own   → create relationships where user owns BOTH endpoints
relationships:delete:own    → delete relationships created by the user
*:*                         → super-admin pattern (all actions on all types, future)
```

### 2. Abstract vs Concrete Object Types

| Category | Description | Examples |
|----------|-------------|----------|
| **Abstract (base)** | The CTE root table `objects_service.objects`. Permissions apply to untyped objects or as fallbacks. | `objects:*` |
| **Concrete (typed)** | Objects with a specific `object_type_id`. Permissions are type-specific. | `portfolio:*`, `asset:*`, `relationships:*` |

**Key rule:** The `objects:*` permissions apply **only** to abstract/untyped objects. Concrete typed objects use their own type-specific permissions exclusively — no fallback to generic `objects:*` permissions.

### 3. Permission Resolution Model (Two Rules)

#### Rule A: Within-role — No Duplicate Specificity (Assignment Constraint)

Assigning both scoped and unscoped variants for the same type:action to a SINGLE role is redundant and should be prevented or warned against at assignment time:

```
VALID:     user →     portfolio:read:own
VALID:     admin → portfolio:read:all  
INVALID:   user → portfolio:read:own + portfolio:read:all  ← redundant, broadest always wins
```

When multiple roles are involved, each role should assign ONLY the most specific level that covers its scope. If a user has both `portfolio:read:own` (from "user" role) and `portfolio:read:all` (from "admin" role), this is correct — it's an across-roles scenario handled by Rule B.

#### Rule A Edge Case: Duplicate Scoped Variants in Same Role

The API enforces that each role gets at most one scoped variant per type+action pair. However, if a race condition or data corruption causes both `type:action:own` and `type:action:all` to exist for the same role, resolution follows Rule B semantics — broadest scope wins (`:all` overrides `:own`).

This is logged at warning level with: role_id, permission names involved, affected user count. The system treats this scenario identically to the cross-role union case because the practical effect is the same: a broader scope exists and should take precedence.

#### Rule B: Across-roles — Union Then Broadest Wins

When permissions are granted by MULTIPLE roles, they are collected into a union set. If ANY role provides the broader `:all` variant for an action, that bypasses ownership checks entirely regardless of other roles' narrower-scoped variants.

| Available in Union | Resolution | Ownership Check? |
|--------------------|------------|------------------|
| `*:all` only       | Allow unconditionally | No |
| `*:own` only       | Conditional allow     | Yes — on endpoints/creator |
| Both              | Same as `*:all`        | No (broad access wins) |
| Neither           | Deny                  | N/A |

**Important:** Different actions are independent. Having `update:own` + `read:all` means update requires ownership verification but read does not. Actions never interfere with each other. Note that the `create` action uses flat permission names on objects/types (no scoped variants) since you always own what you create; scoped variants for `create` only apply to relationships where endpoint ownership must be verified.

### 4. Relationships as Object Types

Relationships are stored in the same `objects_service.objects` table (with `relationship_type_id` set). In this architecture, relationships are treated as a special system object type with their own permission namespace.

**Scoped permissions:**
```
relationships:create:own      → create relationships where user owns BOTH endpoints
relationships:create:all      → create ANY relationship (admin-level)
relationships:read:own        → read relationships involving at least one owned endpoint  
relationships:read:all        → read ALL relationships (audit/discovery)
relationships:update:own      → update only relationships the user created
relationships:update:all      → update any relationship
relationships:delete:own      → delete only relationships the user created
relationships:delete:all      → delete any relationship
```

### 5. Ownership Check Logic

For scoped `:own` variants, middleware verifies ownership BEFORE allowing access:

| Permission | "Own" means... | Enforcement |
|------------|----------------|-------------|
| `create:own` | User owns BOTH source AND target objects | Check `owner_id` on both endpoint objects |
| `read:own`   | User owns at least ONE endpoint (source OR target) | Check `owner_id` on either object |
| `update:own` / `delete:own` | User created the relationship (`created_by = user_id`) | Check `created_by` field on relationship object |

**Create scoped variants are exception-only:** On plain objects and object-types, `create` uses flat permission names — you always own what you create. On relationships, `create` **does** use scoped variants (`create:own`, `create:all`) because creating a relationship requires ownership of pre-existing endpoint objects (source AND target), not the relationship instance itself.

### 6. Locked-Down by Default

New object types start with **no default permissions**. Explicit grants are required before any role or user can perform actions on objects of that type. This applies during initial rollout AND when new types are registered in the future.

---

## Roles & Permissions Matrix

Each cell shows the single scoped permission variant assigned to that role for the given action. Per Rule A, a role never holds more than one scope level per type:action combination. Cross-role access is determined by Rule B (union then broadest wins).

### Object Types Permissions (registry)

Permissions on the `object_types` table itself (type definitions/schema). The `create` action uses flat permission names since you always own what you create. All roles have read access; write access is restricted to admin and object-type-admin.

| Role | create | read | update | delete |
|------|--------|------|--------|--------|
| admin | `create` | `all` | `all` | `all` |
| object-type-admin | `create` | `all` | `all` | `all` |
| user, relationship-admin, relationship-viewer | — | `all` | — | — |

### Objects Permissions (data instances)

Permissions on actual object records derived from the types above. The `create` action has no scoped variants — it uses flat permission names since you always own what you create.

| Role | create | read | update | delete |
|------|--------|------|--------|--------|
| admin | `create` | `all` | `all` | `all` |
| object-type-admin | — | `all` | — | — |
| user | `create` | `own` | `own` | `own` |

**Scope semantics:**
- `:own` — ownership check required (user must be `created_by` of the object)
- `:all` — no ownership check; unrestricted access to any object of this type

### Relationships Permissions

The `create` action on relationships uses scoped variants because creating a relationship requires ownership of pre-existing endpoint objects (source AND target). This differs from plain object creation, where you always own what you create.

| Role | create | read | update | delete | Notes |
|------|--------|------|--------|--------|-------|
| **admin** | `all` | `all` | `all` | `all` | Full CRUD, no restrictions |
| **relationship-admin** (NEW) | `all` | `all` | `all` | `all` | Dedicated relationship lifecycle management |
| **relationship-viewer** (NEW) | — | `all` | — | — | Read-only, audit/discovery |
| **user** | `own` | `own` | `own` | `own` | Can act on relationships involving owned endpoints; if also assigned `relationship-viewer`, union grants `read:all` |
| **object-type-admin** | — | — | — | — | Not applicable to relationships |

> **Relationship-specific scope semantics:** Unlike plain objects, relationship ownership is asymmetric:
> - `create:own` — user may create a relationship only if they own **both source AND target** endpoint objects
> - `read:own` — user may read a relationship if they own **at least one** of the two endpoints (source OR target)
> - `update:own` / `delete:own` — user may modify/delete only relationships where they are the `created_by`

---

## Permission Check Flow

When a request accesses an object, the permission middleware performs:

```
1. Collect ALL permissions from ALL roles assigned to this user → union set U
2. Determine requested action (create/read/update/delete) for target type T
3. If any role in union grants :all variant for T+action → allow unconditionally
4. If only :own variants exist in union → verify ownership per Table 5 rules
5. No matching permission at all → deny (403)
```

### Example: User Creates Relationship

```
User: test.user (role: user)
Request: POST /api/v1/relationships {source_obj_id: X, target_obj_id: Y}
Permission required: relationships:create

Check flow:
1. Union set U = {create:own, read:own, update:own, delete:own}  ← only :own variants from user role
2. Action = create, scope needed = ? (no :all in union)  
3. Only create:own exists → ownership check required
4. Verify: test.user owns source_obj X? YES. Owns target_obj Y? YES.
5. Result: permitted, relationship created with created_by = user_id
```

### Example: User Lists Relationships

```
User: test.user (role: user)  
Request: GET /api/v1/relationships
Permission required: relationships:read

Check flow:
1. Union set U = {create:own, read:own, update:own, delete:own} 
2. Action = read → only :own in union
3. Ownership check for read:own → user owns at least one endpoint of each relationship returned
4. Result: 200 with filtered list (only relationships where user is an endpoint owner)
```

### Example: Admin Lists All Relationships

```
User: admin (role: admin)
Request: GET /api/v1/relationships  
Permission required: relationships:read

Check flow:
1. Union set U = {create:all, read:all, update:all, delete:all} 
2. Action = read → :all variant exists in union
3. :all takes precedence over :own → no ownership check needed
4. Result: 200 with ALL relationships (no filtering)
```

---

## Object Type Registry

Each object type must be registered in `objects_service.object_types` with:

| Field | Description |
|-------|-------------|
| `id` | Unique identifier |
| `name` | Human-readable name (e.g., "Portfolio") |
| `type_key` | Machine-readable key (e.g., "portfolio") — used as permission namespace prefix |
| `description` | Description of the type |
| `is_system` | Boolean — true for built-in types like "relationships" |
| `parent_type_id` | Optional — for hierarchical type inheritance (future) |

### Reserved System Types

| Type Key | Description |
|----------|-------------|
| `relationships` | Relationship instances (system-managed, uses its own permission namespace) |

### Example: Registering a New Object Type

```sql
INSERT INTO objects_service.object_types (name, type_key, description, is_system)
VALUES ('Document', 'document', 'Document objects', false);
```

After registration, no default permissions are granted — administrator must explicitly assign them via role assignments.

---

## Permission Namespace Conventions

### Format

```
<type_key>:<action>[:<scope>]
```

### Components

| Component | Values | Description |
|-----------|--------|-------------|
| `type_key` | lowercase, alphanumeric + hyphens | The object type's type_key from registry |
| `action` | `create`, `read`, `update`, `delete` | What action is being performed |
| `scope` | `own`, `all` (optional) | Restriction level — present for read/update/delete; absent for create on objects/types. On relationships, scope applies to endpoint ownership checks. |

### Examples of Well-Formed Permissions

```
portfolio:read:own          → read only owned Portfolios  
portfolio:read:all          → read any Portfolio (broad access)
object-types:create         → create new object type definitions (no scope — you always own what you create)
relationships:create        → create relationships (flat name; scoped variants :own/:all exist for endpoint ownership checks)
relationships:delete:own    → delete own relationships only
relationships:read:all      → audit/discovery mode for all relationships
*:*                         → super-admin pattern (future, applies to all types)
```

---

## Transition Strategy

### Current State (As of 2026-05-14)

- Permissions use flat resource names: `objects`, `object-types`, `relationships`
- User role has `objects:read:own` which grants access to owned objects
- Permission middleware checks permissions based on endpoint path, not object type
- Relationships have separate permission namespace (`relationships:*`)

### Target State

- Permissions scoped by object type with explicit scopes: portfolio:read:own, relationships:create:all, etc.
- `objects:*` permissions become abstract-only fallbacks for untyped objects
- Each concrete object type uses its own namespace (portfolio:*, asset:*, relationships:*)
- Permission check collects union from all roles, then resolves scoped variants

### Migration Path

The approach will be implemented incrementally following `object-type-permissions-implementation-plan.md`:

1. **Phase 0:** Update documentation + audit current middleware enforcement behavior
2. **Phase 1:** Add scoped permission definitions + new roles (migrations only, backward compatible)  
3. **Phase 2:** Enhance middleware to enforce scoped variants and implement multi-role union logic
4. **Phase 3:** Update tests + verify all RBAC test scripts pass with new model

**Key invariant during transition:** If no type-specific permission is found for an action, the request is denied (fail-closed). This means existing role assignments must be explicitly migrated to scoped variants before enforcement begins.

---

## Future Considerations

### Hierarchical Object Types

Future extension: object types can have a `parent_type_id` forming a hierarchy:

```
objects (abstract base)
  └── portfolio
        └── investment_portfolio
              └── retirement_account
```

Permission inheritance: `investment_portfolio:read:all` implies `portfolio:read:all`. This would require recursive permission lookup through the type hierarchy.

**Not implemented in initial version** — requires `parent_type_id` column and middleware changes to walk the tree upward when no exact match is found.

### Capability-Based Access (CBAC)

Alternative approach where permissions are expressed as capabilities attached to object instances:

```json
{
  "object_id": 123,
  "owner_id": "user-uuid",  
  "capabilities": [
    { "role": "viewer", "actions": ["read"], "grantee": "user-uuid-2" }
  ]
}
```

**Deferred** — requires significant architectural changes. Not on the roadmap for initial implementation.

### Type-Specific Default Permissions

When a new object type is registered, a webhook or event could notify an admin service to prompt for initial permission grants. This is out of scope for initial implementation but provides a natural extension point as the system grows.

---

## Related Documents

- `rbac-objects-service.md` — Current RBAC implementation details
- `object-type-permissions-implementation-plan.md` — Detailed todo-style implementation plan
- `service-patterns-reference.md` — Service layer patterns  
- `inter-service-communication.md` — How services communicate

---

## Summary

| Principle | Implementation |
|-----------|----------------|
| Object-type-scoped permissions | `<type>:<action>[:<scope>]` format; `create` uses flat names on objects/types (no scope), scoped variants only for relationships (endpoint ownership) |
| Abstract vs concrete | `objects:*` for untyped base; `<type>:*` exclusively for typed objects |
| Permission resolution (within-role) | Duplicate assignments blocked — each role gets ONE scoped level per type:action pair |
| Permission resolution (across-roles) | Union of all roles, broadest (`:all`) wins over narrower (`:own`) |
| Actions are independent | `update:own` + `read:all` → update needs ownership, read does not. Actions never interfere with each other. |
| Assignment constraint | POST/PATCH/PUT on role_permissions blocks duplicate scoped variants for same type+action on one role |
| Relationships | Special system object type with its own scoped permission namespace; `create` can have scope (endpoint ownership) unlike other types |
| New types | Locked down by default — explicit grants required |
| Roles | 5 roles: admin, object-type-admin, user, relationship-admin (NEW), relationship-viewer (NEW) |
