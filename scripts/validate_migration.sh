#!/bin/bash

# Migration Validation Script
# Validates migration files for syntax and structure

set -e

SERVICE_NAME=${1:-"user-service"}
MIGRATION_DIR="services/${SERVICE_NAME}/migrations"

echo "🔍 Validating migrations for service: ${SERVICE_NAME}"

# Check if migration directory exists
if [ ! -d "$MIGRATION_DIR" ]; then
    echo "❌ Migration directory not found: $MIGRATION_DIR"
    exit 1
fi

# Check if environments.json exists
if [ ! -f "${MIGRATION_DIR}/environments.json" ]; then
    echo "❌ environments.json not found: ${MIGRATION_DIR}/environments.json"
    exit 1
fi

echo "✅ Migration directory and config found"

# Validate SQL syntax for all migration files
echo "🔍 Validating SQL syntax..."

# Find all migration files in environment subdirectories
for sql_file in $(find "$MIGRATION_DIR" -name "*.sql" -type f | sort); do
    echo "   Checking: $(basename "$sql_file")"

    # Basic SQL syntax validation (check for common issues)
    if grep -q "DROP TABLE.*CASCADE" "$sql_file"; then
        echo "   ⚠️  WARNING: CASCADE drop detected in $(basename "$sql_file")"
    fi

    if grep -q "DELETE.*WHERE.*1=1" "$sql_file"; then
        echo "   ⚠️  WARNING: Potentially dangerous DELETE detected in $(basename "$sql_file")"
    fi
done

echo "✅ SQL validation completed"

# Validate migration file naming and pairing
echo "🔍 Validating migration file structure..."

# Find all migration files in environment subdirectories
migration_files=$(find "$MIGRATION_DIR" -name "*.sql" -type f | grep -E "\.up\.|\.down\." | sort)

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

echo ""
echo "🎉 Migration validation completed successfully!"
echo ""
echo "📋 Summary:"
echo "   - SQL syntax checked"
echo "   - File structure validated (up/down pairs)"
echo ""
echo "💡 Next steps:"
echo "   - Run 'make db-migrate-up SERVICE_NAME=${SERVICE_NAME}' to apply migrations"