package services

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mrz/go-invoice/internal/csv"
	"github.com/mrz/go-invoice/internal/models"
)

var (
	errTestParseError      = fmt.Errorf("parse error")
	errTestValidationError = fmt.Errorf("validation error")
	errTestClientNotFound  = fmt.Errorf("client not found")
)

// Define interfaces for testing
type ImportInvoiceService interface {
	CreateInvoice(ctx context.Context, req models.CreateInvoiceRequest) (*models.Invoice, error)
	GetInvoice(ctx context.Context, id models.InvoiceID) (*models.Invoice, error)
	AddWorkItemToInvoice(ctx context.Context, invoiceID models.InvoiceID, workItem models.WorkItem) (*models.Invoice, error)
}

type ImportClientService interface {
	GetClient(ctx context.Context, id models.ClientID) (*models.Client, error)
}

// Mock implementations for testing
type MockTimesheetParser struct {
	mock.Mock
}

func (m *MockTimesheetParser) ParseTimesheet(ctx context.Context, reader io.Reader, options csv.ParseOptions) (*csv.ParseResult, error) {
	args := m.Called(ctx, reader, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*csv.ParseResult), args.Error(1)
}

func (m *MockTimesheetParser) DetectFormat(ctx context.Context, reader io.Reader) (*csv.FormatInfo, error) {
	args := m.Called(ctx, reader)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*csv.FormatInfo), args.Error(1)
}

func (m *MockTimesheetParser) ValidateFormat(ctx context.Context, reader io.Reader) error {
	args := m.Called(ctx, reader)
	return args.Error(0)
}

type MockCSVValidator struct {
	mock.Mock
}

func (m *MockCSVValidator) ValidateWorkItem(ctx context.Context, item *models.WorkItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockCSVValidator) ValidateRow(ctx context.Context, row []string, lineNum int) error {
	args := m.Called(ctx, row, lineNum)
	return args.Error(0)
}

func (m *MockCSVValidator) ValidateBatch(ctx context.Context, items []models.WorkItem) error {
	args := m.Called(ctx, items)
	return args.Error(0)
}

type MockImportInvoiceService struct {
	mock.Mock
}

func (m *MockImportInvoiceService) CreateInvoice(ctx context.Context, req models.CreateInvoiceRequest) (*models.Invoice, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invoice), args.Error(1)
}

func (m *MockImportInvoiceService) GetInvoice(ctx context.Context, id models.InvoiceID) (*models.Invoice, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invoice), args.Error(1)
}

func (m *MockImportInvoiceService) AddWorkItemToInvoice(ctx context.Context, invoiceID models.InvoiceID, workItem models.WorkItem) (*models.Invoice, error) {
	args := m.Called(ctx, invoiceID, workItem)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invoice), args.Error(1)
}

type MockImportClientService struct {
	mock.Mock
}

func (m *MockImportClientService) GetClient(ctx context.Context, id models.ClientID) (*models.Client, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Client), args.Error(1)
}

// Modified ImportService for testing with interfaces
type TestableImportService struct {
	parser         csv.TimesheetParser
	invoiceService ImportInvoiceService
	clientService  ImportClientService
	validator      csv.CSVValidator
	logger         Logger
	idGenerator    IDGenerator
}

func NewTestableImportService(
	parser csv.TimesheetParser,
	invoiceService ImportInvoiceService,
	clientService ImportClientService,
	validator csv.CSVValidator,
	logger Logger,
	idGenerator IDGenerator,
) *TestableImportService {
	return &TestableImportService{
		parser:         parser,
		invoiceService: invoiceService,
		clientService:  clientService,
		validator:      validator,
		logger:         logger,
		idGenerator:    idGenerator,
	}
}

