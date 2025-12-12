#!/bin/bash
# update-docker-prefix.sh - Update Docker project prefix and all derived names

set -e

# Input validation
NEW_PREFIX="$1"
ENV_FILE="${2:-.env.development}"

if [ -z "$NEW_PREFIX" ]; then
    echo "Usage: $0 <new-docker-prefix> [env-file]"
    echo "Examples:"
    echo "  $0 myproject                    # Update .env.development"
    echo "  $0 myproject .env.production    # Update .env.production"
    echo "  $0 myproject .env               # Update .env"
    echo ""
    echo "This will update DOCKER_PROJECT_PREFIX and all derived Docker names"
    echo "in the specified environment file."
    exit 1
fi

# Validate prefix format (alphanumeric, hyphens, underscores only)
if ! echo "$NEW_PREFIX" | grep -qE '^[a-zA-Z0-9_-]+$'; then
    echo "Error: Invalid prefix format. Use only letters, numbers, hyphens, and underscores."
    exit 1
fi

# Check for reasonable length
if [ ${#NEW_PREFIX} -gt 50 ]; then
    echo "Error: Prefix too long (max 50 characters)"
    exit 1
fi

# Check if env file exists
if [ ! -f "$ENV_FILE" ]; then
    echo "Error: Environment file '$ENV_FILE' not found"
    exit 1
fi

# Function to detect current prefix
detect_current_prefix() {
    local env_file="$1"
    local prefix=$(grep "^DOCKER_PROJECT_PREFIX=" "$env_file" | cut -d'=' -f2)
    if [ -z "$prefix" ]; then
        echo "service-boilerplate"  # Default for new files
    else
        echo "$prefix"
    fi
}

# Detect current prefix
CURRENT_PREFIX=$(detect_current_prefix "$ENV_FILE")

# Check if new prefix is different
if [ "$NEW_PREFIX" = "$CURRENT_PREFIX" ]; then
    echo "Info: New prefix '$NEW_PREFIX' is the same as current prefix. No changes needed."
    exit 0
fi

echo "🔄 Updating Docker project prefix in $ENV_FILE: $CURRENT_PREFIX → $NEW_PREFIX"
echo

# Function to backup file
backup_file() {
    local file="$1"
    cp "$file" "${file}.backup.$(date +%Y%m%d_%H%M%S)"
    echo "📋 Backed up: $file"
}

# Function to update file
update_file() {
    local file="$1"
    # Update DOCKER_PROJECT_PREFIX
    sed -i "s|^DOCKER_PROJECT_PREFIX=.*|DOCKER_PROJECT_PREFIX=$NEW_PREFIX|" "$file"

    # Update all derived names that contain the current prefix
    sed -i "s|$CURRENT_PREFIX|$NEW_PREFIX|g" "$file"

    echo "✓ Updated: $file"
}

# Backup and update the specified env file
echo "📝 Updating environment file..."
backup_file "$ENV_FILE"
update_file "$ENV_FILE"

echo
echo "🔍 Validating changes..."

# Basic validation - check if docker-compose can parse with new values
if command -v docker-compose &> /dev/null; then
    if [ -f "docker/docker-compose.yml" ]; then
        echo "Testing docker-compose configuration..."
        if docker-compose -f docker/docker-compose.yml config --quiet 2>/dev/null; then
            echo "✅ Docker Compose configuration is valid"
        else
            echo "⚠️  Warning: Docker Compose configuration may have issues"
            echo "   Check your $ENV_FILE file for any syntax errors"
        fi
    fi
fi

echo
echo "✅ Docker prefix update complete!"
echo
echo "🎯 What was changed in $ENV_FILE:"
echo "  • DOCKER_PROJECT_PREFIX updated to: $NEW_PREFIX"
echo "  • All container names, volumes, and networks updated"
echo "  • Examples: ${CURRENT_PREFIX}-api-gateway → ${NEW_PREFIX}-api-gateway"
echo
echo "🚀 Next steps:"
echo "  • Run 'docker-compose down' if containers are running"
echo "  • Run 'make dev' or 'make prod' to start with new names"
echo "  • Update any scripts/docs that reference old container names"
echo
echo "📋 Backup created: ${ENV_FILE}.backup.* (restore with: cp ${ENV_FILE}.backup.* $ENV_FILE)"