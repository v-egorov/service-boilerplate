package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
)

// TODO: OpenTelemetry has been removed - need to return it back later on
//
// ObjectTypeRepository defines operations for object type hierarchical data
type ObjectTypeRepository interface {
	Repository

	// Basic CRUD operations
	Create(ctx context.Context, input *models.CreateObjectTypeRequest) (*models.ObjectType, error)
	GetByID(ctx context.Context, id int64) (*models.ObjectType, error)
	GetByName(ctx context.Context, name string) (*models.ObjectType, error)
	Update(ctx context.Context, id int64, input *models.UpdateObjectTypeRequest) (*models.ObjectType, error)
	Delete(ctx context.Context, id int64) error

	// Hierarchical operations with eager loading
	GetTree(ctx context.Context, rootID *int64) ([]*models.ObjectType, error)
	GetChildren(ctx context.Context, parentID int64) ([]*models.ObjectType, error)
	GetDescendants(ctx context.Context, rootID int64, maxDepth *int) ([]*models.ObjectType, error)
	GetAncestors(ctx context.Context, id int64) ([]*models.ObjectType, error)
	GetPath(ctx context.Context, id int64) ([]*models.ObjectType, error)

	// Query operations
	List(ctx context.Context, filter *models.ObjectTypeFilter) ([]*models.ObjectType, error)
	Search(ctx context.Context, query string, limit int) ([]*models.ObjectType, error)

	// Validation and business logic
	ValidateParentChild(ctx context.Context, parentID, childID int64) error
	ValidateMove(ctx context.Context, id int64, newParentID *int64) error
	CanDelete(ctx context.Context, id int64) (bool, error)
	GetSubtreeObjectCount(ctx context.Context, id int64) (int64, error)
}

// objectTypeRepository implements ObjectTypeRepository
type objectTypeRepository struct {
	db      DBInterface
	options *RepositoryOptions
	metrics *RepositoryMetrics
}

// NewObjectTypeRepository creates a new ObjectTypeRepository instance
func NewObjectTypeRepository(db DBInterface, options *RepositoryOptions) ObjectTypeRepository {
	if options == nil {
		options = DefaultRepositoryOptions()
	}

	return &objectTypeRepository{
		db:      db,
		options: options,
		metrics: &RepositoryMetrics{LastResetAt: time.Now()},
	}
}

// DB implements Repository interface
func (r *objectTypeRepository) DB() DBInterface {
	return r.db
}

// Options implements Repository interface
func (r *objectTypeRepository) Options() *RepositoryOptions {
	return r.options
}

// Metrics implements Repository interface
func (r *objectTypeRepository) Metrics() *RepositoryMetrics {
	return r.metrics
}

// ResetMetrics implements Repository interface
func (r *objectTypeRepository) ResetMetrics() {
	r.metrics.Reset()
}

// Healthy implements Repository interface
func (r *objectTypeRepository) Healthy(ctx context.Context) error {
	return nil // TODO: Implement health check
}

// Create creates a new object type
func (r *objectTypeRepository) Create(ctx context.Context, input *models.CreateObjectTypeRequest) (*models.ObjectType, error) {
	r.metrics.QueryCount++

	// Validate parent if specified
	if input.ParentTypeID != nil {
		if err := r.validateParentExists(ctx, *input.ParentTypeID); err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("invalid parent: %w", err)
		}
	}

	query := `
		INSERT INTO object_types (
			name, parent_type_id, concrete_table_name, description, is_sealed, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING id, created_at, updated_at`

	var objectType models.ObjectType
	objectType.Name = input.Name
	objectType.ParentTypeID = input.ParentTypeID
	objectType.ConcreteTableName = input.ConcreteTableName
	objectType.Description = input.Description

	// Set defaults
	isSealed := false
	if input.IsSealed != nil {
		isSealed = *input.IsSealed
	}
	objectType.IsSealed = isSealed

	// Handle metadata
	if input.Metadata != nil {
		if err := objectType.SetMetadataMap(input.Metadata); err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to set metadata: %w", err)
		}
	}

	err := r.db.QueryRow(ctx, query,
		objectType.Name, objectType.ParentTypeID, objectType.ConcreteTableName,
		objectType.Description, objectType.IsSealed, objectType.Metadata,
	).Scan(&objectType.ID, &objectType.CreatedAt, &objectType.UpdatedAt)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to create object type: %w", err)
	}

	// Load eager-loaded relationships
	objectType.Children = []models.ObjectType{}
	objectType.ParentType = nil
	if objectType.ParentTypeID != nil {
		parent, _ := r.GetByID(ctx, *objectType.ParentTypeID)
		objectType.ParentType = parent
	}

	return &objectType, nil
}

