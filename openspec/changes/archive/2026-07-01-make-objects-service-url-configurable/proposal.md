## Why

The API gateway has a hardcoded URL for objects-service with no environment-variable override and no development localhost remapping, unlike auth-service and user-service which both support `AUTH_SERVICE_URL` / `USER_SERVICE_URL` env vars. This asymmetry prevents flexible deployment of the api-gateway in environments where objects-service runs on a non-default port or host, and makes local development without Docker less consistent across services.

## What Changes

- Add `OBJECTS_SERVICE_URL` environment variable support for objects-service URL resolution in the API gateway
- Apply development localhost remapping to objects-service (matching auth/user pattern)
- Update docker-compose.yml to include `OBJECTS_SERVICE_URL` env var for api-gateway container

## Capabilities

### Modified Capabilities
- `api-gateway-routing`: Backend Service URL Resolution requirement — objects-service now resolves from environment variable with fallback default and dev localhost remapping, matching the existing pattern for auth-service and user-service.

## Impact

- `api-gateway/cmd/main.go` — add OBJECTS_SERVICE_URL resolution logic (mirrors existing auth/user pattern)
- `docker/docker-compose.yml` — add OBJECTS_SERVICE_URL env var to api-gateway service definition
- No API contract changes; no breaking changes for existing deployments using defaults.