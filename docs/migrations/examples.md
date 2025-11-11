# Migration Examples

## Real-World Migration Scenarios

### Example 1: Adding User Profiles Feature

**Scenario:** Add user profile functionality with bio, avatar, and preferences.

#### Step 1: Generate Migration
```bash
make db-migrate-generate NAME=add_user_profiles TYPE=table
```

#### Step 2: Edit Migration Files

**`000005_add_user_profiles.up.sql`:**
```sql
-- Migration: 000005_add_user_profiles
-- Description: Add user profiles table with bio and preferences
-- Created: Auto-generated

-- Create user profiles table
CREATE TABLE IF NOT EXISTS user_service.user_profiles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES user_service.users(id) ON DELETE CASCADE,
    bio TEXT,
    avatar_url VARCHAR(500),
    website VARCHAR(255),
    location VARCHAR(100),
    timezone VARCHAR(50) DEFAULT 'UTC',
    theme VARCHAR(20) DEFAULT 'light' CHECK (theme IN ('light', 'dark', 'auto')),
    email_notifications BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_profiles_user_id ON user_service.user_profiles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_profiles_theme ON user_service.user_profiles(theme);
CREATE INDEX IF NOT EXISTS idx_user_profiles_created_at ON user_service.user_profiles(created_at);

-- Add table comment
COMMENT ON TABLE user_service.user_profiles IS 'Extended user profile information and preferences';
COMMENT ON COLUMN user_service.user_profiles.bio IS 'User biography text (max 1000 characters)';
COMMENT ON COLUMN user_service.user_profiles.theme IS 'UI theme preference';
```

**`000005_add_user_profiles.down.sql`:**
```sql
-- Migration Rollback: 000005_add_user_profiles
-- Description: Remove user profiles table
-- Created: Auto-generated

-- Drop table (CASCADE removes indexes and constraints)
DROP TABLE IF EXISTS user_service.user_profiles CASCADE;
```

#### Step 3: Update Application Code

**`services/user-service/internal/models/user.go`:**
```go
// Add profile fields to User model
type User struct {
    ID                int       `json:"id" db:"id"`
    Email             string    `json:"email" db:"email"`
    FirstName         string    `json:"first_name" db:"first_name"`
    LastName          string    `json:"last_name" db:"last_name"`
    CreatedAt         time.Time `json:"created_at" db:"created_at"`
    UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`

    // New profile fields
    Profile           *UserProfile `json:"profile,omitempty" db:"-"`
}

// New UserProfile model
type UserProfile struct {
    ID                 int    `json:"id" db:"id"`
    UserID             int    `json:"user_id" db:"user_id"`
    Bio                string `json:"bio" db:"bio"`
    AvatarURL          string `json:"avatar_url" db:"avatar_url"`
    Website            string `json:"website" db:"website"`
    Location           string `json:"location" db:"location"`
    Timezone           string `json:"timezone" db:"timezone"`
    Theme              string `json:"theme" db:"theme"`
    EmailNotifications bool   `json:"email_notifications" db:"email_notifications"`
    CreatedAt          time.Time `json:"created_at" db:"created_at"`
    UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}
```

**`services/user-service/internal/repository/user_repository.go`:**
```go
// Add profile-related methods
func (r *UserRepository) GetUserWithProfile(ctx context.Context, id int) (*models.User, error) {
    query := `
        SELECT
            u.id, u.email, u.first_name, u.last_name, u.created_at, u.updated_at,
            p.id as profile_id, p.bio, p.avatar_url, p.website, p.location,
            p.timezone, p.theme, p.email_notifications, p.created_at as profile_created_at
        FROM user_service.users u
        LEFT JOIN user_service.user_profiles p ON u.id = p.user_id
        WHERE u.id = $1`

    user := &models.User{}
    var profile models.UserProfile

    err := r.db.QueryRow(ctx, query, id).Scan(
        &user.ID, &user.Email, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt,
        &profile.ID, &profile.Bio, &profile.AvatarURL, &profile.Website, &profile.Location,
        &profile.Timezone, &profile.Theme, &profile.EmailNotifications, &profile.CreatedAt,
    )

    if err != nil {
        return nil, err
    }

    user.Profile = &profile
    return user, nil
}

func (r *UserRepository) CreateProfile(ctx context.Context, profile *models.UserProfile) error {
    query := `
        INSERT INTO user_service.user_profiles
        (user_id, bio, avatar_url, website, location, timezone, theme, email_notifications)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, created_at`

    return r.db.QueryRow(ctx, query,
        profile.UserID, profile.Bio, profile.AvatarURL, profile.Website,
        profile.Location, profile.Timezone, profile.Theme, profile.EmailNotifications,
    ).Scan(&profile.ID, &profile.CreatedAt)
}
```

#### Step 4: Test the Migration

```bash
# Validate migration
make db-validate

