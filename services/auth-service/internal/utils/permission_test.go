package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePermission(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    PermissionSpec
		expectError bool
	}{
		{
			name:  "flat permission - objects:create",
			input: "objects:create",
			expected: PermissionSpec{
				Resource: "objects",
				Action:   "create",
				Scope:    "",
			},
			expectError: false,
		},
		{
			name:  "flat permission - object-types:delete",
			input: "object-types:delete",
			expected: PermissionSpec{
				Resource: "object-types",
				Action:   "delete",
				Scope:    "",
			},
			expectError: false,
		},
		{
			name:  "scoped permission - relationships:read:own",
			input: "relationships:read:own",
			expected: PermissionSpec{
				Resource: "relationships",
				Action:   "read",
				Scope:    "own",
			},
			expectError: false,
		},
		{
			name:  "scoped permission - relationships:create:all",
			input: "relationships:create:all",
			expected: PermissionSpec{
				Resource: "relationships",
				Action:   "create",
				Scope:    "all",
			},
			expectError: false,
		},
		{
			name:  "scoped permission - relationships:update:own",
			input: "relationships:update:own",
			expected: PermissionSpec{
				Resource: "relationships",
				Action:   "update",
				Scope:    "own",
			},
			expectError: false,
		},
		{
			name:  "empty string",
			input: "",
			expected: PermissionSpec{},
			expectError: true,
		},
		{
			name:  "single part only",
			input: "onlyone",
			expected: PermissionSpec{},
			expectError: true,
		},
		{
			name:  "four parts - parsed as scoped with invalid scope value (validated separately)",
			input: "a:b:c:d",
			expected: PermissionSpec{
				Resource: "a",
				Action:   "b",
				Scope:    "c:d",
			},
			expectError: false,
		},
		{
			name:  "empty resource between colons",
			input: ":action",
			expected: PermissionSpec{},
			expectError: true,
		},
		{
			name:  "empty action between colons",
			input: "resource:",
			expected: PermissionSpec{},
			expectError: true,
		},
		{
			name:  "only colon separator",
			input: ":",
			expected: PermissionSpec{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePermission(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Resource, result.Resource)
				assert.Equal(t, tt.expected.Action, result.Action)
				assert.Equal(t, tt.expected.Scope, result.Scope)
			}
		})
	}
}

func TestValidatePermission(t *testing.T) {
	tests := []struct {
		name        string
		spec        PermissionSpec
		expectError bool
	}{
		{
			name: "valid flat objects permission",
			spec: PermissionSpec{Resource: "objects", Action: "create", Scope: ""},
			expectError: false,
		},
		{
			name: "valid flat object-types permission",
			spec: PermissionSpec{Resource: "object-types", Action: "read", Scope: ""},
			expectError: false,
		},
		{
			name:  "invalid - empty resource",
			spec:  PermissionSpec{Resource: "", Action: "create", Scope: ""},
			expectError: true,
		},
		{
			name:  "invalid - empty action",
			spec:  PermissionSpec{Resource: "objects", Action: "", Scope: ""},
			expectError: true,
		},
		{
			name:  "invalid scope value",
			spec:  PermissionSpec{Resource: "objects", Action: "read", Scope: "invalid"},
			expectError: true,
		},
		{
			name:  "invalid - relationships without scope",
			spec:  PermissionSpec{Resource: "relationships", Action: "create", Scope: ""},
			expectError: true,
		},
		{
			name:  "invalid - relationships read without scope",
			spec:  PermissionSpec{Resource: "relationships", Action: "read", Scope: ""},
			expectError: true,
		},
		{
			name:  "invalid - relationships update without scope",
			spec:  PermissionSpec{Resource: "relationships", Action: "update", Scope: ""},
			expectError: true,
		},
		{
			name:  "invalid - relationships delete without scope",
			spec:  PermissionSpec{Resource: "relationships", Action: "delete", Scope: ""},
			expectError: true,
		},
		{
			name: "valid scoped relationship permission",
			spec: PermissionSpec{Resource: "relationships", Action: "create", Scope: "own"},
			expectError: false,
		},
		{
			name: "valid scoped read permission for relationships",
			spec: PermissionSpec{Resource: "relationships", Action: "read", Scope: "all"},
			expectError: false,
		},
		{
			name: "valid scoped update permission for relationships",
			spec: PermissionSpec{Resource: "relationships", Action: "update", Scope: "own"},
			expectError: false,
		},
		{
			name: "valid scoped delete permission for relationships",
			spec: PermissionSpec{Resource: "relationships", Action: "delete", Scope: "all"},
			expectError: false,
		},
		{
			name: "valid objects with scope (optional)",
			spec: PermissionSpec{Resource: "objects", Action: "read", Scope: "own"},
			expectError: false,
		},
		{
			name:  "invalid scope from multi-colon input",
			spec:  PermissionSpec{Resource: "a", Action: "b", Scope: "c:d"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePermission(tt.spec)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetResourceAction(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectResource    string
		expectAction      string
		expectError       bool
	}{
		{
			name:           "flat permission",
			input:          "objects:create",
			expectResource: "objects",
			expectAction:   "create",
			expectError:    false,
		},
		{
			name:           "scoped permission",
			input:          "relationships:read:own",
			expectResource: "relationships",
			expectAction:   "read",
			expectError:    false,
		},
		{
			name:        "empty name",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource, action, err := GetResourceAction(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectResource, resource)
				assert.Equal(t, tt.expectAction, action)
			}
		})
	}
}

func TestPermissionRoundTrip(t *testing.T) {
	inputs := []string{
		"objects:create",
		"object-types:read:all",
		"relationships:read:own",
		"relationships:update:all",
		"relationships:delete:own",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			spec, err := ParsePermission(input)
			assert.NoError(t, err)

			err = ValidatePermission(spec)
			assert.NoError(t, err)

			resource, action, err := GetResourceAction(input)
			assert.NoError(t, err)
			assert.Equal(t, spec.Resource, resource)
			assert.Equal(t, spec.Action, action)
		})
	}
}
