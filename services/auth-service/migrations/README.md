# auth-service Migrations

This directory contains database migrations for the auth-service service.

## Structure

```
migrations/
├── environments.json          # Environment configuration (directory per env)
├── development/              # Development migrations (7 files)
├── staging/                  # Staging migrations (5 files)
├── production/               # Production migrations (5 files)
└── docs/                     # Migration documentation
```

## Running Migrations

**Always run in sequence** - init must run once before up:

```bash
# Step 1: Initialize (run once per service)
make db-migrate-init SERVICE_NAME=auth-service

# Step 2: Apply all pending migrations
make db-migrate-up SERVICE_NAME=auth-service

# Step 3: Check status
make db-migrate-status SERVICE_NAME=auth-service

# Step 4: Rollback if needed
make db-migrate-down SERVICE_NAME=auth-service
```

## Environment Differences

- **Development**: 7 migrations (includes test data: roles, permissions seed)
- **Staging**: 5 migrations (excludes dev-only test data)
- **Production**: 5 migrations (excludes dev-only test data)

## Schema Information

This service uses the `auth_service` database schema. All tables are created within this schema to ensure proper isolation between services.

## Migration Guidelines

1. Each migration should be numbered sequentially within its environment
2. Include both up and down migrations
3. Always use schema-qualified table names (e.g., `auth_service.table_name`)
4. Test migrations thoroughly before committing
5. Development-specific migrations should have `-- Environment: development` in the header
6. Environment-specific directories contain different migration counts - do not copy between environments without re-numbering