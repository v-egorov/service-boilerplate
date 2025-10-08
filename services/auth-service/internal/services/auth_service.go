package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/client"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/repository"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo       *repository.AuthRepository
	userClient *client.UserClient
	jwtUtils   *utils.JWTUtils
	logger     *logrus.Logger
}

func NewAuthService(repo *repository.AuthRepository, userClient *client.UserClient, jwtUtils *utils.JWTUtils, logger *logrus.Logger) *AuthService {
	return &AuthService{
		repo:       repo,
		userClient: userClient,
		jwtUtils:   jwtUtils,
		logger:     logger,
	}
}

func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest, ipAddress, userAgent string) (*models.TokenResponse, error) {
	tracer := otel.Tracer("auth-service")

	ctx, span := tracer.Start(ctx, "auth.login",
		trace.WithAttributes(
			attribute.String("user.email", req.Email),
			attribute.String("client.ip", ipAddress),
			attribute.String("auth.operation", "login"),
		))
	defer span.End()

	s.logger.WithFields(logrus.Fields{
		"email": req.Email,
		"ip":    ipAddress,
	}).Info("Login attempt")

	// Call user service to get user with password hash
	userLogin, err := s.userClient.GetUserWithPasswordByEmail(ctx, req.Email)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get user from user service")
		s.logger.WithError(err).Error("Failed to get user from user service")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(userLogin.PasswordHash), []byte(req.Password)); err != nil {
		s.logger.WithField("email", req.Email).Warn("Invalid password")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Use the actual user ID from user service
	userID := userLogin.User.ID
	email := userLogin.User.Email

	// Get user roles
	roles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user roles")
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	// Generate tokens
	accessToken, err := s.jwtUtils.GenerateAccessToken(userID, email, roleNames, 15*time.Minute)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate access token")
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtUtils.GenerateRefreshToken(userID, 7*24*time.Hour)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate refresh token")
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Hash and store tokens
	accessTokenHash := s.hashToken(accessToken)
	refreshTokenHash := s.hashToken(refreshToken)

	// Store access token
	accessTokenModel := &models.AuthToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: accessTokenHash,
		TokenType: "access",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	if err := s.repo.CreateAuthToken(ctx, accessTokenModel); err != nil {
		s.logger.WithError(err).Error("Failed to store access token")
		return nil, fmt.Errorf("failed to store access token: %w", err)
	}

	// Store refresh token
	refreshTokenModel := &models.AuthToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: refreshTokenHash,
		TokenType: "refresh",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.repo.CreateAuthToken(ctx, refreshTokenModel); err != nil {
		s.logger.WithError(err).Error("Failed to store refresh token")
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Create session
	session := &models.UserSession{
		ID:           uuid.New(),
		UserID:       userID,
		SessionToken: s.hashToken(uuid.New().String()),
		IPAddress:    &ipAddress,
		UserAgent:    &userAgent,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.repo.CreateUserSession(ctx, session); err != nil {
		s.logger.WithError(err).Error("Failed to create session")
		// Don't fail the login if session creation fails
	}

	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("user.email", email),
		attribute.Int("auth.tokens_created", 2), // access + refresh
		attribute.Bool("auth.session_created", true),
	)
	span.SetStatus(codes.Ok, "Login successful")

	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"email":   email,
	}).Info("Login successful")

	return &models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900, // 15 minutes
		User: models.UserInfo{
			ID:        userID,
			Email:     email,
			FirstName: userLogin.User.FirstName,
			LastName:  userLogin.User.LastName,
			Roles:     roleNames,
		},
	}, nil
}

