# Specification: API Gateway Services Tests

## Purpose

Defines the test coverage requirements for the api-gateway service, specifying what behaviors must be verified through unit tests across its core packages: ServiceRegistry (services), middleware functions (auth, CORS, request-id), and handler endpoints (liveness, ping, readiness).

## Requirements

### Requirement: ServiceRegistry tests cover all public methods
The test suite MUST verify correct behavior for all four ServiceRegistry public methods: RegisterService, GetServiceURL, ListServices, and UnregisterService. Each method must have positive scenarios (expected success paths) and negative scenarios (error conditions where applicable).

#### Scenario: RegisterService stores service mapping
- **WHEN** RegisterService is called with a valid name and URL
- **THEN** the service appears in subsequent GetServiceURL and ListServices calls

#### Scenario: GetServiceURL returns registered URL
- **WHEN** a service has been registered via RegisterService
- **THEN** GetServiceURL returns the exact URL that was provided during registration

#### Scenario: GetServiceURL returns error for unknown service
- **WHEN** GetServiceURL is called with a name that was never registered
- **THEN** it returns an error containing the service name and "not found"

#### Scenario: ListServices returns all registered services
- **WHEN** multiple services have been registered via RegisterService
- **THEN** ListServices returns a map containing all registered service-name-to-URL mappings

#### Scenario: UnregisterService removes service from registry
- **WHEN** UnregisterService is called with an existing service name
- **THEN** subsequent GetServiceURL calls for that name return "not found" error and ListServices no longer includes it

### Requirement: AuthMiddleware tests cover authorization scenarios
The test suite MUST verify correct behavior for the AuthMiddleware gin.HandlerFunc, testing valid Bearer tokens, missing Authorization headers, non-Bearer formats, and proper request continuation.

#### Scenario: Valid Bearer token allows request to proceed
- **WHEN** a request includes "Authorization: Bearer any-token" header
- **THEN** the middleware calls c.Next() and does not write an error response

#### Scenario: Missing Authorization header returns 401
- **WHEN** a request has no Authorization header
- **THEN** the middleware responds with HTTP 401 and aborts the request (c.Abort is called)

#### Scenario: Non-Bearer authorization format returns 401
- **WHEN** a request includes "Authorization: Basic xyz" header
- **THEN** the middleware responds with HTTP 401 and aborts the request

### Requirement: CORSMiddleware tests cover CORS handling
The test suite MUST verify correct behavior for the CORSMiddleware gin.HandlerFunc, testing OPTIONS preflight requests, standard method routing, and proper CORS headers.

#### Scenario: OPTIONS request returns NoContent status
- **WHEN** a request uses HTTP OPTIONS method
- **THEN** the middleware responds with HTTP 204 (No Content) without calling Next()

#### Scenario: Standard methods proceed through middleware
- **WHEN** a request uses GET, POST, PUT, or DELETE method
- **THEN** the middleware calls c.Next() and sets Access-Control-Allow-Origin header

#### Scenario: CORS headers are set on response
- **WHEN** any non-OPTIONS request passes through CORSMiddleware
- **THEN** response includes Access-Control-Allow-Methods and Access-Control-Allow-Headers headers

### Requirement: GatewayHandler liveness and ping tests
The test suite MUST verify that LivenessHandler returns a JSON response with status "ok" and all metadata fields (timestamp, gateway name, version), and PingHandler returns a JSON response with status "pong" and matching metadata.

#### Scenario: LivenessHandler returns ok status with metadata
- **WHEN** a GET request hits the liveness handler endpoint
- **THEN** the response has HTTP 200 with JSON body containing `"status": "ok"` and `gateway`/`version` fields from config

#### Scenario: PingHandler returns pong status with metadata
- **WHEN** a GET request hits the ping handler endpoint
- **THEN** the response has HTTP 200 with JSON body containing `"status": "pong"` and `gateway`/`version` fields from config

### Requirement: GatewayHandler readiness logic tests
The test suite MUST verify that ReadinessHandler returns HTTP 200 when at least one service is registered in the backend registry, and HTTP 503 with an error message when no services are registered.

#### Scenario: ReadinessHandler returns ok when services available
- **WHEN** the service registry has one or more registered services
- **THEN** ReadinessHandler responds with HTTP 200 and JSON body containing `"status": "ok"`

#### Scenario: ReadinessHandler returns error when no services registered
- **WHEN** the service registry is empty (no services registered)
- **THEN** ReadinessHandler responds with HTTP 503 and JSON body containing `"status": "error"` and a message about no services available

### Requirement: Logging middleware responseWriter tests
The test suite MUST verify that the responseWriter struct correctly captures HTTP status codes from WriteHeader calls and accumulates byte counts from Write calls.

#### Scenario: responseWriter captures correct status code
- **WHEN** WriteHeader is called with a specific HTTP status code (e.g., 404)
- **THEN** the internal `status` field reflects that exact code

#### Scenario: responseWriter accumulates written bytes correctly
- **WHEN** Write is called multiple times with different byte slices
- **THEN** the internal `size` field equals the sum of all write lengths