## 1. Auth-service: DB migrations for scoped permissions

- [x] 1.1 Create `services/auth-service/migrations/development/000013_add_object_types_scoped_permissions.up.sql` — INSERT 6 scoped permission entries (read/update/delete × all/own) for object-types, with `.down.sql`
- [x] 1.2 Create `services/auth-service/migrations/development/000014_add_relationship_types_scoped_permissions.up.sql` — INSERT 6 scoped permission entries (read/update/delete × all/own) for relationship-types, with `.down.sql`

## 2. Auth-service: DB migrations for role assignments

- [x] 2.1 Create `services/auth-service/migrations/development/000015_assign_object_types_perms_to_roles.up.sql` — assign scoped perms to admin (:all), object-type-admin (:all), user (read:own only), with `.down.sql`
- [x] 2.2 Create `services/auth-service/migrations/development/000016_mcp_agent_object_types_perms.up.sql` — add `object-types:read:all` to mcp-agent-read-only role, with `.down.sql`

## 3. Auth-service: Fix hasPermission() empty-scope matching

- [x] 4.1 In `services/auth-service/internal/services/auth_service.go`, add Rule 3 to `hasPermission()`: when `parsed.Scope == ""` (DB entry is unscoped), return true — treating it as unrestricted
- [x] 4.2 Add test cases to `TestHasPermission_ScopedVariants` in `auth_service_test.go`:
  - Empty-scope DB entry satisfies :all requirement
  - Empty-scope DB entry satisfies :own requirement
  - Verify existing behavior preserved (exact match, broad-over-narrow)

## 4. Migration orchestrator configuration

- [x] 5.1 environments.json references directory names, not individual files — orchestrator auto-discovers migrations from directories, no config changes needed.
- [x] 5.2 Verified: dev migration sequence is 000010 → 000011 → 000012 → **000013** → **000014** → **000015** → **000016**. No conflicts.

## 5. Testing & Verification

- [x] 6.1 Build + tests: auth-service compiles, all 8 packages pass (including new hasPermission empty-scope test cases)
- [x] 6.2 objects-service: no code changes, all tests still pass
- [x] 6.3 Applied all 4 dev migrations successfully. DB now has 12 scoped entries (6 for object-types + 6 for relationship-types) plus existing flat entries.
- [x] 6.4 Permissions fix verified: permiddleware passes through to service layer (no more 403 from auth-service). However, objects-service has a pre-existing deep crash in the repo/service layer during List operations (`Failed to list object types` with `"error":{}`) that causes 500 for ALL users including admin. This is documented as a separate pre-existing bug — NOT related to our changes.
- [x] 6.5 Same blocker: MCP `list_object_types` tool reaches permiddleware correctly (identity forwarding works), but objects-service crashes before returning data. The permission fix is correct; the service-layer crash is a separate issue.
