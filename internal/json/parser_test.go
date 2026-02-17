package json

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mrz1836/go-invoice/internal/csv"
	"github.com/mrz1836/go-invoice/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// JSONParserTestSuite defines the test suite for JSON parser functionality
type JSONParserTestSuite struct {
	suite.Suite

	parser      *JSONParser
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

// SetupTest runs before each test
func (suite *JSONParserTestSuite) SetupTest() {
	suite.logger = &MockLogger{}
	suite.validator = &MockValidator{}
	suite.idGenerator = &MockIDGenerator{}
	suite.parser = NewJSONParser(suite.validator, suite.logger, suite.idGenerator)
}

// TestNewJSONParser tests the constructor
func (suite *JSONParserTestSuite) TestNewJSONParser() {
	parser := NewJSONParser(suite.validator, suite.logger, suite.idGenerator)
	suite.NotNil(parser)
	suite.Equal(suite.validator, parser.validator)
	suite.Equal(suite.logger, parser.logger)
	suite.Equal(suite.idGenerator, parser.idGenerator)
}

// TestNewJSONParserWithNilDependencies tests constructor with nil dependencies
func (suite *JSONParserTestSuite) TestNewJSONParserWithNilDependencies() {
	parser := NewJSONParser(nil, nil, nil)
	suite.NotNil(parser)
	suite.Nil(parser.validator)
	suite.Nil(parser.logger)
	suite.Nil(parser.idGenerator)
}

// TestParseTimesheetStructuredFormat tests parsing valid structured JSON format
func (suite *JSONParserTestSuite) TestParseTimesheetStructuredFormat() {
	jsonData := `{
		"metadata": {
			"client": "Acme Corp",
			"period": "January 2024",
			"description": "Monthly timesheet",
			"currency": "USD",
			"total_hours": 16.5,
			"total_amount": 1650.00
		},
		"work_items": [
			{
				"date": "2024-01-15",
				"hours": 8.0,
				"rate": 100.00,
				"description": "Development work",
				"project": "Project A",
				"category": "coding"
			},
			{
				"date": "2024-01-16",
				"hours": 8.5,
				"rate": 100.00,
				"description": "Bug fixes",
				"billable": true
			}
		]
	}`

	reader := strings.NewReader(jsonData)
	options := csv.ParseOptions{}
	ctx := context.Background()

	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().NoError(err)
	suite.Require().NotNil(result)
	suite.Equal(2, result.TotalRows)
	suite.Equal(2, result.SuccessRows)
	suite.Equal(0, result.ErrorRows)
	suite.Len(result.WorkItems, 2)
	suite.Equal("JSON", result.Format)

	// Verify first work item
	firstItem := result.WorkItems[0]
	expectedDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	suite.Equal(expectedDate, firstItem.Date)
	suite.InEpsilon(8.0, firstItem.Hours, 0.001)
	suite.InEpsilon(100.0, firstItem.Rate, 0.001)
	suite.Equal("Development work", firstItem.Description)
	suite.InEpsilon(800.0, firstItem.Total, 0.001)
}

// TestParseTimesheetSimpleFormat tests parsing simple array JSON format
func (suite *JSONParserTestSuite) TestParseTimesheetSimpleFormat() {
	jsonData := `[
		{
			"date": "2024-01-15",
			"hours": 8.0,
			"rate": 100.00,
			"description": "Development work"
		},
		{
			"date": "2024-01-16",
			"hours": 6.5,
			"rate": 100.00,
			"description": "Code review"
		}
	]`

	reader := strings.NewReader(jsonData)
	options := csv.ParseOptions{}
	ctx := context.Background()

	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().NoError(err)
	suite.Require().NotNil(result)
	suite.Equal(2, result.TotalRows)
	suite.Equal(2, result.SuccessRows)
	suite.Equal(0, result.ErrorRows)
	suite.Len(result.WorkItems, 2)
	suite.Equal("JSON", result.Format)
}

// TestParseTimesheetContextCancellation tests context cancellation
func (suite *JSONParserTestSuite) TestParseTimesheetContextCancellation() {
	jsonData := `[{"date": "2024-01-15", "hours": 8.0, "rate": 100.00, "description": "Work"}]`
	reader := strings.NewReader(jsonData)
	options := csv.ParseOptions{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().Error(err)
	suite.Equal(context.Canceled, err)
	suite.Nil(result)
}

// TestParseTimesheetEmptyJSON tests parsing empty JSON data
func (suite *JSONParserTestSuite) TestParseTimesheetEmptyJSON() {
	reader := strings.NewReader("")
	options := csv.ParseOptions{}
	ctx := context.Background()

	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().Error(err)
	suite.Equal(ErrJSONFileEmpty, err)
	suite.Nil(result)
}

// TestParseTimesheetInvalidJSON tests parsing malformed JSON data
func (suite *JSONParserTestSuite) TestParseTimesheetInvalidJSON() {
	jsonData := `{not valid json`
	reader := strings.NewReader(jsonData)
	options := csv.ParseOptions{}
	ctx := context.Background()

	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().Error(err)
	suite.Require().ErrorIs(err, ErrInvalidJSONFormat)
	suite.Nil(result)
}

// TestParseTimesheetEmptyWorkItems tests parsing JSON with empty work items array
func (suite *JSONParserTestSuite) TestParseTimesheetEmptyWorkItems() {
	jsonData := `{"work_items": []}`
	reader := strings.NewReader(jsonData)
	options := csv.ParseOptions{}
	ctx := context.Background()

	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().Error(err)
	suite.Equal(ErrNoWorkItems, err)
	suite.Nil(result)
}

// TestParseTimesheetEmptySimpleArray tests parsing empty simple array format
func (suite *JSONParserTestSuite) TestParseTimesheetEmptySimpleArray() {
	jsonData := `[]`
	reader := strings.NewReader(jsonData)
	options := csv.ParseOptions{}
	ctx := context.Background()

	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().Error(err)
	suite.Equal(ErrNoWorkItems, err)
	suite.Nil(result)
}

// TestParseTimesheetMissingDate tests parsing work item with missing date
func (suite *JSONParserTestSuite) TestParseTimesheetMissingDate() {
	jsonData := `[
		{
			"hours": 8.0,
			"rate": 100.00,
			"description": "Work without date"
		}
	]`

	reader := strings.NewReader(jsonData)
	options := csv.ParseOptions{}
	ctx := context.Background()

	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().NoError(err)
	suite.Require().NotNil(result)
	suite.Equal(1, result.TotalRows)
	suite.Equal(0, result.SuccessRows)
	suite.Equal(1, result.ErrorRows)
	suite.Len(result.Errors, 1)
	suite.Equal("date", result.Errors[0].Column)
	suite.Contains(result.Errors[0].Message, "date is required")
}

// TestParseTimesheetMissingDescription tests parsing work item with missing description
func (suite *JSONParserTestSuite) TestParseTimesheetMissingDescription() {
	jsonData := `[
		{
			"date": "2024-01-15",
			"hours": 8.0,
			"rate": 100.00
		}
	]`

	reader := strings.NewReader(jsonData)
	options := csv.ParseOptions{}
	ctx := context.Background()

	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().NoError(err)
	suite.Require().NotNil(result)
	suite.Equal(1, result.TotalRows)
	suite.Equal(0, result.SuccessRows)
	suite.Equal(1, result.ErrorRows)
	suite.Len(result.Errors, 1)
	suite.Equal("description", result.Errors[0].Column)
	suite.Contains(result.Errors[0].Message, "description is required")
}

// TestParseTimesheetInvalidDate tests parsing work item with invalid date format
func (suite *JSONParserTestSuite) TestParseTimesheetInvalidDate() {
	jsonData := `[
		{
			"date": "not-a-valid-date",
			"hours": 8.0,
			"rate": 100.00,
			"description": "Work with invalid date"
		}
	]`

	reader := strings.NewReader(jsonData)
	options := csv.ParseOptions{}
	ctx := context.Background()

	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().NoError(err)
	suite.Require().NotNil(result)
	suite.Equal(1, result.TotalRows)
	suite.Equal(0, result.SuccessRows)
	suite.Equal(1, result.ErrorRows)
	suite.Len(result.Errors, 1)
	suite.Equal("date", result.Errors[0].Column)
	suite.Contains(result.Errors[0].Message, "invalid date format")
}

// TestParseTimesheetMixedValidInvalid tests parsing with mix of valid and invalid items
func (suite *JSONParserTestSuite) TestParseTimesheetMixedValidInvalid() {
	jsonData := `[
		{
			"date": "2024-01-15",
			"hours": 8.0,
			"rate": 100.00,
			"description": "Valid item"
		},
		{
			"hours": 4.0,
			"rate": 100.00,
			"description": "Missing date"
		},
		{
			"date": "2024-01-17",
			"hours": 6.0,
			"rate": 100.00,
			"description": "Another valid item"
		}
	]`

	reader := strings.NewReader(jsonData)
	options := csv.ParseOptions{}
	ctx := context.Background()

	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().NoError(err)
	suite.Require().NotNil(result)
	suite.Equal(3, result.TotalRows)
	suite.Equal(2, result.SuccessRows)
	suite.Equal(1, result.ErrorRows)
	suite.Len(result.WorkItems, 2)
	suite.Len(result.Errors, 1)
}

// TestParseTimesheetLogging tests that logging works correctly
func (suite *JSONParserTestSuite) TestParseTimesheetLogging() {
	jsonData := `[{"date": "2024-01-15", "hours": 8.0, "rate": 100.00, "description": "Work"}]`
	reader := strings.NewReader(jsonData)
	options := csv.ParseOptions{}
	ctx := context.Background()

	_, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().NoError(err)
	suite.NotEmpty(suite.logger.messages)

	// Find start message
	var foundStart, foundComplete bool
	for _, msg := range suite.logger.messages {
		if msg.Level == "INFO" && strings.Contains(msg.Msg, "starting JSON timesheet parsing") {
			foundStart = true
		}
		if msg.Level == "INFO" && strings.Contains(msg.Msg, "JSON parsing completed") {
			foundComplete = true
		}
	}
	suite.True(foundStart, "Expected to find start logging message")
	suite.True(foundComplete, "Expected to find completion logging message")
}

// TestParseTimesheetMetadataLogging tests that metadata is logged when present
func (suite *JSONParserTestSuite) TestParseTimesheetMetadataLogging() {
	jsonData := `{
		"metadata": {
			"client": "Test Client",
			"period": "January 2024"
		},
		"work_items": [
			{"date": "2024-01-15", "hours": 8.0, "rate": 100.00, "description": "Work"}
		]
	}`

	reader := strings.NewReader(jsonData)
	options := csv.ParseOptions{}
	ctx := context.Background()

	_, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().NoError(err)

	// Find metadata logging message
	var foundMetadata bool
	for _, msg := range suite.logger.messages {
		if msg.Level == "INFO" && strings.Contains(msg.Msg, "JSON metadata found") {
			foundMetadata = true
			break
		}
	}
	suite.True(foundMetadata, "Expected to find metadata logging message")
}

// TestParseDateFormats tests various date format parsing
func (suite *JSONParserTestSuite) TestParseDateFormats() {
	tests := []struct {
		name        string
		dateStr     string
		expectError bool
	}{
		{"ISODate", "2024-01-15", false},
		{"ISODateTime", "2024-01-15T14:30:00", false},
		{"USFormat", "01/15/2024", false},
		{"EUFormat", "15/01/2024", false},
		{"ShortUS", "1/15/2024", false},
		{"ShortEU", "15/1/2024", false},
		{"InvalidFormat", "2024/15/01", true},
		{"EmptyString", "", true},
		{"GarbageString", "not-a-date", true},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			date, err := suite.parser.parseDate(tt.dateStr, nil)

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

// TestParseDateWithCustomFormats tests parsing with custom date formats
func (suite *JSONParserTestSuite) TestParseDateWithCustomFormats() {
	customFormats := []string{
		"02-Jan-2006",
		"2006.01.02",
	}

	tests := []struct {
		name        string
		dateStr     string
		expectError bool
	}{
		{"CustomFormat1", "15-Jan-2024", false},
		{"CustomFormat2", "2024.01.15", false},
		{"UnsupportedFormat", "2024-01-15", true}, // Not in custom formats
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			date, err := suite.parser.parseDate(tt.dateStr, customFormats)

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

// TestDetectFormat tests format detection functionality
func (suite *JSONParserTestSuite) TestDetectFormat() {
	tests := []struct {
		name        string
		data        string
		expectError bool
	}{
		{
			name:        "ValidJSONObject",
			data:        `{"work_items": []}`,
			expectError: false,
		},
		{
			name:        "ValidJSONArray",
			data:        `[{"date": "2024-01-15"}]`,
			expectError: false,
		},
		{
			name:        "ValidJSONString",
			data:        `"just a string"`,
			expectError: false,
		},
		{
			name:        "InvalidJSON",
			data:        `{not valid json`,
			expectError: true,
		},
		{
			name:        "EmptyString",
			data:        ``,
			expectError: true,
		},
		{
			name:        "PlainText",
			data:        `This is not JSON`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			reader := strings.NewReader(tt.data)
			ctx := context.Background()

			format, err := suite.parser.DetectFormat(ctx, reader)

			if tt.expectError {
				suite.Require().Error(err)
				suite.Nil(format)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(format)
				suite.Equal("JSON", format.Name)
				suite.Equal(rune(0), format.Delimiter) // No delimiter for JSON
				suite.False(format.HasHeader)
			}
		})
	}
}

// TestDetectFormatContextCancellation tests format detection with context cancellation
func (suite *JSONParserTestSuite) TestDetectFormatContextCancellation() {
	reader := strings.NewReader(`{"test": true}`)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	format, err := suite.parser.DetectFormat(ctx, reader)

	suite.Require().Error(err)
	suite.Equal(context.Canceled, err)
	suite.Nil(format)
}

// TestValidateFormat tests format validation
func (suite *JSONParserTestSuite) TestValidateFormat() {
	tests := []struct {
		name        string
		data        string
		expectError bool
	}{
		{
			name:        "ValidJSON",
			data:        `{"work_items": []}`,
			expectError: false,
		},
		{
			name:        "InvalidJSON",
			data:        `{not valid}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			reader := strings.NewReader(tt.data)
			ctx := context.Background()

			err := suite.parser.ValidateFormat(ctx, reader)

			if tt.expectError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

// TestValidateFormatContextCancellation tests format validation with context cancellation
func (suite *JSONParserTestSuite) TestValidateFormatContextCancellation() {
	reader := strings.NewReader(`{"test": true}`)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := suite.parser.ValidateFormat(ctx, reader)

	suite.Require().Error(err)
	suite.Equal(context.Canceled, err)
}

// TestParseTimesheetZeroValues tests parsing work items with zero values for hours and rate
func (suite *JSONParserTestSuite) TestParseTimesheetZeroValues() {
	jsonData := `[
		{
			"date": "2024-01-15",
			"hours": 0,
			"rate": 0,
			"description": "Zero hours and rate"
		}
	]`

	reader := strings.NewReader(jsonData)
	options := csv.ParseOptions{}
	ctx := context.Background()

	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().NoError(err)
	suite.Require().NotNil(result)
	suite.Equal(1, result.SuccessRows)
	suite.InDelta(0.0, result.WorkItems[0].Hours, 1e-9)
	suite.InDelta(0.0, result.WorkItems[0].Rate, 1e-9)
	suite.InDelta(0.0, result.WorkItems[0].Total, 1e-9)
}

// TestParseTimesheetNegativeValues tests parsing work items with negative values
func (suite *JSONParserTestSuite) TestParseTimesheetNegativeValues() {
	jsonData := `[
		{
			"date": "2024-01-15",
			"hours": -8.0,
			"rate": 100.00,
			"description": "Negative hours (credit)"
		}
	]`

	reader := strings.NewReader(jsonData)
	options := csv.ParseOptions{}
	ctx := context.Background()

	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().NoError(err)
	suite.Require().NotNil(result)
	suite.Equal(1, result.SuccessRows)
	suite.InEpsilon(-8.0, result.WorkItems[0].Hours, 0.001)
	suite.InEpsilon(-800.0, result.WorkItems[0].Total, 0.001)
}

// TestParseTimesheetLargeFile tests parsing a large JSON file
func (suite *JSONParserTestSuite) TestParseTimesheetLargeFile() {
	// Build large JSON array
	var builder strings.Builder
	builder.WriteString("[")
	for i := 0; i < 100; i++ {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString(fmt.Sprintf(`{
			"date": "2024-01-%02d",
			"hours": 8.0,
			"rate": 100.00,
			"description": "Work item %d"
		}`, (i%28)+1, i+1))
	}
	builder.WriteString("]")

	reader := strings.NewReader(builder.String())
	options := csv.ParseOptions{}
	ctx := context.Background()

	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().NoError(err)
	suite.Equal(100, result.TotalRows)
	suite.Equal(100, result.SuccessRows)
	suite.Len(result.WorkItems, 100)
}

// TestParseTimesheetIDGeneration tests that IDs are generated for each work item
func (suite *JSONParserTestSuite) TestParseTimesheetIDGeneration() {
	jsonData := `[
		{"date": "2024-01-15", "hours": 8.0, "rate": 100.00, "description": "Item 1"},
		{"date": "2024-01-16", "hours": 8.0, "rate": 100.00, "description": "Item 2"}
	]`

	reader := strings.NewReader(jsonData)
	options := csv.ParseOptions{}
	ctx := context.Background()

	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().NoError(err)
	suite.Len(result.WorkItems, 2)

	// Verify unique IDs
	suite.NotEmpty(result.WorkItems[0].ID)
	suite.NotEmpty(result.WorkItems[1].ID)
	suite.NotEqual(result.WorkItems[0].ID, result.WorkItems[1].ID)
}

// TestParseTimesheetCustomIDGenerator tests with custom ID generator
func (suite *JSONParserTestSuite) TestParseTimesheetCustomIDGenerator() {
	suite.idGenerator.generateFunc = func() string {
		return "custom-id-123"
	}
	suite.parser = NewJSONParser(suite.validator, suite.logger, suite.idGenerator)

	jsonData := `[{"date": "2024-01-15", "hours": 8.0, "rate": 100.00, "description": "Work"}]`
	reader := strings.NewReader(jsonData)
	options := csv.ParseOptions{}
	ctx := context.Background()

	result, err := suite.parser.ParseTimesheet(ctx, reader, options)

	suite.Require().NoError(err)
	suite.Equal("custom-id-123", result.WorkItems[0].ID)
}

// TestParseTimesheetCreatedAtTimestamp tests that CreatedAt is set
func (suite *JSONParserTestSuite) TestParseTimesheetCreatedAtTimestamp() {
	jsonData := `[{"date": "2024-01-15", "hours": 8.0, "rate": 100.00, "description": "Work"}]`
	reader := strings.NewReader(jsonData)
	options := csv.ParseOptions{}
	ctx := context.Background()

	beforeParse := time.Now()
	result, err := suite.parser.ParseTimesheet(ctx, reader, options)
	afterParse := time.Now()

	suite.Require().NoError(err)

	createdAt := result.WorkItems[0].CreatedAt
	suite.True(createdAt.After(beforeParse) || createdAt.Equal(beforeParse))
	suite.True(createdAt.Before(afterParse) || createdAt.Equal(afterParse))
}

// TestJSONParserTestSuite runs the JSON parser test suite
func TestJSONParserTestSuite(t *testing.T) {
	suite.Run(t, new(JSONParserTestSuite))
}

// TestParseDateEdgeCases tests edge cases for date parsing
func TestParseDateEdgeCases(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockValidator{}
	idGenerator := &MockIDGenerator{}
	parser := NewJSONParser(validator, logger, idGenerator)

	tests := []struct {
		name         string
		dateStr      string
		wantErr      bool
		expectedYear int
	}{
		{
			name:         "LeapYearDate",
			dateStr:      "2024-02-29",
			wantErr:      false,
			expectedYear: 2024,
		},
		{
			name:    "InvalidLeapYearDate",
			dateStr: "2023-02-29",
			wantErr: true,
		},
		{
			name:    "Day32",
			dateStr: "2024-01-32",
			wantErr: true,
		},
		{
			name:    "Month13",
			dateStr: "2024-13-01",
			wantErr: true,
		},
		// "0001-01-01" parses correctly but IsZero() returns true since
		// Go's time.Time zero value is 0001-01-01 00:00:00 UTC
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.parseDate(tt.dateStr, nil)

			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, result.IsZero())
			} else {
				require.NoError(t, err)
				assert.False(t, result.IsZero())
				if tt.expectedYear > 0 {
					assert.Equal(t, tt.expectedYear, result.Year())
				}
			}
		})
	}
}

// TestParseStructuredFormatDirectly tests the internal structured format parsing
func TestParseStructuredFormatDirectly(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockValidator{}
	idGenerator := &MockIDGenerator{}
	parser := NewJSONParser(validator, logger, idGenerator)

	tests := []struct {
		name         string
		data         []byte
		wantItems    int
		wantMetadata bool
		wantErr      bool
		errContains  string
	}{
		{
			name: "ValidStructuredWithMetadata",
			data: []byte(`{
				"metadata": {"client": "Test"},
				"work_items": [{"date": "2024-01-15", "hours": 8, "rate": 100, "description": "Work"}]
			}`),
			wantItems:    1,
			wantMetadata: true,
			wantErr:      false,
		},
		{
			name: "ValidStructuredNoMetadata",
			data: []byte(`{
				"work_items": [{"date": "2024-01-15", "hours": 8, "rate": 100, "description": "Work"}]
			}`),
			wantItems:    1,
			wantMetadata: false,
			wantErr:      false,
		},
		{
			name:    "NotStructuredFormat",
			data:    []byte(`[{"date": "2024-01-15"}]`),
			wantErr: true,
			// JSON array can't be unmarshaled into struct - returns unmarshal error
		},
		{
			name:    "InvalidJSON",
			data:    []byte(`{invalid`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, metadata, err := parser.parseStructuredFormat(tt.data)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, items, tt.wantItems)
				if tt.wantMetadata {
					assert.NotNil(t, metadata)
				}
			}
		})
	}
}

// TestParseSimpleFormatDirectly tests the internal simple format parsing
func TestParseSimpleFormatDirectly(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockValidator{}
	idGenerator := &MockIDGenerator{}
	parser := NewJSONParser(validator, logger, idGenerator)

	tests := []struct {
		name      string
		data      []byte
		wantItems int
		wantErr   bool
	}{
		{
			name:      "ValidSimpleArray",
			data:      []byte(`[{"date": "2024-01-15", "hours": 8, "rate": 100, "description": "Work"}]`),
			wantItems: 1,
			wantErr:   false,
		},
		{
			name:      "EmptyArray",
			data:      []byte(`[]`),
			wantItems: 0,
			wantErr:   false,
		},
		{
			name:    "InvalidJSON",
			data:    []byte(`[invalid`),
			wantErr: true,
		},
		{
			name:    "NotAnArray",
			data:    []byte(`{"work_items": []}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := parser.parseSimpleFormat(tt.data)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, items, tt.wantItems)
			}
		})
	}
}
