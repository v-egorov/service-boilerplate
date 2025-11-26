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

# Auto-detect services with migrations
SERVICES_WITH_MIGRATIONS := $(shell find services -name "migrations" -type d \
  -exec test -f {}/dependencies.json \; -print 2>/dev/null \
  | sed 's|/migrations||' | sed 's|services/||' | sort)
```

### Migration-Enabled Services
For services to participate in database migrations, they must have:

```
services/{service-name}/
└── migrations/
    ├── dependencies.json    # Required for detection
    ├── environments.json    # Required for orchestrator
    ├── README.md           # Optional documentation
    └── *.sql               # Migration files
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
3. For migrations: ensure `migrations/dependencies.json` exists
4. Run `make` to see if SERVICES variable includes it

### Migration Issues
If migrations don't run:

1. Verify `services/{name}/migrations/dependencies.json` exists
2. Check `environments.json` is present
3. Ensure JSON files are valid
4. Run `make db-migrate-validate SERVICE_NAME={name}`

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
│   ├── dependencies.json
│   ├── environments.json
│   ├── 000001_initial.up.sql
│   └── 000001_initial.down.sql
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