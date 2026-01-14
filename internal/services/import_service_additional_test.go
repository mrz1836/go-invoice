package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mrz/go-invoice/internal/csv"
	"github.com/mrz/go-invoice/internal/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// AdditionalImportServiceTestSuite provides additional test coverage for ImportService
type AdditionalImportServiceTestSuite struct {
	suite.Suite
}

func TestAdditionalImportServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AdditionalImportServiceTestSuite))
}

// Test NewImportService constructor
func (suite *AdditionalImportServiceTestSuite) TestNewImportService() {
	parser := new(MockTimesheetParser)
	invoiceService := NewInvoiceService(nil, nil, nil, nil)
	clientService := NewClientService(nil, nil, nil, nil)
	validator := new(MockCSVValidator)
	logger := new(MockLogger)
	idGenerator := new(MockIDGenerator)

	service := NewImportService(parser, invoiceService, clientService, validator, logger, idGenerator)

	suite.Require().NotNil(service)
}

// Test ValidateImport method
func (suite *AdditionalImportServiceTestSuite) TestValidateImport() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	suite.Run("successful_validation", func() {
		parser := new(MockTimesheetParser)
		validator := new(MockCSVValidator)
		logger := new(MockLogger)
		idGenerator := new(MockIDGenerator)

		workItems := []models.WorkItem{
			{
				Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				Hours:       8.0,
				Rate:        125.0,
				Total:       1000.0,
				Description: "Development work",
			},
		}

		parseResult := &csv.ParseResult{
			WorkItems:   workItems,
			TotalRows:   1,
			SuccessRows: 1,
			ErrorRows:   0,
		}

		req := csv.ValidateImportRequest{
			Options: csv.ParseOptions{},
		}

		reader := strings.NewReader("date,hours,rate,description\n2024-01-15,8.0,125.0,Development work")

		// Setup expectations
		parser.On("ParseTimesheet", ctx, mock.Anything, req.Options).Return(parseResult, nil)
		validator.On("ValidateBatch", ctx, workItems).Return(nil)

		// Test ValidateImport
		importService := NewImportService(parser, nil, nil, validator, logger, idGenerator)
		result, err := importService.ValidateImport(ctx, reader, req)

		// Verify
		suite.Require().NoError(err)
		suite.Require().NotNil(result)
		suite.True(result.Valid)
		suite.Equal(parseResult, result.ParseResult)
		suite.InEpsilon(1000.0, result.EstimatedTotal, 0.001)

		// Assert mock expectations
		parser.AssertExpectations(suite.T())
		validator.AssertExpectations(suite.T())
	})

	suite.Run("validation_with_parse_errors", func() {
		parser := new(MockTimesheetParser)
		validator := new(MockCSVValidator)
		logger := new(MockLogger)
		idGenerator := new(MockIDGenerator)

		parseResult := &csv.ParseResult{
			WorkItems:   []models.WorkItem{},
			TotalRows:   1,
			SuccessRows: 0,
			ErrorRows:   1,
		}

		req := csv.ValidateImportRequest{
			Options: csv.ParseOptions{},
		}

		reader := strings.NewReader("invalid,csv,data")

		// Setup expectations
		parser.On("ParseTimesheet", ctx, mock.Anything, req.Options).Return(parseResult, errTestParseError)

		// Test ValidateImport
		importService := NewImportService(parser, nil, nil, validator, logger, idGenerator)
		result, err := importService.ValidateImport(ctx, reader, req)

		// Verify
		suite.Require().Error(err)
		suite.Require().NotNil(result)
		suite.False(result.Valid)
		suite.Equal(parseResult, result.ParseResult)
		suite.Contains(result.Suggestions, "Check file format and field mappings")

		// Assert mock expectations
		parser.AssertExpectations(suite.T())
	})

	suite.Run("validation_with_batch_errors", func() {
		parser := new(MockTimesheetParser)
		validator := new(MockCSVValidator)
		logger := new(MockLogger)
		idGenerator := new(MockIDGenerator)

		workItems := []models.WorkItem{
			{
				Date:        time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC), // Saturday
				Hours:       12.0,                                         // High hours
				Rate:        125.0,
				Total:       1500.0,
				Description: "Weekend overtime",
			},
		}

		parseResult := &csv.ParseResult{
			WorkItems:   workItems,
			TotalRows:   1,
			SuccessRows: 1,
			ErrorRows:   0,
		}

		req := csv.ValidateImportRequest{
			Options: csv.ParseOptions{},
		}

		reader := strings.NewReader("date,hours,rate,description\n2024-01-13,12.0,125.0,Weekend overtime")

		// Setup expectations
		parser.On("ParseTimesheet", ctx, mock.Anything, req.Options).Return(parseResult, nil)
		validator.On("ValidateBatch", ctx, workItems).Return(errTestValidationError)

		// Test ValidateImport
		importService := NewImportService(parser, nil, nil, validator, logger, idGenerator)
		result, err := importService.ValidateImport(ctx, reader, req)

		// Verify
		suite.Require().NoError(err) // Validation errors don't cause ValidateImport to fail
		suite.Require().NotNil(result)
		suite.False(result.Valid)     // But the result should be invalid
		suite.Len(result.Warnings, 2) // Weekend work and high hours warnings
		suite.Contains(result.Suggestions, "Review work item validation rules")

		// Assert mock expectations
		parser.AssertExpectations(suite.T())
		validator.AssertExpectations(suite.T())
	})

	suite.Run("context_canceled", func() {
		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		req := csv.ValidateImportRequest{
			Options: csv.ParseOptions{},
		}

		reader := strings.NewReader("test data")

		// Test ValidateImport
		importService := NewImportService(nil, nil, nil, nil, nil, nil)
		result, err := importService.ValidateImport(canceledCtx, reader, req)

		// Verify
		suite.Require().Error(err)
		suite.Equal(context.Canceled, err)
		suite.Nil(result)
	})
}

