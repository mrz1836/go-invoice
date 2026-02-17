package csv

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mrz1836/go-invoice/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// CSVParserTestSuite defines the test suite for CSV parser functionality
type CSVParserTestSuite struct {
	suite.Suite

	parser      *CSVParser
	validator   *MockValidator
	logger      *MockLogger
	idGenerator *MockIDGenerator
}

// MockLogger provides a test implementation of the Logger interface
type MockLogger struct {
	messages []LogMessage
}

type LogMessage struct {
	Level  string
	Msg    string
	Fields []any
}

func (l *MockLogger) Info(msg string, fields ...any) {
	l.messages = append(l.messages, LogMessage{Level: "INFO", Msg: msg, Fields: fields})
}

func (l *MockLogger) Error(msg string, fields ...any) {
	l.messages = append(l.messages, LogMessage{Level: "ERROR", Msg: msg, Fields: fields})
}

func (l *MockLogger) Debug(msg string, fields ...any) {
	l.messages = append(l.messages, LogMessage{Level: "DEBUG", Msg: msg, Fields: fields})
}

func (l *MockLogger) Reset() {
	l.messages = nil
}

// MockValidator provides a test implementation of the CSVValidator interface
type MockValidator struct {
	validateWorkItemFunc func(ctx context.Context, item *models.WorkItem) error
	validateRowFunc      func(ctx context.Context, row []string, lineNum int) error
	validateBatchFunc    func(ctx context.Context, items []models.WorkItem) error
}

// MockIDGenerator provides a test implementation of the IDGenerator interface
type MockIDGenerator struct {
	generateFunc func() string
	counter      int
}

func (m *MockIDGenerator) GenerateID() string {
	if m.generateFunc != nil {
		return m.generateFunc()
	}
	m.counter++
	return fmt.Sprintf("test-id-%d", m.counter)
}

func (v *MockValidator) ValidateWorkItem(ctx context.Context, item *models.WorkItem) error {
	if v.validateWorkItemFunc != nil {
		return v.validateWorkItemFunc(ctx, item)
	}
	return nil
}

func (v *MockValidator) ValidateRow(ctx context.Context, row []string, lineNum int) error {
	if v.validateRowFunc != nil {
		return v.validateRowFunc(ctx, row, lineNum)
	}
	return nil
}

func (v *MockValidator) ValidateBatch(ctx context.Context, items []models.WorkItem) error {
	if v.validateBatchFunc != nil {
		return v.validateBatchFunc(ctx, items)
	}
	return nil
}

// SetupTest runs before each test
func (suite *CSVParserTestSuite) SetupTest() {
	suite.logger = &MockLogger{}
	suite.validator = &MockValidator{}
	suite.idGenerator = &MockIDGenerator{}
	suite.parser = NewCSVParser(suite.validator, suite.logger, suite.idGenerator)
}

// TestNewCSVParser tests the constructor
func (suite *CSVParserTestSuite) TestNewCSVParser() {
	parser := NewCSVParser(suite.validator, suite.logger, suite.idGenerator)
	suite.NotNil(parser)
	suite.Equal(suite.validator, parser.validator)
	suite.Equal(suite.logger, parser.logger)
	suite.Equal(suite.idGenerator, parser.idGenerator)
}

// TestParseTimesheetValidCSV tests parsing valid CSV data
func (suite *CSVParserTestSuite) TestParseTimesheetValidCSV() {
	// Generate valid dates for test data
	validDate := time.Now().AddDate(-1, 0, 0)
	date1 := validDate.Format("2006-01-02")
	date2 := validDate.AddDate(0, 0, 1).Format("2006-01-02")
	date3 := validDate.AddDate(0, 0, 2).Format("2006-01-02")

	csvData := fmt.Sprintf(`Date,Hours,Rate,Description
%s,8.0,100.00,Development work
%s,6.5,100.00,Bug fixes and testing
%s,4.0,100.00,Code review`, date1, date2, date3)

	reader := strings.NewReader(csvData)
	options := ParseOptions{
		Format:          "standard",
		ContinueOnError: false,
	}

	ctx := context.Background()
	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().NoError(err)
	suite.Require().NotNil(result)
	suite.Equal(3, result.TotalRows)
	suite.Equal(3, result.SuccessRows)
	suite.Equal(0, result.ErrorRows)
	suite.Len(result.WorkItems, 3)
	suite.Equal("standard", result.Format)

	// Verify first work item
	firstItem := result.WorkItems[0]
	expectedDate := validDate
	suite.Equal(expectedDate.Year(), firstItem.Date.Year())
	suite.Equal(expectedDate.Month(), firstItem.Date.Month())
	suite.Equal(expectedDate.Day(), firstItem.Date.Day())
	suite.InEpsilon(8.0, firstItem.Hours, 0.001)
	suite.InEpsilon(100.0, firstItem.Rate, 0.001)
	suite.Equal("Development work", firstItem.Description)
	suite.InEpsilon(800.0, firstItem.Total, 0.001)
}

