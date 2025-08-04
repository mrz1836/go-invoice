# Common makefile commands & variables between projects
include .make/common.mk

# Common Golang makefile commands & variables between projects
include .make/go.mk

## Set default repository details if not provided
REPO_NAME  ?= go-invoice
REPO_OWNER ?= mrz1836

.PHONY: rebuild-global
rebuild-global: ## Rebuild app, clear local cache, install globally as go-invoice
	@echo "Downloading fresh dependencies..."
	@go mod download
	@echo "Building and installing globally as go-invoice..."
	@go build -o $$(go env GOPATH)/bin/go-invoice ./cmd/go-invoice
