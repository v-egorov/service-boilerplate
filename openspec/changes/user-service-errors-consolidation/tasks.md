## 1. Create common/errors/typed.go with shared typed error structs

- [x] 1.1 Write `common/errors/typed.go` containing ValidationError, ConflictError, NotFoundError, InternalError (copied verbatim from user-service/internal/models/errors.go)
- [x] 1.2 Include constructor functions: NewValidationError, NewConflictError, NewNotFoundError, NewInternalError
- [x] 1.3 Verify build: `go build ./common/...`

## 2. Update user-service models/errors.go with re-export aliases (type + function)

- [x] 2.1 Remove type definitions for ValidationError, ConflictError, NotFoundError, InternalError from `user-service/internal/models/errors.go`
- [x] 2.2 Add type aliases: `type ValidationError = errors.ValidationError` etc.
- [x] 2.3 Add constructor function aliases: `var NewValidationError = errors.NewValidationError` etc. (eliminates need for service-layer import changes)
- [x] 2.4 Verify no remaining local definitions in models/errors.go — only re-export aliases remain

## 3. Build, test, and verify

- [x] 3.1 Build passes: `go build ./services/user-service/...` (zero service-layer changes needed thanks to function aliases)
- [x] 3.2 Tests pass: all 5 packages pass (handlers 0.012s, models 0.011s, services 0.107s, repository 0.008s)
- [x] 3.3 Handler type-switches on `models.ValidationError` etc. continue working — Go type aliases preserve struct identity
- [x] 3.4 Zero service-layer file changes required (import paths unchanged)
