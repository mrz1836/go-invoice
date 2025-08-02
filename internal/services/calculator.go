package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/mrz/go-invoice/internal/models"
)

// InvoiceCalculator provides calculation services for invoices
// Follows dependency injection pattern with consumer-driven interfaces
type InvoiceCalculator struct {
	logger Logger
}

// NewInvoiceCalculator creates a new invoice calculator with injected dependencies
func NewInvoiceCalculator(logger Logger) *InvoiceCalculator {
	return &InvoiceCalculator{
		logger: logger,
	}
}

// CalculationResult represents the result of invoice calculations
type CalculationResult struct {
	Subtotal          float64               `json:"subtotal"`
	TaxAmount         float64               `json:"tax_amount"`
	Total             float64               `json:"total"`
	TaxRate           float64               `json:"tax_rate"`
	WorkItemCount     int                   `json:"work_item_count"`
	TotalHours        float64               `json:"total_hours"`
	AverageHourlyRate float64               `json:"average_hourly_rate"`
	CalculatedAt      time.Time             `json:"calculated_at"`
	Breakdown         *CalculationBreakdown `json:"breakdown,omitempty"`
}

// CalculationBreakdown provides detailed breakdown of calculations
type CalculationBreakdown struct {
	WorkItemTotals     []WorkItemCalculation `json:"work_item_totals"`
	TaxCalculation     TaxCalculation        `json:"tax_calculation"`
	RoundingAdjustment float64               `json:"rounding_adjustment"`
	CurrencyDetails    CurrencyDetails       `json:"currency_details"`
}

// WorkItemCalculation represents calculation details for a single work item
type WorkItemCalculation struct {
	ID          string    `json:"id"`
	Date        time.Time `json:"date"`
	Hours       float64   `json:"hours"`
	Rate        float64   `json:"rate"`
	Subtotal    float64   `json:"subtotal"`
	Description string    `json:"description"`
}

// TaxCalculation represents tax calculation details
type TaxCalculation struct {
	TaxableAmount float64 `json:"taxable_amount"`
	TaxRate       float64 `json:"tax_rate"`
	TaxAmount     float64 `json:"tax_amount"`
	TaxType       string  `json:"tax_type"` // "VAT", "GST", "Sales", etc.
	RoundedAmount float64 `json:"rounded_amount"`
}

// CurrencyDetails represents currency-specific calculation details
type CurrencyDetails struct {
	Currency      string `json:"currency"`
	Symbol        string `json:"symbol"`
	DecimalPlaces int    `json:"decimal_places"`
	RoundingMode  string `json:"rounding_mode"`
}

// CalculationOptions represents options for invoice calculations
type CalculationOptions struct {
	TaxRate          float64 `json:"tax_rate"`
	Currency         string  `json:"currency"`
	DecimalPlaces    int     `json:"decimal_places"`
	RoundingMode     string  `json:"rounding_mode"` // "round", "floor", "ceil"
	IncludeBreakdown bool    `json:"include_breakdown"`
	TaxType          string  `json:"tax_type"`
}

// CalculateInvoiceTotals calculates all totals for an invoice
func (c *InvoiceCalculator) CalculateInvoiceTotals(ctx context.Context, invoice *models.Invoice, options *CalculationOptions) (*CalculationResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if invoice == nil {
		return nil, fmt.Errorf("invoice cannot be nil")
	}

	if options == nil {
		options = &CalculationOptions{
			TaxRate:          0.0,
			Currency:         "USD",
			DecimalPlaces:    2,
			RoundingMode:     "round",
			IncludeBreakdown: false,
			TaxType:          "VAT",
		}
	}

	start := time.Now()
	c.logger.Debug("calculating invoice totals", "invoice_id", invoice.ID, "work_items", len(invoice.WorkItems))

	// Calculate subtotal from work items
	subtotal, totalHours, workItemCalcs, err := c.calculateWorkItemTotals(ctx, invoice.WorkItems, options)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate work item totals: %w", err)
	}

	// Calculate tax
	taxAmount := c.calculateTax(subtotal, options.TaxRate, options)

	// Calculate total
	total := subtotal + taxAmount

	// Apply rounding
	subtotal = c.roundAmount(subtotal, options)
	taxAmount = c.roundAmount(taxAmount, options)
	total = c.roundAmount(total, options)

	// Calculate average hourly rate
	averageRate := 0.0
	if totalHours > 0 {
		averageRate = subtotal / totalHours
		averageRate = c.roundAmount(averageRate, options)
	}

	result := &CalculationResult{
		Subtotal:          subtotal,
		TaxAmount:         taxAmount,
		Total:             total,
		TaxRate:           options.TaxRate,
		WorkItemCount:     len(invoice.WorkItems),
		TotalHours:        totalHours,
		AverageHourlyRate: averageRate,
		CalculatedAt:      time.Now(),
	}

	// Add breakdown if requested
	if options.IncludeBreakdown {
		result.Breakdown = &CalculationBreakdown{
			WorkItemTotals: workItemCalcs,
			TaxCalculation: TaxCalculation{
				TaxableAmount: subtotal,
				TaxRate:       options.TaxRate,
				TaxAmount:     taxAmount,
				TaxType:       options.TaxType,
				RoundedAmount: taxAmount,
			},
			RoundingAdjustment: 0.0, // Could be calculated if needed
			CurrencyDetails: CurrencyDetails{
				Currency:      options.Currency,
				Symbol:        c.getCurrencySymbol(options.Currency),
				DecimalPlaces: options.DecimalPlaces,
				RoundingMode:  options.RoundingMode,
			},
		}
	}

	calcTime := time.Since(start)
	c.logger.Info("invoice totals calculated",
		"invoice_id", invoice.ID,
		"subtotal", subtotal,
		"tax", taxAmount,
		"total", total,
		"calc_time_ms", calcTime.Milliseconds())

	return result, nil
}

