# Migration API Reference

## Command Reference

### Core Migration Commands

#### `make db-migrate-up`
Apply all pending migrations to the database.

```bash
# Apply all pending migrations
make db-migrate-up

# Output:
📈 Running migrations up...
1/u initial (54.183163ms)
2/u add_user_profiles (70.112169ms)
3/u dev_test_data (15.234567ms)
✅ All migrations applied successfully
```

**Options:**
- Works with environment-specific migrations
- Automatically resolves dependencies
- Provides execution timing

**Error Handling:**
- Stops on first failure
- Leaves database in "dirty" state if failed
- Requires manual intervention to fix

#### `make db-migrate-down`
Rollback the last applied migration.

```bash
# Rollback last migration
make db-migrate-down

# Output:
⏪ Running migrations down...
3/d dev_test_data (12.345678ms)
✅ Migration rollback completed
```

**Options:**
- Only rolls back one migration at a time
- Safe rollback with dependency checking
- Preserves data integrity

#### `make db-migrate-status`
Show the current migration status and applied versions.

```bash
# Check migration status
make db-migrate-status

# Output:
database: postgres://postgres:postgres@postgres:5432/service_db?sslmode=disable

000001_initial ............................................... 2024-09-07T20:15:23Z
000002_add_user_profiles ...................................... 2024-09-07T20:15:23Z
000003_dev_test_data .......................................... 2024-09-07T20:15:24Z

No migrations to apply
```

**Status Indicators:**
- ✅ Applied migration (shows timestamp)
- ⏳ Pending migration (not yet applied)
- ❌ Failed migration (dirty state)



### Advanced Migration Commands

#### `make db-migrate-generate NAME={name} TYPE={type}`
Generate a new migration with templates and documentation.

```bash
# Generate a table migration
make db-migrate-generate NAME=add_user_notifications TYPE=table

# Generate an index migration
make db-migrate-generate NAME=optimize_user_queries TYPE=index

# Generate a data migration
make db-migrate-generate NAME=migrate_user_data TYPE=data
```

**Parameters:**
- `NAME`: Descriptive name (snake_case recommended)
- `TYPE`: Migration type (table, index, data, schema)

**Generated Files:**
- `services/user-service/migrations/development/00000X_{name}.up.sql`
- `services/user-service/migrations/development/00000X_{name}.down.sql`
- `services/user-service/migrations/docs/migration_00000X.md`

#### `make db-validate`
Comprehensive validation of migration files and dependencies.

```bash
# Validate all migrations
make db-validate

# Output:
🔍 Validating migrations for service: user-service
✅ Dependencies file exists
🔍 Validating SQL syntax...
   Checking: 000001_initial.up.sql
   Checking: 000002_profiles.up.sql
✅ SQL validation completed
🔍 Validating migration file structure...
   Found 3 migration groups
✅ Migration file structure validated
🔍 Validating migration dependencies...
   Found 3 migrations in dependencies file
✅ Dependency validation completed
```

**Validation Checks:**
- SQL syntax validation
- File structure (up/down pairs)
- Schema qualification
- JSON config validity

#### `make db-migration-deps`
Display migration information (legacy command).

```bash
# Show migration info
make db-migration-deps
```
000003: ["000001"]
000004: ["000003"]
```

**Dependency Format:**
- `migration_id: [dependency1, dependency2, ...]`
- Empty array `[]` means no dependencies
- Used for execution order planning

---

### Data Management Commands

#### `make db-seed`
Apply basic database seeding with test data.

```bash
# Seed with basic data
make db-seed

# Output:
🌱 Seeding database with test data...
INSERT 0 5
✅ Database seeded with 5 test users
```

**Behavior:**
- Inserts test data into `user_service.users`
- Skips existing records (ON CONFLICT DO NOTHING)
- Safe to run multiple times

#### `make db-seed-enhanced ENV={environment}`
Environment-specific database seeding.

```bash
# Development seeding
make db-seed-enhanced ENV=development

# Staging seeding
make db-seed-enhanced ENV=staging

# Production seeding
make db-seed-enhanced ENV=production
```

**Environment Options:**
- `development`: Includes test users and development data
- `staging`: Minimal data for testing
- `production`: Only essential system data

**Features:**
- Environment-aware data loading
- Duplicate prevention
- Validation of seed file existence

#### `make db-clean`
Clean all data from tables while preserving schema.

```bash
# Clean all data
make db-clean

