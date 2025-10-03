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
        <a href="magefiles/magefile.go" target="_blank">
          <img src="https://img.shields.io/badge/Magefile-supported-brightgreen?style=flat&logo=probot&logoColor=white" alt="Magefile Supported">
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
go install github.com/mrz1836/go-invoice/cmd/go-invoice@latest
go install github.com/mrz1836/go-invoice/cmd/go-invoice-mcp@latest

# 2. Initialize storage and set up your business configuration
go-invoice init
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
go install github.com/mrz1836/go-invoice/cmd/go-invoice@latest

# Initialize storage and set up your business configuration
go-invoice init
go-invoice config setup

# Add your first client
go-invoice client create --name "Acme Corp" --email "billing@acme.com" --phone "+1-555-0123"

# Import timesheet data (CSV or JSON) and create invoice
go-invoice import create timesheet.csv --client "Acme Corp" --description "August 2025 Services"
# OR for JSON:
go-invoice import create timesheet.json --client "Acme Corp" --description "August 2025 Services"

# Generate HTML from the invoice
go-invoice generate invoice <invoice-id> --output invoice.html

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
go-invoice client create --name "Acme Corp" --email "billing@acme.com"
go-invoice import create timesheet.csv --client "Acme Corp" --description "Monthly Services"
go-invoice invoice update INV-2025-001 --date 2025-08-07  # Auto-calculates due date
go-invoice generate invoice INV-2025-001 --output invoice.html
```

**🎯 21 MCP Tools Available:** Invoice creation, client management, CSV/JSON import, HTML generation, reporting, and more - all accessible through natural conversation.

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
- **Cryptocurrency service fee** - Optional configurable fee for crypto payments

**👥 Client Management**
- Add, edit, and manage client information
- Client contact details and billing addresses
- Client activity tracking and soft delete

**📄 Invoice Generation**
- Professional HTML invoice generation
- Automatic invoice numbering with configurable prefixes
- Tax calculation and subtotal management
- Multiple invoice statuses (draft, sent, paid, overdue, voided)
- **Flexible line items** - Support for hourly, fixed, and quantity-based billing on the same invoice

**⏱️ Time Tracking & Billing**
- CSV timesheet import from popular time tracking tools
- **Hourly billing** - Traditional time-based work with hours × rate
- **Fixed fees** - Flat amounts for retainers, setup fees, monthly charges
- **Quantity-based pricing** - Unit pricing for materials, licenses, subscriptions
- Mixed billing models on same invoice (e.g., hourly work + monthly retainer + materials)
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

## 💰 Flexible Line Items

go-invoice supports **three types of line items** on the same invoice, giving you complete flexibility for any billing scenario:

### Line Item Types

#### 1. ⏱️ Hourly Billing (Time-based)
Traditional hourly work with automatic calculation: **Hours × Rate = Total**

```bash
# Via CLI
go-invoice invoice add-line-item INV-001 \
  --description "Development work on authentication module" \
  --hours 8 --rate 125

# Via Claude (natural language)
# "Add 8 hours of development work at $125/hour to INV-001"
```

#### 2. 💵 Fixed Amount (Flat Fees)
One-time charges, retainers, setup fees, monthly charges

```bash
# Via CLI
go-invoice invoice add-line-item INV-001 \
  --type fixed \
  --description "Monthly Retainer - August 2025" \
  --amount 2000

# Via Claude (natural language)
# "Add a $2000 monthly retainer to INV-001"
```

#### 3. 📦 Quantity-based (Unit Pricing)
Materials, licenses, subscriptions: **Quantity × Unit Price = Total**

```bash
# Via CLI
go-invoice invoice add-line-item INV-001 \
  --type quantity \
  --description "SSL certificates" \
  --quantity 3 --unit-price 50

# Via Claude (natural language)
# "Add 3 SSL certificates at $50 each to INV-001"
```

### Real-World Example: Mixed Billing

Create an invoice combining all three billing types:

```bash
# 1. Start with hourly work
go-invoice invoice add-line-item INV-001 \
  --description "Development - 40 hours" \
  --hours 40 --rate 125
  # Subtotal: $5,000

# 2. Add project setup fee
go-invoice invoice add-line-item INV-001 \
  --type fixed \
  --description "Project Setup & Configuration" \
  --amount 500
  # Subtotal: $5,500

