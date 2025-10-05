package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/v-egorov/service-boilerplate/common/database"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/models"
)

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
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
