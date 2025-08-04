# 📚 go-invoice MCP Tools Examples

This directory contains comprehensive examples for using go-invoice through the MCP (Model Context Protocol) tools system with Claude Desktop. These examples demonstrate how to leverage natural language interactions for efficient invoice management workflows.

## 📁 Directory Structure

```
examples/
├── README.md                           # This file - MCP tools guidance
├── mcp-tools/                          # MCP-specific examples and setups
│   ├── claude-desktop-config.json      # Claude Desktop MCP configuration
│   ├── tool-usage-examples.md          # Natural language tool examples
│   └── workflow-automation.md          # Automation pattern examples
├── timesheet-standard.csv              # Basic CSV format example
├── timesheet-excel.csv                 # Excel export format example  
├── timesheet-tabs.tsv                  # Tab-separated format example
├── templates/                          # Custom invoice templates
│   ├── modern-invoice.html             # Modern gradient design template
│   └── minimal-invoice.html            # Clean minimal template
├── scripts/                            # Automation scripts
│   ├── monthly-billing.sh              # Monthly billing automation
│   └── setup-client.sh                 # New client setup helper
├── advanced/                           # Advanced usage examples
│   ├── multi-rate-timesheet.csv        # Different rates per task
│   ├── european-format.csv             # European date format (DD/MM/YYYY)
│   └── project-phases.csv              # Project-based billing
├── config/                             # Configuration examples
│   └── sample.env                      # Environment variables template
└── workflows/                          # Complete workflow examples
    └── complete-example.md             # Step-by-step tutorial
```

## 🤖 MCP Tools Integration

### What is MCP?

The Model Context Protocol (MCP) enables Claude Desktop to interact directly with go-invoice through a standardized tool interface. This means you can manage invoices, clients, and billing workflows using natural language conversations with Claude.

### Available Tool Categories

1. **Invoice Management** (7 tools)
   - Create, update, list, and manage invoices
   - Add/remove work items from invoices
   - Natural language invoice operations

2. **Client Management** (5 tools)
   - Create and manage client information
   - Update contact details and billing preferences
   - Search and organize client relationships

3. **Data Import** (3 tools)
   - Import timesheet CSV files
   - Validate data before import
   - Preview import results

4. **Document Generation** (3 tools)
   - Generate HTML invoices with templates
   - Create summaries and reports
   - Export data in various formats

5. **Configuration** (3 tools)
   - Validate system configuration
   - Display current settings
   - Initialize new configurations

## 🚀 Quick Start with MCP Tools

### 1. Setting Up Claude Desktop

First, configure Claude Desktop to use the go-invoice MCP server:

```json
{
  "mcpServers": {
    "go-invoice": {
      "command": "/path/to/go-invoice-mcp",
      "args": ["--config", "/path/to/config.json"],
      "env": {
        "GO_INVOICE_DATA_DIR": "/path/to/invoice-data"
      }
    }
  }
}
```

### 2. Natural Language Invoice Creation

**With Claude Desktop (conversational):**

**You**: "I need to create an invoice for Acme Corp for 40 hours of web development at $85/hour"

**Claude**: I'll create that invoice for you. Let me start by creating Acme Corp as a client if they don't exist yet, then create the invoice.

```json
{
  "tool": "client_create",
  "parameters": {
    "name": "Acme Corp",
    "email": "billing@acme.com"
  }
}
```

```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "Acme Corp",
    "work_items": [
      {
        "description": "Web development",
        "hours": 40,
        "rate": 85
      }
    ],
    "due_days": 30
  }
}
```

Your invoice has been created! The total is $3,400. Would you like me to generate an HTML version?

### 3. CSV Import with Natural Language

**You**: "I have a timesheet CSV file I'd like to import. Can you help me process it and create invoices?"

**Claude**: I'll help you import your timesheet data. Let me first preview the file to understand its structure, then import it.

```json
{
  "tool": "import_preview",
  "parameters": {
    "file_path": "/path/to/timesheet.csv"
  }
}
```

```json
{
  "tool": "import_csv",
  "parameters": {
    "file_path": "/path/to/timesheet.csv",
    "import_mode": "group_by_client",
    "create_invoices": true
  }
}
```

Perfect! I've imported your timesheet and created invoices for 3 clients. Would you like me to generate HTML documents for all of them?

### 4. Document Generation

**You**: "Generate professional HTML invoices for all my pending invoices"

**Claude**: I'll generate HTML invoices for all your pending invoices using the professional template.

```json
{
  "tool": "invoice_list",
  "parameters": {
    "status_filter": "pending",
    "include_details": true
  }
}
```

```json
{
  "tool": "generate_html",
  "parameters": {
    "batch_invoices": ["INV-001", "INV-002", "INV-003"],
    "template": "professional",
    "output_dir": "./client-invoices/"
  }
}
```

Done! I've generated professional HTML invoices for 3 pending invoices and saved them to your client-invoices directory.

## 📄 CSV Format Examples

### Standard Format (`timesheet-standard.csv`)
```csv
date,description,hours,rate
2024-01-15,Backend API development,8.0,100.00
2024-01-16,Database optimization and queries,6.5,100.00
```

