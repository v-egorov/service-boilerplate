package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	// SERVICE_IMPORT_MODELS
)

// Service type alias - will be replaced during template processing
type Service = struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ServiceRepository struct {
	db     *pgxpool.Pool
	logger *logrus.Logger
}

func NewServiceRepository(db *pgxpool.Pool, logger *logrus.Logger) *ServiceRepository {
	return &ServiceRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ServiceRepository) Create(ctx context.Context, service *Service) (*Service, error) {
	query := `
		INSERT INTO services (name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	now := time.Now()
	service.CreatedAt = now
	service.UpdatedAt = now

	err := r.db.QueryRow(ctx, query,
		service.Name,
		service.Description,
		service.CreatedAt,
		service.UpdatedAt,
	).Scan(&service.ID)

	if err != nil {
		r.logger.WithError(err).Error("Failed to create service")
		return nil, err
	}

	r.logger.WithField("id", service.ID).Info("Service created in database")
	return service, nil
}

func (r *ServiceRepository) GetByID(ctx context.Context, id int64) (*Service, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM services
		WHERE id = $1`

	var service Service
	err := r.db.QueryRow(ctx, query, id).Scan(
		&service.ID,
		&service.Name,
		&service.Description,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.WithField("id", id).Debug("Service not found")
			return nil, err
		}
		r.logger.WithError(err).Error("Failed to get service by ID")
		return nil, err
	}

	return &service, nil
}

func (r *ServiceRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) (*Service, error) {
	// Build dynamic update query
	query := "UPDATE services SET updated_at = $1"
	args := []interface{}{time.Now()}
	argCount := 1

	if name, ok := updates["name"].(string); ok {
		argCount++
		query += ", name = $" + string(rune(argCount+'0'))
		args = append(args, name)
	}

	if description, ok := updates["description"].(string); ok {
		argCount++
		query += ", description = $" + string(rune(argCount+'0'))
		args = append(args, description)
	}

	query += " WHERE id = $" + string(rune(argCount+1+'0'))
	args = append(args, id)

	query += " RETURNING id, name, description, created_at, updated_at"

	var service Service
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&service.ID,
		&service.Name,
		&service.Description,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err != nil {
		r.logger.WithError(err).Error("Failed to update service")
		return nil, err
	}

	r.logger.WithField("id", id).Info("Service updated in database")
	return &service, nil
}

func (r *ServiceRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM services WHERE id = $1"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.WithError(err).Error("Failed to delete service")
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.WithField("id", id).Warn("No service found to delete")
		return sql.ErrNoRows
	}

	r.logger.WithField("id", id).Info("Service deleted from database")
	return nil
}

func (r *ServiceRepository) List(ctx context.Context, limit, offset int) ([]*Service, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM services
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		r.logger.WithError(err).Error("Failed to list services")
		return nil, err
	}
	defer rows.Close()

	var services []*Service
	for rows.Next() {
		var service Service
		err := rows.Scan(
			&service.ID,
			&service.Name,
			&service.Description,
			&service.CreatedAt,
			&service.UpdatedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan service row")
			return nil, err
		}
		services = append(services, &service)
	}

	if err := rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating service rows")
		return nil, err
	}

	r.logger.WithField("count", len(services)).Debug("Services listed from database")
	return services, nil
}

func (r *ServiceRepository) Count(ctx context.Context) (int64, error) {
	query := "SELECT COUNT(*) FROM services"

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		r.logger.WithError(err).Error("Failed to count services")
		return 0, err
	}

	return count, nil
}