// CalculateWorkItemTotal calculates the total for a single work item
func (c *InvoiceCalculator) CalculateWorkItemTotal(ctx context.Context, hours, rate float64) (float64, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	if hours < 0 {
		return 0, fmt.Errorf("hours cannot be negative: %f", hours)
	}

	if rate < 0 {
		return 0, fmt.Errorf("rate cannot be negative: %f", rate)
	}

	total := hours * rate

	// Round to 2 decimal places to avoid floating point precision issues
	total = math.Round(total*100) / 100

	c.logger.Debug("work item total calculated", "hours", hours, "rate", rate, "total", total)
	return total, nil
}

// ValidateCalculation validates calculation inputs
func (c *InvoiceCalculator) ValidateCalculation(ctx context.Context, invoice *models.Invoice, options *CalculationOptions) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if invoice == nil {
		return fmt.Errorf("invoice cannot be nil")
	}

	if options == nil {
		return fmt.Errorf("calculation options cannot be nil")
	}

	// Validate tax rate
	if options.TaxRate < 0 || options.TaxRate > 1 {
		return fmt.Errorf("tax rate must be between 0 and 1, got: %f", options.TaxRate)
	}

	// Validate decimal places
	if options.DecimalPlaces < 0 || options.DecimalPlaces > 10 {
		return fmt.Errorf("decimal places must be between 0 and 10, got: %d", options.DecimalPlaces)
	}

	// Validate rounding mode
	validRoundingModes := []string{"round", "floor", "ceil"}
	validMode := false
	for _, mode := range validRoundingModes {
		if options.RoundingMode == mode {
			validMode = true
			break
		}
	}
	if !validMode {
		return fmt.Errorf("invalid rounding mode '%s', must be one of: round, floor, ceil", options.RoundingMode)
	}

	// Validate work items
	for i, item := range invoice.WorkItems {
		if item.Hours < 0 {
			return fmt.Errorf("work item %d has negative hours: %f", i, item.Hours)
		}
		if item.Rate < 0 {
			return fmt.Errorf("work item %d has negative rate: %f", i, item.Rate)
		}
	}

	c.logger.Debug("calculation validation passed", "invoice_id", invoice.ID)
	return nil
}

// RecalculateInvoice recalculates and updates invoice totals
func (c *InvoiceCalculator) RecalculateInvoice(ctx context.Context, invoice *models.Invoice, taxRate float64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if invoice == nil {
		return fmt.Errorf("invoice cannot be nil")
	}

	c.logger.Debug("recalculating invoice", "invoice_id", invoice.ID, "tax_rate", taxRate)

	options := &CalculationOptions{
		TaxRate:       taxRate,
		Currency:      "USD", // Default currency
		DecimalPlaces: 2,
		RoundingMode:  "round",
	}

	result, err := c.CalculateInvoiceTotals(ctx, invoice, options)
	if err != nil {
		return fmt.Errorf("failed to calculate invoice totals: %w", err)
	}

	// Update invoice with calculated values
	invoice.Subtotal = result.Subtotal
	invoice.TaxRate = taxRate
	invoice.TaxAmount = result.TaxAmount
	invoice.Total = result.Total
	invoice.UpdatedAt = time.Now()
	invoice.Version++

	c.logger.Info("invoice recalculated", "invoice_id", invoice.ID, "total", invoice.Total)
	return nil
}

