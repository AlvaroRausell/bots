.PHONY: build test clean install help dev

GOBIN ?= $(HOME)/.local/bin

# Build the bots CLI
build:
	mkdir -p dist
	go build -o dist/bots ./cmd/bots

# Run tests
test:
	go test ./...

# Install to system PATH
install:
	mkdir -p "$(GOBIN)"
	GOBIN="$(GOBIN)" go install ./cmd/bots

# Clean build artifacts
clean:
	rm -rf dist
	go clean

# Run the installer TUI
install-mcp: build
	./dist/bots install

# Start MCP server (for testing)
mcp: build
	./dist/bots mcp serve

# Development mode - rebuild and run
dev: build
	./dist/bots help

# Create .bots directory structure
init:
	mkdir -p .bots/logs .bots/tasks .bots/skills/session-persistence

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	go vet ./...

# Full check - format, lint, test, build
check: fmt lint test build

help:
	@echo "Bots CLI - Session Persistence & Decision Tracking"
	@echo ""
	@echo "Usage:"
	@echo "  make build       Build the bots CLI"
	@echo "  make test        Run tests"
	@echo "  make install     Install to $(GOBIN)"
	@echo "  make clean       Remove build artifacts"
	@echo "  make install-mcp Run interactive MCP installer"
	@echo "  make mcp         Start MCP server"
	@echo "  make init        Create .bots directory structure"
	@echo "  make fmt         Format Go code"
	@echo "  make lint        Lint Go code"
	@echo "  make check       Run fmt, lint, test, build"
	@echo "  make dev         Build and run help"
	@echo "  make help        Show this help message"
