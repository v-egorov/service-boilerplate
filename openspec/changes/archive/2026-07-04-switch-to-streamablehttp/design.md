## Context

The mcp-server currently uses MCP's SSE transport (mcp-go v0.55.1 `NewSSEServer`), which requires two endpoints:
- `GET /sse` — establishes the session, returns an endpoint event with a message submission URL (`/message?sessionId=abc`)
- `POST /message?sessionId=...` — clients POST JSON-RPC messages to this path

The API Gateway proxies both endpoints with special handling: it rewrites `/mcp/message` → `/message` and catches `http.ErrAbortHandler` panics from the SSE reverse proxy. This creates three problems:
1. Path rewriting adds fragility (the gateway must know mcp-go's internal routing)
2. Test scripts must parse SSE streams with timing-sensitive sleep loops
3. The panic recovery is a workaround for control-flow panics, not a real error

StreamableHTTP (mcp-go v0.55.1 `NewStreamableHTTPServer`) uses a single endpoint (`/mcp`) with HTTP headers for session management:
- Session ID in `MCP-Session-ID` header (not URL query param)
- POST requests are synchronous — response comes back immediately
- Optional SSE upgrade on POST responses for async notifications (not used by our read-only tools)

## Goals / Non-Goals

**Goals:**
- Replace SSE with StreamableHTTP as the sole transport in mcp-server
- Enable stateful sessions from day one to support future mutable MCP operations
- Eliminate all SSE-specific code: path rewriting, panic recovery, stream parsing in tests
- Make gateway proxying trivial (single endpoint, no header/body manipulation)

**Non-Goals:**
- Adding mutable tool surface area (create/update/delete objects) — auth plumbing for writes is prepared but not implemented
- Supporting both SSE and StreamableHTTP simultaneously (no backward compatibility)
- Migrating other services to StreamableHTTP (this change is MCP-server specific)
- Implementing server-to-client notification streaming via GET /mcp (supported by mcp-go but not needed today)

## Decisions

### Decision 1: Use StreamableHTTP with stateful sessions (`WithStateful(true)`)

**Rationale:** Stateless sessions generate a new session ID per request, which works for purely synchronous tool calls but doesn't support future mutable operations where identity needs to be established once and reused. Stateful sessions use sticky session management — the first POST establishes the session, and subsequent requests echo back `MCP-Session-ID` in the header.

**Alternatives considered:**
- Stateless (`StatelessGeneratingSessionIdManager`) — simpler, but requires re-authenticating on every call for mutable operations later
- Custom session manager with JWT caching — more control, but unnecessary complexity when built-in stateful mode works

### Decision 2: Single endpoint `POST /mcp` on gateway (no path rewriting)

**Rationale:** StreamableHTTP uses a single HTTP endpoint that handles all JSON-RPC communication via POST. The gateway forwards requests at `/mcp` without any path manipulation — incoming path matches outgoing path exactly. This eliminates the existing fragile rewrite logic (`if strings.HasPrefix(c.Request.URL.Path, "/mcp/message") { c.Request.URL.Path = "/message" }`).

**Alternatives considered:**
- Keep two routes (GET /sse + POST /message) for backward compatibility — rejected per proposal scope
- Add a catch-all route at `/message` as a safety net — too magical, could shadow future endpoints

### Decision 3: Remove `http.ErrAbortHandler` panic recovery in gateway

**Rationale:** This recovery was needed because `httputil.ReverseProxy` panics with `http.ErrAbortHandler` when an SSE stream disconnects (control flow, not a real error). StreamableHTTP has no long-lived streams — each request is a discrete POST with a synchronous response. No panic can occur from client disconnect during streaming because there's no streaming connection to disconnect.

### Decision 4: Keep gateway-trust skip for GET-only reads; prepare Authorization forwarding for writes

**Rationale:** The existing auth model (gateway injects `X-User-ID` → objects-service permiddleware skips auth-service check for GET requests) works perfectly for read-only tools and shouldn't change. For future mutable operations, the identity forwarding code (`identity.go`) already supports forwarding `Authorization` headers when present in context — this will be used by write tools that need real JWT validation through objects-service's normal RBAC path.

## Architecture

```
┌────────────── Client ──────────────┐      ┌────────────── API Gateway ──────────────┐     ┌────────────── mcp-server ──────────────┐
│                                      │      │                                            │     │                                       │
│ POST /mcp                            │      │  POST /mcp                                 │     │  NewStreamableHTTPServer                │
│   Content-Type: application/json     │─────▶│  ProxyMCPRequest                           │─────▶│    WithEndpointPath("/mcp")             │
│   MCP-Session-ID: abc                │      │    No path rewrite                         │     │    WithStateful(true)                   │
│                                      │      │    Body read + rewind (unchanged)          │     │                                       │
│ Response ◀───────────────────────────│◀────│  Response ◀────────────────────────────────│◀────│  handlePost                           │
│ HTTP/200                             │      │                                            │     │    → resolveSessionIdManager           │
│   Content-Type: application/json     │      │  POST /mcp (subsequent calls)              │     │    → creates stateful session          │
│   MCP-Session-ID: abc                │─────▶│    Same single endpoint, no rewrite        │     │    → stores user identity in context   │
│   Body: {result}                     │      │                                            │     │                                       │
└──────────────────────────────────────┘      └────────────────────────────────────────────┘     └───────────────────────────────────────────┘

Auth chain (unchanged semantics):
  mcp-server extracts identity from request.Header → stores in context
  httpGet() reads from context → forwards X-User-* headers to objects-service
  objects-service permiddleware: gateway-trust skip for GET requests only
```

## Risks / Trade-offs

| Risk | Impact | Mitigation |
|------|--------|------------|
| `WithStateful(true)` requires sticky sessions in multi-instance deployments | Low — mcp-server runs as single container in Docker Compose, no scaling concern | Acceptable for dev mode; document for production if scaling needed |
| Body reading + rewinding in gateway proxy still happens for all POSTs | Negligible overhead — same as SSE path; future optimization possible by detecting GET requests (no body) | Defer — not a performance bottleneck |
| Test script rewrite loses SSE-specific edge case coverage | Low — no one tests SSE disconnect scenarios in dev mode; synchronous POST tests are simpler and more reliable | Tests now verify actual tool execution, which is the real goal |
| StreamableHTTP probe fails on some network configurations (corporate proxies, etc.) | Very low — adapter probes SH first, falls back to SSE automatically if needed. We're removing SSE so this means "use SH or nothing" | Document that clients must support StreamableHTTP transport |

## Migration Plan

**Deployment steps:**
1. Deploy updated mcp-server (StreamableHTTP) and gateway (single route) simultaneously — no rolling update needed since they deploy together via Docker Compose
2. Restart all services: `make down && make dev-detached`
3. Verify with e2e test script: `bash scripts/test-mcp-e2e.sh`
4. Verify adapter connection works

**Rollback:** Revert the single commit — mcp-server returns to SSE, gateway restores two routes + path rewriting. No data migration needed (sessions are in-memory only).

## Open Questions

1. **Should we add `WithSessionIdleTTL()` for cleanup of abandoned sessions?** Not urgent — sessions persist until container restart. Can be added later if memory pressure becomes an issue with many concurrent clients.
