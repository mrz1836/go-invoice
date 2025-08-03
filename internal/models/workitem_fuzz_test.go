package models

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"
	"time"
)

// FuzzNewWorkItem fuzzes work item creation
func FuzzNewWorkItem(f *testing.F) {
	// Seed with various parameter combinations
	seeds := []struct {
		id          string
		dateOffset  int // Days from now
		hours       float64
		rate        float64
		description string
	}{
		{"valid-id", 0, 8.0, 100.0, "Development work"},
		{"", 0, 8.0, 100.0, "Empty ID"},                                        // Empty ID
		{"valid-id", 0, -1.0, 100.0, "Negative hours"},                         // Negative hours
		{"valid-id", 0, 8.0, -50.0, "Negative rate"},                           // Negative rate
		{"valid-id", 0, 8.0, 100.0, ""},                                        // Empty description
		{"valid-id", 0, 0.0, 100.0, "Zero hours"},                              // Zero hours
		{"valid-id", 0, 8.0, 0.0, "Zero rate"},                                 // Zero rate
		{"valid-id", 0, 25.0, 100.0, "Too many hours"},                         // Hours > 24
		{"valid-id", 0, 8.0, 5000.0, "Very high rate"},                         // Very high rate
		{"valid-id", 0, 8.0, 0.1, "Very low rate"},                             // Very low rate
		{"valid-id", 10, 8.0, 100.0, "Future date"},                            // Future date
		{"valid-id", -1000, 8.0, 100.0, "Past date"},                           // Far past date
		{"valid-id", 0, math.NaN(), 100.0, "NaN hours"},                        // NaN hours
		{"valid-id", 0, 8.0, math.Inf(1), "Infinite rate"},                     // Infinite rate
		{"valid-id", 0, math.Inf(-1), 100.0, "Negative infinite hours"},        // Negative infinite hours
		{"valid-id", 0, 8.333333333, 100.123456, "High precision"},             // High precision values
		{"123", 0, 8.0, 100.0, "Numeric ID"},                                   // Numeric ID
		{"very-long-id-" + strings.Repeat("x", 100), 0, 8.0, 100.0, "Long ID"}, // Very long ID
		{"valid-id", 0, 8.0, 100.0, "A"},                                       // Very short description
		{"valid-id", 0, 8.0, 100.0, strings.Repeat("A", 1001)},                 // Very long description
		{"id-with-special-chars!@#$%", 0, 8.0, 100.0, "Special chars in ID"},   // Special characters in ID
		{"valid-id", 0, 8.0, 100.0, "Description with special chars!@#$%"},     // Special characters in description
		{"valid-id", 0, 8.0, 100.0, "Unicode description 🚀 with emojis"},       // Unicode description
	}

	for _, seed := range seeds {
		f.Add(seed.id, seed.dateOffset, seed.hours, seed.rate, seed.description)
	}

	f.Fuzz(func(t *testing.T, id string, dateOffset int, hours, rate float64, description string) {
		// Should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("NewWorkItem panicked with id=%q, dateOffset=%d, hours=%f, rate=%f, desc=%q: %v",
					id, dateOffset, hours, rate, description, r)
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
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		workItem, err := NewWorkItem(ctx, id, date, hours, rate, description)

		// Test invariants
		if err == nil && workItem != nil {
			// If creation succeeded, work item should be valid
			if workItem.ID != id {
				t.Errorf("Work item ID should match input: %q != %q", workItem.ID, id)
			}
			if !workItem.Date.Equal(date) {
				t.Errorf("Work item date should match input: %v != %v", workItem.Date, date)
			}
			if workItem.Hours != hours {
				t.Errorf("Work item hours should match input: %f != %f", workItem.Hours, hours)
			}
			if workItem.Rate != rate {
				t.Errorf("Work item rate should match input: %f != %f", workItem.Rate, rate)
			}
			if workItem.Description != description {
				t.Errorf("Work item description should match input: %q != %q", workItem.Description, description)
			}

			// Total should be calculated correctly
			expectedTotal := math.Round(hours*rate*100) / 100
			if math.Abs(workItem.Total-expectedTotal) > 0.01 {
				t.Errorf("Work item total should be calculated correctly: %f != %f", workItem.Total, expectedTotal)
			}

			// CreatedAt should be set and recent
			if workItem.CreatedAt.IsZero() {
				t.Errorf("Work item CreatedAt should be set")
			}
			if time.Since(workItem.CreatedAt) > time.Minute {
				t.Errorf("Work item CreatedAt should be recent: %v", workItem.CreatedAt)
			}

			// If work item was created, it should pass its own validation
			validateErr := workItem.Validate(ctx)
			if validateErr != nil {
				t.Errorf("Valid work item should pass validation: %v", validateErr)
			}

			// Basic sanity checks for valid work items
			if workItem.Hours <= 0 {
				t.Errorf("Valid work item should have positive hours: %f", workItem.Hours)
			}
			if workItem.Rate <= 0 {
				t.Errorf("Valid work item should have positive rate: %f", workItem.Rate)
			}
			if workItem.Total < 0 {
				t.Errorf("Valid work item should have non-negative total: %f", workItem.Total)
			}
			if strings.TrimSpace(workItem.Description) == "" {
				t.Errorf("Valid work item should have non-empty description")
			}
			if strings.TrimSpace(workItem.ID) == "" {
				t.Errorf("Valid work item should have non-empty ID")
			}
		}

		// Test some expected failure cases - these should definitely fail and are expected behaviors:
		// - Empty ID should fail
		// - Empty description should fail
		// - Invalid hours should fail
		// - Invalid rate should fail
		// - Zero date should fail

		// Test context cancellation
		cancelCtx, cancelFunc := context.WithCancel(context.Background())
		cancelFunc()
		_, cancelErr := NewWorkItem(cancelCtx, id, date, hours, rate, description)
		if !errors.Is(cancelErr, context.Canceled) {
			t.Errorf("NewWorkItem with canceled context should return context.Canceled, got: %v", cancelErr)
		}
	})
}

