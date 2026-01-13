# Phase 1: Database Migrations

**Estimated Time**: 2 hours
**Status**: ⬜ Not Started
**Dependencies**: None

## Overview

Replace the existing migration files with the new schema that supports hierarchical object types and objects, flexible attributes, and comprehensive audit fields.

## Tasks

### 1.1 Create New Migration Up Script

**File**: `migrations/000001_initial.up.sql`

**Steps**:
1. Delete existing content
2. Write new schema with:
   - `object_types` table with self-referencing parent_type_id
   - `objects` table with dual ID system (BIGINT + UUID)
   - Foreign key constraints
   - Indexes for performance
   - Triggers for cycle detection and audit fields

**Key Schema Details**:

```sql
-- Object Types Table
CREATE TABLE object_types (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    parent_type_id BIGINT REFERENCES object_types(id) ON DELETE RESTRICT,
    concrete_table_name VARCHAR(255) UNIQUE,
    description TEXT,
    is_sealed BOOLEAN DEFAULT FALSE,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    -- Not sure re first constraint - concrecte types not always should have it's own table
    -- CONSTRAINT valid_concrete_table_name CHECK (
    --     (parent_type_id IS NULL AND concrete_table_name IS NULL) OR
    --     (parent_type_id IS NOT NULL AND concrete_table_name IS NOT NULL)
    -- ),
    CONSTRAINT no_self_parent CHECK (parent_type_id IS NULL OR parent_type_id != id)
);

-- Indexes
CREATE INDEX idx_object_types_parent ON object_types(parent_type_id);
CREATE INDEX idx_object_types_name ON object_types(name);
CREATE INDEX idx_object_types_sealed ON object_types(is_sealed) WHERE is_sealed = TRUE;

-- Objects Table
CREATE TABLE objects (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id UUID NOT NULL DEFAULT gen_random_uuid(), -- make sure it's v5 UUID
    object_type_id BIGINT NOT NULL REFERENCES object_types(id) ON DELETE RESTRICT,
    parent_object_id BIGINT REFERENCES objects(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Audit
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    version BIGINT DEFAULT 0,
    created_by VARCHAR(255) DEFAULT 'system',
    updated_by VARCHAR(255) DEFAULT 'system',
 
    -- Metadata and labels
    metadata JSONB DEFAULT '{}'::jsonb,
    tags TEXT[] DEFAULT '{}',

    status VARCHAR(50) DEFAULT 'active' CHECK (
      status IN ('active', 'inactive', 'archived', 'deleted', 'pending')
    ),
    CONSTRAINT no_self_parent CHECK (parent_object_id IS NULL OR parent_object_id != id),
     -- Unique constraint on external ID
    CONSTRAINT objects_public_id_uniq UNIQUE (public_id)
);

-- Indexes
CREATE INDEX idx_objects_type_id ON objects(object_type_id);
CREATE INDEX idx_objects_type_status ON objects(object_type_id, status);
CREATE INDEX idx_objects_status ON objects(status);
CREATE INDEX idx_objects_parent_id ON objects(parent_object_id);
CREATE INDEX idx_objects_public_id ON objects(public_id);
CREATE INDEX idx_objects_metadata_gin ON objects USING GIN (metadata);
CREATE INDEX idx_objects_tags_gin ON objects USING GIN (tags);
CREATE INDEX idx_objects_created_at ON objects(created_at DESC);
CREATE INDEX idx_objects_updated_at ON objects(updated_at DESC);
CREATE INDEX idx_objects_deleted_at ON objects(deleted_at) WHERE deleted_at IS NOT NULL;

-- Triggers for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_object_types_updated_at
    BEFORE UPDATE ON object_types
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_objects_updated_at
    BEFORE UPDATE ON objects
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

**Verification**:
```bash
psql -U postgres -d objects_service -f migrations/000001_initial.up.sql
```

---

### 1.2 Create New Migration Down Script

**File**: `migrations/000001_initial.down.sql`

**Steps**:
1. Write rollback script that:
   - Drop triggers
   - Drop indexes
   - Drop objects table
   - Drop object_types table

```sql
-- Drop triggers
DROP TRIGGER IF EXISTS update_object_types_updated_at ON object_types;
DROP TRIGGER IF EXISTS update_objects_updated_at ON objects;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_object_types_parent;
DROP INDEX IF EXISTS idx_object_types_name;

DROP INDEX IF EXISTS idx_objects_type_id;
DROP INDEX IF EXISTS idx_objects_type_status;
DROP INDEX IF EXISTS idx_objects_parent_id;
DROP INDEX IF EXISTS idx_objects_public_id;
DROP INDEX IF EXISTS idx_objects_status;
DROP INDEX IF EXISTS idx_objects_metadata_gin;
DROP INDEX IF EXISTS idx_objects_tags_gin;
DROP INDEX IF EXISTS idx_objects_created_at;
DROP INDEX IF EXISTS idx_objects_updated_at;
DROP INDEX IF EXISTS idx_objects_deleted_at;

-- Drop tables
DROP TABLE IF EXISTS objects;
DROP TABLE IF EXISTS object_types;
```

**Verification**:
```bash
psql -U postgres -d objects_service -f migrations/000001_initial.down.sql
```

---

### 1.3 Update Dependencies File

**File**: `migrations/dependencies.json`

**Steps**:
1. Clear existing dependencies
2. This is initial migration, so no dependencies needed

```json
{
  "000001_initial": []
}
```

---

### 1.4 Update Environments File

**File**: `migrations/environments.json`

**Steps**:
1. Update to include development environment
2. Add reference to test data migration (will be created in Phase 7)

```json
{
  "development": [
    "000001_initial",
    "000002_dev_tax_test_data"
  ],
  "production": [
    "000001_initial"
  ],
  "staging": [
    "000001_initial"
  ]
}
```

---

## Checklist

- [ ] Replace `migrations/000001_initial.up.sql` with new schema
- [ ] Replace `migrations/000001_initial.down.sql` with rollback script
- [ ] Update `migrations/dependencies.json`
- [ ] Update `migrations/environments.json`
- [ ] Test migration up: `migrate up`
- [ ] Test migration down: `migrate down`
- [ ] Verify table structure: `\d object_types`, `\d objects`
- [ ] Verify indexes: `\di`
- [ ] Verify triggers exist in database
- [ ] Update progress.md

## Testing

```bash
# Test migration up
cd services/objects-service
go run cmd/migrate/main.go up

# Verify tables
psql postgresql://postgres:password@localhost:5432/objects_service -c "\dt"

# Test migration down
go run cmd/migrate/main.go down

# Test migration up again
go run cmd/migrate/main.go up
```

## Common Issues

**Issue**: Migration fails due to existing schema
**Solution**: Drop existing tables manually or ensure clean database

**Issue**: UUID generation not working
**Solution**: Ensure PostgreSQL has pgcrypto extension enabled: `CREATE EXTENSION IF NOT EXISTS pgcrypto;`

**Issue**: JSONB operators not recognized
**Solution**: Ensure PostgreSQL version is 12+ for full JSONB support

## Next Phase

Proceed to [Phase 2: Models Layer](phase-02-models.md) once all tasks in this phase are complete.
