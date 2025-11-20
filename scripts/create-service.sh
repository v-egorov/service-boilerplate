#!/bin/bash

set -e

# Cleanup function
cleanup() {
    local exit_code=$?
    if [ $exit_code -ne 0 ] && [ -d "$SERVICE_DIR" ]; then
        echo "Script failed, cleaning up..."
        rm -rf "$SERVICE_DIR"
        echo "Cleanup completed"
    fi
    exit $exit_code
}

trap cleanup EXIT

# Default values
CREATE_DB_SCHEMA=true
UPDATE_DOCS=true

# Check for help first
if [[ $# -eq 1 && ($1 == "-h" || $1 == "--help") ]]; then
    echo "Usage: $0 <service-name> <port> [options]"
    echo ""
    echo "Arguments:"
    echo "  service-name    Name of the service (lowercase, hyphens allowed)"
    echo "  port           Port number (1024-65535)"
    echo ""
    echo "Options:"
    echo "  --no-db-schema    Skip database schema creation"
    echo "  --no-docs-update  Skip documentation updates"
    echo "  -h, --help        Show this help message"
    echo ""
    echo "Example: $0 product-service 8082"
    echo "Example: $0 user-service 8081 --no-db-schema"
    exit 0
fi

# Parse options (only at the end)
for arg in "$@"; do
    case $arg in
    --no-db-schema)
        CREATE_DB_SCHEMA=false
        ;;
    --no-docs-update)
        UPDATE_DOCS=false
        ;;
    esac
done

# Remove options from arguments to get positional args
ARGS=()
for arg in "$@"; do
    if [[ $arg != --* ]]; then
        ARGS+=("$arg")
    fi
done

