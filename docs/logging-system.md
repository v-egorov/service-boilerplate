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

| Variable                          | Default (Prod) | Default (Dev) | Description                                                   |
| --------------------------------- | -------------- | ------------- | ------------------------------------------------------------- |
| `LOGGING_LEVEL`                   | `info`         | `debug`       | Log level: `debug`, `info`, `warn`, `error`, `fatal`, `panic` |
| `LOGGING_FORMAT`                  | `json`         | `json`        | Log format: `json` (structured), `text` (colors enabled)     |
| `LOGGING_OUTPUT`                  | `stdout`       | `file`        | Output destination: `stdout`, `file`                          |
| `LOGGING_DUAL_OUTPUT`             | `true`         | `true`        | Enable dual output (file + stdout) when `LOGGING_OUTPUT=file` |
| `LOGGING_STRIP_ANSI_FROM_FILES`  | `true`         | `true`        | Strip ANSI codes from file logs (console colors preserved)   |

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
  strip_ansi_from_files: true  # Default: true
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

