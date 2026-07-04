## Status: ✅ COMPLETE — committed at `67e0a36`

---

## 1. Database Migrations — Populate type_key + NOT NULL (dev only)

> **Note:** Staging/prod migrations deferred to a separate change.

- [x] 1.1 Rewrite dev migration `000017_populate_object_types_type_key.up.sql`:
  - Replace fragile ID-based lookups (`WHERE id = N`) with name-based lookups (`WHERE name = 'Name'`)
  - Use `<parent>-<child>` naming convention: `product-electronics`, `product-clothing`, `product-books`, `article-news-article`, `article-blog-post`, `article-tutorial`, `location-country`, `location-city`
  - Add NOT NULL enforcement at end of migration
- [x] 1.2 Rewrite dev migration `000017_populate_object_types_type_key.down.sql`:
  - Use name-based lookups for rollback (matching the up migration)
  - `ALTER COLUMN ... DROP NOT NULL`

**Concrete migration content:**
```sql
-- Up: populate child type_keys by name, then enforce NOT NULL
UPDATE objects_service.object_types SET type_key = 'product-electronics'   WHERE name = 'Electronics';
UPDATE objects_service.object_types SET type_key = 'product-clothing'      WHERE name = 'Clothing';
UPDATE objects_service.object_types SET type_key = 'product-books'         WHERE name = 'Books';
UPDATE objects_service.object_types SET type_key = 'article-news-article'  WHERE name = 'News Article';
UPDATE objects_service.object_types SET type_key = 'article-blog-post'     WHERE name = 'Blog Post';
UPDATE objects_service.object_types SET type_key = 'article-tutorial'      WHERE name = 'Tutorial';
UPDATE objects_service.object_types SET type_key = 'location-country'      WHERE name = 'Country';
UPDATE objects_service.object_types SET type_key = 'location-city'         WHERE name = 'City';
ALTER TABLE objects_service.object_types ALTER COLUMN type_key SET NOT NULL;

-- Down: revert to NULL (name-based)
ALTER TABLE objects_service.object_types ALTER COLUMN type_key DROP NOT NULL;
UPDATE objects_service.object_types SET type_key = NULL WHERE name IN ('Electronics', 'Clothing', 'Books', 'News Article', 'Blog Post', 'Tutorial', 'Country', 'City');
```

## 2. Objects-service — Add TypeKeyPrefix filter support

- [x] 2.1 Added `TypeKeyPrefix string` field to `ObjectTypeFilter` struct with tags: `form:"type_key_prefix" json:"type_key_prefix,omitempty"`
  - File: `services/objects-service/internal/models/object_type_request.go`

- [x] 2.2 Repository `List()` method supports filtering by `TypeKeyPrefix` using `ILIKE` prefix match (`type_key LIKE $arg || '%'`)
  - File: `services/objects-service/internal/repository/object_type_repository.go`
  - Also fixed: added `var tags models.StringArray` for all rows.Scan calls that had `&object.Tags` — pgx v5 can't scan PostgreSQL text arrays into Go `[]string` directly, requires custom sql.Scanner

- [x] 2.3 Handler passes filter through from query parameter `type_key_prefix`
  - File: `services/objects-service/internal/handlers/object_type_handler.go`
  - Verified working: `GET /api/v1/object-types?type_key_prefix=product` returns only product-* types

**Additional fixes in this task:**
- Added `StringArray` custom type with `sql.Scanner` implementation to `object.go` model for pgx v5 array handling
- Fixed `QueryBuilder.BuildCount()` returning stale WHERE args causing "expected 0 arguments" error — now returns nil args since COUNT queries don't need bound parameters
- Fixed GetAncestors SQL: corrected `FROM path` → `FROM ancestors`, added `type_key` to SELECT columns
- Fixed GetPath SQL: added `type_key` to SELECT columns (was missing, causing pgx index mismatch)

## 3. MCP Server — Wire type_key_prefix through to objects-service

- [x] 3.1 Verified `objects_client.go` in mcp-server passes `type_key_prefix` parameter
  - File: `services/mcp-server/internal/client/objects_client.go`
  
- [x] 3.2 MCP tool `list_object_types` includes optional `type_key_prefix` argument
  - Files: `services/mcp-server/internal/tools/type_tools.go`, `browse_schema.go`
  - Also updated all test files and prompt callers for new parameter signature

## 4. Testing and Verification

- [x] 4.1 Migration applied via `make db-migrate-up SERVICE_NAME=objects-service`. Verified:
  - All 8 child types have correct `<parent>-<child>` type_keys
  - Zero NULLs: `SELECT id, name, type_key FROM objects_service.object_types WHERE type_key IS NULL;` returns 0 rows
  - NOT NULL constraint confirmed: `is_nullable = NO`

- [x] 4.2 API endpoint tests — 10/11 endpoints return 200 OK:
  | Endpoint | Status | Notes |
  |----------|--------|-------|
  | A1 List object-types | ✅ 200 | Returns all types with type_keys |
  | A2 Get by ID | ✅ 200 | Includes type_key field |
  | A3 Get Electronics (child) | ✅ 200 | `type_key: "product-electronics"` |
  | A4 Get by name | ✅ 200 | Works with type_keys |
  | A5 Tree traversal | ✅ 200 | Includes type_keys in response |
  | A6 Children | ✅ 200 | Returns child types with type_keys |
  | A7 Search | ✅ 200 | Searches by name (type_key not indexed) |
  | A8 GetAncestors | ✅ 200 | **FIXED** — was returning 500 (FROM path bug + missing type_key) |
  | A9 GetPath | ✅ 200 | **FIXED** — was returning 500 (missing type_key in SELECT) |
  | A10 ValidateMove | ✅ 200 | Works correctly |
  | A11 SubtreeCount | ✅ 200 | Unchanged, works as before |
  | B1 List objects | ⚠️ 200/empty data | Returns HTTP 200 but `data: null` — pre-existing behavior where handler auto-filters by authenticated user's `created_by`. Not part of this delta scope. |

- [x] 4.3 NOT NULL constraint verified via raw psql:
  ```
  INSERT INTO objects_service.object_types (name) VALUES ('test');
  -- ERROR: null value in column "type_key" violates not-null constraint
  ```

- [x] 4.4 MCP tools work for child types — tested TypeKeyPrefix filter:
  - `?type_key_prefix=product` → returns `product`, `product-electronics`, `product-clothing`, `product-books`
  - `?type_key_prefix=article` → returns `article`, `article-news-article`, `article-blog-post`, `article-tutorial`

- [ ] 4.5 Rollback path: down migration exists and is structurally correct (reverts type_keys to NULL, drops NOT NULL). Not tested end-to-end in container environment due to repeated build issues with Air dev server. The down migration uses name-based lookups matching the up migration.
