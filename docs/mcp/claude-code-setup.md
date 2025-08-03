# Claude Code Setup Guide

This guide will help you set up go-invoice MCP integration with Claude Code using stdio transport. Claude Code provides powerful terminal integration that's perfect for developers and power users who want to manage invoices through command-line interfaces and automation.

## Overview

Claude Code uses stdio transport to communicate with the go-invoice MCP server, providing:
- **Terminal Integration**: Direct command-line access
- **Resource References**: Advanced file and data references
- **Slash Commands**: Quick shortcuts for common operations
- **Script Integration**: Perfect for automation and CI/CD
- **Project Scope**: Deep integration with your development workflow

## Prerequisites

### System Requirements
- **Operating System**: Windows 10+, macOS 10.15+, or Linux
- **Claude Code**: Latest version installed
- **go-invoice**: Latest version with MCP support
- **Terminal Access**: Command-line interface

### Installation Check
Verify installations:
```bash
# Check go-invoice installation
go-invoice --version
go-invoice-mcp --help

# Check Claude Code installation
claude --version
claude mcp list
```

If go-invoice isn't installed:
```bash
# Download and install latest release
curl -L https://github.com/mrz1836/go-invoice/releases/latest/download/go-invoice-linux-amd64 -o go-invoice
chmod +x go-invoice
sudo mv go-invoice /usr/local/bin/

# Or use package manager
# brew install go-invoice  # macOS
# apt install go-invoice   # Debian/Ubuntu
```

## Step 1: Project Setup

### Create Invoice Project Directory
```bash
# Create project directory
mkdir -p ~/invoice-project
cd ~/invoice-project

# Initialize project structure
mkdir -p {data,templates,exports,imports,backups}

# Set up configuration directory
mkdir -p ~/.go-invoice
```

### Directory Structure
```
invoice-project/
├── data/           # Invoice and client data
├── templates/      # Custom invoice templates  
├── exports/        # Generated documents
├── imports/        # CSV files and imports
├── backups/        # Data backups
└── .env.config     # Project-specific config
```

## Step 2: Business Configuration

Create business configuration at `~/.go-invoice/.env.config`:

```bash
# Core Business Information
BUSINESS_NAME="Your Business Name"
BUSINESS_ADDRESS="123 Business St\nCity, State 12345\nCountry"
BUSINESS_EMAIL="billing@yourbusiness.com"
BUSINESS_PHONE="+1 (555) 123-4567"
BUSINESS_WEBSITE="https://yourbusiness.com"

# Tax and Legal
BUSINESS_TAX_ID="12-3456789"
BUSINESS_VAT_ID="GB123456789"  # If applicable

# Payment Configuration
PAYMENT_TERMS="Net 30"
PAYMENT_INSTRUCTIONS="Payment due within 30 days. Late fees may apply."

# Banking Details
BANK_NAME="Your Bank Name"
BANK_ACCOUNT="****1234"
BANK_ROUTING="123456789"
BANK_IBAN="GB29 NWBK 6016 1331 9268 19"  # If applicable
BANK_SWIFT="ABCDGB2L"  # If applicable

# Invoice Settings
INVOICE_PREFIX="INV"
INVOICE_START_NUMBER=1000
INVOICE_DUE_DAYS=30
CURRENCY="USD"
VAT_RATE=0.0

# Storage Configuration
DATA_DIR="~/.go-invoice/data"
BACKUP_DIR="~/.go-invoice/backups"
AUTO_BACKUP=true
BACKUP_INTERVAL="12h"
RETENTION_DAYS=365
```

### Developer-Specific Configuration
Add development-focused settings:

```bash
# Development Settings  
MCP_LOG_LEVEL="debug"
MCP_LOG_FILE="~/.go-invoice/mcp-claude-code.log"
MCP_TRANSPORT="stdio"

# Performance Settings
MCP_READ_TIMEOUT="60s"
MCP_WRITE_TIMEOUT="60s"
MCP_MAX_MESSAGE_SIZE="10485760"

# Integration Settings
ENABLE_PROJECT_SCOPE=true
ENABLE_RESOURCE_PATTERNS=true
ENABLE_SLASH_COMMANDS=true
```

## Step 3: MCP Server Configuration

Create MCP configuration at `~/.go-invoice/mcp-config.json`:

