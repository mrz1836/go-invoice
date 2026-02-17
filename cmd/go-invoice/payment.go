package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/mrz1836/go-invoice/internal/blockchain"
	"github.com/mrz1836/go-invoice/internal/config"
	"github.com/mrz1836/go-invoice/internal/models"
	"github.com/mrz1836/go-invoice/internal/services"
	"github.com/spf13/cobra"
)

var (
	// ErrUnsupportedPaymentMethod is returned when an unsupported payment method is specified
	ErrUnsupportedPaymentMethod = errors.New("unsupported payment method")
	// ErrNoUSDCAddress is returned when no USDC address is configured
	ErrNoUSDCAddress = errors.New("no USDC address configured")
	// ErrNoBSVAddress is returned when no BSV address is configured
	ErrNoBSVAddress = errors.New("no BSV address configured")
)

// buildPaymentCommand creates the payment command with subcommands
func (a *App) buildPaymentCommand() *cobra.Command {
	paymentCmd := &cobra.Command{
		Use:   "payment",
		Short: "Payment verification and management commands",
		Long:  "Verify cryptocurrency payments on-chain and manage invoice payment status",
	}

	// Add payment subcommands
	paymentCmd.AddCommand(a.buildPaymentVerifyCommand())

	return paymentCmd
}

// buildPaymentVerifyCommand creates the payment verify subcommand
func (a *App) buildPaymentVerifyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify [invoice-id-or-number]",
		Short: "Verify cryptocurrency payment for an invoice",
		Long: `Verify that payment has been received on-chain for an invoice.

This command checks the blockchain for payments to the invoice's cryptocurrency
address and automatically marks the invoice as paid if payment is verified.

Currently supports:
  • USDC (Ethereum stablecoin) - exact amount matching
  • BSV (Bitcoin SV) - COMING SOON

The command will:
1. Get the payment address for the invoice (per-invoice or global config)
2. Query the blockchain provider for the current balance
3. Verify payment amount matches invoice total
4. Optionally mark invoice as PAID if verified`,
		Example: `  # Verify USDC payment (default)
  go-invoice payment verify INV-001

  # Verify USDC payment on testnet
  go-invoice payment verify INV-001 --testnet

  # Check payment without updating invoice (dry run)
  go-invoice payment verify INV-001 --dry-run

  # Verify BSV payment (when implemented)
  go-invoice payment verify INV-001 --method BSV`,
		Args: cobra.ExactArgs(1),
		RunE: a.runPaymentVerify,
	}

	// Add flags
	cmd.Flags().String("method", "USDC", "Payment method to verify (USDC, BSV)")
	cmd.Flags().Bool("testnet", false, "Use testnet for blockchain queries")
	cmd.Flags().Bool("dry-run", false, "Check payment without updating invoice status")
	cmd.Flags().String("etherscan-api-key", "", "Etherscan API key (optional, for higher rate limits)")

	return cmd
}

// runPaymentVerify handles the payment verify command
func (a *App) runPaymentVerify(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	invoiceIdentifier := args[0]

	// Get flags
	paymentMethod, _ := cmd.Flags().GetString("method")
	testnet, _ := cmd.Flags().GetBool("testnet")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	etherscanAPIKey, _ := cmd.Flags().GetString("etherscan-api-key")

	// Load configuration
	configPath, _ := cmd.Flags().GetString("config")
	config, err := a.configService.LoadConfig(ctx, configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Use config API key as fallback if flag not provided
	if etherscanAPIKey == "" && config.Business.CryptoPayments.EtherscanAPIKey != "" {
		etherscanAPIKey = config.Business.CryptoPayments.EtherscanAPIKey
	}

	// Create storage and services
	invoiceStorage, _ := a.createStorageInstances(config.Storage.DataDir)
	idGen := services.NewUUIDGenerator()
	invoiceService := services.NewInvoiceService(invoiceStorage, nil, a.logger, idGen)
	paymentService := services.NewPaymentService(invoiceStorage, a.logger)

	// Get invoice
	invoice, err := a.getInvoiceByIDOrNumber(ctx, invoiceService, invoiceIdentifier)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	a.logger.Printf("🔍 Verifying payment for invoice %s\n", invoice.Number)
	a.logger.Printf("   Invoice Total: %.2f %s\n", invoice.Total, config.Invoice.Currency)
	a.logger.Printf("   Status: %s\n", invoice.Status)
	a.logger.Println("")

	// Create blockchain provider based on payment method
	provider, err := a.createBlockchainProvider(paymentMethod, testnet, etherscanAPIKey)
	if err != nil {
		return fmt.Errorf("failed to create blockchain provider: %w", err)
	}

	a.logger.Printf("   Provider: %s\n", provider.Name())

	// Determine payment method
	var method models.PaymentMethod
	switch paymentMethod {
	case "USDC":
		method = models.PaymentMethodUSDC
	case "BSV":
		method = models.PaymentMethodBSV
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedPaymentMethod, paymentMethod)
	}

	// Get payment address for display
	address, err := a.getInvoicePaymentAddress(invoice, config, method)
	if err != nil {
		return err
	}
	a.logger.Printf("   Payment Address: %s\n\n", address)

	// Verify payment
	verificationConfig := services.PaymentVerificationConfig{
		PaymentMethod:      method,
		DefaultUSDCAddress: config.Business.CryptoPayments.USDCAddress,
		DefaultBSVAddress:  config.Business.CryptoPayments.BSVAddress,
	}

	result, err := paymentService.VerifyPayment(ctx, invoice, provider, verificationConfig)
	if err != nil {
		a.logger.Printf("❌ Payment verification failed: %v\n", err)
		return err
	}

	// Display results
	a.displayPaymentVerificationResult(result, invoice, dryRun)

	// Mark invoice as paid if verification successful and not dry run
	if result.IsSuccessful() && !dryRun {
		if invoice.Status != models.StatusPaid {
			if err := paymentService.MarkInvoiceAsPaid(ctx, invoice.ID, result); err != nil {
				a.logger.Printf("⚠️  Warning: Failed to update invoice status: %v\n", err)
				return err
			}
			a.logger.Println("")
			a.logger.Println("🎉 Congratulations! Invoice has been marked as PAID!")
			a.logger.Println("")
		}
	}

	return nil
}

