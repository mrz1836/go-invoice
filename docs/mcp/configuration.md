# Configuration Guide

This comprehensive guide covers all configuration options for the go-invoice MCP integration, including security settings, performance optimization, and advanced features.

## Overview

The go-invoice MCP server uses a layered configuration system:

1. **Business Configuration** - Your company details and invoice settings
2. **MCP Server Configuration** - Transport, security, and logging settings
3. **Claude Configuration** - Client-specific MCP server setup
4. **Runtime Configuration** - Environment variables and runtime options

## Business Configuration

Configure your business information and invoice defaults in `~/.go-invoice/.env.config`:

### Core Business Information

```bash
# Required: Basic business details
BUSINESS_NAME="Your Company Name"
BUSINESS_ADDRESS="123 Business Street\nCity, State 12345\nCountry"
BUSINESS_EMAIL="billing@yourcompany.com"
BUSINESS_PHONE="+1 (555) 123-4567"

# Optional: Additional business details
BUSINESS_WEBSITE="https://yourcompany.com"
BUSINESS_LOGO_PATH="~/.go-invoice/logo.png"
BUSINESS_TAGLINE="Professional Services Excellence"
```

### Tax and Legal Information

```bash
# Tax identification numbers
BUSINESS_TAX_ID="12-3456789"          # US Federal EIN
BUSINESS_VAT_ID="GB123456789"         # EU VAT number (if applicable)
BUSINESS_GST_NUMBER="123456789"       # GST number (if applicable)

# Legal structure and registration
BUSINESS_REGISTRATION_NUMBER="REG123456"
BUSINESS_LICENSE_NUMBER="LIC789012"
```

### Payment Configuration

```bash
# Payment terms and instructions
PAYMENT_TERMS="Net 30"
PAYMENT_INSTRUCTIONS="Payment due within 30 days of invoice date. Late fees may apply after due date."

# Accepted payment methods
PAYMENT_METHODS="Bank Transfer, Check, Credit Card"
PAYMENT_LATE_FEE="2.5%"               # Monthly late fee percentage
PAYMENT_DISCOUNT_TERMS="2/10 Net 30"  # Early payment discount
```

### Banking Information

```bash
# Primary bank account
BANK_NAME="Your Bank Name"
BANK_ACCOUNT="****1234"               # Masked account number
BANK_ROUTING="123456789"              # US routing number
BANK_IBAN="GB29 NWBK 6016 1331 9268 19"  # International IBAN
BANK_SWIFT="ABCDGB2L"                 # SWIFT/BIC code

# Secondary accounts (optional)
BANK_NAME_2="Secondary Bank"
BANK_ACCOUNT_2="****5678"
```

### Invoice Defaults

```bash
# Invoice numbering
INVOICE_PREFIX="INV"                  # Invoice number prefix
INVOICE_START_NUMBER=1000             # Starting invoice number
INVOICE_DUE_DAYS=30                   # Default days until due

# Currency and tax settings
CURRENCY="USD"                        # ISO currency code
VAT_RATE=0.0                         # VAT/tax rate (0.08 = 8%)
TAX_INCLUSIVE=false                   # Whether prices include tax

# Invoice content defaults
INVOICE_FOOTER="Thank you for your business!"
INVOICE_TERMS="Payment terms apply as stated. Please remit payment promptly."
```

### Storage Configuration

```bash
# Data storage locations
DATA_DIR="~/.go-invoice/data"         # Primary data directory
BACKUP_DIR="~/.go-invoice/backups"    # Backup storage location
TEMPLATE_DIR="~/.go-invoice/templates" # Custom template directory
EXPORT_DIR="~/Documents/Invoices"     # Export destination

# Backup settings
AUTO_BACKUP=true                      # Enable automatic backups
BACKUP_INTERVAL="12h"                 # Backup frequency
RETENTION_DAYS=365                    # Days to retain backups
BACKUP_COMPRESS=true                  # Compress backup files
```

### Localization

```bash
# Language and locale settings
LANGUAGE="en"                         # Language code
LOCALE="en_US.UTF-8"                 # System locale
TIMEZONE="America/New_York"           # Timezone for dates
DATE_FORMAT="2006-01-02"             # Go date format
CURRENCY_SYMBOL="$"                   # Currency display symbol
```

## MCP Server Configuration

Configure the MCP server behavior in `~/.go-invoice/mcp-config.json`:

### Basic Server Settings

