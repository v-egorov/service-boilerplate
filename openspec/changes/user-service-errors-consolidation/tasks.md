## 1. Create common/errors/typed.go with shared typed error structs

- [ ] 1.1 Write `common/errors/typed.go` containing ValidationError, ConflictError, NotFoundError, InternalError (copied verbatim from user-service/internal/models/errors.go)
- [ ] 1.2 Include constructor functions: NewValidationError, NewConflictError, NewNotFoundError, NewInternalError
- [ ] 1.3 Verify build: `go build ./common/...`

## 2. Update user-service models/errors.go with re-export aliases

- [ ] 2.1 Remove type definitions for ValidationError, ConflictError, NotFoundError, InternalError from `user-service/internal/models/errors.go`
- [ ] 2.2 Add re-export aliases: `type ValidationError = errors.ValidationError` etc., pointing to common/errors types
- [ ] 2.3 Keep auth-specific types (if any) in models/errors.go — currently none

## 3. Update user-service service-layer references (~68 refs)

- [ ] 3.1 In `services/user-service/internal/services/user_service.go`: replace all `models.NewValidationError` → `errors.NewValidationError`, `models.NewConflictError` → `errors.NewConflictError`, `models.NewNotFoundError` → `errors.NewNotFoundError`
- [ ] 3.2 Add `"github.com/v-egorov/service-boilerplate/common/errors"` import to user_service.go

## 4. Update handler dispatcher imports

- [ ] 4.1 In `user-service/internal/handlers/user_handler.go`: update type-switch cases to reference common/errors types (the alias makes this work automatically, but the switch must still compile with same struct identity)
- [ ] 4.2 Verify that `handleServiceError` method still compiles — aliased types have identical struct identity so type-switch continues working

## 5. Build, test, and verify

- [ ] 5.1 Run `make build-user-service` — must compile without errors
- [ ] 5.2 Run `make test-user-service` — all tests pass (pay attention to handler tests that use typed error types)
- [ ] 5.3 Verify no remaining local definitions of ValidationError/ConflictError/NotFoundError/InternalError in user-service/models/errors.go (only re-export aliases should remain)
