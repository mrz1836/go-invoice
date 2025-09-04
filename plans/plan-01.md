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

### Phase 0: Foundation Alignment (AGENTS.md Compliance)
**Objective**: Ensure implementation plan fully aligns with established conventions and standards

**Duration**: 30 minutes

**Implementation Agent**: Use Claude Code with go-expert-developer persona

**Key Alignment Areas:**
1. **Context-First Design**: All operations must accept `context.Context` as first parameter
2. **Interface Philosophy**: Follow "accept interfaces, return concrete types" pattern
3. **Error Handling Excellence**: Implement comprehensive error wrapping and context
4. **Testing Standards**: Use testify suite with table-driven tests and descriptive names
5. **Security Integration**: Include vulnerability scanning and dependency verification
6. **No Global State**: Enforce dependency injection patterns throughout

**Enhanced Architecture Principles:**
- Context flows through entire call stack for cancellation and timeout support
- Consumer-driven interface design with minimal, focused contracts
- Comprehensive error handling with actionable messages and proper wrapping
- Dependency injection eliminates global state and improves testability
- Security-first approach with automated vulnerability scanning

**Verification Steps:**
```bash
# Enhanced security and quality validation
govulncheck ./...
go mod verify
golangci-lint run
go test -race ./...
go test -cover ./...
go vet ./...
```

**Success Criteria:**
- ✅ All function signatures include context.Context as first parameter
- ✅ Interfaces defined at point of use (consumer-driven design)
- ✅ Error messages provide clear context and actionable guidance
- ✅ Test coverage exceeds 90% using testify patterns
- ✅ No security vulnerabilities detected in dependencies
- ✅ All linting passes per AGENTS.md standards
- ✅ Dependency injection used throughout (no global state)

---

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

// ConfigService demonstrates context-first design and dependency injection
type ConfigService struct {
    logger Logger
    validator Validator
}

func NewConfigService(logger Logger, validator Validator) *ConfigService {
    return &ConfigService{
        logger: logger,
        validator: validator,
    }
}

func (s *ConfigService) LoadConfig(ctx context.Context, path string) (*Config, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Implementation with proper error wrapping
    config, err := s.readConfigFile(ctx, path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config from %s: %w", path, err)
    }

    if err := s.validator.ValidateConfig(ctx, config); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }

    return config, nil
}
```

**Verification Steps:**
```bash
# 1. Enhanced security and quality checks
govulncheck ./...
go mod verify
golangci-lint run

# 2. Build the application
go build -o go-invoice ./cmd/go-invoice

# 3. Run comprehensive tests with testify
go test -v -race -cover ./...

# 4. Test configuration loading with context
./go-invoice config validate

# 5. Display loaded configuration
./go-invoice config show

# 6. Run with example config
cp .env.config.example .env.config && ./go-invoice config show

# 7. Verify context cancellation works
timeout 1s ./go-invoice config validate --slow-operation
```

**Success Criteria:**
- ✅ Project builds successfully from go-template base
- ✅ Configuration loads from .env.config file with context support
- ✅ CLI responds to basic commands and respects context cancellation
- ✅ Configuration validation catches invalid inputs with clear error messages
- ✅ Help text displays properly for all commands
- ✅ All operations accept context.Context as first parameter
- ✅ Dependency injection used throughout (no global state)
- ✅ Error handling follows AGENTS.md excellence patterns
- ✅ Tests use testify suite with descriptive names
- ✅ No security vulnerabilities in dependencies
- ✅ All linting and formatting passes
- ✅ Final todo: Update the @plans/plan-01-status.md file with the results of the implementation

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

// Storage interface defined at point of use (consumer-driven design)
type InvoiceStorage interface {
    SaveInvoice(ctx context.Context, invoice *Invoice) error
    GetInvoice(ctx context.Context, id string) (*Invoice, error)
    ListInvoices(ctx context.Context, filter InvoiceFilter) ([]*Invoice, error)
    UpdateInvoice(ctx context.Context, invoice *Invoice) error
    DeleteInvoice(ctx context.Context, id string) error
}

// Service accepts interface, returns concrete type
type InvoiceService struct {
    storage InvoiceStorage
    logger  Logger
    idGen   IDGenerator
}

func NewInvoiceService(storage InvoiceStorage, logger Logger, idGen IDGenerator) *InvoiceService {
    return &InvoiceService{
        storage: storage,
        logger:  logger,
        idGen:   idGen,
    }
}

func (s *InvoiceService) CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*Invoice, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("invalid create request: %w", err)
    }

    invoice := &Invoice{
        ID:        s.idGen.GenerateID(),
        Number:    req.Number,
        Date:      req.Date,
        DueDate:   req.DueDate,
        Client:    req.Client,
        Status:    "draft",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    if err := s.storage.SaveInvoice(ctx, invoice); err != nil {
        return nil, fmt.Errorf("failed to save invoice %s: %w", invoice.ID, err)
    }

    s.logger.Info("invoice created successfully", "id", invoice.ID, "number", invoice.Number)
    return invoice, nil
}
```