```json
{
  "server": {
    "name": "go-invoice-mcp",
    "version": "1.0.0",
    "description": "Natural language invoice management for developers"
  },
  "transport": {
    "type": "stdio",
    "readTimeout": "60s",
    "writeTimeout": "60s", 
    "maxMessageSize": 10485760,
    "enableLogging": true,
    "logLevel": "debug"
  },
  "security": {
    "enableSandbox": true,
    "allowedPaths": [
      "~/.go-invoice",
      "~/invoice-project",
      "~/Documents/Invoices",
      "~/Downloads",
      "~/Desktop",
      "/tmp"
    ],
    "blockedPaths": [
      "/etc/passwd",
      "/etc/shadow",
      "/var/log",
      "/usr/bin",
      "/sbin"
    ],
    "environmentWhitelist": [
      "HOME",
      "USER",
      "PATH",
      "GO_INVOICE_HOME", 
      "MCP_LOG_LEVEL",
      "MCP_TRANSPORT",
      "MCP_LOG_FILE",
      "EDITOR",
      "TERM"
    ]
  },
  "logging": {
    "enabled": true,
    "level": "debug",
    "file": "~/.go-invoice/mcp-claude-code.log",
    "maxSize": "50MB",
    "maxBackups": 10,
    "enableConsole": true
  },
  "features": {
    "enableResourcePatterns": true,
    "enableSlashCommands": true,
    "enableProjectScope": true,
    "enableFileWatching": false
  }
}
```

## Step 4: Claude Code Configuration

Configure Claude Code to use the go-invoice MCP server:

### Create MCP Configuration File
```bash
# Create Claude Code config directory
mkdir -p ~/.config/claude/mcp

# Create configuration file
cat > ~/.config/claude/mcp/go-invoice.json << 'EOF'
{
  "mcpServers": {
    "go-invoice": {
      "name": "Go Invoice MCP Server",
      "description": "Natural language invoice management through go-invoice CLI",
      "command": "go-invoice-mcp",
      "args": [
        "--stdio",
        "--config",
        "~/.go-invoice/mcp-config.json"
      ],
      "env": {
        "GO_INVOICE_HOME": "~/.go-invoice",
        "MCP_LOG_LEVEL": "debug",
        "MCP_TRANSPORT": "stdio",
        "MCP_LOG_FILE": "~/.go-invoice/mcp-claude-code.log"
      },
      "capabilities": {
        "tools": true,
        "resources": true,
        "prompts": false
      },
      "resources": {
        "patterns": [
          "@invoice:*",
          "@client:*", 
          "@timesheet:*",
          "@config:*"
        ],
        "handlers": {
          "@invoice": {
            "description": "Reference invoices by ID",
            "examples": ["@invoice:INV-2025-001", "@invoice:latest"]
          },
          "@client": {
            "description": "Reference clients by name or ID", 
            "examples": ["@client:\"Acme Corp\"", "@client:CL-001"]
          },
          "@timesheet": {
            "description": "Reference timesheet files",
            "examples": ["@timesheet:./hours.csv", "@timesheet:~/timesheets/august.csv"]
          },
          "@config": {
            "description": "Reference configuration sections",
            "examples": ["@config:invoice_defaults", "@config:storage_path"]
          }
        }
      },
      "slashCommands": {
        "prefix": "/mcp__go_invoice__",
        "commands": [
          {
            "name": "create_invoice",
            "description": "Create a new invoice",
            "aliases": ["new_invoice", "invoice_new"]
          },
          {
            "name": "list_invoices", 
            "description": "List all invoices with filters",
            "aliases": ["show_invoices", "invoices"]
          },
          {
            "name": "import_csv",
            "description": "Import timesheet from CSV",
            "aliases": ["import_timesheet", "add_hours"]
          },
          {
            "name": "generate_html",
            "description": "Generate HTML invoice", 
            "aliases": ["export_invoice", "make_invoice"]
          },
          {
            "name": "show_config",
            "description": "Display configuration",
            "aliases": ["config", "settings"]
          }
        ]
      },
      "metadata": {
        "version": "1.0.0",
        "author": "go-invoice",
        "projectScope": true,
        "documentation": "https://github.com/mrz1836/go-invoice/docs/mcp/claude-code-setup.md"
      }
    }
  }
}
EOF
```

## Step 5: Test the Configuration

### 1. Test MCP Server Standalone
```bash
# Test stdio transport
echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {"tools": {}}}}' | go-invoice-mcp --stdio --config ~/.go-invoice/mcp-config.json

# Test configuration validation  
go-invoice-mcp --config ~/.go-invoice/mcp-config.json --validate

# Test business configuration
go-invoice config validate
```

