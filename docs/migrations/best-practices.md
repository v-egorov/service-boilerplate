# Migration Best Practices

## Core Principles

### 1. Always Create Both Directions

**✅ Good Practice:**
```bash
# Generate migration with both up and down
make db-migration-generate NAME=add_user_settings TYPE=table

# Results in:
# - 000005_add_user_settings.up.sql
# - 000005_add_user_settings.down.sql
```

**❌ Anti-Pattern:**
```bash
# Only up migration - no rollback path
000005_add_user_settings.up.sql  # Only this file exists
```

### 2. Schema Qualification is Mandatory

**✅ Correct:**
```sql
-- Always use schema-qualified table names
SELECT * FROM user_service.users;
INSERT INTO user_service.user_settings VALUES (...);
UPDATE user_service.users SET name = 'John' WHERE id = 1;
```

**❌ Incorrect:**
```sql
-- Will fail or use wrong schema
SELECT * FROM users;  -- Assumes public schema
```

### 3. Test Migrations Before Committing

**Pre-Commit Checklist:**
```bash
# 1. Validate syntax and dependencies
make db-validate

# 2. Test migration application
make db-migrate-up

# 3. Test rollback
make db-migrate-down

# 4. Verify data integrity
make db-tables
make db-counts

# 5. Test with application
# Start services and verify functionality
```

## Migration Design Guidelines

### Table Design Best Practices

#### Primary Keys
```sql
-- ✅ Use SERIAL for auto-incrementing IDs
CREATE TABLE user_service.user_settings (
    id SERIAL PRIMARY KEY,
    -- other columns
);

-- ✅ Use UUID for distributed systems
CREATE TABLE user_service.api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- other columns
);
```

#### Foreign Keys
```sql
-- ✅ Define foreign keys with appropriate actions
CREATE TABLE user_service.user_profiles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES user_service.users(id) ON DELETE CASCADE,
    -- other columns
);

-- ✅ Use descriptive constraint names
ALTER TABLE user_service.user_profiles
ADD CONSTRAINT fk_user_profiles_user_id
FOREIGN KEY (user_id) REFERENCES user_service.users(id) ON DELETE CASCADE;
```

#### Indexes
```sql
-- ✅ Index foreign keys automatically
CREATE INDEX idx_user_profiles_user_id ON user_service.user_profiles(user_id);

-- ✅ Index columns used in WHERE clauses
CREATE INDEX idx_users_email ON user_service.users(email);
CREATE INDEX idx_users_created_at ON user_service.users(created_at DESC);

-- ✅ Consider composite indexes for complex queries
CREATE INDEX idx_users_name ON user_service.users(last_name, first_name);

-- ✅ Use partial indexes for filtered queries
CREATE INDEX idx_users_active ON user_service.users(created_at)
WHERE deleted_at IS NULL;
```

### Data Types and Constraints

```sql
-- ✅ Use appropriate data types
CREATE TABLE user_service.users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    is_active BOOLEAN DEFAULT true,
    settings JSONB,  -- For flexible configuration
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ✅ Add check constraints for data validation
ALTER TABLE user_service.users
ADD CONSTRAINT chk_email_format CHECK (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');

ALTER TABLE user_service.users
ADD CONSTRAINT chk_phone_format CHECK (phone IS NULL OR phone ~ '^\+?[0-9\s\-\(\)]+$');
```

## Performance Considerations

### Large Table Operations

```sql
-- ✅ Use CONCURRENTLY for index creation (non-blocking)
CREATE INDEX CONCURRENTLY idx_large_table_column ON user_service.large_table(column);

-- ✅ Batch large data updates
UPDATE user_service.large_table
SET status = 'processed'
WHERE id IN (
    SELECT id FROM user_service.large_table
    WHERE status = 'pending'
    LIMIT 1000
);

-- ✅ Use appropriate transaction sizes
-- For large operations, consider smaller batch sizes
```

### Migration Timing

```bash
# ✅ Schedule during low-traffic periods
# Check current load before migration
make db-status

# Monitor during migration
watch -n 5 'make db-status'

# ✅ Use maintenance windows for production
# Coordinate with operations team
```

## Error Handling and Rollback

### Comprehensive Rollback Planning

```sql
-- UP Migration: Add new column
ALTER TABLE user_service.users ADD COLUMN phone VARCHAR(20);

-- Populate with default values
UPDATE user_service.users SET phone = NULL WHERE phone IS NULL;

-- DOWN Migration: Remove column
ALTER TABLE user_service.users DROP COLUMN IF EXISTS phone;
```

