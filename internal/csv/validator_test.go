package csv

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mrz/go-invoice/internal/models"
)

// WorkItemValidatorTestSuite defines the test suite for work item validator functionality
type WorkItemValidatorTestSuite struct {
	suite.Suite

	validator *WorkItemValidator
	logger    *MockLogger
}

// SetupTest runs before each test
func (suite *WorkItemValidatorTestSuite) SetupTest() {
	suite.logger = &MockLogger{}
	suite.validator = NewWorkItemValidator(suite.logger)
}

// TestNewWorkItemValidator tests the constructor
func (suite *WorkItemValidatorTestSuite) TestNewWorkItemValidator() {
	validator := NewWorkItemValidator(suite.logger)
	suite.NotNil(validator)
	suite.Equal(suite.logger, validator.logger)
	suite.NotEmpty(validator.rules, "should have validation rules")
}

// TestValidateWorkItemSuccess tests successful work item validation
func (suite *WorkItemValidatorTestSuite) TestValidateWorkItemSuccess() {
	ctx := context.Background()
	workItem, err := models.NewWorkItem(ctx, "test-123", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), 8.0, 100.0, "Development work")
	suite.Require().NoError(err)

	err = suite.validator.ValidateWorkItem(ctx, workItem)
	suite.Require().NoError(err)
}

// TestValidateWorkItemContextCancellation tests context cancellation
func (suite *WorkItemValidatorTestSuite) TestValidateWorkItemContextCancellation() {
	workItem, err := models.NewWorkItem(context.Background(), "test-123", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), 8.0, 100.0, "Development work")
	suite.Require().NoError(err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = suite.validator.ValidateWorkItem(ctx, workItem)
	suite.Require().Error(err)
	suite.Equal(context.Canceled, err)
}

// TestValidateWorkItemNil tests validation of nil work item
func (suite *WorkItemValidatorTestSuite) TestValidateWorkItemNil() {
	ctx := context.Background()
	err := suite.validator.ValidateWorkItem(ctx, nil)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "work item cannot be nil")
}

