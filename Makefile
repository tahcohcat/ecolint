.PHONY: test test-verbose test-coverage test-race build clean lint install help

# Default target
help: ## Show this help message
	@echo "🌱 ecolint - Environment File Linter"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ { printf "  %-15s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# Test targets
test: ## Run all tests
	@echo "🧪 Running tests..."
	go test ./...

test-verbose: ## Run tests with verbose output
	@echo "🧪 Running tests (verbose)..."
	go test -v ./...

test-coverage: ## Run tests with coverage report
	@echo "🧪 Running tests with coverage..."
	go test -cover ./...
	@echo ""
	@echo "📊 Generating detailed coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

test-race: ## Run tests with race detection
	@echo "🧪 Running tests with race detection..."
	go test -race ./...

test-rules: ## Run only rules package tests
	@echo "🧪 Running rules tests..."
	go test -v ./rules

test-parse: ## Run only parse package tests
	@echo "🧪 Running parse tests..."
	go test -v ./parse

# Build targets
build: ## Build the binary
	@echo "🔨 Building ecolint..."
	go build -o ecolint cmd/ecolint/main.go
	@echo "✅ Binary built: ./ecolint"

build-all: ## Build binaries for all platforms
	@echo "🔨 Building for all platforms..."
	GOOS=linux GOARCH=amd64 go build -o dist/ecolint-linux-amd64 cmd/ecolint/main.go
	GOOS=darwin GOARCH=amd64 go build -o dist/ecolint-darwin-amd64 cmd/ecolint/main.go
	GOOS=darwin GOARCH=arm64 go build -o dist/ecolint-darwin-arm64 cmd/ecolint/main.go
	GOOS=windows GOARCH=amd64 go build -o dist/ecolint-windows-amd64.exe cmd/ecolint/main.go
	@echo "✅ All binaries built in dist/"

install: build ## Install the binary to $GOPATH/bin
	@echo "📦 Installing ecolint..."
	mv ecolint $(GOPATH)/bin/ecolint
	@echo "✅ ecolint installed to $(GOPATH)/bin/ecolint"

# Development targets
lint: ## Run golangci-lint
	@echo "🔍 Running linter..."
	golangci-lint run

fmt: ## Format code
	@echo "🎨 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	go vet ./...

tidy: ## Tidy up module dependencies
	@echo "🧹 Tidying dependencies..."
	go mod tidy

# Example targets
examples: build ## Create example files for testing
	@echo "📁 Creating example files..."
	@mkdir -p examples/env
	@cat > examples/env/okay.env << 'EOF'
# Good environment file example
DATABASE_URL=postgresql://user:pass@localhost:5432/mydb
API_KEY=your-api-key-here
PORT=8080
NODE_ENV=development
DEBUG=true
LOG_LEVEL=info
EOF
	@cat > examples/env/issues.env << 'EOF'
# Environment file with various issues
DATABASE_URL=postgresql://user:pass@localhost:5432/mydb
API_KEY=your-api-key-here
PORT=8080

# Duplicate variable (duplicate rule)
PORT=3000

# Another duplicate
API_KEY=different-key

# Empty value (empty_values rule)
EMPTY_VAR=

# Malformed line (syntax parsing)
MALFORMED LINE WITHOUT EQUALS

# Invalid key with spaces (convention rule)
INVALID KEY=value

# Security issue (security rule)
REAL_SECRET=sk_live_51H1234567890abcdef

# Convention violations
databasePassword=secret123
database-host=localhost
CONFIG=generic-config
DB_PWD=secret
PATH=/override/system/path
EOF
	@cat > .ecolint.yaml << 'EOF'
# ecolint configuration file
required_vars:
  - DATABASE_URL
  - API_KEY
  - PORT

rules:
  duplicate: true
  missing: true
  syntax: true
  empty_values: true
  security: true
  convention: true

output:
  format: "pretty"
  color: true
EOF
	@echo "✅ Example files created:"
	@echo "  - examples/env/okay.env (clean file)"
	@echo "  - examples/env/issues.env (file with issues)"
	@echo "  - .ecolint.yaml (sample config)"

test-examples: examples ## Test the example files
	@echo "🧪 Testing example files..."
	@echo ""
	@echo "Testing clean file (should show no issues):"
	./ecolint lint examples/env/okay.env || true
	@echo ""
	@echo "Testing problematic file (should show multiple issues):"
	./ecolint lint examples/env/issues.env || true

demo: examples ## Run a demonstration
	@echo "🎪 ecolint demonstration"
	@echo ""
	@echo "1. Testing a clean environment file:"
	@echo "   $ ecolint lint examples/env/okay.env"
	@./ecolint lint examples/env/okay.env || true
	@echo ""
	@echo "2. Testing a problematic environment file:"
	@echo "   $ ecolint lint examples/env/issues.env"
	@./ecolint lint examples/env/issues.env || true
	@echo ""
	@echo "3. Quick syntax check:"
	@echo "   $ ecolint check examples/env/okay.env"
	@./ecolint check examples/env/okay.env || true

# Cleanup targets
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	rm -f ecolint
	rm -f coverage.out coverage.html
	rm -rf dist/
	@echo "✅ Clean complete"

clean-examples: ## Clean example files
	@echo "🧹 Cleaning example files..."
	rm -rf examples/
	rm -f .ecolint.yaml
	@echo "✅ Example files cleaned"

# CI targets
ci-test: ## Run tests for CI
	@echo "🚀 Running CI tests..."
	go test -v -race -cover ./...

ci-lint: ## Run linter for CI (if available)
	@echo "🚀 Running CI linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed, skipping..."; \
		go vet ./...; \
	fi

ci: ci-test ci-lint ## Run all CI checks