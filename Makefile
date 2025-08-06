# Common makefile commands & variables between projects
include .make/common.mk

# Common Golang makefile commands & variables between projects
include .make/go.mk

## Set default repository details if not provided
REPO_NAME  ?= go-invoice
REPO_OWNER ?= mrz1836

.PHONY: rebuild-global rebuild-mcp run-mcp
rebuild-global: ## Rebuild app, clear local cache, install globally as go-invoice, and build MCP server
	@echo "Step 1: Removing existing binaries..."
	@rm -f bin/go-invoice
	@rm -f bin/go-invoice-mcp
	@rm -f $$(go env GOPATH)/bin/go-invoice
	@rm -f $$(go env GOPATH)/bin/go-invoice-mcp
	@echo "Step 2: Verifying binaries are deleted..."
	@if [ -f bin/go-invoice ]; then echo "ERROR: bin/go-invoice still exists!" && exit 1; fi
	@if [ -f bin/go-invoice-mcp ]; then echo "ERROR: bin/go-invoice-mcp still exists!" && exit 1; fi
	@if [ -f $$(go env GOPATH)/bin/go-invoice ]; then echo "ERROR: global go-invoice still exists!" && exit 1; fi
	@if [ -f $$(go env GOPATH)/bin/go-invoice-mcp ]; then echo "ERROR: global go-invoice-mcp still exists!" && exit 1; fi
	@echo "✓ All binaries successfully deleted"
	@echo "Step 3: Cleaning Go build cache to ensure fresh template embedding..."
	#@go clean -modcache
	#@go clean -i ./...
	@echo "Step 4: Downloading fresh dependencies..."
	@go mod download
	@echo "Step 5: Building go-invoice with embedded templates..."
	@go build -o bin/go-invoice ./cmd/go-invoice
	@echo "Step 6: Installing go-invoice globally..."
	@cp -f bin/go-invoice $$(go env GOPATH)/bin/go-invoice
	@echo "Step 7: Building go-invoice-mcp..."
	@go build -o bin/go-invoice-mcp ./cmd/go-invoice-mcp
	@echo "Step 8: Installing go-invoice-mcp globally..."
	@cp -f bin/go-invoice-mcp $$(go env GOPATH)/bin/go-invoice-mcp
	@echo "✅ Complete rebuild successful - all binaries rebuilt with fresh embedded templates"

run-mcp: ## Run the MCP server with stdio transport
	@echo "Starting go-invoice MCP server..."
	@./bin/go-invoice-mcp --stdio

rebuild-run-mcp: rebuild-global run-mcp ## Rebuild and run the MCP server
