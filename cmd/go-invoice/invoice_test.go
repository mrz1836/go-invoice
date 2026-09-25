package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-invoice/internal/cli"
	"github.com/mrz1836/go-invoice/internal/config"
	"github.com/mrz1836/go-invoice/internal/models"
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

	t.Run("WirePaymentDetailsPropagated", func(t *testing.T) {
		wireCfg := *cfg
		wireCfg.Business.PaymentInstructions = "Include the invoice number with your payment"
		wireCfg.Business.DomesticWire = config.DomesticWire{
			Enabled:         true,
			BeneficiaryName: "Acme Corp",
			BankName:        "Example Bank",
			AccountNumber:   "0000123456",
			RoutingNumber:   "123456789",
			AccountType:     "Checking",
			Reference:       "Invoice number",
		}
		wireCfg.Business.InternationalWire = config.InternationalWire{
			Enabled:            true,
			BeneficiaryName:    "Acme Corp",
			BeneficiaryAddress: "1 Example Way, Springfield, US",
			BankName:           "Example Bank",
			BankAddress:        "2 Example Plaza, Springfield, US",
			SWIFT:              "TESTUS33",
			IBAN:               "GB00TEST00000000000000",
			Reference:          "Invoice number",
		}

		data := app.createInvoiceData(&models.Invoice{ID: "wire-007", Number: "WIRE-007"}, &wireCfg)

		assert.Equal(t, wireCfg.Business.PaymentInstructions, data.Business.PaymentInstructions)
		assert.Equal(t, wireCfg.Business.DomesticWire, data.Business.DomesticWire)
		assert.Equal(t, wireCfg.Business.InternationalWire, data.Business.InternationalWire)
	})
}

func TestRenderDefaultTemplatePaymentSection(t *testing.T) {
	app := &App{
		logger: cli.NewLogger(false),
	}

	baseConfig := func() *config.Config {
		return &config.Config{
			Business: config.BusinessConfig{
				Name:         "Test Business",
				Address:      "123 Test St",
				Email:        "test@example.com",
				PaymentTerms: "Net 30",
			},
			Invoice: config.InvoiceConfig{
				Currency: "USD",
			},
		}
	}

	renderHTML := func(t *testing.T, cfg *config.Config, invoice *models.Invoice) string {
		t.Helper()
		ctx := context.Background()
		renderer, err := app.createRenderService(ctx, cfg)
		require.NoError(t, err)

		html, err := app.renderInvoice(ctx, renderer, app.createInvoiceData(invoice, cfg), "default")
		require.NoError(t, err)
		return html
	}

	newInvoice := func() *models.Invoice {
		return &models.Invoice{
			ID:      "render-001",
			Number:  "RENDER-001",
			Date:    time.Now(),
			DueDate: time.Now().AddDate(0, 0, 30),
			Client:  models.Client{ID: "client-001", Name: "Acme Corp"},
		}
	}

	t.Run("NoPaymentMethodsOmitsHeading", func(t *testing.T) {
		html := renderHTML(t, baseConfig(), newInvoice())

		assert.Contains(t, html, "Payment Terms")
		assert.NotContains(t, html, "Payment Methods:")
		assert.NotContains(t, html, "ACH")
	})

	t.Run("RenderedMethodEmitsHeadingOnce", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Business.CryptoPayments.USDCEnabled = true
		cfg.Business.CryptoPayments.USDCAddress = "0x0000000000000000000000000000000000000001"

		html := renderHTML(t, cfg, newInvoice())

		assert.Equal(t, 1, strings.Count(html, "Payment Methods:"))
		assert.Contains(t, html, "USDC Cryptocurrency:")
	})

	t.Run("EnabledMethodWithoutAddressOmitsHeading", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Business.CryptoPayments.USDCEnabled = true

		html := renderHTML(t, cfg, newInvoice())

		assert.NotContains(t, html, "Payment Methods:")
	})

	t.Run("PaymentInstructionsRendered", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Business.PaymentInstructions = "Include the invoice number with your payment"

		html := renderHTML(t, cfg, newInvoice())

		assert.Contains(t, html, "Include the invoice number with your payment")
	})

	t.Run("CryptoFeeNoticeRecommendsWireTransfer", func(t *testing.T) {
		invoice := newInvoice()
		invoice.CryptoFee = 25.0

		html := renderHTML(t, baseConfig(), invoice)

		assert.Contains(t, html, "please pay by bank wire transfer (USD)")
		assert.NotContains(t, html, "ACH Bank Transfer")
	})
}

