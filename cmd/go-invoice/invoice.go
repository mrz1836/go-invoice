package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/mrz/go-invoice/internal/cli"
	"github.com/mrz/go-invoice/internal/config"
	"github.com/mrz/go-invoice/internal/models"
	"github.com/mrz/go-invoice/internal/services"
	"github.com/mrz/go-invoice/internal/storage"
	jsonStorage "github.com/mrz/go-invoice/internal/storage/json"
	"github.com/spf13/cobra"
)

// Invoice command errors
var (
	ErrClientNameRequired          = fmt.Errorf("client name is required (use --client or --interactive)")
	ErrCannotDeletePaidInvoice     = fmt.Errorf("cannot delete paid invoice")
	ErrNoUpdatesSpecified          = fmt.Errorf("no updates specified")
	ErrEmailRequiredForNewClient   = fmt.Errorf("email is required when creating a new client")
	ErrCannotUpdateInvoiceStatus   = fmt.Errorf("cannot update invoice with status")
	ErrInvalidStatus               = fmt.Errorf("invalid status")
	ErrSpecifyMoreSpecific         = fmt.Errorf("please specify a more specific client name")
	ErrClientNotFound              = fmt.Errorf("client not found")
	ErrNoClientsFound              = fmt.Errorf("no clients found matching")
	ErrMultipleClientsFound        = fmt.Errorf("multiple clients found matching")
	ErrHourlyLineItemRequiresFlags = fmt.Errorf("hourly line items require --hours and --rate flags")
	ErrFixedLineItemRequiresAmount = fmt.Errorf("fixed line items require --amount flag")
	ErrQuantityLineItemRequiresAll = fmt.Errorf("quantity line items require --quantity and --unit-price flags")
	ErrInvalidLineItemType         = fmt.Errorf("invalid line item type (must be hourly, fixed, or quantity)")
)

// getInvoiceByIDOrNumber is a helper function to get an invoice by ID or number
func (a *App) getInvoiceByIDOrNumber(ctx context.Context, invoiceService *services.InvoiceService, identifier string) (*models.Invoice, error) {
	// Try by ID first
	invoice, err := invoiceService.GetInvoice(ctx, models.InvoiceID(identifier))
	if err != nil {
		// If not found by ID, try by number
		invoice, err = invoiceService.GetInvoiceByNumber(ctx, identifier)
		if err != nil {
			return nil, fmt.Errorf("%w: '%s'", models.ErrInvoiceNotFound, identifier)
		}
	}
	return invoice, nil
}

// buildInvoiceCommand creates the invoice command with all subcommands
func (a *App) buildInvoiceCommand() *cobra.Command {
	// Ensure cli package is marked as used (a.logger is *cli.SimpleLogger)
	_ = (*cli.SimpleLogger)(nil)
	invoiceCmd := &cobra.Command{
		Use:   "invoice",
		Short: "Invoice management commands",
		Long:  "Create, list, show, update, and delete invoices",
	}

	// Add invoice subcommands
	invoiceCmd.AddCommand(a.buildInvoiceCreateCommand())
	invoiceCmd.AddCommand(a.buildInvoiceListCommand())
	invoiceCmd.AddCommand(a.buildInvoiceShowCommand())
	invoiceCmd.AddCommand(a.buildInvoiceUpdateCommand())
	invoiceCmd.AddCommand(a.buildInvoiceDeleteCommand())
	invoiceCmd.AddCommand(a.buildInvoiceAddLineItemCommand())

	return invoiceCmd
}

// buildInvoiceCreateCommand creates the invoice create subcommand
func (a *App) buildInvoiceCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new invoice",
		Long: `Create a new invoice with the specified client and options.

The command will automatically generate the next invoice number based on your configuration.
If the specified client doesn't exist and --create-client is used, it will be created.`,
		Example: `  # Create invoice for existing client
  go-invoice invoice create --client "Acme Corp"

  # Create invoice with specific dates
  go-invoice invoice create --client "Acme Corp" --date 2024-01-15 --due-date 2024-02-15

  # Create invoice and client if needed
  go-invoice invoice create --client "New Client" --create-client --email "client@example.com"

  # Interactive mode
  go-invoice invoice create --interactive`,
		RunE: a.runInvoiceCreate,
	}

	// Add flags
	cmd.Flags().String("client", "", "Client name or ID (required unless --interactive)")
	cmd.Flags().String("date", "", "Invoice date (default: today)")
	cmd.Flags().String("due-date", "", "Due date (default: based on payment terms)")
	cmd.Flags().String("description", "", "Invoice description")
	cmd.Flags().Bool("interactive", false, "Interactive mode to prompt for missing information")
	cmd.Flags().Bool("create-client", false, "Create client if it doesn't exist")
	cmd.Flags().String("email", "", "Client email (required when creating new client)")
	cmd.Flags().String("address", "", "Client address (when creating new client)")
	cmd.Flags().String("phone", "", "Client phone (when creating new client)")

	return cmd
}

