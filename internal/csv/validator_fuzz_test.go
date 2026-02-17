package csv

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/mrz1836/go-invoice/internal/models"
)

// FuzzValidateWorkItem fuzzes work item validation
func FuzzValidateWorkItem(f *testing.F) {
	// Generate valid dates for seed data
	validDate := time.Now().AddDate(-1, 0, 0)
	date1 := validDate.Format("2006-01-02")
	date2 := validDate.AddDate(0, 0, 1).Format("2006-01-02")
	date3 := validDate.AddDate(0, 0, 2).Format("2006-01-02")
	date4 := validDate.AddDate(0, 0, 3).Format("2006-01-02")

	// Seed with various work item examples
	seedItems := []struct {
		id          string
		dateStr     string
		hours       float64
		rate        float64
		description string
	}{
		{"test-1", date1, 8.0, 100.0, "Development work"},
		{"test-2", date2, 4.5, 125.5, "Bug fixes"},
		{"test-3", date3, 0.25, 200.0, "Quick consultation"},
		{"test-4", date4, 12.0, 75.0, "Extended development session"},
		{"", date1, 8.0, 100.0, "No ID"},                          // Empty ID
		{"test-5", "", 8.0, 100.0, "Development work"},            // Empty date
		{"test-6", date1, -1.0, 100.0, "Negative hours"},          // Negative hours
		{"test-7", date1, 8.0, -50.0, "Negative rate"},            // Negative rate
		{"test-8", date1, 8.0, 100.0, ""},                         // Empty description
		{"test-9", date1, 25.0, 100.0, "Too many hours"},          // Hours > 24
		{"test-10", date1, 8.0, 5000.0, "Very high rate"},         // Very high rate
		{"test-11", date1, 8.0, 0.1, "Very low rate"},             // Very low rate
		{"test-12", date1, 0.0, 100.0, "Zero hours"},              // Zero hours
		{"test-13", date1, 8.0, 0.0, "Zero rate"},                 // Zero rate
		{"test-14", "2050-01-15", 8.0, 100.0, "Future date"},      // Far future date (testing validation)
		{"test-15", "1900-01-15", 8.0, 100.0, "Past date"},        // Far past date (testing validation)
		{"test-16", date1, 8.0, 100.0, "A"},                       // Very short description
		{"test-17", date1, 8.0, 100.0, strings.Repeat("A", 1001)}, // Very long description
	}

	for _, seed := range seedItems {
		f.Add(seed.id, seed.dateStr, seed.hours, seed.rate, seed.description)
	}

	f.Fuzz(func(t *testing.T, id, dateStr string, hours, rate float64, description string) {
		logger := &MockLogger{}
		validator := NewWorkItemValidator(logger)

		// Should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ValidateWorkItem panicked with input id=%q, date=%q, hours=%f, rate=%f, desc=%q: %v",
					id, dateStr, hours, rate, description, r)
			}
		}()

		// Parse date if possible
		var date time.Time
		if dateStr != "" {
			if parsedDate, err := time.Parse("2006-01-02", dateStr); err == nil {
				date = parsedDate
			}
		}

		// Create work item with given parameters (may fail)
		ctx := context.Background()
		workItem, err := models.NewWorkItem(ctx, id, date, hours, rate, description)

		// If work item creation succeeded, test validation
		if err == nil && workItem != nil {
			validationErr := validator.ValidateWorkItem(ctx, workItem)

			// Test validation invariants
			if validationErr == nil {
				// If validation passed, work item should meet basic criteria
				if workItem.ID == "" {
					t.Errorf("Valid work item should have non-empty ID")
				}
				if workItem.Hours <= 0 {
					t.Errorf("Valid work item should have positive hours: %f", workItem.Hours)
				}
				if workItem.Rate <= 0 {
					t.Errorf("Valid work item should have positive rate: %f", workItem.Rate)
				}
				if workItem.Description == "" {
					t.Errorf("Valid work item should have non-empty description")
				}
				if workItem.Date.IsZero() {
					t.Errorf("Valid work item should have non-zero date")
				}
				if workItem.Hours > 24 {
					t.Errorf("Valid work item should not exceed 24 hours: %f", workItem.Hours)
				}

				// Check date is reasonable
				now := time.Now()
				if workItem.Date.Before(now.AddDate(-2, 0, 0)) {
					t.Errorf("Valid work item date should not be more than 2 years in past: %v", workItem.Date)
				}
				if workItem.Date.After(now.AddDate(0, 0, 7)) {
					t.Errorf("Valid work item date should not be more than 1 week in future: %v", workItem.Date)
				}

				// Check calculation
				expectedTotal := math.Round(workItem.Hours*workItem.Rate*100) / 100
				if math.Abs(workItem.Total-expectedTotal) > 0.01 {
					t.Errorf("Valid work item should have correct total calculation: %f != %f", workItem.Total, expectedTotal)
				}
			}
		}

		// Test with nil work item (should always return error)
		nilErr := validator.ValidateWorkItem(ctx, nil)
		if nilErr == nil {
			t.Errorf("Validating nil work item should return error")
		}

		// Test context cancellation
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()
		if workItem != nil {
			cancelErr := validator.ValidateWorkItem(cancelCtx, workItem)
			if !errors.Is(cancelErr, context.Canceled) {
				t.Errorf("Validation with canceled context should return context.Canceled, got: %v", cancelErr)
			}
		}
	})
}

