# Troubleshooting Guide

## Common Issues and Solutions

### Air Not Starting

#### Problem
```bash
air: command not found
```

#### Solutions

**Install Air globally:**
```bash
go install github.com/air-verse/air@v1.52.3
```

**Check Go version compatibility:**
```bash
go version  # Should be 1.19+
```

**Verify PATH:**
```bash
which air
echo $PATH
```

**Install in Docker container:**
```bash
docker-compose exec api-gateway which air
```

### Build Errors

#### Problem
```bash
build-errors.log contains compilation errors
```

#### Solutions

**Check build logs:**
```bash
cat api-gateway/build-errors.log
```

**View real-time build output:**
```bash
make logs
```

**Test manual build:**
```bash
cd api-gateway
go build ./cmd
```

**Check Go modules:**
```bash
go mod tidy
go mod download
```

**Verify import paths:**
```bash
go list -m all
```

### File Watching Issues

#### Problem
Air not detecting file changes

#### Solutions

**Check file permissions:**
```bash
ls -la api-gateway/
```

**Verify watched directories:**
```bash
# Check if files are in watched paths
find . -name "*.go" | head -10
```

**Test file system events:**
```bash
# On Linux
inotifywait -m . -r -e modify,create,delete
```

**Use polling mode (fallback):**
```toml
# .air.toml
poll = true
poll_interval = 1000
```

**Check exclude patterns:**
```bash
# Files matching exclude patterns are ignored
ls api-gateway/tmp/
```

### Docker Development Issues

#### Problem
Hot reload not working in Docker containers

#### Solutions

**Verify volume mounting:**
```bash
docker-compose exec api-gateway ls -la /app/api-gateway
```

**Check file permissions in container:**
```bash
docker-compose exec api-gateway id
docker-compose exec api-gateway ls -la /app
```

**Fix permission issues:**
```bash
# Fix ownership
docker-compose exec api-gateway chown -R $(id -u):$(id -g) /app

# Or rebuild containers
make build-dev
```

**Verify Air installation in container:**
```bash
docker-compose exec api-gateway air -v
```

**Check container working directory:**
```bash
docker-compose exec api-gateway pwd
```

### Port Conflicts

#### Problem
```bash
bind: address already in use
```

#### Solutions

**Check port usage:**
```bash
lsof -i :8080
netstat -tulpn | grep :8080
```

**Change ports in override file:**
```yaml
# docker/docker-compose.override.yml
services:
  api-gateway:
    ports:
      - "8081:8080"  # Change host port
```

**Stop conflicting services:**
```bash
# Find and stop process
sudo fuser -k 8080/tcp
```

### Database Connection Issues

#### Problem
Services can't connect to database in development

#### Solutions

**Check database container:**
```bash
docker-compose ps postgres
```

**Verify database logs:**
```bash
docker-compose logs postgres
```

**Test database connection:**
```bash
docker-compose exec postgres psql -U postgres -d service_db
```

**Check environment variables:**
```bash
docker-compose exec user-service env | grep DATABASE
```

**Wait for database readiness:**
```yaml
# docker/docker-compose.yml
depends_on:
  postgres:
    condition: service_healthy
```

### Memory and Performance Issues

#### Problem
High memory usage or slow rebuilds

#### Solutions

**Optimize exclude patterns:**
```toml
# .air.toml
exclude_dir = ["assets", "tmp", "vendor", "testdata", "docker", "node_modules"]
exclude_regex = ["_test.go", "_mock.go"]
```

**Reduce build delay:**
```toml
delay = 500  # milliseconds
```

**Limit watched files:**
```toml
include_ext = ["go", "html", "yaml"]
```

**Monitor resource usage:**
```bash
docker stats
```

### Environment Variable Problems

#### Problem
Configuration not loading correctly

#### Solutions

**Check environment file:**
```bash
cat .env
```

**Verify variable loading:**
```bash
docker-compose exec api-gateway env | grep APP_ENV
```

**Test with explicit variables:**
```bash
APP_ENV=development make dev
```

**Check docker-compose configuration:**
```bash
docker-compose config
```

### Log Analysis Issues

#### Problem
Can't see Air logs or build output

#### Solutions

**Check log levels:**
```bash
# In container
docker-compose exec api-gateway air -d
```

**View build error logs:**
```bash
cat api-gateway/build-errors.log
```

**Enable verbose logging:**
```toml
# .air.toml
[log]
main_only = false
time = true
```

**Check Docker logs:**
```bash
docker-compose logs -f --tail=100
```

### Service Health Check Failures

#### Problem
Health checks failing in production

#### Solutions

**Test health endpoint manually:**
```bash
curl http://localhost:8080/health
```

**Check service logs:**
```bash
docker-compose logs api-gateway
```

**Verify health check configuration:**
```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 10s
```

**Check service startup time:**
```bash
docker-compose ps
```

### Go Module Issues

#### Problem
Dependency or module resolution errors

#### Solutions

**Clean module cache:**
```bash
go clean -modcache
go mod download
```

**Update dependencies:**
```bash
go get -u ./...
go mod tidy
```

**Check for conflicting versions:**
```bash
go list -m all
```

**Rebuild with clean cache:**
```bash
go clean -cache
go build ./cmd
```

### Docker Compose Issues

#### Problem
Container orchestration problems

