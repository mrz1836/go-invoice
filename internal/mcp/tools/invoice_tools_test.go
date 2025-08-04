package tools

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// InvoiceToolsTestSuite provides test suite for invoice management tools
type InvoiceToolsTestSuite struct {
	suite.Suite
}

// SetupTest initializes each test
func (suite *InvoiceToolsTestSuite) SetupTest() {
	// No setup needed now that context is passed to individual tests
}

// TestCreateInvoiceManagementTools tests that all invoice tools are created properly
func (suite *InvoiceToolsTestSuite) TestCreateInvoiceManagementTools() {
	tools := CreateInvoiceManagementTools()

	// Verify we get all 7 expected tools
	suite.Len(tools, 7, "Expected 7 invoice management tools")

	// Verify tool names are correct
	expectedNames := []string{
		"invoice_create",
		"invoice_list",
		"invoice_show",
		"invoice_update",
		"invoice_delete",
		"invoice_add_item",
		"invoice_remove_item",
	}

	actualNames := make([]string, len(tools))
	for i, tool := range tools {
		actualNames[i] = tool.Name
	}

	suite.ElementsMatch(expectedNames, actualNames, "Tool names should match expected values")

	// Verify all tools have required fields
	for _, tool := range tools {
		suite.NotEmpty(tool.Name, "Tool name should not be empty")
		suite.NotEmpty(tool.Description, "Tool description should not be empty")
		suite.NotNil(tool.InputSchema, "Tool input schema should not be nil")
		suite.NotEmpty(tool.Examples, "Tool should have examples")
		suite.Equal(CategoryInvoiceManagement, tool.Category, "Tool should be in invoice management category")
		suite.NotEmpty(tool.CLICommand, "Tool should have CLI command")
		suite.NotEmpty(tool.CLIArgs, "Tool should have CLI args")
		suite.NotEmpty(tool.Version, "Tool should have version")
		suite.Greater(tool.Timeout, time.Duration(0), "Tool should have positive timeout")
	}
}

// TestInvoiceToolSchemas tests that all tools have valid input schemas
func (suite *InvoiceToolsTestSuite) TestInvoiceToolSchemas() {
	tools := CreateInvoiceManagementTools()

	for _, tool := range tools {
		schema := tool.InputSchema

		// All schemas should be objects
		schemaType, exists := schema["type"]
		suite.True(exists, "Schema should have type field for tool %s", tool.Name)
		suite.Equal("object", schemaType, "Schema type should be object for tool %s", tool.Name)

		// All schemas should have properties
		properties, exists := schema["properties"]
		suite.True(exists, "Schema should have properties for tool %s", tool.Name)
		suite.IsType(map[string]interface{}{}, properties, "Properties should be map for tool %s", tool.Name)

		// Properties should not be empty
		propsMap := properties.(map[string]interface{})
		suite.NotEmpty(propsMap, "Properties should not be empty for tool %s", tool.Name)

		// Should not allow additional properties
		additionalProps, exists := schema["additionalProperties"]
		if exists {
			suite.False(additionalProps.(bool), "Should not allow additional properties for tool %s", tool.Name)
		}
	}
}

// TestInvoiceToolExamples tests that all tools have comprehensive examples
func (suite *InvoiceToolsTestSuite) TestInvoiceToolExamples() {
	tools := CreateInvoiceManagementTools()

	for _, tool := range tools {
		suite.GreaterOrEqual(len(tool.Examples), 2, "Tool %s should have at least 2 examples", tool.Name)

		for i, example := range tool.Examples {
			suite.NotEmpty(example.Description, "Example %d description should not be empty for tool %s", i, tool.Name)
			suite.NotNil(example.Input, "Example %d input should not be nil for tool %s", i, tool.Name)
			suite.NotEmpty(example.Input, "Example %d input should not be empty for tool %s", i, tool.Name)
			suite.NotEmpty(example.ExpectedOutput, "Example %d expected output should not be empty for tool %s", i, tool.Name)
			suite.NotEmpty(example.UseCase, "Example %d use case should not be empty for tool %s", i, tool.Name)
		}
	}
}

// TestRegisterInvoiceManagementTools tests tool registration functionality
func (suite *InvoiceToolsTestSuite) TestRegisterInvoiceManagementTools() {
	// Create a mock registry for testing
	validator := NewDefaultInputValidator(&TestLogger{})
	registry := NewDefaultToolRegistry(validator, &TestLogger{})

	// Register invoice tools
	ctx := context.Background()
	err := RegisterInvoiceManagementTools(ctx, registry)
	suite.Require().NoError(err, "Should be able to register invoice management tools")

	// Verify tools are registered
	tools := CreateInvoiceManagementTools()
	for _, expectedTool := range tools {
		registeredTool, getErr := registry.GetTool(ctx, expectedTool.Name)
		suite.Require().NoError(getErr, "Should be able to get registered tool %s", expectedTool.Name)
		suite.Equal(expectedTool.Name, registeredTool.Name, "Registered tool name should match")
		suite.Equal(expectedTool.Category, registeredTool.Category, "Registered tool category should match")
	}

	// Verify tools are in correct category
	categoryTools, err := registry.ListTools(ctx, CategoryInvoiceManagement)
	suite.Require().NoError(err, "Should be able to list tools by category")
	suite.Len(categoryTools, 7, "Should have 7 tools in invoice management category")
}

// TestRegisterInvoiceManagementToolsContextCancellation tests context cancellation
func (suite *InvoiceToolsTestSuite) TestRegisterInvoiceManagementToolsContextCancellation() {
	validator := NewDefaultInputValidator(&TestLogger{})
	registry := NewDefaultToolRegistry(validator, &TestLogger{})

	// Create canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should fail with context cancellation
	err := RegisterInvoiceManagementTools(ctx, registry)
	suite.Require().Error(err, "Should fail with canceled context")
	suite.Equal(context.Canceled, err, "Error should be context.Canceled")
}

// TestLogger provides a test implementation of the Logger interface
type TestLogger struct{}

func (l *TestLogger) Debug(_ string, _ ...interface{}) {}
func (l *TestLogger) Info(_ string, _ ...interface{})  {}
func (l *TestLogger) Warn(_ string, _ ...interface{})  {}
func (l *TestLogger) Error(_ string, _ ...interface{}) {}

// TestInvoiceToolsTestSuite runs the test suite
func TestInvoiceToolsTestSuite(t *testing.T) {
	suite.Run(t, new(InvoiceToolsTestSuite))
}
