## Context

The mcp-server exposes 4 tools (list_object_types, get_object_type, list_objects, get_object) that call objects-service via HTTP REST API. Tool handlers return `*mcp.CallToolResult` with both `Content` and `StructuredContent`.

**Current wire format for a successful tool call:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [{"type":"text","text":"[{\"id\":1,...}]"}],     // ~7KB: full JSON serialization of data
    "structuredContent": [{id:1,name:"Article"}, ...],           // ~7KB: raw objects (bare array!)
    "isError": false
  }
}
```

**Two problems:**

1. `structuredContent` is a bare JSON array for list tools. The MCP spec requires structuredContent to be "a JSON object" (dict/map). Python SDK v1.26.0 enforces this with Pydantic: `StructuredContent = dict[str, Any] | None`. Arrays are rejected at the type level.

2. **Token waste:** Both `content[]` and `structuredContent` carry semantically identical data in different formats — effectively doubling context window usage per tool call. For `list_object_types` with 15 types, that's ~7-8KB of duplicated JSON in every response.

**Root causes:**
- Go slices marshal to JSON arrays (`[...]`), maps marshal to JSON objects (`{}`). Our handlers pass raw Go values directly as `StructuredContent`, so arrays flow through unmodified.
- We double-encode: full JSON string in `content[]` + raw data in `structuredContent`. The spec says this is a "SHOULD" for backward compatibility, but no backward-compatible clients exist — all our MCP SDKs (mcp-go Go, Python 1.26+) natively support structuredContent.

## Goals / Non-Goals

**Goals:**
- Wrap array returns in `{"types": [...]}` or `{"objects": [...]}` (domain-specific keys) so structuredContent is always a JSON object
- Add output schemas to all 4 tools using mcp-go's `WithOutputSchema[T]()`
- Replace full-data double-encoding with summary-only `content[]` — eliminates ~7KB token waste per tool call
- Wire format compatibility: clients can access data at `result.structuredContent.types` / `result.structuredContent.objects` for list operations

**Non-Goals:**
- Fixing missing filter params (`parent_type_id`, `type_key_prefix`) — these are implementation gaps in existing spec requirements, handled as tasks but not spec-level changes

## Decisions

### Decision 1: Wrap array returns using domain-specific keys

**Choice:** Use `{"types": [...]}` for list_object_types and `{"objects": [...]}` for list_objects.

**Rationale:**
- Minimal change to tool handler return values — just wrap the slice before assignment
- Python SDK accepts any `dict[str, Any]`, so domain-specific key wrappers pass validation
- The MCP spec only requires structuredContent to be "a JSON object" — it does not prescribe internal keys; the outputSchema defines the contract
- Domain-specific keys are self-documenting and align with how clients already access the data (e.g., `result.structuredContent.types` is clearer than `result.structuredContent.items`)
- Clients already handle both array and dict structuredContent (e2e test has fallback logic)
- Single-object tools (get_object_type, get_object) already return maps → no wrapping needed

**Alternatives considered:**
1. **Change to `content[]` only, drop structuredContent for arrays** — Loses typed data; clients must parse text field. Not ideal.
2. **Use generic key "items" or "data"** — Works but less self-documenting than domain-specific keys.
3. **Make structuredContent nullable for arrays** — Would require conditional logic per tool; adds complexity without benefit.

### Decision 2: Define output schemas as typed Go structs

**Choice:** Create dedicated result types (e.g., `ListObjectTypesResult`, `GetObjectTypeResult`) and use `WithOutputSchema[T]()`.

**Rationale:**
- mcp-go's `WithOutputSchema[T]()` uses Go generics to derive JSON Schema from struct tags — no manual schema writing needed
- Type safety: handler return values must match the declared type (enforced by compiler)
- `tools/list` response will include accurate output schemas for client validation

**Alternatives considered:**
1. **Manual JSON Schema strings** — Error-prone, not compile-time checked, duplicates type info.
2. **No output schemas** — Keeps current state; Python SDK can't validate results. Not acceptable.

### Decision 3: Use separate result types instead of modifying existing models

**Choice:** Create thin wrapper structs for tool outputs rather than modifying the existing `ObjectType` and `Object` models from objects-service client layer.

**Rationale:**
- Existing models (`internal/client/objects_client.go`) carry raw API response fields; adding output-schema-specific tags would pollute them
- Wrapper types are small (just `{Types []T}`) and live in the tools package where they belong
- Objects-service client models remain focused on HTTP transport contract

### Decision 4: Replace double-encoded `content[]` with summary-only text

**Choice:** Every tool handler replaces its full-data JSON string in `content[]` with a short human-readable summary. The structured data lives exclusively in `structuredContent`.

```go
// Before (current — ~7KB per list call)
return &mcp.CallToolResult{
    Content:         []mcp.Content{mcp.NewTextContent(string(data))},  // full JSON
    StructuredContent: types,
}, nil

// After (~30 chars)
return &mcp.CallToolResult{
    Content:         []mcp.Content{mcp.NewTextContent("list_object_types returned 15 types")},
    StructuredContent: ListObjectTypesResult{Items: types},
}, nil
```

**Summary format per tool:**
| Tool | Summary |
|------|---------|
| `list_object_types` | `"list_object_types returned N types"` (N = count) |
| `get_object_type` | `"get_object_type: <name> (id=<id>)"` |
| `list_objects` | `"list_objects returned N objects for type <type_name>"` |
| `get_object` | `"get_object: <object_name> (id=<public_id>)"` |

**Rationale:**
- Eliminates ~7KB token waste per tool call — context window cost drops from ~8KB to ~30 bytes for the text field
- MCP spec says double-encoding is a "SHOULD" for backward compatibility, not a MUST. No legacy clients exist that need it.
- Summary in `content[]` gives Claude Desktop / UI clients something human-readable to display while structuredContent carries full typed data for programmatic access
- Error paths keep their descriptive error messages (unchanged — they're already short strings)

**Alternatives considered:**
1. **Empty content[]** (`[]`) — Zero token cost but gives no UI feedback in clients that only render content[]. Not user-friendly.
2. **Keep full JSON** — Current state. Token waste is unacceptable for list operations with many results.

**Example:**
```go
// In type_tools.go
type ListObjectTypesResult struct {
    Types []ObjectType `json:"types"`
}
// In object_tools.go
type ListObjectsResult struct {
    Objects []Object `json:"objects"`
}
```



## Risks / Trade-offs

| Risk | Impact | Mitigation |
|------|--------|------------|
| Clients accessing `structuredContent` directly expect array | Medium | E2e test already handles both formats; update client docs |
| Output schema types don't match actual objects-service response data | Low | Add tests to verify schema conformance |
| Adding result types increases maintenance surface | Low | Types are thin wrappers (~5 lines each) |

## Migration Plan

1. Add output schema structs in `internal/tools/` package
2. Update tool handlers to wrap array results and pass schemas via `WithOutputSchema[T]()`
3. Run e2e tests — verify both Python-style dict access (`result["types"]` / `result["objects"]`) and legacy fallback still work
4. No database or config changes required

## Open Questions

1. Should resources (hierarchy) and prompts also declare output schemas? Currently not in scope since they use different result types (`ReadResourceResult`, `GetPromptResult`). This is a best-practice consideration that should be assessed separately if it becomes relevant; tracked as a deferred item.
