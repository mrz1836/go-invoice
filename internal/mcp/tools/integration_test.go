package tools

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// ToolIntegrationTestSuite tests the complete tool registry and discovery integration.
//
// This test suite validates that all 21 tools are properly registered and that
// the discovery, validation, and initialization systems work together correctly.
type ToolIntegrationTestSuite struct {
	suite.Suite

	logger     *IntegrationTestLogger
	components *ToolSystemComponents
}

// SetupTest initializes the test environment for each test.
func (suite *ToolIntegrationTestSuite) SetupTest() {
	suite.logger = NewIntegrationTestLogger()
}

// TestCompleteToolSystemInitialization validates the entire tool system can be initialized.
func (suite *ToolIntegrationTestSuite) TestCompleteToolSystemInitialization() {
	// Initialize the complete tool system
	ctx := context.Background()
	components, err := InitializeToolSystem(ctx, suite.logger)
	suite.Require().NoError(err, "Tool system initialization should succeed")
	suite.Require().NotNil(components, "Components should not be nil")

	suite.components = components

	// Validate registry is populated
	suite.NotNil(components.Registry, "Registry should not be nil")
	suite.NotNil(components.Validator, "Validator should not be nil")
	suite.NotNil(components.DiscoveryService, "Discovery service should not be nil")
	suite.NotNil(components.Metrics, "Metrics should not be nil")

	// Validate tool count
	allTools, err := components.Registry.ListTools(ctx, "")
	suite.Require().NoError(err, "Listing all tools should succeed")
	suite.Len(allTools, 21, "Should have exactly 21 tools registered")

	// Validate category count
	categories, err := components.Registry.GetCategories(ctx)
	suite.Require().NoError(err, "Getting categories should succeed")
	suite.Len(categories, 5, "Should have exactly 5 categories")

	// Validate expected categories are present
	expectedCategories := map[CategoryType]bool{
		CategoryInvoiceManagement: false,
		CategoryClientManagement:  false,
		CategoryDataImport:        false,
		CategoryDataExport:        false,
		CategoryConfiguration:     false,
	}

	for _, category := range categories {
		expectedCategories[category] = true
	}

	for category, found := range expectedCategories {
		suite.True(found, "Category %s should be present", category)
	}
}

// TestToolDiscoveryIntegration validates the discovery service works with all tools.
func (suite *ToolIntegrationTestSuite) TestToolDiscoveryIntegration() {
	suite.setupComponents()

	// Test search functionality
	searchCriteria := &ToolSearchCriteria{
		Query:             "invoice",
		MaxResults:        20,
		MinRelevanceScore: 0.1,
		SortBy:            "relevance",
		SortOrder:         "desc",
	}

	ctx := context.Background()
	results, err := suite.components.DiscoveryService.SearchTools(ctx, searchCriteria)
	suite.Require().NoError(err, "Search should succeed")
	suite.NotEmpty(results, "Search for 'invoice' should return results")

	// Validate search results have relevance scores
	for _, result := range results {
		suite.NotNil(result.Tool, "Search result should have a tool")
		suite.GreaterOrEqual(result.RelevanceScore, 0.0, "Relevance score should be non-negative")
		suite.NotEmpty(result.MatchContext, "Match context should be provided")
	}
}

// TestCategoryBasedDiscovery validates category-based tool discovery.
func (suite *ToolIntegrationTestSuite) TestCategoryBasedDiscovery() {
	suite.setupComponents()
	ctx := context.Background()

	// Test discovery for each category
	categories := []CategoryType{
		CategoryInvoiceManagement,
		CategoryClientManagement,
		CategoryDataImport,
		CategoryDataExport,
		CategoryConfiguration,
	}

	expectedToolCounts := map[CategoryType]int{
		CategoryInvoiceManagement: 7,
		CategoryClientManagement:  5,
		CategoryDataImport:        3,
		CategoryDataExport:        3,
		CategoryConfiguration:     3,
	}

	for _, category := range categories {
		result, err := suite.components.DiscoveryService.DiscoverToolsByCategory(ctx, category)
		suite.Require().NoError(err, "Category discovery should succeed for %s", category)

		expectedCount := expectedToolCounts[category]
		suite.Equal(expectedCount, result.ToolCount, "Category %s should have %d tools", category, expectedCount)
		suite.Len(result.Tools, expectedCount, "Tools slice should match tool count")
		suite.Equal(category, result.Category, "Category should match requested category")
		suite.NotEmpty(result.RelatedCategories, "Should have related categories")
	}
}

// TestToolValidationIntegration validates that tool validation works for all tools.
func (suite *ToolIntegrationTestSuite) TestToolValidationIntegration() {
	suite.setupComponents()
	ctx := context.Background()

	// Get all tools
	allTools, err := suite.components.Registry.ListTools(ctx, "")
	suite.Require().NoError(err, "Listing tools should succeed")

	// Test validation with empty input (should fail for tools with required fields)
	emptyInput := map[string]interface{}{}

	validationAttempts := 0
	validationErrors := 0
	for _, tool := range allTools {
		validationAttempts++
		err := suite.components.Registry.ValidateToolInput(ctx, tool.Name, emptyInput)
		if err != nil {
			validationErrors++
			// Validation errors should be descriptive
			suite.NotEmpty(err.Error(), "Validation error should not be empty")
		}
	}

	// We should have attempted to validate all tools
	suite.Equal(21, validationAttempts, "Should validate all 21 tools")

	// Some tools might have validation errors with empty input
	suite.T().Logf("Validation attempts: %d, Validation errors: %d", validationAttempts, validationErrors)
}

