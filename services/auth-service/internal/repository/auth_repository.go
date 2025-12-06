package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/v-egorov/service-boilerplate/common/database"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/models"
)

// DBPoolInterface defines the database operations needed
type DBPoolInterface interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

type AuthRepository struct {
	db DBPoolInterface
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{db: db}
}

// NewAuthRepositoryWithInterface creates a repository with a custom database interface (for testing)
func NewAuthRepositoryWithInterface(db DBPoolInterface) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateAuthToken(ctx context.Context, token *models.AuthToken) error {
	query := `
		INSERT INTO auth_service.auth_tokens (id, user_id, token_hash, token_type, expires_at)
		VALUES ($1, $2, $3, $4, $5)`

	return database.TraceDBInsert(ctx, "auth_tokens", query, func(ctx context.Context) error {
		_, err := r.db.Exec(ctx, query,
			token.ID, token.UserID, token.TokenHash, token.TokenType, token.ExpiresAt)
		return err
	})
}

func (r *AuthRepository) GetAuthTokenByHash(ctx context.Context, tokenHash string) (*models.AuthToken, error) {
	query := `
		SELECT id, user_id, token_hash, token_type, expires_at, revoked_at, created_at, updated_at
		FROM auth_service.auth_tokens
		WHERE token_hash = $1 AND (revoked_at IS NULL OR revoked_at > NOW())`

	var token models.AuthToken
	err := database.TraceDBQuery(ctx, "auth_tokens", query, func(ctx context.Context) error {
		return r.db.QueryRow(ctx, query, tokenHash).Scan(
			&token.ID, &token.UserID, &token.TokenHash, &token.TokenType,
			&token.ExpiresAt, &token.RevokedAt, &token.CreatedAt, &token.UpdatedAt)
	})
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *AuthRepository) RevokeAuthToken(ctx context.Context, tokenID uuid.UUID) error {
	query := `UPDATE auth_service.auth_tokens SET revoked_at = NOW() WHERE id = $1`
	return database.TraceDBUpdate(ctx, "auth_tokens", query, func(ctx context.Context) error {
		_, err := r.db.Exec(ctx, query, tokenID)
		return err
	})
}

func (r *AuthRepository) RevokeUserTokens(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE auth_service.auth_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`
	return database.TraceDBUpdate(ctx, "auth_tokens", query, func(ctx context.Context) error {
		_, err := r.db.Exec(ctx, query, userID)
		return err
	})
}

func (r *AuthRepository) CreateUserSession(ctx context.Context, session *models.UserSession) error {
	query := `
		INSERT INTO auth_service.user_sessions (id, user_id, session_token, ip_address, user_agent, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	return database.TraceDBInsert(ctx, "user_sessions", query, func(ctx context.Context) error {
		_, err := r.db.Exec(ctx, query,
			session.ID, session.UserID, session.SessionToken,
			session.IPAddress, session.UserAgent, session.ExpiresAt)
		return err
	})
}

func (r *AuthRepository) GetUserSession(ctx context.Context, sessionToken string) (*models.UserSession, error) {
	query := `
		SELECT id, user_id, session_token, ip_address, user_agent, expires_at, created_at
		FROM auth_service.user_sessions
		WHERE session_token = $1 AND expires_at > NOW()`

	var session models.UserSession
	err := r.db.QueryRow(ctx, query, sessionToken).Scan(
		&session.ID, &session.UserID, &session.SessionToken,
		&session.IPAddress, &session.UserAgent, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *AuthRepository) DeleteUserSession(ctx context.Context, sessionID uuid.UUID) error {
	query := `DELETE FROM auth_service.user_sessions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, sessionID)
	return err
}

func (r *AuthRepository) DeleteExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM auth_service.user_sessions WHERE expires_at <= NOW()`
	_, err := r.db.Exec(ctx, query)
	return err
}

func (r *AuthRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.created_at
		FROM auth_service.roles r
		JOIN auth_service.user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1`

	var roles []models.Role
	err := database.TraceDBQuery(ctx, "roles,user_roles", query, func(ctx context.Context) error {
		rows, err := r.db.Query(ctx, query, userID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var role models.Role
			err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
			if err != nil {
				return err
			}
			roles = append(roles, role)
		}
		return rows.Err()
	})
	return roles, err
}

func (r *AuthRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]models.Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.resource, p.action, p.created_at
		FROM auth_service.permissions p
		JOIN auth_service.role_permissions rp ON p.id = rp.permission_id
		JOIN auth_service.user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var permission models.Permission
		err := rows.Scan(&permission.ID, &permission.Name, &permission.Resource, &permission.Action, &permission.CreatedAt)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, rows.Err()
}

func (r *AuthRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `
		INSERT INTO auth_service.user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, role_id) DO NOTHING`

	_, err := r.db.Exec(ctx, query, userID, roleID)
	return err
}

func (r *AuthRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `DELETE FROM auth_service.user_roles WHERE user_id = $1 AND role_id = $2`
	_, err := r.db.Exec(ctx, query, userID, roleID)
	return err
}

func (r *AuthRepository) GetRoleByName(ctx context.Context, roleName string) (*models.Role, error) {
	query := `SELECT id, name, description, created_at FROM auth_service.roles WHERE name = $1`

	var role models.Role
	err := r.db.QueryRow(ctx, query, roleName).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *AuthRepository) CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM auth_service.permissions p
			JOIN auth_service.role_permissions rp ON p.id = rp.permission_id
			JOIN auth_service.user_roles ur ON rp.role_id = ur.role_id
			WHERE ur.user_id = $1 AND p.resource = $2 AND p.action = $3
		)`

	var exists bool
	err := r.db.QueryRow(ctx, query, userID, resource, action).Scan(&exists)
	return exists, err
}

// Role CRUD methods
func (r *AuthRepository) CreateRole(ctx context.Context, role *models.Role) (*models.Role, error) {
	query := `
		INSERT INTO auth_service.roles (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at`

	err := r.db.QueryRow(ctx, query, role.Name, role.Description).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *AuthRepository) ListRoles(ctx context.Context) ([]models.Role, error) {
	query := `SELECT id, name, description, created_at FROM auth_service.roles ORDER BY name`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *AuthRepository) GetRole(ctx context.Context, roleID uuid.UUID) (*models.Role, error) {
	query := `SELECT id, name, description, created_at FROM auth_service.roles WHERE id = $1`

	var role models.Role
	err := r.db.QueryRow(ctx, query, roleID).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *AuthRepository) UpdateRole(ctx context.Context, roleID uuid.UUID, name, description string) (*models.Role, error) {
	query := `
		UPDATE auth_service.roles
		SET name = $2, description = $3
		WHERE id = $1
		RETURNING id, name, description, created_at`

	var role models.Role
	err := r.db.QueryRow(ctx, query, roleID, name, description).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *AuthRepository) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	query := `DELETE FROM auth_service.roles WHERE id = $1`
	_, err := r.db.Exec(ctx, query, roleID)
	return err
}

