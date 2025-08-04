# Troubleshooting Guide

This guide covers common issues, solutions, and debugging techniques for the go-invoice MCP integration with Claude Desktop and Claude Code.

## Quick Diagnostics

### System Health Check
```bash
# Check go-invoice installation
go-invoice --version
go-invoice config validate

# Check MCP server
go-invoice-mcp --version --validate-config ~/.go-invoice/mcp-config.json

# Check Claude connection (Claude Code)
claude mcp --server go-invoice --health

# Test basic functionality
go-invoice client list
go-invoice invoice list
```

### Log Analysis
```bash
# Check MCP server logs
tail -f ~/.go-invoice/mcp-server.log

# Check Claude Code logs (if configured)
tail -f ~/.go-invoice/mcp-claude-code.log

# Check system logs for errors
grep -i "go-invoice\|mcp" /var/log/syslog
```

## Common Issues

### 1. Installation and Setup Issues

#### Problem: Command Not Found
```bash
$ go-invoice --version
bash: go-invoice: command not found
```

**Solution:**
```bash
# Check if installed
which go-invoice

# Install if missing
curl -L https://github.com/mrz1836/go-invoice/releases/latest/download/go-invoice-$(uname -s)-$(uname -m) -o go-invoice
chmod +x go-invoice
sudo mv go-invoice /usr/local/bin/

# Or use package manager
# brew install go-invoice     # macOS
# apt install go-invoice      # Debian/Ubuntu

# Verify installation
go-invoice --version
```

#### Problem: Permission Denied
```bash
$ go-invoice config validate
Permission denied: ~/.go-invoice/config.yaml
```

**Solution:**
```bash
# Fix file permissions
chmod 600 ~/.go-invoice/.env.config
chmod 700 ~/.go-invoice

# Fix ownership if needed
sudo chown -R $USER:$USER ~/.go-invoice

# Check permissions
ls -la ~/.go-invoice/
```

#### Problem: Configuration Directory Missing
```bash
$ go-invoice config validate
Error: configuration directory not found
```

**Solution:**
```bash
# Create configuration directory
mkdir -p ~/.go-invoice
chmod 700 ~/.go-invoice

# Initialize configuration
go-invoice config init

# Verify setup
go-invoice config validate
```

### 2. MCP Server Issues

#### Problem: Server Won't Start
```bash
$ go-invoice-mcp --stdio
Error: failed to initialize MCP server
```

**Diagnosis:**
```bash
# Check configuration syntax
go-invoice-mcp --config ~/.go-invoice/mcp-config.json --validate

# Test with minimal config
cat > /tmp/minimal-mcp.json << 'EOF'
{
  "server": {"name": "go-invoice-mcp"},
  "transport": {"type": "stdio"},
  "security": {"enableSandbox": false},
  "logging": {"enabled": true, "level": "debug"}
}
EOF

go-invoice-mcp --config /tmp/minimal-mcp.json --stdio
```

**Solution:**
```bash
# Fix common configuration issues
jq '.' ~/.go-invoice/mcp-config.json  # Validate JSON syntax

# Reset to default configuration
go-invoice-mcp --config ~/.go-invoice/mcp-config.json --reset-config

# Check dependencies
ldd $(which go-invoice-mcp)  # Linux
otool -L $(which go-invoice-mcp)  # macOS
```

#### Problem: HTTP Transport Issues (Claude Desktop)
```bash
Error: failed to start HTTP server: bind: address already in use
```

**Diagnosis:**
```bash
# Check port usage
netstat -tulpn | grep :8080
lsof -i :8080

# Check server logs
grep -i "http\|port\|bind" ~/.go-invoice/mcp-server.log
```

**Solution:**
```bash
# Use automatic port assignment
{
  "transport": {
    "type": "http",
    "port": 0  // Auto-assign port
  }
}

# Or specify different port
{
  "transport": {
    "type": "http", 
    "port": 8081
  }
}

# Kill process using port
sudo kill $(lsof -t -i:8080)
```

#### Problem: stdio Transport Issues (Claude Code)
```bash
Error: broken pipe
Error: unexpected EOF
```

