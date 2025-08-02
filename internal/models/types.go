package models

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// InvoiceID provides type-safe invoice identification.
type InvoiceID string

// ClientID provides type-safe client identification.
type ClientID string

// Invoice statuses
const (
	StatusDraft   = "draft"
	StatusSent    = "sent"
	StatusPaid    = "paid"
	StatusOverdue = "overdue"
	StatusVoided  = "voided"
)

// Validation patterns
var (
	invoiceIDPattern = regexp.MustCompile(`^[A-Z0-9-]+$`)
	emailPattern     = regexp.MustCompile(`^[a-zA-Z0-9]+([._%-+][a-zA-Z0-9]+)*@[a-zA-Z0-9]+([.-][a-zA-Z0-9]+)*\.[a-zA-Z]{2,}$`)
)

// Predefined errors for common validation failures
var (
	ErrNameRequired        = errors.New("name cannot be empty")
	ErrEmailRequired       = errors.New("email cannot be empty")
	ErrDescriptionRequired = errors.New("description cannot be empty")
)

// ValidationError represents a validation error with context
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
}

// Validator interface for validation operations
type Validator interface {
	ValidateInvoice(ctx context.Context, invoice *Invoice) error
	ValidateWorkItem(ctx context.Context, item *WorkItem) error
	ValidateClient(ctx context.Context, client *Client) error
}

// InvoiceFilter represents filtering options for invoice queries
type InvoiceFilter struct {
	Status      string    `json:"status,omitempty"`
	ClientID    ClientID  `json:"client_id,omitempty"`
	DateFrom    time.Time `json:"date_from,omitempty"`
	DateTo      time.Time `json:"date_to,omitempty"`
	DueDateFrom time.Time `json:"due_date_from,omitempty"`
	DueDateTo   time.Time `json:"due_date_to,omitempty"`
	AmountMin   float64   `json:"amount_min,omitempty"`
	AmountMax   float64   `json:"amount_max,omitempty"`
	Limit       int       `json:"limit,omitempty"`
	Offset      int       `json:"offset,omitempty"`
}

// Validate validates the invoice filter parameters
func (f *InvoiceFilter) Validate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	var errors []ValidationError

	// Validate status if provided
	if f.Status != "" {
		validStatuses := []string{StatusDraft, StatusSent, StatusPaid, StatusOverdue, StatusVoided}
		valid := false
		for _, status := range validStatuses {
			if f.Status == status {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, ValidationError{
				Field:   "status",
				Message: fmt.Sprintf("must be one of: %s", strings.Join(validStatuses, ", ")),
				Value:   f.Status,
			})
		}
	}

	// Validate date ranges
	if !f.DateFrom.IsZero() && !f.DateTo.IsZero() && f.DateFrom.After(f.DateTo) {
		errors = append(errors, ValidationError{
			Field:   "date_range",
			Message: "date_from must be before date_to",
			Value:   fmt.Sprintf("%v - %v", f.DateFrom, f.DateTo),
		})
	}

	if !f.DueDateFrom.IsZero() && !f.DueDateTo.IsZero() && f.DueDateFrom.After(f.DueDateTo) {
		errors = append(errors, ValidationError{
			Field:   "due_date_range",
			Message: "due_date_from must be before due_date_to",
			Value:   fmt.Sprintf("%v - %v", f.DueDateFrom, f.DueDateTo),
		})
	}

	// Validate amount ranges
	if f.AmountMin < 0 {
		errors = append(errors, ValidationError{
			Field:   "amount_min",
			Message: "must be non-negative",
			Value:   f.AmountMin,
		})
	}

	if f.AmountMax < 0 {
		errors = append(errors, ValidationError{
			Field:   "amount_max",
			Message: "must be non-negative",
			Value:   f.AmountMax,
		})
	}

	if f.AmountMin > 0 && f.AmountMax > 0 && f.AmountMin > f.AmountMax {
		errors = append(errors, ValidationError{
			Field:   "amount_range",
			Message: "amount_min must be less than or equal to amount_max",
			Value:   fmt.Sprintf("%.2f - %.2f", f.AmountMin, f.AmountMax),
		})
	}

	// Validate pagination
	if f.Limit < 0 {
		errors = append(errors, ValidationError{
			Field:   "limit",
			Message: "must be non-negative",
			Value:   f.Limit,
		})
	}

	if f.Offset < 0 {
		errors = append(errors, ValidationError{
			Field:   "offset",
			Message: "must be non-negative",
			Value:   f.Offset,
		})
	}

	if len(errors) > 0 {
		var messages []string
		for _, err := range errors {
			messages = append(messages, err.Error())
		}
		return fmt.Errorf("filter validation failed: %s", strings.Join(messages, "; "))
	}

	return nil
}

