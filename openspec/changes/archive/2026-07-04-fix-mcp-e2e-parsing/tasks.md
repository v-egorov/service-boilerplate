## 1. Fix E2E Test Script Parsing Paths

- [x] 1.1 Step 5: Fix jq path from `.result.content[0].structuredContent` → `.result.structuredContent`
- [x] 1.2 Step 6: Fix jq path from `.result.content[0].structuredContent.name` → `.result.structuredContent.name`
- [x] 1.3 Step 6: Fix SSE data prefix handling — use `extract_json()` helper (was raw grep with `data: ` prefix)

## 2. Verification

- [x] 2.1 Run `bash scripts/test-mcp-e2e.sh` — all 6 steps pass
- [x] 2.2 Confirmed: list_object_types returns 15 types with populated type_keys (product-electronics, article-blog-post, etc.)