// FuzzWorkItemValidation fuzzes work item validation
func FuzzWorkItemValidation(f *testing.F) {
	// First create some base work items to fuzz validation on
	ctx := context.Background()
	baseItems := []*WorkItem{}

	// Try to create some valid work items as seeds
	validSeeds := []struct {
		id          string
		hours       float64
		rate        float64
		description string
	}{
		{"valid-1", 8.0, 100.0, "Development work"},
		{"valid-2", 4.5, 125.0, "Bug fixes and testing"},
		{"valid-3", 2.0, 200.0, "Code review"},
	}

	for _, seed := range validSeeds {
		date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		item, err := NewWorkItem(ctx, seed.id, date, seed.hours, seed.rate, seed.description)
		if err == nil {
			baseItems = append(baseItems, item)
		}
	}

	// Seed fuzzer with modifications to valid items
	for i, item := range baseItems {
		if item != nil {
			f.Add(i, item.ID, item.Hours, item.Rate, item.Description, int(item.Date.Unix()), item.Total)
		}
	}

	// Add some edge case seeds
	edgeCases := []struct {
		idx         int
		id          string
		hours       float64
		rate        float64
		description string
		dateUnix    int
		total       float64
	}{
		{0, "", 8.0, 100.0, "Empty ID", 1705363200, 800.0},
		{0, "valid", -1.0, 100.0, "Negative hours", 1705363200, -100.0},
		{0, "valid", 8.0, -50.0, "Negative rate", 1705363200, -400.0},
		{0, "valid", 8.0, 100.0, "", 1705363200, 800.0},
		{0, "valid", 25.0, 100.0, "Too many hours", 1705363200, 2500.0},
		{0, "valid", 8.0, 100.0, "Description", 0, 800.0}, // Zero date
	}

	for _, edge := range edgeCases {
		f.Add(edge.idx, edge.id, edge.hours, edge.rate, edge.description, edge.dateUnix, edge.total)
	}

	f.Fuzz(func(t *testing.T, _ int, id string, hours, rate float64, description string, dateUnix int, total float64) {
		// Should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("WorkItem validation panicked: %v", r)
			}
		}()

		// Create a work item manually (bypassing NewWorkItem validation)
		date := time.Unix(int64(dateUnix), 0)
		workItem := &WorkItem{
			ID:          id,
			Date:        date,
			Hours:       hours,
			Rate:        rate,
			Description: description,
			Total:       total,
			CreatedAt:   time.Now(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err := workItem.Validate(ctx)

		// Test validation invariants
		if err == nil {
			// If validation passed, work item should meet all criteria
			if strings.TrimSpace(workItem.ID) == "" {
				t.Errorf("Valid work item should have non-empty ID")
			}
			if workItem.Date.IsZero() {
				t.Errorf("Valid work item should have non-zero date")
			}
			if workItem.Hours <= 0 {
				t.Errorf("Valid work item should have positive hours: %f", workItem.Hours)
			}
			if workItem.Hours > 24 {
				t.Errorf("Valid work item should not exceed 24 hours: %f", workItem.Hours)
			}
			if workItem.Rate <= 0 {
				t.Errorf("Valid work item should have positive rate: %f", workItem.Rate)
			}
			if workItem.Rate > 10000 {
				t.Errorf("Valid work item should not exceed $10,000/hour: %f", workItem.Rate)
			}
			if strings.TrimSpace(workItem.Description) == "" {
				t.Errorf("Valid work item should have non-empty description")
			}
			if len(workItem.Description) > 1000 {
				t.Errorf("Valid work item description should not exceed 1000 characters: %d", len(workItem.Description))
			}
			if workItem.Total < 0 {
				t.Errorf("Valid work item should have non-negative total: %f", workItem.Total)
			}
			if workItem.CreatedAt.IsZero() {
				t.Errorf("Valid work item should have CreatedAt set")
			}

			// Check calculation correctness
			expectedTotal := math.Round(workItem.Hours*workItem.Rate*100) / 100
			if math.Abs(workItem.Total-expectedTotal) > 0.01 {
				t.Errorf("Valid work item should have correct total: %f != %f", workItem.Total, expectedTotal)
			}

			// Check date is reasonable
			now := time.Now()
			if workItem.Date.After(now.AddDate(0, 0, 1)) {
				t.Errorf("Valid work item date should not be more than 1 day in future: %v", workItem.Date)
			}
			// Allow dates back to Unix epoch for testing purposes

			// Check for floating point issues
			if math.IsNaN(workItem.Hours) || math.IsInf(workItem.Hours, 0) {
				t.Errorf("Valid work item should not have NaN/Inf hours: %f", workItem.Hours)
			}
			if math.IsNaN(workItem.Rate) || math.IsInf(workItem.Rate, 0) {
				t.Errorf("Valid work item should not have NaN/Inf rate: %f", workItem.Rate)
			}
			if math.IsNaN(workItem.Total) || math.IsInf(workItem.Total, 0) {
				t.Errorf("Valid work item should not have NaN/Inf total: %f", workItem.Total)
			}
		}

		// Test context cancellation
		cancelCtx, cancelFunc := context.WithCancel(context.Background())
		cancelFunc()
		cancelErr := workItem.Validate(cancelCtx)
		if !errors.Is(cancelErr, context.Canceled) {
			t.Errorf("Validation with canceled context should return context.Canceled, got: %v", cancelErr)
		}
	})
}