# Apply migration
make db-migrate-up

# Check table structure
make db-tables

# Verify indexes
make db-schema

# Test rollback
make db-migrate-down
make db-migrate-up
```

---

### Example 2: Performance Optimization Migration

**Scenario:** Add indexes and optimize queries for user search functionality.

#### Step 1: Generate Migration
```bash
make db-migrate-generate NAME=optimize_user_search TYPE=index
```

#### Step 2: Edit Migration Files

**`000006_optimize_user_search.up.sql`:**
```sql
-- Migration: 000006_optimize_user_search
-- Description: Add indexes for user search and filtering performance
-- Created: Auto-generated

-- Composite index for name searches
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_name_search
ON user_service.users (last_name, first_name);

-- Partial index for active users only
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_active_recent
ON user_service.users (created_at DESC)
WHERE created_at > CURRENT_TIMESTAMP - INTERVAL '90 days';

-- Index for email domain analysis
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email_domain
ON user_service.users ((split_part(email, '@', 2)));

-- Index for user profiles search
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_profiles_location
ON user_service.user_profiles (location)
WHERE location IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_profiles_website
ON user_service.user_profiles (website)
WHERE website IS NOT NULL;

-- Add comments
COMMENT ON INDEX user_service.idx_users_name_search IS 'Optimizes user name searches and sorting';
COMMENT ON INDEX user_service.idx_users_active_recent IS 'Optimizes queries for recently active users';
COMMENT ON INDEX user_service.idx_users_email_domain IS 'Supports email domain analytics';
```

**`000006_optimize_user_search.down.sql`:**
```sql
-- Migration Rollback: 000006_optimize_user_search
-- Description: Remove performance optimization indexes
-- Created: Auto-generated

-- Drop indexes (CONCURRENTLY not needed for DROP)
DROP INDEX IF EXISTS user_service.idx_users_name_search;
DROP INDEX IF EXISTS user_service.idx_users_active_recent;
DROP INDEX IF EXISTS user_service.idx_users_email_domain;
DROP INDEX IF EXISTS user_service.idx_user_profiles_location;
DROP INDEX IF EXISTS user_service.idx_user_profiles_website;
```

#### Step 3: Performance Testing

```sql
-- Test query performance before and after migration
EXPLAIN ANALYZE
SELECT * FROM user_service.users
WHERE last_name LIKE 'Smith%'
ORDER BY first_name;

EXPLAIN ANALYZE
SELECT u.*, p.location
FROM user_service.users u
LEFT JOIN user_service.user_profiles p ON u.id = p.user_id
WHERE u.created_at > CURRENT_TIMESTAMP - INTERVAL '30 days'
ORDER BY u.created_at DESC;
```

---

### Example 3: Data Migration with Business Logic

**Scenario:** Migrate existing user data to add default profile entries.

#### Step 1: Generate Migration
```bash
make db-migrate-generate NAME=migrate_user_profiles_data TYPE=data
```

#### Step 2: Edit Migration Files

**`000007_migrate_user_profiles_data.up.sql`:**
```sql
-- Migration: 000007_migrate_user_profiles_data
-- Description: Create default profile entries for existing users
-- Created: Auto-generated

-- Create default profiles for users who don't have one
INSERT INTO user_service.user_profiles (
    user_id,
    timezone,
    theme,
    email_notifications,
    created_at,
    updated_at
)
SELECT
    u.id,
    'UTC' as timezone,
    'light' as theme,
    true as email_notifications,
    CURRENT_TIMESTAMP as created_at,
    CURRENT_TIMESTAMP as updated_at
FROM user_service.users u
LEFT JOIN user_service.user_profiles p ON u.id = p.user_id
WHERE p.id IS NULL;

-- Update existing profiles with default values
UPDATE user_service.user_profiles
SET
    timezone = COALESCE(timezone, 'UTC'),
    theme = COALESCE(theme, 'light'),
    email_notifications = COALESCE(email_notifications, true),
    updated_at = CURRENT_TIMESTAMP
WHERE timezone IS NULL
   OR theme IS NULL
   OR email_notifications IS NULL;

-- Log migration results
DO $$
DECLARE
    profiles_created INTEGER;
    profiles_updated INTEGER;
