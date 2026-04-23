package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
)

var (
	ErrRelationshipNotFound  = errors.New("relationship not found")
	ErrDuplicateRelationship = errors.New("relationship already exists")
	ErrSourceObjectNotFound  = errors.New("source object not found")
	ErrTargetObjectNotFound  = errors.New("target object not found")
	ErrCircularRelationship  = errors.New("creating this relationship would create a cycle")
	ErrCardinalityViolation  = errors.New("relationship cardinality constraint violated")
	ErrSourceTargetSame      = errors.New("source and target cannot be the same")
)

const (
	StatusActive     = "active"
	StatusInactive   = "inactive"
	StatusDeprecated = "deprecated"
)

type RelationshipService interface {
	Create(ctx context.Context, input *models.CreateRelationshipRequest) (*models.Relationship, error)
	GetByPublicID(ctx context.Context, publicID uuid.UUID) (*models.Relationship, error)
	Update(ctx context.Context, publicID uuid.UUID, input *models.UpdateRelationshipRequest) (*models.Relationship, error)
	Delete(ctx context.Context, publicID uuid.UUID) error
	List(ctx context.Context, filter *models.RelationshipFilter) ([]*models.Relationship, error)
	GetForObject(ctx context.Context, objectPublicID uuid.UUID, filter *models.RelationshipFilterForType) ([]*models.Relationship, error)
	GetForObjectByType(ctx context.Context, objectPublicID uuid.UUID, typeKey string) ([]*models.Relationship, error)
	GetRelatedObjects(ctx context.Context, objectPublicID uuid.UUID, typeKey *string) ([]*models.Object, error)
}

type relationshipService struct {
	repo                 repository.RelationshipRepository
	relationshipTypeRepo repository.RelationshipTypeRepository
	objectRepo           repository.ObjectRepository
}

func NewRelationshipService(repo repository.RelationshipRepository, relTypeRepo repository.RelationshipTypeRepository, objectRepo repository.ObjectRepository) RelationshipService {
	return &relationshipService{
		repo:                 repo,
		relationshipTypeRepo: relTypeRepo,
		objectRepo:           objectRepo,
	}
}

func (s *relationshipService) Create(ctx context.Context, input *models.CreateRelationshipRequest) (*models.Relationship, error) {
	input.SetDefaults()

	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", repository.ErrInvalidInput, err)
	}

	sourcePublicID, err := uuid.Parse(input.SourceObjectPublicID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid source_object_id format", repository.ErrInvalidInput)
	}

	targetPublicID, err := uuid.Parse(input.TargetObjectPublicID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid target_object_id format", repository.ErrInvalidInput)
	}

	if sourcePublicID == targetPublicID {
		return nil, fmt.Errorf("%w: %s", ErrSourceTargetSame, repository.ErrInvalidInput)
	}

	sourceObject, err := s.objectRepo.GetByPublicID(ctx, sourcePublicID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrSourceObjectNotFound, repository.ErrNotFound)
	}

	targetObject, err := s.objectRepo.GetByPublicID(ctx, targetPublicID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTargetObjectNotFound, repository.ErrNotFound)
	}

	relType, err := s.relationshipTypeRepo.GetByTypeKey(ctx, input.RelationshipTypeKey)
	if err != nil {
		return nil, fmt.Errorf("%w: type_key '%s' not found", ErrRelationshipTypeNotFound, input.RelationshipTypeKey)
	}

	exists, err := s.repo.Exists(ctx, sourceObject.ID, targetObject.ID, relType.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to check relationship existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("%w: relationship already exists", ErrDuplicateRelationship)
	}

	isCircular, err := s.repo.CheckCircular(ctx, sourceObject.ID, targetObject.ID, relType.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to check circular relationship: %w", err)
	}
	if isCircular {
		return nil, fmt.Errorf("%w: %s", ErrCircularRelationship, repository.ErrInvalidInput)
	}

	if err := s.validateCardinality(ctx, sourceObject.ID, targetObject.ID, relType); err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, input)
}

