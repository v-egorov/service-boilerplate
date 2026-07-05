package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ObjectsClient wraps HTTP calls to objects-service REST API.
type ObjectsClient struct {
	baseURL   string
	httpClient *http.Client
}

// NewObjectsClient creates a new client for the objects-service backend.
func NewObjectsClient(baseURL string, timeout time.Duration) *ObjectsClient {
	return &ObjectsClient{
		baseURL: baseURL,
		httpClient: &http.Client{Timeout: timeout},
	}
}

// --- Object Type endpoints ---

type ObjectTypeListResponse struct {
	Data []map[string]interface{} `json:"data"`
}

func (c *ObjectsClient) ListObjectTypes(ctx context.Context, typeKeyPrefix string, parentTypeID *int64) ([]map[string]interface{}, error) {
	targetURL := fmt.Sprintf("%s/api/v1/object-types", c.baseURL)
	if typeKeyPrefix != "" || parentTypeID != nil {
		params := url.Values{}
		if typeKeyPrefix != "" {
			params.Set("type_key_prefix", typeKeyPrefix)
		}
		if parentTypeID != nil {
			params.Set("parent_type_id", fmt.Sprintf("%d", *parentTypeID))
		}
		targetURL = targetURL + "?" + params.Encode()
	}
	resp, err := c.httpGet(ctx, targetURL)
	if err != nil {
		return nil, fmt.Errorf("list object types: %w", err)
	}
	defer resp.Body.Close()

	var result ObjectTypeListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode list response: %w", err)
	}
	return result.Data, nil
}

type ObjectTypeDetailResponse struct {
	Data map[string]interface{} `json:"data"`
}

func (c *ObjectsClient) GetObjectTypeByID(ctx context.Context, id int64) (map[string]interface{}, error) {
	resp, err := c.httpGet(ctx, fmt.Sprintf("%s/api/v1/object-types/%d", c.baseURL, id))
	if err != nil {
		return nil, fmt.Errorf("get object type by id %d: %w", id, err)
	}
	defer resp.Body.Close()

	var result ObjectTypeDetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode detail response: %w", err)
	}
	return result.Data, nil
}

func (c *ObjectsClient) GetObjectTypeByName(ctx context.Context, name string) (map[string]interface{}, error) {
	resp, err := c.httpGet(ctx, fmt.Sprintf("%s/api/v1/object-types/name/%s", c.baseURL, name))
	if err != nil {
		return nil, fmt.Errorf("get object type by name %q: %w", name, err)
	}
	defer resp.Body.Close()

	var result ObjectTypeDetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode detail response: %w", err)
	}
	return result.Data, nil
}

func (c *ObjectsClient) GetObjectTypeTree(ctx context.Context, id int64) (map[string]interface{}, error) {
	resp, err := c.httpGet(ctx, fmt.Sprintf("%s/api/v1/object-types/%d/tree", c.baseURL, id))
	if err != nil {
		return nil, fmt.Errorf("get object type tree for %d: %w", id, err)
	}
	defer resp.Body.Close()

	var result ObjectTypeDetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode tree response: %w", err)
	}
	return result.Data, nil
}

// GetRootTree fetches all types and builds the root hierarchy.
type HierarchyResponse struct {
	Data []map[string]interface{} `json:"data"`
}

func (c *ObjectsClient) GetRootTree(ctx context.Context) ([]map[string]interface{}, error) {
	resp, err := c.httpGet(ctx, fmt.Sprintf("%s/api/v1/object-types", c.baseURL))
	if err != nil {
		return nil, fmt.Errorf("fetch hierarchy: %w", err)
	}
	defer resp.Body.Close()

	var result HierarchyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode hierarchy response: %w", err)
	}
	// Return types where parent_type_id is null (root types with children)
	var roots []map[string]interface{}
	for _, typ := range result.Data {
		if parentID, ok := typ["parent_type_id"]; !ok || parentID == nil {
			roots = append(roots, typ)
		}
	}
	return roots, nil
}

// --- Object endpoints ---

func (c *ObjectsClient) ListObjects(ctx context.Context, objectTypeID int64, limit int, offset int, typeKeyPrefix string) ([]map[string]interface{}, map[string]any, error) {
	params := fmt.Sprintf("?object_type_id=%d&limit=%d&offset=%d", objectTypeID, limit, offset)
	if typeKeyPrefix != "" {
		params = params + "&type_key_prefix=" + url.QueryEscape(typeKeyPrefix)
	}
	fullURL := c.baseURL + "/api/v1/objects" + params
	resp, err := c.httpGet(ctx, fullURL)
	if err != nil {
		return nil, nil, fmt.Errorf("list objects: %w", err)
	}
	defer resp.Body.Close()

	// Decode both data and pagination metadata from objects-service response
	var result struct {
		Data       []map[string]interface{} `json:"data"`
		Pagination map[string]any           `json:"pagination"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, nil, fmt.Errorf("decode list response: %w", err)
	}

	// Normalize pagination values to float64 for consistency with JSON number handling
	pagination := result.Pagination
	if pagination != nil {
		for k, v := range pagination {
			switch val := v.(type) {
			case float64:
				// Already correct type from JSON decoding
			default:
				pagination[k] = val
			}
		}
	}
	return result.Data, pagination, nil
}

type ObjectDetailResponse struct {
	Data map[string]interface{} `json:"data"`
}

func (c *ObjectsClient) GetObjectByPublicID(ctx context.Context, publicID string) (map[string]interface{}, error) {
	resp, err := c.httpGet(ctx, fmt.Sprintf("%s/api/v1/objects/public-id/%s", c.baseURL, publicID))
	if err != nil {
		return nil, fmt.Errorf("get object by public_id %q: %w", publicID, err)
	}
	defer resp.Body.Close()

	var result ObjectDetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode detail response: %w", err)
	}
	return result.Data, nil
}

// forwardHeaders is the allow-list of identity headers to copy from inbound context
// onto outbound requests to objects-service.
var forwardHeaders = []string{"X-User-ID", "X-User-Email", "X-User-Roles"}

// httpGet performs a GET request with an optional identity context. When the context
// carries identity (via WithIdentity), only headers in forwardHeaders are copied onto
// the outbound request; when nil, the request proceeds without identity headers.
func (c *ObjectsClient) httpGet(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Copy allow-listed identity headers from context when present.
	if hdr := IdentityFromContext(ctx); hdr != nil {
		for _, key := range forwardHeaders {
			if val := hdr.Get(key); val != "" {
				req.Header.Set(key, val)
			}
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("objects-service returned status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// HealthCheckURL returns the objects-service base URL for health check purposes.
func (c *ObjectsClient) HealthCheckURL() string {
	return c.baseURL
}
