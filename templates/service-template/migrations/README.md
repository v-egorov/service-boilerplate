# SERVICE_NAME Migrations

This directory contains database migrations for the SERVICE_NAME service.

## Structure

- `000001_initial.up.sql` / `000001_initial.down.sql` - Initial schema creation
- `development/` - Development-specific migrations (test data, etc.)
- `docs/` - Migration documentation
- `dependencies.json` - Migration dependencies and metadata
- `environments.json` - Environment-specific migration configurations

## Running Migrations

Use the provided scripts to manage migrations:

```bash
# Apply all pending migrations
make db-migrate-up SERVICE_NAME=SERVICE_NAME

# Rollback last migration
make db-migrate-down SERVICE_NAME=SERVICE_NAME

# Check migration status
make db-migrate-status SERVICE_NAME=SERVICE_NAME
```

## Development Environment

The development environment includes test data that can be loaded using:

```bash
make db-seed SERVICE_NAME=SERVICE_NAME ENV=development
```

## Schema Information

This service uses the `SCHEMA_NAME` database schema. All tables are created within this schema to ensure proper isolation between services.

## Migration Guidelines

1. Each migration should be numbered sequentially
2. Include both up and down migrations
3. Always use schema-qualified table names (e.g., `SCHEMA_NAME.table_name`)
4. Test migrations thoroughly before committing
5. Document complex migrations in the `docs/` directory
6. Update `dependencies.json` with new migration metadata