// TestSpecificToolRetrieval validates that specific tools can be retrieved correctly.
func (suite *ToolIntegrationTestSuite) TestSpecificToolRetrieval() {
	suite.setupComponents()
	ctx := context.Background()

	// Test retrieving specific well-known tools
	expectedTools := []string{
		"invoice_create",
		"invoice_list",
		"client_create",
		"import_csv",
		"generate_html",
		"config_show",
	}

	for _, toolName := range expectedTools {
		tool, err := suite.components.Registry.GetTool(ctx, toolName)
		suite.Require().NoError(err, "Tool %s should be retrievable", toolName)
		suite.Equal(toolName, tool.Name, "Tool name should match")
		suite.NotEmpty(tool.Description, "Tool should have description")
		suite.NotNil(tool.InputSchema, "Tool should have input schema")
		suite.NotEmpty(tool.CLICommand, "Tool should have CLI command")
		suite.Greater(tool.Timeout, time.Duration(0), "Tool should have positive timeout")
	}
}

// TestRegistrationMetrics validates that registry metrics are correct.
func (suite *ToolIntegrationTestSuite) TestRegistrationMetrics() {
	suite.setupComponents()
	ctx := context.Background()

	metrics, err := suite.components.Registry.GetRegistrationMetrics(ctx)
	suite.Require().NoError(err, "Getting metrics should succeed")

	suite.Equal(21, metrics.TotalTools, "Should have 21 total tools")
	suite.Equal(5, metrics.TotalCategories, "Should have 5 total categories")
	suite.NotZero(metrics.Uptime, "Should have non-zero uptime")

	// Validate tool distribution
	expectedDistribution := map[CategoryType]int{
		CategoryInvoiceManagement: 7,
		CategoryClientManagement:  5,
		CategoryDataImport:        3,
		CategoryDataExport:        3,
		CategoryConfiguration:     3,
	}

	for category, expectedCount := range expectedDistribution {
		actualCount := metrics.ToolsByCategory[category]
		suite.Equal(expectedCount, actualCount, "Category %s should have %d tools", category, expectedCount)
	}
}

// TestContextCancellation validates that context cancellation is properly handled.
func (suite *ToolIntegrationTestSuite) TestContextCancellation() {
	suite.setupComponents()

	// Create a canceled context
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	// Test that operations respect cancellation
	_, err := suite.components.Registry.ListTools(canceledCtx, "")
	suite.Equal(context.Canceled, err, "Should return context.Canceled")

	_, err = suite.components.DiscoveryService.SearchTools(canceledCtx, &ToolSearchCriteria{})
	suite.Equal(context.Canceled, err, "Should return context.Canceled")
}

// setupComponents initializes the components if not already done.
func (suite *ToolIntegrationTestSuite) setupComponents() {
	if suite.components == nil {
		ctx := context.Background()
		components, err := InitializeToolSystem(ctx, suite.logger)
		suite.Require().NoError(err, "Tool system initialization should succeed")
		suite.components = components
	}
}

// TestToolIntegrationSuite runs the integration test suite.
func TestToolIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ToolIntegrationTestSuite))
}

// IntegrationTestLogger provides a test implementation of the Logger interface for integration tests.
type IntegrationTestLogger struct {
	messages []IntegrationLogMessage
}

// IntegrationLogMessage represents a logged message for testing.
type IntegrationLogMessage struct {
	Level   string
	Message string
	KVPairs []interface{}
}

// NewIntegrationTestLogger creates a new test logger.
func NewIntegrationTestLogger() *IntegrationTestLogger {
	return &IntegrationTestLogger{
		messages: make([]IntegrationLogMessage, 0),
	}
}

// Debug logs debug-level messages.
func (l *IntegrationTestLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.messages = append(l.messages, IntegrationLogMessage{
		Level:   "debug",
		Message: msg,
		KVPairs: keysAndValues,
	})
}

// Info logs info-level messages.
func (l *IntegrationTestLogger) Info(msg string, keysAndValues ...interface{}) {
	l.messages = append(l.messages, IntegrationLogMessage{
		Level:   "info",
		Message: msg,
		KVPairs: keysAndValues,
	})
}

// Warn logs warning-level messages.
func (l *IntegrationTestLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.messages = append(l.messages, IntegrationLogMessage{
		Level:   "warn",
		Message: msg,
		KVPairs: keysAndValues,
	})
}

// Error logs error-level messages.
func (l *IntegrationTestLogger) Error(msg string, keysAndValues ...interface{}) {
	l.messages = append(l.messages, IntegrationLogMessage{
		Level:   "error",
		Message: msg,
		KVPairs: keysAndValues,
	})
}

// GetMessages returns all logged messages.
func (l *IntegrationTestLogger) GetMessages() []IntegrationLogMessage {
	return l.messages
}

// GetMessagesWithLevel returns messages of a specific level.
func (l *IntegrationTestLogger) GetMessagesWithLevel(level string) []IntegrationLogMessage {
	var filtered []IntegrationLogMessage
	for _, msg := range l.messages {
		if msg.Level == level {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}
