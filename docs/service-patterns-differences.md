# Service Patterns Differences

This document identifies differences between services that should be standardized in future improvements.

---

## Overview

The boilerplate contains two service implementations:
- **user-service** - User management service
- **objects-service** - Object/taxonomy management service

Both implement similar functionality but have some inconsistencies that should be addressed in future development.

---

## Identified Differences

### 1. Error Handling Approach

| Aspect | user-service | objects-service |
|--------|--------------|-----------------|
| Implementation | Custom error types | Package-level error variables |
| Location | `models/errors.go` | Package vars in each layer |
| Error types | `ValidationError`, `NotFoundError`, `ConflictError`, `InternalError` | `ErrNotFound`, `ErrAlreadyExists`, etc. |

**user-service approach:**
```go
// internal/models/errors.go
type NotFoundError struct {
    Resource string
    ID       string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}
```

**objects-service approach:**
```go
// internal/services/user_service.go
var (
    ErrUserNotFound = errors.New("user not found")
    ErrUserExists   = errors.New("user already exists")
)
```

**Recommendation:** Choose one approach and standardize across all services.

---

### 2. Database Schema Qualification

| Aspect | user-service | objects-service |
|--------|--------------|-----------------|
| Table references | `app.users` | `objects_service.object_types` |
| Consistency | Mixed | Consistent |

**user-service:**
```sql
SELECT * FROM app.users WHERE id = $1
```

**objects-service:**
```sql
SELECT * FROM objects_service.object_types WHERE id = $1
```

**Recommendation:** Always use schema-qualified table names (objects-service approach).

---

### 3. Constructor Patterns

| Aspect | user-service | objects-service |
|--------|--------------|-----------------|
| Multiple constructors | ✅ Yes | ❌ Single |
| Test interfaces | ✅ Yes | Partial |

**user-service:**
```go
// Production
func NewUserService(repo *repository.UserRepository, logger *logrus.Logger) *UserService

// Testing
func NewUserServiceWithInterface(repo UserRepositoryInterface, logger *logrus.Logger) *UserService
```

**Recommendation:** Use two-constructor pattern for better testability.

---

### 4. Repository Interface Location

| Aspect | user-service | objects-service |
|--------|--------------|-----------------|
| Interface definition | In repository package | In repository package |
| Testing interface | Duplicated in services | Duplicated in services |

**Recommendation:** Keep interfaces in repository package, reference from services.

---

### 5. Database Tracing

| Aspect | user-service | objects-service |
|--------|--------------|-----------------|
| DB tracing | ✅ Uses wrappers | ✅ Uses wrappers |
| Consistency | Good | Good |

---

### 6. Business Operation Tracing

| Aspect | user-service | objects-service |
|--------|--------------|-----------------|
| Manual spans | Not implemented | Not implemented |
| Gap | Both services | Both services |

**Note:** Neither service implements business operation tracing (see [tracing-implementation-guide.md](tracing-implementation-guide.md)).

---

### 7. Handler Error Mapping

| Aspect | user-service | objects-service |
|--------|--------------|-----------------|
| Centralized | ✅ Yes | ✅ Yes |
| Implementation | `handleError` method | `handleError` method |

Both services follow similar patterns for error handling.

---

## Future Improvements Checklist

- [ ] Standardize error handling approach across services
- [ ] Add two-constructor pattern where missing
- [ ] Ensure schema-qualified queries everywhere
- [ ] Implement business operation tracing in all services
- [ ] Create shared error types in common package (optional)

---

## References

- [Service Patterns Reference](service-patterns-reference.md)
- [Tracing Implementation Guide](tracing-implementation-guide.md)
