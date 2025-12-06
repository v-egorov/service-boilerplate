# Testify Mocking Approach for Go Testing

## Overview

This document describes the testify mocking approach for unit testing in Go, including when to use it, how to implement it, and examples. Testify provides a powerful mocking framework that integrates seamlessly with its assertion library.

## References

- **Testify Repository**: [github.com/stretchr/testify](https://github.com/stretchr/testify)
- **Mock Package Documentation**: [pkg.go.dev/github.com/stretchr/testify/mock](https://pkg.go.dev/github.com/stretchr/testify/mock)
- **Testify Documentation**: [docs.gofiber.dev](https://docs.gofiber.dev) (general testing docs)

## When to Use Testify Mocking

### Beneficial Scenarios

1. **Call Verification**: When you need to verify that methods were called with specific arguments
2. **Interaction Testing**: Testing workflows with multiple method calls and sequences
3. **Simple Interface Mocking**: For straightforward interfaces with few methods
4. **Team Consistency**: When testify mocking is the standard in your team/organization

### When Manual Mocks Are Better

1. **Complex Database Operations**: Like our current repository testing with complex result sets
2. **Full Control Requirements**: When you need complete control over mock behavior
3. **Performance-Critical Testing**: Manual mocks have less overhead
4. **Legacy Integration**: Easier integration with existing codebases

## Testify Mocking Basics

### Core Concepts

- **Mock Embedding**: Embed `mock.Mock` in your mock struct
- **Expectation Setting**: Use `On().Return()` to define expected calls
- **Call Verification**: Use `AssertExpectations(t)` for automatic verification

### Basic Structure

```go
import "github.com/stretchr/testify/mock"

// Mock implementation
type MyMock struct {
    mock.Mock
}

func (m *MyMock) MyMethod(arg string) (string, error) {
    args := m.Called(arg)
    return args.String(0), args.Error(1)
}

// Usage in test
func TestSomething(t *testing.T) {
    mock := &MyMock{}
    mock.On("MyMethod", "input").Return("output", nil).Once()

    // Use mock in your code
    result, err := someFunction(mock)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "output", result)

    // Verify all expectations were met
    mock.AssertExpectations(t)
}
```

## Detailed Examples

### Example 1: Call Verification & Interaction Testing

**Scenario**: Testing a service that calls multiple external APIs and you need to ensure each API was called exactly once with correct parameters.

```go
// Interface to mock
type ExternalAPI interface {
    ValidateUser(userID string) error
    SendNotification(userID, message string) error
    LogActivity(userID, action string) error
}

// Testify Mock
type MockExternalAPI struct {
    mock.Mock
}

func (m *MockExternalAPI) ValidateUser(userID string) error {
    args := m.Called(userID)
    return args.Error(0)
}

func (m *MockExternalAPI) SendNotification(userID, message string) error {
    args := m.Called(userID, message)
    return args.Error(0)
}

func (m *MockExternalAPI) LogActivity(userID, action string) error {
    args := m.Called(userID, action)
    return args.Error(0)
}

// Test implementation
func TestUserService_OnboardUser(t *testing.T) {
    mockAPI := &MockExternalAPI{}

    // Set up expectations - specific call verification
    mockAPI.On("ValidateUser", "user123").Return(nil).Once()
    mockAPI.On("SendNotification", "user123", "Welcome to our platform!").Return(nil).Once()
    mockAPI.On("LogActivity", "user123", "user_onboarded").Return(nil).Once()

    service := NewUserService(mockAPI)
    err := service.OnboardUser("user123")

    // Assert business logic
    assert.NoError(t, err)

    // Verify ALL expected calls were made exactly as specified
    mockAPI.AssertExpectations(t)
}
```

**Benefits**:
- Automatic verification that each method was called exactly once
- Exact argument matching
- Clear failure messages if expectations aren't met

### Example 2: Simple Interface Mocking

**Scenario**: Mocking a simple notification service for testing user registration.

```go
// Interface
type Notifier interface {
    SendEmail(to, subject, body string) error
    SendSMS(to, message string) error
}

// Testify Mock
type MockNotifier struct {
    mock.Mock
}

func (m *MockNotifier) SendEmail(to, subject, body string) error {
    args := m.Called(to, subject, body)
    return args.Error(0)
}

func (m *MockNotifier) SendSMS(to, message string) error {
    args := m.Called(to, message)
    return args.Error(0)
}

// Test
func TestAuthService_RegisterUser(t *testing.T) {
    mockNotifier := &MockNotifier{}

    // Simple expectation - just return success
    mockNotifier.On("SendEmail", "user@example.com", "Welcome!", mock.AnythingOfType("string")).Return(nil)
    mockNotifier.On("SendSMS", "+1234567890", "Your verification code is 123456").Return(nil)

    service := NewAuthService(mockNotifier)
    user, err := service.RegisterUser("user@example.com", "+1234567890")

    assert.NoError(t, err)
    assert.NotNil(t, user)

    mockNotifier.AssertExpectations(t)
}
```

**Benefits**:
- Less boilerplate than manual mocks
- `mock.AnythingOfType()` for flexible argument matching
- Standard testify patterns

### Example 3: Complex Interaction Testing

**Scenario**: Testing a payment processor with multiple steps and error handling.

```go
type PaymentGateway interface {
    ValidatePaymentMethod(methodID string) error
    AuthorizePayment(methodID string, amount float64) (string, error)  // returns transaction ID
    CapturePayment(transactionID string, amount float64) error
    RefundPayment(transactionID string, amount float64) error
}

type MockPaymentGateway struct {
    mock.Mock
}

func (m *MockPaymentGateway) ValidatePaymentMethod(methodID string) error {
    args := m.Called(methodID)
    return args.Error(0)
}

func (m *MockPaymentGateway) AuthorizePayment(methodID string, amount float64) (string, error) {
    args := m.Called(methodID, amount)
    return args.String(0), args.Error(1)
}

func (m *MockPaymentGateway) CapturePayment(transactionID string, amount float64) error {
    args := m.Called(transactionID, amount)
    return args.Error(0)
}

func (m *MockPaymentGateway) RefundPayment(transactionID string, amount float64) error {
    args := m.Called(transactionID, amount)
    return args.Error(0)
}

func TestPaymentService_ProcessPayment(t *testing.T) {
    mockGateway := &MockPaymentGateway{}

    // Complex workflow testing
    mockGateway.On("ValidatePaymentMethod", "pm_123").Return(nil).Once()
    mockGateway.On("AuthorizePayment", "pm_123", 99.99).Return("txn_456", nil).Once()
    mockGateway.On("CapturePayment", "txn_456", 99.99).Return(nil).Once()

    service := NewPaymentService(mockGateway)
    err := service.ProcessPayment("pm_123", 99.99)

    assert.NoError(t, err)
    mockGateway.AssertExpectations(t)
}

func TestPaymentService_ProcessPayment_ValidationFailure(t *testing.T) {
    mockGateway := &MockPaymentGateway{}

    // Test error path - validation fails
    mockGateway.On("ValidatePaymentMethod", "pm_invalid").Return(errors.New("invalid payment method")).Once()
    // Note: Authorize and Capture should NOT be called

    service := NewPaymentService(mockGateway)
    err := service.ProcessPayment("pm_invalid", 99.99)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "invalid payment method")

    // Verify that subsequent calls were NOT made
    mockGateway.AssertExpectations(t)
}
```

**Benefits**:
- Clear testing of complex workflows
- Easy verification of call sequences
- Built-in support for different return values and errors

## Advanced Features

### Argument Matchers

```go
// Exact matching
mock.On("Method", "exact_arg").Return(result)

// Any value of type
mock.On("Method", mock.AnythingOfType("string")).Return(result)

// Custom matchers
mock.On("Method", mock.MatchedBy(func(arg string) bool {
    return len(arg) > 5
})).Return(result)
```

### Call Counts

```go
// Exactly once
mock.On("Method").Return(result).Once()

// At least once
mock.On("Method").Return(result).Maybe()

// Specific count
mock.On("Method").Return(result).Times(3)

// Any number of times
mock.On("Method").Return(result).AnyTimes()
```

### Return Values

```go
// Single return value
mock.On("Method").Return("result")

// Multiple return values
mock.On("Method").Return("result", nil)

// Error returns
mock.On("Method").Return("", errors.New("error"))
```

## Comparison with Manual Mocks

| Aspect | Testify Mocks | Manual Mocks (Our Current) |
|--------|---------------|---------------------------|
| **Setup** | `mock.On().Return()` | Function pointers |
| **Call Verification** | Built-in `AssertExpectations()` | Manual tracking |
| **Flexibility** | Medium (framework constraints) | Very High (full control) |
| **Boilerplate** | Low | Medium-High |
| **Learning Curve** | Low | None (custom approach) |
| **Complex Scenarios** | Challenging | Excellent |
| **Simple Scenarios** | Excellent | Good |

## Integration with Our Codebase

### Current Repository Testing
- **Recommendation**: Continue with manual mocks
- **Reason**: Complex database operations require fine-grained control over `pgx.Rows`, transactions, and error scenarios

### Service Layer Testing
- **Recommendation**: Consider testify mocks for external dependencies
- **Example**: Mocking HTTP clients, external APIs, notification services

### New Features
- **Recommendation**: Evaluate based on complexity
- **Simple interfaces**: Use testify mocks
- **Complex interactions**: Consider manual mocks

## Best Practices

1. **Use Appropriate Tools**: Choose based on complexity and verification needs
2. **Clear Expectations**: Set specific, clear expectations for each test
3. **Verify Calls**: Always use `AssertExpectations(t)` to ensure all expected calls were made
4. **Test Error Paths**: Don't just test happy paths - test error conditions
5. **Documentation**: Document mock expectations clearly in test comments

## Conclusion

Testify mocking is excellent for:
- Call verification and interaction testing
- Simple interface mocking
- Team consistency and standards
- Scenarios where built-in verification is valuable

Manual mocks (like our current approach) are better for:
- Complex database operations
- Full behavioral control requirements
- Performance-critical testing
- Integration with existing codebases

Choose the approach that best fits your testing needs and team preferences.