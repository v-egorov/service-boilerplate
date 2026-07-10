## 1. Define domain-specific sentinel errors in auth_service.go

- [x] 1.1 Add `errors` import to services/auth_service.go if not present (check current imports)
- [x] 1.2 Define `ErrRoleInUse = errors.New("role has assigned users and cannot be deleted")` near existing service-layer error definitions
- [x] 1.3 Define `ErrPermissionInUse = errors.New("permission has assigned roles and cannot be deleted")` alongside ErrRoleInUse

## 2. Wrap DeleteRole/DeletePermission business rule returns with domain sentinels

- [x] 2.1 In DeleteRole (~line 668): replace raw string return with `return fmt.Errorf("role deletion blocked: %w", ErrRoleInUse)`
- [x] 2.2 In DeletePermission (~line 747): replace raw string return with `return fmt.Errorf("permission deletion blocked: %w", ErrPermissionInUse)`

## 3. Move DB duplicate detection from handlers to service layer CreateRole/CreatePermission

- [x] 3.1 Verify `"strings"` import exists in auth_service.go (needed for strings.Contains)
- [x] 3.2 In CreateRole (~line 610): after `s.repo.CreateRole()` fails, add:
  ```go
  if err != nil {
      s.logger.WithError(err).Error("Failed to create role")
      // Check if it's a constraint violation (duplicate role name)
      if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "23505") {
          return nil, fmt.Errorf("role already exists: %w", errors_pkg.ErrAlreadyExists)
      }
      return nil, fmt.Errorf("failed to create role: %w", err)
  }
  ```
- [x] 3.3 In CreatePermission (~line 689): same pattern after `s.repo.CreatePermission()` fails

## 4. Update HandleAuthError dispatcher with new sentinel cases

- [x] 4.1 Add import for auth-service services package in auth_handler.go (check current imports)
- [x] 4.2 After existing ErrNotFound case, add:
  ```go
  case errors.Is(err, services.ErrRoleInUse),
      errors.Is(err, services.ErrPermissionInUse):
      statusCode = http.StatusUnprocessableEntity
      errorMessage = err.Error()
      errorType = "validation_error"
  ```
- [x] 4.3 Add ErrAlreadyExists case (for DB duplicate key detection now in service layer):
  ```go
  case errors.Is(err, errors_pkg.ErrAlreadyExists):
      statusCode = http.StatusConflict
      errorMessage = err.Error()
      errorType = "conflict"
  ```

## 5. Remove inline strings.Contains("duplicate key") from CreateRole/CreatePermission handlers

- [x] 5.1 In CreateRole handler (~line 425): remove the `if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "23505")` block entirely — service layer now wraps with ErrAlreadyExists
- [x] 5.2 In CreatePermission handler (~line 568): same removal

## 6. Build and verify compilation

- [x] 6.1 Run `go build ./services/auth-service/...` — must compile without errors
- [x] 6.2 Verify no unused imports remain (check for stale strings import if no longer needed)

## 7. Update tests for new error paths

- [x] 7.1 Review auth_service_test.go — verify DeleteRole/DeletePermission tests still pass with wrapped sentinels
- [x] 7.2 Review auth_handler_test.go — update mock errors to use wrapped sentinels where handlers test CreateRole/CreatePermission duplicate detection
- [x] 7.3 Run `go test ./services/auth-service/...` — all tests must pass

## 8. Final verification

- [x] 8.1 Verify zero behavioral improvement: DeleteRole/DeletePermission now return HTTP 422 (not 500) for "resource in use" conditions
- [x] 8.2 CreateRole/CreatePermission no longer contain inline strings.Contains() detection — all error paths flow through HandleAuthError
- [x] 8.3 Confirm HandleAuthError dispatcher has exactly 4 sentinel cases: ErrUnauthorized (401), ErrNotFound (404), domain sentinels (422), ErrAlreadyExists (409)
