# type_key Infrastructure and Universal Permission Routing Plan

## Status: APPROVED

Plan committed to git. Implementation will proceed one task at a time — each task fully completed and tested before moving to the next. Tasks are independent where possible, sequential where dependencies exist.

## Overview

Replace hardcoded permission strings in Gin route groups with data-driven, `type_key`-based routing so that adding a new object type requires only database rows — no code changes. Simultaneously unify the ownership model to a single `created_by == user_id` check across all types (relationships included).

**What becomes data-driven:** Permission checking middleware constructs permission strings from `RouteConfig{TypeKey, HTTPMethod}` instead of hardcoded permission names like `"objects:read:own"`.

**What stays type-specific and manual:** Each object type still needs its own route group in `main.go`, its own handler (`*_handler.go`), its own service, its own repository. This does NOT change — there is no dynamic URL generation or universal handler.

## Context

### Current State (from analysis at docs/analysis/)
- **CTI pattern**: Relationships ARE objects — they share the same base table, use `object_id` PK/FK into concrete tables (`objects_relationships`, `objects_relationship_types`)
- **Permissions contradiction**: The architecture doc treats relationships as "special" with 3 distinct ownership models, but CTI establishes them as regular object descendants
- **Critical bugs identified** (15 total):
  - Relationships can't be accessed at all — routes use flat permission names (`relationships:create`) that don't match scoped names in DB (`relationships:create:own`/`:all`) → always returns 403
  - Object.List leaks ALL objects to users with only `:own` scope (no filtering by created_by)
  - Relationship handler never sets `created_by` — all relationships get `"system"` as creator
  - No ownership checks on any relationship endpoints
  - `handleServiceError` uses exact comparison (`==`) instead of `errors.Is()` → 404s return as 500

### What type_key does today vs what it should do
- **Today**: The architecture doc mentions `type_key` in the architecture, but the column doesn't exist on `object_types`. The permission system uses hardcoded string prefixes (`objects`, `relationships`, `object-types`) — there's no data-driven mapping from object_type_id to permission namespace.
- **Target**: Each registered type has a `type_key` (e.g., `"portfolio"`, `"relationships"`). Permission strings are constructed as `<type_key>:<action>[:scope]`. Routes don't hardcode permission names; they derive them from the route's type context at runtime.

### Scope boundaries
- **In scope**: Schema change (`type_key` column), seed data, model/repository updates, universal route/permission middleware (replaces hardcoded strings with `RouteConfig{TypeKey, HTTPMethod}`), unified ownership checks, auth-service consistency fixes (action column + tracing)
- **Out of scope**: Database migrations re-creation, test suite rewrite, gateway changes, JWT/token infrastructure changes, endpoint ownership validation for relationship creation, dynamic URL generation, universal type-specific handlers

## What is data-driven vs what stays manual

| Layer | Dynamic? | Example |
|-------|----------|---------|
| **Middleware (permission check)** | Yes — uses `type_key` to construct permission strings | Same middleware code for all types: `RouteConfig{TypeKey: "portfolio", HTTPMethod: "GET"}` → checks `portfolio:read:all`, `portfolio:read:own` |
| **Routes in main.go** | No — each type still needs its own route group wired manually | `v1.Group("/portfolios")` added by hand, like `/objects` and `/relationships` today |
| **Handlers, models, services** | No — remain type-specific per type | `portfolio_handler.go`, `portfolio_service.go`, etc. Same pattern as `relationship_handler.go` today |

### How route groups work after this change

Current (hardcoded permission strings):
```go
// main.go - each route group has its own path and hardcoded permissions:
objectsRead := v1.Group("/objects")
objectsRead.Use(permissionMiddleware("objects:read:all", "objects:read:own"))
{
    objectsRead.GET("/:id", objectHandler.GetByID)
}

relationshipsRead := v1.Group("/relationships")
relationshipsRead.Use(permissionMiddleware("relationships:read"))  // flat — doesn't match DB!
{
    relationshipsRead.GET("", relationshipHandler.List)
}
```

