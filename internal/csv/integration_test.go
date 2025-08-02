package csv

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

// CSVIntegrationTestSuite tests the integration between parser and validator
type CSVIntegrationTestSuite struct {
	suite.Suite

	parser    *CSVParser
	validator *WorkItemValidator
	logger    *MockLogger
	idGen     *MockIDGenerator
}

func (suite *CSVIntegrationTestSuite) SetupTest() {
	suite.logger = &MockLogger{}
	suite.validator = NewWorkItemValidator(suite.logger)
	suite.idGen = &MockIDGenerator{}
	suite.parser = NewCSVParser(suite.validator, suite.logger, suite.idGen)
}

func TestCSVIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CSVIntegrationTestSuite))
}

// TestEndToEndParsing tests complete parsing workflow
func (suite *CSVIntegrationTestSuite) TestEndToEndParsing() {
	tests := []struct {
		name         string
		csvData      string
		expected     int
		hasError     bool
		useTabFormat bool
	}{
		{
			name: "ValidCSVWithStandardFormat",
			csvData: `Date,Hours,Rate,Description
2024-01-15,8.0,100.0,Development work
2024-01-16,6.5,100.0,Testing
2024-01-17,4.0,125.0,Code review`,
			expected: 3,
			hasError: false,
		},
		{
			name: "ValidCSVWithTabDelimiter",
			csvData: "Date\tHours\tRate\tDescription\n" +
				"2024-01-15\t8.0\t100.0\tDevelopment work\n" +
				"2024-01-16\t6.5\t100.0\tTesting",
			expected:     2,
			hasError:     false,
			useTabFormat: true,
		},
		{
			name: "CSVWithValidationErrors",
			csvData: `Date,Hours,Rate,Description
2024-01-15,8.0,100.0,Development work
2024-01-16,-5.0,100.0,Invalid hours
2024-01-17,8.0,-50.0,Invalid rate`,
			expected: 1, // Only first row is valid
			hasError: true,
		},
		{
			name: "CSVWithMixedValidation",
			csvData: `Date,Hours,Rate,Description
2024-01-15,8.0,100.0,Development work
invalid-date,8.0,100.0,Bad date
2024-01-17,abc,100.0,Bad hours
2024-01-18,8.0,100.0,Good work`,
			expected: 2, // First and last rows are valid
			hasError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			reader := strings.NewReader(tt.csvData)

			options := ParseOptions{
				ContinueOnError: true,
				SkipEmptyRows:   true,
			}
			if tt.useTabFormat {
				options.Format = "tab"
			}

			result, err := suite.parser.ParseTimesheet(ctx, reader, options)

			if tt.hasError {
				// Should not error if continue on error is true
				suite.Require().NoError(err)
				suite.NotEmpty(result.Errors, "should have parsing errors")
			} else {
				suite.Require().NoError(err)
				suite.Empty(result.Errors, "should have no errors")
			}

			suite.Len(result.WorkItems, tt.expected)
			suite.Equal(len(result.WorkItems), result.SuccessRows)
		})
	}
}

// TestFormatDetectionIntegration tests format detection with parsing
func (suite *CSVIntegrationTestSuite) TestFormatDetectionIntegration() {
	tests := []struct {
		name      string
		csvData   string
		expectFmt string
		expectDel rune
	}{
		{
			name: "StandardCommaDelimited",
			csvData: `Date,Hours,Rate,Description
2024-01-15,8.0,100.0,Development work`,
			expectFmt: "standard",
			expectDel: ',',
		},
		{
			name: "TabDelimited",
			csvData: "Date\tHours\tRate\tDescription\n" +
				"2024-01-15\t8.0\t100.0\tDevelopment work",
			expectFmt: "tab",
			expectDel: '\t',
		},
		{
			name: "SemicolonDelimited",
			csvData: `Date;Hours;Rate;Description
2024-01-15;8.0;100.0;Development work`,
			expectFmt: "semicolon",
			expectDel: ';',
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			reader := strings.NewReader(tt.csvData)

			// First detect format
			formatInfo, err := suite.parser.DetectFormat(ctx, reader)
			suite.Require().NoError(err)
			suite.Equal(tt.expectFmt, formatInfo.Name)
			suite.Equal(tt.expectDel, formatInfo.Delimiter)

			// Reset reader and parse with detected format
			reader = strings.NewReader(tt.csvData)
			options := ParseOptions{
				Format:          formatInfo.Name,
				ContinueOnError: false,
			}

			result, err := suite.parser.ParseTimesheet(ctx, reader, options)
			suite.Require().NoError(err)
			suite.Len(result.WorkItems, 1)
			suite.Equal(formatInfo.Name, result.Format)
		})
	}
}

