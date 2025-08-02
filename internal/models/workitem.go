package models

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

// NewWorkItem creates a new work item with validation and automatic total calculation
func NewWorkItem(ctx context.Context, id string, date time.Time, hours, rate float64, description string) (*WorkItem, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Calculate total with proper rounding
	total := math.Round(hours*rate*100) / 100

	item := &WorkItem{
		ID:          id,
		Date:        date,
		Hours:       hours,
		Rate:        rate,
		Description: description,
		Total:       total,
		CreatedAt:   time.Now(),
	}

	// Validate the new work item
	if err := item.Validate(ctx); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrWorkItemValidationFailed, err)
	}

	return item, nil
}

// Validate performs comprehensive validation of the work item
func (w *WorkItem) Validate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	expectedTotal := math.Round(w.Hours*w.Rate*100) / 100

	return NewValidationBuilder().
		AddRequired("id", w.ID).
		AddTimeRequired("date", w.Date).
		AddDateNotFuture("date", w.Date, 24).
		AddFloatValidation("hours", w.Hours, 24, "24 hours per entry").
		AddFloatValidation("rate", w.Rate, 10000, "$10,000 per hour").
		AddRequired("description", w.Description).
		AddMaxLength("description", w.Description, 1000).
		AddCalculationValidation("total", w.Total, expectedTotal).
		AddNonNegative("total", w.Total).
		AddTimeRequired("created_at", w.CreatedAt).
		Build(ErrWorkItemValidationFailed)
}

// UpdateHours updates the hours worked and recalculates the total
func (w *WorkItem) UpdateHours(ctx context.Context, hours float64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if hours <= 0 {
		return ErrHoursMustBePositive
	}

	if hours > 24 {
		return ErrHoursExceedLimit
	}

	w.Hours = hours
	w.Total = math.Round(w.Hours*w.Rate*100) / 100

	return nil
}

// UpdateRate updates the hourly rate and recalculates the total
func (w *WorkItem) UpdateRate(ctx context.Context, rate float64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if rate <= 0 {
		return ErrRateMustBePositive
	}

	if rate > 10000 {
		return ErrRateExceedsLimit
	}

	w.Rate = rate
	w.Total = math.Round(w.Hours*w.Rate*100) / 100

	return nil
}

// UpdateDescription updates the work item description
func (w *WorkItem) UpdateDescription(ctx context.Context, description string) error {
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

	w.Description = description
	return nil
}

// GetFormattedTotal returns the total formatted as a currency string
func (w *WorkItem) GetFormattedTotal() string {
	return fmt.Sprintf("$%.2f", w.Total)
}

// GetFormattedRate returns the rate formatted as a currency string
func (w *WorkItem) GetFormattedRate() string {
	return fmt.Sprintf("$%.2f", w.Rate)
}

// GetFormattedHours returns the hours formatted to 2 decimal places
func (w *WorkItem) GetFormattedHours() string {
	return fmt.Sprintf("%.2f", w.Hours)
}