// runInvoiceCreate handles the invoice create command
func (a *App) runInvoiceCreate(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	// Load configuration
	configPath, _ := cmd.Flags().GetString("config")
	config, err := a.configService.LoadConfig(ctx, configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create storage and services
	invoiceStorage, clientStorage := a.createStorageInstances(config.Storage.DataDir)
	idGen := services.NewUUIDGenerator()
	invoiceService := services.NewInvoiceService(invoiceStorage, clientStorage, a.logger, idGen)
	clientService := services.NewClientService(clientStorage, invoiceStorage, a.logger, idGen)

	// Get flags
	clientName, _ := cmd.Flags().GetString("client")
	dateStr, _ := cmd.Flags().GetString("date")
	dueDateStr, _ := cmd.Flags().GetString("due-date")
	description, _ := cmd.Flags().GetString("description")
	interactive, _ := cmd.Flags().GetBool("interactive")
	createClient, _ := cmd.Flags().GetBool("create-client")

	// Interactive mode
	if interactive {
		return a.runInvoiceCreateInteractive(ctx, invoiceService, clientService, config)
	}

	// Validate required fields
	if clientName == "" {
		return ErrClientNameRequired
	}

	// Parse dates
	invoiceDate := time.Now()
	if dateStr != "" {
		parsedDate, parseErr := time.Parse("2006-01-02", dateStr)
		if parseErr != nil {
			return fmt.Errorf("invalid date format (use YYYY-MM-DD): %w", parseErr)
		}
		invoiceDate = parsedDate
	}

	dueDate := invoiceDate.AddDate(0, 0, config.Invoice.DefaultDueDays)

	if dueDateStr != "" {
		parsedDueDate, parseErr := time.Parse("2006-01-02", dueDateStr)
		if parseErr != nil {
			return fmt.Errorf("invalid due date format (use YYYY-MM-DD): %w", parseErr)
		}
		dueDate = parsedDueDate
	}

	// Find or create client
	client, err := a.findOrCreateClient(ctx, clientService, clientName, createClient, cmd)
	if err != nil {
		return err
	}

	// Generate next invoice number
	nextNumber := a.generateNextInvoiceNumber(ctx, invoiceService, config.Invoice.Prefix, config.Invoice.StartNumber)

	// Create invoice
	req := models.CreateInvoiceRequest{
		Number:      nextNumber,
		Date:        invoiceDate,
		DueDate:     dueDate,
		ClientID:    client.ID,
		Description: description,
	}

	invoice, err := invoiceService.CreateInvoice(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	// Display success message
	a.logger.Printf("✅ Invoice created successfully!\n")
	a.logger.Printf("   Invoice Number: %s\n", invoice.Number)
	a.logger.Printf("   Client: %s\n", client.Name)
	a.logger.Printf("   Date: %s\n", invoice.Date.Format("2006-01-02"))
	a.logger.Printf("   Due Date: %s\n", invoice.DueDate.Format("2006-01-02"))
	a.logger.Printf("   Status: %s\n", invoice.Status)
	a.logger.Printf("\n")
	a.logger.Printf("💡 Next steps:\n")
	a.logger.Printf("   • Import work items: go-invoice import --file hours.csv --invoice %s\n", invoice.ID)
	a.logger.Printf("   • Generate invoice: go-invoice generate %s\n", invoice.ID)

	return nil
}

// buildInvoiceListCommand creates the invoice list subcommand
func (a *App) buildInvoiceListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List invoices",
		Long:  "List all invoices with optional filtering and sorting",
		Example: `  # List all invoices
  go-invoice invoice list

  # Filter by status
  go-invoice invoice list --status unpaid

  # Filter by client
  go-invoice invoice list --client "Acme Corp"

  # Filter by date range
  go-invoice invoice list --from 2024-01-01 --to 2024-12-31

  # Sort by amount descending
  go-invoice invoice list --sort amount --desc

  # Output as JSON
  go-invoice invoice list --output json`,
		RunE: a.runInvoiceList,
	}

	// Add flags
	cmd.Flags().String("status", "", "Filter by status (draft, sent, paid, overdue, canceled)")
	cmd.Flags().String("client", "", "Filter by client name or ID")
	cmd.Flags().String("from", "", "Filter from date (YYYY-MM-DD)")
	cmd.Flags().String("to", "", "Filter to date (YYYY-MM-DD)")
	cmd.Flags().String("sort", "date", "Sort by field (date, amount, status, client)")
	cmd.Flags().Bool("desc", false, "Sort in descending order")
	cmd.Flags().String("output", "table", "Output format (table, json, csv)")
	cmd.Flags().Int("limit", 0, "Limit number of results (0 = no limit)")
	cmd.Flags().Bool("summary", false, "Show summary statistics")

	return cmd
}

// runInvoiceList handles the invoice list command
func (a *App) runInvoiceList(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	// Load configuration
	configPath, _ := cmd.Flags().GetString("config")
	config, err := a.configService.LoadConfig(ctx, configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create storage and services
	invoiceStorage, clientStorage := a.createStorageInstances(config.Storage.DataDir)
	idGen := services.NewUUIDGenerator()
	invoiceService := services.NewInvoiceService(invoiceStorage, clientStorage, a.logger, idGen)
	clientService := services.NewClientService(clientStorage, invoiceStorage, a.logger, idGen)

	// Build filter from flags
	filter, err := a.buildInvoiceFilter(ctx, cmd, clientService)
	if err != nil {
		return err
	}

	// Get invoices
	result, err := invoiceService.ListInvoices(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to list invoices: %w", err)
	}
	invoices := result.Invoices

	// Get output format
	outputFormat, _ := cmd.Flags().GetString("output")
	showSummary, _ := cmd.Flags().GetBool("summary")

	// Display results based on format
	switch outputFormat {
	case "json":
		return a.outputInvoicesJSON(invoices)
	case "csv":
		a.outputInvoicesCSV(invoices)
		return nil
	default:
		if err := a.outputInvoicesTable(ctx, invoices, clientService); err != nil {
			return err
		}
		if showSummary {
			a.displayInvoiceSummary(invoices, config.Invoice.Currency)
		}
		return nil
	}
}

// buildInvoiceShowCommand creates the invoice show subcommand
func (a *App) buildInvoiceShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [invoice-id]",
		Short: "Show invoice details",
		Long:  "Display detailed information about a specific invoice",
		Args:  cobra.ExactArgs(1),
		Example: `  # Show invoice by ID
  go-invoice invoice show INV-001

  # Output as JSON
  go-invoice invoice show INV-001 --output json

  # Show with work items
  go-invoice invoice show INV-001 --show-items`,
		RunE: a.runInvoiceShow,
	}

	// Add flags
	cmd.Flags().String("output", "text", "Output format (text, json, yaml)")
	cmd.Flags().Bool("show-items", false, "Show detailed work items")
	cmd.Flags().Bool("show-history", false, "Show status history")

	return cmd
}

