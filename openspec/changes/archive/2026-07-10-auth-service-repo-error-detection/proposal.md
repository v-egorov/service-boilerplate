## Why

auth-service's repository layer (`auth_repository.go`) does NOT detect `sql.ErrNoRows` at the repository boundary. When a single-row query returns zero rows, pgx propagates raw `pgx.ErrNoRows` through `%w` wrapping:

```go
// auth_repository.go — WRONG pattern (Rule 7 violation)
func (r *AuthRepository) GetRole(ctx context.Context, roleID uuid.UUID) (*models.Role, error) {
    err := r.db.QueryRow(ctx, query, id).Scan(...)
    if err != nil {
        return nil, fmt.Errorf("failed to get role: %w", err)  // raw ErrNoRows passes through!
    }
}
```

This violates **Rule 7** of `docs/go-coding-conventions.md`, which mandates that all repository `Get*()` methods detect `sql.ErrNoRows` and return the `repository.ErrNotFound` sentinel directly (unwrapped). The service layer then does a direct pointer comparison (`err == ErrNotFound`) which fails on wrapped errors, causing "not found" intent to be lost.

In auth-service this means:
- Service-layer sql.ErrNoRows detection is scattered across ~2 handler methods that check it manually
- Other repo GET methods have no not-found handling at all — callers get raw pgx errors with 500 status codes
- Inconsistent pattern vs objects-service (which does repo-level detection in every Get* method)

## What Changes

- Create `auth-service/internal/repository/interfaces.go` with sentinel aliases (`ErrNotFound = common/errors.ErrNotFound`) and the repository interface definition
- Add sql.ErrNoRows → ErrNotFound detection to all single-row GET methods in `auth_repository.go`:
  - `GetAuthTokenByHash` — JWT token lookup (used by Logout, RefreshToken, ValidateToken)
  - `GetUserSession` — session lookup
  - `GetRoleByName` — role name lookup (used during user registration for default role)
  - `GetRole` — UUID-based role lookup (used in permission CRUD assignment/validation)
  - `GetPermission` — UUID-based permission lookup (used in role-permission management)
- Service layer removes its scattered manual sql.ErrNoRows checks (no longer needed at that layer)
- No API behavior change — only internal error flow improvement

## Capabilities

### Modified Capabilities

- **common-errors**: ADD `ErrNotFound` sentinel usage documentation for auth-service repository methods. The sentinel was already defined in the package but never adopted by auth-service's repo layer.

## Impact

- **Affected code**: `auth-service/internal/repository/interfaces.go` (new file), `auth-service/internal/repository/auth_repository.go` (~5 GET methods modified), `auth-service/internal/services/auth_service.go` (~2 manual sql.ErrNoRows checks removed)
- **API behavior change**: None — errors that were already wrapped with ErrUnauthorized/ErrNotFound in the service layer continue to work; errors that weren't wrapped now get proper ErrNotFound at repo level so handler dispatcher can classify them correctly
- **Breaking change**: No