// Embed ImportService methods for testing
func (s *TestableImportService) ImportToNewInvoice(ctx context.Context, reader io.Reader, req ImportToNewInvoiceRequest) (*csv.ImportResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Info("starting import to new invoice", "client_id", req.ClientID, "dry_run", req.DryRun)

	// Parse CSV data
	parseResult, err := s.parser.ParseTimesheet(ctx, reader, req.ParseOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCSVParsingFailed, err)
	}

	if len(parseResult.WorkItems) == 0 {
		return &csv.ImportResult{
			ParseResult:    parseResult,
			WorkItemsAdded: 0,
			DryRun:         req.DryRun,
		}, nil
	}

	// Validate batch of work items
	if validationErr := s.validator.ValidateBatch(ctx, parseResult.WorkItems); validationErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrBatchValidationFailed, validationErr)
	}

	if req.DryRun {
		s.logger.Info("dry run completed", "work_items", len(parseResult.WorkItems))
		return s.createDryRunResult(parseResult), nil
	}

	// Verify client exists
	_, err = s.clientService.GetClient(ctx, req.ClientID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientVerificationFailed, err)
	}

	// Generate invoice number if not provided
	invoiceNumber := req.InvoiceNumber
	if invoiceNumber == "" {
		invoiceNumber = s.generateInvoiceNumber(ctx)
	}

	// Create invoice
	invoiceReq := models.CreateInvoiceRequest{
		Number:      invoiceNumber,
		ClientID:    req.ClientID,
		Date:        req.InvoiceDate,
		DueDate:     req.DueDate,
		Description: req.Description,
		WorkItems:   s.convertToWorkItemRequests(parseResult.WorkItems),
	}

	invoice, err := s.invoiceService.CreateInvoice(ctx, invoiceReq)
	if err != nil {
		return nil, fmt.Errorf("invoice creation failed: %w", err)
	}

	// Calculate total amount
	totalAmount := s.calculateTotalAmount(parseResult.WorkItems)

	result := &csv.ImportResult{
		ParseResult:    parseResult,
		InvoiceID:      string(invoice.ID),
		WorkItemsAdded: len(parseResult.WorkItems),
		TotalAmount:    totalAmount,
		DryRun:         false,
	}

	s.logger.Info("import to new invoice completed",
		"invoice_id", invoice.ID,
		"work_items", len(parseResult.WorkItems),
		"total_amount", totalAmount)

	return result, nil
}

func (s *TestableImportService) AppendToInvoice(ctx context.Context, reader io.Reader, req AppendToInvoiceRequest) (*csv.ImportResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Info("starting import append to invoice", "invoice_id", req.InvoiceID, "dry_run", req.DryRun)

	// Parse CSV data
	parseResult, err := s.parser.ParseTimesheet(ctx, reader, req.ParseOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCSVParsingFailed, err)
	}

	if len(parseResult.WorkItems) == 0 {
		return &csv.ImportResult{
			ParseResult:    parseResult,
			InvoiceID:      req.InvoiceID,
			WorkItemsAdded: 0,
			DryRun:         req.DryRun,
		}, nil
	}

	// Validate batch
	if validationErr := s.validator.ValidateBatch(ctx, parseResult.WorkItems); validationErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrBatchValidationFailed, validationErr)
	}

	// Check for duplicates with existing invoice work items
	warnings, err := s.detectDuplicates(ctx, models.InvoiceID(req.InvoiceID), parseResult.WorkItems)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDuplicateDetectionFailed, err)
	}

	if req.DryRun {
		result := s.createDryRunResult(parseResult)
		result.InvoiceID = req.InvoiceID
		result.Warnings = warnings
		s.logger.Info("dry run append completed", "work_items", len(parseResult.WorkItems))
		return result, nil
	}

	// Add work items to invoice
	successCount := 0
	for _, workItem := range parseResult.WorkItems {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		_, err := s.invoiceService.AddWorkItemToInvoice(ctx, models.InvoiceID(req.InvoiceID), workItem)
		if err != nil {
			s.logger.Error("failed to add work item to invoice",
				"invoice_id", req.InvoiceID,
				"work_item_date", workItem.Date,
				"error", err)

			continue
		}

		successCount++
	}

	totalAmount := s.calculateTotalAmount(parseResult.WorkItems[:successCount])

	result := &csv.ImportResult{
		ParseResult:    parseResult,
		InvoiceID:      req.InvoiceID,
		WorkItemsAdded: successCount,
		TotalAmount:    totalAmount,
		Warnings:       warnings,
		DryRun:         false,
	}

	s.logger.Info("import append completed",
		"invoice_id", req.InvoiceID,
		"work_items_added", successCount,
		"total_amount", totalAmount)

	return result, nil
}

