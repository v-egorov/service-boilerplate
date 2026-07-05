## Problem Analysis

### Bug 1: Where() Hardcoded Placeholder

```
QueryBuilder.Where("object_type_id = $1", 3)
→ query: "WHERE object_type_id = $1"   args: [3]   argIndex: 2

QueryBuilder.Where("created_by = $1", uuid_str)
→ query: "WHERE object_type_id = $1 AND created_by = $1"   args: [3, uuid_str]   argIndex: 3

PostgreSQL resolves: WHERE object_type_id = 3 AND created_by = 3
                                ^^^^^^^ varchar       → CRASH (42883)
```

The `argIndex` field is incremented but never used in Where() — it's only consumed by methods like `WhereJsonContains`, `WhereDateRange`, and `WhereTagsContain`. The `WhereIn()` method correctly generates dynamic placeholders.

### Bug 2: Self-Filtering Design Issue

```
Request Flow:
  Gateway → X-User-ID header injected
    ↓
  objects-service handler List()
    ↓
  filter.UserID = &userID   ← unconditionally applied!
    ↓
  SQL: WHERE created_by = '0000...0001' (MCP agent UUID)
    ↓
  Zero results → empty data or 500 crash (when combined with other filters)
```

Permission checks are already handled at the gateway level via permiddleware. The handler-level self-filter is redundant and incorrect for MCP agents who have read-all permissions.

### Bug 3: WhereTagsContain() Issues

- Branches for `i==0` and `i>0` produce identical output (both `AND`)
- `argIndex += len(tags)` jumps by the full array length instead of incrementing per-tag
- Should use OR logic for subsequent tags: `(tag1 = ANY(tags)) OR (tag2 = ANY(tags))`

## Design

### Fix 1: Where() Dynamic Placeholder Injection

Replace hardcoded `$1` in the condition string with sequential indices starting from current `argIndex`. This handles any placeholder format (`$1`, `$123`) and ensures each Where() call gets distinct placeholders.

```go
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
    runArgIdx := qb.argIndex
    // Scan condition for $N patterns and replace with correct indices
    result := []byte{}
    for i := 0; i < len(condition); {
        if condition[i] == '$' && i+1 < len(condition) {
            numStart := i + 1
            for numStart < len(condition) && condition[numStart] >= '0' && condition[numStart] <= '9' {
                numStart++
            }
            result = append(result, []byte("$"+itoa(runArgIdx))...)
            runArgIdx++
            i = numStart
        } else {
            result = append(result, condition[i])
            i++
        }
    }
    
    if qb.query == "" || !contains(qb.query, "WHERE") {
        qb.query += "WHERE " + string(result) + " "
    } else {
        qb.query += "AND " + string(result) + " "
    }
    qb.args = append(qb.args, args...)
    qb.argIndex = runArgIdx
    return qb
}
```

### Fix 2: Remove Self-Filtering from List Handlers

Remove the unconditional `filter.UserID = &userID` assignment from both `object_handler.go` and `relationship_handler.go`. The permission middleware at the gateway already enforces access control. If clients need ownership filtering, they can pass an explicit query parameter (future enhancement).

```diff
  // Remove these lines from List():
- userID := middleware.GetAuthenticatedUserID(c)
- if userID != "" {
-     filter.UserID = &userID
- }
```

### Fix 3: WhereTagsContain() OR Logic and argIndex

Replace identical branches with proper OR logic for subsequent tags, incrementing argIndex by 1 per tag.

```go
func (qb *QueryBuilder) WhereTagsContain(tags []string) *QueryBuilder {
    firstTag := true
    for _, tag := range tags {
        if firstTag {
            qb.query += "AND ($" + itoa(qb.argIndex) + " = ANY(tags)) "
        } else {
            qb.query += "OR ($" + itoa(qb.argIndex) + " = ANY(tags)) "
        }
        firstTag = false
        qb.args = append(qb.args, tag)
        qb.argIndex++
    }
    return qb
}
```

## Risks and Tradeoffs

| Risk | Mitigation |
|------|-----------|
| Some callers may rely on implicit ownership filtering | No known consumers in dev mode; explicit opt-in query param can be added later if needed |
| Where() placeholder replacement could break edge cases (e.g., `$1` appearing in string literals) | Not currently an issue — all Where() calls use parameterized conditions. Future: consider a more robust parser if string-literal placeholders become necessary |
| argIndex tracking across mixed Where/WhereJsonContains/WhereDateRange calls | Already works correctly for non-Where methods; this fix brings Where into alignment |

## Testing Strategy

1. Unit test QueryBuilder with 3+ chained Where() calls → verify distinct `$N` indices in output SQL
2. Integration test: list objects with `object_type_id` + user ID filter → should return results, not crash
3. Verify existing tests remain green (no behavioral change for single-filter queries)
