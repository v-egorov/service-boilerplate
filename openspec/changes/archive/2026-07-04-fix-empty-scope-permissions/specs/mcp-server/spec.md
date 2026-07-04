## MODIFIED Requirements

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
