# Testify Suite Testing Guide

## Overview

Testify's `suite` package provides a structured way to organize tests with shared setup/teardown logic, test fixtures, and better test organization. This guide covers when and how to use testify suites for effective testing.

## References

- **Testify Repository**: [github.com/stretchr/testify](https://github.com/stretchr/testify)
- **Suite Package**: [pkg.go.dev/github.com/stretchr/testify/suite](https://pkg.go.dev/github.com/stretchr/testify/suite)

## When to Use Testify Suites

### Beneficial Scenarios

1. **Shared Setup/Teardown**: Tests that need common database setup, API server startup, or resource initialization
2. **Test Fixtures**: When multiple tests operate on the same data structures
3. **Complex Test Organization**: Large test files with many related test methods
4. **Integration Testing**: Tests that require external services or complex state management

### When Traditional Tests Are Better

1. **Simple Unit Tests**: Single-method tests with minimal setup
2. **Performance**: Suite overhead for very simple tests
3. **One-off Tests**: Tests that don't share common setup

## Suite Structure

### Basic Suite Structure

```go
import (
    "testing"
    "github.com/stretchr/testify/suite"
)

// TestSuite embeds suite.Suite
type UserServiceTestSuite struct {
    suite.Suite
    // Shared test data
    service *UserService
    testUsers []*User
    db *sql.DB
}

// SetupSuite runs once before all tests in the suite
func (suite *UserServiceTestSuite) SetupSuite() {
    // One-time setup (database connection, external services)
    db, err := setupTestDatabase()
    suite.Require().NoError(err)
    suite.db = db
}

// TearDownSuite runs once after all tests in the suite
func (suite *UserServiceTestSuite) TearDownSuite() {
    // One-time cleanup
    if suite.db != nil {
        suite.db.Close()
    }
}

// SetupTest runs before each test method
func (suite *UserServiceTestSuite) SetupTest() {
    // Per-test setup
    suite.service = NewUserService(suite.db)

    // Create test data
    suite.testUsers = []*User{
        {ID: 1, Email: "user1@example.com", Name: "User 1"},
        {ID: 2, Email: "user2@example.com", Name: "User 2"},
    }
}

// TearDownTest runs after each test method
func (suite *UserServiceTestSuite) TearDownTest() {
    // Per-test cleanup
    // Reset database state, clean up files, etc.
}

// Test methods
func (suite *UserServiceTestSuite) TestCreateUser() {
    user := &User{Email: "new@example.com", Name: "New User"}
    err := suite.service.CreateUser(user)
    suite.NoError(err)
    suite.NotZero(user.ID)
}

func (suite *UserServiceTestSuite) TestGetUser() {
    user, err := suite.service.GetUser(suite.testUsers[0].ID)
    suite.NoError(err)
    suite.Equal(suite.testUsers[0].Email, user.Email)
}

// Run the suite
func TestUserServiceTestSuite(t *testing.T) {
    suite.Run(t, new(UserServiceTestSuite))
}
```

## Advanced Suite Patterns

### Database Testing Suite

```go
type DatabaseTestSuite struct {
    suite.Suite
    db *sql.DB
    tx *sql.Tx
}

func (suite *DatabaseTestSuite) SetupSuite() {
    // Connect to test database
    db, err := sql.Open("postgres", testDBURL)
    suite.Require().NoError(err)
    suite.db = db
}

func (suite *DatabaseTestSuite) SetupTest() {
    // Start transaction for each test
    tx, err := suite.db.Begin()
    suite.Require().NoError(err)
    suite.tx = tx
}

func (suite *DatabaseTestSuite) TearDownTest() {
    // Rollback transaction to reset state
    if suite.tx != nil {
        suite.tx.Rollback()
    }
}

func (suite *DatabaseTestSuite) TearDownSuite() {
    if suite.db != nil {
        suite.db.Close()
    }
}

func (suite *DatabaseTestSuite) TestUserCreation() {
    // Use suite.tx instead of suite.db for transactional testing
    repo := NewUserRepository(suite.tx)
    user := &User{Email: "test@example.com"}
    err := repo.Create(user)
    suite.NoError(err)
}
```

### HTTP API Testing Suite

```go
type APITestSuite struct {
    suite.Suite
    server *httptest.Server
    client *http.Client
    baseURL string
}

func (suite *APITestSuite) SetupSuite() {
    // Start test server
    router := setupTestRouter()
    suite.server = httptest.NewServer(router)
    suite.baseURL = suite.server.URL
    suite.client = &http.Client{Timeout: 10 * time.Second}
}

func (suite *APITestSuite) TearDownSuite() {
    if suite.server != nil {
        suite.server.Close()
    }
}

func (suite *APITestSuite) TestCreateUserAPI() {
    userData := `{"email":"test@example.com","name":"Test User"}`
    resp, err := suite.client.Post(
        suite.baseURL+"/api/users",
        "application/json",
        strings.NewReader(userData),
    )

    suite.NoError(err)
    suite.Equal(http.StatusCreated, resp.StatusCode)

    var response map[string]interface{}
    err = json.NewDecoder(resp.Body).Decode(&response)
    suite.NoError(err)
    suite.Contains(response, "id")
}

func (suite *APITestSuite) TestGetUserAPI() {
    // First create a user
    suite.TestCreateUserAPI()

    // Then test getting the user
    resp, err := suite.client.Get(suite.baseURL + "/api/users/1")
    suite.NoError(err)
    suite.Equal(http.StatusOK, resp.StatusCode)
}
```

### Mock Testing Suite

```go
type ServiceTestSuite struct {
    suite.Suite
    service *UserService
    mockRepo *MockUserRepository
    mockEmail *MockEmailService
}

func (suite *ServiceTestSuite) SetupTest() {
    suite.mockRepo = &MockUserRepository{}
    suite.mockEmail = &MockEmailService{}
    suite.service = NewUserService(suite.mockRepo, suite.mockEmail)
}

func (suite *ServiceTestSuite) TearDownTest() {
    // Verify all mock expectations
    suite.mockRepo.AssertExpectations(suite.T())
    suite.mockEmail.AssertExpectations(suite.T())
}

func (suite *ServiceTestSuite) TestRegisterUser_Success() {
    // Setup mocks
    suite.mockRepo.On("CreateUser", mock.AnythingOfType("*User")).Return(nil).Run(func(args mock.Arguments) {
        user := args.Get(0).(*User)
        user.ID = 123
    })
    suite.mockEmail.On("SendWelcomeEmail", "user@example.com").Return(nil)

    // Test
    user, err := suite.service.RegisterUser("user@example.com", "password")

    // Assert
    suite.NoError(err)
    suite.NotNil(user)
    suite.Equal(uint(123), user.ID)
}

func (suite *ServiceTestSuite) TestRegisterUser_EmailFailure() {
    // Setup mocks
    suite.mockRepo.On("CreateUser", mock.AnythingOfType("*User")).Return(nil)
    suite.mockEmail.On("SendWelcomeEmail", "user@example.com").Return(errors.New("SMTP error"))

    // Test
    user, err := suite.service.RegisterUser("user@example.com", "password")

    // Assert
    suite.Error(err)
    suite.Nil(user)
}
```

## Suite Organization Patterns

### Suite Inheritance

```go
// Base suite with common functionality
type BaseTestSuite struct {
    suite.Suite
    db *sql.DB
}

func (suite *BaseTestSuite) SetupSuite() {
    suite.db = setupTestDB()
}

func (suite *BaseTestSuite) TearDownSuite() {
    if suite.db != nil {
        suite.db.Close()
    }
}

// Specific test suites inherit from base
type UserTestSuite struct {
    BaseTestSuite
    userService *UserService
}

func (suite *UserTestSuite) SetupTest() {
    suite.userService = NewUserService(suite.db)
}

// Sub-suites for different concerns
type UserCreationTestSuite struct {
    UserTestSuite
}

func (suite *UserCreationTestSuite) TestCreateValidUser() {
    // Test implementation
}

func (suite *UserCreationTestSuite) TestCreateInvalidUser() {
    // Test implementation
}
```

### Suite Composition

```go
type IntegrationTestSuite struct {
    suite.Suite
    databaseSuite *DatabaseTestSuite
    apiSuite *APITestSuite
}

func (suite *IntegrationTestSuite) SetupSuite() {
    suite.databaseSuite = &DatabaseTestSuite{}
    suite.apiSuite = &APITestSuite{}

    // Setup both suites
    suite.databaseSuite.SetupSuite()
    suite.apiSuite.SetupSuite()
}

func (suite *IntegrationTestSuite) TestFullUserFlow() {
    // Use both database and API components
    // Create user in DB
    // Call API to verify user
    // Check email was sent
}
```

## Best Practices

### 1. Clear Separation of Concerns

```go
// Good: Separate suites for different layers
type RepositoryTestSuite struct { /* DB tests */ }
type ServiceTestSuite struct { /* Business logic tests */ }
type HandlerTestSuite struct { /* HTTP tests */ }

// Bad: Mixing concerns in one suite
type MixedTestSuite struct {
    // Database, service, and HTTP tests all mixed together
}
```

### 2. Appropriate Setup Levels

```go
func (suite *UserTestSuite) SetupSuite() {
    // Expensive operations: DB connection, external services
    suite.db = connectToTestDB()
}

func (suite *UserTestSuite) SetupTest() {
    // Per-test setup: Clean tables, create fixtures
    suite.cleanDatabase()
    suite.createTestUsers()
}

func (suite *UserTestSuite) TearDownTest() {
    // Per-test cleanup: Reset state
    suite.rollbackTransaction()
}
```

### 3. Test Naming Conventions

```go
func (suite *UserServiceTestSuite) TestCreateUser_ValidInput() {}
func (suite *UserServiceTestSuite) TestCreateUser_DuplicateEmail() {}
func (suite *UserServiceTestSuite) TestCreateUser_InvalidEmail() {}
func (suite *UserServiceTestSuite) TestGetUser_ExistingUser() {}
func (suite *UserServiceTestSuite) TestGetUser_NonExistingUser() {}
```

### 4. Suite Method Organization

```go
type UserServiceTestSuite struct {
    suite.Suite
    service *UserService
    mockRepo *MockUserRepository
}

// Setup methods
func (suite *UserServiceTestSuite) SetupSuite() {}
func (suite *UserServiceTestSuite) SetupTest() {}
func (suite *UserServiceTestSuite) TearDownTest() {}

// Test methods grouped by functionality
func (suite *UserServiceTestSuite) TestCreateUser() {}
func (suite *UserServiceTestSuite) TestCreateUser_Validation() {}

func (suite *UserServiceTestSuite) TestGetUser() {}
func (suite *UserServiceTestSuite) TestGetUser_NotFound() {}

func (suite *UserServiceTestSuite) TestUpdateUser() {}
func (suite *UserServiceTestSuite) TestDeleteUser() {}
```

### 5. Helper Methods

```go
type UserServiceTestSuite struct {
    suite.Suite
    service *UserService
}

// Helper methods for common test operations
func (suite *UserServiceTestSuite) createTestUser(email, name string) *User {
    user := &User{Email: email, Name: name}
    err := suite.service.CreateUser(user)
    suite.Require().NoError(err)
    return user
}

func (suite *UserServiceTestSuite) assertUserEqual(expected, actual *User) {
    suite.Equal(expected.ID, actual.ID)
    suite.Equal(expected.Email, actual.Email)
    suite.Equal(expected.Name, actual.Name)
}

// Usage in tests
func (suite *UserServiceTestSuite) TestUpdateUser() {
    user := suite.createTestUser("test@example.com", "Test User")

    user.Name = "Updated Name"
    err := suite.service.UpdateUser(user)
    suite.NoError(err)

    updated, err := suite.service.GetUser(user.ID)
    suite.NoError(err)
    suite.assertUserEqual(user, updated)
}
```

## Integration with Our Codebase

### Current Testing Structure
Our current tests are organized as individual test functions:
- `TestAuthService_Login()`
- `TestAuthRepository_CreateAuthToken()`
- etc.

### Potential Suite Usage

```go
// Example: Auth Service Suite
type AuthServiceTestSuite struct {
    suite.Suite
    service *AuthService
    mockRepo *MockAuthRepository
    mockJWT *MockJWTUtils
    testUsers []*User
}

func (suite *AuthServiceTestSuite) SetupTest() {
    suite.mockRepo = &MockAuthRepository{}
    suite.mockJWT = &MockJWTUtils{}
    suite.service = NewAuthService(suite.mockRepo, nil, suite.mockJWT, nil)
}

func (suite *AuthServiceTestSuite) TestLogin_Success() {
    // Test implementation
}

func (suite *AuthServiceTestSuite) TestLogin_InvalidCredentials() {
    // Test implementation
}

// This would replace our current individual test functions
func TestAuthServiceTestSuite(t *testing.T) {
    suite.Run(t, new(AuthServiceTestSuite))
}
```

### When to Consider Suites

1. **Complex Setup**: Tests requiring database setup, external services
2. **Shared Fixtures**: Multiple tests operating on same test data
3. **Integration Tests**: Tests spanning multiple layers
4. **Large Test Files**: When individual test files become unwieldy

## Common Pitfalls

### 1. Overusing Setup Methods

```go
// Bad: Too much setup in SetupTest
func (suite *MySuite) SetupTest() {
    // 50 lines of setup code
    suite.setupDatabase()
    suite.createUsers()
    suite.setupPermissions()
    suite.initializeServices()
    // ...
}

// Good: Keep setup focused
func (suite *MySuite) SetupTest() {
    suite.service = NewService(suite.db)
}
```

### 2. Suite Inheritance Issues

```go
// Problematic: Deep inheritance hierarchies
type BaseSuite struct { suite.Suite }
type DatabaseSuite struct { BaseSuite }
type APISuite struct { DatabaseSuite }
type IntegrationSuite struct { APISuite }

// Better: Composition over inheritance
type IntegrationSuite struct {
    suite.Suite
    *DatabaseSuite
    *APISuite
}
```

### 3. Ignoring Test Isolation

```go
// Bad: Tests affect each other
func (suite *MySuite) TestCreateUser() {
    suite.service.CreateUser(&User{Email: "test@example.com"})
}

func (suite *MySuite) TestGetUser() {
    // Assumes TestCreateUser ran first
    user, _ := suite.service.GetUser("test@example.com")
    suite.NotNil(user)
}

// Good: Each test is independent
func (suite *MySuite) SetupTest() {
    suite.cleanDatabase()
}

func (suite *MySuite) TestCreateUser() {
    user := &User{Email: "test@example.com"}
    err := suite.service.CreateUser(user)
    suite.NoError(err)
}

func (suite *MySuite) TestGetUser() {
    // Create user for this specific test
    user := suite.createTestUser("test@example.com")
    found, err := suite.service.GetUser(user.ID)
    suite.NoError(err)
    suite.Equal(user.ID, found.ID)
}
```

## Conclusion

Testify suites are powerful for organizing complex tests with shared setup/teardown logic. Use them when you have:

- Shared test fixtures or setup code
- Multiple related test methods
- Complex initialization requirements
- Integration testing scenarios

For simple unit tests, individual test functions are often sufficient and more straightforward.