**Verification Steps:**
```bash
# 1. Run security and quality checks
govulncheck ./...
go mod verify
golangci-lint run

# 2. Run comprehensive tests with testify
go test -v -race -cover ./internal/storage/...
go test -v -race -cover ./internal/models/...

# 3. Initialize storage
./go-invoice init

# 4. Verify storage directory creation
ls -la ~/.go-invoice/

# 5. Test invoice creation with context
./go-invoice invoice create --client "Test Client"

# 6. List invoices
./go-invoice invoice list

# 7. Test context cancellation
timeout 1s ./go-invoice invoice create --client "Test" --slow-operation
```

**Success Criteria:**
- ✅ Storage directory initializes correctly
- ✅ Invoice models serialize/deserialize properly
- ✅ CRUD operations work for invoices with context support
- ✅ Storage handles concurrent access safely with proper locking
- ✅ Error handling provides clear, actionable feedback with proper wrapping
- ✅ All storage operations accept context.Context as first parameter
- ✅ Consumer-driven interfaces defined at point of use
- ✅ Dependency injection used throughout (no global state)
- ✅ Tests use testify suite with table-driven patterns
- ✅ Context cancellation works correctly for long operations
- ✅ No security vulnerabilities in dependencies
- ✅ All linting and race condition checks pass
- ✅ Final todo: Update the @plans/plan-01-status.md file with the results of the implementation

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

// CSV parser interface defined at point of use
type TimesheetParser interface {
    ParseTimesheet(ctx context.Context, reader io.Reader) ([]WorkItem, error)
}

type CSVParser struct {
    validator WorkItemValidator
    logger    Logger
}

func NewCSVParser(validator WorkItemValidator, logger Logger) *CSVParser {
    return &CSVParser{
        validator: validator,
        logger:    logger,
    }
}

func (p *CSVParser) ParseTimesheet(ctx context.Context, reader io.Reader) ([]WorkItem, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    csvReader := csv.NewReader(reader)
    rows, err := csvReader.ReadAll()
    if err != nil {
        return nil, fmt.Errorf("failed to read CSV data: %w", err)
    }

    var workItems []WorkItem
    for i, row := range rows {
        if i == 0 {
            continue // Skip header
        }

        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }

        workItem, err := p.parseRow(ctx, row)
        if err != nil {
            return nil, fmt.Errorf("failed to parse row %d: %w", i+1, err)
        }

        if err := p.validator.ValidateWorkItem(ctx, workItem); err != nil {
            return nil, fmt.Errorf("validation failed for row %d: %w", i+1, err)
        }

        workItems = append(workItems, workItem)
    }

    p.logger.Info("successfully parsed timesheet", "rows", len(workItems))
    return workItems, nil
}
```

**Verification Steps:**
```bash
# 1. Run security and quality checks
govulncheck ./...
go mod verify
golangci-lint run

# 2. Run comprehensive tests with testify
go test -v -race -cover ./internal/csv/...
go test -v -race -cover ./internal/services/...

# 3. Import CSV data with context
./go-invoice import --file examples/timesheet.csv --invoice INV-001

# 4. Verify imported data
./go-invoice invoice show INV-001

# 5. Test append functionality
./go-invoice import --file more-hours.csv --invoice INV-001 --append

# 6. Test validation with clear error messages
./go-invoice import --file invalid.csv --invoice INV-001

