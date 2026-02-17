package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-invoice/internal/models"
)

// MockLogger implements the Logger interface for testing
type MockLogger struct {
	logs []LogEntry
}

type LogEntry struct {
	Level   string
	Message string
	Fields  []interface{}
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.logs = append(m.logs, LogEntry{Level: "debug", Message: msg, Fields: fields})
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.logs = append(m.logs, LogEntry{Level: "info", Message: msg, Fields: fields})
}

func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.logs = append(m.logs, LogEntry{Level: "warn", Message: msg, Fields: fields})
}

func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.logs = append(m.logs, LogEntry{Level: "error", Message: msg, Fields: fields})
}

func (m *MockLogger) GetLogs() []LogEntry {
	return m.logs
}

func (m *MockLogger) Reset() {
	m.logs = nil
}

// CalculatorTestSuite defines a test suite for the InvoiceCalculator
type CalculatorTestSuite struct {
	suite.Suite

	calculator *InvoiceCalculator
	logger     *MockLogger
}

// SetupTest runs before each test method
func (suite *CalculatorTestSuite) SetupTest() {
	suite.logger = &MockLogger{}
	suite.calculator = NewInvoiceCalculator(suite.logger)
}

// TestCalculateWorkItemTotal tests individual work item calculation
func (suite *CalculatorTestSuite) TestCalculateWorkItemTotal() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		hours       float64
		rate        float64
		expected    float64
		expectError bool
	}{
		{
			name:     "ValidCalculation",
			hours:    8.0,
			rate:     125.00,
			expected: 1000.00,
		},
		{
			name:     "PartialHours",
			hours:    7.5,
			rate:     100.00,
			expected: 750.00,
		},
		{
			name:     "ZeroHours",
			hours:    0.0,
			rate:     125.00,
			expected: 0.00,
		},
		{
			name:     "ZeroRate",
			hours:    8.0,
			rate:     0.0,
			expected: 0.00,
		},
		{
			name:        "NegativeHours",
			hours:       -1.0,
			rate:        125.00,
			expectError: true,
		},
		{
			name:        "NegativeRate",
			hours:       8.0,
			rate:        -125.00,
			expectError: true,
		},
		{
			name:     "FloatingPointPrecision",
			hours:    1.5,
			rate:     33.33,
			expected: 50.00, // Rounded from 49.995
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			total, err := suite.calculator.CalculateWorkItemTotal(ctx, tc.hours, tc.rate)

			if tc.expectError {
				suite.Error(err)
			} else {
				suite.Require().NoError(err)
				suite.InDelta(tc.expected, total, 0.01, "Expected %.2f, got %.2f", tc.expected, total)
			}
		})
	}
}

