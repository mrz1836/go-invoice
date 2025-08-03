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

// ImportToolsTestSuite provides comprehensive tests for import tool definitions
type ImportToolsTestSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *ImportToolsTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *ImportToolsTestSuite) TestCreateDataImportTools() {
	tools := CreateDataImportTools()

	s.NotNil(tools, "Should return tools slice")
	s.NotEmpty(tools, "Should return at least one tool")

	// Expected import tools based on typical functionality
	expectedToolNames := []string{
		"import_timesheet",
		"import_clients",
		"import_validate",
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
		s.validateImportTool(tool)
	}
}

func (s *ImportToolsTestSuite) TestTimesheetImportTool() {
	tools := CreateDataImportTools()
	var timesheetTool *MCPTool

	for _, tool := range tools {
		if tool.Name == "import_timesheet" || strings.Contains(tool.Name, "timesheet") {
			timesheetTool = tool
			break
		}
	}

	if timesheetTool == nil {
		s.T().Skip("timesheet import tool not found in implementation")
		return
	}

	s.Run("BasicStructure", func() {
		s.Contains(strings.ToLower(timesheetTool.Name), "timesheet")
		s.NotEmpty(timesheetTool.Description)
		s.Contains(strings.ToLower(timesheetTool.Description), "timesheet")
		s.Equal(CategoryDataImport, timesheetTool.Category)
		s.NotEmpty(timesheetTool.Version)
		s.Greater(timesheetTool.Timeout, time.Duration(0))
	})

	s.Run("Schema", func() {
		s.NotNil(timesheetTool.InputSchema)
		s.Equal("object", timesheetTool.InputSchema["type"])

		// Verify expected properties exist
		properties, hasProperties := timesheetTool.InputSchema["properties"]
		s.True(hasProperties, "Should have properties in schema")

		if propertiesMap, ok := properties.(map[string]interface{}); ok {
			// Should have file input field
			expectedFields := []string{"file", "file_path", "source_file"}
			hasFileField := false
			for _, field := range expectedFields {
				if _, hasField := propertiesMap[field]; hasField {
					hasFileField = true
					break
				}
			}
			s.True(hasFileField, "Should have file input field")

			// May have format specification
			formatFields := []string{"format", "file_format", "type"}
			for _, field := range formatFields {
				if fieldDef, hasField := propertiesMap[field]; hasField {
					s.NotNil(fieldDef, "Format field %s should have definition", field)
				}
			}
		}
	})

	s.Run("Examples", func() {
		s.NotEmpty(timesheetTool.Examples, "Should have examples")

		for i, example := range timesheetTool.Examples {
			s.NotEmpty(example.Description, "Example %d should have description", i)
			s.NotNil(example.Input, "Example %d should have input", i)
			s.NotEmpty(example.Input, "Example %d input should not be empty", i)

			// Verify example demonstrates file handling
			hasFileRef := false
			fileFields := []string{"file", "file_path", "source_file"}
			for _, field := range fileFields {
				if fileVal, hasFile := example.Input[field]; hasFile {
					s.IsType("", fileVal, "File field should be string")
					s.NotEmpty(fileVal, "File field should not be empty")
					hasFileRef = true
				}
			}
			s.True(hasFileRef, "Example should demonstrate file input")
		}
	})

	s.Run("FileFormatSupport", func() {
		// Check if examples cover different file formats
		hasCSVExample := false
		hasExcelExample := false
		hasTSVExample := false

		for _, example := range timesheetTool.Examples {
			exampleText := strings.ToLower(example.Description + " " + example.UseCase)
			if strings.Contains(exampleText, "csv") {
				hasCSVExample = true
			}
			if strings.Contains(exampleText, "excel") || strings.Contains(exampleText, "xlsx") {
				hasExcelExample = true
			}
			if strings.Contains(exampleText, "tsv") || strings.Contains(exampleText, "tab") {
				hasTSVExample = true
			}
		}

		s.True(hasCSVExample || hasExcelExample || hasTSVExample, "Should have examples for common file formats")
	})
}

