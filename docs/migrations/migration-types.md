# Migration Types and Templates

## Overview

The migration system supports different types of database changes, each with specific templates and best practices. Understanding these types helps you choose the right approach for your database changes.

## Migration Types

### 1. Table Migrations (`TYPE=table`)

**Purpose**: Create, modify, or drop database tables

**Use Cases**:
- Adding new entities (users, products, orders)
- Modifying existing table structures
- Creating junction tables for many-to-many relationships

**Template Structure**:

```sql
-- UP Migration
CREATE TABLE IF NOT EXISTS user_service.user_settings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES user_service.users(id),
    setting_key VARCHAR(100) NOT NULL,
    setting_value TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id, setting_key)
);

CREATE INDEX IF NOT EXISTS idx_user_settings_user_id ON user_service.user_settings(user_id);
CREATE INDEX IF NOT EXISTS idx_user_settings_key ON user_service.user_settings(setting_key);

COMMENT ON TABLE user_service.user_settings IS 'User-specific application settings';
```

```sql
-- DOWN Migration
DROP TABLE IF EXISTS user_service.user_settings CASCADE;
```

**Best Practices**:
- Always include primary keys
- Add appropriate indexes for performance
- Use foreign key constraints for data integrity
- Add comments for documentation
- Consider CASCADE behavior for foreign keys

### 2. Index Migrations (`TYPE=index`)

**Purpose**: Add or remove database indexes for performance optimization

**Use Cases**:
- Optimizing slow queries
- Adding indexes for new query patterns
- Removing unused indexes to save space

**Template Structure**:

```sql
-- UP Migration
-- Add indexes for user search optimization
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email_search ON user_service.users(email);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_name_search ON user_service.users(first_name, last_name);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_created_at_desc ON user_service.users(created_at DESC);

-- Add partial index for active users
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_active ON user_service.users(created_at)
WHERE created_at > CURRENT_TIMESTAMP - INTERVAL '1 year';
```

```sql
-- DOWN Migration
DROP INDEX CONCURRENTLY IF EXISTS user_service.idx_users_email_search;
DROP INDEX CONCURRENTLY IF EXISTS user_service.idx_users_name_search;
DROP INDEX CONCURRENTLY IF EXISTS user_service.idx_users_created_at_desc;
DROP INDEX CONCURRENTLY IF EXISTS user_service.idx_users_active;
```

**Best Practices**:
- Use `CONCURRENTLY` to avoid blocking writes
- Test index impact on write performance
- Monitor index usage with `pg_stat_user_indexes`
- Consider partial indexes for filtered queries
- Drop unused indexes to reduce maintenance overhead

### 3. Data Migrations (`TYPE=data`)

**Purpose**: Modify existing data in the database

**Use Cases**:
- Updating existing records with new values
- Data cleanup and normalization
- Populating new columns with default values
- Migrating data between table structures

**Template Structure**:

```sql
-- UP Migration
-- Update user records to set default theme
UPDATE user_service.users
SET theme = 'light'
WHERE theme IS NULL;

-- Normalize email addresses to lowercase
UPDATE user_service.users
SET email = LOWER(email)
WHERE email != LOWER(email);

-- Populate created_at for legacy records
UPDATE user_service.users
SET created_at = '2024-01-01 00:00:00+00'::timestamptz
WHERE created_at IS NULL;
```

```sql
-- DOWN Migration
-- Note: Data rollbacks are often complex or impossible
-- Consider backup strategies for data migrations

-- Revert theme changes (if original values were preserved)
UPDATE user_service.users
SET theme = NULL
WHERE theme = 'light'
  AND created_at > '2024-09-07'; -- Only revert recently changed records

-- Note: Email normalization and created_at population may not be reversible
```

**Best Practices**:
- Test on subset of data first
- Include WHERE clauses to limit impact
- Consider transaction size for large datasets
- Document irreversible changes
- Have backup strategy ready

### 4. Schema Migrations (`TYPE=schema`)

**Purpose**: Modify database schema structure (constraints, defaults, etc.)

**Use Cases**:
- Adding/modifying constraints
- Changing column defaults
- Modifying column types (with caution)
- Adding check constraints
- Schema-level optimizations

**Template Structure**:

```sql
-- UP Migration
-- Add check constraint for email format
ALTER TABLE user_service.users
ADD CONSTRAINT chk_users_email_format
CHECK (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');

-- Add default value for new column
ALTER TABLE user_service.users
ALTER COLUMN theme SET DEFAULT 'light';

-- Add NOT NULL constraint (after populating existing data)
ALTER TABLE user_service.users
ALTER COLUMN email SET NOT NULL;
```

