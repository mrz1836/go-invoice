# Claude Code Integration Guide

This guide explains how to integrate the go-invoice MCP server with Claude Code using stdio transport for project-level invoice management.

## Overview

The go-invoice MCP server provides seamless invoice management within Claude Code through slash commands and resource mentions. Work with invoices, clients, and timesheets directly in your development environment using natural language.

## Prerequisites

- Claude Code installed
- go-invoice CLI installed and configured
- Go 1.21+ (for building the MCP server)

## Quick Setup

### Global Setup

Run the automated setup script:

```bash
./scripts/setup-claude-integration.sh --code
```

### Project-Specific Setup

Navigate to your project directory and run:

```bash
/path/to/go-invoice/scripts/setup-claude-code-integration.sh
```

This will:
- Create project invoice structure
- Configure Claude Code for this project
- Generate sample templates and timesheets
- Test the stdio transport

## Manual Setup

### 1. Build the MCP Server

```bash
cd /path/to/go-invoice
go build -o bin/go-invoice-mcp ./cmd/mcp-server
```

### 2. Global Configuration

Create Claude Code MCP configuration:

**Location**: `~/.config/claude-code/mcp_config.json`

```json
{
  "mcpServers": {
    "go-invoice": {
      "name": "Go Invoice MCP Server",
      "description": "Natural language invoice management",
      "command": "go-invoice-mcp",
      "args": ["--stdio", "--config", "~/.go-invoice/mcp-config.json"],
      "env": {
        "GO_INVOICE_HOME": "~/.go-invoice",
        "MCP_TRANSPORT": "stdio"
      },
      "capabilities": {
        "tools": true,
        "resources": true
      }
    }
  }
}
```

### 3. Project Configuration

Create `.claude_config.json` in your project root:

```json
{
  "mcp": {
    "servers": {
      "go-invoice": {
        "command": "./bin/go-invoice-mcp",
        "args": ["--stdio", "--config", "./.go-invoice/config.json"],
        "workingDirectory": ".",
        "env": {
          "GO_INVOICE_PROJECT": "./",
          "MCP_TRANSPORT": "stdio"
        },
        "projectSettings": {
          "projectName": "MyProject",
          "defaultClient": "Acme Corp",
          "invoicePrefix": "INV",
          "autoImportPath": "./timesheets",
          "outputPath": "./invoices"
        }
      }
    }
  }
}
```

## Project Structure

The integration creates a standardized project structure:

```
your-project/
├── .claude_config.json          # Claude Code configuration
├── .go-invoice/
│   ├── config.json             # Project invoice settings
│   ├── data/                   # Invoice data storage
│   └── logs/                   # MCP server logs
├── invoices/
│   ├── drafts/                 # Draft invoices
│   ├── sent/                   # Sent invoices
│   └── paid/                   # Paid invoices
├── timesheets/
│   ├── pending/                # Unprocessed timesheets
│   └── processed/              # Imported timesheets
└── templates/
    └── invoice.html            # Invoice template
```

## Slash Commands

Use these slash commands to interact with go-invoice:

### Invoice Management

- `/invoice` - Create a new invoice
- `/mcp__go_invoice__list_invoices` - List all invoices
- `/mcp__go_invoice__show_invoice` - Display invoice details
- `/mcp__go_invoice__update_invoice` - Update invoice information

### Data Import

- `/import` - Import timesheet from CSV
- `/mcp__go_invoice__import_excel` - Import from Excel file

### Document Generation

- `/generate` - Generate HTML invoice
- `/mcp__go_invoice__generate_pdf` - Generate PDF invoice

### Configuration

- `/mcp__go_invoice__show_config` - Display current configuration
- `/mcp__go_invoice__summary` - Show project invoice summary

## Resource Mentions

Reference invoice data using @ mentions:

### Invoice References

- `@invoice:INV-2025-001` - Reference specific invoice
- `@invoice:latest` - Reference most recent invoice
- `@invoice:draft` - Reference draft invoices

### Client References

- `@client:"Acme Corp"` - Reference client by name
- `@client:CL-001` - Reference client by ID

### Timesheet References

- `@timesheet:./hours.csv` - Reference local timesheet
- `@timesheet:~/timesheets/january.csv` - Reference timesheet by path

### Configuration References

- `@config:invoice_defaults` - Reference invoice defaults
- `@config:storage_path` - Reference storage configuration

## Usage Examples

### Creating an Invoice

```
/invoice

Create an invoice for @client:"TechCorp Solutions" for January 2025.
Import hours from @timesheet:./timesheets/january.csv
Use rate $175/hour and include 8.5% tax.
```

### Importing Timesheet Data

```
/import @timesheet:./timesheets/weekly-hours.csv

Parse this CSV and create separate invoices for each client.
Set the invoice date to the last day of the timesheet period.
```

### Generating Documents

```
/generate @invoice:INV-2025-003

Generate HTML for this invoice and save to ./invoices/sent/
Use the custom template at ./templates/branded-invoice.html
```

### Project Summary

```
/mcp__go_invoice__summary

Show me:
- Total invoiced amount this month
- Outstanding invoices
- Top 3 clients by revenue
- Recent timesheet imports
```

## Workflow Integration

### Development Project Billing

```
# Import hours at end of sprint
/import @timesheet:./sprint-hours.csv

# Create invoice for main client
/invoice @client:"Primary Client"
Add all development hours from this sprint
Include code review and documentation time

# Generate and review
/generate @invoice:latest
```

### Freelance Project Management

