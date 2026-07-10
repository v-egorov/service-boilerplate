## ADDED Requirements

### Requirement: Auth-service defines domain-specific sentinel errors for business rule violations

Auth-service service layer MUST define domain-specific sentinel error variables for business rule conditions that cannot be expressed with generic sentinels (ErrUnauthorized, ErrNotFound, etc.). These follow the same pattern as objects-service's service-layer sentinels (`ErrRelationshipTypeInUse`, `ErrCircularRelationship`, etc.) and are matched by HandleAuthError via `errors.Is()`.

The following domain-specific sentinels SHALL be defined in `auth-service/internal/services/auth_service.go`:

| Sentinel | HTTP Status | Description |
|----------|-------------|-------------|
| `ErrRoleInUse` | 422 (unprocessable) | Attempt to delete a role that has assigned users |
| `ErrPermissionInUse` | 422 (unprocessable) | Attempt to delete a permission that has roles assigned |

Each sentinel is defined as:
```go
var ErrRoleInUse = errors.New("role has assigned users and cannot be deleted")
var ErrPermissionInUse = errors.New("permission has assigned roles and cannot be deleted")
```

Service layer methods wrap these with `%w` when returning:
```go
return fmt.Errorf("cannot delete role: %d users are assigned to this role", userCount)
// BECOMES:
return fmt.Errorf("role deletion blocked: %w", ErrRoleInUse)
```

#### Scenario: DeleteRole returns ErrRoleInUse via errors.Is() matching

- **WHEN** `DeleteRole` is called for a role that has 3 assigned users
- **THEN** it detects the constraint and returns: `fmt.Errorf("role deletion blocked: %w", ErrRoleInUse)` so that HandleAuthError dispatcher can match via `errors.Is()` and return HTTP 422

#### Scenario: DeletePermission returns ErrPermissionInUse via errors.Is() matching

- **WHEN** `DeletePermission` is called for a permission that has 5 roles assigned
- **THEN** it detects the constraint and returns: `fmt.Errorf("permission deletion blocked: %w", ErrPermissionInUse)` so that HandleAuthError dispatcher can match via `errors.Is()` and return HTTP 422

#### Scenario: HandleAuthError dispatches domain sentinels to correct HTTP status codes

- **WHEN** a service-layer error wrapped with `ErrRoleInUse` reaches the handler
- **THEN** HandleAuthError matches it via `errors.Is(err, services.ErrRoleInUse)` and returns HTTP 422 Unprocessable Entity with error type "validation_error" — NOT HTTP 500