// runInvoiceShow handles the invoice show command
func (a *App) runInvoiceShow(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	invoiceID := args[0]

	// Load configuration
	configPath, _ := cmd.Flags().GetString("config")
	config, err := a.configService.LoadConfig(ctx, configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create storage and services
	invoiceStorage, clientStorage := a.createStorageInstances(config.Storage.DataDir)
	idGen := services.NewUUIDGenerator()
	invoiceService := services.NewInvoiceService(invoiceStorage, clientStorage, a.logger, idGen)
	clientService := services.NewClientService(clientStorage, invoiceStorage, a.logger, idGen)

	// Get invoice - try by ID first, then by number
	invoice, err := a.getInvoiceByIDOrNumber(ctx, invoiceService, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	// Get client
	client, err := clientService.GetClient(ctx, invoice.Client.ID)
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	// Get output format
	outputFormat, _ := cmd.Flags().GetString("output")
	showItems, _ := cmd.Flags().GetBool("show-items")
	showHistory, _ := cmd.Flags().GetBool("show-history")

	// Display based on format
	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(invoice, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal invoice: %w", err)
		}
		a.logger.Println(string(data))
	case "yaml":
		// For now, we'll use JSON format for YAML output
		// In a real implementation, we would use a YAML library
		data, err := json.MarshalIndent(invoice, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal invoice: %w", err)
		}
		a.logger.Println(string(data))
	default:
		a.displayInvoiceDetails(invoice, client, config.Invoice.Currency, showItems, showHistory)
	}

	return nil
}

// buildInvoiceUpdateCommand creates the invoice update subcommand
func (a *App) buildInvoiceUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [invoice-id]",
		Short: "Update an invoice",
		Long: `Update invoice details such as dates, status, or description.

Note: You cannot update invoices that are already paid or canceled.
Work items should be managed through the import command.`,
		Args: cobra.ExactArgs(1),
		Example: `  # Update invoice status
  go-invoice invoice update INV-001 --status sent

  # Update due date
  go-invoice invoice update INV-001 --due-date 2024-02-28

  # Update description
  go-invoice invoice update INV-001 --description "January consulting services"

  # Interactive update
  go-invoice invoice update INV-001 --interactive`,
		RunE: a.runInvoiceUpdate,
	}

	// Add flags
	cmd.Flags().String("status", "", "Update status (draft, sent, paid, overdue, canceled)")
	cmd.Flags().String("date", "", "Update invoice date (YYYY-MM-DD)")
	cmd.Flags().String("due-date", "", "Update due date (YYYY-MM-DD)")
	cmd.Flags().String("description", "", "Update description")
	cmd.Flags().String("notes", "", "Update internal notes")
	cmd.Flags().Bool("interactive", false, "Interactive mode to select fields to update")

	return cmd
}

// runInvoiceUpdate handles the invoice update command
func (a *App) runInvoiceUpdate(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	invoiceID := args[0]

	// Setup and validation
	invoiceService, invoice, config, err := a.setupUpdateCommand(ctx, cmd, invoiceID)
	if err != nil {
		return err
	}

	// Check if interactive mode
	if interactive, _ := cmd.Flags().GetBool("interactive"); interactive {
		return a.runInvoiceUpdateInteractive(ctx, invoiceService, invoice)
	}

	// Build update request - use the actual invoice ID from the retrieved invoice
	req, hasUpdates, err := a.buildUpdateRequest(cmd, string(invoice.ID), config)
	if err != nil {
		return err
	}

	if !hasUpdates {
		return ErrNoUpdatesSpecified
	}

	// Perform update and display results
	return a.executeUpdateAndDisplay(ctx, invoiceService, invoice, req)
}

// setupUpdateCommand sets up the invoice service and validates the invoice
func (a *App) setupUpdateCommand(ctx context.Context, cmd *cobra.Command, invoiceID string) (*services.InvoiceService, *models.Invoice, *config.Config, error) {
	// Load configuration
	configPath, _ := cmd.Flags().GetString("config")
	config, err := a.configService.LoadConfig(ctx, configPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create storage and services
	invoiceStorage, clientStorage := a.createStorageInstances(config.Storage.DataDir)
	idGen := services.NewUUIDGenerator()
	invoiceService := services.NewInvoiceService(invoiceStorage, clientStorage, a.logger, idGen)

	// Get current invoice - try by ID first, then by number
	invoice, err := a.getInvoiceByIDOrNumber(ctx, invoiceService, invoiceID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Check if invoice can be updated
	if invoice.Status == models.StatusPaid || invoice.Status == models.StatusVoided {
		return nil, nil, nil, fmt.Errorf("%w: %s", ErrCannotUpdateInvoiceStatus, invoice.Status)
	}

	return invoiceService, invoice, config, nil
}

// buildUpdateRequest builds the update request from command flags
func (a *App) buildUpdateRequest(cmd *cobra.Command, invoiceID string, cfg *config.Config) (models.UpdateInvoiceRequest, bool, error) {
	req := models.UpdateInvoiceRequest{
		ID: models.InvoiceID(invoiceID),
	}

	hasUpdates := false
	dateChanged := false

	// Update status
	if status, _ := cmd.Flags().GetString("status"); status != "" {
		if err := a.validateAndSetStatus(&req, status); err != nil {
			return req, false, err
		}
		hasUpdates = true
	}

	// Update invoice date
	if dateStr, _ := cmd.Flags().GetString("date"); dateStr != "" {
		if err := a.validateAndSetInvoiceDate(&req, dateStr); err != nil {
			return req, false, err
		}
		hasUpdates = true
		dateChanged = true
	}

	// Update due date
	dueDateStr, _ := cmd.Flags().GetString("due-date")
	if dueDateStr != "" {
		// User explicitly set due date, use it
		if err := a.validateAndSetDueDate(&req, dueDateStr); err != nil {
			return req, false, err
		}
		hasUpdates = true
	} else if dateChanged && req.Date != nil {
		// Invoice date changed but no explicit due date provided
		// Auto-calculate due date based on net terms
		dueDays := cfg.Invoice.DefaultDueDays
		if dueDays == 0 {
			dueDays = 30 // Default to 30 days if not configured
		}
		newDueDate := req.Date.AddDate(0, 0, dueDays)
		req.DueDate = &newDueDate
		a.logger.Printf("   Note: Due date automatically adjusted to %d days from invoice date\n", dueDays)
	}

	// Update description
	if description, _ := cmd.Flags().GetString("description"); description != "" {
		req.Description = &description
		hasUpdates = true
	}

	// Handle notes (not yet supported)
	if notes, _ := cmd.Flags().GetString("notes"); notes != "" {
		a.logger.Debug("notes update not yet supported", "notes", notes)
	}

	return req, hasUpdates, nil
}

// validateAndSetStatus validates and sets the status in the update request
func (a *App) validateAndSetStatus(req *models.UpdateInvoiceRequest, status string) error {
	validStatuses := []string{"draft", "sent", "paid", "overdue", "voided"}

	for _, vs := range validStatuses {
		if status == vs {
			req.Status = &status
			return nil
		}
	}

	return fmt.Errorf("%w: %s (must be one of: %s)", ErrInvalidStatus, status, strings.Join(validStatuses, ", "))
}

// validateAndSetInvoiceDate validates and sets the invoice date in the update request
func (a *App) validateAndSetInvoiceDate(req *models.UpdateInvoiceRequest, dateStr string) error {
	invoiceDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid invoice date format (use YYYY-MM-DD): %w", err)
	}
	req.Date = &invoiceDate
	return nil
}

// validateAndSetDueDate validates and sets the due date in the update request
func (a *App) validateAndSetDueDate(req *models.UpdateInvoiceRequest, dueDateStr string) error {
	dueDate, err := time.Parse("2006-01-02", dueDateStr)
	if err != nil {
		return fmt.Errorf("invalid due date format (use YYYY-MM-DD): %w", err)
	}
	req.DueDate = &dueDate
	return nil
}

// executeUpdateAndDisplay performs the update and displays results
func (a *App) executeUpdateAndDisplay(ctx context.Context, invoiceService *services.InvoiceService, originalInvoice *models.Invoice, req models.UpdateInvoiceRequest) error {
	// Perform update
	updatedInvoice, err := invoiceService.UpdateInvoice(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	// Display success message
	a.displayUpdateResults(originalInvoice, updatedInvoice, req)
	return nil
}

// displayUpdateResults displays the update results to the user
func (a *App) displayUpdateResults(original, updated *models.Invoice, req models.UpdateInvoiceRequest) {
	a.logger.Printf("✅ Invoice updated successfully!\n")
	a.logger.Printf("   Invoice Number: %s\n", updated.Number)

	if req.Status != nil {
		a.logger.Printf("   Status: %s → %s\n", original.Status, updated.Status)
	}

	if req.Date != nil {
		a.logger.Printf("   Invoice Date: %s → %s\n",
			original.Date.Format("2006-01-02"),
			updated.Date.Format("2006-01-02"))
	}

	if req.DueDate != nil {
		a.logger.Printf("   Due Date: %s → %s\n",
			original.DueDate.Format("2006-01-02"),
			updated.DueDate.Format("2006-01-02"))
	}

	if req.Description != nil {
		a.logger.Printf("   Description updated\n")
	}
}

// buildInvoiceDeleteCommand creates the invoice delete subcommand
func (a *App) buildInvoiceDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [invoice-id]",
		Short: "Delete an invoice",
		Long: `Delete an invoice from the system.

By default, this performs a soft delete (marks as deleted but retains data).
Use --hard to permanently remove the invoice and all associated data.

Note: You cannot delete invoices that are paid or have associated transactions.`,
		Args: cobra.ExactArgs(1),
		Example: `  # Soft delete (default)
  go-invoice invoice delete INV-001

  # Hard delete with confirmation
  go-invoice invoice delete INV-001 --hard

  # Force delete without confirmation (dangerous!)
  go-invoice invoice delete INV-001 --hard --force`,
		RunE: a.runInvoiceDelete,
	}

	// Add flags
	cmd.Flags().Bool("hard", false, "Permanently delete invoice (cannot be undone)")
	cmd.Flags().Bool("force", false, "Skip confirmation prompt")

	return cmd
}