func TestRenderDefaultTemplateWireTransfer(t *testing.T) {
	app := &App{
		logger: cli.NewLogger(false),
	}

	// wireConfig enables both wire sets so each test proves the client's wire type drives selection
	wireConfig := func() *config.Config {
		return &config.Config{
			Business: config.BusinessConfig{
				Name:         "Test Business",
				Address:      "123 Test St",
				Email:        "test@example.com",
				PaymentTerms: "Net 30",
				DomesticWire: config.DomesticWire{
					Enabled:         true,
					BeneficiaryName: "Acme Corp",
					BankName:        "Example Bank",
					AccountNumber:   "0000123456",
					RoutingNumber:   "123456789",
					AccountType:     "Checking",
					Reference:       "Domestic memo",
				},
				InternationalWire: config.InternationalWire{
					Enabled:            true,
					BeneficiaryName:    "Acme Corp",
					BeneficiaryAddress: "1 Example Way, Springfield, US",
					BankName:           "Example Bank",
					BankAddress:        "2 Example Plaza, Springfield, US",
					SWIFT:              "TESTUS33",
					IBAN:               "GB00TEST00000000000000",
					IntermediaryBank:   "Example Correspondent Bank",
					IntermediarySWIFT:  "CORRUS33",
					Reference:          "International memo",
				},
			},
			Invoice: config.InvoiceConfig{
				Currency: "USD",
			},
		}
	}

	renderHTML := func(t *testing.T, cfg *config.Config, invoice *models.Invoice) string {
		t.Helper()
		ctx := context.Background()
		renderer, err := app.createRenderService(ctx, cfg)
		require.NoError(t, err)

		html, err := app.renderInvoice(ctx, renderer, app.createInvoiceData(invoice, cfg), "default")
		require.NoError(t, err)
		return html
	}

	newInvoice := func(wireType models.WireType) *models.Invoice {
		return &models.Invoice{
			ID:      "render-wire-001",
			Number:  "RENDER-WIRE-001",
			Date:    time.Now(),
			DueDate: time.Now().AddDate(0, 0, 30),
			Client:  models.Client{ID: "client-001", Name: "Acme Corp", WireType: wireType},
			WorkItems: []models.WorkItem{
				{ID: "work-001", Date: time.Now(), Hours: 1, Rate: 100, Description: "Consulting", Total: 100},
			},
			Subtotal: 100,
			Total:    100,
		}
	}

	domesticOnly := []string{"Domestic Wire Transfer (USD):", "ABA Routing Number: 123456789", "Account Number: 0000123456", "Domestic memo"}
	internationalOnly := []string{"International Wire Transfer (USD):", "SWIFT/BIC: TESTUS33", "IBAN: GB00TEST00000000000000", "International memo"}

	t.Run("WireFeeRendersTotalsLineAndNotice", func(t *testing.T) {
		invoice := newInvoice(models.WireTypeInternational)
		invoice.WireFee = 20.0
		invoice.Total = 120.0

		html := renderHTML(t, wireConfig(), invoice)

		assert.Contains(t, html, "Wire Transfer Service Fee:")
		assert.Contains(t, html, "Wire Transfer Service Fee Notice:")
		assert.Contains(t, html, "A $20.00 service fee has been applied for bank wire transfer processing.")
		assert.Contains(t, html, "$120.00")
		assert.NotContains(t, html, "Cryptocurrency Service Fee")
	})

	t.Run("NoWireFeeOmitsTotalsLineAndNotice", func(t *testing.T) {
		html := renderHTML(t, wireConfig(), newInvoice(models.WireTypeInternational))

		assert.NotContains(t, html, "Wire Transfer Service Fee")
	})

	t.Run("BothFeesRenderTogether", func(t *testing.T) {
		invoice := newInvoice(models.WireTypeInternational)
		invoice.CryptoFee = 25.0
		invoice.WireFee = 20.0
		invoice.Total = 145.0

		html := renderHTML(t, wireConfig(), invoice)

		assert.Contains(t, html, "Cryptocurrency Service Fee:")
		assert.Contains(t, html, "Wire Transfer Service Fee:")
		assert.Less(t, strings.Index(html, "Cryptocurrency Service Fee:"), strings.Index(html, "Wire Transfer Service Fee:"))
	})

	t.Run("DomesticClientRendersOnlyDomesticSet", func(t *testing.T) {
		html := renderHTML(t, wireConfig(), newInvoice(models.WireTypeDomestic))

		for _, want := range domesticOnly {
			assert.Contains(t, html, want)
		}
		assert.Contains(t, html, "Beneficiary: Acme Corp")
		assert.Contains(t, html, "Bank: Example Bank")
		assert.Contains(t, html, "Account Type: Checking")
		for _, absent := range internationalOnly {
			assert.NotContains(t, html, absent)
		}
		assert.Equal(t, 1, strings.Count(html, "Payment Methods:"))
		assert.NotContains(t, html, "ACH Bank Transfer")
	})

	t.Run("InternationalClientRendersOnlyInternationalSet", func(t *testing.T) {
		html := renderHTML(t, wireConfig(), newInvoice(models.WireTypeInternational))

		for _, want := range internationalOnly {
			assert.Contains(t, html, want)
		}
		assert.Contains(t, html, "Beneficiary: Acme Corp")
		assert.Contains(t, html, "Beneficiary Address: 1 Example Way, Springfield, US")
		assert.Contains(t, html, "Beneficiary Bank: Example Bank")
		assert.Contains(t, html, "Bank Address: 2 Example Plaza, Springfield, US")
		assert.Contains(t, html, "Intermediary Bank: Example Correspondent Bank")
		assert.Contains(t, html, "Intermediary SWIFT/BIC: CORRUS33")
		for _, absent := range domesticOnly {
			assert.NotContains(t, html, absent)
		}
		assert.Equal(t, 1, strings.Count(html, "Payment Methods:"))
		assert.NotContains(t, html, "ACH Bank Transfer")
	})

	t.Run("InternationalOptionalFieldsOmittedWhenEmpty", func(t *testing.T) {
		cfg := wireConfig()
		cfg.Business.InternationalWire.IBAN = ""
		cfg.Business.InternationalWire.AccountNumber = "9876543210"
		cfg.Business.InternationalWire.IntermediaryBank = ""
		cfg.Business.InternationalWire.IntermediarySWIFT = ""

		html := renderHTML(t, cfg, newInvoice(models.WireTypeInternational))

		assert.Contains(t, html, "Account Number: 9876543210")
		assert.NotContains(t, html, "IBAN:")
		assert.NotContains(t, html, "Intermediary")
	})

	t.Run("NoWireTypeRendersNeitherSet", func(t *testing.T) {
		for _, wireType := range []models.WireType{models.WireTypeNone, ""} {
			html := renderHTML(t, wireConfig(), newInvoice(wireType))

			for _, absent := range append(append([]string{}, domesticOnly...), internationalOnly...) {
				assert.NotContains(t, html, absent, "wire type %q", wireType)
			}
			assert.NotContains(t, html, "Payment Methods:", "wire type %q", wireType)
		}
	})

	t.Run("SelectedSetDisabledRendersNothing", func(t *testing.T) {
		cfg := wireConfig()
		cfg.Business.DomesticWire.Enabled = false

		html := renderHTML(t, cfg, newInvoice(models.WireTypeDomestic))

		for _, absent := range append(append([]string{}, domesticOnly...), internationalOnly...) {
			assert.NotContains(t, html, absent)
		}
		assert.NotContains(t, html, "Payment Methods:")
	})

	t.Run("WireBlockSharesHeadingWithOtherMethods", func(t *testing.T) {
		cfg := wireConfig()
		cfg.Business.CryptoPayments.USDCEnabled = true
		cfg.Business.CryptoPayments.USDCAddress = "0x0000000000000000000000000000000000000001"

		html := renderHTML(t, cfg, newInvoice(models.WireTypeInternational))

		assert.Equal(t, 1, strings.Count(html, "Payment Methods:"))
		assert.Less(t, strings.Index(html, "Payment Methods:"), strings.Index(html, "International Wire Transfer (USD):"))
		assert.Less(t, strings.Index(html, "International Wire Transfer (USD):"), strings.Index(html, "USDC Cryptocurrency:"))
	})
}