// Test helper functions that have lower coverage
func (suite *AdditionalImportServiceTestSuite) TestHelperFunctions() {
	suite.Run("abs_function", func() {
		// Test positive number
		result := abs(5.5)
		suite.InEpsilon(5.5, result, 0.001)

		// Test negative number
		result = abs(-5.5)
		suite.InEpsilon(5.5, result, 0.001)

		// Test zero
		result = abs(0.0)
		suite.InDelta(0.0, result, 0.001)
	})

	suite.Run("generateInvoiceNumber", func() {
		ctx := context.Background()
		service := NewImportService(nil, nil, nil, nil, nil, nil)

		number := service.generateInvoiceNumber(ctx)
		suite.NotEmpty(number)
		suite.Contains(number, "INV-")

		// Test multiple generations to ensure format consistency
		for i := 0; i < 3; i++ {
			testNumber := service.generateInvoiceNumber(ctx)
			suite.NotEmpty(testNumber)
			suite.Contains(testNumber, "INV-")
			// Each should follow the pattern INV-YYYYMMDD-HHMMSS
			suite.Regexp(`^INV-\d{8}-\d{6}$`, testNumber)
		}
	})

	suite.Run("calculateTotalAmount", func() {
		service := NewImportService(nil, nil, nil, nil, nil, nil)

		// Test with multiple work items
		workItems := []models.WorkItem{
			{Total: 100.0},
			{Total: 200.0},
			{Total: 300.0},
		}
		total := service.calculateTotalAmount(workItems)
		suite.InEpsilon(600.0, total, 0.001)

		// Test with empty slice
		total = service.calculateTotalAmount([]models.WorkItem{})
		suite.InDelta(0.0, total, 0.001)

		// Test with single item
		workItems = []models.WorkItem{{Total: 150.75}}
		total = service.calculateTotalAmount(workItems)
		suite.InEpsilon(150.75, total, 0.001)
	})

	suite.Run("convertToWorkItemRequests", func() {
		service := NewImportService(nil, nil, nil, nil, nil, nil)

		workItems := []models.WorkItem{
			{
				Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				Hours:       8.0,
				Rate:        125.0,
				Total:       1000.0,
				Description: "Development work",
			},
		}

		result := service.convertToWorkItemRequests(workItems)
		suite.Equal(workItems, result)
	})

	suite.Run("createDryRunResult", func() {
		service := NewImportService(nil, nil, nil, nil, nil, nil)

		workItems := []models.WorkItem{
			{Total: 100.0},
			{Total: 200.0},
		}

		parseResult := &csv.ParseResult{
			WorkItems:   workItems,
			TotalRows:   2,
			SuccessRows: 2,
			ErrorRows:   0,
		}

		result := service.createDryRunResult(parseResult)
		suite.Require().NotNil(result)
		suite.Equal(parseResult, result.ParseResult)
		suite.Equal(2, result.WorkItemsAdded)
		suite.InEpsilon(300.0, result.TotalAmount, 0.001)
		suite.True(result.DryRun)
	})

	suite.Run("workItemsAreSimilar", func() {
		service := NewImportService(nil, nil, nil, nil, nil, nil)
		date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

		item1 := models.WorkItem{
			Date:  date,
			Hours: 8.0,
		}

		// Similar item (within tolerance)
		item2 := models.WorkItem{
			Date:  date,
			Hours: 8.05, // Within 0.1 tolerance
		}

		// Different hours (outside tolerance)
		item3 := models.WorkItem{
			Date:  date,
			Hours: 8.2, // Outside 0.1 tolerance
		}

		// Different date
		item4 := models.WorkItem{
			Date:  date.AddDate(0, 0, 1), // Different date
			Hours: 8.0,
		}

		suite.True(service.workItemsAreSimilar(item1, item2))
		suite.False(service.workItemsAreSimilar(item1, item3))
		suite.False(service.workItemsAreSimilar(item1, item4))

		// Edge case: exactly at tolerance boundary
		item5 := models.WorkItem{
			Date:  date,
			Hours: 8.11, // Slightly outside tolerance boundary (0.11 > 0.1)
		}
		suite.False(service.workItemsAreSimilar(item1, item5))

		// Edge case: just within tolerance
		item6 := models.WorkItem{
			Date:  date,
			Hours: 8.09, // Just within tolerance
		}
		suite.True(service.workItemsAreSimilar(item1, item6))
	})

	suite.Run("generateValidationSuggestions", func() {
		service := NewImportService(nil, nil, nil, nil, nil, nil)

		// Test with error rows
		parseResult := &csv.ParseResult{
			WorkItems:   []models.WorkItem{},
			TotalRows:   3,
			SuccessRows: 1,
			ErrorRows:   2,
		}

		suggestions := service.generateValidationSuggestions(parseResult, nil)
		suite.Contains(suggestions, "Check data format in rows with errors")
		suite.Contains(suggestions, "Ensure dates are in YYYY-MM-DD format")
		suite.Contains(suggestions, "Verify numeric fields (hours, rates) contain valid numbers")

		// Test with batch validation error
		parseResult2 := &csv.ParseResult{
			WorkItems:   []models.WorkItem{{Date: time.Now(), Hours: 8.0}},
			TotalRows:   1,
			SuccessRows: 1,
			ErrorRows:   0,
		}

		suggestions2 := service.generateValidationSuggestions(parseResult2, errTestValidationError)
		suite.Contains(suggestions2, "Review work item validation rules")
		suite.Contains(suggestions2, "Check for unusual values (very high hours, extreme rates)")

		// Test with empty result
		parseResult3 := &csv.ParseResult{
			WorkItems:   []models.WorkItem{},
			TotalRows:   0,
			SuccessRows: 0,
			ErrorRows:   0,
		}

		suggestions3 := service.generateValidationSuggestions(parseResult3, nil)
		suite.Contains(suggestions3, "File appears to be empty or header-only")
		suite.Contains(suggestions3, "Ensure CSV contains data rows after header")
	})

	suite.Run("generateValidationWarnings", func() {
		service := NewImportService(nil, nil, nil, nil, nil, nil)

		workItems := []models.WorkItem{
			{
				Date:        time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC), // Saturday
				Hours:       12.0,                                         // High hours
				Rate:        125.0,
				Total:       1500.0,
				Description: "Weekend overtime",
			},
			{
				Date:        time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC), // Sunday
				Hours:       6.0,                                          // Normal hours
				Rate:        125.0,
				Total:       750.0,
				Description: "Sunday work",
			},
			{
				Date:        time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC), // Monday
				Hours:       8.0,                                          // Normal hours
				Rate:        125.0,
				Total:       1000.0,
				Description: "Regular work",
			},
		}

		warnings := service.generateValidationWarnings(workItems)

		// Should have 3 warnings: 2 weekend work + 1 high hours
		suite.Len(warnings, 3)

		// Check weekend warnings
		weekendWarnings := 0
		highHoursWarnings := 0
		for _, warning := range warnings {
			if warning.Type == "weekend_work" {
				weekendWarnings++
			}
			if warning.Type == "high_hours" {
				highHoursWarnings++
			}
		}
		suite.Equal(2, weekendWarnings)   // Saturday and Sunday
		suite.Equal(1, highHoursWarnings) // Only Saturday has >10 hours
	})
}

