package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
)

// ObjectRepository defines operations for object data with advanced querying capabilities
type ObjectRepository interface {
	Repository

	// Basic CRUD operations
	Create(ctx context.Context, input *models.CreateObjectRequest) (*models.Object, error)
	GetByID(ctx context.Context, id int64) (*models.Object, error)
	GetByPublicID(ctx context.Context, publicID uuid.UUID) (*models.Object, error)
	GetByName(ctx context.Context, name string) (*models.Object, error)
	Update(ctx context.Context, id int64, input *models.UpdateObjectRequest) (*models.Object, error)
	Delete(ctx context.Context, id int64) error

	// Advanced querying operations
	List(ctx context.Context, filter *models.ObjectFilter) ([]*models.Object, int64, error)
	Search(ctx context.Context, query string, limit int) ([]*models.Object, error)

	// Metadata and tag operations
	FindByMetadata(ctx context.Context, key, value string) ([]*models.Object, error)
	FindByTags(ctx context.Context, tags []string, matchAll bool) ([]*models.Object, error)
	UpdateMetadata(ctx context.Context, id int64, metadata map[string]interface{}, updatedBy string) error
	AddTags(ctx context.Context, id int64, tags []string, updatedBy string) error
	RemoveTags(ctx context.Context, id int64, tags []string, updatedBy string) error

	// Hierarchical operations (simplified - no eager loading for performance)
	GetChildren(ctx context.Context, parentID int64) ([]*models.Object, error)
	GetDescendants(ctx context.Context, rootID int64, maxDepth *int) ([]*models.Object, error)
	GetAncestors(ctx context.Context, id int64) ([]*models.Object, error)
	GetPath(ctx context.Context, id int64) ([]*models.Object, error)

	// Bulk operations
	BulkCreate(ctx context.Context, objects []*models.CreateObjectRequest) ([]*models.Object, error)
	BulkUpdate(ctx context.Context, ids []int64, updates *models.UpdateObjectRequest) ([]*models.Object, error)
	BulkDelete(ctx context.Context, ids []int64) error

	// Business logic and validation
	ValidateObjectType(ctx context.Context, objectTypeID int64) error
	ValidateParentChild(ctx context.Context, parentID, childID int64) error
	CanDelete(ctx context.Context, id int64) (bool, error)
	GetObjectStats(ctx context.Context, filter *models.ObjectFilter) (*ObjectStats, error)
}

// ObjectStats contains statistics about objects
type ObjectStats struct {
	Total    int64            `json:"total"`
	ByStatus map[string]int64 `json:"by_status"`
	ByType   map[int64]int64  `json:"by_type"`
	ByTags   map[string]int64 `json:"by_tags"`
	Recent   int64            `json:"recent"` // Created in last 30 days
}

// objectRepository implements ObjectRepository with performance focus (no eager loading)
type objectRepository struct {
	db      DBInterface
	options *RepositoryOptions
	metrics *RepositoryMetrics
}

// NewObjectRepository creates a new ObjectRepository instance
func NewObjectRepository(db DBInterface, options *RepositoryOptions) ObjectRepository {
	if options == nil {
		options = DefaultRepositoryOptions()
	}

	return &objectRepository{
		db:      db,
		options: options,
		metrics: &RepositoryMetrics{LastResetAt: time.Now()},
	}
}

// DB implements Repository interface
func (r *objectRepository) DB() DBInterface {
	return r.db
}

// Options implements Repository interface
func (r *objectRepository) Options() *RepositoryOptions {
	return r.options
}

// Metrics implements Repository interface
func (r *objectRepository) Metrics() *RepositoryMetrics {
	return r.metrics
}

// ResetMetrics implements Repository interface
func (r *objectRepository) ResetMetrics() {
	r.metrics.Reset()
}

// Healthy implements Repository interface
func (r *objectRepository) Healthy(ctx context.Context) error {
	var result int
	return r.db.QueryRow(ctx, "SELECT 1").Scan(&result)
}

// Create creates a new object
func (r *objectRepository) Create(ctx context.Context, input *models.CreateObjectRequest) (*models.Object, error) {

	r.metrics.QueryCount++

	// Validate object type exists
	if err := r.ValidateObjectType(ctx, input.ObjectTypeID); err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("invalid object type: %w", err)
	}

	// Validate parent if specified
	if input.ParentObjectID != nil {
		if err := r.validateParent(ctx, *input.ParentObjectID); err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("invalid parent: %w", err)
		}
	}

	query := `
		INSERT INTO objects (
			public_id, object_type_id, parent_object_id, name, description, 
			metadata, tags, status, version, created_by, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING id, created_at, updated_at`

	var object models.Object
	object.PublicID = uuid.New()
	object.ObjectTypeID = input.ObjectTypeID
	object.ParentObjectID = input.ParentObjectID
	object.Name = input.Name
	object.Description = input.Description
	object.Tags = input.Tags
	object.Status = models.StatusActive // Default status
	object.Version = 1
	if input.CreatedBy != "" {
		object.CreatedBy = input.CreatedBy
		object.UpdatedBy = input.CreatedBy
	} else {
		object.CreatedBy = "system"
		object.UpdatedBy = "system"
	}

	// Handle metadata
	if input.Metadata != nil {
		if err := object.SetMetadataMap(input.Metadata); err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to set metadata: %w", err)
		}
	}

	err := r.db.QueryRow(ctx, query,
		object.PublicID, object.ObjectTypeID, object.ParentObjectID,
		object.Name, object.Description, object.Metadata, object.Tags,
		object.Status, object.Version, object.CreatedBy, object.UpdatedBy,
	).Scan(&object.ID, &object.CreatedAt, &object.UpdatedAt)

	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to create object: %w", err)
	}

	// Load ObjectType for eager loading (as planned)
	if objectType, err := r.getObjectTypeByID(ctx, object.ObjectTypeID); err == nil {
		object.ObjectType = objectType
	}

	return &object, nil
}

