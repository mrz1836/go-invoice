package models

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type InvoiceTestSuite struct {
	suite.Suite

	ctx        context.Context //nolint:containedctx // Test suite context is acceptable
	cancelFunc context.CancelFunc
}

func (suite *InvoiceTestSuite) SetupTest() {
	suite.ctx, suite.cancelFunc = context.WithTimeout(context.Background(), 5*time.Second)
}

func (suite *InvoiceTestSuite) TearDownTest() {
	suite.cancelFunc()
}

func TestInvoiceTestSuite(t *testing.T) {
	suite.Run(t, new(InvoiceTestSuite))
}

// createTestInvoice is a helper function to create a basic test invoice
func createTestInvoice(t *testing.T, _ context.Context) *Invoice {
	t.Helper()

	return &Invoice{
		ID:        "INV-TEST-001",
		Number:    "TEST-001",
		Date:      time.Now(),
		DueDate:   time.Now().AddDate(0, 0, 30),
		Client:    Client{ID: "CLIENT-001", Name: "Test Client"},
		Status:    StatusDraft,
		TaxRate:   0.0,
		WorkItems: []WorkItem{},
		LineItems: []LineItem{},
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (suite *InvoiceTestSuite) TestNewInvoice() {
	t := suite.T()

	tests := []struct {
		name        string
		id          InvoiceID
		number      string
		date        time.Time
		dueDate     time.Time
		client      Client
		taxRate     float64
		expectError bool
		errorMsg    string
	}{
		{
			name:    "ValidInvoice",
			id:      "INV-001",
			number:  "INV-2024-001",
			date:    time.Now(),
			dueDate: time.Now().AddDate(0, 0, 30),
			client: Client{
				ID:        "CLIENT-001",
				Name:      "Test Client",
				Email:     "test@example.com",
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			taxRate:     0.1,
			expectError: false,
		},
		{
			name:    "EmptyID",
			id:      "",
			number:  "INV-2024-001",
			date:    time.Now(),
			dueDate: time.Now().AddDate(0, 0, 30),
			client: Client{
				ID:        "CLIENT-001",
				Name:      "Test Client",
				Email:     "test@example.com",
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			taxRate:     0.1,
			expectError: true,
			errorMsg:    "invoice validation failed",
		},
		{
			name:    "InvalidInvoiceNumber",
			id:      "INV-001",
			number:  "inv-2024-001", // lowercase not allowed
			date:    time.Now(),
			dueDate: time.Now().AddDate(0, 0, 30),
			client: Client{
				ID:        "CLIENT-001",
				Name:      "Test Client",
				Email:     "test@example.com",
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			taxRate:     0.1,
			expectError: true,
			errorMsg:    "must contain only uppercase letters, numbers, and hyphens",
		},
		{
			name:    "DueDateBeforeInvoiceDate",
			id:      "INV-001",
			number:  "INV-2024-001",
			date:    time.Now(),
			dueDate: time.Now().AddDate(0, 0, -1), // yesterday
			client: Client{
				ID:        "CLIENT-001",
				Name:      "Test Client",
				Email:     "test@example.com",
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			taxRate:     0.1,
			expectError: true,
			errorMsg:    "must be on or after invoice date",
		},
		{
			name:    "InvalidTaxRate",
			id:      "INV-001",
			number:  "INV-2024-001",
			date:    time.Now(),
			dueDate: time.Now().AddDate(0, 0, 30),
			client: Client{
				ID:        "CLIENT-001",
				Name:      "Test Client",
				Email:     "test@example.com",
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			taxRate:     1.5, // > 1.0
			expectError: true,
			errorMsg:    "must be between 0 and 1",
		},
		{
			name:    "InvalidClient",
			id:      "INV-001",
			number:  "INV-2024-001",
			date:    time.Now(),
			dueDate: time.Now().AddDate(0, 0, 30),
			client: Client{
				ID:        "",
				Name:      "Test Client",
				Email:     "test@example.com",
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			taxRate:     0.1,
			expectError: true,
			errorMsg:    "client validation failed",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			invoice, err := NewInvoice(suite.ctx, tt.id, tt.number, tt.date, tt.dueDate, tt.client, tt.taxRate)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, invoice)
			} else {
				require.NoError(t, err)
				require.NotNil(t, invoice)
				assert.Equal(t, tt.id, invoice.ID)
				assert.Equal(t, tt.number, invoice.Number)
				assert.Equal(t, StatusDraft, invoice.Status)
				assert.InDelta(t, tt.taxRate, invoice.TaxRate, 1e-9)
				assert.Empty(t, invoice.WorkItems)
				assert.InDelta(t, 0.0, invoice.Subtotal, 1e-9)
				assert.InDelta(t, 0.0, invoice.TaxAmount, 1e-9)
				assert.InDelta(t, 0.0, invoice.Total, 1e-9)
				assert.Equal(t, 1, invoice.Version)
			}
		})
	}
}

func (suite *InvoiceTestSuite) TestNewInvoiceWithContext() {
	t := suite.T()

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := Client{
		ID:        "CLIENT-001",
		Name:      "Test Client",
		Email:     "test@example.com",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	invoice, err := NewInvoice(ctx, "INV-001", "INV-2024-001", time.Now(), time.Now().AddDate(0, 0, 30), client, 0.1)
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Nil(t, invoice)
}

func (suite *InvoiceTestSuite) TestInvoiceValidate() {
	t := suite.T()

	tests := []struct {
		name        string
		invoice     Invoice
		expectError bool
		errorMsg    string
	}{
		{
			name: "ValidInvoice",
			invoice: Invoice{
				ID:      "INV-001",
				Number:  "INV-2024-001",
				Date:    time.Now(),
				DueDate: time.Now().AddDate(0, 0, 30),
				Client: Client{
					ID:        "CLIENT-001",
					Name:      "Test Client",
					Email:     "test@example.com",
					Active:    true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Status:    StatusDraft,
				TaxRate:   0.1,
				Subtotal:  100.0,
				TaxAmount: 10.0,
				Total:     110.0,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Version:   1,
			},
			expectError: false,
		},
		{
			name: "EmptyID",
			invoice: Invoice{
				ID:      "",
				Number:  "INV-2024-001",
				Date:    time.Now(),
				DueDate: time.Now().AddDate(0, 0, 30),
				Client: Client{
					ID:        "CLIENT-001",
					Name:      "Test Client",
					Email:     "test@example.com",
					Active:    true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Status:    StatusDraft,
				TaxRate:   0.1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Version:   1,
			},
			expectError: true,
			errorMsg:    "validation failed for field 'id': is required",
		},
		{
			name: "InvalidStatus",
			invoice: Invoice{
				ID:      "INV-001",
				Number:  "INV-2024-001",
				Date:    time.Now(),
				DueDate: time.Now().AddDate(0, 0, 30),
				Client: Client{
					ID:        "CLIENT-001",
					Name:      "Test Client",
					Email:     "test@example.com",
					Active:    true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Status:    "invalid-status",
				TaxRate:   0.1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Version:   1,
			},
			expectError: true,
			errorMsg:    "validation failed for field 'status': must be one of:",
		},
		{
			name: "NegativeSubtotal",
			invoice: Invoice{
				ID:      "INV-001",
				Number:  "INV-2024-001",
				Date:    time.Now(),
				DueDate: time.Now().AddDate(0, 0, 30),
				Client: Client{
					ID:        "CLIENT-001",
					Name:      "Test Client",
					Email:     "test@example.com",
					Active:    true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Status:    StatusDraft,
				TaxRate:   0.1,
				Subtotal:  -100.0,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Version:   1,
			},
			expectError: true,
			errorMsg:    "validation failed for field 'subtotal': must be non-negative",
		},
		{
			name: "InvalidVersion",
			invoice: Invoice{
				ID:      "INV-001",
				Number:  "INV-2024-001",
				Date:    time.Now(),
				DueDate: time.Now().AddDate(0, 0, 30),
				Client: Client{
					ID:        "CLIENT-001",
					Name:      "Test Client",
					Email:     "test@example.com",
					Active:    true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Status:    StatusDraft,
				TaxRate:   0.1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Version:   0,
			},
			expectError: true,
			errorMsg:    "validation failed for field 'version': must be at least 1",
		},
		{
			name: "UpdatedBeforeCreated",
			invoice: Invoice{
				ID:      "INV-001",
				Number:  "INV-2024-001",
				Date:    time.Now(),
				DueDate: time.Now().AddDate(0, 0, 30),
				Client: Client{
					ID:        "CLIENT-001",
					Name:      "Test Client",
					Email:     "test@example.com",
					Active:    true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Status:    StatusDraft,
				TaxRate:   0.1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now().AddDate(0, 0, -1),
				Version:   1,
			},
			expectError: true,
			errorMsg:    "validation failed for field 'updated_at': must be on or after created_at",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.invoice.Validate(suite.ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func (suite *InvoiceTestSuite) TestAddWorkItem() {
	t := suite.T()

	invoice := &Invoice{
		ID:        "INV-001",
		Number:    "INV-2024-001",
		Date:      time.Now(),
		DueDate:   time.Now().AddDate(0, 0, 30),
		Status:    StatusDraft,
		TaxRate:   0.1,
		WorkItems: []WorkItem{},
		Version:   1,
	}

	workItem := WorkItem{
		ID:          "ITEM-001",
		Date:        time.Now(),
		Hours:       8.0,
		Rate:        100.0,
		Description: "Development work",
		Total:       800.0,
		CreatedAt:   time.Now(),
	}

	err := invoice.AddWorkItem(suite.ctx, workItem)
	require.NoError(t, err)

	assert.Len(t, invoice.WorkItems, 1)
	assert.InDelta(t, 800.0, invoice.Subtotal, 1e-9)
	assert.InDelta(t, 80.0, invoice.TaxAmount, 1e-9)
	assert.InDelta(t, 880.0, invoice.Total, 1e-9)
	assert.Equal(t, 2, invoice.Version)

	// Add another work item
	workItem2 := WorkItem{
		ID:          "ITEM-002",
		Date:        time.Now(),
		Hours:       4.0,
		Rate:        150.0,
		Description: "Consulting work",
		Total:       600.0,
		CreatedAt:   time.Now(),
	}

	err = invoice.AddWorkItem(suite.ctx, workItem2)
	require.NoError(t, err)

	assert.Len(t, invoice.WorkItems, 2)
	assert.InDelta(t, 1400.0, invoice.Subtotal, 1e-9)
	assert.InDelta(t, 140.0, invoice.TaxAmount, 1e-9)
	assert.InDelta(t, 1540.0, invoice.Total, 1e-9)
	assert.Equal(t, 3, invoice.Version)
}

func (suite *InvoiceTestSuite) TestAddWorkItemValidation() {
	t := suite.T()

	invoice := &Invoice{
		ID:        "INV-001",
		Number:    "INV-2024-001",
		Date:      time.Now(),
		DueDate:   time.Now().AddDate(0, 0, 30),
		Status:    StatusDraft,
		TaxRate:   0.1,
		WorkItems: []WorkItem{},
		Version:   1,
	}

	// Invalid work item (missing ID)
	invalidItem := WorkItem{
		ID:          "",
		Date:        time.Now(),
		Hours:       8.0,
		Rate:        100.0,
		Description: "Development work",
		Total:       800.0,
		CreatedAt:   time.Now(),
	}

	err := invoice.AddWorkItem(suite.ctx, invalidItem)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid work item")
	assert.Empty(t, invoice.WorkItems)
	assert.Equal(t, 1, invoice.Version) // Version should not increment on error
}

func (suite *InvoiceTestSuite) TestRemoveWorkItem() {
	t := suite.T()

	invoice := &Invoice{
		ID:      "INV-001",
		Number:  "INV-2024-001",
		Date:    time.Now(),
		DueDate: time.Now().AddDate(0, 0, 30),
		Status:  StatusDraft,
		TaxRate: 0.1,
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
			{
				ID:          "ITEM-002",
				Date:        time.Now(),
				Hours:       4.0,
				Rate:        150.0,
				Description: "Consulting work",
				Total:       600.0,
				CreatedAt:   time.Now(),
			},
		},
		Subtotal:  1400.0,
		TaxAmount: 140.0,
		Total:     1540.0,
		Version:   1,
	}

	// Remove first item
	err := invoice.RemoveWorkItem(suite.ctx, "ITEM-001")
	require.NoError(t, err)

	assert.Len(t, invoice.WorkItems, 1)
	assert.Equal(t, "ITEM-002", invoice.WorkItems[0].ID)
	assert.InDelta(t, 600.0, invoice.Subtotal, 1e-9)
	assert.InDelta(t, 60.0, invoice.TaxAmount, 1e-9)
	assert.InDelta(t, 660.0, invoice.Total, 1e-9)
	assert.Equal(t, 2, invoice.Version)

	// Try to remove non-existent item
	err = invoice.RemoveWorkItem(suite.ctx, "ITEM-999")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "work item not found: ITEM-999")
	assert.Len(t, invoice.WorkItems, 1)
	assert.Equal(t, 2, invoice.Version) // Version should not increment on error
}

func (suite *InvoiceTestSuite) TestRecalculateTotals() {
	t := suite.T()

	tests := []struct {
		name             string
		workItems        []WorkItem
		taxRate          float64
		expectedSubtotal float64
		expectedTax      float64
		expectedTotal    float64
	}{
		{
			name:             "NoWorkItems",
			workItems:        []WorkItem{},
			taxRate:          0.1,
			expectedSubtotal: 0.0,
			expectedTax:      0.0,
			expectedTotal:    0.0,
		},
		{
			name: "SingleWorkItem",
			workItems: []WorkItem{
				{Total: 1000.0},
			},
			taxRate:          0.1,
			expectedSubtotal: 1000.0,
			expectedTax:      100.0,
			expectedTotal:    1100.0,
		},
		{
			name: "MultipleWorkItems",
			workItems: []WorkItem{
				{Total: 1000.0},
				{Total: 500.0},
				{Total: 250.0},
			},
			taxRate:          0.15,
			expectedSubtotal: 1750.0,
			expectedTax:      262.50,
			expectedTotal:    2012.50,
		},
		{
			name: "NoTax",
			workItems: []WorkItem{
				{Total: 1000.0},
			},
			taxRate:          0.0,
			expectedSubtotal: 1000.0,
			expectedTax:      0.0,
			expectedTotal:    1000.0,
		},
		{
			name: "FloatingPointPrecision",
			workItems: []WorkItem{
				{Total: 33.33},
				{Total: 66.67},
			},
			taxRate:          0.1,
			expectedSubtotal: 100.0,
			expectedTax:      10.0,
			expectedTotal:    110.0,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			invoice := &Invoice{
				WorkItems: tt.workItems,
				TaxRate:   tt.taxRate,
			}

			err := invoice.RecalculateTotals(suite.ctx)
			require.NoError(t, err)

			assert.InDelta(t, tt.expectedSubtotal, invoice.Subtotal, 1e-9)
			assert.InDelta(t, tt.expectedTax, invoice.TaxAmount, 1e-9)
			assert.InDelta(t, tt.expectedTotal, invoice.Total, 1e-9)
		})
	}
}

func (suite *InvoiceTestSuite) TestUpdateStatus() {
	t := suite.T()

	invoice := &Invoice{
		ID:        "INV-001",
		Status:    StatusDraft,
		Version:   1,
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	originalUpdatedAt := invoice.UpdatedAt

	// Valid status update
	err := invoice.UpdateStatus(suite.ctx, StatusSent)
	require.NoError(t, err)
	assert.Equal(t, StatusSent, invoice.Status)
	assert.Equal(t, 1, invoice.Version) // Version not incremented by UpdateStatus, done by storage layer
	assert.True(t, invoice.UpdatedAt.After(originalUpdatedAt))

	// Invalid status
	err = invoice.UpdateStatus(suite.ctx, "invalid-status")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
	assert.Equal(t, StatusSent, invoice.Status)
	assert.Equal(t, 1, invoice.Version) // Version unchanged after failed update

	// Business rule: can't void a paid invoice
	invoice.Status = StatusPaid
	err = invoice.UpdateStatus(suite.ctx, StatusVoided)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot void a paid invoice")
	assert.Equal(t, StatusPaid, invoice.Status)
}

func (suite *InvoiceTestSuite) TestIsOverdue() {
	t := suite.T()

	tests := []struct {
		name     string
		status   string
		dueDate  time.Time
		expected bool
	}{
		{
			name:     "NotOverdueFutureDueDate",
			status:   StatusSent,
			dueDate:  time.Now().AddDate(0, 0, 7),
			expected: false,
		},
		{
			name:     "OverduePastDueDate",
			status:   StatusSent,
			dueDate:  time.Now().AddDate(0, 0, -7),
			expected: true,
		},
		{
			name:     "PaidNotOverdue",
			status:   StatusPaid,
			dueDate:  time.Now().AddDate(0, 0, -7),
			expected: false,
		},
		{
			name:     "VoidedNotOverdue",
			status:   StatusVoided,
			dueDate:  time.Now().AddDate(0, 0, -7),
			expected: false,
		},
		{
			name:     "DraftOverdue",
			status:   StatusDraft,
			dueDate:  time.Now().AddDate(0, 0, -1),
			expected: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			invoice := &Invoice{
				Status:  tt.status,
				DueDate: tt.dueDate,
			}
			assert.Equal(t, tt.expected, invoice.IsOverdue())
		})
	}
}

func (suite *InvoiceTestSuite) TestGetAgeInDays() {
	t := suite.T()

	tests := []struct {
		name     string
		date     time.Time
		expected int
	}{
		{
			name:     "Today",
			date:     time.Now(),
			expected: 0,
		},
		{
			name:     "Yesterday",
			date:     time.Now().AddDate(0, 0, -1),
			expected: 1,
		},
		{
			name:     "LastWeek",
			date:     time.Now().AddDate(0, 0, -7),
			expected: 7,
		},
		{
			name:     "LastMonth",
			date:     time.Now().AddDate(0, -1, 0),
			expected: 30, // approximately
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			invoice := &Invoice{
				Date: tt.date,
			}
			age := invoice.GetAgeInDays()
			// Allow for small differences due to exact timing
			assert.InDelta(t, tt.expected, age, 1)
		})
	}
}

func (suite *InvoiceTestSuite) TestGetDaysUntilDue() {
	t := suite.T()

	tests := []struct {
		name     string
		dueDate  time.Time
		expected int
	}{
		{
			name:     "DueToday",
			dueDate:  time.Now(),
			expected: 0,
		},
		{
			name:     "DueTomorrow",
			dueDate:  time.Now().AddDate(0, 0, 1),
			expected: 1,
		},
		{
			name:     "DueNextWeek",
			dueDate:  time.Now().AddDate(0, 0, 7),
			expected: 7,
		},
		{
			name:     "OverdueYesterday",
			dueDate:  time.Now().AddDate(0, 0, -1),
			expected: -1,
		},
		{
			name:     "OverdueLastWeek",
			dueDate:  time.Now().AddDate(0, 0, -7),
			expected: -7,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			invoice := &Invoice{
				DueDate: tt.dueDate,
			}
			days := invoice.GetDaysUntilDue()
			// Allow for small differences due to exact timing
			assert.InDelta(t, tt.expected, days, 1)
		})
	}
}

func (suite *InvoiceTestSuite) TestSetCryptoFee() {
	t := suite.T()

	tests := []struct {
		name                  string
		cryptoPaymentsEnabled bool
		feeEnabled            bool
		feeAmount             float64
		workItemTotal         float64
		taxRate               float64
		expectedCryptoFee     float64
		expectedSubtotal      float64
		expectedTaxAmount     float64
		expectedTotal         float64
	}{
		{
			name:                  "CryptoEnabledAndFeeEnabled",
			cryptoPaymentsEnabled: true,
			feeEnabled:            true,
			feeAmount:             25.00,
			workItemTotal:         1000.00,
			taxRate:               0.10,
			expectedCryptoFee:     25.00,
			expectedSubtotal:      1000.00,
			expectedTaxAmount:     102.50,  // Tax on (1000 + 25)
			expectedTotal:         1127.50, // 1000 + 25 + 102.50
		},
		{
			name:                  "CryptoEnabledButFeeDisabled",
			cryptoPaymentsEnabled: true,
			feeEnabled:            false,
			feeAmount:             25.00,
			workItemTotal:         1000.00,
			taxRate:               0.10,
			expectedCryptoFee:     0.00,
			expectedSubtotal:      1000.00,
			expectedTaxAmount:     100.00,  // Tax on 1000 only
			expectedTotal:         1100.00, // 1000 + 100
		},
		{
			name:                  "CryptoDisabledFeeEnabled",
			cryptoPaymentsEnabled: false,
			feeEnabled:            true,
			feeAmount:             25.00,
			workItemTotal:         1000.00,
			taxRate:               0.10,
			expectedCryptoFee:     0.00,
			expectedSubtotal:      1000.00,
			expectedTaxAmount:     100.00,
			expectedTotal:         1100.00,
		},
		{
			name:                  "BothDisabled",
			cryptoPaymentsEnabled: false,
			feeEnabled:            false,
			feeAmount:             25.00,
			workItemTotal:         1000.00,
			taxRate:               0.10,
			expectedCryptoFee:     0.00,
			expectedSubtotal:      1000.00,
			expectedTaxAmount:     100.00,
			expectedTotal:         1100.00,
		},
		{
			name:                  "ZeroFeeAmount",
			cryptoPaymentsEnabled: true,
			feeEnabled:            true,
			feeAmount:             0.00,
			workItemTotal:         1000.00,
			taxRate:               0.10,
			expectedCryptoFee:     0.00,
			expectedSubtotal:      1000.00,
			expectedTaxAmount:     100.00,
			expectedTotal:         1100.00,
		},
		{
			name:                  "LargeFeeAmount",
			cryptoPaymentsEnabled: true,
			feeEnabled:            true,
			feeAmount:             100.00,
			workItemTotal:         5000.00,
			taxRate:               0.10,
			expectedCryptoFee:     100.00,
			expectedSubtotal:      5000.00,
			expectedTaxAmount:     510.00,  // Tax on (5000 + 100)
			expectedTotal:         5610.00, // 5000 + 100 + 510
		},
		{
			name:                  "NoWorkItemsWithCryptoFee",
			cryptoPaymentsEnabled: true,
			feeEnabled:            true,
			feeAmount:             25.00,
			workItemTotal:         0.00,
			taxRate:               0.10,
			expectedCryptoFee:     25.00,
			expectedSubtotal:      0.00,
			expectedTaxAmount:     2.50,  // Tax on 25
			expectedTotal:         27.50, // 0 + 25 + 2.50
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Create invoice with work item
			invoice := &Invoice{
				ID:        "INV-001",
				Number:    "INV-2024-001",
				Date:      time.Now(),
				DueDate:   time.Now().AddDate(0, 0, 30),
				Status:    StatusDraft,
				TaxRate:   tt.taxRate,
				WorkItems: []WorkItem{},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Version:   1,
			}

			// Add work item if specified
			if tt.workItemTotal > 0 {
				workItem := WorkItem{
					ID:          "ITEM-001",
					Date:        time.Now(),
					Hours:       8.0,
					Rate:        tt.workItemTotal / 8.0,
					Description: "Test work",
					Total:       tt.workItemTotal,
					CreatedAt:   time.Now(),
				}
				invoice.WorkItems = append(invoice.WorkItems, workItem)
			}

			// Recalculate base totals first
			err := invoice.RecalculateTotals(suite.ctx)
			require.NoError(t, err)

			// Set crypto fee
			err = invoice.SetCryptoFee(suite.ctx, tt.cryptoPaymentsEnabled, tt.feeEnabled, tt.feeAmount)
			require.NoError(t, err)

			// Verify results
			assert.InDelta(t, tt.expectedCryptoFee, invoice.CryptoFee, 1e-9, "crypto fee mismatch")
			assert.InDelta(t, tt.expectedSubtotal, invoice.Subtotal, 1e-9, "subtotal mismatch")
			assert.InDelta(t, tt.expectedTaxAmount, invoice.TaxAmount, 1e-9, "tax amount mismatch")
			assert.InDelta(t, tt.expectedTotal, invoice.Total, 1e-9, "total mismatch")
		})
	}
}

func (suite *InvoiceTestSuite) TestRecalculateTotalsWithCryptoFee() {
	t := suite.T()

	tests := []struct {
		name             string
		workItems        []WorkItem
		cryptoFee        float64
		taxRate          float64
		expectedSubtotal float64
		expectedTax      float64
		expectedTotal    float64
	}{
		{
			name: "SingleWorkItemWithCryptoFee",
			workItems: []WorkItem{
				{Total: 1000.0},
			},
			cryptoFee:        25.00,
			taxRate:          0.10,
			expectedSubtotal: 1000.00,
			expectedTax:      102.50,  // (1000 + 25) * 0.10
			expectedTotal:    1127.50, // 1000 + 25 + 102.50
		},
		{
			name: "MultipleWorkItemsWithCryptoFee",
			workItems: []WorkItem{
				{Total: 1000.0},
				{Total: 500.0},
				{Total: 250.0},
			},
			cryptoFee:        50.00,
			taxRate:          0.10,
			expectedSubtotal: 1750.00,
			expectedTax:      180.00,  // (1750 + 50) * 0.10
			expectedTotal:    1980.00, // 1750 + 50 + 180
		},
		{
			name: "ZeroTaxWithCryptoFee",
			workItems: []WorkItem{
				{Total: 1000.0},
			},
			cryptoFee:        25.00,
			taxRate:          0.00,
			expectedSubtotal: 1000.00,
			expectedTax:      0.00,
			expectedTotal:    1025.00, // 1000 + 25 + 0
		},
		{
			name: "HighTaxWithCryptoFee",
			workItems: []WorkItem{
				{Total: 1000.0},
			},
			cryptoFee:        25.00,
			taxRate:          0.20,
			expectedSubtotal: 1000.00,
			expectedTax:      205.00,  // (1000 + 25) * 0.20
			expectedTotal:    1230.00, // 1000 + 25 + 205
		},
		{
			name:             "NoCryptoFee",
			workItems:        []WorkItem{{Total: 1000.0}},
			cryptoFee:        0.00,
			taxRate:          0.10,
			expectedSubtotal: 1000.00,
			expectedTax:      100.00,
			expectedTotal:    1100.00,
		},
		{
			name:             "CryptoFeeOnlyNoWorkItems",
			workItems:        []WorkItem{},
			cryptoFee:        25.00,
			taxRate:          0.10,
			expectedSubtotal: 0.00,
			expectedTax:      2.50,  // 25 * 0.10
			expectedTotal:    27.50, // 0 + 25 + 2.50
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			invoice := &Invoice{
				WorkItems: tt.workItems,
				CryptoFee: tt.cryptoFee,
				TaxRate:   tt.taxRate,
			}

			err := invoice.RecalculateTotals(suite.ctx)
			require.NoError(t, err)

			assert.InDelta(t, tt.expectedSubtotal, invoice.Subtotal, 1e-9, "subtotal mismatch")
			assert.InDelta(t, tt.expectedTax, invoice.TaxAmount, 1e-9, "tax amount mismatch")
			assert.InDelta(t, tt.expectedTotal, invoice.Total, 1e-9, "total mismatch")
		})
	}
}

// TestInvoiceAddLineItem tests adding line items to invoices
func (suite *InvoiceTestSuite) TestInvoiceAddLineItem() {
	t := suite.T()
	ctx := suite.ctx

	t.Run("AddHourlyLineItem", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)

		hours := 8.0
		rate := 125.0
		lineItem := LineItem{
			ID:          "line-1",
			Type:        LineItemTypeHourly,
			Date:        time.Now(),
			Description: "Development work",
			Hours:       &hours,
			Rate:        &rate,
			Total:       1000.0,
			CreatedAt:   time.Now(),
		}

		err := invoice.AddLineItem(ctx, lineItem)
		require.NoError(t, err)
		assert.Len(t, invoice.LineItems, 1)
		assert.InDelta(t, 1000.0, invoice.Subtotal, 1e-9)
	})

	t.Run("AddFixedLineItem", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)

		amount := 2000.0
		lineItem := LineItem{
			ID:          "line-1",
			Type:        LineItemTypeFixed,
			Date:        time.Now(),
			Description: "Monthly Retainer",
			Amount:      &amount,
			Total:       2000.0,
			CreatedAt:   time.Now(),
		}

		err := invoice.AddLineItem(ctx, lineItem)
		require.NoError(t, err)
		assert.Len(t, invoice.LineItems, 1)
		assert.InDelta(t, 2000.0, invoice.Subtotal, 1e-9)
	})

	t.Run("AddQuantityLineItem", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)

		quantity := 3.0
		unitPrice := 50.0
		lineItem := LineItem{
			ID:          "line-1",
			Type:        LineItemTypeQuantity,
			Date:        time.Now(),
			Description: "SSL Certificates",
			Quantity:    &quantity,
			UnitPrice:   &unitPrice,
			Total:       150.0,
			CreatedAt:   time.Now(),
		}

		err := invoice.AddLineItem(ctx, lineItem)
		require.NoError(t, err)
		assert.Len(t, invoice.LineItems, 1)
		assert.InDelta(t, 150.0, invoice.Subtotal, 1e-9)
	})

	t.Run("AddMultipleLineItems", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)

		// Add hourly
		hours := 8.0
		rate := 125.0
		lineItem1 := LineItem{
			ID:          "line-1",
			Type:        LineItemTypeHourly,
			Date:        time.Now(),
			Description: "Development",
			Hours:       &hours,
			Rate:        &rate,
			Total:       1000.0,
			CreatedAt:   time.Now(),
		}
		err := invoice.AddLineItem(ctx, lineItem1)
		require.NoError(t, err)

		// Add fixed
		amount := 500.0
		lineItem2 := LineItem{
			ID:          "line-2",
			Type:        LineItemTypeFixed,
			Date:        time.Now(),
			Description: "Setup Fee",
			Amount:      &amount,
			Total:       500.0,
			CreatedAt:   time.Now(),
		}
		err = invoice.AddLineItem(ctx, lineItem2)
		require.NoError(t, err)

		assert.Len(t, invoice.LineItems, 2)
		assert.InDelta(t, 1500.0, invoice.Subtotal, 1e-9)
	})

	t.Run("AddInvalidLineItem", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)

		// Missing required fields
		lineItem := LineItem{
			ID:          "line-1",
			Type:        LineItemTypeHourly,
			Date:        time.Now(),
			Description: "Invalid",
			Total:       1000.0,
			CreatedAt:   time.Now(),
		}

		err := invoice.AddLineItem(ctx, lineItem)
		require.Error(t, err)
	})
}

