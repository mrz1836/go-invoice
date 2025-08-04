# go-invoice MCP Integration

Welcome to the go-invoice Model Context Protocol (MCP) integration guide. This powerful integration allows you to manage invoices, clients, and business workflows using natural language through Claude Desktop and Claude Code.

## Quick Start

The go-invoice MCP server provides 21 tools across 5 categories, enabling you to:
- Create and manage invoices through conversation
- Import timesheet data from CSV files
- Generate professional HTML invoices
- Manage client information and relationships
- Configure and validate your invoice system

## What is MCP?

Model Context Protocol (MCP) is a standard that allows Claude to interact with external tools and data sources. The go-invoice MCP integration enables Claude to:

- **Understand your business context** - Access your invoice and client data
- **Execute invoice operations** - Create invoices, import data, generate documents
- **Provide intelligent assistance** - Natural language queries about your business
- **Automate workflows** - Complex multi-step invoice processing

## Platform Support

### Claude Desktop (HTTP Transport)
Perfect for business users who want a visual interface with natural language invoice management.

**Key Features:**
- Point-and-click interface with conversational AI
- Real-time invoice preview and editing
- Visual client management
- Integrated document generation
- Automatic backup and sync

### Claude Code (stdio Transport)  
Ideal for developers and power users who prefer command-line integration.

**Key Features:**
- Terminal-based invoice management
- Direct file system integration
- Advanced scripting capabilities
- Developer-friendly resource references
- Slash command shortcuts

## Available Tools

### 📋 Invoice Management (7 tools)
Create, update, and manage invoices throughout their lifecycle.

| Tool | Description |
|------|-------------|
| `invoice_create` | Create new invoices with automatic calculations |
| `invoice_list` | List and filter existing invoices |
| `invoice_show` | Display detailed invoice information |
| `invoice_update` | Modify invoice metadata and status |
| `invoice_delete` | Remove invoices from the system |
| `invoice_add_item` | Add work items to existing invoices |
| `invoice_remove_item` | Remove work items from invoices |

### 📥 Data Import (3 tools)
Import timesheet data and external information.

| Tool | Description |
|------|-------------|
| `import_csv` | Import timesheet data from CSV files |
| `import_validate` | Validate import data before processing |
| `import_preview` | Preview data changes before import |

### 📄 Data Export (3 tools)
Generate and export professional documents.

| Tool | Description |
|------|-------------|
| `generate_html` | Generate professional HTML invoices |
| `generate_summary` | Create business summaries and reports |
| `export_data` | Export data in various formats |

### 👥 Client Management (5 tools)
Manage client information and relationships.

| Tool | Description |
|------|-------------|
| `client_create` | Add new clients to the system |
| `client_list` | List and search existing clients |
| `client_show` | Display detailed client information |
| `client_update` | Modify client contact details |
| `client_delete` | Remove clients from the system |

### ⚙️ Configuration (3 tools)
System setup and validation tools.

| Tool | Description |
|------|-------------|
| `config_show` | Display current configuration |
| `config_validate` | Validate system setup |
| `config_init` | Initialize new configuration |

## Setup Guides

Choose your platform to get started:

### 🖥️ Claude Desktop Setup
**Recommended for:** Business users, visual interface preference, document generation

[**→ Claude Desktop Setup Guide**](claude-desktop-setup.md)

- HTTP transport configuration
- Visual interface setup
- Document generation workflows
- Business user examples

### 💻 Claude Code Setup  
**Recommended for:** Developers, command-line preference, automation scripts

[**→ Claude Code Setup Guide**](claude-code-setup.md)

- stdio transport configuration
- Terminal integration
- Resource reference patterns
- Developer workflows

## Common Use Cases

### 💼 Freelancers & Consultants
```
"Create an invoice for Acme Corp for 40 hours of web development at $85/hour, due in 30 days"
```

### 🏢 Small Businesses
```
"Import this month's timesheet data and generate invoices for all clients with outstanding hours"
```

### 🧾 Accounting Integration
```
"Export all invoices from Q1 2024 in CSV format for QuickBooks import"
```

### 📊 Business Analytics
```
"Show me a summary of revenue by client for the last 6 months"
```

## Security & Best Practices

- **File System Access**: Sandboxed to your invoice directory
- **Data Validation**: All inputs validated before processing
- **Backup Strategy**: Automatic backups of invoice data
- **Access Control**: Environment-based security configuration

## Documentation Structure

| Document | Purpose |
|----------|---------|
| [Configuration Guide](configuration.md) | Detailed configuration options and security settings |
| [Troubleshooting](troubleshooting.md) | Common issues and solutions |
| [Claude Desktop Setup](claude-desktop-setup.md) | HTTP transport setup guide |
| [Claude Code Setup](claude-code-setup.md) | stdio transport setup guide |

## Support

- **Documentation**: Complete guides and examples
- **Error Messages**: Detailed validation and troubleshooting info
- **Logging**: Comprehensive debug information available
- **Community**: GitHub repository for issues and feature requests

## Getting Started

1. **Choose your platform**: Claude Desktop or Claude Code
2. **Follow the setup guide**: Platform-specific configuration
3. **Test the integration**: Verify tools are working
4. **Start invoicing**: Begin with simple commands
5. **Explore advanced features**: Automation and workflows

Ready to transform your invoice management? Choose your setup guide above and start using natural language to manage your business in minutes!