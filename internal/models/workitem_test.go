package models

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type WorkItemTestSuite struct {
	suite.Suite

	ctx        context.Context //nolint:containedctx // Test suite context is acceptable
	cancelFunc context.CancelFunc
}

func (suite *WorkItemTestSuite) SetupTest() {
	suite.ctx, suite.cancelFunc = context.WithTimeout(context.Background(), 5*time.Second)
}

func (suite *WorkItemTestSuite) TearDownTest() {
	suite.cancelFunc()
}

func TestWorkItemTestSuite(t *testing.T) {
	suite.Run(t, new(WorkItemTestSuite))
}

func (suite *WorkItemTestSuite) TestNewWorkItem() {
	t := suite.T()

	tests := []struct {
		name          string
		id            string
		date          time.Time
		hours         float64
		rate          float64
		description   string
		expectError   bool
		errorMsg      string
		expectedTotal float64
	}{
		{
			name:          "ValidWorkItem",
			id:            "ITEM-001",
			date:          time.Now(),
			hours:         8.0,
			rate:          100.0,
			description:   "Development work",
			expectError:   false,
			expectedTotal: 800.0,
		},
		{
			name:          "ValidWorkItemWithDecimals",
			id:            "ITEM-002",
			date:          time.Now(),
			hours:         7.5,
			rate:          125.50,
			description:   "Consulting work",
			expectError:   false,
			expectedTotal: 941.25,
		},
		{
			name:        "EmptyID",
			id:          "",
			date:        time.Now(),
			hours:       8.0,
			rate:        100.0,
			description: "Development work",
			expectError: true,
			errorMsg:    "validation failed for field 'id': is required",
		},
		{
			name:        "ZeroHours",
			id:          "ITEM-003",
			date:        time.Now(),
			hours:       0,
			rate:        100.0,
			description: "Development work",
			expectError: true,
			errorMsg:    "validation failed for field 'hours': must be greater than 0",
		},
		{
			name:        "NegativeHours",
			id:          "ITEM-004",
			date:        time.Now(),
			hours:       -5.0,
			rate:        100.0,
			description: "Development work",
			expectError: true,
			errorMsg:    "validation failed for field 'hours': must be greater than 0",
		},
		{
			name:        "TooManyHours",
			id:          "ITEM-005",
			date:        time.Now(),
			hours:       25.0,
			rate:        100.0,
			description: "Development work",
			expectError: true,
			errorMsg:    "validation failed for field 'hours': cannot exceed 24 hours per entry",
		},
		{
			name:        "ZeroRate",
			id:          "ITEM-006",
			date:        time.Now(),
			hours:       8.0,
			rate:        0,
			description: "Development work",
			expectError: true,
			errorMsg:    "validation failed for field 'rate': must be greater than 0",
		},
		{
			name:        "ExcessiveRate",
			id:          "ITEM-007",
			date:        time.Now(),
			hours:       8.0,
			rate:        10001.0,
			description: "Development work",
			expectError: true,
			errorMsg:    "validation failed for field 'rate': cannot exceed $10,000 per hour",
		},
		{
			name:        "EmptyDescription",
			id:          "ITEM-008",
			date:        time.Now(),
			hours:       8.0,
			rate:        100.0,
			description: "",
			expectError: true,
			errorMsg:    "validation failed for field 'description': is required",
		},
		{
			name:        "FutureDate",
			id:          "ITEM-009",
			date:        time.Now().AddDate(0, 0, 2), // 2 days in future
			hours:       8.0,
			rate:        100.0,
			description: "Development work",
			expectError: true,
			errorMsg:    "validation failed for field 'date': cannot be more than 1 day in the future",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			item, err := NewWorkItem(suite.ctx, tt.id, tt.date, tt.hours, tt.rate, tt.description)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, item)
			} else {
				require.NoError(t, err)
				require.NotNil(t, item)
				assert.Equal(t, tt.id, item.ID)
				assert.Equal(t, tt.date.Truncate(time.Second), item.Date.Truncate(time.Second))
				assert.InDelta(t, tt.hours, item.Hours, 1e-9)
				assert.InDelta(t, tt.rate, item.Rate, 1e-9)
				assert.Equal(t, tt.description, item.Description)
				assert.InDelta(t, tt.expectedTotal, item.Total, 1e-9)
				assert.False(t, item.CreatedAt.IsZero())
			}
		})
	}
}

