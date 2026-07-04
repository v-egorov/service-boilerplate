## Context

The permission system uses a two-layer model: auth-service stores permissions in `auth_service.permissions` (flat or scoped) and resolves them via `hasPermission()`, while objects-service's permiddleware constructs permission strings from `RouteConfig{TypeKey, HTTPMethod}` and calls auth-service to check.

**Current state:**
- `objects` resource has both flat (`create`) and scoped (`read:all/own`, `update:all/own`, `delete:all/own`) permissions ✅ working
- `relationships` resource has all scoped variants ✅ working  
- `object-types` resource only has flat entries — permiddleware generates scoped strings for GET/PUT/DELETE but no matching DB entries ❌ broken
- `relationship-types` resource same pattern as object-types ❌ broken
- `hasPermission()` in auth-service matches exact scopes and allows `:own ← :all`, but cannot match empty-scope (`""`) against required scoped strings

The architecture spec (`docs/object-type-permissions-architecture.md`) defines scoped permission variants for object-types (admin gets `:all`, user gets `:own`), but those entries were never added to the database during the June 2026 type_key implementation.

## Goals / Non-Goals

**Goals:**
- Add missing scoped permission entries for `object-types` and `relationship-types` to auth-service DB
- Assign new scoped permissions to roles per architecture spec matrix (admin/object-type-admin get `:all`, user gets `read:own`)
- Fix `hasPermission()` empty-scope matching — treat unscoped as unrestricted
- Update MCP agent's permission grants to include object-types read permissions

**Non-Goals:**
- No changes to objects-service code or permiddleware logic (it already generates correct scoped strings)
- No schema changes to auth-service tables
- No migration of existing flat-permission resources (users, profile, etc.) — they work fine through their own permission paths
- No changes to the `create` action on object-types/relationship-types (stays flat per architecture spec)

## Decisions

### Decision 1: Use separate migrations for object-types and relationship-types scoped entries
**Rationale:** They are different resources with different role assignment patterns. Keeping them in separate migration files makes each independently reviewable and reversible. Object-types needs 8 new permission entries + role assignments; relationship-types needs 4 (create stays flat).

### Decision 2: Fix `hasPermission()` rather than forcing all resources to use scoped permissions
**Rationale:** The architecture spec intentionally uses flat permissions for some actions (e.g., `object-types:create` — you always own what you create) and some resources (users, profile, roles) legitimately stay flat. Adding a defensive rule that empty scope = unrestricted is more general-purpose and prevents future bugs when new flat-only resources are added.

### Decision 3: MCP agent gets object-types:read:all (not :own)
**Rationale:** The MCP agent operates on behalf of the developer who has full admin access to object types. There's no meaningful "ownership" concept for type definitions — any user with `object-types:read` should see all defined types, not just ones they created. `:all` is semantically correct here.

### Decision 4: Add scoped entries via INSERT ON CONFLICT DO NOTHING (not DELETE+INSERT)
**Rationale:** Idempotent migrations are preferred. Using `ON CONFLICT (name) DO NOTHING` allows safe reapplication during development and staging without data loss or ordering issues. The down migration simply deletes the added rows by name filter.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| New scoped permissions change existing behavior for roles that only had flat perms | Minimal impact — admin already has read:all on objects, extending to object-types is consistent; user role gains read:own which is narrower than current unscoped "read" (fail-closed), but this aligns with architecture spec intent |
| `hasPermission()` empty-scope rule could mask data inconsistencies | The rule is semantically correct: empty scope = unrestricted. If a permission was meant to be scoped, it should have been added in the migration. This doesn't hide bugs — it handles the "no boundary defined" case correctly |
| Migration ordering — must add permissions before assigning to roles | Migrations are sequential within each env; role assignment migration references permission names so it only works after they exist. Use `ON CONFLICT DO NOTHING` for safety |

## Migration Plan

1. **Pre-deploy:** Run migrations on dev/staging first (dev has 000013/000014, staging mirrors)
2. **Deploy auth-service changes** (hasPermission fix + DB migrations) — single deploy step
3. **Verify:** Test object-types and relationship-types endpoints through gateway with admin JWT
4. **Rollback:** Down migrations drop the 8+4 new permission entries; hasPermission change is backward-compatible (empty scope was already failing, so rollback restores same behavior)

## Open Questions

- None — all decisions captured above.
