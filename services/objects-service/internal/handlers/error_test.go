package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/services"

	"github.com/google/uuid"
)

// mockService is a minimal mock that implements ObjectServiceInterface for testing HandleError
type mockErrorHandlerService struct{}

func (m *mockErrorHandlerService) Create(ctx context.Context, req *models.CreateObjectRequest) (*models.Object, error) { return nil, assert.AnError }
func (m *mockErrorHandlerService) GetByID(ctx context.Context, id int64) (*models.Object, error)                        { return nil, assert.AnError }
func (m *mockErrorHandlerService) GetByPublicID(ctx context.Context, publicID uuid.UUID) (*models.Object, error)        { return nil, assert.AnError }
func (m *mockErrorHandlerService) GetByName(ctx context.Context, name string) (*models.Object, error)                   { return nil, assert.AnError }
func (m *mockErrorHandlerService) Update(ctx context.Context, id int64, req *models.UpdateObjectRequest) (*models.Object, error) {
	return nil, assert.AnError
}
func (m *mockErrorHandlerService) Delete(ctx context.Context, id int64) error              { return assert.AnError }
func (m *mockErrorHandlerService) List(ctx context.Context, filter *models.ObjectFilter) ([]*models.Object, int64, error) {
	return nil, 0, assert.AnError
}
func (m *mockErrorHandlerService) Search(ctx context.Context, query string, limit int) ([]*models.Object, error)       { return nil, assert.AnError }
func (m *mockErrorHandlerService) FindByMetadata(ctx context.Context, key, value string) ([]*models.Object, error)     { return nil, assert.AnError }
func (m *mockErrorHandlerService) FindByTags(ctx context.Context, tags []string, matchAll bool) ([]*models.Object, error) {
	return nil, assert.AnError
}
func (m *mockErrorHandlerService) UpdateMetadata(ctx context.Context, id int64, metadata map[string]interface{}, updatedBy string) error {
	return assert.AnError
}
func (m *mockErrorHandlerService) AddTags(ctx context.Context, id int64, tags []string, updatedBy string) error       { return assert.AnError }
func (m *mockErrorHandlerService) RemoveTags(ctx context.Context, id int64, tags []string, updatedBy string) error    { return assert.AnError }
func (m *mockErrorHandlerService) GetChildren(ctx context.Context, parentID int64) ([]*models.Object, error)          { return nil, assert.AnError }
func (m *mockErrorHandlerService) GetDescendants(ctx context.Context, rootID int64, maxDepth *int) ([]*models.Object, error) {
	return nil, assert.AnError
}
func (m *mockErrorHandlerService) GetAncestors(ctx context.Context, id int64) ([]*models.Object, error)               { return nil, assert.AnError }
func (m *mockErrorHandlerService) GetPath(ctx context.Context, id int64) ([]*models.Object, error)                    { return nil, assert.AnError }
func (m *mockErrorHandlerService) BulkCreate(ctx context.Context, objects []*models.CreateObjectRequest) ([]*models.Object, error) {
	return nil, assert.AnError
}
func (m *mockErrorHandlerService) BulkUpdate(ctx context.Context, ids []int64, updates *models.UpdateObjectRequest) ([]*models.Object, error) {
	return nil, assert.AnError
}
func (m *mockErrorHandlerService) BulkDelete(ctx context.Context, ids []int64) error                          { return assert.AnError }
func (m *mockErrorHandlerService) ValidateParentChild(ctx context.Context, parentID, childID int64) error     { return assert.AnError }
func (m *mockErrorHandlerService) GetObjectStats(ctx context.Context, filter *models.ObjectFilter) (*repository.ObjectStats, error) {
	return nil, assert.AnError
}

// mockRelTypeService for relationship type handler tests
type mockRelTypeHandlerService struct{}

func (m *mockRelTypeHandlerService) Create(ctx context.Context, req *models.CreateRelationshipTypeRequest) (*models.RelationshipType, error) {
	return nil, assert.AnError
}
func (m *mockRelTypeHandlerService) GetByTypeKey(ctx context.Context, typeKey string) (*models.RelationshipType, error) { return nil, assert.AnError }
func (m *mockRelTypeHandlerService) Update(ctx context.Context, typeKey string, req *models.UpdateRelationshipTypeRequest) (*models.RelationshipType, error) {
	return nil, assert.AnError
}
func (m *mockRelTypeHandlerService) Delete(ctx context.Context, typeKey string) error                          { return assert.AnError }
func (m *mockRelTypeHandlerService) List(ctx context.Context, filter *models.RelationshipTypeFilter) ([]*models.RelationshipType, error) {
	return nil, assert.AnError
}

func TestHandleError_NilIsNoOp(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// Should not write anything
	HandleError(c, nil, "test-id")
	assert.Equal(t, http.StatusOK, w.Code, "nil error should produce no response body")
}

