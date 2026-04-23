package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
)

type RelationshipRepository interface {
	Repository

	Create(ctx context.Context, input *models.CreateRelationshipRequest) (*models.Relationship, error)
	GetByObjectID(ctx context.Context, objectID int64) (*models.Relationship, error)
	GetByPublicID(ctx context.Context, publicID uuid.UUID) (*models.Relationship, error)
	Update(ctx context.Context, objectID int64, input *models.UpdateRelationshipRequest) (*models.Relationship, error)
	Delete(ctx context.Context, objectID int64) error
	DeleteByPublicID(ctx context.Context, publicID uuid.UUID) error

	List(ctx context.Context, filter *models.RelationshipFilter) ([]*models.Relationship, error)
	GetForObject(ctx context.Context, objectPublicID uuid.UUID, filter *models.RelationshipFilterForType) ([]*models.Relationship, error)
	GetForObjectByType(ctx context.Context, objectPublicID uuid.UUID, typeKey string) ([]*models.Relationship, error)
	GetRelatedObjects(ctx context.Context, objectPublicID uuid.UUID, typeKey *string) ([]*models.Object, error)

	Exists(ctx context.Context, sourceObjectID, targetObjectID int64, typeObjectID int64) (bool, error)
	CountForObject(ctx context.Context, objectID int64, typeKey *string) (int, error)
	GetByTypeKey(ctx context.Context, typeKey string) ([]*models.Relationship, error)

	CheckCircular(ctx context.Context, sourceObjectID, targetObjectID, typeObjectID int64) (bool, error)
}

type relationshipRepository struct {
	db         DBInterface
	options    *RepositoryOptions
	metrics    *RepositoryMetrics
	objectRepo ObjectRepository
}

func NewRelationshipRepository(db DBInterface, options *RepositoryOptions, objectRepo ObjectRepository) RelationshipRepository {
	if options == nil {
		options = DefaultRepositoryOptions()
	}

	return &relationshipRepository{
		db:         db,
		options:    options,
		metrics:    &RepositoryMetrics{LastResetAt: time.Now()},
		objectRepo: objectRepo,
	}
}

func (r *relationshipRepository) DB() DBInterface {
	return r.db
}

func (r *relationshipRepository) Options() *RepositoryOptions {
	return r.options
}

func (r *relationshipRepository) Metrics() *RepositoryMetrics {
	return r.metrics
}

func (r *relationshipRepository) ResetMetrics() {
	r.metrics.Reset()
}

func (r *relationshipRepository) Healthy(ctx context.Context) error {
	var result int
	return r.db.QueryRow(ctx, "SELECT 1").Scan(&result)
}

func (r *relationshipRepository) Create(ctx context.Context, input *models.CreateRelationshipRequest) (*models.Relationship, error) {
	r.metrics.QueryCount++

	input.SetDefaults()

	if err := input.Validate(); err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	sourcePublicID, err := uuid.Parse(input.SourceObjectPublicID)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("invalid source_object_id format: %w", err)
	}

	targetPublicID, err := uuid.Parse(input.TargetObjectPublicID)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("invalid target_object_id format: %w", err)
	}

	sourceObject, err := r.objectRepo.GetByPublicID(ctx, sourcePublicID)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("source object not found: %w", err)
	}

	targetObject, err := r.objectRepo.GetByPublicID(ctx, targetPublicID)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("target object not found: %w", err)
	}

	relTypeRepo := NewRelationshipTypeRepository(r.db, r.options)
	relType, err := relTypeRepo.GetByTypeKey(ctx, input.RelationshipTypeKey)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("relationship type not found: %w", err)
	}

	objectID, err := r.createBaseObject(ctx, input.CreatedBy)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to create base object: %w", err)
	}

	var metadata []byte
	if input.RelationshipMetadata != nil {
		metadata, err = json.Marshal(input.RelationshipMetadata)
		if err != nil {
			r.metrics.ErrorCount++
			_ = r.deleteBaseObject(ctx, objectID)
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
	} else {
		metadata = []byte("{}")
	}

	status := input.Status
	if status == "" {
		status = models.StatusActive
	}

	query := `
		INSERT INTO objects_service.objects_relationships (
			object_id, source_object_id, target_object_id, relationship_type_id,
			status, relationship_metadata, created_by, updated_by,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING object_id, source_object_id, target_object_id, relationship_type_id,
			status, relationship_metadata, created_by, updated_by, created_at, updated_at`

	var rel models.Relationship
	var relMeta []byte

	err = r.db.QueryRow(ctx, query,
		objectID, sourceObject.ID, targetObject.ID, relType.ObjectID,
		status, metadata, input.CreatedBy, input.CreatedBy,
	).Scan(
		&rel.ObjectID, &rel.SourceObjectID, &rel.TargetObjectID, &rel.RelationshipTypeID,
		&rel.Status, &relMeta, &rel.CreatedBy, &rel.UpdatedBy, &rel.CreatedAt, &rel.UpdatedAt,
	)
	if err != nil {
		r.metrics.ErrorCount++
		_ = r.deleteBaseObject(ctx, objectID)
		return nil, fmt.Errorf("failed to create relationship: %w", err)
	}

	rel.RelationshipMetadata = relMeta
	rel.SourceObjectPublicID = sourceObject.PublicID
	rel.TargetObjectPublicID = targetObject.PublicID
	rel.RelationshipTypeKey = relType.TypeKey

	return &rel, nil
}