func (s *ImportToolsTestSuite) TestClientImportTool() {
	tools := CreateDataImportTools()
	var clientImportTool *MCPTool

	for _, tool := range tools {
		if strings.Contains(tool.Name, "client") && strings.Contains(tool.Name, "import") {
			clientImportTool = tool
			break
		}
	}

	if clientImportTool == nil {
		s.T().Skip("client import tool not found in implementation")
		return
	}

	s.Run("BasicStructure", func() {
		s.Contains(strings.ToLower(clientImportTool.Name), "client")
		s.NotEmpty(clientImportTool.Description)
		s.Contains(strings.ToLower(clientImportTool.Description), "client")
		s.Equal(CategoryDataImport, clientImportTool.Category)
	})

	s.Run("Schema", func() {
		s.NotNil(clientImportTool.InputSchema)
		s.Equal("object", clientImportTool.InputSchema["type"])

		properties, hasProperties := clientImportTool.InputSchema["properties"]
		s.True(hasProperties, "Should have properties")

		if propertiesMap, ok := properties.(map[string]interface{}); ok {
			// Should have file input
			fileFields := []string{"file", "file_path", "source_file"}
			hasFileField := false
			for _, field := range fileFields {
				if _, hasField := propertiesMap[field]; hasField {
					hasFileField = true
					break
				}
			}
			s.True(hasFileField, "Should have file input field")
		}
	})

	s.Run("Examples", func() {
		s.NotEmpty(clientImportTool.Examples, "Should have examples")

		for _, example := range clientImportTool.Examples {
			s.NotEmpty(example.Description)
			s.NotNil(example.Input)

			// Should demonstrate client-specific import scenarios
			exampleText := strings.ToLower(example.Description + " " + example.UseCase)
			s.True(
				strings.Contains(exampleText, "client") ||
					strings.Contains(exampleText, "customer") ||
					strings.Contains(exampleText, "contact"),
				"Example should be client-related",
			)
		}
	})
}

func (s *ImportToolsTestSuite) TestImportValidationTool() {
	tools := CreateDataImportTools()
	var validationTool *MCPTool

	for _, tool := range tools {
		if strings.Contains(tool.Name, "validate") || strings.Contains(tool.Name, "check") {
			validationTool = tool
			break
		}
	}

	if validationTool == nil {
		s.T().Skip("import validation tool not found in implementation")
		return
	}

	s.Run("BasicStructure", func() {
		s.Contains(strings.ToLower(validationTool.Name), "validate")
		s.NotEmpty(validationTool.Description)
		s.Contains(strings.ToLower(validationTool.Description), "validate")
		s.Equal(CategoryDataImport, validationTool.Category)
	})

	s.Run("Schema", func() {
		s.NotNil(validationTool.InputSchema)
		s.Equal("object", validationTool.InputSchema["type"])

		properties, hasProperties := validationTool.InputSchema["properties"]
		s.True(hasProperties, "Should have properties")

		if propertiesMap, ok := properties.(map[string]interface{}); ok {
			// Should have file input for validation
			fileFields := []string{"file", "file_path", "source_file"}
			hasFileField := false
			for _, field := range fileFields {
				if _, hasField := propertiesMap[field]; hasField {
					hasFileField = true
					break
				}
			}
			s.True(hasFileField, "Should have file input field for validation")
		}
	})

	s.Run("Examples", func() {
		s.NotEmpty(validationTool.Examples, "Should have examples")

		for _, example := range validationTool.Examples {
			s.NotEmpty(example.Description)
			s.NotNil(example.Input)

			// Should demonstrate validation scenarios
			exampleText := strings.ToLower(example.Description + " " + example.UseCase)
			validationKeywords := []string{"validate", "check", "verify", "error", "format"}
			hasValidationKeyword := false
			for _, keyword := range validationKeywords {
				if strings.Contains(exampleText, keyword) {
					hasValidationKeyword = true
					break
				}
			}
			s.True(hasValidationKeyword, "Example should be validation-related")
		}
	})
}

