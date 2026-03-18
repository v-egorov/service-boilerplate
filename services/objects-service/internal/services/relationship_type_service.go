package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
)

var (
	ErrRelationshipTypeNotFound  = errors.New("relationship type not found")
	ErrDuplicateRelationshipType = errors.New("relationship type already exists")
	ErrInvalidCardinality        = errors.New("invalid cardinality value")
	ErrInvalidReverseType        = errors.New("reverse type key does not exist")
	ErrRelationshipTypeInUse     = errors.New("relationship type is in use and cannot be deleted")
	ErrInvalidCountConstraint    = errors.New("min_count cannot exceed max_count")
	ErrTypeKeyRequired           = errors.New("type_key is required")
	ErrCardinalityRequired       = errors.New("cardinality is required")
)

// RelationshipTypeService defines operations for relationship type management
type RelationshipTypeService interface {
	Create(ctx context.Context, req *models.CreateRelationshipTypeRequest) (*models.RelationshipType, error)
	GetByTypeKey(ctx context.Context, typeKey string) (*models.RelationshipType, error)
	Update(ctx context.Context, typeKey string, req *models.UpdateRelationshipTypeRequest) (*models.RelationshipType, error)
	Delete(ctx context.Context, typeKey string) error
	List(ctx context.Context, filter *models.RelationshipTypeFilter) ([]*models.RelationshipType, error)
}

// relationshipTypeService implements RelationshipTypeService
type relationshipTypeService struct {
	repo repository.RelationshipTypeRepository
}

// NewRelationshipTypeService creates a new RelationshipTypeService instance
func NewRelationshipTypeService(repo repository.RelationshipTypeRepository) RelationshipTypeService {
	return &relationshipTypeService{repo: repo}
}

// Create creates a new relationship type
func (s *relationshipTypeService) Create(ctx context.Context, req *models.CreateRelationshipTypeRequest) (*models.RelationshipType, error) {
	// Validate type_key
	if req.TypeKey == "" {
		return nil, fmt.Errorf("%w: %s", ErrTypeKeyRequired, repository.ErrInvalidInput)
	}

	// Validate cardinality
	if req.Cardinality == "" {
		return nil, fmt.Errorf("%w: %s", ErrCardinalityRequired, repository.ErrInvalidInput)
	}

	if !models.IsValidCardinality(req.Cardinality) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidCardinality, repository.ErrInvalidInput)
	}

	// Check if type_key already exists
	exists, err := s.repo.Exists(ctx, req.TypeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check relationship type existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("%w: type_key '%s' already exists", ErrDuplicateRelationshipType, req.TypeKey)
	}

	// Validate reverse_type_key if provided
	if req.ReverseTypeKey != nil && *req.ReverseTypeKey != "" {
		// Check that reverse type exists (but not self-reference)
		if *req.ReverseTypeKey == req.TypeKey {
			return nil, fmt.Errorf("%w: cannot reference itself", ErrInvalidReverseType)
		}

		reverseExists, err := s.repo.Exists(ctx, *req.ReverseTypeKey)
		if err != nil {
			return nil, fmt.Errorf("failed to check reverse type existence: %w", err)
		}
		if !reverseExists {
			return nil, fmt.Errorf("%w: type_key '%s' does not exist", ErrInvalidReverseType, *req.ReverseTypeKey)
		}
	}

	// Validate count constraints
	if req.MinCount > req.MaxCount && req.MaxCount >= 0 {
		return nil, fmt.Errorf("%w: min_count cannot exceed max_count", ErrInvalidCountConstraint)
	}

	// Set defaults
	if req.MinCount < 0 {
		req.MinCount = 0
	}
	if req.MaxCount < -1 {
		req.MaxCount = -1
	}

	return s.repo.Create(ctx, req)
}

// GetByTypeKey retrieves a relationship type by type_key
func (s *relationshipTypeService) GetByTypeKey(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
	if typeKey == "" {
		return nil, fmt.Errorf("%w: type_key is required", repository.ErrInvalidInput)
	}

	rt, err := s.repo.GetByTypeKey(ctx, typeKey)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: type_key '%s'", ErrRelationshipTypeNotFound, typeKey)
		}
		return nil, fmt.Errorf("failed to get relationship type: %w", err)
	}

	return rt, nil
}

// Update updates a relationship type
func (s *relationshipTypeService) Update(ctx context.Context, typeKey string, req *models.UpdateRelationshipTypeRequest) (*models.RelationshipType, error) {
	if typeKey == "" {
		return nil, fmt.Errorf("%w: type_key is required", repository.ErrInvalidInput)
	}

	// Get existing relationship type
	existing, err := s.repo.GetByTypeKey(ctx, typeKey)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: type_key '%s'", ErrRelationshipTypeNotFound, typeKey)
		}
		return nil, fmt.Errorf("failed to get relationship type: %w", err)
	}

	// Validate cardinality if provided
	if req.Cardinality != nil && !models.IsValidCardinality(*req.Cardinality) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidCardinality, repository.ErrInvalidInput)
	}

	// Validate reverse_type_key if provided
	if req.ReverseTypeKey != nil && *req.ReverseTypeKey != "" {
		// Allow self-reference if it's being removed (set to empty)
		if *req.ReverseTypeKey != typeKey {
			reverseExists, err := s.repo.Exists(ctx, *req.ReverseTypeKey)
			if err != nil {
				return nil, fmt.Errorf("failed to check reverse type existence: %w", err)
			}
			if !reverseExists {
				return nil, fmt.Errorf("%w: type_key '%s' does not exist", ErrInvalidReverseType, *req.ReverseTypeKey)
			}
		}
	}

	// Validate count constraints if provided
	if req.MinCount != nil && req.MaxCount != nil {
		if *req.MinCount > *req.MaxCount && *req.MaxCount >= 0 {
			return nil, fmt.Errorf("%w: min_count cannot exceed max_count", ErrInvalidCountConstraint)
		}
	}

	return s.repo.Update(ctx, existing.ObjectID, req)
}

// Delete deletes a relationship type
func (s *relationshipTypeService) Delete(ctx context.Context, typeKey string) error {
	if typeKey == "" {
		return fmt.Errorf("%w: type_key is required", repository.ErrInvalidInput)
	}

	// Get existing relationship type
	existing, err := s.repo.GetByTypeKey(ctx, typeKey)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: type_key '%s'", ErrRelationshipTypeNotFound, typeKey)
		}
		return fmt.Errorf("failed to get relationship type: %w", err)
	}

	// TODO: Check if relationship type is in use before deletion
	// This will be implemented in Phase R2 when we have relationship instances

	return s.repo.Delete(ctx, existing.ObjectID)
}

// List retrieves relationship types with filtering and pagination
func (s *relationshipTypeService) List(ctx context.Context, filter *models.RelationshipTypeFilter) ([]*models.RelationshipType, error) {
	if filter == nil {
		filter = &models.RelationshipTypeFilter{}
	}

	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}

	return s.repo.List(ctx, filter)
}
