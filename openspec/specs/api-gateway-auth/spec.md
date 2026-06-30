# Specification: API Gateway Authentication & Authorization

## Purpose
Defines the API gateway's contract for authenticating inbound requests and authorizing access to protected routes: how the JWT public key is resolved and cached, how Bearer tokens are validated, how revoked tokens are rejected, how user identity flows through the request context and into proxied backend requests, and how route-level role-based access control is enforced.

## Requirements

### Requirement: JWT Public Key Resolution
The gateway SHALL resolve the RSA public key used for JWT validation through a prioritized chain: (1) the `jwt.public_key` field in `config.yaml`, (2) a PEM-encoded key fetched from the auth-service `/public-key` endpoint with retries, (3) the `JWT_PUBLIC_KEY` environment variable. The resolved key SHALL be cached with a 1-hour TTL and periodically refreshed every 30 minutes. If a refresh fails, the expired cached key MUST be retained as a fallback. If no key is available through any source, JWT validation SHALL be disabled and the gateway MUST trust `X-User-*` headers from upstream.

#### Scenario: Key fetched from auth-service
- **WHEN** `jwt.public_key` is empty in config AND auth-service is healthy
- **THEN** the gateway fetches an RSA PEM key from `GET /public-key` and caches it

#### Scenario: Retry with exponential backoff
- **WHEN** the auth-service `/public-key` request fails
- **THEN** the gateway retries up to 10 times with exponential backoff starting at 1 second and capped at 30 seconds

#### Scenario: JWT_PUBLIC_KEY environment variable fallback
- **WHEN** config key is empty AND auth-service fetch fails AND `JWT_PUBLIC_KEY` is set
- **THEN** the gateway parses the PEM-encoded RSA key from `JWT_PUBLIC_KEY`

#### Scenario: Cached key persists across refresh failures
- **WHEN** a periodic cache refresh fails
- **THEN** the previously cached key (even if expired) continues to be used

#### Scenario: No key available — gateway trust mode
- **WHEN** no key is resolvable through any source
- **THEN** JWT validation is skipped and the gateway trusts identity from `X-User-*` request headers

### Requirement: JWT Token Validation
The gateway SHALL validate Bearer tokens from the `Authorization` header. Supported signing algorithms SHALL include HMAC (HS256, HS384, HS512) and RSA (RS256, RS384, RS512). On success, the user identity (`user_id`, `email`, `roles`, `token_type`) SHALL be extracted from JWT claims and stored in the request context. When `Authorization` is absent, the middleware MUST fall back to `X-User-*` gateway headers if present.

#### Scenario: Valid Bearer token
- **WHEN** a request carries `Authorization: Bearer <valid-jwt>`
- **THEN** the token is parsed, signature is validated, and `user_id`, `email`, `roles`, `token_type` are set in the request context

#### Scenario: Invalid or expired token
- **WHEN** the JWT fails parsing, signature validation, or is expired
- **THEN** the gateway responds with HTTP 401 and body `{"error": "Invalid token"}`

#### Scenario: Malformed Authorization header
- **WHEN** the `Authorization` header does not start with `Bearer `
- **THEN** the gateway responds with HTTP 401 and body `{"error": "Invalid authorization format"}`

#### Scenario: Claims type mismatch
- **WHEN** JWT claims cannot be cast to the expected JWTClaims struct
- **THEN** the gateway responds with HTTP 401 and body `{"error": "Invalid token claims"}`

#### Scenario: Fallback to gateway identity headers
- **WHEN** `Authorization` header is absent AND `X-User-ID` header is present
- **THEN** user identity is read from `X-User-ID`, `X-User-Email`, and `X-User-Roles` headers and stored in the request context

#### Scenario: No credentials at all
- **WHEN** neither `Authorization` nor `X-User-*` headers are present
- **THEN** the request proceeds without an authenticated user identity in context

### Requirement: Token Revocation Check
When a token revocation checker is wired up, the gateway SHALL query the auth-service `POST /api/v1/auth/validate-token` endpoint after JWT validation to determine if the token has been revoked. If the revocation service is unreachable, the token MUST be treated as revoked (fail-closed).

#### Scenario: Token passes revocation check
- **WHEN** the revocation checker is active AND auth-service `/validate-token` returns HTTP 200
- **THEN** the request proceeds normally

