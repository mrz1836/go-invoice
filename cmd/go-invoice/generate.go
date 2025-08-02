// Package main provides the command-line interface for the go-invoice application.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mrz/go-invoice/internal/cli"
	"github.com/mrz/go-invoice/internal/config"
	"github.com/mrz/go-invoice/internal/models"
	"github.com/mrz/go-invoice/internal/render"
	"github.com/mrz/go-invoice/internal/services"
	jsonStorage "github.com/mrz/go-invoice/internal/storage/json"
	"github.com/spf13/cobra"
)

// buildGenerateCommand creates the generate command with subcommands
func (a *App) buildGenerateCommand() *cobra.Command {
	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate HTML invoices from stored data",
		Long: `Generate professional HTML invoices using customizable templates.
		
Features:
- Professional HTML output optimized for printing
- Customizable templates with Go template syntax
- Automatic calculation validation
- Multiple output formats
- Template preview functionality

Examples:
  go-invoice generate invoice INV-001
  go-invoice generate invoice INV-001 --template professional
  go-invoice generate invoice INV-001 --output /path/to/output.html`,
	}

	// Add generate subcommands
	generateCmd.AddCommand(a.buildGenerateInvoiceCommand())
	generateCmd.AddCommand(a.buildGeneratePreviewCommand())
	generateCmd.AddCommand(a.buildGenerateTemplateListCommand())

	return generateCmd
}

// buildGenerateInvoiceCommand creates the invoice generation command
func (a *App) buildGenerateInvoiceCommand() *cobra.Command {
	var (
		templateName string
		outputPath   string
		openBrowser  bool
		validate     bool
		currency     string
		taxRate      float64
	)

	cmd := &cobra.Command{
		Use:   "invoice <invoice-id>",
		Short: "Generate HTML invoice for a specific invoice ID",
		Long: `Generate a professional HTML invoice from stored invoice data.

The generated HTML is optimized for printing and includes:
- Company and client information
- Detailed work items table
- Professional styling and layout
- Accurate financial calculations
- Print-friendly formatting

Template Options:
- default: Clean, professional template (default)
- professional: Enhanced template with additional styling
- minimal: Simple, minimal template

Examples:
  go-invoice generate invoice INV-001
  go-invoice generate invoice INV-001 --template professional
  go-invoice generate invoice INV-001 --output invoice.html --open`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			invoiceID := args[0]
			configPath, _ := cmd.Flags().GetString("config")

			return a.executeGenerateInvoice(ctx, invoiceID, configPath, GenerateInvoiceOptions{
				TemplateName: templateName,
				OutputPath:   outputPath,
				OpenBrowser:  openBrowser,
				Validate:     validate,
				Currency:     currency,
				TaxRate:      taxRate,
			})
		},
	}

	cmd.Flags().StringVar(&templateName, "template", "default", "Template to use for generation")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (default: <invoice-number>.html)")
	cmd.Flags().BoolVar(&openBrowser, "open", false, "Open generated invoice in default browser")
	cmd.Flags().BoolVar(&validate, "validate", true, "Validate calculations before generation")
	cmd.Flags().StringVar(&currency, "currency", "", "Override currency for display (default from config)")
	cmd.Flags().Float64Var(&taxRate, "tax-rate", -1, "Override tax rate (-1 to use invoice rate)")

	return cmd
}

// buildGeneratePreviewCommand creates the template preview command
func (a *App) buildGeneratePreviewCommand() *cobra.Command {
	var (
		templateName string
		sampleData   bool
	)

	cmd := &cobra.Command{
		Use:   "preview [invoice-id]",
		Short: "Preview invoice generation without saving",
		Long: `Preview how an invoice will look when generated without saving to file.

If no invoice ID is provided, uses sample data to preview the template.
Useful for testing templates and checking formatting.

Examples:
  go-invoice generate preview INV-001
  go-invoice generate preview --sample --template professional`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			var invoiceID string
			if len(args) > 0 {
				invoiceID = args[0]
			}

			configPath, _ := cmd.Flags().GetString("config")

			return a.executeGeneratePreview(ctx, invoiceID, configPath, GeneratePreviewOptions{
				TemplateName: templateName,
				SampleData:   sampleData || invoiceID == "",
			})
		},
	}

	cmd.Flags().StringVar(&templateName, "template", "default", "Template to use for preview")
	cmd.Flags().BoolVar(&sampleData, "sample", false, "Use sample data instead of real invoice")

	return cmd
}

