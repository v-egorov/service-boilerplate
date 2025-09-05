# Service Boilerplate Makefile

# Project variables
PROJECT_NAME := service-boilerplate
API_GATEWAY_DIR := api-gateway
USER_SERVICE_DIR := services/user-service
BUILD_DIR := build
DOCKER_COMPOSE_FILE := docker/docker-compose.yml

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

.PHONY: run
run: run-gateway run-user-service ## Run all services locally

.PHONY: run-gateway
run-gateway: ## Run API Gateway
	@echo "Running API Gateway..."
	@cd $(API_GATEWAY_DIR) && $(GO) run ./cmd

.PHONY: run-user-service
run-user-service: ## Run User Service
	@echo "Running User Service..."
	@cd $(USER_SERVICE_DIR) && $(GO) run ./cmd

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
	@$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) build

.PHONY: docker-run
docker-run: ## Run services with Docker Compose
	@echo "Starting services with Docker..."
	@$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) up -d

.PHONY: docker-stop
docker-stop: ## Stop all Docker containers
	@echo "Stopping Docker containers..."
	@$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) down

.PHONY: docker-logs
docker-logs: ## Show Docker container logs
	@echo "Showing container logs..."
	@$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) logs -f

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