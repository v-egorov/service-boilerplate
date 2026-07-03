# Specification: API Gateway Services Tests

## ADDED Requirements

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