// TestInvoiceRemoveLineItem tests removing line items from invoices
func (suite *InvoiceTestSuite) TestInvoiceRemoveLineItem() {
	t := suite.T()
	ctx := suite.ctx

	t.Run("RemoveExistingLineItem", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)

		// Add line item
		hours := 8.0
		rate := 125.0
		lineItem := LineItem{
			ID:          "line-1",
			Type:        LineItemTypeHourly,
			Date:        time.Now(),
			Description: "Development",
			Hours:       &hours,
			Rate:        &rate,
			Total:       1000.0,
			CreatedAt:   time.Now(),
		}
		err := invoice.AddLineItem(ctx, lineItem)
		require.NoError(t, err)

		// Remove it
		err = invoice.RemoveLineItem(ctx, "line-1")
		require.NoError(t, err)
		assert.Empty(t, invoice.LineItems)
		assert.InDelta(t, 0.0, invoice.Subtotal, 1e-9)
	})

	t.Run("RemoveNonExistentLineItem", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)

		err := invoice.RemoveLineItem(ctx, "non-existent")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrLineItemNotFound)
	})
}

// TestInvoiceMigrateWorkItemsToLineItems tests migration from WorkItems to LineItems
func (suite *InvoiceTestSuite) TestInvoiceMigrateWorkItemsToLineItems() {
	t := suite.T()
	ctx := suite.ctx

	t.Run("MigrateWorkItemsSuccessfully", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)

		// Add work items (old format)
		workItem := WorkItem{
			ID:          "work-1",
			Date:        time.Now(),
			Hours:       8.0,
			Rate:        125.0,
			Description: "Development",
			Total:       1000.0,
			CreatedAt:   time.Now(),
		}
		invoice.WorkItems = append(invoice.WorkItems, workItem)

		// Migrate
		err := invoice.MigrateWorkItemsToLineItems(ctx)
		require.NoError(t, err)

		assert.Len(t, invoice.LineItems, 1)
		assert.Equal(t, workItem.ID, invoice.LineItems[0].ID)
		assert.Equal(t, LineItemTypeHourly, invoice.LineItems[0].Type)
		assert.InDelta(t, workItem.Total, invoice.LineItems[0].Total, 1e-9)
	})

	t.Run("NoMigrationIfLineItemsExist", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)

		// Add work item
		workItem := WorkItem{
			ID:          "work-1",
			Date:        time.Now(),
			Hours:       8.0,
			Rate:        125.0,
			Description: "Development",
			Total:       1000.0,
			CreatedAt:   time.Now(),
		}
		invoice.WorkItems = append(invoice.WorkItems, workItem)

		// Add line item
		hours := 4.0
		rate := 150.0
		lineItem := LineItem{
			ID:          "line-1",
			Type:        LineItemTypeHourly,
			Date:        time.Now(),
			Description: "Other work",
			Hours:       &hours,
			Rate:        &rate,
			Total:       600.0,
			CreatedAt:   time.Now(),
		}
		invoice.LineItems = append(invoice.LineItems, lineItem)

		initialLineItemCount := len(invoice.LineItems)

		// Migrate - should not add anything since LineItems already exist
		err := invoice.MigrateWorkItemsToLineItems(ctx)
		require.NoError(t, err)

		assert.Len(t, invoice.LineItems, initialLineItemCount)
	})
}

