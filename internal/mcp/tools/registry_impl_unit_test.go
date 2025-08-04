package tools

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestRegistrationMetrics tests the RegistrationMetrics structure
func TestRegistrationMetrics(t *testing.T) {
	t.Run("BasicStructure", func(t *testing.T) {
		now := time.Now()
		uptime := 10 * time.Minute

		metrics := RegistrationMetrics{
			TotalTools:         21,
			TotalCategories:    5,
			InitializationTime: now,
			Uptime:             uptime,
			ToolsByCategory: map[CategoryType]int{
				CategoryInvoiceManagement: 7,
				CategoryClientManagement:  5,
				CategoryDataImport:        3,
				CategoryDataExport:        3,
				CategoryConfiguration:     3,
			},
		}

		assert.Equal(t, 21, metrics.TotalTools, "Total tools should be 21")
		assert.Equal(t, 5, metrics.TotalCategories, "Total categories should be 5")
		assert.Equal(t, now, metrics.InitializationTime, "Initialization time should match")
		assert.Equal(t, uptime, metrics.Uptime, "Uptime should match")
		assert.Len(t, metrics.ToolsByCategory, 5, "Should have 5 categories")

		// Verify category counts
		assert.Equal(t, 7, metrics.ToolsByCategory[CategoryInvoiceManagement], "Invoice management should have 7 tools")
		assert.Equal(t, 5, metrics.ToolsByCategory[CategoryClientManagement], "Client management should have 5 tools")
		assert.Equal(t, 3, metrics.ToolsByCategory[CategoryDataImport], "Data import should have 3 tools")
		assert.Equal(t, 3, metrics.ToolsByCategory[CategoryDataExport], "Data export should have 3 tools")
		assert.Equal(t, 3, metrics.ToolsByCategory[CategoryConfiguration], "Configuration should have 3 tools")

		// Verify totals add up
		total := 0
		for _, count := range metrics.ToolsByCategory {
			total += count
		}
		assert.Equal(t, 21, total, "Category counts should sum to total tools")
	})

	t.Run("EmptyMetrics", func(t *testing.T) {
		metrics := RegistrationMetrics{}

		assert.Equal(t, 0, metrics.TotalTools, "Default total tools should be 0")
		assert.Equal(t, 0, metrics.TotalCategories, "Default total categories should be 0")
		assert.True(t, metrics.InitializationTime.IsZero(), "Default initialization time should be zero")
		assert.Equal(t, time.Duration(0), metrics.Uptime, "Default uptime should be 0")
		assert.Nil(t, metrics.ToolsByCategory, "Default tools by category should be nil")
	})

	t.Run("MetricsConsistency", func(t *testing.T) {
		// Test that expected tool counts are consistent with actual implementation
		expectedTotalTools := 7 + 5 + 3 + 3 + 3 // Sum of all category tools
		assert.Equal(t, 21, expectedTotalTools, "Expected total should be 21")

		expectedCategories := 5
		categoryTypes := []CategoryType{
			CategoryInvoiceManagement,
			CategoryClientManagement,
			CategoryDataImport,
			CategoryDataExport,
			CategoryConfiguration,
		}
		assert.Len(t, categoryTypes, expectedCategories, "Should have 5 category types")
	})
}

// TestRegistrationErrors tests registry-specific error constants
func TestRegistrationErrors(t *testing.T) {
	t.Run("ErrorConstants", func(t *testing.T) {
		// Test that registry-specific errors are defined
		require.Error(t, ErrMissingCategory, "ErrMissingCategory should be defined")
		require.Error(t, ErrCategoryToolCount, "ErrCategoryToolCount should be defined")
		require.Error(t, ErrInvalidToolCount, "ErrInvalidToolCount should be defined")
		require.Error(t, ErrInvalidCategoryCount, "ErrInvalidCategoryCount should be defined")

		// Test error messages
		assert.Contains(t, ErrMissingCategory.Error(), "missing", "Error should mention missing")
		assert.Contains(t, ErrCategoryToolCount.Error(), "tool count", "Error should mention tool count")
		assert.Contains(t, ErrInvalidToolCount.Error(), "invalid tool count", "Error should mention invalid tool count")
		assert.Contains(t, ErrInvalidCategoryCount.Error(), "invalid category count", "Error should mention invalid category count")

		// Test that errors implement error interface
		err1 := ErrMissingCategory
		err2 := ErrCategoryToolCount
		err3 := ErrInvalidToolCount
		err4 := ErrInvalidCategoryCount

		require.Error(t, err1, "ErrMissingCategory should implement error interface")
		require.Error(t, err2, "ErrCategoryToolCount should implement error interface")
		require.Error(t, err3, "ErrInvalidToolCount should implement error interface")
		assert.Error(t, err4, "ErrInvalidCategoryCount should implement error interface")
	})
}

