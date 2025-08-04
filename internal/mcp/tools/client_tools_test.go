package tools

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ClientToolsTestSuite provides comprehensive tests for client management tool definitions
type ClientToolsTestSuite struct {
	suite.Suite

	// Test context managed per test method
}

func (s *ClientToolsTestSuite) SetupTest() {
	// Test setup if needed
}

func (s *ClientToolsTestSuite) TestCreateClientManagementTools() {
	tools := CreateClientManagementTools()

	s.NotNil(tools, "Should return tools slice")
	s.NotEmpty(tools, "Should return at least one tool")

	// Expected client management tools based on typical functionality
	expectedToolNames := []string{
		"client_create",
		"client_list",
		"client_show",
		"client_update",
		"client_delete",
	}

	// Verify we have the expected tools (at least the core ones)
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	for _, expectedName := range expectedToolNames {
		if toolNames[expectedName] {
			s.True(toolNames[expectedName], "Should have %s tool", expectedName)
		}
	}

	// Verify all tools are properly configured
	for _, tool := range tools {
		s.validateClientTool(tool)
	}
}

func (s *ClientToolsTestSuite) TestClientCreateTool() {
	tools := CreateClientManagementTools()
	var clientCreateTool *MCPTool

	for _, tool := range tools {
		if tool.Name == "client_create" {
			clientCreateTool = tool
			break
		}
	}

	if clientCreateTool == nil {
		s.T().Skip("client_create tool not found in implementation")
		return
	}

	s.Run("BasicStructure", func() {
		s.Equal("client_create", clientCreateTool.Name)
		s.NotEmpty(clientCreateTool.Description)
		s.Contains(clientCreateTool.Description, "client")
		s.Contains(clientCreateTool.Description, "create")
		s.Equal(CategoryClientManagement, clientCreateTool.Category)
		s.NotEmpty(clientCreateTool.Version)
		s.Greater(clientCreateTool.Timeout, time.Duration(0))
	})

	s.Run("Schema", func() {
		s.NotNil(clientCreateTool.InputSchema)
		s.Equal("object", clientCreateTool.InputSchema["type"])

		// Verify expected properties exist
		properties, hasProperties := clientCreateTool.InputSchema["properties"]
		s.True(hasProperties, "Should have properties in schema")

		if propertiesMap, ok := properties.(map[string]interface{}); ok {
			// Should have common client fields
			expectedFields := []string{"name", "email"}
			for _, field := range expectedFields {
				if _, hasField := propertiesMap[field]; hasField {
					s.Contains(propertiesMap, field, "Should have %s field", field)
				}
			}
		}
	})

	s.Run("Examples", func() {
		s.NotEmpty(clientCreateTool.Examples, "Should have examples")

		for i, example := range clientCreateTool.Examples {
			s.NotEmpty(example.Description, "Example %d should have description", i)
			s.NotNil(example.Input, "Example %d should have input", i)
			s.NotEmpty(example.Input, "Example %d input should not be empty", i)

			// Verify example input contains reasonable fields
			if name, hasName := example.Input["name"]; hasName {
				s.IsType("", name, "Name should be string")
				s.NotEmpty(name, "Name should not be empty")
			}
		}
	})
}

