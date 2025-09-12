# Boilerplate CLI

A command-line utility for automating business-logic related operations in the service boilerplate.

## Installation

```bash
cd cli
go build -o boilerplate-cli .
```

## Configuration

The CLI uses a YAML configuration file. You can:

1. Use the default config file: `boilerplate-cli.yaml`
2. Specify a custom config file: `boilerplate-cli --config /path/to/config.yaml`
3. Use environment variables with prefix `BOILERPLATE_CLI_`

### Configuration File Example

```yaml
# Environment settings
environment: development

# Service configuration
gateway_url: "http://localhost:8080"
service_urls:
  user-service: "http://localhost:8081"
  api-gateway: "http://localhost:8080"
default_port: 8080
timeout: 30
retry_attempts: 3

# API configuration
base_url: "http://localhost:8080"
version: "v1"
api_timeout: 30

# Database configuration
db_host: "localhost"
db_port: 5432
db_user: "postgres"
db_password: "postgres"
db_database: "service_db"
db_ssl_mode: "disable"

# Logging configuration
log_level: "info"
log_format: "text"
log_output: "stdout"
```

## Usage

### Global Flags

- `--config, -c`: Config file path
- `--env, -e`: Environment (development/production)
- `--verbose, -v`: Verbose output
- `--json`: JSON output format

### Services Commands

#### List Services
```bash
boilerplate-cli services list
boilerplate-cli services list --json
```

#### Check Service Status
```bash
boilerplate-cli services status
boilerplate-cli services status user-service
boilerplate-cli services status --json
```

#### Direct API Call
```bash
boilerplate-cli services call user-service GET /api/v1/users
boilerplate-cli services call user-service POST /api/v1/users --body '{"name":"John"}'
boilerplate-cli services call user-service GET /api/v1/users --header "Authorization=Bearer token"
```

#### Gateway Proxy
```bash
boilerplate-cli services proxy user-service GET /api/v1/users
```

### Data Commands

#### Seed Data
```bash
boilerplate-cli data seed user-service test-users
```

#### Export Data
```bash
boilerplate-cli data export user-service users
```

#### Validate Data
```bash
boilerplate-cli data validate user-service
```

#### Clean Data
```bash
boilerplate-cli data cleanup user-service
```

### Business Operations

#### User Operations
```bash
boilerplate-cli ops user create john@example.com "John Doe"
boilerplate-cli ops user list
boilerplate-cli ops user update 123 '{"name":"Jane Doe"}'
boilerplate-cli ops user delete 123
```

#### Workflow Execution
```bash
boilerplate-cli ops workflow user-onboarding
```

### Development Commands

#### Scaffold CRUD
```bash
boilerplate-cli dev scaffold Product product-service
```

#### Run Tests
```bash
boilerplate-cli dev test user-service
boilerplate-cli dev test user-service --watch
```

#### Database Migrations
```bash
boilerplate-cli dev migrate user-service up
boilerplate-cli dev migrate user-service down
```

#### Stream Logs
```bash
boilerplate-cli dev logs user-service
boilerplate-cli dev logs user-service --follow
```

### Health Commands

#### Comprehensive Health Check
```bash
boilerplate-cli health check
```

#### Individual Health Checks
```bash
boilerplate-cli health services
boilerplate-cli health database
boilerplate-cli health dependencies
```

## Examples

### Development Workflow

```bash
# Check all services are running
boilerplate-cli health check

# List available services
boilerplate-cli services list

# Create a test user
boilerplate-cli ops user create test@example.com "Test User"

# Check user was created
boilerplate-cli services call user-service GET /api/v1/users

# Run tests
boilerplate-cli dev test user-service

# Clean up test data
boilerplate-cli data cleanup user-service
```

### Production Operations

```bash
# Set production environment
export BOILERPLATE_CLI_ENVIRONMENT=production

# Check production services
boilerplate-cli health check

# Get service status in JSON format
boilerplate-cli services status --json

# Execute business workflow
boilerplate-cli ops workflow data-migration
```

## Architecture

The CLI is built with a modular architecture:

- **cmd/**: CLI commands and subcommands
- **internal/config/**: Configuration management
- **internal/client/**: HTTP API client with retry logic
- **internal/discovery/**: Service discovery and registry
- **pkg/**: Public packages and utilities

## Features

- ✅ **Service Discovery**: Automatically discovers and monitors services
- ✅ **Health Monitoring**: Real-time health checks for all services
- ✅ **Configuration Management**: Flexible config with environment overrides
- ✅ **HTTP Client**: Production-ready API client with retry logic
- ✅ **Command Structure**: Intuitive CLI with help and examples
- ✅ **Output Formats**: Both human-readable and JSON formats
- ✅ **Error Handling**: Comprehensive error reporting and logging
- ✅ **User Operations**: Complete CRUD operations for user management
- ✅ **Data Operations**: Seed, export, validate, and cleanup data
- ✅ **Workflow Orchestration**: Predefined business workflows
- 🚧 Advanced monitoring and logging (Phase 3)

## Contributing

1. Commands should follow the established patterns
2. Add appropriate error handling and logging
3. Include help text and examples
4. Test commands with both running and stopped services
5. Update this README when adding new features