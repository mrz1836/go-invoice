package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mrz/go-invoice/internal/cli"
	"github.com/mrz/go-invoice/internal/config"
	"github.com/mrz/go-invoice/internal/csv"
	"github.com/mrz/go-invoice/internal/models"
	"github.com/mrz/go-invoice/internal/services"
	"github.com/mrz/go-invoice/internal/storage/json"
)

// SimpleIDGenerator provides a simple ID generator for tests
type SimpleIDGenerator struct {
	counter int
}

func NewSimpleIDGenerator() *SimpleIDGenerator {
	return &SimpleIDGenerator{counter: 0}
}

func (g *SimpleIDGenerator) GenerateInvoiceID(ctx context.Context) (models.InvoiceID, error) {
	g.counter++
	return models.InvoiceID(fmt.Sprintf("test-inv-%d", g.counter)), nil
}

func (g *SimpleIDGenerator) GenerateClientID(ctx context.Context) (models.ClientID, error) {
	g.counter++
	return models.ClientID(fmt.Sprintf("test-client-%d", g.counter)), nil
}

func (g *SimpleIDGenerator) GenerateWorkItemID(ctx context.Context) (string, error) {
	g.counter++
	return fmt.Sprintf("test-wi-%d", g.counter), nil
}

func (g *SimpleIDGenerator) GenerateID() string {
	g.counter++
	return fmt.Sprintf("test-id-%d", g.counter)
}

// IntegrationTestSuite defines the integration test suite
type IntegrationTestSuite struct {
	suite.Suite
	tempDir        string
	logger         *cli.SimpleLogger
	configService  *config.ConfigService
	storage        *json.JSONStorage
	invoiceService *services.InvoiceService
	clientService  *services.ClientService
	csvParser      *csv.CSVParser
	testConfig     *config.Config
	idGenerator    *SimpleIDGenerator
}

// SetupSuite runs once before all tests in the suite
func (suite *IntegrationTestSuite) SetupSuite() {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "go-invoice-integration-*")
	require.NoError(suite.T(), err)
	suite.tempDir = tempDir

	// Initialize logger
	suite.logger = cli.NewLogger(false)

	// Initialize ID generator
	suite.idGenerator = NewSimpleIDGenerator()

	// Initialize config service and create test configuration
	validator := config.NewSimpleValidator(suite.logger)
	suite.configService = config.NewConfigService(suite.logger, validator)
	suite.testConfig = &config.Config{
		Business: config.BusinessConfig{
			Name:         "Test Integration Business",
			Address:      "123 Integration St\nTestville, TV 12345",
			Email:        "integration@test.com",
			Phone:        "+1-555-INTEG",
			PaymentTerms: "Net 30",
		},
		Invoice: config.InvoiceConfig{
			Prefix:      "INT",
			StartNumber: 1000,
			Currency:    "USD",
			VATRate:     0.10,
		},
		Storage: config.StorageConfig{
			DataDir: suite.tempDir,
		},
	}

	// Initialize storage
	suite.storage = json.NewJSONStorage(suite.tempDir, suite.logger)

	// Create required directories
	os.MkdirAll(filepath.Join(suite.tempDir, "invoices"), 0o755)
	os.MkdirAll(filepath.Join(suite.tempDir, "clients"), 0o755)

	// Initialize services with correct parameters
	suite.invoiceService = services.NewInvoiceService(
		suite.storage,     // InvoiceStorage
		suite.storage,     // ClientStorage (JSONStorage implements both)
		suite.logger,      // Logger
		suite.idGenerator, // IDGenerator
	)
	suite.clientService = services.NewClientService(
		suite.storage,     // ClientStorage
		suite.storage,     // InvoiceStorage
		suite.logger,      // Logger
		suite.idGenerator, // IDGenerator
	)

	// Initialize CSV parser
	csvValidator := csv.NewWorkItemValidator(suite.logger)
	suite.csvParser = csv.NewCSVParser(csvValidator, suite.logger, suite.idGenerator)
}

// TearDownSuite runs once after all tests in the suite
func (suite *IntegrationTestSuite) TearDownSuite() {
	os.RemoveAll(suite.tempDir)
}

// SetupTest runs before each test
func (suite *IntegrationTestSuite) SetupTest() {
	// Clear storage between tests
	ctx := context.Background()

	// List and delete all invoices
	result, _ := suite.storage.ListInvoices(ctx, models.InvoiceFilter{})
	if result != nil {
		for _, invoice := range result.Invoices {
			suite.storage.DeleteInvoice(ctx, invoice.ID)
		}
	}
}

