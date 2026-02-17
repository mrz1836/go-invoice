package main

import (
	"testing"
	"time"

	"github.com/mrz1836/go-invoice/internal/cli"
	"github.com/mrz1836/go-invoice/internal/config"
	"github.com/mrz1836/go-invoice/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildInvoiceRecalculateCommand(t *testing.T) {
	app := &App{
		logger: cli.NewLogger(false),
	}

	cmd := app.buildInvoiceRecalculateCommand()

	// Test command basic properties
	assert.Equal(t, "recalculate", cmd.Use[:11], "Command Use should start with 'recalculate'")
	assert.Equal(t, "Recalculate invoice totals", cmd.Short)
	assert.NotEmpty(t, cmd.Long, "Command should have long description")
	assert.NotEmpty(t, cmd.Example, "Command should have examples")

	// Test that command requires exactly 1 argument
	assert.NotNil(t, cmd.Args, "Command should have Args validator")

	// Verify RunE is set
	assert.NotNil(t, cmd.RunE, "Command should have RunE function")
}

func TestCreateInvoiceData(t *testing.T) {
	app := &App{
		logger: cli.NewLogger(false),
	}

	cfg := &config.Config{
		Business: config.BusinessConfig{
			Name:    "Test Business",
			Address: "123 Test St",
			Phone:   "555-1234",
			Email:   "test@example.com",
		},
		Invoice: config.InvoiceConfig{
			Currency: "USD",
		},
	}

	t.Run("TotalHoursWithOnlyWorkItems", func(t *testing.T) {
		invoice := &models.Invoice{
			ID:     "test-001",
			Number: "TEST-001",
			Date:   time.Now(),
			WorkItems: []models.WorkItem{
				{
					ID:          "work-1",
					Date:        time.Now(),
					Hours:       8.0,
					Rate:        100.0,
					Description: "Development",
					Total:       800.0,
					CreatedAt:   time.Now(),
				},
				{
					ID:          "work-2",
					Date:        time.Now(),
					Hours:       4.0,
					Rate:        150.0,
					Description: "Consulting",
					Total:       600.0,
					CreatedAt:   time.Now(),
				},
			},
			LineItems: []models.LineItem{},
		}

		data := app.createInvoiceData(invoice, cfg)

		assert.InDelta(t, 12.0, data.TotalHours, 0.01, "Should count hours from WorkItems")
	})

	t.Run("TotalHoursWithOnlyLineItems", func(t *testing.T) {
		hours1 := 10.0
		rate1 := 125.0
		hours2 := 5.0
		rate2 := 150.0

		invoice := &models.Invoice{
			ID:        "test-002",
			Number:    "TEST-002",
			Date:      time.Now(),
			WorkItems: []models.WorkItem{},
			LineItems: []models.LineItem{
				{
					ID:          "line-1",
					Type:        models.LineItemTypeHourly,
					Date:        time.Now(),
					Description: "Development",
					Hours:       &hours1,
					Rate:        &rate1,
					Total:       1250.0,
					CreatedAt:   time.Now(),
				},
				{
					ID:          "line-2",
					Type:        models.LineItemTypeHourly,
					Date:        time.Now(),
					Description: "Consulting",
					Hours:       &hours2,
					Rate:        &rate2,
					Total:       750.0,
					CreatedAt:   time.Now(),
				},
			},
		}

		data := app.createInvoiceData(invoice, cfg)

		assert.InDelta(t, 15.0, data.TotalHours, 0.01, "Should count hours from hourly LineItems")
	})

	t.Run("TotalHoursWithBothItemTypes", func(t *testing.T) {
		hours := 8.0
		rate := 125.0

		invoice := &models.Invoice{
			ID:     "test-003",
			Number: "TEST-003",
			Date:   time.Now(),
			WorkItems: []models.WorkItem{
				{
					ID:          "work-1",
					Date:        time.Now(),
					Hours:       5.0,
					Rate:        100.0,
					Description: "Legacy work",
					Total:       500.0,
					CreatedAt:   time.Now(),
				},
			},
			LineItems: []models.LineItem{
				{
					ID:          "line-1",
					Type:        models.LineItemTypeHourly,
					Date:        time.Now(),
					Description: "New work",
					Hours:       &hours,
					Rate:        &rate,
					Total:       1000.0,
					CreatedAt:   time.Now(),
				},
			},
		}

		data := app.createInvoiceData(invoice, cfg)

		assert.InDelta(t, 13.0, data.TotalHours, 0.01, "Should count hours from both WorkItems and LineItems")
	})

	t.Run("TotalHoursWithMixedLineItemTypes", func(t *testing.T) {
		hours := 10.0
		rate := 125.0
		fixedAmount := 1000.0

		invoice := &models.Invoice{
			ID:        "test-004",
			Number:    "TEST-004",
			Date:      time.Now(),
			WorkItems: []models.WorkItem{},
			LineItems: []models.LineItem{
				{
					ID:          "line-1",
					Type:        models.LineItemTypeHourly,
					Date:        time.Now(),
					Description: "Hourly work",
					Hours:       &hours,
					Rate:        &rate,
					Total:       1250.0,
					CreatedAt:   time.Now(),
				},
				{
					ID:          "line-2",
					Type:        models.LineItemTypeFixed,
					Date:        time.Now(),
					Description: "Fixed price work",
					Amount:      &fixedAmount,
					Total:       1000.0,
					CreatedAt:   time.Now(),
				},
			},
		}

		data := app.createInvoiceData(invoice, cfg)

		// Should only count hours from hourly items, not fixed
		assert.InDelta(t, 10.0, data.TotalHours, 0.01, "Should only count hours from hourly LineItems, not fixed")
	})

	t.Run("TotalHoursWithQuantityLineItems", func(t *testing.T) {
		hours := 8.0
		rate := 125.0
		quantity := 5.0
		unitPrice := 100.0

		invoice := &models.Invoice{
			ID:        "test-005",
			Number:    "TEST-005",
			Date:      time.Now(),
			WorkItems: []models.WorkItem{},
			LineItems: []models.LineItem{
				{
					ID:          "line-1",
					Type:        models.LineItemTypeHourly,
					Date:        time.Now(),
					Description: "Hourly work",
					Hours:       &hours,
					Rate:        &rate,
					Total:       1000.0,
					CreatedAt:   time.Now(),
				},
				{
					ID:          "line-2",
					Type:        models.LineItemTypeQuantity,
					Date:        time.Now(),
					Description: "Licenses",
					Quantity:    &quantity,
					UnitPrice:   &unitPrice,
					Total:       500.0,
					CreatedAt:   time.Now(),
				},
			},
		}

		data := app.createInvoiceData(invoice, cfg)

		// Should only count hours from hourly items, not quantity items
		assert.InDelta(t, 8.0, data.TotalHours, 0.01, "Should only count hours from hourly LineItems")
	})

	t.Run("InvoiceDataStructure", func(t *testing.T) {
		invoice := &models.Invoice{
			ID:        "test-006",
			Number:    "TEST-006",
			Date:      time.Now(),
			Subtotal:  5000.0,
			Total:     5025.0,
			CryptoFee: 25.0,
			WorkItems: []models.WorkItem{},
			LineItems: []models.LineItem{},
		}

		data := app.createInvoiceData(invoice, cfg)

		require.NotNil(t, data, "Invoice data should not be nil")
		assert.Equal(t, invoice.Number, data.Number, "Invoice should be embedded")
		assert.InDelta(t, invoice.Subtotal, data.Subtotal, 0.01, "Subtotal should be preserved")
		assert.InDelta(t, invoice.Total, data.Total, 0.01, "Total should be preserved")
		assert.InDelta(t, invoice.CryptoFee, data.CryptoFee, 0.01, "CryptoFee should be preserved")
		assert.Equal(t, cfg.Business.Name, data.Business.Name, "Business info should be populated")
		assert.Equal(t, "USD", data.Config.Currency, "Config should be populated")
		assert.Equal(t, "$", data.Config.CurrencySymbol, "Currency symbol should be set")
	})
}