// CreateInvoiceRequest represents a request to create a new invoice
type CreateInvoiceRequest struct {
	Number      string     `json:"number"`
	ClientID    ClientID   `json:"client_id"`
	Date        time.Time  `json:"date"`
	DueDate     time.Time  `json:"due_date"`
	Description string     `json:"description,omitempty"`
	WorkItems   []WorkItem `json:"work_items,omitempty"`
}

// Validate validates the create invoice request
func (r *CreateInvoiceRequest) Validate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	var errors []ValidationError

	// Validate invoice number
	if strings.TrimSpace(r.Number) == "" {
		errors = append(errors, ValidationError{
			Field:   "number",
			Message: "is required",
			Value:   r.Number,
		})
	} else if !invoiceIDPattern.MatchString(r.Number) {
		errors = append(errors, ValidationError{
			Field:   "number",
			Message: "must contain only uppercase letters, numbers, and hyphens",
			Value:   r.Number,
		})
	}

	// Validate client ID
	if strings.TrimSpace(string(r.ClientID)) == "" {
		errors = append(errors, ValidationError{
			Field:   "client_id",
			Message: "is required",
			Value:   r.ClientID,
		})
	}

	// Validate dates
	if r.Date.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "date",
			Message: "is required",
			Value:   r.Date,
		})
	}

	if r.DueDate.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "due_date",
			Message: "is required",
			Value:   r.DueDate,
		})
	}

	if !r.Date.IsZero() && !r.DueDate.IsZero() && r.DueDate.Before(r.Date) {
		errors = append(errors, ValidationError{
			Field:   "due_date",
			Message: "must be on or after invoice date",
			Value:   fmt.Sprintf("due: %v, invoice: %v", r.DueDate, r.Date),
		})
	}

	// Validate work items if provided
	for i, item := range r.WorkItems {
		if err := item.Validate(ctx); err != nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("work_items[%d]", i),
				Message: err.Error(),
				Value:   item,
			})
		}
	}

	if len(errors) > 0 {
		var messages []string
		for _, err := range errors {
			messages = append(messages, err.Error())
		}
		return fmt.Errorf("create invoice request validation failed: %s", strings.Join(messages, "; "))
	}

	return nil
}

// UpdateInvoiceRequest represents a request to update an invoice
type UpdateInvoiceRequest struct {
	ID          InvoiceID  `json:"id"`
	Number      *string    `json:"number,omitempty"`
	Date        *time.Time `json:"date,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Status      *string    `json:"status,omitempty"`
	Description *string    `json:"description,omitempty"`
}

// Validate validates the update invoice request
func (r *UpdateInvoiceRequest) Validate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	var errors []ValidationError

	// Validate ID
	if strings.TrimSpace(string(r.ID)) == "" {
		errors = append(errors, ValidationError{
			Field:   "id",
			Message: "is required",
			Value:   r.ID,
		})
	}

	// Validate number if provided
	if r.Number != nil {
		if strings.TrimSpace(*r.Number) == "" {
			errors = append(errors, ValidationError{
				Field:   "number",
				Message: "cannot be empty",
				Value:   *r.Number,
			})
		} else if !invoiceIDPattern.MatchString(*r.Number) {
			errors = append(errors, ValidationError{
				Field:   "number",
				Message: "must contain only uppercase letters, numbers, and hyphens",
				Value:   *r.Number,
			})
		}
	}

	// Validate status if provided
	if r.Status != nil {
		validStatuses := []string{StatusDraft, StatusSent, StatusPaid, StatusOverdue, StatusVoided}
		valid := false
		for _, status := range validStatuses {
			if *r.Status == status {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, ValidationError{
				Field:   "status",
				Message: fmt.Sprintf("must be one of: %s", strings.Join(validStatuses, ", ")),
				Value:   *r.Status,
			})
		}
	}

	// Validate date consistency if both dates are provided
	if r.Date != nil && r.DueDate != nil && r.DueDate.Before(*r.Date) {
		errors = append(errors, ValidationError{
			Field:   "due_date",
			Message: "must be on or after invoice date",
			Value:   fmt.Sprintf("due: %v, invoice: %v", *r.DueDate, *r.Date),
		})
	}

	if len(errors) > 0 {
		var messages []string
		for _, err := range errors {
			messages = append(messages, err.Error())
		}
		return fmt.Errorf("update invoice request validation failed: %s", strings.Join(messages, "; "))
	}

	return nil
}
