# MCP Server for Objects-Service — Exploration Notes

## Date: 2026-07-01

**Status**: Decisions finalized — ready for change proposal.

---

## Context: What objects-service provides

```
objects-service architecture:
┌─────────────────────────────────────────────────────┐
│  ObjectType (schema/type hierarchy)                 │
│    └── Object (instance, hierarchical tree)         │
│       ├── Metadata (JSONB)                          │
│       ├── Tags ([]string)                           │
│       └── Relationships (edges between objects)     │
│           └── RelationshipType (itself an Object!)  │
├────────────┬─────────────┬──────────────┬───────────┤
│  ObjectType│   Object    │   Relation-  │ RelationshipType
│ Handler    │   Handler   │ shipHandler  │ Handler
├────────────┼─────────────┼──────────────┼───────────┤
│ CRUD +     │ CRUD +      │ CRUD +       │ CRUD
│ tree ops   │ search/     │ graph queries│           │
│            │ find by     │              │           │
│            │ metadata    │              │           │
└────────────┴─────────────┴──────────────┴───────────┘

Service interface (~30+ methods):
- ObjectType: CRUD + tree traversal (children, parents)
- Object: CRUD + search + find_by_metadata + find_by_tags + bulk ops + graph queries (ancestors/descendants/path/children)
- RelationshipType: CRUD
- Relationship: CRUD + graph queries
```

---

## Approach Chosen: New Service Behind API Gateway

```
┌─────────────┐     ┌──────────────────┐   ┌──────────────────┐   ┌──────────────────┐
│  LLM Client │SSE  │  api-gateway     ├──▶│  mcp-server      ├──▶│  objects-service │
│  (Hermes,   │     │  (JWT auth)      │HTTP│  (new service)   │REST│  (existing API)  │
│   pi agent) │     └──────────────────┘   │                  │API│                  │
└─────────────┘                            └──────────────────┘   └──────────────────┘
                                           No DB access —                               
                                           calls objects-service HTTP REST API only
```

**Pros:**
- Clean separation of concerns — MCP server is its own service
- Reuses api-gateway's JWT auth + X-User-* header forwarding (same pattern objects-service uses)
- Calls objects-service via HTTP REST API — zero DB-level coupling, no repository code duplication
- Independent lifecycle and deployment
- Can use mcp-go SDK directly (production-quality, well-maintained)
- Consistent with existing architecture — follows the same "new service behind gateway" pattern

**Cons:**
- New service in docker-compose (but standard pattern, not novel)
- api-gateway does NOT support WebSocket upgrades — transport limited to SSE
- Needs its own config (standard) but NO database connection needed — HTTP-only layer

**Why this over alternatives:**
- **vs embedded**: Separates concerns cleanly; no gin/mcp framework coupling in one process
- **vs gateway adapter**: Gateway has no business logic layer; MCP needs rich tool definitions that don't belong in a proxy

---

## Key Design Questions (Unresolved)

### 1. Transport Layer
MCP supports multiple transports:

| Transport | Pros | Cons | Decision |
|-----------|------|------|----------|
| **SSE** (Server-Sent Events) | HTTP-based, works behind any proxy/load balancer, api-gateway can forward it, simplest to implement | One-way server→client only; MCP client must reconnect for requests | ✅ **Chosen** — fits gateway architecture |
| **Stdio** | CLI/tool invocation style, no network overhead, standard for local MCP tools | Requires direct process launch, doesn't work behind api-gateway proxy, not suitable for multi-agent scenarios | ❌ Not viable behind gateway |
| WebSockets | Bidirectional, low latency, true long-lived connection | api-gateway does NOT support WebSocket upgrades — would break the architecture | ❌ Out of scope |

**Note**: SSE is actually ideal here because MCP's request/response pattern over HTTP already works fine via SSE (the client sends requests on the same channel), and it fits perfectly behind our api-gateway proxy.

### 2. Tool Surface Area — First Delta Scope

**Decision**: Read-only only for first delta. Skip write operations, skip auth complexity.