### 2. Test Resource Patterns
```bash
# Initialize sample data
go-invoice config init

# Create test client
go-invoice client create --name "Test Client" --email "test@example.com"

# Create test invoice
go-invoice invoice create --client "Test Client" --due-date "30 days"
```

### 3. Test in Claude Code

Start Claude Code and verify MCP integration:

```bash
# Start Claude Code with MCP
claude --mcp-config ~/.config/claude/mcp/go-invoice.json

# Or if using project-specific config
cd ~/invoice-project
claude --mcp-auto-detect
```

Test basic commands:
```
@config:show
```
```
/mcp__go_invoice__list_invoices
```
```
"Show me all my clients"
```

## Step 6: Resource Reference Patterns

Claude Code supports powerful resource reference patterns for seamless integration:

### Invoice References
```bash
# Reference specific invoice
@invoice:INV-2025-001

# Reference latest invoice
@invoice:latest

# Reference invoices by client
@invoice:client:"Acme Corp"

# Reference invoices by date range
@invoice:range:2024-01-01:2024-12-31
```

### Client References
```bash
# Reference by client name
@client:"Acme Corporation"

# Reference by client ID
@client:CL-001

# Reference all clients with filter
@client:active
@client:overdue
```

### Timesheet References
```bash
# Reference local CSV file
@timesheet:./january-hours.csv

# Reference file in home directory
@timesheet:~/timesheets/2024/january.csv

# Reference with specific format
@timesheet:./hours.csv:format:excel
```

### Configuration References
```bash
# Reference entire configuration
@config:all

# Reference specific section
@config:business
@config:invoice_defaults  
@config:storage

# Reference specific setting
@config:business.name
@config:invoice.currency
```

## Step 7: Slash Commands

Quick shortcuts for common operations:

### Available Commands
```bash
# Invoice Management
/mcp__go_invoice__create_invoice
/mcp__go_invoice__list_invoices
/mcp__go_invoice__show_invoice
/mcp__go_invoice__update_invoice

# Client Management  
/mcp__go_invoice__create_client
/mcp__go_invoice__list_clients
/mcp__go_invoice__show_client

# Data Operations
/mcp__go_invoice__import_csv
/mcp__go_invoice__generate_html
/mcp__go_invoice__export_data

# Configuration
/mcp__go_invoice__show_config
/mcp__go_invoice__validate_config
```

### Usage Examples
```bash
# Quick invoice creation
/mcp__go_invoice__create_invoice --client "Acme Corp" --hours 40 --rate 85

# List recent invoices
/mcp__go_invoice__list_invoices --status unpaid --limit 10

# Generate invoice document
/mcp__go_invoice__generate_html --invoice INV-2025-001 --output ./exports/
```

## Step 8: Development Workflows

### Daily Invoice Management
```bash
# Morning: Check status
"Show me all unpaid invoices due this week"

# Import timesheet data
"Import @timesheet:./daily-hours.csv and update existing invoices"

# Generate documents for new invoices
"Generate HTML invoices for all invoices created today"
```

### Monthly Reporting
```bash
# Revenue analysis
"Generate a revenue summary for @client:all for the last month"

# Export for accounting
"Export all invoices from last month to @timesheet:./exports/monthly-report.csv"

# Client review
"Show me which clients haven't been invoiced in the last 60 days"
```

### Automation Scripts

Create shell scripts for common tasks:

```bash
#!/bin/bash
# daily-invoice-workflow.sh

# Import timesheet data
claude mcp --server go-invoice "Import @timesheet:./today.csv"

# Generate invoices for clients with > 8 hours
claude mcp --server go-invoice "Create invoices for any client with more than 8 hours of unbilled time"

# Generate HTML documents
claude mcp --server go-invoice "Generate HTML invoices for all invoices created today and save to ./exports/"

# Send summary email (with additional tooling)
claude mcp --server go-invoice "Generate a summary of today's invoice activity" | mail -s "Daily Invoice Summary" admin@company.com
```

## Advanced Configuration

### Custom Resource Handlers

Add custom resource patterns in your MCP config:

