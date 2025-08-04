package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/mrz/go-invoice/internal/cli"
	"github.com/mrz/go-invoice/internal/config"
	"github.com/mrz/go-invoice/internal/storage"
	jsonStorage "github.com/mrz/go-invoice/internal/storage/json"
	"github.com/spf13/cobra"
)

// Version information set by build process
var (
	Version = "dev"     //nolint:gochecknoglobals // Build-time version information
	Commit  = "unknown" //nolint:gochecknoglobals // Build-time commit information
	Date    = "unknown" //nolint:gochecknoglobals // Build-time date information
)

// defaultInvoiceTemplate contains the default invoice HTML template
const defaultInvoiceTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Invoice {{.Invoice.Number}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
            padding: 40px 20px;
        }
        .header {
            display: flex;
            justify-content: space-between;
            margin-bottom: 40px;
            padding-bottom: 20px;
            border-bottom: 2px solid #f0f0f0;
        }
        .invoice-title {
            font-size: 32px;
            font-weight: 300;
            color: #2c3e50;
            margin: 0;
        }
        .invoice-number {
            text-align: right;
            color: #7f8c8d;
        }
        .invoice-number strong {
            color: #2c3e50;
        }
        .parties {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 40px;
            margin-bottom: 40px;
        }
        .party h3 {
            color: #2c3e50;
            margin-bottom: 10px;
        }
        .party p {
            margin: 5px 0;
            color: #555;
        }
        .items-table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 40px;
        }
        .items-table th {
            text-align: left;
            padding: 12px;
            border-bottom: 2px solid #e9ecef;
            color: #495057;
            font-weight: 600;
        }
        .items-table td {
            padding: 12px;
            border-bottom: 1px solid #e9ecef;
        }
        .items-table .amount {
            text-align: right;
        }
        .totals {
            margin-left: auto;
            width: 300px;
        }
        .totals-row {
            display: flex;
            justify-content: space-between;
            padding: 8px 0;
        }
        .totals-row.total {
            font-weight: bold;
            font-size: 1.2em;
            border-top: 2px solid #e9ecef;
            padding-top: 16px;
        }
        .footer {
            margin-top: 60px;
            padding-top: 30px;
            border-top: 1px solid #e9ecef;
            text-align: center;
            color: #6c757d;
        }
        @media print {
            body {
                padding: 0;
            }
            .header {
                page-break-after: avoid;
            }
        }
    </style>