// GetByID retrieves an object by ID with minimal eager loading
func (r *objectRepository) GetByID(ctx context.Context, id int64) (*models.Object, error) {

	r.metrics.QueryCount++

	query := `
		SELECT id, public_id, object_type_id, parent_object_id, name, description,
			   metadata, tags, status, version, created_by, updated_by,
			   created_at, updated_at, deleted_at
		FROM objects
		WHERE id = $1`

	var object models.Object
	var parentObjectID sql.NullInt64
	var deletedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, id).Scan(
		&object.ID, &object.PublicID, &object.ObjectTypeID, &parentObjectID,
		&object.Name, &object.Description, &object.Metadata, &object.Tags,
		&object.Status, &object.Version, &object.CreatedBy, &object.UpdatedBy,
		&object.CreatedAt, &object.UpdatedAt, &deletedAt,
	)

	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	if parentObjectID.Valid {
		object.ParentObjectID = &parentObjectID.Int64
	}
	if deletedAt.Valid {
		object.DeletedAt = &deletedAt.Time
	}

	// Load ObjectType for eager loading (as planned)
	if objectType, err := r.getObjectTypeByID(ctx, object.ObjectTypeID); err == nil {
		object.ObjectType = objectType
	}

	return &object, nil
}

// GetByPublicID retrieves an object by public ID
func (r *objectRepository) GetByPublicID(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {

	r.metrics.QueryCount++

	query := `SELECT id FROM objects WHERE public_id = $1`
	var id int64
	err := r.db.QueryRow(ctx, query, publicID).Scan(&id)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get object by public ID: %w", err)
	}

	return r.GetByID(ctx, id)
}

// GetByName retrieves an object by name
func (r *objectRepository) GetByName(ctx context.Context, name string) (*models.Object, error) {

	r.metrics.QueryCount++

	query := `
		SELECT id FROM objects 
		WHERE name = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC LIMIT 1`

	var id int64
	err := r.db.QueryRow(ctx, query, name).Scan(&id)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get object by name: %w", err)
	}

	return r.GetByID(ctx, id)
}

// Update updates an existing object with optimistic locking
func (r *objectRepository) Update(ctx context.Context, id int64, input *models.UpdateObjectRequest) (*models.Object, error) {

	r.metrics.QueryCount++

	// Get current object for version checking
	current, err := r.GetByID(ctx, id)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get current object: %w", err)
	}

	// Check version for optimistic locking
	if input.Version != nil && *input.Version != current.Version {
		return nil, ErrOptimisticLock
	}

	// Validate changes
	if input.ObjectTypeID != nil {
		if err := r.ValidateObjectType(ctx, *input.ObjectTypeID); err != nil {
			return nil, fmt.Errorf("invalid object type: %w", err)
		}
	}

	if input.ParentObjectID != nil {
		if *input.ParentObjectID == id {
			return nil, fmt.Errorf("object cannot be its own parent")
		}
		if err := r.validateParent(ctx, *input.ParentObjectID); err != nil {
			return nil, fmt.Errorf("invalid parent: %w", err)
		}
	}

	// Build dynamic update query
	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if input.ObjectTypeID != nil {
		setClauses = append(setClauses, fmt.Sprintf("object_type_id = $%d", argIndex))
		args = append(args, *input.ObjectTypeID)
		argIndex++
	}

	if input.ParentObjectID != nil {
		setClauses = append(setClauses, fmt.Sprintf("parent_object_id = $%d", argIndex))
		args = append(args, *input.ParentObjectID)
		argIndex++
	}

	if input.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *input.Name)
		argIndex++
	}

	if input.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *input.Description)
		argIndex++
	}

	if input.Metadata != nil {
		metadataBytes, err := json.Marshal(*input.Metadata)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		setClauses = append(setClauses, fmt.Sprintf("metadata = $%d", argIndex))
		args = append(args, metadataBytes)
		argIndex++
	}

	if input.Tags != nil {
		setClauses = append(setClauses, fmt.Sprintf("tags = $%d", argIndex))
		args = append(args, *input.Tags)
		argIndex++
	}

	if input.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *input.Status)
		argIndex++
	}

	if len(setClauses) == 0 {
		return current, nil // No changes
	}

	// Add version increment and timestamps
	setClauses = append(setClauses, fmt.Sprintf("version = version + 1"))
	setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")
	setClauses = append(setClauses, fmt.Sprintf("updated_by = $%d", argIndex))
	if input.UpdatedBy != "" {
		args = append(args, input.UpdatedBy)
	} else {
		args = append(args, "system")
	}
	argIndex++

	query := fmt.Sprintf(`
		UPDATE objects 
		SET %s 
		WHERE id = $%d AND deleted_at IS NULL AND version = $%d
		RETURNING updated_at, version`,
		strings.Join(setClauses, ", "),
		argIndex,
		argIndex+1,
	)

	args = append(args, id, current.Version)

	var updatedAt sql.NullTime
	var newVersion int64
	err = r.db.QueryRow(ctx, query, args...).Scan(&updatedAt, &newVersion)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to update object: %w", err)
	}

	// Return updated object
	return r.GetByID(ctx, id)
}

