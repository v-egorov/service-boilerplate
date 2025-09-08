# Migration Troubleshooting Guide

## Common Issues and Solutions

### 1. Migration Fails to Apply

#### Problem
```bash
make db-migrate-up
# Error: migration did not apply cleanly
```

#### Solutions

**Check Migration Status:**
```bash
make db-migrate-status
# Look for "dirty" state indicating failed migration
```

**Validate Migration Files:**
```bash
make db-validate
# Check for syntax errors or missing files
```

**Check Database Logs:**
```bash
make logs
# Look for PostgreSQL error messages
```

**Manual Application:**
```bash
# Connect to database
make db-connect

# Check current schema
\dt user_service.*

# Apply migration manually
\i services/user-service/migrations/000001_initial.up.sql
```

#### Prevention
- Always run `make db-validate` before applying migrations
- Test migrations in development environment first
- Check database logs for detailed error messages

---

### 2. Rollback Fails

#### Problem
```bash
make db-migrate-down
# Error: rollback did not apply cleanly
```

#### Solutions

**Check Down Migration:**
```sql
-- Verify down migration exists and is correct
-- Common issues:
-- 1. Missing down migration file
-- 2. Incorrect SQL syntax in down migration
-- 3. Dependencies not handled properly
```

**Force Rollback:**
```bash
# Go to specific version
make db-migrate-goto VERSION=000002

# Or reset and reapply
make db-fresh
make db-setup
```

**Manual Rollback:**
```bash
make db-connect
\i services/user-service/migrations/000003_feature.down.sql
```

#### Prevention
- Always create both up and down migrations
- Test rollbacks in development
- Document any irreversible operations

---

### 3. Schema Qualification Errors

#### Problem
```bash
ERROR: relation "users" does not exist
```

#### Solutions

**Check Schema Usage:**
```sql
-- Incorrect
SELECT * FROM users;

-- Correct
SELECT * FROM user_service.users;
```

**Update Repository Code:**
```go
// In user_repository.go
query := `SELECT * FROM user_service.users WHERE id = $1`
```

**Fix Migration Files:**
```sql
-- Ensure all table references are schema-qualified
CREATE TABLE user_service.new_table (...);
INSERT INTO user_service.existing_table VALUES (...);
```

#### Prevention
- Use find-and-replace to add schema prefixes
- Run validation to catch missing qualifications
- Set up code templates with schema prefixes

---

### 4. Dependency Conflicts

#### Problem
```bash
make db-validate
# ERROR: Migration 000003 depends on 000002, but 000002 not found
```

#### Solutions

**Check Dependencies File:**
```json
// services/user-service/migrations/dependencies.json
{
  "migrations": {
    "000003": {
      "depends_on": ["000002"],  // Ensure 000002 exists
      "description": "Feature migration"
    }
  }
}
```

**Verify Migration Files:**
```bash
ls services/user-service/migrations/
# Ensure all referenced migrations exist
```

**Update Dependencies:**
```bash
# Regenerate dependencies
make db-migration-generate NAME=new_feature TYPE=table
# This updates dependencies.json automatically
```

#### Prevention
- Keep dependencies.json synchronized with migration files
- Use automated generation to maintain dependencies
- Review dependencies during code reviews

---

### 5. Performance Issues

#### Problem
```bash
time make db-migrate-up
# Takes longer than expected
```

#### Solutions

**Add CONCURRENTLY to Indexes:**
```sql
-- Instead of:
CREATE INDEX idx_table_column ON user_service.table(column);

-- Use:
CREATE INDEX CONCURRENTLY idx_table_column ON user_service.table(column);
```

**Batch Large Operations:**
```sql
-- Instead of updating all records at once:
UPDATE user_service.large_table SET status = 'processed';

-- Use batching:
UPDATE user_service.large_table
SET status = 'processed'
WHERE id IN (
    SELECT id FROM user_service.large_table
    WHERE status = 'pending'
    LIMIT 1000
);
```

**Monitor Progress:**
```bash
# Watch migration progress
watch -n 5 'make db-migrate-status'

# Check database locks
make db-connect
SELECT * FROM pg_locks WHERE NOT granted;
```

#### Prevention
- Test migrations with production-sized data
- Use `CONCURRENTLY` for index operations
- Schedule large migrations during maintenance windows

---

### 6. Environment-Specific Issues

#### Problem
```bash
# Works in development, fails in staging
make db-migrate-up
```

#### Solutions

**Check Environment Configuration:**
```json
// services/user-service/migrations/environments.json
{
  "environments": {
    "staging": {
      "migrations": ["000005_staging_only.up.sql"],
      "config": {
        "allow_destructive_operations": false
      }
    }
  }
}
```

**Verify Environment Variables:**
```bash
echo $MIGRATION_ENV
# Should match environment name
```

**Test in Staging:**
```bash
# Use staging-specific seeding
make db-seed-enhanced ENV=staging

# Test migrations in staging
make db-migrate-up
```

#### Prevention
- Test migrations in staging before production
- Use environment-specific configurations
- Document environment differences

---

### 7. Data Integrity Issues

#### Problem
```bash
# Migration succeeds but data is corrupted
make db-counts
# Shows unexpected results
```

#### Solutions

**Check Data Before Migration:**
```sql
-- Count records before migration
SELECT COUNT(*) FROM user_service.users;

-- Check for NULL values
SELECT COUNT(*) FROM user_service.users WHERE email IS NULL;
```

**Validate After Migration:**
```sql
-- Verify constraints
SELECT * FROM user_service.users WHERE LENGTH(email) < 5;

-- Check foreign keys
SELECT * FROM user_service.user_profiles p
LEFT JOIN user_service.users u ON p.user_id = u.id
WHERE u.id IS NULL;
```