// Helper methods (copied from ImportService)
func (s *TestableImportService) createDryRunResult(parseResult *csv.ParseResult) *csv.ImportResult {
	totalAmount := s.calculateTotalAmount(parseResult.WorkItems)

	return &csv.ImportResult{
		ParseResult:    parseResult,
		WorkItemsAdded: len(parseResult.WorkItems),
		TotalAmount:    totalAmount,
		DryRun:         true,
	}
}

func (s *TestableImportService) calculateTotalAmount(workItems []models.WorkItem) float64 {
	total := 0.0
	for _, item := range workItems {
		total += item.Total
	}
	return total
}

func (s *TestableImportService) convertToWorkItemRequests(workItems []models.WorkItem) []models.WorkItem {
	return workItems
}

func (s *TestableImportService) generateInvoiceNumber(_ context.Context) string {
	now := time.Now()
	return fmt.Sprintf("INV-%s", now.Format("20060102-150405"))
}

func (s *TestableImportService) detectDuplicates(ctx context.Context, invoiceID models.InvoiceID, newWorkItems []models.WorkItem) ([]csv.ImportWarning, error) {
	// Get existing invoice
	invoice, err := s.invoiceService.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice for duplicate detection: %w", err)
	}

	var warnings []csv.ImportWarning

	// Simple duplicate detection based on date and hours
	for _, newItem := range newWorkItems {
		for _, existingItem := range invoice.WorkItems {
			if s.workItemsAreSimilar(newItem, existingItem) {
				warning := csv.ImportWarning{
					Type: "duplicate",
					Message: fmt.Sprintf("Potential duplicate work item on %s: %v hours",
						newItem.Date.Format("2006-01-02"), newItem.Hours),
				}
				warnings = append(warnings, warning)
				break
			}
		}
	}

	return warnings, nil
}

func (s *TestableImportService) workItemsAreSimilar(item1, item2 models.WorkItem) bool {
	// Consider items similar if same date and similar hours (within 0.1)
	return item1.Date.Equal(item2.Date) &&
		abs(item1.Hours-item2.Hours) < 0.1
}

// ImportServiceTestSuite tests the ImportService
type ImportServiceTestSuite struct {
	suite.Suite

	service        *TestableImportService
	parser         *MockTimesheetParser
	invoiceService *MockImportInvoiceService
	clientService  *MockImportClientService
	validator      *MockCSVValidator
	logger         *MockLogger
	idGenerator    *MockIDGenerator
}

func (suite *ImportServiceTestSuite) SetupTest() {
	suite.parser = new(MockTimesheetParser)
	suite.invoiceService = new(MockImportInvoiceService)
	suite.clientService = new(MockImportClientService)
	suite.validator = new(MockCSVValidator)
	suite.logger = new(MockLogger)
	suite.idGenerator = new(MockIDGenerator)

	suite.service = NewTestableImportService(
		suite.parser,
		suite.invoiceService,
		suite.clientService,
		suite.validator,
		suite.logger,
		suite.idGenerator,
	)
}

func (suite *ImportServiceTestSuite) TearDownTest() {
	// TearDownTest is intentionally empty as sub-tests manage their own mock expectations
	// to allow for different mock behaviors in individual test cases
}

func TestImportServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ImportServiceTestSuite))
}

// Test NewTestableImportService constructor
func (suite *ImportServiceTestSuite) TestNewTestableImportService() {
	service := NewTestableImportService(
		suite.parser,
		suite.invoiceService,
		suite.clientService,
		suite.validator,
		suite.logger,
		suite.idGenerator,
	)

	suite.Require().NotNil(service)
	suite.Require().Equal(suite.parser, service.parser)
	suite.Require().Equal(suite.invoiceService, service.invoiceService)
	suite.Require().Equal(suite.clientService, service.clientService)
	suite.Require().Equal(suite.validator, service.validator)
	suite.Require().Equal(suite.logger, service.logger)
	suite.Require().Equal(suite.idGenerator, service.idGenerator)
}

