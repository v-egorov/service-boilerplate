# Objects Relationships System

## Overview

This document describes the design and implementation plan for adding a general-purpose M:N relationship system to the objects-service. The system uses the Class Table Inheritance (CTI) pattern where relationship types and instances are represented as objects with concrete tables.

## Background

The objects-service currently supports:
- **Object type hierarchies**: Parent-child relationships within `object_types` table
- **Object hierarchies**: Parent-child relationships within `objects` table via `parent_object_id`
- **JSONB metadata**: Flexible attributes stored in `metadata` JSONB field

However, there is no support for:
- Cross-type relationships (e.g., linking a Product to a Category)
- M:N relationships (many-to-many)
- Relationship metadata (type, cardinality, attributes)
- Directional relationships beyond parent-child

## Design Decisions

### CTI Pattern for Relationships

Both relationship types and relationship instances are implemented using the CTI pattern:

1. **Relationship Types** are objects (instances in `objects` table) with type "RelationshipType"
2. **Relationship Instances** are objects (instances in `objects` table) with type "Relationship"
3. Each has a corresponding concrete CTI table extending the base `objects` table

### Why CTI for Both Types and Instances?

- Relationship types become first-class objects (can use existing object features: metadata, tags, versioning, RBAC)
- Single unified CTI pattern from base `objects` table
- Relationship types can have their own relationships (meta-relationships)
- Simpler, more consistent architecture

### Simplified CTI Approach

This implementation uses explicit, type-specific patterns rather than the full dynamic CTI infrastructure described in Phase 10. This approach:
- Uses straightforward, explicit SQL queries (no reflection/generics)
- Relationship tables are "known" types, not dynamically queried
- Can serve as an example for future CTI implementations
- Simpler to understand and maintain

### Bidirectional Relationships

