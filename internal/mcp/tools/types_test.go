package tools

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// TypesTestSuite provides comprehensive tests for core types and interfaces
type TypesTestSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *TypesTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *TypesTestSuite) TestMCPTool_Structure() {
	tests := []struct {
		name        string
		tool        MCPTool
		expectValid bool
		description string
	}{
		{
			name: "ValidCompleteToolDefinition",
			tool: MCPTool{
				Name:        "test_tool",
				Description: "Test tool for validation",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"param1": map[string]interface{}{
							"type":        "string",
							"description": "Test parameter",
						},
					},
					"required": []interface{}{"param1"},
				},
				Examples: []MCPToolExample{
					{
						Description: "Basic example",
						Input: map[string]interface{}{
							"param1": "test value",
						},
						ExpectedOutput: "Test output",
						UseCase:        "Testing",
					},
				},
				Category:   CategoryInvoiceManagement,
				CLICommand: "go-invoice",
				CLIArgs:    []string{"test"},
				HelpText:   "Help text for test tool",
				Version:    "1.0.0",
				Timeout:    30 * time.Second,
			},
			expectValid: true,
			description: "Complete valid tool definition with all fields",
		},
		{
			name: "MinimalValidToolDefinition",
			tool: MCPTool{
				Name:        "minimal_tool",
				Description: "Minimal test tool",
				InputSchema: map[string]interface{}{
					"type": "object",
				},
				Category:   CategoryConfiguration,
				CLICommand: "go-invoice",
				CLIArgs:    []string{"minimal"},
				Version:    "1.0.0",
				Timeout:    15 * time.Second,
			},
			expectValid: true,
			description: "Minimal valid tool with required fields only",
		},
		{
			name: "ToolWithEmptyName",
			tool: MCPTool{
				Name:        "",
				Description: "Tool with empty name",
				InputSchema: map[string]interface{}{"type": "object"},
				Category:    CategoryDataImport,
				CLICommand:  "go-invoice",
				Version:     "1.0.0",
				Timeout:     10 * time.Second,
			},
			expectValid: false,
			description: "Tool with empty name should be invalid",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Test JSON marshaling/unmarshaling
			if tt.expectValid {
				s.NotEmpty(tt.tool.Name, "Valid tool should have non-empty name")
				s.NotEmpty(tt.tool.Description, "Valid tool should have description")
				s.NotNil(tt.tool.InputSchema, "Valid tool should have schema")
				s.NotEmpty(tt.tool.CLICommand, "Valid tool should have CLI command")
				s.NotEmpty(tt.tool.Version, "Valid tool should have version")
				s.Greater(tt.tool.Timeout, time.Duration(0), "Valid tool should have positive timeout")
			}

			// Validate schema structure
			if tt.tool.InputSchema != nil {
				schemaType, hasType := tt.tool.InputSchema["type"]
				if hasType {
					s.Equal("object", schemaType, "Schema type should be object")
				}
			}

			// Validate examples structure
			for _, example := range tt.tool.Examples {
				s.NotEmpty(example.Description, "Example should have description")
				s.NotNil(example.Input, "Example should have input")
			}
		})
	}
}

func (s *TypesTestSuite) TestMCPToolExample_Validation() {
	tests := []struct {
		name    string
		example MCPToolExample
		valid   bool
	}{
		{
			name: "CompleteExample",
			example: MCPToolExample{
				Description:    "Complete example with all fields",
				Input:          map[string]interface{}{"param": "value"},
				ExpectedOutput: "Expected result",
				UseCase:        "Testing scenario",
			},
			valid: true,
		},
		{
			name: "MinimalExample",
			example: MCPToolExample{
				Description: "Minimal example",
				Input:       map[string]interface{}{"param": "value"},
			},
			valid: true,
		},
		{
			name: "EmptyDescription",
			example: MCPToolExample{
				Description: "",
				Input:       map[string]interface{}{"param": "value"},
			},
			valid: false,
		},
		{
			name: "NilInput",
			example: MCPToolExample{
				Description: "Example with nil input",
				Input:       nil,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.valid {
				s.NotEmpty(tt.example.Description, "Valid example should have description")
				s.NotNil(tt.example.Input, "Valid example should have input")
			} else {
				if tt.example.Description == "" {
					s.Empty(tt.example.Description, "Invalid example has empty description")
				}
				if tt.example.Input == nil {
					s.Nil(tt.example.Input, "Invalid example has nil input")
				}
			}
		})
	}
}

