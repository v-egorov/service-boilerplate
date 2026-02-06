package services

import (
	"context"
	"fmt"

	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
)

type ObjectTypeService interface {
	Create(ctx context.Context, req *models.CreateObjectTypeRequest) (*models.ObjectType, error)
	GetByID(ctx context.Context, id int64) (*models.ObjectType, error)
	GetByName(ctx context.Context, name string) (*models.ObjectType, error)
	Update(ctx context.Context, id int64, req *models.UpdateObjectTypeRequest) (*models.ObjectType, error)
	Delete(ctx context.Context, id int64) error
	GetTree(ctx context.Context, rootID *int64) ([]*models.ObjectType, error)
	GetChildren(ctx context.Context, parentID int64) ([]*models.ObjectType, error)
	GetDescendants(ctx context.Context, rootID int64, maxDepth *int) ([]*models.ObjectType, error)
	GetAncestors(ctx context.Context, id int64) ([]*models.ObjectType, error)
	GetPath(ctx context.Context, id int64) ([]*models.ObjectType, error)
	List(ctx context.Context, filter *models.ObjectTypeFilter) ([]*models.ObjectType, error)
	Search(ctx context.Context, query string, limit int) ([]*models.ObjectType, error)
	ValidateMove(ctx context.Context, id int64, newParentID *int64) error
	GetSubtreeObjectCount(ctx context.Context, id int64) (int64, error)
}

type objectTypeService struct {
	repo repository.ObjectTypeRepository
}

func NewObjectTypeService(repo repository.ObjectTypeRepository) ObjectTypeService {
	return &objectTypeService{repo: repo}
}

func (s *objectTypeService) Create(ctx context.Context, req *models.CreateObjectTypeRequest) (*models.ObjectType, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required: %w", repository.ErrInvalidInput)
	}

	return s.repo.Create(ctx, req)
}

func (s *objectTypeService) GetByID(ctx context.Context, id int64) (*models.ObjectType, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id: %w", repository.ErrInvalidInput)
	}

	objectType, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, fmt.Errorf("object type not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get object type: %w", err)
	}

	return objectType, nil
}

func (s *objectTypeService) GetByName(ctx context.Context, name string) (*models.ObjectType, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required: %w", repository.ErrInvalidInput)
	}

	objectType, err := s.repo.GetByName(ctx, name)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, fmt.Errorf("object type not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get object type: %w", err)
	}

	return objectType, nil
}

func (s *objectTypeService) Update(ctx context.Context, id int64, req *models.UpdateObjectTypeRequest) (*models.ObjectType, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id: %w", repository.ErrInvalidInput)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("object type not found: %w", err)
	}

	if req.Name != nil && *req.Name == "" {
		return nil, fmt.Errorf("name cannot be empty: %w", repository.ErrInvalidInput)
	}

	if req.IsSealed != nil && *req.IsSealed && existing.IsSealed {
		return nil, fmt.Errorf("cannot modify sealed object type: %w", repository.ErrInvalidInput)
	}

	return s.repo.Update(ctx, id, req)
}

func (s *objectTypeService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid id: %w", repository.ErrInvalidInput)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("object type not found: %w", err)
	}

	if existing.IsSealed {
		return fmt.Errorf("cannot delete sealed object type: %w", repository.ErrInvalidInput)
	}

	objectCount, err := s.repo.GetSubtreeObjectCount(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check object count: %w", err)
	}

	if objectCount > 0 {
		return fmt.Errorf("cannot delete object type with existing objects: %w", repository.ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *objectTypeService) GetTree(ctx context.Context, rootID *int64) ([]*models.ObjectType, error) {
	return s.repo.GetTree(ctx, rootID)
}

func (s *objectTypeService) GetChildren(ctx context.Context, parentID int64) ([]*models.ObjectType, error) {
	return s.repo.GetChildren(ctx, parentID)
}

func (s *objectTypeService) GetDescendants(ctx context.Context, rootID int64, maxDepth *int) ([]*models.ObjectType, error) {
	return s.repo.GetDescendants(ctx, rootID, maxDepth)
}

func (s *objectTypeService) GetAncestors(ctx context.Context, id int64) ([]*models.ObjectType, error) {
	return s.repo.GetAncestors(ctx, id)
}

func (s *objectTypeService) GetPath(ctx context.Context, id int64) ([]*models.ObjectType, error) {
	return s.repo.GetPath(ctx, id)
}

func (s *objectTypeService) List(ctx context.Context, filter *models.ObjectTypeFilter) ([]*models.ObjectType, error) {
	return s.repo.List(ctx, filter)
}

func (s *objectTypeService) Search(ctx context.Context, query string, limit int) ([]*models.ObjectType, error) {
	if query == "" {
		return nil, fmt.Errorf("search query is required: %w", repository.ErrInvalidInput)
	}

	return s.repo.Search(ctx, query, limit)
}

func (s *objectTypeService) ValidateMove(ctx context.Context, id int64, newParentID *int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid id: %w", repository.ErrInvalidInput)
	}

	if newParentID != nil && *newParentID <= 0 {
		return fmt.Errorf("invalid new parent id: %w", repository.ErrInvalidInput)
	}

	return s.repo.ValidateMove(ctx, id, newParentID)
}

func (s *objectTypeService) GetSubtreeObjectCount(ctx context.Context, id int64) (int64, error) {
	if id <= 0 {
		return 0, fmt.Errorf("invalid id: %w", repository.ErrInvalidInput)
	}

	return s.repo.GetSubtreeObjectCount(ctx, id)
}
