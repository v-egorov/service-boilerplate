package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
)

// RelationshipTypeRepository defines operations for relationship type data
type RelationshipTypeRepository interface {
	Repository

	Create(ctx context.Context, input *models.CreateRelationshipTypeRequest) (*models.RelationshipType, error)
	GetByID(ctx context.Context, id int64) (*models.RelationshipType, error)
	GetByTypeKey(ctx context.Context, typeKey string) (*models.RelationshipType, error)
	Update(ctx context.Context, id int64, input *models.UpdateRelationshipTypeRequest) (*models.RelationshipType, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter *models.RelationshipTypeFilter) ([]*models.RelationshipType, error)
	Exists(ctx context.Context, typeKey string) (bool, error)
	GetByReverseTypeKey(ctx context.Context, reverseKey string) (*models.RelationshipType, error)
}

// relationshipTypeRepository implements RelationshipTypeRepository
type relationshipTypeRepository struct {
	db      DBInterface
	options *RepositoryOptions
	metrics *RepositoryMetrics
}

// NewRelationshipTypeRepository creates a new RelationshipTypeRepository instance
func NewRelationshipTypeRepository(db DBInterface, options *RepositoryOptions) RelationshipTypeRepository {
	if options == nil {
		options = DefaultRepositoryOptions()
	}

	return &relationshipTypeRepository{
		db:      db,
		options: options,
		metrics: &RepositoryMetrics{LastResetAt: time.Now()},
	}
}

// DB implements Repository interface
func (r *relationshipTypeRepository) DB() DBInterface {
	return r.db
}

// Options implements Repository interface
func (r *relationshipTypeRepository) Options() *RepositoryOptions {
	return r.options
}

// Metrics implements Repository interface
func (r *relationshipTypeRepository) Metrics() *RepositoryMetrics {
	return r.metrics
}

// ResetMetrics implements Repository interface
func (r *relationshipTypeRepository) ResetMetrics() {
	r.metrics.Reset()
}

// Healthy implements Repository interface
func (r *relationshipTypeRepository) Healthy(ctx context.Context) error {
	var result int
	return r.db.QueryRow(ctx, "SELECT 1").Scan(&result)
}

// Create creates a new relationship type
func (r *relationshipTypeRepository) Create(ctx context.Context, input *models.CreateRelationshipTypeRequest) (*models.RelationshipType, error) {
	r.metrics.QueryCount++

	// First, create an object in the objects table to serve as the CTI base
	objectID, err := r.createBaseObject(ctx, input.CreatedBy)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to create base object: %w", err)
	}

	// Serialize validation rules
	var validationRules []byte
	if input.ValidationRules != nil {
		validationRules, err = json.Marshal(input.ValidationRules)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to marshal validation rules: %w", err)
		}
	} else {
		validationRules = []byte("{}")
	}

	// Set default relationship name if not provided
	relationshipName := input.RelationshipName
	if relationshipName == "" {
		relationshipName = input.TypeKey
	}

	// Set default cardinality
	cardinality := input.Cardinality
	if cardinality == "" {
		cardinality = models.CardinalityManyToMany
	}

	query := `
		INSERT INTO objects_service.objects_relationship_types (
			object_id, type_key, relationship_name, reverse_type_key,
			cardinality, required, min_count, max_count, validation_rules,
			created_by, updated_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		RETURNING object_id, type_key, relationship_name, reverse_type_key,
			cardinality, required, min_count, max_count, validation_rules,
			created_by, updated_by, created_at, updated_at`

	var rt models.RelationshipType
	var reverseTypeKey *string

	err = r.db.QueryRow(ctx, query,
		objectID, input.TypeKey, relationshipName, input.ReverseTypeKey,
		cardinality, input.Required, input.MinCount, input.MaxCount, validationRules,
		input.CreatedBy, input.UpdatedBy,
	).Scan(
		&rt.ObjectID, &rt.TypeKey, &rt.RelationshipName, &reverseTypeKey,
		&rt.Cardinality, &rt.Required, &rt.MinCount, &rt.MaxCount, &rt.ValidationRules,
		&rt.CreatedBy, &rt.UpdatedBy, &rt.CreatedAt, &rt.UpdatedAt,
	)

	if err != nil {
		r.metrics.ErrorCount++
		// If we fail to create the relationship type, clean up the base object
		_ = r.deleteBaseObject(ctx, objectID)
		return nil, fmt.Errorf("failed to create relationship type: %w", err)
	}

	rt.ReverseTypeKey = reverseTypeKey
	return &rt, nil
}

