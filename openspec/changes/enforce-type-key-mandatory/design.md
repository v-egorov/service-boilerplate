# Design: Enforce type_key as Mandatory Identifier for Object Types

## Context

The `type_key` column was added to `objects_service.object_types` in June 2026 (commit `d9ab677`) as the canonical string identifier for object types. It is used by permission middleware (`permiddleware/permission.go`) to construct scoped permission strings like `{type_key}:read:all`. The plan stated "every object type has one" — but only root-level types were seeded.

**Current DB state (15 rows):**
| Group | Count | type_key status | Names |
|-------|-------|-----------------|-------|
| Root types | 6 | ✅ populated | Category, Product, Article, Location, RelationshipType, Relationship |
| Child types | 8 | ❌ NULL | Electronics, Clothing, Books, News Article, Blog Post, Tutorial, Country, City |
| Test data | 1 | ✅ populated | Test Type |

The column has a UNIQUE constraint (`object_types_type_key_unique`) but PostgreSQL allows multiple NULLs under unique constraints. The repository layer scans `type_key` into a non-pointer Go `string`, so pgx v5 crashes with `ScanArgError` when any row has SQL NULL.

**Affected endpoints (all return 500 when result set contains child types):**
- `GET /api/v1/object-types` — List()
- `GET /api/v1/object-types/:id` — GetByID()
- `GET /api/v1/object-types/name/:name` — GetByName()
- `GET /api/v1/object-types/:id/tree` — GetTree()
- `GET /api/v1/object-types/:id/children` — GetChildren()
- `GET /api/v1/object-types/:id/descendants` — GetDescendants()

## Goals / Non-Goals

**Goals:**
- Populate `type_key` for all 8 child types using deterministic naming convention
- Convert column to NOT NULL so the invariant is enforced at DB level
- Populate dev environment with correct type_key values (staging/prod migrations deferred to a separate change)
- Add `TypeKeyPrefix` filter to `ObjectTypeFilter` for querying by namespace

**Non-Goals:**
- Renaming existing type_keys (all 6 root types already have correct names)
- Changing the permission middleware routing logic (it uses hardcoded TypeKey in RouteConfig per route group, not runtime type_key lookup)
- Adding auto-generation of type_key on create — require from caller (validated as `required,min=1,max=100,alphanumascii`)

**Naming convention examples:**
| Parent type | Child name | Generated `type_key` |
|-------------|-----------|---------------------|
| Product (`product`) | Electronics | `product-electronics` |
| Product (`product`) | Clothing | `product-clothing` |
| Product (`product`) | Books | `product-books` |
| Article (`article`) | News Article | `article-news-article` |
| Article (`article`) | Blog Post | `article-blog-post` |
| Article (`article`) | Tutorial | `article-tutorial` |
| Location (`location`) | Country | `location-country` |
| Location (`location`) | City | `location-city` |

## Decisions

### Decision 1: Naming convention for child type_keys

**Choice:** `<parent_type_key>-<child_name_slugified>`

Examples:
| Parent | Child | Generated type_key |
|--------|-------|-------------------|
| Product (product) | Electronics | `product-electronics` |
| Product (product) | Clothing | `product-clothing` |
| Article (article) | News Article | `article-news-article` |
| Location (location) | Country | `location-country` |

**Rationale:**
- Uses the parent's type_key as a namespace prefix, maintaining hierarchical organization
- Slugified child name keeps it readable and unique within each parent
- Deterministic — same input always produces same output, making migrations idempotent
- Follows existing convention: root types use single-word names (`category`, `product`), children add suffix

**Alternatives considered:**
1. **Just the child slug** (`electronics`, `clothing`) — Rejected: loses hierarchical context; could conflict if different parents had children with same name.
2. **Inherit parent's type_key** (all children under Product share `product`) — Rejected: violates unique constraint; each type needs its own identifier for permission strings and API identity.
3. **Auto-increment numeric suffix** (`product-1`, `product-2`) — Rejected: fragile, not human-readable, breaks with re-seeding.

### Decision 2: Migration approach (name-based lookups)

**Choice:** All UPDATEs use `WHERE name = 'ChildName'` (not ID references).

**Choice:** All UPDATEs use `WHERE name = 'ChildName'` (not ID references).

```sql
-- Example migration pattern
UPDATE objects_service.object_types SET type_key = 'product-electronics' WHERE name = 'Electronics';
UPDATE objects_service.object_types SET type_key = 'product-clothing'   WHERE name = 'Clothing';
-- ... etc.
ALTER TABLE objects_service.object_types ALTER COLUMN type_key SET NOT NULL;
```