func (s *relationshipService) validateCardinality(ctx context.Context, sourceObjectID, targetObjectID int64, relType *models.RelationshipType) error {
	switch relType.Cardinality {
	case models.CardinalityOneToOne:
		srcCount, _ := s.repo.CountForObject(ctx, sourceObjectID, &relType.TypeKey)
		tgtCount, _ := s.repo.CountForObject(ctx, targetObjectID, &relType.TypeKey)
		if srcCount >= 1 || tgtCount >= 1 {
			return fmt.Errorf("%w: one_to_one cardinality violated", ErrCardinalityViolation)
		}

	case models.CardinalityOneToMany:
		tgtCount, _ := s.repo.CountForObject(ctx, targetObjectID, &relType.TypeKey)
		if tgtCount >= 1 {
			return fmt.Errorf("%w: one_to_many target already has source", ErrCardinalityViolation)
		}
		if relType.MaxCount >= 0 {
			srcCount, _ := s.repo.CountForObject(ctx, sourceObjectID, &relType.TypeKey)
			if srcCount >= relType.MaxCount {
				return fmt.Errorf("%w: source reached max_count %d", ErrCardinalityViolation, relType.MaxCount)
			}
		}

	case models.CardinalityManyToOne:
		srcCount, _ := s.repo.CountForObject(ctx, sourceObjectID, &relType.TypeKey)
		if srcCount >= 1 {
			return fmt.Errorf("%w: many_to_one source already has target", ErrCardinalityViolation)
		}
		if relType.MaxCount >= 0 {
			tgtCount, _ := s.repo.CountForObject(ctx, targetObjectID, &relType.TypeKey)
			if tgtCount >= relType.MaxCount {
				return fmt.Errorf("%w: target reached max_count %d", ErrCardinalityViolation, relType.MaxCount)
			}
		}

	case models.CardinalityManyToMany:
		if relType.MaxCount >= 0 {
			srcCount, _ := s.repo.CountForObject(ctx, sourceObjectID, &relType.TypeKey)
			if srcCount >= relType.MaxCount {
				return fmt.Errorf("%w: source reached max_count %d", ErrCardinalityViolation, relType.MaxCount)
			}
		}
	}

	return nil
}

func (s *relationshipService) GetByPublicID(ctx context.Context, publicID uuid.UUID) (*models.Relationship, error) {
	rel, err := s.repo.GetByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: relationship not found", ErrRelationshipNotFound)
		}
		return nil, err
	}
	return rel, nil
}

func (s *relationshipService) Update(ctx context.Context, publicID uuid.UUID, input *models.UpdateRelationshipRequest) (*models.Relationship, error) {
	rel, err := s.repo.GetByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: relationship not found", ErrRelationshipNotFound)
		}
		return nil, err
	}

	objectID := rel.ObjectID
	return s.repo.Update(ctx, objectID, input)
}

func (s *relationshipService) Delete(ctx context.Context, publicID uuid.UUID) error {
	rel, err := s.repo.GetByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: relationship not found", ErrRelationshipNotFound)
		}
		return err
	}

	return s.repo.Delete(ctx, rel.ObjectID)
}

func (s *relationshipService) List(ctx context.Context, filter *models.RelationshipFilter) ([]*models.Relationship, error) {
	if filter == nil {
		filter = &models.RelationshipFilter{}
	}
	return s.repo.List(ctx, filter)
}

func (s *relationshipService) GetForObject(ctx context.Context, objectPublicID uuid.UUID, filter *models.RelationshipFilterForType) ([]*models.Relationship, error) {
	return s.repo.GetForObject(ctx, objectPublicID, filter)
}

func (s *relationshipService) GetForObjectByType(ctx context.Context, objectPublicID uuid.UUID, typeKey string) ([]*models.Relationship, error) {
	_, err := s.objectRepo.GetByPublicID(ctx, objectPublicID)
	if err != nil {
		return nil, fmt.Errorf("%w: object not found", repository.ErrNotFound)
	}
	return s.repo.GetForObjectByType(ctx, objectPublicID, typeKey)
}

func (s *relationshipService) GetRelatedObjects(ctx context.Context, objectPublicID uuid.UUID, typeKey *string) ([]*models.Object, error) {
	return s.repo.GetRelatedObjects(ctx, objectPublicID, typeKey)
}
