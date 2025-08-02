package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mrz/go-invoice/internal/cli"
	"github.com/mrz/go-invoice/internal/config"
	"github.com/mrz/go-invoice/internal/storage"
	jsonStorage "github.com/mrz/go-invoice/internal/storage/json"
	"github.com/spf13/cobra"
)

// Version information set by build process
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
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
	configCmd.AddCommand(a.buildConfigValidateCommand())
	configCmd.AddCommand(a.buildConfigShowCommand())

	return configCmd
}

// buildConfigValidateCommand creates the config validate subcommand
func (a *App) buildConfigValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file",
		Long:  "Validate the configuration file for syntax and required fields",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			fmt.Println("✅ Configuration is valid")
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
		RunE: func(cmd *cobra.Command, args []string) error {
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

// buildInitCommand creates the init command for storage initialization
func (a *App) buildInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize storage system",
		Long: `Initialize the local storage system by creating the necessary directories 
and metadata files for invoice and client data.

This command must be run before using other invoice management commands.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			configPath, _ := cmd.Flags().GetString("config")

			a.logger.Info("initializing storage system")
			fmt.Println("🔧 Initializing go-invoice storage...")

			// Load configuration to get storage settings
			config, err := a.configService.LoadConfig(ctx, configPath)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Initialize storage system
			if err := a.initializeStorage(ctx, config); err != nil {
				fmt.Printf("❌ Storage initialization failed: %v\n", err)
				return err
			}

			fmt.Println("✅ Storage system initialized successfully!")
			fmt.Printf("   Data directory: %s\n", config.Storage.DataDir)
			fmt.Printf("   Backup directory: %s\n", config.Storage.BackupDir)
			fmt.Println()
			fmt.Println("💡 Next steps:")
			fmt.Println("   • Create clients with: go-invoice client create")
			fmt.Println("   • Create invoices with: go-invoice invoice create")
			fmt.Println("   • View help with: go-invoice --help")

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
		fmt.Println("⚠️  Storage is already initialized")
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
	fmt.Println("📋 Current Configuration")
	fmt.Println("========================")
	fmt.Println()

	fmt.Println("🏢 Business Information:")
	fmt.Printf("  Name: %s\n", config.Business.Name)
	fmt.Printf("  Email: %s\n", config.Business.Email)
	fmt.Printf("  Address: %s\n", config.Business.Address)
	if config.Business.Phone != "" {
		fmt.Printf("  Phone: %s\n", config.Business.Phone)
	}
	if config.Business.Website != "" {
		fmt.Printf("  Website: %s\n", config.Business.Website)
	}
	fmt.Printf("  Payment Terms: %s\n", config.Business.PaymentTerms)
	fmt.Println()

	fmt.Println("🧾 Invoice Settings:")
	fmt.Printf("  Prefix: %s\n", config.Invoice.Prefix)
	fmt.Printf("  Start Number: %d\n", config.Invoice.StartNumber)
	fmt.Printf("  Currency: %s\n", config.Invoice.Currency)
	fmt.Printf("  Default Due Days: %d\n", config.Invoice.DefaultDueDays)
	if config.Invoice.VATRate > 0 {
		fmt.Printf("  VAT Rate: %.1f%%\n", config.Invoice.VATRate*100)
	}
	fmt.Println()

	fmt.Println("💾 Storage Settings:")
	fmt.Printf("  Data Directory: %s\n", config.Storage.DataDir)
	fmt.Printf("  Backup Directory: %s\n", config.Storage.BackupDir)
	fmt.Printf("  Auto Backup: %v\n", config.Storage.AutoBackup)
	if config.Storage.AutoBackup {
		fmt.Printf("  Backup Interval: %v\n", config.Storage.BackupInterval)
	}
	fmt.Println()
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