```json
{
  "server": {
    "name": "go-invoice-mcp",
    "version": "1.0.0",
    "description": "Natural language invoice management",
    "author": "go-invoice",
    "documentation": "https://github.com/mrz1836/go-invoice/docs/mcp/"
  }
}
```

### Transport Configuration

#### stdio Transport (Claude Code)
```json
{
  "transport": {
    "type": "stdio",
    "readTimeout": "30s",
    "writeTimeout": "30s",
    "maxMessageSize": 10485760,
    "enableLogging": true,
    "logLevel": "info",
    "bufferSize": 8192
  }
}
```

#### HTTP Transport (Claude Desktop)
```json
{
  "transport": {
    "type": "http",
    "host": "localhost",
    "port": 0,
    "readTimeout": "30s",
    "writeTimeout": "30s",
    "maxMessageSize": 10485760,
    "enableLogging": true,
    "logLevel": "info",
    "enableCORS": false,
    "allowedOrigins": ["http://localhost"]
  }
}
```

### Security Configuration

```json
{
  "security": {
    "enableSandbox": true,
    "allowedPaths": [
      "~/.go-invoice",
      "~/Documents/Invoices",
      "~/Downloads",
      "~/Desktop"
    ],
    "blockedPaths": [
      "/etc",
      "/var/log",
      "/usr/bin",
      "/sbin",
      "/root"
    ],
    "environmentWhitelist": [
      "HOME",
      "USER",
      "PATH",
      "GO_INVOICE_HOME",
      "MCP_LOG_LEVEL",
      "MCP_TRANSPORT"
    ],
    "maxFileSize": "100MB",
    "maxConcurrentOperations": 10,
    "enableInputValidation": true,
    "enableOutputSanitization": true
  }
}
```

### Logging Configuration

```json
{
  "logging": {
    "enabled": true,
    "level": "info",
    "file": "~/.go-invoice/mcp-server.log",
    "maxSize": "50MB",
    "maxBackups": 10,
    "maxAge": 30,
    "enableConsole": false,
    "enableStructured": true,
    "enableTimestamp": true,
    "enableCaller": false
  }
}
```

### Feature Configuration

```json
{
  "features": {
    "enableResourcePatterns": true,
    "enableSlashCommands": true,
    "enableProjectScope": true,
    "enableFileWatching": false,
    "enableCaching": true,
    "cacheSize": "100MB",
    "cacheTTL": "1h"
  }
}
```

## Claude Configuration

### Claude Desktop Configuration

For Claude Desktop (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "go-invoice": {
      "name": "Go Invoice MCP Server",
      "description": "Natural language invoice management",
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
      "timeout": 30000,
      "retryLimit": 3,
      "retryDelay": 1000,
      "metadata": {
        "version": "1.0.0",
        "author": "go-invoice",
        "documentation": "https://github.com/mrz1836/go-invoice/docs/mcp/"
      }
    }
  }
}
```

### Claude Code Configuration

For Claude Code (`~/.config/claude/mcp/go-invoice.json`):

```json
{
  "mcpServers": {
    "go-invoice": {
      "name": "Go Invoice MCP Server",
      "description": "Natural language invoice management for developers",
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
            "description": "Reference invoices by ID or criteria",
            "examples": [
              "@invoice:INV-2025-001",
              "@invoice:latest",
              "@invoice:unpaid",
              "@invoice:client:\"Acme Corp\""
            ]
          },
          "@client": {
            "description": "Reference clients by name or ID",
            "examples": [
              "@client:\"Acme Corporation\"",
              "@client:CL-001",
              "@client:active",
              "@client:overdue"
            ]
          },
          "@timesheet": {
            "description": "Reference timesheet files and data",
            "examples": [
              "@timesheet:./hours.csv",
              "@timesheet:~/timesheets/january.csv",
              "@timesheet:format:excel"
            ]
          },
          "@config": {
            "description": "Reference configuration sections",
            "examples": [
              "@config:business",
              "@config:invoice_defaults",
              "@config:storage"
            ]
          }
        }
      },
      "timeout": 60000,
      "retryLimit": 2,
      "metadata": {
        "version": "1.0.0",
        "author": "go-invoice",
        "projectScope": true,
        "documentation": "https://github.com/mrz1836/go-invoice/docs/mcp/claude-code-setup.md"
      }
    }
  }
}
```

## Environment Variables

Control runtime behavior with environment variables:

### MCP Server Environment

```bash
# Core MCP settings
export MCP_TRANSPORT="stdio"          # Transport type: stdio or http
export MCP_LOG_LEVEL="info"           # Log level: debug, info, warn, error
export MCP_LOG_FILE="~/.go-invoice/mcp.log"  # Log file location