# 7. Test context cancellation on large imports
timeout 2s ./go-invoice import --file large-timesheet.csv --invoice INV-001
```

**Success Criteria:**
- ✅ CSV files parse correctly with proper validation and context support
- ✅ Work items import with accurate calculations and proper error handling
- ✅ Append mode adds to existing invoices without data corruption
- ✅ Invalid CSV data produces helpful, actionable error messages
- ✅ Import handles various CSV formats gracefully with clear feedback
- ✅ All CSV operations accept context.Context for cancellation support
- ✅ Consumer-driven interfaces used for parser abstraction
- ✅ Dependency injection eliminates global state
- ✅ Tests use testify suite with comprehensive edge-case coverage
- ✅ Context cancellation works for large file imports
- ✅ Error wrapping provides clear operation context
- ✅ No security vulnerabilities in CSV parsing dependencies
- ✅ Race condition testing passes for concurrent imports
- ✅ Final todo: Update the @plans/plan-01-status.md file with the results of the implementation

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
- `internal/render/interface.go` - Renderer interface definition
- `internal/services/calculator.go` - Invoice calculation service
- `cmd/go-invoice/cmd/generate.go` - Generate command

**Enhanced Template Engine Design:**
```go
// Renderer interface defined at point of use
type InvoiceRenderer interface {
    RenderInvoice(ctx context.Context, invoice *Invoice) (string, error)
    ValidateTemplate(ctx context.Context, templatePath string) error
}

type TemplateRenderer struct {
    logger    Logger
    validator TemplateValidator
    cache     TemplateCache
}

func NewTemplateRenderer(logger Logger, validator TemplateValidator, cache TemplateCache) *TemplateRenderer {
    return &TemplateRenderer{
        logger:    logger,
        validator: validator,
        cache:     cache,
    }
}

func (r *TemplateRenderer) RenderInvoice(ctx context.Context, invoice *Invoice) (string, error) {
    select {
    case <-ctx.Done():
        return "", ctx.Err()
    default:
    }

    tmpl, err := r.cache.GetTemplate(ctx, "default")
    if err != nil {
        return "", fmt.Errorf("failed to load template: %w", err)
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, invoice); err != nil {
        return "", fmt.Errorf("template execution failed for invoice %s: %w", invoice.ID, err)
    }

    r.logger.Info("invoice rendered successfully", "id", invoice.ID, "size", buf.Len())
    return buf.String(), nil
}
```

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
# 1. Run security and quality checks
govulncheck ./...
go mod verify
golangci-lint run

# 2. Run comprehensive tests with testify
go test -v -race -cover ./internal/render/...
go test -v -race -cover ./internal/services/...

# 3. Generate invoice HTML with context
./go-invoice generate INV-001

# 4. Open in browser
open invoices/INV-001.html

# 5. Test print preview
./go-invoice generate INV-001 --preview

# 6. Generate with custom template
./go-invoice generate INV-001 --template custom

# 7. Test context cancellation on large templates
timeout 2s ./go-invoice generate INV-001 --complex-template
```

**Success Criteria:**
- ✅ Invoice generates valid HTML output with context support
- ✅ Templates render with correct data and proper error handling
- ✅ Print layout looks professional across browsers
- ✅ Calculations display accurately with validation
- ✅ Custom templates load correctly with security validation
- ✅ All template operations accept context.Context for cancellation
- ✅ Template engine uses dependency injection (no global state)
- ✅ Error messages provide clear guidance for template issues
- ✅ Tests use testify suite with template rendering edge cases
- ✅ Context cancellation works for complex template generation
- ✅ Security scanning passes for template dependencies
- ✅ Race condition testing passes for concurrent rendering
- ✅ Final todo: Update the @plans/plan-01-status.md file with the results of the implementation

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
# 1. Run security and quality checks
govulncheck ./...
go mod verify
golangci-lint run

# 2. Run comprehensive CLI tests with testify
go test -v -race -cover ./cmd/go-invoice/...
go test -v -race -cover ./internal/cli/...

# 3. Test full workflow with context
./go-invoice invoice create --client "Acme Corp" --interactive

# 4. Import hours
./go-invoice import --file hours.csv --invoice INV-001

# 5. Generate invoice
./go-invoice generate INV-001

# 6. Test search with context
./go-invoice invoice search --client "Acme"

# 7. Test filtering
./go-invoice invoice list --status unpaid --month 2024-01