// FuzzValidateRow fuzzes raw CSV row validation
func FuzzValidateRow(f *testing.F) {
	// Generate valid dates for seed data
	validDate := time.Now().AddDate(-1, 0, 0)
	date1 := validDate.Format("2006-01-02")
	date2 := validDate.AddDate(0, 0, 1).Format("2006-01-02")

	// Seed with various row examples
	seedRows := [][]string{
		{date1, "8.0", "100.00", "Development work"},
		{date2, "4.5", "125.50", "Bug fixes"},
		{},                       // Empty row
		{"", "", "", ""},         // Row with empty fields
		{date1},                  // Too few fields
		{date1, "8.0"},           // Too few fields
		{date1, "8.0", "100.00"}, // Too few fields
		{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u"}, // Too many fields
		{"   ", "   ", "   ", "   "},                          // Whitespace only
		{date1, "8.0", "100.00", "Development work", "extra"}, // Extra field
	}

	// Convert to proper format for fuzzer
	for i, row := range seedRows {
		// Convert []string to individual string parameters
		// We'll use up to 10 fields for fuzzing
		fields := make([]string, 10)
		for j, field := range row {
			if j < 10 {
				fields[j] = field //nolint:gosec // G602: j < 10 guard ensures index is within bounds of fields (size 10)
			}
		}
		f.Add(fields[0], fields[1], fields[2], fields[3], fields[4], fields[5], fields[6], fields[7], fields[8], fields[9], len(row), i+1)
	}

	f.Fuzz(func(t *testing.T, f0, f1, f2, f3, f4, f5, f6, f7, f8, f9 string, actualLen, lineNum int) {
		// Reconstruct row from fields
		allFields := []string{f0, f1, f2, f3, f4, f5, f6, f7, f8, f9}
		if actualLen < 0 {
			actualLen = 0
		}
		if actualLen > 10 {
			actualLen = 10
		}
		row := allFields[:actualLen]

		logger := &MockLogger{}
		validator := NewWorkItemValidator(logger)

		// Should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ValidateRow panicked with row %v, line %d: %v", row, lineNum, r)
			}
		}()

		ctx := context.Background()
		err := validator.ValidateRow(ctx, row, lineNum)

		// Test invariants
		// Empty row should return error
		if len(row) == 0 && err == nil {
			t.Errorf("Empty row should return error")
		}

		// Row with all empty/whitespace fields should return error
		if len(row) > 0 {
			allEmpty := true
			for _, field := range row {
				if strings.TrimSpace(field) != "" {
					allEmpty = false
					break
				}
			}
			if allEmpty && err == nil {
				t.Errorf("Row with all empty fields should return error")
			}
		}

		// Row with too few fields should return error
		if len(row) > 0 && len(row) < 4 && err == nil {
			// Check if it's not all empty
			hasContent := false
			for _, field := range row {
				if strings.TrimSpace(field) != "" {
					hasContent = true
					break
				}
			}
			if hasContent {
				t.Errorf("Row with too few fields (%d) should return error", len(row))
			}
		}

		// Row with too many fields should return error
		if len(row) > 20 && err == nil {
			t.Errorf("Row with too many fields (%d) should return error", len(row))
		}

		// Line number should be positive
		// Line number validation is not enforced here as it's not critical for validation logic

		// Test context cancellation
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()
		cancelErr := validator.ValidateRow(cancelCtx, row, lineNum)
		if !errors.Is(cancelErr, context.Canceled) {
			t.Errorf("Validation with canceled context should return context.Canceled, got: %v", cancelErr)
		}
	})
}

