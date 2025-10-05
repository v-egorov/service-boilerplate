# Logging System Documentation

## Overview

The service boilerplate implements a comprehensive logging system designed for microservices architecture. It provides structured JSON logging with multiple output options, automatic log rotation, and seamless integration with Docker logging infrastructure.

## Tech Stack

### Core Libraries

- **[Logrus](https://github.com/sirupsen/logrus)**: Primary logging library providing structured logging with levels, fields, and JSON formatting
- **[Lumberjack](https://github.com/natefinch/lumberjack)**: Log rotation library for file-based logging with size limits and backup management
- **[Viper](https://github.com/spf13/viper)**: Configuration management for environment variables and config files

### Key Features

- **Structured JSON Logging**: Consistent log format across all services
- **Colorful Text Logging**: Enhanced visual formatting with color-coded HTTP methods and status codes
- **Multiple Output Modes**: File-only, stdout-only, or dual output
- **Automatic Rotation**: Size-based rotation with configurable limits
- **Docker Integration**: Compatible with Docker logging drivers
- **Service Context**: Automatic service identification in logs
- **Request Tracking**: Middleware for HTTP request/response logging

## Configuration

### Environment Variables

| Variable                        | Default (Prod) | Default (Dev) | Description                                                   |
| ------------------------------- | -------------- | ------------- | ------------------------------------------------------------- |
| `LOGGING_LEVEL`                 | `info`         | `debug`       | Log level: `debug`, `info`, `warn`, `error`, `fatal`, `panic` |
| `LOGGING_FORMAT`                | `json`         | `json`        | Log format: `json` (structured), `text` (colors enabled)      |
| `LOGGING_OUTPUT`                | `stdout`       | `file`        | Output destination: `stdout`, `file`                          |
| `LOGGING_DUAL_OUTPUT`           | `true`         | `true`        | Enable dual output (file + stdout) when `LOGGING_OUTPUT=file` |
| `LOGGING_STRIP_ANSI_FROM_FILES` | `true`         | `true`        | Strip ANSI codes from file logs (console colors preserved)    |

### Configuration Structure

```go
type LoggingConfig struct {
    Level              string `mapstructure:"level"`                // debug, info, warn, error, fatal, panic
    Format             string `mapstructure:"format"`               // json, text
    Output             string `mapstructure:"output"`               // stdout, file
    DualOutput         bool   `mapstructure:"dual_output"`          // Enable dual output mode
    StripANSIFromFiles bool   `mapstructure:"strip_ansi_from_files"` // Strip ANSI codes from file logs
}
```

## Colorful Formatting

### Visual Enhancements

The text logging format includes enhanced visual formatting to improve log readability:

- **HTTP Methods**: Background color-coded by method type

  - `GET` → Green background
  - `POST` → Blue background
  - `PUT` → Yellow background
  - `PATCH` → Magenta background
  - `DELETE` → Red background

- **HTTP Status Codes**: Background color-coded by response range

  - `2xx` → Green background (success)
  - `3xx` → Yellow background (redirect)
  - `4xx` → Red background (client error)
  - `5xx` → Red background (server error)

- **Service Names**: Unique background colors for easy identification

  - `api-gateway` → Magenta background
  - `user-service` → Cyan background
  - `auth-service` → Blue background

- **Log Levels**: Background colors for error visibility

  - `ERROR` → Red background
  - `WARN` → Yellow background
  - `INFO` → Cyan foreground

- **Error Text**: Red background highlighting for error keywords

### Example Colored Output

```
time="2025-09-27T18:27:50Z" level=warn msg="Request completed with client error" method=GET path=/api/v1/users service=api-gateway status=401
```

**Visual Result (in terminal with background colors):**

- `[36mINFO[0m` - Log level in cyan foreground
- `[36mduration_ms[0m` - Field names in cyan foreground
- `[42mGET[0m` - HTTP method with **green background**
- `[45mapi-gateway[0m` - Service name with **magenta background**
- `[41merror[0m` - Error text with **red background** (when applicable)
- `status=401` - Status code with red background (4xx error)

**Accessibility Note**: Background colors provide better contrast and visibility for users with color vision deficiencies.

## ANSI Code Handling

### Automatic File Cleaning

The logging system automatically strips ANSI escape codes from file outputs while preserving colors in console output:

- **Console Output**: Full ANSI colors maintained for development readability
- **File Output**: ANSI codes automatically removed for clean, parseable text logs
- **Dual Output**: Colors displayed on stdout, clean text written to files

### Configuration

```yaml
logging:
  strip_ansi_from_files: true # Default: true
```

### Benefits

- **Readable Logfiles**: No terminal escape sequences cluttering log files
- **Tool Compatibility**: Works seamlessly with log analysis and monitoring tools
- **Performance**: No overhead for console color display
- **Accessibility**: Clean text logs remain accessible for all users

### Implementation

File outputs are automatically wrapped with an ANSI stripping writer that removes escape sequences before writing to disk, ensuring logfiles contain only plain text content.

## Logging Modes

### 1. Standard Output (Default)

```bash
LOGGING_OUTPUT=stdout
LOGGING_DUAL_OUTPUT=false  # Not applicable
```

**Characteristics:**

- Logs written to stdout/stderr
- Captured by Docker logging driver
- No persistent file storage
- Suitable for containerized environments

**Example Output:**

```json
{
  "time": "2025-09-24T10:30:00Z",
  "level": "info",
  "msg": "Service started",
  "service": "api-gateway"
}
```

### 2. File Output

```bash
LOGGING_OUTPUT=file
LOGGING_DUAL_OUTPUT=false
```

**Characteristics:**

- Logs written to files only
- Automatic rotation with lumberjack
- Persistent storage across container restarts
- No console output

**File Location:** `/app/logs/{service-name}.log`

### 3. Dual Output (Recommended for Development)

```bash
LOGGING_OUTPUT=file
LOGGING_DUAL_OUTPUT=true
```

**Characteristics:**

- Logs written to both files AND stdout/stderr
- File rotation + Docker logging integration
- Best of both worlds for debugging
- Default for development environment

## File Structure

### Log Directories

```
docker/volumes/
├── api-gateway/
│   └── logs/
│       └── api-gateway.log
├── user-service/
│   └── logs/
│       └── user-service.log
└── auth-service/
    └── logs/
        └── auth-service.log
```

### Volume Configuration

Each service has dedicated log volumes defined in `docker-compose.yml`:

```yaml
volumes:
  api_gateway_logs:
    name: ${API_GATEWAY_LOGS_VOLUME}
    driver: local
    driver_opts:
      type: none
      o: bind
      device: ${PWD}/docker/volumes/api-gateway/logs
```

## Log Rotation

### Lumberjack Configuration

```go
lumberjack.Logger{
    Filename:   "/app/logs/service.log",  // Log file path
    MaxSize:    10,                       // Max size in MB before rotation
    MaxBackups: 3,                        // Number of backup files to keep
    MaxAge:     28,                       // Max age in days to keep files
    Compress:   true,                     // Compress rotated files
}
```

### Rotation Behavior

- **Size Limit**: 10MB per log file
- **Backup Files**: Up to 3 rotated files kept
- **Age Limit**: Files older than 28 days are deleted
- **Compression**: Rotated files are gzip compressed
- **Naming**: `service.log.1`, `service.log.2`, etc.

## Service Integration

### Logger Initialization

Each service initializes logging in `cmd/main.go`:

```go
logger := logging.NewLogger(logging.Config{
    Level:       cfg.Logging.Level,
    Format:      cfg.Logging.Format,
    Output:      cfg.Logging.Output,
    DualOutput:  cfg.Logging.DualOutput,
    ServiceName: cfg.App.Name,
})
```

### Request Logging Middleware

HTTP requests are automatically logged with structured data:

```go
serviceLogger := logging.NewServiceRequestLogger(logger.Logger, cfg.App.Name)
router.Use(serviceLogger.RequestResponseLogger())
```

**Request Log Fields:**

- `duration_ms`: Response time in milliseconds
- `ip`: Client IP address
- `method`: HTTP method
- `path`: Request path
- `request_id`: Unique request identifier
- `response_size`: Response body size in bytes
- `service`: Service name
- `status`: HTTP status code
- `timestamp`: ISO 8601 timestamp
- `user_agent`: Client user agent
- `user_id`: Authenticated user ID (if available)

## Logger Types and Usage

### Three-Tier Logging Architecture

The service boilerplate implements a comprehensive three-tier logging approach designed for different stakeholders and use cases:

#### 1. Application Logger (`h.logger`)

**Purpose**: General application debugging and monitoring

- **Target Audience**: Developers and DevOps teams
- **Content**: Business logic events, error conditions, request correlation, debugging information
- **Retention**: Short-term (development and troubleshooting)
- **Example Usage**:

  ```go
  h.logger.WithFields(logrus.Fields{
      "user_id": user.ID,
      "request_id": requestID,
  }).Info("User created successfully")
  ```

#### 2. Standard Logger (`h.standardLogger`)

**Purpose**: Structured business operation logging with consistent schema

- **Target Audience**: Business analysts, automated monitoring systems, metrics collection
- **Content**: User actions, operation types, success/failure status, standardized business events
- **Retention**: Medium-term (analytics and metrics)
- **Example Usage**:

  ```go
  h.standardLogger.UserOperation(requestID, user.ID.String(), "create", true, nil)
  ```

#### 3. Audit Logger (`h.auditLogger`)

**Purpose**: Security and compliance auditing for sensitive operations

- **Target Audience**: Security teams, auditors, compliance officers
- **Content**: Security events, authentication attempts, user creation/modification, sensitive operations with full trace correlation
- **Retention**: Long-term (compliance and legal requirements)
- **Example Usage**:

  ```go
  h.auditLogger.LogUserCreation(requestID, user.ID.String(), ipAddress, userAgent, traceID, spanID, true, "")
  ```

### Logger Usage Guidelines

| Logger             | When to Use                                                  | Log Level               | Trace Correlation         | Compliance |
| ------------------ | ------------------------------------------------------------ | ----------------------- | ------------------------- | ---------- |
| `h.logger`         | Application debugging, error handling, business logic events | `debug/info/warn/error` | Partial (request_id)      | No         |
| `h.standardLogger` | Business operations, user actions, metrics collection        | `info`                  | Via request_id            | No         |
| `h.auditLogger`    | Security events, authentication, compliance actions          | `warn/error`            | Full (trace_id + span_id) | Yes        |

### Best Practices

- **Use `h.logger`** for debugging application flow and technical errors
- **Use `h.standardLogger`** for business metrics and user activity tracking
- **Use `h.auditLogger`** for security events and compliance logging
- **Always include trace information** in audit logs for full observability
- **Use appropriate log levels** based on the logger type and audience

### Integration with Observability Stack

All three logger types integrate with the centralized logging infrastructure:

- **Application logs** appear in Grafana dashboards for monitoring
- **Standard logs** feed into business analytics and alerting systems
- **Audit logs** are correlated with distributed traces in Jaeger for security investigations

## Audit Method Creation

### Overview

When implementing new entity types or security events in your services, you need to add corresponding audit logging methods to `common/logging/audit.go`. This ensures consistent audit logging across all services with proper trace correlation and structured data.

### When to Add Audit Methods

Add new audit methods when:
- Creating new entity types (users, orders, products, etc.)
- Implementing security-sensitive operations
- Adding authentication or authorization features
- Creating custom business operations that require audit trails

### Method Naming Convention

Follow this pattern for audit method names:
```
Log<EntityType><Operation>
```

**Examples:**
- `LogUserCreation` - for user account creation
- `LogOrderUpdate` - for order modifications
- `LogProductDeletion` - for product removal
- `LogPermissionGrant` - for permission assignments

### Method Signature

All audit methods follow this standard signature:

```go
func (al *AuditLogger) Log<EntityType><Operation>(
    actorUserID, requestID, targetID, ipAddress, userAgent, traceID, spanID string,
    success bool, errorMsg string
)
```

**Parameters:**
- `actorUserID`: ID of the authenticated user performing the action (populates UserID field)
- `requestID`: Unique request identifier for correlation
- `targetID`: ID of the target entity/user being operated on (populates EntityID field)
- `ipAddress`: Client IP address
- `userAgent`: Client user agent string
- `traceID`: Distributed tracing trace ID
- `spanID`: Distributed tracing span ID
- `success`: Boolean indicating operation success
- `errorMsg`: Error message if operation failed (empty string if successful)

### AuditEvent Field Usage

The AuditEvent struct contains both UserID and EntityID fields with distinct purposes:

- **UserID**: Identifies "who initiated the operation" (the authenticated user performing the action)
- **EntityID**: Identifies "the object upon which operation is performed" (the target entity)

**Usage Patterns:**
- Authentication operations: Set UserID to the authenticating user, EntityID empty
- Entity operations: Set EntityID to the target entity, UserID to the actor (when available)
- User management: Set EntityID to the target user, UserID to the admin/system performing the action

### Implementation Pattern

Use existing methods as templates. Here's the standard implementation pattern:

```go
// Log<EntityType><Operation> logs <entity type> <operation> events
func (al *AuditLogger) Log<EntityType><Operation>(actorUserID, requestID, targetID, ipAddress, userAgent, traceID, spanID string, success bool, errorMsg string) {
    event := AuditEvent{
        Timestamp: time.Now().UTC(),
        EventType: "<entity_type>_<operation>",  // e.g., "user_creation", "order_update"
        Service:   al.serviceName,
        UserID:    actorUserID,  // Who performed the action
        EntityID:  targetID,     // Object upon which operation is performed
        RequestID: requestID,
        IPAddress: ipAddress,
        UserAgent: userAgent,
        Resource:  "<entity_type>",  // e.g., "user", "order", "product"
        Action:    "<operation>",     // e.g., "create", "update", "delete"
        TraceID:   traceID,
        SpanID:    spanID,
    }

    if success {
        event.Result = "success"
    } else {
        event.Result = "failure"
        event.Error = errorMsg
    }

    al.logEvent(event)
}
```

**Note:** For operations where the authenticated actor is not available (public endpoints, system operations), pass an empty string for actorUserID.

### Generic Methods Available

For common CRUD operations, you can use the existing generic methods:

- `LogEntityCreation()` - Generic entity creation
- `LogEntityUpdate()` - Generic entity update
- `LogEntityDeletion()` - Generic entity deletion

However, prefer service-specific methods for better audit trail clarity and filtering.

### Adding New Audit Methods

1. **Open `common/logging/audit.go`**
2. **Add your method following the pattern above**
3. **Update service handlers to call the new method**
4. **Test the audit logging integration**

**Example: Adding Product Audit Methods**

```go
// LogProductCreation logs product creation events
func (al *AuditLogger) LogProductCreation(actorUserID, requestID, productID, ipAddress, userAgent, traceID, spanID string, success bool, errorMsg string) {
    event := AuditEvent{
        Timestamp: time.Now().UTC(),
        EventType: "product_creation",
        Service:   al.serviceName,
        UserID:    actorUserID,  // Who created the product
        EntityID:  productID,    // The product being created
        RequestID: requestID,
        IPAddress: ipAddress,
        UserAgent: userAgent,
        Resource:  "product",
        Action:    "create",
        TraceID:   traceID,
        SpanID:    spanID,
    }

    if success {
        event.Result = "success"
    } else {
        event.Result = "failure"
        event.Error = errorMsg
    }

    al.logEvent(event)
}

// LogProductUpdate logs product update events
func (al *AuditLogger) LogProductUpdate(actorUserID, requestID, productID, ipAddress, userAgent, traceID, spanID string, success bool, errorMsg string) {
    event := AuditEvent{
        Timestamp: time.Now().UTC(),
        EventType: "product_update",
        Service:   al.serviceName,
        UserID:    actorUserID,  // Who updated the product
        EntityID:  productID,    // The product being updated
        RequestID: requestID,
        IPAddress: ipAddress,
        UserAgent: userAgent,
        Resource:  "product",
        Action:    "update",
        TraceID:   traceID,
        SpanID:    spanID,
    }

    if success {
        event.Result = "success"
    } else {
        event.Result = "failure"
        event.Error = errorMsg
    }

    al.logEvent(event)
}
```

### Usage in Service Handlers

After adding the audit method, use it in your service handlers. Always pass the authenticated user ID when available:

```go
func (h *ProductHandler) CreateProduct(c *gin.Context) {
    // Extract trace information
    traceID := tracing.GetTraceID(c.Request.Context())
    spanID := tracing.GetSpanID(c.Request.Context())

    // Get authenticated user ID (if available from middleware/context)
    // For now, using empty string until proper auth middleware is implemented
    actorUserID := getAuthenticatedUserID(c) // Returns "" if not authenticated

    // Business logic...
    product, err := h.service.CreateProduct(c.Request.Context(), req)

    if err != nil {
        h.auditLogger.LogProductCreation(
            actorUserID,                    // Who attempted to create
            c.GetHeader("X-Request-ID"),    // Request correlation
            "",                             // No product ID for failures
            c.ClientIP(),
            c.GetHeader("User-Agent"),
            traceID,
            spanID,
            false,                          // success = false
            err.Error(),
        )
        // Handle error...
    }

    // Success case
    h.auditLogger.LogProductCreation(
        actorUserID,                        // Who created the product
        c.GetHeader("X-Request-ID"),        // Request correlation
        product.ID.String(),                // The created product ID
        c.ClientIP(),
        c.GetHeader("User-Agent"),
        traceID,
        spanID,
        true,                               // success = true
        "",                                 // no error
    )

    // Return response...
}
```

**Authentication Context:** The `actorUserID` parameter should be extracted from the request context or JWT token. For public endpoints or when authentication is not required, pass an empty string.

### Best Practices

- **Consistent Naming**: Use clear, descriptive entity and operation names
- **Trace Correlation**: Always include traceID and spanID for full observability
- **Error Details**: Provide meaningful error messages for failed operations
- **Resource Clarity**: Use specific resource names rather than generic "entity"
- **Event Type Uniqueness**: Ensure event_type values are unique across services
- **Documentation**: Add comments explaining what each audit method logs

### Testing Audit Methods

Test your audit methods by:
1. Making requests that trigger the audit logging
2. Checking logs in Grafana/Loki for the expected audit events
3. Verifying trace correlation in Jaeger
4. Ensuring all required fields are present in the audit logs

## Usage Examples

### Development Setup

```bash
# Structured JSON logging is enabled by default in development
# These are the default settings in docker-compose.override.yml:
LOGGING_LEVEL=debug          # Enhanced debugging
LOGGING_FORMAT=json          # JSON format for structured logging
LOGGING_OUTPUT=file          # File output for persistence
LOGGING_DUAL_OUTPUT=true     # Both file and stdout for Docker logs

# View logs:
make logs                    # See all services (JSON format)
docker logs <service-name>   # Individual service logs

# For colorful text logging in development:
LOGGING_FORMAT=text          # Switch to text format with colors
```

### Production Setup

```bash
# File-only logging for production
LOGGING_LEVEL=info
LOGGING_OUTPUT=file
LOGGING_DUAL_OUTPUT=false
```

### Docker Compose Override

```yaml
services:
  api-gateway:
    environment:
      - LOGGING_LEVEL=debug
      - LOGGING_OUTPUT=file
      - LOGGING_DUAL_OUTPUT=true
```

## Debug Printing and Output

### Overview

Debug printing is essential for development and troubleshooting. The logging system provides structured debug logging that integrates seamlessly with your existing logging infrastructure. Always use the structured logger instead of `fmt.Printf()` or `logrus.Printf()` for debug output.

### Recommended Approaches

#### 1. Structured Debug Logging (Primary Method)

**Repository Layer Example:**

```go
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
    r.logger.WithField("user_id", id).Debug("Starting user lookup by ID")

    query := `SELECT id, email, password_hash, first_name, last_name, created_at, updated_at FROM user_service.users WHERE id = $1`

    user := &models.User{}
    err := database.TraceDBQuery(ctx, "user_service.users", query, func(ctx context.Context) error {
        return r.db.QueryRow(ctx, query, id).Scan(
            &user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)
    })

    if err != nil {
        r.logger.WithFields(logrus.Fields{
            "user_id": id,
            "error_type": fmt.Sprintf("%T", err),
            "query": query,
        }).Debug("User lookup failed with database error")
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    r.logger.WithFields(logrus.Fields{
        "user_id": user.ID,
        "email": user.Email,
        "operation_duration": "calculated_if_needed",
    }).Debug("User lookup completed successfully")

    return user, nil
}
```

**Handler Layer Example:**

```go
func (h *UserHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")
    requestID := c.GetHeader("X-Request-ID")

    h.logger.WithFields(logrus.Fields{
        "user_id": userID,
        "request_id": requestID,
        "method": c.Request.Method,
        "path": c.Request.URL.Path,
    }).Debug("Processing get user request")

    // Business logic...

    h.logger.WithFields(logrus.Fields{
        "user_id": userID,
        "request_id": requestID,
        "status_code": 200,
        "response_size": "calculated_size",
    }).Debug("Get user request completed")
}
```

#### 2. Service Logger Debug Output

**Using Service-Specific Loggers:**

```go
func (h *UserHandler) CreateUser(c *gin.Context) {
    // Use the service logger for structured debug output
    h.standardLogger.Debug("User creation initiated",
        "ip", c.ClientIP(),
        "user_agent", c.GetHeader("User-Agent"))

    // Business logic...

    h.auditLogger.LogAuthAttempt(requestID, ipAddress, userAgent, "user_creation", true, "success")
}
```

#### 3. Temporary Debug Prints (Development Only)

**For quick debugging during development (remove before commit):**

```go
// Temporary debug - REMOVE BEFORE COMMIT
fmt.Printf("DEBUG: Processing user ID %s with email %s\n", userID, email)

// Or use structured temporary debug
logrus.WithFields(logrus.Fields{
    "user_id": userID,
    "email": email,
    "step": "validation",
}).Warn("TEMP DEBUG: User validation step")  // Use Warn to ensure visibility
```

### Debug Configuration

#### Environment Setup

**Enable Debug Logging:**

```bash
# Set log level to debug
LOGGING_LEVEL=debug

# For development with colors
LOGGING_FORMAT=text
LOGGING_OUTPUT=file
LOGGING_DUAL_OUTPUT=true
```

**Docker Compose Override:**

```yaml
services:
  user-service:
    environment:
      - LOGGING_LEVEL=debug
      - LOGGING_FORMAT=text # For colored output
```

#### Configuration Validation

**Test Debug Output:**

```bash
# Start services
make dev

# Make a request to trigger debug logs
curl http://localhost:8080/api/v1/users

# View debug logs
docker logs service-boilerplate-user-service | grep '"level":"debug"'
```

### Best Practices

#### ✅ Do's

- **Use Debug Level**: Always use `logger.Debug()` not `logger.Info()`
- **Include Context**: Add relevant fields (user_id, request_id, operation, etc.)
- **Structured Fields**: Use `WithField()` or `WithFields()` for structured data
- **Request Correlation**: Include `request_id` for tracing requests across services
- **Performance Context**: Log operation duration, query parameters, result counts

#### ❌ Don'ts

- **Don't Use Printf**: Avoid `fmt.Printf()`, `logrus.Printf()` in production code
- **Don't Log Secrets**: Never log passwords, tokens, or sensitive data
- **Don't Over-Log**: Debug logs should be meaningful, not verbose spam
- **Don't Commit Temp Debug**: Remove temporary debug prints before committing
- **Don't Use Wrong Level**: Debug logs disappear in production - use appropriate levels

### Debug Log Examples

#### Database Operation Debug

```json
{
  "time": "2025-09-27T10:30:00Z",
  "level": "debug",
  "msg": "Starting user lookup by ID",
  "service": "user-service",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "request_id": "abc-123-def"
}
```

#### Business Logic Debug

```json
{
  "time": "2025-09-27T10:30:00Z",
  "level": "debug",
  "msg": "User validation completed",
  "service": "user-service",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "validation_result": "passed",
  "checks_performed": ["email_format", "password_strength", "duplicate_check"]
}
```

#### Error Context Debug

```json
{
  "time": "2025-09-27T10:30:00Z",
  "level": "debug",
  "msg": "Database query failed",
  "service": "user-service",
  "error_type": "*pgconn.PgError",
  "query": "SELECT * FROM users WHERE id = $1",
  "parameters": ["123e4567-e89b-12d3-a456-426614174000"],
  "error_code": "23505"
}
```

### Performance Considerations

- **Debug Level Filtering**: Debug logs are filtered out in production (no performance impact)
- **Structured Overhead**: Minimal overhead for field addition vs unstructured strings
- **Memory Usage**: Debug logs consume memory buffers even when filtered
- **I/O Impact**: Dual output writes debug logs to both files and stdout

### Integration with Tracing

**Debug logs complement tracing:**

```go
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
    r.logger.WithField("user_id", id).Debug("Starting user deletion")

    query := `DELETE FROM user_service.users WHERE id = $1`

    var result pgconn.CommandTag
    err := database.TraceDBDelete(ctx, "user_service.users", query, func(ctx context.Context) error {
        var execErr error
        result, execErr = r.db.Exec(ctx, query, id)
        return execErr
    })

    if err != nil {
        r.logger.WithError(err).WithField("user_id", id).Debug("User deletion failed")
        return err
    }

    r.logger.WithFields(logrus.Fields{
        "user_id": id,
        "rows_affected": result.RowsAffected(),
    }).Debug("User deletion completed")

    return nil
}
```

### Troubleshooting Debug Issues

#### Debug Logs Not Appearing

**Check Configuration:**

```bash
# Verify log level
docker exec service-boilerplate-user-service env | grep LOGGING_LEVEL

# Check if service restarted after config change
docker-compose restart user-service
```

**Verify Code:**

```go
// Ensure you're using the correct logger instance
r.logger.Debug("message")  // ✅ Correct
logrus.Debug("message")    // ❌ Wrong - bypasses service context
```

#### Debug Logs Too Verbose

**Filter Debug Output:**

```bash
# View only debug logs
docker logs service-boilerplate-user-service 2>&1 | jq 'select(.level == "debug")'

# Filter by specific field
docker logs service-boilerplate-user-service 2>&1 | jq 'select(.user_id == "123")'
```

## Log Analysis

### Viewing Logs

**Docker Logs:**

```bash
docker logs service-boilerplate-api-gateway
```

**File Logs:**

```bash
tail -f docker/volumes/api-gateway/logs/api-gateway.log
```

**Filtered Logs:**

```bash
# Find errors in last hour
grep '"level":"error"' docker/volumes/*/logs/*.log | jq '.msg'

# Find slow requests (>1000ms)
grep '"duration_ms":[0-9]\{4,\}' docker/volumes/*/logs/*.log
```

### Log Structure

```json
{
  "time": "2025-09-24T10:30:00Z",
  "level": "info",
  "msg": "Request completed successfully",
  "duration_ms": 150,
  "ip": "192.168.1.100",
  "method": "GET",
  "path": "/api/v1/users",
  "request_id": "abc-123-def",
  "response_size": 2048,
  "service": "user-service",
  "status": 200,
  "timestamp": "2025-09-24T10:30:00Z",
  "user_agent": "Mozilla/5.0...",
  "user_id": "user-456"
}
```

## Troubleshooting

### Common Issues

#### Logs Not Appearing in Files

**Symptoms:** Log files are empty or missing
**Solutions:**

1. Check volume mounts: `docker inspect <container>` → verify `/app/logs` mount
2. Verify permissions: Container should have write access to log directory
3. Check service configuration: Ensure `LOGGING_OUTPUT=file`

#### Logs Not Appearing in Docker

**Symptoms:** `docker logs` shows no application logs
**Solutions:**

1. Verify dual output is enabled: `LOGGING_DUAL_OUTPUT=true`
2. Check if service is using file-only mode
3. Ensure stdout/stderr are not redirected

#### Log Rotation Not Working

**Symptoms:** Single large log file, no rotation
**Solutions:**

1. Check lumberjack configuration in code
2. Verify file permissions for rotation
3. Ensure sufficient disk space

#### Permission Denied Errors

**Symptoms:** Logger initialization fails
**Solutions:**

1. Create log directories: `make create-volumes-dirs`
2. Check volume ownership: `ls -la docker/volumes/*/logs/`
3. Verify Docker user permissions

### Debug Commands

```bash
# Check volume mounts
docker inspect service-boilerplate-api-gateway | jq '.[0].Mounts'

# View real-time logs
docker logs -f service-boilerplate-api-gateway

# Check log file permissions
ls -la docker/volumes/*/logs/

# Test logging configuration
curl http://localhost:8080/health
tail docker/volumes/api-gateway/logs/api-gateway.log
```

## Performance Considerations

### Dual Output Impact

- **Minimal Overhead**: `io.MultiWriter` adds negligible performance cost
- **Disk I/O**: Dual output doubles write operations
- **Memory**: Lumberjack buffers writes for efficiency

### Recommendations

- **Development**: Use dual output for debugging
- **Production**: Use file-only output to reduce I/O overhead
- **High Traffic**: Monitor disk I/O and adjust rotation settings

## Service Creation Integration

When creating new services with `create-service.sh`, logging is automatically configured:

- ✅ Environment variables added to `docker-compose.yml` and `docker-compose.override.yml`
- ✅ Log volume mounts configured
- ✅ Volume directories created
- ✅ Dual output enabled by default in development

## Future Enhancements

### Planned Features

- **Centralized Logging**: Integration with ELK stack or similar
- **Log Shipping**: Automatic log forwarding to external systems
- **Metrics Integration**: Log-based metrics collection
- **Structured Fields**: Additional context fields for better observability

### Configuration Extensions

- **Custom Formatters**: Support for additional log formats
- **External Writers**: Integration with cloud logging services
- **Conditional Logging**: Environment-based log filtering

---

## Quick Reference

### Environment Variables Summary

```bash
# Core settings
LOGGING_LEVEL=info
LOGGING_FORMAT=json
LOGGING_OUTPUT=file
LOGGING_DUAL_OUTPUT=true
LOGGING_STRIP_ANSI_FROM_FILES=true

# File locations
# Logs: docker/volumes/{service}/logs/{service}.log
# Rotation: 10MB, 3 backups, 28 days, compressed
```

### Key Commands

```bash
# View logs
docker logs service-boilerplate-api-gateway
tail -f docker/volumes/api-gateway/logs/api-gateway.log

# Clean logs
make clean-logs

# Create volumes
make create-volumes-dirs
```

This logging system provides robust, scalable logging infrastructure suitable for production microservices deployments.

## Known Issues

- In prod environment log volumes might not created (not present in docker-compose.yml) - need to check
- Filename for logs in prod environment not properly set by service - for all services service-boilerplate.log filename is used
- Based on prev item - there is a possibility that other env vars not created / not propagated properly to running containers
- NewServiceRequestLogger referenced above probably not used anymore - need to examine
