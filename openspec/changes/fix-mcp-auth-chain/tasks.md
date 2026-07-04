## 1. Identity plumbing in client package

- [x] 1.1 Create `services/mcp-server/internal/client/identity.go` with `type identityCtxKey struct{}`, `func WithIdentity(ctx context.Context, hdr http.Header) context.Context` (nil-safe), and `func IdentityFromContext(ctx context.Context) http.Header`
- [x] 1.2 Add unexported `var forwardHeaders = []string{"X-User-ID", "X-User-Email", "X-User-Roles"}` in `objects_client.go` (single source of truth for the forward list)

## 2. Thread context through ObjectsClient

- [x] 2.1 Change `httpGet(url)` → `httpGet(ctx, url)`; read `IdentityFromContext(ctx)` and, when non-nil, copy only the headers in `forwardHeaders` onto the outbound request; when nil, proceed with no identity headers
- [x] 2.2 Add a leading `ctx context.Context` param to the 7 data methods and pass it into `httpGet`: `ListObjectTypes`, `GetObjectTypeByID`, `GetObjectTypeByName`, `GetObjectTypeTree`, `GetRootTree`, `ListObjects`, `GetObjectByPublicID`
- [x] 2.3 Leave `HealthCheckURL()` untouched (returns a string, no HTTP, no identity)

## 3. Inject identity at handler entry points

- [x] 3.1 `internal/tools/type_tools.go`: add `ctx = mcpclient.WithIdentity(ctx, request.Header)` before each `objClient.*` call in `list_object_types` and `get_object_type`
- [x] 3.2 `internal/tools/object_tools.go`: same one-liner before each call in `list_objects` and `get_object`
- [x] 3.3 `internal/resources/hierarchy.go`: same one-liner before `objClient.GetRootTree(ctx)`
- [x] 3.4 `internal/prompts/browse_schema.go`: same one-liner before `objClient.ListObjectTypes(ctx)`
- [x] 3.5 `internal/prompts/get_object_info.go`: same one-liner before `objClient.ListObjects(ctx, ...)`
- [x] 3.6 Leave `internal/handlers/health_handler.go` untouched — its `checkObjectsServiceHealth` intentionally hits `/health` unauthenticated and uses `HealthCheckURL()` directly

## 4. Regression test (proves the fix)

- [x] 4.1 Add `services/mcp-server/internal/client/objects_client_identity_test.go`: spin up a stub `httptest` objects-service that records received headers, call `ListObjects(client.WithIdentity(ctx, hdr), ...)` and assert exactly `X-User-ID`, `X-User-Email`, `X-User-Roles` arrive with original values — fails before the fix, passes after
- [x] 4.2 Add a case asserting non-identity inbound headers (e.g. `Authorization`, `Cookie`) are NOT forwarded
- [x] 4.3 Add a case asserting a context with no identity yields a request with no `X-User-*` headers (proceeds, not rejected)

## 5. Build, verify, known-issues follow-up

- [x] 5.1 `make build-mcp-server` compiles cleanly; `make test-mcp-server` passes (existing health tests stay green)
- [x] 5.2 Update `docs/known-issues.md` MCP Server Identity Forwarding section: correct the root-cause line ("mcp-go's session context does not carry HTTP headers" → it does; the gap was handler+client not reading it), mark the section as resolved by this change with a pointer to `design.md`
