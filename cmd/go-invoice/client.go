package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/mrz1836/go-invoice/internal/models"
	"github.com/mrz1836/go-invoice/internal/services"

	// removed unused imports
	"github.com/spf13/cobra"
)

// buildClientCommand creates the client command with all subcommands
func (a *App) buildClientCommand() *cobra.Command {
	clientCmd := &cobra.Command{
		Use:   "client",
		Short: "Client management commands",
		Long:  "Create, list, show, update, and delete clients",
	}

	// Add client subcommands
	clientCmd.AddCommand(a.buildClientCreateCommand())
	clientCmd.AddCommand(a.buildClientListCommand())
	clientCmd.AddCommand(a.buildClientShowCommand())
	clientCmd.AddCommand(a.buildClientUpdateCommand())
	clientCmd.AddCommand(a.buildClientDeleteCommand())

	return clientCmd
}

// buildClientCreateCommand creates the client create command
func (a *App) buildClientCreateCommand() *cobra.Command {
	var name, email, phone, address, taxID string
	var cryptoFeeEnabled bool
	var cryptoFeeAmount float64
	var lateFeeEnabled bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new client",
		Long:  "Create a new client with contact information",
		Example: `  go-invoice client create --name "Acme Corp" --email "contact@acme.com"
  go-invoice client create --name "John Smith" --email "john@example.com" --phone "+1-555-123-4567"
  go-invoice client create --name "Acme Company" --email "billing@acme.com" --crypto-fee --crypto-fee-amount 25.00 --late-fee`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Load configuration
			configPath, _ := cmd.Flags().GetString("config")
			config, err := a.configService.LoadConfig(ctx, configPath)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Create storage and services
			invoiceStorage, clientStorage := a.createStorageInstances(config.Storage.DataDir)
			idGen := services.NewUUIDGenerator()
			clientService := services.NewClientService(clientStorage, invoiceStorage, a.logger, idGen)

			// Create client request
			req := models.CreateClientRequest{
				Name:             name,
				Email:            email,
				Phone:            phone,
				Address:          address,
				TaxID:            taxID,
				CryptoFeeEnabled: cryptoFeeEnabled,
				CryptoFeeAmount:  cryptoFeeAmount,
				LateFeeEnabled:   lateFeeEnabled,
			}

			client, err := clientService.CreateClient(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			a.logger.Info("Client created successfully", "name", client.Name, "id", client.ID)
			if cryptoFeeEnabled {
				a.logger.Printf("💰 Crypto service fee enabled: $%.2f\n", cryptoFeeAmount)
			}
			if lateFeeEnabled {
				a.logger.Printf("⚠️  Late fee policy enabled (1.5%% per month / 18%% APR)\n")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Client name (required)")
	cmd.Flags().StringVar(&email, "email", "", "Client email (required)")
	cmd.Flags().StringVar(&phone, "phone", "", "Client phone number")
	cmd.Flags().StringVar(&address, "address", "", "Client address")
	cmd.Flags().StringVar(&taxID, "tax-id", "", "Tax ID (EIN, VAT number, etc.)")
	cmd.Flags().BoolVar(&cryptoFeeEnabled, "crypto-fee", false, "Enable cryptocurrency service fee for this client")
	cmd.Flags().Float64Var(&cryptoFeeAmount, "crypto-fee-amount", 25.00, "Cryptocurrency service fee amount")
	cmd.Flags().BoolVar(&lateFeeEnabled, "late-fee", true, "Enable late fee policy on invoices (default: true)")

	if err := cmd.MarkFlagRequired("name"); err != nil {
		return cmd
	}
	if err := cmd.MarkFlagRequired("email"); err != nil {
		return cmd
	}

	return cmd
}

// buildClientListCommand creates the client list command
func (a *App) buildClientListCommand() *cobra.Command {
	var outputFormat string
	var activeOnly, inactiveOnly bool
	var search string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List clients",
		Long:  "List all clients with filtering options",
		Example: `  go-invoice client list
  go-invoice client list --search "Acme"
  go-invoice client list --inactive --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Load configuration
			configPath, _ := cmd.Flags().GetString("config")
			config, err := a.configService.LoadConfig(ctx, configPath)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Create storage and services
			invoiceStorage, clientStorage := a.createStorageInstances(config.Storage.DataDir)
			_ = invoiceStorage // unused in this command

			// Determine active filter
			activeFilter := !inactiveOnly

			// List clients
			result, err := clientStorage.ListClients(ctx, activeFilter, limit, 0)
			if err != nil {
				return fmt.Errorf("failed to list clients: %w", err)
			}

			// Filter by search if provided
			filteredClients := result.Clients
			if search != "" {
				var filtered []*models.Client
				searchLower := strings.ToLower(search)
				for _, client := range result.Clients {
					if strings.Contains(strings.ToLower(client.Name), searchLower) ||
						strings.Contains(strings.ToLower(client.Email), searchLower) {
						filtered = append(filtered, client)
					}
				}
				filteredClients = filtered
			}

			// Output results
			switch outputFormat {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(filteredClients)

			default:
				if len(filteredClients) == 0 {
					a.logger.Info("No clients found")
					return nil
				}

				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				if _, err := fmt.Fprintln(w, "ID\tNAME\tEMAIL\tPHONE\tSTATUS"); err != nil {
					return fmt.Errorf("failed to write header: %w", err)
				}
				if _, err := fmt.Fprintln(w, "---\t----\t-----\t-----\t------"); err != nil {
					return fmt.Errorf("failed to write separator: %w", err)
				}

				for _, client := range filteredClients {
					status := "Active"
					if !client.Active {
						status = "Inactive"
					}
					if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
						client.ID, client.Name, client.Email, client.Phone, status); err != nil {
						return fmt.Errorf("failed to write client data: %w", err)
					}
				}
				if err := w.Flush(); err != nil {
					return fmt.Errorf("failed to flush output: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&outputFormat, "output", "table", "Output format (table, json)")
	cmd.Flags().BoolVar(&activeOnly, "active", false, "Show only active clients")
	cmd.Flags().BoolVar(&inactiveOnly, "inactive", false, "Show only inactive clients")
	cmd.Flags().StringVar(&search, "search", "", "Search clients by name or email")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of clients to return")

	return cmd
}

// buildClientShowCommand creates the client show command
func (a *App) buildClientShowCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "show [client-id or name]",
		Short: "Show client details",
		Long:  "Display detailed information about a specific client",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Load configuration
			configPath, _ := cmd.Flags().GetString("config")
			config, err := a.configService.LoadConfig(ctx, configPath)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Create storage and services
			invoiceStorage, clientStorage := a.createStorageInstances(config.Storage.DataDir)

			// Try to find by ID first
			client, err := clientStorage.GetClient(ctx, models.ClientID(args[0]))
			if err != nil {
				// Try to find by name
				listResult, listErr := clientStorage.ListClients(ctx, true, 100, 0)
				if listErr != nil {
					return fmt.Errorf("failed to search clients: %w", listErr)
				}

				var matches []*models.Client
				searchLower := strings.ToLower(args[0])
				for _, c := range listResult.Clients {
					if strings.Contains(strings.ToLower(c.Name), searchLower) {
						matches = append(matches, c)
					}
				}

				if len(matches) == 0 {
					return fmt.Errorf("%w: %s", models.ErrClientNotFound, args[0])
				}
				if len(matches) > 1 {
					return fmt.Errorf("%w matching '%s'", models.ErrMultipleClientsFound, args[0])
				}
				client = matches[0]
			}

			// Get invoice statistics
			idGen := services.NewUUIDGenerator()
			invoiceService := services.NewInvoiceService(invoiceStorage, clientStorage, a.logger, idGen)
			filter := models.InvoiceFilter{}
			result, err := invoiceService.ListInvoices(ctx, filter)
			if err != nil {
				return fmt.Errorf("failed to list invoices: %w", err)
			}

			// Filter invoices for this client
			var clientInvoices []*models.Invoice
			for _, inv := range result.Invoices {
				if inv.Client.ID == client.ID {
					clientInvoices = append(clientInvoices, inv)
				}
			}

			// Output results
			switch outputFormat {
			case "json":
				output := map[string]interface{}{
					"client":        client,
					"invoice_count": len(clientInvoices),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(output)

			default:
				// Output client details to console
				if _, err := fmt.Fprintf(os.Stdout, "Client Details:\n"); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
				if _, err := fmt.Fprintf(os.Stdout, "  ID:       %s\n", client.ID); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
				if _, err := fmt.Fprintf(os.Stdout, "  Name:     %s\n", client.Name); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
				if _, err := fmt.Fprintf(os.Stdout, "  Email:    %s\n", client.Email); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
				if _, err := fmt.Fprintf(os.Stdout, "  Phone:    %s\n", client.Phone); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
				if _, err := fmt.Fprintf(os.Stdout, "  Address:  %s\n", client.Address); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
				if _, err := fmt.Fprintf(os.Stdout, "  Tax ID:   %s\n", client.TaxID); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
				status := "Active"
				if !client.Active {
					status = "Inactive"
				}
				if _, err := fmt.Fprintf(os.Stdout, "  Status:   %s\n", status); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
				if _, err := fmt.Fprintf(os.Stdout, "  Created:  %s\n", client.CreatedAt.Format(time.RFC3339)); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
				if _, err := fmt.Fprintf(os.Stdout, "\nInvoice Summary:\n"); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
				if _, err := fmt.Fprintf(os.Stdout, "  Total Invoices: %d\n", len(clientInvoices)); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&outputFormat, "output", "table", "Output format (table, json)")

	return cmd
}

// buildClientUpdateCommand creates the client update command
func (a *App) buildClientUpdateCommand() *cobra.Command {
	var name, email, phone, address, taxID string
	var activate, deactivate bool
	var cryptoFeeEnabled bool
	var cryptoFeeAmount float64
	var lateFeeEnabled bool

	cmd := &cobra.Command{
		Use:   "update [client-id or name]",
		Short: "Update client information",
		Long:  "Update client contact information and status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Load configuration
			configPath, _ := cmd.Flags().GetString("config")
			config, err := a.configService.LoadConfig(ctx, configPath)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Create storage and services
			invoiceStorage, clientStorage := a.createStorageInstances(config.Storage.DataDir)
			idGen := services.NewUUIDGenerator()
			clientService := services.NewClientService(clientStorage, invoiceStorage, a.logger, idGen)

			// Find client
			client, err := clientStorage.GetClient(ctx, models.ClientID(args[0]))
			if err != nil {
				// Try to find by name
				listResult, listErr := clientStorage.ListClients(ctx, true, 100, 0)
				if listErr != nil {
					return fmt.Errorf("failed to search clients: %w", listErr)
				}

				var matches []*models.Client
				searchLower := strings.ToLower(args[0])
				for _, c := range listResult.Clients {
					if strings.Contains(strings.ToLower(c.Name), searchLower) {
						matches = append(matches, c)
					}
				}

				if len(matches) == 0 {
					return fmt.Errorf("%w: %s", models.ErrClientNotFound, args[0])
				}
				if len(matches) > 1 {
					return fmt.Errorf("%w matching '%s'", models.ErrMultipleClientsFound, args[0])
				}
				client = matches[0]
			}

			// Update fields
			updated := false
			if cmd.Flags().Changed("name") {
				client.Name = name
				updated = true
			}
			if cmd.Flags().Changed("email") {
				client.Email = email
				updated = true
			}
			if cmd.Flags().Changed("phone") {
				client.Phone = phone
				updated = true
			}
			if cmd.Flags().Changed("address") {
				client.Address = address
				updated = true
			}
			if cmd.Flags().Changed("tax-id") {
				client.TaxID = taxID
				updated = true
			}
			if activate && deactivate {
				return models.ErrCannotActivateDeactivate
			}
			if activate {
				client.Active = true
				updated = true
			}
			if deactivate {
				client.Active = false
				updated = true
			}
			if cmd.Flags().Changed("crypto-fee") {
				client.CryptoFeeEnabled = cryptoFeeEnabled
				updated = true
			}
			if cmd.Flags().Changed("crypto-fee-amount") {
				client.CryptoFeeAmount = cryptoFeeAmount
				updated = true
			}
			if cmd.Flags().Changed("late-fee") {
				client.LateFeeEnabled = lateFeeEnabled
				updated = true
			}

			if !updated {
				return models.ErrNoUpdatesSpecified
			}

			// Update client
			_, err = clientService.UpdateClient(ctx, client)
			if err != nil {
				return fmt.Errorf("failed to update client: %w", err)
			}

			a.logger.Info("Client updated successfully", "name", client.Name)
			if client.CryptoFeeEnabled {
				a.logger.Printf("💰 Crypto service fee: $%.2f\n", client.CryptoFeeAmount)
			}
			if client.LateFeeEnabled {
				a.logger.Printf("⚠️  Late fee policy enabled (1.5%% per month / 18%% APR)\n")
			} else {
				a.logger.Printf("ℹ️  Late fee policy disabled for this client\n")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Update client name")
	cmd.Flags().StringVar(&email, "email", "", "Update client email")
	cmd.Flags().StringVar(&phone, "phone", "", "Update client phone")
	cmd.Flags().StringVar(&address, "address", "", "Update client address")
	cmd.Flags().StringVar(&taxID, "tax-id", "", "Update tax ID")
	cmd.Flags().BoolVar(&activate, "activate", false, "Activate client")
	cmd.Flags().BoolVar(&deactivate, "deactivate", false, "Deactivate client")
	cmd.Flags().BoolVar(&cryptoFeeEnabled, "crypto-fee", false, "Enable cryptocurrency service fee for this client")
	cmd.Flags().Float64Var(&cryptoFeeAmount, "crypto-fee-amount", 25.00, "Cryptocurrency service fee amount")
	cmd.Flags().BoolVar(&lateFeeEnabled, "late-fee", true, "Enable late fee policy on invoices")

	return cmd
}

// buildClientDeleteCommand creates the client delete command
func (a *App) buildClientDeleteCommand() *cobra.Command {
	var force, hardDelete bool

	cmd := &cobra.Command{
		Use:   "delete [client-id or name]",
		Short: "Delete a client",
		Long:  "Delete or deactivate a client (soft delete by default)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Load configuration
			configPath, _ := cmd.Flags().GetString("config")
			config, err := a.configService.LoadConfig(ctx, configPath)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Create storage and services
			invoiceStorage, clientStorage := a.createStorageInstances(config.Storage.DataDir)
			idGen := services.NewUUIDGenerator()
			clientService := services.NewClientService(clientStorage, invoiceStorage, a.logger, idGen)

			// Find client
			client, err := clientStorage.GetClient(ctx, models.ClientID(args[0]))
			if err != nil {
				// Try to find by name
				listResult, listErr := clientStorage.ListClients(ctx, true, 100, 0)
				if listErr != nil {
					return fmt.Errorf("failed to search clients: %w", listErr)
				}

				var matches []*models.Client
				searchLower := strings.ToLower(args[0])
				for _, c := range listResult.Clients {
					if strings.Contains(strings.ToLower(c.Name), searchLower) {
						matches = append(matches, c)
					}
				}

				if len(matches) == 0 {
					return fmt.Errorf("%w: %s", models.ErrClientNotFound, args[0])
				}
				if len(matches) > 1 {
					return fmt.Errorf("%w matching '%s'", models.ErrMultipleClientsFound, args[0])
				}
				client = matches[0]
			}

			// Confirm deletion
			if !force {
				a.logger.Info("Deletion confirmation", "message", fmt.Sprintf("Are you sure you want to delete client '%s'? (y/N): ", client.Name))
				var response string
				if _, scanErr := fmt.Scanln(&response); scanErr != nil {
					return fmt.Errorf("failed to read response: %w", scanErr)
				}
				if strings.ToLower(response) != "y" {
					a.logger.Info("Deletion canceled")
					return nil
				}
			}

			// Delete client
			if hardDelete {
				if deleteErr := clientService.DeleteClient(ctx, client.ID); deleteErr != nil {
					return fmt.Errorf("failed to delete client: %w", deleteErr)
				}
				a.logger.Info("Client permanently deleted", "name", client.Name)
			} else {
				// Soft delete (deactivate)
				client.Active = false
				_, err = clientService.UpdateClient(ctx, client)
				if err != nil {
					return fmt.Errorf("failed to deactivate client: %w", err)
				}
				a.logger.Info("Client deactivated", "name", client.Name)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&hardDelete, "hard", false, "Permanently delete client (cannot be undone)")

	return cmd
}
