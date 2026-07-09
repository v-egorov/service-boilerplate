## 1. Create common/errors package — infra.go

- [ ] 1.1 Create `common/errors/infra.go` with `ErrNotFound`, `ErrAlreadyExists`, `ErrInvalidInput` sentinel variables using `fmt.Errorf()`
- [ ] 1.2 Add a brief package-level doc comment explaining the purpose and usage pattern (import, wrap via `%w`)

## 2. Create common/errors package — auth.go

- [ ] 2.1 Create `common/errors/auth.go` with `ErrUnauthorized`, `ErrForbidden` sentinel variables using `fmt.Errorf()`
- [ ] 2.2 Add brief doc comment for the auth-specific sentinels

## 3. Verify compilation and tests pass (no service changes)

- [ ] 3.1 Run `go build ./common/errors/...` to confirm package compiles with no errors
- [ ] 3.2 Run `make test-all` or equivalent to verify all existing service tests still pass — this delta must be zero-touch for existing services
- [ ] 3.3 Write a simple unit test in `common/errors/infra_test.go` verifying each sentinel is non-nil, has correct Error() string, and matches itself via `errors.Is()`

## 4. Validate integration with existing codebase

- [ ] 4.1 Verify no import cycle — common/errors must not import any service-specific packages
- [ ] 4.2 Run `go vet ./common/...` to catch any issues
- [ ] 4.3 Confirm the package is importable by all three services without modifying their go.mod (it's already in the same module)
