# CLI Utility Plan for Business Logic Operations

## Overview

This document outlines the design and implementation plan for a CLI utility that will serve as a foundation for automating business-logic related operations in the service boilerplate. The CLI will complement existing tools (create-service.sh, Makefile) by focusing on runtime operations and business logic automation.

## Core Purpose

Create a CLI tool that provides:
- Service interaction and orchestration
- Business logic operation automation
- Development and operational workflows
- Foundation for future service API integrations

## CLI Command Structure

```
boilerplate-cli [global-flags] <command> [subcommand] [flags]

Commands:
  services     - Service management and interaction
  data         - Data operations and management
  ops          - Business operations and workflows
  dev          - Development utilities
  health       - Health checks and monitoring

Global Flags:
  --env, -e     - Environment (development/production)
  --config, -c  - Config file path
  --verbose, -v - Verbose output
  --json        - JSON output format
```

## Detailed Command Design

### services - Service Management

```
services list                    - List all registered services
services status <service>        - Get service status and health
services call <service> <endpoint> - Direct API call to service
services proxy <service> <method> <path> - Proxy request through gateway
```

### data - Data Operations

```
data seed <service> <fixture>     - Seed service with test data
data export <service> <table>     - Export data from service
data validate <service>          - Validate service data integrity
data cleanup <service>           - Clean test data from service
```

### ops - Business Operations

```
ops user create <email> <name>    - Create user via user-service API
ops user list [--filter]         - List users with filtering
ops user update <id> <data>      - Update user via API
ops user delete <id>             - Delete user via API
ops workflow <name>              - Execute predefined business workflow
```

### dev - Development Utilities

```
dev scaffold <entity> <service>   - Generate CRUD operations for entity
dev test <service> [--watch]      - Run service tests with options
dev migrate <service> <action>    - Database migration operations
dev logs <service> [--follow]     - Stream service logs
```

### health - Monitoring

```
health check                     - Comprehensive health check all services
health services                  - Check individual service health
health database                  - Database connectivity and health
health dependencies              - Check service dependencies
```

## Architecture Design

### Service Integration Layer

- **API Client**: HTTP client with service discovery
- **Gateway Proxy**: Route through API gateway when available
- **Direct Service Calls**: Bypass gateway for development
- **Service Registry**: Dynamic service discovery from docker-compose/env

### Configuration Management

- **Multi-source Config**: Environment variables, config files, CLI flags
- **Service-specific Configs**: Per-service configuration overrides
- **Environment Profiles**: Development, staging, production profiles

### Business Logic Foundation

- **Operation Templates**: Predefined operation workflows
- **Data Transformers**: Request/response transformation
- **Validation Layer**: Input validation and error handling
- **Retry Logic**: Automatic retry for transient failures

## Integration with Existing Boilerplate

### Complementary to Existing Tools

- **create-service.sh**: Service creation (keep separate)
- **Makefile**: Build/deployment (keep separate)
- **CLI**: Runtime operations and business logic

### Shared Components

- **Common Libraries**: Reuse logging, config, database packages
- **Service Registry**: Integrate with API gateway service discovery
- **Environment Config**: Use same .env and config.yaml patterns

## Implementation Approach

### Phase 1: Core Infrastructure

1. CLI framework setup (Cobra/viper)
2. Service discovery and API client
3. Basic service interaction commands
4. Configuration management

### Phase 2: Business Operations

1. User service API integration
2. Data operations (seed, export, validate)
3. Workflow orchestration
4. Error handling and logging

### Phase 3: Advanced Features

1. Multi-service operations
2. Custom workflow definitions
3. Monitoring and observability
4. Plugin system for extensibility

## Key Benefits

- **Unified Interface**: Single tool for all service operations
- **Business Logic Focus**: Operations vs infrastructure management
- **Extensible**: Foundation for future service APIs
- **Developer Experience**: Streamlined workflows and automation
- **Production Ready**: Robust error handling and monitoring

## Next Steps

1. **Finalize command structure** based on specific business requirements
2. **Define service API contracts** for consistent integration
3. **Create operation templates** for common business workflows
4. **Implement core CLI framework** with service discovery
5. **Add comprehensive testing** and documentation

## Technical Requirements

### Dependencies

- Go 1.23+
- Cobra CLI framework
- Viper configuration management
- HTTP client with retry logic
- JSON/YAML parsing
- Service discovery mechanism

### Project Structure

```
cli/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   ├── services.go        # Service management commands
│   ├── data.go            # Data operations commands
│   ├── ops.go             # Business operations commands
│   ├── dev.go             # Development utilities
│   └── health.go          # Health monitoring commands
├── internal/
│   ├── client/            # API client implementations
│   ├── config/            # Configuration management
│   ├── discovery/         # Service discovery
│   └── workflows/         # Business workflow definitions
├── pkg/
│   ├── api/               # API client interfaces
│   ├── models/            # Data models
│   └── utils/             # Utility functions
├── config.yaml            # CLI configuration
├── main.go                # Application entry point
└── go.mod
```

## Success Criteria

- [ ] CLI provides unified interface for service operations
- [ ] Business logic operations are easily extensible
- [ ] Integration with existing boilerplate is seamless
- [ ] Error handling and logging are comprehensive
- [ ] Documentation and examples are complete
- [ ] Testing coverage meets project standards

---

*This CLI will serve as the operational foundation for the boilerplate, enabling automated business logic operations while maintaining separation from infrastructure management tools.*