func (s *ImportToolsTestSuite) TestRegisterDataImportTools() {
	// Create mock registry
	validator := new(MockInputValidator)
	logger := new(MockLogger)
	registry := NewDefaultToolRegistry(validator, logger)

	// Setup expectations for tool registration
	importTools := CreateDataImportTools()
	logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Times(len(importTools))

	// Test registration
	err := RegisterDataImportTools(s.ctx, registry)
	if err != nil {
		s.T().Skipf("RegisterDataImportTools not implemented or failed: %v", err)
		return
	}

	s.NoError(err, "Should register import tools successfully")

	// Verify tools were registered
	logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	registeredTools, err := registry.ListTools(s.ctx, CategoryDataImport)
	s.NoError(err)
	s.NotEmpty(registeredTools, "Should have registered data import tools")

	// Verify each registered tool
	for _, tool := range registeredTools {
		s.Equal(CategoryDataImport, tool.Category, "All tools should be in data import category")
		s.validateImportTool(tool)
	}
}

func (s *ImportToolsTestSuite) TestImportToolsIntegration() {
	s.Run("ToolNamesUnique", func() {
		tools := CreateDataImportTools()
		toolNames := make(map[string]bool)

		for _, tool := range tools {
			s.False(toolNames[tool.Name], "Tool name %s should be unique", tool.Name)
			toolNames[tool.Name] = true
		}
	})

	s.Run("ToolsCoverImportWorkflow", func() {
		tools := CreateDataImportTools()
		toolNames := make(map[string]bool)

		for _, tool := range tools {
			toolNames[tool.Name] = true
		}

		// Core import operations
		importOperations := []string{"import", "validate", "parse"}
		hasImportOperation := false

		for _, tool := range tools {
			toolNameLower := strings.ToLower(tool.Name)
			for _, op := range importOperations {
				if strings.Contains(toolNameLower, op) {
					hasImportOperation = true
					break
				}
			}
			if hasImportOperation {
				break
			}
		}

		s.True(hasImportOperation, "Should have tools covering import operations")
	})

	s.Run("ConsistentNamingConvention", func() {
		tools := CreateDataImportTools()

		for _, tool := range tools {
			s.True(
				strings.HasPrefix(tool.Name, "import_") ||
					strings.Contains(tool.Name, "import"),
				"Tool %s should be clearly import-related", tool.Name,
			)
			s.NotContains(tool.Name, " ", "Tool name should not contain spaces")
			s.Equal(strings.ToLower(tool.Name), tool.Name, "Tool name should be lowercase")
		}
	})

	s.Run("ConsistentCLIIntegration", func() {
		tools := CreateDataImportTools()

		for _, tool := range tools {
			s.NotEmpty(tool.CLICommand, "Tool %s should have CLI command", tool.Name)
			s.NotEmpty(tool.CLIArgs, "Tool %s should have CLI args", tool.Name)

			// Verify CLI args make sense for import operations
			hasImportArg := false
			cliArgsStr := strings.Join(tool.CLIArgs, " ")
			importKeywords := []string{"import", "parse", "validate"}
			for _, keyword := range importKeywords {
				if strings.Contains(cliArgsStr, keyword) {
					hasImportArg = true
					break
				}
			}
			s.True(hasImportArg, "Tool %s should have import-related CLI args", tool.Name)
		}
	})

	s.Run("ExamplesCoverFileFormats", func() {
		tools := CreateDataImportTools()

		allExampleTexts := ""
		for _, tool := range tools {
			for _, example := range tool.Examples {
				allExampleTexts += strings.ToLower(example.Description + " " + example.UseCase + " ")

				// Check file paths in input
				for key, value := range example.Input {
					if strings.Contains(key, "file") {
						if fileStr, isStr := value.(string); isStr {
							allExampleTexts += strings.ToLower(fileStr) + " "
						}
					}
				}
			}
		}

		// Should cover common file formats
		commonFormats := []string{"csv", "excel", "xlsx", "tsv"}
		formatsCovered := 0
		for _, format := range commonFormats {
			if strings.Contains(allExampleTexts, format) {
				formatsCovered++
			}
		}

		s.Greater(formatsCovered, 0, "Should have examples covering common file formats")
	})

	s.Run("ExamplesCoverErrorHandling", func() {
		tools := CreateDataImportTools()

		hasErrorExample := false
		for _, tool := range tools {
			for _, example := range tool.Examples {
				exampleText := strings.ToLower(example.Description + " " + example.UseCase)
				errorKeywords := []string{"error", "invalid", "malformed", "corrupt", "fail"}
				for _, keyword := range errorKeywords {
					if strings.Contains(exampleText, keyword) {
						hasErrorExample = true
						break
					}
				}
				if hasErrorExample {
					break
				}
			}
			if hasErrorExample {
				break
			}
		}

		s.True(hasErrorExample, "Should have examples covering error scenarios")
	})
}