// TestCalculateInvoiceTotals tests comprehensive invoice calculation
func (suite *CalculatorTestSuite) TestCalculateInvoiceTotals() {
	ctx := context.Background()

	// Create test invoice
	invoice := suite.createTestInvoice()

	testCases := []struct {
		name             string
		options          *CalculationOptions
		expectedSubtotal float64
		expectedTax      float64
		expectedTotal    float64
		expectError      bool
	}{
		{
			name: "NoTax",
			options: &CalculationOptions{
				TaxRate:       0.0,
				Currency:      "USD",
				DecimalPlaces: 2,
				RoundingMode:  "round",
			},
			expectedSubtotal: 2312.50,
			expectedTax:      0.00,
			expectedTotal:    2312.50,
		},
		{
			name: "WithTax",
			options: &CalculationOptions{
				TaxRate:       0.10,
				Currency:      "USD",
				DecimalPlaces: 2,
				RoundingMode:  "round",
			},
			expectedSubtotal: 2312.50,
			expectedTax:      231.25,
			expectedTotal:    2543.75,
		},
		{
			name: "HighTaxRate",
			options: &CalculationOptions{
				TaxRate:       0.25,
				Currency:      "EUR",
				DecimalPlaces: 2,
				RoundingMode:  "round",
			},
			expectedSubtotal: 2312.50,
			expectedTax:      578.13,
			expectedTotal:    2890.63,
		},
		{
			name: "RoundingModeFloor",
			options: &CalculationOptions{
				TaxRate:       0.10,
				Currency:      "USD",
				DecimalPlaces: 2,
				RoundingMode:  "floor",
			},
			expectedSubtotal: 2312.50,
			expectedTax:      231.25,
			expectedTotal:    2543.75,
		},
		{
			name: "HighTaxRate",
			options: &CalculationOptions{
				TaxRate:       1.5, // > 1.0 but valid for demonstration
				Currency:      "USD",
				DecimalPlaces: 2,
				RoundingMode:  "round",
			},
			expectedSubtotal: 2312.50,
			expectedTax:      3468.75, // 2312.50 * 1.5
			expectedTotal:    5781.25,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result, err := suite.calculator.CalculateInvoiceTotals(ctx, invoice, tc.options)

			if tc.expectError {
				suite.Error(err)
			} else {
				suite.Require().NoError(err)
				suite.NotNil(result)
				suite.InDelta(tc.expectedSubtotal, result.Subtotal, 0.01)
				suite.InDelta(tc.expectedTax, result.TaxAmount, 0.01)
				suite.InDelta(tc.expectedTotal, result.Total, 0.01)
				suite.Equal(3, result.WorkItemCount, "Should have 3 work items (no line items in this test)")
				suite.InDelta(18.5, result.TotalHours, 0.01)
			}
		})
	}
}

// TestCalculateInvoiceTotalsWithBreakdown tests calculation with detailed breakdown
func (suite *CalculatorTestSuite) TestCalculateInvoiceTotalsWithBreakdown() {
	ctx := context.Background()
	invoice := suite.createTestInvoice()

	options := &CalculationOptions{
		TaxRate:          0.10,
		Currency:         "USD",
		DecimalPlaces:    2,
		RoundingMode:     "round",
		IncludeBreakdown: true,
		TaxType:          "VAT",
	}

	result, err := suite.calculator.CalculateInvoiceTotals(ctx, invoice, options)

	suite.Require().NoError(err)
	suite.NotNil(result.Breakdown)
	suite.Len(result.Breakdown.WorkItemTotals, 3)
	suite.Equal("VAT", result.Breakdown.TaxCalculation.TaxType)
	suite.Equal("USD", result.Breakdown.CurrencyDetails.Currency)
	suite.Equal("$", result.Breakdown.CurrencyDetails.Symbol)
}

