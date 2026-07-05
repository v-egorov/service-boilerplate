## ADDED Requirements

### Requirement: MCP server structuredContent conforms to MCP spec object type
The mcp-server MUST serialize `structuredContent` as a JSON object (map/dict), never as a bare array. When tools return collections, the result MUST be wrapped in an object with an `"items"` key, e.g., `{"items": [...]}`. This ensures compatibility with the MCP specification's requirement that structuredContent is "a JSON object" and with Python SDK 1.26.0+ which enforces `dict[str, Any] | None`.

#### Scenario: list_object_types returns structuredContent as wrapped array
- **WHEN** the agent calls list_object_types and receives a CallToolResult
- **THEN** result.structuredContent is an object with key `"items"` containing the array of ObjectType entries (e.g., `{"items": [{id:1,...}, {id:2,...}]}`)

#### Scenario: get_object_type returns structuredContent as single object
- **WHEN** the agent calls get_object_type and receives a CallToolResult
- **THEN** result.structuredContent is the ObjectType entry directly (a JSON object, not wrapped in an array)

#### Scenario: list_objects returns structuredContent as wrapped array
- **WHEN** the agent calls list_objects and receives a CallToolResult
- **THEN** result.structuredContent is an object with key `"items"` containing the paginated Object results

#### Scenario: get_object returns structuredContent as single object
- **WHEN** the agent calls get_object and receives a CallToolResult
- **THEN** result.structuredContent is the Object entry directly (a JSON object)

### Requirement: MCP server declares output schemas on all tools
Every MCP tool MUST declare its output schema using mcp-go's `WithOutputSchema[T]()` option, where T is a Go struct matching the expected structuredContent shape. This enables client-side validation via Python SDK and accurate tool metadata from `tools/list`.

#### Scenario: tools/list includes outputSchema for each tool
- **WHEN** an MCP client calls tools/list
- **THEN** every returned tool object includes an `"outputSchema"` field with a valid JSON Schema describing the structuredContent shape

#### Scenario: Output schemas match actual return types
- **WHEN** any tool is called and returns structuredContent
- **THEN** the structuredContent value conforms to that tool's declared outputSchema
