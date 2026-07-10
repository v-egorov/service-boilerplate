## 1. Wrap service-layer errors with common/errors sentinels (auth_service.go)

- [x] 1.1 Added `"github.com/v-egorov/service-boilerplate/common/errors"` import to auth_service.go
- [x] 1.2 Wrapped "invalid credentials" returns (~2 sites: Login, Register fallback) → fmt.Errorf("...: %w", errors_pkg.ErrUnauthorized)
- [x] 1.3 Wrapped token-related failures (token not found, revoked, invalid, expired, wrong type) → fmt.Errorf("...: %w", errors_pkg.ErrUnauthorized) (~9 sites across Logout and RefreshToken methods)
- [x] 1.4 Wrapped sql.ErrNoRows detections in GetPermission and GetRole → fmt.Errorf("...: %w", errors_pkg.ErrNotFound) (2 sites at lines ~764, ~773)
- [x] 1.5 "cannot delete role/permission" business rule violations left as-is (no sentinel needed — handler has hardcoded status codes for these)

## 2. Key rotation manager — NO CHANGES

- [x] 2.1 key_rotation_manager.go errors are infrastructure failures with no clear sentinel mapping; left as-is since they'll default to 500 in dispatcher
- [x] 2.2 No import needed

## 3. Extract HandleAuthError dispatcher on AuthHandler struct

- [x] 3.1 Added shared `HandleAuthError(c *gin.Context, err error)` method to auth_handler.go:
  - Logs the error with request ID context
  - Uses errors.Is() to match ErrUnauthorized → HTTP 401 "unauthorized" (uses err.Error() as message)
  - Uses errors.Is() to match ErrNotFound → HTTP 404 "not_found"
  - Falls through default to HTTP 500 "internal_error"
- [x] 3.2 Kept existing type-switches for auth-specific types (ScopedVariantConflictError, PermissionParseError) that use errors.As() — these stay as-is in their respective methods

## 4. Replace inline errorResponse()/validationError() calls in auth_handler.go (~76 sites)

- [x] 4.1 Login: replaced service-layer error handling with HandleAuthError (1 site)
- [x] 4.2 Register: replaced service-layer error handling with HandleAuthError (1 site)
- [x] 4.3 Logout: kept as-is for auth failures (no handler response on failure, just logs — no change needed)
- [x] 4.4 RefreshToken: replaced errorResponse() call with HandleAuthError (1 site)
- [x] 4.5 GetCurrentUser: replaced all inline errorResponse() calls with HandleAuthError (~6 sites)
- [x] 4.6 ValidateToken: replaced errorResponse() calls with HandleAuthError (~3 sites)
- [x] 4.7 RotateKeys: replaced errorResponse() calls with HandleAuthError (2 sites)
- [x] 4.8 AuthMiddleware: kept inline errorResponse() — middleware returns early, different pattern
- [x] 4.9 Role CRUD methods (CreateRole, UpdateRole, DeleteRole): replaced service-layer error handling with HandleAuthError (~6 sites across methods)
- [x] 4.10 Permission CRUD methods (CreatePermission, UpdatePermission, DeletePermission): replaced service-layer error handling with HandleAuthError (~6 sites across methods)
- [x] 4.11 Assign/Remove Permission-to-Role: replaced errorResponse() calls with HandleAuthError (~3 sites per method)
- [x] 4.12 Role-User management (AssignRoleToUser, RemoveRoleFromUser, UpdateUserRoles): replaced errorResponse() calls with HandleAuthError (~3 sites across methods)

## 5. Replace inline validationError() calls for request-parsing failures

- [x] 5.1 Login/Register/RefreshToken: kept validationError() — these are request parsing (JSON binding), not service errors
- [x] 5.2 Role/Permission CRUD: kept validationError() — same pattern, request parsing failures always → 400
- [x] 5.3 Verified NO-OP task — confirmed that no request-parsing calls were mistakenly replaced in step 4 (15 remaining validationError/errorResponse calls are all request-validation, context-validation, or domain-specific)

## 6. Re-export typed error types from models/errors.go

- [x] 6.1 Replaced local type definitions for ValidationError, ConflictError, NotFoundError, InternalError with re-export aliases to common/errors (type aliases preserving struct identity)
- [x] 6.2 Added constructor function aliases (var NewXxx = errors.NewXxx) for backward compat
- [x] 6.3 Kept PermissionParseError and ScopedVariantConflictError as local auth-service types

## 7. Build, test, and verify

- [x] 7.1 `make build-auth-service` compiles without errors
- [x] 7.2 `go test ./services/auth-service/...` — all tests pass (handler tests updated to use wrapped sentinels in mocks)
- [x] 7.3 Verified ~60 inline `errorResponse()` calls for service-layer errors replaced; 15 remaining are request-validation/context/validationError (correct to keep)
- [x] 7.4 HandleAuthError handles all three sentinel types correctly: ErrUnauthorized→401, ErrNotFound→404, default→500
