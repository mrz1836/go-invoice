# Claude Desktop Integration Guide

This guide explains how to integrate the go-invoice MCP server with Claude Desktop using HTTP transport.

## Overview

The go-invoice MCP server provides natural language invoice management capabilities directly within Claude Desktop. You can create invoices, manage clients, import timesheets, and generate documents using conversational commands.

## Prerequisites

- Claude Desktop installed
- go-invoice CLI installed and configured
- Go 1.21+ (for building the MCP server)

## Quick Setup

Run the automated setup command:

```bash
go-invoice config setup-claude --desktop
```

This will:
- Build the MCP server
- Configure Claude Desktop
- Test the HTTP transport
- Verify the installation

## Manual Setup

### 1. Build the MCP Server

```bash
cd /path/to/go-invoice
go build -o bin/go-invoice-mcp ./cmd/mcp-server
```

### 2. Configure Claude Desktop

Edit your Claude Desktop MCP configuration file:

**macOS**: `~/Library/Application Support/Claude/mcp_servers.json`

Add the go-invoice server configuration:

```json
{
  "mcpServers": {
    "go-invoice": {
      "command": "/path/to/go-invoice/bin/go-invoice-mcp",
      "args": [
        "--http",
        "--port", "0",
        "--config", "~/.go-invoice/mcp-config.json"
      ],
      "env": {
        "GO_INVOICE_HOME": "~/.go-invoice",
        "MCP_LOG_LEVEL": "info",
        "MCP_TRANSPORT": "http"
      }
    }
  }
}
```

### 3. Restart Claude Desktop

Restart Claude Desktop to load the new MCP server configuration.

## Configuration Files

### Main Configuration (`~/.go-invoice/mcp-config.json`)

The main configuration file controls server behavior:

```json
{
  "server": {
    "name": "go-invoice-mcp",
    "version": "1.0.0",
    "transport": {
      "autoDetect": true,
      "http": {
        "enabled": true,
        "host": "127.0.0.1",
        "port": 0,
        "readTimeout": "30s",
        "writeTimeout": "30s"
      }
    },
    "health": {
      "enabled": true,
      "interval": "30s",
      "checks": {
        "cli": true,
        "storage": true
      }
    }
  }
}
```

### Tool Categories Configuration

Tools are organized into categories for better UX:

```json
{
  "categories": {
    "invoice_management": {
      "name": "Invoice Management",
      "description": "Create, list, and manage invoices",
      "tools": ["invoice_create", "invoice_list", "invoice_show"]
    },
    "client_management": {
      "name": "Client Management",
      "description": "Manage client information",
      "tools": ["client_create", "client_list", "client_show"]
    }
  }
}
```

## Available Tools

### Invoice Management

- **invoice_create**: Create a new invoice
- **invoice_list**: List invoices with filtering
- **invoice_show**: Display invoice details
- **invoice_update**: Update invoice information

### Client Management

- **client_create**: Add a new client
- **client_list**: List all clients
- **client_show**: Display client details
- **client_update**: Update client information

### Data Import/Export

- **import_csv**: Import timesheet data from CSV
- **generate_html**: Generate HTML invoice
- **export_data**: Export invoice data

### Configuration

- **show_config**: Display current configuration
- **update_config**: Modify configuration settings

## Usage Examples

### Creating an Invoice

```
Create an invoice for Acme Corporation for consulting work done in January 2025.
Use 40 hours at $150/hour.
```

### Importing Timesheet Data

```
Import the CSV file at ~/Documents/january-timesheet.csv and create invoices
for each client automatically.
```

### Generating Documents

```
Generate an HTML invoice for invoice INV-2025-001 and save it to the
~/Documents/invoices/ folder.
```

### Managing Clients

```
Add a new client: TechCorp Solutions
Address: 123 Tech Street, San Francisco, CA 94105
Email: billing@techcorp.com
Default rate: $175/hour
```

## Health Monitoring

The MCP server includes built-in health monitoring:

- **CLI Availability**: Verifies go-invoice CLI is accessible
- **Storage Access**: Checks data directory permissions
- **Memory Usage**: Monitors memory consumption
- **Performance**: Tracks response times

Health status is available through Claude Desktop's MCP status indicators.

## Logging

Logs are written to `~/.go-invoice/logs/mcp-server.log` by default.

Log levels:
- `debug`: Detailed operation logs
- `info`: General information (default)
- `warn`: Warning messages
- `error`: Error messages

Configure logging in the main configuration file:

```json
{
  "logging": {
    "level": "info",
    "format": "json",
    "outputs": [
      {
        "type": "file",
        "path": "~/.go-invoice/mcp-server.log"
      }
    ]
  }
}
```

## Security

### Sandboxing

The MCP server operates in a security sandbox:

- **Command Restrictions**: Only go-invoice commands are allowed
- **Path Restrictions**: Access limited to invoice directories
- **File Size Limits**: Maximum file size enforced
- **Execution Timeouts**: Commands have time limits

### Audit Logging

All operations are logged for audit purposes:

```json
{
  "timestamp": "2025-01-20T10:30:00Z",
  "user": "desktop-user",
  "action": "invoice_create",
  "client": "Acme Corp",
  "result": "success"
}
```

## Troubleshooting

### Server Not Starting

1. Check if go-invoice CLI is accessible:
   ```bash
   go-invoice version
   ```

2. Verify configuration file syntax:
   ```bash
   jq . ~/.go-invoice/mcp-config.json
   ```

3. Check permissions on data directory:
   ```bash
   ls -la ~/.go-invoice/
   ```

### Tools Not Available

1. Restart Claude Desktop completely
2. Check MCP server logs for errors
3. Verify HTTP transport is listening:
   ```bash
   lsof -i :8080  # Or configured port
   ```

### Performance Issues

1. Check system resources:
   ```bash
   top -p $(pgrep go-invoice-mcp)
   ```

2. Review performance logs in health monitoring
3. Adjust concurrency settings in configuration

## Advanced Configuration

### Custom Tool Categories

Add custom tool groupings:

```json
{
  "tools": {
    "categories": [
      "custom_reporting",
      "automation"
    ]
  }
}
```

### Performance Tuning

Optimize for your usage patterns:

```json
{
  "performance": {
    "caching": {
      "enabled": true,
      "ttl": "10m"
    },
    "concurrency": {
      "maxConcurrentTools": 10
    }
  }
}
```

### Integration with External Systems

Configure webhooks for external notifications:

```json
{
  "hooks": {
    "invoice_created": {
      "url": "https://api.example.com/webhooks/invoice",
      "method": "POST"
    }
  }
}
```

## Support

- **Documentation**: [go-invoice docs](https://github.com/mrz1836/go-invoice)
- **Issues**: [GitHub Issues](https://github.com/mrz1836/go-invoice/issues)
- **Logs**: Check `~/.go-invoice/logs/` for detailed error information
