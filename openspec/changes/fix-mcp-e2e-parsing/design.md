## Context

The e2e test script `scripts/test-mcp-e2e.sh` validates the full MCP request chain: Gateway SSE → mcp-server → objects-service. It currently reports Step 5 (list_object_types tool call) as failed, but inspection of the actual response shows all data flows correctly through every hop — including populated type_keys from the enforce-type-key-mandatory fix.

The failure is purely a test parsing bug: the script accesses `.result.content[0].structuredContent` but mcp-go places `StructuredContent` at root level of `result`, as a sibling of `content[]`.

### MCP Response Structure (mcp-go v0.55.1)

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {"type":"text","text":"[{\"id\":1,\"name\":\"Article\",...}]"}
    ],
    "structuredContent": [
      {"id":1,"name":"Article","type_key":"article",...},
      ...
    ]
  }
}
```

The `content[]` array contains double-encoded JSON (text field is a stringified JSON array). The `structuredContent` at root level contains the actual parsed objects — this is what tools should use.

This dual-format is intentional per mcp-go: *"For backwards compatibility, a tool that returns structured content SHOULD also return functionally equivalent unstructured content."*

## Goals / Non-Goals

**Goals:**
- Fix test script JSON path from `.result.content[0].structuredContent` to `.result.structuredContent` in Steps 5 and 6
- Ensure all 6 e2e steps pass (was 5/6, with Step 6 skipped due to cascading failure)
- Keep the fix minimal — no application code changes

**Non-Goals:**
- No changes to mcp-server or objects-service behavior
- No changes to SSE transport, Gateway routing, or auth chain
- No new test coverage or additional test cases

## Decisions

### Decision 1: Fix only the two broken jq paths

The script has exactly two lines that access `.result.content[0].structuredContent`:
- Line 191 (Step 5 — list_object_types)
- Line 233 (Step 6 — get_object_type)

Both need to become `.result.structuredContent`. This is a one-character path change per line (remove `[0]`).

### Decision 2: No spec changes needed

This is purely an e2e test fix. The MCP protocol behavior is correct — mcp-go serializes `StructuredContent` at root level per spec. No delta specs required.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Test passes but actual chain has a latent bug | The response data includes 15 types with populated type_keys — if the test now parses this correctly, it confirms the full chain works |
| jq path change might break on error responses (no structuredContent) | Script already handles empty/null via `// empty` fallback — unchanged behavior for errors |