# 3. Add materials/licenses
go-invoice invoice add-line-item INV-001 \
  --type quantity \
  --description "Development licenses (annual)" \
  --quantity 2 --unit-price 99
  # Final Total: $5,698
```

### Invoice Display

Line items are intelligently displayed based on type:

| Date  | Description                                              | Details      | Amount        |
|-------|----------------------------------------------------------|--------------|---------------|
| Aug 1 | Development - 40 hours<br><small>Hourly</small>          | 40h @ $125/h | $5,000.00     |
| Aug 1 | Project Setup & Configuration<br><small>Fixed</small>    | —            | $500.00       |
| Aug 1 | Development licenses (annual)<br><small>Quantity</small> | 2 × $99      | $198.00       |
|       |                                                          | **Total**    | **$5,698.00** |

### Benefits

✅ **Flexibility** - Mix different billing models on one invoice
✅ **Clarity** - Clear display shows exactly what was charged
✅ **Accuracy** - Automatic calculations prevent errors
✅ **Professional** - Clean, itemized invoices for clients
✅ **Backward Compatible** - Works alongside existing time-based work items

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
git clone https://github.com/mrz1836/go-invoice.git
cd go-invoice

# Build the application
magex devBuild
```

### Install via Go

```bash
go install github.com/mrz1836/go-invoice@latest
```

### Verify Installation

```bash
go-invoice --version
```

</details>

<br/>

## ⚙️ Configuration

<details>
<summary><strong>💰 Cryptocurrency Service Fee</strong></summary>

### Overview

When cryptocurrency payments are enabled (USDC or BSV), you can optionally apply a service fee to cover the costs of crypto payment processing, exchange fees, and conversion overhead.

### How It Works

The crypto service fee is configured **per-client** basis, giving you fine-grained control over which clients incur the fee.

1. **Client-Level Configuration**: Enable crypto fee when creating or updating a client
2. **Automatic Application**: When crypto payments are enabled AND the client has crypto fee enabled, the fee is automatically added to invoices
3. **Separate Line Item**: The fee appears as "Cryptocurrency Service Fee" in the invoice totals
4. **Taxable Amount**: The fee is added to the subtotal before tax calculation
5. **Clear Disclaimer**: A notice is displayed on invoices explaining how to avoid the fee

### Invoice Display

The crypto service fee appears in two places on generated invoices:

**Totals Section:**
```
Subtotal:                     $5,000.00
Cryptocurrency Service Fee:      $25.00
Tax (10%):                      $502.50
────────────────────────────────────────
Total:                        $5,527.50
```

**Payment Section:**
```
💰 Cryptocurrency Service Fee Notice:
A $25.00 service fee has been applied for cryptocurrency
payment processing and conversion.
To avoid this fee, please use ACH Bank Transfer (USD).
```

### Example: Enable for Specific Client

```bash
# 1. Create a client with crypto service fee enabled
go-invoice client create \
  --name "Acme Company" \
  --email "billing@bsvassociation.com" \
  --crypto-fee \
  --crypto-fee-amount 25.00

# 2. Create invoice for this client - fee is automatically applied
go-invoice invoice create \
  --client "Acme Company" \
  --description "Q1 2025 Consulting Services"

# 3. Generate HTML - the $25 crypto fee will be included
go-invoice generate invoice INV-2025-001
```

### Enable Crypto Payments Globally

To use crypto service fees, you must first enable cryptocurrency payments in your configuration:

```bash
# In your .env.config file
USDC_ENABLED=true
USDC_ADDRESS="0xYourUSDCWalletAddress"
# OR
BSV_ENABLED=true
BSV_ADDRESS="YourBSVWalletAddress"
```

Then enable the crypto fee for specific clients using the CLI commands above.

### Benefits

- **Per-Client Control**: Enable crypto fees only for specific clients
- **Cost Recovery**: Recover cryptocurrency exchange and processing fees
- **Transparency**: Clearly communicate fees to clients upfront
- **Flexibility**: Configurable amount per client or use global default
- **ACH Incentive**: Encourages clients to use fee-free ACH transfers

### Update Existing Client

To add crypto fee to an existing client:

```bash
go-invoice client update "Acme Company" \
  --crypto-fee \
  --crypto-fee-amount 25.00
```