func TestHandleError_ErrNotFound_404(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		status   int
		errorTyp string
	}{
		{"repo ErrNotFound", repository.ErrNotFound, http.StatusNotFound, "not_found"},
		{"service ErrObjectNotFound (wrapped)", fmt.Errorf("object not found: %w", repository.ErrNotFound), http.StatusNotFound, "not_found"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			HandleError(c, tc.err, "req-123")

			assert.Equal(t, tc.status, w.Code, "status code mismatch for %v", tc.name)
			var body map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &body)
			assert.Equal(t, tc.errorTyp, body["type"], "error type mismatch for %v", tc.name)
		})
	}
}

func TestHandleError_ErrConflict_409(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{"repo ErrOptimisticLock", repository.ErrOptimisticLock},
		{"repo ErrVersionConflict", repository.ErrVersionConflict},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			HandleError(c, tc.err, "req-123")

			assert.Equal(t, http.StatusConflict, w.Code, "status code mismatch for %v", tc.name)
			var body map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &body)
			assert.Equal(t, "conflict", body["type"], "error type mismatch for %v", tc.name)
		})
	}
}

func TestHandleError_ErrValidation_422(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	err := services.ErrCircularRelationship
	HandleError(c, err, "req-123")

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, "validation_error", body["type"])
}

func TestHandleError_ErrBadRequest_400(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{"repo ErrInvalidInput", repository.ErrInvalidInput},
		{"service ErrTypeKeyRequired", services.ErrTypeKeyRequired},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			HandleError(c, tc.err, "req-123")

			assert.Equal(t, http.StatusBadRequest, w.Code, "status code mismatch for %v", tc.name)
			var body map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &body)
			assert.Equal(t, "validation_error", body["type"], "error type mismatch for %v", tc.name)
		})
	}
}

func TestHandleError_UnknownErr_500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	unknownErr := fmt.Errorf("some unexpected error")
	HandleError(c, unknownErr, "req-123")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, "internal_error", body["type"])
}

func TestHandleError_RequestsAllSentinels(t *testing.T) {
	// Exhaustive test: every sentinel error must have a registered case.
	// This is the regression guard — if a new sentinel is added and not handled here, it will fail at runtime with 500.
	sentinels := []struct {
		name   string
		err    error
		status int
	}{
		// Repository sentinels
		{"repo:ErrNotFound", repository.ErrNotFound, http.StatusNotFound},
		{"repo:ErrInvalidInput", repository.ErrInvalidInput, http.StatusBadRequest},
		{"repo:ErrOptimisticLock", repository.ErrOptimisticLock, http.StatusConflict},
		{"repo:ErrVersionConflict", repository.ErrVersionConflict, http.StatusConflict},

		// Relationship service sentinels
		{"rel:ErrRelationshipNotFound", services.ErrRelationshipNotFound, http.StatusNotFound},
		{"rel:ErrDuplicateRelationship", services.ErrDuplicateRelationship, http.StatusConflict},
		{"rel:ErrSourceObjectNotFound", services.ErrSourceObjectNotFound, http.StatusNotFound},
		{"rel:ErrTargetObjectNotFound", services.ErrTargetObjectNotFound, http.StatusNotFound},
		{"rel:ErrCircularRelationship", services.ErrCircularRelationship, http.StatusUnprocessableEntity},
		{"rel:ErrCardinalityViolation", services.ErrCardinalityViolation, http.StatusUnprocessableEntity},
		{"rel:ErrSourceTargetSame", services.ErrSourceTargetSame, http.StatusBadRequest},

		// RelationshipType service sentinels
		{"reltype:ErrRelationshipTypeNotFound", services.ErrRelationshipTypeNotFound, http.StatusNotFound},
		{"reltype:ErrDuplicateRelationshipType", services.ErrDuplicateRelationshipType, http.StatusConflict},
		{"reltype:ErrInvalidCardinality", services.ErrInvalidCardinality, http.StatusUnprocessableEntity},
		{"reltype:ErrInvalidReverseType", services.ErrInvalidReverseType, http.StatusUnprocessableEntity},
		{"reltype:ErrRelationshipTypeInUse", services.ErrRelationshipTypeInUse, http.StatusConflict},
		{"reltype:ErrInvalidCountConstraint", services.ErrInvalidCountConstraint, http.StatusBadRequest},
		{"reltype:ErrTypeKeyRequired", services.ErrTypeKeyRequired, http.StatusBadRequest},
		{"reltype:ErrCardinalityRequired", services.ErrCardinalityRequired, http.StatusBadRequest},
	}

	for _, s := range sentinels {
		t.Run(s.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			HandleError(c, s.err, "req-id")

			assert.Equal(t, s.status, w.Code, "%s: expected HTTP %d but got %d", s.name, s.status, w.Code)
			var body map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &body)
			// Verify meta.request_id is always present
			meta, ok := body["meta"].(map[string]interface{})
			assert.True(t, ok, "%s: expected 'meta' object", s.name)
			assert.Equal(t, "req-id", meta["request_id"], "%s: request_id mismatch", s.name)
		})
	}
}
