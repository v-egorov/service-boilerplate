# Air Configuration Guide

## Overview

Air is configured using TOML files (`.air.toml`) located in each service directory. This document explains all configuration options and their current settings.

## Configuration Files

- `api-gateway/.air.toml` - API Gateway hot reload configuration
- `services/user-service/.air.toml` - User Service hot reload configuration

## Configuration Structure

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/api-gateway"
  cmd = "go build -o ./tmp/api-gateway ./cmd"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "docker"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

## Build Configuration

### Core Build Settings

| Setting | Current Value | Description |
|---------|---------------|-------------|
| `bin` | `./tmp/api-gateway` | Output binary path |
| `cmd` | `go build -o ./tmp/api-gateway ./cmd` | Build command |
| `delay` | `1000` | Delay before rebuild (milliseconds) |
| `kill_delay` | `0s` | Delay before killing old process |

### File Monitoring

| Setting | Current Value | Description |
|---------|---------------|-------------|
| `include_ext` | `["go", "tpl", "tmpl", "html"]` | File extensions to watch |
| `exclude_dir` | `["assets", "tmp", "vendor", "testdata", "docker"]` | Directories to ignore |
| `exclude_regex` | `["_test.go"]` | File patterns to exclude |
| `exclude_unchanged` | `false` | Exclude unchanged files |
| `follow_symlink` | `false` | Follow symbolic links |

### Advanced Build Options

| Setting | Current Value | Description |
|---------|---------------|-------------|
| `args_bin` | `[]` | Arguments passed to binary |
| `full_bin` | `""` | Full path to binary (overrides `bin`) |
| `poll` | `false` | Use polling instead of events |
| `poll_interval` | `0` | Polling interval (when polling enabled) |
| `rerun` | `false` | Rerun binary instead of replacing |
| `rerun_delay` | `500` | Delay before rerun (milliseconds) |
| `send_interrupt` | `false` | Send interrupt signal before kill |
| `stop_on_root` | `false` | Stop on root directory changes |

## Logging Configuration

### Build Logs

| Setting | Current Value | Description |
|---------|---------------|-------------|
| `log` | `build-errors.log` | Build error log file |

### Console Output

| Setting | Current Value | Description |
|---------|---------------|-------------|
| `main_only` | `false` | Show only main process output |
| `time` | `false` | Show timestamps in logs |

## Color Configuration

Air uses colors to distinguish different types of output:

| Setting | Current Value | Description |
|---------|---------------|-------------|
| `app` | `""` | Application output color |
| `build` | `"yellow"` | Build process color |
| `main` | `"magenta"` | Main process color |
| `runner` | `"green"` | Runner color |
| `watcher` | `"cyan"` | File watcher color |

## Screen Configuration

| Setting | Current Value | Description |
|---------|---------------|-------------|
| `clear_on_rebuild` | `false` | Clear screen on rebuild |
| `keep_scroll` | `true` | Keep scroll position |

## Miscellaneous

| Setting | Current Value | Description |
|---------|---------------|-------------|
| `clean_on_exit` | `false` | Clean temporary files on exit |

## Directory Structure

```
service-root/
├── .air.toml          # Air configuration
├── tmp/              # Build artifacts (auto-created)
├── cmd/              # Main application entry point
├── internal/         # Internal packages
└── build-errors.log  # Build error logs
```

## Customization Examples

### Adding New File Types to Watch

```toml
include_ext = ["go", "tpl", "tmpl", "html", "yaml", "yml"]
```

### Excluding Additional Directories

```toml
exclude_dir = ["assets", "tmp", "vendor", "testdata", "docker", "node_modules"]
```

### Changing Build Delay

```toml
delay = 500  # Faster rebuilds (500ms)
```

### Enabling Polling Mode

```toml
poll = true
poll_interval = 1000  # Check every second
```

### Custom Build Command

```toml
cmd = "go build -tags=dev -o ./tmp/api-gateway ./cmd"
```

## Environment-Specific Configuration

You can use environment variables in your `.air.toml`:

```toml
[build]
  cmd = "go build -o ./tmp/$SERVICE_NAME ./cmd"
```

## Best Practices

1. **Keep build delay reasonable** (500-2000ms) to avoid excessive CPU usage
2. **Exclude unnecessary files** to improve performance
3. **Use descriptive binary names** for easier debugging
4. **Configure colors** for better log readability
5. **Monitor build logs** for compilation errors

## Troubleshooting Configuration

- Check `build-errors.log` for compilation issues
- Verify file paths in `include_ext` and `exclude_dir`
- Ensure `cmd` can execute from the service root directory
- Test configuration with `air -c .air.toml` locally

## Docker Configuration

### Docker Compose Override File (`docker-compose.override.yml`)

#### Purpose and Role

The `docker-compose.override.yml` file is a **development-specific configuration file** that extends and modifies the base `docker-compose.yml` configuration. It follows Docker Compose's [override mechanism](https://docs.docker.com/compose/extends/#multiple-compose-files) to provide development-friendly settings without modifying the production configuration.