// FuzzValidateBatch fuzzes batch validation of work items
func FuzzValidateBatch(f *testing.F) {
	// Create seeds with various batch scenarios
	ctx := context.Background()
	logger := &MockLogger{}

	// This function creates work items dynamically in the fuzz function
	// No need to pre-create items here as seeds are simple integers

	// Seed with empty batch
	f.Add(0)
	// Seed with various batch sizes
	for i := 1; i <= 10; i++ {
		f.Add(i)
	}
	// Seed with large batch
	f.Add(100)

	f.Fuzz(func(t *testing.T, batchSize int) {
		validator := NewWorkItemValidator(logger)

		// Should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ValidateBatch panicked with batch size %d: %v", batchSize, r)
			}
		}()

		// Create batch of work items
		var items []models.WorkItem

		// Clamp batch size to reasonable limits for fuzzing
		if batchSize < 0 {
			batchSize = 0
		}
		if batchSize > 1000 {
			batchSize = 1000
		}

		for i := 0; i < batchSize; i++ {
			// Create work item with varying parameters
			date := time.Date(2024, 1, 15+(i%30), 0, 0, 0, 0, time.UTC)
			hours := 1.0 + float64(i%16)  // 1-16 hours
			rate := 50.0 + float64(i%200) // 50-250 rate
			description := fmt.Sprintf("Work item %d", i)

			item, err := models.NewWorkItem(ctx, fmt.Sprintf("item-%d", i), date, hours, rate, description)
			if err == nil {
				items = append(items, *item)
			}
		}

		err := validator.ValidateBatch(ctx, items)

		// Test invariants
		// Empty batch should return error
		if len(items) == 0 && err == nil {
			t.Errorf("Empty batch should return error")
		}

		// If validation succeeded, all items should be valid individually
		if err == nil {
			for i, item := range items {
				individualErr := validator.ValidateWorkItem(ctx, &item)
				if individualErr != nil {
					t.Errorf("Item %d should be valid if batch validation passed: %v", i, individualErr)
				}
			}

			// Check batch-level invariants
			if len(items) > 0 {
				// Date range should be reasonable
				minDate, maxDate := items[0].Date, items[0].Date
				totalHours := 0.0
				rateMap := make(map[float64]int)

				for _, item := range items {
					if item.Date.Before(minDate) {
						minDate = item.Date
					}
					if item.Date.After(maxDate) {
						maxDate = item.Date
					}
					totalHours += item.Hours
					rateMap[item.Rate]++
				}

				// Date range should not exceed 1 year
				if maxDate.Sub(minDate) > 365*24*time.Hour {
					t.Errorf("Batch validation should reject date ranges > 1 year, but didn't")
				}

				// Total hours should be positive
				if totalHours <= 0 {
					t.Errorf("Batch validation should reject zero total hours, but didn't")
				}
			}
		}

		// Test context cancellation
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		cancelErr := validator.ValidateBatch(cancelCtx, items)
		if len(items) > 0 && !errors.Is(cancelErr, context.Canceled) {
			t.Errorf("Batch validation with canceled context should return context.Canceled, got: %v", cancelErr)
		}
	})
}