# 8. Test context cancellation across CLI commands
timeout 2s ./go-invoice invoice list --slow-query
```

**Success Criteria:**
- ✅ All commands function as documented with context support
- ✅ Interactive mode provides good UX with proper error handling
- ✅ Search and filters work accurately with clear feedback
- ✅ Error messages are helpful, actionable, and properly wrapped
- ✅ Help text is comprehensive and clear
- ✅ All CLI operations accept context.Context for cancellation
- ✅ Dependency injection used throughout CLI layer
- ✅ Command handlers implement proper error wrapping patterns
- ✅ Tests use testify suite with comprehensive CLI interaction coverage
- ✅ Context cancellation works across all CLI commands
- ✅ Interactive prompts handle context cancellation gracefully
- ✅ Security scanning passes for CLI dependencies
- ✅ Race condition testing passes for concurrent CLI usage
- ✅ Final todo: Update the @plans/plan-01-status.md file with the results of the implementation

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
// Unit test example using testify suite
func TestInvoiceCalculation(t *testing.T) {
    tests := []struct {
        name        string
        workItems   []WorkItem
        taxRate     float64
        expected    float64
        expectError bool
    }{
        {
            name: "ValidCalculationWithTax",
            workItems: []WorkItem{
                {Hours: 10, Rate: 100, Total: 1000},
                {Hours: 5, Rate: 150, Total: 750},
            },
            taxRate:  0.1,
            expected: 1925, // (1000 + 750) * 1.1
        },
        {
            name:        "EmptyWorkItems",
            workItems:   []WorkItem{},
            taxRate:     0.1,
            expected:    0,
            expectError: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            calculator := NewInvoiceCalculator()

            total, err := calculator.CalculateTotal(ctx, tt.workItems, tt.taxRate)

            if tt.expectError {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.InDelta(t, tt.expected, total, 0.01)
            }
        })
    }
}

// Integration test example using testify suite
func TestFullInvoiceWorkflow(t *testing.T) {
    ctx := context.Background()

    // Setup test dependencies with dependency injection
    storage := NewMemoryStorage()
    parser := NewCSVParser(NewValidator(), NewLogger())
    renderer := NewTemplateRenderer(NewLogger())
    service := NewInvoiceService(storage, NewLogger(), NewIDGenerator())

    // Create invoice
    invoice, err := service.CreateInvoice(ctx, CreateInvoiceRequest{
        Number: "TEST-001",
        Client: Client{Name: "Test Client"},
        Date:   time.Now(),
    })
    require.NoError(t, err)
    require.NotNil(t, invoice)

    // Import CSV
    csvData := strings.NewReader("Date,Hours,Rate,Description\n2024-01-01,8,100,Development")
    workItems, err := parser.ParseTimesheet(ctx, csvData)
    require.NoError(t, err)
    require.Len(t, workItems, 1)

    // Add work items to invoice
    err = service.AddWorkItems(ctx, invoice.ID, workItems)
    require.NoError(t, err)

    // Generate HTML
    html, err := renderer.RenderInvoice(ctx, invoice)
    require.NoError(t, err)
    require.Contains(t, html, "TEST-001")
    require.Contains(t, html, "Test Client")
}
```

**Verification Steps:**
```bash
# 1. Run comprehensive security and quality checks
govulncheck ./...
go mod verify
gitleaks detect --source . --log-opts="--all" --verbose

# 2. Run unit tests with testify and race detection
go test -v -race -cover ./...

# 3. Run tests with coverage threshold validation
go test -cover ./... | grep -E "coverage: [0-9]+" | awk '{if ($2 < 90) exit 1}'

# 4. Run integration tests
go test -v -race ./test/...

# 5. Run comprehensive linting per AGENTS.md standards
golangci-lint run
go vet ./...
gofumpt -l .
goimports -l .

# 6. Test context cancellation across the application
go test -v -run TestContext ./...

# 7. Build and test release
goreleaser build --snapshot --clean

# 8. Verify no global state or init functions
grep -r "var.*=" internal/ cmd/ | grep -v test | grep -v const
grep -r "func init()" internal/ cmd/
```

**Success Criteria:**
- ✅ Test coverage exceeds 90% using testify suite patterns
- ✅ All critical paths have comprehensive tests with edge cases
- ✅ Documentation is clear, complete, and follows AGENTS.md standards
- ✅ Examples demonstrate key workflows with context handling
- ✅ CI/CD pipeline passes all enhanced checks including security scanning
- ✅ All tests use testify assertions and table-driven patterns
- ✅ Context cancellation tested across all operations
- ✅ Error handling tested with proper wrapping verification
- ✅ No global state or init functions detected in codebase
- ✅ Race condition testing passes for all concurrent operations
- ✅ Dependency injection patterns verified in all tests
- ✅ Security vulnerabilities scan clean (govulncheck)
- ✅ Secret detection passes (gitleaks)
- ✅ All linting and formatting per AGENTS.md standards passes
- ✅ Final todo: Update the @plans/plan-01-status.md file with the results of the implementation

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

## Claude Code MCP Integration

