# Common makefile commands & variables between projects
include .make/common.mk

# Common Golang makefile commands & variables between projects
include .make/go.mk

## Set default repository details if not provided
REPO_NAME  ?= go-invoice
REPO_OWNER ?= mrz1836

.PHONY: rebuild-global rebuild-mcp run-mcp
rebuild-global: ## Rebuild app, clear local cache, install globally as go-invoice, and build MCP server
	@echo "Downloading fresh dependencies..."
	@go mod download
	@echo "Building and installing globally as go-invoice..."
	@go build -o $$(go env GOPATH)/bin/go-invoice ./cmd/go-invoice
	@echo "Force rebuilding MCP server..."
	@rm -f bin/go-invoice-mcp
	@go build -o bin/go-invoice-mcp ./cmd/go-invoice-mcp
	@echo "Installing MCP server globally..."
	@cp bin/go-invoice-mcp $$(go env GOPATH)/bin/go-invoice-mcp
	@echo "MCP server rebuilt and installed successfully"

run-mcp: ## Run the MCP server with stdio transport
	@echo "Starting go-invoice MCP server..."
	@./bin/go-invoice-mcp --stdio

rebuild-run-mcp: rebuild-global run-mcp ## Rebuild and run the MCP server
