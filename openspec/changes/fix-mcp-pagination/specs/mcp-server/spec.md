# Delta: mcp-server pagination fix

## MODIFIED Requirements

### Requirement: MCP server exposes list_objects tool
The mcp-server MUST expose a tool named `list_objects` that retrieves objects from objects-service filtered by object type. The tool SHALL accept at minimum a required object_type_id parameter and optional pagination parameters (limit, offset). The limit parameter specifies the maximum number of results to return; the offset parameter specifies how many results to skip. These match the objects-service API contract (c.Query("limit"), c.Query("offset")).

#### Scenario: List objects by type returns paginated results
- **WHEN** the agent calls list_objects with a valid object_type_id
- **THEN** the server returns a list of Object instances from objects-service including ID, public_id, name, status, tags, metadata, and associated ObjectType

#### Scenario: List objects respects pagination parameters
- **WHEN** the agent calls list_objects with limit=20 and offset=0
- **THEN** the server returns at most 20 results starting from the first matching object with pagination metadata (total, limit, offset) in the response

### Requirement: MCP server declares output schemas on all tools
Every MCP tool MUST declare its output schema using mcp-go's `WithOutputSchema[T]()` option, where T is a Go struct matching the expected structuredContent shape. This enables client-side validation via Python SDK and accurate tool metadata from `tools/list`.

#### Scenario: tools/list includes outputSchema for each tool
- **WHEN** an MCP client calls tools/list
- **THEN** every returned tool object includes an `"outputSchema"` field with a valid JSON Schema describing the structuredContent shape

#### Scenario: Output schemas match actual return types
- **WHEN** any tool is called and returns structuredContent
- **THEN** the structuredContent value conforms to that tool's declared outputSchema

### Requirement: MCP server structuredContent conforms to MCP spec object type
The mcp-server MUST serialize `structuredContent` as a JSON object (map/dict), never as a bare array. When tools return collections, the result MUST be wrapped in an object with domain-specific keys so that structuredContent is always a valid JSON object. For list_objects, this includes both the objects array and pagination metadata.

#### Scenario: list_object_types returns structuredContent as wrapped array
- **WHEN** the agent calls list_object_types and receives a CallToolResult
- **THEN** result.structuredContent is an object with key `"types"` containing the array of ObjectType entries (e.g., `{"types": [{id:1,...}, {id:2,...}]}`)

#### Scenario: get_object_type returns structuredContent as wrapped single object
- **WHEN** the agent calls get_object_type and receives a CallToolResult
- **THEN** result.structuredContent is an object with key `"item"` containing the ObjectType entry (e.g., `{"item": {id:1,...}}`) — a JSON object, not wrapped in an array

#### Scenario: list_objects returns structuredContent as wrapped array with pagination metadata
- **WHEN** the agent calls list_objects and receives a CallToolResult
- **THEN** result.structuredContent is an object with key `"objects"` containing the paginated Object results AND a `"pagination"` key with total, limit, offset (e.g., `{"objects": [{id:1,...}, ...], "pagination": {"total": 56, "limit": 20, "offset": 0}}`)

#### Scenario: get_object returns structuredContent as wrapped single object
- **WHEN** the agent calls get_object and receives a CallToolResult
- **THEN** result.structuredContent is an object with key `"object"` containing the Object entry (e.g., `{"object": {id:1,...}}`) — a JSON object, not wrapped in an array
