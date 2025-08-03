# Claude Desktop Setup Guide

This guide will help you set up go-invoice MCP integration with Claude Desktop using HTTP transport. Claude Desktop provides a visual interface that's perfect for business users who want to manage invoices through natural conversation.

## Overview

Claude Desktop uses HTTP transport to communicate with the go-invoice MCP server, providing:
- **Visual Interface**: Point-and-click invoice management
- **Real-time Updates**: Immediate feedback on operations
- **Document Preview**: Visual invoice generation and preview
- **Secure Communication**: HTTP-based encrypted communication

## Prerequisites

### System Requirements
- **Operating System**: Windows 10+, macOS 10.15+, or Linux
- **Claude Desktop**: Version 1.0.0 or later
- **go-invoice**: Latest version installed
- **Network Access**: Local HTTP server capability

### Installation Check
Verify go-invoice is installed:
```bash
go-invoice --version
go-invoice-mcp --help
```

If not installed, download from the [releases page](https://github.com/mrz1836/go-invoice/releases).

## Step 1: Configuration Directory Setup

Create your configuration directory:

```bash
# Create configuration directory
mkdir -p ~/.go-invoice

# Set appropriate permissions
chmod 750 ~/.go-invoice
```

## Step 2: Business Configuration

Create your business configuration file at `~/.go-invoice/.env.config`:

```bash
# Business Information
BUSINESS_NAME="Your Business Name"
BUSINESS_ADDRESS="123 Business St, City, State 12345"
BUSINESS_EMAIL="billing@yourbusiness.com"
BUSINESS_PHONE="+1 (555) 123-4567"
BUSINESS_WEBSITE="https://yourbusiness.com"

# Payment Information
PAYMENT_TERMS="Net 30"
BANK_NAME="Your Bank Name"
BANK_ACCOUNT="****1234"
BANK_ROUTING="123456789"

# Invoice Settings
INVOICE_PREFIX="INV"
INVOICE_START_NUMBER=1000
CURRENCY="USD"
VAT_RATE=0.0

# Storage Settings
DATA_DIR="~/.go-invoice/data"
AUTO_BACKUP=true
BACKUP_INTERVAL="24h"
```

### Configuration Options

| Setting | Description | Example |
|---------|-------------|---------|
| `BUSINESS_NAME` | Your company name | "Acme Consulting LLC" |
| `BUSINESS_ADDRESS` | Full business address | "123 Main St, City, ST 12345" |
| `BUSINESS_EMAIL` | Contact email for invoices | "billing@company.com" |
| `PAYMENT_TERMS` | Default payment terms | "Net 30", "Due on receipt" |
| `CURRENCY` | Default currency code | "USD", "EUR", "GBP" |
| `VAT_RATE` | Tax rate as decimal | 0.08 for 8%, 0.0 for none |

## Step 3: MCP Server Configuration

Create the MCP server configuration at `~/.go-invoice/mcp-config.json`:

```json
{
  "server": {
    "name": "go-invoice-mcp",
    "version": "1.0.0",
    "description": "Natural language invoice management"
  },
  "transport": {
    "type": "http",
    "host": "localhost",
    "port": 0,
    "readTimeout": "30s",
    "writeTimeout": "30s",
    "maxMessageSize": 10485760,
    "enableLogging": true,
    "logLevel": "info"
  },
  "security": {
    "enableSandbox": true,
    "allowedPaths": [
      "~/.go-invoice",
      "~/Documents/Invoices",
      "~/Downloads"
    ],
    "blockedPaths": [
      "/etc",
      "/var",
      "/usr",
      "/bin",
      "/sbin"
    ],
    "environmentWhitelist": [
      "HOME",
      "USER",
      "GO_INVOICE_HOME",
      "MCP_LOG_LEVEL",
      "MCP_TRANSPORT"
    ]
  },
  "logging": {
    "enabled": true,
    "level": "info",
    "file": "~/.go-invoice/mcp-server.log",
    "maxSize": "10MB",
    "maxBackups": 5,
    "enableConsole": false
  }
}
```

### Security Configuration

The security section ensures safe operation:

- **Sandbox Mode**: Restricts file system access
- **Allowed Paths**: Directories the server can access
- **Blocked Paths**: System directories to protect
- **Environment Whitelist**: Approved environment variables

## Step 4: Claude Desktop Configuration

Add the go-invoice MCP server to Claude Desktop's configuration:

### Windows Configuration
Edit `%APPDATA%\Claude\claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "go-invoice": {
      "name": "Go Invoice MCP Server",
      "description": "Natural language invoice management through go-invoice CLI",
      "command": "go-invoice-mcp",
      "args": [
        "--http",
        "--config",
        "~/.go-invoice/mcp-config.json",
        "--port",
        "0"
      ],
      "env": {
        "GO_INVOICE_HOME": "~/.go-invoice",
        "MCP_LOG_LEVEL": "info",
        "MCP_TRANSPORT": "http",
        "MCP_SERVER_NAME": "go-invoice-mcp"
      },
      "capabilities": {
        "tools": true,
        "resources": false,
        "prompts": false
      },
      "metadata": {
        "version": "1.0.0",
        "author": "go-invoice",
        "documentation": "https://github.com/mrz1836/go-invoice/docs/mcp/claude-desktop-setup.md"
      }
    }
  }
}
```

### macOS Configuration
Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "go-invoice": {
      "name": "Go Invoice MCP Server",
      "description": "Natural language invoice management through go-invoice CLI",
      "command": "go-invoice-mcp",
      "args": [
        "--http",
        "--config",
        "~/.go-invoice/mcp-config.json",
        "--port",
        "0"
      ],
      "env": {
        "GO_INVOICE_HOME": "~/.go-invoice",
        "MCP_LOG_LEVEL": "info",
        "MCP_TRANSPORT": "http"
      },
      "capabilities": {
        "tools": true,
        "resources": false,
        "prompts": false
      }
    }
  }
}
```

### Linux Configuration
Edit `~/.config/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "go-invoice": {
      "name": "Go Invoice MCP Server", 
      "description": "Natural language invoice management through go-invoice CLI",
      "command": "go-invoice-mcp",
      "args": [
        "--http",
        "--config",
        "~/.go-invoice/mcp-config.json",
        "--port",
        "0"
      ],
      "env": {
        "GO_INVOICE_HOME": "~/.go-invoice",
        "MCP_LOG_LEVEL": "info",
        "MCP_TRANSPORT": "http"
      },
      "capabilities": {
        "tools": true,
        "resources": false,
        "prompts": false
      }
    }
  }
}
```

## Step 5: Test the Configuration

### 1. Test MCP Server Standalone
```bash
# Test the MCP server directly
go-invoice-mcp --http --config ~/.go-invoice/mcp-config.json --port 8080