# Performance settings
export MCP_READ_TIMEOUT="30s"         # Read operation timeout
export MCP_WRITE_TIMEOUT="30s"        # Write operation timeout
export MCP_MAX_MESSAGE_SIZE="10485760" # Maximum message size in bytes
export MCP_BUFFER_SIZE="8192"         # I/O buffer size

# Security settings
export MCP_ENABLE_SANDBOX="true"      # Enable security sandbox
export MCP_MAX_FILE_SIZE="104857600"  # Maximum file size (100MB)
export MCP_MAX_CONCURRENT_OPS="10"    # Max concurrent operations

# Feature flags
export MCP_ENABLE_CACHING="true"      # Enable response caching
export MCP_CACHE_SIZE="104857600"     # Cache size in bytes
export MCP_CACHE_TTL="3600"           # Cache TTL in seconds
export MCP_ENABLE_METRICS="true"      # Enable performance metrics
```

### go-invoice Environment

```bash
# Application settings
export GO_INVOICE_HOME="~/.go-invoice"    # Home directory
export GO_INVOICE_CONFIG="~/.go-invoice/config.yaml"  # Config file
export GO_INVOICE_DATA_DIR="~/.go-invoice/data"       # Data directory

# Debug and development
export GO_INVOICE_DEBUG="false"       # Enable debug mode
export GO_INVOICE_PROFILE="false"     # Enable profiling
export GO_INVOICE_TRACE="false"       # Enable tracing

# Integration settings
export GO_INVOICE_MCP_ENABLED="true"  # Enable MCP integration
export GO_INVOICE_MCP_TOOLS="all"     # Available tools: all, basic, advanced
```

## Performance Configuration

### Optimization Settings

#### For High-Volume Usage
```json
{
  "transport": {
    "readTimeout": "120s",
    "writeTimeout": "120s",
    "maxMessageSize": 52428800,
    "bufferSize": 65536
  },
  "security": {
    "maxConcurrentOperations": 50
  },
  "features": {
    "enableCaching": true,
    "cacheSize": "500MB",
    "cacheTTL": "30m"
  },
  "logging": {
    "level": "warn",
    "enableConsole": false
  }
}
```

#### For Low-Resource Systems
```json
{
  "transport": {
    "readTimeout": "10s",
    "writeTimeout": "10s",
    "maxMessageSize": 1048576,
    "bufferSize": 2048
  },
  "security": {
    "maxConcurrentOperations": 3
  },
  "features": {
    "enableCaching": false
  },
  "logging": {
    "level": "error",
    "maxSize": "10MB",
    "maxBackups": 3
  }
}
```

### Memory Optimization

```bash
# Memory limits
export MCP_MEMORY_LIMIT="512MB"       # Process memory limit
export MCP_HEAP_SIZE="256MB"          # Heap size limit
export MCP_STACK_SIZE="8MB"           # Stack size limit

# Garbage collection tuning
export GOGC=100                        # GC trigger percentage
export GOMEMLIMIT="512MiB"            # Go memory limit
```

## Security Configuration

### Sandbox Settings

```json
{
  "security": {
    "enableSandbox": true,
    "sandboxMode": "strict",
    "allowedPaths": [
      "~/.go-invoice",
      "~/Documents/Invoices",
      "~/Downloads"
    ],
    "blockedPaths": [
      "/etc",
      "/var",
      "/usr/bin",
      "/sbin",
      "/boot",
      "/sys",
      "/proc"
    ],
    "allowedExtensions": [
      ".csv",
      ".json",
      ".html",
      ".pdf",
      ".txt",
      ".md"
    ],
    "blockedExtensions": [
      ".exe",
      ".bat",
      ".sh",
      ".bin"
    ]
  }
}
```

### Access Control

```json
{
  "security": {
    "enableUserValidation": true,
    "allowedUsers": ["invoice-user", "admin"],
    "enableGroupValidation": true,
    "allowedGroups": ["invoice", "users"],
    "enableIPRestriction": false,
    "allowedIPs": ["127.0.0.1", "::1"],
    "enableRateLimit": true,
    "rateLimitRequests": 100,
    "rateLimitWindow": "1m"
  }
}
```

### Encryption Settings

```json
{
  "security": {
    "enableEncryption": true,
    "encryptionKey": "~/.go-invoice/encryption.key",
    "encryptionAlgorithm": "AES-256-GCM",
    "enableDataEncryption": true,
    "enableLogEncryption": false,
    "enableBackupEncryption": true
  }
}
```

## Development Configuration

### Debug Settings

```json
{
  "development": {
    "enableDebug": true,
    "enableProfiler": true,
    "enableTracing": true,
    "enableMetrics": true,
    "metricsPort": 9090,
    "profilerPort": 6060,
    "enableReload": true,
    "reloadPatterns": ["*.json", "*.yaml"]
  }
}
```

### Testing Configuration

```json
{
  "testing": {
    "enableTestMode": true,
    "testDataDir": "~/.go-invoice/test-data",
    "enableMockData": true,
    "mockClients": 10,
    "mockInvoices": 50,
    "enableValidation": true,
    "strictValidation": false
  }
}
```

## Configuration Validation

### Validate Business Configuration
```bash
# Validate business settings
go-invoice config validate

