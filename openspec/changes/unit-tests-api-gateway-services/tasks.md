## 1. ServiceRegistry tests (services/registry_test.go)

- [x] 1.1 Create registry_test.go with test helper to build a logger for tests and register services
- [x] 1.2 Test RegisterService stores service mapping correctly
- [x] 1.3 Test GetServiceURL returns registered URL on success path
- [x] 1.4 Test GetServiceURL returns "not found" error for unknown service names
- [x] 1.5 Test ListServices returns all registered services in a map
- [x] 1.6 Test UnregisterService removes service and subsequent GetServiceURL fails

## 2. AuthMiddleware tests (middleware/auth_test.go)

- [x] 2.1 Create auth_test.go with helper to build gin.Context for middleware testing
- [x] 2.2 Test AuthMiddleware allows valid Bearer token requests through
- [x] 2.3 Test AuthMiddleware returns 401 when Authorization header is missing
- [x] 2.4 Test AuthMiddleware returns 401 for non-Bearer authorization format

## 3. CORSMiddleware tests (middleware/auth_test.go)

- [x] 3.1 Add CORSMiddleware test to verify OPTIONS requests return HTTP 204 NoContent
- [x] 3.2 Verify standard method requests (GET/POST/PUT/DELETE) proceed through middleware with CORS headers set

## 4. RequestIDMiddleware tests (middleware/auth_test.go)

- [x] 4.1 Test RequestIDMiddleware preserves existing X-Request-ID header from inbound request
- [x] 4.2 Test RequestIDMiddleware generates new UUID when no X-Request-ID is present in request