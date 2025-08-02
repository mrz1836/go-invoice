package csv

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mrz/go-invoice/internal/models"
)

// CSV validation errors
var (
	ErrWorkItemNil      = fmt.Errorf("work item cannot be nil")
	ErrRowEmpty         = fmt.Errorf("row is empty")
	ErrRowNoData        = fmt.Errorf("row contains no data")
	ErrNoWorkItems      = fmt.Errorf("no work items to validate")
	ErrWorkDateEmpty    = fmt.Errorf("work date cannot be empty")
	ErrWorkDateFuture   = fmt.Errorf("work date is too far in the future")
	ErrWorkDatePast     = fmt.Errorf("work date is too far in the past")
	ErrHoursNotPositive = fmt.Errorf("hours must be positive")
)

// WorkItemValidator implements CSVValidator with comprehensive validation rules
type WorkItemValidator struct {
	logger Logger
	rules  []ValidationRule
}

// NewWorkItemValidator creates a new work item validator with dependency injection
func NewWorkItemValidator(logger Logger) *WorkItemValidator {
	validator := &WorkItemValidator{
		logger: logger,
	}

	// Initialize standard validation rules
	validator.rules = validator.createStandardRules()

	return validator
}

// ValidateWorkItem validates a single work item with context support
func (v *WorkItemValidator) ValidateWorkItem(ctx context.Context, item *models.WorkItem) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if item == nil {
		return ErrWorkItemNil
	}

	// Apply all validation rules
	for _, rule := range v.rules {
		if rule.Validator != nil {
			if err := rule.Validator(ctx, item); err != nil {
				return fmt.Errorf("validation rule '%s' failed: %w", rule.Name, err)
			}
		}
	}

	// Use the model's built-in validation as final check
	if err := item.Validate(ctx); err != nil {
		return fmt.Errorf("work item model validation failed: %w", err)
	}

	return nil
}

// ValidateRow validates a raw CSV row before parsing
func (v *WorkItemValidator) ValidateRow(ctx context.Context, row []string, lineNum int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if len(row) == 0 {
		return ErrRowEmpty
	}

	// Check for completely empty row (all fields empty)
	isEmpty := true
	for _, field := range row {
		if strings.TrimSpace(field) != "" {
			isEmpty = false
			break
		}
	}

	if isEmpty {
		return ErrRowNoData
	}

	// Apply row-level validation rules
	for _, rule := range v.rules {
		if rule.RowValidator != nil {
			if err := rule.RowValidator(ctx, row, lineNum); err != nil {
				return fmt.Errorf("row validation rule '%s' failed: %w", rule.Name, err)
			}
		}
	}

	return nil
}

// ValidateBatch validates a batch of work items for consistency
func (v *WorkItemValidator) ValidateBatch(ctx context.Context, items []models.WorkItem) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if len(items) == 0 {
		return ErrNoWorkItems
	}

	v.logger.Debug("validating work items batch", "count", len(items))

	// Validate each item individually
	for i, item := range items {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := v.ValidateWorkItem(ctx, &item); err != nil {
			return fmt.Errorf("work item %d validation failed: %w", i+1, err)
		}
	}

	// Batch-level validations
	if err := v.validateDateRange(items); err != nil {
		return fmt.Errorf("date range validation failed: %w", err)
	}

	if err := v.validateRateConsistency(items); err != nil {
		return fmt.Errorf("rate consistency validation failed: %w", err)
	}

	if err := v.validateTotalHours(items); err != nil {
		return fmt.Errorf("total hours validation failed: %w", err)
	}

	v.logger.Debug("batch validation completed successfully", "items", len(items))

	return nil
}

// createStandardRules creates the standard set of validation rules
func (v *WorkItemValidator) createStandardRules() []ValidationRule {
	return []ValidationRule{
		{
			Name:        "DateValidation",
			Description: "Validates work item dates are reasonable",
			Validator:   v.validateDate,
		},
		{
			Name:        "HoursValidation",
			Description: "Validates hours are within reasonable limits",
			Validator:   v.validateHours,
		},
		{
			Name:        "RateValidation",
			Description: "Validates hourly rates are positive and reasonable",
			Validator:   v.validateRate,
		},
		{
			Name:        "DescriptionValidation",
			Description: "Validates work descriptions are meaningful",
			Validator:   v.validateDescription,
		},
		{
			Name:        "TotalValidation",
			Description: "Validates calculated totals are correct",
			Validator:   v.validateTotal,
		},
		{
			Name:         "RowFormatValidation",
			Description:  "Validates CSV row format and field count",
			RowValidator: v.validateRowFormat,
		},
	}
}

// Individual validation rule implementations

func (v *WorkItemValidator) validateDate(ctx context.Context, item *models.WorkItem) error {
	now := time.Now()

	// Check if date is not zero
	if item.Date.IsZero() {
		return ErrWorkDateEmpty
	}

	// Check if date is not too far in the future (more than 1 week)
	futureLimit := now.AddDate(0, 0, 7)
	if item.Date.After(futureLimit) {
		return fmt.Errorf("%w (more than 1 week from now): %s", ErrWorkDateFuture,
			item.Date.Format("2006-01-02"))
	}

	// Check if date is not too far in the past (more than 2 years)
	pastLimit := now.AddDate(-2, 0, 0)
	if item.Date.Before(pastLimit) {
		return fmt.Errorf("%w (more than 2 years ago): %s", ErrWorkDatePast,
			item.Date.Format("2006-01-02"))
	}

	return nil
}

