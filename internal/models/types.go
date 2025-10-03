package models

import (
	"context"
	"errors"
	"fmt"
	"math"
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

// ValidInvoiceStatuses contains all valid invoice status values
var ValidInvoiceStatuses = []string{StatusDraft, StatusSent, StatusPaid, StatusOverdue, StatusVoided} //nolint:gochecknoglobals // Constant-like status validation slice

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

// ValidationBuilder provides a systematic way to build validation error lists
// while reducing cyclomatic complexity in validation functions
type ValidationBuilder struct {
	errors []ValidationError
}

// NewValidationBuilder creates a new validation builder
func NewValidationBuilder() *ValidationBuilder {
	return &ValidationBuilder{
		errors: make([]ValidationError, 0),
	}
}

// WithContext sets the context for the validation builder
// Deprecated: Pass context directly to methods that need it
func (vb *ValidationBuilder) WithContext(_ context.Context) *ValidationBuilder {
	// Context is no longer stored - this method is kept for backward compatibility
	return vb
}

// AddRequired adds a required field validation error if the value is empty
func (vb *ValidationBuilder) AddRequired(field, value string) *ValidationBuilder {
	if strings.TrimSpace(value) == "" {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: "is required",
			Value:   value,
		})
	}
	return vb
}

// AddMaxLength adds a max length validation error if the value exceeds the limit
func (vb *ValidationBuilder) AddMaxLength(field, value string, maxLen int) *ValidationBuilder {
	if len(value) > maxLen {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("cannot exceed %d characters", maxLen),
			Value:   len(value),
		})
	}
	return vb
}

// AddMinLength adds a min length validation error if the value is below the limit
func (vb *ValidationBuilder) AddMinLength(field, value string, minLen int) *ValidationBuilder {
	trimmed := strings.TrimSpace(value)
	if value != "" && len(trimmed) < minLen {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be at least %d characters", minLen),
			Value:   len(value),
		})
	}
	return vb
}

// AddLengthRange adds a length range validation error if the value is outside the range
func (vb *ValidationBuilder) AddLengthRange(field, value string, minLen, maxLen int) *ValidationBuilder {
	if value != "" && (len(value) < minLen || len(value) > maxLen) {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be between %d and %d characters", minLen, maxLen),
			Value:   len(value),
		})
	}
	return vb
}

// AddEmail adds an email validation error if the value is not a valid email
func (vb *ValidationBuilder) AddEmail(field, value string) *ValidationBuilder {
	if value != "" && !emailPattern.MatchString(value) {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: "must be a valid email address",
			Value:   value,
		})
	}
	return vb
}

// AddTimeRequired adds a required time validation error if the time is zero
func (vb *ValidationBuilder) AddTimeRequired(field string, value time.Time) *ValidationBuilder {
	if value.IsZero() {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: "is required",
			Value:   value,
		})
	}
	return vb
}

// AddTimeOrder adds a time order validation error if before is not before after
func (vb *ValidationBuilder) AddTimeOrder(field string, before, after time.Time, beforeName, afterName string) *ValidationBuilder {
	if !before.IsZero() && !after.IsZero() && after.Before(before) {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be on or after %s", beforeName),
			Value:   fmt.Sprintf("%s: %v, %s: %v", afterName, after, beforeName, before),
		})
	}
	return vb
}

// AddCustom adds a custom validation error
func (vb *ValidationBuilder) AddCustom(field, message string, value interface{}) *ValidationBuilder {
	vb.errors = append(vb.errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
	return vb
}

// AddIf adds a validation error only if the condition is true
func (vb *ValidationBuilder) AddIf(condition bool, field, message string, value interface{}) *ValidationBuilder {
	if condition {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: message,
			Value:   value,
		})
	}
	return vb
}

// HasErrors returns true if there are any validation errors
func (vb *ValidationBuilder) HasErrors() bool {
	return len(vb.errors) > 0
}

// AddNonNegative adds a non-negative validation error if the value is negative
func (vb *ValidationBuilder) AddNonNegative(field string, value float64) *ValidationBuilder {
	if value < 0 {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: "must be non-negative",
			Value:   value,
		})
	}
	return vb
}

// AddNonNegativeInt adds a non-negative validation error if the value is negative
func (vb *ValidationBuilder) AddNonNegativeInt(field string, value int) *ValidationBuilder {
	if value < 0 {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: "must be non-negative",
			Value:   value,
		})
	}
	return vb
}

// AddDateRange adds a date range validation error if from is after to
func (vb *ValidationBuilder) AddDateRange(field string, from, to time.Time, fromName, toName string) *ValidationBuilder {
	if !from.IsZero() && !to.IsZero() && from.After(to) {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s must be before %s", fromName, toName),
			Value:   fmt.Sprintf("%v - %v", from, to),
		})
	}
	return vb
}

