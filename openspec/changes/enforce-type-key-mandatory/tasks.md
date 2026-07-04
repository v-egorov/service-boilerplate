## 1. Database Migrations — Populate type_key + NOT NULL (dev only)

> **Note:** Staging/prod migrations deferred to a separate change.

- [ ] 1.1 Rewrite dev migration `000017_populate_object_types_type_key.up.sql`:
  - Replace fragile ID-based lookups (`WHERE id = N`) with name-based lookups (`WHERE name = 'Name'`)
  - Use `<parent>-<child>` naming convention: `product-electronics`, `product-clothing`, `product-books`, `article-news-article`, `article-blog-post`, `article-tutorial`, `location-country`, `location-city`
  - Add NOT NULL enforcement at end of migration
- [ ] 1.2 Rewrite dev migration `000017_populate_object_types_type_key.down.sql`:
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

- [ ] 2.1 Add `TypeKeyPrefix string` field to `ObjectTypeFilter` struct in `services/objects-service/internal/models/object_type_request.go`. Include validation tag: `form:"type_key_prefix" json:"type_key_prefix,omitempty"`.

- [ ] 2.2 Update repository `List()` method in `object_type_repository.go` to support filtering by `TypeKeyPrefix` using `ILIKE` prefix match (`type_key LIKE $arg || '%'`).

- [ ] 2.3 Update handler to pass the filter through from query parameters (check if this is already wired up or needs new binding).

## 3. MCP Server — Wire type_key_prefix through to objects-service

- [ ] 3.1 Verify `objects_client.go` in mcp-server passes `type_key_prefix` parameter when calling objects-service `/api/v1/object-types` endpoint. If not present, add it.

- [ ] 3.2 Update MCP tool `list_object_types` arguments to include optional `type_key_prefix` string parameter (verify tool definition matches spec).

## 4. Testing and Verification

- [ ] 4.1 Run migration orchestrator for dev: verify all 8 child types now have non-empty type_key values. Query: `SELECT id, name, type_key FROM objects_service.object_types WHERE type_key IS NULL;` should return 0 rows.

- [ ] 4.2 Re-run `/scripts/test-objects-api.sh` — expect all 14/14 endpoints to pass (previously 6 were failing with pgx crashes).

- [ ] 4.3 Test NOT NULL constraint: attempt a raw psql INSERT without type_key and verify it fails with `null value in column "type_key" violates not-null constraint`.

- [ ] 4.4 Verify MCP tools work for child types — test that `list_object_types` returns `product-electronics`, `article-blog-post`, etc. without errors.

- [ ] 4.5 Test rollback path: run `.down.sql` migrations, verify type_keys revert to NULL and pgx crashes return (confirming rollback works as expected). Then re-run up migration to restore state.