</details>

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
BUSINESS_WEBSITE=yourbusiness.com
PAYMENT_TERMS=Net-30

# Invoice Settings
INVOICE_PREFIX=INV
INVOICE_START_NUMBER=1000
INVOICE_DUE_DAYS=30  # Auto-calculates due dates
CURRENCY=USD

# Tax Settings
TAX_RATE=0.10  # 10% tax
TAX_ENABLED=true

# Payment Methods
ACH_ENABLED=true
USDC_ENABLED=false
USDC_ADDRESS="0xYourUSDCWalletAddress"
BSV_ENABLED=false
BSV_ADDRESS="YourBSVWalletAddress"

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
# Add a new client with all details
go-invoice client create \
  --name "Acme Corporation" \
  --email "billing@acme.com" \
  --address "456 Client Ave, Client City, CC 67890" \
  --phone "+1-555-0199" \
  --tax-id "EIN-12-3456789" \
  --approver-contacts "John Doe, Finance Dept"

# List all clients with optional filters
go-invoice client list
go-invoice client list --active-only --name-search "Acme"

# View client details and invoice history
go-invoice client show --client "Acme Corporation" --include-invoices

# Update client information
go-invoice client update --client "Acme Corporation" --email "newbilling@acme.com"

# Deactivate a client (soft delete preserves data)
go-invoice client delete --client "Acme Corporation" --soft-delete
```

</details>

<details>
<summary><strong>📄 Invoice Management</strong></summary>

```bash
# Create a new invoice with optional work items
go-invoice invoice create \
  --client "Acme Corporation" \
  --description "August 2025 Development Services" \
  --date 2025-08-07  # Due date auto-calculated based on net terms

# Add work items to existing invoice
go-invoice invoice add-item \
  INV-2025-001 \
  --description "Frontend development" \
  --hours 8.5 \
  --rate 125.00 \
  --date 2025-08-01

# List all invoices with filters
go-invoice invoice list
go-invoice invoice list --status sent --from-date 2025-08-01
go-invoice invoice list --client "Acme" --include-summary

# Update invoice (including date which auto-updates due date)
go-invoice invoice update INV-2025-001 --date 2025-08-07
go-invoice invoice update INV-2025-001 --status sent
go-invoice invoice update INV-2025-001 --status paid

# Generate HTML invoice
go-invoice generate invoice INV-2025-001 --output invoice-august.html
go-invoice generate invoice INV-2025-001 --template professional --open
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

### Supported Import Formats

#### CSV Format
```csv
date,description,hours,rate
2025-08-01,Frontend development and testing,8.5,125.00
2025-08-02,Backend API implementation,7.25,135.00
2025-08-03,Code review and documentation,3.0,100.00
```

#### JSON Format (Array)
```json
[
  {
    "date": "2025-08-01",
    "description": "Frontend development",
    "hours": 8.5,
    "rate": 125.00
  },
  {
    "date": "2025-08-02",
    "description": "Backend API implementation",
    "hours": 7.25,
    "rate": 135.00
  }
]
```

#### JSON Format (Structured)
```json
{
  "metadata": {
    "client": "Acme Corp",
    "period": "August 2025",
    "project": "Website Redesign"
  },
  "work_items": [
    {
      "date": "2025-08-01",
      "description": "Frontend development",
      "hours": 8.5,
      "rate": 125.00
    }
  ]
}
```

### Import Commands

