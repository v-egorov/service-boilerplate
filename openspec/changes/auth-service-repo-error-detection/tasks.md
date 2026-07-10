## 1. Create repository/interfaces.go with sentinel aliases and interface definition

- [ ] 1.1 Create `services/auth-service/internal/repository/interfaces.go` as new file
- [ ] 1.2 Add import for `"github.com/v-egorov/service-boilerplate/common/errors"` aliased as `errors_pkg`
- [ ] 1.3 Define sentinel alias: `ErrNotFound = errors_pkg.ErrNotFound` (package-level var)
- [ ] 1.4 Define RepositoryInterface with all current repo method signatures (copy from auth_repository.go method receivers, excluding constructor functions)
- [ ] 1.5 Add doc comment referencing Rule 7 of docs/go-coding-conventions.md

## 2. Modify GetAuthTokenByHash — JWT token lookup used by Logout/RefreshToken/ValidateToken

- [ ] 2.1 Add `errors` import to auth_repository.go if not present
- [ ] 2.2 After the TraceDBQuery call in GetAuthTokenByHash (line ~50), add sql.ErrNoRows detection: check `if errors.Is(err, sql.ErrNoRows)` → return `nil, ErrNotFound`
- [ ] 2.3 Keep existing error handling for non-ErrNoRows cases (`return nil, err`)

## 3. Modify GetUserSession — session lookup used by auth middleware

- [ ] 3.1 After the QueryRow call in GetUserSession (line ~97), add sql.ErrNoRows detection → return `nil, ErrNotFound`
- [ ] 3.2 Keep existing error handling for non-ErrNoRows cases

## 4. Modify GetRoleByName — role name lookup used during user registration for default role assignment

- [ ] 4.1 After the QueryRow call in GetRoleByName (line ~202), add sql.ErrNoRows detection → return `nil, ErrNotFound`
- [ ] 4.2 Keep existing error handling for non-ErrNoRows cases

## 5. Modify GetRole — UUID-based role lookup used in permission CRUD assignment/validation

- [ ] 5.1 After the QueryRow call in GetRole (line ~265), add sql.ErrNoRows detection → return `nil, ErrNotFound`
- [ ] 5.2 Keep existing error handling for non-ErrNoRows cases

## 6. Modify GetPermission — UUID-based permission lookup used in role-permission management

- [ ] 6.1 After the QueryRow call in GetPermission (line ~335), add sql.ErrNoRows detection → return `nil, ErrNotFound`
- [ ] 6.2 Keep existing error handling for non-ErrNoRows cases

## 7. Remove manual sql.ErrNoRows checks from service layer (auth_service.go)

- [ ] 7.1 In AssignPermissionToRole (~line 764): remove `if errors.Is(err, sql.ErrNoRows)` block that wraps with ErrNotFound — repo now returns ErrNotFound directly
- [ ] 7.2 In AssignPermissionToRole (~line 773): remove `if errors.Is(err, sql.ErrNoRows)` block for GetRole call — same rationale
- [ ] 7.3 Verify no other manual sql.ErrNoRows checks exist in auth_service.go (Delta 1 only covered these 2)

## 8. Build and verify compilation

- [ ] 8.1 Run `go build ./services/auth-service/...` — must compile without errors
- [ ] 8.2 Verify no unused imports remain after removing sql.ErrNoRows checks from service layer

## 9. Update tests for new error paths

- [ ] 9.1 Review auth_repository_test.go — update mocks that return pgx.ErrNoRows to instead return ErrNotFound where the affected methods are tested
- [ ] 9.2 Verify handler tests still pass (HandleAuthError dispatcher already handles ErrNotFound → 404)
- [ ] 9.3 Run `go test ./services/auth-service/...` — all tests must pass

## 10. Final verification

- [ ] 10.1 Verify zero behavioral change: compiled binary returns same HTTP status codes for existing error paths (401 for auth failures, 500 for infra)
- [ ] 10.2 Confirm all single-row GET methods now have sql.ErrNoRows detection at repo level (5 methods modified)
- [ ] 10.3 Confirm collection-returning methods (ListRoles, ListPermissions, GetUserRoles, etc.) are NOT modified — they return empty slices per Rule 1
