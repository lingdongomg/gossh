.PHONY: build run clean install test coverage test-verbose

# Application name
APP_NAME := gossh
VERSION := 1.2.0

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

# Run tests with coverage
coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with verbose output
test-verbose:
	$(GOTEST) -v -race ./...

# Install the application
# Priority: GOBIN > GOPATH/bin > ~/.local/bin (Linux/macOS) or ~/bin (fallback)
install:
ifeq ($(OS),Windows_NT)
	@if defined GOBIN ( \
		$(GOBUILD) $(LDFLAGS) -o $(GOBIN)\$(BINARY_NAME) . \
	) else if defined GOPATH ( \
		$(GOBUILD) $(LDFLAGS) -o $(GOPATH)\bin\$(BINARY_NAME) . \
	) else ( \
		$(GOBUILD) $(LDFLAGS) -o $(USERPROFILE)\go\bin\$(BINARY_NAME) . \
	)
	@echo "Installed to Go bin directory. Make sure it's in your PATH."
else
	@if [ -n "$(GOBIN)" ]; then \
		$(GOBUILD) $(LDFLAGS) -o $(GOBIN)/$(BINARY_NAME) . && \
		echo "Installed to $(GOBIN)/$(BINARY_NAME)"; \
	elif [ -n "$(GOPATH)" ]; then \
		$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) . && \
		echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"; \
	else \
		mkdir -p $(HOME)/.local/bin && \
		$(GOBUILD) $(LDFLAGS) -o $(HOME)/.local/bin/$(BINARY_NAME) . && \
		echo "Installed to $(HOME)/.local/bin/$(BINARY_NAME)"; \
	fi
	@echo "Make sure the install directory is in your PATH."
endif

# Install to system-wide location (requires sudo on Linux/macOS)
install-system:
ifeq ($(OS),Windows_NT)
	@echo "On Windows, please run as Administrator and copy bin/$(BINARY_NAME) to a directory in PATH"
	@echo "Common locations: C:\\Windows\\System32 or create C:\\Tools and add to PATH"
else
	sudo cp $(OUT_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "Installed to /usr/local/bin/$(BINARY_NAME)"
endif

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
	@echo "  build          - Build the application to bin/"
	@echo "  run            - Run the application directly"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  test           - Run tests"
	@echo "  install        - Install to user bin (GOBIN > GOPATH/bin > ~/.local/bin)"
	@echo "  install-system - Install to /usr/local/bin (requires sudo)"
	@echo "  build-all      - Build for all platforms"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  help           - Show this help"
