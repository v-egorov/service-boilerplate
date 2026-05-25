package utils

import (
	"strings"

	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/models"
)

// PermissionSpec represents the parsed components of a permission name.
type PermissionSpec struct {
	Resource string // e.g., "relationships", "objects"
	Action   string // e.g., "read", "create"
	Scope    string // "" for flat, "own"/"all" for scoped
}

// ParsePermission splits a permission name into its components.
// Valid formats: "objects:create" (2 parts = flat), "relationships:read:own" (3 parts = scoped)
func ParsePermission(name string) (PermissionSpec, error) {
	if name == "" {
		return PermissionSpec{}, models.PermissionParseError{
			Name:   name,
			Reason: "permission name is empty",
		}
	}

	parts := strings.SplitN(name, ":", 3)

	if len(parts) < 2 || len(parts) > 3 {
		return PermissionSpec{}, models.PermissionParseError{
			Name:   name,
			Reason: "expected 'resource:action' or 'resource:action:scope'",
		}
	}

	resource := parts[0]
	action := parts[1]

	if resource == "" || action == "" {
		return PermissionSpec{}, models.PermissionParseError{
			Name:   name,
			Reason: "permission name contains empty field",
		}
	}

	scope := ""
	if len(parts) == 3 {
		scope = parts[2]
	}

	return PermissionSpec{
		Resource: resource,
		Action:   action,
		Scope:    scope,
	}, nil
}

// ValidatePermission checks business rules after parsing.
func ValidatePermission(spec PermissionSpec) error {
	if spec.Resource == "" {
		return models.ValidationError{Field: "resource", Message: "must not be empty"}
	}

	if spec.Action == "" {
		return models.ValidationError{Field: "action", Message: "must not be empty"}
	}

	if spec.Scope != "" && spec.Scope != "own" && spec.Scope != "all" {
		return models.ValidationError{
			Field:   "scope",
			Message: "invalid scope value; must be 'own', 'all', or empty",
		}
	}

	if spec.Resource == "relationships" && spec.Scope == "" {
		return models.ValidationError{
			Field:   "scope",
			Message: "relationship permissions always require a scope suffix (e.g., relationships:" + spec.Action + ":own)",
		}
	}

	return nil
}

// GetResourceAction extracts the resource+action pair from a permission name.
func GetResourceAction(name string) (string, string, error) {
	spec, err := ParsePermission(name)
	if err != nil {
		return "", "", models.PermissionParseError{
			Name:   name,
			Reason: "failed to extract resource/action",
		}
	}
	return spec.Resource, spec.Action, nil
}