# Check arguments
if [ ${#ARGS[@]} -ne 2 ]; then
    echo "Usage: $0 <service-name> <port> [options]"
    echo "Use -h or --help for more information"
    exit 1
fi

SERVICE_NAME=${ARGS[0]}
PORT=${ARGS[1]}

# Validate service name (lowercase, hyphens allowed)
if [[ ! $SERVICE_NAME =~ ^[a-z-]+[a-z]$ ]]; then
    echo "Error: Service name must be lowercase and contain only letters and hyphens"
    echo "Examples: user-service, product-catalog, order-management"
    exit 1
fi

# Check for reserved names
RESERVED_NAMES=("api-gateway" "postgres" "migration" "database" "db")
for reserved in "${RESERVED_NAMES[@]}"; do
    if [ "$SERVICE_NAME" = "$reserved" ]; then
        echo "Error: Service name '$SERVICE_NAME' is reserved"
        exit 1
    fi
done

# Validate port
if [[ ! $PORT =~ ^[0-9]+$ ]] || [ $PORT -lt 1024 ] || [ $PORT -gt 65535 ]; then
    echo "Error: Port must be a number between 1024 and 65535"
    exit 1
fi

# Check if port is already in use
if lsof -Pi :$PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo "Warning: Port $PORT appears to be in use"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check if port conflicts with existing services in docker-compose
if grep -q ":$PORT:" docker/docker-compose.yml 2>/dev/null; then
    echo "Error: Port $PORT is already used by another service in docker-compose.yml"
    exit 1
fi

SERVICE_DIR="services/$SERVICE_NAME"
TEMPLATE_DIR="templates/service-template"

echo "Creating new service: $SERVICE_NAME on port $PORT"
echo "Database schema creation: $CREATE_DB_SCHEMA"
echo "Documentation update: $UPDATE_DOCS"

# Check if service already exists
if [ -d "$SERVICE_DIR" ]; then
    echo "Error: Service $SERVICE_NAME already exists"
    exit 1
fi

# Check if template exists
if [ ! -d "$TEMPLATE_DIR" ]; then
    echo "Error: Service template not found at $TEMPLATE_DIR"
    echo "Please ensure the template directory exists"
    exit 1
fi

# Copy template
echo "Copying template..."
if ! cp -r "$TEMPLATE_DIR" "$SERVICE_DIR"; then
    echo "Error: Failed to copy template directory"
    exit 1
fi

# Replace variables in files
echo "Customizing service..."

# Function to replace variables in file
replace_vars() {
    local file=$1

    # Replace import placeholders first (before PORT replacement)
    sed -i "s|// ENTITY_IMPORT_HANDLERS|\"github.com/v-egorov/service-boilerplate/services/$SERVICE_NAME/internal/handlers\"|g" "$file"
    sed -i "s|// ENTITY_IMPORT_REPOSITORY|\"github.com/v-egorov/service-boilerplate/services/$SERVICE_NAME/internal/repository\"|g" "$file"
    sed -i "s|// ENTITY_IMPORT_SERVICES|\"github.com/v-egorov/service-boilerplate/services/$SERVICE_NAME/internal/services\"|g" "$file"
    sed -i "s|// ENTITY_IMPORT_MODELS|\"github.com/v-egorov/service-boilerplate/services/$SERVICE_NAME/internal/models\"|g" "$file"

    # Replace hardcoded template imports
    sed -i "s|github.com/v-egorov/service-boilerplate/templates/service-template/internal/handlers|github.com/v-egorov/service-boilerplate/services/$SERVICE_NAME/internal/handlers|g" "$file"
    # Remove aliased template imports
    sed -i "/repository \"github.com\/v-egorov\/service-boilerplate\/templates\/service-template\/internal\/repository\"/d" "$file"
    sed -i "/services \"github.com\/v-egorov\/service-boilerplate\/templates\/service-template\/internal\/services\"/d" "$file"

    # Replace other placeholders
    sed -i "s|// ENTITY_REPO_INIT|entityRepo := repository.NewEntityRepository(db.GetPool(), logger.Logger)|g" "$file"
    sed -i "s|// ENTITY_SERVICE_INIT|entityService := services.NewEntityService(entityRepo, logger.Logger)|g" "$file"
    sed -i "s|// ENTITY_HANDLER_INIT|entityHandler := handlers.NewEntityHandler(entityService, logger.Logger)|g" "$file"
    sed -i "s|// ENTITY_ROUTES|v1 := router.Group(\"/api/v1\")\n\t{\n\t\tentities := v1.Group(\"/entities\")\n\t\t{\n\t\t\tv1.POST(\"\", entityHandler.CreateEntity)\n\t\t\tv1.GET(\"/:id\", entityHandler.GetEntity)\n\t\t\tv1.PUT(\"/:id\", entityHandler.UpdateEntity)\n\t\t\tv1.DELETE(\"/:id\", entityHandler.DeleteEntity)\n\t\t\tv1.GET(\"\", entityHandler.ListEntities)\n\t\t}\n\t}|g" "$file"

    # Replace service name and port
    sed -i "s/SERVICE_NAME/$SERVICE_NAME/g" "$file"
    sed -i "s/PORT/$PORT/g" "$file"

    # Replace schema name
    SCHEMA_VALUE=$(echo "$SERVICE_NAME" | sed 's/-/_/g')
    sed -i "s/SCHEMA_NAME/$SCHEMA_VALUE/g" "$file"

    # Replace import paths
    sed -i "s|github.com/v-egorov/service-boilerplate/services/SERVICE_NAME|github.com/v-egorov/service-boilerplate/services/$SERVICE_NAME|g" "$file"

    # Replace service-specific placeholders
    sed -i "s/{{SERVICE_NAME}}/$SERVICE_NAME/g" "$file"
    sed -i "s/{{PORT}}/$PORT/g" "$file"
}

# Find and replace in all files
find "$SERVICE_DIR" -type f \( -name "*.go" -o -name "*.yaml" -o -name "*.md" -o -name "*.sql" -o -name "*.toml" -o -name "*.json" -o -name "Dockerfile*" \) | while read -r file; do
    replace_vars "$file"
done

# Update docker-compose.yml
echo "Updating docker-compose.yml..."
SERVICE_NAME_UPPER=$(echo $SERVICE_NAME | tr '[:lower:]' '[:upper:]' | tr '-' '_')

# Create temporary files for service and volume definitions
SERVICE_DEF_FILE=$(mktemp)
VOLUME_DEF_FILE=$(mktemp)

# Create service definition
cat >"$SERVICE_DEF_FILE" <<EOF

  $SERVICE_NAME:
    build:
      context: ..
      dockerfile: services/$SERVICE_NAME/Dockerfile
    image: \${${SERVICE_NAME_UPPER}_SERVICE_IMAGE}
    container_name: \${${SERVICE_NAME_UPPER}_SERVICE_CONTAINER}
    ports:
      - "\${${SERVICE_NAME_UPPER}_SERVICE_PORT:-$PORT}:\${${SERVICE_NAME_UPPER}_SERVICE_PORT:-$PORT}"
    environment:
      - APP_ENV=\${APP_ENV:-production}
      - SERVER_PORT=\${${SERVICE_NAME_UPPER}_SERVICE_PORT:-$PORT}
      - DATABASE_HOST=\${POSTGRES_NAME}
      - DATABASE_PORT=5432
      - DATABASE_USER=\${DATABASE_USER:-postgres}
      - DATABASE_PASSWORD=\${DATABASE_PASSWORD:-postgres}
      - DATABASE_NAME=\${DATABASE_NAME:-service_db}
      - DATABASE_SSL_MODE=disable
      - LOGGING_LEVEL=\${LOGGING_LEVEL:-info}
      - LOGGING_FORMAT=\${LOGGING_FORMAT:-json}
      - LOGGING_OUTPUT=file
      - LOGGING_DUAL_OUTPUT=\${LOGGING_DUAL_OUTPUT:-true}
      - LOGGING_STRIP_ANSI_FROM_FILES=\${LOGGING_STRIP_ANSI_FROM_FILES:-true}
      - AUTH_SERVICE_URL=\${AUTH_SERVICE_URL:-http://auth-service:8083}
      - USER_SERVICE_URL=\${USER_SERVICE_URL:-http://user-service:8081}
    depends_on:
      postgres:
        condition: service_healthy
    volumes:
      - ${SERVICE_NAME}_logs:/app/logs
    networks:
      service-network:
        aliases:
          - ${SERVICE_NAME}
          - ${SERVICE_NAME}-service
          - ${SERVICE_NAME}-svc
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:\${${SERVICE_NAME_UPPER}_SERVICE_PORT:-$PORT}/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
    restart: unless-stopped
EOF

# Create volume definition
cat >"$VOLUME_DEF_FILE" <<EOF

  ${SERVICE_NAME}_service_tmp:
    name: \${${SERVICE_NAME_UPPER}_SERVICE_TMP_VOLUME}
    driver: local
    driver_opts:
      type: none
      o: bind
      device: \${PWD}/docker/volumes/${SERVICE_NAME}/tmp

  ${SERVICE_NAME}_logs:
    name: \${${SERVICE_NAME_UPPER}_LOGS_VOLUME}
    driver: local
    driver_opts:
      type: none
      o: bind
      device: \${PWD}/docker/volumes/${SERVICE_NAME}/logs
EOF

# Insert service definition before volumes section
awk '
BEGIN { found=0 }
/^volumes:/ && !found {
    while ((getline line < "'$SERVICE_DEF_FILE'") > 0) {
        print line
    }
    found=1
}
{ print }
' docker/docker-compose.yml >docker-compose.tmp && mv docker-compose.tmp docker/docker-compose.yml

# Insert volume definition before networks section
awk '
BEGIN { found=0 }
/^networks:/ && !found {
    while ((getline line < "'$VOLUME_DEF_FILE'") > 0) {
        print line
    }
    found=1
}
{ print }
' docker/docker-compose.yml >docker-compose.tmp && mv docker-compose.tmp docker/docker-compose.yml

# Clean up temporary files
rm -f "$SERVICE_DEF_FILE" "$VOLUME_DEF_FILE"

# Update Promtail volume mounts in docker-compose.yml
echo "Updating Promtail volume mounts..."
PROMTAIL_VOLUME_LINE="      - ./volumes/${SERVICE_NAME}/logs:/var/log/${SERVICE_NAME}:ro"

# Find the promtail service and add the volume mount after the last volume line
awk '
BEGIN { in_promtail=0; in_volumes=0 }
/^  promtail:/ { in_promtail=1 }
/^    volumes:/ && in_promtail { in_volumes=1 }
/^    [a-zA-Z]/ && in_promtail && in_volumes && !/^    volumes:/ && !/^      - / {
    # Insert the new volume mount before the next service property
    print "'"$PROMTAIL_VOLUME_LINE"'"
    in_volumes=0
}
{ print }
' docker/docker-compose.yml >docker-compose.tmp && mv docker-compose.tmp docker/docker-compose.yml

# Update docker-compose.override.yml
echo "Updating docker-compose.override.yml..."
cat >>docker/docker-compose.override.yml <<EOF

  $SERVICE_NAME:
    build:
      context: ..
      dockerfile: services/$SERVICE_NAME/Dockerfile.dev
    environment:
      - APP_ENV=development
      - LOGGING_LEVEL=debug
      - LOGGING_FORMAT=json
      - LOGGING_OUTPUT=file
      - LOGGING_DUAL_OUTPUT=${LOGGING_DUAL_OUTPUT:-true}
      - LOGGING_STRIP_ANSI_FROM_FILES=${LOGGING_STRIP_ANSI_FROM_FILES:-true}
    volumes:
      - ../services/$SERVICE_NAME:/app/services/$SERVICE_NAME
      - ../common:/app/common
      - ${SERVICE_NAME}_service_tmp:/app/services/$SERVICE_NAME/tmp
      - ${SERVICE_NAME}_logs:/app/logs
    ports:
      - "\${${SERVICE_NAME_UPPER}_SERVICE_PORT}:\${${SERVICE_NAME_UPPER}_SERVICE_PORT}"
    working_dir: /app/services/$SERVICE_NAME
    networks:
      service-network:
        aliases:
          - ${SERVICE_NAME}
          - ${SERVICE_NAME}-service
          - ${SERVICE_NAME}-svc
EOF

# Update .env.development file
echo "Updating .env.development file..."
# Add service configuration to .env.development file
echo "" >>.env.development
echo "# $SERVICE_NAME Service Configuration" >>.env.development
echo "${SERVICE_NAME_UPPER}_SERVICE_NAME=$SERVICE_NAME" >>.env.development
echo "${SERVICE_NAME_UPPER}_SERVICE_PORT=$PORT" >>.env.development
echo "${SERVICE_NAME_UPPER}_SERVICE_IMAGE=docker-$SERVICE_NAME" >>.env.development
echo "${SERVICE_NAME_UPPER}_SERVICE_CONTAINER=service-boilerplate-$SERVICE_NAME" >>.env.development
echo "${SERVICE_NAME_UPPER}_SERVICE_TMP_VOLUME=service-boilerplate-${SERVICE_NAME}-tmp" >>.env.development
echo "${SERVICE_NAME_UPPER}_LOGS_VOLUME=service-boilerplate-${SERVICE_NAME}-logs" >>.env.development

# Update Makefile
echo "Updating Makefile..."
cat >>Makefile <<EOF

.PHONY: build-$SERVICE_NAME
build-$SERVICE_NAME: ## Build $SERVICE_NAME
	@echo "Building $SERVICE_NAME..."
	@mkdir -p \$(BUILD_DIR)
	@cd services/$SERVICE_NAME && \$(GOBUILD) -o ../\$(BUILD_DIR)/$SERVICE_NAME ./cmd

.PHONY: run-$SERVICE_NAME
run-$SERVICE_NAME: ## Run $SERVICE_NAME
	@echo "Running $SERVICE_NAME..."
	@cd services/$SERVICE_NAME && \$(GO) run ./cmd

.PHONY: test-$SERVICE_NAME
test-$SERVICE_NAME: ## Run $SERVICE_NAME tests
	@echo "Running $SERVICE_NAME tests..."
	@cd services/$SERVICE_NAME && \$(GOTEST) ./...

.PHONY: air-$SERVICE_NAME
air-$SERVICE_NAME: ## Run $SERVICE_NAME with Air locally
	@echo "Starting $SERVICE_NAME with Air..."
	@cd services/$SERVICE_NAME && air
EOF

# Update main build and run targets
sed -i "/^build:/ s/##.*/ build-$SERVICE_NAME&/" Makefile
sed -i "/^run:/ s/##.*/ run-$SERVICE_NAME&/" Makefile

# Register service with API gateway
echo "Registering service with API gateway..."
# Find the service registration section in main.go
REGISTRATION_LINE=$(grep -n "serviceRegistry.RegisterService" api-gateway/cmd/main.go | tail -1 | cut -d: -f1)
if [ -n "$REGISTRATION_LINE" ]; then
    # Insert after the last registration
    sed -i "${REGISTRATION_LINE}a \\\\n\\t// Register $SERVICE_NAME\\n\\tserviceRegistry.RegisterService(\"$SERVICE_NAME\", \"http://$SERVICE_NAME:$PORT\")" api-gateway/cmd/main.go
else
    echo "Warning: Could not find service registration section in API gateway"
fi

# Create volume directories
echo "Creating volume directories..."
mkdir -p "docker/volumes/$SERVICE_NAME/tmp"
mkdir -p "docker/volumes/$SERVICE_NAME/logs"

# Database schema creation (optional)
if [ "$CREATE_DB_SCHEMA" = true ]; then
    echo "Creating database schema for $SERVICE_NAME..."

    # Add schema to .env.development
    SCHEMA_VAR="${SERVICE_NAME_UPPER}_SCHEMA"
    SCHEMA_VALUE=$(echo "$SERVICE_NAME" | sed 's/-/_/g')

    if ! grep -q "^${SCHEMA_VAR}=" .env.development; then
        echo "${SCHEMA_VAR}=${SCHEMA_VALUE}" >>.env.development
        echo "Added ${SCHEMA_VAR}=${SCHEMA_VALUE} to .env.development"
    fi

    # Actually create the schema in database
    echo "Creating schema $SCHEMA_VALUE in database..."
    docker-compose exec postgres psql -U postgres -d service_db -c "CREATE SCHEMA IF NOT EXISTS $SCHEMA_VALUE;" 2>/dev/null ||
        echo "Note: Database may not be running yet. Schema will be created when migrations run."

    if [ -f "services/$SERVICE_NAME/migrations/000001_initial.up.sql" ]; then
        echo "Database schema created. Run migrations with:"
        echo "  make db-migrate-init SERVICE_NAME=$SERVICE_NAME"
        echo "  make db-migrate-up SERVICE_NAME=$SERVICE_NAME"
        echo "   - db-migrate-init creates the migration tracking table"
        echo "   - db-migrate-up runs the actual migrations"
    fi
fi

# Update Promtail configuration
echo "Updating Promtail configuration..."
if [ -f "docker/promtail-config.yml" ]; then
    # Create temporary file for new job configuration
    JOB_CONFIG_FILE=$(mktemp)

    # Create Promtail job for the new service
    cat >"$JOB_CONFIG_FILE" <<EOF

  # $SERVICE_NAME logs
  - job_name: $SERVICE_NAME
    static_configs:
      - targets:
          - localhost
        labels:
          job: $SERVICE_NAME
          service: $SERVICE_NAME
          __path__: /var/log/$SERVICE_NAME/*.log
    pipeline_stages:
      - json:
          expressions:
            level: level
            timestamp: timestamp
            service: service
            request_id: request_id
            method: method
            path: path
            status: status
            duration_ms: duration_ms
      - labels:
          level:
          service:
          request_id:
          method:
          path:
          status:
      - timestamp:
          source: timestamp
          format: RFC3339
EOF

    # Append the job configuration to the end of scrape_configs
    cat "$JOB_CONFIG_FILE" >> docker/promtail-config.yml

    # Clean up temporary file
    rm -f "$JOB_CONFIG_FILE"

    echo "Added $SERVICE_NAME to Promtail configuration"
fi

# Update documentation
if [ "$UPDATE_DOCS" = true ]; then
    echo "Updating documentation..."
    if [ -f "docs/future_development_plan.md" ]; then
        # Add service to the planned features section
        sed -i "/## Planned Features/a - [x] Add $SERVICE_NAME service (completed)" docs/future_development_plan.md
    fi
fi

echo ""
echo "🎉 Service $SERVICE_NAME created successfully!"
echo ""
echo "📋 Next steps:"
echo "1. Review and customize the generated code in services/$SERVICE_NAME/"
echo "2. Update API gateway routes in api-gateway/cmd/main.go for $SERVICE_NAME"
if [ "$CREATE_DB_SCHEMA" = true ]; then
    echo "3. Run database migrations: make db-migrate-up SERVICE_NAME=$SERVICE_NAME"
    echo "   - This will create the migration tracking table and enable enhanced migration features"
fi
echo "4. Build the service: make build-$SERVICE_NAME"
echo "5. Start the service: make run-$SERVICE_NAME"
echo "6. Test the service: make test-$SERVICE_NAME"
echo ""
echo "🔧 Available commands:"
echo "  make build-$SERVICE_NAME     - Build the service"
echo "  make run-$SERVICE_NAME       - Run the service locally"
echo "  make air-$SERVICE_NAME       - Run with hot reload"
echo "  make test-$SERVICE_NAME      - Run service tests"
echo "  make up                     - Start all services with Docker"
echo "  make dev                    - Start development environment"
echo ""
echo "📁 Service structure:"
echo "  services/$SERVICE_NAME/"
echo "  ├── cmd/main.go"
echo "  ├── internal/"
echo "  │   ├── handlers/"
echo "  │   ├── models/"
echo "  │   ├── repository/"
echo "  │   └── services/"
echo "  ├── migrations/"
echo "  ├── config.yaml"
echo "  ├── Dockerfile"
echo "  └── README.md"
