# 💸 go-invoice
> AI-powered invoice management through natural conversation with Claude Desktop and Code.

<table>
  <thead>
    <tr>
      <th>CI&nbsp;/&nbsp;CD</th>
      <th>Quality&nbsp;&amp;&nbsp;Security</th>
      <th>Docs&nbsp;&amp;&nbsp;Meta</th>
      <th>Community</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td valign="top" align="left">
        <a href="https://github.com/mrz1836/go-invoice/releases">
          <img src="https://img.shields.io/github/release-pre/mrz1836/go-invoice?logo=github&style=flat&v=1" alt="Latest Release">
        </a><br/>
        <a href="https://github.com/mrz1836/go-invoice/actions">
          <img src="https://img.shields.io/github/actions/workflow/status/mrz1836/go-invoice/fortress.yml?branch=master&logo=github&style=flat" alt="Build Status">
        </a><br/>
        <a href="https://github.com/mrz1836/go-invoice/commits/master">
          <img src="https://img.shields.io/github/last-commit/mrz1836/go-invoice?style=flat&logo=clockify&logoColor=white" alt="Last commit">
        </a>
      </td>
      <td valign="top" align="left">
        <a href="https://goreportcard.com/report/github.com/mrz1836/go-invoice">
          <img src="https://goreportcard.com/badge/github.com/mrz1836/go-invoice?style=flat&v=1" alt="Go Report Card">
        </a><br/>
        <a href="https://app.codecov.io/gh/mrz1836/go-invoice/tree/master">
          <img src="https://codecov.io/gh/mrz1836/go-invoice/branch/master/graph/badge.svg?style=flat" alt="Code Coverage">
        </a><br/>
        <a href="https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck">
          <img src="https://img.shields.io/badge/security-govulncheck-blue?style=flat&logo=springsecurity&logoColor=white" alt="Security Scanning">
        </a><br/>
        <a href=".github/SECURITY.md">
          <img src="https://img.shields.io/badge/security-policy-blue?style=flat&logo=springsecurity&logoColor=white" alt="Security Policy">
        </a>
      </td>
      <td valign="top" align="left">
        <a href="https://golang.org/">
          <img src="https://img.shields.io/github/go-mod/go-version/mrz1836/go-invoice?style=flat" alt="Go version">
        </a><br/>
        <a href="https://pkg.go.dev/github.com/mrz1836/go-invoice">
          <img src="https://pkg.go.dev/badge/github.com/mrz1836/go-invoice.svg?style=flat" alt="Go docs">
        </a><br/>
        <a href="docs/mcp/">
          <img src="https://img.shields.io/badge/MCP-compatible-8A2BE2?style=flat&logo=anthropic&logoColor=white" alt="MCP Compatible">
        </a><br/>
        <a href=".github/AGENTS.md">
          <img src="https://img.shields.io/badge/AGENTS.md-found-40b814?style=flat&logo=openai" alt="AI Agent Rules">
        </a><br/>
        <a href="Makefile" target="_blank">
          <img src="https://img.shields.io/badge/Makefile-supported-brightgreen?style=flat&logo=probot&logoColor=white" alt="Makefile Supported">
        </a><br/>
      </td>
      <td valign="top" align="left">
        <a href="https://github.com/mrz1836/go-invoice/graphs/contributors">
          <img src="https://img.shields.io/github/contributors/mrz1836/go-invoice?style=flat&logo=contentful&logoColor=white" alt="Contributors">
        </a><br/>
        <a href="https://github.com/sponsors/mrz">
          <img src="https://img.shields.io/badge/sponsor-mrz-181717.svg?logo=github&style=flat" alt="Sponsor">
        </a><br/>
        <a href="https://github.com/mrz1836/go-invoice/stargazers">
          <img src="https://img.shields.io/github/stars/mrz1836/go-invoice?style=social?v=1" alt="Stars">
        </a>
      </td>
    </tr>
  </tbody>
</table>

<br/>

