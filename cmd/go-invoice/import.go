package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mrz1836/go-invoice/internal/csv"
	"github.com/mrz1836/go-invoice/internal/models"
	"github.com/mrz1836/go-invoice/internal/services"
	jsonStorage "github.com/mrz1836/go-invoice/internal/storage/json"
	"github.com/spf13/cobra"
)

// Import command errors
var (
	ErrClientIDRequired  = fmt.Errorf("client ID is required (use --client flag)")
	ErrInvoiceIDRequired = fmt.Errorf("invoice ID is required (use --invoice flag)")
)

// detectFileFormat detects the format based on file extension
func detectFileFormat(filename, specifiedFormat string) string {
	// If format is explicitly specified and not "auto", use it
	if specifiedFormat != "" && specifiedFormat != "auto" {
		return specifiedFormat
	}

	// Auto-detect based on file extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return "json"
	case ".csv":
		return "csv"
	case ".tsv", ".tab":
		return "tsv"
	case ".xlsx", ".xls":
		return "excel"
	default:
		// Default to CSV if unknown
		return "csv"
	}
}

// buildImportCommand creates the import command with subcommands
func (a *App) buildImportCommand() *cobra.Command {
	importCmd := &cobra.Command{
		Use:   "import",
		Short: "Import timesheet data from CSV or JSON files",
		Long: `Import work hours from CSV or JSON files into invoices.

Supports multiple formats:

CSV Formats:
- Standard CSV (RFC 4180)
- Excel CSV exports
- Google Sheets CSV exports
- Tab-separated values (TSV)
- Custom delimiter formats

JSON Formats:
- Simple array format: [{"date": "2025-08-01", "hours": 8, "rate": 150, "description": "Work"}]
- Structured format with metadata: {"metadata": {...}, "work_items": [...]}

Can create new invoices or append to existing ones.`,
	}

	// Add import subcommands
	importCmd.AddCommand(a.buildImportCreateCommand())
	importCmd.AddCommand(a.buildImportAppendCommand())
	importCmd.AddCommand(a.buildImportValidateCommand())

	return importCmd
}

// buildImportCreateCommand creates the import command for new invoices
func (a *App) buildImportCreateCommand() *cobra.Command {
	var (
		clientID      string
		invoiceNumber string
		description   string
		invoiceDate   string
		dueDate       string
		dryRun        bool
		interactive   bool
		format        string
	)

	cmd := &cobra.Command{
		Use:   "create <file>",
		Short: "Import data into a new invoice",
		Long: `Import timesheet data from a CSV or JSON file and create a new invoice.

For CSV files, columns should include:
- date (work date)
- hours (hours worked)
- rate (hourly rate)
- description (work description)

For JSON files, use either format:
- Array: [{"date": "2025-08-01", "hours": 8, "rate": 150, "description": "Work"}]
- Structured: {"metadata": {...}, "work_items": [...]}

Format is auto-detected from file extension, or use --format flag.

Examples:
  go-invoice import create timesheet.csv --client CLIENT_001
  go-invoice import create timesheet.json --client CLIENT_001
  go-invoice import create data.txt --format json --client CLIENT_001`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			dataFile := args[0]
			configPath, _ := cmd.Flags().GetString("config")

			if clientID == "" {
				return ErrClientIDRequired
			}

			return a.executeImportCreate(ctx, dataFile, configPath, ImportCreateOptions{
				ClientID:      clientID,
				InvoiceNumber: invoiceNumber,
				Description:   description,
				InvoiceDate:   invoiceDate,
				DueDate:       dueDate,
				DryRun:        dryRun,
				Interactive:   interactive,
				Format:        format,
			})
		},
	}

	cmd.Flags().StringVar(&clientID, "client", "", "Client ID for the new invoice (required)")
	cmd.Flags().StringVar(&invoiceNumber, "number", "", "Invoice number (auto-generated if not provided)")
	cmd.Flags().StringVar(&description, "description", "", "Invoice description")
	cmd.Flags().StringVar(&invoiceDate, "date", "", "Invoice date (YYYY-MM-DD, default: today)")
	cmd.Flags().StringVar(&dueDate, "due-date", "", "Due date (YYYY-MM-DD, default: 30 days from invoice date)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate only, don't create invoice")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Interactive mode for resolving ambiguous data")
	cmd.Flags().StringVar(&format, "format", "auto", "Import format (auto, csv, json, excel, tsv)")

	return cmd
}

