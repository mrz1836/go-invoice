package models

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ErrorsTestSuite struct {
	suite.Suite
}

func TestErrorsSuite(t *testing.T) {
	suite.Run(t, new(ErrorsTestSuite))
}

func (s *ErrorsTestSuite) TestCommonValidationErrors() {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ValidationFailed",
			err:      ErrValidationFailed,
			expected: "validation failed",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, tt.err.Error())
			s.Error(tt.err)
		})
	}
}

func (s *ErrorsTestSuite) TestClientRelatedErrors() {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ClientValidationFailed",
			err:      ErrClientValidationFailed,
			expected: "client validation failed",
		},
		{
			name:     "ClientNameTooLong",
			err:      ErrClientNameTooLong,
			expected: "name cannot exceed 200 characters",
		},
		{
			name:     "ClientEmailInvalid",
			err:      ErrClientEmailInvalid,
			expected: "email must be a valid email address",
		},
		{
			name:     "ClientPhoneInvalid",
			err:      ErrClientPhoneInvalid,
			expected: "phone must be between 10 and 20 characters",
		},
		{
			name:     "ClientAddressTooLong",
			err:      ErrClientAddressTooLong,
			expected: "address cannot exceed 500 characters",
		},
		{
			name:     "ClientTaxIDTooLong",
			err:      ErrClientTaxIDTooLong,
			expected: "tax ID cannot exceed 50 characters",
		},
		{
			name:     "CreateClientRequestInvalid",
			err:      ErrCreateClientRequestInvalid,
			expected: "create client request validation failed",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, tt.err.Error())
			s.Error(tt.err)
		})
	}
}

func (s *ErrorsTestSuite) TestInvoiceRelatedErrors() {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "InvoiceValidationFailed",
			err:      ErrInvoiceValidationFailed,
			expected: "invoice validation failed",
		},
		{
			name:     "WorkItemNotFound",
			err:      ErrWorkItemNotFound,
			expected: "work item not found",
		},
		{
			name:     "InvalidStatus",
			err:      ErrInvalidStatus,
			expected: "invalid status",
		},
		{
			name:     "CannotVoidPaidInvoice",
			err:      ErrCannotVoidPaidInvoice,
			expected: "cannot void a paid invoice",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, tt.err.Error())
			s.Error(tt.err)
		})
	}
}

func (s *ErrorsTestSuite) TestWorkItemRelatedErrors() {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "WorkItemValidationFailed",
			err:      ErrWorkItemValidationFailed,
			expected: "work item validation failed",
		},
		{
			name:     "HoursMustBePositive",
			err:      ErrHoursMustBePositive,
			expected: "hours must be greater than 0",
		},
		{
			name:     "HoursExceedLimit",
			err:      ErrHoursExceedLimit,
			expected: "hours cannot exceed 24 per entry",
		},
		{
			name:     "RateMustBePositive",
			err:      ErrRateMustBePositive,
			expected: "rate must be greater than 0",
		},
		{
			name:     "RateExceedsLimit",
			err:      ErrRateExceedsLimit,
			expected: "rate cannot exceed $10,000 per hour",
		},
		{
			name:     "DescriptionTooLong",
			err:      ErrDescriptionTooLong,
			expected: "description cannot exceed 1000 characters",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, tt.err.Error())
			s.Error(tt.err)
		})
	}
}

func (s *ErrorsTestSuite) TestRequestValidationErrors() {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "FilterValidationFailed",
			err:      ErrFilterValidationFailed,
			expected: "filter validation failed",
		},
		{
			name:     "CreateInvoiceRequestInvalid",
			err:      ErrCreateInvoiceRequestInvalid,
			expected: "create invoice request validation failed",
		},
		{
			name:     "UpdateInvoiceRequestInvalid",
			err:      ErrUpdateInvoiceRequestInvalid,
			expected: "update invoice request validation failed",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, tt.err.Error())
			s.Error(tt.err)
		})
	}
}

