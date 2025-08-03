# go-invoice MCP Documentation Index

Welcome to the comprehensive documentation for go-invoice Model Context Protocol (MCP) integration. This documentation provides everything you need to set up and use natural language invoice management with Claude Desktop and Claude Code.

## 📖 Documentation Overview

| Document | Description | Audience | Estimated Reading Time |
|----------|-------------|----------|----------------------|
| [**README.md**](README.md) | Main overview and quick start guide | All users | 5 minutes |
| [**User Guide**](user-guide.md) | Complete user guide with examples and workflows | All users | 20 minutes |
| [**Claude Desktop Setup**](claude-desktop-setup.md) | HTTP transport setup for Claude Desktop | Business users | 15 minutes |
| [**Claude Code Setup**](claude-code-setup.md) | stdio transport setup for Claude Code | Developers | 20 minutes |
| [**Configuration Guide**](configuration.md) | Detailed configuration options and security | System administrators | 25 minutes |
| [**Troubleshooting**](troubleshooting.md) | Common issues and debugging | All users | 15 minutes |

## 🚀 Quick Start Path

### For Business Users (Claude Desktop)
1. Read [README.md](README.md) for overview
2. Follow [Claude Desktop Setup Guide](claude-desktop-setup.md)
3. Use [User Guide](user-guide.md) for daily operations
4. Reference [Troubleshooting](troubleshooting.md) if needed

### For Developers (Claude Code)
1. Read [README.md](README.md) for overview
2. Follow [Claude Code Setup Guide](claude-code-setup.md)
3. Use [User Guide](user-guide.md) for workflows
4. Customize with [Configuration Guide](configuration.md)

### For System Administrators
1. Read [README.md](README.md) for overview
2. Study [Configuration Guide](configuration.md) for security
3. Follow appropriate setup guide for your platform
4. Keep [Troubleshooting](troubleshooting.md) handy for support

## 🎯 Use Case Guides

