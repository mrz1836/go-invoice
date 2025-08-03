package tools

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockInputValidator provides a mock implementation of InputValidator for testing
type MockInputValidator struct {
	mock.Mock
}

func (m *MockInputValidator) ValidateAgainstSchema(ctx context.Context, input map[string]interface{}, schema map[string]interface{}) error {
	args := m.Called(ctx, input, schema)
	return args.Error(0)
}

func (m *MockInputValidator) ValidateRequired(ctx context.Context, input map[string]interface{}, requiredFields []string) error {
	args := m.Called(ctx, input, requiredFields)
	return args.Error(0)
}

func (m *MockInputValidator) ValidateFormat(ctx context.Context, fieldName string, value interface{}, format string) error {
	args := m.Called(ctx, fieldName, value, format)
	return args.Error(0)
}

func (m *MockInputValidator) BuildValidationError(ctx context.Context, fieldPath, message string, suggestions []string) error {
	args := m.Called(ctx, fieldPath, message, suggestions)
	return args.Error(0)
}

// RegistryTestSuite provides comprehensive tests for the tool registry implementation
type RegistryTestSuite struct {
	suite.Suite
	registry  *DefaultToolRegistry
	validator *MockInputValidator
	logger    *MockLogger
	ctx       context.Context
}

func (s *RegistryTestSuite) SetupTest() {
	s.validator = new(MockInputValidator)
	s.logger = new(MockLogger)
	s.registry = NewDefaultToolRegistry(s.validator, s.logger)
	s.ctx = context.Background()
}

func (s *RegistryTestSuite) TearDownTest() {
	s.validator.AssertExpectations(s.T())
	s.logger.AssertExpectations(s.T())
}

func (s *RegistryTestSuite) TestNewDefaultToolRegistry() {
	s.Run("ValidCreation", func() {
		validator := new(MockInputValidator)
		logger := new(MockLogger)
		registry := NewDefaultToolRegistry(validator, logger)

		s.NotNil(registry, "Registry should be created")
		s.Equal(validator, registry.validator, "Validator should be assigned")
		s.Equal(logger, registry.logger, "Logger should be assigned")
		s.NotNil(registry.tools, "Tools map should be initialized")
		s.NotNil(registry.categories, "Categories map should be initialized")
		s.Empty(registry.tools, "Tools map should start empty")
		s.Empty(registry.categories, "Categories map should start empty")
	})

	s.Run("NilValidatorPanic", func() {
		logger := new(MockLogger)
		s.Panics(func() {
			NewDefaultToolRegistry(nil, logger)
		}, "Should panic with nil validator")
	})

	s.Run("NilLoggerPanic", func() {
		validator := new(MockInputValidator)
		s.Panics(func() {
			NewDefaultToolRegistry(validator, nil)
		}, "Should panic with nil logger")
	})
}