After (data-driven middleware, same route structure):
```go
// main.go - same paths and handlers; only the middleware call changes:
objectsRead := v1.Group("/objects")
objectsRead.Use(permissionMiddleware(RouteConfig{TypeKey: "objects", HTTPMethod: "GET"}))
{
    objectsRead.GET("/:id", objectHandler.GetByID)
}

relationshipsRead := v1.Group("/relationships")
relationshipsRead.Use(permissionMiddleware(RouteConfig{TypeKey: "relationships", HTTPMethod: "GET"}))
{
    relationshipsRead.GET("", relationshipHandler.List)
}
```

When adding a new type (e.g., `portfolio`), you still manually add routes in main.go — but the middleware call becomes uniform and predictable. The plan does NOT attempt to auto-discover or auto-register types at runtime.

## Tasks

### Task 1: Schema — add type_key to object_types
**File:** `services/objects-service/migrations/{development,staging,production}/000010_add_type_key.up.sql` (and `.down.sql`)
- Add `type_key VARCHAR(100)` column to `object_types` table (nullable initially)
- Down migration: drop column if it was empty

### Task 2: Seed existing type_keys
**File:** `services/objects-service/migrations/{development,staging,production}/000011_seed_type_keys.up.sql` (and `.down.sql`)
- Update existing rows with their type_key values based on name mapping:
  - `"object-types"` → `type_key = "object-types"`
  - `"Objects"` → `type_key = "objects"`
  - `"RelationshipType"` → `type_key = "relationship-types"`
  - `"Relationship"` → `type_key = "relationships"`
- Add unique constraint on `type_key` (after seeding)

### Task 3: Update ObjectType model and repository queries
**Files:**
- `services/objects-service/internal/models/object_type.go` — add `TypeKey string` field with JSON tag
- `services/objects-service/internal/repository/object_type_repository.go` — update all SELECT/INSERT/UPDATE queries to include `type_key`; ensure creation handles it

### Task 4: Create permission middleware (replaces existing permiddleware entirely, no backward compat needed)
**Files:**
- `services/objects-service/internal/permiddleware/universal_permission.go` (new file)
- This replaces the current per-route hardcoded approach. The middleware receives a route config that specifies: the HTTP method and type_key. It dynamically constructs permission strings like `type_key + ":" + action + ":scope"`.

**Design:**
```go
// RouteConfig holds runtime-per-route configuration. TypeKey is required — every object type has one.
type RouteConfig struct {
    TypeKey      string // e.g. "relationships", "objects" (required, never empty)
    HTTPMethod   string // "GET", "POST", "PUT", "DELETE"
}

// permissionMiddleware creates a gin.HandlerFunc from config.
// Replaces the existing permiddleware.NewPermissionMiddleware entirely — no backward compatibility needed.
func permissionMiddleware(cfg RouteConfig, authClient client.AuthClient, logger *logrus.Logger) gin.HandlerFunc { ... }
```

**Permission mapping:**
| HTTP Method | Action | Permission strings checked (OR logic) |
|-------------|--------|--------------------------------------|
| POST        | create | `{type_key}:create` (flat — you own what you create) |
| GET         | read   | `{type_key}:read:all`, `{type_key}:read:own` |
| PUT         | update | `{type_key}:update:all`, `{type_key}:update:own` |
| DELETE      | delete | `{type_key}:delete:all`, `{type_key}:delete:own` |

**How type_key is resolved at runtime:**
- Each route group specifies its `TypeKey`. For objects endpoints, it's `"objects"`. For relationship endpoints, it's `"relationships"`. This comes from the route config in main.go — not hardcoded permission strings.
- The middleware queries auth-service with dynamically constructed permission string(s).

### Task 5: Refactor main.go routes to use new permission middleware
**File:** `services/objects-service/cmd/main.go`
- Replace all `permissionMiddleware("<hardcoded-string>")` calls with `permissionMiddleware(RouteConfig{TypeKey: "...", HTTPMethod: "..."})`
- Remove hardcoded permission strings entirely — only type_key and method remain as route config