```bash
# Import timesheet (auto-detects CSV or JSON format) and create new invoice
go-invoice import create timesheet.csv \
  --client "Acme Corporation" \
  --description "August 2025 Services" \
  --date 2025-08-07

# Import JSON data with same command (format auto-detected)
go-invoice import create timesheet.json \
  --client "Acme Corporation" \
  --description "August 2025 Services"

# Append data to existing invoice (CSV or JSON)
go-invoice import append timesheet.csv \
  INV-2025-001 \
  --skip-duplicates

# Preview import before executing
go-invoice import preview timesheet.csv \
  --client "Acme Corporation" \
  --show-totals --show-warnings

# Validate format before importing
go-invoice import validate timesheet.json

# Import with custom configuration
go-invoice import create timesheet.csv \
  --client "Acme Corporation" \
  --default-rate 125.00 \
  --description "Monthly development work" \
  --currency USD \
  --due-days 30
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
# Generate HTML invoice with default template
go-invoice generate invoice INV-2025-001 \
  --output invoice.html

# Generate with professional template
go-invoice generate invoice INV-2025-001 \
  --template professional \
  --output invoice.html \
  --open  # Open in browser after generation

# Preview invoice generation without saving
go-invoice generate preview INV-2025-001

# List available templates
go-invoice generate templates
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
magex bench

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
# Run linting
magex lint

# Run all tests
magex test

# Run tests with race detection
magex test:race

# Run integration tests
go test -v -run TestIntegrationSuite

# Generate test coverage
magex test:cover

# Build main application only
magex buildMain

# Build MCP server only
magex buildMCP

# Build both applications (default target)
magex buildAll

# Build development version with "dev" tag
magex devBuild

# Build development versions of both binaries
magex devBuildAll

# Install main binary to $GOPATH/bin
magex installMain

# Install MCP server to $GOPATH/bin
magex installMCP

# Install both binaries to $GOPATH/bin
magex installAll

# Clean build artifacts
magex clean

# Clean all artifacts and installed binaries
magex cleanAll

# Run performance benchmarks
magex bench
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
magex test

# Run tests with race detection (slower but thorough)
magex test:race

# Run integration tests only
go test -v -run TestIntegrationSuite

# Run tests with coverage report
magex test:cover

# Run benchmarks
magex bench

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



<br/>

## 📄 Examples

### 🤖 Natural Language Workflow

<details>
<summary><strong>📋 Complete MCP Tools Reference</strong></summary>

#### Client Management Tools
- **client_create** - Create new clients with full contact details and approver contacts
- **client_list** - List and filter clients with search options
- **client_show** - Display detailed client information and invoice history
- **client_update** - Modify client information and status
- **client_delete** - Remove or deactivate clients

#### Invoice Management Tools
- **invoice_create** - Create invoices with optional work items
- **invoice_list** - List and filter invoices by status, date, client
- **invoice_show** - Display comprehensive invoice details
- **invoice_update** - Update invoice status, dates (auto-calculates due dates)
- **invoice_delete** - Remove invoices with safety checks
- **invoice_add_item** - Add work items to existing invoices
- **invoice_remove_item** - Remove work items from invoices

#### Import/Export Tools
- **import_csv** - Import timesheet data from CSV or JSON files (auto-detects format)
- **import_preview** - Preview import results without making changes
- **import_validate** - Validate CSV/JSON file structure before import
- **export_data** - Export invoices in CSV, JSON, XML, Excel formats

#### Generation & Reports Tools
- **generate_html** - Create HTML invoices with customizable templates
- **generate_summary** - Generate business reports and analytics

#### Configuration Tools
- **config_init** - Initialize system configuration
- **config_show** - Display current configuration
- **config_validate** - Validate configuration integrity

#### Example MCP Usage in Claude:
```
User: "Create an invoice for Acme Corp"
Claude: [Uses invoice_create tool]

User: "Import my timesheet.json file"
Claude: [Uses import_csv tool - auto-detects JSON format]

User: "Update invoice INV-2025-001 date to August 7th"
Claude: [Uses invoice_update tool with --date flag, auto-calculates due date]