go-invoice includes comprehensive support for Claude Code through the Model Context Protocol (MCP), enabling natural language invoice management within your development environment. The MCP integration operates alongside the existing CLI without any disruption to current workflows.

### Integration Overview

The MCP server provides a bridge between Claude Code and the go-invoice CLI, allowing developers to manage invoices through natural language while maintaining all the security and functionality of the command-line interface.

**Key Benefits:**
- **Natural Language Interface**: Manage invoices through conversational interactions
- **Seamless Integration**: Works alongside existing CLI workflows without disruption
- **Local Operation**: All processing happens locally with no external dependencies
- **Resource Access**: Direct access to invoice and client data through "@" mentions
- **Slash Commands**: Custom commands exposed as `/mcp__go_invoice__*` patterns

### Local stdio Transport Configuration

Claude Code integrates with go-invoice through local stdio transport, providing secure and efficient communication.

#### MCP Server Configuration
```json
{
  "mcpServers": {
    "go-invoice": {
      "command": "go-invoice-mcp",
      "args": ["--stdio"],
      "env": {
        "GO_INVOICE_HOME": "~/.go-invoice",
        "MCP_LOG_LEVEL": "info"
      }
    }
  }
}
```

#### Project-Scope Configuration
For project-specific invoice management, configure the MCP server at the project level:

```json
// .claude_config.json (project root)
{
  "mcp": {
    "servers": {
      "go-invoice": {
        "command": "./bin/go-invoice-mcp",
        "args": ["--stdio", "--config", "./invoice-config.json"],
        "workingDirectory": "."
      }
    }
  }
}
```

### Claude Code Integration Patterns

#### Slash Commands
Claude Code exposes go-invoice functionality through slash commands:

```bash
# Invoice management commands
/invoice
/mcp__go_invoice__list_invoices
/mcp__go_invoice__show_invoice
/import
/generate

# Configuration commands
/mcp__go_invoice__show_config
/mcp__go_invoice__validate_setup
```

#### Resource Mentions
Access invoice data directly using "@" mention patterns:

```bash
# Reference specific invoices
@invoice:INV-2025-001
@invoice:latest
@invoice:unpaid

# Reference client data
@client:acme-corp
@client:all

# Reference timesheet files
@timesheet:./hours/january-2025.csv
```

#### Example Interactions

**Creating and Managing Invoices:**
```
User: /invoice Create an invoice for Acme Corp

Claude: I'll create a new invoice for Acme Corp using the go-invoice MCP tool.

[MCP tool execution: invoice_create with client="Acme Corp"]

✅ Invoice INV-2025-003 created successfully for Acme Corp
- Invoice ID: INV-2025-003
- Client: Acme Corp
- Status: Draft
- Due Date: 2025-09-01

User: Import my timesheet data into @invoice:INV-2025-003

Claude: I'll import your timesheet data into invoice INV-2025-003.

[MCP tool execution: import_csv with invoice_id="INV-2025-003"]

✅ Imported 24 work items totaling 156 hours
- Total Amount: $15,600.00
- Date Range: 2025-08-01 to 2025-08-31
- Ready for generation
```

**Data Analysis and Reporting:**
```
User: Show me all unpaid invoices from @client:acme-corp

Claude: Let me retrieve all unpaid invoices for Acme Corp.

[MCP tool execution: list_invoices with filters]

📊 Found 3 unpaid invoices for Acme Corp:
- INV-2025-001: $8,400 (Due: 2025-08-15) - OVERDUE
- INV-2025-002: $12,200 (Due: 2025-09-01)
- INV-2025-003: $15,600 (Due: 2025-09-01)

Total Outstanding: $36,200

User: Generate HTML for all of these invoices

Claude: I'll generate HTML invoices for all 3 unpaid Acme Corp invoices.

[MCP tool execution: generate_invoice for each]

✅ Generated 3 HTML invoices:
- invoice-2025-001.html (Ready for sending)
- invoice-2025-002.html (Ready for sending)
- invoice-2025-003.html (Ready for sending)
```

### Security and Local Operation

The MCP integration prioritizes security through local-only operation and comprehensive validation:

**Local Security Measures:**
- **Stdio Transport**: All communication through secure local stdio pipes
- **Command Validation**: All CLI commands validated before execution
- **File System Sandboxing**: Restricted file access to invoice directories only
- **Input Sanitization**: All user inputs sanitized to prevent injection attacks
- **No Network Dependencies**: Completely offline operation