// TestInvoiceRecalculateTotals tests total calculation with all item types (WorkItems and LineItems)
func (suite *InvoiceTestSuite) TestInvoiceRecalculateTotals() {
	t := suite.T()
	ctx := suite.ctx

	t.Run("CalculateWithOnlyLineItems", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)

		// Add hourly line item
		hours := 8.0
		rate := 125.0
		lineItem1 := LineItem{
			ID:          "line-1",
			Type:        LineItemTypeHourly,
			Date:        time.Now(),
			Description: "Development",
			Hours:       &hours,
			Rate:        &rate,
			Total:       1000.0,
			CreatedAt:   time.Now(),
		}
		invoice.LineItems = append(invoice.LineItems, lineItem1)

		// Add fixed line item
		amount := 500.0
		lineItem2 := LineItem{
			ID:          "line-2",
			Type:        LineItemTypeFixed,
			Date:        time.Now(),
			Description: "Setup",
			Amount:      &amount,
			Total:       500.0,
			CreatedAt:   time.Now(),
		}
		invoice.LineItems = append(invoice.LineItems, lineItem2)

		err := invoice.RecalculateTotals(ctx)
		require.NoError(t, err)

		assert.InDelta(t, 1500.0, invoice.Subtotal, 1e-9)
		assert.InDelta(t, 1500.0, invoice.Total, 1e-9)
	})

	t.Run("CalculateWithBothWorkItemsAndLineItems", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)

		// Add work item
		workItem := WorkItem{
			ID:          "work-1",
			Date:        time.Now(),
			Hours:       4.0,
			Rate:        100.0,
			Description: "Old work",
			Total:       400.0,
			CreatedAt:   time.Now(),
		}
		invoice.WorkItems = append(invoice.WorkItems, workItem)

		// Add line item
		hours := 8.0
		rate := 125.0
		lineItem := LineItem{
			ID:          "line-1",
			Type:        LineItemTypeHourly,
			Date:        time.Now(),
			Description: "New work",
			Hours:       &hours,
			Rate:        &rate,
			Total:       1000.0,
			CreatedAt:   time.Now(),
		}
		invoice.LineItems = append(invoice.LineItems, lineItem)

		err := invoice.RecalculateTotals(ctx)
		require.NoError(t, err)

		// Should include both work items and line items
		assert.InDelta(t, 1400.0, invoice.Subtotal, 1e-9)
		assert.InDelta(t, 1400.0, invoice.Total, 1e-9)
	})

	t.Run("CalculateWithCryptoFee", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)
		invoice.TaxRate = 0.005 // 0.5% fee rate

		// Add $5000 fixed line item (simulating the exact bug scenario)
		amount := 5000.0
		lineItem := LineItem{
			ID:          "line-1",
			Type:        LineItemTypeFixed,
			Date:        time.Now(),
			Description: "Repository Maintenance",
			Amount:      &amount,
			Total:       5000.0,
			CreatedAt:   time.Now(),
		}
		invoice.LineItems = append(invoice.LineItems, lineItem)

		// Set crypto fee
		invoice.CryptoFee = 25.0

		err := invoice.RecalculateTotals(ctx)
		require.NoError(t, err)

		// Should be: subtotal $5000 + crypto fee $25 = $5025 (no tax in this test, but fee is there)
		// Tax on (subtotal + crypto fee) = (5000 + 25) * 0.005 = 25.125
		// Total = subtotal + crypto fee + tax = 5000 + 25 + 25.13 = 5050.13
		assert.InDelta(t, 5000.0, invoice.Subtotal, 0.01, "Subtotal should be $5000")
		assert.InDelta(t, 25.13, invoice.TaxAmount, 0.01, "Tax on (5000+25) at 0.5% should be $25.13")
		assert.InDelta(t, 5050.13, invoice.Total, 0.01, "Total should be $5050.13")
	})

	t.Run("CalculateWithMultipleLineItemTypes", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)
		invoice.TaxRate = 0.0 // No tax to simplify

		// Add hourly line item
		hours := 10.0
		rate := 150.0
		hourlyItem := LineItem{
			ID:          "line-1",
			Type:        LineItemTypeHourly,
			Date:        time.Now(),
			Description: "Development",
			Hours:       &hours,
			Rate:        &rate,
			Total:       1500.0,
			CreatedAt:   time.Now(),
		}
		invoice.LineItems = append(invoice.LineItems, hourlyItem)

		// Add fixed line item
		fixedAmount := 2000.0
		fixedItem := LineItem{
			ID:          "line-2",
			Type:        LineItemTypeFixed,
			Date:        time.Now(),
			Description: "Setup Fee",
			Amount:      &fixedAmount,
			Total:       2000.0,
			CreatedAt:   time.Now(),
		}
		invoice.LineItems = append(invoice.LineItems, fixedItem)

		// Add quantity-based line item
		quantity := 5.0
		unitPrice := 100.0
		quantityItem := LineItem{
			ID:          "line-3",
			Type:        LineItemTypeQuantity,
			Date:        time.Now(),
			Description: "Licenses",
			Quantity:    &quantity,
			UnitPrice:   &unitPrice,
			Total:       500.0,
			CreatedAt:   time.Now(),
		}
		invoice.LineItems = append(invoice.LineItems, quantityItem)

		err := invoice.RecalculateTotals(ctx)
		require.NoError(t, err)

		// Total should be 1500 + 2000 + 500 = 4000
		assert.InDelta(t, 4000.0, invoice.Subtotal, 1e-9, "Subtotal should include all line item types")
		assert.InDelta(t, 4000.0, invoice.Total, 1e-9, "Total should equal subtotal with no tax")
	})

	t.Run("CalculateWithEmptyInvoice", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)

		err := invoice.RecalculateTotals(ctx)
		require.NoError(t, err)

		assert.InDelta(t, 0.0, invoice.Subtotal, 1e-9)
		assert.InDelta(t, 0.0, invoice.Total, 1e-9)
	})

	t.Run("CalculateWithOnlyWorkItems", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)

		// Add legacy work items only
		workItem1 := WorkItem{
			ID:          "work-1",
			Date:        time.Now(),
			Hours:       8.0,
			Rate:        100.0,
			Description: "Legacy work",
			Total:       800.0,
			CreatedAt:   time.Now(),
		}
		invoice.WorkItems = append(invoice.WorkItems, workItem1)

		workItem2 := WorkItem{
			ID:          "work-2",
			Date:        time.Now(),
			Hours:       4.0,
			Rate:        150.0,
			Description: "More legacy work",
			Total:       600.0,
			CreatedAt:   time.Now(),
		}
		invoice.WorkItems = append(invoice.WorkItems, workItem2)

		err := invoice.RecalculateTotals(ctx)
		require.NoError(t, err)

		// Should still work for backward compatibility
		assert.InDelta(t, 1400.0, invoice.Subtotal, 1e-9, "Should calculate WorkItems correctly")
		assert.InDelta(t, 1400.0, invoice.Total, 1e-9)
	})
}

