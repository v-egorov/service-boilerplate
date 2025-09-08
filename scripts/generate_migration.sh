#!/bin/bash

# Automated Migration Generation Script
# Generates new migration files with proper structure and templates

set -e

# Configuration
SERVICE_NAME=${1:-"user-service"}
MIGRATION_NAME=${2:-"new_migration"}
MIGRATION_TYPE=${3:-"table"}  # table, index, data, schema

if [ -z "$SERVICE_NAME" ] || [ -z "$MIGRATION_NAME" ]; then
    echo "❌ Usage: $0 <service-name> <migration-name> [migration-type]"
    echo "   service-name: e.g., user-service"
    echo "   migration-name: e.g., add_user_preferences"
    echo "   migration-type: table (default), index, data, schema"
    exit 1
fi

MIGRATION_DIR="services/${SERVICE_NAME}/migrations"
TEMPLATES_DIR="${MIGRATION_DIR}/templates"

# Create directories if they don't exist
mkdir -p "$MIGRATION_DIR"
mkdir -p "$TEMPLATES_DIR"

# Get next migration number
LAST_MIGRATION=$(find "$MIGRATION_DIR" -name "*.up.sql" | sort | tail -1 | grep -oE '[0-9]+' || echo "0")
NEXT_NUM=$((LAST_MIGRATION + 1))
MIGRATION_ID=$(printf "%06d" $NEXT_NUM)

echo "🚀 Generating migration: ${MIGRATION_ID}_${MIGRATION_NAME}"
echo "   Service: $SERVICE_NAME"
echo "   Type: $MIGRATION_TYPE"

# Generate migration files
UP_FILE="${MIGRATION_DIR}/${MIGRATION_ID}_${MIGRATION_NAME}.up.sql"
DOWN_FILE="${MIGRATION_DIR}/${MIGRATION_ID}_${MIGRATION_NAME}.down.sql"

# Derive table name from migration name (remove 'add_' prefix if present)
TABLE_NAME=$(echo "$MIGRATION_NAME" | sed 's/^add_//')

# Schema name (convert service name to schema name)
SCHEMA_NAME=$(echo "$SERVICE_NAME" | sed 's/-/_/g')

# Generate UP migration based on type
case $MIGRATION_TYPE in
    "table")
        cat > "$UP_FILE" << EOF
-- Migration: ${MIGRATION_ID}_${MIGRATION_NAME}
-- Description: Add new table to ${SCHEMA_NAME} schema
-- Created: $(date)

-- Create table
CREATE TABLE IF NOT EXISTS ${SCHEMA_NAME}.${TABLE_NAME} (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes if needed
-- CREATE INDEX IF NOT EXISTS idx_${TABLE_NAME}_created_at ON ${SCHEMA_NAME}.${TABLE_NAME}(created_at);

-- Add comments
COMMENT ON TABLE ${SCHEMA_NAME}.${TABLE_NAME} IS 'Table for ${TABLE_NAME} functionality';
EOF

        cat > "$DOWN_FILE" << EOF
-- Migration Rollback: ${MIGRATION_ID}_${MIGRATION_NAME}
-- Description: Remove ${TABLE_NAME} table from ${SCHEMA_NAME} schema
-- Created: $(date)

-- Drop table (with CASCADE to remove dependencies)
DROP TABLE IF EXISTS ${SCHEMA_NAME}.${TABLE_NAME} CASCADE;
EOF
        ;;

    "index")
        cat > "$UP_FILE" << EOF
-- Migration: ${MIGRATION_ID}_${MIGRATION_NAME}
-- Description: Add indexes for performance optimization
-- Created: $(date)

-- Add indexes to existing tables
-- Example: CREATE INDEX IF NOT EXISTS idx_users_email ON ${SCHEMA_NAME}.users(email);
-- Example: CREATE INDEX IF NOT EXISTS idx_users_created_at ON ${SCHEMA_NAME}.users(created_at DESC);

-- Add your index creation statements here
EOF

        cat > "$DOWN_FILE" << EOF
-- Migration Rollback: ${MIGRATION_ID}_${MIGRATION_NAME}
-- Description: Remove indexes
-- Created: $(date)

-- Drop indexes (be careful with CONCURRENTLY in production)
-- Example: DROP INDEX IF EXISTS ${SCHEMA_NAME}.idx_users_email;
-- Example: DROP INDEX IF EXISTS ${SCHEMA_NAME}.idx_users_created_at;

-- Add your index removal statements here
EOF
        ;;

    "data")
        cat > "$UP_FILE" << EOF
-- Migration: ${MIGRATION_ID}_${MIGRATION_NAME}
-- Description: Data migration for ${MIGRATION_NAME}
-- Created: $(date)

-- Data migration operations
-- Be careful with large data operations in production

-- Example: UPDATE ${SCHEMA_NAME}.users SET updated_at = CURRENT_TIMESTAMP WHERE updated_at IS NULL;

