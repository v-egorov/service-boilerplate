# Service Patterns Reference

This document provides copy-pasteable code examples for building services in this boilerplate.

---

## 1. Model Layer

### Domain Model + TableName

```go
// internal/models/user.go
package models

import (
    "time"

    "github.com/google/uuid"
)

type User struct {
    ID           uuid.UUID `json:"id" db:"id"`
    Email        string    `json:"email" db:"email"`
    PasswordHash string    `json:"-" db:"password_hash"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

func (u *User) TableName() string {
    return "users"
}

// ToResponse converts domain model to API response
func (u *User) ToResponse() *UserResponse {
    return &UserResponse{
        ID:        u.ID,
        Email:     u.Email,
        CreatedAt: u.CreatedAt.Format(time.RFC3339),
        UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
    }
}
```

### Request DTOs

```go
// internal/models/user_request.go
package models

import "github.com/google/uuid"

type CreateUserRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

type UpdateUserRequest struct {
    Email    *string `json:"email,omitempty" binding:"omitempty,email"`
    Password *string `json:"password,omitempty" binding:"omitempty,min=8"`
}

type UserFilter struct {
    Email   string
    Limit   int
    Offset  int
    SortBy  string
    SortOrder string
}

type UserResponse struct {
    ID        uuid.UUID `json:"id"`
    Email     string    `json:"email"`
    CreatedAt string   `json:"created_at"`
    UpdatedAt string   `json:"updated_at"`
}
```

---

## 2. Repository Layer

### Interface Definition

```go
// internal/repository/user_repository.go
package repository

import (
    "context"

    "github.com/google/uuid"
    "github.com/v-egorov/service-boilerplate/services/user-service/internal/models"
)

type UserRepository interface {
    Repository // Base interface (DB, Options, Metrics, etc.)
    Create(ctx context.Context, user *models.User) (*models.User, error)
    GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    Update(ctx context.Context, id uuid.UUID, user *models.User) (*models.User, error)
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, filter *models.UserFilter) ([]*models.User, error)
    ExistsByEmail(ctx context.Context, email string) (bool, error)
}
```

### Implementation

```go
// internal/repository/user_repository.go
package repository

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/v-egorov/service-boilerplate/common/database"
    "github.com/v-egorov/service-boilerplate/services/user-service/internal/models"
)

type userRepository struct {
    db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
    return &userRepository{db: db}
}

