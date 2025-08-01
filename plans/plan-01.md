# go-invoice MVP - Implementation Plan

## Executive Summary

go-invoice is a CLI-driven invoice generation tool designed for freelancers and contractors to efficiently convert time tracking data into professional, printer-friendly invoices. The tool accepts CSV input of hours worked, stores invoice data locally in JSON format, and generates HTML invoices using Go templates.

**Key Architecture Decisions**:
- **Local-First Storage**: JSON files for simplicity and portability, no database dependencies
- **CSV Import**: Standard format for time tracking data import from any spreadsheet tool
- **Go Templates**: Native Go templating for invoice generation with full customization
- **Configuration via .env**: Simple key-value configuration for business details
- **Modular Architecture**: Clean separation between CLI, business logic, and storage layers
- **Printer-Optimized Output**: HTML/CSS designed specifically for print media
- **Extensible Design**: Architecture supports future enhancements (PDF, email, cloud storage)

## Vision Statement

go-invoice embodies the principle of simplicity without sacrificing functionality. It provides freelancers with a streamlined workflow from time tracking to professional invoice generation, all through intuitive command-line operations. The tool prioritizes:

- **Developer-First Design**: Built by developers, for developers and tech-savvy professionals
- **Data Portability**: Your data stays local and accessible in standard formats
- **Workflow Integration**: Seamlessly fits into existing time tracking workflows
- **Professional Output**: Generates invoices that look polished and business-ready
- **Minimal Dependencies**: Leverages Go's standard library wherever possible
- **Incremental Enhancement**: Start simple, add features as needed

## System Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   CSV Import    │────▶│   CLI Commands  │────▶│ Invoice Engine  │
│  (Time Data)    │     │    (Cobra)      │     │  (Business)     │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                │                         │
                                ▼                         ▼
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Configuration  │     │  JSON Storage   │     │ HTML Templates  │
│  (.env.config)  │     │   (Invoices)    │     │  (Go Templates) │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                                          │
                                                          ▼
                                                ┌─────────────────┐
                                                │  HTML Output    │
                                                │ (Print-Ready)   │
                                                └─────────────────┘
```

Configuration structure leverages environment variables for business details while maintaining invoice data in structured JSON files for easy manipulation and backup.

## Implementation Roadmap

### Phase 1: Core Infrastructure and Configuration
**Objective**: Establish project foundation with configuration management and basic CLI structure

**Implementation Agent**: Use Claude Code with go-expert-developer persona for all Go code implementation

**Implementation Steps:**
1. Fork and initialize project from go-template repository
2. Set up configuration management for .env.config file
3. Create base CLI structure using Cobra
4. Implement configuration validation and loading
5. Set up project structure following Go best practices

**Files to Create/Modify:**
- `cmd/go-invoice/main.go` - Main entry point with CLI initialization
- `internal/config/config.go` - Configuration management and validation
- `internal/config/types.go` - Configuration type definitions
- `.env.config.example` - Example configuration file
- `go.mod` - Update module name and dependencies

**Configuration Structure:**
```go
type Config struct {
    Business BusinessConfig
    Invoice  InvoiceConfig
}

type BusinessConfig struct {
    Name          string
    Address       string
    Phone         string
    Email         string
    TaxID         string
    PaymentTerms  string
    BankDetails   BankDetails
}
```

**Verification Steps:**
```bash
# 1. Build the application
go build -o go-invoice ./cmd/go-invoice

# 2. Test configuration loading
./go-invoice config validate

# 3. Display loaded configuration
./go-invoice config show

# 4. Run with example config
cp .env.config.example .env.config && ./go-invoice config show
```

**Success Criteria:**
- ✅ Project builds successfully from go-template base
- ✅ Configuration loads from .env.config file
- ✅ CLI responds to basic commands
- ✅ Configuration validation catches invalid inputs
- ✅ Help text displays properly for all commands
- ✅ Final todo: Update the @plans/plan-[number]-status.md file with the results of the implementation

### Phase 2: Data Models and Storage Layer
**Objective**: Implement invoice data models and JSON-based storage system

**Implementation Agent**: Use Claude Code with go-expert-developer persona for Go implementation

**Implementation Steps:**
1. Define invoice and work item data models
2. Implement JSON storage interface
3. Create file-based storage implementation
4. Add storage initialization and validation
5. Implement CRUD operations for invoices

**Files to Create/Modify:**
- `internal/models/invoice.go` - Invoice and work item models
- `internal/storage/interface.go` - Storage interface definition
- `internal/storage/json.go` - JSON file storage implementation
- `internal/storage/errors.go` - Storage-specific error types
- `cmd/go-invoice/cmd/init.go` - Storage initialization command

**Data Models:**
```go
type Invoice struct {
    ID          string      `json:"id"`
    Number      string      `json:"number"`
    Date        time.Time   `json:"date"`
    DueDate     time.Time   `json:"due_date"`
    Client      Client      `json:"client"`
    WorkItems   []WorkItem  `json:"work_items"`
    Status      string      `json:"status"`
    CreatedAt   time.Time   `json:"created_at"`
    UpdatedAt   time.Time   `json:"updated_at"`
}