#### Scenario: Token has been revoked
- **WHEN** the revocation checker is active AND auth-service `/validate-token` returns a non-200 response
- **THEN** the gateway responds with HTTP 401 and body `{"error": "Token has been revoked"}`

#### Scenario: Revocation service unreachable
- **WHEN** the revocation checker is active AND the request to auth-service `/validate-token` fails (network error or timeout)
- **THEN** the gateway responds with HTTP 401 and body `{"error": "Token has been revoked"}`

> **Note (known gap):** The HTTP-based revocation checker is currently only wired up when the JWT public key is successfully fetched from auth-service at startup. When the key is loaded from `config.yaml` or the `JWT_PUBLIC_KEY` environment variable, the `revocationChecker` remains nil and no revocation check is performed. This is a gap because token revocation is independent of the key source — a service should be able to check revocation regardless of how it obtained the public key. Expected to be addressed in a future change.

### Requirement: User Identity in Request Context
After successful authentication, the gateway SHALL store the following values in the Gin request context: `user_id` (string UUID), `user_email` (string), `user_roles` ([]string), and `token_type` (string). Helper functions `GetAuthenticatedUserID`, `GetAuthenticatedUserEmail`, and `GetAuthenticatedUserRoles` SHALL be available for downstream middleware and handlers to read these values.

#### Scenario: Identity accessible to downstream handlers
- **WHEN** a request is authenticated via JWT
- **THEN** `c.Get("user_id")` returns the user's UUID, `c.Get("user_email")` returns the email, and `c.Get("user_roles")` returns the roles slice

#### Scenario: Unauthenticated request has empty identity
- **WHEN** a request proceeds without authentication
- **THEN** `GetAuthenticatedUserID(c)` returns an empty string and `GetAuthenticatedUserRoles(c)` returns an empty slice

### Requirement: Identity Header Forwarding to Backends
When proxying a request to a backend service, the gateway SHALL forward the authenticated user's identity via HTTP headers: `X-User-ID` set to the user's UUID, `X-User-Email` set to the email, and `X-User-Roles` set to a comma-wrapped list in the format `,role1,role2,`. These headers SHALL only be set when the user has been authenticated.

#### Scenario: Authenticated identity forwarded
- **WHEN** a proxied request has `user_id`, `user_email`, and `user_roles` in context
- **THEN** the backend receives `X-User-ID`, `X-User-Email`, and `X-User-Roles` headers with the corresponding values

#### Scenario: Unauthenticated request omits identity headers
- **WHEN** a proxied request has no `user_id` in context
- **THEN** no `X-User-*` headers are sent to the backend

#### Scenario: Roles serialized in comma-wrapped format
- **WHEN** the authenticated user has roles `["admin", "user"]`
- **THEN** the `X-User-Roles` header value is `,admin,user,`

### Requirement: Route-Level Access Control
The gateway SHALL enforce authentication and role requirements on route groups using `RequireAuth` and `RequireRole` middleware. `RequireAuth` MUST reject requests without an authenticated `user_id` in context with HTTP 401. `RequireRole` MUST reject requests where the user does not possess at least one of the required roles with HTTP 403.

#### Scenario: Protected route with authentication
- **WHEN** a request hits a route protected by `RequireAuth` with a valid `user_id` in context
- **THEN** the request proceeds to the handler

#### Scenario: Protected route without authentication
- **WHEN** a request hits a route protected by `RequireAuth` without `user_id` in context
- **THEN** the gateway responds with HTTP 401 and body `{"error": "Authentication required"}`

#### Scenario: Admin-only route with admin role
- **WHEN** a request hits a route protected by `RequireRole("admin")` and the user possesses the "admin" role
- **THEN** the request proceeds to the handler

#### Scenario: Admin-only route without admin role
- **WHEN** a request hits a route protected by `RequireRole("admin")` and the user lacks the "admin" role
- **THEN** the gateway responds with HTTP 403 and body `{"error": "Insufficient permissions"}`

#### Scenario: Role check is OR, not AND
- **WHEN** a route is protected by `RequireRole("admin", "moderator")` and the user has role "moderator" but not "admin"
- **THEN** the request proceeds — any single matching role is sufficient
