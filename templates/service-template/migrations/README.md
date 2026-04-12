# SERVICE_NAME Migrations

This directory contains database migrations for the SERVICE_NAME service.

## Structure

```
migrations/
├── environments.json          # Environment configuration
├── development/              # Development migrations (2 files)
│   ├── 000001_initial.up.sql
│   ├── 000001_initial.down.sql
│   ├── 000002_dev_test_data.up.sql
│   └── 000002_dev_test_data.down.sql
├── staging/                  # Staging migrations (1 file)
│   ├── 000001_initial.up.sql
│   └── 000001_initial.down.sql
├── production/               # Production migrations (1 file)
│   ├── 000001_initial.up.sql
│   └── 000001_initial.down.sql
└── docs/                     # Migration documentation
```

## Running Migrations

**Important:** Always run in sequence - init must run once before up.

```bash
# Step 1: Initialize tracking (run once per service)
make db-migrate-init SERVICE_NAME=SERVICE_NAME

# Step 2: Apply all pending migrations
make db-migrate-up SERVICE_NAME=SERVICE_NAME

# Step 3: Check migration status
make db-migrate-status SERVICE_NAME=SERVICE_NAME

# Step 4: Rollback if needed
make db-migrate-down SERVICE_NAME=SERVICE_NAME
```

## Environment Differences

- **Development**: 2 migrations (includes test data)
- **Staging**: 1 migration (excludes dev-only test data)
- **Production**: 1 migration (excludes dev-only test data)

## Schema Information

This service uses the `SCHEMA_NAME` database schema. All tables are created within this schema to ensure proper isolation between services.

## Migration Guidelines

1. Each migration should be numbered sequentially within its environment
2. Include both up and down migrations
3. Always use schema-qualified table names (e.g., `SCHEMA_NAME.table_name`)
4. Test migrations thoroughly before committing
5. Development-specific migrations should have `-- Environment: development` in the header
6. Environment-specific directories contain different migration counts - do not copy between environments without re-numbering