# Output:
🧹 Cleaning database data...
TRUNCATE TABLE
✅ Database data cleaned (schema preserved)
```

**Behavior:**
- Uses `TRUNCATE` with `CASCADE`
- Preserves table structure and indexes
- Fast operation for large datasets

---

### Database Inspection Commands

#### `make db-tables`
List all tables in the database with structure information.

```bash
# List all tables
make db-tables

# Output:
📊 Table Structures:
     table_name
-------------------
 schema_migrations
 users
(2 rows)

                                           Table "user_service.users"
    Column   |           Type           | Collation | Nullable |                    Default
------------+--------------------------+-----------+----------+------------------------------------------------
  id         | integer                  |           | not null | nextval('user_service.users_id_seq'::regclass)
  email      | character varying(255)   |           | not null |
  first_name | character varying(100)   |           | not null |
  last_name  | character varying(100)   |           | not null |
  created_at | timestamp with time zone |           |          | CURRENT_TIMESTAMP
  updated_at | timestamp with time zone |           |          | CURRENT_TIMESTAMP

Service-Specific Migration Tracking:
- user_service_schema_migrations (tracks user-service migrations)
- dynamic_service_schema_migrations (tracks dynamic-service migrations)
- Each service maintains independent migration history
```

**Information Provided:**
- Table names by schema
- Detailed column information
- Indexes and constraints
- Foreign key relationships

#### `make db-counts`
Show row counts and sizes for all tables.

```bash
# Show table statistics
make db-counts

# Output:
🔢 Table Statistics:
    schema    | table | size  | row_count
--------------+-------+-------+-----------
 user_service | users | 72 kB |         5
(1 row)
```

**Metrics:**
- Table size on disk
- Row count per table
- Schema organization
- Storage efficiency

#### `make db-schema`
Show database schema information and indexes.

```bash
# Show schema details
make db-schema

# Output:
📋 Database Schema:
                List of relations
 Schema |       Name       |   Type   |  Owner
--------+------------------+----------+----------
 public | schema_migrations| table    | postgres
 user_service | users       | table    | postgres

📊 Indexes:
                                      List of relations
    Schema    |                Name                | Type  |  Owner   |     Table
--------------+-------------------------------------+-------+----------+---------------
 user_service | users_pkey                         | index | postgres | users
 user_service | idx_users_email                    | index | postgres | users
 user_service | idx_users_created_at               | index | postgres | users
```

---

### Database Connection Commands

#### `make db-connect`
Open an interactive PostgreSQL shell.

```bash
# Connect to database
make db-connect

# Output:
psql (15.14)
Type "help" for help.

service_db=#
```

**Features:**
- Direct database access
- Full SQL command execution
- Schema inspection capabilities
- Manual migration testing

#### `make db-status`
Show database connection status and basic health metrics.

```bash
# Check database status
make db-status

# Output:
📊 Database Status:
                 PostgreSQL 15.14 on x86_64-pc-linux-musl, compiled by gcc (Alpine 14.2.0) 14.2.0, 64-bit
Active connections: 3
Database size: 12 MB
```

#### `make db-health`
Perform database health check and connectivity test.

```bash
# Health check
make db-health

# Output:
🏥 Database Health Check:
✅ Database is healthy and accepting connections
Connection time: 2ms
Active connections: 3
```

---

### Utility Commands

#### `make db-backup`
Create a timestamped database backup.

```bash
# Create backup
make db-backup

# Output:
💾 Creating database backup...
✅ Database backup created: backup_20240907_201500.sql
   Size: 4.2K
```

**Features:**
- Timestamped backup files
- Size reporting
- Non-destructive operation
- Ready for restoration

#### `make db-restore FILE={filename}`
Restore database from a backup file.

```bash
# Restore from backup
make db-restore FILE=backup_20240907_201500.sql

# Output:
🔄 Restoring database from backup_20240907_201500.sql...
✅ Database restored from backup_20240907_201500.sql
```

**Safety Features:**
- File existence validation
- Backup integrity checking
- Confirmation prompts for destructive operations

#### `make db-fresh`
Complete database reset with volume cleanup.

```bash
# Fresh database setup
make db-fresh

