## 1. GatewayHandler tests (handlers/gateway_test.go)

- [ ] 1.1 Create gateway_test.go with test helper to build GatewayHandler from config and registry
- [ ] 1.2 Test LivenessHandler returns HTTP 200 with JSON body containing status "ok", timestamp, gateway name, version
- [ ] 1.3 Test PingHandler returns HTTP 200 with JSON body containing status "pong" and metadata fields
- [ ] 1.4 Test ReadinessHandler returns HTTP 200 when services are registered in backend registry
- [ ] 1.5 Test ReadinessHandler returns HTTP 503 when no services are registered

## 2. Logging middleware tests (middleware/logging_test.go)

- [ ] 2.1 Create logging_test.go with test helper for responseWriter and logger setup
- [ ] 2.2 Test responseWriter captures correct status code from WriteHeader call
- [ ] 2.3 Test responseWriter accumulates written byte counts across multiple Write calls
- [ ] 2.4 Test RequestResponseLogger skip-paths: requests to /health are not logged by the middleware