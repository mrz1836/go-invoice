package executor

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// BridgeBuildersTestSuite tests the arg builder functions
type BridgeBuildersTestSuite struct {
	suite.Suite

	bridge      *CLIBridge
	logger      *MockLogger
	executor    *MockCommandExecutor
	fileHandler *MockFileHandler
}

func (suite *BridgeBuildersTestSuite) SetupTest() {
	suite.logger = new(MockLogger)
	suite.executor = new(MockCommandExecutor)
	suite.fileHandler = new(MockFileHandler)

	// Setup logger expectations
	suite.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	suite.logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	suite.logger.On("Warn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	suite.logger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	suite.bridge = NewCLIBridge(suite.logger, suite.executor, suite.fileHandler, "")
}

// TestBuildInvoiceCreateArgsWithClientID tests invoice creation with client_id
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceCreateArgsWithClientID() {
	input := map[string]interface{}{
		"client_id":   "client-123",
		"description": "Test Invoice",
	}

	args, err := suite.bridge.buildInvoiceCreateArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "--client-id")
	suite.Contains(args, "client-123")
	suite.Contains(args, "--description")
	suite.Contains(args, "Test Invoice")
}

// TestBuildInvoiceCreateArgsWithClientName tests invoice creation with client_name
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceCreateArgsWithClientName() {
	input := map[string]interface{}{
		"client_name": "ACME Corp",
	}

	args, err := suite.bridge.buildInvoiceCreateArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "--client")
	suite.Contains(args, "ACME Corp")
}

// TestBuildInvoiceCreateArgsWithClientEmail tests invoice creation with client_email
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceCreateArgsWithClientEmail() {
	input := map[string]interface{}{
		"client_email": "client@example.com",
	}

	args, err := suite.bridge.buildInvoiceCreateArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "--client-email")
	suite.Contains(args, "client@example.com")
}

// TestBuildInvoiceCreateArgsMissingClient tests invoice creation without client
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceCreateArgsMissingClient() {
	input := map[string]interface{}{
		"description": "Test Invoice",
	}

	args, err := suite.bridge.buildInvoiceCreateArgs(input)

	suite.Require().ErrorIs(err, ErrMissingRequired)
	suite.Nil(args)
}

// TestBuildInvoiceCreateArgsWithAllOptions tests invoice creation with all options
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceCreateArgsWithAllOptions() {
	input := map[string]interface{}{
		"client_id":    "client-123",
		"description":  "Full Invoice",
		"invoice_date": "2024-01-15",
		"due_date":     "2024-02-15",
	}

	args, err := suite.bridge.buildInvoiceCreateArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "--date")
	suite.Contains(args, "2024-01-15")
	suite.Contains(args, "--due")
	suite.Contains(args, "2024-02-15")
}

// TestBuildInvoiceCreateArgsWithWorkItems tests invoice creation with work items
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceCreateArgsWithWorkItems() {
	input := map[string]interface{}{
		"client_id": "client-123",
		"work_items": []interface{}{
			map[string]interface{}{
				"description": "Development work",
				"hours":       float64(8),
				"rate":        float64(150),
				"date":        "2024-01-15",
			},
		},
	}

	args, err := suite.bridge.buildInvoiceCreateArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "--add-item-description")
	suite.Contains(args, "Development work")
	suite.Contains(args, "--add-item-hours")
	suite.Contains(args, "--add-item-rate")
	suite.Contains(args, "--add-item-date")
}

// TestBuildInvoiceCreateArgsWithCreateClient tests invoice creation with create_client_if_missing
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceCreateArgsWithCreateClient() {
	input := map[string]interface{}{
		"client_name":              "New Client",
		"create_client_if_missing": true,
		"new_client_email":         "new@example.com",
		"new_client_phone":         "123-456-7890",
		"new_client_address":       "123 Main St",
	}

	args, err := suite.bridge.buildInvoiceCreateArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "--create-client")
	suite.Contains(args, "--new-client-email")
	suite.Contains(args, "new@example.com")
	suite.Contains(args, "--new-client-phone")
	suite.Contains(args, "--new-client-address")
}

// TestBuildInvoiceListArgs tests invoice list with filters
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceListArgs() {
	input := map[string]interface{}{
		"status":      "pending",
		"client_name": "ACME Corp",
		"from_date":   "2024-01-01",
		"to_date":     "2024-12-31",
	}

	args, err := suite.bridge.buildInvoiceListArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "--output")
	suite.Contains(args, "json")
	suite.Contains(args, "--status")
	suite.Contains(args, "pending")
	suite.Contains(args, "--client")
	suite.Contains(args, "--from")
	suite.Contains(args, "--to")
}

