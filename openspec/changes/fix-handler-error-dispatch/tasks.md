## 1. Create shared error dispatcher

- [x] 1.1 Create `services/objects-service/internal/handlers/error.go` with `HandleError(c *gin.Context, err error, requestID string)` function
- [x] 1.2 Implement switch cases for all ~30 sentinel errors from repository and service layers (see design.md Decision 3 ordering)
- [x] 1.3 Each case writes `c.JSON(statusCode, gin.H{"error": ..., "type": ..., "meta": gin.H{"request_id": requestID}})` following API Response Standards
- [x] 1.4 Verify: the function handles nil errors gracefully (no-op / return)

## 2. Update ObjectHandler — replace handleServiceError calls

- [x] 2.1 `object_handler.go`: Replace all ~25 `h.handleServiceError(c, err, "...", requestID)` calls with a single shared call pattern
- [x] 2.2 Remove the `handleServiceError` method (lines ~75-98) from object_handler.go
- [x] 2.3 Verify: each endpoint that returns a sentinel error now gets correct status code (e.g., GetByID → 404, Create with duplicate → 409)

## 3. Update ObjectTypeHandler — replace handleServiceError calls

- [x] 3.1 `object_type_handler.go`: Replace all ~15 `h.handleServiceError(c, err, "...", requestID)` calls
- [x] 3.2 Remove the broken `handleServiceError` method (switch nil only) from object_type_handler.go
- [x] 3.3 Verify: Create with duplicate type_key → 409 conflict; GetByID not found → 404

## 4. Update RelationshipHandler — replace handleError calls

- [x] 4.1 `relationship_handler.go`: Replace all ~8 `h.handleError(c, requestID, err, "op")` calls
- [x] 4.2 Remove the existing `handleError` method (lines with switch) from relationship_handler.go
- [x] 4.3 Verify: All cases still map correctly (ErrCircularRelationship → 422, ErrNotFound → 404, etc.)

## 5. Update RelationshipTypeHandler — replace handleError calls

- [x] 5.1 `relationship_type_handler.go`: Replace all ~6 `h.handleError(c, requestID, err, "op")` calls
- [x] 5.2 Remove the existing `handleError` method from relationship_type_handler.go
- [x] 5.3 Verify: ErrRelationshipTypeNotFound → 404; ErrDuplicateRelationshipType → 409

## 6. Update tests — fix expected status codes

- [x] 6.1 Run `go test ./services/objects-service/internal/handlers/... -v` to identify all failing assertions
- [x] 6.2 Update any test mocks or expectations that assert HTTP 500 for errors that should now return 404/409/422
- [x] 6.3 Run `go test ./services/objects-service/internal/... -v` — all tests pass

## 7. Add handler error-dispatch regression tests

Currently zero handler tests assert on HTTP status codes for error paths.

- [x] 7.1 Create `error_test.go` in `internal/handlers/` with a mock service that returns each sentinel error type
- [x] 7.2 Test: `HandleError` maps `ErrNotFound` → 404 (covers object, object_type, relationship, relationship_type variants)
- [x] 7.3 Test: `HandleError` maps duplicate/conflict errors → 409
- [x] 7.4 Test: `HandleError` maps validation errors (circular, cardinality) → 422/400
- [x] 7.5 Test: `HandleError` maps unknown errors → 500
- [x] 7.6 Test: `HandleError(nil)` is a no-op (no response written)
- [x] 7.7 Run tests — all pass with explicit status code assertions

## 7. Build and deploy verification

- [x] 7.1 Run `go build ./services/objects-service/...` — expect zero errors
- [x] 7.2 Verify: `curl 'http://localhost:8085/api/v1/objects/999999' -H 'X-User-ID: dev.admin@example.com'` → returns HTTP 404 with `"type": "not_found"` (was 500)
- [x] 7.3 Verify: `curl -X POST 'http://localhost:8085/api/v1/object-types' -H 'Content-Type: application/json' -d '{"name":"dup","type_key":"object-types"}'` → returns HTTP 409 with `"type": "conflict"` (was 500)
- [x] 7.4 Verify MCP tool calls return proper error responses instead of opaque 500
