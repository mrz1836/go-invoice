# go-invoice MCP Integration Documentation

This directory contains comprehensive documentation for the Model Context Protocol (MCP) integration of go-invoice, enabling natural language invoice management through Claude Desktop and Claude Code.

## Documentation Index

### Quick Start Guides

- **[MCP Integration Overview](mcp-integration.md)** - Architecture overview and general concepts
- **[Claude Desktop Integration](claude-desktop-integration.md)** - Setup and usage for Claude Desktop (HTTP transport)
- **[Claude Code Integration](claude-code-integration.md)** - Setup and usage for Claude Code (stdio transport)

### Setup and Installation

#### Automated Setup

```bash
# Complete setup for both platforms
./scripts/setup-claude-integration.sh

# Claude Desktop only
./scripts/setup-claude-integration.sh --desktop

# Claude Code only
./scripts/setup-claude-integration.sh --code

# Project-specific Claude Code setup
./scripts/setup-claude-code-integration.sh
```

#### Manual Setup

1. **Prerequisites**
   - go-invoice CLI installed and configured
   - Claude Desktop or Claude Code installed
   - Go 1.21+ for building MCP server

2. **Build MCP Server**
   ```bash
   go build -o bin/go-invoice-mcp ./cmd/mcp-server
   ```

3. **Platform Configuration**
   - See [Claude Desktop Integration](claude-desktop-integration.md#manual-setup)
   - See [Claude Code Integration](claude-code-integration.md#manual-setup)

## Key Features

### Natural Language Interface

Interact with go-invoice using natural language:

```
Create an invoice for Acme Corporation for consulting work in January.
Use 40 hours at $150/hour with 8.5% tax.
```

### Dual Transport Support

- **HTTP Transport**: Claude Desktop integration for application-level management
- **stdio Transport**: Claude Code integration for project-level management

### Advanced Capabilities

#### Claude Code Features

- **Slash Commands**: `/mcp__go_invoice__create_invoice`
- **Resource Mentions**: `@invoice:INV-2025-001`, `@client:"Acme Corp"`
- **Project Integration**: Per-project configuration and data isolation
- **Workspace Watching**: Automatic detection of timesheet changes

#### Claude Desktop Features

- **Tool Categories**: Organized tool groupings for better UX
- **Global Management**: Cross-project invoice and client management
- **Background Service**: Always-available invoice management
- **Health Monitoring**: Built-in system health checks

## Configuration

### Configuration Files

```
configs/
├── mcp-config.json                 # Main MCP server configuration
├── logging.yaml                    # Logging configuration
├── claude-desktop/
│   ├── mcp_servers.json           # Claude Desktop server registration
│   └── tools_config.json          # Tool categorization
└── claude-code/
    ├── mcp_config.json            # Global Claude Code configuration
    ├── project_config.json        # Project template
    └── .claude_config.json.example # Example project config
```

### Project Structure

For Claude Code integration, projects follow this structure:

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

## Usage Examples

### Creating Invoices

#### Claude Desktop
```
Create an invoice for TechCorp Solutions for Q1 2025 consulting.
Include 120 hours at $175/hour.
Add 8.5% tax and set payment terms to 30 days.
```

#### Claude Code
```
/mcp__go_invoice__create_invoice

Create invoice for @client:"TechCorp Solutions" 
Import hours from @timesheet:./timesheets/q1-hours.csv
Use $175/hour rate with 8.5% tax
```

### Managing Clients

```
Add a new client:
- Name: Acme Corporation
- Address: 123 Business Ave, San Francisco, CA 94105
- Email: billing@acme.com
- Default rate: $150/hour
- Payment terms: Net 30
```

### Importing Timesheets

```
Import the CSV file at ~/Documents/january-timesheet.csv
Parse entries by client and create separate invoices
Set invoice date to last day of timesheet period
```

### Generating Documents

```
Generate HTML invoice for @invoice:INV-2025-001
Use template ./templates/branded-invoice.html
Save to ./invoices/sent/ directory
```

## Available Tools

### Invoice Management
- `invoice_create` - Create new invoices
- `invoice_list` - List invoices with filtering
- `invoice_show` - Display invoice details
- `invoice_update` - Modify existing invoices

### Client Management
- `client_create` - Add new clients
- `client_list` - List all clients
- `client_show` - Display client information
- `client_update` - Modify client details

### Data Import/Export
- `import_csv` - Import timesheet data from CSV
- `generate_html` - Generate HTML invoices
- `generate_pdf` - Generate PDF invoices
- `export_data` - Export invoice data

### Configuration
- `show_config` - Display current configuration
- `invoice_summary` - Show project summary

## Security

### Sandbox Protection

The MCP server operates within a security sandbox:

- **Command Whitelisting**: Only go-invoice commands allowed
- **Path Restrictions**: Access limited to invoice directories
- **File Size Limits**: Maximum 50MB file operations
- **Execution Timeouts**: 5-minute maximum execution time

### Audit Logging

All operations are logged for compliance and debugging:

```json
{
  "timestamp": "2025-01-20T10:30:00Z",
  "user": "desktop-user",
  "action": "invoice_create",
  "client": "Acme Corp",
  "result": "success",
  "duration": "250ms"
}
```

## Performance

### Optimization Features

- **Caching**: Tool metadata and configuration caching
- **Concurrency**: Up to 5 concurrent tool executions
- **Health Monitoring**: Continuous system health checks
- **Resource Limits**: Memory and CPU usage controls

### Monitoring

#### Health Endpoints (HTTP transport)
- `GET /health` - Basic health status
- `GET /health/detailed` - Detailed system information
- `GET /metrics` - Performance metrics

#### Log Analysis
```bash
# View error rate
grep "ERROR" ~/.go-invoice/logs/mcp-server.log | wc -l

# Monitor response times  
grep "duration" ~/.go-invoice/logs/mcp-server.log | awk '{print $NF}'

# Most used tools
grep "tool_call" ~/.go-invoice/logs/mcp-server.log | cut -d' ' -f5 | sort | uniq -c
```

## Troubleshooting

### Common Issues

#### MCP Server Won't Start
1. Verify go-invoice CLI is installed: `go-invoice version`
2. Check configuration syntax: `jq . ~/.go-invoice/mcp-config.json`
3. Review permissions: `ls -la ~/.go-invoice/`
4. Check logs: `tail ~/.go-invoice/logs/mcp-server.log`

#### Tools Not Available in Claude
1. Restart Claude application completely
2. Verify MCP server configuration
3. Test transport connectivity
4. Check security sandbox settings

#### Performance Issues
1. Monitor system resources: `top -p $(pgrep go-invoice-mcp)`
2. Review concurrent operation limits
3. Check caching configuration
4. Analyze logs for bottlenecks

### Debug Mode

Enable detailed logging:

```bash
# Environment variable
export MCP_LOG_LEVEL=debug

# Or in configuration
{
  "logging": {
    "level": "debug"
  }
}
```

## Development

### Building from Source

```bash
# Build MCP server
go build -o bin/go-invoice-mcp ./cmd/mcp-server

# Run tests
go test ./internal/mcp/...

# Build for multiple platforms
make build-all
```

### Testing

#### Unit Tests
```bash
go test -v ./internal/mcp/...
```

#### Integration Tests
```bash
# Test stdio transport
echo '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}' | \
  ./bin/go-invoice-mcp --stdio

# Test HTTP transport
./bin/go-invoice-mcp --http --port 8080 &
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}'
```

### Contributing

We welcome contributions! See the main [CONTRIBUTING.md](../CONTRIBUTING.md) for:

- Code style guidelines
- Pull request process
- Issue reporting
- Testing requirements

## Support

### Documentation
- **Main Repository**: [go-invoice on GitHub](https://github.com/mrz1836/go-invoice)
- **MCP Specification**: [Model Context Protocol](https://spec.modelcontextprotocol.io/)
- **Claude Documentation**: [Anthropic Claude Docs](https://docs.anthropic.com/)

### Getting Help
- **Issues**: [GitHub Issues](https://github.com/mrz1836/go-invoice/issues)
- **Discussions**: [GitHub Discussions](https://github.com/mrz1836/go-invoice/discussions)
- **Wiki**: [Project Wiki](https://github.com/mrz1836/go-invoice/wiki)

### Community
- **Discord**: [Join our Discord](https://discord.gg/go-invoice) *(if available)*
- **Slack**: [Slack Workspace](https://go-invoice.slack.com) *(if available)*

## Changelog

### Version 1.0.0 (2025-01-20)

#### New Features
- Initial MCP server implementation
- Dual transport support (HTTP + stdio)
- Claude Desktop integration
- Claude Code integration with project-scope support
- Comprehensive security sandbox
- Health monitoring system
- Automated setup scripts

#### Tool Categories
- Invoice Management (5 tools)
- Client Management (4 tools)  
- Data Import/Export (4 tools)
- Configuration (2 tools)

#### Security Features
- Command whitelisting
- Path restrictions
- File size limits
- Execution timeouts
- Audit logging

## License

This MCP integration is part of the go-invoice project and follows the same license terms. See [LICENSE](../LICENSE) for details.

---

**Next Steps**: Choose your platform and follow the appropriate integration guide:

- **[Claude Desktop Users](claude-desktop-integration.md)** - Application-level invoice management
- **[Claude Code Users](claude-code-integration.md)** - Project-level invoice management
- **[Developers](mcp-integration.md)** - Architecture and technical details