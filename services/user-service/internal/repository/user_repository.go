package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/vegorov/service-boilerplate/services/user-service/internal/models"
)

type UserRepository struct {
	db     *pgxpool.Pool
	logger *logrus.Logger
}

func NewUserRepository(db *pgxpool.Pool, logger *logrus.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (email, first_name, last_name, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query, user.Email, user.FirstName, user.LastName).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		r.logger.WithError(err).Error("Failed to create user")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.WithField("user_id", user.ID).Info("User created successfully")
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	query := `SELECT id, email, first_name, last_name, created_at, updated_at FROM users WHERE id = $1`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.WithError(err).Error("Failed to get user by ID")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, id int, user *models.User) (*models.User, error) {
	query := `
		UPDATE users
		SET email = $1, first_name = $2, last_name = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING id, email, first_name, last_name, created_at, updated_at`

	err := r.db.QueryRow(ctx, query, user.Email, user.FirstName, user.LastName, id).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		r.logger.WithError(err).Error("Failed to update user")
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	r.logger.WithField("user_id", user.ID).Info("User updated successfully")
	return user, nil
}

func (r *UserRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
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
	query := `SELECT id, email, first_name, last_name, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		r.logger.WithError(err).Error("Failed to list users")
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan user")
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating over users")
		return nil, fmt.Errorf("error iterating over users: %w", err)
	}

	return users, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, email, first_name, last_name, created_at, updated_at FROM users WHERE email = $1`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.WithError(err).Error("Failed to get user by email")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}
