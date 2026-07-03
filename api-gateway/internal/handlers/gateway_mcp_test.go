package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/api-gateway/internal/services"
)

func TestProxyMCPRequest_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock MCP server that validates injected identity headers
	mcpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		userEmail := r.Header.Get("X-User-Email")
		userRoles := r.Header.Get("X-User-Roles")

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: {\"type\":\"test\"}\n\n"))

		// Identity must be a UUID (not string "mcp-agent") so auth-service uuid.MustParse() doesn't panic
		if userID == "" {
			t.Error("Expected non-empty X-User-ID (UUID)")
		}
		if userEmail != "mcp-agent@system.internal" {
			t.Errorf("Expected X-User-Email=mcp-agent@system.internal, got %s", userEmail)
		}
		if !strings.Contains(userRoles, "mcp-agent-read-only") {
			t.Errorf("Expected X-User-Roles to contain mcp-agent-read-only, got %s", userRoles)
		}

		w.Write([]byte("event: message\ndata: {\"jsonrpc\":\"2.0\",\"result\":null}\n\n"))
	}))
	defer mcpServer.Close()

	// Setup registry with MCP server
	registry := services.NewServiceRegistry(&logrus.Logger{})
	registry.RegisterService("mcp-server", mcpServer.URL)
	handler, _ := newTestGatewayHandler(t, registry)

	// Create a gin router and serve it via httptest.Server (for proper http.CloseNotifier support)
	router := gin.New()
	router.GET("/mcp/sse", handler.ProxyMCPRequest())

	ginServer := httptest.NewServer(router)
	defer ginServer.Close()

	// Make actual HTTP request through the server
	resp, err := http.Get(ginServer.URL + "/mcp/sse")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)

	// Verify response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(body, "{\"type\":\"test\"}") {
		t.Errorf("Expected SSE response body to contain test data, got: %s", body)
	}
}

func TestProxyMCPRequest_ServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup registry WITHOUT MCP server
	registry := services.NewServiceRegistry(&logrus.Logger{})
	handler, _ := newTestGatewayHandler(t, registry)

	router := gin.New()
	router.GET("/mcp/sse", handler.ProxyMCPRequest())

	ginServer := httptest.NewServer(router)
	defer ginServer.Close()

	resp, err := http.Get(ginServer.URL + "/mcp/sse")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503 (service unavailable), got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if body["error"] != "MCP service unavailable" {
		t.Errorf("Expected error message 'MCP service unavailable', got '%v'", body["error"])
	}
}

func TestProxyMCPRequest_POSTMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var receivedBody string
	mcpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		receivedBody = string(bodyBytes)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: {\"jsonrpc\":\"2.0\",\"result\":null}\n\n"))
	}))
	defer mcpServer.Close()

	registry := services.NewServiceRegistry(&logrus.Logger{})
	registry.RegisterService("mcp-server", mcpServer.URL)
	handler, _ := newTestGatewayHandler(t, registry)

	router := gin.New()
	router.POST("/mcp/sse", handler.ProxyMCPRequest())

	ginServer := httptest.NewServer(router)
	defer ginServer.Close()

	reqBody := `{"jsonrpc":"2.0","method":"tools/call","params":{"name":"list_object_types"}}`
	resp, err := http.Post(ginServer.URL+"/mcp/sse", "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	if receivedBody == "" {
		t.Error("Expected request body to be forwarded")
	}
	if !strings.Contains(receivedBody, "list_object_types") {
		t.Errorf("Expected forwarded body to contain 'list_object_types', got: %s", receivedBody)
	}
}

func TestMultiHandler_MCPRoutes(t *testing.T) {
	mainRouter := gin.New()
	mainRouter.GET("/api/test", func(c *gin.Context) {
		c.String(200, "main-router-response")
	})

	mcpRouter := gin.New()
	mcpRouter.GET("/mcp/sse", func(c *gin.Context) {
		c.String(200, "mcp-router-response")
	})

	logger := &logrus.Logger{}
	multiHandler := &MultiHandler{
		MainRouter:  mainRouter,
		HTTPHandler: mcpRouter,
		Prefix:      "/mcp",
		Logger:      logger,
	}

	// Test MCP route goes to mcpRouter
	req1 := httptest.NewRequest("GET", "/mcp/sse", nil)
	rec1 := httptest.NewRecorder()
	multiHandler.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Errorf("MCP route: expected status 200, got %d", rec1.Code)
	}
	if !strings.Contains(rec1.Body.String(), "mcp-router-response") {
		t.Errorf("MCP route: expected mcp-router response, got: %s", rec1.Body.String())
	}

	// Test main router route goes to mainRouter
	req2 := httptest.NewRequest("GET", "/api/test", nil)
	rec2 := httptest.NewRecorder()
	multiHandler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("Main route: expected status 200, got %d", rec2.Code)
	}
	if !strings.Contains(rec2.Body.String(), "main-router-response") {
		t.Errorf("Main route: expected main-router response, got: %s", rec2.Body.String())
	}

	// Test /mcp/health goes to mcpRouter (prefix match)
	req3 := httptest.NewRequest("GET", "/mcp/health", nil)
	rec3 := httptest.NewRecorder()
	multiHandler.ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusNotFound {
		t.Errorf("/mcp/health should go to mcpRouter (404), got %d", rec3.Code)
	}
}

func TestMultiHandler_NonMCPRoutes(t *testing.T) {
	mainRouter := gin.New()
	mainRouter.GET("/ready", func(c *gin.Context) {
		c.String(200, "ready")
	})
	mainRouter.GET("/health", func(c *gin.Context) {
		c.String(200, "health")
	})

	mcpRouter := gin.New()
	mcpRouter.GET("/mcp/sse", func(c *gin.Context) {
		c.String(200, "sse")
	})

	logger := &logrus.Logger{}
	multiHandler := &MultiHandler{
		MainRouter:  mainRouter,
		HTTPHandler: mcpRouter,
		Prefix:      "/mcp",
		Logger:      logger,
	}

	testCases := []struct {
		path     string
		expected string
	}{
		{"/ready", "ready"},
		{"/health", "health"},
		{"/api/v1/users", ""}, // Will be 404 but should go to mainRouter
	}

	for _, tc := range testCases {
		req := httptest.NewRequest("GET", tc.path, nil)
		rec := httptest.NewRecorder()
		multiHandler.ServeHTTP(rec, req)

		if tc.expected != "" && !strings.Contains(rec.Body.String(), tc.expected) {
			t.Errorf("Path %s: expected '%s', got: %s", tc.path, tc.expected, rec.Body.String())
		}
	}
}