**Diagnosis:**
```bash
# Test stdio manually
echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}}' | go-invoice-mcp --stdio

# Check process communication
ps aux | grep go-invoice-mcp
strace -p $(pgrep go-invoice-mcp) 2>&1 | grep -E "read|write"
```

**Solution:**
```bash
# Increase buffer sizes
{
  "transport": {
    "type": "stdio",
    "bufferSize": 65536,
    "readTimeout": "60s",
    "writeTimeout": "60s"
  }
}

# Check for process conflicts
pkill go-invoice-mcp
go-invoice-mcp --stdio --config ~/.go-invoice/mcp-config.json
```

### 3. Claude Desktop Integration Issues

#### Problem: MCP Server Not Detected
Claude Desktop doesn't show go-invoice tools.

**Diagnosis:**
```bash
# Check Claude Desktop configuration
# Windows: %APPDATA%\Claude\claude_desktop_config.json  
# macOS: ~/Library/Application Support/Claude/claude_desktop_config.json
# Linux: ~/.config/Claude/claude_desktop_config.json

# Validate JSON syntax
cat ~/.config/Claude/claude_desktop_config.json | jq '.'

# Check logs (if available)
# Claude Desktop > Settings > Logs
```

**Solution:**
```bash
# Ensure correct configuration format
{
  "mcpServers": {
    "go-invoice": {
      "command": "go-invoice-mcp",
      "args": ["--http", "--config", "~/.go-invoice/mcp-config.json"],
      "env": {
        "GO_INVOICE_HOME": "~/.go-invoice"
      }
    }
  }
}

# Restart Claude Desktop completely
# Kill all Claude processes and restart

# Verify command path
which go-invoice-mcp
ls -la $(which go-invoice-mcp)
```

#### Problem: Connection Timeout
```bash
Error: MCP server connection timeout
```

**Solution:**
```bash
# Increase timeout in Claude configuration
{
  "mcpServers": {
    "go-invoice": {
      "command": "go-invoice-mcp",
      "timeout": 60000,  // 60 seconds
      "retryLimit": 3,
      "retryDelay": 2000
    }
  }
}

# Optimize MCP server performance
{
  "transport": {
    "readTimeout": "30s",
    "writeTimeout": "30s"
  },
  "logging": {
    "level": "warn"  // Reduce logging overhead
  }
}
```

### 4. Claude Code Integration Issues

#### Problem: Resource Patterns Not Working
```bash
Error: resource pattern @invoice:INV-001 not recognized
```

**Diagnosis:**
```bash
# Check Claude Code MCP configuration
cat ~/.config/claude/mcp/go-invoice.json

# Test resource patterns manually
claude mcp --server go-invoice --test-resource "@invoice:list"
```

**Solution:**
```bash
# Ensure resource patterns are enabled
{
  "resources": {
    "patterns": [
      "@invoice:*",
      "@client:*", 
      "@timesheet:*",
      "@config:*"
    ]
  },
  "capabilities": {
    "resources": true
  }
}

# Verify in MCP configuration
{
  "features": {
    "enableResourcePatterns": true
  }
}
```

#### Problem: Slash Commands Not Available
Slash commands like `/mcp__go_invoice__create_invoice` don't work.

**Solution:**
```bash
# Enable slash commands in configuration
{
  "slashCommands": {
    "prefix": "/mcp__go_invoice__",
    "commands": [
      {
        "name": "create_invoice",
        "description": "Create a new invoice"
      }
    ]
  }
}

# Ensure feature is enabled
{
  "features": {
    "enableSlashCommands": true
  }
}

# Restart Claude Code
claude --restart-mcp
```

### 5. File System and Security Issues

#### Problem: Path Access Denied
```bash
Error: access to path "/home/user/Documents" denied
```

**Diagnosis:**
```bash
# Check sandbox configuration
grep -A 10 "allowedPaths" ~/.go-invoice/mcp-config.json

# Test path access
go-invoice-mcp --test-path "/home/user/Documents"
```

