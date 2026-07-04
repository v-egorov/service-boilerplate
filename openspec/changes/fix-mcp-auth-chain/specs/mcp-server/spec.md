## ADDED Requirements

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
