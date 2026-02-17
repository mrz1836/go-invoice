package csv

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/mrz1836/go-invoice/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// CSVTypesTestSuite defines the test suite for CSV type validation
type CSVTypesTestSuite struct {
	suite.Suite
}

// TestCSVTypesTestSuite runs the CSV types test suite
func TestCSVTypesTestSuite(t *testing.T) {
	suite.Run(t, new(CSVTypesTestSuite))
}

// TestParseOptionsJSONMarshaling tests JSON marshaling/unmarshaling for ParseOptions
func (suite *CSVTypesTestSuite) TestParseOptionsJSONMarshaling() {
	tests := []struct {
		name    string
		options ParseOptions
	}{
		{
			name: "CompleteParseOptions",
			options: ParseOptions{
				Format:          "excel",
				ContinueOnError: true,
				SkipEmptyRows:   true,
				DateFormat:      "2006-01-02",
			},
		},
		{
			name: "MinimalParseOptions",
			options: ParseOptions{
				Format: "standard",
			},
		},
		{
			name: "CustomDateFormat",
			options: ParseOptions{
				Format:          "tab",
				ContinueOnError: false,
				SkipEmptyRows:   false,
				DateFormat:      "01/02/2006",
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			jsonData, err := json.Marshal(tt.options)
			suite.Require().NoError(err)

			var unmarshaled ParseOptions
			err = json.Unmarshal(jsonData, &unmarshaled)
			suite.Require().NoError(err)

			suite.Equal(tt.options, unmarshaled)
		})
	}
}

