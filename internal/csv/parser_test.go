package csv

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mrz/go-invoice/internal/models"
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
	csvData := `Date,Hours,Rate,Description
2024-01-15,8.0,100.00,Development work
2024-01-16,6.5,100.00,Bug fixes and testing
2024-01-17,4.0,100.00,Code review`

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
	expectedDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	suite.Equal(expectedDate, firstItem.Date)
	suite.InEpsilon(8.0, firstItem.Hours, 0.001)
	suite.InEpsilon(100.0, firstItem.Rate, 0.001)
	suite.Equal("Development work", firstItem.Description)
	suite.InEpsilon(800.0, firstItem.Total, 0.001)
}

// TestParseTimesheetContextCancellation tests context cancellation
func (suite *CSVParserTestSuite) TestParseTimesheetContextCancellation() {
	csvData := `Date,Hours,Rate,Description
2024-01-15,8.0,100.00,Development work`

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
	csvData := `Date,Hours,Rate,Description
2024-01-15,invalid_hours,100.00,Development work`

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
	csvData := `Date,Hours,Rate,Description
2024-01-15,8.0,100.00,Development work
2024-01-16,invalid_hours,100.00,Bug fixes
2024-01-17,4.0,100.00,Code review`

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
	csvData := `Date,Hours,Rate,Description
2024-01-15,8.0,100.00,Development work`

	// Mock validator to return error
	suite.validator.validateWorkItemFunc = func(ctx context.Context, item *models.WorkItem) error {
		return fmt.Errorf("validation error")
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
	tests := []struct {
		name     string
		csvData  string
		format   string
		expected int
	}{
		{
			name: "TabSeparatedValues",
			csvData: "Date\tHours\tRate\tDescription\n" +
				"2024-01-15\t8.0\t100.00\tDevelopment work",
			format:   "tab",
			expected: 1,
		},
		{
			name: "SemicolonSeparated",
			csvData: "Date;Hours;Rate;Description\n" +
				"2024-01-15;8.0;100.00;Development work",
			format:   "semicolon",
			expected: 1,
		},
		{
			name: "ExcelFormat",
			csvData: `Date,Hours,Rate,Description
"2024-01-15","8.0","100.00","Development work"`,
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
			csvData := tt.headers + "\n2024-01-15,8.0,100.00,Development work"
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
	tests := []struct {
		name           string
		csvData        string
		expectedFormat string
		expectedDelim  rune
	}{
		{
			name:           "StandardCSV",
			csvData:        "Date,Hours,Rate,Description\n2024-01-15,8.0,100.00,Work",
			expectedFormat: "standard",
			expectedDelim:  ',',
		},
		{
			name:           "TabSeparated",
			csvData:        "Date\tHours\tRate\tDescription\n2024-01-15\t8.0\t100.00\tWork",
			expectedFormat: "tab",
			expectedDelim:  '\t',
		},
		{
			name:           "SemicolonSeparated",
			csvData:        "Date;Hours;Rate;Description\n2024-01-15;8.0;100.00;Work",
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
	tests := []struct {
		name        string
		dateStr     string
		expectError bool
	}{
		{"ISO8601", "2024-01-15", false},
		{"USFormat", "01/15/2024", false},
		{"EUFormat", "15/01/2024", false},
		{"MonthName", "Jan 15, 2024", false},
		{"FullMonthName", "January 15, 2024", false},
		{"WithTime", "2024-01-15 14:30:00", false},
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
	var csvBuilder strings.Builder
	csvBuilder.WriteString("Date,Hours,Rate,Description\n")

	// Add 100 rows
	for i := 1; i <= 100; i++ {
		csvBuilder.WriteString("2024-01-15,8.0,100.00,Work item ")
		csvBuilder.WriteString(string(rune('0' + i%10)))
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
	csvData := `Date,Hours,Rate,Description
2024-01-15,,100.00,Development work`

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
	csvData := `Date,Hours,Rate,Description
2024-01-15,8.0,100.00,Development work`

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
			row:     []string{"2024-01-15", "8.0", "100.00", "Development work"},
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
			row:     []string{"2024-01-15", "invalid", "100.00", "Development work"},
			lineNum: 1,
			wantErr: true,
		},
		{
			name:    "InvalidRate",
			row:     []string{"2024-01-15", "8.0", "invalid", "Development work"},
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