// buildImportAppendCommand creates the import command for existing invoices
func (a *App) buildImportAppendCommand() *cobra.Command {
	var (
		invoiceID   string
		dryRun      bool
		interactive bool
		format      string
	)

	cmd := &cobra.Command{
		Use:   "append <file>",
		Short: "Import data and append to existing invoice",
		Long: `Import timesheet data from a CSV or JSON file and append work items to an existing invoice.

The invoice must be in draft status to accept additional work items.

Examples:
  go-invoice import append timesheet.csv --invoice INV-001
  go-invoice import append timesheet.json --invoice INV-001`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			dataFile := args[0]
			configPath, _ := cmd.Flags().GetString("config")

			if invoiceID == "" {
				return ErrInvoiceIDRequired
			}

			return a.executeImportAppend(ctx, dataFile, configPath, ImportAppendOptions{
				InvoiceID:   invoiceID,
				DryRun:      dryRun,
				Interactive: interactive,
				Format:      format,
			})
		},
	}

	cmd.Flags().StringVar(&invoiceID, "invoice", "", "Invoice ID to append to (required)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate only, don't append to invoice")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Interactive mode for resolving ambiguous data")
	cmd.Flags().StringVar(&format, "format", "auto", "Import format (auto, csv, json, excel, tsv)")

	return cmd
}

