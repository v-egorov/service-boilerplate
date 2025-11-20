# Service URL Configuration Guide

This guide explains the service URL configuration architecture implemented in the service-boilerplate project.

## Overview

The project uses a **hybrid approach** combining environment variables and platform service discovery for clean, maintainable service communication.

## Architecture

### Configuration Hierarchy

```
1. Environment Variables (highest priority)
   ├── AUTH_SERVICE_URL=http://localhost:8083 (development override)
   └── USER_SERVICE_URL=http://localhost:8081 (development override)

2. Platform Service Discovery (Docker/Kubernetes defaults)
   ├── AUTH_SERVICE_URL=http://auth-service:8083
   └── USER_SERVICE_URL=http://user-service:8081

3. Client-Level API Versioning
   ├── /api/v1/users (user service endpoints)
   ├── /api/v1/auth/validate-token (auth service endpoints)
   └── /public-key (auth service endpoints)
```

## Implementation

### API Gateway Configuration

```go
// Environment-based service URL configuration
authServiceURL := os.Getenv("AUTH_SERVICE_URL")
if authServiceURL == "" {
    authServiceURL = "http://auth-service:8083" // Docker service discovery default
}

userServiceURL := os.Getenv("USER_SERVICE_URL")
if userServiceURL == "" {
    userServiceURL = "http://user-service:8081" // Docker service discovery default
}

// Apply development overrides
if cfg.App.Environment == "development" && os.Getenv("DOCKER_ENV") != "true" {
    authServiceURL = strings.Replace(authServiceURL, "auth-service", "localhost", 1)
    userServiceURL = strings.Replace(userServiceURL, "user-service", "localhost", 1)
}
```

### Auth Service Configuration

```go
// Clean base URL from environment
userServiceURL := os.Getenv("USER_SERVICE_URL")
if userServiceURL == "" {
    userServiceURL = "http://user-service:8081" // Docker default
}

// Client handles API versioning
userClient := client.NewUserClient(userServiceURL, logger)
```

### User Service Client

```go
// Base URL + API versioning in client
func (c *UserClient) CreateUser(ctx context.Context, req *CreateUserRequest) (*UserData, error) {
    url := fmt.Sprintf("%s/api/v1/users", c.baseURL) // API version handled here
    // ... implementation
}
```

## Docker Compose Configuration

### Development Environment

```yaml
# docker/docker-compose.override.yml
services:
  api-gateway:
    environment:
      - AUTH_SERVICE_URL=http://auth-service:8083
      - USER_SERVICE_URL=http://user-service:8081

  auth-service:
    environment:
      - USER_SERVICE_URL=http://user-service:8081

  user-service:
    environment:
      # User service doesn't need other service URLs
```

### Production Environment

```yaml
# Production docker-compose.yml or Kubernetes manifests
services:
  api-gateway:
    environment:
      - AUTH_SERVICE_URL=http://auth-service.prod.company.com:8083
      - USER_SERVICE_URL=http://user-service.prod.company.company.com:8081

  auth-service:
    environment:
      - USER_SERVICE_URL=http://user-service.prod.company.com:8081
```

## Benefits

### ✅ Clean Architecture
- **Separation of Concerns**: URLs vs API versioning
- **Platform Agnostic**: Works with Docker, Kubernetes, or manual deployment
- **Environment Flexibility**: Easy switching between dev/staging/prod

### ✅ Maintainability
- **No Hard-coded URLs**: All service locations configurable
- **Single Source of Truth**: Environment variables control service locations
- **Deployment-Specific**: Configuration lives with deployment manifests

### ✅ Scalability
- **Dynamic Service Discovery**: Platform handles service location resolution
- **Easy Service Addition**: New services just need environment variables
- **Zero Code Changes**: Environment configuration only

### ✅ Developer Experience
- **Local Development**: Automatic localhost overrides
- **Testing**: Environment variables make testing flexible
- **Debugging**: Clear configuration hierarchy

## Migration from Config-Based Approach

### Before (Config-Based)
```yaml
# config.yaml
services:
  auth_service_url: "http://auth-service:8083"
  user_service_url: "http://user-service:8081"
```

```go
// Code
authURL := cfg.GetServiceURL("auth", "http://auth-service:8083")
```

### After (Environment-Based)
```yaml
# docker-compose.yml
environment:
  - AUTH_SERVICE_URL=http://auth-service:8083
  - USER_SERVICE_URL=http://user-service:8081
```

```go
// Code
authURL := os.Getenv("AUTH_SERVICE_URL")
if authURL == "" {
    authURL = "http://auth-service:8083"
}
```

## Best Practices

### 1. Environment Variable Naming
- Use `UPPER_SNAKE_CASE` for environment variables
- Follow pattern: `{SERVICE_NAME}_SERVICE_URL`
- Example: `AUTH_SERVICE_URL`, `USER_SERVICE_URL`

### 2. Default Values
- Always provide sensible Docker/Kubernetes defaults
- Use service names that resolve via platform discovery
- Example: `http://auth-service:8083`

### 3. Development Overrides
- Use localhost for local development
- Apply overrides based on environment detection
- Example: Replace `auth-service` with `localhost`

### 4. API Versioning
- Handle API versions in client code, not URLs
- Use consistent versioning across services
- Example: `/api/v1/` prefix in client methods

### 5. Documentation
- Document required environment variables
- Include examples for different environments
- Update deployment manifests accordingly

## Troubleshooting

### Service Connection Issues
```bash
# Check environment variables
docker exec service-boilerplate-api-gateway env | grep SERVICE_URL

# Test service connectivity
docker exec service-boilerplate-api-gateway curl http://auth-service:8083/health

# Check platform DNS resolution
docker exec service-boilerplate-api-gateway nslookup auth-service
```

### Configuration Debugging
```bash
# Verify environment variables are set
echo $AUTH_SERVICE_URL
echo $USER_SERVICE_URL

# Check service logs for connection errors
docker logs service-boilerplate-api-gateway | grep "auth-service\|user-service"
```

## Related Documentation

- [Security Architecture](./security-architecture.md) - JWT and service communication security
- [JWT Key Rotation](./jwt-key-rotation.md) - Dynamic key distribution
- [Troubleshooting](./troubleshooting-auth-logging.md) - Service communication issues
- [Service Creation Guide](./service-creation-guide.md) - Adding new services</content>
<parameter name="filePath">docs/service-url-configuration.md