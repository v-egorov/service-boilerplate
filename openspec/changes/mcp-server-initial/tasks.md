## 1. Service Scaffolding

- [ ] 1.1 Create directory structure: `services/mcp-server/cmd/main.go`, `internal/client/`, `internal/tools/`, `internal/resources/`, `internal/prompts/`
- [ ] 1.2 Create config.yaml with objects-service URL, port (8095), and logging config following existing service patterns
- [ ] 1.3 Initialize go.mod wrapper referencing mcp-go SDK dependency (github.com/mark3labs/mcp-go v0.x)
- [ ] 1.4 Wire up main.go: load config → init logger → start MCP server on configured port

## 2. HTTP Client for objects-service

- [ ] 2.1 Create `internal/client/objects_client.go` with HTTP client wrapping objects-service REST API calls
- [ ] 2.2 Implement ObjectTypeClient methods: List(), GetByID(), GetByName() — mapping to existing objects-service endpoints
- [ ] 2.3 Implement ObjectClient methods: GetByPublicID(), List(type_key, page, page_size) — mapping to existing objects-service endpoints
- [ ] 2.4 Write unit tests for HTTP client mocking (object_client_test.go)

## 3. MCP Server Core Setup

- [ ] 3.1 Initialize mcp-go server with SSE transport in main.go
- [ ] 3.2 Register the `list_object_types` tool with mcp-go including description, input schema, and handler function
- [ ] 3.3 Register the `get_object_type` tool (parameters: id as integer OR type_key as string) with handler
- [ ] 3.4 Register the `list_objects` tool (parameters: object_type_id required, page/page_size optional) with handler
- [ ] 3.5 Register the `get_object` tool (parameter: public_id as UUID string) with handler
- [ ] 3.6 Implement each tool's handler to call objects-client methods and format MCP-compliant results

## 4. Resource Registration

- [ ] 4.1 Create `internal/resources/type_hierarchy.go` — register `objects-types` resource URI template
- [ ] 4.2 Implement resource read handler that fetches full type hierarchy from ObjectTypeClient.GetTree()
- [ ] 4.3 Return resource content as JSON representing the complete object type tree

## 5. Prompt Templates

- [ ] 5.1 Create `internal/prompts/browse_schema.go` — register prompt template named `browse_schema` with optional type_key_prefix parameter
- [ ] 5.2 Implement browse_schema handler: fetch types, format message guiding agent toward exploration
- [ ] 5.3 Create `internal/prompts/get_object_info.go` — register prompt template named `get_object_info` with required object_type_id parameter
- [ ] 5.4 Implement get_object_info handler: summarize objects of given type, suggest next tool calls

## 6. Docker Compose Integration

- [ ] 6.1 Add mcp-server entry to docker-compose.yml (service name: mcp-server, image: service-boilerplate-mcp-server, ports: 8095)
- [ ] 6.2 Configure mcp-server depends_on objects-service and network placement (same compose network as other services)
- [ ] 6.3 Mount config.yaml into container following existing volume mount pattern

## 7. API Gateway Routing & Header Injection

- [ ] 7.1 Add `/mcp/*` route group in `api-gateway/cmd/main.go` (following existing pattern for other service routes)
- [ ] 7.2 Register SSE endpoint: `GET /mcp/sse → ProxyRequest("mcp-server")`
- [ ] 7.3 Ensure gateway injects `X-User-ID: mcp-agent` header for `/mcp/*` requests (modify ProxyRequest or add route-specific middleware)
- [ ] 7.4 Verify SSE streaming works through reverse proxy — chunked transfer encoding passthrough, no premature connection close
- [ ] 7.5 Ensure gateway doesn't apply JWT validation to `/mcp/*` routes (unprotected for delta 1)

## 8. Validation & Testing

- [ ] 8.1 Run `make build-mcp-server` and verify clean compilation
- [ ] 8.2 Start services via `make dev-detached`, verify mcp-server container starts successfully
- [ ] 8.3 Test SSE connection: curl GET /mcp/sse through gateway, verify HTTP 200 with text/event-stream content type
- [ ] 8.4 Test each tool call (list_object_types, get_object_type, list_objects, get_object) through MCP protocol
- [ ] 8.5 Test resource read via objects-types URI
- [ ] 8.6 Test both prompt templates invoke correctly
