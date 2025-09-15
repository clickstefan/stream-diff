# Stream-Diff Modern Makefile with 2025 Best Practices

# Build variables
BINARY_NAME = stream-diff
MAIN_PACKAGE = ./main.go
BUILD_DIR = build
COVERAGE_DIR = coverage

# Version info (can be overridden by CI)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILT_BY ?= $(shell whoami)

# Go build flags
LDFLAGS = -ldflags="-X 'data-comparator/cmd.version=$(VERSION)' \
                    -X 'data-comparator/cmd.commit=$(COMMIT)' \
                    -X 'data-comparator/cmd.date=$(DATE)' \
                    -X 'data-comparator/cmd.builtBy=$(BUILT_BY)'"

# Go tools
GOLANGCI_LINT_VERSION = v1.63.4
GOVULNCHECK_VERSION = latest

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

##@ Development

.PHONY: setup
setup: ## Install development dependencies
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/securecodewarrior/sast-scan@latest
	@echo "Development tools installed!"

.PHONY: deps
deps: ## Download and verify dependencies
	go mod download
	go mod verify
	go mod tidy

##@ Building

.PHONY: build
build: clean deps ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: build-all
build-all: clean deps ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)
	
	# macOS
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	
	# Windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	
	@echo "Multi-platform build complete!"

.PHONY: install
install: build ## Install the application
	go install $(LDFLAGS) $(MAIN_PACKAGE)

##@ Testing

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	go test -race -vet=off ./...

.PHONY: test-verbose
test-verbose: ## Run tests with verbose output
	@echo "Running tests with verbose output..."
	go test -v -race -vet=off ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	go test -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	go tool cover -func=$(COVERAGE_DIR)/coverage.out
	@echo "Coverage report generated: $(COVERAGE_DIR)/coverage.html"

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem -run=^$$ ./...

.PHONY: test-integration
test-integration: build ## Run integration tests
	@echo "Running integration tests..."
	./scripts/integration-tests.sh

##@ Code Quality

.PHONY: lint
lint: ## Run linters
	@echo "Running golangci-lint..."
	golangci-lint run --verbose

.PHONY: lint-fix
lint-fix: ## Run linters with auto-fix
	@echo "Running golangci-lint with auto-fix..."
	golangci-lint run --fix --verbose

.PHONY: format
format: ## Format code
	@echo "Formatting code..."
	gofmt -s -w .
	goimports -w .

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

.PHONY: staticcheck
staticcheck: ## Run staticcheck
	@echo "Running staticcheck..."
	staticcheck ./...

.PHONY: complexity
complexity: ## Check cyclomatic complexity
	@echo "Checking cyclomatic complexity..."
	gocyclo -over 15 .

##@ Security

.PHONY: security
security: ## Run security checks
	@echo "Running security checks..."
	gosec -quiet ./...

.PHONY: vuln-check
vuln-check: ## Check for known vulnerabilities
	@echo "Checking for vulnerabilities..."
	govulncheck ./...

.PHONY: mod-verify
mod-verify: ## Verify module dependencies
	@echo "Verifying module dependencies..."
	go mod verify

##@ Quality Gates

.PHONY: quality-check
quality-check: deps vet lint staticcheck security vuln-check test ## Run all quality checks

.PHONY: pre-commit
pre-commit: format quality-check ## Run pre-commit checks

.PHONY: ci-check
ci-check: quality-check test-coverage ## Run CI checks

##@ Documentation

.PHONY: docs
docs: ## Generate documentation
	@echo "Generating documentation..."
	go doc -all ./... > docs/api.txt
	@echo "Documentation generated in docs/"

##@ Cleanup

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)
	rm -rf $(COVERAGE_DIR)
	go clean -cache
	go clean -testcache

.PHONY: clean-all
clean-all: clean ## Clean everything including dependencies
	go clean -modcache

##@ Utilities

.PHONY: version
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"
	@echo "Built by: $(BUILT_BY)"

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

.PHONY: deps-audit
deps-audit: ## Audit dependencies for security issues
	@echo "Auditing dependencies..."
	go list -json -deps ./... | nancy sleuth

.PHONY: run
run: build ## Build and run the application
	./$(BUILD_DIR)/$(BINARY_NAME) --help

.PHONY: release
release: clean build-all test-coverage quality-check ## Prepare release artifacts
	@echo "Release preparation complete!"
	@echo "Artifacts in $(BUILD_DIR)/"
	@echo "Coverage report in $(COVERAGE_DIR)/"

# Default target
.DEFAULT_GOAL := help