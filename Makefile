# Service Boilerplate Makefile
MAKEFLAGS += --no-print-directory

# Project variables
API_GATEWAY_DIR := api-gateway
USER_SERVICE_DIR := services/user-service
CLI_DIR := cli
BUILD_DIR := build
DOCKER_COMPOSE_FILE := docker/docker-compose.yml
DOCKER_COMPOSE_OVERRIDE_FILE := docker/docker-compose.override.yml

# Network variables (simplified for compatibility)
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
DOCKER_COMPOSE := docker compose

# Environment-specific configuration
APP_ENV ?= development
ENV_FILE := .env
ifneq ("$(wildcard .env.$(APP_ENV))","")
    ENV_FILE := .env.$(APP_ENV)
endif

# Load environment variables from .env file
define load_env
$(foreach var, $(shell cat $(ENV_FILE) | grep -v '^#' | grep -v '^$$' | grep -v '^SERVICE_NAME=' | cut -d'=' -f1), $(eval $(var) := $(shell grep '^$(var)=' $(ENV_FILE) | cut -d'=' -f2)))
endef
$(call load_env)

# Service URL variables (can be overridden via environment)
AUTH_SERVICE_URL ?= http://auth-service:8083
USER_SERVICE_URL ?= http://user-service:8081
API_GATEWAY_URL ?= http://localhost:8080

# Base image variables for smart cleanup
POSTGRES_IMAGE := postgres:15-alpine
GOLANG_BUILD_IMAGE := golang:1.23-alpine
ALPINE_RUNTIME_IMAGE := alpine:latest
MIGRATION_IMAGE := migrate/migrate:latest

# Docker cleanup configuration
DOCKER_CLEANUP_MODE ?= smart

# Dynamic service variable loading from environment-specific .env file
# Extract all service containers, images, and volumes from .env
SERVICE_CONTAINERS := $(shell grep "_CONTAINER=" $(ENV_FILE) | grep -v "POSTGRES_CONTAINER" | grep -v -E "(LOKI|PROMTAIL|GRAFANA)_CONTAINER" | cut -d'=' -f2)
SERVICE_IMAGES := $(shell grep "_IMAGE=" $(ENV_FILE) | grep -v "POSTGRES_IMAGE\|MIGRATION_IMAGE" | grep -v -E "(LOKI|PROMTAIL|GRAFANA)_IMAGE" | cut -d'=' -f2)
SERVICE_VOLUMES := $(shell grep "_VOLUME=" $(ENV_FILE) | grep -v "POSTGRES_VOLUME\|MIGRATION_TMP_VOLUME" | grep -v -E "(LOKI_DATA|PROMTAIL_POSITIONS|GRAFANA_DATA)_VOLUME" | cut -d'=' -f2)

# Extract monitoring containers, images, and volumes from .env
MONITORING_CONTAINERS := $(shell grep "_CONTAINER=" $(ENV_FILE) | grep -E "(LOKI|PROMTAIL|GRAFANA)_CONTAINER" | cut -d'=' -f2)
MONITORING_IMAGES := $(shell grep "_IMAGE=" $(ENV_FILE) | grep -E "(LOKI|PROMTAIL|GRAFANA)_IMAGE" | cut -d'=' -f2)
MONITORING_VOLUMES := $(shell grep "_VOLUME=" $(ENV_FILE) | grep -E "(LOKI_DATA|PROMTAIL_POSITIONS|GRAFANA_DATA|JAEGER_DATA)_VOLUME" | cut -d'=' -f2)
MONITORING_VOLUME_DIRS := grafana loki promtail jaeger

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo '🚀 QUICK START:'
	@echo '  make dev-bootstrap - 🛠️  Bootstrap DEVELOPMENT environment (auto DB setup, hot reload)'
	@echo '  make dev           - 🛠️  Start DEVELOPMENT environment (hot reload, debug logs)'
	@echo '  make prod          - 🚀 Start PRODUCTION environment (pre-built images)'

	@echo ''
	@echo '📋 Available Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ''
	@echo '💡 DEVELOPMENT vs PRODUCTION:'
	@echo '  • make dev  : Hot reload, volume mounts, debug logging, development tools'
	@echo '  • make prod : Pre-built optimized images, production settings'

	@echo ''
	@echo 'CLI Commands:'
	@echo '  build-cli          - Build CLI utility'
	@echo '  build-all          - Build all services and CLI'
	@echo '  run-cli            - Build and run CLI utility'
	@echo '  test-cli           - Run CLI tests'
	@echo '  test-all           - Run all tests (services + CLI)'
	@echo '  clean-cli          - Clean CLI build artifacts'
	@echo ''
	@echo 'Health & Monitoring:'
	@echo '  health             - Comprehensive health check of all services'
	@echo '  health-services    - Check HTTP health endpoints'
	@echo '  health-containers  - Check Docker container status'
	@echo '  health-database    - Check database connectivity'
	@echo '  health-network     - Check Docker network status'
	@echo '  health-volumes     - Check volume mount status'
	@echo ''
	@echo 'Database Commands:'
	@echo '  db-connect         - Connect to database shell'
	@echo '  db-status          - Show database status and connections'
	@echo '  db-health          - Check database health and connectivity'
	@echo '  db-create          - Create database if it does not exist'
	@echo '  db-drop            - Drop database (with confirmation)'
	@echo '  db-recreate        - Recreate database from scratch'
	@echo ''
	@echo 'Migration Management (Orchestrator):'
	@echo '  db-migrate-init    - Initialize migration tracking for service'
	@echo '  db-migrate         - Run migrations for all services (or specific with SERVICE_NAME=)'
	@echo '  db-migrate-up      - Run migrations up for specific service'
	@echo '  db-migrate-down    - Run migrations down for specific service'
	@echo '  db-migrate-status  - Show migration status for specific service'
	@echo '  db-migrate-list    - List migration executions for specific service'
	@echo '  db-migrate-validate - Validate migrations for specific service'
	@echo '  db-migrate-create NAME= - Create migration file'
	@echo '  db-migrate-generate NAME= TYPE= - Generate migration with templates'
	@echo '  db-migrate-file-list  - List migration files'
	@echo ''
	@echo 'Data Management:'
	@echo '  db-seed            - Seed database with test data'
	@echo '  db-seed-enhanced ENV= - Environment-specific seeding'
	@echo '  db-dump            - Create database dump'
	@echo '  db-restore FILE=   - Restore database from dump'
	@echo '  db-clean           - Clean all data from tables'
	@echo '  db-schema          - Show database schema'
	@echo '  db-tables          - List all tables and structure'
	@echo '  db-counts          - Show row counts for all tables'
	@echo '  db-setup           - Complete database setup'
	@echo '  db-reset-dev       - Reset database for development'
	@echo '  db-fresh           - Complete reset with volume cleanup'
	@echo ''
	@echo 'Cleaning Commands:'
	@echo '  clean              - Clean basic Go build artifacts'
	@echo '  clean-all          - Complete clean for fresh start'
	@echo '  clean-go           - Clean Go artifacts and cache'
	@echo '  clean-cli          - Clean CLI build artifacts'
	@echo '  clean-docker       - Clean project Docker artifacts'
	@echo '  clean-volumes      - Clean Docker volumes'
	@echo '  clean-logs         - Clean log files'
	@echo '  clean-cache        - Clean caches and temp files'
	@echo '  clean-test         - Clean test artifacts'
	@echo '  fresh-start        - Complete reset and setup'
	@echo '  clean-all-confirm  - Clean all with confirmation'
	@echo ''
	@echo 'Docker Management:'
	@echo '  docker-reset           - Complete project Docker environment reset'
	@echo '  docker-reset-confirm   - Reset with confirmation prompt'
	@echo '  docker-recreate        - Recreate project Docker environment'
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
build: build-gateway build-user-service build-auth-service build-objects-service ## Build all services

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

.PHONY: build-cli
build-cli: ## Build CLI utility
	@echo "Building CLI utility..."
	@mkdir -p $(BUILD_DIR)
	@cd $(CLI_DIR) && $(GOBUILD) -o ../$(BUILD_DIR)/boilerplate-cli ./main.go
	@echo "✅ CLI built successfully: $(BUILD_DIR)/boilerplate-cli"

.PHONY: build-all
build-all: build build-cli ## Build all services and CLI
	@echo "✅ All components built successfully"

.PHONY: check-prod-safety
check-prod-safety:
	@if [ -f ".git" ] && [ -d "api-gateway" ] && [ "$(APP_ENV)" != "production" ] && [ "$(FORCE_PROD)" != "true" ]; then \
		echo "⚠️  WARNING: You appear to be in a DEVELOPMENT environment."; \
		echo "   Production mode uses pre-built images without hot reload."; \
		echo "   For development with hot reload, use: make dev"; \
		echo ""; \
		read -p "Continue with production mode? (y/N): " confirm; \
		if [ "$$confirm" != "y" ] && [ "$$confirm" != "Y" ]; then \
			echo "❌ Production start cancelled. Use 'make dev' for development."; \
			exit 1; \
		fi; \
	fi