### Task 6: Fix relationship handler created_by assignment
**File:** `services/objects-service/internal/handlers/relationship_handler.go`
- In `Create()`, set `req.CreatedBy = middleware.GetAuthenticatedUserID(c)` before calling service (currently missing — all relationships get `"system"`)
- Add `checkOwnership()` method to RelationshipHandler (identical logic to ObjectHandler.checkOwnership) for Update/Delete endpoints

### Task 7: Unified ownership model across handlers and repos
**Files:**
- `services/objects-service/internal/handlers/object_handler.go` — extract `checkOwnership` into a shared function or embeddable struct; ensure it uses the same pattern as relationship handler's new checkOwnership
- `services/objects-service/internal/repository/object_repository.go` — in `List()`, accept optional `created_by` filter and apply when user matched `:own` scope (fixes BUG-2)
- `services/objects-service/internal/handlers/object_handler.go` — pass ownership scope to repo based on which permission matched (`:own` vs `:all`)
- `services/objects-service/internal/repository/relationship_repository.go` — add optional `created_by` filter parameter to `List()` and apply when `:own` scope is active

### Task 8: Fix handleServiceError to use errors.Is()
**File:** `services/objects-service/internal/handlers/object_handler.go`
- Change exact comparison (`err == repository.ErrOptimisticLock`) to `errors.Is(err, ...)` so wrapped service-layer errors match correctly (fixes BUG-5 / 404→500 mapping)

### Task 9: Fix bulk operations permission model
**File:** `services/objects-service/cmd/main.go`
- Split the single bulk route group into three separate groups:
  - POST `/bulk` → only requires `{type_key}:create`
  - PUT `/bulk` → requires `{type_key}:update:all` OR `{type_key}:update:own`
  - DELETE `/bulk` → requires `{type_key}:delete:all` OR `{type_key}:delete:own`

### Task 10: Auth-service — fix action column consistency + tracing
**Files:**
- `services/auth-service/migrations/{development,staging,production}/000012_fix_action_column.up.sql` (and `.down.sql`) — standardize the `action` column to store only base action (`read`, `create`, etc.) for all permissions; strip scope suffixes from objects scoped permissions that currently have it baked in
- `services/auth-service/internal/repository/auth_repository.go` — add `database.TraceDBQuery()` wrapper around the `GetUserPermissions` query (fixes ISSUE-15)

### Task 11: Update architecture documentation
**Files:**
- `docs/object-type-permissions-architecture.md` — remove "three ownership models" section for relationships; clarify endpoint ownership is service-layer validation, not permission logic; add type_key → resource mapping table
- `docs/object-type-permissions-implementation-plan.md` — mark completed phases, note the shift to universal data-driven approach

### Task 12: Create "How To Add A New Object Type" guide (in docs/)
**File:** `docs/adding-new-object-types-guide.md` (new document)

Step-by-step procedure for two scenarios:

#### Scenario A: No CTI table needed (simple types with JSONB metadata only)

These are object types whose data fits entirely in the base `objects` table + JSONB metadata. Example: a "Document" type that needs name, description, status, and arbitrary metadata fields stored as JSONB.

**Steps:**
1. **Register the type in auth-service permissions** (`auth_service.permissions`):
   - Insert permission records for each action (e.g., `document:create`, `document:read:all`, `document:read:own`, etc.)
   - Assign to roles via `auth_service.role_permissions` (which roles get which permissions)

2. **Add the type_key to object_types** (`object_types`):
   ```sql
   INSERT INTO objects_service.object_types (name, description, is_sealed, metadata, created_at, updated_at)
   VALUES ('Document', 'Simple document type with JSONB metadata', false, '{}', NOW(), NOW());

   -- Update the newly inserted row:
   UPDATE object_types SET type_key = 'document' WHERE name = 'Document';
   ```