### Freelancers & Consultants
**Recommended Path:**
- [Claude Desktop Setup](claude-desktop-setup.md) → [User Guide: Freelancer Workflow](user-guide.md#freelancer-workflow)

**Key Features:**
- Simple invoice creation through conversation
- CSV timesheet import
- Professional document generation
- Client relationship management

### Small Businesses
**Recommended Path:**
- [Claude Desktop Setup](claude-desktop-setup.md) → [User Guide: Business Scenarios](user-guide.md#business-scenarios)

**Key Features:**
- Multi-client invoice management
- Recurring billing setup
- Financial reporting
- Tax compliance features

### Development Teams
**Recommended Path:**
- [Claude Code Setup](claude-code-setup.md) → [User Guide: Advanced Workflows](user-guide.md#advanced-workflows)

**Key Features:**
- Terminal integration
- Resource reference patterns
- Automation scripting
- CI/CD integration

### Enterprise Deployments
**Recommended Path:**
- [Configuration Guide](configuration.md) → Platform Setup → [User Guide](user-guide.md)

**Key Features:**
- Advanced security configuration
- Accounting system integration
- Compliance reporting
- Multi-user management

## 🛠️ Technical Reference

### System Requirements
- **go-invoice**: Version 1.0.0 or later
- **Claude Desktop**: Latest version (for HTTP transport)
- **Claude Code**: Latest version (for stdio transport)
- **Operating System**: Windows 10+, macOS 10.15+, or Linux

### Architecture Overview
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Claude Client │    │  MCP Transport  │    │ go-invoice CLI  │
│ (Desktop/Code)  │◄──►│  (HTTP/stdio)   │◄──►│   + MCP Server  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Available Tools (21 Total)
- **Invoice Management**: 7 tools for complete invoice lifecycle
- **Data Import**: 3 tools for timesheet and external data
- **Data Export**: 3 tools for document generation
- **Client Management**: 5 tools for customer relationships
- **Configuration**: 3 tools for system management

## 📋 Tool Categories

### Invoice Management Tools
| Tool | Description | Common Usage |
|------|-------------|--------------|
| `invoice_create` | Create new invoices | "Create invoice for Acme Corp for 40 hours at $85/hour" |
| `invoice_list` | List and filter invoices | "Show me all unpaid invoices" |
| `invoice_show` | Display invoice details | "Show invoice INV-2025-001" |
| `invoice_update` | Modify existing invoices | "Mark invoice INV-2025-001 as paid" |
| `invoice_delete` | Remove invoices | "Delete test invoice" |
| `invoice_add_item` | Add work items | "Add 8 hours to Acme Corp invoice" |
| `invoice_remove_item` | Remove work items | "Remove last item from invoice" |

### Data Import Tools
| Tool | Description | Common Usage |
|------|-------------|--------------|
| `import_csv` | Import timesheet data | "Import ~/Downloads/hours.csv" |
| `import_validate` | Validate import data | "Validate import file before processing" |
| `import_preview` | Preview import changes | "Show what would be imported" |

### Data Export Tools
| Tool | Description | Common Usage |
|------|-------------|--------------|
| `generate_html` | Generate HTML invoices | "Generate HTML for INV-2025-001" |
| `generate_summary` | Create business reports | "Generate revenue summary for Q1" |
| `export_data` | Export various formats | "Export client data to CSV" |

### Client Management Tools
| Tool | Description | Common Usage |
|------|-------------|--------------|
| `client_create` | Add new clients | "Create client 'Acme Corp' with email acme@example.com" |
| `client_list` | List clients | "Show all clients" |
| `client_show` | Display client details | "Show Acme Corp details" |
| `client_update` | Update client info | "Update Acme Corp email address" |
| `client_delete` | Remove clients | "Delete test client" |

### Configuration Tools
| Tool | Description | Common Usage |
|------|-------------|--------------|
| `config_show` | Display configuration | "Show my business configuration" |
| `config_validate` | Validate setup | "Validate system configuration" |
| `config_init` | Initialize configuration | "Set up new business profile" |

## 🔒 Security Features

### Sandbox Security
- File system access restricted to allowed paths
- Command injection prevention
- Input validation and sanitization
- Resource limit enforcement

### Data Protection
- Automatic encrypted backups
- Secure configuration storage
- Environment variable isolation
- Audit logging

### Access Control
- User-based permissions
- Path-based restrictions
- Environment variable whitelist
- Rate limiting

## 📊 Common Workflows

### Daily Operations
```
"Show me today's outstanding invoices"
"Import timesheet and create invoices for clients with >8 hours"
"Generate HTML invoices for all new invoices"
```

### Weekly Management
```
"Create weekly summary report"
"Follow up on overdue invoices"
"Export completed invoices for accounting"
```

### Monthly Processes
```
"Generate monthly revenue report by client"
"Create recurring invoices for retainer clients"
"Backup all invoice data"
```

## 🆘 Support Resources

### Self-Help Resources
1. **[Troubleshooting Guide](troubleshooting.md)** - Common issues and solutions
2. **[Configuration Reference](configuration.md)** - Complete configuration options
3. **[User Guide Examples](user-guide.md)** - Practical usage examples

### Debug Information
```bash
# System health check
go-invoice --version
go-invoice config validate
go-invoice-mcp --version

# Log analysis
tail -f ~/.go-invoice/mcp-server.log
```

### Community Support
- **GitHub Issues**: Bug reports and feature requests
- **Documentation**: Comprehensive guides and references
- **Examples**: Real-world usage patterns

## 📈 Getting the Most Value

### Optimization Tips
1. **Start Simple**: Begin with basic invoice creation
2. **Import Existing Data**: Load your current clients and invoices
3. **Establish Patterns**: Create consistent workflows
4. **Automate Repetitive Tasks**: Use batch operations
5. **Customize for Your Business**: Adapt templates and processes

### Best Practices
- Use consistent naming conventions for clients
- Regularly backup your invoice data
- Keep configuration files secure
- Monitor system logs for issues
- Update to latest versions regularly

## 🎓 Learning Path

### Beginner (Week 1)
- [ ] Complete platform setup
- [ ] Create first client and invoice
- [ ] Generate first HTML document
- [ ] Import sample timesheet data

### Intermediate (Week 2-3)
- [ ] Set up automated workflows
- [ ] Customize invoice templates
- [ ] Integrate with existing tools
- [ ] Establish backup procedures

### Advanced (Month 2+)
- [ ] Develop custom automation scripts
- [ ] Optimize performance for your workload
- [ ] Implement compliance procedures
- [ ] Train team members on usage

Ready to get started? Choose your platform and follow the appropriate setup guide to begin your natural language invoice management journey!