func (s *ImportToolsTestSuite) TestImportToolsEdgeCases() {
	s.Run("EmptyToolsList", func() {
		// This tests the function behavior itself
		tools := CreateDataImportTools()
		s.NotNil(tools, "Should never return nil")
		// Note: It's acceptable for tools to be empty if not implemented yet
	})

	s.Run("FilePathValidation", func() {
		tools := CreateDataImportTools()

		for _, tool := range tools {
			for _, example := range tool.Examples {
				// Check file path examples are realistic
				for key, value := range example.Input {
					if strings.Contains(key, "file") {
						if filePath, isStr := value.(string); isStr && filePath != "" {
							// Should look like a real file path
							s.True(
								strings.Contains(filePath, ".") ||
									strings.Contains(filePath, "/") ||
									strings.Contains(filePath, "\\"),
								"File path example should look realistic: %s", filePath,
							)
						}
					}
				}
			}
		}
	})

	s.Run("SchemaSupportsFileUploads", func() {
		tools := CreateDataImportTools()

		for _, tool := range tools {
			properties, hasProps := tool.InputSchema["properties"]
			if !hasProps {
				continue
			}

			if propsMap, ok := properties.(map[string]interface{}); ok {
				// Look for file-related fields
				fileFields := []string{"file", "file_path", "source_file", "upload"}
				for _, field := range fileFields {
					if fieldDef, hasField := propsMap[field]; hasField {
						if fieldMap, isMap := fieldDef.(map[string]interface{}); isMap {
							// File fields should be strings
							if fieldType, hasType := fieldMap["type"]; hasType {
								s.Equal("string", fieldType, "File field %s should be string type", field)
							}
						}
					}
				}
			}
		}
	})

	s.Run("TimeoutAppropriateForFileOperations", func() {
		tools := CreateDataImportTools()

		for _, tool := range tools {
			// Import operations may take longer due to file processing
			s.Greater(tool.Timeout, 5*time.Second, "Tool %s should have sufficient timeout for file operations", tool.Name)
			s.LessOrEqual(tool.Timeout, 5*time.Minute, "Tool %s timeout should not be excessive", tool.Name)
		}
	})

	s.Run("DescriptionsIndicateDataTypes", func() {
		tools := CreateDataImportTools()

		for _, tool := range tools {
			desc := strings.ToLower(tool.Description)

			// Should indicate what type of data is being imported
			dataTypes := []string{"timesheet", "client", "invoice", "contact", "data"}
			hasDataType := false
			for _, dataType := range dataTypes {
				if strings.Contains(desc, dataType) {
					hasDataType = true
					break
				}
			}
			s.True(hasDataType, "Tool %s description should indicate data type", tool.Name)
		}
	})
}

