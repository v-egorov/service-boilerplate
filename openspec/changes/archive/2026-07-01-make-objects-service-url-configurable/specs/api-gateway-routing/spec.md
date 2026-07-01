# Delta Specification: API Gateway Routing — Backend Service URL Resolution

## MODIFIED Requirements

### Requirement: Backend Service URL Resolution
At startup, the API gateway SHALL resolve the base URL of each backend service (auth-service, user-service, objects-service) and register it under a logical service name for proxy routing. The resolved URL for each service MUST be the proxy target used for that service's logical name. Each service URL MAY be overridden via an environment variable (`AUTH_SERVICE_URL`, `USER_SERVICE_URL`, `OBJECTS_SERVICE_URL`) with built-in Docker service-discovery defaults as fallbacks.

#### Scenario: Environment-provided URL overrides the default
- **WHEN** `AUTH_SERVICE_URL` is set in the environment
- **THEN** the gateway registers auth-service at the URL provided by `AUTH_SERVICE_URL`

#### Scenario: OBJECTS_SERVICE_URL provides objects-service override
- **WHEN** `OBJECTS_SERVICE_URL` is set in the environment
- **THEN** the gateway registers objects-service at the URL provided by `OBJECTS_SERVICE_URL`

#### Scenario: Built-in default applies when no environment variable is set
- **WHEN** `USER_SERVICE_URL` is not set in the environment
- **THEN** the gateway registers user-service at a built-in Docker service-discovery default URL

#### Scenario: Development localhost remapping for all services
- **WHEN** the application environment is "development" AND `DOCKER_ENV` is not "true"
- **THEN** the host component of auth-service, user-service, and objects-service URLs is replaced with "localhost"

### Requirement: Service Registration
The API gateway SHALL maintain an in-memory registry of backend services, mapping a logical service name to a base URL. Registration MUST be thread-safe and MUST overwrite an existing entry for the same name.

#### Scenario: Register a new service
- **WHEN** a service is registered with name "auth-service" and URL `http://auth-service:<port>` (where <port> comes from environment or infrastructure defaults)
- **THEN** subsequent lookups for "auth-service" return the registered URL