# In another terminal, test the health endpoint
curl http://localhost:8080/health
```

You should see a response like:
```json
{
  "status": "ok",
  "transport": "http",
  "ready": true,
  "messagesReceived": 0,
  "messagesSent": 0,
  "uptime": "5s"
}
```

### 2. Validate Configuration
```bash
# Validate your business configuration
go-invoice config validate

# Initialize sample data if needed
go-invoice config init
```

### 3. Test in Claude Desktop

1. **Start Claude Desktop**
2. **Verify MCP Connection**: Look for "go-invoice" in the available tools
3. **Test Basic Commands**:

```
"Show me my configuration"
```

```
"Create a test client called 'Test Company' with email test@example.com"
```

```
"List all my clients"
```

## Step 6: Usage Examples

### Creating Your First Invoice

```
"Create an invoice for Acme Corp for 40 hours of web development work at $85 per hour, due in 30 days"
```

Expected response:
```
I'll create an invoice for Acme Corp. Let me set this up with the details you provided:

- Client: Acme Corp
- Work: 40 hours of web development 
- Rate: $85/hour
- Total: $3,400
- Due date: 30 days from today

[Invoice created successfully with ID: INV-2024-001]
```

### Importing Timesheet Data

```
"Import the timesheet data from ~/Downloads/january-hours.csv and create invoices for each client"
```

### Generating Professional Invoices

```
"Generate an HTML invoice for INV-2024-001 and save it to ~/Documents/Invoices/"
```

## Advanced Configuration

### Custom Invoice Templates

1. Create templates directory:
```bash
mkdir -p ~/.go-invoice/templates
```

2. Add custom HTML template:
```html
<!-- ~/.go-invoice/templates/professional.html -->
<!DOCTYPE html>
<html>
<head>
    <title>Invoice {{.InvoiceNumber}}</title>
    <style>
        /* Your custom styling */
    </style>
</head>
<body>
    <!-- Your custom invoice layout -->
</body>
</html>
```

3. Use custom template:
```
"Generate an invoice using the professional template for INV-2024-001"
```

### Performance Optimization

For high-volume usage, adjust these settings in `mcp-config.json`:

```json
{
  "transport": {
    "readTimeout": "60s",
    "writeTimeout": "60s",
    "maxMessageSize": 52428800
  },
  "logging": {
    "level": "warn",
    "enableConsole": false
  }
}
```

### Backup Configuration

Enable automatic backups:

```bash
# In your .env.config file
AUTO_BACKUP=true
BACKUP_INTERVAL="6h"
RETENTION_DAYS=90
BACKUP_DIR="~/.go-invoice/backups"
```

## Security Best Practices

### 1. File System Security
- Keep invoice data in dedicated directory (`~/.go-invoice`)
- Set restrictive permissions: `chmod 750 ~/.go-invoice`
- Regularly backup your data directory

### 2. Network Security
- HTTP server binds only to localhost by default
- Consider using firewall rules if needed
- Monitor server logs for unusual activity

### 3. Configuration Security
- Protect configuration files: `chmod 640 ~/.go-invoice/mcp-config.json`
- Use environment variables for sensitive data
- Regularly rotate any API keys or credentials

### 4. Data Protection
- Enable automatic backups
- Test backup restoration regularly
- Consider encryption for sensitive client data

## Troubleshooting

### Common Issues

1. **Server Won't Start**
   ```bash
   # Check configuration syntax
   go-invoice-mcp --config ~/.go-invoice/mcp-config.json --validate
   
   # Check port availability
   netstat -an | grep :8080
   ```

2. **Claude Can't Connect**
   - Verify Claude Desktop configuration file syntax
   - Check that go-invoice-mcp is in PATH
   - Review Claude Desktop logs

3. **Permission Errors**
   ```bash
   # Fix directory permissions
   chmod -R 750 ~/.go-invoice
   
   # Check file ownership
   ls -la ~/.go-invoice
   ```

For more troubleshooting help, see the [Troubleshooting Guide](troubleshooting.md).

## Next Steps

1. **Explore Tools**: Try all 21 available tools
2. **Set Up Workflows**: Create automated invoice processes
3. **Customize Templates**: Design professional invoice layouts
4. **Monitor Usage**: Review logs and optimize performance

You're now ready to use natural language for professional invoice management with Claude Desktop!