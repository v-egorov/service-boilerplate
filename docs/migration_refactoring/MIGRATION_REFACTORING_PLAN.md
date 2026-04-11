# Migration System Refactoring Plan - Option 1

**Status:** PHASE 1 COMPLETE - AUTH-SERVICE  
**Approach:** Separate migration paths per environment  
**Goal:** Use golang-migrate natively with environment-specific directories

---

## Executive Summary

This plan reorganizes migrations to use golang-migrate's native capabilities:
1. Create environment-specific migration directories (`development/`, `staging/`, `production/`)
2. Copy/move migration files to appropriate directories
3. Simplify orchestrator to use `m.Up()` and `m.Steps()`
4. Update `environments.json` to reference directories instead of file arrays
5. Keep schema-per-service architecture with `search_path`

---

## Benefits

| Benefit | Before | After |
|---------|--------|-------|
| **golang-migrate usage** | Manual SQL execution | Native `m.Up()` |
| **Migration tracking** | Empty `schema_migrations` | Properly populated via golang-migrate |
| **Environment isolation** | Complex filtering | Separate directories |
| **Rollback behavior** | Custom logic | Native `m.Steps()` |
| **Code complexity** | 460+ lines with filtering | ~330 lines, simpler logic |
| **Developer experience** | Environment tags in SQL | Clear directory structure |

---

## Implementation Progress

### Phase 1: Auth-Service ✅ COMPLETE
- [x] Create environment directories (`development/`, `staging/`, `production/`)
- [x] Copy migration files to appropriate directories
  - Development: 7 migrations (14 files)
  - Staging: 5 migrations (10 files)
  - Production: 5 migrations (10 files)
- [x] Remove root-level migration files
- [x] Update `environments.json` to use directory names
- [x] Refactor `RunMigrationsUp()` - simplified to 32 lines
- [x] Refactor `RunMigrationsDown()` - simplified to 32 lines
- [x] Remove deprecated `isMigrationForEnvironment()` method
- [x] Rename binary from `migration-orchestrator` to `migrate-wrapper`
- [x] Build and test
  - Init: ✅ Schema + tracking table created
  - Migrations up: ✅ All migrations applied, tracking populated
  - Idempotency: ✅ Running again shows no pending
  - Staging environment: ✅ Works independently

### Phase 2: Objects-Service ⏳ PENDING
- [ ] Same steps as Phase 1
- Development: 6 migrations
- Staging/Production: 5 migrations (exclude dev-only `000002_dev_tax_test_data`)

### Phase 3: User-Service ⏳ PENDING
- [ ] Same steps as Phase 1
- [ ] Move seed files to environment directories
  - `scripts/seeds/development/*` → `migrations/development/`
  - `scripts/seeds/staging/*` → `migrations/staging/`
- Development: 6 migrations
- Staging: 4 migrations (includes `000005_dev_add_object_admin`)
- Production: 3 migrations (excludes `000005_dev_add_object_admin`)

### Phase 4: Documentation ⏳ PENDING
- [x] Update MIGRATION_REFACTORING_PLAN.md (this file)
- [ ] Create `docs/migration-guidelines.md`
- [ ] Update service README files

---

## Migration Directory Structure

### Before
```
services/auth-service/migrations/
├── 000001_initial.up.sql
├── 000002_jwt_keys.up.sql
├── 000003_key_rotation.up.sql
├── 000005_object_permissions.up.sql
├── 000008_relationship_permissions.up.sql
├── 000009_dev_relationship_permissions_seed.up.sql
├── 000010_dev_relationship_read_permission_for_users.up.sql
└── environments.json (with file arrays)
```

### After
```
services/auth-service/migrations/
├── development/
│   ├── 000001_initial.up.sql
│   ├── 000002_jwt_keys.up.sql
│   ├── 000003_key_rotation.up.sql
│   ├── 000005_object_permissions.up.sql
│   ├── 000008_relationship_permissions.up.sql
│   ├── 000009_dev_relationship_permissions_seed.up.sql
│   └── 000010_dev_relationship_read_permission_for_users.up.sql
├── staging/
│   ├── 000001_initial.up.sql
│   ├── 000002_jwt_keys.up.sql
│   ├── 000003_key_rotation.up.sql
│   ├── 000005_object_permissions.up.sql
│   └── 000008_relationship_permissions.up.sql
└── production/
    └── (same as staging)
```

---

## Configuration Changes

### Before (environments.json)
```json
{
  "environments": {
    "development": {
      "migrations": [
        "000001_initial.up.sql",
        "000002_jwt_keys.up.sql",
        "... (7 files)"
      ],
      "seed_files": []
    }
  }
}
```