**Solution:**
```bash
# Add paths to allowed list
{
  "security": {
    "allowedPaths": [
      "~/.go-invoice",
      "~/Documents",
      "~/Downloads",
      "~/Desktop",
      "/tmp"
    ]
  }
}

# Or disable sandbox for testing (not recommended for production)
{
  "security": {
    "enableSandbox": false
  }
}
```

#### Problem: File Size Limit Exceeded
```bash
Error: file size exceeds limit: 150MB > 100MB
```

**Solution:**
```bash
# Increase file size limits
{
  "security": {
    "maxFileSize": "500MB"
  },
  "transport": {
    "maxMessageSize": 52428800  // 50MB
  }
}
```

### 6. Data and Business Logic Issues

#### Problem: Invalid Business Configuration
```bash
Error: business name is required
Error: invalid VAT rate: must be between 0 and 1
```

**Diagnosis:**
```bash
# Validate business configuration
go-invoice config validate

# Show current configuration
go-invoice config show
```

**Solution:**
```bash
# Fix required business fields in .env.config
BUSINESS_NAME="Your Company Name"
BUSINESS_ADDRESS="123 Business St, City, State 12345"
BUSINESS_EMAIL="billing@company.com"

# Fix VAT rate (use decimal, not percentage)
VAT_RATE=0.08  # For 8%
VAT_RATE=0.0   # For no tax

# Validate after changes
go-invoice config validate
```

#### Problem: Invoice Creation Fails
```bash
Error: client not found: "Acme Corp"
Error: invalid date format: "next month"
```

**Diagnosis:**
```bash
# Check available clients
go-invoice client list

# Test date parsing
go-invoice config test-date "next month"
```

**Solution:**
```bash
# Create client first
"Create a new client named 'Acme Corp' with email acme@example.com"

# Use supported date formats
"2024-03-15"           # ISO format
"March 15, 2024"       # Natural language
"30 days"              # Relative dates
"next Friday"          # Relative dates
```

### 7. Performance Issues

#### Problem: Slow Response Times
Operations take too long to complete.

**Diagnosis:**
```bash
# Enable performance logging
export MCP_LOG_LEVEL="debug"
export MCP_ENABLE_METRICS="true"

# Monitor resource usage
htop -p $(pgrep go-invoice-mcp)
iostat -x 1
```

**Solution:**
```bash
# Optimize configuration for performance
{
  "transport": {
    "readTimeout": "120s",
    "writeTimeout": "120s",
    "maxMessageSize": 52428800
  },
  "features": {
    "enableCaching": true,
    "cacheSize": "500MB",
    "cacheTTL": "30m"
  },
  "logging": {
    "level": "warn"  // Reduce logging overhead
  }
}

# Optimize system resources
# Increase available memory
# Use SSD storage for data directory
# Close unnecessary applications
```

#### Problem: Memory Usage High
```bash
Error: out of memory
Process killed by OOM killer
```

**Solution:**
```bash
# Set memory limits
export MCP_MEMORY_LIMIT="512MB"
export GOMEMLIMIT="512MiB"

# Optimize garbage collection
export GOGC=50  # More aggressive GC

# Configure limits in MCP config
{
  "security": {
    "maxConcurrentOperations": 5,
    "maxFileSize": "50MB"
  },
  "features": {
    "cacheSize": "100MB"
  }
}
```

## Advanced Debugging

### Enable Debug Mode

```bash
# Set debug environment variables
export MCP_LOG_LEVEL="debug"
export MCP_TRACE_ENABLED="true"
export GO_INVOICE_DEBUG="true"

# Run with debug output
go-invoice-mcp --debug --config ~/.go-invoice/mcp-config.json
```

### Trace Network Communication

#### stdio Transport
```bash
# Trace stdio communication
strace -e trace=read,write -p $(pgrep go-invoice-mcp)

# Log all stdio messages
{
  "logging": {
    "level": "debug",
    "enableConsole": true
  },
  "development": {
    "enableTracing": true,
    "traceStdio": true
  }
}
```

#### HTTP Transport
```bash
# Monitor HTTP requests
tcpdump -i lo port 8080

# Enable HTTP request logging
{
  "transport": {
    "enableLogging": true,
    "logLevel": "debug"
  },
  "logging": {
    "level": "debug"
  }
}
```

