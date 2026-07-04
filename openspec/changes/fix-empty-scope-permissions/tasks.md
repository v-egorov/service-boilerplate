## 1. Auth-service: DB migrations for scoped permissions

- [ ] 1.1 Create `services/auth-service/migrations/{development,staging}/000013_add_object_types_scoped_permissions.up.sql` — INSERT 6 scoped permission entries (read/update/delete × all/own) for object-types, with `.down.sql`
- [ ] 1.2 Create `services/auth-service/migrations/{production}/000009_add_object_types_scoped_permissions.up.sql` and `.down.sql` (renumbered: prod has no 06/07 dev-only migrations)
- [ ] 1.3 Create `services/auth-service/migrations/{development,staging}/000014_add_relationship_types_scoped_permissions.up.sql` — INSERT 4 scoped permission entries for relationship-types, with `.down.sql`
- [ ] 1.4 Create `services/auth-service/migrations/{production}/000010_add_relationship_types_scoped_permissions.up.sql` and `.down.sql` (renumbered)

## 2. Auth-service: DB migrations for role assignments

- [ ] 2.1 Create `services/auth-service/migrations/{development,staging}/000015_assign_object_types_perms_to_roles.up.sql` — assign scoped perms to admin (:all), object-type-admin (:all), user (read:own only), with `.down.sql`
- [ ] 2.2 Create `services/auth-service/migrations/{production}/000011_assign_object_types_perms_to_roles.up.sql` and `.down.sql` (renumbered)

## 3. Auth-service: DB migration for MCP agent object-types permission

- [ ] 3.1 Update existing `000012_mcp_agent_permissions.up.sql` to also insert `object-types:read:all` and assign it to mcp-agent-read-only role, or create a new migration `000016` specifically for this addition (if 000012 is already committed), with `.down.sql`
- [ ] 3.2 Create production equivalent renumbered

## 4. Auth-service: Fix hasPermission() empty-scope matching

- [ ] 4.1 In `services/auth-service/internal/services/auth_service.go`, add Rule 3 to `hasPermission()`: when `parsed.Scope == ""` (DB entry is unscoped), return true — treating it as unrestricted
- [ ] 4.2 Add test cases to `TestHasPermission_ScopedVariants` in `auth_service_test.go`:
  - Empty-scope DB entry satisfies :all requirement
  - Empty-scope DB entry satisfies :own requirement
  - Verify existing behavior preserved (exact match, broad-over-narrow)

## 5. Objects-service: Migration orchestrator configuration update

- [ ] 5.1 Update `migration-orchestrator/environments.json` to include new migration file numbers for dev/staging/production environments if needed (check current numbering)
- [ ] 5.2 Verify no conflicts with existing migration sequences in each environment

## 6. Testing & Verification

- [ ] 6.1 Run `make build-auth-service && make test-auth-service` — verify all tests pass including new hasPermission cases
- [ ] 6.2 Run `make build-objects-service && make test-objects-service` — no code changes but verify nothing broke
- [ ] 6.3 Apply migrations to dev: `make db-migrate-up SERVICE_NAME=auth-service` and verify 8+4 new permissions exist in DB
- [ ] 6.4 Test object-types GET/PUT/DELETE through gateway with admin JWT — should return 200 (not 403)
- [ ] 6.5 Test MCP `list_object_types` tool — should work without gateway-trust skip path