```json
{
  "resources": {
    "patterns": [
      "@project:*",
      "@template:*",
      "@backup:*"
    ],
    "handlers": {
      "@project": {
        "description": "Reference project-specific data",
        "examples": ["@project:web-redesign", "@project:mobile-app"]
      },
      "@template": {
        "description": "Reference invoice templates",
        "examples": ["@template:professional", "@template:minimal"]
      }
    }
  }
}
```

### Environment-Specific Configuration

Development environment:
```bash
# .env.development
MCP_LOG_LEVEL="debug"
INVOICE_PREFIX="DEV"
AUTO_BACKUP=false
```

Production environment:
```bash
# .env.production
MCP_LOG_LEVEL="warn"
INVOICE_PREFIX="INV"
AUTO_BACKUP=true
BACKUP_INTERVAL="1h"
```

### Integration with External Tools

Git integration:
```bash
# .git/hooks/pre-commit
#!/bin/bash
# Validate invoices before commit
claude mcp --server go-invoice "Validate all invoice data for consistency"
```

CI/CD integration:
```yaml
# .github/workflows/invoice-validation.yml
name: Validate Invoices
on: [push, pull_request]
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Validate Invoice Data
        run: |
          claude mcp --server go-invoice "Validate all invoice and client data"
```

## Performance Optimization

### Large Dataset Handling
```json
{
  "transport": {
    "readTimeout": "120s",
    "writeTimeout": "120s",
    "maxMessageSize": 52428800
  },
  "logging": {
    "level": "warn",
    "maxSize": "100MB",
    "maxBackups": 20
  }
}
```

### Memory Optimization
```bash
# In your environment
export MCP_MEMORY_LIMIT="512MB"
export MCP_CACHE_SIZE="100MB"
export MCP_MAX_CONCURRENT_OPERATIONS="10"
```

## Security Best Practices

### 1. File System Security
```bash
# Secure configuration files
chmod 600 ~/.go-invoice/.env.config
chmod 600 ~/.config/claude/mcp/go-invoice.json

# Secure data directory
chmod 700 ~/.go-invoice/data
```

### 2. Process Security
```bash
# Run with limited permissions
sudo -u invoice-user claude mcp --server go-invoice

# Use systemd for process management
sudo systemctl --user enable claude-mcp@go-invoice
```

### 3. Network Security
```bash
# Disable network access if not needed
export MCP_NO_NETWORK=true

# Use local file operations only  
export MCP_LOCAL_ONLY=true
```

### 4. Data Protection
```bash
# Enable encryption for sensitive data
export MCP_ENCRYPT_DATA=true
export MCP_ENCRYPTION_KEY_FILE="~/.go-invoice/encryption.key"

# Automatic secure backups
export BACKUP_ENCRYPT=true
export BACKUP_COMPRESS=true
```

## Troubleshooting

### Common Issues

1. **stdio Communication Problems**
   ```bash
   # Test stdio manually
   echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}' | go-invoice-mcp --stdio
   
   # Check process communication
   ps aux | grep go-invoice-mcp
   lsof -p $(pgrep go-invoice-mcp)
   ```

2. **Resource Pattern Issues**
   ```bash
   # Validate resource patterns
   claude mcp --server go-invoice --validate-resources
   
   # Test specific pattern
   claude mcp --server go-invoice --test-resource "@invoice:INV-2025-001"
   ```

3. **Performance Issues**
   ```bash
   # Check logs for bottlenecks
   tail -f ~/.go-invoice/mcp-claude-code.log
   
   # Monitor resource usage
   htop -p $(pgrep go-invoice-mcp)
   ```

4. **Configuration Problems**
   ```bash
   # Validate all configurations
   go-invoice config validate
   claude mcp --validate-config ~/.config/claude/mcp/go-invoice.json
   
   # Reset to defaults
   go-invoice config init --force
   ```

### Debug Mode

Enable comprehensive debugging:

```bash
# Set debug environment
export MCP_LOG_LEVEL="debug"
export MCP_TRACE_ENABLED="true"
export MCP_DEBUG_STDIO="true"

# Run with debugging
claude --debug --mcp-debug mcp --server go-invoice "list all invoices"
```

## Next Steps

1. **Master Resource Patterns**: Learn advanced resource referencing
2. **Create Automation Scripts**: Build workflows for your business
3. **Integrate with Other Tools**: Connect with Git, CI/CD, accounting software
4. **Customize Templates**: Design professional invoice layouts
5. **Monitor and Optimize**: Track performance and improve workflows

You're now ready to use Claude Code for powerful, automated invoice management with natural language and advanced developer features!