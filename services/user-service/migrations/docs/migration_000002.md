# Migration: 000002_add_user_profiles

## Overview
**Type:** table
**Service:** user-service
**Schema:** user_service
**Created:** Auto-generated
**Risk Level:** Medium

## Description
Adds user profiles table to store extended user information including bio, avatar, website, location, timezone preferences, theme settings, and notification preferences.

## Changes Made

### Database Changes
- Creates `user_service.user_profiles` table with the following columns:
  - `id` (SERIAL PRIMARY KEY)
  - `user_id` (INTEGER, foreign key to users.id with CASCADE delete)
  - `bio` (TEXT, user biography)
  - `avatar_url` (VARCHAR(500), profile picture URL)
  - `website` (VARCHAR(255), personal website)
  - `location` (VARCHAR(100), user location)
  - `timezone` (VARCHAR(50), default 'UTC')
  - `theme` (VARCHAR(20), default 'light', CHECK constraint for 'light'/'dark'/'auto')
  - `email_notifications` (BOOLEAN, default true)
  - `created_at` (TIMESTAMP WITH TIME ZONE)
  - `updated_at` (TIMESTAMP WITH TIME ZONE)

### Indexes Created
- `idx_user_profiles_user_id` on `user_service.user_profiles(user_id)`
- `idx_user_profiles_theme` on `user_service.user_profiles(theme)`
- `idx_user_profiles_created_at` on `user_service.user_profiles(created_at)`

### Constraints Added
- UNIQUE constraint on `user_id` (one profile per user)
- Foreign key constraint to `user_service.users(id)` with CASCADE delete
- CHECK constraint on `theme` column

## Affected Tables
- `user_service.user_profiles` (created)
- `user_service.users` (referenced via foreign key)

## Rollback Plan
Drops the user_profiles table with CASCADE to remove all indexes and constraints.

## Testing
- [x] Migration applies successfully
- [x] Table structure is correct
- [x] Foreign key constraints work
- [x] Indexes improve query performance
- [x] Rollback removes table cleanly
- [x] Application integration works

## Notes
This migration establishes the foundation for user profile management. The table supports common profile features like themes, notifications, and contact information.

## Risk Assessment
- **Risk Level:** Medium
- **Estimated Duration:** 45 seconds
- **Rollback Safety:** Safe (no data loss on rollback)
- **Dependencies:** 000001_initial