// TestParseTimesheetContextCancellation tests context cancellation
func (suite *CSVParserTestSuite) TestParseTimesheetContextCancellation() {
	validDate := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	csvData := fmt.Sprintf(`Date,Hours,Rate,Description
%s,8.0,100.00,Development work`, validDate)

	reader := strings.NewReader(csvData)
	options := ParseOptions{Format: "standard"}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().Error(err)
	suite.Equal(context.Canceled, err)
	suite.Nil(result)
}

// TestParseTimesheetEmptyCSV tests parsing empty CSV data
func (suite *CSVParserTestSuite) TestParseTimesheetEmptyCSV() {
	reader := strings.NewReader("")
	options := ParseOptions{Format: "standard"}

	ctx := context.Background()
	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().Error(err)
	suite.Contains(err.Error(), "CSV file is empty")
	suite.Nil(result)
}

// TestParseTimesheetInvalidCSV tests parsing malformed CSV data
func (suite *CSVParserTestSuite) TestParseTimesheetInvalidCSV() {
	validDate := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	csvData := fmt.Sprintf(`Date,Hours,Rate,Description
%s,invalid_hours,100.00,Development work`, validDate)

	reader := strings.NewReader(csvData)
	options := ParseOptions{
		Format:          "standard",
		ContinueOnError: false,
	}

	ctx := context.Background()
	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().Error(err)
	suite.Contains(err.Error(), "parsing failed at line 2")
	suite.Nil(result)
}

// TestParseTimesheetContinueOnError tests parsing with error continuation
func (suite *CSVParserTestSuite) TestParseTimesheetContinueOnError() {
	validDate := time.Now().AddDate(-1, 0, 0)
	date1 := validDate.Format("2006-01-02")
	date2 := validDate.AddDate(0, 0, 1).Format("2006-01-02")
	date3 := validDate.AddDate(0, 0, 2).Format("2006-01-02")

	csvData := fmt.Sprintf(`Date,Hours,Rate,Description
%s,8.0,100.00,Development work
%s,invalid_hours,100.00,Bug fixes
%s,4.0,100.00,Code review`, date1, date2, date3)

	reader := strings.NewReader(csvData)
	options := ParseOptions{
		Format:          "standard",
		ContinueOnError: true,
	}

	ctx := context.Background()
	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().NoError(err)
	suite.Require().NotNil(result)
	suite.Equal(3, result.TotalRows)
	suite.Equal(2, result.SuccessRows)
	suite.Equal(1, result.ErrorRows)
	suite.Len(result.WorkItems, 2)
	suite.Len(result.Errors, 1)

	// Verify error details
	parseError := result.Errors[0]
	suite.Equal(3, parseError.Line) // Line 3 has the invalid hours
	suite.Contains(parseError.Message, "invalid hours")
}

// TestParseTimesheetValidationError tests parsing with validation errors
func (suite *CSVParserTestSuite) TestParseTimesheetValidationError() {
	validDate := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	csvData := fmt.Sprintf(`Date,Hours,Rate,Description
%s,8.0,100.00,Development work`, validDate)

	// Mock validator to return error
	suite.validator.validateWorkItemFunc = func(_ context.Context, _ *models.WorkItem) error {
		return fmt.Errorf("validation error") //nolint:err113 // Test-specific error
	}

	reader := strings.NewReader(csvData)
	options := ParseOptions{
		Format:          "standard",
		ContinueOnError: false,
	}

	ctx := context.Background()
	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().Error(err)
	suite.Contains(err.Error(), "validation failed at line 2")
	suite.Nil(result)
}

