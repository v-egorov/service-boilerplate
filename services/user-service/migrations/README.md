# Database Migrations - User Service

## Overview

This directory contains database migrations for the user service using a schema-per-service architecture. All user-related tables are created in the `user_service` schema.

## Directory Structure

```
migrations/
├── README.md                          # This file
├── dependencies.json                  # Migration dependencies and metadata
├── environments.json                  # Environment-specific configuration
├── 000001_initial.up.sql             # Initial schema creation
├── 000001_initial.down.sql           # Initial schema rollback
├── 000002_add_user_profiles.up.sql   # User profiles feature
├── 000002_add_user_profiles.down.sql # User profiles rollback
├── development/                       # Development-specific migrations
│   ├── 000003_dev_test_data.up.sql
│   └── 000003_dev_test_data.down.sql
├── staging/                          # Staging-specific migrations
├── production/                       # Production-specific migrations
├── docs/                             # Migration documentation
└── templates/                        # Migration templates
```

## Migration Naming Convention

- **Format**: `NNNNNN_description.up.sql` / `NNNNNN_description.down.sql`
- **NNNNNN**: Zero-padded migration number (000001, 000002, etc.)
- **description**: Snake_case description of the migration purpose
- **.up.sql**: Migration to apply the changes
- **.down.sql**: Migration to rollback the changes

## Schema Architecture

All user service tables are created in the `user_service` schema:

```sql
-- Example table creation
CREATE TABLE user_service.users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## Migration Workflow

### 1. Create New Migration
```bash
# Generate migration with template
make db-migration-generate NAME=add_user_preferences TYPE=table

# Or create manually
# Copy template and modify
```

### 2. Validate Migration
```bash
# Validate syntax and dependencies
make db-validate
```

### 3. Test Migration
```bash
# Test on development environment
make db-migrate-up

# Check results
make db-tables
make db-counts
```

### 4. Rollback if Needed
```bash
# Rollback last migration
make db-migrate-down

# Rollback to specific version
make db-migrate-goto VERSION=000001
```

## Environment-Specific Migrations

### Development Environment
- Contains test data and development-specific features
- Safe to modify and reset frequently
- Includes debug indexes and test users

### Staging Environment
- Mirrors production structure
- Contains minimal test data
- Used for pre-production validation

### Production Environment
- Only essential migrations
- Requires approval for high-risk changes
- Includes performance optimizations

## Dependencies Management

Migration dependencies are tracked in `dependencies.json`:

```json
{
  "migrations": {
    "000001_initial": {
      "description": "Create user_service schema and users table",
      "depends_on": [],
      "affects_tables": ["user_service.users"],
      "estimated_duration": "30s",
      "risk_level": "low",
      "rollback_safe": true
    }
  }
}
```

### Checking Dependencies
```bash
# View dependency graph
make db-migration-deps
```

## Best Practices

### Migration Writing
- ✅ Always include both `.up.sql` and `.down.sql` files
- ✅ Use schema-qualified table names (`user_service.table_name`)
- ✅ Include comments explaining the migration purpose
- ✅ Test rollbacks thoroughly
- ✅ Keep migrations idempotent when possible

### Risk Assessment
- **Low Risk**: Adding indexes, constraints, comments
- **Medium Risk**: Adding columns, modifying data
- **High Risk**: Dropping tables, changing column types

### Performance Considerations
- Avoid long-running operations in production migrations
- Use `CONCURRENTLY` for index creation in production
- Consider table locking implications
- Test migration performance on production-sized data

## Troubleshooting

### Common Issues

1. **Migration Fails**: Check SQL syntax and dependencies
2. **Rollback Fails**: Verify down migration is correct
3. **Performance Issues**: Use `EXPLAIN` to analyze slow queries
4. **Lock Conflicts**: Use `CONCURRENTLY` for index operations

### Recovery Procedures

1. **Failed Migration**:
   ```bash
   # Check migration status
   make db-migrate-status

   # Manual rollback if needed
   make db-migrate-down
   ```

2. **Data Inconsistency**:
   ```bash
   # Create backup
   make db-backup

   # Restore from backup if needed
   make db-restore FILE=backup_file.sql
   ```

## Commands Reference

```bash
# Basic migration operations
make db-migrate-up              # Run all pending migrations
make db-migrate-down            # Rollback last migration
make db-migrate-status          # Show migration status
make db-migrate-goto VERSION=000001  # Go to specific version

# Advanced operations
make db-migration-generate NAME=feature TYPE=table  # Generate new migration
make db-validate               # Validate migrations
make db-migration-deps         # Show dependencies
make db-backup                 # Create backup

# Data operations
make db-seed                   # Basic seeding
make db-seed-enhanced ENV=development  # Environment-specific seeding
make db-clean                  # Clean all data

# Inspection
make db-tables                 # List tables
make db-counts                 # Show row counts
make db-schema                 # Show schema info
```

## Contributing

1. Always create both up and down migrations
2. Update `dependencies.json` with new migration info
3. Add documentation in `docs/` directory
4. Test migrations in development environment
5. Validate with `make db-validate` before committing

## Related Documentation

- [Database Setup Guide](../../docs/database-setup.md)
- [API Documentation](../../docs/api.md)
- [Deployment Guide](../../docs/deployment.md)