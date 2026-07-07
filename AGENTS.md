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
- **CRITICAL: Never switch `.air.toml` `poll = false` to `poll = true`.** Inotify works correctly — all 5 services use it. Suspected "Air didn't see the change" is always a code-level bug; verify by running the binary directly (`./tmp/<service-name>`) before touching Air config.
- **To trigger an Air rebuild during development**, use `touch -m <file.go>` (the `-m` flag performs a metadata write syscall that Docker volume mounts translate into inotify IN_MODIFY events). Plain `touch` does NOT trigger Air — it only updates stat info without the kernel notification path.
- **`touch` does NOT trigger Air rebuilds.** Docker volume mounts don't translate metadata-only changes into inotify write events. Use `echo "" >> file.go` (actual content write) or open/save in vim to force a rebuild.
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

## Error Handling Conventions

Prefer typed domain errors over generic wrapped errors for business logic. Use `errors.Is()` to match sentinel errors and type assertions to inspect structured error fields.

**Rule of thumb:**
- **Typed struct errors** (`ValidationError`, `ConflictError`, etc.) — use when the handler needs structured data (field name, resource ID) beyond just "what went wrong"
- **Sentinel errors** (`var ErrNotFound = errors.New(...)`) — use for simple cross-package signaling where only the error type matters (used with `errors.Is()`)
- **`fmt.Errorf("...: %w", err)` wrapping** — use discretely, one level per layer. Never wrap more than once per call stack frame; let the outermost layer provide context

Example at each layer:
- Repository returns sentinel or typed error for domain conditions (`ErrNotFound`, `ValidationError`)
- Service wraps with `fmt.Errorf("failed to create user: %w", err)` — one level of wrapping, not two
- Handler matches via `errors.Is()` for sentinels, type switch/`errors.As()` for typed errors

**Anti-patterns:**
- `fmt.Errorf("something failed: %w", fmt.Errorf("inner: %w", err))` — double-wrapping in one call chain
- Mixing `==` comparison with wrapped errors (`err == fmt.Errorf(...)`) — always use `errors.Is()`

## Planning & Implementation Workflow

For systematic feature development, see [Development Workflow](docs/workflow.md).

**Quick reference:**
- Plans stored in `docs/plans/YYYY-MM-DD-<feature-name>.md`
- Use plan template with Status, Context, Tasks, Development Approach, Progress Tracking
- Review plans with revdiff (automatic via plugin or manual with `/revdiff`)
- Track progress with checkboxes `[x]` and status updates
- See [workflow.md](docs/workflow.md) for complete documentation

## Git Workflow

**AI assistant makes code changes but does NOT automatically commit.**

**Typical workflow:**
1. Assistant makes code changes (build runs, tests pass)
2. Review changes with `git diff`
3. User review and stage changes with `git add <files>` or staging hunks when ready
4. Say "lets commit" or "now please commit" to request a commit message and commit

**Example:**
```bash
# Review changes
git diff

# Stage specific files (if needed)
git add services/objects-service/internal/handlers/object_handler.go

# Request commit
"lets commit" → Assistant checks `git diff --cached`, crafts message, executes `git commit`
```

**When "lets commit" is requested, the assistant should:**
1. Check only staged changes via `git diff --cached --stat` and `git diff --cached`
2. Craft a commit message based on what's actually staged
3. Execute `git commit` — nothing more

**The assistant never stages any files during a "lets commit" request, regardless of unstaged changes.** It never alters which hunks or files are included in a commit. All staging decisions are the user's responsibility.

**Why this workflow:**
- **Review control** - User sees all changes before committing
- **Quality assurance** - Verify tests pass after changes
- **Commit messages** - Assistant crafts descriptive messages, user approves
- **Incremental commits** - Can stage specific files separately

## Compaction

When system provides a compaction summary (e.g., "What did we do so far?"), always write it out to `docs/compaction/` with a timestamped filename before responding. The file should contain the full session state: goal, progress, key decisions, next steps, and relevant files.

Rules:
- Create folder if missing: `mkdir -p docs/compaction`
- Filename format: `YYYYMMDD-HHmmss.md` (use current UTC timestamp)
- Write the compaction message verbatim — do not summarize or truncate it
- This preserves session state for later recall and audit

## OpenSpec

OpenSpec is the change management workflow for this project. All operations use the CLI.

**Key rules:**
- Always use `npx openspec` prefix (never bare `openspec` or direct file manipulation)
- Use `--json` with `status`, `validate`, etc. when you need programmatic output (for checking artifact completion, parsing task status)
- Delta specs in archives serve as historical reference — they document what changed at that point in time

