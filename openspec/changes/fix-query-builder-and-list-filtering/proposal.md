## Why

Two pre-existing bugs in objects-service cause persistent 500 Internal Server Error responses and incorrect empty results when listing objects:

1. **QueryBuilder.Where() hardcodes `$1` placeholder** — Every call to `.Where("column = $1", value)` injects a literal `$1`, regardless of how many WHERE clauses already exist. When both `object_type_id` and user ID filters are active (the common case for authenticated API calls), PostgreSQL resolves BOTH placeholders to the same argument (`arg[0]`), comparing a varchar column against a bigint value → SQLSTATE 42883 crash.

2. **List handler unconditionally self-filters by user ID** — The List endpoint always adds `WHERE created_by = <authenticated_user_id>` regardless of whether the caller requested it. MCP agents with read-all permissions get zero results because no objects are created by the agent's UUID. This is a design issue: permission checks should be handled at the gateway level (permiddleware), not silently filtered in the repository layer.

3. **QueryBuilder.WhereTagsContain() has identical branches and wrong argIndex increment** — Code paths for `i==0` and `i>0` are identical, and `argIndex += len(tags)` instead of `++` causes placeholder indices to jump incorrectly when combining tags with other WHERE clauses.

These bugs make the objects-service `/api/v1/objects` endpoint crash whenever any two filters combine (object_type_id + user ID being the most common), breaking both direct API usage and MCP server tool calls.

## What Changes

- **QueryBuilder.Where()** — Replace hardcoded `$1` with dynamic placeholder indices starting from `argIndex`, matching the pattern already used by WhereJsonContains, WhereDateRange, and other methods
- **QueryBuilder.WhereTagsContain()** — Fix identical branches to use OR logic for subsequent tags; fix argIndex increment to `++` instead of `+= len(tags)`
- **List handlers (object + relationship)** — Remove unconditional self-filtering by user ID. Permission enforcement remains at the gateway level via permiddleware

## Capabilities

### Modified Capabilities
- `objects-service`: add requirements that QueryBuilder WHERE clauses use distinct placeholder indices and List endpoints do not apply implicit ownership filtering

### New Capabilities
(none)

## Impact

- **Code:** `services/objects-service/internal/repository/database.go` (~20 lines changed), `services/objects-service/internal/handlers/object_handler.go` (~5 lines removed), `services/objects-service/internal/handlers/relationship_handler.go` (~3 lines removed)
- **Untouched:** migrations, Docker config, other services, MCP server, API gateway
- **Tests:** existing repository tests need updating (QueryBuilder test with multiple Where calls); new tests for multi-filter queries
- **Dev mode only impact:** the self-filter removal changes default List behavior — clients relying on implicit ownership filtering will now see all objects. No known consumers rely on this behavior in dev mode.
