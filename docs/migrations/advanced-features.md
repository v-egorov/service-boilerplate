# Advanced Migration Features

## Service-Specific Migration Tracking

### Overview

Each service maintains its own migration tracking table to ensure complete isolation between services. This allows multiple services to coexist in the same database with independent migration histories.

### Migration Table Naming

Migration tables follow the pattern: `{service_name}_schema_migrations`

Examples:
- `user_service_schema_migrations` - Tracks user-service migrations
- `dynamic_service_schema_migrations` - Tracks dynamic-service migrations
- `product_service_schema_migrations` - Tracks product-service migrations

### Benefits

- **Service Isolation**: Each service can migrate independently
- **Version Independence**: Services can have different migration versions
- **Rollback Safety**: Rolling back one service doesn't affect others
- **Parallel Development**: Multiple services can be developed simultaneously

### Database Architecture

```
service_db (PostgreSQL)
├── public
│   ├── user_service_schema_migrations
│   ├── dynamic_service_schema_migrations
│   └── [other_service]_schema_migrations
├── user_service
│   ├── users, user_profiles, user_settings
├── dynamic_service
│   └── entities
```

## Migration Dependencies

### Overview

Migration dependencies ensure that migrations are applied in the correct order within each service. The system automatically tracks and validates dependencies to prevent inconsistent database states.

### Dependency Configuration

Dependencies are defined in `services/{service}/migrations/dependencies.json`:

```json
{
  "migrations": {
    "000001": {
      "description": "Create user_service schema and users table",
      "depends_on": [],
      "affects_tables": ["user_service.users"],
      "estimated_duration": "30s",
      "risk_level": "low",
      "rollback_safe": true
    },
    "000002": {
      "description": "Add user profiles functionality",
      "depends_on": ["000001"],
      "affects_tables": ["user_service.users"],
      "estimated_duration": "45s",
      "risk_level": "medium",
      "rollback_safe": true
    },
    "000003": {
      "description": "Add development test data",
      "depends_on": ["000001"],
      "affects_tables": ["user_service.users"],
      "estimated_duration": "15s",
      "risk_level": "low",
      "rollback_safe": true,
      "environment": "development"
    }
  }
}
```

### Dependency Validation

```bash
# Check dependency graph
make db-migration-deps

# Output:
000001: []
000002: ["000001"]
000003: ["000001"]
000004: ["000003"]
```

### Circular Dependency Detection

The system automatically detects circular dependencies:

```json
{
  "migrations": {
    "000001": { "depends_on": ["000003"] },
    "000002": { "depends_on": ["000001"] },
    "000003": { "depends_on": ["000002"] }  // Circular dependency!
  }
}
```

Validation will fail with:
```
❌ Circular dependency detected: 000001 -> 000003 -> 000002 -> 000001
```

## Environment-Specific Migrations

### Directory Structure

```
services/user-service/migrations/
├── common/                    # All environments
│   ├── 000001_initial.up.sql
│   └── 000002_core_features.up.sql
├── development/              # Development only
│   ├── 000003_dev_data.up.sql
│   └── 000004_debug_indexes.up.sql
├── staging/                  # Staging only
│   └── 000005_perf_indexes.up.sql
└── production/               # Production only
    └── 000006_prod_optimizations.up.sql
```

### Environment Configuration

Defined in `environments.json`:

```json
{
  "environments": {
    "development": {
      "description": "Development environment with test data",
      "migrations": ["000003_dev_data.up.sql", "000004_debug_indexes.up.sql"],
      "config": {
        "allow_destructive_operations": true,
        "skip_validation": false,
        "auto_rollback_on_failure": true
      }
    },
    "staging": {
      "description": "Pre-production testing",
      "migrations": ["000005_perf_indexes.up.sql"],
      "config": {
        "allow_destructive_operations": false,
        "skip_validation": false,
        "auto_rollback_on_failure": true,
        "require_approval": true
      }
    },
    "production": {
      "description": "Live production environment",
      "migrations": ["000006_prod_optimizations.up.sql"],
      "config": {
        "allow_destructive_operations": false,
        "skip_validation": false,
        "auto_rollback_on_failure": false,
        "require_approval": true,
        "maintenance_window_required": true,
        "backup_required": true
      }
    }
  },
  "current_environment": "development"
}
```

### Environment-Specific Execution

```bash
# Development environment
export MIGRATION_ENV=development
make db-migrate-up

# Staging environment
export MIGRATION_ENV=staging
make db-migrate-up

# Production environment (with approval)
export MIGRATION_ENV=production
make db-migrate-up
```

## Advanced Validation Features

### Pre-Flight Checks

The validation system performs comprehensive checks:

