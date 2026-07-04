# Proposal: Fix MCP E2E Test Script Parsing

## Problem

The e2e test script `scripts/test-mcp-e2e.sh` reports a false failure on Step 5 (list_object_types tool call) even though the full request chain works correctly end-to-end.

**What actually happens:**
```
Gateway → mcp-server → objects-service → DB → response flows back through all hops ✅
```

The test script fails because it uses the wrong JSON path to extract structured content from the MCP JSON-RPC response:

```bash
# ❌ Current (broken): looks for structuredContent nested inside content array element
TYPE_NAMES=$(echo "$LIST_RESPONSE" | jq -r '.result.content[0].structuredContent // empty')

# ✅ Correct: StructuredContent is a top-level field on result, alongside content
TYPE_NAMES=$(echo "$LIST_RESPONSE" | jq -r '.result.structuredContent // empty')
```

## Root Cause

The mcp-go library serializes `CallToolResult` with **two parallel fields**:
- `content[]` — unstructured text (double-encoded JSON string for backwards compatibility)
- `structuredContent` — parsed objects, at the root level of result

```json
{
  "result": {
    "content": [{"type":"text","text":"[{\"id\":1,...}]"}],
    "structuredContent": [{...parsed types...}]   ← HERE, not inside content[0]!
  }
}
```

The test script incorrectly assumes `structuredContent` is nested inside the first content element. This affects:
- **Step 5** — list_object_types parsing (fails to find data)
- **Step 6** — get_object_type parsing (skips because TYPE_NAMES is empty from Step 5 failure)

## Scope

Fix only the test script — no application code changes needed. The MCP server and full request chain are working correctly.

## Impact

After fix: all 6 e2e steps will pass, confirming Gateway → mcp-server → objects-service data flow with populated type_keys.
