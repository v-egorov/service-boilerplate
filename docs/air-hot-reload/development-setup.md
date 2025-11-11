# Development Setup Guide

## Prerequisites

Before setting up Air hot reload, ensure you have:

- Docker and Docker Compose installed
- Go 1.23+ (for local development)
- Make utility
- Git

## Quick Start

### Option 1: Full Development Environment (Recommended)

```bash
# Clone the repository
git clone <repository-url>
cd service-boilerplate

# Start all services with hot reload
make dev
```

### Option 2: Individual Services

```bash
# Start API Gateway only
make air-gateway

# Start User Service only
make air-user-service
```

### Option 3: Local Development (without Docker)

```bash
# Install Air globally
go install github.com/air-verse/air@v1.52.3

# Start API Gateway
cd api-gateway && air

# Start User Service (in another terminal)
cd services/user-service && air
```

## Environment Setup

### Environment Variables

Create a `.env` file in the project root:

```bash
# Development environment
APP_ENV=development
LOGGING_LEVEL=debug

# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/service_db?sslmode=disable

# Service ports
API_GATEWAY_PORT=8080
USER_SERVICE_PORT=8081
```

### Docker Development Environment

#### Docker Compose Override Workflow

The development setup uses **Docker Compose's override mechanism** with `docker-compose.override.yml`:

##### How Override Files Work
```bash
# Docker Compose automatically merges files:
docker-compose up
# = docker-compose.yml + docker-compose.override.yml

# Production deployment (override ignored):
docker-compose -f docker-compose.yml up
```

##### Override File Purpose
- **Development Optimizations**: Hot reload, debugging, source mounting
- **Environment Separation**: Dev settings don't affect production
- **Automatic Loading**: No special commands needed for development
- **Configuration Override**: Dev settings take precedence over base config

##### Development Override Configuration
```yaml
# docker/docker-compose.override.yml
services:
  api-gateway:
    # Development Dockerfile with Air
    build:
      dockerfile: api-gateway/Dockerfile.dev

    # Development environment
    environment:
      - APP_ENV=development
      - LOGGING_LEVEL=debug

    # Live source code mounting
    volumes:
      - ../api-gateway:/app/api-gateway
      - ${API_GATEWAY_TMP_VOLUME}:/app/api-gateway/tmp

    # Development ports
    ports:
      - "${API_GATEWAY_PORT}:${API_GATEWAY_PORT}"

    # Service discovery aliases
    networks:
      service-network:
        aliases:
          - ${API_GATEWAY_NAME}
          - gateway
          - api

  user-service:
    build:
      dockerfile: services/user-service/Dockerfile.dev
    environment:
      - APP_ENV=development
      - LOGGING_LEVEL=debug
    volumes:
      - ../services/user-service:/app/services/user-service
      - ${USER_SERVICE_TMP_VOLUME}:/app/services/user-service/tmp
    ports:
      - "${USER_SERVICE_PORT}:${USER_SERVICE_PORT}"
    networks:
      service-network:
        aliases:
          - ${USER_SERVICE_NAME}
          - users
          - user-svc

  postgres:
    environment:
      - POSTGRES_DB=${DATABASE_NAME}
      - POSTGRES_USER=${DATABASE_USER}
      - POSTGRES_PASSWORD=${DATABASE_PASSWORD}
    volumes:
      - ${POSTGRES_VOLUME}:/var/lib/postgresql/data
    networks:
      service-network:
        aliases:
          - ${POSTGRES_NAME}
          - db
          - database
```

##### Override Workflow in Action

1. **Start Development**:
   ```bash
   make dev
   # Docker Compose loads: base.yml + override.yml
   # Result: Development containers with hot reload
   ```

2. **Make Code Changes**:
   ```bash
   vim api-gateway/internal/handlers/gateway.go
   # Override mounts source → changes immediately available
   # Air detects changes → automatic rebuild
   ```

3. **Debug and Test**:
   ```bash
   # Override provides development tools
   docker-compose exec api-gateway sh
   curl http://gateway:8080/health  # Using network alias
   ```

4. **Deploy to Production**:
   ```bash
   make up
   # Only loads base.yml → production containers
   # Override file ignored → clean production deployment
   ```

##### Key Override Benefits

