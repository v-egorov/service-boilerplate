#!/bin/bash
# setup-fork.sh - Configure forked repository with new module path

set -e

# Input validation
NEW_MODULE_PATH="$1"
if [ -z "$NEW_MODULE_PATH" ]; then
    echo "Usage: $0 <new-module-path>"
    echo "Example: $0 github.com/newuser/myproject"
    exit 1
fi

# Validate module path format (basic check for Go module format)
if ! echo "$NEW_MODULE_PATH" | grep -qE '^([a-zA-Z0-9._-]+/)*[a-zA-Z0-9._-]+$'; then
    echo "Error: Invalid module path format. Expected format: domain.com/user/repo"
    exit 1
fi

OLD_MODULE_PATH="github.com/v-egorov/service-boilerplate"

echo "🔄 Replacing module path: $OLD_MODULE_PATH → $NEW_MODULE_PATH"
echo

# Function to replace in file - replace any module path containing service-boilerplate
replace_in_file() {
    local file="$1"
    if [ -f "$file" ]; then
        # Replace any github.com path that ends with "service-boilerplate" (any user/org)
        sed -i "s|github\.com/[^/]\+/service-boilerplate|$NEW_MODULE_PATH|g" "$file"
        echo "✓ Updated: $file"
    fi
}

# Files to update
echo "📝 Updating Go module files..."
replace_in_file "go.mod"
replace_in_file "cli/go.mod"
replace_in_file "migration-orchestrator/go.mod"
replace_in_file "services/auth-service/go.mod"
replace_in_file "services/user-service/go.mod"

echo
echo "🔍 Finding and updating Go source files..."
# Find all .go files and update them
find . -name "*.go" -type f -exec grep -l "$OLD_MODULE_PATH" {} \; | while read -r file; do
    replace_in_file "$file"
done

echo
echo "📜 Updating scripts..."
replace_in_file "scripts/create-service.sh"

echo
echo "🔧 Running 'go mod tidy' on all modules..."

# Function to run go mod tidy in a directory
run_go_mod_tidy() {
    local dir="$1"
    if [ -f "$dir/go.mod" ]; then
        echo "Running go mod tidy in $dir..."
        (cd "$dir" && go mod tidy)
    fi
}

run_go_mod_tidy "."
run_go_mod_tidy "cli"
run_go_mod_tidy "migration-orchestrator"
run_go_mod_tidy "services/auth-service"
run_go_mod_tidy "services/user-service"

echo
echo "✅ Module path replacement complete!"
echo
echo "🎯 Next steps:"
echo "1. Edit .env file for your Docker naming preferences"
echo "2. Run 'make dev' to start development environment"
echo
echo "Your project is now ready with module path: $NEW_MODULE_PATH"