### Handling Irreversible Changes

```sql
-- For irreversible data changes, document clearly
-- UP Migration: Normalize email addresses
UPDATE user_service.users SET email = LOWER(email);

-- DOWN Migration: Cannot reverse normalization
-- NOTE: Email normalization is irreversible
-- Consider backup strategies for rollbacks
```

### Transaction Management

```sql
-- ✅ Use transactions for multi-statement migrations
BEGIN;

-- Multiple operations
CREATE TABLE user_service.temp_table (...);
INSERT INTO user_service.temp_table SELECT ... FROM user_service.source_table;
UPDATE user_service.target_table SET ... FROM user_service.temp_table;
DROP TABLE user_service.temp_table;

COMMIT;
```

## Environment-Specific Practices

### Development Environment

```bash
# ✅ Safe to experiment and iterate quickly
make db-migration-generate NAME=experimental_feature TYPE=table

# ✅ Use development-specific data
make db-seed-enhanced ENV=development

# ✅ Test destructive operations safely
# (Development allows destructive operations)
```

### Staging Environment

```bash
# ✅ Mirror production constraints
# ✅ Test with production-like data volume
# ✅ Validate performance impact
# ✅ Test rollback procedures
```

### Production Environment

```bash
# ✅ Require explicit approval for migrations
# ✅ Always create backups before migration
make db-backup

# ✅ Schedule during maintenance windows
# ✅ Monitor migration execution closely
# ✅ Have rollback plan ready
# ✅ Test on staging first
```

## Documentation Standards

### Migration Documentation Template

```markdown
# Migration: 000005_add_user_settings

## Overview
**Type:** table
**Service:** user-service
**Schema:** user_service
**Risk Level:** Low
**Estimated Duration:** 30 seconds

## Description
Adds user settings table to store user preferences and configuration options.

## Changes Made

### Database Changes
- Creates `user_service.user_settings` table
- Adds foreign key to `user_service.users`
- Creates performance indexes
- Adds check constraints

### Affected Tables
- `user_service.user_settings` (created)
- `user_service.users` (referenced)

## Rollback Plan
Drops the user_settings table with CASCADE to remove all dependencies.

## Testing Checklist
- [x] Migration applies successfully
- [x] Foreign key constraints work
- [x] Indexes improve query performance
- [x] Rollback removes table cleanly
- [x] Application integration works

## Performance Impact
- **Read Queries:** Improved by indexes
- **Write Queries:** Minimal impact
- **Storage:** ~10% increase due to indexes

## Risk Assessment
- **Risk Level:** Low
- **Data Loss Risk:** None
- **Downtime Required:** No
- **Rollback Difficulty:** Easy
```

## Code Quality Standards

### SQL Style Guidelines

```sql
-- ✅ Consistent formatting
SELECT
    u.id,
    u.email,
    u.first_name,
    u.last_name,
    u.created_at
FROM user_service.users u
WHERE u.created_at > $1
ORDER BY u.created_at DESC
LIMIT $2;

-- ✅ Use meaningful aliases
SELECT
    user.id as user_id,
    user.email,
    profile.bio,
    profile.website
FROM user_service.users user
LEFT JOIN user_service.user_profiles profile ON user.id = profile.user_id;

-- ✅ Comment complex logic
-- Calculate user engagement score based on activity
SELECT
    user_id,
    COUNT(*) as total_actions,
    AVG(EXTRACT(EPOCH FROM (created_at - LAG(created_at) OVER (PARTITION BY user_id ORDER BY created_at)))) as avg_time_between_actions
FROM user_service.user_activities
WHERE created_at > CURRENT_TIMESTAMP - INTERVAL '30 days'
GROUP BY user_id;
```

### Naming Conventions

```sql
-- ✅ Tables: snake_case, plural
user_service.users
user_service.user_profiles
user_service.user_settings

-- ✅ Columns: snake_case
user_id, created_at, updated_at, is_active

-- ✅ Indexes: idx_table_column_pattern
idx_users_email, idx_users_created_at, idx_user_profiles_user_id

-- ✅ Constraints: descriptive names
fk_user_profiles_user_id, chk_email_format, pk_users
```

## Testing Strategies

### Unit Testing Migrations