**Restore from Backup:**
```bash
# Find latest backup
ls -la backup_*.sql | tail -1

# Restore
make db-restore FILE=backup_file.sql
```

#### Prevention
- Always backup before migrations
- Validate data integrity after migrations
- Test with realistic data volumes

---

### 8. Connection Issues

#### Problem
```bash
make db-migrate-up
# ERROR: connection refused
```

#### Solutions

**Check Database Status:**
```bash
make db-status
# Should show database is accessible
```

**Verify Services:**
```bash
make ps
# Should show postgres container running
```

**Restart Services:**
```bash
make down
make up
```

**Check Connection String:**
```bash
# Verify environment variables
echo $DATABASE_HOST
echo $DATABASE_PORT
echo $DATABASE_NAME
```

#### Prevention
- Ensure services are running before migrations
- Check network connectivity
- Verify database credentials

---

### 9. Lock Conflicts

#### Problem
```bash
make db-migrate-up
# ERROR: relation is being used by active queries
```

#### Solutions

**Check Active Connections:**
```sql
make db-connect
SELECT * FROM pg_stat_activity WHERE state != 'idle';
```

**Terminate Blocking Queries:**
```sql
-- Be careful with this in production!
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE state = 'active' AND pid != pg_backend_pid();
```

**Use CONCURRENTLY:**
```sql
-- For index creation
CREATE INDEX CONCURRENTLY idx_table_column ON user_service.table(column);
```

**Schedule During Low Traffic:**
```bash
# Check current activity
make db-status

# Wait for low traffic period
```

#### Prevention
- Schedule migrations during maintenance windows
- Use `CONCURRENTLY` for index operations
- Monitor database activity before migrations

---

### 10. Disk Space Issues

#### Problem
```bash
make db-migrate-up
# ERROR: no space left on device
```

#### Solutions

**Check Disk Usage:**
```bash
df -h
docker system df
```

**Clean Up Old Backups:**
```bash
# Remove old backups (keep last 5)
ls -t backup_*.sql | tail -n +6 | xargs rm
```

**Free Docker Space:**
```bash
docker system prune -a
```

**Monitor During Migration:**
```bash
# Watch disk usage
watch -n 10 'df -h'
```

#### Prevention
- Monitor disk space before migrations
- Clean up old backups regularly
- Plan for space requirements of large migrations

---

## Advanced Troubleshooting

### Database Recovery

```bash
# Complete recovery procedure
echo "🚨 Starting database recovery"

# 1. Stop all services
make down

# 2. Find latest backup
LATEST_BACKUP=$(ls -t backup_*.sql | head -1)
echo "Using backup: $LATEST_BACKUP"

# 3. Reset database
make db-fresh

# 4. Restore from backup
make db-restore FILE="$LATEST_BACKUP"

# 5. Restart services
make up

# 6. Verify recovery
make db-status
make db-counts
```

### Migration Debugging

```bash
# Enable verbose logging
export MIGRATION_DEBUG=true

# Run with detailed output
make db-migrate-up 2>&1 | tee migration.log

# Analyze log
grep -i error migration.log
grep -i warning migration.log
```

### Performance Analysis

```sql
-- Check slow queries during migration
SELECT
    query,
    total_exec_time,
    mean_exec_time,
    calls
FROM pg_stat_statements
ORDER BY total_exec_time DESC
LIMIT 10;

-- Check table bloat
SELECT
    schemaname,
    tablename,
    n_dead_tup,
    n_live_tup
FROM pg_stat_user_tables
WHERE schemaname = 'user_service'
ORDER BY n_dead_tup DESC;
```

## Diagnostic Commands

### Quick Health Check
```bash
# Run all diagnostic commands
echo "=== Database Status ==="
make db-status

echo "=== Migration Status ==="
make db-migrate-status

echo "=== Validation Results ==="
make db-validate

echo "=== Table Counts ==="
make db-counts

echo "=== Connection Test ==="
make db-connect -c "SELECT version();"
```

### Emergency Commands
```bash
# Force migration reset (CAUTION!)
make db-fresh

# Emergency rollback
make db-migrate-goto VERSION=000001

# Database restart
make down
make up
```

## Prevention Strategies

### Pre-Migration Checklist
- [ ] Run `make db-validate`
- [ ] Create backup with `make db-backup`
- [ ] Check disk space availability
- [ ] Review migration for destructive operations
- [ ] Test in staging environment first
- [ ] Schedule during low-traffic period
- [ ] Have rollback plan ready

### Monitoring Setup
```bash
# Set up monitoring
watch -n 30 'make db-status'

# Log migration progress
make db-migrate-up 2>&1 | tee "migration_$(date +%Y%m%d_%H%M%S).log"
```

### Automated Testing
```bash
# Test migration in isolated environment
docker-compose -f docker-compose.test.yml up -d
make db-migrate-up
make db-migrate-down
docker-compose -f docker-compose.test.yml down -v
```

## Getting Help

### Documentation Resources
- **[Getting Started](../getting-started.md)**: Basic migration workflow
- **[Best Practices](../best-practices.md)**: Guidelines and recommendations
- **[API Reference](../api-reference.md)**: Complete command reference
- **[Examples](../examples.md)**: Real-world migration examples

### Community Support
- Check existing issues in the repository
- Review migration logs for error patterns
- Test in isolated environment before production
- Document solutions for future reference

### Emergency Contacts
- Database administrator
- DevOps team
- System monitoring alerts
- Backup recovery procedures

---

**Remember**: When in doubt, **create a backup first** and test thoroughly in a non-production environment!