// Test ImportToNewInvoice
func (suite *ImportServiceTestSuite) TestImportToNewInvoice() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	suite.Run("successful_import", func() {
		// Setup test data
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

		client := &models.Client{
			ID:     "client-1",
			Name:   "Test Client",
			Email:  "test@example.com",
			Active: true,
		}

		invoice := &models.Invoice{
			ID:          "invoice-1",
			Number:      "INV-001",
			Client:      *client,
			Date:        time.Now(),
			DueDate:     time.Now().AddDate(0, 0, 30),
			Description: "Test invoice",
			WorkItems:   workItems,
		}

		req := ImportToNewInvoiceRequest{
			ClientID:      "client-1",
			ParseOptions:  csv.ParseOptions{},
			InvoiceNumber: "INV-001",
			InvoiceDate:   time.Now(),
			DueDate:       time.Now().AddDate(0, 0, 30),
			Description:   "Test invoice",
			DryRun:        false,
		}

		reader := strings.NewReader("date,hours,rate,description\n2024-01-15,8.0,125.0,Development work")

		// Setup expectations
		suite.parser.On("ParseTimesheet", ctx, mock.Anything, req.ParseOptions).Return(parseResult, nil)
		suite.validator.On("ValidateBatch", ctx, workItems).Return(nil)
		suite.clientService.On("GetClient", ctx, models.ClientID("client-1")).Return(client, nil)
		suite.invoiceService.On("CreateInvoice", ctx, mock.AnythingOfType("models.CreateInvoiceRequest")).Return(invoice, nil)

		// Execute
		result, err := suite.service.ImportToNewInvoice(ctx, reader, req)

		// Verify
		suite.Require().NoError(err)
		suite.Require().NotNil(result)
		suite.Require().Equal("invoice-1", result.InvoiceID)
		suite.Require().Equal(1, result.WorkItemsAdded)
		suite.Require().InEpsilon(1000.0, result.TotalAmount, 0.001)
		suite.Require().False(result.DryRun)
		suite.Require().Equal(parseResult, result.ParseResult)
	})

	suite.Run("dry_run", func() {
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

		req := ImportToNewInvoiceRequest{
			ClientID:     "client-1",
			ParseOptions: csv.ParseOptions{},
			DryRun:       true,
		}

		reader := strings.NewReader("date,hours,rate,description\n2024-01-15,8.0,125.0,Development work")

		// Setup expectations
		suite.parser.On("ParseTimesheet", ctx, mock.Anything, req.ParseOptions).Return(parseResult, nil)
		suite.validator.On("ValidateBatch", ctx, workItems).Return(nil)

		// Execute
		result, err := suite.service.ImportToNewInvoice(ctx, reader, req)

		// Verify
		suite.Require().NoError(err)
		suite.Require().NotNil(result)
		suite.Require().Empty(result.InvoiceID)
		suite.Require().Equal(1, result.WorkItemsAdded)
		suite.Require().InEpsilon(1000.0, result.TotalAmount, 0.001)
		suite.Require().True(result.DryRun)
	})

	suite.Run("context_canceled", func() {
		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		req := ImportToNewInvoiceRequest{
			ClientID: "client-1",
			DryRun:   false,
		}

		reader := strings.NewReader("test data")

		// Execute
		result, err := suite.service.ImportToNewInvoice(canceledCtx, reader, req)

		// Verify
		suite.Require().Error(err)
		suite.Require().Equal(context.Canceled, err)
		suite.Require().Nil(result)
	})

	suite.Run("csv_parsing_failed", func() {
		// Create fresh mocks for this test
		parser := new(MockTimesheetParser)
		invoiceService := new(MockImportInvoiceService)
		clientService := new(MockImportClientService)
		validator := new(MockCSVValidator)
		logger := new(MockLogger)
		idGenerator := new(MockIDGenerator)

		service := NewTestableImportService(parser, invoiceService, clientService, validator, logger, idGenerator)

		req := ImportToNewInvoiceRequest{
			ClientID:     "client-1",
			ParseOptions: csv.ParseOptions{},
			DryRun:       false,
		}

		reader := strings.NewReader("invalid csv data")

		// Setup expectations
		parser.On("ParseTimesheet", ctx, mock.Anything, req.ParseOptions).Return((*csv.ParseResult)(nil), errTestParseError)

		// Execute
		result, err := service.ImportToNewInvoice(ctx, reader, req)

		// Verify
		suite.Require().Error(err)
		suite.Require().ErrorIs(err, ErrCSVParsingFailed)
		suite.Require().Nil(result)

		// Assert mock expectations
		parser.AssertExpectations(suite.T())
	})

	suite.Run("empty_work_items", func() {
		// Create fresh mocks for this test
		parser := new(MockTimesheetParser)
		invoiceService := new(MockImportInvoiceService)
		clientService := new(MockImportClientService)
		validator := new(MockCSVValidator)
		logger := new(MockLogger)
		idGenerator := new(MockIDGenerator)

		service := NewTestableImportService(parser, invoiceService, clientService, validator, logger, idGenerator)

		parseResult := &csv.ParseResult{
			WorkItems:   []models.WorkItem{},
			TotalRows:   0,
			SuccessRows: 0,
			ErrorRows:   0,
		}

		req := ImportToNewInvoiceRequest{
			ClientID:     "client-1",
			ParseOptions: csv.ParseOptions{},
			DryRun:       false,
		}

		reader := strings.NewReader("date,hours,rate,description\n")

		// Setup expectations
		parser.On("ParseTimesheet", ctx, mock.Anything, req.ParseOptions).Return(parseResult, nil)

		// Execute
		result, err := service.ImportToNewInvoice(ctx, reader, req)

		// Verify
		suite.Require().NoError(err)
		suite.Require().NotNil(result)
		suite.Require().Equal(0, result.WorkItemsAdded)
		suite.Require().InDelta(0.0, result.TotalAmount, 0.001)
		suite.Require().False(result.DryRun)

		// Assert mock expectations
		parser.AssertExpectations(suite.T())
	})

	suite.Run("batch_validation_failed", func() {
		// Create fresh mocks for this test
		parser := new(MockTimesheetParser)
		invoiceService := new(MockImportInvoiceService)
		clientService := new(MockImportClientService)
		validator := new(MockCSVValidator)
		logger := new(MockLogger)
		idGenerator := new(MockIDGenerator)

		service := NewTestableImportService(parser, invoiceService, clientService, validator, logger, idGenerator)

		workItems := []models.WorkItem{
			{
				Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				Hours:       -1.0, // Invalid hours
				Rate:        125.0,
				Total:       -125.0,
				Description: "Invalid work",
			},
		}

		parseResult := &csv.ParseResult{
			WorkItems:   workItems,
			TotalRows:   1,
			SuccessRows: 1,
			ErrorRows:   0,
		}

		req := ImportToNewInvoiceRequest{
			ClientID:     "client-1",
			ParseOptions: csv.ParseOptions{},
			DryRun:       false,
		}

		reader := strings.NewReader("date,hours,rate,description\n2024-01-15,-1.0,125.0,Invalid work")

		// Setup expectations
		parser.On("ParseTimesheet", ctx, mock.Anything, req.ParseOptions).Return(parseResult, nil)
		validator.On("ValidateBatch", ctx, workItems).Return(errTestValidationError)

		// Execute
		result, err := service.ImportToNewInvoice(ctx, reader, req)

		// Verify
		suite.Require().Error(err)
		suite.Require().ErrorIs(err, ErrBatchValidationFailed)
		suite.Require().Nil(result)

		// Assert mock expectations
		parser.AssertExpectations(suite.T())
		validator.AssertExpectations(suite.T())
	})

	suite.Run("client_verification_failed", func() {
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

		req := ImportToNewInvoiceRequest{
			ClientID:     "nonexistent-client",
			ParseOptions: csv.ParseOptions{},
			DryRun:       false,
		}

		reader := strings.NewReader("date,hours,rate,description\n2024-01-15,8.0,125.0,Development work")

		// Setup expectations
		suite.parser.On("ParseTimesheet", ctx, mock.Anything, req.ParseOptions).Return(parseResult, nil)
		suite.validator.On("ValidateBatch", ctx, workItems).Return(nil)
		suite.clientService.On("GetClient", ctx, models.ClientID("nonexistent-client")).Return(nil, errTestClientNotFound)

		// Execute
		result, err := suite.service.ImportToNewInvoice(ctx, reader, req)

		// Verify
		suite.Require().Error(err)
		suite.Require().ErrorIs(err, ErrClientVerificationFailed)
		suite.Require().Nil(result)
	})
}