// createBaseObject creates an object entry in the objects table for CTI
func (r *relationshipTypeRepository) createBaseObject(ctx context.Context, createdBy string) (int64, error) {
	// Get the RelationshipType object_type_id
	var typeID int64
	err := r.db.QueryRow(ctx, `
		SELECT id FROM objects_service.object_types WHERE name = 'RelationshipType'
	`).Scan(&typeID)
	if err != nil {
		return 0, fmt.Errorf("RelationshipType object type not found: %w", err)
	}

	// Insert the object
	var objectID int64
	err = r.db.QueryRow(ctx, `
		INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, 'RelationshipType:' || $2, 'active', NOW(), NOW())
		RETURNING id
	`, typeID, createdBy).Scan(&objectID)
	if err != nil {
		return 0, fmt.Errorf("failed to create base object: %w", err)
	}

	return objectID, nil
}

// deleteBaseObject deletes the base object from objects table
func (r *relationshipTypeRepository) deleteBaseObject(ctx context.Context, objectID int64) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM objects_service.objects WHERE id = $1
	`, objectID)
	return err
}

// GetByID retrieves a relationship type by object ID
func (r *relationshipTypeRepository) GetByID(ctx context.Context, id int64) (*models.RelationshipType, error) {
	r.metrics.QueryCount++

	query := `
		SELECT object_id, type_key, relationship_name, reverse_type_key,
			cardinality, required, min_count, max_count, validation_rules,
			created_by, updated_by, created_at, updated_at
		FROM objects_service.objects_relationship_types
		WHERE object_id = $1`

	var rt models.RelationshipType
	var reverseTypeKey *string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&rt.ObjectID, &rt.TypeKey, &rt.RelationshipName, &reverseTypeKey,
		&rt.Cardinality, &rt.Required, &rt.MinCount, &rt.MaxCount, &rt.ValidationRules,
		&rt.CreatedBy, &rt.UpdatedBy, &rt.CreatedAt, &rt.UpdatedAt,
	)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get relationship type by id: %w", err)
	}

	rt.ReverseTypeKey = reverseTypeKey
	return &rt, nil
}

// GetByTypeKey retrieves a relationship type by type_key
func (r *relationshipTypeRepository) GetByTypeKey(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
	r.metrics.QueryCount++

	query := `
		SELECT object_id, type_key, relationship_name, reverse_type_key,
			cardinality, required, min_count, max_count, validation_rules,
			created_by, updated_by, created_at, updated_at
		FROM objects_service.objects_relationship_types
		WHERE type_key = $1`

	var rt models.RelationshipType
	var reverseTypeKey *string

	err := r.db.QueryRow(ctx, query, typeKey).Scan(
		&rt.ObjectID, &rt.TypeKey, &rt.RelationshipName, &reverseTypeKey,
		&rt.Cardinality, &rt.Required, &rt.MinCount, &rt.MaxCount, &rt.ValidationRules,
		&rt.CreatedBy, &rt.UpdatedBy, &rt.CreatedAt, &rt.UpdatedAt,
	)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get relationship type by type_key: %w", err)
	}

	rt.ReverseTypeKey = reverseTypeKey
	return &rt, nil
}

// Update updates an existing relationship type
func (r *relationshipTypeRepository) Update(ctx context.Context, id int64, input *models.UpdateRelationshipTypeRequest) (*models.RelationshipType, error) {
	r.metrics.QueryCount++

	// Get current relationship type
	current, err := r.GetByID(ctx, id)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, err
	}

	// Build dynamic update query
	updates := []string{}
	args := []interface{}{}
	argNum := 1

	if input.RelationshipName != nil {
		updates = append(updates, fmt.Sprintf("relationship_name = $%d", argNum))
		args = append(args, *input.RelationshipName)
		argNum++
	}

	if input.ReverseTypeKey != nil {
		updates = append(updates, fmt.Sprintf("reverse_type_key = $%d", argNum))
		args = append(args, *input.ReverseTypeKey)
		argNum++
	}

	if input.Cardinality != nil {
		updates = append(updates, fmt.Sprintf("cardinality = $%d", argNum))
		args = append(args, *input.Cardinality)
		argNum++
	}

	if input.Required != nil {
		updates = append(updates, fmt.Sprintf("required = $%d", argNum))
		args = append(args, *input.Required)
		argNum++
	}

	if input.MinCount != nil {
		updates = append(updates, fmt.Sprintf("min_count = $%d", argNum))
		args = append(args, *input.MinCount)
		argNum++
	}

	if input.MaxCount != nil {
		updates = append(updates, fmt.Sprintf("max_count = $%d", argNum))
		args = append(args, *input.MaxCount)
		argNum++
	}

	if input.ValidationRules != nil {
		rulesJSON, err := json.Marshal(input.ValidationRules)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to marshal validation rules: %w", err)
		}
		updates = append(updates, fmt.Sprintf("validation_rules = $%d", argNum))
		args = append(args, rulesJSON)
		argNum++
	}

	// Always update updated_at and updated_by
	updates = append(updates, fmt.Sprintf("updated_at = NOW()"))
	if input.UpdatedBy != "" {
		updates = append(updates, fmt.Sprintf("updated_by = $%d", argNum))
		args = append(args, input.UpdatedBy)
		argNum++
	}

	// Add WHERE clause
	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE objects_service.objects_relationship_types
		SET %s
		WHERE object_id = $%d
		RETURNING object_id, type_key, relationship_name, reverse_type_key,
			cardinality, required, min_count, max_count, validation_rules,
			created_by, updated_by, created_at, updated_at`,
		strings.Join(updates, ", "), argNum)

	var rt models.RelationshipType
	var reverseTypeKey *string

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&rt.ObjectID, &rt.TypeKey, &rt.RelationshipName, &reverseTypeKey,
		&rt.Cardinality, &rt.Required, &rt.MinCount, &rt.MaxCount, &rt.ValidationRules,
		&rt.CreatedBy, &rt.UpdatedBy, &rt.CreatedAt, &rt.UpdatedAt,
	)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to update relationship type: %w", err)
	}

	rt.ReverseTypeKey = reverseTypeKey
	rt.RelationshipName = current.RelationshipName // Keep original if not updated
	if input.RelationshipName != nil {
		rt.RelationshipName = *input.RelationshipName
	}

	return &rt, nil
}

