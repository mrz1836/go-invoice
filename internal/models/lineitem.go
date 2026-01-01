package models

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

// LineItemType represents the type of line item
type LineItemType string

const (
	// LineItemTypeHourly represents hourly billing (hours × rate)
	LineItemTypeHourly LineItemType = "hourly"
	// LineItemTypeFixed represents fixed amount billing (flat fee, retainer)
	LineItemTypeFixed LineItemType = "fixed"
	// LineItemTypeQuantity represents quantity-based billing (quantity × unit price)
	LineItemTypeQuantity LineItemType = "quantity"
)

// ValidLineItemTypes contains all valid line item type values
//
//nolint:gochecknoglobals // Constant-like type validation slice required for validation
var ValidLineItemTypes = []string{
	string(LineItemTypeHourly),
	string(LineItemTypeFixed),
	string(LineItemTypeQuantity),
}

// LineItem represents a flexible invoice line item that supports multiple billing types
type LineItem struct {
	ID          string       `json:"id"`
	Type        LineItemType `json:"type"`
	Date        time.Time    `json:"date"`
	EndDate     *time.Time   `json:"end_date,omitempty"` // Optional end date for date ranges (e.g., monthly retainers)
	Description string       `json:"description"`

	// For hourly items (Type == LineItemTypeHourly)
	Hours *float64 `json:"hours,omitempty"`
	Rate  *float64 `json:"rate,omitempty"`

	// For fixed items (Type == LineItemTypeFixed)
	Amount *float64 `json:"amount,omitempty"`

	// For quantity items (Type == LineItemTypeQuantity)
	Quantity  *float64 `json:"quantity,omitempty"`
	UnitPrice *float64 `json:"unit_price,omitempty"`

	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}

// NewHourlyLineItem creates a new hourly-based line item
func NewHourlyLineItem(ctx context.Context, id string, date time.Time, hours, rate float64, description string) (*LineItem, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	total := math.Round(hours*rate*100) / 100

	item := &LineItem{
		ID:          id,
		Type:        LineItemTypeHourly,
		Date:        date,
		Hours:       &hours,
		Rate:        &rate,
		Description: description,
		Total:       total,
		CreatedAt:   time.Now(),
	}

	if err := item.Validate(ctx); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLineItemValidationFailed, err)
	}

	return item, nil
}

// NewFixedLineItem creates a new fixed-amount line item (retainer, flat fee)
func NewFixedLineItem(ctx context.Context, id string, date time.Time, amount float64, description string) (*LineItem, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	total := math.Round(amount*100) / 100

	item := &LineItem{
		ID:          id,
		Type:        LineItemTypeFixed,
		Date:        date,
		Amount:      &amount,
		Description: description,
		Total:       total,
		CreatedAt:   time.Now(),
	}

	if err := item.Validate(ctx); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLineItemValidationFailed, err)
	}

	return item, nil
}

// NewQuantityLineItem creates a new quantity-based line item
func NewQuantityLineItem(ctx context.Context, id string, date time.Time, quantity, unitPrice float64, description string) (*LineItem, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	total := math.Round(quantity*unitPrice*100) / 100

	item := &LineItem{
		ID:          id,
		Type:        LineItemTypeQuantity,
		Date:        date,
		Quantity:    &quantity,
		UnitPrice:   &unitPrice,
		Description: description,
		Total:       total,
		CreatedAt:   time.Now(),
	}

	if err := item.Validate(ctx); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLineItemValidationFailed, err)
	}

	return item, nil
}