type WorkItem struct {
    Date        time.Time `json:"date"`
    Hours       float64   `json:"hours"`
    Rate        float64   `json:"rate"`
    Description string    `json:"description"`
    Total       float64   `json:"total"`
}
```

**Verification Steps:**
```bash
# 1. Initialize storage
./go-invoice init

# 2. Verify storage directory creation
ls -la ~/.go-invoice/

# 3. Test invoice creation
./go-invoice invoice create --client "Test Client"

# 4. List invoices
./go-invoice invoice list
```

**Success Criteria:**
- ✅ Storage directory initializes correctly
- ✅ Invoice models serialize/deserialize properly
- ✅ CRUD operations work for invoices
- ✅ Storage handles concurrent access safely
- ✅ Error handling provides clear feedback
- ✅ Final todo: Update the @plans/plan-[number]-status.md file with the results of the implementation

### Phase 3: CSV Import and Work Item Management
**Objective**: Implement CSV parsing and work item management functionality

**Implementation Agent**: Use Claude Code with go-expert-developer persona

**Implementation Steps:**
1. Implement CSV parser for time tracking data
2. Create work item import logic with validation
3. Add ability to append work items to existing invoices
4. Implement work item aggregation and calculations
5. Add CSV format validation and error reporting

**Files to Create/Modify:**
- `internal/csv/parser.go` - CSV parsing logic
- `internal/csv/validator.go` - CSV format validation
- `internal/services/import.go` - Import service for work items
- `cmd/go-invoice/cmd/import.go` - Import CLI command
- `examples/timesheet.csv` - Example CSV format

**CSV Import Logic:**
```go
type CSVRow struct {
    Date        string
    Hours       string
    Rate        string
    Description string
}

func ParseTimesheet(reader io.Reader) ([]WorkItem, error) {
    // Parse CSV with validation
    // Convert to WorkItem structs
    // Calculate totals
}
```

**Verification Steps:**
```bash
# 1. Import CSV data
./go-invoice import --file examples/timesheet.csv --invoice INV-001

# 2. Verify imported data
./go-invoice invoice show INV-001

# 3. Test append functionality
./go-invoice import --file more-hours.csv --invoice INV-001 --append

# 4. Test validation
./go-invoice import --file invalid.csv --invoice INV-001
```

**Success Criteria:**
- ✅ CSV files parse correctly with proper validation
- ✅ Work items import with accurate calculations
- ✅ Append mode adds to existing invoices
- ✅ Invalid CSV data produces helpful error messages
- ✅ Import handles various CSV formats gracefully
- ✅ Final todo: Update the @plans/plan-[number]-status.md file with the results of the implementation

### Phase 4: Invoice Generation and Template System
**Objective**: Implement HTML invoice generation with customizable templates

**Implementation Agent**: Use Claude Code with template design expertise

**Implementation Steps:**
1. Create base HTML invoice template with Go template syntax
2. Implement template rendering engine
3. Add printer-friendly CSS styling
4. Create template customization system
5. Implement invoice calculation logic

**Files to Create/Modify:**
- `templates/invoice/default.html` - Default invoice template
- `templates/invoice/styles.css` - Printer-optimized styles
- `internal/render/engine.go` - Template rendering engine
- `internal/services/calculator.go` - Invoice calculation service
- `cmd/go-invoice/cmd/generate.go` - Generate command

**Template Structure:**
```html
<!DOCTYPE html>
<html>
<head>
    <style>
        @media print {
            /* Printer-specific styles */
        }
    </style>
</head>
<body>
    <div class="invoice">
        <h1>Invoice #{{.Number}}</h1>
        <!-- Invoice content with Go template syntax -->
    </div>
</body>
</html>
```

**Verification Steps:**
```bash
# 1. Generate invoice HTML
./go-invoice generate INV-001

# 2. Open in browser
open invoices/INV-001.html

# 3. Test print preview
./go-invoice generate INV-001 --preview