#### Solutions

**Check compose file syntax:**
```bash
docker-compose config
```

**Rebuild containers:**
```bash
docker-compose down
docker-compose up --build
```

**Clear Docker cache:**
```bash
docker system prune -f
```

**Check network connectivity:**
```bash
docker-compose exec api-gateway ping user-service
```

### IDE Integration Issues

#### Problem
IDE not recognizing changes or debugging not working

#### Solutions

**Check file watching in IDE:**
```bash
# VS Code: Ensure files are not excluded
# GoLand: Check file watchers
```

**Restart IDE file watchers:**
```bash
# VS Code: Reload window
# GoLand: Invalidate caches
```

**Verify build configurations:**
```bash
# Check IDE run configurations match Makefile
```

### Performance Optimization

#### Slow Rebuilds

**Solutions:**
```toml
# .air.toml - Optimize for speed
delay = 300
exclude_unchanged = true
kill_delay = "0s"
```

#### High CPU Usage

**Solutions:**
```bash
# Limit container resources
docker-compose.yml
services:
  api-gateway:
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
```

#### Large Build Context

**Solutions:**
```dockerfile
# Use .dockerignore
node_modules/
.git/
docs/
*.log
```

### Advanced Troubleshooting

#### Debug Air Configuration

```bash
# Test configuration
air -c .air.toml --help

# Debug mode
air -c .air.toml -d
```

#### Monitor File System Events

```bash
# Linux
sudo apt install inotify-tools
inotifywait -m . -r -e modify

# macOS
fswatch -r .
```

#### Network Debugging

```bash
# Test service connectivity
docker-compose exec api-gateway curl http://user-service:8081/health

# Check DNS resolution
docker-compose exec api-gateway nslookup user-service
```

#### Log Aggregation

```bash
# Collect all logs
docker-compose logs > debug.log

# Filter by time
docker-compose logs --since "1h"
```

### Emergency Recovery

#### Complete Reset

```bash
# Stop everything
make down

# Clean up
docker system prune -f
docker volume prune -f

# Rebuild from scratch
make build-dev
make dev
```

#### Force Rebuild

```bash
# Force rebuild containers
docker-compose build --no-cache

# Clear Go cache
go clean -cache -modcache
```

### Getting Help

#### Log Collection

```bash
# Collect diagnostic information
{
  echo "=== Docker Status ==="
  docker-compose ps
  echo -e "\n=== Container Logs ==="
  docker-compose logs --tail=50
  echo -e "\n=== Air Configuration ==="
  cat api-gateway/.air.toml
  echo -e "\n=== Environment ==="
  docker-compose exec api-gateway env | grep -E "(APP_ENV|LOGGING)"
} > diagnostics.log
```

#### Version Information

```bash
# Check versions
go version
docker --version
docker-compose --version
air -v
```

### Prevention Best Practices

1. **Regular Maintenance**
   ```bash
   # Clean up regularly
   docker system prune -f
   go clean -modcache
   ```

2. **Monitor Resources**
   ```bash
   # Set up monitoring
   docker stats
   ```

3. **Backup Configurations**
   ```bash
   # Keep backups of working configs
   cp .air.toml .air.toml.backup
   ```

4. **Test Changes**
   ```bash
    # Test before committing
    make build-dev
   make dev
   ```

## Related Documentation

- [Development Setup](development-setup.md)
- [Configuration](configuration.md)
- [Production Mode](production-mode.md)
- [Examples](examples.md)</content>
</xai:function_call">### 📋 **Troubleshooting Documentation Created**

I've created comprehensive troubleshooting documentation covering:

#### **🔧 Common Issues & Solutions**
- **Air Installation**: Command not found, version compatibility
- **Build Errors**: Compilation failures, module issues
- **File Watching**: Detection problems, permission issues
- **Docker Integration**: Volume mounting, container permissions
- **Port Conflicts**: Address already in use, port changes
- **Database Connections**: Container networking, dependency issues

#### **⚡ Performance Optimization**
- **Memory Usage**: High resource consumption fixes
- **Slow Rebuilds**: Delay optimization, exclude patterns
- **CPU Usage**: Resource limits, monitoring
- **Build Context**: Docker ignore files, caching

#### **🔍 Advanced Debugging**
- **Log Analysis**: Build errors, real-time monitoring
- **Environment Variables**: Configuration loading issues
- **Health Checks**: Service monitoring problems
- **Go Modules**: Dependency resolution, cache issues

#### **🐳 Docker-Specific Issues**
- **Compose Problems**: Configuration syntax, networking
- **Container Permissions**: File ownership, volume access
- **Resource Limits**: CPU and memory constraints
- **Network Debugging**: Service connectivity, DNS

#### **🛠️ IDE Integration**
- **File Watching**: IDE recognition problems
- **Debug Configuration**: Run configuration mismatches
- **Cache Issues**: IDE cache invalidation

#### **🚨 Emergency Recovery**
- **Complete Reset**: Full environment cleanup
- **Force Rebuild**: Clean container rebuilds
- **Diagnostic Collection**: Log aggregation for support

#### **📋 Next Steps**
Ready to create the final examples documentation, which will include:
- Real-world usage scenarios
- Command examples and workflows
- Configuration examples
- Integration patterns

Would you like me to continue with the examples documentation?