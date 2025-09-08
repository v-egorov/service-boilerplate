# Getting Started with Database Migrations

## Prerequisites

Before you begin, ensure you have:

- ✅ PostgreSQL 15+ running
- ✅ Docker and Docker Compose installed
- ✅ Make utility available
- ✅ Services running (`make up`)
- ✅ Database initialized (`make db-setup`)

## Your First Migration

### Step 1: Generate Migration

Let's create a migration to add user preferences:

```bash
make db-migration-generate NAME=add_user_preferences TYPE=table
```

This creates:
- `services/user-service/migrations/000005_add_user_preferences.up.sql`
- `services/user-service/migrations/000005_add_user_preferences.down.sql`
- `services/user-service/migrations/docs/migration_000005.md`
- Updates `dependencies.json`

### Step 2: Edit the Migration Files

**Edit `000005_add_user_preferences.up.sql`:**

```sql
-- Migration: 000005_add_user_preferences
-- Description: Add user preferences table
-- Created: Auto-generated

-- Create user preferences table
CREATE TABLE IF NOT EXISTS user_service.user_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES user_service.users(id) ON DELETE CASCADE,
    theme VARCHAR(20) DEFAULT 'light',
    language VARCHAR(10) DEFAULT 'en',
    notifications_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes
CREATE INDEX IF NOT EXISTS idx_user_preferences_user_id ON user_service.user_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_user_preferences_theme ON user_service.user_preferences(theme);

-- Add comments
COMMENT ON TABLE user_service.user_preferences IS 'User preference settings';
COMMENT ON COLUMN user_service.user_preferences.theme IS 'UI theme preference (light/dark)';
```

**Edit `000005_add_user_preferences.down.sql`:**

```sql
-- Migration Rollback: 000005_add_user_preferences
-- Description: Remove user preferences table
-- Created: Auto-generated

-- Drop table (with CASCADE to remove dependencies)
DROP TABLE IF EXISTS user_service.user_preferences CASCADE;
```

### Step 3: Update Documentation

Edit `docs/migration_000005.md`:

```markdown
# Migration: 000005_add_user_preferences

## Overview
**Type:** table
**Service:** user-service
**Schema:** user_service
**Risk Level:** Low

## Description
Adds user preferences table to store user-specific settings like theme, language, and notification preferences.

## Changes Made
- Creates `user_service.user_preferences` table
- Adds foreign key to `user_service.users`
- Creates performance indexes
- Adds table and column comments

## Rollback Plan
Drops the user_preferences table with CASCADE to remove all dependencies.

## Testing
- [x] Migration applies successfully
- [x] Foreign key constraints work
- [x] Indexes improve query performance
- [x] Rollback removes table cleanly
```

### Step 4: Validate Migration

```bash
# Validate syntax and dependencies
make db-validate

# Check migration dependencies
make db-migration-deps
```

### Step 5: Test Migration

```bash
# Create backup before testing
make db-backup

# Apply migration
make db-migrate-up

# Check status
make db-migrate-status

# Verify table creation
make db-tables
```

### Step 6: Test Rollback

```bash
# Test rollback
make db-migrate-down

# Verify table removal
make db-tables

# Re-apply migration
make db-migrate-up
```

## Understanding the Structure

### Migration Files

Each migration consists of:
- **`.up.sql`**: SQL to apply the changes
- **`.down.sql`**: SQL to rollback the changes
- **Documentation**: Markdown file explaining the migration

### Naming Convention

```
{NNNN}_{descriptive_name}.{up|down}.sql

Examples:
- 000001_initial.up.sql
- 000002_add_user_profiles.up.sql
- 000003_dev_test_data.up.sql
```

### Schema Qualification

All table references must include the schema:

```sql
-- ✅ Correct
SELECT * FROM user_service.users;

-- ❌ Incorrect
SELECT * FROM users;
```

## Environment-Specific Migrations

### Development Environment

```bash
# Generate development-specific migration
make db-migration-generate NAME=add_dev_test_data TYPE=data

# This creates files in development/ subdirectory
# Only runs in development environment
```

### Production Considerations

For production migrations:
1. Always create backups first
2. Test on staging environment
3. Schedule during maintenance windows
4. Have rollback plan ready
5. Monitor execution time

## Common Commands

```bash
# Migration lifecycle
make db-migration-generate NAME=feature TYPE=table  # Create
make db-validate                                   # Validate
make db-migrate-up                                 # Apply
make db-migrate-status                             # Check status
make db-migrate-down                               # Rollback

# Data management
make db-seed                                       # Basic seeding
make db-seed-enhanced ENV=development             # Environment seeding
make db-clean                                      # Clean data

# Utilities
make db-backup                                     # Create backup
make db-tables                                     # List tables
make db-counts                                     # Show statistics
```

## Next Steps

Now that you understand the basics:

1. **[Migration Types](./migration-types.md)**: Learn about different migration categories
2. **[Advanced Features](./advanced-features.md)**: Explore dependencies and environments
3. **[Best Practices](./best-practices.md)**: Follow recommended guidelines
4. **[Examples](./examples.md)**: See real-world migration examples

## Troubleshooting

### Migration Fails to Apply

```bash
# Check migration status
make db-migrate-status

# Validate migration files
make db-validate

# Check database logs
make logs
```

### Rollback Issues

```bash
# Force rollback to specific version
make db-migrate-goto VERSION=000004

# Check current migration state
make db-migrate-status
```

### Permission Issues

```bash
# Ensure database user has proper permissions
make db-connect

# Check user permissions in psql
\du
```

## Getting Help

- 📖 **[Migration Types](./migration-types.md)** for different migration types
- 🔧 **[Troubleshooting](./troubleshooting.md)** for common issues
- 💡 **[Best Practices](./best-practices.md)** for guidelines
- 📚 **[API Reference](./api-reference.md)** for complete commands

---

**🎉 Congratulations!** You've successfully created and applied your first migration. The migration system is now ready for your development workflow.