// Delete soft-deletes an object
func (r *objectRepository) Delete(ctx context.Context, id int64) error {

	r.metrics.QueryCount++

	// Check if can delete
	canDelete, err := r.CanDelete(ctx, id)
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("failed to check delete constraints: %w", err)
	}
	if !canDelete {
		return fmt.Errorf("object cannot be deleted")
	}

	query := `
		UPDATE objects 
		SET deleted_at = CURRENT_TIMESTAMP, status = 'deleted', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL`

	_, err = r.db.Exec(ctx, query, id)
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// List retrieves objects with filtering and pagination
func (r *objectRepository) List(ctx context.Context, filter *models.ObjectFilter) ([]*models.Object, int64, error) {

	r.metrics.QueryCount++

	qb := NewQueryBuilder()
	qb.Select(
		"id", "public_id", "object_type_id", "parent_object_id", "name", "description",
		"metadata", "tags", "status", "version", "created_by", "updated_by",
		"created_at", "updated_at", "deleted_at",
	).From("objects")

	// Apply filters
	if filter.Name != "" {
		qb.Where("name ILIKE $1", fmt.Sprintf("%%%s%%", filter.Name))
	}

	if filter.ObjectTypeID != nil {
		qb.Where("object_type_id = $1", *filter.ObjectTypeID)
	}

	if filter.ParentObjectID != nil {
		qb.Where("parent_object_id = $1", *filter.ParentObjectID)
	}

	if filter.Status != "" {
		qb.Where("status = $1", filter.Status)
	}

	if len(filter.Tags) > 0 {
		qb.WhereTagsContain(filter.Tags)
	}

	if filter.CreatedAfter != nil {
		qb.WhereDateRange("created_at", *filter.CreatedAfter, time.Time{})
	}

	if filter.CreatedBefore != nil {
		qb.WhereDateRange("created_at", time.Time{}, *filter.CreatedBefore)
	}

	if filter.MetadataKey != "" && filter.MetadataValue != "" {
		qb.WhereJsonContains("metadata", map[string]interface{}{
			filter.MetadataKey: filter.MetadataValue,
		})
	}

	// Always exclude deleted objects unless explicitly requested
	if filter.Status != models.StatusDeleted {
		qb.Where("deleted_at IS NULL")
	}

	// Get total count first
	countQuery, countArgs := qb.BuildCount()
	var total int64
	err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, 0, fmt.Errorf("failed to count objects: %w", err)
	}

	// Apply sorting
	if filter.SortBy != "" {
		orderDir := "ASC"
		if filter.SortOrder == "desc" {
			orderDir = "DESC"
		}
		qb.OrderBy(fmt.Sprintf("%s %s", filter.SortBy, orderDir))
	} else {
		qb.OrderByDesc("created_at")
	}

	// Apply pagination
	if filter.Limit > 0 {
		qb.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		qb.Offset(filter.Offset)
	}

	query, args := qb.Build()
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, 0, fmt.Errorf("failed to query objects: %w", err)
	}
	defer rows.Close()

	var objects []*models.Object
	for rows.Next() {
		var object models.Object
		var parentObjectID sql.NullInt64
		var deletedAt sql.NullTime

		err := rows.Scan(
			&object.ID, &object.PublicID, &object.ObjectTypeID, &parentObjectID,
			&object.Name, &object.Description, &object.Metadata, &object.Tags,
			&object.Status, &object.Version, &object.CreatedBy, &object.UpdatedBy,
			&object.CreatedAt, &object.UpdatedAt, &deletedAt,
		)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, 0, fmt.Errorf("failed to scan object row: %w", err)
		}

		if parentObjectID.Valid {
			object.ParentObjectID = &parentObjectID.Int64
		}
		if deletedAt.Valid {
			object.DeletedAt = &deletedAt.Time
		}

		objects = append(objects, &object)
	}

	// Load ObjectType for all objects if needed
	if len(objects) > 0 {
		r.loadObjectTypesForObjects(ctx, objects)
	}

	return objects, total, nil
}

// Helper methods

