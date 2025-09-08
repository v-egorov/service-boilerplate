#!/bin/bash

# Enhanced Database Seeding Script
# Handles environment-specific seeding with dependency management

set -e

ENVIRONMENT=${1:-"development"}
SERVICE_NAME=${2:-"user-service"}
FORCE=${3:-"false"}

echo "🌱 Enhanced Database Seeding"
echo "   Environment: $ENVIRONMENT"
echo "   Service: $SERVICE_NAME"

# Environment configuration
case $ENVIRONMENT in
    "development")
        SEED_FILES=(
            "scripts/seeds/base/users.sql"
            "scripts/seeds/development/dev_users.sql"
        )
        ;;
    "staging")
        SEED_FILES=(
            "scripts/seeds/base/users.sql"
        )
        ;;
    "production")
        SEED_FILES=(
            "scripts/seeds/base/users.sql"
        )
        ;;
    *)
        echo "❌ Unknown environment: $ENVIRONMENT"
        echo "   Supported: development, staging, production"
        exit 1
        ;;
esac

# Check if database is accessible
echo "🔍 Checking database connectivity..."
if ! docker-compose --env-file .env -f docker/docker-compose.yml exec postgres psql -U postgres -d service_db -c "SELECT 1;" &>/dev/null; then
    echo "❌ Database not accessible"
    exit 1
fi

echo "✅ Database connection established"

# Validate seed files exist
echo "🔍 Validating seed files..."
for seed_file in "${SEED_FILES[@]}"; do
    if [ ! -f "$seed_file" ]; then
        echo "❌ Seed file not found: $seed_file"
        exit 1
    fi
    echo "   ✅ Found: $seed_file"
done

# Check current data count (for development environments)
if [ "$ENVIRONMENT" = "development" ]; then
    echo "📊 Checking current data state..."
    CURRENT_COUNT=$(docker-compose --env-file .env -f docker/docker-compose.yml exec postgres psql -U postgres -d service_db -t -c "SELECT COUNT(*) FROM user_service.users;" 2>/dev/null || echo "0")
    echo "   Current users: $CURRENT_COUNT"

    if [ "$CURRENT_COUNT" -gt 2 ] && [ "$FORCE" != "true" ]; then
        echo "⚠️  Development data already exists ($CURRENT_COUNT users)"
        echo "   Use '$0 development user-service true' to force reseed"
        exit 0
    fi
fi

# Execute seed files
echo "🚀 Executing seed files..."
for seed_file in "${SEED_FILES[@]}"; do
    echo "   Seeding: $(basename "$seed_file")"

    if cat "$seed_file" | docker-compose --env-file .env -f docker/docker-compose.yml exec -T postgres psql -U postgres -d service_db 2>/dev/null; then
        echo "   ✅ Successfully seeded: $(basename "$seed_file")"
    else
        echo "   ❌ Failed to seed: $(basename "$seed_file")"
        exit 1
    fi
done

# Post-seed validation
echo "🔍 Running post-seed validation..."
USER_COUNT=$(docker-compose --env-file .env -f docker/docker-compose.yml exec postgres psql -U postgres -d service_db -t -c "SELECT COUNT(*) FROM user_service.users;" 2>/dev/null || echo "0")

echo ""
echo "🎉 Seeding completed successfully!"
echo "📊 Final state:"
echo "   Users in database: $USER_COUNT"
echo "   Environment: $ENVIRONMENT"
echo "   Service: $SERVICE_NAME"

# Environment-specific summary
case $ENVIRONMENT in
    "development")
        echo ""
        echo "💡 Development notes:"
        echo "   - Test users created for development"
        echo "   - Safe to modify/delete development data"
        echo "   - Run tests against this data"
        ;;
    "staging")
        echo ""
        echo "💡 Staging notes:"
        echo "   - Minimal data for testing"
        echo "   - Mirrors production structure"
        echo "   - Use for pre-production validation"
        ;;
    "production")
        echo ""
        echo "💡 Production notes:"
        echo "   - Only essential system data"
        echo "   - No test data included"
        echo "   - Ready for production use"
        ;;
esac

echo ""
echo "🔄 Next steps:"
echo "   - Run 'make db-counts' to verify data"
echo "   - Test application functionality"
echo "   - Run integration tests"