func (r *relationshipRepository) createBaseObject(ctx context.Context, createdBy string) (int64, error) {
	var typeID int64
	err := r.db.QueryRow(ctx, `
		SELECT id FROM objects_service.object_types WHERE name = 'Relationship'
	`).Scan(&typeID)
	if err != nil {
		return 0, fmt.Errorf("Relationship object type not found: %w", err)
	}

	var objectID int64
	err = r.db.QueryRow(ctx, `
		INSERT INTO objects_service.objects (public_id, object_type_id, name, status, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, 'Relationship', 'active', NOW(), NOW())
		RETURNING id
	`, typeID).Scan(&objectID)
	if err != nil {
		return 0, fmt.Errorf("failed to create base object: %w", err)
	}

	return objectID, nil
}

func (r *relationshipRepository) deleteBaseObject(ctx context.Context, objectID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM objects_service.objects WHERE id = $1`, objectID)
	return err
}

func (r *relationshipRepository) GetByObjectID(ctx context.Context, objectID int64) (*models.Relationship, error) {
	r.metrics.QueryCount++

	query := `
		SELECT 
			r.object_id, r.source_object_id, r.target_object_id, r.relationship_type_id,
			r.status, r.relationship_metadata, r.created_by, r.updated_by,
			r.created_at, r.updated_at,
			s.public_id, t.public_id, rt.type_key
		FROM objects_service.objects_relationships r
		JOIN objects_service.objects s ON r.source_object_id = s.id
		JOIN objects_service.objects t ON r.target_object_id = t.id
		JOIN objects_service.objects_relationship_types rt ON r.relationship_type_id = rt.object_id
		WHERE r.object_id = $1`

	var rel models.Relationship
	var metadata []byte

	err := r.db.QueryRow(ctx, query, objectID).Scan(
		&rel.ObjectID, &rel.SourceObjectID, &rel.TargetObjectID, &rel.RelationshipTypeID,
		&rel.Status, &metadata, &rel.CreatedBy, &rel.UpdatedBy,
		&rel.CreatedAt, &rel.UpdatedAt,
		&rel.SourceObjectPublicID, &rel.TargetObjectPublicID, &rel.RelationshipTypeKey,
	)
	if err != nil {
		r.metrics.ErrorCount++
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get relationship: %w", err)
	}

	rel.RelationshipMetadata = metadata
	return &rel, nil
}

func (r *relationshipRepository) GetByPublicID(ctx context.Context, publicID uuid.UUID) (*models.Relationship, error) {
	r.metrics.QueryCount++

	var objectID int64
	err := r.db.QueryRow(ctx, `
		SELECT id FROM objects_service.objects WHERE public_id = $1 AND object_type_id = (
			SELECT id FROM objects_service.object_types WHERE name = 'Relationship'
		)
	`, publicID).Scan(&objectID)
	if err != nil {
		r.metrics.ErrorCount++
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get object id: %w", err)
	}

	return r.GetByObjectID(ctx, objectID)
}

func (r *relationshipRepository) Update(ctx context.Context, objectID int64, input *models.UpdateRelationshipRequest) (*models.Relationship, error) {
	r.metrics.QueryCount++

	current, err := r.GetByObjectID(ctx, objectID)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, err
	}

	updates := []string{}
	args := []interface{}{}
	argNum := 1

	if input.Status != nil {
		updates = append(updates, fmt.Sprintf("status = $%d", argNum))
		args = append(args, *input.Status)
		argNum++
	}

	if input.RelationshipMetadata != nil {
		metadata, err := json.Marshal(input.RelationshipMetadata)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		updates = append(updates, fmt.Sprintf("relationship_metadata = $%d", argNum))
		args = append(args, metadata)
		argNum++
	}

	updates = append(updates, "updated_at = NOW()")
	if input.UpdatedBy != "" {
		updates = append(updates, fmt.Sprintf("updated_by = $%d", argNum))
		args = append(args, input.UpdatedBy)
		argNum++
	}

	args = append(args, objectID)
	query := fmt.Sprintf(`
		UPDATE objects_service.objects_relationships
		SET %s
		WHERE object_id = $%d
		RETURNING object_id, source_object_id, target_object_id, relationship_type_id,
			status, relationship_metadata, created_by, updated_by, created_at, updated_at`,
		strings.Join(updates, ", "), argNum)

	var rel models.Relationship
	var metadata []byte

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&rel.ObjectID, &rel.SourceObjectID, &rel.TargetObjectID, &rel.RelationshipTypeID,
		&rel.Status, &metadata, &rel.CreatedBy, &rel.UpdatedBy,
		&rel.CreatedAt, &rel.UpdatedAt,
	)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to update relationship: %w", err)
	}

	rel.RelationshipMetadata = metadata
	rel.SourceObjectPublicID = current.SourceObjectPublicID
	rel.TargetObjectPublicID = current.TargetObjectPublicID
	rel.RelationshipTypeKey = current.RelationshipTypeKey

	return &rel, nil
}

func (r *relationshipRepository) Delete(ctx context.Context, objectID int64) error {
	r.metrics.QueryCount++

	_, err := r.db.Exec(ctx, `DELETE FROM objects_service.objects_relationships WHERE object_id = $1`, objectID)
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("failed to delete relationship: %w", err)
	}

	_, err = r.db.Exec(ctx, `DELETE FROM objects_service.objects WHERE id = $1`, objectID)
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("failed to delete base object: %w", err)
	}

	return nil
}

func (r *relationshipRepository) DeleteByPublicID(ctx context.Context, publicID uuid.UUID) error {
	rel, err := r.GetByPublicID(ctx, publicID)
	if err != nil {
		return err
	}
	return r.Delete(ctx, rel.ObjectID)
}

func (r *relationshipRepository) List(ctx context.Context, filter *models.RelationshipFilter) ([]*models.Relationship, error) {
	r.metrics.QueryCount++

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = r.options.DefaultPageSize
	}
	if filter.PageSize > r.options.MaxPageSize {
		filter.PageSize = r.options.MaxPageSize
	}

	whereClauses := []string{"1=1"}
	args := []interface{}{}
	argNum := 1

	if filter.SourceObjectPublicID != nil && *filter.SourceObjectPublicID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("s.public_id::text = $%d", argNum))
		args = append(args, *filter.SourceObjectPublicID)
		argNum++
	}

	if filter.TargetObjectPublicID != nil && *filter.TargetObjectPublicID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("t.public_id::text = $%d", argNum))
		args = append(args, *filter.TargetObjectPublicID)
		argNum++
	}

	if filter.RelationshipTypeKey != nil && *filter.RelationshipTypeKey != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("rt.type_key = $%d", argNum))
		args = append(args, *filter.RelationshipTypeKey)
		argNum++
	}

	if filter.Status != nil && *filter.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("r.status = $%d", argNum))
		args = append(args, *filter.Status)
		argNum++
	}

	where := strings.Join(whereClauses, " AND ")

	orderBy := "r.created_at"
	if filter.SortBy != "" {
		orderBy = filter.SortBy
	}
	order := "DESC"
	if filter.SortOrder == "asc" {
		order = "ASC"
	}

	offset := (filter.Page - 1) * filter.PageSize

	query := fmt.Sprintf(`
		SELECT 
			r.object_id, r.source_object_id, r.target_object_id, r.relationship_type_id,
			r.status, r.relationship_metadata, r.created_by, r.updated_by,
			r.created_at, r.updated_at,
			s.public_id, t.public_id, rt.type_key
		FROM objects_service.objects_relationships r
		JOIN objects_service.objects s ON r.source_object_id = s.id
		JOIN objects_service.objects t ON r.target_object_id = t.id
		JOIN objects_service.objects_relationship_types rt ON r.relationship_type_id = rt.object_id
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, where, orderBy, order, argNum, argNum+1)

	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to list relationships: %w", err)
	}
	defer rows.Close()

	var rels []*models.Relationship
	for rows.Next() {
		var rel models.Relationship
		var metadata []byte

		err := rows.Scan(
			&rel.ObjectID, &rel.SourceObjectID, &rel.TargetObjectID, &rel.RelationshipTypeID,
			&rel.Status, &metadata, &rel.CreatedBy, &rel.UpdatedBy,
			&rel.CreatedAt, &rel.UpdatedAt,
			&rel.SourceObjectPublicID, &rel.TargetObjectPublicID, &rel.RelationshipTypeKey,
		)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to scan relationship: %w", err)
		}
		rel.RelationshipMetadata = metadata
		rels = append(rels, &rel)
	}

	if err := rows.Err(); err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("error iterating relationships: %w", err)
	}

	return rels, nil
}