# Check specific sections
go-invoice config validate --section business
go-invoice config validate --section invoice
go-invoice config validate --section storage
```

### Validate MCP Configuration
```bash
# Validate MCP server config
go-invoice-mcp --config ~/.go-invoice/mcp-config.json --validate

# Test configuration with dry run
go-invoice-mcp --config ~/.go-invoice/mcp-config.json --dry-run

# Validate Claude configuration
claude mcp --validate-config ~/.config/claude/mcp/go-invoice.json
```

### Configuration Schema

The configuration files follow JSON Schema specifications. Validation schemas are available:

- [Business Configuration Schema](../schemas/business-config.schema.json)
- [MCP Server Configuration Schema](../schemas/mcp-server-config.schema.json)
- [Claude Configuration Schema](../schemas/claude-config.schema.json)

## Configuration Templates

### Small Business Template
```bash
# Copy template configuration
cp ~/.go-invoice/templates/small-business.env ~/.go-invoice/.env.config
cp ~/.go-invoice/templates/small-business-mcp.json ~/.go-invoice/mcp-config.json
```

### Enterprise Template
```bash
# Copy enterprise configuration
cp ~/.go-invoice/templates/enterprise.env ~/.go-invoice/.env.config
cp ~/.go-invoice/templates/enterprise-mcp.json ~/.go-invoice/mcp-config.json
```

### Developer Template
```bash
# Copy developer configuration
cp ~/.go-invoice/templates/developer.env ~/.go-invoice/.env.config
cp ~/.go-invoice/templates/developer-mcp.json ~/.go-invoice/mcp-config.json
```

## Troubleshooting Configuration

### Common Configuration Issues

1. **File Permissions**
   ```bash
   # Fix configuration file permissions
   chmod 600 ~/.go-invoice/.env.config
   chmod 600 ~/.go-invoice/mcp-config.json
   chmod 600 ~/.config/claude/mcp/go-invoice.json
   ```

2. **Path Resolution**
   ```bash
   # Test path resolution
   go-invoice config paths --test

   # Expand paths
   go-invoice config paths --expand
   ```

3. **Environment Variables**
   ```bash
   # Check environment
   go-invoice config env --show

   # Validate environment
   go-invoice config env --validate
   ```

### Configuration Debugging

Enable debug logging to troubleshoot configuration issues:

```bash
# Enable configuration debugging
export MCP_CONFIG_DEBUG="true"
export GO_INVOICE_CONFIG_DEBUG="true"

# Run with debug output
go-invoice-mcp --config ~/.go-invoice/mcp-config.json --debug
```

## Best Practices

### 1. Security Best Practices
- Use environment variables for sensitive data
- Set restrictive file permissions (600 for config files)
- Enable sandbox mode in production
- Regularly rotate encryption keys
- Monitor access logs

### 2. Performance Best Practices
- Tune timeouts based on your workload
- Enable caching for read-heavy operations
- Use appropriate log levels (warn/error in production)
- Monitor memory usage and set limits
- Enable compression for large datasets

### 3. Maintenance Best Practices
- Backup configuration files regularly
- Version control your configurations
- Test configuration changes in development
- Document custom settings
- Review security settings regularly

### 4. Development Best Practices
- Use templates for consistent configurations
- Validate configurations before deployment
- Enable debug logging during development
- Use environment-specific configurations
- Test with production-like data volumes

Ready to optimize your go-invoice MCP integration? Use these configuration options to customize the system for your specific needs and environment.