```bash
make db-validate

# Output:
🔍 Validating migrations for service: user-service
✅ Dependencies file exists
🔍 Validating SQL syntax...
   Checking: 000001_initial.up.sql
   Checking: 000002_profiles.up.sql
   ⚠️  WARNING: CASCADE drop detected in 000002_profiles.down.sql
✅ SQL validation completed
🔍 Validating migration file structure...
   Found 3 migration groups
✅ Migration file structure validated
🔍 Validating migration dependencies...
   Found 3 migrations in dependencies file
✅ Dependency validation completed
```

### SQL Syntax Validation

- **Basic Syntax**: Checks for valid SQL statements
- **Schema Qualification**: Ensures all tables use proper schema prefixes
- **Dangerous Operations**: Warns about CASCADE drops, DELETE without WHERE, etc.
- **Best Practices**: Suggests improvements for performance and maintainability

### Dependency Graph Analysis

- **Topological Sorting**: Ensures migrations can be applied in order
- **Missing Dependencies**: Detects migrations that depend on non-existent migrations
- **Orphaned Migrations**: Finds migrations not referenced in dependency graph
- **Impact Analysis**: Shows which tables are affected by each migration

## Migration Hooks and Automation

### Pre/Post Migration Hooks

Create hook scripts for automation:

```bash
# pre_migrate_000001.sh
#!/bin/bash
echo "🚀 Starting migration 000001"
echo "📊 Current database size: $(make db-counts | grep user_service | wc -l) tables"

# Backup before migration
make db-backup

# Notify team
curl -X POST -H 'Content-type: application/json' \
  --data '{"text":"Starting migration 000001"}' \
  $SLACK_WEBHOOK_URL
```

```bash
# post_migrate_000001.sh
#!/bin/bash
echo "✅ Migration 000001 completed"
echo "📊 New database size: $(make db-counts | grep user_service | wc -l) tables"

# Verify migration success
if make db-tables | grep -q "user_service.users"; then
  echo "✅ Users table created successfully"
else
  echo "❌ Users table not found"
  exit 1
fi

# Run post-migration tests
make test-all
```



## Performance Monitoring

### Migration Execution Tracking

```bash
# Track migration performance
time make db-migrate-up

# Output:
📈 Running migrations up...
1/u initial (54.183163ms)
2/u add_user_profiles (70.112169ms)
3/u dev_test_data (15.234567ms)

real    0m0.140s
user    0m0.020s
sys     0m0.010s
```

### Performance Metrics Collection

```sql
-- Create performance tracking table
CREATE TABLE user_service.migration_performance (
    migration_id VARCHAR(20) PRIMARY KEY,
    execution_time_ms INTEGER NOT NULL,
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    environment VARCHAR(20) NOT NULL,
    success BOOLEAN NOT NULL
);

-- Track migration performance
INSERT INTO user_service.migration_performance
VALUES ('000001', 54183, CURRENT_TIMESTAMP, 'development', true);
```

### Slow Migration Detection

```bash
# Check for slow migrations
make db-migrate-status | awk '$2 > 30000 {print "⚠️  Slow migration: " $1 " (" $2 "ms)"}'

# Output:
⚠️  Slow migration: 000002 (70112ms)
```

## Backup and Recovery Integration

### Automated Backup Strategy

```bash
# pre_backup.sh - Run before migrations
#!/bin/bash
BACKUP_FILE="backup_pre_migration_$(date +%Y%m%d_%H%M%S).sql"

echo "💾 Creating pre-migration backup: $BACKUP_FILE"
docker-compose --env-file .env -f docker/docker-compose.yml exec postgres \
  pg_dump -U postgres -d service_db --no-owner --no-privileges > "$BACKUP_FILE"

if [ $? -eq 0 ]; then
  echo "✅ Backup created: $BACKUP_FILE ($(du -h "$BACKUP_FILE" | cut -f1))"
else
  echo "❌ Backup failed"
  exit 1
fi
```

### Recovery Procedures

```bash
# recovery.sh - Automated recovery
#!/bin/bash
echo "🔄 Starting database recovery"

# Find latest backup
LATEST_BACKUP=$(ls -t backup_*.sql | head -1)

if [ -z "$LATEST_BACKUP" ]; then
  echo "❌ No backup files found"
  exit 1
fi

echo "📁 Using backup: $LATEST_BACKUP"

# Restore from backup
make db-restore FILE="$LATEST_BACKUP"

if [ $? -eq 0 ]; then
  echo "✅ Database restored from $LATEST_BACKUP"
else
  echo "❌ Restore failed"
  exit 1
fi
```

## Multi-Service Coordination

### Cross-Service Dependencies