// GetByID retrieves an object type by ID with eager loading
func (r *objectTypeRepository) GetByID(ctx context.Context, id int64) (*models.ObjectType, error) {
	r.metrics.QueryCount++

	query := `
		SELECT id, name, parent_type_id, concrete_table_name, description, is_sealed, metadata, created_at, updated_at
		FROM object_types
		WHERE id = $1`

	var objectType models.ObjectType
	var parentID sql.NullInt64

	err := r.db.QueryRow(ctx, query, id).Scan(
		&objectType.ID, &objectType.Name, &parentID, &objectType.ConcreteTableName,
		&objectType.Description, &objectType.IsSealed, &objectType.Metadata,
		&objectType.CreatedAt, &objectType.UpdatedAt,
	)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get object type: %w", err)
	}

	if parentID.Valid {
		objectType.ParentTypeID = &parentID.Int64
	}

	// Eager load children
	children, err := r.GetChildren(ctx, objectType.ID)
	if err == nil {
		objectType.Children = make([]models.ObjectType, len(children))
		for i, child := range children {
			objectType.Children[i] = *child
		}
	} else {
		objectType.Children = []models.ObjectType{}
	}

	// Eager load parent
	if objectType.ParentTypeID != nil {
		parent, err := r.GetByID(ctx, *objectType.ParentTypeID)
		if err == nil {
			objectType.ParentType = parent
		}
	}

	return &objectType, nil
}

// GetByName retrieves an object type by name with eager loading
func (r *objectTypeRepository) GetByName(ctx context.Context, name string) (*models.ObjectType, error) {
	r.metrics.QueryCount++

	query := `
		SELECT id, name, parent_type_id, concrete_table_name, description, is_sealed, metadata, created_at, updated_at
		FROM object_types 
		WHERE name = $1`

	var objectType models.ObjectType
	var parentID sql.NullInt64

	err := r.db.QueryRow(ctx, query, name).Scan(
		&objectType.ID, &objectType.Name, &parentID, &objectType.ConcreteTableName,
		&objectType.Description, &objectType.IsSealed, &objectType.Metadata,
		&objectType.CreatedAt, &objectType.UpdatedAt,
	)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get object type by name: %w", err)
	}

	if parentID.Valid {
		objectType.ParentTypeID = &parentID.Int64
	}

	// Load relationships
	objectType.Children = []models.ObjectType{}
	if objectType.ParentTypeID != nil {
		parent, _ := r.GetByID(ctx, *objectType.ParentTypeID)
		objectType.ParentType = parent
	}

	return &objectType, nil
}

// Update updates an existing object type
func (r *objectTypeRepository) Update(ctx context.Context, id int64, input *models.UpdateObjectTypeRequest) (*models.ObjectType, error) {
	r.metrics.QueryCount++

	// Get current object type for versioning and validation
	current, err := r.GetByID(ctx, id)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get current object type: %w", err)
	}

	// Validate parent change if specified
	if input.ParentTypeID != nil {
		if *input.ParentTypeID == id {
			return nil, fmt.Errorf("object type cannot be its own parent")
		}
		if err := r.ValidateParentChild(ctx, *input.ParentTypeID, id); err != nil {
			return nil, fmt.Errorf("invalid parent relationship: %w", err)
		}
	}

	// Build dynamic update query
	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if input.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *input.Name)
		argIndex++
	}

	if input.ParentTypeID != nil {
		setClauses = append(setClauses, fmt.Sprintf("parent_type_id = $%d", argIndex))
		args = append(args, *input.ParentTypeID)
		argIndex++
	}

	if input.ConcreteTableName != nil {
		setClauses = append(setClauses, fmt.Sprintf("concrete_table_name = $%d", argIndex))
		args = append(args, *input.ConcreteTableName)
		argIndex++
	}

	if input.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *input.Description)
		argIndex++
	}

	if input.IsSealed != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_sealed = $%d", argIndex))
		args = append(args, *input.IsSealed)
		argIndex++
	}

	if input.Metadata != nil {
		// Convert metadata to JSON
		metadataBytes, err := json.Marshal(*input.Metadata)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		setClauses = append(setClauses, fmt.Sprintf("metadata = $%d", argIndex))
		args = append(args, metadataBytes)
		argIndex++
	}

	if len(setClauses) == 0 {
		return current, nil // No changes
	}

	// Add updated_at
	setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")

	query := fmt.Sprintf(`
		UPDATE object_types 
		SET %s 
		WHERE id = $%d
		RETURNING updated_at`,
		strings.Join(setClauses, ", "),
		argIndex,
	)

	args = append(args, id)

	var updatedAt sql.NullTime
	err = r.db.QueryRow(ctx, query, args...).Scan(&updatedAt)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to update object type: %w", err)
	}

	// Return updated object with eager loading
	return r.GetByID(ctx, id)
}

