# Migration Troubleshooting Guide

## Common Issues and Solutions

### 1. Migration Not Applied

#### Symptoms
- Migration files exist but tables not created
- `schema_migrations` shows lower version than expected

#### Diagnosis
```bash
# Check current migration version
make db-migrate-status SERVICE_NAME=user-service

# List migration files
ls services/user-service/migrations/development/
```

#### Solutions

**Verify sequence:**
- Ensure migration files are numbered sequentially (001, 002, 003...)
- Check both `.up.sql` and `.down.sql` exist for each migration

**Check JSON config:**
```bash
# Verify environments.json is valid
cat services/user-service/migrations/environments.json | jq .
```

**Check file permissions:**
```bash
ls -la services/user-service/migrations/development/
```

---

### 2. Rollback Fails

#### Symptoms
- `db-migrate-down` returns error
- Database state becomes inconsistent

#### Diagnosis
```bash
# Check current state
docker exec service-boilerplate-postgres psql -U postgres -d service_db \
  -c "SELECT * FROM user_service.schema_migrations;"
```

#### Solutions

**Verify down file exists:**
```bash
ls services/user-service/migrations/development/*_down.sql
```

**Check for data dependencies:**
```sql
-- Before dropping tables, check for dependencies
SELECT 
    tc.table_name, 
    kcu.column_name
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
    ON tc.constraint_name = kcu.constraint_name
WHERE tc.table_schema = 'user_service';
```

**Manual cleanup if needed:**
```bash
# Drop schema and reinitialize (dev only!)
docker exec service-boilerplate-postgres psql -U postgres -d service_db \
  -c "DROP SCHEMA user_service CASCADE; CREATE SCHEMA user_service;"

make db-migrate-init SERVICE_NAME=user-service
make db-migrate-up SERVICE_NAME=user-service
```

---

### 3. Migration Path Not Found

#### Symptoms
- Error: "migration path does not exist"
- Error: "file does not exist"

#### Diagnosis
```bash
# Verify migration directories exist
ls -d services/*/migrations/*/
```

#### Solutions

**Check environment directories:**
```
services/{service}/migrations/
├── development/    # Must exist
├── staging/        # Must exist (even if empty)
└── production/     # Must exist (even if empty)
```

**Verify migration directory matches config:**
```json
// environments.json should specify:
{
  "environments": {
    "development": {
      "migrations": "development"
    }
  }
}
```

---

### 4. Dirty Database State

#### Symptoms
- Migration version shows `dirty: true`
- Cannot apply or rollback migrations

#### Diagnosis
```bash
docker exec service-boilerplate-postgres psql -U postgres -d service_db \
  -c "SELECT * FROM user_service.schema_migrations;"
```

#### Solutions

**Fix dirty state (development only):**
```bash
# Option 1: Force version to known state
docker exec service-boilerplate-postgres psql -U postgres -d service_db \
  -c "UPDATE user_service.schema_migrations SET dirty = false;"

# Option 2: Drop and reinitialize
docker exec service-boilerplate-postgres psql -U postgres -d service_db \
  -c "DROP SCHEMA user_service CASCADE; CREATE SCHEMA user_service;"

make db-migrate-init SERVICE_NAME=user-service
make db-migrate-up SERVICE_NAME=user-service
```

---

### 5. Performance Issues

#### Symptoms
- Migration takes longer than expected

#### Solutions

**Use CONCURRENTLY for indexes:**
```sql
-- Good: creates index without locking writes
CREATE INDEX CONCURRENTLY idx_table_column 
    ON user_service.table(column);

-- Avoid: blocks writes during index creation
CREATE INDEX idx_table_column ON user_service.table(column);
```

**Batch large operations:**
```sql
-- Instead of:
UPDATE user_service.large_table SET status = 'processed';

-- Use batches:
UPDATE user_service.large_table
SET status = 'processed'
WHERE id IN (
    SELECT id FROM user_service.large_table
    WHERE status = 'pending'
    LIMIT 1000
);
```

---

### 6. Service Not Detected

#### Symptoms
- Service not included in migration targets

#### Diagnosis
```bash
# Check what services are detected
make -n db-migrate-up | grep SERVICES
```

#### Solutions

**Verify migration structure:**
```bash
# Must have:
services/{service}/migrations/
├── environments.json    # Required
├── development/         # Required
├── staging/            # Required
└── production/         # Required
```

**Validate JSON:**
```bash
# Check JSON is valid
cat services/user-service/migrations/environments.json | jq .
```

---

## Prevention Best Practices

1. **Always test both up and down** before committing
2. **Use sequential numbering** within each environment
3. **Add environment headers** to dev-only migrations
4. **Keep development clean** - drop and reinitialize often
5. **Use CONCURRENTLY** for index creation in production

## Getting Help

If issues persist:

1. Check migration logs: `tail -f docker/volumes/{service}/logs/{service}.log`
2. Verify database connectivity
3. Review SQL syntax in migration files
4. Check PostgreSQL version compatibility