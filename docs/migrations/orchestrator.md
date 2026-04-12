# Migration Wrapper (migrate-wrapper)

## Overview

The migration wrapper (`migrate-wrapper`) is a simple Go CLI wrapper around golang-migrate that handles environment-specific migrations and database schema isolation.

## What It Does

- Loads `environments.json` to determine which migration directory to use
- Executes golang-migrate CLI with proper database connection (using search_path for schema isolation)
- Provides simple commands: init, up, down, status, list

## What It Is NOT

This is **not** the enterprise migration orchestrator originally planned. The following features were **not implemented**:

- ❌ Dependency resolution between migrations
- ❌ Risk assessment for migrations
- ❌ Enhanced tracking with `migration_executions` table
- ❌ Approval workflows
- ❌ Backup integration
- ❌ Intelligent rollback with dependency checking

## Architecture

```
migration-orchestrator/
├── cmd/                    # CLI commands
│   ├── init.go            # Initialize tracking for new services
│   ├── up.go              # Run migrations up
│   ├── down.go            # Rollback migrations
│   ├── status.go          # Show migration status
│   └── list.go            # List all migrations
├── internal/
│   ├── orchestrator/      # Core logic
│   └── database/          # Database connection
└── pkg/
    ├── types/             # Data structures
    └── utils/             # Logging utilities
```

## Usage

### Initialize (run once per service)

```bash
make db-migrate-init SERVICE_NAME=user-service
```

### Apply Migrations

```bash
make db-migrate-up SERVICE_NAME=user-service
```

### Rollback

```bash
make db-migrate-down SERVICE_NAME=user-service
```

### Check Status

```bash
make db-migrate-status SERVICE_NAME=user-service
```

## Database Schema

Each service gets its own PostgreSQL schema with a `schema_migrations` table for tracking:

```sql
-- Example for user_service schema
CREATE SCHEMA user_service;

-- golang-migrate creates this automatically
CREATE TABLE user_service.schema_migrations (
    version bigint NOT NULL PRIMARY KEY,
    dirty boolean NOT NULL
);
```

## Configuration

Each service has an `environments.json` in its migrations directory:

```json
{
  "environments": {
    "development": {
      "migrations": "development"
    },
    "staging": {
      "migrations": "staging"
    },
    "production": {
      "migrations": "production"
    }
  },
  "current_environment": "development"
}
```

## Migration File Structure

Each environment directory contains sequential migration files:

```
services/user-service/migrations/
├── development/
│   ├── 000001_initial.up.sql
│   ├── 000001_initial.down.sql
│   ├── 000002_add_user_profiles.up.sql
│   └── ...
├── staging/
│   └── ...
└── production/
    └── ...
```

## Important Notes

1. Migrations must be numbered sequentially within each environment
2. Both `.up.sql` and `.down.sql` files are required
3. The wrapper uses `search_path` to ensure migrations affect the correct service schema

---

**Note**: This is a simplified implementation. For more advanced features like dependency resolution and risk assessment, consider extending this wrapper or using a dedicated migration tool.