### Memory and Performance Profiling

```bash
# Enable profiling
export MCP_ENABLE_PROFILER="true"
export MCP_PROFILER_PORT="6060"

# Start with profiling
go-invoice-mcp --profile --config ~/.go-invoice/mcp-config.json

# Access profiler
go tool pprof http://localhost:6060/debug/pprof/heap
go tool pprof http://localhost:6060/debug/pprof/profile
```

### Database and Storage Debugging

```bash
# Check data directory
ls -la ~/.go-invoice/data/

# Validate data integrity
go-invoice data validate --verbose

# Backup before debugging
go-invoice data backup --output ~/.go-invoice/debug-backup.tar.gz

# Repair corrupted data
go-invoice data repair --dry-run
go-invoice data repair --force
```

## Error Code Reference

### MCP Server Error Codes

| Code | Description | Solution |
|------|-------------|----------|
| `E001` | Configuration file not found | Create configuration file |
| `E002` | Invalid JSON syntax | Validate JSON syntax |
| `E003` | Transport initialization failed | Check transport configuration |
| `E004` | Security sandbox violation | Update allowed paths |
| `E005` | Tool registration failed | Check tool definitions |
| `E006` | Input validation failed | Validate input parameters |
| `E007` | File system access denied | Check file permissions |
| `E008` | Resource limit exceeded | Increase resource limits |
| `E009` | Network communication failed | Check network connectivity |
| `E010` | Process timeout | Increase timeout values |

### Business Logic Error Codes

| Code | Description | Solution |
|------|-------------|----------|
| `B001` | Client not found | Create client first |
| `B002` | Invoice not found | Check invoice ID |
| `B003` | Invalid date format | Use supported date formats |
| `B004` | Invalid currency code | Use ISO currency codes |
| `B005` | Invalid tax rate | Use decimal format (0.08 for 8%) |
| `B006` | Duplicate invoice number | Check existing invoices |
| `B007` | Missing required field | Provide all required fields |
| `B008` | Invalid file format | Use supported file formats |
| `B009` | Data validation failed | Check data integrity |
| `B010` | Export failed | Check export directory permissions |

## Getting Help

### Log Collection
When reporting issues, collect relevant logs:

```bash
# Collect all logs
mkdir -p /tmp/go-invoice-logs
cp ~/.go-invoice/mcp-server.log /tmp/go-invoice-logs/
cp ~/.go-invoice/mcp-claude-code.log /tmp/go-invoice-logs/
go-invoice config show > /tmp/go-invoice-logs/config.txt
go-invoice --version > /tmp/go-invoice-logs/version.txt

# Create support bundle
tar -czf go-invoice-support-$(date +%Y%m%d).tar.gz -C /tmp go-invoice-logs
```

### System Information
Include system information with bug reports:

```bash
# System information
uname -a
go version
which go-invoice go-invoice-mcp
ls -la $(which go-invoice) $(which go-invoice-mcp)

# Claude information
claude --version  # For Claude Code
# Check Claude Desktop > About for version info
```

### Debug Checklist

Before reporting issues, verify:

- [ ] Latest version installed (`go-invoice --version`)
- [ ] Configuration files valid (`go-invoice config validate`)
- [ ] Permissions correct (`ls -la ~/.go-invoice/`)
- [ ] Logs checked for errors
- [ ] Network connectivity (for HTTP transport)
- [ ] File system access (for file operations)
- [ ] Resource limits adequate
- [ ] Environment variables set correctly

### Community Support

- **GitHub Issues**: Report bugs and feature requests
- **Documentation**: Comprehensive guides and references
- **Discord/Forums**: Community discussion and help
- **Email Support**: Direct support for enterprise users

### Professional Support

For enterprise deployments:
- Priority support tickets
- Custom configuration assistance
- Performance optimization consulting
- Integration support for existing systems

Most issues can be resolved using this troubleshooting guide. If you continue to experience problems, collect the debugging information above and reach out through the appropriate support channel.