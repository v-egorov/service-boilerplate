## 1. Create common/errors package — infra.go

- [x] 1.1 Created `common/errors/infra.go` with `ErrNotFound`, `ErrAlreadyExists`, `ErrInvalidInput` sentinel variables
- [x] 1.2 Package doc explains usage: import → wrap with `%w` → match via `errors.Is()`

## 2. Create common/errors package — auth.go

- [x] 2.1 Created `common/errors/auth.go` with `ErrUnauthorized`, `ErrForbidden` sentinel variables
- [x] 2.2 Auth sentinels documented in package-level comments (infra.go covers shared pattern)

## 3. Verify compilation and tests pass (no service changes)

- [x] 3.1 Build passes: `go build ./common/errors/...` exits clean
- [x] 3.2 All existing service tests pass (objects ✅, user ✅, auth ✅, mcp-server ✅)
- [x] 3.3 Unit tests: infra_test.go with 5 tests (non-nil, Error() strings, self-match, wrapped matching, cross-mismatch)

## 4. Validate integration with existing codebase

- [x] 4.1 No import cycles: only imports `fmt`, zero service dependencies
- [x] 4.2 `go vet ./common/errors/...` passes clean
- [x] 4.3 All services share module `github.com/v-egorov/service-boilerplate` — importable without go.mod changes