// displayPaymentVerificationResult displays the payment verification result
func (a *App) displayPaymentVerificationResult(result *models.PaymentVerification, invoice *models.Invoice, dryRun bool) {
	a.logger.Println("📊 Verification Results")
	a.logger.Println("═══════════════════════")
	a.logger.Println("")

	switch result.Status {
	case models.PaymentStatusVerified:
		a.logger.Println("✅ Payment VERIFIED")
		a.logger.Printf("   Amount Received: %.2f %s\n", result.ReceivedAmount, result.Currency)
		if result.TransactionHash != "" {
			a.logger.Printf("   Transaction: %s\n", result.TransactionHash)
		}
		if result.ConfirmedAt != nil {
			a.logger.Printf("   Confirmed: %s\n", result.ConfirmedAt.Format("2006-01-02 15:04:05"))
		}

	case models.PaymentStatusOverpaid:
		a.logger.Println("✅ Payment VERIFIED (Overpaid)")
		overpayment := result.ReceivedAmount - result.ExpectedAmount
		a.logger.Printf("   Amount Received: %.2f %s (%.2f over)\n",
			result.ReceivedAmount, result.Currency, overpayment)
		if result.TransactionHash != "" {
			a.logger.Printf("   Transaction: %s\n", result.TransactionHash)
		}

	case models.PaymentStatusNotFound:
		a.logger.Println("❌ Payment NOT FOUND")
		a.logger.Printf("   Expected Amount: %.2f %s\n", result.ExpectedAmount, result.Currency)
		a.logger.Printf("   Current Balance: %.2f %s\n", result.ReceivedAmount, result.Currency)
		a.logger.Println("")
		a.logger.Println("💡 Payment Instructions:")
		a.logger.Printf("   Send exactly %.2f %s to:\n", invoice.Total, result.Currency)
		a.logger.Printf("   %s\n", result.WalletAddress)

	case models.PaymentStatusPartial:
		a.logger.Println("⚠️  PARTIAL Payment Detected")
		remaining := result.ExpectedAmount - result.ReceivedAmount
		a.logger.Printf("   Amount Received: %.2f %s\n", result.ReceivedAmount, result.Currency)
		a.logger.Printf("   Amount Required: %.2f %s\n", result.ExpectedAmount, result.Currency)
		a.logger.Printf("   Remaining: %.2f %s\n", remaining, result.Currency)
		a.logger.Println("")
		a.logger.Printf("💡 Please send an additional %.2f %s to complete payment\n", remaining, result.Currency)

	case models.PaymentStatusPending:
		a.logger.Println("⏳ Payment PENDING Confirmation")
		a.logger.Printf("   Amount: %.2f %s\n", result.ReceivedAmount, result.Currency)
		a.logger.Println("   Waiting for blockchain confirmation...")
	}

	a.logger.Println("")
	a.logger.Printf("Checked: %s via %s\n",
		result.VerifiedAt.Format("2006-01-02 15:04:05"),
		result.VerifiedBy)

	if dryRun && result.IsSuccessful() {
		a.logger.Println("")
		a.logger.Println("ℹ️  Dry run mode: Invoice status not updated")
		a.logger.Println("   Remove --dry-run flag to mark invoice as paid")
	}

	a.logger.Println("")
}

// createBlockchainProvider creates a blockchain provider based on payment method
func (a *App) createBlockchainProvider(paymentMethod string, testnet bool, apiKey string) (blockchain.Provider, error) {
	switch paymentMethod {
	case "USDC":
		return blockchain.NewEtherscanProvider(apiKey, testnet), nil
	case "BSV":
		return blockchain.NewBSVProvider(testnet), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPaymentMethod, paymentMethod)
	}
}

// getInvoicePaymentAddress gets the payment address for an invoice
func (a *App) getInvoicePaymentAddress(invoice *models.Invoice, config *config.Config, method models.PaymentMethod) (string, error) {
	switch method {
	case models.PaymentMethodUSDC:
		if invoice.USDCAddressOverride != nil && *invoice.USDCAddressOverride != "" {
			return *invoice.USDCAddressOverride, nil
		}
		if config.Business.CryptoPayments.USDCAddress != "" {
			return config.Business.CryptoPayments.USDCAddress, nil
		}
		return "", ErrNoUSDCAddress

	case models.PaymentMethodBSV:
		if invoice.BSVAddressOverride != nil && *invoice.BSVAddressOverride != "" {
			return *invoice.BSVAddressOverride, nil
		}
		if config.Business.CryptoPayments.BSVAddress != "" {
			return config.Business.CryptoPayments.BSVAddress, nil
		}
		return "", ErrNoBSVAddress

	case models.PaymentMethodACH, models.PaymentMethodWire, models.PaymentMethodOther:
		return "", fmt.Errorf("%w for non-crypto payments: %s", ErrUnsupportedPaymentMethod, method)

	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedPaymentMethod, method)
	}
}
