package services

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

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
	// Validate input
	if err := s.validateCreateUserRequest(req); err != nil {
		return nil, err
	}

	// Check if user already exists
	existing, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Check for different types of database errors
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "no rows") || strings.Contains(errMsg, "not found") {
			// User doesn't exist, which is fine - we can create them
			existing = nil
		} else {
			// Actual database error
			s.logger.WithError(err).Error("Failed to check existing user")
			return nil, models.NewInternalError("checking existing user", err)
		}
	}

	// If user exists, return conflict error
	if existing != nil {
		return nil, models.NewConflictError("User", "email", req.Email)
	}

	user := &models.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	created, err := s.repo.Create(ctx, user)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create user in repository")

		// Check if it's a constraint violation (duplicate key)
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") ||
			strings.Contains(err.Error(), "already exists") {
			return nil, models.NewConflictError("User", "email", req.Email)
		}

		return nil, models.NewInternalError("creating user", err)
	}

	return s.toResponse(created), nil
}

// validateCreateUserRequest validates the user creation request
func (s *UserService) validateCreateUserRequest(req *models.CreateUserRequest) error {
	// Validate email
	if req.Email == "" {
		return models.NewValidationError("email", "email is required")
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		return models.NewValidationError("email", "invalid email format")
	}

	// Validate first name
	if req.FirstName == "" {
		return models.NewValidationError("first_name", "first name is required")
	}

	if len(req.FirstName) < 2 {
		return models.NewValidationError("first_name", "first name must be at least 2 characters")
	}

	if len(req.FirstName) > 100 {
		return models.NewValidationError("first_name", "first name must be less than 100 characters")
	}

	// Validate last name
	if req.LastName == "" {
		return models.NewValidationError("last_name", "last name is required")
	}

	if len(req.LastName) < 2 {
		return models.NewValidationError("last_name", "last name must be at least 2 characters")
	}

	if len(req.LastName) > 100 {
		return models.NewValidationError("last_name", "last name must be less than 100 characters")
	}

	return nil
}

func (s *UserService) GetUser(ctx context.Context, id int) (*models.UserResponse, error) {
	if id <= 0 {
		return nil, models.NewValidationError("id", "user ID must be positive")
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user in service")

		// Check for different types of "not found" errors
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "no rows") || strings.Contains(errMsg, "not found") {
			return nil, models.NewNotFoundError("User", "id", fmt.Sprintf("%d", id))
		}

		return nil, models.NewInternalError("getting user", err)
	}

	return s.toResponse(user), nil
}

func (s *UserService) ReplaceUser(ctx context.Context, id int, req *models.ReplaceUserRequest) (*models.UserResponse, error) {
	if id <= 0 {
		return nil, models.NewValidationError("id", "user ID must be positive")
	}

	// Validate replace request (all fields required)
	if err := s.validateReplaceUserRequest(req); err != nil {
		return nil, err
	}

	// Check if user exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get existing user for replacement")

		if strings.Contains(err.Error(), "not found") {
			return nil, models.NewNotFoundError("User", "id", fmt.Sprintf("%d", id))
		}

		return nil, models.NewInternalError("getting user for replacement", err)
	}

	// Replace all fields (full replacement)
	user := &models.User{
		ID:        id,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		CreatedAt: existing.CreatedAt, // Preserve creation time
	}

	updated, err := s.repo.Update(ctx, id, user)
	if err != nil {
		s.logger.WithError(err).Error("Failed to replace user in repository")

		// Check for constraint violations
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			return nil, models.NewConflictError("User", "email", req.Email)
		}

		return nil, models.NewInternalError("replacing user", err)
	}

	return s.toResponse(updated), nil
}

func (s *UserService) UpdateUser(ctx context.Context, id int, req *models.UpdateUserRequest) (*models.UserResponse, error) {
	if id <= 0 {
		return nil, models.NewValidationError("id", "user ID must be positive")
	}

	// Validate update request
	if err := s.validateUpdateUserRequest(req); err != nil {
		return nil, err
	}

	// Get existing user
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get existing user for update")

		if strings.Contains(err.Error(), "not found") {
			return nil, models.NewNotFoundError("User", "id", fmt.Sprintf("%d", id))
		}

		return nil, models.NewInternalError("getting user for update", err)
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
		s.logger.WithError(err).Error("Failed to update user in repository")

		// Check for constraint violations
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			return nil, models.NewConflictError("User", "email", req.Email)
		}

		return nil, models.NewInternalError("updating user", err)
	}

	return s.toResponse(updated), nil
}

// validateReplaceUserRequest validates the user replace request (all fields required)
func (s *UserService) validateReplaceUserRequest(req *models.ReplaceUserRequest) error {
	// Validate email (required)
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return models.NewValidationError("email", "invalid email format")
	}

	// Validate first name (required)
	if len(req.FirstName) < 2 {
		return models.NewValidationError("first_name", "first name must be at least 2 characters")
	}
	if len(req.FirstName) > 100 {
		return models.NewValidationError("first_name", "first name must be less than 100 characters")
	}

	// Validate last name (required)
	if len(req.LastName) < 2 {
		return models.NewValidationError("last_name", "last name must be at least 2 characters")
	}
	if len(req.LastName) > 100 {
		return models.NewValidationError("last_name", "last name must be less than 100 characters")
	}

	return nil
}

// validateUpdateUserRequest validates the user update request
func (s *UserService) validateUpdateUserRequest(req *models.UpdateUserRequest) error {
	// At least one field must be provided
	if req.Email == "" && req.FirstName == "" && req.LastName == "" {
		return models.NewValidationError("request", "at least one field must be provided for update")
	}

	// Validate email if provided
	if req.Email != "" {
		if _, err := mail.ParseAddress(req.Email); err != nil {
			return models.NewValidationError("email", "invalid email format")
		}
	}

	// Validate first name if provided
	if req.FirstName != "" {
		if len(req.FirstName) < 2 {
			return models.NewValidationError("first_name", "first name must be at least 2 characters")
		}
		if len(req.FirstName) > 100 {
			return models.NewValidationError("first_name", "first name must be less than 100 characters")
		}
	}

	// Validate last name if provided
	if req.LastName != "" {
		if len(req.LastName) < 2 {
			return models.NewValidationError("last_name", "last name must be at least 2 characters")
		}
		if len(req.LastName) > 100 {
			return models.NewValidationError("last_name", "last name must be less than 100 characters")
		}
	}

	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, id int) error {
	if id <= 0 {
		return models.NewValidationError("id", "user ID must be positive")
	}

	err := s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.WithError(err).Error("Failed to delete user in repository")

		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "no rows") || strings.Contains(errMsg, "not found") {
			return models.NewNotFoundError("User", "id", fmt.Sprintf("%d", id))
		}

		return models.NewInternalError("deleting user", err)
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
