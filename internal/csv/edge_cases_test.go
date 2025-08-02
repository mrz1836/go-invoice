package csv

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/mrz/go-invoice/internal/models"
)

// CSVEdgeCasesTestSuite tests edge cases and boundary conditions
type CSVEdgeCasesTestSuite struct {
	suite.Suite
	parser    *CSVParser
	validator *WorkItemValidator
	logger    *MockLogger
	idGen     *MockIDGenerator
}

func (suite *CSVEdgeCasesTestSuite) SetupTest() {
	suite.logger = &MockLogger{}
	suite.validator = NewWorkItemValidator(suite.logger)
	suite.idGen = &MockIDGenerator{}
	suite.parser = NewCSVParser(suite.validator, suite.logger, suite.idGen)
}

func TestCSVEdgeCasesTestSuite(t *testing.T) {
	suite.Run(t, new(CSVEdgeCasesTestSuite))
}

// TestMalformedCSVData tests handling of malformed CSV data
func (suite *CSVEdgeCasesTestSuite) TestMalformedCSVData() {
	tests := []struct {
		name        string
		csvData     string
		expectError bool
		expectRows  int
	}{
		{
			name: "UnquotedCommasInFields",
			csvData: `Date,Hours,Rate,Description
2024-01-15,8.0,100.0,Work with, unquoted comma`,
			expectError: true, // CSV parser correctly rejects unquoted commas
			expectRows:  0,
		},
		{
			name: "MismatchedQuotes",
			csvData: `Date,Hours,Rate,Description
2024-01-15,8.0,100.0,"Unmatched quote`,
			expectError: true,
			expectRows:  0,
		},
		{
			name: "InconsistentFieldCount",
			csvData: `Date,Hours,Rate,Description
2024-01-15,8.0,100.0,Normal work
2024-01-16,7.0,Extra field,100.0,Too many fields
2024-01-17,6.0,Missing field`,
			expectError: true, // Should error with continue off
			expectRows:  0,    // No rows processed due to error
		},
		{
			name:        "OnlyHeaders",
			csvData:     `Date,Hours,Rate,Description`,
			expectError: false,
			expectRows:  0,
		},
		{
			name:        "EmptyFile",
			csvData:     ``,
			expectError: true,
			expectRows:  0,
		},
		{
			name: "WhitespaceOnly",
			csvData: `   
   
   `,
			expectError: true,
			expectRows:  0,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			reader := strings.NewReader(tt.csvData)
			options := ParseOptions{ContinueOnError: true}

			result, err := suite.parser.ParseTimesheet(ctx, reader, options)

			if tt.expectError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}

			if result != nil {
				suite.Len(result.WorkItems, tt.expectRows)
			} else {
				suite.Equal(0, tt.expectRows)
			}
		})
	}
}

// TestDateFormatEdgeCases tests various date format edge cases
func (suite *CSVEdgeCasesTestSuite) TestDateFormatEdgeCases() {
	tests := []struct {
		name      string
		dateValue string
		valid     bool
	}{
		// Valid formats (based on parser.go parseDate implementation)
		{"ISO8601", "2024-01-15", true},
		{"USFormat", "01/15/2024", true},
		{"EUFormat", "15/01/2024", true}, // Parser uses "/" not "."
		{"AlternativeISO", "2024/01/15", true},
		{"MonthName", "Jan 15, 2024", true},
		{"FullMonthName", "January 15, 2024", true},

		// Edge cases
		{"LeapYear", "2024-02-29", true},
		{"NonLeapYear", "2023-02-29", false},
		{"InvalidMonth", "2024-13-01", false},
		{"InvalidDay", "2024-01-32", false},
		{"ZeroMonth", "2024-00-15", false},
		{"ZeroDay", "2024-01-00", false},

		// Boundary dates (based on validator rules: 2 years past, 1 week future)
		{"FarPast", "1900-01-01", false},   // More than 2 years ago
		{"FarFuture", "2100-01-01", false}, // More than 1 week from now
		{"RecentPast", "2024-01-01", true}, // Within 2 years
		{"Tomorrow", "2025-08-03", true},   // Within 1 week of now

		// Malformed dates
		{"Empty", "", false},
		{"JustYear", "2024", false},
		{"JustMonth", "01", false},
		{"UnsupportedDotFormat", "15.01.2024", false}, // Parser doesn't support dots
		{"ShortYear", "24-01-15", false},              // Not in supported formats
		{"Mixed", "24/Jan/2024", false},
		{"SimpleTime", "2024-01-15 10:30", false},
		{"ISOWithTime", "2024-01-15 15:04:05", true}, // This format IS supported
		{"Timezone", "2024-01-15T10:30Z", false},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Properly escape dates that contain commas
			dateValue := tt.dateValue
			if strings.Contains(dateValue, ",") {
				dateValue = `"` + dateValue + `"`
			}
			csvData := fmt.Sprintf("Date,Hours,Rate,Description\n%s,8.0,100.0,Test work", dateValue)

			ctx := context.Background()
			reader := strings.NewReader(csvData)
			options := ParseOptions{ContinueOnError: true}

			result, err := suite.parser.ParseTimesheet(ctx, reader, options)

			if tt.valid {
				suite.Require().NoError(err)
				if result != nil {
					suite.Len(result.WorkItems, 1)
					suite.Empty(result.Errors)
				}
			} else {
				// May not error if continue on error is true
				if err == nil && result != nil {
					suite.NotEmpty(result.Errors, "should have validation errors")
					suite.Empty(result.WorkItems)
				}
			}
		})
	}
}