-- Add your data migration statements here
EOF

        cat > "$DOWN_FILE" << EOF
-- Migration Rollback: ${MIGRATION_ID}_${MIGRATION_NAME}
-- Description: Rollback data migration
-- Created: $(date)

-- Rollback data changes
-- This might be complex for data migrations - consider backup strategies

-- Example: UPDATE ${SCHEMA_NAME}.users SET updated_at = NULL WHERE updated_at = CURRENT_TIMESTAMP;

-- Add your data rollback statements here
EOF
        ;;

    "schema")
        cat > "$UP_FILE" << EOF
-- Migration: ${MIGRATION_ID}_${MIGRATION_NAME}
-- Description: Schema changes for ${MIGRATION_NAME}
-- Created: $(date)

-- Schema modification operations
-- Examples: ALTER TABLE, ADD COLUMN, MODIFY constraints, etc.

-- Example: ALTER TABLE ${SCHEMA_NAME}.users ADD COLUMN phone VARCHAR(20);
-- Example: ALTER TABLE ${SCHEMA_NAME}.users ADD CONSTRAINT chk_phone_format CHECK (phone ~ '^[0-9+\-\s]+$');

-- Add your schema modification statements here
EOF

        cat > "$DOWN_FILE" << EOF
-- Migration Rollback: ${MIGRATION_ID}_${MIGRATION_NAME}
-- Description: Rollback schema changes
-- Created: $(date)

-- Rollback schema changes
-- Be very careful with schema rollbacks in production

-- Example: ALTER TABLE ${SCHEMA_NAME}.users DROP COLUMN IF EXISTS phone;

-- Add your schema rollback statements here
EOF
        ;;

    *)
        echo "❌ Unknown migration type: $MIGRATION_TYPE"
        echo "   Supported types: table, index, data, schema"
        exit 1
        ;;
esac

# Update dependencies file
DEPENDENCIES_FILE="${MIGRATION_DIR}/dependencies.json"

if [ -f "$DEPENDENCIES_FILE" ]; then
    # Get last migration ID for dependency
    LAST_MIGRATION_ID=$(jq -r '.migrations | keys | last' "$DEPENDENCIES_FILE" 2>/dev/null || echo "")

    # Add new migration to dependencies
    jq --arg id "${MIGRATION_ID}" \
       --arg name "${MIGRATION_NAME}" \
       --arg type "${MIGRATION_TYPE}" \
       --arg depends_on "${LAST_MIGRATION_ID:-[]}" \
       '.migrations[$id] = {
         "description": $name,
         "depends_on": ($depends_on | if . == "[]" then [] else [$depends_on] end),
         "migration_type": $type,
         "affects_tables": [],
         "estimated_duration": "30s",
         "risk_level": "medium",
         "rollback_safe": true,
         "created_at": (now | strftime("%Y-%m-%dT%H:%M:%SZ"))
       }' "$DEPENDENCIES_FILE" > "${DEPENDENCIES_FILE}.tmp" && mv "${DEPENDENCIES_FILE}.tmp" "$DEPENDENCIES_FILE"
fi

# Create documentation template
DOC_FILE="${MIGRATION_DIR}/docs/migration_${MIGRATION_ID}.md"
mkdir -p "${MIGRATION_DIR}/docs"

cat > "$DOC_FILE" << EOF
# Migration: ${MIGRATION_ID}_${MIGRATION_NAME}

## Overview
**Type:** ${MIGRATION_TYPE}
**Service:** ${SERVICE_NAME}
**Schema:** ${SCHEMA_NAME}
**Created:** $(date)

## Description
Brief description of what this migration does.

## Changes Made

### Database Changes
- List the specific database changes made

### Affected Tables
- ${SCHEMA_NAME}.[table_name]

## Rollback Plan
Describe how to rollback this migration if needed.

## Testing
- [ ] Unit tests updated
- [ ] Integration tests pass
- [ ] Data integrity verified
- [ ] Performance impact assessed

## Notes
Any additional notes or considerations.

## Risk Assessment
- **Risk Level:** [Low/Medium/High]
- **Estimated Duration:** [Time estimate]
- **Rollback Safety:** [Safe/Requires Caution/Dangerous]
EOF

echo "✅ Migration files created:"
echo "   📄 UP:   $UP_FILE"
echo "   📄 DOWN: $DOWN_FILE"
echo "   📋 DOC:  $DOC_FILE"

if [ -f "$DEPENDENCIES_FILE" ]; then
    echo "   🔗 DEPS: Updated $DEPENDENCIES_FILE"
fi

echo ""
echo "🎉 Migration generation completed!"
echo ""
echo "📋 Next steps:"
echo "   1. Edit the generated SQL files with your specific changes"
echo "   2. Update the documentation in $DOC_FILE"
echo "   3. Run './scripts/validate_migration.sh $SERVICE_NAME' to validate"
echo "   4. Test the migration with 'make db-migrate-up'"
echo ""
echo "💡 Pro tip: Always test migrations on a copy of production data first!"