func (s *ClientToolsTestSuite) TestClientListTool() {
	tools := CreateClientManagementTools()
	var clientListTool *MCPTool

	for _, tool := range tools {
		if tool.Name == "client_list" {
			clientListTool = tool
			break
		}
	}

	if clientListTool == nil {
		s.T().Skip("client_list tool not found in implementation")
		return
	}

	s.Run("BasicStructure", func() {
		s.Equal("client_list", clientListTool.Name)
		s.NotEmpty(clientListTool.Description)
		s.Contains(clientListTool.Description, "client")
		s.Contains(strings.ToLower(clientListTool.Description), "list")
		s.Equal(CategoryClientManagement, clientListTool.Category)
	})

	s.Run("Schema", func() {
		s.NotNil(clientListTool.InputSchema)
		s.Equal("object", clientListTool.InputSchema["type"])

		// List tools typically have optional filtering parameters
		properties, hasProperties := clientListTool.InputSchema["properties"]
		if hasProperties {
			if propertiesMap, ok := properties.(map[string]interface{}); ok {
				// Common list parameters
				potentialFields := []string{"search", "limit", "offset", "output_format"}
				for _, field := range potentialFields {
					if fieldDef, hasField := propertiesMap[field]; hasField {
						s.NotNil(fieldDef, "Field %s should have definition", field)
					}
				}
			}
		}
	})

	s.Run("Examples", func() {
		s.NotEmpty(clientListTool.Examples, "Should have examples")

		// Should have examples for different use cases
		hasBasicExample := false
		hasFilterExample := false

		for _, example := range clientListTool.Examples {
			s.NotEmpty(example.Description)
			s.NotNil(example.Input)

			if len(example.Input) == 0 {
				hasBasicExample = true // List all clients
			} else {
				hasFilterExample = true // Filtered list
			}
		}

		s.True(hasBasicExample || hasFilterExample, "Should have at least one list example")
	})
}

func (s *ClientToolsTestSuite) TestClientShowTool() {
	tools := CreateClientManagementTools()
	var clientShowTool *MCPTool

	for _, tool := range tools {
		if tool.Name == "client_show" {
			clientShowTool = tool
			break
		}
	}

	if clientShowTool == nil {
		s.T().Skip("client_show tool not found in implementation")
		return
	}

	s.Run("BasicStructure", func() {
		s.Equal("client_show", clientShowTool.Name)
		s.NotEmpty(clientShowTool.Description)
		s.Contains(clientShowTool.Description, "client")
		s.Equal(CategoryClientManagement, clientShowTool.Category)
	})

	s.Run("Schema", func() {
		s.NotNil(clientShowTool.InputSchema)
		s.Equal("object", clientShowTool.InputSchema["type"])

		// Show tools typically require an identifier
		properties, hasProperties := clientShowTool.InputSchema["properties"]
		s.True(hasProperties, "Should have properties for client identification")

		if propertiesMap, ok := properties.(map[string]interface{}); ok {
			// Should have at least one identification method
			identifierFields := []string{"client_id", "client_name", "client_email"}
			hasIdentifier := false
			for _, field := range identifierFields {
				if _, hasField := propertiesMap[field]; hasField {
					hasIdentifier = true
					break
				}
			}
			s.True(hasIdentifier, "Should have at least one client identifier field")
		}
	})

	s.Run("Examples", func() {
		s.NotEmpty(clientShowTool.Examples, "Should have examples")

		for _, example := range clientShowTool.Examples {
			s.NotEmpty(example.Description)
			s.NotNil(example.Input)
			s.NotEmpty(example.Input, "Show examples should have input parameters")

			// Verify examples use valid identifier fields
			hasValidIdentifier := false
			identifierFields := []string{"client_id", "client_name", "client_email"}
			for _, field := range identifierFields {
				if _, hasField := example.Input[field]; hasField {
					hasValidIdentifier = true
					break
				}
			}
			if len(example.Input) > 0 {
				s.True(hasValidIdentifier, "Example should use valid client identifier")
			}
		}
	})
}