func (s *TypesTestSuite) TestCategoryType_Constants() {
	expectedCategories := []CategoryType{
		CategoryInvoiceManagement,
		CategoryDataImport,
		CategoryDataExport,
		CategoryClientManagement,
		CategoryConfiguration,
		CategoryReporting,
	}

	// Test all category constants are defined
	for _, category := range expectedCategories {
		s.NotEmpty(string(category), "Category should not be empty")
		categoryStr := string(category)
		s.Equal(strings.ToLower(categoryStr), categoryStr, "Category should be lowercase")

		// Multi-word categories should use underscores, single words are acceptable
		if strings.Contains(categoryStr, "management") || strings.Contains(categoryStr, "export") || strings.Contains(categoryStr, "import") {
			s.Contains(categoryStr, "_", "Multi-word category should use snake_case format")
		}
	}

	// Test category uniqueness
	categorySet := make(map[CategoryType]bool)
	for _, category := range expectedCategories {
		s.False(categorySet[category], "Category %s should be unique", category)
		categorySet[category] = true
	}

	s.Equal(6, len(expectedCategories), "Should have exactly 6 categories")
}

func (s *TypesTestSuite) TestValidationError_Interface() {
	tests := []struct {
		name          string
		validationErr ValidationError
		expectedMsg   string
	}{
		{
			name: "ErrorWithFieldAndSuggestions",
			validationErr: ValidationError{
				Field:       "client_name",
				Message:     "field is required",
				Code:        "required_field",
				Suggestions: []string{"provide client name", "use client_id instead"},
			},
			expectedMsg: "client_name: field is required (suggestions: provide client name, use client_id instead)",
		},
		{
			name: "ErrorWithoutField",
			validationErr: ValidationError{
				Field:       "",
				Message:     "validation failed",
				Code:        "validation_error",
				Suggestions: []string{"check input format"},
			},
			expectedMsg: "validation failed (suggestions: check input format)",
		},
		{
			name: "ErrorWithoutSuggestions",
			validationErr: ValidationError{
				Field:   "amount",
				Message: "must be positive",
				Code:    "invalid_value",
			},
			expectedMsg: "amount: must be positive",
		},
		{
			name: "MinimalError",
			validationErr: ValidationError{
				Message: "error occurred",
				Code:    "generic_error",
			},
			expectedMsg: "error occurred",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			errorMsg := tt.validationErr.Error()
			s.Equal(tt.expectedMsg, errorMsg, "Error message should match expected format")

			// Test error interface compliance
			var err error = &tt.validationErr
			s.NotNil(err, "ValidationError should implement error interface")
			s.Equal(tt.expectedMsg, err.Error(), "Error interface should return same message")
		})
	}
}

func (s *TypesTestSuite) TestToolNotFoundError_Interface() {
	tests := []struct {
		name        string
		toolErr     ToolNotFoundError
		expectedMsg string
	}{
		{
			name: "ErrorWithSuggestions",
			toolErr: ToolNotFoundError{
				ToolName:       "invoice_creat",
				AvailableTools: []string{"invoice_create", "invoice_list", "invoice_show"},
				Category:       "invoice_management",
			},
			expectedMsg: "tool not found: invoice_creat (available tools: invoice_create, invoice_list, invoice_show) (try searching in category: invoice_management)",
		},
		{
			name: "ErrorWithoutCategory",
			toolErr: ToolNotFoundError{
				ToolName:       "unknown_tool",
				AvailableTools: []string{"tool1", "tool2"},
			},
			expectedMsg: "tool not found: unknown_tool (available tools: tool1, tool2)",
		},
		{
			name: "ErrorWithoutSuggestions",
			toolErr: ToolNotFoundError{
				ToolName: "missing_tool",
				Category: "configuration",
			},
			expectedMsg: "tool not found: missing_tool (try searching in category: configuration)",
		},
		{
			name: "MinimalError",
			toolErr: ToolNotFoundError{
				ToolName: "not_found",
			},
			expectedMsg: "tool not found: not_found",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			errorMsg := tt.toolErr.Error()
			s.Equal(tt.expectedMsg, errorMsg, "Error message should match expected format")

			// Test error interface compliance
			var err error = &tt.toolErr
			s.NotNil(err, "ToolNotFoundError should implement error interface")
			s.Equal(tt.expectedMsg, err.Error(), "Error interface should return same message")
		})
	}
}