**Configuration Security:**
```json
{
  "security": {
    "allowedCommands": [
      "go-invoice",
      "config",
      "invoice",
      "import",
      "generate"
    ],
    "restrictedPaths": [
      "~/.go-invoice/",
      "./invoices/",
      "./timesheets/"
    ],
    "maxExecutionTime": "30s",
    "enableLogging": true
  }
}
```

### Installation and Setup

**Prerequisites:**
- go-invoice CLI installed and configured
- Claude Code with MCP support
- Go 1.24+ for building from source

**Quick Setup:**
```bash
# Build MCP server
go build -o bin/go-invoice-mcp ./cmd/go-invoice-mcp

# Install for user-scope access
cp bin/go-invoice-mcp ~/bin/

# Configure Claude Code (automatic via setup script)
./scripts/setup-claude-code-integration.sh

# Test integration
echo "Test MCP integration" | go-invoice-mcp --stdio --test
```

**Manual Configuration:**
Add to your Claude Code MCP configuration:
```json
{
  "go-invoice": {
    "command": "go-invoice-mcp",
    "args": ["--stdio"],
    "env": {
      "GO_INVOICE_HOME": "~/.go-invoice"
    }
  }
}
```

### Development Workflow Integration

The MCP integration enhances development workflows by providing contextual invoice management:

**During Development:**
```
User: I just finished the authentication module for @client:startup-xyz.
      Log 8 hours at $150/hour to their current invoice.

Claude: I'll add 8 hours of authentication development work to StartupXYZ's current invoice.

[Creates work item and adds to invoice]

✅ Added work item:
- Date: 2025-08-03
- Hours: 8.0
- Rate: $150/hour
- Description: Authentication module development
- Total: $1,200
- Added to: INV-2025-004 (StartupXYZ)
```

**Project Completion:**
```
User: The project is complete. Generate final invoices for all clients
      and show me the total earnings for this month.

Claude: I'll generate final invoices and calculate monthly earnings.

[Generates invoices and calculates totals]

✅ Generated 5 final invoices
📊 August 2025 Summary:
- Total Hours: 168
- Total Earnings: $25,200
- Clients Invoiced: 5
- Average Rate: $150/hour
```

This integration transforms invoice management from a separate administrative task into a natural part of the development workflow, accessible through the same interface used for coding and project management.

## Implementation Timeline

- **Session 0**: Foundation Alignment (Phase 0) - 30 minutes
- **Session 1**: Core Infrastructure (Phase 1) - 2-3 hours
- **Session 2**: Data Models and Storage (Phase 2) - 3-4 hours
- **Session 3**: CSV Import (Phase 3) - 2-3 hours
- **Session 4**: Template System (Phase 4) - 3-4 hours
- **Session 5**: CLI Commands (Phase 5) - 2-3 hours
- **Session 6**: Testing and Documentation (Phase 6) - 3-4 hours

Total estimated time: 15.5-21.5 hours across 7 focused sessions

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
- **Go Version**: Compatible with Go 1.24+
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

**Key improvements in this enhanced plan:**
- **AGENTS.md Compliance**: Full alignment with established engineering standards
- **Context-First Architecture**: All operations support cancellation and timeout
- **Security-First Approach**: Integrated vulnerability scanning and dependency verification
- **Excellence in Error Handling**: Comprehensive error wrapping with actionable messages
- **Testify Integration**: Comprehensive test coverage using established patterns
- **Dependency Injection**: Eliminates global state for better testability
- **Consumer-Driven Interfaces**: Minimal, focused contracts defined at point of use
- **Phased Approach**: Each phase delivers working functionality with quality gates
- **Performance First**: Efficient operations even with large datasets
- **Future-Ready**: Architecture supports PDF, email, and cloud storage additions

This implementation exemplifies established Go patterns from AGENTS.md:
- **Standard Project Layout** following go-template structure
- **Interface-Based Design** for pluggable components (accept interfaces, return concrete types)
- **Context-Aware Operations** for proper cancellation and timeout handling
- **Structured Logging** for debugging and monitoring
- **Configuration as Code** with validation and defaults
- **No Global State** - dependency injection throughout
- **No init() Functions** - explicit initialization patterns
- **Error Handling Excellence** - comprehensive wrapping and context
- **Security Integration** - govulncheck, go mod verify, gitleaks
- **Testing Standards** - testify suite with table-driven tests

go-invoice positions itself as the go-to solution for developers who value engineering excellence, simplicity, control, and professional results in their invoicing workflow.
