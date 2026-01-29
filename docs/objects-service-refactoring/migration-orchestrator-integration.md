# Migration Orchestrator Integration

## Overview

The **Migration Orchestrator** provides enterprise-grade migration management for multi-service architectures. This guide explains how to integrate it into the **objects-service** following the pattern used by `user-service` and `auth-service`.

## Why Use the Orchestrator

| Feature | Manual SQL Files | Migration Orchestrator |
|---------|----------------|-----------------|
| Dependency Management | ❌ Manual | ✅ Topological sorting |
| Environment-Specific Migrations | ❌ Manual | ✅ `development/`, `staging`, `production` sets |
| Execution Tracking | ❌ Manual logs | ✅ Database-level tracking |
| Risk Assessment | ❌ Manual review | ✅ Automated warnings |
| Intelligent Rollback | ❌ Manual steps | ✅ State-aware rollback |
| Cross-Service Dependencies | ❌ Manual coordination | ✅ Automatic resolution |
| Validation | ❌ Manual checks | ✅ Schema validation |

## Architecture

### Orchestrator Structure

```
migration-orchestrator/
├── cmd/
│   ├── init.go              # Initialize tracking
│   ├── up.go               # Run migrations up
│   ├── down.go             # Run migrations down
│   ├── status.go           # Show migration status
│   ├── validate.go         # Validate migration integrity
│   └── list.go             # List all migrations
├── internal/
│   ├── orchestrator/         # Core orchestration logic
│   │   └── orchestrator.go # Main implementation
│   ├── integration_test.go   # Integration tests
│   ├── database/            # Database connection management
│   └── config/             # Configuration loading
└── pkg/
    ├── types/               # Data structures
    │   └── migration.go
    └── utils/               # Logging and utilities
```

### Service Schema Extension

The orchestrator extends each service's schema with a tracking table:

```sql
-- Added to each service database
CREATE TABLE service_schema.migration_executions (
    id BIGSERIAL PRIMARY KEY,
    migration_id VARCHAR(255) NOT NULL,
    migration_version VARCHAR(255) NOT NULL,
    environment VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed', 'rolled_back')),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_ms BIGINT,
    executed_by VARCHAR(255),
    checksum VARCHAR(255),
    dependencies JSONB,
    metadata JSONB,
    error_message TEXT,
    rollback_version VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_migration_env_status ON service_schema.migration_executions(environment, status);
```

## File Structure for Objects-Service

Following the pattern from other services, create this structure:

```
services/objects-service/
├── migrations/
│   ├── dependencies.json           # Migration dependencies
│   ├── environments.json             # Environment-specific migrations
│   ├── 000001_initial.up.sql      # Base schema
│   ├── 000001_initial.down.sql    # Rollback script
│   ├── development/
│   │   └── 000002_dev_tax_test_data.up.sql
│   └── 000002_dev_tax_test_data.down.sql
└── Makefile                      # Orchestrator commands (optional)
```

### Configuration Files

**dependencies.json** - Defines migration dependencies:

```json
{
  "migrations": {
    "000001_initial": {
      "description": "Create object_types and objects tables",
      "depends_on": [],
      "affects_tables": ["object_types", "objects"],
      "risk_level": "high",
      "estimated_duration": "5m",
      "rollback_safe": true
    },
    "000002_dev_tax_test_data": {
      "description": "Load development taxonomy test data",
      "depends_on": ["000001_initial"],
      "affects_tables": ["object_types", "objects"],
      "risk_level": "low",
      "estimated_duration": "1m",
      "rollback_safe": true
    }
  }
}
```

**environments.json** - Environment-specific migration sets:

```json
{
  "development": {
    "description": "Development environment",
    "migrations": [
      "000001_initial",
      "000002_dev_tax_test_data"
    ],
    "seed_files": []
  },
  "staging": {
    "description": "Staging environment",
    "migrations": [
      "000001_initial"
    ],
    "seed_files": []
  },
  "production": {
    "description": "production environment",
    "migrations": [
      "000001_initial"
    ],
    "seed_files": []
  }
}
}
```

## Integration Options

### Option 1: CLI Commands (Simple)

Run orchestrator from any directory:

```bash
# Navigate to project root
cd /home/vegorov/ai/service-boilerplate

# Initialize tracking for objects-service
cd ../migration-orchestrator
make db-migrate-init SERVICE_NAME=objects-service

# Run migrations up
cd ../migration-orchestrator
make db-migrate-up SERVICE_NAME=objects-service

# Check status
cd ../migration-orchestrator
make db-migrate-status SERVICE_NAME=objects-service

# Validate migrations
cd ../migration-orchestrator
make db-migrate-validate SERVICE_NAME=objects-service
```

