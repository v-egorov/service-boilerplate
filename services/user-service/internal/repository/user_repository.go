package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/common/database"
	"github.com/v-egorov/service-boilerplate/services/user-service/internal/models"
)

// DBInterface defines the database operations needed for testing
type DBInterface interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type UserRepository struct {
	db     DBInterface
	logger *logrus.Logger
}

func NewUserRepository(db *pgxpool.Pool, logger *logrus.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

// NewUserRepositoryWithInterface creates a UserRepository with a database interface for testing
func NewUserRepositoryWithInterface(db DBInterface, logger *logrus.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		INSERT INTO user_service.users (email, password_hash, first_name, last_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	err := database.TraceDBInsert(ctx, "user_service.users", query, func(ctx context.Context) error {
		return r.db.QueryRow(ctx, query, user.Email, user.PasswordHash, user.FirstName, user.LastName).Scan(
			&user.ID, &user.CreatedAt, &user.UpdatedAt)
	})
	if err != nil {
		r.logger.WithError(err).Error("Failed to create user")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.WithField("user_id", user.ID).Info("User created successfully")
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `SELECT id, email, password_hash, first_name, last_name, created_at, updated_at FROM user_service.users WHERE id = $1`

	user := &models.User{}
	err := database.TraceDBQuery(ctx, "user_service.users", query, func(ctx context.Context) error {
		return r.db.QueryRow(ctx, query, id).Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)
	})
	if err != nil {
		r.logger.WithError(err).WithField("error_type", fmt.Sprintf("%T", err)).Error("Failed to get user by ID - checking error type")
		if strings.Contains(err.Error(), "no rows in result set") {
			r.logger.WithField("user_id", id).Info("User not found - no rows detected")
			return nil, fmt.Errorf("user not found")
		}
		r.logger.WithError(err).Error("Failed to get user by ID")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, user *models.User) (*models.User, error) {
	query := `
		UPDATE user_service.users
		SET email = $1, password_hash = $2, first_name = $3, last_name = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING id, email, password_hash, first_name, last_name, created_at, updated_at`

	err := database.TraceDBUpdate(ctx, "user_service.users", query, func(ctx context.Context) error {
		return r.db.QueryRow(ctx, query, user.Email, user.PasswordHash, user.FirstName, user.LastName, id).Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.WithError(err).Error("Failed to update user")
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	r.logger.WithField("user_id", user.ID).Info("User updated successfully")
	return user, nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM user_service.users WHERE id = $1`

	var result pgconn.CommandTag
	err := database.TraceDBDelete(ctx, "user_service.users", query, func(ctx context.Context) error {
		var execErr error
		result, execErr = r.db.Exec(ctx, query, id)
		return execErr
	})
	if err != nil {
		r.logger.WithError(err).Error("Failed to delete user")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	r.logger.WithField("user_id", id).Info("User deleted successfully")
	return nil
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `SELECT id, email, password_hash, first_name, last_name, created_at, updated_at FROM user_service.users ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	var users []*models.User
	err := database.TraceDBQuery(ctx, "user_service.users", query, func(ctx context.Context) error {
		rows, err := r.db.Query(ctx, query, limit, offset)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			user := &models.User{}
			err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)
			if err != nil {
				return fmt.Errorf("failed to scan user: %w", err)
			}
			users = append(users, user)
		}

		return rows.Err()
	})
	if err != nil {
		r.logger.WithError(err).Error("Failed to list users")
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, email, password_hash, first_name, last_name, created_at, updated_at FROM user_service.users WHERE email = $1`

	user := &models.User{}
	err := database.TraceDBQuery(ctx, "user_service.users", query, func(ctx context.Context) error {
		return r.db.QueryRow(ctx, query, email).Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.WithError(err).Error("Failed to get user by email")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}