// TestInvoiceSetCryptoFee tests the SetCryptoFee method with all item types
func (suite *InvoiceTestSuite) TestInvoiceSetCryptoFee() {
	t := suite.T()
	ctx := suite.ctx

	t.Run("SetCryptoFeeWithLineItems", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)
		invoice.TaxRate = 0.0 // No tax to simplify

		// Add $5000 line item (the exact bug scenario)
		amount := 5000.0
		lineItem := LineItem{
			ID:          "line-1",
			Type:        LineItemTypeFixed,
			Date:        time.Now(),
			Description: "Repository Maintenance",
			Amount:      &amount,
			Total:       5000.0,
			CreatedAt:   time.Now(),
		}
		invoice.LineItems = append(invoice.LineItems, lineItem)

		// Initially calculate without crypto fee
		err := invoice.RecalculateTotals(ctx)
		require.NoError(t, err)
		assert.InDelta(t, 5000.0, invoice.Subtotal, 0.01)
		assert.InDelta(t, 5000.0, invoice.Total, 0.01)

		// Now set crypto fee - THIS WAS THE BUG
		err = invoice.SetCryptoFee(ctx, true, true, 25.0)
		require.NoError(t, err)

		// After setting crypto fee, subtotal should STILL be $5000 (not $0!)
		assert.InDelta(t, 5000.0, invoice.Subtotal, 0.01, "Subtotal should remain $5000 after SetCryptoFee")
		assert.InDelta(t, 25.0, invoice.CryptoFee, 0.01, "Crypto fee should be $25")
		assert.InDelta(t, 5025.0, invoice.Total, 0.01, "Total should be $5025 (5000+25)")
	})

	t.Run("SetCryptoFeeWithWorkItems", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)
		invoice.TaxRate = 0.0

		// Add work item
		workItem := WorkItem{
			ID:          "work-1",
			Date:        time.Now(),
			Hours:       10.0,
			Rate:        100.0,
			Description: "Development",
			Total:       1000.0,
			CreatedAt:   time.Now(),
		}
		invoice.WorkItems = append(invoice.WorkItems, workItem)

		err := invoice.RecalculateTotals(ctx)
		require.NoError(t, err)

		// Set crypto fee
		err = invoice.SetCryptoFee(ctx, true, true, 10.0)
		require.NoError(t, err)

		assert.InDelta(t, 1000.0, invoice.Subtotal, 0.01, "Subtotal should remain $1000")
		assert.InDelta(t, 10.0, invoice.CryptoFee, 0.01)
		assert.InDelta(t, 1010.0, invoice.Total, 0.01, "Total should be $1010")
	})

	t.Run("SetCryptoFeeWithBothItemTypes", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)
		invoice.TaxRate = 0.0

		// Add work item
		workItem := WorkItem{
			ID:          "work-1",
			Date:        time.Now(),
			Hours:       5.0,
			Rate:        100.0,
			Description: "Legacy work",
			Total:       500.0,
			CreatedAt:   time.Now(),
		}
		invoice.WorkItems = append(invoice.WorkItems, workItem)

		// Add line item
		amount := 1000.0
		lineItem := LineItem{
			ID:          "line-1",
			Type:        LineItemTypeFixed,
			Date:        time.Now(),
			Description: "New work",
			Amount:      &amount,
			Total:       1000.0,
			CreatedAt:   time.Now(),
		}
		invoice.LineItems = append(invoice.LineItems, lineItem)

		err := invoice.RecalculateTotals(ctx)
		require.NoError(t, err)
		assert.InDelta(t, 1500.0, invoice.Subtotal, 0.01)

		// Set crypto fee
		err = invoice.SetCryptoFee(ctx, true, true, 15.0)
		require.NoError(t, err)

		// Must include BOTH WorkItems and LineItems
		assert.InDelta(t, 1500.0, invoice.Subtotal, 0.01, "Subtotal should include both item types")
		assert.InDelta(t, 15.0, invoice.CryptoFee, 0.01)
		assert.InDelta(t, 1515.0, invoice.Total, 0.01, "Total should be $1515")
	})

	t.Run("DisableCryptoFee", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)
		invoice.TaxRate = 0.0

		amount := 1000.0
		lineItem := LineItem{
			ID:          "line-1",
			Type:        LineItemTypeFixed,
			Date:        time.Now(),
			Description: "Work",
			Amount:      &amount,
			Total:       1000.0,
			CreatedAt:   time.Now(),
		}
		invoice.LineItems = append(invoice.LineItems, lineItem)

		// Set crypto fee first
		err := invoice.SetCryptoFee(ctx, true, true, 10.0)
		require.NoError(t, err)
		assert.InDelta(t, 10.0, invoice.CryptoFee, 0.01)

		// Now disable it
		err = invoice.SetCryptoFee(ctx, false, false, 0.0)
		require.NoError(t, err)

		assert.InDelta(t, 1000.0, invoice.Subtotal, 0.01)
		assert.InDelta(t, 0.0, invoice.CryptoFee, 0.01, "Crypto fee should be zero when disabled")
		assert.InDelta(t, 1000.0, invoice.Total, 0.01)
	})

	t.Run("SetCryptoFeeWithTax", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)
		invoice.TaxRate = 0.005 // 0.5% tax

		amount := 5000.0
		lineItem := LineItem{
			ID:          "line-1",
			Type:        LineItemTypeFixed,
			Date:        time.Now(),
			Description: "Repository Maintenance",
			Amount:      &amount,
			Total:       5000.0,
			CreatedAt:   time.Now(),
		}
		invoice.LineItems = append(invoice.LineItems, lineItem)

		// Set crypto fee
		err := invoice.SetCryptoFee(ctx, true, true, 25.0)
		require.NoError(t, err)

		// Tax is calculated on (subtotal + crypto fee)
		// Tax = (5000 + 25) * 0.005 = 25.125 rounded to 25.13
		// Total = 5000 + 25 + 25.13 = 5050.13
		assert.InDelta(t, 5000.0, invoice.Subtotal, 0.01)
		assert.InDelta(t, 25.0, invoice.CryptoFee, 0.01)
		assert.InDelta(t, 25.13, invoice.TaxAmount, 0.01, "Tax should be calculated on subtotal+fee")
		assert.InDelta(t, 5050.13, invoice.Total, 0.01)
	})
}

