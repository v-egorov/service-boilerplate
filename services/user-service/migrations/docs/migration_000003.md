# Migration: 000003_dev_test_data

## Overview
**Type:** data
**Service:** user-service
**Schema:** user_service
**Environment:** development
**Created:** Auto-generated
**Risk Level:** Low

## Description
Adds development test data including sample users and development-specific indexes. This migration only runs in the development environment and provides test data for development and testing purposes.

## Changes Made

### Database Changes
- Inserts development test users:
  - `dev.admin@example.com` (Dev Admin)
  - `test.user@example.com` (Test User)
  - `qa.tester@example.com` (QA Tester)

### Indexes Created
- `idx_users_dev_email` on `user_service.users(email)` WHERE email LIKE 'dev.%'
- `idx_users_test_email` on `user_service.users(email)` WHERE email LIKE 'test.%'

## Affected Tables
- `user_service.users` (test data inserted)

## Rollback Plan
Removes the test users and drops the development-specific indexes.

## Testing
- [x] Migration applies successfully in development
- [x] Test users are created
- [x] Indexes are created
- [x] Rollback removes test data cleanly
- [x] Does not run in production environment

## Notes
This migration is environment-specific and only executes in development. It provides consistent test data for development and automated testing. The ON CONFLICT DO NOTHING clause ensures it can be run multiple times safely.

## Risk Assessment
- **Risk Level:** Low
- **Estimated Duration:** 15 seconds
- **Rollback Safety:** Safe
- **Dependencies:** 000001_initial
- **Environment:** Development only