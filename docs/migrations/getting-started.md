# Getting Started with Database Migrations

## Prerequisites

Before you begin, ensure you have:

- ✅ PostgreSQL 15+ running
- ✅ Docker and Docker Compose installed
- ✅ Make utility available
- ✅ Services running (`make dev-detached`)
- ✅ Database initialized

## Your First Migration

### Step 1: Generate Migration

```bash
make db-migrate-generate NAME=add_user_preferences TYPE=table SERVICE_NAME=user-service
```

This creates migration files in the appropriate environment directory (default: development/).

### Step 2: Edit the Migration Files

Edit the generated files in `services/user-service/migrations/development/`:

**`00000X_add_user_preferences.up.sql`:**
```sql
-- Migration: 00000X_add_user_preferences
-- Description: Add user preferences table
-- Environment: development

-- Create user preferences table
CREATE TABLE IF NOT EXISTS user_service.user_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES user_service.users(id) ON DELETE CASCADE,
    theme VARCHAR(50) DEFAULT 'light',
    language VARCHAR(10) DEFAULT 'en',
    notifications_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add foreign key index
CREATE INDEX idx_user_preferences_user_id 
    ON user_service.user_preferences(user_id);
```

**`00000X_add_user_preferences.down.sql`:**
```sql
-- Rollback: 00000X_add_user_preferences
DROP TABLE IF EXISTS user_service.user_preferences;
```

### Step 3: Run Migrations

**Important:** Always run in sequence - init first, then up.

```bash
# Step 1: Initialize tracking (run once per service)
make db-migrate-init SERVICE_NAME=user-service

# Step 2: Apply all pending migrations
make db-migrate-up SERVICE_NAME=user-service
```

### Step 4: Verify

```bash
# Check migration status
make db-migrate-status SERVICE_NAME=user-service

# Verify tables created
docker exec service-boilerplate-postgres psql -U postgres -d service_db -c "\dt user_service.*"
```

## Migration File Structure

Each service has environment-specific directories:

```
services/{service}/migrations/
├── environments.json          # Configuration
├── development/              # Development migrations (sequential)
│   ├── 000001_*.up.sql
│   └── 000001_*.down.sql
├── staging/                  # Staging migrations (sequential)
└── production/               # Production migrations (sequential)
```

## Key Rules

1. **Sequential numbering**: Each environment must have sequential migration numbers (001, 002, 003...)
2. **Both directions**: Always create both `.up.sql` and `.down.sql` files
3. **Schema qualification**: Always use schema-qualified table names (e.g., `user_service.table_name`)
4. **Environment headers**: Add `-- Environment: development/staging` comments for dev-specific migrations
5. **Test thoroughly**: Always test both up and down migrations

## Migration Commands

```bash
# Initialize (run once)
make db-migrate-init SERVICE_NAME=user-service

# Apply migrations
make db-migrate-up SERVICE_NAME=user-service

# Rollback
make db-migrate-down SERVICE_NAME=user-service

# Check status
make db-migrate-status SERVICE_NAME=user-service

# Generate new migration
make db-migrate-generate NAME=add_feature TYPE=table SERVICE_NAME=user-service
```

## Troubleshooting

### Migration fails to apply
- Check SQL syntax
- Verify schema-qualified table names
- Ensure both up and down files exist

### Rollback fails
- Verify down migration is correct
- Check for data dependencies

### Service not detected
- Ensure `environments.json` exists in migrations directory
- Verify directory is named correctly (e.g., `development/`, not `dev/`)