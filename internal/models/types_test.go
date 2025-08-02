package models

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TypesTestSuite struct {
	suite.Suite
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func (suite *TypesTestSuite) SetupTest() {
	suite.ctx, suite.cancelFunc = context.WithTimeout(context.Background(), 5*time.Second)
}

func (suite *TypesTestSuite) TearDownTest() {
	suite.cancelFunc()
}

func TestTypesTestSuite(t *testing.T) {
	suite.Run(t, new(TypesTestSuite))
}

func (suite *TypesTestSuite) TestValidationError() {
	t := suite.T()

	tests := []struct {
		name        string
		err         ValidationError
		expectedMsg string
	}{
		{
			name: "StringValue",
			err: ValidationError{
				Field:   "email",
				Message: "must be a valid email address",
				Value:   "invalid-email",
			},
			expectedMsg: "validation failed for field 'email': must be a valid email address (value: invalid-email)",
		},
		{
			name: "NumericValue",
			err: ValidationError{
				Field:   "amount",
				Message: "must be greater than 0",
				Value:   -100.50,
			},
			expectedMsg: "validation failed for field 'amount': must be greater than 0 (value: -100.5)",
		},
		{
			name: "NilValue",
			err: ValidationError{
				Field:   "optional_field",
				Message: "is missing",
				Value:   nil,
			},
			expectedMsg: "validation failed for field 'optional_field': is missing (value: <nil>)",
		},
		{
			name: "ComplexValue",
			err: ValidationError{
				Field:   "date_range",
				Message: "start date must be before end date",
				Value:   "2024-12-01 - 2024-01-01",
			},
			expectedMsg: "validation failed for field 'date_range': start date must be before end date (value: 2024-12-01 - 2024-01-01)",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			assert.Equal(t, tt.expectedMsg, tt.err.Error())
		})
	}
}

