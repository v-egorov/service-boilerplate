# Boilerplate CLI Utility - Comprehensive Documentation

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Installation](#installation)
4. [Configuration](#configuration)
5. [Command Reference](#command-reference)
6. [API Reference](#api-reference)
7. [Examples](#examples)
8. [Troubleshooting](#troubleshooting)
9. [Contributing](#contributing)
10. [Changelog](#changelog)

## Overview

The Boilerplate CLI is a comprehensive command-line utility designed to automate business-logic related operations in the service boilerplate ecosystem. It provides a unified interface for service interaction, data management, workflow orchestration, and system monitoring.

### Key Features

- **🔍 Service Discovery**: Automatic discovery and health monitoring of services
- **📊 Advanced Monitoring**: Comprehensive health checks with performance metrics
- **🔄 Multi-Service Operations**: Coordinated operations across multiple services
- **⚙️ Business Logic Automation**: User management, data operations, and workflows
- **📈 Performance Monitoring**: Response time tracking and uptime analytics
- **🎯 Custom Workflows**: User-defined workflow creation and execution
- **🔧 Extensible Architecture**: Modular design for future enhancements

### Use Cases

- **Development**: Rapid service testing and data seeding
- **Operations**: System health monitoring and maintenance
- **CI/CD**: Automated testing and deployment workflows
- **Business Operations**: User management and data synchronization
- **Monitoring**: Real-time health checks and performance analysis

## Architecture

### Core Components

```
boilerplate-cli/
├── cmd/                    # CLI command definitions
│   ├── root.go            # Root command and global configuration
│   ├── services.go        # Service management commands
│   ├── data.go            # Data operations commands
│   ├── ops.go             # Business operations commands
│   ├── dev.go             # Development utilities
│   └── health.go          # Health monitoring commands
├── internal/               # Internal packages
│   ├── client/            # HTTP API client with retry logic
│   ├── config/            # Configuration management
│   ├── discovery/         # Service discovery mechanisms
│   ├── monitoring/        # Advanced monitoring and metrics
│   └── workflows/         # Workflow orchestration
│       ├── workflows.go   # Single-service workflows
│       ├── multi_service.go # Multi-service operations
│       └── custom.go      # Custom workflow management
├── pkg/                   # Public packages
│   └── utils/             # Shared utilities
│       ├── errors.go      # Error handling utilities
│       └── logging.go     # Logging utilities
├── boilerplate-cli.yaml   # Default configuration file
├── main.go                # Application entry point
└── go.mod                 # Go module definition
```

### Design Principles

1. **Modularity**: Each feature is implemented as a separate package
2. **Extensibility**: Plugin architecture for adding new commands
3. **Configuration-Driven**: Environment-specific behavior via config files
4. **Error Resilience**: Comprehensive error handling and recovery
5. **Performance**: Efficient API calls with connection pooling and caching
6. **Observability**: Detailed logging and metrics collection

### Data Flow

```
User Input → Command Parser → Configuration → Service Discovery → API Client → Service → Response → Output Formatter
```

## Installation

### Prerequisites

- Go 1.23 or later
- Access to the service boilerplate environment
- Network connectivity to target services

### Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd service-boilerplate/cli

# Build the CLI
go build -o boilerplate-cli .

# Verify installation
./boilerplate-cli --version
```

### Development Setup

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Build with race detection
go build -race -o boilerplate-cli .

# Install for system-wide use (optional)
sudo cp boilerplate-cli /usr/local/bin/
```

### Docker Integration

```bash
# Build Docker image
docker build -t boilerplate-cli .

# Run in container
docker run --network service-boilerplate-network boilerplate-cli health check
```

## Configuration

### Configuration Sources

The CLI supports multiple configuration sources in order of precedence:

1. **Command-line flags** (highest precedence)
2. **Environment variables** (prefix: `BOILERPLATE_CLI_`)
3. **Configuration file** (`boilerplate-cli.yaml` or custom path)
4. **Default values** (lowest precedence)

### Configuration File

Create a `boilerplate-cli.yaml` file in your working directory or home directory:

```yaml
# Environment settings
environment: development

# Service configuration
gateway_url: "http://localhost:8080"
service_urls:
  user-service: "http://localhost:8081"
  api-gateway: "http://localhost:8080"
  product-service: "http://localhost:8082"
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

# Custom workflows file
workflows_file: "custom-workflows.json"
```

### Environment Variables

```bash
# Set environment
export BOILERPLATE_CLI_ENVIRONMENT=production

# Configure service URLs
export BOILERPLATE_CLI_GATEWAY_URL="https://api.production.com"
export BOILERPLATE_CLI_SERVICE_URLS_USER_SERVICE="https://user.production.com"

# Database configuration
export BOILERPLATE_CLI_DB_HOST="prod-db.example.com"
export BOILERPLATE_CLI_DB_PASSWORD="secure-password"

# Logging
export BOILERPLATE_CLI_LOG_LEVEL="debug"
```

### Command-Line Configuration

```bash
# Use custom config file
boilerplate-cli --config /path/to/custom-config.yaml services list

# Override environment
boilerplate-cli --env production health check

# Enable verbose logging
boilerplate-cli --verbose services status user-service

# JSON output
boilerplate-cli --json health check
```

## Command Reference

### Global Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--config` | `-c` | Configuration file path | `$HOME/.boilerplate-cli.yaml` |
| `--env` | `-e` | Environment (development/production) | `development` |
| `--verbose` | `-v` | Enable verbose logging | `false` |
| `--json` | | Output in JSON format | `false` |
| `--help` | `-h` | Show help information | |

### Services Commands

#### `services list`

List all discovered services with their health status.

```bash
boilerplate-cli services list
boilerplate-cli services list --json
```

**Output:**
```
Available Services:
==================
✅ api-gateway: http://localhost:8080 (healthy)
✅ user-service: http://localhost:8081 (healthy)
❌ product-service: http://localhost:8082 (unhealthy)
```

#### `services status [service]`

Get detailed status information for a specific service or all services.

```bash
boilerplate-cli services status
boilerplate-cli services status user-service
boilerplate-cli services status --json
```

**Output:**
```
Service: user-service
   Status: ✅ healthy
   Response Time: 45ms
   Uptime: 99.5%
   Total Requests: 150
   Errors: 2
   Last Checked: 14:30:25
```

#### `services call <service> <method> <endpoint>`

Make a direct API call to a service.

```bash
boilerplate-cli services call user-service GET /api/v1/users
boilerplate-cli services call user-service POST /api/v1/users --body '{"email":"user@example.com","first_name":"John","last_name":"Doe"}'
boilerplate-cli services call user-service GET /api/v1/users --header "Authorization=Bearer token"
```

**Flags:**
- `--body`, `-b`: Request body (JSON string)
- `--header`, `-H`: Request headers (key=value pairs)

#### `services proxy <service> <method> <path>`

Proxy a request through the API gateway.

```bash
boilerplate-cli services proxy user-service GET /users
boilerplate-cli services proxy user-service POST /users --body '{"email":"user@example.com"}'
```

### Data Commands

#### `data seed <service> <fixture>`

Seed a service with predefined test data.

```bash
boilerplate-cli data seed user-service base
boilerplate-cli data seed user-service development
```

**Available fixtures:**
- `base`: Essential system data
- `development`: Test users and sample data

#### `data export <service> <table>`

Export data from a service table.

```bash
boilerplate-cli data export user-service users
boilerplate-cli data export user-service users --format json --output users.json
```

**Flags:**
- `--format`, `-f`: Export format (json, csv) - default: json
- `--output`, `-o`: Output file path

#### `data validate <service>`

Validate data integrity in a service.

```bash
boilerplate-cli data validate user-service
boilerplate-cli data validate user-service --json
```

**Output:**
```
🔍 Validating user-service data integrity...
✅ Found 150 users in user-service
📧 Valid emails: 148/150
👤 Valid names: 150/150
✅ All data validation checks passed!
```

#### `data cleanup <service>`

Clean up test data from a service.

```bash
boilerplate-cli data cleanup user-service --pattern example.com
boilerplate-cli data cleanup user-service --pattern test --force
```

**Flags:**
- `--pattern`, `-p`: Email pattern to match for cleanup
- `--force`, `-f`: Skip confirmation prompt

### Business Operations

#### User Operations

##### `ops user create <email> <first_name> <last_name>`

Create a new user.

```bash
boilerplate-cli ops user create john.doe@example.com "John" "Doe"
```

##### `ops user list`

List users with pagination.

```bash
boilerplate-cli ops user list
boilerplate-cli ops user list --limit 20 --offset 40
```

**Flags:**
- `--limit`, `-l`: Number of users to retrieve (default: 10)
- `--offset`, `-o`: Offset for pagination (default: 0)

##### `ops user update <id> <email> <first_name> <last_name>`

Update an existing user.

```bash
boilerplate-cli ops user update 123 john.doe@example.com "John" "Smith"
```

##### `ops user delete <id>`

Delete a user.

```bash
boilerplate-cli ops user delete 123
boilerplate-cli ops user delete 123 --force
```

**Flags:**
- `--force`, `-f`: Skip confirmation prompt

#### Workflow Operations

##### `ops workflow list`

List available predefined workflows.

```bash
boilerplate-cli ops workflow list
```

##### `ops workflow <name>`

Execute a predefined workflow.

```bash
boilerplate-cli ops workflow user-onboarding
boilerplate-cli ops workflow data-initialization
boilerplate-cli ops workflow system-health-check
```

**Available workflows:**
- `user-onboarding`: Complete user creation and verification
- `data-initialization`: Initialize system with base data
- `system-health-check`: Comprehensive health check

#### Multi-Service Operations

##### `ops multi list`

List available multi-service operations.

```bash
boilerplate-cli ops multi list
```

##### `ops multi execute <operation>`

Execute a multi-service operation.

```bash
boilerplate-cli ops multi execute system-bootstrap
boilerplate-cli ops multi execute data-synchronization
boilerplate-cli ops multi execute service-migration
```

**Available operations:**
- `system-bootstrap`: Initialize all services and verify connectivity
- `data-synchronization`: Synchronize data across services
- `service-migration`: Migrate data with compatibility verification

### Health Commands

#### `health check`

Comprehensive system health check.

```bash
boilerplate-cli health check
boilerplate-cli health check --verbose
boilerplate-cli health check --json
```

**Output:**
```
🏥 Running comprehensive system health check...
============================================
📊 System Overview:
  • Total Services: 3
  • Healthy Services: 3
  • Unhealthy Services: 0
  • Average Response Time: 45ms
  • System Uptime: 99.8%
  • Check Time: 14:30:25

📋 Service Details:
==================
✅ user-service
  Status: healthy
  Response Time: 42ms
  Uptime: 99.5%
  Total Requests: 150
  Errors: 2
  Last Checked: 14:30:25

🎉 System Health: EXCELLENT
```

#### `health services`

Check individual service health.

```bash
boilerplate-cli health services
boilerplate-cli health services --json
```

#### `health database`

Check database connectivity and health.

```bash
boilerplate-cli health database
boilerplate-cli health database --json
```

#### `health dependencies`

Analyze service dependencies.

```bash
boilerplate-cli health dependencies
boilerplate-cli health dependencies --json
```

### Development Commands

#### `dev scaffold <entity> <service>`

Generate CRUD operations for an entity (future implementation).

```bash
boilerplate-cli dev scaffold Product product-service
```

#### `dev test <service>`

Run tests for a service (future implementation).

```bash
boilerplate-cli dev test user-service
boilerplate-cli dev test user-service --watch
```

#### `dev migrate <service> <action>`

Run database migrations (future implementation).

```bash
boilerplate-cli dev migrate user-service up
boilerplate-cli dev migrate user-service down
```

#### `dev logs <service>`

Stream logs from a service (future implementation).

```bash
boilerplate-cli dev logs user-service
boilerplate-cli dev logs user-service --follow
```

## API Reference

### Service Discovery API

#### `discovery.ServiceRegistry`

```go
type ServiceRegistry struct {
    // Service registry implementation
}

func NewServiceRegistry(config *config.Config) *ServiceRegistry
func (sr *ServiceRegistry) DiscoverServices() ([]*ServiceInfo, error)
func (sr *ServiceRegistry) GetService(name string) (*ServiceInfo, error)
func (sr *ServiceRegistry) GetAllServices() []*ServiceInfo
func (sr *ServiceRegistry) IsServiceHealthy(name string) bool
func (sr *ServiceRegistry) GetServiceURL(name string) string
```

#### `discovery.ServiceInfo`

```go
type ServiceInfo struct {
    Name    string `json:"name"`
    URL     string `json:"url"`
    Status  string `json:"status"`
    Version string `json:"version,omitempty"`
}
```

### API Client API

#### `client.APIClient`

```go
type APIClient struct {
    // HTTP client with retry logic
}

func NewAPIClient(config *config.Config) *APIClient
func (c *APIClient) MakeRequest(req *Request) (*Response, error)
func (c *APIClient) Get(url string, headers map[string]string) (*Response, error)
func (c *APIClient) Post(url string, body interface{}, headers map[string]string) (*Response, error)
func (c *APIClient) Put(url string, body interface{}, headers map[string]string) (*Response, error)
func (c *APIClient) Delete(url string, headers map[string]string) (*Response, error)
func (c *APIClient) CallService(service, method, endpoint string, body interface{}, headers map[string]string) (*Response, error)
```

### Monitoring API

#### `monitoring.Monitor`

```go
type Monitor struct {
    // Advanced monitoring and metrics
}

func NewMonitor(config *config.Config, serviceReg *discovery.ServiceRegistry, apiClient *client.APIClient) *Monitor
func (m *Monitor) CheckServiceHealth(serviceName string) (*HealthStatus, error)
func (m *Monitor) GetSystemMetrics() (*SystemMetrics, error)
func (m *Monitor) GetServiceMetrics(serviceName string, limit int) ([]*ServiceMetrics, error)
func (m *Monitor) GetHealthStatus() map[string]*HealthStatus
```

### Workflow API

#### `workflows.WorkflowExecutor`

```go
type WorkflowExecutor struct {
    // Single-service workflow execution
}

func NewWorkflowExecutor(config *config.Config, apiClient *client.APIClient) *WorkflowExecutor
func (we *WorkflowExecutor) ExecuteWorkflow(name string) error
func (we *WorkflowExecutor) GetWorkflow(name string) (*Workflow, error)
func (we *WorkflowExecutor) ListWorkflows() map[string]*Workflow
```

#### `workflows.MultiServiceExecutor`

```go
type MultiServiceExecutor struct {
    // Multi-service operation execution
}

func NewMultiServiceExecutor(config *config.Config, apiClient *client.APIClient) *MultiServiceExecutor
func (mse *MultiServiceExecutor) ExecuteOperation(operation *MultiServiceOperation) (*MultiServiceResult, error)
func (mse *MultiServiceExecutor) GetPredefinedOperations() map[string]*MultiServiceOperation
```

## Examples

### Development Workflow

```bash
# 1. Check system health
boilerplate-cli health check

# 2. List available services
boilerplate-cli services list

# 3. Seed development data
boilerplate-cli data seed user-service development

# 4. Create test users
boilerplate-cli ops user create alice.developer@example.com "Alice" "Developer"
boilerplate-cli ops user create bob.manager@example.com "Bob" "Manager"

# 5. List users
boilerplate-cli ops user list --limit 10

# 6. Export user data
boilerplate-cli data export user-service users --output users_backup.json

# 7. Validate data integrity
boilerplate-cli data validate user-service

# 8. Run system health check
boilerplate-cli ops workflow system-health-check
```

### Production Operations

```bash
# Set production environment
export BOILERPLATE_CLI_ENVIRONMENT=production

# Comprehensive health check
boilerplate-cli health check --json | jq .

# Check service dependencies
boilerplate-cli health dependencies

# Execute system bootstrap
boilerplate-cli ops multi execute system-bootstrap

# Monitor system health
boilerplate-cli health services --json

# Data synchronization across services
boilerplate-cli ops multi execute data-synchronization
```

### CI/CD Integration

```bash
#!/bin/bash
# CI/CD pipeline script

# Health check before deployment
if ! boilerplate-cli health check --json | jq -e '.uptime_percentage > 95'; then
    echo "System health check failed"
    exit 1
fi

# Run data validation
boilerplate-cli data validate user-service

# Execute deployment workflow
boilerplate-cli ops multi execute service-migration

# Post-deployment health check
boilerplate-cli health check
```

### Monitoring Dashboard

```bash
#!/bin/bash
# Simple monitoring dashboard

while true; do
    clear
    echo "=== Service Boilerplate Monitoring Dashboard ==="
    echo "Timestamp: $(date)"
    echo

    boilerplate-cli health check

    echo
    echo "Press Ctrl+C to exit..."
    sleep 30
done
```

### Custom Workflow Creation

```bash
# Create a custom workflow file
cat > custom-workflows.json << EOF
{
  "user-registration": {
    "name": "User Registration",
    "description": "Complete user registration process",
    "steps": [
      {
        "name": "Create User Account",
        "description": "Create user in authentication service",
        "service": "user-service",
        "method": "POST",
        "endpoint": "/api/v1/users",
        "body": {
          "email": "{{email}}",
          "first_name": "{{first_name}}",
          "last_name": "{{last_name}}"
        },
        "required": true
      },
      {
        "name": "Send Welcome Email",
        "description": "Send welcome email to user",
        "service": "notification-service",
        "method": "POST",
        "endpoint": "/api/v1/emails/welcome",
        "body": {
          "email": "{{email}}",
          "name": "{{first_name}}"
        },
        "required": false
      }
    ]
  }
}
EOF

# Execute custom workflow
boilerplate-cli ops workflow user-registration
```

## Troubleshooting

### Common Issues

#### Services Not Found

**Problem:** `services list` shows no services or unhealthy status.

**Solutions:**
```bash
# Check if services are running
docker ps | grep service-boilerplate

# Verify service URLs in configuration
boilerplate-cli --verbose services list

# Test direct service connectivity
curl http://localhost:8081/health
```

#### API Call Failures

**Problem:** API calls return errors or timeouts.

**Solutions:**
```bash
# Check service health
boilerplate-cli health services

# Verify configuration
boilerplate-cli --verbose services status user-service

# Test with different timeout
boilerplate-cli services call user-service GET /health
```

#### Configuration Issues

**Problem:** CLI doesn't use expected configuration.

**Solutions:**
```bash
# Check configuration file location
ls -la ~/.boilerplate-cli.yaml

# Validate configuration
boilerplate-cli --verbose --config /path/to/config.yaml services list

# Use environment variables
export BOILERPLATE_CLI_ENVIRONMENT=development
boilerplate-cli services list
```

#### Permission Issues

**Problem:** Access denied to services or files.

**Solutions:**
```bash
# Check file permissions
ls -la boilerplate-cli

# Run with appropriate permissions
sudo ./boilerplate-cli services list

# Check Docker network permissions
docker network ls | grep service-boilerplate
```

### Debug Mode

Enable verbose logging for detailed troubleshooting:

```bash
boilerplate-cli --verbose services list
boilerplate-cli --verbose health check
```

### Logs and Diagnostics

```bash
# Check CLI logs
boilerplate-cli --verbose 2>&1 | tee cli-debug.log

# Test individual components
boilerplate-cli services status user-service --json
boilerplate-cli health database --json
```

### Performance Issues

```bash
# Check response times
boilerplate-cli health services

# Monitor system metrics
boilerplate-cli health check --json

# Test with different configurations
boilerplate-cli --config optimized-config.yaml health check
```

## Contributing

### Development Setup

```bash
# Clone repository
git clone <repository-url>
cd service-boilerplate/cli

# Install dependencies
go mod download

# Run tests
go test ./...

# Build and test
go build -o boilerplate-cli .
./boilerplate-cli --help
```

### Code Structure

Follow the established patterns:

```
cmd/
├── command_name.go     # Command implementation
└── command_name_test.go # Command tests

internal/
├── package_name/
│   ├── package_name.go     # Main implementation
│   ├── package_name_test.go # Unit tests
│   └── types.go           # Type definitions
```

### Adding New Commands

1. **Create command file** in `cmd/` directory
2. **Implement command structure** following existing patterns
3. **Add to root command** in `cmd/root.go`
4. **Add tests** for the new command
5. **Update documentation** in README and this guide

Example:

```go
// cmd/newcommand.go
func newNewCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "newcommand",
        Short: "Description of new command",
        Long:  `Detailed description of new command.`,
        RunE: func(cmd *cobra.Command, args []string) error {
            // Implementation
            return nil
        },
    }

    // Add flags
    cmd.Flags().StringP("flag", "f", "", "Description of flag")

    return cmd
}

// cmd/root.go
rootCmd.AddCommand(
    newNewCommand(),
)
```

### Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/client

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...
```

### Code Quality

- Follow Go naming conventions
- Add comprehensive error handling
- Include detailed logging
- Write unit tests for all functions
- Add documentation comments
- Use consistent formatting

### Pull Request Process

1. **Fork** the repository
2. **Create** a feature branch
3. **Implement** your changes
4. **Add tests** for new functionality
5. **Update documentation**
6. **Run tests** and ensure they pass
7. **Submit** pull request with detailed description

## Changelog

### Version 1.0.0 (Current)

#### Features Added
- ✅ **Core Infrastructure**: CLI framework with Cobra and Viper
- ✅ **Service Discovery**: Automatic service detection and health monitoring
- ✅ **API Client**: HTTP client with retry logic and error handling
- ✅ **Configuration Management**: Multi-source configuration support
- ✅ **User Operations**: Complete CRUD operations for user management
- ✅ **Data Operations**: Seed, export, validate, and cleanup functionality
- ✅ **Workflow Orchestration**: Predefined and custom workflow execution
- ✅ **Multi-Service Operations**: Coordinated operations across services
- ✅ **Advanced Monitoring**: Comprehensive health checks and metrics
- ✅ **Performance Monitoring**: Response time tracking and uptime analytics

#### Architecture Improvements
- ✅ **Modular Design**: Clean separation of concerns
- ✅ **Extensible Framework**: Plugin-ready architecture
- ✅ **Error Resilience**: Comprehensive error handling and recovery
- ✅ **Logging System**: Structured logging with multiple levels
- ✅ **Configuration System**: Environment-aware configuration management

#### Documentation
- ✅ **Comprehensive Guide**: Complete usage documentation
- ✅ **API Reference**: Detailed API documentation
- ✅ **Examples**: Practical usage examples
- ✅ **Troubleshooting**: Common issues and solutions

### Future Releases

#### Planned Features (v2.0.0)
- 🔄 **Plugin System**: Extensible command modules
- 📊 **Advanced Analytics**: Trend analysis and predictive monitoring
- 🌐 **REST API**: External system integration
- 🔗 **Distributed Operations**: Cross-cluster coordination
- 📱 **Web Dashboard**: Graphical monitoring interface

#### Enhancements (v1.1.0)
- 🔧 **Custom Workflow Editor**: GUI for workflow creation
- 📈 **Metrics Dashboard**: Real-time monitoring dashboard
- 🔒 **Authentication**: Secure API authentication
- 📋 **Batch Operations**: Bulk data operations
- 🎯 **Service Templates**: Automated service scaffolding

---

## Support

For support and questions:

- **Documentation**: This comprehensive guide
- **Issues**: GitHub issue tracker
- **Discussions**: GitHub discussions
- **Contributing**: See contributing guidelines above

## License

This project is licensed under the MIT License. See LICENSE file for details.

---

*This documentation is automatically generated and maintained. Last updated: $(date)*</content>
</xai:function_call