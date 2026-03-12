# Service Boilerplate

brevity is good

## Project Overview
- This is a boilerplate project for building microservices
- Always assume development mode (no production exists yet)
- User can re-create dev environment from scratch at any time

## Docker & Compose
- Compose files: `./docker/docker-compose.yml` (base) and `./docker/docker-compose.override.yml` (dev overrides)
- Services run as containers, communicate via compose network
- If YAML changes cause errors - stop and transfer control to user

## Project Structure
- **api-gateway/** - API Gateway service
- **services/** - Microservices (auth-service, objects-service, user-service, etc.)
- **templates/service-template/** - Template for new services (do not compile/test directly)
- **scripts/** - Utility scripts including `create-service.sh` to instantiate new services
- **migration-orchestrator/** - Custom Go application for managing DB migrations

## Development Workflow
- Builds and tests run on host machine to simplify workflow
- Air hot-reload inside containers - source changes trigger automatic rebuild/restart
- Key Makefile targets:
  - `make dev` - start services in development mode (blocks, tails logs - do not use in agentic mode)
  - `make dev-detached` - start services in development mode (detached, returns immediately)
  - `make down` - stop all services
  - `make logs` - stream logs from all containers (can be large)
  - `make status` - output status of services
  - `make build-<service-name>` - build specific service
  - `make test-<service-name>` - run tests for specific service

## Services
- Each service produces logs in: `./docker/volumes/[service-name]/logs/[service-name].log`
- Access logs via: `tail -f ./docker/volumes/[service-name]/logs/[service-name].log`
- Each service has `config.yaml` in its root directory

## Authentication & Authorization
- **API Gateway**: Validates JWT, forwards user identity via X-User-ID, X-User-Email, X-User-Roles headers
- **Internal Services**: Trust gateway headers when JWT secret is nil (development mode)
- **Common Middleware**: `common/middleware/auth.go` handles header parsing

## Database Migrations

The project uses migration-orchestrator (golang-migrate + custom tracking) to manage database schema changes.

### Migration Files Location
- Service migrations: `services/<service-name>/migrations/`
- Development migrations: `services/<service-name>/migrations/development/`
- Naming: `NNNN_dev_<description>.up.sql` and `.down.sql`

### Configuration Files (required for each new migration)
Each new migration requires updates to TWO files:
1. **dependencies.json** - migration entry with description, depends_on, affects_tables, estimated_duration, risk_level, rollback_safe, environment
2. **environments.json** - add migration file path to target environment's migrations array

### Applying Migrations
- `make db-migrate SERVICE_NAME=<service-name>` - apply all pending migrations
- `make db-migrate-down SERVICE_NAME=<service-name>` - rollback one migration

### Important Rules
- NEVER apply migrations directly via psql - always use migration orchestrator
- Configuration-level data should be seeded via migrations, not direct SQL

## Testing
- Tests use testify framework
- See `./docs/testify-overview.md` for testing approach
