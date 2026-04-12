# Migration Guide

This document combines best practices and real-world examples for database migrations.

## Core Principles

### 1. Always Create Both Directions

**✅ Good Practice:**
```bash
# Generate migration with both up and down
make db-migrate-generate NAME=add_user_settings TYPE=table SERVICE_NAME=user-service

# Results in:
# - 00000X_add_user_settings.up.sql
# - 00000X_add_user_settings.down.sql
```

**❌ Anti-Pattern:**
```bash
# Only up migration - no rollback path
00000X_add_user_settings.up.sql  # Only this file exists
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
-- Missing schema prefix
SELECT * FROM users;  -- May query wrong schema!
```

### 3. Sequential Numbering

Each environment must have sequential migration numbers:

```bash
# Development: 001, 002, 003, 004, 005, 006, 007
# Staging:    001, 002, 003, 004, 005
# Production: 001, 002, 003, 004, 005
```

### 4. Environment Awareness

Add headers to indicate which environments a migration applies to:

```sql
-- Environment: development
-- Or --
-- Environment: all
-- Or --
-- Environment: development/staging
```

## Examples

### Example 1: Adding a New Table

**Scenario:** Add user profiles table with bio and preferences.

**Step 1: Generate Migration**
```bash
make db-migrate-generate NAME=add_user_profiles TYPE=table SERVICE_NAME=user-service
```

**Step 2: Edit Up Migration**
```sql
-- Migration: 000002_add_user_profiles
-- Description: Add user profiles table
-- Environment: all

CREATE TABLE IF NOT EXISTS user_service.user_profiles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES user_service.users(id) ON DELETE CASCADE,
    bio TEXT,
    avatar_url VARCHAR(500),
    website VARCHAR(255),
    location VARCHAR(100),
    timezone VARCHAR(50) DEFAULT 'UTC',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_profiles_user_id 
    ON user_service.user_profiles(user_id);
```

**Step 3: Edit Down Migration**
```sql
-- Rollback: 000002_add_user_profiles
DROP TABLE IF EXISTS user_service.user_profiling CASCADE;
```

### Example 2: Adding Dev-Only Test Data

**Scenario:** Add test users for development only.

```sql
-- Migration: 000003_dev_test_data
-- Description: Add test users for development
-- Environment: development

INSERT INTO user_service.users (email, password_hash, first_name, last_name, created_at, updated_at)
VALUES 
    ('dev.admin@example.com', '$2a$10$xxxxx', 'Dev', 'Admin', NOW(), NOW()),
    ('test.user@example.com', '$2a$10$xxxxx', 'Test', 'User', NOW(), NOW());
```

### Example 3: Adding an Index

**Scenario:** Add index for better query performance.

```sql
-- Migration: 000004_add_email_index
-- Description: Add index on users email column
-- Environment: all

CREATE INDEX CONCURRENTLY idx_users_email 
    ON user_service.users(email);

-- Down:
DROP INDEX IF EXISTS idx_users_email;
```

**Note:** Use `CONCURRENTLY` in production to avoid table locks.

### Example 4: Adding a Column

**Scenario:** Add phone number to users table.

```sql
-- Migration: 000005_add_phone
-- Description: Add phone column to users
-- Environment: all

ALTER TABLE user_service.users 
    ADD COLUMN phone VARCHAR(20);

ALTER TABLE user_service.users 
    ADD CONSTRAINT users_phone_check 
    CHECK (phone IS NULL OR phone ~ '^\+?[0-9]{10,20}$');

-- Down:
ALTER TABLE user_service.users 
    DROP COLUMN IF EXISTS phone;
```

### Example 5: Rolling Back Safely

**Before running down migration, check for dependencies:**

```sql
-- Check if other tables reference the table you're about to drop
SELECT 
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
    ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage ccu
    ON ccu.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_schema = 'user_service'
    AND ccu.table_name = 'users';
```

## Common Patterns

### Table Creation with Standard Columns

```sql
CREATE TABLE IF NOT EXISTS user_service.example (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add update trigger
CREATE OR REPLACE FUNCTION user_service.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_example_updated_at
    BEFORE UPDATE ON user_service.example
    FOR EACH ROW
    EXECUTE FUNCTION user_service.update_updated_at_column();
```

### Data Migration with Temporary Table

```sql
-- Migration: 00000X_migrate_data
-- Description: Migrate data from old column to new structure

-- Step 1: Add new columns
ALTER TABLE user_service.users ADD COLUMN phone VARCHAR(20);

-- Step 2: Copy data (if needed)
UPDATE user_service.users 
SET phone = contact_info->>'phone'
WHERE contact_info IS NOT NULL;

-- Step 3: Drop old column (in separate migration for safety)
-- ALTER TABLE user_service.users DROP COLUMN IF EXISTS contact_info;
```

## Performance Tips

### Index Creation

```sql
-- Good: Non-blocking index creation
CREATE INDEX CONCURRENTLY idx_table_column ON user_service.table(column);

-- Avoid: Blocks writes during index creation
CREATE INDEX idx_table_column ON user_service.table(column);
```

### Large Data Updates

```sql
-- Bad: May cause long lock
UPDATE user_service.large_table SET status = 'processed';

-- Good: Batch processing
UPDATE user_service.large_table
SET status = 'processed'
WHERE id IN (
    SELECT id FROM user_service.large_table
    WHERE status = 'pending'
    LIMIT 1000
);
```

## Testing Migrations

Always test in development:

```bash
# 1. Reset database
docker exec service-boilerplate-postgres psql -U postgres -d service_db \
  -c "DROP SCHEMA user_service CASCADE; CREATE SCHEMA user_service;"

# 2. Initialize and run migrations
make db-migrate-init SERVICE_NAME=user-service
make db-migrate-up SERVICE_NAME=user-service

# 3. Verify tables created
docker exec service-boilerplate-postgres psql -U postgres -d service_db \
  -c "\dt user_service.*"

# 4. Test rollback
make db-migrate-down SERVICE_NAME=user-service

# 5. Re-apply
make db-migrate-up SERVICE_NAME=user-service
```

## Summary

| Rule | Description |
|------|-------------|
| Both directions | Always create up and down |
| Schema qualification | Use `schema.table` |
| Sequential | Number sequentially per environment |
| Test thoroughly | Test up and down in dev |
| Use CONCURRENTLY | For index creation in production |