# 4. Generate with custom template
./go-invoice generate INV-001 --template custom
```

**Success Criteria:**
- ✅ Invoice generates valid HTML output
- ✅ Templates render with correct data
- ✅ Print layout looks professional
- ✅ Calculations display accurately
- ✅ Custom templates load correctly
- ✅ Final todo: Update the @plans/plan-[number]-status.md file with the results of the implementation

### Phase 5: Complete CLI Implementation
**Objective**: Implement full CLI functionality with all commands and options

**Implementation Agent**: Use Claude Code for CLI implementation

**Implementation Steps:**
1. Implement all invoice management commands
2. Add search and filter functionality
3. Create interactive prompts for complex operations
4. Add command aliases and shortcuts
5. Implement comprehensive help system

**Files to Create/Modify:**
- `cmd/go-invoice/cmd/list.go` - List invoices command
- `cmd/go-invoice/cmd/update.go` - Update invoice command
- `cmd/go-invoice/cmd/delete.go` - Delete invoice command
- `cmd/go-invoice/cmd/search.go` - Search functionality
- `internal/cli/prompt.go` - Interactive prompt utilities

**CLI Commands:**
```bash
go-invoice init                          # Initialize storage
go-invoice config [validate|show]        # Configuration management
go-invoice import --file <csv> --invoice <id>  # Import work items
go-invoice invoice create --client <name>      # Create invoice
go-invoice invoice list [--status <status>]    # List invoices
go-invoice invoice show <id>                   # Show invoice details
go-invoice invoice update <id> [options]       # Update invoice
go-invoice invoice delete <id>                 # Delete invoice
go-invoice generate <id> [--template <name>]   # Generate HTML
```

**Verification Steps:**
```bash
# 1. Test full workflow
./go-invoice invoice create --client "Acme Corp" --interactive

# 2. Import hours
./go-invoice import --file hours.csv --invoice INV-001

# 3. Generate invoice
./go-invoice generate INV-001

# 4. Test search
./go-invoice invoice search --client "Acme"

# 5. Test filtering
./go-invoice invoice list --status unpaid --month 2024-01
```

**Success Criteria:**
- ✅ All commands function as documented
- ✅ Interactive mode provides good UX
- ✅ Search and filters work accurately
- ✅ Error messages are helpful and actionable
- ✅ Help text is comprehensive and clear
- ✅ Final todo: Update the @plans/plan-[number]-status.md file with the results of the implementation

### Phase 6: Testing and Documentation
**Objective**: Ensure code quality with comprehensive testing and documentation

**Implementation Agent**: Use Claude Code with testing expertise

**Implementation Steps:**
1. Write unit tests for all packages
2. Create integration tests for workflows
3. Add example files and tutorials
4. Write comprehensive README
5. Create man page documentation

**Files to Create/Modify:**
- `*_test.go` - Unit tests for all packages
- `test/integration_test.go` - Integration test suite
- `README.md` - Comprehensive documentation
- `docs/tutorial.md` - Step-by-step tutorial
- `examples/` - Example files and templates

**Testing Strategy:**
```go
// Unit test example
func TestInvoiceCalculation(t *testing.T) {
    // Test invoice total calculation
    // Test tax calculation
    // Test date handling
}

// Integration test example
func TestFullInvoiceWorkflow(t *testing.T) {
    // Create invoice
    // Import CSV
    // Generate HTML
    // Verify output
}
```

**Verification Steps:**
```bash
# 1. Run unit tests
go test ./...

# 2. Run with coverage
go test -cover ./...

# 3. Run integration tests
go test ./test/...

# 4. Run linter
golangci-lint run

# 5. Build and test release
goreleaser build --snapshot --clean
```

**Success Criteria:**
- ✅ Test coverage exceeds 80%
- ✅ All critical paths have tests
- ✅ Documentation is clear and complete
- ✅ Examples demonstrate key workflows
- ✅ CI/CD pipeline passes all checks
- ✅ Final todo: Update the @plans/plan-[number]-status.md file with the results of the implementation

## Configuration Examples

### Basic Configuration
```env
# Business Information
BUSINESS_NAME="John Doe Consulting"
BUSINESS_ADDRESS="123 Main St, Suite 100, City, State 12345"
BUSINESS_PHONE="+1-555-123-4567"
BUSINESS_EMAIL="billing@johndoe.com"
BUSINESS_TAX_ID="12-3456789"