func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.UserInfo, error) {
	s.logger.WithField("email", req.Email).Info("Registration attempt")

	// Call user service to create user
	createUserReq := &client.CreateUserRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	userData, err := s.userClient.CreateUser(ctx, createUserReq)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create user in user service")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Use the actual user ID from user service
	userID := userData.ID

	// Assign default role
	defaultRole, err := s.repo.GetRoleByName(ctx, "user")
	if err != nil {
		s.logger.WithError(err).Error("Failed to get default role")
		return nil, fmt.Errorf("failed to get default role: %w", err)
	}

	if err := s.repo.AssignRoleToUser(ctx, userID, defaultRole.ID); err != nil {
		s.logger.WithError(err).Error("Failed to assign default role")
		return nil, fmt.Errorf("failed to assign default role: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"email":   req.Email,
	}).Info("Registration successful")

	return &models.UserInfo{
		ID:        userID,
		Email:     userData.Email,
		FirstName: userData.FirstName,
		LastName:  userData.LastName,
		Roles:     []string{"user"},
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, tokenString string) error {
	tracer := otel.Tracer("auth-service")

	ctx, span := tracer.Start(ctx, "auth.logout",
		trace.WithAttributes(
			attribute.String("auth.operation", "logout"),
		))
	defer span.End()

	claims, err := s.jwtUtils.ValidateToken(tokenString)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid token during logout")
		return fmt.Errorf("invalid token: %w", err)
	}

	span.SetAttributes(
		attribute.String("user.id", claims.UserID.String()),
		attribute.String("token.type", claims.TokenType),
	)

	// Get token from database
	token, err := s.repo.GetAuthTokenByHash(ctx, s.hashToken(tokenString))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Token not found in database during logout")
		s.logger.WithError(err).Warn("Token not found in database during logout")
		return fmt.Errorf("token not found: %w", err)
	}

	// Revoke token
	if err := s.repo.RevokeAuthToken(ctx, token.ID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to revoke token")
		s.logger.WithError(err).Error("Failed to revoke token")
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	span.SetAttributes(
		attribute.Bool("token.revoked", true),
		attribute.String("token.id", token.ID.String()),
	)
	span.SetStatus(codes.Ok, "Logout successful")

	s.logger.WithField("user_id", claims.UserID).Info("Logout successful")
	return nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *models.RefreshTokenRequest) (*models.TokenResponse, error) {
	tracer := otel.Tracer("auth-service")

	ctx, span := tracer.Start(ctx, "auth.refresh_token",
		trace.WithAttributes(
			attribute.String("auth.operation", "refresh"),
		))
	defer span.End()

	s.logger.Info("Token refresh attempt")

	// Validate refresh token
	claims, err := s.jwtUtils.ValidateToken(req.RefreshToken)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid refresh token")
		s.logger.WithError(err).Warn("Invalid refresh token")
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	span.SetAttributes(
		attribute.String("user.id", claims.UserID.String()),
		attribute.String("token.type", claims.TokenType),
	)

	if claims.TokenType != "refresh" {
		s.logger.Warn("Non-refresh token used for refresh")
		return nil, fmt.Errorf("invalid token type for refresh")
	}

	// Get token from database
	token, err := s.repo.GetAuthTokenByHash(ctx, s.hashToken(req.RefreshToken))
	if err != nil {
		s.logger.WithError(err).Warn("Refresh token not found in database")
		return nil, fmt.Errorf("refresh token not found: %w", err)
	}

	// Revoke old refresh token
	if err := s.repo.RevokeAuthToken(ctx, token.ID); err != nil {
		s.logger.WithError(err).Error("Failed to revoke old refresh token")
		return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
	}

	// Get user roles
	roles, err := s.repo.GetUserRoles(ctx, claims.UserID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user roles")
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	// Generate new tokens
	accessToken, err := s.jwtUtils.GenerateAccessToken(claims.UserID, claims.Email, roleNames, 15*time.Minute)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate new access token")
		return nil, fmt.Errorf("failed to generate new access token: %w", err)
	}

	refreshToken, err := s.jwtUtils.GenerateRefreshToken(claims.UserID, 7*24*time.Hour)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate new refresh token")
		return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	// Store new tokens
	accessTokenHash := s.hashToken(accessToken)
	refreshTokenHash := s.hashToken(refreshToken)

	newAccessToken := &models.AuthToken{
		ID:        uuid.New(),
		UserID:    claims.UserID,
		TokenHash: accessTokenHash,
		TokenType: "access",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	if err := s.repo.CreateAuthToken(ctx, newAccessToken); err != nil {
		s.logger.WithError(err).Error("Failed to store new access token")
		return nil, fmt.Errorf("failed to store new access token: %w", err)
	}

	newRefreshToken := &models.AuthToken{
		ID:        uuid.New(),
		UserID:    claims.UserID,
		TokenHash: refreshTokenHash,
		TokenType: "refresh",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.repo.CreateAuthToken(ctx, newRefreshToken); err != nil {
		s.logger.WithError(err).Error("Failed to store new refresh token")
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	span.SetAttributes(
		attribute.Int("auth.tokens_created", 2), // new access + refresh
		attribute.Int("auth.tokens_revoked", 1), // old refresh token
	)
	span.SetStatus(codes.Ok, "Token refresh successful")

	s.logger.WithField("user_id", claims.UserID).Info("Token refresh successful")

	return &models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900,
		User: models.UserInfo{
			ID:    claims.UserID,
			Email: claims.Email,
			Roles: roleNames,
		},
	}, nil
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID, email string) (*models.UserInfo, error) {
	// Call user service to get user details by email
	userData, err := s.userClient.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user from user service")
		// Fallback to basic info if user service call fails
		s.logger.Warn("Falling back to basic user info due to user service error")
	}

	// Get user roles
	roles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user roles")
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	userInfo := &models.UserInfo{
		ID:    userID,
		Email: email,
		Roles: roleNames,
	}

	// Populate user details from user service if available
	if userData != nil {
		userInfo.FirstName = userData.FirstName
		userInfo.LastName = userData.LastName
	}

	return userInfo, nil
}

func (s *AuthService) GetPublicKeyPEM() ([]byte, error) {
	return s.jwtUtils.GetPublicKeyPEM()
}

func (s *AuthService) RotateKeys(ctx context.Context) error {
	return s.RotateKeysWithReason(ctx, "manual")
}

func (s *AuthService) RotateKeysWithReason(ctx context.Context, reason string) error {
	s.logger.WithField("reason", reason).Info("Starting JWT key rotation")

	if err := s.jwtUtils.RotateKeys(ctx); err != nil {
		s.logger.WithError(err).Error("Failed to rotate JWT keys")
		return fmt.Errorf("failed to rotate JWT keys: %w", err)
	}

	s.logger.WithField("reason", reason).Info("JWT key rotation completed successfully")
	return nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*utils.JWTClaims, error) {
	tracer := otel.Tracer("auth-service")

	ctx, span := tracer.Start(ctx, "auth.validate_token",
		trace.WithAttributes(
			attribute.String("auth.operation", "validate"),
			attribute.Bool("token.requires_db_check", true),
		))
	defer span.End()

	claims, err := s.jwtUtils.ValidateToken(tokenString)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "JWT validation failed")
		return nil, err
	}

	span.SetAttributes(
		attribute.String("user.id", claims.UserID.String()),
		attribute.String("token.type", claims.TokenType),
	)

	// Check if token exists in database and is not revoked
	token, err := s.repo.GetAuthTokenByHash(ctx, s.hashToken(tokenString))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Token not found in database")
		return nil, fmt.Errorf("token not found in database: %w", err)
	}

	if token.RevokedAt != nil {
		span.SetAttributes(
			attribute.Bool("token.is_revoked", true),
			attribute.String("token.revoked_at", token.RevokedAt.String()),
		)
		span.SetStatus(codes.Error, "Token has been revoked")
		return nil, fmt.Errorf("token has been revoked")
	}

	span.SetAttributes(
		attribute.Bool("token.is_revoked", false),
		attribute.String("token.expires_at", token.ExpiresAt.String()),
	)
	span.SetStatus(codes.Ok, "Token validation successful")

	return claims, nil
}