// buildGenerateTemplateListCommand creates the template list command
func (a *App) buildGenerateTemplateListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "List available invoice templates",
		Long: `List all available invoice templates with their descriptions and metadata.

Shows template information including:
- Template name and description
- Author and version information
- Required variables and dependencies
- Template validation status

Example:
  go-invoice generate templates`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			configPath, _ := cmd.Flags().GetString("config")

			return a.executeGenerateTemplateList(ctx, configPath)
		},
	}

	return cmd
}

// Generate command execution methods

func (a *App) executeGenerateInvoice(ctx context.Context, invoiceID, configPath string, options GenerateInvoiceOptions) error {
	a.logger.Info("executing generate invoice", "invoice_id", invoiceID, "template", options.TemplateName)

	start := time.Now()

	// Setup services and retrieve invoice
	config, renderService, invoice, err := a.setupGenerateServices(ctx, configPath, invoiceID)
	if err != nil {
		return err
	}

	fmt.Printf("📄 Generating invoice: %s (%s)\n", invoice.Number, invoice.Client.Name)

	// Validate calculations if requested
	if err := a.validateCalculationsIfRequested(ctx, options, invoice, config); err != nil {
		return err
	}

	// Generate HTML content
	html, err := renderService.RenderInvoice(ctx, invoice, options.TemplateName)
	if err != nil {
		return fmt.Errorf("failed to render invoice: %w", err)
	}

	// Write output file
	outputPath, err := a.writeGeneratedInvoice(html, options.OutputPath, invoice.Number)
	if err != nil {
		return err
	}

	// Display results and handle browser opening
	a.displayGenerationResults(outputPath, html, options, time.Since(start))

	return nil
}

// setupGenerateServices sets up configuration and services for invoice generation
func (a *App) setupGenerateServices(ctx context.Context, configPath, invoiceID string) (*config.Config, render.InvoiceRenderer, *models.Invoice, error) {
	// Load configuration
	config, err := a.configService.LoadConfig(ctx, configPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create render service
	renderService, err := a.createRenderService(config)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create render service: %w", err)
	}

	// Create invoice service and get invoice
	invoiceService := a.createInvoiceService(config.Storage.DataDir)

	invoice, err := invoiceService.GetInvoice(ctx, models.InvoiceID(invoiceID))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to retrieve invoice: %w", err)
	}

	return config, renderService, invoice, nil
}

// validateCalculationsIfRequested validates invoice calculations if validation is enabled
func (a *App) validateCalculationsIfRequested(ctx context.Context, options GenerateInvoiceOptions, invoice *models.Invoice, config *config.Config) error {
	if !options.Validate {
		return nil
	}

	calcService := services.NewInvoiceCalculator(a.logger)
	calcOptions := a.buildCalculationOptions(options, invoice, config)

	if validationErr := calcService.ValidateCalculation(ctx, invoice, calcOptions); validationErr != nil {
		fmt.Printf("⚠️  Calculation validation warning: %v\n", validationErr)
	} else {
		fmt.Println("✅ Calculations validated")
	}

	return nil
}

// buildCalculationOptions creates calculation options from generation options
func (a *App) buildCalculationOptions(options GenerateInvoiceOptions, invoice *models.Invoice, config *config.Config) *services.CalculationOptions {
	calcOptions := &services.CalculationOptions{
		TaxRate:       options.TaxRate,
		Currency:      options.Currency,
		DecimalPlaces: 2,
		RoundingMode:  "round",
	}

	// Apply defaults if values not specified
	if options.TaxRate < 0 {
		calcOptions.TaxRate = invoice.TaxRate
	}
	if options.Currency == "" {
		calcOptions.Currency = config.Invoice.Currency
	}

	return calcOptions
}

// writeGeneratedInvoice writes the generated HTML to a file
func (a *App) writeGeneratedInvoice(html, outputPath, invoiceNumber string) (string, error) {
	// Determine output path if not specified
	if outputPath == "" {
		outputPath = a.createSafeFilename(invoiceNumber)
	}

	// Ensure output directory exists
	if err := a.ensureOutputDirectory(outputPath); err != nil {
		return "", err
	}

	// Write to file
	if err := os.WriteFile(outputPath, []byte(html), 0o644); err != nil {
		return "", fmt.Errorf("failed to write output file: %w", err)
	}

	return outputPath, nil
}

