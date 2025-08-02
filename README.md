# 💸 go-invoice
> A simple, fast, and efficient CLI application for managing invoices and time tracking.

[![Go version](https://img.shields.io/github/go-mod/go-version/mrz/go-invoice?style=flat)](https://golang.org/)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen?style=flat)](https://github.com/mrz/go-invoice)
[![Go Report Card](https://img.shields.io/badge/go%20report-A+-brightgreen?style=flat)](https://goreportcard.com/report/github.com/mrz/go-invoice)
[![License](https://img.shields.io/github/license/mrz/go-invoice.svg?style=flat)](LICENSE)

<br/>

## 🗂️ Table of Contents
* [Features](#-features)
* [Quick Start](#-quick-start)
* [Installation](#-installation)
* [Configuration](#-configuration)
* [Usage](#-usage)
* [CSV Import](#-csv-import)
* [Templates](#-templates)
* [Development](#-development)
* [Testing](#-testing)
* [Contributing](#-contributing)
* [License](#-license)

<br/>

## ✨ Features

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
- Comprehensive test coverage (90%+)
- Clean architecture with dependency injection
- Concurrent-safe operations
- Extensive error handling and validation

<br/>

## 🚀 Quick Start

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

<br/>

## 📦 Installation

### Prerequisites
- **Go 1.21 or later** – [Download Go](https://golang.org/dl/)
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

<br/>

## ⚙️ Configuration

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

<br/>

## 🖥️ Usage

### Client Management

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

### Invoice Management

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

### Batch Operations

```bash
# Send all draft invoices
go-invoice invoice send --all-drafts

# Generate overdue report
go-invoice report overdue --format html --output overdue-report.html

# Export invoice data
go-invoice export invoices --format json --output invoices-backup.json
```

<br/>

## 📊 CSV Import

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

<br/>

## 🎨 Templates

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

<br/>

## 🛠️ Development

### Project Structure

```
go-invoice/
├── cmd/                    # CLI application entry point
├── internal/
│   ├── cli/               # CLI interface and prompting
│   ├── config/            # Configuration management
│   ├── csv/               # CSV parsing and validation
│   ├── models/            # Domain models and types
│   ├── render/            # Template rendering
│   ├── services/          # Business logic services
│   └── storage/           # Data persistence layer
├── templates/             # HTML templates
├── examples/              # Usage examples
├── docs/                  # Documentation
└── integration_test.go    # Integration tests
```

### Architecture Principles

- **Context-First Design**: All operations support context cancellation
- **Dependency Injection**: Services use constructor injection
- **Interface Segregation**: Small, focused interfaces for testability
- **Error Handling**: Comprehensive error handling with context
- **Concurrent Safety**: All operations are thread-safe

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
```

<br/>

## 🧪 Testing

### Test Coverage

This project maintains **90%+ test coverage** with comprehensive test suites:

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test complete workflows end-to-end
- **Benchmark Tests**: Performance validation
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

<br/>

## 🔒 Security

### Validation and Sanitization

- Input validation on all user data
- SQL injection prevention (no SQL used)
- Path traversal protection for file operations
- Email format validation
- Numeric range validation for hours and rates

### Safe File Operations

- Atomic file writes with temporary files
- Proper file permissions (0644 for data files)
- Directory creation with safe permissions
- Cleanup of temporary files

### Error Handling

- No sensitive information in error messages
- Proper error logging without data exposure
- Context-aware error propagation

<br/>

## 📈 Performance

### Benchmarks

Run performance benchmarks to verify system performance:

```bash
make bench
```

### Optimization Features

- **Concurrent File Operations**: Safe concurrent access to data files
- **Atomic Writes**: Prevent data corruption during concurrent access
- **Efficient CSV Parsing**: Stream-based parsing for large files
- **Template Caching**: Compiled templates for faster rendering
- **Minimal Memory Allocation**: Efficient data structures and algorithms

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

### Phase 1: MVP ✅
- [x] Basic CLI interface
- [x] Client management
- [x] Invoice creation
- [x] CSV import
- [x] HTML generation
- [x] Test coverage 90%+

### Phase 2: Enhancements
- [ ] PDF generation
- [ ] Email integration
- [ ] Payment tracking
- [ ] Recurring invoices
- [ ] Multi-currency support

### Phase 3: Advanced Features
- [ ] Web interface
- [ ] API endpoints
- [ ] Database integration
- [ ] Advanced reporting
- [ ] Integration with accounting software

<br/>

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**[⭐ Star this repo](https://github.com/mrz/go-invoice)** if you find it helpful!

Made with ❤️ by developers, for developers.

</div>