.PHONY: prod
prod: check-prod-safety ## 🚀 Start services in PRODUCTION mode (pre-built images, no hot reload)
	@echo "🏭 Starting PRODUCTION environment..."
	@echo "⚠️  WARNING: This uses pre-built production images without hot reload!"
	@echo "   For development/debugging with hot reload, use: make dev"
	@echo ""
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) up -d
	@echo "✅ Production services started! Use 'make logs' to view logs."



.PHONY: smart-start
smart-start: ## 🧠 Smart start - automatically detects environment and uses appropriate mode
	@if [ "$(APP_ENV)" = "production" ] || [ "$(FORCE_PROD)" = "true" ]; then \
		echo "🏭 Detected PRODUCTION environment - starting with optimized images..."; \
		$(MAKE) prod; \
	else \
		echo "🛠️  Detected DEVELOPMENT environment - starting with hot reload..."; \
		$(MAKE) dev; \
	fi

.PHONY: start
start: build-prod prod ## ⚠️  DEPRECATED: Use 'make smart-start' or specify 'make prod'/'make dev'
	@echo "✅ Services built and started successfully"

.PHONY: down
down: ## Stop all services
	@echo "Stopping services..."
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) down
	@echo "Services stopped."

.PHONY: dev-bootstrap
dev-bootstrap: create-volumes-dirs build-migrate-wrapper  ## Bootstrap development environment with database setup
	@echo "🚀 Bootstrapping development environment..."
	@echo "📁 Creating volume directories..."
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) up postgres -d
	@echo "⏳ Waiting for postgres to be ready..."
	@sleep 5
	$(MAKE) db-setup
	@echo "✅ Database ready, starting all services..."
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) -f $(DOCKER_COMPOSE_OVERRIDE_FILE) up

.PHONY: dev
dev: create-volumes-dirs  ## Start services in development mode with hot reload (blocks - use dev-detached for automation)
	@echo "Starting development environment with hot reload..."
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) -f $(DOCKER_COMPOSE_OVERRIDE_FILE) up

.PHONY: dev-detached
dev-detached: create-volumes-dirs  ## Start services in development mode (detached, for automation)
	@echo "Starting development environment (detached)..."
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) -f $(DOCKER_COMPOSE_OVERRIDE_FILE) up -d

.PHONY: build-dev
build-dev: ## Build development images with Air
	@echo "Building development images..."
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) -f $(DOCKER_COMPOSE_OVERRIDE_FILE) build

.PHONY: status
status: ## Show current environment status and running services
	@echo "📊 Environment Status:"
	@echo "  APP_ENV: $(APP_ENV)"
	@echo "  Services running:"
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep $(DOCKER_PROJECT_PREFIX) || echo "    No $(DOCKER_PROJECT_PREFIX) containers running"
	@echo ""
	@echo "💡 Quick Commands:"
	@echo "  make dev    - Start development environment (hot reload)"
	@echo "  make prod   - Start production environment (pre-built)"
	@echo "  make down   - Stop all services"
	@echo "  make logs   - View service logs"

.PHONY: logs
logs: ## Show service logs
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) logs -f









.PHONY: run-cli
run-cli: build-cli ## Build and run CLI utility
	@echo "Running CLI utility..."
	@./$(BUILD_DIR)/boilerplate-cli



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

.PHONY: test-cli
test-cli: ## Run CLI tests
	@echo "Running CLI tests..."
	@cd $(CLI_DIR) && $(GOTEST) ./...

.PHONY: test-all
test-all: test test-cli ## Run all tests (services + CLI)
	@echo "✅ All tests completed"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)

.PHONY: clean-cli
clean-cli: ## Clean CLI build artifacts
	@echo "Cleaning CLI artifacts..."
	@rm -f $(BUILD_DIR)/boilerplate-cli
	@rm -f $(CLI_DIR)/boilerplate-cli
	@rm -f $(CLI_DIR)/main
	@find $(CLI_DIR) -name "*.test" -type f -delete 2>/dev/null || true
	@find $(CLI_DIR) -name "*.out" -type f -delete 2>/dev/null || true
	@echo "✅ CLI artifacts cleaned"

.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting code..."
	@$(GOFMT) ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	@$(GOVET) ./...

.PHONY: lint
lint: ## Run golangci-lint on all Go modules
	@echo "Running linter on all Go modules..."
	@echo "Linting: api-gateway"
	@cd api-gateway && golangci-lint run ./... 2>&1 | grep -v "typechecking error" || true
	@echo "Linting: services (auth-service, user-service, objects-service)"
	@cd services && golangci-lint run ./... 2>&1 | grep -v "typechecking error" || true
	@echo "Linting: cli"
	@cd cli && golangci-lint run ./... 2>&1 | grep -v "typechecking error" || true
	@echo "Linting: common"
	@cd common && golangci-lint run ./... 2>&1 | grep -v "typechecking error" || true
	@echo "Linting: migration-orchestrator"
	@cd migration-orchestrator && golangci-lint run ./... 2>&1 | grep -v "typechecking error" || true

.PHONY: check
check: fmt vet lint ## Run fmt, vet, and lint

.PHONY: build-prod
build-prod: ## Build production Docker images
	@echo "Building production Docker images..."
	@$(DOCKER_COMPOSE) --env-file .env -f $(DOCKER_COMPOSE_FILE) build

.PHONY: docker-logs
docker-logs: ## Show Docker container logs (legacy)
	@echo "Showing container logs..."
	@$(DOCKER_COMPOSE) --env-file .env -f $(DOCKER_COMPOSE_FILE) logs -f

.PHONY: db-reset
db-reset: db-rollback db-migrate ## Reset database (down + up) - Updated to use new targets

# ============================================================================
# 🗄️  DATABASE MANAGEMENT TARGETS
# ============================================================================

# Database variables (loaded from .env file at runtime)
DATABASE_USER ?= postgres
DATABASE_PASSWORD ?= postgres
DATABASE_HOST ?= postgres
DATABASE_PORT ?= 5432
DATABASE_NAME ?= service_db
DATABASE_SSL_MODE ?= disable
# SERVICE_NAME defaults to empty (run all services) or can be set to specific service
MIGRATION_IMAGE ?= migrate/migrate:latest
ORCHESTRATOR_IMAGE ?= migrate-wrapper:latest

# Auto-detect all services
SERVICES := $(shell ls services/ | grep -E '.*-service$$' | sort)

# Auto-detect services with migrations (exclude empty directories)
SERVICES_WITH_MIGRATIONS := $(shell find services -name "migrations" -type d -exec test -f {}/environments.json \; -print 2>/dev/null | sed 's|/migrations||' | sed 's|services/||' | sort)

# Fixed migration order: auth-service must run first (creates roles/permissions that user-service depends on)
SERVICE_MIGRATION_ORDER := auth-service user-service objects-service
POSTGRES_NAME ?= postgres

# Database URL construction for targets
DATABASE_URL := postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DATABASE_HOST):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=$(DATABASE_SSL_MODE)

## Database Connection & Access
.PHONY: db-connect
db-connect: ## Connect to database shell
	@echo "🔌 Connecting to database..."
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) exec postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME)

.PHONY: db-status
db-status: ## Show database status and connections
	@echo "📊 Database Status:"
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) exec postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) -c "SELECT version();" 2>/dev/null || echo "❌ Database not accessible"
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) exec postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) -c "SELECT count(*) as active_connections FROM pg_stat_activity;" 2>/dev/null || echo "❌ Cannot query connections"

.PHONY: db-health
db-health: ## Check database health and connectivity
	@echo "🏥 Database Health Check:"
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) exec postgres pg_isready -U $(DATABASE_USER) -d $(DATABASE_NAME) -h $(DATABASE_HOST) -p $(DATABASE_PORT)
	@if [ $$? -eq 0 ]; then \
		echo "✅ Database is healthy and accepting connections"; \
	else \
		echo "❌ Database health check failed"; \
	fi

## Database Management
.PHONY: db-create
db-create: ## Create database if it doesn't exist
	@echo "🆕 Creating database $(DATABASE_NAME)..."
	@$(DOCKER_COMPOSE) --env-file .env -f $(DOCKER_COMPOSE_FILE) exec postgres psql -U postgres -c "CREATE DATABASE $(DATABASE_NAME);" 2>/dev/null || echo "ℹ️  Database $(DATABASE_NAME) already exists or creation failed"

