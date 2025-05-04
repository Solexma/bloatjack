.PHONY: all build test clean deps lint dist fmt fmt-check install-tools

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
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/bloatjack

test:
	go test -v ./...

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
	go run cmd/bloatjack/main.go

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
		GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $$output ./cmd/bloatjack; \
	done

# Release helpers
release: dist
	@echo "Creating release for version $(VERSION)"
	@echo "Please ensure you have:"
	@echo "1. Updated CHANGELOG.md"
	@echo "2. Created and pushed a git tag"
	@echo "3. Built binaries for all platforms"
	@echo "4. Created a GitHub release" 