func (r *objectRepository) validateParent(ctx context.Context, parentID int64) error {
	query := `SELECT COUNT(*) FROM objects WHERE id = $1 AND deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query, parentID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to validate parent: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("parent does not exist or is deleted")
	}
	return nil
}

func (r *objectRepository) ValidateObjectType(ctx context.Context, objectTypeID int64) error {
	query := `SELECT COUNT(*) FROM object_types WHERE id = $1`
	var count int64
	err := r.db.QueryRow(ctx, query, objectTypeID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to validate object type: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("object type does not exist")
	}
	return nil
}

func (r *objectRepository) getObjectTypeByID(ctx context.Context, id int64) (*models.ObjectType, error) {
	query := `
		SELECT id, name, parent_type_id, concrete_table_name, description, is_sealed, metadata, created_at, updated_at
		FROM object_types WHERE id = $1`

	var objectType models.ObjectType
	var parentID sql.NullInt64

	err := r.db.QueryRow(ctx, query, id).Scan(
		&objectType.ID, &objectType.Name, &parentID, &objectType.ConcreteTableName,
		&objectType.Description, &objectType.IsSealed, &objectType.Metadata,
		&objectType.CreatedAt, &objectType.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if parentID.Valid {
		objectType.ParentTypeID = &parentID.Int64
	}

	return &objectType, nil
}

func (r *objectRepository) loadObjectTypesForObjects(ctx context.Context, objects []*models.Object) {
	if len(objects) == 0 {
		return
	}

	// Collect unique object type IDs
	typeIDs := make(map[int64]bool)
	for _, obj := range objects {
		typeIDs[obj.ObjectTypeID] = true
	}

	// Batch load object types
	ids := make([]int64, 0, len(typeIDs))
	for id := range typeIDs {
		ids = append(ids, id)
	}

	query := fmt.Sprintf(`
		SELECT id, name, parent_type_id, concrete_table_name, description, is_sealed, metadata, created_at, updated_at
		FROM object_types WHERE id = ANY($1::bigint[])`)

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return // Silently fail - object types will remain nil
	}
	defer rows.Close()

	typeMap := make(map[int64]*models.ObjectType)
	for rows.Next() {
		var objectType models.ObjectType
		var parentID sql.NullInt64

		err := rows.Scan(
			&objectType.ID, &objectType.Name, &parentID, &objectType.ConcreteTableName,
			&objectType.Description, &objectType.IsSealed, &objectType.Metadata,
			&objectType.CreatedAt, &objectType.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if parentID.Valid {
			objectType.ParentTypeID = &parentID.Int64
		}

		typeMap[objectType.ID] = &objectType
	}

	// Assign object types to objects
	for _, obj := range objects {
		if objectType, exists := typeMap[obj.ObjectTypeID]; exists {
			obj.ObjectType = objectType
		}
	}
}

func (r *objectRepository) Search(ctx context.Context, searchQuery string, limit int) ([]*models.Object, error) {
	r.metrics.QueryCount++

	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, public_id, object_type_id, parent_object_id, name, description,
			   metadata, tags, status, version, created_by, updated_by,
			   created_at, updated_at, deleted_at
		FROM objects
		WHERE deleted_at IS NULL
		  AND (name ILIKE $1 OR description ILIKE $1 OR tags @> $2::text[])
		ORDER BY
			CASE WHEN name ILIKE $1 THEN 0 ELSE 1 END,
			created_at DESC
		LIMIT $3`

	pattern := fmt.Sprintf("%%%s%%", searchQuery)
	rows, err := r.db.Query(ctx, query, pattern, []string{searchQuery}, limit)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to search objects: %w", err)
	}
	defer rows.Close()

	var objects []*models.Object
	for rows.Next() {
		var object models.Object
		var parentObjectID sql.NullInt64
		var deletedAt sql.NullTime

		err := rows.Scan(
			&object.ID, &object.PublicID, &object.ObjectTypeID, &parentObjectID,
			&object.Name, &object.Description, &object.Metadata, &object.Tags,
			&object.Status, &object.Version, &object.CreatedBy, &object.UpdatedBy,
			&object.CreatedAt, &object.UpdatedAt, &deletedAt,
		)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to scan object row: %w", err)
		}

		if parentObjectID.Valid {
			object.ParentObjectID = &parentObjectID.Int64
		}
		if deletedAt.Valid {
			object.DeletedAt = &deletedAt.Time
		}

		objects = append(objects, &object)
	}

	if len(objects) > 0 {
		r.loadObjectTypesForObjects(ctx, objects)
	}

	return objects, nil
}

func (r *objectRepository) FindByMetadata(ctx context.Context, key, value string) ([]*models.Object, error) {
	r.metrics.QueryCount++

	query := `
		SELECT id, public_id, object_type_id, parent_object_id, name, description,
			   metadata, tags, status, version, created_by, updated_by,
			   created_at, updated_at, deleted_at
		FROM objects
		WHERE deleted_at IS NULL
		  AND metadata @> $1::jsonb
		ORDER BY created_at DESC`

	metadataFilter := map[string]interface{}{
		key: value,
	}
	metadataJSON, _ := json.Marshal(metadataFilter)

	rows, err := r.db.Query(ctx, query, metadataJSON)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to find by metadata: %w", err)
	}
	defer rows.Close()

	var objects []*models.Object
	for rows.Next() {
		var object models.Object
		var parentObjectID sql.NullInt64
		var deletedAt sql.NullTime

		err := rows.Scan(
			&object.ID, &object.PublicID, &object.ObjectTypeID, &parentObjectID,
			&object.Name, &object.Description, &object.Metadata, &object.Tags,
			&object.Status, &object.Version, &object.CreatedBy, &object.UpdatedBy,
			&object.CreatedAt, &object.UpdatedAt, &deletedAt,
		)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to scan object row: %w", err)
		}

		if parentObjectID.Valid {
			object.ParentObjectID = &parentObjectID.Int64
		}
		if deletedAt.Valid {
			object.DeletedAt = &deletedAt.Time
		}

		objects = append(objects, &object)
	}

	if len(objects) > 0 {
		r.loadObjectTypesForObjects(ctx, objects)
	}

	return objects, nil
}