func (s *ErrorsTestSuite) TestServiceLevelErrors() {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ClientNotFound",
			err:      ErrClientNotFound,
			expected: "client not found",
		},
		{
			name:     "ClientInactive",
			err:      ErrClientInactive,
			expected: "client is inactive",
		},
		{
			name:     "InvoiceIDEmpty",
			err:      ErrInvoiceIDEmpty,
			expected: "invoice ID cannot be empty",
		},
		{
			name:     "CannotDeletePaidInvoice",
			err:      ErrCannotDeletePaidInvoice,
			expected: "cannot delete paid invoice",
		},
		{
			name:     "CannotAddWorkItemToNonDraft",
			err:      ErrCannotAddWorkItemToNonDraft,
			expected: "can only add work items to draft invoices",
		},
		{
			name:     "CannotRemoveWorkItemFromNonDraft",
			err:      ErrCannotRemoveWorkItemFromNonDraft,
			expected: "can only remove work items from draft invoices",
		},
		{
			name:     "CannotSendNonDraftInvoice",
			err:      ErrCannotSendNonDraftInvoice,
			expected: "can only send draft invoices",
		},
		{
			name:     "CannotSendEmptyInvoice",
			err:      ErrCannotSendEmptyInvoice,
			expected: "cannot send invoice with no work items",
		},
		{
			name:     "CannotMarkNonSentAsPaid",
			err:      ErrCannotMarkNonSentAsPaid,
			expected: "can only mark sent or overdue invoices as paid",
		},
		{
			name:     "InvoiceNumberExists",
			err:      ErrInvoiceNumberExists,
			expected: "invoice number already exists",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, tt.err.Error())
			s.Error(tt.err)
		})
	}
}

func (s *ErrorsTestSuite) TestClientServiceErrors() {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ClientIDEmpty",
			err:      ErrClientIDEmpty,
			expected: "client ID cannot be empty",
		},
		{
			name:     "ClientCannotBeNil",
			err:      ErrClientCannotBeNil,
			expected: "client cannot be nil",
		},
		{
			name:     "EmailEmpty",
			err:      ErrEmailEmpty,
			expected: "email cannot be empty",
		},
		{
			name:     "ClientHasActiveInvoices",
			err:      ErrClientHasActiveInvoices,
			expected: "cannot delete client with active invoices - please mark all invoices as paid or voided first",
		},
		{
			name:     "CannotDeactivateClientWithActiveInvoices",
			err:      ErrCannotDeactivateClientWithActiveInvoices,
			expected: "cannot deactivate client with active invoices",
		},
		{
			name:     "ClientEmailExists",
			err:      ErrClientEmailExists,
			expected: "client with email already exists",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, tt.err.Error())
			s.Error(tt.err)
		})
	}
}

func (s *ErrorsTestSuite) TestRenderErrors() {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "TemplateNotFound",
			err:      ErrTemplateNotFound,
			expected: "template not found",
		},
		{
			name:     "TemplateCannotReload",
			err:      ErrTemplateCannotReload,
			expected: "template cannot be reloaded (no source path)",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, tt.err.Error())
			s.Error(tt.err)
		})
	}
}

func (s *ErrorsTestSuite) TestErrorWrapping() {
	tests := []struct {
		name        string
		baseErr     error
		wrapMessage string
		expected    string
	}{
		{
			name:        "ClientValidationFailedWrapped",
			baseErr:     ErrClientValidationFailed,
			wrapMessage: "failed to create client",
			expected:    "failed to create client: client validation failed",
		},
		{
			name:        "InvoiceValidationFailedWrapped",
			baseErr:     ErrInvoiceValidationFailed,
			wrapMessage: "operation failed",
			expected:    "operation failed: invoice validation failed",
		},
		{
			name:        "WorkItemNotFoundWrapped",
			baseErr:     ErrWorkItemNotFound,
			wrapMessage: "update failed",
			expected:    "update failed: work item not found",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			wrappedErr := fmt.Errorf("%s: %w", tt.wrapMessage, tt.baseErr)
			s.Equal(tt.expected, wrappedErr.Error())
			s.ErrorIs(wrappedErr, tt.baseErr)
		})
	}
}

