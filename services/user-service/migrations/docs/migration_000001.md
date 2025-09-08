# Migration: 000001_initial

## Overview
**Type:** schema
**Service:** user-service
**Schema:** user_service
**Created:** Auto-generated
**Risk Level:** Low

## Description
Initial migration that creates the user_service schema and the users table with basic user management functionality.

## Changes Made

### Database Changes
- Create `user_service` schema
- Create `user_service.users` table with the following columns:
  - `id` (SERIAL PRIMARY KEY)
  - `email` (VARCHAR(255) UNIQUE NOT NULL)
  - `first_name` (VARCHAR(100) NOT NULL)
  - `last_name` (VARCHAR(100) NOT NULL)
  - `created_at` (TIMESTAMP WITH TIME ZONE, defaults to CURRENT_TIMESTAMP)
  - `updated_at` (TIMESTAMP WITH TIME ZONE, defaults to CURRENT_TIMESTAMP)

### Indexes Created
- `idx_users_email` on `user_service.users(email)`
- `idx_users_created_at` on `user_service.users(created_at DESC)`

## Affected Tables
- `user_service.users` (created)

## Rollback Plan
The down migration will:
1. Drop the indexes
2. Drop the users table
3. Drop the user_service schema

## Testing
- [x] Migration applies successfully
- [x] Table structure is correct
- [x] Indexes are created
- [x] Rollback works properly
- [x] Application can connect and use the table

## Notes
This is the foundational migration for the user service. All subsequent migrations depend on this one.

## Risk Assessment
- **Risk Level:** Low
- **Estimated Duration:** 30 seconds
- **Rollback Safety:** Safe (no data loss on rollback)
- **Dependencies:** None