</head>
<body>
    <div class="header">
        <h1 class="invoice-title">Invoice</h1>
        <div class="invoice-number">
            <strong>{{.Invoice.Number}}</strong><br>
            {{.Invoice.Date}}
        </div>
    </div>

    <div class="parties">
        <div class="party">
            <h3>From</h3>
            <p><strong>{{.Business.Name}}</strong></p>
            {{if .Business.Address}}<p>{{.Business.Address}}</p>{{end}}
            {{if .Business.Email}}<p>{{.Business.Email}}</p>{{end}}
            {{if .Business.Phone}}<p>{{.Business.Phone}}</p>{{end}}
        </div>
        <div class="party">
            <h3>To</h3>
            <p><strong>{{.Client.Name}}</strong></p>
            {{if .Client.Address}}<p>{{.Client.Address}}</p>{{end}}
            {{if .Client.Email}}<p>{{.Client.Email}}</p>{{end}}
            {{if .Client.Phone}}<p>{{.Client.Phone}}</p>{{end}}
        </div>
    </div>

    <table class="items-table">
        <thead>
            <tr>
                <th>Date</th>
                <th>Description</th>
                <th>Hours</th>
                <th>Rate</th>
                <th class="amount">Amount</th>
            </tr>
        </thead>
        <tbody>
            {{range .Invoice.Items}}
            <tr>
                <td>{{.Date}}</td>
                <td>{{.Description}}</td>
                <td>{{.Quantity}}</td>
                <td>${{.Rate}}/hr</td>
                <td class="amount">${{.Total}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>

    <div class="totals">
        <div class="totals-row">
            <span>Subtotal</span>
            <span>${{.Invoice.Subtotal}}</span>
        </div>
        {{if gt .Invoice.Tax 0}}
        <div class="totals-row">
            <span>Tax ({{.Invoice.TaxRate}}%)</span>
            <span>${{.Invoice.Tax}}</span>
        </div>
        {{end}}
        <div class="totals-row total">
            <span>Total</span>
            <span>${{.Invoice.Total}}</span>
        </div>
    </div>

    <div class="footer">
        <p>Payment is due within {{.Invoice.PaymentTerms}} days.</p>
        <p>Thank you for your business!</p>
    </div>
</body>
</html>`

// App represents the main application with dependency injection
type App struct {
	logger        *cli.SimpleLogger
	configService *config.ConfigService
	rootCmd       *cobra.Command
}

// NewApp creates a new application instance with dependency injection
func NewApp() *App {
	logger := cli.NewLogger(false) // Debug mode controlled by flag
	validator := config.NewSimpleValidator(logger)
	configService := config.NewConfigService(logger, validator)

	app := &App{
		logger:        logger,
		configService: configService,
	}

	app.rootCmd = app.buildRootCommand()
	return app
}

// buildRootCommand constructs the root command with all subcommands
func (a *App) buildRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "go-invoice",
		Short: "A CLI-driven invoice generation tool for freelancers and contractors",
		Long: `go-invoice is a simple, efficient CLI tool for generating professional invoices.
It converts time tracking data from CSV files into printer-friendly HTML invoices,
with local JSON storage and customizable templates.

Key features:
- CSV import from any time tracking tool
- Professional HTML invoice generation
- Local JSON storage (no database required)
- Customizable business configuration
- Printer-optimized output`,
		Version: fmt.Sprintf("%s (%s, built %s)", Version, Commit, Date),
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// Get debug flag from command
			debug, _ := cmd.Flags().GetBool("debug")
			if debug {
				a.logger = cli.NewLogger(true)
				// Update config service with debug logger
				validator := config.NewSimpleValidator(a.logger)
				a.configService = config.NewConfigService(a.logger, validator)
			}
			return nil
		},
	}

	// Add persistent flags
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().String("config", ".env.config", "Path to configuration file")

	// Add subcommands
	rootCmd.AddCommand(a.buildConfigCommand())
	rootCmd.AddCommand(a.buildInitCommand())
	rootCmd.AddCommand(a.buildInvoiceCommand())
	rootCmd.AddCommand(a.buildImportCommand())
	rootCmd.AddCommand(a.buildGenerateCommand())

	return rootCmd
}

// buildConfigCommand creates the config command with subcommands
func (a *App) buildConfigCommand() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management commands",
		Long:  "Manage application configuration including validation and display",
	}

	// Add config subcommands
	configCmd.AddCommand(a.buildConfigSetupCommand())
	configCmd.AddCommand(a.buildConfigSetupClaudeCommand())
	configCmd.AddCommand(a.buildConfigValidateCommand())
	configCmd.AddCommand(a.buildConfigShowCommand())

	return configCmd
}

// buildConfigSetupCommand creates the config setup subcommand
func (a *App) buildConfigSetupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Set up initial business configuration",
		Long: `Interactive setup wizard to configure your business information, 
invoice settings, and storage preferences. This will create a .env.config 
file with your settings.

This wizard will prompt you for:
- Business information (name, address, email, etc.)
- Invoice settings (prefix, currency, VAT rate, etc.)
- Payment terms and banking details
- Storage directory preferences`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			configPath, _ := cmd.Flags().GetString("config")
			if configPath == "" {
				configPath = ".env.config"
			}

			return a.runConfigSetup(ctx, configPath)
		},
	}
}

// buildConfigValidateCommand creates the config validate subcommand
func (a *App) buildConfigValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file",
		Long:  "Validate the configuration file for syntax and required fields",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			configPath, _ := cmd.Flags().GetString("config")

			a.logger.Info("validating configuration", "path", configPath)

			config, err := a.configService.LoadConfig(ctx, configPath)
			if err != nil {
				return fmt.Errorf("configuration validation failed: %w", err)
			}

			a.logger.Info("configuration is valid", "business", config.Business.Name,
				"currency", config.Invoice.Currency, "prefix", config.Invoice.Prefix)
			a.logger.Println("✅ Configuration is valid")
			return nil
		},
	}
}

// buildConfigShowCommand creates the config show subcommand
func (a *App) buildConfigShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display current configuration",
		Long:  "Display the current configuration with sensitive data masked",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			configPath, _ := cmd.Flags().GetString("config")

			config, err := a.configService.LoadConfig(ctx, configPath)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			a.displayConfig(config)
			return nil
		},
	}
}

// buildInvoiceCommand creates the invoice command with all subcommands
func (a *App) buildInvoiceCommand() *cobra.Command {
	return a.buildInvoiceCommandImpl()
}

// buildImportCommand creates the import command for CSV import functionality
func (a *App) buildImportCommand() *cobra.Command {
	// Placeholder - this would be implemented in import.go
	return &cobra.Command{
		Use:   "import",
		Short: "Import work items from CSV",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("import command not yet implemented")
		},
	}
}

// buildGenerateCommand creates the generate command for invoice generation
func (a *App) buildGenerateCommand() *cobra.Command {
	// Placeholder - this would be implemented in generate.go
	return &cobra.Command{
		Use:   "generate",
		Short: "Generate invoice documents",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("generate command not yet implemented")
		},
	}
}

// buildConfigSetupClaudeCommand creates the config setup-claude subcommand
func (a *App) buildConfigSetupClaudeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "setup-claude",
		Short: "Set up Claude integration",
		Long:  "Configure go-invoice for integration with Claude AI assistant",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("claude setup command not yet implemented")
		},
	}
}

// buildInitCommand creates the init command for storage initialization
func (a *App) buildInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize storage system",
		Long: `Initialize the local storage system by creating the necessary directories 
and metadata files for invoice and client data.

This command must be run before using other invoice management commands.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			configPath, _ := cmd.Flags().GetString("config")

			a.logger.Info("initializing storage system")
			a.logger.Println("🔧 Initializing go-invoice storage...")

			// Load configuration to get storage settings
			config, err := a.configService.LoadConfig(ctx, configPath)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Initialize storage system
			if err := a.initializeStorage(ctx, config); err != nil {
				a.logger.Printf("❌ Storage initialization failed: %v\n", err)
				return err
			}

			a.logger.Println("✅ Storage system initialized successfully!")
			a.logger.Printf("   Data directory: %s\n", config.Storage.DataDir)
			a.logger.Printf("   Backup directory: %s\n", config.Storage.BackupDir)
			a.logger.Println("")
			a.logger.Println("💡 Next steps:")
			a.logger.Println("   • Create clients with: go-invoice client create")
			a.logger.Println("   • Create invoices with: go-invoice invoice create")
			a.logger.Println("   • View help with: go-invoice --help")

			return nil
		},
	}
}