// TestNumericFieldEdgeCases tests boundary conditions for numeric fields
func (suite *CSVEdgeCasesTestSuite) TestNumericFieldEdgeCases() {
	tests := []struct {
		name        string
		hours       string
		rate        string
		expectValid bool
		expectHours float64
		expectRate  float64
	}{
		// Valid cases
		{"NormalValues", "8.0", "100.0", true, 8.0, 100.0},
		{"IntegerValues", "8", "100", true, 8.0, 100.0},
		{"DecimalPrecision", "8.25", "125.50", true, 8.25, 125.50},

		// Boundary cases (based on validator: 0 < hours <= 24, 1 <= rate <= 1000)
		{"MaxHours", "24.0", "100.0", true, 24.0, 100.0},
		{"MinHours", "0.01", "100.0", true, 0.01, 100.0},
		{"MinRate", "8.0", "1.0", true, 8.0, 1.0},
		{"MaxRate", "8.0", "1000.0", true, 8.0, 1000.0},

		// Invalid cases (based on validator rules: hours > 0 && <= 24, rate > 0 && <= 1000)
		{"NegativeHours", "-1.0", "100.0", false, 0, 0},
		{"NegativeRate", "8.0", "-50.0", false, 0, 0},
		{"ExcessiveHours", "25.0", "100.0", false, 0, 0}, // > 24 hours
		{"ExcessiveRate", "8.0", "1001.0", false, 0, 0},  // > 1000 rate
		{"ZeroHours", "0.0", "100.0", false, 0, 0},       // hours must be > 0
		{"ZeroRate", "8.0", "0.0", false, 0, 0},          // rate must be > 0
		{"VeryLowRate", "8.0", "0.5", false, 0, 0},       // rate < 1.0

		// Precision edge cases (validator limits hours to 2 decimal places)
		{"HighPrecisionHours", "8.123456", "100.0", false, 0, 0},          // > 2 decimal places
		{"HighPrecisionRate", "8.0", "100.123456", true, 8.0, 100.123456}, // Rate precision OK
		{"VerySmallHours", "0.01", "100.0", true, 0.01, 100.0},
		{"VerySmallRate", "8.0", "0.01", false, 0, 0}, // Rate < 1.0

		// String values
		{"AlphabeticHours", "eight", "100.0", false, 0, 0},
		{"AlphabeticRate", "8.0", "hundred", false, 0, 0},
		{"EmptyHours", "", "100.0", false, 0, 0},
		{"EmptyRate", "8.0", "", false, 0, 0},

		// Special numeric values
		{"InfinityHours", "inf", "100.0", false, 0, 0},
		{"InfinityRate", "8.0", "inf", false, 0, 0},
		{"NaNHours", "NaN", "100.0", false, 0, 0},
		{"NaNRate", "8.0", "NaN", false, 0, 0},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			csvData := fmt.Sprintf("Date,Hours,Rate,Description\n2024-01-15,%s,%s,Test work", tt.hours, tt.rate)

			ctx := context.Background()
			reader := strings.NewReader(csvData)
			options := ParseOptions{ContinueOnError: true}

			result, err := suite.parser.ParseTimesheet(ctx, reader, options)

			if tt.expectValid {
				suite.Require().NoError(err)
				if result != nil {
					suite.Len(result.WorkItems, 1)
					suite.Empty(result.Errors)

					if len(result.WorkItems) > 0 {
						workItem := result.WorkItems[0]
						suite.InEpsilon(tt.expectHours, workItem.Hours, 1e-9)
						suite.InEpsilon(tt.expectRate, workItem.Rate, 1e-9)
					}
				}
			} else {
				// Should have errors or no result
				if err == nil {
					if result != nil {
						// Either parsing errors or no work items
						hasErrors := len(result.Errors) > 0
						noWorkItems := len(result.WorkItems) == 0
						suite.True(hasErrors || noWorkItems,
							"Expected either errors (%d) or no work items (%d)",
							len(result.Errors), len(result.WorkItems))
					}
					// If result is nil, that's also acceptable for invalid cases
				}
			}
		})
	}
}

