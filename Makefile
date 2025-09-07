# Service Boilerplate Makefile

# Project variables
PROJECT_NAME := service-boilerplate
API_GATEWAY_DIR := api-gateway
USER_SERVICE_DIR := services/user-service
BUILD_DIR := build
DOCKER_COMPOSE_FILE := docker/docker-compose.yml

# Network variables (simplified for compatibility)
NETWORK_NAME := service-boilerplate-network
NETWORK_DRIVER := bridge

# Go variables
GO := go
GOMOD := $(GO) mod
GOBUILD := $(GO) build
GOTEST := $(GO) test
GOCLEAN := $(GO) clean
GOFMT := $(GO) fmt
GOVET := $(GO) vet

# Docker variables
DOCKER := docker
DOCKER_COMPOSE := docker-compose

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ''
	@echo 'Cleaning Commands:'
	@echo '  clean              - Clean basic Go build artifacts'
	@echo '  clean-all          - Complete clean for fresh start'
	@echo '  clean-go           - Clean Go artifacts and cache'
	@echo '  clean-docker       - Clean Docker containers/images'
	@echo '  clean-volumes      - Clean Docker volumes'
	@echo '  clean-logs         - Clean log files'
	@echo '  clean-cache        - Clean caches and temp files'
	@echo '  clean-test         - Clean test artifacts'
	@echo '  fresh-start        - Complete reset and setup'
	@echo '  clean-all-confirm  - Clean all with confirmation'
	@echo ''
	@echo 'Network Commands:'
	@echo '  network-create     - Create custom Docker network'
	@echo '  network-inspect    - Inspect Docker network'
	@echo '  network-ls         - List Docker networks'
	@echo '  network-clean      - Clean up unused networks'
	@echo '  network-remove     - Remove custom network'

.PHONY: setup
setup: ## Initialize project (download deps, setup tools)
	@echo "Setting up project..."
	@$(GOMOD) download
	@$(GOMOD) tidy

.PHONY: build
build: build-gateway build-user-service ## Build all services

.PHONY: build-gateway
build-gateway: ## Build API Gateway
	@echo "Building API Gateway..."
	@mkdir -p $(BUILD_DIR)
	@cd $(API_GATEWAY_DIR) && $(GOBUILD) -o ../$(BUILD_DIR)/api-gateway ./cmd

.PHONY: build-user-service
build-user-service: ## Build User Service
	@echo "Building User Service..."
	@mkdir -p $(BUILD_DIR)
	@cd $(USER_SERVICE_DIR) && $(GOBUILD) -o ../$(BUILD_DIR)/user-service ./cmd

.PHONY: up
up: ## Start all services with Docker (PRIMARY)
	@echo "Starting services with Docker..."
	@$(DOCKER_COMPOSE) --env-file .env up -d
	@echo "Services started! Use 'make logs' to view logs."

.PHONY: down
down: ## Stop all services
	@echo "Stopping services..."
	@$(DOCKER_COMPOSE) --env-file .env down
	@echo "Services stopped."

.PHONY: dev
dev: ## Start services in development mode with hot reload
	@echo "Starting development environment with hot reload..."
	@$(DOCKER_COMPOSE) --env-file .env -f $(DOCKER_COMPOSE_FILE) -f docker/docker-compose.override.yml up

.PHONY: dev-build
dev-build: ## Build development images with Air
	@echo "Building development images..."
	@$(DOCKER_COMPOSE) --env-file .env -f $(DOCKER_COMPOSE_FILE) -f docker/docker-compose.override.yml build

.PHONY: air-gateway
air-gateway: ## Run API Gateway with Air locally
	@echo "Starting API Gateway with Air..."
	@cd $(API_GATEWAY_DIR) && air

.PHONY: air-user-service
air-user-service: ## Run User Service with Air locally
	@echo "Starting User Service with Air..."
	@cd $(USER_SERVICE_DIR) && air

.PHONY: logs
logs: ## Show service logs
	@$(DOCKER_COMPOSE) --env-file .env logs -f

.PHONY: run-local
run-local: ## Run all services locally (SECONDARY)
	@echo "Starting all services locally..."
	@cd $(API_GATEWAY_DIR) && $(GO) run ./cmd &
	@sleep 2
	@cd $(USER_SERVICE_DIR) && $(GO) run ./cmd &
	@echo "All services started locally. Use 'make stop-local' to stop them."

.PHONY: run-gateway-local
run-gateway-local: ## Run API Gateway locally
	@echo "Running API Gateway locally..."
	@cd $(API_GATEWAY_DIR) && $(GO) run ./cmd

.PHONY: run-user-service-local
run-user-service-local: ## Run User Service locally
	@echo "Running User Service locally..."
	@cd $(USER_SERVICE_DIR) && $(GO) run ./cmd