func (v *WorkItemValidator) validateHours(ctx context.Context, item *models.WorkItem) error {
	if item.Hours <= 0 {
		return fmt.Errorf("%w, got %v", ErrHoursNotPositive, item.Hours)
	}

	if item.Hours > 24 {
		return fmt.Errorf("hours cannot exceed 24 per day, got %v", item.Hours)
	}

	// Warn about unusual hours (more than 12 per day)
	if item.Hours > 12 {
		v.logger.Debug("unusually high hours detected", "hours", item.Hours, "date", item.Date)
	}

	// Check for reasonable precision (max 2 decimal places)
	hoursStr := fmt.Sprintf("%.2f", item.Hours)
	if parsedHours, err := strconv.ParseFloat(hoursStr, 64); err == nil {
		if parsedHours != item.Hours {
			// Hours has more than 2 decimal places
			return fmt.Errorf("hours should not have more than 2 decimal places, got %v", item.Hours)
		}
	}

	return nil
}

func (v *WorkItemValidator) validateRate(ctx context.Context, item *models.WorkItem) error {
	if item.Rate <= 0 {
		return fmt.Errorf("hourly rate must be positive, got %v", item.Rate)
	}

	// Check for reasonable rate limits
	if item.Rate < 1 {
		return fmt.Errorf("hourly rate seems too low: $%v per hour", item.Rate)
	}

	if item.Rate > 1000 {
		return fmt.Errorf("hourly rate seems too high: $%v per hour", item.Rate)
	}

	// Warn about unusual rates
	if item.Rate > 500 {
		v.logger.Debug("unusually high rate detected", "rate", item.Rate, "date", item.Date)
	}

	return nil
}

func (v *WorkItemValidator) validateDescription(ctx context.Context, item *models.WorkItem) error {
	description := strings.TrimSpace(item.Description)

	if description == "" {
		return fmt.Errorf("work description cannot be empty")
	}

	if len(description) < 3 {
		return fmt.Errorf("work description too short: '%s' (minimum 3 characters)", description)
	}

	if len(description) > 500 {
		return fmt.Errorf("work description too long: %d characters (maximum 500)", len(description))
	}

	// Check for placeholder or generic descriptions
	genericDescriptions := []string{
		"work", "development", "coding", "programming", "task", "project",
		"meeting", "call", "todo", "fix", "bug", "feature",
	}

	lowerDesc := strings.ToLower(description)
	for _, generic := range genericDescriptions {
		if lowerDesc == generic {
			return fmt.Errorf("work description too generic: '%s' (please be more specific)", description)
		}
	}

	return nil
}

func (v *WorkItemValidator) validateTotal(ctx context.Context, item *models.WorkItem) error {
	expectedTotal := item.Hours * item.Rate
	tolerance := 0.01 // Allow small floating point differences

	if item.Total < expectedTotal-tolerance || item.Total > expectedTotal+tolerance {
		return fmt.Errorf("total amount %v does not match calculated value %v (hours: %v, rate: %v)",
			item.Total, expectedTotal, item.Hours, item.Rate)
	}

	return nil
}

func (v *WorkItemValidator) validateRowFormat(ctx context.Context, row []string, lineNum int) error {
	// Check minimum number of fields
	if len(row) < 4 {
		return fmt.Errorf("row has %d fields, expected at least 4 (date, hours, rate, description)", len(row))
	}

	// Check for excessively long rows
	if len(row) > 20 {
		return fmt.Errorf("row has %d fields, which seems excessive (maximum expected: 20)", len(row))
	}

	return nil
}

// Batch validation methods

func (v *WorkItemValidator) validateDateRange(items []models.WorkItem) error {
	if len(items) <= 1 {
		return nil // No range to validate
	}

	var minDate, maxDate time.Time

	for i, item := range items {
		if i == 0 {
			minDate = item.Date
			maxDate = item.Date

			continue
		}

		if item.Date.Before(minDate) {
			minDate = item.Date
		}
		if item.Date.After(maxDate) {
			maxDate = item.Date
		}
	}

	// Check if date range is reasonable (not more than 1 year)
	if maxDate.Sub(minDate) > 365*24*time.Hour {
		return fmt.Errorf("work item date range is too large: %s to %s (more than 1 year)",
			minDate.Format("2006-01-02"), maxDate.Format("2006-01-02"))
	}

	return nil
}

func (v *WorkItemValidator) validateRateConsistency(items []models.WorkItem) error {
	if len(items) <= 1 {
		return nil
	}

	// Group rates and check for consistency
	rateCount := make(map[float64]int)
	for _, item := range items {
		rateCount[item.Rate]++
	}

	// If more than 3 different rates, warn about potential inconsistency
	if len(rateCount) > 3 {
		v.logger.Debug("multiple different rates detected", "unique_rates", len(rateCount))
	}

	return nil
}

func (v *WorkItemValidator) validateTotalHours(items []models.WorkItem) error {
	totalHours := 0.0
	for _, item := range items {
		totalHours += item.Hours
	}

	// Warn about unusual total hours
	if totalHours > 200 {
		v.logger.Debug("large total hours detected", "total_hours", totalHours)
	}

	if totalHours == 0 {
		return fmt.Errorf("total hours cannot be zero")
	}

	return nil
}

// AddCustomRule allows adding custom validation rules
func (v *WorkItemValidator) AddCustomRule(rule ValidationRule) {
	v.rules = append(v.rules, rule)
	v.logger.Debug("custom validation rule added", "rule", rule.Name)
}

// RemoveRule removes a validation rule by name
func (v *WorkItemValidator) RemoveRule(ruleName string) bool {
	for i, rule := range v.rules {
		if rule.Name == ruleName {
			v.rules = append(v.rules[:i], v.rules[i+1:]...)
			v.logger.Debug("validation rule removed", "rule", ruleName)

			return true
		}
	}

	return false
}

// GetRules returns all active validation rules
func (v *WorkItemValidator) GetRules() []ValidationRule {
	return append([]ValidationRule{}, v.rules...) // Return a copy
}
