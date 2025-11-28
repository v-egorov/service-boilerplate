# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.1] - 2025-11-28

### Changed

- Removed unused `MIGRATION_PATH` variable from .env files and documentation.
- Updated Makefile to use dynamic variable references for Docker networks and monitoring volumes.

### Fixed

- Hard-coded network references in migration commands replaced with dynamic `$(NETWORK_NAME)`.
- Migration orchestrator initialization command fixed (was incorrectly calling `up` instead of `init`).
- Removed redundant variable definitions from Makefile to eliminate duplication with .env files.
- Fixed monitoring volume cleanup to use dynamic volume names.
- Updated service naming convention enforcement in `create-service.sh`.

## [1.1.0] - 2025-11-26

### Added

- Dynamic service detection in Makefile for automatic integration of new services.
- Centralized variable loading from `.env` files to reduce duplication.
- Service naming conventions documentation for consistent development.

### Changed

- Refactored Makefile to load environment variables dynamically using `eval` and `sed`.
- Restored migration dependency resolution with improved error handling.
- Updated migration system to exclude `SERVICE_NAME` from env loading to prevent overrides.

### Fixed

- Migration execution for multiple services with proper dependency ordering.
- Service name passing issues in orchestrator commands.
- Automatic detection of services created via `create-service.sh`.

### Technical Details

- `SERVICES` now auto-detects services ending with `-service`.
- `SERVICES_WITH_MIGRATIONS` requires `dependencies.json` for migration-enabled services.
- Variable loading macro filters comments, empties, and excludes `SERVICE_NAME`.

## [1.0.0] - 2025-09-01

Initial release with microservice boilerplate, API gateway, PostgreSQL migrations, and distributed tracing.