#### Key Characteristics

- **Automatic Loading**: Docker Compose automatically merges `docker-compose.override.yml` with `docker-compose.yml` when running `docker-compose up`
- **Development Focused**: Contains settings optimized for local development (hot reload, debugging, volume mounts)
- **Non-Intrusive**: Changes in override file don't affect production deployments
- **Override Priority**: Settings in override file take precedence over base configuration

#### Usage Workflow

##### 1. Development Environment Setup
```bash
# Start development environment (automatically uses override)
make dev
# or
docker-compose up

# Both commands automatically merge:
# docker-compose.yml + docker-compose.override.yml
```

##### 2. Production Deployment
```bash
# Production uses only base configuration
docker-compose -f docker-compose.yml up

# Or with explicit file specification
make up
```

##### 3. Override-Specific Commands
```bash
# View merged configuration
docker-compose config

# View only override configuration
docker-compose -f docker-compose.override.yml config

# Validate override file
docker-compose -f docker-compose.override.yml config --quiet
```

#### Configuration Differences

| Aspect | `docker-compose.yml` | `docker-compose.override.yml` |
|--------|---------------------|------------------------------|
| **Purpose** | Production-ready base config | Development enhancements |
| **Dockerfiles** | `Dockerfile` (optimized) | `Dockerfile.dev` (with Air) |
| **Environment** | `APP_ENV=production` | `APP_ENV=development` |
| **Volumes** | Minimal (data only) | Full source code mounts |
| **Ports** | Configurable via env | Development ports |
| **Dependencies** | Health checks | Development tools |
| **Logging** | JSON format | Text format with debug |

#### Override File Contents

##### Development-Specific Settings
```yaml
# Development Dockerfile with Air hot reload
build:
  dockerfile: api-gateway/Dockerfile.dev

# Development environment
environment:
  - APP_ENV=development
  - LOGGING_LEVEL=debug

# Source code volume mounts for hot reload
volumes:
  - ../api-gateway:/app/api-gateway
  - ${API_GATEWAY_TMP_VOLUME}:/app/api-gateway/tmp

# Development ports
ports:
  - "${API_GATEWAY_PORT}:${API_GATEWAY_PORT}"

# Network aliases for service discovery
networks:
  service-network:
    aliases:
      - ${API_GATEWAY_NAME}
      - gateway
      - api
```

##### Workflow Integration

1. **Local Development**:
   ```bash
   # Override enables hot reload and debugging
   make dev
   # Files changed → Air rebuilds → No restart needed
   ```

2. **Testing Changes**:
   ```bash
   # Override mounts source code
   vim api-gateway/internal/handlers/gateway.go
   # Changes immediately available in container
   ```

3. **Debugging**:
   ```bash
   # Override provides debug environment
   docker-compose exec api-gateway sh
   # Full debugging tools available
   ```

#### Best Practices

##### File Organization
```
docker/
├── docker-compose.yml          # Base production config
├── docker-compose.override.yml # Development overrides
└── docker-compose.test.yml     # Testing overrides (optional)
```

##### Environment Separation
```bash
# Development (with override)
docker-compose up

# Production (without override)
docker-compose -f docker-compose.yml up

# Testing (with test override)
docker-compose -f docker-compose.yml -f docker-compose.test.yml up
```

##### Version Control
- ✅ Commit `docker-compose.override.yml` (contains development config)
- ✅ Include in `.gitignore` if it contains secrets
- ❌ Never commit sensitive data in override files

#### Troubleshooting Override Issues

##### Common Problems
```bash
# Check merged configuration
docker-compose config

# Validate override file syntax
docker-compose -f docker-compose.override.yml config

# Check if override is being loaded
docker-compose ps  # Should show development containers
```

##### Override Not Loading
- Ensure file is named exactly `docker-compose.override.yml`
- Check file is in same directory as `docker-compose.yml`
- Verify YAML syntax is valid
- Use `docker-compose config` to see merged result

##### Environment Variable Issues
```bash
# Check variable values
docker-compose exec api-gateway env | grep API_GATEWAY

# Override .env file location
docker-compose --env-file .env.dev up
```

#### Advanced Usage

##### Multiple Override Files
```bash
# Development with custom settings
docker-compose -f docker-compose.yml -f docker-compose.override.yml -f docker-compose.custom.yml up

# CI/CD pipeline
docker-compose -f docker-compose.yml -f docker-compose.ci.yml up
```

##### Conditional Overrides
```yaml
# Override only in development
services:
  api-gateway:
    # Only applies when override is loaded
    environment:
      - DEBUG=true
      - HOT_RELOAD=true
```

### Environment Variables
Docker configuration is fully controlled via `.env` file:

