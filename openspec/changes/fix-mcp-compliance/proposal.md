## Why

The mcp-server's CallToolResult serialization is non-compliant with the MCP specification. When tools return collections (e.g., `list_object_types`, `list_objects`), `structuredContent` is serialized as a JSON array, but the spec requires it to be a "JSON object" — Python SDK 1.26.0 enforces this with `dict[str, Any] | None` and rejects arrays. Additionally, no output schemas are declared on any tool, preventing clients from validating results and mcp-go's optional output validation from functioning.

## What Changes

- **Wrap array returns in an object with domain-specific keys**: Tools returning collections wrap their data — `{"types": [...]}` for list_object_types and `{"objects": [...]}` for list_objects — instead of a bare array, conforming to the spec requirement that structuredContent be a JSON object.
- **Add output schemas**: Each tool declares its output schema using mcp-go's `WithOutputSchema[T]()`, enabling server-side validation (opt-in via `WithOutputSchemaValidation`) and giving clients accurate type information from `tools/list`.
- **Fix `list_object_types` filter params**: Add missing `parent_type_id` parameter to the MCP tool, matching spec requirement.
- **Fix `list_objects` cross-type query**: Implement `type_key_prefix` filtering on list_objects, resolving objects across a type namespace as specified.

## Capabilities

### Modified Capabilities
- `mcp-server`: structuredContent serialization (array → object wrapper), output schemas on all 4 tools, additional filter parameters for list_object_types and list_objects

## Impact

- **Code**: `services/mcp-server/internal/tools/type_tools.go`, `object_tools.go` — tool handlers, return types, schema declarations
- **Spec**: `openspec/specs/mcp-server/spec.md` — delta spec with MODIFIED requirements
- **Clients**: Python MCP SDK 1.26.0+ will correctly parse structuredContent; Go mcp-go clients unaffected (already accepted arrays)
- **Breaking**: The wire format of `structuredContent` changes from `[...]` to `{"types": [...]}` or `{"objects": [...]}` (domain-specific keys) — any client directly accessing structuredContent without the wrapper will need updating. This is a compliance fix, not an API change.
