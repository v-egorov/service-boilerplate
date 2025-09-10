package services

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	// SERVICE_IMPORT_MODELS
	// SERVICE_IMPORT_REPOSITORY
)

// Service type alias - will be replaced during template processing
type Service = struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   interface{}
	UpdatedAt   interface{}
}

type ServiceService struct {
	repo   Repository
	logger *logrus.Logger
}

func NewServiceService(repo Repository, logger *logrus.Logger) *ServiceService {
	return &ServiceService{
		repo:   repo,
		logger: logger,
	}
}

func (s *ServiceService) Create(ctx context.Context, req CreateServiceRequest) (*ServiceResponse, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		s.logger.WithError(err).Error("Invalid create service request")
		return nil, err
	}

	// Create service model
	service := &Service{
		Name:        req.Name,
		Description: req.Description,
	}

	// Save to database
	createdService, err := s.repo.Create(ctx, service)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create service in database")
		return nil, err
	}

	s.logger.WithField("id", createdService.ID).Info("Service created successfully")
	return s.toResponse(createdService), nil
}

func (s *ServiceService) GetByID(ctx context.Context, id int64) (*ServiceResponse, error) {
	if id <= 0 {
		return nil, errors.New("invalid service ID")
	}

	service, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to get service by ID")
		return nil, err
	}

	return s.toResponse(service), nil
}

func (s *ServiceService) Update(ctx context.Context, id int64, req UpdateServiceRequest) (*ServiceResponse, error) {
	if id <= 0 {
		return nil, errors.New("invalid service ID")
	}

	// Validate request
	if err := s.validateUpdateRequest(req); err != nil {
		s.logger.WithError(err).Error("Invalid update service request")
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
	updatedService, err := s.repo.Update(ctx, id, updates)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to update service")
		return nil, err
	}

	s.logger.WithField("id", id).Info("Service updated successfully")
	return s.toResponse(updatedService), nil
}

func (s *ServiceService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid service ID")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to delete service")
		return err
	}

	s.logger.WithField("id", id).Info("Service deleted successfully")
	return nil
}

func (s *ServiceService) List(ctx context.Context, limit, offset int) ([]*ServiceResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	services, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list services")
		return nil, err
	}

	var responses []*ServiceResponse
	for _, service := range services {
		responses = append(responses, s.toResponse(service))
	}

	s.logger.WithField("count", len(responses)).Debug("Services listed successfully")
	return responses, nil
}

// Validation methods
func (s *ServiceService) validateCreateRequest(req CreateServiceRequest) error {
	if req.Name == "" {
		return errors.New("service name is required")
	}
	if len(req.Name) > 100 {
		return errors.New("service name must be less than 100 characters")
	}
	if len(req.Description) > 500 {
		return errors.New("service description must be less than 500 characters")
	}
	return nil
}

func (s *ServiceService) validateUpdateRequest(req UpdateServiceRequest) error {
	if req.Name != nil && *req.Name == "" {
		return errors.New("service name cannot be empty")
	}
	if req.Name != nil && len(*req.Name) > 100 {
		return errors.New("service name must be less than 100 characters")
	}
	if req.Description != nil && len(*req.Description) > 500 {
		return errors.New("service description must be less than 500 characters")
	}
	return nil
}

// Helper methods
func (s *ServiceService) toResponse(service *Service) *ServiceResponse {
	return &ServiceResponse{
		ID:          service.ID,
		Name:        service.Name,
		Description: service.Description,
		CreatedAt:   formatTime(service.CreatedAt),
		UpdatedAt:   formatTime(service.UpdatedAt),
	}
}

func formatTime(t interface{}) string {
	// Placeholder for time formatting - will be replaced during template processing
	return "2023-01-01T00:00:00Z"
}

// Repository interface for dependency injection
type Repository interface {
	Create(ctx context.Context, service *Service) (*Service, error)
	GetByID(ctx context.Context, id int64) (*Service, error)
	Update(ctx context.Context, id int64, updates map[string]interface{}) (*Service, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*Service, error)
}

// Request/Response types
type CreateServiceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateServiceRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type ServiceResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