// TestParseResultJSONMarshaling tests JSON marshaling/unmarshaling for ParseResult
func (suite *CSVTypesTestSuite) TestParseResultJSONMarshaling() {
	workItem := models.WorkItem{
		ID:          "test-1",
		Date:        time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Hours:       8.5,
		Rate:        75.0,
		Description: "Software development",
		Total:       637.50,
		CreatedAt:   time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	parseError := ParseError{
		Line:       5,
		Column:     "Hours",
		Value:      "invalid",
		Message:    "Invalid number format",
		Suggestion: "Use decimal format like 8.5",
		Row:        []string{"2025-01-15", "invalid", "75.0", "Test work"},
	}

	result := ParseResult{
		WorkItems:   []models.WorkItem{workItem},
		TotalRows:   10,
		SuccessRows: 9,
		ErrorRows:   1,
		Errors:      []ParseError{parseError},
		HeaderMap:   map[string]int{"Date": 0, "Hours": 1, "Rate": 2, "Description": 3},
		Format:      "standard",
	}

	jsonData, err := json.Marshal(result)
	suite.Require().NoError(err)

	var unmarshaled ParseResult
	err = json.Unmarshal(jsonData, &unmarshaled)
	suite.Require().NoError(err)

	suite.Equal(result.TotalRows, unmarshaled.TotalRows)
	suite.Equal(result.SuccessRows, unmarshaled.SuccessRows)
	suite.Equal(result.ErrorRows, unmarshaled.ErrorRows)
	suite.Equal(result.Format, unmarshaled.Format)
	suite.Len(unmarshaled.WorkItems, 1)
	suite.Len(unmarshaled.Errors, 1)
	suite.Equal(parseError.Line, unmarshaled.Errors[0].Line)
	suite.Equal(parseError.Message, unmarshaled.Errors[0].Message)
}

// TestParseErrorJSONMarshaling tests JSON marshaling/unmarshaling for ParseError
func (suite *CSVTypesTestSuite) TestParseErrorJSONMarshaling() {
	tests := []struct {
		name  string
		error ParseError
	}{
		{
			name: "CompleteParseError",
			error: ParseError{
				Line:       10,
				Column:     "Rate",
				Value:      "not-a-number",
				Message:    "Invalid rate format",
				Suggestion: "Enter a decimal number like 75.00",
				Row:        []string{"2023-01-01", "8", "not-a-number", "Work description"},
			},
		},
		{
			name: "MinimalParseError",
			error: ParseError{
				Line:    1,
				Column:  "Date",
				Value:   "invalid-date",
				Message: "Invalid date format",
			},
		},
		{
			name: "EmptyRowError",
			error: ParseError{
				Line:       25,
				Column:     "",
				Value:      "",
				Message:    "Empty row encountered",
				Suggestion: "Remove empty rows or enable skip_empty_rows option",
				Row:        []string{},
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			jsonData, err := json.Marshal(tt.error)
			suite.Require().NoError(err)

			var unmarshaled ParseError
			err = json.Unmarshal(jsonData, &unmarshaled)
			suite.Require().NoError(err)

			suite.Equal(tt.error, unmarshaled)
		})
	}
}

// TestFormatInfoJSONMarshaling tests JSON marshaling/unmarshaling for FormatInfo
func (suite *CSVTypesTestSuite) TestFormatInfoJSONMarshaling() {
	tests := []struct {
		name   string
		format FormatInfo
	}{
		{
			name: "StandardCSVFormat",
			format: FormatInfo{
				Name:      "standard",
				Delimiter: ',',
				HasHeader: true,
				Encoding:  "UTF-8",
			},
		},
		{
			name: "TabDelimitedFormat",
			format: FormatInfo{
				Name:      "tab",
				Delimiter: '\t',
				HasHeader: true,
				Encoding:  "UTF-8",
			},
		},
		{
			name: "SemicolonDelimitedFormat",
			format: FormatInfo{
				Name:      "excel",
				Delimiter: ';',
				HasHeader: false,
				Encoding:  "ISO-8859-1",
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			jsonData, err := json.Marshal(tt.format)
			suite.Require().NoError(err)

			var unmarshaled FormatInfo
			err = json.Unmarshal(jsonData, &unmarshaled)
			suite.Require().NoError(err)

			suite.Equal(tt.format, unmarshaled)
		})
	}
}

// TestValidationRuleJSONMarshaling tests JSON marshaling/unmarshaling for ValidationRule
func (suite *CSVTypesTestSuite) TestValidationRuleJSONMarshaling() {
	// Functions cannot be marshaled to JSON, so they should be omitted
	rule := ValidationRule{
		Name:        "hours_positive",
		Description: "Hours must be positive",
		Validator: func(ctx context.Context, item *models.WorkItem) error {
			if item.Hours <= 0 {
				return assert.AnError
			}
			return nil
		},
		RowValidator: func(ctx context.Context, row []string, lineNum int) error {
			return nil
		},
	}

	jsonData, err := json.Marshal(rule)
	suite.Require().NoError(err)

	var unmarshaled ValidationRule
	err = json.Unmarshal(jsonData, &unmarshaled)
	suite.Require().NoError(err)

	// Functions should be nil after unmarshaling
	suite.Equal(rule.Name, unmarshaled.Name)
	suite.Equal(rule.Description, unmarshaled.Description)
	suite.Nil(unmarshaled.Validator)
	suite.Nil(unmarshaled.RowValidator)
}

// TestImportRequestJSONMarshaling tests JSON marshaling/unmarshaling for ImportRequest
func (suite *CSVTypesTestSuite) TestImportRequestJSONMarshaling() {
	options := ParseOptions{
		Format:          "standard",
		ContinueOnError: true,
		SkipEmptyRows:   true,
		DateFormat:      "2006-01-02",
	}

	request := ImportRequest{
		// Reader is interface{} and marked with json:"-", so it won't be marshaled
		Options:     options,
		ClientID:    "client-123",
		InvoiceID:   "invoice-456",
		DryRun:      true,
		Interactive: false,
	}

	jsonData, err := json.Marshal(request)
	suite.Require().NoError(err)

	var unmarshaled ImportRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	suite.Require().NoError(err)

	suite.Equal(request.Options, unmarshaled.Options)
	suite.Equal(request.ClientID, unmarshaled.ClientID)
	suite.Equal(request.InvoiceID, unmarshaled.InvoiceID)
	suite.Equal(request.DryRun, unmarshaled.DryRun)
	suite.Equal(request.Interactive, unmarshaled.Interactive)
	suite.Nil(unmarshaled.Reader) // Should be nil since it's not marshaled
}

// TestImportResultJSONMarshaling tests JSON marshaling/unmarshaling for ImportResult
func (suite *CSVTypesTestSuite) TestImportResultJSONMarshaling() {
	parseResult := &ParseResult{
		TotalRows:   5,
		SuccessRows: 4,
		ErrorRows:   1,
		Format:      "standard",
	}

	warnings := []ImportWarning{
		{
			Type:    "duplicate",
			Message: "Potential duplicate work item found",
			Line:    3,
		},
		{
			Type:    "date_range",
			Message: "Work item date is in the future",
			Line:    5,
		},
	}

	result := ImportResult{
		ParseResult:    parseResult,
		InvoiceID:      "invoice-789",
		WorkItemsAdded: 4,
		TotalAmount:    1250.75,
		Warnings:       warnings,
		DryRun:         false,
	}

	jsonData, err := json.Marshal(result)
	suite.Require().NoError(err)

	var unmarshaled ImportResult
	err = json.Unmarshal(jsonData, &unmarshaled)
	suite.Require().NoError(err)

	suite.Equal(result.InvoiceID, unmarshaled.InvoiceID)
	suite.Equal(result.WorkItemsAdded, unmarshaled.WorkItemsAdded)
	suite.InEpsilon(result.TotalAmount, unmarshaled.TotalAmount, 1e-9)
	suite.Equal(result.DryRun, unmarshaled.DryRun)
	suite.Len(unmarshaled.Warnings, 2)
	suite.Equal(warnings[0].Type, unmarshaled.Warnings[0].Type)
	suite.Equal(warnings[1].Message, unmarshaled.Warnings[1].Message)
}

// TestImportWarningJSONMarshaling tests JSON marshaling/unmarshaling for ImportWarning
func (suite *CSVTypesTestSuite) TestImportWarningJSONMarshaling() {
	tests := []struct {
		name    string
		warning ImportWarning
	}{
		{
			name: "DuplicateWarning",
			warning: ImportWarning{
				Type:    "duplicate",
				Message: "Duplicate work item detected",
				Line:    10,
			},
		},
		{
			name: "DateRangeWarning",
			warning: ImportWarning{
				Type:    "date_range",
				Message: "Date is outside expected range",
				Line:    15,
			},
		},
		{
			name: "GeneralWarning",
			warning: ImportWarning{
				Type:    "general",
				Message: "General warning message",
				Line:    0, // No specific line
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			jsonData, err := json.Marshal(tt.warning)
			suite.Require().NoError(err)

			var unmarshaled ImportWarning
			err = json.Unmarshal(jsonData, &unmarshaled)
			suite.Require().NoError(err)

			suite.Equal(tt.warning, unmarshaled)
		})
	}
}

// TestValidateImportRequestJSONMarshaling tests JSON marshaling/unmarshaling for ValidateImportRequest
func (suite *CSVTypesTestSuite) TestValidateImportRequestJSONMarshaling() {
	options := ParseOptions{
		Format:        "excel",
		SkipEmptyRows: true,
		DateFormat:    "01/02/2006",
	}

	request := ValidateImportRequest{
		// Reader is interface{} and marked with json:"-"
		Options: options,
	}

	jsonData, err := json.Marshal(request)
	suite.Require().NoError(err)

	var unmarshaled ValidateImportRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	suite.Require().NoError(err)

	suite.Equal(request.Options, unmarshaled.Options)
	suite.Nil(unmarshaled.Reader)
}

// TestValidationResultJSONMarshaling tests JSON marshaling/unmarshaling for ValidationResult
func (suite *CSVTypesTestSuite) TestValidationResultJSONMarshaling() {
	parseResult := &ParseResult{
		TotalRows:   10,
		SuccessRows: 8,
		ErrorRows:   2,
		Format:      "standard",
	}

	warnings := []ImportWarning{
		{Type: "duplicate", Message: "Duplicate found", Line: 5},
	}

	result := ValidationResult{
		Valid:          false,
		ParseResult:    parseResult,
		Warnings:       warnings,
		Suggestions:    []string{"Fix date formats", "Remove duplicate entries"},
		EstimatedTotal: 2575.50,
	}

	jsonData, err := json.Marshal(result)
	suite.Require().NoError(err)

	var unmarshaled ValidationResult
	err = json.Unmarshal(jsonData, &unmarshaled)
	suite.Require().NoError(err)

	suite.Equal(result.Valid, unmarshaled.Valid)
	suite.InEpsilon(result.EstimatedTotal, unmarshaled.EstimatedTotal, 1e-9)
	suite.Len(unmarshaled.Suggestions, 2)
	suite.Equal(result.Suggestions[0], unmarshaled.Suggestions[0])
}

// TestBatchImportRequestJSONMarshaling tests JSON marshaling/unmarshaling for BatchImportRequest
func (suite *CSVTypesTestSuite) TestBatchImportRequestJSONMarshaling() {
	requests := []ImportRequest{
		{
			Options:   ParseOptions{Format: "standard"},
			ClientID:  "client-1",
			InvoiceID: "invoice-1",
		},
		{
			Options:   ParseOptions{Format: "excel"},
			ClientID:  "client-2",
			InvoiceID: "invoice-2",
		},
	}

	batchOptions := BatchOptions{
		ContinueOnError: true,
		MaxConcurrency:  4,
		ProgressReport:  true,
	}

	batchRequest := BatchImportRequest{
		Requests: requests,
		Options:  batchOptions,
	}

	jsonData, err := json.Marshal(batchRequest)
	suite.Require().NoError(err)

	var unmarshaled BatchImportRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	suite.Require().NoError(err)

	suite.Len(unmarshaled.Requests, 2)
	suite.Equal(batchRequest.Options, unmarshaled.Options)
}

// TestBatchOptionsJSONMarshaling tests JSON marshaling/unmarshaling for BatchOptions
func (suite *CSVTypesTestSuite) TestBatchOptionsJSONMarshaling() {
	tests := []struct {
		name    string
		options BatchOptions
	}{
		{
			name: "DefaultBatchOptions",
			options: BatchOptions{
				ContinueOnError: true,
				MaxConcurrency:  2,
				ProgressReport:  false,
			},
		},
		{
			name: "HighConcurrencyOptions",
			options: BatchOptions{
				ContinueOnError: false,
				MaxConcurrency:  10,
				ProgressReport:  true,
			},
		},
		{
			name: "SingleThreadOptions",
			options: BatchOptions{
				ContinueOnError: true,
				MaxConcurrency:  1,
				ProgressReport:  true,
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			jsonData, err := json.Marshal(tt.options)
			suite.Require().NoError(err)

			var unmarshaled BatchOptions
			err = json.Unmarshal(jsonData, &unmarshaled)
			suite.Require().NoError(err)

			suite.Equal(tt.options, unmarshaled)
		})
	}
}

// TestBatchResultJSONMarshaling tests JSON marshaling/unmarshaling for BatchResult
func (suite *CSVTypesTestSuite) TestBatchResultJSONMarshaling() {
	results := []ImportResult{
		{
			InvoiceID:      "invoice-1",
			WorkItemsAdded: 5,
			TotalAmount:    500.00,
			DryRun:         false,
		},
		{
			InvoiceID:      "invoice-2",
			WorkItemsAdded: 3,
			TotalAmount:    375.50,
			DryRun:         false,
		},
	}

	batchResult := BatchResult{
		TotalRequests:   2,
		SuccessRequests: 2,
		FailedRequests:  0,
		Results:         results,
		TotalWorkItems:  8,
		TotalAmount:     875.50,
	}

	jsonData, err := json.Marshal(batchResult)
	suite.Require().NoError(err)

	var unmarshaled BatchResult
	err = json.Unmarshal(jsonData, &unmarshaled)
	suite.Require().NoError(err)

	suite.Equal(batchResult.TotalRequests, unmarshaled.TotalRequests)
	suite.Equal(batchResult.SuccessRequests, unmarshaled.SuccessRequests)
	suite.Equal(batchResult.FailedRequests, unmarshaled.FailedRequests)
	suite.Equal(batchResult.TotalWorkItems, unmarshaled.TotalWorkItems)
	suite.InEpsilon(batchResult.TotalAmount, unmarshaled.TotalAmount, 1e-9)
	suite.Len(unmarshaled.Results, 2)
}

// TestProgressReportJSONMarshaling tests JSON marshaling/unmarshaling for ProgressReport
func (suite *CSVTypesTestSuite) TestProgressReportJSONMarshaling() {
	tests := []struct {
		name     string
		progress ProgressReport
	}{
		{
			name: "StartingProgress",
			progress: ProgressReport{
				Operation:     "Parsing CSV file",
				TotalRows:     100,
				ProcessedRows: 0,
				SuccessRows:   0,
				ErrorRows:     0,
				Percentage:    0.0,
			},
		},
		{
			name: "MidProgress",
			progress: ProgressReport{
				Operation:     "Processing work items",
				TotalRows:     100,
				ProcessedRows: 50,
				SuccessRows:   48,
				ErrorRows:     2,
				Percentage:    50.0,
			},
		},
		{
			name: "CompletedProgress",
			progress: ProgressReport{
				Operation:     "Import completed",
				TotalRows:     100,
				ProcessedRows: 100,
				SuccessRows:   95,
				ErrorRows:     5,
				Percentage:    100.0,
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			jsonData, err := json.Marshal(tt.progress)
			suite.Require().NoError(err)

			var unmarshaled ProgressReport
			err = json.Unmarshal(jsonData, &unmarshaled)
			suite.Require().NoError(err)

			suite.Equal(tt.progress, unmarshaled)
		})
	}
}

// TestFormatDetectionResultJSONMarshaling tests JSON marshaling/unmarshaling for FormatDetectionResult
func (suite *CSVTypesTestSuite) TestFormatDetectionResultJSONMarshaling() {
	primaryFormat := &FormatInfo{
		Name:      "standard",
		Delimiter: ',',
		HasHeader: true,
		Encoding:  "UTF-8",
	}

	alternateFormats := []*FormatInfo{
		{
			Name:      "excel",
			Delimiter: ';',
			HasHeader: true,
			Encoding:  "UTF-8",
		},
		{
			Name:      "tab",
			Delimiter: '\t',
			HasHeader: false,
			Encoding:  "UTF-8",
		},
	}

	detection := FormatDetectionResult{
		PrimaryFormat:    primaryFormat,
		AlternateFormats: alternateFormats,
		Confidence:       0.85,
		Suggestions:      []string{"Consider using standard CSV format", "Check for missing headers"},
	}

	jsonData, err := json.Marshal(detection)
	suite.Require().NoError(err)

	var unmarshaled FormatDetectionResult
	err = json.Unmarshal(jsonData, &unmarshaled)
	suite.Require().NoError(err)

	suite.Equal(detection.PrimaryFormat.Name, unmarshaled.PrimaryFormat.Name)
	suite.Equal(detection.PrimaryFormat.Delimiter, unmarshaled.PrimaryFormat.Delimiter)
	suite.InEpsilon(detection.Confidence, unmarshaled.Confidence, 1e-9)
	suite.Len(unmarshaled.AlternateFormats, 2)
	suite.Len(unmarshaled.Suggestions, 2)
}

// TestDuplicateWorkItemJSONMarshaling tests JSON marshaling/unmarshaling for DuplicateWorkItem
func (suite *CSVTypesTestSuite) TestDuplicateWorkItemJSONMarshaling() {
	importedItem := &models.WorkItem{
		ID:          "imported-1",
		Date:        time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Hours:       8.0,
		Rate:        75.0,
		Description: "Software development",
		Total:       600.0,
	}

	existingItem := &models.WorkItem{
		ID:          "existing-1",
		Date:        time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Hours:       8.0,
		Rate:        75.0,
		Description: "Software development work",
		Total:       600.0,
	}

	duplicate := DuplicateWorkItem{
		ImportedItem: importedItem,
		ExistingItem: existingItem,
		MatchScore:   0.95,
		MatchReasons: []string{"Same date", "Same hours", "Similar description"},
	}

	jsonData, err := json.Marshal(duplicate)
	suite.Require().NoError(err)

	var unmarshaled DuplicateWorkItem
	err = json.Unmarshal(jsonData, &unmarshaled)
	suite.Require().NoError(err)

	suite.InEpsilon(duplicate.MatchScore, unmarshaled.MatchScore, 1e-9)
	suite.Equal(duplicate.MatchReasons, unmarshaled.MatchReasons)
	suite.Equal(duplicate.ImportedItem.ID, unmarshaled.ImportedItem.ID)
	suite.Equal(duplicate.ExistingItem.ID, unmarshaled.ExistingItem.ID)
}

// TestDuplicateDetectionResultJSONMarshaling tests JSON marshaling/unmarshaling for DuplicateDetectionResult
func (suite *CSVTypesTestSuite) TestDuplicateDetectionResultJSONMarshaling() {
	duplicates := []DuplicateWorkItem{
		{
			ImportedItem: &models.WorkItem{ID: "imported-1"},
			ExistingItem: &models.WorkItem{ID: "existing-1"},
			MatchScore:   0.90,
			MatchReasons: []string{"Same date", "Same hours"},
		},
	}

	result := DuplicateDetectionResult{
		PotentialDuplicates: duplicates,
		TotalChecked:        50,
		DuplicatesFound:     1,
		Confidence:          0.75,
	}

	jsonData, err := json.Marshal(result)
	suite.Require().NoError(err)

	var unmarshaled DuplicateDetectionResult
	err = json.Unmarshal(jsonData, &unmarshaled)
	suite.Require().NoError(err)

	suite.Equal(result.TotalChecked, unmarshaled.TotalChecked)
	suite.Equal(result.DuplicatesFound, unmarshaled.DuplicatesFound)
	suite.InEpsilon(result.Confidence, unmarshaled.Confidence, 1e-9)
	suite.Len(unmarshaled.PotentialDuplicates, 1)
}

// TestAggregationRulesJSONMarshaling tests JSON marshaling/unmarshaling for AggregationRules
func (suite *CSVTypesTestSuite) TestAggregationRulesJSONMarshaling() {
	tests := []struct {
		name  string
		rules AggregationRules
	}{
		{
			name: "DefaultAggregationRules",
			rules: AggregationRules{
				GroupByDate:        true,
				GroupByDescription: false,
				MergeThreshold:     0.25,
				RoundingPrecision:  2,
			},
		},
		{
			name: "StrictAggregationRules",
			rules: AggregationRules{
				GroupByDate:        true,
				GroupByDescription: true,
				MergeThreshold:     1.0,
				RoundingPrecision:  4,
			},
		},
		{
			name: "NoAggregationRules",
			rules: AggregationRules{
				GroupByDate:        false,
				GroupByDescription: false,
				MergeThreshold:     0.0,
				RoundingPrecision:  2,
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			jsonData, err := json.Marshal(tt.rules)
			suite.Require().NoError(err)

			var unmarshaled AggregationRules
			err = json.Unmarshal(jsonData, &unmarshaled)
			suite.Require().NoError(err)

			suite.Equal(tt.rules, unmarshaled)
		})
	}
}

// TestAggregationResultJSONMarshaling tests JSON marshaling/unmarshaling for AggregationResult
func (suite *CSVTypesTestSuite) TestAggregationResultJSONMarshaling() {
	originalItems := []models.WorkItem{
		{ID: "1", Hours: 4.0, Rate: 75.0, Total: 300.0},
		{ID: "2", Hours: 4.0, Rate: 75.0, Total: 300.0},
	}

	aggregatedItems := []models.WorkItem{
		{ID: "aggregated-1", Hours: 8.0, Rate: 75.0, Total: 600.0},
	}

	result := AggregationResult{
		OriginalItems:   originalItems,
		AggregatedItems: aggregatedItems,
		ReductionCount:  1,
		TotalHours:      8.0,
		TotalAmount:     600.0,
	}

	jsonData, err := json.Marshal(result)
	suite.Require().NoError(err)

	var unmarshaled AggregationResult
	err = json.Unmarshal(jsonData, &unmarshaled)
	suite.Require().NoError(err)

	suite.Equal(result.ReductionCount, unmarshaled.ReductionCount)
	suite.InEpsilon(result.TotalHours, unmarshaled.TotalHours, 1e-9)
	suite.InEpsilon(result.TotalAmount, unmarshaled.TotalAmount, 1e-9)
	suite.Len(unmarshaled.OriginalItems, 2)
	suite.Len(unmarshaled.AggregatedItems, 1)
}

// TestCSVValidatorInterface tests that CSVValidator interface requirements are met
func (suite *CSVTypesTestSuite) TestCSVValidatorInterface() {
	// This is a compile-time test - if this compiles, the interface is correctly defined
	var _ CSVValidator = (*mockCSVValidator)(nil)
}

// mockCSVValidator implements CSVValidator for testing
type mockCSVValidator struct{}

func (m *mockCSVValidator) ValidateWorkItem(ctx context.Context, item *models.WorkItem) error {
	return nil
}

func (m *mockCSVValidator) ValidateRow(ctx context.Context, row []string, lineNum int) error {
	return nil
}

func (m *mockCSVValidator) ValidateBatch(ctx context.Context, items []models.WorkItem) error {
	return nil
}

// TestJSONFieldNames tests that JSON field names match expectations for CSV types
func (suite *CSVTypesTestSuite) TestJSONFieldNames() {
	parseOptions := ParseOptions{
		Format:          "standard",
		ContinueOnError: true,
		SkipEmptyRows:   true,
		DateFormat:      "2006-01-02",
	}

	jsonData, err := json.Marshal(parseOptions)
	suite.Require().NoError(err)
	jsonStr := string(jsonData)

	// Verify JSON field names match struct tags
	suite.Contains(jsonStr, `"format":`)
	suite.Contains(jsonStr, `"continue_on_error":`)
	suite.Contains(jsonStr, `"skip_empty_rows":`)
	suite.Contains(jsonStr, `"date_format":`)

	// Test other important types
	parseError := ParseError{
		Line:       1,
		Column:     "test",
		Value:      "val",
		Message:    "msg",
		Suggestion: "suggestion",
		Row:        []string{"row"},
	}

	jsonData, err = json.Marshal(parseError)
	suite.Require().NoError(err)
	jsonStr = string(jsonData)

	suite.Contains(jsonStr, `"line":`)
	suite.Contains(jsonStr, `"column":`)
	suite.Contains(jsonStr, `"value":`)
	suite.Contains(jsonStr, `"message":`)
	suite.Contains(jsonStr, `"suggestion":`)
	suite.Contains(jsonStr, `"row":`)
}

// TestEmptyStructSerialization tests serialization of empty/zero-value structs
func (suite *CSVTypesTestSuite) TestEmptyStructSerialization() {
	emptyTypes := []interface{}{
		ParseOptions{},
		ParseResult{},
		ParseError{},
		FormatInfo{},
		ImportRequest{},
		ImportResult{},
		ImportWarning{},
		ValidateImportRequest{},
		ValidationResult{},
		BatchImportRequest{},
		BatchOptions{},
		BatchResult{},
		ProgressReport{},
		FormatDetectionResult{},
		DuplicateWorkItem{},
		DuplicateDetectionResult{},
		AggregationRules{},
		AggregationResult{},
	}

	for i, emptyType := range emptyTypes {
		suite.Run(fmt.Sprintf("EmptyStruct_%d", i), func() {
			jsonData, err := json.Marshal(emptyType)
			suite.Require().NoError(err)

			// Create new instance of same type for unmarshaling
			switch v := emptyType.(type) {
			case ParseOptions:
				var unmarshaled ParseOptions
				err = json.Unmarshal(jsonData, &unmarshaled)
				suite.Require().NoError(err)
				suite.Equal(v, unmarshaled)
			case ParseResult:
				var unmarshaled ParseResult
				err = json.Unmarshal(jsonData, &unmarshaled)
				suite.Require().NoError(err)
				suite.Equal(v, unmarshaled)
			case ParseError:
				var unmarshaled ParseError
				err = json.Unmarshal(jsonData, &unmarshaled)
				suite.Require().NoError(err)
				suite.Equal(v, unmarshaled)
			case FormatInfo:
				var unmarshaled FormatInfo
				err = json.Unmarshal(jsonData, &unmarshaled)
				suite.Require().NoError(err)
				suite.Equal(v, unmarshaled)
			case ImportRequest:
				var unmarshaled ImportRequest
				err = json.Unmarshal(jsonData, &unmarshaled)
				suite.Require().NoError(err)
				// Only compare JSON-serializable fields
				suite.Equal(v.Options, unmarshaled.Options)
				suite.Equal(v.ClientID, unmarshaled.ClientID)
				suite.Equal(v.InvoiceID, unmarshaled.InvoiceID)
				suite.Equal(v.DryRun, unmarshaled.DryRun)
				suite.Equal(v.Interactive, unmarshaled.Interactive)
			default:
				// Generic handling for other types
				suite.NotEmpty(jsonData)
			}
		})
	}
}

// Standalone tests for specific edge cases

// TestRuneJSONSerialization tests that rune fields serialize correctly
func TestRuneJSONSerialization(t *testing.T) {
	formats := []FormatInfo{
		{Name: "comma", Delimiter: ','},
		{Name: "tab", Delimiter: '\t'},
		{Name: "semicolon", Delimiter: ';'},
		{Name: "pipe", Delimiter: '|'},
		{Name: "space", Delimiter: ' '},
	}

	for _, format := range formats {
		t.Run(format.Name, func(t *testing.T) {
			jsonData, err := json.Marshal(format)
			require.NoError(t, err)

			var unmarshaled FormatInfo
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)

			assert.Equal(t, format.Delimiter, unmarshaled.Delimiter)
		})
	}
}