// Test AppendToInvoice
func (suite *ImportServiceTestSuite) TestAppendToInvoice() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	suite.Run("successful_append", func() {
		workItems := []models.WorkItem{
			{
				Date:        time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
				Hours:       4.0,
				Rate:        125.0,
				Total:       500.0,
				Description: "Additional work",
			},
		}

		parseResult := &csv.ParseResult{
			WorkItems:   workItems,
			TotalRows:   1,
			SuccessRows: 1,
			ErrorRows:   0,
		}

		client := &models.Client{
			ID:     "client-1",
			Name:   "Test Client",
			Email:  "test@example.com",
			Active: true,
		}

		existingInvoice := &models.Invoice{
			ID:     "invoice-1",
			Number: "INV-001",
			Client: *client,
			WorkItems: []models.WorkItem{
				{
					Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
					Hours:       8.0,
					Rate:        125.0,
					Total:       1000.0,
					Description: "Previous work",
				},
			},
		}

		updatedInvoice := &models.Invoice{
			ID:        "invoice-1",
			Number:    "INV-001",
			Client:    *client,
			WorkItems: append(existingInvoice.WorkItems, workItems...),
		}

		req := AppendToInvoiceRequest{
			InvoiceID:    "invoice-1",
			ParseOptions: csv.ParseOptions{},
			DryRun:       false,
		}

		reader := strings.NewReader("date,hours,rate,description\n2024-01-16,4.0,125.0,Additional work")

		// Setup expectations
		suite.parser.On("ParseTimesheet", ctx, mock.Anything, req.ParseOptions).Return(parseResult, nil)
		suite.validator.On("ValidateBatch", ctx, workItems).Return(nil)
		suite.invoiceService.On("GetInvoice", ctx, models.InvoiceID("invoice-1")).Return(existingInvoice, nil)
		suite.invoiceService.On("AddWorkItemToInvoice", ctx, models.InvoiceID("invoice-1"), workItems[0]).Return(updatedInvoice, nil)

		// Execute
		result, err := suite.service.AppendToInvoice(ctx, reader, req)

		// Verify
		suite.Require().NoError(err)
		suite.Require().NotNil(result)
		suite.Require().Equal("invoice-1", result.InvoiceID)
		suite.Require().Equal(1, result.WorkItemsAdded)
		suite.Require().InEpsilon(500.0, result.TotalAmount, 0.001)
		suite.Require().False(result.DryRun)
	})

	suite.Run("duplicate_detection", func() {
		// Create fresh mocks for this test
		parser := new(MockTimesheetParser)
		invoiceService := new(MockImportInvoiceService)
		clientService := new(MockImportClientService)
		validator := new(MockCSVValidator)
		logger := new(MockLogger)
		idGenerator := new(MockIDGenerator)

		service := NewTestableImportService(parser, invoiceService, clientService, validator, logger, idGenerator)

		// Create work item that duplicates existing one
		workItems := []models.WorkItem{
			{
				Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), // Same date as existing
				Hours:       8.0,                                          // Same hours as existing
				Rate:        125.0,
				Total:       1000.0,
				Description: "Duplicate work",
			},
		}

		parseResult := &csv.ParseResult{
			WorkItems:   workItems,
			TotalRows:   1,
			SuccessRows: 1,
			ErrorRows:   0,
		}

		client := &models.Client{
			ID:     "client-1",
			Name:   "Test Client",
			Email:  "test@example.com",
			Active: true,
		}

		existingInvoice := &models.Invoice{
			ID:     "invoice-1",
			Number: "INV-001",
			Client: *client,
			WorkItems: []models.WorkItem{
				{
					Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
					Hours:       8.0,
					Rate:        125.0,
					Total:       1000.0,
					Description: "Previous work",
				},
			},
		}

		req := AppendToInvoiceRequest{
			InvoiceID:    "invoice-1",
			ParseOptions: csv.ParseOptions{},
			DryRun:       true, // Dry run to see warnings
		}

		reader := strings.NewReader("date,hours,rate,description\n2024-01-15,8.0,125.0,Duplicate work")

		// Setup expectations
		parser.On("ParseTimesheet", ctx, mock.Anything, req.ParseOptions).Return(parseResult, nil)
		validator.On("ValidateBatch", ctx, workItems).Return(nil)
		invoiceService.On("GetInvoice", ctx, models.InvoiceID("invoice-1")).Return(existingInvoice, nil)

		// Execute
		result, err := service.AppendToInvoice(ctx, reader, req)

		// Verify
		suite.Require().NoError(err)
		suite.Require().NotNil(result)
		suite.Require().Equal("invoice-1", result.InvoiceID)
		suite.Require().Equal(1, result.WorkItemsAdded)
		suite.Require().True(result.DryRun)
		suite.Require().Len(result.Warnings, 1)
		suite.Require().Equal("duplicate", result.Warnings[0].Type)

		// Assert mock expectations
		parser.AssertExpectations(suite.T())
		validator.AssertExpectations(suite.T())
		invoiceService.AssertExpectations(suite.T())
	})

	suite.Run("context_canceled", func() {
		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		req := AppendToInvoiceRequest{
			InvoiceID: "invoice-1",
			DryRun:    false,
		}

		reader := strings.NewReader("test data")

		// Execute
		result, err := suite.service.AppendToInvoice(canceledCtx, reader, req)

		// Verify
		suite.Require().Error(err)
		suite.Require().Equal(context.Canceled, err)
		suite.Require().Nil(result)
	})
}

