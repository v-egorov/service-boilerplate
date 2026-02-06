package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
)

type ObjectService interface {
	Create(ctx context.Context, req *models.CreateObjectRequest) (*models.Object, error)
	GetByID(ctx context.Context, id int64) (*models.Object, error)
	GetByPublicID(ctx context.Context, publicID uuid.UUID) (*models.Object, error)
	GetByName(ctx context.Context, name string) (*models.Object, error)
	Update(ctx context.Context, id int64, req *models.UpdateObjectRequest) (*models.Object, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter *models.ObjectFilter) ([]*models.Object, int64, error)
	Search(ctx context.Context, query string, limit int) ([]*models.Object, error)
	FindByMetadata(ctx context.Context, key, value string) ([]*models.Object, error)
	FindByTags(ctx context.Context, tags []string, matchAll bool) ([]*models.Object, error)
	UpdateMetadata(ctx context.Context, id int64, metadata map[string]interface{}) error
	AddTags(ctx context.Context, id int64, tags []string) error
	RemoveTags(ctx context.Context, id int64, tags []string) error
	GetChildren(ctx context.Context, parentID int64) ([]*models.Object, error)
	GetDescendants(ctx context.Context, rootID int64, maxDepth *int) ([]*models.Object, error)
	GetAncestors(ctx context.Context, id int64) ([]*models.Object, error)
	GetPath(ctx context.Context, id int64) ([]*models.Object, error)
	BulkCreate(ctx context.Context, objects []*models.CreateObjectRequest) ([]*models.Object, error)
	BulkUpdate(ctx context.Context, ids []int64, updates *models.UpdateObjectRequest) ([]*models.Object, error)
	BulkDelete(ctx context.Context, ids []int64) error
	ValidateParentChild(ctx context.Context, parentID, childID int64) error
	GetObjectStats(ctx context.Context, filter *models.ObjectFilter) (*repository.ObjectStats, error)
}

type objectService struct {
	repo           repository.ObjectRepository
	objectTypeRepo repository.ObjectTypeRepository
}

func NewObjectService(repo repository.ObjectRepository, objectTypeRepo repository.ObjectTypeRepository) ObjectService {
	return &objectService{
		repo:           repo,
		objectTypeRepo: objectTypeRepo,
	}
}

func (s *objectService) Create(ctx context.Context, req *models.CreateObjectRequest) (*models.Object, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required: %w", repository.ErrInvalidInput)
	}

	if req.ObjectTypeID <= 0 {
		return nil, fmt.Errorf("object_type_id is required: %w", repository.ErrInvalidInput)
	}

	objectType, err := s.objectTypeRepo.GetByID(ctx, req.ObjectTypeID)
	if err != nil {
		return nil, fmt.Errorf("invalid object type: %w", err)
	}

	if objectType.IsSealed {
		return nil, fmt.Errorf("cannot create objects of sealed type: %w", repository.ErrInvalidInput)
	}

	if req.ParentObjectID != nil {
		parent, err := s.repo.GetByID(ctx, *req.ParentObjectID)
		if err != nil {
			return nil, fmt.Errorf("invalid parent object: %w", err)
		}

		if parent.ObjectTypeID != req.ObjectTypeID {
			return nil, fmt.Errorf("parent object must be of same type: %w", repository.ErrInvalidInput)
		}

		if err := s.repo.ValidateParentChild(ctx, *req.ParentObjectID, 0); err != nil {
			return nil, fmt.Errorf("invalid parent-child relationship: %w", err)
		}
	}

	return s.repo.Create(ctx, req)
}

func (s *objectService) GetByID(ctx context.Context, id int64) (*models.Object, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id: %w", repository.ErrInvalidInput)
	}

	obj, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, fmt.Errorf("object not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return obj, nil
}

