# Specification: mcp-server

## Purpose

Define the MCP server service that provides AI agent access to objects-service through a standardized MCP protocol interface. The server exposes read-only tools, resources, and prompt templates for exploring object types, objects, and their relationships — all routed through the API Gateway's SSE transport without JWT validation (delta 1).
## Requirements
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

### Requirement: MCP agent identity has valid auth-service permissions
The api-gateway MUST inject a system-level user UUID (not string "mcp-agent") into `X-User-ID` header for all `/mcp/*` requests. A corresponding user must exist in auth-service with the `mcp-agent-read-only` role, which grants read permissions (`objects:read:all`, `objects:read:own`, and `object-types:read:all`) to ensure objects-service permiddleware permission checks succeed without requiring per-user JWT validation.

#### Scenario: Gateway injects UUID-based identity for MCP requests
- **WHEN** an MCP client sends a request through api-gateway's `/mcp/*` endpoint
- **THEN** the gateway sets `X-User-ID` to a valid UUID, `X-User-Roles` includes `mcp-agent-read-only`, and objects-service permiddleware successfully calls auth-service CheckPermission without panic

#### Scenario: Auth-service permission check succeeds for MCP agent on objects
- **WHEN** objects-service permiddleware calls auth-service `/api/v1/auth/permissions/check` with the MCP agent's UUID requesting `"objects:read:all"` or `"objects:read:own"`
- **THEN** auth-service resolves permissions via `user_roles → role_permissions → permissions`, finds `objects:read:all` and `objects:read:own`, and returns `allowed=true`

#### Scenario: Auth-service permission check succeeds for MCP agent on object-types
- **WHEN** objects-service permiddleware calls auth-service `/api/v1/auth/permissions/check` with the MCP agent's UUID requesting `"object-types:read:all"` or `"object-types:read:own"`
- **THEN** auth-service resolves permissions via `user_roles → role_permissions → permissions`, finds `object-types:read:all`, and returns `allowed=true`

### Requirement: MCP server forwards caller identity to objects-service
For every tool, resource, and prompt handler that calls objects-service, the mcp-server MUST attach the gateway-injected identity headers (`X-User-ID`, `X-User-Email`, `X-User-Roles`) from the handler's inbound request onto the outbound HTTP request to objects-service. Only headers in the explicit identity forward list SHALL be copied; no other inbound headers SHALL be forwarded wholesale.

#### Scenario: Tool call carries identity through to objects-service
- **WHEN** a client invokes a tool (e.g. `list_objects`) through the gateway and the inbound request carries `X-User-ID`, `X-User-Email`, and `X-User-Roles`
- **THEN** the outbound request from mcp-server to objects-service contains the same three headers with their original values, and objects-service permiddleware sees a non-empty user context and authorizes the read

#### Scenario: Resource read carries identity through to objects-service
- **WHEN** an agent reads the `objects-types` resource and the inbound request carries the identity headers
- **THEN** the outbound request to objects-service contains the same three identity headers

#### Scenario: Prompt invocation carries identity through to objects-service
- **WHEN** an agent invokes a prompt (e.g. `browse_schema`, `get_object_info`) and the inbound request carries the identity headers
- **THEN** the outbound request to objects-service contains the same three identity headers

#### Scenario: Non-identity inbound headers are not forwarded
- **WHEN** the inbound request contains headers outside the identity forward list (e.g. `User-Agent`, `Accept`, arbitrary custom headers)
- **THEN** none of those headers appear on the outbound request to objects-service

#### Scenario: Absent identity context proceeds without identity headers
- **WHEN** an outbound call to objects-service is made from a code path with no inbound identity in its context (e.g. health/readiness checks, background callers)
- **THEN** the request proceeds with no `X-User-*` headers rather than being rejected

### Requirement: Objects-service trusts gateway identity for MCP reads without JWT
For requests that carry a non-empty `X-User-ID` header but no `Authorization` header, objects-service's permission middleware SHALL skip the auth-service permission check for GET requests and trust the gateway-injected identity directly. This path SHALL only activate when both conditions are met: `X-User-ID` is present AND `Authorization` is absent. When `Authorization` is present, the normal RBAC path through auth-service's CheckPermission SHALL be used regardless of `X-User-ID`.

#### Scenario: MCP tool call with gateway identity succeeds without JWT
- **WHEN** a GET request arrives at objects-service with `X-User-ID: 00000000-0000-4000-8000-000000000001` and no `Authorization` header
- **THEN** the permiddleware skips the `auth-service.CheckPermission()` call, sets `matched_permissions` to the required permissions, and allows the request to proceed

#### Scenario: Request with both identity and JWT uses normal RBAC path
- **WHEN** a GET request arrives at objects-service with both `X-User-ID` and a valid `Authorization: Bearer <jwt>` header
- **THEN** the permiddleware calls `auth-service.CheckPermission()` with the JWT token and follows the normal RBAC authorization flow

#### Scenario: Request with no identity and no JWT is still rejected
- **WHEN** a request arrives at objects-service with no `X-User-ID` and no `Authorization` header
- **THEN** the permiddleware returns 401 "Authentication required" — the gateway-trust shortcut does NOT bypass the identity check

#### Scenario: Non-GET request with gateway identity without JWT uses normal RBAC path
- **WHEN** a POST, PUT, or DELETE request arrives at objects-service with `X-User-ID` but no `Authorization` header
- **THEN** the permiddleware calls `auth-service.CheckPermission()` with an empty JWT token — the gateway-trust shortcut only applies to GET requests

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