**Rationale:**
- Names are stable across environments and re-seeding events
- IDs differ between dev (15 rows) and prod (fewer rows), so ID-based lookups would break
- Follows the pattern established in migration `000011` which correctly used name-based lookups for root types

**Alternatives considered:**
1. **Self-referencing subquery** (`SET type_key = parent.type_key || '-' || LOWER(translate(name, ' ', '-'))`) — Rejected: adds complexity; explicit UPDATEs are more transparent and easier to audit/rollback.
2. **CTE with recursive join** — Rejected: over-engineered for 8 rows.

### Decision 3: NOT NULL enforcement via migration (not model change)

**Choice:** Enforce at DB level with `ALTER COLUMN ... SET NOT NULL`. No Go model changes needed since the field is already a non-pointer `string` and validation on create request is already `required`.

```sql
-- Step 1: Populate all NULLs first
UPDATE ... WHERE type_key IS NULL;

-- Step 2: Then enforce NOT NULL (will fail if any NULL remains)
ALTER TABLE objects_service.object_types ALTER COLUMN type_key SET NOT NULL;
```

**Rationale:**
- The Go model already uses `TypeKey string` (non-pointer), so pgx scanning requires non-NULL values anyway
- DB-level NOT NULL provides the strongest guarantee — even direct psql inserts would fail
- No code changes needed in repository, handler, or service layers
- Request validation (`binding:"required"`) + DB constraint = defense in depth

### Decision 4: Migration numbering

**Choice:** One migration pair for dev environment only (staging/prod deferred).

| Environment | Latest migration | New migration number | File name |
|-------------|-----------------|---------------------|-----------|
| Development | 000012 (test data) | 000018 | `000018_enforce_type_key_not_null.up.sql` |

**Rationale:**
- Dev is the only active environment; staging/prod migrations will be handled in a separate change
- Next available number after 000017 (the existing draft migration) is 000018

### Decision 5: Filter support — add TypeKeyPrefix to ObjectTypeFilter

**Choice:** Add a new optional filter field `TypeKeyPrefix string` to `ObjectTypeFilter` and pass it through to repository queries. This enables MCP tools (`list_object_types`) to filter by namespace prefix as documented in the spec.

```go
type ObjectTypeFilter struct {
    // ... existing fields ...
    TypeKeyPrefix string `json:"type_key_prefix,omitempty" form:"type_key_prefix"`
}
```

**Rationale:**
- The mcp-server spec already documents a `type_key_prefix` filter parameter for `list_object_types`, but it's not implemented in the objects-service API layer
- Adding this is a natural extension of existing filter infrastructure (ILIKE prefix match)
- Low risk — purely additive, no behavior change for existing callers

## Risks / Trade-offs

| Risk | Impact | Mitigation |
|------|--------|------------|
| Migration fails if unexpected NULLs exist | All environments blocked | Migration runs in a transaction; rollback on failure. Pre-check query to count NULLs before applying. |
| Naming convention conflicts with future child types | Manual type_key assignment needed | Convention is documented in `adding-new-object-types-guide.md`; new child types should follow the pattern explicitly. If conflict arises, migration provides correct value and developer can override via API. |
| NOT NULL constraint breaks existing code that assumes nullable TypeKey | Low risk — model field is already non-pointer string; all callers expect a string value | No code change needed. The crash was caused by NULL data, not nullable type. |
| Staging/production have fewer rows than dev | Migration UPDATEs are no-ops for missing names | All UPDATEs use `WHERE name = 'X'` — silently skip non-existent rows (same as 000011 pattern). NOT NULL only affects rows that exist. |

## Migration Plan

### Development (only)
1. Rewrite existing migration `000017_populate_object_types_type_key` with name-based lookups and NOT NULL enforcement
2. Run migration orchestrator: `make db-migrate-up SERVICE_NAME=objects-service`
3. Verify all 8 child types now have type_key values (query returns 0 NULLs)
4. Re-run test script: `/scripts/test-objects-api.sh` — expect 6 previously-failing endpoints to return 200 OK

> **Note:** Staging and production migrations are deferred to a separate change once the dev migration is validated.

### Rollback Strategy
The migration includes a `.down.sql` that:
1. Sets all populated type_keys back to NULL (using name-based WHERE clauses)
2. `ALTER COLUMN ... DROP NOT NULL`
3. **WARNING**: After rollback, the same pgx crash behavior returns — this is a one-way fix until data is repopulated

## Open Questions

1. **`concrete_table_name` purpose:** Currently nullable and unused by permission routing or repository queries. It is a placeholder column designed for future CTI support where a subtype would have its own table (similar to how `objects_service.objects_relationship_types` exists as a separate table from the base `objects`). For types whose instances fit the base schema, it should eventually be populated with `'objects'`. Not mandatory for this change — deferred to future hardening.