func (suite *TypesTestSuite) TestInvoiceFilterValidate() {
	t := suite.T()

	tests := []struct {
		name        string
		filter      InvoiceFilter
		expectError bool
		errorMsg    string
	}{
		{
			name:        "ValidEmptyFilter",
			filter:      InvoiceFilter{},
			expectError: false,
		},
		{
			name: "ValidFilterWithAllFields",
			filter: InvoiceFilter{
				Status:      StatusPaid,
				ClientID:    "CLIENT-001",
				DateFrom:    time.Now().AddDate(0, -1, 0),
				DateTo:      time.Now(),
				DueDateFrom: time.Now(),
				DueDateTo:   time.Now().AddDate(0, 1, 0),
				AmountMin:   100.0,
				AmountMax:   1000.0,
				Limit:       50,
				Offset:      0,
			},
			expectError: false,
		},
		{
			name: "InvalidStatus",
			filter: InvoiceFilter{
				Status: "invalid-status",
			},
			expectError: true,
			errorMsg:    "validation failed for field 'status': must be one of:",
		},
		{
			name: "DateFromAfterDateTo",
			filter: InvoiceFilter{
				DateFrom: time.Now(),
				DateTo:   time.Now().AddDate(0, -1, 0),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'date_range': date_from must be before date_to",
		},
		{
			name: "DueDateFromAfterDueDateTo",
			filter: InvoiceFilter{
				DueDateFrom: time.Now(),
				DueDateTo:   time.Now().AddDate(0, -1, 0),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'due_date_range': due_date_from must be before due_date_to",
		},
		{
			name: "NegativeAmountMin",
			filter: InvoiceFilter{
				AmountMin: -100.0,
			},
			expectError: true,
			errorMsg:    "validation failed for field 'amount_min': must be non-negative",
		},
		{
			name: "NegativeAmountMax",
			filter: InvoiceFilter{
				AmountMax: -100.0,
			},
			expectError: true,
			errorMsg:    "validation failed for field 'amount_max': must be non-negative",
		},
		{
			name: "AmountMinGreaterThanMax",
			filter: InvoiceFilter{
				AmountMin: 1000.0,
				AmountMax: 100.0,
			},
			expectError: true,
			errorMsg:    "validation failed for field 'amount_range': amount_min must be less than or equal to amount_max",
		},
		{
			name: "NegativeLimit",
			filter: InvoiceFilter{
				Limit: -10,
			},
			expectError: true,
			errorMsg:    "validation failed for field 'limit': must be non-negative",
		},
		{
			name: "NegativeOffset",
			filter: InvoiceFilter{
				Offset: -10,
			},
			expectError: true,
			errorMsg:    "validation failed for field 'offset': must be non-negative",
		},
		{
			name: "ValidDateRangeWithSameDay",
			filter: InvoiceFilter{
				DateFrom: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				DateTo:   time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
			},
			expectError: false,
		},
		{
			name: "ValidAmountRangeWithSameValue",
			filter: InvoiceFilter{
				AmountMin: 500.0,
				AmountMax: 500.0,
			},
			expectError: false,
		},
		{
			name: "ZeroAmounts",
			filter: InvoiceFilter{
				AmountMin: 0,
				AmountMax: 0,
			},
			expectError: false,
		},
		{
			name: "OnlyDateFrom",
			filter: InvoiceFilter{
				DateFrom: time.Now().AddDate(0, -1, 0),
			},
			expectError: false,
		},
		{
			name: "OnlyDateTo",
			filter: InvoiceFilter{
				DateTo: time.Now(),
			},
			expectError: false,
		},
		{
			name: "MultipleValidationErrors",
			filter: InvoiceFilter{
				Status:    "invalid",
				AmountMin: -100,
				Limit:     -5,
			},
			expectError: true,
			errorMsg:    "filter validation failed:",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.filter.Validate(suite.ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func (suite *TypesTestSuite) TestInvoiceFilterValidateWithContext() {
	t := suite.T()

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	filter := InvoiceFilter{
		Status: StatusPaid,
		Limit:  10,
	}

	err := filter.Validate(ctx)
	assert.Equal(t, context.Canceled, err)
}

func (suite *TypesTestSuite) TestCreateInvoiceRequestValidate() {
	t := suite.T()

	tests := []struct {
		name        string
		request     CreateInvoiceRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "ValidRequest",
			request: CreateInvoiceRequest{
				Number:   "INV-2024-001",
				ClientID: "CLIENT-001",
				Date:     time.Now(),
				DueDate:  time.Now().AddDate(0, 0, 30),
			},
			expectError: false,
		},
		{
			name: "ValidRequestWithDescription",
			request: CreateInvoiceRequest{
				Number:      "INV-2024-001",
				ClientID:    "CLIENT-001",
				Date:        time.Now(),
				DueDate:     time.Now().AddDate(0, 0, 30),
				Description: "Monthly retainer invoice",
			},
			expectError: false,
		},
		{
			name: "ValidRequestWithWorkItems",
			request: CreateInvoiceRequest{
				Number:   "INV-2024-001",
				ClientID: "CLIENT-001",
				Date:     time.Now(),
				DueDate:  time.Now().AddDate(0, 0, 30),
				WorkItems: []WorkItem{
					{
						ID:          "ITEM-001",
						Date:        time.Now(),
						Hours:       8.0,
						Rate:        100.0,
						Description: "Development work",
						Total:       800.0,
						CreatedAt:   time.Now(),
					},
				},
			},
			expectError: false,
		},
		{
			name: "EmptyNumber",
			request: CreateInvoiceRequest{
				Number:   "",
				ClientID: "CLIENT-001",
				Date:     time.Now(),
				DueDate:  time.Now().AddDate(0, 0, 30),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'number': is required",
		},
		{
			name: "WhitespaceNumber",
			request: CreateInvoiceRequest{
				Number:   "   ",
				ClientID: "CLIENT-001",
				Date:     time.Now(),
				DueDate:  time.Now().AddDate(0, 0, 30),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'number': is required",
		},
		{
			name: "InvalidNumberFormat",
			request: CreateInvoiceRequest{
				Number:   "inv-2024-001", // lowercase not allowed
				ClientID: "CLIENT-001",
				Date:     time.Now(),
				DueDate:  time.Now().AddDate(0, 0, 30),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'number': must contain only uppercase letters, numbers, and hyphens",
		},
		{
			name: "NumberWithSpecialChars",
			request: CreateInvoiceRequest{
				Number:   "INV_2024@001",
				ClientID: "CLIENT-001",
				Date:     time.Now(),
				DueDate:  time.Now().AddDate(0, 0, 30),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'number': must contain only uppercase letters, numbers, and hyphens",
		},
		{
			name: "EmptyClientID",
			request: CreateInvoiceRequest{
				Number:   "INV-2024-001",
				ClientID: "",
				Date:     time.Now(),
				DueDate:  time.Now().AddDate(0, 0, 30),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'client_id': is required",
		},
		{
			name: "ZeroDate",
			request: CreateInvoiceRequest{
				Number:   "INV-2024-001",
				ClientID: "CLIENT-001",
				Date:     time.Time{},
				DueDate:  time.Now().AddDate(0, 0, 30),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'date': is required",
		},
		{
			name: "ZeroDueDate",
			request: CreateInvoiceRequest{
				Number:   "INV-2024-001",
				ClientID: "CLIENT-001",
				Date:     time.Now(),
				DueDate:  time.Time{},
			},
			expectError: true,
			errorMsg:    "validation failed for field 'due_date': is required",
		},
		{
			name: "DueDateBeforeDate",
			request: CreateInvoiceRequest{
				Number:   "INV-2024-001",
				ClientID: "CLIENT-001",
				Date:     time.Now(),
				DueDate:  time.Now().AddDate(0, 0, -7),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'due_date': must be on or after invoice date",
		},
		{
			name: "InvalidWorkItem",
			request: CreateInvoiceRequest{
				Number:   "INV-2024-001",
				ClientID: "CLIENT-001",
				Date:     time.Now(),
				DueDate:  time.Now().AddDate(0, 0, 30),
				WorkItems: []WorkItem{
					{
						ID:          "", // Invalid - empty ID
						Date:        time.Now(),
						Hours:       8.0,
						Rate:        100.0,
						Description: "Development work",
						Total:       800.0,
						CreatedAt:   time.Now(),
					},
				},
			},
			expectError: true,
			errorMsg:    "validation failed for field 'work_items[0]':",
		},
		{
			name: "MultipleErrors",
			request: CreateInvoiceRequest{
				Number:   "",
				ClientID: "",
				Date:     time.Time{},
				DueDate:  time.Time{},
			},
			expectError: true,
			errorMsg:    "create invoice request validation failed:",
		},
		{
			name: "SameDateAndDueDate",
			request: CreateInvoiceRequest{
				Number:   "INV-2024-001",
				ClientID: "CLIENT-001",
				Date:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				DueDate:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.request.Validate(suite.ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func (suite *TypesTestSuite) TestUpdateInvoiceRequestValidate() {
	t := suite.T()

	tests := []struct {
		name        string
		request     UpdateInvoiceRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "ValidUpdateNumber",
			request: UpdateInvoiceRequest{
				ID:     "INV-001",
				Number: ptrString("INV-2024-002"),
			},
			expectError: false,
		},
		{
			name: "ValidUpdateDate",
			request: UpdateInvoiceRequest{
				ID:   "INV-001",
				Date: ptrTime(time.Now()),
			},
			expectError: false,
		},
		{
			name: "ValidUpdateStatus",
			request: UpdateInvoiceRequest{
				ID:     "INV-001",
				Status: ptrString(StatusPaid),
			},
			expectError: false,
		},
		{
			name: "ValidUpdateMultipleFields",
			request: UpdateInvoiceRequest{
				ID:          "INV-001",
				Number:      ptrString("INV-2024-002"),
				Date:        ptrTime(time.Now()),
				DueDate:     ptrTime(time.Now().AddDate(0, 0, 30)),
				Status:      ptrString(StatusSent),
				Description: ptrString("Updated description"),
			},
			expectError: false,
		},
		{
			name: "EmptyID",
			request: UpdateInvoiceRequest{
				ID:     "",
				Number: ptrString("INV-2024-002"),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'id': is required",
		},
		{
			name: "WhitespaceID",
			request: UpdateInvoiceRequest{
				ID:     "   ",
				Number: ptrString("INV-2024-002"),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'id': is required",
		},
		{
			name: "EmptyNumber",
			request: UpdateInvoiceRequest{
				ID:     "INV-001",
				Number: ptrString(""),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'number': cannot be empty",
		},
		{
			name: "WhitespaceNumber",
			request: UpdateInvoiceRequest{
				ID:     "INV-001",
				Number: ptrString("   "),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'number': cannot be empty",
		},
		{
			name: "InvalidNumberFormat",
			request: UpdateInvoiceRequest{
				ID:     "INV-001",
				Number: ptrString("inv-2024-001"),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'number': must contain only uppercase letters, numbers, and hyphens",
		},
		{
			name: "InvalidStatus",
			request: UpdateInvoiceRequest{
				ID:     "INV-001",
				Status: ptrString("invalid-status"),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'status': must be one of:",
		},
		{
			name: "DueDateBeforeDate",
			request: UpdateInvoiceRequest{
				ID:      "INV-001",
				Date:    ptrTime(time.Now()),
				DueDate: ptrTime(time.Now().AddDate(0, 0, -7)),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'due_date': must be on or after invoice date",
		},
		{
			name: "ValidOnlyID",
			request: UpdateInvoiceRequest{
				ID: "INV-001",
			},
			expectError: false,
		},
		{
			name: "ValidWithNilFields",
			request: UpdateInvoiceRequest{
				ID:          "INV-001",
				Number:      nil,
				Date:        nil,
				DueDate:     nil,
				Status:      nil,
				Description: nil,
			},
			expectError: false,
		},
		{
			name: "ValidEmptyDescription",
			request: UpdateInvoiceRequest{
				ID:          "INV-001",
				Description: ptrString(""),
			},
			expectError: false,
		},
		{
			name: "ValidDatesWithSameValue",
			request: UpdateInvoiceRequest{
				ID:      "INV-001",
				Date:    ptrTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				DueDate: ptrTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.request.Validate(suite.ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func (suite *TypesTestSuite) TestStatusConstants() {
	t := suite.T()

	// Verify all status constants are defined correctly
	assert.Equal(t, "draft", StatusDraft)
	assert.Equal(t, "sent", StatusSent)
	assert.Equal(t, "paid", StatusPaid)
	assert.Equal(t, "overdue", StatusOverdue)
	assert.Equal(t, "voided", StatusVoided)

	// Verify they are all unique
	statuses := []string{StatusDraft, StatusSent, StatusPaid, StatusOverdue, StatusVoided}
	uniqueStatuses := make(map[string]bool)
	for _, status := range statuses {
		uniqueStatuses[status] = true
	}
	assert.Len(t, uniqueStatuses, len(statuses))
}

func (suite *TypesTestSuite) TestValidationPatterns() {
	t := suite.T()

	// Test invoice ID pattern
	validInvoiceIDs := []string{
		"INV-001",
		"INV-2024-001",
		"INVOICE-123",
		"A",
		"123",
		"ABC-123-XYZ",
		"TEST-INVOICE-2024-01-01",
	}

	for _, id := range validInvoiceIDs {
		suite.Run("ValidInvoiceID_"+id, func() {
			assert.True(t, invoiceIDPattern.MatchString(id), "Expected %s to be valid", id)
		})
	}

	invalidInvoiceIDs := []string{
		"inv-001",      // lowercase
		"INV_001",      // underscore
		"INV 001",      // space
		"INV@001",      // special char
		"",             // empty
		"INV-001!",     // special char at end
		"inv-2024-001", // mixed case
		"Invoice-001",  // mixed case
	}

	for _, id := range invalidInvoiceIDs {
		suite.Run("InvalidInvoiceID_"+id, func() {
			assert.False(t, invoiceIDPattern.MatchString(id), "Expected %s to be invalid", id)
		})
	}

	// Test email pattern
	validEmails := []string{
		"test@example.com",
		"user.name@example.com",
		"user+tag@example.com",
		"test@sub.example.com",
		"test123@example.co.uk",
		"a@b.co",
		"test_user@example.com",
		"TEST@EXAMPLE.COM",
		"123@example.com",
		"test.multiple.dots@example.com",
		"test+multiple+plus@example.com",
	}

	for _, email := range validEmails {
		suite.Run("ValidEmail_"+strings.ReplaceAll(email, "@", "_at_"), func() {
			assert.True(t, emailPattern.MatchString(email), "Expected %s to be valid", email)
		})
	}

	invalidEmails := []string{
		"notanemail",
		"@example.com",
		"test@",
		"test@@example.com",
		"test @example.com",
		"test@example",
		"test.example.com",
		"",
		"test@.com",
		"test@example.",
		"test..double@example.com",
		"test@example..com",
	}

	for _, email := range invalidEmails {
		suite.Run("InvalidEmail_"+strings.ReplaceAll(email, "@", "_at_"), func() {
			assert.False(t, emailPattern.MatchString(email), "Expected %s to be invalid", email)
		})
	}
}

// Helper functions for creating pointers
func ptrString(s string) *string {
	return &s
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
