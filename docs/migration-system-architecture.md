# Migration System Architecture

## Overview

This document describes the current migration system architecture - what was actually implemented vs what was originally planned.

## What Was Implemented

### Current Components

| Component | Status | Description |
|-----------|--------|-------------|
| golang-migrate CLI | ✅ | Used for actual migration execution |
| migrate-wrapper | ✅ | Simple Go CLI wrapper around golang-migrate |
| environments.json | ✅ | Per-service configuration for environment directories |
| Schema-per-service | ✅ | Each service gets its own PostgreSQL schema |
| Environment directories | ✅ | development/, staging/, production/ per service |
| Sequential numbering | ✅ | Migrations numbered sequentially per environment |

### What Was NOT Implemented

The following features were planned but **never implemented**:

| Feature | Description |
|---------|-------------|
| `dependencies.json` | Dependency tracking between migrations |
| `migration_executions` table | Enhanced tracking with execution history |
| Dependency resolution | Automatic ordering based on migration dependencies |
| Risk assessment | Warnings for high-impact changes |
| Approval workflows | Production approval process |
| Backup integration | Automatic backups before migrations |
| Intelligent rollback | Rollback with dependency checking |

## Current Architecture

### Components

```
migrate-wrapper (Go CLI)
    │
    ├── Loads environments.json
    ├── Determines migration directory (dev/staging/prod)
    └── Executes golang-migrate CLI
            │
            └── golang-migrate
                    │
                    └── PostgreSQL (schema-per-service)
```

### Database Schema

Each service has its own schema with standard golang-migrate tracking:

```sql
-- Example: user_service schema
CREATE SCHEMA user_service;

-- golang-migrate creates this automatically
CREATE TABLE user_service.schema_migrations (
    version bigint NOT NULL PRIMARY KEY,
    dirty boolean NOT NULL DEFAULT false
);
```

### Configuration

Each service has `environments.json` in its migrations directory:

```json
{
  "environments": {
    "development": {
      "migrations": "development",
      "config": { "allow_destructive_operations": true }
    },
    "staging": {
      "migrations": "staging",
      "config": { "require_approval": true }
    },
    "production": {
      "migrations": "production",
      "config": { "require_approval": true, "backup_required": true }
    }
  },
  "current_environment": "development"
}
```

### Migration Flow

```
1. db-migrate-init (once per service)
   └── Creates service schema + schema_migrations table

2. db-migrate-up
   ├── Load environments.json
   ├── Get migration directory (e.g., "development")
   └── Execute: migrate -path .../development -database ... up

3. db-migrate-down
   └── Execute: migrate -path .../development -database ... down N
```

## Docker Integration

The migration wrapper runs in Docker:

```dockerfile
# From Dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -o migrate-wrapper .

FROM alpine:latest
COPY --from=builder /build/migrate-wrapper /usr/local/bin/
COPY --from=migrate/migrate:latest /migrate /usr/local/bin/
ENTRYPOINT ["migrate-wrapper"]
```

Usage via Makefile:
```bash
make db-migrate-init SERVICE_NAME=user-service
make db-migrate-up SERVICE_NAME=user-service
make db-migrate-down SERVICE_NAME=user-service
```

## Key Differences from Original Plan

| Original Plan | Actual Implementation |
|---------------|------------------------|
| Custom dependency resolution | No dependency tracking - use sequential numbering |
| migration_executions table | Not implemented - uses basic schema_migrations |
| Risk assessment | Not implemented |
| Approval workflows | Not implemented |
| Backup integration | Not implemented |

## Future Considerations

If more advanced features are needed, consider:

1. **Add dependencies.json back** - If migration ordering becomes complex
2. **Enhance tracking** - Add migration_executions table if more history needed
3. **Risk assessment** - Add pre-migration checks for large table changes
4. **Approval workflow** - Add manual approval step for production

---

**Last Updated**: April 2026
**Version**: 2.0