func (r *objectRepository) FindByTags(ctx context.Context, tags []string, matchAll bool) ([]*models.Object, error) {
	r.metrics.QueryCount++

	var query string
	var args []interface{}

	if matchAll {
		query = `
			SELECT id, public_id, object_type_id, parent_object_id, name, description,
				   metadata, tags, status, version, created_by, updated_by,
				   created_at, updated_at, deleted_at
			FROM objects
			WHERE deleted_at IS NULL
			  AND tags @> $1::text[]
			ORDER BY created_at DESC`
		args = append(args, tags)
	} else {
		query = `
			SELECT id, public_id, object_type_id, parent_object_id, name, description,
				   metadata, tags, status, version, created_by, updated_by,
				   created_at, updated_at, deleted_at
			FROM objects
			WHERE deleted_at IS NULL
			  AND tags && $1::text[]
			ORDER BY created_at DESC`
		args = append(args, tags)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to find by tags: %w", err)
	}
	defer rows.Close()

	var objects []*models.Object
	for rows.Next() {
		var object models.Object
		var parentObjectID sql.NullInt64
		var deletedAt sql.NullTime

		err := rows.Scan(
			&object.ID, &object.PublicID, &object.ObjectTypeID, &parentObjectID,
			&object.Name, &object.Description, &object.Metadata, &object.Tags,
			&object.Status, &object.Version, &object.CreatedBy, &object.UpdatedBy,
			&object.CreatedAt, &object.UpdatedAt, &deletedAt,
		)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to scan object row: %w", err)
		}

		if parentObjectID.Valid {
			object.ParentObjectID = &parentObjectID.Int64
		}
		if deletedAt.Valid {
			object.DeletedAt = &deletedAt.Time
		}

		objects = append(objects, &object)
	}

	if len(objects) > 0 {
		r.loadObjectTypesForObjects(ctx, objects)
	}

	return objects, nil
}

func (r *objectRepository) UpdateMetadata(ctx context.Context, id int64, metadata map[string]interface{}, updatedBy string) error {
	r.metrics.QueryCount++

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	userToSet := updatedBy
	if userToSet == "" {
		userToSet = "system"
	}

	query := `
		UPDATE objects
		SET metadata = $1, updated_at = CURRENT_TIMESTAMP, updated_by = $3
		WHERE id = $2 AND deleted_at IS NULL`

	_, err = r.db.Exec(ctx, query, metadataJSON, id, userToSet)
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	return nil
}

func (r *objectRepository) AddTags(ctx context.Context, id int64, tags []string, updatedBy string) error {
	r.metrics.QueryCount++

	if len(tags) == 0 {
		return nil
	}

	userToSet := updatedBy
	if userToSet == "" {
		userToSet = "system"
	}

	query := `
		UPDATE objects
		SET tags = tags || $1::text[], updated_at = CURRENT_TIMESTAMP, updated_by = $3
		WHERE id = $2 AND deleted_at IS NULL`

	_, err := r.db.Exec(ctx, query, tags, id, userToSet)
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("failed to add tags: %w", err)
	}

	return nil
}

func (r *objectRepository) RemoveTags(ctx context.Context, id int64, tags []string, updatedBy string) error {
	r.metrics.QueryCount++

	if len(tags) == 0 {
		return nil
	}

	userToSet := updatedBy
	if userToSet == "" {
		userToSet = "system"
	}

	query := `
		UPDATE objects
		SET tags = tags - $1::text[], updated_at = CURRENT_TIMESTAMP, updated_by = $3
		WHERE id = $2 AND deleted_at IS NULL`

	_, err := r.db.Exec(ctx, query, tags, id, userToSet)
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("failed to remove tags: %w", err)
	}

	return nil
}

func (r *objectRepository) GetChildren(ctx context.Context, parentID int64) ([]*models.Object, error) {
	r.metrics.QueryCount++

	query := `
		SELECT id, public_id, object_type_id, parent_object_id, name, description,
			   metadata, tags, status, version, created_by, updated_by,
			   created_at, updated_at, deleted_at
		FROM objects
		WHERE parent_object_id = $1 AND deleted_at IS NULL
		ORDER BY name ASC`

	rows, err := r.db.Query(ctx, query, parentID)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get children: %w", err)
	}
	defer rows.Close()

	var objects []*models.Object
	for rows.Next() {
		var object models.Object
		var parentObjectID sql.NullInt64
		var deletedAt sql.NullTime

		err := rows.Scan(
			&object.ID, &object.PublicID, &object.ObjectTypeID, &parentObjectID,
			&object.Name, &object.Description, &object.Metadata, &object.Tags,
			&object.Status, &object.Version, &object.CreatedBy, &object.UpdatedBy,
			&object.CreatedAt, &object.UpdatedAt, &deletedAt,
		)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to scan child row: %w", err)
		}

		if parentObjectID.Valid {
			object.ParentObjectID = &parentObjectID.Int64
		}
		if deletedAt.Valid {
			object.DeletedAt = &deletedAt.Time
		}

		objects = append(objects, &object)
	}

	if len(objects) > 0 {
		r.loadObjectTypesForObjects(ctx, objects)
	}

	return objects, nil
}