// TestFloatingPointPrecision tests floating point calculation edge cases
func (suite *CSVEdgeCasesTestSuite) TestFloatingPointPrecision() {
	tests := []struct {
		name  string
		hours float64
		rate  float64
		total float64
		valid bool
	}{
		{"ExactMatch", 8.0, 100.0, 800.0, true},
		{"SmallDifference", 8.33, 120.0, 999.60, true},
		{"RoundingError", 8.33, 120.0, 999.61, true},       // Small differences are accepted
		{"PrecisionLimit", 1.0 / 3.0, 300.0, 100.0, false}, // Too many decimal places
		{"LargeDifference", 8.0, 100.0, 900.0, true},       // Total mismatches are accepted
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			csvData := fmt.Sprintf("Date,Hours,Rate,Description,Total\n2024-01-15,%.6f,%.2f,Test work,%.2f",
				tt.hours, tt.rate, tt.total)

			ctx := context.Background()
			reader := strings.NewReader(csvData)
			options := ParseOptions{ContinueOnError: true}

			result, err := suite.parser.ParseTimesheet(ctx, reader, options)

			if tt.valid {
				suite.Require().NoError(err)
				suite.Len(result.WorkItems, 1)
				suite.Empty(result.Errors)
			} else {
				if err == nil {
					suite.NotEmpty(result.Errors)
				}
			}
		})
	}
}

// TestDescriptionEdgeCases tests description field edge cases
func (suite *CSVEdgeCasesTestSuite) TestDescriptionEdgeCases() {
	tests := []struct {
		name        string
		description string
		valid       bool
	}{
		{"NormalDescription", "Development work", true},
		{"EmptyDescription", "", false},
		{"WhitespaceOnly", "   ", false},
		{"SingleChar", "X", false},                                // < 3 chars
		{"ThreeCharMin", "ABC", true},                             // Min valid length
		{"VeryLongDescription", strings.Repeat("A", 1000), false}, // > 500 chars
		{"MaxLengthDescription", strings.Repeat("A", 500), true},  // Max valid length
		{"SpecialCharacters", "Work w/ special chars: @#$%^&*()", true},
		{"UnicodeCharacters", "Работа с unicode символами", true},
		{"Newlines", "Work with\nnewlines", true},
		{"Tabs", "Work with\ttabs", true},
		{"Quotes", `Work with "quotes"`, true},
		{"GenericDescription", "work", false}, // Too generic
		{"AnotherGeneric", "task", false},     // Too generic
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Need to properly escape the description for CSV
			escapedDesc := strings.ReplaceAll(tt.description, `"`, `""`)
			if strings.ContainsAny(tt.description, ",\n\r\"") {
				escapedDesc = `"` + escapedDesc + `"`
			}

			csvData := fmt.Sprintf("Date,Hours,Rate,Description\n2024-01-15,8.0,100.0,%s", escapedDesc)

			ctx := context.Background()
			reader := strings.NewReader(csvData)
			options := ParseOptions{ContinueOnError: true}

			result, err := suite.parser.ParseTimesheet(ctx, reader, options)

			if tt.valid {
				suite.Require().NoError(err)
				suite.Len(result.WorkItems, 1)
				suite.Empty(result.Errors)
			} else {
				if err == nil {
					suite.NotEmpty(result.Errors)
				}
			}
		})
	}
}

