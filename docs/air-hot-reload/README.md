# Air Hot Reload Documentation

## Overview

This project uses [Air](https://github.com/air-verse/air) for hot reloading during development. Air is a live-reloading command line utility for developing Go applications that automatically rebuilds and restarts your application when file changes are detected.

## Key Features

- **Automatic Rebuilding**: Monitors Go files and rebuilds on changes
- **Live Restart**: Seamlessly replaces the running binary
- **Fast Compilation**: Incremental builds for quick development cycles
- **Docker Integration**: Works within Docker containers for consistent environments
- **Environment Control**: Different configurations for development vs production

## Quick Start

### Development Mode (with Hot Reload)

```bash
# Start all services with hot reload
make dev

# Or start individual services
make air-gateway      # API Gateway only
make air-user-service # User Service only
```

### Production Mode (without Hot Reload)

```bash
# Start services with pre-compiled binaries
make up
```

## Architecture

### Development Setup
- Uses `Dockerfile.dev` with Air installed
- Mounts source code as volumes
- Runs Air in watch mode
- Environment: `APP_ENV=development`

### Production Setup
- Uses standard `Dockerfile` for optimized builds
- Pre-compiled binaries
- Environment: `APP_ENV=production`

## File Structure

```
docs/air-hot-reload/
├── README.md              # This file - main overview
├── configuration.md       # Air configuration details
├── development-setup.md   # Getting started guide
├── production-mode.md     # Production deployment
├── troubleshooting.md     # Common issues & solutions
└── examples.md           # Usage scenarios

# Configuration Files
api-gateway/.air.toml
services/user-service/.air.toml

# Docker Files
api-gateway/Dockerfile.dev
services/user-service/Dockerfile.dev
docker/docker-compose.override.yml
```

## How It Works

1. **File Monitoring**: Air watches `.go` files for changes
2. **Build Trigger**: On file change, triggers `go build`
3. **Binary Replacement**: New binary replaces the running process
4. **Zero Downtime**: Process replacement happens seamlessly

## Configuration

Air is configured via `.air.toml` files in each service directory:

- `api-gateway/.air.toml` - API Gateway configuration
- `services/user-service/.air.toml` - User Service configuration

See [configuration.md](configuration.md) for detailed settings.

## Development Workflow

1. Start development environment: `make dev`
2. Make code changes in your editor
3. Air automatically detects changes and rebuilds
4. Test your changes immediately
5. View logs: `make logs`

## Environment Variables

- `APP_ENV=development` - Enables Air hot reload
- `APP_ENV=production` - Uses pre-compiled binaries
- `LOGGING_LEVEL=debug` - Enhanced logging in development

## Performance Tips

- Air excludes test files (`*_test.go`) from watching
- Build artifacts stored in `tmp/` directory
- 1-second delay prevents excessive rebuilds
- Color-coded logs for better visibility

## Troubleshooting

See [troubleshooting.md](troubleshooting.md) for common issues and solutions.

## Related Documentation

- [Development Setup](development-setup.md)
- [Configuration Details](configuration.md)
- [Production Mode](production-mode.md)
- [Examples](examples.md)
- [Troubleshooting](troubleshooting.md)</content>
</xai:function_call">### 📋 **Air Hot Reload Documentation Structure Created**

I've successfully created the documentation directory and main README file. The structure includes:

#### **📁 Directory Structure**
```
docs/air-hot-reload/
├── README.md              # Main overview and quick start
├── configuration.md       # Air configuration details  
├── development-setup.md   # Getting started guide
├── production-mode.md     # Production deployment
├── troubleshooting.md     # Common issues & solutions
└── examples.md           # Usage scenarios
```

#### **✅ Main README.md Created**
The main documentation file includes:
- **Overview** of Air hot reload functionality
- **Quick Start** commands for development and production
- **Architecture** explanation (dev vs prod setups)
- **File Structure** mapping
- **How It Works** technical explanation
- **Development Workflow** step-by-step guide
- **Performance Tips** and best practices

#### **🎯 Key Features Documented**
- Automatic rebuilding on file changes
- Docker integration for consistent environments  
- Environment-based configuration switching
- Zero-downtime binary replacement
- Color-coded logging and monitoring

#### **📋 Next Steps**
The foundation is set! Ready to create the detailed documentation files:
1. **Configuration Guide** - Deep dive into `.air.toml` settings
2. **Development Setup** - Complete getting started guide  
3. **Production Mode** - Dev vs prod differences
4. **Troubleshooting** - Common issues and solutions
5. **Examples** - Real-world usage scenarios

Would you like me to continue with creating the configuration documentation next?