func (r *relationshipRepository) GetForObject(ctx context.Context, objectPublicID uuid.UUID, filter *models.RelationshipFilterForType) ([]*models.Relationship, error) {
	r.metrics.QueryCount++

	object, err := r.objectRepo.GetByPublicID(ctx, objectPublicID)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("object not found: %w", err)
	}

	if filter == nil {
		filter = &models.RelationshipFilterForType{}
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = r.options.DefaultPageSize
	}
	if filter.PageSize > r.options.MaxPageSize {
		filter.PageSize = r.options.MaxPageSize
	}

	whereClauses := []string{"(r.source_object_id = $1 OR r.target_object_id = $1)"}
	args := []interface{}{object.ID}
	argNum := 2

	if filter.Status != nil && *filter.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("r.status = $%d", argNum))
		args = append(args, *filter.Status)
		argNum++
	}

	where := strings.Join(whereClauses, " AND ")

	offset := (filter.Page - 1) * filter.PageSize

	query := fmt.Sprintf(`
		SELECT 
			r.object_id, r.source_object_id, r.target_object_id, r.relationship_type_id,
			r.status, r.relationship_metadata, r.created_by, r.updated_by,
			r.created_at, r.updated_at,
			s.public_id, t.public_id, rt.type_key
		FROM objects_service.objects_relationships r
		JOIN objects_service.objects s ON r.source_object_id = s.id
		JOIN objects_service.objects t ON r.target_object_id = t.id
		JOIN objects_service.objects_relationship_types rt ON r.relationship_type_id = rt.object_id
		WHERE %s
		ORDER BY r.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argNum, argNum+1)

	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get relationships for object: %w", err)
	}
	defer rows.Close()

	var rels []*models.Relationship
	for rows.Next() {
		var rel models.Relationship
		var metadata []byte

		err := rows.Scan(
			&rel.ObjectID, &rel.SourceObjectID, &rel.TargetObjectID, &rel.RelationshipTypeID,
			&rel.Status, &metadata, &rel.CreatedBy, &rel.UpdatedBy,
			&rel.CreatedAt, &rel.UpdatedAt,
			&rel.SourceObjectPublicID, &rel.TargetObjectPublicID, &rel.RelationshipTypeKey,
		)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to scan relationship: %w", err)
		}
		rel.RelationshipMetadata = metadata
		rels = append(rels, &rel)
	}

	return rels, nil
}

func (r *relationshipRepository) GetForObjectByType(ctx context.Context, objectPublicID uuid.UUID, typeKey string) ([]*models.Relationship, error) {
	r.metrics.QueryCount++

	object, err := r.objectRepo.GetByPublicID(ctx, objectPublicID)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("object not found: %w", err)
	}

	query := `
		SELECT 
			r.object_id, r.source_object_id, r.target_object_id, r.relationship_type_id,
			r.status, r.relationship_metadata, r.created_by, r.updated_by,
			r.created_at, r.updated_at,
			s.public_id, t.public_id, rt.type_key
		FROM objects_service.objects_relationships r
		JOIN objects_service.objects s ON r.source_object_id = s.id
		JOIN objects_service.objects t ON r.target_object_id = t.id
		JOIN objects_service.objects_relationship_types rt ON r.relationship_type_id = rt.object_id
		WHERE (r.source_object_id = $1 OR r.target_object_id = $1)
			AND rt.type_key = $2
		ORDER BY r.created_at DESC`

	rows, err := r.db.Query(ctx, query, object.ID, typeKey)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get relationships for object by type: %w", err)
	}
	defer rows.Close()

	var rels []*models.Relationship
	for rows.Next() {
		var rel models.Relationship
		var metadata []byte

		err := rows.Scan(
			&rel.ObjectID, &rel.SourceObjectID, &rel.TargetObjectID, &rel.RelationshipTypeID,
			&rel.Status, &metadata, &rel.CreatedBy, &rel.UpdatedBy,
			&rel.CreatedAt, &rel.UpdatedAt,
			&rel.SourceObjectPublicID, &rel.TargetObjectPublicID, &rel.RelationshipTypeKey,
		)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to scan relationship: %w", err)
		}
		rel.RelationshipMetadata = metadata
		rels = append(rels, &rel)
	}

	return rels, nil
}

func (r *relationshipRepository) GetRelatedObjects(ctx context.Context, objectPublicID uuid.UUID, typeKey *string) ([]*models.Object, error) {
	r.metrics.QueryCount++

	object, err := r.objectRepo.GetByPublicID(ctx, objectPublicID)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("object not found: %w", err)
	}

	var rows Rows
	var query string

	if typeKey != nil && *typeKey != "" {
		query = `
			SELECT DISTINCT o.*
			FROM objects_service.objects_relationships r
			JOIN objects_service.objects o ON r.target_object_id = o.id
			JOIN objects_service.objects_relationship_types rt ON r.relationship_type_id = rt.object_id
			WHERE r.source_object_id = $1 AND rt.type_key = $2
			UNION
			SELECT DISTINCT o.*
			FROM objects_service.objects_relationships r
			JOIN objects_service.objects o ON r.source_object_id = o.id
			JOIN objects_service.objects_relationship_types rt ON r.relationship_type_id = rt.object_id
			WHERE r.target_object_id = $1 AND rt.type_key = $2`
		rows, err = r.db.Query(ctx, query, object.ID, *typeKey)
	} else {
		query = `
			SELECT DISTINCT o.*
			FROM objects_service.objects_relationships r
			JOIN objects_service.objects o ON r.target_object_id = o.id
			WHERE r.source_object_id = $1
			UNION
			SELECT DISTINCT o.*
			FROM objects_service.objects_relationships r
			JOIN objects_service.objects o ON r.source_object_id = o.id
			WHERE r.target_object_id = $1`
		rows, err = r.db.Query(ctx, query, object.ID)
	}

	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get related objects: %w", err)
	}
	defer rows.Close()

	var objects []*models.Object
	for rows.Next() {
		var obj models.Object
		err := rows.Scan(
			&obj.ID, &obj.PublicID, &obj.ObjectTypeID, &obj.ParentObjectID,
			&obj.Name, &obj.Description, &obj.Metadata, &obj.Tags,
			&obj.Status, &obj.Version, &obj.CreatedBy, &obj.UpdatedBy,
			&obj.CreatedAt, &obj.UpdatedAt, &obj.DeletedAt,
		)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to scan object: %w", err)
		}
		objects = append(objects, &obj)
	}

	return objects, nil
}

func (r *relationshipRepository) Exists(ctx context.Context, sourceObjectID, targetObjectID int64, typeObjectID int64) (bool, error) {
	r.metrics.QueryCount++

	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM objects_service.objects_relationships
			WHERE source_object_id = $1 AND target_object_id = $2 AND relationship_type_id = $3
		)
	`, sourceObjectID, targetObjectID, typeObjectID).Scan(&exists)
	if err != nil {
		r.metrics.ErrorCount++
		return false, fmt.Errorf("failed to check relationship existence: %w", err)
	}

	return exists, nil
}