// TestValidateWorkItemDateValidation tests date validation rules
func (suite *WorkItemValidatorTestSuite) TestValidateWorkItemDateValidation() {
	tests := []struct {
		name    string
		date    time.Time
		wantErr bool
		errMsg  string
	}{
		{
			name:    "ValidRecentDate",
			date:    time.Now().AddDate(0, 0, -1), // Yesterday
			wantErr: false,
		},
		{
			name:    "ValidToday",
			date:    time.Now(),
			wantErr: false,
		},
		{
			name:    "TooFarInFuture",
			date:    time.Now().AddDate(0, 0, 2), // 2 days in future - models validation allows only 1 day
			wantErr: true,
			errMsg:  "more than 1 day in the future",
		},
		{
			name:    "TooFarInPast",
			date:    time.Now().AddDate(-3, 0, 0), // 3 years ago
			wantErr: true,
			errMsg:  "too far in the past",
		},
		{
			name:    "ZeroDate",
			date:    time.Time{},
			wantErr: true,
			errMsg:  "work date cannot be empty",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			var workItem *models.WorkItem

			// For dates that would fail models.NewWorkItem validation, create manually
			if tt.date.IsZero() || tt.date.After(time.Now().AddDate(0, 0, 1)) || tt.date.Before(time.Now().AddDate(-2, 0, 0)) {
				workItem = &models.WorkItem{
					ID:          "test-123",
					Date:        tt.date,
					Hours:       8.0,
					Rate:        100.0,
					Description: "Development work",
					Total:       800.0,
					CreatedAt:   time.Now(),
				}
			} else {
				// Use models.NewWorkItem for valid dates
				var err error
				workItem, err = models.NewWorkItem(ctx, "test-123", tt.date, 8.0, 100.0, "Development work")
				suite.Require().NoError(err)
			}

			err := suite.validator.ValidateWorkItem(ctx, workItem)

			if tt.wantErr {
				suite.Require().Error(err)
				if tt.errMsg != "" {
					suite.Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

// TestValidateWorkItemHoursValidation tests hours validation rules
func (suite *WorkItemValidatorTestSuite) TestValidateWorkItemHoursValidation() {
	tests := []struct {
		name    string
		hours   float64
		wantErr bool
		errMsg  string
	}{
		{
			name:    "ValidHours",
			hours:   8.0,
			wantErr: false,
		},
		{
			name:    "ValidPartialHours",
			hours:   7.5,
			wantErr: false,
		},
		{
			name:    "ZeroHours",
			hours:   0.0,
			wantErr: true,
			errMsg:  "hours must be positive",
		},
		{
			name:    "NegativeHours",
			hours:   -1.0,
			wantErr: true,
			errMsg:  "hours must be positive",
		},
		{
			name:    "TooManyHours",
			hours:   25.0,
			wantErr: true,
			errMsg:  "hours cannot exceed 24",
		},
		{
			name:    "TooManyDecimalPlaces",
			hours:   8.123456,
			wantErr: true,
			errMsg:  "should not have more than 2 decimal places",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			workItem := &models.WorkItem{
				ID:          "test-123",
				Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				Hours:       tt.hours,
				Rate:        100.0,
				Description: "Development work",
				Total:       tt.hours * 100.0,
				CreatedAt:   time.Now(),
			}

			err := suite.validator.ValidateWorkItem(ctx, workItem)

			if tt.wantErr {
				suite.Require().Error(err)
				if tt.errMsg != "" {
					suite.Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

// TestValidateWorkItemRateValidation tests rate validation rules
func (suite *WorkItemValidatorTestSuite) TestValidateWorkItemRateValidation() {
	tests := []struct {
		name    string
		rate    float64
		wantErr bool
		errMsg  string
	}{
		{
			name:    "ValidRate",
			rate:    100.0,
			wantErr: false,
		},
		{
			name:    "HighButValidRate",
			rate:    500.0,
			wantErr: false,
		},
		{
			name:    "ZeroRate",
			rate:    0.0,
			wantErr: true,
			errMsg:  "hourly rate must be positive",
		},
		{
			name:    "NegativeRate",
			rate:    -10.0,
			wantErr: true,
			errMsg:  "hourly rate must be positive",
		},
		{
			name:    "TooLowRate",
			rate:    0.5,
			wantErr: true,
			errMsg:  "hourly rate seems too low",
		},
		{
			name:    "TooHighRate",
			rate:    1001.0,
			wantErr: true,
			errMsg:  "hourly rate seems too high",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			workItem := &models.WorkItem{
				ID:          "test-123",
				Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				Hours:       8.0,
				Rate:        tt.rate,
				Description: "Development work",
				Total:       8.0 * tt.rate,
				CreatedAt:   time.Now(),
			}

			err := suite.validator.ValidateWorkItem(ctx, workItem)

			if tt.wantErr {
				suite.Require().Error(err)
				if tt.errMsg != "" {
					suite.Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

// TestValidateWorkItemDescriptionValidation tests description validation rules
func (suite *WorkItemValidatorTestSuite) TestValidateWorkItemDescriptionValidation() {
	tests := []struct {
		name        string
		description string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "ValidDescription",
			description: "Development work on user authentication",
			wantErr:     false,
		},
		{
			name:        "EmptyDescription",
			description: "",
			wantErr:     true,
			errMsg:      "work description cannot be empty",
		},
		{
			name:        "WhitespaceOnlyDescription",
			description: "   ",
			wantErr:     true,
			errMsg:      "work description cannot be empty",
		},
		{
			name:        "TooShortDescription",
			description: "ab",
			wantErr:     true,
			errMsg:      "work description too short",
		},
		{
			name:        "TooLongDescription",
			description: strings.Repeat("a", 501),
			wantErr:     true,
			errMsg:      "work description too long",
		},
		{
			name:        "GenericDescription",
			description: "work",
			wantErr:     true,
			errMsg:      "work description too generic",
		},
		{
			name:        "AnotherGenericDescription",
			description: "development",
			wantErr:     true,
			errMsg:      "work description too generic",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			workItem := &models.WorkItem{
				ID:          "test-123",
				Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				Hours:       8.0,
				Rate:        100.0,
				Description: tt.description,
				Total:       800.0,
				CreatedAt:   time.Now(),
			}

			err := suite.validator.ValidateWorkItem(ctx, workItem)

			if tt.wantErr {
				suite.Require().Error(err)
				if tt.errMsg != "" {
					suite.Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

// TestValidateWorkItemTotalValidation tests total calculation validation
func (suite *WorkItemValidatorTestSuite) TestValidateWorkItemTotalValidation() {
	tests := []struct {
		name    string
		hours   float64
		rate    float64
		total   float64
		wantErr bool
		errMsg  string
	}{
		{
			name:    "CorrectTotal",
			hours:   8.0,
			rate:    100.0,
			total:   800.0,
			wantErr: false,
		},
		{
			name:    "CorrectTotalWithDecimals",
			hours:   7.5,
			rate:    125.50,
			total:   941.25,
			wantErr: false,
		},
		{
			name:    "IncorrectTotal",
			hours:   8.0,
			rate:    100.0,
			total:   900.0, // Wrong total
			wantErr: true,
			errMsg:  "total amount",
		},
		{
			name:    "NegativeTotal",
			hours:   8.0,
			rate:    100.0,
			total:   -800.0,
			wantErr: true,
			errMsg:  "total amount",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			workItem := &models.WorkItem{
				ID:          "test-123",
				Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				Hours:       tt.hours,
				Rate:        tt.rate,
				Description: "Development work",
				Total:       tt.total,
				CreatedAt:   time.Now(),
			}

			err := suite.validator.ValidateWorkItem(ctx, workItem)

			if tt.wantErr {
				suite.Require().Error(err)
				if tt.errMsg != "" {
					suite.Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

// TestValidateRowSuccess tests successful row validation
func (suite *WorkItemValidatorTestSuite) TestValidateRowSuccess() {
	ctx := context.Background()
	row := []string{"2024-01-15", "8.0", "100.00", "Development work"}

	err := suite.validator.ValidateRow(ctx, row, 1)
	suite.Require().NoError(err)
}

// TestValidateRowContextCancellation tests row validation with context cancellation
func (suite *WorkItemValidatorTestSuite) TestValidateRowContextCancellation() {
	row := []string{"2024-01-15", "8.0", "100.00", "Development work"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := suite.validator.ValidateRow(ctx, row, 1)
	suite.Require().Error(err)
	suite.Equal(context.Canceled, err)
}

// TestValidateRowErrors tests various row validation errors
func (suite *WorkItemValidatorTestSuite) TestValidateRowErrors() {
	tests := []struct {
		name    string
		row     []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "ValidRow",
			row:     []string{"2024-01-15", "8.0", "100.00", "Development work"},
			wantErr: false,
		},
		{
			name:    "EmptyRow",
			row:     []string{},
			wantErr: true,
			errMsg:  "row is empty",
		},
		{
			name:    "AllEmptyFields",
			row:     []string{"", "", "", ""},
			wantErr: true,
			errMsg:  "row contains no data",
		},
		{
			name:    "WhitespaceOnlyRow",
			row:     []string{"  ", "  ", "  ", "  "},
			wantErr: true,
			errMsg:  "row contains no data",
		},
		{
			name:    "TooFewFields",
			row:     []string{"2024-01-15", "8.0"},
			wantErr: true,
			errMsg:  "has 2 fields, expected at least 4",
		},
		{
			name:    "TooManyFields",
			row:     append([]string{"2024-01-15", "8.0", "100.00", "Development work"}, make([]string, 21)...), // 25 fields total
			wantErr: true,
			errMsg:  "has 25 fields, which seems excessive",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			err := suite.validator.ValidateRow(ctx, tt.row, 1)

			if tt.wantErr {
				suite.Require().Error(err)
				if tt.errMsg != "" {
					suite.Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

// TestValidateBatchSuccess tests successful batch validation
func (suite *WorkItemValidatorTestSuite) TestValidateBatchSuccess() {
	ctx := context.Background()
	workItems := []models.WorkItem{
		{
			ID:          "test-1",
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Hours:       8.0,
			Rate:        100.0,
			Description: "Development work",
			Total:       800.0,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "test-2",
			Date:        time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
			Hours:       6.0,
			Rate:        100.0,
			Description: "Testing and debugging",
			Total:       600.0,
			CreatedAt:   time.Now(),
		},
	}

	err := suite.validator.ValidateBatch(ctx, workItems)
	suite.Require().NoError(err)
}

// TestValidateBatchContextCancellation tests batch validation with context cancellation
func (suite *WorkItemValidatorTestSuite) TestValidateBatchContextCancellation() {
	workItems := []models.WorkItem{
		{
			ID:          "test-1",
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Hours:       8.0,
			Rate:        100.0,
			Description: "Development work",
			Total:       800.0,
			CreatedAt:   time.Now(),
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := suite.validator.ValidateBatch(ctx, workItems)
	suite.Require().Error(err)
	suite.Equal(context.Canceled, err)
}

// TestValidateBatchErrors tests various batch validation errors
func (suite *WorkItemValidatorTestSuite) TestValidateBatchErrors() {
	tests := []struct {
		name      string
		workItems []models.WorkItem
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "EmptyBatch",
			workItems: []models.WorkItem{},
			wantErr:   true,
			errMsg:    "no work items to validate",
		},
		{
			name: "InvalidWorkItem",
			workItems: []models.WorkItem{
				{
					ID:          "", // Invalid empty ID
					Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
					Hours:       8.0,
					Rate:        100.0,
					Description: "Development work",
					Total:       800.0,
					CreatedAt:   time.Now(),
				},
			},
			wantErr: true,
			errMsg:  "work item 1 validation failed",
		},
		{
			name: "ZeroTotalHours",
			workItems: []models.WorkItem{
				{
					ID:          "test-1",
					Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
					Hours:       0.0, // Zero hours
					Rate:        100.0,
					Description: "Development work",
					Total:       0.0,
					CreatedAt:   time.Now(),
				},
			},
			wantErr: true,
			errMsg:  "work item 1 validation failed",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			err := suite.validator.ValidateBatch(ctx, tt.workItems)

			if tt.wantErr {
				suite.Require().Error(err)
				if tt.errMsg != "" {
					suite.Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

// TestValidateBatchDateRange tests date range validation in batches
func (suite *WorkItemValidatorTestSuite) TestValidateBatchDateRange() {
	ctx := context.Background()

	// Create work items with dates spanning more than 1 year, but within individual validation limits
	// Use dates from 1.5 years ago to 6 months ago to avoid individual validation issues
	baseDate := time.Now().AddDate(-1, -6, 0) // 1.5 years ago (within individual validation limit)
	workItems := []models.WorkItem{
		{
			ID:          "test-1",
			Date:        baseDate, // 1.5 years ago
			Hours:       8.0,
			Rate:        100.0,
			Description: "Old development work",
			Total:       800.0,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "test-2",
			Date:        time.Now().AddDate(0, -5, 0), // 5 months ago (more than 1 year later than first item)
			Hours:       8.0,
			Rate:        100.0,
			Description: "Recent development work",
			Total:       800.0,
			CreatedAt:   time.Now(),
		},
	}

	err := suite.validator.ValidateBatch(ctx, workItems)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "date range is too large")
}

// TestCustomRules tests adding and removing custom validation rules
func (suite *WorkItemValidatorTestSuite) TestCustomRules() {
	// Add a custom rule
	customRule := ValidationRule{
		Name:        "CustomTest",
		Description: "Test custom rule",
		Validator: func(_ context.Context, item *models.WorkItem) error {
			if item.Description == "forbidden" {
				return assert.AnError
			}
			return nil
		},
	}

	suite.validator.AddCustomRule(customRule)

	// Test that the custom rule is applied
	ctx := context.Background()
	workItem := &models.WorkItem{
		ID:          "test-1",
		Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Hours:       8.0,
		Rate:        100.0,
		Description: "forbidden",
		Total:       800.0,
		CreatedAt:   time.Now(),
	}

	err := suite.validator.ValidateWorkItem(ctx, workItem)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "CustomTest")

	// Remove the custom rule
	removed := suite.validator.RemoveRule("CustomTest")
	suite.True(removed)

	// Test that the rule is no longer applied
	err = suite.validator.ValidateWorkItem(ctx, workItem)
	suite.Require().NoError(err) // Should pass now since custom rule is removed

	// Try to remove non-existent rule
	removed = suite.validator.RemoveRule("NonExistent")
	suite.False(removed)
}

// TestGetRules tests getting all validation rules
func (suite *WorkItemValidatorTestSuite) TestGetRules() {
	rules := suite.validator.GetRules()
	suite.NotEmpty(rules)

	// Verify we get standard rules
	ruleNames := make(map[string]bool)
	for _, rule := range rules {
		ruleNames[rule.Name] = true
	}

	expectedRules := []string{
		"DateValidation",
		"HoursValidation",
		"RateValidation",
		"DescriptionValidation",
		"TotalValidation",
		"RowFormatValidation",
	}

	for _, expected := range expectedRules {
		suite.True(ruleNames[expected], "Expected rule %s not found", expected)
	}
}

// TestValidatorLogging tests that the validator logs appropriately
func (suite *WorkItemValidatorTestSuite) TestValidatorLogging() {
	ctx := context.Background()
	workItems := []models.WorkItem{
		{
			ID:          "test-1",
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Hours:       8.0,
			Rate:        100.0,
			Description: "Development work",
			Total:       800.0,
			CreatedAt:   time.Now(),
		},
	}

	err := suite.validator.ValidateBatch(ctx, workItems)
	suite.Require().NoError(err)

	// Check that logging occurred
	suite.NotEmpty(suite.logger.messages)

	// Find debug message
	var foundDebug bool
	for _, msg := range suite.logger.messages {
		if msg.Level == "DEBUG" && strings.Contains(msg.Msg, "validating work items batch") {
			foundDebug = true
			break
		}
	}
	suite.True(foundDebug, "Expected to find batch validation logging message")
}

// TestHighHoursWarning tests that high hours trigger debug logging
func (suite *WorkItemValidatorTestSuite) TestHighHoursWarning() {
	ctx := context.Background()
	workItem := &models.WorkItem{
		ID:          "test-1",
		Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Hours:       15.0, // High hours
		Rate:        100.0,
		Description: "Long development session",
		Total:       1500.0,
		CreatedAt:   time.Now(),
	}

	err := suite.validator.ValidateWorkItem(ctx, workItem)
	suite.Require().NoError(err) // Should still pass validation

	// Check that warning was logged
	var foundWarning bool
	for _, msg := range suite.logger.messages {
		if msg.Level == "DEBUG" && strings.Contains(msg.Msg, "unusually high hours detected") {
			foundWarning = true
			break
		}
	}
	suite.True(foundWarning, "Expected to find high hours warning")
}

// TestHighRateWarning tests that high rates trigger debug logging
func (suite *WorkItemValidatorTestSuite) TestHighRateWarning() {
	ctx := context.Background()
	workItem := &models.WorkItem{
		ID:          "test-1",
		Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Hours:       8.0,
		Rate:        600.0, // High rate
		Description: "Expert consultation",
		Total:       4800.0,
		CreatedAt:   time.Now(),
	}

	err := suite.validator.ValidateWorkItem(ctx, workItem)
	suite.Require().NoError(err) // Should still pass validation

	// Check that warning was logged
	var foundWarning bool
	for _, msg := range suite.logger.messages {
		if msg.Level == "DEBUG" && strings.Contains(msg.Msg, "unusually high rate detected") {
			foundWarning = true
			break
		}
	}
	suite.True(foundWarning, "Expected to find high rate warning")
}

// TestWorkItemValidatorTestSuite runs the work item validator test suite
func TestWorkItemValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(WorkItemValidatorTestSuite))
}

// Additional standalone tests for edge cases

// TestValidateRateConsistency tests rate consistency checking
func TestValidateRateConsistency(t *testing.T) {
	logger := &MockLogger{}
	validator := NewWorkItemValidator(logger)

	// Create work items with many different rates
	workItems := make([]models.WorkItem, 5)
	for i := 0; i < 5; i++ {
		workItems[i] = models.WorkItem{
			ID:          "test-" + string(rune('1'+i)),
			Date:        time.Date(2024, 1, 15+i, 0, 0, 0, 0, time.UTC),
			Hours:       8.0,
			Rate:        float64(100 + i*10), // Different rates: 100, 110, 120, 130, 140
			Description: "Development work",
			Total:       8.0 * float64(100+i*10),
			CreatedAt:   time.Now(),
		}
	}

	ctx := context.Background()
	err := validator.ValidateBatch(ctx, workItems)
	require.NoError(t, err) // Should pass validation

	// Check that rate inconsistency was logged
	var foundWarning bool
	for _, msg := range logger.messages {
		if msg.Level == "DEBUG" && strings.Contains(msg.Msg, "multiple different rates detected") {
			foundWarning = true
			break
		}
	}
	assert.True(t, foundWarning, "Expected to find rate inconsistency warning")
}

// TestValidateTotalHoursWarning tests total hours warning
func TestValidateTotalHoursWarning(t *testing.T) {
	logger := &MockLogger{}
	validator := NewWorkItemValidator(logger)

	// Create many work items to exceed total hours threshold
	workItems := make([]models.WorkItem, 30) // 30 items * 8 hours = 240 hours
	for i := 0; i < 30; i++ {
		workItems[i] = models.WorkItem{
			ID:          "test-" + string(rune('1'+i%10)),
			Date:        time.Date(2024, 1, 15+i, 0, 0, 0, 0, time.UTC),
			Hours:       8.0,
			Rate:        100.0,
			Description: "Development work",
			Total:       800.0,
			CreatedAt:   time.Now(),
		}
	}

	ctx := context.Background()
	err := validator.ValidateBatch(ctx, workItems)
	require.NoError(t, err) // Should pass validation

	// Check that high total hours was logged
	var foundWarning bool
	for _, msg := range logger.messages {
		if msg.Level == "DEBUG" && strings.Contains(msg.Msg, "large total hours detected") {
			foundWarning = true
			break
		}
	}
	assert.True(t, foundWarning, "Expected to find large total hours warning")
}
