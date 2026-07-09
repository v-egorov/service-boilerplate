## 1. Remove ErrAlreadyExists from repository and handler dispatcher

- [ ] 1.1 In `repository/interfaces.go`: remove the `ErrAlreadyExists` var definition (keep all other vars)
- [ ] 2.1 In `handlers/error.go`: remove the `case errors.Is(err, repository.ErrAlreadyExists)` branch from HandleError()
- [ ] 3.1 In `handlers/error_test.go`: remove test entries referencing ErrAlreadyExists

## 2. Migrate ErrNotFound to common/errors in repository layer

- [ ] 4.1 Add `"github.com/v-egorov/service-boilerplate/common/errors"` import to `repository/interfaces.go`
- [ ] 4.2 Replace `ErrNotFound = fmt.Errorf("resource not found")` with a type alias: `var ErrNotFound = errors.ErrNotFound` (keeps existing code working) OR remove and update all return sites directly — prefer direct replacement for clarity

## 3. Migrate ErrInvalidInput to common/errors in repository layer

- [ ] 5.1 Replace `ErrInvalidInput = fmt.Errorf("invalid input")` with reference to `common/errors.ErrInvalidInput` (same approach as step 2)

## 4. Update service-layer references to ErrNotFound and ErrInvalidInput

- [ ] 6.1 In `services/object_service.go`: replace all `repository.ErrNotFound` and `repository.ErrInvalidInput` references
- [ ] 7.1 In `services/object_type_service.go`: same migration
- [ ] 8.1 In `services/relationship_service.go`: same migration
- [ ] 9.1 In `services/relationship_type_service.go`: same migration

## 5. Update handler dispatcher to use common/errors sentinels

- [ ] 10.1 In `handlers/error.go`: change sentinel references from `repository.ErrNotFound` and `repository.ErrInvalidInput` to `errors.ErrNotFound` and `errors.ErrInvalidInput`
- [ ] 10.2 Add `"github.com/v-egorov/service-boilerplate/common/errors"` import if not present

## 6. Update test data to use common/errors sentinels

- [ ] 11.1 In `handlers/error_test.go`: update all test entries referencing ErrNotFound and ErrInvalidInput to use common/errors types
- [ ] 11.2 Verify no remaining references to repository.ErrNotFound or repository.ErrInvalidInput outside the interfaces.go file (where type aliases may remain)

## 7. Build, test, and verify

- [ ] 12.1 Run `make build-objects-service` — must compile without errors
- [ ] 12.2 Run `make test-objects-service` — all tests pass
- [ ] 12.3 Verify no references to ErrAlreadyExists remain anywhere in the codebase (definition, dispatcher case, or test data)
