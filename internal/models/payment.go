package models

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrInvoiceIDRequired is returned when invoice_id is missing
	ErrInvoiceIDRequired = errors.New("invoice_id is required")
	// ErrPaymentMethodRequired is returned when payment_method is missing
	ErrPaymentMethodRequired = errors.New("payment_method is required")
	// ErrInvalidPaymentMethod is returned when an invalid payment method is provided
	ErrInvalidPaymentMethod = errors.New("invalid payment_method")
)

// PaymentStatus represents the status of a payment verification
type PaymentStatus string

const (
	// PaymentStatusVerified indicates payment has been verified on-chain
	PaymentStatusVerified PaymentStatus = "verified"
	// PaymentStatusPending indicates payment is pending confirmation
	PaymentStatusPending PaymentStatus = "pending"
	// PaymentStatusNotFound indicates no payment was found
	PaymentStatusNotFound PaymentStatus = "not_found"
	// PaymentStatusPartial indicates partial payment was found
	PaymentStatusPartial PaymentStatus = "partial"
	// PaymentStatusOverpaid indicates more than required amount was received
	PaymentStatusOverpaid PaymentStatus = "overpaid"
)

// PaymentMethod represents the payment method used
type PaymentMethod string

const (
	// PaymentMethodUSDC represents payment via USDC stablecoin
	PaymentMethodUSDC PaymentMethod = "USDC"
	// PaymentMethodBSV represents payment via Bitcoin SV
	PaymentMethodBSV PaymentMethod = "BSV"
	// PaymentMethodACH represents payment via ACH bank transfer
	PaymentMethodACH PaymentMethod = "ACH"
	// PaymentMethodWire represents payment via wire transfer
	PaymentMethodWire PaymentMethod = "Wire"
	// PaymentMethodOther represents other payment methods
	PaymentMethodOther PaymentMethod = "Other"
)

// PaymentVerification represents the result of verifying a payment on-chain
type PaymentVerification struct {
	InvoiceID       InvoiceID       `json:"invoice_id"`
	Status          PaymentStatus   `json:"status"`
	Method          PaymentMethod   `json:"method"`
	ExpectedAmount  float64         `json:"expected_amount"`
	ReceivedAmount  float64         `json:"received_amount"`
	Currency        string          `json:"currency"`
	WalletAddress   string          `json:"wallet_address"`
	TransactionHash string          `json:"transaction_hash,omitempty"`
	BlockNumber     int64           `json:"block_number,omitempty"`
	ConfirmedAt     *time.Time      `json:"confirmed_at,omitempty"`
	VerifiedAt      time.Time       `json:"verified_at"`
	VerifiedBy      string          `json:"verified_by"` // Provider name
	Notes           string          `json:"notes,omitempty"`
	Metadata        PaymentMetadata `json:"metadata,omitempty"`
}

// PaymentMetadata contains additional payment verification metadata
type PaymentMetadata struct {
	ProviderName     string    `json:"provider_name"`
	CheckedAt        time.Time `json:"checked_at"`
	AddressReused    bool      `json:"address_reused"`    // Warning if address used by multiple invoices
	MultiplePayments int       `json:"multiple_payments"` // Number of payment transactions found
}

// IsSuccessful returns true if payment was successfully verified (verified or overpaid)
func (pv *PaymentVerification) IsSuccessful() bool {
	return pv.Status == PaymentStatusVerified || pv.Status == PaymentStatusOverpaid
}

// IsSufficient returns true if received amount is >= expected amount
func (pv *PaymentVerification) IsSufficient() bool {
	return pv.ReceivedAmount >= pv.ExpectedAmount
}

// AmountDifference returns the difference between received and expected amounts
// Positive means overpayment, negative means underpayment
func (pv *PaymentVerification) AmountDifference() float64 {
	return pv.ReceivedAmount - pv.ExpectedAmount
}

// VerifyPaymentRequest represents a request to verify payment for an invoice
type VerifyPaymentRequest struct {
	InvoiceID     InvoiceID     `json:"invoice_id"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	DryRun        bool          `json:"dry_run"` // If true, don't update invoice status
}

// Validate validates the verify payment request
func (r *VerifyPaymentRequest) Validate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if r.InvoiceID == "" {
		return ErrInvoiceIDRequired
	}

	if r.PaymentMethod == "" {
		return ErrPaymentMethodRequired
	}

	// Validate payment method
	validMethods := []PaymentMethod{
		PaymentMethodUSDC,
		PaymentMethodBSV,
		PaymentMethodACH,
		PaymentMethodWire,
		PaymentMethodOther,
	}

	valid := false
	for _, method := range validMethods {
		if r.PaymentMethod == method {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("%w: %s", ErrInvalidPaymentMethod, r.PaymentMethod)
	}

	return nil
}