// TestBuildInvoiceListArgsEmpty tests invoice list with no filters
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceListArgsEmpty() {
	input := map[string]interface{}{}

	args, err := suite.bridge.buildInvoiceListArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "--output")
	suite.Contains(args, "json")
}

// TestBuildInvoiceShowArgsWithID tests invoice show with invoice_id
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceShowArgsWithID() {
	input := map[string]interface{}{
		"invoice_id": "inv-123",
	}

	args, err := suite.bridge.buildInvoiceShowArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "inv-123")
	suite.Contains(args, "--output")
	suite.Contains(args, "json")
}

// TestBuildInvoiceShowArgsWithNumber tests invoice show with invoice_number
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceShowArgsWithNumber() {
	input := map[string]interface{}{
		"invoice_number": "INV-2024-001",
	}

	args, err := suite.bridge.buildInvoiceShowArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "INV-2024-001")
}

// TestBuildInvoiceShowArgsMissingIdentifier tests invoice show without identifier
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceShowArgsMissingIdentifier() {
	input := map[string]interface{}{}

	args, err := suite.bridge.buildInvoiceShowArgs(input)

	suite.Require().ErrorIs(err, ErrMissingRequired)
	suite.Nil(args)
}

// TestBuildInvoiceUpdateArgs tests invoice update with fields
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceUpdateArgs() {
	input := map[string]interface{}{
		"invoice_id":  "inv-123",
		"status":      "paid",
		"description": "Updated description",
		"due_date":    "2024-03-01",
	}

	args, err := suite.bridge.buildInvoiceUpdateArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "inv-123")
	suite.Contains(args, "--status")
	suite.Contains(args, "paid")
	suite.Contains(args, "--description")
	suite.Contains(args, "--due")
}

// TestBuildInvoiceUpdateArgsMissingIdentifier tests update without identifier
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceUpdateArgsMissingIdentifier() {
	input := map[string]interface{}{
		"status": "paid",
	}

	args, err := suite.bridge.buildInvoiceUpdateArgs(input)

	suite.Require().ErrorIs(err, ErrMissingRequired)
	suite.Nil(args)
}

// TestBuildInvoiceUpdateArgsMissingFields tests update without update fields
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceUpdateArgsMissingFields() {
	input := map[string]interface{}{
		"invoice_id": "inv-123",
	}

	args, err := suite.bridge.buildInvoiceUpdateArgs(input)

	suite.Require().ErrorIs(err, ErrMissingUpdateFields)
	suite.Nil(args)
}

// TestBuildInvoiceDeleteArgs tests invoice delete with force
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceDeleteArgs() {
	input := map[string]interface{}{
		"invoice_id": "inv-123",
		"force":      true,
	}

	args, err := suite.bridge.buildInvoiceDeleteArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "inv-123")
	suite.Contains(args, "--force")
}

// TestBuildInvoiceDeleteArgsMissingID tests delete without invoice_id
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceDeleteArgsMissingID() {
	input := map[string]interface{}{}

	args, err := suite.bridge.buildInvoiceDeleteArgs(input)

	suite.Require().ErrorIs(err, ErrMissingRequired)
	suite.Nil(args)
}

// TestBuildInvoiceAddItemArgs tests adding item with all fields
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceAddItemArgs() {
	input := map[string]interface{}{
		"invoice_id":  "inv-123",
		"description": "Development work",
		"hours":       float64(8),
		"rate":        float64(150),
		"date":        "2024-01-15",
	}

	args, err := suite.bridge.buildInvoiceAddItemArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "inv-123")
	suite.Contains(args, "--description")
	suite.Contains(args, "--hours")
	suite.Contains(args, "--rate")
	suite.Contains(args, "--date")
}

// TestBuildInvoiceAddItemArgsMissingInvoice tests adding item without invoice
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceAddItemArgsMissingInvoice() {
	input := map[string]interface{}{
		"description": "Work",
	}

	args, err := suite.bridge.buildInvoiceAddItemArgs(input)

	suite.Require().ErrorIs(err, ErrMissingRequired)
	suite.Nil(args)
}