Bidirectional relationships are handled via `reverse_type_key` field:
- Each relationship type can define its reverse (e.g., "contains" → "contained_by")
- No automatic synchronization of relationship instances
- Single metadata field shared between directions

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│ object_types                                                     │
│ ├── id=1: "RelationshipType"  (special marker)                 │
│ ├── id=2: "Relationship"      (special marker)                 │
│ └── ... other types                                             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ (instances created as objects)
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ objects (base CTI table)                                         │
│ ├── id=100: RelationshipType instance "contains"               │
│ ├── id=101: RelationshipType instance "belongs_to"            │
│ ├── id=200: Relationship instance (PortfolioA → AssetX)         │
│ └── id=201: Relationship instance (DocA → DocB)                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ (CTI concrete tables)
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ objects_relationship_types (CTI concrete)                       │
│ ├── object_id PK → objects.id                                   │
│ ├── type_key VARCHAR(100) NOT NULL UNIQUE                        │
│ ├── relationship_name VARCHAR(255)                               │
│ ├── reverse_type_key VARCHAR(100) NULL                          │
│ ├── cardinality VARCHAR(20) NOT NULL                            │
│ ├── required BOOLEAN DEFAULT FALSE                               │
│ ├── min_count INTEGER DEFAULT 0                                 │
│ ├── max_count INTEGER DEFAULT -1                                │
│ ├── validation_rules JSONB                                       │
│ ├── created_by, updated_by VARCHAR(255)                         │
│ └── created_at, updated_at TIMESTAMP                            │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ objects_relationships (CTI concrete)                            │
│ ├── object_id PK → objects.id                                    │
│ ├── source_object_id → objects.id                                │
│ ├── target_object_id → objects.id                                │
│ ├── relationship_type_id → objects.id                            │
│ ├── status VARCHAR(20) DEFAULT 'active'                         │
│ ├── relationship_metadata JSONB                                 │
│ ├── created_by, updated_by VARCHAR(255)                         │
│ ├── created_at, updated_at TIMESTAMP                            │
│ ├── INDEX on source_object_id                                    │
│ ├── INDEX on target_object_id                                    │
│ └── INDEX on (relationship_type_id, status)                     │
└─────────────────────────────────────────────────────────────────┘
```

## Field Definitions

### objects_relationship_types

| Field | Type | Description |
|-------|------|-------------|
| object_id | BIGINT | FK to objects(id) - the object representing this type |
| type_key | VARCHAR(100) | Unique identifier for API use (e.g., "belongs_to") |
| relationship_name | VARCHAR(255) | Human-readable name (e.g., "belongs to") |
| reverse_type_key | VARCHAR(100) | Points to reverse relationship type_key |
| cardinality | VARCHAR(20) | one_to_one, one_to_many, many_to_one, many_to_many |
| required | BOOLEAN | Whether this relationship is required |
| min_count | INTEGER | Minimum number of relationships allowed |
| max_count | INTEGER | Maximum number (-1 = unlimited) |
| validation_rules | JSONB | Custom validation rules |
| created_by, updated_by | VARCHAR(255) | Audit fields |
| created_at, updated_at | TIMESTAMP | Timestamps |

### objects_relationships

| Field | Type | Description |
|-------|------|-------------|
| object_id | BIGINT | FK to objects(id) - the object representing this relationship |
| source_object_id | BIGINT | The source/tail of the relationship |
| target_object_id | BIGINT | The target/head of the relationship |
| relationship_type_id | BIGINT | FK to objects(id) - points to relationship type object |
| status | VARCHAR(20) | active, inactive, deprecated |
| relationship_metadata | JSONB | Relationship-specific attributes |
| created_by, updated_by | VARCHAR(255) | Audit fields |
| created_at, updated_at | TIMESTAMP | Timestamps |

## Implementation Phases

### Phase R1: Relationship Type System

Implementation of the relationship type management system.

**Estimated Time:** 8-10 hours

**See:** [phase-r1-relationship-types.md](phase-r1-relationship-types.md)

### Phase R2: Relationship Instance System

Implementation of relationship instance management and queries.

**Estimated Time:** 12-15 hours

**See:** [phase-r2-relationship-instances.md](phase-r2-relationship-instances.md)

### Phase R3: Advanced Features

Additional features for relationship system (future work).

**Estimated Time:** 8-10 hours (optional)

**See:** [phase-r3-advanced-features.md](phase-r3-advanced-features.md)

## API Endpoints

### Relationship Types

```
POST   /api/v1/relationship-types          # Create type
GET    /api/v1/relationship-types         # List all
GET    /api/v1/relationship-types/{key}  # Get by type_key
PUT    /api/v1/relationship-types/{key}  # Update
DELETE /api/v1/relationship-types/{key}  # Delete
```

### Relationships

```
POST   /api/v1/relationships                        # Create relationship
GET    /api/v1/relationships                      # List with filters
GET    /api/v1/relationships/{public_id}           # Get by ID
DELETE /api/v1/relationships/{public_id}           # Delete
GET    /api/v1/objects/{id}/relationships         # All relationships for object
GET    /api/v1/objects/{id}/relationships/{key}   # Relationships by type
```

## Testing

### Unit Tests

Follow existing objects-service patterns:
- Manual mock structs for repositories
- Testify assertions
- Test files in same package directories

### Integration Tests

- End-to-end test script: `scripts/test-relationships-e2e.sh`
- Dev seed migrations for test data

### Dev Migrations

- `development/00000X_dev_relationship_types.up.sql` - Seed relationship types
- `development/00000X_dev_relationships.up.sql` - Seed test relationships

## Out of Scope (Separate Topics)

- **RBAC for relationships**: Will require extending RBAC architecture to support object-type-based permissions
- **Dynamic CTI infrastructure**: Full dynamic CTI pattern from Phase 10
- **Natural object identifiers**: Whole separate task for human-readable identifiers

## Related Documentation

- [Phase 10: Class Table Inheritance](../phase-10-class-table-inheritance.md) - CTI pattern details
- [Objects Service Refactoring README](../README.md) - Overall refactoring progress

## Status

| Phase | Status | Notes |
|-------|--------|-------|
| R1: Relationship Types | **COMPLETED** | 14 unit tests passing |
| R2: Relationship Instances | Not Started | |
| R3: Advanced Features | Not Started | Future work |

## Example Use Cases

### Portfolio Management

```
Portfolio "contains" → Asset (many-to-many)
Asset "belongs_to" → Portfolio (many-to-many)
Asset "derived_from" → UnderlyingAsset (many-to-one)
```

### Document Management

```
Document "references" → Document (many-to-many)
Report "contains" → Chart (one-to-many)
Chart "visualizes" → Dataset (many-to-one)
```

### Organizational Structure

```
Employee "reports_to" → Employee (many-to-one)
Department "has" → Employee (one-to-many)
Project "assigned_to" → Employee (many-to-many)
```
