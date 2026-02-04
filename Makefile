.PHONY: help build run test clean migrate-up migrate-down db-create

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building..."
	go build -o customer-support-api ./cmd/api

run: ## Run the application
	@echo "Running..."
	go run ./cmd/api/main.go

dev: ## Run with live reload (requires air: go install github.com/cosmtrek/air@latest)
	@echo "Running with live reload..."
	air

test: ## Run all tests
	@echo "Running tests..."
	go test ./... -v

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	go test ./tests/unit/... -v

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	go test ./tests/integration/... -v

test-e2e: ## Run end-to-end tests
	@echo "Running E2E tests..."
	go test ./tests/e2e/... -v

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f customer-support-api
	rm -f coverage.out coverage.html
	rm -f data/*.db

db-create: ## Create and initialize the database
	@echo "Creating database..."
	@mkdir -p data
	@if [ -f data/chatbot.db ]; then \
		echo "Database already exists. Use 'make clean' first to recreate."; \
	else \
		sqlite3 data/chatbot.db < migrations/001_sqlite_schema.sql; \
		echo "Database created successfully!"; \
	fi

db-reset: ## Reset the database (WARNING: destroys all data)
	@echo "Resetting database..."
	@rm -f data/chatbot.db
	@make db-create

migrate-up: db-create ## Run database migrations
	@echo "Migrations complete!"

migrate-down: ## Drop all tables (WARNING: destroys all data)
	@echo "Dropping all tables..."
	@rm -f data/chatbot.db
	@echo "Database dropped!"

deps: ## Install dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

lint: ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	golangci-lint run

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t customer-support-api:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env customer-support-api:latest