func (s *AuthService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}

// GetUserRoles returns roles for a user (already exists but making sure it's accessible)
func (s *AuthService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error) {
	return s.repo.GetUserRoles(ctx, userID)
}

// Role Management Service Methods
func (s *AuthService) CreateRole(ctx context.Context, name, description string) (*models.Role, error) {
	var descPtr *string
	if description != "" {
		descPtr = &description
	}

	role := &models.Role{
		Name:        name,
		Description: descPtr,
	}

	createdRole, err := s.repo.CreateRole(ctx, role)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create role")
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"role_id":   createdRole.ID,
		"role_name": createdRole.Name,
	}).Info("Role created successfully")

	return createdRole, nil
}

func (s *AuthService) ListRoles(ctx context.Context) ([]models.Role, error) {
	roles, err := s.repo.ListRoles(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list roles")
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	return roles, nil
}

func (s *AuthService) GetRole(ctx context.Context, roleID uuid.UUID) (*models.Role, error) {
	role, err := s.repo.GetRole(ctx, roleID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get role")
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return role, nil
}

func (s *AuthService) UpdateRole(ctx context.Context, roleID uuid.UUID, name, description string) (*models.Role, error) {
	updatedRole, err := s.repo.UpdateRole(ctx, roleID, name, description)
	if err != nil {
		s.logger.WithError(err).Error("Failed to update role")
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"role_id":   updatedRole.ID,
		"role_name": updatedRole.Name,
	}).Info("Role updated successfully")

	return updatedRole, nil
}

func (s *AuthService) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	// Check if role is in use
	userCount, err := s.repo.CountUsersWithRole(ctx, roleID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to check role usage")
		return fmt.Errorf("failed to check role usage: %w", err)
	}

	if userCount > 0 {
		return fmt.Errorf("cannot delete role: %d users are assigned to this role", userCount)
	}

	err = s.repo.DeleteRole(ctx, roleID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to delete role")
		return fmt.Errorf("failed to delete role: %w", err)
	}

	s.logger.WithField("role_id", roleID).Info("Role deleted successfully")
	return nil
}