// TestFloatPrecisionInCSVTypes tests floating point precision in CSV types
func TestFloatPrecisionInCSVTypes(t *testing.T) {
	tests := []struct {
		name  string
		value float64
	}{
		{"Zero", 0.0},
		{"Small", 0.001},
		{"Typical", 125.75},
		{"HighPrecision", 123.456789012},
		{"Large", 999999.99},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test in AggregationRules
			rules := AggregationRules{
				MergeThreshold: tt.value,
			}

			jsonData, err := json.Marshal(rules)
			require.NoError(t, err)

			var unmarshaledRules AggregationRules
			err = json.Unmarshal(jsonData, &unmarshaledRules)
			require.NoError(t, err)

			if tt.value == 0.0 {
				assert.InDelta(t, tt.value, unmarshaledRules.MergeThreshold, 1e-9)
			} else {
				assert.InEpsilon(t, tt.value, unmarshaledRules.MergeThreshold, 1e-9)
			}

			// Test in AggregationResult
			result := AggregationResult{
				TotalHours:  tt.value,
				TotalAmount: tt.value * 75.0, // Some rate
			}

			jsonData, err = json.Marshal(result)
			require.NoError(t, err)

			var unmarshaledResult AggregationResult
			err = json.Unmarshal(jsonData, &unmarshaledResult)
			require.NoError(t, err)

			if tt.value == 0.0 {
				assert.InDelta(t, tt.value, unmarshaledResult.TotalHours, 1e-9)
				assert.InDelta(t, tt.value*75.0, unmarshaledResult.TotalAmount, 1e-9)
			} else {
				assert.InEpsilon(t, tt.value, unmarshaledResult.TotalHours, 1e-9)
				assert.InEpsilon(t, tt.value*75.0, unmarshaledResult.TotalAmount, 1e-9)
			}
		})
	}
}

