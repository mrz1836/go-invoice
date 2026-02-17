package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mrz1836/go-invoice/internal/blockchain"
	"github.com/mrz1836/go-invoice/internal/models"
	"github.com/mrz1836/go-invoice/internal/storage"
)

var (
	// ErrNoPaymentAddress is returned when no payment address is configured
	ErrNoPaymentAddress = errors.New("no payment address configured")
	// ErrNoUSDCAddress is returned when no USDC address is configured
	ErrNoUSDCAddress = errors.New("no USDC address configured")
	// ErrNoBSVAddress is returned when no BSV address is configured
	ErrNoBSVAddress = errors.New("no BSV address configured")
	// ErrUnsupportedPaymentMethodForBlockchain is returned when payment method doesn't support blockchain verification
	ErrUnsupportedPaymentMethodForBlockchain = errors.New("unsupported payment method for blockchain verification")
)

// PaymentService provides payment verification and management operations
type PaymentService struct {
	invoiceStorage storage.InvoiceStorage
	logger         Logger // Logger interface is defined in invoice_service.go
}

// NewPaymentService creates a new payment service
func NewPaymentService(
	invoiceStorage storage.InvoiceStorage,
	logger Logger,
) *PaymentService {
	return &PaymentService{
		invoiceStorage: invoiceStorage,
		logger:         logger,
	}
}

// VerifyPayment verifies payment for an invoice using a blockchain provider
func (s *PaymentService) VerifyPayment(
	ctx context.Context,
	invoice *models.Invoice,
	provider blockchain.Provider,
	config PaymentVerificationConfig,
) (*models.PaymentVerification, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Info("verifying payment", "invoice_id", invoice.ID, "provider", provider.Name())

	// Determine which cryptocurrency address to use based on payment method
	address, token, err := s.getPaymentAddress(invoice, config)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment address: %w", err)
	}

	if address == "" {
		return nil, fmt.Errorf("%w for %s", ErrNoPaymentAddress, config.PaymentMethod)
	}

	s.logger.Debug("checking payment address", "address", address, "token", token, "expected_amount", invoice.Total)

	// Get current balance
	balanceResult, err := provider.GetBalance(ctx, address, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance from %s: %w", provider.Name(), err)
	}

	// Create payment verification result
	verification := &models.PaymentVerification{
		InvoiceID:      invoice.ID,
		Method:         config.PaymentMethod,
		ExpectedAmount: invoice.Total,
		ReceivedAmount: balanceResult.Balance,
		Currency:       string(token),
		WalletAddress:  address,
		VerifiedAt:     time.Now(),
		VerifiedBy:     provider.Name(),
		Metadata: models.PaymentMetadata{
			ProviderName: provider.Name(),
			CheckedAt:    time.Now(),
		},
	}

	// Try to get transaction details if balance is sufficient
	if balanceResult.Balance >= invoice.Total {
		txs, txErr := s.getRelevantTransactions(ctx, provider, address, token, invoice)
		if txErr != nil {
			s.logger.Debug("failed to get transactions", "error", txErr)
			// Don't fail verification if we can't get transactions, balance is enough
		} else if len(txs) > 0 {
			// Use the most recent transaction
			mostRecent := txs[len(txs)-1]
			verification.TransactionHash = mostRecent.Hash
			verification.BlockNumber = mostRecent.BlockNumber
			if mostRecent.Confirmed {
				verification.ConfirmedAt = &mostRecent.Timestamp
			}
			verification.Metadata.MultiplePayments = len(txs)
		}
	}

	// Determine payment status
	verification.Status = s.determinePaymentStatus(verification)

	s.logger.Info("payment verification completed",
		"invoice_id", invoice.ID,
		"status", verification.Status,
		"received", verification.ReceivedAmount,
		"expected", verification.ExpectedAmount)

	return verification, nil
}

// MarkInvoiceAsPaid updates an invoice status to paid with payment details
func (s *PaymentService) MarkInvoiceAsPaid(
	ctx context.Context,
	invoiceID models.InvoiceID,
	verification *models.PaymentVerification,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.logger.Info("marking invoice as paid", "invoice_id", invoiceID)

	// Get current invoice
	invoice, err := s.invoiceStorage.GetInvoice(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	// Check if invoice is already paid
	if invoice.Status == models.StatusPaid {
		s.logger.Debug("invoice already marked as paid", "invoice_id", invoiceID)
		return nil // Idempotent operation
	}

	// Build payment notes
	notes := s.buildPaymentNotes(verification)

	// Update invoice status
	paidStatus := models.StatusPaid
	updateReq := models.UpdateInvoiceRequest{
		ID:     invoiceID,
		Status: &paidStatus,
	}

	// Add description with payment details if transaction hash available
	if verification.TransactionHash != "" {
		desc := fmt.Sprintf("%s\n\nPayment Details:\n%s", invoice.Description, notes)
		updateReq.Description = &desc
	}

	// Update the invoice object
	invoice.Status = *updateReq.Status
	if updateReq.Description != nil {
		invoice.Description = *updateReq.Description
	}

	// Perform update (storage layer handles version increment)
	err = s.invoiceStorage.UpdateInvoice(ctx, invoice)
	if err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}

	s.logger.Info("invoice marked as paid", "invoice_id", invoiceID, "method", verification.Method)
	return nil
}