func (r *objectRepository) GetDescendants(ctx context.Context, rootID int64, maxDepth *int) ([]*models.Object, error) {
	r.metrics.QueryCount++

	query := `
		WITH RECURSIVE descendants AS (
			SELECT id, public_id, object_type_id, parent_object_id, name, description,
				   metadata, tags, status, version, created_by, updated_by,
				   created_at, updated_at, deleted_at, 1 as depth
			FROM objects WHERE id = $1 AND deleted_at IS NULL

			UNION ALL

			SELECT o.id, o.public_id, o.object_type_id, o.parent_object_id, o.name, o.description,
				   o.metadata, o.tags, o.status, o.version, o.created_by, o.updated_by,
				   o.created_at, o.updated_at, o.deleted_at, d.depth + 1
			FROM objects o
			INNER JOIN descendants d ON o.parent_object_id = d.id
			WHERE o.deleted_at IS NULL
		)
		SELECT id, public_id, object_type_id, parent_object_id, name, description,
			   metadata, tags, status, version, created_by, updated_by,
			   created_at, updated_at, deleted_at
		FROM descendants`

	args := []interface{}{rootID}
	if maxDepth != nil {
		query += " WHERE depth <= $2"
		args = append(args, *maxDepth)
	}

	query += " ORDER BY depth, name"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get descendants: %w", err)
	}
	defer rows.Close()

	var objects []*models.Object
	for rows.Next() {
		var object models.Object
		var parentObjectID sql.NullInt64
		var deletedAt sql.NullTime

		err := rows.Scan(
			&object.ID, &object.PublicID, &object.ObjectTypeID, &parentObjectID,
			&object.Name, &object.Description, &object.Metadata, &object.Tags,
			&object.Status, &object.Version, &object.CreatedBy, &object.UpdatedBy,
			&object.CreatedAt, &object.UpdatedAt, &deletedAt,
		)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to scan descendant row: %w", err)
		}

		if parentObjectID.Valid {
			object.ParentObjectID = &parentObjectID.Int64
		}
		if deletedAt.Valid {
			object.DeletedAt = &deletedAt.Time
		}

		objects = append(objects, &object)
	}

	if len(objects) > 0 {
		r.loadObjectTypesForObjects(ctx, objects)
	}

	return objects, nil
}

func (r *objectRepository) GetAncestors(ctx context.Context, id int64) ([]*models.Object, error) {
	r.metrics.QueryCount++

	query := `
		WITH RECURSIVE ancestors AS (
			SELECT id, public_id, object_type_id, parent_object_id, name, description,
				   metadata, tags, status, version, created_by, updated_by,
				   created_at, updated_at, deleted_at, 1 as level
			FROM objects WHERE id = $1 AND deleted_at IS NULL

			UNION ALL

			SELECT o.id, o.public_id, o.object_type_id, o.parent_object_id, o.name, o.description,
				   o.metadata, o.tags, o.status, o.version, o.created_by, o.updated_by,
				   o.created_at, o.updated_at, o.deleted_at, a.level + 1
			FROM objects o
			INNER JOIN ancestors a ON o.id = a.parent_object_id
			WHERE o.deleted_at IS NULL
		)
		SELECT id, public_id, object_type_id, parent_object_id, name, description,
			   metadata, tags, status, version, created_by, updated_by,
			   created_at, updated_at, deleted_at
		FROM ancestors
		WHERE id != $1
		ORDER BY level DESC`

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get ancestors: %w", err)
	}
	defer rows.Close()

	var objects []*models.Object
	for rows.Next() {
		var object models.Object
		var parentObjectID sql.NullInt64
		var deletedAt sql.NullTime

		err := rows.Scan(
			&object.ID, &object.PublicID, &object.ObjectTypeID, &parentObjectID,
			&object.Name, &object.Description, &object.Metadata, &object.Tags,
			&object.Status, &object.Version, &object.CreatedBy, &object.UpdatedBy,
			&object.CreatedAt, &object.UpdatedAt, &deletedAt,
		)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to scan ancestor row: %w", err)
		}

		if parentObjectID.Valid {
			object.ParentObjectID = &parentObjectID.Int64
		}
		if deletedAt.Valid {
			object.DeletedAt = &deletedAt.Time
		}

		objects = append(objects, &object)
	}

	if len(objects) > 0 {
		r.loadObjectTypesForObjects(ctx, objects)
	}

	return objects, nil
}