BEGIN
    SELECT COUNT(*) INTO profiles_created
    FROM user_service.user_profiles
    WHERE created_at >= CURRENT_TIMESTAMP - INTERVAL '1 minute';

    SELECT COUNT(*) INTO profiles_updated
    FROM user_service.user_profiles
    WHERE updated_at >= CURRENT_TIMESTAMP - INTERVAL '1 minute'
      AND created_at < CURRENT_TIMESTAMP - INTERVAL '1 minute';

    RAISE NOTICE 'Migration completed: % profiles created, % profiles updated',
                profiles_created, profiles_updated;
END $$;
```

**`000007_migrate_user_profiles_data.down.sql`:**
```sql
-- Migration Rollback: 000007_migrate_user_profiles_data
-- Description: Remove auto-generated profile data
-- WARNING: This will delete user profile data!
-- Created: Auto-generated

-- NOTE: This rollback is potentially destructive
-- In production, consider backing up profile data before rollback

-- Remove default profiles created by this migration
-- (Profiles created in the last hour that have default values)
DELETE FROM user_service.user_profiles
WHERE created_at >= CURRENT_TIMESTAMP - INTERVAL '1 hour'
  AND bio IS NULL
  AND avatar_url IS NULL
  AND website IS NULL
  AND location IS NULL
  AND timezone = 'UTC'
  AND theme = 'light'
  AND email_notifications = true;

-- Reset updated fields to NULL for partial rollbacks
UPDATE user_service.user_profiles
SET
    timezone = NULL,
    theme = NULL,
    email_notifications = NULL
WHERE updated_at >= CURRENT_TIMESTAMP - INTERVAL '1 hour'
  AND created_at < CURRENT_TIMESTAMP - INTERVAL '1 hour';
```

#### Step 3: Validation Queries

```sql
-- Verify migration results
SELECT
    COUNT(*) as total_users,
    COUNT(p.id) as users_with_profiles,
    COUNT(CASE WHEN p.timezone = 'UTC' THEN 1 END) as default_timezones,
    COUNT(CASE WHEN p.theme = 'light' THEN 1 END) as default_themes
FROM user_service.users u
LEFT JOIN user_service.user_profiles p ON u.id = p.user_id;
```

---

### Example 4: Environment-Specific Migration

**Scenario:** Add development test data that shouldn't exist in production.

#### Step 1: Create Environment-Specific Migration
```bash
# Create development directory
mkdir -p services/user-service/migrations/development

# Create migration file
cat > services/user-service/migrations/development/000008_dev_test_users.up.sql << 'EOF'
-- Development Migration: Add test users for development
-- This migration only runs in development environment
-- Created: Manual

-- Insert development test users
INSERT INTO user_service.users (email, first_name, last_name) VALUES
    ('dev.tester@example.com', 'Dev', 'Tester'),
    ('qa.engineer@example.com', 'QA', 'Engineer'),
    ('product.manager@example.com', 'Product', 'Manager'),
    ('ux.designer@example.com', 'UX', 'Designer')
ON CONFLICT (email) DO NOTHING;

-- Add corresponding profiles
INSERT INTO user_service.user_profiles (
    user_id,
    bio,
    location,
    timezone
)
SELECT
    u.id,
    CASE
        WHEN u.email = 'dev.tester@example.com' THEN 'Software Developer specializing in backend systems'
        WHEN u.email = 'qa.engineer@example.com' THEN 'Quality Assurance Engineer with 5+ years experience'
        WHEN u.email = 'product.manager@example.com' THEN 'Product Manager focused on user experience'
        WHEN u.email = 'ux.designer@example.com' THEN 'UX Designer passionate about user-centered design'
    END as bio,
    CASE
        WHEN u.email = 'dev.tester@example.com' THEN 'San Francisco, CA'
        WHEN u.email = 'qa.engineer@example.com' THEN 'New York, NY'
        WHEN u.email = 'product.manager@example.com' THEN 'Austin, TX'
        WHEN u.email = 'ux.designer@example.com' THEN 'Seattle, WA'
    END as location,
    'America/Los_Angeles' as timezone
FROM user_service.users u
WHERE u.email IN (
    'dev.tester@example.com',
    'qa.engineer@example.com',
    'product.manager@example.com',
    'ux.designer@example.com'
)
ON CONFLICT (user_id) DO NOTHING;
EOF
```

#### Step 2: Create Down Migration
```bash
cat > services/user-service/migrations/development/000008_dev_test_users.down.sql << 'EOF'
-- Development Migration Rollback: Remove test users
-- This rollback only runs in development environment
-- Created: Manual