### Option 2: Import as Go Library

Integrate orchestrator into main.go:

```go
package main

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    
    orchestrator "github.com/v-egorov/service-boilerplate/migration-orchestrator/internal/orchestrator"
    
    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/handlers"
    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/services"
    "github.com/v-egorov/service-boilerplate/common/database"
)

    "github.com/v-egorov/service-boilerplate/services/objects-service/pkg/config"
)

func main() {
    cfg, err := config.Load()
    logger := logrus.New()
    
    // Database connection
    db, err := database.NewPostgresDB(database.Config{
        Host:        cfg.DB.Host,
        Port:        cfg.DB.Port,
        User:        cfg.DB.User,
        Password:    cfg.DB.Password,
        Database:    cfg.DB.Database,
        SSLMode:     cfg.DB.SSLMode,
        MaxConns:    cfg.DB.MaxConns,
        MinConns:    cfg.DB.MinConns,
        MaxConnIdle: time.Hour,
        MaxConnLife: 24 * time.Hour,
    }, logger)
    
    defer db.Close()
    
    // Run migrations on startup (recommended for production)
    if err := runMigrationsOnStartup(ctx, db, cfg, logger); err != nil {
        logger.Fatalf("Failed to run migrations: %v", err)
    }
    
    // ... rest of application startup ...
}

func runMigrationsOnStartup(ctx context.Context, db *database.PostgresDB, cfg *config.Config, logger *logrus.Logger) error {
    logger.Info("Running migrations...")
    
    orchestratorConfig := orchestrator.Config{
        ServiceName: "objects-service",
        BasePath:   "services/objects-service/migrations",
        Environment: os.Getenv("APP_ENV"), // development, staging, production
        DB:         db.GetPool(),
    }
    
    client := orchestrator.NewClient(orchestratorConfig)
    
    err := client.Migrate(ctx)
    if err != nil {
        return fmt.Errorf("failed to run migrations: %w", err)
    }
    
    logger.Info("Migrations completed successfully")
    return nil
}
```

### Option 3: Orchestrator as Middleware

Run migrations before each request (development only):

```go
func migrationMiddleware(logger *logrus.Logger, orchestratorClient *orchestrator.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        if os.Getenv("APP_ENV") == "development" {
            logger.Debug("Checking for pending migrations...")
            
            status, err := orchestratorClient.GetStatus(ctx)
            if err != nil {
                logger.WithError(err).Error("Failed to check migration status")
                c.Next()
                return
            }
            
            if hasPendingMigrations(status) {
                logger.Info("Running pending migrations...")
                if err := orchestratorClient.Migrate(ctx); err != nil {
                    logger.WithError(err).Error("Failed to run migrations")
                    c.JSON(500, gin.H{"error": "migration failed"})
                    c.Abort()
                    return
                }
            }
        }
        
        c.Next()
    }
}

func hasPendingMigrations(status *orchestrator.StatusResponse) bool {
    for _, svc := range status.Services {
        if svc.Status == "pending" || svc.Status == "failed" {
            return true
        }
    }
    return false
}
```

## Common Commands

### Initialization

```bash
# From migration-orchestrator directory
make db-migrate-init SERVICE_NAME=objects-service

# From project root
cd migration-orchestrator && make db-migrate-init SERVICE_NAME=objects-service
```

Creates `service_schema.migration_executions` table in objects-service database.

### Running Migrations

```bash
# Run all pending migrations
make db-migrate-up SERVICE_NAME=objects-service

# Run specific migration
make db-migrate-up SERVICE_NAME=objects-service COUNT=1

# Run specific step of a migration
make db-migrate-up SERVICE_NAME=objects-service STEP=1

# Rollback last migration
make db-migrate-down SERVICE_NAME=objects-service

# Rollback multiple migrations
make db-migrate-down SERVICE_NAME=objects-service STEPS=3
```

### Status & Validation

```bash
# Show overall status
make db-migrate-status SERVICE_NAME=objects-service

# Show detailed status for all services
make db-migrate-status

# Validate migration integrity
make db-migrate-validate SERVICE_NAME=objects-service

# List all migrations
make db-migrate-list SERVICE_NAME=objects-service

# Show execution history
make db-migrate-history SERVICE_NAME=objects-service LIMIT=20
```

### Advanced Operations

```bash
# Force rerun failed migration
make db-migrate-up SERVICE_NAME=objects-service FORCE=true

# Dry-run (preview)
make db-migrate-up SERVICE_NAME=objects-service DRY_RUN=true

# Check for conflicts
make db-migrate-conflicts SERVICE_NAME=objects-service

# Generate rollback plan
make db-migrate-rollback-plan SERVICE_NAME=objects-service
```

