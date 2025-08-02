package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mrz/go-invoice/internal/csv"
	"github.com/mrz/go-invoice/internal/models"
	"github.com/mrz/go-invoice/internal/services"
	jsonStorage "github.com/mrz/go-invoice/internal/storage/json"
	"github.com/spf13/cobra"
)

// Import command errors
var (
	ErrClientIDRequired  = fmt.Errorf("client ID is required (use --client flag)")
	ErrInvoiceIDRequired = fmt.Errorf("invoice ID is required (use --invoice flag)")
)

// buildImportCommand creates the import command with subcommands
func (a *App) buildImportCommand() *cobra.Command {
	importCmd := &cobra.Command{
		Use:   "import",
		Short: "Import timesheet data from CSV files",
		Long: `Import work hours from CSV files into invoices.

Supports multiple CSV formats including:
- Standard CSV (RFC 4180)
- Excel CSV exports
- Google Sheets CSV exports
- Tab-separated values (TSV)
- Custom delimiter formats

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
		dryRun        bool
		interactive   bool
		format        string
	)

	cmd := &cobra.Command{
		Use:   "create <csv-file>",
		Short: "Import CSV data into a new invoice",
		Long: `Import timesheet data from a CSV file and create a new invoice.

The CSV file should contain columns for:
- date (work date)
- hours (hours worked) 
- rate (hourly rate)
- description (work description)

Example:
  go-invoice import create timesheet.csv --client CLIENT_001`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			csvFile := args[0]
			configPath, _ := cmd.Flags().GetString("config")

			if clientID == "" {
				return ErrClientIDRequired
			}

			return a.executeImportCreate(ctx, csvFile, configPath, ImportCreateOptions{
				ClientID:      clientID,
				InvoiceNumber: invoiceNumber,
				Description:   description,
				DryRun:        dryRun,
				Interactive:   interactive,
				Format:        format,
			})
		},
	}

	cmd.Flags().StringVar(&clientID, "client", "", "Client ID for the new invoice (required)")
	cmd.Flags().StringVar(&invoiceNumber, "number", "", "Invoice number (auto-generated if not provided)")
	cmd.Flags().StringVar(&description, "description", "", "Invoice description")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate only, don't create invoice")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Interactive mode for resolving ambiguous data")
	cmd.Flags().StringVar(&format, "format", "", "Force specific CSV format (standard, excel, tab)")

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
		Use:   "append <csv-file>",
		Short: "Import CSV data and append to existing invoice",
		Long: `Import timesheet data from a CSV file and append work items to an existing invoice.

The invoice must be in draft status to accept additional work items.

Example:
  go-invoice import append timesheet.csv --invoice INV-001`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			csvFile := args[0]
			configPath, _ := cmd.Flags().GetString("config")

			if invoiceID == "" {
				return ErrInvoiceIDRequired
			}

			return a.executeImportAppend(ctx, csvFile, configPath, ImportAppendOptions{
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
	cmd.Flags().StringVar(&format, "format", "", "Force specific CSV format (standard, excel, tab)")

	return cmd
}

// buildImportValidateCommand creates the validation command
func (a *App) buildImportValidateCommand() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "validate <csv-file>",
		Short: "Validate CSV file format and data",
		Long: `Validate a CSV file for import compatibility without actually importing data.

Checks for:
- Valid CSV format and structure
- Required columns (date, hours, rate, description)
- Data type validation
- Business rule compliance
- Potential issues and warnings

Example:
  go-invoice import validate timesheet.csv`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			csvFile := args[0]
			configPath, _ := cmd.Flags().GetString("config")

			return a.executeImportValidate(ctx, csvFile, configPath, ImportValidateOptions{
				Format: format,
			})
		},
	}

	cmd.Flags().StringVar(&format, "format", "", "Force specific CSV format (standard, excel, tab)")

	return cmd
}

// Import command execution methods

func (a *App) executeImportCreate(ctx context.Context, csvFile, configPath string, options ImportCreateOptions) error {
	a.logger.Info("executing import create", "file", csvFile, "client", options.ClientID)

	// Load configuration
	config, err := a.configService.LoadConfig(ctx, configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create import service
	importService, err := a.createImportService(config.Storage.DataDir)
	if err != nil {
		return fmt.Errorf("failed to create import service: %w", err)
	}

	// Open CSV file
	file, err := os.Open(csvFile)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// Prepare import request
	req := services.ImportToNewInvoiceRequest{
		ClientID:     models.ClientID(options.ClientID),
		ParseOptions: a.createParseOptions(options.Format),
		DryRun:       options.DryRun,
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

func (a *App) executeImportAppend(ctx context.Context, csvFile, configPath string, options ImportAppendOptions) error {
	a.logger.Info("executing import append", "file", csvFile, "invoice", options.InvoiceID)

	// Load configuration
	config, err := a.configService.LoadConfig(ctx, configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create import service
	importService, err := a.createImportService(config.Storage.DataDir)
	if err != nil {
		return fmt.Errorf("failed to create import service: %w", err)
	}

	// Open CSV file
	file, err := os.Open(csvFile)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// Prepare import request
	req := services.AppendToInvoiceRequest{
		InvoiceID:    options.InvoiceID,
		ParseOptions: a.createParseOptions(options.Format),
		DryRun:       options.DryRun,
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

func (a *App) executeImportValidate(ctx context.Context, csvFile, configPath string, options ImportValidateOptions) error {
	a.logger.Info("executing import validation", "file", csvFile)

	// Load configuration
	config, err := a.configService.LoadConfig(ctx, configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create import service
	importService, err := a.createImportService(config.Storage.DataDir)
	if err != nil {
		return fmt.Errorf("failed to create import service: %w", err)
	}

	// Open CSV file
	file, err := os.Open(csvFile)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// Prepare validation request
	req := csv.ValidateImportRequest{
		Options: a.createParseOptions(options.Format),
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

func (a *App) createImportService(dataDir string) (*services.ImportService, error) {
	// Create storage
	storage := jsonStorage.NewJSONStorage(dataDir, a.logger)

	// Create services with dependency injection
	invoiceService := services.NewInvoiceService(storage, storage, a.logger, &SimpleIDGenerator{})
	clientService := services.NewClientService(storage, storage, a.logger, &SimpleIDGenerator{})

	// Create CSV components
	validator := csv.NewWorkItemValidator(a.logger)
	parser := csv.NewCSVParser(validator, a.logger, &SimpleIDGenerator{})

	// Create import service
	importService := services.NewImportService(parser, invoiceService, clientService, validator, a.logger, &SimpleIDGenerator{})

	return importService, nil
}

func (a *App) createParseOptions(format string) csv.ParseOptions {
	options := csv.ParseOptions{
		ContinueOnError: false,
		SkipEmptyRows:   true,
	}

	if format != "" {
		options.Format = format
	} else {
		options.Format = "standard" // Default format
	}

	return options
}

func (a *App) displayImportResult(result *csv.ImportResult, isDryRun bool) {
	if isDryRun {
		fmt.Println("🔍 Dry Run Results")
		fmt.Println("==================")
	} else {
		fmt.Println("✅ Import Results")
		fmt.Println("=================")
	}

	fmt.Printf("Work Items: %d\n", result.WorkItemsAdded)
	fmt.Printf("Total Amount: $%.2f\n", result.TotalAmount)

	if result.InvoiceID != "" {
		fmt.Printf("Invoice ID: %s\n", result.InvoiceID)
	}

	// Display parsing statistics
	if result.ParseResult != nil {
		fmt.Printf("\nParsing Details:\n")
		fmt.Printf("  Total Rows: %d\n", result.ParseResult.TotalRows)
		fmt.Printf("  Success Rows: %d\n", result.ParseResult.SuccessRows)
		if result.ParseResult.ErrorRows > 0 {
			fmt.Printf("  Error Rows: %d\n", result.ParseResult.ErrorRows)
		}
	}

	// Display warnings
	if len(result.Warnings) > 0 {
		fmt.Printf("\n⚠️  Warnings:\n")
		for _, warning := range result.Warnings {
			fmt.Printf("  - %s\n", warning.Message)
		}
	}

	// Display errors
	if result.ParseResult != nil && len(result.ParseResult.Errors) > 0 {
		fmt.Printf("\n❌ Errors:\n")
		for _, parseError := range result.ParseResult.Errors {
			fmt.Printf("  Line %d: %s\n", parseError.Line, parseError.Message)
		}
	}

	fmt.Println()
}

func (a *App) displayValidationResult(result *csv.ValidationResult) {
	if result.Valid {
		fmt.Println("✅ Validation Passed")
	} else {
		fmt.Println("❌ Validation Failed")
	}
	fmt.Println("====================")

	if result.ParseResult != nil {
		fmt.Printf("Work Items: %d\n", len(result.ParseResult.WorkItems))
		fmt.Printf("Total Rows: %d\n", result.ParseResult.TotalRows)
		fmt.Printf("Success Rows: %d\n", result.ParseResult.SuccessRows)
		if result.ParseResult.ErrorRows > 0 {
			fmt.Printf("Error Rows: %d\n", result.ParseResult.ErrorRows)
		}
	}

	fmt.Printf("Estimated Total: $%.2f\n", result.EstimatedTotal)

	// Display warnings
	if len(result.Warnings) > 0 {
		fmt.Printf("\n⚠️  Warnings:\n")

		for _, warning := range result.Warnings {
			fmt.Printf("  - %s\n", warning.Message)
		}
	}

	// Display suggestions
	if len(result.Suggestions) > 0 {
		fmt.Printf("\n💡 Suggestions:\n")
		for _, suggestion := range result.Suggestions {
			fmt.Printf("  - %s\n", suggestion)
		}
	}

	// Display errors
	if result.ParseResult != nil && len(result.ParseResult.Errors) > 0 {
		fmt.Printf("\n❌ Errors:\n")
		for _, parseError := range result.ParseResult.Errors {
			fmt.Printf("  Line %d: %s\n", parseError.Line, parseError.Message)
		}
	}

	fmt.Println()
}

// Option types for import commands

type ImportCreateOptions struct {
	ClientID      string
	InvoiceNumber string
	Description   string
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

func (g *SimpleIDGenerator) GenerateInvoiceID(ctx context.Context) (models.InvoiceID, error) {
	return models.InvoiceID(fmt.Sprintf("inv_%d", time.Now().Unix())), nil
}

func (g *SimpleIDGenerator) GenerateClientID(ctx context.Context) (models.ClientID, error) {
	return models.ClientID(fmt.Sprintf("client_%d", time.Now().Unix())), nil
}

func (g *SimpleIDGenerator) GenerateWorkItemID(ctx context.Context) (string, error) {
	return fmt.Sprintf("work_%d", time.Now().Unix()), nil
}

// GenerateID implements csv.IDGenerator interface
func (g *SimpleIDGenerator) GenerateID() string {
	return fmt.Sprintf("work_%d", time.Now().Unix())
}
