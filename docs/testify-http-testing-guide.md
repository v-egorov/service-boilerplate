# Testify HTTP Testing Guide

## Overview

Testify provides HTTP testing utilities through the `httptest` package integration and custom helpers. This guide covers HTTP testing patterns, request/response testing, and integration with testify assertions.

## References

- **Testify Repository**: [github.com/stretchr/testify](https://github.com/stretchr/testify)
- **HTTP Testing**: [pkg.go.dev/net/http/httptest](https://pkg.go.dev/net/http/httptest)
- **Testify HTTP Helpers**: Custom patterns for HTTP testing

## HTTP Testing Fundamentals

### Basic HTTP Test Structure

```go
import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestAPIEndpoint(t *testing.T) {
    // Create request
    req, err := http.NewRequest("GET", "/api/users", nil)
    require.NoError(t, err)

    // Create response recorder
    rr := httptest.NewRecorder()

    // Create handler and serve
    handler := http.HandlerFunc(userHandler)
    handler.ServeHTTP(rr, req)

    // Assert response
    assert.Equal(t, http.StatusOK, rr.Code)
    assert.Contains(t, rr.Body.String(), "users")
}
```

### HTTP Test Suite Pattern

```go
type HTTPTestSuite struct {
    suite.Suite
    server *httptest.Server
    client *http.Client
    baseURL string
}

func (suite *HTTPTestSuite) SetupSuite() {
    // Create test server with your router
    router := setupTestRouter()
    suite.server = httptest.NewServer(router)
    suite.baseURL = suite.server.URL

    // Configure client
    suite.client = &http.Client{
        Timeout: 10 * time.Second,
    }
}

func (suite *HTTPTestSuite) TearDownSuite() {
    if suite.server != nil {
        suite.server.Close()
    }
}

func (suite *HTTPTestSuite) TestGetUsers() {
    resp, err := suite.client.Get(suite.baseURL + "/api/users")
    suite.Require().NoError(err)
    defer resp.Body.Close()

    suite.Equal(http.StatusOK, resp.StatusCode)
    suite.Contains(resp.Header.Get("Content-Type"), "application/json")

    var users []User
    err = json.NewDecoder(resp.Body).Decode(&users)
    suite.NoError(err)
    suite.NotEmpty(users)
}

func TestHTTPTestSuite(t *testing.T) {
    suite.Run(t, new(HTTPTestSuite))
}
```

## Request Testing Patterns

### Creating Test Requests

```go
func createTestRequest(method, url string, body io.Reader) *http.Request {
    req, _ := http.NewRequest(method, url, body)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer test-token")
    return req
}

func TestUserCreation(t *testing.T) {
    userData := `{"email":"test@example.com","name":"Test User"}`
    req := createTestRequest("POST", "/api/users", strings.NewReader(userData))

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(createUserHandler)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusCreated, rr.Code)

    var response map[string]interface{}
    err := json.NewDecoder(rr.Body).Decode(&response)
    assert.NoError(t, err)
    assert.Contains(t, response, "id")
}
```

### Request with Context

```go
func TestHandlerWithContext(t *testing.T) {
    req := httptest.NewRequest("GET", "/api/data", nil)

    // Add context values
    ctx := context.WithValue(req.Context(), "user_id", "123")
    req = req.WithContext(ctx)

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(dataHandler)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)
}
```

### Multipart Form Data

```go
func createMultipartRequest(t *testing.T, url string, params map[string]string, fileField, filename, content string) *http.Request {
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    // Add form fields
    for key, value := range params {
        writer.WriteField(key, value)
    }

    // Add file
    part, err := writer.CreateFormFile(fileField, filename)
    require.NoError(t, err)
    _, err = part.Write([]byte(content))
    require.NoError(t, err)
    writer.Close()

    req := httptest.NewRequest("POST", url, body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    return req
}

func TestFileUpload(t *testing.T) {
    params := map[string]string{"title": "Test File"}
    req := createMultipartRequest(t, "/api/upload", params, "file", "test.txt", "file content")

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(uploadHandler)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)
}
```

## Response Testing Patterns

### JSON Response Testing

```go
func assertJSONResponse(t *testing.T, rr *httptest.ResponseRecorder, expectedStatus int, expectedBody interface{}) {
    assert.Equal(t, expectedStatus, rr.Code)
    assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

    var actualBody interface{}
    err := json.NewDecoder(rr.Body).Decode(&actualBody)
    assert.NoError(t, err)

    // For complex assertions, unmarshal to specific type
    if expectedMap, ok := expectedBody.(map[string]interface{}); ok {
        actualMap := actualBody.(map[string]interface{})
        for key, expectedValue := range expectedMap {
            assert.Equal(t, expectedValue, actualMap[key])
        }
    } else {
        assert.Equal(t, expectedBody, actualBody)
    }
}

func TestGetUser(t *testing.T) {
    req := httptest.NewRequest("GET", "/api/users/123", nil)
    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(getUserHandler)
    handler.ServeHTTP(rr, req)

    expectedResponse := map[string]interface{}{
        "id": "123",
        "email": "user@example.com",
        "name": "Test User",
    }

    assertJSONResponse(t, rr, http.StatusOK, expectedResponse)
}
```

### Error Response Testing

```go
func assertErrorResponse(t *testing.T, rr *httptest.ResponseRecorder, expectedStatus int, expectedError string) {
    assert.Equal(t, expectedStatus, rr.Code)

    var errorResponse map[string]interface{}
    err := json.NewDecoder(rr.Body).Decode(&errorResponse)
    assert.NoError(t, err)

    assert.Contains(t, errorResponse, "error")
    assert.Contains(t, errorResponse["error"].(string), expectedError)
}

func TestCreateUser_InvalidData(t *testing.T) {
    invalidData := `{"email":"invalid-email","name":""}`
    req := createTestRequest("POST", "/api/users", strings.NewReader(invalidData))

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(createUserHandler)
    handler.ServeHTTP(rr, req)

    assertErrorResponse(t, rr, http.StatusBadRequest, "invalid email")
}
```

### Header Testing

```go
func assertResponseHeaders(t *testing.T, rr *httptest.ResponseRecorder, expectedHeaders map[string]string) {
    for key, expectedValue := range expectedHeaders {
        actualValue := rr.Header().Get(key)
        assert.Equal(t, expectedValue, actualValue, "Header %s should match", key)
    }
}

func TestAPIResponseHeaders(t *testing.T) {
    req := httptest.NewRequest("GET", "/api/data", nil)
    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(dataHandler)
    handler.ServeHTTP(rr, req)

    expectedHeaders := map[string]string{
        "Content-Type": "application/json",
        "X-API-Version": "v1",
        "Cache-Control": "no-cache",
    }

    assertResponseHeaders(t, rr, expectedHeaders)
}
```

## Middleware Testing

### Authentication Middleware Testing

```go
func TestAuthMiddleware(t *testing.T) {
    // Test without auth header
    req := httptest.NewRequest("GET", "/api/protected", nil)
    rr := httptest.NewRecorder()

    handler := authMiddleware(http.HandlerFunc(protectedHandler))
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusUnauthorized, rr.Code)

    // Test with valid auth header
    req = httptest.NewRequest("GET", "/api/protected", nil)
    req.Header.Set("Authorization", "Bearer valid-token")
    rr = httptest.NewRecorder()

    handler.ServeHTTP(rr, req)
    assert.Equal(t, http.StatusOK, rr.Code)
}
```

### CORS Middleware Testing

```go
func TestCORSMiddleware(t *testing.T) {
    req := httptest.NewRequest("OPTIONS", "/api/data", nil)
    req.Header.Set("Origin", "http://localhost:3000")
    req.Header.Set("Access-Control-Request-Method", "POST")

    rr := httptest.NewRecorder()

    handler := corsMiddleware(http.HandlerFunc(dataHandler))
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)
    assert.Equal(t, "http://localhost:3000", rr.Header().Get("Access-Control-Allow-Origin"))
    assert.Contains(t, rr.Header().Get("Access-Control-Allow-Methods"), "POST")
}
```

## Integration Testing Patterns

### Full API Flow Testing

```go
func TestUserLifecycle(t *testing.T) {
    suite := &HTTPTestSuite{}
    suite.SetupSuite()
    defer suite.TearDownSuite()

    // 1. Create user
    userData := `{"email":"lifecycle@example.com","name":"Lifecycle User"}`
    resp, err := suite.client.Post(
        suite.baseURL+"/api/users",
        "application/json",
        strings.NewReader(userData),
    )
    require.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    var createdUser map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&createdUser)
    userID := createdUser["id"].(string)
    resp.Body.Close()

    // 2. Get user
    resp, err = suite.client.Get(suite.baseURL + "/api/users/" + userID)
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)

    var fetchedUser map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&fetchedUser)
    assert.Equal(t, userID, fetchedUser["id"])
    resp.Body.Close()

    // 3. Update user
    updateData := `{"name":"Updated Lifecycle User"}`
    req, _ := http.NewRequest("PUT", suite.baseURL+"/api/users/"+userID, strings.NewReader(updateData))
    req.Header.Set("Content-Type", "application/json")
    resp, err = suite.client.Do(req)
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    resp.Body.Close()

    // 4. Delete user
    req, _ = http.NewRequest("DELETE", suite.baseURL+"/api/users/"+userID, nil)
    resp, err = suite.client.Do(req)
    require.NoError(t, err)
    assert.Equal(t, http.StatusNoContent, resp.StatusCode)
    resp.Body.Close()

    // 5. Verify user is gone
    resp, err = suite.client.Get(suite.baseURL + "/api/users/" + userID)
    require.NoError(t, err)
    assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
```

### Database Integration Testing

```go
type APIDatabaseTestSuite struct {
    suite.Suite
    server *httptest.Server
    db *sql.DB
    client *http.Client
}

func (suite *APIDatabaseTestSuite) SetupSuite() {
    // Setup database
    db, err := setupTestDatabase()
    suite.Require().NoError(err)
    suite.db = db

    // Setup API server
    router := setupRouter(suite.db)
    suite.server = httptest.NewServer(router)
    suite.client = &http.Client{Timeout: 10 * time.Second}
}

func (suite *APIDatabaseTestSuite) TearDownSuite() {
    if suite.server != nil {
        suite.server.Close()
    }
    if suite.db != nil {
        suite.db.Close()
    }
}

func (suite *APIDatabaseTestSuite) TestCreateUser_Integration() {
    userData := `{"email":"integration@example.com","name":"Integration User"}`

    resp, err := suite.client.Post(
        suite.server.URL+"/api/users",
        "application/json",
        strings.NewReader(userData),
    )
    suite.Require().NoError(err)
    defer resp.Body.Close()

    suite.Equal(http.StatusCreated, resp.StatusCode)

    // Verify in database
    var count int
    err = suite.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", "integration@example.com").Scan(&count)
    suite.NoError(err)
    suite.Equal(1, count)
}
```

## Best Practices

### 1. Test Organization

```go
// Group related tests
func TestUserAPI(t *testing.T) {
    t.Run("Create", testCreateUser)
    t.Run("Get", testGetUser)
    t.Run("Update", testUpdateUser)
    t.Run("Delete", testDeleteUser)
    t.Run("List", testListUsers)
}

func TestUserAPI_Errors(t *testing.T) {
    t.Run("InvalidData", testCreateUserInvalidData)
    t.Run("NotFound", testGetUserNotFound)
    t.Run("Unauthorized", testProtectedEndpointUnauthorized)
}
```

### 2. Request/Response Helpers

```go
// Helper for creating authenticated requests
func createAuthenticatedRequest(t *testing.T, method, url string, body io.Reader, token string) *http.Request {
    req, err := http.NewRequest(method, url, body)
    require.NoError(t, err)

    if token != "" {
        req.Header.Set("Authorization", "Bearer "+token)
    }
    req.Header.Set("Content-Type", "application/json")

    return req
}

// Helper for asserting common response patterns
func assertAPIResponse(t *testing.T, rr *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
    assert.Equal(t, expectedStatus, rr.Code)
    assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

    var response map[string]interface{}
    err := json.NewDecoder(rr.Body).Decode(&response)
    assert.NoError(t, err)

    return response
}
```

### 3. Test Data Management

```go
func setupTestUser(t *testing.T, client *http.Client, baseURL string) string {
    userData := `{"email":"test@example.com","name":"Test User"}`
    resp, err := client.Post(baseURL+"/api/users", "application/json", strings.NewReader(userData))
    require.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    var user map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&user)
    return user["id"].(string)
}

func cleanupTestUser(t *testing.T, client *http.Client, baseURL, userID string) {
    req, _ := http.NewRequest("DELETE", baseURL+"/api/users/"+userID, nil)
    resp, err := client.Do(req)
    if err == nil {
        resp.Body.Close()
    }
}
```

### 4. Performance Testing

```go
func BenchmarkAPIEndpoint(b *testing.B) {
    req := httptest.NewRequest("GET", "/api/users", nil)
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(userHandler)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        rr = httptest.NewRecorder() // Reset for each iteration
        handler.ServeHTTP(rr, req)
    }
}
```

### 5. Concurrent Request Testing

```go
func TestConcurrentRequests(t *testing.T) {
    const numRequests = 10
    var wg sync.WaitGroup
    results := make(chan int, numRequests)

    for i := 0; i < numRequests; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()

            req := httptest.NewRequest("GET", "/api/data", nil)
            rr := httptest.NewRecorder()
            handler := http.HandlerFunc(dataHandler)
            handler.ServeHTTP(rr, req)

            results <- rr.Code
        }()
    }

    wg.Wait()
    close(results)

    for statusCode := range results {
        assert.Equal(t, http.StatusOK, statusCode)
    }
}
```

## Integration with Our Codebase

### Current HTTP Testing

Our current handler tests use basic httptest patterns:
```go
func TestAuthHandler_Login(t *testing.T) {
    // Create request
    req := createTestRequest("POST", "/auth/login", loginData)

    // Create recorder
    rr := httptest.NewRecorder()

    // Execute
    handler := http.HandlerFunc(authHandler.Login)
    handler.ServeHTTP(rr, req)

    // Assert
    assert.Equal(t, http.StatusOK, rr.Code)
}
```

### Enhanced Patterns for Our Codebase

```go
// Helper for our auth handlers
func createAuthRequest(t *testing.T, method, path string, body interface{}, token string) *http.Request {
    var bodyReader io.Reader
    if body != nil {
        bodyBytes, err := json.Marshal(body)
        require.NoError(t, err)
        bodyReader = bytes.NewReader(bodyBytes)
    }

    req, err := http.NewRequest(method, path, bodyReader)
    require.NoError(t, err)

    req.Header.Set("Content-Type", "application/json")
    if token != "" {
        req.Header.Set("Authorization", "Bearer "+token)
    }

    return req
}

// Usage in our tests
func TestAuthHandler_Login(t *testing.T) {
    loginData := map[string]string{
        "email": "user@example.com",
        "password": "password123",
    }

    req := createAuthRequest(t, "POST", "/auth/login", loginData, "")
    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(authHandler.Login)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)

    var response map[string]interface{}
    json.NewDecoder(rr.Body).Decode(&response)
    assert.Contains(t, response, "access_token")
    assert.Contains(t, response, "refresh_token")
}
```

## Common Pitfalls

### 1. Not Closing Response Bodies

```go
// Bad - memory leak
resp, _ := client.Get(url)
// resp.Body not closed

// Good - proper cleanup
resp, err := client.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()
```

### 2. Testing Implementation Details

```go
// Bad - tests internal JSON structure
var response map[string]interface{}
json.NewDecoder(rr.Body).Decode(&response)
assert.Equal(t, "internal_field", response["internal_key"])

// Good - tests behavior
assert.Equal(t, http.StatusOK, rr.Code)
assert.Contains(t, rr.Body.String(), "success")
```

### 3. Ignoring HTTP Status Codes

```go
// Bad - only checks response body
resp, _ := client.Get(url)
var data map[string]interface{}
json.NewDecoder(resp.Body).Decode(&data)
// No status code check

// Good - always check status
resp, err := client.Get(url)
require.NoError(t, err)
assert.Equal(t, http.StatusOK, resp.StatusCode)
```

### 4. Race Conditions in Concurrent Tests

```go
// Bad - shared state between tests
var globalCounter int

func TestConcurrentEndpoint(t *testing.T) {
    globalCounter++
    // Test logic using globalCounter
}

// Good - isolated test state
func TestConcurrentEndpoint(t *testing.T) {
    counter := 0
    // Test logic using local counter
}
```

## Conclusion

HTTP testing with testify provides powerful patterns for testing web APIs. Key principles:

1. **Use httptest.NewRecorder()** for response recording
2. **Always check status codes** and headers
3. **Close response bodies** to prevent memory leaks
4. **Use suites** for complex HTTP testing scenarios
5. **Test error conditions** thoroughly
6. **Keep tests isolated** and independent

These patterns ensure robust, maintainable HTTP API tests that integrate seamlessly with testify's assertion capabilities.