// buildImportValidateCommand creates the validation command
func (a *App) buildImportValidateCommand() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "validate <file>",
		Short: "Validate file format and data",
		Long: `Validate a CSV or JSON file for import compatibility without actually importing data.

Checks for:
- Valid file format and structure (CSV or JSON)
- Required fields (date, hours, rate, description)
- Data type validation
- Business rule compliance
- Potential issues and warnings

Examples:
  go-invoice import validate timesheet.csv
  go-invoice import validate timesheet.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			dataFile := args[0]
			configPath, _ := cmd.Flags().GetString("config")

			return a.executeImportValidate(ctx, dataFile, configPath, ImportValidateOptions{
				Format: format,
			})
		},
	}

	cmd.Flags().StringVar(&format, "format", "auto", "Import format (auto, csv, json, excel, tsv)")

	return cmd
}

// Import command execution methods

func (a *App) executeImportCreate(ctx context.Context, dataFile, configPath string, options ImportCreateOptions) error {
	// Detect file format
	fileFormat := detectFileFormat(dataFile, options.Format)
	a.logger.Info("executing import create", "file", dataFile, "client", options.ClientID, "format", fileFormat)

	// Load configuration
	config, err := a.configService.LoadConfig(ctx, configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create import service
	importService := a.createImportService(config.Storage.DataDir)

	// Open data file
	file, err := os.Open(dataFile) // #nosec G304 -- User-provided file path is expected in CLI
	if err != nil {
		return fmt.Errorf("failed to open data file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			a.logger.Error("failed to close file", "error", closeErr)
		}
	}()

	// Parse dates or use defaults
	invoiceDate := time.Now()
	if options.InvoiceDate != "" {
		invoiceDate, err = time.Parse("2006-01-02", options.InvoiceDate)
		if err != nil {
			return fmt.Errorf("invalid invoice date format (use YYYY-MM-DD): %w", err)
		}
	}

	dueDate := invoiceDate.AddDate(0, 0, 30) // Default: 30 days from invoice date
	if options.DueDate != "" {
		dueDate, err = time.Parse("2006-01-02", options.DueDate)
		if err != nil {
			return fmt.Errorf("invalid due date format (use YYYY-MM-DD): %w", err)
		}
	}

	// Prepare import request
	req := services.ImportToNewInvoiceRequest{
		ClientID:     models.ClientID(options.ClientID),
		InvoiceDate:  invoiceDate,
		DueDate:      dueDate,
		ParseOptions: a.createParseOptions(fileFormat),
		DryRun:       options.DryRun,
		Format:       fileFormat,
	}

	if options.InvoiceNumber != "" {
		req.InvoiceNumber = options.InvoiceNumber
	}

	if options.Description != "" {
		req.Description = options.Description
	}

	// Execute import
	result, err := importService.ImportToNewInvoice(ctx, file, req)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	// Display results
	a.displayImportResult(result, options.DryRun)

	return nil
}

func (a *App) executeImportAppend(ctx context.Context, dataFile, configPath string, options ImportAppendOptions) error {
	// Detect file format
	fileFormat := detectFileFormat(dataFile, options.Format)
	a.logger.Info("executing import append", "file", dataFile, "invoice", options.InvoiceID, "format", fileFormat)

	// Load configuration
	config, err := a.configService.LoadConfig(ctx, configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create import service
	importService := a.createImportService(config.Storage.DataDir)

	// Open data file
	file, err := os.Open(dataFile) // #nosec G304 -- User-provided file path is expected in CLI
	if err != nil {
		return fmt.Errorf("failed to open data file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			a.logger.Error("failed to close file", "error", closeErr)
		}
	}()

	// Get invoice by ID or number
	invoiceService := a.createInvoiceService(config.Storage.DataDir)

	// Try to get invoice by ID first, then by number
	invoice, err := invoiceService.GetInvoice(ctx, models.InvoiceID(options.InvoiceID))
	if err != nil {
		// If not found by ID, try by number
		invoice, err = invoiceService.GetInvoiceByNumber(ctx, options.InvoiceID)
		if err != nil {
			return fmt.Errorf("failed to find invoice '%s': %w", options.InvoiceID, err)
		}
	}

	// Prepare import request using the resolved invoice ID
	req := services.AppendToInvoiceRequest{
		InvoiceID:    string(invoice.ID),
		ParseOptions: a.createParseOptions(fileFormat),
		DryRun:       options.DryRun,
		Format:       fileFormat,
	}

	// Execute import
	result, err := importService.AppendToInvoice(ctx, file, req)
	if err != nil {
		return fmt.Errorf("import append failed: %w", err)
	}

	// Display results
	a.displayImportResult(result, options.DryRun)

	return nil
}

func (a *App) executeImportValidate(ctx context.Context, dataFile, configPath string, options ImportValidateOptions) error {
	// Detect file format
	fileFormat := detectFileFormat(dataFile, options.Format)
	a.logger.Info("executing import validation", "file", dataFile, "format", fileFormat)

	// Load configuration
	config, err := a.configService.LoadConfig(ctx, configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create import service
	importService := a.createImportService(config.Storage.DataDir)

	// Open data file
	file, err := os.Open(dataFile) // #nosec G304 -- User-provided file path is expected in CLI
	if err != nil {
		return fmt.Errorf("failed to open data file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			a.logger.Error("failed to close file", "error", closeErr)
		}
	}()

	// Prepare validation request
	req := csv.ValidateImportRequest{
		Options: a.createParseOptions(fileFormat),
	}

	// Execute validation
	result, err := importService.ValidateImport(ctx, file, req)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Display validation results
	a.displayValidationResult(result)

	return nil
}

// Helper methods

func (a *App) createImportService(dataDir string) *services.ImportService {
	// Create storage
	storage := jsonStorage.NewJSONStorage(dataDir, a.logger)

	// Create services with dependency injection
	invoiceService := services.NewInvoiceService(storage, storage, a.logger, &SimpleIDGenerator{})
	clientService := services.NewClientService(storage, storage, a.logger, &SimpleIDGenerator{})

	// Create CSV components (validator is shared between CSV and JSON parsers)
	validator := csv.NewWorkItemValidator(a.logger)
	csvParser := csv.NewCSVParser(validator, a.logger, &SimpleIDGenerator{})

	// Create import service (JSON parser will be created internally)
	importService := services.NewImportService(csvParser, invoiceService, clientService, validator, a.logger, &SimpleIDGenerator{})

	return importService
}

func (a *App) createParseOptions(format string) csv.ParseOptions {
	options := csv.ParseOptions{
		ContinueOnError: false,
		SkipEmptyRows:   true,
		Format:          format,
	}

	// Set default format if not specified
	if options.Format == "" {
		options.Format = "standard"
	}

	return options
}

func (a *App) displayImportResult(result *csv.ImportResult, isDryRun bool) {
	if isDryRun {
		a.logger.Println("🔍 Dry Run Results")
		a.logger.Println("==================")
	} else {
		a.logger.Println("✅ Import Results")
		a.logger.Println("=================")
	}

	a.logger.Printf("Work Items: %d\n", result.WorkItemsAdded)
	a.logger.Printf("Total Amount: $%.2f\n", result.TotalAmount)

	if result.InvoiceID != "" {
		a.logger.Printf("Invoice ID: %s\n", result.InvoiceID)
	}

	// Display parsing statistics
	if result.ParseResult != nil {
		a.logger.Printf("\nParsing Details:\n")
		a.logger.Printf("  Total Rows: %d\n", result.ParseResult.TotalRows)
		a.logger.Printf("  Success Rows: %d\n", result.ParseResult.SuccessRows)
		if result.ParseResult.ErrorRows > 0 {
			a.logger.Printf("  Error Rows: %d\n", result.ParseResult.ErrorRows)
		}
	}

	// Display warnings
	if len(result.Warnings) > 0 {
		a.logger.Printf("\n⚠️  Warnings:\n")
		for _, warning := range result.Warnings {
			a.logger.Printf("  - %s\n", warning.Message)
		}
	}

	// Display errors
	if result.ParseResult != nil && len(result.ParseResult.Errors) > 0 {
		a.logger.Printf("\n❌ Errors:\n")
		for _, parseError := range result.ParseResult.Errors {
			a.logger.Printf("  Line %d: %s\n", parseError.Line, parseError.Message)
		}
	}

	a.logger.Println("")
}

func (a *App) displayValidationResult(result *csv.ValidationResult) {
	if result.Valid {
		a.logger.Println("✅ Validation Passed")
	} else {
		a.logger.Println("❌ Validation Failed")
	}
	a.logger.Println("====================")

	if result.ParseResult != nil {
		a.logger.Printf("Work Items: %d\n", len(result.ParseResult.WorkItems))
		a.logger.Printf("Total Rows: %d\n", result.ParseResult.TotalRows)
		a.logger.Printf("Success Rows: %d\n", result.ParseResult.SuccessRows)
		if result.ParseResult.ErrorRows > 0 {
			a.logger.Printf("Error Rows: %d\n", result.ParseResult.ErrorRows)
		}
	}

	a.logger.Printf("Estimated Total: $%.2f\n", result.EstimatedTotal)

	// Display warnings
	if len(result.Warnings) > 0 {
		a.logger.Printf("\n⚠️  Warnings:\n")

		for _, warning := range result.Warnings {
			a.logger.Printf("  - %s\n", warning.Message)
		}
	}

	// Display suggestions
	if len(result.Suggestions) > 0 {
		a.logger.Printf("\n💡 Suggestions:\n")
		for _, suggestion := range result.Suggestions {
			a.logger.Printf("  - %s\n", suggestion)
		}
	}

	// Display errors
	if result.ParseResult != nil && len(result.ParseResult.Errors) > 0 {
		a.logger.Printf("\n❌ Errors:\n")
		for _, parseError := range result.ParseResult.Errors {
			a.logger.Printf("  Line %d: %s\n", parseError.Line, parseError.Message)
		}
	}

	a.logger.Println("")
}

// Option types for import commands

type ImportCreateOptions struct {
	ClientID      string
	InvoiceNumber string
	Description   string
	InvoiceDate   string
	DueDate       string
	DryRun        bool
	Interactive   bool
	Format        string
}

type ImportAppendOptions struct {
	InvoiceID   string
	DryRun      bool
	Interactive bool
	Format      string
}

type ImportValidateOptions struct {
	Format string
}

// SimpleIDGenerator provides basic ID generation for the import service
type SimpleIDGenerator struct{}

func (g *SimpleIDGenerator) GenerateInvoiceID(_ context.Context) (models.InvoiceID, error) {
	return models.InvoiceID(fmt.Sprintf("inv_%d", time.Now().Unix())), nil
}

func (g *SimpleIDGenerator) GenerateClientID(_ context.Context) (models.ClientID, error) {
	return models.ClientID(fmt.Sprintf("client_%d", time.Now().Unix())), nil
}

func (g *SimpleIDGenerator) GenerateWorkItemID(_ context.Context) (string, error) {
	return fmt.Sprintf("work_%d", time.Now().Unix()), nil
}

// GenerateID implements csv.IDGenerator interface
func (g *SimpleIDGenerator) GenerateID() string {
	return fmt.Sprintf("work_%d", time.Now().Unix())
}