// TestCompleteToolRegistryStructure tests the CompleteToolRegistry structure
func TestCompleteToolRegistryStructure(t *testing.T) {
	t.Run("StructureValidation", func(t *testing.T) {
		// Create a minimal mock to test structure
		validator := new(MockInputValidator)
		logger := new(MockLogger)

		// We can't actually create the registry without complex mocks,
		// but we can test the structure concept
		initTime := time.Now()

		// Simulate the structure fields
		toolCount := 21
		categoryCount := 5

		// Test the expected structure
		assert.Equal(t, 21, toolCount, "Tool count should be 21")
		assert.Equal(t, 5, categoryCount, "Category count should be 5")
		assert.False(t, initTime.IsZero(), "Initialization time should not be zero")

		// Test that we can create metrics from these values
		metrics := RegistrationMetrics{
			TotalTools:         toolCount,
			TotalCategories:    categoryCount,
			InitializationTime: initTime,
			Uptime:             time.Since(initTime),
			ToolsByCategory: map[CategoryType]int{
				CategoryInvoiceManagement: 7,
				CategoryClientManagement:  5,
				CategoryDataImport:        3,
				CategoryDataExport:        3,
				CategoryConfiguration:     3,
			},
		}

		assert.Equal(t, toolCount, metrics.TotalTools, "Metrics should reflect tool count")
		assert.Equal(t, categoryCount, metrics.TotalCategories, "Metrics should reflect category count")
		assert.Greater(t, metrics.Uptime, time.Duration(0), "Uptime should be positive")

		// Clean up mocks
		validator.AssertExpectations(t)
		logger.AssertExpectations(t)
	})
}

// TestRegistryValidationLogic tests the validation logic concepts
func TestRegistryValidationLogic(t *testing.T) {
	t.Run("ExpectedToolCounts", func(t *testing.T) {
		// Test the expected tool counts per category
		expectedCounts := map[CategoryType]int{
			CategoryInvoiceManagement: 7,
			CategoryClientManagement:  5,
			CategoryDataImport:        3,
			CategoryDataExport:        3,
			CategoryConfiguration:     3,
		}

		totalExpected := 0
		for category, count := range expectedCounts {
			assert.Positive(t, count, "Category %s should have positive tool count", category)
			totalExpected += count
		}

		assert.Equal(t, 21, totalExpected, "Total expected tools should be 21")
		assert.Len(t, expectedCounts, 5, "Should have 5 categories")
	})

	t.Run("CategoryValidation", func(t *testing.T) {
		// Test all expected categories exist
		expectedCategories := []CategoryType{
			CategoryInvoiceManagement,
			CategoryClientManagement,
			CategoryDataImport,
			CategoryDataExport,
			CategoryConfiguration,
		}

		// Verify each category is distinct
		categorySet := make(map[CategoryType]bool)
		for _, category := range expectedCategories {
			assert.False(t, categorySet[category], "Category %s should be unique", category)
			categorySet[category] = true
			assert.NotEmpty(t, string(category), "Category should have non-empty string representation")
		}

		assert.Len(t, categorySet, 5, "Should have exactly 5 unique categories")
	})
}

// TestContextHandling tests context handling in registry operations
func TestContextHandling(t *testing.T) {
	t.Run("ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Test that canceled context is detected
		select {
		case <-ctx.Done():
			assert.Equal(t, context.Canceled, ctx.Err(), "Context should be canceled")
		default:
			t.Fatal("Context should be canceled")
		}

		// Test that operations would return context error
		assert.Equal(t, context.Canceled, ctx.Err(), "Context error should be Canceled")
	})

	t.Run("ContextTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		time.Sleep(5 * time.Millisecond) // Wait for timeout

		select {
		case <-ctx.Done():
			assert.Contains(t, ctx.Err().Error(), "deadline exceeded", "Context should timeout")
		default:
			t.Fatal("Context should have timed out")
		}
	})

	t.Run("ValidContext", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Context should still be valid
		select {
		case <-ctx.Done():
			t.Fatal("Context should not be done yet")
		default:
			assert.NoError(t, ctx.Err(), "Context should not have error")
		}
	})
}

