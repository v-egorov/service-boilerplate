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

# Docker naming variables (loaded from .env)
DOCKER_PROJECT_PREFIX := service-boilerplate
DOCKER_IMAGE_PREFIX := docker
DOCKER_CONTAINER_PREFIX := $(DOCKER_PROJECT_PREFIX)
DOCKER_NETWORK_PREFIX := $(DOCKER_PROJECT_PREFIX)
DOCKER_VOLUME_PREFIX := $(DOCKER_PROJECT_PREFIX)
API_GATEWAY_NAME := api-gateway
USER_SERVICE_NAME := user-service
POSTGRES_NAME := postgres
API_GATEWAY_IMAGE := $(DOCKER_IMAGE_PREFIX)-$(API_GATEWAY_NAME)
USER_SERVICE_IMAGE := $(DOCKER_IMAGE_PREFIX)-$(USER_SERVICE_NAME)
API_GATEWAY_CONTAINER := $(DOCKER_CONTAINER_PREFIX)-$(API_GATEWAY_NAME)
USER_SERVICE_CONTAINER := $(DOCKER_CONTAINER_PREFIX)-$(USER_SERVICE_NAME)
POSTGRES_CONTAINER := $(DOCKER_CONTAINER_PREFIX)-$(POSTGRES_NAME)
NETWORK_NAME := $(DOCKER_NETWORK_PREFIX)-network
POSTGRES_VOLUME := $(DOCKER_VOLUME_PREFIX)-postgres-data
API_GATEWAY_TMP_VOLUME := $(DOCKER_VOLUME_PREFIX)-api-gateway-tmp
USER_SERVICE_TMP_VOLUME := $(DOCKER_VOLUME_PREFIX)-user-service-tmp

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ''
	@echo 'Database Commands:'
	@echo '  db-connect         - Connect to database shell'
	@echo '  db-status          - Show database status and connections'
	@echo '  db-health          - Check database health and connectivity'
	@echo '  db-create          - Create database if it does not exist'
	@echo '  db-drop            - Drop database (with confirmation)'
	@echo '  db-recreate        - Recreate database from scratch'
	@echo '  db-migrate-up      - Run migrations up'
	@echo '  db-migrate-down    - Run migrations down'
	@echo '  db-migrate-status  - Show migration status'
	@echo '  db-migrate-goto VERSION= - Go to specific version'
	@echo '  db-migrate-validate - Validate migration files'
	@echo '  db-migration-create NAME= - Create migration file'
	@echo '  db-migration-list  - List migration files'
	@echo '  db-seed            - Seed database with test data'
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
migrate-up: ## Run database migrations (legacy - use db-migrate-up)
	@echo "⚠️  This target is deprecated. Use 'make db-migrate-up' instead."
	$(MAKE) db-migrate-up

.PHONY: migrate-down
migrate-down: ## Rollback database migrations (legacy - use db-migrate-down)
	@echo "⚠️  This target is deprecated. Use 'make db-migrate-down' instead."
	$(MAKE) db-migrate-down

.PHONY: db-reset
db-reset: db-rollback db-migrate ## Reset database (down + up) - Updated to use new targets

# Database variables (loaded from .env file at runtime)
DATABASE_USER ?= postgres
DATABASE_PASSWORD ?= postgres
DATABASE_HOST ?= postgres
DATABASE_PORT ?= 5432
DATABASE_NAME ?= service_db
DATABASE_SSL_MODE ?= disable
SERVICE_NAME ?= user-service
MIGRATION_IMAGE ?= migrate/migrate:latest
POSTGRES_NAME ?= postgres

# Database URL construction for targets
DATABASE_URL := postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DATABASE_HOST):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=$(DATABASE_SSL_MODE)

# ============================================================================
# 🗄️  DATABASE MANAGEMENT TARGETS
# ============================================================================

## Database Connection & Access
.PHONY: db-connect
db-connect: ## Connect to database shell
	@echo "🔌 Connecting to database..."
	@docker-compose --env-file .env exec postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME)