// Permission Management Service Methods
func (s *AuthService) CreatePermission(ctx context.Context, name, resource, action string) (*models.Permission, error) {
	permission := &models.Permission{
		Name:     name,
		Resource: resource,
		Action:   action,
	}

	createdPermission, err := s.repo.CreatePermission(ctx, permission)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create permission")
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"permission_id":   createdPermission.ID,
		"permission_name": createdPermission.Name,
	}).Info("Permission created successfully")

	return createdPermission, nil
}

func (s *AuthService) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	permissions, err := s.repo.ListPermissions(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list permissions")
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	return permissions, nil
}

func (s *AuthService) GetPermission(ctx context.Context, permissionID uuid.UUID) (*models.Permission, error) {
	permission, err := s.repo.GetPermission(ctx, permissionID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get permission")
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return permission, nil
}

func (s *AuthService) UpdatePermission(ctx context.Context, permissionID uuid.UUID, name, resource, action string) (*models.Permission, error) {
	updatedPermission, err := s.repo.UpdatePermission(ctx, permissionID, name, resource, action)
	if err != nil {
		s.logger.WithError(err).Error("Failed to update permission")
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"permission_id":   updatedPermission.ID,
		"permission_name": updatedPermission.Name,
	}).Info("Permission updated successfully")

	return updatedPermission, nil
}

func (s *AuthService) DeletePermission(ctx context.Context, permissionID uuid.UUID) error {
	// Check if permission is in use
	roleCount, err := s.repo.CountRolesWithPermission(ctx, permissionID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to check permission usage")
		return fmt.Errorf("failed to check permission usage: %w", err)
	}

	if roleCount > 0 {
		return fmt.Errorf("cannot delete permission: %d roles are assigned this permission", roleCount)
	}

	err = s.repo.DeletePermission(ctx, permissionID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to delete permission")
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	s.logger.WithField("permission_id", permissionID).Info("Permission deleted successfully")
	return nil
}

// Role-Permission Management Service Methods
func (s *AuthService) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	err := s.repo.AssignPermissionToRole(ctx, roleID, permissionID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to assign permission to role")
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"role_id":       roleID,
		"permission_id": permissionID,
	}).Info("Permission assigned to role successfully")

	return nil
}

func (s *AuthService) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	err := s.repo.RemovePermissionFromRole(ctx, roleID, permissionID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to remove permission from role")
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"role_id":       roleID,
		"permission_id": permissionID,
	}).Info("Permission removed from role successfully")

	return nil
}

func (s *AuthService) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]models.Permission, error) {
	permissions, err := s.repo.GetRolePermissions(ctx, roleID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get role permissions")
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	return permissions, nil
}

// User Role Management Service Methods
func (s *AuthService) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	err := s.repo.AssignRoleToUser(ctx, userID, roleID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to assign role to user")
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"role_id": roleID,
	}).Info("Role assigned to user successfully")

	return nil
}

func (s *AuthService) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	err := s.repo.RemoveRoleFromUser(ctx, userID, roleID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to remove role from user")
		return fmt.Errorf("failed to remove role from user: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"role_id": roleID,
	}).Info("Role removed from user successfully")

	return nil
}

func (s *AuthService) UpdateUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	err := s.repo.UpdateUserRoles(ctx, userID, roleIDs)
	if err != nil {
		s.logger.WithError(err).Error("Failed to update user roles")
		return fmt.Errorf("failed to update user roles: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"role_count": len(roleIDs),
	}).Info("User roles updated successfully")

	return nil
}