.PHONY: db-drop
db-drop: ## Drop database (with confirmation)
	@echo "💥 WARNING: This will drop database $(DATABASE_NAME) and ALL its data!"
	@echo "This action cannot be undone."
	@echo ""
	@read -p "Are you sure you want to drop the database? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		docker compose --env-file .env -f $(DOCKER_COMPOSE_FILE) exec postgres psql -U postgres -c "DROP DATABASE IF EXISTS $(DATABASE_NAME);"; \
		echo "✅ Database $(DATABASE_NAME) dropped successfully"; \
	else \
		echo "❌ Database drop cancelled"; \
	fi

.PHONY: db-recreate
db-recreate: db-drop db-create ## Recreate database from scratch
	@echo "🔄 Database $(DATABASE_NAME) recreated successfully"

## Migration Management (Docker-Based)
# Service-specific migration table name (e.g., user_service_schema_migrations)
.PHONY: db-migrate-init
db-migrate-init: ## Initialize migration tracking for service using orchestrator
	@echo "🔧 Initializing migration tracking for $(SERVICE_NAME) with orchestrator..."
	@if [ -z "$(SERVICE_NAME)" ]; then \
		echo "❌ Error: Please specify SERVICE_NAME (e.g., make db-migrate-init SERVICE_NAME=user-service)"; \
		exit 1; \
	fi
	@docker run --rm --network $(NETWORK_NAME) \
		--env-file $(ENV_FILE) \
		-e DB_HOST=$(POSTGRES_NAME) \
		-e DB_PORT=$(DATABASE_PORT) \
		-e DB_USER=$(DATABASE_USER) \
		-e DB_PASSWORD=$(DATABASE_PASSWORD) \
		-e DB_NAME=$(DATABASE_NAME) \
		-e DB_SSL_MODE=$(DATABASE_SSL_MODE) \
		-v $(PWD)/services:/services \
		$(ORCHESTRATOR_IMAGE) \
		init $(SERVICE_NAME)

.PHONY: db-migrate-up
db-migrate-up: ## Run migrations up using orchestrator
	@echo "📈 Running migrations up for $(SERVICE_NAME) with orchestrator..."
	@if [ -z "$(SERVICE_NAME)" ]; then \
		echo "❌ Error: Please specify SERVICE_NAME (e.g., make db-migrate-up SERVICE_NAME=user-service)"; \
		exit 1; \
	fi
	@docker run --rm --network $(NETWORK_NAME) \
		--env-file $(ENV_FILE) \
		-e DB_HOST=$(POSTGRES_NAME) \
		-e DB_PORT=$(DATABASE_PORT) \
		-e DB_USER=$(DATABASE_USER) \
		-e DB_PASSWORD=$(DATABASE_PASSWORD) \
		-e DB_NAME=$(DATABASE_NAME) \
		-e DB_SSL_MODE=$(DATABASE_SSL_MODE) \
		-v $(PWD)/services:/services \
		-e APP_ENV="${ENV:?APP_ENV must be set via ENV=staging or APP_ENV=staging}" \
		$(ORCHESTRATOR_IMAGE) \
		up $(SERVICE_NAME) --env "${ENV:?APP_ENV must be set via ENV=staging or APP_ENV=staging}"

.PHONY: db-migrate-down
db-migrate-down: ## Run migrations down using orchestrator
	@echo "⏪ Running migrations down for $(SERVICE_NAME) with orchestrator..."
	@if [ -z "$(SERVICE_NAME)" ]; then \
		echo "❌ Error: Please specify SERVICE_NAME (e.g., make db-migrate-down SERVICE_NAME=user-service)"; \
		exit 1; \
	fi
	@docker run --rm --network $(NETWORK_NAME) \
		--env-file $(ENV_FILE) \
		-e DB_HOST=$(POSTGRES_NAME) \
		-e DB_PORT=$(DATABASE_PORT) \
		-e DB_USER=$(DATABASE_USER) \
		-e DB_PASSWORD=$(DATABASE_PASSWORD) \
		-e DB_NAME=$(DATABASE_NAME) \
		-e DB_SSL_MODE=$(DATABASE_SSL_MODE) \
		-v $(PWD)/services:/services \
		$(ORCHESTRATOR_IMAGE) \
		down $(SERVICE_NAME) 1 --env $(APP_ENV)

.PHONY: db-migrate-status
db-migrate-status: ## Show migration status using orchestrator
	@echo "📊 Migration status for $(SERVICE_NAME) with orchestrator..."
	@if [ -z "$(SERVICE_NAME)" ]; then \
		echo "❌ Error: Please specify SERVICE_NAME (e.g., make db-migrate-status SERVICE_NAME=user-service)"; \
		exit 1; \
	fi
	@docker run --rm --network $(NETWORK_NAME) \
		--env-file $(ENV_FILE) \
		-e DB_HOST=$(POSTGRES_NAME) \
		-e DB_PORT=$(DATABASE_PORT) \
		-e DB_USER=$(DATABASE_USER) \
		-e DB_PASSWORD=$(DATABASE_PASSWORD) \
		-e DB_NAME=$(DATABASE_NAME) \
		-e DB_SSL_MODE=$(DATABASE_SSL_MODE) \
		-v $(PWD)/services:/services \
		$(ORCHESTRATOR_IMAGE) \
		status $(SERVICE_NAME) --env $(APP_ENV)

.PHONY: db-migrate-list
db-migrate-list: ## List all migrations using orchestrator
	@echo "📋 Migration list for $(SERVICE_NAME) with orchestrator..."
	@if [ -z "$(SERVICE_NAME)" ]; then \
		echo "❌ Error: Please specify SERVICE_NAME (e.g., make db-migrate-list SERVICE_NAME=user-service)"; \
		exit 1; \
	fi
	@docker run --rm --network $(NETWORK_NAME) \
		--env-file $(ENV_FILE) \
		-e DB_HOST=$(POSTGRES_NAME) \
		-e DB_PORT=$(DATABASE_PORT) \
		-e DB_USER=$(DATABASE_USER) \
		-e DB_PASSWORD=$(DATABASE_PASSWORD) \
		-e DB_NAME=$(DATABASE_NAME) \
		-e DB_SSL_MODE=$(DATABASE_SSL_MODE) \
		-v $(PWD)/services:/services \
		$(ORCHESTRATOR_IMAGE) \
		list $(SERVICE_NAME) --env $(APP_ENV)

.PHONY: db-migrate-validate
db-migrate-validate: ## Validate migrations using orchestrator
	@echo "✅ Validating migrations for $(SERVICE_NAME) with orchestrator..."
	@if [ -z "$(SERVICE_NAME)" ]; then \
		echo "❌ Error: Please specify SERVICE_NAME (e.g., make db-migrate-validate SERVICE_NAME=user-service)"; \
		exit 1; \
	fi
	@docker run --rm --network $(NETWORK_NAME) \
		--env-file $(ENV_FILE) \
		-e DB_HOST=$(POSTGRES_NAME) \
		-e DB_PORT=$(DATABASE_PORT) \
		-e DB_USER=$(DATABASE_USER) \
		-e DB_PASSWORD=$(DATABASE_PASSWORD) \
		-e DB_NAME=$(DATABASE_NAME) \
		-e DB_SSL_MODE=$(DATABASE_SSL_MODE) \
		-v $(PWD)/services:/services \
		$(ORCHESTRATOR_IMAGE) \
		validate $(SERVICE_NAME) --env $(APP_ENV)

.PHONY: db-migrate
db-migrate: ## Run migrations for all services (or specific service if SERVICE_NAME is set) using orchestrator
	@if [ -n "$(SERVICE_NAME)" ]; then \
		echo "📈 Running migrations for service: $(SERVICE_NAME)"; \
		$(MAKE) db-migrate-init SERVICE_NAME=$(SERVICE_NAME); \
		$(MAKE) db-migrate-up SERVICE_NAME=$(SERVICE_NAME); \
	else \
		echo "📈 Running migrations for all services: $(SERVICE_MIGRATION_ORDER)"; \
		$(MAKE) db-migrate-all; \
	fi

.PHONY: db-migrate-all
db-migrate-all: build-migrate-wrapper ## Run migrations for all services in correct order
	@echo "📈 Running migrations in order: $(SERVICE_MIGRATION_ORDER)"
	@for service in $(SERVICE_MIGRATION_ORDER); do \
		echo "📈 Running migrations for $$service..."; \
		$(MAKE) db-migrate-init SERVICE_NAME=$$service; \
		$(MAKE) db-migrate-up SERVICE_NAME=$$service || echo "❌ Failed to migrate $$service"; \
	done
	@echo "✅ All service migrations completed successfully"

.PHONY: db-rollback
db-rollback: db-migrate-down ## Rollback last migration (alias for db-migrate-down)

.PHONY: db-migrate-create
db-migrate-create: ## Create new migration file (NAME=add_users_table)
	@echo "📝 Creating migration: $(NAME)"
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Error: Please specify NAME (e.g., make db-migrate-create NAME=add_users_table)"; \
		exit 1; \
	fi
	@docker run --rm \
		-v $(PWD)/services/$(SERVICE_NAME)/migrations:/migrations \
		$(MIGRATION_IMAGE) \
		create -ext sql -dir /migrations -seq $(NAME)

