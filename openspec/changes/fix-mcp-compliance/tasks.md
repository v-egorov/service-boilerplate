## 1. Define output schema result types

- [ ] 1.1 Create `ListObjectTypesResult` struct in type_tools.go: `{Types []ObjectType}` with json:"types" tag (domain-specific key)
- [ ] 1.2 Create `GetObjectTypeResult` struct in type_tools.go: single ObjectType (no wrapper needed for single object)
- [ ] 1.3 Create `ListObjectsResult` struct in object_tools.go: `{Objects []Object}` with json:"objects" tag (domain-specific key)
- [ ] 1.4 Create `GetObjectResult` struct in object_tools.go: single Object (no wrapper needed)

## 2. Add output schemas to tool registrations

- [ ] 2.1 In type_tools.go: add `mcp.WithOutputSchema[ListObjectTypesResult]()` to list_object_types tool definition
- [ ] 2.2 In type_tools.go: add `mcp.WithOutputSchema[GetObjectTypeResult]()` to get_object_type tool definition
- [ ] 2.3 In object_tools.go: add `mcp.WithOutputSchema[ListObjectsResult]()` to list_objects tool definition
- [ ] 2.4 In object_tools.go: add `mcp.WithOutputSchema[GetObjectResult]()` to get_object tool definition

## 3. Fix structuredContent serialization for array tools

- [ ] 3.1 In type_tools.go — list_object_types handler: wrap result as `ListObjectTypesResult{Items: types}` before assigning StructuredContent; replace full-data JSON in Content with summary string
- [ ] 3.2 In object_tools.go — list_objects handler: wrap result as `ListObjectsResult{Items: objects}` before assigning StructuredContent; replace full-data JSON in Content with summary string
- [ ] 3.3 Verify get_object_type and get_object return maps (not arrays) — no wrapping needed for single-object tools
- [ ] 3.4 In type_tools.go — get_object_type handler: replace full-data JSON in Content with `"get_object_type: <name> (id=<id>)"` summary
- [ ] 3.5 In object_tools.go — get_object handler: replace full-data JSON in Content with `"get_object: <object_name> (id=<public_id>)"` summary

## 4. Add missing filter parameter: parent_type_id to list_object_types

- [ ] 4.1 Add `ParentTypeID *int64` field to `ListObjectTypesParams` struct in type_tools.go
- [ ] 4.2 In objects_client.go: add `ParentTypeID int64` field to `ListObjectTypesParams` and wire it as query param `parent_type_id`
- [ ] 4.3 Verify objects-service repository supports parent_type_id filter (check object_repository.go for existing method)

## 5. Add missing cross-type query: type_key_prefix on list_objects

- [ ] 5.1 In object_tools.go: add `TypeKeyPrefix *string` field to `ListObjectsParams` struct
- [ ] 5.2 In objects_client.go: add `TypeKeyPrefix string` field and wire as query param `type_key_prefix`
- [ ] 5.3 Verify objects-service supports type_key_prefix filter (check object_repository.go — may need new method for cross-type resolution)

## 6. Build, test, and verify

- [ ] 6.1 Run `go build ./...` in services/mcp-server to verify compilation
- [ ] 6.2 Run `make dev-detached` to restart mcp-server through gateway
- [ ] 6.3 Run `bash scripts/test-mcp-e2e.sh` — all steps must pass (0 failures); update structuredContent access from `.items` to domain-specific keys (`types`, `objects`)
- [ ] 6.4 Verify tools/list output includes outputSchema for each tool via Python HTTP client test
- [ ] 6.5 Verify structuredContent is a dict with correct domain key (`types` or `objects`) for list operations, not bare array
- [ ] 6.6 Verify content[] text fields are ~30 chars (summary) instead of multi-KB JSON strings — confirm token savings


