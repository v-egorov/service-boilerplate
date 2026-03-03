package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type AuthClient interface {
	CheckPermission(ctx context.Context, userID, permission string) (bool, error)
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
	GetUserRoles(ctx context.Context, userID string) ([]string, error)
}

type authClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
}

type AuthClientConfig struct {
	BaseURL string
	Timeout time.Duration
}

func NewAuthClient(cfg AuthClientConfig, logger *logrus.Logger) AuthClient {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	return &authClient{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger,
	}
}

type checkPermissionRequest struct {
	UserID     string `json:"user_id"`
	Permission string `json:"permission"`
}

type checkPermissionResponse struct {
	Allowed    bool   `json:"allowed"`
	UserID     string `json:"user_id"`
	Permission string `json:"permission"`
}

func (c *authClient) CheckPermission(ctx context.Context, userID, permission string) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/auth/permissions/check", c.baseURL)

	reqBody := checkPermissionRequest{
		UserID:     userID,
		Permission: permission,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to call auth-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("auth-service returned status %d", resp.StatusCode)
	}

	var result checkPermissionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Allowed, nil
}

func (c *authClient) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/auth/users/%s/permissions", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call auth-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth-service returned status %d", resp.StatusCode)
	}

	var result struct {
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Permissions, nil
}

func (c *authClient) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/auth/users/%s/roles", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call auth-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth-service returned status %d", resp.StatusCode)
	}

	var result struct {
		Roles []string `json:"roles"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Roles, nil
}