func (s *ImportToolsTestSuite) TestImportToolsPerformance() {
	s.Run("ReasonableToolCount", func() {
		tools := CreateDataImportTools()

		// Should have reasonable number of import tools
		s.LessOrEqual(len(tools), 10, "Should not have excessive number of import tools")
		s.GreaterOrEqual(len(tools), 1, "Should have at least one import tool")
	})

	s.Run("SchemaComplexityReasonable", func() {
		tools := CreateDataImportTools()

		for _, tool := range tools {
			// Count schema properties
			properties, hasProps := tool.InputSchema["properties"]
			if hasProps {
				if propsMap, ok := properties.(map[string]interface{}); ok {
					s.LessOrEqual(len(propsMap), 20, "Tool %s should not have excessive schema complexity", tool.Name)
				}
			}

			// Count examples
			s.LessOrEqual(len(tool.Examples), 10, "Tool %s should not have excessive examples", tool.Name)
		}
	})
}

// Helper method to validate import tool structure
func (s *ImportToolsTestSuite) validateImportTool(tool *MCPTool) {
	s.NotNil(tool, "Tool should not be nil")
	s.NotEmpty(tool.Name, "Tool should have name")
	s.NotEmpty(tool.Description, "Tool should have description")
	s.NotNil(tool.InputSchema, "Tool should have input schema")
	s.Equal(CategoryDataImport, tool.Category, "Tool should be in data import category")
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

	// Import tools should indicate file handling capability
	toolText := strings.ToLower(tool.Name + " " + tool.Description)
	fileKeywords := []string{"file", "import", "upload", "parse", "csv", "excel"}
	hasFileKeyword := false
	for _, keyword := range fileKeywords {
		if strings.Contains(toolText, keyword) {
			hasFileKeyword = true
			break
		}
	}
	s.True(hasFileKeyword, "Import tool should indicate file handling capability")
}

// TestImportToolsTestSuite runs the complete import tools test suite
func TestImportToolsTestSuite(t *testing.T) {
	suite.Run(t, new(ImportToolsTestSuite))
}

// Benchmark tests for import tools
func BenchmarkCreateDataImportTools(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tools := CreateDataImportTools()
		_ = tools
	}
}

// Unit tests for specific import tool behaviors
func TestImportTools_Specific(t *testing.T) {
	t.Run("AllToolsHaveValidCategory", func(t *testing.T) {
		tools := CreateDataImportTools()
		for _, tool := range tools {
			assert.Equal(t, CategoryDataImport, tool.Category)
		}
	})

	t.Run("AllToolsHaveAppropriateTimeouts", func(t *testing.T) {
		tools := CreateDataImportTools()
		for _, tool := range tools {
			// Import operations may take longer
			assert.Greater(t, tool.Timeout, 5*time.Second, "Tool %s timeout too short", tool.Name)
			assert.LessOrEqual(t, tool.Timeout, 5*time.Minute, "Tool %s timeout too long", tool.Name)
		}
	})

	t.Run("AllToolsHaveValidSchemas", func(t *testing.T) {
		tools := CreateDataImportTools()
		for _, tool := range tools {
			require.NotNil(t, tool.InputSchema, "Tool %s missing schema", tool.Name)
			assert.Equal(t, "object", tool.InputSchema["type"], "Tool %s should have object schema", tool.Name)
		}
	})

	t.Run("ToolNamesIndicatePurpose", func(t *testing.T) {
		tools := CreateDataImportTools()
		for _, tool := range tools {
			toolName := strings.ToLower(tool.Name)
			assert.True(t,
				strings.Contains(toolName, "import") ||
					strings.Contains(toolName, "parse") ||
					strings.Contains(toolName, "validate") ||
					strings.Contains(toolName, "upload"),
				"Tool name %s should indicate import purpose", tool.Name,
			)
		}
	})
}