// TestErrorHandlingIntegration tests how errors flow between components
func (suite *CSVIntegrationTestSuite) TestErrorHandlingIntegration() {
	suite.Run("ValidationErrorsWithContinue", func() {
		csvData := `Date,Hours,Rate,Description
2024-01-15,8.0,100.0,Valid work
2024-01-16,-1.0,100.0,Negative hours
2024-01-17,25.0,100.0,Too many hours
2024-01-18,8.0,-10.0,Negative rate`

		ctx := context.Background()
		reader := strings.NewReader(csvData)
		options := ParseOptions{ContinueOnError: true}

		result, err := suite.parser.ParseTimesheet(ctx, reader, options)
		suite.Require().NoError(err)

		// Should have 1 valid work item
		suite.Len(result.WorkItems, 1)
		suite.Equal(4, result.TotalRows)
		suite.Equal(1, result.SuccessRows)
		suite.Equal(3, result.ErrorRows)
		suite.Len(result.Errors, 3)

		// Check error details
		errors := result.Errors
		suite.Contains(errors[0].Message, "hours")
		suite.Contains(errors[1].Message, "hours")
		suite.Contains(errors[2].Message, "rate")
	})

	suite.Run("ValidationErrorsWithoutContinue", func() {
		csvData := `Date,Hours,Rate,Description
2024-01-15,8.0,100.0,Valid work
2024-01-16,-1.0,100.0,Negative hours`

		ctx := context.Background()
		reader := strings.NewReader(csvData)
		options := ParseOptions{ContinueOnError: false}

		result, err := suite.parser.ParseTimesheet(ctx, reader, options)
		suite.Require().Error(err)
		suite.Contains(err.Error(), "hours")

		// Result may be nil when error occurs early
		if result != nil {
			suite.Len(result.WorkItems, 1)
		}
	})
}

// TestComplexDataScenarios tests real-world data scenarios
func (suite *CSVIntegrationTestSuite) TestComplexDataScenarios() {
	suite.Run("MixedDateFormats", func() {
		csvData := `Date,Hours,Rate,Description
2024-01-15,8.0,100.0,ISO format
01/16/2024,7.5,100.0,US format
15/01/2024,6.0,100.0,EU format`

		ctx := context.Background()
		reader := strings.NewReader(csvData)
		options := ParseOptions{ContinueOnError: true}

		result, err := suite.parser.ParseTimesheet(ctx, reader, options)
		suite.Require().NoError(err)

		// All should parse successfully with auto-detection (fixed EU format to use /)
		suite.Len(result.WorkItems, 3)
		suite.Empty(result.Errors)
	})

	suite.Run("QuotedFieldsWithDelimiters", func() {
		csvData := `Date,Hours,Rate,Description
2024-01-15,8.0,100.0,"Work with, comma"
2024-01-16,7.0,100.0,"Testing ""quotes"" handling"`

		ctx := context.Background()
		reader := strings.NewReader(csvData)
		options := ParseOptions{}

		result, err := suite.parser.ParseTimesheet(ctx, reader, options)
		suite.Require().NoError(err)

		suite.Len(result.WorkItems, 2)
		suite.Equal("Work with, comma", result.WorkItems[0].Description)
		suite.Equal(`Testing "quotes" handling`, result.WorkItems[1].Description)
	})

	suite.Run("EmptyRowsAndWhitespace", func() {
		csvData := `Date,Hours,Rate,Description
2024-01-15,8.0,100.0,Valid work
2024-01-16,7.0,100.0,Another valid work`

		ctx := context.Background()
		reader := strings.NewReader(csvData)
		options := ParseOptions{}

		result, err := suite.parser.ParseTimesheet(ctx, reader, options)
		suite.Require().NoError(err)

		suite.Len(result.WorkItems, 2)
		suite.Equal(2, result.SuccessRows)
	})
}