// initializeStorage sets up the storage system using the provided configuration
func (a *App) initializeStorage(ctx context.Context, config *config.Config) error {
	// Create storage instance
	storage := a.createJSONStorage(config.Storage.DataDir)

	// Check if already initialized
	if initialized, err := storage.IsInitialized(ctx); err != nil {
		return fmt.Errorf("failed to check initialization status: %w", err)
	} else if initialized {
		a.logger.Info("storage already initialized")
		a.logger.Println("⚠️  Storage is already initialized")
		return nil
	}

	// Initialize storage
	if err := storage.Initialize(ctx); err != nil {
		return fmt.Errorf("storage initialization failed: %w", err)
	}

	// Validate the initialized storage
	if err := storage.Validate(ctx); err != nil {
		return fmt.Errorf("storage validation failed: %w", err)
	}

	a.logger.Info("storage system initialized", "data_dir", config.Storage.DataDir)
	return nil
}

// createJSONStorage creates a new JSON storage instance
func (a *App) createJSONStorage(dataDir string) storage.StorageInitializer {
	return jsonStorage.NewJSONStorage(dataDir, a.logger)
}

// displayConfig prints the configuration in a user-friendly format
func (a *App) displayConfig(config *config.Config) {
	a.logger.Println("📋 Current Configuration")
	a.logger.Println("========================")
	a.logger.Println("")

	a.logger.Println("🏢 Business Information:")
	a.logger.Printf("  Name: %s\n", config.Business.Name)
	a.logger.Printf("  Email: %s\n", config.Business.Email)
	a.logger.Printf("  Address: %s\n", config.Business.Address)
	if config.Business.Phone != "" {
		a.logger.Printf("  Phone: %s\n", config.Business.Phone)
	}
	if config.Business.Website != "" {
		a.logger.Printf("  Website: %s\n", config.Business.Website)
	}
	a.logger.Printf("  Payment Terms: %s\n", config.Business.PaymentTerms)
	a.logger.Println("")

	a.logger.Println("🧾 Invoice Settings:")
	a.logger.Printf("  Prefix: %s\n", config.Invoice.Prefix)
	a.logger.Printf("  Start Number: %d\n", config.Invoice.StartNumber)
	a.logger.Printf("  Currency: %s\n", config.Invoice.Currency)
	a.logger.Printf("  Default Due Days: %d\n", config.Invoice.DefaultDueDays)
	if config.Invoice.VATRate > 0 {
		a.logger.Printf("  VAT Rate: %.1f%%\n", config.Invoice.VATRate*100)
	}
	a.logger.Println("")

	a.logger.Println("💾 Storage Settings:")
	a.logger.Printf("  Data Directory: %s\n", config.Storage.DataDir)
	a.logger.Printf("  Backup Directory: %s\n", config.Storage.BackupDir)
	a.logger.Printf("  Auto Backup: %v\n", config.Storage.AutoBackup)
	if config.Storage.AutoBackup {
		a.logger.Printf("  Backup Interval: %v\n", config.Storage.BackupInterval)
	}
	a.logger.Println("")
}