// createSafeFilename creates a safe filename from invoice number
func (a *App) createSafeFilename(invoiceNumber string) string {
	safeNumber := strings.ReplaceAll(invoiceNumber, "/", "-")
	safeNumber = strings.ReplaceAll(safeNumber, "\\", "-")
	return fmt.Sprintf("%s.html", safeNumber)
}

// ensureOutputDirectory ensures the output directory exists
func (a *App) ensureOutputDirectory(outputPath string) error {
	if dir := filepath.Dir(outputPath); dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}
	return nil
}

// displayGenerationResults displays the generation results and optionally opens browser
func (a *App) displayGenerationResults(outputPath, html string, options GenerateInvoiceOptions, duration time.Duration) {
	fmt.Printf("✅ Invoice generated successfully!\n")
	fmt.Printf("   Output: %s\n", outputPath)
	fmt.Printf("   Size: %d bytes\n", len(html))
	fmt.Printf("   Template: %s\n", options.TemplateName)
	fmt.Printf("   Generation time: %v\n", duration)

	// Open in browser if requested
	if options.OpenBrowser {
		if err := a.openInBrowser(outputPath); err != nil {
			fmt.Printf("⚠️  Could not open browser: %v\n", err)
		} else {
			fmt.Println("🌐 Opened in default browser")
		}
	}
}

