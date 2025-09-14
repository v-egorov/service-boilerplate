# Migration: 000001_initial

## Overview
**Type:** schema
**Service:** SERVICE_NAME
**Schema:** SERVICE_NAME
**Created:** Auto-generated
**Risk Level:** Low

## Description
Initial migration that creates the SERVICE_NAME schema and the entities table with basic entity management functionality.

## Changes Made

### Database Changes
- Create `SERVICE_NAME` schema
- Create `SERVICE_NAME.entities` table with the following columns:
  - `id` (SERIAL PRIMARY KEY)
  - `name` (VARCHAR(255) NOT NULL)
  - `description` (TEXT)
  - `created_at` (TIMESTAMP WITH TIME ZONE, defaults to CURRENT_TIMESTAMP)
  - `updated_at` (TIMESTAMP WITH TIME ZONE, defaults to CURRENT_TIMESTAMP)

### Indexes Created
- `idx_entities_name` on `SERVICE_NAME.entities(name)`
- `idx_entities_created_at` on `SERVICE_NAME.entities(created_at DESC)`

## Affected Tables
- `SERVICE_NAME.entities` (created)

## Rollback Plan
The down migration will:
1. Drop the indexes
2. Drop the entities table
3. Drop the SERVICE_NAME schema

## Testing
- [x] Migration applies successfully
- [x] Table structure is correct
- [x] Indexes are created
- [x] Rollback works properly
- [x] Application can connect and use the table

## Notes
This is the foundational migration for the SERVICE_NAME service. All subsequent migrations depend on this one.

## Risk Assessment
- **Risk Level:** Low
- **Estimated Duration:** 30 seconds
- **Rollback Safety:** Safe (no data loss on rollback)
- **Dependencies:** None