```json
// services/user-service/migrations/dependencies.json
{
  "migrations": {
    "000005": {
      "description": "Add user notifications",
      "depends_on": ["000001"],
      "cross_service_dependencies": {
        "notification-service": ["000002"]
      }
    }
  }
}
```

### Service Coordination Script

```bash
# deploy_all_services.sh
#!/bin/bash
SERVICES=("user-service" "notification-service" "api-gateway")

for service in "${SERVICES[@]}"; do
  echo "🚀 Deploying $service"

  # Run service-specific migrations
  cd "services/$service"
  make db-migrate-up

  if [ $? -ne 0 ]; then
    echo "❌ Migration failed for $service"
    exit 1
  fi

  # Deploy service
  make deploy

  cd ../..
done

echo "✅ All services deployed successfully"
```

## Advanced Rollback Strategies

### Point-in-Time Recovery

```bash
# Rollback to specific migration


# Rollback multiple migrations


# Safe rollback with validation
make db-rollback-safe MIGRATION=000004
```

### Selective Rollback

```bash
# Rollback specific migration without affecting others
make db-rollback-selective MIGRATION=000005

# This keeps other migrations intact while rolling back only the specified one
```

### Data-Aware Rollback

```sql
-- Check for data dependencies before rollback
SELECT COUNT(*) as dependent_records
FROM user_service.user_profiles
WHERE user_id IN (
  SELECT id FROM user_service.users
  WHERE created_at > '2024-09-01'
);

-- Only rollback if no dependent data
DO $$
BEGIN
  IF (SELECT COUNT(*) FROM user_service.user_profiles) = 0 THEN
    -- Safe to rollback
    DROP TABLE user_service.user_profiles;
  ELSE
    -- Cannot rollback due to dependent data
    RAISE EXCEPTION 'Cannot rollback: dependent data exists';
  END IF;
END $$;
```

## Monitoring and Alerting

### Migration Dashboard

```bash
# migration_monitor.sh
#!/bin/bash
echo "📊 Migration Status Dashboard"
echo "================================"

echo "🔄 Current Status:"
make db-migrate-status

echo ""
echo "📈 Performance Metrics:"
echo "Last migration time: $(grep "migration_time" /tmp/migration_metrics.log | tail -1)"

echo ""
echo "⚠️  Warnings:"
make db-validate 2>&1 | grep "WARNING" || echo "No warnings"

echo ""
echo "🚨 Alerts:"
if make db-migrate-status | grep -q "dirty"; then
  echo "❌ Database in dirty state - requires attention"
fi
```

### Automated Notifications

```bash
# notify_migration_status.sh
#!/bin/bash
WEBHOOK_URL="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"

# Send success notification
curl -X POST -H 'Content-type: application/json' \
  --data "{
    \"text\": \"✅ Migration completed successfully\",
    \"attachments\": [
      {
        \"color\": \"good\",
        \"fields\": [
          {\"title\": \"Service\", \"value\": \"user-service\", \"short\": true},
          {\"title\": \"Environment\", \"value\": \"$ENVIRONMENT\", \"short\": true},
          {\"title\": \"Duration\", \"value\": \"$DURATION\", \"short\": true}
        ]
      }
    ]
  }" \
  $WEBHOOK_URL
```

## Best Practices for Advanced Features

### Dependency Management
- Keep dependency chains short (max 3-4 levels)
- Document why dependencies exist
- Regularly review and refactor dependency graph
- Use dependency visualization tools

### Environment Handling
- Test environment-specific migrations thoroughly
- Document environment differences
- Use feature flags for environment-specific behavior
- Maintain separate validation rules per environment

### Performance Optimization
- Monitor migration execution times
- Optimize large data migrations with batching
- Use appropriate indexing strategies
- Consider maintenance windows for production

### Error Handling
- Implement comprehensive error handling
- Provide clear error messages
- Include recovery procedures
- Log all migration activities

## Troubleshooting Advanced Features

### Dependency Issues

```bash
# Check for missing dependencies
make db-migration-deps | grep "missing"

# Validate dependency graph
make db-validate | grep "dependency"
```

### Environment Conflicts

```bash
# Check environment configuration
cat services/user-service/migrations/environments.json

# Validate environment-specific files
find services/user-service/migrations -name "*${ENVIRONMENT}*" -type f
```

### Performance Problems

```bash
# Monitor slow migrations
time make db-migrate-up

# Check database locks
make db-connect
SELECT * FROM pg_locks WHERE NOT granted;
```

## Next Steps

- **[Best Practices](./best-practices.md)**: Comprehensive guidelines
- **[Troubleshooting](./troubleshooting.md)**: Advanced problem solving
- **[API Reference](./api-reference.md)**: Complete command reference
- **[Examples](./examples.md)**: Advanced migration examples