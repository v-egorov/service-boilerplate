# objects-service Migrations

This directory contains database migrations for the objects-service service.

## Structure

- `000001_initial.up.sql` / `000001_initial.down.sql` - Initial schema creation
- `development/` - Development environment migrations (6 migrations including test data)
- `staging/` - Staging environment migrations (5 migrations)
- `production/` - Production environment migrations (5 migrations)
- `docs/` - Migration documentation
- `environments.json` - Environment configuration

## Running Migrations

Use the provided scripts to manage migrations:

```bash
# Apply all pending migrations
make db-migrate-up objects-service=objects-service

# Rollback last migration
make db-migrate-down objects-service=objects-service

# Check migration status
make db-migrate-status objects-service=objects-service
```

## Development Environment

The development environment includes test data that can be loaded using:

```bash
make db-seed objects-service=objects-service ENV=development
```

## Schema Information

This service uses the `objects_service` database schema. All tables are created within this schema to ensure proper isolation between services.

## Migration Guidelines

1. Each migration should be numbered sequentially
2. Include both up and down migrations
3. Always use schema-qualified table names (e.g., `objects_service.table_name`)
4. Test migrations thoroughly before committing
5. Document complex migrations in the `docs/` directory
6. Update `dependencies.json` with new migration metadata