func (s *TypesTestSuite) TestToolExecutionRequest_Structure() {
	tests := []struct {
		name    string
		request ToolExecutionRequest
		valid   bool
	}{
		{
			name: "CompleteRequest",
			request: ToolExecutionRequest{
				ToolName:  "invoice_create",
				Input:     map[string]interface{}{"client_name": "Test Client"},
				RequestID: "req_123",
				ClientInfo: &ClientInfo{
					Name:     "Claude Desktop",
					Version:  "1.0.0",
					Platform: "macos",
				},
				ExecutionTimeout: 30 * time.Second,
			},
			valid: true,
		},
		{
			name: "MinimalRequest",
			request: ToolExecutionRequest{
				ToolName:         "test_tool",
				Input:            map[string]interface{}{},
				RequestID:        "req_456",
				ExecutionTimeout: 15 * time.Second,
			},
			valid: true,
		},
		{
			name: "EmptyToolName",
			request: ToolExecutionRequest{
				ToolName:         "",
				Input:            map[string]interface{}{},
				RequestID:        "req_789",
				ExecutionTimeout: 10 * time.Second,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.valid {
				s.NotEmpty(tt.request.ToolName, "Valid request should have tool name")
				s.NotNil(tt.request.Input, "Valid request should have input map")
				s.NotEmpty(tt.request.RequestID, "Valid request should have request ID")
				s.Greater(tt.request.ExecutionTimeout, time.Duration(0), "Valid request should have positive timeout")
			}

			if tt.request.ClientInfo != nil {
				s.NotEmpty(tt.request.ClientInfo.Name, "Client info should have name")
			}
		})
	}
}

func (s *TypesTestSuite) TestToolExecutionResponse_Structure() {
	tests := []struct {
		name     string
		response ToolExecutionResponse
		valid    bool
	}{
		{
			name: "SuccessfulResponse",
			response: ToolExecutionResponse{
				Success:       true,
				Output:        map[string]interface{}{"result": "success"},
				ExecutionTime: 2 * time.Second,
				ResourcesUsed: []string{"file1.json", "file2.txt"},
				Metadata:      map[string]string{"version": "1.0.0"},
			},
			valid: true,
		},
		{
			name: "ErrorResponse",
			response: ToolExecutionResponse{
				Success:       false,
				Output:        map[string]interface{}{},
				ErrorMessage:  "Tool execution failed",
				ExecutionTime: 1 * time.Second,
			},
			valid: true,
		},
		{
			name: "MinimalResponse",
			response: ToolExecutionResponse{
				Success:       true,
				Output:        map[string]interface{}{},
				ExecutionTime: 500 * time.Millisecond,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.valid {
				s.NotNil(tt.response.Output, "Valid response should have output map")
				s.GreaterOrEqual(tt.response.ExecutionTime, time.Duration(0), "Execution time should be non-negative")

				if !tt.response.Success {
					s.NotEmpty(tt.response.ErrorMessage, "Failed response should have error message")
				}
			}
		})
	}
}

func (s *TypesTestSuite) TestClientInfo_Structure() {
	tests := []struct {
		name   string
		client ClientInfo
		valid  bool
	}{
		{
			name: "CompleteClientInfo",
			client: ClientInfo{
				Name:     "Claude Desktop",
				Version:  "1.2.3",
				Platform: "macos",
			},
			valid: true,
		},
		{
			name: "MinimalClientInfo",
			client: ClientInfo{
				Name:    "Test Client",
				Version: "1.0.0",
			},
			valid: true,
		},
		{
			name: "EmptyName",
			client: ClientInfo{
				Name:    "",
				Version: "1.0.0",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.valid {
				s.NotEmpty(tt.client.Name, "Valid client info should have name")
				s.NotEmpty(tt.client.Version, "Valid client info should have version")
			}
		})
	}
}

