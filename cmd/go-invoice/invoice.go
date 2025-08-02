package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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
func (a *App) runInvoiceCreate(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("client name is required (use --client or --interactive)")
	}

	// Parse dates
	invoiceDate := time.Now()
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return fmt.Errorf("invalid date format (use YYYY-MM-DD): %w", err)
		}
		invoiceDate = parsedDate
	}

	dueDate := invoiceDate.AddDate(0, 0, config.Invoice.DefaultDueDays)
	if dueDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dueDateStr)
		if err != nil {
			return fmt.Errorf("invalid due date format (use YYYY-MM-DD): %w", err)
		}
		dueDate = parsedDate
	}

	// Find or create client
	client, err := a.findOrCreateClient(ctx, clientService, clientName, createClient, cmd)
	if err != nil {
		return err
	}

	// Generate next invoice number
	nextNumber, err := a.generateNextInvoiceNumber(ctx, invoiceService, config.Invoice.Prefix, config.Invoice.StartNumber)
	if err != nil {
		return fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Create invoice
	req := models.CreateInvoiceRequest{
		Number:      nextNumber,
		Date:        invoiceDate,
		DueDate:     dueDate,
		ClientID:    models.ClientID(client.ID),
		Description: description,
	}

	invoice, err := invoiceService.CreateInvoice(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	// Display success message
	fmt.Printf("✅ Invoice created successfully!\n")
	fmt.Printf("   Invoice Number: %s\n", invoice.Number)
	fmt.Printf("   Client: %s\n", client.Name)
	fmt.Printf("   Date: %s\n", invoice.Date.Format("2006-01-02"))
	fmt.Printf("   Due Date: %s\n", invoice.DueDate.Format("2006-01-02"))
	fmt.Printf("   Status: %s\n", invoice.Status)
	fmt.Printf("\n")
	fmt.Printf("💡 Next steps:\n")
	fmt.Printf("   • Import work items: go-invoice import --file hours.csv --invoice %s\n", invoice.ID)
	fmt.Printf("   • Generate invoice: go-invoice generate %s\n", invoice.ID)

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
	cmd.Flags().String("status", "", "Filter by status (draft, sent, paid, overdue, cancelled)")
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
func (a *App) runInvoiceList(cmd *cobra.Command, args []string) error {
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
	filter, err := a.buildInvoiceFilter(cmd, clientService, ctx)
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
		return a.outputInvoicesCSV(invoices)
	default:
		if err := a.outputInvoicesTable(invoices, clientService, ctx); err != nil {
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

	// Get invoice
	invoice, err := invoiceService.GetInvoice(ctx, models.InvoiceID(invoiceID))
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	// Get client
	client, err := clientService.GetClient(ctx, models.ClientID(invoice.Client.ID))
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
		fmt.Println(string(data))
	case "yaml":
		// For now, we'll use JSON format for YAML output
		// In a real implementation, we would use a YAML library
		data, err := json.MarshalIndent(invoice, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal invoice: %w", err)
		}
		fmt.Println(string(data))
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
		
Note: You cannot update invoices that are already paid or cancelled.
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
	cmd.Flags().String("status", "", "Update status (draft, sent, paid, overdue, cancelled)")
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

	// Get current invoice
	invoice, err := invoiceService.GetInvoice(ctx, models.InvoiceID(invoiceID))
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	// Check if invoice can be updated
	if invoice.Status == models.StatusPaid || invoice.Status == models.StatusVoided {
		return fmt.Errorf("cannot update invoice with status: %s", invoice.Status)
	}

	// Get flags
	status, _ := cmd.Flags().GetString("status")
	dueDateStr, _ := cmd.Flags().GetString("due-date")
	description, _ := cmd.Flags().GetString("description")
	notes, _ := cmd.Flags().GetString("notes")
	interactive, _ := cmd.Flags().GetBool("interactive")

	// Interactive mode
	if interactive {
		return a.runInvoiceUpdateInteractive(ctx, invoiceService, invoice)
	}

	// Build update request
	req := models.UpdateInvoiceRequest{
		ID: models.InvoiceID(invoiceID),
	}

	hasUpdates := false

	// Update status
	if status != "" {
		// Validate status is one of the allowed values
		validStatuses := []string{"draft", "sent", "paid", "overdue", "voided"}
		validStatus := false
		for _, vs := range validStatuses {
			if status == vs {
				validStatus = true
				break
			}
		}
		if !validStatus {
			return fmt.Errorf("invalid status: %s (must be one of: %s)", status, strings.Join(validStatuses, ", "))
		}
		req.Status = &status
		hasUpdates = true
	}

	// Update due date
	if dueDateStr != "" {
		dueDate, err := time.Parse("2006-01-02", dueDateStr)
		if err != nil {
			return fmt.Errorf("invalid due date format (use YYYY-MM-DD): %w", err)
		}
		req.DueDate = &dueDate
		hasUpdates = true
	}

	// Update description
	if description != "" {
		req.Description = &description
		hasUpdates = true
	}

	// Notes field doesn't exist in UpdateInvoiceRequest
	// TODO: Add notes support when field is added to model
	if notes != "" {
		// For now, we'll ignore notes updates
		a.logger.Debug("notes update not yet supported", "notes", notes)
	}

	if !hasUpdates {
		return fmt.Errorf("no updates specified")
	}

	// Perform update
	updatedInvoice, err := invoiceService.UpdateInvoice(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	// Display success message
	fmt.Printf("✅ Invoice updated successfully!\n")
	fmt.Printf("   Invoice Number: %s\n", updatedInvoice.Number)
	if req.Status != nil {
		fmt.Printf("   Status: %s → %s\n", invoice.Status, updatedInvoice.Status)
	}
	if req.DueDate != nil {
		fmt.Printf("   Due Date: %s → %s\n", invoice.DueDate.Format("2006-01-02"), updatedInvoice.DueDate.Format("2006-01-02"))
	}
	if req.Description != nil {
		fmt.Printf("   Description updated\n")
	}
	// Notes not yet supported in UpdateInvoiceRequest

	return nil
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

	// Get invoice to verify it exists and check status
	invoice, err := invoiceService.GetInvoice(ctx, models.InvoiceID(invoiceID))
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	// Check if invoice can be deleted
	if invoice.Status == models.StatusPaid {
		return fmt.Errorf("cannot delete paid invoice")
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

		fmt.Printf("⚠️  About to %s invoice %s\n", deleteType, invoice.Number)
		fmt.Printf("   Client: %s\n", invoice.Client.Name)
		fmt.Printf("   Date: %s\n", invoice.Date.Format("2006-01-02"))
		fmt.Printf("   Total: %.2f\n", invoice.Total)
		fmt.Printf("\n")

		if hardDelete {
			fmt.Printf("❗ This action CANNOT be undone!\n")
		}

		fmt.Printf("Are you sure you want to continue? (yes/no): ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "yes" {
			fmt.Println("❌ Delete cancelled")
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
		fmt.Printf("✅ Invoice %s permanently deleted\n", invoice.Number)
	} else {
		err = invoiceService.DeleteInvoice(ctx, models.InvoiceID(invoiceID))
		if err != nil {
			return fmt.Errorf("failed to delete invoice: %w", err)
		}
		fmt.Printf("✅ Invoice %s deleted\n", invoice.Number)
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
		fmt.Printf("Multiple clients found matching '%s':\n", clientName)
		for i, client := range clients {
			fmt.Printf("  %d. %s (%s)\n", i+1, client.Name, client.Email)
		}
		return nil, fmt.Errorf("please specify a more specific client name")
	}

	// No client found
	if !createIfMissing {
		return nil, fmt.Errorf("client '%s' not found (use --create-client to create)", clientName)
	}

	// Create new client
	email, _ := cmd.Flags().GetString("email")
	if email == "" {
		return nil, fmt.Errorf("email is required when creating a new client")
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

	fmt.Printf("✅ Created new client: %s\n", client.Name)
	return client, nil
}

// buildInvoiceFilter builds an invoice filter from command flags
func (a *App) buildInvoiceFilter(cmd *cobra.Command, clientService *services.ClientService, ctx context.Context) (models.InvoiceFilter, error) {
	filter := models.InvoiceFilter{}

	// Status filter
	if status, _ := cmd.Flags().GetString("status"); status != "" {
		// Validate status
		validStatuses := []string{"draft", "sent", "paid", "overdue", "voided"}
		valid := false
		for _, vs := range validStatuses {
			if status == vs {
				valid = true
				break
			}
		}
		if !valid {
			return filter, fmt.Errorf("invalid status: %s", status)
		}
		filter.Status = status
	}

	// Client filter
	if clientName, _ := cmd.Flags().GetString("client"); clientName != "" {
		clients, err := a.searchClientsByName(ctx, clientService, clientName)
		if err != nil {
			return filter, fmt.Errorf("failed to search for client: %w", err)
		}
		if len(clients) == 0 {
			return filter, fmt.Errorf("no clients found matching: %s", clientName)
		}
		if len(clients) > 1 {
			return filter, fmt.Errorf("multiple clients found matching: %s", clientName)
		}
		filter.ClientID = models.ClientID(clients[0].ID)
	}

	// Date range filter
	if fromStr, _ := cmd.Flags().GetString("from"); fromStr != "" {
		fromDate, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			return filter, fmt.Errorf("invalid from date format (use YYYY-MM-DD): %w", err)
		}
		filter.DateFrom = fromDate
	}

	if toStr, _ := cmd.Flags().GetString("to"); toStr != "" {
		toDate, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			return filter, fmt.Errorf("invalid to date format (use YYYY-MM-DD): %w", err)
		}
		// Set to end of day
		toDate = toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filter.DateTo = toDate
	}

	// Limit
	limit, _ := cmd.Flags().GetInt("limit")
	if limit > 0 {
		filter.Limit = limit
	}

	return filter, nil
}

// Output formatting methods

func (a *App) outputInvoicesJSON(invoices []*models.Invoice) error {
	data, err := json.MarshalIndent(invoices, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal invoices: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func (a *App) outputInvoicesCSV(invoices []*models.Invoice) error {
	// CSV header
	fmt.Println("Number,Date,DueDate,ClientName,Status,SubTotal,Tax,Total")

	// CSV rows
	for _, inv := range invoices {
		fmt.Printf("%s,%s,%s,%s,%s,%.2f,%.2f,%.2f\n",
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
	return nil
}

func (a *App) outputInvoicesTable(invoices []*models.Invoice, clientService *services.ClientService, ctx context.Context) error {
	if len(invoices) == 0 {
		fmt.Println("No invoices found")
		return nil
	}

	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Header
	fmt.Fprintln(w, "NUMBER\tCLIENT\tDATE\tDUE DATE\tSTATUS\tAMOUNT")
	fmt.Fprintln(w, "------\t------\t----\t--------\t------\t------")

	// Rows
	for _, inv := range invoices {
		// Client name is already in invoice
		clientName := inv.Client.Name

		// Format status with color (in a real terminal)
		status := string(inv.Status)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%.2f\n",
			inv.Number,
			clientName,
			inv.Date.Format("2006-01-02"),
			inv.DueDate.Format("2006-01-02"),
			status,
			inv.Total,
		)
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

	fmt.Printf("\n📊 Summary\n")
	fmt.Printf("─────────\n")
	fmt.Printf("Total Invoices: %d\n", len(invoices))
	fmt.Printf("  Draft: %d\n", draftCount)
	fmt.Printf("  Sent: %d\n", sentCount)
	fmt.Printf("  Paid: %d\n", paidCount)
	fmt.Printf("  Overdue: %d\n", overdueCount)
	fmt.Printf("\n")
	fmt.Printf("Total Amount: %.2f %s\n", totalAmount, currency)
	fmt.Printf("  Paid: %.2f %s\n", paidAmount, currency)
	fmt.Printf("  Unpaid: %.2f %s\n", unpaidAmount, currency)
}

func (a *App) displayInvoiceDetails(invoice *models.Invoice, client *models.Client, currency string, showItems, showHistory bool) {
	fmt.Printf("📄 Invoice %s\n", invoice.Number)
	fmt.Printf("════════════════════\n")
	fmt.Printf("\n")

	fmt.Printf("Client: %s\n", client.Name)
	fmt.Printf("Email: %s\n", client.Email)
	if client.Address != "" {
		fmt.Printf("Address: %s\n", client.Address)
	}
	fmt.Printf("\n")

	fmt.Printf("Date: %s\n", invoice.Date.Format("2006-01-02"))
	fmt.Printf("Due Date: %s\n", invoice.DueDate.Format("2006-01-02"))
	fmt.Printf("Status: %s\n", invoice.Status)

	if invoice.Description != "" {
		fmt.Printf("Description: %s\n", invoice.Description)
	}

	fmt.Printf("\n")
	fmt.Printf("💰 Financial Summary\n")
	fmt.Printf("──────────────────\n")
	fmt.Printf("Subtotal: %.2f %s\n", invoice.Subtotal, currency)
	if invoice.TaxAmount > 0 {
		fmt.Printf("Tax: %.2f %s\n", invoice.TaxAmount, currency)
	}
	fmt.Printf("Total: %.2f %s\n", invoice.Total, currency)

	if showItems && len(invoice.WorkItems) > 0 {
		fmt.Printf("\n")
		fmt.Printf("📋 Work Items\n")
		fmt.Printf("───────────\n")

		for i, item := range invoice.WorkItems {
			fmt.Printf("\n%d. %s\n", i+1, item.Description)
			fmt.Printf("   Date: %s\n", item.Date.Format("2006-01-02"))
			fmt.Printf("   Hours: %.2f @ %.2f/hour = %.2f %s\n",
				item.Hours, item.Rate, item.Total, currency)
		}
	}

	// Notes field not yet available in Invoice model

	fmt.Printf("\n")
	fmt.Printf("🕒 Timestamps\n")
	fmt.Printf("───────────\n")
	fmt.Printf("Created: %s\n", invoice.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", invoice.UpdatedAt.Format("2006-01-02 15:04:05"))
}

// Interactive mode helpers

func (a *App) runInvoiceCreateInteractive(ctx context.Context, invoiceService *services.InvoiceService, clientService *services.ClientService, config *config.Config) error {
	fmt.Println("🔨 Create New Invoice - Interactive Mode")
	fmt.Println("=====================================")
	fmt.Println()

	prompter := cli.NewPrompter(a.logger)

	// Prompt for client
	fmt.Println("Step 1: Select or create client")
	fmt.Println("-------------------------------")

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

		index, _, err := prompter.PromptSelect(ctx, "Select a client:", options, -1)
		if err != nil {
			return fmt.Errorf("client selection cancelled: %w", err)
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
		fmt.Println("No clients found. Let's create one.")
		client, err = a.createClientInteractive(ctx, clientService, prompter)
		if err != nil {
			return err
		}
	}

	fmt.Printf("\n✅ Client selected: %s\n\n", client.Name)

	// Prompt for invoice details
	fmt.Println("Step 2: Invoice details")
	fmt.Println("----------------------")

	// Invoice date
	invoiceDate, err := prompter.PromptDate(ctx, "Invoice date", time.Now())
	if err != nil {
		return fmt.Errorf("date selection cancelled: %w", err)
	}

	// Due date
	defaultDueDate := invoiceDate.AddDate(0, 0, config.Invoice.DefaultDueDays)
	dueDate, err := prompter.PromptDate(ctx, "Due date", defaultDueDate)
	if err != nil {
		return fmt.Errorf("due date selection cancelled: %w", err)
	}

	// Description
	description, err := prompter.PromptString(ctx, "Invoice description (optional)", "")
	if err != nil {
		return fmt.Errorf("description input cancelled: %w", err)
	}

	// Generate invoice number
	nextNumber, err := a.generateNextInvoiceNumber(ctx, invoiceService, config.Invoice.Prefix, config.Invoice.StartNumber)
	if err != nil {
		return fmt.Errorf("failed to generate invoice number: %w", err)
	}

	fmt.Printf("\n📋 Invoice Summary:\n")
	fmt.Printf("   Number: %s\n", nextNumber)
	fmt.Printf("   Client: %s\n", client.Name)
	fmt.Printf("   Date: %s\n", invoiceDate.Format("2006-01-02"))
	fmt.Printf("   Due Date: %s\n", dueDate.Format("2006-01-02"))
	if description != "" {
		fmt.Printf("   Description: %s\n", description)
	}
	fmt.Printf("\n")

	// Confirm creation
	confirmed, err := prompter.PromptConfirm(ctx, "Create this invoice?")
	if err != nil {
		return fmt.Errorf("confirmation cancelled: %w", err)
	}

	if !confirmed {
		fmt.Println("❌ Invoice creation cancelled")
		return nil
	}

	// Create invoice
	req := models.CreateInvoiceRequest{
		Number:      nextNumber,
		Date:        invoiceDate,
		DueDate:     dueDate,
		ClientID:    models.ClientID(client.ID),
		Description: description,
	}

	invoice, err := invoiceService.CreateInvoice(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	// Display success message
	fmt.Printf("\n✅ Invoice created successfully!\n")
	fmt.Printf("   Invoice Number: %s\n", invoice.Number)
	fmt.Printf("   Invoice ID: %s\n", invoice.ID)
	fmt.Printf("\n")
	fmt.Printf("💡 Next steps:\n")
	fmt.Printf("   • Import work items: go-invoice import --file hours.csv --invoice %s\n", invoice.ID)
	fmt.Printf("   • Generate invoice: go-invoice generate %s\n", invoice.ID)

	return nil
}

func (a *App) runInvoiceUpdateInteractive(ctx context.Context, invoiceService *services.InvoiceService, invoice *models.Invoice) error {
	fmt.Printf("🔧 Update Invoice %s - Interactive Mode\n", invoice.Number)
	fmt.Println("=====================================")
	fmt.Println()

	prompter := cli.NewPrompter(a.logger)

	// Show current invoice details
	fmt.Println("Current Invoice Details:")
	fmt.Printf("   Status: %s\n", invoice.Status)
	fmt.Printf("   Date: %s\n", invoice.Date.Format("2006-01-02"))
	fmt.Printf("   Due Date: %s\n", invoice.DueDate.Format("2006-01-02"))
	fmt.Printf("   Description: %s\n", invoice.Description)
	fmt.Println()

	// Select what to update
	options := []string{
		"Update status",
		"Update due date",
		"Update description",
		"Cancel (no changes)",
	}

	index, _, err := prompter.PromptSelect(ctx, "What would you like to update?", options, -1)
	if err != nil {
		return fmt.Errorf("selection cancelled: %w", err)
	}

	if index == 3 {
		fmt.Println("❌ Update cancelled")
		return nil
	}

	req := models.UpdateInvoiceRequest{
		ID: models.InvoiceID(invoice.ID),
	}

	switch index {
	case 0: // Update status
		statuses := []string{"draft", "sent", "paid", "overdue", "voided"}
		statusIndex, newStatus, err := prompter.PromptSelect(ctx, "Select new status:", statuses, -1)
		if err != nil {
			return fmt.Errorf("status selection cancelled: %w", err)
		}
		_ = statusIndex // unused
		req.Status = &newStatus

	case 1: // Update due date
		newDueDate, err := prompter.PromptDate(ctx, "New due date", invoice.DueDate)
		if err != nil {
			return fmt.Errorf("due date selection cancelled: %w", err)
		}
		req.DueDate = &newDueDate

	case 2: // Update description
		newDescription, err := prompter.PromptString(ctx, "New description", invoice.Description)
		if err != nil {
			return fmt.Errorf("description input cancelled: %w", err)
		}
		req.Description = &newDescription
	}

	// Confirm update
	fmt.Println("\n📋 Update Summary:")
	if req.Status != nil {
		fmt.Printf("   Status: %s → %s\n", invoice.Status, *req.Status)
	}
	if req.DueDate != nil {
		fmt.Printf("   Due Date: %s → %s\n", invoice.DueDate.Format("2006-01-02"), req.DueDate.Format("2006-01-02"))
	}
	if req.Description != nil {
		fmt.Printf("   Description: %s → %s\n", invoice.Description, *req.Description)
	}
	fmt.Println()

	confirmed, err := prompter.PromptConfirm(ctx, "Apply these changes?")
	if err != nil {
		return fmt.Errorf("confirmation cancelled: %w", err)
	}

	if !confirmed {
		fmt.Println("❌ Update cancelled")
		return nil
	}

	// Perform update
	updatedInvoice, err := invoiceService.UpdateInvoice(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	fmt.Printf("\n✅ Invoice %s updated successfully!\n", updatedInvoice.Number)
	return nil
}

// createClientInteractive creates a new client through interactive prompts
func (a *App) createClientInteractive(ctx context.Context, clientService *services.ClientService, prompter *cli.Prompter) (*models.Client, error) {
	fmt.Println("\n📝 Create New Client")
	fmt.Println("-------------------")

	// Prompt for client details
	name, err := prompter.PromptStringRequired(ctx, "Client name")
	if err != nil {
		return nil, fmt.Errorf("name input cancelled: %w", err)
	}

	email, err := prompter.PromptStringRequired(ctx, "Client email")
	if err != nil {
		return nil, fmt.Errorf("email input cancelled: %w", err)
	}

	phone, err := prompter.PromptString(ctx, "Client phone (optional)", "")
	if err != nil {
		return nil, fmt.Errorf("phone input cancelled: %w", err)
	}

	address, err := prompter.PromptString(ctx, "Client address (optional)", "")
	if err != nil {
		return nil, fmt.Errorf("address input cancelled: %w", err)
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

	fmt.Printf("✅ Client '%s' created successfully!\n", client.Name)
	return client, nil
}

// generateNextInvoiceNumber generates the next available invoice number
func (a *App) generateNextInvoiceNumber(ctx context.Context, invoiceService *services.InvoiceService, prefix string, startNumber int) (string, error) {
	// Get all invoices to find the highest number
	filter := models.InvoiceFilter{}
	result, err := invoiceService.ListInvoices(ctx, filter)
	if err != nil {
		return "", fmt.Errorf("failed to list invoices: %w", err)
	}

	// Find the highest invoice number with the same prefix
	highestNum := startNumber - 1
	for _, inv := range result.Invoices {
		if strings.HasPrefix(inv.Number, prefix) {
			// Extract the numeric part
			numPart := strings.TrimPrefix(inv.Number, prefix)
			numPart = strings.TrimPrefix(numPart, "-")

			if num, err := strconv.Atoi(numPart); err == nil && num > highestNum {
				highestNum = num
			}
		}
	}

	// Generate next number
	nextNum := highestNum + 1
	return fmt.Sprintf("%s-%03d", prefix, nextNum), nil
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
