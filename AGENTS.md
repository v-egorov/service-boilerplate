- brevity is good
- project uses Docker Compose, compose files located at ./docker/ directory.
- there are two compose files:
  - ./docker/docker-compose.yml - base compose
  - ./docker/docker-compose.override.yml - development environment oriented compose, overriding and extending base compose
- this is a boilerplate project intended to be used as a base to develop 'real' project via forking. Consequently:
  - assume that we are working in 'development' mode all the time.
  - no need to worry about backward compatibility and breaking changes.
  - production environment planned for derivative projects, but does not exists for this boilerplate
  - user can re-create dev environment from scratch at any time
- project consists of
  - api-gateway located at ./api-gateway/
  - set of services located at ./services/
  - template to instantiate new service located at ./templates/service-template/ - do not try to compile, vet, test or fix service template directly - it will not work
  - set of utilities scripts located as ./scripts/
  - script to instantiate new service - ./scripts/create-service.sh
  - custom ./migration-orchestrator/ implemented as docker container with go application to manage migrations
- services running as containers in Docker, communicate via compose network, do not try to run services outside of docker
- builds and tests, however, can be run on host machine to simplify development workflow
- project uses Air based hot-reload - so when source is changed, corresponding service[s] compiled and restarted automatically
- project extensively uses Makefile for pretty much all common tasks. Most important targets are:
  - `make dev` - start all services in development mode. Air is started inside containers, compile and start each service.
  - `make down` - stop all services
  - `make logs` - stream logs from all containers. Warning: amount of output from this command can be fairly large, use with caution
  - `make status` - output status of services
  - `make build-<service-name>` - build service-name
  - `make test-<service-name>` - run tests for service-name
- each service produces it's own log, accessible via mounted folder on a host: ./docker/volumes/[service-name]/logs/[service-name].log - these logs available for inspection via tail/grep etc
- if after changing any YAML files (including docker-compose\*.yml) you'll encounter errors - stop and transfer control to user for inspection
## Database Migrations

The project uses migration-orchestrator (golang-migrate + custom tracking) to manage database schema changes.

### Migration Files Location
- Service migrations: `services/<service-name>/migrations/`
- Development migrations: `services/<service-name>/migrations/development/`
- Naming: `NNNN_dev_<description>.up.sql` and `.down.sql`

### Configuration Files (required for each new migration)
Each new migration requires updates to TWO files in the service's migrations folder:
1. **dependencies.json** - add migration entry with description, depends_on, affects_tables, estimated_duration, risk_level, rollback_safe, environment
2. **environments.json** - add migration file path to target environment's migrations array

### Applying Migrations
- `make db-migrate SERVICE_NAME=<service-name>` - apply all pending migrations
- `make db-migrate-down SERVICE_NAME=<service-name>` - rollback one migration

### Important Rules
- NEVER apply migrations directly via psql - always use migration orchestrator
- Always use migration orchestrator for both applying and rolling back
- Configuration-level data should be seeded via migrations, not direct SQL

## Tests

- tests should be implemented using testify framework. Overview of testing approach for project described in ./docs/testify-overview.md