```
FIRST DELTA — Path B: types + minimal object queries:
├── list_object_types     ← browse all object types
├── get_object_type       ← by ID or type_key
├── list_objects          ← filter objects (type_key only, no advanced search)
└── get_object            ← by public_id

FUTURE DELTAS (write operations):
├── create_object         ← with type validation (skipped for now)
├── update_object         ← partial updates
├── delete_object         ← soft delete
├── add_remove_tags       ← tag management
└── update_metadata       ← JSONB mutations
```

**Rationale**: Read-only is safe, doesn't require auth, and covers the most valuable agent use cases (browsing, searching, understanding schema). The MCP server will run behind api-gateway which handles JWT + user-context forwarding, but for read-only tools we don't need to check permissions at all — agents just browse data.

### 3. Authentication Model — First Delta

**Decision**: Skip auth entirely for first delta.

The MCP server runs behind api-gateway, which already:
1. Validates JWT tokens
2. Forwards user identity via `X-User-ID`, `X-User-Email`, `X-User-Roles` headers
3. Objects-service trusts these headers in dev mode (when jwtSecret is nil — see `common/middleware/auth.go`)

For read-only tools, we don't need per-user permission checks. The MCP server can:
- Skip auth middleware entirely for first delta
- Or optionally pass through the user context header if needed later
- Auth + RBAC comes in a future delta when write operations are added

**How objects-service handles this today**: `common/middleware/auth.go` has a hybrid path — when jwtSecret is nil, it reads from gateway headers instead of validating JWT. This is the pattern we follow.

### 4. Resource URIs — What data should be "readable"?

MCP isn't just tools — resources let LLMs browse structured data.

For first delta, focus on the most fundamental resource: **object types** (the schema layer). Objects themselves can be accessed via tools (`get_object`, `list_objects`).

```
FIRST DELTA RESOURCES:
├── objects-types              ← full type hierarchy (read-only schema browser)

FUTURE RESOURCES:
├── object/{public_id}         ← single object + metadata
├── object-tree/{object_id}    ← ancestry chain
└── relationships/{source_id}  ← outgoing edges from an object
```

### 5. Prompt Templates — Reusable Agent Workflows

**Decision**: Include minimal prompts from the start (2-3 thin wrappers).

**Why include them?**
Prompt templates are actually the LOWEST-effort MCP feature:
- A prompt = name + description + argument schema + message-returning function
- Argument schemas **reuse** existing tool params — no duplicate work
- They're metadata on top of existing functions, not new business logic
- The agent gets significantly better UX (named workflows vs raw API surface)

**Scope creep risk**: Not in implementation cost (~50 lines), but in adding too many variants nobody uses.

**Concrete first-delta proposal — exactly 2 prompts:**
1. `browse_schema` — "show me all object types and their hierarchy" (wraps list_object_types + get_object_type)
2. `get_object_info` — "tell me everything about this object including its type and metadata" (wraps list_objects + get_object)

These are thin wrappers around existing tool calls. The agent gets better UX, minimal additional code.

---

## Decisions Made

| Decision | Value |
|----------|-------|
| Architecture | New service behind api-gateway (follows existing pattern) |
| Transport | SSE only — gateway can't do WebSocket |
| Scope (delta 1) | Path B: 4 read-only tools + object types resource + 2 prompt templates |
| Auth | Skipped for delta 1 (gateway handles JWT, read-only = no RBAC needed) |
| Consumers | Hermes Agent / pi agent (open-source agentic solutions) |
| Context | Internal dev use |
| Prompt templates | Include minimal set — 2 thin wrappers around core tools |
| Write operations | Skip for now — future deltas |

## Implementation Notes

- **mcp-go** is the Go SDK to use (production-quality, well-maintained)
- SSE transport means MCP client connects once via HTTP GET `/sse`, then sends requests on same channel
- api-gateway routes: `POST /mcp-server/sse` → mcp-server service
- First delta should be **lean** — just the absolutely needed MCP server infra + minimal API surface

---

## Next Steps

- [x] Explore architecture options → new service behind gateway chosen
- [x] Decide transport → SSE (gateway limitation)
- [x] Define scope → read-only first delta, lean
- [x] Auth model → skip for delta 1
- [ ] Create OpenSpec change proposal with design/specs/tasks