User: "Show me all unpaid invoices from last month"
Claude: [Uses invoice_list tool with filters]
```

</details>

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
# 1. Initialize and configure go-invoice
go-invoice init
go-invoice config setup  # Interactive setup wizard

# Or create a .env.config file:
cat > .env.config << EOF
BUSINESS_NAME="Freelance Developer"
BUSINESS_EMAIL="billing@freelancer.dev"
BUSINESS_WEBSITE="freelancer.dev"
BUSINESS_PHONE="+1-555-0123"
PAYMENT_TERMS="Net-30"
CURRENCY="USD"
TAX_RATE="0.08"
TAX_ENABLED="true"
INVOICE_DUE_DAYS="30"  # Auto-calculates due dates

# Optional: Enable cryptocurrency payments
USDC_ENABLED="true"
USDC_ADDRESS="0xYourUSDCWalletAddress"
BSV_ENABLED="false"
BSV_ADDRESS=""
EOF

# 2. Add a client with full details
go-invoice client create \
  --name "TechCorp Solutions" \
  --email "accounting@techcorp.com" \
  --address "789 Tech Blvd, Innovation City, IC 54321" \
  --phone "+1-555-9876" \
  --approver-contacts "Jane Smith, CFO" \
  --tax-id "EIN-98-7654321"

# 3. Create timesheet data (CSV or JSON)
# CSV format:
cat > january-timesheet.csv << EOF
date,description,hours,rate
2025-01-15,Frontend development and UI design,8.5,125.00
2025-01-16,Backend API implementation,7.25,135.00
2025-01-17,Code review and documentation,3.0,100.00
EOF

# JSON format (alternative):
cat > january-timesheet.json << EOF
[
  {"date": "2025-01-15", "description": "Frontend development", "hours": 8.5, "rate": 125.00},
  {"date": "2025-01-16", "description": "Backend API", "hours": 7.25, "rate": 135.00}
]
EOF

# 4. Import and create invoice (format auto-detected)
INVOICE_NUMBER=$(go-invoice import create january-timesheet.csv \
  --client "TechCorp Solutions" \
  --description "January 2025 Development Services" \
  --date 2025-01-31 \
  --output json | jq -r '.number')

echo "Created invoice: $INVOICE_NUMBER"
# Output: Created invoice: INV-2025-001

# 5. Generate professional HTML invoice (with comma-separated currency)
go-invoice generate invoice $INVOICE_NUMBER \
  --output january-invoice.html \
  --template professional

# 6. Update invoice date (auto-updates due date to 30 days later)
go-invoice invoice update $INVOICE_NUMBER --date 2025-02-01
# Due date automatically becomes 2025-03-03

# 7. Mark invoice as sent
go-invoice invoice update $INVOICE_NUMBER --status sent

# 8. View invoice details
go-invoice invoice show $INVOICE_NUMBER

# 9. Later, mark as paid
go-invoice invoice update $INVOICE_NUMBER --status paid

# 10. List all invoices for the month (as a summary)
go-invoice invoice list \
  --from-date 2025-01-01 \
  --to-date 2025-01-31 \
  --include-summary
```

### Automation Example

```bash
#!/bin/bash
# Monthly invoice automation script

CLIENT="TechCorp Solutions"
MONTH=$(date +%B-%Y)
TIMESHEET="timesheets/${MONTH}-timesheet.csv"
INVOICE_FILE="invoices/${MONTH}-invoice.html"

# Import timesheet and create invoice
INVOICE_NUMBER=$(go-invoice import create "$TIMESHEET" --client "$CLIENT" --description "$MONTH Development Services" --output json | jq -r '.number')

# Generate HTML invoice
go-invoice generate invoice "$INVOICE_NUMBER" --output "$INVOICE_FILE"

# Update invoice status to sent
go-invoice invoice update "$INVOICE_NUMBER" --status sent

echo "Invoice $INVOICE_NUMBER created and sent for $CLIENT"
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

**Import fails (CSV or JSON)**
```bash
# Validate file format first (auto-detects CSV/JSON)
go-invoice import validate timesheet.csv
go-invoice import validate data.json

# Preview import to see what will happen
go-invoice import preview timesheet.csv --client "Acme Corp"

# Check supported import commands
go-invoice import --help
```

**Template rendering issues**
```bash
# Test with default template
go-invoice generate invoice INV-1001 --output test.html

# List available templates
go-invoice generate templates

# Preview generation without saving
go-invoice generate preview INV-1001
```

### Debug Mode

Enable verbose logging for troubleshooting:

```bash
# Enable debug logging
go-invoice --debug invoice create --client "Test Client"

# Or check specific command help
go-invoice [command] --help
```

<br/>

## 📊 Roadmap

### v1.0: Core Features ✅
- [x] CLI interface with comprehensive commands
- [x] Client management with approver contacts and tax IDs
- [x] Invoice creation with automatic due date calculation
- [x] CSV and JSON import with validation and preview
- [x] Professional HTML generation with comma-separated currency
- [x] Cryptocurrency payment methods (USDC, BSV)
- [x] **Cryptocurrency Service Fee** - Configurable fee for crypto payment processing
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

### Quick Contribution Guide

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Make** your changes with tests
4. **Run** the test suite (`magex test`)
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
magex test

# Start developing!
```

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

**[⭐ Star this repo](https://github.com/mrz1836/go-invoice)** if you find it helpful!

Made with ❤️ by developers, for developers.

</div>
