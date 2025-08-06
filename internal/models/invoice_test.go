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