```sql
-- DOWN Migration
-- Remove constraints (be careful with dependencies)
ALTER TABLE user_service.users
DROP CONSTRAINT IF EXISTS chk_users_email_format;

-- Remove default value
ALTER TABLE user_service.users
ALTER COLUMN theme DROP DEFAULT;

-- Remove NOT NULL constraint
ALTER TABLE user_service.users
ALTER COLUMN email DROP NOT NULL;
```

**Best Practices**:
- Test constraint impact on existing data
- Use `NOT VALID` initially for check constraints if needed
- Validate constraints don't break existing data
- Consider performance impact of constraints

## Advanced Migration Patterns

### Composite Migrations

Sometimes a single migration needs multiple types of changes:

```sql
-- UP Migration: Add user notifications feature
-- 1. Create table (table migration)
CREATE TABLE user_service.user_notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES user_service.users(id),
    type VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Add indexes (index migration)
CREATE INDEX idx_user_notifications_user_id ON user_service.user_notifications(user_id);
CREATE INDEX idx_user_notifications_created_at ON user_service.user_notifications(created_at);

-- 3. Add constraint (schema migration)
ALTER TABLE user_service.user_notifications
ADD CONSTRAINT chk_notification_type
CHECK (type IN ('email', 'sms', 'push'));

-- 4. Populate initial data (data migration)
INSERT INTO user_service.user_notifications (user_id, type, message)
SELECT id, 'email', 'Welcome to our platform!'
FROM user_service.users
WHERE created_at > CURRENT_TIMESTAMP - INTERVAL '1 day';
```

### Environment-Specific Migrations

```sql
-- Development: Add test data
INSERT INTO user_service.users (email, first_name, last_name)
VALUES ('test@example.com', 'Test', 'User');

-- Staging: Add performance test data
-- (Larger dataset for performance testing)

-- Production: Add only essential data
-- (Minimal, verified data only)
```

## Risk Assessment Matrix

| Migration Type | Risk Level | Rollback Difficulty | Testing Required |
|----------------|------------|-------------------|------------------|
| Table Creation | Low | Easy | Basic validation |
| Index Addition | Low | Easy | Performance testing |
| Column Addition | Medium | Medium | Data population testing |
| Constraint Addition | Medium | Medium | Data validation testing |
| Data Migration | High | Hard | Comprehensive testing |
| Column Type Change | High | Very Hard | Full system testing |
| Table Drop | Critical | Impossible | Extensive testing |

## Choosing the Right Type

### Quick Decision Guide

**For new features:**
- Need new data storage? → `table`
- Optimizing existing queries? → `index`
- Adding business rules? → `schema`
- Populating new fields? → `data`

**For refactoring:**
- Changing data structure? → `table` + `data`
- Improving performance? → `index`
- Adding validation? → `schema`

**For maintenance:**
- Cleaning up data? → `data`
- Removing unused structures? → Check rollback impact
- Performance tuning? → `index`

## Template Customization

You can create custom templates for your specific needs:

```bash
# Create custom template directory
mkdir -p services/user-service/migrations/templates

# Add custom table template
cat > services/user-service/migrations/templates/audit_table.sql << 'EOF'
-- Audit table template
CREATE TABLE IF NOT EXISTS user_service.{table_name}_audit (
    id SERIAL PRIMARY KEY,
    {table_name}_id INTEGER NOT NULL,
    action VARCHAR(10) NOT NULL, -- INSERT, UPDATE, DELETE
    old_values JSONB,
    new_values JSONB,
    changed_by INTEGER,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_{table_name}_audit_{table_name}_id ON user_service.{table_name}_audit({table_name}_id);
CREATE INDEX idx_{table_name}_audit_changed_at ON user_service.{table_name}_audit(changed_at);
EOF
```

## Validation Checklist

Before applying migrations:

- [ ] **Syntax Check**: Valid SQL syntax
- [ ] **Schema Qualification**: All tables use `user_service.` prefix
- [ ] **Dependencies**: Required migrations applied first
- [ ] **Rollback Plan**: Down migration exists and is tested
- [ ] **Data Safety**: No data loss for existing records
- [ ] **Performance Impact**: Indexes and constraints won't slow down critical queries
- [ ] **Environment Testing**: Tested in development environment
- [ ] **Documentation**: Migration purpose and impact documented

## Next Steps

- **[Advanced Features](./advanced-features.md)**: Learn about dependencies and environments
- **[Best Practices](./best-practices.md)**: Follow recommended guidelines
- **[Examples](./examples.md)**: See real-world migration examples
- **[Troubleshooting](./troubleshooting.md)**: Handle common issues