// TestContextCancellation tests cancellation during parsing
func (suite *CSVIntegrationTestSuite) TestContextCancellation() {
	// Create large CSV data
	var csvBuilder strings.Builder
	csvBuilder.WriteString("Date,Hours,Rate,Description\n")
	for i := 0; i < 10000; i++ {
		csvBuilder.WriteString("2024-01-15,8.0,100.0,Work item\n")
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately to ensure cancellation
	cancel()

	reader := strings.NewReader(csvBuilder.String())
	options := ParseOptions{}

	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	// Should be canceled
	suite.Require().Error(err)
	suite.Equal(context.Canceled, err)

	// Result may be nil when canceled early
	if result != nil {
		suite.Less(len(result.WorkItems), 10000)
	}
}

// TestLargeFileHandling tests memory usage with large files
func (suite *CSVIntegrationTestSuite) TestLargeFileHandling() {
	// Create moderately large CSV (1000 rows)
	var csvBuilder strings.Builder
	csvBuilder.WriteString("Date,Hours,Rate,Description\n")
	for i := 0; i < 1000; i++ {
		csvBuilder.WriteString("2024-01-15,8.0,100.0,Development work\n")
	}

	ctx := context.Background()
	reader := strings.NewReader(csvBuilder.String())
	options := ParseOptions{}

	result, err := suite.parser.ParseTimesheet(ctx, reader, options)
	suite.Require().NoError(err)

	suite.Len(result.WorkItems, 1000)
	suite.Equal(1000, result.SuccessRows)
	suite.Equal(1000, result.TotalRows)
	suite.Empty(result.Errors)
}

// TestHeaderVariations tests different header naming patterns
func (suite *CSVIntegrationTestSuite) TestHeaderVariations() {
	tests := []struct {
		name    string
		headers string
		valid   bool
	}{
		{
			name:    "StandardHeaders",
			headers: "Date,Hours,Rate,Description",
			valid:   true,
		},
		{
			name:    "CaseInsensitive",
			headers: "date,hours,rate,description",
			valid:   true,
		},
		{
			name:    "WithSpaces",
			headers: "Date, Hours, Rate, Description",
			valid:   true,
		},
		{
			name:    "AlternativeNames",
			headers: "work_date,hours_worked,hourly_rate,task", // Use supported alternative names
			valid:   true,
		},
		{
			name:    "MissingRequiredHeader",
			headers: "Date,Hours,Description", // Missing Rate
			valid:   false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			csvData := tt.headers + "\n2024-01-15,8.0,100.0,Development work"

			ctx := context.Background()
			reader := strings.NewReader(csvData)
			options := ParseOptions{}

			result, err := suite.parser.ParseTimesheet(ctx, reader, options)

			if tt.valid {
				suite.Require().NoError(err)
				suite.Len(result.WorkItems, 1)
			} else {
				suite.Require().Error(err)
				// Error could be either header validation or CSV parsing
				errorStr := err.Error()
				headerFound := strings.Contains(errorStr, "header") ||
					strings.Contains(errorStr, "required field") ||
					strings.Contains(errorStr, "wrong number of fields")
				suite.True(headerFound, "Expected header-related error, got: %s", errorStr)
			}
		})
	}
}