func (s *ErrorsTestSuite) TestErrorsIsComparison() {
	tests := []struct {
		name     string
		err1     error
		err2     error
		expected bool
	}{
		{
			name:     "SameErrors",
			err1:     ErrClientValidationFailed,
			err2:     ErrClientValidationFailed,
			expected: true,
		},
		{
			name:     "DifferentErrors",
			err1:     ErrClientValidationFailed,
			err2:     ErrInvoiceValidationFailed,
			expected: false,
		},
		{
			name:     "WrappedErrorMatchesBase",
			err1:     fmt.Errorf("context: %w", ErrClientNotFound),
			err2:     ErrClientNotFound,
			expected: true,
		},
		{
			name:     "WrappedErrorDoesNotMatchDifferent",
			err1:     fmt.Errorf("context: %w", ErrClientNotFound),
			err2:     ErrClientInactive,
			expected: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.expected {
				s.ErrorIs(tt.err1, tt.err2)
			} else {
				s.NotErrorIs(tt.err1, tt.err2)
			}
		})
	}
}

func (s *ErrorsTestSuite) TestErrorChaining() {
	baseErr := ErrClientNotFound
	contextErr := fmt.Errorf("service layer error: %w", baseErr)
	handlerErr := fmt.Errorf("handler error: %w", contextErr)

	// Test that we can unwrap through the chain
	s.Require().ErrorIs(handlerErr, baseErr)
	s.Require().ErrorIs(handlerErr, contextErr)
	s.Require().ErrorIs(contextErr, baseErr)

	// Test error messages include full chain
	s.Contains(handlerErr.Error(), "handler error")
	s.Contains(handlerErr.Error(), "service layer error")
	s.Contains(handlerErr.Error(), "client not found")
}

func (s *ErrorsTestSuite) TestAllErrorsAreNonNil() {
	errorsList := []struct {
		name string
		err  error
	}{
		{"ErrValidationFailed", ErrValidationFailed},
		{"ErrClientValidationFailed", ErrClientValidationFailed},
		{"ErrClientNameTooLong", ErrClientNameTooLong},
		{"ErrClientEmailInvalid", ErrClientEmailInvalid},
		{"ErrClientPhoneInvalid", ErrClientPhoneInvalid},
		{"ErrClientAddressTooLong", ErrClientAddressTooLong},
		{"ErrClientTaxIDTooLong", ErrClientTaxIDTooLong},
		{"ErrCreateClientRequestInvalid", ErrCreateClientRequestInvalid},
		{"ErrInvoiceValidationFailed", ErrInvoiceValidationFailed},
		{"ErrWorkItemNotFound", ErrWorkItemNotFound},
		{"ErrInvalidStatus", ErrInvalidStatus},
		{"ErrCannotVoidPaidInvoice", ErrCannotVoidPaidInvoice},
		{"ErrWorkItemValidationFailed", ErrWorkItemValidationFailed},
		{"ErrHoursMustBePositive", ErrHoursMustBePositive},
		{"ErrHoursExceedLimit", ErrHoursExceedLimit},
		{"ErrRateMustBePositive", ErrRateMustBePositive},
		{"ErrRateExceedsLimit", ErrRateExceedsLimit},
		{"ErrDescriptionTooLong", ErrDescriptionTooLong},
		{"ErrFilterValidationFailed", ErrFilterValidationFailed},
		{"ErrCreateInvoiceRequestInvalid", ErrCreateInvoiceRequestInvalid},
		{"ErrUpdateInvoiceRequestInvalid", ErrUpdateInvoiceRequestInvalid},
		{"ErrClientNotFound", ErrClientNotFound},
		{"ErrClientInactive", ErrClientInactive},
		{"ErrInvoiceIDEmpty", ErrInvoiceIDEmpty},
		{"ErrCannotDeletePaidInvoice", ErrCannotDeletePaidInvoice},
		{"ErrCannotAddWorkItemToNonDraft", ErrCannotAddWorkItemToNonDraft},
		{"ErrCannotRemoveWorkItemFromNonDraft", ErrCannotRemoveWorkItemFromNonDraft},
		{"ErrCannotSendNonDraftInvoice", ErrCannotSendNonDraftInvoice},
		{"ErrCannotSendEmptyInvoice", ErrCannotSendEmptyInvoice},
		{"ErrCannotMarkNonSentAsPaid", ErrCannotMarkNonSentAsPaid},
		{"ErrInvoiceNumberExists", ErrInvoiceNumberExists},
		{"ErrClientIDEmpty", ErrClientIDEmpty},
		{"ErrClientCannotBeNil", ErrClientCannotBeNil},
		{"ErrEmailEmpty", ErrEmailEmpty},
		{"ErrClientHasActiveInvoices", ErrClientHasActiveInvoices},
		{"ErrCannotDeactivateClientWithActiveInvoices", ErrCannotDeactivateClientWithActiveInvoices},
		{"ErrClientEmailExists", ErrClientEmailExists},
		{"ErrTemplateNotFound", ErrTemplateNotFound},
		{"ErrTemplateCannotReload", ErrTemplateCannotReload},
	}

	for _, tt := range errorsList {
		s.Run(tt.name, func() {
			s.Require().Error(tt.err, "Error %s should not be nil", tt.name)
			s.NotEmpty(tt.err.Error(), "Error %s should have a non-empty message", tt.name)
		})
	}
}

