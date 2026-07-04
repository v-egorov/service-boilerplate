## Why

The permission system has a two-layer gap that prevents object-types and relationship-types from working with scoped permissions, and leaves `hasPermission()` unable to match empty-scope DB entries against scoped requirements. This is visible in E2E MCP testing: `list_object_types` returns 403 "Insufficient permissions" for all users including admins because permiddleware generates `{type}:read:all/{own}` but the DB only has flat `{type}:read` entries, and `hasPermission()` has no rule to match empty scope against required scopes.

This is a pre-existing gap from the June 2026 type_key implementation plan — the architecture spec defines scoped permission variants for object-types (admin gets `:all`, user gets `:own`) but those scoped entries were never added to the database. The matching logic bug compounds it by affecting any future resource added with flat permissions.

## What Changes

- **Add 6 scoped permission entries** to auth-service DB for `object-types`: `read:all`, `read:own`, `update:all`, `update:own`, `delete:all`, `delete:own` (create stays flat per architecture spec)
- **Add 4 scoped permission entries** for `relationship-types`: same scoped variants (create stays flat — you always own what you create)
- **Assign new scoped permissions to roles**: admin and object-type-admin get `:all` on all actions; user gets `read:own` only
- **Fix `hasPermission()` in auth-service** to treat empty-scope DB entries as unrestricted (match any required scope level) — one-line defensive rule

## Capabilities

### New Capabilities
- `auth-service-permission-resolution`: Permission matching logic including empty-scope-as-unrestricted rule

### Modified Capabilities
- `mcp-server`: MCP agent's permission grants now include object-types and relationship-types scoped entries, enabling list_object_types tool to work without gateway-trust skip path

## Impact

- **Auth service**: DB migrations (4 new migration pairs in dev: scoped entries for object-types, scoped entries for relationship-types, role assignments, MCP agent permissions), `hasPermission()` logic change
- **Objects service**: No code changes — permiddleware already generates correct scoped strings; fixes are entirely in auth-service layer
- **MCP server**: Indirectly fixed — `list_object_types` tool now works through normal RBAC path when MCP agent has object-types read permissions assigned to its role
- **Existing tests**: `TestHasPermission_ScopedVariants` gains new test cases for empty-scope matching; existing behavior preserved
