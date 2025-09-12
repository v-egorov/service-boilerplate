package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/v-egorov/service-boilerplate/cli/internal/config"
)

// APIClient handles HTTP requests to services
type APIClient struct {
	config     *config.Config
	httpClient *http.Client
}

// NewAPIClient creates a new API client
func NewAPIClient(cfg *config.Config) *APIClient {
	return &APIClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.API.Timeout) * time.Second,
		},
	}
}

// Request represents an API request
type Request struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    interface{}       `json:"body,omitempty"`
}

// Response represents an API response
type Response struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers,omitempty"`
	Body       interface{}         `json:"body,omitempty"`
	Error      string              `json:"error,omitempty"`
}

// MakeRequest makes an HTTP request with retry logic
func (c *APIClient) MakeRequest(req *Request) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.config.Services.RetryAttempts; attempt++ {
		resp, err := c.doRequest(req)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Don't retry on the last attempt
		if attempt < c.config.Services.RetryAttempts {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", c.config.Services.RetryAttempts+1, lastErr)
}

// doRequest performs the actual HTTP request
func (c *APIClient) doRequest(req *Request) (*Response, error) {
	var bodyReader io.Reader

	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Make the request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	response := &Response{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Header,
	}

	// Parse JSON response if possible
	if len(respBody) > 0 {
		var jsonBody interface{}
		if err := json.Unmarshal(respBody, &jsonBody); err == nil {
			response.Body = jsonBody
		} else {
			response.Body = string(respBody)
		}
	}

	// Check for HTTP errors
	if httpResp.StatusCode >= 400 {
		response.Error = fmt.Sprintf("HTTP %d: %s", httpResp.StatusCode, string(respBody))
		return response, fmt.Errorf("HTTP error %d: %s", httpResp.StatusCode, string(respBody))
	}

	return response, nil
}

// Get makes a GET request
func (c *APIClient) Get(url string, headers map[string]string) (*Response, error) {
	return c.MakeRequest(&Request{
		Method:  "GET",
		URL:     url,
		Headers: headers,
	})
}

// Post makes a POST request
func (c *APIClient) Post(url string, body interface{}, headers map[string]string) (*Response, error) {
	return c.MakeRequest(&Request{
		Method:  "POST",
		URL:     url,
		Headers: headers,
		Body:    body,
	})
}

// Put makes a PUT request
func (c *APIClient) Put(url string, body interface{}, headers map[string]string) (*Response, error) {
	return c.MakeRequest(&Request{
		Method:  "PUT",
		URL:     url,
		Headers: headers,
		Body:    body,
	})
}

// Delete makes a DELETE request
func (c *APIClient) Delete(url string, headers map[string]string) (*Response, error) {
	return c.MakeRequest(&Request{
		Method:  "DELETE",
		URL:     url,
		Headers: headers,
	})
}

// CallService makes a request to a specific service
func (c *APIClient) CallService(serviceName, method, endpoint string, body interface{}, headers map[string]string) (*Response, error) {
	serviceURL := c.config.GetServiceURL(serviceName)
	url := fmt.Sprintf("%s%s", serviceURL, endpoint)

	req := &Request{
		Method:  method,
		URL:     url,
		Headers: headers,
		Body:    body,
	}

	return c.MakeRequest(req)
}