func (s *RegistryTestSuite) TestRegisterTool() {
	tests := []struct {
		name        string
		tool        *MCPTool
		expectError bool
		errorMsg    string
		setupMocks  func()
	}{
		{
			name: "ValidToolRegistration",
			tool: &MCPTool{
				Name:        "test_tool",
				Description: "Test tool for registration",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"param": map[string]interface{}{
							"type": "string",
						},
					},
				},
				Category:   CategoryConfiguration,
				CLICommand: "go-invoice",
				CLIArgs:    []string{"test"},
				Version:    "1.0.0",
				Timeout:    30 * time.Second,
			},
			expectError: false,
			setupMocks: func() {
				s.logger.On("Info", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
			},
		},
		{
			name:        "NilTool",
			tool:        nil,
			expectError: true,
			errorMsg:    "tool cannot be nil",
			setupMocks:  func() {},
		},
		{
			name: "EmptyToolName",
			tool: &MCPTool{
				Name:        "",
				Description: "Tool with empty name",
				InputSchema: map[string]interface{}{"type": "object"},
				Category:    CategoryConfiguration,
				CLICommand:  "test",
				Version:     "1.0.0",
				Timeout:     10 * time.Second,
			},
			expectError: true,
			errorMsg:    "tool name cannot be empty",
			setupMocks: func() {
				s.logger.On("Error", mock.Anything, mock.Anything).Once()
			},
		},
		{
			name: "EmptyDescription",
			tool: &MCPTool{
				Name:        "test_tool",
				Description: "",
				InputSchema: map[string]interface{}{"type": "object"},
				Category:    CategoryConfiguration,
				CLICommand:  "test",
				Version:     "1.0.0",
				Timeout:     10 * time.Second,
			},
			expectError: true,
			errorMsg:    "tool description cannot be empty",
			setupMocks: func() {
				s.logger.On("Error", mock.Anything, mock.Anything).Once()
			},
		},
		{
			name: "NilInputSchema",
			tool: &MCPTool{
				Name:        "test_tool",
				Description: "Test tool",
				InputSchema: nil,
				Category:    CategoryConfiguration,
				CLICommand:  "test",
				Version:     "1.0.0",
				Timeout:     10 * time.Second,
			},
			expectError: true,
			errorMsg:    "tool input schema cannot be nil",
			setupMocks: func() {
				s.logger.On("Error", mock.Anything, mock.Anything).Once()
			},
		},
		{
			name: "InvalidCategory",
			tool: &MCPTool{
				Name:        "test_tool",
				Description: "Test tool",
				InputSchema: map[string]interface{}{"type": "object"},
				Category:    CategoryType("invalid_category"),
				CLICommand:  "test",
				Version:     "1.0.0",
				Timeout:     10 * time.Second,
			},
			expectError: true,
			errorMsg:    "invalid tool category",
			setupMocks: func() {
				s.logger.On("Error", mock.Anything, mock.Anything).Once()
			},
		},
		{
			name: "InvalidTimeout",
			tool: &MCPTool{
				Name:        "test_tool",
				Description: "Test tool",
				InputSchema: map[string]interface{}{"type": "object"},
				Category:    CategoryConfiguration,
				CLICommand:  "test",
				Version:     "1.0.0",
				Timeout:     0,
			},
			expectError: true,
			errorMsg:    "tool timeout must be between 1 second and 10 minutes",
			setupMocks: func() {
				s.logger.On("Error", mock.Anything, mock.Anything).Once()
			},
		},
		{
			name: "InvalidSchemaType",
			tool: &MCPTool{
				Name:        "test_tool",
				Description: "Test tool",
				InputSchema: map[string]interface{}{"type": "array"},
				Category:    CategoryConfiguration,
				CLICommand:  "test",
				Version:     "1.0.0",
				Timeout:     10 * time.Second,
			},
			expectError: true,
			errorMsg:    "tool input schema type must be 'object'",
			setupMocks: func() {
				s.logger.On("Error", mock.Anything, mock.Anything).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			err := s.registry.RegisterTool(s.ctx, tt.tool)

			if tt.expectError {
				s.Error(err, "Should return registration error")
				if tt.errorMsg != "" {
					s.Contains(err.Error(), tt.errorMsg, "Error message should contain expected text")
				}
			} else {
				s.NoError(err, "Should register tool successfully")

				// Verify tool was added to registry
				tools, err := s.registry.ListTools(s.ctx, "")
				s.NoError(err)
				s.Len(tools, 1, "Should have one registered tool")
				s.Equal(tt.tool.Name, tools[0].Name)

				// Verify category mapping was updated
				categories, err := s.registry.GetCategories(s.ctx)
				s.NoError(err)
				s.Contains(categories, tt.tool.Category)
			}
		})
	}
}

func (s *RegistryTestSuite) TestRegisterTool_DuplicateName() {
	tool1 := &MCPTool{
		Name:        "duplicate_tool",
		Description: "First tool",
		InputSchema: map[string]interface{}{"type": "object"},
		Category:    CategoryConfiguration,
		CLICommand:  "test1",
		Version:     "1.0.0",
		Timeout:     10 * time.Second,
	}

	tool2 := &MCPTool{
		Name:        "duplicate_tool",
		Description: "Second tool",
		InputSchema: map[string]interface{}{"type": "object"},
		Category:    CategoryDataImport,
		CLICommand:  "test2",
		Version:     "1.0.0",
		Timeout:     10 * time.Second,
	}

	// Setup mocks
	s.logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()

	// Register first tool
	err := s.registry.RegisterTool(s.ctx, tool1)
	s.NoError(err, "First tool should register successfully")

	// Attempt to register duplicate
	err = s.registry.RegisterTool(s.ctx, tool2)
	s.Error(err, "Should fail to register duplicate tool name")
	s.Contains(err.Error(), "tool already registered", "Error should mention duplicate registration")
}

