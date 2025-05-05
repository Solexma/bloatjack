.PHONY: all build test clean deps lint dist fmt fmt-check install-tools test-examples

# Variables
BINARY_NAME=bloatjack
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"
DIST_DIR=dist
PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64
GOBIN=$(shell go env GOPATH)/bin
PATH:=$(GOBIN):$(PATH)
SHELL:=/bin/bash

all: deps test build

deps: install-tools
	go mod download
	go mod tidy

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./

test:
	@echo "Running tests and generating coverage report..."
	go test -v -coverprofile=coverage.out ./...
	@echo "\nCoverage Summary:"
	@go tool cover -func=coverage.out

# Test example configurations
test-examples:
	@echo "Testing example Docker Compose configurations..."
	@echo "1. Starting microservices stack..."
	@docker-compose -f examples/microservices-stack.yml up -d
	@echo "2. Running Bloatjack scan..."
	@./bin/$(BINARY_NAME) scan || true
	@echo "3. Running Bloatjack tune (dry-run)..."
	@./bin/$(BINARY_NAME) tune --dry-run || true
	@echo "4. Cleaning up..."
	@docker-compose -f examples/microservices-stack.yml down
	@echo "Example tests completed!"

clean:
	rm -rf bin/ $(DIST_DIR)/
	go clean

lint:
	golangci-lint run

# Tool installation
install-tools:
	@echo "Installing development tools..."
	@if ! command -v goimports &> /dev/null; then \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@echo "Development tools installed successfully"

# Formatting helpers
fmt: install-tools
	@echo "Formatting code..."
	@go fmt ./...
	@gofmt -s -w .
	@$(GOBIN)/goimports -w .
	@echo "Code formatting completed"

fmt-check: install-tools
	@echo "Checking code format..."
	@test -z $$(gofmt -l .)
	@test -z $$($(GOBIN)/goimports -l .)
	@echo "Code format check passed"

# Development helpers
dev:
	@if [ "$(ARGS)" = "" ]; then \
		go run main.go; \
	else \
		go run main.go $(ARGS); \
	fi

# Alias for dev with rules command
rules:
	@make dev ARGS="rules"

# Alias for dev with scan command
scan:
	@make dev ARGS="scan"

# Alias for dev with tune command
tune:
	@make dev ARGS="tune"

# Alias for dev with ui command
ui:
	@make dev ARGS="ui"

# Distribution helpers
dist: clean
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		output="$(DIST_DIR)/$(BINARY_NAME)-$$os-$$arch"; \
		if [ "$$os" = "windows" ]; then \
			output="$$output.exe"; \
		fi; \
		echo "Building $$output"; \
		GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $$output ./; \
		chmod +x $$output; \
	done

# Release helpers
release: dist
	@echo "Creating release for version $(VERSION)"
	@echo "Please ensure you have:"
	@echo "1. Updated CHANGELOG.md"
	@echo "2. Created and pushed a git tag"
	@echo "3. Built binaries for all platforms"
	@echo "4. Created a GitHub release" 