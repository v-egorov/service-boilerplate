## 1. Fix object_repository — Add ErrNoRows checks

- [ ] 1.1 In `object_repository.GetByID()`: add `if errors.Is(err, sql.ErrNoRows) { return nil, ErrNotFound }` before the `%w` wrapping (after line ~209)
- [ ] 1.2 In `object_repository.GetByName()`: add same check before the `%w` wrapping (after line ~265)

## 2. Fix object_type_repository — Add ErrNoRows check

- [ ] 2.1 In `object_type_repository.GetByID()`: add `if errors.Is(err, sql.ErrNoRows) { return nil, ErrNotFound }` before the `%w` wrapping (after line ~180)

## 3. Remove workaround from handler dispatcher

- [ ] 3.1 In `handlers/error.go`: remove the `case errors.Is(err, sql.ErrNoRows)` fallback case
- [ ] 3.2 Verify that `database/sql` import is no longer needed (if it was only used for this case) — if other code uses it, keep the import

## 4. Update handler dispatcher tests

- [ ] 4.1 In `handlers/error_test.go`: add test verifying `repository.ErrNotFound` from repository layer maps to HTTP 404 without needing sql.ErrNoRows fallback
- [ ] 4.2 Remove any test that explicitly tests the sql.ErrNoRows fallback path (it's now dead code)

## 5. Build and verify

- [ ] 5.1 Run `make build-objects-service` to confirm compilation
- [ ] 5.2 Run `make test-objects-service` to confirm all tests pass
- [ ] 5.3 Verify no `[Recovery] panic recovered` entries in handler dispatcher tests (no unexpected panics)
