.PHONY: build run clean install test

# Application name
APP_NAME := gossh
VERSION := 0.1.0

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GORUN := $(GOCMD) run
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Build flags
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

# Output directory
OUT_DIR := bin

# Detect OS for correct binary extension
ifeq ($(OS),Windows_NT)
    BINARY_EXT := .exe
else
    BINARY_EXT :=
endif

BINARY_NAME := $(APP_NAME)$(BINARY_EXT)

# Default target
all: build

# Build the application
build:
	@mkdir -p $(OUT_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(OUT_DIR)/$(BINARY_NAME) .

# Run the application
run:
	$(GORUN) .

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(OUT_DIR)

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run tests
test:
	$(GOTEST) -v ./...

# Install the application to GOPATH/bin
install:
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) .

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@mkdir -p $(OUT_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(OUT_DIR)/$(APP_NAME)-linux-amd64 .

build-darwin:
	@mkdir -p $(OUT_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(OUT_DIR)/$(APP_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(OUT_DIR)/$(APP_NAME)-darwin-arm64 .

build-windows:
	@mkdir -p $(OUT_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(OUT_DIR)/$(APP_NAME)-windows-amd64.exe .

# Format code
fmt:
	$(GOCMD) fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Help
help:
	@echo "Available targets:"
	@echo "  build       - Build the application"
	@echo "  run         - Run the application"
	@echo "  clean       - Clean build artifacts"
	@echo "  deps        - Download and tidy dependencies"
	@echo "  test        - Run tests"
	@echo "  install     - Install to GOPATH/bin"
	@echo "  build-all   - Build for all platforms"
	@echo "  fmt         - Format code"
	@echo "  lint        - Lint code"
	@echo "  help        - Show this help"