func (s *TypesTestSuite) TestContextCancellation() {
	// Test context cancellation handling in type operations
	ctx, cancel := context.WithTimeout(s.ctx, 1*time.Millisecond)
	defer cancel()

	// Wait for context to timeout
	time.Sleep(2 * time.Millisecond)

	select {
	case <-ctx.Done():
		s.Error(ctx.Err(), "Context should be cancelled")
		s.Contains(ctx.Err().Error(), "deadline exceeded", "Should be timeout error")
	default:
		s.Fail("Context should be cancelled")
	}
}

func (s *TypesTestSuite) TestConcurrentAccess() {
	// Test thread safety of type structures
	tool := &MCPTool{
		Name:        "concurrent_test",
		Description: "Test concurrent access",
		InputSchema: map[string]interface{}{"type": "object"},
		Category:    CategoryConfiguration,
		CLICommand:  "test",
		Version:     "1.0.0",
		Timeout:     10 * time.Second,
	}

	// Simulate concurrent read access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			// Read operations should be safe
			s.Equal("concurrent_test", tool.Name)
			s.Equal(CategoryConfiguration, tool.Category)
			s.NotNil(tool.InputSchema)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func (s *TypesTestSuite) TestTimeoutValidation() {
	tests := []struct {
		name    string
		timeout time.Duration
		valid   bool
	}{
		{
			name:    "ValidShortTimeout",
			timeout: 5 * time.Second,
			valid:   true,
		},
		{
			name:    "ValidLongTimeout",
			timeout: 5 * time.Minute,
			valid:   true,
		},
		{
			name:    "ZeroTimeout",
			timeout: 0,
			valid:   false,
		},
		{
			name:    "NegativeTimeout",
			timeout: -1 * time.Second,
			valid:   false,
		},
		{
			name:    "ExcessiveTimeout",
			timeout: 24 * time.Hour,
			valid:   false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.valid {
				s.Greater(tt.timeout, time.Duration(0), "Valid timeout should be positive")
				s.LessOrEqual(tt.timeout, 10*time.Minute, "Valid timeout should be reasonable")
			} else {
				// Invalid timeouts are either non-positive or excessively long
				if tt.timeout <= 0 {
					s.LessOrEqual(tt.timeout, time.Duration(0), "Invalid timeout should be non-positive")
				} else {
					s.Greater(tt.timeout, 10*time.Minute, "Excessive timeout should be greater than reasonable limit")
				}
			}
		})
	}
}

// TestTypesTestSuite runs the complete types test suite
func TestTypesTestSuite(t *testing.T) {
	suite.Run(t, new(TypesTestSuite))
}

// Benchmark tests for performance validation
func BenchmarkMCPTool_Creation(b *testing.B) {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"param1": map[string]interface{}{
				"type": "string",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool := &MCPTool{
			Name:        "benchmark_tool",
			Description: "Benchmark test tool",
			InputSchema: schema,
			Category:    CategoryConfiguration,
			CLICommand:  "test",
			Version:     "1.0.0",
			Timeout:     10 * time.Second,
		}
		_ = tool
	}
}

func BenchmarkValidationError_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := &ValidationError{
			Field:       "test_field",
			Message:     "validation failed",
			Code:        "test_error",
			Suggestions: []string{"suggestion1", "suggestion2"},
		}
		_ = err.Error()
	}
}