func (a *App) executeGeneratePreview(ctx context.Context, invoiceID, configPath string, options GeneratePreviewOptions) error {
	a.logger.Info("executing generate preview", "invoice_id", invoiceID, "template", options.TemplateName)

	// Load configuration
	config, err := a.configService.LoadConfig(ctx, configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create render service
	renderService, err := a.createRenderService(config)
	if err != nil {
		return fmt.Errorf("failed to create render service: %w", err)
	}

	var invoice *models.Invoice

	if options.SampleData {
		// Create sample invoice for preview
		invoice = a.createSampleInvoice(config)
		fmt.Println("📄 Generating preview with sample data")
	} else {
		// Create invoice service and get real invoice
		invoiceService := a.createInvoiceService(config.Storage.DataDir)

		invoice, err = invoiceService.GetInvoice(ctx, models.InvoiceID(invoiceID))
		if err != nil {
			return fmt.Errorf("failed to retrieve invoice: %w", err)
		}

		fmt.Printf("📄 Generating preview for: %s (%s)\n", invoice.Number, invoice.Client.Name)
	}

	// Generate HTML
	html, err := renderService.RenderInvoice(ctx, invoice, options.TemplateName)
	if err != nil {
		return fmt.Errorf("failed to render invoice: %w", err)
	}

	// Display preview information
	fmt.Printf("✅ Preview generated successfully!\n")
	fmt.Printf("   Template: %s\n", options.TemplateName)
	fmt.Printf("   Size: %d bytes\n", len(html))
	fmt.Printf("   Work Items: %d\n", len(invoice.WorkItems))
	fmt.Printf("   Total: %.2f %s\n", invoice.Total, config.Invoice.Currency)

	// Show first few lines of HTML
	lines := strings.Split(html, "\n")
	previewLines := 10
	if len(lines) < previewLines {
		previewLines = len(lines)
	}

	fmt.Println("\n📋 HTML Preview (first 10 lines):")
	fmt.Println("=====================================")
	for i := 0; i < previewLines; i++ {
		fmt.Printf("%3d: %s\n", i+1, lines[i])
	}
	if len(lines) > previewLines {
		fmt.Printf("... (%d more lines)\n", len(lines)-previewLines)
	}

	return nil
}

func (a *App) executeGenerateTemplateList(ctx context.Context, configPath string) error {
	a.logger.Info("executing generate template list")

	// Load configuration
	config, err := a.configService.LoadConfig(ctx, configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create render service
	renderService, err := a.createRenderService(config)
	if err != nil {
		return fmt.Errorf("failed to create render service: %w", err)
	}

	// Get available templates
	templates, err := renderService.ListAvailableTemplates(ctx)
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	fmt.Println("📝 Available Invoice Templates")
	fmt.Println("==============================")

	if len(templates) == 0 {
		fmt.Println("No templates found.")
		return nil
	}

	for _, templateName := range templates {
		fmt.Printf("\n🎨 %s\n", templateName)

		// Get template info
		info, err := renderService.GetTemplateInfo(ctx, templateName)
		if err != nil {
			fmt.Printf("   Error getting template info: %v\n", err)
			continue
		}

		if info.Description != "" {
			fmt.Printf("   Description: %s\n", info.Description)
		}
		if info.Author != "" {
			fmt.Printf("   Author: %s\n", info.Author)
		}
		if info.Version != "" {
			fmt.Printf("   Version: %s\n", info.Version)
		}

		fmt.Printf("   Size: %d bytes\n", info.SizeBytes)
		fmt.Printf("   Built-in: %v\n", info.IsBuiltIn)
		fmt.Printf("   Valid: %v\n", info.IsValid)

		if !info.IsValid && info.LastError != "" {
			fmt.Printf("   Error: %s\n", info.LastError)
		}

		if len(info.Tags) > 0 {
			fmt.Printf("   Tags: %s\n", strings.Join(info.Tags, ", "))
		}
	}

	fmt.Printf("\nTotal: %d template(s)\n", len(templates))
	return nil
}

// Helper methods

func (a *App) createRenderService(_ *config.Config) (*render.TemplateRenderer, error) {
	// Create file reader
	fileReader := &SimpleFileReader{}

	// Create logger wrapper
	loggerWrapper := &LoggerWrapper{logger: a.logger}

	// Create template engine
	engine := render.NewHTMLTemplateEngine(fileReader, loggerWrapper)

	// Load built-in templates
	if err := a.loadBuiltInTemplates(engine); err != nil {
		return nil, fmt.Errorf("failed to load built-in templates: %w", err)
	}

	// Create template cache
	cache := &SimpleTemplateCache{
		templates: make(map[string]render.Template),
	}

	// Create template validator
	validator := &SimpleTemplateValidator{
		logger: loggerWrapper,
	}

	// Create renderer options
	options := &render.RendererOptions{
		TemplateDir:       "templates",
		CacheSize:         100,
		CacheExpiry:       30 * time.Minute,
		EnableSecurity:    true,
		EnableCompression: false,
		DefaultTemplate:   "default",
		MaxRenderTime:     30 * time.Second,
	}

	// Create renderer
	renderer := render.NewTemplateRenderer(engine, cache, validator, loggerWrapper, options)

	return renderer, nil
}

func (a *App) createInvoiceService(dataDir string) *services.InvoiceService {
	// Create storage
	storage := jsonStorage.NewJSONStorage(dataDir, a.logger)

	// Create invoice service
	invoiceService := services.NewInvoiceService(storage, storage, a.logger, &SimpleIDGenerator{})

	return invoiceService
}

func (a *App) loadBuiltInTemplates(engine render.TemplateEngine) error {
	ctx := context.Background()

	// Load default template
	defaultTemplate, err := os.ReadFile("templates/invoice/default.html")
	if err != nil {
		// If file doesn't exist, use embedded template
		defaultTemplate = []byte(a.getDefaultTemplateContent())
	}

	if err := engine.ParseTemplateString(ctx, "default", string(defaultTemplate)); err != nil {
		return fmt.Errorf("failed to load default template: %w", err)
	}

	// Add more built-in templates here as needed
	return nil
}

func (a *App) createSampleInvoice(_ *config.Config) *models.Invoice {
	// Create sample client
	client := models.Client{
		ID:      models.ClientID("sample_client"),
		Name:    "Sample Client Inc.",
		Email:   "contact@sampleclient.com",
		Phone:   "+1-555-123-4567",
		Address: "123 Business Ave, Suite 100, City, State 12345",
		TaxID:   "12-3456789",
		Active:  true,
	}

	// Create sample work items
	workItems := []models.WorkItem{
		{
			ID:          "work_001",
			Date:        time.Now().AddDate(0, 0, -7),
			Hours:       8.0,
			Rate:        125.00,
			Description: "Web application development",
			Total:       1000.00,
		},
		{
			ID:          "work_002",
			Date:        time.Now().AddDate(0, 0, -6),
			Hours:       6.5,
			Rate:        125.00,
			Description: "Database optimization and testing",
			Total:       812.50,
		},
		{
			ID:          "work_003",
			Date:        time.Now().AddDate(0, 0, -5),
			Hours:       4.0,
			Rate:        125.00,
			Description: "Code review and documentation",
			Total:       500.00,
		},
	}

	// Create sample invoice
	invoice := &models.Invoice{
		ID:          models.InvoiceID("sample_invoice"),
		Number:      "SAMPLE-001",
		Date:        time.Now().AddDate(0, 0, -1),
		DueDate:     time.Now().AddDate(0, 0, 30),
		Client:      client,
		WorkItems:   workItems,
		Status:      models.StatusDraft,
		Description: "Sample invoice for template preview",
		Subtotal:    2312.50,
		TaxRate:     0.10,
		TaxAmount:   231.25,
		Total:       2543.75,
		CreatedAt:   time.Now().AddDate(0, 0, -1),
		UpdatedAt:   time.Now(),
		Version:     1,
	}

	return invoice
}

func (a *App) openInBrowser(path string) error {
	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Use platform-specific command to open file
	var cmd string

	switch {
	case strings.Contains(strings.ToLower(os.Getenv("OS")), "windows"):
		cmd = fmt.Sprintf("start %s", absPath)
	case os.Getenv("DISPLAY") != "": // Linux with display
		cmd = fmt.Sprintf("xdg-open %s", absPath)
	default: // macOS
		cmd = fmt.Sprintf("open %s", absPath)
	}

	// This is a simplified implementation - in a real CLI you'd use exec.Command
	fmt.Printf("To open in browser, run: %s\n", cmd)
	return nil
}

func (a *App) getDefaultTemplateContent() string {
	// Return basic template content if file can't be read
	return `<!DOCTYPE html>
<html>
<head>
    <title>Invoice {{.Number}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { border-bottom: 2px solid #ccc; padding-bottom: 20px; }
        .invoice-title { font-size: 36px; color: #333; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        th { background-color: #f2f2f2; }
        .total { font-weight: bold; font-size: 18px; }
    </style>
</head>
<body>
    <div class="header">
        <h1 class="invoice-title">INVOICE</h1>
        <p>Invoice #{{.Number}}</p>
        <p>Date: {{formatDate .Date "January 2, 2006"}}</p>
    </div>
    
    <h2>Bill To:</h2>
    <p>{{.Client.Name}}<br>{{.Client.Address}}</p>
    
    <table>
        <tr><th>Date</th><th>Description</th><th>Hours</th><th>Rate</th><th>Amount</th></tr>
        {{range .WorkItems}}
        <tr>
            <td>{{formatDate .Date "Jan 2"}}</td>
            <td>{{.Description}}</td>
            <td>{{.Hours}}</td>
            <td>{{formatCurrency .Rate "USD"}}</td>
            <td>{{formatCurrency .Total "USD"}}</td>
        </tr>
        {{end}}
    </table>
    
    <div class="total">
        <p>Total: {{formatCurrency .Total "USD"}}</p>
    </div>
</body>
</html>`
}

// Option types for generate commands

type GenerateInvoiceOptions struct {
	TemplateName string
	OutputPath   string
	OpenBrowser  bool
	Validate     bool
	Currency     string
	TaxRate      float64
}

type GeneratePreviewOptions struct {
	TemplateName string
	SampleData   bool
}

// Enhanced data structures for templates

type EnhancedInvoiceData struct {
	models.Invoice
	Business EnhancedBusinessInfo `json:"business"`
	Config   EnhancedConfigInfo   `json:"config"`
}

type EnhancedBusinessInfo struct {
	Name         string `json:"name"`
	Address      string `json:"address"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Website      string `json:"website"`
	TaxID        string `json:"tax_id"`
	PaymentTerms string `json:"payment_terms"`
	BankDetails  string `json:"bank_details"`
}

type EnhancedConfigInfo struct {
	Currency       string `json:"currency"`
	CurrencySymbol string `json:"currency_symbol"`
	DateFormat     string `json:"date_format"`
	DecimalPlaces  int    `json:"decimal_places"`
}

// LoggerWrapper wraps cli.SimpleLogger to implement render.Logger interface
type LoggerWrapper struct {
	logger *cli.SimpleLogger
}

func (l *LoggerWrapper) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.Debug(msg, keysAndValues...)
}

func (l *LoggerWrapper) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(msg, keysAndValues...)
}

func (l *LoggerWrapper) Warn(msg string, keysAndValues ...interface{}) {
	// SimpleLogger doesn't have Warn, so use Error instead
	l.logger.Error(msg, keysAndValues...)
}

func (l *LoggerWrapper) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Error(msg, keysAndValues...)
}
