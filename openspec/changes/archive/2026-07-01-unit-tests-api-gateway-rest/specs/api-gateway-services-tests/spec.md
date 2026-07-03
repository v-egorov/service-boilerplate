# Delta Specification: API Gateway Services Tests — Handlers and Logging

## ADDED Requirements

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