// TestValidateCalculation tests input validation
func (suite *CalculatorTestSuite) TestValidateCalculation() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		invoice     *models.Invoice
		options     *CalculationOptions
		expectError bool
	}{
		{
			name:    "ValidInput",
			invoice: suite.createTestInvoice(),
			options: &CalculationOptions{
				TaxRate:       0.10,
				Currency:      "USD",
				DecimalPlaces: 2,
				RoundingMode:  "round",
			},
		},
		{
			name:        "NilInvoice",
			invoice:     nil,
			options:     &CalculationOptions{},
			expectError: true,
		},
		{
			name:        "NilOptions",
			invoice:     suite.createTestInvoice(),
			options:     nil,
			expectError: true,
		},
		{
			name:    "InvalidTaxRate",
			invoice: suite.createTestInvoice(),
			options: &CalculationOptions{
				TaxRate:       -0.1,
				Currency:      "USD",
				DecimalPlaces: 2,
				RoundingMode:  "round",
			},
			expectError: true,
		},
		{
			name:    "InvalidDecimalPlaces",
			invoice: suite.createTestInvoice(),
			options: &CalculationOptions{
				TaxRate:       0.10,
				Currency:      "USD",
				DecimalPlaces: -1,
				RoundingMode:  "round",
			},
			expectError: true,
		},
		{
			name:    "InvalidRoundingMode",
			invoice: suite.createTestInvoice(),
			options: &CalculationOptions{
				TaxRate:       0.10,
				Currency:      "USD",
				DecimalPlaces: 2,
				RoundingMode:  "invalid",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.calculator.ValidateCalculation(ctx, tc.invoice, tc.options)

			if tc.expectError {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

// TestRecalculateInvoice tests invoice recalculation
func (suite *CalculatorTestSuite) TestRecalculateInvoice() {
	ctx := context.Background()
	invoice := suite.createTestInvoice()

	// Store original values
	originalTotal := invoice.Total
	originalVersion := invoice.Version

	err := suite.calculator.RecalculateInvoice(ctx, invoice, 0.15)

	suite.Require().NoError(err)
	suite.NotEqual(originalTotal, invoice.Total)
	suite.Equal(originalVersion+1, invoice.Version)
	suite.InDelta(2659.38, invoice.Total, 0.01) // 2312.50 + (2312.50 * 0.15)
}

// TestGetCalculationSummary tests summary calculation for multiple invoices
func (suite *CalculatorTestSuite) TestGetCalculationSummary() {
	ctx := context.Background()

	testCases := []struct {
		name          string
		invoices      []*models.Invoice
		expectedCount int
		expectedTotal float64
		expectedHours float64
		expectError   bool
	}{
		{
			name:          "SingleInvoice",
			invoices:      []*models.Invoice{suite.createTestInvoice()},
			expectedCount: 1,
			expectedTotal: 2312.50,
			expectedHours: 18.5,
		},
		{
			name:          "MultipleInvoices",
			invoices:      []*models.Invoice{suite.createTestInvoice(), suite.createTestInvoice()},
			expectedCount: 2,
			expectedTotal: 4625.00,
			expectedHours: 37.0,
		},
		{
			name:          "EmptyInvoices",
			invoices:      []*models.Invoice{},
			expectedCount: 0,
			expectedTotal: 0.0,
			expectedHours: 0.0,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			summary, err := suite.calculator.GetCalculationSummary(ctx, tc.invoices)

			if tc.expectError {
				suite.Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Equal(tc.expectedCount, summary.InvoiceCount)
				suite.InDelta(tc.expectedTotal, summary.TotalSubtotal, 0.01)
				suite.InDelta(tc.expectedHours, summary.TotalHours, 0.01)
			}
		})
	}
}

// TestContextCancellation tests context cancellation handling
func (suite *CalculatorTestSuite) TestContextCancellation() {
	// Create a canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	invoice := suite.createTestInvoice()
	options := &CalculationOptions{
		TaxRate:       0.10,
		Currency:      "USD",
		DecimalPlaces: 2,
		RoundingMode:  "round",
	}

	// Test that canceled context is handled properly
	_, err := suite.calculator.CalculateInvoiceTotals(ctx, invoice, options)
	suite.Require().Error(err)
	suite.Equal(context.Canceled, err)
}

// TestRoundAmount tests the roundAmount method with different rounding modes
func (suite *CalculatorTestSuite) TestRoundAmount() {
	testCases := []struct {
		name     string
		amount   float64
		options  *CalculationOptions
		expected float64
	}{
		{
			name:     "NilOptions",
			amount:   123.456,
			options:  nil,
			expected: 123.46, // Default 2 decimal places, round mode
		},
		{
			name:   "NegativeDecimalPlaces",
			amount: 123.456,
			options: &CalculationOptions{
				DecimalPlaces: -1,
				RoundingMode:  "round",
			},
			expected: 123.46, // Default 2 decimal places
		},
		{
			name:   "FloorRounding",
			amount: 123.456,
			options: &CalculationOptions{
				DecimalPlaces: 2,
				RoundingMode:  "floor",
			},
			expected: 123.45,
		},
		{
			name:   "CeilRounding",
			amount: 123.456,
			options: &CalculationOptions{
				DecimalPlaces: 2,
				RoundingMode:  "ceil",
			},
			expected: 123.46,
		},
		{
			name:   "DefaultRounding",
			amount: 123.456,
			options: &CalculationOptions{
				DecimalPlaces: 2,
				RoundingMode:  "round",
			},
			expected: 123.46,
		},
		{
			name:   "UnknownRoundingMode",
			amount: 123.456,
			options: &CalculationOptions{
				DecimalPlaces: 2,
				RoundingMode:  "unknown",
			},
			expected: 123.46, // Defaults to round
		},
		{
			name:   "ThreeDecimalPlaces",
			amount: 123.4567,
			options: &CalculationOptions{
				DecimalPlaces: 3,
				RoundingMode:  "round",
			},
			expected: 123.457,
		},
		{
			name:   "ZeroDecimalPlaces",
			amount: 123.456,
			options: &CalculationOptions{
				DecimalPlaces: 0,
				RoundingMode:  "round",
			},
			expected: 123.0,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result := suite.calculator.roundAmount(tc.amount, tc.options)
			suite.InDelta(tc.expected, result, 0.001, "Expected %.3f, got %.3f", tc.expected, result)
		})
	}
}

// TestGetCurrencySymbol tests the getCurrencySymbol method
func (suite *CalculatorTestSuite) TestGetCurrencySymbol() {
	testCases := []struct {
		name           string
		currency       string
		expectedSymbol string
	}{
		{"USD", "USD", "$"},
		{"EUR", "EUR", "€"},
		{"GBP", "GBP", "£"},
		{"CAD", "CAD", "C$"},
		{"AUD", "AUD", "A$"},
		{"JPY", "JPY", "¥"},
		{"CHF", "CHF", "CHF"},
		{"SEK", "SEK", "kr"},
		{"NOK", "NOK", "kr"},
		{"DKK", "DKK", "kr"},
		{"UnknownCurrency", "XYZ", "XYZ"}, // Should return currency code
		{"EmptyString", "", ""},           // Should return empty string
		{"Lowercase", "usd", "usd"},       // Case sensitive, should return currency
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result := suite.calculator.getCurrencySymbol(tc.currency)
			suite.Equal(tc.expectedSymbol, result)
		})
	}
}

