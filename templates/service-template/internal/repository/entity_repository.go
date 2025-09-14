package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	// ENTITY_IMPORT_MODELS
)

// Import models package for Entity type
// This will be replaced with proper import during template processing

type EntityRepository struct {
	db     *pgxpool.Pool
	logger *logrus.Logger
}

func NewEntityRepository(db *pgxpool.Pool, logger *logrus.Logger) *EntityRepository {
	return &EntityRepository{
		db:     db,
		logger: logger,
	}
}

func (r *EntityRepository) Create(ctx context.Context, entity *models.Entity) (*models.Entity, error) {
	query := `
		INSERT INTO entities (name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	now := time.Now()
	entity.CreatedAt = now
	entity.UpdatedAt = now

	err := r.db.QueryRow(ctx, query,
		entity.Name,
		entity.Description,
		entity.CreatedAt,
		entity.UpdatedAt,
	).Scan(&entity.ID)

	if err != nil {
		r.logger.WithError(err).Error("Failed to create entity")
		return nil, err
	}

	r.logger.WithField("id", entity.ID).Info("Entity created in database")
	return entity, nil
}

func (r *EntityRepository) GetByID(ctx context.Context, id int64) (*models.Entity, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM entities
		WHERE id = $1`

	var entity models.Entity
	err := r.db.QueryRow(ctx, query, id).Scan(
		&entity.ID,
		&entity.Name,
		&entity.Description,
		&entity.CreatedAt,
		&entity.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.WithField("id", id).Debug("Entity not found")
			return nil, err
		}
		r.logger.WithError(err).Error("Failed to get entity by ID")
		return nil, err
	}

	return &entity, nil
}

func (r *EntityRepository) Replace(ctx context.Context, id int64, entity *models.Entity) (*models.Entity, error) {
	if id <= 0 {
		return nil, errors.New("invalid entity ID")
	}

	query := `
		UPDATE entities
		SET name = $1, description = $2, updated_at = $3
		WHERE id = $4
		RETURNING id, name, description, created_at, updated_at`

	entity.UpdatedAt = time.Now()

	var updatedEntity models.Entity
	err := r.db.QueryRow(ctx, query,
		entity.Name,
		entity.Description,
		entity.UpdatedAt,
		id,
	).Scan(
		&updatedEntity.ID,
		&updatedEntity.Name,
		&updatedEntity.Description,
		&updatedEntity.CreatedAt,
		&updatedEntity.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.WithField("id", id).Warn("Entity not found for replacement")
			return nil, errors.New("entity not found")
		}
		r.logger.WithError(err).WithField("id", id).Error("Failed to replace entity")
		return nil, err
	}

	r.logger.WithField("id", id).Info("Entity replaced in database")
	return &updatedEntity, nil
}

func (r *EntityRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) (*models.Entity, error) {
	// Build dynamic update query
	query := "UPDATE entities SET updated_at = $1"
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

	var entity models.Entity
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&entity.ID,
		&entity.Name,
		&entity.Description,
		&entity.CreatedAt,
		&entity.UpdatedAt,
	)

	if err != nil {
		r.logger.WithError(err).Error("Failed to update entity")
		return nil, err
	}

	r.logger.WithField("id", id).Info("Entity updated in database")
	return &entity, nil
}

func (r *EntityRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM entities WHERE id = $1"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.WithError(err).Error("Failed to delete entity")
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.WithField("id", id).Warn("No entity found to delete")
		return sql.ErrNoRows
	}

	r.logger.WithField("id", id).Info("Entity deleted from database")
	return nil
}

func (r *EntityRepository) List(ctx context.Context, limit, offset int) ([]*models.Entity, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM entities
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		r.logger.WithError(err).Error("Failed to list entities")
		return nil, err
	}
	defer rows.Close()

	var entities []*models.Entity
	for rows.Next() {
		var entity models.Entity
		err := rows.Scan(
			&entity.ID,
			&entity.Name,
			&entity.Description,
			&entity.CreatedAt,
			&entity.UpdatedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan entity row")
			return nil, err
		}
		entities = append(entities, &entity)
	}

	if err := rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating entity rows")
		return nil, err
	}

	r.logger.WithField("count", len(entities)).Debug("Entities listed from database")
	return entities, nil
}

func (r *EntityRepository) Count(ctx context.Context) (int64, error) {
	query := "SELECT COUNT(*) FROM entities"

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		r.logger.WithError(err).Error("Failed to count entities")
		return 0, err
	}

	return count, nil
}
