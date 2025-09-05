#!/bin/bash

set -e

# Check arguments
if [ $# -ne 2 ]; then
    echo "Usage: $0 <service-name> <port>"
    echo "Example: $0 product-service 8082"
    exit 1
fi

SERVICE_NAME=$1
PORT=$2

# Validate service name (lowercase, hyphens allowed)
if [[ ! $SERVICE_NAME =~ ^[a-z-]+[a-z]$ ]]; then
    echo "Error: Service name must be lowercase and contain only letters and hyphens"
    exit 1
fi

# Validate port
if [[ ! $PORT =~ ^[0-9]+$ ]] || [ $PORT -lt 1024 ] || [ $PORT -gt 65535 ]; then
    echo "Error: Port must be a number between 1024 and 65535"
    exit 1
fi

SERVICE_DIR="services/$SERVICE_NAME"
TEMPLATE_DIR="templates/service-template"

echo "Creating new service: $SERVICE_NAME on port $PORT"

# Check if service already exists
if [ -d "$SERVICE_DIR" ]; then
    echo "Error: Service $SERVICE_NAME already exists"
    exit 1
fi

# Copy template
echo "Copying template..."
cp -r "$TEMPLATE_DIR" "$SERVICE_DIR"

# Replace variables in files
echo "Customizing service..."

# Function to replace variables in file
replace_vars() {
    local file=$1
    sed -i "s/{{SERVICE_NAME}}/$SERVICE_NAME/g" "$file"
    sed -i "s/{{PORT}}/$PORT/g" "$file"
}

# Find and replace in all files
find "$SERVICE_DIR" -type f -name "*.go" -o -name "*.yaml" -o -name "*.md" -o -name "*.sql" | while read -r file; do
    replace_vars "$file"
done

# Update docker-compose.yml
echo "Updating docker-compose.yml..."
cat >> docker/docker-compose.yml << EOF

  $SERVICE_NAME:
    build:
      context: ../services/$SERVICE_NAME
      dockerfile: Dockerfile
    container_name: service-boilerplate-$SERVICE_NAME
    ports:
      - "$PORT:$PORT"
    environment:
      - APP_ENV=production
      - SERVER_PORT=$PORT
      - DATABASE_HOST=postgres
      - DATABASE_PORT=5432
      - DATABASE_USER=postgres
      - DATABASE_PASSWORD=postgres
      - DATABASE_NAME=service_db
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - service-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:$PORT/health"]
      interval: 30s
      timeout: 10s
      retries: 3
EOF

# Update Makefile
echo "Updating Makefile..."
cat >> Makefile << EOF

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
EOF

# Update main build and run targets
sed -i "/^build:/ s/$/ build-$SERVICE_NAME/" Makefile
sed -i "/^run:/ s/$/ run-$SERVICE_NAME/" Makefile

# Register service with API gateway
echo "Registering service with API gateway..."
cat >> api-gateway/cmd/main.go << EOF
	// Register $SERVICE_NAME
	serviceRegistry.RegisterService("$SERVICE_NAME", "http://localhost:$PORT")
EOF

echo "Service $SERVICE_NAME created successfully!"
echo ""
echo "Next steps:"
echo "1. Update API gateway routes in api-gateway/cmd/main.go"
echo "2. Run 'make build-$SERVICE_NAME' to build the service"
echo "3. Run 'make run-$SERVICE_NAME' to start the service"
echo "4. Update docker-compose.yml health checks if needed"