func (s *objectService) GetByPublicID(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
	if publicID == uuid.Nil {
		return nil, fmt.Errorf("invalid public id: %w", repository.ErrInvalidInput)
	}

	obj, err := s.repo.GetByPublicID(ctx, publicID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, fmt.Errorf("object not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return obj, nil
}

func (s *objectService) GetByName(ctx context.Context, name string) (*models.Object, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required: %w", repository.ErrInvalidInput)
	}

	obj, err := s.repo.GetByName(ctx, name)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, fmt.Errorf("object not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return obj, nil
}

func (s *objectService) Update(ctx context.Context, id int64, req *models.UpdateObjectRequest) (*models.Object, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id: %w", repository.ErrInvalidInput)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("object not found: %w", err)
	}

	if req.ObjectTypeID != nil {
		objectType, err := s.objectTypeRepo.GetByID(ctx, *req.ObjectTypeID)
		if err != nil {
			return nil, fmt.Errorf("invalid object type: %w", err)
		}

		if objectType.IsSealed {
			return nil, fmt.Errorf("cannot change to sealed type: %w", repository.ErrInvalidInput)
		}
	}

	if req.ParentObjectID != nil {
		if *req.ParentObjectID == id {
			return nil, fmt.Errorf("object cannot be its own parent: %w", repository.ErrInvalidInput)
		}

		parent, err := s.repo.GetByID(ctx, *req.ParentObjectID)
		if err != nil {
			return nil, fmt.Errorf("invalid parent object: %w", err)
		}

		newTypeID := req.ObjectTypeID
		if newTypeID == nil {
			newTypeID = &existing.ObjectTypeID
		}

		if parent.ObjectTypeID != *newTypeID {
			return nil, fmt.Errorf("parent object must be of same type: %w", repository.ErrInvalidInput)
		}
	}

	return s.repo.Update(ctx, id, req)
}

func (s *objectService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid id: %w", repository.ErrInvalidInput)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("object not found: %w", err)
	}

	objectType, err := s.objectTypeRepo.GetByID(ctx, existing.ObjectTypeID)
	if err == nil && objectType.IsSealed {
		return fmt.Errorf("cannot delete objects of sealed type: %w", repository.ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *objectService) List(ctx context.Context, filter *models.ObjectFilter) ([]*models.Object, int64, error) {
	if filter != nil && filter.ObjectTypeID != nil {
		_, err := s.objectTypeRepo.GetByID(ctx, *filter.ObjectTypeID)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid object type: %w", err)
		}
	}

	return s.repo.List(ctx, filter)
}

func (s *objectService) Search(ctx context.Context, query string, limit int) ([]*models.Object, error) {
	if query == "" {
		return nil, fmt.Errorf("search query is required: %w", repository.ErrInvalidInput)
	}

	return s.repo.Search(ctx, query, limit)
}

func (s *objectService) FindByMetadata(ctx context.Context, key, value string) ([]*models.Object, error) {
	if key == "" {
		return nil, fmt.Errorf("metadata key is required: %w", repository.ErrInvalidInput)
	}

	return s.repo.FindByMetadata(ctx, key, value)
}

func (s *objectService) FindByTags(ctx context.Context, tags []string, matchAll bool) ([]*models.Object, error) {
	if len(tags) == 0 {
		return nil, fmt.Errorf("at least one tag is required: %w", repository.ErrInvalidInput)
	}

	return s.repo.FindByTags(ctx, tags, matchAll)
}

func (s *objectService) UpdateMetadata(ctx context.Context, id int64, metadata map[string]interface{}) error {
	if id <= 0 {
		return fmt.Errorf("invalid id: %w", repository.ErrInvalidInput)
	}

	if metadata == nil {
		return fmt.Errorf("metadata cannot be nil: %w", repository.ErrInvalidInput)
	}

	return s.repo.UpdateMetadata(ctx, id, metadata)
}

func (s *objectService) AddTags(ctx context.Context, id int64, tags []string) error {
	if id <= 0 {
		return fmt.Errorf("invalid id: %w", repository.ErrInvalidInput)
	}

	if len(tags) == 0 {
		return nil
	}

	return s.repo.AddTags(ctx, id, tags)
}

func (s *objectService) RemoveTags(ctx context.Context, id int64, tags []string) error {
	if id <= 0 {
		return fmt.Errorf("invalid id: %w", repository.ErrInvalidInput)
	}

	if len(tags) == 0 {
		return nil
	}

	return s.repo.RemoveTags(ctx, id, tags)
}

func (s *objectService) GetChildren(ctx context.Context, parentID int64) ([]*models.Object, error) {
	if parentID <= 0 {
		return nil, fmt.Errorf("invalid parent id: %w", repository.ErrInvalidInput)
	}

	return s.repo.GetChildren(ctx, parentID)
}

func (s *objectService) GetDescendants(ctx context.Context, rootID int64, maxDepth *int) ([]*models.Object, error) {
	if rootID <= 0 {
		return nil, fmt.Errorf("invalid root id: %w", repository.ErrInvalidInput)
	}

	return s.repo.GetDescendants(ctx, rootID, maxDepth)
}

func (s *objectService) GetAncestors(ctx context.Context, id int64) ([]*models.Object, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id: %w", repository.ErrInvalidInput)
	}

	return s.repo.GetAncestors(ctx, id)
}

func (s *objectService) GetPath(ctx context.Context, id int64) ([]*models.Object, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id: %w", repository.ErrInvalidInput)
	}

	return s.repo.GetPath(ctx, id)
}

func (s *objectService) BulkCreate(ctx context.Context, objects []*models.CreateObjectRequest) ([]*models.Object, error) {
	if len(objects) == 0 {
		return []*models.Object{}, nil
	}

	for i, obj := range objects {
		if obj.ObjectTypeID <= 0 {
			return nil, fmt.Errorf("object[%d]: object_type_id is required: %w", i, repository.ErrInvalidInput)
		}

		objectType, err := s.objectTypeRepo.GetByID(ctx, obj.ObjectTypeID)
		if err != nil {
			return nil, fmt.Errorf("object[%d]: invalid object type: %w", i, err)
		}

		if objectType.IsSealed {
			return nil, fmt.Errorf("object[%d]: cannot create objects of sealed type: %w", i, repository.ErrInvalidInput)
		}
	}

	return s.repo.BulkCreate(ctx, objects)
}

func (s *objectService) BulkUpdate(ctx context.Context, ids []int64, updates *models.UpdateObjectRequest) ([]*models.Object, error) {
	if len(ids) == 0 {
		return []*models.Object{}, nil
	}

	if updates == nil {
		return nil, fmt.Errorf("updates cannot be nil: %w", repository.ErrInvalidInput)
	}

	return s.repo.BulkUpdate(ctx, ids, updates)
}

func (s *objectService) BulkDelete(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	return s.repo.BulkDelete(ctx, ids)
}

func (s *objectService) ValidateParentChild(ctx context.Context, parentID, childID int64) error {
	if parentID <= 0 {
		return fmt.Errorf("invalid parent id: %w", repository.ErrInvalidInput)
	}

	if childID <= 0 {
		return fmt.Errorf("invalid child id: %w", repository.ErrInvalidInput)
	}

	return s.repo.ValidateParentChild(ctx, parentID, childID)
}

func (s *objectService) GetObjectStats(ctx context.Context, filter *models.ObjectFilter) (*repository.ObjectStats, error) {
	if filter != nil && filter.ObjectTypeID != nil {
		_, err := s.objectTypeRepo.GetByID(ctx, *filter.ObjectTypeID)
		if err != nil {
			return nil, fmt.Errorf("invalid object type: %w", err)
		}
	}

	return s.repo.GetObjectStats(ctx, filter)
}
