# Specification: API Gateway Observability

## Purpose
Defines the API gateway's contract for request logging, metrics collection and exposure, alert monitoring, and cross-origin resource sharing (CORS). Covers how HTTP requests are logged with structured fields at severity-appropriate levels, how request-level metrics are collected via MetricsCollector and exposed through `/api/v1/metrics`, how alert thresholds on error rate and response time trigger active alerts available at `/api/v1/alerts`, and how CORS headers are applied to all responses.

## Requirements

### Requirement: Request/Response Logging
The gateway SHALL log every incoming HTTP request and its corresponding response using structured JSON logging. Standard gin logger output MUST include method, path, status code, duration, and IP address. Health check endpoints (`/health`, `/ready`, `/live`, `/ping`, `/status`) MUST be excluded from standard request/response logs to avoid noise.

#### Scenario: Request logged with required fields
- **WHEN** a non-health-check request is processed by the gateway
- **THEN** the log entry includes method, path, status code, duration_ms, and IP address in JSON format

#### Scenario: Health endpoints excluded from standard logging
- **WHEN** a request hits `/health`, `/ready`, `/live`, `/ping`, or `/status`
- **THEN** no standard gin logger entry is generated for the request

> **Note (unused capability):** A `DetailedRequestLogger` middleware exists that captures additional fields (request body up to 1MB, response size, user agent, error details) and logs at severity-appropriate levels. It is NOT wired up in the current pipeline — only the basic `RequestResponseLogger()` is active. The detailed logger can be enabled by replacing `RequestResponseLogger()` with `DetailedRequestLogger()` in the middleware chain.

### Requirement: Metrics Collection
The gateway SHALL collect per-request metrics including HTTP method, path, status code, duration, authenticated user ID (if present), request ID, and an error flag (true when status >= 400). These metrics MUST be stored in a `MetricsCollector` instance that is accessible via the logger's `GetMetricsCollector()` method.

#### Scenario: Metrics recorded for each request
- **WHEN** a request is processed through the logging middleware
- **THEN** the gateway records the request's method, path, status code, duration, user ID (if authenticated), request ID, and error flag in the metrics collector

#### Scenario: Authenticated user ID included when available
- **WHEN** an authenticated request is logged
- **THEN** the recorded metric includes the `user_id` from the Gin context

### Requirement: Metrics Endpoint Exposure
The gateway SHALL expose collected metrics at `GET /api/v1/metrics`. The endpoint MUST return a JSON response containing the current state of all recorded metrics. This endpoint does NOT require authentication and is available in all environments.

#### Scenario: Unauthenticated access to metrics
- **WHEN** any request hits `GET /api/v1/metrics`
- **THEN** the gateway returns HTTP 200 with a JSON body containing collected metrics data

> **Note (public endpoint gap):** The `/api/v1/metrics` endpoint is currently unauthenticated. This is acceptable in development but should be guarded by authentication or IP restrictions in production deployments. Expected to be addressed in a future change.

### Requirement: Alert Monitoring Configuration
The gateway SHALL monitor for high error rates and slow response times using configurable thresholds when alerting is enabled. The monitoring interval, error rate threshold (fraction), and response time threshold (milliseconds) are configured via `alerting.enabled`, `error_rate_threshold`, `response_time_threshold_ms`, and `alert_interval_minutes`. When alerting is disabled (`enabled: false`), no monitoring occurs.

#### Scenario: Monitoring enabled with default thresholds
- **WHEN** `alerting.enabled` is true in config
- **THEN** the gateway checks metrics against an error rate threshold (default 0.1 / 10%) and a response time threshold (default 5000ms) at configured intervals

#### Scenario: Monitoring disabled
- **WHEN** `alerting.enabled` is false in config
- **THEN** no alert monitoring occurs regardless of other alerting configuration values

### Requirement: Alerts Endpoint Exposure
The gateway SHALL expose active alerts at `GET /api/v1/alerts`. The endpoint MUST return a JSON response listing currently active alerts. When no alerts are active, the response SHOULD contain an empty list or equivalent representation. This endpoint does NOT require authentication and is available in all environments.

#### Scenario: Alerts returned when active
- **WHEN** alert monitoring detects conditions exceeding configured thresholds
- **THEN** `GET /api/v1/alerts` returns a JSON body listing the active alerts

#### Scenario: Empty response when no alerts active
- **WHEN** no alerting conditions are currently exceeded
- **THEN** `GET /api/v1/alerts` returns an empty list or equivalent representation in JSON format

> **Note (public endpoint gap):** The `/api/v1/alerts` endpoint is currently unauthenticated. This exposes operational metadata about the gateway's health state to external consumers without authorization, and should be guarded by authentication or IP restrictions in production deployments. Expected to be addressed alongside the metrics endpoint guard in a future change.

### Requirement: CORS Header Application
The gateway SHALL include cross-origin resource sharing (CORS) headers on ALL responses, regardless of whether the request passes authentication. The headers MUST include `Access-Control-Allow-Origin: *`, `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`, and `Access-Control-Allow-Headers: Content-Type, Authorization, X-Request-ID`.

#### Scenario: Standard response includes CORS headers
- **WHEN** any request is processed by the gateway
- **THEN** the response includes `Access-Control-Allow-Origin: *`, `Access-Control-Allow-Methods` (GET/POST/PUT/DELETE/OPTIONS), and `Access-Control-Allow-Headers` (Content-Type, Authorization, X-Request-ID)

#### Scenario: OPTIONS preflight returns 204 No Content
- **WHEN** a request uses the HTTP OPTIONS method
- **THEN** the gateway responds with HTTP 204 No Content after setting CORS headers and does NOT forward to any handler or proxy