// runInvoiceDelete handles the invoice delete command
func (a *App) runInvoiceDelete(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	invoiceID := args[0]

	// Load configuration
	configPath, _ := cmd.Flags().GetString("config")
	config, err := a.configService.LoadConfig(ctx, configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create storage and services
	invoiceStorage, clientStorage := a.createStorageInstances(config.Storage.DataDir)
	idGen := services.NewUUIDGenerator()
	invoiceService := services.NewInvoiceService(invoiceStorage, clientStorage, a.logger, idGen)

	// Get invoice to verify it exists and check status - try by ID first, then by number
	invoice, err := a.getInvoiceByIDOrNumber(ctx, invoiceService, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	// Check if invoice can be deleted
	if invoice.Status == models.StatusPaid {
		return ErrCannotDeletePaidInvoice
	}

	// Get storage for hard delete
	invoiceStorage, _ = a.createStorageInstances(config.Storage.DataDir)

	// Get flags
	hardDelete, _ := cmd.Flags().GetBool("hard")
	force, _ := cmd.Flags().GetBool("force")

	// Confirmation prompt
	if !force {
		deleteType := "soft delete"
		if hardDelete {
			deleteType = "PERMANENTLY DELETE"
		}

		a.logger.Printf("⚠️  About to %s invoice %s\n", deleteType, invoice.Number)
		a.logger.Printf("   Client: %s\n", invoice.Client.Name)
		a.logger.Printf("   Date: %s\n", invoice.Date.Format("2006-01-02"))
		a.logger.Printf("   Total: %.2f\n", invoice.Total)
		a.logger.Printf("\n")

		if hardDelete {
			a.logger.Printf("❗ This action CANNOT be undone!\n")
		}

		a.logger.Printf("Are you sure you want to continue? (yes/no): ")

		var response string
		if _, scanErr := fmt.Scanln(&response); scanErr != nil {
			// Handle non-interactive environments gracefully
			if strings.Contains(scanErr.Error(), "EOF") {
				a.logger.Printf("Non-interactive environment detected. Use --force to bypass confirmation.\n")
				return fmt.Errorf("%w: use --force flag to delete without confirmation", models.ErrConfirmationRequired)
			}
			a.logger.Printf("Error reading input: %v\n", scanErr)
			return fmt.Errorf("failed to read user input: %w", scanErr)
		}
		if strings.ToLower(response) != "yes" {
			a.logger.Println("❌ Delete canceled")
			return nil
		}
	}

	// Perform delete
	if hardDelete {
		// For hard delete, we'll delete from storage directly
		err = invoiceStorage.DeleteInvoice(ctx, models.InvoiceID(invoiceID))
		if err != nil {
			return fmt.Errorf("failed to delete invoice: %w", err)
		}
		a.logger.Printf("✅ Invoice %s permanently deleted\n", invoice.Number)
	} else {
		err = invoiceService.DeleteInvoice(ctx, models.InvoiceID(invoiceID))
		if err != nil {
			return fmt.Errorf("failed to delete invoice: %w", err)
		}
		a.logger.Printf("✅ Invoice %s deleted\n", invoice.Number)
	}

	return nil
}

// Helper methods

// createStorageInstances creates invoice and client storage instances
func (a *App) createStorageInstances(dataDir string) (storage.InvoiceStorage, storage.ClientStorage) {
	// Use a.logger directly - it satisfies the Logger interface
	jsonStore := jsonStorage.NewJSONStorage(dataDir, a.logger)
	return jsonStore, jsonStore
}

// findOrCreateClient finds an existing client or creates a new one if allowed
func (a *App) findOrCreateClient(ctx context.Context, clientService *services.ClientService, clientName string, createIfMissing bool, cmd *cobra.Command) (*models.Client, error) {
	// Try to find existing client
	clients, err := a.searchClientsByName(ctx, clientService, clientName)
	if err != nil {
		return nil, fmt.Errorf("failed to search for client: %w", err)
	}

	// If found, return the first exact match or the only result
	for _, client := range clients {
		if strings.EqualFold(client.Name, clientName) {
			return client, nil
		}
	}

	if len(clients) == 1 {
		return clients[0], nil
	}

	if len(clients) > 1 {
		// Multiple partial matches found
		a.logger.Printf("Multiple clients found matching '%s':\n", clientName)
		for i, client := range clients {
			a.logger.Printf("  %d. %s (%s)\n", i+1, client.Name, client.Email)
		}
		return nil, ErrSpecifyMoreSpecific
	}

	// No client found
	if !createIfMissing {
		return nil, fmt.Errorf("%w '%s' (use --create-client to create)", ErrClientNotFound, clientName)
	}

	// Create new client
	email, _ := cmd.Flags().GetString("email")
	if email == "" {
		return nil, ErrEmailRequiredForNewClient
	}

	address, _ := cmd.Flags().GetString("address")
	phone, _ := cmd.Flags().GetString("phone")

	req := models.CreateClientRequest{
		Name:    clientName,
		Email:   email,
		Address: address,
		Phone:   phone,
	}

	client, err := clientService.CreateClient(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	a.logger.Printf("✅ Created new client: %s\n", client.Name)
	return client, nil
}

// buildInvoiceFilter builds an invoice filter from command flags
func (a *App) buildInvoiceFilter(ctx context.Context, cmd *cobra.Command, clientService *services.ClientService) (models.InvoiceFilter, error) {
	filter := models.InvoiceFilter{}

	// Build status filter
	if err := a.buildStatusFilter(cmd, &filter); err != nil {
		return filter, err
	}

	// Build client filter
	if err := a.buildClientFilter(ctx, cmd, clientService, &filter); err != nil {
		return filter, err
	}

	// Build date range filter
	if err := a.buildDateRangeFilter(cmd, &filter); err != nil {
		return filter, err
	}

	// Build limit filter
	a.buildLimitFilter(cmd, &filter)

	return filter, nil
}

// buildStatusFilter builds the status filter from command flags
func (a *App) buildStatusFilter(cmd *cobra.Command, filter *models.InvoiceFilter) error {
	status, _ := cmd.Flags().GetString("status")
	if status == "" {
		return nil
	}

	validStatuses := []string{"draft", "sent", "paid", "overdue", "voided"}
	for _, vs := range validStatuses {
		if status == vs {
			filter.Status = status
			return nil
		}
	}

	return fmt.Errorf("%w: %s", ErrInvalidStatus, status)
}

// buildClientFilter builds the client filter from command flags
func (a *App) buildClientFilter(ctx context.Context, cmd *cobra.Command, clientService *services.ClientService, filter *models.InvoiceFilter) error {
	clientName, _ := cmd.Flags().GetString("client")
	if clientName == "" {
		return nil
	}

	clients, err := a.searchClientsByName(ctx, clientService, clientName)
	if err != nil {
		return fmt.Errorf("failed to search for client: %w", err)
	}

	if len(clients) == 0 {
		return fmt.Errorf("%w: %s", ErrNoClientsFound, clientName)
	}

	if len(clients) > 1 {
		return fmt.Errorf("%w: %s", ErrMultipleClientsFound, clientName)
	}

	filter.ClientID = clients[0].ID
	return nil
}

// buildDateRangeFilter builds the date range filter from command flags
func (a *App) buildDateRangeFilter(cmd *cobra.Command, filter *models.InvoiceFilter) error {
	// From date
	if fromStr, _ := cmd.Flags().GetString("from"); fromStr != "" {
		fromDate, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			return fmt.Errorf("invalid from date format (use YYYY-MM-DD): %w", err)
		}
		filter.DateFrom = fromDate
	}

	// To date
	if toStr, _ := cmd.Flags().GetString("to"); toStr != "" {
		toDate, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			return fmt.Errorf("invalid to date format (use YYYY-MM-DD): %w", err)
		}
		// Set to end of day
		toDate = toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filter.DateTo = toDate
	}

	return nil
}

// buildLimitFilter builds the limit filter from command flags
func (a *App) buildLimitFilter(cmd *cobra.Command, filter *models.InvoiceFilter) {
	if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
		filter.Limit = limit
	}
}

