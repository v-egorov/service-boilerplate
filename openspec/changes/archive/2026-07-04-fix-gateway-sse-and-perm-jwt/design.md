## Context

Two infrastructure bugs prevent the MCP server from functioning. Both are already investigated in `docs/known-issues.md` (sections "Gateway SSE Panic" and "MCP Auth Chain Permission Check Gap"). This design formalizes the fixes.

### SSE panic

`ProxyMCPRequest()` uses `httputil.ReverseProxy` to stream SSE to clients. When a client disconnects (or the SSE stream naturally ends), Go's reverse proxy calls `panic(http.ErrAbortHandler)` as a **control flow mechanism** — not a real error, just signaling "I'm done, don't write headers." Gin's `gin.Recovery()` middleware catches this as a real panic and writes a 500 response (and truncates the trace span), logging:

```
[Recovery] panic recovered: net/http: abort Handler
[WARNING] Headers were already written. Wanted to override status code 200 with 500
```

The fix is standard Go: wrap the call in a deferred recover that discriminates `http.ErrAbortHandler` from real panics.

### Permission check gap

After `fix-mcp-auth-chain`, identity headers flow correctly: `Gateway → mcp-server → objects-service`. Authentication is fine — `common/middleware/auth.go` reads `X-User-ID` and sets `user_id` in gin context when `jwtSecret` is nil (dev mode).

But authorization breaks at objects-service → auth-service:

```
permiddleware.NewPermissionMiddleware():
  jwtToken = c.GetHeader("Authorization")  // empty — no JWT on MCP path
  authClient.CheckPermission(userID, permission, jwtToken)
        │
        ▼
  auth-service: RequireAuth() → user_id empty → 401
```

The `authClient.CheckPermission()` call creates a bare HTTP request to auth-service — only `Content-Type` and optionally `Authorization` (empty here). It does not forward `X-User-*` headers. Auth-service's JWT middleware never sets user context, so `RequireAuth()` rejects.

In production this path is unreachable — services are isolated behind the gateway, and the gateway always validates JWTs. But in dev mode with MCP (no JWT validation by design), the round-trip to auth-service is the gap.

## Goals / Non-Goals

**Goals:**
- Fix SSE reverse proxy crashes so persistent event streams work
- Restore MCP tool call functionality (currently every tool returns 401 from auth-service)
- Keep changes minimal and scoped to dev-mode paths only
- Preserve existing RBAC behavior for all authenticated (JWT-bearing) requests

**Non-Goals:**
- Adding JWT validation at mcp-server or MCP routes
- Implementing API-key-based MCP client identity
- Forwarding `X-User-*` headers through `authClient` to auth-service (Option A — rejected as implicit trust papering)
- Injecting a shared internal token across services (Option C — overengineered for dev mode)
- Modifying auth-service's JWT or RequireAuth middleware behavior

## Decisions

### Decision 1: Recover http.ErrAbortHandler in ProxyMCPRequest

Wrap `proxy.ServeHTTP(c.Writer, c.Request)` with a deferred recover:

```go
defer func() {
    if r := recover(); r != nil {
        if _, ok := r.(http.ErrAbortHandler); !ok {
            panic(r) // Re-panic if not from ReverseProxy abort
        }
        // http.ErrAbortHandler is expected — do nothing
    }
}()
```

**Why not move this to the gin middleware layer:**
- `ErrAbortHandler` is specific to `httputil.ReverseProxy` behavior, not a general concern
- Gin's `Recovery()` is global — modifying it would affect all handlers
- Inline recover is idiomatic Go for this exact scenario (see Go docs for httputil.ReverseProxy)

**Why not use a ResponseWriter wrapper or custom transport:**
- Overkill — the deferred recover is 5 lines and handles the problem at the exact call site
- ResponseWriter wrappers add complexity without benefit for this single use case
- Custom transport would bypass ReverseProxy entirely, breaking SSE streaming (chunked transfer encoding)

### Decision 2: Skip auth-service permission check in permiddleware for gateway-trusted GET requests

In `permiddleware.go:NewPermissionMiddleware()`, add a check before the `authClient.CheckPermission()` loop:

```go
// When the gateway vouches for identity (X-User-ID present) but no JWT token
// is available (MCP dev-mode path), skip the auth-service round-trip and trust
// the gateway directly. In production, all requests carry valid JWTs, so this
// branch never triggers.
if jwtToken == "" && userID != "" && cfg.HTTPMethod == "GET" {
    c.Set("matched_permissions", requiredPermissions)
    c.Next()
    return
}
```

**Rationale:**

| Option | Approach | Why we rejected |
|--------|----------|-----------------|
| A | Forward `X-User-ID` to auth-service | Implicit trust — auth-service can't distinguish "gateway vouched" from "header spoofed." JWT middleware works but the semantics are muddy. |
| **B (chosen)** | Skip auth-service call when no JWT + identity present | Clean separation — gateway IS the authorizer for dev-mode MCP reads. One file, ~3 lines. In production branch never triggers. |
| C | Shared internal token | Overengineered — new infrastructure for a dev-mode concern. |

**Why only GET requests:**
The MCP server is read-only (delta 1). GET is the only HTTP method used. Restricting to GET makes the intent explicit and prevents accidental write bypass.

**Why check `userID != ""` (not just `jwtToken == ""`):**
A request with no identity AND no JWT should still fail — that's unauthenticated traffic. The skip only applies when identity IS present (gateway vouched) but the normal JWT-based authorization path can't complete.

**Dev-only safety:**
This is NOT a general authorization bypass. In production:
- The gateway always validates JWTs → `Authorization` header is always present → `jwtToken` is never empty
- Service isolation (Kubernetes, no external routes to objects-service) prevents direct access
- The `jwtToken == ""` branch is unreachable

## Risks / Trade-offs

- **MCP reads bypass RBAC in dev mode** → the `mcp-agent` system user effectively has unrestricted read access. Mitigation: this is intentional for delta 1; proper per-user MCP authorization requires API keys + real JWT injection (separate scope, documented in known-issues.md).
- **Permiddleware has two authorization paths** → one for JWT-bearing requests (normal RBAC via auth-service), one for gateway-trusted identity (skip). Mitigation: the conditions are mutually exclusive (`jwtToken == "" && userID != ""`), and in production the second path is unreachable.
- **ErrAbortHandler recovery is inline, not reusable** → if another handler needs `httputil.ReverseProxy`, the recover pattern must be duplicated. Mitigation: the current `ProxyRequest` handler does NOT need this (regular proxying doesn't trigger the panic the same way — it's SSE-specific). If it becomes common, extract to a helper later.