// Delete soft-deletes an object type
func (r *objectTypeRepository) Delete(ctx context.Context, id int64) error {
	r.metrics.QueryCount++

	// Check if can delete
	canDelete, err := r.CanDelete(ctx, id)
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("failed to check delete constraints: %w", err)
	}
	if !canDelete {
		return fmt.Errorf("object type cannot be deleted")
	}

	// For simplicity, we'll do a hard delete since the current schema doesn't have soft delete
	query := `DELETE FROM object_types WHERE id = $1`
	_, err = r.db.Exec(ctx, query, id)
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("failed to delete object type: %w", err)
	}

	return nil
}

// GetTree retrieves the complete tree starting from root (or all roots if rootID is nil)
func (r *objectTypeRepository) GetTree(ctx context.Context, rootID *int64) ([]*models.ObjectType, error) {
	r.metrics.QueryCount++

	// Use recursive CTE for efficient tree fetching
	query := `
		WITH RECURSIVE object_tree AS (
			-- Base case: root nodes
			SELECT 
				id, name, parent_type_id, concrete_table_name, description, is_sealed, metadata,
				created_at, updated_at
			FROM object_types 
			WHERE ($1::bigint IS NULL AND parent_type_id IS NULL) OR id = $1::bigint
			
			UNION ALL
			
			-- Recursive case: children
			SELECT 
				ot.id, ot.name, ot.parent_type_id, ot.concrete_table_name, ot.description, ot.is_sealed, ot.metadata,
				ot.created_at, ot.updated_at
			FROM object_types ot
			INNER JOIN object_tree t ON ot.parent_type_id = t.id
		)
		SELECT * FROM object_tree ORDER BY name;`

	rows, err := r.db.Query(ctx, query, rootID)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get object tree: %w", err)
	}
	defer rows.Close()

	var objectTypes []*models.ObjectType
	idMap := make(map[int64]*models.ObjectType)

	for rows.Next() {
		var objectType models.ObjectType
		var parentID sql.NullInt64

		err := rows.Scan(
			&objectType.ID, &objectType.Name, &parentID, &objectType.ConcreteTableName,
			&objectType.Description, &objectType.IsSealed, &objectType.Metadata,
			&objectType.CreatedAt, &objectType.UpdatedAt,
		)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to scan object type row: %w", err)
		}

		if parentID.Valid {
			objectType.ParentTypeID = &parentID.Int64
		}

		objectType.Children = []models.ObjectType{}
		objectTypes = append(objectTypes, &objectType)
		idMap[objectType.ID] = &objectType
	}

	// Build hierarchical structure
	var roots []*models.ObjectType
	for _, obj := range objectTypes {
		if obj.ParentTypeID != nil {
			if parent, exists := idMap[*obj.ParentTypeID]; exists {
				parent.Children = append(parent.Children, *obj)
				obj.ParentType = parent
			}
		} else {
			roots = append(roots, obj)
		}
	}

	return roots, nil
}

