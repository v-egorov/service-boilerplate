package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func (c *ObjectsClient) ListObjectTypes() ([]map[string]interface{}, error) {
	resp, err := c.httpGet(fmt.Sprintf("%s/api/v1/object-types", c.baseURL))
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

func (c *ObjectsClient) GetObjectTypeByID(id int64) (map[string]interface{}, error) {
	resp, err := c.httpGet(fmt.Sprintf("%s/api/v1/object-types/%d", c.baseURL, id))
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

func (c *ObjectsClient) GetObjectTypeByName(name string) (map[string]interface{}, error) {
	resp, err := c.httpGet(fmt.Sprintf("%s/api/v1/object-types/name/%s", c.baseURL, name))
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

func (c *ObjectsClient) GetObjectTypeTree(id int64) (map[string]interface{}, error) {
	resp, err := c.httpGet(fmt.Sprintf("%s/api/v1/object-types/%d/tree", c.baseURL, id))
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

func (c *ObjectsClient) GetRootTree() ([]map[string]interface{}, error) {
	resp, err := c.httpGet(fmt.Sprintf("%s/api/v1/object-types", c.baseURL))
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

type ObjectListResponse struct {
	Data []map[string]interface{} `json:"data"`
}

func (c *ObjectsClient) ListObjects(objectTypeID int64, page int, pageSize int) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/objects?object_type_id=%d&page=%d&page_size=%d", c.baseURL, objectTypeID, page, pageSize)
	resp, err := c.httpGet(url)
	if err != nil {
		return nil, fmt.Errorf("list objects: %w", err)
	}
	defer resp.Body.Close()

	var result ObjectListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode list response: %w", err)
	}
	return result.Data, nil
}

type ObjectDetailResponse struct {
	Data map[string]interface{} `json:"data"`
}

func (c *ObjectsClient) GetObjectByPublicID(publicID string) (map[string]interface{}, error) {
	resp, err := c.httpGet(fmt.Sprintf("%s/api/v1/objects/public-id/%s", c.baseURL, publicID))
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

// httpGet performs a GET request and returns the response body for JSON decoding.
// HealthCheckURL returns the objects-service base URL for health check purposes.
// This is used by the MCP server's own health handler to verify backend connectivity.
func (c *ObjectsClient) HealthCheckURL() string {
	return c.baseURL
}

func (c *ObjectsClient) httpGet(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
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