### Excel Export Format (`timesheet-excel.csv`)
```csv
Date,Hours Worked,Hourly Rate,Work Description
01/15/2024,8,100,"Backend API development"
01/16/2024,6.5,100,"Database optimization, query performance"
```

### Multi-Rate Format (`advanced/multi-rate-timesheet.csv`)
```csv
date,description,hours,rate
2024-01-15,Senior consultation,4.0,150.00
2024-01-15,Regular development,4.0,125.00
2024-01-16,Junior mentoring,3.0,75.00
```

## 🎨 Template Examples

### Modern Template Features
- Gradient background design
- Professional typography
- Print-friendly layout
- Status badges
- Responsive design

### Minimal Template Features
- Clean, simple design
- Black and white aesthetic
- Focused on content
- Lightweight HTML
- Classic professional look

## 🤖 Automation Scripts

### Monthly Billing Script (`scripts/monthly-billing.sh`)

Automates the complete monthly billing process:

```bash
# Process current month for all clients
./examples/scripts/monthly-billing.sh

# Process specific month/year
./examples/scripts/monthly-billing.sh 12 2023
```

Features:
- ✅ Validates all timesheets
- ✅ Imports data for each client
- ✅ Creates and sends invoices
- ✅ Generates summary reports
- ✅ Creates backups
- ✅ Detailed logging

### Client Setup Script (`scripts/setup-client.sh`)

Sets up a new client with complete workflow:

```bash
# Basic setup
./examples/scripts/setup-client.sh "Client Name"

# Full setup with email and rate
./examples/scripts/setup-client.sh "Client Name" "email@client.com" "125.00"
```

Creates:
- ✅ Client record in go-invoice
- ✅ Directory structure for timesheets/invoices
- ✅ Sample timesheet template
- ✅ Quick command scripts
- ✅ Client list for automation

## 🏗️ Advanced Examples

### Multi-Rate Billing
Handle different rates for different types of work:
- Senior consultation: $150/hour
- Regular development: $125/hour
- Code review: $100/hour
- Documentation: $75/hour

### Project Phase Billing
Organize work by project phases:
- Phase 1: Analysis & Design
- Phase 2: Core Development  
- Phase 3: Frontend & UI
- Phase 4: Testing & QA
- Phase 5: Deployment & Launch

### European Date Formats
Support for DD/MM/YYYY date formats:
```bash
go-invoice import csv examples/advanced/european-format.csv \
  --client "EU Client" \
  --date-format "02/01/2006"
```

## ⚙️ Configuration Examples

### Environment Variables (`config/sample.env`)

Complete configuration template with:
- Business information
- Banking details
- Invoice settings
- Payment terms
- Integration settings
- Security options

Copy and customize:
```bash
cp examples/config/sample.env .env
# Edit .env with your business details
```

## 📋 Complete Workflow Example

See `workflows/complete-example.md` for a comprehensive step-by-step tutorial covering:

1. **Client Setup** - Adding clients to the system
2. **Time Tracking** - Preparing CSV data
3. **Validation** - Checking data before import
4. **Import Process** - Loading timesheet data
5. **Invoice Creation** - Generating professional invoices
6. **Status Management** - Tracking sent/paid status
7. **Automation** - Scaling to multiple clients
8. **Integration** - Working with other tools

## 🛠️ Usage Patterns

### For Freelancers
- Use `timesheet-standard.csv` format
- Customize `modern-invoice.html` template
- Set up monthly automation with `monthly-billing.sh`

### For Consultants  
- Use `multi-rate-timesheet.csv` for different service rates
- Organize by `project-phases.csv` for complex projects
- Use `minimal-invoice.html` for corporate clients

### For Agencies
- Use automation scripts for multiple clients
- Set up team-specific rate structures
- Implement project-based billing workflows

## 🔧 Customization Tips

### Templates
1. Copy default template: `cp templates/invoice/default.html my-template.html`
2. Modify CSS styles and layout
3. Test with sample data: `go-invoice invoice generate --template my-template.html`
4. Add custom branding and colors

### Scripts
1. Copy automation scripts to your project
2. Modify paths and configuration variables
3. Add custom logic for your workflow
4. Set up cron jobs for automatic execution

### CSV Formats
1. Export from your time tracking tool
2. Map columns to go-invoice format
3. Test with `--dry-run` flag first
4. Create templates for consistency

## 📖 Additional Resources

- **Tutorial**: `docs/tutorial.md` - Comprehensive user guide
- **Documentation**: `README.md` - Full feature documentation
- **CLI Help**: `go-invoice help` - Built-in command help
- **Templates**: `templates/` - Default invoice templates

## 🤝 Contributing Examples

Have a useful example or workflow? Contributions are welcome!

1. Fork the repository
2. Add your example to the appropriate directory
3. Include documentation and comments
4. Test with sample data
5. Submit a pull request

## 💡 Tips for Success

1. **Start Simple**: Begin with basic CSV format and default template
2. **Validate Early**: Always use `--dry-run` and `--validate` flags
3. **Automate Gradually**: Start manual, then add automation as you scale
4. **Backup Regularly**: Use the backup features in automation scripts
5. **Test Templates**: Verify custom templates with sample data first
6. **Stay Organized**: Use consistent naming and directory structures

---

**Happy Invoicing!** 🎉

For more help, see the main documentation or run `go-invoice help` for built-in assistance.