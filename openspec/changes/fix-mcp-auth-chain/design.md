## Context

The MCP server exposes read-only tools/resources/prompts that proxy to objects-service. The gateway injects `X-User-ID`, `X-User-Email`, `X-User-Roles` on every `/mcp/*` request, and mcp-go's SSE transport (v0.55.1, `server/sse.go:651`) copies those headers into the message context; `server/request_handler.go:125+` then propagates them onto every handler's `request.Header` (`CallToolRequest`, `ReadResourceRequest`, `GetPromptRequest`).

The break is purely at hop 2: `ObjectsClient.httpGet()` in `services/mcp-server/internal/client/objects_client.go` builds a new `*http.Request` with no headers. objects-service permiddleware sees an empty user context and rejects every call with 401. The existing `mcp-server` spec covers header preservation only on gateway → mcp-server; the onward hop was never specified, so this is both a code fix and a spec gap.

This design supersedes the four options explored in `docs/mcp-auth-chain-decisions.md`, which proposed Option A (explicit `http.Header` param). After review we are going with context-based injection (Option C) instead — see Decisions.

## Goals / Non-Goals

**Goals:**
- Restore identity flow on the mcp-server → objects-service hop so tool calls succeed end-to-end
- Use a single idiom (request `context.Context`) consistent with how identity is already threaded inside the backend services (`common/middleware/auth.go`)
- Keep a tight, explicit forward list so only the three identity headers cross the boundary (nothing wholesale-copied)
- Add a regression test that fails today and proves the fix

**Non-Goals:**
- Adding JWT validation at mcp-server (gateway-trust model stays)
- Forwarding identity from health checks (they hit `/health` unauthenticated by design)
- Fixing migration 00012 user-role linkage (separate known issue)
- Adding `list_object_types` filter params or full recursive hierarchy resource (separate spec gaps)
- Validating `scripts/test-mcp-e2e.sh` end-to-end (separate)

## Decisions

### Decision: Context-based identity injection over explicit `http.Header` param (Option C over A)

`ObjectsClient` data methods gain a leading `ctx context.Context` parameter; `httpGet(ctx, url)` reads identity out of the context and copies the allow-listed headers onto the outbound request. Handlers do one line at entry: `ctx = client.WithIdentity(ctx, request.Header)`.

**Rationale (why C over A):**
- **Idiomatic consistency.** Identity already flows via request context inside the backend services (`common/middleware/auth.go` sets `user_id`/`user_email`/`user_roles` on the gin context). Using context out of mcp-server keeps one model for the whole codebase; Option A would split identity into two mechanisms.
- **Extensibility.** The project has a tracing guide in scope. W3C `Traceparent` (and any future request-scoped header) rides the same context for free. With Option A, each new header type means re-touching every call site to thread another param.
- **Equal call-site effort.** Both options require touching every call site — Option A adds a param, Option C adds a `WithIdentity` line. Neither is heavier in practice.

**Alternatives considered:**
- *Option A (explicit `http.Header` param).* Simplest in isolation but diverges from the codebase's existing identity-via-context pattern and pays the same per-call-site cost.
- *Option B (wrapper/embedding).* Rejected: Go has no virtual dispatch on embedded types, so a wrapper cannot intercept `httpGet` without overriding every public method — high complexity for a simple header-forwarding problem.
- *Option D (closure injector with shared state + mutex).* Overkill for per-request processing; introduces concurrency concerns Option C avoids entirely.

### Decision: Unexported forward list, single source of truth

`httpGet` copies headers from a single unexported `var forwardHeaders = []string{"X-User-ID", "X-User-Email", "X-User-Roles"}`. Unexported deliberately: the only consumer is `httpGet`, exporting it would invite accidental cross-package mutation of the forwarding contract, and the regression test is behavior-based (assert exactly these arrive) rather than slice-iteration-based.

### Decision: No-identity-in-context ⇒ proceed with no identity headers

When `IdentityFromContext(ctx)` returns nil (health checks, background callers, unit tests), `httpGet` builds the request with no identity headers and proceeds. This is the semantic health checks depend on — they intentionally hit `/health` unauthenticated. We document this as an intentional behavior (not an implementation detail) so a future change does not "tighten" it into a rejection without revisiting health checks.

### Decision: Scope identity to the 7 data methods; leave health code path untouched

`HealthCheckURL()` returns a string (no HTTP) and `checkObjectsServiceHealth()` builds its own request to `/health` — neither carries identity, neither needs the `ctx` threading. The health handler keeps calling `objClient.HealthCheckURL()` exactly as today.

## Risks / Trade-offs

- **Signature churn on 7 methods** → every call site in tools/resources/prompts is touched once. Mitigation: mechanical change, one-line handler edits, compile-enforced (the build breaks immediately if a call site is missed).
- **Handler could forget the `WithIdentity` line** → a new tool/resource/prompt added later might silently skip identity and regress to 401. Mitigation: the regression test covers one tool end-to-end as a sentinel; an explicit note in tasks.md and the spec scenario codifies the expectation.
- **Context carries the header map by reference** → if a handler mutated `request.Header` between `WithIdentity` and the outbound call, the change would be visible. Mitigation: handlers never mutate inbound headers; copying is unnecessary complexity and not added.
- **Forward list is a maintenance point** → adding `Traceparent` later means editing one slice. This is the intended trade-off: one explicit list over wholesale forwarding.
