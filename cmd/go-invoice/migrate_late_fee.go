package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// buildMigrateLateFeeCommand creates a command to enable late fees for all existing clients
func (a *App) buildMigrateLateFeeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate-late-fee",
		Short: "Enable late fee policy for all existing clients",
		Long:  "Updates all existing clients to enable the late fee policy by default",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Load configuration
			configPath, _ := cmd.Flags().GetString("config")
			config, err := a.configService.LoadConfig(ctx, configPath)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Create storage instances
			_, clientStorage := a.createStorageInstances(config.Storage.DataDir)

			// Get all clients
			result, err := clientStorage.ListClients(ctx, false, 0, 0)
			if err != nil {
				return fmt.Errorf("failed to list clients: %w", err)
			}

			if len(result.Clients) == 0 {
				a.logger.Info("No clients found to migrate")
				return nil
			}

			// Update each client to enable late fees
			updated := 0
			skipped := 0
			for _, client := range result.Clients {
				if client.LateFeeEnabled {
					a.logger.Info("Client already has late fee enabled, skipping", "name", client.Name)
					skipped++
					continue
				}

				client.LateFeeEnabled = true
				if err := clientStorage.UpdateClient(ctx, client); err != nil {
					a.logger.Error("failed to update client", "name", client.Name, "error", err)
					continue
				}

				a.logger.Info("Enabled late fee policy for client", "name", client.Name)
				updated++
			}

			a.logger.Info("Migration complete",
				"total", len(result.Clients),
				"updated", updated,
				"skipped", skipped,
			)
			return nil
		},
	}

	return cmd
}