// Validate performs comprehensive validation of the line item based on its type
func (l *LineItem) Validate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	builder := NewValidationBuilder().
		AddRequired("id", l.ID).
		AddTimeRequired("date", l.Date).
		AddDateNotFuture("date", l.Date, 24).
		AddRequired("description", l.Description).
		AddMaxLength("description", l.Description, 1000).
		AddNonNegative("total", l.Total).
		AddTimeRequired("created_at", l.CreatedAt)

	// Validate optional EndDate if provided
	if l.EndDate != nil {
		if l.EndDate.Before(l.Date) {
			builder.AddCustom("end_date", "cannot be before date", l.EndDate)
		}
	}

	// Validate type-specific fields
	switch l.Type {
	case LineItemTypeHourly:
		if l.Hours == nil {
			builder.AddCustom("hours", "is required for hourly line items", nil)
		} else {
			expectedTotal := math.Round(*l.Hours**l.Rate*100) / 100
			builder.
				AddFloatValidation("hours", *l.Hours, 24, "24 hours per entry").
				AddFloatValidation("rate", *l.Rate, 10000, "$10,000 per hour").
				AddCalculationValidation("total", l.Total, expectedTotal)
		}

		if l.Rate == nil {
			builder.AddCustom("rate", "is required for hourly line items", nil)
		}

		// Ensure fixed/quantity fields are nil
		if l.Amount != nil {
			builder.AddCustom("amount", "should not be set for hourly line items", l.Amount)
		}
		if l.Quantity != nil || l.UnitPrice != nil {
			builder.AddCustom("quantity/unit_price", "should not be set for hourly line items", nil)
		}

	case LineItemTypeFixed:
		if l.Amount == nil {
			builder.AddCustom("amount", "is required for fixed line items", nil)
		} else {
			expectedTotal := math.Round(*l.Amount*100) / 100
			builder.
				AddPositive("amount", *l.Amount).
				AddMaxValue("amount", *l.Amount, 1000000, "$1,000,000").
				AddCalculationValidation("total", l.Total, expectedTotal)
		}

		// Ensure hourly/quantity fields are nil
		if l.Hours != nil || l.Rate != nil {
			builder.AddCustom("hours/rate", "should not be set for fixed line items", nil)
		}
		if l.Quantity != nil || l.UnitPrice != nil {
			builder.AddCustom("quantity/unit_price", "should not be set for fixed line items", nil)
		}

	case LineItemTypeQuantity:
		if l.Quantity == nil {
			builder.AddCustom("quantity", "is required for quantity line items", nil)
		} else {
			expectedTotal := math.Round(*l.Quantity**l.UnitPrice*100) / 100
			builder.
				AddFloatValidation("quantity", *l.Quantity, 10000, "10,000 units").
				AddFloatValidation("unit_price", *l.UnitPrice, 100000, "$100,000 per unit").
				AddCalculationValidation("total", l.Total, expectedTotal)
		}

		if l.UnitPrice == nil {
			builder.AddCustom("unit_price", "is required for quantity line items", nil)
		}

		// Ensure hourly/fixed fields are nil
		if l.Hours != nil || l.Rate != nil {
			builder.AddCustom("hours/rate", "should not be set for quantity line items", nil)
		}
		if l.Amount != nil {
			builder.AddCustom("amount", "should not be set for quantity line items", l.Amount)
		}

	default:
		builder.AddValidOption("type", string(l.Type), ValidLineItemTypes)
	}

	return builder.Build(ErrLineItemValidationFailed)
}

// RecalculateTotal recalculates the total based on the line item type
func (l *LineItem) RecalculateTotal(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	switch l.Type {
	case LineItemTypeHourly:
		if l.Hours != nil && l.Rate != nil {
			l.Total = math.Round(*l.Hours**l.Rate*100) / 100
		}
	case LineItemTypeFixed:
		if l.Amount != nil {
			l.Total = math.Round(*l.Amount*100) / 100
		}
	case LineItemTypeQuantity:
		if l.Quantity != nil && l.UnitPrice != nil {
			l.Total = math.Round(*l.Quantity**l.UnitPrice*100) / 100
		}
	default:
		return fmt.Errorf("%w: unsupported line item type: %s", ErrInvalidLineItemType, l.Type)
	}

	return nil
}

// GetFormattedTotal returns the total formatted as a currency string
func (l *LineItem) GetFormattedTotal() string {
	return fmt.Sprintf("$%.2f", l.Total)
}

// GetDetails returns a human-readable string describing the line item details
func (l *LineItem) GetDetails() string {
	switch l.Type {
	case LineItemTypeHourly:
		if l.Hours != nil && l.Rate != nil {
			return fmt.Sprintf("%.2f hours @ $%.2f/hr", *l.Hours, *l.Rate)
		}
	case LineItemTypeFixed:
		return "Fixed amount"
	case LineItemTypeQuantity:
		if l.Quantity != nil && l.UnitPrice != nil {
			return fmt.Sprintf("%.2f × $%.2f", *l.Quantity, *l.UnitPrice)
		}
	}
	return ""
}

// UpdateDescription updates the line item description
func (l *LineItem) UpdateDescription(ctx context.Context, description string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	description = strings.TrimSpace(description)
	if description == "" {
		return ErrDescriptionRequired
	}

	if len(description) > 1000 {
		return ErrDescriptionTooLong
	}

	l.Description = description
	return nil
}

// ConvertWorkItemToLineItem converts a WorkItem to a LineItem (for migration)
func ConvertWorkItemToLineItem(ctx context.Context, wi WorkItem) (*LineItem, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	hours := wi.Hours
	rate := wi.Rate

	return &LineItem{
		ID:          wi.ID,
		Type:        LineItemTypeHourly,
		Date:        wi.Date,
		Description: wi.Description,
		Hours:       &hours,
		Rate:        &rate,
		Total:       wi.Total,
		CreatedAt:   wi.CreatedAt,
	}, nil
}