// TestParseTimesheetDifferentFormats tests parsing different CSV formats
func (suite *CSVParserTestSuite) TestParseTimesheetDifferentFormats() {
	validDate := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")

	tests := []struct {
		name     string
		csvData  string
		format   string
		expected int
	}{
		{
			name: "TabSeparatedValues",
			csvData: "Date\tHours\tRate\tDescription\n" +
				validDate + "\t8.0\t100.00\tDevelopment work",
			format:   "tab",
			expected: 1,
		},
		{
			name: "SemicolonSeparated",
			csvData: "Date;Hours;Rate;Description\n" +
				validDate + ";8.0;100.00;Development work",
			format:   "semicolon",
			expected: 1,
		},
		{
			name: "ExcelFormat",
			csvData: fmt.Sprintf(`Date,Hours,Rate,Description
"%s","8.0","100.00","Development work"`, validDate),
			format:   "excel",
			expected: 1,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			reader := strings.NewReader(tt.csvData)
			options := ParseOptions{Format: tt.format}

			ctx := context.Background()
			result, err := suite.parser.ParseTimesheet(ctx, reader, options)

			suite.Require().NoError(err)
			suite.Equal(tt.expected, result.SuccessRows)
		})
	}
}

// TestParseTimesheetHeaderVariations tests different header name variations
func (suite *CSVParserTestSuite) TestParseTimesheetHeaderVariations() {
	validDate := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")

	tests := []struct {
		name    string
		headers string
		wantErr bool
	}{
		{
			name:    "StandardHeaders",
			headers: "Date,Hours,Rate,Description",
			wantErr: false,
		},
		{
			name:    "AlternativeHeaders",
			headers: "Work_Date,Duration,Hourly_Rate,Task",
			wantErr: false,
		},
		{
			name:    "CaseInsensitiveHeaders",
			headers: "DATE,HOURS,RATE,DESCRIPTION",
			wantErr: false,
		},
		{
			name:    "MissingRequiredHeader",
			headers: "Date,Hours,Description", // Missing Rate
			wantErr: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			csvData := tt.headers + "\n" + validDate + ",8.0,100.00,Development work"
			reader := strings.NewReader(csvData)
			options := ParseOptions{Format: "standard"}

			ctx := context.Background()
			result, err := suite.parser.ParseTimesheet(ctx, reader, options)

			if tt.wantErr {
				suite.Require().Error(err)
				suite.Nil(result)
			} else {
				suite.Require().NoError(err)
				suite.NotNil(result)
			}
		})
	}
}

// TestDetectFormat tests format detection functionality
func (suite *CSVParserTestSuite) TestDetectFormat() {
	validDate := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")

	tests := []struct {
		name           string
		csvData        string
		expectedFormat string
		expectedDelim  rune
	}{
		{
			name:           "StandardCSV",
			csvData:        "Date,Hours,Rate,Description\n" + validDate + ",8.0,100.00,Work",
			expectedFormat: "standard",
			expectedDelim:  ',',
		},
		{
			name:           "TabSeparated",
			csvData:        "Date\tHours\tRate\tDescription\n" + validDate + "\t8.0\t100.00\tWork",
			expectedFormat: "tab",
			expectedDelim:  '\t',
		},
		{
			name:           "SemicolonSeparated",
			csvData:        "Date;Hours;Rate;Description\n" + validDate + ";8.0;100.00;Work",
			expectedFormat: "semicolon",
			expectedDelim:  ';',
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			reader := strings.NewReader(tt.csvData)
			ctx := context.Background()

			format, err := suite.parser.DetectFormat(ctx, reader)

			suite.Require().NoError(err)
			suite.Equal(tt.expectedFormat, format.Name)
			suite.Equal(tt.expectedDelim, format.Delimiter)
		})
	}
}

// TestDetectFormatContextCancellation tests format detection with context cancellation
func (suite *CSVParserTestSuite) TestDetectFormatContextCancellation() {
	reader := strings.NewReader("Date,Hours,Rate,Description\n")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	format, err := suite.parser.DetectFormat(ctx, reader)

	suite.Require().Error(err)
	suite.Equal(context.Canceled, err)
	suite.Nil(format)
}

// TestDetectFormatEmptyFile tests format detection with empty file
func (suite *CSVParserTestSuite) TestDetectFormatEmptyFile() {
	reader := strings.NewReader("")
	ctx := context.Background()

	format, err := suite.parser.DetectFormat(ctx, reader)

	suite.Require().Error(err)
	suite.Contains(err.Error(), "cannot detect format of empty file")
	suite.Nil(format)
}