// Output formatting methods

func (a *App) outputInvoicesJSON(invoices []*models.Invoice) error {
	data, err := json.MarshalIndent(invoices, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal invoices: %w", err)
	}
	a.logger.Println(string(data))
	return nil
}

func (a *App) outputInvoicesCSV(invoices []*models.Invoice) {
	// CSV header
	a.logger.Println("Number,Date,DueDate,ClientName,Status,SubTotal,Tax,Total")

	// CSV rows
	for _, inv := range invoices {
		a.logger.Printf("%s,%s,%s,%s,%s,%.2f,%.2f,%.2f\n",
			inv.Number,
			inv.Date.Format("2006-01-02"),
			inv.DueDate.Format("2006-01-02"),
			inv.Client.Name,
			inv.Status,
			inv.Subtotal,
			inv.TaxAmount,
			inv.Total,
		)
	}
}

func (a *App) outputInvoicesTable(_ context.Context, invoices []*models.Invoice, _ *services.ClientService) error {
	if len(invoices) == 0 {
		a.logger.Println("No invoices found")
		return nil
	}

	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() {
		if err := w.Flush(); err != nil {
			// Log error but don't fail the function since this is cleanup
			fmt.Fprintf(os.Stderr, "Warning: failed to flush tabwriter: %v\n", err)
		}
	}()

	// Header
	if _, err := fmt.Fprintln(w, "NUMBER\tCLIENT\tDATE\tDUE DATE\tSTATUS\tAMOUNT"); err != nil {
		return fmt.Errorf("failed to write table header: %w", err)
	}
	if _, err := fmt.Fprintln(w, "------\t------\t----\t--------\t------\t------"); err != nil {
		return fmt.Errorf("failed to write table separator: %w", err)
	}

	// Rows
	for _, inv := range invoices {
		// Client name is already in invoice
		clientName := inv.Client.Name

		// Format status with color (in a real terminal)
		status := inv.Status

		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%.2f\n",
			inv.Number,
			clientName,
			inv.Date.Format("2006-01-02"),
			inv.DueDate.Format("2006-01-02"),
			status,
			inv.Total,
		); err != nil {
			return fmt.Errorf("failed to write table row for invoice %s: %w", inv.Number, err)
		}
	}

	return nil
}