# Output:
💾 Cleaning Docker volumes...
🔧 Handling PostgreSQL volume permissions...
✅ Docker volumes cleaned
🆕 Creating database service_db...
ℹ️  Database service_db already exists or creation failed
📈 Running migrations up...
✅ Fresh database ready
```

**Use Cases:**
- Development environment reset
- Testing clean slate scenarios
- Troubleshooting corrupted databases

---

## Environment Variables

### Migration Configuration

```bash
# Environment selection
export MIGRATION_ENV=development  # development, staging, production

# Debug mode
export MIGRATION_DEBUG=true       # Enable verbose logging

# Custom paths
export MIGRATION_DIR=custom/path  # Custom migration directory
export BACKUP_DIR=/custom/backups # Custom backup location
```

### Database Configuration

```bash
# Connection settings
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export DATABASE_NAME=service_db
export DATABASE_USER=postgres
export DATABASE_PASSWORD=postgres
export DATABASE_SSL_MODE=disable
```

## Exit Codes

### Success Codes
- `0`: Command executed successfully
- Migration applied, validation passed, backup created

### Error Codes
- `1`: General error (validation failed, migration error)
- `2`: File not found (missing migration files, backup files)
- `3`: Connection error (database unreachable)
- `4`: Permission error (insufficient database permissions)
- `5`: Dependency error (missing migration dependencies)

## File Formats

### Migration Files (`*.up.sql`, `*.down.sql`)

```sql
-- Migration: 000005_add_user_settings
-- Description: Add user settings table
-- Created: Auto-generated

-- UP: Changes to apply
CREATE TABLE user_service.user_settings (...);

-- DOWN: Changes to rollback
DROP TABLE user_service.user_settings CASCADE;
```

### Environment Configuration (`environments.json`)

```json
{
  "environments": {
    "development": {
      "migrations": "development",
      "config": {
        "allow_destructive_operations": true
      }
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

## Performance Metrics

### Typical Execution Times

| Command | Typical Time | Factors |
|---------|-------------|---------|
| `db-validate` | 2-5 seconds | Number of migration files |
| `db-migrate-up` | 10-60 seconds | Migration complexity |
| `db-seed` | 1-3 seconds | Data volume |
| `db-backup` | 5-30 seconds | Database size |
| `db-tables` | 1-2 seconds | Number of tables |

### Resource Usage

- **Memory**: Minimal (mostly file I/O)
- **Disk**: Backup size + temporary files
- **Network**: Database connections only
- **CPU**: SQL parsing and execution

## Integration Examples

### CI/CD Pipeline

```yaml
# .github/workflows/migrations.yml
name: Database Migrations
on:
  push:
    paths:
      - 'services/*/migrations/**'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Validate Migrations
        run: make db-validate

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Test Migrations
        run: |
          make db-migrate-up
          make db-seed-enhanced ENV=development
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Validate migrations before commit
if git diff --cached --name-only | grep -q "migrations/"; then
    echo "🔍 Validating migrations..."
    make db-validate

    if [ $? -ne 0 ]; then
        echo "❌ Migration validation failed"
        exit 1
    fi
fi
```

## Troubleshooting Quick Reference

| Issue | Command | Solution |
|-------|---------|----------|
| Migration fails | `make db-validate` | Check syntax and dependencies |
| Rollback fails | `make db-migrate-status` | Check dirty state |
| Connection error | `make db-health` | Verify database connectivity |
| Performance issue | `time make db-migrate-up` | Measure execution time |
| Data inconsistency | `make db-counts` | Check row counts |
| Dependency conflict | `make db-migration-deps` | Review dependency graph |

## Command Dependencies

```mermaid
graph TD
    A[db-setup] --> B[db-create]
    A --> C[db-migrate-up]
    A --> D[db-seed]

    C --> E[db-validate]
    C --> F[db-migration-deps]

    G[db-migrate-generate] --> H[db-validate]

    I[db-restore] --> J[db-backup]

    K[db-fresh] --> L[db-clean]
    K --> M[db-setup]
```

This API reference provides comprehensive coverage of all migration system commands, their usage patterns, and integration examples.