### After (environments.json)
```json
{
  "environments": {
    "development": {
      "migrations": "development",
      "seed_files": []
    },
    "staging": {
      "migrations": "staging",
      "seed_files": []
    },
    "production": {
      "migrations": "production",
      "seed_files": []
    }
  }
}
```

---

## Code Changes

### Before: RunMigrationsUp (150+ lines)
- Load environment config
- Filter migrations by environment tags in SQL files
- Manually execute each migration via SQL
- Track versions manually

### After: RunMigrationsUp (32 lines)
```go
func (o *Orchestrator) RunMigrationsUp(ctx context.Context, environment string) error {
    config, err := o.LoadMigrationConfig()
    if err != nil { return err }
    
    envConfig, exists := config.Environments[environment]
    if !exists { return fmt.Errorf("environment '%s' not found", environment) }
    
    migrationPath := filepath.Join(o.servicePath, "migrations", envConfig.Migrations)
    m, err := o.createMigrateInstance(migrationPath)
    if err != nil { return err }
    defer m.Close()
    
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }
    
    return nil
}
```

### Key Simplifications
1. No environment filtering logic needed
2. Use golang-migrate's native `m.Up()` for execution
3. No manual version tracking
4. `schema_migrations` table populated automatically by golang-migrate

---

## Commands

After refactoring:

```bash
# Initialize tracking for a service
make db-migrate-init SERVICE_NAME=auth-service ENV=development

# Run migrations (uses environment from ENV=)
make db-migrate-up SERVICE_NAME=auth-service ENV=development

# Rollback last migration
make db-migrate-down SERVICE_NAME=auth-service ENV=development

# Check status
make db-migrate-status SERVICE_NAME=auth-service ENV=development

# List migrations
make db-migrate-list SERVICE_NAME=auth-service ENV=development

# Validate migrations
make db-migrate-validate SERVICE_NAME=auth-service ENV=development
```

---

## Testing Results

### Auth-Service Phase 1 Testing
| Test | Result | Notes |
|------|--------|-------|
| Init command | ✅ PASS | Schema + tracking table created |
| Migrations up | ✅ PASS | All 7 development migrations applied |
| Tracking populated | ✅ PASS | `schema_migrations` shows version 10 |
| Idempotency | ✅ PASS | Running again shows no pending |
| Staging environment | ✅ PASS | 5 migrations applied, tracked |
| Rollback | ⚠️ PARTIAL | golang-migrate requires sequential down files |

### Rollback Limitation
The current rollback test revealed that golang-migrate expects sequential migration files with down migrations. Our non-sequential versioning (1, 2, 3, 5, 8, 9, 10) can cause issues when rolling back because golang-migrate looks for the previous version's down file.

**Workaround**: The `m.Steps(1)` rollback works correctly for single-step rollbacks. Multi-step rollbacks may require manual intervention or fixing the down migration file naming.

---

## Migration File Guidelines

### File Naming
- Standard format: `NNNN_description.up.sql` and `NNNN_description.down.sql`
- Environment tags no longer needed (replaced by directory structure)
- Keep version numbers sequential within each environment

### Environment Tags
- **Optional but recommended** for documentation
- Not used by the orchestrator anymore
- Can be removed from SQL files or kept for human reference

Example:
```sql
-- Environment: all (optional - for documentation)
CREATE SCHEMA IF NOT EXISTS auth_service;
```

---

## Future Improvements

1. **Rollback enhancement**: Investigate golang-migrate's handling of non-sequential migrations
2. **Seed files**: Move to environment directories (User-service pending)
3. **Service template**: Update to reflect new structure
4. **Documentation**: Create comprehensive migration guidelines

---

## Timeline

| Phase | Status | Duration |
|-------|--------|----------|
| Phase 1 (Auth-Service) | ✅ Complete | 2 hours |
| Phase 2 (Objects-Service) | ⏳ Pending | 30 minutes |
| Phase 3 (User-Service) | ⏳ Pending | 30 minutes |
| Phase 4 (Documentation) | ⏳ Partial | 1 hour |
| **Remaining** | | ~1.5 hours |

---

## Success Criteria

- [x] Auth-service migrations work with environment directories
- [x] `schema_migrations` table populated correctly
- [x] Idempotent migration runs
- [x] Staging environment isolated from development
- [x] Binary renamed to `migrate-wrapper`
- [x] Documentation updated

- [ ] Objects-service migrations tested
- [ ] User-service migrations tested (including seed files)
- [ ] Rollback fully tested with non-sequential versions
- [ ] Service template updated
- [ ] Migration guidelines documented

---

**Last Updated:** 2026-04-11  
**Implementation Status:** Phase 1 COMPLETE (Auth-Service)