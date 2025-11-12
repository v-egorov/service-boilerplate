# Migration System Architecture & Implementation Plan

## Overview

This document outlines the architecture and implementation plan for a sophisticated database migration system that enhances the current golang-migrate setup with advanced features including environment-specific migrations, dependency management, risk assessment, and comprehensive audit trails.

## Current State Analysis

### Existing Components
- **Migration Tool**: golang-migrate (docker image: `migrate/migrate:latest`)
- **Configuration**: `environments.json` and `dependencies.json` per service
- **Execution**: Makefile targets calling golang-migrate directly
- **Database Architecture**: Single database with schema-per-service approach
  - Migration tracking: `public.auth_service_schema_migrations`, `public.user_service_schema_migrations`
  - Service data: `auth_service.*`, `user_service.*` tables
- **Tracking**: Basic version numbers only (golang-migrate's schema_migrations tables)

### Limitations
- No environment-specific migration support
- No dependency resolution or risk assessment
- Minimal audit trail (version + dirty flag only)
- No rollback intelligence or diagnostics
- Manual environment filtering required

## Proposed Architecture

### Core Components

#### 1. Migration Orchestrator (CLI Tool)
**Purpose**: Intelligent migration management with advanced features
**Technology**: Go CLI application using cobra/viper for configuration
**Responsibilities**:
- Parse `environments.json` and `dependencies.json`
- Environment-specific migration filtering
- Dependency resolution and validation
- Risk assessment and safety checks
- Enhanced tracking and audit logging
- Intelligent rollback operations

#### 2. Enhanced Tracking Database Schema
**Purpose**: Comprehensive migration execution history
**Location**: Dedicated table in each service schema
**Features**:
- Full execution history with timestamps
- Environment tracking
- Dependency relationships
- Risk levels and safety flags
- Error logging and diagnostics
- Rollback history

#### 3. JSON Configuration System
**Existing**: `environments.json`, `dependencies.json`
**Enhancements**:
- Environment-specific migration definitions
- Dependency graphs with risk assessment
- Rollback safety indicators
- Migration metadata (duration estimates, affects_tables)

#### 4. Docker Integration
**Current**: golang-migrate in separate container
**Enhanced**:
- Custom Docker image with orchestrator + golang-migrate
- Environment variable passing
- Volume mounting for configurations

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    Migration Orchestrator                   │
│  ┌─────────────────────────────────────────────────────────┐│
│  │  CLI Tool (Go + Cobra)                                  ││
│  │                                                         ││
│  │  • Command Parser (up, down, status, validate)          ││
│  │  • Configuration Loader (JSON configs)                  ││
│  │  • Dependency Resolver                                  ││
│  │  • Risk Assessor                                        ││
│  │  • Execution Coordinator                                ││
│  └─────────────────────────────────────────────────────────┘│
│                                                             │
│  Orchestrates:                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │  golang-migrate (Execution Engine)                      ││
│  │                                                         │  │
│  │  • SQL Execution                                        │  │
│  │  • Version Tracking                                     │  │
│  │  • Database Locking                                     │  │
│  │  • Transaction Management                               │  │
│  └─────────────────────────────────────────────────────────┘  │
└─────────────────────┼─────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                 Enhanced Tracking                          │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │  migration_executions Table                           │  │
│  │                                                         │  │
│  │  • Execution History                                   │  │
│  │  • Environment Tracking                                │  │
│  │  • Dependency Metadata                                  │  │
│  │  • Risk Assessment                                      │  │
│  │  • Rollback History                                     │  │
│  └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                      ▲
                      │
┌─────────────────────────────────────────────────────────────┐
│              JSON Configuration                            │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │  environments.json + dependencies.json                │  │
│  │                                                         │  │
│  │  • Environment Definitions                             │  │
│  │  • Migration Dependencies                              │  │
│  │  • Risk Levels & Safety Flags                          │  │
│  │  • Migration Metadata                                  │  │
│  └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Implementation Plan

### Phase 1: Architecture Design & Planning (Current)
- [x] Analyze current migration system limitations
- [x] Design enhanced architecture with orchestrator
- [x] Define enhanced tracking schema
- [x] Create implementation roadmap
- [x] Document architecture and plan

### Phase 2: Core Orchestrator Development
**Duration**: 2-3 weeks
**Deliverables**:
- Go CLI application with cobra framework
- JSON configuration parsing
- Basic migration execution coordination
- Enhanced tracking database schema
- Docker image with orchestrator + golang-migrate

**Tasks**:
1. Set up Go project structure with cobra CLI
2. Implement configuration loading (environments.json, dependencies.json)
3. Create migration execution coordinator
4. Implement enhanced tracking table schema (schema-per-service)
5. Add automatic schema-qualified table creation
6. Build custom Docker image
7. Basic integration testing with schema handling

### Phase 3: Advanced Features Implementation
**Duration**: 2-3 weeks
**Deliverables**:
- Dependency resolution engine
- Risk assessment system
- Intelligent rollback operations
- Comprehensive diagnostics
- Environment validation

**Tasks**:
1. Implement dependency graph resolution
2. Add risk level checking and warnings
3. Create rollback intelligence (dependency-aware)
4. Implement comprehensive status reporting
5. Add migration validation commands
6. Environment-specific operation handling

### Phase 4: Integration & Testing
**Duration**: 1-2 weeks
**Deliverables**:
- Makefile integration
- CI/CD pipeline updates
- Comprehensive testing suite
- Documentation updates
- Migration from current system

**Tasks**:
1. Update Makefile targets to use orchestrator
2. Create migration testing framework
3. Implement CI/CD integration
4. Write comprehensive documentation
5. Perform migration testing across environments
6. Update existing services to use new system

### Phase 5: Production Deployment & Monitoring
**Duration**: 1 week
**Deliverables**:
- Production deployment
- Monitoring and alerting
- Performance optimization
- Operational runbooks

**Tasks**:
1. Deploy to staging environment
2. Performance testing and optimization
3. Monitoring dashboard setup
4. Operational documentation
5. Production deployment with rollback plans

## Technical Specifications

### Migration Orchestrator CLI

#### Command Structure
```bash
migration-orchestrator [global-flags] <command> [command-flags] [args]

Commands:
  up          Apply migrations
  down        Rollback migrations
  status      Show migration status
  validate    Validate migration configurations
  plan        Show migration execution plan
  rollback    Intelligent rollback operations

Global Flags:
  --service string       Service name (required)
  --environment string   Environment (development/staging/production)
  --config string        Configuration directory
  --dry-run              Show what would be executed
  --verbose              Enable verbose output
```

#### Example Usage
```bash
# Apply all migrations for user-service in development
migration-orchestrator up --service user-service --environment development

# Show detailed status with dependencies
migration-orchestrator status --service user-service --verbose

# Validate migration configurations
migration-orchestrator validate --service user-service

# Intelligent rollback with dependency checking
migration-orchestrator rollback --migration 000004_settings --service user-service
```

### Enhanced Tracking Schema

#### Service-Specific migration_executions Tables
Following the schema-per-service architecture, enhanced tracking tables are created in each service schema:

```sql
-- For user-service
CREATE TABLE user_service.migration_executions (

-- For auth-service
CREATE TABLE auth_service.migration_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    migration_id VARCHAR(255) NOT NULL,
    service_name VARCHAR(100) NOT NULL,
    environment VARCHAR(50) NOT NULL,
    execution_status VARCHAR(20) NOT NULL CHECK (execution_status IN ('success', 'failed', 'rolled_back')),
    applied_at TIMESTAMP WITH TIME ZONE,
    rolled_back_at TIMESTAMP WITH TIME ZONE,
    execution_duration INTERVAL,
    error_message TEXT,
    checksum VARCHAR(64),
    applied_by VARCHAR(100),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(service_name, migration_id, environment)
);

-- Indexes for performance (schema-qualified)
CREATE INDEX idx_user_service_migration_executions_service_env ON user_service.migration_executions(service_name, environment);
CREATE INDEX idx_user_service_migration_executions_status ON user_service.migration_executions(execution_status);
CREATE INDEX idx_user_service_migration_executions_applied_at ON user_service.migration_executions(applied_at);

-- Similar indexes for auth_service.migration_executions
CREATE INDEX idx_auth_service_migration_executions_service_env ON auth_service.migration_executions(service_name, environment);
CREATE INDEX idx_auth_service_migration_executions_status ON auth_service.migration_executions(execution_status);
CREATE INDEX idx_auth_service_migration_executions_applied_at ON auth_service.migration_executions(applied_at);
```

#### Metadata JSONB Structure
```json
{
  "dependencies": ["000001", "000002"],
  "risk_level": "medium",
  "rollback_safe": true,
  "estimated_duration": "30s",
  "affects_tables": ["user_service.users", "user_service.profiles"],
  "migration_type": "table",
  "description": "Add user profiles functionality"
}
```

### Schema-Per-Service Architecture Considerations

#### Migration Tracking Strategy
- **golang-migrate tables**: Remain in `public` schema for compatibility
  - `public.user_service_schema_migrations`
  - `public.auth_service_schema_migrations`
- **Enhanced tracking tables**: Created in service-specific schemas
  - `user_service.migration_executions`
  - `auth_service.migration_executions`
- **Service data tables**: Already in service schemas
  - `user_service.users`, `user_service.user_profiles`, etc.
  - `auth_service.roles`, `auth_service.permissions`, etc.

#### Benefits of This Approach
- **Data isolation**: Service-specific audit trails stay with service data
- **Schema consistency**: All service-related tables in same schema
- **Cross-schema queries**: Enhanced tracking can join with service data
- **Migration compatibility**: golang-migrate continues working unchanged

#### Orchestrator Responsibilities
- Create service-specific `migration_executions` tables automatically
- Track executions in appropriate service schemas
- Maintain compatibility with existing golang-migrate tables
- Support cross-schema operations for reporting and analysis

### Docker Integration

#### Custom Docker Image
```dockerfile
FROM golang:1.23-alpine AS builder

# Build migration orchestrator
WORKDIR /build
COPY . .
RUN go build -o migration-orchestrator ./cmd

FROM alpine:latest

# Install dependencies
RUN apk add --no-cache postgresql-client jq

# Copy golang-migrate binary
COPY --from=migrate/migrate:latest /migrate /usr/local/bin/migrate

# Copy orchestrator
COPY --from=builder /build/migration-orchestrator /usr/local/bin/

# Set working directory
WORKDIR /workspace

# Default entrypoint
ENTRYPOINT ["migration-orchestrator"]
```

#### Usage in Makefile
```makefile
db-migrate:
	@echo "📈 Running migrations with orchestrator..."
	@docker run --rm --network service-boilerplate-network \
		--env-file $(ENV_FILE) \
		-e DATABASE_URL="postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(POSTGRES_NAME):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=$(DATABASE_SSL_MODE)" \
		-v $(PWD)/services:/services \
		migration-orchestrator:latest \
		up --service $(SERVICE_NAME) --environment $(APP_ENV)
```

## Alternative Architecture Options

### Option 1: Library-Based Orchestrator (Current Recommendation)
**Pros**: Standalone CLI tool, reusable across projects, comprehensive features
**Cons**: Additional complexity, separate deployment
**Implementation**: Go CLI application with cobra

### Option 2: Enhanced Makefile-Only Approach
**Pros**: Simpler, no new tools to maintain
**Cons**: Limited features, harder to test, less reusable
**Implementation**: Complex shell scripts in Makefile

### Option 3: Database-Native Solution
**Pros**: All logic in database, audit trails built-in
**Cons**: Harder to version control, language limitations
**Implementation**: PostgreSQL functions and tables

### Option 4: Kubernetes Operator
**Pros**: Cloud-native, automated, scalable
**Cons**: Overkill for current scale, complex deployment
**Implementation**: Kubernetes custom controller

## Risk Assessment & Mitigation

### Technical Risks
- **JSON Parsing Complexity**: Mitigated by comprehensive testing and validation
- **Dependency Resolution**: Implement cycle detection and validation
- **Database Locking**: Use golang-migrate's proven locking mechanism
- **Schema-Per-Service Complexity**: Proper cross-schema operations and permissions
- **Environment Conflicts**: Strict environment isolation and validation

### Operational Risks
- **Migration Failures**: Comprehensive error handling and rollback capabilities
- **Performance Impact**: Async execution and progress tracking
- **Team Adoption**: Comprehensive documentation and training
- **Rollback Complexity**: Intelligent dependency-aware rollback planning

## Success Metrics

### Functional Metrics
- ✅ Environment-specific migrations execute correctly
- ✅ Dependency resolution prevents invalid migration order
- ✅ Risk assessment provides appropriate warnings
- ✅ Rollback operations are dependency-aware
- ✅ Audit trail provides complete execution history

### Performance Metrics
- Migration execution time < 30 seconds for typical deployments
- Status queries return in < 2 seconds
- Memory usage < 100MB during execution
- No database locks held for > 5 minutes

### Quality Metrics
- 100% test coverage for orchestrator logic
- Zero data loss in rollback scenarios
- Comprehensive error messages and diagnostics
- Full compatibility with existing golang-migrate migrations

## Migration Path

### From Current System
1. **Phase 1**: Deploy orchestrator alongside existing system
2. **Phase 2**: Migrate one service at a time to new system
3. **Phase 3**: Update CI/CD pipelines
4. **Phase 4**: Deprecate old golang-migrate-only approach
5. **Phase 5**: Full production deployment

### Backward Compatibility
- Existing golang-migrate migrations continue to work
- Enhanced tracking is additive (doesn't break existing data)
- Gradual migration path with rollback capability

## Conclusion

The proposed migration orchestrator architecture provides a significant enhancement over the current golang-migrate setup by adding:

- **Intelligent environment handling** via JSON configurations
- **Dependency resolution and risk assessment** for safe deployments
- **Comprehensive audit trails** with execution history and diagnostics
- **Advanced rollback capabilities** with dependency awareness
- **Schema-per-service architecture support** with proper data isolation
- **Operational excellence** features for enterprise-grade database management

The **CLI-based orchestrator approach** (Option 1) is recommended as it provides the best balance of features, maintainability, and operational benefits while:

- **Leveraging golang-migrate's proven execution engine**
- **Respecting the existing schema-per-service database architecture**
- **Maintaining backward compatibility with existing migrations**
- **Providing enterprise-grade audit trails and operational features**

This architecture transforms the migration system from basic version tracking to sophisticated database change management suitable for enterprise production environments with multi-service, schema-isolated databases.
