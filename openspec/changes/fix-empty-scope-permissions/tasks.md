## 1. Auth-service: DB migrations for scoped permissions

- [ ] 1.1 Create `services/auth-service/migrations/development/000013_add_object_types_scoped_permissions.up.sql` — INSERT 6 scoped permission entries (read/update/delete × all/own) for object-types, with `.down.sql`
- [ ] 1.2 Create `services/auth-service/migrations/development/000014_add_relationship_types_scoped_permissions.up.sql` — INSERT 4 scoped permission entries for relationship-types, with `.down.sql`

## 2. Auth-service: DB migrations for role assignments

- [ ] 2.1 Create `services/auth-service/migrations/development/000015_assign_object_types_perms_to_roles.up.sql` — assign scoped perms to admin (:all), object-type-admin (:all), user (read:own only), with `.down.sql`
- [ ] 2.2 Create `services/auth-service/migrations/development/000016_mcp_agent_object_types_perms.up.sql` — add `object-types:read:all` to mcp-agent-read-only role, with `.down.sql`



## 4. Auth-service: Fix hasPermission() empty-scope matching

- [ ] 4.1 In `services/auth-service/internal/services/auth_service.go`, add Rule 3 to `hasPermission()`: when `parsed.Scope == ""` (DB entry is unscoped), return true — treating it as unrestricted
- [ ] 4.2 Add test cases to `TestHasPermission_ScopedVariants` in `auth_service_test.go`:
  - Empty-scope DB entry satisfies :all requirement
  - Empty-scope DB entry satisfies :own requirement
  - Verify existing behavior preserved (exact match, broad-over-narrow)

## 5. Objects-service: Migration orchestrator configuration update

- [ ] 5.1 Update `migration-orchestrator/environments.json` to include the 4 new dev migration files (000013–000016)
- [ ] 5.2 Verify no conflicts with existing migration sequences

## 6. Testing & Verification

- [ ] 6.1 Run `make build-auth-service && make test-auth-service` — verify all tests pass including new hasPermission cases
- [ ] 6.2 Run `make build-objects-service && make test-objects-service` — no code changes but verify nothing broke
- [ ] 6.3 Apply dev migrations: `make db-migrate-up SERVICE_NAME=auth-service` and verify 10 new permissions (6 object-types + 4 relationship-types) exist in DB
- [ ] 6.4 Test object-types GET/PUT/DELETE through gateway with admin JWT — should return 200 (not 403)
- [ ] 6.5 Test MCP `list_object_types` tool — should work without gateway-trust skip path