- **🔄 Hot Reload**: Source mounting enables live updates
- **🐛 Debugging**: Development tools and logging
- **🏷️ Service Discovery**: Network aliases for easy connectivity
- **⚙️ Environment Control**: Dev vs prod settings separation
- **📦 Volume Management**: Configurable temp and data volumes
- **🚀 Automatic**: No special commands needed

## Development Workflow

### 1. Start Development Environment

```bash
make dev
```

This command:
- Builds development Docker images with Air
- Starts all services with hot reload enabled
- Mounts source code as volumes
- Sets `APP_ENV=development`

### 2. Make Code Changes

Edit any `.go` file in your preferred editor:

```bash
# Example: Edit API Gateway handler
vim api-gateway/internal/handlers/gateway.go
```

### 3. Automatic Rebuild

Air automatically:
- Detects file changes
- Triggers `go build`
- Replaces the running binary
- Displays build status with colors

### 4. View Logs

```bash
# View all service logs
make logs

# View specific service logs
docker-compose logs -f api-gateway
docker-compose logs -f user-service
```

### 5. Test Changes

```bash
# Test API Gateway
curl http://localhost:8080/health

# Test User Service
curl http://localhost:8081/health
```

## Development Commands

### Makefile Commands

| Command | Description |
|---------|-------------|
| `make dev` | Start all services with hot reload |
| `make air-gateway` | Start API Gateway with Air |
| `make air-user-service` | Start User Service with Air |
| `make logs` | View all service logs |
| `make down` | Stop all services |
| `make build-dev` | Rebuild development images |

### Docker Commands

```bash
# View running containers
docker-compose ps

# Execute into a container
docker-compose exec api-gateway sh

# View container logs
docker-compose logs -f api-gateway

# Restart a service
docker-compose restart api-gateway
```

### Local Development Commands

```bash
# Install Air (one-time)
go install github.com/air-verse/air@v1.52.3

# Start with custom config
air -c .air.toml

# Start with debug logging
air -d

# Build without running
air -b
```

## File Structure for Development

```
service-boilerplate/
├── api-gateway/
│   ├── .air.toml           # Air configuration
│   ├── Dockerfile.dev      # Development Docker image
│   ├── cmd/main.go         # Application entry point
│   └── internal/           # Source code
├── services/user-service/
│   ├── .air.toml           # Air configuration
│   ├── Dockerfile.dev      # Development Docker image
│   ├── cmd/main.go         # Application entry point
│   └── internal/           # Source code
├── docker/
│   ├── docker-compose.yml          # Production config
│   └── docker-compose.override.yml # Development overrides
├── Makefile                 # Development commands
└── .env                     # Environment variables
```

## Docker Development Images

### Dockerfile.dev Structure

```dockerfile
# Development stage with Go for hot reloading
FROM golang:1.23-alpine

# Install curl for health checks
RUN apk --no-cache add curl

WORKDIR /app

# Install Air for hot reloading
RUN go install github.com/air-verse/air@v1.52.3

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Copy Air configuration
COPY .air.toml .

# Expose port
EXPOSE 8080

# Default command for development
CMD ["air", "-c", ".air.toml"]
```

### Key Features

- **Air Installation**: Latest version (v1.52.3)
- **Dependency Caching**: Efficient layer caching
- **Source Mounting**: Live code updates via volumes
- **Health Checks**: curl for service monitoring

## Troubleshooting Development Setup

### Common Issues

#### Air Not Starting
```bash
# Check if Air is installed
which air

# Install Air if missing
go install github.com/air-verse/air@v1.52.3
```

#### Build Errors
```bash
# Check build logs
cat api-gateway/build-errors.log

# View real-time logs
make logs
```

#### Port Conflicts
```bash
# Check port usage
lsof -i :8080

# Change ports in docker-compose.override.yml
ports:
  - "8081:8080"  # Host:Container
```

#### File Permission Issues
```bash
# Fix permissions in container
docker-compose exec api-gateway chown -R $(id -u):$(id -g) /app
```

### Performance Optimization

#### Exclude Unnecessary Files
Update `.air.toml`:
```toml
exclude_dir = ["assets", "tmp", "vendor", "testdata", "docker", "node_modules"]
```

#### Adjust Build Delay
```toml
delay = 500  # Faster rebuilds
```

#### Use Polling Mode (if file watching fails)
```toml
poll = true
poll_interval = 1000
```

## Database Setup for Development

### PostgreSQL Container

The development environment includes PostgreSQL:

```yaml
postgres:
  environment:
    - POSTGRES_DB=${DATABASE_NAME}
    - POSTGRES_USER=${DATABASE_USER}
    - POSTGRES_PASSWORD=${DATABASE_PASSWORD}
  volumes:
    - ${POSTGRES_VOLUME}:/var/lib/postgresql/data
    - ./docker/init.sql:/docker-entrypoint-initdb.d/init.sql
```

### Database Migrations with Docker Container

Migrations are handled by a dedicated migration service using Docker:

```yaml
# Migration Service (profiles: migration)
migration:
  image: ${MIGRATION_IMAGE}
  container_name: ${MIGRATION_CONTAINER_NAME}
  volumes:
    - ./services/${SERVICE_NAME}/migrations:/migrations:ro
  environment:
    - DATABASE_URL=postgres://${DATABASE_USER}:${DATABASE_PASSWORD}@${POSTGRES_NAME}:${DATABASE_PORT}/${DATABASE_NAME}?sslmode=${DATABASE_SSL_MODE}
  networks:
    - service-network
```

#### Migration Commands

```bash
# Run migrations up
make db-migrate-up

# Run migrations down
make db-migrate-down

# Check migration status
make db-migrate-status



# Create new migration
make db-migrate-create NAME=add_users_table

# Validate migrations
make db-migrate-validate

# Reset database
make db-reset
```

#### Migration Workflow

1. **Create Migration**:
   ```bash
   make db-migrate-create NAME=add_user_profiles
   # Creates: services/user-service/migrations/000002_add_user_profiles.up.sql
   ```

2. **Write Migration SQL**:
   ```sql
   -- Add new columns to users table
   ALTER TABLE users ADD COLUMN profile_image_url VARCHAR(500);
   ALTER TABLE users ADD COLUMN bio TEXT;
   ```

3. **Run Migration**:
   ```bash
   make db-migrate-up
   # Applies all pending migrations
   ```

4. **Verify Status**:
   ```bash
   make db-migrate-status
   # Shows current migration version
   ```

## IDE Integration

### VS Code

Add to `.vscode/tasks.json`:
```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Start API Gateway",
      "type": "shell",
      "command": "make",
      "args": ["air-gateway"],
      "group": "build"
    }
  ]
}
```

### GoLand/IntelliJ

- Use "Go Build" run configuration
- Set working directory to service root
- Enable "Run after build" for auto-restart

## Next Steps

- [Configuration Details](../configuration.md)
- [Production Mode](../production-mode.md)
- [Troubleshooting](../troubleshooting.md)
- [Examples](../examples.md)</content>
</xai:function_call">### 📋 **Development Setup Documentation Created**

I've created a comprehensive development setup guide covering:

#### **🚀 Quick Start Options**
- **Full Environment**: `make dev` for complete setup
- **Individual Services**: `make air-gateway` or `make air-user-service`
- **Local Development**: Direct Air usage without Docker

#### **⚙️ Environment Configuration**
- **Environment Variables**: `.env` file setup
- **Docker Overrides**: Development-specific configuration
- **Volume Mounting**: Live code updates

#### **🔄 Development Workflow**
- **Step-by-step Process**: From startup to testing
- **Automatic Rebuilds**: File change detection and compilation
- **Log Monitoring**: Real-time build and application logs
- **Health Checks**: Service verification commands

#### **🛠️ Development Commands**
- **Makefile Integration**: All development commands explained
- **Docker Commands**: Container management and debugging
- **Local Commands**: Direct Air usage options

#### **🏗️ Docker Development Setup**
- **Dockerfile.dev Analysis**: Development image structure
- **Key Features**: Air installation, dependency caching, health checks
- **Volume Configuration**: Source code mounting for live updates

#### **🔧 Troubleshooting Section**
- **Common Issues**: Air installation, build errors, port conflicts
- **Performance Tips**: File exclusion, build delay optimization
- **File Permissions**: Docker volume permission fixes

#### **💾 Database Integration**
- **PostgreSQL Setup**: Development database configuration
- **Migration Commands**: Database schema management
- **Data Persistence**: Volume configuration for data retention

#### **🖥️ IDE Integration**
- **VS Code Tasks**: Development task configuration
- **GoLand/IntelliJ**: Run configuration setup

#### **📋 Next Steps**
Ready to create the production mode documentation, which will cover:
- Production vs development differences
- Optimized Docker builds
- Binary compilation without Air
- Deployment considerations

Would you like me to continue with the production mode documentation?