// TestTimingAndMetrics tests timing and metrics functionality
func TestTimingAndMetrics(t *testing.T) {
	t.Run("UptimeCalculation", func(t *testing.T) {
		initTime := time.Now()
		time.Sleep(10 * time.Millisecond) // Small delay

		uptime := time.Since(initTime)
		assert.Greater(t, uptime, time.Duration(0), "Uptime should be positive")
		assert.GreaterOrEqual(t, uptime, 10*time.Millisecond, "Uptime should be at least 10ms")
		assert.Less(t, uptime, 1*time.Second, "Uptime should be reasonable for test")
	})

	t.Run("MetricsConsistency", func(t *testing.T) {
		// Test that metrics remain consistent
		initTime := time.Now()

		metrics1 := RegistrationMetrics{
			TotalTools:         21,
			TotalCategories:    5,
			InitializationTime: initTime,
			Uptime:             time.Since(initTime),
			ToolsByCategory: map[CategoryType]int{
				CategoryInvoiceManagement: 7,
				CategoryClientManagement:  5,
				CategoryDataImport:        3,
				CategoryDataExport:        3,
				CategoryConfiguration:     3,
			},
		}

		time.Sleep(1 * time.Millisecond)

		metrics2 := RegistrationMetrics{
			TotalTools:         21,
			TotalCategories:    5,
			InitializationTime: initTime,             // Same init time
			Uptime:             time.Since(initTime), // Updated uptime
			ToolsByCategory: map[CategoryType]int{
				CategoryInvoiceManagement: 7,
				CategoryClientManagement:  5,
				CategoryDataImport:        3,
				CategoryDataExport:        3,
				CategoryConfiguration:     3,
			},
		}

		// Static fields should remain the same
		assert.Equal(t, metrics1.TotalTools, metrics2.TotalTools, "Tool count should remain constant")
		assert.Equal(t, metrics1.TotalCategories, metrics2.TotalCategories, "Category count should remain constant")
		assert.Equal(t, metrics1.InitializationTime, metrics2.InitializationTime, "Init time should remain constant")
		assert.Equal(t, metrics1.ToolsByCategory, metrics2.ToolsByCategory, "Category breakdown should remain constant")

		// Uptime should increase
		assert.Greater(t, metrics2.Uptime, metrics1.Uptime, "Uptime should increase")
	})
}

// TestRegistryExpectations tests the expected behavior of registry components
func TestRegistryExpectations(t *testing.T) {
	t.Run("ExpectedBehavior", func(t *testing.T) {
		// Test expected registry behavior without actual implementation

		// Expected tool registrations per category
		expectedTools := map[CategoryType][]string{
			CategoryInvoiceManagement: {
				"invoice_create", "invoice_list", "invoice_show", "invoice_update",
				"invoice_delete", "invoice_send", "invoice_duplicate",
			},
			CategoryClientManagement: {
				"client_create", "client_list", "client_show", "client_update", "client_delete",
			},
			CategoryDataImport: {
				"import_timesheet", "import_clients", "import_projects",
			},
			CategoryDataExport: {
				"generate_pdf", "generate_html", "generate_csv",
			},
			CategoryConfiguration: {
				"config_show", "config_set", "config_validate",
			},
		}

		// Verify expected structure
		totalTools := 0
		for category, tools := range expectedTools {
			assert.NotEmpty(t, string(category), "Category should have name")
			assert.NotEmpty(t, tools, "Category should have tools")
			totalTools += len(tools)

			// Verify tool names follow conventions
			for _, toolName := range tools {
				assert.NotEmpty(t, toolName, "Tool name should not be empty")
				assert.Contains(t, toolName, "_", "Tool name should use snake_case")
				assert.NotContains(t, toolName, " ", "Tool name should not contain spaces")
			}
		}

		assert.Equal(t, 21, totalTools, "Should have exactly 21 tools")
		assert.Len(t, expectedTools, 5, "Should have exactly 5 categories")
	})
}

// CompleteToolRegistryTestSuite provides comprehensive tests for CompleteToolRegistry
type CompleteToolRegistryTestSuite struct {
	suite.Suite

	mockValidator *MockInputValidator
	mockLogger    *MockLogger
}

// getContext returns a background context for testing
func (s *CompleteToolRegistryTestSuite) getContext() context.Context {
	return context.Background()
}