// TestBuildInvoiceRemoveItemArgsWithID tests removing item with item_id
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceRemoveItemArgsWithID() {
	input := map[string]interface{}{
		"invoice_id": "inv-123",
		"item_id":    "item-456",
	}

	args, err := suite.bridge.buildInvoiceRemoveItemArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "inv-123")
	suite.Contains(args, "--item-id")
	suite.Contains(args, "item-456")
}

// TestBuildInvoiceRemoveItemArgsWithIndex tests removing item with item_index
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceRemoveItemArgsWithIndex() {
	input := map[string]interface{}{
		"invoice_id": "inv-123",
		"item_index": float64(1),
	}

	args, err := suite.bridge.buildInvoiceRemoveItemArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "inv-123")
	suite.Contains(args, "--index")
}

// TestBuildInvoiceRemoveItemArgsMissingItemIdentifier tests removing item without identifiers
func (suite *BridgeBuildersTestSuite) TestBuildInvoiceRemoveItemArgsMissingItemIdentifier() {
	input := map[string]interface{}{
		"invoice_id": "inv-123",
	}

	args, err := suite.bridge.buildInvoiceRemoveItemArgs(input)

	// The actual error wraps ErrMissingRequired
	suite.Require().ErrorIs(err, ErrMissingRequired)
	suite.Nil(args)
}

// TestBuildClientCreateArgs tests client creation
func (suite *BridgeBuildersTestSuite) TestBuildClientCreateArgs() {
	input := map[string]interface{}{
		"name":    "ACME Corp",
		"email":   "contact@acme.com",
		"phone":   "123-456-7890",
		"address": "123 Main St",
	}

	args, err := suite.bridge.buildClientCreateArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "--name")
	suite.Contains(args, "ACME Corp")
	suite.Contains(args, "--email")
	suite.Contains(args, "--phone")
	suite.Contains(args, "--address")
}

// TestBuildClientCreateArgsMissingName tests client creation without name
func (suite *BridgeBuildersTestSuite) TestBuildClientCreateArgsMissingName() {
	input := map[string]interface{}{
		"email": "contact@acme.com",
	}

	args, err := suite.bridge.buildClientCreateArgs(input)

	suite.Require().ErrorIs(err, ErrMissingRequired)
	suite.Nil(args)
}

// TestBuildClientListArgs tests client list with filters
func (suite *BridgeBuildersTestSuite) TestBuildClientListArgs() {
	input := map[string]interface{}{
		"search": "ACME",
		"active": true,
	}

	args, err := suite.bridge.buildClientListArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "--output")
	suite.Contains(args, "json")
	suite.Contains(args, "--search")
	suite.Contains(args, "ACME")
	suite.Contains(args, "--active")
}

// TestBuildClientShowArgs tests client show
func (suite *BridgeBuildersTestSuite) TestBuildClientShowArgs() {
	input := map[string]interface{}{
		"client_id": "client-123",
	}

	args, err := suite.bridge.buildClientShowArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "client-123")
	suite.Contains(args, "--output")
	suite.Contains(args, "json")
}

// TestBuildClientShowArgsMissingID tests client show without ID
func (suite *BridgeBuildersTestSuite) TestBuildClientShowArgsMissingID() {
	input := map[string]interface{}{}

	args, err := suite.bridge.buildClientShowArgs(input)

	suite.Require().ErrorIs(err, ErrMissingClientIdentifier)
	suite.Nil(args)
}

// TestBuildClientUpdateArgs tests client update
func (suite *BridgeBuildersTestSuite) TestBuildClientUpdateArgs() {
	input := map[string]interface{}{
		"client_id": "client-123",
		"name":      "Updated Name",
		"email":     "new@example.com",
	}

	args, err := suite.bridge.buildClientUpdateArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "client-123")
	suite.Contains(args, "--name")
	suite.Contains(args, "--email")
}

// TestBuildClientUpdateArgsMissingID tests client update without ID
func (suite *BridgeBuildersTestSuite) TestBuildClientUpdateArgsMissingID() {
	input := map[string]interface{}{
		"name": "Updated Name",
	}

	args, err := suite.bridge.buildClientUpdateArgs(input)

	suite.Require().ErrorIs(err, ErrMissingClientIDOrName)
	suite.Nil(args)
}