// TestBasicInvoiceCreation tests basic invoice creation functionality
func (suite *IntegrationTestSuite) TestBasicInvoiceCreation() {
	ctx := context.Background()

	// Create a client
	clientReq := models.CreateClientRequest{
		Name:    "Test Client",
		Email:   "test@client.com",
		Address: "123 Client St",
	}

	client, err := suite.clientService.CreateClient(ctx, clientReq)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), client.ID)
	assert.Equal(suite.T(), "Test Client", client.Name)

	// Create invoice request with proper structure
	now := time.Now()
	invoiceReq := models.CreateInvoiceRequest{
		Number:   "INT-001",
		ClientID: client.ID,
		Date:     now,
		DueDate:  now.Add(30 * 24 * time.Hour),
		WorkItems: []models.WorkItem{
			{
				ID:          "wi-001",
				Date:        now,
				Description: "Development Work",
				Hours:       8.0,
				Rate:        100.0,
				Total:       800.0,
				CreatedAt:   now,
			},
		},
	}

	// Create invoice
	invoice, err := suite.invoiceService.CreateInvoice(ctx, invoiceReq)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), invoice.ID)
	assert.Equal(suite.T(), client.ID, invoice.Client.ID)
	assert.Len(suite.T(), invoice.WorkItems, 1)
	assert.Equal(suite.T(), models.StatusDraft, invoice.Status)

	// Verify basic calculations (subtotal should be correct even if tax isn't applied)
	assert.Equal(suite.T(), 800.0, invoice.Subtotal)
	assert.Equal(suite.T(), "INT-001", invoice.Number)
}

// TestCSVParsingWorkflow tests CSV parsing functionality
func (suite *IntegrationTestSuite) TestCSVParsingWorkflow() {
	ctx := context.Background()

	// Create CSV test data
	csvContent := `date,description,hours,rate
2024-01-15,Website Design,8.5,95.00
2024-01-16,Frontend Development,7.25,110.00`

	csvFile := filepath.Join(suite.tempDir, "test_timesheet.csv")
	err := os.WriteFile(csvFile, []byte(csvContent), 0o644)
	require.NoError(suite.T(), err)

	// Open and parse CSV file
	file, err := os.Open(csvFile)
	require.NoError(suite.T(), err)
	defer file.Close()

	// Parse CSV data
	result, err := suite.csvParser.ParseTimesheet(ctx, file, csv.ParseOptions{})
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.WorkItems, 2)

	// Verify parsed data
	workItem1 := result.WorkItems[0]
	assert.Equal(suite.T(), "Website Design", workItem1.Description)
	assert.Equal(suite.T(), 8.5, workItem1.Hours)
	assert.Equal(suite.T(), 95.0, workItem1.Rate)

	workItem2 := result.WorkItems[1]
	assert.Equal(suite.T(), "Frontend Development", workItem2.Description)
	assert.Equal(suite.T(), 7.25, workItem2.Hours)
	assert.Equal(suite.T(), 110.0, workItem2.Rate)
}

// TestInvoiceListingAndFiltering tests invoice listing functionality
func (suite *IntegrationTestSuite) TestInvoiceListingAndFiltering() {
	ctx := context.Background()

	// Create a client
	clientReq := models.CreateClientRequest{
		Name:  "Listing Test Client",
		Email: "listing@test.com",
	}
	client, err := suite.clientService.CreateClient(ctx, clientReq)
	require.NoError(suite.T(), err)

	// Create multiple invoices
	now := time.Now()
	invoice1Req := models.CreateInvoiceRequest{
		Number:   "INT-003",
		ClientID: client.ID,
		Date:     now,
		DueDate:  now.Add(30 * 24 * time.Hour),
		WorkItems: []models.WorkItem{
			{
				ID:          "wi-003",
				Date:        now,
				Description: "Draft Work",
				Hours:       1,
				Rate:        100,
				Total:       100,
				CreatedAt:   now,
			},
		},
	}

	invoice2Req := models.CreateInvoiceRequest{
		Number:   "INT-004",
		ClientID: client.ID,
		Date:     now,
		DueDate:  now.Add(30 * 24 * time.Hour),
		WorkItems: []models.WorkItem{
			{
				ID:          "wi-004",
				Date:        now,
				Description: "Second Work",
				Hours:       2,
				Rate:        100,
				Total:       200,
				CreatedAt:   now,
			},
		},
	}

	// Create both invoices
	invoice1, err := suite.invoiceService.CreateInvoice(ctx, invoice1Req)
	require.NoError(suite.T(), err)

	invoice2, err := suite.invoiceService.CreateInvoice(ctx, invoice2Req)
	require.NoError(suite.T(), err)

	// List all invoices
	result, err := suite.storage.ListInvoices(ctx, models.InvoiceFilter{})
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), result.Invoices, 2)

	// Verify both invoices are present
	invoiceNumbers := []string{result.Invoices[0].Number, result.Invoices[1].Number}
	assert.Contains(suite.T(), invoiceNumbers, "INT-003")
	assert.Contains(suite.T(), invoiceNumbers, "INT-004")

	// Test getting individual invoices
	retrievedInvoice1, err := suite.invoiceService.GetInvoice(ctx, invoice1.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "INT-003", retrievedInvoice1.Number)

	retrievedInvoice2, err := suite.invoiceService.GetInvoice(ctx, invoice2.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "INT-004", retrievedInvoice2.Number)
}