// TestFormatDetectionEdgeCases tests edge cases in format detection
func (suite *CSVEdgeCasesTestSuite) TestFormatDetectionEdgeCases() {
	tests := []struct {
		name        string
		csvData     string
		expectError bool
		expectFmt   string
	}{
		{
			name: "MixedDelimiters",
			csvData: `Date,Hours;Rate,Description
2024-01-15,8.0;100.0,Work`,
			expectError: true, // Ambiguous format
		},
		{
			name: "NoDelimiters",
			csvData: `DateHoursRateDescription
20240115810000Work`,
			expectError: true, // Can't detect format
		},
		{
			name: "OnlyCommasInQuotes",
			csvData: `"Date,Time","Hours","Rate","Description"
"2024,01,15","8.0","100.0","Work"`,
			expectError: false,
			expectFmt:   "standard",
		},
		{
			name: "VeryFewSamples",
			csvData: `A,B
1,2`,
			expectError: true, // Not enough columns to be work items
		},
		{
			name:        "TooManyColumns",
			csvData:     strings.Repeat("Col,", 100) + "LastCol\n" + strings.Repeat("Val,", 100) + "LastVal",
			expectError: true, // Too many columns for work items
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			reader := strings.NewReader(tt.csvData)

			formatInfo, err := suite.parser.DetectFormat(ctx, reader)

			if tt.expectError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Equal(tt.expectFmt, formatInfo.Name)
			}
		})
	}
}

// TestValidatorEdgeCases tests validator-specific edge cases
func (suite *CSVEdgeCasesTestSuite) TestValidatorEdgeCases() {
	ctx := context.Background()

	// Static error variable for test validator
	errNoWeekendsWork := fmt.Errorf("no work allowed on weekends")

	suite.Run("CustomRuleAddition", func() {
		// Add a custom rule that rejects work on weekends
		customRule := ValidationRule{
			Name:        "no_weekends",
			Description: "No work allowed on weekends",
			Validator: func(ctx context.Context, item *models.WorkItem) error {
				if item.Date.Weekday() == time.Saturday || item.Date.Weekday() == time.Sunday {
					return errNoWeekendsWork
				}
				return nil
			},
		}
		suite.validator.AddCustomRule(customRule)

		// Test weekend work (should fail)
		weekendWork, err := models.NewWorkItem(ctx, "test", time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC), 8.0, 100.0, "Weekend work") // Saturday
		suite.Require().NoError(err)

		err = suite.validator.ValidateWorkItem(ctx, weekendWork)
		suite.Require().Error(err)
		suite.Contains(err.Error(), "weekends")

		// Test weekday work (should pass)
		weekdayWork, err := models.NewWorkItem(ctx, "test", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), 8.0, 100.0, "Weekday work") // Monday
		suite.Require().NoError(err)

		err = suite.validator.ValidateWorkItem(ctx, weekdayWork)
		suite.Require().NoError(err)

		// Remove the rule
		suite.validator.RemoveRule("no_weekends")

		// Weekend work should now pass
		err = suite.validator.ValidateWorkItem(ctx, weekendWork)
		suite.Require().NoError(err)
	})

	suite.Run("BatchValidationConsistency", func() {
		// Create work items with inconsistent rates
		workItems := []models.WorkItem{}

		item1, _ := models.NewWorkItem(ctx, "test1", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), 8.0, 100.0, "Work 1")
		item2, _ := models.NewWorkItem(ctx, "test2", time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC), 8.0, 200.0, "Work 2")
		item3, _ := models.NewWorkItem(ctx, "test3", time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC), 8.0, 150.0, "Work 3")
		item4, _ := models.NewWorkItem(ctx, "test4", time.Date(2024, 1, 18, 0, 0, 0, 0, time.UTC), 8.0, 250.0, "Work 4")

		// Need more than 3 different rates to trigger warning
		workItems = append(workItems, *item1, *item2, *item3, *item4)

		// ValidateBatch returns an error, not warnings - but logs warnings
		err := suite.validator.ValidateBatch(ctx, workItems)
		suite.NoError(err, "batch validation should not error, just log warnings")

		// Check that rate inconsistency was logged (similar to existing tests)
		var foundWarning bool
		for _, msg := range suite.logger.messages {
			if msg.Level == "DEBUG" && strings.Contains(msg.Msg, "multiple different rates detected") {
				foundWarning = true
				break
			}
		}
		suite.True(foundWarning, "should have logged rate inconsistency warning")
	})
}

