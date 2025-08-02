package services

import (
	"context"
	"testing"
	"time"

	"github.com/mrz/go-invoice/internal/models"
	"github.com/stretchr/testify/suite"
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
				suite.Equal(3, result.WorkItemCount)
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
	suite.Equal(3, len(result.Breakdown.WorkItemTotals))
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
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	invoice := suite.createTestInvoice()
	options := &CalculationOptions{
		TaxRate:       0.10,
		Currency:      "USD",
		DecimalPlaces: 2,
		RoundingMode:  "round",
	}

	// Test that cancelled context is handled properly
	_, err := suite.calculator.CalculateInvoiceTotals(ctx, invoice, options)
	suite.Error(err)
	suite.Equal(context.Canceled, err)
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
