# Migration: 000001_initial

## Overview
**Type:** schema
**Service:** auth-service
**Schema:** auth_service
**Created:** Auto-generated
**Risk Level:** Low

## Description
Initial migration that creates the auth_service schema and the entities table with basic entity management functionality.

## Changes Made

### Database Changes
- Create `auth_service` schema
- Create `auth_service.entities` table with the following columns:
  - `id` (BIGSERIAL PRIMARY KEY)
  - `name` (VARCHAR(100) NOT NULL UNIQUE)
  - `description` (TEXT)
  - `created_at` (TIMESTAMP WITH TIME ZONE, defaults to CURRENT_TIMESTAMP)
  - `updated_at` (TIMESTAMP WITH TIME ZONE, defaults to CURRENT_TIMESTAMP)

### Indexes Created
- `idx_entities_name` on `auth_service.entities(name)`
- `idx_entities_created_at` on `auth_service.entities(created_at)`

### Triggers Created
- `update_entities_updated_at` trigger to automatically update `updated_at` column

## Affected Tables
- `auth_service.entities` (created)

## Rollback Plan
The down migration will:
1. Drop the trigger
2. Drop the function
3. Drop the indexes
4. Drop the entities table
5. Drop the auth_service schema

## Testing
- [x] Migration applies successfully
- [x] Table structure is correct
- [x] Indexes are created
- [x] Triggers work correctly
- [x] Rollback works properly
- [x] Application can connect and use the table

## Notes
This is the foundational migration for the auth-service service. All subsequent migrations depend on this one.

## Risk Assessment
- **Risk Level:** Low
- **Estimated Duration:** 30 seconds
- **Rollback Safety:** Safe (no data loss on rollback)
- **Dependencies:** None