// Always use schema-qualified table names
func (r *userRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
    query := `
        INSERT INTO app.users (id, email, password_hash, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), NOW())
        RETURNING id, email, password_hash, created_at, updated_at`

    // Use database tracing wrapper
    err := database.TraceDBInsert(ctx, "users", query, func(ctx context.Context) error {
        return r.db.QueryRow(ctx, query,
            user.ID, user.Email, user.PasswordHash,
        ).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    return user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
    query := `SELECT id, email, password_hash, created_at, updated_at FROM app.users WHERE id = $1`

    var user models.User
    err := database.TraceDBQuery(ctx, "users", query, func(ctx context.Context) error {
        return r.db.QueryRow(ctx, query, id).Scan(
            &user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
        )
    })
    if err == pgx.ErrNoRows {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    return &user, nil
}
```

---

## 3. Service Layer

### Interface + Implementation

```go
// internal/services/user_service.go
package services

import (
    "context"
    "errors"
    "fmt"

    "github.com/google/uuid"
    "github.com/v-egorov/service-boilerplate/services/user-service/internal/models"
    "github.com/v-egorov/service-boilerplate/services/user-service/internal/repository"
)

var (
    ErrUserNotFound = errors.New("user not found")
    ErrUserExists   = errors.New("user already exists")
    ErrInvalidInput = errors.New("invalid input")
)

type UserService interface {
    Create(ctx context.Context, req *models.CreateUserRequest) (*models.User, error)
    GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    Update(ctx context.Context, id uuid.UUID, req *models.UpdateUserRequest) (*models.User, error)
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, filter *models.UserFilter) ([]*models.User, error)
}

type userService struct {
    repo UserRepositoryInterface
}

func NewUserService(repo UserRepositoryInterface) UserService {
    return &userService{repo: repo}
}

func (s *userService) Create(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
    // Validation
    if req.Email == "" {
        return nil, fmt.Errorf("email is required: %w", ErrInvalidInput)
    }

    // Business logic
    exists, _ := s.repo.ExistsByEmail(ctx, req.Email)
    if exists {
        return nil, fmt.Errorf("user already exists: %w", ErrUserExists)
    }

    // Hash password (bcrypt in real implementation)
    user := &models.User{
        ID:        uuid.New(),
        Email:     req.Email,
        PasswordHash: "hashed_" + req.Password,
    }

    return s.repo.Create(ctx, user)
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
    if id == uuid.Nil {
        return nil, fmt.Errorf("invalid id: %w", ErrInvalidInput)
    }

    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, repository.ErrNotFound) {
            return nil, fmt.Errorf("user not found: %w", ErrUserNotFound)
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    return user, nil
}
```

---

## 4. Handler Layer

### Handler + Error Mapping

```go
// internal/handlers/user_handler.go
package handlers

import (
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/v-egorov/service-boilerplate/services/user-service/internal/models"
    "github.com/v-egorov/service-boilerplate/services/user-service/internal/services"
)

type UserHandler struct {
    service UserServiceInterface
    logger  *logrus.Logger
}

func NewUserHandler(service services.UserService, logger *logrus.Logger) *UserHandler {
    return &UserHandler{service: service, logger: logger}
}

func (h *UserHandler) Create(c *gin.Context) {
    requestID := c.GetHeader("X-Request-ID")

    var req models.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.WithField("request_id", requestID).WithError(err).Error("Invalid request")
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    user, err := h.service.Create(c.Request.Context(), &req)
    if err != nil {
        h.handleError(c, requestID, err, "create user")
        return
    }

    c.JSON(http.StatusCreated, user.ToResponse())
}

func (h *UserHandler) GetByID(c *gin.Context) {
    requestID := c.GetHeader("X-Request-ID")
    idStr := c.Param("id")

    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }

    user, err := h.service.GetByID(c.Request.Context(), id)
    if err != nil {
        h.handleError(c, requestID, err, "get user")
        return
    }

    c.JSON(http.StatusOK, user.ToResponse())
}

func (h *UserHandler) handleError(c *gin.Context, requestID string, err error, operation string) {
    h.logger.WithField("request_id", requestID).WithError(err).Error(operation)

    switch {
    case errors.Is(err, services.ErrUserNotFound):
        c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
    case errors.Is(err, services.ErrUserExists):
        c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
    case errors.Is(err, services.ErrInvalidInput):
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
    default:
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
    }
}
```

---

## 5. Testing Patterns

### Manual Mock Repository

```go
// internal/services/user_service_test.go
package services

import (
    "context"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/v-egorov/service-boilerplate/services/user-service/internal/models"
)

type mockUserRepository struct {
    createFunc        func(ctx context.Context, user *models.User) (*models.User, error)
    getByIDFunc       func(ctx context.Context, id uuid.UUID) (*models.User, error)
    getByEmailFunc    func(ctx context.Context, email string) (*models.User, error)
    updateFunc        func(ctx context.Context, id uuid.UUID, user *models.User) (*models.User, error)
    deleteFunc        func(ctx context.Context, id uuid.UUID) error
    listFunc          func(ctx context.Context, filter *models.UserFilter) ([]*models.User, error)
    existsByEmailFunc func(ctx context.Context, email string) (bool, error)
}

// Implement ALL repository interface methods
func (m *mockUserRepository) DB() interface{}                           { return nil }
func (m *mockUserRepository) Options() *repository.RepositoryOptions      { return nil }
func (m *mockUserRepository) Metrics() *repository.RepositoryMetrics    { return nil }
func (m *mockUserRepository) ResetMetrics()                             {}
func (m *mockUserRepository) Healthy(ctx context.Context) error         { return nil }

func (m *mockUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
    if m.createFunc != nil {
        return m.createFunc(ctx, user)
    }
    return user, nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
    if m.getByIDFunc != nil {
        return m.getByIDFunc(ctx, id)
    }
    return &models.User{ID: id}, nil
}

func (m *mockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
    if m.existsByEmailFunc != nil {
        return m.existsByEmailFunc(ctx, email)
    }
    return false, nil
}
```

### Test Function

```go
func TestUserService_Create_Success(t *testing.T) {
    mockRepo := &mockUserRepository{
        createFunc: func(ctx context.Context, user *models.User) (*models.User, error) {
            return user, nil
        },
        existsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
            return false, nil
        },
    }

    service := NewUserService(mockRepo)
    req := &models.CreateUserRequest{
        Email:    "test@example.com",
        Password: "password123",
    }

    result, err := service.Create(context.Background(), req)

    assert.NoError(t, err)
    assert.Equal(t, req.Email, result.Email)
}

func TestUserService_Create_DuplicateEmail(t *testing.T) {
    mockRepo := &mockUserRepository{
        existsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
            return true, nil
        },
    }

    service := NewUserService(mockRepo)
    req := &models.CreateUserRequest{
        Email:    "test@example.com",
        Password: "password123",
    }

    result, err := service.Create(context.Background(), req)

    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "already exists")
}
```

---

## 6. Main.go Integration

```go
// cmd/main.go
package main

import (
    "github.com/v-egorov/service-boilerplate/services/user-service/internal/handlers"
    "github.com/v-egorov/service-boilerplate/services/user-service/internal/repository"
    "github.com/v-egorov/service-boilerplate/services/user-service/internal/services"
)

func main() {
    // ... config and logger setup ...

    // Initialize repository
    userRepo := repository.NewUserRepository(db)

    // Initialize service
    userService := services.NewUserService(userRepo)

    // Initialize handler
    userHandler := handlers.NewUserHandler(userService, logger)

    // Register routes
    users := v1.Group("/users")
    users.POST("", userHandler.Create)
    users.GET("/:id", userHandler.GetByID)
}
```

---

## Common Components Usage

### Database with Tracing

```go
import "github.com/v-egorov/service-boilerplate/common/database"

// Wrap database operations
database.TraceDBQuery(ctx, "table_name", query, func(ctx context.Context) error {
    return r.db.QueryRow(ctx, query, args...).Scan(dest...)
})
```

### Logging

```go
import "github.com/v-egorov/service-boilerplate/common/logging"

standardLogger := logging.NewStandardLogger(logger, "service-name")
standardLogger.EntityOperation(requestID, userID, entityID, "create", true, nil)
```

### Middleware

```go
import "github.com/v-egorov/service-boilerplate/common/middleware"

// Development: trust gateway headers (jwtSecret = nil)
router.Use(middleware.JWTMiddleware(nil, logger, nil))
```
