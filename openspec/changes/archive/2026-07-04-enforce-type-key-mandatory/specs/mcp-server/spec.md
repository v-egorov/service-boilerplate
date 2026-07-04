## MODIFIED Requirements

### Requirement: MCP server exposes list_object_types tool
The mcp-server MUST expose a tool named `list_object_types` that retrieves all object types from objects-service, returning the full type hierarchy. The tool SHALL accept optional filter parameters (`type_key_prefix`, `parent_type_id`) to narrow results. Objects-service MUST support these filter parameters and return results where child types (those with `type_key` values containing a parent namespace prefix) are included without errors.

#### Scenario: Filtered listing returns matching types only
- **WHEN** the agent calls list_object_types with a type_key prefix filter (e.g., "doc")
- **THEN** the server returns only object types whose type_key starts with the given prefix

#### Scenario: Filtered listing includes child types without error
- **WHEN** the agent calls list_object_types with no filters or with a broad prefix like "product"
- **THEN** the server returns all matching types including children (e.g., `product-electronics`, `product-clothing`) with valid non-empty type_key values

#### Scenario: Listing by parent_type_id filter works for child types
- **WHEN** the agent calls list_object_types with a valid parent_type_id parameter
- **THEN** the server returns only direct children of that parent type, each with a valid type_key (e.g., querying Product's children returns Electronics, Clothing, Books)

### Requirement: MCP server exposes get_object_type tool
The mcp-server MUST expose a tool named `get_object_type` that retrieves a single object type by its ID or type_key. The tool SHALL return the complete type definition including parent, children, metadata, and whether it is sealed. All returned types — root and child — MUST have a non-empty `type_key` field.

#### Scenario: Get child type by ID returns full type details with valid type_key
- **WHEN** the agent calls get_object_type with a valid integer ID for a child type (e.g., Electronics, id=5)
- **THEN** the server returns the matching ObjectType with all fields including a non-empty `type_key` value (e.g., `"product-electronics"`)

#### Scenario: Get object type by type_key lookup works for child types
- **WHEN** the agent calls get_object_type with a valid type_key string parameter for a child type (e.g., "product-clothing")
- **THEN** the server returns the matching ObjectType using the same fields as ID lookup

## ADDED Requirements

### Requirement: MCP list_objects respects object type hierarchy via type_key
The mcp-server MUST support filtering objects by `type_key_prefix` in addition to `object_type_id`. When a user wants all objects across a type namespace (e.g., all products and product variants), the tool SHALL resolve child type_keys matching the prefix and query them.

#### Scenario: List objects with type_key_prefix returns cross-type results
- **WHEN** the agent calls list_objects with `type_key_prefix="product"` instead of a single object_type_id
- **THEN** the server resolves all matching type_keys (e.g., "product", "product-electronics", "product-clothing") and returns objects from all those types combined

### Requirement: All MCP responses include mandatory non-empty type_key
Every MCP tool response that includes an ObjectType MUST have a non-empty `type_key` field. This is guaranteed by the database-level NOT NULL constraint on `objects_service.object_types.type_key`.

#### Scenario: Every type_key in object-type responses is non-empty
- **WHEN** any MCP tool returns an ObjectType (list_object_types, get_object_type, hierarchy resource)
- **THEN** the `type_key` field in every returned object is a non-empty string matching the pattern `[a-z][a-z0-9-]*`
