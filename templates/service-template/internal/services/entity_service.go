package services

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	// ENTITY_IMPORT_MODELS
)

// Import models package for Entity type
// This will be replaced with proper import during template processing

type EntityService struct {
	repo   Repository
	logger *logrus.Logger
}

func NewEntityService(repo Repository, logger *logrus.Logger) *EntityService {
	return &EntityService{
		repo:   repo,
		logger: logger,
	}
}

func (s *EntityService) Create(ctx context.Context, req CreateEntityRequest) (*EntityResponse, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		s.logger.WithError(err).Error("Invalid create entity request")
		return nil, err
	}

	// Create entity model
	entity := &models.Entity{
		Name:        req.Name,
		Description: req.Description,
	}

	// Save to database
	createdEntity, err := s.repo.Create(ctx, entity)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create entity in database")
		return nil, err
	}

	s.logger.WithField("id", createdEntity.ID).Info("Entity created successfully")
	return s.toResponse(createdEntity), nil
}

func (s *EntityService) GetByID(ctx context.Context, id int64) (*EntityResponse, error) {
	if id <= 0 {
		return nil, errors.New("invalid entity ID")
	}

	entity, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to get entity by ID")
		return nil, err
	}

	return s.toResponse(entity), nil
}

func (s *EntityService) Update(ctx context.Context, id int64, req UpdateEntityRequest) (*EntityResponse, error) {
	if id <= 0 {
		return nil, errors.New("invalid entity ID")
	}

	// Validate request
	if err := s.validateUpdateRequest(req); err != nil {
		s.logger.WithError(err).Error("Invalid update entity request")
		return nil, err
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	// Update in database
	updatedEntity, err := s.repo.Update(ctx, id, updates)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to update entity")
		return nil, err
	}

	s.logger.WithField("id", id).Info("Entity updated successfully")
	return s.toResponse(updatedEntity), nil
}

func (s *EntityService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid entity ID")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to delete entity")
		return err
	}

	s.logger.WithField("id", id).Info("Entity deleted successfully")
	return nil
}

func (s *EntityService) List(ctx context.Context, limit, offset int) ([]*EntityResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	entities, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list entities")
		return nil, err
	}

	var responses []*EntityResponse
	for _, entity := range entities {
		responses = append(responses, s.toResponse(entity))
	}

	s.logger.WithField("count", len(responses)).Debug("Entities listed successfully")
	return responses, nil
}

// Validation methods
func (s *EntityService) validateCreateRequest(req CreateEntityRequest) error {
	if req.Name == "" {
		return errors.New("entity name is required")
	}
	if len(req.Name) > 100 {
		return errors.New("entity name must be less than 100 characters")
	}
	if len(req.Description) > 500 {
		return errors.New("entity description must be less than 500 characters")
	}
	return nil
}

func (s *EntityService) validateUpdateRequest(req UpdateEntityRequest) error {
	if req.Name != nil && *req.Name == "" {
		return errors.New("entity name cannot be empty")
	}
	if req.Name != nil && len(*req.Name) > 100 {
		return errors.New("entity name must be less than 100 characters")
	}
	if req.Description != nil && len(*req.Description) > 500 {
		return errors.New("entity description must be less than 500 characters")
	}
	return nil
}

// Helper methods
func (s *EntityService) toResponse(entity *models.Entity) *EntityResponse {
	return &EntityResponse{
		ID:          entity.ID,
		Name:        entity.Name,
		Description: entity.Description,
		CreatedAt:   formatTime(entity.CreatedAt),
		UpdatedAt:   formatTime(entity.UpdatedAt),
	}
}

func formatTime(t interface{}) string {
	// Placeholder for time formatting - will be replaced during template processing
	return "2023-01-01T00:00:00Z"
}

// Repository interface for dependency injection
type Repository interface {
	Create(ctx context.Context, entity *models.Entity) (*models.Entity, error)
	GetByID(ctx context.Context, id int64) (*models.Entity, error)
	Update(ctx context.Context, id int64, updates map[string]interface{}) (*models.Entity, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*models.Entity, error)
}

// Request/Response types
type CreateEntityRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateEntityRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type EntityResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