// TestInvoiceHelperMethods tests helper methods for line items
func (suite *InvoiceTestSuite) TestInvoiceHelperMethods() {
	t := suite.T()
	ctx := suite.ctx

	t.Run("HasOnlyWorkItems", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)
		workItem := WorkItem{
			ID:          "work-1",
			Date:        time.Now(),
			Hours:       8.0,
			Rate:        125.0,
			Description: "Development",
			Total:       1000.0,
			CreatedAt:   time.Now(),
		}
		invoice.WorkItems = append(invoice.WorkItems, workItem)

		assert.True(t, invoice.HasOnlyWorkItems())
		assert.False(t, invoice.HasLineItems())
	})

	t.Run("HasLineItems", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)
		hours := 8.0
		rate := 125.0
		lineItem := LineItem{
			ID:          "line-1",
			Type:        LineItemTypeHourly,
			Date:        time.Now(),
			Description: "Development",
			Hours:       &hours,
			Rate:        &rate,
			Total:       1000.0,
			CreatedAt:   time.Now(),
		}
		invoice.LineItems = append(invoice.LineItems, lineItem)

		assert.False(t, invoice.HasOnlyWorkItems())
		assert.True(t, invoice.HasLineItems())
	})

	t.Run("GetAllItems", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)

		// Add work item
		workItem := WorkItem{
			ID:          "work-1",
			Date:        time.Now(),
			Hours:       4.0,
			Rate:        100.0,
			Description: "Old work",
			Total:       400.0,
			CreatedAt:   time.Now(),
		}
		invoice.WorkItems = append(invoice.WorkItems, workItem)

		// Add line items
		hours := 8.0
		rate := 125.0
		lineItem := LineItem{
			ID:          "line-1",
			Type:        LineItemTypeHourly,
			Date:        time.Now(),
			Description: "New work",
			Hours:       &hours,
			Rate:        &rate,
			Total:       1000.0,
			CreatedAt:   time.Now(),
		}
		invoice.LineItems = append(invoice.LineItems, lineItem)

		allItems := invoice.GetAllItems()
		assert.Len(t, allItems, 2)
	})

	t.Run("TotalHours", func(t *testing.T) {
		invoice := createTestInvoice(t, ctx)

		// Add work item
		workItem := WorkItem{
			ID:          "work-1",
			Date:        time.Now(),
			Hours:       4.0,
			Rate:        100.0,
			Description: "Old work",
			Total:       400.0,
			CreatedAt:   time.Now(),
		}
		invoice.WorkItems = append(invoice.WorkItems, workItem)

		// Add hourly line item
		hours := 8.0
		rate := 125.0
		lineItem1 := LineItem{
			ID:          "line-1",
			Type:        LineItemTypeHourly,
			Date:        time.Now(),
			Description: "New work",
			Hours:       &hours,
			Rate:        &rate,
			Total:       1000.0,
			CreatedAt:   time.Now(),
		}
		invoice.LineItems = append(invoice.LineItems, lineItem1)

		// Add fixed line item (should not count towards hours)
		amount := 500.0
		lineItem2 := LineItem{
			ID:          "line-2",
			Type:        LineItemTypeFixed,
			Date:        time.Now(),
			Description: "Setup",
			Amount:      &amount,
			Total:       500.0,
			CreatedAt:   time.Now(),
		}
		invoice.LineItems = append(invoice.LineItems, lineItem2)

		totalHours := invoice.TotalHours()
		assert.InDelta(t, 12.0, totalHours, 1e-9) // 4 from work item + 8 from hourly line item
	})
}

