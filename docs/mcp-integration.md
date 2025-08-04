# go-invoice MCP Integration

This document provides an overview of the Model Context Protocol (MCP) integration for go-invoice, enabling natural language invoice management through Claude Desktop and Claude Code.

## Architecture Overview

The go-invoice MCP server acts as a bridge between Claude's conversational interface and the go-invoice CLI, providing seamless invoice management through natural language commands.

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Claude        │    │   go-invoice    │    │   go-invoice    │
│   Desktop/Code  │◄──►│   MCP Server    │◄──►│   CLI           │
└─────────────────┘    └─────────────────┘    └─────────────────┘
     (HTTP/stdio)         (Transport Layer)        (Commands)
```

### Core Components

#### 1. Transport Layer (`internal/mcp/transport.go`)

Provides dual transport support with automatic detection:

- **stdio Transport**: Used by Claude Code for project-level integration
- **HTTP Transport**: Used by Claude Desktop for application-level integration

```go
type Transport interface {
    Type() TransportType
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Receive(ctx context.Context) (*MCPRequest, error)
    Send(ctx context.Context, response *MCPResponse) error
    IsHealthy(ctx context.Context) bool
}
```

#### 2. Health Monitoring (`internal/mcp/health.go`)

Comprehensive health checking system:

- CLI availability verification
- Storage accessibility checks
- Performance monitoring
- Memory usage tracking

#### 3. Security Sandbox

Enforces security constraints:

- Command whitelisting (only go-invoice commands)
- Path restrictions (limited to invoice directories)
- File size limits
- Execution timeouts

#### 4. Configuration Management

Flexible configuration system supporting:

- Global server settings
- Transport-specific configuration
- Platform-specific optimizations
- Project-level overrides

## Supported Platforms

### Claude Desktop Integration

**Transport**: HTTP  
**Configuration**: `~/Library/Application Support/Claude/mcp_servers.json`  
**Use Case**: Application-level invoice management

Features:
- Global invoice management across all projects
- Rich UI integration with tool categories
- Background service operation
- Shared client and template management

### Claude Code Integration

**Transport**: stdio  
**Configuration**: `.mcp.json` (project-level)  
**Use Case**: Project-specific invoice management

Features:
- Project-scoped invoice management
- Natural language MCP tool invocation
- Resource mentions (`@invoice:`, `@client:`, `@timesheet:`)
- Workspace file watching
- Development workflow integration

## Tool Categories

Tools are organized into logical categories for better user experience:

### Invoice Management
- `invoice_create` - Create new invoices
- `invoice_list` - List invoices with filtering
- `invoice_show` - Display invoice details
- `invoice_update` - Modify existing invoices
- `invoice_delete` - Remove invoices

### Client Management
- `client_create` - Add new clients
- `client_list` - List all clients
- `client_show` - Display client information
- `client_update` - Modify client details
- `client_delete` - Remove clients

### Data Import/Export
- `import_csv` - Import timesheet data from CSV
- `import_excel` - Import from Excel files
- `generate_html` - Generate HTML invoices
- `generate_pdf` - Generate PDF invoices
- `export_data` - Export invoice data

### Configuration
- `show_config` - Display current settings
- `update_config` - Modify configuration
- `invoice_summary` - Show project summary

## Data Flow

### 1. Request Processing

```
Claude → MCP Server → Validation → CLI Execution → Response → Claude
```

1. **Request Reception**: Transport layer receives MCP request
2. **Validation**: Security sandbox validates command and parameters
3. **CLI Execution**: go-invoice CLI command executed with parameters
4. **Response Processing**: CLI output formatted as MCP response
5. **Transport**: Response sent back through appropriate transport

### 2. Error Handling

- **Validation Errors**: Parameter validation failures
- **CLI Errors**: go-invoice command execution failures
- **Transport Errors**: Communication layer issues
- **System Errors**: File system, permissions, or resource issues

All errors are logged and returned as structured MCP error responses.

## Configuration

### Main Configuration (`~/.go-invoice/mcp-config.json`)

```json
{
  "server": {
    "name": "go-invoice-mcp",
    "version": "1.0.0",
    "transport": {
      "autoDetect": true,
      "stdio": {
        "enabled": true,
        "bufferSize": 65536
      },
      "http": {
        "enabled": true,
        "host": "127.0.0.1",
        "port": 0
      }
    }
  },
  "security": {
    "sandbox": {
      "enabled": true,
      "allowedCommands": ["go-invoice", "invoice", "client"],
      "allowedPaths": ["~/.go-invoice", "~/Documents/invoices"]
    }
  }
}
```

### Platform-Specific Configuration

#### Claude Desktop (`configs/claude-desktop/`)
- `mcp_servers.json` - Server registration
- `tools_config.json` - Tool categorization and UI preferences

#### Claude Code (`configs/claude-code/`)
- `mcp_config.json` - Global Claude Code configuration
- `project_config.json` - Project template
- `.mcp.json.example` - Example project MCP configuration

## Security

### Sandbox Security

The MCP server operates within a strict security sandbox:

#### Command Restrictions
- Only whitelisted go-invoice commands are allowed
- Parameters are validated against schemas
- No arbitrary command execution

#### Path Restrictions
- Access limited to invoice-related directories
- Blocked paths include system directories (`/etc`, `/sys`, `/proc`)
- File operations restricted to allowed extensions

#### Resource Limits
- Maximum file size enforcement (50MB default)
- Execution timeouts (5 minutes default)
- Memory usage monitoring

### Audit Logging

All operations are logged for security and compliance:

```json
{
  "timestamp": "2025-01-20T10:30:00Z",
  "user": "desktop-user",
  "transport": "http",
  "action": "invoice_create",
  "parameters": {
    "client": "Acme Corp",
    "amount": 1500.00
  },
  "result": "success",
  "duration": "250ms"
}
```

## Performance

### Optimization Features

#### Caching
- Tool metadata caching (5-minute TTL)
- Client information caching
- Configuration caching

#### Concurrency
- Maximum 5 concurrent tool executions
- Request queuing (100 request buffer)
- Non-blocking health checks

#### Monitoring
- Response time tracking
- Memory usage monitoring
- Goroutine leak detection

### Performance Metrics

Key performance indicators tracked:

- **Response Time**: P50, P95, P99 percentiles
- **Success Rate**: Percentage of successful operations
- **Memory Usage**: Current and peak memory consumption
- **Concurrent Operations**: Active tool executions

## Development

### Building the MCP Server

```bash
# Build for current platform
go build -o bin/go-invoice-mcp ./cmd/mcp-server