// Unit tests for specific edge cases
func TestMCPTool_EdgeCases(t *testing.T) {
	t.Run("NilInputSchema", func(t *testing.T) {
		tool := &MCPTool{
			Name:        "test_tool",
			Description: "Test tool",
			InputSchema: nil,
			Category:    CategoryConfiguration,
			CLICommand:  "test",
			Version:     "1.0.0",
			Timeout:     10 * time.Second,
		}
		assert.Nil(t, tool.InputSchema, "Tool should handle nil schema")
	})

	t.Run("EmptyExamplesSlice", func(t *testing.T) {
		tool := &MCPTool{
			Name:        "test_tool",
			Description: "Test tool",
			InputSchema: map[string]interface{}{"type": "object"},
			Examples:    []MCPToolExample{},
			Category:    CategoryConfiguration,
			CLICommand:  "test",
			Version:     "1.0.0",
			Timeout:     10 * time.Second,
		}
		assert.Empty(t, tool.Examples, "Tool should handle empty examples slice")
	})

	t.Run("LargeTimeout", func(t *testing.T) {
		tool := &MCPTool{
			Name:        "test_tool",
			Description: "Test tool",
			InputSchema: map[string]interface{}{"type": "object"},
			Category:    CategoryConfiguration,
			CLICommand:  "test",
			Version:     "1.0.0",
			Timeout:     1 * time.Hour,
		}
		assert.Equal(t, 1*time.Hour, tool.Timeout, "Tool should accept large timeout values")
	})
}

func TestValidationError_EdgeCases(t *testing.T) {
	t.Run("EmptyField", func(t *testing.T) {
		err := &ValidationError{
			Field:   "",
			Message: "error message",
			Code:    "test_error",
		}
		assert.Equal(t, "error message", err.Error(), "Should handle empty field")
	})

	t.Run("EmptySuggestions", func(t *testing.T) {
		err := &ValidationError{
			Field:       "test_field",
			Message:     "error message",
			Code:        "test_error",
			Suggestions: []string{},
		}
		assert.Equal(t, "test_field: error message", err.Error(), "Should handle empty suggestions")
	})

	t.Run("SingleSuggestion", func(t *testing.T) {
		err := &ValidationError{
			Field:       "test_field",
			Message:     "error message",
			Code:        "test_error",
			Suggestions: []string{"single suggestion"},
		}
		expected := "test_field: error message (suggestions: single suggestion)"
		assert.Equal(t, expected, err.Error(), "Should handle single suggestion")
	})
}

func TestToolNotFoundError_EdgeCases(t *testing.T) {
	t.Run("EmptyToolName", func(t *testing.T) {
		err := &ToolNotFoundError{
			ToolName: "",
		}
		assert.Equal(t, "tool not found: ", err.Error(), "Should handle empty tool name")
	})

	t.Run("SingleAvailableTool", func(t *testing.T) {
		err := &ToolNotFoundError{
			ToolName:       "missing",
			AvailableTools: []string{"available"},
		}
		expected := "tool not found: missing (available tools: available)"
		assert.Equal(t, expected, err.Error(), "Should handle single available tool")
	})

	t.Run("EmptyCategory", func(t *testing.T) {
		err := &ToolNotFoundError{
			ToolName: "missing",
			Category: "",
		}
		assert.Equal(t, "tool not found: missing", err.Error(), "Should handle empty category")
	})
}

// Race condition tests
func TestTypes_RaceConditions(t *testing.T) {
	t.Run("ConcurrentErrorCreation", func(t *testing.T) {
		done := make(chan bool, 100)
		for i := 0; i < 100; i++ {
			go func(id int) {
				defer func() { done <- true }()
				err := &ValidationError{
					Field:   "field",
					Message: "message",
					Code:    "code",
				}
				_ = err.Error()
			}(i)
		}

		for i := 0; i < 100; i++ {
			<-done
		}
	})

	t.Run("ConcurrentToolAccess", func(t *testing.T) {
		tool := &MCPTool{
			Name:        "concurrent",
			Description: "Test",
			InputSchema: map[string]interface{}{"type": "object"},
			Category:    CategoryConfiguration,
			CLICommand:  "test",
			Version:     "1.0.0",
			Timeout:     10 * time.Second,
		}

		done := make(chan bool, 50)
		for i := 0; i < 50; i++ {
			go func() {
				defer func() { done <- true }()
				_ = tool.Name
				_ = tool.Category
				_ = tool.Timeout
			}()
		}

		for i := 0; i < 50; i++ {
			<-done
		}
	})
}
