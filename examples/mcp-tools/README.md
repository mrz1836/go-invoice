# go-invoice MCP Tools Examples

This directory contains setup and configuration examples for using go-invoice with Claude Desktop through the Model Context Protocol (MCP).

## Files in this Directory

- `claude-desktop-config.json` - Complete Claude Desktop configuration
- `tool-usage-examples.md` - Natural language interaction examples
- `workflow-automation.md` - Advanced automation patterns

## Quick Setup Guide

### 1. Install go-invoice MCP Server

```bash
# Build the MCP server
cd cmd/go-invoice-mcp
go build -o go-invoice-mcp

# Make it executable and move to your preferred location
chmod +x go-invoice-mcp
mv go-invoice-mcp /usr/local/bin/
```

### 2. Configure Claude Desktop

Copy the configuration from `claude-desktop-config.json` to your Claude Desktop configuration file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
**Linux**: `~/.config/Claude/claude_desktop_config.json`

### 3. Set Environment Variables

```bash
# Add to your shell profile (.bashrc, .zshrc, etc.)
export GO_INVOICE_DATA_DIR="/path/to/your/invoice-data"
export GO_INVOICE_CONFIG_FILE="/path/to/your/config.json"
```

### 4. Start Claude Desktop

Restart Claude Desktop after making the configuration changes. The go-invoice tools should now be available.

## Verification

To verify the setup works, ask Claude:

> "What invoice management tools are available?"

Claude should respond with information about the 21 available go-invoice MCP tools across 5 categories.

## Getting Help

- See `tool-usage-examples.md` for conversational examples
- See `workflow-automation.md` for advanced automation patterns
- Refer to the main documentation in `/docs/` for complete tool reference