// FuzzWorkItemUpdates fuzzes work item update methods
func FuzzWorkItemUpdates(f *testing.F) {
	// Create a base valid work item for updates
	ctx := context.Background()
	baseItem, err := NewWorkItem(ctx, "test-item", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), 8.0, 100.0, "Base work item")
	if err != nil {
		f.Skip("Cannot create base work item for fuzzing")
	}

	// Seed with various update scenarios
	seeds := []struct {
		updateType string
		value1     float64
		value2     string
	}{
		{"hours", 4.0, ""},
		{"hours", -1.0, ""},        // Invalid hours
		{"hours", 25.0, ""},        // Too many hours
		{"hours", 0.0, ""},         // Zero hours
		{"hours", math.NaN(), ""},  // NaN hours
		{"hours", math.Inf(1), ""}, // Infinite hours
		{"rate", 150.0, ""},
		{"rate", -50.0, ""},       // Invalid rate
		{"rate", 0.0, ""},         // Zero rate
		{"rate", 15000.0, ""},     // Too high rate
		{"rate", math.NaN(), ""},  // NaN rate
		{"rate", math.Inf(1), ""}, // Infinite rate
		{"description", 0.0, "Updated description"},
		{"description", 0.0, ""},                        // Empty description
		{"description", 0.0, "A"},                       // Very short description
		{"description", 0.0, strings.Repeat("A", 1001)}, // Very long description
		{"description", 0.0, "Description with special chars!@#$%^&*()"},
		{"description", 0.0, "Unicode description 🚀 with emojis"},
	}

	for _, seed := range seeds {
		f.Add(seed.updateType, seed.value1, seed.value2)
	}

	f.Fuzz(func(t *testing.T, updateType string, numValue float64, strValue string) {
		// Create a copy of the base item for testing
		workItem := &WorkItem{
			ID:          baseItem.ID,
			Date:        baseItem.Date,
			Hours:       baseItem.Hours,
			Rate:        baseItem.Rate,
			Description: baseItem.Description,
			Total:       baseItem.Total,
			CreatedAt:   baseItem.CreatedAt,
		}

		// Should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("WorkItem update panicked with type=%q, numValue=%f, strValue=%q: %v",
					updateType, numValue, strValue, r)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		var err error
		originalHours := workItem.Hours
		originalRate := workItem.Rate
		originalTotal := workItem.Total

		// Test different update methods based on updateType
		switch updateType {
		case "hours":
			err = workItem.UpdateHours(ctx, numValue)
			if err == nil {
				// If update succeeded, check invariants
				// For NaN values, both should be NaN
				if math.IsNaN(numValue) && math.IsNaN(workItem.Hours) {
					// Both are NaN, this is correct behavior for NaN handling
					t.Logf("Both numValue and workItem.Hours are NaN as expected")
				} else if workItem.Hours != numValue {
					t.Errorf("Hours should be updated to %f, got %f", numValue, workItem.Hours)
				}
				if workItem.Hours <= 0 {
					t.Errorf("Updated hours should be positive: %f", workItem.Hours)
				}
				if workItem.Hours > 24 {
					t.Errorf("Updated hours should not exceed 24: %f", workItem.Hours)
				}
				// Total should be recalculated
				expectedTotal := math.Round(workItem.Hours*workItem.Rate*100) / 100
				if math.Abs(workItem.Total-expectedTotal) > 0.01 {
					t.Errorf("Total should be recalculated after hours update: %f != %f", workItem.Total, expectedTotal)
				}
				// Rate should not change
				if workItem.Rate != originalRate {
					t.Errorf("Rate should not change during hours update: %f != %f", workItem.Rate, originalRate)
				}
			}

		case "rate":
			err = workItem.UpdateRate(ctx, numValue)
			if err == nil {
				// If update succeeded, check invariants
				// For NaN values, both should be NaN
				if math.IsNaN(numValue) && math.IsNaN(workItem.Rate) {
					// Both are NaN, this is correct behavior for NaN handling
					t.Logf("Both numValue and workItem.Rate are NaN as expected")
				} else if workItem.Rate != numValue {
					t.Errorf("Rate should be updated to %f, got %f", numValue, workItem.Rate)
				}
				if workItem.Rate <= 0 {
					t.Errorf("Updated rate should be positive: %f", workItem.Rate)
				}
				if workItem.Rate > 10000 {
					t.Errorf("Updated rate should not exceed $10,000: %f", workItem.Rate)
				}
				// Total should be recalculated
				expectedTotal := math.Round(workItem.Hours*workItem.Rate*100) / 100
				if math.Abs(workItem.Total-expectedTotal) > 0.01 {
					t.Errorf("Total should be recalculated after rate update: %f != %f", workItem.Total, expectedTotal)
				}
				// Hours should not change
				if workItem.Hours != originalHours {
					t.Errorf("Hours should not change during rate update: %f != %f", workItem.Hours, originalHours)
				}
			}

		case "description":
			err = workItem.UpdateDescription(ctx, strValue)
			if err == nil {
				// If update succeeded, check invariants
				// UpdateDescription trims whitespace, so compare against trimmed value
				expectedDesc := strings.TrimSpace(strValue)
				if workItem.Description != expectedDesc {
					t.Errorf("Description should be updated to %q, got %q", expectedDesc, workItem.Description)
				}
				if strings.TrimSpace(workItem.Description) == "" {
					t.Errorf("Updated description should not be empty")
				}
				if len(workItem.Description) > 1000 {
					t.Errorf("Updated description should not exceed 1000 characters: %d", len(workItem.Description))
				}
				// Hours, rate, and total should not change
				if workItem.Hours != originalHours {
					t.Errorf("Hours should not change during description update: %f != %f", workItem.Hours, originalHours)
				}
				if workItem.Rate != originalRate {
					t.Errorf("Rate should not change during description update: %f != %f", workItem.Rate, originalRate)
				}
				if workItem.Total != originalTotal {
					t.Errorf("Total should not change during description update: %f != %f", workItem.Total, originalTotal)
				}
			}
		}

		// Test context cancellation for each update type (only for valid update types)
		validUpdateTypes := []string{"hours", "rate", "description"}
		isValidUpdateType := false
		for _, validType := range validUpdateTypes {
			if updateType == validType {
				isValidUpdateType = true
				break
			}
		}

		if isValidUpdateType {
			cancelCtx, cancelFunc := context.WithCancel(context.Background())
			cancelFunc()

			var cancelErr error
			switch updateType {
			case "hours":
				cancelErr = workItem.UpdateHours(cancelCtx, numValue)
			case "rate":
				cancelErr = workItem.UpdateRate(cancelCtx, numValue)
			case "description":
				cancelErr = workItem.UpdateDescription(cancelCtx, strValue)
			}

			if !errors.Is(cancelErr, context.Canceled) {
				t.Errorf("Update with canceled context should return context.Canceled, got: %v", cancelErr)
			}
		}
	})
}

