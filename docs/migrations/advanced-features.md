# Advanced Migration Features

## Service-Specific Migration Tracking

### Overview

Each service maintains its own migration tracking table to ensure complete isolation between services. This allows multiple services to coexist in the same database with independent migration histories.

### Migration Table Naming

Migration tables follow the pattern: `{service_name}.schema_migrations`

Examples:
- `auth_service.schema_migrations`
- `user_service.schema_migrations`
- `objects_service.schema_migrations`

### Benefits

- **Service Isolation**: Each service can migrate independently
- **Version Independence**: Services can have different migration versions
- **Rollback Safety**: Rolling back one service doesn't affect others
- **Parallel Development**: Multiple services can be developed simultaneously

### Database Architecture

```
service_db (PostgreSQL)
├── auth_service
│   ├── schema_migrations
│   └── tables
├── user_service
│   ├── schema_migrations
│   └── tables
└── objects_service
    ├── schema_migrations
    └── tables
```

## Environment-Specific Migrations

### Overview

Each service has separate migration directories for different environments. This allows:
- Development to include test data migrations
- Staging/Production to exclude dev-only migrations

### Directory Structure

```
services/{service}/migrations/
├── environments.json      # Configuration
├── development/           # Dev migrations (more files)
├── staging/              # Staging migrations
└── production/           # Production migrations
```

### Configuration

Each environment is configured in `environments.json`:

```json
{
  "environments": {
    "development": {
      "migrations": "development",
      "config": { "allow_destructive_operations": true }
    },
    "staging": {
      "migrations": "staging"
    },
    "production": {
      "migrations": "production"
    }
  }
}
```

### Sequential Numbering

Each environment directory must have **sequential** migration numbers:
- Development: 001, 002, 003... (e.g., 7 files)
- Staging: 001, 002, 003... (e.g., 5 files)
- Production: 001, 002, 003... (e.g., 5 files)

**Important**: Staging and production don't simply omit files - they need re-numbered migrations to maintain sequential order for golang-migrate.

## Best Practices

### 1. Always Use Both Directions
Every migration must have both `.up.sql` and `.down.sql` files.

### 2. Schema Qualification
Always use schema-qualified table names:
```sql
CREATE TABLE user_service.users (...);
NOT CREATE TABLE users (...);
```

### 3. Environment Markers
Add headers to indicate which environments a migration applies to:
```sql
-- Environment: development
-- Or --
-- Environment: all
```

### 4. Test Thoroughly
Always test both up and down migrations in development before committing.

### 5. Keep Development Clean
For development, it's safe to drop and recreate schemas:
```bash
docker exec service-boilerplate-postgres psql -U postgres -d service_db \
  -c "DROP SCHEMA user_service CASCADE; CREATE SCHEMA user_service;"
```

## Next Steps

See also:
- [Getting Started](./getting-started.md)
- [Troubleshooting](./troubleshooting.md)
- [Migration Types](./migration-types.md)