// TestValidateFormat tests format validation
func (suite *CSVParserTestSuite) TestValidateFormat() {
	// Test supported format
	reader := strings.NewReader("Date,Hours,Rate,Description\n")
	ctx := context.Background()

	err := suite.parser.ValidateFormat(ctx, reader)
	suite.Require().NoError(err)
}

// TestValidateFormatContextCancellation tests format validation with context cancellation
func (suite *CSVParserTestSuite) TestValidateFormatContextCancellation() {
	reader := strings.NewReader("Date,Hours,Rate,Description\n")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := suite.parser.ValidateFormat(ctx, reader)

	suite.Require().Error(err)
	suite.Equal(context.Canceled, err)
}

// TestParseDateFormats tests various date format parsing
func (suite *CSVParserTestSuite) TestParseDateFormats() {
	// Generate valid date for test data (using day 15 to avoid ambiguity)
	validDate := time.Date(time.Now().Year(), time.Now().Month(), 15, 0, 0, 0, 0, time.UTC).AddDate(-1, 0, 0)
	isoDate := validDate.Format("2006-01-02")
	usDate := validDate.Format("01/02/2006")
	euDate := validDate.Format("02/01/2006")
	monthName := validDate.Format("Jan 2, 2006")
	fullMonth := validDate.Format("January 2, 2006")
	withTime := validDate.Format("2006-01-02 15:04:05")

	tests := []struct {
		name        string
		dateStr     string
		expectError bool
	}{
		{"ISO8601", isoDate, false},
		{"USFormat", usDate, false},
		{"EUFormat", euDate, false},
		{"MonthName", monthName, false},
		{"FullMonthName", fullMonth, false},
		{"WithTime", withTime, false},
		{"InvalidFormat", "2024/15/01", true},
		{"EmptyString", "", true},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			date, err := suite.parser.parseDate(tt.dateStr)

			if tt.expectError {
				suite.Require().Error(err)
				suite.True(date.IsZero())
			} else {
				suite.Require().NoError(err)
				suite.False(date.IsZero())
			}
		})
	}
}

// TestNormalizeHeaderName tests header name normalization
func (suite *CSVParserTestSuite) TestNormalizeHeaderName() {
	tests := []struct {
		input    string
		expected string
	}{
		{"Date", "date"},
		{"HOURS", "hours"},
		{"Hourly_Rate", "rate"},
		{"Work_Description", "description"},
		{"  Task  ", "description"},
		{"billing_rate", "rate"},
		{"time", "hours"},
		{"notes", "description"},
		{"unknown_field", "unknown_field"},
	}

	for _, tt := range tests {
		suite.Run(tt.input, func() {
			result := suite.parser.normalizeHeaderName(tt.input)
			suite.Equal(tt.expected, result)
		})
	}
}

// TestParseTimesheetLargeFile tests parsing a large CSV file with context checks
func (suite *CSVParserTestSuite) TestParseTimesheetLargeFile() {
	// Create a large CSV with many rows
	validDate := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	var csvBuilder strings.Builder
	csvBuilder.WriteString("Date,Hours,Rate,Description\n")

	// Add 100 rows
	for i := 1; i <= 100; i++ {
		csvBuilder.WriteString(validDate + ",8.0,100.00,Work item ")
		csvBuilder.WriteRune(rune('0' + i%10))
		csvBuilder.WriteString("\n")
	}

	reader := strings.NewReader(csvBuilder.String())
	options := ParseOptions{Format: "standard"}

	ctx := context.Background()
	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().NoError(err)
	suite.Equal(100, result.TotalRows)
	suite.Equal(100, result.SuccessRows)
	suite.Len(result.WorkItems, 100)
}

// TestParseTimesheetMissingFields tests parsing CSV with missing fields
func (suite *CSVParserTestSuite) TestParseTimesheetMissingFields() {
	validDate := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	csvData := fmt.Sprintf(`Date,Hours,Rate,Description
%s,,100.00,Development work`, validDate)

	reader := strings.NewReader(csvData)
	options := ParseOptions{
		Format:          "standard",
		ContinueOnError: true,
	}

	ctx := context.Background()
	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().NoError(err)
	suite.Equal(1, result.TotalRows)
	suite.Equal(0, result.SuccessRows)
	suite.Equal(1, result.ErrorRows)
	suite.Len(result.Errors, 1)
	suite.Contains(result.Errors[0].Message, "field is empty: hours")
}

