#!/bin/bash
# Wrapper script for go-invoice MCP commands
# This is a workaround for MCP connection issues

# Set environment variables
export GO_INVOICE_HOME="/Users/mrz/.go-invoice"
export GO_INVOICE_PROJECT="/Users/mrz/projects/go-invoice"
export GO_INVOICE_CLI_PATH="/Users/mrz/projects/go-invoice/bin/go-invoice"
export GO_INVOICE_CONFIG_PATH="/Users/mrz/projects/go-invoice/.env.config"

# Function to show invoice
invoice_show() {
    local identifier="$1"
    "${GO_INVOICE_CLI_PATH}" invoice show "$identifier" --output json
}

# Function to list invoices
invoice_list() {
    "${GO_INVOICE_CLI_PATH}" invoice list --output json
}

# Main command handler
case "$1" in
    "invoice_show")
        invoice_show "$2"
        ;;
    "invoice_list")
        invoice_list
        ;;
    *)
        echo "Usage: $0 {invoice_show|invoice_list} [args]"
        exit 1
        ;;
esac