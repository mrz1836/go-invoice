package models

import "fmt"

// Model validation errors
var (
	// Common validation errors
	ErrValidationFailed = fmt.Errorf("validation failed")

	// Client-related errors
	ErrClientValidationFailed     = fmt.Errorf("client validation failed")
	ErrClientNameTooLong          = fmt.Errorf("name cannot exceed 200 characters")
	ErrClientEmailInvalid         = fmt.Errorf("email must be a valid email address")
	ErrClientPhoneInvalid         = fmt.Errorf("phone must be between 10 and 20 characters")
	ErrClientAddressTooLong       = fmt.Errorf("address cannot exceed 500 characters")
	ErrClientTaxIDTooLong         = fmt.Errorf("tax ID cannot exceed 50 characters")
	ErrCreateClientRequestInvalid = fmt.Errorf("create client request validation failed")

	// Invoice-related errors
	ErrInvoiceValidationFailed = fmt.Errorf("invoice validation failed")
	ErrWorkItemNotFound        = fmt.Errorf("work item not found")
	ErrInvalidStatus           = fmt.Errorf("invalid status")
	ErrCannotVoidPaidInvoice   = fmt.Errorf("cannot void a paid invoice")

	// Work item-related errors
	ErrWorkItemValidationFailed = fmt.Errorf("work item validation failed")
	ErrHoursMustBePositive      = fmt.Errorf("hours must be greater than 0")
	ErrHoursExceedLimit         = fmt.Errorf("hours cannot exceed 24 per entry")
	ErrRateMustBePositive       = fmt.Errorf("rate must be greater than 0")
	ErrRateExceedsLimit         = fmt.Errorf("rate cannot exceed $10,000 per hour")
	ErrDescriptionTooLong       = fmt.Errorf("description cannot exceed 1000 characters")

	// Request validation errors
	ErrFilterValidationFailed      = fmt.Errorf("filter validation failed")
	ErrCreateInvoiceRequestInvalid = fmt.Errorf("create invoice request validation failed")
	ErrUpdateInvoiceRequestInvalid = fmt.Errorf("update invoice request validation failed")
)