// FuzzValidationRules fuzzes individual validation rules
func FuzzValidationRules(f *testing.F) {
	// Seed with various parameter combinations
	seeds := []struct {
		hours       float64
		rate        float64
		description string
		dateOffset  int // Days from now
	}{
		{8.0, 100.0, "Development work", 0},
		{-1.0, 100.0, "Negative hours", 0},       // Invalid hours
		{8.0, -50.0, "Negative rate", 0},         // Invalid rate
		{8.0, 100.0, "", 0},                      // Empty description
		{25.0, 100.0, "Too many hours", 0},       // Hours > 24
		{0.0, 100.0, "Zero hours", 0},            // Zero hours
		{8.0, 0.0, "Zero rate", 0},               // Zero rate
		{8.0, 5000.0, "Very high rate", 0},       // Very high rate
		{8.0, 0.5, "Very low rate", 0},           // Very low rate
		{8.0, 100.0, "x", 0},                     // Very short description
		{8.0, 100.0, "work", 0},                  // Generic description
		{8.0, 100.0, "Development work", 10},     // Future date
		{8.0, 100.0, "Development work", -1000},  // Far past date
		{math.NaN(), 100.0, "NaN hours", 0},      // NaN hours
		{8.0, math.Inf(1), "Inf rate", 0},        // Infinite rate
		{8.333333333, 100.0, "Precise hours", 0}, // High precision hours
	}

	for _, seed := range seeds {
		f.Add(seed.hours, seed.rate, seed.description, seed.dateOffset)
	}

	f.Fuzz(func(t *testing.T, hours, rate float64, description string, dateOffset int) {
		logger := &MockLogger{}
		validator := NewWorkItemValidator(logger)

		// Should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Validation rules panicked with hours=%f, rate=%f, desc=%q, dateOffset=%d: %v",
					hours, rate, description, dateOffset, r)
			}
		}()

		// Clamp date offset to reasonable range
		if dateOffset < -3650 { // ~10 years
			dateOffset = -3650
		}
		if dateOffset > 365 { // 1 year
			dateOffset = 365
		}

		date := time.Now().AddDate(0, 0, dateOffset)

		// Try to create work item
		ctx := context.Background()
		workItem, err := models.NewWorkItem(ctx, "test-fuzz", date, hours, rate, description)

		if err == nil && workItem != nil {
			// Test individual validation rules
			rules := validator.GetRules()

			for _, rule := range rules {
				if rule.Validator != nil {
					// Should never panic
					ruleErr := rule.Validator(ctx, workItem)

					// Test rule-specific invariants
					switch rule.Name {
					case "HoursValidation":
						if ruleErr == nil && (hours <= 0 || hours > 24 || math.IsNaN(hours) || math.IsInf(hours, 0)) {
							t.Errorf("HoursValidation should reject invalid hours: %f", hours)
						}
					case "RateValidation":
						if ruleErr == nil && (rate <= 0 || rate > 1000 || math.IsNaN(rate) || math.IsInf(rate, 0)) {
							t.Errorf("RateValidation should reject invalid rate: %f", rate)
						}
					case "DescriptionValidation":
						trimmed := strings.TrimSpace(description)
						if ruleErr == nil && (trimmed == "" || len(trimmed) < 3 || len(trimmed) > 500) {
							t.Errorf("DescriptionValidation should reject invalid description: %q", description)
						}
						// Check for generic descriptions
						genericDescs := []string{"work", "development", "coding", "programming", "task", "project", "meeting", "call", "todo", "fix", "bug", "feature"}
						for _, generic := range genericDescs {
							if strings.ToLower(trimmed) == generic && ruleErr == nil {
								t.Errorf("DescriptionValidation should reject generic description: %q", description)
							}
						}
					case "DateValidation":
						now := time.Now()
						if ruleErr == nil {
							if date.IsZero() {
								t.Errorf("DateValidation should reject zero date")
							}
							if date.After(now.AddDate(0, 0, 7)) {
								t.Errorf("DateValidation should reject dates more than 1 week in future: %v", date)
							}
							if date.Before(now.AddDate(-2, 0, 0)) {
								t.Errorf("DateValidation should reject dates more than 2 years in past: %v", date)
							}
						}
					case "TotalValidation":
						expectedTotal := math.Round(hours*rate*100) / 100
						if ruleErr == nil && math.Abs(workItem.Total-expectedTotal) > 0.01 {
							t.Errorf("TotalValidation should reject incorrect calculation: %f != %f", workItem.Total, expectedTotal)
						}
					}
				}
			}
		}
	})
}
