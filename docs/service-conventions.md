# Service Naming Conventions

This document outlines the naming conventions and directory structure required for services in the Golang Service Boilerplate to ensure automatic detection and integration with the Makefile and build system.

## Naming Rules

### Service Name Format
- **Required Suffix**: All services must end with `-service`
- **Allowed Characters**: Lowercase letters, numbers, and hyphens
- **Examples**:
  - ✅ `user-service`
  - ✅ `auth-service`
  - ✅ `payment-service`
  - ✅ `notification-service`
  - ❌ `user_svc` (wrong suffix)
  - ❌ `UserService` (uppercase)
  - ❌ `user.service` (dot instead of hyphen)

### Directory Structure
Services must be placed in the `services/` directory with the exact service name:

```
services/
├── user-service/
├── auth-service/
└── payment-service/
```

## Automatic Detection

### Makefile Integration
The Makefile automatically detects services using these patterns:

```makefile
# Auto-detect all services
SERVICES := $(shell ls services/ | grep -E '.*-service$$' | sort)

# Auto-detect services with migrations (checks for environments.json)
SERVICES_WITH_MIGRATIONS := $(shell find services -name "migrations" -type d \
  -exec test -f {}/environments.json \; -print 2>/dev/null \
  | sed 's|/migrations||' | sed 's|services/||' | sort)
```

### Migration-Enabled Services
For services to participate in database migrations, they must have:

```
services/{service-name}/
└── migrations/
    ├── environments.json    # Required for configuration
    ├── development/         # Development migrations
    ├── staging/            # Staging migrations
    ├── production/          # Production migrations
    └── README.md           # Optional documentation
```

## create-service.sh Integration

The `create-service.sh` script enforces these conventions:

- Validates service name ends with `-service`
- Creates proper directory structure
- Generates migration configuration files
- Updates API gateway registration

### Usage
```bash
# Correct usage
./scripts/create-service.sh my-feature-service 8084

# This will create:
# - services/my-feature-service/
# - Automatic Makefile integration
# - Migration setup (if --no-db-schema not used)
```

## Benefits

### Maintenance
- No manual Makefile updates required for new services
- Consistent naming across the project
- Automatic integration with build and deployment scripts

### Development
- `make build-my-feature-service` works immediately
- Migration commands include new services automatically
- Service discovery in API gateway works out-of-the-box

### Operations
- Docker containers use predictable naming
- Monitoring and logging integrate seamlessly
- Scaling and deployment scripts work without modification

## Troubleshooting

### Service Not Detected
If a service isn't picked up by the Makefile:

1. Check name ends with `-service`
2. Verify directory exists in `services/`
3. For migrations: ensure `migrations/environments.json` exists
4. Run `make` to see if SERVICES variable includes it

### Migration Issues
If migrations don't run:

1. Verify `services/{name}/migrations/environments.json` exists
2. Ensure migration files are in environment directories (development/, staging/, production/)
3. Ensure JSON files are valid
4. Run `make db-migrate-validate SERVICE_NAME={name}`

## Route Naming Conventions

This section documents the API routing conventions to prevent Gin router conflicts.

### The Problem

Gin cannot handle two different wildcard parameters under the same path prefix. This causes a panic on startup:

```go
// CONFLICTS - causes panic:
router.GET("/objects/:id", handler1)
router.GET("/objects/:public_id", handler2)
```

Error: `panic: ':public_id' in new path '/api/v1/objects/:public_id/relationships' conflicts with existing wildcard ':id'`

### The Solution

Use an explicit literal path prefix for secondary ID types:

```go
// WORKS - separate path prefixes:
router.GET("/objects/:id", handler1)               // internal ID (integer)
router.GET("/objects/public-id/:public_id", handler2) // explicit prefix for UUID
```

Now `:id` and `:public_id` are in different path hierarchies, no conflict.

### When to Apply

| Scenario | Pattern | Example |
|----------|---------|---------|
| Resource has internal int ID + external UUID | Use `/public-id/` prefix for UUID | `/objects/:id`, `/objects/public-id/:public_id` |
| Resource uses string key | Use `:key` directly | `/relationship-types/:type_key` |
| Resource uses only UUID (no internal ID) | Use directly | No conflict, can use `:public_id` directly |

### Examples from Codebase

#### Objects (Internal ID + UUID)

```go
objectsRead.GET("/:id", objectHandler.GetByID)                              // internal ID
objectsRead.GET("/public-id/:public_id", objectHandler.GetByPublicID)      // UUID
objectsRead.GET("/:id/children", objectHandler.GetChildren)           // nested under internal ID
objectsRead.GET("/public-id/:public_id/relationships", ...)       // nested under UUID
```

#### Relationship Types (String Key)

```go
relationshipTypesRead.GET("/:type_key", ...)  // direct, no conflict
relationshipTypesRead.GET("", ...)           // list
```

#### Relationships (UUID Only)

```go
relationshipsRead.GET("/:public_id", ...)  // direct, no internal :id exists
```

### Adding New Routes

When adding new routes to a service:

1. **Check existing routes** - Look for `:id` wildcard in same path prefix
2. **If conflict exists** - Use explicit prefix (e.g., `/public-id/`)
3. **If no conflict** - Use direct parameter (e.g., `:public_id`)

### Debugging Route Conflicts

If you encounter a route conflict:

```
panic: ':parameter' in new path conflicts with existing wildcard ':existing'
```

1. Find both routes in same group
2. Rename one to use explicit prefix
3. Rebuild and test

## Migration from Old Naming

If you have services with different naming:

1. Rename the directory to match `-service` convention
2. Update any hardcoded references in configuration files
3. Test with `make dev-bootstrap`
4. Update API gateway registration if needed

## Examples

### Complete Service Structure
```
services/user-service/
├── cmd/
│   └── main.go
├── internal/
│   ├── handlers/
│   ├── services/
│   └── models/
├── migrations/
│   ├── environments.json
│   ├── development/
│   │   ├── 000001_initial.up.sql
│   │   └── 000001_initial.down.sql
│   ├── staging/
│   └── production/
├── Dockerfile
├── Dockerfile.dev
├── go.mod
└── README.md
```

### Makefile Commands Available
Once created with proper naming:
```bash
make build-user-service
make run-user-service
make test-user-service
make db-migrate-init SERVICE_NAME=user-service
make db-migrate-up SERVICE_NAME=user-service
```

## Version History

- **v1.1.0**: Introduced dynamic service detection and naming conventions
- **v1.0.0**: Initial service creation with manual Makefile updates