// AddAmountRange adds an amount range validation error if minVal is greater than maxVal
func (vb *ValidationBuilder) AddAmountRange(field string, minVal, maxVal float64) *ValidationBuilder {
	if minVal > 0 && maxVal > 0 && minVal > maxVal {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: "amount_min must be less than or equal to amount_max",
			Value:   fmt.Sprintf("%.2f - %.2f", minVal, maxVal),
		})
	}
	return vb
}

// AddValidOption adds a validation error if value is not in the list of valid options
func (vb *ValidationBuilder) AddValidOption(field, value string, validOptions []string) *ValidationBuilder {
	if value != "" {
		valid := false
		for _, option := range validOptions {
			if value == option {
				valid = true
				break
			}
		}
		if !valid {
			vb.errors = append(vb.errors, ValidationError{
				Field:   field,
				Message: fmt.Sprintf("must be one of: %s", strings.Join(validOptions, ", ")),
				Value:   value,
			})
		}
	}
	return vb
}

// Build builds the final validation error or returns nil if no errors
func (vb *ValidationBuilder) Build(baseError error) error {
	if len(vb.errors) == 0 {
		return nil
	}

	messages := make([]string, 0, len(vb.errors))
	for _, err := range vb.errors {
		messages = append(messages, err.Error())
	}
	return fmt.Errorf("%w: %s", baseError, strings.Join(messages, "; "))
}

// AddValidFloat adds a validation error if the value is NaN or Inf
func (vb *ValidationBuilder) AddValidFloat(field string, value float64) *ValidationBuilder {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: "must be a valid number",
			Value:   value,
		})
	}
	return vb
}

// AddPositive adds a validation error if the value is not positive
func (vb *ValidationBuilder) AddPositive(field string, value float64) *ValidationBuilder {
	if value <= 0 {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: "must be greater than 0",
			Value:   value,
		})
	}
	return vb
}

// AddMaxValue adds a validation error if the value exceeds the maximum
func (vb *ValidationBuilder) AddMaxValue(field string, value, maxVal float64, unit string) *ValidationBuilder {
	if value > maxVal {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("cannot exceed %s", unit),
			Value:   value,
		})
	}
	return vb
}

// AddDateNotFuture adds a validation error if the date is more than specified hours in the future
func (vb *ValidationBuilder) AddDateNotFuture(field string, value time.Time, allowedFutureHours int) *ValidationBuilder {
	if value.After(time.Now().Add(time.Duration(allowedFutureHours) * time.Hour)) {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("cannot be more than %d day in the future", allowedFutureHours/24),
			Value:   value,
		})
	}
	return vb
}

// AddCalculationValidation adds a validation error if the calculation is incorrect
func (vb *ValidationBuilder) AddCalculationValidation(field string, actual, expected float64) *ValidationBuilder {
	if math.Abs(actual-expected) > 0.01 {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("incorrect calculation, expected %.2f", expected),
			Value:   actual,
		})
	}
	return vb
}

// AddFloatValidation adds comprehensive float validation (valid, positive, max)
func (vb *ValidationBuilder) AddFloatValidation(field string, value, maxVal float64, unit string) *ValidationBuilder {
	return vb.
		AddValidFloat(field, value).
		AddPositive(field, value).
		AddMaxValue(field, value, maxVal, unit)
}

// AddPattern adds a pattern validation error if the value doesn't match the regex
func (vb *ValidationBuilder) AddPattern(field, value string, pattern *regexp.Regexp, message string) *ValidationBuilder {
	if value != "" && !pattern.MatchString(value) {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: message,
			Value:   value,
		})
	}
	return vb
}

// AddWorkItems validates a slice of work items
func (vb *ValidationBuilder) AddWorkItems(ctx context.Context, field string, items []WorkItem) *ValidationBuilder {
	for i, item := range items {
		if err := item.Validate(ctx); err != nil {
			vb.errors = append(vb.errors, ValidationError{
				Field:   fmt.Sprintf("%s[%d]", field, i),
				Message: err.Error(),
				Value:   item,
			})
		}
	}
	return vb
}

// AddRequiredPointer adds a required pointer field validation if pointer is nil or empty
func (vb *ValidationBuilder) AddRequiredPointer(field string, value *string, message string) *ValidationBuilder {
	if value != nil && strings.TrimSpace(*value) == "" {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: message,
			Value:   *value,
		})
	}
	return vb
}

// AddPatternPointer adds a pattern validation for pointer string if not nil
func (vb *ValidationBuilder) AddPatternPointer(field string, value *string, pattern *regexp.Regexp, message string) *ValidationBuilder {
	if value != nil && *value != "" && !pattern.MatchString(*value) {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: message,
			Value:   *value,
		})
	}
	return vb
}