// TestSliceHandling tests handling of various slice types in CSV types
func TestSliceHandling(t *testing.T) {
	// Test empty slices
	parseResult := ParseResult{
		WorkItems: []models.WorkItem{},
		Errors:    []ParseError{},
		HeaderMap: map[string]int{},
	}

	jsonData, err := json.Marshal(parseResult)
	require.NoError(t, err)

	var unmarshaled ParseResult
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Empty(t, unmarshaled.WorkItems)
	assert.Empty(t, unmarshaled.Errors)
	assert.Empty(t, unmarshaled.HeaderMap)

	// Test nil slices
	batchResult := BatchResult{
		Results: nil, // nil slice
	}

	jsonData, err = json.Marshal(batchResult)
	require.NoError(t, err)

	var unmarshaledBatch BatchResult
	err = json.Unmarshal(jsonData, &unmarshaledBatch)
	require.NoError(t, err)

	assert.Nil(t, unmarshaledBatch.Results)
}

// TestMapHandling tests map field JSON serialization
func TestMapHandling(t *testing.T) {
	headerMap := map[string]int{
		"Date":        0,
		"Hours":       1,
		"Rate":        2,
		"Description": 3,
		"Client":      4,
	}

	parseResult := ParseResult{
		HeaderMap: headerMap,
	}

	jsonData, err := json.Marshal(parseResult)
	require.NoError(t, err)

	var unmarshaled ParseResult
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, headerMap, unmarshaled.HeaderMap)

	// Test empty map
	emptyResult := ParseResult{
		HeaderMap: map[string]int{},
	}

	jsonData, err = json.Marshal(emptyResult)
	require.NoError(t, err)

	var unmarshaledEmpty ParseResult
	err = json.Unmarshal(jsonData, &unmarshaledEmpty)
	require.NoError(t, err)

	assert.Empty(t, unmarshaledEmpty.HeaderMap)
}