.PHONY: db-status
db-status: ## Show database status and connections
	@echo "📊 Database Status:"
	@docker-compose --env-file .env exec postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) -c "SELECT version();" 2>/dev/null || echo "❌ Database not accessible"
	@docker-compose --env-file .env exec postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) -c "SELECT count(*) as active_connections FROM pg_stat_activity;" 2>/dev/null || echo "❌ Cannot query connections"

.PHONY: db-health
db-health: ## Check database health and connectivity
	@echo "🏥 Database Health Check:"
	@docker-compose --env-file .env exec postgres pg_isready -U $(DATABASE_USER) -d $(DATABASE_NAME) -h $(DATABASE_HOST) -p $(DATABASE_PORT)
	@if [ $$? -eq 0 ]; then \
		echo "✅ Database is healthy and accepting connections"; \
	else \
		echo "❌ Database health check failed"; \
	fi

## Database Management
.PHONY: db-create
db-create: ## Create database if it doesn't exist
	@echo "🆕 Creating database $(DATABASE_NAME)..."
	@docker-compose --env-file .env exec postgres psql -U postgres -c "CREATE DATABASE $(DATABASE_NAME);" 2>/dev/null || echo "ℹ️  Database $(DATABASE_NAME) already exists or creation failed"

.PHONY: db-drop
db-drop: ## Drop database (with confirmation)
	@echo "💥 WARNING: This will drop database $(DATABASE_NAME) and ALL its data!"
	@echo "This action cannot be undone."
	@echo ""
	@read -p "Are you sure you want to drop the database? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		docker-compose --env-file .env exec postgres psql -U postgres -c "DROP DATABASE IF EXISTS $(DATABASE_NAME);"; \
		echo "✅ Database $(DATABASE_NAME) dropped successfully"; \
	else \
		echo "❌ Database drop cancelled"; \
	fi

.PHONY: db-recreate
db-recreate: db-drop db-create ## Recreate database from scratch
	@echo "🔄 Database $(DATABASE_NAME) recreated successfully"

## Migration Management (Docker-Based)
.PHONY: db-migrate-up
db-migrate-up: ## Run migrations up using migration container
	@echo "📈 Running migrations up..."
	@docker run --rm --network service-boilerplate-network \
		-v $(PWD)/services/$(SERVICE_NAME)/migrations:/migrations \
		$(MIGRATION_IMAGE) \
		-path /migrations \
		-database "postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(POSTGRES_NAME):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=$(DATABASE_SSL_MODE)" \
		up

.PHONY: db-migrate-down
db-migrate-down: ## Run migrations down using migration container
	@echo "⏪ Running migrations down..."
	@docker run --rm --network service-boilerplate-network \
		-v $(PWD)/services/$(SERVICE_NAME)/migrations:/migrations \
		$(MIGRATION_IMAGE) \
		-path /migrations \
		-database "postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(POSTGRES_NAME):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=$(DATABASE_SSL_MODE)" \
		down 1

.PHONY: db-migrate-status
db-migrate-status: ## Show migration status using migration container
	@echo "📋 Migration status:"
	@docker run --rm --network service-boilerplate-network \
		-v $(PWD)/services/$(SERVICE_NAME)/migrations:/migrations \
		$(MIGRATION_IMAGE) \
		-path /migrations \
		-database "postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(POSTGRES_NAME):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=$(DATABASE_SSL_MODE)" \
		version

.PHONY: db-migrate
db-migrate: db-migrate-up ## Run all pending migrations (alias for db-migrate-up)

.PHONY: db-rollback
db-rollback: db-migrate-down ## Rollback last migration (alias for db-migrate-down)

.PHONY: db-migrate-goto
db-migrate-goto: ## Go to specific migration version (VERSION=001)
	@echo "🎯 Going to migration version $(VERSION)..."
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Error: Please specify VERSION (e.g., make db-migrate-goto VERSION=001)"; \
		exit 1; \
	fi
	@docker run --rm --network service-boilerplate-network \
		-v $(PWD)/services/$(SERVICE_NAME)/migrations:/migrations \
		$(MIGRATION_IMAGE) \
		-path /migrations \
		-database "postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(POSTGRES_NAME):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=$(DATABASE_SSL_MODE)" \
		goto $(VERSION)