// GetChildren retrieves direct children of an object type
func (r *objectTypeRepository) GetChildren(ctx context.Context, parentID int64) ([]*models.ObjectType, error) {
	r.metrics.QueryCount++

	query := `
		SELECT id, name, parent_type_id, concrete_table_name, description, is_sealed, metadata, created_at, updated_at
		FROM object_types
		WHERE parent_type_id = $1
		ORDER BY name ASC`

	rows, err := r.db.Query(ctx, query, parentID)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get children: %w", err)
	}
	defer rows.Close()

	var children []*models.ObjectType
	for rows.Next() {
		var objectType models.ObjectType
		var parentID sql.NullInt64

		err := rows.Scan(
			&objectType.ID, &objectType.Name, &parentID, &objectType.ConcreteTableName,
			&objectType.Description, &objectType.IsSealed, &objectType.Metadata,
			&objectType.CreatedAt, &objectType.UpdatedAt,
		)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to scan child row: %w", err)
		}

		if parentID.Valid {
			objectType.ParentTypeID = &parentID.Int64
		}

		objectType.Children = []models.ObjectType{} // Initialize empty children for eager loading
		children = append(children, &objectType)
	}

	return children, nil
}

// Helper methods

func (r *objectTypeRepository) validateParentExists(ctx context.Context, parentID int64) error {
	query := `SELECT COUNT(*) FROM object_types WHERE id = $1`
	var count int64
	err := r.db.QueryRow(ctx, query, parentID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to validate parent: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("parent does not exist")
	}
	return nil
}

func (r *objectTypeRepository) ValidateParentChild(ctx context.Context, parentID, childID int64) error {
	// Check if parent would create a circular dependency
	query := `
		WITH RECURSIVE descendants AS (
			SELECT id, parent_type_id FROM object_types WHERE id = $1
			UNION ALL
			SELECT ot.id, ot.parent_type_id 
			FROM object_types ot
			INNER JOIN descendants d ON ot.id = d.parent_type_id
		)
		SELECT COUNT(*) FROM descendants WHERE id = $2`

	var count int64
	err := r.db.QueryRow(ctx, query, parentID, childID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to validate parent-child relationship: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("circular dependency detected")
	}

	return nil
}

func (r *objectTypeRepository) CanDelete(ctx context.Context, id int64) (bool, error) {
	// Check if has children
	query := `SELECT COUNT(*) FROM object_types WHERE parent_type_id = $1`
	var childCount int64
	err := r.db.QueryRow(ctx, query, id).Scan(&childCount)
	if err != nil {
		return false, fmt.Errorf("failed to check children: %w", err)
	}

	if childCount > 0 {
		return false, nil
	}

	// Check if has objects (assuming objects table exists)
	query = `SELECT COUNT(*) FROM objects WHERE object_type_id = $1 AND deleted_at IS NULL`
	var objectCount int64
	err = r.db.QueryRow(ctx, query, id).Scan(&objectCount)
	if err != nil {
		// If objects table doesn't exist, assume no objects
		return true, nil
	}

	return objectCount == 0, nil
}

// Placeholder methods to be implemented

// TODO: Implement GetDescendants - hierarchical query with maxDepth parameter using recursive CTE
func (r *objectTypeRepository) GetDescendants(ctx context.Context, rootID int64, maxDepth *int) ([]*models.ObjectType, error) {
	return nil, fmt.Errorf("GetDescendants not implemented yet")
}

// TODO: Implement GetAncestors - hierarchical query moving up the tree
func (r *objectTypeRepository) GetAncestors(ctx context.Context, id int64) ([]*models.ObjectType, error) {
	return nil, fmt.Errorf("GetAncestors not implemented yet")
}

// TODO: Implement GetPath - hierarchical query getting full path from root to node
func (r *objectTypeRepository) GetPath(ctx context.Context, id int64) ([]*models.ObjectType, error) {
	return nil, fmt.Errorf("GetPath not implemented yet")
}

// TODO: Implement List - filtered listing with pagination
func (r *objectTypeRepository) List(ctx context.Context, filter *models.ObjectTypeFilter) ([]*models.ObjectType, error) {
	return nil, fmt.Errorf("List not implemented yet")
}

// TODO: Implement Search - text search across object types
func (r *objectTypeRepository) Search(ctx context.Context, query string, limit int) ([]*models.ObjectType, error) {
	return nil, fmt.Errorf("Search not implemented yet")
}

// TODO: Implement ValidateMove - validate moving an object type to a new parent
func (r *objectTypeRepository) ValidateMove(ctx context.Context, id int64, newParentID *int64) error {
	return fmt.Errorf("ValidateMove not implemented yet")
}

// TODO: Implement GetSubtreeObjectCount - count all objects in a subtree
func (r *objectTypeRepository) GetSubtreeObjectCount(ctx context.Context, id int64) (int64, error) {
	return 0, fmt.Errorf("GetSubtreeObjectCount not implemented yet")
}
