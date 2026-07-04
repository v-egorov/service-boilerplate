## Why

Agents (Hermes Agent, pi agent) need programmatic access to the objects-service data model — specifically the object type hierarchy and basic object queries. Currently there is no MCP-compatible interface for AI agents to browse or query this data, creating a gap between what agents can do with our services and what they should be able to do in development workflows.

## What Changes

- Introduce **mcp-server** as a new microservice that exposes objects-service capabilities via the Model Context Protocol (MCP) using SSE transport
- MCP server calls existing objects-service HTTP REST API — zero database-level coupling
- Expose 4 read-only tools across two categories:
  - *Type management*: `list_object_types`, `get_object_type` (browse schema)
  - *Object queries*: `list_objects`, `get_object` (query instances by type or ID)
- Register 1 MCP resource: `objects-types` (full schema hierarchy as browseable data at URI). Objects themselves are query-based (require parameters), so they belong as tools not resources.
- Provide 2 prompt templates as thin wrappers around core tools for better agent UX
- Add mcp-server to docker-compose alongside existing services, behind api-gateway

## Capabilities

### New Capabilities
- `mcp-server`: Model Context Protocol server exposing objects-service read-only capabilities (tools, resources, prompts) via SSE transport

### Modified Capabilities
<!-- None for this delta — no existing spec requirements change -->

## Impact

- **New**: `services/mcp-server/` directory with full service structure (cmd/, internal/client, internal/tools, internal/resources, internal/prompts, config.yaml)
- **New**: `services/auth-service/migrations/development/NNNNXX_mcp_agent_role.up.sql` — create mcp-agent role with read permissions
- **Modified**: `docker/docker-compose.yml` — add mcp-server service + network entry
- **Modified**: `api-gateway/cmd/main.go` — add `/mcp/*` route group (SSE endpoint, unprotected), inject `X-User-ID: mcp-agent` header for downstream objects-service user context
- **Migration**: auth-service migration to create `mcp-agent` service account role with read permissions on object-types and objects
- **Dependency**: `github.com/mark3labs/mcp-go` SDK (Go module for MCP server implementation)
- **Runtime**: mcp-server connects to objects-service via HTTP REST API (no direct DB access)