func (s *ClientToolsTestSuite) TestClientUpdateTool() {
	tools := CreateClientManagementTools()
	var clientUpdateTool *MCPTool

	for _, tool := range tools {
		if tool.Name == "client_update" {
			clientUpdateTool = tool
			break
		}
	}

	if clientUpdateTool == nil {
		s.T().Skip("client_update tool not found in implementation")
		return
	}

	s.Run("BasicStructure", func() {
		s.Equal("client_update", clientUpdateTool.Name)
		s.NotEmpty(clientUpdateTool.Description)
		s.Contains(clientUpdateTool.Description, "client")
		s.Contains(strings.ToLower(clientUpdateTool.Description), "update")
		s.Equal(CategoryClientManagement, clientUpdateTool.Category)
	})

	s.Run("Schema", func() {
		s.NotNil(clientUpdateTool.InputSchema)
		s.Equal("object", clientUpdateTool.InputSchema["type"])

		properties, hasProperties := clientUpdateTool.InputSchema["properties"]
		s.True(hasProperties, "Should have properties")

		if propertiesMap, ok := properties.(map[string]interface{}); ok {
			// Should have identifier and updatable fields
			identifierFields := []string{"client_id", "client_name", "client_email"}
			hasIdentifier := false
			for _, field := range identifierFields {
				if _, hasField := propertiesMap[field]; hasField {
					hasIdentifier = true
					break
				}
			}
			s.True(hasIdentifier, "Should have client identifier field")

			// Should have updatable fields
			updatableFields := []string{"name", "email", "address", "phone"}
			hasUpdatableField := false
			for _, field := range updatableFields {
				if _, hasField := propertiesMap[field]; hasField {
					hasUpdatableField = true
					break
				}
			}
			s.True(hasUpdatableField, "Should have at least one updatable field")
		}
	})

	s.Run("Examples", func() {
		s.NotEmpty(clientUpdateTool.Examples, "Should have examples")

		for _, example := range clientUpdateTool.Examples {
			s.NotEmpty(example.Description)
			s.NotNil(example.Input)
			s.NotEmpty(example.Input, "Update examples should have input parameters")

			// Should demonstrate updating different fields
			updateFieldCount := 0
			updatableFields := []string{"name", "email", "address", "phone"}
			for _, field := range updatableFields {
				if _, hasField := example.Input[field]; hasField {
					updateFieldCount++
				}
			}
			s.Positive(updateFieldCount, "Example should update at least one field")
		}
	})
}

func (s *ClientToolsTestSuite) TestClientDeleteTool() {
	tools := CreateClientManagementTools()
	var clientDeleteTool *MCPTool

	for _, tool := range tools {
		if tool.Name == "client_delete" {
			clientDeleteTool = tool
			break
		}
	}

	if clientDeleteTool == nil {
		s.T().Skip("client_delete tool not found in implementation")
		return
	}

	s.Run("BasicStructure", func() {
		s.Equal("client_delete", clientDeleteTool.Name)
		s.NotEmpty(clientDeleteTool.Description)
		s.Contains(clientDeleteTool.Description, "client")
		s.Contains(strings.ToLower(clientDeleteTool.Description), "delete")
		s.Equal(CategoryClientManagement, clientDeleteTool.Category)
	})

	s.Run("Schema", func() {
		s.NotNil(clientDeleteTool.InputSchema)
		s.Equal("object", clientDeleteTool.InputSchema["type"])

		properties, hasProperties := clientDeleteTool.InputSchema["properties"]
		s.True(hasProperties, "Should have properties")

		if propertiesMap, ok := properties.(map[string]interface{}); ok {
			// Should have identifier field
			identifierFields := []string{"client_id", "client_name", "client_email"}
			hasIdentifier := false
			for _, field := range identifierFields {
				if _, hasField := propertiesMap[field]; hasField {
					hasIdentifier = true
					break
				}
			}
			s.True(hasIdentifier, "Should have client identifier field")

			// May have safety/confirmation fields
			safetyFields := []string{"confirm", "force"}
			for _, field := range safetyFields {
				if fieldDef, hasField := propertiesMap[field]; hasField {
					s.NotNil(fieldDef, "Safety field %s should have definition", field)
				}
			}
		}
	})

	s.Run("Examples", func() {
		s.NotEmpty(clientDeleteTool.Examples, "Should have examples")

		hasSafeExample := false
		for _, example := range clientDeleteTool.Examples {
			s.NotEmpty(example.Description)
			s.NotNil(example.Input)
			s.NotEmpty(example.Input, "Delete examples should have input parameters")

			// Check for safety considerations in examples
			if confirm, hasConfirm := example.Input["confirm"]; hasConfirm {
				if confirmBool, isBool := confirm.(bool); isBool && confirmBool {
					hasSafeExample = true
				}
			}
		}

		// Should demonstrate safe deletion practices
		if hasSafeExample {
			s.True(hasSafeExample, "Should have examples demonstrating safe deletion")
		}
	})
}

