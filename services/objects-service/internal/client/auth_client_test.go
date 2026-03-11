package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckPermission_Allowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req checkPermissionRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "user-123", req.UserID)
		assert.Equal(t, "objects:create", req.Permission)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(checkPermissionResponse{
			Allowed:    true,
			UserID:     "user-123",
			Permission: "objects:create",
		})
	}))
	defer server.Close()

	client := NewAuthClient(AuthClientConfig{
		BaseURL: server.URL,
	}, nil)

	allowed, err := client.CheckPermission(context.Background(), "user-123", "objects:create", "")
	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestCheckPermission_Denied(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(checkPermissionResponse{
			Allowed:    false,
			UserID:     "user-123",
			Permission: "objects:delete:all",
		})
	}))
	defer server.Close()

	client := NewAuthClient(AuthClientConfig{
		BaseURL: server.URL,
	}, nil)

	allowed, err := client.CheckPermission(context.Background(), "user-123", "objects:delete:all", "")
	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestCheckPermission_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewAuthClient(AuthClientConfig{
		BaseURL: server.URL,
	}, nil)

	allowed, err := client.CheckPermission(context.Background(), "user-123", "objects:create", "")
	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "500")
}

func TestGetUserPermissions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/auth/users/user-123/permissions", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			Permissions []string `json:"permissions"`
		}{
			Permissions: []string{"objects:create", "objects:read:own", "object-types:read"},
		})
	}))
	defer server.Close()

	client := NewAuthClient(AuthClientConfig{
		BaseURL: server.URL,
	}, nil)

	permissions, err := client.GetUserPermissions(context.Background(), "user-123", "")
	assert.NoError(t, err)
	assert.Equal(t, []string{"objects:create", "objects:read:own", "object-types:read"}, permissions)
}

func TestGetUserRoles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/auth/users/user-123/roles", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			Roles []string `json:"roles"`
		}{
			Roles: []string{"user", "object-type-admin"},
		})
	}))
	defer server.Close()

	client := NewAuthClient(AuthClientConfig{
		BaseURL: server.URL,
	}, nil)

	roles, err := client.GetUserRoles(context.Background(), "user-123", "")
	assert.NoError(t, err)
	assert.Equal(t, []string{"user", "object-type-admin"}, roles)
}

func TestAuthClient_DefaultTimeout(t *testing.T) {
	client := NewAuthClient(AuthClientConfig{
		BaseURL: "http://localhost:8083",
	}, nil)

	ac := client.(*authClient)
	assert.Equal(t, 10*time.Second, ac.httpClient.Timeout)
}