# Invoice Settings
INVOICE_PREFIX="INV"
INVOICE_START_NUMBER="1000"
PAYMENT_TERMS="Net 30"
```

### Full Configuration with Payment Details
```env
# Business Information
BUSINESS_NAME="Acme Development Services"
BUSINESS_ADDRESS="456 Tech Plaza, Floor 3, San Francisco, CA 94105"
BUSINESS_PHONE="+1-415-555-0100"
BUSINESS_EMAIL="invoices@acmedev.io"
BUSINESS_TAX_ID="94-7654321"
BUSINESS_WEBSITE="https://acmedev.io"

# Payment Information
PAYMENT_TERMS="Due upon receipt"
BANK_NAME="First National Bank"
BANK_ACCOUNT="****4567"
BANK_ROUTING="123456789"
PAYMENT_INSTRUCTIONS="Wire transfers preferred. Check payable to Acme Development Services."

# Invoice Settings
INVOICE_PREFIX="ACME"
INVOICE_START_NUMBER="2024001"
INVOICE_FOOTER="Thank you for your business!"
```

### Minimal Freelancer Configuration
```env
# Essential Information Only
BUSINESS_NAME="Jane Smith - Freelance Developer"
BUSINESS_EMAIL="jane@example.com"
PAYMENT_TERMS="Net 15"
INVOICE_PREFIX="JS"
```

### International Configuration
```env
# Business Information
BUSINESS_NAME="Global Tech Solutions Ltd"
BUSINESS_ADDRESS="10 Downing Street, London, UK SW1A 2AA"
BUSINESS_PHONE="+44-20-7123-4567"
BUSINESS_EMAIL="accounts@globaltech.uk"
BUSINESS_VAT_ID="GB123456789"

# Payment Information
PAYMENT_TERMS="30 days from invoice date"
BANK_NAME="Barclays Bank UK"
BANK_IBAN="GB82BARC20201512345678"
BANK_SWIFT="BARCGB22"
CURRENCY="GBP"

# Invoice Settings
INVOICE_PREFIX="GTS"
INVOICE_START_NUMBER="5000"
VAT_RATE="20"
```

## Implementation Timeline

- **Session 1**: Core Infrastructure (Phase 1) - 2-3 hours
- **Session 2**: Data Models and Storage (Phase 2) - 3-4 hours
- **Session 3**: CSV Import (Phase 3) - 2-3 hours
- **Session 4**: Template System (Phase 4) - 3-4 hours
- **Session 5**: CLI Commands (Phase 5) - 2-3 hours
- **Session 6**: Testing and Documentation (Phase 6) - 3-4 hours

Total estimated time: 15-21 hours across 6 focused sessions

## Success Metrics

### Functionality
- **CSV Import**: Successfully imports standard timesheet formats without data loss
- **Invoice Generation**: Produces valid, printer-friendly HTML within 1 second
- **Data Integrity**: No data corruption across 1000+ invoice operations
- **Template Rendering**: Supports custom templates with full variable substitution

### Performance
- **Startup Time**: CLI responds within 100ms for all commands
- **File Operations**: Handles 10,000 invoices without performance degradation
- **Memory Usage**: Maintains under 50MB memory footprint during operations
- **Concurrent Access**: Supports multiple simultaneous read operations

### Compatibility
- **Go Version**: Compatible with Go 1.21+
- **Platform Support**: Runs on Linux, macOS, and Windows
- **CSV Formats**: Accepts Excel, Google Sheets, and standard RFC 4180 CSV
- **Output Format**: Generates HTML5 compliant output with CSS3 styling

### Developer Experience
- **Setup Time**: From clone to first invoice in under 5 minutes
- **Documentation**: Every command documented with examples
- **Error Messages**: Clear, actionable error messages with suggested fixes
- **Extensibility**: New commands can be added in under 30 lines of code

## Conclusion

go-invoice represents a focused approach to invoice management, prioritizing developer workflow and professional output. By leveraging Go's strengths and maintaining a clean architecture, the tool provides immediate value while remaining extensible for future enhancements.

**Key improvements in this plan:**
- **Phased Approach**: Each phase delivers working functionality
- **Testing Focus**: Comprehensive testing ensures reliability
- **User-Centric Design**: Commands mirror natural workflow
- **Performance First**: Efficient operations even with large datasets
- **Future-Ready**: Architecture supports PDF, email, and cloud storage additions

This implementation follows established Go patterns:
- **Standard Project Layout** following go-template structure
- **Interface-Based Design** for pluggable components
- **Context-Aware Operations** for proper cancellation
- **Structured Logging** for debugging and monitoring
- **Configuration as Code** with validation and defaults

go-invoice positions itself as the go-to solution for developers who value simplicity, control, and professional results in their invoicing workflow.