// Test error paths for ImportToNewInvoice that are harder to test in the main test
func (suite *AdditionalImportServiceTestSuite) TestImportToNewInvoiceErrorPaths() {
	suite.Run("invoice_creation_failed_skipped", func() {
		// Skip complex integration test that requires deep service mocking
		suite.T().Skip("Complex integration test skipped - requires refactoring for proper service mocking")
	})
}

// Test error variables coverage
func (suite *AdditionalImportServiceTestSuite) TestErrorVariables() {
	suite.Run("error_variables_exist", func() {
		// Test that error variables are properly defined
		suite.NotEmpty(ErrCSVParsingFailed.Error())
		suite.NotEmpty(ErrBatchValidationFailed.Error())
		suite.NotEmpty(ErrClientVerificationFailed.Error())
		suite.NotEmpty(ErrDuplicateDetectionFailed.Error())
		suite.NotEmpty(ErrBatchImportFailed.Error())

		// Test error wrapping behavior
		suite.Contains(ErrCSVParsingFailed.Error(), "CSV parsing failed")
		suite.Contains(ErrBatchValidationFailed.Error(), "batch validation failed")
		suite.Contains(ErrClientVerificationFailed.Error(), "client verification failed")
		suite.Contains(ErrDuplicateDetectionFailed.Error(), "duplicate detection failed")
		suite.Contains(ErrBatchImportFailed.Error(), "batch import failed")
	})
}
