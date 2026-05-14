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
portfolio:read              → read any Portfolio (no scope = default, requires ownership check)
portfolio:read:own          → read only owned Portfolios  
portfolio:read:all          → read any Portfolio (broad access, no ownership check)
asset:create                → create Assets (same as asset:create:default)
relationships:delete:own    → delete relationships created by the user
*:*                         → super-admin (all actions on all types, future)
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
VALID:     user → portfolio:read:only
VALID:     admin → portfolio:read:all  
INVALID:   user → portfolio:read:own + portfolio:read:all  ← redundant, broadest always wins
```

When multiple roles are involved, each role should assign ONLY the most specific level that covers its scope. If a user has both `portfolio:read:own` (from "user" role) and `portfolio:read:all` (from "admin" role), this is correct — it's an across-roles scenario handled by Rule B.

#### Rule B: Across-roles — Union Then Broadest Wins

When permissions are granted by MULTIPLE roles, they are collected into a union set. If ANY role provides the broader `:all` variant for an action, that bypasses ownership checks entirely regardless of other roles' narrower-scoped variants.

| Available in Union | Resolution | Ownership Check? |
|--------------------|------------|------------------|
| `*:all` only       | Allow unconditionally | No |
| `*:own` only       | Conditional allow     | Yes — on endpoints/creator |
| Both              | Same as `*:all`        | No (broad access wins) |
| Neither           | Deny                  | N/A |

**Important:** Different actions are independent. Having `create:own` + `read:all` means create requires ownership verification but read does not. Actions never interfere with each other.

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

**Unscoped permissions:** No ownership check is performed — if a role grants broad access, the action proceeds immediately.

### 6. Locked-Down by Default

New object types start with **no default permissions**. Explicit grants are required before any role or user can perform actions on objects of that type. This applies during initial rollout AND when new types are registered in the future.

---

## Roles & Permissions Matrix

### Object Types Permissions

| Role | create | update | delete |
|------|--------|--------|--------|
| admin | ✓ | ✓ | ✓ |
| object-type-admin | ✓ | ✓ | ✓ |
| user | — | — | — |
| relationship-admin, relationship-viewer | — | — | — |

### Objects Permissions

| Role | create | read | update | delete | Scope constraint |
|------|--------|------|--------|--------|------------------|
| admin | all 3 variants | both variants | both variants | both variants | None |
| object-type-admin | — | read:all | own, all | own, all | :own checks ownership |
| user | create (default) | own, all | own, all | own, all | :own checks ownership |

### Relationships Permissions

| Role | create | read | update | delete | Notes |
|------|--------|------|--------|--------|-------|
| **admin** | `:own`, `:all` | `:own`, `:all` | `:own`, `:all` | `:own`, `:all` | Full CRUD, no restrictions |
| **relationship-admin** (NEW) | same as admin | same as admin | same as admin | same as admin | Dedicated for relationship lifecycle mgmt |
| **relationship-viewer** (NEW) | — | `:own`, `:all` | — | — | Read-only, audit/discovery |
| **user** | `:own` only | `:own` only | `:own` only | `:own` only | Own objects endpoint constraint |
| **object-type-admin** | — | — | — | — | Not applicable to relationships |

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
1. Union set U = {create:own, create:all, read:own, read:all, update:own, update:all, delete:own, delete:all} 
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
| `scope` | `own`, `all` (required) | Restriction level — mandatory for all permissions |

### Examples of Well-Formed Permissions

```
portfolio:read              → no scope specified, middleware will require ownership check
portfolio:read:own          → read only owned Portfolios  
portfolio:read:all          → read any Portfolio (broad access)
asset:create                → create Assets (default scope, requires ownership verification)
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

**No immediate migrations required.** The approach will be implemented incrementally following `object-type-permissions-implementation-plan.md`:

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
| Object-type-scoped permissions | `<type>:<action>[:<scope>]` format, scope always present |
| Abstract vs concrete | `objects:*` for untyped base; `<type>:*` exclusively for typed objects |
| Permission resolution (within-role) | Duplicate assignments blocked — each role gets ONE scoped level per type:action pair |
| Permission resolution (across-roles) | Union of all roles, broadest (`:all`) wins over narrower (`:own`) |
| Actions are independent | `create:own` + `read:all` → create needs ownership, read does not |
| Assignment constraint | POST/PATCH/PUT on role_permissions blocks duplicate scoped variants for same type+action on one role |
| Relationships | Special system object type with its own scoped permission namespace |
| New types | Locked down by default — explicit grants required |
| Roles | 5 roles: admin, object-type-admin, user, relationship-admin (NEW), relationship-viewer (NEW) |