func (a *App) displayInvoiceSummary(invoices []*models.Invoice, currency string) {
	var totalAmount, paidAmount, unpaidAmount float64
	var draftCount, sentCount, paidCount, overdueCount int

	for _, inv := range invoices {
		totalAmount += inv.Total

		switch inv.Status {
		case models.StatusDraft:
			draftCount++
			unpaidAmount += inv.Total
		case models.StatusSent:
			sentCount++
			unpaidAmount += inv.Total
		case models.StatusPaid:
			paidCount++
			paidAmount += inv.Total
		case models.StatusOverdue:
			overdueCount++
			unpaidAmount += inv.Total
		}
	}

	a.logger.Printf("\n📊 Summary\n")
	a.logger.Printf("─────────\n")
	a.logger.Printf("Total Invoices: %d\n", len(invoices))
	a.logger.Printf("  Draft: %d\n", draftCount)
	a.logger.Printf("  Sent: %d\n", sentCount)
	a.logger.Printf("  Paid: %d\n", paidCount)
	a.logger.Printf("  Overdue: %d\n", overdueCount)
	a.logger.Printf("\n")
	a.logger.Printf("Total Amount: %.2f %s\n", totalAmount, currency)
	a.logger.Printf("  Paid: %.2f %s\n", paidAmount, currency)
	a.logger.Printf("  Unpaid: %.2f %s\n", unpaidAmount, currency)
}

func (a *App) displayInvoiceDetails(invoice *models.Invoice, client *models.Client, currency string, showItems, _ bool) {
	a.logger.Printf("📄 Invoice %s\n", invoice.Number)
	a.logger.Printf("════════════════════\n")
	a.logger.Printf("\n")

	a.logger.Printf("Client: %s\n", client.Name)
	a.logger.Printf("Email: %s\n", client.Email)
	if client.Address != "" {
		a.logger.Printf("Address: %s\n", client.Address)
	}
	a.logger.Printf("\n")

	a.logger.Printf("Date: %s\n", invoice.Date.Format("2006-01-02"))
	a.logger.Printf("Due Date: %s\n", invoice.DueDate.Format("2006-01-02"))
	a.logger.Printf("Status: %s\n", invoice.Status)

	if invoice.Description != "" {
		a.logger.Printf("Description: %s\n", invoice.Description)
	}

	a.logger.Printf("\n")
	a.logger.Printf("💰 Financial Summary\n")
	a.logger.Printf("──────────────────\n")
	a.logger.Printf("Subtotal: %.2f %s\n", invoice.Subtotal, currency)
	if invoice.TaxAmount > 0 {
		a.logger.Printf("Tax: %.2f %s\n", invoice.TaxAmount, currency)
	}
	a.logger.Printf("Total: %.2f %s\n", invoice.Total, currency)

	if showItems && len(invoice.WorkItems) > 0 {
		a.logger.Printf("\n")
		a.logger.Printf("📋 Work Items\n")
		a.logger.Printf("───────────\n")

		for i, item := range invoice.WorkItems {
			a.logger.Printf("\n%d. %s\n", i+1, item.Description)
			a.logger.Printf("   Date: %s\n", item.Date.Format("2006-01-02"))
			a.logger.Printf("   Hours: %.2f @ %.2f/hour = %.2f %s\n",
				item.Hours, item.Rate, item.Total, currency)
		}
	}

	// Notes field not yet available in Invoice model

	a.logger.Printf("\n")
	a.logger.Printf("🕒 Timestamps\n")
	a.logger.Printf("───────────\n")
	a.logger.Printf("Created: %s\n", invoice.CreatedAt.Format("2006-01-02 15:04:05"))
	a.logger.Printf("Updated: %s\n", invoice.UpdatedAt.Format("2006-01-02 15:04:05"))
}

// Interactive mode helpers

