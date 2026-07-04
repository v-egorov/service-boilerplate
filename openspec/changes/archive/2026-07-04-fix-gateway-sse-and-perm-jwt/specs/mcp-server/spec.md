## ADDED Requirements

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