```
# Setup new project
Set up invoicing for new project "E-commerce Platform"
Default client: @client:"RetailCorp"
Hourly rate: $150
Invoice monthly on the 1st

# Track work
Import @timesheet:./weekly-hours.csv every Friday
Auto-categorize as "Development", "Meetings", or "Admin"

# Monthly billing
/invoice @client:"RetailCorp" 
Include all unbilled hours from this month
Apply 10% discount for early payment
```

## Configuration

### Project Settings

Customize behavior in `.go-invoice/config.json`:

```json
{
  "storage_path": "./.go-invoice/data",
  "invoice_defaults": {
    "currency": "USD",
    "tax_rate": 0.085,
    "payment_terms": 30,
    "prefix": "INV"
  },
  "import": {
    "csv": {
      "delimiter": ",",
      "date_format": "2006-01-02",
      "time_format": "15:04"
    }
  },
  "export": {
    "html": {
      "output_dir": "./invoices",
      "open_after_generate": true
    }
  }
}
```

### Workspace Integration

Configure file watching in `.claude_config.json`:

```json
{
  "workspace": {
    "invoices": {
      "path": "./invoices",
      "watch": true,
      "autoOpen": true
    },
    "timesheets": {
      "path": "./timesheets", 
      "watch": true,
      "patterns": ["*.csv", "*.xlsx"]
    }
  }
}
```

## Templates

### Invoice Template

Create custom templates in `./templates/`:

```html
<!DOCTYPE html>
<html>
<head>
    <title>Invoice {{.Invoice.Number}}</title>
    <style>
        /* Custom styling */
    </style>
</head>
<body>
    <h1>Invoice {{.Invoice.Number}}</h1>
    
    <div class="client-info">
        <h2>Bill To:</h2>
        <p>{{.Client.Name}}</p>
        <p>{{.Client.Address}}</p>
    </div>
    
    <table class="items">
        {{range .Invoice.Items}}
        <tr>
            <td>{{.Description}}</td>
            <td>{{.Quantity}} hrs</td>
            <td>${{.Rate}}/hr</td>
            <td>${{.Total}}</td>
        </tr>
        {{end}}
    </table>
    
    <div class="total">
        <strong>Total: ${{.Invoice.Total}}</strong>
    </div>
</body>
</html>
```

### CSV Timesheet Format

Standard format for importing timesheets:

```csv
Date,Hours,Description,Project,Rate
2025-01-15,2.5,"API development",Core,150
2025-01-16,4.0,"Database optimization",Core,150
2025-01-17,1.5,"Client meeting",Admin,150
```

## Shortcuts and Aliases

Define custom shortcuts in `.claude_config.json`:

```json
{
  "shortcuts": {
    "invoice": {
      "description": "Quick invoice creation",
      "server": "go-invoice",
      "tool": "invoice_create"
    },
    "bill": {
      "description": "Generate and send invoice",
      "server": "go-invoice", 
      "tool": "generate_html"
    },
    "import": {
      "description": "Import timesheet",
      "server": "go-invoice",
      "tool": "import_csv"
    }
  }
}
```

## Logging and Debugging

### Enable Debug Logging

Set log level in project configuration:

```json
{
  "env": {
    "MCP_LOG_LEVEL": "debug",
    "MCP_LOG_FILE": "./.go-invoice/debug.log"
  }
}
```

### View Logs

```bash
tail -f ./.go-invoice/logs/mcp.log
```

### Test stdio Connection

```bash
echo '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}' | \
  ./bin/go-invoice-mcp --stdio --config ./.go-invoice/config.json
```

## Troubleshooting

### Command Not Found

1. Verify MCP server is built:
   ```bash
   ls -la ./bin/go-invoice-mcp
   ```

2. Check Claude Code configuration:
   ```bash
   cat .claude_config.json
   ```

3. Restart Claude Code

### Resource Mentions Not Working

1. Check resource patterns in configuration
2. Verify file paths are correct
3. Review workspace settings

### Import Failures

1. Validate CSV format and headers
2. Check file permissions
3. Review import configuration

### Performance Issues

1. Check project size and file count
2. Reduce logging verbosity
3. Optimize workspace watching patterns

## Advanced Features

### Custom Tool Integration

Extend functionality with custom tools:

```json
{
  "extensions": {
    "customTools": {
      "enabled": true,
      "path": "./.go-invoice/custom-tools"
    }
  }
}
```

### Webhook Integration

Configure external notifications:

```json
{
  "hooks": {
    "invoice_created": {
      "enabled": true,
      "url": "https://api.slack.com/webhook",
      "template": "New invoice {{.Invoice.Number}} created for {{.Client.Name}}"
    }
  }
}
```

### Automated Workflows

Set up recurring tasks:

```json
{
  "automation": {
    "monthly_billing": {
      "enabled": true,
      "schedule": "1 0 1 * *",
      "action": "create_invoices_for_all_clients"
    }
  }
}
```

## Best Practices

### Project Organization

- Keep timesheets in a consistent format
- Use descriptive invoice prefixes (e.g., `PROJ-2025-001`)
- Organize templates by client or project type
- Maintain separate directories for drafts and sent invoices

### Data Management

- Backup `.go-invoice/data/` regularly
- Version control templates and configuration
- Use gitignore for logs and temporary files

### Collaboration

- Share project configuration with team members
- Document custom templates and workflows
- Use consistent client naming conventions

## Support

- **Documentation**: [go-invoice docs](https://github.com/mrz1836/go-invoice)
- **Issues**: [GitHub Issues](https://github.com/mrz1836/go-invoice/issues)
- **Claude Code MCP**: [MCP Documentation](https://docs.anthropic.com/en/docs/claude-code/mcp)