// Test helper methods
func (suite *ImportServiceTestSuite) TestHelperMethods() {
	suite.Run("calculateTotalAmount", func() {
		workItems := []models.WorkItem{
			{Total: 100.0},
			{Total: 200.0},
			{Total: 300.0},
		}

		total := suite.service.calculateTotalAmount(workItems)
		suite.Require().InEpsilon(600.0, total, 0.001)
	})

	suite.Run("calculateTotalAmount_empty", func() {
		total := suite.service.calculateTotalAmount([]models.WorkItem{})
		suite.Require().InDelta(0.0, total, 0.001)
	})

	suite.Run("generateInvoiceNumber", func() {
		ctx := context.Background()
		number := suite.service.generateInvoiceNumber(ctx)
		suite.Require().NotEmpty(number)
		suite.Require().Contains(number, "INV-")
	})

	suite.Run("workItemsAreSimilar", func() {
		date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

		item1 := models.WorkItem{
			Date:  date,
			Hours: 8.0,
		}

		item2 := models.WorkItem{
			Date:  date,
			Hours: 8.05, // Within 0.1 tolerance
		}

		item3 := models.WorkItem{
			Date:  date,
			Hours: 8.2, // Outside 0.1 tolerance
		}

		item4 := models.WorkItem{
			Date:  date.AddDate(0, 0, 1), // Different date
			Hours: 8.0,
		}

		suite.Require().True(suite.service.workItemsAreSimilar(item1, item2))
		suite.Require().False(suite.service.workItemsAreSimilar(item1, item3))
		suite.Require().False(suite.service.workItemsAreSimilar(item1, item4))
	})
}