3. **Create model** (`internal/models/document.go`):
   - Define `Document` struct with base fields from `Object` (name, description, status, metadata JSONB) plus any Document-specific fields that fit in the base table
   - Add request/response DTOs for Create/Update/List operations

4. **Create repository** (`internal/repository/document_repository.go`):
   - Implement CRUD queries against `objects_service.objects` where `object_type_id = (SELECT id FROM object_types WHERE type_key = 'document')`
   - No CTI table joins needed — all data lives in base table + metadata JSONB

5. **Create service** (`internal/services/document_service.go`):
   - Wire repository, add business validation
   - Uses the same ownership model as objects: `created_by == userID` for `:own` checks

6. **Create handler** (`internal/handlers/document_handler.go`):
   - Implement Create/GetByID/List/Update/Delete handlers
   - Reuse shared `checkOwnership()` utility from Task 7
   - Map service errors using shared error mapping pattern (like `handleServiceError`)

7. **Wire in main.go**:
   ```go
   documentRepo := repository.NewDocumentRepository(pgDatabase, repoOptions)
   documentService := services.NewDocumentService(documentRepo)
   documentHandler := handlers.NewDocumentHandler(documentService, logger.Logger)

   documentsGroup := v1.Group("/documents")
   documentsRead := v1.Group("/documents")
   // ... more groups for create/update/delete as needed

   documentsRead.Use(permissionMiddleware(RouteConfig{TypeKey: "document", HTTPMethod: "GET"}))
   {
       documentsRead.GET("/:id", documentHandler.GetByID)
       documentsRead.GET("", documentHandler.List)
   }
   
   documentsGroup.Use(permissionMiddleware(RouteConfig{TypeKey: "document", HTTPMethod: "POST"}))
   {
       // POST routes for create, etc.
   }
   ```

8. **Add to objects-service main.go service initialization** (same pattern as today):
   - Add `documentRepo`, `documentService`, `documentHandler` to the init block where they're created from repositories
   - Wire route groups in the router section

#### Scenario B: CTI table needed (types with type-specific columns)

These are object types that need their own concrete table alongside the base objects table, following the same pattern as relationships today. Example: a "Product" type needing `sku`, `price`, `inventory_count` as native SQL columns (not JSONB).

**Steps:**
1. **Create DB migration for CTI table**:
   ```sql
   -- In services/objects-service/migrations/{env}/NNNNN_create_products_cti.up.sql
   CREATE TABLE objects_service.products (
       object_id BIGINT PRIMARY KEY REFERENCES objects_service.objects(id) ON DELETE CASCADE,
       sku VARCHAR(100) NOT NULL UNIQUE,
       price DECIMAL(10,2) NOT NULL,
       inventory_count INTEGER DEFAULT 0,
       created_by VARCHAR(255),
       updated_by VARCHAR(255),
       created_at TIMESTAMPTZ DEFAULT NOW(),
       updated_at TIMESTAMPTZ DEFAULT NOW()
   );

   COMMENT ON TABLE objects_service.products IS 'CTI concrete table for Product instances';
   
   CREATE INDEX idx_products_sku ON products(sku);
   ```

2. **Register the type in auth-service permissions** (same as Scenario A)

3. **Add type_key to object_types** (same as Scenario A, plus potentially `concrete_table_name = "products"`)

4. **Create base model** (`internal/models/product.go`):
   - Define `Product` struct with both base fields (`Object` embedded or composed) and CTI-specific fields
   - Base model uses the shared `Object` type for common fields
   - Request/response DTOs include both base and CTI-specific fields

5. **Create repository** (`internal/repository/product_repository.go`):
   - CRUD operations use CTE pattern: join `objects_service.objects` with `products` via `object_id`
   - Example query pattern (same as relationship_repository today):
     ```sql
     WITH product_data AS (
         SELECT 
             o.id, o.public_id, o.name, o.description, o.status, o.metadata,
             p.sku, p.price, p.inventory_count
         FROM objects_service.objects o
         INNER JOIN products_service.products p ON o.id = p.object_id
         WHERE o.object_type_id = (SELECT id FROM object_types WHERE type_key = 'product')
           AND o.deleted_at IS NULL
     )
     SELECT * FROM product_data;
     ```
   - Create: insert into `objects` first → get `object_id` → insert into CTI table using same ID
   - Delete: delete from CTI table → then delete from objects (on cascade)

