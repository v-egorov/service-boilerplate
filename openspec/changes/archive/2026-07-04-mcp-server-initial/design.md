## Context

Agents (Hermes Agent, pi agent) need programmatic access to objects-service data — specifically the object type hierarchy and basic object queries. Currently there is no MCP-compatible interface for AI agents to browse or query this data. The existing api-gateway proxies HTTP REST calls but does not speak MCP protocol. Objects-service has a rich internal API (~30+ methods) across ObjectType, Object, RelationshipType, and Relationship domains, but only HTTP endpoints are exposed externally.

The service boilerplate already follows a consistent pattern: new services are standalone Go binaries behind api-gateway with their own docker-compose entry, config.yaml, and directory structure under `services/`. This change follows that exact pattern.

## Goals / Non-Goals

**Goals:**
- Implement mcp-server as a new microservice using the mcp-go SDK
- Expose 4 read-only tools for browsing object types and objects
- Register MCP resource for the full object type hierarchy
- Provide 2 prompt templates as thin wrappers around core tools
- Integrate into docker-compose alongside existing services
- Route through api-gateway (no direct external access)

**Non-Goals:**
- Write operations (create/update/delete objects or types) — future delta
- Authentication/RBAC on MCP endpoints for delta 1 — routes unprotected, service account identity injected downstream via gateway headers
- Direct database access — mcp-server calls objects-service HTTP REST API only
- WebSocket transport — api-gateway confirmed does NOT support WebSocket upgrades (gin router handles standard HTTP only); SSE is sufficient
- Prompt templates beyond the 2 core workflows defined in this delta

## Decisions

### Decision: Standalone service behind api-gateway (not embedded, not gateway adapter)

**Chosen**: mcp-server is a new Go binary under `services/mcp-server/`, registered in docker-compose, routed through api-gateway.

```
LLM Client → SSE → api-gateway (JWT auth) → HTTP REST → mcp-server → HTTP REST → objects-service
```

**Rationale**: Clean separation of concerns; consistent with existing service pattern. Gateway has no business logic layer for MCP tool definitions. Embedded approach couples gin + mcp-go in one process.

### Decision: SSE transport over WebSocket or Stdio

SSE is the only viable transport behind api-gateway (no WebSocket upgrades). MCP clients connect once via `GET /mcp/sse` and send requests on the same channel. The SSE stream carries server→client messages; client→server requests are sent as HTTP POST to the endpoint URL provided in the initial SSE event.

**Rationale**: Fits gateway architecture perfectly. SSE is just chunked HTTP — reverse proxy handles it natively. Stdio requires direct process launch (not viable behind proxy). WebSocket would require gateway changes we don't want yet.

### Decision: HTTP REST API to objects-service (not direct DB)

mcp-server calls existing objects-service HTTP endpoints — no database connection, no repository code duplication. This creates an extra hop but eliminates schema coupling and leverages existing handler/service layers.

**Rationale**: Zero database-level coupling, no need to replicate repository queries or know PostgreSQL schemas. Objects-service already handles validation, error mapping, and response formatting. The tradeoff is one extra HTTP hop per request (acceptable for dev/internal use).

### Decision: Unprotected MCP routes with service account identity injection

MCP routes (`/mcp/*`) are unprotected in api-gateway for delta 1 — no JWT validation required. The gateway injects a `X-User-ID: mcp-agent` header as part of the proxy request, providing downstream objects-service with a valid user context.

**How it works:**
```
LLM Client → GET /mcp/sse (no auth) → api-gateway (skips JWT validation for /mcp/*)
                                            ↓ injects X-User-ID: mcp-agent
                                        mcp-server → HTTP REST → objects-service
                                                            ✓ has user context to proceed
```

**Service account setup**: A migration will create an `mcp-agent` role in auth-service with read permissions on `object-types` and `objects`. Objects-service permiddleware checks this identity — since routes are unprotected, we only grant read:all (no write operations possible anyway).

**Future protected routes**: When MCP endpoints need per-user authentication:
- Option A: Require JWT Bearer token in SSE connection URL (`/mcp/sse?token=...`)
- Option B: Add API key-based auth endpoint separate from standard login flow
- Both require changes to api-gateway to pass user context on unauthenticated MCP requests

**Rationale**: Reduces complexity for initial implementation. Write operations (which need per-user RBAC) are out of scope anyway. Dev-only service, internal use only.

### Decision: mcp-go SDK

Use `github.com/mark3labs/mcp-go` — production-quality Go SDK with SSE transport support, well-maintained, and used by major MCP clients. Version 0.x (latest stable).

**Rationale**: Best-in-class Go MCP implementation. Handles SSE transport, tool registration, resource management, and prompt templates out of the box.

### Decision: Tool Surface Area — Path B (2 categories, 4 tools)

Tools fall into **two conceptual categories** for this delta:

**Type management tools (schema layer):**
- `list_object_types` → ObjectTypeService.List()
- `get_object_type` → ObjectTypeService.GetByID() or GetByName()

**Object query tools (instance layer):**
- `list_objects` → ObjectService.List() with type_key filter only
- `get_object` → ObjectService.GetByPublicID()

Future deltas will add search, metadata filtering, tags, graph queries, and write operations.

### Decision: 2 Prompt Templates as thin wrappers

- `browse_schema` → wraps list_object_types + get_object_type
- `get_object_info` → wraps list_objects + get_object

Prompts are metadata on top of existing tools — argument schemas reused, minimal code (~50 lines total). They improve agent UX by providing named workflows instead of raw API surface.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| HTTP hop adds latency (client → gateway → mcp-server → objects-service) | Acceptable for dev/internal use; measurable in future if needed |
| mcp-server depends on objects-service availability | Same dependency model as all existing services behind gateway |
| SSE long-lived connections increase resource usage | Standard pattern; connection count limited by agent concurrency (typically 1-2) |
| Gateway reverse-proxy may buffer SSE chunked responses | Gin's httputil.ReverseProxy handles Transfer-Encoding: chunked natively; monitor in dev |