// TestErrorHandling tests various error scenarios
func (suite *IntegrationTestSuite) TestErrorHandling() {
	ctx := context.Background()

	// Test invalid CSV data
	invalidCSV := `date,description,hours,rate
invalid-date,Test Item,8,100`

	csvFile := filepath.Join(suite.tempDir, "invalid_test.csv")
	err := os.WriteFile(csvFile, []byte(invalidCSV), 0o644)
	require.NoError(suite.T(), err)

	file, err := os.Open(csvFile)
	require.NoError(suite.T(), err)
	defer file.Close()

	// This should return an error due to invalid date
	_, err = suite.csvParser.ParseTimesheet(ctx, file, csv.ParseOptions{})
	assert.Error(suite.T(), err)

	// Test getting non-existent invoice
	_, err = suite.invoiceService.GetInvoice(ctx, models.InvoiceID("non-existent"))
	assert.Error(suite.T(), err)

	// Test context cancellation
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately

	_, err = suite.invoiceService.GetInvoice(cancelCtx, models.InvoiceID("any-id"))
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), context.Canceled, err)
}

// TestCompleteWorkflow tests a complete workflow from client creation to invoice management
func (suite *IntegrationTestSuite) TestCompleteWorkflow() {
	ctx := context.Background()

	// Step 1: Create a client
	clientReq := models.CreateClientRequest{
		Name:    "Complete Workflow Client",
		Email:   "complete@workflow.com",
		Address: "789 Workflow Ave",
	}

	client, err := suite.clientService.CreateClient(ctx, clientReq)
	require.NoError(suite.T(), err)

	// Step 2: Create CSV data and parse it
	csvContent := `date,description,hours,rate
2024-01-15,Initial Setup and Configuration,4.0,120.00
2024-01-16,Frontend Development and UI Design,6.5,120.00
2024-01-17,Unit Testing and Bug Fixes,2.0,100.00`

	csvFile := filepath.Join(suite.tempDir, "workflow_timesheet.csv")
	err = os.WriteFile(csvFile, []byte(csvContent), 0o644)
	require.NoError(suite.T(), err)

	file, err := os.Open(csvFile)
	require.NoError(suite.T(), err)
	defer file.Close()

	parseResult, err := suite.csvParser.ParseTimesheet(ctx, file, csv.ParseOptions{})
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), parseResult.WorkItems, 3)

	// Step 3: Create invoice with parsed work items
	now := time.Now()
	invoiceReq := models.CreateInvoiceRequest{
		Number:    "WF-001",
		ClientID:  client.ID,
		Date:      now,
		DueDate:   now.Add(30 * 24 * time.Hour),
		WorkItems: parseResult.WorkItems,
	}

	invoice, err := suite.invoiceService.CreateInvoice(ctx, invoiceReq)
	require.NoError(suite.T(), err)

	// Step 4: Verify the complete invoice
	assert.Equal(suite.T(), "WF-001", invoice.Number)
	assert.Equal(suite.T(), client.ID, invoice.Client.ID)
	assert.Len(suite.T(), invoice.WorkItems, 3)
	assert.Equal(suite.T(), models.StatusDraft, invoice.Status)

	// Calculate expected total: 4*120 + 6.5*120 + 2*100 = 480 + 780 + 200 = 1460
	expectedSubtotal := 4.0*120.0 + 6.5*120.0 + 2.0*100.0
	assert.Equal(suite.T(), expectedSubtotal, invoice.Subtotal)

	// Step 5: Verify we can list and retrieve the invoice
	listResult, err := suite.storage.ListInvoices(ctx, models.InvoiceFilter{})
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), listResult.Invoices, 1)
	assert.Equal(suite.T(), "WF-001", listResult.Invoices[0].Number)

	retrievedInvoice, err := suite.invoiceService.GetInvoice(ctx, invoice.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), invoice.ID, retrievedInvoice.ID)
	assert.Equal(suite.T(), "WF-001", retrievedInvoice.Number)
}

// TestIntegrationSuite runs the integration test suite
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