// TestBuildClientDeleteArgs tests client delete with force
func (suite *BridgeBuildersTestSuite) TestBuildClientDeleteArgs() {
	input := map[string]interface{}{
		"client_id": "client-123",
		"force":     true,
	}

	args, err := suite.bridge.buildClientDeleteArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "client-123")
	suite.Contains(args, "--force")
}

// TestBuildClientDeleteArgsMissingID tests client delete without ID
func (suite *BridgeBuildersTestSuite) TestBuildClientDeleteArgsMissingID() {
	input := map[string]interface{}{}

	args, err := suite.bridge.buildClientDeleteArgs(input)

	suite.Require().ErrorIs(err, ErrMissingClientIDOrName)
	suite.Nil(args)
}

// TestBuildImportCSVArgs tests CSV import for new invoice
func (suite *BridgeBuildersTestSuite) TestBuildImportCSVArgs() {
	input := map[string]interface{}{
		"file_path":   "/tmp/timesheet.csv",
		"client_id":   "client-123",
		"description": "Imported items",
	}

	args, err := suite.bridge.buildImportCSVArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "/tmp/timesheet.csv")
	suite.Contains(args, "--client-id")
	suite.Contains(args, "--description")
}

// TestBuildImportCSVArgsAppendMode tests CSV import in append mode
func (suite *BridgeBuildersTestSuite) TestBuildImportCSVArgsAppendMode() {
	input := map[string]interface{}{
		"file_path":   "/tmp/timesheet.csv",
		"import_mode": "append_invoice",
		"invoice_id":  "inv-123",
	}

	args, err := suite.bridge.buildImportCSVArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "/tmp/timesheet.csv")
	suite.Contains(args, "--invoice")
	suite.Contains(args, "inv-123")
}

// TestBuildImportCSVArgsMissingFile tests CSV import without file
func (suite *BridgeBuildersTestSuite) TestBuildImportCSVArgsMissingFile() {
	input := map[string]interface{}{
		"client_id": "client-123",
	}

	args, err := suite.bridge.buildImportCSVArgs(input)

	suite.Require().ErrorIs(err, ErrMissingRequired)
	suite.Nil(args)
}

// TestBuildGenerateHTMLArgs tests HTML generation
func (suite *BridgeBuildersTestSuite) TestBuildGenerateHTMLArgs() {
	input := map[string]interface{}{
		"invoice_id":  "inv-123",
		"output_path": "/tmp/invoice.html",
		"template":    "modern",
	}

	args, err := suite.bridge.buildGenerateHTMLArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "inv-123")
	suite.Contains(args, "--output")
	suite.Contains(args, "/tmp/invoice.html")
	suite.Contains(args, "--template")
	suite.Contains(args, "modern")
}

// TestBuildGenerateHTMLArgsMissingInvoice tests HTML generation without invoice
func (suite *BridgeBuildersTestSuite) TestBuildGenerateHTMLArgsMissingInvoice() {
	input := map[string]interface{}{
		"output_path": "/tmp/invoice.html",
	}

	args, err := suite.bridge.buildGenerateHTMLArgs(input)

	suite.Require().ErrorIs(err, ErrMissingRequired)
	suite.Nil(args)
}

// TestBuildConfigShowArgs tests config show (uses --output json)
func (suite *BridgeBuildersTestSuite) TestBuildConfigShowArgs() {
	input := map[string]interface{}{
		"section": "database",
	}

	args, err := suite.bridge.buildConfigShowArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "--output")
	suite.Contains(args, "json")
	suite.Contains(args, "--section")
	suite.Contains(args, "database")
}

// TestBuildConfigValidateArgs tests config validate
func (suite *BridgeBuildersTestSuite) TestBuildConfigValidateArgs() {
	input := map[string]interface{}{}

	args, err := suite.bridge.buildConfigValidateArgs(input)

	suite.Require().NoError(err)
	// Verify args is returned (could be empty or nil for no options)
	suite.Empty(args)
}

// TestBuildConfigInitArgs tests config init (uses --force, not --format)
func (suite *BridgeBuildersTestSuite) TestBuildConfigInitArgs() {
	input := map[string]interface{}{
		"force":    true,
		"template": "default",
	}

	args, err := suite.bridge.buildConfigInitArgs(input)

	suite.Require().NoError(err)
	suite.Contains(args, "--force")
	suite.Contains(args, "--template")
	suite.Contains(args, "default")
}

func TestBridgeBuildersTestSuite(t *testing.T) {
	suite.Run(t, new(BridgeBuildersTestSuite))
}