6. **Create service** (`internal/services/product_service.go`):
   - Wire both product repository and object repository (for base object operations)
   - Business validation for CTI-specific constraints
   - Same ownership model as all other types: `created_by == userID`

7. **Create handler** (`internal/handlers/product_handler.go`):
   - Implement Create/GetByID/List/Update/Delete handlers with CTE-aware responses
   - Reuse shared `checkOwnership()` utility from Task 7
   - Map service errors using shared error mapping pattern

8. **Wire in main.go** (same structure as Scenario A, but with CTI-aware handler)

## Development Approach
- Testing: use testify framework as per docs/testify-overview.md; run `make test-<service-name>` after each change
- Use Air hot-reload for services (automatic rebuild/restart on file changes)
- Log files at `./docker/volumes/[service-name]/logs/[service-name].log`
- Make small, focused changes — complete each task fully before moving to the next
- CRITICAL: All tests must pass before starting next task
- Update this plan when scope changes during implementation

## Progress Tracking
- [x] Task 1: Schema migration — add type_key column (000010) ✅
- [ ] Task 2: Seed existing type_keys (000011)
- [ ] Task 3: Update ObjectType model and repository queries
- [ ] Task 4: Create permission middleware (replaces existing permiddleware entirely, no backward compat needed)
- [ ] Task 5: Refactor main.go routes to use new permission middleware
- [ ] Task 6: Fix relationship handler created_by assignment + checkOwnership
- [ ] Task 7: Unified ownership — List filtering by created_by for objects and relationships
- [ ] Task 8: Fix handleServiceError → errors.Is()
- [ ] Task 9: Fix bulk operations permission model
- [ ] Task 10: Auth-service fixes (action column + tracing)
- [ ] Task 11: Update architecture documentation
- [ ] Task 12: Create "How To Add A New Object Type" guide

## Acceptance Criteria
1. `type_key` column exists on `object_types` with all 4 existing types seeded and unique constraint applied
2. No hardcoded permission strings remain in main.go routes — all resolved via universal middleware from type_key + HTTP method
3. Relationship endpoints work correctly (no more 403 due to flat vs scoped mismatch)
4. Object.List filters by `created_by` when user has only `:own` scope verified
5. Relationships set `created_by` = authenticated user on creation
6. All relationship ownership checks use unified model (`created_by == userID`) — no special-cased endpoint ownership in permission layer
7. `handleServiceError` correctly maps repository errors to proper HTTP status codes using `errors.Is()`
8. Bulk operations use per-endpoint permissions instead of requiring ALL simultaneously
9. Auth-service action column is consistent; GetUserPermissions includes distributed tracing
10. Guide document exists with clear step-by-step instructions for both scenarios (with and without CTI tables)

## Post-completion
- Manual verification: start services with `make dev`, test relationship CRUD through API gateway, verify ownership filtering works correctly for both objects and relationships
- User inspection for database changes (migration review)
- Run full test suite: `make test-auth-service` and `make test-objects-service`

**Out of scope notes**: Changes to the API gateway (`api-gateway/`) or existing test suites are explicitly out of scope. During manual verification, the gateway or tests may throw errors due to permission format changes — these will be identified but fixed in a separate increment with their own plan.

## Notes
- The universal middleware approach means adding a new object type requires two manual steps: wiring routes/handler in main.go (same as today), plus using the data-driven permission middleware instead of hardcoded strings. Permission checking itself becomes automatic based on `type_key`.
- This preserves relationship-specific roles (`relationship-admin`, `relationship-viewer`) while removing the architectural contradiction of "relationships are special objects that need different permission rules"