func (s *RegistryTestSuite) TestGetTool() {
	// Register a test tool first
	tool := &MCPTool{
		Name:        "get_test_tool",
		Description: "Tool for get testing",
		InputSchema: map[string]interface{}{"type": "object"},
		Category:    CategoryInvoiceManagement,
		CLICommand:  "test",
		Version:     "1.0.0",
		Timeout:     15 * time.Second,
	}

	s.logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
	err := s.registry.RegisterTool(s.ctx, tool)
	s.Require().NoError(err)

	tests := []struct {
		name        string
		toolName    string
		expectError bool
		errorType   string
		setupMocks  func()
	}{
		{
			name:        "GetExistingTool",
			toolName:    "get_test_tool",
			expectError: false,
			setupMocks: func() {
				s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
			},
		},
		{
			name:        "GetNonExistentTool",
			toolName:    "nonexistent_tool",
			expectError: true,
			errorType:   "*tools.ToolNotFoundError",
			setupMocks: func() {
				s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
			},
		},
		{
			name:        "GetToolWithSimilarName",
			toolName:    "get_test",
			expectError: true,
			errorType:   "*tools.ToolNotFoundError",
			setupMocks: func() {
				s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			retrievedTool, err := s.registry.GetTool(s.ctx, tt.toolName)

			if tt.expectError {
				s.Error(err, "Should return error for non-existent tool")
				s.Nil(retrievedTool, "Should not return tool on error")

				if tt.errorType != "" {
					s.IsType(&ToolNotFoundError{}, err, "Should return ToolNotFoundError")
					toolErr := err.(*ToolNotFoundError)
					s.Equal(tt.toolName, toolErr.ToolName)
				}
			} else {
				s.NoError(err, "Should retrieve tool successfully")
				s.NotNil(retrievedTool, "Should return tool")
				s.Equal(tool.Name, retrievedTool.Name)
				s.Equal(tool.Description, retrievedTool.Description)
				s.Equal(tool.Category, retrievedTool.Category)

				// Verify it's a defensive copy
				s.NotSame(tool, retrievedTool, "Should return defensive copy")
			}
		})
	}
}

func (s *RegistryTestSuite) TestListTools() {
	// Register multiple tools for testing
	tools := []*MCPTool{
		{
			Name:        "invoice_tool",
			Description: "Invoice tool",
			InputSchema: map[string]interface{}{"type": "object"},
			Category:    CategoryInvoiceManagement,
			CLICommand:  "invoice",
			Version:     "1.0.0",
			Timeout:     10 * time.Second,
		},
		{
			Name:        "client_tool",
			Description: "Client tool",
			InputSchema: map[string]interface{}{"type": "object"},
			Category:    CategoryClientManagement,
			CLICommand:  "client",
			Version:     "1.0.0",
			Timeout:     10 * time.Second,
		},
		{
			Name:        "config_tool",
			Description: "Config tool",
			InputSchema: map[string]interface{}{"type": "object"},
			Category:    CategoryConfiguration,
			CLICommand:  "config",
			Version:     "1.0.0",
			Timeout:     10 * time.Second,
		},
		{
			Name:        "another_invoice_tool",
			Description: "Another invoice tool",
			InputSchema: map[string]interface{}{"type": "object"},
			Category:    CategoryInvoiceManagement,
			CLICommand:  "invoice2",
			Version:     "1.0.0",
			Timeout:     10 * time.Second,
		},
	}

	// Register all tools
	s.logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Times(len(tools))
	for _, tool := range tools {
		err := s.registry.RegisterTool(s.ctx, tool)
		s.Require().NoError(err)
	}

	tests := []struct {
		name          string
		category      CategoryType
		expectedCount int
		expectedNames []string
		setupMocks    func()
	}{
		{
			name:          "ListAllTools",
			category:      "",
			expectedCount: 4,
			expectedNames: []string{"another_invoice_tool", "client_tool", "config_tool", "invoice_tool"},
			setupMocks: func() {
				s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
			},
		},
		{
			name:          "ListInvoiceManagementTools",
			category:      CategoryInvoiceManagement,
			expectedCount: 2,
			expectedNames: []string{"another_invoice_tool", "invoice_tool"},
			setupMocks: func() {
				s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
			},
		},
		{
			name:          "ListClientManagementTools",
			category:      CategoryClientManagement,
			expectedCount: 1,
			expectedNames: []string{"client_tool"},
			setupMocks: func() {
				s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
			},
		},
		{
			name:          "ListNonExistentCategory",
			category:      CategoryReporting,
			expectedCount: 0,
			expectedNames: []string{},
			setupMocks: func() {
				s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			listedTools, err := s.registry.ListTools(s.ctx, tt.category)

			s.NoError(err, "Should list tools successfully")
			s.Len(listedTools, tt.expectedCount, "Should return expected number of tools")

			// Verify tools are sorted by name
			toolNames := make([]string, len(listedTools))
			for i, tool := range listedTools {
				toolNames[i] = tool.Name
			}
			s.Equal(tt.expectedNames, toolNames, "Tools should be sorted by name")

			// Verify defensive copies
			if len(listedTools) > 0 {
				originalTool, err := s.registry.GetTool(s.ctx, listedTools[0].Name)
				s.NoError(err)
				s.NotSame(originalTool, listedTools[0], "Should return defensive copies")
			}
		})
	}
}