.PHONY: db-migrate-file-list
db-migrate-file-list: ## List available migration files
	@echo "📁 Migration files in services/$(SERVICE_NAME)/migrations:"
	@ls -la services/$(SERVICE_NAME)/migrations/ 2>/dev/null || echo "No migration files found"

## Migration Orchestrator Targets (Enhanced Migration System)
# Build wrapper binary
.PHONY: build-migrate-wrapper
build-migrate-wrapper: ## Build migrate-wrapper binary
	@echo "🏗️  Building migrate-wrapper..."
	@docker build -t $(ORCHESTRATOR_IMAGE) -f migration-orchestrator/Dockerfile ./migration-orchestrator



## Data Management
.PHONY: db-seed
db-seed: ## Seed database with test data
	@echo "🌱 Seeding database with test data..."
	@cat scripts/seed.sql | $(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) exec -T postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) 2>/dev/null
	@if [ $$? -eq 0 ]; then \
		echo "✅ Database seeded with 5 test users"; \
	else \
		echo "❌ Database seeding failed"; \
	fi

.PHONY: db-seed-enhanced
db-seed-enhanced: ## Environment-specific seeding (ENV=development)
	@echo "🌱 Enhanced database seeding for environment: $(ENV)"
	@./scripts/enhanced_seed.sh $(ENV) user-service

.PHONY: db-migrate-generate
db-migrate-generate: ## Generate new migration with templates (NAME=add_feature TYPE=table)
	@echo "🚀 Generating migration: $(NAME) (type: $(TYPE))"
	@./scripts/generate_migration.sh $(SERVICE_NAME) "$(NAME)" "$(TYPE)"

.PHONY: db-validate
db-validate: ## Validate migration files and dependencies
	@echo "🔍 Validating migrations..."
	@./scripts/validate_migration.sh $(SERVICE_NAME)

.PHONY: db-migration-order
db-migration-order: ## Show migration execution order
	@echo "📋 Migration execution order: $(SERVICE_MIGRATION_ORDER)"

.PHONY: db-backup
db-backup: ## Create timestamped database backup
	@echo "💾 Creating database backup..."
	@BACKUP_FILE="backup_$(shell date +%Y%m%d_%H%M%S).sql"; \
	docker compose --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) -f $(DOCKER_COMPOSE_OVERRIDE_FILE) exec postgres pg_dump -U $(DATABASE_USER) -d $(DATABASE_NAME) --no-owner --no-privileges > "$$BACKUP_FILE" ; \
	if [ $$? -eq 0 ]; then \
		echo "✅ Database backup created: $$BACKUP_FILE"; \
		echo "   Size: $$(du -h "$$BACKUP_FILE" | cut -f1)"; \
	else \
		echo "❌ Database backup failed"; \
	fi

.PHONY: db-dump
db-dump: ## Create database dump
	@echo "💾 Creating database dump..."
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) exec postgres pg_dump -U $(DATABASE_USER) -d $(DATABASE_NAME) --no-owner --no-privileges > db_dump_$(shell date +%Y%m%d_%H%M%S).sql 2>/dev/null
	@if [ $$? -eq 0 ]; then \
		echo "✅ Database dump created: db_dump_$(shell date +%Y%m%d_%H%M%S).sql"; \
	else \
		echo "❌ Database dump failed"; \
	fi

.PHONY: db-restore
db-restore: ## Restore database from dump (usage: make db-restore FILE=dump.sql)
	@echo "🔄 Restoring database from $(FILE)..."
	@if [ -z "$(FILE)" ]; then \
		echo "❌ Error: Please specify FILE variable"; \
		echo "   Usage: make db-restore FILE=path/to/dump.sql"; \
		exit 1; \
	fi
	@if [ ! -f "$(FILE)" ]; then \
		echo "❌ Error: File $(FILE) not found"; \
		exit 1; \
	fi
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) exec -T postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) < $(FILE) 2>/dev/null
	@if [ $$? -eq 0 ]; then \
		echo "✅ Database restored from $(FILE)"; \
	else \
		echo "❌ Database restore failed"; \
	fi

.PHONY: db-clean
db-clean: ## Clean all data from tables (keep schema)
	@echo "🧹 Cleaning database data..."
	@cat scripts/clean.sql | docker compose --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) exec -T postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) 2>/dev/null
	@if [ $$? -eq 0 ]; then \
		echo "✅ Database data cleaned (schema preserved)"; \
	else \
		echo "❌ Database cleanup failed"; \
	fi

## Schema Inspection
.PHONY: db-schema
db-schema: ## Show database schema and tables
	@echo "📋 Database Schema:"
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) exec postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) -c "\dt" 2>/dev/null || echo "❌ Cannot access database schema"
	@echo ""
	@echo "📊 Indexes:"
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) exec postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) -c "\di" 2>/dev/null || echo "❌ Cannot access indexes"

.PHONY: db-tables
db-tables: ## List all tables and their structure
	@echo "📊 Table Structures:"
	@cat scripts/list_tables.sql | docker compose --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) exec -T postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) 2>/dev/null || echo "❌ Cannot list tables"
	@echo ""
	@if docker compose --env-file .env -f $(DOCKER_COMPOSE_FILE) exec postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) -c "\d user_service.users" 2>/dev/null; then \
		echo "✅ Users table structure displayed above"; \
	else \
		echo "⚠️  Users table not found or cannot display structure"; \
	fi

.PHONY: db-counts
db-counts: ## Show row counts and sizes for all tables
	@echo "🔢 Table Statistics:"
	@cat scripts/table_stats.sql | docker compose --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) exec -T postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) 2>/dev/null || echo "❌ Cannot query table statistics"

## Development Workflow Targets
.PHONY: db-setup
db-setup: db-create db-migrate db-seed ## Complete database setup for development using orchestrator
	@echo "🎉 Database setup complete!"
	@echo "   Database: $(DATABASE_NAME)"
	@echo "   Tables: users"
	@echo "   Test data: 5 users created"
	@echo "   Status: Ready for development"
	@echo "   Migration System: Enhanced Orchestrator"

.PHONY: db-reset-dev
db-reset-dev: db-drop db-setup ## Reset database for fresh development start
	@echo "🔄 Database reset complete for development"
	@echo "   All data cleared and fresh schema applied"

.PHONY: db-fresh
db-fresh: clean-volumes db-setup ## Complete database reset with volume cleanup
	@echo "🆕 Fresh database environment ready"
	@echo "   Volumes cleaned and database fully reset"

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
clean-all: down clean-go clean-cli clean-docker clean-volumes clean-logs clean-cache clean-test ## Complete clean for fresh start
	@echo "✅ Complete clean finished! All artifacts removed."

.PHONY: clean-go
clean-go: ## Clean Go build artifacts and cache
	@echo "🧹 Cleaning Go artifacts..."
	@$(GOCLEAN) -r
	@rm -rf $(BUILD_DIR)
	@rm -rf $(API_GATEWAY_DIR)/tmp
	@rm -rf $(USER_SERVICE_DIR)/tmp
	@rm -rf $(CLI_DIR)/tmp
	@find . -name "*.test" -type f -delete 2>/dev/null || true
	@find . -name "*.out" -type f -delete 2>/dev/null || true
	@find . -name "coverage.*" -type f -delete 2>/dev/null || true
	@echo "✅ Go artifacts cleaned"

# Smart Docker cleanup functions
.PHONY: check-image-in-use
check-image-in-use:
	@echo "🔍 Checking if $(IMAGE) is used by running containers..."
	@if docker ps --format "table {{.Image}}" | grep -q "^$(IMAGE)$$" 2>/dev/null; then \
		echo "⚠️  $(IMAGE) is used by running containers - skipping removal"; \
		exit 1; \
	else \
		echo "✅ $(IMAGE) not used by running containers"; \
	fi

.PHONY: check-image-dependencies
check-image-dependencies:
	@echo "🔗 Checking if $(IMAGE) is a base image for others..."
	@if docker images --format "table {{.Repository}}:{{.Tag}}\t{{.ID}}" | \
		grep -v "^$(IMAGE)" | xargs -I {} docker history {} 2>/dev/null | \
		grep -q "$(IMAGE)" 2>/dev/null; then \
		echo "⚠️  $(IMAGE) is a base image for other images - skipping removal"; \
		exit 1; \
	else \
		echo "✅ $(IMAGE) not a base for other images"; \
	fi

.PHONY: check-image-tags
check-image-tags:
	@echo "🏷️  Checking if $(IMAGE) has multiple tags..."
	@TAG_COUNT=$$(docker images $(IMAGE) --format "{{.Repository}}:{{.Tag}}" | wc -l 2>/dev/null || echo "0"); \
	if [ "$$TAG_COUNT" -gt 1 ]; then \
		echo "⚠️  $(IMAGE) has $$TAG_COUNT tags - likely managed by other projects"; \
		exit 1; \
	else \
		echo "✅ $(IMAGE) has single tag ($$TAG_COUNT)"; \
	fi

