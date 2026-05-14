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
- **migration-orchestrator/** - Go wrapper for managing DB migrations via golang-migrate CLI

## Development Workflow
- Builds and tests run on host machine to simplify workflow
- Air hot-reload inside containers - source changes trigger automatic rebuild/restart
- Key Makefile targets:
  - `make dev` - start services in development mode (blocks, tails logs - do not use in agentic mode)
  - `make dev-detached` - start services in development mode (detached, returns once services started)
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

## Development credentials

All accounts use password: `devadmin123`

| Email | Name | Roles |
|-------|------|-------|
| dev.admin@example.com | Dev Admin | admin, user |
| object.admin@example.com | Object Admin | object-type-admin, user |
| test.user@example.com | Test User | user |


## Database objects and development database access

**No local psql** - use Docker:
```bash
docker exec service-boilerplate-postgres psql -U postgres -d service_db
```

**Schemas:**
- `auth_service` - users, roles, permissions, JWT keys
- `objects_service` - object_types, objects, objects_relationship_types
- `user_service` - users, user_profiles, user_settings

**Always use schema-qualified table names:**
```sql
SELECT * FROM objects_service.objects;
SELECT * FROM auth_service.permissions;
```

## Database Migrations

The project uses golang-migrate to manage database schema changes.

### Migration Files Location
- Service migrations: `services/<service-name>/migrations/`
- Environment directories: `development/`, `staging/`, `production/`
- Each environment has its own sequential migration files (000001_, 000002_, etc.)
- File naming: `000001_<description>.up.sql` and `.down.sql`

### Configuration Files
Configuration is stored in ONE file:
- **environments.json** - defines migrations directory per environment and config options
- Note: dependencies.json is no longer used

### Applying Migrations

**Always run in sequence:**
1. `make db-migrate-init SERVICE_NAME=<service-name>` - initialize tracking table (run once per service)
2. `make db-migrate-up SERVICE_NAME=<service-name>` - apply all pending migrations
3. `make db-migrate-down SERVICE_NAME=<service-name>` - rollback one migration

**Or run all services at once:**
- `make db-migrate` - runs all services in correct order (auth-service → user-service → objects-service)
- `make db-migration-order` - show migration execution order

**Note:** `db-migrate-init` only needs to be run once when setting up a new service schema. Subsequent runs only need `db-migrate-up`. auth-service must run first because user-service migrations depend on auth_service schema (roles, permissions).

### Development vs Staging/Production
- Development includes test data migrations (more migrations)
- Staging and Production have fewer migrations (excludes dev-only test data)
- Each environment directory contains the appropriate migrations for that environment

### Important Rules
- NEVER apply migrations directly via psql - always use migration orchestrator
- Migrations must be numbered sequentially within each environment (001, 002, 003...)
- Each migration needs both .up.sql and .down.sql files

## Testing
- Tests use testify framework
- See `./docs/testify-overview.md` for testing approach

## Service Development Patterns

This boilerplate follows consistent patterns across all services. See detailed guides:

- [Service Patterns Reference](docs/service-patterns-reference.md) - Code examples for all layers
- [Tracing Implementation Guide](docs/tracing-implementation-guide.md) - HTTP, DB, and business tracing

### Quick Reference

| Layer | Location | Pattern |
|-------|----------|---------|
| Models | `internal/models/` | Domain struct + Request/Response DTOs |
| Repository | `internal/repository/` | Interface + schema-qualified queries |
| Service | `internal/services/` | Two constructors + validation |
| Handler | `internal/handlers/` | Error mapping + logging |
| Tests | `internal/*/*_test.go` | Manual mocks + testify |

### Future Improvements

See [Service Patterns Differences](docs/service-patterns-differences.md) for planned standardization work.

## API Response Standards

### Success Response Format

All successful API responses follow:

```json
{
  "data": { ... },
  "message": "Human-readable success message",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

**Fields:**
- `data` - The response payload (resource or collection)
- `message` - Human-readable success message (debugging/logging)
- `meta` - Machine-readable metadata
  - `request_id` - Unique request ID for distributed tracing

**Examples:**

Create object:
```json
{
  "data": { "id": 1, "name": "Object" },
  "message": "Object created successfully",
  "meta": { "request_id": "abc-123" }
}
```

Get object:
```json
{
  "data": { "id": 1, "name": "Object" },
  "meta": { "request_id": "abc-123" }
}
```

### Error Response Format

```json
{
  "error": "Human-readable error message",
  "type": "<error_type>",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

**Fields:**
- `error` - User-facing error message
- `type` - Machine-readable error type for programmatic handling
- `meta.request_id` - Unique request ID for distributed tracing (same as success responses)

### Error Type Values

| Type | HTTP Status | Description |
|------|-------------|-------------|
| `validation_error` | 400, 422 | Invalid input, missing required fields |
| `unauthorized` | 401 | Authentication failed |
| `permission_denied` | 403 | Authorization failed |
| `not_found` | 404 | Resource not found |
| `conflict` | 409 | Resource conflict (duplicate, version) |
| `internal_error` | 500 | Server error |

### Special Cases

**Field-level validation:**
```json
{
  "error": "email is required",
  "type": "validation_error",
  "field": "email",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

**Resource conflicts:**
```json
{
  "error": "User already exists",
  "type": "conflict",
  "resource": "user",
  "meta": {
    "request_id": "abc-123-xyz"
  }
}
```

### Implementation Rules

1. **Always include `type` field** - Never return errors with only `error` field
2. **Use `errors.Is()` for wrapped errors** - Don't use `==` for error comparison
3. **Never expose `details` field** - Don't include error chain/stack traces in responses
4. **HTTP status codes must match error type** - Status code is primary signal
5. **Keep messages user-friendly** - Avoid technical jargon

### References

- [Full API Response Standards](docs/api-response-standards.md)

## Git Workflow

**AI assistant makes code changes but does NOT automatically commit.**

**Typical workflow:**
1. Assistant makes code changes (build runs, tests pass)
2. Review changes with `git diff`
3. User review and stage changes with `git add <files>` or staging hunks when ready
4. Say "lets commit" to request commit message and commit

**Example:**
```bash
# Review changes
git diff

# Stage specific files (if needed)
git add services/objects-service/internal/handlers/object_handler.go

# Request commit
"lets commit" → Assistant provides commit message, then executes git commit
```

**Why this workflow:**
- **Review control** - User sees all changes before committing
- **Quality assurance** - Verify tests pass after changes
- **Commit messages** - Assistant crafts descriptive messages, user approves
- **Incremental commits** - Can stage specific files separately

## graphify

This project has a graphify knowledge graph at graphify-out/.

Rules:
- Before answering architecture or codebase questions, read graphify-out/GRAPH_REPORT.md for god nodes and community structure
- If graphify-out/wiki/index.md exists, navigate it instead of reading raw files
- For cross-module "how does X relate to Y" questions, prefer `graphify query "<question>"`, `graphify path "<A>" "<B>"`, or `graphify explain "<concept>"` over grep — these traverse the graph's EXTRACTED + INFERRED edges instead of scanning files
- After modifying code files in this session, run `graphify update .` to keep the graph current (AST-only, no API cost)