func (s *ErrorsTestSuite) TestErrorMessagesAreDescriptive() {
	errorsList := []struct {
		name          string
		err           error
		minLength     int
		shouldContain []string
		shouldNotBe   []string
	}{
		{
			name:          "ErrClientNameTooLong",
			err:           ErrClientNameTooLong,
			minLength:     20,
			shouldContain: []string{"name", "exceed", "200", "characters"},
			shouldNotBe:   []string{"error", "fail"},
		},
		{
			name:          "ErrClientEmailInvalid",
			err:           ErrClientEmailInvalid,
			minLength:     20,
			shouldContain: []string{"email", "valid"},
			shouldNotBe:   []string{"error"},
		},
		{
			name:          "ErrCannotVoidPaidInvoice",
			err:           ErrCannotVoidPaidInvoice,
			minLength:     15,
			shouldContain: []string{"cannot", "void", "paid", "invoice"},
			shouldNotBe:   []string{"error"},
		},
		{
			name:          "ErrRateExceedsLimit",
			err:           ErrRateExceedsLimit,
			minLength:     20,
			shouldContain: []string{"rate", "exceed", "$10,000", "hour"},
			shouldNotBe:   []string{"error"},
		},
	}

	for _, tt := range errorsList {
		s.Run(tt.name, func() {
			errMsg := tt.err.Error()
			s.GreaterOrEqual(len(errMsg), tt.minLength, "Error message should be descriptive enough")

			// Check that message contains expected terms
			for _, term := range tt.shouldContain {
				s.Contains(errMsg, term, "Error message should contain '%s'", term)
			}

			// Check that message doesn't contain generic terms
			for _, term := range tt.shouldNotBe {
				s.NotEqual(errMsg, term, "Error message should not be just '%s'", term)
			}
		})
	}
}