// getPaymentAddress determines which payment address to use for verification
func (s *PaymentService) getPaymentAddress(
	invoice *models.Invoice,
	config PaymentVerificationConfig,
) (string, blockchain.TokenType, error) {
	switch config.PaymentMethod {
	case models.PaymentMethodUSDC:
		// Check for invoice-specific override first
		if invoice.USDCAddressOverride != nil && *invoice.USDCAddressOverride != "" {
			return *invoice.USDCAddressOverride, blockchain.TokenTypeUSDC, nil
		}
		// Fall back to global config
		if config.DefaultUSDCAddress != "" {
			return config.DefaultUSDCAddress, blockchain.TokenTypeUSDC, nil
		}
		return "", blockchain.TokenTypeUSDC, ErrNoUSDCAddress

	case models.PaymentMethodBSV:
		// Check for invoice-specific override first
		if invoice.BSVAddressOverride != nil && *invoice.BSVAddressOverride != "" {
			return *invoice.BSVAddressOverride, blockchain.TokenTypeBSV, nil
		}
		// Fall back to global config
		if config.DefaultBSVAddress != "" {
			return config.DefaultBSVAddress, blockchain.TokenTypeBSV, nil
		}
		return "", blockchain.TokenTypeBSV, ErrNoBSVAddress

	case models.PaymentMethodACH, models.PaymentMethodWire, models.PaymentMethodOther:
		return "", "", fmt.Errorf("%w: %s", ErrUnsupportedPaymentMethodForBlockchain, config.PaymentMethod)

	default:
		return "", "", fmt.Errorf("%w: %s", ErrUnsupportedPaymentMethodForBlockchain, config.PaymentMethod)
	}
}

// getRelevantTransactions gets transactions that match the invoice criteria
func (s *PaymentService) getRelevantTransactions(
	ctx context.Context,
	provider blockchain.Provider,
	address string,
	token blockchain.TokenType,
	invoice *models.Invoice,
) ([]blockchain.Transaction, error) {
	// Look for transactions within a reasonable time window
	// From invoice creation to 30 days after due date
	startTime := invoice.CreatedAt
	endTime := invoice.DueDate.AddDate(0, 0, 30)

	minAmount := invoice.Total

	query := blockchain.TransactionQuery{
		Address:   address,
		Token:     token,
		StartTime: &startTime,
		EndTime:   &endTime,
		MinAmount: &minAmount,
	}

	return provider.GetTransactions(ctx, query)
}

// determinePaymentStatus determines the payment status based on amounts
func (s *PaymentService) determinePaymentStatus(verification *models.PaymentVerification) models.PaymentStatus {
	if verification.ReceivedAmount == 0 {
		return models.PaymentStatusNotFound
	}

	if verification.ReceivedAmount < verification.ExpectedAmount {
		// Check if it's close (within 1% tolerance for floating point)
		tolerance := verification.ExpectedAmount * 0.01
		if verification.ReceivedAmount >= verification.ExpectedAmount-tolerance {
			return models.PaymentStatusVerified
		}
		return models.PaymentStatusPartial
	}

	if verification.ReceivedAmount > verification.ExpectedAmount {
		// Check if it's overpaid (more than 1% over)
		tolerance := verification.ExpectedAmount * 0.01
		if verification.ReceivedAmount > verification.ExpectedAmount+tolerance {
			return models.PaymentStatusOverpaid
		}
	}

	return models.PaymentStatusVerified
}

// buildPaymentNotes creates formatted payment notes for the invoice
func (s *PaymentService) buildPaymentNotes(verification *models.PaymentVerification) string {
	var notes strings.Builder

	fmt.Fprintf(&notes, "Payment Method: %s\n", verification.Method)
	fmt.Fprintf(&notes, "Amount: %.2f %s\n", verification.ReceivedAmount, verification.Currency)
	fmt.Fprintf(&notes, "Wallet: %s\n", verification.WalletAddress)

	if verification.TransactionHash != "" {
		fmt.Fprintf(&notes, "Transaction: %s\n", verification.TransactionHash)
	}

	if verification.ConfirmedAt != nil {
		fmt.Fprintf(&notes, "Confirmed: %s\n", verification.ConfirmedAt.Format("2006-01-02 15:04:05"))
	}

	fmt.Fprintf(&notes, "Verified: %s via %s\n",
		verification.VerifiedAt.Format("2006-01-02 15:04:05"),
		verification.VerifiedBy)

	if verification.Status == models.PaymentStatusOverpaid {
		overpayment := verification.ReceivedAmount - verification.ExpectedAmount
		fmt.Fprintf(&notes, "Note: Overpaid by %.2f %s\n", overpayment, verification.Currency)
	}

	return notes.String()
}

// PaymentVerificationConfig contains configuration for payment verification
type PaymentVerificationConfig struct {
	PaymentMethod      models.PaymentMethod
	DefaultUSDCAddress string // From global config
	DefaultBSVAddress  string // From global config
}