| Variable | Default Value | Description |
|----------|---------------|-------------|
| `API_GATEWAY_NAME` | `service-boilerplate-api-gateway` | API Gateway container name |
| `USER_SERVICE_NAME` | `service-boilerplate-user-service` | User Service container name |
| `POSTGRES_NAME` | `service-boilerplate-postgres` | PostgreSQL container name |
| `SERVICE_NETWORK_NAME` | `service-boilerplate-network` | Docker network name |
| `API_GATEWAY_PORT` | `8080` | API Gateway external port |
| `USER_SERVICE_PORT` | `8081` | User Service external port |
| `DATABASE_PORT` | `5432` | PostgreSQL external port |
| `API_GATEWAY_TMP_VOLUME` | `service-boilerplate-api-gateway-tmp` | API Gateway temp volume |
| `USER_SERVICE_TMP_VOLUME` | `service-boilerplate-user-service-tmp` | User Service temp volume |
| `POSTGRES_VOLUME` | `service-boilerplate-postgres-data` | PostgreSQL data volume |
| `DATABASE_NAME` | `service_db` | PostgreSQL database name |
| `DATABASE_USER` | `postgres` | PostgreSQL username |
| `DATABASE_PASSWORD` | `postgres` | PostgreSQL password |
| `MIGRATION_CONTAINER_NAME` | `service-boilerplate-migration` | Migration container name |
| `MIGRATION_TMP_VOLUME` | `service-boilerplate-migration-tmp` | Migration temp volume |
| `MIGRATION_IMAGE` | `migrate/migrate:latest` | Migration tool image |
| `SERVICE_NAME` | `user-service` | Current service for migrations |

### Network Configuration
Services use named networks with multiple aliases for flexible connectivity:

```yaml
# docker/docker-compose.yml
networks:
  service-network:
    name: ${SERVICE_NETWORK_NAME}
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

### Service Aliases
Each service has multiple DNS names for inter-service communication:

```yaml
# API Gateway
networks:
  service-network:
    aliases:
      - ${API_GATEWAY_NAME}
      - gateway
      - api

# User Service
networks:
  service-network:
    aliases:
      - ${USER_SERVICE_NAME}
      - users
      - user-svc

# PostgreSQL
networks:
  service-network:
    aliases:
      - ${POSTGRES_NAME}
      - db
      - database
```

### Migration Service Configuration
Database migrations are handled by a dedicated Docker container:

```yaml
# Migration Service
migration:
  image: ${MIGRATION_IMAGE}
  container_name: ${MIGRATION_CONTAINER_NAME}
  profiles: ["migration"]
  volumes:
    - ./services/${SERVICE_NAME}/migrations:/migrations:ro
    - ${MIGRATION_TMP_VOLUME}:/tmp/migrations
  environment:
    - DATABASE_URL=postgres://${DATABASE_USER}:${DATABASE_PASSWORD}@${POSTGRES_NAME}:${DATABASE_PORT}/${DATABASE_NAME}?sslmode=${DATABASE_SSL_MODE}
  networks:
    - service-network
  command: ["version"]
  depends_on:
    - ${POSTGRES_NAME}
```

#### Migration Profiles
The migration service uses Docker Compose profiles for optional execution:

```bash
# Run with migration service
docker-compose --profile migration up

# Run migrations only
docker-compose --profile migration run --rm migration up

# Normal development (without migration service)
docker-compose up
```

## Related Files

- `api-gateway/.air.toml` - API Gateway configuration
- `services/user-service/.air.toml` - User Service configuration
- `Makefile` - Development commands
- `docker/docker-compose.override.yml` - Docker development setup
- `.env` - Environment variables and naming configuration</content>
</xai:function_call">### 📋 **Configuration Documentation Created**

I've created comprehensive configuration documentation covering:

#### **📄 Configuration Guide Features**
- **Complete Settings Reference** - All `.air.toml` options explained
- **Current Configuration** - Exact values from your project
- **File Monitoring** - Include/exclude patterns and extensions
- **Build Options** - Command, delay, and rebuild settings
- **Logging & Colors** - Output formatting and log files
- **Screen Settings** - Display and scrolling options

#### **🎯 Key Sections**
- **Build Configuration** - Core build settings and file watching
- **Advanced Options** - Polling, rerun, and interrupt handling
- **Customization Examples** - Real-world configuration changes
- **Best Practices** - Performance tips and recommendations
- **Troubleshooting** - Common configuration issues

#### **📊 Configuration Table**
Documented all settings with current values:
- File extensions to watch: `["go", "tpl", "tmpl", "html"]`
- Excluded directories: `["assets", "tmp", "vendor", "testdata", "docker"]`
- Build delay: `1000ms` (prevents excessive rebuilds)
- Test file exclusion: `["_test.go"]`
- Color-coded logging for better visibility

#### **💡 Practical Examples**
- Adding new file types to monitor
- Custom build commands with tags
- Environment variable integration
- Performance optimization settings

#### **📋 Next Steps**
Ready to create the development setup guide next, which will include:
- Getting started instructions
- Docker integration details
- Makefile command explanations
- Step-by-step development workflow

Would you like me to continue with the development setup documentation?