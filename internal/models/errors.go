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

	// Service-level errors
	ErrClientNotFound                   = fmt.Errorf("client not found")
	ErrClientInactive                   = fmt.Errorf("client is inactive")
	ErrInvoiceIDEmpty                   = fmt.Errorf("invoice ID cannot be empty")
	ErrInvoiceNotFound                  = fmt.Errorf("invoice not found")
	ErrConfirmationRequired             = fmt.Errorf("confirmation required")
	ErrCannotDeletePaidInvoice          = fmt.Errorf("cannot delete paid invoice")
	ErrCannotAddWorkItemToNonDraft      = fmt.Errorf("can only add work items to draft invoices")
	ErrCannotRemoveWorkItemFromNonDraft = fmt.Errorf("can only remove work items from draft invoices")
	ErrCannotSendNonDraftInvoice        = fmt.Errorf("can only send draft invoices")
	ErrCannotSendEmptyInvoice           = fmt.Errorf("cannot send invoice with no work items")
	ErrCannotMarkNonSentAsPaid          = fmt.Errorf("can only mark sent or overdue invoices as paid")
	ErrInvoiceNumberExists              = fmt.Errorf("invoice number already exists")

	// Client service errors
	ErrClientIDEmpty                            = fmt.Errorf("client ID cannot be empty")
	ErrClientCannotBeNil                        = fmt.Errorf("client cannot be nil")
	ErrEmailEmpty                               = fmt.Errorf("email cannot be empty")
	ErrClientHasActiveInvoices                  = fmt.Errorf("cannot delete client with active invoices - please mark all invoices as paid or voided first")
	ErrCannotDeactivateClientWithActiveInvoices = fmt.Errorf("cannot deactivate client with active invoices")
	ErrClientEmailExists                        = fmt.Errorf("client with email already exists")

	// Render errors
	ErrTemplateNotFound     = fmt.Errorf("template not found")
	ErrTemplateCannotReload = fmt.Errorf("template cannot be reloaded (no source path)")

	// CLI client command errors
	ErrMultipleClientsFound     = fmt.Errorf("multiple clients found, please be more specific")
	ErrCannotActivateDeactivate = fmt.Errorf("cannot both activate and deactivate client")
	ErrNoUpdatesSpecified       = fmt.Errorf("no updates specified")
)