.PHONY: safe-remove-image
safe-remove-image:
	@echo "🗑️  Attempting to safely remove $(IMAGE)..."
	@if $(MAKE) check-image-in-use IMAGE=$(IMAGE) && \
	   $(MAKE) check-image-dependencies IMAGE=$(IMAGE) && \
	   $(MAKE) check-image-tags IMAGE=$(IMAGE); then \
		echo "🟢 All checks passed - removing $(IMAGE)"; \
		docker rmi $(IMAGE) 2>/dev/null || echo "ℹ️  $(IMAGE) already removed or not found"; \
	else \
		echo "🟡 $(IMAGE) is in use or has dependencies - keeping it"; \
	fi

.PHONY: clean-docker-smart
clean-docker-smart: ## Smart Docker cleanup with safety checks
	@echo "🧠 Smart cleaning of project Docker artifacts..."
	@$(DOCKER_COMPOSE) --env-file .env down --volumes --remove-orphans 2>/dev/null || true
	@echo "🗂️  Removing service containers..."
	@for container in $(SERVICE_CONTAINERS) $(POSTGRES_CONTAINER); do \
		if [ -n "$$container" ]; then \
			echo "  Removing container: $$container"; \
			docker rm $$container 2>/dev/null || true; \
		fi; \
	done
	@echo "🗂️  Removing monitoring containers..."
	@for container in $(MONITORING_CONTAINERS); do \
		if [ -n "$$container" ]; then \
			echo "  Removing container: $$container"; \
			docker rm $$container 2>/dev/null || true; \
		fi; \
	done
	@echo "🖼️  Removing custom project images..."
	@for image in $(SERVICE_IMAGES); do \
		if [ -n "$$image" ]; then \
			echo "  Removing image: $$image"; \
			docker rmi $$image 2>/dev/null || true; \
		fi; \
	done
	@echo "🧠 Smart cleanup of base images..."
	@$(MAKE) safe-remove-image IMAGE=$(MIGRATION_IMAGE)
	@$(MAKE) safe-remove-image IMAGE=$(POSTGRES_IMAGE)
	@$(MAKE) safe-remove-image IMAGE=$(GOLANG_BUILD_IMAGE)
	@$(MAKE) safe-remove-image IMAGE=$(ALPINE_RUNTIME_IMAGE)
	@echo "📊 Removing monitoring images..."
	@for image in $(MONITORING_IMAGES); do \
		if [ -n "$$image" ]; then \
			echo "  Removing image: $$image"; \
			docker rmi $$image 2>/dev/null || true; \
		fi; \
	done
	@echo "💾 Removing service volumes..."
	@for volume in $(SERVICE_VOLUMES) $(POSTGRES_VOLUME); do \
		if [ -n "$$volume" ]; then \
			echo "  Removing volume: $$volume"; \
			docker volume rm $$volume 2>/dev/null || true; \
		fi; \
	done
	@echo "💾 Removing monitoring volumes..."
	@for volume in $(MONITORING_VOLUMES); do \
		echo "  Removing volume: $$volume"; \
		docker volume rm $$volume 2>/dev/null || true; \
	done
	@docker network rm $(NETWORK_NAME) 2>/dev/null || true
	@echo "✅ Smart Docker cleanup completed"

.PHONY: clean-docker-conservative
clean-docker-conservative: ## Conservative Docker cleanup (keeps base images)
	@echo "🐳 Conservative cleaning of project Docker artifacts..."
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) down --volumes --remove-orphans 2>/dev/null || true
	@echo "🗂️  Removing service containers..."
	@for container in $(SERVICE_CONTAINERS) $(POSTGRES_CONTAINER); do \
		if [ -n "$$container" ]; then \
			echo "  Removing container: $$container"; \
			docker rm $$container 2>/dev/null || true; \
		fi; \
	done
	@echo "🗂️  Removing monitoring containers..."
	@for container in $(MONITORING_CONTAINERS); do \
		if [ -n "$$container" ]; then \
			echo "  Removing container: $$container"; \
			docker rm $$container 2>/dev/null || true; \
		fi; \
	done
	@echo "🖼️  Removing custom project images..."
	@for image in $(SERVICE_IMAGES); do \
		if [ -n "$$image" ]; then \
			echo "  Removing image: $$image"; \
			docker rmi $$image 2>/dev/null || true; \
		fi; \
	done
	@docker rmi $(MIGRATION_IMAGE) 2>/dev/null || true
	@echo "📊 Removing monitoring images..."
	@for image in $(MONITORING_IMAGES); do \
		if [ -n "$$image" ]; then \
			echo "  Removing image: $$image"; \
			docker rmi $$image 2>/dev/null || true; \
		fi; \
	done
	@echo "💾 Removing service volumes..."
	@for volume in $(SERVICE_VOLUMES) $(POSTGRES_VOLUME); do \
		if [ -n "$$volume" ]; then \
			echo "  Removing volume: $$volume"; \
			docker volume rm $$volume 2>/dev/null || true; \
		fi; \
	done
	@echo "💾 Removing monitoring volumes..."
	@for volume in $(MONITORING_VOLUMES); do \
		if [ -n "$$volume" ]; then \
			echo "  Removing volume: $$volume"; \
			docker volume rm $$volume 2>/dev/null || true; \
		fi; \
	done
	@docker network rm $(NETWORK_NAME) 2>/dev/null || true
	@echo "✅ Conservative Docker cleanup completed (base images preserved)"

.PHONY: clean-docker-aggressive
clean-docker-aggressive: ## Aggressive Docker cleanup (removes all project images)
	@echo "💥 Aggressive cleaning of project Docker artifacts..."
	@$(DOCKER_COMPOSE) --env-file .env down --volumes --remove-orphans 2>/dev/null || true
	@echo "🗂️  Removing service containers..."
	@for container in $(SERVICE_CONTAINERS) $(POSTGRES_CONTAINER); do \
		if [ -n "$$container" ]; then \
			echo "  Removing container: $$container"; \
			docker rm $$container 2>/dev/null || true; \
		fi; \
	done
	@echo "🗂️  Removing monitoring containers..."
	@for container in $(MONITORING_CONTAINERS); do \
		if [ -n "$$container" ]; then \
			echo "  Removing container: $$container"; \
			docker rm $$container 2>/dev/null || true; \
		fi; \
	done
	@echo "🖼️  Removing all project images..."
	@for image in $(SERVICE_IMAGES); do \
		if [ -n "$$image" ]; then \
			echo "  Removing image: $$image"; \
			docker rmi $$image 2>/dev/null || true; \
		fi; \
	done
	@docker rmi $(MIGRATION_IMAGE) 2>/dev/null || true
	@docker rmi $(POSTGRES_IMAGE) 2>/dev/null || true
	@docker rmi $(GOLANG_BUILD_IMAGE) 2>/dev/null || true
	@docker rmi $(ALPINE_RUNTIME_IMAGE) 2>/dev/null || true
	@echo "📊 Removing monitoring images..."
	@for image in $(MONITORING_IMAGES); do \
		if [ -n "$$image" ]; then \
			echo "  Removing image: $$image"; \
			docker rmi $$image 2>/dev/null || true; \
		fi; \
	done
	@echo "💾 Removing service volumes..."
	@for volume in $(SERVICE_VOLUMES) $(POSTGRES_VOLUME); do \
		if [ -n "$$volume" ]; then \
			echo "  Removing volume: $$volume"; \
			docker volume rm $$volume 2>/dev/null || true; \
		fi; \
	done
	@echo "💾 Removing monitoring volumes..."
	@for volume in $(MONITORING_VOLUMES); do \
		if [ -n "$$volume" ]; then \
			echo "  Removing volume: $$volume"; \
			docker volume rm $$volume 2>/dev/null || true; \
		fi; \
	done
	@docker network rm $(NETWORK_NAME) 2>/dev/null || true
	@echo "✅ Aggressive Docker cleanup completed"

.PHONY: clean-docker
clean-docker: ## Clean project Docker artifacts (mode: $(DOCKER_CLEANUP_MODE))
	@echo "🐳 Cleaning Docker artifacts (mode: $(DOCKER_CLEANUP_MODE))..."
	@if [ "$(DOCKER_CLEANUP_MODE)" = "conservative" ]; then \
		$(MAKE) clean-docker-conservative; \
	elif [ "$(DOCKER_CLEANUP_MODE)" = "aggressive" ]; then \
		$(MAKE) clean-docker-aggressive; \
	else \
		$(MAKE) clean-docker-smart; \
	fi