```go
// migration_test.go
func TestMigration000005(t *testing.T) {
    // Setup test database
    db := setupTestDB()

    // Apply migration
    err := applyMigration(db, "000005_add_user_settings.up.sql")
    assert.NoError(t, err)

    // Verify table exists
    exists, err := tableExists(db, "user_service.user_settings")
    assert.NoError(t, err)
    assert.True(t, exists)

    // Verify constraints
    hasConstraint, err := hasForeignKey(db, "user_service.user_settings", "user_id")
    assert.NoError(t, err)
    assert.True(t, hasConstraint)

    // Test rollback
    err = applyMigration(db, "000005_add_user_settings.down.sql")
    assert.NoError(t, err)

    // Verify table removed
    exists, err = tableExists(db, "user_service.user_settings")
    assert.NoError(t, err)
    assert.False(t, exists)
}
```

### Integration Testing

```bash
# Test migration in isolated environment
docker-compose -f docker-compose.test.yml up -d

# Run migrations
make db-migrate-up

# Run application tests
npm test

# Verify data integrity
make db-counts

# Test rollback
make db-migrate-down

# Clean up
docker-compose -f docker-compose.test.yml down -v
```

## Security Considerations

### Data Protection

```sql
-- ✅ Use parameterized queries (handled by application)
-- ✅ Avoid storing sensitive data in migrations
-- ✅ Use appropriate column types for sensitive data

-- ✅ Encrypt sensitive columns if needed
CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE user_service.users
ADD COLUMN encrypted_ssn bytea;

-- ✅ Implement row-level security if needed
ALTER TABLE user_service.users ENABLE ROW LEVEL SECURITY;

CREATE POLICY users_own_data ON user_service.users
    FOR ALL USING (user_id = current_user_id());
```

### Access Control

```sql
-- ✅ Grant appropriate permissions
GRANT SELECT, INSERT, UPDATE ON user_service.users TO app_user;
GRANT USAGE ON SCHEMA user_service TO app_user;

-- ✅ Revoke dangerous permissions
REVOKE DROP ON user_service.users FROM app_user;
REVOKE TRUNCATE ON user_service.users FROM app_user;
```

## Monitoring and Maintenance

### Migration Monitoring

```bash
# Regular health checks
make db-status
make db-migrate-status

# Performance monitoring
make db-counts

# Storage monitoring
docker-compose exec postgres du -sh /var/lib/postgresql/data
```

### Maintenance Tasks

```bash
# Regular index maintenance
docker-compose exec postgres reindexdb -U postgres service_db

# Analyze tables for query optimization
docker-compose exec postgres psql -U postgres -d service_db -c "ANALYZE;"

# Vacuum for space reclamation
docker-compose exec postgres psql -U postgres -d service_db -c "VACUUM;"
```

## Common Pitfalls and Solutions

### 1. Missing Schema Qualification

**Problem:** Migration fails because table is not found
**Solution:** Always use `user_service.table_name`

### 2. Foreign Key Violations

**Problem:** Cannot drop table due to dependencies
**Solution:** Use `CASCADE` appropriately or drop dependencies first

### 3. Long-Running Migrations

**Problem:** Migration blocks database for too long
**Solution:** Break into smaller migrations or use `CONCURRENTLY`

### 4. Data Loss on Rollback

**Problem:** Rollback cannot restore deleted data
**Solution:** Always backup before destructive operations

### 5. Environment-Specific Failures

**Problem:** Migration works in dev but fails in production
**Solution:** Test in staging environment first

## Continuous Improvement

### Code Reviews for Migrations

**Review Checklist:**
- [ ] Both up and down migrations present
- [ ] Schema qualification used consistently
- [ ] Appropriate indexes added
- [ ] Foreign keys defined correctly
- [ ] Documentation complete
- [ ] Rollback tested
- [ ] Performance impact assessed

### Migration Metrics

```sql
-- Track migration performance
CREATE TABLE user_service.migration_metrics (
    migration_id VARCHAR(20) PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    execution_time_ms INTEGER,
    success BOOLEAN,
    environment VARCHAR(20),
    applied_by VARCHAR(100)
);
```

### Regular Audits

```bash
# Monthly migration audit
make db-validate
make db-migration-deps
make db-backup

# Check for unused indexes
make db-connect
SELECT * FROM pg_stat_user_indexes WHERE idx_scan = 0;
```

## Summary

Following these best practices ensures:

- **Reliability**: Migrations work consistently across environments
- **Maintainability**: Easy to understand and modify
- **Performance**: Optimized for production workloads
- **Safety**: Minimal risk of data loss or downtime
- **Scalability**: Support for growing application needs

Remember: **Migrations are forever** - design them carefully and test thoroughly!