// TestParseTimesheetLogging tests that logging works correctly
func (suite *CSVParserTestSuite) TestParseTimesheetLogging() {
	validDate := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	csvData := fmt.Sprintf(`Date,Hours,Rate,Description
%s,8.0,100.00,Development work`, validDate)

	reader := strings.NewReader(csvData)
	options := ParseOptions{Format: "standard"}

	ctx := context.Background()
	_, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().NoError(err)

	// Check that logging occurred
	suite.NotEmpty(suite.logger.messages)

	// Find start message
	var foundStart bool
	for _, msg := range suite.logger.messages {
		if msg.Level == "INFO" && strings.Contains(msg.Msg, "starting timesheet parsing") {
			foundStart = true
			break
		}
	}
	suite.True(foundStart, "Expected to find start logging message")
}

// TestCSVParserTestSuite runs the CSV parser test suite
func TestCSVParserTestSuite(t *testing.T) {
	suite.Run(t, new(CSVParserTestSuite))
}

// TestParseRow tests individual row parsing functionality
func TestParseRow(t *testing.T) {
	validDate := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	logger := &MockLogger{}
	validator := &MockValidator{}
	idGenerator := &MockIDGenerator{}
	parser := NewCSVParser(validator, logger, idGenerator)

	headerMap := map[string]int{
		"date":        0,
		"hours":       1,
		"rate":        2,
		"description": 3,
	}

	tests := []struct {
		name    string
		row     []string
		lineNum int
		wantErr bool
	}{
		{
			name:    "ValidRow",
			row:     []string{validDate, "8.0", "100.00", "Development work"},
			lineNum: 1,
			wantErr: false,
		},
		{
			name:    "EmptyRow",
			row:     []string{},
			lineNum: 1,
			wantErr: true,
		},
		{
			name:    "InvalidHours",
			row:     []string{validDate, "invalid", "100.00", "Development work"},
			lineNum: 1,
			wantErr: true,
		},
		{
			name:    "InvalidRate",
			row:     []string{validDate, "8.0", "invalid", "Development work"},
			lineNum: 1,
			wantErr: true,
		},
		{
			name:    "InvalidDate",
			row:     []string{"invalid-date", "8.0", "100.00", "Development work"},
			lineNum: 1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			workItem, err := parser.parseRow(ctx, tt.row, headerMap, tt.lineNum)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, workItem)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, workItem)
			}
		})
	}
}

// TestParseDateFormats tests comprehensive date parsing with various formats
func TestParseDateFormats(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockValidator{}
	idGenerator := &MockIDGenerator{}
	parser := NewCSVParser(validator, logger, idGenerator)

	now := time.Now()

	// Helper function to calculate expected year for no-year date formats
	// This mirrors the parser's logic: use current year unless >6 months in future
	expectedYearForNoYearDate := func(month time.Month, day int) int {
		candidateDate := time.Date(now.Year(), month, day, 0, 0, 0, 0, time.UTC)
		sixMonthsFromNow := now.AddDate(0, 6, 0)
		if candidateDate.After(sixMonthsFromNow) {
			return now.Year() - 1
		}
		return now.Year()
	}

	tests := []struct {
		name         string
		dateStr      string
		wantErr      bool
		checkYear    bool
		expectedYear int
		description  string
	}{
		// Full 4-digit year formats (most reliable)
		{
			name:         "ISO format YYYY-MM-DD",
			dateStr:      "2025-09-08",
			wantErr:      false,
			checkYear:    true,
			expectedYear: 2025,
			description:  "Standard ISO 8601 format",
		},
		{
			name:         "US format MM/DD/YYYY",
			dateStr:      "09/08/2025",
			wantErr:      false,
			checkYear:    true,
			expectedYear: 2025,
			description:  "US date format with full year",
		},
		{
			name:         "EU format DD/MM/YYYY",
			dateStr:      "08/09/2025",
			wantErr:      false,
			checkYear:    true,
			expectedYear: 2025,
			description:  "European date format with full year",
		},

		// 2-digit year formats (with smart inference)
		{
			name:         "2-digit year 24 → 2024",
			dateStr:      "9/8/24",
			wantErr:      false,
			checkYear:    true,
			expectedYear: 2024,
			description:  "2-digit year 00-50 maps to 2000-2050",
		},
		{
			name:         "2-digit year 25 → 2025",
			dateStr:      "9/8/25",
			wantErr:      false,
			checkYear:    true,
			expectedYear: 2025,
			description:  "2-digit year 25 maps to 2025",
		},
		{
			name:         "2-digit year 99 → 1999",
			dateStr:      "9/8/99",
			wantErr:      false,
			checkYear:    true,
			expectedYear: 1999,
			description:  "2-digit year 51-99 maps to 1951-1999",
		},
		{
			name:         "2-digit year with leading zeros",
			dateStr:      "01/02/06",
			wantErr:      false,
			checkYear:    true,
			expectedYear: 2006,
			description:  "Padded format with 2-digit year",
		},

		// No-year formats (smart current/previous year inference)
		{
			name:         "No year - month/day only",
			dateStr:      "9/8",
			wantErr:      false,
			checkYear:    true,
			expectedYear: expectedYearForNoYearDate(time.September, 8),
			description:  "Date without year uses current year unless >6 months in future",
		},
		{
			name:         "No year - with leading zeros",
			dateStr:      "01/15",
			wantErr:      false,
			checkYear:    true,
			expectedYear: expectedYearForNoYearDate(time.January, 15),
			description:  "Padded date without year",
		},
		{
			name:         "Month name without year",
			dateStr:      "Sep 8",
			wantErr:      false,
			checkYear:    true,
			expectedYear: expectedYearForNoYearDate(time.September, 8),
			description:  "Month name format without year uses smart inference",
		},

		// Edge cases and invalid formats
		{
			name:        "Invalid date format",
			dateStr:     "not-a-date",
			wantErr:     true,
			description: "Completely invalid date string",
		},
		{
			name:        "Empty string",
			dateStr:     "",
			wantErr:     true,
			description: "Empty date string should error",
		},
		{
			name:        "Whitespace only",
			dateStr:     "   ",
			wantErr:     true,
			description: "Whitespace-only string should error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.parseDate(tt.dateStr)

			if tt.wantErr {
				require.Error(t, err, tt.description)
				assert.True(t, result.IsZero(), "Error case should return zero time")
			} else {
				require.NoError(t, err, tt.description)
				assert.False(t, result.IsZero(), "Valid date should not be zero time")

				if tt.checkYear {
					assert.Equal(t, tt.expectedYear, result.Year(),
						"Year should be %d for input '%s': %s",
						tt.expectedYear, tt.dateStr, tt.description)
				}
			}
		})
	}
}

