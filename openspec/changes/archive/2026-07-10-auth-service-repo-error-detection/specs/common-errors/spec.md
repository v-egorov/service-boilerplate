## ADDED Requirements

### Requirement: Auth-service repository methods detect sql.ErrNoRows at the boundary

The auth-service repository layer MUST detect `sql.ErrNoRows` in all single-row GET methods and return `common/errors.ErrNotFound` directly (unwrapped). This follows Rule 7 of `docs/go-coding-conventions.md` and ensures consistent error flow across services.

Every `Get*()` method that retrieves a single row SHALL check for missing rows before returning:

```go
func (r *AuthRepository) GetRole(ctx context.Context, roleID uuid.UUID) (*models.Role, error) {
    err := r.db.QueryRow(ctx, query, id).Scan(...)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, repository.ErrNotFound  // ← unwrapped sentinel
        }
        return nil, fmt.Errorf("failed to get role: %w", err)
    }
}
```

#### Scenario: GetRole returns ErrNotFound for missing role

- **WHEN** `GetRole` queries for a role that does not exist in the database
- **THEN** it detects `sql.ErrNoRows` and returns `repository.ErrNotFound` (unwrapped sentinel) so that service/handler layers can match via `errors.Is()` or direct comparison

#### Scenario: GetPermission returns ErrNotFound for missing permission

- **WHEN** `GetPermission` queries for a permission ID that does not exist
- **THEN** it detects `sql.ErrNoRows` and returns `repository.ErrNotFound` (unwrapped sentinel)

#### Scenario: GetAuthTokenByHash returns ErrNotFound for missing token

- **WHEN** `GetAuthTokenByHash` queries for an auth token by hash that does not exist
- **THEN** it detects `sql.ErrNoRows` and returns `repository.ErrNotFound` (unwrapped sentinel)

#### Scenario: GetRoleByName returns ErrNotFound for unknown role name

- **WHEN** `GetRoleByName` queries for a role with a name that does not exist
- **THEN** it detects `sql.ErrNoRows` and returns `repository.ErrNotFound` (unwrapped sentinel)

#### Scenario: Collection-returning methods do NOT apply this rule

- **WHEN** a repository method returns a slice (`[]*models.Role`, etc.)
- **THEN** it does NOT check for sql.ErrNoRows — empty results are represented as an empty non-nil slice, not an error (Rule 1 convention)