.PHONY: db-migrate-validate
db-migrate-validate: ## Validate migration files
	@echo "✅ Validating migration files..."
	@docker run --rm --network service-boilerplate-network \
		-v $(PWD)/services/$(SERVICE_NAME)/migrations:/migrations \
		$(MIGRATION_IMAGE) \
		-path /migrations \
		-database "postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(POSTGRES_NAME):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=$(DATABASE_SSL_MODE)" \
		up --dry-run

.PHONY: db-migration-create
db-migration-create: ## Create new migration file (NAME=add_users_table)
	@echo "📝 Creating migration: $(NAME)"
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Error: Please specify NAME (e.g., make db-migration-create NAME=add_users_table)"; \
		exit 1; \
	fi
	@docker run --rm \
		-v $(PWD)/services/$(SERVICE_NAME)/migrations:/migrations \
		$(MIGRATION_IMAGE) \
		create -ext sql -dir /migrations -seq $(NAME)

.PHONY: db-migration-list
db-migration-list: ## List available migration files
	@echo "📁 Migration files in services/$(SERVICE_NAME)/migrations:"
	@ls -la services/$(SERVICE_NAME)/migrations/ 2>/dev/null || echo "No migration files found"

## Data Management
.PHONY: db-seed
db-seed: ## Seed database with test data
	@echo "🌱 Seeding database with test data..."
	@docker-compose --env-file .env exec postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) -c "
	INSERT INTO users (email, first_name, last_name) VALUES
		('john.doe@example.com', 'John', 'Doe'),
		('jane.smith@example.com', 'Jane', 'Smith'),
		('bob.wilson@example.com', 'Bob', 'Wilson'),
		('alice.johnson@example.com', 'Alice', 'Johnson'),
		('charlie.brown@example.com', 'Charlie', 'Brown')
	ON CONFLICT (email) DO NOTHING;" 2>/dev/null
	@if [ $$? -eq 0 ]; then \
		echo "✅ Database seeded with 5 test users"; \
	else \
		echo "❌ Database seeding failed"; \
	fi

.PHONY: db-dump
db-dump: ## Create database dump
	@echo "💾 Creating database dump..."
	@docker-compose --env-file .env exec postgres pg_dump -U $(DATABASE_USER) -d $(DATABASE_NAME) --no-owner --no-privileges > db_dump_$(shell date +%Y%m%d_%H%M%S).sql 2>/dev/null
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
	@docker-compose --env-file .env exec -T postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) < $(FILE) 2>/dev/null
	@if [ $$? -eq 0 ]; then \
		echo "✅ Database restored from $(FILE)"; \
	else \
		echo "❌ Database restore failed"; \
	fi

.PHONY: db-clean
db-clean: ## Clean all data from tables (keep schema)
	@echo "🧹 Cleaning database data..."
	@docker-compose --env-file .env exec postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) -c "
	TRUNCATE TABLE users RESTART IDENTITY CASCADE;" 2>/dev/null
	@if [ $$? -eq 0 ]; then \
		echo "✅ Database data cleaned (schema preserved)"; \
	else \
		echo "❌ Database cleanup failed"; \
	fi

## Schema Inspection
.PHONY: db-schema
db-schema: ## Show database schema and tables
	@echo "📋 Database Schema:"
	@docker-compose --env-file .env exec postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) -c "\dt" 2>/dev/null || echo "❌ Cannot access database schema"
	@echo ""
	@echo "📊 Indexes:"
	@docker-compose --env-file .env exec postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) -c "\di" 2>/dev/null || echo "❌ Cannot access indexes"

