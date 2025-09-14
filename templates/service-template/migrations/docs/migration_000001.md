# Migration: 000001_initial

## Overview
**Type:** schema
**Service:** SERVICE_NAME
**Schema:** SCHEMA_NAME
**Created:** Auto-generated
**Risk Level:** Low

## Description
Initial migration that creates the SCHEMA_NAME schema and the entities table with basic entity management functionality.

## Changes Made

### Database Changes
- Create `SCHEMA_NAME` schema
- Create `SCHEMA_NAME.entities` table with the following columns:
  - `id` (BIGSERIAL PRIMARY KEY)
  - `name` (VARCHAR(100) NOT NULL UNIQUE)
  - `description` (TEXT)
  - `created_at` (TIMESTAMP WITH TIME ZONE, defaults to CURRENT_TIMESTAMP)
  - `updated_at` (TIMESTAMP WITH TIME ZONE, defaults to CURRENT_TIMESTAMP)

### Indexes Created
- `idx_entities_name` on `SCHEMA_NAME.entities(name)`
- `idx_entities_created_at` on `SCHEMA_NAME.entities(created_at)`

### Triggers Created
- `update_entities_updated_at` trigger to automatically update `updated_at` column

## Affected Tables
- `SCHEMA_NAME.entities` (created)

## Rollback Plan
The down migration will:
1. Drop the trigger
2. Drop the function
3. Drop the indexes
4. Drop the entities table
5. Drop the SCHEMA_NAME schema

## Testing
- [x] Migration applies successfully
- [x] Table structure is correct
- [x] Indexes are created
- [x] Triggers work correctly
- [x] Rollback works properly
- [x] Application can connect and use the table

## Notes
This is the foundational migration for the SERVICE_NAME service. All subsequent migrations depend on this one.

## Risk Assessment
- **Risk Level:** Low
- **Estimated Duration:** 30 seconds
- **Rollback Safety:** Safe (no data loss on rollback)
- **Dependencies:** None