// AddValidOptionPointer adds validation for pointer to string option
func (vb *ValidationBuilder) AddValidOptionPointer(field string, value *string, validOptions []string) *ValidationBuilder {
	if value != nil {
		valid := false
		for _, option := range validOptions {
			if *value == option {
				valid = true
				break
			}
		}
		if !valid {
			vb.errors = append(vb.errors, ValidationError{
				Field:   field,
				Message: fmt.Sprintf("must be one of: %s", strings.Join(validOptions, ", ")),
				Value:   *value,
			})
		}
	}
	return vb
}

// AddTimeOrderPointer adds validation for pointer time ordering
func (vb *ValidationBuilder) AddTimeOrderPointer(field string, before, after *time.Time, beforeName, afterName string) *ValidationBuilder {
	if before != nil && after != nil && after.Before(*before) {
		vb.errors = append(vb.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be on or after %s", beforeName),
			Value:   fmt.Sprintf("%s: %v, %s: %v", afterName, *after, beforeName, *before),
		})
	}
	return vb
}

// BuildWithMessage builds the final validation error with custom message or returns nil if no errors
func (vb *ValidationBuilder) BuildWithMessage(message string) error {
	if len(vb.errors) == 0 {
		return nil
	}

	messages := make([]string, 0, len(vb.errors))
	for _, err := range vb.errors {
		messages = append(messages, err.Error())
	}
	return fmt.Errorf("%w: %s: %s", ErrValidationFailed, message, strings.Join(messages, "; "))
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

	return NewValidationBuilder().
		AddValidOption("status", f.Status, ValidInvoiceStatuses).
		AddDateRange("date_range", f.DateFrom, f.DateTo, "date_from", "date_to").
		AddDateRange("due_date_range", f.DueDateFrom, f.DueDateTo, "due_date_from", "due_date_to").
		AddNonNegative("amount_min", f.AmountMin).
		AddNonNegative("amount_max", f.AmountMax).
		AddAmountRange("amount_range", f.AmountMin, f.AmountMax).
		AddNonNegativeInt("limit", f.Limit).
		AddNonNegativeInt("offset", f.Offset).
		BuildWithMessage("filter validation failed")
}

// CreateInvoiceRequest represents a request to create a new invoice
type CreateInvoiceRequest struct {
	Number      string     `json:"number"`
	ClientID    ClientID   `json:"client_id"`
	Date        time.Time  `json:"date"`
	DueDate     time.Time  `json:"due_date"`
	Description string     `json:"description,omitempty"`
	WorkItems   []WorkItem `json:"work_items,omitempty"`
	USDCAddress *string    `json:"usdc_address,omitempty"` // Optional USDC address override for this invoice
	BSVAddress  *string    `json:"bsv_address,omitempty"`  // Optional BSV address override for this invoice
}

// Validate validates the create invoice request
func (r *CreateInvoiceRequest) Validate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return NewValidationBuilder().
		AddRequired("number", r.Number).
		AddPattern("number", r.Number, invoiceIDPattern, "must contain only uppercase letters, numbers, and hyphens").
		AddRequired("client_id", string(r.ClientID)).
		AddTimeRequired("date", r.Date).
		AddTimeRequired("due_date", r.DueDate).
		AddTimeOrder("due_date", r.Date, r.DueDate, "invoice date", "due date").
		AddWorkItems(ctx, "work_items", r.WorkItems).
		BuildWithMessage("create invoice request validation failed")
}

// UpdateInvoiceRequest represents a request to update an invoice
type UpdateInvoiceRequest struct {
	ID          InvoiceID  `json:"id"`
	Number      *string    `json:"number,omitempty"`
	Date        *time.Time `json:"date,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Status      *string    `json:"status,omitempty"`
	Description *string    `json:"description,omitempty"`
	USDCAddress *string    `json:"usdc_address,omitempty"` // Optional USDC address override for this invoice
	BSVAddress  *string    `json:"bsv_address,omitempty"`  // Optional BSV address override for this invoice
}

// Validate validates the update invoice request
func (r *UpdateInvoiceRequest) Validate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return NewValidationBuilder().
		AddRequired("id", string(r.ID)).
		AddRequiredPointer("number", r.Number, "cannot be empty").
		AddPatternPointer("number", r.Number, invoiceIDPattern, "must contain only uppercase letters, numbers, and hyphens").
		AddValidOptionPointer("status", r.Status, ValidInvoiceStatuses).
		AddTimeOrderPointer("due_date", r.Date, r.DueDate, "invoice date", "due date").
		BuildWithMessage("update invoice request validation failed")
}