.PHONY: db-tables
db-tables: ## List all tables and their structure
	@echo "📊 Table Structures:"
	@docker-compose --env-file .env exec postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) -c "
	SELECT table_name FROM information_schema.tables
	WHERE table_schema = 'public'
	ORDER BY table_name;" 2>/dev/null || echo "❌ Cannot list tables"
	@echo ""
	@if docker-compose --env-file .env exec postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) -c "\d users" 2>/dev/null; then \
		echo "✅ Users table structure displayed above"; \
	else \
		echo "⚠️  Users table not found or cannot display structure"; \
	fi

.PHONY: db-counts
db-counts: ## Show row counts and sizes for all tables
	@echo "🔢 Table Statistics:"
	@docker-compose --env-file .env exec postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME) -c "
	SELECT
		schemaname as schema,
		tablename as table,
		pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size,
		n_tup_ins - n_tup_del as row_count
	FROM pg_stat_user_tables
	WHERE schemaname = 'public'
	ORDER BY n_tup_ins - n_tup_del DESC;" 2>/dev/null || echo "❌ Cannot query table statistics"

## Development Workflow Targets
.PHONY: db-setup
db-setup: db-create db-migrate db-seed ## Complete database setup for development
	@echo "🎉 Database setup complete!"
	@echo "   Database: $(DATABASE_NAME)"
	@echo "   Tables: users"
	@echo "   Test data: 5 users created"
	@echo "   Status: Ready for development"

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
clean-docker: ## Clean project Docker containers, images, and networks
	@echo "🐳 Cleaning project Docker artifacts..."
	@docker-compose --env-file .env down --volumes --remove-orphans 2>/dev/null || true
	@docker rm $(API_GATEWAY_CONTAINER) 2>/dev/null || true
	@docker rm $(USER_SERVICE_CONTAINER) 2>/dev/null || true
	@docker rm $(POSTGRES_CONTAINER) 2>/dev/null || true
	@docker rmi $(API_GATEWAY_IMAGE) 2>/dev/null || true
	@docker rmi $(USER_SERVICE_IMAGE) 2>/dev/null || true
	@docker volume rm $(POSTGRES_VOLUME) 2>/dev/null || true
	@docker volume rm $(API_GATEWAY_TMP_VOLUME) 2>/dev/null || true
	@docker volume rm $(USER_SERVICE_TMP_VOLUME) 2>/dev/null || true
	@docker network rm $(NETWORK_NAME) 2>/dev/null || true
	@echo "✅ Project Docker artifacts cleaned"

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

	# Create volume directories
	@echo "   Creating volume directories..."
	@mkdir -p docker/volumes/postgres_data
	@mkdir -p docker/volumes/api-gateway/tmp
	@mkdir -p docker/volumes/user-service/tmp

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

.PHONY: help-docker
help-docker: ## Show Docker management commands
	@echo "🐳 Docker Management Commands:"
	@echo "  docker-reset           - Complete project Docker environment reset"
	@echo "  docker-reset-confirm   - Reset with confirmation prompt"
	@echo "  docker-recreate        - Recreate project Docker environment"
	@echo "  clean-docker           - Clean project Docker artifacts"

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
	@echo "Migration Management:"
	@echo "  db-migrate-up      - Run migrations up"
	@echo "  db-migrate-down    - Run migrations down"
	@echo "  db-migrate-status  - Show migration status"
	@echo "  db-migrate-goto VERSION= - Go to specific version"
	@echo "  db-migrate-validate - Validate migration files"
	@echo "  db-migration-create NAME= - Create migration file"
	@echo "  db-migration-list  - List migration files"
	@echo ""
	@echo "Data Management:"
	@echo "  db-seed            - Seed database with test data"
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
	@echo "Examples:"
	@echo "  make db-setup                    # Setup database for development"
	@echo "  make db-connect                  # Open database shell"
	@echo "  make db-migrate-up               # Run migrations"
	@echo "  make db-migration-create NAME=add_users_table"
	@echo "  make db-migrate-goto VERSION=001 # Go to specific version"
	@echo "  make db-dump                     # Create backup"
	@echo "  make db-restore FILE=dump.sql    # Restore from backup"
	@echo "  make db-seed                     # Add test data"
