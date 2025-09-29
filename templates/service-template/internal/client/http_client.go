package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// HTTPClient provides HTTP client functionality with request ID propagation
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(baseURL string, logger *logrus.Logger) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// DoRequest performs an HTTP request with request ID propagation
func (c *HTTPClient) DoRequest(ctx context.Context, method, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	var requestBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		requestBody = bytes.NewBuffer(jsonData)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type for JSON requests
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Extract and set request ID
	if requestID, ok := ctx.Value("request_id").(string); ok {
		req.Header.Set("X-Request-ID", requestID)
	}

	// Inject OpenTelemetry trace context
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

// Get performs a GET request
func (c *HTTPClient) Get(ctx context.Context, path string, headers map[string]string) (*http.Response, error) {
	return c.DoRequest(ctx, "GET", path, nil, headers)
}

// Post performs a POST request
func (c *HTTPClient) Post(ctx context.Context, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.DoRequest(ctx, "POST", path, body, headers)
}

// Put performs a PUT request
func (c *HTTPClient) Put(ctx context.Context, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.DoRequest(ctx, "PUT", path, body, headers)
}

// Delete performs a DELETE request
func (c *HTTPClient) Delete(ctx context.Context, path string, headers map[string]string) (*http.Response, error) {
	return c.DoRequest(ctx, "DELETE", path, nil, headers)
}

// Close closes the HTTP client
func (c *HTTPClient) Close() {
	// Clean up if needed
}
