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
		return nil, fmt.Errorf("work item validation failed: %w", err)
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

	var errors []ValidationError

	// Validate ID
	if strings.TrimSpace(w.ID) == "" {
		errors = append(errors, ValidationError{
			Field:   "id",
			Message: "is required",
			Value:   w.ID,
		})
	}

	// Validate date
	if w.Date.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "date",
			Message: "is required",
			Value:   w.Date,
		})
	}

	// Validate date is not in the future (more than 1 day ahead to account for timezone differences)
	if w.Date.After(time.Now().Add(24 * time.Hour)) {
		errors = append(errors, ValidationError{
			Field:   "date",
			Message: "cannot be more than 1 day in the future",
			Value:   w.Date,
		})
	}

	// Validate hours
	if w.Hours <= 0 {
		errors = append(errors, ValidationError{
			Field:   "hours",
			Message: "must be greater than 0",
			Value:   w.Hours,
		})
	}

	// Reasonable upper limit for hours per day
	if w.Hours > 24 {
		errors = append(errors, ValidationError{
			Field:   "hours",
			Message: "cannot exceed 24 hours per entry",
			Value:   w.Hours,
		})
	}

	// Validate rate
	if w.Rate <= 0 {
		errors = append(errors, ValidationError{
			Field:   "rate",
			Message: "must be greater than 0",
			Value:   w.Rate,
		})
	}

	// Reasonable upper limit for hourly rate
	if w.Rate > 10000 {
		errors = append(errors, ValidationError{
			Field:   "rate",
			Message: "cannot exceed $10,000 per hour",
			Value:   w.Rate,
		})
	}

	// Validate description
	if strings.TrimSpace(w.Description) == "" {
		errors = append(errors, ValidationError{
			Field:   "description",
			Message: "is required",
			Value:   w.Description,
		})
	}

	// Validate description length
	if len(strings.TrimSpace(w.Description)) > 1000 {
		errors = append(errors, ValidationError{
			Field:   "description",
			Message: "cannot exceed 1000 characters",
			Value:   len(w.Description),
		})
	}

	// Validate total calculation
	expectedTotal := math.Round(w.Hours*w.Rate*100) / 100
	if math.Abs(w.Total-expectedTotal) > 0.01 {
		errors = append(errors, ValidationError{
			Field:   "total",
			Message: fmt.Sprintf("incorrect calculation, expected %.2f", expectedTotal),
			Value:   w.Total,
		})
	}

	// Validate total is reasonable
	if w.Total < 0 {
		errors = append(errors, ValidationError{
			Field:   "total",
			Message: "must be non-negative",
			Value:   w.Total,
		})
	}

	// Validate created timestamp
	if w.CreatedAt.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "created_at",
			Message: "is required",
			Value:   w.CreatedAt,
		})
	}

	if len(errors) > 0 {
		var messages []string
		for _, err := range errors {
			messages = append(messages, err.Error())
		}
		return fmt.Errorf("work item validation failed: %s", strings.Join(messages, "; "))
	}

	return nil
}

// UpdateHours updates the hours worked and recalculates the total
func (w *WorkItem) UpdateHours(ctx context.Context, hours float64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if hours <= 0 {
		return fmt.Errorf("hours must be greater than 0")
	}

	if hours > 24 {
		return fmt.Errorf("hours cannot exceed 24 per entry")
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
		return fmt.Errorf("rate must be greater than 0")
	}

	if rate > 10000 {
		return fmt.Errorf("rate cannot exceed $10,000 per hour")
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
		return fmt.Errorf("description cannot be empty")
	}

	if len(description) > 1000 {
		return fmt.Errorf("description cannot exceed 1000 characters")
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