// TestParseDateFutureInference tests the smart year inference for no-year dates
func TestParseDateFutureInference(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockValidator{}
	idGenerator := &MockIDGenerator{}
	parser := NewCSVParser(validator, logger, idGenerator)

	now := time.Now()

	// Test date that would be >6 months in future (should use previous year)
	// Use a date that's definitely >6 months from now: 8 months from today
	eightMonthsFromNow := now.AddDate(0, 8, 0)
	dateStr := fmt.Sprintf("%d/%d", int(eightMonthsFromNow.Month()), eightMonthsFromNow.Day())

	result, err := parser.parseDate(dateStr)
	require.NoError(t, err)

	// The parser should use current year if <=6 months in future, or previous year if >6 months
	sixMonthsFromNow := now.AddDate(0, 6, 0)
	candidateCurrentYear := time.Date(now.Year(), eightMonthsFromNow.Month(), eightMonthsFromNow.Day(), 0, 0, 0, 0, time.UTC)

	// If the date with current year is more than 6 months from now, use previous year
	if candidateCurrentYear.After(sixMonthsFromNow) {
		assert.Equal(t, now.Year()-1, result.Year(),
			"Date >6 months in future should use previous year")
	} else {
		assert.Equal(t, now.Year(), result.Year(),
			"Date <=6 months in future should use current year")
	}
}

// TestParseDateWithSpaces tests date parsing with extra whitespace
func TestParseDateWithSpaces(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockValidator{}
	idGenerator := &MockIDGenerator{}
	parser := NewCSVParser(validator, logger, idGenerator)

	tests := []struct {
		name    string
		dateStr string
		wantErr bool
	}{
		{
			name:    "Leading spaces",
			dateStr: "  2025-09-08",
			wantErr: false,
		},
		{
			name:    "Trailing spaces",
			dateStr: "2025-09-08  ",
			wantErr: false,
		},
		{
			name:    "Both leading and trailing",
			dateStr: "  2025-09-08  ",
			wantErr: false,
		},
		{
			name:    "Tabs",
			dateStr: "\t2025-09-08\t",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.parseDate(tt.dateStr)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, 2025, result.Year())
				assert.Equal(t, time.Month(9), result.Month())
				assert.Equal(t, 8, result.Day())
			}
		})
	}
}
