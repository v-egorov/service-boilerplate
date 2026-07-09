package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/services"
)

// HandleError dispatches service-layer and repository-layer sentinel errors to
// appropriate HTTP status codes following API Response Standards.
// It writes the JSON response directly; callers should log before calling it.
func HandleError(c *gin.Context, err error, requestID string) {
	if err == nil {
		return
	}

	statusCode := http.StatusInternalServerError
	errorMessage := "Internal server error"
	errorType := "internal_error"

	switch {
	// ── Object service sentinels (wraps repository.ErrNotFound / ErrInvalidInput) ──

	// ── Relationship service sentinels ──
	case errors.Is(err, services.ErrRelationshipNotFound):
		statusCode = http.StatusNotFound
		errorMessage = "Relationship not found"
		errorType = "not_found"
	case errors.Is(err, services.ErrDuplicateRelationship):
		statusCode = http.StatusConflict
		errorMessage = "Relationship already exists"
		errorType = "conflict"
	case errors.Is(err, services.ErrSourceObjectNotFound):
		statusCode = http.StatusNotFound
		errorMessage = "Source object not found"
		errorType = "not_found"
	case errors.Is(err, services.ErrTargetObjectNotFound):
		statusCode = http.StatusNotFound
		errorMessage = "Target object not found"
		errorType = "not_found"
	case errors.Is(err, services.ErrCircularRelationship):
		statusCode = http.StatusUnprocessableEntity
		errorMessage = "Cannot create circular relationship"
		errorType = "validation_error"
	case errors.Is(err, services.ErrCardinalityViolation):
		statusCode = http.StatusUnprocessableEntity
		errorMessage = "Relationship cardinality constraint violated"
		errorType = "validation_error"
	case errors.Is(err, services.ErrSourceTargetSame):
		statusCode = http.StatusBadRequest
		errorMessage = "Source and target cannot be the same"
		errorType = "validation_error"

	// ── Relationship type service sentinels ──
	case errors.Is(err, services.ErrRelationshipTypeNotFound):
		statusCode = http.StatusNotFound
		errorMessage = "Relationship type not found"
		errorType = "not_found"
	case errors.Is(err, services.ErrDuplicateRelationshipType):
		statusCode = http.StatusConflict
		errorMessage = "Relationship type already exists"
		errorType = "conflict"
	case errors.Is(err, services.ErrInvalidCardinality):
		statusCode = http.StatusUnprocessableEntity
		errorMessage = "Invalid cardinality value"
		errorType = "validation_error"
	case errors.Is(err, services.ErrInvalidReverseType):
		statusCode = http.StatusUnprocessableEntity
		errorMessage = "Reverse type key does not exist"
		errorType = "validation_error"
	case errors.Is(err, services.ErrRelationshipTypeInUse):
		statusCode = http.StatusConflict
		errorMessage = "Relationship type is in use and cannot be deleted"
		errorType = "conflict"
	case errors.Is(err, services.ErrInvalidCountConstraint):
		statusCode = http.StatusBadRequest
		errorMessage = "min_count cannot exceed max_count"
		errorType = "validation_error"
	case errors.Is(err, services.ErrTypeKeyRequired):
		statusCode = http.StatusBadRequest
		errorMessage = "type_key is required"
		errorType = "validation_error"
	case errors.Is(err, services.ErrCardinalityRequired):
		statusCode = http.StatusBadRequest
		errorMessage = "cardinality is required"
		errorType = "validation_error"

	// ── Repository-level sentinels (catch-all for object/object_type wrappers) ──
	case errors.Is(err, repository.ErrOptimisticLock),
		errors.Is(err, repository.ErrVersionConflict):
		statusCode = http.StatusConflict
		errorMessage = err.Error()
		errorType = "conflict"

	case errors.Is(err, repository.ErrNotFound):
		statusCode = http.StatusNotFound
		errorMessage = err.Error()
		errorType = "not_found"

	case errors.Is(err, repository.ErrInvalidInput):
		statusCode = http.StatusBadRequest
		errorMessage = err.Error()
		errorType = "validation_error"

	default:
		// Unknown error — fall back to 500
	}

	c.JSON(statusCode, gin.H{
		"error": errorMessage,
		"type":  errorType,
		"meta":  gin.H{"request_id": requestID},
	})
}