func (r *objectRepository) GetPath(ctx context.Context, id int64) ([]*models.Object, error) {
	r.metrics.QueryCount++

	query := `
		WITH RECURSIVE path AS (
			SELECT id, public_id, object_type_id, parent_object_id, name, description,
				   metadata, tags, status, version, created_by, updated_by,
				   created_at, updated_at, deleted_at, 1 as level
			FROM objects WHERE id = $1 AND deleted_at IS NULL

			UNION ALL

			SELECT o.id, o.public_id, o.object_type_id, o.parent_object_id, o.name, o.description,
				   o.metadata, o.tags, o.status, o.version, o.created_by, o.updated_by,
				   o.created_at, o.updated_at, o.deleted_at, p.level + 1
			FROM objects o
			INNER JOIN path p ON o.id = p.parent_object_id
			WHERE o.deleted_at IS NULL
		)
		SELECT id, public_id, object_type_id, parent_object_id, name, description,
			   metadata, tags, status, version, created_by, updated_by,
			   created_at, updated_at, deleted_at
		FROM path
		ORDER BY level DESC`

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get path: %w", err)
	}
	defer rows.Close()

	var objects []*models.Object
	for rows.Next() {
		var object models.Object
		var parentObjectID sql.NullInt64
		var deletedAt sql.NullTime

		err := rows.Scan(
			&object.ID, &object.PublicID, &object.ObjectTypeID, &parentObjectID,
			&object.Name, &object.Description, &object.Metadata, &object.Tags,
			&object.Status, &object.Version, &object.CreatedBy, &object.UpdatedBy,
			&object.CreatedAt, &object.UpdatedAt, &deletedAt,
		)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to scan path row: %w", err)
		}

		if parentObjectID.Valid {
			object.ParentObjectID = &parentObjectID.Int64
		}
		if deletedAt.Valid {
			object.DeletedAt = &deletedAt.Time
		}

		objects = append(objects, &object)
	}

	if len(objects) > 0 {
		r.loadObjectTypesForObjects(ctx, objects)
	}

	return objects, nil
}

func (r *objectRepository) BulkCreate(ctx context.Context, inputs []*models.CreateObjectRequest) ([]*models.Object, error) {
	r.metrics.QueryCount++

	if len(inputs) == 0 {
		return []*models.Object{}, nil
	}

	query := `
		INSERT INTO objects (
			public_id, object_type_id, parent_object_id, name, description,
			metadata, tags, status, version, created_by, updated_by
		) VALUES`

	valueStrings := []string{}
	valueArgs := []interface{}{}
	for i, input := range inputs {
		offset := i * 11
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7, offset+8, offset+9, offset+10, offset+11))

		publicID := uuid.New()
		tags := input.Tags
		status := models.StatusActive
		version := 1
		createdBy := "system"
		updatedBy := "system"

		valueArgs = append(valueArgs,
			publicID, input.ObjectTypeID, input.ParentObjectID, input.Name, input.Description,
			input.Metadata, tags, status, version, createdBy, updatedBy)
	}

	query += strings.Join(valueStrings, ", ")
	query += ` RETURNING id, public_id, object_type_id, parent_object_id, name, description,
					metadata, tags, status, version, created_by, updated_by, created_at, updated_at`

	rows, err := r.db.Query(ctx, query, valueArgs...)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to bulk create objects: %w", err)
	}
	defer rows.Close()

	var objects []*models.Object
	for rows.Next() {
		var object models.Object
		var parentObjectID sql.NullInt64

		err := rows.Scan(
			&object.ID, &object.PublicID, &object.ObjectTypeID, &parentObjectID,
			&object.Name, &object.Description, &object.Metadata, &object.Tags,
			&object.Status, &object.Version, &object.CreatedBy, &object.UpdatedBy,
			&object.CreatedAt, &object.UpdatedAt,
		)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to scan bulk create result: %w", err)
		}

		if parentObjectID.Valid {
			object.ParentObjectID = &parentObjectID.Int64
		}

		objects = append(objects, &object)
	}

	if len(objects) > 0 {
		r.loadObjectTypesForObjects(ctx, objects)
	}

	return objects, nil
}

func (r *objectRepository) BulkUpdate(ctx context.Context, ids []int64, updates *models.UpdateObjectRequest) ([]*models.Object, error) {
	r.metrics.QueryCount++

	if len(ids) == 0 {
		return []*models.Object{}, nil
	}

	if updates == nil {
		return nil, fmt.Errorf("updates cannot be nil")
	}

	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if updates.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *updates.Name)
		argIndex++
	}

	if updates.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *updates.Description)
		argIndex++
	}

	if updates.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *updates.Status)
		argIndex++
	}

	if updates.Metadata != nil {
		metadataBytes, err := json.Marshal(*updates.Metadata)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		setClauses = append(setClauses, fmt.Sprintf("metadata = $%d", argIndex))
		args = append(args, metadataBytes)
		argIndex++
	}

	if updates.Tags != nil {
		setClauses = append(setClauses, fmt.Sprintf("tags = $%d", argIndex))
		args = append(args, *updates.Tags)
		argIndex++
	}

	if len(setClauses) == 0 {
		return []*models.Object{}, nil
	}

	setClauses = append(setClauses, fmt.Sprintf("version = version + 1"))
	setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")
	setClauses = append(setClauses, fmt.Sprintf("updated_by = $%d", argIndex))
	args = append(args, "system")
	argIndex++

	idArray := fmt.Sprintf("$%d", argIndex)
	args = append(args, ids)

	query := fmt.Sprintf(`
		UPDATE objects
		SET %s
		WHERE id = ANY(%s::bigint[]) AND deleted_at IS NULL
		RETURNING id, public_id, object_type_id, parent_object_id, name, description,
				  metadata, tags, status, version, created_by, updated_by,
				  created_at, updated_at, deleted_at`,
		strings.Join(setClauses, ", "),
		idArray,
	)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to bulk update objects: %w", err)
	}
	defer rows.Close()

	var objects []*models.Object
	for rows.Next() {
		var object models.Object
		var parentObjectID sql.NullInt64
		var deletedAt sql.NullTime

		err := rows.Scan(
			&object.ID, &object.PublicID, &object.ObjectTypeID, &parentObjectID,
			&object.Name, &object.Description, &object.Metadata, &object.Tags,
			&object.Status, &object.Version, &object.CreatedBy, &object.UpdatedBy,
			&object.CreatedAt, &object.UpdatedAt, &deletedAt,
		)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to scan bulk update result: %w", err)
		}

		if parentObjectID.Valid {
			object.ParentObjectID = &parentObjectID.Int64
		}
		if deletedAt.Valid {
			object.DeletedAt = &deletedAt.Time
		}

		objects = append(objects, &object)
	}

	if len(objects) > 0 {
		r.loadObjectTypesForObjects(ctx, objects)
	}

	return objects, nil
}