// TestMemoryAndPerformance tests memory usage and performance edge cases
func (suite *CSVEdgeCasesTestSuite) TestMemoryAndPerformance() {
	suite.Run("VeryLargeFields", func() {
		// Create work item with very large description
		largeDesc := strings.Repeat("Very long description ", 1000) // ~22KB description

		csvData := fmt.Sprintf("Date,Hours,Rate,Description\n2024-01-15,8.0,100.0,\"%s\"", largeDesc)

		ctx := context.Background()
		reader := strings.NewReader(csvData)
		options := ParseOptions{}

		result, err := suite.parser.ParseTimesheet(ctx, reader, options)
		// Expect error due to description length validation
		if err != nil {
			suite.Contains(err.Error(), "validation failed")
			return
		}

		// If no error, validate the result
		suite.Len(result.WorkItems, 1)
		if len(result.WorkItems) > 0 {
			suite.Equal(largeDesc, result.WorkItems[0].Description)
		}
	})

	suite.Run("ManySmallRows", func() {
		// Create many small rows to test memory allocation patterns
		var csvBuilder strings.Builder
		csvBuilder.WriteString("Date,Hours,Rate,Description\n")

		for i := 0; i < 100; i++ {
			csvBuilder.WriteString(fmt.Sprintf("2024-01-15,1.0,100.0,Work %d\n", i))
		}

		ctx := context.Background()
		reader := strings.NewReader(csvBuilder.String())
		options := ParseOptions{}

		result, err := suite.parser.ParseTimesheet(ctx, reader, options)
		suite.Require().NoError(err)
		suite.Len(result.WorkItems, 100)
		suite.Equal(100, result.SuccessRows)
	})
}

// TestErrorMessageQuality tests the quality and usefulness of error messages
func (suite *CSVEdgeCasesTestSuite) TestErrorMessageQuality() {
	tests := []struct {
		name               string
		csvData            string
		expectedInMsg      []string
		expectedSuggestion bool
	}{
		{
			name:               "InvalidDate",
			csvData:            "Date,Hours,Rate,Description\n2024-13-01,8.0,100.0,Work",
			expectedInMsg:      []string{"date", "2024-13-01"},
			expectedSuggestion: false,
		},
		{
			name:               "NegativeHours",
			csvData:            "Date,Hours,Rate,Description\n2024-01-15,-5.0,100.0,Work",
			expectedInMsg:      []string{"hours", "-5"},
			expectedSuggestion: false,
		},
		{
			name:               "InvalidRate",
			csvData:            "Date,Hours,Rate,Description\n2024-01-15,8.0,abc,Work",
			expectedInMsg:      []string{"rate", "abc"},
			expectedSuggestion: false,
		},
		{
			name:               "MissingHeader",
			csvData:            "Date,Hours,Description\n2024-01-15,8.0,Work", // Missing Rate
			expectedInMsg:      []string{"rate", "not found"},
			expectedSuggestion: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			reader := strings.NewReader(tt.csvData)
			options := ParseOptions{ContinueOnError: true}

			result, err := suite.parser.ParseTimesheet(ctx, reader, options)

			// Should have either an error or parse errors
			var errorMsg string
			if err != nil {
				errorMsg = err.Error()
			} else if result != nil && len(result.Errors) > 0 {
				errorMsg = result.Errors[0].Message
			}

			suite.NotEmpty(errorMsg, "should have an error message")

			// Check that expected strings are in the error message
			for _, expected := range tt.expectedInMsg {
				suite.Contains(strings.ToLower(errorMsg), strings.ToLower(expected))
			}

			// Check for suggestions if expected
			if tt.expectedSuggestion {
				suite.Require().NotNil(result, "result should not be nil for suggestion check")
				suite.Require().NotEmpty(result.Errors, "should have errors for suggestion check")
				suite.NotEmpty(result.Errors[0].Suggestion, "should have a suggestion")
			}
		})
	}
}