// TestRecalculateInvoiceErrorCases tests error scenarios for RecalculateInvoice
func (suite *CalculatorTestSuite) TestRecalculateInvoiceErrorCases() {
	ctx := context.Background()

	suite.Run("NilInvoice", func() {
		err := suite.calculator.RecalculateInvoice(ctx, nil, 0.10)
		suite.Require().Error(err)
		suite.Contains(err.Error(), "invoice cannot be nil")
	})

	suite.Run("NegativeTaxRate", func() {
		invoice := suite.createTestInvoice()
		// RecalculateInvoice allows negative tax rates but calculateTax returns 0 for taxRate <= 0
		err := suite.calculator.RecalculateInvoice(ctx, invoice, -0.10)
		suite.Require().NoError(err)
		// Verify that negative tax rate is stored but tax amount is 0
		suite.InDelta(-0.10, invoice.TaxRate, 1e-9)
		suite.InDelta(0.0, invoice.TaxAmount, 1e-9) // calculateTax returns 0 for negative rates
	})

	suite.Run("CanceledContext", func() {
		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		invoice := suite.createTestInvoice()
		err := suite.calculator.RecalculateInvoice(canceledCtx, invoice, 0.10)
		suite.Require().Error(err)
		suite.Equal(context.Canceled, err)
	})

	suite.Run("InvoiceWithInvalidWorkItems", func() {
		invoice := suite.createTestInvoice()
		// Add an invalid work item
		invoice.WorkItems = append(invoice.WorkItems, models.WorkItem{
			ID:    "invalid",
			Hours: -1.0, // Negative hours should cause error
			Rate:  125.0,
		})

		err := suite.calculator.RecalculateInvoice(ctx, invoice, 0.10)
		suite.Require().Error(err)
		suite.Contains(err.Error(), "failed to calculate work item totals")
		suite.Contains(err.Error(), "hours cannot be negative")
	})
}

