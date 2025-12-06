# Testify Assertions Guide

## Overview

Testify provides powerful assertion functions through the `assert` and `require` packages. This guide covers when to use each, common patterns, and best practices for effective testing.

## References

- **Testify Repository**: [github.com/stretchr/testify](https://github.com/stretchr/testify)
- **Assert Package**: [pkg.go.dev/github.com/stretchr/testify/assert](https://pkg.go.dev/github.com/stretchr/testify/assert)
- **Require Package**: [pkg.go.dev/github.com/stretchr/testify/require](https://pkg.go.dev/github.com/stretchr/testify/require)

## Assert vs Require

### `assert` Package
- **Behavior**: Test continues even if assertion fails
- **Use When**: You want to collect multiple failures or test doesn't depend on previous assertions
- **Example**: UI testing where multiple elements can be checked independently

### `require` Package
- **Behavior**: Test stops immediately on first failure (calls `t.Fatal()`)
- **Use When**: Subsequent assertions depend on previous ones, or early failure is preferred
- **Example**: Database setup failures should stop the test immediately

## Common Assertion Patterns

### Equality Assertions

```go
// Basic equality
assert.Equal(t, expected, actual)
assert.NotEqual(t, unexpected, actual)

// Deep equality for complex types
assert.EqualValues(t, expected, actual)

// String equality
assert.Equal(t, "expected string", result)

// Numeric equality with tolerance
assert.InDelta(t, 3.14159, math.Pi, 0.01)
```

### Nil and Zero Value Assertions

```go
// Nil checks
assert.Nil(t, result)
assert.NotNil(t, user)

// Zero value checks
assert.Zero(t, count)
assert.NotZero(t, balance)

// Empty checks
assert.Empty(t, slice)
assert.NotEmpty(t, users)
```

### Boolean and Truth Assertions

```go
// Boolean values
assert.True(t, isValid)
assert.False(t, hasErrors)

// Truthy/falsy (Go style)
assert.Equal(t, true, success)
```

### Collection Assertions

```go
// Slice and array assertions
assert.Len(t, users, 5)
assert.Contains(t, names, "Alice")
assert.NotContains(t, errors, "critical")

// Map assertions
assert.Contains(t, userMap, "user123")
assert.Equal(t, "admin", userMap["user123"])

// Subset checks
assert.Subset(t, []string{"a", "b", "c"}, []string{"a", "b"})
```

### Error Assertions

```go
// Error presence
assert.Error(t, err)
assert.NoError(t, result)

// Specific error types
assert.ErrorIs(t, err, sql.ErrNoRows)
assert.ErrorAs(t, err, &customError)

// Error messages
assert.ErrorContains(t, err, "not found")
assert.Contains(t, err.Error(), "validation failed")
```

### Type Assertions

```go
// Type checking
assert.IsType(t, User{}, result)
assert.IsType(t, &User{}, result)

// Kind checking
assert.Equal(t, reflect.Slice, reflect.TypeOf(result).Kind())
```

## Advanced Assertion Patterns

### Custom Assertions

```go
// Custom assertion function
func assertValidUser(t *testing.T, user *User) {
    assert.NotNil(t, user)
    assert.NotEmpty(t, user.ID)
    assert.NotEmpty(t, user.Email)
    assert.Contains(t, user.Email, "@")
}

// Usage
func TestCreateUser(t *testing.T) {
    user, err := createUser(request)
    require.NoError(t, err)
    assertValidUser(t, user)
}
```

### Conditional Assertions

```go
// Only assert if condition is met
if assert.NotNil(t, result) {
    assert.Equal(t, "expected", result.Value)
}

// Assert based on test setup
if isIntegrationTest {
    assert.NotNil(t, dbConnection)
} else {
    assert.Nil(t, dbConnection)
}
```

### Table-Driven Tests with Assertions

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name     string
        email    string
        expected bool
    }{
        {"valid email", "user@example.com", true},
        {"invalid email", "not-an-email", false},
        {"empty email", "", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := validateEmail(tt.email)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

## HTTP Testing Assertions

```go
func TestAPIEndpoint(t *testing.T) {
    // Make HTTP request
    resp, err := http.Get("http://localhost:8080/api/users")
    require.NoError(t, err)
    defer resp.Body.Close()

    // Status code assertions
    assert.Equal(t, http.StatusOK, resp.StatusCode)

    // Header assertions
    assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

    // Body assertions
    var users []User
    err = json.NewDecoder(resp.Body).Decode(&users)
    assert.NoError(t, err)
    assert.NotEmpty(t, users)
}
```

## Database Testing Assertions

```go
func TestUserRepository_GetUser(t *testing.T) {
    // Setup
    userID := uuid.New()

    // Execute
    user, err := repo.GetUser(context.Background(), userID)

    // Assertions
    require.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, userID, user.ID)
    assert.NotEmpty(t, user.Email)
    assert.True(t, user.CreatedAt.After(time.Now().Add(-time.Hour)))
}
```

## Best Practices

### 1. Choose Appropriate Package
```go
// Use require for setup that must succeed
db, err := setupDatabase()
require.NoError(t, err)

// Use assert for independent checks
assert.Equal(t, "expected", result.Field1)
assert.Equal(t, "expected", result.Field2)
```

### 2. Descriptive Failure Messages
```go
// Bad - generic message
assert.Equal(t, expected, actual)

// Good - descriptive message
assert.Equal(t, expected, actual, "user balance should match expected amount")
```

### 3. Group Related Assertions
```go
// Group related checks together
user, err := createUser(request)
require.NoError(t, err)
require.NotNil(t, user)

// User field validations
assert.NotEmpty(t, user.ID)
assert.Equal(t, request.Email, user.Email)
assert.True(t, user.CreatedAt.After(time.Now().Add(-time.Minute)))
```

### 4. Use Appropriate Assertions
```go
// Prefer specific assertions over generic ones
assert.NotNil(t, result)      // Good
assert.True(t, result != nil) // Less clear

assert.Empty(t, slice)        // Good
assert.Equal(t, 0, len(slice)) // Less expressive
```

### 5. Test Error Conditions Thoroughly
```go
func TestCreateUser_ValidationErrors(t *testing.T) {
    tests := []struct {
        name        string
        request     CreateUserRequest
        expectedErr string
    }{
        {"empty email", CreateUserRequest{Password: "pass"}, "email is required"},
        {"weak password", CreateUserRequest{Email: "user@example.com", Password: "123"}, "password too weak"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := createUser(tt.request)
            assert.Error(t, err)
            assert.Contains(t, err.Error(), tt.expectedErr)
        })
    }
}
```

## Integration with Our Codebase

### Current Usage Patterns
Our existing tests use testify assertions effectively:
- `assert.Equal()` for value comparisons
- `assert.NoError()` for error checking
- `assert.NotNil()` for nil checks
- `require.NoError()` for critical setup operations

### Recommended Patterns
```go
// Service layer tests
func TestAuthService_Login(t *testing.T) {
    // Critical setup - use require
    user, err := setupTestUser()
    require.NoError(t, err)

    // Business logic assertions - use assert
    token, err := service.Login(user.Email, "password")
    assert.NoError(t, err)
    assert.NotNil(t, token)
    assert.NotEmpty(t, token.AccessToken)
}

// Repository layer tests
func TestUserRepository_GetUser(t *testing.T) {
    // Setup - use require for database operations
    userID, err := setupTestUser()
    require.NoError(t, err)

    // Query assertions - use assert
    user, err := repo.GetUser(userID)
    assert.NoError(t, err)
    assert.Equal(t, userID, user.ID)
}
```

## Common Pitfalls

### 1. Overusing Require
```go
// Bad - test stops on first failure
require.Equal(t, "expected", result.Field1)
require.Equal(t, "expected", result.Field2) // Never reached if Field1 fails

// Good - collect all failures
assert.Equal(t, "expected", result.Field1)
assert.Equal(t, "expected", result.Field2)
```

### 2. Ignoring Error Details
```go
// Bad - only checks error presence
assert.Error(t, err)

// Good - checks specific error
assert.ErrorIs(t, err, ErrNotFound)
assert.Contains(t, err.Error(), "user not found")
```

### 3. Testing Implementation Details
```go
// Bad - tests internal implementation
assert.Equal(t, 3, service.callCount)

// Good - tests behavior
result, err := service.Process(input)
assert.NoError(t, err)
assert.Equal(t, expectedOutput, result)
```

## Conclusion

Testify assertions provide a rich set of tools for comprehensive testing. Choose between `assert` and `require` based on whether test continuation is desired. Use descriptive messages and group related assertions for maintainable tests.