func (r *relationshipRepository) CountForObject(ctx context.Context, objectID int64, typeKey *string) (int, error) {
	r.metrics.QueryCount++

	var count int
	var err error

	if typeKey != nil && *typeKey != "" {
		err = r.db.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM objects_service.objects_relationships r
			JOIN objects_service.objects_relationship_types rt ON r.relationship_type_id = rt.object_id
			WHERE (r.source_object_id = $1 OR r.target_object_id = $1) AND rt.type_key = $2
		`, objectID, *typeKey).Scan(&count)
	} else {
		err = r.db.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM objects_service.objects_relationships
			WHERE source_object_id = $1 OR target_object_id = $1
		`, objectID).Scan(&count)
	}

	if err != nil {
		r.metrics.ErrorCount++
		return 0, fmt.Errorf("failed to count relationships: %w", err)
	}

	return count, nil
}

func (r *relationshipRepository) GetByTypeKey(ctx context.Context, typeKey string) ([]*models.Relationship, error) {
	r.metrics.QueryCount++

	query := `
		SELECT 
			r.object_id, r.source_object_id, r.target_object_id, r.relationship_type_id,
			r.status, r.relationship_metadata, r.created_by, r.updated_by,
			r.created_at, r.updated_at,
			s.public_id, t.public_id, rt.type_key
		FROM objects_service.objects_relationships r
		JOIN objects_service.objects s ON r.source_object_id = s.id
		JOIN objects_service.objects t ON r.target_object_id = t.id
		JOIN objects_service.objects_relationship_types rt ON r.relationship_type_id = rt.object_id
		WHERE rt.type_key = $1
		ORDER BY r.created_at DESC`

	rows, err := r.db.Query(ctx, query, typeKey)
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to get relationships by type: %w", err)
	}
	defer rows.Close()

	var rels []*models.Relationship
	for rows.Next() {
		var rel models.Relationship
		var metadata []byte

		err := rows.Scan(
			&rel.ObjectID, &rel.SourceObjectID, &rel.TargetObjectID, &rel.RelationshipTypeID,
			&rel.Status, &metadata, &rel.CreatedBy, &rel.UpdatedBy,
			&rel.CreatedAt, &rel.UpdatedAt,
			&rel.SourceObjectPublicID, &rel.TargetObjectPublicID, &rel.RelationshipTypeKey,
		)
		if err != nil {
			r.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to scan relationship: %w", err)
		}
		rel.RelationshipMetadata = metadata
		rels = append(rels, &rel)
	}

	return rels, nil
}