func (suite *WorkItemTestSuite) TestNewWorkItemWithContext() {
	t := suite.T()

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	item, err := NewWorkItem(ctx, "ITEM-001", time.Now(), 8.0, 100.0, "Development work")
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Nil(t, item)
}

func (suite *WorkItemTestSuite) TestWorkItemValidate() {
	t := suite.T()

	tests := []struct {
		name        string
		workItem    WorkItem
		expectError bool
		errorMsg    string
	}{
		{
			name: "ValidWorkItem",
			workItem: WorkItem{
				ID:          "ITEM-001",
				Date:        time.Now(),
				Hours:       8.0,
				Rate:        100.0,
				Description: "Development work",
				Total:       800.0,
				CreatedAt:   time.Now(),
			},
			expectError: false,
		},
		{
			name: "ZeroDate",
			workItem: WorkItem{
				ID:          "ITEM-002",
				Date:        time.Time{},
				Hours:       8.0,
				Rate:        100.0,
				Description: "Development work",
				Total:       800.0,
				CreatedAt:   time.Now(),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'date': is required",
		},
		{
			name: "LongDescription",
			workItem: WorkItem{
				ID:          "ITEM-003",
				Date:        time.Now(),
				Hours:       8.0,
				Rate:        100.0,
				Description: strings.Repeat("a", 1001),
				Total:       800.0,
				CreatedAt:   time.Now(),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'description': cannot exceed 1000 characters",
		},
		{
			name: "IncorrectTotal",
			workItem: WorkItem{
				ID:          "ITEM-004",
				Date:        time.Now(),
				Hours:       8.0,
				Rate:        100.0,
				Description: "Development work",
				Total:       900.0, // Should be 800.0
				CreatedAt:   time.Now(),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'total': incorrect calculation",
		},
		{
			name: "NegativeTotal",
			workItem: WorkItem{
				ID:          "ITEM-005",
				Date:        time.Now(),
				Hours:       8.0,
				Rate:        100.0,
				Description: "Development work",
				Total:       -800.0,
				CreatedAt:   time.Now(),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'total': must be non-negative",
		},
		{
			name: "MissingCreatedAt",
			workItem: WorkItem{
				ID:          "ITEM-006",
				Date:        time.Now(),
				Hours:       8.0,
				Rate:        100.0,
				Description: "Development work",
				Total:       800.0,
				CreatedAt:   time.Time{},
			},
			expectError: true,
			errorMsg:    "validation failed for field 'created_at': is required",
		},
		{
			name: "WhitespaceOnlyDescription",
			workItem: WorkItem{
				ID:          "ITEM-007",
				Date:        time.Now(),
				Hours:       8.0,
				Rate:        100.0,
				Description: "   ",
				Total:       800.0,
				CreatedAt:   time.Now(),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'description': is required",
		},
		{
			name: "FractionalHoursAndRate",
			workItem: WorkItem{
				ID:          "ITEM-008",
				Date:        time.Now(),
				Hours:       7.5,
				Rate:        125.75,
				Description: "Consulting work",
				Total:       943.13, // 7.5 * 125.75 = 943.125, rounded to 943.13
				CreatedAt:   time.Now(),
			},
			expectError: false,
		},
		{
			name: "EdgeCaseValidDate",
			workItem: WorkItem{
				ID:          "ITEM-009",
				Date:        time.Now().Add(23 * time.Hour), // Just under 24 hours future
				Hours:       8.0,
				Rate:        100.0,
				Description: "Development work",
				Total:       800.0,
				CreatedAt:   time.Now(),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.workItem.Validate(suite.ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func (suite *WorkItemTestSuite) TestUpdateHours() {
	t := suite.T()

	workItem := &WorkItem{
		ID:          "ITEM-001",
		Date:        time.Now(),
		Hours:       8.0,
		Rate:        100.0,
		Description: "Development work",
		Total:       800.0,
		CreatedAt:   time.Now(),
	}

	tests := []struct {
		name          string
		newHours      float64
		expectError   bool
		errorMsg      string
		expectedTotal float64
	}{
		{
			name:          "ValidUpdate",
			newHours:      10.0,
			expectError:   false,
			expectedTotal: 1000.0,
		},
		{
			name:          "ValidDecimalHours",
			newHours:      6.5,
			expectError:   false,
			expectedTotal: 650.0,
		},
		{
			name:        "ZeroHours",
			newHours:    0,
			expectError: true,
			errorMsg:    "hours must be greater than 0",
		},
		{
			name:        "NegativeHours",
			newHours:    -5.0,
			expectError: true,
			errorMsg:    "hours must be greater than 0",
		},
		{
			name:        "ExcessiveHours",
			newHours:    25.0,
			expectError: true,
			errorMsg:    "hours cannot exceed 24 per entry",
		},
		{
			name:          "MaximumValidHours",
			newHours:      24.0,
			expectError:   false,
			expectedTotal: 2400.0,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset work item for each test
			workItem.Hours = 8.0
			workItem.Total = 800.0

			err := workItem.UpdateHours(suite.ctx, tt.newHours)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				// Values should remain unchanged
				assert.InDelta(t, 8.0, workItem.Hours, 1e-9)
				assert.InDelta(t, 800.0, workItem.Total, 1e-9)
			} else {
				require.NoError(t, err)
				assert.InDelta(t, tt.newHours, workItem.Hours, 1e-9)
				assert.InDelta(t, tt.expectedTotal, workItem.Total, 1e-9)
			}
		})
	}
}

func (suite *WorkItemTestSuite) TestUpdateRate() {
	t := suite.T()

	workItem := &WorkItem{
		ID:          "ITEM-001",
		Date:        time.Now(),
		Hours:       8.0,
		Rate:        100.0,
		Description: "Development work",
		Total:       800.0,
		CreatedAt:   time.Now(),
	}

	tests := []struct {
		name          string
		newRate       float64
		expectError   bool
		errorMsg      string
		expectedTotal float64
	}{
		{
			name:          "ValidUpdate",
			newRate:       150.0,
			expectError:   false,
			expectedTotal: 1200.0,
		},
		{
			name:          "ValidDecimalRate",
			newRate:       125.75,
			expectError:   false,
			expectedTotal: 1006.0, // 8 * 125.75 = 1006
		},
		{
			name:        "ZeroRate",
			newRate:     0,
			expectError: true,
			errorMsg:    "rate must be greater than 0",
		},
		{
			name:        "NegativeRate",
			newRate:     -50.0,
			expectError: true,
			errorMsg:    "rate must be greater than 0",
		},
		{
			name:        "ExcessiveRate",
			newRate:     10001.0,
			expectError: true,
			errorMsg:    "rate cannot exceed $10,000 per hour",
		},
		{
			name:          "MaximumValidRate",
			newRate:       10000.0,
			expectError:   false,
			expectedTotal: 80000.0,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset work item for each test
			workItem.Rate = 100.0
			workItem.Total = 800.0

			err := workItem.UpdateRate(suite.ctx, tt.newRate)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				// Values should remain unchanged
				assert.InEpsilon(t, 100.0, workItem.Rate, 1e-9)
				assert.InEpsilon(t, 800.0, workItem.Total, 1e-9)
			} else {
				require.NoError(t, err)
				assert.InEpsilon(t, tt.newRate, workItem.Rate, 1e-9)
				assert.InEpsilon(t, tt.expectedTotal, workItem.Total, 1e-9)
			}
		})
	}
}

func (suite *WorkItemTestSuite) TestUpdateDescription() {
	t := suite.T()

	workItem := &WorkItem{
		ID:          "ITEM-001",
		Date:        time.Now(),
		Hours:       8.0,
		Rate:        100.0,
		Description: "Development work",
		Total:       800.0,
		CreatedAt:   time.Now(),
	}

	tests := []struct {
		name         string
		newDesc      string
		expectError  bool
		errorMsg     string
		expectedDesc string
	}{
		{
			name:         "ValidUpdate",
			newDesc:      "Updated development work",
			expectError:  false,
			expectedDesc: "Updated development work",
		},
		{
			name:         "TrimmedDescription",
			newDesc:      "  Trimmed description  ",
			expectError:  false,
			expectedDesc: "Trimmed description",
		},
		{
			name:        "EmptyDescription",
			newDesc:     "",
			expectError: true,
			errorMsg:    "description cannot be empty",
		},
		{
			name:        "WhitespaceOnlyDescription",
			newDesc:     "   ",
			expectError: true,
			errorMsg:    "description cannot be empty",
		},
		{
			name:        "TooLongDescription",
			newDesc:     strings.Repeat("a", 1001),
			expectError: true,
			errorMsg:    "description cannot exceed 1000 characters",
		},
		{
			name:         "MaxLengthDescription",
			newDesc:      strings.Repeat("a", 1000),
			expectError:  false,
			expectedDesc: strings.Repeat("a", 1000),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset work item for each test
			workItem.Description = "Development work"

			err := workItem.UpdateDescription(suite.ctx, tt.newDesc)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				// Value should remain unchanged
				assert.Equal(t, "Development work", workItem.Description)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedDesc, workItem.Description)
			}
		})
	}
}

func (suite *WorkItemTestSuite) TestContextCancellation() {
	t := suite.T()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	workItem := &WorkItem{
		ID:          "ITEM-001",
		Date:        time.Now(),
		Hours:       8.0,
		Rate:        100.0,
		Description: "Development work",
		Total:       800.0,
		CreatedAt:   time.Now(),
	}

	// Test all update methods with canceled context
	err := workItem.UpdateHours(ctx, 10.0)
	assert.Equal(t, context.Canceled, err)

	err = workItem.UpdateRate(ctx, 150.0)
	assert.Equal(t, context.Canceled, err)

	err = workItem.UpdateDescription(ctx, "Updated description")
	assert.Equal(t, context.Canceled, err)

	// Validate with canceled context
	err = workItem.Validate(ctx)
	assert.Equal(t, context.Canceled, err)
}

func (suite *WorkItemTestSuite) TestFormattedMethods() {
	t := suite.T()

	tests := []struct {
		name          string
		workItem      WorkItem
		expectedTotal string
		expectedRate  string
		expectedHours string
	}{
		{
			name: "WholeNumbers",
			workItem: WorkItem{
				Hours: 8.0,
				Rate:  100.0,
				Total: 800.0,
			},
			expectedTotal: "$800.00",
			expectedRate:  "$100.00",
			expectedHours: "8.00",
		},
		{
			name: "DecimalNumbers",
			workItem: WorkItem{
				Hours: 7.5,
				Rate:  125.75,
				Total: 943.13,
			},
			expectedTotal: "$943.13",
			expectedRate:  "$125.75",
			expectedHours: "7.50",
		},
		{
			name: "ZeroValues",
			workItem: WorkItem{
				Hours: 0.0,
				Rate:  0.0,
				Total: 0.0,
			},
			expectedTotal: "$0.00",
			expectedRate:  "$0.00",
			expectedHours: "0.00",
		},
		{
			name: "LargeValues",
			workItem: WorkItem{
				Hours: 24.0,
				Rate:  10000.0,
				Total: 240000.0,
			},
			expectedTotal: "$240000.00",
			expectedRate:  "$10000.00",
			expectedHours: "24.00",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			assert.Equal(t, tt.expectedTotal, tt.workItem.GetFormattedTotal())
			assert.Equal(t, tt.expectedRate, tt.workItem.GetFormattedRate())
			assert.Equal(t, tt.expectedHours, tt.workItem.GetFormattedHours())
		})
	}
}