func (s *ErrorsTestSuite) TestErrorUniqueness() {
	// Collect all error messages to ensure they are unique
	errorMessages := make(map[string]string)
	duplicates := make([]string, 0)

	errorsList := []struct {
		name string
		err  error
	}{
		{"ErrValidationFailed", ErrValidationFailed},
		{"ErrClientValidationFailed", ErrClientValidationFailed},
		{"ErrClientNameTooLong", ErrClientNameTooLong},
		{"ErrClientEmailInvalid", ErrClientEmailInvalid},
		{"ErrClientPhoneInvalid", ErrClientPhoneInvalid},
		{"ErrClientAddressTooLong", ErrClientAddressTooLong},
		{"ErrClientTaxIDTooLong", ErrClientTaxIDTooLong},
		{"ErrCreateClientRequestInvalid", ErrCreateClientRequestInvalid},
		{"ErrInvoiceValidationFailed", ErrInvoiceValidationFailed},
		{"ErrWorkItemNotFound", ErrWorkItemNotFound},
		{"ErrInvalidStatus", ErrInvalidStatus},
		{"ErrCannotVoidPaidInvoice", ErrCannotVoidPaidInvoice},
		{"ErrWorkItemValidationFailed", ErrWorkItemValidationFailed},
		{"ErrHoursMustBePositive", ErrHoursMustBePositive},
		{"ErrHoursExceedLimit", ErrHoursExceedLimit},
		{"ErrRateMustBePositive", ErrRateMustBePositive},
		{"ErrRateExceedsLimit", ErrRateExceedsLimit},
		{"ErrDescriptionTooLong", ErrDescriptionTooLong},
		{"ErrFilterValidationFailed", ErrFilterValidationFailed},
		{"ErrCreateInvoiceRequestInvalid", ErrCreateInvoiceRequestInvalid},
		{"ErrUpdateInvoiceRequestInvalid", ErrUpdateInvoiceRequestInvalid},
		{"ErrClientNotFound", ErrClientNotFound},
		{"ErrClientInactive", ErrClientInactive},
		{"ErrInvoiceIDEmpty", ErrInvoiceIDEmpty},
		{"ErrCannotDeletePaidInvoice", ErrCannotDeletePaidInvoice},
		{"ErrCannotAddWorkItemToNonDraft", ErrCannotAddWorkItemToNonDraft},
		{"ErrCannotRemoveWorkItemFromNonDraft", ErrCannotRemoveWorkItemFromNonDraft},
		{"ErrCannotSendNonDraftInvoice", ErrCannotSendNonDraftInvoice},
		{"ErrCannotSendEmptyInvoice", ErrCannotSendEmptyInvoice},
		{"ErrCannotMarkNonSentAsPaid", ErrCannotMarkNonSentAsPaid},
		{"ErrInvoiceNumberExists", ErrInvoiceNumberExists},
		{"ErrClientIDEmpty", ErrClientIDEmpty},
		{"ErrClientCannotBeNil", ErrClientCannotBeNil},
		{"ErrEmailEmpty", ErrEmailEmpty},
		{"ErrClientHasActiveInvoices", ErrClientHasActiveInvoices},
		{"ErrCannotDeactivateClientWithActiveInvoices", ErrCannotDeactivateClientWithActiveInvoices},
		{"ErrClientEmailExists", ErrClientEmailExists},
		{"ErrTemplateNotFound", ErrTemplateNotFound},
		{"ErrTemplateCannotReload", ErrTemplateCannotReload},
	}

	for _, tt := range errorsList {
		message := tt.err.Error()
		if existingError, exists := errorMessages[message]; exists {
			duplicates = append(duplicates, fmt.Sprintf("%s and %s both have message: '%s'", existingError, tt.name, message))
		} else {
			errorMessages[message] = tt.name
		}
	}

	s.Empty(duplicates, "Found duplicate error messages: %v", duplicates)
}

// Benchmark tests for error operations
func BenchmarkErrorCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ErrClientValidationFailed
	}
}

func BenchmarkErrorWrapping(b *testing.B) {
	baseErr := ErrClientNotFound
	for i := 0; i < b.N; i++ {
		_ = fmt.Errorf("operation failed: %w", baseErr)
	}
}

func BenchmarkErrorUnwrapping(b *testing.B) {
	wrappedErr := fmt.Errorf("operation failed: %w", ErrClientNotFound)
	for i := 0; i < b.N; i++ {
		_ = errors.Is(wrappedErr, ErrClientNotFound)
	}
}