.PHONY: stop-local
stop-local: ## Stop all locally running services
	@echo "Stopping local services..."
	@pkill -f "go run ./cmd" || true
	@echo "Local services stopped."

.PHONY: test
test: ## Run all tests
	@echo "Running tests..."
	@$(GOTEST) ./...

.PHONY: test-gateway
test-gateway: ## Run API Gateway tests
	@echo "Running API Gateway tests..."
	@cd $(API_GATEWAY_DIR) && $(GOTEST) ./...

.PHONY: test-user-service
test-user-service: ## Run User Service tests
	@echo "Running User Service tests..."
	@cd $(USER_SERVICE_DIR) && $(GOTEST) ./...

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)

.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting code..."
	@$(GOFMT) ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	@$(GOVET) ./...

.PHONY: lint
lint: ## Run golangci-lint
	@echo "Running linter..."
	@golangci-lint run

.PHONY: check
check: fmt vet lint ## Run fmt, vet, and lint

.PHONY: docker-build
docker-build: ## Build all Docker images
	@echo "Building Docker images..."
	@$(DOCKER_COMPOSE) --env-file .env -f $(DOCKER_COMPOSE_FILE) build

.PHONY: docker-logs
docker-logs: ## Show Docker container logs (legacy)
	@echo "Showing container logs..."
	@$(DOCKER_COMPOSE) --env-file .env -f $(DOCKER_COMPOSE_FILE) logs -f

.PHONY: migrate-up
migrate-up: ## Run database migrations
	@echo "Running database migrations..."
	@migrate -path services/user-service/migrations -database $(DATABASE_URL) up

.PHONY: migrate-down
migrate-down: ## Rollback database migrations
	@echo "Rolling back database migrations..."
	@migrate -path services/user-service/migrations -database $(DATABASE_URL) down 1

.PHONY: db-reset
db-reset: migrate-down migrate-up ## Reset database (down + up)

.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GOMOD) download

.PHONY: tidy
tidy: ## Clean up go.mod
	@echo "Tidying go.mod..."
	@$(GOMOD) tidy

.PHONY: create-service
create-service: ## Create new service (usage: make create-service SERVICE_NAME=name PORT=8082)
	@echo "Creating new service: $(SERVICE_NAME)"
	@./scripts/create-service.sh $(SERVICE_NAME) $(PORT)

# Cleaning Commands
.PHONY: clean-all
clean-all: clean-go clean-docker clean-volumes clean-logs clean-cache clean-test ## Complete clean for fresh start
	@echo "✅ Complete clean finished! All artifacts removed."

.PHONY: clean-go
clean-go: ## Clean Go build artifacts and cache
	@echo "🧹 Cleaning Go artifacts..."
	@$(GOCLEAN) -r
	@rm -rf $(BUILD_DIR)
	@rm -rf $(API_GATEWAY_DIR)/tmp
	@rm -rf $(USER_SERVICE_DIR)/tmp
	@find . -name "*.test" -type f -delete 2>/dev/null || true
	@find . -name "*.out" -type f -delete 2>/dev/null || true
	@find . -name "coverage.*" -type f -delete 2>/dev/null || true
	@echo "✅ Go artifacts cleaned"

.PHONY: clean-docker
clean-docker: ## Clean Docker containers, images, and networks
	@echo "🐳 Cleaning Docker artifacts..."
	@docker-compose --env-file .env down --volumes --remove-orphans 2>/dev/null || true
	@docker system prune -f --volumes 2>/dev/null || true
	@docker volume prune -f 2>/dev/null || true
	@docker network prune -f 2>/dev/null || true
	@echo "✅ Docker artifacts cleaned"

.PHONY: clean-volumes
clean-volumes: ## Clean Docker volumes and persistent data
	@echo "💾 Cleaning Docker volumes..."
	@docker-compose --env-file .env down -v 2>/dev/null || true
	@docker volume rm $$(docker volume ls -q | grep -E "(postgres_data|api_gateway|user_service)" 2>/dev/null) 2>/dev/null || true
	@echo "🔧 Handling PostgreSQL volume permissions..."
	@if [ -d "docker/volumes/postgres_data" ]; then \
		echo "   Using Docker to clean PostgreSQL data..."; \
		docker run --rm -v $(PWD)/docker/volumes/postgres_data:/var/lib/postgresql/data alpine sh -c "rm -rf /var/lib/postgresql/data/* 2>/dev/null || true" 2>/dev/null || true; \
	fi
	@rm -rf docker/volumes/api-gateway/ 2>/dev/null || true
	@rm -rf docker/volumes/user-service/ 2>/dev/null || true
	@rm -rf docker/volumes/ 2>/dev/null || true
	@echo "✅ Docker volumes cleaned"