func (s *RegistryTestSuite) TestValidateToolInput() {
	// Register a test tool first
	tool := &MCPTool{
		Name:        "validation_test_tool",
		Description: "Tool for validation testing",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
				"amount": map[string]interface{}{
					"type": "number",
				},
			},
			"required": []interface{}{"name"},
		},
		Category:   CategoryConfiguration,
		CLICommand: "test",
		Version:    "1.0.0",
		Timeout:    10 * time.Second,
	}

	s.logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
	err := s.registry.RegisterTool(s.ctx, tool)
	s.Require().NoError(err)

	tests := []struct {
		name        string
		toolName    string
		input       map[string]interface{}
		expectError bool
		setupMocks  func()
	}{
		{
			name:     "ValidInput",
			toolName: "validation_test_tool",
			input: map[string]interface{}{
				"name":   "Test Name",
				"amount": 100.0,
			},
			expectError: false,
			setupMocks: func() {
				s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice() // GetTool + successful validation
				s.validator.On("ValidateAgainstSchema", s.ctx, mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
		{
			name:     "InvalidInput",
			toolName: "validation_test_tool",
			input: map[string]interface{}{
				"amount": 100.0,
			},
			expectError: true,
			setupMocks: func() {
				s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Times(3) // GetTool + failed validation + error log
				validationErr := &ValidationError{Field: "name", Message: "required", Code: "required_field"}
				s.validator.On("ValidateAgainstSchema", s.ctx, mock.Anything, mock.Anything).Return(validationErr).Once()
			},
		},
		{
			name:        "NonExistentTool",
			toolName:    "nonexistent_tool",
			input:       map[string]interface{}{},
			expectError: true,
			setupMocks: func() {
				s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			err := s.registry.ValidateToolInput(s.ctx, tt.toolName, tt.input)

			if tt.expectError {
				s.Error(err, "Should return validation error")
			} else {
				s.NoError(err, "Should pass validation")
			}
		})
	}
}

func (s *RegistryTestSuite) TestGetCategories() {
	// Start with empty registry
	s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
	categories, err := s.registry.GetCategories(s.ctx)
	s.NoError(err)
	s.Empty(categories, "Should return empty categories for empty registry")

	// Register tools in different categories
	tools := []*MCPTool{
		{
			Name:        "tool1",
			Description: "Tool 1",
			InputSchema: map[string]interface{}{"type": "object"},
			Category:    CategoryInvoiceManagement,
			CLICommand:  "test1",
			Version:     "1.0.0",
			Timeout:     10 * time.Second,
		},
		{
			Name:        "tool2",
			Description: "Tool 2",
			InputSchema: map[string]interface{}{"type": "object"},
			Category:    CategoryClientManagement,
			CLICommand:  "test2",
			Version:     "1.0.0",
			Timeout:     10 * time.Second,
		},
		{
			Name:        "tool3",
			Description: "Tool 3",
			InputSchema: map[string]interface{}{"type": "object"},
			Category:    CategoryInvoiceManagement, // Same category as tool1
			CLICommand:  "test3",
			Version:     "1.0.0",
			Timeout:     10 * time.Second,
		},
	}

	s.logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Times(len(tools))
	for _, tool := range tools {
		err := s.registry.RegisterTool(s.ctx, tool)
		s.Require().NoError(err)
	}

	s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
	categories, err = s.registry.GetCategories(s.ctx)
	s.NoError(err)
	s.Len(categories, 2, "Should return 2 unique categories")

	// Verify categories are sorted
	expectedCategories := []CategoryType{CategoryClientManagement, CategoryInvoiceManagement}
	s.Equal(expectedCategories, categories, "Categories should be sorted alphabetically")
}

func (s *RegistryTestSuite) TestContextCancellation() {
	tests := []struct {
		name     string
		testFunc func(context.Context) error
	}{
		{
			name: "RegisterToolCancellation",
			testFunc: func(ctx context.Context) error {
				tool := &MCPTool{
					Name:        "cancel_test",
					Description: "Test",
					InputSchema: map[string]interface{}{"type": "object"},
					Category:    CategoryConfiguration,
					CLICommand:  "test",
					Version:     "1.0.0",
					Timeout:     10 * time.Second,
				}
				return s.registry.RegisterTool(ctx, tool)
			},
		},
		{
			name: "GetToolCancellation",
			testFunc: func(ctx context.Context) error {
				_, err := s.registry.GetTool(ctx, "test_tool")
				return err
			},
		},
		{
			name: "ListToolsCancellation",
			testFunc: func(ctx context.Context) error {
				_, err := s.registry.ListTools(ctx, "")
				return err
			},
		},
		{
			name: "ValidateToolInputCancellation",
			testFunc: func(ctx context.Context) error {
				return s.registry.ValidateToolInput(ctx, "test_tool", map[string]interface{}{})
			},
		},
		{
			name: "GetCategoriesCancellation",
			testFunc: func(ctx context.Context) error {
				_, err := s.registry.GetCategories(ctx)
				return err
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx, cancel := context.WithCancel(s.ctx)
			cancel() // Cancel immediately

			err := tt.testFunc(ctx)
			s.Error(err, "Should return context cancellation error")
			s.Equal(context.Canceled, err, "Should be context.Canceled error")
		})
	}
}

func (s *RegistryTestSuite) TestConcurrentOperations() {
	s.Run("ConcurrentRegistration", func() {
		var wg sync.WaitGroup
		errors := make(chan error, 100)

		// Setup mocks for concurrent registrations
		s.logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Times(100)
		s.logger.On("Error", mock.Anything, mock.Anything).Maybe() // Some may fail due to duplicates

		// Attempt to register 100 tools concurrently
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				tool := &MCPTool{
					Name:        fmt.Sprintf("concurrent_tool_%d", id),
					Description: fmt.Sprintf("Concurrent tool %d", id),
					InputSchema: map[string]interface{}{"type": "object"},
					Category:    CategoryConfiguration,
					CLICommand:  "test",
					Version:     "1.0.0",
					Timeout:     10 * time.Second,
				}
				err := s.registry.RegisterTool(s.ctx, tool)
				if err != nil {
					errors <- err
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for unexpected errors (should be none for unique tool names)
		var errorCount int
		for err := range errors {
			s.T().Logf("Unexpected error: %v", err)
			errorCount++
		}
		s.Equal(0, errorCount, "Should have no errors for unique tool names")

		// Verify all tools were registered
		tools, err := s.registry.ListTools(s.ctx, "")
		s.NoError(err)
		s.Len(tools, 100, "Should have 100 registered tools")
	})

	s.Run("ConcurrentGetOperations", func() {
		// Register a tool first
		tool := &MCPTool{
			Name:        "concurrent_get_test",
			Description: "Test tool for concurrent gets",
			InputSchema: map[string]interface{}{"type": "object"},
			Category:    CategoryConfiguration,
			CLICommand:  "test",
			Version:     "1.0.0",
			Timeout:     10 * time.Second,
		}

		s.logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
		err := s.registry.RegisterTool(s.ctx, tool)
		s.Require().NoError(err)

		var wg sync.WaitGroup
		errors := make(chan error, 100)

		// Setup mocks for concurrent gets
		s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Times(100)

		// Perform 100 concurrent get operations
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				retrievedTool, err := s.registry.GetTool(s.ctx, "concurrent_get_test")
				if err != nil {
					errors <- err
				} else if retrievedTool.Name != "concurrent_get_test" {
					errors <- fmt.Errorf("unexpected tool name: %s", retrievedTool.Name)
				}
			}()
		}

		wg.Wait()
		close(errors)

		// Check for errors
		var errorCount int
		for err := range errors {
			s.T().Logf("Unexpected error: %v", err)
			errorCount++
		}
		s.Equal(0, errorCount, "Should have no errors for concurrent get operations")
	})
}

func (s *RegistryTestSuite) TestToolDefensiveCopying() {
	originalTool := &MCPTool{
		Name:        "copy_test_tool",
		Description: "Tool for testing defensive copying",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"param": map[string]interface{}{
					"type": "string",
				},
			},
		},
		Examples: []MCPToolExample{
			{
				Description: "Test example",
				Input:       map[string]interface{}{"param": "value"},
			},
		},
		Category:   CategoryConfiguration,
		CLICommand: "test",
		CLIArgs:    []string{"arg1", "arg2"},
		Version:    "1.0.0",
		Timeout:    15 * time.Second,
	}

	s.logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
	err := s.registry.RegisterTool(s.ctx, originalTool)
	s.Require().NoError(err)

	// Modify original tool after registration
	originalTool.Description = "Modified description"
	originalTool.Examples[0].Description = "Modified example"
	originalTool.CLIArgs[0] = "modified_arg"

	// Retrieve tool and verify it wasn't affected by modifications
	s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
	retrievedTool, err := s.registry.GetTool(s.ctx, "copy_test_tool")
	s.NoError(err)

	s.Equal("Tool for testing defensive copying", retrievedTool.Description, "Description should not be modified")
	s.Equal("Test example", retrievedTool.Examples[0].Description, "Example should not be modified")
	s.Equal("arg1", retrievedTool.CLIArgs[0], "CLI args should not be modified")
}