// runConfigSetup runs the interactive configuration setup wizard
func (a *App) runConfigSetup(ctx context.Context, configPath string) error {
	prompter := cli.NewPrompter(a.logger)

	a.logger.Println("🚀 Welcome to go-invoice setup!")
	a.logger.Println("This wizard will help you configure your business information and invoice settings.")
	a.logger.Println("")

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		a.logger.Printf("⚠️  Configuration file already exists at: %s\n", configPath)
		overwrite, err := prompter.PromptBool(ctx, "Do you want to overwrite it?", false)
		if err != nil {
			return fmt.Errorf("failed to get user confirmation: %w", err)
		}
		if !overwrite {
			a.logger.Println("Setup canceled. Existing configuration preserved.")
			return nil
		}
		a.logger.Println("")
	}

	a.logger.Println("📋 Let's start with your business information...")
	a.logger.Println("")

	// Collect business information
	businessName, err := prompter.PromptStringRequired(ctx, "Business Name")
	if err != nil {
		return fmt.Errorf("failed to get business name: %w", err)
	}

	businessEmail, err := prompter.PromptStringRequired(ctx, "Business Email")
	if err != nil {
		return fmt.Errorf("failed to get business email: %w", err)
	}

	businessAddress, err := prompter.PromptStringRequired(ctx, "Business Address (use \\n for line breaks)")
	if err != nil {
		return fmt.Errorf("failed to get business address: %w", err)
	}

	businessPhone, err := prompter.PromptString(ctx, "Business Phone (optional)", "")
	if err != nil {
		return fmt.Errorf("failed to get business phone: %w", err)
	}

	businessWebsite, err := prompter.PromptString(ctx, "Business Website (optional)", "")
	if err != nil {
		return fmt.Errorf("failed to get business website: %w", err)
	}

	paymentTerms, err := prompter.PromptString(ctx, "Payment Terms", "Net 30")
	if err != nil {
		return fmt.Errorf("failed to get payment terms: %w", err)
	}

	a.logger.Println("")
	a.logger.Println("💳 Banking information (optional)...")
	a.logger.Println("")

	bankName, err := prompter.PromptString(ctx, "Bank Name (optional)", "")
	if err != nil {
		return fmt.Errorf("failed to get bank name: %w", err)
	}

	bankAccount, err := prompter.PromptString(ctx, "Bank Account Number (optional)", "")
	if err != nil {
		return fmt.Errorf("failed to get bank account: %w", err)
	}

	bankRouting, err := prompter.PromptString(ctx, "Bank Routing Number (optional)", "")
	if err != nil {
		return fmt.Errorf("failed to get bank routing: %w", err)
	}

	paymentInstructions, err := prompter.PromptString(ctx, "Payment Instructions (optional)", "")
	if err != nil {
		return fmt.Errorf("failed to get payment instructions: %w", err)
	}

	a.logger.Println("")
	a.logger.Println("🧾 Invoice settings...")
	a.logger.Println("")

	invoicePrefix, err := prompter.PromptString(ctx, "Invoice Prefix", "INV")
	if err != nil {
		return fmt.Errorf("failed to get invoice prefix: %w", err)
	}

	invoiceStartNumber, err := prompter.PromptInt(ctx, "Invoice Start Number", 1000)
	if err != nil {
		return fmt.Errorf("failed to get invoice start number: %w", err)
	}

	// Currency selection
	currencies := []string{"USD", "EUR", "GBP", "CAD", "AUD", "JPY", "Other"}
	_, currency, err := prompter.PromptSelect(ctx, "Select your currency:", currencies, 0)
	if err != nil {
		return fmt.Errorf("failed to get currency: %w", err)
	}

	if currency == "Other" {
		currency, err = prompter.PromptStringRequired(ctx, "Enter your currency code (e.g., SEK, NOK)")
		if err != nil {
			return fmt.Errorf("failed to get custom currency: %w", err)
		}
	}

	vatRate, err := prompter.PromptFloat(ctx, "VAT/Tax Rate (as decimal, e.g., 0.10 for 10%)", 0.0)
	if err != nil {
		return fmt.Errorf("failed to get VAT rate: %w", err)
	}

	invoiceDueDays, err := prompter.PromptInt(ctx, "Default invoice due days", 30)
	if err != nil {
		return fmt.Errorf("failed to get invoice due days: %w", err)
	}

	a.logger.Println("")
	a.logger.Println("💾 Storage settings...")
	a.logger.Println("")

	defaultDataDir := filepath.Join(os.Getenv("HOME"), ".go-invoice")
	dataDir, err := prompter.PromptString(ctx, "Data Directory", defaultDataDir)
	if err != nil {
		return fmt.Errorf("failed to get data directory: %w", err)
	}

	autoBackup, err := prompter.PromptBool(ctx, "Enable automatic backups?", true)
	if err != nil {
		return fmt.Errorf("failed to get auto backup setting: %w", err)
	}

	// Generate the configuration file content
	configContent := a.generateConfigFileContent(
		businessName, businessEmail, businessAddress, businessPhone, businessWebsite,
		paymentTerms, bankName, bankAccount, bankRouting, paymentInstructions,
		invoicePrefix, invoiceStartNumber, currency, vatRate, invoiceDueDays,
		dataDir, autoBackup,
	)

	// Write the configuration file
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	// Create templates directory and default invoice template
	if err := a.createDefaultTemplate(); err != nil {
		a.logger.Printf("⚠️  Warning: Could not create default template: %v\n", err)
		a.logger.Println("   You may need to create templates/invoice.html manually")
	}

	a.logger.Println("")
	a.logger.Println("✅ Configuration setup completed successfully!")
	a.logger.Printf("📁 Configuration saved to: %s\n", configPath)
	a.logger.Println("")
	a.logger.Println("💡 Next steps:")
	a.logger.Println("   • Initialize storage with: go-invoice init")
	a.logger.Println("   • Add clients with: go-invoice client add")
	a.logger.Println("   • View your config with: go-invoice config show")

	return nil
}

