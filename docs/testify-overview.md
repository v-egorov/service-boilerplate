# Testify Testing Framework Documentation

## Overview

This documentation suite provides comprehensive guidance on using the testify testing framework for Go applications. Testify is a powerful toolkit that enhances Go's built-in testing capabilities with assertions, mocking, suite testing, and HTTP testing utilities.

## Documentation Index

### Core Components

1. **[Testify Mocking Guide](testify-mocking-guide.md)**
   - When to use testify mocking vs. manual mocks
   - Complete examples of call verification and interaction testing
   - Comparison with manual mocking approaches
   - Integration recommendations for our codebase

2. **[Testify Assertions Guide](testify-assertions-guide.md)**
   - `assert` vs `require` usage patterns
   - Common assertion types (equality, nil, collections, errors)
   - Advanced patterns and best practices
   - Integration with our existing test patterns

3. **[Testify Suite Testing Guide](testify-suite-testing-guide.md)**
   - Structured test organization with shared setup/teardown
   - Database and HTTP testing suites
   - Best practices for suite organization
   - When to use suites vs. individual tests

4. **[Testify HTTP Testing Guide](testify-http-testing-guide.md)**
   - HTTP request/response testing patterns
   - Middleware testing techniques
   - Integration testing with databases
   - API lifecycle testing examples

## Quick Reference

### Installation

```bash
go get github.com/stretchr/testify
```

### Basic Usage

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "github.com/stretchr/testify/mock"
)

func TestBasic(t *testing.T) {
    // Assertions
    assert.Equal(t, expected, actual)
    assert.NoError(t, err)
    assert.NotNil(t, result)

    // Or with require (stops on first failure)
    require.NoError(t, err)
    require.NotNil(t, result)
}

func TestWithMock(t *testing.T) {
    mock := &MyMock{}
    mock.On("Method", "arg").Return("result", nil).Once()

    result, err := service.Method("arg")

    assert.NoError(t, err)
    assert.Equal(t, "result", result)
    mock.AssertExpectations(t)
}

type MyTestSuite struct {
    suite.Suite
    // shared state
}

func (suite *MyTestSuite) TestSomething() {
    suite.Equal(expected, actual)
}

func TestMyTestSuite(t *testing.T) {
    suite.Run(t, new(MyTestSuite))
}
```

## Decision Framework

### When to Use Testify Components

| Component | Use When | Example Scenarios |
|-----------|----------|-------------------|
| **Assertions** | Always (replaces manual `t.Errorf()`) | Value comparisons, error checking, collection validation |
| **Mocking** | Call verification needed, simple interfaces | External API testing, interaction verification |
| **Suites** | Shared setup/teardown, complex test organization | Database testing, API integration tests |
| **HTTP Testing** | API endpoint testing, middleware validation | REST API testing, authentication flows |

### When to Use Manual Approaches

| Component | Use When | Our Current Usage |
|-----------|----------|-------------------|
| **Manual Mocks** | Complex control needed, performance critical | Repository database testing |
| **Individual Tests** | Simple, independent tests | Basic unit tests |
| **Built-in httptest** | Basic HTTP testing without testify features | Simple handler tests |

## Integration with Our Codebase

### Current Test Structure

```
services/auth-service/
├── internal/
│   ├── handlers/
│   │   └── auth_handler_test.go      # HTTP handler tests
│   ├── services/
│   │   └── auth_service_test.go      # Business logic tests
│   └── repository/
│       └── auth_repository_test.go   # Database layer tests
```

### Test Coverage Status

- **Handlers**: 74.2% coverage (26/26 methods tested)
- **Services**: 71.4% coverage (25/25 methods tested)
- **Repository**: 71.0% coverage (18/30 methods tested)

### Recommended Patterns

#### Repository Layer (Current: Manual Mocks)
```go
// Continue with manual mocks for complex DB operations
func TestAuthRepository_GetUserRoles(t *testing.T) {
    mockDB := &MockDBPool{}
    // Full control over pgx.Rows iteration
}
```

#### Service Layer (Potential: Testify Mocks)
```go
// Consider testify mocks for external dependencies
func TestAuthService_SendNotification(t *testing.T) {
    mockNotifier := &MockNotifier{}
    mockNotifier.On("SendEmail", "user@example.com").Return(nil)
    // Built-in call verification
}
```

#### Handler Layer (Current: httptest + testify)
```go
// Continue with current pattern
func TestAuthHandler_Login(t *testing.T) {
    req := httptest.NewRequest("POST", "/auth/login", nil)
    rr := httptest.NewRecorder()
    // testify assertions for response validation
}
```

## Best Practices Summary

### 1. Choose Appropriate Tools
- **Assertions**: Always use testify (assert/require)
- **Mocking**: Manual for complex DB, testify for simple interfaces
- **Suites**: For shared setup, individual tests for simple cases
- **HTTP**: httptest + testify assertions

### 2. Test Organization
- **Group related tests** with `t.Run()`
- **Use suites** for complex setup scenarios
- **Keep tests isolated** and independent
- **Document test intent** clearly

### 3. Assertion Patterns
- **Use `require`** for critical setup that must succeed
- **Use `assert`** for independent checks
- **Provide descriptive messages** for failures
- **Test error conditions** thoroughly

### 4. Mocking Guidelines
- **Manual mocks** for complex database operations
- **Testify mocks** for call verification and simple interfaces
- **Always verify expectations** with `AssertExpectations(t)`
- **Keep mocks focused** on the interface being tested

### 5. HTTP Testing
- **Always check status codes** and content types
- **Close response bodies** to prevent memory leaks
- **Test error responses** and edge cases
- **Use suites** for complex API testing

## Migration Path

### Phase 1: Assessment (Current)
- ✅ Established manual mocking patterns for repository testing
- ✅ Consistent testify assertion usage
- ✅ Basic HTTP testing with httptest

### Phase 2: Selective Adoption
- Consider testify mocks for new service layer external dependencies
- Evaluate suite testing for complex integration scenarios
- Enhance HTTP testing patterns as needed

### Phase 3: Optimization
- Standardize mocking approaches based on proven patterns
- Implement suite testing where it reduces boilerplate
- Optimize test performance and maintainability

## Conclusion

Testify provides excellent tools for Go testing, but the right approach depends on the context:

- **Use testify assertions** in all tests (they're superior to manual checks)
- **Use manual mocks** for complex database operations (like our current repository tests)
- **Use testify mocks** for simple interfaces where call verification is valuable
- **Use suite testing** when shared setup/teardown provides clear benefits
- **Use HTTP testing patterns** consistently across API tests

Our current approach is well-suited to our needs, with room for selective testify adoption as new testing scenarios arise.