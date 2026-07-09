## 1. Fix object_repository — Add ErrNoRows checks

- [x] 1.1 In `object_repository.GetByID()`: add `if errors.Is(err, sql.ErrNoRows) { return nil, ErrNotFound }` before the `%w` wrapping (after line ~209)
- [x] 1.2 In `object_repository.GetByName()`: add same check before the `%w` wrapping (after line ~265)

## 2. Fix object_type_repository — Add ErrNoRows check

- [x] 2.1 In `object_type_repository.GetByID()`: add `if errors.Is(err, sql.ErrNoRows) { return nil, ErrNotFound }` before the `%w` wrapping (after line ~180)

## 3. Remove workaround from handler dispatcher

- [x] 3.1 In `handlers/error.go`: remove the `case errors.Is(err, sql.ErrNoRows)` fallback case
- [x] 3.2 Verified: `database/sql` import removed from error.go — only used by sql.ErrNoRows fallback

## 4. Update handler dispatcher tests

- [x] 4.1 Existing `TestHandleError_ErrNotFound_404` already covers this: tests both unwrapped ErrNotFound and wrapped ErrNotFound via `%w`, verifying correct 404 mapping
- [x] 4.2 No sql.ErrNoRows fallback test exists — no cleanup needed

## 5. Build and verify

- [x] 5.1 Build passes: `make build-objects-service`
- [x] 5.2 All 9 packages pass (handlers 0.008s, repository 0.003s)
- [x] 5.3 No panics detected — test output clean