## Migration File Guidelines

### Naming Convention

- Format: `NNNNN_description`
- Example: `000001_initial`, `000002_dev_tax_test_data`

### Structure

**Up Migration Files**:
1. Define schema changes
2. Include indexes and constraints
3. Add comments for documentation
4. Idempotent (can run multiple times safely)

**Down Migration Files**:
1. Reverse all up operations
2. Clean up created data
3. Restore previous state
4. Must be complete and tested

### Dependencies

```json
{
  "migrations": {
    "000001_initial": {
      "depends_on": [],
      "affects_tables": ["table1", "table2"],
      "risk_level": "low|medium|high",
      "estimated_duration": "30s|5m",
      "rollback_safe": true
    }
  }
}
```

### Risk Levels

- **low**: No data loss, quick rollback, minimal downtime
- **medium**: Some data loss possible, requires backup, moderate downtime
- **high**: Significant data loss possible, requires careful testing, extended downtime

### Metadata Fields

```json
{
  "000001_initial": {
    "metadata": {
      "author": "Your Name",
      "jira_ticket": "TICKET-123",
      "requires_downtime": false,
      "notes": "Initial schema for objects-service",
      "rollback_plan": "Simple: DROP TABLE IF EXISTS"
    }
  }
}
```

## Environment-Specific Migrations

### Development

```json
{
  "development": {
    "description": "Development environment",
    "migrations": [
      "000001_initial",
      "000002_dev_tax_test_data"
    ]
  }
}
```

### Production

```json
{
  "production": {
    "description": "Production environment",
    "migrations": [
      "000001_initial"
    ]
  }
}
```

### Staging

```json
{
  "staging": {
    "description": "Staging environment",
    "migrations": [
      "000001_initial"
    ]
  }
}
```

## Integration Checklist

### Phase 1: Setup

- [ ] Create `migrations/dependencies.json`
- [ ] Create `migrations/environments.json`
- [ ] Initialize orchestrator tracking: `make db-migrate-init SERVICE_NAME=objects-service`
- [ ] Verify `service_schema.migration_executions` table created

### Phase 2: Create Migrations

- [ ] Create `000001_initial.up.sql` with base schema
- [ ] Create `000001_initial.down.sql` with rollback
- [ ] Create `development/000002_dev_tax_test_data.up.sql`
- [ ] Create `development/000002_dev_tax_test_data.down.sql`

### Phase 3: Testing

- [ ] Test migrations up: `make db-migrate-up SERVICE_NAME=objects-service`
- [ ] Test migrations down: `make db-migrate-down SERVICE_NAME=objects-service`
- [ ] Validate: `make db-migrate-validate SERVICE_NAME=objects-service`
- [ ] Verify tracking table populated

### Phase 4: Service Integration

**Option A - CLI Only**:
- [ ] Update README with orchestrator commands
- [ ] Add Makefile with orchestrator targets

**Option B - Go Library**:
- [ ] Import orchestrator into `cmd/main.go`
- [ ] Add migration runner function
- [ ] Add migration middleware (development only)

### Phase 5: Deployment

- [ ] Validate production migrations: `make db-migrate-validate SERVICE_NAME=objects-service`
- [ ] Backup database before deployment
- [ ] Run migrations in production
- [ ] Verify tracking table
- [ ] Monitor execution logs

## Orchestrator vs Manual Migrations

| Aspect | Manual | Orchestrator |
|--------|--------|------------|
| Dependency Order | Manual (error-prone) | Automatic topological sort |
| Environment Awareness | File-based | Database-level tracking |
| Rollback Safety | Manual steps | State-aware |
| Execution History | Logs only | Database tracking |
| Validation | Manual checks | Automated schema validation |
| Rollback Intelligence | Manual | Cross-migration awareness |
| Concurrency | Manual locking | Transaction-level locking |
| Monitoring | Manual logs | Real-time execution tracking |
| Alerting | Manual review | Automatic risk detection |

## Best Practices

### 1. Always Test in Development

```bash
# Test migrations locally
make db-migrate-up SERVICE_NAME=objects-service ENVIRONMENT=development

# Rollback and re-run
make db-migrate-down SERVICE_NAME=objects-service
make db-migrate-up SERVICE_NAME=objects-service
```

### 2. Use Dry-Run First

```bash
# Preview what will happen
make db-migrate-up SERVICE_NAME=objects-service DRY_RUN=true

# Check for conflicts
make db-migrate-conflicts SERVICE_NAME=objects-service
```

### 3. Validate Before Production

```bash
# Full validation
make db-migrate-validate SERVICE_NAME=objects-service

# Check for pending migrations
make db-migrate-status SERVICE_NAME=objects-service
```