// TestInvoiceCryptoAddressOverride tests the crypto address override functionality
func (suite *InvoiceTestSuite) TestInvoiceCryptoAddressOverride() {
	defaultUSDCAddress := "0xDefaultUSDCAddress123456789"
	defaultBSVAddress := "DefaultBSVAddress123456789"
	customUSDCAddress := "0xCustomUSDCAddress987654321"
	customBSVAddress := "CustomBSVAddress987654321"

	suite.Run("NoOverride_UsesDefault", func() {
		t := suite.T()
		invoice := createTestInvoice(t, suite.ctx)

		// No override set
		assert.Nil(t, invoice.USDCAddressOverride)
		assert.Nil(t, invoice.BSVAddressOverride)

		// Should return default addresses
		assert.Equal(t, defaultUSDCAddress, invoice.GetUSDCAddress(defaultUSDCAddress))
		assert.Equal(t, defaultBSVAddress, invoice.GetBSVAddress(defaultBSVAddress))

		// Should not have overrides
		assert.False(t, invoice.HasUSDCAddressOverride())
		assert.False(t, invoice.HasBSVAddressOverride())
	})

	suite.Run("WithOverride_UsesCustomAddress", func() {
		t := suite.T()
		invoice := createTestInvoice(t, suite.ctx)

		// Set overrides
		invoice.USDCAddressOverride = &customUSDCAddress
		invoice.BSVAddressOverride = &customBSVAddress

		// Should return custom addresses
		assert.Equal(t, customUSDCAddress, invoice.GetUSDCAddress(defaultUSDCAddress))
		assert.Equal(t, customBSVAddress, invoice.GetBSVAddress(defaultBSVAddress))

		// Should have overrides
		assert.True(t, invoice.HasUSDCAddressOverride())
		assert.True(t, invoice.HasBSVAddressOverride())
	})

	suite.Run("EmptyStringOverride_UsesEmptyNotDefault", func() {
		t := suite.T()
		invoice := createTestInvoice(t, suite.ctx)

		// Set empty string overrides (different from nil)
		emptyString := ""
		invoice.USDCAddressOverride = &emptyString
		invoice.BSVAddressOverride = &emptyString

		// Should return empty strings, not defaults
		assert.Empty(t, invoice.GetUSDCAddress(defaultUSDCAddress))
		assert.Empty(t, invoice.GetBSVAddress(defaultBSVAddress))

		// Should not have overrides (empty string is not considered an override)
		assert.False(t, invoice.HasUSDCAddressOverride())
		assert.False(t, invoice.HasBSVAddressOverride())
	})

	suite.Run("MixedOverride_OnlyUSDC", func() {
		t := suite.T()
		invoice := createTestInvoice(t, suite.ctx)

		// Set only USDC override
		invoice.USDCAddressOverride = &customUSDCAddress

		// USDC should use custom, BSV should use default
		assert.Equal(t, customUSDCAddress, invoice.GetUSDCAddress(defaultUSDCAddress))
		assert.Equal(t, defaultBSVAddress, invoice.GetBSVAddress(defaultBSVAddress))

		assert.True(t, invoice.HasUSDCAddressOverride())
		assert.False(t, invoice.HasBSVAddressOverride())
	})

	suite.Run("MixedOverride_OnlyBSV", func() {
		t := suite.T()
		invoice := createTestInvoice(t, suite.ctx)

		// Set only BSV override
		invoice.BSVAddressOverride = &customBSVAddress

		// USDC should use default, BSV should use custom
		assert.Equal(t, defaultUSDCAddress, invoice.GetUSDCAddress(defaultUSDCAddress))
		assert.Equal(t, customBSVAddress, invoice.GetBSVAddress(defaultBSVAddress))

		assert.False(t, invoice.HasUSDCAddressOverride())
		assert.True(t, invoice.HasBSVAddressOverride())
	})

	suite.Run("EmptyDefaultWithOverride", func() {
		t := suite.T()
		invoice := createTestInvoice(t, suite.ctx)

		// Set custom override
		invoice.USDCAddressOverride = &customUSDCAddress

		// Should use custom even when default is empty
		assert.Equal(t, customUSDCAddress, invoice.GetUSDCAddress(""))
		assert.True(t, invoice.HasUSDCAddressOverride())
	})

	suite.Run("EmptyDefaultNoOverride", func() {
		t := suite.T()
		invoice := createTestInvoice(t, suite.ctx)

		// No override set
		assert.Nil(t, invoice.USDCAddressOverride)

		// Should return empty default
		assert.Empty(t, invoice.GetUSDCAddress(""))
		assert.False(t, invoice.HasUSDCAddressOverride())
	})
}