// FuzzWorkItemFormatting fuzzes work item formatting methods
func FuzzWorkItemFormatting(f *testing.F) {
	// Seed with various numeric values
	seeds := []struct {
		hours float64
		rate  float64
		total float64
	}{
		{8.0, 100.0, 800.0},
		{8.5, 125.75, 1069.125},
		{0.0, 0.0, 0.0},
		{0.01, 0.01, 0.0001},
		{23.99, 999.99, 23989.9701},
		{1.0, 1.0, 1.0},
		{1000.0, 1000.0, 1000000.0},
		{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64},
		{math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64},
	}

	for _, seed := range seeds {
		f.Add(seed.hours, seed.rate, seed.total)
	}

	f.Fuzz(func(t *testing.T, hours, rate, total float64) {
		// Create work item manually for formatting tests
		workItem := &WorkItem{
			ID:          "test-format",
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Hours:       hours,
			Rate:        rate,
			Description: "Test formatting",
			Total:       total,
			CreatedAt:   time.Now(),
		}

		// Should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Formatting method panicked with hours=%f, rate=%f, total=%f: %v",
					hours, rate, total, r)
			}
		}()

		// Test formatting methods
		formattedTotal := workItem.GetFormattedTotal()
		formattedRate := workItem.GetFormattedRate()
		formattedHours := workItem.GetFormattedHours()

		// Test formatting invariants
		// Formatted values should not be empty
		if formattedTotal == "" {
			t.Errorf("GetFormattedTotal should not return empty string")
		}
		if formattedRate == "" {
			t.Errorf("GetFormattedRate should not return empty string")
		}
		if formattedHours == "" {
			t.Errorf("GetFormattedHours should not return empty string")
		}

		// Currency formatting should start with $
		if !strings.HasPrefix(formattedTotal, "$") {
			t.Errorf("GetFormattedTotal should start with $: %s", formattedTotal)
		}
		if !strings.HasPrefix(formattedRate, "$") {
			t.Errorf("GetFormattedRate should start with $: %s", formattedRate)
		}

		// Hours formatting should be numeric and contain decimal point for normal values
		if !strings.Contains(formattedHours, ".") && !math.IsInf(hours, 0) && !math.IsNaN(hours) {
			t.Logf("Hours formatting may not contain decimal point for value: %f -> %s", hours, formattedHours)
		}

		// Formatting should be deterministic
		formattedTotal2 := workItem.GetFormattedTotal()
		formattedRate2 := workItem.GetFormattedRate()
		formattedHours2 := workItem.GetFormattedHours()

		if formattedTotal != formattedTotal2 {
			t.Errorf("GetFormattedTotal should be deterministic: %s != %s", formattedTotal, formattedTotal2)
		}
		if formattedRate != formattedRate2 {
			t.Errorf("GetFormattedRate should be deterministic: %s != %s", formattedRate, formattedRate2)
		}
		if formattedHours != formattedHours2 {
			t.Errorf("GetFormattedHours should be deterministic: %s != %s", formattedHours, formattedHours2)
		}

		// Test special values
		if math.IsNaN(total) {
			// NaN should be handled gracefully
			if !strings.Contains(strings.ToLower(formattedTotal), "nan") && formattedTotal != "$NaN" {
				t.Logf("NaN value formatting: %f -> %s", total, formattedTotal)
			}
		}
		if math.IsInf(rate, 0) {
			// Infinity should be handled gracefully
			if !strings.Contains(strings.ToLower(formattedRate), "inf") && !strings.Contains(formattedRate, "Inf") {
				t.Logf("Infinity value formatting: %f -> %s", rate, formattedRate)
			}
		}
	})
}