func (r *objectRepository) BulkDelete(ctx context.Context, ids []int64) error {
	r.metrics.QueryCount++

	if len(ids) == 0 {
		return nil
	}

	idArray := fmt.Sprintf("$1::bigint[]")
	query := fmt.Sprintf(`
		UPDATE objects
		SET deleted_at = CURRENT_TIMESTAMP, status = 'deleted', updated_at = CURRENT_TIMESTAMP
		WHERE id = ANY(%s) AND deleted_at IS NULL`, idArray)

	_, err := r.db.Exec(ctx, query, ids)
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("failed to bulk delete objects: %w", err)
	}

	return nil
}

func (r *objectRepository) ValidateParentChild(ctx context.Context, parentID, childID int64) error {
	r.metrics.QueryCount++

	if parentID == childID {
		return fmt.Errorf("object cannot be its own parent")
	}

	query := `
		WITH RECURSIVE descendants AS (
			SELECT id, parent_object_id FROM objects WHERE id = $1 AND deleted_at IS NULL
			UNION ALL
			SELECT o.id, o.parent_object_id
			FROM objects o
			INNER JOIN descendants d ON o.parent_object_id = d.id
			WHERE o.deleted_at IS NULL
		)
		SELECT COUNT(*) FROM descendants WHERE id = $2`

	var count int64
	err := r.db.QueryRow(ctx, query, parentID, childID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to validate parent-child: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("circular dependency detected")
	}

	return nil
}

func (r *objectRepository) CanDelete(ctx context.Context, id int64) (bool, error) {
	// Check if has children
	query := `SELECT COUNT(*) FROM objects WHERE parent_object_id = $1 AND deleted_at IS NULL`
	var childCount int64
	err := r.db.QueryRow(ctx, query, id).Scan(&childCount)
	if err != nil {
		return false, fmt.Errorf("failed to check children: %w", err)
	}

	if childCount > 0 {
		return false, nil
	}

	return true, nil
}

func (r *objectRepository) GetObjectStats(ctx context.Context, filter *models.ObjectFilter) (*ObjectStats, error) {
	r.metrics.QueryCount++

	stats := &ObjectStats{
		ByStatus: make(map[string]int64),
		ByType:   make(map[int64]int64),
		ByTags:   make(map[string]int64),
	}

	baseQuery := "FROM objects"
	whereClauses := []string{}
	if filter == nil || filter.Status != models.StatusDeleted {
		whereClauses = append(whereClauses, "deleted_at IS NULL")
	}

	if filter != nil && filter.ObjectTypeID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("object_type_id = $1"))
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	totalQuery := fmt.Sprintf("SELECT COUNT(*) %s %s", baseQuery, whereSQL)
	args := []interface{}{}
	if filter != nil && filter.ObjectTypeID != nil {
		args = append(args, *filter.ObjectTypeID)
	}

	var total int64
	err := r.db.QueryRow(ctx, totalQuery, args...).Scan(&total)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to count total objects: %w", err)
	}
	stats.Total = total

	statusQuery := fmt.Sprintf(`
		SELECT status, COUNT(*)
		FROM objects
		WHERE deleted_at IS NULL
		GROUP BY status`)

	rows, err := r.db.Query(ctx, statusQuery)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get status stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}
		stats.ByStatus[status] = count
	}

	typeQuery := fmt.Sprintf(`
		SELECT object_type_id, COUNT(*)
		FROM objects
		WHERE deleted_at IS NULL
		GROUP BY object_type_id`)

	rows, err = r.db.Query(ctx, typeQuery)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get type stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var typeID int64
		var count int64
		if err := rows.Scan(&typeID, &count); err != nil {
			continue
		}
		stats.ByType[typeID] = count
	}

	recentQuery := `
		SELECT COUNT(*)
		FROM objects
		WHERE deleted_at IS NULL
		  AND created_at > CURRENT_TIMESTAMP - INTERVAL '30 days'`

	var recent int64
	err = r.db.QueryRow(ctx, recentQuery).Scan(&recent)
	if err == nil {
		stats.Recent = recent
	}

	return stats, nil
}