// GetCalculationSummary returns a summary of calculations for reporting
func (c *InvoiceCalculator) GetCalculationSummary(ctx context.Context, invoices []*models.Invoice) (*CalculationSummary, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if len(invoices) == 0 {
		return &CalculationSummary{}, nil
	}

	c.logger.Debug("calculating summary", "invoice_count", len(invoices))

	summary := &CalculationSummary{
		InvoiceCount: len(invoices),
		CalculatedAt: time.Now(),
	}

	var totalSubtotal, totalTax, totalAmount, totalHours float64

	for _, invoice := range invoices {
		totalSubtotal += invoice.Subtotal
		totalTax += invoice.TaxAmount
		totalAmount += invoice.Total

		for _, item := range invoice.WorkItems {
			totalHours += item.Hours
		}
	}

	summary.TotalSubtotal = c.roundAmount(totalSubtotal, &CalculationOptions{DecimalPlaces: 2})
	summary.TotalTax = c.roundAmount(totalTax, &CalculationOptions{DecimalPlaces: 2})
	summary.TotalAmount = c.roundAmount(totalAmount, &CalculationOptions{DecimalPlaces: 2})
	summary.TotalHours = c.roundAmount(totalHours, &CalculationOptions{DecimalPlaces: 2})

	if totalHours > 0 {
		summary.AverageRate = c.roundAmount(totalSubtotal/totalHours, &CalculationOptions{DecimalPlaces: 2})
	}

	if len(invoices) > 0 {
		summary.AverageInvoiceAmount = c.roundAmount(totalAmount/float64(len(invoices)), &CalculationOptions{DecimalPlaces: 2})
	}

	c.logger.Info("calculation summary completed", "invoices", len(invoices), "total", summary.TotalAmount)
	return summary, nil
}

// Helper methods

// calculateWorkItemTotals calculates totals for all work items
func (c *InvoiceCalculator) calculateWorkItemTotals(ctx context.Context, workItems []models.WorkItem, options *CalculationOptions) (float64, float64, []WorkItemCalculation, error) {
	var subtotal, totalHours float64
	var workItemCalcs []WorkItemCalculation

	for _, item := range workItems {
		select {
		case <-ctx.Done():
			return 0, 0, nil, ctx.Err()
		default:
		}

		itemTotal, err := c.CalculateWorkItemTotal(ctx, item.Hours, item.Rate)
		if err != nil {
			return 0, 0, nil, fmt.Errorf("failed to calculate work item %s total: %w", item.ID, err)
		}

		subtotal += itemTotal
		totalHours += item.Hours

		if options.IncludeBreakdown {
			workItemCalcs = append(workItemCalcs, WorkItemCalculation{
				ID:          item.ID,
				Date:        item.Date,
				Hours:       item.Hours,
				Rate:        item.Rate,
				Subtotal:    itemTotal,
				Description: item.Description,
			})
		}
	}

	return subtotal, totalHours, workItemCalcs, nil
}

// calculateTax calculates tax amount based on subtotal and rate
func (c *InvoiceCalculator) calculateTax(subtotal, taxRate float64, options *CalculationOptions) float64 {
	if taxRate <= 0 {
		return 0
	}

	taxAmount := subtotal * taxRate
	return c.roundAmount(taxAmount, options)
}

// roundAmount rounds an amount based on the specified options
func (c *InvoiceCalculator) roundAmount(amount float64, options *CalculationOptions) float64 {
	if options == nil || options.DecimalPlaces < 0 {
		return math.Round(amount*100) / 100
	}

	multiplier := math.Pow(10, float64(options.DecimalPlaces))

	switch options.RoundingMode {
	case "floor":
		return math.Floor(amount*multiplier) / multiplier
	case "ceil":
		return math.Ceil(amount*multiplier) / multiplier
	default: // "round"
		return math.Round(amount*multiplier) / multiplier
	}
}

// getCurrencySymbol returns the symbol for a currency code
func (c *InvoiceCalculator) getCurrencySymbol(currency string) string {
	symbols := map[string]string{
		"USD": "$",
		"EUR": "€",
		"GBP": "£",
		"CAD": "C$",
		"AUD": "A$",
		"JPY": "¥",
		"CHF": "CHF",
		"SEK": "kr",
		"NOK": "kr",
		"DKK": "kr",
	}

	if symbol, exists := symbols[currency]; exists {
		return symbol
	}

	return currency // Return currency code if symbol not found
}

// CalculationSummary represents a summary of calculations across multiple invoices
type CalculationSummary struct {
	InvoiceCount         int       `json:"invoice_count"`
	TotalSubtotal        float64   `json:"total_subtotal"`
	TotalTax             float64   `json:"total_tax"`
	TotalAmount          float64   `json:"total_amount"`
	TotalHours           float64   `json:"total_hours"`
	AverageRate          float64   `json:"average_rate"`
	AverageInvoiceAmount float64   `json:"average_invoice_amount"`
	CalculatedAt         time.Time `json:"calculated_at"`
}