// generateConfigFileContent creates the .env configuration file content
func (a *App) generateConfigFileContent(
	businessName, businessEmail, businessAddress, businessPhone, businessWebsite,
	paymentTerms, bankName, bankAccount, bankRouting, paymentInstructions,
	invoicePrefix string, invoiceStartNumber int, currency string, vatRate float64,
	invoiceDueDays int, dataDir string, autoBackup bool,
) string {
	var content strings.Builder

	content.WriteString("# go-invoice Configuration\n")
	content.WriteString("# Generated by setup wizard\n\n")

	content.WriteString("# Business Information\n")
	content.WriteString(fmt.Sprintf("BUSINESS_NAME=%s\n", businessName))
	content.WriteString(fmt.Sprintf("BUSINESS_EMAIL=%s\n", businessEmail))
	content.WriteString(fmt.Sprintf("BUSINESS_ADDRESS=%s\n", businessAddress))
	if businessPhone != "" {
		content.WriteString(fmt.Sprintf("BUSINESS_PHONE=%s\n", businessPhone))
	}
	if businessWebsite != "" {
		content.WriteString(fmt.Sprintf("BUSINESS_WEBSITE=%s\n", businessWebsite))
	}
	content.WriteString(fmt.Sprintf("PAYMENT_TERMS=%s\n", paymentTerms))
	content.WriteString("\n")

	if bankName != "" || bankAccount != "" || bankRouting != "" || paymentInstructions != "" {
		content.WriteString("# Banking Information\n")
		if bankName != "" {
			content.WriteString(fmt.Sprintf("BANK_NAME=%s\n", bankName))
		}
		if bankAccount != "" {
			content.WriteString(fmt.Sprintf("BANK_ACCOUNT=%s\n", bankAccount))
		}
		if bankRouting != "" {
			content.WriteString(fmt.Sprintf("BANK_ROUTING=%s\n", bankRouting))
		}
		if paymentInstructions != "" {
			content.WriteString(fmt.Sprintf("PAYMENT_INSTRUCTIONS=%s\n", paymentInstructions))
		}
		content.WriteString("\n")
	}

	content.WriteString("# Invoice Settings\n")
	content.WriteString(fmt.Sprintf("INVOICE_PREFIX=%s\n", invoicePrefix))
	content.WriteString(fmt.Sprintf("INVOICE_START_NUMBER=%d\n", invoiceStartNumber))
	content.WriteString(fmt.Sprintf("CURRENCY=%s\n", currency))
	if vatRate > 0 {
		content.WriteString(fmt.Sprintf("VAT_RATE=%.4f\n", vatRate))
	}
	content.WriteString(fmt.Sprintf("INVOICE_DUE_DAYS=%d\n", invoiceDueDays))
	content.WriteString("\n")

	content.WriteString("# Storage Settings\n")
	content.WriteString(fmt.Sprintf("DATA_DIR=%s\n", dataDir))
	content.WriteString(fmt.Sprintf("AUTO_BACKUP=%s\n", strconv.FormatBool(autoBackup)))

	return content.String()
}

// createDefaultTemplate creates the templates directory and default invoice template
func (a *App) createDefaultTemplate() error {
	// Create templates directory
	templatesDir := "templates"
	if err := os.MkdirAll(templatesDir, 0o750); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	// Check if invoice.html already exists
	templatePath := filepath.Join(templatesDir, "invoice.html")
	if _, err := os.Stat(templatePath); err == nil {
		a.logger.Println("📄 Invoice template already exists, skipping creation")
		return nil
	}

	// Write the embedded template
	if err := os.WriteFile(templatePath, []byte(defaultInvoiceTemplate), 0o600); err != nil {
		return fmt.Errorf("failed to write invoice template: %w", err)
	}

	a.logger.Println("📄 Created default invoice template at templates/invoice.html")
	return nil
}

// Execute runs the application with context cancellation support
func (a *App) Execute() error {
	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		a.logger.Info("received interrupt signal, shutting down...")
		cancel()
	}()

	// Execute with context
	return a.rootCmd.ExecuteContext(ctx)
}

func main() {
	app := NewApp()

	if err := app.Execute(); err != nil {
		app.logger.Error("application failed", "error", err)
		os.Exit(1)
	}
}
