## 1. Remove ErrAlreadyExists from repository and handler dispatcher

- [x] 1.1 In `repository/interfaces.go`: removed `ErrAlreadyExists` var definition (kept all other vars)
- [x] 2.1 In `handlers/error.go`: removed the `case errors.Is(err, repository.ErrAlreadyExists)` branch from HandleError()
- [x] 3.1 In `handlers/error_test.go`: removed test entry referencing ErrAlreadyExists

## 2. Migrate ErrNotFound and ErrInvalidInput to common/errors in repository layer (via aliases)

- [x] 4.1 Added `"github.com/v-egorov/service-boilerplate/common/errors"` import to `repository/interfaces.go`
- [x] 4.2 Replaced `ErrNotFound = fmt.Errorf("resource not found")` with alias: `var ErrNotFound = errors.ErrNotFound`
- [x] 5.1 Replaced `ErrInvalidInput = fmt.Errorf("invalid input")` with alias: `var ErrInvalidInput = errors.ErrInvalidInput`

**Note:** Using aliases (not direct replacement) preserves backward compatibility — all existing service-layer code that references `repository.ErrNotFound` or `repository.ErrInvalidInput` continues working because the alias is the same pointer value. Service-layer changes are NOT required.

## 3. Build, test, and verify

- [x] 12.1 Build passes: `make build-objects-service`
- [x] 12.2 Tests pass: all 7 packages pass (handlers 0.009s, repository 0.006s)
- [x] 12.3 Verified zero references to ErrAlreadyExists remain in the codebase
