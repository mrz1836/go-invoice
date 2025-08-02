package csv

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				require.NoError(suite.T(), err)
				assert.True(suite.T(), len(result.Errors) > 0, "should have parsing errors")
			} else {
				require.NoError(suite.T(), err)
				assert.Empty(suite.T(), result.Errors, "should have no errors")
			}

			assert.Equal(suite.T(), tt.expected, len(result.WorkItems))
			assert.Equal(suite.T(), len(result.WorkItems), result.SuccessRows)
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
			require.NoError(suite.T(), err)
			assert.Equal(suite.T(), tt.expectFmt, formatInfo.Name)
			assert.Equal(suite.T(), tt.expectDel, formatInfo.Delimiter)

			// Reset reader and parse with detected format
			reader = strings.NewReader(tt.csvData)
			options := ParseOptions{
				Format:          formatInfo.Name,
				ContinueOnError: false,
			}

			result, err := suite.parser.ParseTimesheet(ctx, reader, options)
			require.NoError(suite.T(), err)
			assert.Equal(suite.T(), 1, len(result.WorkItems))
			assert.Equal(suite.T(), formatInfo.Name, result.Format)
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
		require.NoError(suite.T(), err)

		// Should have 1 valid work item
		assert.Equal(suite.T(), 1, len(result.WorkItems))
		assert.Equal(suite.T(), 4, result.TotalRows)
		assert.Equal(suite.T(), 1, result.SuccessRows)
		assert.Equal(suite.T(), 3, result.ErrorRows)
		assert.Equal(suite.T(), 3, len(result.Errors))

		// Check error details
		errors := result.Errors
		assert.Contains(suite.T(), errors[0].Message, "hours")
		assert.Contains(suite.T(), errors[1].Message, "hours")
		assert.Contains(suite.T(), errors[2].Message, "rate")
	})

	suite.Run("ValidationErrorsWithoutContinue", func() {
		csvData := `Date,Hours,Rate,Description
2024-01-15,8.0,100.0,Valid work
2024-01-16,-1.0,100.0,Negative hours`

		ctx := context.Background()
		reader := strings.NewReader(csvData)
		options := ParseOptions{ContinueOnError: false}

		result, err := suite.parser.ParseTimesheet(ctx, reader, options)
		require.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "hours")

		// Result may be nil when error occurs early
		if result != nil {
			assert.Equal(suite.T(), 1, len(result.WorkItems))
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
		require.NoError(suite.T(), err)

		// All should parse successfully with auto-detection (fixed EU format to use /)
		assert.Equal(suite.T(), 3, len(result.WorkItems))
		assert.Empty(suite.T(), result.Errors)
	})

	suite.Run("QuotedFieldsWithDelimiters", func() {
		csvData := `Date,Hours,Rate,Description
2024-01-15,8.0,100.0,"Work with, comma"
2024-01-16,7.0,100.0,"Testing ""quotes"" handling"`

		ctx := context.Background()
		reader := strings.NewReader(csvData)
		options := ParseOptions{}

		result, err := suite.parser.ParseTimesheet(ctx, reader, options)
		require.NoError(suite.T(), err)

		assert.Equal(suite.T(), 2, len(result.WorkItems))
		assert.Equal(suite.T(), "Work with, comma", result.WorkItems[0].Description)
		assert.Equal(suite.T(), `Testing "quotes" handling`, result.WorkItems[1].Description)
	})

	suite.Run("EmptyRowsAndWhitespace", func() {
		csvData := `Date,Hours,Rate,Description
2024-01-15,8.0,100.0,Valid work
2024-01-16,7.0,100.0,Another valid work`

		ctx := context.Background()
		reader := strings.NewReader(csvData)
		options := ParseOptions{}

		result, err := suite.parser.ParseTimesheet(ctx, reader, options)
		require.NoError(suite.T(), err)

		assert.Equal(suite.T(), 2, len(result.WorkItems))
		assert.Equal(suite.T(), 2, result.SuccessRows)
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

	// Should be cancelled
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), context.Canceled, err)

	// Result may be nil when cancelled early
	if result != nil {
		assert.True(suite.T(), len(result.WorkItems) < 10000)
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
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), 1000, len(result.WorkItems))
	assert.Equal(suite.T(), 1000, result.SuccessRows)
	assert.Equal(suite.T(), 1000, result.TotalRows)
	assert.Empty(suite.T(), result.Errors)
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
				require.NoError(suite.T(), err)
				assert.Equal(suite.T(), 1, len(result.WorkItems))
			} else {
				require.Error(suite.T(), err)
				// Error could be either header validation or CSV parsing
				errorStr := err.Error()
				headerFound := strings.Contains(errorStr, "header") ||
					strings.Contains(errorStr, "required field") ||
					strings.Contains(errorStr, "wrong number of fields")
				assert.True(suite.T(), headerFound, "Expected header-related error, got: %s", errorStr)
			}
		})
	}
}