# Build for multiple platforms
GOOS=darwin GOARCH=amd64 go build -o bin/go-invoice-mcp-darwin-amd64 ./cmd/mcp-server
GOOS=linux GOARCH=amd64 go build -o bin/go-invoice-mcp-linux-amd64 ./cmd/mcp-server
GOOS=windows GOARCH=amd64 go build -o bin/go-invoice-mcp-windows-amd64.exe ./cmd/mcp-server
```

### Testing

#### Unit Tests
```bash
go test ./internal/mcp/...
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

#### Load Testing
```bash
# Generate load on HTTP transport
for i in {1..100}; do
  curl -X POST http://localhost:8080/mcp \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","method":"tools/list","params":{},"id":'$i'}' &
done
```

### Debugging

#### Enable Debug Logging
```bash
export MCP_LOG_LEVEL=debug
./bin/go-invoice-mcp --stdio --config ~/.go-invoice/mcp-config.json
```

#### Trace Mode
```bash
export MCP_TRACE=true
./bin/go-invoice-mcp --stdio --config ~/.go-invoice/mcp-config.json
```

## Deployment

### Automated Setup

Use the provided setup scripts:

```bash
# Full setup (both platforms)
go-invoice config setup-claude

# Claude Desktop only
go-invoice config setup-claude --desktop

# Claude Code only  
go-invoice config setup-claude --code

# Update existing installation
go-invoice config setup-claude --update
```

### Manual Deployment

1. **Build MCP Server**
   ```bash
   go build -o bin/go-invoice-mcp ./cmd/mcp-server
   ```

2. **Deploy Configuration**
   ```bash
   mkdir -p ~/.go-invoice/config
   cp configs/mcp-config.json ~/.go-invoice/config/
   ```

3. **Configure Platform**
   - Claude Desktop: Update MCP servers configuration
   - Claude Code: Create project `.mcp.json` and `.claude/settings.json`

4. **Verify Installation**
   ```bash
   ./bin/go-invoice-mcp --version
   go-invoice version
   ```

## Monitoring

### Health Endpoints

The HTTP transport exposes health endpoints:

- `GET /health` - Basic health check
- `GET /health/detailed` - Detailed system status
- `GET /metrics` - Prometheus-compatible metrics

### Log Analysis

Monitor key log patterns:

```bash
# Error rate
grep "ERROR" ~/.go-invoice/logs/mcp-server.log | wc -l

# Response times
grep "duration" ~/.go-invoice/logs/mcp-server.log | awk '{print $NF}'

# Most used tools
grep "tool_call" ~/.go-invoice/logs/mcp-server.log | cut -d' ' -f5 | sort | uniq -c
```

## Migration

### From CLI-Only Usage

1. **Backup Existing Data**
   ```bash
   cp -r ~/.go-invoice ~/.go-invoice.backup
   ```

2. **Install MCP Server**
   ```bash
   go-invoice config setup-claude
   ```

3. **Verify Data Integrity**
   ```bash
   go-invoice invoice list
   ```

### Between Platforms

Configuration and data are shared between Claude Desktop and Claude Code, allowing seamless migration between platforms.

## Troubleshooting

### Common Issues

#### Server Won't Start
1. Check go-invoice CLI installation
2. Verify configuration file syntax
3. Check directory permissions
4. Review log files for errors

#### Tools Not Available
1. Restart Claude application
2. Verify MCP server configuration
3. Check transport connectivity
4. Review security sandbox settings

#### Performance Issues
1. Monitor system resources
2. Check concurrent operation limits
3. Review caching configuration
4. Analyze log files for bottlenecks

### Support Resources

- **Documentation**: [GitHub Wiki](https://github.com/mrz1836/go-invoice/wiki)
- **Issues**: [GitHub Issues](https://github.com/mrz1836/go-invoice/issues)
- **Discussions**: [GitHub Discussions](https://github.com/mrz1836/go-invoice/discussions)
- **MCP Specification**: [MCP Docs](https://spec.modelcontextprotocol.io/)

## Future Enhancements

### Planned Features

- **Real-time Collaboration**: Multi-user invoice management
- **Advanced Templates**: Dynamic template generation
- **Workflow Automation**: Custom business rule engine
- **External Integrations**: QuickBooks, Xero, Stripe integration
- **Mobile Support**: Mobile-optimized interfaces
- **API Gateway**: RESTful API exposure

### Contributing

We welcome contributions! See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines on:

- Reporting issues
- Submitting pull requests
- Code style requirements
- Testing procedures
- Documentation standards