.PHONY: clean-logs
clean-logs: ## Clean log files
	@echo "📝 Cleaning log files..."
	@find . -name "*.log" -type f -delete 2>/dev/null || true
	@find . -name "build-errors.log" -type f -delete 2>/dev/null || true
	@rm -rf logs/ 2>/dev/null || true
	@echo "✅ Log files cleaned"

.PHONY: clean-cache
clean-cache: ## Clean Go caches and temporary files
	@echo "🗂️  Cleaning caches and temporary files..."
	@go clean -cache 2>/dev/null || true
	@go clean -modcache 2>/dev/null || true
	@find . -name ".DS_Store" -type f -delete 2>/dev/null || true
	@find . -name "Thumbs.db" -type f -delete 2>/dev/null || true
	@find . -name "*.bak" -type f -delete 2>/dev/null || true
	@find . -name "*.old" -type f -delete 2>/dev/null || true
	@find . -name "*.tmp" -type f -delete 2>/dev/null || true
	@echo "✅ Caches and temporary files cleaned"

.PHONY: clean-test
clean-test: ## Clean test artifacts
	@echo "🧪 Cleaning test artifacts..."
	@find . -name "*.cover" -type f -delete 2>/dev/null || true
	@find . -name "*.coverprofile" -type f -delete 2>/dev/null || true
	@find . -name "coverage.txt" -type f -delete 2>/dev/null || true
	@find . -name "coverage.html" -type f -delete 2>/dev/null || true
	@rm -rf test-results/ 2>/dev/null || true
	@echo "✅ Test artifacts cleaned"

.PHONY: fresh-start
fresh-start: clean-all setup ## Complete reset and setup
	@echo "🔄 Fresh start complete! Ready for development."

.PHONY: clean-all-confirm
clean-all-confirm: ## Clean all with confirmation prompt
	@echo "⚠️  This will remove ALL build artifacts, Docker volumes, and caches!"
	@echo "This includes database data and cannot be undone."
	@echo ""
	@read -p "Are you sure you want to proceed? (y/N): " confirm && \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		$(MAKE) clean-all; \
		echo "✅ Clean operation completed successfully."; \
	else \
		echo "❌ Clean operation cancelled."; \
	fi

# Network Management Commands
.PHONY: network-create
network-create: ## Create custom Docker network
	@echo "🌐 Creating service network..."
	@docker network create \
		--driver $(NETWORK_DRIVER) \
		--subnet $(NETWORK_SUBNET) \
		--gateway $(NETWORK_GATEWAY) \
		--label com.service-boilerplate.network=backend \
		--label com.service-boilerplate.project=service-boilerplate \
		$(NETWORK_NAME) 2>/dev/null || echo "Network $(NETWORK_NAME) already exists"

.PHONY: network-inspect
network-inspect: ## Inspect Docker network
	@echo "🔍 Inspecting service network..."
	@docker network inspect $(NETWORK_NAME) || echo "Network $(NETWORK_NAME) not found"

.PHONY: network-ls
network-ls: ## List Docker networks
	@echo "📋 Docker networks:"
	@docker network ls

.PHONY: network-clean
network-clean: ## Clean up Docker networks
	@echo "🧹 Cleaning up unused networks..."
	@docker network prune -f
	@echo "✅ Unused networks cleaned"

.PHONY: network-remove
network-remove: ## Remove custom network
	@echo "🗑️  Removing service network..."
	@docker network rm $(NETWORK_NAME) 2>/dev/null || echo "Network $(NETWORK_NAME) not found or in use"

.PHONY: help-network
help-network: ## Show network commands
	@echo "🌐 Network Commands:"
	@echo "  network-create     - Create custom Docker network"
	@echo "  network-inspect    - Inspect Docker network"
	@echo "  network-ls         - List Docker networks"
	@echo "  network-clean      - Clean up unused networks"
	@echo "  network-remove     - Remove custom network"

.PHONY: help-clean
help-clean: ## Show cleaning commands
	@echo "🧹 Cleaning Commands:"
	@echo "  clean              - Clean basic Go build artifacts"
	@echo "  clean-all          - Complete clean for fresh start"
	@echo "  clean-go           - Clean Go artifacts and cache"
	@echo "  clean-docker       - Clean Docker containers/images"
	@echo "  clean-volumes      - Clean Docker volumes"
	@echo "  clean-logs         - Clean log files"
	@echo "  clean-cache        - Clean caches and temp files"
	@echo "  clean-test         - Clean test artifacts"
	@echo "  fresh-start        - Complete reset and setup"
	@echo "  clean-all-confirm  - Clean all with confirmation"