func (a *App) runInvoiceCreateInteractive(ctx context.Context, invoiceService *services.InvoiceService, clientService *services.ClientService, config *config.Config) error {
	a.logger.Println("🔨 Create New Invoice - Interactive Mode")
	a.logger.Println("=====================================")
	a.logger.Println("")

	prompter := cli.NewPrompter(a.logger)

	// Prompt for client
	a.logger.Println("Step 1: Select or create client")
	a.logger.Println("-------------------------------")

	// List existing clients
	clientResult, err := clientService.ListClients(ctx, true, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to list clients: %w", err)
	}

	var client *models.Client
	if len(clientResult.Clients) > 0 {
		// Add option to create new client
		options := []string{"Create new client"}
		for _, c := range clientResult.Clients {
			options = append(options, fmt.Sprintf("%s (%s)", c.Name, c.Email))
		}

		index, _, selectErr := prompter.PromptSelect(ctx, "Select a client:", options, -1)
		if selectErr != nil {
			return fmt.Errorf("client selection canceled: %w", selectErr)
		}

		if index == 0 {
			// Create new client
			client, err = a.createClientInteractive(ctx, clientService, prompter)
			if err != nil {
				return err
			}
		} else {
			client = clientResult.Clients[index-1]
		}
	} else {
		// No clients exist, create new one
		a.logger.Println("No clients found. Let's create one.")
		client, err = a.createClientInteractive(ctx, clientService, prompter)
		if err != nil {
			return err
		}
	}

	a.logger.Printf("\n✅ Client selected: %s\n\n", client.Name)

	// Prompt for invoice details
	a.logger.Println("Step 2: Invoice details")
	a.logger.Println("----------------------")

	// Invoice date
	invoiceDate, err := prompter.PromptDate(ctx, "Invoice date", time.Now())
	if err != nil {
		return fmt.Errorf("date selection canceled: %w", err)
	}

	// Due date
	defaultDueDate := invoiceDate.AddDate(0, 0, config.Invoice.DefaultDueDays)
	dueDate, err := prompter.PromptDate(ctx, "Due date", defaultDueDate)
	if err != nil {
		return fmt.Errorf("due date selection canceled: %w", err)
	}

	// Description
	description, err := prompter.PromptString(ctx, "Invoice description (optional)", "")
	if err != nil {
		return fmt.Errorf("description input canceled: %w", err)
	}

	// Generate invoice number
	nextNumber := a.generateNextInvoiceNumber(ctx, invoiceService, config.Invoice.Prefix, config.Invoice.StartNumber)

	a.logger.Printf("\n📋 Invoice Summary:\n")
	a.logger.Printf("   Number: %s\n", nextNumber)
	a.logger.Printf("   Client: %s\n", client.Name)
	a.logger.Printf("   Date: %s\n", invoiceDate.Format("2006-01-02"))
	a.logger.Printf("   Due Date: %s\n", dueDate.Format("2006-01-02"))
	if description != "" {
		a.logger.Printf("   Description: %s\n", description)
	}
	a.logger.Printf("\n")

	// Confirm creation
	confirmed, err := prompter.PromptConfirm(ctx, "Create this invoice?")
	if err != nil {
		return fmt.Errorf("confirmation canceled: %w", err)
	}

	if !confirmed {
		a.logger.Println("❌ Invoice creation canceled")
		return nil
	}

	// Create invoice
	req := models.CreateInvoiceRequest{
		Number:      nextNumber,
		Date:        invoiceDate,
		DueDate:     dueDate,
		ClientID:    client.ID,
		Description: description,
	}

	invoice, err := invoiceService.CreateInvoice(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	// Display success message
	a.logger.Printf("\n✅ Invoice created successfully!\n")
	a.logger.Printf("   Invoice Number: %s\n", invoice.Number)
	a.logger.Printf("   Invoice ID: %s\n", invoice.ID)
	a.logger.Printf("\n")
	a.logger.Printf("💡 Next steps:\n")
	a.logger.Printf("   • Import work items: go-invoice import --file hours.csv --invoice %s\n", invoice.ID)
	a.logger.Printf("   • Generate invoice: go-invoice generate %s\n", invoice.ID)

	return nil
}

func (a *App) runInvoiceUpdateInteractive(ctx context.Context, invoiceService *services.InvoiceService, invoice *models.Invoice) error {
	a.logger.Printf("🔧 Update Invoice %s - Interactive Mode\n", invoice.Number)
	a.logger.Println("=====================================")
	a.logger.Println("")

	prompter := cli.NewPrompter(a.logger)

	// Show current invoice details
	a.logger.Println("Current Invoice Details:")
	a.logger.Printf("   Status: %s\n", invoice.Status)
	a.logger.Printf("   Date: %s\n", invoice.Date.Format("2006-01-02"))
	a.logger.Printf("   Due Date: %s\n", invoice.DueDate.Format("2006-01-02"))
	a.logger.Printf("   Description: %s\n", invoice.Description)
	a.logger.Println("")

	// Select what to update
	options := []string{
		"Update status",
		"Update due date",
		"Update description",
		"Cancel (no changes)",
	}

	index, _, err := prompter.PromptSelect(ctx, "What would you like to update?", options, -1)
	if err != nil {
		return fmt.Errorf("selection canceled: %w", err)
	}

	if index == 3 {
		a.logger.Println("❌ Update canceled")
		return nil
	}

	req := models.UpdateInvoiceRequest{
		ID: invoice.ID,
	}

	switch index {
	case 0: // Update status
		statuses := []string{"draft", "sent", "paid", "overdue", "voided"}
		statusIndex, newStatus, selectErr := prompter.PromptSelect(ctx, "Select new status:", statuses, -1)
		if selectErr != nil {
			return fmt.Errorf("status selection canceled: %w", selectErr)
		}
		_ = statusIndex // unused
		req.Status = &newStatus

	case 1: // Update due date
		newDueDate, promptErr := prompter.PromptDate(ctx, "New due date", invoice.DueDate)
		if promptErr != nil {
			return fmt.Errorf("due date selection canceled: %w", promptErr)
		}
		req.DueDate = &newDueDate

	case 2: // Update description
		newDescription, promptErr := prompter.PromptString(ctx, "New description", invoice.Description)
		if promptErr != nil {
			return fmt.Errorf("description input canceled: %w", promptErr)
		}
		req.Description = &newDescription
	}

	// Confirm update
	a.logger.Println("\n📋 Update Summary:")
	if req.Status != nil {
		a.logger.Printf("   Status: %s → %s\n", invoice.Status, *req.Status)
	}
	if req.DueDate != nil {
		a.logger.Printf("   Due Date: %s → %s\n", invoice.DueDate.Format("2006-01-02"), req.DueDate.Format("2006-01-02"))
	}
	if req.Description != nil {
		a.logger.Printf("   Description: %s → %s\n", invoice.Description, *req.Description)
	}
	a.logger.Println("")

	confirmed, err := prompter.PromptConfirm(ctx, "Apply these changes?")
	if err != nil {
		return fmt.Errorf("confirmation canceled: %w", err)
	}

	if !confirmed {
		a.logger.Println("❌ Update canceled")
		return nil
	}

	// Perform update
	updatedInvoice, err := invoiceService.UpdateInvoice(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	a.logger.Printf("\n✅ Invoice %s updated successfully!\n", updatedInvoice.Number)
	return nil
}

// createClientInteractive creates a new client through interactive prompts
func (a *App) createClientInteractive(ctx context.Context, clientService *services.ClientService, prompter *cli.Prompter) (*models.Client, error) {
	a.logger.Println("\n📝 Create New Client")
	a.logger.Println("-------------------")

	// Prompt for client details
	name, err := prompter.PromptStringRequired(ctx, "Client name")
	if err != nil {
		return nil, fmt.Errorf("name input canceled: %w", err)
	}

	email, err := prompter.PromptStringRequired(ctx, "Client email")
	if err != nil {
		return nil, fmt.Errorf("email input canceled: %w", err)
	}

	phone, err := prompter.PromptString(ctx, "Client phone (optional)", "")
	if err != nil {
		return nil, fmt.Errorf("phone input canceled: %w", err)
	}

	address, err := prompter.PromptString(ctx, "Client address (optional)", "")
	if err != nil {
		return nil, fmt.Errorf("address input canceled: %w", err)
	}

	// Create client
	req := models.CreateClientRequest{
		Name:    name,
		Email:   email,
		Phone:   phone,
		Address: address,
	}

	client, err := clientService.CreateClient(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	a.logger.Printf("✅ Client '%s' created successfully!\n", client.Name)
	return client, nil
}

// generateNextInvoiceNumber generates the next available invoice number using timestamp format
func (a *App) generateNextInvoiceNumber(_ context.Context, _ *services.InvoiceService, prefix string, _ int) string {
	// Generate invoice number based on current date and time
	now := time.Now()
	return fmt.Sprintf("%s-%s", prefix, now.Format("20060102-150405"))
}

// searchClientsByName searches for clients by name
func (a *App) searchClientsByName(ctx context.Context, clientService *services.ClientService, name string) ([]*models.Client, error) {
	// Get all clients and filter by name
	// This is a simple implementation - in a real system, you'd want server-side search
	result, err := clientService.ListClients(ctx, true, 0, 0) // activeOnly=true, no limit
	if err != nil {
		return nil, fmt.Errorf("failed to list clients: %w", err)
	}

	var matches []*models.Client
	nameLower := strings.ToLower(name)

	for _, client := range result.Clients {
		if strings.Contains(strings.ToLower(client.Name), nameLower) {
			matches = append(matches, client)
		}
	}

	return matches, nil
}

// buildInvoiceAddLineItemCommand creates the invoice add-line-item subcommand
func (a *App) buildInvoiceAddLineItemCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-line-item [invoice-id-or-number]",
		Short: "Add a custom line item to an invoice",
		Long: `Add a custom line item to an existing invoice. Supports hourly, fixed, and quantity-based line items.

Line Item Types:
  hourly   - Time-based billing (hours × rate)
  fixed    - Flat fee or fixed amount (retainers, setup fees)
  quantity - Quantity-based billing (quantity × unit price)`,
		Example: `  # Add hourly work item (default type)
  go-invoice invoice add-line-item INV-001 --description "Development work" --hours 8 --rate 125

  # Add monthly retainer (fixed amount)
  go-invoice invoice add-line-item INV-001 --type fixed --description "Monthly Retainer - August" --amount 2000

  # Add flat setup fee
  go-invoice invoice add-line-item INV-001 --type fixed --description "Project Setup Fee" --amount 500

  # Add quantity-based item (licenses, materials)
  go-invoice invoice add-line-item INV-001 --type quantity --description "SSL Certificates" --quantity 2 --unit-price 50`,
		Args: cobra.ExactArgs(1),
		RunE: a.runInvoiceAddLineItem,
	}

	// Common flags
	cmd.Flags().String("type", "hourly", "Line item type: hourly, fixed, or quantity")
	cmd.Flags().String("description", "", "Line item description (required)")
	cmd.Flags().String("date", "", "Line item date (default: today)")

	// Hourly flags
	cmd.Flags().Float64("hours", 0, "Hours worked (for hourly type)")
	cmd.Flags().Float64("rate", 0, "Hourly rate (for hourly type)")

	// Fixed flags
	cmd.Flags().Float64("amount", 0, "Fixed amount (for fixed type)")

	// Quantity flags
	cmd.Flags().Float64("quantity", 0, "Quantity (for quantity type)")
	cmd.Flags().Float64("unit-price", 0, "Unit price (for quantity type)")

	// Mark description as required
	_ = cmd.MarkFlagRequired("description")

	return cmd
}

// runInvoiceAddLineItem executes the invoice add-line-item command
func (a *App) runInvoiceAddLineItem(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get invoice identifier
	invoiceIdentifier := args[0]

	// Parse flags
	lineItemType, _ := cmd.Flags().GetString("type")
	description, _ := cmd.Flags().GetString("description")
	dateStr, _ := cmd.Flags().GetString("date")

	// Hourly flags
	hours, _ := cmd.Flags().GetFloat64("hours")
	rate, _ := cmd.Flags().GetFloat64("rate")

	// Fixed flags
	amount, _ := cmd.Flags().GetFloat64("amount")

	// Quantity flags
	quantity, _ := cmd.Flags().GetFloat64("quantity")
	unitPrice, _ := cmd.Flags().GetFloat64("unit-price")

	// Parse date
	var itemDate time.Time
	if dateStr != "" {
		var err error
		itemDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return fmt.Errorf("invalid date format (use YYYY-MM-DD): %w", err)
		}
	} else {
		itemDate = time.Now()
	}

	// Get config path from flag
	configPath, _ := cmd.Flags().GetString("config")

	// Load configuration
	config, err := a.configService.LoadConfig(ctx, configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize storage and services
	invoiceStorage, clientStorage := a.createStorageInstances(config.Storage.DataDir)
	idGen := services.NewUUIDGenerator()
	invoiceService := services.NewInvoiceService(invoiceStorage, clientStorage, a.logger, idGen)

	// Get invoice
	invoice, err := a.getInvoiceByIDOrNumber(ctx, invoiceService, invoiceIdentifier)
	if err != nil {
		return err
	}

	// Create line item based on type
	var lineItem models.LineItem

	switch models.LineItemType(lineItemType) {
	case models.LineItemTypeHourly:
		if hours == 0 || rate == 0 {
			return ErrHourlyLineItemRequiresFlags
		}
		lineItem = models.LineItem{
			Type:        models.LineItemTypeHourly,
			Date:        itemDate,
			Description: description,
			Hours:       &hours,
			Rate:        &rate,
			Total:       hours * rate,
			CreatedAt:   time.Now(),
		}

	case models.LineItemTypeFixed:
		if amount == 0 {
			return ErrFixedLineItemRequiresAmount
		}
		lineItem = models.LineItem{
			Type:        models.LineItemTypeFixed,
			Date:        itemDate,
			Description: description,
			Amount:      &amount,
			Total:       amount,
			CreatedAt:   time.Now(),
		}

	case models.LineItemTypeQuantity:
		if quantity == 0 || unitPrice == 0 {
			return ErrQuantityLineItemRequiresAll
		}
		lineItem = models.LineItem{
			Type:        models.LineItemTypeQuantity,
			Date:        itemDate,
			Description: description,
			Quantity:    &quantity,
			UnitPrice:   &unitPrice,
			Total:       quantity * unitPrice,
			CreatedAt:   time.Now(),
		}

	default:
		return fmt.Errorf("%w: %s", ErrInvalidLineItemType, lineItemType)
	}

	// Add line item to invoice
	updatedInvoice, err := invoiceService.AddLineItemToInvoice(ctx, invoice.ID, lineItem)
	if err != nil {
		return fmt.Errorf("failed to add line item: %w", err)
	}

	// Display success message
	a.logger.Printf("✅ Line item added to invoice %s\n\n", updatedInvoice.Number)
	a.logger.Printf("Type:        %s\n", lineItem.Type)
	a.logger.Printf("Description: %s\n", description)
	a.logger.Printf("Details:     %s\n", lineItem.GetDetails())
	a.logger.Printf("Amount:      %s\n\n", lineItem.GetFormattedTotal())
	a.logger.Printf("Updated Total: $%.2f\n", updatedInvoice.Total)

	return nil
}