// Delete deletes a relationship type
func (r *relationshipTypeRepository) Delete(ctx context.Context, id int64) error {
	r.metrics.QueryCount++

	// First get the type_key to check if it's in use
	rt, err := r.GetByID(ctx, id)
	if err != nil {
		r.metrics.ErrorCount++
		return err
	}

	// Delete from the CTI table first (will cascade to objects table)
	_, err = r.db.Exec(ctx, `
		DELETE FROM objects_service.objects_relationship_types WHERE object_id = $1
	`, id)
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("failed to delete relationship type: %w", err)
	}

	// Delete the base object (should cascade, but doing explicitly for clarity)
	_, err = r.db.Exec(ctx, `DELETE FROM objects_service.objects WHERE id = $1`, id)
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("failed to delete base object: %w", err)
	}

	_ = rt // Use to avoid unused warning
	return nil
}

// List retrieves relationship types with filtering and pagination
func (r *relationshipTypeRepository) List(ctx context.Context, filter *models.RelationshipTypeFilter) ([]*models.RelationshipType, error) {
	r.metrics.QueryCount++

	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = r.options.DefaultPageSize
	}
	if filter.PageSize > r.options.MaxPageSize {
		filter.PageSize = r.options.MaxPageSize
	}

	// Build query
	whereClauses := []string{}
	args := []interface{}{}
	argNum := 1

	if filter.Cardinality != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("cardinality = $%d", argNum))
		args = append(args, filter.Cardinality)
		argNum++
	}

	if filter.Required != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("required = $%d", argNum))
		args = append(args, *filter.Required)
		argNum++
	}

	// Build WHERE clause
	where := ""
	if len(whereClauses) > 0 {
		where = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Build ORDER BY
	orderBy := "type_key"
	if filter.SortBy != "" {
		orderBy = filter.SortBy
	}
	order := "ASC"
	if filter.SortOrder == "desc" {
		order = "DESC"
	}
	orderClause := fmt.Sprintf("ORDER BY %s %s", orderBy, order)

	// Calculate offset
	offset := (filter.Page - 1) * filter.PageSize

	// Build and execute query
	query := fmt.Sprintf(`
		SELECT object_id, type_key, relationship_name, reverse_type_key,
			cardinality, required, min_count, max_count, validation_rules,
			created_by, updated_by, created_at, updated_at
		FROM objects_service.objects_relationship_types
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, where, orderClause, argNum, argNum+1)

	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to list relationship types: %w", err)
	}
	defer rows.Close()

	var types []*models.RelationshipType
	for rows.Next() {
		var rt models.RelationshipType
		var reverseTypeKey *string

		err := rows.Scan(
			&rt.ObjectID, &rt.TypeKey, &rt.RelationshipName, &reverseTypeKey,
			&rt.Cardinality, &rt.Required, &rt.MinCount, &rt.MaxCount, &rt.ValidationRules,
			&rt.CreatedBy, &rt.UpdatedBy, &rt.CreatedAt, &rt.UpdatedAt,
		)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to scan relationship type: %w", err)
		}
		rt.ReverseTypeKey = reverseTypeKey
		types = append(types, &rt)
	}

	if err := rows.Err(); err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("error iterating relationship types: %w", err)
	}

	return types, nil
}

// Exists checks if a relationship type exists
func (r *relationshipTypeRepository) Exists(ctx context.Context, typeKey string) (bool, error) {
	r.metrics.QueryCount++

	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM objects_service.objects_relationship_types WHERE type_key = $1)
	`, typeKey).Scan(&exists)
	if err != nil {
		r.metrics.ErrorCount++
		return false, fmt.Errorf("failed to check relationship type existence: %w", err)
	}

	return exists, nil
}

// GetByReverseTypeKey retrieves a relationship type by reverse_type_key
func (r *relationshipTypeRepository) GetByReverseTypeKey(ctx context.Context, reverseKey string) (*models.RelationshipType, error) {
	r.metrics.QueryCount++

	query := `
		SELECT object_id, type_key, relationship_name, reverse_type_key,
			cardinality, required, min_count, max_count, validation_rules,
			created_by, updated_by, created_at, updated_at
		FROM objects_service.objects_relationship_types
		WHERE reverse_type_key = $1`

	var rt models.RelationshipType
	var reverseTypeKey *string

	err := r.db.QueryRow(ctx, query, reverseKey).Scan(
		&rt.ObjectID, &rt.TypeKey, &rt.RelationshipName, &reverseTypeKey,
		&rt.Cardinality, &rt.Required, &rt.MinCount, &rt.MaxCount, &rt.ValidationRules,
		&rt.CreatedBy, &rt.UpdatedBy, &rt.CreatedAt, &rt.UpdatedAt,
	)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get relationship type by reverse_type_key: %w", err)
	}

	rt.ReverseTypeKey = reverseTypeKey
	return &rt, nil
}