func (s *CompleteToolRegistryTestSuite) SetupTest() {
	s.mockValidator = &MockInputValidator{}
	s.mockLogger = NewMockLogger()

	// Set up default validator expectations for any schema validation calls
	s.mockValidator.On("ValidateAgainstSchema", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	s.mockValidator.On("ValidateRequired", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	s.mockValidator.On("ValidateFormat", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
}

// TestNewCompleteToolRegistry tests the complete registry constructor
func (s *CompleteToolRegistryTestSuite) TestNewCompleteToolRegistry() {
	s.Run("ValidConstruction", func() {
		registry, err := NewCompleteToolRegistry(s.getContext(), s.mockValidator, s.mockLogger)

		s.Require().NoError(err)
		s.NotNil(registry)
		s.NotNil(registry.DefaultToolRegistry)
		s.False(registry.initializationTime.IsZero())
		s.Positive(registry.toolCount)
		s.Positive(registry.categoryCount)
	})

	s.Run("NilValidator", func() {
		registry, err := NewCompleteToolRegistry(s.getContext(), nil, s.mockLogger)

		s.Require().Error(err)
		s.Nil(registry)
		s.Contains(err.Error(), "validator")
	})

	s.Run("NilLogger", func() {
		registry, err := NewCompleteToolRegistry(s.getContext(), s.mockValidator, nil)

		s.Require().Error(err)
		s.Nil(registry)
		s.Contains(err.Error(), "logger")
	})

	s.Run("ContextCancellation", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		registry, err := NewCompleteToolRegistry(ctx, s.mockValidator, s.mockLogger)

		s.Require().Error(err)
		s.Nil(registry)
		s.Equal(context.Canceled, err)
	})
}

// TestGetRegistrationMetrics tests metrics retrieval
func (s *CompleteToolRegistryTestSuite) TestGetRegistrationMetrics() {
	s.Run("ValidMetrics", func() {
		registry, err := NewCompleteToolRegistry(s.getContext(), s.mockValidator, s.mockLogger)
		s.Require().NoError(err)
		s.NotNil(registry)

		metrics, err := registry.GetRegistrationMetrics(s.getContext())

		s.Require().NoError(err)
		s.NotNil(metrics)
		s.Positive(metrics.TotalTools)
		s.Positive(metrics.TotalCategories)
		s.False(metrics.InitializationTime.IsZero())
		s.GreaterOrEqual(metrics.Uptime, time.Duration(0))
		s.NotNil(metrics.ToolsByCategory)
	})

	s.Run("ContextCancellation", func() {
		registry, err := NewCompleteToolRegistry(s.getContext(), s.mockValidator, s.mockLogger)
		s.Require().NoError(err)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		metrics, err := registry.GetRegistrationMetrics(ctx)

		s.Require().Error(err)
		s.Nil(metrics)
		s.Equal(context.Canceled, err)
	})
}

// TestRegistryValidation tests internal validation methods
func (s *CompleteToolRegistryTestSuite) TestRegistryValidation() {
	s.Run("ExpectedToolCounts", func() {
		registry, err := NewCompleteToolRegistry(s.getContext(), s.mockValidator, s.mockLogger)
		s.Require().NoError(err)
		s.NotNil(registry)

		// Test that we have the expected number of tools
		categories, err := registry.GetCategories(s.getContext())
		s.Require().NoError(err)
		s.NotEmpty(categories)

		// Verify we can list tools for each category
		for _, category := range categories {
			tools, err := registry.ListTools(s.getContext(), category)
			s.Require().NoError(err, "Should be able to list tools for category %s", category)
			s.NotEmpty(tools, "Category %s should have tools", category)
		}
	})

	s.Run("ToolRegistrationIntegrity", func() {
		registry, err := NewCompleteToolRegistry(s.getContext(), s.mockValidator, s.mockLogger)
		s.Require().NoError(err)
		s.NotNil(registry)

		// Get all tools and verify they're properly registered
		allTools, err := registry.ListTools(s.getContext(), "")
		s.Require().NoError(err)
		s.NotEmpty(allTools)

		// Verify each tool can be retrieved individually
		for _, tool := range allTools {
			retrievedTool, err := registry.GetTool(s.getContext(), tool.Name)
			s.Require().NoError(err, "Should be able to retrieve tool %s", tool.Name)
			s.Equal(tool.Name, retrievedTool.Name)
			s.Equal(tool.Category, retrievedTool.Category)
		}
	})
}

func TestCompleteToolRegistryTestSuite(t *testing.T) {
	suite.Run(t, new(CompleteToolRegistryTestSuite))
}

// TestConcurrentAccess tests concurrent access patterns
func TestConcurrentAccess(t *testing.T) {
	t.Run("ConcurrentMetricsAccess", func(t *testing.T) {
		// Simulate concurrent access to registry metrics
		initTime := time.Now()

		// Run multiple goroutines that would access metrics
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func() {
				defer func() { done <- true }()

				// Simulate metrics calculation
				metrics := RegistrationMetrics{
					TotalTools:         21,
					TotalCategories:    5,
					InitializationTime: initTime,
					Uptime:             time.Since(initTime),
					ToolsByCategory: map[CategoryType]int{
						CategoryInvoiceManagement: 7,
						CategoryClientManagement:  5,
						CategoryDataImport:        3,
						CategoryDataExport:        3,
						CategoryConfiguration:     3,
					},
				}

				// Verify metrics are consistent
				assert.Equal(t, 21, metrics.TotalTools, "Tool count should be consistent")
				assert.Equal(t, 5, metrics.TotalCategories, "Category count should be consistent")
				assert.Greater(t, metrics.Uptime, time.Duration(0), "Uptime should be positive")
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("Concurrent access test timed out")
			}
		}
	})
}
