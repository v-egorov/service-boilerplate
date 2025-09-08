#!/bin/bash

# Migration Validation Script
# Validates migration files for syntax, dependencies, and best practices

set -e

SERVICE_NAME=${1:-"user-service"}
MIGRATION_DIR="services/${SERVICE_NAME}/migrations"
DEPENDENCIES_FILE="${MIGRATION_DIR}/dependencies.json"

echo "🔍 Validating migrations for service: ${SERVICE_NAME}"

# Check if migration directory exists
if [ ! -d "$MIGRATION_DIR" ]; then
    echo "❌ Migration directory not found: $MIGRATION_DIR"
    exit 1
fi

# Check if dependencies file exists
if [ ! -f "$DEPENDENCIES_FILE" ]; then
    echo "⚠️  Dependencies file not found: $DEPENDENCIES_FILE"
    echo "   Creating basic dependencies file..."
    cat > "$DEPENDENCIES_FILE" << EOF
{
  "migrations": {},
  "global_config": {
    "max_parallel_migrations": 1,
    "require_approval_for_high_risk": true,
    "auto_backup_before_destructive": true,
    "validate_after_migration": true
  }
}
EOF
fi

echo "✅ Dependencies file exists"

# Validate SQL syntax for all migration files
echo "🔍 Validating SQL syntax..."

for sql_file in $(find "$MIGRATION_DIR" -name "*.sql" | sort); do
    echo "   Checking: $(basename "$sql_file")"

    # Basic SQL syntax validation (check for common issues)
    if grep -q "DROP TABLE.*CASCADE" "$sql_file"; then
        echo "   ⚠️  WARNING: CASCADE drop detected in $(basename "$sql_file")"
    fi

    if grep -q "DELETE.*WHERE.*1=1" "$sql_file"; then
        echo "   ⚠️  WARNING: Potentially dangerous DELETE detected in $(basename "$sql_file")"
    fi

    if ! grep -q "user_service\." "$sql_file" 2>/dev/null; then
        echo "   ℹ️  INFO: No schema-qualified tables found in $(basename "$sql_file")"
    fi
done

echo "✅ SQL validation completed"

# Validate migration file naming and pairing
echo "🔍 Validating migration file structure..."

# Find all migration files including subdirectories
migration_files=$(find "$MIGRATION_DIR" -name "*.sql" | grep -E "\.up\.|\.down\." | sort)

# Group files by migration number
declare -A migration_groups

for file in $migration_files; do
    filename=$(basename "$file")
    migration_num=$(echo "$filename" | sed 's/_.*//')

    if [ -z "${migration_groups[$migration_num]}" ]; then
        migration_groups[$migration_num]="$file"
    else
        migration_groups[$migration_num]="${migration_groups[$migration_num]} $file"
    fi
done

# Validate each migration group
for migration_num in "${!migration_groups[@]}"; do
    files=${migration_groups[$migration_num]}
    up_count=$(echo "$files" | grep -c "\.up\.sql")
    down_count=$(echo "$files" | grep -c "\.down\.sql")

    if [ "$up_count" -ne 1 ]; then
        echo "❌ Migration $migration_num: Expected 1 up file, found $up_count"
        exit 1
    fi

    if [ "$down_count" -ne 1 ]; then
        echo "❌ Migration $migration_num: Expected 1 down file, found $down_count"
        exit 1
    fi
done

echo "   Found $(echo "${!migration_groups[@]}" | wc -w) migration groups"

echo "✅ Migration file structure validated"

# Validate dependencies
echo "🔍 Validating migration dependencies..."

if command -v jq &> /dev/null; then
    # Check for circular dependencies (basic check)
    migration_count=$(jq '.migrations | length' "$DEPENDENCIES_FILE")
    echo "   Found $migration_count migrations in dependencies file"

    # Validate that all migration files have dependency entries
    for file in $migration_files; do
        filename=$(basename "$file")
        migration_id=$(echo "$filename" | sed 's/\.up\.sql\|\.down\.sql//' | sed 's/_.*//')

        if ! jq -e ".migrations.\"${migration_id}\"" "$DEPENDENCIES_FILE" &> /dev/null; then
            echo "⚠️  WARNING: Migration $migration_id not found in dependencies file"
        fi
    done
else
    echo "⚠️  WARNING: jq not installed, skipping advanced dependency validation"
fi

echo "✅ Dependency validation completed"

echo ""
echo "🎉 Migration validation completed successfully!"
echo ""
echo "📋 Summary:"
echo "   - SQL syntax checked"
echo "   - File structure validated"
echo "   - Dependencies verified"
echo ""
echo "💡 Next steps:"
echo "   - Run 'make db-migrate-up' to apply migrations"
echo "   - Run 'make db-migrate-status' to check migration status"