// TestCalculateWorkItemTotalsErrorCases tests error scenarios for calculateWorkItemTotals
func (suite *CalculatorTestSuite) TestCalculateWorkItemTotalsErrorCases() {
	ctx := context.Background()
	options := &CalculationOptions{
		TaxRate:          0.10,
		Currency:         "USD",
		DecimalPlaces:    2,
		RoundingMode:     "round",
		IncludeBreakdown: false,
	}

	suite.Run("EmptyWorkItems", func() {
		subtotal, hours, breakdown, err := suite.calculator.calculateWorkItemTotals(ctx, []models.WorkItem{}, options)
		suite.Require().NoError(err)
		suite.InDelta(0.0, subtotal, 0.01)
		suite.InDelta(0.0, hours, 0.01)
		suite.Empty(breakdown)
	})

	suite.Run("WorkItemWithNegativeHours", func() {
		workItems := []models.WorkItem{
			{
				ID:    "invalid",
				Hours: -1.0,
				Rate:  125.0,
			},
		}

		subtotal, hours, breakdown, err := suite.calculator.calculateWorkItemTotals(ctx, workItems, options)
		suite.Require().Error(err)
		suite.Contains(err.Error(), "failed to calculate work item")
		suite.InDelta(0.0, subtotal, 0.01)
		suite.InDelta(0.0, hours, 0.01)
		suite.Nil(breakdown)
	})

	suite.Run("WorkItemWithNegativeRate", func() {
		workItems := []models.WorkItem{
			{
				ID:    "invalid",
				Hours: 8.0,
				Rate:  -125.0,
			},
		}

		subtotal, hours, breakdown, err := suite.calculator.calculateWorkItemTotals(ctx, workItems, options)
		suite.Require().Error(err)
		suite.Contains(err.Error(), "failed to calculate work item")
		suite.InDelta(0.0, subtotal, 0.01)
		suite.InDelta(0.0, hours, 0.01)
		suite.Nil(breakdown)
	})

	suite.Run("CanceledContext", func() {
		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		workItems := []models.WorkItem{
			{
				ID:    "valid",
				Hours: 8.0,
				Rate:  125.0,
			},
		}

		subtotal, hours, breakdown, err := suite.calculator.calculateWorkItemTotals(canceledCtx, workItems, options)
		suite.Require().Error(err)
		suite.Equal(context.Canceled, err)
		suite.InDelta(0.0, subtotal, 0.01)
		suite.InDelta(0.0, hours, 0.01)
		suite.Nil(breakdown)
	})

	suite.Run("WithBreakdownEnabled", func() {
		optionsWithBreakdown := &CalculationOptions{
			TaxRate:          0.10,
			Currency:         "USD",
			DecimalPlaces:    2,
			RoundingMode:     "round",
			IncludeBreakdown: true,
		}

		workItems := []models.WorkItem{
			{
				ID:    "valid1",
				Hours: 8.0,
				Rate:  125.0,
			},
		}

		subtotal, hours, breakdown, err := suite.calculator.calculateWorkItemTotals(ctx, workItems, optionsWithBreakdown)
		suite.Require().NoError(err)
		suite.InDelta(1000.0, subtotal, 0.01)
		suite.InDelta(8.0, hours, 0.01)
		suite.Len(breakdown, 1)
		suite.Equal("valid1", breakdown[0].ID)
		suite.InDelta(1000.0, breakdown[0].Subtotal, 0.01)
	})
}