func (s *RegistryTestSuite) TestValidationIntegration() {
	// Register a tool with complex schema
	tool := &MCPTool{
		Name:        "complex_validation_tool",
		Description: "Tool with complex validation schema",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"client_name": map[string]interface{}{
					"type":      "string",
					"minLength": 1.0,
					"maxLength": 100.0,
				},
				"amount": map[string]interface{}{
					"type":    "number",
					"minimum": 0.01,
					"maximum": 1000000.0,
				},
				"email": map[string]interface{}{
					"type":   "string",
					"format": "email",
				},
			},
			"required": []interface{}{"client_name", "amount"},
		},
		Category:   CategoryInvoiceManagement,
		CLICommand: "test",
		Version:    "1.0.0",
		Timeout:    20 * time.Second,
	}

	s.logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
	err := s.registry.RegisterTool(s.ctx, tool)
	s.Require().NoError(err)

	tests := []struct {
		name        string
		input       map[string]interface{}
		expectError bool
		setupMocks  func()
	}{
		{
			name: "ValidComplexInput",
			input: map[string]interface{}{
				"client_name": "Acme Corporation",
				"amount":      500.75,
				"email":       "contact@acme.com",
			},
			expectError: false,
			setupMocks: func() {
				s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice()
				s.validator.On("ValidateAgainstSchema", s.ctx, mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
		{
			name: "InvalidComplexInput",
			input: map[string]interface{}{
				"client_name": "",              // Empty required field
				"amount":      -100.0,          // Negative amount
				"email":       "invalid-email", // Invalid email format
			},
			expectError: true,
			setupMocks: func() {
				s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Times(3)
				validationErr := &ValidationError{
					Field:   "client_name",
					Message: "field cannot be empty",
					Code:    "empty_field",
				}
				s.validator.On("ValidateAgainstSchema", s.ctx, mock.Anything, mock.Anything).Return(validationErr).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			err := s.registry.ValidateToolInput(s.ctx, "complex_validation_tool", tt.input)

			if tt.expectError {
				s.Error(err, "Should return validation error")
			} else {
				s.NoError(err, "Should pass validation")
			}
		})
	}
}

// TestRegistryTestSuite runs the complete registry test suite
func TestRegistryTestSuite(t *testing.T) {
	suite.Run(t, new(RegistryTestSuite))
}

// Benchmark tests for performance validation
func BenchmarkDefaultToolRegistry_GetTool(b *testing.B) {
	validator := new(MockInputValidator)
	logger := new(MockLogger)
	logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	registry := NewDefaultToolRegistry(validator, logger)
	ctx := context.Background()

	// Register a test tool
	tool := &MCPTool{
		Name:        "benchmark_tool",
		Description: "Benchmark test tool",
		InputSchema: map[string]interface{}{"type": "object"},
		Category:    CategoryConfiguration,
		CLICommand:  "test",
		Version:     "1.0.0",
		Timeout:     10 * time.Second,
	}

	_ = registry.RegisterTool(ctx, tool)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.GetTool(ctx, "benchmark_tool")
	}
}

func BenchmarkDefaultToolRegistry_ListTools(b *testing.B) {
	validator := new(MockInputValidator)
	logger := new(MockLogger)
	logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	registry := NewDefaultToolRegistry(validator, logger)
	ctx := context.Background()

	// Register multiple tools
	for i := 0; i < 100; i++ {
		tool := &MCPTool{
			Name:        fmt.Sprintf("tool_%d", i),
			Description: fmt.Sprintf("Tool %d", i),
			InputSchema: map[string]interface{}{"type": "object"},
			Category:    CategoryConfiguration,
			CLICommand:  "test",
			Version:     "1.0.0",
			Timeout:     10 * time.Second,
		}
		_ = registry.RegisterTool(ctx, tool)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.ListTools(ctx, "")
	}
}

// Unit tests for specific edge cases
func TestDefaultToolRegistry_EdgeCases(t *testing.T) {
	validator := new(MockInputValidator)
	logger := new(MockLogger)
	registry := NewDefaultToolRegistry(validator, logger)
	ctx := context.Background()

	t.Run("EmptyToolRegistry", func(t *testing.T) {
		logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
		tools, err := registry.ListTools(ctx, CategoryInvoiceManagement)
		assert.NoError(t, err)
		assert.Empty(t, tools, "Should return empty slice for non-existent category")
	})

	t.Run("GetToolFromEmptyRegistry", func(t *testing.T) {
		logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
		tool, err := registry.GetTool(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, tool)
		assert.IsType(t, &ToolNotFoundError{}, err)
	})

	t.Run("ValidateInputForNonExistentTool", func(t *testing.T) {
		logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
		err := registry.ValidateToolInput(ctx, "nonexistent", map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot validate input for unknown tool")
	})

	t.Run("GetCategoriesFromEmptyRegistry", func(t *testing.T) {
		logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
		categories, err := registry.GetCategories(ctx)
		assert.NoError(t, err)
		assert.Empty(t, categories)
	})
}

func TestToolValidation_EdgeCases(t *testing.T) {
	validator := new(MockInputValidator)
	logger := new(MockLogger)
	registry := NewDefaultToolRegistry(validator, logger)

	t.Run("ToolWithExcessiveTimeout", func(t *testing.T) {
		tool := &MCPTool{
			Name:        "excessive_timeout_tool",
			Description: "Tool with excessive timeout",
			InputSchema: map[string]interface{}{"type": "object"},
			Category:    CategoryConfiguration,
			CLICommand:  "test",
			Version:     "1.0.0",
			Timeout:     24 * time.Hour, // Excessive timeout
		}

		logger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
		err := registry.RegisterTool(context.Background(), tool)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tool timeout must be between 1 second and 10 minutes")
	})

	t.Run("ToolWithEmptyVersion", func(t *testing.T) {
		tool := &MCPTool{
			Name:        "no_version_tool",
			Description: "Tool without version",
			InputSchema: map[string]interface{}{"type": "object"},
			Category:    CategoryConfiguration,
			CLICommand:  "test",
			Version:     "", // Empty version
			Timeout:     10 * time.Second,
		}

		logger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
		err := registry.RegisterTool(context.Background(), tool)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tool version cannot be empty")
	})

	t.Run("ToolWithEmptyCLICommand", func(t *testing.T) {
		tool := &MCPTool{
			Name:        "no_cli_tool",
			Description: "Tool without CLI command",
			InputSchema: map[string]interface{}{"type": "object"},
			Category:    CategoryConfiguration,
			CLICommand:  "", // Empty CLI command
			Version:     "1.0.0",
			Timeout:     10 * time.Second,
		}

		logger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
		err := registry.RegisterTool(context.Background(), tool)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tool CLI command cannot be empty")
	})
}