func (r *relationshipRepository) CheckCircular(ctx context.Context, sourceObjectID, targetObjectID, typeObjectID int64) (bool, error) {
	r.metrics.QueryCount++

	var typeKey string
	err := r.db.QueryRow(ctx, `
		SELECT type_key FROM objects_service.objects_relationship_types WHERE object_id = $1
	`, typeObjectID).Scan(&typeKey)
	if err != nil {
		r.metrics.ErrorCount++
		return false, fmt.Errorf("failed to get relationship type: %w", err)
	}

	var cardinality string
	err = r.db.QueryRow(ctx, `
		SELECT cardinality FROM objects_service.objects_relationship_types WHERE object_id = $1
	`, typeObjectID).Scan(&cardinality)
	if err != nil {
		r.metrics.ErrorCount++
		return false, fmt.Errorf("failed to get cardinality: %w", err)
	}

	if cardinality != models.CardinalityOneToMany && cardinality != models.CardinalityManyToOne {
		return false, nil
	}

	var isCircular bool
	err = r.db.QueryRow(ctx, `
		WITH RECURSIVE relationship_path AS (
			SELECT target_object_id, 1 as depth
			FROM objects_service.objects_relationships
			WHERE source_object_id = $1 AND relationship_type_id = $2

			UNION

			SELECT r.target_object_id, rp.depth + 1
			FROM objects_service.objects_relationships r
			INNER JOIN relationship_path rp ON r.source_object_id = rp.target_object_id
			WHERE r.relationship_type_id = $2 AND rp.depth < 100
		)
		SELECT EXISTS(SELECT 1 FROM relationship_path WHERE target_object_id = $1)
	`, targetObjectID, typeObjectID).Scan(&isCircular)
	if err != nil {
		r.metrics.ErrorCount++
		return false, fmt.Errorf("failed to check circular relationship: %w", err)
	}

	return isCircular, nil
}
