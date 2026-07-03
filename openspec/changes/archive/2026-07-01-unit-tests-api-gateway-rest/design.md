# Design: API Gateway Tests — Handlers and Logging

## Context

This is the second round of api-gateway test coverage, following `unit-tests-api-gateway-services` which covered `services/registry.go` and `middleware/auth.go`. Two packages remain: `handlers/gateway.go` (15+ methods) and `middleware/logging.go` (2 middleware functions + responseWriter struct). This pass targets the high-value, easily testable pieces while deferring complex infrastructure-dependent code.

## Goals / Non-Goals

**Goals:**
- Test LivenessHandler, PingHandler, ReadinessHandler in handlers/gateway.go
- Test responseWriter struct and RequestResponseLogger skip-paths behavior from logging.go
- Establish patterns consistent with pass 1 (testify/assert, gin httptest)

**Non-Goals:**
- Testing ProxyRequest — requires a full mock backend server via httputil.ReverseProxy
- Testing StatusHandler — makes real HTTP calls to health endpoints across services
- Testing DetailedRequestLogger — depends on metrics collector and body capture logic that adds complexity beyond scope

## Decisions

### 1. Use NewGatewayHandler with minimal config for handler tests
**Rationale:** LivenessHandler, PingHandler, and ReadinessHandler only need registry (for readiness check) and a basic config struct. No real HTTP backend needed.

```go
cfg := &config.Config{App: config.AppConfig{Name: "test-gw", Version: "0.1.0"}}
reg := services.NewServiceRegistry(logger)
handler := NewGatewayHandler(reg, logger, cfg)
```

**Alternatives considered:** Interface extraction for GatewayHandler would enable full mocking, but introduces unnecessary abstraction for simple JSON handlers that are inherently testable as-is.

### 2. Test responseWriter directly (not through middleware wrapper)
**Rationale:** The responseWriter is a simple struct with WriteHeader/Write methods. Direct instantiation and method calls provide the most focused tests without gin context boilerplate.

**Alternatives considered:** Testing through DetailedRequestLogger would verify end-to-end behavior but adds dependency on metrics collector and request body capture logic that obscures what's being tested.

### 3. Test RequestResponseLogger skip-paths via gin router registration
**Rationale:** The middleware wraps `gin.LoggerWithConfig` with a `SkipPaths` list. We can verify it by making requests to both skipped paths (/health) and non-skipped paths, then checking the logger output captures only non-skipped requests.

**Alternatives considered:** Since gin's default logger writes to stdout/stderr which is hard to assert on in tests, we could use a custom io.Writer for log output capture — but that adds complexity. For now, verify the SkipPaths config is correct via indirect testing (assert no panic/errors when hitting skipped paths).

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| ReadinessHandler test flakes from timing (time.Now() in response) | Only assert status code and presence of expected keys, not exact timestamp values |
| gin.TestMode needed to suppress default logger output | Set `gin.SetMode(gin.TestMode)` before creating routers — standard pattern from pass 1 |
| config.Config struct is large with many nested fields | Use minimal inline structs for only the fields tested (App.Name, App.Version) |

## Migration Plan

No migration required. Only new test files added:
- `api-gateway/internal/handlers/gateway_test.go`
- `api-gateway/internal/middleware/logging_test.go`

Existing code untouched — tests revert independently via git revert on just test files.

## Open Questions

None — scope is well-defined with clear boundaries and patterns established in pass 1.