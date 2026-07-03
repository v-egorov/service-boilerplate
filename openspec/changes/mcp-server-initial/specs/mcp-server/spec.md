## ADDED Requirements

### Requirement: MCP server exposes list_object_types tool
The mcp-server MUST expose a tool named `list_object_types` that retrieves all object types from objects-service, returning the full type hierarchy. The tool SHALL accept optional filter parameters (type_key prefix, parent_type_id) to narrow results.

#### Scenario: List all object types returns complete hierarchy
- **WHEN** the agent calls list_object_types with no filters
- **THEN** the server returns all object types from objects-service including their ID, name, type_key, parent relationships, and metadata in MCP tool result format

#### Scenario: Filtered listing returns matching types only
- **WHEN** the agent calls list_object_types with a type_key prefix filter (e.g., "doc")
- **THEN** the server returns only object types whose type_key starts with the given prefix

### Requirement: MCP server exposes get_object_type tool
The mcp-server MUST expose a tool named `get_object_type` that retrieves a single object type by its ID or type_key. The tool SHALL return the complete type definition including parent, children, metadata, and whether it is sealed.

#### Scenario: Get object type by ID returns full type details
- **WHEN** the agent calls get_object_type with a valid integer ID parameter
- **THEN** the server returns the matching ObjectType from objects-service with all fields (ID, name, type_key, parent_type_id, description, is_sealed, metadata)

#### Scenario: Get object type by type_key returns same result
- **WHEN** the agent calls get_object_type with a valid type_key string parameter
- **THEN** the server returns the matching ObjectType from objects-service using the same fields as ID lookup

#### Scenario: Get non-existent type key returns error
- **WHEN** the agent calls get_object_type with an invalid or missing type_key and no valid ID
- **THEN** the server returns an MCP tool result indicating the type was not found

### Requirement: MCP server exposes list_objects tool
The mcp-server MUST expose a tool named `list_objects` that retrieves objects from objects-service filtered by object type. The tool SHALL accept at minimum a required object_type_id parameter and optional pagination parameters (page, page_size).

#### Scenario: List objects by type returns paginated results
- **WHEN** the agent calls list_objects with a valid object_type_id
- **THEN** the server returns a list of Object instances from objects-service including ID, public_id, name, status, tags, metadata, and associated ObjectType

#### Scenario: List objects respects pagination parameters
- **WHEN** the agent calls list_objects with page=1 and page_size=20
- **THEN** the server returns at most 20 results starting from the first matching object with correct pagination metadata in the response

### Requirement: MCP server exposes get_object tool
The mcp-server MUST expose a tool named `get_object` that retrieves a single object by its public_id. The tool SHALL return the complete object including all fields and associated ObjectType.

#### Scenario: Get object by public_id returns full details
- **WHEN** the agent calls get_object with a valid UUID public_id parameter
- **THEN** the server returns the matching Object from objects-service with all fields (ID, public_id, name, description, metadata, tags, status, version, created_by, updated_by) and its associated ObjectType

#### Scenario: Get non-existent object returns error
- **WHEN** the agent calls get_object with an invalid or missing public_id
- **THEN** the server returns an MCP tool result indicating the object was not found

### Requirement: MCP server registers objects-types resource
The mcp-server MUST register a resource URI template `objects-types` that provides the full object type hierarchy as structured data. The resource SHALL return all ObjectType entries with their parent-child relationships, suitable for agents to browse the schema.

#### Scenario: Read objects-types resource returns complete hierarchy
- **WHEN** an agent reads the `objects-types` resource URI
- **THEN** the server returns a JSON representation of the full object type tree from objects-service including all types with their hierarchical parent-child relationships

### Requirement: MCP server exposes browse_schema prompt template
The mcp-server MUST provide a prompt template named `browse_schema` that combines list_object_types and get_object_type to let agents explore the schema incrementally. The prompt SHALL accept optional parameters for filtering (e.g., type_key_prefix) and return structured messages suitable as agent conversation context.

#### Scenario: Browse schema returns initial exploration message
- **WHEN** an agent invokes the browse_schema prompt with no arguments
- **THEN** the server returns a message instructing the agent on available object types and offering next-step tool calls (get_object_type for specific types)

### Requirement: MCP server exposes get_object_info prompt template
The mcp-server MUST provide a prompt template named `get_object_info` that combines list_objects and get_object to let agents explore objects within a type. The prompt SHALL accept an object_type_id parameter and return structured messages guiding the agent toward object-level queries.

#### Scenario: Get object info for a type returns exploration guidance
- **WHEN** an agent invokes the get_object_info prompt with a valid object_type_id
- **THEN** the server returns a message summarizing available objects of that type and offering next-step tool calls (get_object for specific instances)

### Requirement: MCP server uses SSE transport via api-gateway
The mcp-server MUST expose an SSE endpoint at `/mcp/sse` that MCP clients connect to. The api-gateway MUST route incoming requests from the path prefix `/mcp/*` to the mcp-server service using standard HTTP reverse proxying with chunked transfer encoding support for SSE streams.

#### Scenario: Client establishes SSE connection through gateway
- **WHEN** an MCP client sends GET /mcp/sse through api-gateway
- **THEN** api-gateway forwards the request to mcp-server, which responds with HTTP 200 and Content-Type: text/event-stream, providing an endpoint event with the message submission URL

#### Scenario: Tool call reaches objects-service via gateway-mcp pipeline
- **WHEN** a client POSTs a tool_call request through api-gateway to the MCP message endpoint
- **THEN** the request is forwarded to mcp-server which executes the tool and returns the result via SSE stream, with each hop (gateway → mcp-server) preserving headers and body
