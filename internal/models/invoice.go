package models

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

// Invoice represents a complete invoice entity
type Invoice struct {
	ID          InvoiceID  `json:"id"`
	Number      string     `json:"number"`
	Date        time.Time  `json:"date"`
	DueDate     time.Time  `json:"due_date"`
	Client      Client     `json:"client"`
	WorkItems   []WorkItem `json:"work_items"`
	Status      string     `json:"status"`
	Description string     `json:"description,omitempty"`
	Subtotal    float64    `json:"subtotal"`
	CryptoFee   float64    `json:"crypto_fee"`
	TaxRate     float64    `json:"tax_rate"`
	TaxAmount   float64    `json:"tax_amount"`
	Total       float64    `json:"total"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Version     int        `json:"version"` // For optimistic locking
}

// WorkItem represents a single work entry on an invoice
type WorkItem struct {
	ID          string    `json:"id"`
	Date        time.Time `json:"date"`
	Hours       float64   `json:"hours"`
	Rate        float64   `json:"rate"`
	Description string    `json:"description"`
	Total       float64   `json:"total"`
	CreatedAt   time.Time `json:"created_at"`
}

// Client represents customer information
type Client struct {
	ID               ClientID  `json:"id"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone,omitempty"`
	Address          string    `json:"address,omitempty"`
	TaxID            string    `json:"tax_id,omitempty"`
	ApproverContacts string    `json:"approver_contacts,omitempty"`
	Active           bool      `json:"active"`
	CryptoFeeEnabled bool      `json:"crypto_fee_enabled"`
	CryptoFeeAmount  float64   `json:"crypto_fee_amount,omitempty"`
	LateFeeEnabled   bool      `json:"late_fee_enabled"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// NewInvoice creates a new invoice with validation
func NewInvoice(ctx context.Context, id InvoiceID, number string, date, dueDate time.Time, client Client, taxRate float64) (*Invoice, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	now := time.Now()
	invoice := &Invoice{
		ID:        id,
		Number:    number,
		Date:      date,
		DueDate:   dueDate,
		Client:    client,
		WorkItems: make([]WorkItem, 0),
		Status:    StatusDraft,
		TaxRate:   taxRate,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}

	// Calculate initial totals (will be zero for empty work items)
	if err := invoice.RecalculateTotals(ctx); err != nil {
		return nil, fmt.Errorf("failed to calculate initial totals: %w", err)
	}

	// Validate the new invoice
	if err := invoice.Validate(ctx); err != nil {
		return nil, fmt.Errorf("invoice validation failed: %w", err)
	}

	return invoice, nil
}

// Validate performs comprehensive validation of the invoice
func (i *Invoice) Validate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	var errors []ValidationError

	// Validate all components
	i.validateBasicFields(&errors)
	i.validateDates(&errors)
	i.validateStatus(&errors)
	i.validateClientAndWorkItems(ctx, &errors)
	i.validateFinancials(&errors)
	i.validateTimestamps(&errors)
	i.validateVersion(&errors)

	return i.formatValidationErrors(errors)
}

// validateBasicFields validates ID and number fields
func (i *Invoice) validateBasicFields(errors *[]ValidationError) {
	if strings.TrimSpace(string(i.ID)) == "" {
		*errors = append(*errors, ValidationError{
			Field:   "id",
			Message: "is required",
			Value:   i.ID,
		})
	}

	if strings.TrimSpace(i.Number) == "" {
		*errors = append(*errors, ValidationError{
			Field:   "number",
			Message: "is required",
			Value:   i.Number,
		})
		return
	}

	if !invoiceIDPattern.MatchString(i.Number) {
		*errors = append(*errors, ValidationError{
			Field:   "number",
			Message: "must contain only uppercase letters, numbers, and hyphens",
			Value:   i.Number,
		})
	}
}

// validateDates validates date and due_date fields
func (i *Invoice) validateDates(errors *[]ValidationError) {
	if i.Date.IsZero() {
		*errors = append(*errors, ValidationError{
			Field:   "date",
			Message: "is required",
			Value:   i.Date,
		})
	}

	if i.DueDate.IsZero() {
		*errors = append(*errors, ValidationError{
			Field:   "due_date",
			Message: "is required",
			Value:   i.DueDate,
		})
		return
	}

	if !i.Date.IsZero() && i.DueDate.Before(i.Date) {
		*errors = append(*errors, ValidationError{
			Field:   "due_date",
			Message: "must be on or after invoice date",
			Value:   fmt.Sprintf("due: %v, invoice: %v", i.DueDate, i.Date),
		})
	}
}

// validateStatus validates the invoice status
func (i *Invoice) validateStatus(errors *[]ValidationError) {
	validStatuses := []string{StatusDraft, StatusSent, StatusPaid, StatusOverdue, StatusVoided}

	for _, status := range validStatuses {
		if i.Status == status {
			return
		}
	}

	*errors = append(*errors, ValidationError{
		Field:   "status",
		Message: fmt.Sprintf("must be one of: %s", strings.Join(validStatuses, ", ")),
		Value:   i.Status,
	})
}

// validateClientAndWorkItems validates client and work items
func (i *Invoice) validateClientAndWorkItems(ctx context.Context, errors *[]ValidationError) {
	if err := i.Client.Validate(ctx); err != nil {
		*errors = append(*errors, ValidationError{
			Field:   "client",
			Message: err.Error(),
			Value:   i.Client,
		})
	}

	for idx, item := range i.WorkItems {
		if err := item.Validate(ctx); err != nil {
			*errors = append(*errors, ValidationError{
				Field:   fmt.Sprintf("work_items[%d]", idx),
				Message: err.Error(),
				Value:   item,
			})
		}
	}
}

// validateFinancials validates financial amounts
func (i *Invoice) validateFinancials(errors *[]ValidationError) {
	if i.TaxRate < 0 || i.TaxRate > 1 {
		*errors = append(*errors, ValidationError{
			Field:   "tax_rate",
			Message: "must be between 0 and 1",
			Value:   i.TaxRate,
		})
	}

	if i.Subtotal < 0 {
		*errors = append(*errors, ValidationError{
			Field:   "subtotal",
			Message: "must be non-negative",
			Value:   i.Subtotal,
		})
	}

	if i.TaxAmount < 0 {
		*errors = append(*errors, ValidationError{
			Field:   "tax_amount",
			Message: "must be non-negative",
			Value:   i.TaxAmount,
		})
	}

	if i.Total < 0 {
		*errors = append(*errors, ValidationError{
			Field:   "total",
			Message: "must be non-negative",
			Value:   i.Total,
		})
	}
}

// validateTimestamps validates created_at and updated_at timestamps
func (i *Invoice) validateTimestamps(errors *[]ValidationError) {
	if i.CreatedAt.IsZero() {
		*errors = append(*errors, ValidationError{
			Field:   "created_at",
			Message: "is required",
			Value:   i.CreatedAt,
		})
	}

	if i.UpdatedAt.IsZero() {
		*errors = append(*errors, ValidationError{
			Field:   "updated_at",
			Message: "is required",
			Value:   i.UpdatedAt,
		})
		return
	}

	if !i.CreatedAt.IsZero() && i.UpdatedAt.Before(i.CreatedAt) {
		*errors = append(*errors, ValidationError{
			Field:   "updated_at",
			Message: "must be on or after created_at",
			Value:   fmt.Sprintf("updated: %v, created: %v", i.UpdatedAt, i.CreatedAt),
		})
	}
}

// validateVersion validates the version field
func (i *Invoice) validateVersion(errors *[]ValidationError) {
	if i.Version < 1 {
		*errors = append(*errors, ValidationError{
			Field:   "version",
			Message: "must be at least 1",
			Value:   i.Version,
		})
	}
}

// formatValidationErrors formats validation errors into a single error
func (i *Invoice) formatValidationErrors(errors []ValidationError) error {
	if len(errors) == 0 {
		return nil
	}

	messages := make([]string, 0, len(errors))
	for _, err := range errors {
		messages = append(messages, err.Error())
	}
	return fmt.Errorf("%w: %s", ErrInvoiceValidationFailed, strings.Join(messages, "; "))
}

// AddWorkItem adds a work item to the invoice and recalculates totals
func (i *Invoice) AddWorkItem(ctx context.Context, item WorkItem) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Validate the work item
	if err := item.Validate(ctx); err != nil {
		return fmt.Errorf("invalid work item: %w", err)
	}

	// Add the item
	i.WorkItems = append(i.WorkItems, item)

	// Recalculate totals
	if err := i.RecalculateTotals(ctx); err != nil {
		return fmt.Errorf("failed to recalculate totals after adding work item: %w", err)
	}

	// Update timestamp and version
	i.UpdatedAt = time.Now()
	i.Version++

	return nil
}

// AddWorkItemWithoutVersionIncrement adds a work item without incrementing version
// This is used for bulk operations where version will be handled by storage
func (i *Invoice) AddWorkItemWithoutVersionIncrement(ctx context.Context, item WorkItem) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Validate the work item
	if err := item.Validate(ctx); err != nil {
		return fmt.Errorf("invalid work item: %w", err)
	}

	// Add the item
	i.WorkItems = append(i.WorkItems, item)

	// Recalculate totals
	if err := i.RecalculateTotals(ctx); err != nil {
		return fmt.Errorf("failed to recalculate totals after adding work item: %w", err)
	}

	// Update timestamp but NOT version (for bulk operations)
	i.UpdatedAt = time.Now()

	return nil
}

// RemoveWorkItem removes a work item by ID and recalculates totals
func (i *Invoice) RemoveWorkItem(ctx context.Context, itemID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Find and remove the item
	found := false
	for idx, item := range i.WorkItems {
		if item.ID == itemID {
			// Remove item by slicing
			i.WorkItems = append(i.WorkItems[:idx], i.WorkItems[idx+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("%w: %s", ErrWorkItemNotFound, itemID)
	}

	// Recalculate totals
	if err := i.RecalculateTotals(ctx); err != nil {
		return fmt.Errorf("failed to recalculate totals after removing work item: %w", err)
	}

	// Update timestamp and version
	i.UpdatedAt = time.Now()
	i.Version++

	return nil
}

// RecalculateTotals recalculates all financial totals based on work items
func (i *Invoice) RecalculateTotals(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Calculate subtotal from work items
	subtotal := 0.0
	for _, item := range i.WorkItems {
		subtotal += item.Total
	}

	// Round to avoid floating point precision issues
	i.Subtotal = math.Round(subtotal*100) / 100

	// Calculate tax amount on (subtotal + crypto fee)
	taxableAmount := i.Subtotal + i.CryptoFee
	i.TaxAmount = math.Round(taxableAmount*i.TaxRate*100) / 100

	// Calculate total (subtotal + crypto fee + tax)
	i.Total = math.Round((taxableAmount+i.TaxAmount)*100) / 100

	return nil
}

// SetCryptoFee sets the cryptocurrency service fee if applicable
func (i *Invoice) SetCryptoFee(ctx context.Context, cryptoPaymentsEnabled, feeEnabled bool, feeAmount float64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Apply crypto service fee if crypto payments are enabled and fee is enabled
	if cryptoPaymentsEnabled && feeEnabled {
		i.CryptoFee = feeAmount
	} else {
		i.CryptoFee = 0.0
	}

	// Recalculate totals with the new crypto fee
	return i.RecalculateTotals(ctx)
}

// UpdateStatus updates the invoice status with validation
func (i *Invoice) UpdateStatus(ctx context.Context, newStatus string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Validate new status
	validStatuses := []string{StatusDraft, StatusSent, StatusPaid, StatusOverdue, StatusVoided}
	valid := false
	for _, status := range validStatuses {
		if newStatus == status {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("%w: '%s', must be one of: %s", ErrInvalidStatus, newStatus, strings.Join(validStatuses, ", "))
	}

	// Business rule validation (example: can't void a paid invoice)
	if i.Status == StatusPaid && newStatus == StatusVoided {
		return ErrCannotVoidPaidInvoice
	}

	// Update status
	i.Status = newStatus
	i.UpdatedAt = time.Now()
	// Version should only be incremented by the storage layer during save
	// i.Version++

	return nil
}

// IsOverdue checks if the invoice is overdue
func (i *Invoice) IsOverdue() bool {
	return i.Status != StatusPaid && i.Status != StatusVoided && time.Now().After(i.DueDate)
}

// GetAgeInDays returns the age of the invoice in days
func (i *Invoice) GetAgeInDays() int {
	return int(time.Since(i.Date).Hours() / 24)
}

// GetDaysUntilDue returns the number of days until the due date (negative if overdue)
func (i *Invoice) GetDaysUntilDue() int {
	return int(time.Until(i.DueDate).Hours() / 24)
}