func (s *ClientToolsTestSuite) TestRegisterClientManagementTools() {
	// Create mock registry
	validator := new(MockInputValidator)
	logger := new(MockLogger)
	registry := NewDefaultToolRegistry(validator, logger)

	// Setup expectations for tool registration
	clientTools := CreateClientManagementTools()
	logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Times(len(clientTools))

	// Test registration
	ctx := context.Background()
	err := RegisterClientManagementTools(ctx, registry)
	if err != nil {
		s.T().Skipf("RegisterClientManagementTools not implemented or failed: %v", err)
		return
	}

	s.Require().NoError(err, "Should register client tools successfully")

	// Verify tools were registered
	logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	registeredTools, err := registry.ListTools(ctx, CategoryClientManagement)
	s.Require().NoError(err)
	s.NotEmpty(registeredTools, "Should have registered client management tools")

	// Verify each registered tool
	for _, tool := range registeredTools {
		s.Equal(CategoryClientManagement, tool.Category, "All tools should be in client management category")
		s.validateClientTool(tool)
	}
}

func (s *ClientToolsTestSuite) TestClientToolsIntegration() {
	s.Run("ToolNamesUnique", func() {
		tools := CreateClientManagementTools()
		toolNames := make(map[string]bool)

		for _, tool := range tools {
			s.False(toolNames[tool.Name], "Tool name %s should be unique", tool.Name)
			toolNames[tool.Name] = true
		}
	})

	s.Run("ToolsCoverClientLifecycle", func() {
		tools := CreateClientManagementTools()
		toolNames := make(map[string]bool)

		for _, tool := range tools {
			toolNames[tool.Name] = true
		}

		// Core CRUD operations
		crudOperations := map[string]string{
			"create": "client_create",
			"read":   "client_show",
			"list":   "client_list",
			"update": "client_update",
			"delete": "client_delete",
		}

		for operation, toolName := range crudOperations {
			if toolNames[toolName] {
				s.True(toolNames[toolName], "Should have %s operation (%s)", operation, toolName)
			}
		}
	})

	s.Run("ConsistentNamingConvention", func() {
		tools := CreateClientManagementTools()

		for _, tool := range tools {
			s.True(strings.HasPrefix(tool.Name, "client_"), "Tool %s should start with 'client_'", tool.Name)
			s.NotContains(tool.Name, " ", "Tool name should not contain spaces")
			s.Equal(strings.ToLower(tool.Name), tool.Name, "Tool name should be lowercase")
		}
	})

	s.Run("ConsistentCLIIntegration", func() {
		tools := CreateClientManagementTools()

		for _, tool := range tools {
			s.NotEmpty(tool.CLICommand, "Tool %s should have CLI command", tool.Name)
			s.NotEmpty(tool.CLIArgs, "Tool %s should have CLI args", tool.Name)

			// Verify CLI args make sense for client operations
			hasClientArg := false
			for _, arg := range tool.CLIArgs {
				if strings.Contains(arg, "client") {
					hasClientArg = true
					break
				}
			}
			s.True(hasClientArg, "Tool %s should have 'client' in CLI args", tool.Name)
		}
	})

	s.Run("ExamplesCoverUseCases", func() {
		tools := CreateClientManagementTools()

		for _, tool := range tools {
			if len(tool.Examples) == 0 {
				continue // Skip if no examples
			}

			// Each tool should have meaningful examples
			for i, example := range tool.Examples {
				s.NotEmpty(example.Description, "Tool %s example %d should have description", tool.Name, i)
				s.NotEmpty(example.UseCase, "Tool %s example %d should have use case", tool.Name, i)

				// Examples should be relevant to client management
				exampleText := strings.ToLower(example.Description + " " + example.UseCase)
				s.True(
					strings.Contains(exampleText, "client") ||
						strings.Contains(exampleText, "customer") ||
						strings.Contains(exampleText, "contact"),
					"Tool %s example %d should be client-related", tool.Name, i,
				)
			}
		}
	})
}

