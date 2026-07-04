## ADDED Requirements

### Requirement: hasPermission treats empty scope as unrestricted
The auth-service permission resolution function (`hasPermission`) SHALL treat a database permission entry with an empty scope (no `:own` or `:all` suffix) as unrestricted — it matches any required scope level (`:all`, `:own`, or no scope). This ensures that flat/unscoped permissions functionally satisfy scoped permission requirements.

#### Scenario: Empty-scope DB entry satisfies :all requirement
- **WHEN** a user has the database permission `"object-types:read"` (scope = empty) and permiddleware requests `"object-types:read:all"`
- **THEN** `hasPermission()` returns `true` because an unscoped permission grants unrestricted access

#### Scenario: Empty-scope DB entry satisfies :own requirement
- **WHEN** a user has the database permission `"object-types:read"` (scope = empty) and permiddleware requests `"object-types:read:own"`
- **THEN** `hasPermission()` returns `true` because an unscoped permission grants unrestricted access, which covers ownership checks

#### Scenario: Empty-scope DB entry matches exact empty scope requirement
- **WHEN** a user has the database permission `"objects:create"` (scope = empty) and permiddleware requests `"objects:create"` (scope = empty)
- **THEN** `hasPermission()` returns `true` via the existing exact-match rule

#### Scenario: :all DB entry still satisfies :own requirement (existing behavior preserved)
- **WHEN** a user has `"object-types:read:all"` and permiddleware requests `"object-types:read:own"`
- **THEN** `hasPermission()` returns `true` via the existing broad-wins-over-narrow rule

#### Scenario: :own DB entry does NOT satisfy :all requirement (existing behavior preserved)
- **WHEN** a user has `"object-types:read:own"` and permiddleware requests `"object-types:read:all"`
- **THEN** `hasPermission()` returns `false` — narrow permission cannot grant broad access

#### Scenario: Different resource does not match regardless of scope
- **WHEN** a user has `"relationships:read:all"` and permiddleware requests `"object-types:read:all"`
- **THEN** `hasPermission()` returns `false` — resource mismatch is checked before scope comparison
