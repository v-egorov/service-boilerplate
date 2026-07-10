// Package repository provides data access for auth-service.
// Rule 7 (docs/go-coding-conventions.md): All single-row Get*() methods MUST detect
// sql.ErrNoRows and return ErrNotFound directly (unwrapped sentinel) to ensure consistent
// error flow across services. Collection-returning methods return empty non-nil slices,
// not errors.
package repository

import (
	"context"

	"github.com/google/uuid"
	errors "github.com/v-egorov/service-boilerplate/common/errors"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/models"
)

// ErrNotFound is an alias for common/errors.ErrNotFound so that service-layer code
// referencing repository.ErrNotFound continues to work via pointer equality (err == ErrNotFound).
var ErrNotFound = errors.ErrNotFound

// RepositoryInterface defines all operations supported by AuthRepository.
type RepositoryInterface interface {
	CreateAuthToken(ctx context.Context, token *models.AuthToken) error
	GetAuthTokenByHash(ctx context.Context, tokenHash string) (*models.AuthToken, error)
	RevokeAuthToken(ctx context.Context, tokenID uuid.UUID) error
	RevokeUserTokens(ctx context.Context, userID uuid.UUID) error
	CreateUserSession(ctx context.Context, session *models.UserSession) error
	GetUserSession(ctx context.Context, sessionToken string) (*models.UserSession, error)
	DeleteUserSession(ctx context.Context, sessionID uuid.UUID) error
	DeleteExpiredSessions(ctx context.Context) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]models.Permission, error)
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetRoleByName(ctx context.Context, roleName string) (*models.Role, error)
	CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
	CreateRole(ctx context.Context, role *models.Role) (*models.Role, error)
	ListRoles(ctx context.Context) ([]models.Role, error)
	GetRole(ctx context.Context, roleID uuid.UUID) (*models.Role, error)
	UpdateRole(ctx context.Context, roleID uuid.UUID, name, description string) (*models.Role, error)
	DeleteRole(ctx context.Context, roleID uuid.UUID) error
	CreatePermission(ctx context.Context, permission *models.Permission) (*models.Permission, error)
	ListPermissions(ctx context.Context) ([]models.Permission, error)
	GetPermission(ctx context.Context, permissionID uuid.UUID) (*models.Permission, error)
	UpdatePermission(ctx context.Context, permissionID uuid.UUID, name, resource, action string) (*models.Permission, error)
	DeletePermission(ctx context.Context, permissionID uuid.UUID) error
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]models.Permission, error)
	UpdateUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error
	CountUsersWithRole(ctx context.Context, roleID uuid.UUID) (int, error)
	CountRolesWithPermission(ctx context.Context, permissionID uuid.UUID) (int, error)
	DetectConflictingPermission(ctx context.Context, roleID uuid.UUID, permissionName string) error
	ReplaceScopedPermission(ctx context.Context, roleID uuid.UUID, newPermissionName string, newPermissionID uuid.UUID) error
}