// TestCalculateTaxEdgeCases tests edge cases for calculateTax method
func (suite *CalculatorTestSuite) TestCalculateTaxEdgeCases() {
	testCases := []struct {
		name     string
		subtotal float64
		options  *CalculationOptions
		expected float64
	}{
		{
			name:     "ZeroSubtotal",
			subtotal: 0.0,
			options: &CalculationOptions{
				TaxRate:       0.10,
				DecimalPlaces: 2,
				RoundingMode:  "round",
			},
			expected: 0.0,
		},
		{
			name:     "ZeroTaxRate",
			subtotal: 1000.0,
			options: &CalculationOptions{
				TaxRate:       0.0,
				DecimalPlaces: 2,
				RoundingMode:  "round",
			},
			expected: 0.0,
		},
		{
			name:     "NilOptions",
			subtotal: 1000.0,
			options:  nil,
			expected: 0.0, // Default behavior with nil options
		},
		{
			name:     "VerySmallSubtotal",
			subtotal: 0.01,
			options: &CalculationOptions{
				TaxRate:       0.10,
				DecimalPlaces: 2,
				RoundingMode:  "round",
			},
			expected: 0.00, // Rounds to 0
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			taxRate := 0.0
			if tc.options != nil {
				taxRate = tc.options.TaxRate
			}
			result := suite.calculator.calculateTax(tc.subtotal, taxRate, tc.options)
			suite.InDelta(tc.expected, result, 0.001, "Expected %.3f, got %.3f", tc.expected, result)
		})
	}
}

// Helper methods

func (suite *CalculatorTestSuite) createTestInvoice() *models.Invoice {
	workItems := []models.WorkItem{
		{
			ID:          "work_001",
			Date:        time.Now().AddDate(0, 0, -7),
			Hours:       8.0,
			Rate:        125.00,
			Description: "Web development",
			Total:       1000.00,
		},
		{
			ID:          "work_002",
			Date:        time.Now().AddDate(0, 0, -6),
			Hours:       6.5,
			Rate:        125.00,
			Description: "Database optimization",
			Total:       812.50,
		},
		{
			ID:          "work_003",
			Date:        time.Now().AddDate(0, 0, -5),
			Hours:       4.0,
			Rate:        125.00,
			Description: "Code review",
			Total:       500.00,
		},
	}

	client := models.Client{
		ID:     models.ClientID("test_client"),
		Name:   "Test Client",
		Email:  "test@example.com",
		Active: true,
	}

	return &models.Invoice{
		ID:        models.InvoiceID("test_invoice"),
		Number:    "TEST-001",
		Date:      time.Now().AddDate(0, 0, -1),
		DueDate:   time.Now().AddDate(0, 0, 30),
		Client:    client,
		WorkItems: workItems,
		Status:    models.StatusDraft,
		Subtotal:  2312.50,
		TaxRate:   0.10,
		TaxAmount: 231.25,
		Total:     2543.75,
		CreatedAt: time.Now().AddDate(0, 0, -1),
		UpdatedAt: time.Now(),
		Version:   1,
	}
}

