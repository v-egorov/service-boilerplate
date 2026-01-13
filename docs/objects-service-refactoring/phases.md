# Implementation Phases

This document provides an index and overview of all implementation phases. Each phase has its own detailed document linked below.

## Phase Overview

| Phase | Name | Estimated Time | Status |
|-------|------|----------------|--------|
| [Phase 1](phase-01-migrations.md) | Database Migrations | 2 hours | ⬜ Not Started |
| [Phase 2](phase-02-models.md) | Models Layer | 2.5 hours | ⬜ Not Started |
| [Phase 3](phase-03-repositories.md) | Repository Layer | 4 hours | ⬜ Not Started |
| [Phase 4](phase-04-services.md) | Service Layer | 4 hours | ⬜ Not Started |
| [Phase 5](phase-05-handlers.md) | Handlers Layer | 4 hours | ⬜ Not Started |
| [Phase 6](phase-06-main.md) | Main Application | 1 hour | ⬜ Not Started |
| [Phase 7](phase-07-test-data.md) | Development Test Data | 1 hour | ⬜ Not Started |
| [Phase 8](phase-08-tests.md) | Tests | 4 hours | ⬜ Not Started |
| [Phase 9](phase-09-documentation.md) | Documentation | 2 hours | ⬜ Not Started |
| [Phase 10](phase-10-class-table-inheritance.md) | Class Table Inheritance (Future) | 8-10 hours | ⬜ Not Started |

**Total Estimated Time**: ~32.5-34.5 hours (with Phase 10)

## Prerequisites

Before starting any phase, ensure:
1. All [design questions](design-questions.md) have been answered
2. The objects-service repository is available
3. Go 1.25+ is installed
4. PostgreSQL 15+ is accessible
5. Database migrations tool is configured

## Phase Dependencies

```
Phase 1 (Migrations)
    ↓
Phase 2 (Models)
    ↓
Phase 3 (Repositories) ← Phase 2
    ↓
Phase 4 (Services) ← Phase 3
    ↓
Phase 5 (Handlers) ← Phase 4
    ↓
Phase 6 (Main) ← Phase 5
    ↓
Phase 7 (Test Data) ← Phase 6
    ↓
Phase 8 (Tests) ← Phase 5
    ↓
Phase 9 (Documentation) ← Phase 8
```

## Quick Reference

### Files to Delete (All Phases Combined)
```
internal/models/entity.go
internal/models/entity_test.go
internal/repository/entity_repository.go
internal/repository/entity_repository_test.go
internal/services/entity_service.go
internal/services/entity_service_test.go
internal/handlers/entity_handler.go
internal/handlers/entity_handler_test.go
migrations/000001_initial.up.sql
migrations/000001_initial.down.sql
migrations/development/000003_dev_test_data.up.sql
```

### Files to Create (All Phases Combined)
```
internal/models/object_type.go
internal/models/object.go
internal/models/object_type_request.go
internal/models/object_request.go
internal/repository/object_type_repository.go
internal/repository/object_repository.go
internal/services/object_type_service.go
internal/services/object_service.go
internal/handlers/object_type_handler.go
internal/handlers/object_handler.go
migrations/development/000002_dev_tax_test_data.up.sql
```

### Files to Update (All Phases Combined)
```
migrations/000001_initial.up.sql
migrations/000001_initial.down.sql
migrations/dependencies.json
migrations/environments.json
cmd/main.go
services/objects-service/README.md
docs/objects-service-refactoring/progress.md
```

## Implementation Notes

1. **Sequence**: Follow phases in order due to dependencies
2. **Testing**: Run tests after each phase to ensure correctness
3. **Commit**: Consider creating git commits after major milestones
4. **Documentation**: Update progress.md after completing each phase
5. **Validation**: Test the API manually after Phase 6 (Main Application)

## Next Steps

1. Review and answer all questions in [design-questions.md](design-questions.md)
2. Start with [Phase 1: Database Migrations](phase-01-migrations.md)
3. Track progress in [progress.md](progress.md)