.PHONY: clean-volumes
clean-volumes: ## Clean Docker volumes and persistent data
	@echo "💾 Cleaning Docker volumes..."
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) --file $(DOCKER_COMPOSE_FILE) --file $(DOCKER_COMPOSE_OVERRIDE_FILE) down --volumes
	@echo "🔧 Cleaning volume data using Docker containers..."
	@mkdir -p docker/volumes  # Ensure parent directory exists with correct ownership
	@echo "📁 Removing postgres volume..."
	@docker run --rm -v $(PWD)/docker/volumes:/volumes alpine sh -c "rm -rf /volumes/postgres_data";
	@echo "📁 Removing monitoring volumes..."
	@for dir in $(MONITORING_VOLUME_DIRS); do \
		echo "  Cleaning $$dir volume data..."; \
		docker run --rm -v $(PWD)/docker/volumes/$$dir:/data alpine sh -c "rm -rf /data/*" 2>/dev/null || true; \
	done
	@if [ -d "docker/volumes" ]; then \
		for dir in docker/volumes/*/; do \
			if [ -d "$$dir" ]; then \
				service_name=$$(basename "$$dir"); \
				# Skip monitoring volumes as they're handled above \
				skip_monitoring=false; \
				for monitor_dir in $(MONITORING_VOLUME_DIRS); do \
					if [ "$$service_name" = "$$monitor_dir" ]; then \
						skip_monitoring=true; \
						break; \
					fi; \
				done; \
				if [ "$$skip_monitoring" = "false" ]; then \
					echo " 📁 Cleaning $$service_name volumes..."; \
					docker run --rm -v $(PWD)/$$dir:/data alpine sh -c "rm -rf /data/*"; \
				fi; \
			fi; \
		done; \
	fi
	@if [ -d "tmp" ]; then \
		echo "  📁 Cleaning migration temp files..."; \
		docker run --rm -v $(PWD)/tmp:/data alpine sh -c "rm -rf /data/migrations 2>/dev/null || true" 2>/dev/null || true; \
	fi
	@echo "🗑️  Removing empty volume directories..."
	@find docker/volumes -type d -empty -delete 2>/dev/null || true
	@rmdir docker/volumes 2>/dev/null || true
	@rmdir tmp 2>/dev/null || true
	@echo "✅ Docker volumes cleaned"

.PHONY: clean-logs
clean-logs: ## Clean log files
	@echo "📝 Cleaning log files..."
	@find . -name "*.log" -type f -delete 2>/dev/null || true
	@find . -name "build-errors.log" -type f -delete 2>/dev/null || true
	@rm -rf logs/ 2>/dev/null || true
	@if [ -d "docker/volumes" ]; then \
		for service_dir in docker/volumes/*/; do \
			if [ -d "$$service_dir/logs" ]; then \
				echo "  Cleaning logs in $$service_dir"; \
				rm -rf $$service_dir/logs/*.log 2>/dev/null || true; \
				rm -rf $$service_dir/logs/*.gz 2>/dev/null || true; \
			fi; \
		done; \
	fi
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

# Docker Environment Management
.PHONY: docker-reset
docker-reset: ## Complete project Docker environment reset
	@echo "🔄 Starting complete project Docker reset..."

	# Stop and remove project containers (exact names)
	@echo "   Stopping project containers..."
	@docker stop $(API_GATEWAY_CONTAINER) 2>/dev/null || true
	@docker stop $(USER_SERVICE_CONTAINER) 2>/dev/null || true
	@docker stop $(POSTGRES_CONTAINER) 2>/dev/null || true

	@echo "   Removing project containers..."
	@docker rm $(API_GATEWAY_CONTAINER) 2>/dev/null || true
	@docker rm $(USER_SERVICE_CONTAINER) 2>/dev/null || true
	@docker rm $(POSTGRES_CONTAINER) 2>/dev/null || true

	# Remove project images (exact names)
	@echo "   Removing project images..."
	@docker rmi $(API_GATEWAY_IMAGE) 2>/dev/null || true
	@docker rmi $(USER_SERVICE_IMAGE) 2>/dev/null || true

	# Remove project volumes (exact names)
	@echo "   Removing project volumes..."
	@docker volume rm $(POSTGRES_VOLUME) 2>/dev/null || true
	@docker volume rm $(API_GATEWAY_TMP_VOLUME) 2>/dev/null || true
	@docker volume rm $(USER_SERVICE_TMP_VOLUME) 2>/dev/null || true

	# Remove project networks (exact names)
	@echo "   Removing project networks..."
	@docker network rm $(NETWORK_NAME) 2>/dev/null || true

	# Clean up volume directories
	@echo "   Cleaning volume directories..."
	@rm -rf docker/volumes/ 2>/dev/null || true

	@echo "✅ Project Docker environment reset complete"
	@echo "   Run 'make docker-recreate' to recreate from scratch"

.PHONY: docker-reset-confirm
docker-reset-confirm: ## Reset project Docker environment with confirmation
	@echo "🔄 Project Docker Environment Reset"
	@echo ""
	@echo "This will remove:"
	@echo "  • Container: $(API_GATEWAY_CONTAINER)"
	@echo "  • Container: $(USER_SERVICE_CONTAINER)"
	@echo "  • Container: $(POSTGRES_CONTAINER)"
	@echo "  • Image: $(API_GATEWAY_IMAGE)"
	@echo "  • Image: $(USER_SERVICE_IMAGE)"
	@echo "  • Volume: $(POSTGRES_VOLUME)"
	@echo "  • Volume: $(API_GATEWAY_TMP_VOLUME)"
	@echo "  • Volume: $(USER_SERVICE_TMP_VOLUME)"
	@echo "  • Network: $(NETWORK_NAME)"
	@echo "  • All volume data and directories"
	@echo ""
	@echo "The environment can be recreated with: make docker-recreate"
	@echo ""
	@read -p "Are you sure you want to reset the project Docker environment? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		$(MAKE) docker-reset; \
	else \
		echo "❌ Reset cancelled"; \
	fi

.PHONY: create-volumes-dirs
create-volumes-dirs: ## (Re)create volumes directories
	@echo "🔄 Recreating volumes directories..."

	# Check for root ownership (informational only - manual fix required)
	@if [ -d "docker/volumes" ] && [ "$$(stat -c %U docker/volumes 2>/dev/null)" = "root" ]; then \
		echo "⚠️  WARNING: docker/volumes is owned by root. This may cause permission issues."; \
		echo "   To fix: sudo chown -R $$(id -u):$$(id -g) docker/volumes"; \
	fi

	# Create volume directories
	@echo "   Creating volume directories..."
	@mkdir -p docker/volumes/postgres_data
	@for service in $$(grep "_TMP_VOLUME=" $(ENV_FILE) | cut -d'=' -f2 | sed 's/$(DOCKER_PROJECT_PREFIX)-//' | sed 's/-tmp$$//'); do \
		echo "   Creating directory for $$service..."; \
		mkdir -p docker/volumes/$$service/tmp; \
	done
	@for service in $$(grep "_LOGS_VOLUME=" $(ENV_FILE) | cut -d'=' -f2 | sed 's/$(DOCKER_PROJECT_PREFIX)-//' | sed 's/-logs$$//'); do \
		echo "   Creating logs directory for $$service..."; \
		mkdir -p docker/volumes/$$service/logs; \
	done
	@echo "   Creating monitoring volume directories..."
	@echo "   Creating directory for grafana..."
	@mkdir -p docker/volumes/grafana/data
	@echo "   Creating directory for jaeger..."
	@mkdir -p docker/volumes/jaeger/data
	@echo "   Creating directory for loki..."
	@mkdir -p docker/volumes/loki/data
	@echo "   Creating directory for promtail..."
	@mkdir -p docker/volumes/promtail/positions

.PHONY: docker-recreate
docker-recreate: create-volumes-dirs ## Recreate project Docker environment from scratch
	@echo "🔄 Recreating project Docker environment..."

	# Build images from scratch
	@echo "   Building images from scratch..."
	@make docker-build

	# Start services
	@echo "   Starting services..."
	@make up

	@echo "✅ Project Docker environment recreated"
	@echo "   Services should be available at:"
	@echo "   • API Gateway: http://localhost:8080"
	@echo "   • User Service: http://localhost:8081"
	@echo "   • PostgreSQL: localhost:5432"

.PHONY: help-network
help-network: ## Show network commands
	@echo "🌐 Network Commands:"
	@echo "  network-create     - Create custom Docker network"
	@echo "  network-inspect    - Inspect Docker network"
	@echo "  network-ls         - List Docker networks"
	@echo "  network-clean      - Clean up unused networks"
	@echo "  network-remove     - Remove custom network"

# ============================================================================
# 🏥 HEALTH & MONITORING TARGETS
# ============================================================================

.PHONY: health
health: ## Comprehensive health check of all services
	@echo "🏥 Service Boilerplate Health Check"
	@echo "=================================="
	@echo ""
	@echo "🔍 Checking container status..."
	@$(MAKE) health-containers
	@echo ""
	@echo "🌐 Checking service endpoints..."
	@$(MAKE) health-services
	@echo ""
	@echo "🗄️  Checking database connectivity..."
	@$(MAKE) health-database
	@echo ""
	@echo "📡 Checking network status..."
	@$(MAKE) health-network
	@echo ""
	@echo "💾 Checking volume mounts..."
	@$(MAKE) health-volumes
	@echo ""
	@echo "✅ Health check completed!"

.PHONY: health-containers
health-containers: ## Check Docker container status
	@echo "🐳 Container Status:"
	@CONTAINERS="$$(docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep $(DOCKER_PROJECT_PREFIX))"; \
	if [ -n "$$CONTAINERS" ]; then \
		echo "$$CONTAINERS" | while read line; do \
			if echo "$$line" | grep -q "Up"; then \
				echo "  ✅ $$line"; \
			else \
				echo "  ❌ $$line"; \
			fi; \
		done; \
	else \
		echo "  ⚠️  No $(DOCKER_PROJECT_PREFIX) containers running"; \
	fi

.PHONY: health-services
health-services: ## Check HTTP health endpoints
	@echo "🌐 Service Health Endpoints:"
	@API_GATEWAY_HEALTH=$$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health 2>/dev/null || echo "000"); \
	if [ "$$API_GATEWAY_HEALTH" = "200" ]; then \
		echo "  ✅ API Gateway (localhost:8080/health) - HTTP $$API_GATEWAY_HEALTH"; \
	else \
		echo "  ❌ API Gateway (localhost:8080/health) - HTTP $$API_GATEWAY_HEALTH"; \
	fi
	@USER_SERVICE_HEALTH=$$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/health 2>/dev/null || echo "000"); \
	if [ "$$USER_SERVICE_HEALTH" = "200" ]; then \
		echo "  ✅ User Service (localhost:8081/health) - HTTP $$USER_SERVICE_HEALTH"; \
	else \
		echo "  ❌ User Service (localhost:8081/health) - HTTP $$USER_SERVICE_HEALTH"; \
	fi

.PHONY: health-database
health-database: ## Check database connectivity
	@echo "🗄️  Database Connectivity:"
	@DB_STATUS=$$(docker compose --env-file .env -f $(DOCKER_COMPOSE_FILE) exec -T postgres pg_isready -U $(DATABASE_USER) -d $(DATABASE_NAME) -h $(DATABASE_HOST) -p $(DATABASE_PORT) 2>/dev/null || echo "failed"); \
	if echo "$$DB_STATUS" | grep -q "accepting connections"; then \
		echo "  ✅ PostgreSQL accepting connections"; \
		CONNECTIONS=$$(docker compose --env-file .env -f $(DOCKER_COMPOSE_FILE) exec -T postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) -c "SELECT count(*) as active_connections FROM pg_stat_activity;" 2>/dev/null | tail -3 | head -1 | tr -d ' ' || echo "unknown"); \
		echo "  📊 Active connections: $$CONNECTIONS"; \
	else \
		echo "  ❌ PostgreSQL not accepting connections"; \
	fi

.PHONY: health-network
health-network: ## Check Docker network status
	@echo "📡 Docker Network Status:"
	@NETWORK_STATUS=$$(docker network ls --format "table {{.Name}}\t{{.Driver}}" | grep $(NETWORK_NAME) || echo "not found"); \
	if echo "$$NETWORK_STATUS" | grep -q $(NETWORK_NAME); then \
		echo "  ✅ Network $(NETWORK_NAME) exists"; \
		CONNECTED_CONTAINERS=$$(docker network inspect $(NETWORK_NAME) --format '{{range .Containers}}{{.Name}} {{end}}' 2>/dev/null || echo "unknown"); \
		if [ "$$CONNECTED_CONTAINERS" != "unknown" ] && [ -n "$$CONNECTED_CONTAINERS" ]; then \
			echo "  🔗 Connected containers: $$CONNECTED_CONTAINERS"; \
		else \
			echo "  ⚠️  No containers connected to network"; \
		fi; \
	else \
		echo "  ❌ Network $(NETWORK_NAME) not found"; \
	fi

.PHONY: health-volumes
health-volumes: ## Check volume mount status
	@echo "💾 Docker Volume Status:"
	@VOLUMES="$$(docker volume ls --format "table {{.Name}}" | grep $(DOCKER_PROJECT_PREFIX))"; \
	if [ -n "$$VOLUMES" ]; then \
		echo "$$VOLUMES" | while read volume; do \
			if [ "$$volume" != "NAME" ]; then \
				VOLUME_PATH=$$(docker volume inspect $$volume --format '{{.Mountpoint}}' 2>/dev/null || echo "unknown"); \
				if [ -d "$$VOLUME_PATH" ]; then \
					echo "  ✅ $$volume mounted at $$VOLUME_PATH"; \
				else \
					echo "  ❌ $$volume mount point not accessible"; \
				fi; \
			fi; \
		done; \
	else \
		echo "  ⚠️  No $(DOCKER_PROJECT_PREFIX) volumes found"; \
	fi
	@HOST_VOLUMES="docker/volumes/postgres_data docker/volumes/api-gateway docker/volumes/user-service tmp"; \
	for volume in $$HOST_VOLUMES; do \
		if [ -d "$$volume" ]; then \
			FILE_COUNT=$$(find $$volume -type f 2>/dev/null | wc -l); \
			echo "  📁 $$volume exists ($$FILE_COUNT files)"; \
		else \
			echo "  ℹ️  $$volume directory not present"; \
		fi; \
	done

.PHONY: clean-docker-report
clean-docker-report: ## Report on Docker cleanup status
	@echo "📊 Docker Cleanup Report:"
	@echo ""
	@echo "🔍 Current cleanup mode: $(DOCKER_CLEANUP_MODE)"
	@echo ""
	@echo "🖼️  Images that would be removed in $(DOCKER_CLEANUP_MODE) mode:"
	@if [ "$(DOCKER_CLEANUP_MODE)" = "conservative" ]; then \
		echo "  • $(API_GATEWAY_IMAGE)"; \
		echo "  • $(USER_SERVICE_IMAGE)"; \
		echo "  • $(MIGRATION_IMAGE)"; \
	elif [ "$(DOCKER_CLEANUP_MODE)" = "aggressive" ]; then \
		echo "  • $(API_GATEWAY_IMAGE)"; \
		echo "  • $(USER_SERVICE_IMAGE)"; \
		echo "  • $(MIGRATION_IMAGE)"; \
		echo "  • $(POSTGRES_IMAGE)"; \
		echo "  • $(GOLANG_BUILD_IMAGE)"; \
		echo "  • $(ALPINE_RUNTIME_IMAGE)"; \
	else \
		echo "  • $(API_GATEWAY_IMAGE) (always)"; \
		echo "  • $(USER_SERVICE_IMAGE) (always)"; \
		echo "  • $(MIGRATION_IMAGE) (if safe)"; \
		echo "  • $(POSTGRES_IMAGE) (if safe)"; \
		echo "  • $(GOLANG_BUILD_IMAGE) (if safe)"; \
		echo "  • $(ALPINE_RUNTIME_IMAGE) (if safe)"; \
	fi
	@echo ""
	@echo "🏃 Running containers:"
	@docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Status}}" 2>/dev/null || echo "  No running containers"
	@echo ""
	@echo "🖼️  Current project images:"
	@docker images --format "table {{.Repository}}:{{.Tag}}\t{{.Size}}" | grep -E "($(DOCKER_PROJECT_PREFIX)|migrate|postgres|golang|alpine)" 2>/dev/null || echo "  No project images found"
	@echo ""
	@echo "💡 To change cleanup mode: make clean-docker DOCKER_CLEANUP_MODE=conservative"

.PHONY: clean-docker-dry-run
clean-docker-dry-run: ## Preview what would be cleaned (dry run)
	@echo "🔍 Docker Cleanup Dry Run (mode: $(DOCKER_CLEANUP_MODE))"
	@echo "This shows what WOULD be cleaned, but nothing is actually removed."
	@echo ""
	$(MAKE) clean-docker-report
	@echo ""
	@echo "💡 To actually clean: make clean-docker"
	@echo "💡 To change mode: make clean-docker DOCKER_CLEANUP_MODE=conservative"

.PHONY: help-docker
help-docker: ## Show Docker management commands
	@echo "🐳 Docker Management Commands:"
	@echo "  docker-reset           - Complete project Docker environment reset"
	@echo "  docker-reset-confirm   - Reset with confirmation prompt"
	@echo "  docker-recreate        - Recreate project Docker environment"
	@echo "  clean-docker           - Clean project Docker artifacts"
	@echo "  clean-docker-report    - Report on cleanup status"
	@echo "  clean-docker-dry-run   - Preview cleanup without removing anything"
	@echo ""
	@echo "🧠 Smart Cleanup Modes:"
	@echo "  DOCKER_CLEANUP_MODE=smart       - Intelligent cleanup (default)"
	@echo "  DOCKER_CLEANUP_MODE=conservative - Keeps base images"
	@echo "  DOCKER_CLEANUP_MODE=aggressive   - Removes all images"
	@echo ""
	@echo "⚠️  Note: Volume cleanup may require sudo for root-owned files"
	@echo "   created by Docker containers."
	@echo ""
	@echo "📝 Examples:"
	@echo "  make clean-docker                           # Smart cleanup"
	@echo "  make clean-docker DOCKER_CLEANUP_MODE=conservative"
	@echo "  make clean-docker-report                    # See what would be cleaned"

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

.PHONY: help-health
help-health: ## Show health and monitoring commands
	@echo "🏥 Health & Monitoring Commands:"
	@echo "  health             - Comprehensive health check of all services"
	@echo "  health-services    - Check HTTP health endpoints only"
	@echo "  health-containers  - Check Docker container status only"
	@echo "  health-database    - Check database connectivity only"
	@echo "  health-network     - Check Docker network status only"
	@echo "  health-volumes     - Check volume mount status only"
	@echo ""
	@echo "💡 Health Check Features:"
	@echo "  • Real-time status monitoring"
	@echo "  • HTTP endpoint validation"
	@echo "  • Database connectivity checks"
	@echo "  • Docker infrastructure validation"
	@echo "  • Color-coded results (✅ ❌ ⚠️ ℹ️)"
	@echo "  • CI/CD pipeline friendly"

.PHONY: logs-grafana
logs-grafana: ## Open Grafana UI for centralized logging (http://localhost:3000)
	@echo "🌐 Opening Grafana for centralized logging..."
	@echo "   URL: http://localhost:3000"
	@echo "   Username: admin"
	@echo "   Password: admin"
	@echo ""
	@echo "📊 Pre-configured dashboards:"
	@echo "   • Service Boilerplate - Logs"
	@echo ""
	@echo "🔍 Useful LogQL queries:"
	@echo "   • All service logs: {job=~\".*\"}"
	@echo "   • API Gateway logs: {service=\"api-gateway\"}"
	@echo "   • Error logs only: {level=\"error\"}"
	@echo "   • Request logs: {method=~\"GET|POST\"}"

.PHONY: logs-loki
logs-loki: ## View Loki service logs
	@echo "📊 Loki Service Logs:"
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) logs -f loki

.PHONY: logs-promtail
logs-promtail: ## View Promtail log shipping logs
	@echo "📤 Promtail Log Shipping Logs:"
	@$(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) logs -f promtail

.PHONY: loki-status
loki-status: ## Check Loki stack health and status
	@echo "📊 Loki Stack Status:"
	@echo ""
	@echo "🐳 Container Status:"
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "(loki|promtail|grafana)" || echo "   No Loki stack containers running"
	@echo ""
	@echo "🌐 Service Endpoints:"
	@echo "   • Grafana UI: http://localhost:3000"
	@echo "   • Loki API: http://localhost:3100"
	@echo "   • Jaeger UI: http://localhost:16686"
	@echo ""
	@echo "💾 Data Volumes:"
	@docker volume ls --format "table {{.Name}}" | grep -E "(loki|grafana)" || echo "   No Loki volumes found"

.PHONY: help-loki
help-loki: ## Show Loki logging stack commands
	@echo "📊 Loki Logging Stack Commands:"
	@echo "  logs-grafana       - View centralized logs in Grafana UI"
	@echo "  logs-loki          - View Loki service logs"
	@echo "  logs-promtail      - View Promtail log shipping logs"
	@echo "  loki-status        - Check Loki stack health"
	@echo ""
	@echo "🌐 Service URLs:"
	@echo "  • Grafana: http://localhost:3000 (admin/admin)"
	@echo "  • Loki: http://localhost:3100"
	@echo "  • Jaeger: http://localhost:16686"
	@echo ""
	@echo "📋 Log Queries (Grafana/Loki):"
	@echo "  • All logs: {job=~\".*\"}"
	@echo "  • API Gateway: {service=\"api-gateway\"}"
	@echo "  • User Service: {service=\"user-service\"}"
	@echo "  • Auth Service: {service=\"auth-service\"}"
	@echo "  • Errors only: {level=\"error\"}"

.PHONY: help-db
help-db: ## Show database commands
	@echo "🗄️  Database Commands:"
	@echo ""
	@echo "Connection & Access:"
	@echo "  db-connect         - Connect to database shell"
	@echo "  db-status          - Show database status and connections"
	@echo "  db-health          - Check database health and connectivity"
	@echo ""
	@echo "Database Management:"
	@echo "  db-create          - Create database if it does not exist"
	@echo "  db-drop            - Drop database (with confirmation)"
	@echo "  db-recreate        - Recreate database from scratch"
	@echo ""
	@echo "Migration Management (Orchestrator):"
	@echo "  db-migrate-init    - Initialize migration tracking for service"
	@echo "  db-migrate         - Run migrations for all services (or specific with SERVICE_NAME=)"
	@echo "  db-migrate-up      - Run migrations up for specific service"
	@echo "  db-migrate-down    - Run migrations down for specific service"
	@echo "  db-migrate-status  - Show migration status for specific service"
	@echo "  db-migrate-list    - List migration executions for specific service"
	@echo "  db-migrate-validate - Validate migrations for specific service"
	@echo "  db-migrate-create NAME= - Create migration file"
	@echo "  db-migrate-generate NAME= TYPE= - Generate migration with templates"
	@echo "  db-migrate-file-list  - List migration files"
	@echo ""
	@echo "Data Management:"
	@echo "  db-seed            - Seed database with test data"
	@echo "  db-seed-enhanced ENV= - Environment-specific seeding"
	@echo "  db-dump            - Create database dump"
	@echo "  db-restore FILE=   - Restore database from dump"
	@echo "  db-clean           - Clean all data from tables"
	@echo ""
	@echo "Schema Inspection:"
	@echo "  db-schema          - Show database schema"
	@echo "  db-tables          - List all tables and structure"
	@echo "  db-counts          - Show row counts for all tables"
	@echo ""
	@echo "Development Workflow:"
	@echo "  db-setup           - Complete database setup"
	@echo "  db-reset-dev       - Reset database for development"
	@echo "  db-fresh           - Complete reset with volume cleanup"
	@echo ""
	@echo "Advanced Features:"
	@echo "  db-validate        - Validate migration files and dependencies"
	@echo "  db-migration-order - Show migration execution order"
	@echo "  db-backup          - Create timestamped database backup"
	@echo ""
	@echo "Examples:"
	@echo "  make db-setup                    # Setup database for development"
	@echo "  make db-connect                  # Open database shell"
	@echo "  make db-migrate-init SERVICE_NAME=user-service  # Initialize migrations"
	@echo "  make db-migrate-up SERVICE_NAME=user-service    # Run migrations"
	@echo "  make db-migrate-generate NAME=add_user_preferences TYPE=table"
	@echo "  make db-seed-enhanced ENV=development"
	@echo "  make db-validate                 # Validate all migrations"
	@echo "  make db-backup                   # Create backup before changes"

.PHONY: build-auth-service
build-auth-service: ## Build auth-service
	@echo "Building auth-service..."
	@mkdir -p $(BUILD_DIR)
	@cd services/auth-service && $(GOBUILD) -o ../$(BUILD_DIR)/auth-service ./cmd

.PHONY: run-auth-service
run-auth-service: ## Run auth-service
	@echo "Running auth-service..."
	@cd services/auth-service && $(GO) run ./cmd

.PHONY: test-auth-service
test-auth-service: ## Run auth-service tests
	@echo "Running auth-service tests..."
	@cd services/auth-service && $(GOTEST) ./...

.PHONY: air-auth-service
air-auth-service: ## Run auth-service with Air in Docker
	@echo "Starting auth-service with Air..."
	@cd services/auth-service && air

.PHONY: build-objects-service
build-objects-service: ## Build objects-service
	@echo "Building objects-service..."
	@mkdir -p $(BUILD_DIR)
	@cd services/objects-service && $(GOBUILD) -o ../$(BUILD_DIR)/objects-service ./cmd

.PHONY: run-objects-service
run-objects-service: ## Run objects-service
	@echo "Running objects-service..."
	@cd services/objects-service && $(GO) run ./cmd

.PHONY: test-objects-service
test-objects-service: ## Run objects-service tests
	@echo "Running objects-service tests..."
	@cd services/objects-service && $(GOTEST) ./...

.PHONY: air-objects-service
air-objects-service: ## Run objects-service with Air locally
	@echo "Starting objects-service with Air..."
	@cd services/objects-service && air
