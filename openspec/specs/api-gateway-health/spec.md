# Specification: API Gateway Health & Status Endpoints

## Purpose

Defines the gateway contract for exposing process liveness, service dependency readiness and comprehensive system status at `/health`, `/ready` (also aliased as `/live`) and `/status`. Covers how each endpoint determines its response code based on backend checks with configurable timeouts. Per-service entries contain URL, rounded to milliseconds from `monitoring.health_check_timeout`, last checked timestamp in RFC339 format, error message if not 200 OK is returned backends that fail their health check with the overall system state (healthy/degraded/unhealthy) and gateway info including memory usage when enabled.

## Requirements

### Requirement: Liveness Endpoint

The gateway SHALL expose a liveness probe at `GET /health`. The endpoint MUST return an immediate 200 OK if the process is alive — it does NOT verify any backend service dependencies or registry health status. Response JSON body contains `"status": "ok"`, timestamp in RFC339 UTC, and version fields from config.yaml values (`app.name` and `version`).

#### Scenario: Process responds with liveness

- **WHEN** a request hits `/health` while the gateway is running
- **THEN** returns HTTP 200 OK JSON body containing `"status": "ok"`, timestamp, name/version fields from config (no dependency checks performed)

### Requirement: Readiness Endpoint Behavior and Alias Confusion Note

The `GET /ready` endpoint MUST verify that at least one backend service is registered in the registry before returning HTTP 200 OK. If zero services are returned as ready status with message `"No services registered"` instead of standard readiness payload matching liveness format (timestamp, name/version).

> **Naming note:** `/live` on this gateway goes to LivenessHandler — it does NOT perform dependency checks or confirm service availability unlike what the name suggests; only `/ready` performs registry-based verification.

#### Scenario: Services available returns ready

- WHEN at least one registered and `GET /ready` called  
  THEN HTTP 200 OK status ok with timestamp/name/version fields matching liveness endpoint format

### Requirement: Health Check Concurrency and Timeout

The gateway SHALL check all registered backend services concurrently using goroutines with WaitGroup synchronization. Each individual health endpoint (`<service_url>/health`) MUST use an HTTP client timeout equal to `monitoring.health_check_timeout` seconds from config.yaml value.

#### Scenario: Timeout on one backend does not block others

- **WHEN** three services registered but auth-service responds within timeout while user/objects fail with timeouts/non‑200 responses  
  THEN overall system is `"degraded"` (not unhealthy) because some passed and some failed

#### Scenario: Timeout on one backend does not block others

- **WHEN** three services registered but auth-service responds within timeout while user/objects fail with timeouts/non‑200 responses  
  THEN overall system is `"degraded"` (not unhealthy) because some passed and some failed

### Requirement: Overall System State Calculation Logic Based On Concurrent Service Check Results Using These Rules

The gateway MUST calculate aggregated status from concurrent backend checks using these rules: no services registered → healthy; all checked are healthy zero failures → **healthy**; at least one but not all fail (partial failure) → degraded (`"degraded"`); ALL backends down completely (all unhealthy/failures) → fully down/unavailable `"unhealthy"`. Gateway self-health is always included in total check count and treated as passing since it's the gateway itself responding.

#### Scenario: All services healthy system reports full availability

- **WHEN** all registered backend return HTTP 200 at their `/health` endpoints within `monitoring.health_check_timeout` seconds

### Requirement: Response Format and Gateway Info

The gateway SHALL include structured information in all health endpoints. Liveness/readiness responses MUST contain timestamp, name/version fields as JSON objects. The status endpoint additionally includes a `gateway` object containing Name (`config.app.name`), Version (from app config), Environment (from `config.app.environment`) with optional MemoryUsage sub-object — allocated bytes, total alloc, sys from `runtime.ReadMemStats` enabled when `monitoring.enable_detailed_metrics=true` in config — and Uptime formatted as `<hours>h<minutes>m<seconds>s`, or `<minutes>m<seconds>s`, or `<seconds>`s based on magnitude.

#### Scenario: Gateway info includes memory when detailed metrics enabled

- **WHEN** `/status` is called and `monitoring.enable_detailed_metrics=true` in config.yaml  
  THEN response contains nested `"memory_usage"` object with alloc, total_alloc, sys fields populated from runtime stats

---

## Notes & Observations

| Item | Detail |
| --- | --- |
| **Liveness vs Readiness** | `/health` checks process only; `/ready` verifies registry not empty. Both skipped by gin logger via `SkipPaths`. |
| **Naming confusion (`/live`)** | `/live` is aliased to LivenessHandler, NOT readiness handler. It returns basic status ok (same as `/health`). Only `/ready` performs dependency checks. |
| **Concurrent health checks** | Checks run in goroutines with WaitGroup synchronization. Each uses `monitoring.health_check_timeout` seconds (default 5s). Failed checks logged at Warn level — no blocking of other checks. |
| **Status calculation rules** | Empty registry → healthy; all passed → healthy; partial failures → degraded; all failed → unhealthy. Gateway self-health always included as +1 in total/passed counts. |
| **Uptime format** | Calculated from `time.Since(startTime)` at startup → HHMMSS if > 0h, MMSS if > 0m, else S only (no leading zeros). |

## What I didn't include

- Health check caching/interval mechanism — currently each `/status` call triggers fresh concurrent checks with no in-memory cache or scheduled background refresh. There's no existing interval-based health-checking logic so nothing to gap-flag here except maybe "health data is computed on-demand not cached" which could be an optimization opportunity but isn't a behavioral requirement yet since it doesn't affect what the endpoint returns today — just performance characteristics under load.