// Permission CRUD methods
func (r *AuthRepository) CreatePermission(ctx context.Context, permission *models.Permission) (*models.Permission, error) {
	query := `
		INSERT INTO auth_service.permissions (name, resource, action)
		VALUES ($1, $2, $3)
		RETURNING id, name, resource, action, created_at`

	err := r.db.QueryRow(ctx, query, permission.Name, permission.Resource, permission.Action).Scan(
		&permission.ID, &permission.Name, &permission.Resource, &permission.Action, &permission.CreatedAt)
	if err != nil {
		return nil, err
	}
	return permission, nil
}

func (r *AuthRepository) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	query := `SELECT id, name, resource, action, created_at FROM auth_service.permissions ORDER BY resource, action`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var permission models.Permission
		err := rows.Scan(&permission.ID, &permission.Name, &permission.Resource, &permission.Action, &permission.CreatedAt)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, rows.Err()
}

func (r *AuthRepository) GetPermission(ctx context.Context, permissionID uuid.UUID) (*models.Permission, error) {
	query := `SELECT id, name, resource, action, created_at FROM auth_service.permissions WHERE id = $1`

	var permission models.Permission
	err := r.db.QueryRow(ctx, query, permissionID).Scan(
		&permission.ID, &permission.Name, &permission.Resource, &permission.Action, &permission.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *AuthRepository) UpdatePermission(ctx context.Context, permissionID uuid.UUID, name, resource, action string) (*models.Permission, error) {
	query := `
		UPDATE auth_service.permissions
		SET name = $2, resource = $3, action = $4
		WHERE id = $1
		RETURNING id, name, resource, action, created_at`

	var permission models.Permission
	err := r.db.QueryRow(ctx, query, permissionID, name, resource, action).Scan(
		&permission.ID, &permission.Name, &permission.Resource, &permission.Action, &permission.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *AuthRepository) DeletePermission(ctx context.Context, permissionID uuid.UUID) error {
	query := `DELETE FROM auth_service.permissions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, permissionID)
	return err
}

// Role-Permission management
func (r *AuthRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	query := `
		INSERT INTO auth_service.role_permissions (role_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT (role_id, permission_id) DO NOTHING`
	_, err := r.db.Exec(ctx, query, roleID, permissionID)
	return err
}

func (r *AuthRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	query := `DELETE FROM auth_service.role_permissions WHERE role_id = $1 AND permission_id = $2`
	_, err := r.db.Exec(ctx, query, roleID, permissionID)
	return err
}

func (r *AuthRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]models.Permission, error) {
	query := `
		SELECT p.id, p.name, p.resource, p.action, p.created_at
		FROM auth_service.permissions p
		JOIN auth_service.role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action`

	rows, err := r.db.Query(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var permission models.Permission
		err := rows.Scan(&permission.ID, &permission.Name, &permission.Resource, &permission.Action, &permission.CreatedAt)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, rows.Err()
}

// User role management (bulk operations)
func (r *AuthRepository) UpdateUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	// Start transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Remove all existing roles for user
	_, err = tx.Exec(ctx, `DELETE FROM auth_service.user_roles WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	// Assign new roles
	for _, roleID := range roleIDs {
		_, err = tx.Exec(ctx, `
			INSERT INTO auth_service.user_roles (user_id, role_id)
			VALUES ($1, $2)`, userID, roleID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// Utility methods for validation
func (r *AuthRepository) CountUsersWithRole(ctx context.Context, roleID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM auth_service.user_roles WHERE role_id = $1`

	var count int
	err := r.db.QueryRow(ctx, query, roleID).Scan(&count)
	return count, err
}

func (r *AuthRepository) CountRolesWithPermission(ctx context.Context, permissionID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM auth_service.role_permissions WHERE permission_id = $1`

	var count int
	err := r.db.QueryRow(ctx, query, permissionID).Scan(&count)
	return count, err
}
