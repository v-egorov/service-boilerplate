## 1. Update api-gateway URL resolution in main.go

- [ ] 1.1 Add `OBJECTS_SERVICE_URL` environment variable resolution with Docker service-discovery default fallback (3 lines: get env var, check empty, set default)
- [ ] 1.2 Add development localhost remapping for objects-service URLs when env is "development" and DOCKER_ENV != "true" (matching existing auth/user pattern, ~4 lines)
- [ ] 1.3 Replace the hardcoded `serviceRegistry.RegisterService("objects-service", "http://objects-service:8085")` call with the resolved `objectsServiceURL` variable

## 2. Update docker-compose.yml configuration

- [ ] 2.1 Add `OBJECTS_SERVICE_URL=${OBJECTS_SERVICE_URL:-http://objects-service:8085}` to api-gateway service environment section in `docker/docker-compose.yml` (mirrors existing AUTH/USER SERVICE URL entries)

## 3. Verify no regressions
- [ ] 3.1 Build the api-gateway service and verify compilation succeeds
- [ ] 3.2 Run api-gateway tests to confirm no regressions