### 4. Monitor Production Deployments

```bash
# Real-time monitoring
watch -n 1 'tail -f /var/log/objects-service/migrations.log'

# Check status
while true; do
    make db-migrate-status SERVICE_NAME=objects-service
    sleep 60
done
```

### 5. Rollback Plan

Before major deployments:
```bash
# Generate rollback plan
make db-migrate-rollback-plan SERVICE_NAME=objects-service

# Save to safe location
cp rollback-plan.json /backups/objects-service-$(date +%Y%m%d).json
```

## Troubleshooting

### Common Issues

**Issue**: Service not found by orchestrator

**Solution**: Ensure `SERVICE_NAME=objects-service` and `BASEPATH` are correct:
```bash
ls services/objects-service/migrations/  # Should see dependencies.json and environments.json
```

**Issue**: Database connection failed

**Solution**: Check `DATABASE_URL` environment variable and connectivity:
```bash
echo $DATABASE_URL

# Test connection
psql "$DATABASE_URL" -c "SELECT version()"
```

**Issue**: Migration stuck on "pending" status

**Solution**: Check for conflicts:
```bash
make db-migrate-conflicts SERVICE_NAME=objects-service
```

**Issue**: Rollback failed due to conflicts

**Solution**: Use intelligent rollback:
```bash
make db-migrate-down SERVICE_NAME=objects-service INTELLIGENT_ROLLBACK=true
```

**Issue**: Cross-service dependency not resolved

**Solution**: Check orchestration status:
```bash
make db-migrate-status SERVICE_NAME=all
```

## Makefile Integration

Optional but recommended for convenience:

```makefile
# Objects Service
.PHONY: help

help:
	@echo "Objects Service - Orchestrator Integration"
	@echo ""
	@echo "Available targets:"
	@echo "  migrate-init      - Initialize orchestrator tracking"
	@echo "  migrate-up        - Run migrations up"
	@echo "  migrate-down      - Roll back migrations"
	@echo "  migrate-status     - Show migration status"
	@echo "  migrate-validate   - Validate migrations"
	@echo "  migrate-list       - List all migrations"
	@echo  migrate-history   - Show execution history"
	@echo ""

.PHONY: migrate-init migrate-up migrate-down migrate-status migrate-validate migrate-list

migrate-init:
	cd ../migration-orchestrator && \
	make db-migrate-init SERVICE_NAME=objects-service

migrate-up:
	cd ../migration-orchestrator && \
	make db-migrate-up SERVICE_NAME=objects-service

migrate-down:
	cd ../migration-orchestrator && \
	make db-migrate-down SERVICE_NAME=objects-service

migrate-status:
	cd ../migration-orchestrator && \
	make db-migrate-status SERVICE_NAME=objects-service

migrate-validate:
	cd ../migration-orchestrator && \
		make db-migrate-validate SERVICE_NAME=objects-service

migrate-list:
	cd ../migration-orchestrator && \
		make db-migrate-list SERVICE_NAME=objects-service

migrate-history:
	cd ../migration-orchestrator && \
		make db-migrate-history SERVICE_NAME=objects-service LIMIT=20
```

## Quick Reference

### Orchestrator Commands

| Command | Description |
|---------|-------------|
| `db-migrate-init` | Initialize tracking for a service |
| `db-migrate-up` | Run migrations up (supports COUNT, FORCE, STEP, DRY_RUN) |
| `db-migrate-down` | Rollback migrations (supports STEPS, INTELLIGENT_ROLLBACK) |
| `db-migrate-status` | Show migration status for all services |
| `db-migrate-validate` | Validate migration integrity |
| `db-migrate-list` | List all migrations for a service |
| `db-migrate-history` | Show execution history |
| `db-migrate-conflicts` | Check for migration conflicts |

### Environment Variables

| Variable | Description | Example |
|---------|-------------|--------|
| `SERVICE_NAME` | Service identifier | `objects-service` |
| `BASEPATH` | Path to migrations | `services/objects-service/migrations` |
| `DATABASE_URL` | Database connection | `postgresql://...` |
| `APP_ENV` | Environment | `development|staging|production` |
| `DRY_RUN` | Preview execution | `true|false` |
| `FORCE` | Force rerun | `true|false` |

## Next Steps

1. Create `migrations/dependencies.json`
2. Create `migrations/environments.json`
3. Create `000001_initial.up.sql` (Phase 1)
4. Initialize orchestrator: `make db-migrate-init SERVICE_NAME=objects-service`
5. Test migrations locally
6. Update `phase-01-migrations.md` with orchestrator integration
7. Update `progress.md` when phase complete

See [phase-01-migrations.md](phase-01-migrations.md) for detailed migration steps.
