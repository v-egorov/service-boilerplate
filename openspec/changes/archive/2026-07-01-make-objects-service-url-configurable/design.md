# Design: Make objects-service URL Configurable in API Gateway

## Context

The api-gateway resolves backend service URLs at startup via environment variables (`AUTH_SERVICE_URL`, `USER_SERVICE_URL`) with built-in Docker service-discovery defaults. The objects-service URL is currently hardcoded as a string literal `http://objects-service:8085` with no env override and no development localhost remapping — an asymmetry that prevents flexible deployment of the api-gateway when objects-service runs on a non-default port or host.

## Goals / Non-Goals

**Goals:**
- Add `OBJECTS_SERVICE_URL` environment variable support for objects-service URL resolution, mirroring the existing pattern used by auth-service and user-service.
- Apply development localhost remapping to objects-service URLs (matching auth/user behavior).
- Update docker-compose.yml to include the env var for api-gateway container.

**Non-Goals:**
- Refactor service URL resolution into a shared utility or config section — that's a larger architectural change better addressed after learning this workflow.
- Add objects-service URL to `config.yaml` — current pattern uses env vars, not config file entries.

## Approach

1. In `api-gateway/cmd/main.go`, add the same env-var resolution + default + dev remapping pattern for objects-service:
   ```go
   objectsServiceURL := os.Getenv("OBJECTS_SERVICE_URL")
   if objectsServiceURL == "" {
       objectsServiceURL = "http://objects-service:8085" // Docker service discovery default
   }
   if cfg.App.Environment == "development" && os.Getenv("DOCKER_ENV") != "true" {
       objectsServiceURL = strings.Replace(objectsServiceURL, "objects-service", "localhost", 1)
   }
   serviceRegistry.RegisterService("objects-service", objectsServiceURL)
   ```

2. In `docker/docker-compose.yml`, add to api-gateway env section:
   ```yaml
   - OBJECTS_SERVICE_URL=${OBJECTS_SERVICE_URL:-http://objects-service:8085}
   ```

3. No changes needed to route definitions, proxy logic, or any other file — only the URL registration call at startup needs modification.

## Risk Assessment

- **Breaking change:** None for existing deployments using defaults. The env var is optional; if unset, behavior is identical (hardcoded fallback).
- **Test risk:** Low — no new code paths, just a different way to configure an existing behavior. Existing tests should continue passing.