## 🗂️ Table of Contents
* [Quick Start](#-quick-start)
* [Natural Language Interface](#-natural-language-interface)
* [Features](#-features)
* [Claude Integration](#-claude-integration)
* [Installation](#-installation)
* [Configuration](#-configuration)
* [Traditional CLI Usage](#-traditional-cli-usage)
* [CSV Import](#-csv-import)
* [Templates](#-templates)
* [Performance](#-performance)
* [Development](#-development)
* [Testing](#-testing)
* [AI Compliance](#-ai-compliance)
* [Maintainers](#-maintainers)
* [Contributing](#-contributing)
* [License](#-license)

<br/>


## 🚀 Quick Start

### Natural Language Setup (Recommended)

```bash
# 1. Install go-invoice
go install github.com/mrz/go-invoice@latest

# 2. Set up your business configuration
go-invoice config setup

# 3. Setup Claude integration
go-invoice config setup-claude

# 4. Start using natural language!
# In Claude Desktop: "Create an invoice for Acme Corp for 40 hours of development at $125/hour"
# In Claude Code: "Create an invoice for Acme Corp" (just type naturally, no commands needed)
```

<details>
<summary><strong>📋 Traditional CLI Quick Start</strong></summary>

```bash
# Install go-invoice
go install github.com/mrz/go-invoice@latest

# Set up your business configuration
go-invoice config setup

# Add your first client
go-invoice client add --name "Acme Corp" --email "billing@acme.com"

# Import timesheet data
go-invoice import csv timesheet.csv --client "Acme Corp"

# Generate an invoice
go-invoice invoice create --client "Acme Corp" --output invoice.html

# View your invoice in the browser
open invoice.html
```

</details>

<br/>

## 🤖 Natural Language Interface

**Manage your invoices through conversation with Claude Desktop and Claude Code**

### Talk to Claude Desktop
```
User: "Create an invoice for Acme Corp for website redesign, 40 hours at $125/hour"
Claude: ✅ Created invoice INV-2025-001 for Acme Corp 
        📊 Total: $5,000.00 (40 hours × $125.00)

User: "Import my timesheet.csv and generate the final HTML invoice"
Claude: ✅ Imported 14 work items totaling 112 hours
        📄 Generated invoice-2025-001.html

User: "Show me all unpaid invoices from last month"
Claude: 📋 Found 3 unpaid invoices:
        • INV-2024-015 - TechCorp ($2,400) - Due Jan 31
        • INV-2024-018 - StartupXYZ ($1,800) - Due Feb 5
        • INV-2024-021 - Consulting LLC ($3,200) - Due Feb 10
```

### Use Claude Code with Natural Language
```
User: Create an invoice for Acme Corp for 40 hours at $125/hour
Claude: ✅ Created invoice INV-2025-001 for Acme Corp
        📊 Total: $5,000.00 (40 hours × $125.00)

User: Import timesheet.csv and add it to @invoice:INV-2025-001
Claude: ✅ Imported 14 work items into invoice INV-2025-001
        📊 Updated total: $6,750.00 (54 hours total)

User: Generate the HTML for @invoice:INV-2025-001
Claude: ✅ Generated invoice-2025-001.html in current directory
```

### Traditional CLI (Also Available)
```bash
go-invoice client add --name "Acme Corp" --email "billing@acme.com"
go-invoice import csv timesheet.csv --client "Acme Corp"
go-invoice invoice create --client "Acme Corp" --output invoice.html
```

**🎯 21 MCP Tools Available:** Invoice creation, client management, CSV import, HTML generation, reporting, and more - all accessible through natural conversation.

<br/>

## ✨ Features

**🤖 Natural Language Interface**
- Manage invoices through conversation with Claude Desktop and Claude Code
- 21 MCP tools accessible via natural language commands
- Resource mentions in Claude Code (@invoice:, @client:, @timesheet:)
- Dual transport support (HTTP for Desktop, stdio for Code)

**⚡ Production Performance**
- Sub-microsecond response times (0.0008ms average)
- >1.4M operations per second throughput capability
- 100+ concurrent requests supported
- Efficient memory usage (~1KB per operation)

**🛡️ Security First**
- Local-only operation (no external network dependencies)
- 64 security test cases covering injection prevention
- Comprehensive command sandboxing and validation
- Complete audit trail for all operations

**🏢 Business Management**
- Complete business profile setup with contact information
- Configurable tax rates and payment terms
- Multi-currency support

**👥 Client Management**
- Add, edit, and manage client information
- Client contact details and billing addresses
- Client activity tracking and soft delete

**📄 Invoice Generation**
- Professional HTML invoice generation
- Automatic invoice numbering with configurable prefixes
- Tax calculation and subtotal management
- Multiple invoice statuses (draft, sent, paid, overdue, voided)

**⏱️ Time Tracking**
- CSV timesheet import from popular time tracking tools
- Manual work item entry with hours and rates
- Flexible date formats and validation
- Work item descriptions with intelligent validation

**🔧 Developer-Friendly**
- Context-first design throughout the application
- Comprehensive test coverage with security and performance validation
- Clean architecture with dependency injection
- Concurrent-safe operations
- Extensive error handling and validation

<br/>

## 🎯 Claude Integration

<details>
<summary><strong>🖥️ Claude Desktop Setup (HTTP Transport)</strong></summary>

```bash
# Run the setup command
go-invoice config setup-claude --desktop

# Or manually add to Claude Desktop config:
# ~/.config/claude-desktop/mcp_servers.json
{
  "mcpServers": {
    "go-invoice": {
      "command": "/path/to/go-invoice-mcp",
      "args": ["--transport", "http", "--port", "8080"],
      "env": {
        "DATA_DIR": "/path/to/your/invoice/data"
      }
    }
  }
}
```

**Features:**
- Full natural language conversation
- All 21 MCP tools available
- HTTP transport for reliable communication
- Automatic tool discovery and categorization

</details>

<details>
<summary><strong>💻 Claude Code Setup (stdio Transport)</strong></summary>

```bash
# Setup for current project
go-invoice config setup-claude --code

# Or manually create .mcp.json:
{
  "mcpServers": {
    "go-invoice": {
      "command": "/path/to/go-invoice-mcp",
      "args": ["--stdio"],
      "env": {
        "GO_INVOICE_HOME": "${HOME}/.go-invoice"
      }
    }
  }
}
```

**Features:**
- Natural language interface - just describe what you want
- Resource mentions: `@invoice:`, `@client:`, `@timesheet:`
- Project-scope configuration
- stdio transport for fast local communication

</details>

<details>
<summary><strong>🛠️ Platform Comparison</strong></summary>

| Feature               | Claude Desktop         | Claude Code                      |
|-----------------------|------------------------|----------------------------------|
| **Transport**         | HTTP                   | stdio                            |
| **Interface**         | Natural conversation   | Natural language + mentions      |
| **Setup**             | Global configuration   | Project-specific                 |
| **Performance**       | < 200ms                | < 100ms                          |
| **Tools Available**   | All 21 tools           | All 21 tools                     |
| **Resource Mentions** | Not supported          | @invoice:, @client:, @timesheet: |
| **Best For**          | Business conversations | Development workflows            |

</details>

<details>
<summary><strong>🔧 Troubleshooting Claude Integration</strong></summary>

**MCP Server Not Found:**
```bash
# Verify installation
go-invoice-mcp --version

# Test MCP server directly
go-invoice-mcp --transport stdio --test
```

**Connection Issues:**
```bash
# Check Claude Desktop logs
tail -f ~/.config/claude-desktop/logs/mcp.log

# Test HTTP transport
curl http://localhost:8080/mcp -d '{"jsonrpc":"2.0","method":"ping","id":"test"}'
```

**Tool Discovery Issues:**
```bash
# Verify business configuration
go-invoice config show

# Test tool discovery
go-invoice-mcp --transport stdio --list-tools
```

For detailed troubleshooting, see our [comprehensive troubleshooting guide](docs/mcp/troubleshooting.md).

</details>

<br/>

## 📦 Installation

<details>
<summary><strong>📋 Installation Options</strong></summary>

### Prerequisites
- **Go 1.24 or later** – [Download Go](https://golang.org/dl/)
- **Git** – For version control

### Install from Source

```bash
# Clone the repository
git clone https://github.com/mrz/go-invoice.git
cd go-invoice

# Build the application
make build-go

# Install globally (optional)
make install
```

### Install via Go

```bash
go install github.com/mrz/go-invoice@latest
```

### Verify Installation

```bash
go-invoice --version
```

</details>

<br/>

## ⚙️ Configuration

<details>
<summary><strong>🔧 Business Configuration</strong></summary>

### Initial Setup

Run the setup wizard to configure your business information:

```bash
go-invoice config setup
```

This will prompt you for:
- Business name and contact information
- Default tax rates and currency
- Invoice numbering preferences
- Payment terms and banking details

### Manual Configuration

Alternatively, set up using environment variables or configuration files:

```bash
# Environment variables
export BUSINESS_NAME="Your Business Name"
export BUSINESS_EMAIL="billing@yourbusiness.com"
export BUSINESS_ADDRESS="123 Business St, City, State 12345"
export PAYMENT_TERMS="Net 30"
export CURRENCY="USD"
export VAT_RATE="0.10"  # 10% tax rate
```

### Configuration File

Create a `.env` file in your working directory:

```bash
# Business Information
BUSINESS_NAME=Your Business Name
BUSINESS_ADDRESS=123 Business St\nCity, State 12345
BUSINESS_EMAIL=billing@yourbusiness.com
BUSINESS_PHONE=+1-555-0123
PAYMENT_TERMS=Net 30

# Invoice Settings
INVOICE_PREFIX=INV
INVOICE_START_NUMBER=1000
CURRENCY=USD
VAT_RATE=0.10

# Storage Settings
DATA_DIR=./data
AUTO_BACKUP=true
```

</details>

<br/>

## 🖥️ Traditional CLI Usage

<details>
<summary><strong>👥 Client Management</strong></summary>

```bash
# Add a new client
go-invoice client add \
  --name "Acme Corporation" \
  --email "billing@acme.com" \
  --address "456 Client Ave, Client City, CC 67890" \
  --phone "+1-555-0199"

# List all clients
go-invoice client list

# View client details
go-invoice client show --name "Acme Corporation"

# Update client information
go-invoice client update --name "Acme Corporation" --email "newbilling@acme.com"

# Deactivate a client (soft delete)
go-invoice client deactivate --name "Acme Corporation"
```

</details>

<details>
<summary><strong>📄 Invoice Management</strong></summary>

```bash
# Create a new invoice
go-invoice invoice create \
  --client "Acme Corporation" \
  --description "Monthly development services" \
  --output invoice-001.html

# Add work items manually
go-invoice invoice add-work \
  --invoice INV-1001 \
  --description "Frontend development" \
  --hours 8 \
  --rate 125.00 \
  --date 2024-01-15

# List all invoices
go-invoice invoice list

# Filter invoices by status
go-invoice invoice list --status sent

# Update invoice status
go-invoice invoice send --invoice INV-1001
go-invoice invoice mark-paid --invoice INV-1001

# Generate HTML output
go-invoice invoice generate --invoice INV-1001 --output invoice-1001.html
```

</details>

<details>
<summary><strong>📦 Batch Operations</strong></summary>

```bash
# Send all draft invoices
go-invoice invoice send --all-drafts

# Generate overdue report
go-invoice report overdue --format html --output overdue-report.html

# Export invoice data
go-invoice export invoices --format json --output invoices-backup.json
```

</details>

<br/>

## 📊 CSV Import

<details>
<summary><strong>📁 CSV Import Features</strong></summary>

go-invoice supports importing timesheet data from popular time tracking applications.

### Supported CSV Format

```csv
date,description,hours,rate
2024-01-15,Frontend development and testing,8.5,125.00
2024-01-16,Backend API implementation,7.25,135.00
2024-01-17,Code review and documentation,3.0,100.00
```

### Import Commands

```bash
# Import CSV timesheet
go-invoice import csv timesheet.csv \
  --client "Acme Corporation" \
  --validate

# Import with custom date format
go-invoice import csv timesheet.csv \
  --client "Acme Corporation" \
  --date-format "01/02/2006"

# Preview import without saving
go-invoice import csv timesheet.csv \
  --client "Acme Corporation" \
  --dry-run

# Import and create invoice immediately
go-invoice import csv timesheet.csv \
  --client "Acme Corporation" \
  --create-invoice \
  --output invoice.html
```

### Supported Date Formats

- **ISO Format**: `2006-01-02`
- **US Format**: `01/02/2006`, `1/2/2006`
- **European Format**: `02/01/2006`, `2/1/2006`
- **Named Months**: `Jan 2, 2006`, `January 2, 2006`

### Validation Rules

- **Hours**: Must be between 0.1 and 24.0
- **Rate**: Must be between $1.00 and $1,000.00
- **Description**: 3-500 characters, must be specific (no generic terms like "work" or "development")
- **Date**: Must be within the last 2 years and not in the future

</details>

<br/>

## 🎨 Templates

<details>
<summary><strong>🎨 Template Customization</strong></summary>

### Default Template

go-invoice includes a professional HTML template with:
- Clean, modern design
- Print-friendly layout
- Automatic tax calculations
- Professional formatting
- Company branding area

### Custom Templates

Create custom invoice templates using Go's `text/template` syntax:

```html
<!DOCTYPE html>
<html>
<head>
    <title>Invoice {{.Number}}</title>
    <style>
        /* Your custom CSS */
    </style>
</head>
<body>
    <h1>Invoice {{.Number}}</h1>
    
    <div class="business">
        <h2>{{.Config.Business.Name}}</h2>
        <p>{{.Config.Business.Address}}</p>
        <p>{{.Config.Business.Email}}</p>
    </div>
    
    <div class="client">
        <h3>Bill To:</h3>
        <p>{{.Client.Name}}</p>
        <p>{{.Client.Address}}</p>
    </div>
    
    <table class="work-items">
        <tr>
            <th>Description</th>
            <th>Hours</th>
            <th>Rate</th>
            <th>Total</th>
        </tr>
        {{range .WorkItems}}
        <tr>
            <td>{{.Description}}</td>
            <td>{{.Hours}}</td>
            <td>${{.Rate | printf "%.2f"}}</td>
            <td>${{.Total | printf "%.2f"}}</td>
        </tr>
        {{end}}
    </table>
    
    <div class="totals">
        <p>Subtotal: ${{.Subtotal | printf "%.2f"}}</p>
        <p>Tax ({{.TaxRate | printf "%.1f"}}%): ${{.TaxAmount | printf "%.2f"}}</p>
        <p><strong>Total: ${{.Total | printf "%.2f"}}</strong></p>
    </div>
</body>
</html>
```

### Using Custom Templates

```bash
# Generate invoice with custom template
go-invoice invoice generate \
  --invoice INV-1001 \
  --template custom-template.html \
  --output invoice.html
```

</details>

<br/>

## ⚡ Performance

**Production-Grade Performance Metrics:**

```
🚀 Response Times
  Simple operations: 0.0008ms average (sub-microsecond)
  Complex operations: < 100ms average  
  MCP tool execution: < 2s average

📈 Throughput
  Operations per second: >1.4M sustained
  Concurrent requests: 100+ supported
  Memory per operation: ~1KB efficient

🔒 Security Validation
  Test cases passed: 64/64 security tests
  Attack vectors blocked: Command injection, path traversal, privilege escalation
  Sandboxing: Comprehensive command and file access restrictions
```

<details>
<summary><strong>🔍 Performance Testing Details</strong></summary>

### Benchmark Results
```bash
# Run performance benchmarks
make bench

# Sample results:
BenchmarkSimpleServerOperations/ping_request-10         1486044    860.1 ns/op    1034 B/op    9 allocs/op
BenchmarkSimpleServerOperations/initialize_request-10   1494889    805.4 ns/op    1008 B/op   10 allocs/op
BenchmarkBasicTransportOperations/transport_health_check-10  9636212    112.7 ns/op
BenchmarkResponseTimeValidation/response_time_under_target-10  1480423    884.0 ns/op
```

### Security Test Coverage
- Command injection prevention: 15 test cases
- Path traversal protection: 12 test cases  
- Sandbox enforcement: 18 test cases
- Environment security: 8 test cases
- File handler security: 11 test cases

### Load Testing
- Concurrent users: Up to 100 simultaneous
- Burst capacity: 50 requests instantly
- Sustained load: 50 operations/second minimum
- Memory efficiency: Linear scaling, no leaks

</details>

<br/>

## 🛠️ Development

<details>
<summary><strong>🏗️ Project Architecture</strong></summary>

### Project Structure

```
go-invoice/
├── cmd/
│   ├── go-invoice/        # Traditional CLI application
│   └── go-invoice-mcp/    # MCP server binary
├── internal/
│   ├── cli/               # CLI interface and prompting
│   ├── config/            # Configuration management
│   ├── csv/               # CSV parsing and validation
│   ├── models/            # Domain models and types
│   ├── render/            # Template rendering
│   ├── services/          # Business logic services
│   ├── storage/           # Data persistence layer
│   └── mcp/               # MCP server implementation
│       ├── tools/         # 21 MCP tools (5 categories)
│       ├── executor/      # Secure command execution
│       ├── schemas/       # JSON Schema definitions
│       └── types/         # MCP protocol types
├── docs/mcp/              # MCP integration documentation
├── configs/               # Claude integration configs
├── scripts/               # Setup automation scripts
├── templates/             # HTML templates
└── examples/              # Usage examples
```

### Architecture Principles

- **Context-First Design**: All operations support context cancellation
- **Dependency Injection**: Services use constructor injection  
- **Interface Segregation**: Small, focused interfaces for testability
- **Error Handling**: Comprehensive error handling with context
- **Concurrent Safety**: All operations are thread-safe
- **MCP Protocol Compliance**: Full MCP 2024-11-05 specification support
- **Dual Transport**: HTTP (Claude Desktop) and stdio (Claude Code)
- **Security First**: Comprehensive sandboxing and validation

### MCP Integration Architecture

```
┌─────────────────┐    ┌─────────────────┐
│  Claude Desktop │    │   Claude Code   │
│   (HTTP/JSON)   │    │   (stdio/JSON)  │
└─────────┬───────┘    └─────────┬───────┘
          │                      │
          └──────┐        ┌──────┘
                 ▼        ▼
        ┌─────────────────────────┐
        │    MCP Server (Go)      │
        │  ┌─────────────────┐    │
        │  │  Tool Registry  │    │
        │  │   (21 tools)    │    │
        │  └─────────────────┘    │
        │  ┌─────────────────┐    │
        │  │ Secure Executor │    │
        │  │  (Sandboxed)    │    │
        │  └─────────────────┘    │
        └─────────┬───────────────┘
                  │
                  ▼
        ┌─────────────────┐
        │  go-invoice CLI │
        │   (Business     │
        │    Logic)       │
        └─────────────────┘
```

### Build Commands

```bash
# Install development dependencies
make mod-download

# Run linting
make lint

# Run all tests
make test

# Run tests with race detection
make test-race

# Run integration tests
go test -v -run TestIntegrationSuite

# Generate test coverage
make coverage

# Build binary
make build-go

# Install locally
make install

# Build MCP server
make build-mcp

# Run MCP integration tests
make test-mcp

# Run security tests
make test-security

# Run performance benchmarks
make bench
```

</details>

<br/>

## 🧪 Testing

<details>
<summary><strong>🧪 Comprehensive Test Suite</strong></summary>

### Test Coverage

This project maintains comprehensive test coverage with multiple test categories:

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: End-to-end MCP workflows for both Claude platforms
- **Security Tests**: 64 test cases covering injection prevention and sandboxing  
- **Performance Tests**: Sub-microsecond response time validation
- **Race Detection**: Concurrent safety testing

### Running Tests

```bash
# Run all tests (fast)
make test

# Run tests with race detection (slower but thorough)
make test-race

# Run integration tests only
go test -v -run TestIntegrationSuite

# Run tests with coverage report
make coverage

# Run benchmarks
make bench

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Test Categories

1. **models_test.go** - Domain model validation and business logic
2. **storage_test.go** - Data persistence and retrieval operations  
3. **services_test.go** - Business service integration
4. **csv_test.go** - CSV parsing and validation
5. **cli_test.go** - CLI interface and user interaction
6. **integration_test.go** - End-to-end workflow testing
7. **mcp/integration_test.go** - MCP workflow testing (both transports)
8. **mcp/security_test.go** - Security validation (64 test cases)
9. **mcp/performance_test.go** - Performance benchmarking

### Security Test Results
```
✅ Command Injection Prevention: 15/15 tests passed
✅ Path Traversal Protection: 12/12 tests passed  
✅ Sandbox Enforcement: 18/18 tests passed
✅ Environment Security: 8/8 tests passed
✅ File Handler Security: 11/11 tests passed
✅ Total Security Tests: 64/64 passed
```

### Performance Test Results
```
✅ Simple Operations: 0.0008ms (target: <100ms)
✅ Complex Operations: <100ms (target: <2s)
✅ Concurrent Requests: 100+ (target: minimum 5)
✅ Throughput: >1.4M ops/sec (target: 50 ops/sec)
✅ Memory Efficiency: ~1KB per operation
```

</details>

<br/>


## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](.github/CONTRIBUTING.md) for details.

### Quick Contribution Guide

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Make** your changes with tests
4. **Run** the test suite (`make test`)
5. **Commit** your changes (`git commit -m 'Add amazing feature'`)
6. **Push** to the branch (`git push origin feature/amazing-feature`)
7. **Open** a Pull Request

### Development Setup

```bash
# Clone your fork
git clone https://github.com/yourusername/go-invoice.git
cd go-invoice

# Install dependencies
go mod download

# Run tests to verify setup
make test

# Start developing!
```

<br/>

## 📄 Examples

### 🤖 Natural Language Workflow

**Freelancer Monthly Invoice with Claude Desktop:**
```
User: "I need to create an invoice for TechCorp Solutions for January 2025. 
      I worked 85 hours at $135/hour on web development."

Claude: ✅ Created invoice INV-2025-001 for TechCorp Solutions
        📊 Total: $11,475.00 (85 hours × $135.00)
        📄 Generated invoice-2025-001.html

User: "Import additional hours from my january-timesheet.csv file"

Claude: ✅ Imported 12 additional work items (15.5 hours)
        📊 Updated total: $13,567.50 (100.5 hours total)
        📄 Regenerated invoice-2025-001.html

User: "Send the invoice and show me all unpaid invoices"

Claude: ✅ Invoice INV-2025-001 marked as sent
        📋 Unpaid invoices:
        • INV-2025-001 - TechCorp Solutions ($13,567.50) - Due Feb 30
        • INV-2024-089 - StartupXYZ ($2,400.00) - Due Jan 15 (OVERDUE)
```

<details>
<summary><strong>📋 Traditional CLI Workflow</strong></summary>

### Complete Workflow Example

```bash
# 1. Set up your business
export BUSINESS_NAME="Freelance Developer"
export BUSINESS_EMAIL="billing@freelancer.dev"
export CURRENCY="USD"
export VAT_RATE="0.08"

# 2. Add a client
go-invoice client add \
  --name "TechCorp Solutions" \
  --email "accounting@techcorp.com" \
  --address "789 Tech Blvd, Innovation City, IC 54321"

# 3. Import time tracking data
go-invoice import csv january-timesheet.csv \
  --client "TechCorp Solutions" \
  --validate

# 4. Create and generate invoice
go-invoice invoice create \
  --client "TechCorp Solutions" \
  --description "January 2024 Development Services" \
  --output january-invoice.html

# 5. Send the invoice (updates status)
go-invoice invoice send --invoice INV-1001

# 6. Later, mark as paid
go-invoice invoice mark-paid --invoice INV-1001
```

### Automation Example

```bash
#!/bin/bash
# Monthly invoice automation script

CLIENT="TechCorp Solutions"
MONTH=$(date +%B-%Y)
TIMESHEET="timesheets/${MONTH}-timesheet.csv"
INVOICE_FILE="invoices/${MONTH}-invoice.html"

# Import timesheet
go-invoice import csv "$TIMESHEET" --client "$CLIENT"

# Create invoice
go-invoice invoice create \
  --client "$CLIENT" \
  --description "$MONTH Development Services" \
  --output "$INVOICE_FILE"

# Send invoice
INVOICE_ID=$(go-invoice invoice list --client "$CLIENT" --status draft --format json | jq -r '.[0].id')
go-invoice invoice send --invoice "$INVOICE_ID"

echo "Invoice $INVOICE_ID created and sent for $CLIENT"
```

</details>

<br/>

## 🔧 Troubleshooting

### Common Issues

**Configuration not found**
```bash
# Ensure configuration is set
go-invoice config show

# Or run setup wizard
go-invoice config setup
```

**CSV import fails**
```bash
# Validate CSV format first
go-invoice import csv timesheet.csv --validate --dry-run

# Check supported date formats
go-invoice help import
```

**Template rendering issues**
```bash
# Test with default template
go-invoice invoice generate --invoice INV-1001 --output test.html

# Validate custom template syntax
go-invoice template validate custom-template.html
```

### Debug Mode

Enable verbose logging for troubleshooting:

```bash
export DEBUG=true
go-invoice invoice create --client "Test Client" --verbose
```

<br/>

## 📊 Roadmap

### v1.0: Core Features ✅
- [x] CLI interface with comprehensive commands
- [x] Client management system
- [x] Invoice creation and management
- [x] CSV import with validation
- [x] Professional HTML generation
- [x] **MCP Integration** - Natural language interface for Claude Desktop and Claude Code
- [x] **21 MCP Tools** - Complete invoice management via AI conversation
- [x] **Dual Transport Support** - HTTP (Claude Desktop) and stdio (Claude Code)
- [x] **Production Security** - 64 security tests, comprehensive sandboxing
- [x] **Performance Validation** - Sub-microsecond response times
- [x] Comprehensive test coverage and documentation

### v2.0: Enhanced Features (Planned)
- [ ] PDF generation with customizable templates
- [ ] Email integration for automated invoice delivery  
- [ ] Payment tracking and reconciliation
- [ ] Recurring invoices and subscription billing
- [ ] Enhanced multi-currency support
- [ ] **Advanced MCP Tools** - Additional Claude integration features

### v3.0: Enterprise Features (Future)
- [ ] Web interface for team collaboration
- [ ] REST API endpoints for third-party integration
- [ ] Database backend for enterprise scale
- [ ] Advanced reporting and analytics
- [ ] Integration with popular accounting software (QuickBooks, Xero)
- [ ] **Claude Enterprise Integration** - Team-wide AI invoice management

<br/>

## 🤖 AI Compliance

go-invoice includes comprehensive AI assistant guidelines and native Claude integration:

- **[AGENTS.md](.github/AGENTS.md)** — Complete rules for coding style, workflows, and best practices
- **[CLAUDE.md](.github/CLAUDE.md)** — Guidelines for AI assistant integration
- **[.cursorrules](.cursorrules)** — Machine-readable policies for Cursor and similar tools
- **[sweep.yaml](.github/sweep.yaml)** — Configuration for Sweep AI code review
- **[docs/mcp/](docs/mcp)** — Complete MCP integration documentation for Claude Desktop and Claude Code

### Native Claude Integration Features
- **MCP Protocol Compliance**: Full MCP 2024-11-05 specification support
- **21 MCP Tools**: Natural language access to all invoice management features
- **Dual Platform Support**: Works with both Claude Desktop (HTTP) and Claude Code (stdio)
- **Production Security**: Local-only operation with comprehensive validation
- **Performance Optimized**: Sub-microsecond response times for AI interactions

These ensure that both AI assistants and the MCP integration follow the same high standards, maintaining code quality and security across all AI-powered interactions.

<br/>

## 👥 Maintainers

| [<img src="https://github.com/mrz1836.png" height="50" width="50" alt="Maintainer" />](https://github.com/mrz1836) |
|:------------------------------------------------------------------------------------------------------------------:|
|                                         [mrz](https://github.com/mrz1836)                                          |

<br/>

## 🤝 Contributing

We welcome contributions from the community! Please read our [contributing guidelines](.github/CONTRIBUTING.md) and [code of conduct](.github/CODE_OF_CONDUCT.md).

### How Can I Help?

All kinds of contributions are welcome! :raised_hands:

- **⭐ Star the project** to show your support
- **🐛 Report bugs** through GitHub issues
- **💡 Suggest features** with detailed use cases
- **📝 Improve documentation** with examples and clarity
- **🔧 Submit pull requests** with bug fixes or new features
- **💬 Join discussions** and help other users

<br/>

## 📝 License

[![License](https://img.shields.io/github/license/mrz1836/go-invoice.svg?style=flat&v=1)](LICENSE)

<br/>
<br/>

<div align="center">

**[⭐ Star this repo](https://github.com/mrz/go-invoice)** if you find it helpful!

Made with ❤️ by developers, for developers.

</div>
