.PHONY: build clean test help

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: ## Build all services
	@echo "Building all services..."
	go build -o bin/gateway ./cmd/gateway
	go build -o bin/usermanager ./cmd/usermanager
	go build -o bin/dhtnode ./cmd/dhtnode
	go build -o bin/replicator ./cmd/replicator
	go build -o bin/migrator ./cmd/migrator

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

run-gateway: ## Run gateway service
	go run ./cmd/gateway

run-usermanager: ## Run usermanager service
	go run ./cmd/usermanager

run-dhtnode: ## Run dhtnode service
	go run ./cmd/dhtnode

run-replicator: ## Run replicator service
	go run ./cmd/replicator
