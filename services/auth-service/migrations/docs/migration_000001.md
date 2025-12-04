# Migration: 000001_initial

## Overview
**Type:** schema
**Service:** auth-service
**Schema:** auth_service
**Created:** Auto-generated
**Risk Level:** Low

## Description
Initial migration that creates the auth_service schema with authentication and authorization tables.

## Changes Made

### Database Changes
- Create `auth_service` schema
- Create authentication and authorization tables:
  - `auth_tokens` - JWT token storage
  - `user_sessions` - User session management
  - `roles` - RBAC roles
  - `permissions` - RBAC permissions
  - `role_permissions` - Role-permission relationships
  - `user_roles` - User-role assignments

### Indexes Created
- Various indexes for performance optimization on auth tokens, sessions, roles, and permissions

### Triggers Created
- `update_auth_tokens_updated_at` trigger for auth_tokens table

## Affected Tables
- `auth_service.auth_tokens` (created)
- `auth_service.user_sessions` (created)
- `auth_service.roles` (created)
- `auth_service.permissions` (created)
- `auth_service.role_permissions` (created)
- `auth_service.user_roles` (created)

## Rollback Plan
The down migration will drop all created tables and the schema.

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