// TestCalculateInvoiceTotalsWithLineItems tests that LineItems are included in subtotal calculation
func (suite *CalculatorTestSuite) TestCalculateInvoiceTotalsWithLineItems() {
	ctx := context.Background()

	suite.Run("InvoiceWithOnlyLineItems", func() {
		// Create an invoice with ONLY line items (no work items)
		amount := 5000.0
		invoice := &models.Invoice{
			ID:     models.InvoiceID("test_line_items"),
			Number: "TEST-LINEITEMS-001",
			Date:   time.Now(),
			LineItems: []models.LineItem{
				{
					ID:          "line_001",
					Type:        models.LineItemTypeFixed,
					Date:        time.Now(),
					Description: "Repository Maintenance",
					Amount:      &amount,
					Total:       5000.0,
				},
			},
			WorkItems: []models.WorkItem{}, // No work items
			Status:    models.StatusDraft,
			Version:   1,
		}

		options := &CalculationOptions{
			TaxRate:       0.0,
			Currency:      "USD",
			DecimalPlaces: 2,
			RoundingMode:  "round",
		}

		result, err := suite.calculator.CalculateInvoiceTotals(ctx, invoice, options)

		suite.Require().NoError(err)
		suite.NotNil(result)
		suite.InDelta(5000.0, result.Subtotal, 0.01, "Subtotal should include LineItems")
		suite.InDelta(0.0, result.TaxAmount, 0.01)
		suite.InDelta(5000.0, result.Total, 0.01)
		suite.Equal(1, result.WorkItemCount, "WorkItemCount should include LineItems")
	})

	suite.Run("InvoiceWithLineItemsAndCryptoFee", func() {
		// Test the specific bug scenario: $5000 line item + $25 crypto fee = $5025 total
		amount := 5000.0
		invoice := &models.Invoice{
			ID:     models.InvoiceID("test_crypto_fee"),
			Number: "TEST-CRYPTO-001",
			Date:   time.Now(),
			LineItems: []models.LineItem{
				{
					ID:          "line_001",
					Type:        models.LineItemTypeFixed,
					Date:        time.Now(),
					Description: "Repository Maintenance",
					Amount:      &amount,
					Total:       5000.0,
				},
			},
			WorkItems: []models.WorkItem{},
			Status:    models.StatusDraft,
			Version:   1,
		}

		// Crypto fee is 0.5% of subtotal, which would be $25 for $5000
		options := &CalculationOptions{
			TaxRate:       0.005, // 0.5% fee
			Currency:      "USD",
			DecimalPlaces: 2,
			RoundingMode:  "round",
		}

		result, err := suite.calculator.CalculateInvoiceTotals(ctx, invoice, options)

		suite.Require().NoError(err)
		suite.NotNil(result)
		suite.InDelta(5000.0, result.Subtotal, 0.01, "Subtotal should be $5000")
		suite.InDelta(25.0, result.TaxAmount, 0.01, "Fee should be $25")
		suite.InDelta(5025.0, result.Total, 0.01, "Total should be $5025")
	})

	suite.Run("InvoiceWithBothWorkItemsAndLineItems", func() {
		// Create an invoice with BOTH work items AND line items
		amount := 5000.0
		invoice := &models.Invoice{
			ID:     models.InvoiceID("test_mixed"),
			Number: "TEST-MIXED-001",
			Date:   time.Now(),
			LineItems: []models.LineItem{
				{
					ID:          "line_001",
					Type:        models.LineItemTypeFixed,
					Date:        time.Now(),
					Description: "Repository Maintenance",
					Amount:      &amount,
					Total:       5000.0,
				},
			},
			WorkItems: []models.WorkItem{
				{
					ID:          "work_001",
					Date:        time.Now(),
					Hours:       8.0,
					Rate:        125.0,
					Description: "Development work",
					Total:       1000.0,
				},
			},
			Status:  models.StatusDraft,
			Version: 1,
		}

		options := &CalculationOptions{
			TaxRate:       0.0,
			Currency:      "USD",
			DecimalPlaces: 2,
			RoundingMode:  "round",
		}

		result, err := suite.calculator.CalculateInvoiceTotals(ctx, invoice, options)

		suite.Require().NoError(err)
		suite.NotNil(result)
		suite.InDelta(6000.0, result.Subtotal, 0.01, "Subtotal should include both WorkItems ($1000) and LineItems ($5000)")
		suite.InDelta(0.0, result.TaxAmount, 0.01)
		suite.InDelta(6000.0, result.Total, 0.01)
		suite.Equal(2, result.WorkItemCount, "WorkItemCount should include both types")
		suite.InDelta(8.0, result.TotalHours, 0.01, "TotalHours should only count WorkItems")
	})

	suite.Run("InvoiceWithMultipleLineItemTypes", func() {
		// Test different types of line items: fixed, hourly, quantity
		fixedAmount := 5000.0
		hourlyRate := 125.0
		hourlyHours := 8.0
		quantity := 10.0
		unitPrice := 50.0

		invoice := &models.Invoice{
			ID:     models.InvoiceID("test_multiple_types"),
			Number: "TEST-TYPES-001",
			Date:   time.Now(),
			LineItems: []models.LineItem{
				{
					ID:          "line_fixed",
					Type:        models.LineItemTypeFixed,
					Date:        time.Now(),
					Description: "Fixed amount",
					Amount:      &fixedAmount,
					Total:       5000.0,
				},
				{
					ID:          "line_hourly",
					Type:        models.LineItemTypeHourly,
					Date:        time.Now(),
					Description: "Hourly work",
					Quantity:    &hourlyHours,
					UnitPrice:   &hourlyRate,
					Total:       1000.0,
				},
				{
					ID:          "line_quantity",
					Type:        models.LineItemTypeQuantity,
					Date:        time.Now(),
					Description: "Quantity-based",
					Quantity:    &quantity,
					UnitPrice:   &unitPrice,
					Total:       500.0,
				},
			},
			WorkItems: []models.WorkItem{},
			Status:    models.StatusDraft,
			Version:   1,
		}

		options := &CalculationOptions{
			TaxRate:       0.1, // 10% tax
			Currency:      "USD",
			DecimalPlaces: 2,
			RoundingMode:  "round",
		}

		result, err := suite.calculator.CalculateInvoiceTotals(ctx, invoice, options)

		suite.Require().NoError(err)
		suite.NotNil(result)
		suite.InDelta(6500.0, result.Subtotal, 0.01, "Subtotal should include all LineItem types: $5000 + $1000 + $500")
		suite.InDelta(650.0, result.TaxAmount, 0.01, "Tax should be 10% of $6500")
		suite.InDelta(7150.0, result.Total, 0.01, "Total should be $6500 + $650")
		suite.Equal(3, result.WorkItemCount, "Should count all 3 LineItems")
	})

	suite.Run("EmptyInvoice", func() {
		// Test invoice with no work items or line items
		invoice := &models.Invoice{
			ID:        models.InvoiceID("test_empty"),
			Number:    "TEST-EMPTY-001",
			Date:      time.Now(),
			LineItems: []models.LineItem{},
			WorkItems: []models.WorkItem{},
			Status:    models.StatusDraft,
			Version:   1,
		}

		options := &CalculationOptions{
			TaxRate:       0.1,
			Currency:      "USD",
			DecimalPlaces: 2,
			RoundingMode:  "round",
		}

		result, err := suite.calculator.CalculateInvoiceTotals(ctx, invoice, options)

		suite.Require().NoError(err)
		suite.NotNil(result)
		suite.InDelta(0.0, result.Subtotal, 0.01, "Empty invoice should have $0 subtotal")
		suite.InDelta(0.0, result.TaxAmount, 0.01)
		suite.InDelta(0.0, result.Total, 0.01)
		suite.Equal(0, result.WorkItemCount)
	})
}

// Run the test suite
func TestCalculatorTestSuite(t *testing.T) {
	suite.Run(t, new(CalculatorTestSuite))
}

// Benchmark tests for performance verification

func BenchmarkCalculateWorkItemTotal(b *testing.B) {
	logger := &MockLogger{}
	calculator := NewInvoiceCalculator(logger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = calculator.CalculateWorkItemTotal(ctx, 8.0, 125.00)
	}
}

func BenchmarkCalculateInvoiceTotals(b *testing.B) {
	logger := &MockLogger{}
	calculator := NewInvoiceCalculator(logger)
	ctx := context.Background()

	// Create test data
	suite := &CalculatorTestSuite{}
	suite.logger = logger
	suite.calculator = calculator
	invoice := suite.createTestInvoice()

	options := &CalculationOptions{
		TaxRate:       0.10,
		Currency:      "USD",
		DecimalPlaces: 2,
		RoundingMode:  "round",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = calculator.CalculateInvoiceTotals(ctx, invoice, options)
	}
}
