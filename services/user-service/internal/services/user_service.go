package services

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/services/user-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/user-service/internal/repository"
)

type UserService struct {
	repo   *repository.UserRepository
	logger *logrus.Logger
}

func NewUserService(repo *repository.UserRepository, logger *logrus.Logger) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.UserResponse, error) {
	// Check if user already exists
	existing, err := s.repo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	user := &models.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	created, err := s.repo.Create(ctx, user)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create user in service")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return s.toResponse(created), nil
}

func (s *UserService) GetUser(ctx context.Context, id int) (*models.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user in service")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return s.toResponse(user), nil
}

func (s *UserService) UpdateUser(ctx context.Context, id int, req *models.UpdateUserRequest) (*models.UserResponse, error) {
	// Get existing user
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Update fields if provided
	if req.Email != "" {
		existing.Email = req.Email
	}
	if req.FirstName != "" {
		existing.FirstName = req.FirstName
	}
	if req.LastName != "" {
		existing.LastName = req.LastName
	}

	updated, err := s.repo.Update(ctx, id, existing)
	if err != nil {
		s.logger.WithError(err).Error("Failed to update user in service")
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.toResponse(updated), nil
}

func (s *UserService) DeleteUser(ctx context.Context, id int) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.WithError(err).Error("Failed to delete user in service")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]*models.UserResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	users, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list users in service")
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	var responses []*models.UserResponse
	for _, user := range users {
		responses = append(responses, s.toResponse(user))
	}

	return responses, nil
}

func (s *UserService) toResponse(user *models.User) *models.UserResponse {
	return &models.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
