# Migration: 000004_add_user_settings

## Overview
**Type:** table
**Service:** user-service
**Schema:** user_service
**Created:** Вс 07 сен 2025 23:54:28 MSK
**Risk Level:** Low

## Description
Adds user settings table to store flexible key-value application settings for users. This allows storing various user preferences and configuration options in a generic, extensible way.

## Changes Made

### Database Changes
- Creates `user_service.user_settings` table with the following columns:
  - `id` (SERIAL PRIMARY KEY)
  - `user_id` (INTEGER, foreign key to users.id with CASCADE delete)
  - `setting_key` (VARCHAR(100) NOT NULL, setting identifier)
  - `setting_value` (TEXT, setting value - can store JSON or plain text)
  - `created_at` (TIMESTAMP WITH TIME ZONE)
  - `updated_at` (TIMESTAMP WITH TIME ZONE)

### Indexes Created
- `idx_user_settings_user_id` on `user_service.user_settings(user_id)`
- `idx_user_settings_key` on `user_service.user_settings(setting_key)`
- `idx_user_settings_created_at` on `user_service.user_settings(created_at)`

### Constraints Added
- UNIQUE constraint on `(user_id, setting_key)` (one setting per user per key)
- Foreign key constraint to `user_service.users(id)` with CASCADE delete

## Affected Tables
- `user_service.user_settings` (created)
- `user_service.users` (referenced via foreign key)

## Rollback Plan
Drops the user_settings table with CASCADE to remove all indexes and constraints.

## Testing
- [x] Migration applies successfully
- [x] Table structure is correct
- [x] Foreign key constraints work
- [x] Unique constraints prevent duplicates
- [x] Indexes improve query performance
- [x] Rollback removes table cleanly

## Notes
This flexible settings table allows storing various user preferences like UI settings, notification preferences, or custom configuration options. The setting_value column can store JSON for complex settings or simple text values.

## Risk Assessment
- **Risk Level:** Low
- **Estimated Duration:** 30 seconds
- **Rollback Safety:** Safe (no data loss on rollback)
- **Dependencies:** 000002_add_user_profiles