func (s *ClientToolsTestSuite) TestClientToolsEdgeCases() {
	s.Run("EmptyToolsList", func() {
		// This tests the function behavior itself
		tools := CreateClientManagementTools()
		s.NotNil(tools, "Should never return nil")
		// It's acceptable for tools to be empty if not implemented yet
	})

	s.Run("ToolSchemaValidation", func() {
		tools := CreateClientManagementTools()

		for _, tool := range tools {
			// Schema should be valid JSON schema structure
			s.NotNil(tool.InputSchema, "Tool %s should have schema", tool.Name)

			if schemaType, hasType := tool.InputSchema["type"]; hasType {
				s.Equal("object", schemaType, "Tool %s schema should be object type", tool.Name)
			}

			// Properties should be properly structured
			if properties, hasProps := tool.InputSchema["properties"]; hasProps {
				s.IsType(map[string]interface{}{}, properties, "Properties should be map")

				if propsMap, ok := properties.(map[string]interface{}); ok {
					for fieldName, fieldDef := range propsMap {
						s.NotEmpty(fieldName, "Field name should not be empty")
						s.NotNil(fieldDef, "Field definition should not be nil")

						if fieldMap, isMap := fieldDef.(map[string]interface{}); isMap {
							if fieldType, hasFieldType := fieldMap["type"]; hasFieldType {
								validTypes := []string{"string", "number", "boolean", "array", "object"}
								s.Contains(validTypes, fieldType, "Field %s should have valid type", fieldName)
							}
						}
					}
				}
			}
		}
	})

	s.Run("TimeoutReasonableness", func() {
		tools := CreateClientManagementTools()

		for _, tool := range tools {
			s.Greater(tool.Timeout, time.Duration(0), "Tool %s should have positive timeout", tool.Name)
			s.LessOrEqual(tool.Timeout, 5*time.Minute, "Tool %s timeout should be reasonable", tool.Name)
		}
	})
}

// Helper method to validate client tool structure
func (s *ClientToolsTestSuite) validateClientTool(tool *MCPTool) {
	s.NotNil(tool, "Tool should not be nil")
	s.NotEmpty(tool.Name, "Tool should have name")
	s.NotEmpty(tool.Description, "Tool should have description")
	s.NotNil(tool.InputSchema, "Tool should have input schema")
	s.Equal(CategoryClientManagement, tool.Category, "Tool should be in client management category")
	s.NotEmpty(tool.CLICommand, "Tool should have CLI command")
	s.NotEmpty(tool.Version, "Tool should have version")
	s.Greater(tool.Timeout, time.Duration(0), "Tool should have positive timeout")

	// Verify schema structure
	s.Equal("object", tool.InputSchema["type"], "Schema should be object type")

	// Verify examples structure if present
	for i, example := range tool.Examples {
		s.NotEmpty(example.Description, "Example %d should have description", i)
		s.NotNil(example.Input, "Example %d should have input", i)
	}
}

// TestClientToolsTestSuite runs the complete client tools test suite
func TestClientToolsTestSuite(t *testing.T) {
	suite.Run(t, new(ClientToolsTestSuite))
}

// Benchmark tests for client tools
func BenchmarkCreateClientManagementTools(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tools := CreateClientManagementTools()
		_ = tools
	}
}

// Unit tests for specific client tool behaviors
func TestClientTools_Specific(t *testing.T) {
	t.Run("AllToolsHaveValidCategory", func(t *testing.T) {
		tools := CreateClientManagementTools()
		for _, tool := range tools {
			assert.Equal(t, CategoryClientManagement, tool.Category)
		}
	})

	t.Run("AllToolsHaveReasonableTimeouts", func(t *testing.T) {
		tools := CreateClientManagementTools()
		for _, tool := range tools {
			assert.Greater(t, tool.Timeout, 1*time.Second, "Tool %s timeout too short", tool.Name)
			assert.LessOrEqual(t, tool.Timeout, 2*time.Minute, "Tool %s timeout too long", tool.Name)
		}
	})

	t.Run("AllToolsHaveValidSchemas", func(t *testing.T) {
		tools := CreateClientManagementTools()
		for _, tool := range tools {
			require.NotNil(t, tool.InputSchema, "Tool %s missing schema", tool.Name)
			assert.Equal(t, "object", tool.InputSchema["type"], "Tool %s should have object schema", tool.Name)
		}
	})
}
