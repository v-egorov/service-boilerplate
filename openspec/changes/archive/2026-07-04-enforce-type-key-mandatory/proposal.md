# Proposal: Enforce type_key as Mandatory Identifier for Object Types

## Why

The `type_key` column was introduced in June 2026 as the canonical string identifier (handle) for object types, used by permission middleware to construct scoped permission strings (`{type_key}:{action}:scope`). However, only root-level types were seeded — child types created during development have `type_key IS NULL`, causing:

1. **pgx v5 crashes**: Repository queries scan `type_key` into a non-pointer `string`, and pgx refuses to convert SQL NULL → Go string (`ScanArgError`). Every API endpoint that returns object type data (List, GetByID, Search, Tree, Children) fails with 500 for any result set containing child types.
2. **Broken permissions**: Permission middleware requires a non-empty `type_key` in every route config. Child types cannot participate in the permission system because they have no identifier.
3. **API contract gap**: The request model validates `TypeKey: "required"` on creation, but there's no database-level enforcement (column is nullable), so existing data violates the intended invariant.

The June 2026 implementation left the column nullable and only seeded 6 of 15 rows. This change completes that work by enforcing the NOT NULL constraint and populating all remaining type_keys using a deterministic naming convention.

## What Changes

- **Migration**: Populate `type_key` for all child types (`Electronics`, `Clothing`, `Books`, `News Article`, `Blog Post`, `Tutorial`, `Country`, `City`) using name-based lookups with the convention `<parent_type_key>-<child_name_slugified>`. Convert column to NOT NULL.
- **Migration**: Apply same population logic across all environments (dev, staging, production).
- **API contract**: `type_key` becomes a mandatory field in create responses and appears as a required identifier in all API responses for object types.
- **Filter support**: Add `TypeKeyPrefix` filter to `ObjectTypeFilter` so clients can query by type namespace.

## Capabilities

### Modified Capabilities

- `mcp-server`: MCP tools (`list_object_types`, `get_object_type`) will return valid `type_key` values for all types (currently crashes on child types).
- `auth-service-permission-resolution`: Permission middleware routing depends on non-empty `type_key`; fixing data ensures permission checks work for routes that access child-type objects.

## Impact

| Area | Changes |
|------|---------|
| **Migrations** | Dev migration `000018` — populates NULL type_keys for child types and converts column to NOT NULL. Staging/prod migrations deferred to a separate change.
| **Database schema** | `type_key` column changes from nullable to NOT NULL |
| **Repository layer** | No code changes needed — queries already SELECT type_key, crash is purely a data issue |
| **Model layer** | No code changes needed — `TypeKey string` model + `required` validation on create request are correct |
| **API responses** | Object type responses will now include valid `type_key` for all rows (was empty string for NULL) |
| **MCP server** | `list_object_types()` and `get_object_type()` tools will work for all types |
| **Objects-service tests** | Existing integration tests may need test data updated if they reference child type IDs |
