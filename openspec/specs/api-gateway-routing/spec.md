# Specification: API Gateway Routing

## Purpose
Defines the API gateway's contract for routing inbound HTTP requests to backend services: how backend service URLs are resolved and registered, how route prefixes map to logical service names, how reverse proxying is executed with error handling, and how request ID and distributed trace context are propagated to backends.

## Requirements

### Requirement: Service Registration
The API gateway SHALL maintain an in-memory registry of backend services, mapping a logical service name to a base URL. Registration MUST be thread-safe and MUST overwrite an existing entry for the same name.

#### Scenario: Register a new service
- **WHEN** a service is registered with name "auth-service" and URL "http://auth-service:<port>" (where <port> comes from environment or infrastructure defaults)
- **THEN** subsequent lookups for "auth-service" return the registered URL

#### Scenario: Re-register an existing service
- **WHEN** a service name is registered again with a different URL
- **THEN** the registry stores the new URL and the previous URL is no longer returned

### Requirement: Backend Service URL Resolution
At startup, the API gateway SHALL resolve the base URL of each backend service (auth-service, user-service, objects-service) and register it under a logical service name for proxy routing. The resolved URL for each service MUST be the proxy target used for that service's logical name.

#### Scenario: Environment-provided URL overrides the default
- **WHEN** `AUTH_SERVICE_URL` is set in the environment
- **THEN** the gateway registers auth-service at the URL provided by `AUTH_SERVICE_URL`

#### Scenario: Built-in default applies when no environment variable is set
- **WHEN** `USER_SERVICE_URL` is not set in the environment
- **THEN** the gateway registers user-service at a built-in Docker service-discovery default URL

#### Scenario: Development localhost remapping
- **WHEN** the application environment is "development" AND `DOCKER_ENV` is not "true"
- **THEN** the host component of the auth-service and user-service URLs is replaced with "localhost"

#### Scenario: objects-service URL is not environment-configurable
- **WHEN** the gateway starts
- **THEN** objects-service is registered at a built-in default URL with no environment override and no development remapping applied

> **Note (known asymmetry/bug):** objects-service is the only backend service whose URL is hardcoded in `main.go`, with no environment-variable override and no development localhost remapping, unlike auth-service and user-service. This is documented here as current behavior and is the subject of the first OpenSpec-driven change after the baseline is established.

### Requirement: Route-to-Service Mapping
The API gateway SHALL map inbound route prefixes to a backend service in the registry: `/api/v1/auth/*` and `/api/v1/users/*` MUST route to their respective services, and the `/api/v1/object-types/*`, `/api/v1/objects/*`, `/api/v1/relationship-types/*`, `/api/v1/relationships/*` groups MUST route to objects-service.

#### Scenario: Auth routes target auth-service
- **WHEN** a request hits `/api/v1/auth/login`
- **THEN** the gateway proxies the request to the auth-service backend

#### Scenario: Users routes target user-service
- **WHEN** a request hits `/api/v1/users/42`
- **THEN** the gateway proxies the request to the user-service backend

#### Scenario: Objects-related routes target objects-service
- **WHEN** a request hits `/api/v1/objects`, `/api/v1/object-types`, `/api/v1/relationship-types`, or `/api/v1/relationships`
- **THEN** the gateway proxies the request to the objects-service backend

### Requirement: Reverse Proxy Execution
For each routed request, the gateway SHALL execute a reverse proxy to the resolved backend service URL. The proxy MUST preserve the HTTP method, path, query string, and request body, and MUST set the request Host, scheme, and host to the target service.

#### Scenario: Successful proxy
- **WHEN** a request is routed to a registered service
- **THEN** the gateway forwards it to the backend and returns the backend's response to the client

#### Scenario: Unknown service requested
- **WHEN** a routed request references a service name not in the registry
- **THEN** the gateway responds with HTTP 503 "Service unavailable"

#### Scenario: Invalid service URL
- **WHEN** the registered service URL cannot be parsed
- **THEN** the gateway responds with HTTP 500 "Internal server error"

#### Scenario: Backend unreachable
- **WHEN** the reverse proxy fails to reach the backend service
- **THEN** the gateway responds with HTTP 502 "Service unavailable"

### Requirement: Request ID Propagation
The gateway SHALL set an `X-Request-ID` header on proxied requests. If the inbound request carries an `X-Request-ID`, the gateway MUST reuse it; otherwise it MUST generate a new UUID.

#### Scenario: Inbound request ID is forwarded
- **WHEN** an inbound request carries header `X-Request-ID: abc-123`
- **THEN** the proxied request to the backend carries `X-Request-ID: abc-123`

#### Scenario: Missing request ID is generated
- **WHEN** an inbound request has no `X-Request-ID` header
- **THEN** the gateway generates a UUID and sets it as `X-Request-ID` on the proxied request

### Requirement: Distributed Trace Context Propagation
When tracing is enabled, the gateway SHALL inject the current OpenTelemetry trace context into the headers of proxied requests using the W3C TraceContext text-map propagator.

#### Scenario: Trace context forwarded to backend
- **WHEN** tracing is enabled and a request is proxied
- **THEN** the proxied request headers carry the traceparent/tracestate propagated from the inbound request context