-- Remove test user profiles first (due to foreign key)
DELETE FROM user_service.user_profiles
WHERE user_id IN (
    SELECT id FROM user_service.users
    WHERE email IN (
        'dev.tester@example.com',
        'qa.engineer@example.com',
        'product.manager@example.com',
        'ux.designer@example.com'
    )
);

-- Remove test users
DELETE FROM user_service.users
WHERE email IN (
    'dev.tester@example.com',
    'qa.engineer@example.com',
    'product.manager@example.com',
    'ux.designer@example.com'
);
EOF
```

#### Step 3: Update Environment Configuration
```json
// services/user-service/migrations/environments.json
{
  "environments": {
    "development": {
      "migrations": [
        "development/000008_dev_test_users.up.sql"
      ],
      "config": {
        "allow_destructive_operations": true,
        "skip_validation": false
      }
    }
  }
}
```

#### Step 4: Test Environment-Specific Execution
```bash
# Development environment
export MIGRATION_ENV=development
make db-migrate-up

# Check that test users exist
make db-seed-enhanced ENV=development

# Verify test data
docker-compose exec postgres psql -U postgres -d service_db \
  -c "SELECT email, first_name, last_name FROM user_service.users WHERE email LIKE 'dev.%' OR email LIKE 'qa.%';"
```

---

### Example 5: Complex Schema Refactoring

**Scenario:** Split user contact information into separate table.

#### Step 1: Generate Migration
```bash
make db-migrate-generate NAME=refactor_user_contacts TYPE=table
```

#### Step 2: Complex Migration with Data Migration

**`000009_refactor_user_contacts.up.sql`:**
```sql
-- Migration: 000009_refactor_user_contacts
-- Description: Extract contact information into separate table
-- Created: Auto-generated

-- Create contacts table
CREATE TABLE IF NOT EXISTS user_service.user_contacts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES user_service.users(id) ON DELETE CASCADE,
    contact_type VARCHAR(20) NOT NULL CHECK (contact_type IN ('email', 'phone', 'address')),
    contact_value TEXT NOT NULL,
    is_primary BOOLEAN DEFAULT false,
    verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id, contact_type, contact_value)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_user_contacts_user_id ON user_service.user_contacts(user_id);
CREATE INDEX IF NOT EXISTS idx_user_contacts_type ON user_service.user_contacts(contact_type);
CREATE INDEX IF NOT EXISTS idx_user_contacts_primary ON user_service.user_contacts(user_id, is_primary);

-- Migrate existing email data
INSERT INTO user_service.user_contacts (
    user_id,
    contact_type,
    contact_value,
    is_primary,
    verified_at
)
SELECT
    id,
    'email',
    email,
    true,
    created_at  -- Assume emails were verified at account creation
FROM user_service.users
WHERE email IS NOT NULL;

-- Migrate phone data from profiles (if exists)
INSERT INTO user_service.user_contacts (
    user_id,
    contact_type,
    contact_value,
    is_primary
)
SELECT
    user_id,
    'phone',
    phone,
    true
FROM user_service.user_profiles
WHERE phone IS NOT NULL AND phone != '';

-- Add check constraints
ALTER TABLE user_service.user_contacts
ADD CONSTRAINT chk_contact_value_not_empty
CHECK (length(trim(contact_value)) > 0);

-- Update users table to remove redundant columns later
-- (Keep for now to ensure rollback works)

-- Log migration results
DO $$
DECLARE
    emails_migrated INTEGER;
    phones_migrated INTEGER;
BEGIN
    SELECT COUNT(*) INTO emails_migrated
    FROM user_service.user_contacts
    WHERE contact_type = 'email';

    SELECT COUNT(*) INTO phones_migrated
    FROM user_service.user_contacts
    WHERE contact_type = 'phone';

    RAISE NOTICE 'Contact migration completed: % emails, % phones migrated',
                emails_migrated, phones_migrated;
END $$;
```

**`000009_refactor_user_contacts.down.sql`:**
```sql
-- Migration Rollback: 000009_refactor_user_contacts
-- Description: Rollback contact information refactoring
-- WARNING: This will lose contact verification data
-- Created: Auto-generated

-- NOTE: This rollback requires careful data reconstruction
-- Some data loss may occur for verified contacts

-- Restore email data to users table
UPDATE user_service.users
SET email = uc.contact_value
FROM user_service.user_contacts uc
WHERE uc.user_id = users.id
  AND uc.contact_type = 'email'
  AND uc.is_primary = true;

-- Restore phone data to profiles
UPDATE user_service.user_profiles
SET phone = uc.contact_value
FROM user_service.user_contacts uc
WHERE uc.user_id = user_profiles.user_id
  AND uc.contact_type = 'phone'
  AND uc.is_primary = true;

