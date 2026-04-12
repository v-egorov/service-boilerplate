# Database Migrations System

## Overview

The Database Migrations System uses golang-migrate via a Go wrapper (`migrate-wrapper`) to manage database schema changes across multiple services and environments. Each service has its own database schema with environment-specific migration directories.

## Key Features

- **Environment Awareness**: Different migrations for dev/staging/production via separate directories
- **Sequential Numbering**: Each environment has strictly sequential migration files
- **Schema Isolation**: Each service uses its own PostgreSQL schema
- **Simple CLI**: Direct golang-migrate CLI wrapper for reliable migrations

## Architecture

### Schema-per-Service Approach

```
service_db (PostgreSQL)
├── auth_service
│   ├── schema_migrations (golang-migrate tracking)
│   └── tables (users, roles, permissions, etc.)
├── user_service
│   ├── schema_migrations
│   └── tables (users, user_profiles, user_settings)
└── objects_service
    ├── schema_migrations
    └── tables (object_types, objects, etc.)
```

### Migration Structure

```
services/{service}/migrations/
├── environments.json          # Environment configuration
├── development/              # Development migrations (sequential)
│   ├── 000001_*.up.sql
│   ├── 000001_*.down.sql
│   └── ...
├── staging/                  # Staging migrations
│   ├── 000001_*.up.sql
│   └── ...
├── production/               # Production migrations
│   ├── 000001_*.up.sql
│   └── ...
└── docs/                     # Service-specific migration docs
```

## Quick Start

### 1. Generate a New Migration

```bash
# Generate a table migration (specify service)
make db-migrate-generate NAME=add_user_profiles TYPE=table SERVICE_NAME=user-service
```

### 2. Run Migrations

**Important:** Always run in sequence - init first, then up.

```bash
# Step 1: Initialize tracking (run once per service)
make db-migrate-init SERVICE_NAME=user-service

# Step 2: Apply all pending migrations
make db-migrate-up SERVICE_NAME=user-service

# Step 3: Check status
make db-migrate-status SERVICE_NAME=user-service
```

### 3. Rollback

```bash
# Rollback last migration
make db-migrate-down SERVICE_NAME=user-service
```

## Environment Configuration

Each service has an `environments.json` file that specifies which migration directory to use:

```json
{
  "environments": {
    "development": {
      "migrations": "development",
      "config": { "allow_destructive_operations": true }
    },
    "staging": {
      "migrations": "staging",
      "config": { "allow_destructive_operations": false }
    },
    "production": {
      "migrations": "production",
      "config": { "allow_destructive_operations": false }
    }
  },
  "current_environment": "development"
}
```

## Environment Differences

- **Development**: Includes test data migrations (more migrations)
- **Staging**: Excludes dev-only test data migrations
- **Production**: Excludes dev-only test data migrations

## Important Rules

1. **Sequential numbering**: Migrations must be numbered sequentially within each environment (001, 002, 003...)
2. **Always create both up and down**: Each migration needs both `.up.sql` and `.down.sql` files
3. **Never apply via psql**: Always use the migration wrapper
4. **Run init before up**: `db-migrate-init` must be run once before `db-migrate-up`

## Documentation

- **[Getting Started](./getting-started.md)**: Basic setup and first migration
- **[Best Practices](./guide.md)**: Guidelines and real-world examples
- **[Troubleshooting](./troubleshooting.md)**: Common issues and solutions
- **[Migration Types](./migration-types.md)**: Different migration categories
- **[Advanced Features](./advanced-features.md)**: Environment-specific migrations

---

**Version**: 2.0.0
**Last Updated**: April 2026
**Compatibility**: PostgreSQL 15+, golang-migrate v4.19.1