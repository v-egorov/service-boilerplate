## Why

The api-gateway still has 2 of 4 internal packages without tests: `handlers/gateway.go` and `middleware/logging.go`. This second pass targets the high-value, easily testable functions — simple health-check handlers (LivenessHandler, PingHandler), ReadinessHandler logic, and the logging middleware's responseWriter wrapper. ProxyRequest and StatusHandler are deferred due to their need for a full mock backend server infrastructure.

## What Changes

- Add unit tests for `api-gateway/internal/handlers/gateway.go`:
  - LivenessHandler: returns JSON with status "ok", timestamp, gateway name, version
  - PingHandler: returns JSON with status "pong" and metadata
  - ReadinessHandler: returns 200 when services registered, 503 when empty registry
- Add unit tests for `api-gateway/internal/middleware/logging.go`:
  - responseWriter: captures WriteHeader (status code) and Write (response size)
  - RequestResponseLogger: skip-paths list is applied (/health, /ready, etc. not logged)

## Capabilities

**No spec-level behavior changes.** This extends test coverage for existing capabilities already defined in `api-gateway-services-tests`.

### New Capabilities
- `api-gateway-rest-tests`: Unit tests for GatewayHandler (Liveness, Ping, Readiness) and logging middleware responseWriter + skip-paths

## Impact

- **New files:** `handlers/gateway_test.go`, `middleware/logging_test.go`
- **Existing files modified:** none (tests are additive only)