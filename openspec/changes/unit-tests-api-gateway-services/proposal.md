## Why

The api-gateway service has zero test coverage across all 5 source files. This is a serious gap that makes any future refactoring or bug fixes risky since there's no safety net to catch regressions. The ServiceRegistry (services/registry.go) and auth/CORS/request-id middleware (middleware/auth.go) are the highest-value targets: pure logic with no external dependencies, making them quick wins for establishing test infrastructure patterns.

## What Changes

- Add unit tests for `api-gateway/internal/services` — cover ServiceRegistry RegisterService, GetServiceURL, ListServices, UnregisterService
- Add unit tests for `api-gateway/internal/middleware/auth.go` — cover AuthMiddleware (valid/invalid/missing headers), CORSMiddleware (OPTIONS, standard methods), RequestIDMiddleware (existing header, generated UUID)
- Establish test patterns consistent with existing service test conventions (testify assertions, gin httptest for middleware, interface-based mocks where needed)

## Capabilities

### New Capabilities
- `api-gateway-services-tests`: Unit tests for ServiceRegistry and auth/CORS/request-id middleware packages

**No spec-level behavior changes** — this is purely a testing concern. No delta specs required.

## Impact

- **New files:** `services/registry_test.go`, `middleware/auth_test.go`
- **Existing files modified:** none (tests are additive, no production code changes)
- **Dependencies:** testify for assertions, gin httptest for middleware — both already in go.mod as dev dependencies via other services