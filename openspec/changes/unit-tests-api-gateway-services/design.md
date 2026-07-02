# Design: Unit Tests for API Gateway Services

## Context

The api-gateway service has zero test coverage across all 5 source files in the `api-gateway/internal/` directory. This is a critical gap that makes any future refactoring or bug fixes risky since there's no safety net to catch regressions. The ServiceRegistry (`services/registry.go`) and auth/CORS/request-id middleware (`middleware/auth.go`) are pure logic with no external dependencies, making them ideal first targets for establishing test patterns in this service.

## Goals / Non-Goals

**Goals:**
- Add comprehensive unit tests for `api-gateway/internal/services` (ServiceRegistry) — 4 public methods to cover
- Add unit tests for `api-gateway/internal/middleware/auth.go` (AuthMiddleware, CORSMiddleware, RequestIDMiddleware) — 3 middleware functions
- Establish test patterns consistent with existing service conventions (testify assertions, gin httptest for HTTP handlers)
- No production code changes required — only adding test files

**Non-Goals:**
- Testing `handlers/gateway.go` — requires complex mock backend server setup, defer to separate change
- Testing `middleware/logging.go` — depends on metrics collector with deep gin internals, defer to later
- Integration tests — out of scope for this initial coverage effort

## Decisions

### 1. Use testify/assert for assertions (not require)
**Rationale:** Existing service test suites in auth-service and objects-service use `assert` throughout. Consistency matters more than strictness for unit tests where individual test isolation is guaranteed by Go's testing framework.

**Alternatives considered:** `require` would stop on first assertion failure, but since each test function runs independently, there's no benefit over `assert`.

### 2. Use httptest.NewRecorder() for middleware tests
**Rationale:** gin provides `httptest.NewRequest()` and `httptest.NewRecorder()` which work with gin's Context wrapper. This is the standard pattern from go testing docs and is used consistently across service test suites in this project.

**Alternatives considered:** A full httptest.Server would be overkill for middleware that just sets headers and calls Next().

### 3. No mocks needed for ServiceRegistry
**Rationale:** The registry has no interfaces or external dependencies — it's a plain struct with map operations. Direct instantiation and method calls are sufficient; mocking adds complexity without value.

**Alternatives considered:** Interface extraction would enable mocking, but introduces unnecessary abstraction for a simple in-memory data structure that is inherently testable as-is.

### 4. Use gin.Default() router for middleware tests
**Rationale:** Middleware registration requires a route group or engine. Using `gin.Default()` provides the minimal setup needed to attach and invoke middleware without boilerplate.

**Alternatives considered:** Creating a bare `&gin.Engine{}` would work but lacks default recovery middleware that some edge cases might depend on.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Test flakiness from goroutines in registry (RWMutex) | Use sync.WaitGroup or sequential test ordering; avoid concurrent tests for same registry instance |
| gin context state leakage between middleware tests | Create fresh `gin.Context` per test using `httptest.NewRequest()` + `NewRecorder()` — no shared state |
| Test coverage gaps from complex error paths | Accept that initial run focuses on happy paths and obvious error conditions; expand in follow-up changes if needed |

## Migration Plan

No migration required. Only new files are added:
- `api-gateway/internal/services/registry_test.go`
- `api-gateway/internal/middleware/auth_test.go`

Existing code is untouched, so no rollback strategy needed — tests can be reverted independently via git revert on just the test files.

## Open Questions

None identified — this scope is well-contained with clear boundaries and established patterns from sibling services.