-- Drop contacts table
DROP TABLE IF EXISTS user_service.user_contacts CASCADE;
```

#### Step 3: Application Code Updates

```go
// New Contact model
type UserContact struct {
    ID          int       `json:"id" db:"id"`
    UserID      int       `json:"user_id" db:"user_id"`
    Type        string    `json:"type" db:"contact_type"`
    Value       string    `json:"value" db:"contact_value"`
    IsPrimary   bool      `json:"is_primary" db:"is_primary"`
    VerifiedAt  *time.Time `json:"verified_at" db:"verified_at"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Repository methods for contacts
func (r *UserRepository) GetUserContacts(ctx context.Context, userID int) ([]*models.UserContact, error) {
    query := `SELECT * FROM user_service.user_contacts WHERE user_id = $1 ORDER BY is_primary DESC, created_at ASC`
    // Implementation
}

func (r *UserRepository) AddUserContact(ctx context.Context, contact *models.UserContact) error {
    query := `
        INSERT INTO user_service.user_contacts
        (user_id, contact_type, contact_value, is_primary)
        VALUES ($1, $2, $3, $4)`
    // Implementation
}
```

---

## Testing Patterns

### Unit Testing Migrations

```go
func TestMigration000005(t *testing.T) {
    // Setup isolated test database
    db := setupTestDB(t)

    // Apply migration
    err := applyMigration(db, "000005_add_user_profiles.up.sql")
    assert.NoError(t, err)

    // Verify table structure
    exists, err := tableExists(db, "user_service.user_profiles")
    assert.NoError(t, err)
    assert.True(t, exists)

    // Verify constraints
    hasFK, err := hasForeignKey(db, "user_service.user_profiles", "user_id")
    assert.NoError(t, err)
    assert.True(t, hasFK)

    // Test data insertion
    userID := insertTestUser(t, db)
    err = insertUserProfile(db, userID, "Test bio", "America/New_York")
    assert.NoError(t, err)

    // Verify rollback
    err = applyMigration(db, "000005_add_user_profiles.down.sql")
    assert.NoError(t, err)

    exists, err = tableExists(db, "user_service.user_profiles")
    assert.NoError(t, err)
    assert.False(t, exists)
}
```

### Integration Testing

```bash
# Test migration in isolated environment
docker-compose -f docker-compose.test.yml up -d

# Run full migration suite
make db-migrate-up
make db-seed-enhanced ENV=development

# Run application tests
make test-all

# Test rollback scenarios

make db-migrate-up

# Clean up
docker-compose -f docker-compose.test.yml down -v
```

### Performance Testing

```sql
-- Benchmark migration performance
\timing on

-- Test migration execution time
SELECT clock_timestamp();

-- Run migration
\i services/user-service/migrations/000005_add_user_profiles.up.sql

SELECT clock_timestamp();

-- Test query performance after migration
EXPLAIN (ANALYZE, BUFFERS)
SELECT u.*, p.bio
FROM user_service.users u
LEFT JOIN user_service.user_profiles p ON u.id = p.user_id
WHERE u.created_at > CURRENT_TIMESTAMP - INTERVAL '30 days';
```

---

## Common Patterns and Templates

### Adding Audit Columns

```sql
-- Standard audit columns for new tables
CREATE TABLE user_service.table_name (
    id SERIAL PRIMARY KEY,
    -- Business columns here
    created_by INTEGER,
    updated_by INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add audit indexes
CREATE INDEX idx_table_name_created_at ON user_service.table_name(created_at);
CREATE INDEX idx_table_name_updated_at ON user_service.table_name(updated_at);
```

### Soft Delete Pattern

```sql
-- Add soft delete to existing table
ALTER TABLE user_service.table_name
ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN deleted_by INTEGER;

-- Add partial index for performance
CREATE INDEX idx_table_name_not_deleted
ON user_service.table_name (id)
WHERE deleted_at IS NULL;

-- Update queries to exclude deleted records
SELECT * FROM user_service.table_name
WHERE deleted_at IS NULL;
```

### Polymorphic Associations

```sql
-- Generic activity/log table
CREATE TABLE user_service.activities (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES user_service.users(id),
    activity_type VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id INTEGER NOT NULL,
    description TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_activities_user_id ON user_service.activities(user_id);
CREATE INDEX idx_activities_entity ON user_service.activities(entity_type, entity_id);
CREATE INDEX idx_activities_created_at ON user_service.activities(created_at DESC);
```

These examples demonstrate the full range of migration scenarios you might encounter, from simple table additions to complex data refactoring with business logic.